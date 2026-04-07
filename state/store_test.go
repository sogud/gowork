package state

import (
	"sync"
	"testing"
	"time"
)

func TestNewStateStore(t *testing.T) {
	initialState := NewAppState()
	store := NewStateStore(initialState, nil)

	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	state := store.GetState()
	if state.CurrentScreen != ScreenHome {
		t.Errorf("Expected initial screen to be ScreenHome, got %v", state.CurrentScreen)
	}
}

func TestStateStoreGetState(t *testing.T) {
	initialState := NewAppState()
	initialState.CurrentScreen = ScreenMonitor
	store := NewStateStore(initialState, nil)

	state := store.GetState()

	if state.CurrentScreen != ScreenMonitor {
		t.Errorf("Expected screen to be Monitor, got %v", state.CurrentScreen)
	}

	// Verify it's a copy
	state.CurrentScreen = ScreenConfig
	state2 := store.GetState()
	if state2.CurrentScreen == ScreenConfig {
		t.Errorf("GetState should return a copy, not the original")
	}
}

func TestStateStoreUpdateState(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	updated := false
	store.UpdateState(func(state AppState) AppState {
		state.CurrentScreen = ScreenConfig
		updated = true
		return state
	})

	if !updated {
		t.Errorf("Update function was not called")
	}

	state := store.GetState()
	if state.CurrentScreen != ScreenConfig {
		t.Errorf("Expected screen to be updated to Config, got %v", state.CurrentScreen)
	}
}

func TestStateStoreSubscribe(t *testing.T) {
	notifier := NewNotifier()
	store := NewStateStore(NewAppState(), notifier)

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		receivedEvent = &event
		wg.Done()
	})

	// Trigger an update
	store.SetScreen(ScreenConfig)

	// Wait for notification
	wg.Wait()

	if receivedEvent == nil {
		t.Errorf("Expected to receive event notification")
	}
}

func TestStateStoreUnsubscribe(t *testing.T) {
	notifier := NewNotifier()
	store := NewStateStore(NewAppState(), notifier)

	receivedCount := 0
	sub := store.Subscribe(func(event StateChangeEvent) {
		receivedCount++
	})

	// First update
	store.SetScreen(ScreenConfig)
	time.Sleep(10 * time.Millisecond)

	if receivedCount != 1 {
		t.Errorf("Expected 1 event, got %d", receivedCount)
	}

	// Unsubscribe
	store.Unsubscribe(sub)

	// Second update
	store.SetScreen(ScreenMonitor)
	time.Sleep(10 * time.Millisecond)

	if receivedCount != 1 {
		t.Errorf("Expected still 1 event after unsubscribe, got %d", receivedCount)
	}
}

func TestStateStoreSetScreen(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetScreen(ScreenMonitor)

	state := store.GetState()
	if state.CurrentScreen != ScreenMonitor {
		t.Errorf("Expected screen to be Monitor, got %v", state.CurrentScreen)
	}
	if state.FocusIndex != 0 {
		t.Errorf("Expected FocusIndex to be reset to 0, got %v", state.FocusIndex)
	}
}

func TestStateStoreSetFocusIndex(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetFocusIndex(5)

	state := store.GetState()
	if state.FocusIndex != 5 {
		t.Errorf("Expected FocusIndex to be 5, got %v", state.FocusIndex)
	}
}

func TestStateStoreSetTaskDescription(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetTaskDescription("test task description")

	state := store.GetState()
	if state.TaskInput.TaskDescription != "test task description" {
		t.Errorf("Expected TaskDescription to be set, got %v", state.TaskInput.TaskDescription)
	}
}

func TestStateStoreSetWorkflowType(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetWorkflowType(WorkflowParallel)

	state := store.GetState()
	if state.TaskInput.WorkflowType != WorkflowParallel {
		t.Errorf("Expected WorkflowType to be Parallel, got %v", state.TaskInput.WorkflowType)
	}
}

func TestStateStoreSetSelectedAgents(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	agents := []string{"agent1", "agent2", "agent3"}
	store.SetSelectedAgents(agents)

	state := store.GetState()
	if len(state.TaskInput.SelectedAgents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(state.TaskInput.SelectedAgents))
	}

	// Verify it's a copy
	agents[0] = "modified"
	if state.TaskInput.SelectedAgents[0] == "modified" {
		t.Errorf("SelectedAgents should be a copy")
	}
}

