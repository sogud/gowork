// Package styles provides styling definitions for the TUI.
package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color represents an RGB color.
type Color struct {
	R uint8
	G uint8
	B uint8
}

// ToLipgloss converts a Color to a lipgloss.Color.
func (c Color) ToLipgloss() lipgloss.Color {
	return lipgloss.Color(c.Hex())
}

// String returns the hex color string.
func (c Color) String() string {
	return c.Hex()
}

// Hex returns the hex color string in #RRGGBB format.
func (c Color) Hex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// FontStyle defines font styling options.
type FontStyle struct {
	Bold      bool
	Italic    bool
	Underline bool
}

// Apply applies the font style to a lipgloss.Style.
func (f FontStyle) Apply(style lipgloss.Style) lipgloss.Style {
	if f.Bold {
		style = style.Bold(true)
	}
	if f.Italic {
		style = style.Italic(true)
	}
	if f.Underline {
		style = style.Underline(true)
	}
	return style
}

// Theme defines the color and font theme for the TUI.
type Theme struct {
	// Colors
	Primary    Color
	Secondary  Color
	Background Color
	Surface    Color
	Text       Color
	TextMuted  Color
	Success    Color
	Warning    Color
	Error      Color
	Border     Color

	// Fonts
	TitleFont  FontStyle
	HeaderFont FontStyle
	BodyFont   FontStyle
	CodeFont   FontStyle
}

// DefaultTheme creates a new theme with default colors and fonts.
func DefaultTheme() Theme {
	return Theme{
		// Colors - Dark theme inspired by modern terminal apps
		Primary:    Color{R: 139, G: 92, B: 246}, // Purple
		Secondary:  Color{R: 59, G: 130, B: 246},  // Blue
		Background: Color{R: 17, G: 24, B: 39},    // Dark blue-gray
		Surface:    Color{R: 30, G: 41, B: 59},    // Lighter blue-gray
		Text:       Color{R: 241, G: 245, B: 249}, // Almost white
		TextMuted:  Color{R: 148, G: 163, B: 184}, // Gray
		Success:    Color{R: 34, G: 197, B: 94},   // Green
		Warning:    Color{R: 251, G: 191, B: 36},  // Amber
		Error:      Color{R: 239, G: 68, B: 68},   // Red
		Border:     Color{R: 71, G: 85, B: 105},   // Slate

		// Fonts
		TitleFont: FontStyle{
			Bold:      true,
			Italic:    false,
			Underline: false,
		},
		HeaderFont: FontStyle{
			Bold:      true,
			Italic:    false,
			Underline: false,
		},
		BodyFont: FontStyle{
			Bold:      false,
			Italic:    false,
			Underline: false,
		},
		CodeFont: FontStyle{
			Bold:      false,
			Italic:    false,
			Underline: false,
		},
	}
}

// CreateStyle creates a lipgloss.Style with the given options.
func (t Theme) CreateStyle(fg, bg Color, bold, italic, underline bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(fg.ToLipgloss()).
		Background(bg.ToLipgloss())

	if bold {
		style = style.Bold(true)
	}
	if italic {
		style = style.Italic(true)
	}
	if underline {
		style = style.Underline(true)
	}

	return style
}

// GetStatusColor returns the appropriate color for a status string.
func (t Theme) GetStatusColor(status string) Color {
	switch strings.ToLower(status) {
	case "running", "completed", "success":
		return t.Success
	case "waiting", "pending":
		return t.TextMuted
	case "failed", "error":
		return t.Error
	default:
		return t.Text
	}
}

// TitleStyle returns a styled string for titles.
func (t Theme) TitleStyle(text string) string {
	return t.TitleFont.Apply(
		lipgloss.NewStyle().Foreground(t.Primary.ToLipgloss()),
	).Render(text)
}

// HeaderStyle returns a styled string for headers.
func (t Theme) HeaderStyle(text string) string {
	return t.HeaderFont.Apply(
		lipgloss.NewStyle().Foreground(t.Secondary.ToLipgloss()),
	).Render(text)
}

// ErrorStyle returns a styled string for error messages.
func (t Theme) ErrorStyle(text string) string {
	return lipgloss.NewStyle().Foreground(t.Error.ToLipgloss()).Render(text)
}

// SuccessStyle returns a styled string for success messages.
func (t Theme) SuccessStyle(text string) string {
	return lipgloss.NewStyle().Foreground(t.Success.ToLipgloss()).Render(text)
}

// WarningStyle returns a styled string for warning messages.
func (t Theme) WarningStyle(text string) string {
	return lipgloss.NewStyle().Foreground(t.Warning.ToLipgloss()).Render(text)
}

// MutedStyle returns a styled string for muted text.
func (t Theme) MutedStyle(text string) string {
	return lipgloss.NewStyle().Foreground(t.TextMuted.ToLipgloss()).Render(text)
}