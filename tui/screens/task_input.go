// Package screens provides individual screen implementations for the TUI.
package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/state"
	"github.com/sogud/gowork/tui/styles"
)

// InputField constants for cursor field tracking.
const (
	FieldTaskDescription = 0
	FieldWorkflowType    = 1
	FieldAgentSelection  = 2
	FieldMaxIterations   = 3
)

// Available workflow types for selection.
var workflowTypes = []state.WorkflowType{
	state.WorkflowSequential,
	state.WorkflowParallel,
	state.WorkflowLoop,
	state.WorkflowDynamic,
}

// TaskInputScreen is the screen for creating and submitting new tasks.
type TaskInputScreen struct {
	theme         styles.Theme
	layout        styles.Layout
	taskInput     textinput.Model
	iterationsInput textinput.Model
	initialized   bool
}

// NewTaskInputScreen creates a new TaskInputScreen instance.
func NewTaskInputScreen() *TaskInputScreen {
	theme := styles.DefaultTheme()
	layout := styles.NewLayout(theme)

	// Initialize task description input
	taskInput := textinput.New()
	taskInput.Placeholder = "输入任务描述..."
	taskInput.Focus()
	taskInput.CharLimit = 500
	taskInput.Width = 40

	// Initialize max iterations input
	iterationsInput := textinput.New()
	iterationsInput.Placeholder = "3"
	iterationsInput.CharLimit = 3
	iterationsInput.Width = 5

	return &TaskInputScreen{
		theme:           theme,
		layout:          layout,
		taskInput:       taskInput,
		iterationsInput: iterationsInput,
		initialized:     false,
	}
}

// Name returns the screen's identifier.
func (t *TaskInputScreen) Name() state.Screen {
	return state.ScreenTaskInput
}

// initFromState initializes the text inputs from the current state.
func (t *TaskInputScreen) initFromState(model *state.AppState) {
	if t.initialized {
		return
	}

	// Set task description from state
	if model.TaskInput.TaskDescription != "" {
		t.taskInput.SetValue(model.TaskInput.TaskDescription)
	}

	// Set max iterations from state
	t.iterationsInput.SetValue(fmt.Sprintf("%d", model.TaskInput.MaxIterations))

	// Set focus based on cursor field
	t.updateFocus(model.TaskInput.CursorField)

	t.initialized = true
}

// updateFocus updates which input field is focused.
func (t *TaskInputScreen) updateFocus(field int) {
	switch field {
	case FieldTaskDescription:
		t.taskInput.Focus()
		t.iterationsInput.Blur()
	case FieldMaxIterations:
		t.taskInput.Blur()
		t.iterationsInput.Focus()
	default:
		t.taskInput.Blur()
		t.iterationsInput.Blur()
	}
}

// Render renders the TaskInput screen content.
func (t *TaskInputScreen) Render(model *state.AppState) string {
	t.initFromState(model)

	var sections []string

	// Title
	title := t.renderTitle()
	sections = append(sections, title)

	// Task description section
	taskSection := t.renderTaskDescription(model)
	sections = append(sections, taskSection)

	// Workflow type section
	workflowSection := t.renderWorkflowType(model)
	sections = append(sections, workflowSection)

	// Agent selection section
	agentSection := t.renderAgentSelection(model)
	sections = append(sections, agentSection)

	// Max iterations section (for Loop/Dynamic modes)
	if model.TaskInput.WorkflowType == state.WorkflowLoop ||
		model.TaskInput.WorkflowType == state.WorkflowDynamic {
		iterSection := t.renderMaxIterations(model)
		sections = append(sections, iterSection)
	}

	return t.layout.JoinVertical(sections...)
}

