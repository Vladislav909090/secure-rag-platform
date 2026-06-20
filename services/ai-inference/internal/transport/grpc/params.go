package grpc

import (
	aiinferencev1 "secure-rag-platform/api/gen/go/aiinference/v1"
	"secure-rag-platform/services/ai-inference/internal/config"
	"secure-rag-platform/services/ai-inference/internal/usecase"
)

func generationParamsFromProto(params *aiinferencev1.GenerationParams) usecase.GenerationParams {
	if params == nil {
		return usecase.GenerationParams{}
	}

	return usecase.GenerationParams{
		Temperature:      float32FromWrapper(params.GetTemperature()),
		TopP:             float32FromWrapper(params.GetTopP()),
		MaxTokens:        int32FromWrapper(params.GetMaxTokens()),
		PresencePenalty:  float32FromWrapper(params.GetPresencePenalty()),
		FrequencyPenalty: float32FromWrapper(params.GetFrequencyPenalty()),
	}
}

func float32FromWrapper(value interface{ GetValue() float32 }) *float32 {
	if value == nil {
		return nil
	}
	v := value.GetValue()

	return &v
}

func int32FromWrapper(value interface{ GetValue() int32 }) *int32 {
	if value == nil {
		return nil
	}
	v := value.GetValue()

	return &v
}

func taskTypeFromProto(taskType aiinferencev1.TaskType) (config.TaskType, error) {
	switch taskType {
	case aiinferencev1.TaskType_TASK_TYPE_UNSPECIFIED:
		return "", nil
	case aiinferencev1.TaskType_TASK_TYPE_GENERATION:
		return config.TaskGeneration, nil
	case aiinferencev1.TaskType_TASK_TYPE_EMBEDDING:
		return config.TaskEmbedding, nil
	default:
		return "", usecase.ErrInvalidArgument
	}
}

func taskTypeToProto(task config.TaskType) aiinferencev1.TaskType {
	switch task {
	case config.TaskGeneration:
		return aiinferencev1.TaskType_TASK_TYPE_GENERATION
	case config.TaskEmbedding:
		return aiinferencev1.TaskType_TASK_TYPE_EMBEDDING
	default:
		return aiinferencev1.TaskType_TASK_TYPE_UNSPECIFIED
	}
}
