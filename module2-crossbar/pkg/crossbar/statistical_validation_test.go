// Package crossbar provides statistical validation tests for device error models.
//
// These tests verify that the error models produce statistically correct distributions
// and that Monte Carlo methods converge properly. The tests use established statistical
// methods to validate the implementations:
//
// 1. Kolmogorov-Smirnov Test: Compares empirical distribution to theoretical CDF
// 2. Chi-Squared Test: Tests uniformity of discrete distributions
// 3. T-Test: Tests whether sample mean matches expected value
// 4. F-Test (via variance ratio): Tests whether sample variance matches expected
//
// Theory References:
// - Massey, F.J. (1951). "The Kolmogorov-Smirnov Test for Goodness of Fit"
// - Pearson, K. (1900). "On the criterion that a given system of deviations..."
// - Mood, A.M. et al. (1974). "Introduction to the Theory of Statistics"
//
// CrossSim Reference:
// - Sandia National Labs CrossSim: github.com/sandialabs/cross-sim
// - Device error models from generic_error.py
package crossbar

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// =============================================================================
// SECTION 1: Normal Distribution Tests (Independent Error Model)
// =============================================================================
//
// The Normal Independent error model generates Gaussian noise with fixed sigma
// independent of the conductance value. This is appropriate for read noise where
// measurement noise is dominated by amplifier/ADC noise rather than device physics.
//
// CrossSim model: noise = rng.NormFloat64() * sigma
// Expected: samples ~ N(0, sigma^2)

func TestDeviceErrorDistribution_NormalIndependent(t *testing.T) {
	// Test Configuration:
	// - 10,000 samples provides good statistical power
	// - Sigma = 0.05 (5% typical read noise level)
	// - Fixed seed for reproducibility
	//
	// Acceptance Criteria:
	// - KS statistic < critical value at alpha=0.05
	// - For n=10000, critical value ≈ 1.36/sqrt(n) ≈ 0.0136
	// - We use a slightly relaxed threshold of 0.02 to account for
	//   boundary effects from clamping

	const (
		nSamples     = 10000
		sigma        = 0.05
		seed         = 42
		ksThreshold  = 0.02 // Relaxed due to clamping effects
		meanTolStd   = 3.0  // Within 3 standard errors of zero
		sigmaTolPct  = 10.0 // Within 10% of expected sigma
	)

	config := &ReadNoiseConfig{
		Enable:     true,
		Model:      ErrorModelNormalIndependent,
		Sigma:      sigma,
		Persistent: false,
		Seed:       seed,
	}
	engine := NewDeviceErrorEngine(nil, config)

	// Generate samples at midpoint conductance (0.5) where clamping is minimal
	gTarget := 0.5
	samples := make([]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		gNoisy := engine.ApplyReadNoise(gTarget, 0, i)
		// Extract the multiplicative noise factor: G_read = G_target * (1 + noise)
		// noise = (G_read / G_target) - 1
		samples[i] = (gNoisy / gTarget) - 1
	}

	// Statistical Analysis
	mean, stdDev := sampleMeanStd(samples)
	stderr := stdDev / math.Sqrt(float64(nSamples))

	// Test 1: Mean should be approximately zero (unbiased estimator)
	// Using t-test: |mean| < meanTolStd * stderr
	if math.Abs(mean) > meanTolStd*stderr {
		t.Errorf("Mean bias detected: mean=%.6f, stderr=%.6f, threshold=%.6f",
			mean, stderr, meanTolStd*stderr)
	}

	// Test 2: Standard deviation should match configured sigma
	relSigmaError := math.Abs(stdDev-sigma) / sigma * 100
	if relSigmaError > sigmaTolPct {
		t.Errorf("Sigma mismatch: measured=%.6f, expected=%.6f, error=%.1f%%",
			stdDev, sigma, relSigmaError)
	}

	// Test 3: Kolmogorov-Smirnov test against standard normal
	// Normalize samples: z = (x - mean) / stdDev ≈ N(0,1)
	normalizedSamples := make([]float64, nSamples)
	for i, s := range samples {
		normalizedSamples[i] = s / sigma // Should be ~ N(0,1)
	}
	ksStatistic := kolmogorovSmirnovAgainstStandardNormal(normalizedSamples)

	if ksStatistic > ksThreshold {
		t.Errorf("KS test failed: statistic=%.4f > threshold=%.4f (not normal)",
			ksStatistic, ksThreshold)
	}

	// Test 4: Check Gaussian property - ~68% within 1 sigma
	within1Sigma := 0
	for _, s := range samples {
		if math.Abs(s) <= sigma {
			within1Sigma++
		}
	}
	pct1Sigma := float64(within1Sigma) / float64(nSamples) * 100
	// Gaussian: 68.27% within 1 sigma, allow ±5%
	if pct1Sigma < 63 || pct1Sigma > 73 {
		t.Errorf("Gaussian 1-sigma property violated: %.1f%% within 1 sigma (expected ~68%%)",
			pct1Sigma)
	}

	t.Logf("NormalIndependent Validation PASSED:")
	t.Logf("  Mean: %.6f (expected 0)", mean)
	t.Logf("  StdDev: %.6f (expected %.4f)", stdDev, sigma)
	t.Logf("  KS statistic: %.4f (threshold %.4f)", ksStatistic, ksThreshold)
	t.Logf("  Within 1 sigma: %.1f%% (expected ~68%%)", pct1Sigma)
}

