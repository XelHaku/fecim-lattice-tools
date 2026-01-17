// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"math"
	"math/rand"
)

// Dropout implements standard dropout regularization.
// During training, randomly zeroes elements with probability p.
// During inference, passes input unchanged.
type Dropout struct {
	Rate     float64 // Dropout probability (0.0 - 1.0)
	Training bool    // Training mode flag
}

// NewDropout creates a new dropout layer.
func NewDropout(rate float64) *Dropout {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Dropout{
		Rate:     rate,
		Training: false,
	}
}

// Forward applies dropout to 1D input.
func (d *Dropout) Forward(input []float64) []float64 {
	if !d.Training || d.Rate == 0 {
		return input
	}

	output := make([]float64, len(input))
	scale := 1.0 / (1.0 - d.Rate)

	for i := range output {
		if rand.Float64() >= d.Rate {
			output[i] = input[i] * scale
		}
	}

	return output
}

// Forward2D applies dropout to 2D input.
func (d *Dropout) Forward2D(input [][]float64) [][]float64 {
	if !d.Training || d.Rate == 0 {
		return input
	}

	output := make([][]float64, len(input))
	scale := 1.0 / (1.0 - d.Rate)

	for i := range output {
		output[i] = make([]float64, len(input[i]))
		for j := range output[i] {
			if rand.Float64() >= d.Rate {
				output[i][j] = input[i][j] * scale
			}
		}
	}

	return output
}

// Forward3D applies dropout to 3D input (e.g., feature maps).
func (d *Dropout) Forward3D(input [][][]float64) [][][]float64 {
	if !d.Training || d.Rate == 0 {
		return input
	}

	output := make([][][]float64, len(input))
	scale := 1.0 / (1.0 - d.Rate)

	for c := range output {
		output[c] = make([][]float64, len(input[c]))
		for h := range output[c] {
			output[c][h] = make([]float64, len(input[c][h]))
			for w := range output[c][h] {
				if rand.Float64() >= d.Rate {
					output[c][h][w] = input[c][h][w] * scale
				}
			}
		}
	}

	return output
}

// DropoutMask generates a dropout mask for later reuse (e.g., in backward pass).
func (d *Dropout) DropoutMask(shape []int) []float64 {
	size := 1
	for _, s := range shape {
		size *= s
	}

	mask := make([]float64, size)
	scale := 1.0 / (1.0 - d.Rate)

	for i := range mask {
		if rand.Float64() >= d.Rate {
			mask[i] = scale
		}
	}

	return mask
}

// GaussianNoise injects Gaussian noise during training.
// Acts as regularization, similar to weight noise in analog hardware.
// This simulates the inherent noise in analog CIM devices.
type GaussianNoise struct {
	Stddev   float64 // Standard deviation of noise
	Training bool    // Training mode flag
}

// NewGaussianNoise creates a new Gaussian noise layer.
func NewGaussianNoise(stddev float64) *GaussianNoise {
	return &GaussianNoise{
		Stddev:   stddev,
		Training: false,
	}
}

// Forward applies Gaussian noise to 1D input.
func (gn *GaussianNoise) Forward(input []float64) []float64 {
	if !gn.Training || gn.Stddev == 0 {
		return input
	}

	output := make([]float64, len(input))
	for i := range output {
		output[i] = input[i] + rand.NormFloat64()*gn.Stddev
	}

	return output
}

// Forward2D applies Gaussian noise to 2D input.
func (gn *GaussianNoise) Forward2D(input [][]float64) [][]float64 {
	if !gn.Training || gn.Stddev == 0 {
		return input
	}

	output := make([][]float64, len(input))
	for i := range output {
		output[i] = make([]float64, len(input[i]))
		for j := range output[i] {
			output[i][j] = input[i][j] + rand.NormFloat64()*gn.Stddev
		}
	}

	return output
}

// MultiplicativeNoise applies multiplicative Gaussian noise.
// More realistic model for analog device variation.
// Output = input * (1 + N(0, stddev))
type MultiplicativeNoise struct {
	Stddev   float64
	Training bool
}

// NewMultiplicativeNoise creates multiplicative noise layer.
func NewMultiplicativeNoise(stddev float64) *MultiplicativeNoise {
	return &MultiplicativeNoise{
		Stddev:   stddev,
		Training: false,
	}
}

// Forward applies multiplicative noise.
func (mn *MultiplicativeNoise) Forward(input []float64) []float64 {
	if !mn.Training || mn.Stddev == 0 {
		return input
	}

	output := make([]float64, len(input))
	for i := range output {
		output[i] = input[i] * (1 + rand.NormFloat64()*mn.Stddev)
	}

	return output
}

// Forward2D applies multiplicative noise to 2D input.
func (mn *MultiplicativeNoise) Forward2D(input [][]float64) [][]float64 {
	if !mn.Training || mn.Stddev == 0 {
		return input
	}

	output := make([][]float64, len(input))
	for i := range output {
		output[i] = make([]float64, len(input[i]))
		for j := range output[i] {
			output[i][j] = input[i][j] * (1 + rand.NormFloat64()*mn.Stddev)
		}
	}

	return output
}

