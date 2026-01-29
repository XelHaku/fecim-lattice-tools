// Package core provides the dual-mode inference engine for FeCIM MNIST demo.
// It implements both full-precision (FP) and compute-in-memory (CIM) paths
// to demonstrate the impact of 30-level ferroelectric quantization.
package core

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/physics"
)

// Note: logger is shared with network.go (var log = logging.NewLogger("mnist-core"))

// FeCIMLevels is the number of discrete conductance levels in FeCIM hardware.
// This is the key physical constraint from Dr. Tour's research.
// Re-exported from shared/physics for backward compatibility.
const FeCIMLevels = physics.DefaultLevels

// QuantizeWeights quantizes FP weights to N discrete levels
// using symmetric range [-W_max, +W_max] with linear mapping.
//
// Physical Mapping:
// Each quantization level represents a discrete conductance state in the FeCIM device.
// In hardware, these map to conductance values: G = Gmin + (Gmax - Gmin) * normalized_level
// where Gmin = 10µS and Gmax = 100µS (from docs).
// For computational efficiency, we store normalized values [−W_max, +W_max] but the
// conceptual mapping to [Gmin, Gmax] is implicit.
func QuantizeWeights(fpWeights [][]float64, levels int) ([][]float64, error) {
	log.Input("QuantizeWeights", map[string]interface{}{
		"levels": levels,
		"rows":   len(fpWeights),
		"cols":   func() int { if len(fpWeights) > 0 { return len(fpWeights[0]) }; return 0 }(),
	})

	if levels < 2 {
		err := fmt.Errorf("quantize: levels must be >= 2, got %d", levels)
		log.ErrorContext("QuantizeWeights", err, nil)
		return nil, err
	}

	rows := len(fpWeights)
	if rows == 0 {
		return fpWeights, nil
	}
	cols := len(fpWeights[0])
	if cols == 0 {
		return fpWeights, nil
	}

	// 1. Find global max magnitude (symmetric)
	wMax := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if abs := math.Abs(fpWeights[i][j]); abs > wMax {
				wMax = abs
			}
		}
	}

	if wMax == 0 {
		return fpWeights, nil // All zeros
	}

	// 2. Quantize to integer bins [0, levels-1]
	// Each bin represents a discrete conductance level in the FeCIM device
	quantized := make([][]float64, rows)
	levelStep := 2.0 * wMax / float64(levels-1) // Level spacing

	for i := 0; i < rows; i++ {
		quantized[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Map [-wMax, +wMax] → [0, 1]
			normalized := (fpWeights[i][j] + wMax) / (2.0 * wMax)

			// Quantize to bin (conductance level)
			bin := int(math.Round(normalized * float64(levels-1)))

			// Clamp
			if bin < 0 {
				bin = 0
			}
			if bin >= levels {
				bin = levels - 1
			}

			// Map back to [-wMax, +wMax]
			// In hardware: this represents G = Gmin + bin/(levels-1) * (Gmax - Gmin)
			quantized[i][j] = -wMax + float64(bin)*levelStep
		}
	}

	log.Calculation("QuantizeWeights", map[string]interface{}{
		"wMax":       wMax,
		"levelStep":  levelStep,
		"dims":       fmt.Sprintf("%dx%d", rows, cols),
	}, "quantized")

	return quantized, nil
}

// QuantizeBias quantizes bias vector to N discrete levels.
func QuantizeBias(fpBias []float64, levels int) ([]float64, error) {
	log.Trace("QuantizeBias: levels=%d, len=%d", levels, len(fpBias))

	if len(fpBias) == 0 {
		return fpBias, nil
	}
	// Wrap as 2D array for code reuse
	wrapped := [][]float64{fpBias}
	quantized, err := QuantizeWeights(wrapped, levels)
	if err != nil {
		log.ErrorContext("QuantizeBias", err, map[string]interface{}{"levels": levels})
		return nil, err
	}
	return quantized[0], nil
}

// QuantizationStats returns quantization metrics for analysis.
type QuantizationStats struct {
	OriginalRange  float64 // [-W_max, +W_max]
	QuantizedRange float64
	LevelSpacing   float64
	NumDistinct    int     // Unique values after quantization
	MSE            float64 // Mean squared error
	PSNR           float64 // Peak signal-to-noise ratio (dB)
	MaxError       float64 // Maximum absolute error
}