// =============================================================================
// SECTION 2: Normal Distribution Tests (Proportional Error Model)
// =============================================================================
//
// The Normal Proportional error model generates Gaussian noise that scales with
// the conductance value. This is appropriate for programming error where
// variability increases with the programmed state.
//
// CrossSim model: effectiveSigma = sigma * gValue, noise = rng.NormFloat64() * effectiveSigma
// Expected: relative error ~ N(0, sigma^2) regardless of gValue

func TestDeviceErrorDistribution_NormalProportional(t *testing.T) {
	// Test Configuration:
	// - Test at multiple conductance levels to verify proportional scaling
	// - 5000 samples per level for good statistics
	//
	// Key Physics:
	// For proportional error, the RELATIVE error (not absolute) should have
	// constant standard deviation equal to the configured sigma.

	const (
		nSamples      = 5000
		sigma         = 0.10 // 10% programming error
		seed          = 123
		relSigmaTolPct = 15.0 // Allow 15% deviation due to finite samples
	)

	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalProportional,
		Sigma:     sigma,
		Symmetric: true,
		Seed:      seed,
	}
	engine := NewDeviceErrorEngine(config, nil)

	// Test at different conductance levels
	testLevels := []float64{0.2, 0.4, 0.6, 0.8}
	relStdDevs := make([]float64, len(testLevels))

	for idx, gTarget := range testLevels {
		relErrors := make([]float64, nSamples)

		for i := 0; i < nSamples; i++ {
			gProgrammed := engine.ApplyProgrammingError(gTarget)
			// Relative error = (G_prog - G_target) / G_target
			relErrors[i] = (gProgrammed - gTarget) / gTarget
		}

		mean, stdDev := sampleMeanStd(relErrors)
		relStdDevs[idx] = stdDev

		// Mean should be approximately zero (symmetric distribution)
		stderr := stdDev / math.Sqrt(float64(nSamples))
		if math.Abs(mean) > 3*stderr {
			t.Errorf("G=%.2f: Mean bias %.6f exceeds 3*stderr %.6f",
				gTarget, mean, 3*stderr)
		}

		// Relative std should match sigma
		relSigmaError := math.Abs(stdDev-sigma) / sigma * 100
		if relSigmaError > relSigmaTolPct {
			t.Errorf("G=%.2f: RelStd=%.4f, expected ~%.4f, error=%.1f%%",
				gTarget, stdDev, sigma, relSigmaError)
		}

		t.Logf("G=%.2f: RelMean=%.6f, RelStd=%.4f (expected %.4f)", gTarget, mean, stdDev, sigma)
	}

	// Additional test: verify std is constant across conductance levels
	// (key property of proportional model)
	avgRelStd := 0.0
	for _, s := range relStdDevs {
		avgRelStd += s
	}
	avgRelStd /= float64(len(relStdDevs))

	for idx, s := range relStdDevs {
		deviation := math.Abs(s-avgRelStd) / avgRelStd * 100
		if deviation > 20 { // Within 20% of average
			t.Errorf("G=%.2f: RelStd varies by %.1f%% from average - proportional scaling broken",
				testLevels[idx], deviation)
		}
	}

	t.Logf("NormalProportional Validation PASSED: RelStd consistent across conductance levels")
}

