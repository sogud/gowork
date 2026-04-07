package model

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/genai"
)

func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		if cfg.BaseURL == "" {
			t.Error("BaseURL should not be empty")
		}
		if cfg.ModelName == "" {
			t.Error("ModelName should not be empty")
		}
		if cfg.HTTPClient == nil {
			t.Error("HTTPClient should not be nil")
		}
		if cfg.Timeout == 0 {
			t.Error("Timeout should not be zero")
		}
	})

	t.Run("CustomConfig", func(t *testing.T) {
		cfg := &Config{
			BaseURL:   "http://custom:11434",
			ModelName: "llama2",
		}
		if cfg.BaseURL != "http://custom:11434" {
			t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "http://custom:11434")
		}
		if cfg.ModelName != "llama2" {
			t.Errorf("ModelName = %q, want %q", cfg.ModelName, "llama2")
		}
	})
}

func TestOllamaModel_Name(t *testing.T) {
	cfg := DefaultConfig()
	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	if got := model.Name(); got != cfg.ModelName {
		t.Errorf("Name() = %q, want %q", got, cfg.ModelName)
	}
}

func TestOllamaModel_GenerateContent_NonStreaming(t *testing.T) {
	tests := []struct {
		name       string
		serverResp string
		wantText   string
		wantErr    bool
	}{
		{
			name:       "simple response",
			serverResp: `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Hello, world!"},"done":true}`,
			wantText:   "Hello, world!",
			wantErr:    false,
		},
		{
			name:       "empty response",
			serverResp: `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":""},"done":true}`,
			wantText:   "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/chat" {
					t.Errorf("unexpected path: %s", r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: %s", r.Method)
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			cfg := &Config{
				BaseURL:   server.URL,
				ModelName: "gemma:4b",
				HTTPClient: http.DefaultClient,
			}

			model := &ollamaModel{
				name:    cfg.ModelName,
				baseURL: cfg.BaseURL,
				client:  cfg.HTTPClient,
			}

			req := &adkmodel.LLMRequest{
				Model: cfg.ModelName,
				Contents: []*genai.Content{
					{
						Role:  "user",
						Parts: []*genai.Part{{Text: "Hello"}},
					},
				},
			}

			ctx := context.Background()
			var gotResponse *adkmodel.LLMResponse
			var gotErr error

			for resp, err := range model.GenerateContent(ctx, req, false) {
				gotResponse = resp
				gotErr = err
				break // Non-streaming, only one iteration
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GenerateContent() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if !tt.wantErr && gotResponse != nil {
				if len(gotResponse.Content.Parts) == 0 {
					t.Error("Content.Parts is empty")
					return
				}
				if gotResponse.Content.Parts[0].Text != tt.wantText {
					t.Errorf("Text = %q, want %q", gotResponse.Content.Parts[0].Text, tt.wantText)
				}
			}
		})
	}
}

func TestOllamaModel_GenerateContent_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not be called
		t.Error("unexpected request to server")
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	ctx := context.Background()
	var gotErr error

	for _, err := range model.GenerateContent(ctx, nil, false) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error for nil request, got nil")
	}
	if gotErr != nil && !strings.Contains(gotErr.Error(), "request cannot be nil") {
		t.Errorf("error = %v, want error containing 'request cannot be nil'", gotErr)
	}
}

func TestOllamaModel_GenerateContent_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	req := &adkmodel.LLMRequest{
		Model: cfg.ModelName,
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: "Hello"}},
			},
		},
	}

	ctx := context.Background()
	var gotErr error

	for _, err := range model.GenerateContent(ctx, req, false) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error for API error, got nil")
	}
	if gotErr != nil && !strings.Contains(gotErr.Error(), "500") {
		t.Errorf("error = %v, want error containing '500'", gotErr)
	}
	if gotErr != nil && !strings.Contains(gotErr.Error(), "model not found") {
		t.Errorf("error = %v, want error containing 'model not found'", gotErr)
	}
}

func TestOllamaModel_GenerateContent_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	req := &adkmodel.LLMRequest{
		Model: cfg.ModelName,
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: "Hello"}},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var gotErr error
	for _, err := range model.GenerateContent(ctx, req, false) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error for context cancellation, got nil")
	}
}

