package screens

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

func TestHistoryScreenName(t *testing.T) {
	history := NewHistoryScreen(nil)
	if history.Name() != state.ScreenHistory {
		t.Errorf("Expected name to be ScreenHistory, got %v", history.Name())
	}
}

func TestHistoryScreenHelpText(t *testing.T) {
	history := NewHistoryScreen(nil)
	helpText := history.HelpText()
	if helpText == "" {
		t.Error("HelpText should not be empty")
	}
	// Should contain navigation hints
	if !strings.Contains(helpText, "arrows") {
		t.Error("HelpText should mention arrow navigation")
	}
	if !strings.Contains(helpText, "Enter") {
		t.Error("HelpText should mention Enter key")
	}
	if !strings.Contains(helpText, "search") {
		t.Error("HelpText should mention search")
	}
}

func TestHistoryScreenHelpTextSearching(t *testing.T) {
	history := NewHistoryScreen(nil)
	history.SetSearching(true)
	helpText := history.HelpText()
	if !strings.Contains(helpText, "Esc") {
		t.Error("HelpText in search mode should mention Esc")
	}
	if !strings.Contains(helpText, "confirm") {
		t.Error("HelpText in search mode should mention confirm")
	}
}

func TestHistoryScreenRenderBasic(t *testing.T) {
	history := NewHistoryScreen(nil)
	now := time.Now()
	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records: []state.WorkflowRecord{
				{
					ID:        "test-001",
					Task:      "AI trend analysis",
					Type:      state.WorkflowParallel,
					Status:    state.WorkflowCompleted,
					StartTime: now.Add(-1 * time.Hour),
					EndTime:   now,
					Duration:  time.Hour,
				},
				{
					ID:        "test-002",
					Task:      "Architecture comparison",
					Type:      state.WorkflowSequential,
					Status:    state.WorkflowCompleted,
					StartTime: now.Add(-2 * time.Hour),
					EndTime:   now.Add(-1 * time.Hour),
					Duration:  time.Hour,
				},
			},
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	output := history.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain title
	if !strings.Contains(output, "历史记录") {
		t.Error("Render should contain title '历史记录'")
	}

	// Should contain search label
	if !strings.Contains(output, "搜索") {
		t.Error("Render should contain search label")
	}

	// Should contain statistics
	if !strings.Contains(output, "统计") {
		t.Error("Render should contain statistics")
	}

	// Should contain workflow type
	if !strings.Contains(output, "Parallel") {
		t.Error("Render should contain workflow type Parallel")
	}
}

func TestHistoryScreenRenderEmptyRecords(t *testing.T) {
	history := NewHistoryScreen(nil)
	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{},
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	output := history.Render(model)
	if output == "" {
		t.Error("Render should not return empty string even with no records")
	}

	// Should contain title
	if !strings.Contains(output, "历史记录") {
		t.Error("Render should contain title even with no records")
	}

	// Should show empty message
	if !strings.Contains(output, "暂无历史记录") {
		t.Error("Render should show empty message when no records")
	}
}

func TestHistoryScreenRenderDetailView(t *testing.T) {
	history := NewHistoryScreen(nil)
	now := time.Now()
	record := state.WorkflowRecord{
		ID:          "test-detail-001",
		Task:        "Detailed task analysis",
		Type:        state.WorkflowDynamic,
		Status:      state.WorkflowCompleted,
		StartTime:   now.Add(-3 * time.Hour),
		EndTime:     now,
		Duration:    3 * time.Hour,
		FinalOutput: "This is the final output of the workflow",
		Agents: []state.AgentRecord{
			{
				Name:     "researcher",
				Output:   "Research output here",
				Tokens:   1500,
				Duration: 30 * time.Minute,
			},
			{
				Name:     "analyst",
				Output:   "Analysis output here",
				Tokens:   2000,
				Duration: 45 * time.Minute,
			},
		},
	}

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{record},
			SelectedIndex: 0,
			ViewDetail:    true,
			DetailRecord:  &record,
		},
	}

	output := history.Render(model)
	if output == "" {
		t.Error("Render should not return empty string for detail view")
	}

	// Should contain workflow ID
	if !strings.Contains(output, "test-detail") {
		t.Error("Render detail view should contain workflow ID")
	}

	// Should contain task
	if !strings.Contains(output, "Detailed task analysis") {
		t.Error("Render detail view should contain task")
	}

	// Should contain agent names
	if !strings.Contains(output, "researcher") {
		t.Error("Render detail view should contain researcher agent")
	}
	if !strings.Contains(output, "analyst") {
		t.Error("Render detail view should contain analyst agent")
	}

	// Should contain final output section
	if !strings.Contains(output, "最终输出") {
		t.Error("Render detail view should contain final output section")
	}
}

