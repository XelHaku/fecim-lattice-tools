// Package layers provides neural network layer implementations for CIM simulation.
// audio_cnn.go implements neuromorphic audio processing and end-to-end CNN
// inference pipelines for ferroelectric CIM accelerators.
//
// Research basis:
// - Gammatone filterbank: Bio-inspired cochlear model for audio processing
// - Silicon cochlea: 64-channel, 9.6 Hz to 14.6 kHz, 29.7 mW
// - AIMC speech recognition: Near-software accuracy on TI-46-Word dataset
// - ResNet-32 on CIM: 93.7% accuracy on CIFAR-10 with 256×256 crossbars
// - FDCA accelerator: 17.1-18.79 TOPS/W on VGG16/ResNet50
// - Layer tiling: Split large layers across multiple crossbar arrays
//
// Key audio processing features:
// - Gammatone filterbank (24-64 channels)
// - Mel-frequency cepstral coefficients (MFCC)
// - Spike encoding (rate and temporal coding)
// - Event-based audio representation
package layers

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
)

// =============================================================================
// NEUROMORPHIC AUDIO PROCESSING
// =============================================================================

// AudioConfig configures audio processing parameters
type AudioConfig struct {
	// Sampling parameters
	SampleRateHz    int     // Audio sample rate (16000 typical)
	FrameSizeMs     float64 // Analysis frame size (25ms typical)
	FrameStepMs     float64 // Frame step/hop (10ms typical)

	// Filterbank parameters
	NumFilters      int     // Number of mel/gammatone filters (24-64)
	LowFreqHz       float64 // Lowest filter frequency
	HighFreqHz      float64 // Highest filter frequency

	// Feature extraction
	NumMFCC         int     // Number of MFCC coefficients (13 typical)
	UseDelta        bool    // Include delta features
	UseDeltaDelta   bool    // Include delta-delta features

	// Spike encoding
	SpikeThreshold  float64 // Threshold for spike generation
	RefractoryMs    float64 // Refractory period
}

// DefaultAudioConfig returns typical speech processing configuration
func DefaultAudioConfig() *AudioConfig {
	return &AudioConfig{
		SampleRateHz:   16000,
		FrameSizeMs:    25,
		FrameStepMs:    10,
		NumFilters:     40,
		LowFreqHz:      80,
		HighFreqHz:     7600,
		NumMFCC:        13,
		UseDelta:       true,
		UseDeltaDelta:  true,
		SpikeThreshold: 0.5,
		RefractoryMs:   1.0,
	}
}

// GammatoneFilterbank implements bio-inspired cochlear filterbank
type GammatoneFilterbank struct {
	config         *AudioConfig
	centerFreqs    []float64 // Center frequencies for each filter
	filterOrder    int       // Gammatone filter order (typically 4)
	bandwidths     []float64 // ERB-scaled bandwidths
}

// NewGammatoneFilterbank creates a new gammatone filterbank
func NewGammatoneFilterbank(config *AudioConfig) *GammatoneFilterbank {
	if config == nil {
		config = DefaultAudioConfig()
	}

	fb := &GammatoneFilterbank{
		config:      config,
		filterOrder: 4, // Standard gammatone order
	}

	// Calculate ERB-spaced center frequencies
	fb.centerFreqs = fb.calculateERBSpacing()
	fb.bandwidths = fb.calculateBandwidths()

	return fb
}

// calculateERBSpacing computes ERB-spaced center frequencies
func (fb *GammatoneFilterbank) calculateERBSpacing() []float64 {
	// ERB scale: ERB(f) = 24.7 * (4.37 * f/1000 + 1)
	lowERB := fb.hzToERB(fb.config.LowFreqHz)
	highERB := fb.hzToERB(fb.config.HighFreqHz)

	step := (highERB - lowERB) / float64(fb.config.NumFilters-1)

	freqs := make([]float64, fb.config.NumFilters)
	for i := 0; i < fb.config.NumFilters; i++ {
		erb := lowERB + float64(i)*step
		freqs[i] = fb.erbToHz(erb)
	}

	return freqs
}

// hzToERB converts Hz to ERB scale
func (fb *GammatoneFilterbank) hzToERB(hz float64) float64 {
	return 21.4 * math.Log10(4.37*hz/1000+1)
}

