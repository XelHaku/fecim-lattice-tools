// writeverify_a2s.go - Write Verification Circuits and Analog-to-Spike Conversion
// for Ferroelectric Compute-in-Memory (CIM) Systems
//
// This module implements:
// 1. Write verification schemes for FeRAM/FeFET memory
// 2. Sense amplifier designs (DCRS, offset-canceled)
// 3. Analog-to-spike encoding schemes (rate, temporal, burst, TTFS)
// 4. Sigma-delta neuron circuits for efficient conversion
// 5. Integrated CIM write-verify and spike encoding systems
//
// Based on research:
// - Differential Capacitance Read Scheme (DCRS) for FeRAM
// - Offset-Canceled Sense Amplifier design
// - Self-Terminating Write (STW) for MLC ReRAM
// - Sigma-delta neuron encoding

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// WRITE VERIFICATION CIRCUIT CONFIGURATION
// =============================================================================

// WriteVerifyConfig configures write verification circuits
type WriteVerifyConfig struct {
	// Memory technology
	MemoryType       string  // "FeRAM", "FeFET", "RRAM", "PCM"
	CellStructure    string  // "1T1C", "2T2C", "3TnC", "1T"
	NumLevels        int     // MLC levels (2, 4, 8, 16, etc.)

	// Write parameters
	MaxWritePulses   int     // Maximum write-verify iterations
	WritePulseWidth  float64 // Base pulse width in ns
	WriteVoltageStep float64 // Voltage increment per iteration in V
	WriteVoltageMax  float64 // Maximum write voltage in V

	// Verify parameters
	VerifyMargin     float64 // Required margin from threshold (V or Ω)
	VerifyDelay      float64 // Delay between write and verify in ns

	// Sense amplifier
	SenseAmpType     string  // "DCRS", "CurrentSense", "OffsetCanceled"
	OffsetVoltage    float64 // Input offset voltage in mV
	SenseMargin      float64 // Sensing margin in mV

	// Noise and variation
	ProcessVariation float64 // σ of process variation
	NoiseLevel       float64 // Thermal noise level

	// Power optimization
	LowPowerMode     bool    // Enable low-power techniques
	AdaptivePulsing  bool    // Adaptive pulse width/voltage
}

// DefaultWriteVerifyConfig returns default configuration
func DefaultWriteVerifyConfig() *WriteVerifyConfig {
	return &WriteVerifyConfig{
		MemoryType:       "FeFET",
		CellStructure:    "1T",
		NumLevels:        4,
		MaxWritePulses:   10,
		WritePulseWidth:  50.0,  // 50 ns
		WriteVoltageStep: 0.1,   // 100 mV steps
		WriteVoltageMax:  4.0,   // 4V max
		VerifyMargin:     50.0,  // 50 mV margin
		VerifyDelay:      10.0,  // 10 ns
		SenseAmpType:     "OffsetCanceled",
		OffsetVoltage:    20.0,  // 20 mV offset
		SenseMargin:      100.0, // 100 mV
		ProcessVariation: 0.05,
		NoiseLevel:       0.02,
		LowPowerMode:     true,
		AdaptivePulsing:  true,
	}
}

// =============================================================================
// SENSE AMPLIFIER MODELS
// =============================================================================

// SenseAmplifier represents a sense amplifier circuit
type SenseAmplifier struct {
	Config        *WriteVerifyConfig
	Type          string
	OffsetVoltage float64
	Gain          float64
	Bandwidth     float64 // MHz
	Power         float64 // μW
}

// DifferentialCapacitanceReadScheme implements DCRS sense amplifier
type DifferentialCapacitanceReadScheme struct {
	SenseAmplifier

	// DCRS-specific parameters
	BitlineCapacitance float64 // fF
	FerroCapacitance   float64 // fF
	PrechargeVoltage   float64 // V

	// Timing
	WordlineActivation float64 // ns
	PlatelineDelay     float64 // ns - plateline activated after sense
	AccessTime         float64 // ns
}

// NewDCRS creates a new DCRS sense amplifier
func NewDCRS(config *WriteVerifyConfig) *DifferentialCapacitanceReadScheme {
	dcrs := &DifferentialCapacitanceReadScheme{
		SenseAmplifier: SenseAmplifier{
			Config:        config,
			Type:          "DCRS",
			OffsetVoltage: config.OffsetVoltage,
			Gain:          100.0,
			Bandwidth:     500.0, // 500 MHz
			Power:         50.0,  // 50 μW
		},
		BitlineCapacitance: 100.0, // 100 fF
		FerroCapacitance:   50.0,  // 50 fF
		PrechargeVoltage:   1.5,   // 1.5V
		WordlineActivation: 5.0,   // 5 ns
		PlatelineDelay:     40.0,  // 40 ns - after sensing complete
		AccessTime:         45.0,  // 45 ns total
	}
	return dcrs
}

// Sense performs capacitance-based sensing
func (dcrs *DifferentialCapacitanceReadScheme) Sense(cellCapacitance, referenceCapacitance float64) (bool, float64) {
	// DCRS advantage: sense immediately after wordline activation
	// before plateline rise, using capacitance difference

	capDifference := cellCapacitance - referenceCapacitance

	// Convert capacitance difference to voltage
	// V = Q/C, where Q is charge sharing
	voltageSignal := dcrs.PrechargeVoltage * capDifference / (dcrs.BitlineCapacitance + dcrs.FerroCapacitance)

	// Add noise and offset
	noise := rand.NormFloat64() * dcrs.Config.NoiseLevel * voltageSignal
	offset := dcrs.OffsetVoltage / 1000.0 // Convert mV to V

	sensedVoltage := voltageSignal + noise + offset

	// Compare against threshold
	threshold := dcrs.Config.SenseMargin / 1000.0

	return sensedVoltage > threshold, sensedVoltage
}

// OffsetCanceledSenseAmplifier implements offset cancellation
type OffsetCanceledSenseAmplifier struct {
	SenseAmplifier

	// Offset cancellation parameters
	CalibrationPhases int     // Number of calibration cycles
	ResidualOffset    float64 // Remaining offset after cancellation (mV)
	AutoZeroEnable    bool    // Auto-zero technique
	ChopperEnable     bool    // Chopper stabilization
}

// NewOffsetCanceledSA creates an offset-canceled sense amplifier
func NewOffsetCanceledSA(config *WriteVerifyConfig) *OffsetCanceledSenseAmplifier {
	return &OffsetCanceledSenseAmplifier{
		SenseAmplifier: SenseAmplifier{
			Config:        config,
			Type:          "OffsetCanceled",
			OffsetVoltage: config.OffsetVoltage,
			Gain:          200.0,
			Bandwidth:     300.0,
			Power:         30.0,
		},
		CalibrationPhases: 2,
		ResidualOffset:    config.OffsetVoltage * 0.1, // 90% reduction
		AutoZeroEnable:    true,
		ChopperEnable:     false,
	}
}