// Update handles messages and updates the model.
func (t *TaskInputScreen) Update(msg interface{}, model *state.AppState) interface{} {
	t.initFromState(model)

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyEsc:
			// Return to home screen
			return ScreenChangeMsg{Screen: state.ScreenHome}

		case tea.KeyTab:
			// Cycle through input fields
			newField := (model.TaskInput.CursorField + 1) % 4
			model.TaskInput.CursorField = newField
			t.updateFocus(newField)
			return nil

		case tea.KeyShiftTab:
			// Cycle backwards through input fields
			newField := model.TaskInput.CursorField - 1
			if newField < 0 {
				newField = 3
			}
			model.TaskInput.CursorField = newField
			t.updateFocus(newField)
			return nil

		case tea.KeyEnter:
			// Submit task if at least one agent selected and task description provided
			if model.TaskInput.TaskDescription != "" && len(model.TaskInput.SelectedAgents) > 0 {
				return SubmitTaskMsg{
					TaskDescription: model.TaskInput.TaskDescription,
					WorkflowType:    model.TaskInput.WorkflowType,
					SelectedAgents:  model.TaskInput.SelectedAgents,
					MaxIterations:   model.TaskInput.MaxIterations,
				}
			}
			return nil

		case tea.KeyUp:
			// Navigate within current field
			switch model.TaskInput.CursorField {
			case FieldWorkflowType:
				// Select previous workflow type
				currentIdx := t.findWorkflowIndex(model.TaskInput.WorkflowType)
				if currentIdx > 0 {
					model.TaskInput.WorkflowType = workflowTypes[currentIdx-1]
				}
			case FieldAgentSelection:
				// Navigate agent selection up
				if model.FocusIndex > 0 {
					model.FocusIndex--
				}
			}
			return nil

		case tea.KeyDown:
			// Navigate within current field
			switch model.TaskInput.CursorField {
			case FieldWorkflowType:
				// Select next workflow type
				currentIdx := t.findWorkflowIndex(model.TaskInput.WorkflowType)
				if currentIdx < len(workflowTypes)-1 {
					model.TaskInput.WorkflowType = workflowTypes[currentIdx+1]
				}
			case FieldAgentSelection:
				// Navigate agent selection down
				agents := t.getAvailableAgents(model)
				if model.FocusIndex < len(agents)-1 {
					model.FocusIndex++
				}
			}
			return nil

		case tea.KeySpace:
			// Toggle selection based on current field
			switch model.TaskInput.CursorField {
			case FieldWorkflowType:
				// No action for workflow type - use arrows to change
			case FieldAgentSelection:
				// Toggle agent selection
				agents := t.getAvailableAgents(model)
				if model.FocusIndex < len(agents) {
					agentName := agents[model.FocusIndex]
					t.toggleAgentSelection(model, agentName)
				}
			}
			return nil

		case tea.KeyRunes:
			// Handle text input for focused field
			switch model.TaskInput.CursorField {
			case FieldTaskDescription:
				// Update task input
				ti, cmd := t.taskInput.Update(m)
				t.taskInput = ti
				model.TaskInput.TaskDescription = t.taskInput.Value()
				return cmd
			case FieldMaxIterations:
				// Update iterations input (only allow digits)
				if len(m.Runes) > 0 && m.Runes[0] >= '0' && m.Runes[0] <= '9' {
					ii, cmd := t.iterationsInput.Update(m)
					t.iterationsInput = ii
					// Parse and update max iterations
					if val := t.iterationsInput.Value(); val != "" {
						var iter int
						fmt.Sscanf(val, "%d", &iter)
						if iter > 0 {
							model.TaskInput.MaxIterations = iter
						}
					}
					return cmd
				}
			}

		case tea.KeyBackspace:
			// Handle backspace for text inputs
			switch model.TaskInput.CursorField {
			case FieldTaskDescription:
				ti, cmd := t.taskInput.Update(m)
				t.taskInput = ti
				model.TaskInput.TaskDescription = t.taskInput.Value()
				return cmd
			case FieldMaxIterations:
				ii, cmd := t.iterationsInput.Update(m)
				t.iterationsInput = ii
				if val := t.iterationsInput.Value(); val != "" {
					var iter int
					fmt.Sscanf(val, "%d", &iter)
					model.TaskInput.MaxIterations = iter
				} else {
					model.TaskInput.MaxIterations = 3 // default
				}
				return cmd
			}
		}
	}

	// Handle text input updates for focused fields
	if model.TaskInput.CursorField == FieldTaskDescription {
		ti, cmd := t.taskInput.Update(msg)
		t.taskInput = ti
		model.TaskInput.TaskDescription = t.taskInput.Value()
		return cmd
	}
	if model.TaskInput.CursorField == FieldMaxIterations {
		ii, cmd := t.iterationsInput.Update(msg)
		t.iterationsInput = ii
		if val := t.iterationsInput.Value(); val != "" {
			var iter int
			fmt.Sscanf(val, "%d", &iter)
			if iter > 0 {
				model.TaskInput.MaxIterations = iter
			}
		}
		return cmd
	}

	return nil
}