// erbToHz converts ERB scale to Hz
func (fb *GammatoneFilterbank) erbToHz(erb float64) float64 {
	return (math.Pow(10, erb/21.4) - 1) * 1000 / 4.37
}

// calculateBandwidths computes ERB-scaled bandwidths
func (fb *GammatoneFilterbank) calculateBandwidths() []float64 {
	bw := make([]float64, len(fb.centerFreqs))
	for i, f := range fb.centerFreqs {
		// ERB bandwidth: B = 24.7 * (4.37 * f/1000 + 1)
		erb := 24.7 * (4.37*f/1000 + 1)
		// Gammatone bandwidth factor: 1.019
		bw[i] = erb * 1.019
	}
	return bw
}

// Process applies gammatone filterbank to audio frame
func (fb *GammatoneFilterbank) Process(samples []float64) []float64 {
	output := make([]float64, fb.config.NumFilters)

	// For each filter channel
	for ch := 0; ch < fb.config.NumFilters; ch++ {
		// Apply gammatone filter (simplified FIR approximation)
		filtered := fb.applyGammatoneFilter(samples, ch)

		// Compute envelope (half-wave rectification + smoothing)
		envelope := fb.computeEnvelope(filtered)

		// Output is mean envelope energy
		sum := 0.0
		for _, e := range envelope {
			sum += e * e
		}
		output[ch] = math.Sqrt(sum / float64(len(envelope)+1))
	}

	return output
}

// applyGammatoneFilter applies single gammatone filter
func (fb *GammatoneFilterbank) applyGammatoneFilter(samples []float64, channel int) []float64 {
	cf := fb.centerFreqs[channel]
	bw := fb.bandwidths[channel]
	n := fb.filterOrder
	fs := float64(fb.config.SampleRateHz)

	output := make([]float64, len(samples))

	// Gammatone impulse response: g(t) = t^(n-1) * exp(-2πBt) * cos(2πft)
	// Simplified: apply as modulated bandpass
	for i, s := range samples {
		t := float64(i) / fs
		// Envelope
		env := math.Pow(t*1000+0.001, float64(n-1)) * math.Exp(-2*math.Pi*bw*t)
		// Modulation
		carrier := math.Cos(2 * math.Pi * cf * t)
		output[i] = s * env * carrier
	}

	return output
}

// computeEnvelope extracts envelope via half-wave rectification
func (fb *GammatoneFilterbank) computeEnvelope(filtered []float64) []float64 {
	envelope := make([]float64, len(filtered))

	// Half-wave rectification
	for i, v := range filtered {
		if v > 0 {
			envelope[i] = v
		}
	}

	// Low-pass smoothing (simple moving average)
	windowSize := int(fb.config.SampleRateHz / 1000) // 1ms window
	if windowSize < 1 {
		windowSize = 1
	}

	smoothed := make([]float64, len(envelope))
	for i := range envelope {
		sum := 0.0
		count := 0
		for j := i - windowSize; j <= i+windowSize; j++ {
			if j >= 0 && j < len(envelope) {
				sum += envelope[j]
				count++
			}
		}
		smoothed[i] = sum / float64(count)
	}

	return smoothed
}

// =============================================================================
// MFCC FEATURE EXTRACTION
// =============================================================================

// MFCCExtractor extracts Mel-frequency cepstral coefficients
type MFCCExtractor struct {
	config       *AudioConfig
	melFilters   [][]float64 // Mel filterbank matrix
	dctMatrix    [][]float64 // DCT matrix for cepstral conversion
	frameSamples int
	fftSize      int
}

// NewMFCCExtractor creates a new MFCC extractor
func NewMFCCExtractor(config *AudioConfig) *MFCCExtractor {
	if config == nil {
		config = DefaultAudioConfig()
	}

	frameSamples := int(config.FrameSizeMs * float64(config.SampleRateHz) / 1000)
	// FFT size is next power of 2
	fftSize := 1
	for fftSize < frameSamples {
		fftSize *= 2
	}

	ext := &MFCCExtractor{
		config:       config,
		frameSamples: frameSamples,
		fftSize:      fftSize,
	}

	ext.melFilters = ext.createMelFilterbank()
	ext.dctMatrix = ext.createDCTMatrix()

	return ext
}

// hzToMel converts Hz to Mel scale
func (m *MFCCExtractor) hzToMel(hz float64) float64 {
	return 2595 * math.Log10(1+hz/700)
}

