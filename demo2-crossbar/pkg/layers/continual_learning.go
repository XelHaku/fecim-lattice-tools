// Package layers provides neural network layer implementations for CIM deployment.
// This file implements continual learning algorithms for ferroelectric CIM systems.
// Based on research: Nature Electronics 2025 (ferroelectric-memristor), Nature Communications 2025 (FCM)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// ELASTIC WEIGHT CONSOLIDATION (EWC)
// Reduces catastrophic forgetting by 45.7% vs naive sequential training
// =============================================================================

// EWCConfig configures Elastic Weight Consolidation parameters.
type EWCConfig struct {
	// Lambda controls importance of previous task weights (typically 100-10000)
	Lambda float64

	// FisherSamples is number of samples for Fisher Information estimation
	FisherSamples int

	// OnlineEWC enables online EWC with running Fisher accumulation
	OnlineEWC bool

	// Gamma is decay factor for online EWC (typically 0.9-0.99)
	Gamma float64

	// NormalizeFisher normalizes Fisher information across layers
	NormalizeFisher bool

	// CIMQuantization applies CIM-aware quantization to Fisher
	CIMQuantization int
}

// DefaultEWCConfig returns default EWC configuration.
func DefaultEWCConfig() *EWCConfig {
	return &EWCConfig{
		Lambda:          400.0,
		FisherSamples:   200,
		OnlineEWC:       true,
		Gamma:           0.95,
		NormalizeFisher: true,
		CIMQuantization: 6,
	}
}

// EWCRegularizer implements Elastic Weight Consolidation for CIM.
type EWCRegularizer struct {
	Config *EWCConfig

	// Per-task stored quantities
	OptimalWeights    [][][]float64        // theta*_t for each task
	FisherInformation [][][]float64        // F_t for each task
	TaskCount         int                  // Number of tasks seen

	// Online EWC accumulated quantities
	OnlineFisher      [][]float64
	OnlineOptimal     [][]float64
}

// NewEWCRegularizer creates a new EWC regularizer.
func NewEWCRegularizer(config *EWCConfig) *EWCRegularizer {
	if config == nil {
		config = DefaultEWCConfig()
	}
	return &EWCRegularizer{
		Config:            config,
		OptimalWeights:    make([][][]float64, 0),
		FisherInformation: make([][][]float64, 0),
	}
}

// ComputeFisher computes Fisher Information Matrix from gradients.
func (e *EWCRegularizer) ComputeFisher(weights [][]float64, gradientFn func([][]float64) [][]float64) [][]float64 {
	rows, cols := len(weights), len(weights[0])
	fisher := make([][]float64, rows)
	for i := range fisher {
		fisher[i] = make([]float64, cols)
	}

	// Accumulate squared gradients over samples
	for sample := 0; sample < e.Config.FisherSamples; sample++ {
		grads := gradientFn(weights)
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				fisher[i][j] += grads[i][j] * grads[i][j]
			}
		}
	}

	// Average and optionally normalize
	scale := 1.0 / float64(e.Config.FisherSamples)
	maxFisher := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			fisher[i][j] *= scale
			if fisher[i][j] > maxFisher {
				maxFisher = fisher[i][j]
			}
		}
	}

	// Normalize if enabled
	if e.Config.NormalizeFisher && maxFisher > 0 {
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				fisher[i][j] /= maxFisher
			}
		}
	}

	// CIM quantization if enabled
	if e.Config.CIMQuantization > 0 {
		levels := float64(int(1) << e.Config.CIMQuantization)
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				fisher[i][j] = math.Round(fisher[i][j]*levels) / levels
			}
		}
	}

	return fisher
}

// RegisterTask records optimal weights and Fisher after task completion.
func (e *EWCRegularizer) RegisterTask(weights [][]float64, fisher [][]float64) {
	rows, cols := len(weights), len(weights[0])

	// Deep copy weights
	optWeights := make([][]float64, rows)
	for i := range optWeights {
		optWeights[i] = make([]float64, cols)
		copy(optWeights[i], weights[i])
	}

	// Deep copy Fisher
	fisherCopy := make([][]float64, rows)
	for i := range fisherCopy {
		fisherCopy[i] = make([]float64, cols)
		copy(fisherCopy[i], fisher[i])
	}

	if e.Config.OnlineEWC {
		// Online EWC: accumulate with decay
		if e.OnlineFisher == nil {
			e.OnlineFisher = fisherCopy
			e.OnlineOptimal = optWeights
		} else {
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					// Decay old Fisher and add new
					e.OnlineFisher[i][j] = e.Config.Gamma*e.OnlineFisher[i][j] + fisher[i][j]
					// Update optimal with weighted average
					e.OnlineOptimal[i][j] = e.Config.Gamma*e.OnlineOptimal[i][j] +
						(1-e.Config.Gamma)*weights[i][j]
				}
			}
		}
	} else {
		// Standard EWC: store per-task
		e.OptimalWeights = append(e.OptimalWeights, optWeights)
		e.FisherInformation = append(e.FisherInformation, fisherCopy)
	}

	e.TaskCount++
}

