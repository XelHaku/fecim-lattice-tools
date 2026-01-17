// sensory_compiler.go - Neuromorphic Sensory Systems and CIM Compiler/Mapping
// IronLattice Ferroelectric CIM Educational Simulation
//
// This module implements:
// 1. Neuromorphic Sensory Systems:
//    - Silicon retina / Dynamic Vision Sensor (DVS) simulation
//    - Event-based vision processing
//    - Silicon cochlea / acoustic processing
//    - Speech-to-spike encoding
//    - Bio-inspired preprocessing (Gammatone filterbank)
//
// 2. CIM Compiler and Mapping:
//    - Layer partitioning algorithms
//    - Weight replication strategies (PIMCOMP-style)
//    - Core mapping and scheduling
//    - Dataflow optimization
//    - Tiling for crossbar arrays
//
// References:
// - Mead & Mahowald (1992): Silicon retina
// - Delbruck et al. (2008): Dynamic Vision Sensor
// - Nature Electronics (2023): MEMS neuromorphic cochlea
// - PIMCOMP (DAC 2023): Universal compilation framework
// - CIM-MLC (2024): Multi-level compilation stack
// - COMPASS (2025): Resource-constrained crossbar compiler

package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// PART 1: NEUROMORPHIC SENSORY SYSTEMS
// =============================================================================

// DVSPixelConfig configures a single DVS pixel
type DVSPixelConfig struct {
	ContrastThreshold   float64 // Log intensity change threshold (typ. 15-25%)
	RefractoryPeriod    float64 // Minimum time between events (us)
	TemporalResolution  float64 // Time precision (us)
	DarkCurrentNoise    float64 // Noise in low light
	BandwidthHz         float64 // Pixel bandwidth
	PhotodiodeCapacitor float64 // Cp (fF)
}

// DefaultDVSPixelConfig returns typical DVS pixel parameters
func DefaultDVSPixelConfig() *DVSPixelConfig {
	return &DVSPixelConfig{
		ContrastThreshold:   0.20, // 20% contrast sensitivity
		RefractoryPeriod:    1.0,  // 1 us
		TemporalResolution:  1.0,  // 1 us precision
		DarkCurrentNoise:    0.01,
		BandwidthHz:         3e6,   // 3 MHz
		PhotodiodeCapacitor: 100.0, // 100 fF
	}
}

// DVSSensorConfig configures the entire DVS sensor
type DVSSensorConfig struct {
	ResolutionX       int              // Horizontal pixels
	ResolutionY       int              // Vertical pixels
	PixelPitch        float64          // um
	OpticalFormat     string           // "1/2", "1/2.5", "1/5"
	MaxEventRate      float64          // Maximum events per second
	DynamicRangeDB    float64          // HDR capability
	PowerConsumption  float64          // mW
	MinIlluminanceLux float64          // Minimum operating light
	PixelConfig       *DVSPixelConfig
}

// DVSSensorSpecs returns specifications for commercial DVS sensors
func DVSSensorSpecs(model string) *DVSSensorConfig {
	switch model {
	case "IMX636": // Sony/Prophesee HD
		return &DVSSensorConfig{
			ResolutionX:       1280,
			ResolutionY:       720,
			PixelPitch:        4.86,
			OpticalFormat:     "1/2",
			MaxEventRate:      1.066e9, // 1.066 Geps
			DynamicRangeDB:    124,
			PowerConsumption:  15.0, // Estimated
			MinIlluminanceLux: 0.1,
			PixelConfig:       DefaultDVSPixelConfig(),
		}
	case "GenX320": // Prophesee compact
		return &DVSSensorConfig{
			ResolutionX:       320,
			ResolutionY:       320,
			PixelPitch:        6.3,
			OpticalFormat:     "1/5",
			MaxEventRate:      100e6, // 100 Meps
			DynamicRangeDB:    140,
			PowerConsumption:  3.0, // 3 mW typical
			MinIlluminanceLux: 0.05,
			PixelConfig:       DefaultDVSPixelConfig(),
		}
	case "Samsung640": // Samsung DVS
		return &DVSSensorConfig{
			ResolutionX:       640,
			ResolutionY:       480,
			PixelPitch:        9.0,
			OpticalFormat:     "1/3",
			MaxEventRate:      300e6, // 300 Meps
			DynamicRangeDB:    120,
			PowerConsumption:  10.0,
			MinIlluminanceLux: 0.5,
			PixelConfig:       DefaultDVSPixelConfig(),
		}
	default: // Default research sensor
		return &DVSSensorConfig{
			ResolutionX:       128,
			ResolutionY:       128,
			PixelPitch:        10.0,
			OpticalFormat:     "custom",
			MaxEventRate:      10e6,
			DynamicRangeDB:    100,
			PowerConsumption:  5.0,
			MinIlluminanceLux: 1.0,
			PixelConfig:       DefaultDVSPixelConfig(),
		}
	}
}

// DVSEvent represents a single DVS event
type DVSEvent struct {
	X         int     // Pixel X coordinate
	Y         int     // Pixel Y coordinate
	Timestamp float64 // Time in microseconds
	Polarity  int     // +1 (ON) or -1 (OFF)
}

// DVSSensor simulates a Dynamic Vision Sensor
type DVSSensor struct {
	Config          *DVSSensorConfig
	PixelState      [][]float64   // Log intensity state
	LastEventTime   [][]float64   // Last event timestamp per pixel
	EventBuffer     []DVSEvent    // Output event buffer
	TotalEvents     int64
	PowerDissipated float64       // mJ

	// Statistics
	EventRateHz     float64
	DataReduction   float64       // vs frame-based
}

// NewDVSSensor creates a new DVS sensor
func NewDVSSensor(config *DVSSensorConfig) *DVSSensor {
	if config == nil {
		config = DVSSensorSpecs("default")
	}

	// Initialize pixel states
	pixelState := make([][]float64, config.ResolutionY)
	lastEventTime := make([][]float64, config.ResolutionY)
	for y := 0; y < config.ResolutionY; y++ {
		pixelState[y] = make([]float64, config.ResolutionX)
		lastEventTime[y] = make([]float64, config.ResolutionX)
	}

	return &DVSSensor{
		Config:        config,
		PixelState:    pixelState,
		LastEventTime: lastEventTime,
		EventBuffer:   make([]DVSEvent, 0, 10000),
	}
}

