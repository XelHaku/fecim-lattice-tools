//go:build legacy_fyne

// Package gui — device_state_ispp.go
// ISPP (Incremental Step Pulse Programming) state machine with overshoot handling,
// and column-write (half-select) visualization overlay for DeviceState.
package gui

import (
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// ============================================================================
// ISPP STATE MACHINE WITH OVERSHOOT HANDLING
// ============================================================================

// ISPPEngine selects which ISPP implementation to use.
type ISPPEngine int

const (
	ISPPEngineLevel ISPPEngine = iota // Fast, level-based ISPP (legacy)
	ISPPEngineLK                      // Physics-based ISPP using L-K solver
)

func (e ISPPEngine) String() string {
	switch e {
	case ISPPEngineLevel:
		return "Preisach (Level-based)"
	case ISPPEngineLK:
		return "Landau-Khalatnikov (Physics ODE)"
	default:
		return "Unknown"
	}
}

// ISPPResult represents the result of an ISPP iteration
type ISPPResult int

const (
	ISPPResultContinue      ISPPResult = iota // Continue iterating
	ISPPResultVerified                        // Target level reached
	ISPPResultOvershoot                       // Overshoot detected, reset needed
	ISPPResultMaxIterations                   // Max iterations reached
	ISPPResultNotActive                       // ISPP was not active
)

// ISPP constants
const (
	ISPPMaxIterations = 40 // More pulses for finer convergence (matched to shared/physics)
)

// ISPPHistoryPoint stores one write-verify iteration snapshot.
type ISPPHistoryPoint struct {
	Iteration int
	Level     int
	Voltage   float64
}

// ISPPState holds the state of an active ISPP (Incremental Step Pulse Programming) loop
type ISPPState struct {
	Active       bool
	Iteration    int
	MaxIter      int
	TargetRow    int
	TargetCol    int
	TargetLevel  int
	CurrentLevel int
	Voltage      float64
	Direction    HysteresisDirection
	Verified     bool
	Complete     bool
	Success      bool
	History      []ISPPHistoryPoint
}

// StartISPP begins an ISPP loop for a cell
func (ds *DeviceState) StartISPP(row, col, targetLevel, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Early exit if already at target level
	if currentLevel == targetLevel {
		ds.isppState.Active = false
		ds.isppState.Iteration = 0
		ds.isppState.TargetLevel = targetLevel
		ds.isppState.CurrentLevel = currentLevel
		ds.isppState.Verified = true
		ds.isppState.Complete = true
		ds.isppState.Success = true
		ds.isppState.History = []ISPPHistoryPoint{{Iteration: 0, Level: currentLevel, Voltage: 0}}
		return
	}

	// Use shared ISPP calculator to determine direction
	sharedDirection := sharedphysics.GetDirection(currentLevel, targetLevel)

	// Map to local HysteresisDirection type
	var localDirection HysteresisDirection
	switch sharedDirection {
	case sharedphysics.DirectionAscending:
		localDirection = DirectionAscending
	case sharedphysics.DirectionDescending:
		localDirection = DirectionDescending
	default:
		localDirection = DirectionUnknown
	}

	// Ensure voltage calibration is initialized
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	// Calculate starting voltage using shared calculator
	ascending := localDirection == DirectionAscending
	calibratedVoltage := ds.getVoltageForLevelInternal(targetLevel, ascending)
	isppCalc := ds.ensureISPPCalculatorLocked()
	startVoltage := isppCalc.CalculateStartVoltage(calibratedVoltage)

	ds.isppState.Active = true
	ds.isppState.Iteration = 0
	ds.isppState.MaxIter = ISPPMaxIterations
	ds.isppState.TargetRow = row
	ds.isppState.TargetCol = col
	ds.isppState.TargetLevel = targetLevel
	ds.isppState.CurrentLevel = currentLevel
	ds.isppState.Voltage = startVoltage
	ds.isppState.Direction = localDirection
	ds.isppState.Verified = false
	ds.isppState.Complete = false
	ds.isppState.Success = false
	ds.isppState.History = []ISPPHistoryPoint{{Iteration: 0, Level: currentLevel, Voltage: startVoltage}}
}

// ISPPIterate performs one write-verify iteration
// Returns the result indicating whether to continue, success, overshoot, or max iterations
func (ds *DeviceState) ISPPIterate(newCurrentLevel int) ISPPResult {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.isppState.Active {
		return ISPPResultNotActive
	}

	ds.isppState.CurrentLevel = newCurrentLevel
	ds.isppState.Iteration++
	ds.isppState.History = append(ds.isppState.History, ISPPHistoryPoint{
		Iteration: ds.isppState.Iteration,
		Level:     ds.isppState.CurrentLevel,
		Voltage:   ds.isppState.Voltage,
	})

	// Map local direction to shared direction type
	var sharedDirection sharedphysics.HysteresisDirection
	switch ds.isppState.Direction {
	case DirectionAscending:
		sharedDirection = sharedphysics.DirectionAscending
	case DirectionDescending:
		sharedDirection = sharedphysics.DirectionDescending
	default:
		sharedDirection = sharedphysics.DirectionUnknown
	}

	// Use shared ISPP calculator to check result
	isppCalc := ds.ensureISPPCalculatorLocked()
	result := isppCalc.CheckResult(
		ds.isppState.CurrentLevel,
		ds.isppState.TargetLevel,
		sharedDirection,
		ds.isppState.Iteration,
	)

	// Map shared result to local result type and update state
	switch result {
	case sharedphysics.ISPPSuccess:
		ds.isppState.Verified = true
		ds.isppState.Complete = true
		ds.isppState.Success = true
		ds.isppState.Active = false
		return ISPPResultVerified

	case sharedphysics.ISPPOvershoot:
		return ISPPResultOvershoot

	case sharedphysics.ISPPMaxPulses:
		ds.isppState.Complete = true
		ds.isppState.Success = false
		ds.isppState.Active = false
		return ISPPResultMaxIterations

	case sharedphysics.ISPPContinue:
		// Calculate next voltage using shared calculator
		ds.isppState.Voltage = isppCalc.CalculateNextVoltage(ds.isppState.Voltage, sharedDirection)
		return ISPPResultContinue

	default:
		return ISPPResultContinue
	}
}

// HandleOvershoot performs RESET-to-saturation when write overshoots target
// Returns true if reset was performed
func (ds *DeviceState) HandleOvershoot(row, col int) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.isppState.Active {
		return false
	}

	// Reset to saturation based on direction
	if ds.isppState.Direction == DirectionAscending {
		// Ascending overshoot: reset to level 0 (negative saturation)
		ds.isppState.CurrentLevel = 0
		ds.isppState.Direction = DirectionAscending // Keep ascending for retry
	} else {
		// Descending overshoot: reset to max level (positive saturation)
		ds.isppState.CurrentLevel = ds.writeRange.NumLevels - 1
		ds.isppState.Direction = DirectionDescending // Keep descending for retry
	}

	// Ensure voltage calibration is initialized
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	// Recalculate voltage for target from new position
	ascending := ds.isppState.Direction == DirectionAscending
	calibratedVoltage := ds.getVoltageForLevelInternal(ds.isppState.TargetLevel, ascending)

	// Use shared calculator for starting voltage after reset
	if ds.isppCalc != nil {
		ds.isppState.Voltage = ds.isppCalc.CalculateStartVoltage(calibratedVoltage)
	} else {
		ds.isppState.Voltage = calibratedVoltage
	}

	ds.forceResetNextSeq = true

	return true
}