// CalibratedSense performs offset-canceled sensing
func (ocsa *OffsetCanceledSenseAmplifier) CalibratedSense(signalVoltage, referenceVoltage float64) (int, float64) {
	// Phase 1: Sample offset during calibration
	offsetSample := ocsa.SenseAmplifier.OffsetVoltage / 1000.0

	// Phase 2: Apply auto-zero correction
	if ocsa.AutoZeroEnable {
		offsetSample = ocsa.ResidualOffset / 1000.0
	}

	// Amplify difference with residual offset
	difference := signalVoltage - referenceVoltage
	amplified := (difference - offsetSample) * ocsa.Gain

	// Add noise
	noise := rand.NormFloat64() * ocsa.Config.NoiseLevel * 0.1
	result := amplified + noise

	// Quantize to levels
	levelSpacing := 1.0 / float64(ocsa.Config.NumLevels)
	level := int(result / levelSpacing)
	if level < 0 {
		level = 0
	}
	if level >= ocsa.Config.NumLevels {
		level = ocsa.Config.NumLevels - 1
	}

	return level, result
}

// CurrentSenseAmplifier implements current-based sensing
type CurrentSenseAmplifier struct {
	SenseAmplifier

	// Current sensing parameters
	ReferenceCurrent float64 // μA
	SenseResistor    float64 // kΩ
	AsymmetricRatio  float64 // Asymmetric transistor ratio
}

// NewCurrentSenseAmplifier creates a current-based sense amplifier
func NewCurrentSenseAmplifier(config *WriteVerifyConfig) *CurrentSenseAmplifier {
	return &CurrentSenseAmplifier{
		SenseAmplifier: SenseAmplifier{
			Config:        config,
			Type:          "CurrentSense",
			OffsetVoltage: config.OffsetVoltage,
			Gain:          50.0,
			Bandwidth:     400.0,
			Power:         25.0,
		},
		ReferenceCurrent: 10.0,  // 10 μA
		SenseResistor:    10.0,  // 10 kΩ
		AsymmetricRatio:  1.5,   // 53.9% margin improvement
	}
}

// SenseCurrent performs current-based sensing with asymmetric design
func (csa *CurrentSenseAmplifier) SenseCurrent(cellCurrent float64) (bool, float64) {
	// Asymmetric design: different transistor sizes improve margin
	effectiveReference := csa.ReferenceCurrent / csa.AsymmetricRatio

	// Current difference
	currentDiff := cellCurrent - effectiveReference

	// Convert to voltage through sense resistor
	voltage := currentDiff * csa.SenseResistor

	// Add noise
	noise := rand.NormFloat64() * csa.Config.NoiseLevel * voltage

	result := voltage + noise

	// 53.9% improved margin with asymmetric design
	margin := csa.Config.SenseMargin / 1000.0 * 1.539

	return result > margin, result
}

// =============================================================================
// WRITE-VERIFY SCHEMES
// =============================================================================

// WriteVerifyResult contains result of write-verify operation
type WriteVerifyResult struct {
	TargetLevel    int
	AchievedLevel  int
	NumPulses      int
	TotalTime      float64 // ns
	TotalEnergy    float64 // pJ
	Success        bool
	FinalVoltage   float64
	FinalCurrent   float64
}

// IterativeWriteVerify implements standard write-verify scheme
type IterativeWriteVerify struct {
	Config      *WriteVerifyConfig
	SenseAmp    *OffsetCanceledSenseAmplifier

	// Statistics
	TotalWrites       int
	SuccessfulWrites  int
	AveragePulses     float64
	AverageEnergy     float64
}

// NewIterativeWriteVerify creates an iterative write-verify controller
func NewIterativeWriteVerify(config *WriteVerifyConfig) *IterativeWriteVerify {
	return &IterativeWriteVerify{
		Config:   config,
		SenseAmp: NewOffsetCanceledSA(config),
	}
}

// WriteCell performs iterative write-verify on a single cell
func (iwv *IterativeWriteVerify) WriteCell(currentLevel, targetLevel int) *WriteVerifyResult {
	result := &WriteVerifyResult{
		TargetLevel:   targetLevel,
		AchievedLevel: currentLevel,
	}

	iwv.TotalWrites++

	// Calculate target voltage range for level
	levelSpacing := 1.0 / float64(iwv.Config.NumLevels)
	targetVoltageMin := float64(targetLevel) * levelSpacing
	targetVoltageMax := float64(targetLevel+1) * levelSpacing
	targetVoltage := (targetVoltageMin + targetVoltageMax) / 2.0

	// Current state voltage
	currentVoltage := float64(currentLevel) * levelSpacing + rand.NormFloat64()*iwv.Config.ProcessVariation*levelSpacing

	// Iterative write-verify
	writeVoltage := iwv.Config.WriteVoltageStep
	pulseWidth := iwv.Config.WritePulseWidth

	for pulse := 0; pulse < iwv.Config.MaxWritePulses; pulse++ {
		result.NumPulses++

		// Apply write pulse
		direction := 1.0
		if targetVoltage < currentVoltage {
			direction = -1.0
		}

		// Voltage change depends on pulse amplitude and width
		deltaV := direction * writeVoltage * (pulseWidth / 100.0) * 0.1
		currentVoltage += deltaV

		// Add variation
		currentVoltage += rand.NormFloat64() * iwv.Config.ProcessVariation * 0.1

		// Calculate time and energy for this pulse
		result.TotalTime += pulseWidth + iwv.Config.VerifyDelay
		result.TotalEnergy += writeVoltage * 10.0 * (pulseWidth / 1000.0) // pJ estimate

		// Verify
		verifiedLevel, _ := iwv.SenseAmp.CalibratedSense(currentVoltage, 0.5)
		result.AchievedLevel = verifiedLevel

		// Check if within margin
		marginV := iwv.Config.VerifyMargin / 1000.0
		if math.Abs(currentVoltage-targetVoltage) < marginV {
			result.Success = true
			result.FinalVoltage = currentVoltage
			break
		}

		// Adaptive pulsing: increase voltage if not converging
		if iwv.Config.AdaptivePulsing && pulse > 3 {
			writeVoltage = math.Min(writeVoltage+iwv.Config.WriteVoltageStep, iwv.Config.WriteVoltageMax)
		}
	}

	result.FinalVoltage = currentVoltage

	if result.Success {
		iwv.SuccessfulWrites++
	}

	// Update statistics
	iwv.AveragePulses = (iwv.AveragePulses*float64(iwv.TotalWrites-1) + float64(result.NumPulses)) / float64(iwv.TotalWrites)
	iwv.AverageEnergy = (iwv.AverageEnergy*float64(iwv.TotalWrites-1) + result.TotalEnergy) / float64(iwv.TotalWrites)

	return result
}

// SelfTerminatingWrite implements STW for reduced latency
type SelfTerminatingWrite struct {
	Config         *WriteVerifyConfig

	// STW-specific parameters
	TargetConductance float64
	ToleranceBand     float64
	ReuseADC          bool // Reuse in-memory computing ADC
	ReuseTIA          bool // Reuse Trans-Impedance Amplifier
}

