// Package gui provides Fyne-based GUI components for crossbar visualization.
// analysis.go contains IR drop and sneak path analysis functions.
package gui

import (
	"fmt"
	"math"
	"math/rand"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// runIRDropAnalysis updates IR drop heatmap silently (no tab switch).
func (ca *CrossbarApp) runIRDropAnalysis() {
	// Protected read of lastInput
	ca.stateMu.RLock()
	input := ca.lastInput
	ca.stateMu.RUnlock()

	if input == nil || len(input) != ca.config.Cols {
		input = make([]float64, ca.config.Cols)
		for i := range input {
			input[i] = rand.Float64()
		}
	}

	params := crossbar.DefaultWireParams()
	analysis := ca.array.AnalyzeIRDrop(input, params)

	// Protected write to lastIRDropAnalysis
	ca.stateMu.Lock()
	ca.lastIRDropAnalysis = analysis
	ca.stateMu.Unlock()

	irMap := analysis.GetIRDropMap()
	fyne.Do(func() {
		ca.irDropHeatmap.SetData(irMap)
		ca.irDropHeatmap.SetSelection(analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])
	})

	// Add badge to IR Drop tab (C2 accessibility fix - discoverability)
	ca.setTabBadge("IR Drop")
}

// runSneakPathAnalysis updates sneak path heatmap silently (no tab switch).
func (ca *CrossbarApp) runSneakPathAnalysis() {
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	analysis := ca.array.AnalyzeSneakPaths(selectedRow, selectedCol)

	// Protected write to lastSneakAnalysis
	ca.stateMu.Lock()
	ca.lastSneakAnalysis = analysis
	ca.stateMu.Unlock()

	sneakMap := analysis.GetSneakMap()

	// Apply sqrt transformation for better visibility
	for i := range sneakMap {
		for j := range sneakMap[i] {
			sneakMap[i][j] = math.Sqrt(sneakMap[i][j])
		}
	}

	fyne.Do(func() {
		ca.sneakPathHeatmap.SetData(sneakMap)
		ca.sneakPathHeatmap.SetSelection(selectedRow, selectedCol)
	})

	// Add badge to Sneak Paths tab (C2 accessibility fix - discoverability)
	ca.setTabBadge("Sneak Paths")
}

// analyzeIRDrop performs IR drop analysis.
func (ca *CrossbarApp) analyzeIRDrop() {
	// Update mode and educational panel
	ca.modeIndicator.SetMode(DemoModeIRDrop)
	ca.setEducationalContent("Non-Ideality: IR Drop", "IR DROP ANALYSIS\n\nWire resistance causes\nvoltage drop along lines.\n\nCells far from drivers\nsee reduced voltage.\n\nThis affects accuracy:\n• Worst at corners\n• Mitigate with drivers")
	ca.updateStatus("IR DROP | Computing voltage distribution across array (wire resistance model)...")

	// Protected read of lastInput
	ca.stateMu.RLock()
	input := ca.lastInput
	ca.stateMu.RUnlock()

	if input == nil || len(input) != ca.config.Cols {
		input = make([]float64, ca.config.Cols)
		for i := range input {
			input[i] = rand.Float64()
		}
		// Protected write to lastInput
		ca.stateMu.Lock()
		ca.lastInput = input
		ca.stateMu.Unlock()
	}

	// Analyze IR drop
	params := crossbar.DefaultWireParams()
	analysis := ca.array.AnalyzeIRDrop(input, params)

	// Update IR drop heatmap
	irMap := analysis.GetIRDropMap()
	debug.Printf("IR Drop map size: %dx%d, MaxIRDrop: %.6f", len(irMap), len(irMap[0]), analysis.MaxIRDrop)
	// Sample some values to see the range
	if len(irMap) > 0 && len(irMap[0]) > 0 {
		debug.Printf("IR Drop sample values: [0,0]=%.4f, [mid,mid]=%.4f", irMap[0][0], irMap[len(irMap)/2][len(irMap[0])/2])
	}
	fyne.Do(func() {
		ca.irDropHeatmap.SetData(irMap)
		ca.irDropHeatmap.SetSelection(analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])
	})

	// Update stats
	ca.statsLabel.SetText(fmt.Sprintf(
		"IR Drop Analysis\n"+
			"Max IR Drop: %.2f%%\n"+
			"Avg IR Drop: %.2f%%\n"+
			"Variance: %.6f\n"+
			"Worst Cell: [%d, %d]\n\n"+
			"Wire Parameters:\n"+
			"Word Line R: %.1f Ω/cell\n"+
			"Bit Line R: %.1f Ω/cell\n"+
			"Contact R: %.1f Ω",
		analysis.MaxIRDrop*100,
		analysis.AvgIRDrop*100,
		analysis.IRDropVariance,
		analysis.WorstCaseCell[0], analysis.WorstCaseCell[1],
		params.RwordLine, params.RbitLine, params.Rcontact,
	))

	// Update key stat and log
	ca.setKeyStatValue(fmt.Sprintf("Max: %.1f%% drop", analysis.MaxIRDrop*100))

	ca.updateStatus(fmt.Sprintf("IR DROP | Complete: Max voltage drop %.2f%% at corner cell [%d,%d]. Check heatmap!",
		analysis.MaxIRDrop*100, analysis.WorstCaseCell[0], analysis.WorstCaseCell[1]))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// analyzeSneakPaths performs sneak path analysis.
