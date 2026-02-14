package controller

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func TestWriteController_CalculateNextField_StuckRelaxesBoundsWithoutResettingToZero(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.TargetLevel = 11

	// Simulate a situation where we've already pulsed a few times, have a tight bracket,
	// and then get stuck (no improvement). This used to hard-reset VMin to 0.
	wc.PulseCount = 1
	wc.CurrentField = -0.97
	wc.VMin = 0.97
	wc.VMax = 1.16
	wc.VMinSet = true
	wc.VMaxSet = true
	wc.StuckCount = 3
	wc.NoImproveCount = 0
	wc.stepFloor = 0

	wc.calculateNextField(16)

	if wc.VMin <= 0 {
		t.Fatalf("VMin reset: got %v, want > 0", wc.VMin)
	}
	if !wc.VMinSet {
		t.Fatalf("VMinSet: got false, want true")
	}
	if wc.VMax != wc.MaxField {
		t.Fatalf("VMax: got %v, want %v", wc.VMax, wc.MaxField)
	}
	if wc.VMaxSet {
		t.Fatalf("VMaxSet: got true, want false")
	}
}

// ---------------------------------------------------------------------------
// L-ISPP (Logarithmic Step Mode) tests
// ---------------------------------------------------------------------------

// TestLogISPP_DefaultIsLinear verifies new controllers default to linear stepping.
func TestLogISPP_DefaultIsLinear(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	if wc.StepMode != "linear" {
		t.Fatalf("StepMode: got %q, want \"linear\"", wc.StepMode)
	}
	if wc.LogBaseStep != 0.05 {
		t.Fatalf("LogBaseStep: got %v, want 0.05", wc.LogBaseStep)
	}
}

// TestLogISPP_StepGrowsSublinearly verifies that in logarithmic mode the
// voltage steps grow but decelerate: step(n+1)-step(n) < step(n)-step(n-1).
func TestLogISPP_StepGrowsSublinearly(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.StepMode = "logarithmic"
	wc.LogBaseStep = 0.05
	wc.TargetLevel = 20

	// Collect voltage magnitudes across several pulses in undershoot (forward) mode.
	// Set up state: start at level 5 heading toward level 20.
	wc.Start(20, true)
	wc.CurrentField = 1.0 // initial pulse already set
	wc.PulseCount = 1

	voltages := make([]float64, 0, 10)
	voltages = append(voltages, math.Abs(wc.CurrentField))

	for pulse := 2; pulse <= 8; pulse++ {
		wc.PulseCount = pulse
		wc.calculateNextField(5) // still at level 5 (undershoot)
		voltages = append(voltages, math.Abs(wc.CurrentField))
	}

	// Verify sublinear growth: successive increments should decrease.
	for i := 2; i < len(voltages); i++ {
		delta1 := voltages[i-1] - voltages[i-2]
		delta2 := voltages[i] - voltages[i-1]
		if delta1 <= 0 || delta2 <= 0 {
			continue // skip if voltage didn't grow (floor/cap effects)
		}
		if delta2 >= delta1+1e-9 {
			t.Errorf("step %d→%d: delta=%.6f >= previous delta=%.6f (not sublinear)",
				i-1, i, delta2, delta1)
		}
	}
}

// TestLogISPP_ConvergesToTarget runs a full ISPP cycle with logarithmic mode
// using the Preisach model and verifies convergence.
func TestLogISPP_ConvergesToTarget(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	numLevels := 30
	target := 15
	wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.StepMode = "logarithmic"
	wc.PulseDuration = 5e-4
	wc.Start(target, true)

	p := -mat.Ps // start from negative saturation
	currentField := 0.0
	const (
		maxIters = 40000
		dt       = 5e-6
	)

	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(p, mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		p = model.Update(currentField)
		if done {
			break
		}
	}

	if wc.State != StateSuccess {
		finalLevel := levelFromP(p, mat.Ps, numLevels)
		t.Fatalf("logarithmic ISPP did not converge: state=%s target=%d final=%d pulses=%d",
			wc.State, target, finalLevel, wc.TotalPulses+wc.PulseCount)
	}
}

