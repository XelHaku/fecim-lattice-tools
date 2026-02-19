// Package gui provides Fyne-based GUI components for crossbar visualization.
// callbacks.go contains event handlers for heatmap cell interactions.
package gui

import (
	"fmt"

	"fecim-lattice-tools/shared/crossbar"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// syncSelection updates the app-level selection and syncs it to all heatmaps.
func (ca *CrossbarApp) syncSelection(row, col int) {
	ca.stateMu.Lock()
	ca.selectedRow = row
	ca.selectedCol = col
	ca.stateMu.Unlock()

	// Sync to all heatmaps
	if ca.conductanceHeatmap != nil {
		ca.conductanceHeatmap.SetSelection(row, col)
	}
	if ca.irDropHeatmap != nil {
		ca.irDropHeatmap.SetSelection(row, col)
	}
	if ca.sneakPathHeatmap != nil {
		ca.sneakPathHeatmap.SetSelection(row, col)
	}
	if ca.beforeAfterToggle != nil && ca.beforeAfterToggle.leftHeatmap != nil {
		ca.beforeAfterToggle.leftHeatmap.SetSelection(row, col)
		ca.beforeAfterToggle.rightHeatmap.SetSelection(row, col)
	}
}

// onCellTapped handles clicks on heatmap cells.
func (ca *CrossbarApp) onCellTapped(row, col int) {
	ca.modeIndicator.SetMode(int(DemoModeRead))

	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	matrix := ca.array.GetConductanceMatrix()
	value := matrix[row][col]
	level := crossbar.GetLevel(value)

	ca.levelIndicator.SetLevel(level)

	// Generate comprehensive tooltip
	tooltip := sharedwidgets.ConductanceTooltip(row, col, value, ca.array)

	// Display in stats label (formatted for readability)
	ca.statsLabel.SetText(tooltip)

	ca.updateStatus(fmt.Sprintf("READ | Cell [%d,%d] = Level %d/30 (%.2f µS)",
		row, col, level, value*99+1))
	ca.modeIndicator.SetMode(int(DemoModeIdle))
}

// onCellHover handles mouse hover over heatmap cells.
// M3 UX fix: Compact format with most important info first to avoid truncation
func (ca *CrossbarApp) onCellHover(row, col int, value float64) {
	// If mouse is out of bounds (row/col < 0), check if we have a pinned selection
	if row < 0 || col < 0 {
		ca.stateMu.RLock()
		selRow, selCol := ca.selectedRow, ca.selectedCol
		ca.stateMu.RUnlock()

		// UI-015: If a cell is selected, show its info instead of "Hover over..."
		if selRow >= 0 && selCol >= 0 {
			// Fetch the value from the matrix safely
			matrix := ca.array.GetConductanceMatrix()
			if selRow < len(matrix) && selCol < len(matrix[0]) {
				val := matrix[selRow][selCol]
				level := crossbar.GetLevel(val)
				conductanceUS := val*99 + 1
				ca.hoverInfoLabel.SetText(fmt.Sprintf(
					"Selected: [%d,%d] L%d │ %.1f µS",
					selRow, selCol, level, conductanceUS))
				return
			}
		}

		ca.hoverInfoLabel.SetText("Hover over cells for details")
		return
	}
	level := crossbar.GetLevel(value)
	conductanceUS := value*99 + 1

	// Compact format: position, level, conductance (most important first)
	ca.hoverInfoLabel.SetText(fmt.Sprintf(
		"[%d,%d] L%d │ %.1f µS",
		row, col, level, conductanceUS))
}

// onIRDropCellTapped handles clicks on IR Drop heatmap.
func (ca *CrossbarApp) onIRDropCellTapped(row, col int) {
	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	// Protected read of lastIRDropAnalysis and architecture
	ca.stateMu.RLock()
	analysis := ca.lastIRDropAnalysis
	arch := ca.architecture
	ca.stateMu.RUnlock()

	// Generate comprehensive IR drop tooltip with architecture
	tooltip := sharedwidgets.IRDropTooltipWithArch(row, col, analysis, ca.array, arch)
	ca.statsLabel.SetText(tooltip)

	// Update status with key info
	if analysis != nil && row < len(analysis.EffectiveVoltage) &&
		col < len(analysis.EffectiveVoltage[0]) {
		effectiveV := analysis.EffectiveVoltage[row][col]
		dropPercent := (1.0 - effectiveV) * 100
		ca.updateStatus(fmt.Sprintf("IR DROP | Cell [%d,%d]: %.3f V (%.1f%% drop)",
			row, col, effectiveV, dropPercent))

		// m3 UX fix: Update educational content for extreme cells
		isWorstCell := row == analysis.WorstCaseCell[0] && col == analysis.WorstCaseCell[1]
		if isWorstCell {
			ca.setEducationalContent("Worst IR Drop Cell",
				"This cell has the HIGHEST\nvoltage drop in the array.\n\n"+
					"Why? It's farthest from\nthe voltage drivers.\n\n"+
					"Impact:\n"+
					"• Reduced effective voltage\n"+
					"• Lower compute accuracy\n"+
					"• Current loss proportional\n\n"+
					"Mitigation:\n"+
					"• Hierarchical drivers\n"+
					"• Wider metal lines\n"+
					"• Tiled architecture")
		} else if dropPercent > 10 {
			ca.setEducationalContent("High IR Drop Cell",
				"This cell has significant\nvoltage drop (>10%).\n\n"+
					"Position matters: cells\nfar from drivers suffer\nmore voltage drop.\n\n"+
					"The drop follows Ohm's Law:\nV_drop = I × R_wire\n\n"+
					"Click the worst-case cell\nto see maximum impact.")
		}
	}
}

// onIRDropCellHover handles hover on IR Drop heatmap.
// M3 UX fix: Compact format with most important info first
func (ca *CrossbarApp) onIRDropCellHover(row, col int, value float64) {
	// If mouse is out of bounds, check for pinned selection
	if row < 0 || col < 0 {
		ca.stateMu.RLock()
		selRow, selCol := ca.selectedRow, ca.selectedCol
		analysis := ca.lastIRDropAnalysis
		ca.stateMu.RUnlock()

		if selRow >= 0 && selCol >= 0 && analysis != nil &&
			selRow < len(analysis.EffectiveVoltage) &&
			selCol < len(analysis.EffectiveVoltage[0]) {

			effectiveV := analysis.EffectiveVoltage[selRow][selCol]
			dropPercent := (1.0 - effectiveV) * 100
			ca.hoverInfoLabel.SetText(fmt.Sprintf(
				"Selected: [%d,%d] %.1f%% drop │ %.3fV",
				selRow, selCol, dropPercent, effectiveV))
			return
		}

		ca.hoverInfoLabel.SetText("Hover over cells for IR drop")
		return
	}

	// Protected read of lastIRDropAnalysis
	ca.stateMu.RLock()
	analysis := ca.lastIRDropAnalysis
	ca.stateMu.RUnlock()

	// Compact format: position, drop%, voltage (most important first)
	if analysis != nil && row < len(analysis.EffectiveVoltage) &&
		col < len(analysis.EffectiveVoltage[0]) {
		effectiveV := analysis.EffectiveVoltage[row][col]
		dropPercent := (1.0 - effectiveV) * 100

		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] %.1f%% drop │ %.3fV",
			row, col, dropPercent, effectiveV))
	} else {
		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] Run MVM for IR drop", row, col))
	}
}

