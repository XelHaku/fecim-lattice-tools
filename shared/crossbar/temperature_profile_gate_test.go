package crossbar

import (
	"math"
	"testing"
)

func TestTemperatureProfileMasterEnableGate(t *testing.T) {
	cfg := &Config{Rows: 1, Cols: 1, NoiseLevel: 0.0, ADCBits: 12, DACBits: 12}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	if err := arr.ProgramWeight(0, 0, 0.9); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}

	input := []float64{1.0}
	baseOpts := &MVMOptions{
		EnableIRDrop:     false,
		EnableSneakPaths: false,
		EnableVariation:  false,
		EnableDrift:      true,
		Temperature:      400,
	}

	legacy, err := arr.MVMWithNonIdealities(input, baseOpts)
	if err != nil {
		t.Fatalf("legacy MVM: %v", err)
	}

	profileDisabled := &TemperatureProfile{
		Enable:                 false,
		ApplyConductanceWindow: true,
		ApplyVariationNoise:    true,
		ApplyDrift:             true,
	}
	withDisabledProfile, err := arr.MVMWithNonIdealities(input, &MVMOptions{
		EnableIRDrop:       false,
		EnableSneakPaths:   false,
		EnableVariation:    false,
		EnableDrift:        true,
		Temperature:        400,
		TemperatureProfile: profileDisabled,
	})
	if err != nil {
		t.Fatalf("disabled-profile MVM: %v", err)
	}

	if math.Abs(legacy.ActualOutput[0]-withDisabledProfile.ActualOutput[0]) > 1e-9 {
		t.Fatalf("temperature profile should be fully gated by Enable=false; legacy=%v disabled=%v",
			legacy.ActualOutput[0], withDisabledProfile.ActualOutput[0])
	}
}
