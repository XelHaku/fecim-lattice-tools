package core

import (
	"fmt"
	"math"
)

// Infer runs dual-path inference (FP + CIM) and returns comparison results.
// Returns nil if input length doesn't match expected InputSize.
func (net *DualModeNetwork) Infer(input []float64) *InferenceResult {
	// Compute input stats for logging
	inputMin, inputMax, inputMean := 1.0, 0.0, 0.0
	for _, v := range input {
		if v < inputMin {
			inputMin = v
		}
		if v > inputMax {
			inputMax = v
		}
		inputMean += v
	}
	if len(input) > 0 {
		inputMean /= float64(len(input))
	}

	log.Input("Infer", map[string]interface{}{
		"inputLen":  len(input),
		"inputMin":  inputMin,
		"inputMax":  inputMax,
		"inputMean": inputMean,
		"levels":    net.Config.NumLevels,
		"noise":     net.Config.NoiseLevel,
		"adcBits":   net.Config.ADCBits,
		"dacBits":   net.Config.DACBits,
		"singleLyr": net.Config.SingleLayer,
	})

	net.mu.RLock()
	defer net.mu.RUnlock()

	// Validate input length
	if len(input) != net.InputSize {
		log.Error(fmt.Errorf("input length %d != expected %d", len(input), net.InputSize), "Infer")
		return nil
	}

	result := &InferenceResult{}

	var fpOutput, fpProbs []float64
	var fpHidden []float64
	var cimOutput, cimProbs []float64
	var cimHidden []float64

	if net.Config.SingleLayer {
		// ============================================
		// SINGLE-LAYER MODE: (784→10)
		// Simpler architecture for demonstration
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

	// Energy calculation uses the shared model (core energy SSOT).
	est := EstimateInferenceEnergyJ(net.Config, net.InputSize, net.HiddenSize, net.OutputSize)
	result.EnergyUsed = est.TotalJ * 1e6

	log.Calculation("Infer", map[string]interface{}{
		"fpPred":       result.FPPrediction,
		"fpConf":       result.FPConfidence,
		"cimPred":      result.CIMPrediction,
		"cimConf":      result.CIMConfidence,
		"agree":        result.Agree,
		"disagreement": result.Disagreement,
		"energy_uJ":    result.EnergyUsed,
	}, result)

	return result
}

// InferFPOnly runs only the FP path (for fast evaluation).
func (net *DualModeNetwork) InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64) {
	log.Trace("InferFPOnly: start (inputLen=%d)", len(input))

	net.mu.RLock()
	defer net.mu.RUnlock()

	hidden := net.forwardFP(input, net.FPWeights1, net.FPBias1)
	hidden = relu(hidden)
	output := net.forwardFP(hidden, net.FPWeights2, net.FPBias2)
	probs = softmax(output)
	prediction = argmax(probs)
	confidence = probs[prediction]

	log.Trace("InferFPOnly: pred=%d, conf=%.4f", prediction, confidence)
	return
}

// InferCIMOnly runs only the CIM path (for fast evaluation).
// Uses quantized weights to simulate realistic hardware behavior.
func (net *DualModeNetwork) InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64) {
	log.Trace("InferCIMOnly: start (inputLen=%d)", len(input))

	net.mu.RLock()
	defer net.mu.RUnlock()

	dacInput := quantizeDAC(input, net.Config.DACBits)

	// Layer 1: Use QUANTIZED weights (30-level FeCIM quantization)
	hidden := net.forwardCIM(dacInput, net.QuantWeights1, net.QuantBias1)
	hidden = quantizeADC(hidden, net.Config.ADCBits)
	hidden = net.safeNoise(hidden, net.Config.NoiseLevel)
	hidden = relu(hidden)

	// Layer 2: Use QUANTIZED weights (30-level FeCIM quantization)
	output := net.forwardCIM(hidden, net.QuantWeights2, net.QuantBias2)
	output = quantizeADC(output, net.Config.ADCBits)
	output = net.safeNoise(output, net.Config.NoiseLevel)
	probs = softmax(output)
	prediction = argmax(probs)
	if prediction >= 0 && prediction < len(probs) {
		confidence = probs[prediction]
	}

	log.Trace("InferCIMOnly: pred=%d, conf=%.4f", prediction, confidence)
	return
}

// forwardFP performs standard FP matrix multiplication.
// Uses GPU acceleration when available and input is large enough.
func (net *DualModeNetwork) forwardFP(input []float64, weights [][]float64, bias []float64) []float64 {
	// Try GPU path if available and input is large enough to benefit
	if net.useGPU && len(input) >= 128 {
		result, err := net.forwardFPGPU(input, weights, bias)
		if err == nil {
			log.Trace("forwardFP: GPU path used (input=%d, output=%d)", len(input), len(bias))
			return result
		}
		log.Trace("forwardFP: GPU fallback to CPU (err=%v)", err)
		// Fall back to CPU on GPU error (silent fallback)
	}

	// CPU path (original implementation)
	output := make([]float64, len(bias))

	for i := 0; i < len(weights); i++ {
		sum := bias[i]
		for j := 0; j < len(input); j++ {
			sum += weights[i][j] * input[j]
		}
		output[i] = sum
	}

	// Log activation stats
	if len(output) > 0 {
		outMin, outMax, outMean := output[0], output[0], 0.0
		for _, v := range output {
			if v < outMin {
				outMin = v
			}
			if v > outMax {
				outMax = v
			}
			outMean += v
		}
		outMean /= float64(len(output))
		log.Trace("forwardFP: CPU path (input=%d, output=%d, min=%.3f, max=%.3f, mean=%.3f)",
			len(input), len(output), outMin, outMax, outMean)
	}

	return output
}

// forwardCIM performs CIM (Compute-in-Memory) matrix multiplication using conductance-based computation.
//
// Physical Model:
// - Each quantized weight represents a discrete conductance level: G = Gmin + (Gmax-Gmin) * normalized_weight
// - Using Gmin=10µS, Gmax=100µS (from docs)
// - The math result is identical to forwardFP for inference, but models the physical process
//
// The difference from forwardFP is semantic:
// - forwardFP is called with full-precision (float64) weights
// - forwardCIM is called with quantized weights (30-level FeCIM representation)
// - forwardCIM conceptually represents conductance-based analog computation
//
// This wrapper exists for code clarity to distinguish the two inference paths.
func (net *DualModeNetwork) forwardCIM(input []float64, weights [][]float64, bias []float64) []float64 {
	// The quantized weights are already mapped to discrete levels
	// In hardware, these would be conductance levels G = Gmin + (Gmax-Gmin) * normalized_weight
	// For computational efficiency, we keep normalized values but the mapping is:
	// Gmin = 10e-6 S, Gmax = 100e-6 S
	// The multiplication and accumulation is mathematically identical to FP
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
	if len(x) == 0 {
		return nil
	}

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
// Returns -1 if the slice is empty.
func argmax(x []float64) int {
	if len(x) == 0 {
		return -1
	}

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
	if len(values) == 0 {
		return values
	}

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
