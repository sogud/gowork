package components

import (
	"testing"

	"github.com/sogud/gowork/tui/styles"
)

func TestNewProgress(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme)

	if progress.theme != theme {
		t.Error("Expected theme to be set")
	}
	if progress.percent != 0 {
		t.Errorf("Expected default percent 0, got %f", progress.percent)
	}
}

func TestProgressWithPercent(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.5)

	if progress.percent != 0.5 {
		t.Errorf("Expected percent 0.5, got %f", progress.percent)
	}
}

func TestProgressWithWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithWidth(50)

	if progress.width != 50 {
		t.Errorf("Expected width 50, got %d", progress.width)
	}
}

func TestProgressWithLabel(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithLabel("Loading...")

	if progress.label != "Loading..." {
		t.Errorf("Expected label 'Loading...', got '%s'", progress.label)
	}
}

func TestProgressWithState(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithState(ProgressRunning)

	if progress.state != ProgressRunning {
		t.Errorf("Expected state ProgressRunning, got %v", progress.state)
	}
}

func TestProgressView(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.5).WithWidth(30)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestProgressViewRunning(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.5).WithState(ProgressRunning).WithWidth(30)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestProgressViewCompleted(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(1.0).WithState(ProgressCompleted).WithWidth(30)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestProgressViewFailed(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.3).WithState(ProgressFailed).WithWidth(30)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestProgressViewWaiting(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0).WithState(ProgressWaiting).WithWidth(30)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestProgressSetPercent(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme)

	progress.SetPercent(0.75)
	if progress.GetPercent() != 0.75 {
		t.Errorf("Expected percent 0.75, got %f", progress.GetPercent())
	}
}

func TestProgressSetPercentBoundary(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme)

	// Test below 0
	progress.SetPercent(-0.5)
	if progress.GetPercent() != 0 {
		t.Errorf("Expected percent 0 for negative input, got %f", progress.GetPercent())
	}

	// Test above 1
	progress.SetPercent(1.5)
	if progress.GetPercent() != 1 {
		t.Errorf("Expected percent 1 for >1 input, got %f", progress.GetPercent())
	}
}

func TestProgressSetState(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme)

	progress.SetState(ProgressCompleted)
	if progress.GetState() != ProgressCompleted {
		t.Errorf("Expected state ProgressCompleted, got %v", progress.GetState())
	}
}

func TestProgressSetWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme)

	progress.SetWidth(60)
	if progress.width != 60 {
		t.Errorf("Expected width 60, got %d", progress.width)
	}
}

func TestProgressRenderCompact(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.5).WithWidth(20)

	view := progress.RenderCompact()
	if view == "" {
		t.Error("RenderCompact should not be empty")
	}
}

func TestProgressWithShowPercent(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.75).WithShowPercent(true).WithWidth(30)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestProgressFullBar(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(1.0).WithWidth(20)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty for full bar")
	}
}

func TestProgressEmptyBar(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0).WithWidth(20)

	view := progress.View()
	if view == "" {
		t.Error("View should not be empty for empty bar")
	}
}

func TestProgressWithAnimated(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithAnimated(true)

	if progress.animated != true {
		t.Error("Animated should be true")
	}
}

func TestProgressIsComplete(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(1.0)

	if !progress.IsComplete() {
		t.Error("IsComplete should return true for 100%")
	}
}

func TestProgressIsFailed(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithState(ProgressFailed)

	if !progress.IsFailed() {
		t.Error("IsFailed should return true for failed state")
	}
}

func TestProgressReset(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.75).WithState(ProgressCompleted)

	progress.Reset()
	if progress.GetPercent() != 0 {
		t.Errorf("Expected percent 0 after reset, got %f", progress.GetPercent())
	}
	if progress.GetState() != ProgressWaiting {
		t.Errorf("Expected state ProgressWaiting after reset, got %v", progress.GetState())
	}
	if progress.label != "" {
		t.Errorf("Expected empty label after reset, got '%s'", progress.label)
	}
}

func TestProgressRenderWithStatus(t *testing.T) {
	theme := styles.DefaultTheme()
	progress := NewProgress(theme).WithPercent(0.5).WithState(ProgressRunning).WithWidth(30)

	view := progress.RenderWithStatus()
	if view == "" {
		t.Error("RenderWithStatus should not be empty")
	}
}