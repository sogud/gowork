// Package state provides immutable state management for the TUI.
package state

import (
	"time"
)

// Screen represents the current active screen in the TUI.
type Screen int

const (
	// ScreenHome is the main landing screen.
	ScreenHome Screen = iota
	// ScreenTaskInput is where users input their task.
	ScreenTaskInput
	// ScreenMonitor shows real-time workflow execution.
	ScreenMonitor
	// ScreenConfig allows configuration management.
	ScreenConfig
	// ScreenHistory shows past workflow executions.
	ScreenHistory
)

// String returns a human-readable representation of the screen.
func (s Screen) String() string {
	switch s {
	case ScreenHome:
		return "Home"
	case ScreenTaskInput:
		return "Task Input"
	case ScreenMonitor:
		return "Monitor"
	case ScreenConfig:
		return "Config"
	case ScreenHistory:
		return "History"
	default:
		return "Unknown"
	}
}

// WorkflowType represents the type of workflow execution.
type WorkflowType string

const (
	// WorkflowSequential executes agents one after another in order.
	WorkflowSequential WorkflowType = "sequential"
	// WorkflowParallel executes agents concurrently.
	WorkflowParallel WorkflowType = "parallel"
	// WorkflowLoop executes agents iteratively for refinement.
	WorkflowLoop WorkflowType = "loop"
	// WorkflowDynamic lets the coordinator decide execution pattern.
	WorkflowDynamic WorkflowType = "dynamic"
)

// WorkflowStatus represents the current status of a workflow.
type WorkflowStatus int

const (
	// WorkflowPending indicates the workflow is waiting to start.
	WorkflowPending WorkflowStatus = iota
	// WorkflowRunning indicates the workflow is currently executing.
	WorkflowRunning
	// WorkflowCompleted indicates the workflow finished successfully.
	WorkflowCompleted
	// WorkflowFailed indicates the workflow encountered an error.
	WorkflowFailed
)

