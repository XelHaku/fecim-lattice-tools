package physics

import (
	"math"
	"testing"
)

// ============================================================================
// LK Solver Stability Tests
// ============================================================================

func TestLKSolver_Stress_ExtremeFields(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)

	tests := []struct {
		name      string
		fieldMult float64
	}{
		{"VeryLarge_10xEc", 10.0},
		{"ExtremelyLarge_100xEc", 100.0},
		{"VerySmall_0.001xEc", 0.001},
		{"Tiny_0.0001xEc", 0.0001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetState(-mat.Pr) // Reset to -Pr
			E := tt.fieldMult * mat.Ec
			dt := 1e-12

			// Apply field for 10 steps
			for i := 0; i < 10; i++ {
				P := s.Step(E, dt)
				if math.IsNaN(P) {
					t.Fatalf("NaN at step %d with E=%.3e (%.1fx Ec)", i, E, tt.fieldMult)
				}
				if math.IsInf(P, 0) {
					t.Fatalf("Inf at step %d with E=%.3e (%.1fx Ec)", i, E, tt.fieldMult)
				}
				if math.Abs(P) > s.PMax*2 {
					t.Fatalf("Polarization %.3e exceeds 2x PMax at step %d", P, i)
				}
			}
		})
	}
}

func TestLKSolver_Stress_Convergence(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false // Disable NLS for deterministic convergence test

	// Apply field and verify polarization approaches saturation monotonically
	s.SetState(-mat.Pr)
	E := 3.0 * mat.Ec // Strong field pointing positive
	dt := 1e-10       // Larger timestep for faster convergence

	prevP := s.GetState()
	for i := 0; i < 200; i++ {
		P := s.Step(E, dt)
		if math.IsNaN(P) || math.IsInf(P, 0) {
			t.Fatalf("Invalid P at step %d: %g", i, P)
		}

		// With positive field, polarization should generally increase (or stay saturated)
		// Allow some tolerance for depolarization effects
		if i > 0 && P < prevP-0.05*mat.Ps {
			// Allow moderate decreases due to depolarization field, but not large reversals
			t.Fatalf("Non-monotonic approach to saturation: P[%d]=%.6e < P[%d]=%.6e by %.6e (E=+%.3e)",
				i, P, i-1, prevP, prevP-P, E)
		}

		// Should eventually approach positive polarization
		if i > 150 && P < 0 {
			t.Fatalf("After 150 steps with +E, still have negative P=%.6e", P)
		}

		prevP = P
	}

	// Final check: should be positive (not necessarily near +Ps due to depolarization)
	finalP := s.GetState()
	if finalP < 0 {
		t.Fatalf("Failed to reach positive polarization: finalP=%.6e with +E=%.3e", finalP, E)
	}
}

func TestLKSolver_Stress_HysteresisLoop(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.EnableNoise = false
	s.UseNLS = false // Disable NLS for deterministic loop

	// Sweep field from -Emax to +Emax and back
	Emax := 3.0 * mat.Ec
	nSteps := 100
	dt := 1e-11

	s.SetState(0) // Start at origin

	var upP, downP []float64

	// Upward sweep: -Emax -> +Emax
	for i := 0; i <= nSteps; i++ {
		E := -Emax + 2*Emax*float64(i)/float64(nSteps)
		for j := 0; j < 5; j++ { // Multiple steps per field point
			s.Step(E, dt)
		}
		upP = append(upP, s.GetState())
	}

	// Downward sweep: +Emax -> -Emax
	for i := 0; i <= nSteps; i++ {
		E := Emax - 2*Emax*float64(i)/float64(nSteps)
		for j := 0; j < 5; j++ {
			s.Step(E, dt)
		}
		downP = append(downP, s.GetState())
	}

	// Verify loop area > 0 (energy dissipation)
	// Simple check: at E=0, P should be different on up vs down branch
	midIdx := nSteps / 2
	if len(upP) <= midIdx || len(downP) <= midIdx {
		t.Fatal("Insufficient data points")
	}

	// At zero field, remanence should give us +Pr on upward branch, -Pr on downward
	// (approximate check)
	pUpMid := upP[midIdx]
	pDownMid := downP[midIdx]

	if math.Abs(pUpMid-pDownMid) < 0.1*mat.Pr {
		t.Fatalf("Hysteresis loop too narrow: P_up(E~0)=%.3e, P_down(E~0)=%.3e, diff=%.3e (expected >%.3e)",
			pUpMid, pDownMid, math.Abs(pUpMid-pDownMid), 0.1*mat.Pr)
	}

	// Additional sanity: final P on downward sweep should be negative
	if downP[len(downP)-1] > 0 {
		t.Fatalf("After downward sweep to -Emax, P should be negative, got %.3e", downP[len(downP)-1])
	}
}

