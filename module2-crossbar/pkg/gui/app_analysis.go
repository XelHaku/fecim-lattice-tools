//go:build legacy_fyne

// Package gui - Analysis and metrics functions for crossbar app
package gui

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"fecim-lattice-tools/shared/crossbar"
)

// updateEnhancedWidgets updates all enhanced visualization widgets with MVM results
func (ca *CrossbarApp) updateEnhancedWidgets(mvmResult *crossbar.MVMResult) {
	// Protected read of lastInput
	ca.stateMu.RLock()
	input := ca.lastInput
	ca.stateMu.RUnlock()

	// Update metrics panel
	// Estimate baseline accuracy (for demo purposes, use 90%)
	baselineAcc := 90.0
	actualAcc := baselineAcc - mvmResult.AccuracyLoss
	if ca.metricsPanel != nil {
		ca.metricsPanel.UpdateMetrics(
			baselineAcc,
			actualAcc,
			mvmResult.TotalEnergy,
			mvmResult.GPUEquivalentEnergy,
			mvmResult.MACOperations,
			mvmResult.Latency,
		)
	}

	// Update comparison badge
	if ca.comparisonBadge != nil {
		ca.comparisonBadge.UpdateValues(
			fmt.Sprintf("%.2f pJ", mvmResult.TotalEnergy),
			fmt.Sprintf("%.0f pJ", mvmResult.GPUEquivalentEnergy),
			fmt.Sprintf("%.0f× better", mvmResult.EnergyEfficiency),
		)
	}

	// Update accuracy waterfall
	if ca.accuracyWaterfall != nil {
		opts := crossbar.DefaultMVMOptions()
		ca.stateMu.RLock()
		opts.Architecture = ca.architecture
		ca.stateMu.RUnlock()
		opts.Temperature = ca.currentTemperatureK()
		degradation, _ := ca.array.ComputeAccuracyDegradationWithOptions(input, baselineAcc, opts)
		if degradation != nil {
			steps := make([]WaterfallStep, len(degradation.Degradations))
			colors := []color.RGBA{
				{100, 200, 100, 255}, // Green - baseline
				{150, 200, 150, 255}, // Light green
				{200, 200, 100, 255}, // Yellow
				{255, 180, 100, 255}, // Orange
				{255, 100, 100, 255}, // Red
			}
			for i, deg := range degradation.Degradations {
				steps[i] = WaterfallStep{
					Label:    deg.Source,
					Accuracy: deg.AccuracyNow,
					Loss:     deg.Loss,
					Color:    colors[i%len(colors)],
				}
			}
			ca.accuracyWaterfall.SetSteps(steps)
		}
	}

	// Update before/after toggle
	if ca.beforeAfterToggle != nil {
		// Ideal: programmed conductances (no noise/variation)
		idealMatrix := ca.array.GetConductanceMatrix()
		// Actual: effective conductances with per-cell noise factors applied
		actualMatrix := ca.array.GetEffectiveConductanceMatrix()
		ca.beforeAfterToggle.SetData(idealMatrix, actualMatrix)
	}

	// Check if baselines need to be computed (first run after array creation)
	ca.stateMu.RLock()
	baselineIRExists := ca.baselineMaxIRDrop > 0
	baselineSneakExists := ca.baselineMaxSneak > 0
	ca.stateMu.RUnlock()

	// Compute passive baseline if not yet set (needed for legend scaling)
	// Fixed at 100% - sneak ratio shows each cell's contribution relative to signal
	// Same-row/column cells: ~100% (direct parallel paths)
	// Off-diagonal cells: ~30% (series path through 3 cells)
	// With 1T1R/2T1R: near 0% (transistor isolation)
	if !baselineSneakExists {
		centerRow := ca.config.Rows / 2
		centerCol := ca.config.Cols / 2
		getDebug().Printf("=== COMPUTING SNEAK BASELINE ===")
		getDebug().Printf("  Center cell: [%d, %d]", centerRow, centerCol)
		passiveSneak := ca.array.AnalyzeSneakPathsWithIsolation(centerRow, centerCol, 1.0) // 0T1R factor
		getDebug().Printf("  Passive analysis result:")
		getDebug().Printf("    MaxSneakRatio: %.6f (%.1f%%)", passiveSneak.MaxSneakRatio, passiveSneak.MaxSneakRatio*100)
		getDebug().Printf("    AvgSneakRatio: %.6f (%.1f%%)", passiveSneak.AvgSneakRatio, passiveSneak.AvgSneakRatio*100)
		getDebug().Printf("    TotalSneak: %.6e", passiveSneak.TotalSneak)
		getDebug().Printf("    TotalSignal: %.6e", passiveSneak.TotalSignal)
		ca.stateMu.Lock()
		// Use fixed 100% baseline for intuitive visualization
		// Values > 100% will be clamped (cells with higher G than signal)
		ca.baselineMaxSneak = 100.0
		getDebug().Printf("  Baseline Sneak set to fixed: %.2f%%", ca.baselineMaxSneak)
		ca.stateMu.Unlock()
	}

	if !baselineIRExists && input != nil && len(input) > 0 {
		// Use passive wire params (1.5x multiplier for 0T1R)
		passiveParams := crossbar.DefaultWireParams()
		passiveParams.RwordLine *= 1.5
		passiveParams.RbitLine *= 1.5
		ca.applyTemperatureToWireParams(passiveParams)
		passiveIR := ca.array.AnalyzeIRDrop(input, passiveParams)
		ca.stateMu.Lock()
		ca.baselineMaxIRDrop = passiveIR.MaxIRDrop * 100
		if ca.baselineMaxIRDrop < 1 {
			ca.baselineMaxIRDrop = 1
		}
		getDebug().Printf("Baseline IR Drop set from passive: %.2f%%", ca.baselineMaxIRDrop)
		ca.stateMu.Unlock()
	}

	// Update IR drop heatmap
	if mvmResult.IRDropAnalysis != nil {
		ca.stateMu.Lock()
		ca.lastIRDropAnalysis = mvmResult.IRDropAnalysis
		// Use pre-computed passive baseline (don't update - keep fixed for comparison)
		baselineIR := ca.baselineMaxIRDrop
		ca.stateMu.Unlock()

		getDebug().Printf("IR Drop data: MaxDrop=%.4f%%, Baseline=%.2f%%",
			mvmResult.IRDropAnalysis.MaxIRDrop*100, baselineIR)

		// Use passive baseline for legend
		ca.irLegend.SetRange(0, baselineIR)

		// Get IR drop map scaled to baseline (convert percent to fraction)
		irMap := mvmResult.IRDropAnalysis.GetIRDropMapWithScale(baselineIR / 100)
		ca.irDropHeatmap.SetData(irMap)

		// Preserve user's cell selection if they have one, otherwise show worst cell
		ca.stateMu.RLock()
		userRow, userCol := ca.selectedRow, ca.selectedCol
		ca.stateMu.RUnlock()
		if userRow >= 0 && userCol >= 0 {
			// User has a selection - preserve it
			ca.irDropHeatmap.SetSelection(userRow, userCol)
		} else {
			// No user selection - highlight worst-case cell
			ca.irDropHeatmap.SetSelection(
				mvmResult.IRDropAnalysis.WorstCaseCell[0],
				mvmResult.IRDropAnalysis.WorstCaseCell[1],
			)
		}
		ca.irDropHeatmap.Refresh() // Force refresh

		// Add badge to IR Drop tab (C2 accessibility fix - discoverability)
		ca.setTabBadge("IR Drop")
	} else {
		getDebug().Println("Warning: IRDropAnalysis is nil")
	}

	// Update sneak path heatmap
	if mvmResult.SneakPathAnalysis != nil {
		ca.stateMu.Lock()
		ca.lastSneakAnalysis = mvmResult.SneakPathAnalysis
		// Use pre-computed passive baseline (don't update - keep fixed for comparison)
		baselineSneak := ca.baselineMaxSneak
		ca.stateMu.Unlock()

		analysis := mvmResult.SneakPathAnalysis
		maxSneakPercent := analysis.MaxSneakRatio * 100
		avgSneakPercent := analysis.AvgSneakRatio * 100
		getDebug().Printf("=== SNEAK PATH UPDATE ===")
		getDebug().Printf("  MaxSneakRatio: %.6f (%.4f%%)", analysis.MaxSneakRatio, maxSneakPercent)
		getDebug().Printf("  AvgSneakRatio: %.6f (%.4f%%)", analysis.AvgSneakRatio, avgSneakPercent)
		getDebug().Printf("  TotalSneak: %.6e, TotalSignal: %.6e", analysis.TotalSneak, analysis.TotalSignal)
		getDebug().Printf("  Baseline: %.2f%%", baselineSneak)
		getDebug().Printf("  SneakCurrents size: %dx%d", len(analysis.SneakCurrents), func() int {
			if len(analysis.SneakCurrents) > 0 {
				return len(analysis.SneakCurrents[0])
			}
			return 0
		}())

		// Sample some sneak values
		if len(analysis.SneakCurrents) > 0 && len(analysis.SneakCurrents[0]) > 0 {
			mid := len(analysis.SneakCurrents) / 2
			getDebug().Printf("  Sample SneakCurrents[0][0]=%.6e, [mid][mid]=%.6e, [0][mid]=%.6e, [mid][0]=%.6e",
				analysis.SneakCurrents[0][0],
				analysis.SneakCurrents[mid][mid],
				analysis.SneakCurrents[0][mid],
				analysis.SneakCurrents[mid][0])
		}

		// Use passive baseline for legend
		ca.sneakLegend.SetRange(0, baselineSneak)

		// Get sneak map normalized to baseline
		sneakMap := analysis.GetSneakMapWithScale(baselineSneak / 100)
		getDebug().Printf("  SneakMap size: %dx%d", len(sneakMap), func() int {
			if len(sneakMap) > 0 {
				return len(sneakMap[0])
			}
			return 0
		}())
		if len(sneakMap) > 0 && len(sneakMap[0]) > 0 {
			mid := len(sneakMap) / 2
			getDebug().Printf("  Sample SneakMap[0][0]=%.4f, [mid][mid]=%.4f, [0][mid]=%.4f, [mid][0]=%.4f",
				sneakMap[0][0], sneakMap[mid][mid], sneakMap[0][mid], sneakMap[mid][0])
		}

		ca.sneakPathHeatmap.SetData(sneakMap)

		// Add badge to Sneak Paths tab (C2 accessibility fix - discoverability)
		ca.setTabBadge("Sneak Paths")
	} else {
		getDebug().Println("Warning: SneakPathAnalysis is nil")
	}
}

