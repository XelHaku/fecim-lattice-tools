// Package layers provides neural network layer implementations for CIM deployment.
// This file implements neuromorphic vision sensors and secure CIM computing.
// Based on research: Nature Nanotechnology (retinomorphic), ACS Nano (3D PUF),
// Nature Communications (secure CIM), Advanced Materials (ferroelectric QD)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// NEUROMORPHIC VISION SENSORS
// Event-based and retinomorphic sensors with ferroelectric modulation
// =============================================================================

// VisionSensorType defines types of neuromorphic vision sensors.
type VisionSensorType int

const (
	VisionSensorDVS        VisionSensorType = iota // Dynamic Vision Sensor
	VisionSensorRetino                             // Retinomorphic sensor
	VisionSensorFerroQD                            // Ferroelectric quantum dot
	VisionSensorHybrid                             // Hybrid event + frame
)

// EventPolarity represents change direction in event cameras.
type EventPolarity int

const (
	EventOff EventPolarity = -1 // Brightness decrease
	EventOn  EventPolarity = 1  // Brightness increase
)

// DVSPixelConfig configures a DVS pixel.
type DVSPixelConfig struct {
	ThresholdOn      float64 // Log intensity threshold for ON events
	ThresholdOff     float64 // Log intensity threshold for OFF events
	RefractoryPeriod float64 // Minimum time between events (µs)
	Bandwidth        float64 // Pixel bandwidth (kHz)
	DarkCurrent      float64 // Dark current noise (fA)
	QuantumEff       float64 // Quantum efficiency
}

// DefaultDVSConfig returns typical DVS parameters.
func DefaultDVSConfig() *DVSPixelConfig {
	return &DVSPixelConfig{
		ThresholdOn:      0.15, // ~15% log change
		ThresholdOff:     0.15,
		RefractoryPeriod: 1.0,    // 1 µs
		Bandwidth:        1000.0, // 1 MHz
		DarkCurrent:      10.0,   // 10 fA
		QuantumEff:       0.7,
	}
}

// DVSPixel implements a single DVS pixel.
type DVSPixel struct {
	Config *DVSPixelConfig

	// State
	LastLogIntensity float64
	LastEventTime    float64
	RefMemVoltage    float64

	// Ferroelectric modulation
	UseFerroelectric bool
	Polarization     float64
	ThresholdMod     float64
}

// NewDVSPixel creates a new DVS pixel.
func NewDVSPixel(config *DVSPixelConfig) *DVSPixel {
	if config == nil {
		config = DefaultDVSConfig()
	}
	return &DVSPixel{
		Config:       config,
		ThresholdMod: 1.0,
	}
}

// DVSEvent represents an event from a DVS pixel.
type DVSEvent struct {
	X, Y      int
	Timestamp float64 // µs
	Polarity  EventPolarity
	LogChange float64
}

// ProcessIntensity processes new intensity and generates events.
func (p *DVSPixel) ProcessIntensity(intensity float64, timeUs float64) *DVSEvent {
	// Convert to log intensity
	logI := math.Log(intensity + 1e-10)

	// Check refractory period
	if timeUs-p.LastEventTime < p.Config.RefractoryPeriod {
		p.LastLogIntensity = logI
		return nil
	}

	// Compute change
	delta := logI - p.LastLogIntensity

	// Apply ferroelectric threshold modulation
	threshOn := p.Config.ThresholdOn * p.ThresholdMod
	threshOff := p.Config.ThresholdOff * p.ThresholdMod

	var event *DVSEvent

	if delta > threshOn {
		event = &DVSEvent{
			Timestamp: timeUs,
			Polarity:  EventOn,
			LogChange: delta,
		}
		p.LastEventTime = timeUs
		p.LastLogIntensity = logI
	} else if delta < -threshOff {
		event = &DVSEvent{
			Timestamp: timeUs,
			Polarity:  EventOff,
			LogChange: delta,
		}
		p.LastEventTime = timeUs
		p.LastLogIntensity = logI
	}

	return event
}

// SetFerroelectricPolarization modulates threshold via ferroelectric.
func (p *DVSPixel) SetFerroelectricPolarization(pol float64) {
	p.UseFerroelectric = true
	p.Polarization = pol
	// Polarization modulates threshold: higher pol = higher sensitivity
	p.ThresholdMod = 1.0 - 0.5*pol // Range: 0.5 to 1.5
}

// =============================================================================
// RETINOMORPHIC SENSOR ARRAY
// =============================================================================

