// Package layers provides reinforcement learning and 3D chiplet CIM simulation.
//
// This module implements:
// - Memristor-based actor-critic reinforcement learning
// - Temporal difference (TD) learning in hardware
// - Reward-modulated STDP (R-STDP)
// - 3D monolithic CIM integration (M3D-LIME style)
// - Heterogeneous CIM chiplet architecture (3D-CIMlet)
// - Thermal-aware chiplet placement
//
// Based on research from Nature Machine Intelligence 2025, Nature Communications 2023,
// DAC 2024/2025, and ISSCC 2025.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// REINFORCEMENT LEARNING WITH MEMRISTOR CROSSBARS
// =============================================================================

// RLConfig configures the reinforcement learning system.
type RLConfig struct {
	StateSize       int
	ActionSize      int
	HiddenSize      int
	LearningRate    float64
	DiscountFactor  float64 // Gamma
	TDLambda        float64 // Eligibility trace decay
	ExplorationRate float64 // Epsilon for exploration
	UseMemristor    bool
	CrossbarSize    int
}

// MemristorSynapse models a memristor as a synapse.
type MemristorSynapse struct {
	Conductance      float64 // Current conductance (weight)
	MinConductance   float64 // G_min
	MaxConductance   float64 // G_max
	EligibilityTrace float64 // For TD learning
	PreSpikeTime     float64 // Last presynaptic spike
	PostSpikeTime    float64 // Last postsynaptic spike
	LTPThreshold     float64 // Long-term potentiation threshold
	LTDThreshold     float64 // Long-term depression threshold
}

// ActorCriticCIM implements actor-critic RL with memristor arrays.
type ActorCriticCIM struct {
	Config         *RLConfig
	ActorWeights   [][]*MemristorSynapse  // Policy network
	CriticWeights  [][]*MemristorSynapse  // Value network
	ActorBias      []float64
	CriticBias     []float64
	TDError        float64                // Temporal difference error
	ValueEstimate  float64                // V(s)
	Stats          *RLStats
}

// RLStats tracks RL performance metrics.
type RLStats struct {
	TotalReward       float64
	Episodes          int
	AverageReturn     float64
	TDErrorHistory    []float64
	WeightUpdateCount int64
	EnergyPerUpdate   float64 // pJ
}

// NewActorCriticCIM creates a new memristor-based actor-critic system.
func NewActorCriticCIM(config *RLConfig) *ActorCriticCIM {
	ac := &ActorCriticCIM{
		Config:      config,
		ActorBias:   make([]float64, config.ActionSize),
		CriticBias:  make([]float64, 1),
		Stats:       &RLStats{},
	}

	// Initialize actor weights (state -> action)
	ac.ActorWeights = make([][]*MemristorSynapse, config.StateSize)
	for i := 0; i < config.StateSize; i++ {
		ac.ActorWeights[i] = make([]*MemristorSynapse, config.ActionSize)
		for j := 0; j < config.ActionSize; j++ {
			ac.ActorWeights[i][j] = NewMemristorSynapse()
		}
	}

	// Initialize critic weights (state -> value)
	ac.CriticWeights = make([][]*MemristorSynapse, config.StateSize)
	for i := 0; i < config.StateSize; i++ {
		ac.CriticWeights[i] = make([]*MemristorSynapse, 1)
		ac.CriticWeights[i][0] = NewMemristorSynapse()
	}

	return ac
}

// NewMemristorSynapse creates a new memristor synapse.
func NewMemristorSynapse() *MemristorSynapse {
	return &MemristorSynapse{
		Conductance:    0.5 + rand.Float64()*0.5, // Random init 0.5-1.0
		MinConductance: 0.1,
		MaxConductance: 1.0,
		LTPThreshold:   20.0,  // ms
		LTDThreshold:   -20.0, // ms
	}
}

// Forward computes action probabilities given state.
func (ac *ActorCriticCIM) Forward(state []float64) ([]float64, float64) {
	if len(state) != ac.Config.StateSize {
		return nil, 0
	}

	// Actor: compute action logits
	actionLogits := make([]float64, ac.Config.ActionSize)
	for j := 0; j < ac.Config.ActionSize; j++ {
		sum := ac.ActorBias[j]
		for i := 0; i < ac.Config.StateSize; i++ {
			sum += state[i] * ac.ActorWeights[i][j].Conductance
		}
		actionLogits[j] = sum
	}

	// Softmax for action probabilities
	actionProbs := softmax(actionLogits)

	// Critic: compute value estimate
	value := ac.CriticBias[0]
	for i := 0; i < ac.Config.StateSize; i++ {
		value += state[i] * ac.CriticWeights[i][0].Conductance
	}

	ac.ValueEstimate = value
	return actionProbs, value
}

