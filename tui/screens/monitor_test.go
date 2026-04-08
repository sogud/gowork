package screens

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

func TestMonitorScreenName(t *testing.T) {
	monitor := NewMonitorScreen()
	if monitor.Name() != state.ScreenMonitor {
		t.Errorf("Expected name to be ScreenMonitor, got %v", monitor.Name())
	}
}

func TestMonitorScreenHelpText(t *testing.T) {
	monitor := NewMonitorScreen()
	helpText := monitor.HelpText()
	if helpText == "" {
		t.Error("HelpText should not be empty")
	}
	// Should contain keyboard hints
	if !strings.Contains(helpText, "Tab") {
		t.Error("HelpText should mention Tab navigation")
	}
	if !strings.Contains(helpText, "Esc") {
		t.Error("HelpText should mention Esc for back")
	}
	if !strings.Contains(helpText, "Enter") {
		t.Error("HelpText should mention Enter for details")
	}
}

func TestMonitorScreenRenderNoWorkflow(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
	}

	output := monitor.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should indicate no active workflow
	if !strings.Contains(output, "No active workflow") {
		t.Error("Render should indicate no active workflow")
	}
}

func TestMonitorScreenRenderWithWorkflow(t *testing.T) {
	monitor := NewMonitorScreen()

	startTime := time.Now().Add(-45 * time.Second) // Started 45 seconds ago
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    0,
		ActiveWorkflow: &state.WorkflowState{
			ID:        "wf-001",
			Task:      "Research and summarize AI agent trends",
			Type:      state.WorkflowParallel,
			Status:    state.WorkflowRunning,
			StartTime: startTime,
			AgentExecutions: []state.AgentExecution{
				{
					Name:   "researcher",
					Status: state.AgentCompleted,
					Output: "Collected 15 relevant papers...",
					ToolCalls: []state.ToolCallInfo{
						{ToolName: "web_search"},
						{ToolName: "web_search"},
						{ToolName: "web_search"},
					},
				},
				{
					Name:   "analyst",
					Status: state.AgentRunning,
					Output: "Analyzing data trends...",
				},
				{
					Name:   "writer",
					Status: state.AgentWaiting,
				},
				{
					Name:   "reviewer",
					Status: state.AgentWaiting,
				},
			},
		},
	}

	output := monitor.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain workflow header
	if !strings.Contains(output, "Execution Monitor") {
		t.Error("Render should contain Execution Monitor title")
	}

	// Should contain workflow type
	if !strings.Contains(output, "parallel") {
		t.Error("Render should contain workflow type (parallel)")
	}

	// Should contain task description
	if !strings.Contains(output, "Research") {
		t.Error("Render should contain task description")
	}

	// Should contain agent names
	if !strings.Contains(output, "researcher") {
		t.Error("Render should contain researcher agent")
	}
	if !strings.Contains(output, "analyst") {
		t.Error("Render should contain analyst agent")
	}

	// Should contain status text
	if !strings.Contains(output, "Completed") {
		t.Error("Render should show Completed status")
	}
	if !strings.Contains(output, "Running") {
		t.Error("Render should show Running status")
	}
	if !strings.Contains(output, "Waiting") {
		t.Error("Render should show Waiting status")
	}

	// Should contain progress bar characters
	if !strings.Contains(output, "█") && !strings.Contains(output, "░") {
		t.Error("Render should contain progress bar characters")
	}
}

func TestMonitorScreenRenderWithOutputStream(t *testing.T) {
	monitor := NewMonitorScreen()

	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    1, // Focus on analyst
		ActiveWorkflow: &state.WorkflowState{
			ID:        "wf-002",
			Task:      "Test task",
			Type:      state.WorkflowSequential,
			Status:    state.WorkflowRunning,
			StartTime: time.Now(),
			AgentExecutions: []state.AgentExecution{
				{
					Name:   "agent1",
					Status: state.AgentCompleted,
					Output: "Output from agent 1",
				},
				{
					Name:   "agent2",
					Status: state.AgentRunning,
					Output: "Based on collected data, found three main trends:\n1. Multi-agent frameworks...\n2. Tool call enhancement...",
				},
			},
		},
	}

	output := monitor.Render(model)

	// Should show output stream for focused agent
	if !strings.Contains(output, "Output Stream") {
		t.Error("Render should contain Output Stream section")
	}
	if !strings.Contains(output, "agent2") {
		t.Error("Render should show agent2 in output stream header")
	}
}