func TestOllamaModel_GenerateContent_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	req := &adkmodel.LLMRequest{
		Model: cfg.ModelName,
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: "Hello"}},
			},
		},
	}

	ctx := context.Background()
	var gotErr error

	for _, err := range model.GenerateContent(ctx, req, false) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if gotErr != nil && !strings.Contains(gotErr.Error(), "decode") {
		t.Errorf("error = %v, want error containing 'decode'", gotErr)
	}
}

func TestOllamaModel_HealthCheck(t *testing.T) {
	t.Run("healthy server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/tags" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"models":[{"name":"gemma:4b"}]}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		cfg := &Config{
			BaseURL:   server.URL,
			ModelName: "gemma:4b",
			HTTPClient: http.DefaultClient,
		}

		ctx := context.Background()
		model, err := NewOllamaModel(ctx, cfg)
		if err != nil {
			t.Fatalf("NewOllamaModel() error = %v", err)
		}

		if model == nil {
			t.Error("model should not be nil")
		}
	})

	t.Run("unhealthy server - wrong model", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/tags" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"models":[{"name":"llama2"}]}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		cfg := &Config{
			BaseURL:   server.URL,
			ModelName: "gemma:4b",
			HTTPClient: http.DefaultClient,
		}

		ctx := context.Background()
		_, err := NewOllamaModel(ctx, cfg)
		if err == nil {
			t.Error("expected error for unavailable model, got nil")
		}
	})

	t.Run("unhealthy server - connection refused", func(t *testing.T) {
		cfg := &Config{
			BaseURL:   "http://localhost:59999", // Non-existent server
			ModelName: "gemma:4b",
			HTTPClient: http.DefaultClient,
		}

		ctx := context.Background()
		_, err := NewOllamaModel(ctx, cfg)
		if err == nil {
			t.Error("expected error for connection refused, got nil")
		}
	})
}

func TestNewOllamaModel_NilHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[{"name":"gemma:4b"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: nil, // nil HTTPClient
		Timeout:   30 * time.Second,
	}

	ctx := context.Background()
	model, err := NewOllamaModel(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOllamaModel() error = %v", err)
	}

	if model == nil {
		t.Error("model should not be nil")
	}

	// Verify the model can make requests
	ollama := model.(*ollamaModel)
	if ollama.client == nil {
		t.Error("ollamaModel.client should not be nil after NewOllamaModel")
	}
}