// ComputePenalty computes EWC regularization penalty.
func (e *EWCRegularizer) ComputePenalty(weights [][]float64) float64 {
	rows, cols := len(weights), len(weights[0])
	penalty := 0.0

	if e.Config.OnlineEWC && e.OnlineFisher != nil {
		// Online EWC penalty
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				diff := weights[i][j] - e.OnlineOptimal[i][j]
				penalty += e.OnlineFisher[i][j] * diff * diff
			}
		}
	} else {
		// Standard EWC: sum over all tasks
		for t := 0; t < len(e.OptimalWeights); t++ {
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					diff := weights[i][j] - e.OptimalWeights[t][i][j]
					penalty += e.FisherInformation[t][i][j] * diff * diff
				}
			}
		}
	}

	return 0.5 * e.Config.Lambda * penalty
}

// ComputeGradient computes EWC gradient contribution.
func (e *EWCRegularizer) ComputeGradient(weights [][]float64) [][]float64 {
	rows, cols := len(weights), len(weights[0])
	grad := make([][]float64, rows)
	for i := range grad {
		grad[i] = make([]float64, cols)
	}

	if e.Config.OnlineEWC && e.OnlineFisher != nil {
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				diff := weights[i][j] - e.OnlineOptimal[i][j]
				grad[i][j] = e.Config.Lambda * e.OnlineFisher[i][j] * diff
			}
		}
	} else {
		for t := 0; t < len(e.OptimalWeights); t++ {
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					diff := weights[i][j] - e.OptimalWeights[t][i][j]
					grad[i][j] += e.Config.Lambda * e.FisherInformation[t][i][j] * diff
				}
			}
		}
	}

	return grad
}

// =============================================================================
// SYNAPTIC INTELLIGENCE (SI)
// Online importance estimation without explicit Fisher computation
// =============================================================================

// SIConfig configures Synaptic Intelligence parameters.
type SIConfig struct {
	// C controls regularization strength
	C float64

	// Xi is damping parameter to prevent division by zero
	Xi float64

	// CIMEnabled enables CIM-aware modifications
	CIMEnabled bool
}

// SynapticIntelligence implements online continual learning.
type SynapticIntelligence struct {
	Config *SIConfig

	// Running weight importance
	Omega [][]float64

	// Path integral accumulator
	PathIntegral [][]float64

	// Previous weights and gradients
	PrevWeights [][]float64
	PrevLoss    float64

	// CIM conductance bounds for importance weighting
	ConductanceMin float64
	ConductanceMax float64
}

// NewSynapticIntelligence creates a new SI regularizer.
func NewSynapticIntelligence(rows, cols int, config *SIConfig) *SynapticIntelligence {
	if config == nil {
		config = &SIConfig{C: 0.1, Xi: 1e-3, CIMEnabled: true}
	}

	omega := make([][]float64, rows)
	path := make([][]float64, rows)
	prev := make([][]float64, rows)
	for i := range omega {
		omega[i] = make([]float64, cols)
		path[i] = make([]float64, cols)
		prev[i] = make([]float64, cols)
	}

	return &SynapticIntelligence{
		Config:         config,
		Omega:          omega,
		PathIntegral:   path,
		PrevWeights:    prev,
		ConductanceMin: 1e-9,
		ConductanceMax: 1e-6,
	}
}

// UpdatePathIntegral updates the path integral during training.
func (si *SynapticIntelligence) UpdatePathIntegral(weights [][]float64, gradients [][]float64, loss float64) {
	rows, cols := len(weights), len(weights[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Path integral: sum of gradient * weight change
			deltaW := weights[i][j] - si.PrevWeights[i][j]
			si.PathIntegral[i][j] += -gradients[i][j] * deltaW

			// Update previous weights
			si.PrevWeights[i][j] = weights[i][j]
		}
	}
	si.PrevLoss = loss
}

