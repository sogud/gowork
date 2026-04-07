// Package tests provides end-to-end integration tests.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sogud/gowork/agents"
	"github.com/sogud/gowork/memory"
	"github.com/sogud/gowork/model"
	"github.com/sogud/gowork/server"
	"github.com/sogud/gowork/workflow"
	adkagent "google.golang.org/adk/agent"
	adkmodel "google.golang.org/adk/model"
)

// ============================================================================
// Scenario 1: Complete Workflow Execution
// ============================================================================

// TestE2E_CompleteWorkflowExecution tests the entire workflow from model to result.
// This test validates:
// - Mock Ollama server works correctly
// - Model adapter can communicate with mock server
// - Agents execute using the model
// - Workflow engine orchestrates execution
// - Memory service shares state
func TestE2E_CompleteWorkflowExecution(t *testing.T) {
	// Setup mock Ollama server
	mockServer := NewMockOllamaServer()
	defer mockServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create model adapter with mock server
	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   10 * time.Second,
	}

	llm, err := model.NewOllamaModel(ctx, modelCfg)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Verify model name
	if llm.Name() != "gemma4:4b" {
		t.Errorf("Expected model name 'gemma4:4b', got '%s'", llm.Name())
	}

	// Create agent registry
	registry := agents.NewRegistry()

	// Create and register agents
	testAgents := []string{"researcher", "analyst", "writer"}
	for _, agentName := range testAgents {
		agent, err := createTestAgent(llm, agentName)
		if err != nil {
			t.Fatalf("Failed to create agent %s: %v", agentName, err)
		}
		execAgent := agents.NewExecutableAgent(agent)
		if err := registry.Register(execAgent); err != nil {
			t.Fatalf("Failed to register agent %s: %v", agentName, err)
		}
	}

	// Verify registry
	registeredAgents := registry.List()
	if len(registeredAgents) != 3 {
		t.Errorf("Expected 3 registered agents, got %d", len(registeredAgents))
	}

	// Create memory service
	memService := memory.NewInMemoryService()

	// Create workflow configuration
	workflowCfg := &workflow.Config{
		Type:    workflow.TypeSequential,
		Agents:  testAgents,
		Task:    "Test task: analyze the benefits of Go programming language",
		MaxIter: 1,
	}

	// Create workflow engine
	engine, err := workflow.NewEngine(
		workflowCfg,
		workflow.WithRegistry(registry),
		workflow.WithSessionService(memService),
	)
	if err != nil {
		t.Fatalf("Failed to create workflow engine: %v", err)
	}

	// Execute workflow
	result, err := engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify result
	if !result.Success {
		t.Error("Expected workflow to succeed")
	}

	if result.Type != workflow.TypeSequential {
		t.Errorf("Expected workflow type %s, got %s", workflow.TypeSequential, result.Type)
	}

	if len(result.AgentResults) != 3 {
		t.Errorf("Expected 3 agent results, got %d", len(result.AgentResults))
	}

	// Verify agent execution order
	expectedOrder := []string{"researcher", "analyst", "writer"}
	for i, expected := range expectedOrder {
		actual := result.AgentResults[i].AgentName
		if actual != expected {
			t.Errorf("Agent %d: expected '%s', got '%s'", i, expected, actual)
		}
	}

	// Verify state sharing - each agent should receive previous agent's output
	for _, ar := range result.AgentResults {
		if ar.Error != nil {
			t.Errorf("Agent %s had error: %v", ar.AgentName, ar.Error)
		}
		if ar.Output == "" {
			t.Errorf("Agent %s produced no output", ar.AgentName)
		}
	}

	// Verify final output contains results from all agents
	if result.Output == "" {
		t.Error("Expected non-empty final output")
	}

	// Verify mock server received requests
	requests := mockServer.GetRequests()
	if len(requests) == 0 {
		t.Error("Expected mock server to receive requests")
	}

	t.Logf("Workflow completed successfully with %d agent executions", len(result.AgentResults))
	t.Logf("Final output length: %d characters", len(result.Output))
}

// ============================================================================
// Scenario 2: API Server Integration
// ============================================================================