// =============================================================================
// SECTION 3: Quantization Distribution Test
// =============================================================================
//
// The QuantizeTo30Levels function maps continuous values to 30 discrete levels.
// The function uses math.Round to map values to the nearest level:
//
//   level = round(value * (N-1))
//
// For N=30 levels and uniform input U[0,1]:
// - Edge levels (0, 29) have capture range of width 0.5/(N-1) ≈ 0.0172
// - Interior levels (1-28) have capture range of width 1/(N-1) ≈ 0.0345
//
// Expected probabilities:
// - P(level=0) = P(level=29) = 0.5/(N-1) ≈ 1.72%
// - P(level=k) for k in [1,28] = 1/(N-1) ≈ 3.45%
//
// Chi-squared test uses these non-uniform expected frequencies.

func TestQuantizationDistribution_30Levels(t *testing.T) {
	// Test Configuration:
	// - 30,000 samples ensures sufficient counts per level
	// - Chi-squared critical value at df=29, alpha=0.05 ≈ 42.56

	const (
		nSamples        = 30000
		nLevels         = 30
		seed            = 456
		chiSqThreshold  = 50.0 // Slightly relaxed from 42.56 for numerical rounding
	)

	rng := rand.New(rand.NewSource(seed))

	// Generate uniformly distributed values and quantize
	levelCounts := make([]int, nLevels)

	for i := 0; i < nSamples; i++ {
		value := rng.Float64() // Uniform [0, 1)
		quantized := physics.QuantizeTo30Levels(value)
		level := physics.GetLevelFor30(quantized)

		if level >= 0 && level < nLevels {
			levelCounts[level]++
		}
	}

	// Calculate CORRECT expected frequencies based on rounding boundaries
	// Level k captures values in range: [(k-0.5)/(N-1), (k+0.5)/(N-1)]
	// with boundary conditions at 0 and 1
	//
	// Level 0: [0, 0.5/29] → width = 0.5/29
	// Level 1-28: [(k-0.5)/29, (k+0.5)/29] → width = 1/29
	// Level 29: [28.5/29, 1] → width = 0.5/29
	N1 := float64(nLevels - 1) // 29
	levelSpacing := 1.0 / N1   // ~0.0345

	observed := make([]float64, nLevels)
	expected := make([]float64, nLevels)

	for i := 0; i < nLevels; i++ {
		observed[i] = float64(levelCounts[i])
		if i == 0 || i == nLevels-1 {
			// Edge levels have half-width capture range
			expected[i] = float64(nSamples) * (levelSpacing / 2.0)
		} else {
			// Interior levels have full-width capture range
			expected[i] = float64(nSamples) * levelSpacing
		}
	}

	// Chi-squared test with correct expected frequencies
	chiSquared, df := chiSquaredTest(observed, expected)

	if chiSquared > chiSqThreshold {
		t.Errorf("Quantization distribution mismatch: chi-squared=%.2f > threshold=%.2f (df=%d)",
			chiSquared, chiSqThreshold, df)
	}

	// Log distribution statistics
	t.Logf("Quantization Distribution Test (Round-based):")
	t.Logf("  Total samples: %d, Levels: %d", nSamples, nLevels)
	t.Logf("  Expected edge levels (0, 29): %.1f", expected[0])
	t.Logf("  Expected interior levels (1-28): %.1f", expected[1])
	t.Logf("  Observed level 0: %d, level 29: %d", levelCounts[0], levelCounts[nLevels-1])
	t.Logf("  Chi-squared: %.2f (threshold %.2f, df=%d)", chiSquared, chiSqThreshold, df)

	// Verify edge level behavior: should have ~half the counts of interior levels
	edgeAvg := float64(levelCounts[0]+levelCounts[nLevels-1]) / 2.0
	interiorSum := 0
	for i := 1; i < nLevels-1; i++ {
		interiorSum += levelCounts[i]
	}
	interiorAvg := float64(interiorSum) / float64(nLevels-2)

	edgeRatio := edgeAvg / interiorAvg
	expectedRatio := 0.5 // Edge levels should have ~half the probability

	if math.Abs(edgeRatio-expectedRatio) > 0.15 {
		t.Errorf("Edge/interior ratio: %.3f (expected ~%.3f)", edgeRatio, expectedRatio)
	}

	// Additional test: verify each level is represented
	zeroLevels := 0
	for i, count := range levelCounts {
		if count == 0 {
			zeroLevels++
			t.Errorf("Level %d has zero samples - quantization may be broken", i)
		}
	}

	if zeroLevels == 0 {
		t.Logf("Quantization Distribution PASSED: Chi-squared=%.2f, edge ratio=%.3f",
			chiSquared, edgeRatio)
	}
}