// ConsolidateTask updates omega after task completion.
func (si *SynapticIntelligence) ConsolidateTask(finalWeights, initialWeights [][]float64) {
	rows, cols := len(finalWeights), len(finalWeights[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			deltaW := finalWeights[i][j] - initialWeights[i][j]
			denominator := deltaW*deltaW + si.Config.Xi

			// Update importance
			si.Omega[i][j] += si.PathIntegral[i][j] / denominator

			// CIM-aware clipping based on conductance bounds
			if si.Config.CIMEnabled {
				maxOmega := (si.ConductanceMax - si.ConductanceMin) / si.Config.Xi
				if si.Omega[i][j] > maxOmega {
					si.Omega[i][j] = maxOmega
				}
			}

			// Reset path integral
			si.PathIntegral[i][j] = 0
		}
	}
}

// ComputePenalty computes SI regularization penalty.
func (si *SynapticIntelligence) ComputePenalty(weights, referenceWeights [][]float64) float64 {
	penalty := 0.0
	rows, cols := len(weights), len(weights[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			diff := weights[i][j] - referenceWeights[i][j]
			penalty += si.Omega[i][j] * diff * diff
		}
	}

	return si.Config.C * penalty
}

// =============================================================================
// MEMORY AWARE SYNAPSES (MAS)
// Importance based on output sensitivity
// =============================================================================

// MASConfig configures Memory Aware Synapses.
type MASConfig struct {
	Lambda     float64
	NumSamples int
}

// MemoryAwareSynapses implements MAS for CIM.
type MemoryAwareSynapses struct {
	Config     *MASConfig
	Importance [][]float64
	RefWeights [][]float64
}

// NewMemoryAwareSynapses creates a new MAS regularizer.
func NewMemoryAwareSynapses(rows, cols int, config *MASConfig) *MemoryAwareSynapses {
	if config == nil {
		config = &MASConfig{Lambda: 1.0, NumSamples: 100}
	}

	imp := make([][]float64, rows)
	ref := make([][]float64, rows)
	for i := range imp {
		imp[i] = make([]float64, cols)
		ref[i] = make([]float64, cols)
	}

	return &MemoryAwareSynapses{
		Config:     config,
		Importance: imp,
		RefWeights: ref,
	}
}

// ComputeImportance estimates weight importance from output gradients.
func (mas *MemoryAwareSynapses) ComputeImportance(weights [][]float64, outputGradFn func([][]float64) [][]float64) {
	rows, cols := len(weights), len(weights[0])

	// Reset importance
	for i := range mas.Importance {
		for j := range mas.Importance[i] {
			mas.Importance[i][j] = 0
		}
	}

	// Accumulate absolute gradients
	for s := 0; s < mas.Config.NumSamples; s++ {
		grads := outputGradFn(weights)
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				mas.Importance[i][j] += math.Abs(grads[i][j])
			}
		}
	}

	// Normalize
	scale := 1.0 / float64(mas.Config.NumSamples)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			mas.Importance[i][j] *= scale
		}
	}

	// Store reference weights
	for i := 0; i < rows; i++ {
		copy(mas.RefWeights[i], weights[i])
	}
}

// ComputePenalty computes MAS regularization loss.
func (mas *MemoryAwareSynapses) ComputePenalty(weights [][]float64) float64 {
	penalty := 0.0
	rows, cols := len(weights), len(weights[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			diff := weights[i][j] - mas.RefWeights[i][j]
			penalty += mas.Importance[i][j] * diff * diff
		}
	}

	return 0.5 * mas.Config.Lambda * penalty
}

// =============================================================================
// METAPLASTICITY CONTROLLER
// Brain-inspired variable synaptic plasticity for CIM
// =============================================================================

// MetaplasticityMode defines the plasticity regulation mode.
type MetaplasticityMode int

const (
	MetaplasticityBCM MetaplasticityMode = iota // Bienenstock-Cooper-Munro
	MetaplasticityOja                           // Oja's rule
	MetaplasticitySliding                       // Sliding threshold
)

// MetaplasticityConfig configures metaplasticity controller.
type MetaplasticityConfig struct {
	Mode MetaplasticityMode

	// BCM parameters
	BCMTheta0     float64 // Initial modification threshold
	BCMTau        float64 // Time constant for threshold adaptation
	BCMTimeConstant float64

	// Sliding window parameters
	WindowSize    int
	TargetRate    float64

	// CIM-specific parameters
	ConductanceDependent bool
	MinLearningRate      float64
	MaxLearningRate      float64
}

