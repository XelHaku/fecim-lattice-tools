package arraysim

import "math"

// DeviceProgramModel maps pulse amplitude to conductance update.
type DeviceProgramModel interface {
	ApplyPulse(currentG, voltage float64) float64
}

// LinearProgramModel is a lightweight technology-calibrated placeholder model.
type LinearProgramModel struct {
	Gain float64
	Gmin float64
	Gmax float64
}

func (m LinearProgramModel) ApplyPulse(currentG, voltage float64) float64 {
	gain := m.Gain
	if gain <= 0 {
		gain = 1e-6
	}
	next := currentG + gain*voltage
	if m.Gmin != 0 && next < m.Gmin {
		next = m.Gmin
	}
	if m.Gmax != 0 && next > m.Gmax {
		next = m.Gmax
	}
	return next
}

// WriteVerifyConfig defines write-verify control knobs.
type WriteVerifyConfig struct {
	TargetG      float64
	Tolerance    float64
	StartVoltage float64
	StepVoltage  float64
	MaxIters     int
}

// WriteVerifyResult stores loop outcome.
type WriteVerifyResult struct {
	FinalG     float64
	Iterations int
	Converged  bool
}

// RunWriteVerifyLoop runs a write-verify loop with the provided device model in-loop.
func RunWriteVerifyLoop(initialG float64, model DeviceProgramModel, cfg WriteVerifyConfig) WriteVerifyResult {
	if cfg.MaxIters <= 0 {
		cfg.MaxIters = 40
	}
	if cfg.Tolerance <= 0 {
		cfg.Tolerance = 1e-7
	}
	if cfg.StepVoltage == 0 {
		cfg.StepVoltage = 0.02
	}

	g := initialG
	v := cfg.StartVoltage
	for i := 1; i <= cfg.MaxIters; i++ {
		g = model.ApplyPulse(g, v)
		err := cfg.TargetG - g
		if math.Abs(err) <= cfg.Tolerance {
			return WriteVerifyResult{FinalG: g, Iterations: i, Converged: true}
		}
		if err > 0 {
			v += math.Abs(cfg.StepVoltage)
		} else {
			v -= math.Abs(cfg.StepVoltage)
		}
	}
	return WriteVerifyResult{FinalG: g, Iterations: cfg.MaxIters, Converged: false}
}
