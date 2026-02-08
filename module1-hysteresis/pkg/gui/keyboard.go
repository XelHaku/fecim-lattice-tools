package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/keyboard"
)

// setupKeyboard initializes keyboard shortcuts using the shared keyboard manager.
func (a *App) setupKeyboard() {
	km := keyboard.NewManager(a.mainWindow)

	// Register common handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			log.Info("Save shortcut triggered (Ctrl+S)")
			go a.exportPEData()
		},
		keyboard.ActionExport: func() {
			log.Info("Export shortcut triggered (Ctrl+E)")
			go a.exportPEData()
		},
		keyboard.ActionReset: func() {
			log.Info("Reset shortcut triggered (Ctrl+R)")
			a.resetSimulation()
		},
		keyboard.ActionPauseResume: func() {
			a.togglePause()
		},
		keyboard.ActionHelp: func() {
			a.showKeyboardHelp()
		},
		keyboard.ActionNavigateUp: func() {
			a.increaseTemperature()
		},
		keyboard.ActionNavigateDown: func() {
			a.decreaseTemperature()
		},
		keyboard.ActionNavigateLeft: func() {
			a.decreaseEField()
		},
		keyboard.ActionNavigateRight: func() {
			a.increaseEField()
		},
		keyboard.ActionIncrease: func() {
			a.increaseFrequency()
		},
		keyboard.ActionDecrease: func() {
			a.decreaseFrequency()
		},
		keyboard.ActionNextTab: func() {
			a.cycleWaveform()
		},
	})

	// Add module-specific shortcuts
	km.AddCustomShortcut("increase_e", fyne.KeyE, 0, "Increase E-field (Manual mode)")
	km.AddCustomShortcut("decrease_e", fyne.KeyD, 0, "Decrease E-field (Manual mode)")
	km.AddCustomShortcut("increase_temp", fyne.KeyT, 0, "Increase temperature by 25K")
	km.AddCustomShortcut("decrease_temp", fyne.KeyG, 0, "Decrease temperature by 25K")
	km.AddCustomShortcut("increase_freq", fyne.KeyF, 0, "Double frequency")
	km.AddCustomShortcut("decrease_freq", fyne.KeyV, 0, "Halve frequency")
	km.AddCustomShortcut("cycle_waveform", fyne.KeyW, 0, "Cycle waveform mode")

	// Register the manager
	km.Register()

	// Also handle legacy key events for module-specific keys
	a.mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		a.handleKeyPress(ke)
	})
}

// setupShortcuts registers custom keyboard shortcuts (legacy - kept for compatibility)
func (a *App) setupShortcuts() {
	// Register Ctrl+E for export
	ctrlE := &desktop.CustomShortcut{
		KeyName:  fyne.KeyE,
		Modifier: fyne.KeyModifierControl,
	}
	a.mainWindow.Canvas().AddShortcut(ctrlE, func(shortcut fyne.Shortcut) {
		log.Info("Export shortcut triggered (Ctrl+E)")
		go a.exportPEData()
	})

	// Register Ctrl+S for save (same as export for this module)
	ctrlS := &desktop.CustomShortcut{
		KeyName:  fyne.KeyS,
		Modifier: fyne.KeyModifierControl,
	}
	a.mainWindow.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		log.Info("Save shortcut triggered (Ctrl+S)")
		go a.exportPEData()
	})

	// Register Ctrl+R for reset
	ctrlR := &desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: fyne.KeyModifierControl,
	}
	a.mainWindow.Canvas().AddShortcut(ctrlR, func(shortcut fyne.Shortcut) {
		log.Info("Reset shortcut triggered (Ctrl+R)")
		a.resetSimulation()
	})
}

// resetSimulation resets the simulation state
func (a *App) resetSimulation() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.electricField = 0
	if a.useLKSolver() {
		resetP := a.lkDefaultPolarization()
		if a.lkSolver != nil {
			a.lkSolver.SetState(resetP)
			a.lkSolver.Time = 0
			a.polarization = a.lkSolver.GetState()
		} else {
			a.polarization = resetP
		}
	} else if a.preisach != nil {
		a.preisach.Reset()
		a.polarization = 0
		a.normalizedP = 0
	} else {
		a.polarization = 0
		a.normalizedP = 0
	}
	a.normalizedP = 0
	a.syncDiscreteLevelLocked()
	a.resetHistoryLocked()
	a.simTime = 0
	a.timeResAnimating = false
	a.timeResIndex = 0

	if a.eFieldSlider != nil {
		a.eFieldSlider.SetValue(0)
	}
}

// togglePause toggles the pause state
func (a *App) togglePause() {
	a.paused = !a.paused
	if a.paused {
		log.Info("Paused")
		if a.pauseBtn != nil {
			a.pauseBtn.SetText("Resume")
		}
	} else {
		log.Info("Resumed")
		if a.pauseBtn != nil {
			a.pauseBtn.SetText("Pause")
		}
	}
}

// increaseEField increases the E-field by 0.1*Ec (Manual mode only)
func (a *App) increaseEField() {
	if a.waveform != WaveformManual || a.eFieldSlider == nil {
		return
	}
	currentValue := a.eFieldSlider.Value
	newValue := currentValue + 0.1
	if newValue > 2.0 {
		newValue = 2.0
	}
	a.eFieldSlider.SetValue(newValue)
	log.Info("E-field increased to %.2f×Ec", newValue)
}