// NewSelfTerminatingWrite creates an STW controller
func NewSelfTerminatingWrite(config *WriteVerifyConfig) *SelfTerminatingWrite {
	return &SelfTerminatingWrite{
		Config:            config,
		TargetConductance: 0.0,
		ToleranceBand:     0.05, // 5% tolerance
		ReuseADC:          true,
		ReuseTIA:          true,
	}
}

// WriteWithSTW performs self-terminating write
func (stw *SelfTerminatingWrite) WriteWithSTW(currentConductance, targetConductance float64) *WriteVerifyResult {
	result := &WriteVerifyResult{
		TotalTime:   0,
		TotalEnergy: 0,
	}

	stw.TargetConductance = targetConductance
	conductance := currentConductance

	// Calculate target level
	levelSpacing := 1.0 / float64(stw.Config.NumLevels)
	result.TargetLevel = int(targetConductance / levelSpacing)
	result.AchievedLevel = int(currentConductance / levelSpacing)

	// STW: apply write pulse with built-in termination
	writeVoltage := stw.Config.WriteVoltageStep * 2.0 // Initial higher voltage

	for pulse := 0; pulse < stw.Config.MaxWritePulses; pulse++ {
		result.NumPulses++

		// Direction
		direction := 1.0
		if conductance > targetConductance {
			direction = -1.0
		}

		// Apply pulse - STW monitors in real-time
		pulseDuration := 0.0
		pulseStep := stw.Config.WritePulseWidth / 10.0 // Sub-divide pulse

		for subStep := 0; subStep < 10; subStep++ {
			pulseDuration += pulseStep

			// Conductance change
			deltaG := direction * writeVoltage * (pulseStep / 100.0) * 0.05
			conductance += deltaG

			// Real-time monitoring using reused ADC/TIA
			if stw.ReuseADC {
				// Check if target reached
				error := math.Abs(conductance - targetConductance)
				if error < stw.ToleranceBand*targetConductance {
					// Self-terminate
					result.Success = true
					break
				}
			}
		}

		result.TotalTime += pulseDuration + 5.0 // 5ns overhead
		result.TotalEnergy += writeVoltage * 5.0 * (pulseDuration / 1000.0)

		if result.Success {
			break
		}

		// Reduce voltage as we get closer
		error := math.Abs(conductance - targetConductance)
		if error < stw.ToleranceBand*targetConductance*3.0 {
			writeVoltage *= 0.8
		}
	}

	result.FinalCurrent = conductance * 0.5 // I = G * V
	result.AchievedLevel = int(conductance / levelSpacing)

	return result
}

// =============================================================================
// ANALOG-TO-SPIKE CONVERSION CONFIGURATION
// =============================================================================

// SpikeEncodingConfig configures spike encoding
type SpikeEncodingConfig struct {
	// Encoding type
	EncodingType string // "Rate", "Temporal", "Burst", "TTFS", "Phase", "SigmaDelta"

	// Timing parameters
	TimeWindow   float64 // Encoding window in ms
	TimeStep     float64 // Simulation time step in ms
	MaxSpikes    int     // Maximum spikes in window

	// Threshold parameters
	Threshold    float64 // Firing threshold
	ResetValue   float64 // Reset value after spike
	RefractoryPeriod float64 // Refractory period in ms

	// Leaky integrate-and-fire parameters
	LeakRate     float64 // Leak time constant
	Integration  float64 // Integration gain

	// Noise parameters
	NoiseLevel   float64 // Input noise level
	JitterLevel  float64 // Spike timing jitter (ms)
}

// DefaultSpikeEncodingConfig returns default configuration
func DefaultSpikeEncodingConfig() *SpikeEncodingConfig {
	return &SpikeEncodingConfig{
		EncodingType:     "Rate",
		TimeWindow:       10.0,   // 10 ms
		TimeStep:         0.1,    // 0.1 ms
		MaxSpikes:        100,
		Threshold:        1.0,
		ResetValue:       0.0,
		RefractoryPeriod: 1.0,    // 1 ms
		LeakRate:         0.1,
		Integration:      1.0,
		NoiseLevel:       0.02,
		JitterLevel:      0.1,    // 0.1 ms jitter
	}
}

// =============================================================================
// SPIKE ENCODING SCHEMES
// =============================================================================

// SpikeTrain represents a spike train
type SpikeTrain struct {
	SpikeTimes []float64 // Spike times in ms
	NumSpikes  int
	Rate       float64   // Average firing rate (Hz)
	ISIs       []float64 // Inter-spike intervals
}

// SpikeEncoder interface for different encoding schemes
type SpikeEncoder interface {
	Encode(analogValue float64) *SpikeTrain
	Decode(train *SpikeTrain) float64
	GetName() string
	GetEfficiency() float64 // Spikes per unit information
}

// RateEncoder implements rate coding
type RateEncoder struct {
	Config  *SpikeEncodingConfig
	MinRate float64 // Minimum firing rate (Hz)
	MaxRate float64 // Maximum firing rate (Hz)
}

// NewRateEncoder creates a rate encoder
func NewRateEncoder(config *SpikeEncodingConfig) *RateEncoder {
	return &RateEncoder{
		Config:  config,
		MinRate: 10.0,   // 10 Hz minimum
		MaxRate: 200.0,  // 200 Hz maximum
	}
}

// Encode converts analog value to rate-coded spike train
func (re *RateEncoder) Encode(analogValue float64) *SpikeTrain {
	// Clamp input
	if analogValue < 0 {
		analogValue = 0
	}
	if analogValue > 1 {
		analogValue = 1
	}

	// Calculate target rate
	targetRate := re.MinRate + analogValue*(re.MaxRate-re.MinRate)

	// Generate spikes with Poisson process
	train := &SpikeTrain{
		SpikeTimes: make([]float64, 0),
	}

	avgISI := 1000.0 / targetRate // ms between spikes
	currentTime := 0.0
	lastSpikeTime := -re.Config.RefractoryPeriod

	for currentTime < re.Config.TimeWindow {
		// Poisson waiting time
		waitTime := -avgISI * math.Log(1.0-rand.Float64())
		currentTime += waitTime

		// Add noise
		currentTime += rand.NormFloat64() * re.Config.JitterLevel

		// Check refractory period
		if currentTime-lastSpikeTime > re.Config.RefractoryPeriod && currentTime < re.Config.TimeWindow {
			train.SpikeTimes = append(train.SpikeTimes, currentTime)
			lastSpikeTime = currentTime
		}
	}

	train.NumSpikes = len(train.SpikeTimes)
	train.Rate = float64(train.NumSpikes) / re.Config.TimeWindow * 1000.0 // Hz

	// Calculate ISIs
	if train.NumSpikes > 1 {
		train.ISIs = make([]float64, train.NumSpikes-1)
		for i := 1; i < train.NumSpikes; i++ {
			train.ISIs[i-1] = train.SpikeTimes[i] - train.SpikeTimes[i-1]
		}
	}

	return train
}

