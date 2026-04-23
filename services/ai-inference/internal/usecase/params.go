package usecase

import "secure-rag-platform/services/ai-inference/internal/config"

func mergeGenerationParams(defaults config.GenerationDefaults, incoming GenerationParams) GenerationParams {
	merged := GenerationParams{
		Temperature:      cloneFloat32(defaults.Temperature),
		TopP:             cloneFloat32(defaults.TopP),
		MaxTokens:        cloneInt32(defaults.MaxTokens),
		PresencePenalty:  cloneFloat32(defaults.PresencePenalty),
		FrequencyPenalty: cloneFloat32(defaults.FrequencyPenalty),
	}

	if incoming.Temperature != nil {
		merged.Temperature = cloneFloat32(incoming.Temperature)
	}
	if incoming.TopP != nil {
		merged.TopP = cloneFloat32(incoming.TopP)
	}
	if incoming.MaxTokens != nil {
		merged.MaxTokens = cloneInt32(incoming.MaxTokens)
	}
	if incoming.PresencePenalty != nil {
		merged.PresencePenalty = cloneFloat32(incoming.PresencePenalty)
	}
	if incoming.FrequencyPenalty != nil {
		merged.FrequencyPenalty = cloneFloat32(incoming.FrequencyPenalty)
	}

	return merged
}

func cloneFloat32(v *float32) *float32 {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func cloneInt32(v *int32) *int32 {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}