// GetISPPStatus returns the current ISPP state for UI display
func (ds *DeviceState) GetISPPStatus() ISPPState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	status := ds.isppState
	if len(ds.isppState.History) > 0 {
		status.History = append([]ISPPHistoryPoint(nil), ds.isppState.History...)
	}
	return status
}

// CancelISPP aborts the current ISPP loop
func (ds *DeviceState) CancelISPP() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.isppState.Active = false
	ds.isppState.Complete = true
	ds.isppState.Success = false
}

func (ds *DeviceState) ensureISPPCalculatorLocked() *sharedphysics.ISPPCalculator {
	if ds.isppCalc != nil {
		return ds.isppCalc
	}
	ec := 1.0
	numLevels := ds.writeRange.NumLevels
	if ds.material != nil {
		ec = ds.material.CoerciveVoltage()
		numLevels = ds.material.GetNumLevels()
	}
	if ec <= 0 {
		ec = 1.0
	}
	if numLevels < 2 {
		numLevels = sharedphysics.DefaultLevels
	}
	ds.isppCalc = sharedphysics.NewISPPCalculator(ec, numLevels)
	return ds.isppCalc
}

func (ds *DeviceState) beginISPPTracking(row, col, targetLevel, currentLevel int, direction HysteresisDirection, maxIter int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if maxIter <= 0 {
		maxIter = ISPPMaxIterations
	}

	ds.isppState.Active = true
	ds.isppState.Iteration = 0
	ds.isppState.MaxIter = maxIter
	ds.isppState.TargetRow = row
	ds.isppState.TargetCol = col
	ds.isppState.TargetLevel = targetLevel
	ds.isppState.CurrentLevel = currentLevel
	ds.isppState.Voltage = 0
	ds.isppState.Direction = direction
	ds.isppState.Verified = false
	ds.isppState.Complete = false
	ds.isppState.Success = false
	ds.isppState.History = []ISPPHistoryPoint{{Iteration: 0, Level: currentLevel, Voltage: 0}}
}