func (ca *CrossbarApp) analyzeSneakPaths() {
	// Update mode and educational panel
	ca.modeIndicator.SetMode(DemoModeSneakPath)
	ca.setEducationalContent("Non-Ideality: Sneak Paths", "SNEAK PATH ANALYSIS\n\nCurrent can flow through\nunintended paths in passive\ncrossbar arrays.\n\nMitigation strategies:\n• Selector devices\n• 1T1R architecture\n• Threshold switching")
	ca.updateStatus("SNEAK | Computing parasitic current paths for center cell...")

	// Select center cell
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	// Analyze sneak paths
	analysis := ca.array.AnalyzeSneakPaths(selectedRow, selectedCol)

	// Update sneak path heatmap with sqrt transform for better visibility
	sneakMap := analysis.GetSneakMap()
	debug.Printf("Sneak map size: %dx%d", len(sneakMap), len(sneakMap[0]))

	// Apply sqrt transformation to make small variations more visible
	for i := range sneakMap {
		for j := range sneakMap[i] {
			sneakMap[i][j] = math.Sqrt(sneakMap[i][j])
		}
	}

	if len(sneakMap) > 0 && len(sneakMap[0]) > 0 {
		debug.Printf("Sneak sample values (sqrt): [0,0]=%.4f, [mid,mid]=%.4f", sneakMap[0][0], sneakMap[len(sneakMap)/2][len(sneakMap[0])/2])
	}
	fyne.Do(func() {
		ca.sneakPathHeatmap.SetData(sneakMap)
		ca.sneakPathHeatmap.SetSelection(selectedRow, selectedCol)
	})

	// Calculate signal-to-sneak ratio
	snr := 0.0
	if analysis.TotalSneak > 0 {
		snr = analysis.TotalSignal / analysis.TotalSneak
	}

	// Update stats
	ca.statsLabel.SetText(fmt.Sprintf(
		"Sneak Path Analysis\n"+
			"Selected Cell: [%d, %d]\n"+
			"Signal Current: %.6f\n"+
			"Total Sneak: %.6f\n"+
			"Max Sneak/Signal: %.2f%%\n"+
			"Avg Sneak/Signal: %.2f%%\n"+
			"Signal/Sneak Ratio: %.1f:1\n\n"+
			"Impact Assessment:\n%s",
		selectedRow, selectedCol,
		analysis.TotalSignal,
		analysis.TotalSneak,
		analysis.MaxSneakRatio*100,
		analysis.AvgSneakRatio*100,
		snr,
		getImpactAssessment(analysis.MaxSneakRatio),
	))

	// Update key stat and log
	ca.setKeyStatValue(fmt.Sprintf("SNR: %.1f:1", snr))

	ca.updateStatus(fmt.Sprintf("SNEAK | Complete: Signal-to-Sneak ratio %.1f:1. %s",
		snr, getImpactSummary(analysis.MaxSneakRatio)))
	ca.modeIndicator.SetMode(DemoModeIdle)
}

// resetArray resets the array with new random weights.
func (ca *CrossbarApp) resetArray() {
	ca.modeIndicator.SetMode(DemoModeWrite)
	ca.updateStatus("WRITE | Programming random conductance values (30 levels per cell)...")

	ca.programRandomWeights()
	ca.updateConductanceDisplay()

	// Protected clear of state
	ca.stateMu.Lock()
	ca.lastInput = nil
	ca.lastOutput = nil
	ca.stateMu.Unlock()
	ca.conductanceHeatmap.ClearSelection()
	ca.irDropHeatmap.ClearSelection()
	ca.sneakPathHeatmap.ClearSelection()
	ca.statsLabel.SetText(fmt.Sprintf(
		"Array Reset!\n\n"+
			"New random weights programmed\n"+
			"across all %d cells.\n\n"+
			"Each cell randomly assigned\n"+
			"to one of 30 discrete\n"+
			"conductance levels.\n\n"+
			"Ready for MVM operations!",
		ca.config.Rows*ca.config.Cols,
	))

	// Update key stat
	ca.setKeyStatValue(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols))

	ca.setEducationalContent("What You're Seeing", "Matrix-Vector Multiplication\n"+
		"using FeFET crossbar arrays.\n\n"+
		"The array computes I = W × V\n"+
		"using physics (Ohm's Law)\n"+
		"instead of digital logic.\n\n"+
		"All operations happen in\n"+
		"parallel - no sequential ALU!")
	ca.modeIndicator.SetMode(DemoModeIdle)

	// Auto-run MVM after reset
	ca.runEnhancedMVMInstant()
}

// getImpactAssessment returns a text assessment of sneak path severity.
func getImpactAssessment(maxRatio float64) string {
	if maxRatio < 0.01 {
		return "✓ Sneak paths negligible\n  Excellent cell isolation"
	} else if maxRatio < 0.05 {
		return "⚠ Moderate sneak paths\n  Consider selector devices"
	}
	return "✗ Significant sneak paths\n  1T1R or selector required"
}

// getImpactSummary returns a one-line summary for status messages.
func getImpactSummary(maxRatio float64) string {
	if maxRatio < 0.01 {
		return "Negligible parasitic currents - excellent!"
	} else if maxRatio < 0.05 {
		return "Moderate leakage - acceptable for many apps."
	}
	return "High sneak current - needs mitigation!"
}