// HelpText returns the help text for this screen.
func (t *TaskInputScreen) HelpText() string {
	return "Tab 切换字段 / 空格选择 / Enter 提交 / Esc 返回"
}

// renderTitle renders the title section.
func (t *TaskInputScreen) renderTitle() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(t.theme.Primary.ToLipgloss()).
		Bold(true).
		Padding(1, 0).
		Align(lipgloss.Center).
		Width(50)

	return titleStyle.Render("提交新任务")
}

// renderTaskDescription renders the task description input section.
func (t *TaskInputScreen) renderTaskDescription(model *state.AppState) string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(t.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2)

	header := headerStyle.Render("任务描述")

	// Input box style (focused/unfocused)
	focused := model.TaskInput.CursorField == FieldTaskDescription
	boxStyle := t.getInputBoxStyle(focused, 44)

	// Render text input
	inputView := t.taskInput.View()
	inputBox := boxStyle.Render(inputView)

	return t.layout.JoinVertical(header, inputBox)
}

// renderWorkflowType renders the workflow type selection section.
func (t *TaskInputScreen) renderWorkflowType(model *state.AppState) string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(t.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2)

	header := headerStyle.Render("工作流类型")

	// Render radio buttons
	options := t.renderWorkflowOptions(model)

	return t.layout.JoinVertical(header, options)
}

// renderWorkflowOptions renders the workflow type radio button options.
func (t *TaskInputScreen) renderWorkflowOptions(model *state.AppState) string {
	focused := model.TaskInput.CursorField == FieldWorkflowType

	optionStyle := lipgloss.NewStyle().
		Padding(0, 2)

	var optionTexts []string
	for _, wt := range workflowTypes {
		selected := model.TaskInput.WorkflowType == wt
		optionText := t.renderRadioOption(t.workflowTypeLabel(wt), selected, focused && selected)
		optionTexts = append(optionTexts, optionStyle.Render(optionText))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, optionTexts...)
}

// workflowTypeLabel returns the display label for a workflow type.
func (t *TaskInputScreen) workflowTypeLabel(wt state.WorkflowType) string {
	switch wt {
	case state.WorkflowSequential:
		return "Sequential"
	case state.WorkflowParallel:
		return "Parallel"
	case state.WorkflowLoop:
		return "Loop"
	case state.WorkflowDynamic:
		return "Dynamic"
	default:
		return string(wt)
	}
}

// renderRadioOption renders a single radio button option.
func (t *TaskInputScreen) renderRadioOption(label string, selected, focused bool) string {
	var indicator string
	var labelStyle lipgloss.Style

	if selected {
		indicator = "●"
		if focused {
			labelStyle = lipgloss.NewStyle().
				Foreground(t.theme.Primary.ToLipgloss()).
				Bold(true)
		} else {
			labelStyle = lipgloss.NewStyle().
				Foreground(t.theme.Text.ToLipgloss())
		}
	} else {
		indicator = "○"
		if focused {
			labelStyle = lipgloss.NewStyle().
				Foreground(t.theme.TextMuted.ToLipgloss())
		} else {
			labelStyle = lipgloss.NewStyle().
				Foreground(t.theme.TextMuted.ToLipgloss())
		}
	}

	return fmt.Sprintf("%s %s", indicator, labelStyle.Render(label))
}

// renderAgentSelection renders the agent selection section.
func (t *TaskInputScreen) renderAgentSelection(model *state.AppState) string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(t.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2)

	header := headerStyle.Render("选择智能体 (空格选择)")

	// Agent selection box
	agentsContent := t.renderAgentList(model)

	focused := model.TaskInput.CursorField == FieldAgentSelection
	boxStyle := t.getInputBoxStyle(focused, 44)

	agentBox := boxStyle.Render(agentsContent)

	return t.layout.JoinVertical(header, agentBox)
}