// MetaplasticityController implements variable synaptic plasticity.
type MetaplasticityController struct {
	Config *MetaplasticityConfig

	// Per-synapse modification thresholds
	Thresholds [][]float64

	// Activity history for sliding threshold
	ActivityHistory [][][]float64
	HistoryIdx      int

	// Current learning rate modulation
	LearningRates [][]float64
}

// NewMetaplasticityController creates a metaplasticity controller.
func NewMetaplasticityController(rows, cols int, config *MetaplasticityConfig) *MetaplasticityController {
	if config == nil {
		config = &MetaplasticityConfig{
			Mode:                 MetaplasticityBCM,
			BCMTheta0:            0.5,
			BCMTau:               100.0,
			WindowSize:           100,
			TargetRate:           0.1,
			ConductanceDependent: true,
			MinLearningRate:      0.001,
			MaxLearningRate:      0.1,
		}
	}

	thresholds := make([][]float64, rows)
	lr := make([][]float64, rows)
	for i := range thresholds {
		thresholds[i] = make([]float64, cols)
		lr[i] = make([]float64, cols)
		for j := range thresholds[i] {
			thresholds[i][j] = config.BCMTheta0
			lr[i][j] = (config.MinLearningRate + config.MaxLearningRate) / 2
		}
	}

	history := make([][][]float64, config.WindowSize)
	for t := range history {
		history[t] = make([][]float64, rows)
		for i := range history[t] {
			history[t][i] = make([]float64, cols)
		}
	}

	return &MetaplasticityController{
		Config:          config,
		Thresholds:      thresholds,
		ActivityHistory: history,
		LearningRates:   lr,
	}
}

// UpdateBCM updates thresholds using BCM rule.
func (mc *MetaplasticityController) UpdateBCM(activity [][]float64) {
	rows, cols := len(activity), len(activity[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// BCM threshold tracks squared postsynaptic activity
			actSquared := activity[i][j] * activity[i][j]
			mc.Thresholds[i][j] += (actSquared - mc.Thresholds[i][j]) / mc.Config.BCMTau
		}
	}
}

// ComputePlasticity computes BCM plasticity modification.
func (mc *MetaplasticityController) ComputePlasticity(preActivity, postActivity [][]float64) [][]float64 {
	rows, cols := len(preActivity), len(preActivity[0])
	plasticity := make([][]float64, rows)

	for i := range plasticity {
		plasticity[i] = make([]float64, cols)
		for j := range plasticity[i] {
			// BCM rule: phi(y) = y(y - theta)
			post := postActivity[i][j]
			phi := post * (post - mc.Thresholds[i][j])
			plasticity[i][j] = phi * preActivity[i][j]
		}
	}

	return plasticity
}

// ModulateLearningRate adjusts learning rates based on synapse state.
func (mc *MetaplasticityController) ModulateLearningRate(conductances [][]float64) {
	rows, cols := len(conductances), len(conductances[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Higher conductance -> lower plasticity (stable synapse)
			// Lower conductance -> higher plasticity (exploratory synapse)
			g := conductances[i][j]

			if mc.Config.ConductanceDependent {
				// Inverse relationship with conductance
				mc.LearningRates[i][j] = mc.Config.MaxLearningRate -
					(mc.Config.MaxLearningRate-mc.Config.MinLearningRate)*g
				// Clamp to valid range
				if mc.LearningRates[i][j] < mc.Config.MinLearningRate {
					mc.LearningRates[i][j] = mc.Config.MinLearningRate
				}
				if mc.LearningRates[i][j] > mc.Config.MaxLearningRate {
					mc.LearningRates[i][j] = mc.Config.MaxLearningRate
				}
			}
		}
	}
}

// =============================================================================
// NEGATIVE CAPACITANCE CIM (NC-CIM)
// Sub-60 mV/decade switching for ultra-low power CIM
// =============================================================================

// NCCIMConfig configures NC-CIM device simulation.
type NCCIMConfig struct {
	// NC-FET parameters
	SubthresholdSwing float64 // mV/decade (< 60 for NC)
	NCGain            float64 // Internal NC voltage amplification

	// Ferroelectric parameters
	FEThickness     float64 // nm
	CoerciveField   float64 // MV/cm
	Polarization    float64 // μC/cm²

	// Dielectric capacitance
	CDielectric     float64 // fF/μm²
	CFerroelectric  float64 // fF/μm²

	// Operating conditions
	OperatingVoltage float64 // V
	Temperature      float64 // K
}