func TestConvertRequest(t *testing.T) {
	tests := []struct {
		name  string
		input *adkmodel.LLMRequest
		want  *ollamaChatRequest
	}{
		{
			name: "simple text request",
			input: &adkmodel.LLMRequest{
				Model: "gemma:4b",
				Contents: []*genai.Content{
					{
						Role:  "user",
						Parts: []*genai.Part{{Text: "Hello"}},
					},
				},
			},
			want: &ollamaChatRequest{
				Model: "gemma:4b",
				Messages: []ollamaMessage{
					{Role: "user", Content: "Hello"},
				},
				Stream: false,
			},
		},
		{
			name: "multi-turn conversation preserves roles",
			input: &adkmodel.LLMRequest{
				Model: "gemma:4b",
				Contents: []*genai.Content{
					{
						Role:  "user",
						Parts: []*genai.Part{{Text: "Hi"}},
					},
					{
						Role:  "model",
						Parts: []*genai.Part{{Text: "Hello!"}},
					},
					{
						Role:  "user",
						Parts: []*genai.Part{{Text: "How are you?"}},
					},
				},
			},
			want: &ollamaChatRequest{
				Model: "gemma:4b",
				Messages: []ollamaMessage{
					{Role: "user", Content: "Hi"},
					{Role: "assistant", Content: "Hello!"},
					{Role: "user", Content: "How are you?"},
				},
				Stream: false,
			},
		},
		{
			name: "nil request",
			input: nil,
			want:  &ollamaChatRequest{},
		},
		{
			name: "empty contents",
			input: &adkmodel.LLMRequest{
				Model:    "gemma:4b",
				Contents: []*genai.Content{},
			},
			want: &ollamaChatRequest{
				Model:    "gemma:4b",
				Messages: []ollamaMessage{},
				Stream:   false,
			},
		},
		{
			name: "multiple parts in content",
			input: &adkmodel.LLMRequest{
				Model: "gemma:4b",
				Contents: []*genai.Content{
					{
						Role:  "user",
						Parts: []*genai.Part{{Text: "Hello "}, {Text: "World"}},
					},
				},
			},
			want: &ollamaChatRequest{
				Model: "gemma:4b",
				Messages: []ollamaMessage{
					{Role: "user", Content: "Hello World"},
				},
				Stream: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertRequest(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("convertRequest() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConvertResponse(t *testing.T) {
	tests := []struct {
		name  string
		input *ollamaChatResponse
		want  *adkmodel.LLMResponse
	}{
		{
			name: "simple response",
			input: &ollamaChatResponse{
				Model: "gemma:4b",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "Hello, world!",
				},
				Done: true,
			},
			want: &adkmodel.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{{Text: "Hello, world!"}},
				},
				TurnComplete: true,
			},
		},
		{
			name:  "nil response",
			input: nil,
			want:  &adkmodel.LLMResponse{},
		},
		{
			name: "empty content",
			input: &ollamaChatResponse{
				Model: "gemma:4b",
				Message: ollamaMessage{
					Role:    "assistant",
					Content: "",
				},
				Done: true,
			},
			want: &adkmodel.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{{Text: ""}},
				},
				TurnComplete: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertResponse(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("convertResponse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDoChatRequest_ReadsErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), "messages") {
			t.Errorf("request body should contain 'messages', got: %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid request: model not loaded"}`))
	}))
	defer server.Close()

	model := &ollamaModel{
		name:    "gemma:4b",
		baseURL: server.URL,
		client:  http.DefaultClient,
	}

	req := &ollamaChatRequest{
		Model: "gemma:4b",
		Messages: []ollamaMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := model.doChatRequest(context.Background(), req)
	if err == nil {
		t.Error("expected error for bad request, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should contain status code 400, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid request: model not loaded") {
		t.Errorf("error should contain error body, got: %v", err)
	}
}

func TestOllamaModel_GenerateContent_Streaming(t *testing.T) {
	tests := []struct {
		name        string
		streamData  string
		wantChunks  []string
		wantDones   []bool
		wantErr     bool
		errContains string
	}{
		{
			name: "simple streaming response",
			streamData: `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Hello"},"done":false}
{"model":"gemma:4b","created_at":"2024-01-01T00:00:01Z","message":{"role":"assistant","content":" world"},"done":false}
{"model":"gemma:4b","created_at":"2024-01-01T00:00:02Z","message":{"role":"assistant","content":"!"},"done":true}`,
			wantChunks: []string{"Hello", " world", "!"},
			wantDones:  []bool{false, false, true},
			wantErr:    false,
		},
		{
			name: "single chunk response",
			streamData: `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Complete response"},"done":true}`,
			wantChunks: []string{"Complete response"},
			wantDones:  []bool{true},
			wantErr:    false,
		},
		{
			name: "empty chunks in stream",
			streamData: `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":""},"done":false}
{"model":"gemma:4b","created_at":"2024-01-01T00:00:01Z","message":{"role":"assistant","content":"text"},"done":false}
{"model":"gemma:4b","created_at":"2024-01-01T00:00:02Z","message":{"role":"assistant","content":""},"done":true}`,
			wantChunks: []string{"", "text", ""},
			wantDones:  []bool{false, false, true},
			wantErr:    false,
		},
		{
			name: "invalid JSON in stream",
			streamData: `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Hello"},"done":false}
invalid json line
{"model":"gemma:4b","created_at":"2024-01-01T00:00:02Z","message":{"role":"assistant","content":"!"},"done":true}`,
			wantErr:     true,
			errContains: "failed to unmarshal stream chunk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/chat" {
					t.Errorf("unexpected path: %s", r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: %s", r.Method)
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}

				// Verify stream parameter is set
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("failed to read request body: %v", err)
				}
				if !strings.Contains(string(body), `"stream":true`) {
					t.Errorf("request body should contain stream:true, got: %s", string(body))
				}

				w.Header().Set("Content-Type", "application/x-ndjson")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.streamData))
			}))
			defer server.Close()

			cfg := &Config{
				BaseURL:   server.URL,
				ModelName: "gemma:4b",
				HTTPClient: http.DefaultClient,
			}

			model := &ollamaModel{
				name:    cfg.ModelName,
				baseURL: cfg.BaseURL,
				client:  cfg.HTTPClient,
			}

			req := &adkmodel.LLMRequest{
				Model: cfg.ModelName,
				Contents: []*genai.Content{
					{
						Role:  "user",
						Parts: []*genai.Part{{Text: "Hello"}},
					},
				},
			}

			ctx := context.Background()
			var gotChunks []string
			var gotDones []bool
			var gotErr error
			chunkCount := 0

			for resp, err := range model.GenerateContent(ctx, req, true) {
				if err != nil {
					gotErr = err
					break
				}
				if resp != nil && len(resp.Content.Parts) > 0 {
					gotChunks = append(gotChunks, resp.Content.Parts[0].Text)
					gotDones = append(gotDones, resp.TurnComplete)
					chunkCount++
				}
			}

			if tt.wantErr {
				if gotErr == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(gotErr.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", gotErr, tt.errContains)
				}
				return
			}

			if gotErr != nil {
				t.Errorf("unexpected error: %v", gotErr)
				return
			}

			if len(gotChunks) != len(tt.wantChunks) {
				t.Errorf("got %d chunks, want %d", len(gotChunks), len(tt.wantChunks))
				return
			}

			for i, want := range tt.wantChunks {
				if gotChunks[i] != want {
					t.Errorf("chunk[%d] = %q, want %q", i, gotChunks[i], want)
				}
			}

			for i, want := range tt.wantDones {
				if gotDones[i] != want {
					t.Errorf("done[%d] = %v, want %v", i, gotDones[i], want)
				}
			}
		})
	}
}

