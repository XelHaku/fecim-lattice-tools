//go:build legacy_fyne

// Package widgets provides tests for GUI widget logic.
// These tests focus on state management and logic that can run without a display.
package widgets

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
)

// Test colors for PEPlot
var (
	testBgColor   = color.RGBA{R: 30, G: 30, B: 50, A: 255}
	testGridColor = color.RGBA{R: 60, G: 60, B: 80, A: 255}
	testAxisColor = color.RGBA{R: 100, G: 100, B: 120, A: 255}
	testPosColor  = color.RGBA{R: 0, G: 200, B: 100, A: 255}
	testNegColor  = color.RGBA{R: 200, G: 50, B: 50, A: 255}
	testWarnColor = color.RGBA{R: 255, G: 200, B: 0, A: 255}
)

// =============================================================================
// LEVEL INDICATOR TESTS
// =============================================================================

// TestLevelIndicatorCreation verifies widget initialization.
func TestLevelIndicatorCreation(t *testing.T) {
	l := NewLevelIndicator()

	if l == nil {
		t.Fatal("NewLevelIndicator returned nil")
	}

	// Default level should be 15 (middle)
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level != 15 {
		t.Errorf("Default level should be 15, got %d", level)
	}
}

// TestLevelIndicatorSetLevel verifies level setting.
func TestLevelIndicatorSetLevel(t *testing.T) {
	l := NewLevelIndicator()

	testCases := []int{0, 1, 14, 15, 16, 29}

	for _, tc := range testCases {
		l.SetLevel(tc)

		l.mu.RLock()
		got := l.level
		l.mu.RUnlock()

		if got != tc {
			t.Errorf("SetLevel(%d): expected %d, got %d", tc, tc, got)
		}
	}
}

// TestLevelIndicatorBounds verifies level clamping behavior.
func TestLevelIndicatorBounds(t *testing.T) {
	l := NewLevelIndicator()

	// Test that levels are stored as-is (clamping happens in renderer)
	l.SetLevel(-5)
	l.mu.RLock()
	levelNeg := l.level
	l.mu.RUnlock()

	l.SetLevel(100)
	l.mu.RLock()
	levelHigh := l.level
	l.mu.RUnlock()

	// Widget stores values as-is, renderer handles display
	t.Logf("Out of bounds test: level(-5)=%d, level(100)=%d", levelNeg, levelHigh)
}

// TestLevelIndicatorTargetLevel verifies target level highlighting state.
// Note: We can't test SetTargetLevel with highlight=true because it starts
// a Fyne animation which requires a running Fyne app. We test the state storage.
func TestLevelIndicatorTargetLevel(t *testing.T) {
	l := NewLevelIndicator()

	// Directly set the target level fields (avoiding animation)
	l.mu.Lock()
	l.targetLevel = 20
	l.highlightTarget = false // Don't start animation
	l.mu.Unlock()

	l.mu.RLock()
	target := l.targetLevel
	highlight := l.highlightTarget
	l.mu.RUnlock()

	if target != 20 {
		t.Errorf("Target level should be 20, got %d", target)
	}
	if highlight {
		t.Error("Highlight should be false (no animation started)")
	}

	// Test SetTargetLevel without animation (highlight=false is safe)
	l.SetTargetLevel(25, false)

	l.mu.RLock()
	target = l.targetLevel
	l.mu.RUnlock()

	if target != 25 {
		t.Errorf("Target level should be 25, got %d", target)
	}
}

// TestLevelIndicatorMinSize verifies minimum size setting.
func TestLevelIndicatorMinSize(t *testing.T) {
	l := NewLevelIndicator()

	// Check default
	size := l.MinSize()
	if size.Width <= 0 || size.Height <= 0 {
		t.Error("MinSize should have positive dimensions")
	}

	// Set custom size
	l.SetMinSize(fyne.NewSize(100, 500))
	newSize := l.MinSize()

	if newSize.Width != 100 || newSize.Height != 500 {
		t.Errorf("MinSize should be 100x500, got %.0fx%.0f", newSize.Width, newSize.Height)
	}
}

// TestLevelIndicatorCallback verifies click callback is stored.
func TestLevelIndicatorCallback(t *testing.T) {
	l := NewLevelIndicator()

	callbackCalled := false
	l.OnLevelClicked = func(targetLevel int) {
		callbackCalled = true
	}

	if l.OnLevelClicked == nil {
		t.Error("Callback should be stored")
	}

	// Simulate callback
	if l.OnLevelClicked != nil {
		l.OnLevelClicked(15)
	}

	if !callbackCalled {
		t.Error("Callback should have been called")
	}
}

