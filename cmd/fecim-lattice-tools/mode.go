package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	hysheadless "fecim-lattice-tools/module1-hysteresis/pkg/headless"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

func runMode(mode string, engine string) error {
	return runHeadlessEngine(mode, HeadlessEngineOptions{Engine: engine})
}

func normalizeEngine(engine string) string {
	engine = strings.ToLower(strings.TrimSpace(engine))
	switch engine {
	case "", "preisach", "p":
		return "preisach"
	case "lk", "l-k", "landau":
		return "lk"
	default:
		return engine
	}
}

func runHysteresisMode(engine string) error {
	log := logging.NewLogger("hysteresis-mode")
	if stop, enabled := maybeStartHeadlessPprof(log); enabled {
		defer stop()
	}
	engine = normalizeEngine(engine)
	if engine == "" {
		engine = "preisach"
	}
	if engine != "preisach" && engine != "lk" {
		return fmt.Errorf("unknown hysteresis engine %q (expected: preisach|lk)", engine)
	}
	log.Info("Headless hysteresis engine: %s", engine)

	ensureFinite := func(label string, values ...float64) error {
		for _, v := range values {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("non-finite %s=%v", label, v)
			}
		}
		return nil
	}

	mat := func() *physics.HZOMaterial {
		// Optional material selector for headless testing.
		// Examples:
		//   FECIM_MATERIAL=fecim_hzo
		//   FECIM_MATERIAL=literature_superlattice
		//   FECIM_MATERIAL=default_hzo
		name := strings.ToLower(strings.TrimSpace(os.Getenv("FECIM_MATERIAL")))
		switch name {
		case "", "fecim", "fecim_hzo", "fecim-hzo":
			return physics.FeCIMMaterial()
		case "fecim_target", "fecim_hzo_target", "fecim-hzo-target":
			return physics.FeCIMMaterialTarget()
		case "default", "default_hzo", "default-hzo":
			return physics.DefaultHZO()
		case "literature", "literature_superlattice", "superlattice", "cheema", "cheema2020":
			return physics.LiteratureSuperlattice()
		case "cryo", "cryogenic", "cryogenic_hzo", "cryogenic-hzo":
			return physics.CryogenicHZO()
		case "hzo32", "hzo_standard_32", "hzo-standard-32":
			return physics.HZOStandard32()
		case "ftj140", "hzo_ftj_140", "hzo-ftj-140":
			return physics.HZOFJT140()
		case "custom14", "hzo_custom_14", "hzo-custom-14":
			return physics.HZOCustom14()
		case "alscn", "al-scn":
			return physics.AlScN()
		default:
			log.Info("Unknown FECIM_MATERIAL=%q; defaulting to FeCIMMaterial", name)
			return physics.FeCIMMaterial()
		}
	}()

	if mat.Ec > 0 {
		// Headless LK diagnostics assume SI units (V/m). Catch MV/cm or V/cm mixups early.
		// Typical Ec values for HZO presets are O(1e8) V/m (≈ 1 MV/cm).
		if mat.Ec < 1e6 || mat.Ec > 1e10 {
			return fmt.Errorf("material Ec=%g looks non-SI; expected V/m (~1e8). Refuse headless run", mat.Ec)
		}
	}
	numLevels := mat.GetNumLevels()
	if numLevels <= 0 {
		numLevels = physics.DefaultLevels
	}
	midLevel := numLevels / 2

	gmin := mat.Gmin
	gmax := mat.Gmax
	if gmin == 0 && gmax == 0 {
		gmin = 1e-6
		gmax = 100e-6
	}

	rangeFrac := mat.TargetRangeFrac
	if s := strings.TrimSpace(os.Getenv("FECIM_RANGE_FRAC")); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			rangeFrac = v
		}
	}
	if rangeFrac <= 0 || rangeFrac > 1 {
		rangeFrac = 1
	}
	// Many LK presets cannot reliably reach full ±Ps due to depolarization and relaxation.
	// Materials provide TargetRangeFrac as the recommended "reachable" outer range.
	// Clamp upward overrides to avoid pathological headless runs that time out.
	if engine == "lk" && mat.TargetRangeFrac > 0 && mat.TargetRangeFrac <= 1 && rangeFrac > mat.TargetRangeFrac {
		log.Info("Headless rangeFrac=%.3f exceeds material TargetRangeFrac=%.3f; clamping for reachability", rangeFrac, mat.TargetRangeFrac)
		rangeFrac = mat.TargetRangeFrac
	}
	effPs := mat.Ps * rangeFrac

	var dataLogger *hysheadless.HysteresisDataLogger
	if logger, err := hysheadless.NewHysteresisDataLogger(mat.Name); err != nil {
		log.Info("Headless CSV logging disabled: %v", err)
	} else {
		dataLogger = logger
		log.Info("Headless CSV logging enabled: %s", logger.Path())
		defer func() {
			if err := dataLogger.Close(); err != nil {
				log.Info("Headless CSV logging close failed: %v", err)
			}
		}()
	}

	var (
		solver        *physics.LKSolver
		preisachModel *ferroelectric.PreisachModel
		engineTemp    float64
		simTime       float64
		currentP      float64
	)
	switch engine {
	case "lk":
		solver = physics.NewLKSolver()
		solver.ConfigureFromMaterial(mat)
		solver.Temperature = 300
		solver.EnableNoise = false
		solver.UseNLS = false
		if !solver.UseMaterialAlpha {
			solver.UpdateParams()
		}
		engineTemp = solver.Temperature
		currentP = solver.GetState()

		log.Info("LK config: Beta=%.3e Gamma=%.3e Rho=%.3e K_dep=%.3e Q12=%.3e Stress=%.2f GPa SeriesR=%.1f Ohm Thickness=%.2e m Area=%.2e m^2 CurieTemp=%.1fK CurieConst=%.2e UseEffVisc=%v UseNLS=%v Noise=%v UseMaterialAlpha=%v",
			solver.Beta, solver.Gamma, solver.Rho, solver.K_dep, solver.Q12, solver.Stress/1e9, solver.SeriesResistance, solver.Thickness, solver.Area,
			solver.CurieTemp, solver.CurieConst, solver.UseEffectiveViscosity, solver.UseNLS, solver.EnableNoise, solver.UseMaterialAlpha)
		log.Info("LK alpha=%.3e at T=%.1fK", solver.Alpha, solver.Temperature)
	case "preisach":
		preisachModel = ferroelectric.NewPreisachModel(mat)
		preisachModel.SetTemperature(300)
		engineTemp = preisachModel.Temperature
		currentP = preisachModel.Polarization()
		log.Info("Preisach config: Ec=%.3e Ps=%.3e Delta=%.3e Temp=%.1fK",
			mat.Ec, mat.Ps, mat.Ec*0.25, engineTemp)
	}

	stepPolarization := func(E, dt float64) float64 {
		if solver != nil {
			return solver.Step(E, dt)
		}
		return preisachModel.Update(E)
	}

	Emax := mat.Ec * 1.2
	dt := 1e-12

	perfEnabled := logging.IsVerbose(logging.VerbosityDebug)
	perfSteps := 0
	perfDtMin := math.MaxFloat64
	perfDtMax := 0.0
	perfDtSum := 0.0
	perfStepTime := time.Duration(0)
	perfWallTime := time.Duration(0)
	perfDtMinHits := 0
	perfDtMinThreshold := 0.0
	sweepLabel := "LK_SWEEP"
	if engine == "preisach" {
		sweepLabel = "PREISACH_SWEEP"
	}

	resetPerf := func() {
		if !perfEnabled {
			return
		}
		perfSteps = 0
		perfDtMin = math.MaxFloat64
		perfDtMax = 0
		perfDtSum = 0
		perfStepTime = 0
		perfWallTime = 0
		perfDtMinHits = 0
	}

	recordPerf := func(step float64, elapsed time.Duration) {
		if !perfEnabled {
			return
		}
		perfSteps++
		if step < perfDtMin {
			perfDtMin = step
		}
		if step > perfDtMax {
			perfDtMax = step
		}
		perfDtSum += step
		perfStepTime += elapsed
		if perfDtMinThreshold > 0 && step <= perfDtMinThreshold*1.001 {
			perfDtMinHits++
		}
	}

	logPerf := func(label string) {
		if !perfEnabled || perfSteps == 0 {
			return
		}
		dtMean := perfDtSum / float64(perfSteps)
		minShare := 0.0
		if perfSteps > 0 {
			minShare = float64(perfDtMinHits) / float64(perfSteps) * 100.0
		}
		solverMs := perfStepTime.Seconds() * 1000.0
		wallMs := perfWallTime.Seconds() * 1000.0
		solverShare := 0.0
		if wallMs > 0 {
			solverShare = (solverMs / wallMs) * 100.0
		}
		stepNs := 0.0
		if perfSteps > 0 {
			stepNs = perfStepTime.Seconds() * 1e9 / float64(perfSteps)
		}
		log.Info("%s_PERF %s: steps=%d dtMin=%.3e dtMean=%.3e dtMax=%.3e dtMinHits=%d (%.1f%%) solverMs=%.2f wallMs=%.2f solverShare=%.1f%% stepNs=%.1f",
			strings.ToUpper(engine), label, perfSteps, perfDtMin, dtMean, perfDtMax, perfDtMinHits, minShare, solverMs, wallMs, solverShare, stepNs)
	}

	log.Info("Hysteresis diagnostic sweep starting")
	resetPerf()
	perfDtMinThreshold = 0
	sweepStartWall := time.Now()
	for _, E := range []float64{-Emax, -0.5 * Emax, 0, 0.5 * Emax, Emax} {
		if perfEnabled {
			stepStart := time.Now()
			currentP = stepPolarization(E, dt)
			recordPerf(dt, time.Since(stepStart))
		} else {
			currentP = stepPolarization(E, dt)
		}
		simTime += dt
		if err := ensureFinite("sweep", currentP, E); err != nil {
			return err
		}
		recordHeadlessSnapshot(dataLogger, headlessEngineState{time: simTime, temperature: engineTemp, polarization: currentP}, mat, effPs, gmin, gmax, numLevels, nil, dt, E, sweepLabel, "SWEEP", nil)
	}
	perfWallTime = time.Since(sweepStartWall)
	logPerf("SWEEP")

	log.Info("ISPP write-verify sequence starting")
	log.Info("Headless level mapping: effPs=%.3e C/m^2 (rangeFrac=%.3f, matPs=%.3e)", effPs, rangeFrac, mat.Ps)
	calibManager := algo.NewCalibrationManager(numLevels)
	writeController := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, calibManager)

	pulseDuration := mat.Tau
	if pulseDuration <= 0 {
		pulseDuration = 10e-9
	}
	writeController.PulseDuration = pulseDuration
	phaseDuration := pulseDuration / 0.4
	if phaseDuration <= 0 {
		phaseDuration = pulseDuration
	}
	// Keep phase duration in a numerically practical range for headless loops.
	// Some presets define very small Tau, which can make maxSimTime collapse and
	// trigger immediate timeouts before the controller can converge.
	if phaseDuration < 1e-9 {
		phaseDuration = 1e-9
	}

	// Headless integration tests can become extremely slow if we pick time steps that
	// are too small relative to the pulse duration. Provide deterministic, opt-in
	// speed controls via environment variables.
	//
	//   FECIM_ISPP_STEPS_PER_PULSE: target number of simulation steps per pulse (int)
	//   FECIM_HEADLESS_FAST=1: convenience preset for CI (sets steps-per-pulse to 200)
	stepsPerPulse := 2000
	if s := strings.TrimSpace(os.Getenv("FECIM_ISPP_STEPS_PER_PULSE")); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			stepsPerPulse = v
		}
	}
	if strings.TrimSpace(os.Getenv("FECIM_HEADLESS_FAST")) == "1" {
		stepsPerPulse = 200
	}
	if stepsPerPulse < 20 {
		stepsPerPulse = 20
	}

	// Nominal dt is pulseDuration/stepsPerPulse; allow dtMin to get smaller only
	// near coercive points (for stability) but never below 1e-15 s.
	dtNominal := pulseDuration / float64(stepsPerPulse)
	if dtNominal <= 0 {
		dtNominal = 1e-12
	}
	dtMin := dtNominal / 20.0
	if dtMin < 1e-15 {
		dtMin = 1e-15
	}
	dtMax := pulseDuration / 2.0
	if dtMax <= 0 {
		dtMax = dtNominal
	}
	log.Info("LK_DIAG timing: pulseDuration=%.3e s stepsPerPulse=%d dtNominal=%.3e s dtMin=%.3e s dtMax=%.3e s", pulseDuration, stepsPerPulse, dtNominal, dtMin, dtMax)

	gWindow := gmax - gmin
	// Test targets span near-saturation + low + mid-range.
	// Level-to-G: G = gmin + (level-1)/(numLevels-1) * gWindow
	levelToG := func(level int) float64 {
		if level < 1 {
			level = 1
		}
		if level > numLevels {
			level = numLevels
		}
		return gmin + float64(level-1)/float64(numLevels-1)*gWindow
	}

	// Build a deterministic target list that mixes extremes and mid-levels.
	//
	// Controls:
	//   FECIM_ISPP_TARGETS: number of targets to run (default 5)
	//   FECIM_ISPP_TARGET_SEED: deterministic seed for randomized fill (default 1)
	targetCount := 5
	if s := strings.TrimSpace(os.Getenv("FECIM_ISPP_TARGETS")); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			targetCount = v
		}
	}
	if targetCount < 1 {
		targetCount = 1
	}
	if targetCount > numLevels {
		targetCount = numLevels
	}

	seed := int64(1)
	if s := strings.TrimSpace(os.Getenv("FECIM_ISPP_TARGET_SEED")); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			seed = v
		}
	}

	buildTargets := func(numLevels, n int, seed int64) []int {
		if n <= 0 {
			return nil
		}
		if n > numLevels {
			n = numLevels
		}
		mid := int(math.Round(float64(numLevels+1) / 2.0))
		q1 := int(math.Round(0.25 * float64(numLevels)))
		q3 := int(math.Round(0.75 * float64(numLevels)))
		base := []int{1, 2, 3, q1, mid, q3, numLevels - 2, numLevels - 1, numLevels}
		seen := map[int]bool{}
		out := make([]int, 0, n)
		push := func(level int) {
			if level < 1 {
				level = 1
			}
			if level > numLevels {
				level = numLevels
			}
			if len(out) >= n {
				return
			}
			if !seen[level] {
				seen[level] = true
				out = append(out, level)
			}
		}
		for _, lvl := range base {
			push(lvl)
		}
		rng := rand.New(rand.NewSource(seed))
		for len(out) < n {
			lvl := 1 + rng.Intn(numLevels)
			push(lvl)
		}
		// Shuffle deterministically (we want a mixture during the run, not sorted).
		for i := len(out) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			out[i], out[j] = out[j], out[i]
		}
		return out
	}

	parseTargetLevels := func(raw string, numLevels int) []int {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil
		}
		mid := int(math.Round(float64(numLevels+1) / 2.0))
		parts := strings.Split(raw, ",")
		out := make([]int, 0, len(parts))
		for _, p := range parts {
			tok := strings.ToLower(strings.TrimSpace(p))
			if tok == "" {
				continue
			}
			switch tok {
			case "lo", "low", "l":
				out = append(out, 1)
			case "mid", "middle", "m":
				out = append(out, mid)
			case "hi", "high", "h":
				out = append(out, numLevels)
			default:
				if v, err := strconv.Atoi(tok); err == nil {
					if v < 1 {
						v = 1
					}
					if v > numLevels {
						v = numLevels
					}
					out = append(out, v)
				}
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}

	targetLevels := parseTargetLevels(os.Getenv("FECIM_ISPP_TARGET_LEVELS"), numLevels)
	if len(targetLevels) == 0 {
		targetLevels = buildTargets(numLevels, targetCount, seed)
	} else {
		log.Info("Using explicit ISPP target levels from FECIM_ISPP_TARGET_LEVELS=%q => %v", os.Getenv("FECIM_ISPP_TARGET_LEVELS"), targetLevels)
	}
	steps := make([]struct {
		label   string
		targetG float64
	}, 0, len(targetLevels))
	for i, lvl := range targetLevels {
		steps = append(steps, struct {
			label   string
			targetG float64
		}{
			label:   fmt.Sprintf("T%02d_L%02d", i+1, lvl),
			targetG: levelToG(lvl),
		})
	}

	wrdTotalWrites := 0
	wrdSuccessWrites := 0

	for i, step := range steps {
		resetPerf()
		perfDtMinThreshold = dtMin
		targetWallStart := time.Now()
		_, targetLevel := headlessLevelFromConductance(step.targetG, gmin, gmax, numLevels)
		currentField := 0.0
		currentG := physics.PolarizationToConductance(currentP, effPs, gmin, gmax)
		_, currentLevel := headlessLevelFromConductance(currentG, gmin, gmax, numLevels)
		stepStart := simTime
		pulseBudget := 800
		if s := strings.TrimSpace(os.Getenv("FECIM_ISPP_MAX_PULSES")); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 {
				pulseBudget = v
			}
		}
		if strings.TrimSpace(os.Getenv("FECIM_HEADLESS_FAST")) == "1" && pulseBudget < 1200 {
			pulseBudget = 1200
		}
		// Headless-fast LK runs must not spin forever on difficult/unreachable targets.
		// Prefer explicit controller failure/reset to a simulation-time timeout.
		if strings.TrimSpace(os.Getenv("FECIM_HEADLESS_FAST")) == "1" && engine == "lk" {
			writeController.MaxRetries = pulseBudget / 4
			if writeController.MaxRetries < 100 {
				writeController.MaxRetries = 100
			}
			writeController.ForceResetLimit = 1
		}
		maxSimTime := phaseDuration * float64(pulseBudget)
		if maxSimTime < 1e-3 {
			maxSimTime = 1e-3
		}
		wrd := &headlessWRDState{
			phase:         0,
			phaseTimer:    0,
			targetLevel:   targetLevel,
			startLevel:    currentLevel,
			readLevel:     currentLevel,
			retryCount:    0,
			cycleEnergy:   0,
			totalWrites:   wrdTotalWrites,
			successWrites: wrdSuccessWrites,
		}

		recordHeadlessSnapshot(dataLogger, headlessEngineState{time: simTime, temperature: engineTemp, polarization: currentP}, mat, effPs, gmin, gmax, numLevels, writeController, 0, currentField, "ISPP", "", wrd)

		targetDone := false
		maxLevelDelta := 0
		stuckBreakerCount := 0
		lastGuardPulseCount := writeController.GuardPulseCount
		// Track "written P" - the polarization captured at end of pulse (before E→0)
		// This is critical because high K_dep causes P to relax toward 0 when E=0,
		// which would give wrong level readings during VERIFY phase.
		var writtenP float64
		var lastControllerState = controller.StateIdle
		// Track the state at success for accurate reporting (before DISPLAY relaxation)
		var successP float64
		var successLevel int
		for simTime-stepStart < maxSimTime && !targetDone {
			// Capture "written P" at transition from WAIT to VERIFY (before E→0)
			// This gives the true written state before depolarization relaxation
			if writeController.State == controller.StateVerify && lastControllerState == controller.StateWait {
				writtenP = currentP
			}
			lastControllerState = writeController.State

			// Use writtenP for level calculation during VERIFY/SUCCESS/FAILED states
			// to avoid depolarization-induced errors at E=0
			effectiveP := currentP
			if writeController.State == controller.StateVerify ||
				writeController.State == controller.StateSuccess ||
				writeController.State == controller.StateFailed {
				if writtenP != 0 {
					effectiveP = writtenP
				}
			}

			currentG := physics.PolarizationToConductance(effectiveP, effPs, gmin, gmax)
			_, currentLevel = headlessLevelFromConductance(currentG, gmin, gmax, numLevels)
			wrd.readLevel = currentLevel
			levelDelta := currentLevel - targetLevel
			if levelDelta < 0 {
				levelDelta = -levelDelta
			}
			if levelDelta > maxLevelDelta {
				maxLevelDelta = levelDelta
			}

			currentStep := dtNominal
			if mat.Ec > 0 {
				distPlus := math.Abs(currentField - mat.Ec)
				distMinus := math.Abs(currentField + mat.Ec)
				minDist := math.Min(distPlus, distMinus)
				if minDist < 1e7 {
					currentStep = dtMin
				}
			}
			if currentStep > dtMax {
				currentStep = dtMax
			}

			wrd.phaseTimer += currentStep

			switch wrd.phase {
			case 0: // PREP - drive to saturation in opposite direction of target
				// For headless runs we apply fields directly (no slew-rate limiting).
				// The slew limiter is useful for GUI smoothness but it can mask physics issues
				// by never actually reaching the intended field during short phases.
				// For upper targets (>15): saturate NEGATIVE first (level ~1), then write UP
				// For lower targets (<=15): saturate POSITIVE first (level ~30), then write DOWN
				prepE := 0.0
				saturated := false
				// Use a relaxed saturation threshold because depolarization (K_dep) and
				// some LK parameter sets can limit the achievable steady-state |P| well
				// below Ps (especially for literature/superlattice presets).
				// Also use 2.0×Ec drive field for faster saturation.
				if targetLevel > numLevels/2 {
					// Upper target: drive to negative saturation
					prepE = -mat.Ec * 2.0
					saturated = currentP <= -0.50*mat.Ps
				} else {
					// Lower target: drive to positive saturation
					prepE = mat.Ec * 2.0
					saturated = currentP >= 0.50*mat.Ps
				}
				wrd.prepE = prepE

				currentField = prepE

				// Transition only after field reached AND level actually saturated
				if wrd.phaseTimer > phaseDuration*0.25 && math.Abs(currentField-prepE) < 0.01*mat.Ec*2 && saturated {
					if targetLevel > numLevels/2 {
						wrd.saturatedDirection = -1
					} else {
						wrd.saturatedDirection = 1
					}
					// Skip HOLD_RESET (phase 1) to prevent polarization relaxation
					// Go directly to WRITE phase with fromSaturation=true
					fromSaturation := true
					writeController.PulseDuration = phaseDuration * 0.4
					writeController.Start(targetLevel, fromSaturation)
					wrd.phase = 2
					wrd.phaseTimer = 0
				}

			// case 1: HOLD_RESET - DELETED to prevent polarization relaxation
			// Phase 1 is now skipped - transition goes directly from PREP (0) to WRITE (2)

			case 2: // WRITE - delegated to WriteController
				targetField, done := writeController.Update(currentStep, currentField, currentLevel, 0)

				// Headless: apply exactly the controller-requested field.
				// This avoids a pathological interaction where the controller believes
				// E reached its setpoint but the simulation lags behind (and then VERIFY
				// happens at a non-zero residual field like ~0.025×Ec).
				currentField = targetField

				wrd.writeE = targetField

				if done {
					switch writeController.State {
					case controller.StateSuccess:
						// Capture the written state BEFORE DISPLAY phase relaxation
						successP = effectiveP
						successLevel = currentLevel
						wrdTotalWrites++
						wrdSuccessWrites++
						wrd.totalWrites = wrdTotalWrites
						wrd.successWrites = wrdSuccessWrites

						// Learn calibration from successful write (match GUI behavior)
						learnedE := writeController.CurrentField
						targetIdx := targetLevel - 1
						if calibManager != nil && targetIdx >= 0 && targetIdx < len(calibManager.CalibrationUp) {
							if targetLevel > midLevel {
								// Ascending (written from reset negative)
								calibManager.UpdateCalibrationUp(targetIdx, 0, mat.Ec)
								calibManager.CalibrationUp[targetIdx] = learnedE
							} else {
								// Descending (written from reset positive)
								calibManager.UpdateCalibrationDown(targetIdx, 0, mat.Ec)
								calibManager.CalibrationDown[targetIdx] = learnedE
							}
							log.Info("CALIB LEARN: target=%d learnedE=%.3f×Ec (updated via CM)", targetLevel, learnedE/mat.Ec)
						}

						wrd.phase = 5
						wrd.phaseTimer = 0
					case controller.StateForceReset:
						wrd.phase = 0
						wrd.phaseTimer = 0
					case controller.StateFailed:
						wrdTotalWrites++
						wrd.totalWrites = wrdTotalWrites
						wrd.phase = 5
						wrd.phaseTimer = 0
					}
				}

			case 5: // DISPLAY - return to zero, then finish target
				// Headless: snap to zero for verify/display phases.
				currentField = 0

				if wrd.phaseTimer > phaseDuration*0.6 && math.Abs(currentField) < 0.01*mat.Ec*2 {
					targetDone = true
				}
			}

			wrd.retryCount = writeController.RetryCount
			if writeController.GuardPulseCount > lastGuardPulseCount {
				stuckBreakerCount += writeController.GuardPulseCount - lastGuardPulseCount
			}
			lastGuardPulseCount = writeController.GuardPulseCount

			if perfEnabled {
				stepStart := time.Now()
				currentP = stepPolarization(currentField, currentStep)
				recordPerf(currentStep, time.Since(stepStart))
			} else {
				currentP = stepPolarization(currentField, currentStep)
			}
			simTime += currentStep
			if err := ensureFinite("ispp", currentP, currentField, currentStep); err != nil {
				return err
			}
			recordHeadlessSnapshot(dataLogger, headlessEngineState{time: simTime, temperature: engineTemp, polarization: currentP}, mat, effPs, gmin, gmax, numLevels, writeController, currentStep, currentField, "ISPP", "", wrd)
		}

		if !targetDone {
			timeoutErr := fmt.Errorf("ISPP step %d (%s): timed out after %.3fs", i+1, step.label, simTime-stepStart)
			if strings.TrimSpace(os.Getenv("FECIM_HEADLESS_ALLOW_TIMEOUT")) == "1" {
				log.Info("%v (continuing: FECIM_HEADLESS_ALLOW_TIMEOUT=1)", timeoutErr)
			} else {
				return timeoutErr
			}
		}
		perfWallTime = time.Since(targetWallStart)
		logPerf(fmt.Sprintf("ISPP_%s", step.label))

		// Use the captured success state if available, otherwise use relaxed state
		finalP := currentP
		finalG := physics.PolarizationToConductance(finalP, effPs, gmin, gmax)
		_, finalLevel := headlessLevelFromConductance(finalG, gmin, gmax, numLevels)
		success := writeController.State == controller.StateSuccess
		attempts := writeController.PulseCount
		overshoots := writeController.OvershootCount

		// Report the written state (at success) not the relaxed state (after DISPLAY)
		reportP := finalP
		reportLevel := finalLevel
		if success && successP != 0 {
			reportP = successP
			reportLevel = successLevel
		}
		reportG := physics.PolarizationToConductance(reportP, effPs, gmin, gmax)

		log.Info("ISPP step %d (%s): targetG=%.3e targetLevel=%d attempts=%d success=%v overshoots=%d maxLevelDelta=%d stuckBreakers=%d writtenLevel=%d writtenP=%.3e writtenG=%.3e (relaxedP=%.3e)",
			i+1, step.label, step.targetG, targetLevel, attempts, success, overshoots, maxLevelDelta, stuckBreakerCount, reportLevel, reportP, reportG, finalP)
	}

	return nil
}

