// Package layers provides power management and spiking neural network simulation
// for IronLattice ferroelectric compute-in-memory systems.
//
// This module implements:
// - CIM power management (power gating, standby modes)
// - Energy efficiency modeling (TOPS/W)
// - Leaky integrate-and-fire neurons
// - FeFET-based synapses with STDP
// - All-ferroelectric spiking neural networks
// - Spike timing dependent plasticity
//
// Based on research from:
// - ISSCC 2024: 22nm SRAM CIM macros (60-89 TOPS/W)
// - Advanced Science 2024: All-ferroelectric SNN (94.9% accuracy)
// - Nature Communications 2022: Anti-ferroelectric neuron (37 fJ/spike)
// - Wiley 2025: Spintronic memtransistor LIF neurons
package layers

import (
	"math"
	"math/rand"
	"sync"
)

// =============================================================================
// CIM Power Management
// =============================================================================

// PowerModeType represents different power states
type PowerModeType int

const (
	PowerModeActive PowerModeType = iota
	PowerModeStandby
	PowerModeSleep
	PowerModePowerDown
)

// PowerManagementConfig configures CIM power management
type PowerManagementConfig struct {
	ArrayRows          int
	ArrayCols          int
	TechnologyNode     int       // nm (e.g., 22, 28, 65)
	SupplyVoltage      float64   // VDD in volts
	StandbyVoltage     float64   // Reduced voltage for standby
	LeakagePerCell     float64   // Static leakage in pA
	DynamicEnergyMAC   float64   // Energy per MAC in fJ
	ClockFrequencyMHz  float64   // Operating frequency
	PowerGatingEnabled bool
	DVFSEnabled        bool
	MemoryType         string    // "SRAM", "ReRAM", "FeFET"
}

// DefaultPowerManagementConfig returns standard configuration for 22nm
func DefaultPowerManagementConfig() *PowerManagementConfig {
	return &PowerManagementConfig{
		ArrayRows:          64,
		ArrayCols:          64,
		TechnologyNode:     22,
		SupplyVoltage:      0.8,
		StandbyVoltage:     0.5,
		LeakagePerCell:     0.1,      // 0.1 pA for ReRAM (near-zero)
		DynamicEnergyMAC:   0.5,      // 0.5 fJ/MAC
		ClockFrequencyMHz:  500,
		PowerGatingEnabled: true,
		DVFSEnabled:        true,
		MemoryType:         "ReRAM",
	}
}

// CIMPowerManager manages power states for CIM arrays
type CIMPowerManager struct {
	Config           *PowerManagementConfig
	CurrentMode      PowerModeType
	BlockPowerStates [][]bool  // Per-block power gating state

	// Power metrics
	ActivePowerMW    float64
	StandbyPowerMW   float64
	LeakagePowerMW   float64
	TotalEnergyPJ    float64

	// Performance metrics
	EffectiveTOPSW   float64
	PeakTOPSW        float64
	Utilization      float64
}

// NewCIMPowerManager creates a new power manager
func NewCIMPowerManager(config *PowerManagementConfig) *CIMPowerManager {
	if config == nil {
		config = DefaultPowerManagementConfig()
	}

	pm := &CIMPowerManager{
		Config:           config,
		CurrentMode:      PowerModeActive,
		BlockPowerStates: make([][]bool, config.ArrayRows/8), // 8x8 blocks
	}

	for i := range pm.BlockPowerStates {
		pm.BlockPowerStates[i] = make([]bool, config.ArrayCols/8)
		for j := range pm.BlockPowerStates[i] {
			pm.BlockPowerStates[i][j] = true // All blocks powered on initially
		}
	}

	pm.calculateBasePower()
	pm.calculateEfficiency()

	return pm
}

func (pm *CIMPowerManager) calculateBasePower() {
	numCells := pm.Config.ArrayRows * pm.Config.ArrayCols

	// Leakage power based on memory type
	var leakagePerCellPA float64
	switch pm.Config.MemoryType {
	case "ReRAM":
		leakagePerCellPA = 0.1 // Near-zero standby for NVM
	case "FeFET":
		leakagePerCellPA = 0.05 // Very low for ferroelectric
	case "SRAM":
		leakagePerCellPA = 10.0 // Higher for volatile
	default:
		leakagePerCellPA = 1.0
	}

	// Total leakage power in mW
	pm.LeakagePowerMW = float64(numCells) * leakagePerCellPA * 1e-9 // pA to mW

	// Active power: dynamic + leakage
	// Dynamic power = C * V^2 * f * activity
	activity := 0.3 // Average switching activity
	capacitanceFF := 10.0 // Effective capacitance in fF
	pm.ActivePowerMW = float64(numCells) * capacitanceFF * 1e-15 *
		pm.Config.SupplyVoltage * pm.Config.SupplyVoltage *
		pm.Config.ClockFrequencyMHz * 1e6 * activity * 1000 // Convert to mW

	pm.ActivePowerMW += pm.LeakagePowerMW

	// Standby power (reduced voltage)
	voltageRatio := pm.Config.StandbyVoltage / pm.Config.SupplyVoltage
	pm.StandbyPowerMW = pm.LeakagePowerMW * voltageRatio * voltageRatio
}