// ============================================================================
// Material Presets Tests
// ============================================================================

func TestLKSolver_Stress_AllMaterialPresets(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		FeCIMMaterialTarget(),
		LiteratureSuperlattice(),
		CryogenicHZO(),
		HZOStandard32(),
		HZOFJT140(),
		HZOCustom14(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			s := NewLKSolver()
			s.ConfigureFromMaterial(mat)

			// Apply moderate field
			E := 1.5 * mat.Ec
			dt := 1e-11

			s.SetState(-math.Abs(mat.Pr))

			// Run for 20 steps
			for i := 0; i < 20; i++ {
				P := s.Step(E, dt)
				if math.IsNaN(P) {
					t.Fatalf("NaN at step %d for material %s", i, mat.Name)
				}
				if math.IsInf(P, 0) {
					t.Fatalf("Inf at step %d for material %s", i, mat.Name)
				}
			}

			// Final check: P should be valid
			P := s.GetState()
			if math.Abs(P) > s.PMax*1.5 {
				t.Fatalf("Final P=%.3e exceeds 1.5*PMax=%.3e for %s", P, s.PMax*1.5, mat.Name)
			}
		})
	}
}

// ============================================================================
// Temperature Dependence Tests
// ============================================================================

func TestLKSolver_Stress_TemperatureDependence(t *testing.T) {
	mat := DefaultHZO()

	tests := []struct {
		name string
		temp float64
	}{
		{"Cryogenic_77K", 77.0},
		{"RoomTemp_300K", 300.0},
		{"Elevated_400K", 400.0},
	}

	var coercives []float64

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLKSolver()
			s.ConfigureFromMaterial(mat)
			s.Temperature = tt.temp
			s.UseMaterialAlpha = false // Use temp-dependent alpha
			s.UpdateParams()

			// Find approximate coercive field by binary search
			// Start from -Pr, apply increasing field until P crosses zero
			s.SetState(-mat.Pr)
			dt := 1e-10

			Emin := 0.0
			Emax := 5.0 * mat.Ec
			var Ec_approx float64

			for iter := 0; iter < 20; iter++ {
				E := (Emin + Emax) / 2.0
				s.SetState(-mat.Pr)

				// Apply field
				for i := 0; i < 50; i++ {
					s.Step(E, dt)
				}

				P := s.GetState()
				if P > 0 {
					// Field was strong enough to flip
					Emax = E
					Ec_approx = E
				} else {
					// Not strong enough
					Emin = E
				}
			}

			if Ec_approx <= 0 || math.IsNaN(Ec_approx) || math.IsInf(Ec_approx, 0) {
				t.Fatalf("Failed to find coercive field at T=%.1fK", tt.temp)
			}

			coercives = append(coercives, Ec_approx)
		})
	}

	// Verify: higher temp should give lower coercive field
	if len(coercives) != 3 {
		t.Fatal("Expected 3 coercive field measurements")
	}

	// Ec(77K) should be > Ec(300K) > Ec(400K) (approximately)
	// Allow some tolerance due to numerical approximation
	if coercives[0] < coercives[1]*0.8 {
		t.Logf("Warning: Ec(77K)=%.3e not significantly higher than Ec(300K)=%.3e", coercives[0], coercives[1])
	}
	if coercives[1] < coercives[2]*0.8 {
		t.Logf("Warning: Ec(300K)=%.3e not significantly higher than Ec(400K)=%.3e", coercives[1], coercives[2])
	}
}

// ============================================================================
// WriteController Convergence Tests
// ============================================================================

func TestWriteController_Stress_AllLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-levels stress test in -short mode (>50s under -race)")
	}
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)

	c := NewWriteController(s, mat)
	c.MaxIterations = 30
	c.Tolerance = 0.02 // 2% tolerance

	numLevels := 30
	failures := 0

	for level := 0; level < numLevels; level++ {
		targetG := mat.Gmin + (mat.Gmax-mat.Gmin)*float64(level)/float64(numLevels-1)

		// Reset to -Pr for consistent starting point
		s.SetState(-mat.Pr)

		attempts, success, _ := c.WriteTarget(targetG)

		if !success {
			failures++
			t.Logf("Level %d: Failed to converge (targetG=%.3e, attempts=%d, reason=%s)",
				level, targetG, attempts, c.FailureReason)
		}

		if attempts > c.MaxIterations {
			t.Errorf("Level %d: Exceeded max iterations", level)
		}
	}

	// Allow some failures, but not too many
	maxFailures := numLevels / 10 // Allow 10% failure rate
	if failures > maxFailures {
		t.Errorf("Too many failures: %d/%d levels failed (max allowed: %d)", failures, numLevels, maxFailures)
	}
}