// onSneakCellTapped handles clicks on Sneak Path heatmap.
func (ca *CrossbarApp) onSneakCellTapped(row, col int) {
	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	// Get selected target cell for sneak analysis (typically center)
	sneakTargetRow := ca.config.Rows / 2
	sneakTargetCol := ca.config.Cols / 2

	// Protected read of lastSneakAnalysis and architecture
	ca.stateMu.RLock()
	analysis := ca.lastSneakAnalysis
	arch := ca.architecture
	ca.stateMu.RUnlock()

	// Generate comprehensive sneak path tooltip with architecture
	tooltip := sharedwidgets.SneakPathTooltipWithArch(row, col, analysis, sneakTargetRow, sneakTargetCol, ca.array, arch)
	ca.statsLabel.SetText(tooltip)

	// Update status with key info
	if analysis != nil && row < len(analysis.SneakCurrents) &&
		col < len(analysis.SneakCurrents[0]) {
		sneakCurrent := analysis.SneakCurrents[row][col]
		sneakRatio := 0.0
		if analysis.TotalSignal > 0 {
			sneakRatio = sneakCurrent / analysis.TotalSignal * 100
		}
		// Cap display at 100% with note when sneak exceeds signal
		sneakDisplay := sneakRatio
		sneakNote := ""
		if sneakRatio > 100.0 {
			sneakDisplay = 100.0
			sneakNote = fmt.Sprintf(" [actual %.0f%%]", sneakRatio)
		}
		ca.updateStatus(fmt.Sprintf("SNEAK | Cell [%d,%d]: %.6f µA (%.2f%% of signal)%s",
			row, col, sneakCurrent*1e6, sneakDisplay, sneakNote))

		// m3 UX fix: Update educational content for specific cell types
		isTargetCell := row == sneakTargetRow && col == sneakTargetCol
		isRowNeighbor := row == sneakTargetRow && col != sneakTargetCol
		isColNeighbor := col == sneakTargetCol && row != sneakTargetRow

		if isTargetCell {
			ca.setEducationalContent("Target Cell",
				"This is the SELECTED cell\nfor sneak path analysis.\n\n"+
					"Ideal read: Only this cell's\ncurrent should be measured.\n\n"+
					"Reality: Current from OTHER\ncells sneaks through shared\nword/bit lines.\n\n"+
					"Sneak paths cause:\n"+
					"• Read errors\n"+
					"• Increased power\n"+
					"• Crosstalk noise")
		} else if isRowNeighbor {
			ca.setEducationalContent("Row Neighbor Cell",
				"Same ROW as target cell.\n\n"+
					"Shares the WORD LINE with\nthe target cell.\n\n"+
					"Sneak path: Current can flow\nthrough this cell when its\nbit line is partially selected.\n\n"+
					"1T1R architecture adds a\ntransistor to block this path.")
		} else if isColNeighbor {
			ca.setEducationalContent("Column Neighbor Cell",
				"Same COLUMN as target cell.\n\n"+
					"Shares the BIT LINE with\nthe target cell.\n\n"+
					"Sneak path: Current from this\ncell's word line can add to\nthe sensed output current.\n\n"+
					"Half-select voltage schemes\nhelp reduce this leakage.")
		} else if sneakRatio > 5 {
			ca.setEducationalContent("High Sneak Contributor",
				"This diagonal cell contributes\n>5% sneak current.\n\n"+
					"Diagonal cells create the\nworst sneak paths because\nthey connect through TWO\nintermediate cells.\n\n"+
					"Path: Target → Row neighbor\n→ This cell → Col neighbor\n→ Back to bit line\n\n"+
					"High conductance cells\ncontribute more sneak current.")
		}
	}
}

