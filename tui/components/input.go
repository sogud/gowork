package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/tui/styles"
)

// Input is a text input component that implements tea.Model.
type Input struct {
	theme       styles.Theme
	value       string
	placeholder string
	prompt      string
	cursor      int
	width       int
	focused     bool
	maxLength   int
}

// NewInput creates a new Input component with the given theme.
func NewInput(theme styles.Theme) *Input {
	return &Input{
		theme:     theme,
		value:     "",
		cursor:    0,
		focused:   false,
		maxLength: 0, // No limit by default
	}
}

// WithPlaceholder sets the placeholder text for empty input.
func (i *Input) WithPlaceholder(placeholder string) *Input {
	i.placeholder = placeholder
	return i
}

// WithValue sets the initial value of the input.
func (i *Input) WithValue(value string) *Input {
	i.value = value
	i.cursor = len(value)
	return i
}

// WithWidth sets the width of the input field.
func (i *Input) WithWidth(width int) *Input {
	i.width = width
	return i
}

// WithPrompt sets a prompt prefix for the input.
func (i *Input) WithPrompt(prompt string) *Input {
	i.prompt = prompt
	return i
}

// WithMaxLength sets the maximum character length.
func (i *Input) WithMaxLength(max int) *Input {
	i.maxLength = max
	return i
}

// Focus focuses the input field.
func (i *Input) Focus() {
	i.focused = true
}

// Blur unfocuses the input field.
func (i *Input) Blur() {
	i.focused = false
}

// Focused returns whether the input is focused.
func (i Input) Focused() bool {
	return i.focused
}

// GetValue returns the current input value.
func (i Input) GetValue() string {
	return i.value
}

// SetValue sets the input value.
func (i *Input) SetValue(value string) {
	i.value = value
	i.cursor = len(value)
}

// Clear clears the input value.
func (i *Input) Clear() {
	i.value = ""
	i.cursor = 0
}

// SetCursor sets the cursor position.
func (i *Input) SetCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(i.value) {
		pos = len(i.value)
	}
	i.cursor = pos
}

// IsEmpty returns true if the input value is empty.
func (i Input) IsEmpty() bool {
	return i.value == ""
}

// Init initializes the input (tea.Model interface).
func (i *Input) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (tea.Model interface).
func (i *Input) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !i.focused {
		return i, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			// Insert characters at cursor position
			for _, r := range msg.Runes {
				if i.maxLength > 0 && len(i.value) >= i.maxLength {
					break
				}
				i.value = i.value[:i.cursor] + string(r) + i.value[i.cursor:]
				i.cursor++
			}

		case tea.KeyBackspace, tea.KeyDelete:
			// Delete character before cursor
			if i.cursor > 0 {
				i.value = i.value[:i.cursor-1] + i.value[i.cursor:]
				i.cursor--
			}

		case tea.KeyLeft:
			// Move cursor left
			if i.cursor > 0 {
				i.cursor--
			}

		case tea.KeyRight:
			// Move cursor right
			if i.cursor < len(i.value) {
				i.cursor++
			}

		case tea.KeyHome, tea.KeyCtrlA:
			// Move cursor to beginning
			i.cursor = 0

		case tea.KeyEnd, tea.KeyCtrlE:
			// Move cursor to end
			i.cursor = len(i.value)

		case tea.KeyCtrlU:
			// Clear from cursor to beginning
			i.value = i.value[i.cursor:]
			i.cursor = 0

		case tea.KeyCtrlK:
			// Clear from cursor to end
			i.value = i.value[:i.cursor]
		}
	}

	return i, nil
}

// View renders the input (tea.Model interface).
func (i *Input) View() string {
	// Build the display value with cursor
	var displayValue string
	if i.value == "" && i.placeholder != "" && !i.focused {
		displayValue = i.placeholder
	} else {
		displayValue = i.renderValueWithCursor()
	}

	// Apply styles based on focus state
	var style lipgloss.Style
	if i.focused {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(i.theme.Primary.ToLipgloss()).
			Foreground(i.theme.Text.ToLipgloss()).
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(i.theme.Border.ToLipgloss()).
			Foreground(i.theme.TextMuted.ToLipgloss()).
			Padding(0, 1)
	}

	if i.width > 0 {
		style = style.Width(i.width)
	}

	// Add prompt if set
	if i.prompt != "" {
		promptStyle := lipgloss.NewStyle().
			Foreground(i.theme.Primary.ToLipgloss())
		return promptStyle.Render(i.prompt) + style.Render(displayValue)
	}

	return style.Render(displayValue)
}

