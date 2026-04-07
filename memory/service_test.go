package memory

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryService_SetAndGetSharedState(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	appName := "gowork"
	userID := "workflow-123"
	key := "research-findings"
	value := map[string]interface{}{
		"topic":   "AI agents",
		"summary": "Key findings about multi-agent systems",
		"sources": []string{"source1", "source2"},
	}

	// Set state
	err := service.SetSharedState(ctx, appName, userID, key, value)
	if err != nil {
		t.Fatalf("Failed to set state: %v", err)
	}

	// Get state
	retrieved, err := service.GetSharedState(ctx, appName, userID, key)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected non-nil state entry")
	}

	if retrieved.Key != key {
		t.Errorf("Expected key '%s', got '%s'", key, retrieved.Key)
	}

	if retrieved.Value["topic"] != "AI agents" {
		t.Errorf("Expected topic 'AI agents', got '%v'", retrieved.Value["topic"])
	}

	if retrieved.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestInMemoryService_GetSharedState_NotFound(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	retrieved, err := service.GetSharedState(ctx, "app", "user", "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil for nonexistent key, got %v", retrieved)
	}
}

func TestInMemoryService_WorkflowState(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	workflowID := "workflow-123"

	// Update workflow progress
	err := service.UpdateWorkflowProgress(ctx, workflowID, "researcher", "completed")
	if err != nil {
		t.Fatalf("Failed to update workflow progress: %v", err)
	}

	// Get workflow state
	state, err := service.GetWorkflowState(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	if state == nil {
		t.Fatal("Expected non-nil workflow state")
	}

	if state.WorkflowID != workflowID {
		t.Errorf("Expected workflowID '%s', got '%s'", workflowID, state.WorkflowID)
	}

	if state.AgentStatus["researcher"] != "completed" {
		t.Errorf("Expected researcher status 'completed', got '%s'", state.AgentStatus["researcher"])
	}

	if state.CurrentAgent != "researcher" {
		t.Errorf("Expected current agent 'researcher', got '%s'", state.CurrentAgent)
	}

	if state.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
}

func TestInMemoryService_WorkflowProgress_MultipleAgents(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	workflowID := "workflow-456"

	// Update multiple agents
	err := service.UpdateWorkflowProgress(ctx, workflowID, "researcher", "completed")
	if err != nil {
		t.Fatalf("Failed to update researcher: %v", err)
	}

	err = service.UpdateWorkflowProgress(ctx, workflowID, "analyst", "running")
	if err != nil {
		t.Fatalf("Failed to update analyst: %v", err)
	}

	err = service.UpdateWorkflowProgress(ctx, workflowID, "writer", "pending")
	if err != nil {
		t.Fatalf("Failed to update writer: %v", err)
	}

	// Get workflow state
	state, err := service.GetWorkflowState(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	// Verify all agents
	if state.AgentStatus["researcher"] != "completed" {
		t.Errorf("Expected researcher status 'completed', got '%s'", state.AgentStatus["researcher"])
	}

	if state.AgentStatus["analyst"] != "running" {
		t.Errorf("Expected analyst status 'running', got '%s'", state.AgentStatus["analyst"])
	}

	if state.AgentStatus["writer"] != "pending" {
		t.Errorf("Expected writer status 'pending', got '%s'", state.AgentStatus["writer"])
	}

	if state.CurrentAgent != "writer" {
		t.Errorf("Expected current agent 'writer', got '%s'", state.CurrentAgent)
	}
}

func TestInMemoryService_WorkflowState_NotFound(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	state, err := service.GetWorkflowState(ctx, "nonexistent-workflow")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if state != nil {
		t.Errorf("Expected nil for nonexistent workflow, got %v", state)
	}
}

func TestStateManager_SetAndGetSharedState(t *testing.T) {
	service := NewInMemoryService()
	manager := NewStateManager(service)

	ctx := context.Background()
	appName := "gowork"
	userID := "workflow-789"
	key := "analysis-results"
	value := map[string]interface{}{
		"data": "processed data",
		"count": 42,
	}

	err := manager.SetSharedState(ctx, appName, userID, key, value)
	if err != nil {
		t.Fatalf("Failed to set state: %v", err)
	}

	retrieved, err := manager.GetSharedState(ctx, appName, userID, key)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if retrieved.Value["data"] != "processed data" {
		t.Errorf("Expected 'processed data', got '%v'", retrieved.Value["data"])
	}
}

func TestStateManager_ShareBetweenAgents(t *testing.T) {
	service := NewInMemoryService()
	manager := NewStateManager(service)

	ctx := context.Background()
	workflowID := "workflow-sharing"
	fromAgent := "researcher"
	toAgent := "analyst"
	key := "findings"
	value := map[string]interface{}{
		"insights": []string{"insight1", "insight2"},
	}

	err := manager.ShareBetweenAgents(ctx, workflowID, fromAgent, toAgent, key, value)
	if err != nil {
		t.Fatalf("Failed to share between agents: %v", err)
	}

	// Verify the state was stored with the correct key
	expectedKey := workflowID + "/" + key
	retrieved, err := service.GetSharedState(ctx, "gowork", workflowID, expectedKey)
	if err != nil {
		t.Fatalf("Failed to get shared state: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected non-nil state entry")
	}

	// Check that FromAgent is set
	if retrieved.FromAgent != fromAgent {
		t.Errorf("Expected FromAgent '%s', got '%s'", fromAgent, retrieved.FromAgent)
	}
}

func TestStateManager_TrackWorkflowProgress(t *testing.T) {
	service := NewInMemoryService()
	manager := NewStateManager(service)

	ctx := context.Background()
	workflowID := "workflow-tracking"
	agents := []string{"researcher", "analyst", "writer"}

	err := manager.TrackWorkflowProgress(ctx, workflowID, agents)
	if err != nil {
		t.Fatalf("Failed to track workflow progress: %v", err)
	}

	// Verify workflow state was created
	state, err := service.GetWorkflowState(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	if state == nil {
		t.Fatal("Expected non-nil workflow state")
	}

	// All agents should be initialized as pending
	for _, agent := range agents {
		if state.AgentStatus[agent] != "pending" {
			t.Errorf("Expected agent '%s' status 'pending', got '%s'", agent, state.AgentStatus[agent])
		}
	}
}

func TestInMemoryService_ThreadSafety(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	appName := "test-app"
	userID := "test-user"

	// Run concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "concurrent-key"
			value := map[string]interface{}{"id": id}
			_ = service.SetSharedState(ctx, appName, userID, key, value)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no race conditions (just check that we can retrieve)
	_, err := service.GetSharedState(ctx, appName, userID, "concurrent-key")
	if err != nil {
		t.Errorf("Unexpected error after concurrent writes: %v", err)
	}
}

func TestStateEntry_Timestamp(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	before := time.Now()

	err := service.SetSharedState(ctx, "app", "user", "key", map[string]interface{}{"data": "test"})
	if err != nil {
		t.Fatalf("Failed to set state: %v", err)
	}

	after := time.Now()

	retrieved, err := service.GetSharedState(ctx, "app", "user", "key")
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if retrieved.Timestamp.Before(before) || retrieved.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in expected range [%v, %v]", retrieved.Timestamp, before, after)
	}
}

func TestInMemoryService_EmptyAgentName(t *testing.T) {
	service := NewInMemoryService()

	ctx := context.Background()
	workflowID := "workflow-empty"

	// Update with empty agent name should still create workflow state
	err := service.UpdateWorkflowProgress(ctx, workflowID, "", "initialized")
	if err != nil {
		t.Fatalf("Failed to update workflow progress: %v", err)
	}

	state, err := service.GetWorkflowState(ctx, workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	if state == nil {
		t.Fatal("Expected non-nil workflow state")
	}

	// AgentStatus should be empty
	if len(state.AgentStatus) != 0 {
		t.Errorf("Expected empty AgentStatus, got %v", state.AgentStatus)
	}
}