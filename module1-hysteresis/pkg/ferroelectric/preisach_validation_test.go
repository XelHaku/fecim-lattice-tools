package ferroelectric

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// getTestMaterials returns 3 representative materials for validation testing:
// DefaultHZO, FeCIMMaterial, and LiteratureSuperlattice.
func getTestMaterials() []*HZOMaterial {
	return []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		LiteratureSuperlattice(),
	}
}

// TestPreisach_PR1_SaturationSymmetry validates antisymmetry of the major loop (Tier T0)
// Property: P(E) ≈ -P(-E) for the saturation loop
// Test ID: PR1
func TestPreisach_PR1_SaturationSymmetry(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewPreisachModel(mat)
			model.Reset()

			// Trace full major loop with 500 points
			const numPoints = 500
			maxE := mat.Ec * 3.0
			step := (2.0 * maxE) / float64(numPoints-1)

			// Ascending branch: -3*Ec to +3*Ec
			ascendingP := make([]float64, numPoints)
			E := -maxE
			for i := 0; i < numPoints; i++ {
				ascendingP[i] = model.Update(E)
				E += step
			}

			// Descending branch: +3*Ec to -3*Ec
			descendingP := make([]float64, numPoints)
			E = maxE
			for i := 0; i < numPoints; i++ {
				descendingP[i] = model.Update(E)
				E -= step
			}

			// Check antisymmetry: P(E_i) ≈ -P(-E_i)
			// Compare ascending branch with descending branch at opposite fields
			// ascendingP[i] is at E_i = -maxE + i*step
			// We want descendingP at -E_i, which in the descending array is at the same index i
			// (because descending starts at +maxE and goes down)
			maxAsymmetry := 0.0
			for i := 0; i < numPoints; i++ {
				P_asc := ascendingP[i]
				P_desc := descendingP[i]

				// For antisymmetry: P(E) + P(-E) should be small
				asymmetry := math.Abs(P_asc+P_desc) / mat.Ps
				if asymmetry > maxAsymmetry {
					maxAsymmetry = asymmetry
				}
			}

			// Acceptance: max|P(E) + P(-E)| / Ps < 0.01 (1%)
			if maxAsymmetry >= 0.01 {
				t.Errorf("PR1 FAILED: max antisymmetry error %.6f >= 0.01 (1%%)", maxAsymmetry)
			} else {
				t.Logf("PR1 PASSED: max antisymmetry error %.6f < 0.01 (1%%)", maxAsymmetry)
			}
		})
	}
}

// TestPreisach_PR2_SaturationValues validates that saturation polarization is reached (Tier T1)
// Property: P(3*Ec) > 0.70 * Ps
// Test ID: PR2
func TestPreisach_PR2_SaturationValues(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewPreisachModel(mat)
			model.Reset()

			// Apply saturation field
			E_sat := mat.Ec * 3.0
			P_sat := model.Update(E_sat)

			ratio := P_sat / mat.Ps

			// Acceptance: P(3*Ec) > 0.70 * Ps
			if ratio <= 0.70 {
				t.Errorf("PR2 FAILED: P_sat/Ps = %.4f <= 0.70", ratio)
			} else {
				t.Logf("PR2 PASSED: P_sat/Ps = %.4f > 0.70", ratio)
			}
		})
	}
}

// TestPreisach_PR3_CoerciveFieldAccuracy validates coercive field measurement (Tier T1)
// Property: Effective Ec is within reasonable range of material Ec
// Note: Preisach model's effective coercive field (where P=0) depends on the
// distribution shape (Delta parameter). It's an emergent property, not a direct input.
// The material's Ec parameter tunes the distribution width, but doesn't directly
// determine the zero-crossing field.
// Test ID: PR3
func TestPreisach_PR3_CoerciveFieldAccuracy(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewPreisachModel(mat)
			model.Reset()

			// Saturate positive
			model.Update(mat.Ec * 3.0)

			// Trace descending branch to find zero crossing
			const numPoints = 1000
			maxE := mat.Ec * 3.0
			step := (2.0 * maxE) / float64(numPoints-1)

			E_prev := maxE
			P_prev := model.Update(E_prev)
			Ec_measured := 0.0
			found := false

			for i := 1; i < numPoints; i++ {
				E_curr := maxE - float64(i)*step
				P_curr := model.Update(E_curr)

				// Check for zero crossing
				if P_prev > 0 && P_curr <= 0 {
					// Linear interpolation to find exact crossing
					t := -P_prev / (P_curr - P_prev)
					Ec_measured = E_prev + t*(E_curr-E_prev)
					found = true
					break
				}

				E_prev = E_curr
				P_prev = P_curr
			}

			if !found {
				t.Errorf("PR3 FAILED: no zero crossing found in descending branch")
				return
			}

			// Acceptance: Effective |Ec| within factor of 3 (accounts for distribution shape)
			// Typical Preisach with Delta=0.25*Ec has effective Ec ≈ 2.5-3x material Ec
			// Take absolute value since descending branch crosses zero at negative field
			ratio := math.Abs(Ec_measured) / mat.Ec
			if ratio < 0.5 || ratio > 4.0 {
				t.Errorf("PR3 FAILED: |Ec_measured|/Ec = %.4f outside [0.5, 4.0]", ratio)
			} else {
				t.Logf("PR3 PASSED: Ec_measured = %.6f MV/cm, Ec = %.6f MV/cm, |Ec_measured|/Ec = %.4f ∈ [0.5, 4.0]",
					Ec_measured, mat.Ec, ratio)
			}
		})
	}
}

