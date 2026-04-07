package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/tui/styles"
)

func TestNewBorder(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	if border.theme != theme {
		t.Error("Expected theme to be set")
	}
}

func TestBorderRounded(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Rounded()
	rendered := style.Render("test content")

	if rendered == "" {
		t.Error("Rounded border should render content")
	}
}

func TestBorderRoundedWithTitle(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.RoundedWithTitle("Test Title", 40)
	rendered := style.Render("test content")

	if rendered == "" {
		t.Error("RoundedWithTitle should render content")
	}
}

func TestBorderFocused(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Focused()
	rendered := style.Render("test content")

	if rendered == "" {
		t.Error("Focused border should render content")
	}
}

func TestBorderUnfocused(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Unfocused()
	rendered := style.Render("test content")

	if rendered == "" {
		t.Error("Unfocused border should render content")
	}
}

func TestBorderError(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Error()
	rendered := style.Render("test content")

	if rendered == "" {
		t.Error("Error border should render content")
	}
}

func TestBorderSuccess(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Success()
	rendered := style.Render("test content")

	if rendered == "" {
		t.Error("Success border should render content")
	}
}

func TestBorderWithWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Rounded().Width(50)
	if style.GetWidth() != 50 {
		t.Errorf("Expected width 50, got %d", style.GetWidth())
	}
}

func TestBorderWithHeight(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	style := border.Rounded().Height(10)
	if style.GetHeight() != 10 {
		t.Errorf("Expected height 10, got %d", style.GetHeight())
	}
}

func TestBorderApplyToStyle(t *testing.T) {
	theme := styles.DefaultTheme()
	border := NewBorder(theme)

	baseStyle := lipgloss.NewStyle().Padding(1, 2)
	style := border.ApplyToStyle(baseStyle)

	rendered := style.Render("test")
	if rendered == "" {
		t.Error("ApplyToStyle should produce renderable style")
	}
}