// ComputeQuantizationStats computes metrics comparing original and quantized weights.
func ComputeQuantizationStats(original, quantized [][]float64) QuantizationStats {
	stats := QuantizationStats{}

	if len(original) == 0 || len(original[0]) == 0 {
		log.Trace("ComputeQuantizationStats: empty input")
		return stats
	}

	rows := len(original)
	cols := len(original[0])

	// Find range of original weights
	oMin, oMax := original[0][0], original[0][0]
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if original[i][j] < oMin {
				oMin = original[i][j]
			}
			if original[i][j] > oMax {
				oMax = original[i][j]
			}
		}
	}
	stats.OriginalRange = oMax - oMin

	// Find range of quantized weights
	qMin, qMax := quantized[0][0], quantized[0][0]
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if quantized[i][j] < qMin {
				qMin = quantized[i][j]
			}
			if quantized[i][j] > qMax {
				qMax = quantized[i][j]
			}
		}
	}
	stats.QuantizedRange = qMax - qMin

	// Count distinct values
	distinctMap := make(map[float64]bool)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			distinctMap[quantized[i][j]] = true
		}
	}
	stats.NumDistinct = len(distinctMap)

	// Calculate level spacing
	if stats.NumDistinct > 1 {
		stats.LevelSpacing = stats.QuantizedRange / float64(stats.NumDistinct-1)
	}

	// Calculate MSE and max error
	sumSquaredError := 0.0
	maxAbsError := 0.0
	count := float64(rows * cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			diff := original[i][j] - quantized[i][j]
			sumSquaredError += diff * diff
			if absErr := math.Abs(diff); absErr > maxAbsError {
				maxAbsError = absErr
			}
		}
	}

	stats.MSE = sumSquaredError / count
	stats.MaxError = maxAbsError

	// Calculate PSNR (Peak Signal-to-Noise Ratio)
	// PSNR = 10 * log10(MAX^2 / MSE)
	if stats.MSE > 0 {
		maxVal := math.Max(math.Abs(oMin), math.Abs(oMax))
		stats.PSNR = 10 * math.Log10(maxVal*maxVal/stats.MSE)
	} else {
		stats.PSNR = math.Inf(1) // Perfect reconstruction
	}

	log.Calculation("ComputeQuantizationStats", map[string]interface{}{
		"numDistinct":  stats.NumDistinct,
		"levelSpacing": stats.LevelSpacing,
		"MSE":          stats.MSE,
		"PSNR_dB":      stats.PSNR,
	}, stats)

	return stats
}

// AddGaussianNoise adds Gaussian noise to values with given standard deviation.
// noiseLevel is specified as the standard deviation σ (additive noise model).
//
// Physical Model:
// This models read noise from thermal/shot noise in the analog readout circuitry.
// The noise is additive (constant σ) rather than multiplicative, as thermal and shot
// noise do not scale with signal amplitude.
func AddGaussianNoise(values []float64, noiseLevel float64, rng *RandomSource) []float64 {
	log.Trace("AddGaussianNoise: noiseLevel=%.4f, len=%d", noiseLevel, len(values))

	if noiseLevel <= 0 {
		return values
	}

	result := make([]float64, len(values))
	for i, v := range values {
		// Additive Gaussian noise (models thermal/shot read noise)
		// sigma is constant, not proportional to signal
		result[i] = v + rng.NormFloat64()*noiseLevel
	}
	return result
}

// RandomSource provides deterministic random number generation.
type RandomSource struct {
	seed uint64
}

// NewRandomSource creates a new random source with given seed.
func NewRandomSource(seed uint64) *RandomSource {
	return &RandomSource{seed: seed}
}

// NormFloat64 returns a normally distributed float64 using Box-Muller transform.
func (r *RandomSource) NormFloat64() float64 {
	u1 := r.Float64()
	u2 := r.Float64()
	// Box-Muller transform
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

// Float64 returns a uniformly distributed float64 in [0, 1).
func (r *RandomSource) Float64() float64 {
	// Simple xorshift64 PRNG
	r.seed ^= r.seed << 13
	r.seed ^= r.seed >> 7
	r.seed ^= r.seed << 17
	return float64(r.seed) / float64(math.MaxUint64)
}

// Intn returns a random int in [0, n).
func (r *RandomSource) Intn(n int) int {
	return int(r.Float64() * float64(n))
}