func TestHistoryScreenRenderFailedWorkflow(t *testing.T) {
	history := NewHistoryScreen(nil)
	now := time.Now()
	record := state.WorkflowRecord{
		ID:        "test-failed-001",
		Task:      "Failed task",
		Type:      state.WorkflowLoop,
		Status:    state.WorkflowFailed,
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now,
		Duration:  time.Hour,
		Error:     "Timeout exceeded",
	}

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{record},
			SelectedIndex: 0,
			ViewDetail:    true,
			DetailRecord:  &record,
		},
	}

	output := history.Render(model)

	// Should contain error section
	if !strings.Contains(output, "错误信息") {
		t.Error("Render should contain error section for failed workflow")
	}
	if !strings.Contains(output, "Timeout exceeded") {
		t.Error("Render should contain error message")
	}
}

func TestHistoryScreenUpdateArrowKeys(t *testing.T) {
	history := NewHistoryScreen(nil)
	now := time.Now()
	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records: []state.WorkflowRecord{
				{ID: "test-001", Task: "Task 1", StartTime: now},
				{ID: "test-002", Task: "Task 2", StartTime: now.Add(-1 * time.Hour)},
				{ID: "test-003", Task: "Task 3", StartTime: now.Add(-2 * time.Hour)},
			},
			SelectedIndex: 1,
			ViewDetail:    false,
		},
	}

	// Test pressing up arrow
	result := history.Update(tea.KeyMsg{Type: tea.KeyUp}, model)
	if result != nil {
		t.Error("Update for arrow keys should return nil (just updates selection)")
	}
	if model.History.SelectedIndex != 0 {
		t.Errorf("After pressing up, SelectedIndex should be 0, got %d", model.History.SelectedIndex)
	}

	// Test pressing up again (at top, should stay at 0)
	result = history.Update(tea.KeyMsg{Type: tea.KeyUp}, model)
	if model.History.SelectedIndex != 0 {
		t.Errorf("At top, pressing up should keep SelectedIndex at 0, got %d", model.History.SelectedIndex)
	}

	// Test pressing down arrow
	result = history.Update(tea.KeyMsg{Type: tea.KeyDown}, model)
	if model.History.SelectedIndex != 1 {
		t.Errorf("After pressing down, SelectedIndex should be 1, got %d", model.History.SelectedIndex)
	}

	// Test pressing down again
	result = history.Update(tea.KeyMsg{Type: tea.KeyDown}, model)
	if model.History.SelectedIndex != 2 {
		t.Errorf("After pressing down, SelectedIndex should be 2, got %d", model.History.SelectedIndex)
	}

	// Test pressing down at bottom (should stay at max)
	result = history.Update(tea.KeyMsg{Type: tea.KeyDown}, model)
	if model.History.SelectedIndex != 2 {
		t.Errorf("At bottom, pressing down should keep SelectedIndex at 2, got %d", model.History.SelectedIndex)
	}
}

func TestHistoryScreenUpdateEnterDetailView(t *testing.T) {
	history := NewHistoryScreen(nil)
	record := state.WorkflowRecord{
		ID:     "test-enter-001",
		Task:   "Test task",
		Status: state.WorkflowCompleted,
	}

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{record},
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	// Test pressing Enter to view detail
	result := history.Update(tea.KeyMsg{Type: tea.KeyEnter}, model)

	if result != nil {
		t.Error("Update for Enter in normal mode should return nil")
	}
	if !model.History.ViewDetail {
		t.Error("ViewDetail should be true after pressing Enter")
	}
	if model.History.DetailRecord == nil {
		t.Error("DetailRecord should be set after pressing Enter")
	}
	if model.History.DetailRecord.ID != "test-enter-001" {
		t.Errorf("DetailRecord.ID should be test-enter-001, got %s", model.History.DetailRecord.ID)
	}
}

func TestHistoryScreenUpdateEscBackToHome(t *testing.T) {
	history := NewHistoryScreen(nil)
	appState := state.NewAppState()
	appState.CurrentScreen = state.ScreenHistory
	model := &appState

	// Test pressing Esc to return to Home
	result := history.Update(tea.KeyMsg{Type: tea.KeyEsc}, model)

	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenHome {
		t.Errorf("Expected ScreenHome, got %v", switchMsg.Screen)
	}
}