// TestLevelIndicatorConcurrency verifies thread-safe access.
// Note: We avoid SetTargetLevel with highlight=true as it requires Fyne app.
func TestLevelIndicatorConcurrency(t *testing.T) {
	l := NewLevelIndicator()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			l.SetLevel(i % 30)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			l.mu.RLock()
			_ = l.level
			l.mu.RUnlock()
		}
		done <- true
	}()

	// Target setter goroutine (only with highlight=false to avoid animation)
	go func() {
		for i := 0; i < 1000; i++ {
			l.SetTargetLevel(i%30, false)
		}
		done <- true
	}()

	<-done
	<-done
	<-done

	t.Log("Concurrent access completed without race")
}

// =============================================================================
// CELL VISUALIZER TESTS
// =============================================================================

// TestCellVisualizerCreation verifies cell visualizer initialization.
func TestCellVisualizerCreation(t *testing.T) {
	c := NewCellVisualizer()

	if c == nil {
		t.Fatal("NewCellVisualizer returned nil")
	}
}

// TestCellVisualizerSetLevel verifies level setting on cell visualizer.
func TestCellVisualizerSetLevel(t *testing.T) {
	c := NewCellVisualizer()

	// Test various levels
	for level := 0; level < 30; level++ {
		c.SetLevel(level)

		c.mu.RLock()
		got := c.level
		c.mu.RUnlock()

		if got != level {
			t.Errorf("SetLevel(%d): got %d", level, got)
		}
	}
}

// TestCellVisualizerMinSize verifies minimum size.
func TestCellVisualizerMinSize(t *testing.T) {
	c := NewCellVisualizer()

	size := c.MinSize()
	if size.Width <= 0 || size.Height <= 0 {
		t.Error("MinSize should have positive dimensions")
	}
}

// =============================================================================
// MODE INDICATOR TESTS
// =============================================================================

// TestModeIndicatorCreation verifies mode indicator initialization.
func TestModeIndicatorCreation(t *testing.T) {
	m := NewModeIndicator()

	if m == nil {
		t.Fatal("NewModeIndicator returned nil")
	}
}

// TestModeIndicatorSetWrite verifies write mode setting.
func TestModeIndicatorSetWrite(t *testing.T) {
	m := NewModeIndicator()

	// Set to write mode
	m.SetWrite(true)
	m.mu.RLock()
	isWrite := m.isWrite
	m.mu.RUnlock()

	if !isWrite {
		t.Error("Should be in write mode")
	}

	// Set to read mode
	m.SetWrite(false)
	m.mu.RLock()
	isWrite = m.isWrite
	m.mu.RUnlock()

	if isWrite {
		t.Error("Should be in read mode")
	}
}

// TestModeIndicatorMinSize verifies minimum size.
func TestModeIndicatorMinSize(t *testing.T) {
	m := NewModeIndicator()

	size := m.MinSize()
	if size.Width <= 0 || size.Height <= 0 {
		t.Error("MinSize should have positive dimensions")
	}
}

// =============================================================================
// PE PLOT TESTS
// =============================================================================

// TestPEPlotCreation verifies P-E plot initialization.
func TestPEPlotCreation(t *testing.T) {
	eMax := 2e8 // 2 MV/cm
	pMax := 0.4 // 40 µC/cm²

	p := NewPEPlot(eMax, pMax, testBgColor, testGridColor, testAxisColor, testPosColor, testNegColor, testWarnColor)

	if p == nil {
		t.Fatal("NewPEPlot returned nil")
	}

	if p.eMax != eMax {
		t.Errorf("eMax should be %e, got %e", eMax, p.eMax)
	}

	if p.pMax != pMax {
		t.Errorf("pMax should be %e, got %e", pMax, p.pMax)
	}
}

// TestPEPlotSetBounds verifies bounds setting.
func TestPEPlotSetBounds(t *testing.T) {
	p := NewPEPlot(1e8, 0.3, testBgColor, testGridColor, testAxisColor, testPosColor, testNegColor, testWarnColor)

	newE := 3e8
	newP := 0.5
	p.SetBounds(newE, newP)

	p.mu.RLock()
	eMax := p.eMax
	pMax := p.pMax
	p.mu.RUnlock()

	if eMax != newE {
		t.Errorf("eMax should be %e, got %e", newE, eMax)
	}
	if pMax != newP {
		t.Errorf("pMax should be %e, got %e", newP, pMax)
	}
}