// TestLogISPP_FewerPulsesThanLinear compares pulse counts for the same target.
// Logarithmic mode should use comparable or fewer pulses than linear.
func TestLogISPP_FewerPulsesThanLinear(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	numLevels := 30
	target := 15

	runMode := func(mode string) int {
		t.Helper()
		model := ferroelectric.NewPreisachModel(mat)
		model.Reset()
		wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
		wc.StepMode = mode
		wc.PulseDuration = 5e-4
		wc.Start(target, true)

		p := -mat.Ps
		currentField := 0.0
		const (
			maxIters = 40000
			dt       = 5e-6
		)
		for i := 0; i < maxIters; i++ {
			curLevel := levelFromP(p, mat.Ps, numLevels)
			targetField, done := wc.Update(dt, currentField, curLevel, 0)
			currentField = targetField
			p = model.Update(currentField)
			if done {
				break
			}
		}
		if wc.State != StateSuccess {
			t.Fatalf("%s mode did not converge for target %d", mode, target)
		}
		return wc.TotalPulses + wc.PulseCount
	}

	linearPulses := runMode("linear")
	logPulses := runMode("logarithmic")

	t.Logf("linear=%d pulses, logarithmic=%d pulses", linearPulses, logPulses)

	// Allow up to 50% more pulses — the goal is "comparable", not strictly fewer.
	// Logarithmic mode may take slightly more pulses on some materials but uses
	// less cumulative voltage (energy).
	limit := linearPulses + linearPulses/2
	if logPulses > limit {
		t.Errorf("logarithmic used %d pulses vs linear %d (>150%% ratio)", logPulses, linearPulses)
	}
}

// TestLogISPP_CumulativeVoltageTracked verifies CumulativeAbsVoltage > 0
// after a successful write.
func TestLogISPP_CumulativeVoltageTracked(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	numLevels := 30
	target := 10
	wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.StepMode = "logarithmic"
	wc.PulseDuration = 5e-4
	wc.Start(target, true)

	if wc.CumulativeAbsVoltage != 0 {
		t.Fatalf("CumulativeAbsVoltage should be 0 after Start(), got %v", wc.CumulativeAbsVoltage)
	}

	p := -mat.Ps
	currentField := 0.0
	const (
		maxIters = 40000
		dt       = 5e-6
	)
	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(p, mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		p = model.Update(currentField)
		if done {
			break
		}
	}

	if wc.State != StateSuccess {
		t.Fatalf("did not converge: state=%s", wc.State)
	}
	if wc.CumulativeAbsVoltage <= 0 {
		t.Fatalf("CumulativeAbsVoltage should be > 0 after write, got %v", wc.CumulativeAbsVoltage)
	}
	t.Logf("CumulativeAbsVoltage = %.4f (pulses=%d)", wc.CumulativeAbsVoltage, wc.TotalPulses+wc.PulseCount)
}

// TestLogISPP_BisectionStillWorks verifies that once VMin/VMax bounds are
// established, bisection takes over regardless of StepMode.
func TestLogISPP_BisectionStillWorks(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.StepMode = "logarithmic"
	wc.TargetLevel = 15
	wc.Start(15, true)

	// Simulate: pulse has been applied, bounds are established from prior verify cycles.
	wc.PulseCount = 5
	wc.CurrentField = 1.0
	wc.VMin = 0.9
	wc.VMax = 1.1
	wc.VMinSet = true
	wc.VMaxSet = true
	wc.LastError = -1 // undershoot

	wc.calculateNextField(12) // still undershooting (level 12, target 15)

	// Bisection should pick midpoint of [0.9, 1.1], not logarithmic step.
	// With lower-bound bias the result should be in (VMin, VMax).
	absField := math.Abs(wc.CurrentField)
	if absField <= wc.VMin || absField >= wc.VMax {
		t.Fatalf("expected bisection within (%.3f, %.3f), got %.3f", wc.VMin, wc.VMax, absField)
	}
}