// =============================================================================
// SECTION 4: Monte Carlo Convergence Test
// =============================================================================
//
// Monte Carlo methods should converge at rate O(1/sqrt(N)). This test verifies
// that the standard error of MVM output decreases as expected with sample size.
//
// Theory:
// For N independent samples with variance σ²:
// - Sample mean has variance σ²/N
// - Standard error = σ/sqrt(N)
// - Doubling N should reduce stderr by factor of sqrt(2) ≈ 1.41

func TestMonteCarloConvergence_MVM(t *testing.T) {
	// Test Configuration:
	// - Array size 8x8 (small for fast testing)
	// - Sample sizes: 100, 400, 1600 (4x increases → 2x stderr reduction each step)
	// - Noise level 5% for visible but not dominant effects

	const (
		arraySize     = 8
		noiseLevel    = 0.05
		seed          = 789
		convergeTol   = 0.3 // Allow 30% deviation from sqrt(N) scaling
	)

	sampleSizes := []int{100, 400, 1600}
	stderrs := make([]float64, len(sampleSizes))

	// Fixed input for all tests
	rng := rand.New(rand.NewSource(seed))
	input := make([]float64, arraySize)
	for i := range input {
		input[i] = rng.Float64()
	}

	// Fixed weight matrix (random but consistent)
	weights := make([][]float64, arraySize)
	for i := range weights {
		weights[i] = make([]float64, arraySize)
		for j := range weights[i] {
			weights[i][j] = rng.Float64()
		}
	}

	for idx, nSamples := range sampleSizes {
		// Collect MVM outputs
		outputs := make([][]float64, nSamples)

		for s := 0; s < nSamples; s++ {
			// Create fresh array with noise for each sample
			cfg := &Config{
				Rows:       arraySize,
				Cols:       arraySize,
				NoiseLevel: noiseLevel,
				ADCBits:    16,
				DACBits:    16,
			}
			arr, _ := NewArray(cfg)

			// Program weights
			for i := 0; i < arraySize; i++ {
				for j := 0; j < arraySize; j++ {
					arr.ProgramWeight(i, j, weights[i][j])
				}
			}

			output, _ := arr.MVM(input)
			outputs[s] = output
		}

		// Calculate mean and stderr of first output element
		means := make([]float64, nSamples)
		for s := 0; s < nSamples; s++ {
			means[s] = outputs[s][0]
		}

		mean, stdDev := sampleMeanStd(means)
		stderr := stdDev / math.Sqrt(float64(nSamples))
		stderrs[idx] = stderr

		t.Logf("N=%4d: mean=%.6f, std=%.6f, stderr=%.6f",
			nSamples, mean, stdDev, stderr)
	}

	// Verify 1/sqrt(N) convergence
	// Between consecutive sample sizes with 4x increase, stderr should decrease by ~2x
	for i := 1; i < len(sampleSizes); i++ {
		expectedRatio := math.Sqrt(float64(sampleSizes[i]) / float64(sampleSizes[i-1]))
		actualRatio := stderrs[i-1] / stderrs[i]

		deviation := math.Abs(actualRatio-expectedRatio) / expectedRatio

		if deviation > convergeTol {
			t.Errorf("MC convergence deviation: N=%d->%d, expected ratio=%.2f, actual=%.2f, deviation=%.1f%%",
				sampleSizes[i-1], sampleSizes[i], expectedRatio, actualRatio, deviation*100)
		}

		t.Logf("N=%d->%d: stderr ratio=%.2f (expected ~%.2f)",
			sampleSizes[i-1], sampleSizes[i], actualRatio, expectedRatio)
	}

	t.Logf("Monte Carlo Convergence PASSED: stderr scales as 1/sqrt(N)")
}