// SelectAction chooses action using epsilon-greedy or softmax.
func (ac *ActorCriticCIM) SelectAction(actionProbs []float64) int {
	// Epsilon-greedy exploration
	if rand.Float64() < ac.Config.ExplorationRate {
		return rand.Intn(len(actionProbs))
	}

	// Sample from probability distribution
	r := rand.Float64()
	cumProb := 0.0
	for i, prob := range actionProbs {
		cumProb += prob
		if r < cumProb {
			return i
		}
	}
	return len(actionProbs) - 1
}

// ComputeTDError calculates temporal difference error.
func (ac *ActorCriticCIM) ComputeTDError(reward, nextValue float64, done bool) float64 {
	target := reward
	if !done {
		target += ac.Config.DiscountFactor * nextValue
	}
	ac.TDError = target - ac.ValueEstimate
	ac.Stats.TDErrorHistory = append(ac.Stats.TDErrorHistory, ac.TDError)
	return ac.TDError
}

// UpdateWeights performs TD-based weight updates on memristors.
func (ac *ActorCriticCIM) UpdateWeights(state []float64, action int, tdError float64) {
	lr := ac.Config.LearningRate

	// Update critic weights using TD error
	for i := 0; i < ac.Config.StateSize; i++ {
		synapse := ac.CriticWeights[i][0]

		// Update eligibility trace
		synapse.EligibilityTrace *= ac.Config.TDLambda * ac.Config.DiscountFactor
		synapse.EligibilityTrace += state[i]

		// Memristor conductance update
		delta := lr * tdError * synapse.EligibilityTrace
		ac.updateMemristorConductance(synapse, delta)
	}

	// Update actor weights using policy gradient
	for i := 0; i < ac.Config.StateSize; i++ {
		for j := 0; j < ac.Config.ActionSize; j++ {
			synapse := ac.ActorWeights[i][j]

			// Advantage-weighted update
			advantage := tdError
			actionIndicator := 0.0
			if j == action {
				actionIndicator = 1.0
			}

			delta := lr * advantage * state[i] * actionIndicator
			ac.updateMemristorConductance(synapse, delta)
		}
	}

	ac.Stats.WeightUpdateCount++
	ac.Stats.EnergyPerUpdate = 0.5 // pJ per memristor update
}

// updateMemristorConductance applies conductance change with bounds.
func (ac *ActorCriticCIM) updateMemristorConductance(syn *MemristorSynapse, delta float64) {
	syn.Conductance += delta

	// Clamp to valid range
	if syn.Conductance < syn.MinConductance {
		syn.Conductance = syn.MinConductance
	}
	if syn.Conductance > syn.MaxConductance {
		syn.Conductance = syn.MaxConductance
	}
}

// softmax computes softmax probabilities.
func softmax(logits []float64) []float64 {
	maxLogit := logits[0]
	for _, l := range logits {
		if l > maxLogit {
			maxLogit = l
		}
	}

	expSum := 0.0
	probs := make([]float64, len(logits))
	for i, l := range logits {
		probs[i] = math.Exp(l - maxLogit)
		expSum += probs[i]
	}

	for i := range probs {
		probs[i] /= expSum
	}
	return probs
}

// =============================================================================
// REWARD-MODULATED STDP (R-STDP)
// =============================================================================

// RSTDPConfig configures reward-modulated STDP.
type RSTDPConfig struct {
	TauPlus        float64 // LTP time constant (ms)
	TauMinus       float64 // LTD time constant (ms)
	APlus          float64 // LTP amplitude
	AMinus         float64 // LTD amplitude
	RewardWindow   float64 // Time window for reward (ms)
	EligibilityTau float64 // Eligibility trace decay (ms)
}

// RSTDPSynapse implements reward-modulated STDP.
type RSTDPSynapse struct {
	Weight           float64
	EligibilityTrace float64
	Config           *RSTDPConfig
	LastPreSpike     float64
	LastPostSpike    float64
}

// NewRSTDPSynapse creates a new R-STDP synapse.
func NewRSTDPSynapse(config *RSTDPConfig) *RSTDPSynapse {
	return &RSTDPSynapse{
		Weight: 0.5 + rand.Float64()*0.5,
		Config: config,
	}
}

// OnPreSpike handles presynaptic spike arrival.
func (s *RSTDPSynapse) OnPreSpike(time float64) {
	s.LastPreSpike = time

	// Check for LTD (post before pre)
	if s.LastPostSpike > 0 {
		dt := time - s.LastPostSpike
		if dt > 0 && dt < s.Config.RewardWindow {
			// Update eligibility trace for LTD
			stdpChange := -s.Config.AMinus * math.Exp(-dt/s.Config.TauMinus)
			s.EligibilityTrace += stdpChange
		}
	}
}