// TestPreisach_PR4_RemanentPolarization validates remanent polarization (Tier T1)
// Property: P at E=0 after saturation > 0.30 * Ps
// Test ID: PR4
func TestPreisach_PR4_RemanentPolarization(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewPreisachModel(mat)
			model.Reset()

			// Saturate positive
			model.Update(mat.Ec * 3.0)

			// Return to E=0
			P_rem := model.Update(0.0)

			ratio := P_rem / mat.Ps

			// Acceptance: P at E=0 > 0.30 * Ps
			if ratio <= 0.30 {
				t.Errorf("PR4 FAILED: Pr/Ps = %.4f <= 0.30", ratio)
			} else {
				t.Logf("PR4 PASSED: Pr/Ps = %.4f > 0.30", ratio)
			}
		})
	}
}

// TestPreisach_PR5_ReturnPointMemory validates return-point memory property (Tier T0)
// Property: After minor loop excursion, returning to same field gives same polarization
// Test ID: PR5
func TestPreisach_PR5_ReturnPointMemory(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Use raw PreisachStack for precision
			ev := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.25}
			stack := physics.NewPreisachStack(mat.Ec*5, ev)

			// Saturate positive
			stack.Update(mat.Ec * 3.0)

			// Ascending major loop to E1
			E1 := mat.Ec * 0.5
			P_initial := stack.Update(E1)

			// Minor excursion to E2
			E2 := mat.Ec * -0.3
			stack.Update(E2)

			// Return to E1
			P_return := stack.Update(E1)

			// Acceptance: |P_return - P_initial| / Ps < 0.01 (1%)
			// Note: Ideal Preisach has perfect return-point memory, but numerical
			// implementation with finite-precision Everett integrals has small errors
			relError := math.Abs(P_return-P_initial) / mat.Ps
			if relError >= 0.01 {
				t.Errorf("PR5 FAILED: |P_return - P_initial| / Ps = %.6f >= 0.01", relError)
			} else {
				t.Logf("PR5 PASSED: return-point memory error = %.6f < 0.01 (1%%)", relError)
			}
		})
	}
}

// TestPreisach_PR6_WipeOut validates wipe-out property (Tier T0)
// Property: Exceeding a previous turning point erases inner loops
// Test ID: PR6
func TestPreisach_PR6_WipeOut(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Use raw PreisachStack
			ev := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.25}
			stack := physics.NewPreisachStack(mat.Ec*5, ev)

			// Create nested minor loop
			E1 := mat.Ec * 2.0
			stack.Update(E1)

			E2 := mat.Ec * -0.5
			stack.Update(E2)

			E3 := mat.Ec * 0.5
			stack.Update(E3) // Inner loop created

			stackLenBefore := len(stack.Stack)

			// Wipe out by exceeding E1
			E4 := mat.Ec * 3.0
			stack.Update(E4)

			stackLenAfter := len(stack.Stack)

			// Acceptance: stack length decreases (inner loop erased)
			if stackLenAfter >= stackLenBefore {
				t.Errorf("PR6 FAILED: stack length after wipe-out (%d) >= before (%d)",
					stackLenAfter, stackLenBefore)
			} else {
				t.Logf("PR6 PASSED: stack length decreased from %d to %d after wipe-out",
					stackLenBefore, stackLenAfter)
			}
		})
	}
}

// TestPreisach_PR7_CongruentMinorLoops validates congruency of minor loops (Tier T1)
// Property: Minor loops at different bias points should have similar delta_P
// Test ID: PR7
func TestPreisach_PR7_CongruentMinorLoops(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Use raw PreisachStack
			ev := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.25}

			// Test at 3 different bias points
			biasPoints := []float64{-0.5 * mat.Ec, 0.0, 0.5 * mat.Ec}
			deltaE := mat.Ec * 0.3
			deltaPValues := make([]float64, len(biasPoints))

			for i, bias := range biasPoints {
				stack := physics.NewPreisachStack(mat.Ec*5, ev)

				// Saturate and go to bias
				stack.Update(mat.Ec * 3.0)
				stack.Update(bias)

				// Apply minor loop
				P1 := stack.Update(bias)
				P2 := stack.Update(bias + deltaE)
				stack.Update(bias) // Return to bias

				deltaPValues[i] = math.Abs(P2 - P1)
			}

			// Calculate coefficient of variation
			mean := 0.0
			for _, dp := range deltaPValues {
				mean += dp
			}
			mean /= float64(len(deltaPValues))

			variance := 0.0
			for _, dp := range deltaPValues {
				diff := dp - mean
				variance += diff * diff
			}
			variance /= float64(len(deltaPValues))
			stdDev := math.Sqrt(variance)
			cv := stdDev / mean

			// Acceptance: CV < 0.15 (15%)
			if cv >= 0.15 {
				t.Errorf("PR7 FAILED: coefficient of variation %.4f >= 0.15", cv)
				t.Logf("  delta_P values: %v", deltaPValues)
			} else {
				t.Logf("PR7 PASSED: coefficient of variation %.4f < 0.15", cv)
			}
		})
	}
}