// DefaultNCCIMConfig returns default NC-CIM configuration.
func DefaultNCCIMConfig() *NCCIMConfig {
	return &NCCIMConfig{
		SubthresholdSwing: 45.0,  // Sub-60 mV/dec NC
		NCGain:            1.5,
		FEThickness:       10.0,
		CoerciveField:     1.0,
		Polarization:      15.0,
		CDielectric:       10.0,
		CFerroelectric:    8.0,
		OperatingVoltage:  0.5,
		Temperature:       300.0,
	}
}

// NCCIMDevice simulates negative capacitance CIM device.
type NCCIMDevice struct {
	Config       *NCCIMConfig
	Polarization float64
	Conductance  float64
	NCState      float64 // Internal NC amplified voltage
}

// NewNCCIMDevice creates a new NC-CIM device.
func NewNCCIMDevice(config *NCCIMConfig) *NCCIMDevice {
	if config == nil {
		config = DefaultNCCIMConfig()
	}
	return &NCCIMDevice{
		Config:      config,
		Conductance: 0.5,
	}
}

// ComputeNCGain computes the negative capacitance voltage gain.
func (d *NCCIMDevice) ComputeNCGain(vg float64) float64 {
	// NC gain: Av = 1 + |CFE|/Cox
	// In negative capacitance regime, CFE < 0
	cRatio := d.Config.CFerroelectric / d.Config.CDielectric
	gain := 1.0 + cRatio

	// Internal voltage is amplified
	return vg * gain
}

// SubthresholdCurrent computes the sub-threshold current with NC boost.
func (d *NCCIMDevice) SubthresholdCurrent(vg, vd float64) float64 {
	// Boltzmann thermal voltage
	kT := 0.0259 * d.Config.Temperature / 300.0

	// NC-enhanced gate voltage
	vgNC := d.ComputeNCGain(vg)

	// Sub-threshold current with reduced swing
	ssNC := d.Config.SubthresholdSwing / 1000.0 // Convert to V
	i0 := 1e-9                                   // Reference current

	return i0 * math.Exp(vgNC/(ssNC/math.Log(10))) * (1 - math.Exp(-vd/kT))
}

// ProgramConductance programs NC-CIM device with reduced voltage.
func (d *NCCIMDevice) ProgramConductance(targetG, pulseV float64) float64 {
	// NC amplification enables lower programming voltage
	effectiveV := d.ComputeNCGain(pulseV)

	// Ferroelectric switching
	ec := d.Config.CoerciveField * d.Config.FEThickness * 1e-7 // MV/cm * nm -> V
	if math.Abs(effectiveV) > ec {
		// Polarization switching
		sign := 1.0
		if effectiveV < 0 {
			sign = -1.0
		}
		d.Polarization = sign * d.Config.Polarization

		// Update conductance based on polarization
		d.Conductance = (d.Polarization/d.Config.Polarization + 1) / 2
	}

	return d.Conductance
}

// =============================================================================
// UNIFIED FERROELECTRIC-MEMRISTOR DEVICE
// Based on Nature Electronics 2025 unified memory for training and inference
// =============================================================================

// UnifiedFEMemConfig configures unified FE-memristor device.
type UnifiedFEMemConfig struct {
	// Device geometry
	StackThickness float64 // nm (HfO2 + Ti scavenging layer)

	// Ferroelectric mode parameters
	FEEndurance      int64   // Cycles (typically 10^10+)
	FEProgramEnergy  float64 // fJ/bit
	FESwitchingTime  float64 // ns

	// Memristor mode parameters
	MemEndurance     int64   // Cycles (typically 10^6-10^8)
	MemProgramEnergy float64 // pJ/bit
	MemReadEnergy    float64 // fJ/bit

	// State levels
	NumFEStates  int
	NumMemStates int
}

// DefaultUnifiedConfig returns default unified device config.
func DefaultUnifiedConfig() *UnifiedFEMemConfig {
	return &UnifiedFEMemConfig{
		StackThickness:   15.0,
		FEEndurance:      1e12,
		FEProgramEnergy:  10.0,
		FESwitchingTime:  10.0,
		MemEndurance:     1e7,
		MemProgramEnergy: 100.0,
		MemReadEnergy:    1.0,
		NumFEStates:      64,
		NumMemStates:     16,
	}
}

// UnifiedFEMemDevice implements unified ferroelectric-memristor device.
type UnifiedFEMemDevice struct {
	Config *UnifiedFEMemConfig

	// Current state
	FEState     int     // Ferroelectric polarization state
	MemState    int     // Memristor conductance state
	Conductance float64 // Effective conductance

	// Mode tracking
	IsTrainingMode  bool
	WriteCount      int64
	ReadCount       int64
}

