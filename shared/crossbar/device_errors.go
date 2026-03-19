// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"math"
	"math/rand"
	"time"

	"fecim-lattice-tools/shared/mathutil"
)

// ErrorModel defines the type of error distribution.
// Based on CrossSim's generic_error.py models.
type ErrorModel int

const (
	// ErrorModelNone disables error modeling
	ErrorModelNone ErrorModel = iota
	// ErrorModelNormalIndependent applies Gaussian noise with fixed sigma
	ErrorModelNormalIndependent
	// ErrorModelNormalProportional applies Gaussian noise proportional to conductance
	ErrorModelNormalProportional
	// ErrorModelNormalInverseProportional applies Gaussian noise inversely proportional to conductance
	ErrorModelNormalInverseProportional
	// ErrorModelUniformIndependent applies uniform noise with fixed range
	ErrorModelUniformIndependent
	// ErrorModelUniformProportional applies uniform noise proportional to conductance
	ErrorModelUniformProportional
)

// String returns the string representation of the error model.
func (m ErrorModel) String() string {
	switch m {
	case ErrorModelNone:
		return "None"
	case ErrorModelNormalIndependent:
		return "Normal (Independent)"
	case ErrorModelNormalProportional:
		return "Normal (Proportional)"
	case ErrorModelNormalInverseProportional:
		return "Normal (Inverse Proportional)"
	case ErrorModelUniformIndependent:
		return "Uniform (Independent)"
	case ErrorModelUniformProportional:
		return "Uniform (Proportional)"
	default:
		return "Unknown"
	}
}

// ProgrammingErrorConfig configures programming (write) error modeling.
// Programming errors represent the difference between target and actual programmed conductance.
type ProgrammingErrorConfig struct {
	Enable    bool       // Enable programming error
	Model     ErrorModel // Error distribution model
	Sigma     float64    // Standard deviation (as fraction of range, e.g., 0.05 = 5%)
	Symmetric bool       // If true, error can be positive or negative; if false, only positive
	Seed      int64      // Random seed (0 = use time-based seed)
}

// DefaultProgrammingErrorConfig returns default programming error settings.
// Based on typical ReRAM/FeFET programming variability.
func DefaultProgrammingErrorConfig() *ProgrammingErrorConfig {
	return &ProgrammingErrorConfig{
		Enable:    false,
		Model:     ErrorModelNormalProportional,
		Sigma:     0.05, // 5% typical programming error
		Symmetric: true,
		Seed:      0,
	}
}

// ReadNoiseConfig configures read (per-operation) noise modeling.
// Read noise represents measurement noise during inference.
type ReadNoiseConfig struct {
	Enable     bool       // Enable read noise
	Model      ErrorModel // Noise distribution model
	Sigma      float64    // Noise magnitude (as fraction, e.g., 0.01 = 1%)
	Persistent bool       // If true, same noise per read of same cell; if false, varies each read
	Seed       int64      // Random seed (0 = use time-based seed)
}

// DefaultReadNoiseConfig returns default read noise settings.
func DefaultReadNoiseConfig() *ReadNoiseConfig {
	return &ReadNoiseConfig{
		Enable:     false,
		Model:      ErrorModelNormalIndependent,
		Sigma:      0.01, // 1% typical read noise
		Persistent: false,
		Seed:       0,
	}
}

// DeviceErrorEngine handles all device non-ideality modeling.
type DeviceErrorEngine struct {
	progConfig *ProgrammingErrorConfig
	readConfig *ReadNoiseConfig
	progRng    *rand.Rand
	readRng    *rand.Rand

	// Persistent noise storage (for Persistent=true mode)
	persistentNoise map[cellKey]float64
}

type cellKey struct {
	row, col int
}

// NewDeviceErrorEngine creates a new device error engine.
func NewDeviceErrorEngine(progConfig *ProgrammingErrorConfig, readConfig *ReadNoiseConfig) *DeviceErrorEngine {
	if progConfig == nil {
		progConfig = DefaultProgrammingErrorConfig()
	}
	if readConfig == nil {
		readConfig = DefaultReadNoiseConfig()
	}

	// Initialize random number generators
	progSeed := progConfig.Seed
	if progSeed == 0 {
		progSeed = time.Now().UnixNano()
	}
	readSeed := readConfig.Seed
	if readSeed == 0 {
		readSeed = time.Now().UnixNano() + 1
	}

	return &DeviceErrorEngine{
		progConfig:      progConfig,
		readConfig:      readConfig,
		progRng:         rand.New(rand.NewSource(progSeed)),
		readRng:         rand.New(rand.NewSource(readSeed)),
		persistentNoise: make(map[cellKey]float64),
	}
}