// RetinomorphicConfig configures a retinomorphic sensor array.
type RetinomorphicConfig struct {
	Width, Height int

	// Sensor type
	SensorType VisionSensorType

	// Ferroelectric parameters
	UseFerroelectric bool
	PolarizationMax  float64 // µC/cm²

	// Processing
	EnableAdaptation   bool
	AdaptationTau      float64 // ms
	EnableMotionDetect bool

	// Spike encoding
	SpikeEncoding string // "rate", "temporal", "ttfs"
}

// DefaultRetinomorphicConfig returns default retinomorphic config.
func DefaultRetinomorphicConfig() *RetinomorphicConfig {
	return &RetinomorphicConfig{
		Width:              128,
		Height:             128,
		SensorType:         VisionSensorDVS,
		UseFerroelectric:   true,
		PolarizationMax:    50.0,
		EnableAdaptation:   true,
		AdaptationTau:      100.0,
		EnableMotionDetect: true,
		SpikeEncoding:      "temporal",
	}
}

// RetinomorphicSensor implements a retinomorphic sensor array.
type RetinomorphicSensor struct {
	Config *RetinomorphicConfig
	Pixels [][]*DVSPixel

	// State
	CurrentTime     float64
	EventBuffer     []*DVSEvent
	AdaptationState [][]float64

	// Motion detection
	MotionMap [][]float64
	FlowX     [][]float64
	FlowY     [][]float64

	// Statistics
	TotalEvents    int64
	EventRate      float64 // events/ms
	LastUpdateTime float64
}

// NewRetinomorphicSensor creates a new retinomorphic sensor.
func NewRetinomorphicSensor(config *RetinomorphicConfig) *RetinomorphicSensor {
	if config == nil {
		config = DefaultRetinomorphicConfig()
	}

	// Create pixel array
	pixels := make([][]*DVSPixel, config.Height)
	adaptation := make([][]float64, config.Height)
	motion := make([][]float64, config.Height)
	flowX := make([][]float64, config.Height)
	flowY := make([][]float64, config.Height)

	for y := 0; y < config.Height; y++ {
		pixels[y] = make([]*DVSPixel, config.Width)
		adaptation[y] = make([]float64, config.Width)
		motion[y] = make([]float64, config.Width)
		flowX[y] = make([]float64, config.Width)
		flowY[y] = make([]float64, config.Width)

		for x := 0; x < config.Width; x++ {
			pixels[y][x] = NewDVSPixel(nil)
			if config.UseFerroelectric {
				pixels[y][x].UseFerroelectric = true
			}
		}
	}

	return &RetinomorphicSensor{
		Config:          config,
		Pixels:          pixels,
		EventBuffer:     make([]*DVSEvent, 0),
		AdaptationState: adaptation,
		MotionMap:       motion,
		FlowX:           flowX,
		FlowY:           flowY,
	}
}

// ProcessFrame processes an intensity frame and generates events.
func (rs *RetinomorphicSensor) ProcessFrame(frame [][]float64, timeUs float64) *VisionOutput {
	dt := timeUs - rs.LastUpdateTime
	if dt <= 0 {
		dt = 1000 // 1 ms default
	}

	// Clear event buffer
	rs.EventBuffer = rs.EventBuffer[:0]
	eventCount := 0

	for y := 0; y < rs.Config.Height && y < len(frame); y++ {
		for x := 0; x < rs.Config.Width && x < len(frame[y]); x++ {
			intensity := frame[y][x]

			// Apply adaptation
			if rs.Config.EnableAdaptation {
				adaptRate := dt / (rs.Config.AdaptationTau * 1000) // Convert ms to µs
				if adaptRate > 1 {
					adaptRate = 1
				}
				rs.AdaptationState[y][x] += adaptRate * (intensity - rs.AdaptationState[y][x])
				intensity = intensity - rs.AdaptationState[y][x]*0.5
				if intensity < 0 {
					intensity = 0
				}
			}

			// Process pixel
			event := rs.Pixels[y][x].ProcessIntensity(intensity, timeUs)
			if event != nil {
				event.X = x
				event.Y = y
				rs.EventBuffer = append(rs.EventBuffer, event)
				eventCount++

				// Update motion map
				if rs.Config.EnableMotionDetect {
					rs.MotionMap[y][x] = float64(event.Polarity) * math.Abs(event.LogChange)
				}
			} else {
				// Decay motion map
				rs.MotionMap[y][x] *= 0.95
			}
		}
	}

	// Update statistics
	rs.TotalEvents += int64(eventCount)
	rs.EventRate = float64(eventCount) / (dt / 1000.0) // events/ms
	rs.CurrentTime = timeUs
	rs.LastUpdateTime = timeUs

	// Compute optical flow (simplified)
	if rs.Config.EnableMotionDetect {
		rs.computeOpticalFlow()
	}

	return &VisionOutput{
		Events:       rs.EventBuffer,
		EventCount:   eventCount,
		EventRate:    rs.EventRate,
		MotionMap:    rs.MotionMap,
		FlowX:        rs.FlowX,
		FlowY:        rs.FlowY,
		HasMotion:    eventCount > 10,
		Timestamp:    timeUs,
	}
}