func (pm *CIMPowerManager) calculateEfficiency() {
	// Calculate TOPS/W based on array size and power
	numMACs := pm.Config.ArrayRows * pm.Config.ArrayCols // MACs per cycle
	opsPerSecond := float64(numMACs) * pm.Config.ClockFrequencyMHz * 1e6 * 2 // 2 ops/MAC

	// TOPS
	tops := opsPerSecond / 1e12

	// TOPS/W
	if pm.ActivePowerMW > 0 {
		pm.PeakTOPSW = tops / (pm.ActivePowerMW / 1000) // mW to W
	}

	// Effective TOPS/W considering utilization
	pm.Utilization = 0.85 // 85% typical
	pm.EffectiveTOPSW = pm.PeakTOPSW * pm.Utilization
}

// SetPowerMode transitions to a new power mode
func (pm *CIMPowerManager) SetPowerMode(mode PowerModeType) {
	pm.CurrentMode = mode

	switch mode {
	case PowerModeActive:
		// Full power, all blocks enabled
		for i := range pm.BlockPowerStates {
			for j := range pm.BlockPowerStates[i] {
				pm.BlockPowerStates[i][j] = true
			}
		}
	case PowerModeStandby:
		// Reduced voltage, maintain data
		// For NVM, can power down most circuits
	case PowerModeSleep:
		// Minimal power, only retention
	case PowerModePowerDown:
		// Zero power (NVM retains data)
		for i := range pm.BlockPowerStates {
			for j := range pm.BlockPowerStates[i] {
				pm.BlockPowerStates[i][j] = false
			}
		}
	}
}

// PowerGateBlock enables/disables a specific block
func (pm *CIMPowerManager) PowerGateBlock(blockRow, blockCol int, enable bool) {
	if blockRow < len(pm.BlockPowerStates) && blockCol < len(pm.BlockPowerStates[0]) {
		pm.BlockPowerStates[blockRow][blockCol] = enable
	}
}

// GetCurrentPower returns current power consumption in mW
func (pm *CIMPowerManager) GetCurrentPower() float64 {
	switch pm.CurrentMode {
	case PowerModeActive:
		// Count active blocks
		activeBlocks := 0
		totalBlocks := len(pm.BlockPowerStates) * len(pm.BlockPowerStates[0])
		for i := range pm.BlockPowerStates {
			for j := range pm.BlockPowerStates[i] {
				if pm.BlockPowerStates[i][j] {
					activeBlocks++
				}
			}
		}
		ratio := float64(activeBlocks) / float64(totalBlocks)
		return pm.ActivePowerMW * ratio
	case PowerModeStandby:
		return pm.StandbyPowerMW
	case PowerModeSleep:
		return pm.LeakagePowerMW * 0.1 // 10% of leakage
	case PowerModePowerDown:
		return 0 // Zero for NVM
	default:
		return pm.ActivePowerMW
	}
}

// RecordOperation records energy for a CIM operation
func (pm *CIMPowerManager) RecordOperation(numMACs int) {
	energyFJ := float64(numMACs) * pm.Config.DynamicEnergyMAC
	pm.TotalEnergyPJ += energyFJ / 1000 // fJ to pJ
}

// =============================================================================
// Energy Efficiency Models
// =============================================================================

// CIMEnergyModel represents energy efficiency for different CIM implementations
type CIMEnergyModel struct {
	Name            string
	TechnologyNode  int
	ArraySize       int
	Precision       int      // bits
	PeakTOPSW       float64
	AreaEfficiency  float64  // TOPS/mm²
	MemoryType      string
	Architecture    string   // "analog", "digital", "hybrid"
}

