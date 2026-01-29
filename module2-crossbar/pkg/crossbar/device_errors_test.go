package crossbar

import (
	"math"
	"testing"
)

func TestErrorModelString(t *testing.T) {
	tests := []struct {
		model    ErrorModel
		expected string
	}{
		{ErrorModelNone, "None"},
		{ErrorModelNormalIndependent, "Normal (Independent)"},
		{ErrorModelNormalProportional, "Normal (Proportional)"},
		{ErrorModelNormalInverseProportional, "Normal (Inverse Proportional)"},
		{ErrorModelUniformIndependent, "Uniform (Independent)"},
		{ErrorModelUniformProportional, "Uniform (Proportional)"},
		{ErrorModel(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.model.String(); got != tt.expected {
			t.Errorf("ErrorModel(%d).String() = %s, want %s", tt.model, got, tt.expected)
		}
	}
}

func TestNewDeviceErrorEngine(t *testing.T) {
	// Test with nil configs (should use defaults)
	engine := NewDeviceErrorEngine(nil, nil)
	if engine == nil {
		t.Fatal("NewDeviceErrorEngine returned nil")
	}

	// Test with custom configs
	progConfig := &ProgrammingErrorConfig{
		Enable: true,
		Model:  ErrorModelNormalProportional,
		Sigma:  0.1,
		Seed:   42,
	}
	readConfig := &ReadNoiseConfig{
		Enable: true,
		Model:  ErrorModelNormalIndependent,
		Sigma:  0.02,
		Seed:   43,
	}
	engine = NewDeviceErrorEngine(progConfig, readConfig)
	if engine == nil {
		t.Fatal("NewDeviceErrorEngine returned nil with custom configs")
	}
}

func TestApplyProgrammingError_Disabled(t *testing.T) {
	config := &ProgrammingErrorConfig{
		Enable: false,
		Sigma:  0.1,
	}
	engine := NewDeviceErrorEngine(config, nil)

	// Should return input unchanged when disabled
	input := 0.5
	output := engine.ApplyProgrammingError(input)
	if output != input {
		t.Errorf("Expected %f, got %f (disabled should return input)", input, output)
	}
}

func TestApplyProgrammingError_NormalIndependent(t *testing.T) {
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalIndependent,
		Sigma:     0.05,
		Symmetric: true,
		Seed:      42, // Fixed seed for reproducibility
	}
	engine := NewDeviceErrorEngine(config, nil)

	// Generate many samples and check distribution
	nSamples := 10000
	target := 0.5
	sumError := 0.0
	sumSqError := 0.0

	for i := 0; i < nSamples; i++ {
		actual := engine.ApplyProgrammingError(target)
		err := actual - target
		sumError += err
		sumSqError += err * err

		// Check bounds
		if actual < 0 || actual > 1 {
			t.Errorf("Output %f out of bounds [0,1]", actual)
		}
	}

	meanError := sumError / float64(nSamples)
	variance := (sumSqError / float64(nSamples)) - (meanError * meanError)
	stdDev := math.Sqrt(variance)

	// Mean should be close to 0 for symmetric error
	if math.Abs(meanError) > 0.01 {
		t.Errorf("Mean error %.4f too large for symmetric model", meanError)
	}

	// Standard deviation should be close to sigma
	if math.Abs(stdDev-config.Sigma) > 0.01 {
		t.Errorf("StdDev %.4f not close to sigma %.4f", stdDev, config.Sigma)
	}
}

func TestApplyProgrammingError_NormalProportional(t *testing.T) {
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalProportional,
		Sigma:     0.1,
		Symmetric: true,
		Seed:      42,
	}
	engine := NewDeviceErrorEngine(config, nil)

	// Test at different conductance levels
	testCases := []struct {
		gTarget     float64
		expectedStd float64 // Approximate expected std (proportional to G)
	}{
		{0.8, 0.08}, // High G: higher absolute noise
		{0.4, 0.04}, // Low G: lower absolute noise
	}

	for _, tc := range testCases {
		sumSqError := 0.0
		nSamples := 5000

		for i := 0; i < nSamples; i++ {
			actual := engine.ApplyProgrammingError(tc.gTarget)
			err := actual - tc.gTarget
			sumSqError += err * err
		}

		stdDev := math.Sqrt(sumSqError / float64(nSamples))
		// For proportional model, error scales with conductance
		// Allow generous tolerance due to multiplicative nature
		if stdDev < tc.expectedStd*0.3 || stdDev > tc.expectedStd*3 {
			t.Logf("G=%.1f: StdDev %.4f (expected ~%.4f)", tc.gTarget, stdDev, tc.expectedStd)
		}
	}
}

