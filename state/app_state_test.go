package state

import (
	"testing"
	"time"
)

func TestScreenString(t *testing.T) {
	tests := []struct {
		screen   Screen
		expected string
	}{
		{ScreenHome, "Home"},
		{ScreenTaskInput, "Task Input"},
		{ScreenMonitor, "Monitor"},
		{ScreenConfig, "Config"},
		{ScreenHistory, "History"},
		{Screen(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.screen.String(); got != tt.expected {
				t.Errorf("Screen.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWorkflowStatusString(t *testing.T) {
	tests := []struct {
		status   WorkflowStatus
		expected string
	}{
		{WorkflowPending, "Pending"},
		{WorkflowRunning, "Running"},
		{WorkflowCompleted, "Completed"},
		{WorkflowFailed, "Failed"},
		{WorkflowStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("WorkflowStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAgentStatusString(t *testing.T) {
	tests := []struct {
		status   AgentStatus
		expected string
	}{
		{AgentWaiting, "Waiting"},
		{AgentRunning, "Running"},
		{AgentCompleted, "Completed"},
		{AgentFailed, "Failed"},
		{AgentStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("AgentStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTaskInputStateCopy(t *testing.T) {
	original := TaskInputState{
		TaskDescription: "Test task",
		WorkflowType:    WorkflowSequential,
		SelectedAgents:  []string{"agent1", "agent2"},
		MaxIterations:   5,
		CursorField:     1,
	}

	copy := original.Copy()

	// Verify copy is equal
	if copy.TaskDescription != original.TaskDescription {
		t.Errorf("TaskDescription not copied correctly")
	}
	if copy.WorkflowType != original.WorkflowType {
		t.Errorf("WorkflowType not copied correctly")
	}
	if len(copy.SelectedAgents) != len(original.SelectedAgents) {
		t.Errorf("SelectedAgents length mismatch")
	}
	if copy.MaxIterations != original.MaxIterations {
		t.Errorf("MaxIterations not copied correctly")
	}
	if copy.CursorField != original.CursorField {
		t.Errorf("CursorField not copied correctly")
	}

	// Verify it's a deep copy
	copy.SelectedAgents[0] = "modified"
	if original.SelectedAgents[0] == "modified" {
		t.Errorf("SelectedAgents is not a deep copy")
	}
}

func TestToolCallInfoCopy(t *testing.T) {
	ts := time.Now()
	original := ToolCallInfo{
		ToolName:  "test_tool",
		Input:     "test input",
		Output:    "test output",
		Timestamp: ts,
	}

	copy := original.Copy()

	if copy.ToolName != original.ToolName {
		t.Errorf("ToolName not copied correctly")
	}
	if copy.Input != original.Input {
		t.Errorf("Input not copied correctly")
	}
	if copy.Output != original.Output {
		t.Errorf("Output not copied correctly")
	}
	if !copy.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp not copied correctly")
	}
}

func TestAgentExecutionCopy(t *testing.T) {
	ts := time.Now()
	original := AgentExecution{
		Name:       "test_agent",
		Status:     AgentRunning,
		StartTime:  ts,
		Output:     "test output",
		ToolCalls:  []ToolCallInfo{{ToolName: "tool1"}},
		TokensUsed: 100,
		Error:      "test error",
	}

	copy := original.Copy()

	if copy.Name != original.Name {
		t.Errorf("Name not copied correctly")
	}
	if copy.Status != original.Status {
		t.Errorf("Status not copied correctly")
	}
	if len(copy.ToolCalls) != len(original.ToolCalls) {
		t.Errorf("ToolCalls length mismatch")
	}

	// Verify deep copy
	copy.ToolCalls[0].ToolName = "modified"
	if original.ToolCalls[0].ToolName == "modified" {
		t.Errorf("ToolCalls is not a deep copy")
	}
}

func TestWorkflowStateCopy(t *testing.T) {
	ts := time.Now()
	original := WorkflowState{
		ID:              "workflow-123",
		Task:            "test task",
		Type:            WorkflowSequential,
		Status:          WorkflowRunning,
		StartTime:       ts,
		ElapsedTime:     time.Hour,
		AgentExecutions: []AgentExecution{{Name: "agent1"}},
		CurrentAgent:    "agent1",
		FinalOutput:     "final",
	}

	copy := original.Copy()

	if copy.ID != original.ID {
		t.Errorf("ID not copied correctly")
	}
	if len(copy.AgentExecutions) != len(original.AgentExecutions) {
		t.Errorf("AgentExecutions length mismatch")
	}

	// Verify deep copy
	copy.AgentExecutions[0].Name = "modified"
	if original.AgentExecutions[0].Name == "modified" {
		t.Errorf("AgentExecutions is not a deep copy")
	}
}

func TestModelProviderConfigCopy(t *testing.T) {
	original := ModelProviderConfig{
		Type:      "ollama",
		ModelName: "llama2",
		BaseURL:   "http://localhost:11434",
		APIKey:    "secret",
		Timeout:   30,
	}

	copy := original.Copy()

	if copy.Type != original.Type {
		t.Errorf("Type not copied correctly")
	}
	if copy.ModelName != original.ModelName {
		t.Errorf("ModelName not copied correctly")
	}
	if copy.APIKey != original.APIKey {
		t.Errorf("APIKey not copied correctly")
	}
}

func TestConfigStateCopy(t *testing.T) {
	original := ConfigState{
		ModelProvider: ModelProviderConfig{Type: "ollama"},
		Agents:        []AgentConfigEntry{{Name: "agent1"}},
		Tools:         []ToolConfigEntry{{Name: "tool1"}},
		WorkflowDefaults: WorkflowDefaultsConfig{DefaultType: "sequential"},
		EditMode:      true,
		EditField:     2,
	}

	copy := original.Copy()

	if copy.ModelProvider.Type != original.ModelProvider.Type {
		t.Errorf("ModelProvider not copied correctly")
	}
	if len(copy.Agents) != len(original.Agents) {
		t.Errorf("Agents length mismatch")
	}
	if len(copy.Tools) != len(original.Tools) {
		t.Errorf("Tools length mismatch")
	}
	if copy.EditMode != original.EditMode {
		t.Errorf("EditMode not copied correctly")
	}

	// Verify deep copy
	copy.Agents[0].Name = "modified"
	if original.Agents[0].Name == "modified" {
		t.Errorf("Agents is not a deep copy")
	}
}

func TestWorkflowRecordCopy(t *testing.T) {
	ts := time.Now()
	original := WorkflowRecord{
		ID:          "record-123",
		Task:        "test task",
		Type:        WorkflowSequential,
		Status:      WorkflowCompleted,
		StartTime:   ts,
		EndTime:     ts.Add(time.Hour),
		Duration:    time.Hour,
		Agents:      []AgentRecord{{Name: "agent1"}},
		FinalOutput: "final",
		Error:       "",
	}

	copy := original.Copy()

	if copy.ID != original.ID {
		t.Errorf("ID not copied correctly")
	}
	if len(copy.Agents) != len(original.Agents) {
		t.Errorf("Agents length mismatch")
	}

	// Verify deep copy
	copy.Agents[0].Name = "modified"
	if original.Agents[0].Name == "modified" {
		t.Errorf("Agents is not a deep copy")
	}
}

func TestHistoryStateCopy(t *testing.T) {
	record := WorkflowRecord{ID: "record-123"}
	original := HistoryState{
		Records:       []WorkflowRecord{record},
		SearchQuery:   "test",
		SelectedIndex: 1,
		ViewDetail:    true,
		DetailRecord:  &record,
	}

	copy := original.Copy()

	if len(copy.Records) != len(original.Records) {
		t.Errorf("Records length mismatch")
	}
	if copy.SearchQuery != original.SearchQuery {
		t.Errorf("SearchQuery not copied correctly")
	}
	if copy.DetailRecord == nil {
		t.Errorf("DetailRecord is nil")
	}
	if copy.DetailRecord.ID != original.DetailRecord.ID {
		t.Errorf("DetailRecord not copied correctly")
	}

	// Verify deep copy
	copy.Records[0].ID = "modified"
	if original.Records[0].ID == "modified" {
		t.Errorf("Records is not a deep copy")
	}
}

func TestAgentInfoCopy(t *testing.T) {
	original := AgentInfo{
		Name:        "test_agent",
		Description: "Test agent description",
		Status:      AgentRunning,
	}

	copy := original.Copy()

	if copy.Name != original.Name {
		t.Errorf("Name not copied correctly")
	}
	if copy.Description != original.Description {
		t.Errorf("Description not copied correctly")
	}
	if copy.Status != original.Status {
		t.Errorf("Status not copied correctly")
	}
}

func TestAppStateCopy(t *testing.T) {
	ts := time.Now()
	workflow := &WorkflowState{
		ID:        "workflow-123",
		Task:      "test task",
		Type:      WorkflowSequential,
		Status:    WorkflowRunning,
		StartTime: ts,
	}

	original := AppState{
		CurrentScreen: ScreenMonitor,
		FocusIndex:    2,
		TaskInput: TaskInputState{
			TaskDescription: "test",
		},
		ActiveWorkflow: workflow,
		Config: ConfigState{
			EditMode: true,
		},
		History: HistoryState{
			SearchQuery: "query",
		},
		Agents: []AgentInfo{
			{Name: "agent1"},
		},
		StatusMessage: "status",
		Error:         "error",
	}

	copy := original.Copy()

	if copy.CurrentScreen != original.CurrentScreen {
		t.Errorf("CurrentScreen not copied correctly")
	}
	if copy.FocusIndex != original.FocusIndex {
		t.Errorf("FocusIndex not copied correctly")
	}
	if copy.TaskInput.TaskDescription != original.TaskInput.TaskDescription {
		t.Errorf("TaskInput not copied correctly")
	}
	if copy.ActiveWorkflow == nil {
		t.Errorf("ActiveWorkflow is nil")
	}
	if copy.ActiveWorkflow.ID != original.ActiveWorkflow.ID {
		t.Errorf("ActiveWorkflow not copied correctly")
	}
	if len(copy.Agents) != len(original.Agents) {
		t.Errorf("Agents length mismatch")
	}

	// Verify deep copy
	copy.Agents[0].Name = "modified"
	if original.Agents[0].Name == "modified" {
		t.Errorf("Agents is not a deep copy")
	}

	copy.ActiveWorkflow.Task = "modified"
	if original.ActiveWorkflow.Task == "modified" {
		t.Errorf("ActiveWorkflow is not a deep copy")
	}
}

func TestNewAppState(t *testing.T) {
	state := NewAppState()

	if state.CurrentScreen != ScreenHome {
		t.Errorf("Expected initial screen to be ScreenHome, got %v", state.CurrentScreen)
	}
	if state.FocusIndex != 0 {
		t.Errorf("Expected initial FocusIndex to be 0, got %v", state.FocusIndex)
	}
	if state.TaskInput.WorkflowType != WorkflowSequential {
		t.Errorf("Expected default WorkflowType to be Sequential, got %v", state.TaskInput.WorkflowType)
	}
	if state.TaskInput.MaxIterations != 3 {
		t.Errorf("Expected default MaxIterations to be 3, got %v", state.TaskInput.MaxIterations)
	}
	if state.Agents == nil {
		t.Errorf("Expected Agents to be initialized to empty slice, got nil")
	}
	if len(state.Agents) != 0 {
		t.Errorf("Expected Agents to be empty, got %d agents", len(state.Agents))
	}
}