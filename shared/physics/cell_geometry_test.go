package physics

import "testing"

func TestCellGeometry_ElectricFieldUnitsAndMonotonicity(t *testing.T) {
	g := CellGeometry{Thickness: 10e-9, Area: 1e-14}

	E1 := g.ElectricField(1.0)
	E2 := g.ElectricField(2.0)
	if E2 <= E1 {
		t.Fatalf("expected E to increase with V: E(1V)=%g, E(2V)=%g", E1, E2)
	}

	g2 := CellGeometry{Thickness: 20e-9, Area: 1e-14}
	Ehalf := g2.ElectricField(1.0)
	if Ehalf >= E1 {
		t.Fatalf("expected thicker film to reduce field: t1=%g E1=%g, t2=%g E2=%g", g.Thickness, E1, g2.Thickness, Ehalf)
	}

	// Unit sanity: 1V / 10nm = 1e8 V/m.
	want := 1e8
	if rel := (E1 - want) / want; rel < -1e-12 || rel > 1e-12 {
		t.Fatalf("expected ~%g V/m, got %g", want, E1)
	}
}

func TestCellGeometry_ChargeFromPolarizationScaling(t *testing.T) {
	P := 0.3 // C/m^2
	g1 := CellGeometry{Area: 1e-14}
	g2 := CellGeometry{Area: 2e-14}

	Q1 := g1.ChargeFromPolarization(P)
	Q2 := g2.ChargeFromPolarization(P)
	if Q2 <= Q1 {
		t.Fatalf("expected Q to increase with area: Q1=%g Q2=%g", Q1, Q2)
	}

	want := 2 * Q1
	diff := Q2 - want
	if diff < 0 {
		diff = -diff
	}
	if diff > 1e-24 {
		t.Fatalf("expected Q to scale linearly with area: want %g got %g", want, Q2)
	}
}