// exportData exports array weights and analysis to files in the data folder.
func (ca *CrossbarApp) exportData() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Ensure data/crossbar folder exists
	dataDir := filepath.Join("data", "crossbar")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.updateStatus(fmt.Sprintf("Export failed: cannot create data/crossbar folder: %v", err))
		return
	}

	// Export weights CSV to data/crossbar folder
	weightsPath := filepath.Join(dataDir, fmt.Sprintf("crossbar_weights_%s.csv", timestamp))
	if err := ca.array.ExportWeightsCSV(weightsPath); err != nil {
		ca.updateStatus(fmt.Sprintf("Export failed: %v", err))
		return
	}

	// Convert to absolute path for dialog
	absWeightsPath, _ := filepath.Abs(weightsPath)

	// Export analysis JSON to data/crossbar folder
	if ca.lastMVMResult != nil {
		analysisPath := filepath.Join(dataDir, fmt.Sprintf("crossbar_analysis_%s.json", timestamp))
		if err := ca.array.ExportAnalysisJSON(analysisPath, ca.lastMVMResult); err != nil {
			ca.updateStatus(fmt.Sprintf("Export failed: %v", err))
			return
		}
		ca.updateStatus(fmt.Sprintf("Exported: %s, %s", weightsPath, analysisPath))

		// Convert to absolute path for dialog
		absAnalysisPath, _ := filepath.Abs(analysisPath)

		// Show success dialog with file paths
		fyne.Do(func() {
			ca.showExportSuccessDialog([]string{absWeightsPath, absAnalysisPath})
		})
	} else {
		ca.updateStatus(fmt.Sprintf("Exported: %s (run MVM for analysis)", weightsPath))

		// Show success dialog with single file path
		fyne.Do(func() {
			ca.showExportSuccessDialog([]string{absWeightsPath})
		})
	}
}