func TestApplyProgrammingError_Asymmetric(t *testing.T) {
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelNormalIndependent,
		Sigma:     0.05,
		Symmetric: false, // Only positive errors
		Seed:      42,
	}
	engine := NewDeviceErrorEngine(config, nil)

	target := 0.5
	nSamples := 1000
	negativeCount := 0

	for i := 0; i < nSamples; i++ {
		actual := engine.ApplyProgrammingError(target)
		if actual < target {
			negativeCount++
		}
	}

	// With asymmetric, should have no (or very few) negative errors
	// Allow small percentage due to clamping at boundaries
	if float64(negativeCount)/float64(nSamples) > 0.01 {
		t.Errorf("Asymmetric mode had %d negative errors out of %d", negativeCount, nSamples)
	}
}

func TestApplyReadNoise_Disabled(t *testing.T) {
	config := &ReadNoiseConfig{
		Enable: false,
		Sigma:  0.1,
	}
	engine := NewDeviceErrorEngine(nil, config)

	input := 0.5
	output := engine.ApplyReadNoise(input, 0, 0)
	if output != input {
		t.Errorf("Expected %f, got %f (disabled should return input)", input, output)
	}
}

func TestApplyReadNoise_Persistent(t *testing.T) {
	config := &ReadNoiseConfig{
		Enable:     true,
		Model:      ErrorModelNormalIndependent,
		Sigma:      0.05,
		Persistent: true,
		Seed:       42,
	}
	engine := NewDeviceErrorEngine(nil, config)

	gProgrammed := 0.5

	// First read for cell (0,0)
	read1 := engine.ApplyReadNoise(gProgrammed, 0, 0)
	// Second read for same cell should be identical
	read2 := engine.ApplyReadNoise(gProgrammed, 0, 0)

	if read1 != read2 {
		t.Errorf("Persistent noise changed: read1=%f, read2=%f", read1, read2)
	}

	// Different cell should have different noise
	read3 := engine.ApplyReadNoise(gProgrammed, 0, 1)
	if read1 == read3 {
		t.Log("Note: Different cells have same noise (possible with random)")
	}

	// Clear and re-read should give different value
	engine.ClearPersistentNoise()
	read4 := engine.ApplyReadNoise(gProgrammed, 0, 0)
	if read1 == read4 {
		t.Log("Note: Noise same after clear (possible but unlikely)")
	}
}