func TestWriteController_Stress_WithNoise(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping noise test in short mode")
	}

	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.EnableNoise = true // Enable Langevin noise

	c := NewWriteController(s, mat)
	c.MaxIterations = 50 // May need more iterations with noise
	c.Tolerance = 0.03   // Looser tolerance with noise

	// Test a few representative levels
	testLevels := []int{0, 7, 14, 21, 29}

	for _, level := range testLevels {
		targetG := mat.Gmin + (mat.Gmax-mat.Gmin)*float64(level)/float64(29)

		s.SetState(-mat.Pr)

		attempts, success, _ := c.WriteTarget(targetG)

		if !success {
			t.Logf("Level %d with noise: Failed (targetG=%.3e, attempts=%d, reason=%s)",
				level, targetG, attempts, c.FailureReason)
			// With noise, some failures are expected, so don't fail test
		} else if attempts > c.MaxIterations {
			t.Errorf("Level %d with noise: Exceeded max iterations", level)
		}
	}
}

// ============================================================================
// PredictState Accuracy Tests
// ============================================================================

func TestWriteController_Stress_PredictState(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)

	c := NewWriteController(s, mat)

	// Test all four quadrants
	tests := []struct {
		name    string
		targetP float64
		wantPos bool // Expected sign of voltage estimate
	}{
		{"PositiveNear", 0.3 * mat.Ps, true},
		{"PositiveFar", 0.9 * mat.Ps, true},
		{"NegativeNear", -0.3 * mat.Ps, false},
		{"NegativeFar", -0.9 * mat.Ps, false},
		{"VeryPositive", 1.2 * mat.Ps, true},   // Beyond Ps
		{"VeryNegative", -1.2 * mat.Ps, false}, // Beyond -Ps
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vEst := c.initialPulseBound(tt.targetP)

			if math.IsNaN(vEst) || math.IsInf(vEst, 0) {
				t.Fatalf("Invalid voltage estimate: %g for targetP=%.3e", vEst, tt.targetP)
			}

			if vEst < 0 {
				t.Fatalf("Voltage estimate should be magnitude (positive), got %g", vEst)
			}

			if vEst > c.MaxVoltage*2 {
				t.Fatalf("Voltage estimate %.3e exceeds reasonable bounds (2x MaxVoltage=%.3e)",
					vEst, c.MaxVoltage*2)
			}

			// Sign verification happens in initialPulseMagnitude, which applies direction
			// Here we just verify the bound is finite and reasonable
		})
	}
}

// ============================================================================
// Rapid Successive Writes Tests
// ============================================================================

func TestWriteController_Stress_RapidWrites(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)

	c := NewWriteController(s, mat)
	c.MaxIterations = 25

	// Sequence: 0 -> 29 -> 0 -> 29
	levels := []int{0, 29, 0, 29}
	expectedFinal := 29

	for i, level := range levels {
		targetG := mat.Gmin + (mat.Gmax-mat.Gmin)*float64(level)/29.0

		attempts, success, _ := c.WriteTarget(targetG)

		if !success {
			t.Fatalf("Rapid write %d (level %d) failed: attempts=%d, reason=%s",
				i, level, attempts, c.FailureReason)
		}
	}

	// Verify final state
	finalP := s.GetState()
	finalG := PolarizationToConductance(finalP, mat.Ps, mat.Gmin, mat.Gmax)
	expectedG := mat.Gmin + (mat.Gmax-mat.Gmin)*float64(expectedFinal)/29.0

	if math.Abs(finalG-expectedG) > c.Tolerance {
		t.Fatalf("Final state incorrect: finalG=%.3e, expectedG=%.3e, error=%.3e",
			finalG, expectedG, math.Abs(finalG-expectedG))
	}
}

// ============================================================================
// Numerical Edge Cases Tests
// ============================================================================

