// edge_neuromorphic_cim.go - Edge AI and Neuromorphic Sensor Fusion with FeFET CIM
// Part of IronLattice educational demonstrations
// Research iteration 140: TinyML, keyword spotting, event cameras, in-sensor computing

package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// PART 1: EDGE AI WITH FERROELECTRIC CIM
// =============================================================================

// TinyMLConfig configures ultra-low power inference
type TinyMLConfig struct {
	// Model parameters
	ModelType      string  // "kws", "vww", "ad", "ic" (MLPerf Tiny benchmarks)
	InputSize      int     // Input dimension
	HiddenLayers   []int   // Hidden layer sizes
	OutputClasses  int     // Number of output classes
	Quantization   int     // Weight bits (4, 8)

	// Hardware parameters
	ClockFreqKHz   float64 // Operating frequency (kHz)
	SupplyVoltage  float64 // Supply voltage (V)
	Technology     string  // "28nm", "22nm", "12nm"

	// Power parameters
	ActivePowerUW  float64 // Active power (µW)
	LeakagePowerNW float64 // Leakage power (nW)
	EnergyPerInfNJ float64 // Energy per inference (nJ)

	// Memory
	SRAMSizeKB     int     // On-chip SRAM (KB)
	FlashSizeKB    int     // Flash storage (KB)
}

// DefaultKWSConfig returns keyword spotting configuration
func DefaultKWSConfig() *TinyMLConfig {
	return &TinyMLConfig{
		ModelType:      "kws",
		InputSize:      490,    // 49 MFCC frames × 10 coefficients
		HiddenLayers:   []int{256, 256, 256},
		OutputClasses:  12,     // 10 keywords + silence + unknown
		Quantization:   8,
		ClockFreqKHz:   250.0,  // 250 kHz
		SupplyVoltage:  0.5,    // 0.5V subthreshold
		Technology:     "28nm",
		ActivePowerUW:  3.8,    // 3.8 µW (from paper)
		LeakagePowerNW: 100.0,
		EnergyPerInfNJ: 183.4,  // 183.4 nJ/inference
		SRAMSizeKB:     64,
		FlashSizeKB:    512,
	}
}

// FeFETCIMConfig configures FeFET-based CIM for edge
type FeFETCIMConfig struct {
	// Array parameters
	ArrayRows      int     // Crossbar rows
	ArrayCols      int     // Crossbar columns
	CellType       string  // "1FeFET", "2T1FeFET", "1FeFET1R"
	WeightBits     int     // Multi-level cell bits

	// FeFET parameters
	Vth0           float64 // Initial threshold voltage (V)
	MemoryWindow   float64 // Memory window (V)
	OnOffRatio     float64 // ION/IOFF ratio
	EnduranceCycles int64  // Write endurance

	// Performance
	ReadLatencyNS  float64 // Read latency (ns)
	WriteLatencyNS float64 // Write latency (ns)
	EnergyPerMAC   float64 // fJ per MAC

	// Accuracy impact
	C2CVariation   float64 // Cycle-to-cycle variation
	D2DVariation   float64 // Device-to-device variation
}

// DefaultFeFETCIMConfig returns FELIX-style parameters
func DefaultFeFETCIMConfig() *FeFETCIMConfig {
	return &FeFETCIMConfig{
		ArrayRows:       128,
		ArrayCols:       128,
		CellType:        "1FeFET1R",
		WeightBits:      4,
		Vth0:            0.4,
		MemoryWindow:    1.0,
		OnOffRatio:      1e4,
		EnduranceCycles: 1e10,
		ReadLatencyNS:   10.0,
		WriteLatencyNS:  100.0,
		EnergyPerMAC:    0.1,   // 0.1 fJ/MAC
		C2CVariation:    0.03,  // 3%
		D2DVariation:    0.05,  // 5%
	}
}

// KeywordSpottingAccelerator implements always-on KWS
type KeywordSpottingAccelerator struct {
	Config       *TinyMLConfig
	CIMConfig    *FeFETCIMConfig
	MFCCExtractor *MFCCExtractor
	DNNLayers    []*FeFETCIMLayer
	Stats        *KWSStats
}

// KWSStats tracks KWS performance
type KWSStats struct {
	TotalInferences int
	TotalKeywords   int
	TruePositives   int
	FalsePositives  int
	FalseNegatives  int
	Accuracy        float64
	AveragePowerUW  float64
	EnergyPerInfNJ  float64
	LatencyMS       float64
}

// MFCCExtractor extracts Mel-frequency cepstral coefficients
type MFCCExtractor struct {
	NumCoeffs     int       // Number of MFCC coefficients
	NumFilters    int       // Number of mel filters
	FrameSize     int       // Samples per frame
	FrameStride   int       // Frame stride
	SampleRate    int       // Audio sample rate
	FFTSize       int       // FFT size
	MelFilters    [][]float64 // Mel filterbank
}