// Decode converts rate-coded spike train back to analog value
func (re *RateEncoder) Decode(train *SpikeTrain) float64 {
	if train.NumSpikes == 0 {
		return 0
	}

	rate := train.Rate
	if rate < re.MinRate {
		rate = re.MinRate
	}
	if rate > re.MaxRate {
		rate = re.MaxRate
	}

	return (rate - re.MinRate) / (re.MaxRate - re.MinRate)
}

// GetName returns encoder name
func (re *RateEncoder) GetName() string {
	return "RateEncoder"
}

// GetEfficiency returns spikes per information unit
func (re *RateEncoder) GetEfficiency() float64 {
	// Rate coding is less efficient - needs many spikes
	return 0.3 // Low efficiency
}

// TemporalEncoder implements temporal/latency coding
type TemporalEncoder struct {
	Config     *SpikeEncodingConfig
	MaxLatency float64 // Maximum latency for minimum input
	MinLatency float64 // Minimum latency for maximum input
}

// NewTemporalEncoder creates a temporal encoder
func NewTemporalEncoder(config *SpikeEncodingConfig) *TemporalEncoder {
	return &TemporalEncoder{
		Config:     config,
		MaxLatency: config.TimeWindow * 0.9, // 90% of window
		MinLatency: 0.5, // 0.5 ms minimum latency
	}
}

// Encode converts analog value to temporal spike
func (te *TemporalEncoder) Encode(analogValue float64) *SpikeTrain {
	train := &SpikeTrain{
		SpikeTimes: make([]float64, 0, 1),
	}

	// Clamp input
	if analogValue < 0 {
		analogValue = 0
	}
	if analogValue > 1 {
		analogValue = 1
	}

	// Higher value = earlier spike (inverse relationship)
	latency := te.MaxLatency - analogValue*(te.MaxLatency-te.MinLatency)

	// Add jitter
	latency += rand.NormFloat64() * te.Config.JitterLevel

	if latency < te.MinLatency {
		latency = te.MinLatency
	}
	if latency < te.Config.TimeWindow {
		train.SpikeTimes = append(train.SpikeTimes, latency)
		train.NumSpikes = 1
	}

	train.Rate = float64(train.NumSpikes) / te.Config.TimeWindow * 1000.0

	return train
}

// Decode converts temporal spike back to analog value
func (te *TemporalEncoder) Decode(train *SpikeTrain) float64 {
	if train.NumSpikes == 0 {
		return 0
	}

	// First spike time determines value
	latency := train.SpikeTimes[0]
	if latency < te.MinLatency {
		latency = te.MinLatency
	}
	if latency > te.MaxLatency {
		latency = te.MaxLatency
	}

	return (te.MaxLatency - latency) / (te.MaxLatency - te.MinLatency)
}

// GetName returns encoder name
func (te *TemporalEncoder) GetName() string {
	return "TemporalEncoder"
}

// GetEfficiency returns spikes per information unit
func (te *TemporalEncoder) GetEfficiency() float64 {
	// Temporal coding is very efficient - single spike
	return 0.9
}

// BurstEncoder implements burst coding
type BurstEncoder struct {
	Config       *SpikeEncodingConfig
	MaxBurstSize int     // Maximum spikes in burst
	BurstISI     float64 // Inter-spike interval within burst (ms)
	BurstDelay   float64 // Delay before burst
}

// NewBurstEncoder creates a burst encoder
func NewBurstEncoder(config *SpikeEncodingConfig) *BurstEncoder {
	return &BurstEncoder{
		Config:       config,
		MaxBurstSize: 8,
		BurstISI:     0.5, // 0.5 ms within burst
		BurstDelay:   1.0, // 1 ms initial delay
	}
}

// Encode converts analog value to burst-coded spike train
func (be *BurstEncoder) Encode(analogValue float64) *SpikeTrain {
	train := &SpikeTrain{
		SpikeTimes: make([]float64, 0),
	}

	// Clamp input
	if analogValue < 0 {
		analogValue = 0
	}
	if analogValue > 1 {
		analogValue = 1
	}

	// Number of spikes in burst proportional to value
	burstSize := int(math.Round(analogValue * float64(be.MaxBurstSize)))
	if burstSize < 1 && analogValue > 0.05 {
		burstSize = 1
	}

	// Generate burst
	currentTime := be.BurstDelay
	for i := 0; i < burstSize; i++ {
		// Add jitter
		spikeTime := currentTime + rand.NormFloat64()*be.Config.JitterLevel*0.2
		if spikeTime > 0 && spikeTime < be.Config.TimeWindow {
			train.SpikeTimes = append(train.SpikeTimes, spikeTime)
		}
		currentTime += be.BurstISI
	}

	train.NumSpikes = len(train.SpikeTimes)
	train.Rate = float64(train.NumSpikes) / be.Config.TimeWindow * 1000.0

	// Calculate ISIs
	if train.NumSpikes > 1 {
		train.ISIs = make([]float64, train.NumSpikes-1)
		for i := 1; i < train.NumSpikes; i++ {
			train.ISIs[i-1] = train.SpikeTimes[i] - train.SpikeTimes[i-1]
		}
	}

	return train
}

// Decode converts burst-coded spike train back to analog value
func (be *BurstEncoder) Decode(train *SpikeTrain) float64 {
	return float64(train.NumSpikes) / float64(be.MaxBurstSize)
}

// GetName returns encoder name
func (be *BurstEncoder) GetName() string {
	return "BurstEncoder"
}

// GetEfficiency returns spikes per information unit
func (be *BurstEncoder) GetEfficiency() float64 {
	// Burst coding is moderately efficient
	return 0.7
}

// TTFSEncoder implements Time-To-First-Spike coding
type TTFSEncoder struct {
	Config       *SpikeEncodingConfig
	EncodingGain float64 // Gain for input-to-latency mapping
}

// NewTTFSEncoder creates a TTFS encoder
func NewTTFSEncoder(config *SpikeEncodingConfig) *TTFSEncoder {
	return &TTFSEncoder{
		Config:       config,
		EncodingGain: 1.0,
	}
}

// Encode converts analog value to TTFS spike
func (ttfs *TTFSEncoder) Encode(analogValue float64) *SpikeTrain {
	train := &SpikeTrain{
		SpikeTimes: make([]float64, 0, 1),
	}

	// Only generate spike if value exceeds threshold
	if analogValue < 0.01 {
		return train
	}

	// Clamp
	if analogValue > 1 {
		analogValue = 1
	}

	// TTFS: strong input = early spike
	// Using exponential mapping for better dynamic range
	latency := -math.Log(analogValue) * ttfs.Config.TimeWindow / 5.0

	// Clamp latency
	if latency < 0.1 {
		latency = 0.1
	}
	if latency > ttfs.Config.TimeWindow*0.95 {
		// Very weak signal - no spike
		return train
	}

	// Add jitter
	latency += rand.NormFloat64() * ttfs.Config.JitterLevel
	if latency < 0.1 {
		latency = 0.1
	}

	train.SpikeTimes = append(train.SpikeTimes, latency)
	train.NumSpikes = 1
	train.Rate = 1000.0 / ttfs.Config.TimeWindow

	return train
}