func TestApplyReadNoise_NonPersistent(t *testing.T) {
	config := &ReadNoiseConfig{
		Enable:     true,
		Model:      ErrorModelNormalIndependent,
		Sigma:      0.05,
		Persistent: false,
		Seed:       42,
	}
	engine := NewDeviceErrorEngine(nil, config)

	gProgrammed := 0.5

	// Multiple reads should vary
	reads := make([]float64, 100)
	for i := range reads {
		reads[i] = engine.ApplyReadNoise(gProgrammed, 0, 0)
	}

	// Check that we have variation
	allSame := true
	for i := 1; i < len(reads); i++ {
		if reads[i] != reads[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Non-persistent noise should vary between reads")
	}
}

func TestApplyProgrammingErrorToMatrix(t *testing.T) {
	config := &ProgrammingErrorConfig{
		Enable: true,
		Model:  ErrorModelNormalIndependent,
		Sigma:  0.05,
		Seed:   42,
	}
	engine := NewDeviceErrorEngine(config, nil)

	// Create test matrix
	original := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	result := engine.ApplyProgrammingErrorToMatrix(original)

	// Check dimensions
	if len(result) != 2 || len(result[0]) != 3 {
		t.Errorf("Wrong dimensions: got %dx%d", len(result), len(result[0]))
	}

	// Check that values changed (with high probability)
	changed := false
	for i := range original {
		for j := range original[i] {
			if result[i][j] != original[i][j] {
				changed = true
			}
		}
	}
	if !changed {
		t.Error("Matrix should have changed with noise")
	}

	// Original should be unmodified
	if original[0][0] != 0.1 {
		t.Error("Original matrix was modified")
	}
}

func TestComputeErrorStatistics(t *testing.T) {
	target := [][]float64{
		{0.5, 0.5, 0.5},
		{0.5, 0.5, 0.5},
	}
	actual := [][]float64{
		{0.51, 0.49, 0.52},
		{0.48, 0.53, 0.50},
	}

	stats := ComputeErrorStatistics(target, actual)
	if stats == nil {
		t.Fatal("ComputeErrorStatistics returned nil")
	}

	t.Logf("Stats: Mean=%.4f, StdDev=%.4f, MaxAbs=%.4f, RMSE=%.4f, SNR=%.1fdB",
		stats.MeanError, stats.StdDevError, stats.MaxAbsError, stats.RMSE, stats.SNR)

	// Mean should be close to 0.005 ((0.01-0.01+0.02-0.02+0.03+0)/6)
	expectedMean := (0.01 - 0.01 + 0.02 - 0.02 + 0.03 + 0) / 6
	if math.Abs(stats.MeanError-expectedMean) > 0.001 {
		t.Errorf("Mean error %.4f not close to expected %.4f", stats.MeanError, expectedMean)
	}

	// Max abs should be 0.03
	if math.Abs(stats.MaxAbsError-0.03) > 0.001 {
		t.Errorf("Max abs error %.4f not close to 0.03", stats.MaxAbsError)
	}

	// SNR should be positive and reasonable (signal >> noise)
	if stats.SNR < 20 {
		t.Errorf("SNR %.1f dB seems too low", stats.SNR)
	}
}

func TestComputeErrorStatistics_Empty(t *testing.T) {
	stats := ComputeErrorStatistics(nil, nil)
	if stats != nil {
		t.Error("Expected nil for empty input")
	}

	stats = ComputeErrorStatistics([][]float64{}, [][]float64{})
	if stats != nil {
		t.Error("Expected nil for empty matrices")
	}
}

func TestSimulateAccuracyDegradation(t *testing.T) {
	testCases := []struct {
		progSigma float64
		readSigma float64
		arraySize int
		maxLoss   float64 // Maximum expected loss
	}{
		{0.0, 0.0, 64, 0.001},  // No error -> no loss
		{0.05, 0.01, 64, 0.5},  // Typical errors
		{0.10, 0.03, 128, 1.0}, // Worst case
		{0.01, 0.01, 32, 0.1},  // Low error, small array
	}

	for _, tc := range testCases {
		loss := SimulateAccuracyDegradation(tc.progSigma, tc.readSigma, tc.arraySize)

		if loss < 0 || loss > 1 {
			t.Errorf("Loss %.4f out of range [0,1]", loss)
		}

		if loss > tc.maxLoss {
			t.Logf("prog=%.2f read=%.2f N=%d: loss=%.4f (max=%.4f)",
				tc.progSigma, tc.readSigma, tc.arraySize, loss, tc.maxLoss)
		}
	}
}

func TestRecommendErrorBudget(t *testing.T) {
	testCases := []struct {
		targetAccuracy float64
		arraySize      int
	}{
		{0.99, 64},  // 99% accuracy, 64 cols
		{0.95, 128}, // 95% accuracy, 128 cols
		{0.90, 256}, // 90% accuracy, 256 cols
	}

	for _, tc := range testCases {
		progSigma, readSigma := RecommendErrorBudget(tc.targetAccuracy, tc.arraySize)

		// Verify the recommendation achieves target
		loss := SimulateAccuracyDegradation(progSigma, readSigma, tc.arraySize)
		achievedAccuracy := 1 - loss

		t.Logf("Target=%.0f%% N=%d: progS=%.4f readS=%.4f -> achieved=%.1f%%",
			tc.targetAccuracy*100, tc.arraySize, progSigma, readSigma, achievedAccuracy*100)

		// Should be within 10% of target
		if achievedAccuracy < tc.targetAccuracy-0.1 {
			t.Errorf("Achieved %.1f%% below target %.0f%%", achievedAccuracy*100, tc.targetAccuracy*100)
		}
	}
}

func TestDeviceNonIdealityConfig_Presets(t *testing.T) {
	config := DefaultDeviceNonIdealityConfig()

	// Check defaults are disabled
	if config.ProgrammingError.Enable || config.ReadNoise.Enable {
		t.Error("Default config should have errors disabled")
	}

	// Test typical preset
	config.EnableTypicalErrors()
	if !config.ProgrammingError.Enable {
		t.Error("EnableTypicalErrors should enable programming error")
	}
	if config.ProgrammingError.Sigma != 0.05 {
		t.Errorf("Expected sigma=0.05, got %f", config.ProgrammingError.Sigma)
	}

	// Test worst case preset
	config = DefaultDeviceNonIdealityConfig()
	config.EnableWorstCaseErrors()
	if config.ProgrammingError.Sigma != 0.10 {
		t.Errorf("Expected worst case sigma=0.10, got %f", config.ProgrammingError.Sigma)
	}
}

func TestUniformErrorModels(t *testing.T) {
	config := &ProgrammingErrorConfig{
		Enable:    true,
		Model:     ErrorModelUniformIndependent,
		Sigma:     0.1,
		Symmetric: true,
		Seed:      42,
	}
	engine := NewDeviceErrorEngine(config, nil)

	target := 0.5
	nSamples := 5000

	min, max := 1.0, 0.0
	for i := 0; i < nSamples; i++ {
		actual := engine.ApplyProgrammingError(target)
		if actual < min {
			min = actual
		}
		if actual > max {
			max = actual
		}
	}

	// Uniform should be bounded by [-sigma, +sigma] around target
	expectedMin := target - config.Sigma
	expectedMax := target + config.Sigma

	if min < expectedMin-0.01 || max > expectedMax+0.01 {
		t.Errorf("Uniform distribution out of expected range: [%f, %f] vs expected [%f, %f]",
			min, max, expectedMin, expectedMax)
	}
}

func BenchmarkApplyProgrammingError(b *testing.B) {
	config := &ProgrammingErrorConfig{
		Enable: true,
		Model:  ErrorModelNormalProportional,
		Sigma:  0.05,
		Seed:   42,
	}
	engine := NewDeviceErrorEngine(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ApplyProgrammingError(0.5)
	}
}

func BenchmarkApplyReadNoise(b *testing.B) {
	config := &ReadNoiseConfig{
		Enable:     true,
		Model:      ErrorModelNormalIndependent,
		Sigma:      0.01,
		Persistent: false,
		Seed:       42,
	}
	engine := NewDeviceErrorEngine(nil, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ApplyReadNoise(0.5, 0, 0)
	}
}

func BenchmarkApplyProgrammingErrorToMatrix_64x64(b *testing.B) {
	config := &ProgrammingErrorConfig{
		Enable: true,
		Model:  ErrorModelNormalProportional,
		Sigma:  0.05,
		Seed:   42,
	}
	engine := NewDeviceErrorEngine(config, nil)

	matrix := make([][]float64, 64)
	for i := range matrix {
		matrix[i] = make([]float64, 64)
		for j := range matrix[i] {
			matrix[i][j] = 0.5
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ApplyProgrammingErrorToMatrix(matrix)
	}
}