// ProcessFrame simulates DVS response to an intensity frame
func (dvs *DVSSensor) ProcessFrame(frame [][]float64, timestamp float64) []DVSEvent {
	dvs.EventBuffer = dvs.EventBuffer[:0]
	threshold := dvs.Config.PixelConfig.ContrastThreshold
	refractoryPeriod := dvs.Config.PixelConfig.RefractoryPeriod

	for y := 0; y < dvs.Config.ResolutionY && y < len(frame); y++ {
		for x := 0; x < dvs.Config.ResolutionX && x < len(frame[y]); x++ {
			// Compute log intensity
			intensity := frame[y][x]
			if intensity <= 0 {
				intensity = 1e-6
			}
			logIntensity := math.Log(intensity)

			// Check refractory period
			if timestamp-dvs.LastEventTime[y][x] < refractoryPeriod {
				continue
			}

			// Compute contrast change
			deltaLog := logIntensity - dvs.PixelState[y][x]

			// Generate events for significant changes
			if math.Abs(deltaLog) > threshold {
				polarity := 1
				if deltaLog < 0 {
					polarity = -1
				}

				// Add noise
				noise := rand.Float64() * dvs.Config.PixelConfig.DarkCurrentNoise

				event := DVSEvent{
					X:         x,
					Y:         y,
					Timestamp: timestamp + noise,
					Polarity:  polarity,
				}
				dvs.EventBuffer = append(dvs.EventBuffer, event)
				dvs.LastEventTime[y][x] = timestamp
				dvs.PixelState[y][x] = logIntensity
				dvs.TotalEvents++
			}
		}
	}

	// Update statistics
	totalPixels := float64(dvs.Config.ResolutionX * dvs.Config.ResolutionY)
	dvs.EventRateHz = float64(len(dvs.EventBuffer)) / (1e-6) // Events per frame
	dvs.DataReduction = totalPixels / float64(len(dvs.EventBuffer)+1)

	return dvs.EventBuffer
}

// ConvertToSpikes converts DVS events to spike trains for SNN processing
func (dvs *DVSSensor) ConvertToSpikes(events []DVSEvent, numNeurons int, timeWindow float64) [][]float64 {
	// Map pixel coordinates to neuron indices
	spikes := make([][]float64, numNeurons)
	for i := range spikes {
		spikes[i] = make([]float64, 0)
	}

	pixelsPerNeuron := (dvs.Config.ResolutionX * dvs.Config.ResolutionY) / numNeurons
	if pixelsPerNeuron < 1 {
		pixelsPerNeuron = 1
	}

	for _, event := range events {
		pixelIdx := event.Y*dvs.Config.ResolutionX + event.X
		neuronIdx := pixelIdx / pixelsPerNeuron
		if neuronIdx >= numNeurons {
			neuronIdx = numNeurons - 1
		}

		if event.Timestamp < timeWindow {
			spikes[neuronIdx] = append(spikes[neuronIdx], event.Timestamp)
		}
	}

	return spikes
}

// =============================================================================
// Silicon Cochlea
// =============================================================================

// GammatoneFilterConfig configures a gammatone filter for cochlear simulation
type GammatoneFilterConfig struct {
	CenterFreq  float64 // Center frequency (Hz)
	Bandwidth   float64 // ERB bandwidth
	Order       int     // Filter order (typically 4)
	SampleRate  float64 // Sample rate (Hz)
}

// CochleaConfig configures the silicon cochlea
type CochleaConfig struct {
	NumChannels      int              // Number of frequency channels (typ. 64-256)
	MinFreq          float64          // Minimum frequency (Hz)
	MaxFreq          float64          // Maximum frequency (Hz)
	FilterOrder      int              // Gammatone filter order
	SampleRate       float64          // Audio sample rate
	Compression      float64          // Compressive nonlinearity exponent
	AdaptationTau    float64          // Adaptation time constant (ms)
	HalfWaveRectify  bool             // Half-wave rectification
	PowerConsumption float64          // mW
	DynamicRangeDB   float64          // dB
}

// DefaultCochleaConfig returns typical cochlea parameters
func DefaultCochleaConfig() *CochleaConfig {
	return &CochleaConfig{
		NumChannels:      64,
		MinFreq:          100.0,
		MaxFreq:          10000.0,
		FilterOrder:      4,
		SampleRate:       16000.0,
		Compression:      0.3, // Compressive nonlinearity
		AdaptationTau:    50.0,
		HalfWaveRectify:  true,
		PowerConsumption: 0.5, // 0.5 mW
		DynamicRangeDB:   120.0,
	}
}

// SiliconCochlea simulates a neuromorphic cochlea
type SiliconCochlea struct {
	Config          *CochleaConfig
	Filters         []GammatoneFilterConfig
	FilterState     [][]float64    // Filter state per channel
	AdaptationState []float64      // Adaptation per channel
	SpikeThreshold  []float64      // Per-channel threshold
	OutputSpikes    [][]float64    // Spike times per channel

	// Performance
	EnergyPerSpike  float64        // pJ
}

// NewSiliconCochlea creates a new silicon cochlea
func NewSiliconCochlea(config *CochleaConfig) *SiliconCochlea {
	if config == nil {
		config = DefaultCochleaConfig()
	}

	// Create frequency-spaced filters (ERB scale)
	filters := make([]GammatoneFilterConfig, config.NumChannels)
	freqs := erbSpace(config.MinFreq, config.MaxFreq, config.NumChannels)

	for i := 0; i < config.NumChannels; i++ {
		filters[i] = GammatoneFilterConfig{
			CenterFreq: freqs[i],
			Bandwidth:  erbBandwidth(freqs[i]),
			Order:      config.FilterOrder,
			SampleRate: config.SampleRate,
		}
	}

	// Initialize states
	filterState := make([][]float64, config.NumChannels)
	for i := range filterState {
		filterState[i] = make([]float64, config.FilterOrder)
	}

	return &SiliconCochlea{
		Config:          config,
		Filters:         filters,
		FilterState:     filterState,
		AdaptationState: make([]float64, config.NumChannels),
		SpikeThreshold:  make([]float64, config.NumChannels),
		OutputSpikes:    make([][]float64, config.NumChannels),
		EnergyPerSpike:  0.1, // 0.1 pJ typical
	}
}

// erbSpace generates frequencies on the ERB scale
func erbSpace(minFreq, maxFreq float64, numChannels int) []float64 {
	freqs := make([]float64, numChannels)

	// ERB scale conversion
	minERB := 21.4 * math.Log10(0.00437*minFreq+1)
	maxERB := 21.4 * math.Log10(0.00437*maxFreq+1)

	for i := 0; i < numChannels; i++ {
		erb := minERB + float64(i)*(maxERB-minERB)/float64(numChannels-1)
		freqs[i] = (math.Pow(10, erb/21.4) - 1) / 0.00437
	}

	return freqs
}

// erbBandwidth returns the ERB bandwidth at a given frequency
func erbBandwidth(freq float64) float64 {
	return 24.7 * (0.00437*freq + 1)
}

// ProcessAudio processes audio samples through the cochlea
func (cochlea *SiliconCochlea) ProcessAudio(samples []float64) [][]float64 {
	numSamples := len(samples)
	output := make([][]float64, cochlea.Config.NumChannels)
	for i := range output {
		output[i] = make([]float64, numSamples)
	}

	dt := 1.0 / cochlea.Config.SampleRate
	adaptTau := cochlea.Config.AdaptationTau / 1000.0 // Convert to seconds

	for ch := 0; ch < cochlea.Config.NumChannels; ch++ {
		filter := &cochlea.Filters[ch]

		// Simplified gammatone-like filtering
		omega := 2 * math.Pi * filter.CenterFreq
		b := 2 * math.Pi * filter.Bandwidth

		for t := 0; t < numSamples; t++ {
			// First-order approximation of gammatone
			x := samples[t]

			// IIR filter step
			decay := math.Exp(-b * dt)
			cochlea.FilterState[ch][0] = decay*cochlea.FilterState[ch][0] + (1-decay)*x*math.Cos(omega*float64(t)*dt)

			// Output with compression
			y := cochlea.FilterState[ch][0]
			if cochlea.Config.HalfWaveRectify && y < 0 {
				y = 0
			}

			// Compressive nonlinearity
			if y > 0 {
				y = math.Pow(y, cochlea.Config.Compression)
			}

			// Adaptation
			cochlea.AdaptationState[ch] = cochlea.AdaptationState[ch]*math.Exp(-dt/adaptTau) + y*(1-math.Exp(-dt/adaptTau))
			y = y / (cochlea.AdaptationState[ch] + 0.001)

			output[ch][t] = y
		}
	}

	return output
}

