// Package physics provides shared physics utilities for FeCIM simulations.
// This file implements the ISPP (Incremental Step Pulse Programming) calculator
// for precise multi-level programming of ferroelectric memory cells.
//
// ISPP Physics:
//
// Ferroelectric memory requires careful write strategies due to hysteresis:
//   - The polarization state depends on the history of applied fields
//   - A single pulse may undershoot or overshoot the target state
//   - ISPP applies incremental pulses, verifying after each one
//
// Programming flow:
//   1. Apply initial pulse at ~70% of calibrated voltage (conservative start)
//   2. Read back current state and compare to target
//   3. If undershoot: increase voltage by step and pulse again
//   4. If overshoot: must RESET to saturation and restart (hysteresis path-dependent)
//   5. Repeat until target reached or max pulses exceeded
//
// Why overshoot requires reset:
//   Ferroelectric polarization follows a hysteresis loop. If you're going UP
//   (increasing polarization) and overshoot, you can't simply reduce the field
//   to go back down - you'd follow a different branch of the hysteresis loop.
//   Instead, you must saturate to the opposite extreme and approach again.
//
// Typical performance:
//   - Well-calibrated device: 1-3 pulses to converge
//   - Poor calibration: 5-10 pulses with frequent overshoots
//   - Failure modes: stuck bits, excessive overshoots, max iterations
//
// References:
//   - IEEE IEDM 2022: FeFET multi-level programming statistics
//   - Nature Electronics 2023: Analog state precision in FeFETs
//   - IEEE IRPS 2022: Write-verify endurance characteristics
package physics

import (
	"math"
)

// HysteresisDirection indicates the programming direction on the hysteresis loop.
type HysteresisDirection int

const (
	// DirectionUnknown indicates the direction has not been determined.
	DirectionUnknown HysteresisDirection = iota

	// DirectionAscending means going to higher levels (applying positive field).
	// On the P-E loop: moving from -Pr toward +Pr.
	DirectionAscending

	// DirectionDescending means going to lower levels (applying negative field).
	// On the P-E loop: moving from +Pr toward -Pr.
	DirectionDescending
)

// String returns a human-readable representation of the direction.
func (d HysteresisDirection) String() string {
	switch d {
	case DirectionAscending:
		return "Ascending"
	case DirectionDescending:
		return "Descending"
	default:
		return "Unknown"
	}
}

// ISPPResult represents the outcome of an ISPP verification step.
type ISPPResult int

const (
	// ISPPContinue indicates more pulses are needed to reach target.
	ISPPContinue ISPPResult = iota

	// ISPPSuccess indicates the target level was reached within tolerance.
	ISPPSuccess

	// ISPPOvershoot indicates the current level went past the target.
	// Recovery requires reset to saturation and restart.
	ISPPOvershoot

	// ISPPMaxPulses indicates the maximum iteration count was reached without success.
	ISPPMaxPulses
)

// String returns a human-readable representation of the result.
func (r ISPPResult) String() string {
	switch r {
	case ISPPContinue:
		return "Continue"
	case ISPPSuccess:
		return "Success"
	case ISPPOvershoot:
		return "Overshoot"
	case ISPPMaxPulses:
		return "MaxPulses"
	default:
		return "Unknown"
	}
}

// ISPPConfig contains configuration parameters for ISPP programming.
type ISPPConfig struct {
	// StartRatio is the fraction of calibrated voltage to use for the first pulse.
	// Default: 0.7 (start at 70% of calibrated voltage for conservative approach).
	// Lower values reduce overshoot risk but increase pulse count.
	StartRatio float64

	// StepPercent is the voltage increment per pulse, as a fraction of Ec.
	// Default: 0.05 (5% of coercive voltage per step).
	// Smaller steps give finer control but require more pulses.
	StepPercent float64

	// MaxPulses is the maximum number of pulses before declaring failure.
	// Default: 10 (typical ISPP implementations use 5-15).
	MaxPulses int

	// SafetyCap is the maximum voltage as a multiple of Ec.
	// Default: 2.2 (prevents damage from excessive field).
	// Higher values risk device degradation, lower values may prevent reaching extreme levels.
	SafetyCap float64

	// Tolerance is the acceptable error in levels.
	// Default: 0 (exact match required).
	// Setting to 1 allows +-1 level error, reducing pulse count.
	Tolerance int
}

// DefaultISPPConfig returns the default ISPP configuration.
func DefaultISPPConfig() ISPPConfig {
	return ISPPConfig{
		StartRatio:  0.7,
		StepPercent: 0.02, // 2% of Ec per step for fine control (was 5%)
		MaxPulses:   20,   // More pulses for fine convergence (was 10)
		SafetyCap:   2.2,
		Tolerance:   0,
	}
}

