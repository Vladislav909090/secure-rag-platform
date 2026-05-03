package httpupload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	knowledgev1 "secure-rag-platform/services/knowledge/gen/v1"
	"secure-rag-platform/services/knowledge/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const defaultChunkSize = 256 * 1024
const knowledgeDocumentsPrefix = "/knowledge/api/v1/documents"

type UploadHandlers struct {
	client    knowledgev1.KnowledgeServiceClient
	uc        *usecase.DocumentUsecase
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

type uploadVersionMultipartState struct {
	metaSent bool
	fileSent bool
}

func New(client knowledgev1.KnowledgeServiceClient, uc *usecase.DocumentUsecase) *UploadHandlers {
	return &UploadHandlers{client: client, uc: uc, chunkSize: defaultChunkSize}
}

func (h *UploadHandlers) CreateDocument(gateway http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != knowledgeDocumentsPrefix {
			gateway.ServeHTTP(w, r)
			return
		}

		if !isMultipart(r.Header.Get("Content-Type")) {
			gateway.ServeHTTP(w, r)
			return
		}

		if err := h.handleCreateDocumentMultipart(r.Context(), w, r); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
		}
	}
}

func (h *UploadHandlers) UploadVersion(gateway http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Перехватываем скачивание файлов:
		// - GET /knowledge/api/v1/documents/{uuid}/file
		// - GET /knowledge/api/v1/documents/{uuid}/versions/{version}/file
		if r.Method == http.MethodGet &&
			strings.HasPrefix(r.URL.Path, knowledgeDocumentsPrefix+"/") &&
			strings.HasSuffix(r.URL.Path, "/file") &&
			h.uc != nil {
			trimmed := strings.TrimPrefix(r.URL.Path, knowledgeDocumentsPrefix+"/")
			if strings.Contains(trimmed, "/versions/") {
				h.handleDownloadVersionFile(w, r)
				return
			}

			h.handleDownloadFile(w, r)
			return
		}

		if r.Method != http.MethodPost || !strings.HasPrefix(r.URL.Path, knowledgeDocumentsPrefix+"/") || !strings.HasSuffix(r.URL.Path, "/versions") {
			gateway.ServeHTTP(w, r)
			return
		}

		if !isMultipart(r.Header.Get("Content-Type")) {
			gateway.ServeHTTP(w, r)
			return
		}

		docUUID, ok := extractDocumentUUID(r.URL.Path)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid upload path")
			return
		}

		if err := h.handleUploadVersionMultipart(r.Context(), w, r, docUUID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
		}
	}
}

func (h *UploadHandlers) handleCreateDocumentMultipart(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	reader, err := r.MultipartReader()
	if err != nil {
		return err
	}

	stream, err := h.client.CreateDocumentStream(ctx)
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
			return nextErr
		}

		if partErr := h.processCreateDocumentPart(part, stream, state); partErr != nil {
			return partErr
		}
	}

	if !state.metaSent || !state.fileSent {
		return errors.New("multipart payload must contain title and file fields")
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	return writeProtoJSON(w, http.StatusOK, resp)
}

func (h *UploadHandlers) handleUploadVersionMultipart(ctx context.Context, w http.ResponseWriter, r *http.Request, docUUID string) error {
	reader, err := r.MultipartReader()
	if err != nil {
		return err
	}

	stream, err := h.client.UploadVersionStream(ctx)
	if err != nil {
		return err
	}

	state := &uploadVersionMultipartState{}

	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return nextErr
		}

		if partErr := h.processUploadVersionPart(part, stream, docUUID, state); partErr != nil {
			return partErr
		}
	}

	if !state.metaSent || !state.fileSent {
		return errors.New("multipart payload must contain file field")
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	return writeProtoJSON(w, http.StatusOK, resp)
}

func (h *UploadHandlers) processCreateDocumentPart(part *multipart.Part, stream knowledgev1.KnowledgeService_CreateDocumentStreamClient, state *createDocumentMultipartState) error {
	switch part.FormName() {
	case "title":
		value, err := readLimitedPart(part, 1<<20)
		if err != nil {
			return err
		}
		state.title = strings.TrimSpace(string(value))
		return nil
	case "description":
		value, err := readLimitedPart(part, 1<<20)
		if err != nil {
			return err
		}
		state.hasDescription = true
		state.description = string(value)
		return nil
	case "attributes":
		value, err := readLimitedPart(part, 4<<20)
		if err != nil {
			return err
		}
		attrs, err := parseStructAttributes(value)
		if err != nil {
			return err
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

func (h *UploadHandlers) sendCreateDocumentFilePart(part *multipart.Part, stream knowledgev1.KnowledgeService_CreateDocumentStreamClient, state *createDocumentMultipartState) error {
	if state.metaSent {
		return errors.New("multiple file parts are not supported")
	}
	if strings.TrimSpace(state.title) == "" {
		return errors.New("title is required before file part")
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
			Payload: &knowledgev1.CreateDocumentStreamRequest_Chunk{Chunk: &knowledgev1.FileChunk{Data: chunk}},
		})
	}); err != nil {
		return err
	}

	state.fileSent = true
	return nil
}

