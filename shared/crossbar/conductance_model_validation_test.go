package crossbar

import (
	"testing"
)

func TestConductanceModelValidation_M2MVM05(t *testing.T) {
	models := []struct {
		name  string
		model ConductanceModel
		table []float64
	}{
		{name: "Linear", model: ConductanceLinear},
		{name: "Exponential", model: ConductanceExponential},
		{
			name:  "Lookup",
			model: ConductanceLookup,
			table: func() []float64 {
				// Strictly increasing lookup table spanning [GMin, GMax].
				t := make([]float64, DefaultQuantizationLevels)
				for i := range t {
					x := float64(i) / float64(len(t)-1)
					t[i] = GMin + x*(GMax-GMin)
				}
				return t
			}(),
		},
	}

	for _, tc := range models {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{Rows: 1, Cols: 1, ADCBits: 8, DACBits: 8, NoiseLevel: 0, ConductanceModel: tc.model, ConductanceTable: tc.table}
			arr, err := NewArray(cfg)
			if err != nil {
				t.Fatalf("NewArray: %v", err)
			}
			// Ensure table is installed for lookup model.
			if tc.model == ConductanceLookup {
				if err := arr.SetConductanceTable(tc.table); err != nil {
					t.Fatalf("SetConductanceTable: %v", err)
				}
			}

			// Bounds + monotonicity over normalized inputs in [0,1].
			prev := -1.0
			for k := 0; k <= 1000; k++ {
				gNorm := float64(k) / 1000.0
				g := arr.GetPhysicalConductance(gNorm)

				if g < GMin-1e-18 || g > GMax+1e-18 {
					t.Fatalf("G out of range at gNorm=%.3f: G=%.6g, expected [%.6g, %.6g]", gNorm, g, GMin, GMax)
				}
				if k > 0 && g+1e-21 < prev {
					t.Fatalf("G not monotonic at gNorm=%.3f: G=%.6g < prev=%.6g", gNorm, g, prev)
				}
				prev = g
			}
			// Endpoint sanity.
			if g0 := arr.GetPhysicalConductance(0); g0 != arr.GetPhysicalConductance(-1) {
				t.Fatalf("expected clamping at low end")
			}
			if g1 := arr.GetPhysicalConductance(1); g1 != arr.GetPhysicalConductance(2) {
				t.Fatalf("expected clamping at high end")
			}
		})
	}
}