// StandardCIMModels returns published CIM macro specifications
func StandardCIMModels() []*CIMEnergyModel {
	return []*CIMEnergyModel{
		{
			Name:           "ISSCC21_SRAM_Digital",
			TechnologyNode: 22,
			ArraySize:      64,
			Precision:      8,
			PeakTOPSW:      89,
			AreaEfficiency: 16.3,
			MemoryType:     "SRAM",
			Architecture:   "digital",
		},
		{
			Name:           "ISSCC21_ReRAM_4Mb",
			TechnologyNode: 22,
			ArraySize:      256,
			Precision:      8,
			PeakTOPSW:      195.7,
			AreaEfficiency: 25.0,
			MemoryType:     "ReRAM",
			Architecture:   "analog",
		},
		{
			Name:           "ISSCC24_SRAM_Hybrid",
			TechnologyNode: 22,
			ArraySize:      64,
			Precision:      8,
			PeakTOPSW:      60.8,
			AreaEfficiency: 10.0,
			MemoryType:     "SRAM",
			Architecture:   "hybrid",
		},
		{
			Name:           "VLSI22_SRAM_C2C",
			TechnologyNode: 22,
			ArraySize:      64,
			Precision:      8,
			PeakTOPSW:      32.2,
			AreaEfficiency: 4.0,
			MemoryType:     "SRAM",
			Architecture:   "analog",
		},
		{
			Name:           "ASSCC24_SRAM_DriveStrength",
			TechnologyNode: 65,
			ArraySize:      128,
			Precision:      8,
			PeakTOPSW:      687.5,
			AreaEfficiency: 8.0,
			MemoryType:     "SRAM",
			Architecture:   "analog",
		},
		{
			Name:           "FeFET_MLC_28nm",
			TechnologyNode: 28,
			ArraySize:      64,
			Precision:      4,
			PeakTOPSW:      885,
			AreaEfficiency: 30.0,
			MemoryType:     "FeFET",
			Architecture:   "analog",
		},
	}
}

// =============================================================================
// Leaky Integrate-and-Fire Neuron
// =============================================================================

// LIFNeuronConfig configures LIF neuron parameters
type LIFNeuronConfig struct {
	MembraneCapacitance  float64 // Membrane capacitance in pF
	LeakConductance      float64 // Leak conductance in nS
	ThresholdVoltage     float64 // Spike threshold in mV
	ResetVoltage         float64 // Reset voltage in mV
	RestingPotential     float64 // Resting potential in mV
	RefractoryPeriodMs   float64 // Refractory period in ms
	TimeConstantMs       float64 // Membrane time constant
	EnergyPerSpikePJ     float64 // Energy consumption per spike
}

// DefaultLIFNeuronConfig returns biologically-inspired parameters
func DefaultLIFNeuronConfig() *LIFNeuronConfig {
	return &LIFNeuronConfig{
		MembraneCapacitance: 10.0,    // 10 pF
		LeakConductance:     1.0,     // 1 nS
		ThresholdVoltage:    -50.0,   // -50 mV
		ResetVoltage:        -70.0,   // -70 mV
		RestingPotential:    -65.0,   // -65 mV
		RefractoryPeriodMs:  2.0,     // 2 ms
		TimeConstantMs:      10.0,    // 10 ms (τ = C/g)
		EnergyPerSpikePJ:    37.0,    // 37 fJ/spike (anti-ferroelectric neuron)
	}
}

// LIFNeuron implements leaky integrate-and-fire dynamics
type LIFNeuron struct {
	Config            *LIFNeuronConfig
	MembranePotential float64
	LastSpikeTime     float64
	IsRefractory      bool
	SpikeCount        int
	TotalEnergy       float64  // pJ

	// Input integration
	InputCurrent      float64
}

// NewLIFNeuron creates a new LIF neuron
func NewLIFNeuron(config *LIFNeuronConfig) *LIFNeuron {
	if config == nil {
		config = DefaultLIFNeuronConfig()
	}
	return &LIFNeuron{
		Config:            config,
		MembranePotential: config.RestingPotential,
	}
}

// Step advances the neuron by one time step
func (n *LIFNeuron) Step(dt float64, inputCurrent float64) bool {
	// Check refractory period
	if n.IsRefractory {
		n.LastSpikeTime += dt
		if n.LastSpikeTime >= n.Config.RefractoryPeriodMs {
			n.IsRefractory = false
		}
		return false
	}

	// Leaky integration
	// dV/dt = (-(V - V_rest) + R*I) / τ
	// where R = 1/g_leak, τ = C/g_leak
	tau := n.Config.TimeConstantMs
	resistance := 1.0 / n.Config.LeakConductance // MΩ

	dV := (-(n.MembranePotential - n.Config.RestingPotential) +
		resistance*inputCurrent) * dt / tau
	n.MembranePotential += dV

	// Check for spike
	if n.MembranePotential >= n.Config.ThresholdVoltage {
		// Fire spike
		n.MembranePotential = n.Config.ResetVoltage
		n.IsRefractory = true
		n.LastSpikeTime = 0
		n.SpikeCount++
		n.TotalEnergy += n.Config.EnergyPerSpikePJ
		return true
	}

	return false
}