// ISPPCalculator performs ISPP voltage calculations for ferroelectric programming.
type ISPPCalculator struct {
	// Config holds the ISPP configuration parameters.
	Config ISPPConfig

	// Ec is the coercive voltage (V) for this material/device.
	// This is the voltage at which polarization switches, typically:
	//   Ec = coercive_field * film_thickness
	// For 10nm HZO with Ec~1 MV/cm: Ec_voltage ~ 1.0V
	Ec float64

	// NumLevels is the total number of discrete analog states.
	// For FeCIM: 30 levels (4.91 bits/cell).
	NumLevels int
}

// NewISPPCalculator creates a new calculator with default configuration.
//
// Parameters:
//   - ec: Coercive voltage in Volts (material/device dependent)
//   - numLevels: Number of discrete analog states (e.g., 30 for FeCIM)
func NewISPPCalculator(ec float64, numLevels int) *ISPPCalculator {
	return &ISPPCalculator{
		Config:    DefaultISPPConfig(),
		Ec:        ec,
		NumLevels: numLevels,
	}
}

// NewISPPCalculatorWithConfig creates a calculator with custom configuration.
//
// Parameters:
//   - ec: Coercive voltage in Volts
//   - numLevels: Number of discrete analog states
//   - config: Custom ISPP configuration
func NewISPPCalculatorWithConfig(ec float64, numLevels int, config ISPPConfig) *ISPPCalculator {
	return &ISPPCalculator{
		Config:    config,
		Ec:        ec,
		NumLevels: numLevels,
	}
}

// GetDirection determines the programming direction based on current and target levels.
//
// Returns:
//   - DirectionAscending if target > current (need positive field)
//   - DirectionDescending if target < current (need negative field)
//   - DirectionUnknown if target == current (no programming needed)
func GetDirection(currentLevel, targetLevel int) HysteresisDirection {
	if targetLevel > currentLevel {
		return DirectionAscending
	}
	if targetLevel < currentLevel {
		return DirectionDescending
	}
	return DirectionUnknown
}

// CalculateStartVoltage computes the initial pulse voltage.
// Uses a conservative fraction of the calibrated voltage to reduce overshoot risk.
//
// Parameters:
//   - calibratedVoltage: The pre-calibrated voltage for the target level
//
// Returns: The starting voltage (calibratedVoltage * StartRatio)
func (c *ISPPCalculator) CalculateStartVoltage(calibratedVoltage float64) float64 {
	return calibratedVoltage * c.Config.StartRatio
}

// CalculateVoltageStep returns the voltage increment per ISPP pulse.
// Based on a fraction of the coercive voltage.
//
// Returns: Ec * StepPercent
func (c *ISPPCalculator) CalculateVoltageStep() float64 {
	return c.Ec * c.Config.StepPercent
}

// CalculateNextVoltage computes the next pulse voltage after an undershoot.
//
// Parameters:
//   - currentVoltage: The voltage used in the previous pulse
//   - direction: The programming direction (ascending or descending)
//
// Returns: The adjusted voltage, clamped to safety limits
func (c *ISPPCalculator) CalculateNextVoltage(currentVoltage float64, direction HysteresisDirection) float64 {
	step := c.CalculateVoltageStep()

	var nextVoltage float64
	switch direction {
	case DirectionAscending:
		// Ascending: increase voltage (more positive)
		nextVoltage = currentVoltage + step
	case DirectionDescending:
		// Descending: decrease voltage (more negative)
		nextVoltage = currentVoltage - step
	default:
		// No change for unknown direction
		nextVoltage = currentVoltage
	}

	return c.ClampVoltage(nextVoltage, direction)
}

// ClampVoltage enforces safety limits on the programming voltage.
//
// Limits:
//   - Ascending: 0 to +Ec*SafetyCap (positive voltages only)
//   - Descending: -Ec*SafetyCap to 0 (negative voltages only)
//
// Parameters:
//   - voltage: The voltage to clamp
//   - direction: The programming direction
//
// Returns: The clamped voltage within safe bounds
func (c *ISPPCalculator) ClampVoltage(voltage float64, direction HysteresisDirection) float64 {
	maxMagnitude := c.Ec * c.Config.SafetyCap

	switch direction {
	case DirectionAscending:
		// Positive field: clamp to [0, +maxMagnitude]
		if voltage < 0 {
			return 0
		}
		if voltage > maxMagnitude {
			return maxMagnitude
		}
	case DirectionDescending:
		// Negative field: clamp to [-maxMagnitude, 0]
		if voltage > 0 {
			return 0
		}
		if voltage < -maxMagnitude {
			return -maxMagnitude
		}
	}

	return voltage
}