// NewMFCCExtractor creates MFCC feature extractor
func NewMFCCExtractor(numCoeffs, sampleRate int) *MFCCExtractor {
	frameSize := sampleRate * 25 / 1000  // 25ms frame
	frameStride := sampleRate * 10 / 1000 // 10ms stride
	fftSize := 512
	numFilters := 40

	// Create mel filterbank
	melFilters := make([][]float64, numFilters)
	for i := 0; i < numFilters; i++ {
		melFilters[i] = make([]float64, fftSize/2+1)
		// Triangular filter (simplified)
		center := float64(i+1) * float64(fftSize/2) / float64(numFilters+1)
		width := float64(fftSize/2) / float64(numFilters)
		for j := 0; j < fftSize/2+1; j++ {
			dist := math.Abs(float64(j) - center)
			if dist < width {
				melFilters[i][j] = 1 - dist/width
			}
		}
	}

	return &MFCCExtractor{
		NumCoeffs:   numCoeffs,
		NumFilters:  numFilters,
		FrameSize:   frameSize,
		FrameStride: frameStride,
		SampleRate:  sampleRate,
		FFTSize:     fftSize,
		MelFilters:  melFilters,
	}
}

// Extract computes MFCC features from audio
func (mfcc *MFCCExtractor) Extract(audio []float64) [][]float64 {
	numFrames := (len(audio) - mfcc.FrameSize) / mfcc.FrameStride + 1
	if numFrames < 1 {
		numFrames = 1
	}

	features := make([][]float64, numFrames)
	for f := 0; f < numFrames; f++ {
		start := f * mfcc.FrameStride
		end := start + mfcc.FrameSize
		if end > len(audio) {
			end = len(audio)
		}

		// Extract frame
		frame := make([]float64, mfcc.FFTSize)
		for i := start; i < end && i-start < len(frame); i++ {
			// Apply Hamming window
			n := float64(i - start)
			window := 0.54 - 0.46*math.Cos(2*math.Pi*n/float64(end-start-1))
			frame[i-start] = audio[i] * window
		}

		// Simplified FFT magnitude (actual implementation would use real FFT)
		fftMag := make([]float64, mfcc.FFTSize/2+1)
		for k := 0; k < len(fftMag); k++ {
			real, imag := 0.0, 0.0
			for n := 0; n < len(frame); n++ {
				angle := -2 * math.Pi * float64(k*n) / float64(mfcc.FFTSize)
				real += frame[n] * math.Cos(angle)
				imag += frame[n] * math.Sin(angle)
			}
			fftMag[k] = math.Sqrt(real*real + imag*imag)
		}

		// Apply mel filterbank
		melEnergies := make([]float64, mfcc.NumFilters)
		for i := 0; i < mfcc.NumFilters; i++ {
			for j := 0; j < len(fftMag) && j < len(mfcc.MelFilters[i]); j++ {
				melEnergies[i] += fftMag[j] * mfcc.MelFilters[i][j]
			}
			melEnergies[i] = math.Log(melEnergies[i] + 1e-10)
		}

		// DCT to get MFCC (simplified)
		features[f] = make([]float64, mfcc.NumCoeffs)
		for k := 0; k < mfcc.NumCoeffs; k++ {
			for n := 0; n < mfcc.NumFilters; n++ {
				features[f][k] += melEnergies[n] * math.Cos(math.Pi*float64(k)*(float64(n)+0.5)/float64(mfcc.NumFilters))
			}
			features[f][k] *= math.Sqrt(2.0 / float64(mfcc.NumFilters))
		}
	}

	return features
}

// FeFETCIMLayer implements a CIM layer with FeFET cells
type FeFETCIMLayer struct {
	Config      *FeFETCIMConfig
	Weights     [][]float64
	Bias        []float64
	InputSize   int
	OutputSize  int
	Activation  string // "relu", "sigmoid", "none"
}

// NewFeFETCIMLayer creates FeFET CIM layer
func NewFeFETCIMLayer(inputSize, outputSize int, config *FeFETCIMConfig, activation string) *FeFETCIMLayer {
	// Initialize weights with Xavier initialization
	scale := math.Sqrt(2.0 / float64(inputSize+outputSize))
	weights := make([][]float64, outputSize)
	bias := make([]float64, outputSize)

	for i := 0; i < outputSize; i++ {
		weights[i] = make([]float64, inputSize)
		for j := 0; j < inputSize; j++ {
			weights[i][j] = rand.NormFloat64() * scale
		}
	}

	return &FeFETCIMLayer{
		Config:     config,
		Weights:    weights,
		Bias:       bias,
		InputSize:  inputSize,
		OutputSize: outputSize,
		Activation: activation,
	}
}

// Forward performs CIM-based matrix-vector multiplication
func (layer *FeFETCIMLayer) Forward(input []float64) []float64 {
	output := make([]float64, layer.OutputSize)
	cfg := layer.Config

	for i := 0; i < layer.OutputSize; i++ {
		sum := layer.Bias[i]
		for j := 0; j < layer.InputSize && j < len(input); j++ {
			// Add FeFET variability
			c2cNoise := rand.NormFloat64() * cfg.C2CVariation
			d2dNoise := rand.NormFloat64() * cfg.D2DVariation
			effectiveWeight := layer.Weights[i][j] * (1 + c2cNoise + d2dNoise)
			sum += effectiveWeight * input[j]
		}

		// Apply activation
		switch layer.Activation {
		case "relu":
			output[i] = math.Max(0, sum)
		case "sigmoid":
			output[i] = 1.0 / (1.0 + math.Exp(-sum))
		default:
			output[i] = sum
		}
	}

	return output
}

