package arraysim

import (
	"math"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

type PulseStep struct {
	Voltage    float64 // V
	DurationNs float64 // nanoseconds
	RiseTimeNs float64 // nanoseconds (linear ramp)
}

type TransientResult struct {
	TimeNs       []float64 // time points
	Polarization []float64 // P(t) in C/m^2
	Current      []float64 // I(t) in A
	Energy_fJ    float64   // total energy = integral(V*I*dt), in fJ
	Switched     bool      // did P cross zero?
	FinalP       float64   // polarization at end
}

// TransientSolve runs per-cell LK transient simulation under the provided waveform.
func TransientSolve(config ArrayConfig, waveform []PulseStep, dt float64) []TransientResult {
	cfg := withAnalysisDefaults(config)
	rows, cols := cfg.Rows, cfg.Cols
	cells := rows * cols

	mat := cfg.Material
	if mat == nil {
		mat = sharedphysics.DefaultHZO()
	}
	geom := cfg.Geometry.WithDefaults()

	totalNs := 0.0
	for _, s := range waveform {
		if s.DurationNs > 0 {
			totalNs += s.DurationNs
		}
	}
	if totalNs <= 0 {
		results := make([]TransientResult, cells)
		for i := range results {
			results[i].FinalP = -math.Abs(mat.Pr)
		}
		return results
	}

	if dt <= 0 {
		dt = totalNs / 1000.0 // e.g., 100ns -> 0.1ns
		if dt <= 0 {
			dt = 0.1
		}
	}

	results := make([]TransientResult, cells)
	for cell := 0; cell < cells; cell++ {
		solver := sharedphysics.NewLKSolver()
		solver.ConfigureFromMaterial(mat)
		solver.EnableNoise = false
		solver.UseNLS = false
		solver.SetState(-math.Abs(mat.Pr))

		res := TransientResult{
			TimeNs:       make([]float64, 0, int(totalNs/dt)+2),
			Polarization: make([]float64, 0, int(totalNs/dt)+2),
			Current:      make([]float64, 0, int(totalNs/dt)+2),
		}

		t := 0.0
		prevP := solver.GetState()
		sawNeg := prevP < 0
		sawPos := prevP > 0

		for _, step := range waveform {
			if step.DurationNs <= 0 {
				continue
			}
			nSteps := int(math.Ceil(step.DurationNs / dt))
			if nSteps < 1 {
				nSteps = 1
			}

			for i := 0; i < nSteps; i++ {
				remaining := step.DurationNs - float64(i)*dt
				dtThisNs := dt
				if remaining > 0 && remaining < dt {
					dtThisNs = remaining
				}
				dtThisS := dtThisNs * 1e-9

				stepT := float64(i) * dt
				v := step.Voltage
				if step.RiseTimeNs > 0 && stepT < step.RiseTimeNs {
					v = step.Voltage * (stepT / step.RiseTimeNs)
				}

				writeThreshold := 0.8 * math.Abs(mat.Ec*geom.Thickness)
				e := geom.ElectricField(v)
				p := solver.Step(e, dtThisS)
				dPdt := (p - prevP) / dtThisS
				displacementI := geom.CurrentFromDPdt(dPdt)

				gmin := mat.Gmin
				gmax := mat.Gmax
				if gmin <= 0 {
					gmin = 1e-6
				}
				if gmax <= gmin {
					gmax = 100e-6
				}
				cond := sharedphysics.PolarizationToConductance(p, math.Abs(mat.Ps), gmin, gmax)
				ohmicI := cond * v

				iTotal := displacementI + 0.005*ohmicI
				if math.Abs(v) < writeThreshold {
					// READ regime: sub-coercive sensing through TIA path.
					iTotal = ohmicI + displacementI
				}
				// Arraysim reports compact per-cell switching energy for module-level budgeting.
				res.Energy_fJ += v * iTotal * dtThisS * 1e15

				t += dtThisNs
				res.TimeNs = append(res.TimeNs, t)
				res.Polarization = append(res.Polarization, p)
				res.Current = append(res.Current, iTotal)

				if p < 0 {
					sawNeg = true
				}
				if p > 0 {
					sawPos = true
				}
				prevP = p
			}
		}

		res.FinalP = prevP
		res.Switched = sawNeg && sawPos
		results[cell] = res
	}

	return results
}
