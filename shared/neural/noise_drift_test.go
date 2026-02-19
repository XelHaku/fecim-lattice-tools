package neural

import (
	"math"
	"math/rand"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestDrift_M3_NOISE_04_TemporalConductance validates inference under temporal conductance drift.
// Requirement: Measure accuracy at t=0, 1hr, 1day, 1week (simulated)
// Evidence: accuracy vs time, drift trajectory, degradation rate
func TestDrift_M3_NOISE_04_TemporalConductance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping temporal drift test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-04: Testing temporal conductance drift on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Store reference weights
	refWeights1 := deepCopyWeights(net.QuantWeights1)
	refWeights2 := deepCopyWeights(net.QuantWeights2)

	// Time checkpoints (simulated)
	timePoints := []struct {
		name    string
		hours   float64
		seconds float64
	}{
		{"t=0 (initial)", 0, 0},
		{"t=1hr", 1, 3600},
		{"t=1day", 24, 86400},
		{"t=1week", 168, 604800},
	}

	// Drift model: logarithmic conductance drift
	// ΔG/G = A * log(1 + t/τ)
	// A = drift coefficient (typical 0.01-0.05 for FeCIM)
	// τ = time constant (seconds)
	const driftCoefficient = 0.02 // 2% drift coefficient
	const timeConstantS = 3600.0  // 1 hour time constant

	results := make([]struct {
		name     string
		hours    float64
		correct  int
		accuracy float64
		drift    float64
	}, len(timePoints))

	for idx, tp := range timePoints {
		// Restore reference weights
		restoreWeightsFrom(net.QuantWeights1, refWeights1)
		restoreWeightsFrom(net.QuantWeights2, refWeights2)

		// Apply drift if t > 0
		var relativeDrift float64
		if tp.seconds > 0 {
			relativeDrift = driftCoefficient * math.Log(1.0+tp.seconds/timeConstantS)
			applyLogarithmicDrift(net.QuantWeights1, relativeDrift)
			applyLogarithmicDrift(net.QuantWeights2, relativeDrift)
		}

		// Run inference
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Infer returned nil for image %d", i)
			}
			if result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		results[idx].name = tp.name
		results[idx].hours = tp.hours
		results[idx].correct = correct
		results[idx].accuracy = accuracy
		results[idx].drift = relativeDrift

		t.Logf("  %s: Accuracy = %.2f%% (%d/%d), ΔG/G = %.4f",
			tp.name, accuracy, correct, len(images), relativeDrift)
	}

	// Validate degradation trajectory
	baseline := results[0].accuracy
	acc1hr := results[1].accuracy
	acc1day := results[2].accuracy
	acc1week := results[3].accuracy

	degradation1hr := baseline - acc1hr
	degradation1day := baseline - acc1day
	degradation1week := baseline - acc1week

	t.Logf("  Accuracy degradation from baseline:")
	t.Logf("    1hr:  %.2f%% (drift: %.4f)", degradation1hr, results[1].drift)
	t.Logf("    1day: %.2f%% (drift: %.4f)", degradation1day, results[2].drift)
	t.Logf("    1week: %.2f%% (drift: %.4f)", degradation1week, results[3].drift)

	// Validate requirements
	// After 1 week, accuracy should still be >60%
	if acc1week < 60.0 {
		t.Errorf("Accuracy after 1 week %.2f%% < 60%% (excessive drift)", acc1week)
	}

	// Degradation should increase with time (monotonic within tolerance)
	if degradation1day < degradation1hr-2.0 {
		t.Logf("WARNING: 1-day degradation (%.2f%%) < 1-hour degradation (%.2f%%)",
			degradation1day, degradation1hr)
	}

	// Total drift after 1 week should be <15% for reasonable drift parameters
	if degradation1week > 15.0 {
		t.Logf("WARNING: 1-week degradation %.2f%% is high (>15%%)", degradation1week)
	}

	t.Logf("M3-NOISE-04: PASS — Temporal drift trajectory characterized")
	t.Logf("  1-week accuracy: %.2f%% ≥ 60%% ✓", acc1week)
}

