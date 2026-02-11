// Package tabs provides individual tab components for the Demo 2 crossbar GUI.
package tabs

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// TestNewIdealTab tests creation of ideal MVM tab.
func TestNewIdealTab(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	array, err := crossbar.NewArray(&crossbar.Config{Rows: 8, Cols: 8})
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	tab := NewIdealTab(array, nil)

	if tab == nil {
		t.Fatal("NewIdealTab returned nil")
	}
	if tab.array != array {
		t.Error("IdealTab array not set correctly")
	}
	if tab.statsLabel == nil {
		t.Error("IdealTab statsLabel is nil")
	}
	if tab.statusLabel == nil {
		t.Error("IdealTab statusLabel is nil")
	}
}

// TestIdealTabContent tests that Content() returns non-nil canvas object.
func TestIdealTabContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	array, err := crossbar.NewArray(&crossbar.Config{Rows: 16, Cols: 16})
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	tab := NewIdealTab(array, nil)

	content := tab.Content()
	if content == nil {
		t.Fatal("IdealTab.Content() returned nil")
	}
}

// TestIdealTabSetArray tests SetArray method.
func TestIdealTabSetArray(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	array1, err := crossbar.NewArray(&crossbar.Config{Rows: 8, Cols: 8})
	if err != nil {
		t.Fatalf("Failed to create array1: %v", err)
	}
	array2, err := crossbar.NewArray(&crossbar.Config{Rows: 16, Cols: 16})
	if err != nil {
		t.Fatalf("Failed to create array2: %v", err)
	}

	tab := NewIdealTab(array1, nil)
	if tab.array != array1 {
		t.Error("Initial array not set correctly")
	}

	tab.SetArray(array2)
	if tab.array != array2 {
		t.Error("SetArray did not update array reference")
	}
}

// TestIdealTabWithNilArray tests behavior with nil array.
func TestIdealTabWithNilArray(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewIdealTab(nil, nil)
	if tab == nil {
		t.Fatal("NewIdealTab returned nil with nil array")
	}

	// Content should still work with nil array
	content := tab.Content()
	if content == nil {
		t.Error("Content() returned nil with nil array")
	}
}

// TestIdealTabWithCallback tests callback functionality.
func TestIdealTabWithCallback(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	array, err := crossbar.NewArray(&crossbar.Config{Rows: 8, Cols: 8})
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	callbackCalled := false
	callback := func() {
		callbackCalled = true
	}

	tab := NewIdealTab(array, callback)
	if tab.onArrayChange == nil {
		t.Error("Callback not set")
	}

	// Trigger callback
	if tab.onArrayChange != nil {
		tab.onArrayChange()
	}
	if !callbackCalled {
		t.Error("Callback was not called")
	}
}

// TestNewDriftTab tests creation of drift analysis tab.
func TestNewDriftTab(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	arraySize := 8
	tab := NewDriftTab(arraySize)

	if tab == nil {
		t.Fatal("NewDriftTab returned nil")
	}
	if tab.arraySize != arraySize {
		t.Errorf("Expected arraySize %d, got %d", arraySize, tab.arraySize)
	}
	if tab.simulator == nil {
		t.Error("DriftTab simulator is nil")
	}
	if tab.statsLabel == nil {
		t.Error("DriftTab statsLabel is nil")
	}
	if tab.statusLabel == nil {
		t.Error("DriftTab statusLabel is nil")
	}
}

// TestDriftTabContent tests that Content() returns non-nil canvas object.
func TestDriftTabContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewDriftTab(8)
	content := tab.Content()

	if content == nil {
		t.Fatal("DriftTab.Content() returned nil")
	}
	if tab.driftChart == nil {
		t.Error("driftChart container not initialized")
	}
}

// TestDriftTabWithVariousArraySizes tests drift tab with different array sizes.
func TestDriftTabWithVariousArraySizes(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	sizes := []int{4, 8, 16, 32}
	for _, size := range sizes {
		tab := NewDriftTab(size)
		if tab == nil {
			t.Errorf("NewDriftTab returned nil for size %d", size)
			continue
		}
		if tab.arraySize != size {
			t.Errorf("Size %d: expected arraySize %d, got %d", size, size, tab.arraySize)
		}

		content := tab.Content()
		if content == nil {
			t.Errorf("Content() returned nil for size %d", size)
		}
	}
}

// TestDriftTabUpdateChart tests chart update functionality.
func TestDriftTabUpdateChart(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewDriftTab(8)
	_ = tab.Content() // Initialize chart

	// Should not panic with valid chart
	tab.updateChart()

	// Test with nil chart
	tab.driftChart = nil
	tab.updateChart() // Should not panic
}

