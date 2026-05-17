//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"image/color"
	"testing"
)

func TestColorLegendCreation(t *testing.T) {
	legend := NewColorLegend(0.0, 1.0, "mV", true, ViridisColor)
	if legend == nil {
		t.Error("NewColorLegend returned nil")
	}
	if legend.minValue != 0.0 {
		t.Errorf("Expected minValue 0.0, got %f", legend.minValue)
	}
	if legend.maxValue != 1.0 {
		t.Errorf("Expected maxValue 1.0, got %f", legend.maxValue)
	}
	if legend.units != "mV" {
		t.Errorf("Expected units 'mV', got '%s'", legend.units)
	}
	if !legend.vertical {
		t.Error("Expected vertical=true")
	}
}

func TestColorLegendSetRange(t *testing.T) {
	// Skip this test as it requires a Fyne app context for fyne.Do()
	// SetRange is tested in integration tests with the actual GUI
	t.Skip("SetRange requires Fyne app context")
}

func TestViridisColor(t *testing.T) {
	c := ViridisColor(0.5)
	if c.A != 255 {
		t.Errorf("Expected alpha 255, got %d", c.A)
	}
}

func TestBlueWhiteRedColor(t *testing.T) {
	// Test low end (blue)
	c1 := BlueWhiteRedColor(0.0)
	if c1.B != 255 {
		t.Errorf("Expected blue=255 at t=0, got %d", c1.B)
	}

	// Test middle (white) - at t=0.5, s=1.0, so all channels are 255
	c2 := BlueWhiteRedColor(0.5)
	if c2.R != 255 || c2.G != 255 || c2.B != 255 {
		t.Errorf("Expected white at t=0.5, got R=%d G=%d B=%d", c2.R, c2.G, c2.B)
	}

	// Test high end (red)
	c3 := BlueWhiteRedColor(1.0)
	if c3.R != 255 {
		t.Errorf("Expected red=255 at t=1, got %d", c3.R)
	}
}

func TestGreenToRedColor(t *testing.T) {
	c1 := GreenToRedColor(0.0)
	if c1.R != 0 || c1.G < 150 {
		t.Errorf("Expected green at t=0, got R=%d G=%d", c1.R, c1.G)
	}

	c2 := GreenToRedColor(1.0)
	if c2.R != 255 {
		t.Errorf("Expected red at t=1, got R=%d", c2.R)
	}
}

func TestErrorColor(t *testing.T) {
	// Low error (black)
	c1 := ErrorColor(0.0)
	if c1.R != 0 || c1.G != 0 {
		t.Errorf("Expected black at t=0, got R=%d G=%d", c1.R, c1.G)
	}

	// Mid error (yellow)
	c2 := ErrorColor(0.5)
	if c2.R != 255 || c2.G != 200 {
		t.Errorf("Expected yellow at t=0.5, got R=%d G=%d", c2.R, c2.G)
	}

	// High error (red)
	c3 := ErrorColor(1.0)
	if c3.R != 255 || c3.G != 0 {
		t.Errorf("Expected red at t=1, got R=%d G=%d", c3.R, c3.G)
	}
}

func BenchmarkViridisColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ViridisColor(0.5)
	}
}

func BenchmarkBlueWhiteRedColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BlueWhiteRedColor(0.5)
	}
}

// Ensure color functions return valid RGBA
func TestColorFunctionsReturnRGBA(t *testing.T) {
	colorFuncs := []func(float64) color.RGBA{
		ViridisColor,
		BlueWhiteRedColor,
		GreenToRedColor,
		BlueToYellowColor,
		ErrorColor,
	}

	testValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, fn := range colorFuncs {
		for _, v := range testValues {
			c := fn(v)
			if c.A != 255 {
				t.Errorf("Expected alpha=255 for value %f, got %d", v, c.A)
			}
		}
	}
}