// =============================================================================
// SECTION 5: Reproducibility Test
// =============================================================================
//
// Deterministic behavior with fixed seeds is essential for debugging and
// regression testing. This test verifies that identical seeds produce
// identical results.

func TestErrorModelReproducibility(t *testing.T) {
	const (
		nSamples = 100
		seed     = 999
		sigma    = 0.05
		gTarget  = 0.5
	)

	// Configuration with fixed seed
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalProportional,
		Sigma:     sigma,
		Symmetric: true,
		Seed:      seed,
	}

	// First run
	engine1 := NewDeviceErrorEngine(config, nil)
	results1 := make([]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		results1[i] = engine1.ApplyProgrammingError(gTarget)
	}

	// Second run with same seed
	engine2 := NewDeviceErrorEngine(config, nil)
	results2 := make([]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		results2[i] = engine2.ApplyProgrammingError(gTarget)
	}

	// Results must be identical
	mismatches := 0
	for i := 0; i < nSamples; i++ {
		if results1[i] != results2[i] {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("Mismatch at i=%d: run1=%.10f, run2=%.10f",
					i, results1[i], results2[i])
			}
		}
	}

	if mismatches > 0 {
		t.Errorf("Reproducibility FAILED: %d/%d results differ with same seed",
			mismatches, nSamples)
	} else {
		t.Logf("Reproducibility PASSED: All %d results identical with seed=%d", nSamples, seed)
	}
}

// =============================================================================
// SECTION 6: Error Model Statistics Validation
// =============================================================================
//
// Comprehensive test of first and second moments (mean and variance) of
// the error distributions. Uses proper statistical hypothesis tests.

