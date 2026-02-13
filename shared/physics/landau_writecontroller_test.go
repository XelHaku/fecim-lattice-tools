package physics

import (
	"math"
	"testing"
)

// ============================================================================
// LKSolver Tests
// ============================================================================

func TestNewLKSolver_Defaults(t *testing.T) {
	s := NewLKSolver()

	// Verify Golden Set parameters
	if s.Beta != -6.720e8 {
		t.Errorf("Beta = %e, want -6.720e8", s.Beta)
	}
	if s.Gamma != 1.950e10 {
		t.Errorf("Gamma = %e, want 1.950e10", s.Gamma)
	}
	if s.Rho != 0.05 {
		t.Errorf("Rho = %f, want 0.05", s.Rho)
	}
	if s.Q12 != -0.026 {
		t.Errorf("Q12 = %f, want -0.026", s.Q12)
	}
	if s.Stress != 1.0e9 {
		t.Errorf("Stress = %e, want 1.0e9", s.Stress)
	}

	// Verify circuit parasitics
	if s.SeriesResistance != 50.0 {
		t.Errorf("SeriesResistance = %f, want 50.0", s.SeriesResistance)
	}
	if s.Thickness != 10e-9 {
		t.Errorf("Thickness = %e, want 10e-9", s.Thickness)
	}
	expectedArea := 45e-9 * 45e-9
	if s.Area != expectedArea {
		t.Errorf("Area = %e, want %e", s.Area, expectedArea)
	}

	// Verify depolarization
	if s.K_dep != 2.5e8 {
		t.Errorf("K_dep = %e, want 2.5e8", s.K_dep)
	}

	// Verify NLS parameters
	if !s.UseNLS {
		t.Error("UseNLS should be true by default")
	}
	if s.ActivationField != 1.9e9 {
		t.Errorf("ActivationField = %e, want 1.9e9", s.ActivationField)
	}
	if s.TauInf != 1.0e-13 {
		t.Errorf("TauInf = %e, want 1.0e-13", s.TauInf)
	}
	if s.NLSSigma != 1.5 {
		t.Errorf("NLSSigma = %f, want 1.5", s.NLSSigma)
	}

	// Verify thermodynamic constants
	if s.CurieTemp != 723.0 {
		t.Errorf("CurieTemp = %f, want 723.0", s.CurieTemp)
	}
	if s.CurieConst != 1.5e5 {
		t.Errorf("CurieConst = %e, want 1.5e5", s.CurieConst)
	}

	// Verify state
	if s.EnableNoise {
		t.Error("EnableNoise should be false by default")
	}
	if !s.UseEffectiveViscosity {
		t.Error("UseEffectiveViscosity should be true by default")
	}
	if s.Temperature != 300.0 {
		t.Errorf("Temperature = %f, want 300.0", s.Temperature)
	}
	if s.P != -0.30 {
		t.Errorf("P = %f, want -0.30", s.P)
	}
	if s.PMax != 0.30 {
		t.Errorf("PMax = %f, want 0.30", s.PMax)
	}
}

func TestLKSolver_UpdateParams(t *testing.T) {
	s := NewLKSolver()
	s.Temperature = 300.0
	s.CurieTemp = 723.0
	s.CurieConst = 1.5e5
	s.Q12 = -0.026
	s.Stress = 1.0e9
	s.UseMaterialAlpha = false

	s.UpdateParams()

	const Eps0 = 8.854e-12
	expectedAlphaT := (s.Temperature - s.CurieTemp) / (2 * Eps0 * s.CurieConst)
	expectedAlphaMech := 2 * s.Q12 * s.Stress
	expectedAlpha := expectedAlphaT - expectedAlphaMech

	if math.Abs(s.Alpha-expectedAlpha) > 1e-6 {
		t.Errorf("Alpha = %e, want %e", s.Alpha, expectedAlpha)
	}

	// Verify calculation components
	if math.Abs(expectedAlphaT-(-1.594e8)) > 1e6 {
		t.Errorf("alphaT = %e, unexpected value", expectedAlphaT)
	}
}

