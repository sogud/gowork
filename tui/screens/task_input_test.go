package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

func TestNewTaskInputScreen(t *testing.T) {
	taskInput := NewTaskInputScreen()
	if taskInput == nil {
		t.Fatal("NewTaskInputScreen should return non-nil screen")
	}

	if taskInput.Name() != state.ScreenTaskInput {
		t.Errorf("Expected name to be ScreenTaskInput, got %v", taskInput.Name())
	}
}

func TestTaskInputScreenHelpText(t *testing.T) {
	taskInput := NewTaskInputScreen()
	helpText := taskInput.HelpText()

	if helpText == "" {
		t.Error("HelpText should not be empty")
	}

	// Should contain navigation hints
	if !strings.Contains(helpText, "Tab") {
		t.Error("HelpText should mention Tab navigation")
	}
	if !strings.Contains(helpText, "Enter") {
		t.Error("HelpText should mention Enter submission")
	}
	if !strings.Contains(helpText, "Esc") {
		t.Error("HelpText should mention Esc to go back")
	}
}

func TestTaskInputScreenRenderBasic(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenTaskInput,
		TaskInput: state.TaskInputState{
			TaskDescription: "测试任务",
			WorkflowType:    state.WorkflowSequential,
			SelectedAgents:  []string{"researcher"},
			MaxIterations:   3,
			CursorField:     FieldTaskDescription,
		},
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentWaiting},
			{Name: "analyst", Status: state.AgentWaiting},
			{Name: "writer", Status: state.AgentWaiting},
		},
	}

	output := taskInput.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain title
	if !strings.Contains(output, "提交新任务") {
		t.Error("Render should contain title")
	}

	// Should contain task description section
	if !strings.Contains(output, "任务描述") {
		t.Error("Render should contain task description section")
	}

	// Should contain workflow type section
	if !strings.Contains(output, "工作流类型") {
		t.Error("Render should contain workflow type section")
	}

	// Should contain workflow options
	if !strings.Contains(output, "Sequential") {
		t.Error("Render should contain Sequential workflow option")
	}
	if !strings.Contains(output, "Parallel") {
		t.Error("Render should contain Parallel workflow option")
	}
	if !strings.Contains(output, "Loop") {
		t.Error("Render should contain Loop workflow option")
	}
	if !strings.Contains(output, "Dynamic") {
		t.Error("Render should contain Dynamic workflow option")
	}

	// Should contain agent selection section
	if !strings.Contains(output, "选择智能体") {
		t.Error("Render should contain agent selection section")
	}

	// Should contain agent names
	if !strings.Contains(output, "researcher") {
		t.Error("Render should contain researcher agent")
	}
}

func TestTaskInputScreenRenderLoopMode(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenTaskInput,
		TaskInput: state.TaskInputState{
			TaskDescription: "测试任务",
			WorkflowType:    state.WorkflowLoop,
			SelectedAgents:  []string{"researcher"},
			MaxIterations:   5,
			CursorField:     FieldTaskDescription,
		},
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentWaiting},
		},
	}

	output := taskInput.Render(model)

	// Should show max iterations section for Loop mode
	if !strings.Contains(output, "最大迭代次数") {
		t.Error("Render should contain max iterations section for Loop mode")
	}
}

func TestTaskInputScreenRenderDynamicMode(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenTaskInput,
		TaskInput: state.TaskInputState{
			TaskDescription: "测试任务",
			WorkflowType:    state.WorkflowDynamic,
			SelectedAgents:  []string{"researcher"},
			MaxIterations:   4,
			CursorField:     FieldTaskDescription,
		},
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentWaiting},
		},
	}

	output := taskInput.Render(model)

	// Should show max iterations section for Dynamic mode
	if !strings.Contains(output, "最大迭代次数") {
		t.Error("Render should contain max iterations section for Dynamic mode")
	}
}

func TestTaskInputScreenRenderNoAgents(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenTaskInput,
		TaskInput: state.TaskInputState{
			TaskDescription: "测试任务",
			WorkflowType:    state.WorkflowSequential,
			SelectedAgents:  []string{},
			CursorField:     FieldTaskDescription,
		},
		Agents: []state.AgentInfo{},
	}

	output := taskInput.Render(model)

	// Should handle no agents gracefully
	if !strings.Contains(output, "没有可用的智能体") {
		t.Error("Render should show message when no agents available")
	}
}

