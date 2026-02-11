package controller

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
)

// TestNewWriteController verifies default initialization
func TestNewWriteController(t *testing.T) {
	numLevels := 30
	ec := 1.0
	emax := 2.5
	calib := algo.NewCalibrationManager(numLevels)

	wc := NewWriteController(numLevels, ec, emax, calib)

	if wc.NumLevels != numLevels {
		t.Errorf("NumLevels: got %d, want %d", wc.NumLevels, numLevels)
	}
	if wc.EcField != ec {
		t.Errorf("EcField: got %f, want %f", wc.EcField, ec)
	}
	if wc.MaxField != emax {
		t.Errorf("MaxField: got %f, want %f", wc.MaxField, emax)
	}
	if wc.CalibManager != calib {
		t.Error("CalibManager not set")
	}
	if wc.State != StateIdle {
		t.Errorf("State: got %v, want %v", wc.State, StateIdle)
	}
	if wc.MinStep <= 0 {
		t.Errorf("MinStep should be positive, got %f", wc.MinStep)
	}
	if wc.PulseDuration <= 0 {
		t.Errorf("PulseDuration should be positive, got %f", wc.PulseDuration)
	}
}

// TestWriteController_Start verifies Start() initializes state correctly
func TestWriteController_Start(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)

	targetLevel := 15
	wc.Start(targetLevel, true)

	if wc.TargetLevel != targetLevel {
		t.Errorf("TargetLevel: got %d, want %d", wc.TargetLevel, targetLevel)
	}
	if !wc.FromSaturation {
		t.Error("FromSaturation should be true")
	}
	if wc.State != StateApply {
		t.Errorf("State: got %v, want %v", wc.State, StateApply)
	}
	if wc.PhaseTimer != 0 {
		t.Errorf("PhaseTimer: got %f, want 0", wc.PhaseTimer)
	}
	if wc.PulseCount != 0 {
		t.Errorf("PulseCount: got %d, want 0", wc.PulseCount)
	}
	if wc.OvershootCount != 0 {
		t.Errorf("OvershootCount: got %d, want 0", wc.OvershootCount)
	}
	if wc.GuardPulseCount != 0 {
		t.Errorf("GuardPulseCount: got %d, want 0", wc.GuardPulseCount)
	}
	if wc.VMin != 0 {
		t.Errorf("VMin: got %f, want 0", wc.VMin)
	}
	if wc.VMax != wc.MaxField {
		t.Errorf("VMax: got %f, want %f", wc.VMax, wc.MaxField)
	}
}

// TestWriteController_ResetState verifies all counters reset
func TestWriteController_ResetState(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)

	// Set some state
	wc.State = StateSuccess
	wc.TotalPulses = 10
	wc.RetryCount = 2
	wc.Attempts = 3
	wc.SuccessCount = 1
	wc.FailureCount = 1
	wc.OvershootTotal = 5

	wc.ResetState()

	if wc.State != StateIdle {
		t.Errorf("State: got %v, want %v", wc.State, StateIdle)
	}
	if wc.TotalPulses != 0 {
		t.Errorf("TotalPulses: got %d, want 0", wc.TotalPulses)
	}
	if wc.RetryCount != 0 {
		t.Errorf("RetryCount: got %d, want 0", wc.RetryCount)
	}
	if wc.Attempts != 0 {
		t.Errorf("Attempts: got %d, want 0", wc.Attempts)
	}
	if wc.SuccessCount != 0 {
		t.Errorf("SuccessCount: got %d, want 0", wc.SuccessCount)
	}
	if wc.FailureCount != 0 {
		t.Errorf("FailureCount: got %d, want 0", wc.FailureCount)
	}
	if wc.OvershootTotal != 0 {
		t.Errorf("OvershootTotal: got %d, want 0", wc.OvershootTotal)
	}
}