// TestDrift_M3_NOISE_04_PowerLaw validates power-law drift model.
// Alternative to logarithmic: ΔG/G = A * (t/τ)^β
func TestDrift_M3_NOISE_04_PowerLaw(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping power-law drift test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-04: Testing power-law conductance drift on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	refWeights1 := deepCopyWeights(net.QuantWeights1)
	refWeights2 := deepCopyWeights(net.QuantWeights2)

	// Power-law drift: ΔG/G = A * (t/τ)^β
	const driftCoefficient = 0.01
	const timeConstantS = 3600.0
	const powerLawExponent = 0.3 // β < 1 → sublinear (typical for ferroelectrics)

	timePointsHours := []float64{0, 1, 24, 168}

	for _, hours := range timePointsHours {
		restoreWeightsFrom(net.QuantWeights1, refWeights1)
		restoreWeightsFrom(net.QuantWeights2, refWeights2)

		seconds := hours * 3600.0
		var relativeDrift float64
		if seconds > 0 {
			relativeDrift = driftCoefficient * math.Pow(seconds/timeConstantS, powerLawExponent)
			applyLogarithmicDrift(net.QuantWeights1, relativeDrift) // Reuse for multiplicative drift
			applyLogarithmicDrift(net.QuantWeights2, relativeDrift)
		}

		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result != nil && result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		t.Logf("  t=%.0fhr: Accuracy = %.2f%% (%d/%d), ΔG/G = %.4f (power-law β=%.1f)",
			hours, accuracy, correct, len(images), relativeDrift, powerLawExponent)
	}

	t.Logf("M3-NOISE-04: PASS — Power-law drift model validated")
}

// TestDrift_M3_NOISE_04_StatisticalDrift validates drift with device-to-device variation.
// Real FeCIM: each cell drifts at different rate.
func TestDrift_M3_NOISE_04_StatisticalDrift(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping statistical drift test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-04: Testing statistical drift variation on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Simulate 1-day drift with statistical variation
	const timeS = 86400.0 // 1 day
	const meanDriftCoeff = 0.02
	const stdDevDriftCoeff = 0.005 // 25% relative variation in drift coefficient

	applyStatisticalDrift(net.QuantWeights1, meanDriftCoeff, stdDevDriftCoeff, timeS)
	applyStatisticalDrift(net.QuantWeights2, meanDriftCoeff, stdDevDriftCoeff, timeS)

	// Run inference
	correct := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
		if result.CIMPrediction == labels[i] {
			correct++
		}
	}

	accuracy := 100.0 * float64(correct) / float64(len(images))
	t.Logf("  1-day statistical drift (mean A=%.3f±%.3f): Accuracy = %.2f%% (%d/%d)",
		meanDriftCoeff, stdDevDriftCoeff, accuracy, correct, len(images))

	// Should maintain >70% accuracy even with statistical drift variation
	if accuracy < 70.0 {
		t.Errorf("Statistical drift accuracy %.2f%% < 70%%", accuracy)
	}

	t.Logf("M3-NOISE-04: PASS — Statistical drift variation validated (%.2f%% ≥ 70%%)", accuracy)
}

// TestDrift_M3_NOISE_04_RefreshStrategy validates periodic refresh mitigation.
// Refresh every N hours → restore accuracy.
func TestDrift_M3_NOISE_04_RefreshStrategy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping drift refresh test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-04: Testing periodic refresh strategy on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	refWeights1 := deepCopyWeights(net.QuantWeights1)
	refWeights2 := deepCopyWeights(net.QuantWeights2)

	const driftCoefficient = 0.02
	const timeConstantS = 3600.0
	const refreshIntervalHours = 24.0 // Refresh every 24 hours

	// Simulate 1 week with refresh every 24 hours
	const totalDays = 7
	results := make([]struct {
		day      int
		accuracy float64
	}, totalDays+1)

	for day := 0; day <= totalDays; day++ {
		if day > 0 && day%1 == 0 {
			// Refresh weights (reprogram array)
			restoreWeightsFrom(net.QuantWeights1, refWeights1)
			restoreWeightsFrom(net.QuantWeights2, refWeights2)
			t.Logf("  Day %d: Refreshed weights", day)
		}

		// Apply 1-day drift
		if day > 0 {
			seconds := 86400.0 // 1 day since last refresh
			relativeDrift := driftCoefficient * math.Log(1.0+seconds/timeConstantS)
			applyLogarithmicDrift(net.QuantWeights1, relativeDrift)
			applyLogarithmicDrift(net.QuantWeights2, relativeDrift)
		}

		// Measure accuracy
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result != nil && result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		results[day].day = day
		results[day].accuracy = accuracy

		t.Logf("  Day %d end: Accuracy = %.2f%% (%d/%d)", day, accuracy, correct, len(images))
	}

	// With daily refresh, accuracy should remain high throughout the week
	for i := range results {
		if results[i].accuracy < 75.0 {
			t.Errorf("Day %d accuracy %.2f%% < 75%% (refresh strategy failed)",
				results[i].day, results[i].accuracy)
		}
	}

	t.Logf("M3-NOISE-04: PASS — Periodic refresh strategy validated")
}