// onBeforeAfterCellTapped handles clicks on the Ideal vs Actual comparison heatmaps.
func (ca *CrossbarApp) onBeforeAfterCellTapped(row, col int, isIdeal bool) {
	if ca.beforeAfterToggle == nil {
		return
	}

	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	var idealVal, actualVal float64
	if ca.beforeAfterToggle.idealData != nil && row < len(ca.beforeAfterToggle.idealData) &&
		col < len(ca.beforeAfterToggle.idealData[0]) {
		idealVal = ca.beforeAfterToggle.idealData[row][col]
	}
	if ca.beforeAfterToggle.actualData != nil && row < len(ca.beforeAfterToggle.actualData) &&
		col < len(ca.beforeAfterToggle.actualData[0]) {
		actualVal = ca.beforeAfterToggle.actualData[row][col]
	}

	idealLevel := crossbar.GetLevel(idealVal)
	actualLevel := crossbar.GetLevel(actualVal)
	diff := idealVal - actualVal
	diffPercent := 0.0
	if idealVal > 0 {
		diffPercent = (diff / idealVal) * 100
	}

	source := "Actual"
	if isIdeal {
		source = "Ideal"
	}

	tooltip := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"CELL [%d, %d] - IDEAL vs ACTUAL\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Clicked: %s heatmap\n\n"+
			"Ideal Value:\n"+
			"  Conductance:  %.4f (L%d/29)\n"+
			"  Current:      %.2f µA\n\n"+
			"Actual Value:\n"+
			"  Conductance:  %.4f (L%d/29)\n"+
			"  Current:      %.2f µA\n\n"+
			"Degradation:\n"+
			"  Difference:   %.4f (%.1f%%)\n"+
			"  Level shift:  %d levels\n\n"+
			"Impact:\n"+
			"  %s\n",
		row, col,
		source,
		idealVal, idealLevel, idealVal*99+1,
		actualVal, actualLevel, actualVal*99+1,
		math.Abs(diff), math.Abs(diffPercent),
		int(math.Abs(float64(idealLevel-actualLevel))),
		ca.assessDegradationImpact(diffPercent),
	)

	ca.statsLabel.SetText(tooltip)
	ca.updateStatus(fmt.Sprintf("COMPARISON | Cell [%d,%d]: Ideal L%d → Actual L%d (%.1f%% change)",
		row, col, idealLevel, actualLevel, diffPercent))
}

