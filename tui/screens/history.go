// Package screens provides individual screen implementations for the TUI.
package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/state"
	"github.com/sogud/gowork/tui/components"
	"github.com/sogud/gowork/tui/styles"
)

// HistoryScreen shows past workflow executions.
type HistoryScreen struct {
	theme     styles.Theme
	layout    styles.Layout
	border    components.Border
	store     *state.HistoryStore
	searching bool // Whether search input is focused
}

// NewHistoryScreen creates a new HistoryScreen instance.
func NewHistoryScreen(historyStore *state.HistoryStore) *HistoryScreen {
	theme := styles.DefaultTheme()
	layout := styles.NewLayout(theme)
	border := components.NewBorder(theme)
	return &HistoryScreen{
		theme:     theme,
		layout:    layout,
		border:    border,
		store:     historyStore,
		searching: false,
	}
}

// Name returns the screen's identifier.
func (h *HistoryScreen) Name() state.Screen {
	return state.ScreenHistory
}

// Render renders the History screen content.
func (h *HistoryScreen) Render(model *state.AppState) string {
	// If viewing detail, render detail view instead
	if model.History.ViewDetail && model.History.DetailRecord != nil {
		return h.renderDetailView(model)
	}

	var sections []string

	// Title section
	title := h.renderTitle()
	sections = append(sections, title)

	// Search input
	search := h.renderSearchInput(model)
	sections = append(sections, search)

	// History table
	table := h.renderHistoryTable(model)
	sections = append(sections, table)

	// Statistics summary
	stats := h.renderStatistics(model)
	sections = append(sections, stats)

	return h.layout.JoinVertical(sections...)
}

// Update handles messages and updates the model.
func (h *HistoryScreen) Update(msg interface{}, model *state.AppState) interface{} {
	switch m := msg.(type) {
	case tea.KeyMsg:
		// Handle detail view mode
		if model.History.ViewDetail {
			return h.handleDetailViewKeys(m, model)
		}

		// Handle search mode
		if h.searching {
			return h.handleSearchKeys(m, model)
		}

		// Handle normal mode
		return h.handleNormalKeys(m, model)
	}

	return nil
}

// HelpText returns the help text for this screen.
func (h *HistoryScreen) HelpText() string {
	if h.searching {
		return "Enter to confirm / Esc to cancel search"
	}
	return "Press arrows to select / Enter for details / s search / d delete / r re-run / Esc back"
}

// handleNormalKeys handles keyboard input in normal mode.
func (h *HistoryScreen) handleNormalKeys(msg tea.KeyMsg, model *state.AppState) interface{} {
	switch msg.Type {
	case tea.KeyUp:
		// Move selection up
		if len(model.History.Records) > 0 && model.History.SelectedIndex > 0 {
			model.History.SelectedIndex--
		}
		return nil

	case tea.KeyDown:
		// Move selection down
		if len(model.History.Records) > 0 && model.History.SelectedIndex < len(model.History.Records)-1 {
			model.History.SelectedIndex++
		}
		return nil

	case tea.KeyEnter:
		// View detail of selected record
		if len(model.History.Records) > 0 && model.History.SelectedIndex < len(model.History.Records) {
			record := model.History.Records[model.History.SelectedIndex]
			model.History.ViewDetail = true
			model.History.DetailRecord = &record
		}
		return nil

	case tea.KeyEsc:
		// Return to Home screen
		return ScreenChangeMsg{Screen: state.ScreenHome}

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "s", "S":
			// Enter search mode
			h.searching = true
			return nil
		case "d", "D":
			// Delete selected record
			if len(model.History.Records) > 0 && model.History.SelectedIndex < len(model.History.Records) {
				record := model.History.Records[model.History.SelectedIndex]
				if h.store != nil {
					_ = h.store.DeleteRecord(record.ID) // Ignore error for now
				}
				// Refresh records after deletion
				records, _ := h.store.ListRecordsWithFilter(model.History.SearchQuery)
				model.History.Records = records
				// Adjust selection if needed
				if model.History.SelectedIndex >= len(model.History.Records) {
					model.History.SelectedIndex = len(model.History.Records) - 1
					if model.History.SelectedIndex < 0 {
						model.History.SelectedIndex = 0
					}
				}
			}
			return nil
		case "r", "R":
			// Re-run workflow - return a message for app to handle
			if len(model.History.Records) > 0 && model.History.SelectedIndex < len(model.History.Records) {
				record := model.History.Records[model.History.SelectedIndex]
				return RerunWorkflowMsg{
					Task:        record.Task,
					WorkflowType: record.Type,
				}
			}
			return nil
		case "q", "Q":
			return QuitMsg{Quit: true}
		case "?":
			return HelpMsg{Text: h.HelpText()}
		}
	}

	return nil
}

