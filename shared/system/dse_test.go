package system

import "testing"

func TestRunDSE_Defaults(t *testing.T) {
	results := RunDSE(DSEConfig{})
	// 3 sizes × 3 ADC bits × 3 cell bits = 27
	if len(results) != 27 {
		t.Errorf("RunDSE default: got %d results, want 27", len(results))
	}
	for i, r := range results {
		if r.AreaUM2 <= 0 {
			t.Errorf("result[%d].AreaUM2 = %g, want > 0", i, r.AreaUM2)
		}
		if r.EnergyPJ <= 0 {
			t.Errorf("result[%d].EnergyPJ = %g, want > 0", i, r.EnergyPJ)
		}
		if r.LatencyNS <= 0 {
			t.Errorf("result[%d].LatencyNS = %g, want > 0", i, r.LatencyNS)
		}
		if r.PowerUW <= 0 {
			t.Errorf("result[%d].PowerUW = %g, want > 0", i, r.PowerUW)
		}
	}
}

func TestRunDSE_CustomConfig(t *testing.T) {
	cfg := DSEConfig{
		ArraySizes: []int{32, 64},
		ADCBits:    []int{4, 8},
		CellBits:   []int{2, 4},
		TechNode:   Node28nm,
		DeviceType: CellRRAM,
	}
	results := RunDSE(cfg)
	// 2 × 2 × 2 = 8
	if len(results) != 8 {
		t.Errorf("RunDSE custom: got %d results, want 8", len(results))
	}
}

func TestRunDSE_LargerArray_MoreArea(t *testing.T) {
	results64 := RunDSE(DSEConfig{
		ArraySizes: []int{64},
		ADCBits:    []int{4},
		CellBits:   []int{4},
	})
	results128 := RunDSE(DSEConfig{
		ArraySizes: []int{128},
		ADCBits:    []int{4},
		CellBits:   []int{4},
	})
	if results128[0].AreaUM2 <= results64[0].AreaUM2 {
		t.Errorf("128-cell array area %g should be > 64-cell area %g",
			results128[0].AreaUM2, results64[0].AreaUM2)
	}
}