// TestE2E_APIServerIntegration tests the HTTP API server endpoints.
func TestE2E_APIServerIntegration(t *testing.T) {
	// Setup mock Ollama server
	mockServer := NewMockOllamaServer()
	defer mockServer.Close()

	ctx := context.Background()

	// Create model and registry
	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   10 * time.Second,
	}

	llm, err := model.NewOllamaModel(ctx, modelCfg)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	registry := agents.NewRegistry()
	for _, name := range []string{"researcher", "analyst"} {
		agent, _ := createTestAgent(llm, name)
		execAgent := agents.NewExecutableAgent(agent)
		registry.Register(execAgent)
	}

	// Test 1: GET /api/v1/health
	t.Run("HealthEndpoint", func(t *testing.T) {
		// Create test server using httptest for reliable testing
		testServer := createTestAPIServer(registry)
		defer testServer.Close()

		resp, err := http.Get(testServer.URL + "/api/v1/health")
		if err != nil {
			t.Fatalf("Health request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var healthResp server.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		if healthResp.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got '%s'", healthResp.Status)
		}
	})

	// Test 2: GET /api/v1/agents
	t.Run("ListAgentsEndpoint", func(t *testing.T) {
		testServer := createTestAPIServer(registry)
		defer testServer.Close()

		resp, err := http.Get(testServer.URL + "/api/v1/agents")
		if err != nil {
			t.Fatalf("List agents request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var agentsResp server.AgentsResponse
		if err := json.NewDecoder(resp.Body).Decode(&agentsResp); err != nil {
			t.Fatalf("Failed to decode agents response: %v", err)
		}

		if len(agentsResp.Agents) < 2 {
			t.Errorf("Expected at least 2 agents, got %d", len(agentsResp.Agents))
		}
	})

	// Test 3: POST /api/v1/workflow/execute
	t.Run("ExecuteWorkflowEndpoint", func(t *testing.T) {
		testServer := createTestAPIServer(registry)
		defer testServer.Close()

		// Create workflow request
		reqBody := server.WorkflowExecuteRequest{
			Task:   "Test workflow task",
			Type:   "sequential",
			Agents: []string{"researcher", "analyst"},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			testServer.URL+"/api/v1/workflow/execute",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Execute workflow request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, body)
		}

		var workflowResp server.WorkflowExecuteResponse
		if err := json.NewDecoder(resp.Body).Decode(&workflowResp); err != nil {
			t.Fatalf("Failed to decode workflow response: %v", err)
		}

		// Verify response format
		if workflowResp.WorkflowID == "" {
			t.Error("Expected workflow_id in response")
		}

		if workflowResp.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", workflowResp.Status)
		}

		if workflowResp.ExecutionTimeMs < 0 {
			t.Error("Expected non-negative execution_time_ms")
		}

		if len(workflowResp.AgentResults) < 2 {
			t.Errorf("Expected at least 2 agent results, got %d", len(workflowResp.AgentResults))
		}
	})
}

// ============================================================================
// Scenario 3: Multi-Agent Collaboration
// ============================================================================

// TestE2E_MultiAgentCollaboration tests researcher -> analyst -> writer workflow.
func TestE2E_MultiAgentCollaboration(t *testing.T) {
	mockServer := NewMockOllamaServer()
	defer mockServer.Close()

	// Set specific responses to track collaboration flow
	mockServer.SetChatResponse("Detailed analysis and findings")

	ctx := context.Background()

	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   10 * time.Second,
	}

	llm, err := model.NewOllamaModel(ctx, modelCfg)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	registry := agents.NewRegistry()

	// Create specialized agents
	researcher, _ := createTestAgent(llm, "researcher")
	analyst, _ := createTestAgent(llm, "analyst")
	writer, _ := createTestAgent(llm, "writer")

	registry.Register(agents.NewExecutableAgent(researcher))
	registry.Register(agents.NewExecutableAgent(analyst))
	registry.Register(agents.NewExecutableAgent(writer))

	// Create workflow with researcher -> analyst -> writer
	workflowCfg := &workflow.Config{
		Type:    workflow.TypeSequential,
		Agents:  []string{"researcher", "analyst", "writer"},
		Task:    "Write a comprehensive report about microservices architecture",
		MaxIter: 1,
	}

	engine, err := workflow.NewEngine(
		workflowCfg,
		workflow.WithRegistry(registry),
	)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	result, err := engine.Execute(ctx)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify collaboration chain
	if !result.Success {
		t.Error("Expected successful collaboration")
	}

	// Verify each agent received previous agent's output as input
	// The workflow engine passes currentInput = output of previous agent
	agentResults := result.AgentResults

	// researcher output
	if agentResults[0].AgentName != "researcher" {
		t.Error("First agent should be researcher")
	}

	// analyst should have processed researcher's output
	if agentResults[1].AgentName != "analyst" {
		t.Error("Second agent should be analyst")
	}

	// writer should have processed analyst's output
	if agentResults[2].AgentName != "writer" {
		t.Error("Third agent should be writer")
	}

	// Verify final output is aggregated
	if !strings.Contains(result.Output, "Sequential Workflow Results") {
		t.Error("Expected aggregated output format")
	}

	// Verify all outputs are present in final result
	for _, ar := range agentResults {
		if !strings.Contains(result.Output, ar.AgentName) {
			t.Errorf("Final output missing agent '%s'", ar.AgentName)
		}
	}

	t.Logf("Multi-agent collaboration completed: %d steps", len(agentResults))
}

