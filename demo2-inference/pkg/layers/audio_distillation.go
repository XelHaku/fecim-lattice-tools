// Package layers provides neural network layer implementations for CIM deployment.
// This file implements neuromorphic audio processing and knowledge distillation.
// Based on research: Nature Electronics (MEMS cochlea), IEEE (KD for memristor),
// Nano Letters (HZO SNN), arXiv (integrated gradients KD)
package layers

import (
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// NEUROMORPHIC COCHLEA SIMULATION
// Bio-inspired frequency-selective audio processing
// =============================================================================

// CochleaChannelConfig configures a single cochlea frequency channel.
type CochleaChannelConfig struct {
	CenterFrequency float64 // Hz
	Bandwidth       float64 // Hz (Q factor determines this)
	QFactor         float64 // Quality factor

	// Ferroelectric parameters
	UseFeFET        bool
	AdaptationTau   float64 // ms
	ThresholdLevel  float64
}

// CochleaConfig configures the full cochlea model.
type CochleaConfig struct {
	NumChannels     int
	MinFrequency    float64 // Hz
	MaxFrequency    float64 // Hz
	SampleRate      float64 // Hz

	// Processing options
	UseNonlinear    bool    // Compressive nonlinearity
	CompressionExp  float64 // Compression exponent (0.3-0.5 typical)

	// Spike encoding
	SpikeThreshold  float64
	RefractoryMs    float64

	// FeFET integration
	UseFeFETFilters bool
}

// DefaultCochleaConfig returns default cochlea configuration.
func DefaultCochleaConfig() *CochleaConfig {
	return &CochleaConfig{
		NumChannels:     64,
		MinFrequency:    20.0,
		MaxFrequency:    8000.0,
		SampleRate:      16000.0,
		UseNonlinear:    true,
		CompressionExp:  0.4,
		SpikeThreshold:  0.5,
		RefractoryMs:    1.0,
		UseFeFETFilters: true,
	}
}

// CochleaChannel implements a single frequency channel.
type CochleaChannel struct {
	Config *CochleaChannelConfig

	// Filter state (2nd order bandpass)
	FilterState [2]float64

	// Envelope detection
	Envelope        float64
	EnvelopeDecay   float64

	// Spike generation
	MembranePot     float64
	LastSpikeTime   float64
	InRefractory    bool

	// FeFET adaptation state
	AdaptationLevel float64
	Polarization    float64
}

// NeuromorphicCochlea implements bio-inspired audio front-end.
type NeuromorphicCochlea struct {
	Config   *CochleaConfig
	Channels []*CochleaChannel

	// Time tracking
	CurrentTime float64
	SampleDt    float64

	// Output buffers
	FilteredOutput  []float64
	EnvelopeOutput  []float64
	SpikeOutput     []bool
}

// NewNeuromorphicCochlea creates a new cochlea model.
func NewNeuromorphicCochlea(config *CochleaConfig) *NeuromorphicCochlea {
	if config == nil {
		config = DefaultCochleaConfig()
	}

	channels := make([]*CochleaChannel, config.NumChannels)

	// Create channels with log-spaced frequencies (like biological cochlea)
	logMin := math.Log10(config.MinFrequency)
	logMax := math.Log10(config.MaxFrequency)
	logStep := (logMax - logMin) / float64(config.NumChannels-1)

	for i := 0; i < config.NumChannels; i++ {
		freq := math.Pow(10, logMin+float64(i)*logStep)
		qFactor := 5.0 + float64(i)*0.1 // Higher Q at higher frequencies

		channels[i] = &CochleaChannel{
			Config: &CochleaChannelConfig{
				CenterFrequency: freq,
				Bandwidth:       freq / qFactor,
				QFactor:         qFactor,
				UseFeFET:        config.UseFeFETFilters,
				AdaptationTau:   100.0,
				ThresholdLevel:  config.SpikeThreshold,
			},
			EnvelopeDecay:   0.99,
			AdaptationLevel: 0.5,
		}
	}

	return &NeuromorphicCochlea{
		Config:         config,
		Channels:       channels,
		SampleDt:       1000.0 / config.SampleRate, // ms
		FilteredOutput: make([]float64, config.NumChannels),
		EnvelopeOutput: make([]float64, config.NumChannels),
		SpikeOutput:    make([]bool, config.NumChannels),
	}
}

// ProcessSample processes a single audio sample.
func (c *NeuromorphicCochlea) ProcessSample(sample float64) *CochleaOutput {
	c.CurrentTime += c.SampleDt

	for i, ch := range c.Channels {
		// Bandpass filter (2nd order IIR approximation)
		omega := 2.0 * math.Pi * ch.Config.CenterFrequency / c.Config.SampleRate
		alpha := math.Sin(omega) / (2.0 * ch.Config.QFactor)

		// Filter coefficients
		b0 := alpha
		b1 := 0.0
		b2 := -alpha
		a0 := 1.0 + alpha
		a1 := -2.0 * math.Cos(omega)
		a2 := 1.0 - alpha

		// Apply filter
		filtered := (b0/a0)*sample + ch.FilterState[0]
		ch.FilterState[0] = (b1/a0)*sample - (a1/a0)*filtered + ch.FilterState[1]
		ch.FilterState[1] = (b2/a0)*sample - (a2/a0)*filtered

		// Compressive nonlinearity (like hair cells)
		if c.Config.UseNonlinear {
			sign := 1.0
			if filtered < 0 {
				sign = -1.0
			}
			filtered = sign * math.Pow(math.Abs(filtered), c.Config.CompressionExp)
		}

		// Half-wave rectification and envelope detection
		rectified := filtered
		if rectified < 0 {
			rectified = 0
		}
		ch.Envelope = ch.EnvelopeDecay*ch.Envelope + (1-ch.EnvelopeDecay)*rectified

		// FeFET adaptation
		if ch.Config.UseFeFET {
			adaptDt := c.SampleDt / ch.Config.AdaptationTau
			ch.AdaptationLevel = (1-adaptDt)*ch.AdaptationLevel + adaptDt*ch.Envelope
			ch.Polarization = ch.AdaptationLevel

			// Adaptive threshold
			effectiveThreshold := ch.Config.ThresholdLevel * (1.0 + ch.Polarization)
			ch.Config.ThresholdLevel = effectiveThreshold
		}

		// Spike generation (integrate-and-fire)
		ch.MembranePot += ch.Envelope * c.SampleDt

		c.SpikeOutput[i] = false
		if !ch.InRefractory && ch.MembranePot > ch.Config.ThresholdLevel {
			c.SpikeOutput[i] = true
			ch.MembranePot = 0
			ch.LastSpikeTime = c.CurrentTime
			ch.InRefractory = true
		}

		// Refractory period
		if ch.InRefractory && (c.CurrentTime-ch.LastSpikeTime) > c.Config.RefractoryMs {
			ch.InRefractory = false
		}

		c.FilteredOutput[i] = filtered
		c.EnvelopeOutput[i] = ch.Envelope
	}

	return &CochleaOutput{
		FilterBank: append([]float64{}, c.FilteredOutput...),
		Envelopes:  append([]float64{}, c.EnvelopeOutput...),
		Spikes:     append([]bool{}, c.SpikeOutput...),
		Time:       c.CurrentTime,
	}
}

// ProcessFrame processes a frame of audio samples.
func (c *NeuromorphicCochlea) ProcessFrame(samples []float64) []*CochleaOutput {
	outputs := make([]*CochleaOutput, len(samples))
	for i, sample := range samples {
		outputs[i] = c.ProcessSample(sample)
	}
	return outputs
}

// CochleaOutput contains cochlea processing output.
type CochleaOutput struct {
	FilterBank []float64 // Per-channel filtered output
	Envelopes  []float64 // Per-channel envelope
	Spikes     []bool    // Per-channel spike events
	Time       float64   // Current time (ms)
}

// GetSpikeRaster converts spike outputs to raster format.
func (c *NeuromorphicCochlea) GetSpikeRaster(outputs []*CochleaOutput) [][]SpikeEvent {
	raster := make([][]SpikeEvent, c.Config.NumChannels)
	for i := range raster {
		raster[i] = make([]SpikeEvent, 0)
	}

	for _, out := range outputs {
		for ch, spiked := range out.Spikes {
			if spiked {
				raster[ch] = append(raster[ch], SpikeEvent{
					Channel: ch,
					Time:    out.Time,
				})
			}
		}
	}

	return raster
}

// SpikeEvent represents a single spike.
type SpikeEvent struct {
	Channel int
	Time    float64
}

// =============================================================================
// AUDIO FEATURE EXTRACTION FOR CIM
// =============================================================================

// AudioFeatureConfig configures audio feature extraction.
type AudioFeatureConfig struct {
	NumMelBins    int
	NumMFCCs      int
	FrameSize     int
	HopSize       int
	UseDelta      bool
	UseDeltaDelta bool
}

// AudioFeatureExtractor extracts features for CIM inference.
type AudioFeatureExtractor struct {
	Config  *AudioFeatureConfig
	Cochlea *NeuromorphicCochlea

	// Mel filterbank (precomputed)
	MelFilters [][]float64

	// DCT matrix for MFCC
	DCTMatrix [][]float64
}

// NewAudioFeatureExtractor creates an audio feature extractor.
func NewAudioFeatureExtractor(config *AudioFeatureConfig, cochlea *NeuromorphicCochlea) *AudioFeatureExtractor {
	if config == nil {
		config = &AudioFeatureConfig{
			NumMelBins:    40,
			NumMFCCs:      13,
			FrameSize:     400,
			HopSize:       160,
			UseDelta:      true,
			UseDeltaDelta: true,
		}
	}

	// Build mel filterbank
	numFilters := config.NumMelBins
	numChannels := cochlea.Config.NumChannels
	melFilters := make([][]float64, numFilters)

	for i := 0; i < numFilters; i++ {
		melFilters[i] = make([]float64, numChannels)
		// Triangular filters (simplified)
		center := float64(i) / float64(numFilters-1) * float64(numChannels-1)
		width := float64(numChannels) / float64(numFilters) * 2

		for j := 0; j < numChannels; j++ {
			dist := math.Abs(float64(j) - center)
			if dist < width {
				melFilters[i][j] = 1.0 - dist/width
			}
		}
	}

	// Build DCT matrix
	dctMatrix := make([][]float64, config.NumMFCCs)
	for i := 0; i < config.NumMFCCs; i++ {
		dctMatrix[i] = make([]float64, numFilters)
		for j := 0; j < numFilters; j++ {
			dctMatrix[i][j] = math.Cos(math.Pi * float64(i) * (float64(j) + 0.5) / float64(numFilters))
		}
	}

	return &AudioFeatureExtractor{
		Config:     config,
		Cochlea:    cochlea,
		MelFilters: melFilters,
		DCTMatrix:  dctMatrix,
	}
}

// ExtractFeatures extracts MFCC-like features from cochlea output.
func (afe *AudioFeatureExtractor) ExtractFeatures(envelopes []float64) []float64 {
	// Apply mel filterbank
	melEnergies := make([]float64, afe.Config.NumMelBins)
	for i := 0; i < afe.Config.NumMelBins; i++ {
		for j := 0; j < len(envelopes) && j < len(afe.MelFilters[i]); j++ {
			melEnergies[i] += afe.MelFilters[i][j] * envelopes[j]
		}
		// Log compression
		melEnergies[i] = math.Log(melEnergies[i] + 1e-10)
	}

	// Apply DCT to get MFCCs
	mfccs := make([]float64, afe.Config.NumMFCCs)
	for i := 0; i < afe.Config.NumMFCCs; i++ {
		for j := 0; j < afe.Config.NumMelBins; j++ {
			mfccs[i] += afe.DCTMatrix[i][j] * melEnergies[j]
		}
	}

	return mfccs
}

// =============================================================================
// AUDIO CLASSIFICATION SNN
// Spiking neural network for audio classification
// =============================================================================

// AudioSNNConfig configures audio classification SNN.
type AudioSNNConfig struct {
	InputSize     int
	HiddenSizes   []int
	OutputSize    int

	// SNN parameters
	Tau           float64 // Membrane time constant (ms)
	Threshold     float64 // Spike threshold
	Reset         float64 // Reset potential

	// FeFET synapse parameters
	UseFeFETSynapses bool
	SynapseOnOff     float64
}

// AudioSNN implements spiking neural network for audio.
type AudioSNN struct {
	Config *AudioSNNConfig

	// Weights
	Weights [][][]float64

	// Neuron states
	MembranePotentials [][]float64
	SpikeStates        [][]bool

	// Time tracking
	CurrentTime float64
	DtMs        float64
}

// NewAudioSNN creates a new audio SNN.
func NewAudioSNN(config *AudioSNNConfig) *AudioSNN {
	if config == nil {
		config = &AudioSNNConfig{
			InputSize:        64,
			HiddenSizes:      []int{128, 64},
			OutputSize:       10,
			Tau:              20.0,
			Threshold:        1.0,
			Reset:            0.0,
			UseFeFETSynapses: true,
			SynapseOnOff:     1e4,
		}
	}

	// Build layer sizes
	sizes := append([]int{config.InputSize}, config.HiddenSizes...)
	sizes = append(sizes, config.OutputSize)

	// Initialize weights
	weights := make([][][]float64, len(sizes)-1)
	for l := 0; l < len(sizes)-1; l++ {
		inSize, outSize := sizes[l], sizes[l+1]
		weights[l] = make([][]float64, inSize)
		scale := math.Sqrt(2.0 / float64(inSize))
		for i := range weights[l] {
			weights[l][i] = make([]float64, outSize)
			for j := range weights[l][i] {
				weights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}

	// Initialize neuron states
	membranePots := make([][]float64, len(sizes)-1)
	spikeStates := make([][]bool, len(sizes)-1)
	for l := 0; l < len(sizes)-1; l++ {
		membranePots[l] = make([]float64, sizes[l+1])
		spikeStates[l] = make([]bool, sizes[l+1])
	}

	return &AudioSNN{
		Config:             config,
		Weights:            weights,
		MembranePotentials: membranePots,
		SpikeStates:        spikeStates,
		DtMs:               1.0,
	}
}

// Forward processes input spikes through the SNN.
func (snn *AudioSNN) Forward(inputSpikes []bool) []float64 {
	snn.CurrentTime += snn.DtMs

	// Convert input spikes to currents
	inputCurrents := make([]float64, len(inputSpikes))
	for i, spike := range inputSpikes {
		if spike {
			inputCurrents[i] = 1.0
		}
	}

	current := inputCurrents

	// Process through layers
	for l := 0; l < len(snn.Weights); l++ {
		inSize := len(snn.Weights[l])
		outSize := len(snn.Weights[l][0])
		nextCurrents := make([]float64, outSize)

		// Synaptic integration
		for j := 0; j < outSize; j++ {
			synapticInput := 0.0
			for i := 0; i < inSize && i < len(current); i++ {
				synapticInput += current[i] * snn.Weights[l][i][j]
			}

			// Leaky integrate-and-fire
			decay := math.Exp(-snn.DtMs / snn.Config.Tau)
			snn.MembranePotentials[l][j] = decay*snn.MembranePotentials[l][j] + synapticInput

			// Spike generation
			snn.SpikeStates[l][j] = false
			if snn.MembranePotentials[l][j] > snn.Config.Threshold {
				snn.SpikeStates[l][j] = true
				nextCurrents[j] = 1.0
				snn.MembranePotentials[l][j] = snn.Config.Reset
			}
		}

		current = nextCurrents
	}

	// Return output layer membrane potentials (for classification)
	return snn.MembranePotentials[len(snn.MembranePotentials)-1]
}

// Reset resets all neuron states.
func (snn *AudioSNN) Reset() {
	for l := range snn.MembranePotentials {
		for j := range snn.MembranePotentials[l] {
			snn.MembranePotentials[l][j] = 0
			snn.SpikeStates[l][j] = false
		}
	}
	snn.CurrentTime = 0
}

// =============================================================================
// KNOWLEDGE DISTILLATION FOR CIM
// Hardware-aware knowledge transfer from teacher to student
// =============================================================================

// KDConfig configures knowledge distillation.
type KDConfig struct {
	// Temperature for softening
	Temperature float64

	// Loss weights
	AlphaHard    float64 // Weight for hard label loss
	AlphaSoft    float64 // Weight for soft label (KD) loss
	AlphaFeature float64 // Weight for feature matching

	// Hardware-aware parameters
	TargetBits       int     // Target quantization bits
	VariationSigma   float64 // Device variation for training

	// Integrated gradients
	UseIntegratedGrad bool
	NumIGSteps        int
}

// DefaultKDConfig returns default KD configuration.
func DefaultKDConfig() *KDConfig {
	return &KDConfig{
		Temperature:       4.0,
		AlphaHard:         0.5,
		AlphaSoft:         0.5,
		AlphaFeature:      0.1,
		TargetBits:        6,
		VariationSigma:    0.02,
		UseIntegratedGrad: true,
		NumIGSteps:        50,
	}
}

// KnowledgeDistiller implements knowledge distillation for CIM.
type KnowledgeDistiller struct {
	Config *KDConfig

	// Teacher model (full precision)
	TeacherWeights [][][]float64
	TeacherBiases  [][]float64

	// Student model (quantized)
	StudentWeights [][][]float64
	StudentBiases  [][]float64

	// Intermediate features
	TeacherFeatures [][]float64
	StudentFeatures [][]float64

	// Integrated gradients cache
	IGCache map[string][][]float64
}

// NewKnowledgeDistiller creates a new distiller.
func NewKnowledgeDistiller(teacherSizes []int, config *KDConfig) *KnowledgeDistiller {
	if config == nil {
		config = DefaultKDConfig()
	}

	// Initialize teacher weights
	teacherWeights := make([][][]float64, len(teacherSizes)-1)
	teacherBiases := make([][]float64, len(teacherSizes)-1)

	for l := 0; l < len(teacherSizes)-1; l++ {
		inSize, outSize := teacherSizes[l], teacherSizes[l+1]
		teacherWeights[l] = make([][]float64, inSize)
		teacherBiases[l] = make([]float64, outSize)

		scale := math.Sqrt(2.0 / float64(inSize))
		for i := range teacherWeights[l] {
			teacherWeights[l][i] = make([]float64, outSize)
			for j := range teacherWeights[l][i] {
				teacherWeights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}

	// Student will be created during distillation
	return &KnowledgeDistiller{
		Config:         config,
		TeacherWeights: teacherWeights,
		TeacherBiases:  teacherBiases,
		IGCache:        make(map[string][][]float64),
	}
}

// TeacherForward performs forward pass through teacher.
func (kd *KnowledgeDistiller) TeacherForward(input []float64) ([]float64, [][]float64) {
	current := input
	features := make([][]float64, len(kd.TeacherWeights))

	for l := 0; l < len(kd.TeacherWeights); l++ {
		inSize := len(kd.TeacherWeights[l])
		outSize := len(kd.TeacherWeights[l][0])
		output := make([]float64, outSize)

		for j := 0; j < outSize; j++ {
			sum := kd.TeacherBiases[l][j]
			for i := 0; i < inSize && i < len(current); i++ {
				sum += current[i] * kd.TeacherWeights[l][i][j]
			}

			// ReLU for hidden layers
			if l < len(kd.TeacherWeights)-1 && sum < 0 {
				sum = 0
			}
			output[j] = sum
		}

		features[l] = output
		current = output
	}

	return current, features
}

// SoftmaxWithTemperature computes softmax with temperature.
func SoftmaxWithTemperature(logits []float64, temperature float64) []float64 {
	scaled := make([]float64, len(logits))
	maxVal := logits[0]
	for _, v := range logits {
		if v > maxVal {
			maxVal = v
		}
	}

	sum := 0.0
	for i, v := range logits {
		scaled[i] = math.Exp((v - maxVal) / temperature)
		sum += scaled[i]
	}

	for i := range scaled {
		scaled[i] /= sum
	}

	return scaled
}

// KLDivergence computes KL divergence between two distributions.
func KLDivergence(p, q []float64) float64 {
	kl := 0.0
	for i := range p {
		if p[i] > 1e-10 && q[i] > 1e-10 {
			kl += p[i] * math.Log(p[i]/q[i])
		}
	}
	return kl
}

// ComputeIntegratedGradients computes IG for feature importance.
func (kd *KnowledgeDistiller) ComputeIntegratedGradients(input []float64, targetClass int) []float64 {
	baseline := make([]float64, len(input))
	ig := make([]float64, len(input))

	for step := 0; step < kd.Config.NumIGSteps; step++ {
		alpha := float64(step) / float64(kd.Config.NumIGSteps)

		// Interpolated input
		interpolated := make([]float64, len(input))
		for i := range input {
			interpolated[i] = baseline[i] + alpha*(input[i]-baseline[i])
		}

		// Forward pass
		output, _ := kd.TeacherForward(interpolated)

		// Approximate gradient (finite difference)
		delta := 0.001
		for i := range input {
			inputPlus := make([]float64, len(input))
			copy(inputPlus, interpolated)
			inputPlus[i] += delta

			outputPlus, _ := kd.TeacherForward(inputPlus)

			grad := (outputPlus[targetClass] - output[targetClass]) / delta
			ig[i] += grad * (input[i] - baseline[i]) / float64(kd.Config.NumIGSteps)
		}
	}

	return ig
}

// CreateStudentModel creates a compressed student model.
func (kd *KnowledgeDistiller) CreateStudentModel(compressionRatio float64) {
	kd.StudentWeights = make([][][]float64, len(kd.TeacherWeights))
	kd.StudentBiases = make([][]float64, len(kd.TeacherBiases))

	for l := 0; l < len(kd.TeacherWeights); l++ {
		inSize := len(kd.TeacherWeights[l])
		outSize := len(kd.TeacherWeights[l][0])

		// Compressed sizes
		studentInSize := int(float64(inSize) / compressionRatio)
		studentOutSize := int(float64(outSize) / compressionRatio)
		if studentInSize < 1 {
			studentInSize = 1
		}
		if studentOutSize < 1 {
			studentOutSize = 1
		}

		// First and last layer keep original I/O size
		if l == 0 {
			studentInSize = inSize
		}
		if l == len(kd.TeacherWeights)-1 {
			studentOutSize = outSize
		}

		kd.StudentWeights[l] = make([][]float64, studentInSize)
		kd.StudentBiases[l] = make([]float64, studentOutSize)

		scale := math.Sqrt(2.0 / float64(studentInSize))
		for i := range kd.StudentWeights[l] {
			kd.StudentWeights[l][i] = make([]float64, studentOutSize)
			for j := range kd.StudentWeights[l][i] {
				kd.StudentWeights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}
}

// StudentForward performs forward pass through student with variation.
func (kd *KnowledgeDistiller) StudentForward(input []float64, addVariation bool) ([]float64, [][]float64) {
	if kd.StudentWeights == nil {
		return nil, nil
	}

	current := input
	features := make([][]float64, len(kd.StudentWeights))

	for l := 0; l < len(kd.StudentWeights); l++ {
		inSize := len(kd.StudentWeights[l])
		outSize := len(kd.StudentWeights[l][0])
		output := make([]float64, outSize)

		for j := 0; j < outSize; j++ {
			sum := kd.StudentBiases[l][j]
			for i := 0; i < inSize && i < len(current); i++ {
				w := kd.StudentWeights[l][i][j]

				// Add device variation during training
				if addVariation {
					w *= (1.0 + kd.Config.VariationSigma*rand.NormFloat64())
				}

				sum += current[i] * w
			}

			// ReLU for hidden layers
			if l < len(kd.StudentWeights)-1 && sum < 0 {
				sum = 0
			}
			output[j] = sum
		}

		features[l] = output
		current = output
	}

	return current, features
}

// ComputeDistillationLoss computes the full KD loss.
func (kd *KnowledgeDistiller) ComputeDistillationLoss(input []float64, hardLabel int) float64 {
	// Teacher forward
	teacherLogits, teacherFeatures := kd.TeacherForward(input)
	teacherSoft := SoftmaxWithTemperature(teacherLogits, kd.Config.Temperature)

	// Student forward with variation
	studentLogits, studentFeatures := kd.StudentForward(input, true)
	if studentLogits == nil {
		return 0
	}
	studentSoft := SoftmaxWithTemperature(studentLogits, kd.Config.Temperature)
	studentHard := SoftmaxWithTemperature(studentLogits, 1.0)

	// Hard label loss (cross-entropy)
	hardLoss := 0.0
	if hardLabel < len(studentHard) {
		hardLoss = -math.Log(studentHard[hardLabel] + 1e-10)
	}

	// Soft label loss (KL divergence)
	softLoss := KLDivergence(teacherSoft, studentSoft) * kd.Config.Temperature * kd.Config.Temperature

	// Feature matching loss (MSE on intermediate features)
	featureLoss := 0.0
	for l := 0; l < len(teacherFeatures) && l < len(studentFeatures); l++ {
		tFeat := teacherFeatures[l]
		sFeat := studentFeatures[l]
		minLen := len(tFeat)
		if len(sFeat) < minLen {
			minLen = len(sFeat)
		}
		for i := 0; i < minLen; i++ {
			diff := tFeat[i] - sFeat[i]
			featureLoss += diff * diff
		}
	}
	if len(teacherFeatures) > 0 {
		featureLoss /= float64(len(teacherFeatures))
	}

	// Combined loss
	totalLoss := kd.Config.AlphaHard*hardLoss +
		kd.Config.AlphaSoft*softLoss +
		kd.Config.AlphaFeature*featureLoss

	return totalLoss
}

// QuantizeStudentWeights quantizes student weights for CIM.
func (kd *KnowledgeDistiller) QuantizeStudentWeights() {
	levels := float64(int(1) << kd.Config.TargetBits)

	for l := range kd.StudentWeights {
		// Find min/max for layer
		minW, maxW := kd.StudentWeights[l][0][0], kd.StudentWeights[l][0][0]
		for i := range kd.StudentWeights[l] {
			for j := range kd.StudentWeights[l][i] {
				if kd.StudentWeights[l][i][j] < minW {
					minW = kd.StudentWeights[l][i][j]
				}
				if kd.StudentWeights[l][i][j] > maxW {
					maxW = kd.StudentWeights[l][i][j]
				}
			}
		}

		// Quantize
		scale := (maxW - minW) / levels
		if scale == 0 {
			scale = 1
		}
		for i := range kd.StudentWeights[l] {
			for j := range kd.StudentWeights[l][i] {
				normalized := (kd.StudentWeights[l][i][j] - minW) / scale
				quantized := math.Round(normalized)
				kd.StudentWeights[l][i][j] = quantized*scale + minW
			}
		}
	}
}

// =============================================================================
// HARDWARE-AWARE KD TRAINER
// =============================================================================

// HardwareAwareKDConfig configures hardware-aware distillation.
type HardwareAwareKDConfig struct {
	*KDConfig

	// Hardware constraints
	MaxCrossbarSize  int
	ADCBits          int
	DACBits          int

	// Noise injection schedule
	NoiseSchedule    string // "constant", "linear", "cosine"
	InitialNoise     float64
	FinalNoise       float64

	// Pruning during distillation
	EnablePruning    bool
	TargetSparsity   float64
}

// HardwareAwareKDTrainer trains student with hardware constraints.
type HardwareAwareKDTrainer struct {
	Config    *HardwareAwareKDConfig
	Distiller *KnowledgeDistiller

	// Training state
	CurrentEpoch int
	TotalEpochs  int

	// Importance scores for pruning
	WeightImportance [][][]float64
}

// NewHardwareAwareKDTrainer creates a hardware-aware KD trainer.
func NewHardwareAwareKDTrainer(teacherSizes []int, config *HardwareAwareKDConfig) *HardwareAwareKDTrainer {
	if config == nil {
		config = &HardwareAwareKDConfig{
			KDConfig:        DefaultKDConfig(),
			MaxCrossbarSize: 256,
			ADCBits:         6,
			DACBits:         8,
			NoiseSchedule:   "cosine",
			InitialNoise:    0.01,
			FinalNoise:      0.05,
			EnablePruning:   true,
			TargetSparsity:  0.5,
		}
	}

	distiller := NewKnowledgeDistiller(teacherSizes, config.KDConfig)

	return &HardwareAwareKDTrainer{
		Config:    config,
		Distiller: distiller,
	}
}

// GetCurrentNoise returns noise level based on schedule.
func (t *HardwareAwareKDTrainer) GetCurrentNoise() float64 {
	if t.TotalEpochs == 0 {
		return t.Config.InitialNoise
	}

	progress := float64(t.CurrentEpoch) / float64(t.TotalEpochs)

	switch t.Config.NoiseSchedule {
	case "linear":
		return t.Config.InitialNoise + progress*(t.Config.FinalNoise-t.Config.InitialNoise)
	case "cosine":
		return t.Config.InitialNoise + 0.5*(t.Config.FinalNoise-t.Config.InitialNoise)*
			(1-math.Cos(progress*math.Pi))
	default:
		return t.Config.InitialNoise
	}
}

// PruneByImportance prunes weights based on importance scores.
func (t *HardwareAwareKDTrainer) PruneByImportance() {
	if t.Distiller.StudentWeights == nil || !t.Config.EnablePruning {
		return
	}

	// Collect all weight magnitudes
	type weightInfo struct {
		layer, i, j int
		magnitude   float64
	}
	allWeights := make([]weightInfo, 0)

	for l := range t.Distiller.StudentWeights {
		for i := range t.Distiller.StudentWeights[l] {
			for j := range t.Distiller.StudentWeights[l][i] {
				allWeights = append(allWeights, weightInfo{
					layer:     l,
					i:         i,
					j:         j,
					magnitude: math.Abs(t.Distiller.StudentWeights[l][i][j]),
				})
			}
		}
	}

	// Sort by magnitude
	sort.Slice(allWeights, func(a, b int) bool {
		return allWeights[a].magnitude < allWeights[b].magnitude
	})

	// Prune lowest magnitude weights
	numToPrune := int(float64(len(allWeights)) * t.Config.TargetSparsity)
	for i := 0; i < numToPrune; i++ {
		w := allWeights[i]
		t.Distiller.StudentWeights[w.layer][w.i][w.j] = 0
	}
}

// =============================================================================
// DISTILLATION METRICS
// =============================================================================

// KDMetrics contains distillation metrics.
type KDMetrics struct {
	TeacherAccuracy   float64
	StudentAccuracy   float64
	CompressionRatio  float64
	Sparsity          float64
	QuantizationBits  int
	ParameterCount    int64
	EstimatedEnergy   float64 // pJ per inference
}

// ComputeKDMetrics computes distillation metrics.
func ComputeKDMetrics(distiller *KnowledgeDistiller) *KDMetrics {
	// Count parameters
	var teacherParams, studentParams int64

	for l := range distiller.TeacherWeights {
		for i := range distiller.TeacherWeights[l] {
			teacherParams += int64(len(distiller.TeacherWeights[l][i]))
		}
	}

	if distiller.StudentWeights != nil {
		for l := range distiller.StudentWeights {
			for i := range distiller.StudentWeights[l] {
				studentParams += int64(len(distiller.StudentWeights[l][i]))
			}
		}
	}

	// Count zeros for sparsity
	var zeros int64
	if distiller.StudentWeights != nil {
		for l := range distiller.StudentWeights {
			for i := range distiller.StudentWeights[l] {
				for j := range distiller.StudentWeights[l][i] {
					if distiller.StudentWeights[l][i][j] == 0 {
						zeros++
					}
				}
			}
		}
	}

	sparsity := 0.0
	if studentParams > 0 {
		sparsity = float64(zeros) / float64(studentParams)
	}

	compressionRatio := 1.0
	if studentParams > 0 {
		compressionRatio = float64(teacherParams) / float64(studentParams)
	}

	// Estimate energy (simplified)
	macEnergy := 10.0 // fJ per MAC
	estimatedEnergy := float64(studentParams) * macEnergy * (1 - sparsity) / 1000.0 // pJ

	return &KDMetrics{
		CompressionRatio: compressionRatio,
		Sparsity:         sparsity,
		QuantizationBits: distiller.Config.TargetBits,
		ParameterCount:   studentParams,
		EstimatedEnergy:  estimatedEnergy,
	}
}
