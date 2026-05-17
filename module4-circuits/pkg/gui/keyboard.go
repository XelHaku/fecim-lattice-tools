//go:build legacy_fyne

package gui

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/keyboard"
)

// setupKeyboard initializes keyboard shortcuts using the shared keyboard manager.
func (ca *CircuitsApp) setupKeyboard() {
	km := keyboard.NewManager(ca.window)

	// Register common handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			ca.exportData()
		},
		keyboard.ActionExport: func() {
			ca.exportData()
		},
		keyboard.ActionReset: func() {
			ca.resetArrayWeights()
		},
		keyboard.ActionPauseResume: func() {
			ca.toggleAnimation()
		},
		keyboard.ActionHelp: func() {
			ca.showKeyboardHelp()
		},
		keyboard.ActionNavigateUp: func() {
			ca.navigateLevel(1)
		},
		keyboard.ActionNavigateDown: func() {
			ca.navigateLevel(-1)
		},
		keyboard.ActionNavigateLeft: func() {
			ca.navigateCell(-1, 0)
		},
		keyboard.ActionNavigateRight: func() {
			ca.navigateCell(1, 0)
		},
		keyboard.ActionNextTab: func() {
			ca.nextTab()
		},
		keyboard.ActionPrevTab: func() {
			ca.prevTab()
		},
		keyboard.ActionIncrease: func() {
			ca.adjustDACBits(1)
		},
		keyboard.ActionDecrease: func() {
			ca.adjustDACBits(-1)
		},
	})

	// Add module-specific shortcuts
	km.AddCustomShortcut("program", fyne.KeyP, 0, "Program selected cell")
	km.AddCustomShortcut("read", fyne.KeyR, 0, "Read selected cell")
	km.AddCustomShortcut("compute", fyne.KeyC, 0, "Run compute operation")
	km.AddCustomShortcut("write_mode", fyne.KeyW, 0, "Switch to Write mode")
	km.AddCustomShortcut("read_mode", fyne.KeyD, 0, "Switch to Read mode")
	km.AddCustomShortcut("compute_mode", fyne.KeyM, 0, "Switch to Compute mode")
	km.AddCustomShortcut("animate", fyne.KeyA, 0, "Start/stop animation")
	km.AddCustomShortcut("zoom_in", fyne.KeyEqual, 0, "Zoom in array view")
	km.AddCustomShortcut("zoom_out", fyne.KeyMinus, 0, "Zoom out array view")
	km.AddCustomShortcut("fit_view", fyne.KeyF, 0, "Reset zoom to 100%")
	km.AddCustomShortcut("export_snapshot", fyne.KeyE, 0, "Export simulation snapshot")
	km.AddCustomShortcut("undo", fyne.KeyZ, 0, "Undo last operation")

	// Register the manager
	km.Register()

	// Handle additional module-specific keys
	ca.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		ca.handleKeyPress(ke)
	})
}

// handleKeyPress handles module-specific key events
func (ca *CircuitsApp) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyP:
		// Program cell
		ca.programSelectedCell()

	case fyne.KeyR:
		// Read cell (without Ctrl - that's reset)
		ca.readSelectedCell()

	case fyne.KeyC:
		// Compute operation
		ca.runCompute()

	case fyne.KeyW:
		// Switch to Write mode
		ca.setMode(ModeWrite)

	case fyne.KeyD:
		// Switch to Read mode
		ca.setMode(ModeRead)

	case fyne.KeyM:
		// Switch to Compute mode
		ca.setMode(ModeCompute)

	case fyne.KeyA:
		// Toggle animation
		ca.toggleAnimation()

	case fyne.KeySpace:
		// Toggle animation (same as A)
		ca.toggleAnimation()

	case fyne.KeyEqual:
		ca.adjustUnifiedZoom(1)

	case fyne.KeyMinus:
		ca.adjustUnifiedZoom(-1)

	case fyne.KeyF:
		ca.fitUnifiedZoom()

	case fyne.KeyE:
		ca.exportUnifiedSnapshot()

	case fyne.KeyZ:
		ca.undoUnifiedAction()

	case fyne.KeySlash:
		// Show help
		ca.showKeyboardHelp()
	}
}

// navigateLevel adjusts the target level for write operations
func (ca *CircuitsApp) navigateLevel(delta int) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	newLevel := ca.targetLevel + delta
	if newLevel < 0 {
		newLevel = 0
	}
	if newLevel >= ca.quantLevels {
		newLevel = ca.quantLevels - 1
	}
	ca.targetLevel = newLevel

	// Update slider if available
	if ca.opsWriteLevelSlider != nil {
		fyne.Do(func() {
			ca.opsWriteLevelSlider.SetValue(float64(newLevel))
		})
	}
}

// navigateCell moves cell selection
func (ca *CircuitsApp) navigateCell(dx, dy int) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	newCol := ca.selectedCol + dx
	newRow := ca.selectedRow + dy

	if newCol < 0 {
		newCol = 0
	}
	if newCol >= ca.arrayCols {
		newCol = ca.arrayCols - 1
	}
	if newRow < 0 {
		newRow = 0
	}
	if newRow >= ca.arrayRows {
		newRow = ca.arrayRows - 1
	}

	ca.selectedCol = newCol
	ca.selectedRow = newRow

	// Trigger canvas refresh to show new selection
	fyne.Do(func() {
		if ca.sharedArrayCanvas != nil {
			ca.sharedArrayCanvas.Refresh()
		}
	})
}