// ApplyProgrammingError applies programming error to a target conductance.
// Returns the actual (noisy) conductance after programming.
//
// G_programmed = G_target × (1 + noise) for proportional models
// G_programmed = G_target + noise for independent models
func (e *DeviceErrorEngine) ApplyProgrammingError(gTarget float64) float64 {
	if !e.progConfig.Enable {
		return gTarget
	}

	noise := e.generateNoise(e.progConfig.Model, e.progConfig.Sigma, gTarget, e.progRng)

	if !e.progConfig.Symmetric && noise < 0 {
		noise = -noise
	}

	var gProgrammed float64
	switch e.progConfig.Model {
	case ErrorModelNormalProportional, ErrorModelUniformProportional:
		// Multiplicative noise: G_prog = G_target × (1 + noise)
		gProgrammed = gTarget * (1 + noise)
	case ErrorModelNormalInverseProportional:
		// Inverse proportional: higher noise for lower conductance
		gProgrammed = gTarget + noise
	default:
		// Additive noise: G_prog = G_target + noise
		gProgrammed = gTarget + noise
	}

	// Clamp to valid range [0, 1]
	return mathutil.Clamp(gProgrammed, 0, 1)
}

// ApplyReadNoise applies read noise to a conductance value.
// Returns the noisy conductance for this read operation.
//
// G_read = G_programmed × (1 + noise) for typical read noise
func (e *DeviceErrorEngine) ApplyReadNoise(gProgrammed float64, row, col int) float64 {
	if !e.readConfig.Enable {
		return gProgrammed
	}

	var noise float64
	if e.readConfig.Persistent {
		// Use stored noise for this cell
		key := cellKey{row, col}
		if storedNoise, ok := e.persistentNoise[key]; ok {
			noise = storedNoise
		} else {
			noise = e.generateNoise(e.readConfig.Model, e.readConfig.Sigma, gProgrammed, e.readRng)
			e.persistentNoise[key] = noise
		}
	} else {
		// Generate new noise each read
		noise = e.generateNoise(e.readConfig.Model, e.readConfig.Sigma, gProgrammed, e.readRng)
	}

	gRead := gProgrammed * (1 + noise)
	return mathutil.Clamp(gRead, 0, 1)
}

// ClearPersistentNoise clears all stored persistent noise values.
// Call this to simulate device refresh or reprogramming.
func (e *DeviceErrorEngine) ClearPersistentNoise() {
	e.persistentNoise = make(map[cellKey]float64)
}

// generateNoise generates a noise value based on the error model.
func (e *DeviceErrorEngine) generateNoise(model ErrorModel, sigma, gValue float64, rng *rand.Rand) float64 {
	switch model {
	case ErrorModelNone:
		return 0

	case ErrorModelNormalIndependent:
		// Fixed sigma Gaussian
		return rng.NormFloat64() * sigma

	case ErrorModelNormalProportional:
		// Sigma proportional to conductance
		effectiveSigma := sigma * gValue
		if effectiveSigma < 1e-10 {
			effectiveSigma = sigma * 0.01 // Minimum sigma
		}
		return rng.NormFloat64() * effectiveSigma / gValue

	case ErrorModelNormalInverseProportional:
		// Sigma inversely proportional to conductance (higher noise at low G)
		effectiveSigma := sigma
		if gValue > 1e-10 {
			effectiveSigma = sigma / gValue
		}
		// Cap to prevent explosion
		if effectiveSigma > sigma*10 {
			effectiveSigma = sigma * 10
		}
		return rng.NormFloat64() * effectiveSigma

	case ErrorModelUniformIndependent:
		// Uniform distribution [-sigma, +sigma]
		return (rng.Float64()*2 - 1) * sigma

	case ErrorModelUniformProportional:
		// Uniform distribution proportional to conductance
		effectiveSigma := sigma * gValue
		if effectiveSigma < 1e-10 {
			effectiveSigma = sigma * 0.01
		}
		return (rng.Float64()*2 - 1) * effectiveSigma / gValue

	default:
		return 0
	}
}


// ApplyProgrammingErrorToMatrix applies programming error to an entire conductance matrix.
// Returns a new matrix with noisy values (original is not modified).
func (e *DeviceErrorEngine) ApplyProgrammingErrorToMatrix(g [][]float64) [][]float64 {
	rows := len(g)
	if rows == 0 {
		return nil
	}
	cols := len(g[0])

	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = e.ApplyProgrammingError(g[i][j])
		}
	}
	return result
}

// ApplyReadNoiseToMatrix applies read noise to an entire conductance matrix.
// Returns a new matrix with noisy values (original is not modified).
func (e *DeviceErrorEngine) ApplyReadNoiseToMatrix(g [][]float64) [][]float64 {
	rows := len(g)
	if rows == 0 {
		return nil
	}
	cols := len(g[0])

	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = e.ApplyReadNoise(g[i][j], i, j)
		}
	}
	return result
}

// ErrorStatistics contains statistics about applied errors.
type ErrorStatistics struct {
	MeanError       float64 // Mean error (G_actual - G_target)
	StdDevError     float64 // Standard deviation of errors
	MaxAbsError     float64 // Maximum absolute error
	MinAbsError     float64 // Minimum absolute error
	RMSE            float64 // Root mean square error
	SNR             float64 // Signal-to-noise ratio (dB)
	PercentOutliers float64 // Percentage of values > 3 sigma
}