// OnPostSpike handles postsynaptic spike.
func (s *RSTDPSynapse) OnPostSpike(time float64) {
	s.LastPostSpike = time

	// Check for LTP (pre before post)
	if s.LastPreSpike > 0 {
		dt := time - s.LastPreSpike
		if dt > 0 && dt < s.Config.RewardWindow {
			// Update eligibility trace for LTP
			stdpChange := s.Config.APlus * math.Exp(-dt/s.Config.TauPlus)
			s.EligibilityTrace += stdpChange
		}
	}
}

// ApplyReward modulates weight change by reward signal.
func (s *RSTDPSynapse) ApplyReward(reward, learningRate float64) {
	// Three-factor learning: weight change = eligibility * reward
	weightChange := learningRate * s.EligibilityTrace * reward
	s.Weight += weightChange

	// Clamp weight
	s.Weight = math.Max(0, math.Min(1, s.Weight))

	// Decay eligibility trace
	s.EligibilityTrace *= math.Exp(-1.0 / s.Config.EligibilityTau)
}

// DecayEligibility decays the eligibility trace over time.
func (s *RSTDPSynapse) DecayEligibility(dt float64) {
	s.EligibilityTrace *= math.Exp(-dt / s.Config.EligibilityTau)
}

// =============================================================================
// Q-LEARNING HARDWARE ACCELERATOR
// =============================================================================

// QLearningConfig configures the Q-learning accelerator.
type QLearningConfig struct {
	StateCount     int
	ActionCount    int
	LearningRate   float64
	DiscountFactor float64
	InitialEpsilon float64
	EpsilonDecay   float64
	MinEpsilon     float64
	UseCrossbar    bool
}

// QLearningAccelerator implements hardware Q-learning.
type QLearningAccelerator struct {
	Config      *QLearningConfig
	QTable      [][]float64          // Q-values stored in crossbar
	Crossbar    [][]*MemristorSynapse // Memristor implementation
	Epsilon     float64
	Stats       *QLearningStats
}

// QLearningStats tracks Q-learning performance.
type QLearningStats struct {
	ThroughputMSPS   float64 // Million samples per second
	PowerMW          float64 // Milliwatts
	TotalUpdates     int64
	ConvergenceSteps int
}

// NewQLearningAccelerator creates a new Q-learning accelerator.
func NewQLearningAccelerator(config *QLearningConfig) *QLearningAccelerator {
	qa := &QLearningAccelerator{
		Config:  config,
		QTable:  make([][]float64, config.StateCount),
		Epsilon: config.InitialEpsilon,
		Stats:   &QLearningStats{},
	}

	// Initialize Q-table
	for s := 0; s < config.StateCount; s++ {
		qa.QTable[s] = make([]float64, config.ActionCount)
		for a := 0; a < config.ActionCount; a++ {
			qa.QTable[s][a] = rand.Float64() * 0.01 // Small random init
		}
	}

	// Initialize memristor crossbar if enabled
	if config.UseCrossbar {
		qa.Crossbar = make([][]*MemristorSynapse, config.StateCount)
		for s := 0; s < config.StateCount; s++ {
			qa.Crossbar[s] = make([]*MemristorSynapse, config.ActionCount)
			for a := 0; a < config.ActionCount; a++ {
				qa.Crossbar[s][a] = NewMemristorSynapse()
				qa.Crossbar[s][a].Conductance = qa.QTable[s][a] + 0.5
			}
		}
	}

	// Performance estimates based on literature
	// 8×4 Q-Matrix: 222 MSPS, 37 mW
	// 256×16 Q-Matrix: 93 MSPS, 611 mW
	if config.StateCount <= 8 && config.ActionCount <= 4 {
		qa.Stats.ThroughputMSPS = 222
		qa.Stats.PowerMW = 37
	} else if config.StateCount <= 256 && config.ActionCount <= 16 {
		qa.Stats.ThroughputMSPS = 93
		qa.Stats.PowerMW = 611
	} else {
		// Extrapolate
		qa.Stats.ThroughputMSPS = 50
		qa.Stats.PowerMW = 1000
	}

	return qa
}

// SelectAction chooses action using epsilon-greedy policy.
func (qa *QLearningAccelerator) SelectAction(state int) int {
	if rand.Float64() < qa.Epsilon {
		return rand.Intn(qa.Config.ActionCount)
	}

	// Find action with max Q-value
	maxQ := qa.QTable[state][0]
	maxAction := 0
	for a := 1; a < qa.Config.ActionCount; a++ {
		if qa.QTable[state][a] > maxQ {
			maxQ = qa.QTable[state][a]
			maxAction = a
		}
	}
	return maxAction
}