// TestPreisach_EV1_EverettNonNegativity validates non-negativity of Everett function (Tier T0)
// Property: Everett(alpha, beta) >= 0 for all alpha >= beta
// Test ID: EV1
func TestPreisach_EV1_EverettNonNegativity(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			ev := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.25}

			// Grid sweep
			maxE := mat.Ec * 3.0
			step := mat.Ec * 0.1
			violations := 0
			minValue := 0.0

			for alpha := -maxE; alpha <= maxE; alpha += step {
				for beta := -maxE; beta <= alpha; beta += step {
					W := ev.Calculate(alpha, beta)
					if W < -1e-15 {
						violations++
						if W < minValue {
							minValue = W
						}
					}
				}
			}

			// Acceptance: all values >= -1e-15
			if violations > 0 {
				t.Errorf("EV1 FAILED: %d violations of non-negativity, min value = %.16e",
					violations, minValue)
			} else {
				t.Logf("EV1 PASSED: all Everett values >= -1e-15")
			}
		})
	}
}

// TestPreisach_EV2_EverettMonotonicity validates monotonicity of Everett function (Tier T0)
// Property: Everett non-decreasing in alpha, non-increasing in beta
// Test ID: EV2
func TestPreisach_EV2_EverettMonotonicity(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			ev := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.25}

			maxE := mat.Ec * 3.0
			step := mat.Ec * 0.2
			alphaViolations := 0
			betaViolations := 0
			const tolerance = 1e-12

			// Test monotonicity in alpha (for fixed beta)
			for beta := -maxE; beta <= maxE; beta += step {
				W_prev := ev.Calculate(beta, beta) // alpha = beta is minimum
				for alpha := beta + step; alpha <= maxE; alpha += step {
					W_curr := ev.Calculate(alpha, beta)
					if W_curr < W_prev-tolerance {
						alphaViolations++
					}
					W_prev = W_curr
				}
			}

			// Test monotonicity in beta (for fixed alpha)
			for alpha := -maxE; alpha <= maxE; alpha += step {
				W_prev := ev.Calculate(alpha, alpha) // beta = alpha is maximum
				for beta := alpha - step; beta >= -maxE; beta -= step {
					W_curr := ev.Calculate(alpha, beta)
					if W_curr < W_prev-tolerance {
						betaViolations++
					}
					W_prev = W_curr
				}
			}

			// Acceptance: no violations
			if alphaViolations > 0 || betaViolations > 0 {
				t.Errorf("EV2 FAILED: %d alpha violations, %d beta violations",
					alphaViolations, betaViolations)
			} else {
				t.Logf("EV2 PASSED: Everett monotonicity satisfied")
			}
		})
	}
}

// TestPreisach_EV3_EverettBoundary validates boundary conditions of Everett function (Tier T0)
// Property: Everett(a, a) ≈ 0, Everett(3*Ec, -3*Ec) ≈ Ps
// Test ID: EV3
func TestPreisach_EV3_EverettBoundary(t *testing.T) {
	materials := getTestMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			ev := &physics.TanhEverett{Ps: mat.Ps, Ec: mat.Ec, Delta: mat.Ec * 0.25}

			// Test Everett(a, a) ≈ 0 for several alpha values
			alphaValues := []float64{-mat.Ec * 2, -mat.Ec, 0, mat.Ec, mat.Ec * 2}
			maxDiagonalError := 0.0
			for _, alpha := range alphaValues {
				W := ev.Calculate(alpha, alpha)
				absError := math.Abs(W)
				if absError > maxDiagonalError {
					maxDiagonalError = absError
				}
			}

			// Acceptance: diagonal values < 1e-6 (floating-point precision limit for tanh)
			if maxDiagonalError >= 1e-6 {
				t.Errorf("EV3 FAILED: max diagonal error %.10e >= 1e-6", maxDiagonalError)
			} else {
				t.Logf("EV3 PASSED: diagonal condition satisfied (max error %.10e)", maxDiagonalError)
			}

			// Test Everett(3*Ec, -3*Ec) ≈ Ps
			W_max := ev.Calculate(mat.Ec*3, -mat.Ec*3)
			relError := math.Abs(W_max-mat.Ps) / mat.Ps

			// Acceptance: within 5% of Ps
			if relError >= 0.05 {
				t.Errorf("EV3 FAILED: |W_max - Ps| / Ps = %.4f >= 0.05", relError)
			} else {
				t.Logf("EV3 PASSED: W_max/Ps = %.6f (rel_error = %.4f < 0.05)",
					W_max/mat.Ps, relError)
			}
		})
	}
}