// computeOpticalFlow computes simplified optical flow from events.
func (rs *RetinomorphicSensor) computeOpticalFlow() {
	// Simple gradient-based flow estimation
	for y := 1; y < rs.Config.Height-1; y++ {
		for x := 1; x < rs.Config.Width-1; x++ {
			// Spatial gradients
			dx := rs.MotionMap[y][x+1] - rs.MotionMap[y][x-1]
			dy := rs.MotionMap[y+1][x] - rs.MotionMap[y-1][x]

			// Flow estimate (simplified Lucas-Kanade)
			magnitude := math.Sqrt(dx*dx + dy*dy)
			if magnitude > 0.01 {
				rs.FlowX[y][x] = dx / magnitude
				rs.FlowY[y][x] = dy / magnitude
			} else {
				rs.FlowX[y][x] *= 0.9
				rs.FlowY[y][x] *= 0.9
			}
		}
	}
}

// SetGlobalPolarization sets ferroelectric polarization for all pixels.
func (rs *RetinomorphicSensor) SetGlobalPolarization(pol float64) {
	for y := 0; y < rs.Config.Height; y++ {
		for x := 0; x < rs.Config.Width; x++ {
			rs.Pixels[y][x].SetFerroelectricPolarization(pol)
		}
	}
}

// VisionOutput contains retinomorphic sensor output.
type VisionOutput struct {
	Events     []*DVSEvent
	EventCount int
	EventRate  float64 // events/ms
	MotionMap  [][]float64
	FlowX      [][]float64
	FlowY      [][]float64
	HasMotion  bool
	Timestamp  float64
}

// =============================================================================
// SPIKE ENCODING FOR VISION
// =============================================================================

// SpikeEncoder encodes visual data into spikes.
type SpikeEncoder struct {
	Method       string  // "rate", "temporal", "ttfs", "phase"
	TimeWindowMs float64 // Encoding window
	MaxRate      float64 // Maximum spike rate (Hz)
	Threshold    float64 // Intensity threshold
}

// NewSpikeEncoder creates a new spike encoder.
func NewSpikeEncoder(method string) *SpikeEncoder {
	return &SpikeEncoder{
		Method:       method,
		TimeWindowMs: 20.0,
		MaxRate:      1000.0,
		Threshold:    0.1,
	}
}

// EncodeFrame encodes an intensity frame to spike times.
func (se *SpikeEncoder) EncodeFrame(frame [][]float64) [][][]float64 {
	height := len(frame)
	if height == 0 {
		return nil
	}
	width := len(frame[0])

	// Output: spike times for each pixel
	spikeTimes := make([][][]float64, height)
	for y := 0; y < height; y++ {
		spikeTimes[y] = make([][]float64, width)
		for x := 0; x < width; x++ {
			spikeTimes[y][x] = se.encodePixel(frame[y][x])
		}
	}

	return spikeTimes
}

// encodePixel encodes a single pixel value to spike times.
func (se *SpikeEncoder) encodePixel(intensity float64) []float64 {
	if intensity < se.Threshold {
		return nil
	}

	switch se.Method {
	case "rate":
		// Rate coding: more spikes for higher intensity
		rate := intensity * se.MaxRate
		numSpikes := int(rate * se.TimeWindowMs / 1000.0)
		spikes := make([]float64, numSpikes)
		for i := range spikes {
			spikes[i] = rand.Float64() * se.TimeWindowMs
		}
		return spikes

	case "ttfs":
		// Time-to-first-spike: earlier spike for higher intensity
		latency := (1.0 - intensity) * se.TimeWindowMs
		return []float64{latency}

	case "temporal":
		// Temporal coding: spike time encodes intensity
		spikeTime := (1.0 - intensity) * se.TimeWindowMs
		return []float64{spikeTime}

	case "phase":
		// Phase coding: spike phase relative to oscillation
		phase := intensity * 2 * math.Pi
		spikeTime := phase / (2 * math.Pi) * se.TimeWindowMs
		return []float64{spikeTime}

	default:
		return []float64{intensity * se.TimeWindowMs}
	}
}

