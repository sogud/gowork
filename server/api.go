package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sogud/gowork/agents"
	"github.com/sogud/gowork/workflow"
)

// APIServer provides HTTP API endpoints for the workflow system.
type APIServer struct {
	config   Config
	server   *http.Server
	registry *agents.Registry
}

// NewAPIServer creates a new API server with the given configuration.
//
// Parameters:
//   - registry: Agent registry for workflow execution. If nil, placeholder execution is used.
//   - opts: Optional configuration options.
//
// Returns:
//   - *APIServer: The created server
//   - error: An error if configuration is invalid
func NewAPIServer(registry *agents.Registry, opts ...Option) (*APIServer, error) {
	config := DefaultConfig()
	config.Apply(opts...)

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &APIServer{
		config:   config,
		registry: registry,
	}, nil
}

// Start starts the HTTP server.
// This is a blocking call that runs until the server is shut down.
func (s *APIServer) Start() error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *APIServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// registerRoutes registers all API routes.
func (s *APIServer) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/agents", s.handleListAgents)
	mux.HandleFunc("POST /api/v1/workflow/execute", s.handleExecuteWorkflow)
}

// handleHealth handles GET /api/v1/health
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "healthy",
	}
	writeJSON(w, http.StatusOK, response)
}

// handleListAgents handles GET /api/v1/agents
func (s *APIServer) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agentNames := s.getAgentNames()
	response := AgentsResponse{
		Agents: agentNames,
	}
	writeJSON(w, http.StatusOK, response)
}

// handleExecuteWorkflow handles POST /api/v1/workflow/execute
func (s *APIServer) handleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req WorkflowExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create workflow config
	config := &workflow.Config{
		Type:    workflow.ParseType(req.Type),
		Agents:  req.Agents,
		Task:    req.Task,
		MaxIter: req.MaxIter,
	}

	// Create engine options
	var opts []workflow.EngineOption
	if s.registry != nil {
		opts = append(opts, workflow.WithRegistry(s.registry))
	}

	// Create workflow engine
	engine, err := workflow.NewEngine(config, opts...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create workflow engine: %v", err))
		return
	}

	// Execute workflow
	startTime := time.Now()
	result, err := engine.Execute(r.Context())
	executionTime := time.Since(startTime).Milliseconds()

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("workflow execution failed: %v", err))
		return
	}

	// Build response
	response := s.buildWorkflowResponse(result, executionTime)
	writeJSON(w, http.StatusOK, response)
}

// getAgentNames returns the list of available agent names.
func (s *APIServer) getAgentNames() []string {
	if s.registry == nil {
		return []string{"researcher", "analyst", "writer", "reviewer", "coordinator"}
	}
	return s.registry.List()
}

// buildWorkflowResponse builds the workflow response from the result.
func (s *APIServer) buildWorkflowResponse(result *workflow.Result, executionTimeMs int64) WorkflowExecuteResponse {
	// Generate workflow ID (in production, use UUID)
	workflowID := fmt.Sprintf("wf-%d", time.Now().UnixNano())

	// Determine status
	status := "completed"
	if !result.Success {
		status = "failed"
	}

	// Convert agent results to map
	agentResults := make(map[string]interface{})
	for _, ar := range result.AgentResults {
		agentResults[ar.AgentName] = map[string]interface{}{
			"output": ar.Output,
		}
		if ar.Error != nil {
			agentResults[ar.AgentName].(map[string]interface{})["error"] = ar.Error.Error()
		}
	}

	return WorkflowExecuteResponse{
		WorkflowID:      workflowID,
		Status:          status,
		AgentResults:    agentResults,
		ExecutionTimeMs: executionTimeMs,
		Output:          result.Output,
		Iterations:      result.Iterations,
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// Request/Response types

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// AgentsResponse represents the list agents response.
type AgentsResponse struct {
	Agents []string `json:"agents"`
}

// WorkflowExecuteRequest represents a workflow execution request.
type WorkflowExecuteRequest struct {
	Task    string   `json:"task"`
	Type    string   `json:"type"`
	Agents  []string `json:"agents"`
	MaxIter int      `json:"max_iter,omitempty"`
}

// Validate validates the workflow execute request.
func (r *WorkflowExecuteRequest) Validate() error {
	if r.Task == "" {
		return &ValidationError{Field: "task", Message: "task is required"}
	}
	if len(r.Agents) == 0 {
		return &ValidationError{Field: "agents", Message: "at least one agent is required"}
	}
	if r.Type == "" {
		r.Type = "sequential" // default
	}
	// Validate workflow type
	switch r.Type {
	case "sequential", "parallel", "loop", "dynamic":
		// valid
	default:
		return &ValidationError{Field: "type", Message: "invalid workflow type, must be one of: sequential, parallel, loop, dynamic"}
	}
	return nil
}

// WorkflowExecuteResponse represents a workflow execution response.
type WorkflowExecuteResponse struct {
	WorkflowID      string                 `json:"workflow_id"`
	Status          string                 `json:"status"`
	AgentResults    map[string]interface{} `json:"agent_results"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	Output          string                 `json:"output,omitempty"`
	Iterations      int                    `json:"iterations,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}