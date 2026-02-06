package physics

import (
	"math"
	"testing"
)

func TestAdaptiveISPP_PredictState_SignAndClamp(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	ispp := NewAdaptiveISPP(s, mat)

	tests := []struct {
		name    string
		targetP float64
		wantPos bool
	}{
		{name: "positive", targetP: 0.5 * mat.Ps, wantPos: true},
		{name: "negative", targetP: -0.5 * mat.Ps, wantPos: false},
		{name: "near+Ps", targetP: 1.5 * mat.Ps, wantPos: true},
		{name: "near-Ps", targetP: -1.5 * mat.Ps, wantPos: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ispp.PredictState(tt.targetP)
			if tt.wantPos && v <= 0 {
				t.Fatalf("expected positive voltage estimate, got %g", v)
			}
			if !tt.wantPos && v >= 0 {
				t.Fatalf("expected negative voltage estimate, got %g", v)
			}
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("expected finite estimate, got %g", v)
			}
		})
	}
}

func TestAdaptiveISPP_BinarySearchWrite_Converges(t *testing.T) {
	// Use a simplified LK configuration that behaves approximately linearly:
	//   dP/dt = E/rho, with alpha=beta=gamma=K_dep=0.
	// This keeps the test focused on controller logic.
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat := &HZOMaterial{Ps: 1, Ec: 1, Thickness: 1, Tau: 1}
	ispp := NewAdaptiveISPP(s, mat)
	ispp.MinVoltage = -1
	ispp.MaxVoltage = 1
	ispp.PulseWidth = 1
	ispp.TargetTolerance = 1e-3
	ispp.MaxIterations = 25

	cases := []struct {
		name    string
		baseP   float64
		targetP float64
	}{
		{name: "upwards", baseP: -0.4, targetP: 0.2},
		{name: "downwards", baseP: 0.4, targetP: -0.2},
		{name: "small_adjust_positive", baseP: 0.1, targetP: 0.12},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s.SetState(tc.baseP)
			gotP, iters, ok := ispp.BinarySearchWrite(tc.targetP)
			if iters <= 0 {
				t.Fatalf("expected iterations > 0")
			}
			if !ok {
				t.Fatalf("expected convergence: gotP=%.6g targetP=%.6g iters=%d", gotP, tc.targetP, iters)
			}
			tolP := ispp.TargetTolerance * math.Abs(mat.Ps)
			if math.Abs(gotP-tc.targetP) > tolP {
				t.Fatalf("outside tolerance: gotP=%.6g targetP=%.6g tol=%.6g", gotP, tc.targetP, tolP)
			}
			if math.Abs(s.GetState()-gotP) > 1e-12 {
				t.Fatalf("expected solver state to match return: state=%.6g gotP=%.6g", s.GetState(), gotP)
			}
		})
	}
}