// NewKeywordSpottingAccelerator creates KWS accelerator
func NewKeywordSpottingAccelerator(config *TinyMLConfig, cimConfig *FeFETCIMConfig) *KeywordSpottingAccelerator {
	// Create DNN layers
	layers := make([]*FeFETCIMLayer, 0)

	// Input -> First hidden
	prevSize := config.InputSize
	for _, hiddenSize := range config.HiddenLayers {
		layers = append(layers, NewFeFETCIMLayer(prevSize, hiddenSize, cimConfig, "relu"))
		prevSize = hiddenSize
	}

	// Last hidden -> Output
	layers = append(layers, NewFeFETCIMLayer(prevSize, config.OutputClasses, cimConfig, "none"))

	return &KeywordSpottingAccelerator{
		Config:        config,
		CIMConfig:     cimConfig,
		MFCCExtractor: NewMFCCExtractor(10, 16000),
		DNNLayers:     layers,
		Stats:         &KWSStats{},
	}
}

// ProcessAudio performs keyword spotting on audio
func (kws *KeywordSpottingAccelerator) ProcessAudio(audio []float64) (int, float64) {
	// Extract MFCC features
	mfccFeatures := kws.MFCCExtractor.Extract(audio)

	// Flatten features
	input := make([]float64, 0)
	for _, frame := range mfccFeatures {
		input = append(input, frame...)
	}

	// Pad/truncate to expected size
	if len(input) < kws.Config.InputSize {
		padded := make([]float64, kws.Config.InputSize)
		copy(padded, input)
		input = padded
	} else if len(input) > kws.Config.InputSize {
		input = input[:kws.Config.InputSize]
	}

	// Forward through DNN layers
	activation := input
	for _, layer := range kws.DNNLayers {
		activation = layer.Forward(activation)
	}

	// Softmax and argmax
	maxVal := activation[0]
	maxIdx := 0
	expSum := 0.0
	for i, v := range activation {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
		expSum += math.Exp(v - maxVal)
	}
	confidence := math.Exp(activation[maxIdx]-maxVal) / expSum

	// Update stats
	kws.Stats.TotalInferences++
	kws.Stats.EnergyPerInfNJ = kws.Config.EnergyPerInfNJ
	kws.Stats.AveragePowerUW = kws.Config.ActivePowerUW
	kws.Stats.LatencyMS = float64(kws.Config.InputSize) / (kws.Config.ClockFreqKHz)

	return maxIdx, confidence
}

// =============================================================================
// WAKE WORD DETECTION (ULTRA-LOW POWER)
// =============================================================================

// WakeWordConfig configures wake word detector
type WakeWordConfig struct {
	// Model
	NumKeywords    int
	WindowMS       int     // Detection window (ms)
	ThresholdProb  float64 // Detection threshold

	// Power modes
	AlwaysOnPowerNW float64 // Always-on listening power (nW)
	ActivePowerUW   float64 // Active inference power (µW)
	WakeupLatencyMS float64 // Latency to wake up

	// Hardware
	Technology      string
	SupplyVoltage   float64
}

// DefaultWakeWordConfig returns Syntiant NDP-style config
func DefaultWakeWordConfig() *WakeWordConfig {
	return &WakeWordConfig{
		NumKeywords:     10,
		WindowMS:        1000,
		ThresholdProb:   0.8,
		AlwaysOnPowerNW: 150000, // 150 µW (Syntiant NDP100)
		ActivePowerUW:   140.0,
		WakeupLatencyMS: 10.0,
		Technology:      "40nm",
		SupplyVoltage:   0.9,
	}
}

// UltraLowPowerWakeWordConfig returns sub-µW config
func UltraLowPowerWakeWordConfig() *WakeWordConfig {
	return &WakeWordConfig{
		NumKeywords:     35,
		WindowMS:        1000,
		ThresholdProb:   0.7,
		AlwaysOnPowerNW: 510,    // 510 nW (ISSCC record)
		ActivePowerUW:   0.61,   // 0.61 µW
		WakeupLatencyMS: 20.0,
		Technology:      "28nm",
		SupplyVoltage:   0.41,
	}
}

// WakeWordDetector implements always-on wake word detection
type WakeWordDetector struct {
	Config      *WakeWordConfig
	CIMAccel    *KeywordSpottingAccelerator
	AudioBuffer []float64
	BufferIdx   int
	IsAwake     bool
	Stats       *WakeWordStats
}

// WakeWordStats tracks wake word performance
type WakeWordStats struct {
	TotalListeningTimeMS int64
	TotalWakeups         int
	FalseWakeups         int
	MissedWakeups        int
	AveragePowerNW       float64
	BatteryLifeDays      float64
}

