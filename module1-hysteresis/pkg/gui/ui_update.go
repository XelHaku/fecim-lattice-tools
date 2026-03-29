package gui

import (
	"fmt"
	"math"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	sharedphysics "fecim-lattice-tools/shared/physics"
	"fyne.io/fyne/v2"
)

type widgetSnapshot struct {
	phase  phaseWidgetSnapshot
	target targetWidgetSnapshot
}

type phaseWidgetSnapshot struct {
	mode  string
	phase int
}

type targetWidgetSnapshot struct {
	level     int
	highlight bool
	mode      widgets.TargetMode
}

type uiSnapshot struct {
	fE          float64
	pV          float64
	dL          int
	eC          float64
	hE          []float64
	hP          []float64
	effEc       float64
	effPr       float64
	materialPs  float64
	normalizedP float64

	numLevels     int
	waveform      WaveformType
	physicsEngine PhysicsEngine
	paused        bool

	wrdPhase              int
	wrdTargetLevel        int
	wrdReadLevel          int
	wrdTotalWrites        int
	wrdSuccessWrites      int
	wrdTotalEnergyfJ      float64
	manualAnimating       bool
	manualPhase           int
	manualTargetLevel     int
	manualStartLevel      int
	controllerState       controller.WriteState
	controllerTargetLevel int
	controllerField       float64
	controllerPulseTotal  int
	controllerBestLevel   int
	isppPulseLimit        int

	widgets widgetSnapshot
	logText string
}

// ensureUIUpdateLoop starts the async UI update loop exactly once.
func (a *App) ensureUIUpdateLoop() {
	a.uiUpdateOnce.Do(func() {
		a.uiUpdates = make(chan uiSnapshot, 1)
		go a.uiUpdateLoop()
	})
}

// closeUIUpdateLoop closes the uiUpdates channel so the uiUpdateLoop goroutine
// exits cleanly. Safe to call even if the loop was never started, and safe to
// call multiple times (only the first call closes the channel).
func (a *App) closeUIUpdateLoop() {
	a.uiCloseOnce.Do(func() {
		if a.uiUpdates != nil {
			close(a.uiUpdates)
		}
	})
}

// queueUIUpdate sends the latest UI snapshot without blocking physics.
// Safe to call even after closeUIUpdateLoop (returns silently).
func (a *App) queueUIUpdate(snapshot uiSnapshot) {
	// Guard: skip if the module is shutting down (channel may be closed).
	if !a.running.Load() {
		return
	}
	a.ensureUIUpdateLoop()
	select {
	case a.uiUpdates <- snapshot:
		return
	default:
		// Drop stale update and enqueue the latest.
		select {
		case <-a.uiUpdates:
		default:
		}
		select {
		case a.uiUpdates <- snapshot:
		default:
		}
	}
}

// uiUpdateLoop serializes UI updates on the main thread and coalesces frames.
func (a *App) uiUpdateLoop() {
	for snapshot := range a.uiUpdates {
		// Coalesce to the most recent snapshot if multiple are queued.
		for {
			select {
			case newer := <-a.uiUpdates:
				snapshot = newer
			default:
				goto Apply
			}
		}
	Apply:
		fyne.Do(func() {
			a.refreshGUI(snapshot)
		})
	}
}

