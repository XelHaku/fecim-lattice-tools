// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"math"
)

// BatchNormConfig holds batch normalization parameters.
type BatchNormConfig struct {
	NumFeatures int     // Number of features (channels)
	Epsilon     float64 // Small constant for numerical stability
	Momentum    float64 // Running statistics momentum
	Affine      bool    // Learnable gamma and beta
}

// DefaultBatchNormConfig returns default batch normalization config.
func DefaultBatchNormConfig(numFeatures int) *BatchNormConfig {
	return &BatchNormConfig{
		NumFeatures: numFeatures,
		Epsilon:     1e-5,
		Momentum:    0.1,
		Affine:      true,
	}
}

// BatchNorm implements batch normalization for crossbar inference.
// For analog CIM, BN is typically fused with linear layer weights.
type BatchNorm struct {
	config *BatchNormConfig

	// Learnable parameters (if affine)
	Gamma []float64 // Scale parameter
	Beta  []float64 // Shift parameter

	// Running statistics
	RunningMean []float64
	RunningVar  []float64

	// Training mode
	Training bool
}

// NewBatchNorm creates a new batch normalization layer.
func NewBatchNorm(config *BatchNormConfig) *BatchNorm {
	if config == nil {
		config = &BatchNormConfig{
			NumFeatures: 1,
			Epsilon:     1e-5,
			Momentum:    0.1,
			Affine:      true,
		}
	}

	bn := &BatchNorm{
		config:      config,
		Gamma:       make([]float64, config.NumFeatures),
		Beta:        make([]float64, config.NumFeatures),
		RunningMean: make([]float64, config.NumFeatures),
		RunningVar:  make([]float64, config.NumFeatures),
		Training:    false,
	}

	// Initialize gamma to 1, beta to 0, var to 1
	for i := range bn.Gamma {
		bn.Gamma[i] = 1.0
		bn.RunningVar[i] = 1.0
	}

	return bn
}

