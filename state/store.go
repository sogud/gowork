package state

import (
	"sync"
	"time"
)

// StateStore manages the application state with subscriber notifications.
// It implements Copy-on-Write semantics for the state.
type StateStore struct {
	mu       sync.RWMutex
	state    AppState
	notifier StateNotifier
}

// NewStateStore creates a new StateStore with the given initial state.
// If no notifier is provided, a default one is created.
func NewStateStore(initialState AppState, notifier StateNotifier) *StateStore {
	if notifier == nil {
		notifier = NewNotifier()
	}
	return &StateStore{
		state:    initialState,
		notifier: notifier,
	}
}

// GetState returns a copy of the current state.
// This is safe to call from multiple goroutines.
func (s *StateStore) GetState() AppState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Copy()
}

// updateState performs an atomic state update without emitting notifications.
// This is used internally by methods that emit specific typed events.
func (s *StateStore) updateState(update func(AppState) AppState) {
	s.mu.Lock()
	newState := update(s.state.Copy())
	s.state = newState
	s.mu.Unlock()
}

// UpdateState atomically updates the state using a copy-on-write pattern.
// The update function receives a copy of the current state and returns the modified copy.
// A generic EventStateChanged notification is emitted after the update.
// This is safe to call from multiple goroutines.
func (s *StateStore) UpdateState(update func(AppState) AppState) {
	s.updateState(update)

	// Notify subscribers with a generic state change event
	s.notifier.Notify(StateChangeEvent{
		Type:      EventStateChanged,
		Timestamp: now(),
		Data:      nil,
	})
}

// Subscribe registers a handler to receive state change notifications.
// Returns a Subscription that can be used to unsubscribe.
func (s *StateStore) Subscribe(handler func(StateChangeEvent)) Subscription {
	return s.notifier.Subscribe(handler)
}

// Unsubscribe removes a subscription.
func (s *StateStore) Unsubscribe(sub Subscription) {
	s.notifier.Unsubscribe(sub)
}

// SetScreen changes the current screen.
func (s *StateStore) SetScreen(screen Screen) {
	s.UpdateState(func(state AppState) AppState {
		state.CurrentScreen = screen
		state.FocusIndex = 0
		return state
	})
}

// SetFocusIndex changes the focused element index.
func (s *StateStore) SetFocusIndex(index int) {
	s.UpdateState(func(state AppState) AppState {
		state.FocusIndex = index
		return state
	})
}

// SetTaskDescription updates the task description in the TaskInputState.
func (s *StateStore) SetTaskDescription(description string) {
	s.UpdateState(func(state AppState) AppState {
		state.TaskInput.TaskDescription = description
		return state
	})
}

// SetWorkflowType updates the workflow type in the TaskInputState.
func (s *StateStore) SetWorkflowType(workflowType WorkflowType) {
	s.UpdateState(func(state AppState) AppState {
		state.TaskInput.WorkflowType = workflowType
		return state
	})
}

// SetSelectedAgents updates the selected agents in the TaskInputState.
func (s *StateStore) SetSelectedAgents(agents []string) {
	s.UpdateState(func(state AppState) AppState {
		agentsCopy := make([]string, len(agents))
		copy(agentsCopy, agents)
		state.TaskInput.SelectedAgents = agentsCopy
		return state
	})
}

// SetMaxIterations updates the max iterations in the TaskInputState.
func (s *StateStore) SetMaxIterations(maxIter int) {
	s.UpdateState(func(state AppState) AppState {
		state.TaskInput.MaxIterations = maxIter
		return state
	})
}

// SetCursorField updates the cursor field in the TaskInputState.
func (s *StateStore) SetCursorField(field int) {
	s.UpdateState(func(state AppState) AppState {
		state.TaskInput.CursorField = field
		return state
	})
}

// StartWorkflow creates and sets a new active workflow.
func (s *StateStore) StartWorkflow(id, task string, workflowType WorkflowType, agentNames []string) {
	s.updateState(func(state AppState) AppState {
		executions := make([]AgentExecution, len(agentNames))
		for i, name := range agentNames {
			executions[i] = AgentExecution{
				Name:   name,
				Status: AgentWaiting,
			}
		}

		state.ActiveWorkflow = &WorkflowState{
			ID:              id,
			Task:            task,
			Type:            workflowType,
			Status:          WorkflowRunning,
			StartTime:       now(),
			AgentExecutions: executions,
		}

		return state
	})

	// Notify about workflow start
	s.notifier.Notify(StateChangeEvent{
		Type: EventWorkflowStarted,
		Timestamp: now(),
		Data: WorkflowEventData{
			WorkflowID:   id,
			Task:         task,
			WorkflowType: workflowType,
		},
	})
}

// UpdateAgentStatus updates the status of an agent in the active workflow.
func (s *StateStore) UpdateAgentStatus(agentName string, status AgentStatus) {
	var oldStatus AgentStatus
	var workflowID string

	s.updateState(func(state AppState) AppState {
		if state.ActiveWorkflow == nil {
			return state
		}

		workflowID = state.ActiveWorkflow.ID
		for i := range state.ActiveWorkflow.AgentExecutions {
			if state.ActiveWorkflow.AgentExecutions[i].Name == agentName {
				oldStatus = state.ActiveWorkflow.AgentExecutions[i].Status
				state.ActiveWorkflow.AgentExecutions[i].Status = status
				if status == AgentRunning {
					state.ActiveWorkflow.AgentExecutions[i].StartTime = now()
					state.ActiveWorkflow.CurrentAgent = agentName
				}
				break
			}
		}

		return state
	})

	// Notify about agent status change
	if workflowID != "" {
		s.notifier.Notify(StateChangeEvent{
			Type: EventAgentStatusChanged,
			Timestamp: now(),
			Data: AgentStatusEventData{
				AgentName: agentName,
				OldStatus: oldStatus,
				NewStatus: status,
			},
		})
	}
}

