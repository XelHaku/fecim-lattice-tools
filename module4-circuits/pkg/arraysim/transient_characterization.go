package arraysim

import (
	"math"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

// CharacterizeTransientResult extracts write/read timing and energy from a
// transient simulation trace.
func CharacterizeTransientResult(config ArrayConfig, result TransientResult) sharedphysics.CharacterizationResult {
	cfg := withAnalysisDefaults(config)
	pr := math.Abs(cfg.Material.Pr)
	writeThreshold := 0.9 * pr

	writeTimeNs := 0.0
	if writeThreshold > 0 {
		for i, p := range result.Polarization {
			if math.Abs(p) >= writeThreshold {
				if i < len(result.TimeNs) {
					writeTimeNs = result.TimeNs[i]
				}
				break
			}
		}
		if writeTimeNs == 0 && len(result.TimeNs) > 0 {
			writeTimeNs = result.TimeNs[len(result.TimeNs)-1]
		}
	}

	readTimeNs := estimateReadSettlingTimeNs(cfg.Sense, result.Current, result.TimeNs)

	return sharedphysics.CharacterizationResult{
		WriteTimeNs:    writeTimeNs,
		ReadTimeNs:     readTimeNs,
		WriteEnergy_fJ: result.Energy_fJ,
		ReadEnergy_fJ:  result.Energy_fJ,
	}
}

func estimateReadSettlingTimeNs(sense SenseChain, current []float64, timeNs []float64) float64 {
	n := len(current)
	if n == 0 || len(timeNs) == 0 {
		return 0
	}
	if len(timeNs) < n {
		n = len(timeNs)
	}
	if n == 1 {
		return timeNs[0]
	}

	tol := math.Abs(sense.CurrentLSB())
	if tol <= 0 {
		tol = 1e-9
	}

	window := 5
	if n < window {
		window = n
	}
	final := 0.0
	for i := n - window; i < n; i++ {
		final += current[i]
	}
	final /= float64(window)

	for i := 0; i < n; i++ {
		stable := true
		for j := i; j < n; j++ {
			if math.Abs(current[j]-final) > tol {
				stable = false
				break
			}
		}
		if stable {
			return timeNs[i]
		}
	}
	return timeNs[n-1]
}
