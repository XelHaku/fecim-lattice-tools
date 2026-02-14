package gui

import (
	"math"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// calibrateLevelsAtTemperature performs calibration at a specific temperature.
// Sets Preisach temperature before calibrating, stores result in cache.
// MUST be called with a.mu held.
func (a *App) calibrateLevelsAtTemperature(tempK float64) {
	if a.material == nil {
		return
	}
	// Allow fast cancellation when module view/window switches and Stop() is called.
	if !a.running.Load() {
		return
	}

	// Set engine temperature before calibrating
	if a.useLKSolver() {
		if a.lkSolver != nil {
			a.lkSolver.Temperature = tempK
		}
	} else if a.preisach != nil {
		a.preisach.SetTemperature(tempK)
	}
	a.calibrationTemp = tempK

	// Perform calibration
	a.calibrateLevels()

	if !a.running.Load() {
		return
	}

	// Store in cache
	tempKRounded := int(math.Round(tempK))
	if a.tempCalibrations == nil {
		a.tempCalibrations = make(map[int]*TempCalibration)
	}
	a.tempCalibrations[tempKRounded] = &TempCalibration{
		Temperature:     tempK,
		CalibrationUp:   append([]float64(nil), a.calibrationUp...),
		CalibrationDown: append([]float64(nil), a.calibrationDown...),
		CalibUpLow:      append([]float64(nil), a.calibUpLow...),
		CalibUpHigh:     append([]float64(nil), a.calibUpHigh...),
		CalibDownLow:    append([]float64(nil), a.calibDownLow...),
		CalibDownHigh:   append([]float64(nil), a.calibDownHigh...),
		LastErrorUp:     append([]int(nil), a.lastErrorUp...),
		LastErrorDown:   append([]int(nil), a.lastErrorDown...),
		RelaxCompUp:     append([]float64(nil), a.relaxCompUp...),
		RelaxCompDown:   append([]float64(nil), a.relaxCompDown...),
	}

	log.Printf("Level calibration complete for material: %s at %.0fK", a.material.Name, tempK)
}

// calibrateLevels performs a calibration sweep to map field→level relationship.
// This mimics how real ferroelectric memory controllers characterize each device
// and build lookup tables for programming. Called at startup and when material changes.
// MUST be called with a.mu held.
//
// FIXED: Uses binary search with fresh saturation for each level to avoid
// Preisach state corruption from the previous sweep-based approach.
func (a *App) calibrateLevels() {
	if a.material == nil {
		return
	}
	// Allow fast cancellation when module view/window switches and Stop() is called.
	if !a.running.Load() {
		return
	}
	if a.useLKSolver() {
		a.calibrateLevelsLK()
		return
	}
	if a.preisach == nil {
		return
	}

	// Use temperature-corrected Ec from Preisach model
	Ec := a.preisach.GetEffectiveEc()
	if Ec == 0 {
		// Fallback to material Ec if temperature correction returns 0
		Ec = a.material.Ec
	}
	Emax := 2.0 * Ec // Go well beyond saturation for reliable calibration
	numLevels := a.numLevels
	maxLevel := numLevels - 1
	if maxLevel < 1 {
		maxLevel = 1 // Prevent division by zero when numLevels=1
	}

	// Record calibration temperature
	a.calibrationTemp = a.preisach.Temperature

	// Resize calibration arrays if needed (check all arrays to handle partial initialization)
	if len(a.calibrationUp) != numLevels {
		a.calibrationUp = make([]float64, numLevels)
	}
	if len(a.calibrationDown) != numLevels {
		a.calibrationDown = make([]float64, numLevels)
	}
	if len(a.calibUpLow) != numLevels {
		a.calibUpLow = make([]float64, numLevels)
	}
	if len(a.calibUpHigh) != numLevels {
		a.calibUpHigh = make([]float64, numLevels)
	}
	if len(a.calibDownLow) != numLevels {
		a.calibDownLow = make([]float64, numLevels)
	}
	if len(a.calibDownHigh) != numLevels {
		a.calibDownHigh = make([]float64, numLevels)
	}
	if len(a.lastErrorUp) != numLevels {
		a.lastErrorUp = make([]int, numLevels)
	}
	if len(a.lastErrorDown) != numLevels {
		a.lastErrorDown = make([]int, numLevels)
	}
	if len(a.relaxCompUp) != numLevels {
		a.relaxCompUp = make([]float64, numLevels)
	}
	if len(a.relaxCompDown) != numLevels {
		a.relaxCompDown = make([]float64, numLevels)
	}

	// Initialize bounds based on temperature-corrected Ec
	for i := 0; i < numLevels; i++ {
		// Initial bounds: full range (will be narrowed by runtime feedback)
		a.calibUpLow[i] = Ec * 0.3
		a.calibUpHigh[i] = Ec * 2.0
		a.calibDownLow[i] = -Ec * 2.0
		a.calibDownHigh[i] = -Ec * 0.3

		// No static relaxation compensation - not needed in quasistatic Preisach simulation.
		// Switching time (10ns) is negligible vs pulse duration (300ms), so no drift occurs.
		// The adaptive runtime feedback system handles any corrections needed.
		a.relaxCompUp[i] = 0.0
		a.relaxCompDown[i] = 0.0
	}

	// Normalize using effective Ps to match runtime discrete-level mapping
	effPs := a.effectivePsForLevels()
	if effPs <= 0 {
		effPs = a.material.Pr
	}
	if effPs <= 0 {
		effPs = 1.0
	}

	// Helper function to test what level results from a given field
	// starting from negative saturation (for ascending calibration)
	testLevelAscending := func(testE float64) int {
		if !a.running.Load() {
			return 0
		}
		// Reset and saturate negative
		a.preisach.Reset()
		for i := 0; i < 50; i++ {
			if !a.running.Load() {
				return 0
			}
			a.preisach.Update(-Emax)
		}
		a.preisach.Update(0) // At level 1 (negative remanent)

		// Apply test field and return to zero
		if !a.running.Load() {
			return 0
		}
		a.preisach.Update(testE)
		p := a.preisach.Update(0)

		// Normalize by Ps to match runtime discrete-level mapping
		normalizedP := p / effPs
		if normalizedP > 1.0 {
			normalizedP = 1.0
		}
		if normalizedP < -1.0 {
			normalizedP = -1.0
		}
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		}
		if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	// Helper function for descending calibration (from positive saturation)
	testLevelDescending := func(testE float64) int {
		if !a.running.Load() {
			return 0
		}
		// Reset and saturate positive
		a.preisach.Reset()
		for i := 0; i < 50; i++ {
			if !a.running.Load() {
				return 0
			}
			a.preisach.Update(Emax)
		}
		a.preisach.Update(0) // At level N (positive remanent)

		// Apply test field (negative) and return to zero
		if !a.running.Load() {
			return 0
		}
		a.preisach.Update(testE)
		p := a.preisach.Update(0)

		// Normalize by Ps to match runtime discrete-level mapping
		normalizedP := p / effPs
		if normalizedP > 1.0 {
			normalizedP = 1.0
		}
		if normalizedP < -1.0 {
			normalizedP = -1.0
		}
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		}
		if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	// Calibrate ASCENDING using binary search for each level
	// Start with initial estimates based on linear interpolation
	for targetLevel := 1; targetLevel < numLevels; targetLevel++ {
		if !a.running.Load() {
			return
		}
		// Initial estimate: linear interpolation between Ec and 2*Ec
		ratio := float64(targetLevel) / float64(maxLevel)
		initialGuess := Ec * (0.8 + ratio*1.2) // Range: 0.8*Ec to 2.0*Ec

		// Binary search to find exact field - use full bounds range
		lowE := Ec * 0.3 // Match initialized bounds for full coverage
		highE := Emax
		bestE := initialGuess
		bestDiff := numLevels // Start with worst case

		// Binary search with 15 iterations (precision: ~0.003% of range)
		for iter := 0; iter < 15; iter++ {
			if !a.running.Load() {
				return
			}
			midE := (lowE + highE) / 2
			resultLevel := testLevelAscending(midE)

			diff := resultLevel - targetLevel
			if abs(diff) < abs(bestDiff) {
				bestDiff = diff
				bestE = midE
			}

			if resultLevel == targetLevel {
				// Found exact match
				break
			} else if resultLevel < targetLevel {
				// Need higher field
				lowE = midE
			} else {
				// Need lower field
				highE = midE
			}
		}

		a.calibrationUp[targetLevel] = bestE
	}
	// Level 0 is at negative saturation, no field needed from that state
	a.calibrationUp[0] = 0

	// Calibrate DESCENDING using binary search for each level
	for targetLevel := maxLevel - 1; targetLevel >= 0; targetLevel-- {
		if !a.running.Load() {
			return
		}
		// Initial estimate: linear interpolation between -Ec and -2*Ec
		ratio := float64(maxLevel-targetLevel) / float64(maxLevel)
		initialGuess := -Ec * (0.8 + ratio*1.2) // Range: -0.8*Ec to -2.0*Ec

		// Binary search to find exact field (negative values) - use full bounds range
		lowE := -Emax      // More negative
		highE := -Ec * 0.3 // Less negative - match initialized bounds
		bestE := initialGuess
		bestDiff := numLevels

		for iter := 0; iter < 15; iter++ {
			if !a.running.Load() {
				return
			}
			midE := (lowE + highE) / 2
			resultLevel := testLevelDescending(midE)

			diff := resultLevel - targetLevel
			if abs(diff) < abs(bestDiff) {
				bestDiff = diff
				bestE = midE
			}

			if resultLevel == targetLevel {
				break
			} else if resultLevel > targetLevel {
				// Need more negative field (lower E)
				highE = midE
			} else {
				// Need less negative field (higher E)
				lowE = midE
			}
		}

		a.calibrationDown[targetLevel] = bestE
	}
	// Level maxLevel is at positive saturation, no field needed from that state
	a.calibrationDown[maxLevel] = 0

	a.finalizeCalibration(Ec)
}

// calibrateLevelsLK performs calibration using the Landau-Khalatnikov solver.
// It runs the same WriteController loop against the L-K dynamics to derive per-level fields.
// MUST be called with a.mu held.
func (a *App) calibrateLevelsLK() {
	if a.material == nil {
		return
	}
	if !a.running.Load() {
		return
	}

	numLevels := a.numLevels
	if numLevels < 2 {
		numLevels = 2
	}
	maxLevel := numLevels - 1

	// Ensure calibration arrays are sized correctly
	if len(a.calibrationUp) != numLevels {
		a.calibrationUp = make([]float64, numLevels)
	}
	if len(a.calibrationDown) != numLevels {
		a.calibrationDown = make([]float64, numLevels)
	}
	if len(a.calibUpLow) != numLevels {
		a.calibUpLow = make([]float64, numLevels)
	}
	if len(a.calibUpHigh) != numLevels {
		a.calibUpHigh = make([]float64, numLevels)
	}
	if len(a.calibDownLow) != numLevels {
		a.calibDownLow = make([]float64, numLevels)
	}
	if len(a.calibDownHigh) != numLevels {
		a.calibDownHigh = make([]float64, numLevels)
	}
	if len(a.lastErrorUp) != numLevels {
		a.lastErrorUp = make([]int, numLevels)
	}
	if len(a.lastErrorDown) != numLevels {
		a.lastErrorDown = make([]int, numLevels)
	}
	if len(a.relaxCompUp) != numLevels {
		a.relaxCompUp = make([]float64, numLevels)
	}
	if len(a.relaxCompDown) != numLevels {
		a.relaxCompDown = make([]float64, numLevels)
	}

	Ec := a.material.Ec
	if Ec == 0 {
		Ec = 1e8
	}
	for i := 0; i < numLevels; i++ {
		a.calibUpLow[i] = Ec * 0.3
		a.calibUpHigh[i] = Ec * 2.0
		a.calibDownLow[i] = -Ec * 2.0
		a.calibDownHigh[i] = -Ec * 0.3
		a.relaxCompUp[i] = 0
		a.relaxCompDown[i] = 0
		a.lastErrorUp[i] = 0
		a.lastErrorDown[i] = 0
	}

	// Local solver for calibration (deterministic, no noise)
	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(a.material)
	solver.Temperature = a.currentTemperature()
	solver.EnableNoise = false
	solver.UseNLS = false
	if !solver.UseMaterialAlpha {
		solver.UpdateParams()
	}

	// Pulse timing (match headless defaults)
	pulseDuration := a.material.Tau
	if pulseDuration <= 0 {
		pulseDuration = 10e-9
	}

	dtNominal := 1e-4
	dtMin := 1e-6
	dtMax := 0.025
	stableNominal := pulseDuration / 10000.0
	if stableNominal > 0 && stableNominal < dtNominal {
		dtNominal = stableNominal
	}
	if dtNominal <= 0 {
		dtNominal = 1e-12
	}
	if dtMin > dtNominal {
		dtMin = dtNominal
	}
	if pulseDuration > 0 && pulseDuration < dtMax {
		dtMax = pulseDuration
	}

	// Level mapping from polarization (effective Ps for range back-off)
	effPs := a.effectivePsForLevels()
	if effPs == 0 {
		effPs = a.material.Pr
	}
	if effPs == 0 {
		effPs = 1.0
	}

	levelFromP := func(P float64) int {
		normalizedP := P / effPs
		if normalizedP > 1.0 {
			normalizedP = 1.0
		}
		if normalizedP < -1.0 {
			normalizedP = -1.0
		}
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		}
		if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	satP := effPs
	if satP == 0 {
		satP = 0.3
	}

	runWrite := func(targetIdx int, startPol float64) (float64, bool) {
		wc := controller.NewWriteController(numLevels, Ec, Ec*2.5, nil)
		wc.PulseDuration = pulseDuration
		wc.Start(targetIdx+1, true)

		solver.SetState(startPol)
		solver.Time = 0

		currentField := 0.0
		elapsed := 0.0
		maxSimTime := pulseDuration * float64(wc.MaxRetries+100)

		for elapsed < maxSimTime {
			if !a.running.Load() {
				return wc.CurrentField, false
			}
			currentLevel := levelFromP(solver.GetState()) + 1
			step := dtNominal
			if Ec > 0 {
				distPlus := math.Abs(currentField - Ec)
				distMinus := math.Abs(currentField + Ec)
				if math.Min(distPlus, distMinus) < 1e7 {
					step = dtMin
				}
			}
			if step > dtMax {
				step = dtMax
			}

			targetField, done := wc.Update(step, currentField, currentLevel, 0)
			currentField = targetField
			solver.Step(currentField, step)
			elapsed += step

			if done {
				return wc.CurrentField, wc.State == controller.StateSuccess
			}
		}

		return wc.CurrentField, false
	}

	// Ascending calibration (from negative saturation)
	for idx := 1; idx < numLevels; idx++ {
		if !a.running.Load() {
			return
		}
		field, ok := runWrite(idx, -math.Abs(satP))
		if ok {
			a.calibrationUp[idx] = field
		} else {
			a.calibrationUp[idx] = Ec
		}
	}
	a.calibrationUp[0] = 0

	// Descending calibration (from positive saturation)
	for idx := maxLevel - 1; idx >= 0; idx-- {
		if !a.running.Load() {
			return
		}
		field, ok := runWrite(idx, math.Abs(satP))
		if ok {
			a.calibrationDown[idx] = field
		} else {
			a.calibrationDown[idx] = -Ec
		}
	}
	a.calibrationDown[maxLevel] = 0

	a.finalizeCalibration(Ec)
}

// finalizeCalibration enforces monotonicity, resets state, and syncs to CalibrationManager.
// MUST be called with a.mu held.
func (a *App) finalizeCalibration(Ec float64) {
	numLevels := a.numLevels
	if numLevels <= 0 {
		numLevels = 1
	}

	// MONOTONICITY ENFORCEMENT: Fix any non-monotonic values (spikes)
	// For ascending (calibrationUp): E-field must increase with level
	// For descending (calibrationDown): E-field must decrease (more negative) with decreasing level

	// Fix ascending calibration: ensure calibrationUp[i] <= calibrationUp[i+1]
	for i := 1; i < numLevels-1; i++ {
		if a.calibrationUp[i+1] < a.calibrationUp[i] {
			// Spike detected - interpolate from neighbors
			nextValid := a.calibrationUp[i]
			for j := i + 1; j < numLevels; j++ {
				if a.calibrationUp[j] > a.calibrationUp[i] {
					nextValid = a.calibrationUp[j]
					break
				}
			}
			a.calibrationUp[i+1] = (a.calibrationUp[i] + nextValid) / 2
		}
	}

	// Second pass: ensure strict monotonicity from bottom up
	for i := 1; i < numLevels; i++ {
		if a.calibrationUp[i] <= a.calibrationUp[i-1] {
			step := Ec * 0.02 // 2% of Ec per level minimum step
			a.calibrationUp[i] = a.calibrationUp[i-1] + step
		}
	}

	// Fix descending calibration: ensure calibrationDown[i] >= calibrationDown[i-1] (less negative for higher levels)
	for i := numLevels - 2; i > 0; i-- {
		if a.calibrationDown[i-1] > a.calibrationDown[i] {
			prevValid := a.calibrationDown[i]
			for j := i - 1; j >= 0; j-- {
				if a.calibrationDown[j] < a.calibrationDown[i] {
					prevValid = a.calibrationDown[j]
					break
				}
			}
			a.calibrationDown[i-1] = (a.calibrationDown[i] + prevValid) / 2
		}
	}

	// Second pass: ensure strict monotonicity from top down
	for i := numLevels - 2; i >= 0; i-- {
		if a.calibrationDown[i] >= a.calibrationDown[i+1] {
			step := Ec * 0.02 // 2% of Ec per level minimum step
			a.calibrationDown[i] = a.calibrationDown[i+1] - step
		}
	}

	// Reset physics state to neutral after calibration
	if a.useLKSolver() {
		if a.lkSolver != nil {
			resetP := a.lkDefaultPolarization()
			a.lkSolver.SetState(resetP)
			a.lkSolver.Time = 0
			a.polarization = a.lkSolver.GetState()
		} else {
			a.polarization = a.lkDefaultPolarization()
		}
	} else if a.preisach != nil {
		a.preisach.Reset()
		a.polarization = 0
	}
	if !a.useLKSolver() && a.preisach == nil {
		a.polarization = 0
	}
	a.electricField = 0
	a.normalizedP = 0
	a.syncDiscreteLevelLocked()

	// Sync to CalibrationManager (Refactoring)
	if a.calibManager != nil {
		a.calibManager.CalibrationUp = append([]float64(nil), a.calibrationUp...)
		a.calibManager.CalibrationDown = append([]float64(nil), a.calibrationDown...)
		a.calibManager.CalibUpLow = append([]float64(nil), a.calibUpLow...)
		a.calibManager.CalibUpHigh = append([]float64(nil), a.calibUpHigh...)
		a.calibManager.CalibDownLow = append([]float64(nil), a.calibDownLow...)
		a.calibManager.CalibDownHigh = append([]float64(nil), a.calibDownHigh...)
		a.calibManager.LastErrorUp = append([]int(nil), a.lastErrorUp...)
		a.calibManager.LastErrorDown = append([]int(nil), a.lastErrorDown...)
		a.calibManager.RelaxCompUp = append([]float64(nil), a.relaxCompUp...)
		a.calibManager.RelaxCompDown = append([]float64(nil), a.relaxCompDown...)
	}

	a.calibrated = true
}

// enforceMonotonicityUp ensures calibrationUp[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
// MUST be called with a.mu held.
func (a *App) enforceMonotonicityUp(idx int, Ec float64) {
	if idx <= 0 || idx >= len(a.calibrationUp) {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is greater than previous level
	if a.calibrationUp[idx] <= a.calibrationUp[idx-1] {
		a.calibrationUp[idx] = a.calibrationUp[idx-1] + step
	}

	// Ensure value is less than next level (if not at end)
	if idx < len(a.calibrationUp)-1 && a.calibrationUp[idx] >= a.calibrationUp[idx+1] {
		// Push next value up to maintain monotonicity
		a.calibrationUp[idx+1] = a.calibrationUp[idx] + step
	}
}

// enforceMonotonicityDown ensures calibrationDown[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
// MUST be called with a.mu held.
func (a *App) enforceMonotonicityDown(idx int, Ec float64) {
	if idx <= 0 || idx >= len(a.calibrationDown)-1 {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is more negative than next level (lower index = more negative)
	if a.calibrationDown[idx] >= a.calibrationDown[idx+1] {
		a.calibrationDown[idx] = a.calibrationDown[idx+1] - step
	}

	// Ensure value is less negative than previous level
	if idx > 0 && a.calibrationDown[idx] <= a.calibrationDown[idx-1] {
		// Push previous value down to maintain monotonicity
		a.calibrationDown[idx-1] = a.calibrationDown[idx] - step
	}
}

// updateCalibrationUp updates ascending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
// MUST be called with a.mu held.
func (a *App) updateCalibrationUp(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(a.calibrationUp) {
		return
	}

	currentE := a.calibrationUp[targetIdx]
	lastErr := a.lastErrorUp[targetIdx]

	// Update bounds based on error direction
	if levelError > 0 {
		// Overshot (read level > target) → field too strong → update upper bound
		if currentE < a.calibUpHigh[targetIdx] {
			a.calibUpHigh[targetIdx] = currentE
		}
	} else {
		// Undershot (read level < target) → field too weak → update lower bound
		if currentE > a.calibUpLow[targetIdx] {
			a.calibUpLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if a.calibUpLow[targetIdx] < a.calibUpHigh[targetIdx] {
		newVal = (a.calibUpLow[targetIdx] + a.calibUpHigh[targetIdx]) / 2
	} else {
		// Bounds crossed - use direct error-proportional adjustment
		// This is more aggressive than midpoint to escape stuck states
		adjustment := float64(levelError) * Ec * 0.05 // 5% per level error
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Soft constraints: guide towards reasonable range but don't hard-clamp
	// This allows the binary search to explore beyond if physics requires it
	minE := Ec * 0.3 // Absolute minimum: 0.3×Ec
	maxE := Ec * 2.5 // Absolute maximum: 2.5×Ec

	if newVal < minE {
		newVal = minE
	} else if newVal > maxE {
		newVal = maxE
	}

	oldVal := a.calibrationUp[targetIdx]
	a.calibrationUp[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For ascending calibration: higher indices = higher fields
	//   calibrationUp[0] lowest, calibrationUp[29] highest
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal < oldVal {
		// Made field LOWER - cascade to lower indices (they must be even lower)
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := a.calibrationUp[i+1] - step // Must be lower than next
			if a.calibrationUp[i] > maxAllowed {
				a.calibrationUp[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal > oldVal {
		// Made field HIGHER - cascade to higher indices (they must be even higher)
		for i := targetIdx + 1; i < len(a.calibrationUp); i++ {
			minAllowed := a.calibrationUp[i-1] + step // Must be higher than previous
			if a.calibrationUp[i] < minAllowed {
				a.calibrationUp[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call enforceMonotonicityUp - cascade handles it and won't fight with newVal
	a.lastErrorUp[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(a.relaxCompUp) {
		relaxAdjust := 0.02 // 2% target adjustment per retry (more aggressive for faster convergence)
		oldRelaxComp := a.relaxCompUp[targetIdx]
		var targetRelax float64
		if levelError > 0 {
			// Overshot: read level > target, reduce compensation (we overshot too much)
			targetRelax = oldRelaxComp - relaxAdjust*float64(levelError) // Scale by error magnitude
		} else {
			// Undershot: read level < target, increase compensation (need more overshoot)
			targetRelax = oldRelaxComp - relaxAdjust*float64(levelError) // levelError is negative, so this adds
		}
		// Exponential moving average: 70% old + 30% new for smooth convergence
		a.relaxCompUp[targetIdx] = 0.7*oldRelaxComp + 0.3*targetRelax
		// Clamp to reasonable bounds [-0.05, 0.25]
		if a.relaxCompUp[targetIdx] < -0.05 {
			a.relaxCompUp[targetIdx] = -0.05
		} else if a.relaxCompUp[targetIdx] > 0.25 {
			a.relaxCompUp[targetIdx] = 0.25
		}
		if a.relaxCompUp[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_UP[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, a.relaxCompUp[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_UP[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		a.calibUpLow[targetIdx]/1e8, a.calibUpHigh[targetIdx]/1e8)

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationUp, a.calibrationUp)
	}
}

// updateCalibrationDown updates descending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
// MUST be called with a.mu held.
func (a *App) updateCalibrationDown(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(a.calibrationDown) {
		return
	}

	currentE := a.calibrationDown[targetIdx]
	lastErr := a.lastErrorDown[targetIdx]

	// Update bounds based on error direction (descending uses negative fields)
	// calibDownLow = most negative limit (e.g., -2.5×Ec)
	// calibDownHigh = least negative limit (e.g., -0.3×Ec)
	// midpoint searches between them
	if levelError > 0 {
		// Positive error = didn't go down enough = field not negative enough
		// Current field is too weak → becomes new UPPER bound (least negative limit)
		// Next search will be between Low and currentE (more negative)
		if currentE < a.calibDownHigh[targetIdx] {
			a.calibDownHigh[targetIdx] = currentE
		}
	} else {
		// Negative error = went too far down = field too negative
		// Current field is too strong → becomes new LOWER bound (most negative limit)
		// Next search will be between currentE and High (less negative)
		if currentE > a.calibDownLow[targetIdx] {
			a.calibDownLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if a.calibDownLow[targetIdx] < a.calibDownHigh[targetIdx] {
		newVal = (a.calibDownLow[targetIdx] + a.calibDownHigh[targetIdx]) / 2
	} else {
		// Bounds crossed - use direct error-proportional adjustment
		// For descending, positive error means we need MORE negative field
		adjustment := float64(levelError) * Ec * 0.05 // 5% per level error
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Soft constraints: guide towards reasonable range but don't hard-clamp
	minE := -Ec * 2.5 // Absolute minimum (most negative): -2.5×Ec
	maxE := -Ec * 0.3 // Absolute maximum (least negative): -0.3×Ec

	if newVal > maxE {
		newVal = maxE
	} else if newVal < minE {
		newVal = minE
	}

	oldVal := a.calibrationDown[targetIdx]
	a.calibrationDown[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For descending calibration: lower indices = more negative (stronger field)
	//   calibrationDown[0] most negative, calibrationDown[29] least negative (or 0)
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal > oldVal {
		// Made field LESS negative (weaker) - cascade to higher indices
		// Higher indices must be even less negative to maintain monotonicity
		for i := targetIdx + 1; i < len(a.calibrationDown); i++ {
			minAllowed := a.calibrationDown[i-1] + step // Must be less negative than previous
			if a.calibrationDown[i] < minAllowed {
				a.calibrationDown[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal < oldVal {
		// Made field MORE negative (stronger) - cascade to lower indices
		// Lower indices must be even more negative to maintain monotonicity
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := a.calibrationDown[i+1] - step // Must be more negative than next
			if a.calibrationDown[i] > maxAllowed {
				a.calibrationDown[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call enforceMonotonicityDown - cascade handles it and won't fight with newVal
	a.lastErrorDown[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(a.relaxCompDown) {
		relaxAdjust := 0.02 // 2% target adjustment per retry (more aggressive for faster convergence)
		oldRelaxComp := a.relaxCompDown[targetIdx]
		var targetRelax float64
		if levelError < 0 {
			// Went too far down (error < 0), reduce compensation
			targetRelax = oldRelaxComp + relaxAdjust*float64(levelError) // levelError is negative, so this subtracts
		} else {
			// Didn't go down enough (error > 0), increase compensation
			targetRelax = oldRelaxComp + relaxAdjust*float64(levelError) // Scale by error magnitude
		}
		// Exponential moving average: 70% old + 30% new for smooth convergence
		a.relaxCompDown[targetIdx] = 0.7*oldRelaxComp + 0.3*targetRelax
		// Clamp to reasonable bounds [-0.05, 0.25]
		if a.relaxCompDown[targetIdx] < -0.05 {
			a.relaxCompDown[targetIdx] = -0.05
		} else if a.relaxCompDown[targetIdx] > 0.25 {
			a.relaxCompDown[targetIdx] = 0.25
		}
		if a.relaxCompDown[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_DOWN[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, a.relaxCompDown[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_DOWN[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		a.calibDownLow[targetIdx]/1e8, a.calibDownHigh[targetIdx]/1e8)

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationDown, a.calibrationDown)
	}
}