// adjustDACBits adjusts the DAC bits setting
func (ca *CircuitsApp) adjustDACBits(delta int) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	newBits := ca.dacBits + delta
	if newBits < 4 {
		newBits = 4
	}
	if newBits > 8 {
		newBits = 8
	}
	ca.dacBits = newBits

	// Recreate DAC with new bits
	// Note: This would require updating the DAC component
}

// nextTab switches to the next tab
func (ca *CircuitsApp) nextTab() {
	keyboard.SelectNextTab(ca.mainTabs)
}

// prevTab switches to the previous tab
func (ca *CircuitsApp) prevTab() {
	keyboard.SelectPrevTab(ca.mainTabs)
}

// programSelectedCell programs the currently selected cell
func (ca *CircuitsApp) programSelectedCell() {
	if ca.opsProgramBtn != nil && ca.currentMode == ModeWrite {
		ca.opsProgramBtn.OnTapped()
	}
}

// readSelectedCell reads the currently selected cell
func (ca *CircuitsApp) readSelectedCell() {
	if ca.opsReadBtn != nil && ca.currentMode == ModeRead {
		ca.opsReadBtn.OnTapped()
	}
}

// runCompute runs the compute operation
func (ca *CircuitsApp) runCompute() {
	if ca.opsComputeBtn != nil && ca.currentMode == ModeCompute {
		ca.opsComputeBtn.OnTapped()
		return
	}
	if ca.actionComputeBtn != nil && ca.currentMode == ModeCompute {
		ca.actionComputeBtn.OnTapped()
	}
}

func (ca *CircuitsApp) adjustUnifiedZoom(direction int) {
	if ca.zoomSlider == nil || ca.zoomSlider.Step <= 0 {
		return
	}
	ca.zoomSlider.SetValue(ca.zoomSlider.Value + float64(direction)*ca.zoomSlider.Step)
}

func (ca *CircuitsApp) fitUnifiedZoom() {
	if ca.actionFitBtn != nil {
		ca.actionFitBtn.OnTapped()
	}
}

func (ca *CircuitsApp) exportUnifiedSnapshot() {
	ca.exportSimulationData()
}

func (ca *CircuitsApp) undoUnifiedAction() {
	if ca.undoHistoryBtn != nil && !ca.undoHistoryBtn.Disabled() {
		ca.undoHistoryBtn.OnTapped()
	}
}

// toggleAnimation toggles the animation state
func (ca *CircuitsApp) toggleAnimation() {
	if ca.opsAnimateBtn != nil {
		ca.opsAnimateBtn.OnTapped()
	}
}

// setMode switches to the specified operation mode
func (ca *CircuitsApp) setMode(mode OperationMode) {
	ca.mu.Lock()
	ca.currentMode = mode
	ca.mu.Unlock()
	fyne.Do(func() {
		ca.updateModeButtons()
		ca.updateActionButtons()
	})
}

// resetArrayWeights resets the array to random weights
func (ca *CircuitsApp) resetArrayWeights() {
	if ca.opsResetBtn != nil {
		ca.opsResetBtn.OnTapped()
	}
}

// exportData exports the current data
func (ca *CircuitsApp) exportData() {
	// Export functionality - implement based on current mode
	if ca.operationsStatusLabel != nil {
		ca.operationsStatusLabel.SetText("Export: Feature coming soon")
	}
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (ca *CircuitsApp) showKeyboardHelp() {
	helpText := keyboard.FormatHelpMetadata(keyboard.HelpMetadata{
		Sections: []keyboard.ShortcutSection{
			{Title: "Navigation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "↑/↓", Description: "Adjust target level (Write mode)"}, {Key: "←/→", Description: "Navigate selected cell"}, {Key: "Tab", Description: "Next tab"}, {Key: "Shift+Tab", Description: "Previous tab"}}},
			{Title: "Operations", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Space", Description: "Toggle animation"}, {Key: "A", Description: "Toggle animation"}, {Key: "Ctrl+R", Description: "Reset array"}, {Key: "P", Description: "Program selected cell (Write mode)"}, {Key: "R", Description: "Read selected cell (Read mode)"}, {Key: "C", Description: "Run MVM operation (Compute mode)"}, {Key: "Z", Description: "Undo last operation"}, {Key: "E", Description: "Export simulation snapshot"}}},
			{Title: "Mode Switching", Shortcuts: []keyboard.ShortcutMetadata{{Key: "W", Description: "Switch to Write mode"}, {Key: "D", Description: "Switch to Read mode"}, {Key: "M", Description: "Switch to Compute mode"}}},
			{Title: "Configuration", Shortcuts: []keyboard.ShortcutMetadata{{Key: "+/=", Description: "Increase DAC bits"}, {Key: "-", Description: "Decrease DAC bits"}, {Key: "=", Description: "Zoom in array view"}, {Key: "-", Description: "Zoom out array view"}, {Key: "F", Description: "Reset zoom to 100%"}}},
			{Title: "Data", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+S", Description: "Save/Export data"}, {Key: "Ctrl+E", Description: "Export data"}}},
			{Title: "Help", Shortcuts: []keyboard.ShortcutMetadata{{Key: "/ or ?", Description: "Show this help dialog"}}},
		},
		Tips: []string{"Use W/D/M to quickly switch between modes", "Arrow keys navigate the array in all modes", "Up/Down adjusts write level in Write mode", "Press P to program, R to read, C to compute"},
	})
	keyboard.ShowHelpTextDialog(ca.window, "Keyboard Shortcuts", helpText, 420, 450)
}
