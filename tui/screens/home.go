// Package screens provides individual screen implementations for the TUI.
package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/state"
	"github.com/sogud/gowork/tui/styles"
)

// HomeScreen is the main landing screen of the TUI.
type HomeScreen struct {
	theme  styles.Theme
	layout styles.Layout
}

// NewHomeScreen creates a new HomeScreen instance.
func NewHomeScreen() *HomeScreen {
	theme := styles.DefaultTheme()
	layout := styles.NewLayout(theme)
	return &HomeScreen{
		theme:  theme,
		layout: layout,
	}
}

// Name returns the screen's identifier.
func (h *HomeScreen) Name() state.Screen {
	return state.ScreenHome
}

// Render renders the Home screen content.
func (h *HomeScreen) Render(model *state.AppState) string {
	var sections []string

	// Title section
	title := h.renderTitle()
	sections = append(sections, title)

	// Navigation menu
	menu := h.renderNavigationMenu()
	sections = append(sections, menu)

	// Divider
	divider := h.layout.Divider(46)
	sections = append(sections, divider)

	// Agent status overview
	agentStatus := h.renderAgentStatus(model)
	sections = append(sections, agentStatus)

	// Model info
	modelInfo := h.renderModelInfo(model)
	sections = append(sections, modelInfo)

	return h.layout.JoinVertical(sections...)
}

// Update handles messages and updates the model.
func (h *HomeScreen) Update(msg interface{}, model *state.AppState) interface{} {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyRunes:
			switch string(m.Runes) {
			case "1":
				return ScreenChangeMsg{Screen: state.ScreenTaskInput}
			case "2":
				return ScreenChangeMsg{Screen: state.ScreenMonitor}
			case "3":
				return ScreenChangeMsg{Screen: state.ScreenConfig}
			case "4":
				return ScreenChangeMsg{Screen: state.ScreenHistory}
			case "q", "Q":
				return QuitMsg{Quit: true}
			case "?":
				return HelpMsg{Text: h.HelpText()}
			}
		}
	}

	return nil
}

// HelpText returns the help text for this screen.
func (h *HomeScreen) HelpText() string {
	return "Press 1-4 to navigate / q to quit / ? for help"
}

// renderTitle renders the title section.
func (h *HomeScreen) renderTitle() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary.ToLipgloss()).
		Bold(true).
		Padding(1, 0).
		Align(lipgloss.Center)

	title := titleStyle.Render("gowork - 多智能体协作框架")
	return title
}

// renderNavigationMenu renders the navigation menu section.
func (h *HomeScreen) renderNavigationMenu() string {
	menuItemStyle := lipgloss.NewStyle().
		Foreground(h.theme.Text.ToLipgloss()).
		PaddingLeft(4)

	keyStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary.ToLipgloss()).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(h.theme.TextMuted.ToLipgloss())

	menuItems := []struct {
		key     string
		name    string
		desc    string
		screen  string
	}{
		{"1", "提交新任务", "快速开始", "Task Input"},
		{"2", "查看执行监控", "实时状态", "Monitor"},
		{"3", "配置管理", "模型/智能体/工具", "Config"},
		{"4", "历史记录", "已执行工作流", "History"},
	}

	var lines []string
	for _, item := range menuItems {
		key := keyStyle.Render(fmt.Sprintf("[%s]", item.key))
		name := h.theme.HeaderStyle(item.name)
		desc := descStyle.Render(item.desc)
		line := menuItemStyle.Render(fmt.Sprintf("%s %s        %s", key, name, desc))
		lines = append(lines, line)
	}

	return h.layout.JoinVertical(lines...)
}

// renderAgentStatus renders the agent status overview section.
func (h *HomeScreen) renderAgentStatus(model *state.AppState) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2).
		MarginBottom(1)

	header := headerStyle.Render("智能体状态概览")

	// Create status box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(h.theme.Border.ToLipgloss()).
		Padding(0, 1).
		MarginLeft(2)

	// Render agent status items
	agentLines := h.renderAgentStatusItems(model)

	if len(agentLines) == 0 {
		agentLines = []string{"No agents configured"}
	}

	boxContent := h.layout.JoinVertical(agentLines...)
	statusBox := boxStyle.Render(boxContent)

	return h.layout.JoinVertical(header, statusBox)
}

// renderAgentStatusItems renders the individual agent status items.
func (h *HomeScreen) renderAgentStatusItems(model *state.AppState) []string {
	var lines []string

	// Group agents in pairs for layout
	agents := model.Agents
	if len(agents) == 0 {
		return lines
	}

	// Create agent lookup map
	agentMap := make(map[string]state.AgentInfo)
	for _, agent := range agents {
		agentMap[agent.Name] = agent
	}

	// Build pairs: (researcher, analyst), (writer, reviewer), (coordinator, _)
	pairs := [][]string{
		{"researcher", "analyst"},
		{"writer", "reviewer"},
		{"coordinator"},
	}

	itemStyle := lipgloss.NewStyle().Padding(0, 1)

	for _, pair := range pairs {
		var items []string
		for _, agentName := range pair {
			agent, exists := agentMap[agentName]
			if exists {
				item := h.renderSingleAgentStatus(agentName, agent.Status)
				items = append(items, itemStyle.Render(item))
			} else {
				// Show placeholder if agent not found
				item := h.renderSingleAgentStatus(agentName, state.AgentWaiting)
				items = append(items, itemStyle.Render(item))
			}
		}
		line := lipgloss.JoinHorizontal(lipgloss.Top, items...)
		lines = append(lines, line)
	}

	return lines
}

// renderSingleAgentStatus renders a single agent's status.
func (h *HomeScreen) renderSingleAgentStatus(name string, status state.AgentStatus) string {
	nameStyle := lipgloss.NewStyle().
		Foreground(h.theme.Text.ToLipgloss())

	// Status indicator and text
	var statusColor styles.Color
	var statusText string
	var indicator string

	switch status {
	case state.AgentRunning:
		statusColor = h.theme.Success
		statusText = "运行中"
		indicator = "●"
	case state.AgentCompleted:
		statusColor = h.theme.Success
		statusText = "完成"
		indicator = "●"
	case state.AgentFailed:
		statusColor = h.theme.Error
		statusText = "失败"
		indicator = "●"
	case state.AgentWaiting:
		statusColor = h.theme.TextMuted
		statusText = "就绪"
		indicator = "●"
	default:
		statusColor = h.theme.Text
		statusText = status.String()
		indicator = "●"
	}

	indicatorStyle := lipgloss.NewStyle().
		Foreground(statusColor.ToLipgloss())

	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor.ToLipgloss())

	return fmt.Sprintf("%s %s %s", nameStyle.Render(name), indicatorStyle.Render(indicator), statusStyle.Render(statusText))
}

// renderModelInfo renders the current model information.
func (h *HomeScreen) renderModelInfo(model *state.AppState) string {
	modelProvider := model.Config.ModelProvider

	// Build model identifier string
	var modelStr string
	if modelProvider.Type != "" && modelProvider.ModelName != "" {
		modelStr = fmt.Sprintf("%s/%s", modelProvider.Type, modelProvider.ModelName)
	} else if modelProvider.ModelName != "" {
		modelStr = modelProvider.ModelName
	} else {
		modelStr = "未配置"
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(h.theme.TextMuted.ToLipgloss()).
		PaddingLeft(2)

	valueStyle := lipgloss.NewStyle().
		Foreground(h.theme.Text.ToLipgloss())

	label := labelStyle.Render("当前模型:")
	value := valueStyle.Render(modelStr)

	return fmt.Sprintf("%s %s", label, value)
}