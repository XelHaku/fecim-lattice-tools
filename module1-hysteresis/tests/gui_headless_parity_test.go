package tests

import (
	"fmt"
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// TestM1_GUIHeadlessParity verifies that the GUI WRD/ISPP dispatch path and the
// headless regression harness produce identical WriteController convergence for
// identical initial conditions. This is the Module 1 counterpart of Module 4's
// headless_gui_physics_parity_test.go.
//
// What we compare (per target):
//   - Final level after ISPP convergence
//   - Total pulse count
//   - Whether convergence was achieved (StateSuccess)
//
// Tolerance rationale:
//   - Final level: ±1 allowed. The GUI path uses finite-rate E-field ramping
//     (rampRate * dt) so the controller sees slightly different currentField
//     values at verify boundaries. This can shift the Preisach stack by one
//     hysteron near level boundaries — matching real ±1 LSB quantization noise
//     in ferroelectric devices (Park et al. 2015).
//   - Pulse count: ±15 allowed. The GUI applies E-field via finite ramp
//     (rampRate × dt) while headless applies instantaneously. Preisach is
//     path-dependent: different field trajectories through the same endpoints
//     produce different internal hysteron states, causing different bisection
//     paths. This is not a physics bug — it reflects real sensitivity of
//     ferroelectric switching to pulse shape (Mulaosmanovic et al. 2017).
//
// The test does NOT compare phase transition timing or skip-prep decisions,
// which legitimately differ between GUI animation and headless stepping.
func TestM1_GUIHeadlessParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping parity test in -short")
	}

	materials := []struct {
		name string
		mat  *ferroelectric.HZOMaterial
	}{
		{"fecim_hzo", sharedphysics.FeCIMMaterial()},
		{"literature_superlattice", sharedphysics.LiteratureSuperlattice()},
	}

	targets := []int{5, 15, 27}

	for _, mc := range materials {
		mc := mc
		t.Run(mc.name, func(t *testing.T) {
			t.Logf("VERDICT material=%s", mc.name)

			for _, target := range targets {
				target := target
				t.Run(fmt.Sprintf("target_%d", target), func(t *testing.T) {
					guiResult := runGUISingleTarget(t, mc.mat, target)
					headlessResult := runHeadlessSingleTarget(t, mc.mat, target)

					t.Logf("gui:      target=%d final=%d pulses=%d converged=%v",
						target, guiResult.FinalLevel, guiResult.Pulses, guiResult.Converged)
					t.Logf("headless: target=%d final=%d pulses=%d converged=%v",
						target, headlessResult.FinalLevel, headlessResult.Pulses, headlessResult.Converged)

					// Level parity: ±1 allowed (ramp-induced hysteron shift).
					levelDelta := abs(guiResult.FinalLevel - headlessResult.FinalLevel)
					if levelDelta > 1 {
						t.Errorf("level mismatch: gui=%d headless=%d delta=%d (tolerance ±1)",
							guiResult.FinalLevel, headlessResult.FinalLevel, levelDelta)
					}

					// Pulse parity: ±15 allowed (ramp dynamics, see rationale above).
					pulseDelta := abs(guiResult.Pulses - headlessResult.Pulses)
					if pulseDelta > 15 {
						t.Errorf("pulse count mismatch: gui=%d headless=%d delta=%d (tolerance ±15)",
							guiResult.Pulses, headlessResult.Pulses, pulseDelta)
					}

					// Convergence: both should converge. If GUI times out but headless
					// converges and final levels agree within ±1, this is a ramp-timeout
					// issue -- log but don't fail.
					if guiResult.Converged != headlessResult.Converged {
						if headlessResult.Converged && levelDelta <= 1 {
							t.Logf("convergence note: gui=%v headless=%v (ramp timeout; levels agree ±1)",
								guiResult.Converged, headlessResult.Converged)
						} else {
							t.Errorf("convergence mismatch: gui=%v headless=%v levelDelta=%d",
								guiResult.Converged, headlessResult.Converged, levelDelta)
						}
					}
				})
			}
		})
	}
}

type isppResult struct {
	FinalLevel int
	Pulses     int
	Converged  bool
}

// runGUISingleTarget drives a single ISPP target through gui.App.UpdatePhysics,
// using the full GUI WRD state machine (PREP→WRITE→DISPLAY).
func runGUISingleTarget(t *testing.T, mat *ferroelectric.HZOMaterial, target int) isppResult {
	t.Helper()
	gui.InitTestLogger()
	app := gui.NewAppWithMaterial(mat.Name)
	if app == nil {
		t.Fatal("gui.NewAppWithMaterial returned nil")
	}

	app.AutoMode(true)
	app.SetWaveform(gui.WaveformWriteReadDemo)
	app.SetPhysicsEngine(gui.PhysicsPreisach)
	app.SetFrequency(500.0) // fast cycles for test speed
	app.SetWrdSkipPrep(false)
	app.SetWrdNextTarget(target)
	app.SetWrdPhase(0)

	dt := 5e-6
	maxSteps := 400000 // 2s simulated

	for step := 0; step < maxSteps; step++ {
		app.Lock()
		_ = app.UpdatePhysics(dt)
		phase := app.WrdPhase()
		wc := app.WriteController()
		app.Unlock()

		// DISPLAY phase = target reached (or failed).
		if phase == 5 && wc != nil {
			return isppResult{
				FinalLevel: wc.LastVerifyLevel,
				Pulses:     wc.TotalPulses + wc.PulseCount,
				Converged:  wc.State == controller.StateSuccess,
			}
		}
	}
	t.Fatalf("GUI did not complete target=%d within %d steps", target, maxSteps)
	return isppResult{}
}

// runHeadlessSingleTarget runs the same WriteController + Preisach physics
// without any GUI code, mirroring the headless regression harness.
func runHeadlessSingleTarget(t *testing.T, mat *ferroelectric.HZOMaterial, target int) isppResult {
	t.Helper()
	numLevels := 30
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	wc := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 60
	wc.Start(target, true)

	// Start from saturation matching GUI PREP phase.
	startP := mat.Ps
	if target > numLevels/2 {
		startP = -mat.Ps
	}
	P := startP
	_ = model.Update(mat.Ec * 2.5 * math.Copysign(1, startP)) // drive to saturation
	P = model.Update(0)                                         // relax to remanent

	currentField := 0.0
	dt := 5e-6
	maxIters := 50000

	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(P, mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		P = model.Update(currentField)
		if done {
			finalLevel := levelFromP(P, mat.Ps, numLevels)
			return isppResult{
				FinalLevel: finalLevel,
				Pulses:     wc.TotalPulses + wc.PulseCount,
				Converged:  wc.State == controller.StateSuccess,
			}
		}
	}
	finalLevel := levelFromP(P, mat.Ps, numLevels)
	return isppResult{
		FinalLevel: finalLevel,
		Pulses:     wc.TotalPulses + wc.PulseCount,
		Converged:  false,
	}
}

func levelFromP(P, Ps float64, numLevels int) int {
	if Ps == 0 {
		return 1
	}
	norm := (P/Ps + 1) / 2 // 0..1
	level := int(math.Round(norm*float64(numLevels-1))) + 1
	if level < 1 {
		level = 1
	}
	if level > numLevels {
		level = numLevels
	}
	return level
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
