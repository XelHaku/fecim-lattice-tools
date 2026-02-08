package ferroelectric

import (
	"math"
	"testing"
)

// Table-driven property tests for the Preisach (quasi-static) hysteresis engine.
//
// These focus on qualitative invariants that should hold across refactors:
// - Monotonicity on strictly monotone field ramps (ascending/descending branches)
// - Saturation bounds (no runaway polarization)
// - Approximate odd symmetry for unbiased materials (P(E) \approx -P(-E))
//
// Notes/assumptions:
// - The model includes a reversible dielectric contribution that saturates via tanh.
//   Therefore we allow a small overshoot margin over |Ps|.
// - We test symmetry only at deep saturation and along a monotone ramp starting from
//   negative saturation (where Preisach is expected to be approximately odd).
func TestPreisach_TableDriven_MonotonicityBoundsAndSymmetry(t *testing.T) {
	tests := []struct {
		name     string
		mat      *HZOMaterial
		wantOdd  bool
		maxMult  float64 // multiplier on Ec for the drive amplitude
		minSatPs float64 // minimum |P|/Ps required at saturation
	}{
		{
			name:     "DefaultHZO",
			mat:      DefaultHZO(),
			wantOdd:  true,
			maxMult:  5.0,
			minSatPs: 0.70,
		},
		{
			name:     "LiteratureSuperlattice",
			mat:      LiteratureSuperlattice(),
			wantOdd:  true,
			maxMult:  5.0,
			minSatPs: 0.70,
		},
		{
			name:     "AlScN",
			mat:      AlScN(),
			wantOdd:  true,
			maxMult:  5.0,
			minSatPs: 0.70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mat == nil {
				t.Fatal("nil material")
			}

			m := NewPreisachModel(tt.mat)
			Emax := tt.maxMult * tt.mat.Ec
			if Emax <= 0 {
				t.Fatalf("invalid Emax: %.3e", Emax)
			}

			// Ascending ramp: -Emax -> +Emax
			m.Reset()
			_ = m.Update(-Emax) // ensure we start from negative saturation at the same amplitude

			const n = 120
			prevP := m.Polarization()
			minP, maxP := prevP, prevP
			for i := 1; i <= n; i++ {
				E := -Emax + 2*Emax*float64(i)/float64(n)
				P := m.Update(E)
				if math.IsNaN(P) || math.IsInf(P, 0) {
					t.Fatalf("invalid polarization at i=%d E=%.3e: %v", i, E, P)
				}
				// Allow tiny numerical noise; require overall non-decreasing behavior.
				if P+1e-12 < prevP {
					t.Fatalf("ascending ramp not monotone at i=%d: E=%.3e prevP=%.6e P=%.6e", i, E, prevP, P)
				}
				prevP = P
				if P < minP {
					minP = P
				}
				if P > maxP {
					maxP = P
				}
			}

			// Bounds: total polarization should remain near |Ps| (reversible term included).
			// We allow a modest overshoot margin.
			bound := 1.25 * math.Abs(tt.mat.Ps)
			if bound == 0 {
				bound = 1
			}
			if math.Abs(maxP) > bound || math.Abs(minP) > bound {
				t.Fatalf("polarization out of bounds: minP=%.6e maxP=%.6e bound=±%.6e", minP, maxP, bound)
			}

			// Saturation: at +Emax should be strongly positive.
			if maxP < tt.minSatPs*tt.mat.Ps {
				t.Fatalf("did not reach positive saturation: maxP=%.6e Ps=%.6e", maxP, tt.mat.Ps)
			}
			// and at -Emax should be strongly negative (we started at -Emax).
			if minP > -tt.minSatPs*tt.mat.Ps {
				t.Fatalf("did not reach negative saturation: minP=%.6e Ps=%.6e", minP, tt.mat.Ps)
			}

			// Descending ramp: +Emax -> -Emax should be non-increasing.
			prevP = m.Polarization()
			for i := 1; i <= n; i++ {
				E := Emax - 2*Emax*float64(i)/float64(n)
				P := m.Update(E)
				if P-1e-12 > prevP {
					t.Fatalf("descending ramp not monotone at i=%d: E=%.3e prevP=%.6e P=%.6e", i, E, prevP, P)
				}
				prevP = P
			}

			// Approximate odd symmetry at deep saturation: P(+Emax) \approx -P(-Emax).
			// We check this with a fresh model to avoid minor-path dependence.
			if tt.wantOdd {
				m2 := NewPreisachModel(tt.mat)
				pNeg := m2.Update(-Emax)
				pPos := m2.Update(Emax)
				denom := math.Abs(tt.mat.Ps)
				if denom == 0 {
					denom = 1
				}
				asym := math.Abs(pPos + pNeg) / denom
				if asym > 0.03 { // 3% of Ps
					t.Fatalf("odd-symmetry violation at saturation: P(+)=%.6e P(-)=%.6e asym=%.3f%%", pPos, pNeg, 100*asym)
				}
			}
		})
	}
}
