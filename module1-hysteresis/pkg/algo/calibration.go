package algo

import (
	"log"
)

// CalibrationManager handles the storage and update logic for hysteresis loop calibration.
// It manages the voltage levels required to reach specific polarization states.
type CalibrationManager struct {
	NumLevels       int
	CalibrationUp   []float64 // Ascending calibration values
	CalibrationDown []float64 // Descending calibration values

	// Binary search bounds
	CalibUpLow    []float64
	CalibUpHigh   []float64
	CalibDownLow  []float64
	CalibDownHigh []float64

	// State for updates
	LastErrorUp   []int
	LastErrorDown []int
	RelaxCompUp   []float64
	RelaxCompDown []float64
}

// NewCalibrationManager creates a new manager with initialized arrays
func NewCalibrationManager(numLevels int) *CalibrationManager {
	return &CalibrationManager{
		NumLevels:       numLevels,
		CalibrationUp:   make([]float64, numLevels),
		CalibrationDown: make([]float64, numLevels),
		CalibUpLow:      make([]float64, numLevels),
		CalibUpHigh:     make([]float64, numLevels),
		CalibDownLow:    make([]float64, numLevels),
		CalibDownHigh:   make([]float64, numLevels),
		LastErrorUp:     make([]int, numLevels),
		LastErrorDown:   make([]int, numLevels),
		RelaxCompUp:     make([]float64, numLevels),
		RelaxCompDown:   make([]float64, numLevels),
	}
}

