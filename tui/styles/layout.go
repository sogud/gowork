package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout provides common layout styles for the TUI.
type Layout struct {
	theme Theme

	// Core styles
	Box      lipgloss.Style
	Title    lipgloss.Style
	Header   lipgloss.Style
	Body     lipgloss.Style
	Footer   lipgloss.Style
	Focused  lipgloss.Style
	Unfocused lipgloss.Style

	// Text styles
	HelpText    lipgloss.Style
	ErrorText   lipgloss.Style
	SuccessText lipgloss.Style
	WarningText lipgloss.Style
	MutedText   lipgloss.Style
}

// NewLayout creates a new Layout with the given theme.
func NewLayout(theme Theme) Layout {
	border := lipgloss.RoundedBorder()

	return Layout{
		theme: theme,

		// Core styles
		Box: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border.ToLipgloss()).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(theme.Primary.ToLipgloss()).
			Bold(true).
			Padding(0, 1),

		Header: lipgloss.NewStyle().
			Foreground(theme.Secondary.ToLipgloss()).
			Bold(true).
			Padding(0, 1),

		Body: lipgloss.NewStyle().
			Foreground(theme.Text.ToLipgloss()).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(theme.TextMuted.ToLipgloss()).
			Padding(0, 1),

		Focused: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Primary.ToLipgloss()).
			Foreground(theme.Text.ToLipgloss()),

		Unfocused: lipgloss.NewStyle().
			Border(border).
			BorderForeground(theme.Border.ToLipgloss()).
			Foreground(theme.TextMuted.ToLipgloss()),

		// Text styles
		HelpText: lipgloss.NewStyle().
			Foreground(theme.TextMuted.ToLipgloss()).
			Italic(true),

		ErrorText: lipgloss.NewStyle().
			Foreground(theme.Error.ToLipgloss()).
			Bold(true),

		SuccessText: lipgloss.NewStyle().
			Foreground(theme.Success.ToLipgloss()).
			Bold(true),

		WarningText: lipgloss.NewStyle().
			Foreground(theme.Warning.ToLipgloss()).
			Bold(true),

		MutedText: lipgloss.NewStyle().
			Foreground(theme.TextMuted.ToLipgloss()),
	}
}

// BoxWithBorder creates a bordered box with an optional title.
func (l Layout) BoxWithBorder(title string, width, height int) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(l.theme.Border.ToLipgloss()).
		Padding(0, 1)

	if width > 0 {
		style = style.Width(width)
	}
	if height > 0 {
		style = style.Height(height)
	}

	return style
}

// CenterText centers text within a given width.
func (l Layout) CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(text)
}

// JoinVertical joins multiple strings vertically.
func (l Layout) JoinVertical(strs ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, strs...)
}

// JoinHorizontal joins multiple strings horizontally.
func (l Layout) JoinHorizontal(strs ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, strs...)
}

// ProgressBar creates a progress bar with the given percentage and width.
func (l Layout) ProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}

	empty := width - filled
	if empty < 0 {
		empty = 0
	}

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	return l.theme.SuccessStyle(filledBar) + l.MutedText.Render(emptyBar)
}

// TruncateText truncates text to the given maximum length, adding "..." if truncated.
func (l Layout) TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	if maxLen <= 3 {
		return text[:maxLen]
	}

	return text[:maxLen-3] + "..."
}

// FocusedBox returns a focused box style.
func (l Layout) FocusedBox(width int) lipgloss.Style {
	return l.Focused.Width(width)
}

// UnfocusedBox returns an unfocused box style.
func (l Layout) UnfocusedBox(width int) lipgloss.Style {
	return l.Unfocused.Width(width)
}

// StatusBadge creates a styled status badge.
func (l Layout) StatusBadge(status string) string {
	var color Color
	var text string

	switch strings.ToLower(status) {
	case "running":
		color = l.theme.Success
		text = "Running"
	case "completed", "success":
		color = l.theme.Success
		text = "Completed"
	case "waiting", "pending":
		color = l.theme.TextMuted
		text = "Waiting"
	case "failed", "error":
		color = l.theme.Error
		text = "Failed"
	default:
		color = l.theme.Text
		text = status
	}

	style := lipgloss.NewStyle().
		Foreground(color.ToLipgloss()).
		Padding(0, 1).
		Bold(true)

	return style.Render(text)
}

// KeyValue creates a key-value pair display with the key in muted color.
func (l Layout) KeyValue(key, value string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(l.theme.TextMuted.ToLipgloss()).
		Bold(true)

	return keyStyle.Render(key+":") + " " + value
}

// Section creates a section with a header and content.
func (l Layout) Section(header string, content string, width int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(l.theme.Secondary.ToLipgloss()).
		Bold(true).
		Underline(true).
		MarginBottom(1)

	contentStyle := lipgloss.NewStyle().
		Foreground(l.theme.Text.ToLipgloss()).
		PaddingLeft(2)

	return l.JoinVertical(
		headerStyle.Render(header),
		contentStyle.Render(content),
	)
}

// List creates a bullet list from the given items.
func (l Layout) List(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString("  • ")
		sb.WriteString(item)
		sb.WriteString("\n")
	}
	return sb.String()
}

// HeaderLine creates a header line with a title and optional subtitle.
func (l Layout) HeaderLine(title, subtitle string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(l.theme.Primary.ToLipgloss()).
		Bold(true)

	if subtitle != "" {
		subtitleStyle := lipgloss.NewStyle().
			Foreground(l.theme.TextMuted.ToLipgloss())
		return titleStyle.Render(title) + " " + subtitleStyle.Render(subtitle)
	}

	return titleStyle.Render(title)
}

// Divider creates a horizontal divider line.
func (l Layout) Divider(width int) string {
	div := strings.Repeat("─", width)
	return lipgloss.NewStyle().
		Foreground(l.theme.Border.ToLipgloss()).
		Render(div)
}