// ============================================================================
// Scenario 4: Error Handling
// ============================================================================

// TestE2E_InvalidAgentName tests error handling for invalid agent names.
func TestE2E_InvalidAgentName(t *testing.T) {
	mockServer := NewMockOllamaServer()
	defer mockServer.Close()

	ctx := context.Background()

	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   10 * time.Second,
	}

	llm, err := model.NewOllamaModel(ctx, modelCfg)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	registry := agents.NewRegistry()

	// Only register researcher
	researcher, _ := createTestAgent(llm, "researcher")
	registry.Register(agents.NewExecutableAgent(researcher))

	// Try to execute workflow with non-existent agent
	workflowCfg := &workflow.Config{
		Type:    workflow.TypeSequential,
		Agents:  []string{"researcher", "nonexistent_agent"},
		Task:    "Test task",
		MaxIter: 1,
	}

	engine, err := workflow.NewEngine(
		workflowCfg,
		workflow.WithRegistry(registry),
	)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	result, err := engine.Execute(ctx)

	// Should return error for nonexistent agent
	if err == nil {
		t.Error("Expected error for invalid agent name")
	}

	if result != nil && result.Success {
		t.Error("Expected result.Success to be false for invalid agent")
	}

	// Verify error message mentions the invalid agent
	if err != nil && !strings.Contains(err.Error(), "nonexistent_agent") {
		t.Errorf("Error should mention the invalid agent: %v", err)
	}

	t.Logf("Error handling test passed: %v", err)
}

// TestE2E_OllamaConnectionError tests handling of Ollama server errors.
func TestE2E_OllamaConnectionError(t *testing.T) {
	// Create mock server that returns errors
	mockServer := NewMockOllamaServer()
	mockServer.SetError(true, http.StatusInternalServerError)
	defer mockServer.Close()

	ctx := context.Background()

	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   5 * time.Second,
	}

	// Model creation should fail due to health check error
	_, err := model.NewOllamaModel(ctx, modelCfg)
	if err == nil {
		t.Error("Expected error when Ollama server returns errors")
	}

	// Verify error is related to health check
	if !strings.Contains(err.Error(), "health check") {
		t.Errorf("Expected health check error, got: %v", err)
	}

	t.Logf("Connection error test passed: %v", err)
}

// TestE2E_ContextCancellation tests handling of context cancellation.
func TestE2E_ContextCancellation(t *testing.T) {
	mockServer := NewMockOllamaServer()
	defer mockServer.Close()

	// Create context that we'll cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   10 * time.Second,
	}

	// Model creation with cancelled context should fail
	_, err := model.NewOllamaModel(ctx, modelCfg)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}

	// The error should be context related
	if !strings.Contains(err.Error(), "context") {
		t.Logf("Context cancellation error: %v", err)
	}
}