// WeightRegularization provides L1, L2, and elastic net regularization.
type WeightRegularization struct {
	L1Lambda float64 // L1 regularization coefficient (sparsity)
	L2Lambda float64 // L2 regularization coefficient (weight decay)
}

// NewL1Regularization creates L1 (Lasso) regularization.
func NewL1Regularization(lambda float64) *WeightRegularization {
	return &WeightRegularization{
		L1Lambda: lambda,
		L2Lambda: 0,
	}
}

// NewL2Regularization creates L2 (Ridge) regularization.
func NewL2Regularization(lambda float64) *WeightRegularization {
	return &WeightRegularization{
		L1Lambda: 0,
		L2Lambda: lambda,
	}
}

// NewElasticNet creates elastic net regularization (L1 + L2).
func NewElasticNet(l1Lambda, l2Lambda float64) *WeightRegularization {
	return &WeightRegularization{
		L1Lambda: l1Lambda,
		L2Lambda: l2Lambda,
	}
}

// ComputeLoss computes regularization loss for given weights.
func (wr *WeightRegularization) ComputeLoss(weights [][]float64) float64 {
	var l1Loss, l2Loss float64

	for i := range weights {
		for j := range weights[i] {
			w := weights[i][j]
			l1Loss += math.Abs(w)
			l2Loss += w * w
		}
	}

	return wr.L1Lambda*l1Loss + 0.5*wr.L2Lambda*l2Loss
}

// ComputeGradient computes regularization gradient for weight update.
// Returns gradient to be added to main gradient.
func (wr *WeightRegularization) ComputeGradient(weights [][]float64) [][]float64 {
	grad := make([][]float64, len(weights))

	for i := range grad {
		grad[i] = make([]float64, len(weights[i]))
		for j := range grad[i] {
			w := weights[i][j]

			// L1 gradient: sign(w)
			l1Grad := 0.0
			if w > 0 {
				l1Grad = 1.0
			} else if w < 0 {
				l1Grad = -1.0
			}

			// L2 gradient: w
			l2Grad := w

			grad[i][j] = wr.L1Lambda*l1Grad + wr.L2Lambda*l2Grad
		}
	}

	return grad
}

// ApplyWeightDecay applies weight decay directly to weights.
// w_new = w * (1 - lambda * lr)
func (wr *WeightRegularization) ApplyWeightDecay(weights [][]float64, learningRate float64) {
	decay := 1.0 - wr.L2Lambda*learningRate

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] *= decay
		}
	}
}

// AnalogNoiseModel simulates realistic analog CIM device noise.
// Combines multiple noise sources found in FeFET/memristor arrays.
type AnalogNoiseModel struct {
	// Device-to-device variation (spatial)
	D2DVariation float64

	// Cycle-to-cycle variation (temporal)
	C2CVariation float64

	// Read noise (ADC/peripheral)
	ReadNoise float64

	// IR drop effect (position-dependent)
	IRDropCoeff float64

	// Temperature coefficient
	TempCoeff float64

	Training bool
}

// NewAnalogNoiseModel creates realistic analog noise model.
func NewAnalogNoiseModel() *AnalogNoiseModel {
	return &AnalogNoiseModel{
		D2DVariation: 0.05, // 5% device-to-device
		C2CVariation: 0.02, // 2% cycle-to-cycle
		ReadNoise:    0.01, // 1% read noise
		IRDropCoeff:  0.001,
		TempCoeff:    0.0,
		Training:     false,
	}
}

// ApplyNoise applies analog noise model to conductance values.
// Simulates realistic crossbar array non-idealities.
func (anm *AnalogNoiseModel) ApplyNoise(conductances [][]float64) [][]float64 {
	rows := len(conductances)
	cols := len(conductances[0])

	output := make([][]float64, rows)
	for i := range output {
		output[i] = make([]float64, cols)
		for j := range output[i] {
			g := conductances[i][j]

			// Device-to-device variation (fixed per device)
			d2dNoise := rand.NormFloat64() * anm.D2DVariation * math.Abs(g)

			// Cycle-to-cycle variation (varies each operation)
			c2cNoise := rand.NormFloat64() * anm.C2CVariation * math.Abs(g)

			// Read noise
			readNoise := rand.NormFloat64() * anm.ReadNoise * math.Abs(g)

			// IR drop (increases with distance from drivers)
			irDrop := anm.IRDropCoeff * float64(i+j) * math.Abs(g)

			output[i][j] = g + d2dNoise + c2cNoise + readNoise - irDrop
		}
	}

	return output
}

// ApplyOutputNoise applies noise to crossbar output (after MVM).
func (anm *AnalogNoiseModel) ApplyOutputNoise(output []float64) []float64 {
	result := make([]float64, len(output))
	for i := range result {
		noise := rand.NormFloat64() * anm.ReadNoise * math.Abs(output[i])
		result[i] = output[i] + noise
	}
	return result
}