// ConvertToSpikes converts cochlear output to spike trains
func (cochlea *SiliconCochlea) ConvertToSpikes(filterOutput [][]float64, threshold float64) [][]float64 {
	spikes := make([][]float64, cochlea.Config.NumChannels)
	dt := 1.0 / cochlea.Config.SampleRate * 1e6 // Convert to microseconds

	for ch := 0; ch < cochlea.Config.NumChannels; ch++ {
		spikes[ch] = make([]float64, 0)
		membranePotential := 0.0
		refractoryTime := 0.0

		for t := 0; t < len(filterOutput[ch]); t++ {
			timestamp := float64(t) * dt

			if refractoryTime > 0 {
				refractoryTime -= dt
				continue
			}

			membranePotential += filterOutput[ch][t]

			if membranePotential > threshold {
				spikes[ch] = append(spikes[ch], timestamp)
				membranePotential = 0
				refractoryTime = 1000.0 // 1 ms refractory
			}

			// Leak
			membranePotential *= 0.99
		}
	}

	return spikes
}

// =============================================================================
// Speech-to-Spike Encoding
// =============================================================================

// Speech2SpikesConfig configures the speech-to-spikes pipeline
type Speech2SpikesConfig struct {
	SampleRate      float64
	FrameSize       int     // Samples per frame
	HopSize         int     // Overlap
	NumMelFilters   int     // Mel filterbank channels
	NumCepstra      int     // MFCC coefficients (if used)
	SpikeThreshold  float64
	TemporalCoding  bool    // Use spike timing vs rate coding
	LatencyMs       float64 // Processing latency
}

// DefaultSpeech2SpikesConfig returns typical parameters
func DefaultSpeech2SpikesConfig() *Speech2SpikesConfig {
	return &Speech2SpikesConfig{
		SampleRate:     16000,
		FrameSize:      400,  // 25 ms
		HopSize:        160,  // 10 ms
		NumMelFilters:  40,
		NumCepstra:     13,
		SpikeThreshold: 0.5,
		TemporalCoding: true,
		LatencyMs:      5.0,
	}
}

// Speech2Spikes encodes speech to spike trains
type Speech2Spikes struct {
	Config       *Speech2SpikesConfig
	MelFilters   [][]float64

	// Performance metrics
	SpikeRate    float64 // Spikes per second
	Compression  float64 // Data reduction ratio
}

// NewSpeech2Spikes creates a new speech encoder
func NewSpeech2Spikes(config *Speech2SpikesConfig) *Speech2Spikes {
	if config == nil {
		config = DefaultSpeech2SpikesConfig()
	}

	// Create mel filterbank
	melFilters := createMelFilterbank(config.SampleRate, config.FrameSize, config.NumMelFilters)

	return &Speech2Spikes{
		Config:     config,
		MelFilters: melFilters,
	}
}

// createMelFilterbank creates a mel-scale filterbank
func createMelFilterbank(sampleRate float64, frameSize, numFilters int) [][]float64 {
	// Mel scale conversion
	hzToMel := func(hz float64) float64 {
		return 2595 * math.Log10(1+hz/700)
	}
	melToHz := func(mel float64) float64 {
		return 700 * (math.Pow(10, mel/2595) - 1)
	}

	minMel := hzToMel(0)
	maxMel := hzToMel(sampleRate / 2)

	melPoints := make([]float64, numFilters+2)
	for i := range melPoints {
		melPoints[i] = minMel + float64(i)*(maxMel-minMel)/float64(numFilters+1)
	}

	hzPoints := make([]float64, len(melPoints))
	for i, mel := range melPoints {
		hzPoints[i] = melToHz(mel)
	}

	// Convert to FFT bins
	binPoints := make([]int, len(hzPoints))
	nfft := frameSize
	for i, hz := range hzPoints {
		binPoints[i] = int(math.Floor((float64(nfft) + 1) * hz / sampleRate))
	}

	// Create triangular filters
	filters := make([][]float64, numFilters)
	for m := 0; m < numFilters; m++ {
		filters[m] = make([]float64, nfft/2+1)
		for k := binPoints[m]; k < binPoints[m+1]; k++ {
			filters[m][k] = float64(k-binPoints[m]) / float64(binPoints[m+1]-binPoints[m])
		}
		for k := binPoints[m+1]; k < binPoints[m+2]; k++ {
			filters[m][k] = float64(binPoints[m+2]-k) / float64(binPoints[m+2]-binPoints[m+1])
		}
	}

	return filters
}

// Encode converts speech samples to spike trains
func (s2s *Speech2Spikes) Encode(samples []float64) [][]float64 {
	numFrames := (len(samples) - s2s.Config.FrameSize) / s2s.Config.HopSize + 1
	if numFrames < 1 {
		numFrames = 1
	}

	spikes := make([][]float64, s2s.Config.NumMelFilters)
	for i := range spikes {
		spikes[i] = make([]float64, 0)
	}

	frameDuration := float64(s2s.Config.HopSize) / s2s.Config.SampleRate * 1000 // ms

	for frame := 0; frame < numFrames; frame++ {
		start := frame * s2s.Config.HopSize
		end := start + s2s.Config.FrameSize
		if end > len(samples) {
			end = len(samples)
		}

		// Simple energy in each mel band (simplified FFT)
		for m := 0; m < s2s.Config.NumMelFilters; m++ {
			energy := 0.0
			for i := start; i < end && i < len(samples); i++ {
				binIdx := (i - start) * len(s2s.MelFilters[m]) / s2s.Config.FrameSize
				if binIdx < len(s2s.MelFilters[m]) {
					energy += samples[i] * samples[i] * s2s.MelFilters[m][binIdx]
				}
			}
			energy = math.Sqrt(energy / float64(end-start))

			// Convert to spikes
			if energy > s2s.Config.SpikeThreshold {
				timestamp := float64(frame)*frameDuration + s2s.Config.LatencyMs
				if s2s.Config.TemporalCoding {
					// Earlier spike for higher energy
					latency := (1.0 - energy) * frameDuration
					timestamp += latency
				}
				spikes[m] = append(spikes[m], timestamp)
			}
		}
	}

	// Calculate metrics
	totalSpikes := 0
	for _, ch := range spikes {
		totalSpikes += len(ch)
	}
	duration := float64(len(samples)) / s2s.Config.SampleRate
	s2s.SpikeRate = float64(totalSpikes) / duration
	s2s.Compression = float64(len(samples)) / float64(totalSpikes+1)

	return spikes
}

// =============================================================================
// PART 2: CIM COMPILER AND MAPPING
// =============================================================================

// CrossbarSize represents crossbar array dimensions
type CrossbarSize struct {
	Rows    int
	Cols    int
	NumADCs int
	ADCBits int
}