// melToHz converts Mel scale to Hz
func (m *MFCCExtractor) melToHz(mel float64) float64 {
	return 700 * (math.Pow(10, mel/2595) - 1)
}

// createMelFilterbank creates triangular mel filterbank
func (m *MFCCExtractor) createMelFilterbank() [][]float64 {
	numFilters := m.config.NumFilters
	fftBins := m.fftSize/2 + 1

	// Mel-spaced frequencies
	lowMel := m.hzToMel(m.config.LowFreqHz)
	highMel := m.hzToMel(m.config.HighFreqHz)

	melPoints := make([]float64, numFilters+2)
	for i := range melPoints {
		melPoints[i] = lowMel + float64(i)*(highMel-lowMel)/float64(numFilters+1)
	}

	// Convert to Hz and FFT bin indices
	binIndices := make([]int, len(melPoints))
	for i, mel := range melPoints {
		hz := m.melToHz(mel)
		binIndices[i] = int(hz * float64(m.fftSize) / float64(m.config.SampleRateHz))
		if binIndices[i] >= fftBins {
			binIndices[i] = fftBins - 1
		}
	}

	// Create triangular filters
	filters := make([][]float64, numFilters)
	for i := 0; i < numFilters; i++ {
		filters[i] = make([]float64, fftBins)

		left := binIndices[i]
		center := binIndices[i+1]
		right := binIndices[i+2]

		// Rising edge
		for j := left; j < center; j++ {
			if center > left {
				filters[i][j] = float64(j-left) / float64(center-left)
			}
		}

		// Falling edge
		for j := center; j < right; j++ {
			if right > center {
				filters[i][j] = float64(right-j) / float64(right-center)
			}
		}
	}

	return filters
}

// createDCTMatrix creates DCT-II matrix for cepstral conversion
func (m *MFCCExtractor) createDCTMatrix() [][]float64 {
	numMFCC := m.config.NumMFCC
	numFilters := m.config.NumFilters

	dct := make([][]float64, numMFCC)
	for i := 0; i < numMFCC; i++ {
		dct[i] = make([]float64, numFilters)
		for j := 0; j < numFilters; j++ {
			dct[i][j] = math.Cos(math.Pi * float64(i) * (float64(j) + 0.5) / float64(numFilters))
		}
	}

	return dct
}

// ExtractFrame extracts MFCC features from a single frame
func (m *MFCCExtractor) ExtractFrame(samples []float64) []float64 {
	// Zero-pad or truncate to frame size
	frame := make([]float64, m.fftSize)
	copy(frame, samples)

	// Apply Hamming window
	for i := 0; i < len(samples) && i < m.fftSize; i++ {
		window := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(m.frameSamples-1))
		frame[i] *= window
	}

	// Compute power spectrum via FFT
	powerSpectrum := m.computePowerSpectrum(frame)

	// Apply mel filterbank
	melEnergies := make([]float64, m.config.NumFilters)
	for i := 0; i < m.config.NumFilters; i++ {
		sum := 0.0
		for j := 0; j < len(powerSpectrum); j++ {
			sum += powerSpectrum[j] * m.melFilters[i][j]
		}
		// Log energy
		if sum < 1e-10 {
			sum = 1e-10
		}
		melEnergies[i] = math.Log(sum)
	}

	// Apply DCT to get MFCCs
	mfcc := make([]float64, m.config.NumMFCC)
	for i := 0; i < m.config.NumMFCC; i++ {
		sum := 0.0
		for j := 0; j < m.config.NumFilters; j++ {
			sum += melEnergies[j] * m.dctMatrix[i][j]
		}
		mfcc[i] = sum
	}

	return mfcc
}

// computePowerSpectrum computes power spectrum using DFT
func (m *MFCCExtractor) computePowerSpectrum(frame []float64) []float64 {
	n := len(frame)
	spectrum := make([]float64, n/2+1)

	// Simple DFT (for production, use FFT library)
	for k := 0; k <= n/2; k++ {
		var sum complex128
		for t := 0; t < n; t++ {
			angle := -2 * math.Pi * float64(k) * float64(t) / float64(n)
			sum += complex(frame[t], 0) * cmplx.Exp(complex(0, angle))
		}
		spectrum[k] = real(sum)*real(sum) + imag(sum)*imag(sum)
	}

	return spectrum
}

