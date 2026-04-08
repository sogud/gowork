package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

func TestConfigScreenName(t *testing.T) {
	config := NewConfigScreen()
	if config.Name() != state.ScreenConfig {
		t.Errorf("Expected name to be ScreenConfig, got %v", config.Name())
	}
}

func TestConfigScreenHelpText(t *testing.T) {
	config := NewConfigScreen()
	helpText := config.HelpText()
	if helpText == "" {
		t.Error("HelpText should not be empty")
	}
	// Should contain navigation hints
	if !strings.Contains(helpText, "1-4") {
		t.Error("HelpText should mention 1-4 navigation")
	}
	if !strings.Contains(helpText, "Save") {
		t.Error("HelpText should mention Save")
	}
	if !strings.Contains(helpText, "Esc") {
		t.Error("HelpText should mention Esc")
	}
}

func TestConfigScreenRenderBasic(t *testing.T) {
	config := NewConfigScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenConfig,
		Config: state.ConfigState{
			ModelProvider: state.ModelProviderConfig{
				Type:      "ollama",
				ModelName: "gemma4",
				BaseURL:   "http://localhost:11434",
				Timeout:   60,
			},
			Agents: []state.AgentConfigEntry{
				{Name: "researcher", Enabled: true, Description: "研究员"},
				{Name: "analyst", Enabled: true, Description: "分析师"},
				{Name: "writer", Enabled: true, Description: "撰写者"},
				{Name: "reviewer", Enabled: true, Description: "审核员"},
				{Name: "coordinator", Enabled: true, Description: "协调员"},
			},
			WorkflowDefaults: state.WorkflowDefaultsConfig{
				DefaultType: "sequential",
				Timeout:     300,
				MaxIter:     3,
			},
		},
	}

	output := config.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain title
	if !strings.Contains(output, "配置管理") {
		t.Error("Render should contain title '配置管理'")
	}

	// Should contain section headers
	if !strings.Contains(output, "[1] 模型配置") {
		t.Error("Render should contain '[1] 模型配置'")
	}
	if !strings.Contains(output, "[2] 智能体管理") {
		t.Error("Render should contain '[2] 智能体管理'")
	}
	if !strings.Contains(output, "[3] 工具配置") {
		t.Error("Render should contain '[3] 工具配置'")
	}
	if !strings.Contains(output, "[4] 工作流默认设置") {
		t.Error("Render should contain '[4] 工作流默认设置'")
	}

	// Should contain model info
	if !strings.Contains(output, "ollama") {
		t.Error("Render should contain model provider 'ollama'")
	}
	if !strings.Contains(output, "gemma4") {
		t.Error("Render should contain model name 'gemma4'")
	}

	// Should contain agent names
	if !strings.Contains(output, "researcher") {
		t.Error("Render should contain 'researcher' agent")
	}
	if !strings.Contains(output, "analyst") {
		t.Error("Render should contain 'analyst' agent")
	}
}

func TestConfigScreenRenderEmptyConfig(t *testing.T) {
	config := NewConfigScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenConfig,
		Config:        state.ConfigState{},
	}

	output := config.Render(model)
	if output == "" {
		t.Error("Render should not return empty string even with empty config")
	}

	// Should still contain section headers
	if !strings.Contains(output, "配置管理") {
		t.Error("Render should contain title '配置管理'")
	}

	// Should show default agents
	if !strings.Contains(output, "researcher") {
		t.Error("Render should show default 'researcher' agent")
	}
}

func TestConfigScreenRenderWithTools(t *testing.T) {
	config := NewConfigScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenConfig,
		Config: state.ConfigState{
			Tools: []state.ToolConfigEntry{
				{Name: "web_search", Enabled: true, Description: "Web Search"},
				{Name: "file_read", Enabled: false, Description: "File Reader"},
			},
		},
	}

	output := config.Render(model)
	if output == "" {
		t.Error("Render should not return empty string")
	}

	// Should contain tool names
	if !strings.Contains(output, "web_search") {
		t.Error("Render should contain 'web_search' tool")
	}
	if !strings.Contains(output, "file_read") {
		t.Error("Render should contain 'file_read' tool")
	}
}

