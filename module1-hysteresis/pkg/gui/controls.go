package gui

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
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
	// Mode label shows whether slider is in manual or auto mode
	a.eFieldModeLabel = widget.NewLabel("AUTO")
	a.eFieldModeLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Waveform selector
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "Write/Read Demo", "Time-Resolved Switching"}
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
			if a.eFieldModeLabel != nil {
				a.eFieldModeLabel.SetText("MANUAL")
			}
			if a.levelIndicator != nil {
				a.levelIndicator.SetInteractive(true)
			}
			// Lazy calibration: run on first use of manual mode
			if a.needsCalibration {
				log.Printf("Running deferred calibration (manual mode activated)")
				a.calibrateLevels()
				if err := a.saveCalibration(); err != nil {
					log.Printf("Warning: failed to save calibration: %v", err)
				}
				a.needsCalibration = false
			}
		case "Sine Wave":
			a.waveform = WaveformSine
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			if a.eFieldModeLabel != nil {
				a.eFieldModeLabel.SetText("AUTO")
			}
			if a.levelIndicator != nil {
				a.levelIndicator.SetInteractive(false)
			}
		case "Triangle Wave":
			a.waveform = WaveformTriangle
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			if a.eFieldModeLabel != nil {
				a.eFieldModeLabel.SetText("AUTO")
			}
			if a.levelIndicator != nil {
				a.levelIndicator.SetInteractive(false)
			}
		case "Write/Read Demo":
			a.waveform = WaveformWriteReadDemo
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			if a.eFieldModeLabel != nil {
				a.eFieldModeLabel.SetText("AUTO")
			}
			if a.levelIndicator != nil {
				a.levelIndicator.SetInteractive(false)
			}
			// Lazy calibration: run on first use of WRD mode
			if a.needsCalibration {
				log.Printf("Running deferred calibration (Write/Read Demo activated)")
				a.calibrateLevels()
				if err := a.saveCalibration(); err != nil {
					log.Printf("Warning: failed to save calibration: %v", err)
				}
				a.needsCalibration = false
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
			// Initialize phase tracking for first cycle
			a.wrdResetStartP = a.polarization * 100 // Current polarization in µC/cm²
			// Initialize debug log (requires material to be set)
			if a.material != nil {
				a.initDebugLog()
			}
		case "Time-Resolved Switching":
			a.waveform = WaveformTimeResolved
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			if a.eFieldModeLabel != nil {
				a.eFieldModeLabel.SetText("AUTO")
			}
			if a.levelIndicator != nil {
				a.levelIndicator.SetInteractive(false)
			}
			// Reset time-resolved animation state for fresh start
			a.timeResAnimating = false
			a.timeResIndex = 0
			a.timeResDataTimes = nil
			a.timeResDataPols = nil
			a.timeResDataSwitch = nil
		}
	})
	a.waveformSelect.SetSelected("Write/Read Demo")

	// Material button - shows current material, opens picker on click
	a.materialBtn = widget.NewButton(a.material.Name, func() {
		sharedwidgets.ShowMaterialPicker(a.mainWindow, a.getCurrentMaterialID(), func(id string, mat *physics.Material) {
			a.onMaterialPickerSelected(id, mat)
		})
	})

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

		// Run calibration immediately for new level count
		a.mu.Lock()
		a.tempCalibrations = make(map[int]*TempCalibration)
		log.Printf("Running calibration for %d levels...", n)
		a.calibrateLevelsAtTemperature(300)
		if err := a.saveCalibration(); err != nil {
			log.Printf("Warning: failed to save calibration: %v", err)
		}
		a.needsCalibration = false
		a.mu.Unlock()
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
		// Reset physics model
		a.preisach.Reset()
		a.electricField = 0
		a.polarization = 0
		a.normalizedP = 0
		a.discreteLevel = a.numLevels / 2 // Reset to middle of current range

		// Reset trail history
		a.eHistory = a.eHistory[:0]
		a.pHistory = a.pHistory[:0]
		a.simTime = 0

		// Reset WRD (Write/Read Demo) state
		a.wrdPhase = 0
		a.wrdPhaseTimer = 0
		a.wrdTargetLevel = a.numLevels / 2
		a.wrdStartLevel = a.numLevels / 2
		a.wrdReadLevel = 0
		a.wrdRetryCount = 0
		a.wrdTotalWrites = 0
		a.wrdSuccessWrites = 0
		a.wrdTotalEnergyfJ = 0
		a.wrdCycleEnergy = 0

		// Reset manual mode animation state
		a.manualAnimating = false
		a.manualTargetLevel = a.numLevels / 2
		a.manualPhase = 0
		a.manualPhaseTime = 0

		// Reset Time-Resolved animation state
		a.timeResAnimating = false
		a.timeResIndex = 0
		a.timeResDataTimes = nil
		a.timeResDataPols = nil
		a.timeResDataSwitch = nil

		// Clear log entries
		a.logEntries = a.logEntries[:0]
		a.lastLogPhase = -1
		a.mu.Unlock()

		// Update UI elements (outside lock to avoid deadlock with fyne.Do)
		fyne.Do(func() {
			a.eFieldSlider.SetValue(0)
			if a.logText != nil {
				a.logText.SetText("")
			}
		})
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
	}

	// Stress slider (Phase 4.1: Electrostriction control)
	// stressSlider already defined above in previous edit?? No, I replaced it.
	// I need to be careful. The context lines show I overwrote freqSlider.OnChanged.

	// Let's rewrite the block around 		freqLabel.SetText(fmt.Sprintf("Freq: %.2f Hz", v))
	}

	// Stress slider (Phase 4.1: Electrostriction control)
	stressSlider.OnChanged = func(v float64) {
		log.SliderChange("Stress", v)
		a.mu.Lock()
		// Pass stress to physics engine
		a.preisach.SetStress(v) // v is in GPa
		a.mu.Unlock()
		stressLabel.SetText(fmt.Sprintf("Stress: %.1f GPa", v))
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
			// Capture previous temperature before change
			previousTemp := a.preisach.Temperature
			a.onTemperatureChanged(v)
			// Get current temperature after change
			currentTemp := a.preisach.Temperature

			// Clear history if temperature changed significantly (>25K)
			if math.Abs(currentTemp-previousTemp) > 25 {
				a.eHistory = a.eHistory[:0]
				a.pHistory = a.pHistory[:0]
			}

			// Get plot markers with temperature-corrected Ec and nominal Pr
			effEc := a.preisach.GetEffectiveEc()
			// Use material's nominal Pr (not GetEffectivePr which recalculates from current state)
			effPr := a.material.Pr
			a.mu.Unlock()

			// Only update material markers (Ec/Pr lines), NOT the axis bounds
			// The X-axis should stay fixed - only the marker positions change
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
		a.materialBtn,
		a.waveformSelect,
		a.levelsLabel,
		a.levelsEntry,
		container.NewHBox(a.eFieldLabel, a.eFieldModeLabel),
		a.eFieldSlider,
		freqLabel,
		freqSlider,
		tempLabel,
		tempSlider,
		trailLabel,
		trailSlider,
		container.NewHBox(a.pauseBtn, resetBtn, eli5Btn),
		stressLabel,
		stressSlider,
	)
}