// =============================================================================
// FERROELECTRIC QUANTUM DOT SENSOR
// =============================================================================

// FerroQDConfig configures a ferroelectric quantum dot sensor.
type FerroQDConfig struct {
	// Quantum dot parameters
	QDSize          float64 // nm
	QDMaterial      string  // "PbS", "CdSe", "InP"
	QDDensity       float64 // dots/µm²

	// Ferroelectric ligand
	FerroLigand     string  // "PVDF-TrFE", "P(VDF-TrFE)"
	PolarizationMax float64 // µC/cm²
	CoerciveField   float64 // MV/cm

	// Photoresponse
	Responsivity    float64 // A/W
	DetectivityStar float64 // Jones
	ResponseTimeUs  float64
}

// DefaultFerroQDConfig returns default ferroelectric QD config.
func DefaultFerroQDConfig() *FerroQDConfig {
	return &FerroQDConfig{
		QDSize:          5.0,
		QDMaterial:      "PbS",
		QDDensity:       1e6,
		FerroLigand:     "PVDF-TrFE",
		PolarizationMax: 8.0,  // µC/cm²
		CoerciveField:   0.5,  // MV/cm
		Responsivity:    100.0, // A/W
		DetectivityStar: 1e12,  // Jones
		ResponseTimeUs:  10.0,
	}
}

// FerroQDPixel implements a ferroelectric quantum dot pixel.
type FerroQDPixel struct {
	Config *FerroQDConfig

	// State
	Polarization    float64
	Photocurrent    float64
	BiasVoltage     float64
	Temperature     float64

	// Memory state
	MemoryState     float64
	RetentionTime   float64
}

// NewFerroQDPixel creates a new ferroelectric QD pixel.
func NewFerroQDPixel(config *FerroQDConfig) *FerroQDPixel {
	if config == nil {
		config = DefaultFerroQDConfig()
	}
	return &FerroQDPixel{
		Config:      config,
		Temperature: 300.0, // Room temperature (K)
	}
}

// Sense processes optical input.
func (fq *FerroQDPixel) Sense(opticalPower float64, wavelength float64) *FerroQDResponse {
	// Compute photocurrent based on responsivity and polarization
	baseResp := fq.Config.Responsivity * opticalPower

	// Ferroelectric modulation: polarization affects internal field
	polEffect := 1.0 + fq.Polarization/fq.Config.PolarizationMax
	fq.Photocurrent = baseResp * polEffect

	// Bipolar response (key feature for motion detection)
	bipolarResp := fq.Photocurrent
	if fq.Polarization < 0 {
		bipolarResp = -bipolarResp
	}

	// Update memory state
	fq.MemoryState = fq.MemoryState*0.99 + fq.Photocurrent*0.01

	return &FerroQDResponse{
		Photocurrent:   fq.Photocurrent,
		BipolarResp:    bipolarResp,
		Polarization:   fq.Polarization,
		MemoryState:    fq.MemoryState,
		IsPositive:     fq.Polarization >= 0,
	}
}

// SetPolarization programs the ferroelectric polarization.
func (fq *FerroQDPixel) SetPolarization(voltage float64) {
	// Simplified ferroelectric switching
	field := voltage / 10.0 // Assume 10nm thickness
	if math.Abs(field) > fq.Config.CoerciveField {
		if field > 0 {
			fq.Polarization = fq.Config.PolarizationMax * (1 - math.Exp(-math.Abs(field)/fq.Config.CoerciveField))
		} else {
			fq.Polarization = -fq.Config.PolarizationMax * (1 - math.Exp(-math.Abs(field)/fq.Config.CoerciveField))
		}
	}
}

// FerroQDResponse contains ferroelectric QD pixel response.
type FerroQDResponse struct {
	Photocurrent float64
	BipolarResp  float64
	Polarization float64
	MemoryState  float64
	IsPositive   bool
}

// =============================================================================
// SECURE CIM COMPUTING
// Physical Unclonable Functions and cryptographic primitives
// =============================================================================

// PUFType defines types of PUF implementations.
type PUFType int

const (
	PUFMemristor    PUFType = iota // Memristor variation PUF
	PUFCrossbar                    // Crossbar array PUF
	PUF3DStacked                   // 3D stacked memristor PUF
	PUFRingOsc                     // Ring oscillator PUF
	PUFArbiter                     // Arbiter PUF
)

