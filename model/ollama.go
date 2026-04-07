// Package model implements the model.LLM interface for Ollama.
package model

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"

	adkmodel "google.golang.org/adk/model"
	"google.golang.org/genai"
)

// ollamaModel implements model.LLM for Ollama backend.
type ollamaModel struct {
	name    string
	baseURL string
	client  *http.Client
}

// ollamaChatRequest represents a request to Ollama /api/chat endpoint.
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// ollamaMessage represents a single message in the chat.
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatResponse represents a response from Ollama /api/chat endpoint.
type ollamaChatResponse struct {
	Model     string        `json:"model"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
	CreatedAt string        `json:"created_at,omitempty"`
}

// ollamaTagsResponse represents the response from /api/tags endpoint.
type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// NewOllamaModel creates a new Ollama model adapter.
// It performs a health check to verify the Ollama server is running
// and the specified model is available.
func NewOllamaModel(ctx context.Context, cfg *Config) (adkmodel.LLM, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Use default HTTPClient if not provided
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	// Health check: verify server is running and model is available
	if err := healthCheck(ctx, cfg); err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}, nil
}

// healthCheck verifies the Ollama server is accessible and the model is available.
func healthCheck(ctx context.Context, cfg *Config) error {
	url := fmt.Sprintf("%s/api/tags", cfg.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama server at %s: %w", cfg.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	var tagsResp ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return fmt.Errorf("failed to decode tags response: %w", err)
	}

	// Check if the requested model is available
	modelAvailable := false
	for _, model := range tagsResp.Models {
		if model.Name == cfg.ModelName || strings.HasPrefix(model.Name, cfg.ModelName+":") {
			modelAvailable = true
			break
		}
	}

	if !modelAvailable {
		return fmt.Errorf("model %q not found in Ollama server. Available models: %v",
			cfg.ModelName, getModelNames(tagsResp.Models))
	}

	return nil
}

// Name returns the model name.
func (m *ollamaModel) Name() string {
	return m.name
}

// GenerateContent generates content from the model.
// Supports both streaming and non-streaming modes.
func (m *ollamaModel) GenerateContent(ctx context.Context, req *adkmodel.LLMRequest, stream bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		// Handle nil request
		if req == nil {
			yield(nil, fmt.Errorf("request cannot be nil"))
			return
		}

		ollamaReq := convertRequest(req)
		ollamaReq.Stream = stream

		if stream {
			m.generateStream(ctx, ollamaReq, yield)
		} else {
			m.generateNonStream(ctx, ollamaReq, yield)
		}
	}
}

// generateNonStream handles non-streaming generation.
func (m *ollamaModel) generateNonStream(ctx context.Context, req *ollamaChatRequest, yield func(*adkmodel.LLMResponse, error) bool) {
	resp, err := m.doChatRequest(ctx, req)
	if err != nil {
		yield(nil, err)
		return
	}

	yield(convertResponse(resp), nil)
}

// generateStream handles streaming generation from Ollama.
func (m *ollamaModel) generateStream(ctx context.Context, req *ollamaChatRequest, yield func(*adkmodel.LLMResponse, error) bool) {
	resp, err := m.doChatRequestStream(ctx, req)
	if err != nil {
		yield(nil, err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue // Skip empty lines
		}

		var ollamaResp ollamaChatResponse
		if err := json.Unmarshal(line, &ollamaResp); err != nil {
			yield(nil, fmt.Errorf("failed to unmarshal stream chunk: %w", err))
			return
		}

		llmResp := convertResponse(&ollamaResp)
		// Set Partial based on done status
		llmResp.Partial = !ollamaResp.Done

		if !yield(llmResp, nil) {
			// Consumer stopped early, stop producing
			return
		}
	}

	if err := scanner.Err(); err != nil {
		yield(nil, fmt.Errorf("error reading stream: %w", err))
	}
}

// doChatRequest performs the HTTP request to Ollama /api/chat endpoint.
func (m *ollamaModel) doChatRequest(ctx context.Context, req *ollamaChatRequest) (*ollamaChatResponse, error) {
	url := fmt.Sprintf("%s/api/chat", m.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body for more context
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(errorBody))
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ollamaResp, nil
}

// doChatRequestStream performs the HTTP request to Ollama /api/chat endpoint for streaming.
// It returns the raw HTTP response for the caller to read the stream.
func (m *ollamaModel) doChatRequestStream(ctx context.Context, req *ollamaChatRequest) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/chat", m.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Read error response body for more context
		errorBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(errorBody))
	}

	return resp, nil
}

// convertRequest converts LLMRequest to Ollama chat request format.
// It preserves role information for proper multi-turn dialogue support.
func convertRequest(req *adkmodel.LLMRequest) *ollamaChatRequest {
	if req == nil {
		return &ollamaChatRequest{}
	}

	messages := make([]ollamaMessage, 0, len(req.Contents))
	for _, content := range req.Contents {
		// Build text content from all parts
		var textBuilder strings.Builder
		for _, part := range content.Parts {
			textBuilder.WriteString(part.Text)
		}

		// Map genai roles to Ollama roles
		role := content.Role
		if role == "model" {
			role = "assistant"
		}

		messages = append(messages, ollamaMessage{
			Role:    role,
			Content: textBuilder.String(),
		})
	}

	return &ollamaChatRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   false,
	}
}

// convertResponse converts Ollama chat response to LLMResponse format.
func convertResponse(resp *ollamaChatResponse) *adkmodel.LLMResponse {
	if resp == nil {
		return &adkmodel.LLMResponse{}
	}

	return &adkmodel.LLMResponse{
		Content: &genai.Content{
			Role:  "model",
			Parts: []*genai.Part{{Text: resp.Message.Content}},
		},
		TurnComplete: resp.Done,
	}
}

// getModelNames extracts model names from ollamaTagsResponse models.
func getModelNames(models []struct {
	Name string `json:"name"`
}) []string {
	names := make([]string, len(models))
	for i, m := range models {
		names[i] = m.Name
	}
	return names
}