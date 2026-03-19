// Package physics provides shared physics utilities for FeCIM simulations.

package physics

import "fecim-lattice-tools/shared/mathutil"

// CalibrationBounds tracks the binary search bounds for iterative calibration.
// Used by write-verify-retry loops to refine field values.
type CalibrationBounds struct {
	Low  float64 // Lower bound for binary search
	High float64 // Upper bound for binary search
}

// CalibrationState holds the state for a single level's calibration.
type CalibrationState struct {
	Value     float64 // Current calibrated field value
	Bounds    CalibrationBounds
	LastError int // Last error for oscillation detection
}

// Calibrator implements iterative calibration with binary search and
// write-verify-retry logic for ferroelectric memory levels.
type Calibrator struct {
	NumLevels  int
	Ec         float64 // Coercive field for this material
	MinStep    float64 // Minimum step between levels (default: 2% of Ec)
	MaxRetries int     // Maximum retries per target (default: 3)
	Tolerance  int     // Tolerance for "success" in levels (default: 1)

	// Ascending calibration (from negative saturation)
	Up []CalibrationState
	// Descending calibration (from positive saturation)
	Down []CalibrationState
}

// NewCalibrator creates a new calibrator for the given number of levels.
func NewCalibrator(numLevels int, Ec float64) *Calibrator {
	minStep := Ec * 0.02 // 2% of Ec
	c := &Calibrator{
		NumLevels:  numLevels,
		Ec:         Ec,
		MinStep:    minStep,
		MaxRetries: 3,
		Tolerance:  1,
		Up:         make([]CalibrationState, numLevels),
		Down:       make([]CalibrationState, numLevels),
	}

	// Initialize bounds to wide range
	Emax := 2.0 * Ec
	for i := 0; i < numLevels; i++ {
		c.Up[i].Bounds = CalibrationBounds{Low: 0, High: Emax}
		c.Down[i].Bounds = CalibrationBounds{Low: -Emax, High: 0}
	}

	return c
}