// onSneakCellHover handles hover on Sneak Path heatmap.
// M3 UX fix: Compact format with most important info first
func (ca *CrossbarApp) onSneakCellHover(row, col int, value float64) {
	// If mouse is out of bounds, check for pinned selection
	if row < 0 || col < 0 {
		ca.stateMu.RLock()
		selRow, selCol := ca.selectedRow, ca.selectedCol
		analysis := ca.lastSneakAnalysis
		ca.stateMu.RUnlock()

		if selRow >= 0 && selCol >= 0 && analysis != nil &&
			selRow < len(analysis.SneakCurrents) &&
			selCol < len(analysis.SneakCurrents[0]) {

			// Re-calculate sneak ratio for the pinned cell
			sneakRatio := 0.0
			if analysis.TotalSignal > 0 {
				sneakRatio = analysis.SneakCurrents[selRow][selCol] / analysis.TotalSignal * 100
			}
			sneakDisplay := sneakRatio
			overflowMark := ""
			if sneakRatio > 100.0 {
				sneakDisplay = 100.0
				overflowMark = "+"
			}

			// Recalculate context for pinned cell
			selectedRow := ca.config.Rows / 2
			selectedCol := ca.config.Cols / 2
			pathType := "diag"
			if selRow == selectedRow && selCol == selectedCol {
				pathType = "target"
			} else if selRow == selectedRow {
				pathType = "row"
			} else if selCol == selectedCol {
				pathType = "col"
			}

			ca.hoverInfoLabel.SetText(fmt.Sprintf(
				"Selected: [%d,%d] %.1f%%%s sneak │ %s",
				selRow, selCol, sneakDisplay, overflowMark, pathType))
			return
		}

		ca.hoverInfoLabel.SetText("Hover over cells for sneak path")
		return
	}

	// Get selected cell (center)
	selectedRow := ca.config.Rows / 2
	selectedCol := ca.config.Cols / 2

	// Protected read of lastSneakAnalysis
	ca.stateMu.RLock()
	analysis := ca.lastSneakAnalysis
	ca.stateMu.RUnlock()

	// Compact format: position, sneak%, path type (most important first)
	if analysis != nil && row < len(analysis.SneakCurrents) &&
		col < len(analysis.SneakCurrents[0]) {
		sneakRatio := 0.0
		if analysis.TotalSignal > 0 {
			sneakRatio = analysis.SneakCurrents[row][col] / analysis.TotalSignal * 100
		}

		// Cap display at 100% with indicator when sneak exceeds signal
		sneakDisplay := sneakRatio
		overflowMark := ""
		if sneakRatio > 100.0 {
			sneakDisplay = 100.0
			overflowMark = "+"
		}

		// Determine path type (compact)
		pathType := "diag"
		if row == selectedRow && col == selectedCol {
			pathType = "target"
		} else if row == selectedRow {
			pathType = "row"
		} else if col == selectedCol {
			pathType = "col"
		}

		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] %.1f%%%s sneak │ %s",
			row, col, sneakDisplay, overflowMark, pathType))
	} else {
		ca.hoverInfoLabel.SetText(fmt.Sprintf(
			"[%d,%d] Run MVM for sneak analysis", row, col))
	}
}

