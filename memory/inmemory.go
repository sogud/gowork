package memory

import (
	"context"
	"sync"
	"time"

	"google.golang.org/adk/session"
)

// InMemoryService implements the Service interface with in-memory storage.
// It provides thread-safe operations for shared state and workflow tracking.
type InMemoryService struct {
	mu            sync.RWMutex
	sharedState   map[string]map[string]StateEntry // app/user -> key -> entry
	workflowState map[string]*WorkflowState        // workflowID -> state
	session.Service                                // Embed adk-go's in-memory session service
}

// NewInMemoryService creates a new in-memory memory service.
func NewInMemoryService() Service {
	return &InMemoryService{
		sharedState:   make(map[string]map[string]StateEntry),
		workflowState: make(map[string]*WorkflowState),
		Service:       session.InMemoryService(),
	}
}

// GetSharedState retrieves a shared state entry by key.
func (s *InMemoryService) GetSharedState(ctx context.Context, appName, userID, key string) (*StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	appUserKey := appName + "/" + userID
	if states, ok := s.sharedState[appUserKey]; ok {
		if entry, ok := states[key]; ok {
			return &entry, nil
		}
	}
	return nil, nil
}

// SetSharedState stores a shared state entry.
func (s *InMemoryService) SetSharedState(ctx context.Context, appName, userID, key string, value map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	appUserKey := appName + "/" + userID
	if s.sharedState[appUserKey] == nil {
		s.sharedState[appUserKey] = make(map[string]StateEntry)
	}

	s.sharedState[appUserKey][key] = StateEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}
	return nil
}

// SetSharedStateWithAgent stores a shared state entry with the originating agent name.
func (s *InMemoryService) SetSharedStateWithAgent(ctx context.Context, appName, userID, key string, value map[string]interface{}, fromAgent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	appUserKey := appName + "/" + userID
	if s.sharedState[appUserKey] == nil {
		s.sharedState[appUserKey] = make(map[string]StateEntry)
	}

	s.sharedState[appUserKey][key] = StateEntry{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		FromAgent: fromAgent,
	}
	return nil
}

// GetWorkflowState retrieves the workflow execution state.
func (s *InMemoryService) GetWorkflowState(ctx context.Context, workflowID string) (*WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if state, ok := s.workflowState[workflowID]; ok {
		return state, nil
	}
	return nil, nil
}

// UpdateWorkflowProgress updates the workflow execution progress.
// If the workflow doesn't exist, it creates a new workflow state.
func (s *InMemoryService) UpdateWorkflowProgress(ctx context.Context, workflowID, agentName, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create workflow state if it doesn't exist
	if s.workflowState[workflowID] == nil {
		s.workflowState[workflowID] = &WorkflowState{
			WorkflowID:  workflowID,
			AgentStatus: make(map[string]string),
			StartTime:   time.Now(),
		}
	}

	// Update agent status if agent name is provided
	if agentName != "" {
		s.workflowState[workflowID].AgentStatus[agentName] = status
		s.workflowState[workflowID].CurrentAgent = agentName
	}

	return nil
}