// ExtractSequence extracts MFCC features from audio sequence
func (m *MFCCExtractor) ExtractSequence(audio []float64) [][]float64 {
	stepSamples := int(m.config.FrameStepMs * float64(m.config.SampleRateHz) / 1000)
	numFrames := (len(audio) - m.frameSamples) / stepSamples

	if numFrames < 1 {
		numFrames = 1
	}

	features := make([][]float64, numFrames)
	for i := 0; i < numFrames; i++ {
		start := i * stepSamples
		end := start + m.frameSamples
		if end > len(audio) {
			end = len(audio)
		}
		features[i] = m.ExtractFrame(audio[start:end])
	}

	return features
}

// =============================================================================
// SPIKE ENCODING
// =============================================================================

// SpikeEncoder converts analog signals to spike trains
type SpikeEncoder struct {
	config       *AudioConfig
	lastSpikeTimes []float64 // Last spike time per channel
}

// NewSpikeEncoder creates a new spike encoder
func NewSpikeEncoder(numChannels int, config *AudioConfig) *SpikeEncoder {
	if config == nil {
		config = DefaultAudioConfig()
	}

	return &SpikeEncoder{
		config:         config,
		lastSpikeTimes: make([]float64, numChannels),
	}
}

// EncodeRateCoding converts values to spike rates
func (se *SpikeEncoder) EncodeRateCoding(values []float64, timeWindowMs float64) [][]int {
	numChannels := len(values)
	maxSpikes := int(timeWindowMs / se.config.RefractoryMs)

	spikes := make([][]int, numChannels)
	for ch := 0; ch < numChannels; ch++ {
		spikes[ch] = make([]int, 0)

		// Normalize value to [0, 1]
		rate := values[ch]
		if rate < 0 {
			rate = 0
		} else if rate > 1 {
			rate = 1
		}

		// Generate spikes based on rate
		numSpikes := int(rate * float64(maxSpikes))
		for i := 0; i < numSpikes; i++ {
			spikeTime := int(float64(i) * timeWindowMs / float64(numSpikes+1))
			spikes[ch] = append(spikes[ch], spikeTime)
		}
	}

	return spikes
}

// EncodeTemporalCoding converts to time-to-first-spike encoding
func (se *SpikeEncoder) EncodeTemporalCoding(values []float64, maxLatencyMs float64) []float64 {
	spikeTimes := make([]float64, len(values))

	for i, v := range values {
		// Higher value = earlier spike
		if v > se.config.SpikeThreshold {
			// Latency inversely proportional to value
			latency := maxLatencyMs * (1 - v)
			if latency < 0 {
				latency = 0
			}
			spikeTimes[i] = latency
		} else {
			spikeTimes[i] = -1 // No spike
		}
	}

	return spikeTimes
}

// =============================================================================
// CNN INFERENCE PIPELINE
// =============================================================================

// CNNLayerType represents different CNN layer types
type CNNLayerType int

const (
	LayerConv2D    CNNLayerType = iota // Convolution
	LayerMaxPool                        // Max pooling
	LayerAvgPool                        // Average pooling
	LayerFC                             // Fully connected
	LayerBatchNorm                      // Batch normalization
	LayerReLU                           // ReLU activation
	LayerSoftmax                        // Softmax output
	LayerResidual                       // Residual connection
)

// CNNLayerConfig configures a single CNN layer
type CNNLayerConfig struct {
	Type        CNNLayerType
	Name        string

	// Convolution parameters
	InChannels  int
	OutChannels int
	KernelSize  int
	Stride      int
	Padding     int

	// Pooling parameters
	PoolSize    int

	// FC parameters
	InFeatures  int
	OutFeatures int

	// Weights (pre-loaded)
	Weights     [][]float64
	Bias        []float64
}

// CNNPipeline manages end-to-end CNN inference
type CNNPipeline struct {
	layers       []*CNNLayerConfig
	crossbarSize int
	tileMapping  []LayerTileMapping

	// Statistics
	totalOps     int64
	totalLatency float64
}

// LayerTileMapping describes how a layer maps to crossbar tiles
type LayerTileMapping struct {
	LayerIndex  int
	NumTiles    int
	TileRows    int
	TileCols    int
	IsSplit     bool // True if layer spans multiple tiles
}