// refreshSelectedCellTooltip refreshes the tooltip for the currently selected cell.
// Call this after architecture changes or MVM recalculation to update displayed values.
func (ca *CrossbarApp) refreshSelectedCellTooltip() {
	// Get selected cell and architecture with mutex protection
	ca.stateMu.RLock()
	row, col := ca.selectedRow, ca.selectedCol
	arch := ca.architecture
	ca.stateMu.RUnlock()

	// Check if a cell is selected and within bounds
	if row < 0 || col < 0 || row >= ca.config.Rows || col >= ca.config.Cols {
		return
	}

	// Determine which tab is active and refresh the appropriate tooltip
	if ca.tabs == nil || ca.tabs.Selected() == nil {
		return
	}

	switch ca.getBaseTabName(ca.tabs.Selected().Text) {
	case "Conductance":
		matrix := ca.array.GetConductanceMatrix()
		if row < len(matrix) && col < len(matrix[0]) {
			value := matrix[row][col]
			tooltip := sharedwidgets.ConductanceTooltip(row, col, value, ca.array)
			ca.statsLabel.SetText(tooltip)
		}

	case "IR Drop":
		ca.stateMu.RLock()
		analysis := ca.lastIRDropAnalysis
		ca.stateMu.RUnlock()
		tooltip := sharedwidgets.IRDropTooltipWithArch(row, col, analysis, ca.array, arch)
		ca.statsLabel.SetText(tooltip)

	case "Sneak Paths":
		sneakTargetRow := ca.config.Rows / 2
		sneakTargetCol := ca.config.Cols / 2
		ca.stateMu.RLock()
		analysis := ca.lastSneakAnalysis
		ca.stateMu.RUnlock()
		tooltip := sharedwidgets.SneakPathTooltipWithArch(row, col, analysis, sneakTargetRow, sneakTargetCol, ca.array, arch)
		ca.statsLabel.SetText(tooltip)
	}

	// Force refresh to ensure UI updates
	if ca.statsLabel != nil {
		ca.statsLabel.Refresh()
	}
}