// UpdateAgentOutput updates the output of an agent in the active workflow.
func (s *StateStore) UpdateAgentOutput(agentName, output string) {
	var workflowID string

	s.updateState(func(state AppState) AppState {
		if state.ActiveWorkflow == nil {
			return state
		}

		workflowID = state.ActiveWorkflow.ID
		for i := range state.ActiveWorkflow.AgentExecutions {
			if state.ActiveWorkflow.AgentExecutions[i].Name == agentName {
				state.ActiveWorkflow.AgentExecutions[i].Output = output
				break
			}
		}

		return state
	})

	// Notify about output update
	if workflowID != "" {
		s.notifier.Notify(StateChangeEvent{
			Type: EventOutputUpdated,
			Timestamp: now(),
			Data: OutputUpdateEventData{
				WorkflowID: workflowID,
				AgentName:  agentName,
				Output:     output,
			},
		})
	}
}

// AddToolCall adds a tool call to an agent in the active workflow.
func (s *StateStore) AddToolCall(agentName string, toolCall ToolCallInfo) {
	var workflowID string

	s.updateState(func(state AppState) AppState {
		if state.ActiveWorkflow == nil {
			return state
		}

		workflowID = state.ActiveWorkflow.ID
		for i := range state.ActiveWorkflow.AgentExecutions {
			if state.ActiveWorkflow.AgentExecutions[i].Name == agentName {
				state.ActiveWorkflow.AgentExecutions[i].ToolCalls = append(
					state.ActiveWorkflow.AgentExecutions[i].ToolCalls,
					toolCall,
				)
				break
			}
		}

		return state
	})

	// Notify about tool call
	if workflowID != "" {
		s.notifier.Notify(StateChangeEvent{
			Type: EventToolCalled,
			Timestamp: now(),
			Data: ToolCallEventData{
				WorkflowID: workflowID,
				AgentName:  agentName,
				ToolCall:   toolCall,
			},
		})
	}
}

// CompleteWorkflow marks the active workflow as completed and clears it.
func (s *StateStore) CompleteWorkflow(finalOutput string) {
	var workflow *WorkflowState

	s.updateState(func(state AppState) AppState {
		if state.ActiveWorkflow == nil {
			return state
		}

		workflow = state.ActiveWorkflow
		state.ActiveWorkflow.Status = WorkflowCompleted
		state.ActiveWorkflow.FinalOutput = finalOutput
		state.ActiveWorkflow.ElapsedTime = now().Sub(state.ActiveWorkflow.StartTime)

		return state
	})

	// Notify about workflow completion
	if workflow != nil {
		s.notifier.Notify(StateChangeEvent{
			Type: EventWorkflowCompleted,
			Timestamp: now(),
			Data: WorkflowEventData{
				WorkflowID:   workflow.ID,
				Task:         workflow.Task,
				WorkflowType: workflow.Type,
				Status:       WorkflowCompleted,
			},
		})
	}
}

// FailWorkflow marks the active workflow as failed.
func (s *StateStore) FailWorkflow(err string) {
	var workflow *WorkflowState

	s.updateState(func(state AppState) AppState {
		if state.ActiveWorkflow == nil {
			return state
		}

		workflow = state.ActiveWorkflow
		state.ActiveWorkflow.Status = WorkflowFailed
		state.ActiveWorkflow.ElapsedTime = now().Sub(state.ActiveWorkflow.StartTime)

		return state
	})

	// Notify about workflow failure
	if workflow != nil {
		s.notifier.Notify(StateChangeEvent{
			Type: EventWorkflowCompleted,
			Timestamp: now(),
			Data: WorkflowEventData{
				WorkflowID:   workflow.ID,
				Task:         workflow.Task,
				WorkflowType: workflow.Type,
				Status:       WorkflowFailed,
				Error:       err,
			},
		})
	}
}

// SetStatusMessage sets a temporary status message.
func (s *StateStore) SetStatusMessage(message string) {
	s.UpdateState(func(state AppState) AppState {
		state.StatusMessage = message
		return state
	})
}

// SetError sets an error message.
func (s *StateStore) SetError(err string) {
	s.UpdateState(func(state AppState) AppState {
		state.Error = err
		return state
	})
}

// ClearError clears the error message.
func (s *StateStore) ClearError() {
	s.UpdateState(func(state AppState) AppState {
		state.Error = ""
		return state
	})
}

// SetAgents sets the list of available agents.
func (s *StateStore) SetAgents(agents []AgentInfo) {
	s.UpdateState(func(state AppState) AppState {
		agentsCopy := make([]AgentInfo, len(agents))
		for i, a := range agents {
			agentsCopy[i] = a.Copy()
		}
		state.Agents = agentsCopy
		return state
	})
}

// now is a variable for testing purposes.
var now = func() time.Time {
	return time.Now()
}