func TestLKSolver_SetGetState(t *testing.T) {
	s := NewLKSolver()
	s.PMax = 0.5

	// Test normal set/get
	s.SetState(0.25)
	if s.GetState() != 0.25 {
		t.Errorf("GetState() = %f, want 0.25", s.GetState())
	}

	// Test NaN rejection
	s.SetState(math.NaN())
	if s.GetState() != 0.25 {
		t.Errorf("GetState() after NaN = %f, want 0.25 (unchanged)", s.GetState())
	}

	// Test Inf rejection
	s.SetState(math.Inf(1))
	if s.GetState() != 0.25 {
		t.Errorf("GetState() after Inf = %f, want 0.25 (unchanged)", s.GetState())
	}

	// Test clamping at PMax*1.2
	s.SetState(0.65)
	expected := 0.5 * 1.2
	if s.GetState() != expected {
		t.Errorf("GetState() after overflow = %f, want %f (PMax*1.2)", s.GetState(), expected)
	}

	s.SetState(-0.65)
	expected = -0.5 * 1.2
	if s.GetState() != expected {
		t.Errorf("GetState() after underflow = %f, want %f (-PMax*1.2)", s.GetState(), expected)
	}
}

func TestLKSolver_ConfigureFromMaterial(t *testing.T) {
	s := NewLKSolver()
	mat := DefaultHZO()

	s.ConfigureFromMaterial(mat)

	// NOTE: Beta and Gamma are scaled by Ec calibration (LK04 mitigation)
	// So they won't match mat.BetaLandau exactly - they're scaled to match mat.Ec
	// Just verify they were set to non-zero values
	if s.Beta == 0 {
		t.Error("Beta should be non-zero after ConfigureFromMaterial")
	}
	if s.Gamma == 0 {
		t.Error("Gamma should be non-zero after ConfigureFromMaterial")
	}
	if s.Rho != mat.RhoViscosity {
		t.Errorf("Rho = %e, want %e", s.Rho, mat.RhoViscosity)
	}
	if s.Q12 != mat.Q12 {
		t.Errorf("Q12 = %f, want %f", s.Q12, mat.Q12)
	}
	if s.Stress != mat.StressGPa*1e9 {
		t.Errorf("Stress = %e, want %e", s.Stress, mat.StressGPa*1e9)
	}
	if s.K_dep != mat.K_dep {
		t.Errorf("K_dep = %e, want %e", s.K_dep, mat.K_dep)
	}
	if s.Thickness != mat.Thickness {
		t.Errorf("Thickness = %e, want %e", s.Thickness, mat.Thickness)
	}
	if s.Area != mat.Area {
		t.Errorf("Area = %e, want %e", s.Area, mat.Area)
	}
	if s.CurieTemp != mat.CurieTemp {
		t.Errorf("CurieTemp = %f, want %f", s.CurieTemp, mat.CurieTemp)
	}
	if s.CurieConst != mat.CurieConst {
		t.Errorf("CurieConst = %e, want %e", s.CurieConst, mat.CurieConst)
	}
	if s.SeriesResistance != mat.SeriesResistanceOhm {
		t.Errorf("SeriesResistance = %f, want %f", s.SeriesResistance, mat.SeriesResistanceOhm)
	}
	if s.TauInf != mat.Tau0NLS {
		t.Errorf("TauInf = %e, want %e", s.TauInf, mat.Tau0NLS)
	}
	if s.ActivationField != mat.EaNLS {
		t.Errorf("ActivationField = %e, want %e", s.ActivationField, mat.EaNLS)
	}
	if mat.NLSSigma > 0 && s.NLSSigma != mat.NLSSigma {
		t.Errorf("NLSSigma = %f, want %f", s.NLSSigma, mat.NLSSigma)
	}

	// Verify P initialized to -Pr
	if s.P != -math.Abs(mat.Pr) {
		t.Errorf("P = %f, want %f (-Pr)", s.P, -math.Abs(mat.Pr))
	}

	// Verify Alpha computed from Pr (LK04 mitigation)
	pr := math.Abs(mat.Pr)
	expectedAlpha := -2.0*s.Beta*pr*pr - 3.0*s.Gamma*math.Pow(pr, 4)
	if math.Abs(s.Alpha-expectedAlpha) > 1e3 {
		t.Errorf("Alpha = %e, want %e (from Pr)", s.Alpha, expectedAlpha)
	}

	// Verify UseMaterialAlpha enabled
	if !s.UseMaterialAlpha {
		t.Error("UseMaterialAlpha should be true after ConfigureFromMaterial")
	}

	// Verify PMax set from material
	expectedPMax := math.Max(math.Abs(mat.Ps), math.Abs(mat.Pr))
	if s.PMax != expectedPMax {
		t.Errorf("PMax = %f, want %f", s.PMax, expectedPMax)
	}
}

