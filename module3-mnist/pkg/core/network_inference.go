package core

import (
	"math"
)

// Infer runs dual-path inference (FP + CIM) and returns comparison results.
func (net *DualModeNetwork) Infer(input []float64) *InferenceResult {
	net.mu.RLock()
	defer net.mu.RUnlock()

	result := &InferenceResult{}

	var fpOutput, fpProbs []float64
	var fpHidden []float64
	var cimOutput, cimProbs []float64
	var cimHidden []float64
	var totalMACs int

	if net.Config.SingleLayer {
		// ============================================
		// TOUR MODE: Single-Layer (784→10)
		// Matches Dr. Tour's MNIST demo (~87% theoretical max)
		// ============================================

		// FP PATH (single layer)
		fpOutput = net.forwardFP(input, net.SingleLayerWeights, net.SingleLayerBias)
		fpProbs = softmax(fpOutput)
		fpHidden = nil // No hidden layer in Tour mode

		// CIM PATH (single layer with quantization)
		dacInput := quantizeDAC(input, net.Config.DACBits)
		cimOutput = net.forwardCIM(dacInput, net.QuantSingleLayerWeights, net.QuantSingleLayerBias)
		cimOutput = quantizeADC(cimOutput, net.Config.ADCBits)
		cimOutput = net.safeNoise(cimOutput, net.Config.NoiseLevel)
		cimProbs = softmax(cimOutput)
		cimHidden = nil // No hidden layer in Tour mode

		// Energy: single layer only
		totalMACs = net.InputSize * net.OutputSize // 784 × 10 = 7,840 MACs
	} else {
		// ============================================
		// STANDARD MODE: Two-Layer (784→128→10)
		// Higher accuracy (~93%+) due to hidden layer capacity
		// ============================================

		// FP PATH (Ideal Digital)
		fpHidden = net.forwardFP(input, net.FPWeights1, net.FPBias1)
		fpHidden = relu(fpHidden)
		fpOutput = net.forwardFP(fpHidden, net.FPWeights2, net.FPBias2)
		fpProbs = softmax(fpOutput)

		// CIM PATH (Realistic Hardware)
		dacInput := quantizeDAC(input, net.Config.DACBits)

		// Layer 1: Use QUANTIZED weights (30-level FeCIM quantization)
		cimHidden = net.forwardCIM(dacInput, net.QuantWeights1, net.QuantBias1)
		cimHidden = quantizeADC(cimHidden, net.Config.ADCBits)
		cimHidden = net.safeNoise(cimHidden, net.Config.NoiseLevel)
		cimHidden = relu(cimHidden)

		// Layer 2: Use QUANTIZED weights (30-level FeCIM quantization)
		cimOutput = net.forwardCIM(cimHidden, net.QuantWeights2, net.QuantBias2)
		cimOutput = quantizeADC(cimOutput, net.Config.ADCBits)
		cimOutput = net.safeNoise(cimOutput, net.Config.NoiseLevel)
		cimProbs = softmax(cimOutput)

		// Energy: two layers
		macs1 := net.InputSize * net.HiddenSize  // 784 × 128
		macs2 := net.HiddenSize * net.OutputSize // 128 × 10
		totalMACs = macs1 + macs2
	}

	// Store FP results
	result.FPLogits = fpOutput
	result.FPProbabilities = fpProbs
	result.FPPrediction = argmax(fpProbs)
	result.FPConfidence = fpProbs[result.FPPrediction]
	result.FPHidden = fpHidden

	// Store CIM results
	result.CIMLogits = cimOutput
	result.CIMProbabilities = cimProbs
	result.CIMPrediction = argmax(cimProbs)
	result.CIMConfidence = cimProbs[result.CIMPrediction]
	result.CIMHidden = cimHidden

	// ============================================
	// COMPARISON
	// ============================================
	result.Agree = (result.FPPrediction == result.CIMPrediction)
	result.Disagreement = klDivergence(result.FPProbabilities, result.CIMProbabilities)

	// Energy calculation (Jerry et al. IEDM 2017: ~50 fJ/MAC at 30 levels)
	// Energy scales with bits of precision: ~10 fJ/bit per MAC
	// bits = log2(levels), so energy = 10 * log2(levels) fJ/MAC
	// At 30 levels (~4.9 bits): 10 * 4.9 ≈ 50 fJ/MAC (matches literature)
	// At 2 levels (1 bit): 10 * 1 = 10 fJ/MAC (5x more efficient)
	//
	// With per-layer quantization, calculate energy for each layer separately
	var totalEnergy float64
	if net.Config.SingleLayer {
		// Single layer uses Layer1Levels (or NumLevels if not per-layer)
		levels := net.Config.NumLevels
		if net.Config.PerLayerQuant {
			levels = net.Config.Layer1Levels
		}
		bitsPerCell := math.Log2(float64(levels))
		if bitsPerCell < 1 {
			bitsPerCell = 1
		}
		energyPerMAC := 10e-15 * bitsPerCell
		totalEnergy = float64(totalMACs) * energyPerMAC * 1e6
	} else {
		// Two layers: calculate energy separately
		macs1 := net.InputSize * net.HiddenSize
		macs2 := net.HiddenSize * net.OutputSize

		l1Levels := net.Config.NumLevels
		l2Levels := net.Config.NumLevels
		if net.Config.PerLayerQuant {
			l1Levels = net.Config.Layer1Levels
			l2Levels = net.Config.Layer2Levels
		}

		bits1 := math.Log2(float64(l1Levels))
		if bits1 < 1 {
			bits1 = 1
		}
		bits2 := math.Log2(float64(l2Levels))
		if bits2 < 1 {
			bits2 = 1
		}

		energy1 := float64(macs1) * 10e-15 * bits1 * 1e6 // Layer 1 energy in μJ
		energy2 := float64(macs2) * 10e-15 * bits2 * 1e6 // Layer 2 energy in μJ
		totalEnergy = energy1 + energy2
	}
	result.EnergyUsed = totalEnergy

	return result
}

