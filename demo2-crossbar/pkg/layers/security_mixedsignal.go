// Package layers provides CIM security and mixed-signal design simulation.
// This module covers adversarial robustness, side-channel attack modeling,
// weight protection, and mixed-signal peripheral circuit design.
package layers

import (
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// CIM ADVERSARIAL ROBUSTNESS
// =============================================================================

// AdversarialAttackType defines types of adversarial attacks
type AdversarialAttackType int

const (
	AttackPGD       AdversarialAttackType = iota // Projected Gradient Descent
	AttackFGSM                                    // Fast Gradient Sign Method
	AttackCW                                      // Carlini & Wagner
	AttackSquare                                  // Square attack (gradient-free)
	AttackOnePixel                                // One-pixel attack
	AttackAutoAttack                              // AutoAttack ensemble
)

// AdversarialAttackConfig configures an adversarial attack
type AdversarialAttackConfig struct {
	Type           AdversarialAttackType
	Epsilon        float64 // Perturbation budget (L-inf norm)
	NumIterations  int     // For iterative attacks (PGD)
	StepSize       float64 // Step size for iterative attacks
	Confidence     float64 // For CW attack
	NumQueries     int     // For query-based attacks (Square)
	Targeted       bool    // Targeted vs untargeted attack
	TargetClass    int     // Target class for targeted attacks
}

// DefaultPGDConfig returns default PGD attack configuration
func DefaultPGDConfig(epsilon float64) *AdversarialAttackConfig {
	return &AdversarialAttackConfig{
		Type:          AttackPGD,
		Epsilon:       epsilon,
		NumIterations: 40,
		StepSize:      epsilon / 10,
		Targeted:      false,
	}
}

// AdversarialAttacker generates adversarial perturbations
type AdversarialAttacker struct {
	Config *AdversarialAttackConfig
}

// NewAdversarialAttacker creates a new adversarial attacker
func NewAdversarialAttacker(config *AdversarialAttackConfig) *AdversarialAttacker {
	return &AdversarialAttacker{Config: config}
}

// GeneratePerturbation generates adversarial perturbation for input
func (a *AdversarialAttacker) GeneratePerturbation(input []float64, gradient []float64) []float64 {
	perturbation := make([]float64, len(input))

	switch a.Config.Type {
	case AttackFGSM:
		// Fast Gradient Sign Method
		for i := range gradient {
			if gradient[i] > 0 {
				perturbation[i] = a.Config.Epsilon
			} else if gradient[i] < 0 {
				perturbation[i] = -a.Config.Epsilon
			}
		}

	case AttackPGD:
		// Projected Gradient Descent (single step)
		for i := range gradient {
			// Take step in gradient direction
			perturbation[i] = a.Config.StepSize * sign(gradient[i])
			// Project to epsilon ball
			perturbation[i] = clamp(perturbation[i], -a.Config.Epsilon, a.Config.Epsilon)
		}

	case AttackSquare:
		// Square attack - random square perturbations
		size := int(math.Sqrt(float64(len(input))))
		squareSize := size / 4
		startX := rand.Intn(size - squareSize)
		startY := rand.Intn(size - squareSize)

		for y := startY; y < startY+squareSize; y++ {
			for x := startX; x < startX+squareSize; x++ {
				idx := y*size + x
				if idx < len(perturbation) {
					if rand.Float64() < 0.5 {
						perturbation[idx] = a.Config.Epsilon
					} else {
						perturbation[idx] = -a.Config.Epsilon
					}
				}
			}
		}

	case AttackOnePixel:
		// One-pixel attack
		idx := rand.Intn(len(input))
		perturbation[idx] = a.Config.Epsilon
	}

	return perturbation
}

// =============================================================================
// ANALOG CIM NOISE-BASED DEFENSE
// =============================================================================

// CIMNoiseDefenseConfig configures noise-based adversarial defense
type CIMNoiseDefenseConfig struct {
	// Intrinsic noise sources (analog CIM)
	ReadNoiseStd       float64 // Read noise standard deviation
	WriteNoiseStd      float64 // Write noise (programming variability)
	ConductanceVar     float64 // Device-to-device conductance variation
	TemporalDriftRate  float64 // Conductance drift rate

	// Noise injection parameters
	InjectStochasticNoise bool
	InjectedNoiseStd      float64
	UseQuantizationNoise  bool
	QuantizationBits      int

	// Defense effectiveness metrics
	ExpectedRobustGain float64 // Expected robustness improvement
}

// DefaultCIMNoiseDefenseConfig returns default noise-based defense config
func DefaultCIMNoiseDefenseConfig() *CIMNoiseDefenseConfig {
	return &CIMNoiseDefenseConfig{
		ReadNoiseStd:          0.02,   // 2% read noise
		WriteNoiseStd:         0.05,   // 5% write variability
		ConductanceVar:        0.03,   // 3% device variation
		TemporalDriftRate:     0.001,  // 0.1% drift per inference
		InjectStochasticNoise: true,
		InjectedNoiseStd:      0.01,
		UseQuantizationNoise:  true,
		QuantizationBits:      6,
		ExpectedRobustGain:    1.5, // 50% more robust
	}
}

// CIMNoiseDefender implements noise-based adversarial defense
type CIMNoiseDefender struct {
	Config *CIMNoiseDefenseConfig

	// Statistics
	TotalInferences        int64
	AdversarialDetected    int64
	NoiseMitigatedAttacks  int64
}

// NewCIMNoiseDefender creates a new noise-based defender
func NewCIMNoiseDefender(config *CIMNoiseDefenseConfig) *CIMNoiseDefender {
	return &CIMNoiseDefender{Config: config}
}

// ApplyDefense applies noise-based defense to computation
func (d *CIMNoiseDefender) ApplyDefense(values []float64) []float64 {
	d.TotalInferences++
	defended := make([]float64, len(values))

	for i, v := range values {
		// Apply intrinsic noise sources
		noise := 0.0

		// Read noise (recurrent, each inference)
		noise += rand.NormFloat64() * d.Config.ReadNoiseStd * math.Abs(v)

		// Conductance variation (non-recurrent)
		noise += rand.NormFloat64() * d.Config.ConductanceVar * math.Abs(v)

		// Temporal drift
		noise += rand.NormFloat64() * d.Config.TemporalDriftRate * math.Abs(v)

		// Injected stochastic noise
		if d.Config.InjectStochasticNoise {
			noise += rand.NormFloat64() * d.Config.InjectedNoiseStd
		}

		defended[i] = v + noise

		// Quantization noise
		if d.Config.UseQuantizationNoise {
			levels := math.Pow(2, float64(d.Config.QuantizationBits))
			defended[i] = math.Round(defended[i]*levels) / levels
		}
	}

	return defended
}

// EvaluateRobustness evaluates robustness against attack
func (d *CIMNoiseDefender) EvaluateRobustness(
	cleanAccuracy float64,
	attackedAccuracy float64,
	defendedAccuracy float64,
) *RobustnessMetrics {
	return &RobustnessMetrics{
		CleanAccuracy:     cleanAccuracy,
		AttackedAccuracy:  attackedAccuracy,
		DefendedAccuracy:  defendedAccuracy,
		RobustnessGain:    defendedAccuracy - attackedAccuracy,
		AccuracyDrop:      cleanAccuracy - defendedAccuracy,
		DefenseEfficiency: (defendedAccuracy - attackedAccuracy) / (cleanAccuracy - attackedAccuracy),
	}
}

// RobustnessMetrics contains robustness evaluation metrics
type RobustnessMetrics struct {
	CleanAccuracy     float64
	AttackedAccuracy  float64
	DefendedAccuracy  float64
	RobustnessGain    float64 // Improvement over attacked
	AccuracyDrop      float64 // Drop from clean
	DefenseEfficiency float64 // Fraction of attack mitigated
}

// =============================================================================
// SIDE-CHANNEL ATTACK MODELING
// =============================================================================

// SideChannelAttackType defines types of side-channel attacks
type SideChannelAttackType int

const (
	SideChannelPower SideChannelAttackType = iota // Power analysis
	SideChannelEM                                  // Electromagnetic emanations
	SideChannelTiming                              // Timing analysis
	SideChannelCache                               // Cache-based
)

// SideChannelAttackConfig configures a side-channel attack
type SideChannelAttackConfig struct {
	Type               SideChannelAttackType
	NumTraces          int     // Number of power/EM traces
	SamplingRate       float64 // Hz
	NoiseLevel         float64 // Measurement noise
	AttackerCapability string  // "passive", "active", "physical"
}

// SideChannelAttacker models side-channel attacks on CIM
type SideChannelAttacker struct {
	Config *SideChannelAttackConfig

	// Attack state
	CollectedTraces   [][]float64
	ExtractedWeights  []float64
	ExtractionSuccess float64
}

// NewSideChannelAttacker creates a new side-channel attacker
func NewSideChannelAttacker(config *SideChannelAttackConfig) *SideChannelAttacker {
	return &SideChannelAttacker{
		Config:          config,
		CollectedTraces: make([][]float64, 0),
	}
}

// CollectTrace simulates collecting a power/EM trace
func (a *SideChannelAttacker) CollectTrace(weights []float64, inputs []float64) []float64 {
	trace := make([]float64, len(weights))

	for i, w := range weights {
		var inp float64
		if i < len(inputs) {
			inp = inputs[i]
		}

		// Power consumption model: P ∝ V² × activity × capacitance
		// For CIM: Power depends on weight values and input activations
		switch a.Config.Type {
		case SideChannelPower:
			// Power consumption correlates with MAC operations
			activity := math.Abs(w * inp)
			basePower := 0.1 // Base power
			dynamicPower := activity * 0.5
			trace[i] = basePower + dynamicPower + rand.NormFloat64()*a.Config.NoiseLevel

		case SideChannelEM:
			// EM emissions correlate with current flow
			current := w * inp
			trace[i] = math.Abs(current) + rand.NormFloat64()*a.Config.NoiseLevel

		case SideChannelTiming:
			// Timing depends on ADC conversion time (weight-dependent)
			baseTime := 10.0 // ns
			weightDependentTime := math.Abs(w) * 2.0
			trace[i] = baseTime + weightDependentTime + rand.NormFloat64()*a.Config.NoiseLevel
		}
	}

	a.CollectedTraces = append(a.CollectedTraces, trace)
	return trace
}

// ExtractWeights attempts to extract weights from collected traces
func (a *SideChannelAttacker) ExtractWeights() []float64 {
	if len(a.CollectedTraces) < a.Config.NumTraces {
		return nil // Not enough traces
	}

	// Simple correlation-based extraction
	numWeights := len(a.CollectedTraces[0])
	extracted := make([]float64, numWeights)

	for i := 0; i < numWeights; i++ {
		// Average traces for this weight position
		sum := 0.0
		for _, trace := range a.CollectedTraces {
			sum += trace[i]
		}
		avg := sum / float64(len(a.CollectedTraces))

		// Estimate weight from average power/EM
		// This is a simplified model - real attacks use more sophisticated DPA/CPA
		switch a.Config.Type {
		case SideChannelPower:
			extracted[i] = (avg - 0.1) / 0.5 // Inverse of power model
		case SideChannelEM:
			extracted[i] = avg // Direct correlation
		case SideChannelTiming:
			extracted[i] = (avg - 10.0) / 2.0 // Inverse of timing model
		}
	}

	a.ExtractedWeights = extracted
	return extracted
}

// EvaluateExtraction evaluates weight extraction accuracy
func (a *SideChannelAttacker) EvaluateExtraction(trueWeights []float64) float64 {
	if len(a.ExtractedWeights) != len(trueWeights) {
		return 0
	}

	// Calculate correlation coefficient
	n := float64(len(trueWeights))
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for i := range trueWeights {
		x := trueWeights[i]
		y := a.ExtractedWeights[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	correlation := numerator / denominator
	a.ExtractionSuccess = correlation
	return correlation
}

// =============================================================================
// WEIGHT PROTECTION MECHANISMS
// =============================================================================

// WeightProtectionType defines weight protection mechanisms
type WeightProtectionType int

const (
	ProtectionNone       WeightProtectionType = iota
	ProtectionObfuscation                     // Weight obfuscation
	ProtectionEncryption                      // Weight encryption
	ProtectionSplitting                       // Weight splitting across devices
	ProtectionNoise                           // Noise injection
)

// WeightProtectionConfig configures weight protection
type WeightProtectionConfig struct {
	Type            WeightProtectionType
	ObfuscationKey  []float64 // For obfuscation
	EncryptionKey   []byte    // For encryption
	NumSplits       int       // For splitting
	NoiseLevel      float64   // For noise injection
	RefreshInterval int       // Re-obfuscation interval
}

// WeightProtector implements weight protection mechanisms
type WeightProtector struct {
	Config          *WeightProtectionConfig
	ProtectedWeights []float64
	OriginalWeights  []float64
	InferenceCount   int
}

// NewWeightProtector creates a new weight protector
func NewWeightProtector(config *WeightProtectionConfig) *WeightProtector {
	return &WeightProtector{Config: config}
}

// ProtectWeights applies protection to weights
func (p *WeightProtector) ProtectWeights(weights []float64) []float64 {
	p.OriginalWeights = make([]float64, len(weights))
	copy(p.OriginalWeights, weights)

	protected := make([]float64, len(weights))

	switch p.Config.Type {
	case ProtectionObfuscation:
		// XOR-like obfuscation with key
		for i := range weights {
			keyIdx := i % len(p.Config.ObfuscationKey)
			// Obfuscate by adding key (reversible)
			protected[i] = weights[i] + p.Config.ObfuscationKey[keyIdx]
		}

	case ProtectionSplitting:
		// Split weight into random shares
		for i := range weights {
			// Generate random shares that sum to original
			share1 := rand.Float64() * 2 - 1
			share2 := weights[i] - share1
			// Store only one share (other in secure location)
			protected[i] = share1
			// In practice, share2 would be stored elsewhere
			_ = share2
		}

	case ProtectionNoise:
		// Add controlled noise
		for i := range weights {
			protected[i] = weights[i] + rand.NormFloat64()*p.Config.NoiseLevel
		}

	default:
		copy(protected, weights)
	}

	p.ProtectedWeights = protected
	return protected
}

// DeprotectWeights removes protection for computation
func (p *WeightProtector) DeprotectWeights() []float64 {
	p.InferenceCount++

	// Check if refresh needed
	if p.Config.RefreshInterval > 0 && p.InferenceCount%p.Config.RefreshInterval == 0 {
		// Re-obfuscate with new key
		p.refreshProtection()
	}

	switch p.Config.Type {
	case ProtectionObfuscation:
		deprotected := make([]float64, len(p.ProtectedWeights))
		for i := range p.ProtectedWeights {
			keyIdx := i % len(p.Config.ObfuscationKey)
			deprotected[i] = p.ProtectedWeights[i] - p.Config.ObfuscationKey[keyIdx]
		}
		return deprotected

	default:
		return p.OriginalWeights
	}
}

// refreshProtection refreshes the protection with new randomization
func (p *WeightProtector) refreshProtection() {
	// Generate new obfuscation key
	newKey := make([]float64, len(p.Config.ObfuscationKey))
	for i := range newKey {
		newKey[i] = rand.Float64()*2 - 1
	}

	// Re-protect with new key
	for i := range p.ProtectedWeights {
		oldKeyIdx := i % len(p.Config.ObfuscationKey)
		newKeyIdx := i % len(newKey)
		// Remove old obfuscation, add new
		p.ProtectedWeights[i] = p.ProtectedWeights[i] - p.Config.ObfuscationKey[oldKeyIdx] + newKey[newKeyIdx]
	}

	p.Config.ObfuscationKey = newKey
}

// =============================================================================
// MIXED-SIGNAL CIM PERIPHERAL CIRCUITS
// =============================================================================

// PeripheralCircuitType defines types of peripheral circuits
type PeripheralCircuitType int

const (
	PeripheralTIA         PeripheralCircuitType = iota // Transimpedance Amplifier
	PeripheralSA                                        // Sense Amplifier
	PeripheralSC                                        // Switched Capacitor
	PeripheralChargeDomain                              // Charge domain computing
	PeripheralCurrentMirror                             // Current mirror
)

// PeripheralCircuitConfig configures peripheral circuit
type PeripheralCircuitConfig struct {
	Type           PeripheralCircuitType
	Technology     int     // nm process node
	VDD            float64 // Supply voltage
	Bandwidth      float64 // Hz
	Gain           float64 // Amplifier gain
	InputRange     float64 // Input current/voltage range
	NoiseFloor     float64 // Input-referred noise
}

// PeripheralCircuit models a CIM peripheral circuit
type PeripheralCircuit struct {
	Config      *PeripheralCircuitConfig
	Area        float64 // mm²
	Power       float64 // mW
	Latency     float64 // ns
	SNR         float64 // dB
}

// NewPeripheralCircuit creates a new peripheral circuit with calculated metrics
func NewPeripheralCircuit(config *PeripheralCircuitConfig) *PeripheralCircuit {
	pc := &PeripheralCircuit{Config: config}
	pc.calculateMetrics()
	return pc
}

// calculateMetrics calculates area, power, latency metrics
func (pc *PeripheralCircuit) calculateMetrics() {
	techScale := float64(pc.Config.Technology) / 28.0

	switch pc.Config.Type {
	case PeripheralTIA:
		// Transimpedance Amplifier
		// Area: ~0.01 mm² at 28nm
		pc.Area = 0.01 * techScale * techScale
		// Power: ~1-5 mW depending on bandwidth
		pc.Power = 2.0 * (pc.Config.Bandwidth / 1e9) * techScale
		// Latency: ~1/BW
		pc.Latency = 1e9 / pc.Config.Bandwidth
		// SNR depends on noise floor and signal range
		pc.SNR = 20 * math.Log10(pc.Config.InputRange/pc.Config.NoiseFloor)

	case PeripheralSA:
		// Sense Amplifier
		pc.Area = 0.001 * techScale * techScale // Much smaller than TIA
		pc.Power = 0.1 * techScale
		pc.Latency = 1.0 // 1 ns typical
		pc.SNR = 60      // High SNR for digital sensing

	case PeripheralSC:
		// Switched Capacitor circuit
		pc.Area = 0.005 * techScale * techScale
		pc.Power = 0.5 * (pc.Config.Bandwidth / 1e9) * techScale
		pc.Latency = 2e9 / pc.Config.Bandwidth // 2 clock phases
		pc.SNR = 50 + 6*float64(6) // 6-bit equivalent

	case PeripheralChargeDomain:
		// Charge domain computing
		pc.Area = 0.003 * techScale * techScale
		pc.Power = 0.2 * techScale
		pc.Latency = 5.0 // Charge integration time
		pc.SNR = 55

	case PeripheralCurrentMirror:
		// Current mirror
		pc.Area = 0.0005 * techScale * techScale
		pc.Power = 0.05 * techScale
		pc.Latency = 0.5
		pc.SNR = 45
	}
}

// Process processes input through the peripheral circuit
func (pc *PeripheralCircuit) Process(input float64) float64 {
	// Apply gain
	output := input * pc.Config.Gain

	// Add noise based on SNR
	noiseStd := math.Abs(output) / math.Pow(10, pc.SNR/20)
	output += rand.NormFloat64() * noiseStd

	// Clamp to output range
	maxOutput := pc.Config.InputRange * pc.Config.Gain
	output = clamp(output, -maxOutput, maxOutput)

	return output
}

// =============================================================================
// CHARGE DOMAIN COMPUTING
// =============================================================================

// ChargeDomainConfig configures charge domain CIM
type ChargeDomainConfig struct {
	CapacitorValue   float64 // F
	VoltageRange     float64 // V
	ChargeResolution int     // bits
	NumCapacitors    int     // Number of weighted capacitors
	UseC2CLadder     bool    // Use C-2C ladder
}

// ChargeDomainUnit implements charge domain computing
type ChargeDomainUnit struct {
	Config          *ChargeDomainConfig
	Capacitors      []float64 // Capacitor values (for C-2C ladder)
	ChargeState     []float64 // Current charge on each capacitor
	EnergyPerMAC    float64   // fJ/MAC
	AreaEfficiency  float64   // TOPS/mm²
}

// NewChargeDomainUnit creates a new charge domain computing unit
func NewChargeDomainUnit(config *ChargeDomainConfig) *ChargeDomainUnit {
	cdu := &ChargeDomainUnit{
		Config:      config,
		Capacitors:  make([]float64, config.NumCapacitors),
		ChargeState: make([]float64, config.NumCapacitors),
	}

	// Initialize C-2C ladder if enabled
	if config.UseC2CLadder {
		baseC := config.CapacitorValue
		for i := 0; i < config.NumCapacitors; i++ {
			if i == 0 {
				cdu.Capacitors[i] = 2 * baseC
			} else {
				cdu.Capacitors[i] = baseC
			}
		}
	} else {
		// Binary weighted capacitors
		for i := 0; i < config.NumCapacitors; i++ {
			cdu.Capacitors[i] = config.CapacitorValue * math.Pow(2, float64(config.NumCapacitors-1-i))
		}
	}

	// Calculate energy efficiency
	// E = C × V² for each MAC
	totalC := 0.0
	for _, c := range cdu.Capacitors {
		totalC += c
	}
	cdu.EnergyPerMAC = 0.5 * totalC * config.VoltageRange * config.VoltageRange * 1e15 // fJ

	// Area efficiency (assuming 32.2 TOPS/W reference)
	cdu.AreaEfficiency = 4.0 // TOPS/mm² typical for charge domain

	return cdu
}

// ComputeMAC performs MAC operation in charge domain
func (cdu *ChargeDomainUnit) ComputeMAC(weights []float64, input float64) float64 {
	if len(weights) > len(cdu.Capacitors) {
		weights = weights[:len(cdu.Capacitors)]
	}

	// Charge each capacitor with weight × input × voltage
	totalCharge := 0.0
	for i, w := range weights {
		// Q = C × V, where V = weight × input × VDD
		voltage := w * input * cdu.Config.VoltageRange
		charge := cdu.Capacitors[i] * voltage
		cdu.ChargeState[i] = charge
		totalCharge += charge
	}

	// Convert total charge to output voltage
	// V_out = Q_total / C_total
	totalC := 0.0
	for i := 0; i < len(weights); i++ {
		totalC += cdu.Capacitors[i]
	}

	output := totalCharge / totalC

	// Add quantization noise
	levels := math.Pow(2, float64(cdu.Config.ChargeResolution))
	lsb := cdu.Config.VoltageRange / levels
	output = math.Round(output/lsb) * lsb

	return output
}

// =============================================================================
// TRANSIMPEDANCE AMPLIFIER (TIA)
// =============================================================================

// TIAConfig configures a transimpedance amplifier
type TIAConfig struct {
	FeedbackResistance float64 // Ohms (transimpedance)
	Bandwidth          float64 // Hz
	InputCapacitance   float64 // F
	NoiseCurrentDensity float64 // A/sqrt(Hz)
	GainBandwidthProduct float64 // Hz
}

// TIA implements a transimpedance amplifier for CIM current sensing
type TIA struct {
	Config     *TIAConfig
	OutputSNR  float64
	Power      float64 // mW
}

// NewTIA creates a new transimpedance amplifier
func NewTIA(config *TIAConfig) *TIA {
	tia := &TIA{Config: config}

	// Calculate power (simplified model)
	// P ≈ GBW × Cin × VDD² / 10
	vdd := 1.0 // Assume 1V
	tia.Power = config.GainBandwidthProduct * config.InputCapacitance * vdd * vdd / 10 * 1000 // mW

	// Calculate SNR
	// Noise = noise density × sqrt(bandwidth)
	noiseRMS := config.NoiseCurrentDensity * math.Sqrt(config.Bandwidth)
	// Assume 1uA full-scale input
	signalRMS := 1e-6
	tia.OutputSNR = 20 * math.Log10(signalRMS/noiseRMS)

	return tia
}

// Convert converts input current to output voltage
func (tia *TIA) Convert(inputCurrent float64) float64 {
	// V_out = I_in × R_feedback
	output := inputCurrent * tia.Config.FeedbackResistance

	// Add noise
	noiseRMS := tia.Config.NoiseCurrentDensity * math.Sqrt(tia.Config.Bandwidth) * tia.Config.FeedbackResistance
	output += rand.NormFloat64() * noiseRMS

	return output
}

// =============================================================================
// CONFIGURABLE SENSE AMPLIFIER (CSA)
// =============================================================================

// SenseAmpMode defines sense amplifier operating mode
type SenseAmpMode int

const (
	SAModeADC   SenseAmpMode = iota // ADC mode
	SAModeWTA                        // Winner-take-all mode
	SAModeCAM                        // Content-addressable memory mode
)

// ConfigurableSAConfig configures a configurable sense amplifier
type ConfigurableSAConfig struct {
	Mode           SenseAmpMode
	Resolution     int     // bits (for ADC mode)
	NumInputs      int     // Number of inputs (for WTA mode)
	MatchThreshold float64 // For CAM mode
	Technology     int     // nm
}

// ConfigurableSA implements a configurable sense amplifier
type ConfigurableSA struct {
	Config     *ConfigurableSAConfig
	CurrentMode SenseAmpMode
	Area        float64 // mm²
	Energy      float64 // fJ/operation
}

// NewConfigurableSA creates a new configurable sense amplifier
func NewConfigurableSA(config *ConfigurableSAConfig) *ConfigurableSA {
	csa := &ConfigurableSA{
		Config:      config,
		CurrentMode: config.Mode,
	}

	// Calculate area and energy based on mode
	techScale := float64(config.Technology) / 28.0

	switch config.Mode {
	case SAModeADC:
		// SAR ADC mode
		csa.Area = 0.0002 * math.Pow(2, float64(config.Resolution-6)) * techScale * techScale
		csa.Energy = 5.0 * float64(config.Resolution) // fJ

	case SAModeWTA:
		// Winner-take-all mode
		csa.Area = 0.0001 * float64(config.NumInputs) * techScale * techScale
		csa.Energy = 2.0 * float64(config.NumInputs) // fJ

	case SAModeCAM:
		// CAM mode
		csa.Area = 0.00015 * float64(config.NumInputs) * techScale * techScale
		csa.Energy = 3.0 * float64(config.NumInputs) // fJ
	}

	return csa
}

// SetMode configures the sense amplifier mode
func (csa *ConfigurableSA) SetMode(mode SenseAmpMode) {
	csa.CurrentMode = mode
}

// ProcessADC performs ADC conversion
func (csa *ConfigurableSA) ProcessADC(input float64, fullScale float64) int {
	if csa.CurrentMode != SAModeADC {
		return 0
	}

	// Normalize and quantize
	normalized := input / fullScale
	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	levels := int(math.Pow(2, float64(csa.Config.Resolution)))
	return int(normalized * float64(levels-1))
}

// ProcessWTA finds the winner (maximum) among inputs
func (csa *ConfigurableSA) ProcessWTA(inputs []float64) int {
	if csa.CurrentMode != SAModeWTA {
		return -1
	}

	maxIdx := 0
	maxVal := inputs[0]
	for i, v := range inputs {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// ProcessCAM performs content-addressable memory search
func (csa *ConfigurableSA) ProcessCAM(query []float64, stored [][]float64) []int {
	if csa.CurrentMode != SAModeCAM {
		return nil
	}

	matches := make([]int, 0)
	for i, pattern := range stored {
		// Calculate Hamming distance (for binary) or L1 distance (for analog)
		distance := 0.0
		for j := range query {
			if j < len(pattern) {
				distance += math.Abs(query[j] - pattern[j])
			}
		}

		if distance <= csa.Config.MatchThreshold {
			matches = append(matches, i)
		}
	}
	return matches
}

// =============================================================================
// MIXED-SIGNAL CIM MACRO
// =============================================================================

// MixedSignalCIMConfig configures a mixed-signal CIM macro
type MixedSignalCIMConfig struct {
	// Array configuration
	Rows              int
	Cols              int
	CellType          string // "SRAM", "RRAM", "FeFET"

	// DAC configuration
	DACBits           int
	DACType           string // "R2R", "Binary", "Thermometer"

	// ADC configuration
	ADCBits           int
	ADCType           string // "SAR", "Flash", "SC"

	// Peripheral configuration
	UseChargeDomain   bool
	UseTIA            bool
	UseConfigurableSA bool

	// Process parameters
	Technology        int     // nm
	VDD               float64 // V
}

// MixedSignalCIMMacro implements a complete mixed-signal CIM macro
type MixedSignalCIMMacro struct {
	Config          *MixedSignalCIMConfig
	Weights         [][]float64

	// Peripheral circuits
	ChargeDomain    *ChargeDomainUnit
	TIA             *TIA
	SenseAmps       []*ConfigurableSA

	// Performance metrics
	TotalArea       float64 // mm²
	ArrayArea       float64 // mm²
	PeripheralArea  float64 // mm²
	EnergyPerMAC    float64 // fJ
	Throughput      float64 // TOPS
	EnergyEff       float64 // TOPS/W
	AreaEff         float64 // TOPS/mm²
}

// NewMixedSignalCIMMacro creates a new mixed-signal CIM macro
func NewMixedSignalCIMMacro(config *MixedSignalCIMConfig) *MixedSignalCIMMacro {
	macro := &MixedSignalCIMMacro{
		Config:  config,
		Weights: make([][]float64, config.Rows),
	}

	// Initialize weights
	for i := 0; i < config.Rows; i++ {
		macro.Weights[i] = make([]float64, config.Cols)
	}

	// Initialize peripheral circuits
	if config.UseChargeDomain {
		macro.ChargeDomain = NewChargeDomainUnit(&ChargeDomainConfig{
			CapacitorValue:   1e-15, // 1 fF
			VoltageRange:     config.VDD,
			ChargeResolution: config.ADCBits,
			NumCapacitors:    config.Rows,
			UseC2CLadder:     true,
		})
	}

	if config.UseTIA {
		macro.TIA = NewTIA(&TIAConfig{
			FeedbackResistance:   1e6,   // 1 MOhm
			Bandwidth:            1e8,   // 100 MHz
			InputCapacitance:     1e-12, // 1 pF
			NoiseCurrentDensity:  1e-12, // 1 pA/sqrt(Hz)
			GainBandwidthProduct: 1e9,   // 1 GHz
		})
	}

	if config.UseConfigurableSA {
		macro.SenseAmps = make([]*ConfigurableSA, config.Cols)
		for i := 0; i < config.Cols; i++ {
			macro.SenseAmps[i] = NewConfigurableSA(&ConfigurableSAConfig{
				Mode:       SAModeADC,
				Resolution: config.ADCBits,
				Technology: config.Technology,
			})
		}
	}

	macro.calculateMetrics()
	return macro
}

// calculateMetrics calculates performance metrics
func (macro *MixedSignalCIMMacro) calculateMetrics() {
	techScale := float64(macro.Config.Technology) / 28.0

	// Array area
	var cellArea float64
	switch macro.Config.CellType {
	case "SRAM":
		cellArea = 0.5e-6 * techScale * techScale // 0.5 μm² at 28nm
	case "RRAM":
		cellArea = 0.1e-6 * techScale * techScale // 0.1 μm²
	case "FeFET":
		cellArea = 0.2e-6 * techScale * techScale // 0.2 μm²
	default:
		cellArea = 0.5e-6 * techScale * techScale
	}
	macro.ArrayArea = cellArea * float64(macro.Config.Rows*macro.Config.Cols)

	// Peripheral area (typically 30-50% of total)
	dacArea := 0.001 * float64(macro.Config.Rows) * techScale * techScale // mm²
	adcArea := 0.0002 * math.Pow(2, float64(macro.Config.ADCBits-6)) * float64(macro.Config.Cols) * techScale * techScale

	macro.PeripheralArea = dacArea + adcArea
	if macro.ChargeDomain != nil {
		macro.PeripheralArea += 0.001 * techScale * techScale
	}
	if macro.TIA != nil {
		macro.PeripheralArea += 0.01 * float64(macro.Config.Cols) * techScale * techScale
	}

	macro.TotalArea = macro.ArrayArea + macro.PeripheralArea

	// Energy per MAC
	// MAC energy = array + DAC + ADC
	arrayEnergy := 0.5 // fJ for crossbar MAC
	dacEnergy := 5.0 * float64(macro.Config.DACBits) / float64(macro.Config.Rows) // Amortized
	adcEnergy := 5.0 * float64(macro.Config.ADCBits) / float64(macro.Config.Rows) // Amortized

	macro.EnergyPerMAC = arrayEnergy + dacEnergy + adcEnergy

	// Throughput (MACs per second)
	// Assume 100 MHz operation
	clockRate := 100e6 // Hz
	macsPerCycle := float64(macro.Config.Rows * macro.Config.Cols)
	macro.Throughput = clockRate * macsPerCycle / 1e12 // TOPS

	// Energy efficiency
	macro.EnergyEff = 1000 / macro.EnergyPerMAC // TOPS/W

	// Area efficiency
	macro.AreaEff = macro.Throughput / macro.TotalArea // TOPS/mm²
}

// ComputeMVM performs matrix-vector multiplication
func (macro *MixedSignalCIMMacro) ComputeMVM(input []float64) []float64 {
	output := make([]float64, macro.Config.Cols)

	for j := 0; j < macro.Config.Cols; j++ {
		if macro.Config.UseChargeDomain && macro.ChargeDomain != nil {
			// Charge domain computation
			weights := make([]float64, macro.Config.Rows)
			for i := 0; i < macro.Config.Rows; i++ {
				weights[i] = macro.Weights[i][j]
			}
			// Use first input element as representative
			inp := 0.0
			if len(input) > 0 {
				inp = input[0]
			}
			output[j] = macro.ChargeDomain.ComputeMAC(weights, inp)
		} else {
			// Current-domain computation
			current := 0.0
			for i := 0; i < macro.Config.Rows && i < len(input); i++ {
				current += macro.Weights[i][j] * input[i]
			}

			if macro.Config.UseTIA && macro.TIA != nil {
				// Convert current to voltage via TIA
				output[j] = macro.TIA.Convert(current * 1e-6) // Scale to uA
			} else {
				output[j] = current
			}
		}

		// ADC conversion
		if macro.Config.UseConfigurableSA && macro.SenseAmps != nil && j < len(macro.SenseAmps) {
			quantized := macro.SenseAmps[j].ProcessADC(output[j], 1.0)
			output[j] = float64(quantized) / math.Pow(2, float64(macro.Config.ADCBits))
		}
	}

	return output
}

// GetMetrics returns comprehensive metrics
func (macro *MixedSignalCIMMacro) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_area_mm2":     macro.TotalArea,
		"array_area_mm2":     macro.ArrayArea,
		"peripheral_area_mm2": macro.PeripheralArea,
		"energy_per_mac_fJ":  macro.EnergyPerMAC,
		"throughput_TOPS":    macro.Throughput,
		"energy_eff_TOPS_W":  macro.EnergyEff,
		"area_eff_TOPS_mm2":  macro.AreaEff,
		"array_size":         []int{macro.Config.Rows, macro.Config.Cols},
		"technology_nm":      macro.Config.Technology,
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func sign(x float64) float64 {
	if x > 0 {
		return 1.0
	} else if x < 0 {
		return -1.0
	}
	return 0.0
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// =============================================================================
// INTEGRATED SECURE CIM SYSTEM
// =============================================================================

// SecureCIMConfig configures a secure CIM system
type SecureCIMConfig struct {
	// CIM configuration
	MacroConfig     *MixedSignalCIMConfig

	// Security configuration
	NoiseDefense    *CIMNoiseDefenseConfig
	WeightProtection *WeightProtectionConfig

	// Attack modeling
	ExpectedAttacks []AdversarialAttackType
	SideChannelRisk SideChannelAttackType
}

// SecureCIMSystem implements an integrated secure CIM system
type SecureCIMSystem struct {
	Config          *SecureCIMConfig
	Macro           *MixedSignalCIMMacro
	NoiseDefender   *CIMNoiseDefender
	WeightProtector *WeightProtector

	// Security metrics
	AdversarialRobustness float64
	SideChannelResistance float64
	OverallSecurityScore  float64
}

// NewSecureCIMSystem creates a new secure CIM system
func NewSecureCIMSystem(config *SecureCIMConfig) *SecureCIMSystem {
	system := &SecureCIMSystem{
		Config: config,
		Macro:  NewMixedSignalCIMMacro(config.MacroConfig),
	}

	if config.NoiseDefense != nil {
		system.NoiseDefender = NewCIMNoiseDefender(config.NoiseDefense)
	}

	if config.WeightProtection != nil {
		system.WeightProtector = NewWeightProtector(config.WeightProtection)
	}

	system.calculateSecurityScore()
	return system
}

// calculateSecurityScore calculates overall security score
func (s *SecureCIMSystem) calculateSecurityScore() {
	// Adversarial robustness (0-1)
	if s.NoiseDefender != nil {
		// Noise-based defense provides ~50% improvement
		s.AdversarialRobustness = 0.7
	} else {
		s.AdversarialRobustness = 0.3
	}

	// Side-channel resistance (0-1)
	if s.WeightProtector != nil {
		switch s.WeightProtector.Config.Type {
		case ProtectionObfuscation:
			s.SideChannelResistance = 0.6
		case ProtectionEncryption:
			s.SideChannelResistance = 0.9
		case ProtectionSplitting:
			s.SideChannelResistance = 0.8
		case ProtectionNoise:
			s.SideChannelResistance = 0.5
		default:
			s.SideChannelResistance = 0.2
		}
	} else {
		s.SideChannelResistance = 0.2
	}

	// Overall score (geometric mean)
	s.OverallSecurityScore = math.Sqrt(s.AdversarialRobustness * s.SideChannelResistance)
}

// SecureCompute performs secure computation with defenses
func (s *SecureCIMSystem) SecureCompute(input []float64) []float64 {
	// Apply input defense
	defendedInput := input
	if s.NoiseDefender != nil {
		defendedInput = s.NoiseDefender.ApplyDefense(input)
	}

	// Compute MVM
	output := s.Macro.ComputeMVM(defendedInput)

	// Apply output defense
	if s.NoiseDefender != nil {
		output = s.NoiseDefender.ApplyDefense(output)
	}

	return output
}

// GetSecurityMetrics returns security metrics
func (s *SecureCIMSystem) GetSecurityMetrics() map[string]interface{} {
	metrics := s.Macro.GetMetrics()
	metrics["adversarial_robustness"] = s.AdversarialRobustness
	metrics["side_channel_resistance"] = s.SideChannelResistance
	metrics["overall_security_score"] = s.OverallSecurityScore

	if s.NoiseDefender != nil {
		metrics["total_inferences"] = s.NoiseDefender.TotalInferences
		metrics["noise_std"] = s.NoiseDefender.Config.InjectedNoiseStd
	}

	if s.WeightProtector != nil {
		metrics["protection_type"] = s.WeightProtector.Config.Type
		metrics["protection_refreshes"] = s.WeightProtector.InferenceCount / s.WeightProtector.Config.RefreshInterval
	}

	return metrics
}
