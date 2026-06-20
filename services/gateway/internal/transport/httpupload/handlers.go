package httpupload

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/gateway/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	defaultChunkSize     = 256 * 1024
	gatewayDocumentsPath = "/gateway/api/v1/documents"
)

type UploadHandlers struct {
	uc        *usecase.Service
	logger    *slog.Logger
	chunkSize int
}

type createDocumentMultipartState struct {
	title          string
	description    string
	hasDescription bool
	attrs          *structpb.Struct
	metaSent       bool
	fileSent       bool
}

func New(uc *usecase.Service, logger *slog.Logger) *UploadHandlers {
	if logger == nil {
		logger = slog.Default()
	}

	return &UploadHandlers{uc: uc, logger: logger, chunkSize: defaultChunkSize}
}

// CreateDocument перехватывает multipart-загрузку документа через gateway
func (h *UploadHandlers) CreateDocument(gateway http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != gatewayDocumentsPath {
			gateway.ServeHTTP(w, r)
			return
		}

		if !isMultipart(r.Header.Get("Content-Type")) {
			gateway.ServeHTTP(w, r)
			return
		}

		if h.uc == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "service not configured")
			return
		}

		if err := h.handleCreateDocumentMultipart(w, r); err != nil {
			h.logger.WarnContext(r.Context(), "не удалось загрузить документ",
				"component", "gateway.http-upload",
				"error", err,
			)
			writeUsecaseError(w, err)
		}
	}
}

func (h *UploadHandlers) handleCreateDocumentMultipart(w http.ResponseWriter, r *http.Request) error {
	reader, err := r.MultipartReader()
	if err != nil {
		return usecase.ErrInvalidRequest
	}

	stream, err := h.uc.OpenCreateDocumentStream(r.Context(), extractAccessToken(r))
	if err != nil {
		return err
	}

	state := &createDocumentMultipartState{}
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return usecase.ErrInvalidRequest
		}

		if partErr := h.processCreateDocumentPart(part, stream, state); partErr != nil {
			return partErr
		}
	}

	if !state.metaSent || !state.fileSent {
		return usecase.ErrInvalidRequest
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	return writeProtoJSON(w, http.StatusOK, resp)
}

func (h *UploadHandlers) processCreateDocumentPart(
	part *multipart.Part,
	stream knowledgev1.KnowledgeService_CreateDocumentStreamClient,
	state *createDocumentMultipartState,
) error {
	switch part.FormName() {
	case "title":
		value, err := readLimitedPart(part, 1<<20)
		if err != nil {
			return usecase.ErrInvalidRequest
		}
		state.title = strings.TrimSpace(string(value))

		return nil
	case "description":
		value, err := readLimitedPart(part, 1<<20)
		if err != nil {
			return usecase.ErrInvalidRequest
		}
		state.hasDescription = true
		state.description = string(value)

		return nil
	case "attributes":
		value, err := readLimitedPart(part, 4<<20)
		if err != nil {
			return usecase.ErrInvalidRequest
		}
		attrs, err := parseStructAttributes(value)
		if err != nil {
			return usecase.ErrInvalidRequest
		}
		state.attrs = attrs

		return nil
	case "file":
		return h.sendCreateDocumentFilePart(part, stream, state)
	default:
		_, _ = io.Copy(io.Discard, part)
		return nil
	}
}

func (h *UploadHandlers) sendCreateDocumentFilePart(
	part *multipart.Part,
	stream knowledgev1.KnowledgeService_CreateDocumentStreamClient,
	state *createDocumentMultipartState,
) error {
	if state.metaSent || strings.TrimSpace(state.title) == "" {
		return usecase.ErrInvalidRequest
	}

	fileName := part.FileName()
	mimeType := part.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension("." + extension(fileName))
	}

	meta := &knowledgev1.CreateDocumentMeta{
		Title:      state.title,
		FileName:   fileName,
		MimeType:   mimeType,
		Attributes: state.attrs,
	}
	if state.hasDescription {
		meta.Description = &state.description
	}

	if err := stream.Send(&knowledgev1.CreateDocumentStreamRequest{
		Payload: &knowledgev1.CreateDocumentStreamRequest_Meta{Meta: meta},
	}); err != nil {
		return err
	}
	state.metaSent = true

	if err := sendChunks(part, h.chunkSize, func(chunk []byte) error {
		return stream.Send(&knowledgev1.CreateDocumentStreamRequest{
			Payload: &knowledgev1.CreateDocumentStreamRequest_Chunk{
				Chunk: &knowledgev1.FileChunk{Data: chunk},
			},
		})
	}); err != nil {
		return err
	}

	state.fileSent = true

	return nil
}

func readLimitedPart(part *multipart.Part, limit int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(part, limit))
}

func parseStructAttributes(rawJSON []byte) (*structpb.Struct, error) {
	if len(strings.TrimSpace(string(rawJSON))) == 0 {
		return nil, nil
	}

	var raw map[string]any
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		return nil, err
	}

	return structpb.NewStruct(raw)
}

func sendChunks(r io.Reader, chunkSize int, send func([]byte) error) error {
	buf := make([]byte, chunkSize)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if sendErr := send(chunk); sendErr != nil {
				return sendErr
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func writeProtoJSON(w http.ResponseWriter, status int, msg any) error {
	m, ok := msg.(proto.Message)
	if !ok {
		return errors.New("invalid proto response")
	}

	b, err := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}.Marshal(m)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case status.Code(err) == codes.InvalidArgument:
		writeJSONError(w, http.StatusBadRequest, status.Convert(err).Message())
	case status.Code(err) == codes.Unauthenticated:
		writeJSONError(w, http.StatusUnauthorized, status.Convert(err).Message())
	case status.Code(err) == codes.PermissionDenied:
		writeJSONError(w, http.StatusForbidden, status.Convert(err).Message())
	case status.Code(err) == codes.NotFound:
		writeJSONError(w, http.StatusNotFound, status.Convert(err).Message())
	case errors.Is(err, usecase.ErrNotConfigured),
		errors.Is(err, usecase.ErrPolicyRequired),
		errors.Is(err, usecase.ErrPolicyUnavailable):
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, usecase.ErrInvalidRequest):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, usecase.ErrUnauthorized):
		writeJSONError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, usecase.ErrForbidden):
		writeJSONError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, usecase.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + strings.ReplaceAll(message, "\"", "'") + `"}`))
}

func extractAccessToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}

	return value
}

func isMultipart(contentType string) bool {
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	return mediatype == "multipart/form-data"
}

func extension(fileName string) string {
	idx := strings.LastIndex(fileName, ".")
	if idx < 0 || idx+1 >= len(fileName) {
		return ""
	}

	return fileName[idx+1:]
}