// Update performs Q-learning update.
func (qa *QLearningAccelerator) Update(state, action int, reward float64, nextState int, done bool) {
	// Find max Q(s', a')
	maxNextQ := 0.0
	if !done {
		maxNextQ = qa.QTable[nextState][0]
		for a := 1; a < qa.Config.ActionCount; a++ {
			if qa.QTable[nextState][a] > maxNextQ {
				maxNextQ = qa.QTable[nextState][a]
			}
		}
	}

	// Q-learning update
	target := reward + qa.Config.DiscountFactor*maxNextQ
	tdError := target - qa.QTable[state][action]
	qa.QTable[state][action] += qa.Config.LearningRate * tdError

	// Update memristor if enabled
	if qa.Config.UseCrossbar && qa.Crossbar != nil {
		qa.Crossbar[state][action].Conductance = qa.QTable[state][action] + 0.5
	}

	qa.Stats.TotalUpdates++

	// Decay epsilon
	if qa.Epsilon > qa.Config.MinEpsilon {
		qa.Epsilon *= qa.Config.EpsilonDecay
	}
}

// GetQValue retrieves Q-value for state-action pair.
func (qa *QLearningAccelerator) GetQValue(state, action int) float64 {
	return qa.QTable[state][action]
}

// =============================================================================
// 3D MONOLITHIC CIM INTEGRATION (M3D-LIME STYLE)
// =============================================================================

// M3DLayerType specifies the layer type in 3D stack.
type M3DLayerType int

const (
	M3DLayerSiCMOS M3DLayerType = iota // Silicon CMOS logic
	M3DLayerAnalogCIM                   // Analog RRAM CIM
	M3DLayerDigitalCIM                  // Binary RRAM CIM
	M3DLayerTCAM                        // Ternary CAM
	M3DLayerBuffer                      // On-chip buffer (eDRAM/SRAM)
	M3DLayerFeFET                       // Ferroelectric FET CIM
)

// M3DLayer represents a layer in the 3D stack.
type M3DLayer struct {
	Type            M3DLayerType
	Technology      string   // e.g., "HfAlOx RRAM", "CNTFET", "Si CMOS"
	Thickness       float64  // nm
	Area            float64  // mm²
	ArraySize       int      // Rows/cols for memory arrays
	PowerDensity    float64  // W/mm²
	ThermalResist   float64  // K·mm²/W
}

// M3DCIMChip represents a 3D monolithic CIM chip.
type M3DCIMChip struct {
	Config        *M3DCIMConfig
	Layers        []*M3DLayer
	Interconnects []*InterlayerVia
	ThermalModel  *ThermalModel
	Stats         *M3DStats
}

// M3DCIMConfig configures the 3D CIM chip.
type M3DCIMConfig struct {
	TotalLayers      int
	ChipArea         float64 // mm²
	MaxTempC         float64 // Maximum allowed temperature
	ProcessNode      int     // nm
	ILVDensity       int     // Interlayer vias per mm²
	HybridBonding    bool
}

// InterlayerVia represents vertical interconnects.
type InterlayerVia struct {
	FromLayer   int
	ToLayer     int
	Count       int
	PitchUM     float64 // Pitch in micrometers
	ResistOhm   float64 // Via resistance
	CapacitanceFf float64 // Via capacitance
}

// ThermalModel simulates chip thermal behavior.
type ThermalModel struct {
	AmbientTempC    float64
	LayerTemps      []float64
	HotspotTempC    float64
	CoolingCapacity float64 // W
}

// M3DStats tracks 3D chip performance.
type M3DStats struct {
	TotalPowerW       float64
	EnergyEffTOPSW    float64
	AreaEffTOPSmm2    float64
	PeakTempC         float64
	Bandwidth         float64 // GB/s
	Latency           float64 // ns
}

// NewM3DCIMChip creates a new 3D monolithic CIM chip.
func NewM3DCIMChip(config *M3DCIMConfig) *M3DCIMChip {
	chip := &M3DCIMChip{
		Config: config,
		Layers: make([]*M3DLayer, config.TotalLayers),
		Stats:  &M3DStats{},
		ThermalModel: &ThermalModel{
			AmbientTempC:    25.0,
			LayerTemps:      make([]float64, config.TotalLayers),
			CoolingCapacity: 5.0, // 5W typical for edge
		},
	}

	// Initialize default M3D-LIME style stack
	// Layer 0: Si CMOS control logic
	// Layer 1: Analog RRAM CIM
	// Layer 2: Buffer + TCAM

	for i := 0; i < config.TotalLayers; i++ {
		switch i {
		case 0:
			chip.Layers[i] = &M3DLayer{
				Type:          M3DLayerSiCMOS,
				Technology:    "Si CMOS",
				Thickness:     300,
				Area:          config.ChipArea,
				ArraySize:     0, // Logic, no array
				PowerDensity:  0.5,
				ThermalResist: 50,
			}
		case 1:
			chip.Layers[i] = &M3DLayer{
				Type:          M3DLayerAnalogCIM,
				Technology:    "HfAlOx RRAM",
				Thickness:     50,
				Area:          config.ChipArea * 0.8,
				ArraySize:     256,
				PowerDensity:  0.1,
				ThermalResist: 100,
			}
		default:
			chip.Layers[i] = &M3DLayer{
				Type:          M3DLayerBuffer,
				Technology:    "Ta2O5 RRAM + CNTFET",
				Thickness:     50,
				Area:          config.ChipArea * 0.6,
				ArraySize:     128,
				PowerDensity:  0.05,
				ThermalResist: 100,
			}
		}

		chip.ThermalModel.LayerTemps[i] = chip.ThermalModel.AmbientTempC
	}

	// Create interlayer vias
	chip.Interconnects = make([]*InterlayerVia, config.TotalLayers-1)
	for i := 0; i < config.TotalLayers-1; i++ {
		chip.Interconnects[i] = &InterlayerVia{
			FromLayer:     i,
			ToLayer:       i + 1,
			Count:         config.ILVDensity * int(config.ChipArea),
			PitchUM:       1.0,
			ResistOhm:     0.1,
			CapacitanceFf: 0.5,
		}
	}

	chip.calculateStats()
	return chip
}

