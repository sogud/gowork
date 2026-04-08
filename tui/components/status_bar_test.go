package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/tui/styles"
)

func TestNewStatusBar(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	if bar.theme != theme {
		t.Error("Expected theme to be set")
	}
}

func TestStatusBarWithWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80)

	if bar.width != 80 {
		t.Errorf("Expected width 80, got %d", bar.width)
	}
}

func TestStatusBarWithLeftText(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithLeftText("Ready")

	if bar.leftText != "Ready" {
		t.Errorf("Expected leftText 'Ready', got '%s'", bar.leftText)
	}
}

func TestStatusBarWithRightText(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithRightText("Ctrl+C to quit")

	if bar.rightText != "Ctrl+C to quit" {
		t.Errorf("Expected rightText 'Ctrl+C to quit', got '%s'", bar.rightText)
	}
}

func TestStatusBarWithStatus(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithStatus("Processing", StatusInfo)

	if bar.statusMsg != "Processing" {
		t.Errorf("Expected statusMsg 'Processing', got '%s'", bar.statusMsg)
	}
	if bar.statusType != StatusInfo {
		t.Errorf("Expected statusType StatusInfo, got %v", bar.statusType)
	}
}

func TestStatusBarInit(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	cmd := bar.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestStatusBarUpdate(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	// Status bar typically doesn't handle key events directly
	updatedModel, _ := bar.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedBar, ok := updatedModel.(*StatusBar)
	if !ok {
		t.Fatal("Updated model should be a StatusBar")
	}
	// Status bar should remain unchanged for most key events
	_ = updatedBar
}

func TestStatusBarView(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80).WithLeftText("Ready")

	view := bar.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestStatusBarViewWithRightText(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80).WithRightText("Ctrl+C to quit")

	view := bar.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestStatusBarViewWithStatus(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80).WithStatus("Error!", StatusError)

	view := bar.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestStatusBarSetWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	bar.SetWidth(100)
	if bar.width != 100 {
		t.Errorf("Expected width 100, got %d", bar.width)
	}
}

func TestStatusBarSetLeftText(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	bar.SetLeftText("New text")
	if bar.leftText != "New text" {
		t.Errorf("Expected leftText 'New text', got '%s'", bar.leftText)
	}
}

func TestStatusBarSetRightText(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	bar.SetRightText("New right")
	if bar.rightText != "New right" {
		t.Errorf("Expected rightText 'New right', got '%s'", bar.rightText)
	}
}

func TestStatusBarSetStatus(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme)

	bar.SetStatus("Processing...", StatusInfo)
	if bar.statusMsg != "Processing..." {
		t.Errorf("Expected statusMsg 'Processing...', got '%s'", bar.statusMsg)
	}
	if bar.statusType != StatusInfo {
		t.Errorf("Expected statusType StatusInfo, got %v", bar.statusType)
	}
}

func TestStatusBarClearStatus(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithStatus("Some message", StatusInfo)

	bar.ClearStatus()
	if bar.statusMsg != "" {
		t.Errorf("Expected empty statusMsg after ClearStatus, got '%s'", bar.statusMsg)
	}
	if bar.statusType != StatusNone {
		t.Errorf("Expected statusType StatusNone after ClearStatus, got %v", bar.statusType)
	}
}

func TestStatusBarStatusTypes(t *testing.T) {
	theme := styles.DefaultTheme()

	tests := []struct {
		status   StatusType
		expected string
	}{
		{StatusNone, "None"},
		{StatusInfo, "Info"},
		{StatusSuccess, "Success"},
		{StatusWarning, "Warning"},
		{StatusError, "Error"},
	}

	for _, tt := range tests {
		bar := NewStatusBar(theme).WithStatus("test", tt.status)
		if bar.statusType != tt.status {
			t.Errorf("Expected status %s, got %v", tt.expected, bar.statusType)
		}
	}
}

func TestStatusBarRender(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80).WithLeftText("Test")

	view := bar.Render()
	if view == "" {
		t.Error("Render should not be empty")
	}
}

func TestStatusBarRenderWithBorder(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80)

	view := bar.RenderWithBorder()
	if view == "" {
		t.Error("RenderWithBorder should not be empty")
	}
}

func TestStatusBarString(t *testing.T) {
	theme := styles.DefaultTheme()
	bar := NewStatusBar(theme).WithWidth(80)

	str := bar.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}