// NewUnifiedFEMemDevice creates a unified FE-memristor device.
func NewUnifiedFEMemDevice(config *UnifiedFEMemConfig) *UnifiedFEMemDevice {
	if config == nil {
		config = DefaultUnifiedConfig()
	}
	return &UnifiedFEMemDevice{
		Config:         config,
		IsTrainingMode: true,
	}
}

// SetTrainingMode switches to ferroelectric mode for learning.
func (d *UnifiedFEMemDevice) SetTrainingMode() {
	d.IsTrainingMode = true
}

// SetInferenceMode switches to memristor mode for inference.
func (d *UnifiedFEMemDevice) SetInferenceMode() {
	d.IsTrainingMode = false
	// Transfer FE state to memristor
	d.MemState = d.FEState * d.Config.NumMemStates / d.Config.NumFEStates
	d.updateConductance()
}

// Write programs the device (FE mode for training, memristor for inference).
func (d *UnifiedFEMemDevice) Write(targetState int) {
	if d.IsTrainingMode {
		// Ferroelectric capacitor mode - high endurance, faster
		d.FEState = targetState % d.Config.NumFEStates
		d.WriteCount++
	} else {
		// Memristor mode
		d.MemState = targetState % d.Config.NumMemStates
		d.WriteCount++
	}
	d.updateConductance()
}

// Read returns the device state (non-destructive in memristor mode).
func (d *UnifiedFEMemDevice) Read() float64 {
	d.ReadCount++
	return d.Conductance
}

// updateConductance computes effective conductance from state.
func (d *UnifiedFEMemDevice) updateConductance() {
	if d.IsTrainingMode {
		d.Conductance = float64(d.FEState) / float64(d.Config.NumFEStates-1)
	} else {
		d.Conductance = float64(d.MemState) / float64(d.Config.NumMemStates-1)
	}
}

// GetEnergyCost returns estimated energy for the operation.
func (d *UnifiedFEMemDevice) GetEnergyCost(isWrite bool) float64 {
	if isWrite {
		if d.IsTrainingMode {
			return d.Config.FEProgramEnergy
		}
		return d.Config.MemProgramEnergy
	}
	return d.Config.MemReadEnergy
}

// =============================================================================
// FCM MEMRISTOR (Filament Conductivity Modulation)
// Based on Nature Communications 2025 electrochemical ohmic memristors
// =============================================================================

// FCMConfig configures FCM memristor parameters.
type FCMConfig struct {
	// Filament parameters
	FilamentType     string  // "oxide" (TaOx) vs "metallic" (ECM)
	NumStates        int
	OnOffRatio       float64

	// Electrochemical parameters
	ActivationEnergy float64 // eV
	IonMobility      float64 // cm²/V·s

	// Metaplasticity support
	EnableMetaplasticity bool
	PlasticityRange      float64 // Range of plasticity modulation
}

// DefaultFCMConfig returns default FCM memristor configuration.
func DefaultFCMConfig() *FCMConfig {
	return &FCMConfig{
		FilamentType:         "oxide",
		NumStates:            128,
		OnOffRatio:           1e4,
		ActivationEnergy:     0.5,
		IonMobility:          1e-10,
		EnableMetaplasticity: true,
		PlasticityRange:      10.0,
	}
}

// FCMMemristor implements filament conductivity modulation memristor.
type FCMMemristor struct {
	Config *FCMConfig

	// Filament state
	FilamentRadius     float64 // Relative filament radius
	ConductanceState   int
	Conductance        float64

	// Metaplasticity state
	PlasticityFactor   float64 // Current plasticity level (0-1)
	WriteHistory       []float64

	// Temperature and stability
	Temperature        float64
}

// NewFCMMemristor creates a new FCM memristor.
func NewFCMMemristor(config *FCMConfig) *FCMMemristor {
	if config == nil {
		config = DefaultFCMConfig()
	}
	return &FCMMemristor{
		Config:           config,
		FilamentRadius:   0.5,
		PlasticityFactor: 0.5,
		WriteHistory:     make([]float64, 0),
		Temperature:      300.0,
	}
}

// SetState programs the FCM memristor to a specific state.
func (m *FCMMemristor) SetState(state int) {
	m.ConductanceState = state % m.Config.NumStates

	// Oxide filament: conductivity changes via oxygen redistribution
	m.FilamentRadius = float64(m.ConductanceState) / float64(m.Config.NumStates-1)

	// Compute conductance with metaplasticity modulation
	gMin := 1.0 / m.Config.OnOffRatio
	m.Conductance = gMin + (1.0-gMin)*m.FilamentRadius

	// Record write for metaplasticity
	m.WriteHistory = append(m.WriteHistory, float64(state))
}