// onBeforeAfterCellHover handles hover on the Ideal vs Actual comparison heatmaps.
func (ca *CrossbarApp) onBeforeAfterCellHover(row, col int, value float64, isIdeal bool) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells to compare ideal vs actual values")
		return
	}

	if ca.beforeAfterToggle == nil {
		return
	}

	var idealVal, actualVal float64
	if ca.beforeAfterToggle.idealData != nil && row < len(ca.beforeAfterToggle.idealData) &&
		col < len(ca.beforeAfterToggle.idealData[0]) {
		idealVal = ca.beforeAfterToggle.idealData[row][col]
	}
	if ca.beforeAfterToggle.actualData != nil && row < len(ca.beforeAfterToggle.actualData) &&
		col < len(ca.beforeAfterToggle.actualData[0]) {
		actualVal = ca.beforeAfterToggle.actualData[row][col]
	}

	idealLevel := crossbar.GetLevel(idealVal)
	actualLevel := crossbar.GetLevel(actualVal)
	diff := math.Abs(idealVal - actualVal)

	source := "Actual"
	if isIdeal {
		source = "Ideal"
	}

	ca.hoverInfoLabel.SetText(fmt.Sprintf(
		"[%d,%d] %s │ Ideal: L%d (%.3f) │ Actual: L%d (%.3f) │ Δ=%.4f (%d levels)",
		row, col, source, idealLevel, idealVal, actualLevel, actualVal, diff, int(math.Abs(float64(idealLevel-actualLevel)))))
}

// assessDegradationImpact returns a qualitative assessment of degradation.
func (ca *CrossbarApp) assessDegradationImpact(diffPercent float64) string {
	absDiff := math.Abs(diffPercent)
	if absDiff < 1 {
		return "Negligible - within noise margin"
	} else if absDiff < 5 {
		return "Minor - acceptable for most applications"
	} else if absDiff < 10 {
		return "Moderate - may affect precision tasks"
	} else if absDiff < 20 {
		return "Significant - requires compensation"
	}
	return "Critical - exceeds tolerance limits"
}

// getAccuracyStatus returns a status message based on accuracy.
// Note: No fixed target - compares against peer-reviewed benchmarks (96.6-98.24%)
func (ca *CrossbarApp) getAccuracyStatus(accuracy float64) string {
	if accuracy >= 96.0 {
		return "✓ Excellent - matches peer-reviewed benchmarks"
	} else if accuracy >= 90.0 {
		return "✓ Good - within practical range"
	} else if accuracy >= 80.0 {
		return "⚠ Moderate - optimization may help"
	}
	return "⚠ Low - check noise and quantization settings"
}

// showExportSuccessDialog displays a success dialog with export file paths
func (ca *CrossbarApp) showExportSuccessDialog(filePaths []string) {
	if ca.window == nil {
		return
	}

	var message string
	if len(filePaths) == 1 {
		message = fmt.Sprintf("File exported successfully:\n\n%s", filePaths[0])
	} else {
		message = "Files exported successfully:\n\n"
		for _, path := range filePaths {
			message += fmt.Sprintf("• %s\n", path)
		}
	}

	// Simple information dialog with OK button
	dialog.ShowInformation("Export Complete", message, ca.window)
}