// NewWakeWordDetector creates wake word detector
func NewWakeWordDetector(config *WakeWordConfig) *WakeWordDetector {
	tinyConfig := &TinyMLConfig{
		ModelType:      "kws",
		InputSize:      490,
		HiddenLayers:   []int{128, 128},
		OutputClasses:  config.NumKeywords + 2,
		Quantization:   8,
		ClockFreqKHz:   100.0,
		SupplyVoltage:  config.SupplyVoltage,
		Technology:     config.Technology,
		ActivePowerUW:  config.ActivePowerUW,
		EnergyPerInfNJ: config.ActivePowerUW * 10, // ~10ms inference
	}

	bufferSize := config.WindowMS * 16 // 16 kHz sample rate
	return &WakeWordDetector{
		Config:      config,
		CIMAccel:    NewKeywordSpottingAccelerator(tinyConfig, DefaultFeFETCIMConfig()),
		AudioBuffer: make([]float64, bufferSize),
		BufferIdx:   0,
		IsAwake:     false,
		Stats:       &WakeWordStats{},
	}
}

// ProcessSample processes single audio sample (streaming)
func (wwd *WakeWordDetector) ProcessSample(sample float64) (bool, int) {
	wwd.AudioBuffer[wwd.BufferIdx] = sample
	wwd.BufferIdx = (wwd.BufferIdx + 1) % len(wwd.AudioBuffer)
	wwd.Stats.TotalListeningTimeMS++

	// Check for wake word periodically
	if wwd.BufferIdx == 0 {
		keyword, confidence := wwd.CIMAccel.ProcessAudio(wwd.AudioBuffer)
		if confidence > wwd.Config.ThresholdProb && keyword < wwd.Config.NumKeywords {
			wwd.IsAwake = true
			wwd.Stats.TotalWakeups++
			return true, keyword
		}
	}

	return false, -1
}

// EstimateBatteryLife estimates battery life in days
func (wwd *WakeWordDetector) EstimateBatteryLife(batteryMAH float64) float64 {
	// Average power considering duty cycle
	inferenceRate := 1.0 / float64(wwd.Config.WindowMS) * 1000 // inferences per second
	activeDuty := inferenceRate * wwd.Config.WakeupLatencyMS / 1000

	avgPowerUW := wwd.Config.AlwaysOnPowerNW/1000 + activeDuty*wwd.Config.ActivePowerUW
	avgCurrentUA := avgPowerUW / (wwd.Config.SupplyVoltage * 1000) // Simplified

	// Battery life in hours, then days
	batteryLifeHours := (batteryMAH * 1000) / avgCurrentUA
	wwd.Stats.BatteryLifeDays = batteryLifeHours / 24

	return wwd.Stats.BatteryLifeDays
}

// =============================================================================
// PART 2: NEUROMORPHIC SENSOR FUSION
// =============================================================================

// EventCameraConfig configures dynamic vision sensor
type EventCameraConfig struct {
	Width          int     // Pixel width
	Height         int     // Pixel height
	TemporalResUS  float64 // Temporal resolution (µs)
	ContrastThresh float64 // Contrast threshold for event
	RefractoryUS   float64 // Refractory period (µs)
	Latency        float64 // Sensor latency (µs)
	DynamicRange   float64 // Dynamic range (dB)
	Bandwidth      float64 // Event bandwidth (Meps)
}

// DefaultDVSConfig returns DVS128 style config
func DefaultDVSConfig() *EventCameraConfig {
	return &EventCameraConfig{
		Width:          128,
		Height:         128,
		TemporalResUS:  1.0,
		ContrastThresh: 0.15,
		RefractoryUS:   1.0,
		Latency:        15.0,
		DynamicRange:   120.0,
		Bandwidth:      12.0,
	}
}

// HighResDVSConfig returns Prophesee EVK4 style config
func HighResDVSConfig() *EventCameraConfig {
	return &EventCameraConfig{
		Width:          1280,
		Height:         720,
		TemporalResUS:  0.2,
		ContrastThresh: 0.1,
		RefractoryUS:   0.5,
		Latency:        10.0,
		DynamicRange:   140.0,
		Bandwidth:      1200.0,
	}
}

// Event represents a single DVS event
type Event struct {
	X         int     // Pixel x coordinate
	Y         int     // Pixel y coordinate
	Timestamp float64 // Timestamp (µs)
	Polarity  int     // +1 for ON, -1 for OFF
}

// EventStream represents stream of DVS events
type EventStream struct {
	Events    []Event
	StartTime float64
	EndTime   float64
}

// DVSSimulator simulates dynamic vision sensor
type DVSSimulator struct {
	Config        *EventCameraConfig
	LastIntensity [][]float64
	LastEventTime [][]float64
}