// LayerType represents the type of neural network layer
type LayerType int

const (
	LayerTypeFC LayerType = iota
	LayerTypeConv
	LayerTypeDepthwiseConv
	LayerTypeBatchNorm
	LayerTypePooling
	LayerTypeAttention
	LayerTypeResidual
)

// LayerDescriptor describes a layer to be mapped
type LayerDescriptor struct {
	Name           string
	Type           LayerType
	InputShape     []int     // [batch, channels, height, width] or [batch, features]
	OutputShape    []int
	WeightShape    []int     // [out_features, in_features] or [out_ch, in_ch, kH, kW]
	Weights        []float64 // Flattened weights
	Bias           []float64
	Stride         int
	Padding        int
	KernelSize     int
	NumHeads       int       // For attention
	HasResidual    bool
}

// LayerPartition represents a partition of a layer for crossbar mapping
type LayerPartition struct {
	LayerName      string
	PartitionID    int
	RowStart       int
	RowEnd         int
	ColStart       int
	ColEnd         int
	CrossbarID     int
	WeightSlice    []float64
	ReplicationID  int       // Which replica this is
}

// WeightReplicationStrategy defines how weights are replicated
type WeightReplicationStrategy int

const (
	ReplicationBalance  WeightReplicationStrategy = iota // Balance computation
	ReplicationW0H0                                      // Layer size balance
	ReplicationUniform                                   // Uniform replication
	ReplicationGA                                        // Genetic algorithm optimized
)

// DataflowMode defines the dataflow scheduling mode
type DataflowMode int

const (
	DataflowHighThroughput DataflowMode = iota // Pipeline across layers
	DataflowLowLatency                         // Minimize single-sample latency
	DataflowHybrid                             // Adaptive
)

// =============================================================================
// PIMCOMP-style Compiler
// =============================================================================

// PIMCompilerConfig configures the CIM compiler
type PIMCompilerConfig struct {
	CrossbarSize          CrossbarSize
	NumCrossbars          int
	NumTiles              int
	WeightBits            int
	ActivationBits        int
	ReplicationStrategy   WeightReplicationStrategy
	DataflowMode          DataflowMode
	EnableWeightSharing   bool
	EnableInputBroadcast  bool
	MaxReplicationFactor  int
	OptimizeEDP           bool    // Energy-delay product
}

// DefaultPIMCompilerConfig returns default compiler settings
func DefaultPIMCompilerConfig() *PIMCompilerConfig {
	return &PIMCompilerConfig{
		CrossbarSize: CrossbarSize{
			Rows:    128,
			Cols:    128,
			NumADCs: 8,
			ADCBits: 8,
		},
		NumCrossbars:         16,
		NumTiles:             4,
		WeightBits:           8,
		ActivationBits:       8,
		ReplicationStrategy:  ReplicationBalance,
		DataflowMode:         DataflowHighThroughput,
		EnableWeightSharing:  true,
		EnableInputBroadcast: true,
		MaxReplicationFactor: 4,
		OptimizeEDP:          true,
	}
}

// PIMCompiler implements PIMCOMP-style compilation
type PIMCompiler struct {
	Config           *PIMCompilerConfig
	Layers           []LayerDescriptor
	Partitions       []LayerPartition
	Schedule         []ScheduleEntry
	CoreMapping      map[int][]LayerPartition

	// Metrics
	Utilization      float64
	Throughput       float64  // Inferences per second
	Latency          float64  // ms per inference
	EnergyPerInf     float64  // mJ per inference
	MemoryFootprint  int64    // bytes
}

// ScheduleEntry represents a scheduled operation
type ScheduleEntry struct {
	Cycle          int
	CrossbarID     int
	PartitionID    int
	Operation      string   // "MVM", "Activation", "Pooling", "Transfer"
	Dependencies   []int    // Previous entries this depends on
	InputAddress   int
	OutputAddress  int
}

// NewPIMCompiler creates a new compiler
func NewPIMCompiler(config *PIMCompilerConfig) *PIMCompiler {
	if config == nil {
		config = DefaultPIMCompilerConfig()
	}

	return &PIMCompiler{
		Config:      config,
		CoreMapping: make(map[int][]LayerPartition),
	}
}

// AddLayer adds a layer to be compiled
func (compiler *PIMCompiler) AddLayer(layer LayerDescriptor) {
	compiler.Layers = append(compiler.Layers, layer)
}

// Compile performs full compilation
func (compiler *PIMCompiler) Compile() error {
	// Stage 1: Node partitioning
	if err := compiler.partitionLayers(); err != nil {
		return fmt.Errorf("partitioning failed: %v", err)
	}

	// Stage 2: Weight replication
	if err := compiler.replicateWeights(); err != nil {
		return fmt.Errorf("weight replication failed: %v", err)
	}

	// Stage 3: Core mapping
	if err := compiler.mapToCores(); err != nil {
		return fmt.Errorf("core mapping failed: %v", err)
	}

	// Stage 4: Dataflow scheduling
	if err := compiler.scheduleDataflow(); err != nil {
		return fmt.Errorf("scheduling failed: %v", err)
	}

	// Calculate metrics
	compiler.calculateMetrics()

	return nil
}

// partitionLayers divides layers according to crossbar size
func (compiler *PIMCompiler) partitionLayers() error {
	compiler.Partitions = make([]LayerPartition, 0)
	partitionID := 0

	xbarRows := compiler.Config.CrossbarSize.Rows
	xbarCols := compiler.Config.CrossbarSize.Cols

	for _, layer := range compiler.Layers {
		switch layer.Type {
		case LayerTypeFC:
			// Partition fully connected layer
			inFeatures := layer.WeightShape[1]
			outFeatures := layer.WeightShape[0]

			// Partition along input dimension (rows)
			for rowStart := 0; rowStart < inFeatures; rowStart += xbarRows {
				rowEnd := rowStart + xbarRows
				if rowEnd > inFeatures {
					rowEnd = inFeatures
				}

				// Partition along output dimension (columns)
				for colStart := 0; colStart < outFeatures; colStart += xbarCols {
					colEnd := colStart + xbarCols
					if colEnd > outFeatures {
						colEnd = outFeatures
					}

					partition := LayerPartition{
						LayerName:   layer.Name,
						PartitionID: partitionID,
						RowStart:    rowStart,
						RowEnd:      rowEnd,
						ColStart:    colStart,
						ColEnd:      colEnd,
					}

					// Extract weight slice
					partition.WeightSlice = compiler.extractWeightSlice(
						layer.Weights, layer.WeightShape,
						rowStart, rowEnd, colStart, colEnd,
					)

					compiler.Partitions = append(compiler.Partitions, partition)
					partitionID++
				}
			}

		case LayerTypeConv:
			// Use im2col mapping for convolution
			// Input: [batch, in_ch, H, W], Weight: [out_ch, in_ch, kH, kW]
			inChannels := layer.WeightShape[1]
			outChannels := layer.WeightShape[0]
			kH := layer.WeightShape[2]
			kW := layer.WeightShape[3]

			// Unrolled input dimension
			unrolledIn := inChannels * kH * kW

			// Partition similar to FC
			for rowStart := 0; rowStart < unrolledIn; rowStart += xbarRows {
				rowEnd := rowStart + xbarRows
				if rowEnd > unrolledIn {
					rowEnd = unrolledIn
				}

				for colStart := 0; colStart < outChannels; colStart += xbarCols {
					colEnd := colStart + xbarCols
					if colEnd > outChannels {
						colEnd = outChannels
					}

					partition := LayerPartition{
						LayerName:   layer.Name,
						PartitionID: partitionID,
						RowStart:    rowStart,
						RowEnd:      rowEnd,
						ColStart:    colStart,
						ColEnd:      colEnd,
					}

					compiler.Partitions = append(compiler.Partitions, partition)
					partitionID++
				}
			}

		case LayerTypeDepthwiseConv:
			// Each channel processed independently
			numChannels := layer.WeightShape[0]
			kH := layer.WeightShape[2]
			kW := layer.WeightShape[3]
			kernelSize := kH * kW

			channelsPerXbar := xbarCols

			for chStart := 0; chStart < numChannels; chStart += channelsPerXbar {
				chEnd := chStart + channelsPerXbar
				if chEnd > numChannels {
					chEnd = numChannels
				}

				partition := LayerPartition{
					LayerName:   layer.Name,
					PartitionID: partitionID,
					RowStart:    0,
					RowEnd:      kernelSize,
					ColStart:    chStart,
					ColEnd:      chEnd,
				}

				compiler.Partitions = append(compiler.Partitions, partition)
				partitionID++
			}
		}
	}

	return nil
}

