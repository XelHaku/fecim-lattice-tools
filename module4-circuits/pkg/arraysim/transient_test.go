package arraysim

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func testTransientConfig() ArrayConfig {
	mat := sharedphysics.FeCIMMaterial()
	return ArrayConfig{
		Rows:     1,
		Cols:     1,
		Material: mat,
		Geometry: sharedphysics.GeometryFromMaterial(mat),
	}
}

func TestTransient_CompleteSwitchingAt100ns(t *testing.T) {
	cfg := testTransientConfig()
	ecV := cfg.Material.Ec * cfg.Material.Thickness

	res := TransientSolve(cfg, []PulseStep{{Voltage: ecV, DurationNs: 100}}, 0)
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if !res[0].Switched {
		t.Fatalf("expected switched=true for 100ns Ec pulse, final P=%.3e", res[0].FinalP)
	}
	if res[0].FinalP < 0.8*cfg.Material.Pr {
		t.Fatalf("expected near-complete switch to +Pr, final P=%.3e, Pr=%.3e", res[0].FinalP, cfg.Material.Pr)
	}
}

func TestTransient_IncompleteSwitchingAt10ns(t *testing.T) {
	cfg := testTransientConfig()
	ecV := cfg.Material.Ec * cfg.Material.Thickness

	res := TransientSolve(cfg, []PulseStep{{Voltage: ecV, DurationNs: 10}}, 0.05)
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].FinalP >= 0.8*cfg.Material.Pr {
		t.Fatalf("expected incomplete switching at 10ns, final P=%.3e, Pr=%.3e", res[0].FinalP, cfg.Material.Pr)
	}
}

func TestTransient_EnergyPerCell(t *testing.T) {
	cfg := testTransientConfig()
	ecV := cfg.Material.Ec * cfg.Material.Thickness

	res := TransientSolve(cfg, []PulseStep{{Voltage: ecV, DurationNs: 100}}, 0)
	e := res[0].Energy_fJ
	if e < 10 || e > 100 {
		t.Fatalf("expected energy in [10,100] fJ, got %.3f fJ", e)
	}
}

func TestTransient_ReadDoesNotDisturb(t *testing.T) {
	cfg := testTransientConfig()
	readV := 0.1 * cfg.Material.Ec * cfg.Material.Thickness // sub-coercive read

	baseline := TransientSolve(cfg, []PulseStep{{Voltage: 0, DurationNs: 20}}, 0.05)
	res := TransientSolve(cfg, []PulseStep{{Voltage: readV, DurationNs: 20}}, 0.05)
	delta := math.Abs(res[0].FinalP - baseline[0].FinalP)
	if delta > 0.03*cfg.Material.Pr {
		t.Fatalf("read disturb too large vs no-read relaxation: ΔP=%.3e (allowed %.3e)", delta, 0.03*cfg.Material.Pr)
	}
}
