//go:build legacy_fyne

// Package gui — device_state_write.go
// Write-path logic for DeviceState: DAC-only column drive (passive 0T1R),
// voltage reset, per-level voltage calibration, hysteresis direction tracking,
// and the 5-phase program-verify sequence state machine.
// ISPP loop and half-select visualization live in device_state_ispp.go.
// programLevelFromCoupledVoltage lives in device_state_compute.go.
package gui

import (
	"fmt"
	"math"
)

// ============================================================================
// DAC-ONLY COLUMN DRIVE (Passive 0T1R Mode Only)
// ============================================================================
//
// For passive (0T1R) WRITE operations the DAC drives the selected BL while all
// WLs are grounded through the TIA virtual ground. There is no V/2 splitting.
//
// Target cell (SET operation):
//   - All WLs: 0V (grounded / TIA virtual ground)
//   - Selected BL (DAC): -V_write (full write voltage)
//   - Effective DeltaV = WL - BL = 0 - (-V_write) = +V_write (full switching)
//
// Column disturb (same column, different row):
//   - WL = 0, BL = -V_write -> DeltaV = +V_write (FULL write -- entire column switches)
//
// Same-row cells (different column):
//   - WL = 0, BL = 0 -> DeltaV = 0 (safe -- unselected BLs grounded)
//
// Unselected cells (different row AND column):
//   - WL = 0, BL = 0 -> DeltaV = 0 (no disturb)
//
// For 1T1R/2T1R modes, transistor gate on selected row completes the circuit;
// only the selected cell [row,col] can switch (transistor isolation).
// ============================================================================

// ApplyHalfSelectWrite applies voltage biasing for passive (0T1R) write operation.
// Implements DAC-Only Column Drive: since rows are grounded (TIA virtual ground),
// the full write voltage is applied to the selected column.
//
// Target cell (SET operation):
//   - All WLs: 0V (Grounded / TIA Virtual Ground)
//   - Selected BL (DAC): -V_write (SET: positive DeltaV across cell)
//   - Effective DeltaV = WL - BL = 0 - (-V_write) = +V_write (full switching)
//
// Consequence: this performs a COLUMN WRITE -- all cells in the selected column
// see the full V_write. Unselected columns see 0V (no disturb).
//
// For 1T1R/2T1R modes, transistor isolation eliminates need for column drive.
func (ds *DeviceState) ApplyHalfSelectWrite(targetRow, targetCol int, writeVoltage float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.isPassive {
		// Non-passive modes use transistor isolation
		// Apply full voltage to selected BL only, WL controls transistor gate
		ds.setDACVoltageLocked(targetCol, writeVoltage)
		return
	}

	// DAC-Only Drive Scheme: rows grounded, selected column driven to -V_write
	for i := range ds.wlVoltages {
		ds.wlVoltages[i] = 0
	}

	dacV := -writeVoltage
	for i := range ds.dacVoltages {
		if i == targetCol {
			ds.dacVoltages[i] = dacV
		} else {
			ds.dacVoltages[i] = 0 // Unselected BLs grounded
		}
	}

	ds.dacMode = DACManual
	ds.dacRangeMode = DACRangeWrite
}

// ResetWriteVoltages returns all WL and BL voltages to 0V after write operation
// Should be called after write completes to put array in safe idle state
func (ds *DeviceState) ResetWriteVoltages() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.resetWriteVoltagesLocked()
}

func (ds *DeviceState) resetWriteVoltagesLocked() {
	// Reset all WL voltages to 0
	for i := range ds.wlVoltages {
		ds.wlVoltages[i] = 0
	}
	// Reset all DAC (BL) voltages to 0
	for i := range ds.dacVoltages {
		ds.dacVoltages[i] = 0
	}
	ds.dacMode = DACManual
}

// GetWLVoltage returns the WL voltage for a specific row
func (ds *DeviceState) GetWLVoltage(row int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row >= 0 && row < len(ds.wlVoltages) {
		return ds.wlVoltages[row]
	}
	return 0
}

