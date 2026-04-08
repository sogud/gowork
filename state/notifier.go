package state

import (
	"strconv"
	"sync"
	"time"
)

// EventType represents the type of state change event.
type EventType int

const (
	// EventStateChanged is a generic event indicating state has changed.
	// This is emitted by UpdateState for general state updates.
	EventStateChanged EventType = iota
	// EventAgentStatusChanged indicates an agent's status changed.
	EventAgentStatusChanged
	// EventWorkflowStarted indicates a workflow has started.
	EventWorkflowStarted
	// EventWorkflowCompleted indicates a workflow has completed.
	EventWorkflowCompleted
	// EventOutputUpdated indicates an agent's output was updated.
	EventOutputUpdated
	// EventToolCalled indicates a tool was called by an agent.
	EventToolCalled
)

// String returns a human-readable representation of the event type.
func (e EventType) String() string {
	switch e {
	case EventStateChanged:
		return "StateChanged"
	case EventAgentStatusChanged:
		return "AgentStatusChanged"
	case EventWorkflowStarted:
		return "WorkflowStarted"
	case EventWorkflowCompleted:
		return "WorkflowCompleted"
	case EventOutputUpdated:
		return "OutputUpdated"
	case EventToolCalled:
		return "ToolCalled"
	default:
		return "Unknown"
	}
}

// StateChangeEvent represents a state change notification.
type StateChangeEvent struct {
	// Type is the type of event.
	Type EventType
	// Timestamp is when the event occurred.
	Timestamp time.Time
	// Data contains event-specific data.
	Data interface{}
}

// Subscription represents an active subscription to state change events.
type Subscription struct {
	// ID is the unique identifier for this subscription.
	ID string
}

// StateNotifier defines the interface for state change notifications.
type StateNotifier interface {
	// Subscribe registers a handler to receive state change events.
	// Returns a Subscription that can be used to unsubscribe.
	Subscribe(handler func(StateChangeEvent)) Subscription
	// Unsubscribe removes a subscription.
	Unsubscribe(sub Subscription)
	// Notify sends an event to all subscribers.
	Notify(event StateChangeEvent)
}

// notifier implements the StateNotifier interface.
type notifier struct {
	mu         sync.RWMutex
	handlers   map[string]func(StateChangeEvent)
	nextSubID  int64
}

// NewNotifier creates a new StateNotifier instance.
func NewNotifier() StateNotifier {
	return &notifier{
		handlers: make(map[string]func(StateChangeEvent)),
	}
}

// Subscribe registers a handler to receive state change events.
func (n *notifier) Subscribe(handler func(StateChangeEvent)) Subscription {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.nextSubID++
	id := generateSubscriptionID(n.nextSubID)
	n.handlers[id] = handler

	return Subscription{ID: id}
}

// Unsubscribe removes a subscription.
func (n *notifier) Unsubscribe(sub Subscription) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.handlers, sub.ID)
}

// Notify sends an event to all subscribers.
// Events are delivered synchronously to each subscriber in an undefined order.
// If a handler panics, it will not affect other handlers or the caller.
func (n *notifier) Notify(event StateChangeEvent) {
	n.mu.RLock()
	handlersCopy := make([]func(StateChangeEvent), 0, len(n.handlers))
	for _, handler := range n.handlers {
		handlersCopy = append(handlersCopy, handler)
	}
	n.mu.RUnlock()

	// Deliver events outside the lock to prevent deadlocks
	for _, handler := range handlersCopy {
		// Recover from panics to prevent one bad handler from affecting others
		func(h func(StateChangeEvent)) {
			defer func() {
				if r := recover(); r != nil {
					// Log or handle panic as needed
					// For now, we just recover to prevent crashing
				}
			}()
			h(event)
		}(handler)
	}
}

// generateSubscriptionID creates a unique subscription ID.
func generateSubscriptionID(id int64) string {
	return "sub_" + time.Now().Format("20060102150405") + "_" + strconv.FormatInt(id, 10)
}

// AgentStatusEventData contains data for EventAgentStatusChanged events.
type AgentStatusEventData struct {
	// AgentName is the name of the agent whose status changed.
	AgentName string
	// OldStatus is the previous status.
	OldStatus AgentStatus
	// NewStatus is the new status.
	NewStatus AgentStatus
}

// WorkflowEventData contains data for workflow-related events.
type WorkflowEventData struct {
	// WorkflowID is the unique identifier for the workflow.
	WorkflowID string
	// Task is the task description.
	Task string
	// WorkflowType is the type of workflow.
	WorkflowType WorkflowType
	// Status is the workflow status (for completed events).
	Status WorkflowStatus
	// Error contains any error message.
	Error string
}

// OutputUpdateEventData contains data for EventOutputUpdated events.
type OutputUpdateEventData struct {
	// WorkflowID is the workflow identifier.
	WorkflowID string
	// AgentName is the name of the agent.
	AgentName string
	// Output is the agent's output.
	Output string
}

// ToolCallEventData contains data for EventToolCalled events.
type ToolCallEventData struct {
	// WorkflowID is the workflow identifier.
	WorkflowID string
	// AgentName is the name of the agent that called the tool.
	AgentName string
	// ToolCall contains the tool call information.
	ToolCall ToolCallInfo
}