// NewCNNPipeline creates a new CNN inference pipeline
func NewCNNPipeline(crossbarSize int) *CNNPipeline {
	return &CNNPipeline{
		layers:       make([]*CNNLayerConfig, 0),
		crossbarSize: crossbarSize,
		tileMapping:  make([]LayerTileMapping, 0),
	}
}

// AddLayer adds a layer to the pipeline
func (p *CNNPipeline) AddLayer(config *CNNLayerConfig) {
	p.layers = append(p.layers, config)
	p.updateTileMapping(len(p.layers) - 1)
}

// updateTileMapping computes crossbar mapping for layer
func (p *CNNPipeline) updateTileMapping(layerIndex int) {
	layer := p.layers[layerIndex]

	// Calculate layer matrix dimensions
	var rows, cols int
	switch layer.Type {
	case LayerConv2D:
		// im2col mapping: rows = OutChannels, cols = InChannels * K * K
		rows = layer.OutChannels
		cols = layer.InChannels * layer.KernelSize * layer.KernelSize
	case LayerFC:
		rows = layer.OutFeatures
		cols = layer.InFeatures
	default:
		// Non-matrix layers don't need mapping
		p.tileMapping = append(p.tileMapping, LayerTileMapping{
			LayerIndex: layerIndex,
			NumTiles:   0,
		})
		return
	}

	// Calculate number of tiles needed
	tilesRows := (rows + p.crossbarSize - 1) / p.crossbarSize
	tilesCols := (cols + p.crossbarSize - 1) / p.crossbarSize
	numTiles := tilesRows * tilesCols

	p.tileMapping = append(p.tileMapping, LayerTileMapping{
		LayerIndex: layerIndex,
		NumTiles:   numTiles,
		TileRows:   tilesRows,
		TileCols:   tilesCols,
		IsSplit:    numTiles > 1,
	})
}

// Forward runs inference on input
func (p *CNNPipeline) Forward(input [][][]float64) ([]float64, PipelineStats) {
	var stats PipelineStats
	current := input

	for i, layer := range p.layers {
		switch layer.Type {
		case LayerConv2D:
			current, stats.ConvOps = p.forwardConv2D(current, layer)
		case LayerMaxPool:
			current = p.forwardMaxPool(current, layer)
		case LayerAvgPool:
			current = p.forwardAvgPool(current, layer)
		case LayerReLU:
			current = p.forwardReLU(current)
		case LayerFC:
			// Flatten if needed
			flat := p.flatten(current)
			output := p.forwardFC(flat, layer)
			stats.FCOps += int64(len(flat) * len(output))
			// Return FC output directly
			if i == len(p.layers)-1 {
				return output, stats
			}
			// Reshape for next layer (shouldn't happen normally)
			current = [][][]float64{{{0}}} // Placeholder
		case LayerSoftmax:
			flat := p.flatten(current)
			return p.forwardSoftmax(flat), stats
		}
	}

	return p.flatten(current), stats
}

// forwardConv2D performs convolution on crossbar
func (p *CNNPipeline) forwardConv2D(input [][][]float64, layer *CNNLayerConfig) ([][][]float64, int64) {
	inC := len(input)
	inH := len(input[0])
	inW := len(input[0][0])

	outH := (inH + 2*layer.Padding - layer.KernelSize) / layer.Stride + 1
	outW := (inW + 2*layer.Padding - layer.KernelSize) / layer.Stride + 1

	output := make([][][]float64, layer.OutChannels)
	for oc := 0; oc < layer.OutChannels; oc++ {
		output[oc] = make([][]float64, outH)
		for h := 0; h < outH; h++ {
			output[oc][h] = make([]float64, outW)
		}
	}

	ops := int64(0)

	// Convolution with im2col conceptually
	for oc := 0; oc < layer.OutChannels; oc++ {
		for oh := 0; oh < outH; oh++ {
			for ow := 0; ow < outW; ow++ {
				sum := 0.0

				// Dot product over receptive field
				for ic := 0; ic < inC; ic++ {
					for kh := 0; kh < layer.KernelSize; kh++ {
						for kw := 0; kw < layer.KernelSize; kw++ {
							ih := oh*layer.Stride + kh - layer.Padding
							iw := ow*layer.Stride + kw - layer.Padding

							if ih >= 0 && ih < inH && iw >= 0 && iw < inW {
								// Weight index
								wIdx := ic*layer.KernelSize*layer.KernelSize + kh*layer.KernelSize + kw
								if layer.Weights != nil && oc < len(layer.Weights) && wIdx < len(layer.Weights[oc]) {
									sum += input[ic][ih][iw] * layer.Weights[oc][wIdx]
								}
								ops++
							}
						}
					}
				}

				// Add bias
				if layer.Bias != nil && oc < len(layer.Bias) {
					sum += layer.Bias[oc]
				}

				output[oc][oh][ow] = sum
			}
		}
	}

	return output, ops
}