// extractWeightSlice extracts a portion of weights
func (compiler *PIMCompiler) extractWeightSlice(weights []float64, shape []int, rowStart, rowEnd, colStart, colEnd int) []float64 {
	if len(shape) < 2 {
		return nil
	}

	outDim := shape[0]
	inDim := shape[1]

	sliceRows := rowEnd - rowStart
	sliceCols := colEnd - colStart
	slice := make([]float64, sliceRows*sliceCols)

	for row := rowStart; row < rowEnd && row < inDim; row++ {
		for col := colStart; col < colEnd && col < outDim; col++ {
			srcIdx := col*inDim + row
			dstIdx := (row-rowStart)*sliceCols + (col - colStart)
			if srcIdx < len(weights) && dstIdx < len(slice) {
				slice[dstIdx] = weights[srcIdx]
			}
		}
	}

	return slice
}

// replicateWeights determines replication for each partition
func (compiler *PIMCompiler) replicateWeights() error {
	numCrossbars := compiler.Config.NumCrossbars
	numPartitions := len(compiler.Partitions)

	if numPartitions == 0 {
		return nil
	}

	switch compiler.Config.ReplicationStrategy {
	case ReplicationBalance:
		// Balance computation across crossbars
		// Count partitions per layer
		layerPartitions := make(map[string]int)
		for _, p := range compiler.Partitions {
			layerPartitions[p.LayerName]++
		}

		// Assign crossbars proportionally
		crossbarIdx := 0
		for i := range compiler.Partitions {
			compiler.Partitions[i].CrossbarID = crossbarIdx % numCrossbars
			crossbarIdx++
		}

	case ReplicationUniform:
		// Uniform replication factor
		replicationFactor := compiler.Config.MaxReplicationFactor
		newPartitions := make([]LayerPartition, 0)

		for _, p := range compiler.Partitions {
			for rep := 0; rep < replicationFactor; rep++ {
				newP := p
				newP.ReplicationID = rep
				newP.CrossbarID = (p.PartitionID*replicationFactor + rep) % numCrossbars
				newPartitions = append(newPartitions, newP)
			}
		}
		compiler.Partitions = newPartitions

	case ReplicationGA:
		// Genetic algorithm optimization (simplified)
		// Would use actual GA in production
		for i := range compiler.Partitions {
			// Heuristic: balance based on partition size
			size := (compiler.Partitions[i].RowEnd - compiler.Partitions[i].RowStart) *
				(compiler.Partitions[i].ColEnd - compiler.Partitions[i].ColStart)
			compiler.Partitions[i].CrossbarID = (i * size / 1000) % numCrossbars
		}
	}

	return nil
}

// mapToCores assigns partitions to cores
func (compiler *PIMCompiler) mapToCores() error {
	// Group partitions by crossbar
	for _, p := range compiler.Partitions {
		coreID := p.CrossbarID / (compiler.Config.NumCrossbars / compiler.Config.NumTiles)
		compiler.CoreMapping[coreID] = append(compiler.CoreMapping[coreID], p)
	}

	return nil
}

// scheduleDataflow creates the execution schedule
func (compiler *PIMCompiler) scheduleDataflow() error {
	compiler.Schedule = make([]ScheduleEntry, 0)

	switch compiler.Config.DataflowMode {
	case DataflowHighThroughput:
		// Pipeline execution across layers
		cycle := 0

		// Group partitions by layer
		layerOrder := make([]string, 0)
		layerPartitions := make(map[string][]LayerPartition)

		for _, p := range compiler.Partitions {
			if _, exists := layerPartitions[p.LayerName]; !exists {
				layerOrder = append(layerOrder, p.LayerName)
			}
			layerPartitions[p.LayerName] = append(layerPartitions[p.LayerName], p)
		}

		// Schedule layer by layer
		for _, layerName := range layerOrder {
			partitions := layerPartitions[layerName]

			// Schedule all partitions of this layer in parallel where possible
			for _, p := range partitions {
				entry := ScheduleEntry{
					Cycle:       cycle,
					CrossbarID:  p.CrossbarID,
					PartitionID: p.PartitionID,
					Operation:   "MVM",
				}
				compiler.Schedule = append(compiler.Schedule, entry)
			}

			// Accumulate results
			cycle++
			compiler.Schedule = append(compiler.Schedule, ScheduleEntry{
				Cycle:     cycle,
				Operation: "Accumulate",
			})

			// Activation
			cycle++
			compiler.Schedule = append(compiler.Schedule, ScheduleEntry{
				Cycle:     cycle,
				Operation: "Activation",
			})

			cycle++
		}

	case DataflowLowLatency:
		// Minimize latency for single sample
		// Execute partitions in dependency order
		cycle := 0

		for _, p := range compiler.Partitions {
			entry := ScheduleEntry{
				Cycle:       cycle,
				CrossbarID:  p.CrossbarID,
				PartitionID: p.PartitionID,
				Operation:   "MVM",
			}
			compiler.Schedule = append(compiler.Schedule, entry)
			cycle++
		}
	}

	return nil
}

// calculateMetrics computes performance metrics
func (compiler *PIMCompiler) calculateMetrics() {
	numPartitions := len(compiler.Partitions)
	numCrossbars := compiler.Config.NumCrossbars
	xbarSize := compiler.Config.CrossbarSize.Rows * compiler.Config.CrossbarSize.Cols

	// Utilization
	totalCells := numCrossbars * xbarSize
	usedCells := 0
	for _, p := range compiler.Partitions {
		usedCells += (p.RowEnd - p.RowStart) * (p.ColEnd - p.ColStart)
	}
	compiler.Utilization = float64(usedCells) / float64(totalCells) * 100

	// Throughput (simplified model)
	cyclesPerInference := len(compiler.Schedule)
	clockFreqMHz := 100.0 // 100 MHz typical
	compiler.Latency = float64(cyclesPerInference) / clockFreqMHz // ms
	compiler.Throughput = 1000.0 / compiler.Latency              // Inferences/s

	// Energy (simplified)
	energyPerMVM := 0.1 // pJ per MAC
	macsPerPartition := float64(xbarSize)
	totalMACs := float64(numPartitions) * macsPerPartition
	compiler.EnergyPerInf = totalMACs * energyPerMVM / 1e9 // mJ

	// Memory
	compiler.MemoryFootprint = int64(usedCells) * int64(compiler.Config.WeightBits) / 8
}

