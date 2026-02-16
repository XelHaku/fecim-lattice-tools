package controller

import (
	"fmt"
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
// incremental voltage steps decay: each pulse adds less ΔV than the previous.
func TestLogISPP_StepGrowsSublinearly(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.StepMode = "logarithmic"
	wc.LogBaseStep = 0.30 // Large base step so log decay is visible above MinStep floor
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

	// Verify decaying increments: each delta should be smaller than the previous.
	for i := 2; i < len(voltages); i++ {
		delta1 := voltages[i-1] - voltages[i-2]
		delta2 := voltages[i] - voltages[i-1]
		if delta1 <= 0 || delta2 <= 0 {
			continue // skip if voltage didn't grow (floor/cap effects)
		}
		if delta2 >= delta1+1e-9 {
			t.Errorf("pulse %d→%d: delta=%.6f >= previous delta=%.6f (steps should decay)",
				i-1, i, delta2, delta1)
		}
	}

	// Also verify that voltages are strictly increasing (steps are positive).
	for i := 1; i < len(voltages); i++ {
		if voltages[i] <= voltages[i-1] {
			t.Errorf("voltage did not increase: V[%d]=%.6f <= V[%d]=%.6f",
				i, voltages[i], i-1, voltages[i-1])
		}
	}
}

// TestLogISPP_ConvergesToTarget runs full ISPP cycles with logarithmic mode
// using the Preisach model and verifies convergence for multiple targets.
func TestLogISPP_ConvergesToTarget(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	numLevels := 30

	// Test a range of targets including easy (mid) and hard (near-saturation).
	for _, target := range []int{5, 10, 15, 20, 25} {
		t.Run(fmt.Sprintf("target_%d", target), func(t *testing.T) {
			model := ferroelectric.NewPreisachModel(mat)
			model.Reset()

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

			finalLevel := levelFromP(p, mat.Ps, numLevels)
			pulses := wc.TotalPulses + wc.PulseCount
			t.Logf("target=%d final=%d pulses=%d state=%s", target, finalLevel, pulses, wc.State)

			if wc.State != StateSuccess {
				t.Fatalf("logarithmic ISPP did not converge: state=%s target=%d final=%d pulses=%d",
					wc.State, target, finalLevel, pulses)
			}
		})
	}
}

// TestLogISPP_FewerPulsesThanLinear compares pulse counts across multiple targets.
// Logarithmic mode should use comparable or fewer pulses than linear.
func TestLogISPP_FewerPulsesThanLinear(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	numLevels := 30

	runMode := func(mode string, target int) (int, float64) {
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
		return wc.TotalPulses + wc.PulseCount, wc.CumulativeFluence
	}

	for _, target := range []int{5, 10, 15, 20, 25} {
		t.Run(fmt.Sprintf("target_%d", target), func(t *testing.T) {
			linearPulses, linearEnergy := runMode("linear", target)
			logPulses, logEnergy := runMode("logarithmic", target)

			t.Logf("linear: %d pulses, energy=%.2f | logarithmic: %d pulses, energy=%.2f",
				linearPulses, linearEnergy, logPulses, logEnergy)

			// Allow up to 50% more pulses — the goal is "comparable", not strictly fewer.
			limit := linearPulses + linearPulses/2 + 1 // +1 avoids zero-pulse edge case
			if logPulses > limit {
				t.Errorf("logarithmic used %d pulses vs linear %d (>150%% ratio)", logPulses, linearPulses)
			}
		})
	}
}

// TestLogISPP_CumulativeFluenceTracked verifies CumulativeFluence > 0
// after a successful write.
func TestLogISPP_CumulativeFluenceTracked(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	numLevels := 30
	target := 10
	wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.StepMode = "logarithmic"
	wc.PulseDuration = 5e-4
	wc.Start(target, true)

	if wc.CumulativeFluence != 0 {
		t.Errorf("expected CumulativeFluence 0, got %f", wc.CumulativeFluence)
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
	if wc.CumulativeFluence == 0 {
		t.Error("CumulativeFluence should be > 0 after pulses")
	}
	t.Logf("CumulativeFluence = %.4f (pulses=%d)", wc.CumulativeFluence, wc.TotalPulses+wc.PulseCount)
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