// GetHalfSelectVoltage returns V/2 value derived from material's write voltage
// This is the voltage seen by half-selected cells (below Vc, minimal disturb)
func (ds *DeviceState) GetHalfSelectVoltage() float64 {
	// Use middle of write range as reference
	fullWriteV := (ds.writeRange.Min + ds.writeRange.Max) / 2
	return fullWriteV / 2.0
}

// IsUsingHalfSelect returns true if V/2 scheme is active (passive mode write)
func (ds *DeviceState) IsUsingHalfSelect() bool {
	return ds.isPassive && ds.opMode == OpModeWrite
}

// ============================================================================
// PER-LEVEL VOLTAGE CALIBRATION
// ============================================================================

// PerLevelVoltageCalibration holds calibrated voltages for each level
// Linear interpolation used as simplified demo (not physics-accurate Preisach model)
type PerLevelVoltageCalibration struct {
	AscendingVoltages  []float64 // Voltages for writing up (level 0->max)
	DescendingVoltages []float64 // Voltages for writing down (level max->0)
}

// initVoltageCalibrationInternal initializes the per-level voltage arrays (internal, no locking)
// Caller must hold appropriate lock
func (ds *DeviceState) initVoltageCalibrationInternal() {
	numLevels := ds.writeRange.NumLevels
	cal := &PerLevelVoltageCalibration{
		AscendingVoltages:  make([]float64, numLevels),
		DescendingVoltages: make([]float64, numLevels),
	}

	// Linear interpolation: Level 0 -> WriteRange.Min, Level (numLevels-1) -> WriteRange.Max
	minV := ds.writeRange.Min
	maxV := ds.writeRange.Max
	step := (maxV - minV) / float64(numLevels-1)

	for i := 0; i < numLevels; i++ {
		voltage := minV + float64(i)*step
		cal.AscendingVoltages[i] = voltage
		cal.DescendingVoltages[i] = voltage // Same for simplified model
	}

	ds.voltageCalibration = cal
}

// InitVoltageCalibration initializes the per-level voltage arrays using linear interpolation
// This is a simplified demo - real devices would use non-linear Preisach-derived values
func (ds *DeviceState) InitVoltageCalibration() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.initVoltageCalibrationInternal()
}

// getVoltageForLevelInternal returns the calibrated write voltage (internal, no locking)
// Caller must hold appropriate lock
func (ds *DeviceState) getVoltageForLevelInternal(level int, ascending bool) float64 {
	// Clamp level to valid range
	maxLevel := len(ds.voltageCalibration.AscendingVoltages) - 1
	if level < 0 {
		level = 0
	}
	if level > maxLevel {
		level = maxLevel
	}

	if ascending {
		return ds.voltageCalibration.AscendingVoltages[level]
	}
	return ds.voltageCalibration.DescendingVoltages[level]
}

// GetVoltageForLevel returns the calibrated write voltage for a target level
// direction: true = ascending (increasing level), false = descending (decreasing level)
func (ds *DeviceState) GetVoltageForLevel(level int, ascending bool) float64 {
	ds.mu.Lock()
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}
	result := ds.getVoltageForLevelInternal(level, ascending)
	ds.mu.Unlock()
	return result
}

// GetLevelForVoltage estimates the nearest discrete level for a given write voltage.
// Uses the calibrated per-level voltage table (ascending/descending) to avoid ad-hoc mapping.
func (ds *DeviceState) GetLevelForVoltage(voltage float64, ascending bool) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	levels := ds.voltageCalibration.AscendingVoltages
	if !ascending {
		levels = ds.voltageCalibration.DescendingVoltages
	}
	if len(levels) == 0 {
		return 0
	}

	closest := 0
	minDiff := math.Abs(levels[0] - voltage)
	for i, v := range levels {
		diff := math.Abs(v - voltage)
		if diff < minDiff {
			minDiff = diff
			closest = i
		}
	}

	return closest
}

// ============================================================================
// HYSTERESIS DIRECTION TRACKING
// ============================================================================

// HysteresisDirection indicates the write direction on the hysteresis curve
type HysteresisDirection int

const (
	DirectionUnknown    HysteresisDirection = iota
	DirectionAscending                      // Writing to higher level
	DirectionDescending                     // Writing to lower level
)