// ComputeErrorStatistics computes statistics comparing target and actual matrices.
func ComputeErrorStatistics(target, actual [][]float64) *ErrorStatistics {
	if len(target) == 0 || len(actual) == 0 {
		return nil
	}

	rows := len(target)
	cols := len(target[0])
	n := float64(rows * cols)

	var sumError, sumSqError, sumSqTarget float64
	maxAbs := 0.0
	minAbs := math.MaxFloat64

	// Collect all errors for outlier calculation
	errors := make([]float64, 0, rows*cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			err := actual[i][j] - target[i][j]
			absErr := math.Abs(err)

			errors = append(errors, err)
			sumError += err
			sumSqError += err * err
			sumSqTarget += target[i][j] * target[i][j]

			if absErr > maxAbs {
				maxAbs = absErr
			}
			if absErr < minAbs {
				minAbs = absErr
			}
		}
	}

	meanError := sumError / n
	variance := (sumSqError / n) - (meanError * meanError)
	stdDev := math.Sqrt(variance)
	rmse := math.Sqrt(sumSqError / n)

	// SNR = 10 * log10(signal_power / noise_power)
	signalPower := sumSqTarget / n
	noisePower := sumSqError / n
	var snr float64
	if noisePower > 1e-20 {
		snr = 10 * math.Log10(signalPower/noisePower)
	} else {
		snr = math.Inf(1) // No noise
	}

	// Count outliers (> 3 sigma)
	outlierThreshold := 3 * stdDev
	outlierCount := 0
	for _, err := range errors {
		if math.Abs(err-meanError) > outlierThreshold {
			outlierCount++
		}
	}
	percentOutliers := float64(outlierCount) / n * 100

	return &ErrorStatistics{
		MeanError:       meanError,
		StdDevError:     stdDev,
		MaxAbsError:     maxAbs,
		MinAbsError:     minAbs,
		RMSE:            rmse,
		SNR:             snr,
		PercentOutliers: percentOutliers,
	}
}

// SimulateAccuracyDegradation simulates MVM accuracy degradation due to device errors.
// Returns the expected accuracy loss as a fraction (0-1).
func SimulateAccuracyDegradation(progSigma, readSigma float64, arraySize int) float64 {
	// Empirical model based on CrossSim validation
	// Accuracy loss ~ sqrt(N) × (prog_sigma + read_sigma) / sqrt(2)
	// where N is the number of accumulations per output
	n := float64(arraySize)
	totalSigma := math.Sqrt(progSigma*progSigma + readSigma*readSigma)

	// Error compounds with sqrt(N) accumulations
	expectedError := math.Sqrt(n) * totalSigma / math.Sqrt(2)

	// Cap at 100% loss
	if expectedError > 1 {
		expectedError = 1
	}

	return expectedError
}

// RecommendErrorBudget recommends sigma values to achieve target accuracy.
// Returns (progSigma, readSigma) that should maintain targetAccuracy.
func RecommendErrorBudget(targetAccuracy float64, arraySize int) (progSigma, readSigma float64) {
	// Work backwards from accuracy target
	// accuracy = 1 - loss, loss = sqrt(N) × total_sigma / sqrt(2)
	// total_sigma = (1 - accuracy) × sqrt(2) / sqrt(N)
	allowedLoss := 1 - targetAccuracy
	n := float64(arraySize)
	totalSigma := allowedLoss * math.Sqrt(2) / math.Sqrt(n)

	// Split budget equally between programming and read
	eachSigma := totalSigma / math.Sqrt(2)

	return eachSigma, eachSigma
}

// DeviceNonIdealityConfig combines all device non-ideality settings.
type DeviceNonIdealityConfig struct {
	ProgrammingError *ProgrammingErrorConfig
	ReadNoise        *ReadNoiseConfig
}

// DefaultDeviceNonIdealityConfig returns typical device non-ideality settings.
func DefaultDeviceNonIdealityConfig() *DeviceNonIdealityConfig {
	return &DeviceNonIdealityConfig{
		ProgrammingError: DefaultProgrammingErrorConfig(),
		ReadNoise:        DefaultReadNoiseConfig(),
	}
}

// EnableTypicalErrors enables typical error modeling with reasonable defaults.
func (c *DeviceNonIdealityConfig) EnableTypicalErrors() {
	c.ProgrammingError.Enable = true
	c.ProgrammingError.Model = ErrorModelNormalProportional
	c.ProgrammingError.Sigma = 0.05 // 5%

	c.ReadNoise.Enable = true
	c.ReadNoise.Model = ErrorModelNormalIndependent
	c.ReadNoise.Sigma = 0.01 // 1%
}

// EnableWorstCaseErrors enables conservative error modeling.
func (c *DeviceNonIdealityConfig) EnableWorstCaseErrors() {
	c.ProgrammingError.Enable = true
	c.ProgrammingError.Model = ErrorModelNormalProportional
	c.ProgrammingError.Sigma = 0.10 // 10%

	c.ReadNoise.Enable = true
	c.ReadNoise.Model = ErrorModelNormalIndependent
	c.ReadNoise.Sigma = 0.03 // 3%
}