// Reset resets the neuron to initial state
func (n *LIFNeuron) Reset() {
	n.MembranePotential = n.Config.RestingPotential
	n.LastSpikeTime = 0
	n.IsRefractory = false
}

// =============================================================================
// Ferroelectric LIF Neuron (Anti-ferroelectric / MPB-based)
// =============================================================================

// FerroelectricNeuronConfig configures ferroelectric neuron
type FerroelectricNeuronConfig struct {
	DeviceType         string   // "antiferroelectric", "MPB", "FeTFT"
	Polarization       float64  // Pr in µC/cm²
	CoerciveField      float64  // Ec in MV/cm
	ThresholdVoltage   float64  // Threshold in V
	IntegrationTime    float64  // Integration time constant
	EnergyPerSpikeFJ   float64  // Energy per spike in fJ
	Endurance          float64  // Cycles
}

// DefaultFerroelectricNeuronConfig returns standard HZO parameters
func DefaultFerroelectricNeuronConfig() *FerroelectricNeuronConfig {
	return &FerroelectricNeuronConfig{
		DeviceType:        "antiferroelectric",
		Polarization:      20.0,    // 20 µC/cm²
		CoerciveField:     1.0,     // 1 MV/cm
		ThresholdVoltage:  1.5,     // 1.5 V
		IntegrationTime:   1.0,     // 1 ms
		EnergyPerSpikeFJ:  37.0,    // 37 fJ (from Nature Comm 2022)
		Endurance:         1e12,    // 10^12 cycles
	}
}

// FerroelectricLIFNeuron implements LIF using ferroelectric dynamics
type FerroelectricLIFNeuron struct {
	Config              *FerroelectricNeuronConfig
	AccumulatedCharge   float64  // Integrated charge
	PolarizationState   float64  // Current polarization
	SpikeOutput         bool
	SpikeCount          int
	TotalEnergy         float64  // fJ

	// Double-gate configuration for MPB neurons
	TopGateVoltage      float64
	BottomGateVoltage   float64
}

// NewFerroelectricLIFNeuron creates a new ferroelectric neuron
func NewFerroelectricLIFNeuron(config *FerroelectricNeuronConfig) *FerroelectricLIFNeuron {
	if config == nil {
		config = DefaultFerroelectricNeuronConfig()
	}
	return &FerroelectricLIFNeuron{
		Config: config,
	}
}

// Integrate accumulates input and checks for spike
func (fn *FerroelectricLIFNeuron) Integrate(inputVoltage float64, dt float64) bool {
	// Ferroelectric integration: charge accumulates on polarization
	// For anti-ferroelectric: spontaneous depolarization provides leak
	chargeInput := inputVoltage * dt / fn.Config.IntegrationTime
	fn.AccumulatedCharge += chargeInput

	// Spontaneous depolarization (leak)
	leakRate := 0.1 // 10% per time step for AFE
	fn.AccumulatedCharge *= (1.0 - leakRate*dt)

	// Check threshold
	if fn.AccumulatedCharge >= fn.Config.ThresholdVoltage {
		// Spike and reset
		fn.SpikeOutput = true
		fn.AccumulatedCharge = 0
		fn.SpikeCount++
		fn.TotalEnergy += fn.Config.EnergyPerSpikeFJ
		return true
	}

	fn.SpikeOutput = false
	return false
}

// =============================================================================
// FeFET Synapse with STDP
// =============================================================================

// FeFETSynapseConfig configures FeFET synapse parameters
type FeFETSynapseConfig struct {
	NumConductanceLevels int     // MLC levels
	MinConductance       float64 // Minimum G in µS
	MaxConductance       float64 // Maximum G in µS
	ProgramVoltage       float64 // Program voltage
	EraseVoltage         float64 // Erase voltage
	RetentionTime        float64 // Retention in seconds
	Endurance            float64 // Write cycles
	STDPEnabled          bool
	LTPTimeConstant      float64 // LTP time window in ms
	LTDTimeConstant      float64 // LTD time window in ms
	LearningRate         float64
}