// forwardMaxPool performs max pooling
func (p *CNNPipeline) forwardMaxPool(input [][][]float64, layer *CNNLayerConfig) [][][]float64 {
	c := len(input)
	h := len(input[0])
	w := len(input[0][0])

	poolSize := layer.PoolSize
	outH := h / poolSize
	outW := w / poolSize

	output := make([][][]float64, c)
	for ch := 0; ch < c; ch++ {
		output[ch] = make([][]float64, outH)
		for oh := 0; oh < outH; oh++ {
			output[ch][oh] = make([]float64, outW)
			for ow := 0; ow < outW; ow++ {
				maxVal := math.Inf(-1)
				for ph := 0; ph < poolSize; ph++ {
					for pw := 0; pw < poolSize; pw++ {
						val := input[ch][oh*poolSize+ph][ow*poolSize+pw]
						if val > maxVal {
							maxVal = val
						}
					}
				}
				output[ch][oh][ow] = maxVal
			}
		}
	}

	return output
}

// forwardAvgPool performs average pooling
func (p *CNNPipeline) forwardAvgPool(input [][][]float64, layer *CNNLayerConfig) [][][]float64 {
	c := len(input)
	h := len(input[0])
	w := len(input[0][0])

	poolSize := layer.PoolSize
	outH := h / poolSize
	outW := w / poolSize

	output := make([][][]float64, c)
	for ch := 0; ch < c; ch++ {
		output[ch] = make([][]float64, outH)
		for oh := 0; oh < outH; oh++ {
			output[ch][oh] = make([]float64, outW)
			for ow := 0; ow < outW; ow++ {
				sum := 0.0
				for ph := 0; ph < poolSize; ph++ {
					for pw := 0; pw < poolSize; pw++ {
						sum += input[ch][oh*poolSize+ph][ow*poolSize+pw]
					}
				}
				output[ch][oh][ow] = sum / float64(poolSize*poolSize)
			}
		}
	}

	return output
}

// forwardReLU applies ReLU activation
func (p *CNNPipeline) forwardReLU(input [][][]float64) [][][]float64 {
	output := make([][][]float64, len(input))
	for c := range input {
		output[c] = make([][]float64, len(input[c]))
		for h := range input[c] {
			output[c][h] = make([]float64, len(input[c][h]))
			for w := range input[c][h] {
				if input[c][h][w] > 0 {
					output[c][h][w] = input[c][h][w]
				}
			}
		}
	}
	return output
}

// forwardFC performs fully connected layer on crossbar
func (p *CNNPipeline) forwardFC(input []float64, layer *CNNLayerConfig) []float64 {
	output := make([]float64, layer.OutFeatures)

	for o := 0; o < layer.OutFeatures; o++ {
		sum := 0.0
		for i := 0; i < len(input) && i < layer.InFeatures; i++ {
			if layer.Weights != nil && o < len(layer.Weights) && i < len(layer.Weights[o]) {
				sum += input[i] * layer.Weights[o][i]
			}
		}
		if layer.Bias != nil && o < len(layer.Bias) {
			sum += layer.Bias[o]
		}
		output[o] = sum
	}

	return output
}

// forwardSoftmax applies softmax activation
func (p *CNNPipeline) forwardSoftmax(input []float64) []float64 {
	// Find max for numerical stability
	maxVal := input[0]
	for _, v := range input {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute exp and sum
	expSum := 0.0
	output := make([]float64, len(input))
	for i, v := range input {
		output[i] = math.Exp(v - maxVal)
		expSum += output[i]
	}

	// Normalize
	for i := range output {
		output[i] /= expSum
	}

	return output
}

// flatten converts 3D tensor to 1D vector
func (p *CNNPipeline) flatten(input [][][]float64) []float64 {
	c := len(input)
	h := len(input[0])
	w := len(input[0][0])

	flat := make([]float64, c*h*w)
	idx := 0
	for ch := 0; ch < c; ch++ {
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				flat[idx] = input[ch][row][col]
				idx++
			}
		}
	}

	return flat
}

