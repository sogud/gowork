package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConfig_DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Host != "0.0.0.0" {
		t.Errorf("expected host '0.0.0.0', got '%s'", config.Host)
	}
	if config.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Port)
	}
	if config.ReadTimeout != 15*time.Second {
		t.Errorf("expected read timeout 15s, got %v", config.ReadTimeout)
	}
	if config.WriteTimeout != 15*time.Second {
		t.Errorf("expected write timeout 15s, got %v", config.WriteTimeout)
	}
	if config.IdleTimeout != 60*time.Second {
		t.Errorf("expected idle timeout 60s, got %v", config.IdleTimeout)
	}
	if config.ShutdownTimeout != 30*time.Second {
		t.Errorf("expected shutdown timeout 30s, got %v", config.ShutdownTimeout)
	}
}

func TestConfig_Options(t *testing.T) {
	config := DefaultConfig()
	config.Apply(
		WithHost("localhost"),
		WithPort(9090),
		WithReadTimeout(10*time.Second),
		WithWriteTimeout(20*time.Second),
		WithIdleTimeout(30*time.Second),
		WithShutdownTimeout(15*time.Second),
	)

	if config.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", config.Host)
	}
	if config.Port != 9090 {
		t.Errorf("expected port 9090, got %d", config.Port)
	}
	if config.ReadTimeout != 10*time.Second {
		t.Errorf("expected read timeout 10s, got %v", config.ReadTimeout)
	}
	if config.WriteTimeout != 20*time.Second {
		t.Errorf("expected write timeout 20s, got %v", config.WriteTimeout)
	}
	if config.IdleTimeout != 30*time.Second {
		t.Errorf("expected idle timeout 30s, got %v", config.IdleTimeout)
	}
	if config.ShutdownTimeout != 15*time.Second {
		t.Errorf("expected shutdown timeout 15s, got %v", config.ShutdownTimeout)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "invalid port - zero",
			config: Config{
				Host:            "localhost",
				Port:            0,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid port - negative",
			config: Config{
				Host:            "localhost",
				Port:            -1,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Host:            "localhost",
				Port:            70000,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "empty host",
			config: Config{
				Host:            "",
				Port:            8080,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid read timeout",
			config: Config{
				Host:            "localhost",
				Port:            8080,
				ReadTimeout:     0,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid write timeout",
			config: Config{
				Host:            "localhost",
				Port:            8080,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    0,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid idle timeout",
			config: Config{
				Host:            "localhost",
				Port:            8080,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     0,
				ShutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid shutdown timeout",
			config: Config{
				Host:            "localhost",
				Port:            8080,
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewAPIServer(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		wantError bool
	}{
		{
			name:      "default config",
			opts:      nil,
			wantError: false,
		},
		{
			name:      "custom port",
			opts:      []Option{WithPort(9090)},
			wantError: false,
		},
		{
			name:      "invalid config",
			opts:      []Option{WithPort(-1)},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewAPIServer(nil, tt.opts...)
			if (err != nil) != tt.wantError {
				t.Errorf("NewAPIServer() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && server == nil {
				t.Error("expected server to be created")
			}
		})
	}
}

func TestHandleHealth(t *testing.T) {
	server, err := NewAPIServer(nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}
}

func TestHandleListAgents(t *testing.T) {
	server, err := NewAPIServer(nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	rec := httptest.NewRecorder()

	server.handleListAgents(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response AgentsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Agents) == 0 {
		t.Error("expected at least one agent")
	}

	// Check default agents exist
	agentMap := make(map[string]bool)
	for _, agent := range response.Agents {
		agentMap[agent] = true
	}

	expectedAgents := []string{"researcher", "analyst", "writer", "reviewer", "coordinator"}
	for _, expected := range expectedAgents {
		if !agentMap[expected] {
			t.Errorf("expected agent '%s' not found", expected)
		}
	}
}

func TestHandleExecuteWorkflow(t *testing.T) {
	server, err := NewAPIServer(nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name       string
		request    WorkflowExecuteRequest
		wantStatus int
	}{
		{
			name: "valid sequential workflow",
			request: WorkflowExecuteRequest{
				Task:   "研究 Go 并发模式",
				Type:   "sequential",
				Agents: []string{"researcher", "analyst"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid parallel workflow",
			request: WorkflowExecuteRequest{
				Task:   "分析代码质量",
				Type:   "parallel",
				Agents: []string{"analyst", "reviewer"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid loop workflow with max_iter",
			request: WorkflowExecuteRequest{
				Task:    "迭代优化文档",
				Type:    "loop",
				Agents:  []string{"writer", "reviewer"},
				MaxIter: 3,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "default type to sequential",
			request: WorkflowExecuteRequest{
				Task:   "默认类型测试",
				Agents: []string{"researcher"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "missing task",
			request: WorkflowExecuteRequest{
				Type:   "sequential",
				Agents: []string{"researcher"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing agents",
			request: WorkflowExecuteRequest{
				Task: "测试任务",
				Type: "sequential",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid workflow type",
			request: WorkflowExecuteRequest{
				Task:   "测试任务",
				Type:   "invalid",
				Agents: []string{"researcher"},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/workflow/execute", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.handleExecuteWorkflow(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d, body: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				var response WorkflowExecuteResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.WorkflowID == "" {
					t.Error("expected workflow_id to be set")
				}
				if response.Status == "" {
					t.Error("expected status to be set")
				}
				if response.ExecutionTimeMs < 0 {
					t.Error("expected execution_time_ms to be non-negative")
				}
			}
		})
	}
}

func TestHandleExecuteWorkflow_InvalidJSON(t *testing.T) {
	server, err := NewAPIServer(nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflow/execute", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleExecuteWorkflow(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestWorkflowExecuteRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		request   WorkflowExecuteRequest
		wantError bool
	}{
		{
			name: "valid request",
			request: WorkflowExecuteRequest{
				Task:   "test task",
				Type:   "sequential",
				Agents: []string{"researcher"},
			},
			wantError: false,
		},
		{
			name: "empty task",
			request: WorkflowExecuteRequest{
				Type:   "sequential",
				Agents: []string{"researcher"},
			},
			wantError: true,
		},
		{
			name: "empty agents",
			request: WorkflowExecuteRequest{
				Task: "test task",
				Type: "sequential",
			},
			wantError: true,
		},
		{
			name: "invalid type",
			request: WorkflowExecuteRequest{
				Task:   "test task",
				Type:   "invalid",
				Agents: []string{"researcher"},
			},
			wantError: true,
		},
		{
			name: "default type when empty",
			request: WorkflowExecuteRequest{
				Task:   "test task",
				Agents: []string{"researcher"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}

			// Check default type
			if !tt.wantError && tt.request.Type == "" {
				t.Error("expected type to be defaulted to 'sequential'")
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "task", Message: "is required"}
	expected := "task: is required"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{Field: "Port", Message: "must be valid"}
	expected := "config error: Port - must be valid"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestAPIServer_Routes(t *testing.T) {
	server, err := NewAPIServer(nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Create test server using the actual handler
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health":
			server.handleHealth(w, r)
		case "/api/v1/agents":
			server.handleListAgents(w, r)
		case "/api/v1/workflow/execute":
			server.handleExecuteWorkflow(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	t.Run("health endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/health")
		if err != nil {
			t.Fatalf("failed to request health endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("agents endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/agents")
		if err != nil {
			t.Fatalf("failed to request agents endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("workflow execute endpoint", func(t *testing.T) {
		body, _ := json.Marshal(WorkflowExecuteRequest{
			Task:   "test task",
			Type:   "sequential",
			Agents: []string{"researcher"},
		})

		resp, err := http.Post(ts.URL+"/api/v1/workflow/execute", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to request workflow execute endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestAPIServer_Shutdown(t *testing.T) {
	server, err := NewAPIServer(nil, WithPort(18080))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Start server in goroutine
	go func() {
		server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}