// NewDVSSimulator creates DVS simulator
func NewDVSSimulator(config *EventCameraConfig) *DVSSimulator {
	lastInt := make([][]float64, config.Height)
	lastTime := make([][]float64, config.Height)
	for y := 0; y < config.Height; y++ {
		lastInt[y] = make([]float64, config.Width)
		lastTime[y] = make([]float64, config.Width)
		for x := 0; x < config.Width; x++ {
			lastInt[y][x] = 0.5 // Mid-gray initial
		}
	}

	return &DVSSimulator{
		Config:        config,
		LastIntensity: lastInt,
		LastEventTime: lastTime,
	}
}

// ProcessFrame converts frame to events
func (dvs *DVSSimulator) ProcessFrame(frame [][]float64, timestamp float64) *EventStream {
	events := make([]Event, 0)
	cfg := dvs.Config

	for y := 0; y < cfg.Height && y < len(frame); y++ {
		for x := 0; x < cfg.Width && x < len(frame[y]); x++ {
			intensity := frame[y][x]

			// Check refractory period
			if timestamp-dvs.LastEventTime[y][x] < cfg.RefractoryUS {
				continue
			}

			// Compute log intensity change
			if dvs.LastIntensity[y][x] > 0 && intensity > 0 {
				logChange := math.Log(intensity) - math.Log(dvs.LastIntensity[y][x])

				// Generate events based on threshold crossings
				if math.Abs(logChange) >= cfg.ContrastThresh {
					polarity := 1
					if logChange < 0 {
						polarity = -1
					}

					events = append(events, Event{
						X:         x,
						Y:         y,
						Timestamp: timestamp + rand.Float64()*cfg.Latency,
						Polarity:  polarity,
					})

					dvs.LastEventTime[y][x] = timestamp
				}
			}

			dvs.LastIntensity[y][x] = intensity
		}
	}

	return &EventStream{
		Events:    events,
		StartTime: timestamp,
		EndTime:   timestamp + 1000, // 1ms window
	}
}

// =============================================================================
// IN-SENSOR COMPUTING
// =============================================================================

// InSensorConfig configures in-sensor compute pixel
type InSensorConfig struct {
	// Pixel array
	PixelRows      int
	PixelCols      int
	PixelType      string // "MACPix", "1P1R", "OEM"

	// In-pixel compute
	WeightBits     int
	ActivationBits int
	ComputeType    string // "analog", "digital", "hybrid"

	// FeFET parameters (for MACPix)
	FeFETStates    int     // Multi-level states
	PolarizationUC float64 // Polarization (µC/cm²)

	// Performance
	FrameRate      float64 // Frames per second
	PowerPerPixelPW float64 // Power per pixel (pW)
	EnergyPerInfNJ float64 // Energy per inference
}

// DefaultMACPixConfig returns FeFET MACPix config
func DefaultMACPixConfig() *InSensorConfig {
	return &InSensorConfig{
		PixelRows:       64,
		PixelCols:       64,
		PixelType:       "MACPix",
		WeightBits:      4,
		ActivationBits:  8,
		ComputeType:     "analog",
		FeFETStates:     16,
		PolarizationUC:  25.0,
		FrameRate:       1000.0,
		PowerPerPixelPW: 100.0,
		EnergyPerInfNJ:  0.05,
	}
}

// InSensorComputeArray implements in-pixel MAC operations
type InSensorComputeArray struct {
	Config      *InSensorConfig
	Weights     [][]float64 // Stored in FeFET pixels
	PhotoValues [][]float64 // Current photo-generated values
	Stats       *InSensorStats
}

// InSensorStats tracks in-sensor performance
type InSensorStats struct {
	TotalFrames     int
	TotalMACs       int64
	TotalEnergyPJ   float64
	AveragePowerUW  float64
	Throughput      float64 // GOPS
}

// NewInSensorComputeArray creates in-sensor compute array
func NewInSensorComputeArray(config *InSensorConfig) *InSensorComputeArray {
	weights := make([][]float64, config.PixelRows)
	photoValues := make([][]float64, config.PixelRows)

	for y := 0; y < config.PixelRows; y++ {
		weights[y] = make([]float64, config.PixelCols)
		photoValues[y] = make([]float64, config.PixelCols)
		for x := 0; x < config.PixelCols; x++ {
			// Initialize random weights
			weights[y][x] = rand.Float64()*2 - 1
		}
	}

	return &InSensorComputeArray{
		Config:      config,
		Weights:     weights,
		PhotoValues: photoValues,
		Stats:       &InSensorStats{},
	}
}

// ComputeMAC performs in-pixel MAC operation
func (isc *InSensorComputeArray) ComputeMAC(image [][]float64) []float64 {
	cfg := isc.Config
	output := make([]float64, cfg.PixelCols)

	// Each column accumulates weighted sum of pixel values
	for x := 0; x < cfg.PixelCols; x++ {
		sum := 0.0
		for y := 0; y < cfg.PixelRows && y < len(image); y++ {
			pixelValue := 0.0
			if x < len(image[y]) {
				pixelValue = image[y][x]
			}

			// Quantize weight based on FeFET states
			quantWeight := math.Round(isc.Weights[y][x]*float64(cfg.FeFETStates-1)) / float64(cfg.FeFETStates-1)

			// In-pixel multiply-accumulate
			sum += pixelValue * quantWeight
			isc.Stats.TotalMACs++
		}
		output[x] = sum
	}

	isc.Stats.TotalFrames++
	isc.Stats.TotalEnergyPJ += cfg.EnergyPerInfNJ * 1000
	isc.Stats.AveragePowerUW = cfg.PowerPerPixelPW * float64(cfg.PixelRows*cfg.PixelCols) / 1e6

	return output
}

