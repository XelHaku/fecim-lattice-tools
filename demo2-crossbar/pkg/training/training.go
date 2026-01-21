// Package training provides hardware-aware training utilities for crossbar arrays.
// Implements backpropagation with analog non-ideality simulation.
package training

import (
	"fmt"
	"math"
	"math/rand"
)

// TrainingConfig holds training hyperparameters.
type TrainingConfig struct {
	LearningRate float64 // Base learning rate
	WeightDecay  float64 // L2 regularization
	Momentum     float64 // Momentum coefficient
	BatchSize    int     // Mini-batch size
	Epochs       int     // Number of training epochs

	// Hardware-aware parameters
	WeightClipMin  float64 // Minimum weight (conductance)
	WeightClipMax  float64 // Maximum weight (conductance)
	UpdateNoise    float64 // Noise in weight updates (σ)
	QuantizeBits   int     // Weight quantization bits (0 = no quantization)
	AsymmetryRatio float64 // Potentiation/depression asymmetry (1.0 = symmetric)
}

// DefaultTrainingConfig returns default training configuration.
func DefaultTrainingConfig() *TrainingConfig {
	return &TrainingConfig{
		LearningRate:   0.01,
		WeightDecay:    1e-4,
		Momentum:       0.9,
		BatchSize:      32,
		Epochs:         10,
		WeightClipMin:  0.0,
		WeightClipMax:  1.0,
		UpdateNoise:    0.0,
		QuantizeBits:   0,
		AsymmetryRatio: 1.0,
	}
}

// Trainer handles neural network training on crossbar arrays.
type Trainer struct {
	config     *TrainingConfig
	weights    [][][]float64 // Layer weights [layer][out][in]
	biases     [][]float64   // Layer biases [layer][out]
	velocities [][][]float64 // Momentum velocities
	biasVels   [][]float64   // Bias velocities
}

// NewTrainer creates a new trainer with given layer dimensions.
// dims specifies the size of each layer [input, hidden1, hidden2, ..., output]
func NewTrainer(dims []int, config *TrainingConfig) (*Trainer, error) {
	if len(dims) < 2 {
		return nil, fmt.Errorf("need at least 2 layers")
	}
	if config == nil {
		config = DefaultTrainingConfig()
	}

	numLayers := len(dims) - 1
	t := &Trainer{
		config:     config,
		weights:    make([][][]float64, numLayers),
		biases:     make([][]float64, numLayers),
		velocities: make([][][]float64, numLayers),
		biasVels:   make([][]float64, numLayers),
	}

	// Initialize weights with Xavier/Glorot initialization
	for l := 0; l < numLayers; l++ {
		inSize := dims[l]
		outSize := dims[l+1]
		stddev := math.Sqrt(2.0 / float64(inSize+outSize))

		t.weights[l] = make([][]float64, outSize)
		t.velocities[l] = make([][]float64, outSize)
		t.biases[l] = make([]float64, outSize)
		t.biasVels[l] = make([]float64, outSize)

		for i := 0; i < outSize; i++ {
			t.weights[l][i] = make([]float64, inSize)
			t.velocities[l][i] = make([]float64, inSize)
			for j := 0; j < inSize; j++ {
				// Initialize centered at 0.5 for crossbar [0,1] range
				w := rand.NormFloat64()*stddev*0.5 + 0.5
				t.weights[l][i][j] = clip(w, config.WeightClipMin, config.WeightClipMax)
			}
		}
	}

	return t, nil
}

// Forward performs forward pass with activation caching for backprop.
func (t *Trainer) Forward(input []float64) (output []float64, activations [][]float64, preActivations [][]float64) {
	activations = make([][]float64, len(t.weights)+1)
	preActivations = make([][]float64, len(t.weights))

	activations[0] = input
	current := input

	for l := range t.weights {
		// Matrix multiply: z = W * a + b
		z := make([]float64, len(t.biases[l]))
		for i := range z {
			sum := t.biases[l][i]
			for j, w := range t.weights[l][i] {
				// Apply quantization noise during forward pass
				effectiveW := t.applyHardwareNoise(w, true)
				sum += effectiveW * current[j]
			}
			z[i] = sum
		}
		preActivations[l] = z

		// Apply activation
		if l < len(t.weights)-1 {
			current = relu(z)
		} else {
			current = softmax(z)
		}
		activations[l+1] = current
	}

	return current, activations, preActivations
}