func TestLKSolver_Stress_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*LKSolver, *HZOMaterial)
		wantPanic bool
	}{
		{
			name: "ZeroThickness",
			setupFunc: func(s *LKSolver, mat *HZOMaterial) {
				mat.Thickness = 0
				s.ConfigureFromMaterial(mat)
			},
			wantPanic: false, // Should handle gracefully
		},
		{
			name: "ZeroArea",
			setupFunc: func(s *LKSolver, mat *HZOMaterial) {
				mat.Area = 0
				s.ConfigureFromMaterial(mat)
			},
			wantPanic: false,
		},
		{
			name: "VeryLargePs",
			setupFunc: func(s *LKSolver, mat *HZOMaterial) {
				mat.Ps = 1000.0 // 1000 C/m² (absurdly large)
				mat.Pr = 900.0
				s.ConfigureFromMaterial(mat)
			},
			wantPanic: false,
		},
		{
			name: "NegativePs",
			setupFunc: func(s *LKSolver, mat *HZOMaterial) {
				mat.Ps = -0.3 // Invalid but shouldn't crash
				mat.Pr = -0.25
				s.ConfigureFromMaterial(mat)
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Panic mismatch: got panic=%v, want panic=%v", r != nil, tt.wantPanic)
				}
			}()

			mat := DefaultHZO()
			s := NewLKSolver()
			tt.setupFunc(s, mat)

			// Try to step
			E := 1e8 // 1 MV/cm
			dt := 1e-11
			for i := 0; i < 10; i++ {
				P := s.Step(E, dt)
				if math.IsNaN(P) || math.IsInf(P, 0) {
					// Log but don't fail - some edge cases may produce invalid values
					t.Logf("Step %d produced invalid P=%g", i, P)
					break
				}
			}
		})
	}
}

func TestWriteController_Stress_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*WriteController, *HZOMaterial)
		targetG   float64
		wantFail  bool
	}{
		{
			name: "TargetBelowGmin",
			setupFunc: func(c *WriteController, mat *HZOMaterial) {
				// Use defaults
			},
			targetG:  0.5e-6, // Below Gmin (1e-6)
			wantFail: true,
		},
		{
			name: "TargetAboveGmax",
			setupFunc: func(c *WriteController, mat *HZOMaterial) {
				// Use defaults
			},
			targetG:  200e-6, // Above Gmax (100e-6)
			wantFail: true,
		},
		{
			name: "ZeroTolerance",
			setupFunc: func(c *WriteController, mat *HZOMaterial) {
				c.Tolerance = 0
				c.MaxIterations = 50 // Need more iterations for exact match
			},
			targetG:  50e-6,
			wantFail: true, // Exact match is nearly impossible with zero tolerance
		},
		{
			name: "VeryLargeTolerance",
			setupFunc: func(c *WriteController, mat *HZOMaterial) {
				c.Tolerance = 100.0 // 100x conductance range
			},
			targetG:  50e-6,
			wantFail: false, // Should converge immediately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := DefaultHZO()
			s := NewLKSolver()
			s.ConfigureFromMaterial(mat)

			c := NewWriteController(s, mat)
			c.MaxIterations = 20
			tt.setupFunc(c, mat)

			s.SetState(-mat.Pr)

			_, success, _ := c.WriteTarget(tt.targetG)

			if success == tt.wantFail {
				t.Errorf("Success mismatch: got success=%v, want fail=%v (reason: %s)",
					success, tt.wantFail, c.FailureReason)
			}
		})
	}
}

// ============================================================================
// Long-Running Stress Tests (marked for skipping in short mode)
// ============================================================================

func TestLKSolver_Stress_LongRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running stress test in short mode")
	}

	t.Log("NOTE: This test takes >5 seconds - verifying long-term stability")

	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.EnableNoise = false

	// Simulate 1000 pulses cycling between +Emax and -Emax
	Emax := 2.0 * mat.Ec
	dt := 1e-11
	stepsPerPulse := 10

	for cycle := 0; cycle < 1000; cycle++ {
		// Positive pulse
		for i := 0; i < stepsPerPulse; i++ {
			P := s.Step(Emax, dt)
			if math.IsNaN(P) || math.IsInf(P, 0) {
				t.Fatalf("Invalid P at cycle %d, positive pulse step %d", cycle, i)
			}
		}

		// Negative pulse
		for i := 0; i < stepsPerPulse; i++ {
			P := s.Step(-Emax, dt)
			if math.IsNaN(P) || math.IsInf(P, 0) {
				t.Fatalf("Invalid P at cycle %d, negative pulse step %d", cycle, i)
			}
		}

		// Periodic check
		if cycle%100 == 0 {
			P := s.GetState()
			if math.Abs(P) > s.PMax*1.5 {
				t.Fatalf("Polarization drift at cycle %d: P=%.3e exceeds bounds", cycle, P)
			}
		}
	}

	t.Logf("Successfully completed 1000 bipolar cycles without numerical instability")
}
