//go:build legacy_fyne

package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestConstrainSize(t *testing.T) {
	tests := []struct {
		name      string
		allocated fyne.Size
		minSize   fyne.Size
		expect    fyne.Size
	}{
		{
			name:      "allocated larger than min - constrain",
			allocated: fyne.NewSize(200, 150),
			minSize:   fyne.NewSize(100, 100),
			expect:    fyne.NewSize(100, 100),
		},
		{
			name:      "allocated smaller than min - keep allocated",
			allocated: fyne.NewSize(50, 50),
			minSize:   fyne.NewSize(100, 100),
			expect:    fyne.NewSize(50, 50),
		},
		{
			name:      "allocated equal to min",
			allocated: fyne.NewSize(100, 100),
			minSize:   fyne.NewSize(100, 100),
			expect:    fyne.NewSize(100, 100),
		},
		{
			name:      "min size is zero - no constraint",
			allocated: fyne.NewSize(200, 200),
			minSize:   fyne.NewSize(0, 0),
			expect:    fyne.NewSize(200, 200),
		},
		{
			name:      "partial constraint - width only",
			allocated: fyne.NewSize(200, 50),
			minSize:   fyne.NewSize(100, 100),
			expect:    fyne.NewSize(100, 50),
		},
		{
			name:      "partial constraint - height only",
			allocated: fyne.NewSize(50, 200),
			minSize:   fyne.NewSize(100, 100),
			expect:    fyne.NewSize(50, 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstrainSize(tt.allocated, tt.minSize)
			if result != tt.expect {
				t.Errorf("ConstrainSize(%v, %v) = %v, expected %v",
					tt.allocated, tt.minSize, result, tt.expect)
			}
		})
	}
}

func TestCenterInSize(t *testing.T) {
	tests := []struct {
		name      string
		innerSize fyne.Size
		outerSize fyne.Size
		expect    fyne.Position
	}{
		{
			name:      "center small in large",
			innerSize: fyne.NewSize(50, 50),
			outerSize: fyne.NewSize(200, 200),
			expect:    fyne.NewPos(75, 75),
		},
		{
			name:      "same size - position at origin",
			innerSize: fyne.NewSize(100, 100),
			outerSize: fyne.NewSize(100, 100),
			expect:    fyne.NewPos(0, 0),
		},
		{
			name:      "inner larger than outer - negative position",
			innerSize: fyne.NewSize(200, 200),
			outerSize: fyne.NewSize(100, 100),
			expect:    fyne.NewPos(-50, -50),
		},
		{
			name:      "asymmetric sizes",
			innerSize: fyne.NewSize(100, 50),
			outerSize: fyne.NewSize(200, 100),
			expect:    fyne.NewPos(50, 25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CenterInSize(tt.innerSize, tt.outerSize)
			if result != tt.expect {
				t.Errorf("CenterInSize(%v, %v) = %v, expected %v",
					tt.innerSize, tt.outerSize, result, tt.expect)
			}
		})
	}
}

func TestDebugLog(t *testing.T) {
	// Should not panic when called
	DebugLog("test message %d", 42)
}

func TestDebugLayoutCall(t *testing.T) {
	// Reset counters before test
	ResetLayoutCounts()

	// Normal call should return false
	result := DebugLayoutCall("TestWidget", fyne.NewSize(100, 100))
	if result {
		t.Error("first call should not indicate loop")
	}

	// Multiple calls should still be fine
	for i := 0; i < 100; i++ {
		DebugLayoutCall("TestWidget", fyne.NewSize(100, 100))
	}

	// Clean up
	ResetLayoutCounts()
}

func TestDebugRefreshCall(t *testing.T) {
	// Should not panic
	DebugRefreshCall("TestWidget", fyne.NewSize(100, 100))
}

func TestDebugMinSizeCall(t *testing.T) {
	// Should not panic
	DebugMinSizeCall("TestWidget", fyne.NewSize(50, 50))
}

