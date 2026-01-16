// Package layers provides neural network layer implementations for CIM deployment.
// This file implements in-sensor computing and replay-based continual learning.
// Based on research: Nature npj (edge intelligence), Science Advances (retinomorphic),
// Nature Communications (generative replay), NIPS 2017 (deep generative replay)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// RETINOMORPHIC IN-SENSOR COMPUTING
// Bio-inspired visual processing at the sensor edge
// =============================================================================

// RetinaCellType defines the type of retinal processing cell.
type RetinaCellType int

const (
	RetinaCellPhotoreceptor RetinaCellType = iota // Light sensing (rods/cones)
	RetinaCellBipolar                             // Signal relay
	RetinaCellAmacrine                            // Lateral inhibition
	RetinaCellGanglion                            // Output integration
	RetinaCellHorizontal                          // Lateral feedback
)

// PhotoreceptorConfig configures photoreceptor simulation.
type PhotoreceptorConfig struct {
	// Spectral response (RGB sensitivity)
	RedSensitivity   float64
	GreenSensitivity float64
	BlueSensitivity  float64

	// Temporal dynamics
	IntegrationTimeMs float64
	AdaptationTauMs   float64

	// Ferroelectric integration
	UseFeFET          bool
	ThresholdVoltage  float64
	MemoryWindow      float64
}

// DefaultPhotoreceptorConfig returns default photoreceptor configuration.
func DefaultPhotoreceptorConfig() *PhotoreceptorConfig {
	return &PhotoreceptorConfig{
		RedSensitivity:    0.33,
		GreenSensitivity:  0.34,
		BlueSensitivity:   0.33,
		IntegrationTimeMs: 10.0,
		AdaptationTauMs:   100.0,
		UseFeFET:          true,
		ThresholdVoltage:  0.5,
		MemoryWindow:      2.0,
	}
}

// FeFETPhotoreceptor implements ferroelectric photoreceptor cell.
type FeFETPhotoreceptor struct {
	Config *PhotoreceptorConfig

	// State variables
	Photocurrent    float64 // Current light-induced current
	AdaptationLevel float64 // Adaptation state (0-1)
	Polarization    float64 // FeFET polarization state
	OutputCurrent   float64 // Post-adaptation output

	// Temporal filtering
	IntegratedLight float64
	LastUpdateTime  float64
}

// NewFeFETPhotoreceptor creates a new FeFET photoreceptor.
func NewFeFETPhotoreceptor(config *PhotoreceptorConfig) *FeFETPhotoreceptor {
	if config == nil {
		config = DefaultPhotoreceptorConfig()
	}
	return &FeFETPhotoreceptor{
		Config:          config,
		AdaptationLevel: 0.5,
	}
}

// Sense processes incoming light intensity.
func (p *FeFETPhotoreceptor) Sense(r, g, b float64, timeMs float64) float64 {
	// Compute weighted photocurrent
	p.Photocurrent = r*p.Config.RedSensitivity +
		g*p.Config.GreenSensitivity +
		b*p.Config.BlueSensitivity

	// Temporal integration
	dt := timeMs - p.LastUpdateTime
	if dt > 0 {
		alpha := dt / p.Config.IntegrationTimeMs
		if alpha > 1 {
			alpha = 1
		}
		p.IntegratedLight = (1-alpha)*p.IntegratedLight + alpha*p.Photocurrent
	}

	// Light adaptation (Weber-Fechner law)
	adaptDt := dt / p.Config.AdaptationTauMs
	if adaptDt > 1 {
		adaptDt = 1
	}
	p.AdaptationLevel = (1-adaptDt)*p.AdaptationLevel + adaptDt*p.IntegratedLight

	// FeFET modulation
	if p.Config.UseFeFET {
		// Polarization encodes adaptation state
		p.Polarization = p.AdaptationLevel * p.Config.MemoryWindow

		// Threshold modulation based on polarization
		effectiveVth := p.Config.ThresholdVoltage - p.Polarization*0.1
		if effectiveVth < 0.1 {
			effectiveVth = 0.1
		}

		// Output with adapted threshold
		p.OutputCurrent = (p.IntegratedLight - p.AdaptationLevel) / (p.AdaptationLevel + 0.1)
	} else {
		p.OutputCurrent = p.IntegratedLight / (p.AdaptationLevel + 0.1)
	}

	p.LastUpdateTime = timeMs
	return p.OutputCurrent
}

// =============================================================================
// RETINOMORPHIC PIXEL ARRAY
// =============================================================================

