package grpc

import (
	"errors"
	"testing"

	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGenerationParamsFromProto(t *testing.T) {
	params := generationParamsFromProto(&aiinferencev1.GenerationParams{
		Temperature:      wrapperspb.Float(0.7),
		TopP:             wrapperspb.Float(0.9),
		MaxTokens:        wrapperspb.Int32(128),
		PresencePenalty:  wrapperspb.Float(0.1),
		FrequencyPenalty: wrapperspb.Float(0.2),
	})

	if *params.Temperature != 0.7 || *params.TopP != 0.9 || *params.MaxTokens != 128 {
		t.Fatalf("unexpected params: %#v", params)
	}
	if generationParamsFromProto(nil).Temperature != nil {
		t.Fatalf("nil params should produce zero-value usecase params")
	}
}

func TestTaskTypeMapping(t *testing.T) {
	task, err := taskTypeFromProto(aiinferencev1.TaskType_TASK_TYPE_EMBEDDING)
	if err != nil || task != config.TaskEmbedding {
		t.Fatalf("taskTypeFromProto() = %q, %v", task, err)
	}
	if got := taskTypeToProto(config.TaskGeneration); got != aiinferencev1.TaskType_TASK_TYPE_GENERATION {
		t.Fatalf("taskTypeToProto() = %v", got)
	}
}

func TestToGRPCError(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{usecase.ErrInvalidArgument, codes.InvalidArgument},
		{usecase.ErrAliasNotFound, codes.NotFound},
		{usecase.ErrAliasTaskMismatch, codes.FailedPrecondition},
		{usecase.ErrProviderNotConfigured, codes.Unavailable},
		{usecase.ErrProviderFailed, codes.Unavailable},
		{errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		if got := status.Code(toGRPCError(tt.err)); got != tt.code {
			t.Fatalf("toGRPCError(%v) code = %v, want %v", tt.err, got, tt.code)
		}
	}
}