func (ds *DeviceState) updateISPPTracking(iteration int, voltage float64, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.isppState.Iteration = iteration
	ds.isppState.Voltage = voltage
	ds.isppState.CurrentLevel = currentLevel
	ds.isppState.History = append(ds.isppState.History, ISPPHistoryPoint{
		Iteration: iteration,
		Level:     currentLevel,
		Voltage:   voltage,
	})
}

func (ds *DeviceState) endISPPTracking(success bool, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.isppState.Active = false
	ds.isppState.Complete = true
	ds.isppState.Success = success
	ds.isppState.Verified = success
	ds.isppState.CurrentLevel = currentLevel
}

// ============================================================================
// COLUMN-WRITE VISUALIZATION STATE
// ============================================================================

// HalfSelectVoltageRatio is kept for backward compatibility; in DAC-only column drive
// the column disturb voltage equals the full write voltage (ratio = 1.0 effectively).
// The name is historical; do not rely on this being 0.5 (V/2 scheme is not used).
const HalfSelectVoltageRatio = 0.5

// HalfSelectVisualization holds the state for column-write overlay visualization.
// Despite the name, this implements DAC-only column drive, not a V/2 half-select scheme.
type HalfSelectVisualization struct {
	Enabled        bool
	FullVoltage    float64
	HalfVoltage    float64 // Set equal to FullVoltage in DAC-only mode (no V/2 splitting)
	SelectedRow    int
	SelectedCol    int
	HalfSelectRows []int // Rows disturbed at full voltage (same column -- all switch)
	HalfSelectCols []int // Always empty in DAC-only mode (same-row cells see 0V)
}

// EnableHalfSelectVisualization enables the column-write overlay for a passive write operation.
// In DAC-Only Column Drive, all rows in the selected column see full write voltage (disturbed).
// Cells in the same row see 0V (safe -- row is grounded, unselected BL is 0V).
func (ds *DeviceState) EnableHalfSelectVisualization(row, col int, fullVoltage float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.halfSelectState.Enabled = true
	ds.halfSelectState.FullVoltage = fullVoltage
	// In DAC-Only drive the whole column is disturbed at full voltage; no V/2 partial stress.
	ds.halfSelectState.HalfVoltage = fullVoltage

	ds.halfSelectState.SelectedRow = row
	ds.halfSelectState.SelectedCol = col

	// All other rows in the same column are disturbed (WL=0, BL=-V -> DeltaV=+V, full disturb)
	ds.halfSelectState.HalfSelectRows = make([]int, 0)
	for r := 0; r < ds.rows; r++ {
		if r != row {
			ds.halfSelectState.HalfSelectRows = append(ds.halfSelectState.HalfSelectRows, r)
		}
	}

	// Same-row cells see 0V (WL=0, BL=0) -- no disturb, so HalfSelectCols is empty.
	ds.halfSelectState.HalfSelectCols = make([]int, 0)
}

// DisableHalfSelectVisualization disables the V/2 overlay
func (ds *DeviceState) DisableHalfSelectVisualization() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.halfSelectState.Enabled = false
	ds.halfSelectState.HalfSelectRows = nil
	ds.halfSelectState.HalfSelectCols = nil
}

// GetHalfSelectState returns the current V/2 visualization state
func (ds *DeviceState) GetHalfSelectState() HalfSelectVisualization {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.halfSelectState
}

// IsHalfSelected returns true if the given cell is in half-select state
func (ds *DeviceState) IsHalfSelected(row, col int) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if !ds.halfSelectState.Enabled {
		return false
	}

	// Check if in half-select row (same column as selected)
	if col == ds.halfSelectState.SelectedCol {
		for _, r := range ds.halfSelectState.HalfSelectRows {
			if r == row {
				return true
			}
		}
	}

	// Check if in half-select column (same row as selected)
	if row == ds.halfSelectState.SelectedRow {
		for _, c := range ds.halfSelectState.HalfSelectCols {
			if c == col {
				return true
			}
		}
	}

	return false
}
