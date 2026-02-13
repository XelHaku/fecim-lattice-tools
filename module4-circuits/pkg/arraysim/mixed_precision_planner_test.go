package arraysim

import "testing"

func TestPlanMixedPrecisionConfig_FindsValidConfig(t *testing.T) {
	cfg, err := PlanMixedPrecisionConfig(MixedPrecisionPlannerInput{
		AccuracyTarget:  0.92,
		EnergyBudgetPJ:  12.0,
		LatencyBudgetNS: 220.0,
	})
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
	if cfg.ExpectedAccuracy < 0.92 {
		t.Fatalf("accuracy constraint violated: got %.4f", cfg.ExpectedAccuracy)
	}
	if cfg.ExpectedEnergyPJ > 12.0 {
		t.Fatalf("energy constraint violated: got %.4f pJ", cfg.ExpectedEnergyPJ)
	}
	if cfg.ExpectedLatencyNS > 220.0 {
		t.Fatalf("latency constraint violated: got %.4f ns", cfg.ExpectedLatencyNS)
	}
	if cfg.Levels <= 0 || cfg.ADCBits <= 0 || cfg.TileRows <= 0 || cfg.TileCols <= 0 {
		t.Fatalf("invalid structural config: %+v", cfg)
	}
}

func TestPlanMixedPrecisionConfig_NoFeasiblePoint(t *testing.T) {
	_, err := PlanMixedPrecisionConfig(MixedPrecisionPlannerInput{
		AccuracyTarget:  0.99,
		EnergyBudgetPJ:  1.0,
		LatencyBudgetNS: 50.0,
	})
	if err == nil {
		t.Fatal("expected infeasible error, got nil")
	}
}