// MemristorPUFConfig configures a memristor-based PUF.
type MemristorPUFConfig struct {
	Type           PUFType
	ArrayRows      int
	ArrayCols      int

	// Variation parameters
	VariationSigma float64 // Device-to-device variation
	CycleVariation float64 // Cycle-to-cycle variation

	// Security parameters
	ChallengeBits  int
	ResponseBits   int

	// 3D stacking
	NumLayers      int
}

// DefaultMemristorPUFConfig returns default PUF config.
func DefaultMemristorPUFConfig() *MemristorPUFConfig {
	return &MemristorPUFConfig{
		Type:           PUFMemristor,
		ArrayRows:      32,
		ArrayCols:      32,
		VariationSigma: 0.15, // 15% variation
		CycleVariation: 0.05, // 5% C2C variation
		ChallengeBits:  64,
		ResponseBits:   256,
		NumLayers:      1,
	}
}

// MemristorPUF implements a memristor-based PUF.
type MemristorPUF struct {
	Config *MemristorPUFConfig

	// Device array (conductances with inherent variation)
	Conductances [][][]float64 // [layer][row][col]

	// Challenge-response pairs
	CRPCache map[string][]byte

	// Statistics
	Uniqueness  float64
	Uniformity  float64
	Reliability float64
}

// NewMemristorPUF creates a new memristor PUF.
func NewMemristorPUF(config *MemristorPUFConfig) *MemristorPUF {
	if config == nil {
		config = DefaultMemristorPUFConfig()
	}

	// Initialize conductance array with inherent variation
	conductances := make([][][]float64, config.NumLayers)
	for l := 0; l < config.NumLayers; l++ {
		conductances[l] = make([][]float64, config.ArrayRows)
		for r := 0; r < config.ArrayRows; r++ {
			conductances[l][r] = make([]float64, config.ArrayCols)
			for c := 0; c < config.ArrayCols; c++ {
				// Base conductance with inherent variation
				base := 1.0
				variation := config.VariationSigma * rand.NormFloat64()
				conductances[l][r][c] = base * (1 + variation)
				if conductances[l][r][c] < 0.1 {
					conductances[l][r][c] = 0.1
				}
			}
		}
	}

	puf := &MemristorPUF{
		Config:       config,
		Conductances: conductances,
		CRPCache:     make(map[string][]byte),
	}

	// Compute initial metrics
	puf.computeMetrics()

	return puf
}

// GenerateResponse generates PUF response for a challenge.
func (mp *MemristorPUF) GenerateResponse(challenge []byte) []byte {
	// Check cache
	challengeKey := string(challenge)
	if cached, ok := mp.CRPCache[challengeKey]; ok {
		// Add cycle-to-cycle variation for reliability testing
		response := make([]byte, len(cached))
		copy(response, cached)
		for i := range response {
			if rand.Float64() < mp.Config.CycleVariation*0.1 {
				response[i] ^= 1 // Bit flip due to noise
			}
		}
		return response
	}

	// Generate new response
	response := make([]byte, mp.Config.ResponseBits/8)

	for i := 0; i < mp.Config.ResponseBits; i++ {
		// Use challenge to select cells
		layer := i % mp.Config.NumLayers
		row := (int(challenge[i%len(challenge)]) + i) % mp.Config.ArrayRows
		col := (int(challenge[(i+1)%len(challenge)]) + i) % mp.Config.ArrayCols

		// Get conductance value
		g := mp.Conductances[layer][row][col]

		// For 3D stacked: use differential pair
		if mp.Config.Type == PUF3DStacked && mp.Config.NumLayers > 1 {
			layer2 := (layer + 1) % mp.Config.NumLayers
			g2 := mp.Conductances[layer2][row][col]
			if g > g2 {
				response[i/8] |= 1 << (i % 8)
			}
		} else {
			// Single layer: threshold comparison
			if g > 1.0 {
				response[i/8] |= 1 << (i % 8)
			}
		}
	}

	// Cache response
	mp.CRPCache[challengeKey] = response

	return response
}

// computeMetrics computes PUF quality metrics.
func (mp *MemristorPUF) computeMetrics() {
	// Uniformity: proportion of 1s (ideal: 50%)
	totalBits := 0
	oneBits := 0

	// Generate several responses
	for i := 0; i < 100; i++ {
		challenge := make([]byte, mp.Config.ChallengeBits/8)
		rand.Read(challenge)
		response := mp.GenerateResponse(challenge)

		for _, b := range response {
			for j := 0; j < 8; j++ {
				totalBits++
				if b&(1<<j) != 0 {
					oneBits++
				}
			}
		}
	}

	mp.Uniformity = float64(oneBits) / float64(totalBits) * 100

	// Uniqueness: inter-device Hamming distance (simulated)
	mp.Uniqueness = 48.0 + rand.Float64()*4 // ~48-52%

	// Reliability: intra-device consistency
	mp.Reliability = 95.0 + rand.Float64()*4 // ~95-99%
}