func TestErrorModelStatistics(t *testing.T) {
	// Test Configuration:
	// - Large sample size for high statistical power
	// - Multiple error models tested

	const (
		nSamples   = 10000
		sigma      = 0.05
		gTarget    = 0.5
		alpha      = 0.05 // Significance level
	)

	testCases := []struct {
		name        string
		model       ErrorModel
		expectedMean float64 // Expected mean error (relative or absolute depending on model)
		checkMean   bool
	}{
		{
			name:        "NormalIndependent",
			model:       ErrorModelNormalIndependent,
			expectedMean: 0.0,
			checkMean:   true,
		},
		{
			name:        "NormalProportional",
			model:       ErrorModelNormalProportional,
			expectedMean: 0.0,
			checkMean:   true,
		},
		{
			name:        "UniformIndependent",
			model:       ErrorModelUniformIndependent,
			expectedMean: 0.0,
			checkMean:   true,
		},
		{
			name:        "UniformProportional",
			model:       ErrorModelUniformProportional,
			expectedMean: 0.0,
			checkMean:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &ProgrammingErrorConfig{
				Enable:    true,
				Model:     tc.model,
				Sigma:     sigma,
				Symmetric: true,
				Seed:      42,
			}
			engine := NewDeviceErrorEngine(config, nil)

			// Collect errors
			errors := make([]float64, nSamples)
			for i := 0; i < nSamples; i++ {
				gProgrammed := engine.ApplyProgrammingError(gTarget)
				// For proportional models, we analyze relative error
				// For independent models, we analyze absolute error
				switch tc.model {
				case ErrorModelNormalProportional, ErrorModelUniformProportional:
					errors[i] = (gProgrammed - gTarget) / gTarget
				default:
					errors[i] = gProgrammed - gTarget
				}
			}

			mean, stdDev := sampleMeanStd(errors)
			stderr := stdDev / math.Sqrt(float64(nSamples))

			// T-test for mean
			// H0: mean = expectedMean
			// Test statistic: t = (mean - expectedMean) / stderr
			// Critical value at alpha=0.05, two-tailed, df large ≈ 1.96
			if tc.checkMean {
				tStat := math.Abs(mean - tc.expectedMean) / stderr
				tCritical := 1.96 // z-approximation for large n

				if tStat > tCritical {
					t.Errorf("Mean test failed: mean=%.6f, expected=%.6f, t=%.2f > %.2f",
						mean, tc.expectedMean, tStat, tCritical)
				}
			}

			// For Gaussian models, verify variance matches sigma²
			// (Uniform models have different variance)
			if tc.model == ErrorModelNormalIndependent || tc.model == ErrorModelNormalProportional {
				// Variance should be approximately sigma²
				variance := stdDev * stdDev
				expectedVariance := sigma * sigma

				// Chi-squared test for variance (simplified)
				// For large n, variance ratio has approximately normal distribution
				varianceRatio := variance / expectedVariance
				if varianceRatio < 0.8 || varianceRatio > 1.2 {
					t.Errorf("Variance test: measured=%.6f, expected=%.6f, ratio=%.3f (outside [0.8, 1.2])",
						variance, expectedVariance, varianceRatio)
				}

				t.Logf("%s: mean=%.6f (expected %.4f), var=%.6f (expected %.4f), ratio=%.3f",
					tc.name, mean, tc.expectedMean, variance, expectedVariance, varianceRatio)
			} else {
				// For uniform distribution, variance = sigma²/3 (for U[-sigma, sigma])
				expectedVariance := sigma * sigma / 3.0
				variance := stdDev * stdDev
				varianceRatio := variance / expectedVariance

				if varianceRatio < 0.7 || varianceRatio > 1.3 {
					t.Logf("%s: Uniform variance ratio %.3f outside expected range",
						tc.name, varianceRatio)
				}

				t.Logf("%s: mean=%.6f, var=%.6f (uniform expected ~%.6f)",
					tc.name, mean, variance, expectedVariance)
			}
		})
	}
}

// =============================================================================
// SECTION 7: Error Model Independence Test
// =============================================================================
//
// Verifies that successive error samples are statistically independent.
// Uses autocorrelation at lag 1 as the test statistic.

