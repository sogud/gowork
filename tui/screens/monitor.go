// Package screens provides TUI screen implementations.
package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/state"
	"github.com/sogud/gowork/tui/styles"
)

// MonitorScreen displays real-time workflow execution status.
type MonitorScreen struct {
	theme  styles.Theme
	layout styles.Layout
	// expanded indicates if the output stream is in expanded view mode
	expanded bool
}

// NewMonitorScreen creates a new MonitorScreen.
func NewMonitorScreen() *MonitorScreen {
	theme := styles.DefaultTheme()
	layout := styles.NewLayout(theme)
	return &MonitorScreen{
		theme:  theme,
		layout: layout,
	}
}

// Name returns the screen identifier.
func (m *MonitorScreen) Name() state.Screen {
	return state.ScreenMonitor
}

// Render renders the monitor screen content.
func (m *MonitorScreen) Render(model *state.AppState) string {
	if model == nil || model.ActiveWorkflow == nil {
		return m.renderNoWorkflow()
	}

	return m.renderWorkflow(model)
}

// renderNoWorkflow renders when there's no active workflow.
func (m *MonitorScreen) renderNoWorkflow() string {
	title := m.theme.TitleStyle("Execution Monitor")
	message := m.layout.CenterText("No active workflow running", 60)
	help := m.layout.HelpText.Render("Press Esc to return to Home")

	return m.layout.JoinVertical(title, message, help)
}

// renderWorkflow renders the active workflow state.
func (m *MonitorScreen) renderWorkflow(model *state.AppState) string {
	wf := model.ActiveWorkflow

	var sections []string

	// Header section
	sections = append(sections, m.renderHeader(wf))

	// Agent execution status section
	sections = append(sections, m.renderAgentStatus(wf, model.FocusIndex))

	// Divider
	sections = append(sections, m.layout.Divider(50))

	// Real-time output stream section
	sections = append(sections, m.renderOutputStream(wf, model.FocusIndex))

	// Join all sections
	return m.layout.JoinVertical(sections...)
}

// renderHeader renders the workflow header with type, task, elapsed time, and status.
func (m *MonitorScreen) renderHeader(wf *state.WorkflowState) string {
	// Title line: "Execution Monitor - <workflow_type> Workflow"
	title := m.theme.TitleStyle("Execution Monitor")
	workflowType := m.theme.HeaderStyle(fmt.Sprintf("%s Workflow", wf.Type))

	headerLine := title + " - " + workflowType

	// Task line: "Task: <task_description>"
	task := m.layout.TruncateText(wf.Task, 50)
	taskLine := m.layout.KeyValue("Task", task)

	// Elapsed time and status line
	elapsed := m.formatElapsedTime(wf)
	statusText := wf.Status.String()
	statusColor := m.theme.GetStatusColor(statusText)
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor.ToLipgloss()).
		Bold(true)
	status := statusStyle.Render(statusText)

	timeStatusLine := m.layout.KeyValue("Elapsed", elapsed) + "  " + m.layout.KeyValue("Status", status)

	return m.layout.JoinVertical(headerLine, taskLine, timeStatusLine)
}

// formatElapsedTime formats the elapsed time as MM:SS.
func (m *MonitorScreen) formatElapsedTime(wf *state.WorkflowState) string {
	elapsed := wf.ElapsedTime
	if elapsed == 0 && wf.Status == state.WorkflowRunning {
		// Calculate elapsed time if not set (running state)
		elapsed = time.Since(wf.StartTime)
	}

	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// renderAgentStatus renders the agent execution status with progress bars.
func (m *MonitorScreen) renderAgentStatus(wf *state.WorkflowState, focusIndex int) string {
	var lines []string

	// Section header
	header := m.theme.HeaderStyle("Agent Execution Status")
	lines = append(lines, header, "")

	for i, agent := range wf.AgentExecutions {
		line := m.renderAgentLine(agent, i, focusIndex, len(wf.AgentExecutions))
		lines = append(lines, line)

		// Show output preview and tool info for selected or running agent
		if i == focusIndex || agent.Status == state.AgentRunning {
			if agent.Output != "" || len(agent.ToolCalls) > 0 {
				details := m.renderAgentDetails(agent)
				lines = append(lines, details)
			}
		}
		lines = append(lines, "")
	}

	return m.layout.JoinVertical(lines...)
}

// renderAgentLine renders a single agent's status line.
func (m *MonitorScreen) renderAgentLine(agent state.AgentExecution, index, focusIndex, total int) string {
	// Tree structure prefix (├─ or └─)
	var prefix string
	if index == total-1 {
		prefix = "└─ "
	} else {
		prefix = "├─ "
	}

	// Focused indicator
	var focusIndicator string
	if index == focusIndex {
		focusIndicator = lipgloss.NewStyle().
			Foreground(m.theme.Primary.ToLipgloss()).
			Bold(true).
			Render("▶ ")
	} else {
		focusIndicator = "  "
	}

	// Agent name
	nameStyle := lipgloss.NewStyle().Foreground(m.theme.Text.ToLipgloss())
	if index == focusIndex {
		nameStyle = nameStyle.Bold(true)
	}
	name := nameStyle.Render(agent.Name)

	// Progress bar (calculate based on status)
	var progress float64
	switch agent.Status {
	case state.AgentWaiting:
		progress = 0
	case state.AgentRunning:
		progress = 0.5 // Running is shown as 50% by default
	case state.AgentCompleted:
		progress = 1.0
	case state.AgentFailed:
		progress = 1.0
	}

	progressBar := m.layout.ProgressBar(progress, 12)

	// Percentage
	percent := fmt.Sprintf("%3d%%", int(progress*100))

	// Status text
	statusText := agent.Status.String()
	statusColor := m.theme.GetStatusColor(statusText)
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor.ToLipgloss()).
		Bold(true)
	status := statusStyle.Render(statusText)

	return focusIndicator + prefix + name + "  " + progressBar + " " + percent + "  " + status
}