// RetinomorphicPixelConfig configures the pixel array.
type RetinomorphicPixelConfig struct {
	Width, Height int

	// Processing modes
	EnableEdgeDetection    bool
	EnableMotionDetection  bool
	EnableContrastEnhance  bool

	// Lateral inhibition (center-surround)
	SurroundRadius  int
	SurroundWeight  float64
	CenterWeight    float64

	// Temporal difference
	TemporalDiffTau float64
}

// RetinomorphicPixel implements a single in-sensor computing pixel.
type RetinomorphicPixel struct {
	Photoreceptor *FeFETPhotoreceptor
	X, Y          int

	// Processing outputs
	RawIntensity    float64
	AdaptedOutput   float64
	EdgeResponse    float64
	MotionResponse  float64

	// Temporal memory
	PreviousOutput  float64
	OutputHistory   []float64
}

// RetinomorphicArray implements bio-inspired sensor array.
type RetinomorphicArray struct {
	Config *RetinomorphicPixelConfig
	Pixels [][]*RetinomorphicPixel

	// Convolution kernels (stored as FeFET weights)
	EdgeKernel      [][]float64
	MotionKernel    [][]float64

	// Frame buffer for temporal processing
	PreviousFrame   [][]float64
	CurrentFrame    [][]float64
}

// NewRetinomorphicArray creates a new retinomorphic sensor array.
func NewRetinomorphicArray(config *RetinomorphicPixelConfig) *RetinomorphicArray {
	if config == nil {
		config = &RetinomorphicPixelConfig{
			Width:               64,
			Height:              64,
			EnableEdgeDetection: true,
			EnableMotionDetection: true,
			EnableContrastEnhance: true,
			SurroundRadius:      1,
			SurroundWeight:      -0.125,
			CenterWeight:        1.0,
			TemporalDiffTau:     20.0,
		}
	}

	// Initialize pixels
	pixels := make([][]*RetinomorphicPixel, config.Height)
	for y := range pixels {
		pixels[y] = make([]*RetinomorphicPixel, config.Width)
		for x := range pixels[y] {
			pixels[y][x] = &RetinomorphicPixel{
				Photoreceptor: NewFeFETPhotoreceptor(nil),
				X:             x,
				Y:             y,
				OutputHistory: make([]float64, 0),
			}
		}
	}

	// Initialize edge detection kernel (Laplacian of Gaussian approximation)
	edgeKernel := [][]float64{
		{-1, -1, -1},
		{-1, 8, -1},
		{-1, -1, -1},
	}

	// Motion detection kernel (temporal difference)
	motionKernel := [][]float64{
		{0.25, 0.5, 0.25},
		{0.5, 1.0, 0.5},
		{0.25, 0.5, 0.25},
	}

	// Frame buffers
	prevFrame := make([][]float64, config.Height)
	currFrame := make([][]float64, config.Height)
	for y := range prevFrame {
		prevFrame[y] = make([]float64, config.Width)
		currFrame[y] = make([]float64, config.Width)
	}

	return &RetinomorphicArray{
		Config:        config,
		Pixels:        pixels,
		EdgeKernel:    edgeKernel,
		MotionKernel:  motionKernel,
		PreviousFrame: prevFrame,
		CurrentFrame:  currFrame,
	}
}

