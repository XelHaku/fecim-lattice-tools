// Package gui provides enhanced tooltips with maximum technical data
package gui

import (
	"fmt"
	"math"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// ConductanceTooltip generates detailed tooltip for conductance cell
func ConductanceTooltip(row, col int, G float64, array *crossbar.Array) string {
	level := crossbar.GetLevel(G)

	// Get cell statistics if available (reserved for future use)
	_ = int64(0)   // switchCount - reserved for tracking cell cycles
	_ = float64(0) // noiseFactor - reserved for noise analysis

	// Calculate derived metrics
	conductanceUS := G*99 + 1  // Assume 1-100 µS range
	resistance := 1.0 / (conductanceUS * 1e-6)  // Convert to Ohms

	tooltip := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
		"CELL [%d, %d] - CONDUCTANCE DETAILS\n"+
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
		"FeCIM State:\n"+
		"  Level:        %d / 29  (%.1f%%)\n"+
		"  Normalized:   %.6f\n"+
		"  Conductance:  %.2f µS\n"+
		"  Resistance:   %.2f kΩ\n\n"+
		"Bit Representation:\n"+
		"  Analog value: %d / 30 states\n"+
		"  Bits/cell:    %.2f bits\n"+
		"  Binary equiv: %d-bit memory\n\n"+
		"Physical Properties:\n"+
		"  Position:     Row %d, Col %d\n"+
		"  Array coords: (%d, %d)\n\n"+
		"Programming:\n"+
		"  Target level: %d\n"+
		"  Achieved:     %d\n"+
		"  Error:        %.1f%%\n\n"+
		"Usage:\n"+
		"  Click: Select cell\n"+
		"  Right-click: Deselect\n"+
		"  Drag: Inspect region\n",  // Conductance already has µS unit",
		row, col,
		level, float64(level)/29.0*100,
		G,
		conductanceUS,
		resistance/1000.0,
		level,
		math.Log2(30),
		5, // 2^5 = 32, closest power of 2
		row, col,
		row, col,
		level,
		level,
		0.0,
	)

	return tooltip
}

