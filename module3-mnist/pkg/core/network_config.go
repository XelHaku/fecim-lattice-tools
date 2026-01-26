package core

// SetNumLevels updates the quantization levels and re-quantizes weights.
func (net *DualModeNetwork) SetNumLevels(levels int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 1 {
		levels = 1
	}
	if levels > 30 {
		levels = 30
	}
	net.Config.NumLevels = levels
	net.requantizeWeightsLocked()
}

// GetNumLevels returns the current quantization levels.
func (net *DualModeNetwork) GetNumLevels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.NumLevels
}

// SetNoiseLevel updates the noise level for CIM inference.
func (net *DualModeNetwork) SetNoiseLevel(noise float64) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if noise < 0 {
		noise = 0
	}
	if noise > 0.5 {
		noise = 0.5
	}
	net.Config.NoiseLevel = noise
}

// SetADCBits updates the ADC resolution for CIM inference.
func (net *DualModeNetwork) SetADCBits(bits int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if bits < 3 {
		bits = 3
	}
	if bits > 16 {
		bits = 16
	}
	net.Config.ADCBits = bits
}

// SetDACBits updates the DAC resolution for CIM inference.
func (net *DualModeNetwork) SetDACBits(bits int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if bits < 3 {
		bits = 3
	}
	if bits > 16 {
		bits = 16
	}
	net.Config.DACBits = bits
}

// SetSingleLayer enables/disables Tour Mode (single-layer 784→10 architecture).
// When enabled, this matches Dr. Tour's MNIST demo with ~87% theoretical max accuracy.
func (net *DualModeNetwork) SetSingleLayer(enabled bool) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.Config.SingleLayer = enabled
}

// SetPerLayerQuant enables/disables per-layer quantization mode.
// When enabled, Layer1Levels and Layer2Levels are used instead of NumLevels.
func (net *DualModeNetwork) SetPerLayerQuant(enabled bool) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.Config.PerLayerQuant = enabled
	net.requantizeWeightsLocked()
}

// IsPerLayerQuant returns whether per-layer quantization is enabled.
func (net *DualModeNetwork) IsPerLayerQuant() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.PerLayerQuant
}

// SetLayer1Levels sets the quantization levels for layer 1 (hidden layer).
func (net *DualModeNetwork) SetLayer1Levels(levels int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 2 {
		levels = 2
	}
	if levels > 30 {
		levels = 30
	}
	net.Config.Layer1Levels = levels
	if net.Config.PerLayerQuant {
		net.requantizeWeightsLocked()
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
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 2 {
		levels = 2
	}
	if levels > 30 {
		levels = 30
	}
	net.Config.Layer2Levels = levels
	if net.Config.PerLayerQuant {
		net.requantizeWeightsLocked()
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
	if layer1 > 30 {
		layer1 = 30
	}
	if layer2 < 2 {
		layer2 = 2
	}
	if layer2 > 30 {
		layer2 = 30
	}

	net.Config.Layer1Levels = layer1
	net.Config.Layer2Levels = layer2
	net.Config.PerLayerQuant = true
	net.requantizeWeightsLocked()
}

// IsSingleLayer returns whether single-layer (Tour Mode) is enabled.
func (net *DualModeNetwork) IsSingleLayer() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.SingleLayer
}

// safeNoise applies Gaussian noise with thread-safe RNG access.
// This prevents data races when multiple goroutines hold RLock on the network.
func (net *DualModeNetwork) safeNoise(data []float64, noiseLevel float64) []float64 {
	if noiseLevel <= 0 {
		return data
	}
	net.rngMu.Lock()
	result := AddGaussianNoise(data, noiseLevel, net.rng)
	net.rngMu.Unlock()
	return result
}