// DefaultFeFETSynapseConfig returns standard HZO FeFET parameters
func DefaultFeFETSynapseConfig() *FeFETSynapseConfig {
	return &FeFETSynapseConfig{
		NumConductanceLevels: 16,     // 4-bit MLC
		MinConductance:       1.0,    // 1 µS
		MaxConductance:       100.0,  // 100 µS
		ProgramVoltage:       3.0,    // 3V
		EraseVoltage:         -3.0,   // -3V
		RetentionTime:        3.15e7, // 10 years
		Endurance:            1e10,   // 10^10 cycles
		STDPEnabled:          true,
		LTPTimeConstant:      20.0,   // 20 ms
		LTDTimeConstant:      20.0,   // 20 ms
		LearningRate:         0.01,
	}
}

// FeFETSynapse implements a ferroelectric FET synapse
type FeFETSynapse struct {
	Config           *FeFETSynapseConfig
	Conductance      float64
	Weight           float64  // Normalized weight [0, 1]
	LastPreSpikeTime float64
	LastPostSpikeTime float64
	WriteCount       int
}

// NewFeFETSynapse creates a new FeFET synapse
func NewFeFETSynapse(config *FeFETSynapseConfig) *FeFETSynapse {
	if config == nil {
		config = DefaultFeFETSynapseConfig()
	}

	syn := &FeFETSynapse{
		Config: config,
		Weight: 0.5, // Initialize to middle weight
	}
	syn.updateConductance()

	return syn
}

func (s *FeFETSynapse) updateConductance() {
	// Map weight to conductance
	s.Conductance = s.Config.MinConductance +
		s.Weight*(s.Config.MaxConductance-s.Config.MinConductance)
}

// ApplySTDP applies spike-timing-dependent plasticity
func (s *FeFETSynapse) ApplySTDP(preSpikeTime, postSpikeTime float64) {
	if !s.Config.STDPEnabled {
		return
	}

	dt := postSpikeTime - preSpikeTime // Timing difference

	var deltaW float64
	if dt > 0 {
		// Pre before post -> LTP (potentiation)
		deltaW = s.Config.LearningRate * math.Exp(-dt/s.Config.LTPTimeConstant)
	} else {
		// Post before pre -> LTD (depression)
		deltaW = -s.Config.LearningRate * math.Exp(dt/s.Config.LTDTimeConstant)
	}

	// Update weight with bounds
	s.Weight += deltaW
	if s.Weight < 0 {
		s.Weight = 0
	}
	if s.Weight > 1 {
		s.Weight = 1
	}

	s.updateConductance()
	s.WriteCount++
}

// TransmitSpike transmits a pre-synaptic spike
func (s *FeFETSynapse) TransmitSpike(spikeTime float64) float64 {
	s.LastPreSpikeTime = spikeTime
	// Output current proportional to conductance
	return s.Conductance * 1e-3 // Convert µS to mS for current
}

// ReceivePostSpike records post-synaptic spike for STDP
func (s *FeFETSynapse) ReceivePostSpike(spikeTime float64) {
	s.LastPostSpikeTime = spikeTime
	// Apply STDP if we have a recent pre-spike
	if s.LastPreSpikeTime > 0 {
		s.ApplySTDP(s.LastPreSpikeTime, spikeTime)
	}
}

// =============================================================================
// All-Ferroelectric Spiking Neural Network
// =============================================================================

// AllFerroelectricSNNConfig configures the SNN
type AllFerroelectricSNNConfig struct {
	InputSize         int
	HiddenSize        int
	OutputSize        int
	TimeSteps         int
	DtMs              float64  // Time step in ms
	NeuronType        string   // "antiferroelectric", "MPB", "I2FET"
	SynapseType       string   // "FeFET", "FeTFT"
	LearningEnabled   bool
}

// DefaultAllFerroelectricSNNConfig returns standard configuration
func DefaultAllFerroelectricSNNConfig() *AllFerroelectricSNNConfig {
	return &AllFerroelectricSNNConfig{
		InputSize:       784,    // MNIST
		HiddenSize:      128,
		OutputSize:      10,
		TimeSteps:       100,
		DtMs:            1.0,
		NeuronType:      "MPB",
		SynapseType:     "FeFET",
		LearningEnabled: true,
	}
}