func TestConfigScreenUpdateNavigateSections(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig

	// Test pressing '1' to select model section
	result := config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}, &model)
	if model.FocusIndex != int(ConfigSectionModel) {
		t.Errorf("FocusIndex should be %d (ConfigSectionModel), got %d", ConfigSectionModel, model.FocusIndex)
	}
	if result != nil {
		t.Error("Update should return nil for section selection")
	}

	// Test pressing '2' to select agents section
	result = config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, &model)
	if model.FocusIndex != int(ConfigSectionAgents) {
		t.Errorf("FocusIndex should be %d (ConfigSectionAgents), got %d", ConfigSectionAgents, model.FocusIndex)
	}

	// Test pressing '3' to select tools section
	result = config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}, &model)
	if model.FocusIndex != int(ConfigSectionTools) {
		t.Errorf("FocusIndex should be %d (ConfigSectionTools), got %d", ConfigSectionTools, model.FocusIndex)
	}

	// Test pressing '4' to select workflow section
	result = config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}, &model)
	if model.FocusIndex != int(ConfigSectionWorkflow) {
		t.Errorf("FocusIndex should be %d (ConfigSectionWorkflow), got %d", ConfigSectionWorkflow, model.FocusIndex)
	}
}

func TestConfigScreenUpdateNavigateUpDown(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig
	model.FocusIndex = 0

	// Test pressing down to navigate to next section
	config.Update(tea.KeyMsg{Type: tea.KeyDown}, &model)
	if model.FocusIndex != 1 {
		t.Errorf("FocusIndex should be 1 after down key, got %d", model.FocusIndex)
	}

	// Navigate to last section
	model.FocusIndex = int(ConfigSectionWorkflow)

	// Test that down doesn't go past last section
	config.Update(tea.KeyMsg{Type: tea.KeyDown}, &model)
	if model.FocusIndex != int(ConfigSectionWorkflow) {
		t.Errorf("FocusIndex should stay at %d (ConfigSectionWorkflow), got %d", ConfigSectionWorkflow, model.FocusIndex)
	}

	// Test pressing up to navigate to previous section
	config.Update(tea.KeyMsg{Type: tea.KeyUp}, &model)
	if model.FocusIndex != int(ConfigSectionTools) {
		t.Errorf("FocusIndex should be %d (ConfigSectionTools) after up key, got %d", ConfigSectionTools, model.FocusIndex)
	}

	// Navigate to first section
	model.FocusIndex = 0

	// Test that up doesn't go before first section
	config.Update(tea.KeyMsg{Type: tea.KeyUp}, &model)
	if model.FocusIndex != 0 {
		t.Errorf("FocusIndex should stay at 0, got %d", model.FocusIndex)
	}
}

func TestConfigScreenUpdateEscape(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig

	// Test pressing Escape to go back to home screen
	result := config.Update(tea.KeyMsg{Type: tea.KeyEsc}, &model)

	switchMsg, ok := result.(ScreenChangeMsg)
	if !ok {
		t.Errorf("Expected ScreenChangeMsg, got %T", result)
	}
	if switchMsg.Screen != state.ScreenHome {
		t.Errorf("Expected ScreenHome, got %v", switchMsg.Screen)
	}
}

func TestConfigScreenUpdateSave(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig

	// Test pressing 's' to save configuration
	result := config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}, &model)

	saveMsg, ok := result.(SaveConfigMsg)
	if !ok {
		t.Errorf("Expected SaveConfigMsg, got %T", result)
	}
	_ = saveMsg // SaveConfigMsg is empty, just verify type
}