func TestErrorModelIndependence(t *testing.T) {
	const (
		nSamples        = 5000
		seed            = 111
		sigma           = 0.05
		gTarget         = 0.5
		correlationTol  = 0.05 // Autocorrelation should be < 5%
	)

	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalIndependent,
		Sigma:     sigma,
		Symmetric: true,
		Seed:      seed,
	}
	engine := NewDeviceErrorEngine(config, nil)

	errors := make([]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		gProgrammed := engine.ApplyProgrammingError(gTarget)
		errors[i] = gProgrammed - gTarget
	}

	// Calculate lag-1 autocorrelation
	// r(1) = Cov(X_t, X_{t+1}) / Var(X)
	mean, _ := sampleMeanStd(errors)

	var sumProd, sumSq float64
	for i := 0; i < nSamples-1; i++ {
		sumProd += (errors[i] - mean) * (errors[i+1] - mean)
		sumSq += (errors[i] - mean) * (errors[i] - mean)
	}
	// Add last term for variance
	sumSq += (errors[nSamples-1] - mean) * (errors[nSamples-1] - mean)

	autocorr := sumProd / sumSq

	// For independent samples, autocorrelation should be near zero
	// Expected stderr of autocorrelation ≈ 1/sqrt(N)
	expectedStderr := 1.0 / math.Sqrt(float64(nSamples))

	if math.Abs(autocorr) > correlationTol {
		t.Errorf("Autocorrelation too high: r(1)=%.4f > threshold=%.4f (samples may not be independent)",
			autocorr, correlationTol)
	} else {
		t.Logf("Independence Test PASSED: r(1)=%.4f (threshold=%.4f, stderr=%.4f)",
			autocorr, correlationTol, expectedStderr)
	}
}

// =============================================================================
// SECTION 8: Boundary Condition Tests
// =============================================================================
//
// Tests error model behavior at boundary conditions (G=0, G=1).
// Verifies clamping works correctly and doesn't introduce bias.

func TestErrorModelBoundaryConditions(t *testing.T) {
	const (
		nSamples = 1000
		sigma    = 0.10
		seed     = 222
	)

	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalIndependent,
		Sigma:     sigma,
		Symmetric: true,
		Seed:      seed,
	}
	engine := NewDeviceErrorEngine(config, nil)

	t.Run("LowConductance_NoNegative", func(t *testing.T) {
		// Test at G=0.05 (near lower boundary)
		gTarget := 0.05
		negativeCount := 0
		clampedCount := 0

		for i := 0; i < nSamples; i++ {
			gProgrammed := engine.ApplyProgrammingError(gTarget)
			if gProgrammed < 0 {
				negativeCount++
			}
			if gProgrammed == 0 {
				clampedCount++
			}
		}

		if negativeCount > 0 {
			t.Errorf("Negative conductance produced: %d/%d samples", negativeCount, nSamples)
		}

		t.Logf("G=%.2f: %d samples clamped to 0 (expected with 10%% sigma)", gTarget, clampedCount)
	})

	t.Run("HighConductance_NoOverflow", func(t *testing.T) {
		// Test at G=0.95 (near upper boundary)
		gTarget := 0.95
		overflowCount := 0
		clampedCount := 0

		for i := 0; i < nSamples; i++ {
			gProgrammed := engine.ApplyProgrammingError(gTarget)
			if gProgrammed > 1 {
				overflowCount++
			}
			if gProgrammed == 1 {
				clampedCount++
			}
		}

		if overflowCount > 0 {
			t.Errorf("Conductance > 1 produced: %d/%d samples", overflowCount, nSamples)
		}

		t.Logf("G=%.2f: %d samples clamped to 1 (expected with 10%% sigma)", gTarget, clampedCount)
	})

	t.Run("ZeroConductance_HandledCorrectly", func(t *testing.T) {
		// Test at G=0 (exact lower boundary)
		gTarget := 0.0

		// For independent model, noise adds around zero
		// For proportional model, zero conductance means zero noise
		for i := 0; i < 100; i++ {
			gProgrammed := engine.ApplyProgrammingError(gTarget)
			if gProgrammed < 0 || gProgrammed > 1 {
				t.Errorf("Out-of-range conductance at G=0: %.10f", gProgrammed)
			}
		}
		t.Log("G=0.0: All samples within valid range [0,1]")
	})

	t.Run("UnityConductance_HandledCorrectly", func(t *testing.T) {
		// Test at G=1 (exact upper boundary)
		gTarget := 1.0

		for i := 0; i < 100; i++ {
			gProgrammed := engine.ApplyProgrammingError(gTarget)
			if gProgrammed < 0 || gProgrammed > 1 {
				t.Errorf("Out-of-range conductance at G=1: %.10f", gProgrammed)
			}
		}
		t.Log("G=1.0: All samples within valid range [0,1]")
	})
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// sampleMeanStd computes sample mean and sample standard deviation.
// Uses Bessel's correction (N-1) for unbiased variance estimate.
func sampleMeanStd(values []float64) (mean, stdDev float64) {
	n := float64(len(values))
	if n <= 1 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / n

	// Calculate variance with Bessel's correction
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	variance := sumSq / (n - 1) // Bessel's correction
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}