// AllFerroelectricSNN implements a complete SNN using ferroelectric devices
type AllFerroelectricSNN struct {
	Config           *AllFerroelectricSNNConfig
	InputNeurons     []*FerroelectricLIFNeuron
	HiddenNeurons    []*FerroelectricLIFNeuron
	OutputNeurons    []*FerroelectricLIFNeuron
	InputHiddenSyn   [][]*FeFETSynapse  // [input][hidden]
	HiddenOutputSyn  [][]*FeFETSynapse  // [hidden][output]

	// Spike recordings
	InputSpikes      [][]bool  // [time][neuron]
	HiddenSpikes     [][]bool
	OutputSpikes     [][]bool

	// Performance metrics
	TotalEnergy      float64  // fJ
	Accuracy         float64
	InferenceTime    float64  // ms

	mu sync.Mutex
}

// NewAllFerroelectricSNN creates a new all-ferroelectric SNN
func NewAllFerroelectricSNN(config *AllFerroelectricSNNConfig) *AllFerroelectricSNN {
	if config == nil {
		config = DefaultAllFerroelectricSNNConfig()
	}

	snn := &AllFerroelectricSNN{
		Config:          config,
		InputNeurons:    make([]*FerroelectricLIFNeuron, config.InputSize),
		HiddenNeurons:   make([]*FerroelectricLIFNeuron, config.HiddenSize),
		OutputNeurons:   make([]*FerroelectricLIFNeuron, config.OutputSize),
		InputHiddenSyn:  make([][]*FeFETSynapse, config.InputSize),
		HiddenOutputSyn: make([][]*FeFETSynapse, config.HiddenSize),
	}

	// Initialize neurons
	neuronConfig := DefaultFerroelectricNeuronConfig()
	neuronConfig.DeviceType = config.NeuronType

	for i := 0; i < config.InputSize; i++ {
		snn.InputNeurons[i] = NewFerroelectricLIFNeuron(neuronConfig)
	}
	for i := 0; i < config.HiddenSize; i++ {
		snn.HiddenNeurons[i] = NewFerroelectricLIFNeuron(neuronConfig)
	}
	for i := 0; i < config.OutputSize; i++ {
		snn.OutputNeurons[i] = NewFerroelectricLIFNeuron(neuronConfig)
	}

	// Initialize synapses
	synapseConfig := DefaultFeFETSynapseConfig()
	synapseConfig.STDPEnabled = config.LearningEnabled

	for i := 0; i < config.InputSize; i++ {
		snn.InputHiddenSyn[i] = make([]*FeFETSynapse, config.HiddenSize)
		for j := 0; j < config.HiddenSize; j++ {
			snn.InputHiddenSyn[i][j] = NewFeFETSynapse(synapseConfig)
			// Random initialization
			snn.InputHiddenSyn[i][j].Weight = rand.Float64()
			snn.InputHiddenSyn[i][j].updateConductance()
		}
	}

	for i := 0; i < config.HiddenSize; i++ {
		snn.HiddenOutputSyn[i] = make([]*FeFETSynapse, config.OutputSize)
		for j := 0; j < config.OutputSize; j++ {
			snn.HiddenOutputSyn[i][j] = NewFeFETSynapse(synapseConfig)
			snn.HiddenOutputSyn[i][j].Weight = rand.Float64()
			snn.HiddenOutputSyn[i][j].updateConductance()
		}
	}

	return snn
}

