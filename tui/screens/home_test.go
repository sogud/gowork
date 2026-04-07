package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

func TestHomeScreenName(t *testing.T) {
	home := NewHomeScreen()
	if home.Name() != state.ScreenHome {
		t.Errorf("Expected name to be ScreenHome, got %v", home.Name())
	}
}

func TestHomeScreenHelpText(t *testing.T) {
	home := NewHomeScreen()
	helpText := home.HelpText()
	if helpText == "" {
		t.Error("HelpText should not be empty")
	}
	// Should contain navigation hints
	if !strings.Contains(helpText, "1-4") {
		t.Error("HelpText should mention 1-4 navigation")
	}
}

func TestHomeScreenRenderBasic(t *testing.T) {
	home := NewHomeScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenHome,
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentWaiting},
			{Name: "analyst", Status: state.AgentWaiting},
			{Name: "writer", Status: state.AgentWaiting},
			{Name: "reviewer", Status: state.AgentWaiting},
			{Name: "coordinator", Status: state.AgentWaiting},
		},
		Config: state.ConfigState{
			ModelProvider: state.ModelProviderConfig{
				Type:      "ollama",
				ModelName: "gemma4",
			},
		},
	}

	output := home.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain title
	if !strings.Contains(output, "gowork") {
		t.Error("Render should contain gowork title")
	}

	// Should contain navigation menu items
	if !strings.Contains(output, "[1]") {
		t.Error("Render should contain [1] navigation item")
	}
	if !strings.Contains(output, "[2]") {
		t.Error("Render should contain [2] navigation item")
	}
	if !strings.Contains(output, "[3]") {
		t.Error("Render should contain [3] navigation item")
	}
	if !strings.Contains(output, "[4]") {
		t.Error("Render should contain [4] navigation item")
	}

	// Should contain agent names
	if !strings.Contains(output, "researcher") {
		t.Error("Render should contain researcher agent")
	}

	// Should contain model info
	if !strings.Contains(output, "ollama") || !strings.Contains(output, "gemma4") {
		t.Error("Render should contain model info (ollama/gemma4)")
	}
}

func TestHomeScreenRenderEmptyAgents(t *testing.T) {
	home := NewHomeScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenHome,
		Agents:        []state.AgentInfo{},
		Config: state.ConfigState{
			ModelProvider: state.ModelProviderConfig{
				Type:      "ollama",
				ModelName: "gemma4",
			},
		},
	}

	output := home.Render(model)
	if output == "" {
		t.Error("Render should not return empty string even with empty agents")
	}

	// Should still contain title and navigation
	if !strings.Contains(output, "gowork") {
		t.Error("Render should contain gowork title")
	}
}

func TestHomeScreenUpdateNavigateToTaskInput(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()
	model.Agents = []state.AgentInfo{
		{Name: "researcher", Status: state.AgentWaiting},
	}

	// Test pressing '1' to navigate to TaskInput
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}, &model)

	// Result should be a command to change screen
	if result == nil {
		t.Error("Update should return a result for key '1'")
	}

	// Check if it's a ScreenChangeMsg
	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenTaskInput {
		t.Errorf("Expected ScreenTaskInput, got %v", switchMsg.Screen)
	}
}

func TestHomeScreenUpdateNavigateToMonitor(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()
	model.Agents = []state.AgentInfo{
		{Name: "researcher", Status: state.AgentWaiting},
	}

	// Test pressing '2' to navigate to Monitor
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, &model)

	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenMonitor {
		t.Errorf("Expected ScreenMonitor, got %v", switchMsg.Screen)
	}
}

func TestHomeScreenUpdateNavigateToConfig(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()
	model.Agents = []state.AgentInfo{
		{Name: "researcher", Status: state.AgentWaiting},
	}

	// Test pressing '3' to navigate to Config
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}, &model)

	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenConfig {
		t.Errorf("Expected ScreenConfig, got %v", switchMsg.Screen)
	}
}

func TestHomeScreenUpdateNavigateToHistory(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()
	model.Agents = []state.AgentInfo{
		{Name: "researcher", Status: state.AgentWaiting},
	}

	// Test pressing '4' to navigate to History
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}, &model)

	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenHistory {
		t.Errorf("Expected ScreenHistory, got %v", switchMsg.Screen)
	}
}

func TestHomeScreenUpdateQuit(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()

	// Test pressing 'q' to quit
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, &model)

	quitMsg, ok := result.(QuitMsg)
	if !ok {
		t.Errorf("Expected QuitMsg, got %T", result)
	}
	if !quitMsg.Quit {
		t.Error("QuitMsg.Quit should be true")
	}
}

func TestHomeScreenUpdateHelp(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()

	// Test pressing '?' for help
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, &model)

	helpMsg, ok := result.(HelpMsg)
	if !ok {
		t.Errorf("Expected HelpMsg, got %T", result)
	}
	if helpMsg.Text == "" {
		t.Error("HelpMsg.Text should not be empty")
	}
}

func TestHomeScreenUpdateUnknownKey(t *testing.T) {
	home := NewHomeScreen()
	model := state.NewAppState()

	// Test pressing an unknown key
	result := home.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, &model)

	// Should return nil for unknown keys
	if result != nil {
		t.Errorf("Expected nil for unknown key, got %T", result)
	}
}

func TestHomeScreenAgentStatusDisplay(t *testing.T) {
	home := NewHomeScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenHome,
		Agents: []state.AgentInfo{
			{Name: "researcher", Status: state.AgentRunning},
			{Name: "analyst", Status: state.AgentCompleted},
			{Name: "writer", Status: state.AgentFailed},
			{Name: "reviewer", Status: state.AgentWaiting},
			{Name: "coordinator", Status: state.AgentWaiting},
		},
		Config: state.ConfigState{
			ModelProvider: state.ModelProviderConfig{
				Type:      "ollama",
				ModelName: "gemma4",
			},
		},
	}

	output := home.Render(model)

	// Should show status text (using Chinese status labels per spec)
	if !strings.Contains(output, "运行中") {
		t.Error("Render should show 运行中 (Running) status")
	}
	if !strings.Contains(output, "完成") {
		t.Error("Render should show 完成 (Completed) status")
	}
	if !strings.Contains(output, "失败") {
		t.Error("Render should show 失败 (Failed) status")
	}
	if !strings.Contains(output, "就绪") {
		t.Error("Render should show 就绪 (Waiting) status")
	}
}