// updateUI prepares data and queues refreshGUI on the main thread.
// Safe to call without holding a.mu (it snapshots under a read lock).
func (a *App) updateUI() {
	const uiMinInterval = 33 * time.Millisecond
	const uiMaxPlotPoints = 2000
	if !a.lastUIUpdate.IsZero() && time.Since(a.lastUIUpdate) < uiMinInterval {
		return
	}
	a.lastUIUpdate = time.Now()

	// Take a snapshot under lock.
	a.mu.RLock()
	fE := a.electricField
	pV := a.polarization
	dL := a.discreteLevel
	numLevels := a.numLevels
	waveform := a.waveform
	physicsEngine := a.physicsEngine
	paused := a.paused.Load()
	wrdPhase := a.wrdPhase
	wrdTargetLevel := a.wrdTargetLevel
	wrdReadLevel := a.wrdReadLevel
	wrdTotalWrites := a.wrdTotalWrites
	wrdSuccessWrites := a.wrdSuccessWrites
	wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
	manualAnimating := a.manualAnimating
	manualPhase := a.manualPhase
	manualTargetLevel := a.manualTargetLevel
	manualStartLevel := a.manualStartLevel
	ctrlState := controller.StateIdle
	ctrlTargetLevel := 0
	ctrlField := 0.0
	ctrlPulseTotal := 0
	ctrlBestLevel := 0
	isppLimit := isppPulseLimit(a.isppMaxPulses)
	if a.writeController != nil {
		ctrlState = a.writeController.State
		ctrlTargetLevel = a.writeController.TargetLevel
		ctrlField = a.writeController.CurrentField
		ctrlPulseTotal = a.writeController.TotalPulses + a.writeController.PulseCount
		ctrlBestLevel = a.writeController.BestVerifyLevel
	}
	lastLogPhase := a.lastLogPhase
	logEntries := append([]string(nil), a.logEntries...)
	logVerbose := a.logVerbose
	logText := formatLogEntries(logEntries, logVerbose)
	materialEc := 0.0
	effEc := 0.0
	effPr := 0.0
	materialPs := 0.0
	normalizedP := a.normalizedP
	if a.material != nil {
		materialEc = a.material.Ec
		effEc = a.effectiveEc()
		effPr = a.material.Pr
		materialPs = a.material.Ps
	}
	histLen := a.historyLengthLocked()
	stride := 1
	if histLen > uiMaxPlotPoints {
		stride = histLen / uiMaxPlotPoints
		if histLen%uiMaxPlotPoints != 0 {
			stride++
		}
	}
	points := (histLen + stride - 1) / stride
	eHist := make([]float64, 0, points)
	pHist := make([]float64, 0, points)
	for i := 0; i < histLen; i += stride {
		eVal, pVal := a.historyAtLocked(i)
		eHist = append(eHist, eVal)
		pHist = append(pHist, pVal)
	}
	a.mu.RUnlock()

	a.queueUIUpdate(uiSnapshot{
		fE:                    fE,
		pV:                    pV,
		dL:                    dL,
		eC:                    materialEc,
		hE:                    eHist,
		hP:                    pHist,
		effEc:                 effEc,
		effPr:                 effPr,
		materialPs:            materialPs,
		normalizedP:           normalizedP,
		numLevels:             numLevels,
		waveform:              waveform,
		physicsEngine:         physicsEngine,
		paused:                paused,
		wrdPhase:              wrdPhase,
		wrdTargetLevel:        wrdTargetLevel,
		wrdReadLevel:          wrdReadLevel,
		wrdTotalWrites:        wrdTotalWrites,
		wrdSuccessWrites:      wrdSuccessWrites,
		wrdTotalEnergyfJ:      wrdTotalEnergyfJ,
		manualAnimating:       manualAnimating,
		manualPhase:           manualPhase,
		manualTargetLevel:     manualTargetLevel,
		manualStartLevel:      manualStartLevel,
		controllerState:       ctrlState,
		controllerTargetLevel: ctrlTargetLevel,
		controllerField:       ctrlField,
		controllerPulseTotal:  ctrlPulseTotal,
		controllerBestLevel:   ctrlBestLevel,
		isppPulseLimit:        isppLimit,
		widgets:               a.buildWidgetSnapshot(fE, dL, materialEc, waveform, wrdPhase, wrdTargetLevel, manualAnimating, manualPhase, manualTargetLevel, ctrlState, ctrlTargetLevel, lastLogPhase),
		logText:               logText,
	})
}

const wrdPhaseBoundaryLogMinInterval = 400 * time.Millisecond

// defaultISPPPulseLimit is the fallback ISPP pulse budget when not configured.
// This limits the number of write-verify cycles per target to prevent infinite loops.
const defaultISPPPulseLimit = 30

func isppPulseLimit(maxPulses int) int {
	if maxPulses <= 0 {
		return defaultISPPPulseLimit
	}
	return maxPulses * 3
}

func shouldForceResetAfterISPP(ctrl *controller.WriteController, explicitReset bool) bool {
	if explicitReset {
		return true
	}
	if ctrl == nil {
		return false
	}
	return ctrl.MaxOvershootDelta > 3
}

func (a *App) shouldEmitWRDPhaseBoundaryLog(wrdTarget int) bool {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastWrdBoundaryLog.IsZero() {
		a.lastWrdBoundaryLog = now
		a.lastWrdBoundaryLogTarget = wrdTarget
		return true
	}
	if wrdTarget != a.lastWrdBoundaryLogTarget {
		a.lastWrdBoundaryLog = now
		a.lastWrdBoundaryLogTarget = wrdTarget
		return true
	}
	if now.Sub(a.lastWrdBoundaryLog) >= wrdPhaseBoundaryLogMinInterval {
		a.lastWrdBoundaryLog = now
		a.lastWrdBoundaryLogTarget = wrdTarget
		return true
	}
	return false
}