func TestMonitorScreenRenderToolCalls(t *testing.T) {
	monitor := NewMonitorScreen()

	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    0,
		ActiveWorkflow: &state.WorkflowState{
			ID:        "wf-003",
			Task:      "Test task",
			Type:      state.WorkflowDynamic,
			Status:    state.WorkflowRunning,
			StartTime: time.Now(),
			AgentExecutions: []state.AgentExecution{
				{
					Name:   "researcher",
					Status: state.AgentRunning,
					ToolCalls: []state.ToolCallInfo{
						{ToolName: "web_search"},
						{ToolName: "web_search"},
						{ToolName: "file_read"},
					},
				},
			},
		},
	}

	output := monitor.Render(model)

	// Should show tool call summary
	if !strings.Contains(output, "web_search(2)") {
		t.Error("Render should show tool call count (web_search(2))")
	}
	if !strings.Contains(output, "file_read(1)") {
		t.Error("Render should show tool call count (file_read(1))")
	}
}

func TestMonitorScreenUpdateTabKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    0,
		ActiveWorkflow: &state.WorkflowState{
			AgentExecutions: []state.AgentExecution{
				{Name: "agent1"},
				{Name: "agent2"},
				{Name: "agent3"},
			},
		},
	}

	// Press Tab
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyTab}, model)

	// Should return nil (FocusIndex is updated directly in model)
	if result != nil {
		t.Errorf("Expected nil for Tab key, got %T", result)
	}

	// FocusIndex should have changed
	if model.FocusIndex != 1 {
		t.Errorf("Expected FocusIndex to be 1, got %d", model.FocusIndex)
	}

	// Press Tab again to wrap around
	monitor.Update(tea.KeyMsg{Type: tea.KeyTab}, model)
	monitor.Update(tea.KeyMsg{Type: tea.KeyTab}, model)

	if model.FocusIndex != 0 {
		t.Errorf("Expected FocusIndex to wrap to 0, got %d", model.FocusIndex)
	}
}

func TestMonitorScreenUpdateShiftTabKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    0,
		ActiveWorkflow: &state.WorkflowState{
			AgentExecutions: []state.AgentExecution{
				{Name: "agent1"},
				{Name: "agent2"},
				{Name: "agent3"},
			},
		},
	}

	// Press Shift+Tab (cycle backwards)
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyShiftTab}, model)

	if result != nil {
		t.Errorf("Expected nil for Shift+Tab key, got %T", result)
	}

	// FocusIndex should wrap to last agent
	if model.FocusIndex != 2 {
		t.Errorf("Expected FocusIndex to wrap to 2, got %d", model.FocusIndex)
	}
}

func TestMonitorScreenUpdateEnterKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    0,
		ActiveWorkflow: &state.WorkflowState{
			AgentExecutions: []state.AgentExecution{
				{Name: "agent1"},
			},
		},
	}

	// Initially not expanded
	if monitor.expanded {
		t.Error("Monitor should not be expanded initially")
	}

	// Press Enter
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyEnter}, model)

	// Should return nil (toggles expanded state)
	if result != nil {
		t.Errorf("Expected nil for Enter key, got %T", result)
	}

	// Should now be expanded
	if !monitor.expanded {
		t.Error("Monitor should be expanded after Enter")
	}

	// Press Enter again to collapse
	monitor.Update(tea.KeyMsg{Type: tea.KeyEnter}, model)

	if monitor.expanded {
		t.Error("Monitor should not be expanded after second Enter")
	}
}

func TestMonitorScreenUpdateEscKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
	}

	// Press Esc
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyEsc}, model)

	// Should return ScreenChangeMsg to go to Home
	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenHome {
		t.Errorf("Expected ScreenHome, got %v", switchMsg.Screen)
	}
}

func TestMonitorScreenUpdateQuitKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
	}

	// Press 'q'
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, model)

	quitMsg, ok := result.(QuitMsg)
	if !ok {
		t.Errorf("Expected QuitMsg, got %T", result)
	}
	if !quitMsg.Quit {
		t.Error("QuitMsg.Quit should be true")
	}
}

func TestMonitorScreenUpdateHelpKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
	}

	// Press '?'
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, model)

	helpMsg, ok := result.(HelpMsg)
	if !ok {
		t.Errorf("Expected HelpMsg, got %T", result)
	}
	if helpMsg.Text == "" {
		t.Error("HelpMsg.Text should not be empty")
	}
}

func TestMonitorScreenUpdateUnknownKey(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
	}

	// Press an unknown key
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, model)

	// Should return nil for unknown keys
	if result != nil {
		t.Errorf("Expected nil for unknown key, got %T", result)
	}
}