// ProcessFrame performs in-sensor computation on an image frame.
func (ra *RetinomorphicArray) ProcessFrame(image [][][]float64, timeMs float64) *InSensorOutput {
	// image is [height][width][3] RGB
	h, w := ra.Config.Height, ra.Config.Width

	// Copy current to previous
	for y := 0; y < h; y++ {
		copy(ra.PreviousFrame[y], ra.CurrentFrame[y])
	}

	// Phase 1: Photoreceptor sensing with FeFET adaptation
	for y := 0; y < h && y < len(image); y++ {
		for x := 0; x < w && x < len(image[y]); x++ {
			pixel := ra.Pixels[y][x]
			r, g, b := image[y][x][0], image[y][x][1], image[y][x][2]

			// FeFET photoreceptor processing
			pixel.RawIntensity = (r + g + b) / 3.0
			pixel.AdaptedOutput = pixel.Photoreceptor.Sense(r, g, b, timeMs)
			ra.CurrentFrame[y][x] = pixel.AdaptedOutput
		}
	}

	// Phase 2: Center-surround (lateral inhibition) - in-sensor convolution
	csOutput := make([][]float64, h)
	for y := range csOutput {
		csOutput[y] = make([]float64, w)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			center := ra.CurrentFrame[y][x] * ra.Config.CenterWeight
			surround := 0.0
			count := 0

			for dy := -ra.Config.SurroundRadius; dy <= ra.Config.SurroundRadius; dy++ {
				for dx := -ra.Config.SurroundRadius; dx <= ra.Config.SurroundRadius; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					ny, nx := y+dy, x+dx
					if ny >= 0 && ny < h && nx >= 0 && nx < w {
						surround += ra.CurrentFrame[ny][nx]
						count++
					}
				}
			}

			if count > 0 {
				surround = surround / float64(count) * ra.Config.SurroundWeight * float64(count)
			}
			csOutput[y][x] = center + surround
		}
	}

	// Phase 3: Edge detection (in-sensor convolution with FeFET weights)
	edgeOutput := make([][]float64, h)
	for y := range edgeOutput {
		edgeOutput[y] = make([]float64, w)
	}

	if ra.Config.EnableEdgeDetection {
		kSize := len(ra.EdgeKernel)
		kHalf := kSize / 2

		for y := kHalf; y < h-kHalf; y++ {
			for x := kHalf; x < w-kHalf; x++ {
				sum := 0.0
				for ky := 0; ky < kSize; ky++ {
					for kx := 0; kx < kSize; kx++ {
						sum += ra.CurrentFrame[y+ky-kHalf][x+kx-kHalf] * ra.EdgeKernel[ky][kx]
					}
				}
				edgeOutput[y][x] = math.Abs(sum)
				ra.Pixels[y][x].EdgeResponse = edgeOutput[y][x]
			}
		}
	}

	// Phase 4: Motion detection (temporal difference)
	motionOutput := make([][]float64, h)
	for y := range motionOutput {
		motionOutput[y] = make([]float64, w)
	}

	if ra.Config.EnableMotionDetection {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				diff := math.Abs(ra.CurrentFrame[y][x] - ra.PreviousFrame[y][x])
				motionOutput[y][x] = diff
				ra.Pixels[y][x].MotionResponse = diff
			}
		}
	}

	// Compute statistics
	totalEdge := 0.0
	totalMotion := 0.0
	maxEdge := 0.0
	maxMotion := 0.0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			totalEdge += edgeOutput[y][x]
			totalMotion += motionOutput[y][x]
			if edgeOutput[y][x] > maxEdge {
				maxEdge = edgeOutput[y][x]
			}
			if motionOutput[y][x] > maxMotion {
				maxMotion = motionOutput[y][x]
			}
		}
	}

	return &InSensorOutput{
		AdaptedFrame:     csOutput,
		EdgeMap:          edgeOutput,
		MotionMap:        motionOutput,
		MeanEdgeStrength: totalEdge / float64(h*w),
		MeanMotion:       totalMotion / float64(h*w),
		MaxEdge:          maxEdge,
		MaxMotion:        maxMotion,
	}
}

// InSensorOutput contains results of in-sensor preprocessing.
type InSensorOutput struct {
	AdaptedFrame     [][]float64
	EdgeMap          [][]float64
	MotionMap        [][]float64
	MeanEdgeStrength float64
	MeanMotion       float64
	MaxEdge          float64
	MaxMotion        float64
}

// GetSparseEvents extracts events above threshold (event-driven output).
func (ra *RetinomorphicArray) GetSparseEvents(threshold float64) []SensorEvent {
	events := make([]SensorEvent, 0)

	for y := 0; y < ra.Config.Height; y++ {
		for x := 0; x < ra.Config.Width; x++ {
			pixel := ra.Pixels[y][x]

			// Event generation based on change or edge
			if pixel.EdgeResponse > threshold || pixel.MotionResponse > threshold {
				events = append(events, SensorEvent{
					X:         x,
					Y:         y,
					Polarity:  pixel.MotionResponse > 0,
					Intensity: pixel.AdaptedOutput,
					EdgeScore: pixel.EdgeResponse,
				})
			}
		}
	}

	return events
}

// SensorEvent represents a sparse sensor output event.
type SensorEvent struct {
	X, Y      int
	Polarity  bool    // ON (brightness increase) or OFF
	Intensity float64
	EdgeScore float64
}

// =============================================================================
// IN-SENSOR NEURAL NETWORK
// MAC operations with photonic/ferroelectric weights
// =============================================================================

// InSensorNNConfig configures in-sensor neural network.
type InSensorNNConfig struct {
	InputSize    int
	HiddenSizes  []int
	OutputSize   int

	// Weight precision
	WeightBits   int
	ActivationBits int

	// FeFET parameters
	FeFETOnOff    float64
	FeFETVariation float64

	// Energy parameters
	MACEnergyFJ    float64
	SenseEnergyFJ  float64
}

