package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/tui/styles"
)

// ProgressState represents the state of a progress indicator.
type ProgressState int

const (
	ProgressWaiting ProgressState = iota
	ProgressRunning
	ProgressCompleted
	ProgressFailed
)

// Progress is a progress bar component that implements tea.Model.
type Progress struct {
	theme    styles.Theme
	percent  float64
	width    int
	height   int
	state    ProgressState
	status   ProgressState  // Alias for state (test compatibility)
	label    string
	showPct  bool
	animated bool
	animate  bool           // Alias for animated (test compatibility)
}

// NewProgress creates a new Progress component with the given theme.
func NewProgress(theme styles.Theme) *Progress {
	return &Progress{
		theme:    theme,
		percent:  0,
		width:    40,
		state:    ProgressWaiting,
		showPct:  true,
		animated: false,
	}
}

// WithPercent sets the progress percentage (0.0 to 1.0).
func (p *Progress) WithPercent(percent float64) *Progress {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	p.percent = percent
	return p
}

// WithWidth sets the width of the progress bar.
func (p *Progress) WithWidth(width int) *Progress {
	p.width = width
	return p
}

// WithState sets the progress state.
func (p *Progress) WithState(state ProgressState) *Progress {
	p.state = state
	return p
}

// WithLabel sets a label for the progress bar.
func (p *Progress) WithLabel(label string) *Progress {
	p.label = label
	return p
}

// SetLabel sets the label.
func (p *Progress) SetLabel(label string) {
	p.label = label
}

// WithStatus sets the progress state (alias for WithState).
func (p *Progress) WithStatus(state ProgressState) *Progress {
	p.state = state
	p.status = state
	return p
}

// SetStatus sets the progress state (alias for SetState).
func (p *Progress) SetStatus(state ProgressState) {
	p.state = state
	p.status = state
}

// GetStatus returns the current progress state (alias for GetState).
func (p *Progress) GetStatus() ProgressState {
	return p.state
}

// WithHeight sets the height.
func (p *Progress) WithHeight(height int) *Progress {
	p.height = height
	return p
}

// WithShowPercent sets whether to show percentage text.
func (p *Progress) WithShowPercent(show bool) *Progress {
	p.showPct = show
	return p
}

// WithShowPercentage sets whether to show percentage text (alias for WithShowPercent).
func (p *Progress) WithShowPercentage(show bool) *Progress {
	p.showPct = show
	return p
}

// WithAnimated sets whether the progress bar is animated.
func (p *Progress) WithAnimated(animated bool) *Progress {
	p.animated = animated
	p.animate = animated
	return p
}

// WithAnimate sets whether the progress bar is animated.
func (p *Progress) WithAnimate(animate bool) *Progress {
	p.animated = animate
	p.animate = animate
	return p
}

// ViewCompact renders a compact progress view.
func (p *Progress) ViewCompact() string {
	return p.RenderCompact()
}

// Increment adds to the current progress percentage.
func (p *Progress) Increment(amount float64) {
	p.percent += amount
	if p.percent > 1 {
		p.percent = 1
	}
}

// Complete marks the progress as complete.
func (p *Progress) Complete() {
	p.percent = 1.0
	p.state = ProgressCompleted
	p.status = ProgressCompleted
}

// Fail marks the progress as failed.
func (p *Progress) Fail() {
	p.state = ProgressFailed
	p.status = ProgressFailed
}

// SetPercent sets the progress percentage.
func (p *Progress) SetPercent(percent float64) {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	p.percent = percent
}

// GetPercent returns the current progress percentage.
func (p *Progress) GetPercent() float64 {
	return p.percent
}

// SetState sets the progress state.
func (p *Progress) SetState(state ProgressState) {
	p.state = state
}

// GetState returns the current progress state.
func (p *Progress) GetState() ProgressState {
	return p.state
}

// SetWidth sets the width dynamically.
func (p *Progress) SetWidth(width int) {
	p.width = width
}

// Init initializes the progress (tea.Model interface).
func (p *Progress) Init() tea.Cmd {
	return nil
}

// Update handles messages (tea.Model interface).
func (p *Progress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}

// View renders the progress bar (tea.Model interface).
func (p *Progress) View() string {
	return p.Render()
}