// handleSearchKeys handles keyboard input in search mode.
func (h *HistoryScreen) handleSearchKeys(msg tea.KeyMsg, model *state.AppState) interface{} {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel search mode
		h.searching = false
		return nil

	case tea.KeyEnter:
		// Confirm search
		h.searching = false
		// Apply search filter
		if h.store != nil {
			records, _ := h.store.ListRecordsWithFilter(model.History.SearchQuery)
			model.History.Records = records
			model.History.SelectedIndex = 0
		}
		return nil

	case tea.KeyBackspace:
		// Remove last character from search query
		if len(model.History.SearchQuery) > 0 {
			model.History.SearchQuery = model.History.SearchQuery[:len(model.History.SearchQuery)-1]
			// Live search - update records immediately
			if h.store != nil {
				records, _ := h.store.ListRecordsWithFilter(model.History.SearchQuery)
				model.History.Records = records
				model.History.SelectedIndex = 0
			}
		}
		return nil

	case tea.KeyRunes:
		// Add character to search query
		model.History.SearchQuery += string(msg.Runes)
		// Live search - update records immediately
		if h.store != nil {
			records, _ := h.store.ListRecordsWithFilter(model.History.SearchQuery)
			model.History.Records = records
			model.History.SelectedIndex = 0
		}
		return nil
	}

	return nil
}

// handleDetailViewKeys handles keyboard input in detail view mode.
func (h *HistoryScreen) handleDetailViewKeys(msg tea.KeyMsg, model *state.AppState) interface{} {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		// Return to history list
		model.History.ViewDetail = false
		model.History.DetailRecord = nil
		return nil
	}

	return nil
}

// renderTitle renders the title section.
func (h *HistoryScreen) renderTitle() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary.ToLipgloss()).
		Bold(true).
		Padding(1, 0).
		Align(lipgloss.Center)

	title := titleStyle.Render("历史记录")
	return title
}

// renderSearchInput renders the search input field.
func (h *HistoryScreen) renderSearchInput(model *state.AppState) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(h.theme.TextMuted.ToLipgloss()).
		PaddingLeft(2)

	label := labelStyle.Render("搜索:")

	// Input field style
	inputStyle := lipgloss.NewStyle().
		Foreground(h.theme.Text.ToLipgloss())
	if h.searching {
		inputStyle = inputStyle.
			Border(lipgloss.NormalBorder()).
			BorderForeground(h.theme.Primary.ToLipgloss())
	}

	// Render input field with query and cursor
	query := model.History.SearchQuery
	if h.searching {
		query = query + "|" // Cursor indicator
	}

	// Create input box
	inputBox := lipgloss.NewStyle().
		Width(30).
		Render(query)

	inputLine := lipgloss.JoinHorizontal(lipgloss.Top, label, " ", inputStyle.Render(inputBox))
	return inputLine
}

// renderHistoryTable renders the history records table.
func (h *HistoryScreen) renderHistoryTable(model *state.AppState) string {
	// Table container
	tableStyle := h.border.Rounded().Padding(0, 1).MarginLeft(2)
	tableWidth := 50

	// Header row
	headerStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary.ToLipgloss()).
		Bold(true).
		Padding(0, 1)

	header := fmt.Sprintf("%-4s %-12s %-10s %-8s %-14s", "#", "时间", "类型", "状态", "概要")
	headerLine := headerStyle.Render(header)

	// Divider line
	divider := h.border.DividerLine(tableWidth)

	// Build table content
	var rows []string
	rows = append(rows, headerLine)
	rows = append(rows, divider)

	// Render each record
	for i, record := range model.History.Records {
		row := h.renderTableRow(i, record, model.History.SelectedIndex)
		rows = append(rows, row)
	}

	// If no records, show empty message
	if len(model.History.Records) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(h.theme.TextMuted.ToLipgloss()).
			Padding(1, 2)
		rows = append(rows, emptyStyle.Render("暂无历史记录"))
	}

	content := h.layout.JoinVertical(rows...)
	return tableStyle.Width(tableWidth).Render(content)
}