func TestMonitorScreenUpdateNoWorkflow(t *testing.T) {
	monitor := NewMonitorScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    0,
	}

	// Press Tab with no active workflow - should handle gracefully
	result := monitor.Update(tea.KeyMsg{Type: tea.KeyTab}, model)

	// Should return nil (no agents to cycle)
	if result != nil {
		t.Errorf("Expected nil for Tab with no workflow, got %T", result)
	}

	// FocusIndex should not change
	if model.FocusIndex != 0 {
		t.Errorf("FocusIndex should remain 0 with no workflow, got %d", model.FocusIndex)
	}
}

func TestMonitorScreenFormatElapsedTime(t *testing.T) {
	monitor := NewMonitorScreen()

	// Test various durations
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0 * time.Second, "00:00"},
		{30 * time.Second, "00:30"},
		{45 * time.Second, "00:45"},
		{90 * time.Second, "01:30"},
		{120 * time.Second, "02:00"},
		{5 * time.Minute, "05:00"},
	}

	for _, tt := range tests {
		wf := &state.WorkflowState{
			StartTime:   time.Now().Add(-tt.duration),
			ElapsedTime: tt.duration,
			Status:      state.WorkflowCompleted,
		}
		result := monitor.formatElapsedTime(wf)
		if result != tt.expected {
			t.Errorf("Expected %s for duration %v, got %s", tt.expected, tt.duration, result)
		}
	}
}

func TestMonitorScreenWrapText(t *testing.T) {
	monitor := NewMonitorScreen()

	// Test text wrapping
	shortText := "Short text"
	result := monitor.wrapText(shortText, 20)
	if result != shortText {
		t.Error("Short text should not be wrapped")
	}

	// Test long text
	longText := "This is a very long line that should be wrapped"
	result = monitor.wrapText(longText, 10)
	if !strings.Contains(result, "\n") {
		t.Error("Long text should contain newlines after wrapping")
	}

	// Test multi-line text
	multiLine := "Line 1\nLine 2\nLine 3"
	result = monitor.wrapText(multiLine, 20)
	if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") {
		t.Error("Multi-line text should preserve all lines")
	}
}

func TestMonitorScreenFocusIndicator(t *testing.T) {
	monitor := NewMonitorScreen()

	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		FocusIndex:    2, // Focus on third agent
		ActiveWorkflow: &state.WorkflowState{
			ID:        "wf-004",
			Task:      "Test",
			Type:      state.WorkflowParallel,
			Status:    state.WorkflowRunning,
			StartTime: time.Now(),
			AgentExecutions: []state.AgentExecution{
				{Name: "agent1", Status: state.AgentCompleted},
				{Name: "agent2", Status: state.AgentRunning},
				{Name: "agent3", Status: state.AgentWaiting},
			},
		},
	}

	output := monitor.Render(model)

	// Should contain focused agent name prominently
	if !strings.Contains(output, "agent3") {
		t.Error("Render should contain focused agent name (agent3)")
	}
}

func TestMonitorScreenCompletedWorkflow(t *testing.T) {
	monitor := NewMonitorScreen()

	startTime := time.Now().Add(-2 * time.Minute)
	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		ActiveWorkflow: &state.WorkflowState{
			ID:          "wf-005",
			Task:        "Completed task",
			Type:        state.WorkflowSequential,
			Status:      state.WorkflowCompleted,
			StartTime:   startTime,
			ElapsedTime: 2 * time.Minute,
			FinalOutput: "Final result of the workflow",
			AgentExecutions: []state.AgentExecution{
				{Name: "agent1", Status: state.AgentCompleted},
				{Name: "agent2", Status: state.AgentCompleted},
			},
		},
	}

	output := monitor.Render(model)

	// Should show Completed status
	if !strings.Contains(output, "Completed") {
		t.Error("Render should show workflow Completed status")
	}

	// Should show elapsed time
	if !strings.Contains(output, "02:00") {
		t.Error("Render should show elapsed time (02:00)")
	}
}

func TestMonitorScreenFailedWorkflow(t *testing.T) {
	monitor := NewMonitorScreen()

	model := &state.AppState{
		CurrentScreen: state.ScreenMonitor,
		ActiveWorkflow: &state.WorkflowState{
			ID:        "wf-006",
			Task:      "Failed task",
			Type:      state.WorkflowDynamic,
			Status:    state.WorkflowFailed,
			StartTime: time.Now().Add(-30 * time.Second),
			AgentExecutions: []state.AgentExecution{
				{Name: "agent1", Status: state.AgentCompleted},
				{Name: "agent2", Status: state.AgentFailed, Error: "API timeout"},
			},
		},
	}

	output := monitor.Render(model)

	// Should show Failed status
	if !strings.Contains(output, "Failed") {
		t.Error("Render should show workflow Failed status")
	}
}