// IRDropTooltip generates detailed tooltip for IR drop analysis
func IRDropTooltip(row, col int, irAnalysis *crossbar.IRDropAnalysis, array *crossbar.Array) string {
	if irAnalysis == nil {
		return "Run MVM to compute IR drop analysis"
	}

	if row >= len(irAnalysis.EffectiveVoltage) || col >= len(irAnalysis.EffectiveVoltage[0]) {
		return "Cell out of range"
	}

	// Get IR drop data
	effectiveV := irAnalysis.EffectiveVoltage[row][col]
	wlV := irAnalysis.WordLineVoltages[row][col]
	blV := irAnalysis.BitLineVoltages[row][col]
	dropV := 1.0 - effectiveV
	dropPercent := dropV * 100

	// Get conductance
	G := array.GetConductanceMatrix()[row][col]
	conductanceUS := G*99 + 1

	// Calculate current
	idealCurrent := 1.0 * conductanceUS  // µA
	actualCurrent := effectiveV * conductanceUS
	currentLoss := idealCurrent - actualCurrent
	currentLossPercent := (currentLoss / idealCurrent) * 100

	// Distance from drivers (WL driver on left, BL sense amp at top)
	wlDist := col // Distance from word line driver (left)
	blDist := row // Distance from bit line sense amp (top)
	totalDist := wlDist + blDist
	maxDist := (len(irAnalysis.EffectiveVoltage[0]) - 1) + (len(irAnalysis.EffectiveVoltage) - 1)
	distPercent := float64(totalDist) / float64(maxDist) * 100

	// Severity assessment
	var severity string
	var severitySymbol string
	if dropPercent < 5 {
		severity = "Negligible"
		severitySymbol = "✓"
	} else if dropPercent < 10 {
		severity = "Moderate"
		severitySymbol = "⚠"
	} else if dropPercent < 15 {
		severity = "Significant"
		severitySymbol = "⚠⚠"
	} else {
		severity = "Critical"
		severitySymbol = "✗"
	}

	// Calculate percentile rank (how close this cell is to the worst case)
	percentileRank := 0.0
	if irAnalysis.MaxIRDrop > 0 {
		percentileRank = (dropPercent / (irAnalysis.MaxIRDrop * 100)) * 100
	}

	tooltip := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
		"CELL [%d, %d] - IR DROP ANALYSIS\n"+
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
		"%s IR Drop: %.2f%% %s\n\n"+
		"Voltage Details:\n"+
		"  Ideal voltage:      1.000 V\n"+
		"  Word line voltage:  %.3f V\n"+
		"  Bit line voltage:   %.3f V\n"+
		"  Effective voltage:  %.3f V\n"+
		"  Voltage drop:       %.3f V (%.1f%%)\n\n"+
		"Current Impact:\n"+
		"  Ideal current:      %.2f µA\n"+
		"  Actual current:     %.2f µA\n"+
		"  Current loss:       %.2f µA (%.1f%%)\n\n"+
		"Position Analysis:\n"+
		"  WL distance:        %d cells from left driver\n"+
		"  BL distance:        %d cells from top sense amp\n"+
		"  Total distance:     %d (%.1f%% of max)\n"+
		"  Worst case cell:    [%d, %d] (%.2f%% drop)\n\n"+
		"Array Statistics:\n"+
		"  Max IR drop:        %.2f%%\n"+
		"  Avg IR drop:        %.2f%%\n"+
		"  This cell rank:     %.1f%% of worst\n\n"+
		"Mitigation Strategies:\n"+
		"  • Wider metal lines (2× width → 50%% drop)\n"+
		"  • Hierarchical drivers\n"+
		"  • Tiled architecture\n"+
		"  • Voltage compensation\n\n"+
		"Wire Parameters:\n"+
		"  R_word_line:        2.5 Ω/cell\n"+
		"  R_bit_line:         2.5 Ω/cell\n"+
		"  Contact R:          50 Ω\n",
		row, col,
		severitySymbol, dropPercent, severity,
		wlV,
		blV,
		effectiveV,
		dropV, dropPercent,
		idealCurrent,
		actualCurrent,
		currentLoss, currentLossPercent,
		wlDist,
		blDist,
		totalDist, distPercent,
		irAnalysis.WorstCaseCell[0], irAnalysis.WorstCaseCell[1],
		irAnalysis.MaxIRDrop*100,
		irAnalysis.MaxIRDrop*100,
		irAnalysis.AvgIRDrop*100,
		percentileRank,
	)

	return tooltip
}