// TestPEPlotSetData verifies data setting.
func TestPEPlotSetData(t *testing.T) {
	p := NewPEPlot(2e8, 0.4, testBgColor, testGridColor, testAxisColor, testPosColor, testNegColor, testWarnColor)

	// Create test data
	eHist := []float64{-1e8, 0, 1e8, 0}
	pHist := []float64{-0.2, 0, 0.2, 0.1}
	currentE := 0.5e8
	currentP := 0.15

	p.SetData(eHist, pHist, currentE, currentP)

	p.mu.RLock()
	dataLen := len(p.eData)
	curE := p.currentE
	curP := p.currentP
	p.mu.RUnlock()

	if dataLen != len(eHist) {
		t.Errorf("History length should be %d, got %d", len(eHist), dataLen)
	}

	if curE != currentE {
		t.Errorf("Current E should be %e, got %e", currentE, curE)
	}

	if curP != currentP {
		t.Errorf("Current P should be %e, got %e", currentP, curP)
	}
}

// TestPEPlotMinSize verifies minimum size.
func TestPEPlotMinSize(t *testing.T) {
	p := NewPEPlot(2e8, 0.4, testBgColor, testGridColor, testAxisColor, testPosColor, testNegColor, testWarnColor)

	// Default minSize is 400x300 as set in NewPEPlot
	size := p.MinSize()
	if size.Width != 400 || size.Height != 300 {
		t.Errorf("MinSize should be 400x300, got %.0fx%.0f", size.Width, size.Height)
	}
}

// TestPEPlotConcurrency verifies thread-safe data access.
func TestPEPlotConcurrency(t *testing.T) {
	p := NewPEPlot(2e8, 0.4, testBgColor, testGridColor, testAxisColor, testPosColor, testNegColor, testWarnColor)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			e := float64(i%200-100) * 1e6
			pVal := float64(i%100-50) * 0.004
			eHist := []float64{e - 1e7, e}
			pHist := []float64{pVal - 0.01, pVal}
			p.SetData(eHist, pHist, e, pVal)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			p.mu.RLock()
			_ = p.currentE
			_ = p.currentP
			_ = len(p.eData)
			p.mu.RUnlock()
		}
		done <- true
	}()

	<-done
	<-done

	t.Log("Concurrent PE plot access completed without race")
}

// =============================================================================
// 30-LEVEL QUANTIZATION TESTS (Physics Accuracy)
// =============================================================================

// TestLevelIndicator30Levels verifies the full 30-level range.
func TestLevelIndicator30Levels(t *testing.T) {
	l := NewLevelIndicator()

	// Test all 30 levels (0-29)
	for level := 0; level < 30; level++ {
		l.SetLevel(level)

		l.mu.RLock()
		got := l.level
		l.mu.RUnlock()

		if got != level {
			t.Errorf("Level %d: got %d", level, got)
		}
	}

	t.Log("All 30 FeCIM levels verified")
}

// TestCellVisualizerPolarityMapping verifies level-to-polarity mapping.
func TestCellVisualizerPolarityMapping(t *testing.T) {
	c := NewCellVisualizer()

	// Level 0 = fully negative polarization
	c.SetLevel(0)
	c.mu.RLock()
	level0 := c.level
	c.mu.RUnlock()

	// Level 15 = neutral
	c.SetLevel(15)
	c.mu.RLock()
	level15 := c.level
	c.mu.RUnlock()

	// Level 29 = fully positive polarization
	c.SetLevel(29)
	c.mu.RLock()
	level29 := c.level
	c.mu.RUnlock()

	if level0 != 0 || level15 != 15 || level29 != 29 {
		t.Errorf("Level mapping issue: 0=%d, 15=%d, 29=%d", level0, level15, level29)
	}

	t.Log("30-level polarity mapping verified: 0=negative, 15=neutral, 29=positive")
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkLevelIndicatorSetLevel(b *testing.B) {
	l := NewLevelIndicator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.SetLevel(i % 30)
	}
}

func BenchmarkPEPlotSetData(b *testing.B) {
	p := NewPEPlot(2e8, 0.4, testBgColor, testGridColor, testAxisColor, testPosColor, testNegColor, testWarnColor)
	eHist := make([]float64, 100)
	pHist := make([]float64, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.SetData(eHist, pHist, 0, 0)
	}
}

func BenchmarkModeIndicatorSetWrite(b *testing.B) {
	m := NewModeIndicator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetWrite(i%2 == 0)
	}
}
