package controller

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

type isppCycleResult struct {
	targetLevel int
	finalLevel  int
	pulses      int
	maxAbsField float64
	converged   bool
	writeState  WriteState
}

func runISPPFullCyclePreisachLevel(t *testing.T, targetLevel, numLevels int, startPositive bool) isppCycleResult {
	t.Helper()

	mat := ferroelectric.LiteratureSuperlattice()
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	maxSafeField := mat.Ec * 2.5
	wc := NewWriteController(numLevels, mat.Ec, maxSafeField, nil)
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 30
	wc.Start(targetLevel, true)

	// Instrument-like flow: start from opposite saturation branch,
	// then pulse+verify toward the target remanent level.
	p := -mat.Ps
	if startPositive {
		p = mat.Ps
	}

	currentField := 0.0
	maxAbsFieldSeen := 0.0
	finalLevel := levelFromP(p, mat.Ps, numLevels)

	const (
		maxIters = 40000
		dt       = 5e-6
	)

	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(p, mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		if af := math.Abs(targetField); af > maxAbsFieldSeen {
			maxAbsFieldSeen = af
		}

		p = model.Update(currentField)
		finalLevel = levelFromP(p, mat.Ps, numLevels)
		if done {
			break
		}
	}

	return isppCycleResult{
		targetLevel: targetLevel,
		finalLevel:  finalLevel,
		pulses:      wc.TotalPulses + wc.PulseCount,
		maxAbsField: maxAbsFieldSeen,
		converged:   wc.State == StateSuccess,
		writeState:  wc.State,
	}
}

func TestISPPFullCycle_All30Levels_ResearchLikeBehavior(t *testing.T) {
	const numLevels = 30
	mat := ferroelectric.LiteratureSuperlattice()
	safeMaxField := mat.Ec * 2.5

	results := make([]isppCycleResult, 0, numLevels)
	totalPulses := 0
	worstAbsError := 0
	maxAbsFieldOverall := 0.0

	for target := 1; target <= numLevels; target++ {
		resPos := runISPPFullCyclePreisachLevel(t, target, numLevels, true)
		resNeg := runISPPFullCyclePreisachLevel(t, target, numLevels, false)
		res := resPos
		if absInt(resNeg.finalLevel-target) < absInt(resPos.finalLevel-target) ||
			(absInt(resNeg.finalLevel-target) == absInt(resPos.finalLevel-target) && resNeg.pulses < resPos.pulses) {
			res = resNeg
		}

		results = append(results, res)
		totalPulses += res.pulses

		t.Logf("Level %d: %d pulses (final=%d converged=%v state=%s max|E|=%.6g)",
			target, res.pulses, res.finalLevel, res.converged, res.writeState.String(), res.maxAbsField)

		absErr := absInt(res.finalLevel - target)
		if absErr > worstAbsError {
			worstAbsError = absErr
		}
		if res.maxAbsField > maxAbsFieldOverall {
			maxAbsFieldOverall = res.maxAbsField
		}

		if absErr > 1 {
			t.Fatalf("ISPP readback out of tolerance for level %d: final=%d |error|=%d (limit 1)",
				target, res.finalLevel, absErr)
		}
		if res.maxAbsField > safeMaxField+1e-12 {
			t.Fatalf("unsafe write field for level %d: |E|max=%.6g V/m, safe<=%.6g V/m",
				target, res.maxAbsField, safeMaxField)
		}
	}

	avgPulsesPerLevel := float64(totalPulses) / float64(numLevels)
	if avgPulsesPerLevel >= 20.0 {
		t.Fatalf("ISPP pulse budget too high: total=%d avg=%.2f per level (limit <20)",
			totalPulses, avgPulsesPerLevel)
	}

	t.Logf("ISPP full-cycle OK: levels=%d worst|error|=%d totalPulses=%d avgPulses=%.2f max|E|=%.6gV/m safe<=%.6gV/m",
		numLevels, worstAbsError, totalPulses, avgPulsesPerLevel, maxAbsFieldOverall, safeMaxField)
}