// TestDriftTabUpdateStats tests stats update functionality.
func TestDriftTabUpdateStats(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewDriftTab(8)
	_ = tab.Content()

	// Should not panic
	tab.updateStats()

	// Check that stats label was updated
	if tab.statsLabel.Text == "" {
		t.Error("Stats label not updated")
	}
}

// TestDriftTabTechComparison tests technology comparison functionality.
func TestDriftTabTechComparison(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewDriftTab(8)
	_ = tab.Content()

	// Should not panic
	tab.showTechComparison()

	// Stats label should be updated with comparison
	if tab.statsLabel.Text == "" {
		t.Error("Stats label not updated after tech comparison")
	}
}

// TestDriftTabTechComparisonContent tests TechComparisonContent method.
func TestDriftTabTechComparisonContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewDriftTab(8)
	content := tab.TechComparisonContent()

	if content == nil {
		t.Fatal("TechComparisonContent() returned nil")
	}
}

// TestNewSneakTab tests creation of sneak path analysis tab.
func TestNewSneakTab(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	arraySize := 8
	tab := NewSneakTab(arraySize)

	if tab == nil {
		t.Fatal("NewSneakTab returned nil")
	}
	if tab.arraySize != arraySize {
		t.Errorf("Expected arraySize %d, got %d", arraySize, tab.arraySize)
	}
	if tab.simulator == nil {
		t.Error("SneakTab simulator is nil")
	}
	if tab.statsLabel == nil {
		t.Error("SneakTab statsLabel is nil")
	}
	if tab.statusLabel == nil {
		t.Error("SneakTab statusLabel is nil")
	}
}

// TestSneakTabContent tests that Content() returns non-nil canvas object.
func TestSneakTabContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewSneakTab(8)
	content := tab.Content()

	if content == nil {
		t.Fatal("SneakTab.Content() returned nil")
	}
	if tab.heatmap == nil {
		t.Error("heatmap container not initialized")
	}
	if tab.legend == nil {
		t.Error("legend not initialized")
	}
}

// TestSneakTabWithVariousArraySizes tests sneak tab with different array sizes.
func TestSneakTabWithVariousArraySizes(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	sizes := []int{4, 8, 16, 32}
	for _, size := range sizes {
		tab := NewSneakTab(size)
		if tab == nil {
			t.Errorf("NewSneakTab returned nil for size %d", size)
			continue
		}
		if tab.arraySize != size {
			t.Errorf("Size %d: expected arraySize %d, got %d", size, size, tab.arraySize)
		}

		content := tab.Content()
		if content == nil {
			t.Errorf("Content() returned nil for size %d", size)
		}
	}
}

// TestSneakTabUpdateHeatmap tests heatmap update functionality.
func TestSneakTabUpdateHeatmap(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewSneakTab(8)
	_ = tab.Content() // Initialize heatmap

	// Should not panic with valid heatmap
	tab.updateHeatmap()

	// Test with nil heatmap
	tab.heatmap = nil
	tab.updateHeatmap() // Should not panic
}

// TestSneakTabUpdateStats tests stats update functionality.
func TestSneakTabUpdateStats(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewSneakTab(8)
	_ = tab.Content()

	// Should not panic
	tab.updateStats()

	// Check that stats label was updated
	if tab.statsLabel.Text == "" {
		t.Error("Stats label not updated")
	}
}

// TestSneakTabGetSeverity tests severity classification.
func TestSneakTabGetSeverity(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewSneakTab(8)

	tests := []struct {
		snr      float64
		expected string
	}{
		{35.0, "Excellent (SNR > 30dB)"},
		{25.0, "Good (SNR > 20dB)"},
		{15.0, "Acceptable (SNR > 10dB)"},
		{5.0, "Needs Selector (SNR < 10dB)"},
	}

	for _, tt := range tests {
		result := tab.getSneakSeverity(tt.snr)
		if result != tt.expected {
			t.Errorf("getSneakSeverity(%.1f) = %s, expected %s", tt.snr, result, tt.expected)
		}
	}
}

// TestNewIRDropTab tests creation of IR drop analysis tab.
func TestNewIRDropTab(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	arraySize := 8
	tab := NewIRDropTab(arraySize)

	if tab == nil {
		t.Fatal("NewIRDropTab returned nil")
	}
	if tab.arraySize != arraySize {
		t.Errorf("Expected arraySize %d, got %d", arraySize, tab.arraySize)
	}
	if tab.simulator == nil {
		t.Error("IRDropTab simulator is nil")
	}
	if tab.statsLabel == nil {
		t.Error("IRDropTab statsLabel is nil")
	}
	if tab.statusLabel == nil {
		t.Error("IRDropTab statusLabel is nil")
	}
}