// Forward performs forward pass with spike encoding
func (snn *AllFerroelectricSNN) Forward(input []float64) []int {
	snn.mu.Lock()
	defer snn.mu.Unlock()

	dt := snn.Config.DtMs
	timeSteps := snn.Config.TimeSteps

	// Initialize spike recordings
	snn.InputSpikes = make([][]bool, timeSteps)
	snn.HiddenSpikes = make([][]bool, timeSteps)
	snn.OutputSpikes = make([][]bool, timeSteps)

	outputSpikeCount := make([]int, snn.Config.OutputSize)

	for t := 0; t < timeSteps; t++ {
		currentTime := float64(t) * dt

		snn.InputSpikes[t] = make([]bool, snn.Config.InputSize)
		snn.HiddenSpikes[t] = make([]bool, snn.Config.HiddenSize)
		snn.OutputSpikes[t] = make([]bool, snn.Config.OutputSize)

		// Input layer: rate coding
		for i := 0; i < snn.Config.InputSize; i++ {
			// Poisson spike generation based on input intensity
			rate := input[i] * 0.1 // Max 100 Hz
			if rand.Float64() < rate*dt/1000.0 {
				snn.InputSpikes[t][i] = true
			}
		}

		// Hidden layer
		for j := 0; j < snn.Config.HiddenSize; j++ {
			var inputCurrent float64
			for i := 0; i < snn.Config.InputSize; i++ {
				if snn.InputSpikes[t][i] {
					inputCurrent += snn.InputHiddenSyn[i][j].TransmitSpike(currentTime)
				}
			}

			if snn.HiddenNeurons[j].Integrate(inputCurrent, dt) {
				snn.HiddenSpikes[t][j] = true
				// STDP update
				if snn.Config.LearningEnabled {
					for i := 0; i < snn.Config.InputSize; i++ {
						if snn.InputSpikes[t][i] {
							snn.InputHiddenSyn[i][j].ReceivePostSpike(currentTime)
						}
					}
				}
			}
		}

		// Output layer
		for k := 0; k < snn.Config.OutputSize; k++ {
			var inputCurrent float64
			for j := 0; j < snn.Config.HiddenSize; j++ {
				if snn.HiddenSpikes[t][j] {
					inputCurrent += snn.HiddenOutputSyn[j][k].TransmitSpike(currentTime)
				}
			}

			if snn.OutputNeurons[k].Integrate(inputCurrent, dt) {
				snn.OutputSpikes[t][k] = true
				outputSpikeCount[k]++
			}
		}
	}

	return outputSpikeCount
}

// Classify returns the predicted class based on spike counts
func (snn *AllFerroelectricSNN) Classify(input []float64) int {
	spikeCounts := snn.Forward(input)

	// Winner-take-all: highest spike count
	maxCount := 0
	maxIdx := 0
	for i, count := range spikeCounts {
		if count > maxCount {
			maxCount = count
			maxIdx = i
		}
	}

	return maxIdx
}

// GetEnergyConsumption returns total energy in fJ
func (snn *AllFerroelectricSNN) GetEnergyConsumption() float64 {
	var total float64

	// Neuron energy
	for _, n := range snn.InputNeurons {
		total += n.TotalEnergy
	}
	for _, n := range snn.HiddenNeurons {
		total += n.TotalEnergy
	}
	for _, n := range snn.OutputNeurons {
		total += n.TotalEnergy
	}

	// Synapse energy (reads during inference)
	numSynapses := snn.Config.InputSize*snn.Config.HiddenSize +
		snn.Config.HiddenSize*snn.Config.OutputSize
	readEnergyPerSynapse := 1.0 // fJ
	total += float64(numSynapses) * readEnergyPerSynapse

	return total
}

// =============================================================================
// STDP Learning Rule
// =============================================================================

// STDPConfig configures STDP learning
type STDPConfig struct {
	TauPlus       float64 // LTP time constant (ms)
	TauMinus      float64 // LTD time constant (ms)
	APlus         float64 // LTP amplitude
	AMinus        float64 // LTD amplitude
	WMax          float64 // Maximum weight
	WMin          float64 // Minimum weight
	UseSymmetric  bool    // Symmetric vs asymmetric STDP
}

// DefaultSTDPConfig returns standard STDP parameters
func DefaultSTDPConfig() *STDPConfig {
	return &STDPConfig{
		TauPlus:      20.0,
		TauMinus:     20.0,
		APlus:        0.005,
		AMinus:       0.00525, // Slightly stronger LTD
		WMax:         1.0,
		WMin:         0.0,
		UseSymmetric: false,
	}
}

// STDPLearner implements STDP learning rule
type STDPLearner struct {
	Config        *STDPConfig
	PreTraces     []float64  // Eligibility traces for pre-synaptic spikes
	PostTraces    []float64  // Eligibility traces for post-synaptic spikes
	NumPre        int
	NumPost       int
}

// NewSTDPLearner creates a new STDP learner
func NewSTDPLearner(config *STDPConfig, numPre, numPost int) *STDPLearner {
	if config == nil {
		config = DefaultSTDPConfig()
	}
	return &STDPLearner{
		Config:     config,
		PreTraces:  make([]float64, numPre),
		PostTraces: make([]float64, numPost),
		NumPre:     numPre,
		NumPost:    numPost,
	}
}

// UpdateTraces updates eligibility traces
func (sl *STDPLearner) UpdateTraces(dt float64) {
	// Exponential decay
	decayPre := math.Exp(-dt / sl.Config.TauPlus)
	decayPost := math.Exp(-dt / sl.Config.TauMinus)

	for i := range sl.PreTraces {
		sl.PreTraces[i] *= decayPre
	}
	for i := range sl.PostTraces {
		sl.PostTraces[i] *= decayPost
	}
}