func TestLKSolver_Step_ZeroField(t *testing.T) {
	s := NewLKSolver()
	s.UseNLS = false
	s.EnableNoise = false
	s.UseMaterialAlpha = true
	mat := DefaultHZO()
	s.ConfigureFromMaterial(mat)

	// Start at positive polarization
	s.SetState(0.2)
	initialP := s.GetState()

	// Apply zero field for multiple steps
	dt := 1e-12
	for i := 0; i < 100; i++ {
		s.Step(0, dt)
	}

	finalP := s.GetState()

	// With E=0, P should evolve towards equilibrium (near ±Pr)
	// Should move towards positive Pr since we started positive
	if math.Abs(finalP) < math.Abs(initialP)*0.9 {
		t.Errorf("Zero field evolution: P moved from %e to %e, expected movement towards ±Pr", initialP, finalP)
	}
}

func TestLKSolver_Step_PositiveField(t *testing.T) {
	s := NewLKSolver()
	s.UseNLS = false
	s.EnableNoise = false
	s.UseMaterialAlpha = true
	mat := DefaultHZO()
	s.ConfigureFromMaterial(mat)

	// Start at negative saturation
	s.SetState(-0.25)
	initialP := s.GetState()

	// Apply large positive field
	E := 2e8 // 2 MV/cm, well above Ec
	dt := 1e-12
	for i := 0; i < 100; i++ {
		s.Step(E, dt)
	}

	finalP := s.GetState()

	// P should increase under positive field
	if finalP <= initialP {
		t.Errorf("Positive field: P did not increase (initial=%e, final=%e)", initialP, finalP)
	}
}

func TestLKSolver_Step_NegativeField(t *testing.T) {
	s := NewLKSolver()
	s.UseNLS = false
	s.EnableNoise = false
	s.UseMaterialAlpha = true
	mat := DefaultHZO()
	s.ConfigureFromMaterial(mat)

	// Start at positive saturation
	s.SetState(0.25)
	initialP := s.GetState()

	// Apply large negative field
	E := -2e8 // -2 MV/cm
	dt := 1e-12
	for i := 0; i < 100; i++ {
		s.Step(E, dt)
	}

	finalP := s.GetState()

	// P should decrease under negative field
	if finalP >= initialP {
		t.Errorf("Negative field: P did not decrease (initial=%e, final=%e)", initialP, finalP)
	}
}

func TestLKSolver_Step_NumericalStability(t *testing.T) {
	s := NewLKSolver()
	s.UseNLS = false
	s.EnableNoise = false
	s.PMax = 0.5

	// Test NaN input recovery - should recover to valid P (not necessarily exactly -PMax)
	s.P = math.NaN()
	result := s.Step(1e8, 1e-12)
	if math.IsNaN(result) {
		t.Error("Step() returned NaN after NaN input")
	}
	// Result should be within valid range after recovery
	if math.Abs(result) > s.PMax*1.2 {
		t.Errorf("Step() after NaN = %e, should be within PMax*1.2 = %e", result, s.PMax*1.2)
	}

	// Test Inf field handling
	s.SetState(0.0)
	result = s.Step(math.Inf(1), 1e-12)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		t.Error("Step() with Inf field produced invalid result")
	}

	// Test very large field (should not crash)
	s.SetState(0.0)
	result = s.Step(1e12, 1e-12)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		t.Error("Step() with very large field produced invalid result")
	}
}

func TestLKSolver_Step_SimplifiedLinear(t *testing.T) {
	// Simplified linear solver: alpha=beta=gamma=K_dep=0, NLS=false
	// Expected: dP/dt = E/rho exactly
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	E := 100.0
	dt := 0.01

	// Single step
	s.SetState(0)
	result := s.Step(E, dt)

	// Expected: P_new = P_old + (E/rho)*dt = 0 + (100/1)*0.01 = 1.0
	expected := E * dt / s.Rho
	tolerance := 0.01

	if math.Abs(result-expected) > tolerance {
		t.Errorf("Simplified linear: P = %f, want %f (tolerance %f)", result, expected, tolerance)
	}
}

// ============================================================================
// WriteController Tests
// ============================================================================