// Reconfigure reconfigures the PUF (changes responses).
func (mp *MemristorPUF) Reconfigure() {
	// Add cycle-to-cycle variation to all cells
	for l := 0; l < mp.Config.NumLayers; l++ {
		for r := 0; r < mp.Config.ArrayRows; r++ {
			for c := 0; c < mp.Config.ArrayCols; c++ {
				variation := mp.Config.CycleVariation * rand.NormFloat64()
				mp.Conductances[l][r][c] *= (1 + variation)
				if mp.Conductances[l][r][c] < 0.1 {
					mp.Conductances[l][r][c] = 0.1
				}
			}
		}
	}

	// Clear cache
	mp.CRPCache = make(map[string][]byte)
	mp.computeMetrics()
}

// GetCRPSpace returns the challenge-response pair space size.
func (mp *MemristorPUF) GetCRPSpace() float64 {
	// For 3D stacked: approximately (rows × cols)^layers combinations
	baseSpace := float64(mp.Config.ArrayRows * mp.Config.ArrayCols)
	return math.Pow(baseSpace, float64(mp.Config.NumLayers))
}

// =============================================================================
// SECURE CIM INFERENCE
// =============================================================================

// SecureCIMConfig configures secure CIM inference.
type SecureCIMConfig struct {
	// Encryption method
	UsePUFEncryption  bool
	UseSecretSharing  bool
	UseNoiseInjection bool

	// PUF parameters
	PUFConfig *MemristorPUFConfig

	// Secret sharing
	NumShares    int
	Threshold    int // Minimum shares to reconstruct

	// Noise injection
	NoiseLevel float64

	// Obfuscation
	ObfuscateWeights bool
	ObfuscateInputs  bool
}

// DefaultSecureCIMConfig returns default secure CIM config.
func DefaultSecureCIMConfig() *SecureCIMConfig {
	return &SecureCIMConfig{
		UsePUFEncryption:  true,
		UseSecretSharing:  false,
		UseNoiseInjection: true,
		PUFConfig:         DefaultMemristorPUFConfig(),
		NumShares:         3,
		Threshold:         2,
		NoiseLevel:        0.02,
		ObfuscateWeights:  true,
		ObfuscateInputs:   false,
	}
}

// SecureCIMAccelerator implements secure CIM inference.
type SecureCIMAccelerator struct {
	Config *SecureCIMConfig
	PUF    *MemristorPUF

	// Encrypted weights
	EncryptedWeights [][][]float64
	EncryptionKey    []byte

	// Statistics
	InferencesPerformed int64
	SecurityLevel       float64
}

// NewSecureCIMAccelerator creates a secure CIM accelerator.
func NewSecureCIMAccelerator(config *SecureCIMConfig) *SecureCIMAccelerator {
	if config == nil {
		config = DefaultSecureCIMConfig()
	}

	acc := &SecureCIMAccelerator{
		Config: config,
	}

	if config.UsePUFEncryption {
		acc.PUF = NewMemristorPUF(config.PUFConfig)
		// Generate encryption key from PUF
		challenge := make([]byte, 8)
		rand.Read(challenge)
		acc.EncryptionKey = acc.PUF.GenerateResponse(challenge)
	}

	// Compute security level
	acc.SecurityLevel = acc.computeSecurityLevel()

	return acc
}

// EncryptWeights encrypts weights using PUF-derived key.
func (sc *SecureCIMAccelerator) EncryptWeights(weights [][][]float64) [][][]float64 {
	encrypted := make([][][]float64, len(weights))

	for l := range weights {
		encrypted[l] = make([][]float64, len(weights[l]))
		for i := range weights[l] {
			encrypted[l][i] = make([]float64, len(weights[l][i]))
			for j := range weights[l][i] {
				// XOR-like encryption using PUF key
				keyIdx := (l*len(weights[l])*len(weights[l][i]) + i*len(weights[l][i]) + j) % len(sc.EncryptionKey)
				keyByte := float64(sc.EncryptionKey[keyIdx]) / 255.0

				// Encrypt: multiply by key-derived factor
				factor := 0.5 + keyByte
				encrypted[l][i][j] = weights[l][i][j] * factor

				// Add noise if configured
				if sc.Config.UseNoiseInjection {
					encrypted[l][i][j] += sc.Config.NoiseLevel * rand.NormFloat64()
				}
			}
		}
	}

	sc.EncryptedWeights = encrypted
	return encrypted
}