// getCurrentMaterialID returns the ID of the currently selected material.
func (a *App) getCurrentMaterialID() string {
	// Map material names to IDs
	nameToID := map[string]string{
		"HZO (Si-doped)":                        "default_hzo",
		"FeCIM HZO":                             "fecim_hzo",
		"FeCIM HZO (TARGET - NOT DEMONSTRATED)": "fecim_hzo_target",
		"Literature Superlattice (Cheema 2020)": "literature_superlattice",
		"Cryogenic HZO (4K)":                    "cryogenic_hzo",
		"HZO Standard (32 states)":              "hzo_standard_32",
		"HZO FTJ (140 states)":                  "hzo_ftj_140",
		"AlScN (8-16 states)":                   "alscn",
	}

	if a.material != nil {
		if id, ok := nameToID[a.material.Name]; ok {
			return id
		}
	}
	return "default_hzo"
}

// onMaterialPickerSelected handles selection from the material picker dialog.
func (a *App) onMaterialPickerSelected(materialID string, physMat *physics.Material) {
	if physMat == nil {
		return
	}

	log.Printf("Material picker selected: %s (%s)", physMat.Name, materialID)

	// Convert physics.Material to ferroelectric.HZOMaterial
	cfg, err := physics.Load()
	if err != nil {
		log.Printf("Error loading physics config: %v", err)
		return
	}
	hzoMat := ferroelectric.MaterialFromConfig(physMat, cfg)

	// Find index in materials list (for compatibility with existing dropdown)
	var newIdx int
	for i, m := range a.materials {
		if m.Name == hzoMat.Name {
			newIdx = i
			break
		}
	}

	a.mu.Lock()
	// Capture temperature before material change
	savedTemp := a.preisach.Temperature

	a.matIndex = newIdx
	a.material = hzoMat
	// Use fixed high-resolution grid for physics accuracy
	a.preisach = ferroelectric.NewPreisachModel(a.material)

	// Restore temperature
	a.preisach.SetTemperature(savedTemp)

	// Clear history for new material
	a.eHistory = a.eHistory[:0]
	a.pHistory = a.pHistory[:0]

	// Reset simulation state
	a.electricField = 0
	a.polarization = 0
	a.normalizedP = 0
	a.simTime = 0

	// Reset time-resolved animation state
	a.timeResAnimating = false
	a.timeResIndex = 0
	a.timeResDataTimes = nil
	a.timeResDataPols = nil
	a.timeResDataSwitch = nil

	// Reset write/read demo state
	a.wrdPhase = 0
	a.wrdPhaseTimer = 0
	a.wrdTotalWrites = 0
	a.wrdSuccessWrites = 0
	a.wrdTotalEnergyfJ = 0
	a.wrdCycleEnergy = 0

	// Get temperature-corrected Ec and nominal Pr
	effEc := a.preisach.GetEffectiveEc()
	// Use material's nominal Pr (not GetEffectivePr which recalculates from current state)
	effPr := a.material.Pr

	// Update number of levels based on material
	newLevels := a.material.GetNumLevels()
	a.numLevels = newLevels

	// Initialize ISPP calculator with new material parameters
	// a.isppCalc = sharedphysics.NewISPPCalculator(effEc, newLevels)
	// Update Adaptive ISPP
	if a.adaptiveISPP != nil {
		a.adaptiveISPP = sharedphysics.NewAdaptiveISPP(a.lkSolver, a.material)
	}
	// Update level indicator and cell visualizer
	if a.levelIndicator != nil {
		a.levelIndicator.SetNumLevels(newLevels)
	}
	if a.cellViz != nil {
		a.cellViz.SetNumLevels(newLevels)
	}
	// Reset discrete level to middle of new range
	a.discreteLevel = newLevels / 2
	// Reset target levels
	a.wrdTargetLevel = newLevels / 2
	a.wrdStartLevel = newLevels / 2
	a.manualTargetLevel = newLevels / 2
	// Update bits stored
	a.wrdBitsStored = math.Log2(float64(newLevels))
	// Resize calibration arrays
	a.calibrationUp = make([]float64, newLevels)
	a.calibrationDown = make([]float64, newLevels)

	// Mark calibration as stale
	a.calibrated = false
	a.mu.Unlock()

	// Update UI directly (we're already on the UI thread from dialog callback)
	// Update plot bounds and markers
	a.plot.SetBounds(effEc*2.5, effPr*1.2)
	a.plot.SetMaterialParams(effEc, effPr)
	a.plot.SetData(nil, nil, 0, 0)
	a.plot.Refresh()

	// Reset E-field slider
	if a.eFieldSlider != nil {
		a.eFieldSlider.SetValue(0)
	}

	// Update material button text
	if a.materialBtn != nil {
		a.materialBtn.SetText(hzoMat.Name)
	}

	// Update levels entry and label
	if a.levelsEntry != nil {
		a.levelsEntry.SetText(fmt.Sprintf("%d", newLevels))
	}
	if a.levelsLabel != nil {
		a.levelsLabel.SetText(fmt.Sprintf("Levels: %d (%.1f bits)", newLevels, math.Log2(float64(newLevels))))
	}

	// Reset level indicator to middle
	if a.levelIndicator != nil {
		a.levelIndicator.SetLevel(newLevels / 2)
	}

	// Load or run calibration for new material
	a.tempCalibrations = make(map[int]*TempCalibration)
	if !a.loadCalibration() {
		log.Printf("Running calibration for %s...", a.material.Name)
		a.calibrateLevelsAtTemperature(300)
		if err := a.saveCalibration(); err != nil {
			log.Printf("Warning: failed to save calibration: %v", err)
		}
	}
	a.needsCalibration = false
}
