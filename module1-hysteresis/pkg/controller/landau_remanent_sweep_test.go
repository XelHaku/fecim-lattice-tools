package controller

import (
	"hash/fnv"
	"math"
	"sort"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// Remanent staircase diagnostic for polydomain Landau-Khalatnikov.
//
// Goal: with ensemble enabled, verify-at-E=0 should produce a broad remanent
// staircase (not binary), measured as a distribution of (P_rem, level)
// points over a pulse-amplitude sweep.
func TestLandauKEnsemble_RemanentStaircase_Superlattice(t *testing.T) {
	mat := ferroelectric.LiteratureSuperlattice()

	type sweepMetrics struct {
		levels                 []int
		distinctLevels         int
		hash                   uint64
		decreases              int
		worstDrop              int
		maxRelaxDeltaFrac      float64
		levelChangeAtEndOfZero int
		levelCounts            map[int]int
		levelMeanPrem          map[int]float64
	}

	runSweep := func() sweepMetrics {
		solver := sharedphysics.NewLKSolver()
		solver.ConfigureFromMaterial(mat)
		solver.EnableNoise = false
		solver.UseNLS = true
		solver.EnableEnsemble(256, mat, 0)

		dt := 2e-9
		pulseSteps := 6
		relaxSteps := 20
		sweepSteps := 80
		maxFieldScale := 3.5

		ps := math.Abs(mat.Ps)
		if ps == 0 {
			ps = 1
		}

		seen := map[int]bool{}
		levelCounts := map[int]int{}
		levelPSum := map[int]float64{}
		levels := make([]int, 0, sweepSteps+1)
		maxRelaxDeltaFrac := 0.0
		levelChangeAtEndOfZero := 0
		h := fnv.New64a()

		for k := 0; k <= sweepSteps; k++ {
			mag := (maxFieldScale * float64(k) / float64(sweepSteps)) * mat.Ec
			solver.SetState(-mat.Ps)
			for i := 0; i < pulseSteps; i++ {
				solver.Step(mag, dt)
			}

			prevRelaxP := solver.GetState()
			prevRelaxLevel := levelFromP(prevRelaxP, mat.Ps, 30)
			for i := 0; i < relaxSteps; i++ {
				solver.Step(0, dt)
				if i == relaxSteps-2 {
					prevRelaxP = solver.GetState()
					prevRelaxLevel = levelFromP(prevRelaxP, mat.Ps, 30)
				}
			}

			pRem := solver.GetState()
			lvl := levelFromP(pRem, mat.Ps, 30)
			levels = append(levels, lvl)
			seen[lvl] = true
			levelCounts[lvl]++
			levelPSum[lvl] += pRem
			_, _ = h.Write([]byte{byte(lvl)})

			if lvl != prevRelaxLevel {
				// final two E=0 relax steps should be stable at level mapping resolution
				levelChangeAtEndOfZero++
			}
			deltaFrac := math.Abs(pRem-prevRelaxP) / ps
			if deltaFrac > maxRelaxDeltaFrac {
				maxRelaxDeltaFrac = deltaFrac
			}
		}

		decreases := 0
		worstDrop := 0
		for i := 1; i < len(levels); i++ {
			if levels[i] < levels[i-1] {
				decreases++
				drop := levels[i-1] - levels[i]
				if drop > worstDrop {
					worstDrop = drop
				}
			}
		}

		levelMeanPrem := make(map[int]float64, len(levelPSum))
		for lvl, sum := range levelPSum {
			levelMeanPrem[lvl] = sum / float64(levelCounts[lvl])
		}

		return sweepMetrics{
			levels:                 levels,
			distinctLevels:         len(seen),
			hash:                   h.Sum64(),
			decreases:              decreases,
			worstDrop:              worstDrop,
			maxRelaxDeltaFrac:      maxRelaxDeltaFrac,
			levelChangeAtEndOfZero: levelChangeAtEndOfZero,
			levelCounts:            levelCounts,
			levelMeanPrem:          levelMeanPrem,
		}
	}

	m1 := runSweep()
	m2 := runSweep()

	levels := make([]int, 0, len(m1.levelCounts))
	for lvl := range m1.levelCounts {
		levels = append(levels, lvl)
	}
	sort.Ints(levels)
	for _, lvl := range levels {
		t.Logf("distribution level=%02d count=%d meanP_rem=%0.6e C/m^2", lvl, m1.levelCounts[lvl], m1.levelMeanPrem[lvl])
	}
	t.Logf("distinct remanent levels observed: %d (hash=%016x)", m1.distinctLevels, m1.hash)
	t.Logf("monotonicity (diagnostic): decreases=%d worstDrop=%d", m1.decreases, m1.worstDrop)
	t.Logf("remanent stability at E=0: maxDeltaFrac(last-step)=%0.3e levelChanges=%d", m1.maxRelaxDeltaFrac, m1.levelChangeAtEndOfZero)

	if m1.hash != m2.hash || len(m1.levels) != len(m2.levels) {
		t.Fatalf("expected deterministic sweep hash/length; got hash1=%016x hash2=%016x len1=%d len2=%d", m1.hash, m2.hash, len(m1.levels), len(m2.levels))
	}
	for i := range m1.levels {
		if m1.levels[i] != m2.levels[i] {
			t.Fatalf("expected deterministic sweep levels; mismatch at i=%d: run1=%d run2=%d (hash1=%016x hash2=%016x)", i, m1.levels[i], m2.levels[i], m1.hash, m2.hash)
		}
	}

	// Multilevel claim gate: require broad remanent staircase, not binary behavior.
	if m1.distinctLevels < 20 {
		t.Fatalf("multilevel claim failed: expected >=20 distinct remanent levels, got %d", m1.distinctLevels)
	}

	// Verify-at-E=0 stability guard (quantization-level stable with small final-step drift).
	if m1.levelChangeAtEndOfZero != 0 {
		t.Fatalf("remanent not stable at E=0: level changed in final relax step (%d times)", m1.levelChangeAtEndOfZero)
	}
	if m1.maxRelaxDeltaFrac > 1e-3 {
		t.Fatalf("remanent not stable at E=0: maxDeltaFrac(last-step)=%0.3e exceeds 1e-3", m1.maxRelaxDeltaFrac)
	}
}