func TestDebugWindowResize(t *testing.T) {
	// Should not panic
	DebugWindowResize(fyne.NewSize(800, 600))
	DebugWindowResize(fyne.NewSize(1024, 768))
}

func TestDebugInteraction(t *testing.T) {
	// Should not panic
	DebugInteraction("button clicked")
}

func TestResetLayoutCounts(t *testing.T) {
	// Add some counts
	DebugLayoutCall("Widget1", fyne.NewSize(100, 100))
	DebugLayoutCall("Widget2", fyne.NewSize(100, 100))

	// Reset
	ResetLayoutCounts()

	// Counters should be cleared (no way to verify directly, but shouldn't panic)
	DebugLayoutCall("Widget1", fyne.NewSize(100, 100))
}

func TestIsStartupStabilizing(t *testing.T) {
	// This function depends on time.Since(startupTime)
	// Just verify it doesn't panic and returns a bool
	result := IsStartupStabilizing()
	_ = result // Result depends on timing
}

func TestShouldSuppressResize(t *testing.T) {
	tests := []struct {
		name    string
		oldSize fyne.Size
		newSize fyne.Size
	}{
		{
			name:    "no change",
			oldSize: fyne.NewSize(800, 600),
			newSize: fyne.NewSize(800, 600),
		},
		{
			name:    "small change",
			oldSize: fyne.NewSize(800, 600),
			newSize: fyne.NewSize(801, 601),
		},
		{
			name:    "large change",
			oldSize: fyne.NewSize(800, 600),
			newSize: fyne.NewSize(1024, 768),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = ShouldSuppressResize(tt.oldSize, tt.newSize)
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input  float32
		expect float32
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.0, 0.0},
		{-100.5, 100.5},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expect {
			t.Errorf("abs(%v) = %v, expected %v", tt.input, result, tt.expect)
		}
	}
}

func TestWrapSelectCallback(t *testing.T) {
	var called bool
	var receivedValue string

	original := func(value string) {
		called = true
		receivedValue = value
	}

	wrapped := WrapSelectCallback("TestSelect", original)
	wrapped("option1")

	if !called {
		t.Error("original callback should be called")
	}
	if receivedValue != "option1" {
		t.Errorf("expected value 'option1', got '%s'", receivedValue)
	}
}

func TestWrapSelectCallback_NilOriginal(t *testing.T) {
	wrapped := WrapSelectCallback("TestSelect", nil)
	// Should not panic
	wrapped("option1")
}

func TestWrapButtonCallback(t *testing.T) {
	var called bool

	original := func() {
		called = true
	}

	wrapped := WrapButtonCallback("TestButton", original)
	wrapped()

	if !called {
		t.Error("original callback should be called")
	}
}

func TestWrapButtonCallback_NilOriginal(t *testing.T) {
	wrapped := WrapButtonCallback("TestButton", nil)
	// Should not panic
	wrapped()
}

func TestWrapSliderCallback(t *testing.T) {
	var called bool
	var receivedValue float64

	original := func(value float64) {
		called = true
		receivedValue = value
	}

	wrapped := WrapSliderCallback("TestSlider", original)
	wrapped(0.75)

	if !called {
		t.Error("original callback should be called")
	}
	if receivedValue != 0.75 {
		t.Errorf("expected value 0.75, got %v", receivedValue)
	}
}

func TestWrapSliderCallback_NilOriginal(t *testing.T) {
	wrapped := WrapSliderCallback("TestSlider", nil)
	// Should not panic
	wrapped(0.5)
}

func TestGetShortStack(t *testing.T) {
	// When DebugResize is false, should return empty string
	oldDebugResize := DebugResize
	DebugResize = false
	result := getShortStack()
	if result != "" {
		t.Error("getShortStack should return empty when DebugResize is false")
	}
	DebugResize = oldDebugResize
}