// CheckResult evaluates the current ISPP state and determines the next action.
//
// Parameters:
//   - currentLevel: The level read back after the last pulse
//   - targetLevel: The desired target level
//   - direction: The programming direction
//   - pulseCount: Number of pulses applied so far
//
// Returns:
//   - ISPPSuccess: if |currentLevel - targetLevel| <= Tolerance
//   - ISPPOvershoot: if went past target (requires reset)
//   - ISPPMaxPulses: if pulseCount >= MaxPulses
//   - ISPPContinue: otherwise (apply another pulse)
func (c *ISPPCalculator) CheckResult(currentLevel, targetLevel int, direction HysteresisDirection, pulseCount int) ISPPResult {
	// Check for success (within tolerance)
	error := currentLevel - targetLevel
	if int(math.Abs(float64(error))) <= c.Config.Tolerance {
		return ISPPSuccess
	}

	// Check for overshoot
	if c.IsOvershoot(currentLevel, targetLevel, direction) {
		return ISPPOvershoot
	}

	// Check for max pulses reached
	if pulseCount >= c.Config.MaxPulses {
		return ISPPMaxPulses
	}

	// Need more pulses
	return ISPPContinue
}

// IsOvershoot checks if the current level has gone past the target.
//
// Overshoot conditions:
//   - Ascending (going to higher levels): current > target
//   - Descending (going to lower levels): current < target
//
// Parameters:
//   - currentLevel: The level read back after the last pulse
//   - targetLevel: The desired target level
//   - direction: The programming direction
//
// Returns: true if overshoot occurred
func (c *ISPPCalculator) IsOvershoot(currentLevel, targetLevel int, direction HysteresisDirection) bool {
	switch direction {
	case DirectionAscending:
		// Going up: overshoot if we went too high
		return currentLevel > targetLevel
	case DirectionDescending:
		// Going down: overshoot if we went too low
		return currentLevel < targetLevel
	default:
		return false
	}
}

// GetResetVoltage returns the voltage needed to reset to saturation after overshoot.
//
// After overshoot, the device must be reset to the opposite saturation state
// before retrying. This is due to the path-dependent nature of hysteresis.
//
// Reset voltages:
//   - Ascending overshoot: reset to -Ec*SafetyCap (negative saturation)
//   - Descending overshoot: reset to +Ec*SafetyCap (positive saturation)
//
// Parameters:
//   - direction: The direction that caused the overshoot
//
// Returns: The reset voltage to apply
func (c *ISPPCalculator) GetResetVoltage(direction HysteresisDirection) float64 {
	maxMagnitude := c.Ec * c.Config.SafetyCap

	switch direction {
	case DirectionAscending:
		// Was going up and overshot: reset to negative saturation
		return -maxMagnitude
	case DirectionDescending:
		// Was going down and overshot: reset to positive saturation
		return maxMagnitude
	default:
		return 0
	}
}

// LevelError calculates the signed error between current and target levels.
// Positive = current is higher than target, Negative = current is lower.
func LevelError(currentLevel, targetLevel int) int {
	return currentLevel - targetLevel
}

// GetSaturationVoltage returns the voltage for full saturation in a direction.
//
// Parameters:
//   - direction: DirectionAscending for +Pr, DirectionDescending for -Pr
//
// Returns: +Ec*SafetyCap for ascending, -Ec*SafetyCap for descending
func (c *ISPPCalculator) GetSaturationVoltage(direction HysteresisDirection) float64 {
	maxMagnitude := c.Ec * c.Config.SafetyCap

	switch direction {
	case DirectionAscending:
		return maxMagnitude
	case DirectionDescending:
		return -maxMagnitude
	default:
		return 0
	}
}

// EstimatePulsesNeeded estimates the number of pulses to reach target.
// This is a rough estimate based on the voltage step and level spacing.
//
// Parameters:
//   - currentLevel: Current device state
//   - targetLevel: Desired target state
//   - calibratedVoltage: Pre-calibrated voltage for target level
//
// Returns: Estimated number of pulses (1 to MaxPulses)
func (c *ISPPCalculator) EstimatePulsesNeeded(currentLevel, targetLevel int, calibratedVoltage float64) int {
	if currentLevel == targetLevel {
		return 1 // Already there, just verify
	}

	// Estimate based on conservative start and step size
	startV := c.CalculateStartVoltage(calibratedVoltage)
	stepV := c.CalculateVoltageStep()

	// Voltage gap to close from start to target
	voltageGap := math.Abs(calibratedVoltage - startV)

	// Estimated pulses = gap / step, minimum 1
	if stepV > 0 {
		estimated := int(math.Ceil(voltageGap/stepV)) + 1
		if estimated > c.Config.MaxPulses {
			return c.Config.MaxPulses
		}
		if estimated < 1 {
			return 1
		}
		return estimated
	}

	return 1
}