// renderTableRow renders a single table row.
func (h *HistoryScreen) renderTableRow(index int, record state.WorkflowRecord, selectedIndex int) string {
	// Row style based on selection
	rowStyle := lipgloss.NewStyle().Padding(0, 1)
	if index == selectedIndex {
		rowStyle = lipgloss.NewStyle().
			Foreground(h.theme.Primary.ToLipgloss()).
			Background(h.theme.Surface.ToLipgloss()).
			Padding(0, 1)
	}

	// Format time
	timeStr := h.formatTime(record.StartTime)

	// Format type
	typeStr := h.formatWorkflowType(record.Type)

	// Format status
	statusStr := h.formatStatus(record.Status)

	// Truncate task summary
	summary := h.layout.TruncateText(record.Task, 14)

	// Build row
	row := fmt.Sprintf("%-4d %-12s %-10s %-8s %-14s",
		index+1, timeStr, typeStr, statusStr, summary)

	return rowStyle.Render(row)
}

// formatTime formats a timestamp for display.
func (h *HistoryScreen) formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	// If within 24 hours, show time
	if diff < 24*time.Hour {
		return t.Format("15:04:05")
	}

	// If within 7 days, show relative day
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	}

	// Otherwise show date
	return t.Format("01-02")
}

// formatWorkflowType formats a workflow type for display.
func (h *HistoryScreen) formatWorkflowType(t state.WorkflowType) string {
	switch t {
	case state.WorkflowSequential:
		return "Sequential"
	case state.WorkflowParallel:
		return "Parallel"
	case state.WorkflowLoop:
		return "Loop"
	case state.WorkflowDynamic:
		return "Dynamic"
	default:
		return string(t)
	}
}

// formatStatus formats a workflow status for display.
func (h *HistoryScreen) formatStatus(s state.WorkflowStatus) string {
	switch s {
	case state.WorkflowCompleted:
		return "完成"
	case state.WorkflowFailed:
		return "失败"
	case state.WorkflowRunning:
		return "运行"
	case state.WorkflowPending:
		return "等待"
	default:
		return s.String()
	}
}

// renderStatistics renders the statistics summary.
func (h *HistoryScreen) renderStatistics(model *state.AppState) string {
	// Calculate statistics
	total := len(model.History.Records)
	success := 0
	failed := 0

	for _, record := range model.History.Records {
		switch record.Status {
		case state.WorkflowCompleted:
			success++
		case state.WorkflowFailed:
			failed++
		}
	}

	// Build statistics line
	labelStyle := lipgloss.NewStyle().
		Foreground(h.theme.TextMuted.ToLipgloss()).
		PaddingLeft(2)

	valueStyle := lipgloss.NewStyle().
		Foreground(h.theme.Text.ToLipgloss())

	successStyle := lipgloss.NewStyle().
		Foreground(h.theme.Success.ToLipgloss())

	failedStyle := lipgloss.NewStyle().
		Foreground(h.theme.Error.ToLipgloss())

	stats := fmt.Sprintf("%s %s %d %s  %s %s  %s %s",
		labelStyle.Render("统计:"),
		valueStyle.Render("总计"),
		total,
		valueStyle.Render("个工作流"),
		successStyle.Render("成功"),
		successStyle.Render(fmt.Sprintf("%d", success)),
		failedStyle.Render("失败"),
		failedStyle.Render(fmt.Sprintf("%d", failed)),
	)

	return stats
}