// AddLayer adds a custom layer to the 3D stack.
func (chip *M3DCIMChip) AddLayer(layer *M3DLayer, position int) error {
	if position < 0 || position > len(chip.Layers) {
		return fmt.Errorf("invalid position %d", position)
	}

	// Insert layer
	newLayers := make([]*M3DLayer, len(chip.Layers)+1)
	copy(newLayers[:position], chip.Layers[:position])
	newLayers[position] = layer
	copy(newLayers[position+1:], chip.Layers[position:])
	chip.Layers = newLayers

	// Update thermal model
	chip.ThermalModel.LayerTemps = append(chip.ThermalModel.LayerTemps, chip.ThermalModel.AmbientTempC)

	chip.calculateStats()
	return nil
}

// calculateStats computes chip performance metrics.
func (chip *M3DCIMChip) calculateStats() {
	totalPower := 0.0
	for _, layer := range chip.Layers {
		totalPower += layer.PowerDensity * layer.Area
	}
	chip.Stats.TotalPowerW = totalPower

	// Energy efficiency estimate (based on M3D-LIME: 18.3× vs GPU)
	// Assuming GPU baseline of ~1 TOPS/W
	chip.Stats.EnergyEffTOPSW = 18.3 // TOPS/W

	// Area efficiency
	totalArea := chip.Config.ChipArea
	topsCapacity := 10.0 // Assumed
	chip.Stats.AreaEffTOPSmm2 = topsCapacity / totalArea

	// Bandwidth (based on ILV count and frequency)
	bandwidth := float64(len(chip.Interconnects)) * float64(chip.Config.ILVDensity) * 0.001 // GB/s
	chip.Stats.Bandwidth = bandwidth

	// Latency (simplified)
	chip.Stats.Latency = 10.0 // ns

	// Thermal simulation
	chip.simulateThermal()
}

// simulateThermal performs thermal analysis.
func (chip *M3DCIMChip) simulateThermal() {
	// Simple 1D thermal model
	hotspot := chip.ThermalModel.AmbientTempC

	for i := len(chip.Layers) - 1; i >= 0; i-- {
		layer := chip.Layers[i]
		powerDissipation := layer.PowerDensity * layer.Area
		tempRise := powerDissipation * layer.ThermalResist / layer.Area
		chip.ThermalModel.LayerTemps[i] = hotspot + tempRise
		hotspot = chip.ThermalModel.LayerTemps[i]
	}

	chip.ThermalModel.HotspotTempC = hotspot
	chip.Stats.PeakTempC = hotspot
}

// =============================================================================
// HETEROGENEOUS CIM CHIPLET ARCHITECTURE (3D-CIMlet)
// =============================================================================

// ChipletType specifies the chiplet memory technology.
type ChipletType int

const (
	ChipletRRAM     ChipletType = iota // RRAM-based CIM
	ChipletFeFET                        // Ferroelectric FET
	ChipletEDRAM                        // Embedded DRAM (capacitor-less)
	ChipletSRAM                         // SRAM CIM
	ChipletHybrid                       // Mixed technology
)

// CIMChiplet represents a single CIM chiplet.
type CIMChiplet struct {
	ID              int
	Type            ChipletType
	ProcessNode     int     // nm
	Area            float64 // mm²
	ArrayRows       int
	ArrayCols       int
	Precision       int     // bits
	EnergyPerMAC    float64 // pJ
	Latency         float64 // ns
	Endurance       float64 // cycles
	RetentionYears  float64
	Position        ChipletPosition
	ThermalLimit    float64 // °C
}

// ChipletPosition specifies 2.5D/3D placement.
type ChipletPosition struct {
	X, Y   float64 // mm from origin
	Layer  int     // 0 for 2.5D, 1+ for 3D stacking
}

// Interposer represents the silicon interposer.
type Interposer struct {
	Width           float64 // mm
	Height          float64 // mm
	Technology      string  // e.g., "65nm passive", "28nm active"
	TSVDensity      int     // TSVs per mm²
	MetalLayers     int
	MicrobumpPitch  float64 // µm
}