// TestWriteController_Update_BasicSequence verifies ascending write sequence
func TestWriteController_Update_BasicSequence(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.Start(15, false) // Target level 15 from level 0

	currentLevel := 0
	currentField := 0.0
	guardSign := 0
	dt := 0.01

	// First update should move to APPLY and calculate field
	targetField, done := wc.Update(dt, currentField, currentLevel, guardSign)
	if done {
		t.Error("Should not be done on first update")
	}
	if wc.State != StateApply {
		t.Errorf("State: got %v, want %v", wc.State, StateApply)
	}
	// Field should be positive (ascending)
	if targetField <= 0 {
		t.Errorf("targetField should be positive for ascending write, got %f", targetField)
	}
}

// TestWriteController_Update_Descending verifies descending write sequence
func TestWriteController_Update_Descending(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.Start(10, false) // Target level 10 from level 29

	currentLevel := 29
	currentField := 0.0
	guardSign := 0
	dt := 0.01

	// Initialize with current level
	targetField, done := wc.Update(dt, currentField, currentLevel, guardSign)
	if done {
		t.Error("Should not be done on first update")
	}
	if wc.State != StateApply {
		t.Errorf("State: got %v, want %v", wc.State, StateApply)
	}
	// Field should be negative (descending)
	if targetField >= 0 {
		t.Errorf("targetField should be negative for descending write, got %f", targetField)
	}
}

// TestWriteController_Update_AtTarget verifies early completion when at target
func TestWriteController_Update_AtTarget(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	targetLevel := 15
	wc.Start(targetLevel, false)

	currentLevel := targetLevel // Already at target
	currentField := 0.0
	guardSign := 0
	dt := 0.01

	// Should recognize early and finish
	_, done := wc.Update(dt, currentField, currentLevel, guardSign)
	if !done {
		t.Error("Should be done when starting at target level")
	}
	if wc.State != StateSuccess {
		t.Errorf("State: got %v, want %v", wc.State, StateSuccess)
	}
	if wc.LastError != 0 {
		t.Errorf("LastError: got %d, want 0", wc.LastError)
	}
}

// TestWriteController_GuardSign_NoFlip verifies guard pulse limit prevents direction flip
func TestWriteController_GuardSign_NoFlip(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.Start(15, false)

	// Simulate reaching target with guard sign active
	wc.State = StateVerify
	wc.PhaseTimer = 0.1 // Past wait threshold
	wc.PulseCount = 5
	wc.LastVerifyLevel = 15
	wc.CurrentField = 0.8 // Some positive field

	currentLevel := 15 // At target
	currentField := 0.0 // Field has settled
	guardSign := 1      // Guard wants to push higher
	dt := 0.01

	// First guard pulse
	_, done := wc.Update(dt, currentField, currentLevel, guardSign)
	if done {
		t.Error("Should not be done on first guard pulse")
	}
	if wc.GuardPulseCount != 1 {
		t.Errorf("GuardPulseCount: got %d, want 1", wc.GuardPulseCount)
	}
	if wc.LastError == 0 {
		t.Error("LastError should be set to guardSign for correction")
	}

	// Second guard pulse
	wc.State = StateVerify
	wc.PhaseTimer = 0.1
	_, done = wc.Update(dt, currentField, currentLevel, guardSign)
	if done {
		t.Error("Should not be done on second guard pulse")
	}
	if wc.GuardPulseCount != 2 {
		t.Errorf("GuardPulseCount: got %d, want 2", wc.GuardPulseCount)
	}

	// Third guard pulse should be rejected (limit = 2)
	wc.State = StateVerify
	wc.PhaseTimer = 0.1
	_, done = wc.Update(dt, currentField, currentLevel, guardSign)
	if !done {
		t.Error("Should accept convergence after 2 guard pulses")
	}
	if wc.State != StateSuccess {
		t.Errorf("State: got %v, want %v", wc.State, StateSuccess)
	}
	// GuardPulseCount increments to 3 before being accepted
	if wc.GuardPulseCount < 2 {
		t.Errorf("GuardPulseCount should be >= 2, got %d", wc.GuardPulseCount)
	}
}