// String returns a human-readable representation of the workflow status.
func (s WorkflowStatus) String() string {
	switch s {
	case WorkflowPending:
		return "Pending"
	case WorkflowRunning:
		return "Running"
	case WorkflowCompleted:
		return "Completed"
	case WorkflowFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// AgentStatus represents the current status of an agent.
type AgentStatus int

const (
	// AgentWaiting indicates the agent is waiting to be executed.
	AgentWaiting AgentStatus = iota
	// AgentRunning indicates the agent is currently executing.
	AgentRunning
	// AgentCompleted indicates the agent finished successfully.
	AgentCompleted
	// AgentFailed indicates the agent encountered an error.
	AgentFailed
)

// String returns a human-readable representation of the agent status.
func (s AgentStatus) String() string {
	switch s {
	case AgentWaiting:
		return "Waiting"
	case AgentRunning:
		return "Running"
	case AgentCompleted:
		return "Completed"
	case AgentFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// TaskInputState holds the state for the task input screen.
type TaskInputState struct {
	// TaskDescription is the user's task description.
	TaskDescription string
	// WorkflowType is the selected workflow type.
	WorkflowType WorkflowType
	// SelectedAgents is the list of agents selected for the workflow.
	SelectedAgents []string
	// MaxIterations is the maximum number of iterations for loop/dynamic workflows.
	MaxIterations int
	// CursorField indicates which input field is currently focused.
	CursorField int
}

// Copy creates a deep copy of TaskInputState.
func (t TaskInputState) Copy() TaskInputState {
	agents := make([]string, len(t.SelectedAgents))
	copy(agents, t.SelectedAgents)
	return TaskInputState{
		TaskDescription: t.TaskDescription,
		WorkflowType:     t.WorkflowType,
		SelectedAgents:   agents,
		MaxIterations:    t.MaxIterations,
		CursorField:      t.CursorField,
	}
}

// ToolCallInfo represents information about a tool call made by an agent.
type ToolCallInfo struct {
	// ToolName is the name of the tool that was called.
	ToolName string
	// Input is the input provided to the tool.
	Input string
	// Output is the output returned by the tool.
	Output string
	// Timestamp is when the tool call occurred.
	Timestamp time.Time
}

// Copy creates a copy of ToolCallInfo.
func (t ToolCallInfo) Copy() ToolCallInfo {
	return ToolCallInfo{
		ToolName:  t.ToolName,
		Input:     t.Input,
		Output:    t.Output,
		Timestamp: t.Timestamp,
	}
}

// AgentExecution represents the execution state of a single agent.
type AgentExecution struct {
	// Name is the agent's name.
	Name string
	// Status is the current execution status.
	Status AgentStatus
	// StartTime is when the agent started executing.
	StartTime time.Time
	// Output is the agent's output.
	Output string
	// ToolCalls is the list of tool calls made by the agent.
	ToolCalls []ToolCallInfo
	// TokensUsed is the number of tokens consumed by the agent.
	TokensUsed int
	// Error contains any error that occurred during execution.
	Error string
}

// Copy creates a deep copy of AgentExecution.
func (a AgentExecution) Copy() AgentExecution {
	toolCalls := make([]ToolCallInfo, len(a.ToolCalls))
	for i, tc := range a.ToolCalls {
		toolCalls[i] = tc.Copy()
	}
	return AgentExecution{
		Name:       a.Name,
		Status:     a.Status,
		StartTime:  a.StartTime,
		Output:     a.Output,
		ToolCalls:  toolCalls,
		TokensUsed: a.TokensUsed,
		Error:      a.Error,
	}
}

// WorkflowState represents the state of an active workflow.
type WorkflowState struct {
	// ID is the unique identifier for the workflow.
	ID string
	// Task is the task description.
	Task string
	// Type is the workflow type.
	Type WorkflowType
	// Status is the current workflow status.
	Status WorkflowStatus
	// StartTime is when the workflow started.
	StartTime time.Time
	// ElapsedTime is the duration since the workflow started.
	ElapsedTime time.Duration
	// AgentExecutions contains the execution state of each agent.
	AgentExecutions []AgentExecution
	// CurrentAgent is the name of the currently executing agent.
	CurrentAgent string
	// FinalOutput is the final output from the workflow.
	FinalOutput string
}

// Copy creates a deep copy of WorkflowState.
func (w WorkflowState) Copy() WorkflowState {
	executions := make([]AgentExecution, len(w.AgentExecutions))
	for i, e := range w.AgentExecutions {
		executions[i] = e.Copy()
	}
	return WorkflowState{
		ID:              w.ID,
		Task:            w.Task,
		Type:            w.Type,
		Status:          w.Status,
		StartTime:       w.StartTime,
		ElapsedTime:     w.ElapsedTime,
		AgentExecutions: executions,
		CurrentAgent:    w.CurrentAgent,
		FinalOutput:     w.FinalOutput,
	}
}

// ModelProviderConfig holds the configuration for the model provider.
type ModelProviderConfig struct {
	// Type is the provider type (e.g., "ollama", "gemini").
	Type string
	// ModelName is the name of the model to use.
	ModelName string
	// BaseURL is the base URL for the provider API.
	BaseURL string
	// APIKey is the API key for authentication.
	APIKey string
	// Timeout is the request timeout in seconds.
	Timeout int
}

// Copy creates a copy of ModelProviderConfig.
func (m ModelProviderConfig) Copy() ModelProviderConfig {
	return ModelProviderConfig{
		Type:      m.Type,
		ModelName: m.ModelName,
		BaseURL:   m.BaseURL,
		APIKey:    m.APIKey,
		Timeout:   m.Timeout,
	}
}

// AgentConfigEntry represents a single agent configuration entry.
type AgentConfigEntry struct {
	// Name is the agent's name.
	Name string
	// Enabled indicates whether the agent is enabled.
	Enabled bool
	// Description is a human-readable description of the agent.
	Description string
}

// Copy creates a copy of AgentConfigEntry.
func (a AgentConfigEntry) Copy() AgentConfigEntry {
	return AgentConfigEntry{
		Name:        a.Name,
		Enabled:     a.Enabled,
		Description: a.Description,
	}
}

// ToolConfigEntry represents a single tool configuration entry.
type ToolConfigEntry struct {
	// Name is the tool's name.
	Name string
	// Enabled indicates whether the tool is enabled.
	Enabled bool
	// Description is a human-readable description of the tool.
	Description string
}

// Copy creates a copy of ToolConfigEntry.
func (t ToolConfigEntry) Copy() ToolConfigEntry {
	return ToolConfigEntry{
		Name:        t.Name,
		Enabled:     t.Enabled,
		Description: t.Description,
	}
}

// WorkflowDefaultsConfig holds default workflow configuration values.
type WorkflowDefaultsConfig struct {
	// DefaultType is the default workflow type.
	DefaultType string
	// Timeout is the default timeout in seconds.
	Timeout int
	// MaxIter is the default maximum iterations for loop/dynamic workflows.
	MaxIter int
}

// Copy creates a copy of WorkflowDefaultsConfig.
func (w WorkflowDefaultsConfig) Copy() WorkflowDefaultsConfig {
	return WorkflowDefaultsConfig{
		DefaultType: w.DefaultType,
		Timeout:     w.Timeout,
		MaxIter:     w.MaxIter,
	}
}

// ConfigState holds the configuration state for the TUI.
type ConfigState struct {
	// ModelProvider is the model provider configuration.
	ModelProvider ModelProviderConfig
	// Agents is the list of agent configurations.
	Agents []AgentConfigEntry
	// Tools is the list of tool configurations.
	Tools []ToolConfigEntry
	// WorkflowDefaults contains default workflow settings.
	WorkflowDefaults WorkflowDefaultsConfig
	// EditMode indicates whether the config screen is in edit mode.
	EditMode bool
	// EditField indicates which field is being edited.
	EditField int
}

// Copy creates a deep copy of ConfigState.
func (c ConfigState) Copy() ConfigState {
	agents := make([]AgentConfigEntry, len(c.Agents))
	for i, a := range c.Agents {
		agents[i] = a.Copy()
	}
	tools := make([]ToolConfigEntry, len(c.Tools))
	for i, t := range c.Tools {
		tools[i] = t.Copy()
	}
	return ConfigState{
		ModelProvider:    c.ModelProvider.Copy(),
		Agents:           agents,
		Tools:            tools,
		WorkflowDefaults: c.WorkflowDefaults.Copy(),
		EditMode:         c.EditMode,
		EditField:        c.EditField,
	}
}

// AgentRecord represents an agent's execution record in history.
type AgentRecord struct {
	// Name is the agent's name.
	Name string
	// Input is the input provided to the agent.
	Input string
	// Output is the output from the agent.
	Output string
	// Tokens is the number of tokens consumed.
	Tokens int
	// Duration is how long the agent executed.
	Duration time.Duration
	// ToolCalls is the list of tool calls made by the agent.
	ToolCalls []ToolCallInfo
}

// Copy creates a deep copy of AgentRecord.
func (a AgentRecord) Copy() AgentRecord {
	toolCalls := make([]ToolCallInfo, len(a.ToolCalls))
	for i, tc := range a.ToolCalls {
		toolCalls[i] = tc.Copy()
	}
	return AgentRecord{
		Name:      a.Name,
		Input:     a.Input,
		Output:    a.Output,
		Tokens:    a.Tokens,
		Duration:  a.Duration,
		ToolCalls: toolCalls,
	}
}

// WorkflowRecord represents a historical workflow execution record.
type WorkflowRecord struct {
	// ID is the unique identifier for the workflow.
	ID string
	// Task is the task description.
	Task string
	// Type is the workflow type.
	Type WorkflowType
	// Status is the final workflow status.
	Status WorkflowStatus
	// StartTime is when the workflow started.
	StartTime time.Time
	// EndTime is when the workflow completed.
	EndTime time.Time
	// Duration is the total execution duration.
	Duration time.Duration
	// Agents contains the execution records for each agent.
	Agents []AgentRecord
	// FinalOutput is the final output from the workflow.
	FinalOutput string
	// Error contains any error that occurred during execution.
	Error string
}

// Copy creates a deep copy of WorkflowRecord.
func (w WorkflowRecord) Copy() WorkflowRecord {
	agents := make([]AgentRecord, len(w.Agents))
	for i, a := range w.Agents {
		agents[i] = a.Copy()
	}
	return WorkflowRecord{
		ID:          w.ID,
		Task:        w.Task,
		Type:        w.Type,
		Status:      w.Status,
		StartTime:   w.StartTime,
		EndTime:     w.EndTime,
		Duration:    w.Duration,
		Agents:      agents,
		FinalOutput: w.FinalOutput,
		Error:       w.Error,
	}
}

// HistoryState holds the state for the history screen.
type HistoryState struct {
	// Records is the list of historical workflow records.
	Records []WorkflowRecord
	// SearchQuery is the current search filter.
	SearchQuery string
	// SelectedIndex is the currently selected record index.
	SelectedIndex int
	// ViewDetail indicates whether to show record details.
	ViewDetail bool
	// DetailRecord is the record being viewed in detail.
	DetailRecord *WorkflowRecord
}

// Copy creates a deep copy of HistoryState.
func (h HistoryState) Copy() HistoryState {
	records := make([]WorkflowRecord, len(h.Records))
	for i, r := range h.Records {
		records[i] = r.Copy()
	}
	var detailRecord *WorkflowRecord
	if h.DetailRecord != nil {
		copied := h.DetailRecord.Copy()
		detailRecord = &copied
	}
	return HistoryState{
		Records:       records,
		SearchQuery:   h.SearchQuery,
		SelectedIndex: h.SelectedIndex,
		ViewDetail:    h.ViewDetail,
		DetailRecord:  detailRecord,
	}
}

// AgentInfo represents information about an agent for display.
type AgentInfo struct {
	// Name is the agent's name.
	Name string
	// Description is a human-readable description.
	Description string
	// Status is the current agent status.
	Status AgentStatus
}

// Copy creates a copy of AgentInfo.
func (a AgentInfo) Copy() AgentInfo {
	return AgentInfo{
		Name:        a.Name,
		Description: a.Description,
		Status:      a.Status,
	}
}

// AppState is the immutable root state for the TUI.
// All fields should be considered read-only.
// Use StateStore methods to create modified copies.
type AppState struct {
	// CurrentScreen is the currently active screen.
	CurrentScreen Screen
	// FocusIndex is the index of the currently focused element.
	FocusIndex int
	// TaskInput holds the task input screen state.
	TaskInput TaskInputState
	// ActiveWorkflow holds the active workflow state, if any.
	ActiveWorkflow *WorkflowState
	// Config holds the configuration screen state.
	Config ConfigState
	// History holds the history screen state.
	History HistoryState
	// Agents holds information about available agents.
	Agents []AgentInfo
	// StatusMessage is a temporary status message to display.
	StatusMessage string
	// Error is an error message to display, if any.
	Error string
}

// Copy creates a deep copy of AppState.
// This implements the Copy-on-Write pattern for immutability.
func (a AppState) Copy() AppState {
	agents := make([]AgentInfo, len(a.Agents))
	for i, ag := range a.Agents {
		agents[i] = ag.Copy()
	}
	var activeWorkflow *WorkflowState
	if a.ActiveWorkflow != nil {
		copied := a.ActiveWorkflow.Copy()
		activeWorkflow = &copied
	}
	return AppState{
		CurrentScreen: a.CurrentScreen,
		FocusIndex:    a.FocusIndex,
		TaskInput:     a.TaskInput.Copy(),
		ActiveWorkflow: activeWorkflow,
		Config:        a.Config.Copy(),
		History:       a.History.Copy(),
		Agents:        agents,
		StatusMessage: a.StatusMessage,
		Error:         a.Error,
	}
}

// NewAppState creates a new AppState with default values.
func NewAppState() AppState {
	return AppState{
		CurrentScreen: ScreenHome,
		FocusIndex:    0,
		TaskInput: TaskInputState{
			WorkflowType:  WorkflowSequential,
			MaxIterations: 3,
		},
		Agents: []AgentInfo{},
	}
}