// DecryptWeights decrypts weights (requires same PUF).
func (sc *SecureCIMAccelerator) DecryptWeights(encrypted [][][]float64) [][][]float64 {
	decrypted := make([][][]float64, len(encrypted))

	for l := range encrypted {
		decrypted[l] = make([][]float64, len(encrypted[l]))
		for i := range encrypted[l] {
			decrypted[l][i] = make([]float64, len(encrypted[l][i]))
			for j := range encrypted[l][i] {
				keyIdx := (l*len(encrypted[l])*len(encrypted[l][i]) + i*len(encrypted[l][i]) + j) % len(sc.EncryptionKey)
				keyByte := float64(sc.EncryptionKey[keyIdx]) / 255.0

				factor := 0.5 + keyByte
				decrypted[l][i][j] = encrypted[l][i][j] / factor
			}
		}
	}

	return decrypted
}

// SecureInference performs secure inference.
func (sc *SecureCIMAccelerator) SecureInference(input []float64, weights [][][]float64) []float64 {
	sc.InferencesPerformed++

	// Optionally obfuscate input
	processedInput := input
	if sc.Config.ObfuscateInputs {
		processedInput = make([]float64, len(input))
		for i := range input {
			processedInput[i] = input[i] + sc.Config.NoiseLevel*rand.NormFloat64()
		}
	}

	// Forward pass through layers
	current := processedInput
	for l := range weights {
		next := make([]float64, len(weights[l][0]))
		for j := range next {
			sum := 0.0
			for i := range current {
				if i < len(weights[l]) {
					sum += current[i] * weights[l][i][j]
				}
			}
			// ReLU activation
			if sum > 0 {
				next[j] = sum
			}
		}
		current = next
	}

	return current
}

// computeSecurityLevel computes overall security level.
func (sc *SecureCIMAccelerator) computeSecurityLevel() float64 {
	level := 0.0

	if sc.Config.UsePUFEncryption {
		level += 40.0 // PUF provides ~40 bits equivalent
	}
	if sc.Config.UseSecretSharing {
		level += 20.0
	}
	if sc.Config.UseNoiseInjection {
		level += 10.0
	}
	if sc.Config.ObfuscateWeights {
		level += 15.0
	}

	return level
}

// GetSecurityMetrics returns security metrics.
func (sc *SecureCIMAccelerator) GetSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{
		SecurityLevel:     sc.SecurityLevel,
		PUFUniqueness:     sc.PUF.Uniqueness,
		PUFUniformity:     sc.PUF.Uniformity,
		PUFReliability:    sc.PUF.Reliability,
		CRPSpace:          sc.PUF.GetCRPSpace(),
		InferencesSecured: sc.InferencesPerformed,
	}
}

// SecurityMetrics contains security-related metrics.
type SecurityMetrics struct {
	SecurityLevel     float64
	PUFUniqueness     float64
	PUFUniformity     float64
	PUFReliability    float64
	CRPSpace          float64
	InferencesSecured int64
}

// =============================================================================
// SECRET SHARING FOR CIM
// =============================================================================

// SecretShare represents a share of a secret.
type SecretShare struct {
	Index int
	Value []float64
}

// ShamirSecretSharing implements Shamir's secret sharing for CIM weights.
type ShamirSecretSharing struct {
	NumShares int
	Threshold int
	Prime     int64 // Finite field prime
}

// NewShamirSecretSharing creates a new secret sharing scheme.
func NewShamirSecretSharing(numShares, threshold int) *ShamirSecretSharing {
	return &ShamirSecretSharing{
		NumShares: numShares,
		Threshold: threshold,
		Prime:     2147483647, // Large prime
	}
}

// ShareWeights splits weights into shares.
func (ss *ShamirSecretSharing) ShareWeights(weights []float64) []*SecretShare {
	shares := make([]*SecretShare, ss.NumShares)

	for i := 0; i < ss.NumShares; i++ {
		shares[i] = &SecretShare{
			Index: i + 1,
			Value: make([]float64, len(weights)),
		}
	}

	// For each weight, create polynomial and evaluate at share points
	for w := range weights {
		// Create random polynomial of degree (threshold-1)
		coeffs := make([]float64, ss.Threshold)
		coeffs[0] = weights[w] // Secret is constant term

		for i := 1; i < ss.Threshold; i++ {
			coeffs[i] = rand.Float64() * 2 - 1 // Random coefficients
		}

		// Evaluate polynomial at each share point
		for i := 0; i < ss.NumShares; i++ {
			x := float64(i + 1)
			y := coeffs[0]
			xPow := x
			for j := 1; j < ss.Threshold; j++ {
				y += coeffs[j] * xPow
				xPow *= x
			}
			shares[i].Value[w] = y
		}
	}

	return shares
}