// ChipletSystem represents the complete chiplet-based CIM system.
type ChipletSystem struct {
	Config       *ChipletSystemConfig
	Chiplets     []*CIMChiplet
	Interposer   *Interposer
	Interconnect *ChipletInterconnect
	ThermalMap   [][]float64 // 2D thermal distribution
	Stats        *ChipletSystemStats
}

// ChipletSystemConfig configures the chiplet system.
type ChipletSystemConfig struct {
	MaxChiplets       int
	InterposerArea    float64 // mm²
	Is3DStacked       bool
	MaxPowerW         float64
	MaxTempC          float64
	TargetWorkload    string // e.g., "LLM", "CNN", "ViT"
}

// ChipletInterconnect models inter-chiplet communication.
type ChipletInterconnect struct {
	Type            string  // "UCIe", "BoW", "custom"
	BandwidthGBps   float64
	LatencyNS       float64
	EnergyPerBitPJ  float64
}

// ChipletSystemStats tracks system performance.
type ChipletSystemStats struct {
	TotalTOPS         float64
	EnergyEfficiency  float64 // TOPS/W
	AreaEfficiency    float64 // TOPS/mm²
	MemoryCapacity    float64 // GB
	PeakBandwidth     float64 // GB/s
	ThermalHeadroom   float64 // °C below limit
	EDPImprovement    float64 // vs 2D baseline
}

// NewChipletSystem creates a heterogeneous CIM chiplet system.
func NewChipletSystem(config *ChipletSystemConfig) *ChipletSystem {
	system := &ChipletSystem{
		Config:   config,
		Chiplets: make([]*CIMChiplet, 0),
		Interposer: &Interposer{
			Width:          math.Sqrt(config.InterposerArea),
			Height:         math.Sqrt(config.InterposerArea),
			Technology:     "65nm passive",
			TSVDensity:     10000,
			MetalLayers:    4,
			MicrobumpPitch: 55,
		},
		Interconnect: &ChipletInterconnect{
			Type:           "UCIe",
			BandwidthGBps:  256,
			LatencyNS:      2,
			EnergyPerBitPJ: 0.5,
		},
		Stats: &ChipletSystemStats{},
	}

	// Initialize thermal map
	gridSize := 10
	system.ThermalMap = make([][]float64, gridSize)
	for i := range system.ThermalMap {
		system.ThermalMap[i] = make([]float64, gridSize)
		for j := range system.ThermalMap[i] {
			system.ThermalMap[i][j] = 25.0 // Ambient
		}
	}

	return system
}

// AddChiplet adds a chiplet to the system.
func (sys *ChipletSystem) AddChiplet(chiplet *CIMChiplet) error {
	if len(sys.Chiplets) >= sys.Config.MaxChiplets {
		return fmt.Errorf("maximum chiplets (%d) reached", sys.Config.MaxChiplets)
	}

	// Assign ID
	chiplet.ID = len(sys.Chiplets)

	// Auto-place if position not set
	if chiplet.Position.X == 0 && chiplet.Position.Y == 0 {
		sys.autoPlaceChiplet(chiplet)
	}

	sys.Chiplets = append(sys.Chiplets, chiplet)
	sys.calculateStats()
	return nil
}

// autoPlaceChiplet finds optimal placement.
func (sys *ChipletSystem) autoPlaceChiplet(chiplet *CIMChiplet) {
	// Simple grid placement
	n := len(sys.Chiplets)
	gridCols := int(math.Ceil(math.Sqrt(float64(sys.Config.MaxChiplets))))

	row := n / gridCols
	col := n % gridCols

	spacing := sys.Interposer.Width / float64(gridCols)
	chiplet.Position.X = float64(col)*spacing + spacing/2
	chiplet.Position.Y = float64(row)*spacing + spacing/2

	if sys.Config.Is3DStacked && n > 0 {
		// Alternate layers for thermal distribution
		chiplet.Position.Layer = n % 2
	}
}