// OnPreSpike handles pre-synaptic spike
func (sl *STDPLearner) OnPreSpike(preIdx int) {
	sl.PreTraces[preIdx] = 1.0
}

// OnPostSpike handles post-synaptic spike
func (sl *STDPLearner) OnPostSpike(postIdx int) {
	sl.PostTraces[postIdx] = 1.0
}

// ComputeWeightChange computes weight change for a synapse
func (sl *STDPLearner) ComputeWeightChange(preIdx, postIdx int, currentWeight float64) float64 {
	// LTP: post after pre -> use post trace at time of pre spike
	// LTD: pre after post -> use pre trace at time of post spike
	ltp := sl.Config.APlus * sl.PreTraces[preIdx] * (sl.Config.WMax - currentWeight)
	ltd := sl.Config.AMinus * sl.PostTraces[postIdx] * (currentWeight - sl.Config.WMin)

	return ltp - ltd
}

// =============================================================================
// 2D Material SNN Components
// =============================================================================

// I2FETNeuronConfig configures impact ionization FET neuron
type I2FETNeuronConfig struct {
	Material           string   // "WSe2", "MoS2"
	BreakdownVoltage   float64  // V
	ThresholdCurrent   float64  // nA
	EnergyPerSpikePJ   float64  // pJ
	SwitchingTime      float64  // ns
}

// I2FETNeuron implements impact ionization FET spiking neuron
type I2FETNeuron struct {
	Config              *I2FETNeuronConfig
	AccumulatedCurrent  float64
	SpikeOutput         bool
	SpikeCount          int
	TotalEnergy         float64
}

// NewI2FETNeuron creates a new I2FET neuron
func NewI2FETNeuron(config *I2FETNeuronConfig) *I2FETNeuron {
	if config == nil {
		config = &I2FETNeuronConfig{
			Material:          "WSe2",
			BreakdownVoltage:  5.0,
			ThresholdCurrent:  100.0,     // 100 nA
			EnergyPerSpikePJ:  2.0,       // 2 pJ/spike
			SwitchingTime:     10.0,      // 10 ns
		}
	}
	return &I2FETNeuron{Config: config}
}

// Step advances the I2FET neuron
func (n *I2FETNeuron) Step(inputCurrent float64, dt float64) bool {
	// Integration with leakage
	leakRate := 0.05 // 5% per time step
	n.AccumulatedCurrent = n.AccumulatedCurrent*(1-leakRate) + inputCurrent

	// Check for impact ionization threshold
	if n.AccumulatedCurrent >= n.Config.ThresholdCurrent {
		n.SpikeOutput = true
		n.AccumulatedCurrent = 0 // Reset
		n.SpikeCount++
		n.TotalEnergy += n.Config.EnergyPerSpikePJ
		return true
	}

	n.SpikeOutput = false
	return false
}

// =============================================================================
// Performance Metrics
// =============================================================================

// SNNPerformanceMetrics stores SNN performance data
type SNNPerformanceMetrics struct {
	Accuracy           float64
	EnergyPerInference float64  // pJ
	InferenceLatency   float64  // ms
	SpikeActivity      float64  // Average spikes per neuron
	EnergyPerSpike     float64  // fJ
	ThroughputSPS      float64  // Samples per second
}

// ComputeMetrics calculates performance metrics for an SNN
func ComputeMetrics(snn *AllFerroelectricSNN, testSamples int, accuracy float64) *SNNPerformanceMetrics {
	totalEnergy := snn.GetEnergyConsumption() * float64(testSamples)
	inferenceTime := float64(snn.Config.TimeSteps) * snn.Config.DtMs

	// Count total spikes
	totalSpikes := 0
	for _, n := range snn.HiddenNeurons {
		totalSpikes += n.SpikeCount
	}
	for _, n := range snn.OutputNeurons {
		totalSpikes += n.SpikeCount
	}

	numNeurons := snn.Config.HiddenSize + snn.Config.OutputSize
	spikeActivity := float64(totalSpikes) / float64(numNeurons) / float64(testSamples)

	return &SNNPerformanceMetrics{
		Accuracy:           accuracy,
		EnergyPerInference: totalEnergy / float64(testSamples),
		InferenceLatency:   inferenceTime,
		SpikeActivity:      spikeActivity,
		EnergyPerSpike:     37.0, // fJ (anti-ferroelectric)
		ThroughputSPS:      1000.0 / inferenceTime * float64(testSamples),
	}
}