// renderValueWithCursor renders the value with cursor indicator.
func (i Input) renderValueWithCursor() string {
	if i.value == "" {
		// Show cursor placeholder for empty input
		if i.focused {
			cursorStyle := lipgloss.NewStyle().
				Background(i.theme.Primary.ToLipgloss())
			return cursorStyle.Render(" ")
		}
		return ""
	}

	// Truncate if width is set
	maxDisplayLen := i.width - 4 // Account for padding and border
	if maxDisplayLen <= 0 {
		maxDisplayLen = 40 // Default width
	}

	displayValue := i.value
	cursorPos := i.cursor

	// Handle long values - scroll if needed
	if len(displayValue) > maxDisplayLen {
		// Show portion around cursor
		start := cursorPos - maxDisplayLen/2
		if start < 0 {
			start = 0
		}
		if start+maxDisplayLen > len(displayValue) {
			start = len(displayValue) - maxDisplayLen
		}
		displayValue = displayValue[start:start+maxDisplayLen]
		cursorPos = cursorPos - start
	}

	// Build the string with cursor
	before := displayValue[:cursorPos]
	atCursor := ""
	after := displayValue[cursorPos:]

	if i.focused {
		if cursorPos < len(displayValue) {
			atCursor = lipgloss.NewStyle().
				Background(i.theme.Primary.ToLipgloss()).
				Foreground(i.theme.Background.ToLipgloss()).
				Render(string(displayValue[cursorPos]))
			after = displayValue[cursorPos+1:]
		} else {
			// Cursor at end - show placeholder
			atCursor = lipgloss.NewStyle().
				Background(i.theme.Primary.ToLipgloss()).
				Render(" ")
		}
	}

	return before + atCursor + after
}

// SetWidth sets the width dynamically.
func (i *Input) SetWidth(width int) {
	i.width = width
}

// SetHeight has no effect on Input (single line).
func (i *Input) SetHeight(height int) {
	// Input is always single line, height is ignored
}

// GetCursor returns the current cursor position.
func (i Input) GetCursor() int {
	return i.cursor
}

// CharLimit returns the maximum character limit.
func (i Input) CharLimit() int {
	return i.maxLength
}

// EchoMode controls how input is displayed (for password fields).
type EchoMode int

const (
	// EchoNormal displays input normally.
	EchoNormal EchoMode = iota
	// EchoHidden hides all characters (for passwords).
	EchoHidden
	// EchoNone displays nothing (completely hidden).
	EchoNone
)

// WithEchoMode sets the echo mode for password input.
func (i *Input) WithEchoMode(mode EchoMode) *Input {
	// For now, we store this but implementation can be added
	// when needed for password fields
	return i
}

// Placeholder returns the placeholder text.
func (i Input) Placeholder() string {
	return i.placeholder
}

// Prompt returns the prompt text.
func (i Input) Prompt() string {
	return i.prompt
}

// Width returns the configured width.
func (i Input) Width() int {
	return i.width
}

// ViewPlain renders the input without border styling.
func (i Input) ViewPlain() string {
	if i.value == "" && i.placeholder != "" {
		return lipgloss.NewStyle().
			Foreground(i.theme.TextMuted.ToLipgloss()).
			Render(i.placeholder)
	}
	return i.renderValueWithCursor()
}

// ViewWithBorder renders the input with a specific border style.
func (i Input) ViewWithBorder(border Border) string {
	var style lipgloss.Style
	if i.focused {
		style = border.Focused()
	} else {
		style = border.Unfocused()
	}

	if i.width > 0 {
		style = style.Width(i.width)
	}

	return style.Render(i.renderValueWithCursor())
}

// String returns the input value as a string.
func (i Input) String() string {
	return i.value
}

// Len returns the length of the input value.
func (i Input) Len() int {
	return len(i.value)
}

// TrimSpace trims whitespace from the input value.
func (i Input) TrimSpace() Input {
	i.value = strings.TrimSpace(i.value)
	i.cursor = len(i.value)
	return i
}