// GetSchedule returns the compiled schedule
func (compiler *PIMCompiler) GetSchedule() []ScheduleEntry {
	return compiler.Schedule
}

// GetPartitions returns the layer partitions
func (compiler *PIMCompiler) GetPartitions() []LayerPartition {
	return compiler.Partitions
}

// =============================================================================
// CIM-MLC Multi-Level Abstraction
// =============================================================================

// AbstractionLevel defines the hardware abstraction level
type AbstractionLevel int

const (
	AbstractionChip AbstractionLevel = iota
	AbstractionTile
	AbstractionCore
	AbstractionCrossbar
	AbstractionWordline
)

// CIMMLCConfig configures the multi-level compiler
type CIMMLCConfig struct {
	ChipConfig        ChipAbstraction
	TileConfig        TileAbstraction
	CoreConfig        CoreAbstraction
	CrossbarConfig    CrossbarAbstraction
	ComputeMode       ComputeMode
}

// ChipAbstraction describes chip-level architecture
type ChipAbstraction struct {
	NumTiles            int
	InterTileInterconnect string  // "mesh", "bus", "htree"
	GlobalBufferKB      int
	ControllerType      string
}

// TileAbstraction describes tile-level architecture
type TileAbstraction struct {
	NumCores          int
	LocalBufferKB     int
	InterCoreLatency  int  // cycles
}

// CoreAbstraction describes core-level architecture
type CoreAbstraction struct {
	NumCrossbars      int
	AccumulatorBits   int
	ActivationUnit    bool
	PoolingUnit       bool
}

// CrossbarAbstraction describes crossbar-level architecture
type CrossbarAbstraction struct {
	Rows              int
	Cols              int
	CellType          string  // "ReRAM", "FeRAM", "SRAM", "Flash"
	BitsPrecision     int
	ADCsPerCol        int
	DACsPerRow        int
}

// ComputeMode defines the computation granularity
type ComputeMode int

const (
	ComputeModeVector   ComputeMode = iota // Full MVM
	ComputeModeRow                         // Row-by-row
	ComputeModeBlock                       // Block-wise
)

// DefaultCIMMLCConfig returns a default multi-level configuration
func DefaultCIMMLCConfig() *CIMMLCConfig {
	return &CIMMLCConfig{
		ChipConfig: ChipAbstraction{
			NumTiles:              4,
			InterTileInterconnect: "mesh",
			GlobalBufferKB:        256,
			ControllerType:        "centralized",
		},
		TileConfig: TileAbstraction{
			NumCores:         4,
			LocalBufferKB:    64,
			InterCoreLatency: 2,
		},
		CoreConfig: CoreAbstraction{
			NumCrossbars:   4,
			AccumulatorBits: 32,
			ActivationUnit:  true,
			PoolingUnit:     true,
		},
		CrossbarConfig: CrossbarAbstraction{
			Rows:          128,
			Cols:          128,
			CellType:      "FeRAM",
			BitsPrecision: 6,
			ADCsPerCol:    1,
			DACsPerRow:    1,
		},
		ComputeMode: ComputeModeVector,
	}
}

// CIMMLCCompiler implements multi-level compilation
type CIMMLCCompiler struct {
	Config           *CIMMLCConfig

	// Hierarchical mappings
	ChipMapping      []TileMapping
	TileMappings     map[int][]CoreMapping
	CoreMappings     map[int][]CrossbarMapping

	// Performance model
	Latency          float64
	Energy           float64
	Throughput       float64
}

// TileMapping represents mapping at tile level
type TileMapping struct {
	TileID       int
	LayerNames   []string
	Utilization  float64
}

// CoreMapping represents mapping at core level
type CoreMapping struct {
	CoreID       int
	Partitions   []LayerPartition
	BufferUsage  int
}

// CrossbarMapping represents mapping at crossbar level
type CrossbarMapping struct {
	CrossbarID   int
	WeightMatrix [][]float64
	Utilization  float64
}

// NewCIMMLCCompiler creates a new multi-level compiler
func NewCIMMLCCompiler(config *CIMMLCConfig) *CIMMLCCompiler {
	if config == nil {
		config = DefaultCIMMLCConfig()
	}

	return &CIMMLCCompiler{
		Config:       config,
		TileMappings: make(map[int][]CoreMapping),
		CoreMappings: make(map[int][]CrossbarMapping),
	}
}

// =============================================================================
// COMPASS Resource-Constrained Compiler
// =============================================================================

// COMPASSConfig configures the COMPASS compiler
type COMPASSConfig struct {
	OnChipCapacity    int64   // bytes
	ExternalMemBW     float64 // GB/s
	TargetThroughput  float64 // inferences/s
	TargetEDP         float64 // Energy-delay product target
	GAPopulation      int     // Genetic algorithm population
	GAGenerations     int     // Genetic algorithm generations
}

// DefaultCOMPASSConfig returns default COMPASS settings
func DefaultCOMPASSConfig() *COMPASSConfig {
	return &COMPASSConfig{
		OnChipCapacity:   1024 * 1024, // 1 MB
		ExternalMemBW:    10.0,        // 10 GB/s
		TargetThroughput: 100.0,
		TargetEDP:        1.0,
		GAPopulation:     50,
		GAGenerations:    100,
	}
}

// NetworkPartition represents a partition of the network
type NetworkPartition struct {
	PartitionID    int
	Layers         []string
	MemoryRequired int64
	ComputeOps     int64
	Dependencies   []int
}

// COMPASSCompiler implements COMPASS compilation for resource-constrained scenarios
type COMPASSCompiler struct {
	Config         *COMPASSConfig
	Network        []LayerDescriptor
	Partitions     []NetworkPartition

	// Optimization results
	OptimalPartitioning []NetworkPartition
	MemorySchedule      []MemoryTransfer
	ThroughputAchieved  float64
	EDPAchieved         float64
}

// MemoryTransfer represents a memory transfer operation
type MemoryTransfer struct {
	PartitionID   int
	Direction     string  // "load" or "store"
	SizeBytes     int64
	StartTime     float64
	EndTime       float64
}

// NewCOMPASSCompiler creates a new COMPASS compiler
func NewCOMPASSCompiler(config *COMPASSConfig) *COMPASSCompiler {
	if config == nil {
		config = DefaultCOMPASSConfig()
	}

	return &COMPASSCompiler{
		Config: config,
	}
}

// SetNetwork sets the network to compile
func (compiler *COMPASSCompiler) SetNetwork(layers []LayerDescriptor) {
	compiler.Network = layers
}