// Backward performs backward pass to compute gradients.
func (t *Trainer) Backward(activations, preActivations [][]float64, target []float64) (weightGrads [][][]float64, biasGrads [][]float64) {
	numLayers := len(t.weights)
	weightGrads = make([][][]float64, numLayers)
	biasGrads = make([][]float64, numLayers)

	// Output layer delta (cross-entropy + softmax derivative simplifies to output - target)
	delta := make([]float64, len(activations[numLayers]))
	for i := range delta {
		delta[i] = activations[numLayers][i] - target[i]
	}

	// Backpropagate through layers
	for l := numLayers - 1; l >= 0; l-- {
		outSize := len(t.weights[l])
		inSize := len(t.weights[l][0])

		// Compute weight gradients: dW = delta * a^T
		weightGrads[l] = make([][]float64, outSize)
		biasGrads[l] = make([]float64, outSize)

		for i := 0; i < outSize; i++ {
			weightGrads[l][i] = make([]float64, inSize)
			biasGrads[l][i] = delta[i]
			for j := 0; j < inSize; j++ {
				weightGrads[l][i][j] = delta[i] * activations[l][j]
			}
		}

		// Compute delta for next layer (if not input)
		if l > 0 {
			newDelta := make([]float64, inSize)
			for j := 0; j < inSize; j++ {
				sum := 0.0
				for i := 0; i < outSize; i++ {
					sum += t.weights[l][i][j] * delta[i]
				}
				// ReLU derivative
				if preActivations[l-1][j] > 0 {
					newDelta[j] = sum
				}
			}
			delta = newDelta
		}
	}

	return weightGrads, biasGrads
}

// UpdateWeights applies gradients with hardware-aware update rules.
func (t *Trainer) UpdateWeights(weightGrads [][][]float64, biasGrads [][]float64) {
	lr := t.config.LearningRate
	momentum := t.config.Momentum
	decay := t.config.WeightDecay
	asymmetry := t.config.AsymmetryRatio

	for l := range t.weights {
		for i := range t.weights[l] {
			// Update bias
			t.biasVels[l][i] = momentum*t.biasVels[l][i] - lr*biasGrads[l][i]
			t.biases[l][i] += t.biasVels[l][i]

			for j := range t.weights[l][i] {
				grad := weightGrads[l][i][j] + decay*t.weights[l][i][j]

				// Apply momentum
				t.velocities[l][i][j] = momentum*t.velocities[l][i][j] - lr*grad

				// Hardware-aware update
				update := t.velocities[l][i][j]

				// Apply asymmetry (potentiation vs depression)
				if update > 0 {
					update *= asymmetry // Potentiation
				}
				// Depression uses base rate (asymmetry < 1 means easier depression)

				// Apply update noise
				if t.config.UpdateNoise > 0 {
					update += rand.NormFloat64() * t.config.UpdateNoise * math.Abs(update)
				}

				// Update weight
				newW := t.weights[l][i][j] + update

				// Apply hardware constraints
				newW = t.applyHardwareConstraints(newW)
				t.weights[l][i][j] = newW
			}
		}
	}
}

// TrainBatch trains on a single batch and returns loss.
func (t *Trainer) TrainBatch(inputs [][]float64, targets [][]float64) float64 {
	if len(inputs) == 0 {
		return 0
	}

	// Accumulate gradients
	numLayers := len(t.weights)
	accumWeightGrads := make([][][]float64, numLayers)
	accumBiasGrads := make([][]float64, numLayers)

	// Initialize accumulators
	for l := range t.weights {
		accumWeightGrads[l] = make([][]float64, len(t.weights[l]))
		accumBiasGrads[l] = make([]float64, len(t.biases[l]))
		for i := range t.weights[l] {
			accumWeightGrads[l][i] = make([]float64, len(t.weights[l][i]))
		}
	}

	var totalLoss float64
	batchSize := float64(len(inputs))

	for b := range inputs {
		// Forward pass
		output, activations, preActivations := t.Forward(inputs[b])

		// Compute loss (cross-entropy)
		for i, target := range targets[b] {
			if target > 0 && output[i] > 0 {
				totalLoss -= target * math.Log(output[i]+1e-10)
			}
		}

		// Backward pass
		weightGrads, biasGrads := t.Backward(activations, preActivations, targets[b])

		// Accumulate gradients
		for l := range weightGrads {
			for i := range weightGrads[l] {
				accumBiasGrads[l][i] += biasGrads[l][i] / batchSize
				for j := range weightGrads[l][i] {
					accumWeightGrads[l][i][j] += weightGrads[l][i][j] / batchSize
				}
			}
		}
	}

	// Update weights
	t.UpdateWeights(accumWeightGrads, accumBiasGrads)

	return totalLoss / batchSize
}

// applyHardwareNoise adds noise during forward pass.
func (t *Trainer) applyHardwareNoise(w float64, readNoise bool) float64 {
	if !readNoise || t.config.UpdateNoise == 0 {
		return w
	}
	// Smaller noise during read than write
	noise := rand.NormFloat64() * t.config.UpdateNoise * 0.1
	return w + noise
}

// applyHardwareConstraints clips and quantizes weights.
func (t *Trainer) applyHardwareConstraints(w float64) float64 {
	// Clip to valid range
	w = clip(w, t.config.WeightClipMin, t.config.WeightClipMax)

	// Quantize if enabled
	if t.config.QuantizeBits > 0 {
		levels := float64(int(1) << uint(t.config.QuantizeBits))
		w = math.Round(w*levels) / levels
	}

	return w
}

