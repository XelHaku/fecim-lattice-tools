package controller

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// TestISPPConverges_LandauK_Ensemble_Superlattice verifies that verify-at-E=0 ISPP
// converges on representative intermediate targets when LK is run in polydomain
// ensemble mode.
func TestISPPConverges_LandauK_Ensemble_Superlattice(t *testing.T) {
	materials := []struct {
		id  string
		mat *sharedphysics.HZOMaterial
	}{
		{id: "fecim_hzo", mat: ferroelectric.FeCIMMaterial()},
		{id: "literature_superlattice", mat: ferroelectric.LiteratureSuperlattice()},
		{id: "default_hzo", mat: ferroelectric.DefaultHZO()},
	}

	numLevels := 30
	targets := []int{5, 10, 15, 20, 25, 29}
	for _, mc := range materials {
		mc := mc
		t.Run(mc.id, func(t *testing.T) {
			defer func() {
				if t.Failed() {
					t.Logf("VERDICT material=%s result=FAIL", mc.id)
					return
				}
				t.Logf("VERDICT material=%s result=PASS", mc.id)
			}()

			mat := mc.mat
			for _, target := range targets {
				target := target
				t.Run("target_level_"+itoa(target), func(t *testing.T) {
					solver := sharedphysics.NewLKSolver()
					solver.ConfigureFromMaterial(mat)
					solver.EnableNoise = false
					solver.UseNLS = true
					solver.EnableEnsemble(96, mat, 1) // fixed deterministic seed for stable convergence coverage

					wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
					wc.EnableLKMidOptimizations = true
					wc.PulseDuration = 2e-3
					wc.MaxRetries = 60 // increased for cumulative NLS model (Guo APL 2018)
					wc.Start(target, true)

					// Saturate on appropriate side before ISPP.
					startP := mat.Ps
					if target > wc.NumLevels/2 {
						startP = -mat.Ps
					}
					solver.SetState(startP)

					currentField := 0.0
					finalLevel := levelFromP(solver.GetState(), mat.Ps, numLevels)
					dt := 5e-6
					bins := ferroelectric.NewLevelBins(mat.Ps, numLevels, 0.98, 0.15)

					for i := 0; i < 50000; i++ {
						curLevel := levelFromP(solver.GetState(), mat.Ps, numLevels)
						guardSign := 0
						if lvl, inError, delta := bins.LevelForP(solver.GetState()); lvl == target && inError {
							if delta > 0 {
								guardSign = 1
							} else if delta < 0 {
								guardSign = -1
							}
							curLevel = lvl
						}
						targetField, done := wc.Update(dt, currentField, curLevel, guardSign)
						currentField = targetField
						solver.Step(currentField, dt)
						finalLevel = levelFromP(solver.GetState(), mat.Ps, numLevels)
						if done {
							break
						}
					}

					if wc.State != StateSuccess {
						t.Fatalf("landauk-ensemble: did not converge: target=%d final=%d pulses=%d state=%s",
							target, finalLevel, wc.TotalPulses+wc.PulseCount, wc.State)
					}

					// Verify-at-0: force additional settle at E=0 and confirm stable target level.
					for i := 0; i < 400; i++ {
						solver.Step(0, dt)
					}
					finalLevel = levelFromP(solver.GetState(), mat.Ps, numLevels)

					if finalLevel != target {
						t.Fatalf("landauk-ensemble: wrong final level after E=0 verify: target=%d final=%d pulses=%d",
							target, finalLevel, wc.TotalPulses+wc.PulseCount)
					}
					if wc.TotalPulses+wc.PulseCount > 100 { // relaxed for cumulative NLS model
						t.Fatalf("landauk-ensemble: too many pulses: target=%d pulses=%d",
							target, wc.TotalPulses+wc.PulseCount)
					}
					t.Logf("landauk-ensemble OK: target=%d final=%d pulses=%d", target, finalLevel, wc.TotalPulses+wc.PulseCount)
				})
			}
		})
	}
}