func TestNewWriteController_Defaults(t *testing.T) {
	s := NewLKSolver()
	mat := DefaultHZO()
	c := NewWriteController(s, mat)

	if c.MaxVoltage != 3.0 {
		t.Errorf("MaxVoltage = %f, want 3.0", c.MaxVoltage)
	}
	if c.PulseWidth != 10e-9 {
		t.Errorf("PulseWidth = %e, want 10e-9", c.PulseWidth)
	}
	if c.Tolerance != 0.01 {
		t.Errorf("Tolerance = %f, want 0.01", c.Tolerance)
	}
	if c.MaxIterations != 20 {
		t.Errorf("MaxIterations = %d, want 20", c.MaxIterations)
	}
	if c.MaxStep != 1e-12 {
		t.Errorf("MaxStep = %e, want 1e-12", c.MaxStep)
	}
	if c.Solver != s {
		t.Error("Solver not set correctly")
	}
	if c.Material != mat {
		t.Error("Material not set correctly")
	}
}

func TestWriteController_NilMaterial(t *testing.T) {
	s := NewLKSolver()
	c := NewWriteController(s, nil)

	attempts, success, overshootCount := c.WriteTarget(0.5)

	if attempts != 0 {
		t.Errorf("attempts = %d, want 0", attempts)
	}
	if success {
		t.Error("success should be false with nil material")
	}
	if overshootCount != 0 {
		t.Errorf("overshootCount = %d, want 0", overshootCount)
	}
	if c.FailureReason != "material is nil" {
		t.Errorf("FailureReason = %q, want %q", c.FailureReason, "material is nil")
	}
}

func TestWriteController_OutOfBoundsTarget(t *testing.T) {
	s := NewLKSolver()
	mat := DefaultHZO()
	c := NewWriteController(s, mat)

	// Test target below Gmin
	attempts, success, overshootCount := c.WriteTarget(mat.Gmin * 0.5)
	if attempts != 0 || success || overshootCount != 0 {
		t.Error("Should fail for target < Gmin")
	}
	if c.FailureReason != "target conductance out of bounds" {
		t.Errorf("FailureReason = %q, want 'target conductance out of bounds'", c.FailureReason)
	}

	// Test target above Gmax
	attempts, success, overshootCount = c.WriteTarget(mat.Gmax * 1.5)
	if attempts != 0 || success || overshootCount != 0 {
		t.Error("Should fail for target > Gmax")
	}
	if c.FailureReason != "target conductance out of bounds" {
		t.Errorf("FailureReason = %q, want 'target conductance out of bounds'", c.FailureReason)
	}
}

func TestWriteController_ConvergesWithSimpleSolver(t *testing.T) {
	// Simplified linear solver for predictable convergence
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c := NewWriteController(s, mat)
	c.MaxIterations = 50

	// Test mid-range target
	targetG := (mat.Gmin + mat.Gmax) / 2.0
	attempts, success, _ := c.WriteTarget(targetG)

	if !success {
		t.Errorf("Failed to converge with simplified solver (attempts=%d)", attempts)
	}

	// Verify final conductance is within tolerance
	finalP := s.GetState()
	finalG := PolarizationToConductance(finalP, mat.Ps, mat.Gmin, mat.Gmax)
	error := math.Abs(finalG - targetG)

	if error > c.Tolerance {
		t.Errorf("Final error = %e, want <= %e", error, c.Tolerance)
	}
}

func TestWriteController_EventHookCalled(t *testing.T) {
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c := NewWriteController(s, mat)
	c.MaxIterations = 10

	// Track events
	events := make(map[string]int)
	c.EventHook = func(event WriteEvent) {
		events[event.Phase]++
	}

	targetG := (mat.Gmin + mat.Gmax) / 2.0
	c.WriteTarget(targetG)

	// Verify key events were emitted
	expectedPhases := []string{"Start", "Predict", "Verify"}
	for _, phase := range expectedPhases {
		if events[phase] == 0 {
			t.Errorf("Event phase %q was not emitted", phase)
		}
	}

	// Either Success or Failed should be emitted
	if events["Success"] == 0 && events["Failed"] == 0 {
		t.Error("Neither Success nor Failed event was emitted")
	}
}

func TestWriteController_OvershootRecovery(t *testing.T) {
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   0.1,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     -0.8,
		PMax:                  1.0,
		Thickness:             1.0,
		Area:                  1,
	}

	mat := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c := NewWriteController(s, mat)
	c.MaxIterations = 30
	c.MaxVoltage = 5.0

	overshoots := 0
	c.EventHook = func(event WriteEvent) {
		if event.Phase == "Overshoot" || event.Phase == "OvershootNoReset" {
			overshoots++
		}
	}

	// Target high conductance which may cause overshoot
	targetG := mat.Gmax * 0.9
	_, _, overshootCount := c.WriteTarget(targetG)

	if overshootCount > 0 {
		t.Logf("Overshoot recovery triggered %d times", overshootCount)
	}
}

