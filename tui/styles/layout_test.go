package styles

import (
	"testing"
)

func TestLayoutStyles(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	// Test that all layout styles are defined by checking they can render
	tests := []struct {
		name  string
		style func() string
	}{
		{"Box", func() string { return layout.Box.Render("test") }},
		{"Title", func() string { return layout.Title.Render("test") }},
		{"Header", func() string { return layout.Header.Render("test") }},
		{"Body", func() string { return layout.Body.Render("test") }},
		{"Footer", func() string { return layout.Footer.Render("test") }},
		{"Focused", func() string { return layout.Focused.Render("test") }},
		{"Unfocused", func() string { return layout.Unfocused.Render("test") }},
		{"HelpText", func() string { return layout.HelpText.Render("test") }},
		{"ErrorText", func() string { return layout.ErrorText.Render("test") }},
		{"SuccessText", func() string { return layout.SuccessText.Render("test") }},
		{"WarningText", func() string { return layout.WarningText.Render("test") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.style()
			if result == "" {
				t.Errorf("%s style failed to render", tt.name)
			}
		})
	}
}

func TestLayoutBoxWithBorder(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	box := layout.BoxWithBorder("Test Title", 40, 10)
	rendered := box.Render("content")

	if rendered == "" {
		t.Error("BoxWithBorder failed to render")
	}
}

func TestLayoutCenterText(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	centered := layout.CenterText("Test", 20)
	if centered == "" {
		t.Error("CenterText failed")
	}
}

func TestLayoutPadding(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	// Test that padding is applied correctly
	style := layout.Body.Padding(1, 2)
	rendered := style.Render("test")

	if rendered == "" {
		t.Error("Padding style failed to render")
	}
}

func TestLayoutMargin(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	style := layout.Body.Margin(1, 2)
	rendered := style.Render("test")

	if rendered == "" {
		t.Error("Margin style failed to render")
	}
}

func TestLayoutWidth(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	width := 80
	style := layout.Box.Width(width)
	rendered := style.Render("test")

	if rendered == "" {
		t.Error("Width style failed to render")
	}
}

func TestLayoutJoinVertical(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	top := "Top"
	middle := "Middle"
	bottom := "Bottom"

	result := layout.JoinVertical(top, middle, bottom)
	if result == "" {
		t.Error("JoinVertical failed")
	}
}

func TestLayoutJoinHorizontal(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	left := "Left"
	right := "Right"

	result := layout.JoinHorizontal(left, right)
	if result == "" {
		t.Error("JoinHorizontal failed")
	}
}

func TestLayoutProgressBar(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	// Test progress bar at various percentages
	tests := []float64{0, 0.25, 0.5, 0.75, 1.0, 1.5, -0.5}

	for _, pct := range tests {
		bar := layout.ProgressBar(pct, 20)
		if bar == "" {
			t.Errorf("ProgressBar(%f) failed", pct)
		}
	}
}

func TestLayoutTruncateText(t *testing.T) {
	theme := DefaultTheme()
	layout := NewLayout(theme)

	tests := []struct {
		text   string
		maxLen int
	}{
		{"Short", 10},
		{"This is a longer text that should be truncated", 20},
		{"Exact length", 11},
		{"", 10},
	}

	for _, tt := range tests {
		result := layout.TruncateText(tt.text, tt.maxLen)
		if len(result) > tt.maxLen+3 { // +3 for "..."
			t.Errorf("TruncateText result too long: got %d, max %d", len(result), tt.maxLen)
		}
	}
}