// calculateStats computes system performance.
func (sys *ChipletSystem) calculateStats() {
	if len(sys.Chiplets) == 0 {
		return
	}

	totalTOPS := 0.0
	totalPower := 0.0
	totalArea := 0.0
	totalMemory := 0.0

	for _, c := range sys.Chiplets {
		// TOPS = (rows × cols × 2) / latency_ns × 1e-3
		macPerCycle := float64(c.ArrayRows * c.ArrayCols)
		ops := macPerCycle * 2 / c.Latency * 1e-3 // TOPS
		totalTOPS += ops

		// Power from energy per MAC
		power := c.EnergyPerMAC * macPerCycle / c.Latency // mW
		totalPower += power / 1000                        // W

		totalArea += c.Area

		// Memory capacity (weights)
		memBits := float64(c.ArrayRows * c.ArrayCols * c.Precision)
		totalMemory += memBits / 8 / 1e9 // GB
	}

	sys.Stats.TotalTOPS = totalTOPS
	sys.Stats.MemoryCapacity = totalMemory

	if totalPower > 0 {
		sys.Stats.EnergyEfficiency = totalTOPS / totalPower
	}
	if totalArea > 0 {
		sys.Stats.AreaEfficiency = totalTOPS / totalArea
	}

	sys.Stats.PeakBandwidth = sys.Interconnect.BandwidthGBps * float64(len(sys.Chiplets))

	// Thermal headroom
	maxTemp := 25.0
	for _, c := range sys.Chiplets {
		chipTemp := 25 + c.EnergyPerMAC*float64(c.ArrayRows*c.ArrayCols)/c.Area*10
		if chipTemp > maxTemp {
			maxTemp = chipTemp
		}
	}
	sys.Stats.ThermalHeadroom = sys.Config.MaxTempC - maxTemp

	// 3D-CIMlet EDP improvement (from literature: up to 12× vs 2D)
	if sys.Config.Is3DStacked {
		sys.Stats.EDPImprovement = 12.0
	} else {
		sys.Stats.EDPImprovement = 9.3 // 2.5D
	}
}

// CreateHeterogeneousStack creates a typical heterogeneous 3D-CIMlet configuration.
func (sys *ChipletSystem) CreateHeterogeneousStack() {
	// RRAM chiplet for weight storage
	sys.AddChiplet(&CIMChiplet{
		Type:           ChipletRRAM,
		ProcessNode:    28,
		Area:           4.0,
		ArrayRows:      256,
		ArrayCols:      256,
		Precision:      8,
		EnergyPerMAC:   0.3,
		Latency:        5.0,
		Endurance:      1e9,
		RetentionYears: 10,
		ThermalLimit:   85,
	})

	// FeFET chiplet for inference
	sys.AddChiplet(&CIMChiplet{
		Type:           ChipletFeFET,
		ProcessNode:    14,
		Area:           2.0,
		ArrayRows:      128,
		ArrayCols:      128,
		Precision:      4,
		EnergyPerMAC:   0.1,
		Latency:        2.0,
		Endurance:      1e10,
		RetentionYears: 10,
		ThermalLimit:   85,
	})

	// eDRAM chiplet for activations
	sys.AddChiplet(&CIMChiplet{
		Type:           ChipletEDRAM,
		ProcessNode:    7,
		Area:           1.5,
		ArrayRows:      512,
		ArrayCols:      512,
		Precision:      16,
		EnergyPerMAC:   0.05,
		Latency:        1.0,
		Endurance:      1e15, // DRAM endurance
		RetentionYears: 0.01, // Needs refresh
		ThermalLimit:   95,
	})

	// SRAM chiplet for digital operations
	sys.AddChiplet(&CIMChiplet{
		Type:           ChipletSRAM,
		ProcessNode:    5,
		Area:           1.0,
		ArrayRows:      64,
		ArrayCols:      64,
		Precision:      8,
		EnergyPerMAC:   0.8,
		Latency:        0.5,
		Endurance:      1e18, // Effectively unlimited
		RetentionYears: 100,  // With power
		ThermalLimit:   100,
	})
}

// =============================================================================
// DEMONSTRATION AND BENCHMARKING
// =============================================================================