// Decode converts TTFS spike back to analog value
func (ttfs *TTFSEncoder) Decode(train *SpikeTrain) float64 {
	if train.NumSpikes == 0 {
		return 0
	}

	latency := train.SpikeTimes[0]
	if latency < 0.1 {
		latency = 0.1
	}

	// Inverse exponential mapping
	value := math.Exp(-latency * 5.0 / ttfs.Config.TimeWindow)
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	return value
}

// GetName returns encoder name
func (ttfs *TTFSEncoder) GetName() string {
	return "TTFSEncoder"
}

// GetEfficiency returns spikes per information unit
func (ttfs *TTFSEncoder) GetEfficiency() float64 {
	// TTFS is the most efficient - single spike with timing
	return 0.95
}

// =============================================================================
// SIGMA-DELTA NEURON
// =============================================================================

// SigmaDeltaNeuron implements sigma-delta modulation encoding
type SigmaDeltaNeuron struct {
	Config      *SpikeEncodingConfig

	// State variables
	Integrator  float64 // Sigma (integration) state
	Error       float64 // Quantization error

	// Parameters
	Quantizer   int     // Number of quantization levels
	Feedback    float64 // Feedback coefficient
	LeakFactor  float64 // Integrator leak

	// Output
	LastOutput  int     // Last quantized output
}

// NewSigmaDeltaNeuron creates a sigma-delta neuron encoder
func NewSigmaDeltaNeuron(config *SpikeEncodingConfig) *SigmaDeltaNeuron {
	return &SigmaDeltaNeuron{
		Config:     config,
		Integrator: 0,
		Error:      0,
		Quantizer:  2, // Binary for spike/no-spike
		Feedback:   1.0,
		LeakFactor: config.LeakRate,
		LastOutput: 0,
	}
}

// Reset resets neuron state
func (sdn *SigmaDeltaNeuron) Reset() {
	sdn.Integrator = 0
	sdn.Error = 0
	sdn.LastOutput = 0
}

// Step performs one time step of sigma-delta modulation
func (sdn *SigmaDeltaNeuron) Step(input float64) bool {
	// Sigma: integrate input minus feedback
	sdn.Integrator += input - sdn.Feedback*float64(sdn.LastOutput)

	// Apply leak
	sdn.Integrator *= (1.0 - sdn.LeakFactor)

	// Delta: quantize
	spike := false
	if sdn.Integrator >= sdn.Config.Threshold {
		spike = true
		sdn.LastOutput = 1
		sdn.Error = sdn.Integrator - sdn.Config.Threshold
	} else {
		sdn.LastOutput = 0
		sdn.Error = sdn.Integrator
	}

	return spike
}

// Encode converts analog signal to spike train using sigma-delta
func (sdn *SigmaDeltaNeuron) Encode(analogValue float64) *SpikeTrain {
	sdn.Reset()

	train := &SpikeTrain{
		SpikeTimes: make([]float64, 0),
	}

	// Simulate over time window
	numSteps := int(sdn.Config.TimeWindow / sdn.Config.TimeStep)

	for step := 0; step < numSteps; step++ {
		currentTime := float64(step) * sdn.Config.TimeStep

		// Add noise to input
		noisyInput := analogValue + rand.NormFloat64()*sdn.Config.NoiseLevel
		if noisyInput < 0 {
			noisyInput = 0
		}

		// Sigma-delta step
		if sdn.Step(noisyInput * sdn.Config.Integration) {
			train.SpikeTimes = append(train.SpikeTimes, currentTime)
		}
	}

	train.NumSpikes = len(train.SpikeTimes)
	train.Rate = float64(train.NumSpikes) / sdn.Config.TimeWindow * 1000.0

	// Calculate ISIs
	if train.NumSpikes > 1 {
		train.ISIs = make([]float64, train.NumSpikes-1)
		for i := 1; i < train.NumSpikes; i++ {
			train.ISIs[i-1] = train.SpikeTimes[i] - train.SpikeTimes[i-1]
		}
	}

	return train
}

// Decode reconstructs analog value from spike train
func (sdn *SigmaDeltaNeuron) Decode(train *SpikeTrain) float64 {
	// Low-pass filter the spike train
	if train.NumSpikes == 0 {
		return 0
	}

	// Simple averaging (corresponds to ideal LPF)
	numSteps := int(sdn.Config.TimeWindow / sdn.Config.TimeStep)
	return float64(train.NumSpikes) / float64(numSteps) * sdn.Config.Threshold / sdn.Config.Integration
}

// GetName returns encoder name
func (sdn *SigmaDeltaNeuron) GetName() string {
	return "SigmaDeltaNeuron"
}

// GetEfficiency returns spikes per information unit
func (sdn *SigmaDeltaNeuron) GetEfficiency() float64 {
	// Sigma-delta provides good precision with moderate spikes
	return 0.75
}

// =============================================================================
// PHASE ENCODER
// =============================================================================

// PhaseEncoder implements phase-based spike encoding
type PhaseEncoder struct {
	Config         *SpikeEncodingConfig
	ReferenceFreq  float64 // Reference oscillation frequency (Hz)
	NumPhaseBins   int     // Number of phase bins
}

// NewPhaseEncoder creates a phase encoder
func NewPhaseEncoder(config *SpikeEncodingConfig) *PhaseEncoder {
	return &PhaseEncoder{
		Config:        config,
		ReferenceFreq: 40.0, // 40 Hz gamma oscillation
		NumPhaseBins:  8,
	}
}

// Encode converts analog value to phase-coded spike
func (pe *PhaseEncoder) Encode(analogValue float64) *SpikeTrain {
	train := &SpikeTrain{
		SpikeTimes: make([]float64, 0),
	}

	// Clamp input
	if analogValue < 0 {
		analogValue = 0
	}
	if analogValue > 1 {
		analogValue = 1
	}

	// Calculate phase based on value
	// Higher value = earlier phase within cycle
	targetPhase := (1.0 - analogValue) * 2.0 * math.Pi

	// Generate spikes at target phase in each cycle
	period := 1000.0 / pe.ReferenceFreq // ms
	currentTime := 0.0

	for currentTime < pe.Config.TimeWindow {
		// Spike time within this cycle
		spikeTime := currentTime + targetPhase*period/(2.0*math.Pi)

		// Add jitter
		spikeTime += rand.NormFloat64() * pe.Config.JitterLevel

		if spikeTime > currentTime && spikeTime < pe.Config.TimeWindow {
			train.SpikeTimes = append(train.SpikeTimes, spikeTime)
		}

		currentTime += period
	}

	train.NumSpikes = len(train.SpikeTimes)
	train.Rate = float64(train.NumSpikes) / pe.Config.TimeWindow * 1000.0

	return train
}

