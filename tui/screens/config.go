// Package screens provides individual screen implementations for the TUI.
package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/state"
	"github.com/sogud/gowork/tui/styles"
)

// ConfigSection represents the different configuration sections.
type ConfigSection int

const (
	ConfigSectionModel ConfigSection = iota
	ConfigSectionAgents
	ConfigSectionTools
	ConfigSectionWorkflow
)

// SaveConfigMsg is a message to save the configuration.
type SaveConfigMsg struct{}

// ConfigScreen is the configuration management screen.
type ConfigScreen struct {
	theme  styles.Theme
	layout styles.Layout
}

// NewConfigScreen creates a new ConfigScreen instance.
func NewConfigScreen() *ConfigScreen {
	theme := styles.DefaultTheme()
	layout := styles.NewLayout(theme)
	return &ConfigScreen{
		theme:  theme,
		layout: layout,
	}
}

// Name returns the screen's identifier.
func (c *ConfigScreen) Name() state.Screen {
	return state.ScreenConfig
}

// Render renders the Config screen content.
func (c *ConfigScreen) Render(model *state.AppState) string {
	var sections []string

	// Title section
	title := c.renderTitle()
	sections = append(sections, title)

	// Divider after title
	divider := c.layout.Divider(46)
	sections = append(sections, divider)

	// Model configuration section
	modelSection := c.renderModelSection(model)
	sections = append(sections, modelSection)

	// Agents section
	agentsSection := c.renderAgentsSection(model)
	sections = append(sections, agentsSection)

	// Tools section (collapsed)
	toolsSection := c.renderToolsSection(model)
	sections = append(sections, toolsSection)

	// Workflow defaults section (collapsed)
	workflowSection := c.renderWorkflowSection(model)
	sections = append(sections, workflowSection)

	return c.layout.JoinVertical(sections...)
}

// Update handles messages and updates the model.
func (c *ConfigScreen) Update(msg interface{}, model *state.AppState) interface{} {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyEsc:
			// Return to home screen
			return ScreenChangeMsg{Screen: state.ScreenHome}

		case tea.KeyRunes:
			switch string(m.Runes) {
			case "1":
				// Select model config section
				model.FocusIndex = int(ConfigSectionModel)
				return nil
			case "2":
				// Select agents section
				model.FocusIndex = int(ConfigSectionAgents)
				return nil
			case "3":
				// Select tools section
				model.FocusIndex = int(ConfigSectionTools)
				return nil
			case "4":
				// Select workflow section
				model.FocusIndex = int(ConfigSectionWorkflow)
				return nil
			case "s", "S":
				// Save configuration
				return SaveConfigMsg{}
			case " ":
				// Toggle checkbox (for agents/tools)
				return c.handleToggle(model)
			case "?":
				return HelpMsg{Text: c.HelpText()}
			}

		case tea.KeyUp:
			// Navigate up through sections
			if model.FocusIndex > 0 {
				model.FocusIndex--
			}
			return nil

		case tea.KeyDown:
			// Navigate down through sections
			if model.FocusIndex < int(ConfigSectionWorkflow) {
				model.FocusIndex++
			}
			return nil

		case tea.KeyEnter:
			// Enter edit mode for selected section
			model.Config.EditMode = true
			model.Config.EditField = 0
			return nil
		}
	}

	return nil
}

// HelpText returns the help text for this screen.
func (c *ConfigScreen) HelpText() string {
	return "1-4 Select section / Up/Down Navigate / Enter Edit / Space Toggle / s Save / Esc Back"
}

// handleToggle handles toggling checkboxes for agents or tools.
func (c *ConfigScreen) handleToggle(model *state.AppState) interface{} {
	section := ConfigSection(model.FocusIndex)

	switch section {
	case ConfigSectionAgents:
		// Toggle first agent's enabled state if agents exist
		if len(model.Config.Agents) > 0 {
			model.Config.Agents[0].Enabled = !model.Config.Agents[0].Enabled
		}
	case ConfigSectionTools:
		// Toggle first tool's enabled state if tools exist
		if len(model.Config.Tools) > 0 {
			model.Config.Tools[0].Enabled = !model.Config.Tools[0].Enabled
		}
	}

	return nil
}

// renderTitle renders the title section.
func (c *ConfigScreen) renderTitle() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(c.theme.Primary.ToLipgloss()).
		Bold(true).
		Padding(1, 0).
		Align(lipgloss.Center).
		Width(46)

	return titleStyle.Render("配置管理")
}

// renderModelSection renders the model configuration section.
func (c *ConfigScreen) renderModelSection(model *state.AppState) string {
	focused := model.FocusIndex == int(ConfigSectionModel)
	return c.renderSection(
		"[1] 模型配置",
		c.renderModelConfigContent(model),
		focused,
	)
}