// VariationAwareTraining implements variation-aware training for analog CIM.
// During training, applies noise to simulate device variations.
// Based on research showing noise injection improves hardware accuracy.
type VariationAwareTraining struct {
	NoiseModel  *AnalogNoiseModel
	NumSamples  int     // Number of noise samples per forward pass
	Temperature float64 // Temperature for noise scheduling
}

// NewVariationAwareTraining creates variation-aware training wrapper.
func NewVariationAwareTraining() *VariationAwareTraining {
	return &VariationAwareTraining{
		NoiseModel:  NewAnalogNoiseModel(),
		NumSamples:  1,
		Temperature: 1.0,
	}
}

// ForwardWithNoise performs forward pass with noise injection.
// Returns average output over multiple noise samples.
func (vat *VariationAwareTraining) ForwardWithNoise(
	weights [][]float64,
	input []float64,
	forward func([][]float64, []float64) []float64,
) []float64 {
	if vat.NumSamples <= 1 {
		noisyWeights := vat.NoiseModel.ApplyNoise(weights)
		return forward(noisyWeights, input)
	}

	// Multiple samples for more stable gradient estimation
	var outputs [][]float64
	for s := 0; s < vat.NumSamples; s++ {
		noisyWeights := vat.NoiseModel.ApplyNoise(weights)
		outputs = append(outputs, forward(noisyWeights, input))
	}

	// Average outputs
	result := make([]float64, len(outputs[0]))
	for _, out := range outputs {
		for i := range result {
			result[i] += out[i]
		}
	}
	for i := range result {
		result[i] /= float64(vat.NumSamples)
	}

	return result
}

// SpatialDropout implements spatial dropout for CNNs.
// Drops entire feature maps instead of individual elements.
// Better preserves spatial structure.
type SpatialDropout struct {
	Rate     float64
	Training bool
}

// NewSpatialDropout creates spatial dropout layer.
func NewSpatialDropout(rate float64) *SpatialDropout {
	return &SpatialDropout{
		Rate:     rate,
		Training: false,
	}
}

// Forward applies spatial dropout to 3D input [channels][height][width].
func (sd *SpatialDropout) Forward(input [][][]float64) [][][]float64 {
	if !sd.Training || sd.Rate == 0 {
		return input
	}

	numChannels := len(input)
	output := make([][][]float64, numChannels)
	scale := 1.0 / (1.0 - sd.Rate)

	for c := range output {
		if rand.Float64() < sd.Rate {
			// Drop entire channel
			output[c] = make([][]float64, len(input[c]))
			for h := range output[c] {
				output[c][h] = make([]float64, len(input[c][h]))
			}
		} else {
			// Keep and scale channel
			output[c] = make([][]float64, len(input[c]))
			for h := range output[c] {
				output[c][h] = make([]float64, len(input[c][h]))
				for w := range output[c][h] {
					output[c][h][w] = input[c][h][w] * scale
				}
			}
		}
	}

	return output
}

// DropConnect implements DropConnect regularization.
// Drops weights instead of activations.
// More suitable for analog CIM as it simulates weight failures.
type DropConnect struct {
	Rate     float64
	Training bool
}

// NewDropConnect creates DropConnect layer.
func NewDropConnect(rate float64) *DropConnect {
	return &DropConnect{
		Rate:     rate,
		Training: false,
	}
}

// ApplyToWeights applies DropConnect to weight matrix.
func (dc *DropConnect) ApplyToWeights(weights [][]float64) [][]float64 {
	if !dc.Training || dc.Rate == 0 {
		return weights
	}

	output := make([][]float64, len(weights))
	scale := 1.0 / (1.0 - dc.Rate)

	for i := range output {
		output[i] = make([]float64, len(weights[i]))
		for j := range output[i] {
			if rand.Float64() >= dc.Rate {
				output[i][j] = weights[i][j] * scale
			}
		}
	}

	return output
}

// LabelSmoothing implements label smoothing regularization.
// Prevents overconfident predictions.
type LabelSmoothing struct {
	Smoothing  float64 // Smoothing factor (typically 0.1)
	NumClasses int
}

// NewLabelSmoothing creates label smoothing layer.
func NewLabelSmoothing(numClasses int, smoothing float64) *LabelSmoothing {
	return &LabelSmoothing{
		Smoothing:  smoothing,
		NumClasses: numClasses,
	}
}

// SmoothLabels converts hard labels to soft labels.
// Hard label [0,0,1,0] with smoothing=0.1 becomes [0.025,0.025,0.925,0.025]
func (ls *LabelSmoothing) SmoothLabels(labels []float64) []float64 {
	result := make([]float64, len(labels))
	uniform := ls.Smoothing / float64(ls.NumClasses)

	for i := range result {
		result[i] = labels[i]*(1-ls.Smoothing) + uniform
	}

	return result
}

// SmoothLabelsBatch applies label smoothing to batch.
func (ls *LabelSmoothing) SmoothLabelsBatch(labels [][]float64) [][]float64 {
	result := make([][]float64, len(labels))
	for i := range result {
		result[i] = ls.SmoothLabels(labels[i])
	}
	return result
}