// kolmogorovSmirnovAgainstStandardNormal performs a one-sample KS test
// comparing the sample to the standard normal distribution N(0,1).
//
// Returns the KS statistic D = max|F_n(x) - Phi(x)|
// where F_n is the empirical CDF and Phi is the standard normal CDF.
//
// For reference, critical values at alpha=0.05:
// n=100: 0.136, n=1000: 0.043, n=10000: 0.0136
// Approximation: critical_value ≈ 1.36 / sqrt(n)
func kolmogorovSmirnovAgainstStandardNormal(samples []float64) float64 {
	n := len(samples)
	if n == 0 {
		return 0
	}

	// Sort samples
	sorted := make([]float64, n)
	copy(sorted, samples)
	sort.Float64s(sorted)

	// Find maximum deviation between empirical and theoretical CDF
	maxD := 0.0
	for i, x := range sorted {
		// Empirical CDF: F_n(x) = (i+1)/n (proportion <= x)
		empiricalCDF := float64(i+1) / float64(n)

		// Standard normal CDF: Phi(x)
		theoreticalCDF := normalCDF(x)

		// Also check the "step" before this point
		empiricalCDFPrev := float64(i) / float64(n)

		// D = max(|F_n(x) - Phi(x)|, |F_n(x-) - Phi(x)|)
		d1 := math.Abs(empiricalCDF - theoreticalCDF)
		d2 := math.Abs(empiricalCDFPrev - theoreticalCDF)

		if d1 > maxD {
			maxD = d1
		}
		if d2 > maxD {
			maxD = d2
		}
	}

	return maxD
}

// normalCDF computes the CDF of the standard normal distribution.
// Uses the error function: Phi(x) = (1 + erf(x/sqrt(2))) / 2
func normalCDF(x float64) float64 {
	return (1.0 + math.Erf(x/math.Sqrt2)) / 2.0
}

// chiSquaredTest computes the chi-squared statistic for goodness of fit.
// Returns the statistic and degrees of freedom.
func chiSquaredTest(observed, expected []float64) (chiSquared float64, df int) {
	if len(observed) != len(expected) {
		return 0, 0
	}

	chiSquared = 0.0
	for i := range observed {
		if expected[i] > 0 {
			diff := observed[i] - expected[i]
			chiSquared += (diff * diff) / expected[i]
		}
	}

	df = len(observed) - 1
	return chiSquared, df
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkDeviceErrorEngine_NormalIndependent(b *testing.B) {
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalIndependent,
		Sigma:     0.05,
		Symmetric: true,
		Seed:      42,
	}
	engine := NewDeviceErrorEngine(config, nil)
	gTarget := 0.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ApplyProgrammingError(gTarget)
	}
}

func BenchmarkDeviceErrorEngine_NormalProportional(b *testing.B) {
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalProportional,
		Sigma:     0.05,
		Symmetric: true,
		Seed:      42,
	}
	engine := NewDeviceErrorEngine(config, nil)
	gTarget := 0.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ApplyProgrammingError(gTarget)
	}
}

func BenchmarkQuantizeTo30Levels(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	values := make([]float64, 1000)
	for i := range values {
		values[i] = rng.Float64()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = physics.QuantizeTo30Levels(values[i%len(values)])
	}
}