func maybeStartHeadlessPprof(log *logging.Logger) (func(), bool) {
	if strings.TrimSpace(os.Getenv("FECIM_PPROF")) != "1" {
		return nil, false
	}
	addr := strings.TrimSpace(os.Getenv("FECIM_PPROF_ADDR"))
	if addr == "" {
		addr = "127.0.0.1:6060"
	}

	srv := &http.Server{Addr: addr}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info("Headless pprof server error: %v", err)
		}
	}()
	log.Info("Headless pprof enabled at http://%s/debug/pprof/", addr)

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Info("Headless pprof shutdown error: %v", err)
		}
	}
	return stop, true
}

const (
	headlessMVPerCM  = 1e8
	headlessUCPerCM2 = 1e2
)

type headlessWRDState struct {
	phase              int
	phaseTimer         float64
	targetLevel        int
	readLevel          int
	retryCount         int
	cycleEnergy        float64
	totalWrites        int
	successWrites      int
	writeE             float64
	prepE              float64
	settleE            float64
	startLevel         int
	saturatedDirection int // -1 for negative sat, +1 for positive sat, 0 for not saturated
}

type headlessEngineState struct {
	time         float64
	temperature  float64
	polarization float64
}

func headlessLevelFromConductance(g, gmin, gmax float64, numLevels int) (int, int) {
	if numLevels <= 1 || gmax <= gmin {
		return 0, 1
	}
	ratio := (g - gmin) / (gmax - gmin)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	idx := int(math.Round(ratio * float64(numLevels-1)))
	if idx < 0 {
		idx = 0
	} else if idx > numLevels-1 {
		idx = numLevels - 1
	}
	return idx, idx + 1
}