// =============================================================================
// MULTI-MODAL SENSOR FUSION
// =============================================================================

// MultiModalFusionConfig configures multi-sensor fusion
type MultiModalFusionConfig struct {
	// Modalities
	EnableVision    bool
	EnableAudio     bool
	EnableIMU       bool
	EnableTactile   bool

	// Fusion strategy
	FusionLevel     string // "early", "mid", "late", "hierarchical"
	AttentionHeads  int
	FusedDim        int

	// SNN parameters for neuromorphic fusion
	UseSNN          bool
	SNNTimesteps    int
	LIFTau          float64
}

// DefaultMultiModalConfig returns default fusion config
func DefaultMultiModalConfig() *MultiModalFusionConfig {
	return &MultiModalFusionConfig{
		EnableVision:   true,
		EnableAudio:    true,
		EnableIMU:      true,
		EnableTactile:  false,
		FusionLevel:    "mid",
		AttentionHeads: 4,
		FusedDim:       128,
		UseSNN:         true,
		SNNTimesteps:   10,
		LIFTau:         20.0,
	}
}

// SensorModality represents a single sensor input
type SensorModality struct {
	Name        string
	FeatureDim  int
	SampleRate  float64
	Features    []float64
	Spikes      [][]bool // For SNN: [timestep][neuron]
}

// MultiModalFusionSNN implements neuromorphic sensor fusion
type MultiModalFusionSNN struct {
	Config       *MultiModalFusionConfig
	Modalities   map[string]*SensorModality
	FusionLayer  *SNNFusionLayer
	OutputLayer  *SNNOutputLayer
	Stats        *FusionStats
}

// FusionStats tracks fusion performance
type FusionStats struct {
	TotalFusions    int
	Accuracy        float64
	LatencyMS       float64
	PowerUW         float64
	SpikeActivity   float64
}

// SNNFusionLayer implements SNN-based feature fusion
type SNNFusionLayer struct {
	Config       *MultiModalFusionConfig
	Weights      map[string][][]float64 // Per-modality weights
	FusedWeights [][]float64
	Membrane     []float64
	Threshold    float64
	Tau          float64
}

// NewSNNFusionLayer creates SNN fusion layer
func NewSNNFusionLayer(config *MultiModalFusionConfig, modalityDims map[string]int) *SNNFusionLayer {
	weights := make(map[string][][]float64)
	totalInput := 0

	for name, dim := range modalityDims {
		weights[name] = make([][]float64, config.FusedDim)
		for i := 0; i < config.FusedDim; i++ {
			weights[name][i] = make([]float64, dim)
			for j := 0; j < dim; j++ {
				weights[name][i][j] = rand.NormFloat64() * 0.1
			}
		}
		totalInput += dim
	}

	return &SNNFusionLayer{
		Config:    config,
		Weights:   weights,
		Membrane:  make([]float64, config.FusedDim),
		Threshold: 1.0,
		Tau:       config.LIFTau,
	}
}

// FuseSpikes fuses spike trains from multiple modalities
func (sfl *SNNFusionLayer) FuseSpikes(modalitySpikes map[string][]bool, dt float64) []bool {
	cfg := sfl.Config
	outputSpikes := make([]bool, cfg.FusedDim)

	for i := 0; i < cfg.FusedDim; i++ {
		// Accumulate input from all modalities
		totalInput := 0.0
		for name, spikes := range modalitySpikes {
			if weights, exists := sfl.Weights[name]; exists {
				for j := 0; j < len(spikes) && j < len(weights[i]); j++ {
					if spikes[j] {
						totalInput += weights[i][j]
					}
				}
			}
		}

		// LIF dynamics
		sfl.Membrane[i] = sfl.Membrane[i]*math.Exp(-dt/sfl.Tau) + totalInput

		// Check threshold
		if sfl.Membrane[i] >= sfl.Threshold {
			outputSpikes[i] = true
			sfl.Membrane[i] = 0 // Reset
		}
	}

	return outputSpikes
}

// SNNOutputLayer implements output classification
type SNNOutputLayer struct {
	NumClasses int
	Weights    [][]float64
	Membrane   []float64
	SpikeCounts []int
}

// NewSNNOutputLayer creates output layer
func NewSNNOutputLayer(inputDim, numClasses int) *SNNOutputLayer {
	weights := make([][]float64, numClasses)
	for i := 0; i < numClasses; i++ {
		weights[i] = make([]float64, inputDim)
		for j := 0; j < inputDim; j++ {
			weights[i][j] = rand.NormFloat64() * 0.1
		}
	}

	return &SNNOutputLayer{
		NumClasses:  numClasses,
		Weights:     weights,
		Membrane:    make([]float64, numClasses),
		SpikeCounts: make([]int, numClasses),
	}
}