// renderDetailView renders the detail view for a selected record.
func (h *HistoryScreen) renderDetailView(model *state.AppState) string {
	record := model.History.DetailRecord
	if record == nil {
		return h.layout.ErrorText.Render("No record selected")
	}

	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary.ToLipgloss()).
		Bold(true).
		Padding(1, 0)

	// Truncate ID for display (max 8 chars)
	idDisplay := record.ID
	if len(idDisplay) > 8 {
		idDisplay = idDisplay[:8]
	}

	title := titleStyle.Render(fmt.Sprintf("工作流详情 - %s", idDisplay))
	sections = append(sections, title)

	// Basic info box
	infoBox := h.border.Rounded().Padding(0, 1).MarginLeft(2).Width(50)

	var infoLines []string

	// Task
	infoLines = append(infoLines, h.layout.KeyValue("任务", record.Task))

	// Type
	infoLines = append(infoLines, h.layout.KeyValue("类型", h.formatWorkflowType(record.Type)))

	// Status
	statusDisplay := h.formatStatus(record.Status)
	infoLines = append(infoLines, h.layout.KeyValue("状态", statusDisplay))

	// Duration
	durationStr := h.formatDuration(record.Duration)
	infoLines = append(infoLines, h.layout.KeyValue("耗时", durationStr))

	// Start time
	infoLines = append(infoLines, h.layout.KeyValue("开始时间", record.StartTime.Format("2006-01-02 15:04:05")))

	// End time
	infoLines = append(infoLines, h.layout.KeyValue("结束时间", record.EndTime.Format("2006-01-02 15:04:05")))

	infoContent := h.layout.JoinVertical(infoLines...)
	sections = append(sections, infoBox.Render(infoContent))

	// Agent executions
	if len(record.Agents) > 0 {
		agentSection := h.renderAgentExecutions(record.Agents)
		sections = append(sections, agentSection)
	}

	// Final output
	if record.FinalOutput != "" {
		outputSection := h.renderFinalOutput(record.FinalOutput)
		sections = append(sections, outputSection)
	}

	// Error (if failed)
	if record.Status == state.WorkflowFailed && record.Error != "" {
		errorSection := h.renderError(record.Error)
		sections = append(sections, errorSection)
	}

	// Help text for detail view
	helpStyle := lipgloss.NewStyle().
		Foreground(h.theme.TextMuted.ToLipgloss()).
		Padding(1, 2)

	help := helpStyle.Render("按 Esc 或 Enter 返回历史列表")
	sections = append(sections, help)

	return h.layout.JoinVertical(sections...)
}

// renderAgentExecutions renders the agent execution details.
func (h *HistoryScreen) renderAgentExecutions(agents []state.AgentRecord) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2).
		MarginTop(1)

	header := headerStyle.Render("智能体执行详情")

	box := h.border.Rounded().Padding(0, 1).MarginLeft(2).Width(50)

	var lines []string
	for _, agent := range agents {
		agentLine := h.renderAgentLine(agent)
		lines = append(lines, agentLine)
	}

	content := h.layout.JoinVertical(lines...)
	return h.layout.JoinVertical(header, box.Render(content))
}

// renderAgentLine renders a single agent execution line.
func (h *HistoryScreen) renderAgentLine(agent state.AgentRecord) string {
	nameStyle := lipgloss.NewStyle().
		Foreground(h.theme.Primary.ToLipgloss()).
		Bold(true)

	durationStyle := lipgloss.NewStyle().
		Foreground(h.theme.TextMuted.ToLipgloss())

	tokensStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary.ToLipgloss())

	// Truncate output for display
	outputPreview := h.layout.TruncateText(strings.TrimSpace(agent.Output), 30)
	if outputPreview == "" {
		outputPreview = "(无输出)"
	}

	return fmt.Sprintf("%s %s %s - %s",
		nameStyle.Render(agent.Name),
		durationStyle.Render(h.formatDuration(agent.Duration)),
		tokensStyle.Render(fmt.Sprintf("%d tokens", agent.Tokens)),
		outputPreview)
}

// renderFinalOutput renders the final output section.
func (h *HistoryScreen) renderFinalOutput(output string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(h.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2).
		MarginTop(1)

	header := headerStyle.Render("最终输出")

	box := h.border.Success().Padding(0, 1).MarginLeft(2).Width(50)

	// Truncate output if too long
	displayOutput := output
	if len(output) > 500 {
		displayOutput = output[:500] + "..."
	}

	return h.layout.JoinVertical(header, box.Render(displayOutput))
}

// renderError renders the error section.
func (h *HistoryScreen) renderError(err string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(h.theme.Error.ToLipgloss()).
		Bold(true).
		PaddingLeft(2).
		MarginTop(1)

	header := headerStyle.Render("错误信息")

	box := h.border.Error().Padding(0, 1).MarginLeft(2).Width(50)

	return h.layout.JoinVertical(header, box.Render(err))
}

// formatDuration formats a duration for display.
func (h *HistoryScreen) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// RerunWorkflowMsg is a message to re-run a workflow with the same configuration.
type RerunWorkflowMsg struct {
	Task         string
	WorkflowType state.WorkflowType
}

// SetSearching sets the search mode state.
func (h *HistoryScreen) SetSearching(searching bool) {
	h.searching = searching
}

// IsSearching returns whether the screen is in search mode.
func (h *HistoryScreen) IsSearching() bool {
	return h.searching
}