func headlessStateBand(levelIndex int, numLevels int) string {
	if numLevels <= 0 {
		return "UNKNOWN"
	}
	lowThird := numLevels / 3
	highThird := numLevels * 2 / 3
	if levelIndex < lowThird {
		return "NEGATIVE"
	}
	if levelIndex >= highThird {
		return "POSITIVE"
	}
	return "INTERMEDIATE"
}

func headlessWRDPhaseName(phase int) string {
	switch phase {
	case 0:
		return "PREP"
	case 1:
		return "SETTLE"
	case 2:
		// In headless WRD we delegate programming + verify to WriteController.
		// This phase represents the active write loop.
		return "WRITE"
	case 3:
		return "HOLD"
	case 4:
		return "READBACK"
	case 5:
		return "RESULT"
	case 6:
		return "RETRY"
	default:
		return "UNKNOWN"
	}
}

func headlessPhaseForState(state, waveform string) (int, string) {
	if strings.HasSuffix(waveform, "_SWEEP") {
		return -1, waveform
	}
	switch state {
	case "RESETTING":
		return 0, "RESET"
	case "VERIFY":
		return 4, "READ"
	case "SUCCESS":
		return 5, "DISPLAY"
	case "FAILED", "FORCE_RESET":
		return 5, "FAILED"
	case "IDLE":
		return -1, "IDLE"
	default:
		return 2, "WRITE"
	}
}