// ProcessSpikes processes fused spikes for classification
func (sol *SNNOutputLayer) ProcessSpikes(spikes []bool, dt float64) {
	tau := 20.0

	for i := 0; i < sol.NumClasses; i++ {
		input := 0.0
		for j := 0; j < len(spikes) && j < len(sol.Weights[i]); j++ {
			if spikes[j] {
				input += sol.Weights[i][j]
			}
		}

		sol.Membrane[i] = sol.Membrane[i]*math.Exp(-dt/tau) + input

		if sol.Membrane[i] >= 1.0 {
			sol.SpikeCounts[i]++
			sol.Membrane[i] = 0
		}
	}
}

// GetPrediction returns predicted class
func (sol *SNNOutputLayer) GetPrediction() int {
	maxCount := 0
	maxClass := 0
	for i, count := range sol.SpikeCounts {
		if count > maxCount {
			maxCount = count
			maxClass = i
		}
	}
	return maxClass
}

// NewMultiModalFusionSNN creates multi-modal fusion system
func NewMultiModalFusionSNN(config *MultiModalFusionConfig) *MultiModalFusionSNN {
	modalities := make(map[string]*SensorModality)

	if config.EnableVision {
		modalities["vision"] = &SensorModality{
			Name:       "vision",
			FeatureDim: 128,
			SampleRate: 30.0,
		}
	}
	if config.EnableAudio {
		modalities["audio"] = &SensorModality{
			Name:       "audio",
			FeatureDim: 64,
			SampleRate: 16000.0,
		}
	}
	if config.EnableIMU {
		modalities["imu"] = &SensorModality{
			Name:       "imu",
			FeatureDim: 6,
			SampleRate: 100.0,
		}
	}

	modalityDims := make(map[string]int)
	for name, mod := range modalities {
		modalityDims[name] = mod.FeatureDim
	}

	return &MultiModalFusionSNN{
		Config:      config,
		Modalities:  modalities,
		FusionLayer: NewSNNFusionLayer(config, modalityDims),
		OutputLayer: NewSNNOutputLayer(config.FusedDim, 10),
		Stats:       &FusionStats{},
	}
}

// ProcessMultiModal performs multi-modal inference
func (mmf *MultiModalFusionSNN) ProcessMultiModal(inputs map[string][]float64) int {
	cfg := mmf.Config
	dt := 1.0 // 1ms timestep

	// Reset spike counts
	mmf.OutputLayer.SpikeCounts = make([]int, mmf.OutputLayer.NumClasses)

	// Simulate over timesteps
	for t := 0; t < cfg.SNNTimesteps; t++ {
		modalitySpikes := make(map[string][]bool)

		// Convert inputs to spikes (rate coding)
		for name, input := range inputs {
			if mod, exists := mmf.Modalities[name]; exists {
				spikes := make([]bool, mod.FeatureDim)
				for i := 0; i < mod.FeatureDim && i < len(input); i++ {
					// Poisson spike generation
					rate := math.Max(0, input[i])
					if rand.Float64() < rate*dt/1000 {
						spikes[i] = true
					}
				}
				modalitySpikes[name] = spikes
			}
		}

		// Fuse spikes
		fusedSpikes := mmf.FusionLayer.FuseSpikes(modalitySpikes, dt)

		// Output classification
		mmf.OutputLayer.ProcessSpikes(fusedSpikes, dt)
	}

	mmf.Stats.TotalFusions++
	return mmf.OutputLayer.GetPrediction()
}

// =============================================================================
// INTEGRATED EDGE + NEUROMORPHIC DEMO
// =============================================================================

// EdgeNeuromorphicDemo demonstrates integrated capabilities
type EdgeNeuromorphicDemo struct {
	// Edge AI components
	KWSAccel      *KeywordSpottingAccelerator
	WakeWord      *WakeWordDetector

	// Neuromorphic components
	DVS           *DVSSimulator
	InSensor      *InSensorComputeArray
	MultiModal    *MultiModalFusionSNN
}

// NewEdgeNeuromorphicDemo creates integrated demo
func NewEdgeNeuromorphicDemo() *EdgeNeuromorphicDemo {
	return &EdgeNeuromorphicDemo{
		KWSAccel:   NewKeywordSpottingAccelerator(DefaultKWSConfig(), DefaultFeFETCIMConfig()),
		WakeWord:   NewWakeWordDetector(UltraLowPowerWakeWordConfig()),
		DVS:        NewDVSSimulator(DefaultDVSConfig()),
		InSensor:   NewInSensorComputeArray(DefaultMACPixConfig()),
		MultiModal: NewMultiModalFusionSNN(DefaultMultiModalConfig()),
	}
}