// IncrementalUpdate applies incremental conductance change.
func (m *FCMMemristor) IncrementalUpdate(deltaG float64) {
	// Metaplasticity-aware update
	effectiveDelta := deltaG * m.PlasticityFactor * m.Config.PlasticityRange

	// Apply to filament radius
	m.FilamentRadius += effectiveDelta
	if m.FilamentRadius < 0 {
		m.FilamentRadius = 0
	}
	if m.FilamentRadius > 1 {
		m.FilamentRadius = 1
	}

	// Update conductance
	gMin := 1.0 / m.Config.OnOffRatio
	m.Conductance = gMin + (1.0-gMin)*m.FilamentRadius

	// Update state index
	m.ConductanceState = int(m.FilamentRadius * float64(m.Config.NumStates-1))
}

// UpdatePlasticity adjusts plasticity based on recent activity.
func (m *FCMMemristor) UpdatePlasticity() {
	if !m.Config.EnableMetaplasticity || len(m.WriteHistory) < 2 {
		return
	}

	// Compute variance of recent writes
	n := len(m.WriteHistory)
	windowSize := 10
	if n < windowSize {
		windowSize = n
	}

	var sum, sumSq float64
	for i := n - windowSize; i < n; i++ {
		sum += m.WriteHistory[i]
		sumSq += m.WriteHistory[i] * m.WriteHistory[i]
	}
	mean := sum / float64(windowSize)
	variance := sumSq/float64(windowSize) - mean*mean

	// High variance -> high plasticity (exploring)
	// Low variance -> low plasticity (consolidating)
	m.PlasticityFactor = 1.0 / (1.0 + math.Exp(-variance+2.0))
}

// =============================================================================
// CONTINUAL LEARNING CROSSBAR
// Full CIM array with continual learning support
// =============================================================================

// ContinualLearningMode defines the CL algorithm to use.
type ContinualLearningMode int

const (
	CLModeEWC ContinualLearningMode = iota
	CLModeSI
	CLModeMAS
	CLModeMetaplastic
)

// ContinualLearningCrossbar implements CIM array with CL capabilities.
type ContinualLearningCrossbar struct {
	Rows, Cols int

	// Device arrays
	Devices [][]*UnifiedFEMemDevice

	// Continual learning components
	EWC           *EWCRegularizer
	SI            *SynapticIntelligence
	MAS           *MemoryAwareSynapses
	Metaplasticity *MetaplasticityController

	// Current CL mode
	Mode ContinualLearningMode

	// Task tracking
	CurrentTask int
	TaskNames   []string
}

// NewContinualLearningCrossbar creates a CL-enabled crossbar.
func NewContinualLearningCrossbar(rows, cols int, mode ContinualLearningMode) *ContinualLearningCrossbar {
	devices := make([][]*UnifiedFEMemDevice, rows)
	for i := range devices {
		devices[i] = make([]*UnifiedFEMemDevice, cols)
		for j := range devices[i] {
			devices[i][j] = NewUnifiedFEMemDevice(nil)
		}
	}

	clc := &ContinualLearningCrossbar{
		Rows:    rows,
		Cols:    cols,
		Devices: devices,
		Mode:    mode,
	}

	// Initialize CL component based on mode
	switch mode {
	case CLModeEWC:
		clc.EWC = NewEWCRegularizer(nil)
	case CLModeSI:
		clc.SI = NewSynapticIntelligence(rows, cols, nil)
	case CLModeMAS:
		clc.MAS = NewMemoryAwareSynapses(rows, cols, nil)
	case CLModeMetaplastic:
		clc.Metaplasticity = NewMetaplasticityController(rows, cols, nil)
	}

	return clc
}

// GetWeights extracts weights from device conductances.
func (clc *ContinualLearningCrossbar) GetWeights() [][]float64 {
	weights := make([][]float64, clc.Rows)
	for i := range weights {
		weights[i] = make([]float64, clc.Cols)
		for j := range weights[i] {
			weights[i][j] = clc.Devices[i][j].Read()
		}
	}
	return weights
}

// SetWeights programs weights to devices.
func (clc *ContinualLearningCrossbar) SetWeights(weights [][]float64) {
	for i := 0; i < clc.Rows; i++ {
		for j := 0; j < clc.Cols; j++ {
			state := int(weights[i][j] * float64(clc.Devices[i][j].Config.NumFEStates-1))
			clc.Devices[i][j].Write(state)
		}
	}
}

