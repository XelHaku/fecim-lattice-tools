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

// updateEFieldSliderRange updates the E-field slider min/max and range label to
// match the current material's Ec (±2.5×Ec in MV/cm). Must be called on UI thread.
func (a *App) updateEFieldSliderRange() {
	if a.eFieldSlider == nil || a.material == nil {
		return
	}
	ecMVcm := a.effectiveEc() / 1e8
	emax := ecMVcm * 2.5
	a.eFieldSlider.Min = -emax
	a.eFieldSlider.Max = emax
	a.eFieldSlider.Step = emax / 250
	a.eFieldSlider.Refresh()
	if a.eFieldRangeLabel != nil {
		a.eFieldRangeLabel.SetText(fmt.Sprintf("range ±%.2f MV/cm", emax))
	}
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
	// E-field slider — range is ±2.5×Ec in actual MV/cm, updated per material.
	initialEcMVcm := a.material.Ec / 1e8
	initialEmax := initialEcMVcm * 2.5
	a.eFieldSlider = widget.NewSlider(-initialEmax, initialEmax)
	a.eFieldSlider.Step = initialEmax / 250 // ~500 steps across full range
	a.eFieldSlider.Value = 0
	a.eFieldSlider.OnChanged = func(v float64) {
		// Only accept slider input when in Manual mode AND not animating
		if a.waveform == WaveformManual && !a.manualAnimating {
			a.mu.Lock()
			a.electricField = v * 1e8 // MV/cm → V/m
			a.mu.Unlock()
		}
	}
	a.eFieldLabel = widget.NewLabel("E: 0.00 MV/cm")
	// Range label shows the current ±max in MV/cm so the user knows the slider extent
	a.eFieldRangeLabel = widget.NewLabel(fmt.Sprintf("range ±%.2f MV/cm", initialEmax))
	a.eFieldRangeLabel.TextStyle = fyne.TextStyle{Italic: true}
	// Mode label shows whether slider is in manual or auto mode
	a.eFieldModeLabel = widget.NewLabel("AUTO")
	a.eFieldModeLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Waveform selector + inline explainability label
	waveforms := []string{"Manual", "Sine Wave", "Triangle Wave", "ISPP (Write/Read)", "Time-Resolved Switching"}
	waveformDescriptions := map[string]string{
		"Manual":                  "Direct E-field control via slider",
		"Sine Wave":               "Continuous sinusoidal E-field sweep",
		"Triangle Wave":           "Linear ramp E-field sweep (standard P-E loop)",
		"ISPP (Write/Read)":       "Write-Read-Demo: programs target levels via pulse sequence",
		"Time-Resolved Switching": "LK switching dynamics visualization",
	}
	a.waveformHelp = widget.NewLabel(waveformDescriptions["ISPP (Write/Read)"])
	a.waveformHelp.TextStyle = fyne.TextStyle{Italic: true}
	a.waveformSelect = widget.NewSelect(waveforms, func(s string) {
		log.Selection("Waveform", s)
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.waveformHelp != nil {
			a.waveformHelp.SetText(waveformDescriptions[s])
		}
		switch s {
		case "Manual":
			a.waveform = WaveformManual
			a.autoMode = false
			a.resetHistoryLocked()
			a.simTime = 0
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
			a.resetHistoryLocked()
			a.simTime = 0
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
			a.resetHistoryLocked()
			a.simTime = 0
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
			a.resetHistoryLocked()
			a.simTime = 0
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
			a.resetHistoryLocked()
			a.simTime = 0
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
		a.syncLKSolverDomainModeLocked()
	})
	// Default to WRD mode explicitly so PROGRAM/VERIFY/RESULT state indicator is active on startup.
	// (SetSelected callback behavior can vary by toolkit/runtime init ordering.)
	a.waveform = WaveformWriteReadDemo
	a.autoMode = true
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

		// Temperature/stress controls only apply to L-K (Preisach has no intrinsic T/σ physics)
		isLK := s == PhysicsLandau.String()
		if a.tempSlider != nil {
			if isLK {
				a.tempSlider.Enable()
			} else {
				a.tempSlider.SetValue(300) // reset to room temperature
				a.tempSlider.Disable()
			}
		}
		if a.stressSlider != nil {
			if isLK {
				a.stressSlider.Enable()
			} else {
				a.stressSlider.SetValue(1.0) // reset to default
				a.stressSlider.Disable()
			}
		}
	})
	a.physicsSelect.SetSelected(a.physicsEngine.String())

	// ISPP method selector
	isppMethods := []string{"Linear (Standard)", "Logarithmic (A-ISPP)", "DCC (Future)"}
	a.isppMethodSelect = widget.NewSelect(isppMethods, func(s string) {
		log.Selection("ISPPMethod", s)
		a.mu.Lock()
		defer a.mu.Unlock()

		if a.writeController == nil {
			return
		}

		switch s {
		case "Linear (Standard)":
			a.writeController.StepMode = "linear"
		case "Logarithmic (A-ISPP)":
			a.writeController.StepMode = "logarithmic"
		case "DCC (Future)":
			log.Printf("DCC method not yet available - keeping current ISPP mode")
			if a.statusLabel != nil {
				a.statusLabel.SetText("DCC method is not yet available in this build; using current ISPP mode")
			}
			// Capture current mode under lock, then revert via fyne.Do to avoid
			// re-entrant SetSelected within OnChanged (fyne.Do queues for next cycle).
			revertTo := "Linear (Standard)"
			if a.writeController.StepMode == "logarithmic" {
				revertTo = "Logarithmic (A-ISPP)"
			}
			fyne.Do(func() {
				a.isppMethodSelect.SetSelected(revertTo)
			})
		}
	})
	a.isppMethodSelect.SetSelected("Logarithmic (A-ISPP)")

	// Material button - shows current material, opens picker on click
	a.materialBtn = widget.NewButton(a.material.Name, func() {
		a.showMaterialPickerDialog()
	})

	// Levels selector (2-64 levels)
	a.levelsLabel = widget.NewLabel(fmt.Sprintf("LE%d (%.1f bits)", a.numLevels, math.Log2(float64(a.numLevels))))
	a.levelsEntry = widget.NewEntry()
	a.levelsEntry.SetText(fmt.Sprintf("%d", a.numLevels))
	a.levelsEntry.SetPlaceHolder("2-64")
	a.levelsEntry.OnChanged = func(s string) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return
		}
		if n < 2 {
			n = 2
		}
		if n > 64 {
			n = 64
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
		a.levelsLabel.SetText(fmt.Sprintf("LE%d (%.1f bits)", n, bits))

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
	a.wrdRangeLabel.Wrapping = fyne.TextWrapWord
	a.wrdRangeLabel.Truncation = fyne.TextTruncateEllipsis
	a.wrdRangeSlider = widget.NewSlider(0.5, 1.0)
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
		wasPaused := a.paused.Load()
		a.paused.Store(!wasPaused)
		if !wasPaused {
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
	}
	// Mechanical stress control (active physics coupling — L-K only)
	stressSlider := widget.NewSlider(0, 5.0) // 0 to 5 GPa
	a.stressSlider = stressSlider
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

	// Temperature slider (L-K only — Preisach has no intrinsic T dependence)
	tempSlider := widget.NewSlider(200, 700)
	a.tempSlider = tempSlider
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

			// SetMaterialParams calls Refresh() — must run on UI thread.
			fyne.Do(func() {
				a.plot.SetMaterialParams(effEc, effPr)
			})
		}()

		// Update label immediately (don't wait for calibration)
		celsius := v - 273
		tcRatio := v / a.material.CurieTemp * 100
		tempLabel.SetText(fmt.Sprintf("T: %.0f K (%.0f°C) [%.0f%% Tc]", v, celsius, tcRatio))
	}

	// Disable T/stress sliders when default engine is Preisach
	if a.physicsEngine == PhysicsPreisach {
		tempSlider.Disable()
		stressSlider.Disable()
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

	lkPolydomainCheck := widget.NewCheck("Polydomain LK for ISPP", func(on bool) {
		a.mu.Lock()
		a.lkISPPPolydomainEnabled = on
		a.syncLKSolverDomainModeLocked()
		a.mu.Unlock()
	})
	lkPolydomainCheck.SetChecked(a.lkISPPPolydomainEnabled)

	eFieldHeader := container.NewBorder(nil, nil, a.eFieldLabel, a.eFieldModeLabel, nil)
	freqEntryRow := container.NewGridWithColumns(2, freqEntry, unitSelect)
	actionRow := container.NewGridWithColumns(2, a.pauseBtn, resetBtn)
	learnRow := container.NewGridWithColumns(2, eli5Btn, eqBtn)

	levelsGrid := container.NewGridWithColumns(2, a.levelsLabel, a.levelsEntry)
	rangeGrid := container.NewVBox(a.wrdRangeLabel, a.wrdRangeSlider)

	// Helper to create a row with an info button for tooltips
	withInfo := func(content fyne.CanvasObject, tc sharedwidgets.TooltipContent) fyne.CanvasObject {
		return sharedwidgets.WithInfoButton(content, tc, a.mainWindow)
	}

	// Tooltips reference
	tips := sharedwidgets.HysteresisTooltips

	mainSections := container.NewVBox(
		widget.NewCard("Material & Mode", "", container.NewVBox(
			withInfo(a.materialBtn, tips.Material),
			withInfo(a.waveformSelect, tips.Waveform),
			withInfo(a.physicsSelect, tips.PhysicsEngine),
			withInfo(a.isppMethodSelect, tips.ISPPMethod),
		)),
		widget.NewCard("Levels & Range", "", container.NewVBox(
			withInfo(levelsGrid, tips.Levels),
			withInfo(rangeGrid, tips.TargetRange),
		)),
		widget.NewCard("Drive & Timing", "", container.NewVBox(
			withInfo(container.NewVBox(eFieldHeader, a.eFieldSlider, a.eFieldRangeLabel), tips.EField),
			withInfo(container.NewVBox(freqLabel, freqSlider, freqEntryRow), tips.Frequency),
			withInfo(container.NewVBox(timeScaleLabel, timeScaleSlider), tips.TimeScale),
		)),
		widget.NewCard("Run", "", container.NewVBox(
			actionRow,
			learnRow,
		)),
		a.createFORCPanel(),
	)

	advanced := widget.NewAccordion(
		widget.NewAccordionItem("Environment", container.NewVBox(
			withInfo(container.NewVBox(tempLabel, tempSlider), tips.Temperature),
			withInfo(container.NewVBox(stressLabel, stressSlider), tips.Stress),
		)),
		widget.NewAccordionItem("History / Trail", withInfo(container.NewVBox(trailLabel, trailSlider), tips.TrailLength)),
		widget.NewAccordionItem("Advanced Physics", lkPolydomainCheck),
	)

	return container.NewVBox(mainSections, advanced)
}

