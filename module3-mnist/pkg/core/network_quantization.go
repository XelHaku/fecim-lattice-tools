package core

// RequantizeWeights re-quantizes the FP weights to the current number of levels.
// Call this after changing Config.NumLevels.
func (net *DualModeNetwork) RequantizeWeights() {
	log.Debug("RequantizeWeights called")
	net.mu.Lock()
	defer net.mu.Unlock()
	net.requantizeWeightsLocked()
}

// requantizeWeightsLocked performs quantization (must hold lock).
func (net *DualModeNetwork) requantizeWeightsLocked() {
	// Determine quantization levels for each layer
	var l1Levels, l2Levels int

	if net.Config.PerLayerQuant {
		// Use per-layer quantization levels
		l1Levels = net.Config.Layer1Levels
		l2Levels = net.Config.Layer2Levels
	} else {
		// Use uniform quantization levels
		l1Levels = net.Config.NumLevels
		l2Levels = net.Config.NumLevels
	}

	log.Trace("requantizeWeightsLocked: L1=%d, L2=%d, perLayer=%v",
		l1Levels, l2Levels, net.Config.PerLayerQuant)

	// Clamp levels to valid range [2, MaxDemoLevels]
	if l1Levels < 2 {
		l1Levels = 2
	}
	if l1Levels > MaxDemoLevels {
		l1Levels = MaxDemoLevels
	}
	if l2Levels < 2 {
		l2Levels = 2
	}
	if l2Levels > MaxDemoLevels {
		l2Levels = MaxDemoLevels
	}

	// Quantize layer 1 weights with layer1 levels
	// Levels are clamped to [2, 31] above, so these cannot fail
	net.QuantWeights1, _ = QuantizeWeights(net.FPWeights1, l1Levels)

	// Quantize layer 2 weights with layer2 levels
	net.QuantWeights2, _ = QuantizeWeights(net.FPWeights2, l2Levels)

	// Quantize single-layer weights (Calibration Mode) - use l1 levels for input layer
	if len(net.SingleLayerWeights) > 0 {
		net.QuantSingleLayerWeights, _ = QuantizeWeights(net.SingleLayerWeights, l1Levels)
		net.QuantSingleLayerBias, _ = QuantizeBias(net.SingleLayerBias, l1Levels)
	}

	// Quantize biases with corresponding layer levels
	net.QuantBias1, _ = QuantizeBias(net.FPBias1, l1Levels)
	net.QuantBias2, _ = QuantizeBias(net.FPBias2, l2Levels)
}

// GetQuantizationStats returns statistics about the current weight quantization.
func (net *DualModeNetwork) GetQuantizationStats() (layer1Stats, layer2Stats QuantizationStats) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	layer1Stats = ComputeQuantizationStats(net.FPWeights1, net.QuantWeights1)
	layer2Stats = ComputeQuantizationStats(net.FPWeights2, net.QuantWeights2)
	return
}

// GetPerLayerQuantInfo returns the current per-layer quantization configuration.
func (net *DualModeNetwork) GetPerLayerQuantInfo() (enabled bool, l1Levels, l2Levels int) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	enabled = net.Config.PerLayerQuant
	if enabled {
		l1Levels = net.Config.Layer1Levels
		l2Levels = net.Config.Layer2Levels
	} else {
		l1Levels = net.Config.NumLevels
		l2Levels = net.Config.NumLevels
	}
	return
}

// GetFPWeights returns a copy of the FP weights for visualization.
func (net *DualModeNetwork) GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	w1 = copyMatrix(net.FPWeights1)
	w2 = copyMatrix(net.FPWeights2)
	b1 = copySlice(net.FPBias1)
	b2 = copySlice(net.FPBias2)
	return
}

// GetQuantWeights returns a copy of the quantized weights for visualization.
func (net *DualModeNetwork) GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	w1 = copyMatrix(net.QuantWeights1)
	w2 = copyMatrix(net.QuantWeights2)
	b1 = copySlice(net.QuantBias1)
	b2 = copySlice(net.QuantBias2)
	return
}

func copyMatrix(m [][]float64) [][]float64 {
	result := make([][]float64, len(m))
	for i := range m {
		result[i] = make([]float64, len(m[i]))
		copy(result[i], m[i])
	}
	return result
}

func copySlice(s []float64) []float64 {
	result := make([]float64, len(s))
	copy(result, s)
	return result
}
