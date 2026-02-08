package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/keyboard"
)

// setupKeyboard initializes keyboard shortcuts using the shared keyboard manager.
func (ma *MNISTApp) setupKeyboard() {
	km := keyboard.NewManager(ma.window)

	// Register common handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			debug.Println("Save shortcut triggered (Ctrl+S)")
			// Future: Save network weights
			ma.updateStatus("Save: Feature not yet implemented")
		},
		keyboard.ActionExport: func() {
			debug.Println("Export shortcut triggered (Ctrl+E)")
			// Future: Export evaluation results
			ma.updateStatus("Export: Feature not yet implemented")
		},
		keyboard.ActionReset: func() {
			debug.Println("Reset shortcut triggered (Ctrl+R)")
			ma.clearCanvas()
		},
		keyboard.ActionPauseResume: func() {
			ma.toggleAutoDemo()
		},
		keyboard.ActionHelp: func() {
			ma.showKeyboardHelp()
		},
		keyboard.ActionNavigateLeft: func() {
			ma.loadRandomTestDigit()
		},
		keyboard.ActionNavigateRight: func() {
			ma.loadRandomTestDigit()
		},
		keyboard.ActionStepForward: func() {
			ma.loadRandomTestDigit()
		},
	})

	// Add module-specific shortcuts
	km.AddCustomShortcut("clear", fyne.KeyC, 0, "Clear canvas")
	km.AddCustomShortcut("random", fyne.KeyR, 0, "Load random test digit")
	km.AddCustomShortcut("evaluate", fyne.KeyE, 0, "Evaluate network")
	km.AddCustomShortcut("load", fyne.KeyL, 0, "Load test data")

	// Register the manager
	km.Register()

	// Handle additional module-specific keys
	ma.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		ma.handleKeyPress(ke)
	})
}

// handleKeyPress handles module-specific key events
func (ma *MNISTApp) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyC:
		// Clear canvas
		ma.clearCanvas()

	case fyne.KeyR:
		// Load random test digit
		ma.loadRandomTestDigit()

	case fyne.KeyE:
		// Evaluate network (without Ctrl modifier)
		go ma.evaluateNetwork()

	case fyne.KeyL:
		// Load test data
		go ma.loadTestData()

	case fyne.KeySpace:
		// Toggle auto demo
		ma.toggleAutoDemo()

	case fyne.KeySlash:
		// Show help
		ma.showKeyboardHelp()
	}
}

// clearCanvas clears the digit canvas and resets predictions
func (ma *MNISTApp) clearCanvas() {
	if ma.digitCanvas != nil {
		ma.digitCanvas.Clear()
	}
	if ma.predictionDisplay != nil {
		ma.predictionDisplay.SetPrediction(-1, 0)
	}
	if ma.operationLog != nil {
		ma.operationLog.Add("Canvas cleared")
	}
	ma.updateStatus("● IDLE | Canvas cleared")
}

// toggleAutoDemo toggles the automatic demonstration mode
func (ma *MNISTApp) toggleAutoDemo() {
	ma.autoDemoMu.Lock()
	isRunning := ma.autoDemo
	ma.autoDemoMu.Unlock()

	if isRunning {
		ma.stopAutoDemoLoop()
		if ma.demoModeSelect != nil {
			ma.demoModeSelect.SetSelected("Manual")
		}
		ma.updateStatus("Auto demo paused - Press Space to resume")
	} else {
		if ma.demoModeSelect != nil {
			ma.demoModeSelect.SetSelected("Auto Demo")
		}
		ma.updateStatus("Auto demo started - Press Space to pause")
	}
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (ma *MNISTApp) showKeyboardHelp() {
	helpText := `Keyboard Shortcuts:

Navigation:
  ←/→       Load random test digit
  ]         Step forward (load random digit)

Simulation:
  Space     Toggle auto demo pause/resume
  Ctrl+R    Clear canvas

Data:
  Ctrl+S    Save (future)
  Ctrl+E    Export (future)
  C         Clear canvas
  R         Load random test digit
  E         Evaluate network on test data
  L         Load test data

Demo Mode:
  Space     Toggle auto demo

Help:
  / or ?    Show this help dialog

Tips:
• Draw a digit on the canvas to see predictions
• Use R to load random MNIST test samples
• Press E to run full evaluation on test set
• Watch the network activations in the center panel`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(400, 380))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, ma.window)
	helpDialog.Show()
}
