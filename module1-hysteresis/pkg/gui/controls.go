package gui

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func (a *App) createControlsPanel() fyne.CanvasObject {
	// E-field slider (no logging - too frequent)
	a.eFieldSlider = widget.NewSlider(-1.5, 1.5)
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
			a.wrdTargetLevel = rand.Intn(a.numLevels) + 1
			a.wrdStartLevel = a.discreteLevel + 1
			// Reset Dr. Tour demo metrics
			a.wrdTotalWrites = 0
			a.wrdSuccessWrites = 0
			a.wrdTotalEnergyfJ = 0
			a.wrdCycleEnergy = 0
			a.wrdBitsStored = math.Log2(float64(a.numLevels))
			// Initialize debug log (requires material to be set)
			if a.material != nil {
				a.initDebugLog()
			}
		}
	})
	a.waveformSelect.SetSelected("Write/Read Demo")

	// Material selector - dynamically build from AllMaterials()
	matNames := make([]string, len(a.materials))
	for i, m := range a.materials {
		matNames[i] = m.Name
	}
	a.materialSelect = widget.NewSelect(matNames, func(s string) {
		log.Selection("Material", s)
		// Find selected material by name
		var idx int
		for i, m := range a.materials {
			if m.Name == s {
				idx = i
				break
			}
		}
		a.mu.Lock()
		a.matIndex = idx
		a.material = a.materials[idx]
		// Use fixed high-resolution grid (50) for physics accuracy, independent of quantization levels
		a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 50)
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.plot.SetBounds(a.material.Ec*1.5, a.material.Ps*1.2)
		a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
		// Mark calibration as stale (new material needs recalibration)
		a.calibrated = false

		// Update number of levels based on material's capability
		newLevels := a.material.GetNumLevels()
		if newLevels != a.numLevels {
			a.numLevels = newLevels
			// Update level indicator and cell visualizer
			if a.levelIndicator != nil {
				a.levelIndicator.SetNumLevels(newLevels)
			}
			if a.cellViz != nil {
				a.cellViz.SetNumLevels(newLevels)
			}
			// Reset discrete level to middle of new range
			a.discreteLevel = newLevels / 2
			// Reset target levels to be within new bounds
			a.wrdTargetLevel = newLevels / 2
			a.manualTargetLevel = newLevels / 2
			// Resize calibration arrays
			a.calibrationUp = make([]float64, newLevels)
			a.calibrationDown = make([]float64, newLevels)
		}
		a.mu.Unlock()

		// Update levels entry and label in UI thread
		fyne.Do(func() {
			if a.levelsEntry != nil {
				a.levelsEntry.SetText(fmt.Sprintf("%d", newLevels))
			}
			if a.levelsLabel != nil {
				a.levelsLabel.SetText(fmt.Sprintf("Levels: %d (%.1f bits)", newLevels, math.Log2(float64(newLevels))))
			}
		})

		// Recalibrate for new material at current temperature (background to not block UI)
		go func() {
			a.mu.Lock()
			// Clear old calibration cache for new material
			a.tempCalibrations = make(map[int]*TempCalibration)
			// Calibrate at current temperature
			currentTemp := a.preisach.Temperature
			a.calibrateLevelsAtTemperature(currentTemp)
			if err := a.saveCalibration(); err != nil {
				log.Printf("Warning: failed to save calibration: %v", err)
			}
			a.mu.Unlock()
		}()
	})
	a.materialSelect.SetSelected(a.materials[0].Name)

	// Levels selector (2-256 levels)
	a.levelsLabel = widget.NewLabel(fmt.Sprintf("Levels: %d (%.1f bits)", a.numLevels, math.Log2(float64(a.numLevels))))
	a.levelsEntry = widget.NewEntry()
	a.levelsEntry.SetText(fmt.Sprintf("%d", a.numLevels))
	a.levelsEntry.SetPlaceHolder("2-256")
	a.levelsEntry.OnChanged = func(s string) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		if n < 2 {
			n = 2
		}
		if n > 256 {
			n = 256
		}
		log.Selection("Levels", fmt.Sprintf("%d", n))

		a.mu.Lock()
		if n != a.numLevels {
			a.numLevels = n
			// NOTE: Do NOT recreate Preisach model - its grid size is for physics
			// accuracy, not quantization levels. We only change the output quantization.

			// Update level indicator and cell visualizer
			if a.levelIndicator != nil {
				a.levelIndicator.SetNumLevels(n)
			}
			if a.cellViz != nil {
				a.cellViz.SetNumLevels(n)
			}
			// Reset discrete level to middle of new range
			a.discreteLevel = n / 2
			// Reset target levels to be within new bounds
			a.wrdTargetLevel = n / 2
			a.wrdStartLevel = n / 2
			a.manualTargetLevel = n / 2
			// Mark calibration as stale (need new level->field mapping)
			a.calibrated = false
		}
		a.mu.Unlock()

		// Update label
		bits := math.Log2(float64(n))
		a.levelsLabel.SetText(fmt.Sprintf("Levels: %d (%.1f bits)", n, bits))

		// Recalibrate in background at current temperature (new quantization mapping)
		go func() {
			a.mu.Lock()
			// Clear old calibration cache for new level count
			a.tempCalibrations = make(map[int]*TempCalibration)
			// Calibrate at current temperature
			currentTemp := a.preisach.Temperature
			a.calibrateLevelsAtTemperature(currentTemp)
			if err := a.saveCalibration(); err != nil {
				log.Printf("Warning: failed to save calibration: %v", err)
			}
			a.mu.Unlock()
		}()
	}

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

		// Update temperature with calibration handling (runs in background if recalibration needed)
		go func() {
			a.mu.Lock()
			a.onTemperatureChanged(v)
			// Get plot markers with temperature-corrected Ec and Pr
			effEc := a.preisach.GetEffectiveEc()
			effPr := a.preisach.GetEffectivePr()
			a.mu.Unlock()

			// Update plot markers (outside lock, uses fyne.Do internally)
			a.plot.SetMaterialParams(effEc, effPr)
		}()

		// Update label immediately (don't wait for calibration)
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
		a.levelsLabel,
		a.levelsEntry,
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

