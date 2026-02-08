package gui

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

const rangeCalibrationDebounce = 500 * time.Millisecond

func (a *App) rangeFracLabelText(frac float64) string {
	frac = clampRangeFrac(frac)
	a.mu.RLock()
	ps := 0.0
	if a.material != nil {
		ps = a.material.Ps
	}
	a.mu.RUnlock()
	if ps > 0 {
		effective := ps * frac * 100
		return fmt.Sprintf("Target Range: %.1f%% (±%.1f µC/cm²)", frac*100, effective)
	}
	return fmt.Sprintf("Target Range: %.1f%%", frac*100)
}

func (a *App) scheduleRangeCalibration() {
	if a == nil {
		return
	}
	if a.wrdRangeTimer != nil {
		a.wrdRangeTimer.Stop()
	}
	a.wrdRangeTimer = time.AfterFunc(rangeCalibrationDebounce, func() {
		a.mu.Lock()
		if a.material == nil || !a.needsCalibration {
			a.mu.Unlock()
			return
		}
		temp := a.currentTemperature()
		if log != nil {
			log.Printf("Running calibration for target range %.2f...", a.wrdRangeFrac)
		}
		a.tempCalibrations = make(map[int]*TempCalibration)
		a.calibrateLevelsAtTemperature(temp)
		if err := a.saveCalibration(); err != nil {
			if log != nil {
				log.Printf("Warning: failed to save calibration: %v", err)
			}
		}
		a.needsCalibration = false
		a.mu.Unlock()
	})
}

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
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "ISPP (Write/Read)", "Time-Resolved Switching"}
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
		case "ISPP (Write/Read)":
			a.waveform = WaveformWriteReadDemo
			a.autoMode = true
			if a.eFieldSlider != nil {
				a.eFieldSlider.Disable()
			}
			if a.eFieldModeLabel != nil {
				a.eFieldModeLabel.SetText("AUTO")
			}
			if a.levelIndicator != nil {
				a.levelIndicator.SetInteractive(true) // Allow manual target selection
			}
			// Lazy calibration: run on first use of WRD mode
			if a.needsCalibration {
				log.Printf("Running deferred calibration (ISPP (Write/Read) activated)")
				a.calibrateLevels()
				if err := a.saveCalibration(); err != nil {
					log.Printf("Warning: failed to save calibration: %v", err)
				}
				a.needsCalibration = false
			}
			// Reset write/read demo state with improved physics
			startPhase := 2
			skipPrep := true
			if a.useLKSolver() {
				startPhase = 0
				skipPrep = false
			}
			a.wrdPhase = startPhase
			a.wrdPhaseTimer = 0
			a.wrdTargetLevel = rand.Intn(a.numLevels) + 1
			a.wrdNextTargetLevel = 0
			a.wrdStartLevel = a.discreteLevel + 1
			a.wrdLastBranch = 0
			a.wrdForceReset = false
			a.wrdSkipPrep = skipPrep
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
	a.waveformSelect.SetSelected("ISPP (Write/Read)")

	// Physics engine selector
	engines := []string{PhysicsLandau.String(), PhysicsPreisach.String()}
	a.physicsSelect = widget.NewSelect(engines, func(s string) {
		log.Selection("PhysicsEngine", s)
		switch s {
		case PhysicsLandau.String():
			a.setPhysicsEngine(PhysicsLandau)
		case PhysicsPreisach.String():
			a.setPhysicsEngine(PhysicsPreisach)
		}

		// Update plot markers for the active engine
		a.mu.RLock()
		effEc := a.effectiveEc()
		effPr := a.material.Pr
		a.mu.RUnlock()
		if a.plot != nil {
			a.plot.SetMaterialParams(effEc, effPr)
		}
	})
	a.physicsSelect.SetSelected(a.physicsEngine.String())

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
			a.wrdNextTargetLevel = 0
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

	// Target range (effective Ps back-off for level mapping)
	a.wrdRangeLabel = widget.NewLabel(a.rangeFracLabelText(a.wrdRangeFrac))
	a.wrdRangeSlider = widget.NewSlider(0.8, 1.0)
	a.wrdRangeSlider.Step = 0.005
	a.wrdRangeSlider.Value = clampRangeFrac(a.wrdRangeFrac)
	a.wrdRangeSlider.OnChanged = func(v float64) {
		v = clampRangeFrac(v)
		a.mu.Lock()
		if math.Abs(a.wrdRangeFrac-v) < 1e-6 {
			a.mu.Unlock()
			return
		}
		a.wrdRangeFrac = v
		a.calibrated = false
		a.needsCalibration = true
		a.tempCalibrations = make(map[int]*TempCalibration)
		a.mu.Unlock()
		if a.wrdRangeLabel != nil {
			a.wrdRangeLabel.SetText(a.rangeFracLabelText(v))
		}
		a.updateLevelIndicatorRange()
		a.scheduleRangeCalibration()
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
		a.electricField = 0
		// Reset physics model
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

		// Reset trail history
		a.resetHistoryLocked()
		a.simTime = 0

		// Reset WRD (Write/Read Demo) state
		a.wrdPhase = 0
		a.wrdPhaseTimer = 0
		a.wrdTargetLevel = a.numLevels / 2
		a.wrdNextTargetLevel = 0
		a.wrdStartLevel = a.numLevels / 2
		a.wrdReadLevel = 0
		a.wrdRetryCount = 0
		a.wrdLastControllerState = controller.StateIdle
		a.wrdLastControllerPulse = 0
		a.wrdLastProgressLog = 0
		a.wrdLastBranch = 0
		a.wrdForceReset = false
		a.wrdSkipPrep = !a.useLKSolver()
		a.wrdTotalWrites = 0
		a.wrdSuccessWrites = 0
		a.wrdTotalEnergyfJ = 0
		a.wrdCycleEnergy = 0
		if a.writeController != nil {
			a.writeController.ResetState()
		}

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
	// Physics equations button (educational widget)
	eqBtn := widget.NewButton("Eq", func() {
		log.Button("Equation")
		a.showPhysicsEquationsDialog()
	})
	eqBtn.Importance = widget.LowImportance

	formatHz := func(v float64) string {
		if v <= 0 {
			return "0 Hz"
		}
		switch {
		case v >= 1e9:
			return fmt.Sprintf("%.3g GHz", v/1e9)
		case v >= 1e6:
			return fmt.Sprintf("%.3g MHz", v/1e6)
		case v >= 1e3:
			return fmt.Sprintf("%.3g kHz", v/1e3)
		default:
			return fmt.Sprintf("%.3g Hz", v)
		}
	}

	unitFactors := map[string]float64{
		"Hz":  1,
		"kHz": 1e3,
		"MHz": 1e6,
		"GHz": 1e9,
	}
	currentUnit := "Hz"

	// Frequency slider
	freqSlider := widget.NewSlider(0.01, 1.0)
	freqSlider.Step = 0.01
	freqSlider.Value = 0.5
	freqLabel := widget.NewLabel(fmt.Sprintf("Freq: %s", formatHz(freqSlider.Value)))
	freqEntry := widget.NewEntry()
	freqEntry.SetPlaceHolder("Value")
	freqEntry.SetText(fmt.Sprintf("%.3g", freqSlider.Value))
	updatingFreq := false
	unitSelect := widget.NewSelect([]string{"Hz", "kHz", "MHz", "GHz"}, func(s string) {
		if s == "" {
			return
		}
		currentUnit = s
		a.mu.RLock()
		current := a.frequency
		a.mu.RUnlock()
		if updatingFreq {
			return
		}
		updatingFreq = true
		freqEntry.SetText(fmt.Sprintf("%.3g", current/unitFactors[currentUnit]))
		updatingFreq = false
	})
	unitSelect.SetSelected("Hz")
	displayFreqLabel := widget.NewLabel("Display: 0.50 Hz")
	updateDisplayLabel := func(freq, scale float64) {
		displayFreqLabel.SetText(fmt.Sprintf("Display: %s (scale %.3gx)", formatHz(freq*scale), scale))
	}
	setFrequency := func(v float64, source string) {
		if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}
		if updatingFreq {
			return
		}
		updatingFreq = true
		a.mu.Lock()
		a.frequency = v
		// Reset trail when frequency changes
		a.resetHistoryLocked()
		a.simTime = 0
		a.mu.Unlock()
		freqLabel.SetText(fmt.Sprintf("Freq: %s", formatHz(v)))
		updateDisplayLabel(v, a.timeScale)
		if source != "slider" && v >= freqSlider.Min && v <= freqSlider.Max {
			freqSlider.SetValue(v)
		}
		if source != "entry" {
			freqEntry.SetText(fmt.Sprintf("%.3g", v/unitFactors[currentUnit]))
		}
		updatingFreq = false
	}
	freqSlider.OnChanged = func(v float64) {
		log.SliderChange("Frequency", v)
		setFrequency(v, "slider")
	}
	freqEntry.OnSubmitted = func(text string) {
		value, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			a.mu.RLock()
			current := a.frequency
			a.mu.RUnlock()
			freqEntry.SetText(fmt.Sprintf("%.3g", current/unitFactors[currentUnit]))
			return
		}
		a.mu.RLock()
		current := a.frequency
		a.mu.RUnlock()
		log.ValueChange("FrequencyEntry", current, value)
		setFrequency(value*unitFactors[currentUnit], "entry")
	}

	// Time scale slider (slow-motion playback without changing physics)
	timeScaleSlider := widget.NewSlider(0, 9)
	timeScaleSlider.Step = 0.5
	timeScaleSlider.Value = 0
	timeScaleLabel := widget.NewLabel("Time Scale: 1x")
	timeScaleSlider.OnChanged = func(v float64) {
		scale := math.Pow(10, -v)
		log.SliderChange("TimeScale", scale)
		a.mu.Lock()
		a.timeScale = scale
		a.mu.Unlock()
		timeScaleLabel.SetText(fmt.Sprintf("Time Scale: %.3gx", scale))
		a.mu.RLock()
		current := a.frequency
		a.mu.RUnlock()
		updateDisplayLabel(current, scale)
	}
	a.mu.RLock()
	initialFreq := a.frequency
	initialScale := a.timeScale
	a.mu.RUnlock()
	updateDisplayLabel(initialFreq, initialScale)
	// Stress slider (Phase 4.1: Electrostriction control)
	stressSlider := widget.NewSlider(0, 5.0) // 0 to 5 GPa
	stressSlider.Step = 0.1
	stressSlider.Value = 1.0
	stressLabel := widget.NewLabel("Stress: 1.0 GPa")

	// Stress slider (Phase 4.1: Electrostriction control)
	stressSlider.OnChanged = func(v float64) {
		log.SliderChange("Stress", v)
		a.mu.Lock()
		// Pass stress to physics engine
		if a.preisach != nil {
			a.preisach.SetStress(v) // v is in GPa
		}
		if a.lkSolver != nil {
			a.lkSolver.Stress = v * 1e9
		}
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
			previousTemp := a.currentTemperature()
			a.onTemperatureChanged(v)
			// Get current temperature after change
			currentTemp := a.currentTemperature()

			// Clear history if temperature changed significantly (>25K)
			if math.Abs(currentTemp-previousTemp) > 25 {
				a.resetHistoryLocked()
			}

			// Get plot markers with temperature-corrected Ec and nominal Pr
			effEc := a.effectiveEc()
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
	trailSlider := widget.NewSlider(1000, 100000)
	trailSlider.Step = 1000
	trailSlider.Value = float64(a.maxHistory)
	trailLabel := widget.NewLabel(fmt.Sprintf("Trail: %d", a.maxHistory))
	trailSlider.OnChanged = func(v float64) {
		log.SliderChange("TrailLength", v)
		a.mu.Lock()
		a.resizeHistoryLocked(int(v))
		a.mu.Unlock()
		trailLabel.SetText(fmt.Sprintf("Trail: %d", int(v)))
	}

	physicsRow := container.NewBorder(nil, nil, a.physicsSelect, eqBtn, nil)
	eFieldHeader := container.NewBorder(nil, nil, a.eFieldLabel, a.eFieldModeLabel, nil)
	freqEntryRow := container.NewGridWithColumns(2, freqEntry, unitSelect)
	actionRow := container.NewGridWithColumns(3, a.pauseBtn, resetBtn, eli5Btn)

	section := func(title string, content ...fyne.CanvasObject) fyne.CanvasObject {
		return widget.NewCard(title, "", container.NewPadded(container.NewVBox(content...)))
	}

	// Helper to create a row with an info button for tooltips
	withInfo := func(content fyne.CanvasObject, tc sharedwidgets.TooltipContent) fyne.CanvasObject {
		return sharedwidgets.WithInfoButton(content, tc, a.mainWindow)
	}

	// Tooltips reference
	tips := sharedwidgets.HysteresisTooltips

	return container.NewVBox(
		section("Material & Mode",
			withInfo(a.materialBtn, tips.Material),
			withInfo(a.waveformSelect, tips.Waveform),
			withInfo(physicsRow, tips.PhysicsEngine),
		),
		section("Levels & Range",
			withInfo(container.NewVBox(a.levelsLabel, a.levelsEntry), tips.Levels),
			withInfo(container.NewVBox(a.wrdRangeLabel, a.wrdRangeSlider), tips.TargetRange),
		),
		section("Drive & Timing",
			withInfo(container.NewVBox(eFieldHeader, a.eFieldSlider), tips.EField),
			withInfo(container.NewVBox(freqLabel, freqSlider, freqEntryRow), tips.Frequency),
			withInfo(container.NewVBox(timeScaleLabel, timeScaleSlider, displayFreqLabel), tips.TimeScale),
		),
		section("Environment",
			withInfo(container.NewVBox(tempLabel, tempSlider), tips.Temperature),
			withInfo(container.NewVBox(stressLabel, stressSlider), tips.Stress),
		),
		section("Trail",
			withInfo(container.NewVBox(trailLabel, trailSlider), tips.TrailLength),
		),
		section("Run",
			actionRow,
		),
	)
}