// InferFPOnly runs only the FP path (for fast evaluation).
func (net *DualModeNetwork) InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	hidden := net.forwardFP(input, net.FPWeights1, net.FPBias1)
	hidden = relu(hidden)
	output := net.forwardFP(hidden, net.FPWeights2, net.FPBias2)
	probs = softmax(output)
	prediction = argmax(probs)
	confidence = probs[prediction]
	return
}

// InferCIMOnly runs only the CIM path (for fast evaluation).
func (net *DualModeNetwork) InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	dacInput := quantizeDAC(input, net.Config.DACBits)

	hidden := net.forwardCIM(dacInput, net.FPWeights1, net.FPBias1)
	hidden = quantizeADC(hidden, net.Config.ADCBits)
	hidden = net.safeNoise(hidden, net.Config.NoiseLevel)
	hidden = relu(hidden)

	output := net.forwardCIM(hidden, net.FPWeights2, net.FPBias2)
	output = quantizeADC(output, net.Config.ADCBits)
	output = net.safeNoise(output, net.Config.NoiseLevel)
	probs = softmax(output)
	prediction = argmax(probs)
	confidence = probs[prediction]
	return
}

// forwardFP performs standard FP matrix multiplication.
func (net *DualModeNetwork) forwardFP(input []float64, weights [][]float64, bias []float64) []float64 {
	output := make([]float64, len(bias))

	for i := 0; i < len(weights); i++ {
		sum := bias[i]
		for j := 0; j < len(input); j++ {
			sum += weights[i][j] * input[j]
		}
		output[i] = sum
	}

	return output
}

// forwardCIM performs CIM matrix multiplication (same math, but uses quantized weights).
func (net *DualModeNetwork) forwardCIM(input []float64, weights [][]float64, bias []float64) []float64 {
	return net.forwardFP(input, weights, bias)
}

// relu applies ReLU activation.
func relu(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		if v > 0 {
			result[i] = v
		}
	}
	return result
}

// softmax applies softmax activation.
func softmax(x []float64) []float64 {
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	expSum := 0.0
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Exp(v - max)
		expSum += result[i]
	}

	for i := range result {
		result[i] /= expSum
	}

	return result
}

// argmax returns the index of the maximum value.
func argmax(x []float64) int {
	maxIdx := 0
	maxVal := x[0]
	for i, v := range x {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// klDivergence computes KL divergence D(p||q).
func klDivergence(p, q []float64) float64 {
	kl := 0.0
	eps := 1e-10
	for i := range p {
		if p[i] > eps {
			kl += p[i] * math.Log(p[i]/(q[i]+eps))
		}
	}
	return kl
}

// quantizeDAC simulates N-bit DAC quantization of input voltages.
func quantizeDAC(values []float64, bits int) []float64 {
	if bits >= 16 {
		return values // No quantization
	}

	levels := 1 << bits // 2^bits
	result := make([]float64, len(values))

	for i, v := range values {
		// Input values are assumed to be in [0, 1]
		// Clamp and quantize
		v = math.Max(0, math.Min(1, v))
		bin := int(math.Round(v * float64(levels-1)))
		if bin >= levels {
			bin = levels - 1
		}
		result[i] = float64(bin) / float64(levels-1)
	}

	return result
}

// quantizeADC simulates N-bit ADC quantization of output currents.
func quantizeADC(values []float64, bits int) []float64 {
	if bits >= 16 {
		return values // No quantization
	}

	levels := 1 << bits // 2^bits

	// Find range
	vMin, vMax := values[0], values[0]
	for _, v := range values {
		if v < vMin {
			vMin = v
		}
		if v > vMax {
			vMax = v
		}
	}

	vRange := vMax - vMin
	if vRange == 0 {
		return values
	}

	step := vRange / float64(levels-1)

	result := make([]float64, len(values))
	for i, v := range values {
		bin := int(math.Round((v - vMin) / step))
		if bin < 0 {
			bin = 0
		}
		if bin >= levels {
			bin = levels - 1
		}
		result[i] = vMin + float64(bin)*step
	}

	return result
}