func TestOllamaModel_GenerateContent_Streaming_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not be called
		t.Error("unexpected request to server")
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	ctx := context.Background()
	var gotErr error

	for _, err := range model.GenerateContent(ctx, nil, true) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error for nil request, got nil")
	}
	if gotErr != nil && !strings.Contains(gotErr.Error(), "request cannot be nil") {
		t.Errorf("error = %v, want error containing 'request cannot be nil'", gotErr)
	}
}

func TestOllamaModel_GenerateContent_Streaming_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	req := &adkmodel.LLMRequest{
		Model: cfg.ModelName,
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: "Hello"}},
			},
		},
	}

	ctx := context.Background()
	var gotErr error

	for _, err := range model.GenerateContent(ctx, req, true) {
		gotErr = err
		break
	}

	if gotErr == nil {
		t.Error("expected error for API error, got nil")
	}
	if gotErr != nil && !strings.Contains(gotErr.Error(), "500") {
		t.Errorf("error = %v, want error containing '500'", gotErr)
	}
}

func TestOllamaModel_GenerateContent_Streaming_ConsumerStop(t *testing.T) {
	// Test that the producer stops when consumer stops early
	streamData := `{"model":"gemma:4b","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Hello"},"done":false}
{"model":"gemma:4b","created_at":"2024-01-01T00:00:01Z","message":{"role":"assistant","content":" world"},"done":false}
{"model":"gemma:4b","created_at":"2024-01-01T00:00:02Z","message":{"role":"assistant","content":"!"},"done":true}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(streamData))
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL:   server.URL,
		ModelName: "gemma:4b",
		HTTPClient: http.DefaultClient,
	}

	model := &ollamaModel{
		name:    cfg.ModelName,
		baseURL: cfg.BaseURL,
		client:  cfg.HTTPClient,
	}

	req := &adkmodel.LLMRequest{
		Model: cfg.ModelName,
		Contents: []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: "Hello"}},
			},
		},
	}

	ctx := context.Background()
	chunkCount := 0

	// Only consume first chunk
	for resp, err := range model.GenerateContent(ctx, req, true) {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			break
		}
		if resp != nil {
			chunkCount++
			break // Stop after first chunk
		}
	}

	// Should have received exactly one chunk
	if chunkCount != 1 {
		t.Errorf("expected 1 chunk, got %d", chunkCount)
	}

	// The request should have been made once
	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}
}