// OptimizePartitioning finds optimal network partitioning using GA
func (compiler *COMPASSCompiler) OptimizePartitioning() error {
	numLayers := len(compiler.Network)
	if numLayers == 0 {
		return fmt.Errorf("no layers to partition")
	}

	// Calculate memory requirements per layer
	layerMemory := make([]int64, numLayers)
	for i, layer := range compiler.Network {
		layerMemory[i] = int64(len(layer.Weights)) * 4 // 4 bytes per float
	}

	// Simple greedy partitioning (simplified from full GA)
	partitions := make([]NetworkPartition, 0)
	currentPartition := NetworkPartition{
		PartitionID: 0,
		Layers:      make([]string, 0),
	}

	for i, layer := range compiler.Network {
		if currentPartition.MemoryRequired+layerMemory[i] > compiler.Config.OnChipCapacity {
			// Start new partition
			partitions = append(partitions, currentPartition)
			currentPartition = NetworkPartition{
				PartitionID:  len(partitions),
				Layers:       make([]string, 0),
				Dependencies: []int{len(partitions) - 1},
			}
		}

		currentPartition.Layers = append(currentPartition.Layers, layer.Name)
		currentPartition.MemoryRequired += layerMemory[i]
	}

	// Add final partition
	if len(currentPartition.Layers) > 0 {
		partitions = append(partitions, currentPartition)
	}

	compiler.OptimalPartitioning = partitions

	// Schedule memory transfers
	compiler.scheduleMemoryTransfers()

	return nil
}

// scheduleMemoryTransfers creates the memory schedule
func (compiler *COMPASSCompiler) scheduleMemoryTransfers() {
	compiler.MemorySchedule = make([]MemoryTransfer, 0)

	currentTime := 0.0

	for _, partition := range compiler.OptimalPartitioning {
		// Load partition
		loadTime := float64(partition.MemoryRequired) / (compiler.Config.ExternalMemBW * 1e9)

		transfer := MemoryTransfer{
			PartitionID: partition.PartitionID,
			Direction:   "load",
			SizeBytes:   partition.MemoryRequired,
			StartTime:   currentTime,
			EndTime:     currentTime + loadTime,
		}
		compiler.MemorySchedule = append(compiler.MemorySchedule, transfer)
		currentTime += loadTime

		// Compute time (simplified)
		computeTime := float64(partition.ComputeOps) / 1e12 // Assume 1 TOPS
		currentTime += computeTime
	}

	compiler.ThroughputAchieved = 1.0 / currentTime
}

// =============================================================================
// Tiling Optimizer
// =============================================================================

// TilingConfig configures tiling optimization
type TilingConfig struct {
	MaxTileRows       int
	MaxTileCols       int
	MinTileRows       int
	MinTileCols       int
	PreferSquare      bool
	OptimizeForEnergy bool
	OptimizeForArea   bool
}

// TilingResult represents the result of tiling optimization
type TilingResult struct {
	TileRows         int
	TileCols         int
	NumTiles         int
	Utilization      float64
	EstimatedArea    float64  // mm²
	EstimatedEnergy  float64  // pJ per op
}

// TilingOptimizer optimizes crossbar tiling
type TilingOptimizer struct {
	Config          *TilingConfig
	OptimalTiling   TilingResult
}

// NewTilingOptimizer creates a new tiling optimizer
func NewTilingOptimizer(config *TilingConfig) *TilingOptimizer {
	if config == nil {
		config = &TilingConfig{
			MaxTileRows:       256,
			MaxTileCols:       256,
			MinTileRows:       16,
			MinTileCols:       16,
			PreferSquare:      false, // Research shows square not always optimal
			OptimizeForEnergy: true,
			OptimizeForArea:   false,
		}
	}

	return &TilingOptimizer{
		Config: config,
	}
}

// OptimizeTiling finds optimal tile dimensions for a weight matrix
func (opt *TilingOptimizer) OptimizeTiling(rows, cols int) TilingResult {
	bestResult := TilingResult{
		Utilization: 0,
	}
	bestScore := math.MaxFloat64

	// Search tile dimensions
	for tileRows := opt.Config.MinTileRows; tileRows <= opt.Config.MaxTileRows; tileRows *= 2 {
		for tileCols := opt.Config.MinTileCols; tileCols <= opt.Config.MaxTileCols; tileCols *= 2 {
			// Calculate number of tiles needed
			numRowTiles := (rows + tileRows - 1) / tileRows
			numColTiles := (cols + tileCols - 1) / tileCols
			numTiles := numRowTiles * numColTiles

			// Calculate utilization
			totalCells := numTiles * tileRows * tileCols
			usedCells := rows * cols
			utilization := float64(usedCells) / float64(totalCells)

			// Estimate area (larger tiles have better area efficiency but more waste)
			areaPerCell := 0.001 // mm² per cell
			peripheralOverhead := 0.1 * float64(numTiles) // Per-tile overhead
			estimatedArea := float64(totalCells)*areaPerCell + peripheralOverhead

			// Estimate energy (larger tiles have higher IR drop)
			irDropFactor := 1.0 + 0.001*float64(tileRows+tileCols)
			baseEnergy := 0.1 // pJ base
			estimatedEnergy := baseEnergy * irDropFactor

			// Score (lower is better)
			score := 0.0
			if opt.Config.OptimizeForEnergy {
				score += estimatedEnergy * 1000
			}
			if opt.Config.OptimizeForArea {
				score += estimatedArea * 100
			}
			score += (1.0 - utilization) * 50 // Penalize low utilization

			if score < bestScore {
				bestScore = score
				bestResult = TilingResult{
					TileRows:        tileRows,
					TileCols:        tileCols,
					NumTiles:        numTiles,
					Utilization:     utilization * 100,
					EstimatedArea:   estimatedArea,
					EstimatedEnergy: estimatedEnergy,
				}
			}
		}
	}

	opt.OptimalTiling = bestResult
	return bestResult
}

// =============================================================================
// Heterogeneous Crossbar Mapping
// =============================================================================

// HeterogeneousCrossbarConfig configures heterogeneous crossbar arrays
type HeterogeneousCrossbarConfig struct {
	AvailableSizes     []CrossbarSize
	MaxTotalArea       float64  // mm²
	AreaPerCell        float64  // mm² per cell
	PeripheralOverhead float64  // mm² per crossbar
}

// HeterogeneousMapper maps networks to heterogeneous crossbar arrays
type HeterogeneousMapper struct {
	Config           *HeterogeneousCrossbarConfig
	Allocations      []CrossbarAllocation

	// Results
	TotalArea        float64
	AreaReduction    float64  // vs homogeneous
}

// CrossbarAllocation represents allocation to a specific crossbar
type CrossbarAllocation struct {
	CrossbarSize     CrossbarSize
	LayerName        string
	RowRange         [2]int
	ColRange         [2]int
	Utilization      float64
}

// NewHeterogeneousMapper creates a new heterogeneous mapper
func NewHeterogeneousMapper(config *HeterogeneousCrossbarConfig) *HeterogeneousMapper {
	if config == nil {
		config = &HeterogeneousCrossbarConfig{
			AvailableSizes: []CrossbarSize{
				{Rows: 32, Cols: 32},
				{Rows: 64, Cols: 64},
				{Rows: 128, Cols: 128},
				{Rows: 256, Cols: 256},
			},
			MaxTotalArea:       10.0,
			AreaPerCell:        0.0001, // 0.1 um² per cell
			PeripheralOverhead: 0.01,
		}
	}

	return &HeterogeneousMapper{
		Config: config,
	}
}