func TestHistoryScreenUpdateSearchMode(t *testing.T) {
	history := NewHistoryScreen(nil)
	appState := state.NewAppState()
	appState.CurrentScreen = state.ScreenHistory
	model := &appState

	// Test pressing 's' to enter search mode
	result := history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}, model)

	if result != nil {
		t.Error("Update for 's' should return nil (just changes mode)")
	}
	if !history.IsSearching() {
		t.Error("Should be in search mode after pressing 's'")
	}
}

func TestHistoryScreenUpdateSearchTyping(t *testing.T) {
	history := NewHistoryScreen(nil)
	history.SetSearching(true) // Start in search mode

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{},
			SearchQuery:   "",
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	// Test typing 'a'
	_ = history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, model)
	if model.History.SearchQuery != "a" {
		t.Errorf("SearchQuery should be 'a' after typing, got %s", model.History.SearchQuery)
	}

	// Test typing 'b'
	_ = history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}, model)
	if model.History.SearchQuery != "ab" {
		t.Errorf("SearchQuery should be 'ab' after typing, got %s", model.History.SearchQuery)
	}
}

func TestHistoryScreenUpdateSearchBackspace(t *testing.T) {
	history := NewHistoryScreen(nil)
	history.SetSearching(true)

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{},
			SearchQuery:   "test",
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	// Test pressing backspace
	_ = history.Update(tea.KeyMsg{Type: tea.KeyBackspace}, model)
	if model.History.SearchQuery != "tes" {
		t.Errorf("SearchQuery should be 'tes' after backspace, got %s", model.History.SearchQuery)
	}

	// Test pressing backspace again
	_ = history.Update(tea.KeyMsg{Type: tea.KeyBackspace}, model)
	if model.History.SearchQuery != "te" {
		t.Errorf("SearchQuery should be 'te' after backspace, got %s", model.History.SearchQuery)
	}
}

func TestHistoryScreenUpdateSearchConfirm(t *testing.T) {
	history := NewHistoryScreen(nil)
	history.SetSearching(true)

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{},
			SearchQuery:   "test query",
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	// Test pressing Enter to confirm search
	result := history.Update(tea.KeyMsg{Type: tea.KeyEnter}, model)

	if result != nil {
		t.Error("Update for Enter in search mode should return nil")
	}
	if history.IsSearching() {
		t.Error("Should not be in search mode after pressing Enter")
	}
}

func TestHistoryScreenUpdateSearchCancel(t *testing.T) {
	history := NewHistoryScreen(nil)
	history.SetSearching(true)

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{},
			SearchQuery:   "test query",
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	// Test pressing Esc to cancel search
	result := history.Update(tea.KeyMsg{Type: tea.KeyEsc}, model)

	if result != nil {
		t.Error("Update for Esc in search mode should return nil")
	}
	if history.IsSearching() {
		t.Error("Should not be in search mode after pressing Esc")
	}
}

func TestHistoryScreenUpdateDetailViewEsc(t *testing.T) {
	history := NewHistoryScreen(nil)
	record := state.WorkflowRecord{
		ID:     "test-detail-esc",
		Task:   "Test",
		Status: state.WorkflowCompleted,
	}

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{record},
			SelectedIndex: 0,
			ViewDetail:    true,
			DetailRecord:  &record,
		},
	}

	// Test pressing Esc to exit detail view
	result := history.Update(tea.KeyMsg{Type: tea.KeyEsc}, model)

	if result != nil {
		t.Error("Update for Esc in detail view should return nil")
	}
	if model.History.ViewDetail {
		t.Error("ViewDetail should be false after pressing Esc in detail view")
	}
	if model.History.DetailRecord != nil {
		t.Error("DetailRecord should be nil after pressing Esc in detail view")
	}
}

func TestHistoryScreenUpdateDetailViewEnter(t *testing.T) {
	history := NewHistoryScreen(nil)
	record := state.WorkflowRecord{
		ID:     "test-detail-enter",
		Task:   "Test",
		Status: state.WorkflowCompleted,
	}

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{record},
			SelectedIndex: 0,
			ViewDetail:    true,
			DetailRecord:  &record,
		},
	}

	// Test pressing Enter to exit detail view
	history.Update(tea.KeyMsg{Type: tea.KeyEnter}, model)

	if model.History.ViewDetail {
		t.Error("ViewDetail should be false after pressing Enter in detail view")
	}
}

func TestHistoryScreenUpdateQuit(t *testing.T) {
	history := NewHistoryScreen(nil)
	appState := state.NewAppState()
	appState.CurrentScreen = state.ScreenHistory
	model := &appState

	// Test pressing 'q' to quit
	result := history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, model)

	quitMsg, ok := result.(QuitMsg)
	if !ok {
		t.Errorf("Expected QuitMsg, got %T", result)
	}
	if !quitMsg.Quit {
		t.Error("QuitMsg.Quit should be true")
	}
}

