package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// handleKeyPress processes keyboard shortcuts
func (a *App) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyE:
		// Increase E-field by 0.1*Ec (Manual mode only)
		if a.waveform == WaveformManual {
			// Read slider value (UI access safe from main thread)
			currentValue := a.eFieldSlider.Value
			newValue := currentValue + 0.1
			if newValue > 2.0 {
				newValue = 2.0
			}
			a.eFieldSlider.SetValue(newValue)
			log.Info("E-field increased to %.2f×Ec", newValue)
		}

	case fyne.KeyD:
		// Decrease E-field by 0.1*Ec (Manual mode only)
		if a.waveform == WaveformManual {
			// Read slider value (UI access safe from main thread)
			currentValue := a.eFieldSlider.Value
			newValue := currentValue - 0.1
			if newValue < -2.0 {
				newValue = -2.0
			}
			a.eFieldSlider.SetValue(newValue)
			log.Info("E-field decreased to %.2f×Ec", newValue)
		}

	case fyne.KeyT:
		// Increase temperature by 25K
		a.mu.Lock()
		currentTemp := a.preisach.Temperature
		newTemp := currentTemp + 25
		if newTemp > 700 {
			newTemp = 700
		}
		a.mu.Unlock()

		// Handle temperature change with calibration (runs in background)
		go func() {
			a.mu.Lock()
			a.onTemperatureChanged(newTemp)
			a.mu.Unlock()
		}()
		log.Info("Temperature increased to %.0f K", newTemp)

	case fyne.KeyG:
		// Decrease temperature by 25K
		a.mu.Lock()
		currentTemp := a.preisach.Temperature
		newTemp := currentTemp - 25
		if newTemp < 200 {
			newTemp = 200
		}
		a.mu.Unlock()

		// Handle temperature change with calibration (runs in background)
		go func() {
			a.mu.Lock()
			a.onTemperatureChanged(newTemp)
			a.mu.Unlock()
		}()
		log.Info("Temperature decreased to %.0f K", newTemp)

	case fyne.KeyF:
		// Double frequency (up to 1.0 Hz)
		a.mu.Lock()
		newFreq := a.frequency * 2.0
		if newFreq > 1.0 {
			newFreq = 1.0
		}
		a.frequency = newFreq
		// Reset trail when frequency changes
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.mu.Unlock()
		log.Info("Frequency increased to %.2f Hz", newFreq)

	case fyne.KeyV:
		// Halve frequency (down to 0.01 Hz)
		a.mu.Lock()
		newFreq := a.frequency / 2.0
		if newFreq < 0.01 {
			newFreq = 0.01
		}
		a.frequency = newFreq
		// Reset trail when frequency changes
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.mu.Unlock()
		log.Info("Frequency decreased to %.2f Hz", newFreq)

	case fyne.KeyW:
		// Cycle to next waveform
		a.mu.Lock()
		nextWaveform := (a.waveform + 1) % 4
		a.mu.Unlock()

		// Trigger the same logic as waveformSelect.OnChanged
		waveformNames := []string{"Manual", "Sine Wave", "Triangle Wave", "Write/Read Demo"}
		selectedName := waveformNames[nextWaveform]

		// Call the select widget to trigger the OnChanged callback
		a.waveformSelect.SetSelected(selectedName)
		log.Info("Waveform changed to %s", selectedName)

	case fyne.KeySpace:
		// Toggle pause
		a.paused = !a.paused
		if a.paused {
			log.Info("Paused")
			a.pauseBtn.SetText("Resume")
		} else {
			log.Info("Resumed")
			a.pauseBtn.SetText("Pause")
		}

	case fyne.KeyR:
		// Reset simulation
		log.Info("Reset simulation")
		a.mu.Lock()
		a.preisach.Reset()
		a.electricField = 0
		a.polarization = 0
		a.normalizedP = 0
		a.discreteLevel = 15
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.mu.Unlock()
		a.eFieldSlider.SetValue(0)

	case fyne.KeySlash:
		// Show keyboard help dialog (both / and ? key)
		a.showKeyboardHelp()
	}
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (a *App) showKeyboardHelp() {
	helpText := `Keyboard Shortcuts:

E-Field Control (Manual Mode Only):
  E         Increase E-field by 0.1×Ec
  D         Decrease E-field by 0.1×Ec

Temperature Control:
  T         Increase temperature by 25K
  G         Decrease temperature by 25K

Frequency Control:
  F         Double frequency (max 1.0 Hz)
  V         Halve frequency (min 0.01 Hz)

Waveform & Simulation:
  W         Cycle to next waveform
  Space     Toggle pause/resume
  R         Reset simulation

Help:
  / or ?    Show this help dialog

Note: E and D keys only work in Manual mode.
Switch to Manual mode using the waveform selector.`

	// Create a scrollable label for the help text
	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(500, 400))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, a.mainWindow)
	helpDialog.Show()
}