// RunKWSDemo demonstrates keyword spotting
func (demo *EdgeNeuromorphicDemo) RunKWSDemo() map[string]interface{} {
	// Generate synthetic audio (1 second at 16kHz)
	audio := make([]float64, 16000)
	for i := range audio {
		// Simulate speech-like signal
		t := float64(i) / 16000.0
		audio[i] = 0.5 * math.Sin(2*math.Pi*200*t) * math.Exp(-t*2)
		audio[i] += rand.NormFloat64() * 0.1
	}

	keyword, confidence := demo.KWSAccel.ProcessAudio(audio)

	return map[string]interface{}{
		"detected_keyword": keyword,
		"confidence":       confidence,
		"power_uw":         demo.KWSAccel.Stats.AveragePowerUW,
		"energy_nj":        demo.KWSAccel.Stats.EnergyPerInfNJ,
		"latency_ms":       demo.KWSAccel.Stats.LatencyMS,
	}
}

// RunDVSDemo demonstrates event camera processing
func (demo *EdgeNeuromorphicDemo) RunDVSDemo() map[string]interface{} {
	cfg := demo.DVS.Config

	// Generate synthetic moving edge
	frame1 := make([][]float64, cfg.Height)
	frame2 := make([][]float64, cfg.Height)
	for y := 0; y < cfg.Height; y++ {
		frame1[y] = make([]float64, cfg.Width)
		frame2[y] = make([]float64, cfg.Width)
		for x := 0; x < cfg.Width; x++ {
			// Vertical edge moving right
			if x < cfg.Width/2 {
				frame1[y][x] = 0.2
				frame2[y][x] = 0.8
			} else {
				frame1[y][x] = 0.8
				frame2[y][x] = 0.2
			}
		}
	}

	// Process frames
	events1 := demo.DVS.ProcessFrame(frame1, 0)
	events2 := demo.DVS.ProcessFrame(frame2, 1000)

	return map[string]interface{}{
		"frame1_events": len(events1.Events),
		"frame2_events": len(events2.Events),
		"temporal_res":  cfg.TemporalResUS,
		"bandwidth":     cfg.Bandwidth,
	}
}

// RunInSensorDemo demonstrates in-sensor computing
func (demo *EdgeNeuromorphicDemo) RunInSensorDemo() map[string]interface{} {
	cfg := demo.InSensor.Config

	// Generate test image
	image := make([][]float64, cfg.PixelRows)
	for y := 0; y < cfg.PixelRows; y++ {
		image[y] = make([]float64, cfg.PixelCols)
		for x := 0; x < cfg.PixelCols; x++ {
			image[y][x] = rand.Float64()
		}
	}

	// Perform in-sensor MAC
	output := demo.InSensor.ComputeMAC(image)

	return map[string]interface{}{
		"output_dim":    len(output),
		"total_macs":    demo.InSensor.Stats.TotalMACs,
		"power_uw":      demo.InSensor.Stats.AveragePowerUW,
		"pixel_type":    cfg.PixelType,
		"fefet_states":  cfg.FeFETStates,
	}
}

// RunMultiModalDemo demonstrates multi-modal fusion
func (demo *EdgeNeuromorphicDemo) RunMultiModalDemo() map[string]interface{} {
	// Generate synthetic sensor inputs
	inputs := map[string][]float64{
		"vision": make([]float64, 128),
		"audio":  make([]float64, 64),
		"imu":    make([]float64, 6),
	}

	for i := range inputs["vision"] {
		inputs["vision"][i] = rand.Float64() * 100 // Rate in Hz
	}
	for i := range inputs["audio"] {
		inputs["audio"][i] = rand.Float64() * 50
	}
	for i := range inputs["imu"] {
		inputs["imu"][i] = rand.Float64() * 20
	}

	prediction := demo.MultiModal.ProcessMultiModal(inputs)

	return map[string]interface{}{
		"prediction":     prediction,
		"total_fusions":  demo.MultiModal.Stats.TotalFusions,
		"fusion_level":   demo.MultiModal.Config.FusionLevel,
		"snn_timesteps":  demo.MultiModal.Config.SNNTimesteps,
	}
}

// GetPerformanceSummary returns key metrics
func (demo *EdgeNeuromorphicDemo) GetPerformanceSummary() map[string]float64 {
	return map[string]float64{
		// KWS (from papers)
		"kws_power_uw":           3.8,
		"kws_energy_nj":          183.4,
		"kws_accuracy_pct":       90.6,

		// Wake word (ISSCC record)
		"wakeword_power_nw":      510,
		"wakeword_keywords":      35,
		"wakeword_voltage_v":     0.41,

		// FeFET CIM
		"fefet_tops_w":           885,
		"fefet_accuracy_pct":     96.6,
		"fefet_energy_fj_mac":    0.1,

		// In-sensor (MACPix)
		"insensor_states":        16,
		"insensor_frame_rate":    1000,

		// DVS
		"dvs_temporal_res_us":    1.0,
		"dvs_dynamic_range_db":   120,

		// Multi-modal fusion
		"fusion_power_uw":        20,
		"fusion_modalities":      3,
	}
}