// renderAgentList renders the list of selectable agents.
func (t *TaskInputScreen) renderAgentList(model *state.AppState) string {
	agents := t.getAvailableAgents(model)
	if len(agents) == 0 {
		return t.theme.MutedStyle("没有可用的智能体")
	}

	// Group agents in pairs for layout
	var lines []string
	for i := 0; i < len(agents); i += 2 {
		var items []string
		for j := 0; j < 2 && i+j < len(agents); j++ {
			agentName := agents[i+j]
			selected := t.isAgentSelected(model, agentName)
			focused := model.TaskInput.CursorField == FieldAgentSelection &&
				model.FocusIndex == i+j
			item := t.renderCheckboxOption(agentName, selected, focused)
			items = append(items, lipgloss.NewStyle().Padding(0, 1).Render(item))
		}
		line := lipgloss.JoinHorizontal(lipgloss.Top, items...)
		lines = append(lines, line)
	}

	return t.layout.JoinVertical(lines...)
}

// renderCheckboxOption renders a single checkbox option.
func (t *TaskInputScreen) renderCheckboxOption(label string, selected, focused bool) string {
	var indicator string
	var labelStyle lipgloss.Style

	if selected {
		indicator = "[✓]"
		labelStyle = lipgloss.NewStyle().
			Foreground(t.theme.Text.ToLipgloss())
	} else {
		indicator = "[ ]"
		labelStyle = lipgloss.NewStyle().
			Foreground(t.theme.TextMuted.ToLipgloss())
	}

	if focused {
		// Highlight the focused option
		indicatorStyle := lipgloss.NewStyle().
			Foreground(t.theme.Primary.ToLipgloss()).
			Bold(true)
		indicator = indicatorStyle.Render(indicator)
		labelStyle = lipgloss.NewStyle().
			Foreground(t.theme.Primary.ToLipgloss())
	}

	return fmt.Sprintf("%s %s", indicator, labelStyle.Render(label))
}

// renderMaxIterations renders the max iterations input section.
func (t *TaskInputScreen) renderMaxIterations(model *state.AppState) string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(t.theme.Secondary.ToLipgloss()).
		Bold(true).
		PaddingLeft(2)

	workflowLabel := t.workflowTypeLabel(model.TaskInput.WorkflowType)
	header := headerStyle.Render(fmt.Sprintf("最大迭代次数 (%s模式)", workflowLabel))

	// Input box
	focused := model.TaskInput.CursorField == FieldMaxIterations
	boxStyle := t.getInputBoxStyle(focused, 10)

	inputView := t.iterationsInput.View()
	inputBox := boxStyle.Render(inputView)

	return lipgloss.JoinHorizontal(lipgloss.Top, header, " ", inputBox)
}

// getInputBoxStyle returns the appropriate box style based on focus state.
func (t *TaskInputScreen) getInputBoxStyle(focused bool, width int) lipgloss.Style {
	border := lipgloss.RoundedBorder()

	if focused {
		return lipgloss.NewStyle().
			Border(border).
			BorderForeground(t.theme.Primary.ToLipgloss()).
			Padding(0, 1).
			Width(width)
	}

	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(t.theme.Border.ToLipgloss()).
		Padding(0, 1).
		Width(width)
}

// getAvailableAgents returns the list of available agents from the model.
func (t *TaskInputScreen) getAvailableAgents(model *state.AppState) []string {
	var agents []string
	for _, agent := range model.Agents {
		agents = append(agents, agent.Name)
	}
	return agents
}

// isAgentSelected checks if an agent is in the selected agents list.
func (t *TaskInputScreen) isAgentSelected(model *state.AppState, agentName string) bool {
	for _, selected := range model.TaskInput.SelectedAgents {
		if selected == agentName {
			return true
		}
	}
	return false
}

// toggleAgentSelection toggles the selection state of an agent.
func (t *TaskInputScreen) toggleAgentSelection(model *state.AppState, agentName string) {
	// Check if already selected
	for i, selected := range model.TaskInput.SelectedAgents {
		if selected == agentName {
			// Remove from selection
			model.TaskInput.SelectedAgents = append(
				model.TaskInput.SelectedAgents[:i],
				model.TaskInput.SelectedAgents[i+1:]...,
			)
			return
		}
	}

	// Add to selection
	model.TaskInput.SelectedAgents = append(
		model.TaskInput.SelectedAgents,
		agentName,
	)
}

// findWorkflowIndex finds the index of the current workflow type.
func (t *TaskInputScreen) findWorkflowIndex(wt state.WorkflowType) int {
	for i, w := range workflowTypes {
		if w == wt {
			return i
		}
	}
	return 0
}