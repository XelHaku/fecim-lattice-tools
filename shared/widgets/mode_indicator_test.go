//go:build legacy_fyne

package widgets

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
)

func TestNewModeIndicator(t *testing.T) {
	config := ModeIndicatorConfig{
		MinSize: fyne.NewSize(120, 50),
		Styles: map[int]ModeStyle{
			0: {Text: "IDLE", BackgroundColor: color.RGBA{40, 40, 60, 255}},
			1: {Text: "ACTIVE", BackgroundColor: color.RGBA{80, 120, 50, 255}},
		},
	}
	mi := NewModeIndicator(config)

	if mi == nil {
		t.Fatal("NewModeIndicator returned nil")
	}
	if mi.GetMode() != 0 {
		t.Errorf("Expected initial mode 0, got %d", mi.GetMode())
	}
	if mi.MinSize() != config.MinSize {
		t.Errorf("Expected MinSize %v, got %v", config.MinSize, mi.MinSize())
	}
}

func TestModeIndicator_SetMode(t *testing.T) {
	styles := map[int]ModeStyle{
		0: {Text: "OFF"},
		1: {Text: "ON"},
		2: {Text: "ERROR"},
	}
	mi := NewModeIndicator(ModeIndicatorConfig{Styles: styles})

	testCases := []int{0, 1, 2, 1, 0}
	for _, mode := range testCases {
		mi.SetMode(mode)
		if mi.GetMode() != mode {
			t.Errorf("Expected mode %d, got %d", mode, mi.GetMode())
		}
	}
}

func TestModeIndicator_GetCurrentStyle(t *testing.T) {
	styles := map[int]ModeStyle{
		0: {Text: "IDLE", BackgroundColor: color.RGBA{40, 40, 60, 255}},
		1: {Text: "ACTIVE", BackgroundColor: color.RGBA{80, 120, 50, 255}},
	}
	fallback := ModeStyle{Text: "UNKNOWN", BackgroundColor: color.RGBA{100, 100, 100, 255}}
	mi := NewModeIndicator(ModeIndicatorConfig{
		Styles:       styles,
		DefaultStyle: fallback,
	})

	// Test known mode
	mi.SetMode(1)
	style := mi.GetCurrentStyle()
	if style.Text != "ACTIVE" {
		t.Errorf("Expected style text 'ACTIVE', got '%s'", style.Text)
	}

	// Test unknown mode falls back to default
	mi.SetMode(99)
	style = mi.GetCurrentStyle()
	if style.Text != "UNKNOWN" {
		t.Errorf("Expected fallback style text 'UNKNOWN', got '%s'", style.Text)
	}
}

func TestModeIndicator_SetStyle(t *testing.T) {
	mi := NewModeIndicator(ModeIndicatorConfig{})

	// Add a new style
	newStyle := ModeStyle{
		Text:            "CUSTOM",
		BackgroundColor: color.RGBA{200, 100, 50, 255},
		BorderColor:     color.RGBA{255, 150, 100, 255},
	}
	mi.SetStyle(5, newStyle)
	mi.SetMode(5)

	style := mi.GetCurrentStyle()
	if style.Text != "CUSTOM" {
		t.Errorf("Expected style text 'CUSTOM', got '%s'", style.Text)
	}
}

func TestDefaultModeStyles(t *testing.T) {
	styles := DefaultModeStyles()

	expectedModes := []int{0, 1, 2, 3}
	for _, mode := range expectedModes {
		if _, ok := styles[mode]; !ok {
			t.Errorf("DefaultModeStyles missing mode %d", mode)
		}
	}

	// Check that each style has required fields
	for mode, style := range styles {
		if style.Text == "" {
			t.Errorf("Mode %d has empty Text", mode)
		}
	}
}

func TestModeIndicator_DefaultConfig(t *testing.T) {
	// Empty config should use defaults
	mi := NewModeIndicator(ModeIndicatorConfig{})

	if mi.MinSize().Width <= 0 || mi.MinSize().Height <= 0 {
		t.Error("Default MinSize should be positive")
	}

	style := mi.GetCurrentStyle()
	if style.Text == "" {
		t.Error("Default fallback style should have text")
	}
}