// decreaseEField decreases the E-field by 0.1*Ec (Manual mode only)
func (a *App) decreaseEField() {
	if a.waveform != WaveformManual || a.eFieldSlider == nil {
		return
	}
	currentValue := a.eFieldSlider.Value
	newValue := currentValue - 0.1
	if newValue < -2.0 {
		newValue = -2.0
	}
	a.eFieldSlider.SetValue(newValue)
	log.Info("E-field decreased to %.2f×Ec", newValue)
}

// increaseTemperature increases temperature by 25K
func (a *App) increaseTemperature() {
	a.mu.Lock()
	currentTemp := a.currentTemperature()
	newTemp := currentTemp + 25
	if newTemp > 700 {
		newTemp = 700
	}
	a.mu.Unlock()

	go func() {
		a.mu.Lock()
		a.onTemperatureChanged(newTemp)
		a.mu.Unlock()
	}()
	log.Info("Temperature increased to %.0f K", newTemp)
}

// decreaseTemperature decreases temperature by 25K
func (a *App) decreaseTemperature() {
	a.mu.Lock()
	currentTemp := a.currentTemperature()
	newTemp := currentTemp - 25
	if newTemp < 200 {
		newTemp = 200
	}
	a.mu.Unlock()

	go func() {
		a.mu.Lock()
		a.onTemperatureChanged(newTemp)
		a.mu.Unlock()
	}()
	log.Info("Temperature decreased to %.0f K", newTemp)
}

// increaseFrequency doubles the frequency
func (a *App) increaseFrequency() {
	a.mu.Lock()
	newFreq := a.frequency * 2.0
	const minFreq = 1e-9
	if newFreq < minFreq {
		newFreq = minFreq
	}
	a.frequency = newFreq
	a.resetHistoryLocked()
	a.simTime = 0
	a.mu.Unlock()
	log.Info("Frequency increased to %.3g Hz", newFreq)
}

// decreaseFrequency halves the frequency
func (a *App) decreaseFrequency() {
	a.mu.Lock()
	newFreq := a.frequency / 2.0
	const minFreq = 1e-9
	if newFreq < minFreq {
		newFreq = minFreq
	}
	a.frequency = newFreq
	a.resetHistoryLocked()
	a.simTime = 0
	a.mu.Unlock()
	log.Info("Frequency decreased to %.3g Hz", newFreq)
}

// cycleWaveform cycles to the next waveform mode
func (a *App) cycleWaveform() {
	a.mu.Lock()
	nextWaveform := (a.waveform + 1) % 5
	a.mu.Unlock()

	waveformNames := []string{"Manual", "Sine Wave", "Triangle Wave", "ISPP (Write/Read)", "Time-Resolved Switching"}
	selectedName := waveformNames[nextWaveform]

	if a.waveformSelect != nil {
		a.waveformSelect.SetSelected(selectedName)
	}
	log.Info("Waveform changed to %s", selectedName)
}

// handleKeyPress processes keyboard shortcuts
func (a *App) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyE:
		// Increase E-field by 0.1*Ec (Manual mode only)
		// Note: Ctrl+E is handled by setupShortcuts() via Canvas.AddShortcut
		a.increaseEField()

	case fyne.KeyD:
		// Decrease E-field by 0.1*Ec (Manual mode only)
		a.decreaseEField()

	case fyne.KeyT:
		// Increase temperature by 25K
		a.increaseTemperature()

	case fyne.KeyG:
		// Decrease temperature by 25K
		a.decreaseTemperature()

	case fyne.KeyF:
		// Double frequency
		a.increaseFrequency()

	case fyne.KeyV:
		// Halve frequency
		a.decreaseFrequency()

	case fyne.KeyW:
		// Cycle to next waveform
		a.cycleWaveform()

	case fyne.KeySpace:
		// Toggle pause
		a.togglePause()

	case fyne.KeyR:
		// Reset simulation (without Ctrl - legacy support)
		log.Info("Reset simulation")
		a.resetSimulation()

	case fyne.KeySlash:
		// Show keyboard help dialog
		a.showKeyboardHelp()
	}
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (a *App) showKeyboardHelp() {
	helpText := `Keyboard Shortcuts:

Navigation:
  ↑         Increase temperature by 25K
  ↓         Decrease temperature by 25K
  ←         Decrease E-field (Manual mode)
  →         Increase E-field (Manual mode)

Simulation:
  Space     Toggle pause/resume
  Ctrl+R    Reset simulation
  R         Reset simulation (quick)

Data:
  Ctrl+S    Save/Export P-E data
  Ctrl+E    Export P-E data

E-Field Control (Manual Mode Only):
  E         Increase E-field by 0.1×Ec
  D         Decrease E-field by 0.1×Ec

Temperature Control:
  T         Increase temperature by 25K
  G         Decrease temperature by 25K

Frequency Control:
  F         Double frequency
  V         Halve frequency (min 1e-9 Hz)
  +/=       Increase frequency
  -         Decrease frequency

Waveform:
  W         Cycle to next waveform
  Tab       Cycle waveform mode

Help:
  / or ?    Show this help dialog

Note: E and D keys only work in Manual mode.
Switch to Manual mode using the waveform selector.
Exported files are saved to the data/ directory with timestamps.`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(500, 450))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, a.mainWindow)
	helpDialog.Show()
}