func TestTaskInputScreenUpdateEscape(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()

	// Test pressing Escape to return to home
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyEsc}, &model)

	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenHome {
		t.Errorf("Expected ScreenHome, got %v", switchMsg.Screen)
	}
}

func TestTaskInputScreenUpdateTabNavigation(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.CursorField = FieldTaskDescription

	// Test Tab to move to next field
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyTab}, &model)

	// Should not return a message, just update state
	if result != nil {
		t.Errorf("Expected nil for Tab, got %T", result)
	}

	// Should have moved to next field
	if model.TaskInput.CursorField != FieldWorkflowType {
		t.Errorf("Expected CursorField %d, got %d", FieldWorkflowType, model.TaskInput.CursorField)
	}

	// Test cycling through all fields
	for expected := FieldWorkflowType; expected <= FieldMaxIterations; expected++ {
		taskInput.Update(tea.KeyMsg{Type: tea.KeyTab}, &model)
		nextExpected := (expected + 1) % 4
		if model.TaskInput.CursorField != nextExpected {
			t.Errorf("After Tab, expected CursorField %d, got %d", nextExpected, model.TaskInput.CursorField)
		}
	}
}

func TestTaskInputScreenUpdateShiftTabNavigation(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.CursorField = FieldWorkflowType

	// Test Shift+Tab to move to previous field
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyShiftTab}, &model)

	if result != nil {
		t.Errorf("Expected nil for Shift+Tab, got %T", result)
	}

	// Should have moved to previous field
	if model.TaskInput.CursorField != FieldTaskDescription {
		t.Errorf("Expected CursorField %d, got %d", FieldTaskDescription, model.TaskInput.CursorField)
	}
}

func TestTaskInputScreenUpdateWorkflowTypeNavigation(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.CursorField = FieldWorkflowType
	model.TaskInput.WorkflowType = state.WorkflowSequential

	// Test Down arrow to change workflow type
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyDown}, &model)

	if result != nil {
		t.Errorf("Expected nil for Down arrow, got %T", result)
	}

	if model.TaskInput.WorkflowType != state.WorkflowParallel {
		t.Errorf("Expected WorkflowType Parallel, got %v", model.TaskInput.WorkflowType)
	}

	// Test Up arrow to change back
	taskInput.Update(tea.KeyMsg{Type: tea.KeyUp}, &model)

	if model.TaskInput.WorkflowType != state.WorkflowSequential {
		t.Errorf("Expected WorkflowType Sequential, got %v", model.TaskInput.WorkflowType)
	}
}

func TestTaskInputScreenUpdateAgentSelection(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.CursorField = FieldAgentSelection
	model.FocusIndex = 0
	model.Agents = []state.AgentInfo{
		{Name: "researcher", Status: state.AgentWaiting},
		{Name: "analyst", Status: state.AgentWaiting},
		{Name: "writer", Status: state.AgentWaiting},
	}
	model.TaskInput.SelectedAgents = []string{}

	// Test Down arrow to navigate agents
	taskInput.Update(tea.KeyMsg{Type: tea.KeyDown}, &model)
	if model.FocusIndex != 1 {
		t.Errorf("Expected FocusIndex 1, got %d", model.FocusIndex)
	}

	// Test Up arrow to navigate back
	taskInput.Update(tea.KeyMsg{Type: tea.KeyUp}, &model)
	if model.FocusIndex != 0 {
		t.Errorf("Expected FocusIndex 0, got %d", model.FocusIndex)
	}

	// Test Space to toggle agent selection
	taskInput.Update(tea.KeyMsg{Type: tea.KeySpace}, &model)
	if len(model.TaskInput.SelectedAgents) != 1 {
		t.Errorf("Expected 1 selected agent, got %d", len(model.TaskInput.SelectedAgents))
	}
	if model.TaskInput.SelectedAgents[0] != "researcher" {
		t.Errorf("Expected researcher selected, got %v", model.TaskInput.SelectedAgents[0])
	}

	// Test Space again to deselect
	taskInput.Update(tea.KeyMsg{Type: tea.KeySpace}, &model)
	if len(model.TaskInput.SelectedAgents) != 0 {
		t.Errorf("Expected 0 selected agents after deselection, got %d", len(model.TaskInput.SelectedAgents))
	}
}