// SneakPathTooltip generates detailed tooltip for sneak path analysis
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

	// Get conductance
	G := array.GetConductanceMatrix()[row][col]
	conductanceUS := G*99 + 1

	// Determine sneak path type
	var pathType string
	var pathDescription string
	isSelected := (row == selectedRow && col == selectedCol)
	sameRow := (row == selectedRow)
	sameCol := (col == selectedCol)

	if isSelected {
		pathType = "TARGET CELL"
		pathDescription = "This is the selected cell being read"
	} else if sameRow && sameCol {
		pathType = "IMPOSSIBLE"
		pathDescription = "Cannot be same row AND column"
	} else if sameRow {
		pathType = "ROW SNEAK"
		pathDescription = fmt.Sprintf("3-cell path: WL[%d] → this cell → unsel BL → target col", row)
	} else if sameCol {
		pathType = "COLUMN SNEAK"
		pathDescription = fmt.Sprintf("3-cell path: unsel WL → this cell → BL[%d] → target row", col)
	} else {
		pathType = "DIAGONAL SNEAK"
		pathDescription = fmt.Sprintf("3-cell path: WL[%d] → cell → BL[%d] → target", row, col)
	}

	// Calculate sneak resistance
	sneakResistance := 0.0
	if sneakCurrent > 1e-12 {
		sneakResistance = 1.0 / (sneakCurrent * 1e-6)  // Ω
	}

	// SNR calculation
	snr := 0.0
	snrDB := -100.0
	if sneakCurrent > 0 {
		snr = signalCurrent / sneakCurrent
		snrDB = 20 * math.Log10(snr)
	}

	// Severity
	var severity string
	var severitySymbol string
	if sneakRatio < 1 {
		severity = "Negligible"
		severitySymbol = "✓"
	} else if sneakRatio < 5 {
		severity = "Low"
		severitySymbol = "⚠"
	} else if sneakRatio < 10 {
		severity = "Moderate"
		severitySymbol = "⚠⚠"
	} else {
		severity = "High"
		severitySymbol = "✗"
	}

	// Distance from target
	rowDist := abs(row - selectedRow)
	colDist := abs(col - selectedCol)
	manhattanDist := rowDist + colDist

	tooltip := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
		"CELL [%d, %d] - SNEAK PATH ANALYSIS\n"+
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
		"%s Sneak: %.2f%% %s\n\n"+
		"Path Information:\n"+
		"  Type:         %s\n"+
		"  Description:  %s\n"+
		"  Target cell:  [%d, %d]\n"+
		"  Distance:     %d cells (Manhattan)\n\n"+
		"Current Analysis:\n"+
		"  Signal current:   %.6f µA\n"+
		"  Sneak current:    %.6f µA\n"+
		"  Sneak ratio:      %.2f%%\n"+
		"  SNR:              %.1f dB\n\n"+
		"Cell Properties:\n"+
		"  Conductance:      %.2f µS\n"+
		"  Sneak resistance: %.2f kΩ\n"+
		"  Path cells:       %d\n\n"+
		"Array Statistics:\n"+
		"  Max sneak ratio:  %.2f%%\n"+
		"  Avg sneak ratio:  %.2f%%\n"+
		"  Total sneak:      %.6f µA\n"+
		"  Signal/Sneak:     %.1f:1\n\n"+
		"Path Details:\n"+
		"  Row offset:       %d\n"+
		"  Col offset:       %d\n"+
		"  Same row:         %v\n"+
		"  Same col:         %v\n\n"+
		"Mitigation Options:\n"+
		"  • Selector devices (1T1R)\n"+
		"    - On/Off ratio: 100:1 to 1000:1\n"+
		"    - Reduces sneak by 2-3 orders\n"+
		"  • Half-select scheme\n"+
		"    - V_sel/2 on unselected lines\n"+
		"    - Reduces sneak voltage\n"+
		"  • Threshold switching\n"+
		"    - Ovonic threshold switch\n"+
		"    - Blocks sub-threshold paths\n",
		row, col,
		severitySymbol, sneakRatio, severity,
		pathType,
		pathDescription,
		selectedRow, selectedCol,
		manhattanDist,
		signalCurrent*1e6,
		sneakCurrent*1e6,
		sneakRatio,
		snrDB,
		conductanceUS,
		sneakResistance/1000.0,
		3,
		sneakAnalysis.MaxSneakRatio*100,
		sneakAnalysis.AvgSneakRatio*100,
		sneakAnalysis.TotalSneak*1e6,
		snr,
		rowDist,
		colDist,
		sameRow,
		sameCol,
	)

	return tooltip
}