// Decode converts phase-coded spike train back to analog value
func (pe *PhaseEncoder) Decode(train *SpikeTrain) float64 {
	if train.NumSpikes == 0 {
		return 0
	}

	// Average phase across all spikes
	period := 1000.0 / pe.ReferenceFreq
	totalPhase := 0.0

	for _, spikeTime := range train.SpikeTimes {
		// Phase within cycle
		cycleTime := math.Mod(spikeTime, period)
		phase := cycleTime / period * 2.0 * math.Pi
		totalPhase += phase
	}

	avgPhase := totalPhase / float64(train.NumSpikes)
	value := 1.0 - avgPhase/(2.0*math.Pi)

	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	return value
}

// GetName returns encoder name
func (pe *PhaseEncoder) GetName() string {
	return "PhaseEncoder"
}

// GetEfficiency returns spikes per information unit
func (pe *PhaseEncoder) GetEfficiency() float64 {
	// Phase coding is efficient but requires reference oscillation
	return 0.8
}

// =============================================================================
// MULTIPLEXED ENCODER (RATE + TEMPORAL FUSION)
// =============================================================================

// MultiplexedEncoder combines rate and temporal coding
type MultiplexedEncoder struct {
	Config      *SpikeEncodingConfig
	RateWeight  float64 // Weight for rate component
	TTFSWeight  float64 // Weight for TTFS component
	RateEncoder *RateEncoder
	TTFSEncoder *TTFSEncoder
}

// NewMultiplexedEncoder creates a multiplexed encoder
func NewMultiplexedEncoder(config *SpikeEncodingConfig) *MultiplexedEncoder {
	return &MultiplexedEncoder{
		Config:      config,
		RateWeight:  0.5,
		TTFSWeight:  0.5,
		RateEncoder: NewRateEncoder(config),
		TTFSEncoder: NewTTFSEncoder(config),
	}
}

// Encode converts analog value using multiplexed coding
func (me *MultiplexedEncoder) Encode(analogValue float64) *SpikeTrain {
	// Get both encodings
	rateTrain := me.RateEncoder.Encode(analogValue)
	ttfsTrain := me.TTFSEncoder.Encode(analogValue)

	// Merge spike trains
	combined := &SpikeTrain{
		SpikeTimes: make([]float64, 0),
	}

	// Add all spikes
	combined.SpikeTimes = append(combined.SpikeTimes, rateTrain.SpikeTimes...)
	combined.SpikeTimes = append(combined.SpikeTimes, ttfsTrain.SpikeTimes...)

	// Sort by time
	sort.Float64s(combined.SpikeTimes)

	// Remove duplicates within refractory period
	filtered := make([]float64, 0)
	lastSpike := -me.Config.RefractoryPeriod
	for _, t := range combined.SpikeTimes {
		if t-lastSpike >= me.Config.RefractoryPeriod {
			filtered = append(filtered, t)
			lastSpike = t
		}
	}
	combined.SpikeTimes = filtered

	combined.NumSpikes = len(combined.SpikeTimes)
	combined.Rate = float64(combined.NumSpikes) / me.Config.TimeWindow * 1000.0

	return combined
}

// Decode converts multiplexed spike train back to analog value
func (me *MultiplexedEncoder) Decode(train *SpikeTrain) float64 {
	// Use weighted combination of both decodings
	rateValue := me.RateEncoder.Decode(train)
	ttfsValue := me.TTFSEncoder.Decode(train)

	return me.RateWeight*rateValue + me.TTFSWeight*ttfsValue
}

// GetName returns encoder name
func (me *MultiplexedEncoder) GetName() string {
	return "MultiplexedEncoder"
}

// GetEfficiency returns spikes per information unit
func (me *MultiplexedEncoder) GetEfficiency() float64 {
	// Multiplexed achieves 6.4% higher accuracy than single schemes
	return 0.85
}

// =============================================================================
// ENCODER FACTORY
// =============================================================================

// CreateEncoder creates an encoder based on type
func CreateEncoder(config *SpikeEncodingConfig) SpikeEncoder {
	switch config.EncodingType {
	case "Rate":
		return NewRateEncoder(config)
	case "Temporal":
		return NewTemporalEncoder(config)
	case "Burst":
		return NewBurstEncoder(config)
	case "TTFS":
		return NewTTFSEncoder(config)
	case "Phase":
		return NewPhaseEncoder(config)
	case "SigmaDelta":
		return NewSigmaDeltaNeuron(config)
	case "Multiplexed":
		return NewMultiplexedEncoder(config)
	default:
		return NewRateEncoder(config)
	}
}

// =============================================================================
// INTEGRATED CIM WRITE-VERIFY + SPIKE ENCODER SYSTEM
// =============================================================================

// CIMWriteVerifyA2SSystem integrates write verification with spike encoding
type CIMWriteVerifyA2SSystem struct {
	// Configuration
	WriteConfig *WriteVerifyConfig
	SpikeConfig *SpikeEncodingConfig

	// Components
	WriteVerifier *IterativeWriteVerify
	STWController *SelfTerminatingWrite
	SenseAmp      *OffsetCanceledSenseAmplifier

	// Spike encoders
	Encoders map[string]SpikeEncoder

	// Crossbar
	ArrayRows    int
	ArrayCols    int
	Conductances [][]float64

	// Statistics
	TotalWriteOps     int
	TotalEncodeOps    int
	WriteSuccessRate  float64
	AverageLatency    float64 // ns
	AverageEnergy     float64 // pJ
}

// NewCIMWriteVerifyA2SSystem creates an integrated system
func NewCIMWriteVerifyA2SSystem(rows, cols int, writeConfig *WriteVerifyConfig, spikeConfig *SpikeEncodingConfig) *CIMWriteVerifyA2SSystem {
	system := &CIMWriteVerifyA2SSystem{
		WriteConfig:   writeConfig,
		SpikeConfig:   spikeConfig,
		WriteVerifier: NewIterativeWriteVerify(writeConfig),
		STWController: NewSelfTerminatingWrite(writeConfig),
		SenseAmp:      NewOffsetCanceledSA(writeConfig),
		Encoders:      make(map[string]SpikeEncoder),
		ArrayRows:     rows,
		ArrayCols:     cols,
		Conductances:  make([][]float64, rows),
	}

	// Initialize conductances
	for i := 0; i < rows; i++ {
		system.Conductances[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			system.Conductances[i][j] = rand.Float64()
		}
	}

	// Create all encoder types
	encoderTypes := []string{"Rate", "Temporal", "Burst", "TTFS", "Phase", "SigmaDelta", "Multiplexed"}
	for _, encType := range encoderTypes {
		config := *spikeConfig
		config.EncodingType = encType
		system.Encoders[encType] = CreateEncoder(&config)
	}

	return system
}

