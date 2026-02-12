package physics

import (
	"math"
	"testing"
)

func TestBoundaryValue_Preisach_EFieldPoints(t *testing.T) {
	const ec = 1.0
	const eMax = 2.0 // exactly 2*Ec for explicit ±2Ec, ±Emax overlap

	ps := NewPreisachStack(eMax, simpleUniformEverett{sat: eMax})

	tests := []struct {
		name     string
		field    float64
		expected float64
	}{
		{name: "-Emax", field: -eMax, expected: -1.0},
		{name: "-2Ec", field: -2 * ec, expected: -1.0},
		{name: "-Ec", field: -ec, expected: -0.5},
		{name: "0", field: 0, expected: 0.0},
		{name: "+Ec", field: ec, expected: 0.5},
		{name: "+2Ec", field: 2 * ec, expected: 1.0},
		{name: "+Emax", field: eMax, expected: 1.0},
	}

	const tol = 1e-12
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ps.ComputePolarization(tc.field)
			if math.Abs(got-tc.expected) > tol {
				t.Fatalf("ComputePolarization(%g) = %.15f, want %.15f", tc.field, got, tc.expected)
			}
		})
	}
}

func TestBoundaryValue_LevelMapping_PolarizationPoints(t *testing.T) {
	const (
		psat = 1.0
		pr   = 0.6
		gmin = 10e-6
		gmax = 100e-6
	)

	tests := []struct {
		name string
		p    float64
	}{
		{name: "-Ps", p: -psat},
		{name: "-Pr", p: -pr},
		{name: "0", p: 0},
		{name: "+Pr", p: pr},
		{name: "+Ps", p: psat},
	}

	const tol = 1e-15
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotG := PolarizationToConductance(tc.p, psat, gmin, gmax)
			expectedG := gmin + (gmax-gmin)*(tc.p/psat+1)/2
			if math.Abs(gotG-expectedG) > tol {
				t.Fatalf("PolarizationToConductance(%g) = %.15e, want %.15e", tc.p, gotG, expectedG)
			}

			gotP := ConductanceToPolarization(gotG, gmin, gmax, psat)
			if math.Abs(gotP-tc.p) > 1e-12 {
				t.Fatalf("ConductanceToPolarization(PolarizationToConductance(%g)) = %.15f, want %.15f", tc.p, gotP, tc.p)
			}
		})
	}
}

func TestBoundaryValue_ConductancePoints(t *testing.T) {
	gmin := GMin
	gmax := GMax
	gmid := (gmin + gmax) / 2

	tests := []struct {
		name      string
		g         float64
		wantNorm  float64
		wantLevel int
	}{
		{name: "Gmin", g: gmin, wantNorm: 0.0, wantLevel: 0},
		{name: "midpoint", g: gmid, wantNorm: 0.5, wantLevel: DefaultLevels / 2},
		{name: "Gmax", g: gmax, wantNorm: 1.0, wantLevel: DefaultLevels - 1},
	}

	const tol = 1e-12
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNorm := PhysicalToNormalizedRange(tc.g, gmin, gmax)
			if math.Abs(gotNorm-tc.wantNorm) > tol {
				t.Fatalf("PhysicalToNormalizedRange(%e) = %.15f, want %.15f", tc.g, gotNorm, tc.wantNorm)
			}

			gotLevel := ConductanceToLevel(tc.g, DefaultLevels)
			if gotLevel != tc.wantLevel {
				t.Fatalf("ConductanceToLevel(%e) = %d, want %d", tc.g, gotLevel, tc.wantLevel)
			}
		})
	}
}
