package physics

import "math"

// AgingEngine models coupled wake-up, fatigue, and retention degradation.
type AgingEngine struct {
	PrFresh            float64
	WakeupBoost        float64 // fractional max increase by wake-up
	WakeupTauCycles    float64
	FatigueStartCycle  float64
	FatigueRate        float64
	RetentionTauSec    float64
	CycleHistory       []int
	LastCycle          int
	LastPolarizationPr float64
}

func NewAgingEngine(prFresh float64) *AgingEngine {
	return &AgingEngine{
		PrFresh:           prFresh,
		WakeupBoost:       0.18,
		WakeupTauCycles:   250,
		FatigueStartCycle: 1_000,
		FatigueRate:       0.18,
		RetentionTauSec:   3600 * 24 * 30,
		CycleHistory:      make([]int, 0, 64),
	}
}

// ApplyCycle updates aging state for a target cycle and retention hold time.
func (a *AgingEngine) ApplyCycle(cycle int, holdTimeSec float64) float64 {
	if cycle < 1 {
		cycle = 1
	}
	wakeup := 1 + a.WakeupBoost*(1-math.Exp(-float64(min(cycle, 1000))/a.WakeupTauCycles))
	fatigue := 1.0
	if float64(cycle) > a.FatigueStartCycle {
		x := (float64(cycle) - a.FatigueStartCycle) / (1_000_000 - a.FatigueStartCycle)
		if x > 1 {
			x = 1
		}
		fatigue = math.Exp(-a.FatigueRate * x)
	}
	retention := 1.0
	if holdTimeSec > 0 {
		retention = math.Exp(-holdTimeSec / a.RetentionTauSec)
	}
	pr := a.PrFresh * wakeup * fatigue * retention
	a.LastCycle = cycle
	a.LastPolarizationPr = pr
	a.CycleHistory = append(a.CycleHistory, cycle)
	return pr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
