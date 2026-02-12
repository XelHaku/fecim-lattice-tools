package validation

import (
	"fmt"
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	circuitsgui "fecim-lattice-tools/module4-circuits/pkg/gui"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// TestM1M4PhysicsConsistency_DefaultHZO30Levels verifies the shared Module 1 ↔ Module 4
// polarization-to-conductance contract for the default 30-level HZO preset.
func TestM1M4PhysicsConsistency_DefaultHZO30Levels(t *testing.T) {
	matM1 := ferroelectric.DefaultHZO()
	if matM1 == nil {
		t.Fatal("module1 DefaultHZO returned nil")
	}

	if levels := matM1.GetNumLevels(); levels != 30 {
		t.Fatalf("expected DefaultHZO to expose 30 levels, got %d", levels)
	}

	ds := circuitsgui.NewDeviceState(1, 1, nil, nil)
	ds.SetMaterial(matM1)
	matM4 := ds.GetMaterial()
	if matM4 == nil {
		t.Fatal("module4 device state material is nil")
	}

	// Guard against material drift between module re-export and shared material object.
	if matM4.Ps != matM1.Ps || matM4.Gmin != matM1.Gmin || matM4.Gmax != matM1.Gmax {
		t.Fatalf("material parameter drift: m1(Ps=%g,Gmin=%g,Gmax=%g) m4(Ps=%g,Gmin=%g,Gmax=%g)",
			matM1.Ps, matM1.Gmin, matM1.Gmax,
			matM4.Ps, matM4.Gmin, matM4.Gmax)
	}

	states := ferroelectric.NewPreisachModel(matM1).DiscreteStates(30)
	if len(states) != 30 {
		t.Fatalf("expected 30 programmed states, got %d", len(states))
	}

	gmin, gmax := matM4.Gmin, matM4.Gmax
	if !(gmin > 0 && gmax > gmin) {
		t.Fatalf("invalid module4 conductance range: Gmin=%g S Gmax=%g S", gmin, gmax)
	}

	prevG := -math.MaxFloat64
	for i, st := range states {
		level := i // Module 4 level indexing is 0..N-1

		// Shared mapping equation from polarization state to conductance.
		gFromP := sharedphysics.PolarizationToConductance(st.Polarization, matM1.Ps, matM1.Gmin, matM1.Gmax)

		// Module 4 uses material.DiscreteLevel(level, levels) for level→G mapping.
		gM4 := matM4.DiscreteLevel(level, len(states))

		// Check consistency between explicit P→G mapping and Module 4 level mapping.
		tol := math.Max(math.Abs(gM4)*1e-12, 1e-18)
		if math.Abs(gFromP-gM4) > tol {
			t.Fatalf("level %d mismatch: shared P→G=%g S module4=%g S (|Δ|=%g > tol=%g)",
				level, gFromP, gM4, math.Abs(gFromP-gM4), tol)
		}

		// Valid range check in Module 4 conductance bounds.
		if gM4 < gmin-1e-18 || gM4 > gmax+1e-18 {
			t.Fatalf("level %d out of range: G=%g S, allowed=[%g,%g] S", level, gM4, gmin, gmax)
		}

		// Monotonicity: higher programmed level must not decrease conductance.
		if level > 0 && gM4 < prevG-1e-18 {
			t.Fatalf("non-monotonic conductance: level %d G=%g S < previous %g S", level, gM4, prevG)
		}
		prevG = gM4
	}

	t.Logf("M1↔M4 physics contract validated: levels=%d, range=[%.3f, %.3f] µS",
		len(states), gmin*1e6, gmax*1e6)
	t.Logf("units sanity: Ps=%.2f µC/cm^2 (%g C/m^2)", matM1.Ps*100, matM1.Ps)
	t.Log(fmt.Sprintf("mapping: G = Gmin + (Gmax-Gmin)*(P/Ps + 1)/2"))
}