// ProgramWeights programs the crossbar with target weights
func (sys *CIMWriteVerifyA2SSystem) ProgramWeights(weights [][]float64, useSTW bool) [][]WriteVerifyResult {
	results := make([][]WriteVerifyResult, len(weights))

	for i := range weights {
		results[i] = make([]WriteVerifyResult, len(weights[i]))
		for j := range weights[i] {
			targetG := weights[i][j]
			currentG := sys.Conductances[i][j]

			var result *WriteVerifyResult
			if useSTW {
				result = sys.STWController.WriteWithSTW(currentG, targetG)
			} else {
				currentLevel := int(currentG * float64(sys.WriteConfig.NumLevels))
				targetLevel := int(targetG * float64(sys.WriteConfig.NumLevels))
				result = sys.WriteVerifier.WriteCell(currentLevel, targetLevel)
			}

			results[i][j] = *result
			sys.TotalWriteOps++

			// Update conductance
			if result.Success {
				sys.Conductances[i][j] = targetG
			} else {
				// Partial update
				levelSpacing := 1.0 / float64(sys.WriteConfig.NumLevels)
				sys.Conductances[i][j] = float64(result.AchievedLevel) * levelSpacing
			}

			// Update statistics
			sys.AverageLatency = (sys.AverageLatency*float64(sys.TotalWriteOps-1) + result.TotalTime) / float64(sys.TotalWriteOps)
			sys.AverageEnergy = (sys.AverageEnergy*float64(sys.TotalWriteOps-1) + result.TotalEnergy) / float64(sys.TotalWriteOps)
		}
	}

	// Calculate success rate
	successful := 0
	total := 0
	for i := range results {
		for j := range results[i] {
			total++
			if results[i][j].Success {
				successful++
			}
		}
	}
	sys.WriteSuccessRate = float64(successful) / float64(total)

	return results
}

// EncodeInputs encodes analog inputs to spike trains
func (sys *CIMWriteVerifyA2SSystem) EncodeInputs(inputs []float64, encoderType string) []*SpikeTrain {
	encoder, ok := sys.Encoders[encoderType]
	if !ok {
		encoder = sys.Encoders["Rate"]
	}

	trains := make([]*SpikeTrain, len(inputs))
	for i, input := range inputs {
		trains[i] = encoder.Encode(input)
		sys.TotalEncodeOps++
	}

	return trains
}

// ComputeWithSpikes performs compute-in-memory with spike inputs
func (sys *CIMWriteVerifyA2SSystem) ComputeWithSpikes(spikeTrains []*SpikeTrain) []float64 {
	// Time-based accumulation
	outputs := make([]float64, sys.ArrayCols)

	// For each time step in simulation window
	timeStep := sys.SpikeConfig.TimeStep
	numSteps := int(sys.SpikeConfig.TimeWindow / timeStep)

	// Create spike presence arrays for each step
	for step := 0; step < numSteps; step++ {
		currentTime := float64(step) * timeStep

		// Check which inputs have spikes at this time
		for row := 0; row < sys.ArrayRows && row < len(spikeTrains); row++ {
			train := spikeTrains[row]

			// Check for spike in this time window
			hasSpike := false
			for _, spikeTime := range train.SpikeTimes {
				if spikeTime >= currentTime && spikeTime < currentTime+timeStep {
					hasSpike = true
					break
				}
			}

			if hasSpike {
				// Accumulate weighted contributions
				for col := 0; col < sys.ArrayCols; col++ {
					outputs[col] += sys.Conductances[row][col]
				}
			}
		}
	}

	// Normalize
	for i := range outputs {
		outputs[i] /= float64(numSteps)
	}

	return outputs
}

// GetStatistics returns system statistics
func (sys *CIMWriteVerifyA2SSystem) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_write_ops":     sys.TotalWriteOps,
		"total_encode_ops":    sys.TotalEncodeOps,
		"write_success_rate":  sys.WriteSuccessRate,
		"average_latency_ns":  sys.AverageLatency,
		"average_energy_pJ":   sys.AverageEnergy,
		"array_size":          fmt.Sprintf("%dx%d", sys.ArrayRows, sys.ArrayCols),
	}
}

// =============================================================================
// ENCODER BENCHMARK
// =============================================================================

// EncoderBenchmark benchmarks different encoding schemes
type EncoderBenchmark struct {
	Config        *SpikeEncodingConfig
	Encoders      map[string]SpikeEncoder
	TestValues    []float64
	Results       map[string]*EncoderBenchmarkResult
}

// EncoderBenchmarkResult stores benchmark results for one encoder
type EncoderBenchmarkResult struct {
	EncoderName        string
	AverageSpikes      float64
	AverageError       float64
	MaxError           float64
	Efficiency         float64
	TotalSpikes        int
	ReconstructionMSE  float64
}

// NewEncoderBenchmark creates a benchmark
func NewEncoderBenchmark(config *SpikeEncodingConfig) *EncoderBenchmark {
	bench := &EncoderBenchmark{
		Config:     config,
		Encoders:   make(map[string]SpikeEncoder),
		TestValues: make([]float64, 0),
		Results:    make(map[string]*EncoderBenchmarkResult),
	}

	// Create all encoders
	encoderTypes := []string{"Rate", "Temporal", "Burst", "TTFS", "Phase", "SigmaDelta", "Multiplexed"}
	for _, encType := range encoderTypes {
		cfg := *config
		cfg.EncodingType = encType
		bench.Encoders[encType] = CreateEncoder(&cfg)
	}

	// Generate test values
	for i := 0; i <= 100; i++ {
		bench.TestValues = append(bench.TestValues, float64(i)/100.0)
	}

	return bench
}

// Run executes the benchmark
func (bench *EncoderBenchmark) Run() {
	for name, encoder := range bench.Encoders {
		result := &EncoderBenchmarkResult{
			EncoderName: name,
		}

		totalSpikes := 0
		totalError := 0.0
		maxError := 0.0
		totalSqError := 0.0

		for _, value := range bench.TestValues {
			// Encode
			train := encoder.Encode(value)
			totalSpikes += train.NumSpikes

			// Decode
			decoded := encoder.Decode(train)

			// Error
			err := math.Abs(decoded - value)
			totalError += err
			if err > maxError {
				maxError = err
			}
			totalSqError += err * err
		}

		n := float64(len(bench.TestValues))
		result.TotalSpikes = totalSpikes
		result.AverageSpikes = float64(totalSpikes) / n
		result.AverageError = totalError / n
		result.MaxError = maxError
		result.ReconstructionMSE = totalSqError / n
		result.Efficiency = encoder.GetEfficiency()

		bench.Results[name] = result
	}
}

