package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

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
	if ca.mainTabs == nil {
		return
	}
	items := ca.mainTabs.Items
	if len(items) == 0 {
		return
	}
	currentIdx := 0
	for i, item := range items {
		if item == ca.mainTabs.Selected() {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(items)
	ca.mainTabs.Select(items[nextIdx])
}

// prevTab switches to the previous tab
func (ca *CircuitsApp) prevTab() {
	if ca.mainTabs == nil {
		return
	}
	items := ca.mainTabs.Items
	if len(items) == 0 {
		return
	}
	currentIdx := 0
	for i, item := range items {
		if item == ca.mainTabs.Selected() {
			currentIdx = i
			break
		}
	}
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(items) - 1
	}
	ca.mainTabs.Select(items[prevIdx])
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
	helpText := `Keyboard Shortcuts:

Navigation:
  ↑/↓       Adjust target level (Write mode)
  ←/→       Navigate selected cell
  Tab       Next tab
  Shift+Tab Previous tab

Operations:
  Space     Toggle animation
  A         Toggle animation
  Ctrl+R    Reset array
  P         Program selected cell (Write mode)
  R         Read selected cell (Read mode)
  C         Run MVM operation (Compute mode)
  Z         Undo last operation
  E         Export simulation snapshot

Mode Switching:
  W         Switch to Write mode
  D         Switch to Read mode
  M         Switch to Compute mode

Configuration:
  +/=       Increase DAC bits
  -         Decrease DAC bits
  =         Zoom in array view
  -         Zoom out array view
  F         Reset zoom to 100%

Data:
  Ctrl+S    Save/Export data
  Ctrl+E    Export data

Help:
  / or ?    Show this help dialog

Tips:
• Use W/D/M to quickly switch between modes
• Arrow keys navigate the array in all modes
• Up/Down adjusts write level in Write mode
• Press P to program, R to read, C to compute`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(420, 450))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, ca.window)
	helpDialog.Show()
}