// MapNetwork maps a network to heterogeneous crossbars
func (mapper *HeterogeneousMapper) MapNetwork(layers []LayerDescriptor) error {
	mapper.Allocations = make([]CrossbarAllocation, 0)

	// Sort available sizes by area (smallest first)
	sizes := make([]CrossbarSize, len(mapper.Config.AvailableSizes))
	copy(sizes, mapper.Config.AvailableSizes)
	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i].Rows*sizes[i].Cols < sizes[j].Rows*sizes[j].Cols
	})

	homogeneousArea := 0.0
	heterogeneousArea := 0.0

	for _, layer := range layers {
		if len(layer.WeightShape) < 2 {
			continue
		}

		rows := layer.WeightShape[1]
		cols := layer.WeightShape[0]

		// Find best fitting crossbar size
		bestSize := sizes[len(sizes)-1] // Default to largest
		bestWaste := rows * cols        // Initialize with full waste

		for _, size := range sizes {
			if size.Rows >= rows && size.Cols >= cols {
				waste := size.Rows*size.Cols - rows*cols
				if waste < bestWaste {
					bestWaste = waste
					bestSize = size
				}
			}
		}

		// If no single crossbar fits, use largest
		numRowSplits := (rows + bestSize.Rows - 1) / bestSize.Rows
		numColSplits := (cols + bestSize.Cols - 1) / bestSize.Cols

		for rs := 0; rs < numRowSplits; rs++ {
			for cs := 0; cs < numColSplits; cs++ {
				rowStart := rs * bestSize.Rows
				rowEnd := (rs + 1) * bestSize.Rows
				if rowEnd > rows {
					rowEnd = rows
				}

				colStart := cs * bestSize.Cols
				colEnd := (cs + 1) * bestSize.Cols
				if colEnd > cols {
					colEnd = cols
				}

				usedCells := (rowEnd - rowStart) * (colEnd - colStart)
				totalCells := bestSize.Rows * bestSize.Cols

				allocation := CrossbarAllocation{
					CrossbarSize: bestSize,
					LayerName:    layer.Name,
					RowRange:     [2]int{rowStart, rowEnd},
					ColRange:     [2]int{colStart, colEnd},
					Utilization:  float64(usedCells) / float64(totalCells) * 100,
				}

				mapper.Allocations = append(mapper.Allocations, allocation)

				heterogeneousArea += float64(totalCells)*mapper.Config.AreaPerCell +
					mapper.Config.PeripheralOverhead
			}
		}

		// Calculate homogeneous area (using largest size)
		largestSize := sizes[len(sizes)-1]
		homoRowSplits := (rows + largestSize.Rows - 1) / largestSize.Rows
		homoColSplits := (cols + largestSize.Cols - 1) / largestSize.Cols
		homogeneousArea += float64(homoRowSplits*homoColSplits*largestSize.Rows*largestSize.Cols)*
			mapper.Config.AreaPerCell +
			float64(homoRowSplits*homoColSplits)*mapper.Config.PeripheralOverhead
	}

	mapper.TotalArea = heterogeneousArea
	if homogeneousArea > 0 {
		mapper.AreaReduction = (1.0 - heterogeneousArea/homogeneousArea) * 100
	}

	return nil
}

// =============================================================================
// Integrated Sensory-CIM Pipeline
// =============================================================================

// SensoryCIMPipelineConfig configures an end-to-end sensory-to-CIM pipeline
type SensoryCIMPipelineConfig struct {
	SensorType        string  // "DVS", "Cochlea", "Speech"
	ProcessingMode    string  // "spike", "rate", "frame"
	CIMCompilerConfig *PIMCompilerConfig
	EnableFusion      bool    // Fuse sensor preprocessing with CIM
	TargetLatencyMs   float64
	TargetPowerMW     float64
}

// SensoryCIMPipeline represents an integrated sensory-CIM system
type SensoryCIMPipeline struct {
	Config         *SensoryCIMPipelineConfig

	// Sensor components
	DVSSensor      *DVSSensor
	Cochlea        *SiliconCochlea
	Speech2Spikes  *Speech2Spikes

	// CIM compiler
	Compiler       *PIMCompiler

	// Performance
	EndToEndLatency float64
	TotalPower      float64
	DataReduction   float64
}

// NewSensoryCIMPipeline creates a new integrated pipeline
func NewSensoryCIMPipeline(config *SensoryCIMPipelineConfig) *SensoryCIMPipeline {
	if config == nil {
		config = &SensoryCIMPipelineConfig{
			SensorType:        "DVS",
			ProcessingMode:    "spike",
			CIMCompilerConfig: DefaultPIMCompilerConfig(),
			EnableFusion:      true,
			TargetLatencyMs:   10.0,
			TargetPowerMW:     10.0,
		}
	}

	pipeline := &SensoryCIMPipeline{
		Config: config,
	}

	// Initialize sensor
	switch config.SensorType {
	case "DVS":
		pipeline.DVSSensor = NewDVSSensor(DVSSensorSpecs("GenX320"))
	case "Cochlea":
		pipeline.Cochlea = NewSiliconCochlea(DefaultCochleaConfig())
	case "Speech":
		pipeline.Speech2Spikes = NewSpeech2Spikes(DefaultSpeech2SpikesConfig())
	}

	// Initialize compiler
	pipeline.Compiler = NewPIMCompiler(config.CIMCompilerConfig)

	return pipeline
}

// EstimatePerformance estimates end-to-end performance
func (pipeline *SensoryCIMPipeline) EstimatePerformance() {
	sensorLatency := 0.0
	sensorPower := 0.0

	switch pipeline.Config.SensorType {
	case "DVS":
		if pipeline.DVSSensor != nil {
			sensorLatency = 0.15  // 150 us typical
			sensorPower = pipeline.DVSSensor.Config.PowerConsumption
			pipeline.DataReduction = 100.0 // 100x typical
		}
	case "Cochlea":
		if pipeline.Cochlea != nil {
			sensorLatency = 0.1
			sensorPower = pipeline.Cochlea.Config.PowerConsumption
			pipeline.DataReduction = 50.0
		}
	case "Speech":
		if pipeline.Speech2Spikes != nil {
			sensorLatency = pipeline.Speech2Spikes.Config.LatencyMs
			sensorPower = 0.5 // 0.5 mW typical
			pipeline.DataReduction = pipeline.Speech2Spikes.Compression
		}
	}

	// CIM latency and power
	cimLatency := pipeline.Compiler.Latency
	cimPower := pipeline.Compiler.EnergyPerInf / cimLatency // mW

	pipeline.EndToEndLatency = sensorLatency + cimLatency
	pipeline.TotalPower = sensorPower + cimPower
}

// GetPipelineStats returns pipeline statistics
func (pipeline *SensoryCIMPipeline) GetPipelineStats() map[string]float64 {
	return map[string]float64{
		"end_to_end_latency_ms": pipeline.EndToEndLatency,
		"total_power_mw":        pipeline.TotalPower,
		"data_reduction_x":      pipeline.DataReduction,
		"cim_utilization":       pipeline.Compiler.Utilization,
		"cim_throughput":        pipeline.Compiler.Throughput,
	}
}
