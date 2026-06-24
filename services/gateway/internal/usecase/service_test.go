package usecase

import (
	"io"
	"log/slog"
	"testing"

	iamv1 "secure-rag-platform/api/gen/go/iam/v1"
	knowledgev1 "secure-rag-platform/api/gen/go/knowledge/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

type gatewayTestDeps struct {
	rag        *MockRAGServiceClient
	knowledge  *MockKnowledgeServiceClient
	iam        *MockInternalIAMServiceClient
	auth       *MockAuthServiceClient
	users      *MockUserServiceClient
	roles      *MockRoleServiceClient
	attributes *MockAttributeServiceClient
	policy     *MockPolicyAuthorizer
}

func newGatewayTestService(t *testing.T) (*Service, gatewayTestDeps) {
	t.Helper()

	deps := gatewayTestDeps{
		rag:        NewMockRAGServiceClient(t),
		knowledge:  NewMockKnowledgeServiceClient(t),
		iam:        NewMockInternalIAMServiceClient(t),
		auth:       NewMockAuthServiceClient(t),
		users:      NewMockUserServiceClient(t),
		roles:      NewMockRoleServiceClient(t),
		attributes: NewMockAttributeServiceClient(t),
		policy:     NewMockPolicyAuthorizer(t),
	}

	svc := NewService(
		deps.rag,
		deps.knowledge,
		deps.iam,
		deps.auth,
		deps.users,
		deps.roles,
		deps.attributes,
		deps.policy,
		Defaults{TopK: 3, EmbeddingModelAlias: "embed-default", GenerationModelAlias: "gen-default"},
		false,
		false,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	return svc, deps
}

func expectValidToken(deps gatewayTestDeps, token string, subject *iamv1.SubjectContext) {
	deps.iam.EXPECT().
		ValidateAccessToken(mock.Anything, mock.MatchedBy(func(req *iamv1.ValidateAccessTokenRequest) bool {
			return req.GetAccessToken() == token
		})).
		Return(&iamv1.ValidateAccessTokenResponse{Valid: true, Subject: subject}, nil)
}

func gatewaySubject(roles ...string) *iamv1.SubjectContext {
	attrs, err := structpb.NewStruct(map[string]any{"department": "search"})
	if err != nil {
		panic(err)
	}

	return &iamv1.SubjectContext{
		UserId:     "u1",
		Login:      "alice",
		IsActive:   true,
		Roles:      roles,
		Attributes: attrs,
		CtxVer:     3,
	}
}

func gatewayDocument(uuid string, attrs map[string]any) *knowledgev1.Document {
	protoAttrs, err := structpb.NewStruct(attrs)
	if err != nil {
		panic(err)
	}

	return &knowledgev1.Document{
		Id:             7,
		Uuid:           uuid,
		Title:          "title",
		Description:    "description",
		Attributes:     protoAttrs,
		FileName:       "file.txt",
		FileExtension:  "txt",
		MimeType:       "text/plain",
		SizeBytes:      12,
		ChecksumSha256: "checksum",
		StorageKey:     "documents/" + uuid + "/file",
		IndexStatus:    "READY",
		CreatedAt:      "created",
		UpdatedAt:      "updated",
	}
}

func TestGatewayServiceReady(t *testing.T) {
	t.Parallel()

	assert.False(t, (*Service)(nil).Ready())
	assert.False(t, (&Service{}).Ready())

	svc, _ := newGatewayTestService(t)
	assert.True(t, svc.Ready())

	svc.disableAuth = true
	svc.iam = nil
	svc.auth = nil
	svc.users = nil
	svc.roles = nil
	svc.attributes = nil
	assert.True(t, svc.Ready())

	require.NotNil(t, NewService(nil, nil, nil, nil, nil, nil, nil, nil, Defaults{}, false, false, nil))
}