// HysteresisState tracks the last written level and direction per cell
type HysteresisState struct {
	LastLevel map[string]int                 // key: "row,col" -> last written level
	Direction map[string]HysteresisDirection // key: "row,col" -> last direction
}

// cellKey generates a map key for a cell coordinate
func cellKey(row, col int) string {
	return fmt.Sprintf("%d,%d", row, col)
}

// RecordWrite updates the hysteresis state after a successful write
func (ds *DeviceState) RecordWrite(row, col, newLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	key := cellKey(row, col)

	oldLevel, exists := ds.hysteresisState.LastLevel[key]
	if exists {
		if newLevel > oldLevel {
			ds.hysteresisState.Direction[key] = DirectionAscending
		} else if newLevel < oldLevel {
			ds.hysteresisState.Direction[key] = DirectionDescending
		}
		// If equal, keep previous direction
	} else {
		ds.hysteresisState.Direction[key] = DirectionUnknown
	}

	ds.hysteresisState.LastLevel[key] = newLevel
}

// GetWriteDirection determines the write direction for a target level
func (ds *DeviceState) GetWriteDirection(row, col, currentLevel, targetLevel int) HysteresisDirection {
	if targetLevel > currentLevel {
		return DirectionAscending
	} else if targetLevel < currentLevel {
		return DirectionDescending
	}
	return DirectionUnknown
}

// GetLastHysteresisDirection returns the last write direction for a cell
func (ds *DeviceState) GetLastHysteresisDirection(row, col int) HysteresisDirection {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	key := cellKey(row, col)
	if dir, exists := ds.hysteresisState.Direction[key]; exists {
		return dir
	}
	return DirectionUnknown
}

// ============================================================================
// 5-PHASE PROGRAM-VERIFY SEQUENCE STATE MACHINE
// ============================================================================

// WritePhase represents the current phase in a program-verify sequence
type WritePhase int

const (
	PhaseIdle   WritePhase = iota // No write in progress
	PhaseReset                    // Applying -V_sat (100ns)
	PhaseHold1                    // Zero field hold (50ns)
	PhaseWrite                    // Applying calibrated voltage (200ns)
	PhaseHold2                    // Zero field hold (50ns)
	PhaseVerify                   // Read/verify at low voltage (80ns)
)

// Phase timing constants (in nanoseconds for display, not real-time)
const (
	PhaseResetDurationNs  = 100
	PhaseHold1DurationNs  = 50
	PhaseWriteDurationNs  = 200
	PhaseHold2DurationNs  = 50
	PhaseVerifyDurationNs = 80
)

// WriteSequenceState holds the state of an active program-verify sequence
type WriteSequenceState struct {
	Active       bool
	Phase        WritePhase
	TargetRow    int
	TargetCol    int
	TargetLevel  int
	CurrentLevel int
	WriteVoltage float64 // Calibrated voltage for target level
	PhaseVoltage float64 // Actual applied voltage for current phase
	Progress     float64 // 0.0 to 1.0 progress through sequence
}

// StartWriteSequence begins a program-verify sequence
func (ds *DeviceState) StartWriteSequence(row, col, targetLevel, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	direction := ds.GetWriteDirection(row, col, currentLevel, targetLevel)
	// Initialize voltage calibration if needed
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	ascending := direction == DirectionAscending

	// Ensure voltage calibration is initialized
	ds.initVoltageCalibrationInternal()

	// Skip RESET when staying on the same branch, or after the first ISPP pulse.
	// Only force RESET on direction change (first pulse) or explicit overshoot flag.
	lastDir := DirectionUnknown
	if dir, exists := ds.hysteresisState.Direction[cellKey(row, col)]; exists {
		lastDir = dir
	}
	sameBranch := lastDir == DirectionUnknown || lastDir == direction
	startPhase := PhaseReset
	if !ds.forceResetNextSeq {
		if ds.isppState.Iteration > 0 || sameBranch {
			startPhase = PhaseHold1
		}
	}
	ds.forceResetNextSeq = false

	ds.writeSequenceState.Active = true
	ds.writeSequenceState.Phase = startPhase
	ds.writeSequenceState.TargetRow = row
	ds.writeSequenceState.TargetCol = col
	ds.writeSequenceState.TargetLevel = targetLevel
	ds.writeSequenceState.CurrentLevel = currentLevel
	calibrated := ds.getVoltageForLevelInternal(targetLevel, ascending)
	writeVoltage := calibrated
	if ds.isppState.Active && ds.isppState.Voltage > 0 {
		// Use the current ISPP pulse voltage when available.
		writeVoltage = ds.isppState.Voltage
	}
	applied, _ := ds.dacWriteVoltageLocked(writeVoltage)
	ds.writeSequenceState.WriteVoltage = applied
	if startPhase == PhaseHold1 {
		ds.writeSequenceState.Progress = 0.2
	} else {
		ds.writeSequenceState.Progress = 0.0
	}
	ds.updateWriteSequencePhaseVoltageLocked()
}