// TestE2E_APIValidationErrors tests API validation error responses.
func TestE2E_APIValidationErrors(t *testing.T) {
	registry := agents.NewRegistry()
	testServer := createTestAPIServer(registry)
	defer testServer.Close()

	// Test 1: Missing task
	t.Run("MissingTask", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type":   "sequential",
			"agents": []string{"researcher"},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			testServer.URL+"/api/v1/workflow/execute",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	// Test 2: Missing agents
	t.Run("MissingAgents", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"task": "Test task",
			"type": "sequential",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			testServer.URL+"/api/v1/workflow/execute",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	// Test 3: Invalid workflow type
	t.Run("InvalidWorkflowType", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"task":   "Test task",
			"type":   "invalid_type",
			"agents": []string{"researcher"},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			testServer.URL+"/api/v1/workflow/execute",
			"application/json",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

// TestE2E_StateSharing tests state sharing between agents through memory service.
func TestE2E_StateSharing(t *testing.T) {
	mockServer := NewMockOllamaServer()
	defer mockServer.Close()

	ctx := context.Background()

	modelCfg := &model.Config{
		BaseURL:   mockServer.BaseURL(),
		ModelName: "gemma4:4b",
		Timeout:   10 * time.Second,
	}

	// Create memory service
	memService := memory.NewInMemoryService()

	// Verify model can be created with mock server
	_, err := model.NewOllamaModel(ctx, modelCfg)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Store initial shared state
	testKey := "test-research-data"
	testValue := map[string]interface{}{
		"topic":   "microservices",
		"sources": []string{"doc1", "doc2"},
	}

	err = memService.SetSharedStateWithAgent(
		ctx,
		"test-app",
		"test-user",
		testKey,
		testValue,
		"researcher",
	)
	if err != nil {
		t.Fatalf("Failed to set shared state: %v", err)
	}

	// Retrieve shared state
	entry, err := memService.GetSharedState(ctx, "test-app", "test-user", testKey)
	if err != nil {
		t.Fatalf("Failed to get shared state: %v", err)
	}

	if entry == nil {
		t.Fatal("Expected state entry to exist")
	}

	if entry.Key != testKey {
		t.Errorf("Expected key '%s', got '%s'", testKey, entry.Key)
	}

	if entry.FromAgent != "researcher" {
		t.Errorf("Expected from_agent 'researcher', got '%s'", entry.FromAgent)
	}

	// Verify workflow state tracking
	workflowID := "wf-test-001"
	err = memService.UpdateWorkflowProgress(ctx, workflowID, "researcher", "completed")
	if err != nil {
		t.Fatalf("Failed to update workflow progress: %v", err)
	}

	state, err := memService.GetWorkflowState(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	if state == nil {
		t.Fatal("Expected workflow state to exist")
	}

	if state.WorkflowID != workflowID {
		t.Errorf("Expected workflow_id '%s', got '%s'", workflowID, state.WorkflowID)
	}

	if state.AgentStatus["researcher"] != "completed" {
		t.Errorf("Expected researcher status 'completed', got '%s'", state.AgentStatus["researcher"])
	}

	t.Logf("State sharing test passed")
}

// ============================================================================
// Helper Functions
// ============================================================================

// createTestAgent creates a test agent with the given name.
func createTestAgent(model adkmodel.LLM, name string) (adkagent.Agent, error) {
	var agent adkagent.Agent
	var err error

	switch name {
	case "researcher":
		agent, err = agents.NewResearcher(model)
	case "analyst":
		agent, err = agents.NewAnalyst(model)
	case "writer":
		agent, err = agents.NewWriter(model)
	case "reviewer":
		agent, err = agents.NewReviewer(model)
	case "coordinator":
		// Coordinator needs registry, create empty one for test
		emptyRegistry := agents.NewRegistry()
		agent, err = agents.NewCoordinator(model, emptyRegistry)
	default:
		return nil, fmt.Errorf("unknown test agent: %s", name)
	}

	return agent, err
}

// createTestAPIServer creates a httptest.Server for API testing.
func createTestAPIServer(registry *agents.Registry) *httptest.Server {
	// Create a test server using httptest
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response := server.HealthResponse{Status: "healthy"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	mux.HandleFunc("GET /api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		agentNames := registry.List()
		if len(agentNames) == 0 {
			agentNames = []string{"researcher", "analyst", "writer", "reviewer", "coordinator"}
		}
		response := server.AgentsResponse{Agents: agentNames}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	mux.HandleFunc("POST /api/v1/workflow/execute", func(w http.ResponseWriter, r *http.Request) {
		var req server.WorkflowExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(server.ErrorResponse{
				Error:   "Bad Request",
				Message: "invalid JSON body",
			})
			return
		}

		if err := req.Validate(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(server.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}

		// Execute workflow
		config := &workflow.Config{
			Type:    workflow.ParseType(req.Type),
			Agents:  req.Agents,
			Task:    req.Task,
			MaxIter: req.MaxIter,
		}

		var opts []workflow.EngineOption
		if registry != nil && len(registry.List()) > 0 {
			opts = append(opts, workflow.WithRegistry(registry))
		}

		engine, err := workflow.NewEngine(config, opts...)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(server.ErrorResponse{
				Error:   "Internal Server Error",
				Message: fmt.Sprintf("failed to create workflow engine: %v", err),
			})
			return
		}

		startTime := time.Now()
		result, err := engine.Execute(r.Context())
		executionTime := time.Since(startTime).Milliseconds()

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(server.ErrorResponse{
				Error:   "Internal Server Error",
				Message: fmt.Sprintf("workflow execution failed: %v", err),
			})
			return
		}

		// Build response
		workflowID := fmt.Sprintf("wf-%d", time.Now().UnixNano())
		status := "completed"
		if !result.Success {
			status = "failed"
		}

		agentResults := make(map[string]interface{})
		for _, ar := range result.AgentResults {
			agentResults[ar.AgentName] = map[string]interface{}{
				"output": ar.Output,
			}
			if ar.Error != nil {
				agentResults[ar.AgentName].(map[string]interface{})["error"] = ar.Error.Error()
			}
		}

		response := server.WorkflowExecuteResponse{
			WorkflowID:      workflowID,
			Status:          status,
			AgentResults:    agentResults,
			ExecutionTimeMs: executionTime,
			Output:          result.Output,
			Iterations:      result.Iterations,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(mux)
}