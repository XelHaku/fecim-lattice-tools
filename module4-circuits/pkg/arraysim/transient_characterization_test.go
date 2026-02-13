package arraysim

import (
	"math"
	"testing"
)

func TestCharacterizeTransientResult_ExtractsTimingAndEnergy(t *testing.T) {
	cfg := testTransientConfig()
	cfg.Sense = SenseChain{
		TIA: TIAConfig{Rf: 10e3, Vref: 0, Vmin: 0, Vmax: 1.0},
		ADC: ADCConfig{Bits: 10, Vmin: 0, Vmax: 1.0},
	}

	res := TransientResult{
		TimeNs:       []float64{1, 2, 3, 4, 5},
		Polarization: []float64{0.1 * cfg.Material.Pr, 0.4 * cfg.Material.Pr, 0.92 * cfg.Material.Pr, 0.95 * cfg.Material.Pr, cfg.Material.Pr},
		Current:      []float64{1e-6, 2e-6, 1.5e-6, 1.50005e-6, 1.50001e-6},
		Energy_fJ:    42.0,
	}

	char := CharacterizeTransientResult(cfg, res)
	if math.Abs(char.WriteTimeNs-3.0) > 1e-9 {
		t.Fatalf("WriteTimeNs: got %.6f ns, want 3.0 ns", char.WriteTimeNs)
	}
	if char.ReadTimeNs <= 0 {
		t.Fatalf("ReadTimeNs should be > 0, got %.6f", char.ReadTimeNs)
	}
	if char.WriteEnergy_fJ != 42.0 || char.ReadEnergy_fJ != 42.0 {
		t.Fatalf("energy mapping mismatch: write=%.3f read=%.3f want=42.0", char.WriteEnergy_fJ, char.ReadEnergy_fJ)
	}
}