func (a *App) buildWidgetSnapshot(
	fE float64,
	dL int,
	eC float64,
	waveform WaveformType,
	wrdPhase int,
	wrdTargetLevel int,
	manualAnimating bool,
	manualPhase int,
	manualTargetLevel int,
	ctrlState controller.WriteState,
	ctrlTargetLevel int,
	lastLogPhase int,
) widgetSnapshot {
	const (
		wrdPhaseProgram = 0
		wrdPhaseVerify  = 1
		wrdPhaseResult  = 2
	)

	ws := widgetSnapshot{
		phase:  phaseWidgetSnapshot{mode: "", phase: -1},
		target: targetWidgetSnapshot{level: 0, highlight: false, mode: widgets.TargetModeNone},
	}

	if waveform == WaveformWriteReadDemo {
		targetLevel := wrdTargetLevel
		controllerTargetActive := ctrlState != controller.StateIdle
		if ctrlTargetLevel > 0 && controllerTargetActive {
			targetLevel = ctrlTargetLevel
		}
		atTarget := (dL + 1) == targetLevel
		eFieldSettled := false
		if eC > 0 {
			eFieldSettled = math.Abs(fE) < 0.01*eC
		} else {
			eFieldSettled = math.Abs(fE) < 1e-9
		}
		wrdSettled := atTarget && eFieldSettled

		displayPhase := wrdPhaseProgram
		switch ctrlState {
		case controller.StateVerify:
			displayPhase = wrdPhaseVerify
		case controller.StateSuccess, controller.StateFailed:
			displayPhase = wrdPhaseResult
		default:
			if wrdSettled {
				displayPhase = wrdPhaseResult
			} else if lastLogPhase == wrdPhaseVerify {
				displayPhase = wrdPhaseVerify
			}
		}
		ws.phase = phaseWidgetSnapshot{mode: "wrd", phase: displayPhase}

		highlight := wrdPhase >= 2 && wrdPhase <= 5 && !wrdSettled
		targetMode := widgets.TargetModeWrite
		if ctrlState == controller.StateVerify || ctrlState == controller.StateSuccess {
			targetMode = widgets.TargetModeVerify
		}
		ws.target = targetWidgetSnapshot{level: targetLevel, highlight: highlight, mode: targetMode}
		return ws
	}

	if waveform == WaveformManual {
		if manualAnimating {
			// Map manual ISPP phases to WRD display phases (PROGRAM/VERIFY/RESULT)
			displayPhase := 0 // PhaseProgram
			switch manualPhase {
			case 0: // PREP → PROGRAM
				displayPhase = 0
			case 1: // WRITE (controller active)
				switch ctrlState {
				case controller.StateVerify:
					displayPhase = 1 // PhaseVerify
				case controller.StateSuccess, controller.StateFailed:
					displayPhase = 2 // PhaseResult
				default:
					displayPhase = 0 // PhaseProgram
				}
			case 2: // DISPLAY → RESULT
				displayPhase = 2
			}
			ws.phase = phaseWidgetSnapshot{mode: "wrd", phase: displayPhase}
		} else {
			ws.phase = phaseWidgetSnapshot{mode: "", phase: -1}
		}

		atTarget := (dL + 1) == manualTargetLevel
		eFieldSettled := math.Abs(fE) < 0.01*eC
		settled := atTarget && eFieldSettled
		targetMode := widgets.TargetModeWrite
		if ctrlState == controller.StateVerify || ctrlState == controller.StateSuccess {
			targetMode = widgets.TargetModeVerify
		}
		ws.target = targetWidgetSnapshot{level: manualTargetLevel, highlight: manualAnimating && !settled, mode: targetMode}
	}

	return ws
}

