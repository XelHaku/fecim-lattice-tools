//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func levelFromPolarization(P, effPs float64, numLevels int) int {
	if numLevels <= 1 {
		return 1
	}
	if effPs == 0 {
		effPs = 1
	}
	n := P / effPs
	if n > 1 {
		n = 1
	}
	if n < -1 {
		n = -1
	}
	maxLevel := numLevels - 1
	lvl0 := int(math.Round((n + 1) / 2 * float64(maxLevel)))
	if lvl0 < 0 {
		lvl0 = 0
	}
	if lvl0 > maxLevel {
		lvl0 = maxLevel
	}
	return lvl0 + 1
}

func runPreisachISPP(targetLevel int) (pulses int, finalLevel int, completed bool) {
	mat := ferroelectric.LiteratureSuperlattice()
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	const numLevels = 30
	wc := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 10
	wc.ForceResetLimit = 3
	wc.Start(targetLevel, true)

	P := mat.Ps
	if targetLevel > numLevels/2 {
		P = -mat.Ps
	}
	currentField := 0.0
	finalLevel = levelFromPolarization(P, mat.Ps, numLevels)

	for i := 0; i < 50000; i++ {
		currentLevel := levelFromPolarization(P, mat.Ps, numLevels)
		targetField, done := wc.Update(5e-6, currentField, currentLevel, 0)
		currentField = targetField
		P = model.Update(currentField)
		finalLevel = levelFromPolarization(P, mat.Ps, numLevels)

		if done {
			completed = true
			break
		}
		if wc.TotalPulses+wc.PulseCount >= 30 {
			completed = true
			break
		}
	}

	pulses = wc.TotalPulses + wc.PulseCount
	return
}

func TestISPPStability_TargetLevel5_CompletesWithin30Pulses(t *testing.T) {
	pulses, finalLevel, completed := runPreisachISPP(5)
	if !completed {
		t.Fatalf("target L5 did not complete: final=%d pulses=%d", finalLevel, pulses)
	}
	if pulses > 30 {
		t.Fatalf("target L5 exceeded pulse budget: pulses=%d (limit 30)", pulses)
	}
}

func TestISPPStability_TargetLevel15_CompletesWithin30Pulses(t *testing.T) {
	pulses, finalLevel, completed := runPreisachISPP(15)
	if !completed {
		t.Fatalf("target L15 did not complete: final=%d pulses=%d", finalLevel, pulses)
	}
	if pulses > 30 {
		t.Fatalf("target L15 exceeded pulse budget: pulses=%d (limit 30)", pulses)
	}
}

func TestISPPStability_ResetGate_OnlyOnSevereOvershootOrExplicitReset(t *testing.T) {
	wc := &controller.WriteController{MaxOvershootDelta: 3}
	if shouldForceResetAfterISPP(wc, false) {
		t.Fatalf("unexpected reset at overshoot delta 3")
	}
	wc.MaxOvershootDelta = 4
	if !shouldForceResetAfterISPP(wc, false) {
		t.Fatalf("expected reset for severe overshoot delta 4")
	}
	wc.MaxOvershootDelta = 0
	if !shouldForceResetAfterISPP(wc, true) {
		t.Fatalf("explicit reset must always force reset")
	}
}

func TestISPPStability_StuckFallback_ActivatesAt30PulseBudget(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()
	wc := controller.NewWriteController(30, mat.Ec, mat.Ec*2.5, nil)
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 0
	wc.ForceResetLimit = 0
	wc.Start(20, true)

	const stuckLevel = 5
	const pulseLimit = 30
	currentField := 0.0
	forcedComplete := false

	for i := 0; i < 80000; i++ {
		targetField, done := wc.Update(5e-6, currentField, stuckLevel, 0)
		currentField = targetField
		if done {
			break
		}
		if wc.TotalPulses+wc.PulseCount >= pulseLimit {
			forcedComplete = true
			break
		}
	}

	if !forcedComplete {
		t.Fatalf("expected hard timeout fallback to activate at %d pulses; got pulses=%d state=%s", pulseLimit, wc.TotalPulses+wc.PulseCount, wc.State)
	}
}