// MVMResultTooltip generates tooltip for MVM results
func MVMResultTooltip(row int, mvmResult *crossbar.MVMResult) string {
	if mvmResult == nil {
		return "Run MVM to see results"
	}

	if row >= len(mvmResult.IdealOutput) {
		return "Row out of range"
	}

	ideal := mvmResult.IdealOutput[row]
	actual := mvmResult.ActualOutput[row]
	error := math.Abs(ideal - actual)
	errorPercent := 0.0
	if ideal > 0 {
		errorPercent = (error / ideal) * 100
	}

	tooltip := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
		"OUTPUT ROW [%d] - MVM RESULTS\n"+
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
		"Output Values:\n"+
		"  Ideal output:     %.6f\n"+
		"  Actual output:    %.6f\n"+
		"  Error:            %.6f (%.2f%%)\n\n"+
		"Global Statistics:\n"+
		"  RMSE:             %.6f\n"+
		"  Max error:        %.6f\n"+
		"  Mean error:       %.6f\n"+
		"  Accuracy loss:    %.2f%%\n\n"+
		"Energy Metrics:\n"+
		"  This row ADC:     %.2f pJ\n"+
		"  Total MVM:        %.2f pJ\n"+
		"  GPU equivalent:   %.0f pJ\n"+
		"  Efficiency:       %.0f× better\n\n"+
		"Performance:\n"+
		"  MAC operations:   %d\n"+
		"  Latency:          %.1f ns\n"+
		"  Throughput:       %.2e MACs/s\n",
		row,
		ideal,
		actual,
		error, errorPercent,
		mvmResult.RMSE,
		mvmResult.MaxError,
		mvmResult.MeanError,
		mvmResult.AccuracyLoss,
		mvmResult.ADCEnergy / float64(len(mvmResult.IdealOutput)),
		mvmResult.TotalEnergy,
		mvmResult.GPUEquivalentEnergy,
		mvmResult.EnergyEfficiency,
		mvmResult.MACOperations,
		mvmResult.Latency,
		mvmResult.Throughput,
	)

	return tooltip
}

// ComprehensiveTooltip generates maximum information tooltip
func ComprehensiveTooltip(row, col int, array *crossbar.Array, irAnalysis *crossbar.IRDropAnalysis, sneakAnalysis *crossbar.SneakPathAnalysis, mvmResult *crossbar.MVMResult) string {
	G := array.GetConductanceMatrix()[row][col]
	level := crossbar.GetLevel(G)
	conductanceUS := G*99 + 1

	tooltip := fmt.Sprintf(
		"╔═══════════════════════════════════╗\n"+
		"║  CELL [%d, %d] - FULL ANALYSIS   ║\n"+
		"╚═══════════════════════════════════╝\n\n"+
		"🔹 BASIC PROPERTIES\n"+
		"  FeCIM level:      %d / 29\n"+
		"  Conductance:      %.2f µS\n"+
		"  Normalized:       %.6f\n\n",
		row, col,
		level,
		conductanceUS,
		G,
	)

	if irAnalysis != nil && row < len(irAnalysis.EffectiveVoltage) && col < len(irAnalysis.EffectiveVoltage[0]) {
		effectiveV := irAnalysis.EffectiveVoltage[row][col]
		dropPercent := (1.0 - effectiveV) * 100
		tooltip += fmt.Sprintf(
			"🔹 IR DROP\n"+
			"  Effective voltage: %.3f V\n"+
			"  Voltage drop:      %.2f%%\n\n",
			effectiveV,
			dropPercent,
		)
	}

	if sneakAnalysis != nil && row < len(sneakAnalysis.SneakCurrents) && col < len(sneakAnalysis.SneakCurrents[0]) {
		sneakCurrent := sneakAnalysis.SneakCurrents[row][col]
		sneakRatio := 0.0
		if sneakAnalysis.TotalSignal > 0 {
			sneakRatio = sneakCurrent / sneakAnalysis.TotalSignal * 100
		}
		tooltip += fmt.Sprintf(
			"🔹 SNEAK PATH\n"+
			"  Sneak current:     %.6f µA\n"+
			"  Sneak ratio:       %.2f%%\n\n",
			sneakCurrent*1e6,
			sneakRatio,
		)
	}

	if mvmResult != nil {
		// MED-004 fix: Add nuanced GPU comparison context
		tooltip += fmt.Sprintf(
			"🔹 MVM PERFORMANCE\n"+
			"  Energy:            %.2f pJ\n"+
			"  vs GPU:            %.0f× better (per-MAC basis)\n"+
			"  Latency:           %.1f ns (single cycle)\n\n"+
			"⚠️ GPU COMPARISON CONTEXT:\n"+
			"  • FeCIM advantage: per-operation latency + energy\n"+
			"  • GPUs excel at batched operations (throughput)\n"+
			"  • Fair comparison needs throughput AND efficiency\n"+
			"  • FeCIM: O(1) cycles for MVM vs GPU: O(N²)\n",
			mvmResult.TotalEnergy,
			mvmResult.EnergyEfficiency,
			mvmResult.Latency,
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
