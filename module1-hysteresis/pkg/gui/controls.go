package gui

import (
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/ferroelectric"
)

func (a *App) createControlsPanel() fyne.CanvasObject {
	// E-field slider (no logging - too frequent)
	a.eFieldSlider = widget.NewSlider(-2, 2)
	a.eFieldSlider.Step = 0.01
	a.eFieldSlider.Value = 0
	a.eFieldSlider.OnChanged = func(v float64) {
		// Only accept slider input when in Manual mode AND not animating
		if a.waveform == WaveformManual && !a.manualAnimating {
			a.mu.Lock()
			a.electricField = v * a.material.Ec
			a.mu.Unlock()
		}
	}
	a.eFieldLabel = widget.NewLabel("E: 0.00 MV/cm")

	// Waveform selector
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "Write/Read Demo"}
	a.waveformSelect = widget.NewSelect(waveforms, func(s string) {
		log.Selection("Waveform", s)
		a.mu.Lock()
		defer a.mu.Unlock()
		switch s {
		case "Manual":
			a.waveform = WaveformManual
			a.autoMode = false
			if a.eFieldSlider != nil {
				a.eFieldSlider.Enable()
			}
		case "Sine Wave":
			a.waveform = WaveformSine
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
		case "Triangle Wave":
			a.waveform = WaveformTriangle
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
		case "Write/Read Demo":
			a.waveform = WaveformWriteReadDemo
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			// Reset write/read demo state with improved physics
			a.wrdPhase = 0
			a.wrdPhaseTimer = 0
			a.wrdTargetLevel = rand.Intn(30) + 1
			a.wrdStartLevel = a.discreteLevel + 1
			// Reset Dr. Tour demo metrics
			a.wrdTotalWrites = 0
			a.wrdSuccessWrites = 0
			a.wrdTotalEnergyfJ = 0
			a.wrdCycleEnergy = 0
			a.wrdBitsStored = 4.91 // log2(30)
			// Initialize debug log (requires material to be set)
			if a.material != nil {
				a.initDebugLog()
			}
		}
	})
	a.waveformSelect.SetSelected("Sine Wave")

	// Material selector
	matNames := []string{"Default HZO", "Optimized Superlattice", "FeCIM HZO"}
	a.materialSelect = widget.NewSelect(matNames, func(s string) {
		log.Selection("Material", s)
		var idx int
		switch s {
		case "Default HZO":
			idx = 0
		case "Optimized Superlattice":
			idx = 1
		case "FeCIM HZO":
			idx = 2
		}
		a.mu.Lock()
		a.matIndex = idx
		a.material = a.materials[idx]
		a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 30)
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.plot.SetBounds(a.material.Ec*2.5, a.material.Ps*1.2)
		a.mu.Unlock()
	})
	a.materialSelect.SetSelected("Default HZO")

	// Pause/Resume button
	a.pauseBtn = widget.NewButton("Pause", func() {
		a.paused = !a.paused
		if a.paused {
			log.Button("Pause (now paused)")
			a.pauseBtn.SetText("Resume")
		} else {
			log.Button("Resume (now running)")
			a.pauseBtn.SetText("Pause")
		}
	})

	// Reset button
	resetBtn := widget.NewButton("Reset", func() {
		log.Button("Reset")
		a.mu.Lock()
		a.preisach.Reset()
		a.electricField = 0
		a.polarization = 0
		a.normalizedP = 0
		a.discreteLevel = 15
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.eFieldSlider.SetValue(0)
		a.mu.Unlock()
	})

	// ELI5 (Explain Like I'm 5) button
	eli5Btn := widget.NewButton("ELI5", func() {
		log.Button("ELI5")
		a.showELI5Dialog()
	})
	eli5Btn.Importance = widget.LowImportance

	// Frequency slider
	freqSlider := widget.NewSlider(0.01, 1.0)
	freqSlider.Step = 0.01
	freqSlider.Value = 0.5
	freqLabel := widget.NewLabel("Freq: 0.50 Hz")
	freqSlider.OnChanged = func(v float64) {
		log.SliderChange("Frequency", v)
		a.mu.Lock()
		a.frequency = v
		// Reset trail when frequency changes
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0
		a.mu.Unlock()
		freqLabel.SetText(fmt.Sprintf("Freq: %.2f Hz", v))
	}

	// Temperature slider (Dr. Tour recommendation: show HZO's high Tc advantage)
	tempSlider := widget.NewSlider(200, 700)
	tempSlider.Step = 25
	tempSlider.Value = 300 // Room temperature (300K = 27°C)
	tempLabel := widget.NewLabel("T: 300 K (27°C)")
	tempSlider.OnChanged = func(v float64) {
		log.SliderChange("Temperature", v)
		a.mu.Lock()
		// Update Preisach model temperature
		a.preisach.SetTemperature(v)
		a.mu.Unlock()
		celsius := v - 273
		tcRatio := v / a.material.CurieTemp * 100
		tempLabel.SetText(fmt.Sprintf("T: %.0f K (%.0f°C) [%.0f%% Tc]", v, celsius, tcRatio))
	}

	// Trail length slider
	trailSlider := widget.NewSlider(50, 2000)
	trailSlider.Step = 50
	trailSlider.Value = float64(a.maxHistory)
	trailLabel := widget.NewLabel(fmt.Sprintf("Trail: %d", a.maxHistory))
	trailSlider.OnChanged = func(v float64) {
		log.SliderChange("TrailLength", v)
		a.mu.Lock()
		a.maxHistory = int(v)
		// Immediately trim if history exceeds new max
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[len(a.eHistory)-a.maxHistory:]
			a.pHistory = a.pHistory[len(a.pHistory)-a.maxHistory:]
		}
		a.mu.Unlock()
		trailLabel.SetText(fmt.Sprintf("Trail: %d", int(v)))
	}

	// Compact layout
	return container.NewVBox(
		a.materialSelect,
		a.waveformSelect,
		a.eFieldLabel,
		a.eFieldSlider,
		freqLabel,
		freqSlider,
		tempLabel,
		tempSlider,
		trailLabel,
		trailSlider,
		container.NewHBox(a.pauseBtn, resetBtn, eli5Btn),
		widget.NewLabel("← Click level bar"),
	)
}

