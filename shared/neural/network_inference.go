package neural

import (
	"fmt"
	"math"
	"sync"
)

var cimSemanticNoticeOnce sync.Once

type inferScratch struct {
	fpHidden  []float64
	fpOutput  []float64
	fpProbs   []float64
	cimInput  []float64
	cimHidden []float64
	cimOutput []float64
	cimProbs  []float64
}

func ensureLen(buf []float64, n int) []float64 {
	if cap(buf) < n {
		return make([]float64, n)
	}
	return buf[:n]
}

// Infer runs dual-path inference (FP + CIM) and returns comparison results.
// Returns nil if input length doesn't match expected InputSize.
func (net *DualModeNetwork) Infer(input []float64) *InferenceResult {
	net.mu.RLock()
	defer net.mu.RUnlock()

	if len(input) != net.InputSize {
		log.Error(fmt.Errorf("input length %d != expected %d", len(input), net.InputSize), "Infer")
		return nil
	}

	s := net.scratchPool.Get().(*inferScratch)
	defer net.scratchPool.Put(s)
	cimReadLatencyS := 0.0

	if net.Config.SingleLayer {
		s.fpOutput = ensureLen(s.fpOutput, len(net.SingleLayerBias))
		net.forwardFPInto(input, net.SingleLayerWeights, net.SingleLayerBias, s.fpOutput)
		s.fpProbs = softmaxInto(s.fpProbs, s.fpOutput)

		s.cimInput = quantizeDACInto(s.cimInput, input, net.Config.DACBits)
		s.cimOutput = ensureLen(s.cimOutput, len(net.QuantSingleLayerBias))
		net.forwardCIMConductanceInto(s.cimInput, net.QuantSingleLayerWeights, net.QuantSingleLayerBias, s.cimOutput)
		cimReadLatencyS += net.adcReadLatencySecondsLocked(len(net.QuantSingleLayerWeights))
		s.cimOutput = quantizeADCInto(s.cimOutput, s.cimOutput, net.Config.ADCBits)
		s.cimOutput = net.safeNoise(s.cimOutput, net.Config.NoiseLevel)
		s.cimProbs = softmaxInto(s.cimProbs, s.cimOutput)
	} else {
		s.fpHidden = ensureLen(s.fpHidden, len(net.FPBias1))
		net.forwardFPInto(input, net.FPWeights1, net.FPBias1, s.fpHidden)
		reluInPlace(s.fpHidden)
		s.fpOutput = ensureLen(s.fpOutput, len(net.FPBias2))
		net.forwardFPInto(s.fpHidden, net.FPWeights2, net.FPBias2, s.fpOutput)
		s.fpProbs = softmaxInto(s.fpProbs, s.fpOutput)

		s.cimInput = quantizeDACInto(s.cimInput, input, net.Config.DACBits)
		s.cimHidden = ensureLen(s.cimHidden, len(net.QuantBias1))
		net.forwardCIMConductanceInto(s.cimInput, net.QuantWeights1, net.QuantBias1, s.cimHidden)
		cimReadLatencyS += net.adcReadLatencySecondsLocked(len(net.QuantWeights1))
		s.cimHidden = quantizeADCInto(s.cimHidden, s.cimHidden, net.Config.ADCBits)
		s.cimHidden = net.safeNoise(s.cimHidden, net.Config.NoiseLevel)
		reluInPlace(s.cimHidden)

		s.cimOutput = ensureLen(s.cimOutput, len(net.QuantBias2))
		net.forwardCIMConductanceInto(s.cimHidden, net.QuantWeights2, net.QuantBias2, s.cimOutput)
		cimReadLatencyS += net.adcReadLatencySecondsLocked(len(net.QuantWeights2))
		s.cimOutput = quantizeADCInto(s.cimOutput, s.cimOutput, net.Config.ADCBits)
		s.cimOutput = net.safeNoise(s.cimOutput, net.Config.NoiseLevel)
		s.cimProbs = softmaxInto(s.cimProbs, s.cimOutput)
	}

	result := &InferenceResult{
		FPLogits:         append([]float64(nil), s.fpOutput...),
		FPProbabilities:  append([]float64(nil), s.fpProbs...),
		CIMLogits:        append([]float64(nil), s.cimOutput...),
		CIMProbabilities: append([]float64(nil), s.cimProbs...),
		ReadLatencyUS:    cimReadLatencyS * 1e6,
		FPHidden:         nil,
		CIMHidden:        nil,
	}
	if !net.Config.SingleLayer {
		result.FPHidden = append([]float64(nil), s.fpHidden...)
		result.CIMHidden = append([]float64(nil), s.cimHidden...)
	}

	result.FPPrediction = argmax(result.FPProbabilities)
	result.FPConfidence = result.FPProbabilities[result.FPPrediction]
	result.CIMPrediction = argmax(result.CIMProbabilities)
	result.CIMConfidence = result.CIMProbabilities[result.CIMPrediction]
	result.Agree = (result.FPPrediction == result.CIMPrediction)
	result.Disagreement = klDivergence(result.FPProbabilities, result.CIMProbabilities)

	est := EstimateInferenceEnergyJ(net.Config, net.InputSize, net.HiddenSize, net.OutputSize)
	result.EnergyUsed = est.TotalJ * 1e6

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
	if net.useGPU && len(input) >= 128 {
		result, err := net.forwardFPGPU(input, weights, bias)
		if err == nil {
			log.Trace("forwardFP: GPU path used (input=%d, output=%d)", len(input), len(bias))
			return result
		}
		log.Trace("forwardFP: GPU fallback to CPU (err=%v)", err)
		net.emitNotification(fmt.Sprintf("GPU inference failed (%v). Falling back to CPU.", err))
	}

	output := make([]float64, len(bias))
	net.forwardFPInto(input, weights, bias, output)
	return output
}