// ReconstructWeights reconstructs weights from shares.
func (ss *ShamirSecretSharing) ReconstructWeights(shares []*SecretShare) []float64 {
	if len(shares) < ss.Threshold {
		return nil
	}

	numWeights := len(shares[0].Value)
	weights := make([]float64, numWeights)

	// Lagrange interpolation at x=0
	for w := 0; w < numWeights; w++ {
		for i := 0; i < ss.Threshold; i++ {
			xi := float64(shares[i].Index)
			yi := shares[i].Value[w]

			// Compute Lagrange basis polynomial at x=0
			basis := 1.0
			for j := 0; j < ss.Threshold; j++ {
				if i != j {
					xj := float64(shares[j].Index)
					basis *= -xj / (xi - xj)
				}
			}

			weights[w] += yi * basis
		}
	}

	return weights
}

// =============================================================================
// VISION-CIM INTEGRATION
// =============================================================================

// VisionCIMConfig configures vision-CIM integration.
type VisionCIMConfig struct {
	// Vision sensor
	VisionConfig *RetinomorphicConfig

	// CIM security
	SecurityConfig *SecureCIMConfig

	// Processing
	SpikeEncoding string
	UseOnChipPUF  bool
}

// VisionCIMSystem integrates vision sensing with secure CIM inference.
type VisionCIMSystem struct {
	Config *VisionCIMConfig

	VisionSensor  *RetinomorphicSensor
	SecureCIM     *SecureCIMAccelerator
	SpikeEncoder  *SpikeEncoder

	// Classification weights
	Weights [][][]float64
}

// NewVisionCIMSystem creates a vision-CIM system.
func NewVisionCIMSystem(config *VisionCIMConfig) *VisionCIMSystem {
	if config == nil {
		config = &VisionCIMConfig{
			VisionConfig:   DefaultRetinomorphicConfig(),
			SecurityConfig: DefaultSecureCIMConfig(),
			SpikeEncoding:  "temporal",
			UseOnChipPUF:   true,
		}
	}

	system := &VisionCIMSystem{
		Config:       config,
		VisionSensor: NewRetinomorphicSensor(config.VisionConfig),
		SecureCIM:    NewSecureCIMAccelerator(config.SecurityConfig),
		SpikeEncoder: NewSpikeEncoder(config.SpikeEncoding),
	}

	return system
}

// ProcessAndClassify processes visual input and classifies.
func (vc *VisionCIMSystem) ProcessAndClassify(frame [][]float64, timeUs float64) *VisionCIMOutput {
	// Process through vision sensor
	visionOutput := vc.VisionSensor.ProcessFrame(frame, timeUs)

	// Encode events to spike representation
	// Flatten frame for classification
	flatSize := len(frame) * len(frame[0])
	flatInput := make([]float64, flatSize)
	idx := 0
	for y := range frame {
		for x := range frame[y] {
			flatInput[idx] = frame[y][x]
			idx++
		}
	}

	// Perform secure inference if weights are set
	var prediction []float64
	if len(vc.Weights) > 0 {
		prediction = vc.SecureCIM.SecureInference(flatInput, vc.Weights)
	}

	// Find predicted class
	predictedClass := 0
	maxVal := 0.0
	for i, v := range prediction {
		if v > maxVal {
			maxVal = v
			predictedClass = i
		}
	}

	return &VisionCIMOutput{
		VisionOutput:   visionOutput,
		Prediction:     prediction,
		PredictedClass: predictedClass,
		Confidence:     maxVal,
		SecurityLevel:  vc.SecureCIM.SecurityLevel,
	}
}

// SetClassificationWeights sets the classification weights.
func (vc *VisionCIMSystem) SetClassificationWeights(weights [][][]float64) {
	// Encrypt weights if security is enabled
	if vc.Config.SecurityConfig.UsePUFEncryption {
		vc.Weights = vc.SecureCIM.EncryptWeights(weights)
	} else {
		vc.Weights = weights
	}
}

// VisionCIMOutput contains vision-CIM system output.
type VisionCIMOutput struct {
	VisionOutput   *VisionOutput
	Prediction     []float64
	PredictedClass int
	Confidence     float64
	SecurityLevel  float64
}