// Forward applies batch normalization to input.
func (bn *BatchNorm) Forward(input []float64) []float64 {
	if len(input) != bn.config.NumFeatures {
		// Pad or truncate
		result := make([]float64, bn.config.NumFeatures)
		copy(result, input)
		input = result
	}

	output := make([]float64, bn.config.NumFeatures)

	if bn.Training {
		// Compute batch statistics (single sample in this case)
		mean := 0.0
		for _, v := range input {
			mean += v
		}
		mean /= float64(len(input))

		variance := 0.0
		for _, v := range input {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(len(input))

		// Normalize with current batch stats
		for i := range output {
			normalized := (input[i] - mean) / math.Sqrt(variance+bn.config.Epsilon)
			if bn.config.Affine {
				output[i] = bn.Gamma[i]*normalized + bn.Beta[i]
			} else {
				output[i] = normalized
			}
		}

		// Update running statistics
		bn.updateRunningStats(mean, variance)
	} else {
		// Use running statistics for inference
		for i := range output {
			normalized := (input[i] - bn.RunningMean[i]) / math.Sqrt(bn.RunningVar[i]+bn.config.Epsilon)
			if bn.config.Affine {
				output[i] = bn.Gamma[i]*normalized + bn.Beta[i]
			} else {
				output[i] = normalized
			}
		}
	}

	return output
}

// ForwardBatch applies batch normalization to a batch of inputs.
func (bn *BatchNorm) ForwardBatch(batch [][]float64) [][]float64 {
	if len(batch) == 0 {
		return nil
	}

	batchSize := len(batch)
	numFeatures := bn.config.NumFeatures

	output := make([][]float64, batchSize)
	for i := range output {
		output[i] = make([]float64, numFeatures)
	}

	if bn.Training {
		// Compute batch mean
		mean := make([]float64, numFeatures)
		for _, sample := range batch {
			for j := 0; j < numFeatures && j < len(sample); j++ {
				mean[j] += sample[j]
			}
		}
		for j := range mean {
			mean[j] /= float64(batchSize)
		}

		// Compute batch variance
		variance := make([]float64, numFeatures)
		for _, sample := range batch {
			for j := 0; j < numFeatures && j < len(sample); j++ {
				diff := sample[j] - mean[j]
				variance[j] += diff * diff
			}
		}
		for j := range variance {
			variance[j] /= float64(batchSize)
		}

		// Normalize
		for i, sample := range batch {
			for j := 0; j < numFeatures && j < len(sample); j++ {
				normalized := (sample[j] - mean[j]) / math.Sqrt(variance[j]+bn.config.Epsilon)
				if bn.config.Affine {
					output[i][j] = bn.Gamma[j]*normalized + bn.Beta[j]
				} else {
					output[i][j] = normalized
				}
			}
		}

		// Update running statistics (per feature)
		for j := range mean {
			bn.RunningMean[j] = (1-bn.config.Momentum)*bn.RunningMean[j] + bn.config.Momentum*mean[j]
			bn.RunningVar[j] = (1-bn.config.Momentum)*bn.RunningVar[j] + bn.config.Momentum*variance[j]
		}
	} else {
		// Use running statistics
		for i, sample := range batch {
			for j := 0; j < numFeatures && j < len(sample); j++ {
				normalized := (sample[j] - bn.RunningMean[j]) / math.Sqrt(bn.RunningVar[j]+bn.config.Epsilon)
				if bn.config.Affine {
					output[i][j] = bn.Gamma[j]*normalized + bn.Beta[j]
				} else {
					output[i][j] = normalized
				}
			}
		}
	}

	return output
}

func (bn *BatchNorm) updateRunningStats(mean, variance float64) {
	momentum := bn.config.Momentum
	for i := range bn.RunningMean {
		bn.RunningMean[i] = (1-momentum)*bn.RunningMean[i] + momentum*mean
		bn.RunningVar[i] = (1-momentum)*bn.RunningVar[i] + momentum*variance
	}
}

// SetTraining sets training mode.
func (bn *BatchNorm) SetTraining(training bool) {
	bn.Training = training
}

// FuseWithLinear fuses batch normalization with preceding linear layer weights.
// This is the preferred method for analog CIM deployment.
// Returns fused weights and biases: W_fused = gamma/sqrt(var+eps) * W
//                                   b_fused = gamma/sqrt(var+eps) * (b - mean) + beta
func (bn *BatchNorm) FuseWithLinear(weights [][]float64, biases []float64) ([][]float64, []float64) {
	numOutputs := len(weights)
	if numOutputs == 0 {
		return weights, biases
	}
	numInputs := len(weights[0])

	// Compute scale factor for each output
	scale := make([]float64, bn.config.NumFeatures)
	for i := range scale {
		if i < numOutputs {
			scale[i] = bn.Gamma[i] / math.Sqrt(bn.RunningVar[i]+bn.config.Epsilon)
		}
	}

	// Fused weights
	fusedWeights := make([][]float64, numOutputs)
	for i := range fusedWeights {
		fusedWeights[i] = make([]float64, numInputs)
		s := 1.0
		if i < len(scale) {
			s = scale[i]
		}
		for j := range fusedWeights[i] {
			fusedWeights[i][j] = weights[i][j] * s
		}
	}

	// Fused biases
	fusedBiases := make([]float64, numOutputs)
	for i := range fusedBiases {
		s := 1.0
		mean := 0.0
		beta := 0.0
		if i < len(scale) {
			s = scale[i]
			mean = bn.RunningMean[i]
			beta = bn.Beta[i]
		}
		b := 0.0
		if i < len(biases) {
			b = biases[i]
		}
		fusedBiases[i] = s*(b-mean) + beta
	}

	return fusedWeights, fusedBiases
}

// GetFusionScaleShift returns scale and shift for in-memory BN fusion.
// scale[i] = gamma[i] / sqrt(var[i] + eps)
// shift[i] = beta[i] - scale[i] * mean[i]
func (bn *BatchNorm) GetFusionScaleShift() (scale, shift []float64) {
	scale = make([]float64, bn.config.NumFeatures)
	shift = make([]float64, bn.config.NumFeatures)

	for i := range scale {
		scale[i] = bn.Gamma[i] / math.Sqrt(bn.RunningVar[i]+bn.config.Epsilon)
		shift[i] = bn.Beta[i] - scale[i]*bn.RunningMean[i]
	}

	return scale, shift
}

// LoadParameters loads trained BN parameters.
func (bn *BatchNorm) LoadParameters(gamma, beta, runningMean, runningVar []float64) {
	copy(bn.Gamma, gamma)
	copy(bn.Beta, beta)
	copy(bn.RunningMean, runningMean)
	copy(bn.RunningVar, runningVar)
}

// LayerNorm implements layer normalization (alternative to batch norm).
// More suitable for transformers and RNNs.
type LayerNorm struct {
	NumFeatures int
	Epsilon     float64
	Gamma       []float64
	Beta        []float64
}

// NewLayerNorm creates a new layer normalization.
func NewLayerNorm(numFeatures int) *LayerNorm {
	ln := &LayerNorm{
		NumFeatures: numFeatures,
		Epsilon:     1e-5,
		Gamma:       make([]float64, numFeatures),
		Beta:        make([]float64, numFeatures),
	}
	for i := range ln.Gamma {
		ln.Gamma[i] = 1.0
	}
	return ln
}

// Forward applies layer normalization.
func (ln *LayerNorm) Forward(input []float64) []float64 {
	// Compute mean and variance across features
	mean := 0.0
	for _, v := range input {
		mean += v
	}
	mean /= float64(len(input))

	variance := 0.0
	for _, v := range input {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(input))

	// Normalize
	output := make([]float64, len(input))
	for i := range output {
		normalized := (input[i] - mean) / math.Sqrt(variance+ln.Epsilon)
		if i < ln.NumFeatures {
			output[i] = ln.Gamma[i]*normalized + ln.Beta[i]
		} else {
			output[i] = normalized
		}
	}

	return output
}

// RMSNorm implements Root Mean Square Layer Normalization.
// Popular in modern transformers (LLaMA, etc.) - simpler than LayerNorm.
type RMSNorm struct {
	NumFeatures int
	Epsilon     float64
	Gamma       []float64
}

// NewRMSNorm creates a new RMS normalization layer.
func NewRMSNorm(numFeatures int) *RMSNorm {
	rms := &RMSNorm{
		NumFeatures: numFeatures,
		Epsilon:     1e-5,
		Gamma:       make([]float64, numFeatures),
	}
	for i := range rms.Gamma {
		rms.Gamma[i] = 1.0
	}
	return rms
}

// Forward applies RMS normalization.
func (rms *RMSNorm) Forward(input []float64) []float64 {
	// Compute RMS
	sumSq := 0.0
	for _, v := range input {
		sumSq += v * v
	}
	rmsVal := math.Sqrt(sumSq/float64(len(input)) + rms.Epsilon)

	// Normalize
	output := make([]float64, len(input))
	for i := range output {
		normalized := input[i] / rmsVal
		if i < rms.NumFeatures {
			output[i] = rms.Gamma[i] * normalized
		} else {
			output[i] = normalized
		}
	}

	return output
}

// QuantizedBatchNorm implements batch normalization with quantized parameters
// for efficient analog CIM deployment.
type QuantizedBatchNorm struct {
	*BatchNorm
	Bits       int
	ScaleQ     []int32   // Quantized scale
	ShiftQ     []int32   // Quantized shift
	ScaleFP    float64   // Floating-point scale factor
}

// NewQuantizedBatchNorm creates quantized batch norm from trained BN.
func NewQuantizedBatchNorm(bn *BatchNorm, bits int) *QuantizedBatchNorm {
	qbn := &QuantizedBatchNorm{
		BatchNorm: bn,
		Bits:      bits,
		ScaleQ:    make([]int32, bn.config.NumFeatures),
		ShiftQ:    make([]int32, bn.config.NumFeatures),
	}

	// Compute floating-point scale and shift
	scale, shift := bn.GetFusionScaleShift()

	// Find max absolute value for quantization
	maxAbs := 0.0
	for i := range scale {
		if math.Abs(scale[i]) > maxAbs {
			maxAbs = math.Abs(scale[i])
		}
		if math.Abs(shift[i]) > maxAbs {
			maxAbs = math.Abs(shift[i])
		}
	}

	// Quantize
	levels := float64(int32(1) << (bits - 1))
	qbn.ScaleFP = maxAbs / levels

	for i := range scale {
		qbn.ScaleQ[i] = int32(math.Round(scale[i] / qbn.ScaleFP))
		qbn.ShiftQ[i] = int32(math.Round(shift[i] / qbn.ScaleFP))
	}

	return qbn
}

// ForwardQuantized applies quantized batch normalization.
func (qbn *QuantizedBatchNorm) ForwardQuantized(input []float64) []float64 {
	output := make([]float64, len(input))
	for i := range output {
		if i < len(qbn.ScaleQ) {
			// Use quantized parameters
			output[i] = float64(qbn.ScaleQ[i])*qbn.ScaleFP*input[i] + float64(qbn.ShiftQ[i])*qbn.ScaleFP
		} else {
			output[i] = input[i]
		}
	}
	return output
}