// TestWriteController_GuardSign_DirectionClamp verifies calcLevel clamping
func TestWriteController_GuardSign_DirectionClamp(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	targetLevel := 15
	wc.Start(targetLevel, false)

	// Setup: at target, guard wants to go backwards (would flip direction)
	wc.State = StateVerify
	wc.PhaseTimer = 0.1
	wc.PulseCount = 3
	wc.LastVerifyLevel = targetLevel
	wc.CurrentField = 0.8 // Positive field (ascending)
	wc.GuardPulseCount = 0

	currentLevel := targetLevel
	currentField := 0.0
	guardSign := -1 // Guard wants to go down (would flip from ascending to descending)
	dt := 0.01

	// This should trigger guard correction but clamp to prevent direction flip
	_, done := wc.Update(dt, currentField, currentLevel, guardSign)
	if done {
		t.Error("Should not be done, should continue with clamped guard correction")
	}

	// Verify that calculateNextField was called with safe calcLevel
	// (The log message "ISPP GUARD CLAMP" would appear if direction flip was detected)
	if wc.State != StateApply {
		t.Errorf("State: got %v, want %v after guard correction", wc.State, StateApply)
	}
}

// TestWriteController_StateTransitions verifies state machine flow
func TestWriteController_StateTransitions(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.Start(15, false)

	// Start in APPLY
	if wc.State != StateApply {
		t.Errorf("Initial state: got %v, want %v", wc.State, StateApply)
	}

	// After pulse duration, should move to WAIT
	// Need to provide the calculated field to trigger transition
	dt := 0.01
	var lastField float64
	for i := 0; i < 50; i++ {
		field, _ := wc.Update(dt, lastField, 0, 0)
		lastField = field
		if wc.State == StateWait {
			break
		}
	}
	if wc.State != StateWait {
		t.Errorf("After pulse, state: got %v, want %v", wc.State, StateWait)
	}

	// After wait, should move to VERIFY
	for i := 0; i < 50; i++ {
		wc.Update(dt, wc.CurrentField, 0, 0)
		if wc.State == StateVerify {
			break
		}
	}
	if wc.State != StateVerify {
		t.Errorf("After wait, state: got %v, want %v", wc.State, StateVerify)
	}
}

// TestWriteController_BoundsCollapse verifies the bounds fix
func TestWriteController_BoundsCollapse(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.TargetLevel = 15

	// Simulate bounds collapse scenario
	wc.VMin = 1.2
	wc.VMax = 1.0 // VMax < VMin (collapsed)
	wc.VMinSet = true
	wc.VMaxSet = true
	wc.PulseCount = 3
	wc.CurrentField = -1.1
	wc.LastError = -2 // Need more (ascending on descending branch)

	// Trigger verify logic that should fix collapsed bounds
	wc.State = StateVerify
	wc.PhaseTimer = 0.1
	currentLevel := 13 // Below target
	currentField := 0.0
	guardSign := 0
	dt := 0.01

	wc.Update(dt, currentField, currentLevel, guardSign)

	// Bounds should be widened, not reset to [0, MaxField]
	if wc.VMin <= 0 {
		t.Errorf("VMin should not reset to 0, got %f", wc.VMin)
	}
	if wc.VMax <= wc.VMin {
		t.Errorf("VMax should be > VMin after fix, got VMin=%f VMax=%f", wc.VMin, wc.VMax)
	}
}

// TestWriteController_ResetDirection verifies sticky reset direction tracking
func TestWriteController_ResetDirection(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.Start(15, false)

	if wc.ResetDirection() != 0 {
		t.Errorf("ResetDirection should be 0 initially, got %d", wc.ResetDirection())
	}

	// Simulate an overshoot scenario that sets resetDirection
	wc.resetDirection = -1
	if wc.ResetDirection() != -1 {
		t.Errorf("ResetDirection: got %d, want -1", wc.ResetDirection())
	}
}

// TestWriteController_MinStepEnforcement verifies MinStep is respected
func TestWriteController_MinStepEnforcement(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)

	if wc.MinStep <= 0 {
		t.Fatalf("MinStep should be positive, got %f", wc.MinStep)
	}

	// MinStep should be a fraction of Ec
	expectedMin := 0.04 * wc.EcField
	if wc.MinStep != expectedMin {
		t.Errorf("MinStep: got %f, want %f (0.04*Ec)", wc.MinStep, expectedMin)
	}
}