func (h *UploadHandlers) processUploadVersionPart(part *multipart.Part, stream knowledgev1.KnowledgeService_UploadVersionStreamClient, docUUID string, state *uploadVersionMultipartState) error {
	if part.FormName() != "file" {
		_, _ = io.Copy(io.Discard, part)
		return nil
	}

	if state.metaSent {
		return errors.New("multiple file parts are not supported")
	}

	fileName := strings.TrimSpace(part.FileName())
	if fileName == "" {
		return errors.New("uploaded file must include filename")
	}

	mimeType := part.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension("." + extension(fileName))
	}

	if err := stream.Send(&knowledgev1.UploadVersionStreamRequest{
		Payload: &knowledgev1.UploadVersionStreamRequest_Meta{Meta: &knowledgev1.UploadVersionMeta{
			DocumentUuid: docUUID,
			FileName:     fileName,
			MimeType:     mimeType,
		}},
	}); err != nil {
		return err
	}
	state.metaSent = true

	if err := sendChunks(part, h.chunkSize, func(chunk []byte) error {
		return stream.Send(&knowledgev1.UploadVersionStreamRequest{
			Payload: &knowledgev1.UploadVersionStreamRequest_Chunk{Chunk: &knowledgev1.FileChunk{Data: chunk}},
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

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + strings.ReplaceAll(message, "\"", "'") + `"}`))
}

func isMultipart(contentType string) bool {
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediatype == "multipart/form-data"
}

func (h *UploadHandlers) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, knowledgeDocumentsPrefix+"/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[1] != "file" || strings.TrimSpace(parts[0]) == "" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	docUUID := parts[0]

	dl, err := h.uc.DownloadFile(r.Context(), docUUID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrDocumentNotFound),
			errors.Is(err, usecase.ErrVersionNotFound),
			errors.Is(err, usecase.ErrFileNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrDocumentDeleted):
			writeError(w, http.StatusGone, err.Error())
		default:
			// Разворачиваем gRPC status errors после возможных gateway-вызовов.
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				writeError(w, http.StatusNotFound, st.Message())
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	defer dl.Body.Close()
	h.writeDownloadResponse(w, dl)
}

func (h *UploadHandlers) handleDownloadVersionFile(w http.ResponseWriter, r *http.Request) {
	docUUID, versionNumber, ok := extractVersionDownloadPath(r.URL.Path)
	if !ok {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	dl, err := h.uc.DownloadVersionFile(r.Context(), docUUID, versionNumber)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrDocumentNotFound),
			errors.Is(err, usecase.ErrVersionNotFound),
			errors.Is(err, usecase.ErrFileNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, usecase.ErrDocumentDeleted):
			writeError(w, http.StatusGone, err.Error())
		default:
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				writeError(w, http.StatusNotFound, st.Message())
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	defer dl.Body.Close()
	h.writeDownloadResponse(w, dl)
}

func (h *UploadHandlers) writeDownloadResponse(w http.ResponseWriter, dl *usecase.FileDownload) {

	mimeType := dl.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// RFC 6266: указываем и ASCII fallback в кавычках, и имя по RFC 5987.
	// Браузеры предпочитают filename*, если есть оба варианта; Swagger UI использует filename.
	fallback := asciiFilenameFallback(dl.FileName)
	escaped := strings.ReplaceAll(fallback, `"`, `\\"`)
	encodedName := url.PathEscape(dl.FileName)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, escaped, encodedName))
	w.Header().Set("Content-Length", strconv.FormatInt(dl.SizeBytes, 10))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, dl.Body)
}

func extractVersionDownloadPath(path string) (string, int32, bool) {
	trimmed := strings.TrimPrefix(path, knowledgeDocumentsPrefix+"/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 4 || parts[1] != "versions" || parts[3] != "file" || strings.TrimSpace(parts[0]) == "" {
		return "", 0, false
	}

	v, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || v <= 0 {
		return "", 0, false
	}

	return parts[0], int32(v), true
}

func extractDocumentUUID(path string) (string, bool) {
	trimmed := strings.TrimPrefix(path, knowledgeDocumentsPrefix+"/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 2 || parts[1] != "versions" || strings.TrimSpace(parts[0]) == "" {
		return "", false
	}
	return parts[0], true
}

func extension(fileName string) string {
	idx := strings.LastIndex(fileName, ".")
	if idx < 0 || idx+1 >= len(fileName) {
		return ""
	}
	return fileName[idx+1:]
}

func asciiFilenameFallback(fileName string) string {
	if strings.TrimSpace(fileName) == "" {
		return "download.bin"
	}

	ext := extension(fileName)
	name := fileName
	if ext != "" {
		name = strings.TrimSuffix(fileName, "."+ext)
	}

	var b strings.Builder
	b.Grow(len(name))

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		default:
			b.WriteRune('_')
		}
	}

	base := strings.Trim(b.String(), "_-")
	if base == "" {
		base = "download"
	}

	if ext == "" {
		return base + ".bin"
	}

	return base + "." + ext
}
