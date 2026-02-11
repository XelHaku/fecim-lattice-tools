package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

func TestMVMTemperatureProfileConductanceWindowGated(t *testing.T) {
	cfg := &Config{Rows: 1, Cols: 1, NoiseLevel: 0.0, ADCBits: 12, DACBits: 12}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	// Pick a high conductance so widening increases the result.
	if err := arr.ProgramWeight(0, 0, 0.8); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}
	input := []float64{1.0}

	// Legacy behavior (no profile): should be temperature-invariant here.
	resRTLegacy, err := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, Temperature: 300})
	if err != nil {
		t.Fatalf("MVM RT legacy: %v", err)
	}
	resCryoLegacy, err := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, Temperature: 77})
	if err != nil {
		t.Fatalf("MVM cryo legacy: %v", err)
	}
	if math.Abs(resRTLegacy.ActualOutput[0]-resCryoLegacy.ActualOutput[0]) > 1e-9 {
		t.Fatalf("expected legacy output to be temperature-invariant; RT=%v cryo=%v", resRTLegacy.ActualOutput[0], resCryoLegacy.ActualOutput[0])
	}

	// With profile: cryogenic should widen window and increase high-G output.
	prof := &TemperatureProfile{Enable: true, ApplyConductanceWindow: true}
	resRT, _ := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, Temperature: 300, TemperatureProfile: prof})
	resCryo, _ := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, Temperature: 77, TemperatureProfile: prof})
	if resCryo.ActualOutput[0] <= resRT.ActualOutput[0] {
		t.Fatalf("expected cryogenic output > RT with conductance-window scaling; RT=%v cryo=%v", resRT.ActualOutput[0], resCryo.ActualOutput[0])
	}
}

func TestMVMTemperatureProfileNoiseScalingIncreasesErrorAtHighT(t *testing.T) {
	// Deterministic per-cell noise factors.
	rand.Seed(123)
	cfg := &Config{Rows: 8, Cols: 8, NoiseLevel: 0.05, ADCBits: 12, DACBits: 12}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			_ = arr.ProgramWeight(i, j, 0.5)
		}
	}
	input := make([]float64, cfg.Cols)
	for j := range input {
		input[j] = 0.5
	}

	prof := &TemperatureProfile{Enable: true, ApplyVariationNoise: true}
	opts300 := &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: true, Temperature: 300, TemperatureProfile: prof}
	opts600 := &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: true, Temperature: 600, TemperatureProfile: prof}

	res300, err := arr.MVMWithNonIdealities(input, opts300)
	if err != nil {
		t.Fatalf("MVM 300K: %v", err)
	}
	res600, err := arr.MVMWithNonIdealities(input, opts600)
	if err != nil {
		t.Fatalf("MVM 600K: %v", err)
	}

	// With noise scaling, higher temperature should increase variation-driven error.
	if res600.RMSE <= res300.RMSE {
		t.Fatalf("expected RMSE to increase with temperature noise scaling; rmse300=%v rmse600=%v", res300.RMSE, res600.RMSE)
	}
}

func TestMVMTemperatureProfileDriftScalingGated(t *testing.T) {
	cfg := &Config{Rows: 1, Cols: 1, NoiseLevel: 0.0, ADCBits: 12, DACBits: 12}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	_ = arr.ProgramWeight(0, 0, 1.0) // far from midpoint
	input := []float64{1.0}

	// No profile: EnableDrift should have no effect (drift approximation is gated).
	resNoProf, _ := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, EnableDrift: true, Temperature: 400})
	resNoProf2, _ := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, EnableDrift: false, Temperature: 400})
	if math.Abs(resNoProf.ActualOutput[0]-resNoProf2.ActualOutput[0]) > 1e-9 {
		t.Fatalf("expected drift to be gated by TemperatureProfile; got withDrift=%v noDrift=%v", resNoProf.ActualOutput[0], resNoProf2.ActualOutput[0])
	}

	prof := &TemperatureProfile{Enable: true, ApplyDrift: true}
	res300, _ := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, EnableDrift: true, Temperature: 300, TemperatureProfile: prof})
	res400, _ := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, EnableDrift: true, Temperature: 400, TemperatureProfile: prof})

	// At higher temperature, drift scaling should increase effective drift and pull
	// conductance toward midpoint, reducing the output.
	if res400.ActualOutput[0] >= res300.ActualOutput[0] {
		t.Fatalf("expected higher-T drift to reduce high-G output; out300=%v out400=%v", res300.ActualOutput[0], res400.ActualOutput[0])
	}
}
