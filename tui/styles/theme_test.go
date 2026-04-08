package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestThemeColors(t *testing.T) {
	theme := DefaultTheme()

	// Test that all color fields are defined (not empty)
	if theme.Primary == (Color{}) {
		t.Error("Primary color not defined")
	}
	if theme.Secondary == (Color{}) {
		t.Error("Secondary color not defined")
	}
	if theme.Background == (Color{}) {
		t.Error("Background color not defined")
	}
	if theme.Surface == (Color{}) {
		t.Error("Surface color not defined")
	}
	if theme.Text == (Color{}) {
		t.Error("Text color not defined")
	}
	if theme.TextMuted == (Color{}) {
		t.Error("TextMuted color not defined")
	}
	if theme.Success == (Color{}) {
		t.Error("Success color not defined")
	}
	if theme.Warning == (Color{}) {
		t.Error("Warning color not defined")
	}
	if theme.Error == (Color{}) {
		t.Error("Error color not defined")
	}
	if theme.Border == (Color{}) {
		t.Error("Border color not defined")
	}
}

func TestThemeFonts(t *testing.T) {
	theme := DefaultTheme()

	// Test that font styles are defined by checking they can apply
	tests := []struct {
		name string
		font FontStyle
	}{
		{"TitleFont", theme.TitleFont},
		{"HeaderFont", theme.HeaderFont},
		{"BodyFont", theme.BodyFont},
		{"CodeFont", theme.CodeFont},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the font style can be applied
			style := tt.font.Apply(lipgloss.NewStyle())
			rendered := style.Render("test")
			if rendered == "" {
				t.Errorf("%s failed to apply", tt.name)
			}
		})
	}
}

func TestColorToLipgloss(t *testing.T) {
	color := Color{R: 255, G: 128, B: 64}
	lgColor := color.ToLipgloss()

	// Verify it creates a valid lipgloss color
	// The actual color value is computed by lipgloss
	str := lipgloss.NewStyle().Foreground(lgColor).Render("test")
	if str == "" {
		t.Error("Failed to render with lipgloss color")
	}
}

func TestFontStyleApply(t *testing.T) {
	font := FontStyle{
		Bold:      true,
		Italic:    false,
		Underline: true,
	}

	style := font.Apply(lipgloss.NewStyle())
	rendered := style.Render("test")

	if rendered == "" {
		t.Error("Failed to apply font style")
	}
}

func TestThemeCreateStyle(t *testing.T) {
	theme := DefaultTheme()

	style := theme.CreateStyle(
		theme.Primary,
		theme.Background,
		true,  // bold
		false, // italic
		false, // underline
	)

	rendered := style.Render("test")
	if rendered == "" {
		t.Error("Failed to create and render style")
	}
}

func TestThemeGetStatusColor(t *testing.T) {
	theme := DefaultTheme()

	tests := []struct {
		status   string
		expected Color
	}{
		{"running", theme.Success},
		{"completed", theme.Success},
		{"success", theme.Success},
		{"waiting", theme.TextMuted},
		{"pending", theme.TextMuted},
		{"failed", theme.Error},
		{"error", theme.Error},
		{"unknown", theme.Text},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := theme.GetStatusColor(tt.status)
			if got != tt.expected {
				t.Errorf("GetStatusColor(%s) = %v, want %v", tt.status, got, tt.expected)
			}
		})
	}
}