func (net *DualModeNetwork) forwardFPInto(input []float64, weights [][]float64, bias []float64, output []float64) {
	for i := 0; i < len(weights); i++ {
		sum := bias[i]
		for j := 0; j < len(input); j++ {
			sum += weights[i][j] * input[j]
		}
		output[i] = sum
	}
}

// forwardCIM performs conductance-domain CIM MVM using a differential-pair map.
func (net *DualModeNetwork) forwardCIM(input []float64, weights [][]float64, bias []float64) []float64 {
	cimSemanticNoticeOnce.Do(func() {
		log.Info("forwardCIM uses conductance-domain differential MVM (G+/G-) with quantization and peripheral constraints")
	})
	return net.forwardCIMConductance(input, weights, bias)
}

// relu applies ReLU activation.
func relu(x []float64) []float64 {
	result := append([]float64(nil), x...)
	reluInPlace(result)
	return result
}

func reluInPlace(x []float64) {
	for i := range x {
		if x[i] < 0 {
			x[i] = 0
		}
	}
}

// softmax applies softmax activation.
func softmax(x []float64) []float64 {
	return softmaxInto(nil, x)
}

func softmaxInto(dst []float64, x []float64) []float64 {
	if len(x) == 0 {
		return nil
	}
	dst = ensureLen(dst, len(x))

	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	expSum := 0.0
	for i, v := range x {
		e := math.Exp(v - max)
		dst[i] = e
		expSum += e
	}

	inv := 1.0 / expSum
	for i := range dst {
		dst[i] *= inv
	}

	return dst
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
	return quantizeDACInto(nil, values, bits)
}

func quantizeDACInto(dst []float64, values []float64, bits int) []float64 {
	if bits >= 16 {
		return values
	}

	levels := 1 << bits
	dst = ensureLen(dst, len(values))

	outOfRangeCount := 0
	minInput := math.Inf(1)
	maxInput := math.Inf(-1)

	for i, v := range values {
		if v < minInput {
			minInput = v
		}
		if v > maxInput {
			maxInput = v
		}
		if v < 0 || v > 1 {
			outOfRangeCount++
		}
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		bin := int(math.Round(v * float64(levels-1)))
		if bin >= levels {
			bin = levels - 1
		}
		dst[i] = float64(bin) / float64(levels-1)
	}

	if outOfRangeCount > 0 {
		log.Warn("quantizeDAC: clamped %d/%d invalid input values outside [0,1] (min=%.6f, max=%.6f)",
			outOfRangeCount, len(values), minInput, maxInput)
	}

	return dst
}

// quantizeADC simulates N-bit ADC quantization of output currents.
func quantizeADC(values []float64, bits int) []float64 {
	return quantizeADCInto(nil, values, bits)
}

func quantizeADCInto(dst []float64, values []float64, bits int) []float64 {
	if len(values) == 0 {
		return values
	}
	if bits >= 16 {
		return values
	}

	levels := 1 << bits
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
	dst = ensureLen(dst, len(values))
	for i, v := range values {
		bin := int(math.Round((v - vMin) / step))
		if bin < 0 {
			bin = 0
		}
		if bin >= levels {
			bin = levels - 1
		}
		dst[i] = vMin + float64(bin)*step
	}

	return dst
}