func recordHeadlessSnapshot(
	logger *hysheadless.HysteresisDataLogger,
	state headlessEngineState,
	mat *physics.HZOMaterial,
	effPs float64,
	gmin float64,
	gmax float64,
	numLevels int,
	ctrl *controller.WriteController,
	dt float64,
	eField float64,
	waveform string,
	stateOverride string,
	wrd *headlessWRDState,
) {
	if logger == nil || mat == nil {
		return
	}

	currentP := state.polarization
	if effPs == 0 {
		rangeFrac := mat.TargetRangeFrac
		if rangeFrac <= 0 || rangeFrac > 1 {
			rangeFrac = 1
		}
		effPs = mat.Ps * rangeFrac
	}
	currentG := physics.PolarizationToConductance(currentP, effPs, gmin, gmax)
	levelIdx, level := headlessLevelFromConductance(currentG, gmin, gmax, numLevels)
	normalizedP := 0.0
	if effPs != 0 {
		normalizedP = currentP / effPs
		if normalizedP < -1 {
			normalizedP = -1
		} else if normalizedP > 1 {
			normalizedP = 1
		}
	}

	stateLabel := stateOverride
	if ctrl != nil {
		stateLabel = ctrl.State.String()
	}

	wrdPhase := -1
	wrdPhaseName := ""
	wrdPhaseTimer := 0.0
	wrdTargetLevel := 0
	wrdReadLevel := level
	wrdRetryCount := 0
	wrdCycleEnergy := 0.0
	wrdTotalWrites := 0
	wrdSuccessWrites := 0
	wrdWriteE := 0.0
	wrdPrepE := 0.0
	wrdSettleE := 0.0
	wrdStartLevel := 0
	if wrd != nil {
		wrdPhase = wrd.phase
		wrdPhaseName = headlessWRDPhaseName(wrd.phase)
		wrdPhaseTimer = wrd.phaseTimer
		wrdTargetLevel = wrd.targetLevel
		wrdReadLevel = wrd.readLevel
		wrdRetryCount = wrd.retryCount
		wrdCycleEnergy = wrd.cycleEnergy
		wrdTotalWrites = wrd.totalWrites
		wrdSuccessWrites = wrd.successWrites
		wrdWriteE = wrd.writeE
		wrdPrepE = wrd.prepE
		wrdSettleE = wrd.settleE
		wrdStartLevel = wrd.startLevel
	} else {
		wrdPhase, wrdPhaseName = headlessPhaseForState(stateLabel, waveform)
	}

	snapshot := hysheadless.HysteresisSnapshot{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		SimTime:       state.time,
		Dt:            dt,
		Waveform:      waveform,
		AutoMode:      true,
		Material:      mat.Name,
		TemperatureK:  state.temperature,
		EcMVcm:        mat.Ec / headlessMVPerCM,
		PsUcCm2:       mat.Ps * headlessUCPerCM2,
		PrUcCm2:       mat.Pr * headlessUCPerCM2,
		NumLevels:     numLevels,
		LevelIndex:    levelIdx,
		Level:         level,
		StateBand:     headlessStateBand(levelIdx, numLevels),
		EField:        eField,
		EFieldMVcm:    eField / headlessMVPerCM,
		Polarization:  currentP,
		PolarizationU: currentP * headlessUCPerCM2,
		NormalizedP:   normalizedP,

		WrdPhase:       wrdPhase,
		WrdPhaseName:   wrdPhaseName,
		WrdPhaseTimer:  wrdPhaseTimer,
		WrdTargetLevel: wrdTargetLevel,
		WrdReadLevel:   wrdReadLevel,
		WrdRetryCount:  wrdRetryCount,
		WrdCycleEnergy: wrdCycleEnergy,
		WrdTotalWrites: wrdTotalWrites,
		WrdSuccess:     wrdSuccessWrites,
		WrdWriteE:      wrdWriteE,
		WrdPrepE:       wrdPrepE,
		WrdSettleE:     wrdSettleE,
		WrdStartLevel:  wrdStartLevel,

		ControllerState:          stateLabel,
		ControllerPhaseTimer:     0,
		ControllerTargetLevel:    0,
		ControllerCurrentField:   0,
		ControllerCurrentFieldMV: 0,
		ControllerPulseCount:     0,
		ControllerTotalPulses:    0,
		ControllerRetryCount:     0,
		ControllerOvershootCount: 0,
		ControllerOvershootTotal: 0,
		ControllerLastVerify:     0,
		ControllerLastError:      0,
		ControllerVMin:           0,
		ControllerVMax:           0,
		ControllerInitialLevel:   0,
		ControllerFromSaturation: false,
		ControllerResetDirection: 0,
	}

	if ctrl != nil {
		snapshot.ControllerPhaseTimer = ctrl.PhaseTimer
		snapshot.ControllerTargetLevel = ctrl.TargetLevel
		snapshot.ControllerCurrentField = ctrl.CurrentField
		snapshot.ControllerCurrentFieldMV = ctrl.CurrentField / headlessMVPerCM
		snapshot.ControllerPulseCount = ctrl.PulseCount
		snapshot.ControllerTotalPulses = ctrl.TotalPulses
		snapshot.ControllerRetryCount = ctrl.RetryCount
		snapshot.ControllerOvershootCount = ctrl.OvershootCount
		snapshot.ControllerOvershootTotal = ctrl.OvershootTotal
		snapshot.ControllerLastVerify = ctrl.LastVerifyLevel
		snapshot.ControllerLastError = ctrl.LastError
		snapshot.ControllerVMin = ctrl.VMin
		snapshot.ControllerVMax = ctrl.VMax
		snapshot.ControllerInitialLevel = ctrl.InitialLevel
		snapshot.ControllerFromSaturation = ctrl.FromSaturation
		snapshot.ControllerResetDirection = ctrl.ResetDirection()

		if ctrl.EcField != 0 {
			snapshot.ControllerVMinEc = ctrl.VMin / ctrl.EcField
			snapshot.ControllerVMaxEc = ctrl.VMax / ctrl.EcField
		}
	}

	logger.Record(snapshot)
}