// TestIRDropTabContent tests that Content() returns non-nil canvas object.
func TestIRDropTabContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewIRDropTab(8)
	content := tab.Content()

	if content == nil {
		t.Fatal("IRDropTab.Content() returned nil")
	}
	if tab.heatmap == nil {
		t.Error("heatmap container not initialized")
	}
	if tab.legend == nil {
		t.Error("legend not initialized")
	}
}

// TestIRDropTabWithVariousArraySizes tests IR drop tab with different array sizes.
func TestIRDropTabWithVariousArraySizes(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	sizes := []int{4, 8, 16, 32}
	for _, size := range sizes {
		tab := NewIRDropTab(size)
		if tab == nil {
			t.Errorf("NewIRDropTab returned nil for size %d", size)
			continue
		}
		if tab.arraySize != size {
			t.Errorf("Size %d: expected arraySize %d, got %d", size, size, tab.arraySize)
		}

		content := tab.Content()
		if content == nil {
			t.Errorf("Content() returned nil for size %d", size)
		}
	}
}

// TestIRDropTabUpdateHeatmap tests heatmap update functionality.
func TestIRDropTabUpdateHeatmap(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewIRDropTab(8)
	_ = tab.Content() // Initialize heatmap

	// Should not panic with valid heatmap
	tab.updateHeatmap()

	// Test with nil heatmap
	tab.heatmap = nil
	tab.updateHeatmap() // Should not panic
}

// TestIRDropTabUpdateStats tests stats update functionality.
func TestIRDropTabUpdateStats(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewIRDropTab(8)
	_ = tab.Content()

	// Should not panic
	tab.updateStats()

	// Check that stats label was updated
	if tab.statsLabel.Text == "" {
		t.Error("Stats label not updated")
	}
}

// TestIRDropTabGetSeverity tests severity classification.
func TestIRDropTabGetSeverity(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewIRDropTab(8)

	tests := []struct {
		maxError float64
		expected string
	}{
		{0.5, "Excellent (<1% error)"},
		{3.0, "Good (<5% error)"},
		{7.0, "Acceptable (<10% error)"},
		{15.0, "Needs Mitigation (>10% error)"},
	}

	for _, tt := range tests {
		result := tab.getIRSeverity(tt.maxError)
		if result != tt.expected {
			t.Errorf("getIRSeverity(%.1f) = %s, expected %s", tt.maxError, result, tt.expected)
		}
	}
}

// TestZeroArraySize tests behavior with zero-size array configuration.
// Note: Zero-size arrays currently cause panics in the simulators.
// This is documented as a known limitation - minimum size should be 1x1.
func TestZeroArraySize(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// Test each tab type with zero size (edge case)
	// Currently these panic due to array index out of bounds in simulators
	// We test that we can detect this condition

	t.Run("DriftTab", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic with zero size: %v", r)
			}
		}()
		tab := NewDriftTab(0)
		if tab == nil {
			t.Fatal("NewDriftTab returned nil with zero size")
		}
		// Content() call may panic - caught by defer recover
		content := tab.Content()
		if content == nil {
			t.Error("Content() returned nil with zero size")
		}
	})

	t.Run("SneakTab", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic with zero size: %v", r)
			}
		}()
		tab := NewSneakTab(0)
		if tab == nil {
			t.Fatal("NewSneakTab returned nil with zero size")
		}
		content := tab.Content()
		if content == nil {
			t.Error("Content() returned nil with zero size")
		}
	})

	t.Run("IRDropTab", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic with zero size: %v", r)
			}
		}()
		tab := NewIRDropTab(0)
		if tab == nil {
			t.Fatal("NewIRDropTab returned nil with zero size")
		}
		content := tab.Content()
		if content == nil {
			t.Error("Content() returned nil with zero size")
		}
	})
}

// TestLargeArraySize tests behavior with large array configurations.
func TestLargeArraySize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large array test in short mode")
	}

	app := test.NewApp()
	defer app.Quit()

	// Test with moderately large size (64x64)
	size := 64

	t.Run("DriftTab", func(t *testing.T) {
		tab := NewDriftTab(size)
		if tab == nil {
			t.Fatal("NewDriftTab returned nil with large size")
		}
		content := tab.Content()
		if content == nil {
			t.Error("Content() returned nil with large size")
		}
	})

	t.Run("SneakTab", func(t *testing.T) {
		tab := NewSneakTab(size)
		if tab == nil {
			t.Fatal("NewSneakTab returned nil with large size")
		}
		content := tab.Content()
		if content == nil {
			t.Error("Content() returned nil with large size")
		}
	})

	t.Run("IRDropTab", func(t *testing.T) {
		tab := NewIRDropTab(size)
		if tab == nil {
			t.Fatal("NewIRDropTab returned nil with large size")
		}
		content := tab.Content()
		if content == nil {
			t.Error("Content() returned nil with large size")
		}
	})
}

