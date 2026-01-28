// Package gui provides tooltips with progressive disclosure (M1 UX fix)
// Shows key metrics in summary, full details available on request
package gui

import (
	"fmt"
	"math"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// ConductanceTooltip generates tooltip for conductance cell with progressive disclosure
// Shows key metrics first (5-7 lines), then detailed info
func ConductanceTooltip(row, col int, G float64, array *crossbar.Array) string {
	level := crossbar.GetLevel(G)

	// Calculate derived metrics
	conductanceUS := G*99 + 1                  // Assume 1-100 µS range
	resistance := 1.0 / (conductanceUS * 1e-6) // Convert to Ohms

	// Progressive disclosure: Key metrics first (M1 UX fix)
	tooltip := fmt.Sprintf(
		"CELL [%d, %d] - CONDUCTANCE\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Level:       %d / 29 (%.0f%%)\n"+
			"Conductance: %.2f µS\n"+
			"Resistance:  %.2f kΩ\n"+
			"Bits/cell:   %.1f bits\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Details:\n"+
			"  Normalized value: %.6f\n"+
			"  Position: (%d, %d)\n"+
			"  30-level analog storage\n",
		row, col,
		level, float64(level)/29.0*100,
		conductanceUS,
		resistance/1000.0,
		math.Log2(30),
		G,
		row, col,
	)

	return tooltip
}

// IRDropTooltip generates tooltip for IR drop analysis with progressive disclosure
func IRDropTooltip(row, col int, irAnalysis *crossbar.IRDropAnalysis, array *crossbar.Array) string {
	if irAnalysis == nil {
		return "Run MVM to compute IR drop analysis"
	}

	if row >= len(irAnalysis.EffectiveVoltage) || col >= len(irAnalysis.EffectiveVoltage[0]) {
		return "Cell out of range"
	}

	// Get IR drop data
	effectiveV := irAnalysis.EffectiveVoltage[row][col]
	dropPercent := (1.0 - effectiveV) * 100

	// Get conductance and calculate current impact
	G := array.GetConductanceMatrix()[row][col]
	conductanceUS := G*99 + 1
	currentLossPercent := dropPercent // Current loss is proportional to voltage drop

	// Severity assessment
	var severity string
	var severitySymbol string
	if dropPercent < 5 {
		severity = "OK"
		severitySymbol = "✓"
	} else if dropPercent < 10 {
		severity = "Moderate"
		severitySymbol = "⚠"
	} else {
		severity = "High"
		severitySymbol = "✗"
	}

	// Progressive disclosure: Key metrics first (M1 UX fix)
	tooltip := fmt.Sprintf(
		"CELL [%d, %d] - IR DROP\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"%s IR Drop: %.2f%% (%s)\n"+
			"Effective V: %.3f V\n"+
			"Current loss: %.1f%%\n"+
			"Worst cell: [%d,%d] at %.1f%%\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Details:\n"+
			"  Distance from driver: %d cells\n"+
			"  Conductance: %.2f µS\n"+
			"  Array avg drop: %.2f%%\n\n"+
			"Mitigation: wider wires, tiled arch\n",
		row, col,
		severitySymbol, dropPercent, severity,
		effectiveV,
		currentLossPercent,
		irAnalysis.WorstCaseCell[0], irAnalysis.WorstCaseCell[1],
		irAnalysis.MaxIRDrop*100,
		col+row, // Manhattan distance from corner
		conductanceUS,
		irAnalysis.AvgIRDrop*100,
	)

	return tooltip
}

// SneakPathTooltip generates tooltip for sneak path analysis with progressive disclosure
func SneakPathTooltip(row, col int, sneakAnalysis *crossbar.SneakPathAnalysis, selectedRow, selectedCol int, array *crossbar.Array) string {
	if sneakAnalysis == nil {
		return "Run MVM to compute sneak path analysis"
	}

	if row >= len(sneakAnalysis.SneakCurrents) || col >= len(sneakAnalysis.SneakCurrents[0]) {
		return "Cell out of range"
	}

	// Get sneak current data
	sneakCurrent := sneakAnalysis.SneakCurrents[row][col]
	signalCurrent := sneakAnalysis.TotalSignal
	sneakRatio := 0.0
	if signalCurrent > 0 {
		sneakRatio = sneakCurrent / signalCurrent * 100
	}

	// Determine sneak path type
	var pathType string
	isSelected := (row == selectedRow && col == selectedCol)
	sameRow := (row == selectedRow)
	sameCol := (col == selectedCol)

	if isSelected {
		pathType = "TARGET"
	} else if sameRow {
		pathType = "Row"
	} else if sameCol {
		pathType = "Column"
	} else {
		pathType = "Diagonal"
	}

	// SNR calculation
	snr := 0.0
	if sneakCurrent > 0 {
		snr = signalCurrent / sneakCurrent
	}

	// Severity
	var severity string
	var severitySymbol string
	if sneakRatio < 1 {
		severity = "OK"
		severitySymbol = "✓"
	} else if sneakRatio < 5 {
		severity = "Low"
		severitySymbol = "⚠"
	} else {
		severity = "High"
		severitySymbol = "✗"
	}

	// Distance from target
	manhattanDist := abs(row-selectedRow) + abs(col-selectedCol)

	// Progressive disclosure: Key metrics first (M1 UX fix)
	tooltip := fmt.Sprintf(
		"CELL [%d, %d] - SNEAK PATH\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"%s Sneak: %.2f%% (%s)\n"+
			"Path type: %s\n"+
			"SNR: %.1f:1\n"+
			"Target: [%d,%d], dist: %d\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Details:\n"+
			"  Signal: %.3f µA\n"+
			"  Sneak: %.3f µA\n"+
			"  Array max: %.2f%%\n\n"+
			"Mitigation: 1T1R, selectors\n",
		row, col,
		severitySymbol, sneakRatio, severity,
		pathType,
		snr,
		selectedRow, selectedCol, manhattanDist,
		signalCurrent*1e6,
		sneakCurrent*1e6,
		sneakAnalysis.MaxSneakRatio*100,
	)

	return tooltip
}

// MVMResultTooltip generates tooltip for MVM results with progressive disclosure
func MVMResultTooltip(row int, mvmResult *crossbar.MVMResult) string {
	if mvmResult == nil {
		return "Run MVM to see results"
	}

	if row >= len(mvmResult.IdealOutput) {
		return "Row out of range"
	}

	ideal := mvmResult.IdealOutput[row]
	actual := mvmResult.ActualOutput[row]
	errorVal := math.Abs(ideal - actual)
	errorPercent := 0.0
	if ideal > 0 {
		errorPercent = (errorVal / ideal) * 100
	}

	// Progressive disclosure: Key metrics first (M1 UX fix)
	tooltip := fmt.Sprintf(
		"OUTPUT ROW [%d] - MVM\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Ideal:  %.4f\n"+
			"Actual: %.4f\n"+
			"Error:  %.2f%%\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Global: RMSE=%.4f, Loss=%.1f%%\n"+
			"Energy: %.1f pJ (%.0f× vs GPU)\n"+
			"MACs: %d @ %.1f ns\n",
		row,
		ideal,
		actual,
		errorPercent,
		mvmResult.RMSE,
		mvmResult.AccuracyLoss,
		mvmResult.TotalEnergy,
		mvmResult.EnergyEfficiency,
		mvmResult.MACOperations,
		mvmResult.Latency,
	)

	return tooltip
}

// ComprehensiveTooltip generates a compact multi-analysis tooltip with progressive disclosure
func ComprehensiveTooltip(row, col int, array *crossbar.Array, irAnalysis *crossbar.IRDropAnalysis, sneakAnalysis *crossbar.SneakPathAnalysis, mvmResult *crossbar.MVMResult) string {
	G := array.GetConductanceMatrix()[row][col]
	level := crossbar.GetLevel(G)
	conductanceUS := G*99 + 1

	// Progressive disclosure: Compact summary (M1 UX fix)
	tooltip := fmt.Sprintf(
		"CELL [%d, %d] - SUMMARY\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Level: %d/29 | G: %.1f µS\n",
		row, col,
		level,
		conductanceUS,
	)

	if irAnalysis != nil && row < len(irAnalysis.EffectiveVoltage) && col < len(irAnalysis.EffectiveVoltage[0]) {
		dropPercent := (1.0 - irAnalysis.EffectiveVoltage[row][col]) * 100
		tooltip += fmt.Sprintf("IR Drop: %.1f%%\n", dropPercent)
	}

	if sneakAnalysis != nil && row < len(sneakAnalysis.SneakCurrents) && col < len(sneakAnalysis.SneakCurrents[0]) {
		sneakRatio := 0.0
		if sneakAnalysis.TotalSignal > 0 {
			sneakRatio = sneakAnalysis.SneakCurrents[row][col] / sneakAnalysis.TotalSignal * 100
		}
		tooltip += fmt.Sprintf("Sneak: %.1f%%\n", sneakRatio)
	}

	if mvmResult != nil {
		tooltip += fmt.Sprintf("Energy: %.1f pJ (%.0f× vs GPU)\n",
			mvmResult.TotalEnergy,
			mvmResult.EnergyEfficiency,
		)
	}

	return tooltip
}

// Helper function
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