func TestConfigScreenUpdateToggleAgent(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig
	model.FocusIndex = int(ConfigSectionAgents)
	model.Config.Agents = []state.AgentConfigEntry{
		{Name: "researcher", Enabled: true, Description: "研究员"},
		{Name: "analyst", Enabled: true, Description: "分析师"},
	}

	// Test pressing Space to toggle agent
	_ = config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, &model)

	// First agent should be toggled off
	if model.Config.Agents[0].Enabled != false {
		t.Error("First agent should be toggled off after Space key")
	}

	// Toggle again
	_ = config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, &model)

	// First agent should be toggled back on
	if model.Config.Agents[0].Enabled != true {
		t.Error("First agent should be toggled on after second Space key")
	}
}

func TestConfigScreenUpdateToggleTool(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig
	model.FocusIndex = int(ConfigSectionTools)
	model.Config.Tools = []state.ToolConfigEntry{
		{Name: "web_search", Enabled: true, Description: "Web Search"},
	}

	// Test pressing Space to toggle tool
	_ = config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, &model)

	// First tool should be toggled off
	if model.Config.Tools[0].Enabled != false {
		t.Error("First tool should be toggled off after Space key")
	}
}

func TestConfigScreenUpdateEnterEditMode(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig
	model.FocusIndex = int(ConfigSectionModel)

	// Test pressing Enter to enter edit mode
	config.Update(tea.KeyMsg{Type: tea.KeyEnter}, &model)

	if !model.Config.EditMode {
		t.Error("EditMode should be true after Enter key")
	}
	if model.Config.EditField != 0 {
		t.Errorf("EditField should be 0, got %d", model.Config.EditField)
	}
}

func TestConfigScreenUpdateHelp(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig

	// Test pressing '?' for help
	result := config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, &model)

	helpMsg, ok := result.(HelpMsg)
	if !ok {
		t.Errorf("Expected HelpMsg, got %T", result)
	}
	if helpMsg.Text == "" {
		t.Error("HelpMsg.Text should not be empty")
	}
}

func TestConfigScreenUpdateUnknownKey(t *testing.T) {
	config := NewConfigScreen()
	model := state.NewAppState()
	model.CurrentScreen = state.ScreenConfig

	// Test pressing an unknown key
	result := config.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, &model)

	// Should return nil for unknown keys
	if result != nil {
		t.Errorf("Expected nil for unknown key, got %T", result)
	}
}

func TestConfigScreenCheckbox(t *testing.T) {
	config := NewConfigScreen()

	// Test checked checkbox
	checked := config.renderCheckbox(true)
	if !strings.Contains(checked, "\u2713") {
		t.Error("Checked checkbox should contain checkmark")
	}

	// Test unchecked checkbox
	unchecked := config.renderCheckbox(false)
	if strings.Contains(unchecked, "\u2713") {
		t.Error("Unchecked checkbox should not contain checkmark")
	}
}

func TestConfigScreenRenderWithTimeout(t *testing.T) {
	config := NewConfigScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenConfig,
		Config: state.ConfigState{
			ModelProvider: state.ModelProviderConfig{
				Type:      "ollama",
				ModelName: "gemma4",
				BaseURL:   "http://localhost:11434",
				Timeout:   120,
			},
		},
	}

	output := config.Render(model)
	if !strings.Contains(output, "120s") {
		t.Error("Render should show custom timeout value")
	}
}

func TestConfigScreenRenderWithZeroTimeout(t *testing.T) {
	config := NewConfigScreen()
	model := &state.AppState{
		CurrentScreen: state.ScreenConfig,
		Config: state.ConfigState{
			ModelProvider: state.ModelProviderConfig{
				Type:      "ollama",
				ModelName: "gemma4",
				BaseURL:   "http://localhost:11434",
				Timeout:   0,
			},
		},
	}

	output := config.Render(model)
	// Should show default timeout when 0
	if !strings.Contains(output, "60s") {
		t.Error("Render should show default timeout (60s) when timeout is 0")
	}
}
