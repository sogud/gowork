package memory

import (
	"context"
	"time"

	"google.golang.org/adk/session"
)

// Service extends adk-go session service with state sharing capabilities.
// It enables agents to share state during workflow execution.
type Service interface {
	session.Service // Embed adk-go's session service interface

	// State sharing between agents
	GetSharedState(ctx context.Context, appName, userID, key string) (*StateEntry, error)
	SetSharedState(ctx context.Context, appName, userID, key string, value map[string]interface{}) error
	SetSharedStateWithAgent(ctx context.Context, appName, userID, key string, value map[string]interface{}, fromAgent string) error

	// Workflow state tracking
	GetWorkflowState(ctx context.Context, workflowID string) (*WorkflowState, error)
	UpdateWorkflowProgress(ctx context.Context, workflowID, agentName, status string) error
}

// StateEntry represents a shared state entry stored in the memory service.
type StateEntry struct {
	Key       string                 // The key identifier for this state
	Value     map[string]interface{} // The state value
	Timestamp time.Time              // When this state was created/updated
	FromAgent string                 // Which agent created this state
}

// WorkflowState tracks the execution progress of a workflow across agents.
type WorkflowState struct {
	WorkflowID   string            // Unique workflow identifier
	AgentStatus  map[string]string // Agent name -> status (pending, running, completed, failed)
	StartTime    time.Time         // When the workflow started
	CurrentAgent string            // The currently executing agent
}