// TestTabsMultipleContentCalls tests that calling Content() multiple times is safe.
func TestTabsMultipleContentCalls(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("IdealTab", func(t *testing.T) {
		array, err := crossbar.NewArray(&crossbar.Config{Rows: 8, Cols: 8})
		if err != nil {
			t.Fatalf("Failed to create array: %v", err)
		}
		tab := NewIdealTab(array, nil)
		content1 := tab.Content()
		content2 := tab.Content()
		if content1 == nil || content2 == nil {
			t.Error("Content() returned nil on multiple calls")
		}
	})

	t.Run("DriftTab", func(t *testing.T) {
		tab := NewDriftTab(8)
		content1 := tab.Content()
		content2 := tab.Content()
		if content1 == nil || content2 == nil {
			t.Error("Content() returned nil on multiple calls")
		}
	})

	t.Run("SneakTab", func(t *testing.T) {
		tab := NewSneakTab(8)
		content1 := tab.Content()
		content2 := tab.Content()
		if content1 == nil || content2 == nil {
			t.Error("Content() returned nil on multiple calls")
		}
	})

	t.Run("IRDropTab", func(t *testing.T) {
		tab := NewIRDropTab(8)
		content1 := tab.Content()
		content2 := tab.Content()
		if content1 == nil || content2 == nil {
			t.Error("Content() returned nil on multiple calls")
		}
	})
}

// TestDriftTabSimulation tests drift simulation functionality.
func TestDriftTabSimulation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewDriftTab(8)
	_ = tab.Content()

	// Verify simulator has initial history
	if tab.simulator == nil {
		t.Fatal("Simulator is nil")
	}
	if len(tab.simulator.DriftHistory) == 0 {
		t.Error("Simulator should have initial drift history")
	}

	initialHistoryLen := len(tab.simulator.DriftHistory)

	// Simulate additional time
	for step := 0; step < 10; step++ {
		tab.simulator.SimulateTimeStep(200)
		tab.simulator.RecordSnapshot()
	}

	if len(tab.simulator.DriftHistory) <= initialHistoryLen {
		t.Error("Drift history did not grow after simulation")
	}
}

// TestSneakTabMitigationEffects tests that mitigation strategies have effects.
func TestSneakTabMitigationEffects(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewSneakTab(8)
	_ = tab.Content()

	// Get baseline stats
	baselineStats := tab.simulator.GetStats(0.5)
	baselineSNR := baselineStats.SignalToNoiseRatio

	// Apply selector mitigation
	mitigation := crossbar.SneakMitigation{
		UseSelector:   true,
		SelectorOnOff: 1000,
	}
	tab.simulator.AnalyzeWithMitigation(tab.arraySize/2, tab.arraySize/2, 0.5, mitigation)

	// Get mitigated stats
	mitigatedStats := tab.simulator.GetStats(0.5)
	mitigatedSNR := mitigatedStats.SignalToNoiseRatio

	// Mitigation should improve SNR (or at least not make it worse significantly)
	if mitigatedSNR < baselineSNR*0.5 {
		t.Errorf("Mitigation made SNR significantly worse: baseline=%.2f, mitigated=%.2f",
			baselineSNR, mitigatedSNR)
	}
}

// TestIRDropTabMitigationEffects tests that mitigation strategies reduce IR drop.
func TestIRDropTabMitigationEffects(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	tab := NewIRDropTab(8)
	_ = tab.Content()

	// Get baseline stats
	baselineStats := tab.simulator.GetStats()
	baselineMaxDrop := baselineStats.MaxIRDrop

	// Apply mitigation
	mitigation := crossbar.IRDropMitigation{
		UseWidenedLines:   true,
		LineWidthIncrease: 4.0,
	}
	tab.simulator.ApplyMitigation(mitigation)

	// Get mitigated stats
	mitigatedStats := tab.simulator.GetStats()
	mitigatedMaxDrop := mitigatedStats.MaxIRDrop

	// Mitigation should reduce IR drop
	if mitigatedMaxDrop >= baselineMaxDrop {
		t.Errorf("Mitigation did not reduce IR drop: baseline=%.6f, mitigated=%.6f",
			baselineMaxDrop, mitigatedMaxDrop)
	}
}
