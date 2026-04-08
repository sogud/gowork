// Package components provides reusable UI components for the TUI.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/tui/styles"
)

// Border provides reusable border styles with rounded corners.
type Border struct {
	theme styles.Theme

	// Pre-defined border styles
	rounded   lipgloss.Style
	focused   lipgloss.Style
	unfocused lipgloss.Style
	error     lipgloss.Style
	success   lipgloss.Style
}

// NewBorder creates a new Border component with the given theme.
func NewBorder(theme styles.Theme) Border {
	roundedBorder := lipgloss.RoundedBorder()

	return Border{
		theme: theme,

		rounded: lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(theme.Border.ToLipgloss()).
			Padding(0, 1),

		focused: lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(theme.Primary.ToLipgloss()).
			Padding(0, 1),

		unfocused: lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(theme.Border.ToLipgloss()).
			Foreground(theme.TextMuted.ToLipgloss()).
			Padding(0, 1),

		error: lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(theme.Error.ToLipgloss()).
			Padding(0, 1),

		success: lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(theme.Success.ToLipgloss()).
			Padding(0, 1),
	}
}

// Rounded returns a rounded border style.
func (b Border) Rounded() lipgloss.Style {
	return b.rounded
}

// RoundedWithTitle returns a rounded border style with a title.
// The title is rendered in the top border line.
func (b Border) RoundedWithTitle(title string, width int) lipgloss.Style {
	if title == "" {
		return b.rounded.Width(width)
	}

	// Create title style
	titleStyle := lipgloss.NewStyle().
		Foreground(b.theme.Primary.ToLipgloss()).
		Bold(true).
		Padding(0, 1)

	// Calculate title width within border
	maxTitleWidth := width - 4 // Account for border chars and padding
	if maxTitleWidth < 0 {
		maxTitleWidth = 0
	}

	// Truncate title if too long
	displayTitle := title
	if len(title) > maxTitleWidth {
		displayTitle = title[:maxTitleWidth]
	}

	// Render title separately and join with content
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(b.theme.Border.ToLipgloss()).
		Width(width).
		SetString(titleStyle.Render(displayTitle))
}

// Focused returns a focused border style (highlighted border).
func (b Border) Focused() lipgloss.Style {
	return b.focused
}

// Unfocused returns an unfocused border style (muted).
func (b Border) Unfocused() lipgloss.Style {
	return b.unfocused
}

// Error returns an error border style (red border).
func (b Border) Error() lipgloss.Style {
	return b.error
}

// Success returns a success border style (green border).
func (b Border) Success() lipgloss.Style {
	return b.success
}

// ApplyToStyle applies a rounded border to an existing style.
func (b Border) ApplyToStyle(style lipgloss.Style) lipgloss.Style {
	return style.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(b.theme.Border.ToLipgloss())
}

// Box creates a bordered box with optional title.
func (b Border) Box(title string, content string, width, height int) string {
	var boxStyle lipgloss.Style

	if title != "" {
		// Create header with title
		headerStyle := lipgloss.NewStyle().
			Foreground(b.theme.Primary.ToLipgloss()).
			Bold(true).
			Padding(0, 1)

		// Build the box with title on top
		titleLine := headerStyle.Render(title)

		contentStyle := b.rounded.Width(width)
		if height > 0 {
			contentStyle = contentStyle.Height(height - 2) // Account for title and padding
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleLine,
			contentStyle.Render(content),
		)
	}

	boxStyle = b.rounded
	if width > 0 {
		boxStyle = boxStyle.Width(width)
	}
	if height > 0 {
		boxStyle = boxStyle.Height(height)
	}

	return boxStyle.Render(content)
}

// Divider creates a horizontal divider with rounded ends.
func (b Border) Divider(width int) string {
	// Use rounded border characters for a nice divider
	left := "╭"
	right := "╮"
	middle := strings.Repeat("─", width-2)

	dividerStyle := lipgloss.NewStyle().
		Foreground(b.theme.Border.ToLipgloss())

	return dividerStyle.Render(left + middle + right)
}

// DividerLine creates a simple horizontal divider line.
func (b Border) DividerLine(width int) string {
	line := strings.Repeat("─", width)
	return lipgloss.NewStyle().
		Foreground(b.theme.Border.ToLipgloss()).
		Render(line)
}