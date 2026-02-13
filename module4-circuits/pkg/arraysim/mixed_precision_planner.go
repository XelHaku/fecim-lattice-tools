package arraysim

import "fmt"

// MixedPrecisionPlannerInput defines hard constraints and target quality for CIM planning.
type MixedPrecisionPlannerInput struct {
	AccuracyTarget  float64
	EnergyBudgetPJ  float64
	LatencyBudgetNS float64
}

// MixedPrecisionConfig is the planner output.
type MixedPrecisionConfig struct {
	Levels            int
	ADCBits           int
	TileRows          int
	TileCols          int
	ExpectedAccuracy  float64
	ExpectedEnergyPJ  float64
	ExpectedLatencyNS float64
}

// PlanMixedPrecisionConfig searches the discrete design space and returns the
// lowest-energy valid configuration. Ties are broken by lower latency.
func PlanMixedPrecisionConfig(in MixedPrecisionPlannerInput) (MixedPrecisionConfig, error) {
	if in.AccuracyTarget <= 0 || in.AccuracyTarget > 1 {
		return MixedPrecisionConfig{}, fmt.Errorf("accuracy target must be in (0,1], got %.6f", in.AccuracyTarget)
	}
	if in.EnergyBudgetPJ <= 0 {
		return MixedPrecisionConfig{}, fmt.Errorf("energy budget must be > 0, got %.6f", in.EnergyBudgetPJ)
	}
	if in.LatencyBudgetNS <= 0 {
		return MixedPrecisionConfig{}, fmt.Errorf("latency budget must be > 0, got %.6f", in.LatencyBudgetNS)
	}

	levels := []int{8, 16, 32, 64}
	adcBits := []int{4, 5, 6, 7, 8}
	tiles := []int{16, 32, 64, 128}

	best := MixedPrecisionConfig{}
	found := false

	for _, l := range levels {
		for _, b := range adcBits {
			for _, tr := range tiles {
				for _, tc := range tiles {
					acc, en, lat := estimateMixedPrecisionMetrics(l, b, tr, tc)
					if acc < in.AccuracyTarget || en > in.EnergyBudgetPJ || lat > in.LatencyBudgetNS {
						continue
					}
					candidate := MixedPrecisionConfig{
						Levels:            l,
						ADCBits:           b,
						TileRows:          tr,
						TileCols:          tc,
						ExpectedAccuracy:  acc,
						ExpectedEnergyPJ:  en,
						ExpectedLatencyNS: lat,
					}
					if !found || candidate.ExpectedEnergyPJ < best.ExpectedEnergyPJ ||
						(candidate.ExpectedEnergyPJ == best.ExpectedEnergyPJ && candidate.ExpectedLatencyNS < best.ExpectedLatencyNS) {
						best = candidate
						found = true
					}
				}
			}
		}
	}

	if !found {
		return MixedPrecisionConfig{}, fmt.Errorf("no mixed-precision configuration satisfies accuracy>=%.3f, energy<=%.3f pJ, latency<=%.3f ns", in.AccuracyTarget, in.EnergyBudgetPJ, in.LatencyBudgetNS)
	}

	return best, nil
}

func estimateMixedPrecisionMetrics(levels, adcBits, tileRows, tileCols int) (accuracy, energyPJ, latencyNS float64) {
	// Simple monotonic surrogate model:
	// - higher levels/ADC bits improve accuracy but increase cost
	// - larger tiles reduce overhead (lower per-op energy/latency)
	accuracy = 0.86 + 0.03*float64(levels/16) + 0.015*float64(adcBits-4)
	if accuracy > 0.995 {
		accuracy = 0.995
	}

	tileArea := float64(tileRows * tileCols)
	overheadFactor := 1.0 + 2048.0/tileArea
	energyPJ = (3.2 + 0.35*float64(adcBits-4) + 0.25*float64(levels/16)) * overheadFactor
	latencyNS = (55.0 + 7.0*float64(adcBits-4) + 4.0*float64(levels/16)) * overheadFactor
	return accuracy, energyPJ, latencyNS
}