// UpdateCalibrationUp updates ascending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
func (cm *CalibrationManager) UpdateCalibrationUp(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(cm.CalibrationUp) {
		return
	}

	currentE := cm.CalibrationUp[targetIdx]
	lastErr := cm.LastErrorUp[targetIdx]

	// Update bounds based on error direction
	if levelError > 0 {
		// Overshot (read level > target) → field too strong → update upper bound
		if currentE < cm.CalibUpHigh[targetIdx] {
			cm.CalibUpHigh[targetIdx] = currentE
		}
	} else {
		// Undershot (read level < target) → field too weak → update lower bound
		if currentE > cm.CalibUpLow[targetIdx] {
			cm.CalibUpLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if cm.CalibUpLow[targetIdx] < cm.CalibUpHigh[targetIdx] {
		newVal = (cm.CalibUpLow[targetIdx] + cm.CalibUpHigh[targetIdx]) / 2
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

	oldVal := cm.CalibrationUp[targetIdx]
	cm.CalibrationUp[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For ascending calibration: higher indices = higher fields
	//   CalibrationUp[0] lowest, CalibrationUp[29] highest
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal < oldVal {
		// Made field LOWER - cascade to lower indices (they must be even lower)
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := cm.CalibrationUp[i+1] - step // Must be lower than next
			if cm.CalibrationUp[i] > maxAllowed {
				cm.CalibrationUp[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal > oldVal {
		// Made field HIGHER - cascade to higher indices (they must be even higher)
		for i := targetIdx + 1; i < len(cm.CalibrationUp); i++ {
			minAllowed := cm.CalibrationUp[i-1] + step // Must be higher than previous
			if cm.CalibrationUp[i] < minAllowed {
				cm.CalibrationUp[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call EnforceMonotonicityUp - cascade handles it and won't fight with newVal
	cm.LastErrorUp[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(cm.RelaxCompUp) {
		relaxAdjust := 0.02 // 2% target adjustment per retry (more aggressive for faster convergence)
		oldRelaxComp := cm.RelaxCompUp[targetIdx]
		var targetRelax float64
		if levelError > 0 {
			// Overshot: read level > target, reduce compensation (we overshot too much)
			targetRelax = oldRelaxComp - relaxAdjust*float64(levelError) // Scale by error magnitude
		} else {
			// Undershot: read level < target, increase compensation (need more overshoot)
			targetRelax = oldRelaxComp - relaxAdjust*float64(levelError) // levelError is negative, so this adds
		}
		// Exponential moving average: 70% old + 30% new for smooth convergence
		cm.RelaxCompUp[targetIdx] = 0.7*oldRelaxComp + 0.3*targetRelax
		// Clamp to reasonable bounds [-0.05, 0.25]
		if cm.RelaxCompUp[targetIdx] < -0.05 {
			cm.RelaxCompUp[targetIdx] = -0.05
		} else if cm.RelaxCompUp[targetIdx] > 0.25 {
			cm.RelaxCompUp[targetIdx] = 0.25
		}
		if cm.RelaxCompUp[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_UP[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, cm.RelaxCompUp[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_UP[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		cm.CalibUpLow[targetIdx]/1e8, cm.CalibUpHigh[targetIdx]/1e8)
}

// UpdateCalibrationDown updates descending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
func (cm *CalibrationManager) UpdateCalibrationDown(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(cm.CalibrationDown) {
		return
	}

	currentE := cm.CalibrationDown[targetIdx]
	lastErr := cm.LastErrorDown[targetIdx]

	// Update bounds based on error direction (descending uses negative fields)
	// CalibDownLow = most negative limit (e.g., -2.5×Ec)
	// CalibDownHigh = least negative limit (e.g., -0.3×Ec)
	// midpoint searches between them
	if levelError > 0 {
		// Positive error = didn't go down enough = field not negative enough
		// Current field is too weak → becomes new UPPER bound (least negative limit)
		// Next search will be between Low and currentE (more negative)
		if currentE < cm.CalibDownHigh[targetIdx] {
			cm.CalibDownHigh[targetIdx] = currentE
		}
	} else {
		// Negative error = went too far down = field too negative
		// Current field is too strong → becomes new LOWER bound (most negative limit)
		// Next search will be between currentE and High (less negative)
		if currentE > cm.CalibDownLow[targetIdx] {
			cm.CalibDownLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if cm.CalibDownLow[targetIdx] < cm.CalibDownHigh[targetIdx] {
		newVal = (cm.CalibDownLow[targetIdx] + cm.CalibDownHigh[targetIdx]) / 2
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

	oldVal := cm.CalibrationDown[targetIdx]
	cm.CalibrationDown[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For descending calibration: lower indices = more negative (stronger field)
	//   CalibrationDown[0] most negative, CalibrationDown[29] least negative (or 0)
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal > oldVal {
		// Made field LESS negative (weaker) - cascade to higher indices
		// Higher indices must be even less negative to maintain monotonicity
		for i := targetIdx + 1; i < len(cm.CalibrationDown); i++ {
			minAllowed := cm.CalibrationDown[i-1] + step // Must be less negative than previous
			if cm.CalibrationDown[i] < minAllowed {
				cm.CalibrationDown[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal < oldVal {
		// Made field MORE negative (stronger) - cascade to lower indices
		// Lower indices must be even more negative to maintain monotonicity
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := cm.CalibrationDown[i+1] - step // Must be more negative than next
			if cm.CalibrationDown[i] > maxAllowed {
				cm.CalibrationDown[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call EnforceMonotonicityDown - cascade handles it and won't fight with newVal
	cm.LastErrorDown[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(cm.RelaxCompDown) {
		relaxAdjust := 0.02 // 2% target adjustment per retry (more aggressive for faster convergence)
		oldRelaxComp := cm.RelaxCompDown[targetIdx]
		var targetRelax float64
		if levelError < 0 {
			// Went too far down (error < 0), reduce compensation
			targetRelax = oldRelaxComp + relaxAdjust*float64(levelError) // levelError is negative, so this subtracts
		} else {
			// Didn't go down enough (error > 0), increase compensation
			targetRelax = oldRelaxComp + relaxAdjust*float64(levelError) // Scale by error magnitude
		}
		// Exponential moving average: 70% old + 30% new for smooth convergence
		cm.RelaxCompDown[targetIdx] = 0.7*oldRelaxComp + 0.3*targetRelax
		// Clamp to reasonable bounds [-0.05, 0.25]
		if cm.RelaxCompDown[targetIdx] < -0.05 {
			cm.RelaxCompDown[targetIdx] = -0.05
		} else if cm.RelaxCompDown[targetIdx] > 0.25 {
			cm.RelaxCompDown[targetIdx] = 0.25
		}
		if cm.RelaxCompDown[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_DOWN[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, cm.RelaxCompDown[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_DOWN[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		cm.CalibDownLow[targetIdx]/1e8, cm.CalibDownHigh[targetIdx]/1e8)
}

// EnforceMonotonicityUp ensures CalibrationUp[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
func (cm *CalibrationManager) EnforceMonotonicityUp(idx int, Ec float64) {
	if idx <= 0 || idx >= len(cm.CalibrationUp) {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is greater than previous level
	if cm.CalibrationUp[idx] <= cm.CalibrationUp[idx-1] {
		cm.CalibrationUp[idx] = cm.CalibrationUp[idx-1] + step
	}

	// Ensure value is less than next level (if not at end)
	if idx < len(cm.CalibrationUp)-1 && cm.CalibrationUp[idx] >= cm.CalibrationUp[idx+1] {
		// Push next value up to maintain monotonicity
		cm.CalibrationUp[idx+1] = cm.CalibrationUp[idx] + step
	}
}

// EnforceMonotonicityDown ensures CalibrationDown[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
func (cm *CalibrationManager) EnforceMonotonicityDown(idx int, Ec float64) {
	if idx <= 0 || idx >= len(cm.CalibrationDown)-1 {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is more negative than next level (lower index = more negative)
	if cm.CalibrationDown[idx] >= cm.CalibrationDown[idx+1] {
		cm.CalibrationDown[idx] = cm.CalibrationDown[idx+1] - step
	}

	// Ensure value is less negative than previous level
	if idx > 0 && cm.CalibrationDown[idx] <= cm.CalibrationDown[idx-1] {
		// Push previous value down to maintain monotonicity
		cm.CalibrationDown[idx-1] = cm.CalibrationDown[idx] - step
	}
}
