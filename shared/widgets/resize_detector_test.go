package widgets

import (
	"sync/atomic"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestNewResizeDetector(t *testing.T) {
	var called bool
	detector := NewResizeDetector(func(size fyne.Size) {
		called = true
	})
	_ = called // Silence unused variable warning

	if detector == nil {
		t.Fatal("NewResizeDetector should return non-nil")
	}

	if detector.OnResize == nil {
		t.Error("OnResize callback should be set")
	}
}

func TestResizeDetector_CreateRenderer(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	detector := NewResizeDetector(nil)
	renderer := detector.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer should return non-nil")
	}

	// Check MinSize is zero (passive detector)
	minSize := renderer.MinSize()
	if minSize.Width != 0 || minSize.Height != 0 {
		t.Errorf("MinSize should be (0,0), got %v", minSize)
	}

	// Check Objects returns the rectangle
	objects := renderer.Objects()
	if len(objects) != 1 {
		t.Errorf("Objects should return 1 object, got %d", len(objects))
	}
}

func TestResizeDetector_CallbackOnResize(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	var callCount int32
	var lastSize fyne.Size

	detector := NewResizeDetector(func(size fyne.Size) {
		atomic.AddInt32(&callCount, 1)
		lastSize = size
	})

	renderer := detector.CreateRenderer()

	// First layout
	size1 := fyne.NewSize(100, 100)
	renderer.Layout(size1)

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 callback, got %d", callCount)
	}
	if lastSize != size1 {
		t.Errorf("expected size %v, got %v", size1, lastSize)
	}

	// Same size - no callback
	renderer.Layout(size1)
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("same size should not trigger callback, got %d calls", callCount)
	}

	// Different size - callback
	size2 := fyne.NewSize(200, 150)
	renderer.Layout(size2)
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("different size should trigger callback, got %d calls", callCount)
	}
	if lastSize != size2 {
		t.Errorf("expected size %v, got %v", size2, lastSize)
	}
}

func TestResizeDetector_NilCallback(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	detector := NewResizeDetector(nil)
	renderer := detector.CreateRenderer()

	// Should not panic with nil callback
	renderer.Layout(fyne.NewSize(100, 100))
	renderer.Refresh()
}

func TestGetBreakpoint(t *testing.T) {
	tests := []struct {
		width  float32
		expect Breakpoint
	}{
		{300, BreakpointSM},
		{576, BreakpointSM},
		{577, BreakpointMD},
		{768, BreakpointMD},
		{769, BreakpointLG},
		{1024, BreakpointLG},
		{1025, BreakpointXL},
		{1600, BreakpointXL},
		{1601, BreakpointXXL},
		{1920, BreakpointXXL},
	}

	for _, tt := range tests {
		result := GetBreakpoint(tt.width)
		if result != tt.expect {
			t.Errorf("GetBreakpoint(%v) = %v, expected %v", tt.width, result, tt.expect)
		}
	}
}

func TestBreakpointName(t *testing.T) {
	tests := []struct {
		bp     Breakpoint
		expect string
	}{
		{BreakpointSM, "SM (Mobile)"},
		{BreakpointMD, "MD (Tablet)"},
		{BreakpointLG, "LG (1-col)"},
		{BreakpointXL, "XL (2-col)"},
		{BreakpointXXL, "XXL (3-col)"},
		{Breakpoint(99), "Unknown"},
	}

	for _, tt := range tests {
		result := BreakpointName(tt.bp)
		if result != tt.expect {
			t.Errorf("BreakpointName(%v) = %q, expected %q", tt.bp, result, tt.expect)
		}
	}
}

func TestNewResponsiveDetector(t *testing.T) {
	var called bool
	detector := NewResponsiveDetector(func(bp Breakpoint, size fyne.Size) {
		called = true
	})
	_ = called // Silence unused variable warning

	if detector == nil {
		t.Fatal("NewResponsiveDetector should return non-nil")
	}

	// Default breakpoint should be XL
	if detector.currentBreakpoint != BreakpointXL {
		t.Errorf("default breakpoint should be XL, got %v", detector.currentBreakpoint)
	}
}

func TestResponsiveDetector_CurrentBreakpoint(t *testing.T) {
	detector := NewResponsiveDetector(nil)

	bp := detector.CurrentBreakpoint()
	if bp != BreakpointXL {
		t.Errorf("CurrentBreakpoint should return XL by default, got %v", bp)
	}
}

func TestResponsiveDetector_CreateRenderer(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	detector := NewResponsiveDetector(nil)
	renderer := detector.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer should return non-nil")
	}

	minSize := renderer.MinSize()
	if minSize.Width != 0 || minSize.Height != 0 {
		t.Errorf("MinSize should be (0,0), got %v", minSize)
	}
}

func TestResponsiveDetector_BreakpointChangeCallback(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	var callCount int32
	var lastBp Breakpoint

	detector := NewResponsiveDetector(func(bp Breakpoint, size fyne.Size) {
		atomic.AddInt32(&callCount, 1)
		lastBp = bp
	})

	renderer := detector.CreateRenderer()

	// First layout - initializes but doesn't fire callback
	renderer.Layout(fyne.NewSize(1200, 800)) // XL (1024-1600)
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("first layout should not fire callback, got %d calls", callCount)
	}

	// Same breakpoint - no callback
	renderer.Layout(fyne.NewSize(1300, 800)) // Still XL
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("same breakpoint should not fire callback, got %d calls", callCount)
	}

	// Different breakpoint - callback
	renderer.Layout(fyne.NewSize(500, 800)) // SM
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("breakpoint change should fire callback, got %d calls", callCount)
	}
	if lastBp != BreakpointSM {
		t.Errorf("expected BreakpointSM, got %v", lastBp)
	}

	// Another breakpoint change
	renderer.Layout(fyne.NewSize(700, 800)) // MD
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 callbacks, got %d", callCount)
	}
	if lastBp != BreakpointMD {
		t.Errorf("expected BreakpointMD, got %v", lastBp)
	}
}

func TestResponsiveDetector_NilCallback(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	detector := NewResponsiveDetector(nil)
	renderer := detector.CreateRenderer()

	// Should not panic with nil callback
	renderer.Layout(fyne.NewSize(1000, 800))
	renderer.Layout(fyne.NewSize(500, 800)) // Change breakpoint
	renderer.Refresh()
}

func TestBreakpointThresholds(t *testing.T) {
	// Verify threshold constants (M13: <1024=1col, 1024-1600=2col, >1600=3col)
	if ThresholdSM != 576 {
		t.Errorf("ThresholdSM should be 576, got %d", ThresholdSM)
	}
	if ThresholdMD != 768 {
		t.Errorf("ThresholdMD should be 768, got %d", ThresholdMD)
	}
	if ThresholdLG != 1024 {
		t.Errorf("ThresholdLG should be 1024, got %d", ThresholdLG)
	}
	if ThresholdXL != 1600 {
		t.Errorf("ThresholdXL should be 1600, got %d", ThresholdXL)
	}
}

func BenchmarkGetBreakpoint(b *testing.B) {
	widths := []float32{300, 600, 800, 1200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBreakpoint(widths[i%4])
	}
}