// PipelineStats holds pipeline statistics
type PipelineStats struct {
	ConvOps    int64
	FCOps      int64
	TotalOps   int64
	LatencyMs  float64
}

// =============================================================================
// PRE-DEFINED CNN ARCHITECTURES
// =============================================================================

// CreateLeNet5 creates LeNet-5 architecture for MNIST
func CreateLeNet5(crossbarSize int) *CNNPipeline {
	pipeline := NewCNNPipeline(crossbarSize)

	// Conv1: 1x32x32 -> 6x28x28
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerConv2D,
		Name:        "conv1",
		InChannels:  1,
		OutChannels: 6,
		KernelSize:  5,
		Stride:      1,
		Padding:     0,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu1"})

	// Pool1: 6x28x28 -> 6x14x14
	pipeline.AddLayer(&CNNLayerConfig{
		Type:     LayerMaxPool,
		Name:     "pool1",
		PoolSize: 2,
	})

	// Conv2: 6x14x14 -> 16x10x10
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerConv2D,
		Name:        "conv2",
		InChannels:  6,
		OutChannels: 16,
		KernelSize:  5,
		Stride:      1,
		Padding:     0,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu2"})

	// Pool2: 16x10x10 -> 16x5x5
	pipeline.AddLayer(&CNNLayerConfig{
		Type:     LayerMaxPool,
		Name:     "pool2",
		PoolSize: 2,
	})

	// FC1: 400 -> 120
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerFC,
		Name:        "fc1",
		InFeatures:  400,
		OutFeatures: 120,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu3"})

	// FC2: 120 -> 84
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerFC,
		Name:        "fc2",
		InFeatures:  120,
		OutFeatures: 84,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu4"})

	// FC3: 84 -> 10
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerFC,
		Name:        "fc3",
		InFeatures:  84,
		OutFeatures: 10,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerSoftmax, Name: "softmax"})

	return pipeline
}

// CreateSimpleCIFAR creates simple CNN for CIFAR-10
func CreateSimpleCIFAR(crossbarSize int) *CNNPipeline {
	pipeline := NewCNNPipeline(crossbarSize)

	// Conv1: 3x32x32 -> 32x32x32
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerConv2D,
		Name:        "conv1",
		InChannels:  3,
		OutChannels: 32,
		KernelSize:  3,
		Stride:      1,
		Padding:     1,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu1"})

	// Conv2: 32x32x32 -> 32x32x32
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerConv2D,
		Name:        "conv2",
		InChannels:  32,
		OutChannels: 32,
		KernelSize:  3,
		Stride:      1,
		Padding:     1,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu2"})

	// Pool1: 32x32x32 -> 32x16x16
	pipeline.AddLayer(&CNNLayerConfig{
		Type:     LayerMaxPool,
		Name:     "pool1",
		PoolSize: 2,
	})

	// Conv3: 32x16x16 -> 64x16x16
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerConv2D,
		Name:        "conv3",
		InChannels:  32,
		OutChannels: 64,
		KernelSize:  3,
		Stride:      1,
		Padding:     1,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu3"})

	// Pool2: 64x16x16 -> 64x8x8
	pipeline.AddLayer(&CNNLayerConfig{
		Type:     LayerMaxPool,
		Name:     "pool2",
		PoolSize: 2,
	})

	// FC1: 4096 -> 512
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerFC,
		Name:        "fc1",
		InFeatures:  4096,
		OutFeatures: 512,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerReLU, Name: "relu4"})

	// FC2: 512 -> 10
	pipeline.AddLayer(&CNNLayerConfig{
		Type:        LayerFC,
		Name:        "fc2",
		InFeatures:  512,
		OutFeatures: 10,
	})
	pipeline.AddLayer(&CNNLayerConfig{Type: LayerSoftmax, Name: "softmax"})

	return pipeline
}

// =============================================================================
// AUDIO + CNN INTEGRATION
// =============================================================================

// AudioCNNSystem integrates audio processing with CNN classification
type AudioCNNSystem struct {
	gammatone   *GammatoneFilterbank
	mfcc        *MFCCExtractor
	spikeEncoder *SpikeEncoder
	cnnPipeline *CNNPipeline

	// Configuration
	useSpikes    bool
	useMFCC      bool
}