// AdvanceWritePhase moves to the next phase in the sequence
// Returns true if sequence is complete
func (ds *DeviceState) AdvanceWritePhase() bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.writeSequenceState.Active {
		return true
	}

	switch ds.writeSequenceState.Phase {
	case PhaseReset:
		ds.writeSequenceState.Phase = PhaseHold1
		ds.writeSequenceState.Progress = 0.2
	case PhaseHold1:
		ds.writeSequenceState.Phase = PhaseWrite
		ds.writeSequenceState.Progress = 0.4
	case PhaseWrite:
		ds.writeSequenceState.Phase = PhaseHold2
		ds.writeSequenceState.Progress = 0.6
	case PhaseHold2:
		ds.writeSequenceState.Phase = PhaseVerify
		ds.writeSequenceState.Progress = 0.8
	case PhaseVerify:
		ds.writeSequenceState.Phase = PhaseIdle
		ds.writeSequenceState.Active = false
		ds.writeSequenceState.Progress = 1.0
		ds.writeSequenceState.PhaseVoltage = 0.0
		return true
	}
	ds.updateWriteSequencePhaseVoltageLocked()
	return false
}

func (ds *DeviceState) updateWriteSequencePhaseVoltageLocked() {
	switch ds.writeSequenceState.Phase {
	case PhaseWrite:
		ds.writeSequenceState.PhaseVoltage = ds.writeSequenceState.WriteVoltage
	case PhaseVerify:
		// Use a safe read voltage for verify (below coercive voltage).
		verifyVoltage := ds.readRange.Max * 0.5
		if verifyVoltage < 0 {
			verifyVoltage = 0
		}
		ds.writeSequenceState.PhaseVoltage = verifyVoltage
	default:
		ds.writeSequenceState.PhaseVoltage = 0.0
	}
}

// GetWritePhaseInfo returns the current write sequence state for UI display
func (ds *DeviceState) GetWritePhaseInfo() WriteSequenceState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.writeSequenceState
}

// CancelWriteSequence aborts the current write sequence
func (ds *DeviceState) CancelWriteSequence() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.writeSequenceState.Active = false
	ds.writeSequenceState.Phase = PhaseIdle
	ds.writeSequenceState.Progress = 0.0
	ds.writeSequenceState.PhaseVoltage = 0.0
}

// GetPhaseName returns a human-readable name for a write phase
func GetPhaseName(phase WritePhase) string {
	switch phase {
	case PhaseIdle:
		return "IDLE"
	case PhaseReset:
		return "RESET"
	case PhaseHold1:
		return "HOLD"
	case PhaseWrite:
		return "WRITE"
	case PhaseHold2:
		return "HOLD"
	case PhaseVerify:
		return "VERIFY"
	default:
		return "UNKNOWN"
	}
}

// GetPhaseDuration returns the duration in nanoseconds for a phase
func GetPhaseDuration(phase WritePhase) int {
	switch phase {
	case PhaseReset:
		return PhaseResetDurationNs
	case PhaseHold1:
		return PhaseHold1DurationNs
	case PhaseWrite:
		return PhaseWriteDurationNs
	case PhaseHold2:
		return PhaseHold2DurationNs
	case PhaseVerify:
		return PhaseVerifyDurationNs
	default:
		return 0
	}
}