// GetRanking returns encoders ranked by efficiency-adjusted accuracy
func (bench *EncoderBenchmark) GetRanking() []string {
	type rankedEncoder struct {
		name  string
		score float64
	}

	ranked := make([]rankedEncoder, 0)
	for name, result := range bench.Results {
		// Score = efficiency * (1 - normalized_error)
		// Higher is better
		score := result.Efficiency * (1.0 - result.AverageError)
		ranked = append(ranked, rankedEncoder{name, score})
	}

	// Sort by score descending
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	names := make([]string, len(ranked))
	for i, r := range ranked {
		names[i] = r.name
	}

	return names
}

// =============================================================================
// JSON SERIALIZATION
// =============================================================================

// WriteVerifyA2SExport exports system configuration
type WriteVerifyA2SExport struct {
	WriteConfig    *WriteVerifyConfig       `json:"write_config"`
	SpikeConfig    *SpikeEncodingConfig     `json:"spike_config"`
	ArrayDimensions [2]int                  `json:"array_dimensions"`
	Statistics     map[string]interface{}   `json:"statistics"`
	BenchmarkResults map[string]*EncoderBenchmarkResult `json:"benchmark_results,omitempty"`
}

// ExportJSON exports system to JSON
func (sys *CIMWriteVerifyA2SSystem) ExportJSON() ([]byte, error) {
	// Run benchmark
	bench := NewEncoderBenchmark(sys.SpikeConfig)
	bench.Run()

	export := WriteVerifyA2SExport{
		WriteConfig:     sys.WriteConfig,
		SpikeConfig:     sys.SpikeConfig,
		ArrayDimensions: [2]int{sys.ArrayRows, sys.ArrayCols},
		Statistics:      sys.GetStatistics(),
		BenchmarkResults: bench.Results,
	}

	return json.MarshalIndent(export, "", "  ")
}

// =============================================================================
// FECIM INTEGRATED SYSTEM
// =============================================================================

// FeCIMWriteVerifyA2S integrates all components for FeCIM
type FeCIMWriteVerifyA2S struct {
	// Core system
	CIMSystem *CIMWriteVerifyA2SSystem

	// HZO-specific parameters
	HZOParameters struct {
		CoerciveField    float64 // MV/cm
		Remanence        float64 // μC/cm²
		EnduranceCycles  float64
		RetentionYears   float64
	}

	// Optimal configurations
	OptimalWriteScheme   string
	OptimalEncodingScheme string

	// Performance metrics
	InferenceThroughput float64 // GOPS
	EnergyEfficiency    float64 // TOPS/W
	WriteLatency        float64 // ns per weight
}

// NewFeCIMWriteVerifyA2S creates an FeCIM-optimized system
func NewFeCIMWriteVerifyA2S(rows, cols int) *FeCIMWriteVerifyA2S {
	// Configure for HZO ferroelectric
	writeConfig := &WriteVerifyConfig{
		MemoryType:       "FeFET",
		CellStructure:    "1T",
		NumLevels:        16,  // 4-bit MLC
		MaxWritePulses:   8,
		WritePulseWidth:  20.0,   // 20 ns fast write
		WriteVoltageStep: 0.2,
		WriteVoltageMax:  4.0,
		VerifyMargin:     30.0,
		VerifyDelay:      5.0,
		SenseAmpType:     "OffsetCanceled",
		OffsetVoltage:    15.0,   // Improved offset cancellation
		SenseMargin:      80.0,
		ProcessVariation: 0.03,   // 3% variation
		NoiseLevel:       0.01,
		LowPowerMode:     true,
		AdaptivePulsing:  true,
	}

	spikeConfig := &SpikeEncodingConfig{
		EncodingType:     "SigmaDelta",
		TimeWindow:       5.0,   // 5 ms - reduced latency
		TimeStep:         0.05,  // 50 μs resolution
		MaxSpikes:        50,
		Threshold:        0.8,
		ResetValue:       0.0,
		RefractoryPeriod: 0.2,
		LeakRate:         0.05,
		Integration:      1.2,
		NoiseLevel:       0.01,
		JitterLevel:      0.05,
	}

	system := &FeCIMWriteVerifyA2S{
		CIMSystem: NewCIMWriteVerifyA2SSystem(rows, cols, writeConfig, spikeConfig),
	}

	// Set HZO parameters
	system.HZOParameters.CoerciveField = 0.85   // MV/cm
	system.HZOParameters.Remanence = 25.0       // μC/cm²
	system.HZOParameters.EnduranceCycles = 1e10 // 10 billion cycles
	system.HZOParameters.RetentionYears = 10.0

	// Determine optimal schemes
	system.OptimalWriteScheme = "STW"           // Self-terminating for speed
	system.OptimalEncodingScheme = "SigmaDelta" // Best balance

	// Calculate performance
	system.InferenceThroughput = float64(rows*cols) / (spikeConfig.TimeWindow * 1e-3) / 1e9 // GOPS
	system.EnergyEfficiency = 310.0 // TOPS/W (based on Neuro-CIM)
	system.WriteLatency = writeConfig.WritePulseWidth * float64(writeConfig.MaxWritePulses) / 2.0

	return system
}

// Inference performs complete inference with spike encoding
func (ils *FeCIMWriteVerifyA2S) Inference(inputs []float64) []float64 {
	// Encode inputs
	trains := ils.CIMSystem.EncodeInputs(inputs, ils.OptimalEncodingScheme)

	// Compute
	outputs := ils.CIMSystem.ComputeWithSpikes(trains)

	return outputs
}

// ProgramModel programs weights using optimal write scheme
func (ils *FeCIMWriteVerifyA2S) ProgramModel(weights [][]float64) {
	useSTW := ils.OptimalWriteScheme == "STW"
	ils.CIMSystem.ProgramWeights(weights, useSTW)
}

// GetPerformanceReport returns performance summary
func (ils *FeCIMWriteVerifyA2S) GetPerformanceReport() string {
	stats := ils.CIMSystem.GetStatistics()

	report := fmt.Sprintf(`
FeCIM Write-Verify + A2S System Performance
==================================================
Array Size: %s
Write Success Rate: %.1f%%
Average Write Latency: %.1f ns
Average Write Energy: %.2f pJ

HZO Ferroelectric Parameters:
  Coercive Field: %.2f MV/cm
  Remanence: %.1f μC/cm²
  Endurance: %.0e cycles
  Retention: %.0f years

Optimal Configurations:
  Write Scheme: %s
  Encoding Scheme: %s

System Performance:
  Inference Throughput: %.2f GOPS
  Energy Efficiency: %.1f TOPS/W
  Write Latency: %.1f ns/weight
`,
		stats["array_size"],
		stats["write_success_rate"].(float64)*100,
		stats["average_latency_ns"],
		stats["average_energy_pJ"],
		ils.HZOParameters.CoerciveField,
		ils.HZOParameters.Remanence,
		ils.HZOParameters.EnduranceCycles,
		ils.HZOParameters.RetentionYears,
		ils.OptimalWriteScheme,
		ils.OptimalEncodingScheme,
		ils.InferenceThroughput,
		ils.EnergyEfficiency,
		ils.WriteLatency,
	)

	return report
}