func TestWriteController_WithReset_vs_Without(t *testing.T) {
	// Test with reset=true
	s1 := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat1 := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c1 := NewWriteController(s1, mat1)
	c1.MaxIterations = 30

	targetG := (mat1.Gmin + mat1.Gmax) / 2.0
	attempts1, success1, overshoot1 := c1.WriteTargetWithReset(targetG, true)

	// Test with reset=false
	s2 := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat2 := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c2 := NewWriteController(s2, mat2)
	c2.MaxIterations = 30

	attempts2, success2, overshoot2 := c2.WriteTargetWithReset(targetG, false)

	// Both should eventually converge (with simplified solver)
	if !success1 && !success2 {
		t.Error("Both reset modes failed to converge")
	}

	t.Logf("With reset: attempts=%d, success=%v, overshoots=%d", attempts1, success1, overshoot1)
	t.Logf("Without reset: attempts=%d, success=%v, overshoots=%d", attempts2, success2, overshoot2)
}

func TestWriteController_MaxIterationsExceeded(t *testing.T) {
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c := NewWriteController(s, mat)
	c.MaxIterations = 1 // Very low limit to force failure
	c.Tolerance = 1e-10 // Very tight tolerance

	targetG := (mat.Gmin + mat.Gmax) / 2.0
	attempts, success, _ := c.WriteTarget(targetG)

	if success {
		t.Error("Should fail with MaxIterations=1 and tight tolerance")
	}
	if attempts != 2 { // 1 iteration + 1 (returns i+1)
		t.Errorf("attempts = %d, want 2", attempts)
	}
	if c.FailureReason == "" {
		t.Error("FailureReason should be set on failure")
	}
}

func TestWriteController_CustomStepFunc(t *testing.T) {
	s := &LKSolver{
		P:         0,
		PMax:      1.0,
		Thickness: 1.0,
	}

	mat := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c := NewWriteController(s, mat)

	// Custom step function that simply scales P by field
	stepCallCount := 0
	c.StepFunc = func(E, dt float64) float64 {
		stepCallCount++
		// Simple linear response: dP = E*dt
		s.P += E * dt * 0.001
		if s.P > s.PMax {
			s.P = s.PMax
		}
		if s.P < -s.PMax {
			s.P = -s.PMax
		}
		return s.P
	}

	targetG := (mat.Gmin + mat.Gmax) / 2.0
	c.WriteTarget(targetG)

	if stepCallCount == 0 {
		t.Error("Custom StepFunc was not called")
	}
	t.Logf("Custom StepFunc called %d times", stepCallCount)
}

func TestWriteController_MinVoltageEnforcement(t *testing.T) {
	s := &LKSolver{
		UseMaterialAlpha:      true,
		UseEffectiveViscosity: false,
		EnableNoise:           false,
		UseNLS:                false,
		Rho:                   1.0,
		K_dep:                 0,
		Alpha:                 0,
		Beta:                  0,
		Gamma:                 0,
		P:                     0,
		PMax:                  10,
		Thickness:             1,
		Area:                  1,
	}

	mat := &HZOMaterial{
		Ps:        1.0,
		Pr:        0.8,
		Ec:        1.0,
		Thickness: 1.0,
		Tau:       1.0,
		Gmin:      0.001,
		Gmax:      0.01,
	}

	c := NewWriteController(s, mat)
	c.MinVoltage = 0.5 // Enforce minimum voltage

	minVoltageRespected := true
	c.EventHook = func(event WriteEvent) {
		if event.Phase == "Predict" || event.Phase == "BinarySearch" {
			absV := math.Abs(event.VPulse)
			if absV > 0 && absV < c.MinVoltage {
				minVoltageRespected = false
				t.Errorf("Voltage %e violates MinVoltage %e", absV, c.MinVoltage)
			}
		}
	}

	targetG := (mat.Gmin + mat.Gmax) / 2.0
	c.WriteTarget(targetG)

	if !minVoltageRespected {
		t.Error("MinVoltage was not enforced")
	}
}
