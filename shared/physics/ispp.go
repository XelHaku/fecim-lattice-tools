package physics

import "math"

// AdaptiveISPP implements a simple write-verify loop for reaching a target
// polarization using single-pulse evaluation and voltage bracketing.
//
// Each candidate voltage is evaluated from the same pre-pulse baseline
// polarization (the solver state at entry). This avoids trying to do a binary
// search while applying pulses cumulatively (which breaks monotonicity).
//
// NOTE: This is a pedagogical controller; for richer ISPP diagnostics, use
// WriteController in ispp_write.go.
type AdaptiveISPP struct {
	Solver   *LKSolver
	Material *HZOMaterial

	MaxVoltage      float64
	MinVoltage      float64
	TargetTolerance float64 // fraction of Ps
	MaxIterations   int
	PulseWidth      float64 // seconds
}

func NewAdaptiveISPP(solver *LKSolver, mat *HZOMaterial) *AdaptiveISPP {
	pw := 0.0
	if mat != nil {
		pw = mat.Tau
	}
	if pw <= 0 {
		pw = 10e-9
	}
	return &AdaptiveISPP{
		Solver:          solver,
		Material:        mat,
		MaxVoltage:      3.0,
		MinVoltage:      -3.0,
		TargetTolerance: 0.01,
		MaxIterations:   12,
		PulseWidth:      pw,
	}
}

// PredictState estimates an initial voltage guess using a monotonic atanh model:
//
//	P ≈ Ps*tanh(E/Ec)  →  E ≈ Ec*atanh(P/Ps)  →  V ≈ E*d
func (c *AdaptiveISPP) PredictState(targetP float64) float64 {
	if c == nil || c.Material == nil {
		return 0
	}
	Ps := c.Material.Ps
	if Ps == 0 || c.Material.Ec == 0 || c.Material.Thickness == 0 {
		return 0
	}
	ratio := targetP / Ps
	if ratio > 0.95 {
		ratio = 0.95
	}
	if ratio < -0.95 {
		ratio = -0.95
	}
	eEst := c.Material.Ec * math.Atanh(ratio)
	return eEst * c.Material.Thickness
}

func (c *AdaptiveISPP) BinarySearchWrite(targetP float64) (float64, int, bool) {
	if c == nil || c.Solver == nil || c.Material == nil {
		return 0, 0, false
	}
	if c.Material.Ps == 0 || c.Material.Thickness == 0 || c.PulseWidth <= 0 {
		return c.Solver.GetState(), 0, false
	}

	baseP := c.Solver.GetState()

	vMin := c.MinVoltage
	vMax := c.MaxVoltage
	if vMin > vMax {
		vMin, vMax = vMax, vMin
	}
	// Pick a voltage polarity consistent with the direction we need to move.
	if targetP >= baseP {
		if vMin < 0 {
			vMin = 0
		}
	} else {
		if vMax > 0 {
			vMax = 0
		}
	}

	vNext := c.PredictState(targetP)
	if vNext < vMin {
		vNext = vMin
	}
	if vNext > vMax {
		vNext = vMax
	}

	tolP := c.TargetTolerance * math.Abs(c.Material.Ps)
	if tolP <= 0 {
		tolP = 1e-6
	}

	bestP := baseP
	bestErr := math.Abs(bestP - targetP)

	for iter := 1; iter <= c.MaxIterations; iter++ {
		c.Solver.SetState(baseP)
		eField := vNext / c.Material.Thickness
		c.Solver.Step(eField, c.PulseWidth)
		p := c.Solver.GetState()
		err := p - targetP
		absErr := math.Abs(err)

		if absErr < bestErr {
			bestErr = absErr
			bestP = p
		}
		if absErr <= tolP {
			return p, iter, true
		}

		if err < 0 {
			vMin = vNext
		} else {
			vMax = vNext
		}
		vNext = 0.5 * (vMin + vMax)
	}

	c.Solver.SetState(bestP)
	return bestP, c.MaxIterations, bestErr <= tolP
}