// NewAudioCNNSystem creates an integrated audio+CNN system
func NewAudioCNNSystem(audioConfig *AudioConfig, crossbarSize int) *AudioCNNSystem {
	if audioConfig == nil {
		audioConfig = DefaultAudioConfig()
	}

	return &AudioCNNSystem{
		gammatone:    NewGammatoneFilterbank(audioConfig),
		mfcc:         NewMFCCExtractor(audioConfig),
		spikeEncoder: NewSpikeEncoder(audioConfig.NumFilters, audioConfig),
		cnnPipeline:  nil, // Set separately based on task
		useSpikes:    false,
		useMFCC:      true,
	}
}

// SetCNNPipeline sets the CNN pipeline for classification
func (sys *AudioCNNSystem) SetCNNPipeline(pipeline *CNNPipeline) {
	sys.cnnPipeline = pipeline
}

// ProcessAudio processes audio and returns classification
func (sys *AudioCNNSystem) ProcessAudio(audio []float64) ([]float64, AudioCNNStats) {
	var stats AudioCNNStats

	// Extract features
	var features [][]float64
	if sys.useMFCC {
		features = sys.mfcc.ExtractSequence(audio)
		stats.NumFrames = len(features)
		stats.FeatureDim = len(features[0])
	}

	// Convert to CNN input format (treat as 1-channel 2D image)
	if sys.cnnPipeline == nil || len(features) == 0 {
		return nil, stats
	}

	// Reshape features to [1][numFrames][featureDim]
	input := make([][][]float64, 1)
	input[0] = features

	// Run CNN
	output, pipelineStats := sys.cnnPipeline.Forward(input)
	stats.CNNOps = pipelineStats.ConvOps + pipelineStats.FCOps

	return output, stats
}

// AudioCNNStats holds audio+CNN processing statistics
type AudioCNNStats struct {
	NumFrames  int
	FeatureDim int
	CNNOps     int64
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// GenerateTestAudio generates synthetic test audio
func GenerateTestAudio(durationMs float64, sampleRate int, freqHz float64) []float64 {
	numSamples := int(durationMs * float64(sampleRate) / 1000)
	audio := make([]float64, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		audio[i] = math.Sin(2 * math.Pi * freqHz * t)
	}

	return audio
}

// GenerateTestImage generates synthetic test image
func GenerateTestImage(channels, height, width int, seed int64) [][][]float64 {
	rng := rand.New(rand.NewSource(seed))

	image := make([][][]float64, channels)
	for c := 0; c < channels; c++ {
		image[c] = make([][]float64, height)
		for h := 0; h < height; h++ {
			image[c][h] = make([]float64, width)
			for w := 0; w < width; w++ {
				image[c][h][w] = rng.Float64()
			}
		}
	}

	return image
}

// FormatPipelineReport generates CNN pipeline report
func FormatPipelineReport(pipeline *CNNPipeline) string {
	report := "=== CNN Pipeline Report ===\n\n"
	report += fmt.Sprintf("Crossbar Size: %d × %d\n", pipeline.crossbarSize, pipeline.crossbarSize)
	report += fmt.Sprintf("Number of Layers: %d\n\n", len(pipeline.layers))

	report += "Layer Details:\n"
	for i, layer := range pipeline.layers {
		report += fmt.Sprintf("  %d. %s (%s)\n", i+1, layer.Name, layerTypeName(layer.Type))

		if mapping := pipeline.tileMapping[i]; mapping.NumTiles > 0 {
			report += fmt.Sprintf("     Tiles: %d (%d × %d), Split: %v\n",
				mapping.NumTiles, mapping.TileRows, mapping.TileCols, mapping.IsSplit)
		}
	}

	return report
}

func layerTypeName(t CNNLayerType) string {
	names := map[CNNLayerType]string{
		LayerConv2D:    "Conv2D",
		LayerMaxPool:   "MaxPool",
		LayerAvgPool:   "AvgPool",
		LayerFC:        "FC",
		LayerBatchNorm: "BatchNorm",
		LayerReLU:      "ReLU",
		LayerSoftmax:   "Softmax",
		LayerResidual:  "Residual",
	}
	return names[t]
}
