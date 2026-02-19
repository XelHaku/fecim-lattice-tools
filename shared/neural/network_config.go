package neural

import "fmt"

func formatLevelClampMessage(source string, requested, actual int) string {
	return fmt.Sprintf("%s requested %d levels, clamped to %d", source, requested, actual)
}

// SetNumLevels updates the quantization levels and re-quantizes weights.
// Minimum is 2 levels (required by QuantizeWeights), maximum is MaxDemoLevels.
func (net *DualModeNetwork) SetNumLevels(levels int) {
	requested := levels
	clamped := false

	net.mu.Lock()
	oldLevels := net.Config.NumLevels
	if levels < 2 {
		levels = 2
		clamped = true
	}
	if levels > MaxDemoLevels {
		levels = MaxDemoLevels
		clamped = true
	}
	net.Config.NumLevels = levels
	net.requantizeWeightsLocked()
	net.mu.Unlock()

	if clamped {
		net.emitNotification(formatLevelClampMessage("SetNumLevels", requested, levels))
	}
	if levels > FeCIMLevels {
		log.Info("Using overspec quantization level %d (FeCIM max %d) for comparison only", levels, FeCIMLevels)
	}
	log.Debug("SetNumLevels: %d -> %d", oldLevels, levels)
}

// GetNumLevels returns the current quantization levels.
func (net *DualModeNetwork) GetNumLevels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.NumLevels
}

// SetNoiseLevel updates the noise level for CIM inference.
// Noise is clamped to [0.0, 0.20] to match the GUI control range and docs.
func (net *DualModeNetwork) SetNoiseLevel(noise float64) {
	oldNoise := net.Config.NoiseLevel
	net.mu.Lock()
	defer net.mu.Unlock()

	if noise < 0 {
		noise = 0
	}
	if noise > 0.20 {
		noise = 0.20
	}
	net.Config.NoiseLevel = noise

	log.Debug("SetNoiseLevel: %.4f -> %.4f", oldNoise, noise)
}

// SetADCBits updates the ADC resolution for CIM inference.
func (net *DualModeNetwork) SetADCBits(bits int) {
	oldBits := net.Config.ADCBits
	net.mu.Lock()
	defer net.mu.Unlock()

	if bits < 3 {
		bits = 3
	}
	if bits > 16 {
		bits = 16
	}
	net.Config.ADCBits = bits

	log.Debug("SetADCBits: %d -> %d", oldBits, bits)
}

// SetDACBits updates the DAC resolution for CIM inference.
func (net *DualModeNetwork) SetDACBits(bits int) {
	oldBits := net.Config.DACBits
	net.mu.Lock()
	defer net.mu.Unlock()

	if bits < 3 {
		bits = 3
	}
	if bits > 16 {
		bits = 16
	}
	net.Config.DACBits = bits

	log.Debug("SetDACBits: %d -> %d", oldBits, bits)
}

// SetSingleLayer enables/disables Calibration Mode (single-layer 784→10 architecture).
// When enabled, this matches the hardware MNIST demo.
func (net *DualModeNetwork) SetSingleLayer(enabled bool) {
	oldValue := net.Config.SingleLayer
	net.mu.Lock()
	defer net.mu.Unlock()
	net.Config.SingleLayer = enabled

	log.Debug("SetSingleLayer: %v -> %v", oldValue, enabled)
}

// SetPerLayerQuant enables/disables per-layer quantization mode.
// When enabled, Layer1Levels and Layer2Levels are used instead of NumLevels.
func (net *DualModeNetwork) SetPerLayerQuant(enabled bool) {
	oldValue := net.Config.PerLayerQuant
	net.mu.Lock()
	defer net.mu.Unlock()
	net.Config.PerLayerQuant = enabled
	net.requantizeWeightsLocked()

	log.Debug("SetPerLayerQuant: %v -> %v (L1=%d, L2=%d)",
		oldValue, enabled, net.Config.Layer1Levels, net.Config.Layer2Levels)
}

// IsPerLayerQuant returns whether per-layer quantization is enabled.
func (net *DualModeNetwork) IsPerLayerQuant() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.PerLayerQuant
}

// SetLayer1Levels sets the quantization levels for layer 1 (hidden layer).
func (net *DualModeNetwork) SetLayer1Levels(levels int) {
	oldLevels := net.Config.Layer1Levels
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 2 {
		levels = 2
	}
	if levels > MaxDemoLevels {
		levels = MaxDemoLevels
	}
	net.Config.Layer1Levels = levels
	if net.Config.PerLayerQuant {
		net.requantizeWeightsLocked()
		if levels > FeCIMLevels {
			log.Info("Using overspec L1 quantization level %d (FeCIM max %d) for comparison only", levels, FeCIMLevels)
		}
		log.Debug("SetLayer1Levels: %d -> %d (requantized)", oldLevels, levels)
	}
}

// GetLayer1Levels returns the current layer 1 quantization levels.
func (net *DualModeNetwork) GetLayer1Levels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.Layer1Levels
}

// SetLayer2Levels sets the quantization levels for layer 2 (output layer).
func (net *DualModeNetwork) SetLayer2Levels(levels int) {
	oldLevels := net.Config.Layer2Levels
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 2 {
		levels = 2
	}
	if levels > MaxDemoLevels {
		levels = MaxDemoLevels
	}
	net.Config.Layer2Levels = levels
	if net.Config.PerLayerQuant {
		net.requantizeWeightsLocked()
		if levels > FeCIMLevels {
			log.Info("Using overspec L2 quantization level %d (FeCIM max %d) for comparison only", levels, FeCIMLevels)
		}
		log.Debug("SetLayer2Levels: %d -> %d (requantized)", oldLevels, levels)
	}
}

// GetLayer2Levels returns the current layer 2 quantization levels.
func (net *DualModeNetwork) GetLayer2Levels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.Layer2Levels
}

// SetPerLayerLevels sets quantization levels for both layers at once.
func (net *DualModeNetwork) SetPerLayerLevels(layer1, layer2 int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if layer1 < 2 {
		layer1 = 2
	}
	if layer1 > MaxDemoLevels {
		layer1 = MaxDemoLevels
	}
	if layer2 < 2 {
		layer2 = 2
	}
	if layer2 > MaxDemoLevels {
		layer2 = MaxDemoLevels
	}

	net.Config.Layer1Levels = layer1
	net.Config.Layer2Levels = layer2
	net.Config.PerLayerQuant = true
	net.requantizeWeightsLocked()

	if layer1 > FeCIMLevels || layer2 > FeCIMLevels {
		log.Info("Using overspec per-layer levels L1=%d L2=%d (FeCIM max %d) for comparison only", layer1, layer2, FeCIMLevels)
	}
	log.Debug("SetPerLayerLevels: L1=%d, L2=%d (enabled per-layer quant)", layer1, layer2)
}

// IsSingleLayer returns whether single-layer (Calibration Mode) is enabled.
func (net *DualModeNetwork) IsSingleLayer() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.SingleLayer
}

// safeNoise applies Gaussian noise with thread-safe RNG access.
// This prevents data races when multiple goroutines hold RLock on the network.
func (net *DualModeNetwork) safeNoise(data []float64, noiseLevel float64) []float64 {
	components := net.cimNoiseComponentsLocked()
	if components.TotalSigma() == 0 && noiseLevel > 0 {
		components = defaultNoiseComponents(noiseLevel)
	}
	if components.TotalSigma() == 0 {
		return data
	}
	net.rngMu.Lock()
	result := applyDecomposedNoise(data, components, net.rng)
	net.rngMu.Unlock()
	return result
}