// UpdateAscending updates ascending calibration using binary search with bounds tracking.
// Returns the new field value to use for the next attempt.
func (c *Calibrator) UpdateAscending(levelIdx int, levelError int) float64 {
	if levelIdx < 0 || levelIdx >= c.NumLevels {
		return 0
	}

	state := &c.Up[levelIdx]
	currentE := state.Value

	// Update bounds based on error direction
	if levelError > 0 {
		// Overshot (read level > target) -> field too strong -> update upper bound
		if currentE < state.Bounds.High {
			state.Bounds.High = currentE
		}
	} else if levelError < 0 {
		// Undershot (read level < target) -> field too weak -> update lower bound
		if currentE > state.Bounds.Low {
			state.Bounds.Low = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if state.Bounds.Low < state.Bounds.High {
		newVal = (state.Bounds.Low + state.Bounds.High) / 2
	} else {
		// Bounds crossed - use small adjustment
		adjustment := float64(levelError) * c.Ec * 0.01
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if state.LastError != 0 && ((state.LastError > 0 && levelError < 0) || (state.LastError < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Level-dependent field constraints: higher levels need stronger fields
	minE, maxE := c.FieldConstraintsAscending(levelIdx)
	newVal = mathutil.Clamp(newVal, minE, maxE)

	state.Value = newVal
	state.LastError = levelError

	return newVal
}

// UpdateDescending updates descending calibration using binary search with bounds tracking.
// Returns the new field value to use for the next attempt.
func (c *Calibrator) UpdateDescending(levelIdx int, levelError int) float64 {
	if levelIdx < 0 || levelIdx >= c.NumLevels {
		return 0
	}

	state := &c.Down[levelIdx]
	currentE := state.Value

	// Update bounds based on error direction (descending uses negative fields)
	if levelError > 0 {
		// Overshot UP (didn't go down enough) -> field not negative enough -> update lower bound
		if currentE > state.Bounds.Low {
			state.Bounds.Low = currentE
		}
	} else if levelError < 0 {
		// Undershot (went too far down) -> field too negative -> update upper bound
		if currentE < state.Bounds.High {
			state.Bounds.High = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if state.Bounds.Low < state.Bounds.High {
		newVal = (state.Bounds.Low + state.Bounds.High) / 2
	} else {
		// Bounds crossed - use small adjustment
		adjustment := float64(levelError) * c.Ec * 0.01
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if state.LastError != 0 && ((state.LastError > 0 && levelError < 0) || (state.LastError < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Level-dependent field constraints: lower levels need more negative fields
	minE, maxE := c.FieldConstraintsDescending(levelIdx)
	newVal = mathutil.Clamp(newVal, minE, maxE)

	state.Value = newVal
	state.LastError = levelError

	return newVal
}

// FieldConstraintsAscending returns (minE, maxE) for ascending calibration.
// Higher levels need stronger (more positive) fields.
func (c *Calibrator) FieldConstraintsAscending(levelIdx int) (minE, maxE float64) {
	maxLevel := float64(c.NumLevels - 1)
	levelRatio := float64(levelIdx) / maxLevel
	minE = c.Ec * (0.4 + levelRatio*1.0) // Range: 0.4*Ec to 1.4*Ec
	maxE = c.Ec * (0.6 + levelRatio*1.2) // Range: 0.6*Ec to 1.8*Ec
	return
}

// FieldConstraintsDescending returns (minE, maxE) for descending calibration.
// Lower levels need stronger (more negative) fields.
func (c *Calibrator) FieldConstraintsDescending(levelIdx int) (minE, maxE float64) {
	maxLevel := float64(c.NumLevels - 1)
	levelRatio := float64(levelIdx) / maxLevel
	// Invert: low level index = strong negative field
	minE = -c.Ec * (0.6 + (1-levelRatio)*1.2) // Range: -1.8*Ec to -0.6*Ec
	maxE = -c.Ec * (0.4 + (1-levelRatio)*1.0) // Range: -1.4*Ec to -0.4*Ec
	return
}

// EnforceMonotonicityAscending ensures Up[idx] maintains local monotonicity.
// Higher indices should have higher (more positive) field values.
func (c *Calibrator) EnforceMonotonicityAscending(idx int) {
	if idx <= 0 || idx >= len(c.Up) {
		return
	}

	// Ensure value is greater than previous level
	if c.Up[idx].Value <= c.Up[idx-1].Value {
		c.Up[idx].Value = c.Up[idx-1].Value + c.MinStep
	}

	// Ensure value is less than next level (if not at end)
	if idx < len(c.Up)-1 && c.Up[idx].Value >= c.Up[idx+1].Value {
		c.Up[idx+1].Value = c.Up[idx].Value + c.MinStep
	}
}

// EnforceMonotonicityDescending ensures Down[idx] maintains local monotonicity.
// Lower indices should have more negative field values.
func (c *Calibrator) EnforceMonotonicityDescending(idx int) {
	if idx <= 0 || idx >= len(c.Down)-1 {
		return
	}

	// Ensure value is more negative than next level (lower index = more negative)
	if c.Down[idx].Value >= c.Down[idx+1].Value {
		c.Down[idx].Value = c.Down[idx+1].Value - c.MinStep
	}

	// Ensure value is less negative than previous level
	if idx > 0 && c.Down[idx].Value <= c.Down[idx-1].Value {
		c.Down[idx-1].Value = c.Down[idx].Value - c.MinStep
	}
}

// EnforceGlobalMonotonicity fixes all non-monotonic values in both directions.
func (c *Calibrator) EnforceGlobalMonotonicity() {
	// Ascending: ensure Up[i] < Up[i+1]
	for i := 1; i < len(c.Up); i++ {
		if c.Up[i].Value <= c.Up[i-1].Value {
			c.Up[i].Value = c.Up[i-1].Value + c.MinStep
		}
	}

	// Descending: ensure Down[i] > Down[i-1] (less negative)
	for i := len(c.Down) - 2; i >= 0; i-- {
		if c.Down[i].Value >= c.Down[i+1].Value {
			c.Down[i].Value = c.Down[i+1].Value - c.MinStep
		}
	}
}

// GetAscendingValues returns a copy of all ascending calibration values.
func (c *Calibrator) GetAscendingValues() []float64 {
	values := make([]float64, len(c.Up))
	for i, s := range c.Up {
		values[i] = s.Value
	}
	return values
}

// GetDescendingValues returns a copy of all descending calibration values.
func (c *Calibrator) GetDescendingValues() []float64 {
	values := make([]float64, len(c.Down))
	for i, s := range c.Down {
		values[i] = s.Value
	}
	return values
}

// SetAscendingValues sets ascending calibration values from a slice.
func (c *Calibrator) SetAscendingValues(values []float64) {
	for i := 0; i < len(c.Up) && i < len(values); i++ {
		c.Up[i].Value = values[i]
	}
}

// SetDescendingValues sets descending calibration values from a slice.
func (c *Calibrator) SetDescendingValues(values []float64) {
	for i := 0; i < len(c.Down) && i < len(values); i++ {
		c.Down[i].Value = values[i]
	}
}

// VerifyResult represents the result of a write-verify check.
type VerifyResult struct {
	Success     bool // True if read level matches target within tolerance
	ReadLevel   int  // The level that was read back
	Error       int  // readLevel - targetLevel
	ShouldRetry bool // True if should retry (not success and retries remaining)
}

// CheckVerify checks if a write operation succeeded and whether to retry.
func (c *Calibrator) CheckVerify(targetLevel, readLevel, retryCount int) VerifyResult {
	error := readLevel - targetLevel
	success := mathutil.AbsInt(error) <= c.Tolerance
	shouldRetry := !success && retryCount < c.MaxRetries

	return VerifyResult{
		Success:     success,
		ReadLevel:   readLevel,
		Error:       error,
		ShouldRetry: shouldRetry,
	}
}


