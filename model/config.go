package model

import (
	"net/http"
	"time"
)

// Config holds configuration for the Ollama model adapter.
type Config struct {
	// BaseURL is the Ollama server URL (default: http://localhost:11434)
	BaseURL string

	// ModelName is the name of the model to use (default: gemma:4b)
	ModelName string

	// HTTPClient is the HTTP client for making requests
	HTTPClient *http.Client

	// Timeout is the request timeout (default: 30s)
	Timeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:   "http://localhost:11434",
		ModelName: "gemma4:e4b",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}