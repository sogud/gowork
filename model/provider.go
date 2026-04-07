// Package model provides model providers for LLM integration.
package model

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// ProviderType defines the type of model provider.
type ProviderType string

const (
	ProviderOllama ProviderType = "ollama"
	ProviderGemini ProviderType = "gemini"
	ProviderApigee ProviderType = "apigee"
)

// ProviderConfig holds configuration for model providers.
type ProviderConfig struct {
	Type     ProviderType
	Model    string
	BaseURL  string // For Ollama
	APIKey   string // For Gemini
	Timeout  int
	MaxRetry int

	// Gemini specific
	GeminiConfig *genai.ClientConfig

	// Custom provider
	CustomModel adkmodel.LLM
}

// NewModel creates a model provider based on configuration.
func NewModel(ctx context.Context, cfg *ProviderConfig) (adkmodel.LLM, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider config is nil")
	}

	// If custom model is provided, use it directly
	if cfg.CustomModel != nil {
		return cfg.CustomModel, nil
	}

	switch cfg.Type {
	case ProviderOllama:
		return newOllamaModel(ctx, cfg)

	case ProviderGemini:
		return newGeminiModel(ctx, cfg)

	case ProviderApigee:
		return nil, fmt.Errorf("apigee provider not yet implemented")

	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}
}

// newOllamaModel creates an Ollama model provider.
func newOllamaModel(ctx context.Context, cfg *ProviderConfig) (adkmodel.LLM, error) {
	ollamaCfg := &Config{
		BaseURL:   cfg.BaseURL,
		ModelName: cfg.Model,
	}

	if ollamaCfg.BaseURL == "" {
		ollamaCfg.BaseURL = "http://localhost:11434"
	}

	return NewOllamaModel(ctx, ollamaCfg)
}

// newGeminiModel creates a Gemini model provider using adk-go's native implementation.
func newGeminiModel(ctx context.Context, cfg *ProviderConfig) (adkmodel.LLM, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Gemini provider")
	}

	geminiCfg := cfg.GeminiConfig
	if geminiCfg == nil {
		geminiCfg = &genai.ClientConfig{
			APIKey: cfg.APIKey,
		}
	}

	return gemini.NewModel(ctx, cfg.Model, geminiCfg)
}

// NewOllamaModelFromConfig creates an Ollama model from simple config.
func NewOllamaModelFromConfig(ctx context.Context, baseURL, model string) (adkmodel.LLM, error) {
	return NewModel(ctx, &ProviderConfig{
		Type:    ProviderOllama,
		Model:   model,
		BaseURL: baseURL,
	})
}

// NewGeminiModelFromAPIKey creates a Gemini model from API key.
func NewGeminiModelFromAPIKey(ctx context.Context, model, apiKey string) (adkmodel.LLM, error) {
	return NewModel(ctx, &ProviderConfig{
		Type:   ProviderGemini,
		Model:  model,
		APIKey: apiKey,
	})
}

// NewCustomModel creates a model from a custom LLM implementation.
func NewCustomModel(customModel adkmodel.LLM) (adkmodel.LLM, error) {
	return NewModel(nil, &ProviderConfig{
		CustomModel: customModel,
	})
}