// GetWeights returns current weights for export.
func (t *Trainer) GetWeights() [][][]float64 {
	return t.weights
}

// GetBiases returns current biases for export.
func (t *Trainer) GetBiases() [][]float64 {
	return t.biases
}

// SetWeights loads pre-trained weights.
func (t *Trainer) SetWeights(weights [][][]float64, biases [][]float64) error {
	if len(weights) != len(t.weights) {
		return fmt.Errorf("layer count mismatch")
	}
	for l := range weights {
		if len(weights[l]) != len(t.weights[l]) {
			return fmt.Errorf("layer %d output size mismatch", l)
		}
		for i := range weights[l] {
			if len(weights[l][i]) != len(t.weights[l][i]) {
				return fmt.Errorf("layer %d input size mismatch", l)
			}
			copy(t.weights[l][i], weights[l][i])
		}
		if l < len(biases) {
			copy(t.biases[l], biases[l])
		}
	}
	return nil
}

// Predict performs inference and returns predicted class.
func (t *Trainer) Predict(input []float64) int {
	output, _, _ := t.Forward(input)
	maxIdx := 0
	maxVal := output[0]
	for i, v := range output {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// Evaluate returns accuracy on a test set.
func (t *Trainer) Evaluate(inputs [][]float64, labels []int) float64 {
	correct := 0
	for i := range inputs {
		pred := t.Predict(inputs[i])
		if pred == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(inputs))
}

// Helper functions

func relu(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Max(0, v)
	}
	return result
}

func softmax(x []float64) []float64 {
	result := make([]float64, len(x))
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}
	var sum float64
	for i, v := range x {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}
	for i := range result {
		result[i] /= sum
	}
	return result
}

func clip(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// MLCProgrammer handles multi-level cell programming schemes.
type MLCProgrammer struct {
	NumLevels     int     // Number of conductance levels (e.g., 4 for 2-bit)
	PulseWidth    float64 // Base pulse width in nanoseconds
	PulseVoltage  float64 // Base programming voltage
	VerifyEnabled bool    // Enable write-verify
	MaxRetries    int     // Max verification retries
}

// NewMLCProgrammer creates a new MLC programmer.
func NewMLCProgrammer(bits int) *MLCProgrammer {
	return &MLCProgrammer{
		NumLevels:     1 << bits,
		PulseWidth:    200, // 200 ns typical
		PulseVoltage:  2.5, // 2.5V typical
		VerifyEnabled: true,
		MaxRetries:    10,
	}
}

// ComputePulseParams returns pulse parameters for target conductance.
func (p *MLCProgrammer) ComputePulseParams(currentG, targetG float64) (voltage, width float64, numPulses int) {
	// Quantize target to nearest level
	levelSize := 1.0 / float64(p.NumLevels-1)
	targetLevel := math.Round(targetG / levelSize)
	quantizedTarget := targetLevel * levelSize

	// Compute required change
	delta := quantizedTarget - currentG

	if math.Abs(delta) < levelSize*0.1 {
		return 0, 0, 0 // Already at target
	}

	// Program direction determines polarity
	if delta > 0 {
		// Potentiation: positive voltage
		voltage = p.PulseVoltage * (0.8 + 0.4*delta)
	} else {
		// Depression: negative voltage
		voltage = -p.PulseVoltage * (0.8 + 0.4*math.Abs(delta))
	}

	width = p.PulseWidth
	numPulses = int(math.Ceil(math.Abs(delta) / levelSize))

	return voltage, width, numPulses
}

// SimulateProgramming simulates MLC programming with non-idealities.
func (p *MLCProgrammer) SimulateProgramming(currentG, targetG float64) (finalG float64, success bool) {
	levelSize := 1.0 / float64(p.NumLevels-1)
	quantizedTarget := math.Round(targetG/levelSize) * levelSize
	quantizedTarget = clip(quantizedTarget, 0, 1)

	g := currentG
	for retry := 0; retry < p.MaxRetries; retry++ {
		voltage, _, numPulses := p.ComputePulseParams(g, quantizedTarget)

		if numPulses == 0 {
			return g, true
		}

		// Apply pulses with noise
		for i := 0; i < numPulses; i++ {
			deltaG := (voltage / p.PulseVoltage) * levelSize
			// Add programming noise
			deltaG *= (1.0 + rand.NormFloat64()*0.1)
			g += deltaG
			g = clip(g, 0, 1)
		}

		// Verify
		if p.VerifyEnabled {
			readG := g * (1.0 + rand.NormFloat64()*0.02) // Read noise
			if math.Abs(readG-quantizedTarget) < levelSize*0.3 {
				return g, true
			}
		} else {
			return g, true
		}
	}

	return g, false
}

// GetQuantizedLevels returns all valid quantized conductance levels.
func (p *MLCProgrammer) GetQuantizedLevels() []float64 {
	levels := make([]float64, p.NumLevels)
	for i := range levels {
		levels[i] = float64(i) / float64(p.NumLevels-1)
	}
	return levels
}