// refreshGUI updates all UI elements with the latest simulation data.
// Must be called on the main/UI thread.
func (a *App) refreshGUI(snapshot uiSnapshot) {
	fE := snapshot.fE
	pV := snapshot.pV
	dL := snapshot.dL
	hE := snapshot.hE
	hP := snapshot.hP
	effEc := snapshot.effEc
	effPr := snapshot.effPr
	materialPs := snapshot.materialPs
	normalizedP := snapshot.normalizedP
	// Update labels
	a.eFieldLabel.SetText(fmt.Sprintf("E-field: %.3f MV/cm", fE/1e8))
	a.pLabel.SetText(fmt.Sprintf("%.2f µC/cm²", pV*100))
	numLevels := snapshot.numLevels
	if numLevels <= 0 {
		a.mu.RLock()
		numLevels = a.numLevels
		a.mu.RUnlock()
	}
	a.levelLabel.SetText(fmt.Sprintf("%d/%d", dL+1, numLevels))

	// Update state descriptor (divide into thirds)
	var stateText string
	lowThird := numLevels / 3
	highThird := numLevels * 2 / 3
	if dL < lowThird {
		stateText = "Negative P"
	} else if dL >= highThird {
		stateText = "Positive P"
	} else {
		stateText = "Intermediate"
	}
	if a.stateLabel != nil {
		a.stateLabel.SetText(stateText)
	}

	// Update stability indicator (M12)
	if a.stabilityIndicator != nil {
		a.stabilityIndicator.SetLevel(dL+1, numLevels)
	}

	// Update wake-up/fatigue labels (Dr. Tour recommendation)
	cycles := a.wrdTotalWrites
	wakeup := 1.0
	degradation := 0.0
	if a.material != nil && a.material.Pr > 0 {
		cfg := sharedphysics.WakeUpModelConfig{
			PrInitial_Cm2:      a.material.Pr,
			WakeUpGainFraction: 0.2,
			WakeUpTauCycles:    1000,
			FatigueOnsetCycles: 1e6,
			FatigueTauCycles:   2e6,
		}
		if pr, err := sharedphysics.WakeUpPolarization(float64(cycles), cfg); err == nil && a.material.Pr > 0 {
			wakeup = pr / (a.material.Pr * (1 + cfg.WakeUpGainFraction))
			if wakeup > 1 {
				wakeup = 1
			}
			degradation = 1 - pr/a.material.Pr
			if degradation < 0 {
				degradation = 0
			}
		}
	}
	if a.cyclesLabel != nil {
		if cycles >= 1000000 {
			a.cyclesLabel.SetText(fmt.Sprintf("%.1fM", float64(cycles)/1e6))
		} else if cycles >= 1000 {
			a.cyclesLabel.SetText(fmt.Sprintf("%.1fK", float64(cycles)/1e3))
		} else {
			a.cyclesLabel.SetText(fmt.Sprintf("%d", cycles))
		}
	}
	if a.wakeupLabel != nil {
		a.wakeupLabel.SetText(fmt.Sprintf("%.1f%%", wakeup*100))
	}
	if a.fatigueLabel != nil {
		a.fatigueLabel.SetText(fmt.Sprintf("%.4f%%", degradation*100))
	}
	// L01: Update cycle phase indicator based on wakeup and degradation
	if a.cyclePhaseLabel != nil {
		var phase string
		if wakeup < 0.95 {
			phase = "WAKE-UP"
		} else if degradation < 0.0001 { // < 0.01% degradation
			phase = "STABLE"
		} else {
			phase = "FATIGUE"
		}
		a.cyclePhaseLabel.SetText(phase)
	}

	// Update temperature-dependent metrics
	switchedFraction := (normalizedP + 1) / 2
	squareness := 0.0
	if materialPs > 0 {
		squareness = effPr / materialPs
	}

	if a.effEcLabel != nil {
		// Show Ec with ±15% uncertainty (typical device-to-device variation)
		ecVal := effEc / 1e8
		a.effEcLabel.SetText(fmt.Sprintf("Ec(T): %.2f±%.2f MV/cm", ecVal, ecVal*0.15))
	}
	if a.effPrLabel != nil {
		// Show Pr with ±20% uncertainty (typical device-to-device variation)
		prVal := effPr * 100
		a.effPrLabel.SetText(fmt.Sprintf("Pr(T): %.1f±%.1f µC/cm²", prVal, prVal*0.20))
	}
	if a.squarenessLabel != nil {
		a.squarenessLabel.SetText(fmt.Sprintf("Sq: %.2f", squareness))
	}
	if a.switchedLabel != nil {
		a.switchedLabel.SetText(fmt.Sprintf("Sw: %.0f%%", switchedFraction*100))
	}

	// Update phase/target widgets from a single UI snapshot struct (G11b).
	a.mu.RLock()
	lastPhase := a.lastLogPhase
	a.mu.RUnlock()
	widgetState := snapshot.widgets
	if widgetState.phase.mode == "" {
		widgetState = a.buildWidgetSnapshot(
			snapshot.fE,
			snapshot.dL,
			snapshot.eC,
			snapshot.waveform,
			snapshot.wrdPhase,
			snapshot.wrdTargetLevel,
			snapshot.manualAnimating,
			snapshot.manualPhase,
			snapshot.manualTargetLevel,
			snapshot.controllerState,
			snapshot.controllerTargetLevel,
			lastPhase,
		)
	}
	wrdDisplayPhase := widgetState.phase.phase

	if a.phaseIndicator != nil {
		a.phaseIndicator.SetPhase(widgetState.phase.phase, widgetState.phase.mode)
	}

	// Update slider to match current E-field (only if not being manually controlled)
	// During Manual animation, the slider reflects the animated E-field
	// Normalize by Ec for display (-1.5 to +1.5 range)
	// Slider is in actual MV/cm — update it whenever not in manual-idle mode.
	shouldUpdateSlider := snapshot.waveform != WaveformManual || snapshot.manualAnimating
	if shouldUpdateSlider {
		a.eFieldSlider.SetValue(fE / 1e8) // V/m → MV/cm
	}

	// Update status and logging
	if snapshot.paused {
		a.statusLabel.SetText("⏸ Paused")
	} else {
		currentWaveform := snapshot.waveform
		wrdTarget := snapshot.wrdTargetLevel
		wrdRead := snapshot.wrdReadLevel
		wrdTotalWrites := snapshot.wrdTotalWrites
		wrdSuccessWrites := snapshot.wrdSuccessWrites
		wrdTotalEnergyfJ := snapshot.wrdTotalEnergyfJ
		midLevel := numLevels / 2 // Dynamic middle level for direction logic

		switch currentWaveform {
		case WaveformWriteReadDemo:
			var phaseStr string
			if snapshot.wrdPhase == 2 && snapshot.controllerState != controller.StateIdle {
				pulseLimit := snapshot.isppPulseLimit
				if pulseLimit <= 0 {
					pulseLimit = defaultISPPPulseLimit
				}
				state := snapshot.controllerState.String()
				best := snapshot.controllerBestLevel
				if best <= 0 {
					best = dL + 1
				}
				a.statusLabel.SetText(fmt.Sprintf("⚡ ISPP: target L%d, pulse %d/%d, state=%s, best=L%d", wrdTarget, snapshot.controllerPulseTotal, pulseLimit, state, best))
				break
			}
			// Log phase transitions (PROGRAM, VERIFY, RESULT) with boundary throttling.
			if wrdDisplayPhase != lastPhase {
				a.mu.Lock()
				a.lastLogPhase = wrdDisplayPhase
				a.mu.Unlock()
				if a.shouldEmitWRDPhaseBoundaryLog(wrdTarget) {
					a.mu.Lock()
					switch wrdDisplayPhase {
					case 0:
						direction := "+"
						if wrdTarget <= midLevel {
							direction = "-"
						}
						a.addLogEntry(fmt.Sprintf("▓▓ PROGRAM L%d | %sE>Ec", wrdTarget, direction))
					case 1:
						a.addLogEntry(fmt.Sprintf("▒▒ VERIFY L%d | E=0", wrdTarget))
					case 2:
						status := "✓ MATCH"
						if wrdRead != wrdTarget {
							diff := abs(wrdRead - wrdTarget)
							if diff == 1 {
								status = fmt.Sprintf("△ ±1 (got %d)", wrdRead)
							} else {
								status = fmt.Sprintf("✗ miss (got %d)", wrdRead)
							}
						}
						successRate := 0.0
						if wrdTotalWrites > 0 {
							successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
						}
						a.addLogEntry(fmt.Sprintf("●● RESULT L%d %s [%.0f%% rate]", wrdTarget, status, successRate))
					}
					a.mu.Unlock()
				}
			}

			// Enhanced status with energy metrics (using local copies from RLock above)
			energyTotal := wrdTotalEnergyfJ
			writeCount := wrdTotalWrites

			switch wrdDisplayPhase {
			case 0:
				direction := "+"
				if wrdTarget <= midLevel {
					direction = "-"
				}
				phaseStr = fmt.Sprintf("PROGRAM L%d | %sE>Ec", wrdTarget, direction)
			case 1:
				phaseStr = fmt.Sprintf("VERIFY L%d | E=0", wrdTarget)
			case 2:
				successRate := 0.0
				if writeCount > 0 {
					successRate = float64(wrdSuccessWrites) / float64(writeCount) * 100
				}
				if wrdRead == wrdTarget {
					phaseStr = fmt.Sprintf("RESULT L%d ✓ | Ops:%d | %.0f%% | %.0fpJ", wrdRead, writeCount, successRate, energyTotal/1000)
				} else {
					phaseStr = fmt.Sprintf("RESULT L%d (want %d) | Ops:%d | %.0f%%", wrdRead, wrdTarget, writeCount, successRate)
				}
			}
			a.statusLabel.SetText(fmt.Sprintf("⚡ FeCIM Write/Read | %s", phaseStr))
		case WaveformManual:
			// Manual ISPP mode status (mirrors WRD status display)
			animAnimating := snapshot.manualAnimating
			manPhase := snapshot.manualPhase
			manTarget := snapshot.manualTargetLevel

			if animAnimating {
				var phaseStr string
				switch manPhase {
				case 0:
					phaseStr = "PREP (saturating...)"
				case 1:
					if snapshot.controllerState != controller.StateIdle {
						pulseLimit := snapshot.isppPulseLimit
						if pulseLimit <= 0 {
							pulseLimit = defaultISPPPulseLimit
						}
						state := snapshot.controllerState.String()
						best := snapshot.controllerBestLevel
						if best <= 0 {
							best = dL + 1
						}
						phaseStr = fmt.Sprintf("ISPP: pulse %d/%d, state=%s, best=L%d",
							snapshot.controllerPulseTotal, pulseLimit, state, best)
					} else {
						phaseStr = "WRITE (starting...)"
					}
				case 2:
					phaseStr = fmt.Sprintf("RESULT L%d (target L%d)", dL+1, manTarget)
				default:
					phaseStr = fmt.Sprintf("Current: L%d", dL+1)
				}
				a.statusLabel.SetText(fmt.Sprintf("⚡ MANUAL L%d | %s", manTarget, phaseStr))
			} else {
				a.statusLabel.SetText(fmt.Sprintf("Manual L%d | Click level bar to write, or drag E-field slider", dL+1))
			}
		case WaveformTimeResolved:
			a.mu.RLock()
			animating := a.timeResAnimating
			idx := a.timeResIndex
			dataLen := len(a.timeResDataTimes)
			a.mu.RUnlock()

			if animating && dataLen > 0 && idx < dataLen {
				a.mu.RLock()
				currentTime := a.timeResDataTimes[idx]
				switchedCount := a.timeResDataSwitch[idx]
				totalHysterons := len(a.timeResDataSwitch)
				a.mu.RUnlock()

				switchedFrac := float64(switchedCount) / float64(totalHysterons) * 100
				a.statusLabel.SetText(fmt.Sprintf("⚡ Time-Resolved | t=%.1f ns | %.0f%% switched | L%d",
					currentTime*1e9, switchedFrac, dL+1))
			} else {
				a.statusLabel.SetText("⚡ Time-Resolved Switching (KAI Dynamics)")
			}
		default:
			frac := (a.normalizedP + 1) / 2 * 100
			a.statusLabel.SetText(fmt.Sprintf("● Running | t=%.2fs | Switched: %.1f%%", a.simTime, frac))
		}
	}

	// Slide panel removed - was distracting and flickering

	// Update log text from the same snapshot payload.
	if snapshot.logText != "" {
		a.logText.SetText(snapshot.logText)
	} else {
		a.logText.SetText(a.getLogText())
	}

	// Update plot
	engine := snapshot.physicsEngine
	a.plot.SetSpikeFiltering(engine != PhysicsLandau)
	a.plot.SetData(hE, hP, fE, pV)
	a.plot.Refresh()

	// Update literature overlay text panel (if dataset loaded)
	a.updateLiteratureOverlayFromData(hE, hP)

	// Update level indicator (level is 0-indexed, display is 1-indexed)
	a.levelIndicator.SetLevel(dL)

	// Highlight target level from the same precomputed snapshot used by phase/log.
	target := widgetState.target
	a.levelIndicator.SetTargetLevelMode(target.level, target.highlight, target.mode)

	a.levelIndicator.Refresh()

	// Update cell visualizer
	a.cellViz.SetLevel(dL)
	a.cellViz.Refresh()
}
