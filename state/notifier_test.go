package state

import (
	"sync"
	"testing"
	"time"
)

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventStateChanged, "StateChanged"},
		{EventAgentStatusChanged, "AgentStatusChanged"},
		{EventWorkflowStarted, "WorkflowStarted"},
		{EventWorkflowCompleted, "WorkflowCompleted"},
		{EventOutputUpdated, "OutputUpdated"},
		{EventToolCalled, "ToolCalled"},
		{EventType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.expected {
				t.Errorf("EventType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNotifierSubscribe(t *testing.T) {
	n := NewNotifier()

	var receivedEvent *StateChangeEvent
	var wg sync.WaitGroup
	wg.Add(1)

	sub := n.Subscribe(func(event StateChangeEvent) {
		receivedEvent = &event
		wg.Done()
	})

	if sub.ID == "" {
		t.Errorf("Expected subscription ID to be non-empty")
	}

	// Send an event
	testEvent := StateChangeEvent{
		Type:      EventWorkflowStarted,
		Timestamp: time.Now(),
		Data:      nil,
	}
	n.Notify(testEvent)

	// Wait for the event to be received
	wg.Wait()

	if receivedEvent == nil {
		t.Errorf("Expected to receive event")
	}
	if receivedEvent.Type != EventWorkflowStarted {
		t.Errorf("Expected event type %v, got %v", EventWorkflowStarted, receivedEvent.Type)
	}
}

func TestNotifierUnsubscribe(t *testing.T) {
	n := NewNotifier()

	receivedCount := 0
	handler := func(event StateChangeEvent) {
		receivedCount++
	}

	sub := n.Subscribe(handler)

	// Send first event
	n.Notify(StateChangeEvent{Type: EventAgentStatusChanged, Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond) // Give time for event to be processed

	if receivedCount != 1 {
		t.Errorf("Expected to receive 1 event, got %d", receivedCount)
	}

	// Unsubscribe
	n.Unsubscribe(sub)

	// Send second event
	n.Notify(StateChangeEvent{Type: EventAgentStatusChanged, Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)

	if receivedCount != 1 {
		t.Errorf("Expected to still have 1 event after unsubscribe, got %d", receivedCount)
	}
}

func TestNotifierMultipleSubscribers(t *testing.T) {
	n := NewNotifier()

	var wg sync.WaitGroup
	received1 := false
	received2 := false
	received3 := false

	wg.Add(3)

	n.Subscribe(func(event StateChangeEvent) {
		received1 = true
		wg.Done()
	})

	n.Subscribe(func(event StateChangeEvent) {
		received2 = true
		wg.Done()
	})

	n.Subscribe(func(event StateChangeEvent) {
		received3 = true
		wg.Done()
	})

	// Send event
	n.Notify(StateChangeEvent{Type: EventAgentStatusChanged, Timestamp: time.Now()})

	// Wait for all handlers
	wg.Wait()

	if !received1 || !received2 || !received3 {
		t.Errorf("Expected all subscribers to receive event, got: %v, %v, %v", received1, received2, received3)
	}
}

func TestNotifierPanicRecovery(t *testing.T) {
	n := NewNotifier()

	var wg sync.WaitGroup
	received := false

	wg.Add(2)

	// First handler panics
	n.Subscribe(func(event StateChangeEvent) {
		wg.Done()
		panic("test panic")
	})

	// Second handler should still be called
	n.Subscribe(func(event StateChangeEvent) {
		received = true
		wg.Done()
	})

	// Send event - should not panic
	n.Notify(StateChangeEvent{Type: EventAgentStatusChanged, Timestamp: time.Now()})

	// Wait for handlers
	wg.Wait()

	if !received {
		t.Errorf("Expected second handler to receive event despite first handler panic")
	}
}

func TestAgentStatusEventData(t *testing.T) {
	data := AgentStatusEventData{
		AgentName: "test_agent",
		OldStatus: AgentWaiting,
		NewStatus: AgentRunning,
	}

	if data.AgentName != "test_agent" {
		t.Errorf("AgentName not set correctly")
	}
	if data.OldStatus != AgentWaiting {
		t.Errorf("OldStatus not set correctly")
	}
	if data.NewStatus != AgentRunning {
		t.Errorf("NewStatus not set correctly")
	}
}

func TestWorkflowEventData(t *testing.T) {
	data := WorkflowEventData{
		WorkflowID:   "workflow-123",
		Task:         "test task",
		WorkflowType: WorkflowSequential,
		Status:       WorkflowCompleted,
		Error:        "",
	}

	if data.WorkflowID != "workflow-123" {
		t.Errorf("WorkflowID not set correctly")
	}
	if data.Task != "test task" {
		t.Errorf("Task not set correctly")
	}
}

func TestOutputUpdateEventData(t *testing.T) {
	data := OutputUpdateEventData{
		WorkflowID: "workflow-123",
		AgentName: "agent1",
		Output:    "test output",
	}

	if data.WorkflowID != "workflow-123" {
		t.Errorf("WorkflowID not set correctly")
	}
	if data.AgentName != "agent1" {
		t.Errorf("AgentName not set correctly")
	}
	if data.Output != "test output" {
		t.Errorf("Output not set correctly")
	}
}

func TestToolCallEventData(t *testing.T) {
	toolCall := ToolCallInfo{
		ToolName:  "test_tool",
		Input:     "input",
		Output:    "output",
		Timestamp: time.Now(),
	}

	data := ToolCallEventData{
		WorkflowID: "workflow-123",
		AgentName: "agent1",
		ToolCall:  toolCall,
	}

	if data.WorkflowID != "workflow-123" {
		t.Errorf("WorkflowID not set correctly")
	}
	if data.AgentName != "agent1" {
		t.Errorf("AgentName not set correctly")
	}
	if data.ToolCall.ToolName != "test_tool" {
		t.Errorf("ToolCall not set correctly")
	}
}