func TestStateStoreSetMaxIterations(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetMaxIterations(10)

	state := store.GetState()
	if state.TaskInput.MaxIterations != 10 {
		t.Errorf("Expected MaxIterations to be 10, got %v", state.TaskInput.MaxIterations)
	}
}

func TestStateStoreSetCursorField(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetCursorField(3)

	state := store.GetState()
	if state.TaskInput.CursorField != 3 {
		t.Errorf("Expected CursorField to be 3, got %v", state.TaskInput.CursorField)
	}
}

func TestStateStoreStartWorkflow(t *testing.T) {
	notifier := NewNotifier()
	store := NewStateStore(NewAppState(), notifier)

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		if event.Type == EventWorkflowStarted {
			receivedEvent = &event
			wg.Done()
		}
	})

	agents := []string{"agent1", "agent2"}
	store.StartWorkflow("workflow-123", "test task", WorkflowSequential, agents)

	// Wait for event
	wg.Wait()

	state := store.GetState()
	if state.ActiveWorkflow == nil {
		t.Fatal("Expected ActiveWorkflow to be set")
	}
	if state.ActiveWorkflow.ID != "workflow-123" {
		t.Errorf("Expected workflow ID to be workflow-123, got %v", state.ActiveWorkflow.ID)
	}
	if state.ActiveWorkflow.Task != "test task" {
		t.Errorf("Expected task to be 'test task', got %v", state.ActiveWorkflow.Task)
	}
	if state.ActiveWorkflow.Type != WorkflowSequential {
		t.Errorf("Expected type to be Sequential, got %v", state.ActiveWorkflow.Type)
	}
	if state.ActiveWorkflow.Status != WorkflowRunning {
		t.Errorf("Expected status to be Running, got %v", state.ActiveWorkflow.Status)
	}
	if len(state.ActiveWorkflow.AgentExecutions) != 2 {
		t.Errorf("Expected 2 agent executions, got %d", len(state.ActiveWorkflow.AgentExecutions))
	}

	// Verify event was received
	if receivedEvent == nil {
		t.Errorf("Expected to receive WorkflowStarted event")
	}
}

func TestStateStoreUpdateAgentStatus(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	// Start a workflow first
	store.StartWorkflow("workflow-123", "test", WorkflowSequential, []string{"agent1", "agent2"})

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		if event.Type == EventAgentStatusChanged {
			receivedEvent = &event
			wg.Done()
		}
	})

	// Update agent status
	store.UpdateAgentStatus("agent1", AgentRunning)

	// Wait for event
	wg.Wait()

	state := store.GetState()
	found := false
	for _, exec := range state.ActiveWorkflow.AgentExecutions {
		if exec.Name == "agent1" {
			found = true
			if exec.Status != AgentRunning {
				t.Errorf("Expected agent1 status to be Running, got %v", exec.Status)
			}
		}
	}
	if !found {
		t.Errorf("Agent agent1 not found in executions")
	}

	if state.ActiveWorkflow.CurrentAgent != "agent1" {
		t.Errorf("Expected CurrentAgent to be agent1, got %v", state.ActiveWorkflow.CurrentAgent)
	}

	// Verify event
	if receivedEvent == nil {
		t.Errorf("Expected to receive AgentStatusChanged event")
	}
}

func TestStateStoreUpdateAgentOutput(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	// Start a workflow
	store.StartWorkflow("workflow-123", "test", WorkflowSequential, []string{"agent1"})

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		if event.Type == EventOutputUpdated {
			receivedEvent = &event
			wg.Done()
		}
	})

	// Update output
	store.UpdateAgentOutput("agent1", "test output")

	// Wait for event
	wg.Wait()

	state := store.GetState()
	found := false
	for _, exec := range state.ActiveWorkflow.AgentExecutions {
		if exec.Name == "agent1" {
			found = true
			if exec.Output != "test output" {
				t.Errorf("Expected output to be 'test output', got %v", exec.Output)
			}
		}
	}
	if !found {
		t.Errorf("Agent agent1 not found")
	}

	// Verify event
	if receivedEvent == nil {
		t.Errorf("Expected to receive OutputUpdated event")
	}
}

