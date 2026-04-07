package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/tui/styles"
)

func TestNewInput(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme)

	if input.theme != theme {
		t.Error("Expected theme to be set")
	}
	if input.focused {
		t.Error("Input should not be focused by default")
	}
}

func TestInputWithPlaceholder(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithPlaceholder("Enter text...")

	if input.placeholder != "Enter text..." {
		t.Errorf("Expected placeholder 'Enter text...', got '%s'", input.placeholder)
	}
}

func TestInputWithValue(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("initial value")

	if input.value != "initial value" {
		t.Errorf("Expected value 'initial value', got '%s'", input.value)
	}
}

func TestInputWithWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithWidth(50)

	if input.width != 50 {
		t.Errorf("Expected width 50, got %d", input.width)
	}
}

func TestInputFocus(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme)

	// Focus the input
	input.Focus()
	if !input.focused {
		t.Error("Input should be focused after Focus()")
	}

	// Blur the input
	input.Blur()
	if input.focused {
		t.Error("Input should not be focused after Blur()")
	}
}

func TestInputInit(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme)

	cmd := input.Init()
	// Init should return nil (no initial command)
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestInputUpdateKeyMsg(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme)
	input.Focus()

	// Test character input
	updatedModel, _ := input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updatedInput, ok := updatedModel.(*Input)
	if !ok {
		t.Fatal("Updated model should be an Input")
	}
	if updatedInput.GetValue() != "a" {
		t.Errorf("Expected value 'a', got '%s'", updatedInput.GetValue())
	}
}

func TestInputUpdateBackspace(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("test")
	input.Focus()

	// Test backspace
	updatedModel, _ := input.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updatedInput, ok := updatedModel.(*Input)
	if !ok {
		t.Fatal("Updated model should be an Input")
	}
	if updatedInput.GetValue() != "tes" {
		t.Errorf("Expected value 'tes', got '%s'", updatedInput.GetValue())
	}
}

func TestInputUpdateNotFocused(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("test") // Not focused

	// Key messages should be ignored when not focused
	updatedModel, _ := input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updatedInput, ok := updatedModel.(*Input)
	if !ok {
		t.Fatal("Updated model should be an Input")
	}
	if updatedInput.GetValue() != "test" {
		t.Errorf("Value should not change when not focused, got '%s'", updatedInput.GetValue())
	}
}

func TestInputViewFocused(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithWidth(30)
	input.Focus()

	view := input.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestInputViewUnfocused(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithWidth(30).WithValue("test")

	view := input.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestInputViewWithPlaceholder(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithPlaceholder("Enter...").WithWidth(30)

	view := input.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestInputGetValue(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("my value")

	if input.GetValue() != "my value" {
		t.Errorf("Expected GetValue to return 'my value', got '%s'", input.GetValue())
	}
}

func TestInputSetValue(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme)

	input.SetValue("new value")
	if input.GetValue() != "new value" {
		t.Errorf("Expected value 'new value', got '%s'", input.GetValue())
	}
}

func TestInputClear(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("test")

	input.Clear()
	if input.GetValue() != "" {
		t.Errorf("Expected empty value after Clear, got '%s'", input.GetValue())
	}
}

func TestInputSetCursor(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("test")
	input.Focus()

	input.SetCursor(2)
	if input.GetCursor() != 2 {
		t.Errorf("Expected cursor position 2, got %d", input.GetCursor())
	}
}

func TestInputWithPrompt(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithPrompt("> ")

	if input.prompt != "> " {
		t.Errorf("Expected prompt '> ', got '%s'", input.prompt)
	}
}

func TestInputCursorPosition(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme).WithValue("hello")
	input.Focus()

	// Move cursor left
	updatedModel, _ := input.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updatedInput, ok := updatedModel.(*Input)
	if !ok {
		t.Fatal("Updated model should be an Input")
	}
	if updatedInput.GetCursor() != 4 {
		t.Errorf("Expected cursor at 4 after left, got %d", updatedInput.GetCursor())
	}
}

func TestInputEmptyValue(t *testing.T) {
	theme := styles.DefaultTheme()
	input := NewInput(theme)

	if input.GetValue() != "" {
		t.Errorf("Expected empty value, got '%s'", input.GetValue())
	}
	if input.IsEmpty() != true {
		t.Error("IsEmpty should return true for empty input")
	}
}