func (a *App) resetAllState() {
	a.mu.Lock()
	a.resetAllStateLocked()
	a.mu.Unlock()

	// Update UI elements (outside lock to avoid deadlock with fyne.Do)
	fyne.Do(func() {
		if a.eFieldSlider != nil {
			a.eFieldSlider.SetValue(0)
		}
		if a.logText != nil {
			a.logText.SetText("")
		}
		if a.statusLabel != nil {
			a.statusLabel.SetText("RESET — all state cleared")
		}
		if a.cyclesLabel != nil {
			a.cyclesLabel.SetText("0")
		}
		if a.wakeupLabel != nil {
			a.wakeupLabel.SetText("100%")
		}
		if a.fatigueLabel != nil {
			a.fatigueLabel.SetText("0%")
		}
		if a.cyclePhaseLabel != nil {
			a.cyclePhaseLabel.SetText("WAKE-UP")
		}
	})
}

func (a *App) resetAllStateLocked() {
	a.electricField = 0
	a.polarization = 0
	a.normalizedP = 0
	a.simTime = 0

	// Reset physics model state.
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
	}
	a.syncDiscreteLevelLocked()

	// Reset trail/history.
	a.resetHistoryLocked()

	// Reset WRD (Write/Read Demo) state.
	a.wrdPhase = 0
	a.wrdPhaseTimer = 0
	a.wrdTargetLevel = a.numLevels / 2
	a.wrdNextTargetLevel = 0
	a.wrdStartLevel = a.numLevels / 2
	a.wrdReadLevel = 0
	a.wrdRetryCount = 0
	a.wrdWriteE = 0
	a.wrdPrepE = 0
	a.wrdSettleE = 0
	a.wrdPrevP = 0
	a.wrdResetStartP = 0
	a.wrdResetEndP = 0
	a.wrdResetEndLvl = 0
	a.wrdWriteStartP = 0
	a.wrdWriteEndP = 0
	a.wrdWriteEndLvl = 0
	a.wrdReadStartP = 0
	a.wrdLastControllerState = controller.StateIdle
	a.wrdLastControllerPulse = 0
	a.wrdLastProgressLog = 0
	a.wrdLastLogState = controller.StateIdle
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

	// Reset ISPP state.
	a.isppPulseCount = 0
	a.isppCurrentVoltage = 0
	a.isppStartVoltage = 0
	a.isppVoltageStep = 0
	a.isppPhase = 0
	a.isppPhaseTimer = 0
	a.isppLastVerifyLvl = 0
	a.isppTotalPulses = 0
	a.isppLastError = 0
	a.isppOscillationCount = 0
	a.isppUseBisection = false
	a.isppReversedOnce = false
	a.isppLowVoltage = 0
	a.isppHighVoltage = 0

	// Reset manual mode animation state.
	a.manualAnimating = false
	a.manualTargetLevel = a.numLevels / 2
	a.manualStartLevel = 0
	a.manualPhase = 0
	a.manualPhaseTime = 0

	// Reset Time-Resolved animation state.
	a.timeResAnimating = false
	a.timeResIndex = 0
	a.timeResDataTimes = nil
	a.timeResDataPols = nil
	a.timeResDataSwitch = nil

	// Reset logs.
	a.logEntries = a.logEntries[:0]
	a.lastLogPhase = -1
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

	// Update number of levels based on material (GUI-validated range)
	newLevels := a.material.GetNumLevels()
	if newLevels < 2 {
		newLevels = 2
	}
	if newLevels > 64 {
		newLevels = 64
	}
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
	a.syncLKSolverDomainModeLocked()
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
	// Update plot bounds, markers, and clear data. updateLevelIndicatorRange sets
	// the correct pMax (accounting for Ps*rangeFrac headroom) and syncs the plot.
	a.updateLevelIndicatorRange()
	a.plot.SetMaterialParams(effEc, effPr)
	a.plot.SetData(nil, nil, 0, 0)
	a.plot.Refresh()
	a.updateSimVsExpWidget()

	// Reset E-field slider and update range for new material
	a.updateEFieldSliderRange()
	if a.eFieldSlider != nil {
		a.eFieldSlider.SetValue(0)
	}

	// Update material button text
	if a.materialBtn != nil {
		a.materialBtn.SetText(hzoMat.Name)
	}

	// Update info panel material fields
	if a.materialNameLabel != nil {
		a.materialNameLabel.SetText(hzoMat.Name)
	}
	if a.materialPropsLabel != nil {
		a.materialPropsLabel.SetText(fmt.Sprintf("Pr %.1f  Ec %.2f MV/cm",
			hzoMat.Pr*100, hzoMat.Ec/1e8))
	}
	if a.materialBadgeBox != nil {
		confidence := sharedwidgets.Estimated
		if strings.Contains(strings.ToLower(hzoMat.Name), "materlik") || strings.Contains(strings.ToLower(hzoMat.Name), "fecim hzo") {
			confidence = sharedwidgets.Calibrated
		}
		a.materialBadgeBox.Objects = []fyne.CanvasObject{sharedwidgets.NewConfidenceBadge(confidence)}
		a.materialBadgeBox.Refresh()
	}

	// Update levels entry and label
	if a.levelsEntry != nil {
		a.levelsEntry.SetText(fmt.Sprintf("%d", newLevels))
	}
	if a.levelsLabel != nil {
		a.levelsLabel.SetText(fmt.Sprintf("LE%d (%.1f bits)", newLevels, math.Log2(float64(newLevels))))
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