// getCurrentMaterialID returns the ID of the currently selected material.
func (a *App) getCurrentMaterialID() string {
	// Map material names to IDs
	nameToID := map[string]string{
		"HZO (Si-doped)":                        "default_hzo",
		"FeCIM HZO":                             "fecim_hzo",
		"FeCIM HZO (TARGET - NOT DEMONSTRATED)": "fecim_hzo_target",
		"Literature Superlattice":               "literature_superlattice",
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
	return "literature_superlattice"
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
	savedTemp := a.currentTemperature()

	a.matIndex = newIdx
	a.material = hzoMat
	a.wrdRangeFrac = rangeFracForMaterial(a.material)
	// Recreate Preisach model for the new material (used in Preisach mode)
	a.preisach = ferroelectric.NewPreisachModel(a.material)
	a.preisach.SetTemperature(savedTemp)

	// Clear history for new material
	a.resetHistoryLocked()

	// Reset simulation state
	a.electricField = 0
	a.simTime = 0
	if a.lkSolver != nil {
		a.lkSolver.Time = 0
	}

	// Reset time-resolved animation state
	a.timeResAnimating = false
	a.timeResIndex = 0
	a.timeResDataTimes = nil
	a.timeResDataPols = nil
	a.timeResDataSwitch = nil

	// Reset write/read demo state
	a.wrdPhase = 0
	a.wrdPhaseTimer = 0
	a.wrdReadLevel = 0
	a.wrdTotalWrites = 0
	a.wrdSuccessWrites = 0
	a.wrdTotalEnergyfJ = 0
	a.wrdCycleEnergy = 0

	// Get temperature-corrected Ec and nominal Pr
	effEc := a.effectiveEc()
	// Use material's nominal Pr (not GetEffectivePr which recalculates from current state)
	effPr := a.material.Pr

	// Update number of levels based on material
	newLevels := a.material.GetNumLevels()
	a.numLevels = newLevels

	// Recreate calibration manager + write controller for the new material/levels
	a.calibManager = algo.NewCalibrationManager(newLevels)
	a.writeController = controller.NewWriteController(newLevels, a.material.Ec, a.material.Ec*2.5, a.calibManager)
	a.writeController.ResetState()
	a.wrdLastControllerState = controller.StateIdle
	a.wrdLastControllerPulse = 0
	a.wrdLastProgressLog = 0
	a.wrdRetryCount = 0
	a.wrdLastBranch = 0
	a.wrdForceReset = false
	a.wrdSkipPrep = !a.useLKSolver()

	// Initialize ISPP calculator with new material parameters
	// a.isppCalc = sharedphysics.NewISPPCalculator(effEc, newLevels)
	// Update Adaptive ISPP
	if a.adaptiveISPP != nil {
		a.adaptiveISPP = sharedphysics.NewAdaptiveISPP(a.lkSolver, a.material)
	}
	if a.lkSolver != nil {
		a.lkSolver.ConfigureFromMaterial(a.material)
		a.lkSolver.Temperature = savedTemp
	}
	if a.useLKSolver() {
		resetP := a.lkDefaultPolarization()
		if a.lkSolver != nil {
			a.lkSolver.SetState(resetP)
			a.polarization = a.lkSolver.GetState()
		} else {
			a.polarization = resetP
		}
		a.normalizedP = 0
	} else {
		a.polarization = 0
		a.normalizedP = 0
	}
	a.syncDiscreteLevelLocked()
	initialLevel := a.discreteLevel
	// Update level indicator and cell visualizer
	if a.levelIndicator != nil {
		a.levelIndicator.SetNumLevels(newLevels)
	}
	if a.cellViz != nil {
		a.cellViz.SetNumLevels(newLevels)
	}
	// Reset target levels
	a.wrdTargetLevel = initialLevel
	a.wrdNextTargetLevel = 0
	a.wrdStartLevel = initialLevel
	a.manualTargetLevel = initialLevel
	// Update bits stored
	a.wrdBitsStored = math.Log2(float64(newLevels))
	// Resize calibration arrays
	a.calibrationUp = make([]float64, newLevels)
	a.calibrationDown = make([]float64, newLevels)
	a.calibUpLow = make([]float64, newLevels)
	a.calibUpHigh = make([]float64, newLevels)
	a.calibDownLow = make([]float64, newLevels)
	a.calibDownHigh = make([]float64, newLevels)
	a.lastErrorUp = make([]int, newLevels)
	a.lastErrorDown = make([]int, newLevels)
	a.relaxCompUp = make([]float64, newLevels)
	a.relaxCompDown = make([]float64, newLevels)

	// Mark calibration as stale
	a.calibrated = false
	a.mu.Unlock()

	// Update UI directly (we're already on the UI thread from dialog callback)
	// Update plot bounds and markers
	a.plot.SetBounds(effEc*2.5, effPr*1.2)
	a.plot.SetMaterialParams(effEc, effPr)
	a.plot.SetData(nil, nil, 0, 0)
	a.plot.Refresh()
	a.updateLevelIndicatorRange()

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
	if a.wrdRangeLabel != nil {
		a.wrdRangeLabel.SetText(a.rangeFracLabelText(a.wrdRangeFrac))
	}
	if a.wrdRangeSlider != nil {
		a.wrdRangeSlider.SetValue(clampRangeFrac(a.wrdRangeFrac))
	}

	// Reset level indicator to middle
	if a.levelIndicator != nil {
		a.levelIndicator.SetLevel(initialLevel)
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