// RunRLDemo demonstrates the RL-CIM capabilities.
func RunRLDemo() {
	fmt.Println("=== Reinforcement Learning CIM Demo ===")
	fmt.Println()

	// Create actor-critic system
	config := &RLConfig{
		StateSize:       4,
		ActionSize:      2,
		HiddenSize:      16,
		LearningRate:    0.01,
		DiscountFactor:  0.99,
		TDLambda:        0.9,
		ExplorationRate: 0.1,
		UseMemristor:    true,
		CrossbarSize:    64,
	}

	ac := NewActorCriticCIM(config)

	// Simulate CartPole-like environment
	episodes := 100
	maxSteps := 200

	for ep := 0; ep < episodes; ep++ {
		state := []float64{rand.Float64()*0.1 - 0.05, rand.Float64()*0.1 - 0.05,
			rand.Float64()*0.1 - 0.05, rand.Float64()*0.1 - 0.05}
		totalReward := 0.0

		for step := 0; step < maxSteps; step++ {
			// Get action
			actionProbs, value := ac.Forward(state)
			action := ac.SelectAction(actionProbs)

			// Simulate environment (simplified)
			reward := 1.0
			done := rand.Float64() < 0.01 // Small chance of termination
			nextState := make([]float64, 4)
			for i := range nextState {
				nextState[i] = state[i] + rand.Float64()*0.1 - 0.05
			}

			// Get next value
			_, nextValue := ac.Forward(nextState)

			// Compute TD error and update
			tdError := ac.ComputeTDError(reward, nextValue, done)
			ac.UpdateWeights(state, action, tdError)

			totalReward += reward
			state = nextState

			if done {
				break
			}
		}

		ac.Stats.TotalReward += totalReward
		ac.Stats.Episodes++
	}

	ac.Stats.AverageReturn = ac.Stats.TotalReward / float64(ac.Stats.Episodes)

	fmt.Printf("Actor-Critic Results:\n")
	fmt.Printf("  Episodes: %d\n", ac.Stats.Episodes)
	fmt.Printf("  Average Return: %.1f\n", ac.Stats.AverageReturn)
	fmt.Printf("  Weight Updates: %d\n", ac.Stats.WeightUpdateCount)
	fmt.Printf("  Energy per Update: %.2f pJ\n", ac.Stats.EnergyPerUpdate)
	fmt.Println()

	// Q-Learning accelerator
	qConfig := &QLearningConfig{
		StateCount:     64,
		ActionCount:    4,
		LearningRate:   0.1,
		DiscountFactor: 0.99,
		InitialEpsilon: 0.5,
		EpsilonDecay:   0.995,
		MinEpsilon:     0.01,
		UseCrossbar:    true,
	}

	qa := NewQLearningAccelerator(qConfig)

	// Simulate grid world
	for ep := 0; ep < 1000; ep++ {
		state := rand.Intn(qConfig.StateCount)
		for step := 0; step < 100; step++ {
			action := qa.SelectAction(state)
			nextState := (state + action + 1) % qConfig.StateCount
			reward := 0.0
			if nextState == qConfig.StateCount-1 {
				reward = 1.0
			}
			done := nextState == qConfig.StateCount-1
			qa.Update(state, action, reward, nextState, done)
			state = nextState
			if done {
				break
			}
		}
	}

	fmt.Printf("Q-Learning Accelerator Results:\n")
	fmt.Printf("  Throughput: %.0f MSPS\n", qa.Stats.ThroughputMSPS)
	fmt.Printf("  Power: %.0f mW\n", qa.Stats.PowerMW)
	fmt.Printf("  Total Updates: %d\n", qa.Stats.TotalUpdates)
	fmt.Printf("  Final Epsilon: %.4f\n", qa.Epsilon)
	fmt.Println()
}

// Run3DChipletDemo demonstrates the 3D CIM chiplet capabilities.
func Run3DChipletDemo() {
	fmt.Println("=== 3D CIM Chiplet Demo ===")
	fmt.Println()

	// Create M3D-LIME style chip
	m3dConfig := &M3DCIMConfig{
		TotalLayers:   3,
		ChipArea:      9.0, // 3×3 mm
		MaxTempC:      85.0,
		ProcessNode:   28,
		ILVDensity:    100000,
		HybridBonding: true,
	}

	m3dChip := NewM3DCIMChip(m3dConfig)

	fmt.Printf("M3D-LIME Style Chip:\n")
	fmt.Printf("  Layers: %d\n", len(m3dChip.Layers))
	for i, layer := range m3dChip.Layers {
		fmt.Printf("    Layer %d: %s (%s)\n", i, layer.Type, layer.Technology)
	}
	fmt.Printf("  Total Power: %.2f W\n", m3dChip.Stats.TotalPowerW)
	fmt.Printf("  Energy Efficiency: %.1f TOPS/W\n", m3dChip.Stats.EnergyEffTOPSW)
	fmt.Printf("  Peak Temperature: %.1f°C\n", m3dChip.Stats.PeakTempC)
	fmt.Println()

	// Create heterogeneous chiplet system
	chipletConfig := &ChipletSystemConfig{
		MaxChiplets:    8,
		InterposerArea: 400, // 20×20 mm
		Is3DStacked:    true,
		MaxPowerW:      15,
		MaxTempC:       85,
		TargetWorkload: "Edge LLM",
	}

	system := NewChipletSystem(chipletConfig)
	system.CreateHeterogeneousStack()

	fmt.Printf("Heterogeneous CIM Chiplet System:\n")
	fmt.Printf("  Chiplets: %d\n", len(system.Chiplets))
	for _, c := range system.Chiplets {
		fmt.Printf("    [%d] %v: %dnm, %.1fmm², %dx%d\n",
			c.ID, c.Type, c.ProcessNode, c.Area, c.ArrayRows, c.ArrayCols)
	}
	fmt.Printf("  Total TOPS: %.2f\n", system.Stats.TotalTOPS)
	fmt.Printf("  Energy Efficiency: %.1f TOPS/W\n", system.Stats.EnergyEfficiency)
	fmt.Printf("  Memory Capacity: %.3f GB\n", system.Stats.MemoryCapacity)
	fmt.Printf("  Thermal Headroom: %.1f°C\n", system.Stats.ThermalHeadroom)
	fmt.Printf("  EDP Improvement vs 2D: %.1fx\n", system.Stats.EDPImprovement)
	fmt.Println()
}

// Ensure sort is used
var _ = sort.Ints
