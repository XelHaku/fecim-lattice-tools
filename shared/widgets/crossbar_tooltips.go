// Package widgets provides reusable UI components.
// crossbar_tooltips.go provides tooltip string generators for crossbar GUI elements.
// These functions produce progressive-disclosure tooltip text for cell-level analysis.
package widgets

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/crossbar"
)

// ConductanceTooltip generates tooltip for conductance cell with progressive disclosure.
func ConductanceTooltip(row, col int, G float64, array *crossbar.Array) string {
	level := crossbar.GetLevel(G)

	conductanceUS := G*99 + 1
	resistance := 1.0 / (conductanceUS * 1e-6)

	return fmt.Sprintf(
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
			"  30-level baseline (conference claim)\n",
		row, col,
		level, float64(level)/29.0*100,
		conductanceUS,
		resistance/1000.0,
		math.Log2(30),
		G,
		row, col,
	)
}

// IRDropTooltip generates tooltip for IR drop analysis with progressive disclosure.
func IRDropTooltip(row, col int, irAnalysis *crossbar.IRDropAnalysis, array *crossbar.Array) string {
	return IRDropTooltipWithArch(row, col, irAnalysis, array, "")
}

// IRDropTooltipWithArch generates tooltip for IR drop analysis including architecture info.
func IRDropTooltipWithArch(row, col int, irAnalysis *crossbar.IRDropAnalysis, array *crossbar.Array, arch string) string {
	if irAnalysis == nil {
		return "Run MVM to compute IR drop analysis"
	}

	if row >= len(irAnalysis.EffectiveVoltage) || col >= len(irAnalysis.EffectiveVoltage[0]) {
		return "Cell out of range"
	}

	effectiveV := irAnalysis.EffectiveVoltage[row][col]
	dropPercent := (1.0 - effectiveV) * 100

	G := array.GetConductanceMatrix()[row][col]
	conductanceUS := G*99 + 1
	currentLossPercent := dropPercent

	var severity, severitySymbol string
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

	archDisplay := arch
	if archDisplay == "" {
		archDisplay = "0T1R"
	}

	return fmt.Sprintf(
		"CELL [%d, %d] - IR DROP [%s]\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"%s IR Drop: %.2f%% (%s)\n"+
			"Effective V: %.3f V\n"+
			"Current loss: %.1f%%\n"+
			"Worst cell: [%d,%d] at %.1f%%\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Details:\n"+
			"  Architecture: %s\n"+
			"  Distance from driver: %d cells\n"+
			"  Conductance: %.2f µS\n"+
			"  Array avg drop: %.2f%%\n\n"+
			"Mitigation: wider wires, tiled arch\n",
		row, col, archDisplay,
		severitySymbol, dropPercent, severity,
		effectiveV,
		currentLossPercent,
		irAnalysis.WorstCaseCell[0], irAnalysis.WorstCaseCell[1],
		irAnalysis.MaxIRDrop*100,
		archDisplay,
		col+row,
		conductanceUS,
		irAnalysis.AvgIRDrop*100,
	)
}

// SneakPathTooltip generates tooltip for sneak path analysis with progressive disclosure.
func SneakPathTooltip(row, col int, sneakAnalysis *crossbar.SneakPathAnalysis, selectedRow, selectedCol int, array *crossbar.Array) string {
	return SneakPathTooltipWithArch(row, col, sneakAnalysis, selectedRow, selectedCol, array, "")
}

// SneakPathTooltipWithArch generates tooltip for sneak path analysis including architecture info.
func SneakPathTooltipWithArch(row, col int, sneakAnalysis *crossbar.SneakPathAnalysis, selectedRow, selectedCol int, array *crossbar.Array, arch string) string {
	if sneakAnalysis == nil {
		return "Run MVM to compute sneak path analysis"
	}

	if row >= len(sneakAnalysis.SneakCurrents) || col >= len(sneakAnalysis.SneakCurrents[0]) {
		return "Cell out of range"
	}

	sneakCurrent := sneakAnalysis.SneakCurrents[row][col]
	signalCurrent := sneakAnalysis.TotalSignal
	sneakRatio := 0.0
	if signalCurrent > 0 {
		sneakRatio = sneakCurrent / signalCurrent * 100
	}

	sneakRatioDisplay := sneakRatio
	sneakRatioNote := ""
	if sneakRatio > 100.0 {
		sneakRatioNote = fmt.Sprintf(" (actual %.1f%%)", sneakRatio)
		sneakRatioDisplay = 100.0
	}

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

	snr := 0.0
	if sneakCurrent > 0 {
		snr = signalCurrent / sneakCurrent
	}

	var severity, severitySymbol string
	if sneakRatio < 1 {
		severity = "OK"
		severitySymbol = "✓"
	} else if sneakRatio < 5 {
		severity = "Low"
		severitySymbol = "⚠"
	} else if sneakRatio < 100 {
		severity = "High"
		severitySymbol = "✗"
	} else {
		severity = "Critical"
		severitySymbol = "✗"
	}

	manhattanDist := crossbarAbs(row-selectedRow) + crossbarAbs(col-selectedCol)

	archDisplay := arch
	if archDisplay == "" {
		archDisplay = "0T1R"
	}

	archNote := "Mitigation: 1T1R, selectors"
	if arch == "1T1R" || arch == "1T1R GATE" {
		archNote = "1T1R: ~1000× sneak reduction"
	} else if arch == "2T1R" {
		archNote = "2T1R: ~10000× sneak reduction"
	}

	arrayMaxPercent := sneakAnalysis.MaxSneakRatio * 100
	arrayMaxNote := ""
	if arrayMaxPercent > 100.0 {
		arrayMaxNote = fmt.Sprintf(" (actual %.1f%% - sneak > signal)", arrayMaxPercent)
		arrayMaxPercent = 100.0
	}

	return fmt.Sprintf(
		"CELL [%d, %d] - SNEAK PATH [%s]\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"%s Sneak: %.2f%%%s (%s)\n"+
			"Path type: %s\n"+
			"SNR: %.1f:1\n"+
			"Target: [%d,%d], dist: %d\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Details:\n"+
			"  Architecture: %s\n"+
			"  Signal: %.3f µA\n"+
			"  Sneak: %.3f µA\n"+
			"  Array max: %.2f%%%s\n\n"+
			"%s\n",
		row, col, archDisplay,
		severitySymbol, sneakRatioDisplay, sneakRatioNote, severity,
		pathType,
		snr,
		selectedRow, selectedCol, manhattanDist,
		archDisplay,
		signalCurrent*1e6,
		sneakCurrent*1e6,
		arrayMaxPercent, arrayMaxNote,
		archNote,
	)
}

// MVMResultTooltip generates tooltip for MVM results with progressive disclosure.
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

	return fmt.Sprintf(
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
}

// ComprehensiveTooltip generates a compact multi-analysis tooltip with progressive disclosure.
func ComprehensiveTooltip(row, col int, array *crossbar.Array, irAnalysis *crossbar.IRDropAnalysis, sneakAnalysis *crossbar.SneakPathAnalysis, mvmResult *crossbar.MVMResult) string {
	G := array.GetConductanceMatrix()[row][col]
	level := crossbar.GetLevel(G)
	conductanceUS := G*99 + 1

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

// crossbarAbs is a package-local integer absolute value helper.
func crossbarAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
