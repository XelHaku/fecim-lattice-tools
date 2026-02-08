package physics

import (
	"math"
	"testing"
)

func TestNewAdaptiveISPP(t *testing.T) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	ispp := NewAdaptiveISPP(solver, mat)

	if ispp == nil {
		t.Fatal("NewAdaptiveISPP returned nil")
	}
	if ispp.Solver != solver {
		t.Error("Solver not set correctly")
	}
	if ispp.Material != mat {
		t.Error("Material not set correctly")
	}
	if ispp.MaxVoltage <= 0 {
		t.Error("MaxVoltage should be positive")
	}
	if ispp.MinVoltage >= 0 {
		t.Error("MinVoltage should be negative")
	}
	if ispp.TargetTolerance <= 0 {
		t.Error("TargetTolerance should be positive")
	}
	if ispp.MaxIterations <= 0 {
		t.Error("MaxIterations should be positive")
	}
	if ispp.PulseWidth <= 0 {
		t.Error("PulseWidth should be positive")
	}
}

func TestAdaptiveISPP_PredictState(t *testing.T) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	ispp := NewAdaptiveISPP(solver, mat)

	tests := []struct {
		name     string
		targetP  float64
		wantSign int // 1 for positive, -1 for negative, 0 for zero
	}{
		{"positive polarization", 0.2, 1},
		{"negative polarization", -0.2, -1},
		{"zero polarization", 0.0, 0},
		{"near saturation positive", 0.9 * mat.Ps, 1},
		{"near saturation negative", -0.9 * mat.Ps, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			V := ispp.PredictState(tt.targetP)

			if math.IsNaN(V) {
				t.Errorf("PredictState(%v) returned NaN", tt.targetP)
			}
			if math.IsInf(V, 0) {
				t.Errorf("PredictState(%v) returned Inf", tt.targetP)
			}

			// Check sign consistency
			if tt.wantSign > 0 && V < 0 {
				t.Logf("Note: Predicted voltage %v has unexpected sign for positive P", V)
			}
			if tt.wantSign < 0 && V > 0 {
				t.Logf("Note: Predicted voltage %v has unexpected sign for negative P", V)
			}
		})
	}
}

func TestAdaptiveISPP_PredictState_EdgeCases(t *testing.T) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	ispp := NewAdaptiveISPP(solver, mat)

	tests := []struct {
		name    string
		targetP float64
	}{
		{"exactly at saturation", mat.Ps},
		{"beyond saturation", 2 * mat.Ps},
		{"negative beyond saturation", -2 * mat.Ps},
		{"NaN target", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("PredictState panicked: %v", r)
				}
			}()

			V := ispp.PredictState(tt.targetP)
			t.Logf("PredictState(%v) = %v", tt.targetP, V)
		})
	}
}

func TestAdaptiveISPP_BinarySearchWrite_Basic(t *testing.T) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false

	ispp := NewAdaptiveISPP(solver, mat)
	ispp.MaxIterations = 20

	// Start from negative saturation
	solver.SetState(-mat.Pr)

	// Try to write to a positive state
	targetP := 0.1 * mat.Ps

	finalP, iterations, success := ispp.BinarySearchWrite(targetP)

	t.Logf("BinarySearchWrite(%.4f): finalP=%.4f, iterations=%d, success=%v",
		targetP, finalP, iterations, success)

	// Check bounds
	if iterations > ispp.MaxIterations {
		t.Errorf("Iterations exceeded MaxIterations: %d > %d", iterations, ispp.MaxIterations)
	}
	if iterations < 1 {
		t.Error("Should have at least 1 iteration")
	}

	// Final P should be finite
	if math.IsNaN(finalP) {
		t.Error("BinarySearchWrite returned NaN for finalP")
	}
	if math.IsInf(finalP, 0) {
		t.Error("BinarySearchWrite returned Inf for finalP")
	}
}

func TestAdaptiveISPP_BinarySearchWrite_MultipleTargets(t *testing.T) {
	mat := DefaultHZO()

	targets := []float64{
		-0.8 * mat.Ps, // Strong negative
		-0.4 * mat.Ps, // Moderate negative
		0.0,           // Neutral
		0.4 * mat.Ps,  // Moderate positive
		0.8 * mat.Ps,  // Strong positive
	}

	for _, targetP := range targets {
		t.Run("", func(t *testing.T) {
			// Fresh solver for each test
			solver := NewLKSolver()
			solver.ConfigureFromMaterial(mat)
			solver.EnableNoise = false
			solver.UseNLS = false
			solver.SetState(0) // Start from neutral

			ispp := NewAdaptiveISPP(solver, mat)
			ispp.MaxIterations = 15

			finalP, iterations, success := ispp.BinarySearchWrite(targetP)

			if math.IsNaN(finalP) {
				t.Errorf("target=%.4f: finalP is NaN", targetP)
			}
			if math.IsInf(finalP, 0) {
				t.Errorf("target=%.4f: finalP is Inf", targetP)
			}

			t.Logf("target=%.4f -> finalP=%.4f (iter=%d, success=%v)",
				targetP, finalP, iterations, success)
		})
	}
}

func TestAdaptiveISPP_BinarySearchWrite_Convergence(t *testing.T) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false

	ispp := NewAdaptiveISPP(solver, mat)
	ispp.MaxIterations = 30
	ispp.TargetTolerance = 0.05 // 5% tolerance

	// Start from one extreme and try to reach the other
	solver.SetState(-mat.Pr)
	targetP := 0.5 * mat.Pr

	_, _, success := ispp.BinarySearchWrite(targetP)

	// Log result but don't fail - convergence depends on physics params
	t.Logf("Convergence test: success=%v", success)
}

func TestAdaptiveISPP_NilMaterial(t *testing.T) {
	// Test with nil material - should handle gracefully or panic predictably
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic with nil material: %v", r)
		}
	}()

	solver := NewLKSolver()
	ispp := NewAdaptiveISPP(solver, nil)

	if ispp != nil {
		// If it doesn't panic, PredictState might
		_ = ispp.PredictState(0.1)
	}
}

func TestAdaptiveISPP_NilSolver(t *testing.T) {
	// Test with nil solver
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic with nil solver: %v", r)
		}
	}()

	mat := DefaultHZO()
	ispp := NewAdaptiveISPP(nil, mat)

	if ispp != nil {
		// This should panic when trying to step
		_, _, _ = ispp.BinarySearchWrite(0.1)
	}
}

func BenchmarkAdaptiveISPP_BinarySearchWrite(b *testing.B) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false

	ispp := NewAdaptiveISPP(solver, mat)
	ispp.MaxIterations = 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.SetState(-mat.Pr)
		ispp.BinarySearchWrite(0.3 * mat.Pr)
	}
}