// renderAgentDetails renders agent output preview and tool call info.
func (m *MonitorScreen) renderAgentDetails(agent state.AgentExecution) string {
	var lines []string

	// Indentation for tree structure
	indent := "│   "

	// Output preview (truncated)
	if agent.Output != "" {
		outputPreview := m.layout.TruncateText(agent.Output, 40)
		outputLine := indent + m.layout.KeyValue("Output", outputPreview)
		lines = append(lines, outputLine)
	}

	// Tool calls summary
	if len(agent.ToolCalls) > 0 {
		// Count tool calls by name
		toolCounts := make(map[string]int)
		for _, tc := range agent.ToolCalls {
			toolCounts[tc.ToolName]++
		}

		var toolSummaries []string
		for toolName, count := range toolCounts {
			toolSummaries = append(toolSummaries, fmt.Sprintf("%s(%d)", toolName, count))
		}

		toolLine := indent + m.layout.KeyValue("Tools", strings.Join(toolSummaries, ", "))
		lines = append(lines, toolLine)
	}

	return m.layout.JoinVertical(lines...)
}

// renderOutputStream renders the real-time output stream for the selected agent.
func (m *MonitorScreen) renderOutputStream(wf *state.WorkflowState, focusIndex int) string {
	if focusIndex >= len(wf.AgentExecutions) {
		return ""
	}

	agent := wf.AgentExecutions[focusIndex]

	// Section header: "Real-time Output Stream (<agent_name>)"
	header := m.theme.HeaderStyle(fmt.Sprintf("Real-time Output Stream (%s)", agent.Name))

	// Output box
	output := agent.Output
	if output == "" {
		output = m.theme.MutedStyle("No output yet...")
	}

	// Determine box dimensions based on expanded mode
	width := 50
	height := 5
	if m.expanded {
		height = 10
	}

	outputBox := m.layout.BoxWithBorder("", width, height)

	// Wrap output to fit in box
	wrappedOutput := m.wrapText(output, width-4)

	return m.layout.JoinVertical(header, outputBox.Render(wrappedOutput))
}

// wrapText wraps text to fit within a given width.
func (m *MonitorScreen) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result = append(result, line)
		} else {
			// Split long lines
			for len(line) > width {
				result = append(result, line[:width])
				line = line[width:]
			}
			if line != "" {
				result = append(result, line)
			}
		}
	}

	return strings.Join(result, "\n")
}

// Update handles messages and updates the model.
func (m *MonitorScreen) Update(msg interface{}, model *state.AppState) interface{} {
	// Handle key messages
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.Type {
	case tea.KeyTab:
		// Cycle through agents
		if model.ActiveWorkflow != nil && len(model.ActiveWorkflow.AgentExecutions) > 0 {
			model.FocusIndex = (model.FocusIndex + 1) % len(model.ActiveWorkflow.AgentExecutions)
		}
		return nil

	case tea.KeyShiftTab:
		// Cycle backwards through agents
		if model.ActiveWorkflow != nil && len(model.ActiveWorkflow.AgentExecutions) > 0 {
			model.FocusIndex--
			if model.FocusIndex < 0 {
				model.FocusIndex = len(model.ActiveWorkflow.AgentExecutions) - 1
			}
		}
		return nil

	case tea.KeyEnter:
		// Toggle expanded output view
		m.expanded = !m.expanded
		return nil

	case tea.KeyEsc:
		// Return to Home screen
		return ScreenChangeMsg{Screen: state.ScreenHome}

	case tea.KeyRunes:
		// Check for specific key characters
		switch string(keyMsg.Runes) {
		case "q":
			return QuitMsg{Quit: true}
		case "?":
			return HelpMsg{Text: m.HelpText()}
		}
	}

	return nil
}

// HelpText returns the help text for this screen.
func (m *MonitorScreen) HelpText() string {
	return "Tab/Shift+Tab switch agents / Enter view details / Esc back / ? help / q quit"
}