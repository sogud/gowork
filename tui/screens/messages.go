package screens

import (
	"github.com/sogud/gowork/state"
)

// ScreenChangeMsg indicates a request to change screens.
type ScreenChangeMsg struct {
	Screen state.Screen
}

// QuitMsg indicates a request to quit the application.
type QuitMsg struct {
	Quit bool
}

// HelpMsg indicates a request to show help information.
type HelpMsg struct {
	Text string
}

// TickMsg is used for periodic updates like elapsed time.
type TickMsg struct{}

// SubmitTaskMsg is sent when a task should be submitted/started.
type SubmitTaskMsg struct {
	TaskDescription string
	WorkflowType    state.WorkflowType
	SelectedAgents  []string
	MaxIterations   int
}

// FocusFieldMsg is sent to change focus to a specific input field.
type FocusFieldMsg struct {
	Field int
}