// Forward performs matrix-vector multiplication.
func (clc *ContinualLearningCrossbar) Forward(input []float64) []float64 {
	output := make([]float64, clc.Cols)

	for j := 0; j < clc.Cols; j++ {
		for i := 0; i < clc.Rows; i++ {
			g := clc.Devices[i][j].Read()
			output[j] += input[i] * g
		}
	}

	return output
}

// StartTask prepares for a new task.
func (clc *ContinualLearningCrossbar) StartTask(taskName string) {
	clc.TaskNames = append(clc.TaskNames, taskName)

	// Set all devices to training mode
	for i := 0; i < clc.Rows; i++ {
		for j := 0; j < clc.Cols; j++ {
			clc.Devices[i][j].SetTrainingMode()
		}
	}
}

// EndTask consolidates learning after task completion.
func (clc *ContinualLearningCrossbar) EndTask() {
	weights := clc.GetWeights()

	switch clc.Mode {
	case CLModeEWC:
		// Compute Fisher and register task
		fisher := clc.EWC.ComputeFisher(weights, func(w [][]float64) [][]float64 {
			// Placeholder gradient function - in real use, compute from loss
			grads := make([][]float64, len(w))
			for i := range grads {
				grads[i] = make([]float64, len(w[0]))
				for j := range grads[i] {
					grads[i][j] = rand.NormFloat64() * 0.1
				}
			}
			return grads
		})
		clc.EWC.RegisterTask(weights, fisher)

	case CLModeSI:
		// SI consolidation handled during training
		// Get initial weights from task start (would need to store)
		clc.SI.ConsolidateTask(weights, weights) // Simplified

	case CLModeMAS:
		// Compute importance from output gradients
		clc.MAS.ComputeImportance(weights, func(w [][]float64) [][]float64 {
			grads := make([][]float64, len(w))
			for i := range grads {
				grads[i] = make([]float64, len(w[0]))
				for j := range grads[i] {
					grads[i][j] = rand.NormFloat64() * 0.1
				}
			}
			return grads
		})
	}

	clc.CurrentTask++
}

// ComputeCLLoss computes the continual learning regularization loss.
func (clc *ContinualLearningCrossbar) ComputeCLLoss(taskLoss float64) float64 {
	weights := clc.GetWeights()
	clPenalty := 0.0

	switch clc.Mode {
	case CLModeEWC:
		clPenalty = clc.EWC.ComputePenalty(weights)
	case CLModeSI:
		clPenalty = clc.SI.ComputePenalty(weights, weights) // Simplified
	case CLModeMAS:
		clPenalty = clc.MAS.ComputePenalty(weights)
	}

	return taskLoss + clPenalty
}

// SwitchToInference prepares all devices for inference mode.
func (clc *ContinualLearningCrossbar) SwitchToInference() {
	for i := 0; i < clc.Rows; i++ {
		for j := 0; j < clc.Cols; j++ {
			clc.Devices[i][j].SetInferenceMode()
		}
	}
}

// =============================================================================
// PERFORMANCE METRICS
// =============================================================================

// ContinualLearningMetrics tracks CL performance.
type ContinualLearningMetrics struct {
	// Per-task accuracy
	TaskAccuracies     map[string]float64

	// Forgetting metrics
	BackwardTransfer   float64 // Negative = forgetting
	ForwardTransfer    float64 // Knowledge transfer to new tasks
	AverageAccuracy    float64

	// Hardware metrics
	TotalWrites        int64
	TotalReads         int64
	TotalEnergy        float64 // fJ
}

// ComputeForgetting calculates catastrophic forgetting metric.
func ComputeForgetting(accuracyAfterTask, accuracyFinal map[string]float64) float64 {
	if len(accuracyAfterTask) == 0 {
		return 0
	}

	forgetting := 0.0
	count := 0
	for task, accAfter := range accuracyAfterTask {
		if accFinal, ok := accuracyFinal[task]; ok {
			forgetting += accAfter - accFinal
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return forgetting / float64(count)
}

// ComputeForwardTransfer calculates forward transfer metric.
func ComputeForwardTransfer(randomInit, afterTraining map[string]float64) float64 {
	if len(randomInit) == 0 {
		return 0
	}

	transfer := 0.0
	count := 0
	for task, accRandom := range randomInit {
		if accTrained, ok := afterTraining[task]; ok {
			transfer += accTrained - accRandom
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return transfer / float64(count)
}