// Render renders the progress bar.
func (p *Progress) Render() string {
	// Determine colors based on state
	var filledColor, emptyColor lipgloss.Color
	switch p.state {
	case ProgressRunning:
		filledColor = p.theme.Primary.ToLipgloss()
		emptyColor = p.theme.Border.ToLipgloss()
	case ProgressCompleted:
		filledColor = p.theme.Success.ToLipgloss()
		emptyColor = p.theme.Border.ToLipgloss()
	case ProgressFailed:
		filledColor = p.theme.Error.ToLipgloss()
		emptyColor = p.theme.Border.ToLipgloss()
	default: // ProgressWaiting
		filledColor = p.theme.TextMuted.ToLipgloss()
		emptyColor = p.theme.Border.ToLipgloss()
	}

	// Calculate filled width
	filledWidth := int(float64(p.width) * p.percent)
	if filledWidth < 0 {
		filledWidth = 0
	}
	if filledWidth > p.width {
		filledWidth = p.width
	}
	emptyWidth := p.width - filledWidth

	// Build progress bar
	filledStyle := lipgloss.NewStyle().
		Foreground(p.theme.Background.ToLipgloss()).
		Background(filledColor)

	emptyStyle := lipgloss.NewStyle().
		Foreground(p.theme.TextMuted.ToLipgloss()).
		Background(emptyColor)

	filled := filledStyle.Render(repeatChar("█", filledWidth))
	empty := emptyStyle.Render(repeatChar("░", emptyWidth))

	bar := filled + empty

	// Add percentage if enabled
	if p.showPct {
		pctStyle := lipgloss.NewStyle().
			Foreground(p.theme.Text.ToLipgloss()).
			PaddingLeft(1)
		pct := pctStyle.Render(fmt.Sprintf("%3.0f%%", p.percent*100))
		bar = bar + pct
	}

	// Add label if set
	if p.label != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(p.theme.Text.ToLipgloss()).
			PaddingRight(1)
		return labelStyle.Render(p.label) + bar
	}

	return bar
}

// RenderWithStatus renders the progress bar with a status indicator.
func (p *Progress) RenderWithStatus() string {
	var statusIcon string
	switch p.state {
	case ProgressRunning:
		statusIcon = "●"
	case ProgressCompleted:
		statusIcon = "✓"
	case ProgressFailed:
		statusIcon = "✗"
	default:
		statusIcon = "○"
	}

	var statusColor lipgloss.Color
	switch p.state {
	case ProgressRunning:
		statusColor = p.theme.Primary.ToLipgloss()
	case ProgressCompleted:
		statusColor = p.theme.Success.ToLipgloss()
	case ProgressFailed:
		statusColor = p.theme.Error.ToLipgloss()
	default:
		statusColor = p.theme.TextMuted.ToLipgloss()
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		PaddingRight(1)

	return statusStyle.Render(statusIcon) + p.Render()
}

// RenderCompact renders a compact progress indicator.
func (p *Progress) RenderCompact() string {
	var indicator string
	switch p.state {
	case ProgressRunning:
		indicator = "▓▓▓"
	case ProgressCompleted:
		indicator = "✓✓✓"
	case ProgressFailed:
		indicator = "✗✗✗"
	default:
		indicator = "░░░"
	}

	var color lipgloss.Color
	switch p.state {
	case ProgressRunning:
		color = p.theme.Primary.ToLipgloss()
	case ProgressCompleted:
		color = p.theme.Success.ToLipgloss()
	case ProgressFailed:
		color = p.theme.Error.ToLipgloss()
	default:
		color = p.theme.TextMuted.ToLipgloss()
	}

	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(indicator)
}

// String returns the progress as a string.
func (p *Progress) String() string {
	return p.Render()
}

// IsComplete returns true if progress is 100%.
func (p *Progress) IsComplete() bool {
	return p.percent >= 1.0
}

// IsFailed returns true if progress is in failed state.
func (p *Progress) IsFailed() bool {
	return p.state == ProgressFailed
}

// Reset resets the progress to initial state.
func (p *Progress) Reset() {
	p.percent = 0
	p.state = ProgressWaiting
	p.label = ""
}

// Helper function to repeat a character.
func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