func TestTaskInputScreenUpdateSubmitTask(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.TaskDescription = "测试任务描述"
	model.TaskInput.WorkflowType = state.WorkflowParallel
	model.TaskInput.SelectedAgents = []string{"researcher", "writer"}
	model.TaskInput.MaxIterations = 3
	model.Agents = []state.AgentInfo{
		{Name: "researcher", Status: state.AgentWaiting},
		{Name: "writer", Status: state.AgentWaiting},
	}

	// Test Enter to submit task
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyEnter}, &model)

	submitMsg, ok := result.(SubmitTaskMsg)
	if !ok {
		t.Errorf("Expected SubmitTaskMsg, got %T", result)
	}

	if submitMsg.TaskDescription != "测试任务描述" {
		t.Errorf("Expected task description '测试任务描述', got %v", submitMsg.TaskDescription)
	}
	if submitMsg.WorkflowType != state.WorkflowParallel {
		t.Errorf("Expected WorkflowType Parallel, got %v", submitMsg.WorkflowType)
	}
	if len(submitMsg.SelectedAgents) != 2 {
		t.Errorf("Expected 2 selected agents, got %d", len(submitMsg.SelectedAgents))
	}
	if submitMsg.MaxIterations != 3 {
		t.Errorf("Expected MaxIterations 3, got %d", submitMsg.MaxIterations)
	}
}

func TestTaskInputScreenUpdateSubmitTaskNoDescription(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.TaskDescription = "" // Empty description
	model.TaskInput.SelectedAgents = []string{"researcher"}

	// Test Enter with empty description - should not submit
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyEnter}, &model)

	if result != nil {
		t.Errorf("Expected nil for Enter with empty description, got %T", result)
	}
}

func TestTaskInputScreenUpdateSubmitTaskNoAgents(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.TaskDescription = "测试任务"
	model.TaskInput.SelectedAgents = []string{} // No agents selected

	// Test Enter with no agents selected - should not submit
	result := taskInput.Update(tea.KeyMsg{Type: tea.KeyEnter}, &model)

	if result != nil {
		t.Errorf("Expected nil for Enter with no agents, got %T", result)
	}
}

func TestTaskInputScreenTextInput(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.CursorField = FieldTaskDescription

	// Test typing in task description field
	result := taskInput.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'测', '试'},
	}, &model)

	// Should return a command for text input update
	if result == nil {
		t.Error("Expected non-nil result for text input")
	}

	// The task description should be updated
	if model.TaskInput.TaskDescription != "测试" {
		t.Errorf("Expected task description '测试', got '%v'", model.TaskInput.TaskDescription)
	}
}

func TestTaskInputScreenIterationsInput(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := state.NewAppState()
	model.TaskInput.CursorField = FieldMaxIterations
	model.TaskInput.WorkflowType = state.WorkflowLoop

	// Test typing digits in iterations field
	result := taskInput.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'5'},
	}, &model)

	// Should return a command for text input update
	if result == nil {
		t.Error("Expected non-nil result for iterations input")
	}
}

func TestTaskInputScreenWorkflowTypeRadioButtons(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenTaskInput,
		TaskInput: state.TaskInputState{
			WorkflowType: state.WorkflowSequential,
			CursorField:  FieldWorkflowType,
		},
	}

	output := taskInput.Render(model)

	// Selected workflow type should have filled circle
	if !strings.Contains(output, "●") {
		t.Error("Selected workflow type should show filled circle")
	}

	// Unselected types should have empty circle
	if !strings.Contains(output, "○") {
		t.Error("Unselected workflow types should show empty circle")
	}
}

func TestTaskInputScreenAgentCheckboxes(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenTaskInput,
		TaskInput: state.TaskInputState{
			SelectedAgents: []string{"researcher"},
			CursorField:    FieldAgentSelection,
		},
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentWaiting},
			{Name: "analyst", Status: state.AgentWaiting},
		},
		FocusIndex: 0,
	}

	output := taskInput.Render(model)

	// Selected agent should have checked box
	if !strings.Contains(output, "[✓]") {
		t.Error("Selected agent should show checked box")
	}

	// Unselected agent should have empty box
	if !strings.Contains(output, "[ ]") {
		t.Error("Unselected agent should show empty box")
	}
}

