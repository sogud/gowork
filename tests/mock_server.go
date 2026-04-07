// Package tests provides end-to-end testing utilities.
package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// MockOllamaServer provides a mock Ollama server for testing.
// It simulates the Ollama API endpoints without requiring a real Ollama instance.
type MockOllamaServer struct {
	server   *httptest.Server
	baseURL  string
	mu       sync.Mutex
	requests []MockRequest // Recorded requests for verification
	responses MockResponses // Configurable responses
}

// MockRequest records a single request to the mock server.
type MockRequest struct {
	Method string
	Path   string
	Body   string
}

// MockResponses holds configurable responses for different endpoints.
type MockResponses struct {
	// ChatResponse is the response for /api/chat endpoint
	ChatResponse string

	// TagsResponse is the list of available models
	TagsResponse []string

	// StreamResponses is a list of streaming responses for /api/chat
	StreamResponses []string

	// ShouldError causes the server to return errors
	ShouldError bool

	// ErrorCode is the HTTP error code to return
	ErrorCode int
}

// DefaultMockResponses returns sensible default responses for testing.
func DefaultMockResponses() MockResponses {
	return MockResponses{
		ChatResponse: "This is a mock response from the agent.",
		TagsResponse: []string{"gemma4:4b", "llama3:8b", "mistral:7b"},
		StreamResponses: []string{
			"This ", "is ", "a ", "mock ", "streaming ", "response.",
		},
		ShouldError: false,
		ErrorCode:   500,
	}
}

// NewMockOllamaServer creates a new mock Ollama server with default responses.
func NewMockOllamaServer() *MockOllamaServer {
	return NewMockOllamaServerWithResponses(DefaultMockResponses())
}

// NewMockOllamaServerWithResponses creates a mock server with custom responses.
func NewMockOllamaServerWithResponses(responses MockResponses) *MockOllamaServer {
	mock := &MockOllamaServer{
		responses: responses,
		requests:  make([]MockRequest, 0),
	}

	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))
	mock.baseURL = mock.server.URL

	return mock
}

// Close shuts down the mock server.
func (m *MockOllamaServer) Close() {
	m.server.Close()
}

// BaseURL returns the base URL of the mock server.
func (m *MockOllamaServer) BaseURL() string {
	return m.baseURL
}

// GetRequests returns all recorded requests.
func (m *MockOllamaServer) GetRequests() []MockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests
}

// SetChatResponse updates the chat response.
func (m *MockOllamaServer) SetChatResponse(response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses.ChatResponse = response
}

// SetError configures the server to return errors.
func (m *MockOllamaServer) SetError(shouldError bool, errorCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses.ShouldError = shouldError
	m.responses.ErrorCode = errorCode
}

// handleRequest handles incoming HTTP requests.
func (m *MockOllamaServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record request
	body := ""
	if r.Body != nil {
		bodyBytes, _ := json.Marshal(r.Body)
		body = string(bodyBytes)
	}
	m.requests = append(m.requests, MockRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   body,
	})

	// Handle error mode
	if m.responses.ShouldError {
		http.Error(w, "mock error", m.responses.ErrorCode)
		return
	}

	switch r.URL.Path {
	case "/api/tags":
		m.handleTags(w, r)
	case "/api/chat":
		m.handleChat(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleTags handles the /api/tags endpoint.
func (m *MockOllamaServer) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := make([]struct {
		Name string `json:"name"`
	}, len(m.responses.TagsResponse))
	for i, name := range m.responses.TagsResponse {
		models[i].Name = name
	}

	response := struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}{
		Models: models,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleChat handles the /api/chat endpoint.
func (m *MockOllamaServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request to check stream mode
	var req struct {
		Model    string `json:"model"`
		Stream   bool   `json:"stream"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if req.Stream {
		// Streaming response
		for i, chunk := range m.responses.StreamResponses {
			done := i == len(m.responses.StreamResponses)-1
			resp := ollamaChatResponse{
				Model: req.Model,
				Message: ollamaMessage{
					Role:    "assistant",
					Content: chunk,
				},
				Done: done,
			}
			data, _ := json.Marshal(resp)
			w.Write(data)
			w.Write([]byte("\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	} else {
		// Non-streaming response
		resp := ollamaChatResponse{
			Model: req.Model,
			Message: ollamaMessage{
				Role:    "assistant",
				Content: m.responses.ChatResponse,
			},
			Done: true,
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// ollamaMessage represents a message in Ollama format.
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatResponse represents a chat response in Ollama format.
type ollamaChatResponse struct {
	Model     string        `json:"model"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
	CreatedAt string        `json:"created_at,omitempty"`
}

// MockAgentOutput creates predictable agent outputs for testing.
// Each agent type has a specific output pattern.
func MockAgentOutput(agentName string, input string) string {
	switch agentName {
	case "researcher":
		return fmt.Sprintf("Research findings for: %s. Key information gathered.", input)
	case "analyst":
		return fmt.Sprintf("Analysis of: %s. Insights and patterns identified.", input)
	case "writer":
		return fmt.Sprintf("Written content based on: %s. Clear and structured output.", input)
	case "reviewer":
		return fmt.Sprintf("Review of: %s. Quality check completed.", input)
	case "coordinator":
		return fmt.Sprintf("Coordination result for: %s. Task delegated and tracked.", input)
	default:
		return fmt.Sprintf("Output from %s for input: %s", agentName, input)
	}
}

// CreateTestWorkflowConfig creates a workflow configuration for testing.
func CreateTestWorkflowConfig(workflowType string, agents []string, task string) map[string]interface{} {
	return map[string]interface{}{
		"type":     workflowType,
		"agents":   agents,
		"task":     task,
		"max_iter": 3,
	}
}

// VerifySequentialOrder checks that agents were executed in the expected order.
func VerifySequentialOrder(results []interface{}, expectedAgents []string) bool {
	if len(results) != len(expectedAgents) {
		return false
	}

	for i, expected := range expectedAgents {
		resultMap, ok := results[i].(map[string]interface{})
		if !ok {
			return false
		}
		agentName, ok := resultMap["agent_name"].(string)
		if !ok {
			return false
		}
		if agentName != expected {
			return false
		}
	}
	return true
}

// ContainsSubstring checks if a string contains a substring.
func ContainsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}