func TestHistoryScreenUpdateRerun(t *testing.T) {
	history := NewHistoryScreen(nil)
	record := state.WorkflowRecord{
		ID:     "test-rerun-001",
		Task:   "Task to rerun",
		Type:   state.WorkflowParallel,
		Status: state.WorkflowCompleted,
	}

	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records:       []state.WorkflowRecord{record},
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	// Test pressing 'r' to re-run
	result := history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, model)

	rerunMsg, ok := result.(RerunWorkflowMsg)
	if !ok {
		t.Errorf("Expected RerunWorkflowMsg, got %T", result)
	}
	if rerunMsg.Task != "Task to rerun" {
		t.Errorf("RerunWorkflowMsg.Task should be 'Task to rerun', got %s", rerunMsg.Task)
	}
	if rerunMsg.WorkflowType != state.WorkflowParallel {
		t.Errorf("RerunWorkflowMsg.WorkflowType should be Parallel, got %v", rerunMsg.WorkflowType)
	}
}

func TestHistoryScreenUpdateHelp(t *testing.T) {
	history := NewHistoryScreen(nil)
	appState := state.NewAppState()
	appState.CurrentScreen = state.ScreenHistory
	model := &appState

	// Test pressing '?' for help
	result := history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, model)

	helpMsg, ok := result.(HelpMsg)
	if !ok {
		t.Errorf("Expected HelpMsg, got %T", result)
	}
	if helpMsg.Text == "" {
		t.Error("HelpMsg.Text should not be empty")
	}
}

func TestHistoryScreenUpdateUnknownKey(t *testing.T) {
	history := NewHistoryScreen(nil)
	appState := state.NewAppState()
	appState.CurrentScreen = state.ScreenHistory
	model := &appState

	// Test pressing an unknown key
	result := history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, model)

	// Should return nil for unknown keys
	if result != nil {
		t.Errorf("Expected nil for unknown key, got %T", result)
	}
}

func TestHistoryScreenStatistics(t *testing.T) {
	history := NewHistoryScreen(nil)
	now := time.Now()
	model := &state.AppState{
		CurrentScreen: state.ScreenHistory,
		History: state.HistoryState{
			Records: []state.WorkflowRecord{
				{ID: "001", Status: state.WorkflowCompleted, StartTime: now},
				{ID: "002", Status: state.WorkflowCompleted, StartTime: now},
				{ID: "003", Status: state.WorkflowFailed, StartTime: now},
				{ID: "004", Status: state.WorkflowCompleted, StartTime: now},
				{ID: "005", Status: state.WorkflowFailed, StartTime: now},
			},
			SelectedIndex: 0,
			ViewDetail:    false,
		},
	}

	output := history.Render(model)

	// Should show total count
	if !strings.Contains(output, "总计") {
		t.Error("Render should show total count")
	}

	// Should show success count
	if !strings.Contains(output, "成功") {
		t.Error("Render should show success count")
	}

	// Should show failed count
	if !strings.Contains(output, "失败") {
		t.Error("Render should show failed count")
	}
}

func TestHistoryScreenFormatTime(t *testing.T) {
	history := NewHistoryScreen(nil)

	// Test recent time (within 24 hours)
	recent := time.Now().Add(-1 * time.Hour)
	recentStr := history.formatTime(recent)
	if !strings.Contains(recentStr, ":") {
		t.Errorf("Recent time should show HH:MM format, got %s", recentStr)
	}

	// Test older time (more than 24 hours)
	older := time.Now().Add(-48 * time.Hour)
	olderStr := history.formatTime(older)
	if !strings.Contains(olderStr, "天前") && !strings.Contains(olderStr, "-") {
		t.Errorf("Older time should show days ago or date format, got %s", olderStr)
	}
}

func TestHistoryScreenFormatDuration(t *testing.T) {
	history := NewHistoryScreen(nil)

	// Test milliseconds
	ms := 500 * time.Millisecond
	msStr := history.formatDuration(ms)
	if !strings.Contains(msStr, "ms") {
		t.Errorf("Duration < 1s should show ms, got %s", msStr)
	}

	// Test seconds
	sec := 5 * time.Second
	secStr := history.formatDuration(sec)
	if !strings.Contains(secStr, "s") {
		t.Errorf("Duration < 1min should show s, got %s", secStr)
	}

	// Test minutes
	min := 5 * time.Minute
	minStr := history.formatDuration(min)
	if !strings.Contains(minStr, "m") {
		t.Errorf("Duration >= 1min should show m, got %s", minStr)
	}
}