// renderModelConfigContent renders the model configuration content.
func (c *ConfigScreen) renderModelConfigContent(model *state.AppState) string {
	config := model.Config.ModelProvider

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c.theme.Border.ToLipgloss()).
		Padding(0, 1).
		MarginLeft(2)

	keyStyle := lipgloss.NewStyle().
		Foreground(c.theme.TextMuted.ToLipgloss()).
		PaddingLeft(1)

	valueStyle := lipgloss.NewStyle().
		Foreground(c.theme.Text.ToLipgloss())

	provider := config.Type
	if provider == "" {
		provider = "未配置"
	}
	modelName := config.ModelName
	if modelName == "" {
		modelName = "未配置"
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "未配置"
	}
	timeout := fmt.Sprintf("%ds", config.Timeout)
	if config.Timeout == 0 {
		timeout = "60s"
	}

	lines := []string{
		fmt.Sprintf("%s %s", keyStyle.Render("提供者:"), valueStyle.Render(provider)),
		fmt.Sprintf("%s %s", keyStyle.Render("模型:"), valueStyle.Render(modelName)),
		fmt.Sprintf("%s %s", keyStyle.Render("地址:"), valueStyle.Render(baseURL)),
		fmt.Sprintf("%s %s", keyStyle.Render("超时:"), valueStyle.Render(timeout)),
	}

	content := strings.Join(lines, "\n")
	return boxStyle.Render(content)
}

// renderAgentsSection renders the agents management section.
func (c *ConfigScreen) renderAgentsSection(model *state.AppState) string {
	focused := model.FocusIndex == int(ConfigSectionAgents)
	return c.renderSection(
		"[2] 智能体管理",
		c.renderAgentsContent(model),
		focused,
	)
}

// renderAgentsContent renders the agents list content.
func (c *ConfigScreen) renderAgentsContent(model *state.AppState) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c.theme.Border.ToLipgloss()).
		Padding(0, 1).
		MarginLeft(2)

	if len(model.Config.Agents) == 0 {
		// Default agents if none configured
		defaultAgents := []struct {
			name        string
			description string
		}{
			{"researcher", "研究员"},
			{"analyst", "分析师"},
			{"writer", "撰写者"},
			{"reviewer", "审核员"},
			{"coordinator", "协调员"},
		}

		var lines []string
		for _, agent := range defaultAgents {
			checkbox := c.renderCheckbox(true)
			line := fmt.Sprintf("%s %-12s %s", checkbox, agent.name, agent.description)
			lines = append(lines, line)
		}
		content := strings.Join(lines, "\n")
		return boxStyle.Render(content)
	}

	var lines []string
	for _, agent := range model.Config.Agents {
		checkbox := c.renderCheckbox(agent.Enabled)
		desc := agent.Description
		if desc == "" {
			desc = agent.Name
		}
		line := fmt.Sprintf("%s %-12s %s", checkbox, agent.Name, desc)
		lines = append(lines, line)
	}
	content := strings.Join(lines, "\n")
	return boxStyle.Render(content)
}

// renderToolsSection renders the tools configuration section.
func (c *ConfigScreen) renderToolsSection(model *state.AppState) string {
	focused := model.FocusIndex == int(ConfigSectionTools)
	return c.renderSection(
		"[3] 工具配置",
		c.renderToolsContent(model),
		focused,
	)
}

// renderToolsContent renders the tools content.
func (c *ConfigScreen) renderToolsContent(model *state.AppState) string {
	if len(model.Config.Tools) == 0 {
		return c.layout.MutedText.Render("   (未配置工具)")
	}

	var lines []string
	for _, tool := range model.Config.Tools {
		checkbox := c.renderCheckbox(tool.Enabled)
		desc := tool.Description
		if desc == "" {
			desc = tool.Name
		}
		line := fmt.Sprintf("   %s %s: %s", checkbox, tool.Name, desc)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// renderWorkflowSection renders the workflow defaults section.
func (c *ConfigScreen) renderWorkflowSection(model *state.AppState) string {
	focused := model.FocusIndex == int(ConfigSectionWorkflow)
	return c.renderSection(
		"[4] 工作流默认设置",
		c.renderWorkflowContent(model),
		focused,
	)
}

// renderWorkflowContent renders the workflow defaults content.
func (c *ConfigScreen) renderWorkflowContent(model *state.AppState) string {
	workflow := model.Config.WorkflowDefaults

	keyStyle := lipgloss.NewStyle().
		Foreground(c.theme.TextMuted.ToLipgloss()).
		PaddingLeft(2)

	valueStyle := lipgloss.NewStyle().
		Foreground(c.theme.Text.ToLipgloss())

	defaultType := workflow.DefaultType
	if defaultType == "" {
		defaultType = "sequential"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s %s", keyStyle.Render("默认类型:"), valueStyle.Render(defaultType)))
	if workflow.Timeout > 0 {
		lines = append(lines, fmt.Sprintf("%s %s", keyStyle.Render("超时:"), valueStyle.Render(fmt.Sprintf("%ds", workflow.Timeout))))
	}
	if workflow.MaxIter > 0 {
		lines = append(lines, fmt.Sprintf("%s %d", keyStyle.Render("最大迭代:"), workflow.MaxIter))
	}

	return strings.Join(lines, "\n")
}

// renderSection renders a section with a header.
func (c *ConfigScreen) renderSection(header, content string, focused bool) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(c.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2)

	if focused {
		headerStyle = headerStyle.Foreground(c.theme.Primary.ToLipgloss())
	}

	return c.layout.JoinVertical(
		headerStyle.Render(header),
		content,
		"", // Empty line for spacing
	)
}

// renderCheckbox renders a checkbox indicator.
func (c *ConfigScreen) renderCheckbox(checked bool) string {
	if checked {
		return c.theme.SuccessStyle("[\u2713]")
	}
	return c.layout.MutedText.Render("[ ]")
}