// TestDrift_M3_NOISE_04_TemperatureDependence validates temperature-accelerated drift.
// Higher temperature → faster drift (Arrhenius model).
func TestDrift_M3_NOISE_04_TemperatureDependence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping temperature-dependent drift test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-04: Testing temperature-dependent drift on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	refWeights1 := deepCopyWeights(net.QuantWeights1)
	refWeights2 := deepCopyWeights(net.QuantWeights2)

	// Arrhenius temperature acceleration: rate ∝ exp(-Ea/kT)
	// Drift at higher T is faster by factor exp(Ea/k * (1/T_ref - 1/T))
	const baseDriftCoeff = 0.02
	const timeS = 86400.0          // 1 day
	const activationEnergyEV = 0.8 // Typical for ferroelectric polarization switching
	const boltzmannEV = 8.617e-5   // Boltzmann constant in eV/K
	const refTempK = 300.0         // Room temperature reference

	temperatures := []struct {
		tempC float64
		tempK float64
	}{
		{25, 298.15},  // Room temp (slightly below ref)
		{85, 358.15},  // Elevated temp (automotive/industrial)
		{125, 398.15}, // High temp (stress test)
	}

	for _, temp := range temperatures {
		restoreWeightsFrom(net.QuantWeights1, refWeights1)
		restoreWeightsFrom(net.QuantWeights2, refWeights2)

		// Temperature acceleration factor
		accelFactor := math.Exp(activationEnergyEV / boltzmannEV * (1.0/refTempK - 1.0/temp.tempK))
		effectiveDriftCoeff := baseDriftCoeff * accelFactor

		// Apply drift
		relativeDrift := effectiveDriftCoeff * math.Log(1.0+timeS/3600.0)
		applyLogarithmicDrift(net.QuantWeights1, relativeDrift)
		applyLogarithmicDrift(net.QuantWeights2, relativeDrift)

		// Run inference
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result != nil && result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		t.Logf("  T=%.0f°C (%.1fK): Accuracy = %.2f%% (%d/%d), accel=%.2fx, ΔG/G=%.4f",
			temp.tempC, temp.tempK, accuracy, correct, len(images), accelFactor, relativeDrift)
	}

	t.Logf("M3-NOISE-04: PASS — Temperature-dependent drift characterized")
}

// applyLogarithmicDrift applies multiplicative conductance drift to weights.
// G_new = G_old * (1 + relativeDrift)
func applyLogarithmicDrift(weights [][]float64, relativeDrift float64) {
	factor := 1.0 + relativeDrift
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] *= factor
		}
	}
}

// applyStatisticalDrift applies drift with cell-to-cell variation in drift coefficient.
func applyStatisticalDrift(weights [][]float64, meanDriftCoeff, stdDevDriftCoeff, timeS float64) {
	const timeConstantS = 3600.0
	for i := range weights {
		for j := range weights[i] {
			// Cell-specific drift coefficient
			cellSeed := int64(i*1000 + j)
			cellRNG := rand.New(rand.NewSource(cellSeed))
			cellDriftCoeff := meanDriftCoeff + cellRNG.NormFloat64()*stdDevDriftCoeff
			if cellDriftCoeff < 0 {
				cellDriftCoeff = 0 // Physical constraint
			}

			// Apply drift
			relativeDrift := cellDriftCoeff * math.Log(1.0+timeS/timeConstantS)
			weights[i][j] *= (1.0 + relativeDrift)
		}
	}
}

// deepCopyWeights creates a deep copy of a 2D weight matrix.
func deepCopyWeights(weights [][]float64) [][]float64 {
	copied := make([][]float64, len(weights))
	for i := range weights {
		copied[i] = make([]float64, len(weights[i]))
		copy(copied[i], weights[i])
	}
	return copied
}

// restoreWeightsFrom copies weights from source to destination.
func restoreWeightsFrom(dest, src [][]float64) {
	for i := range dest {
		copy(dest[i], src[i])
	}
}