// InSensorNN implements neural network with in-sensor MAC.
type InSensorNN struct {
	Config  *InSensorNNConfig

	// Weights stored as FeFET conductances
	Weights [][][]float64

	// Activation function lookup (quantized)
	ActivationLUT []float64
}

// NewInSensorNN creates an in-sensor neural network.
func NewInSensorNN(config *InSensorNNConfig) *InSensorNN {
	if config == nil {
		config = &InSensorNNConfig{
			InputSize:      64 * 64,
			HiddenSizes:    []int{256},
			OutputSize:     10,
			WeightBits:     6,
			ActivationBits: 8,
			FeFETOnOff:     1e4,
			FeFETVariation: 0.02,
			MACEnergyFJ:    10.0,
			SenseEnergyFJ:  1.0,
		}
	}

	// Initialize weights
	layerSizes := append([]int{config.InputSize}, config.HiddenSizes...)
	layerSizes = append(layerSizes, config.OutputSize)

	weights := make([][][]float64, len(layerSizes)-1)
	for l := 0; l < len(layerSizes)-1; l++ {
		inSize, outSize := layerSizes[l], layerSizes[l+1]
		weights[l] = make([][]float64, inSize)
		for i := range weights[l] {
			weights[l][i] = make([]float64, outSize)
			// Xavier initialization
			scale := math.Sqrt(2.0 / float64(inSize+outSize))
			for j := range weights[l][i] {
				weights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}

	// Build activation LUT (ReLU quantized)
	lutSize := 1 << config.ActivationBits
	lut := make([]float64, lutSize)
	for i := range lut {
		x := float64(i-lutSize/2) / float64(lutSize/2)
		if x > 0 {
			lut[i] = x
		} else {
			lut[i] = 0
		}
	}

	return &InSensorNN{
		Config:        config,
		Weights:       weights,
		ActivationLUT: lut,
	}
}

// ForwardInSensor performs forward pass with in-sensor MAC operations.
func (nn *InSensorNN) ForwardInSensor(sensorOutput [][]float64) ([]float64, *InSensorStats) {
	// Flatten sensor output
	h, w := len(sensorOutput), len(sensorOutput[0])
	input := make([]float64, h*w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			input[y*w+x] = sensorOutput[y][x]
		}
	}

	// Resize if needed
	if len(input) > nn.Config.InputSize {
		input = input[:nn.Config.InputSize]
	} else if len(input) < nn.Config.InputSize {
		padded := make([]float64, nn.Config.InputSize)
		copy(padded, input)
		input = padded
	}

	stats := &InSensorStats{}
	current := input

	// Forward through layers
	for l := 0; l < len(nn.Weights); l++ {
		inSize := len(nn.Weights[l])
		outSize := len(nn.Weights[l][0])
		output := make([]float64, outSize)

		// In-sensor MAC with FeFET weight modulation
		for j := 0; j < outSize; j++ {
			sum := 0.0
			for i := 0; i < inSize && i < len(current); i++ {
				// Add device variation
				variation := 1.0 + nn.Config.FeFETVariation*rand.NormFloat64()
				w := nn.Weights[l][i][j] * variation
				sum += current[i] * w
				stats.MACCount++
			}
			output[j] = sum
		}

		// Quantized activation via LUT
		lutSize := len(nn.ActivationLUT)
		for j := range output {
			// Map to LUT index
			idx := int((output[j] + 1.0) / 2.0 * float64(lutSize-1))
			if idx < 0 {
				idx = 0
			}
			if idx >= lutSize {
				idx = lutSize - 1
			}
			output[j] = nn.ActivationLUT[idx]
		}

		current = output
	}

	// Energy estimation
	stats.SenseEnergy = float64(nn.Config.InputSize) * nn.Config.SenseEnergyFJ
	stats.MACEnergy = float64(stats.MACCount) * nn.Config.MACEnergyFJ
	stats.TotalEnergy = stats.SenseEnergy + stats.MACEnergy

	return current, stats
}

// InSensorStats contains in-sensor processing statistics.
type InSensorStats struct {
	MACCount    int64
	SenseEnergy float64 // fJ
	MACEnergy   float64 // fJ
	TotalEnergy float64 // fJ
}

// =============================================================================
// GENERATIVE REPLAY FOR CONTINUAL LEARNING
// Based on Deep Generative Replay (NIPS 2017)
// =============================================================================

// GeneratorConfig configures the generative replay generator.
type GeneratorConfig struct {
	LatentDim     int
	HiddenSizes   []int
	OutputSize    int

	// VAE parameters
	UseVAE        bool
	KLWeight      float64

	// Training parameters
	LearningRate  float64
	BatchSize     int
}

// SimpleGenerator implements a basic MLP generator for replay.
type SimpleGenerator struct {
	Config *GeneratorConfig

	// MLP weights: latent -> hidden -> output
	Weights [][][]float64
	Biases  [][]float64

	// VAE: additional weights for mean and logvar
	MeanWeights   [][]float64
	LogvarWeights [][]float64
}

// NewSimpleGenerator creates a new generator.
func NewSimpleGenerator(config *GeneratorConfig) *SimpleGenerator {
	if config == nil {
		config = &GeneratorConfig{
			LatentDim:    32,
			HiddenSizes:  []int{128, 256},
			OutputSize:   784,
			UseVAE:       true,
			KLWeight:     1.0,
			LearningRate: 0.001,
			BatchSize:    32,
		}
	}

	// Build layer sizes
	sizes := append([]int{config.LatentDim}, config.HiddenSizes...)
	sizes = append(sizes, config.OutputSize)

	// Initialize weights
	weights := make([][][]float64, len(sizes)-1)
	biases := make([][]float64, len(sizes)-1)

	for l := 0; l < len(sizes)-1; l++ {
		inSize, outSize := sizes[l], sizes[l+1]
		weights[l] = make([][]float64, inSize)
		biases[l] = make([]float64, outSize)

		scale := math.Sqrt(2.0 / float64(inSize))
		for i := range weights[l] {
			weights[l][i] = make([]float64, outSize)
			for j := range weights[l][i] {
				weights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}

	return &SimpleGenerator{
		Config:  config,
		Weights: weights,
		Biases:  biases,
	}
}

// Generate creates synthetic samples from random latent vectors.
func (g *SimpleGenerator) Generate(numSamples int) [][]float64 {
	samples := make([][]float64, numSamples)

	for s := 0; s < numSamples; s++ {
		// Sample from latent space (standard normal)
		z := make([]float64, g.Config.LatentDim)
		for i := range z {
			z[i] = rand.NormFloat64()
		}

		// Forward through generator
		current := z
		for l := 0; l < len(g.Weights); l++ {
			inSize := len(g.Weights[l])
			outSize := len(g.Weights[l][0])
			output := make([]float64, outSize)

			for j := 0; j < outSize; j++ {
				sum := g.Biases[l][j]
				for i := 0; i < inSize && i < len(current); i++ {
					sum += current[i] * g.Weights[l][i][j]
				}

				// Activation: ReLU for hidden, sigmoid for output
				if l < len(g.Weights)-1 {
					if sum < 0 {
						sum = 0
					}
				} else {
					sum = 1.0 / (1.0 + math.Exp(-sum))
				}
				output[j] = sum
			}
			current = output
		}

		samples[s] = current
	}

	return samples
}

// =============================================================================
// GENERATIVE REPLAY CONTINUAL LEARNER
// =============================================================================

// GenerativeReplayConfig configures generative replay system.
type GenerativeReplayConfig struct {
	// Replay ratio: how many replayed vs new samples
	ReplayRatio float64

	// Generator update frequency
	GeneratorUpdateFreq int

	// Solver (classifier) configuration
	SolverHiddenSizes []int

	// Memory budget
	MaxStoredSamples int

	// CIM-specific
	QuantizeBits int
}

// GenerativeReplayCL implements continual learning with generative replay.
type GenerativeReplayCL struct {
	Config    *GenerativeReplayConfig
	Generator *SimpleGenerator

	// Solver (task classifier)
	SolverWeights [][][]float64
	SolverBiases  [][]float64

	// Task tracking
	TaskCount int
	TaskLabels []string

	// Small exemplar memory (for hybrid approach)
	ExemplarMemory [][]float64
	ExemplarLabels []int
}

// NewGenerativeReplayCL creates a new generative replay continual learner.
func NewGenerativeReplayCL(inputSize, outputSize int, config *GenerativeReplayConfig) *GenerativeReplayCL {
	if config == nil {
		config = &GenerativeReplayConfig{
			ReplayRatio:         0.5,
			GeneratorUpdateFreq: 1,
			SolverHiddenSizes:   []int{256, 128},
			MaxStoredSamples:    100,
			QuantizeBits:        6,
		}
	}

	// Create generator
	genConfig := &GeneratorConfig{
		LatentDim:   32,
		HiddenSizes: []int{128, 256},
		OutputSize:  inputSize,
		UseVAE:      true,
	}
	generator := NewSimpleGenerator(genConfig)

	// Create solver
	solverSizes := append([]int{inputSize}, config.SolverHiddenSizes...)
	solverSizes = append(solverSizes, outputSize)

	solverWeights := make([][][]float64, len(solverSizes)-1)
	solverBiases := make([][]float64, len(solverSizes)-1)

	for l := 0; l < len(solverSizes)-1; l++ {
		inSize, outSize := solverSizes[l], solverSizes[l+1]
		solverWeights[l] = make([][]float64, inSize)
		solverBiases[l] = make([]float64, outSize)

		scale := math.Sqrt(2.0 / float64(inSize))
		for i := range solverWeights[l] {
			solverWeights[l][i] = make([]float64, outSize)
			for j := range solverWeights[l][i] {
				solverWeights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}

	return &GenerativeReplayCL{
		Config:         config,
		Generator:      generator,
		SolverWeights:  solverWeights,
		SolverBiases:   solverBiases,
		ExemplarMemory: make([][]float64, 0),
		ExemplarLabels: make([]int, 0),
	}
}

// LearnTask trains on a new task with generative replay.
func (gr *GenerativeReplayCL) LearnTask(taskName string, data [][]float64, labels []int) {
	gr.TaskLabels = append(gr.TaskLabels, taskName)

	// Generate replay samples from previous tasks
	replayCount := int(float64(len(data)) * gr.Config.ReplayRatio)
	replaySamples := gr.Generator.Generate(replayCount)

	// Combine new data with replay
	combinedData := make([][]float64, len(data)+replayCount)
	copy(combinedData, data)
	copy(combinedData[len(data):], replaySamples)

	// Combined labels (replay samples get pseudo-labels from solver)
	combinedLabels := make([]int, len(combinedData))
	copy(combinedLabels, labels)

	for i := len(data); i < len(combinedData); i++ {
		// Generate pseudo-label by forward pass
		output := gr.Forward(combinedData[i])
		maxIdx := 0
		maxVal := output[0]
		for j := 1; j < len(output); j++ {
			if output[j] > maxVal {
				maxVal = output[j]
				maxIdx = j
			}
		}
		combinedLabels[i] = maxIdx
	}

	// Simulated training step (simplified gradient update)
	lr := 0.01
	for epoch := 0; epoch < 10; epoch++ {
		for i := range combinedData {
			// Forward pass
			output := gr.Forward(combinedData[i])

			// Compute error
			target := make([]float64, len(output))
			if combinedLabels[i] < len(target) {
				target[combinedLabels[i]] = 1.0
			}

			// Simple weight update (gradient descent approximation)
			for j := range output {
				error := target[j] - output[j]
				// Update last layer
				lastLayer := len(gr.SolverWeights) - 1
				for k := range gr.SolverWeights[lastLayer] {
					if k < len(combinedData[i]) {
						gr.SolverWeights[lastLayer][k][j] += lr * error * combinedData[i][k]
					}
				}
			}
		}
	}

	// Store exemplars (for hybrid approach)
	for i := 0; i < len(data) && len(gr.ExemplarMemory) < gr.Config.MaxStoredSamples; i++ {
		gr.ExemplarMemory = append(gr.ExemplarMemory, data[i])
		gr.ExemplarLabels = append(gr.ExemplarLabels, labels[i])
	}

	gr.TaskCount++
}

// Forward performs forward pass through solver.
func (gr *GenerativeReplayCL) Forward(input []float64) []float64 {
	current := input

	for l := 0; l < len(gr.SolverWeights); l++ {
		inSize := len(gr.SolverWeights[l])
		outSize := len(gr.SolverWeights[l][0])
		output := make([]float64, outSize)

		for j := 0; j < outSize; j++ {
			sum := gr.SolverBiases[l][j]
			for i := 0; i < inSize && i < len(current); i++ {
				sum += current[i] * gr.SolverWeights[l][i][j]
			}

			// ReLU for hidden, softmax prep for output
			if l < len(gr.SolverWeights)-1 {
				if sum < 0 {
					sum = 0
				}
			}
			output[j] = sum
		}

		// Softmax for last layer
		if l == len(gr.SolverWeights)-1 {
			maxVal := output[0]
			for _, v := range output {
				if v > maxVal {
					maxVal = v
				}
			}
			sum := 0.0
			for j := range output {
				output[j] = math.Exp(output[j] - maxVal)
				sum += output[j]
			}
			for j := range output {
				output[j] /= sum
			}
		}

		current = output
	}

	return current
}

// Predict returns the predicted class for an input.
func (gr *GenerativeReplayCL) Predict(input []float64) int {
	output := gr.Forward(input)
	maxIdx := 0
	maxVal := output[0]
	for j := 1; j < len(output); j++ {
		if output[j] > maxVal {
			maxVal = output[j]
			maxIdx = j
		}
	}
	return maxIdx
}

// =============================================================================
// EXPERIENCE REPLAY BUFFER
// Memory-based replay for edge devices
// =============================================================================

// ReplayBufferConfig configures experience replay buffer.
type ReplayBufferConfig struct {
	MaxSize         int
	SamplingStrategy string // "uniform", "prioritized", "reservoir"
	PriorityAlpha   float64
	PriorityBeta    float64
}

// ReplayBuffer implements experience replay memory.
type ReplayBuffer struct {
	Config *ReplayBufferConfig

	// Storage
	Samples   [][]float64
	Labels    []int
	Priorities []float64
	TaskIDs   []int

	// Reservoir sampling state
	TotalSeen int
}

// NewReplayBuffer creates a new replay buffer.
func NewReplayBuffer(config *ReplayBufferConfig) *ReplayBuffer {
	if config == nil {
		config = &ReplayBufferConfig{
			MaxSize:          1000,
			SamplingStrategy: "reservoir",
			PriorityAlpha:    0.6,
			PriorityBeta:     0.4,
		}
	}

	return &ReplayBuffer{
		Config:     config,
		Samples:    make([][]float64, 0, config.MaxSize),
		Labels:     make([]int, 0, config.MaxSize),
		Priorities: make([]float64, 0, config.MaxSize),
		TaskIDs:    make([]int, 0, config.MaxSize),
	}
}

// Add adds a sample to the buffer.
func (rb *ReplayBuffer) Add(sample []float64, label, taskID int) {
	rb.TotalSeen++

	switch rb.Config.SamplingStrategy {
	case "reservoir":
		// Reservoir sampling: equal probability for all seen samples
		if len(rb.Samples) < rb.Config.MaxSize {
			rb.Samples = append(rb.Samples, sample)
			rb.Labels = append(rb.Labels, label)
			rb.Priorities = append(rb.Priorities, 1.0)
			rb.TaskIDs = append(rb.TaskIDs, taskID)
		} else {
			// Replace with probability maxSize/totalSeen
			if rand.Float64() < float64(rb.Config.MaxSize)/float64(rb.TotalSeen) {
				idx := rand.Intn(rb.Config.MaxSize)
				rb.Samples[idx] = sample
				rb.Labels[idx] = label
				rb.TaskIDs[idx] = taskID
			}
		}

	default: // "uniform" or FIFO
		if len(rb.Samples) < rb.Config.MaxSize {
			rb.Samples = append(rb.Samples, sample)
			rb.Labels = append(rb.Labels, label)
			rb.Priorities = append(rb.Priorities, 1.0)
			rb.TaskIDs = append(rb.TaskIDs, taskID)
		} else {
			// FIFO replacement
			idx := rb.TotalSeen % rb.Config.MaxSize
			rb.Samples[idx] = sample
			rb.Labels[idx] = label
			rb.TaskIDs[idx] = taskID
		}
	}
}

// Sample retrieves a batch of samples.
func (rb *ReplayBuffer) Sample(batchSize int) ([][]float64, []int) {
	if len(rb.Samples) == 0 {
		return nil, nil
	}

	if batchSize > len(rb.Samples) {
		batchSize = len(rb.Samples)
	}

	samples := make([][]float64, batchSize)
	labels := make([]int, batchSize)

	// Uniform random sampling
	for i := 0; i < batchSize; i++ {
		idx := rand.Intn(len(rb.Samples))
		samples[i] = rb.Samples[idx]
		labels[i] = rb.Labels[idx]
	}

	return samples, labels
}

// GetTaskDistribution returns the distribution of tasks in buffer.
func (rb *ReplayBuffer) GetTaskDistribution() map[int]int {
	dist := make(map[int]int)
	for _, tid := range rb.TaskIDs {
		dist[tid]++
	}
	return dist
}

// =============================================================================
// IN-SENSOR CONTINUAL LEARNING SYSTEM
// Combines in-sensor computing with continual learning
// =============================================================================

// InSensorCLSystem combines in-sensor preprocessing with CL.
type InSensorCLSystem struct {
	// In-sensor components
	SensorArray *RetinomorphicArray
	InSensorNN  *InSensorNN

	// Continual learning components
	GenerativeReplay *GenerativeReplayCL
	ReplayBuffer     *ReplayBuffer

	// Task management
	CurrentTask int
	TaskNames   []string

	// Performance tracking
	TaskAccuracies map[string]float64
}

// NewInSensorCLSystem creates an integrated in-sensor CL system.
func NewInSensorCLSystem(sensorWidth, sensorHeight, numClasses int) *InSensorCLSystem {
	pixelConfig := &RetinomorphicPixelConfig{
		Width:               sensorWidth,
		Height:              sensorHeight,
		EnableEdgeDetection: true,
		EnableMotionDetection: true,
		EnableContrastEnhance: true,
		SurroundRadius:      1,
	}

	inputSize := sensorWidth * sensorHeight

	return &InSensorCLSystem{
		SensorArray:      NewRetinomorphicArray(pixelConfig),
		InSensorNN:       NewInSensorNN(&InSensorNNConfig{InputSize: inputSize, OutputSize: numClasses}),
		GenerativeReplay: NewGenerativeReplayCL(inputSize, numClasses, nil),
		ReplayBuffer:     NewReplayBuffer(nil),
		TaskAccuracies:   make(map[string]float64),
	}
}

// ProcessAndLearn processes sensor input and updates CL model.
func (sys *InSensorCLSystem) ProcessAndLearn(image [][][]float64, label int, timeMs float64) []float64 {
	// In-sensor preprocessing
	sensorOutput := sys.SensorArray.ProcessFrame(image, timeMs)

	// Flatten adapted frame
	h, w := len(sensorOutput.AdaptedFrame), len(sensorOutput.AdaptedFrame[0])
	flatInput := make([]float64, h*w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			flatInput[y*w+x] = sensorOutput.AdaptedFrame[y][x]
		}
	}

	// Add to replay buffer
	sys.ReplayBuffer.Add(flatInput, label, sys.CurrentTask)

	// Forward pass
	output := sys.GenerativeReplay.Forward(flatInput)

	return output
}

// StartNewTask initializes a new task.
func (sys *InSensorCLSystem) StartNewTask(taskName string) {
	sys.TaskNames = append(sys.TaskNames, taskName)
	sys.CurrentTask++
}

// GetSystemStats returns system statistics.
func (sys *InSensorCLSystem) GetSystemStats() *InSensorCLStats {
	return &InSensorCLStats{
		NumTasks:         sys.CurrentTask,
		BufferSize:       len(sys.ReplayBuffer.Samples),
		TaskDistribution: sys.ReplayBuffer.GetTaskDistribution(),
		TaskAccuracies:   sys.TaskAccuracies,
	}
}

// InSensorCLStats contains system statistics.
type InSensorCLStats struct {
	NumTasks         int
	BufferSize       int
	TaskDistribution map[int]int
	TaskAccuracies   map[string]float64
}

// =============================================================================
// ENERGY AND PERFORMANCE ESTIMATION
// =============================================================================

// EstimateInSensorEnergy estimates energy for in-sensor system.
func EstimateInSensorEnergy(sensorWidth, sensorHeight int, numMACs int64) *InSensorEnergyEstimate {
	// Based on literature values
	senseEnergyPerPixel := 1.0   // fJ
	macEnergyFeFET := 10.0       // fJ
	dataMovementSaved := 100.0   // fJ per pixel (vs off-chip)

	totalSenseEnergy := float64(sensorWidth*sensorHeight) * senseEnergyPerPixel
	totalMACEnergy := float64(numMACs) * macEnergyFeFET
	savedEnergy := float64(sensorWidth*sensorHeight) * dataMovementSaved

	return &InSensorEnergyEstimate{
		SensingEnergy:     totalSenseEnergy,
		ComputeEnergy:     totalMACEnergy,
		TotalEnergy:       totalSenseEnergy + totalMACEnergy,
		DataMovementSaved: savedEnergy,
		Efficiency:        savedEnergy / (totalSenseEnergy + totalMACEnergy),
	}
}

// InSensorEnergyEstimate contains energy estimation results.
type InSensorEnergyEstimate struct {
	SensingEnergy     float64 // fJ
	ComputeEnergy     float64 // fJ
	TotalEnergy       float64 // fJ
	DataMovementSaved float64 // fJ
	Efficiency        float64 // Ratio of saved/used
}
