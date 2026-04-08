package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/tui/styles"
)

// StatusBar is a status bar component that implements tea.Model.
type StatusBar struct {
	theme      styles.Theme
	width      int
	leftText   string
	rightText  string
	statusMsg  string
	statusType StatusType
}

// StatusType represents the type of status message.
type StatusType int

const (
	StatusNone StatusType = iota
	StatusInfo
	StatusSuccess
	StatusWarning
	StatusError
)

// NewStatusBar creates a new StatusBar component with the given theme.
func NewStatusBar(theme styles.Theme) *StatusBar {
	return &StatusBar{
		theme:     theme,
		width:     80,
		statusType: StatusNone,
	}
}

// WithWidth sets the width of the status bar.
func (s *StatusBar) WithWidth(width int) *StatusBar {
	s.width = width
	return s
}

// WithLeftText sets the left-aligned text.
func (s *StatusBar) WithLeftText(text string) *StatusBar {
	s.leftText = text
	return s
}

// WithRightText sets the right-aligned text.
func (s *StatusBar) WithRightText(text string) *StatusBar {
	s.rightText = text
	return s
}

// WithStatus sets a status message with type.
func (s *StatusBar) WithStatus(msg string, statusType StatusType) *StatusBar {
	s.statusMsg = msg
	s.statusType = statusType
	return s
}

// SetWidth sets the width dynamically.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetLeftText sets the left-aligned text.
func (s *StatusBar) SetLeftText(text string) {
	s.leftText = text
}

// SetRightText sets the right-aligned text.
func (s *StatusBar) SetRightText(text string) {
	s.rightText = text
}

// SetStatus sets a status message with type.
func (s *StatusBar) SetStatus(msg string, statusType StatusType) {
	s.statusMsg = msg
	s.statusType = statusType
}

// ClearStatus clears the current status message.
func (s *StatusBar) ClearStatus() {
	s.statusMsg = ""
	s.statusType = StatusNone
}

// Init initializes the status bar (tea.Model interface).
func (s *StatusBar) Init() tea.Cmd {
	return nil
}

// Update handles messages (tea.Model interface).
func (s *StatusBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

// View renders the status bar (tea.Model interface).
func (s *StatusBar) View() string {
	return s.Render()
}

// Render renders the status bar.
func (s *StatusBar) Render() string {
	// Base style for the status bar
	baseStyle := lipgloss.NewStyle().
		Background(s.theme.Surface.ToLipgloss()).
		Padding(0, 1)

	// Left section
	leftStyle := baseStyle.Copy().
		Foreground(s.theme.Text.ToLipgloss())

	left := leftStyle.Render(s.leftText)

	// Right section (help text/shortcuts)
	rightStyle := baseStyle.Copy().
		Foreground(s.theme.TextMuted.ToLipgloss())

	right := rightStyle.Render(s.rightText)

	// Calculate spacing
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	spacing := s.width - leftLen - rightLen
	if spacing < 0 {
		spacing = 0
	}

	// Build the bar
	bar := left + repeatChar(" ", spacing) + right

	// If there's a status message, render it differently
	if s.statusMsg != "" {
		var statusColor lipgloss.Color
		switch s.statusType {
		case StatusSuccess:
			statusColor = s.theme.Success.ToLipgloss()
		case StatusWarning:
			statusColor = s.theme.Warning.ToLipgloss()
		case StatusError:
			statusColor = s.theme.Error.ToLipgloss()
		default: // StatusInfo or StatusNone
			statusColor = s.theme.Primary.ToLipgloss()
		}

		statusStyle := lipgloss.NewStyle().
			Foreground(statusColor).
			Padding(0, 1)

		statusText := statusStyle.Render(s.statusMsg)

		// Replace center with status
		leftPart := leftStyle.Render(s.leftText)
		rightPart := rightStyle.Render(s.rightText)

		leftW := lipgloss.Width(leftPart)
		statusW := lipgloss.Width(statusText)
		rightW := lipgloss.Width(rightPart)

		totalW := leftW + statusW + rightW
		if totalW < s.width {
			spacing := s.width - totalW
			bar = leftPart + repeatChar(" ", spacing/2) + statusText + repeatChar(" ", spacing-(spacing/2)) + rightPart
		} else {
			bar = leftPart + statusText + rightPart
		}
	}

	return bar
}

// RenderWithBorder renders the status bar with a top border.
func (s *StatusBar) RenderWithBorder() string {
	borderStyle := lipgloss.NewStyle().
		Foreground(s.theme.Border.ToLipgloss())

	border := borderStyle.Render(repeatChar("─", s.width))

	return border + "\n" + s.Render()
}

// String returns the status bar as a string.
func (s *StatusBar) String() string {
	return s.Render()
}
