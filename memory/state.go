package memory

import (
	"context"
	"fmt"
)

// StateManager provides convenient operations for managing shared state.
// It wraps the Service interface with helper methods for common operations.
type StateManager struct {
	service Service
}

// NewStateManager creates a new StateManager with the given service.
func NewStateManager(service Service) *StateManager {
	return &StateManager{
		service: service,
	}
}

// SetSharedState stores a shared state entry in the memory service.
func (s *StateManager) SetSharedState(ctx context.Context, appName, userID, key string, value map[string]interface{}) error {
	return s.service.SetSharedState(ctx, appName, userID, key, value)
}

// GetSharedState retrieves a shared state entry from the memory service.
func (s *StateManager) GetSharedState(ctx context.Context, appName, userID, key string) (*StateEntry, error) {
	return s.service.GetSharedState(ctx, appName, userID, key)
}

// ShareBetweenAgents shares state from one agent to another within a workflow.
// The state is stored with a key prefixed by the workflowID for easy retrieval.
// The FromAgent field is set in the returned StateEntry when retrieved.
func (s *StateManager) ShareBetweenAgents(ctx context.Context, workflowID, fromAgent, toAgent, key string, value map[string]interface{}) error {
	// Create a state entry with metadata
	entryKey := fmt.Sprintf("%s/%s", workflowID, key)

	// Store in the service with the originating agent
	return s.service.SetSharedStateWithAgent(ctx, "gowork", workflowID, entryKey, value, fromAgent)
}

// TrackWorkflowProgress initializes workflow state tracking for the given agents.
// All agents are initialized with "pending" status.
func (s *StateManager) TrackWorkflowProgress(ctx context.Context, workflowID string, agents []string) error {
	// Initialize workflow state by updating progress
	// The UpdateWorkflowProgress method will create the workflow state if it doesn't exist
	for _, agent := range agents {
		err := s.service.UpdateWorkflowProgress(ctx, workflowID, agent, "pending")
		if err != nil {
			return fmt.Errorf("failed to initialize agent '%s' status: %w", agent, err)
		}
	}
	return nil
}

// UpdateAgentStatus updates the status of a specific agent in a workflow.
func (s *StateManager) UpdateAgentStatus(ctx context.Context, workflowID, agentName, status string) error {
	return s.service.UpdateWorkflowProgress(ctx, workflowID, agentName, status)
}

// GetWorkflowState retrieves the current workflow execution state.
func (s *StateManager) GetWorkflowState(ctx context.Context, workflowID string) (*WorkflowState, error) {
	return s.service.GetWorkflowState(ctx, workflowID)
}