func TestStateStoreAddToolCall(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	// Start a workflow
	store.StartWorkflow("workflow-123", "test", WorkflowSequential, []string{"agent1"})

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		if event.Type == EventToolCalled {
			receivedEvent = &event
			wg.Done()
		}
	})

	// Add tool call
	toolCall := ToolCallInfo{
		ToolName:  "test_tool",
		Input:     "input",
		Output:    "output",
		Timestamp: time.Now(),
	}
	store.AddToolCall("agent1", toolCall)

	// Wait for event
	wg.Wait()

	state := store.GetState()
	found := false
	for _, exec := range state.ActiveWorkflow.AgentExecutions {
		if exec.Name == "agent1" {
			found = true
			if len(exec.ToolCalls) != 1 {
				t.Errorf("Expected 1 tool call, got %d", len(exec.ToolCalls))
			}
			if exec.ToolCalls[0].ToolName != "test_tool" {
				t.Errorf("Expected tool name to be 'test_tool', got %v", exec.ToolCalls[0].ToolName)
			}
		}
	}
	if !found {
		t.Errorf("Agent agent1 not found")
	}

	// Verify event
	if receivedEvent == nil {
		t.Errorf("Expected to receive ToolCalled event")
	}
}

func TestStateStoreCompleteWorkflow(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	// Start a workflow
	store.StartWorkflow("workflow-123", "test", WorkflowSequential, []string{"agent1"})

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		if event.Type == EventWorkflowCompleted {
			receivedEvent = &event
			wg.Done()
		}
	})

	// Complete workflow
	store.CompleteWorkflow("final output")

	// Wait for event
	wg.Wait()

	state := store.GetState()
	if state.ActiveWorkflow.Status != WorkflowCompleted {
		t.Errorf("Expected status to be Completed, got %v", state.ActiveWorkflow.Status)
	}
	if state.ActiveWorkflow.FinalOutput != "final output" {
		t.Errorf("Expected FinalOutput to be 'final output', got %v", state.ActiveWorkflow.FinalOutput)
	}
	if state.ActiveWorkflow.ElapsedTime == 0 {
		t.Errorf("Expected ElapsedTime to be set")
	}

	// Verify event
	if receivedEvent == nil {
		t.Errorf("Expected to receive WorkflowCompleted event")
	}
}

func TestStateStoreFailWorkflow(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	// Start a workflow
	store.StartWorkflow("workflow-123", "test", WorkflowSequential, []string{"agent1"})

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedEvent *StateChangeEvent
	store.Subscribe(func(event StateChangeEvent) {
		if event.Type == EventWorkflowCompleted {
			receivedEvent = &event
			wg.Done()
		}
	})

	// Fail workflow
	store.FailWorkflow("test error")

	// Wait for event
	wg.Wait()

	state := store.GetState()
	if state.ActiveWorkflow.Status != WorkflowFailed {
		t.Errorf("Expected status to be Failed, got %v", state.ActiveWorkflow.Status)
	}

	// Verify event
	if receivedEvent == nil {
		t.Errorf("Expected to receive WorkflowCompleted event")
	}
	data, ok := receivedEvent.Data.(WorkflowEventData)
	if !ok {
		t.Errorf("Expected WorkflowEventData")
	}
	if data.Error != "test error" {
		t.Errorf("Expected error to be 'test error', got %v", data.Error)
	}
}

func TestStateStoreSetStatusMessage(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetStatusMessage("test message")

	state := store.GetState()
	if state.StatusMessage != "test message" {
		t.Errorf("Expected StatusMessage to be 'test message', got %v", state.StatusMessage)
	}
}

func TestStateStoreSetError(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetError("test error")

	state := store.GetState()
	if state.Error != "test error" {
		t.Errorf("Expected Error to be 'test error', got %v", state.Error)
	}
}

func TestStateStoreClearError(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	store.SetError("test error")
	store.ClearError()

	state := store.GetState()
	if state.Error != "" {
		t.Errorf("Expected Error to be cleared, got %v", state.Error)
	}
}

func TestStateStoreSetAgents(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	agents := []AgentInfo{
		{Name: "agent1", Description: "Agent 1", Status: AgentWaiting},
		{Name: "agent2", Description: "Agent 2", Status: AgentRunning},
	}
	store.SetAgents(agents)

	state := store.GetState()
	if len(state.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(state.Agents))
	}

	// Verify it's a copy
	agents[0].Name = "modified"
	if state.Agents[0].Name == "modified" {
		t.Errorf("Agents should be a copy")
	}
}

func TestStateStoreConcurrency(t *testing.T) {
	store := NewStateStore(NewAppState(), nil)

	var wg sync.WaitGroup
	numOps := 100

	// Concurrent reads
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.GetState()
		}()
	}

	// Concurrent writes
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.SetFocusIndex(i)
		}(i)
	}

	wg.Wait()

	// If we get here without a race condition, the test passes
}