func TestTaskInputFieldConstants(t *testing.T) {
	// Verify field constants are correct
	if FieldTaskDescription != 0 {
		t.Errorf("Expected FieldTaskDescription = 0, got %d", FieldTaskDescription)
	}
	if FieldWorkflowType != 1 {
		t.Errorf("Expected FieldWorkflowType = 1, got %d", FieldWorkflowType)
	}
	if FieldAgentSelection != 2 {
		t.Errorf("Expected FieldAgentSelection = 2, got %d", FieldAgentSelection)
	}
	if FieldMaxIterations != 3 {
		t.Errorf("Expected FieldMaxIterations = 3, got %d", FieldMaxIterations)
	}
}

func TestTaskInputScreenFindWorkflowIndex(t *testing.T) {
	taskInput := NewTaskInputScreen()

	// Test finding indices for each workflow type
	tests := []struct {
		wt       state.WorkflowType
		expected int
	}{
		{state.WorkflowSequential, 0},
		{state.WorkflowParallel, 1},
		{state.WorkflowLoop, 2},
		{state.WorkflowDynamic, 3},
	}

	for _, tt := range tests {
		idx := taskInput.findWorkflowIndex(tt.wt)
		if idx != tt.expected {
			t.Errorf("For workflow type %v, expected index %d, got %d", tt.wt, tt.expected, idx)
		}
	}
}

func TestTaskInputScreenToggleAgentSelection(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		TaskInput: state.TaskInputState{
			SelectedAgents: []string{},
		},
	}

	// Test adding agent
	taskInput.toggleAgentSelection(model, "researcher")
	if len(model.TaskInput.SelectedAgents) != 1 {
		t.Error("Should have 1 selected agent after adding")
	}
	if model.TaskInput.SelectedAgents[0] != "researcher" {
		t.Error("First selected agent should be researcher")
	}

	// Test adding another agent
	taskInput.toggleAgentSelection(model, "writer")
	if len(model.TaskInput.SelectedAgents) != 2 {
		t.Error("Should have 2 selected agents")
	}

	// Test removing agent
	taskInput.toggleAgentSelection(model, "researcher")
	if len(model.TaskInput.SelectedAgents) != 1 {
		t.Error("Should have 1 selected agent after removing")
	}
	if model.TaskInput.SelectedAgents[0] != "writer" {
		t.Error("Remaining selected agent should be writer")
	}
}

func TestTaskInputScreenIsAgentSelected(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		TaskInput: state.TaskInputState{
			SelectedAgents: []string{"researcher", "writer"},
		},
	}

	// Test checking selected agents
	if !taskInput.isAgentSelected(model, "researcher") {
		t.Error("researcher should be selected")
	}
	if !taskInput.isAgentSelected(model, "writer") {
		t.Error("writer should be selected")
	}
	if taskInput.isAgentSelected(model, "analyst") {
		t.Error("analyst should not be selected")
	}
}

func TestTaskInputScreenGetAvailableAgents(t *testing.T) {
	taskInput := NewTaskInputScreen()
	model := &state.AppState{
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentWaiting},
			{Name: "analyst", Status: state.AgentWaiting},
			{Name: "writer", Status: state.AgentWaiting},
		},
	}

	agents := taskInput.getAvailableAgents(model)
	if len(agents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(agents))
	}

	expectedAgents := []string{"researcher", "analyst", "writer"}
	for i, expected := range expectedAgents {
		if agents[i] != expected {
			t.Errorf("Expected agent %s at index %d, got %s", expected, i, agents[i])
		}
	}
}

func TestTaskInputScreenWorkflowTypeLabel(t *testing.T) {
	taskInput := NewTaskInputScreen()

	tests := []struct {
		wt       state.WorkflowType
		expected string
	}{
		{state.WorkflowSequential, "Sequential"},
		{state.WorkflowParallel, "Parallel"},
		{state.WorkflowLoop, "Loop"},
		{state.WorkflowDynamic, "Dynamic"},
	}

	for _, tt := range tests {
		label := taskInput.workflowTypeLabel(tt.wt)
		if label != tt.expected {
			t.Errorf("For workflow type %v, expected label '%s', got '%s'", tt.wt, tt.expected, label)
		}
	}
}