// Package layers provides neural network layer implementations for CIM inference.
// domain_wall.go implements ferroelectric domain wall computing and on-chip learning.
//
// References:
// - LiNbO3 domain wall memristors: 95.1% MNIST, 4-state memory
// - AlScN nitride FeDMEM: 1-2 nm DW width, 10³ ON/OFF
// - Equilibrium Propagation: local learning for memristor crossbars
// - STDP: bio-inspired on-chip learning

package layers

import (
	"fmt"
	"math"
)

// =============================================================================
// Ferroelectric Domain Wall Computing
// =============================================================================

// DomainWallMaterial defines ferroelectric materials for DW devices
type DomainWallMaterial string

const (
	DWMaterialLiNbO3 DomainWallMaterial = "LiNbO3"
	DWMaterialBFO    DomainWallMaterial = "BiFeO3"
	DWMaterialAlScN  DomainWallMaterial = "AlScN"
	DWMaterialPZT    DomainWallMaterial = "PZT"
	DWMaterialBTO    DomainWallMaterial = "BaTiO3"
)

// DomainWallConfig configures domain wall device properties
type DomainWallConfig struct {
	Material          DomainWallMaterial
	ThicknessNm       float64 // Film thickness
	DomainWallWidthNm float64 // Typically 1-2 nm

	// Electrical properties
	OnOffRatio        float64 // Conductance ratio
	DWConductivity    float64 // S/m for conducting DW
	BulkResistivity   float64 // Ω·m for insulating bulk

	// Switching characteristics
	CoerciveFieldMV   float64 // MV/cm
	SwitchingSpeedNs  float64
	NumStates         int     // Multi-level states

	// Device geometry
	ChannelLengthNm   float64
	ChannelWidthNm    float64

	// Reliability
	Endurance         float64
	RetentionHours    float64
}

// GetDomainWallConfig returns material-specific configuration
func GetDomainWallConfig(material DomainWallMaterial) *DomainWallConfig {
	switch material {
	case DWMaterialLiNbO3:
		return &DomainWallConfig{
			Material:          DWMaterialLiNbO3,
			ThicknessNm:       100.0,
			DomainWallWidthNm: 1.0,
			OnOffRatio:        1e4,
			DWConductivity:    1e-3,
			BulkResistivity:   1e14,
			CoerciveFieldMV:   0.2,
			SwitchingSpeedNs:  100.0,
			NumStates:         4, // 4-state demonstrated
			ChannelLengthNm:   500.0,
			ChannelWidthNm:    100.0,
			Endurance:         1e8,
			RetentionHours:    1000,
		}
	case DWMaterialAlScN:
		return &DomainWallConfig{
			Material:          DWMaterialAlScN,
			ThicknessNm:       50.0,
			DomainWallWidthNm: 1.5,
			OnOffRatio:        1e3,
			DWConductivity:    1e-2,
			BulkResistivity:   1e12,
			CoerciveFieldMV:   5.0, // Higher for nitrides
			SwitchingSpeedNs:  10.0,
			NumStates:         8,
			ChannelLengthNm:   100.0,
			ChannelWidthNm:    50.0,
			Endurance:         1e10,
			RetentionHours:    10000,
		}
	case DWMaterialBFO:
		return &DomainWallConfig{
			Material:          DWMaterialBFO,
			ThicknessNm:       200.0,
			DomainWallWidthNm: 2.0,
			OnOffRatio:        1e5,
			DWConductivity:    1e-4,
			BulkResistivity:   1e10,
			CoerciveFieldMV:   0.3,
			SwitchingSpeedNs:  50.0,
			NumStates:         16,
			ChannelLengthNm:   1000.0,
			ChannelWidthNm:    200.0,
			Endurance:         1e7,
			RetentionHours:    5000,
		}
	default:
		return GetDomainWallConfig(DWMaterialLiNbO3)
	}
}

// DomainWallDevice simulates a ferroelectric domain wall memristor
type DomainWallDevice struct {
	Config            *DomainWallConfig
	NumDomainWalls    int       // Number of conducting DWs
	DomainWallPositions []float64 // Position along channel
	CurrentState      int       // Multi-level state index
	Conductance       float64   // Current conductance
}

// NewDomainWallDevice creates a new DW device
func NewDomainWallDevice(config *DomainWallConfig) *DomainWallDevice {
	if config == nil {
		config = GetDomainWallConfig(DWMaterialLiNbO3)
	}
	return &DomainWallDevice{
		Config:              config,
		DomainWallPositions: make([]float64, 0),
	}
}

// CreateDomainWall creates a new conducting domain wall
func (d *DomainWallDevice) CreateDomainWall(position float64) {
	// Check if position is valid
	if position < 0 || position > d.Config.ChannelLengthNm {
		return
	}

	// Add domain wall
	d.DomainWallPositions = append(d.DomainWallPositions, position)
	d.NumDomainWalls++
	d.updateConductance()
}

// AnnihilateDomainWall removes a domain wall
func (d *DomainWallDevice) AnnihilateDomainWall(index int) {
	if index < 0 || index >= len(d.DomainWallPositions) {
		return
	}

	// Remove domain wall
	d.DomainWallPositions = append(
		d.DomainWallPositions[:index],
		d.DomainWallPositions[index+1:]...,
	)
	d.NumDomainWalls--
	d.updateConductance()
}

// MoveDomainWall moves a DW to a new position (subcoercive field)
func (d *DomainWallDevice) MoveDomainWall(index int, newPosition float64) {
	if index < 0 || index >= len(d.DomainWallPositions) {
		return
	}
	if newPosition < 0 {
		newPosition = 0
	}
	if newPosition > d.Config.ChannelLengthNm {
		newPosition = d.Config.ChannelLengthNm
	}
	d.DomainWallPositions[index] = newPosition
	d.updateConductance()
}

// SetState sets the device to a specific multi-level state
func (d *DomainWallDevice) SetState(state int) {
	if state < 0 {
		state = 0
	}
	if state >= d.Config.NumStates {
		state = d.Config.NumStates - 1
	}

	d.CurrentState = state

	// Number of DWs determines state
	targetDWs := state
	d.DomainWallPositions = make([]float64, targetDWs)

	// Distribute DWs evenly along channel
	for i := 0; i < targetDWs; i++ {
		d.DomainWallPositions[i] = float64(i+1) * d.Config.ChannelLengthNm / float64(targetDWs+1)
	}
	d.NumDomainWalls = targetDWs
	d.updateConductance()
}

// updateConductance calculates device conductance from DW configuration
func (d *DomainWallDevice) updateConductance() {
	// Parallel conduction through domain walls
	// Each DW acts as a conductive filament

	if d.NumDomainWalls == 0 {
		// Only bulk conduction
		area := d.Config.ChannelWidthNm * d.Config.ThicknessNm * 1e-18 // m²
		d.Conductance = area / (d.Config.BulkResistivity * d.Config.ChannelLengthNm * 1e-9)
	} else {
		// DW conduction dominates
		// G = σ_DW × A_DW × N_DW / L
		dwArea := d.Config.DomainWallWidthNm * d.Config.ThicknessNm * 1e-18 // m²
		d.Conductance = d.Config.DWConductivity * dwArea * float64(d.NumDomainWalls) /
			(d.Config.ChannelLengthNm * 1e-9)
	}
}

// GetWeight returns normalized weight [0, 1]
func (d *DomainWallDevice) GetWeight() float64 {
	return float64(d.CurrentState) / float64(d.Config.NumStates-1)
}

// Read performs non-destructive read
func (d *DomainWallDevice) Read(voltage float64) float64 {
	// I = G × V
	return d.Conductance * voltage
}

// EstimateEnergy returns energy per switching event
func (d *DomainWallDevice) EstimateEnergy() float64 {
	// Energy to create/annihilate one DW
	// E ~ E_c × P_s × Volume
	coercive := d.Config.CoerciveFieldMV * 1e8 // V/m
	// Assume polarization ~20 µC/cm²
	polarization := 20e-2 // C/m²
	volume := d.Config.DomainWallWidthNm * d.Config.ChannelWidthNm * d.Config.ThicknessNm * 1e-27 // m³

	return coercive * polarization * volume // Joules
}

// =============================================================================
// Domain Wall Crossbar Array
// =============================================================================

// DomainWallCrossbar implements a crossbar array using DW devices
type DomainWallCrossbar struct {
	Rows      int
	Cols      int
	Devices   [][]*DomainWallDevice
	Config    *DomainWallConfig
}

// NewDomainWallCrossbar creates a DW-based crossbar
func NewDomainWallCrossbar(rows, cols int, material DomainWallMaterial) *DomainWallCrossbar {
	config := GetDomainWallConfig(material)
	crossbar := &DomainWallCrossbar{
		Rows:    rows,
		Cols:    cols,
		Devices: make([][]*DomainWallDevice, rows),
		Config:  config,
	}

	for i := 0; i < rows; i++ {
		crossbar.Devices[i] = make([]*DomainWallDevice, cols)
		for j := 0; j < cols; j++ {
			crossbar.Devices[i][j] = NewDomainWallDevice(config)
		}
	}

	return crossbar
}

// ProgramWeights programs weight matrix into crossbar
func (c *DomainWallCrossbar) ProgramWeights(weights [][]float64) {
	for i := 0; i < c.Rows && i < len(weights); i++ {
		for j := 0; j < c.Cols && j < len(weights[i]); j++ {
			// Map weight to state
			w := weights[i][j]
			if w < 0 {
				w = 0
			}
			if w > 1 {
				w = 1
			}
			state := int(w * float64(c.Config.NumStates-1))
			c.Devices[i][j].SetState(state)
		}
	}
}

// Forward performs matrix-vector multiplication
func (c *DomainWallCrossbar) Forward(input []float64, readVoltage float64) []float64 {
	output := make([]float64, c.Rows)

	for i := 0; i < c.Rows; i++ {
		sum := 0.0
		for j := 0; j < c.Cols && j < len(input); j++ {
			// Current = G × V_input
			current := c.Devices[i][j].Read(input[j] * readVoltage)
			sum += current
		}
		output[i] = sum
	}

	return output
}

// =============================================================================
// On-Chip Learning Algorithms
// =============================================================================

// LearningAlgorithm defines the type of on-chip learning
type LearningAlgorithm string

const (
	LearningSTDP     LearningAlgorithm = "stdp"
	LearningEP       LearningAlgorithm = "equilibrium_propagation"
	LearningDFA      LearningAlgorithm = "direct_feedback_alignment"
	LearningHebbian  LearningAlgorithm = "hebbian"
	LearningBCM      LearningAlgorithm = "bcm" // Bienenstock-Cooper-Munro
)

// OnChipLearnerConfig configures on-chip learning
type OnChipLearnerConfig struct {
	Algorithm         LearningAlgorithm
	LearningRate      float64
	WeightDecay       float64

	// STDP parameters
	TauPlus           float64 // LTP time constant (ms)
	TauMinus          float64 // LTD time constant (ms)
	APlus             float64 // LTP amplitude
	AMinus            float64 // LTD amplitude

	// EP parameters
	Beta              float64 // Nudge strength
	NumFreePhaseIter  int     // Free phase iterations
	NumNudgePhaseIter int     // Nudge phase iterations

	// Hardware constraints
	WeightBits        int
	UpdateNoise       float64
}

// DefaultOnChipLearnerConfig returns typical configuration
func DefaultOnChipLearnerConfig(algorithm LearningAlgorithm) *OnChipLearnerConfig {
	config := &OnChipLearnerConfig{
		Algorithm:         algorithm,
		LearningRate:      0.01,
		WeightDecay:       1e-4,
		TauPlus:           20.0,
		TauMinus:          20.0,
		APlus:             0.1,
		AMinus:            0.1,
		Beta:              0.5,
		NumFreePhaseIter:  20,
		NumNudgePhaseIter: 4,
		WeightBits:        6,
		UpdateNoise:       0.02,
	}
	return config
}

// OnChipLearner implements on-chip learning for memristor arrays
type OnChipLearner struct {
	Config     *OnChipLearnerConfig
	Weights    [][]float64
	Activations [][]float64 // Per-layer activations
	SpikeTimes [][]float64 // For STDP
}

// NewOnChipLearner creates a new on-chip learner
func NewOnChipLearner(config *OnChipLearnerConfig) *OnChipLearner {
	if config == nil {
		config = DefaultOnChipLearnerConfig(LearningSTDP)
	}
	return &OnChipLearner{
		Config: config,
	}
}

// InitWeights initializes weight matrices
func (l *OnChipLearner) InitWeights(layerSizes []int) {
	numLayers := len(layerSizes) - 1
	l.Weights = make([][]float64, numLayers)
	l.Activations = make([][]float64, len(layerSizes))
	l.SpikeTimes = make([][]float64, len(layerSizes))

	for i := 0; i < numLayers; i++ {
		rows := layerSizes[i+1]
		cols := layerSizes[i]
		l.Weights[i] = make([]float64, rows*cols)

		// Xavier initialization
		scale := math.Sqrt(2.0 / float64(rows+cols))
		for j := range l.Weights[i] {
			l.Weights[i][j] = (randFloat()*2 - 1) * scale
		}
	}

	for i := range layerSizes {
		l.Activations[i] = make([]float64, layerSizes[i])
		l.SpikeTimes[i] = make([]float64, layerSizes[i])
	}
}

// =============================================================================
// STDP Learning
// =============================================================================

// STDPUpdate computes weight update using spike-timing-dependent plasticity
func (l *OnChipLearner) STDPUpdate(preSpikes, postSpikes []float64, currentTime float64) []float64 {
	numPre := len(preSpikes)
	numPost := len(postSpikes)
	deltaW := make([]float64, numPre*numPost)

	for i := 0; i < numPost; i++ {
		for j := 0; j < numPre; j++ {
			idx := i*numPre + j

			// Check if both neurons spiked
			if preSpikes[j] > 0 && postSpikes[i] > 0 {
				dt := postSpikes[i] - preSpikes[j]

				var dw float64
				if dt > 0 {
					// Pre before post -> LTP
					dw = l.Config.APlus * math.Exp(-dt/l.Config.TauPlus)
				} else {
					// Post before pre -> LTD
					dw = -l.Config.AMinus * math.Exp(dt/l.Config.TauMinus)
				}

				// Apply learning rate and noise
				dw *= l.Config.LearningRate
				if l.Config.UpdateNoise > 0 {
					dw *= (1.0 + (randFloat()-0.5)*l.Config.UpdateNoise*2)
				}

				deltaW[idx] = dw
			}
		}
	}

	return deltaW
}

// ApplySTDP applies STDP updates to weight matrix
func (l *OnChipLearner) ApplySTDP(layerIdx int, deltaW []float64) {
	if layerIdx >= len(l.Weights) {
		return
	}

	for i := range l.Weights[layerIdx] {
		if i < len(deltaW) {
			// Update weight
			l.Weights[layerIdx][i] += deltaW[i]

			// Apply weight decay
			l.Weights[layerIdx][i] *= (1.0 - l.Config.WeightDecay)

			// Clamp to valid range
			if l.Weights[layerIdx][i] < -1 {
				l.Weights[layerIdx][i] = -1
			}
			if l.Weights[layerIdx][i] > 1 {
				l.Weights[layerIdx][i] = 1
			}
		}
	}
}

// =============================================================================
// Equilibrium Propagation
// =============================================================================

// EPState holds network state for equilibrium propagation
type EPState struct {
	Neurons     [][]float64 // Neuron states per layer
	FreeEnergy  float64
	NudgeEnergy float64
}

// EquilibriumPropagation implements EP learning
type EquilibriumPropagation struct {
	Config      *OnChipLearnerConfig
	Weights     [][][]float64 // [layer][post][pre]
	Biases      [][]float64
	LayerSizes  []int
}

// NewEquilibriumPropagation creates an EP learner
func NewEquilibriumPropagation(layerSizes []int, config *OnChipLearnerConfig) *EquilibriumPropagation {
	if config == nil {
		config = DefaultOnChipLearnerConfig(LearningEP)
	}

	ep := &EquilibriumPropagation{
		Config:     config,
		LayerSizes: layerSizes,
		Weights:    make([][][]float64, len(layerSizes)-1),
		Biases:     make([][]float64, len(layerSizes)),
	}

	// Initialize weights
	for l := 0; l < len(layerSizes)-1; l++ {
		ep.Weights[l] = make([][]float64, layerSizes[l+1])
		for i := 0; i < layerSizes[l+1]; i++ {
			ep.Weights[l][i] = make([]float64, layerSizes[l])
			scale := math.Sqrt(2.0 / float64(layerSizes[l]+layerSizes[l+1]))
			for j := range ep.Weights[l][i] {
				ep.Weights[l][i][j] = (randFloat()*2 - 1) * scale
			}
		}
	}

	// Initialize biases
	for l := range layerSizes {
		ep.Biases[l] = make([]float64, layerSizes[l])
	}

	return ep
}

// FreePhase runs the free phase to find equilibrium
func (ep *EquilibriumPropagation) FreePhase(input []float64) *EPState {
	state := &EPState{
		Neurons: make([][]float64, len(ep.LayerSizes)),
	}

	// Initialize neuron states
	for l := range ep.LayerSizes {
		state.Neurons[l] = make([]float64, ep.LayerSizes[l])
	}

	// Clamp input
	copy(state.Neurons[0], input)

	// Iterate to equilibrium
	for iter := 0; iter < ep.Config.NumFreePhaseIter; iter++ {
		for l := 1; l < len(ep.LayerSizes); l++ {
			for i := range state.Neurons[l] {
				// Compute input from previous layer
				sum := ep.Biases[l][i]
				for j := range state.Neurons[l-1] {
					sum += ep.Weights[l-1][i][j] * state.Neurons[l-1][j]
				}

				// Add input from next layer (if exists) for symmetric connections
				if l < len(ep.LayerSizes)-1 {
					for k := range state.Neurons[l+1] {
						sum += ep.Weights[l][k][i] * state.Neurons[l+1][k]
					}
				}

				// Activation (hard sigmoid for analog implementation)
				state.Neurons[l][i] = hardSigmoid(sum)
			}
		}
	}

	// Compute free energy
	state.FreeEnergy = ep.computeEnergy(state.Neurons)

	return state
}

// NudgePhase runs the nudge phase with target clamping
func (ep *EquilibriumPropagation) NudgePhase(state *EPState, target []float64) *EPState {
	nudgedState := &EPState{
		Neurons: make([][]float64, len(ep.LayerSizes)),
	}

	// Copy current state
	for l := range state.Neurons {
		nudgedState.Neurons[l] = make([]float64, len(state.Neurons[l]))
		copy(nudgedState.Neurons[l], state.Neurons[l])
	}

	// Nudge output toward target
	outputLayer := len(ep.LayerSizes) - 1
	for iter := 0; iter < ep.Config.NumNudgePhaseIter; iter++ {
		// Update hidden layers
		for l := 1; l < len(ep.LayerSizes)-1; l++ {
			for i := range nudgedState.Neurons[l] {
				sum := ep.Biases[l][i]
				for j := range nudgedState.Neurons[l-1] {
					sum += ep.Weights[l-1][i][j] * nudgedState.Neurons[l-1][j]
				}
				if l < len(ep.LayerSizes)-1 {
					for k := range nudgedState.Neurons[l+1] {
						sum += ep.Weights[l][k][i] * nudgedState.Neurons[l+1][k]
					}
				}
				nudgedState.Neurons[l][i] = hardSigmoid(sum)
			}
		}

		// Nudge output layer toward target
		for i := range nudgedState.Neurons[outputLayer] {
			// Blend between free state and target
			sum := ep.Biases[outputLayer][i]
			for j := range nudgedState.Neurons[outputLayer-1] {
				sum += ep.Weights[outputLayer-1][i][j] * nudgedState.Neurons[outputLayer-1][j]
			}
			freeVal := hardSigmoid(sum)
			nudgedState.Neurons[outputLayer][i] = freeVal + ep.Config.Beta*(target[i]-freeVal)
		}
	}

	nudgedState.NudgeEnergy = ep.computeEnergy(nudgedState.Neurons)

	return nudgedState
}

// ComputeWeightUpdate calculates weight updates from free and nudged phases
func (ep *EquilibriumPropagation) ComputeWeightUpdate(freeState, nudgedState *EPState) [][][]float64 {
	deltaW := make([][][]float64, len(ep.Weights))

	for l := range ep.Weights {
		deltaW[l] = make([][]float64, len(ep.Weights[l]))
		for i := range ep.Weights[l] {
			deltaW[l][i] = make([]float64, len(ep.Weights[l][i]))
			for j := range ep.Weights[l][i] {
				// Contrastive Hebbian rule:
				// ΔW = η/β × (s_nudge × s_nudge - s_free × s_free)
				freeProd := freeState.Neurons[l+1][i] * freeState.Neurons[l][j]
				nudgeProd := nudgedState.Neurons[l+1][i] * nudgedState.Neurons[l][j]

				deltaW[l][i][j] = ep.Config.LearningRate / ep.Config.Beta * (nudgeProd - freeProd)
			}
		}
	}

	return deltaW
}

// ApplyUpdate applies weight updates
func (ep *EquilibriumPropagation) ApplyUpdate(deltaW [][][]float64) {
	for l := range ep.Weights {
		for i := range ep.Weights[l] {
			for j := range ep.Weights[l][i] {
				ep.Weights[l][i][j] += deltaW[l][i][j]

				// Weight decay
				ep.Weights[l][i][j] *= (1.0 - ep.Config.WeightDecay)

				// Clamp
				if ep.Weights[l][i][j] < -1 {
					ep.Weights[l][i][j] = -1
				}
				if ep.Weights[l][i][j] > 1 {
					ep.Weights[l][i][j] = 1
				}
			}
		}
	}
}

// computeEnergy calculates network energy (for convergence check)
func (ep *EquilibriumPropagation) computeEnergy(neurons [][]float64) float64 {
	energy := 0.0

	// Sum of squared activations (neuron cost)
	for l := range neurons {
		for _, n := range neurons[l] {
			energy += n * n
		}
	}

	// Interaction energy (from weights)
	for l := 0; l < len(ep.Weights); l++ {
		for i := range ep.Weights[l] {
			for j := range ep.Weights[l][i] {
				energy -= ep.Weights[l][i][j] * neurons[l+1][i] * neurons[l][j]
			}
		}
	}

	return energy
}

// hardSigmoid is an analog-friendly activation
func hardSigmoid(x float64) float64 {
	if x < -2.5 {
		return 0
	}
	if x > 2.5 {
		return 1
	}
	return 0.2*x + 0.5
}

// =============================================================================
// Direct Feedback Alignment
// =============================================================================

// DirectFeedbackAlignment implements DFA for on-chip learning
type DirectFeedbackAlignment struct {
	Config         *OnChipLearnerConfig
	Weights        [][][]float64 // Forward weights
	FeedbackWeights [][][]float64 // Random fixed feedback weights
	LayerSizes     []int
}

// NewDirectFeedbackAlignment creates a DFA learner
func NewDirectFeedbackAlignment(layerSizes []int, config *OnChipLearnerConfig) *DirectFeedbackAlignment {
	if config == nil {
		config = DefaultOnChipLearnerConfig(LearningDFA)
	}

	dfa := &DirectFeedbackAlignment{
		Config:          config,
		LayerSizes:      layerSizes,
		Weights:         make([][][]float64, len(layerSizes)-1),
		FeedbackWeights: make([][][]float64, len(layerSizes)-1),
	}

	outputSize := layerSizes[len(layerSizes)-1]

	for l := 0; l < len(layerSizes)-1; l++ {
		// Forward weights
		dfa.Weights[l] = make([][]float64, layerSizes[l+1])
		for i := range dfa.Weights[l] {
			dfa.Weights[l][i] = make([]float64, layerSizes[l])
			scale := math.Sqrt(2.0 / float64(layerSizes[l]))
			for j := range dfa.Weights[l][i] {
				dfa.Weights[l][i][j] = (randFloat()*2 - 1) * scale
			}
		}

		// Random fixed feedback weights (from output to each layer)
		dfa.FeedbackWeights[l] = make([][]float64, layerSizes[l+1])
		for i := range dfa.FeedbackWeights[l] {
			dfa.FeedbackWeights[l][i] = make([]float64, outputSize)
			for j := range dfa.FeedbackWeights[l][i] {
				dfa.FeedbackWeights[l][i][j] = (randFloat()*2 - 1) * 0.5
			}
		}
	}

	return dfa
}

// Forward runs forward pass
func (dfa *DirectFeedbackAlignment) Forward(input []float64) [][]float64 {
	activations := make([][]float64, len(dfa.LayerSizes))
	activations[0] = input

	for l := 0; l < len(dfa.Weights); l++ {
		activations[l+1] = make([]float64, dfa.LayerSizes[l+1])
		for i := range activations[l+1] {
			sum := 0.0
			for j := range activations[l] {
				sum += dfa.Weights[l][i][j] * activations[l][j]
			}
			activations[l+1][i] = math.Tanh(sum)
		}
	}

	return activations
}

// ComputeUpdate calculates weight updates using DFA
func (dfa *DirectFeedbackAlignment) ComputeUpdate(activations [][]float64, target []float64) [][][]float64 {
	outputLayer := len(dfa.LayerSizes) - 1
	deltaW := make([][][]float64, len(dfa.Weights))

	// Output error
	outputError := make([]float64, len(target))
	for i := range target {
		outputError[i] = target[i] - activations[outputLayer][i]
	}

	// Update each layer using direct feedback
	for l := 0; l < len(dfa.Weights); l++ {
		deltaW[l] = make([][]float64, len(dfa.Weights[l]))
		for i := range dfa.Weights[l] {
			deltaW[l][i] = make([]float64, len(dfa.Weights[l][i]))

			// Compute error signal via random feedback
			errorSignal := 0.0
			for k := range outputError {
				errorSignal += dfa.FeedbackWeights[l][i][k] * outputError[k]
			}

			// Derivative of tanh
			deriv := 1.0 - activations[l+1][i]*activations[l+1][i]
			localGradient := errorSignal * deriv

			// Weight update
			for j := range dfa.Weights[l][i] {
				deltaW[l][i][j] = dfa.Config.LearningRate * localGradient * activations[l][j]
			}
		}
	}

	return deltaW
}

// ApplyUpdate applies DFA weight updates
func (dfa *DirectFeedbackAlignment) ApplyUpdate(deltaW [][][]float64) {
	for l := range dfa.Weights {
		for i := range dfa.Weights[l] {
			for j := range dfa.Weights[l][i] {
				dfa.Weights[l][i][j] += deltaW[l][i][j]
				dfa.Weights[l][i][j] *= (1.0 - dfa.Config.WeightDecay)

				if dfa.Weights[l][i][j] < -2 {
					dfa.Weights[l][i][j] = -2
				}
				if dfa.Weights[l][i][j] > 2 {
					dfa.Weights[l][i][j] = 2
				}
			}
		}
	}
}

// =============================================================================
// Performance Metrics
// =============================================================================

// OnChipLearningMetrics holds learning performance data
type OnChipLearningMetrics struct {
	Algorithm        LearningAlgorithm
	EnergyPerUpdate  float64 // Joules
	UpdateLatencyUs  float64
	AccuracyVsBP     float64 // Percentage relative to backprop
	MemoryRequired   int64   // Bytes
	HardwareOverhead float64 // Fraction of compute area
}

// EstimateOnChipMetrics estimates on-chip learning performance
func EstimateOnChipMetrics(algorithm LearningAlgorithm, layerSizes []int) *OnChipLearningMetrics {
	metrics := &OnChipLearningMetrics{
		Algorithm: algorithm,
	}

	totalParams := int64(0)
	for i := 0; i < len(layerSizes)-1; i++ {
		totalParams += int64(layerSizes[i] * layerSizes[i+1])
	}

	switch algorithm {
	case LearningSTDP:
		metrics.EnergyPerUpdate = 10e-15 * float64(totalParams) // 10 fJ per synapse
		metrics.UpdateLatencyUs = 1.0                           // 1 µs for spike-based
		metrics.AccuracyVsBP = 85.0                             // Lower than BP
		metrics.MemoryRequired = totalParams * 2                // Weights only
		metrics.HardwareOverhead = 0.1                          // Simple circuits

	case LearningEP:
		metrics.EnergyPerUpdate = 50e-15 * float64(totalParams) // More iterations
		metrics.UpdateLatencyUs = 100.0                         // Multiple phases
		metrics.AccuracyVsBP = 95.0                             // Near BP accuracy
		metrics.MemoryRequired = totalParams*2 + int64(sum(layerSizes))*8
		metrics.HardwareOverhead = 0.3 // Analog circuits for phases

	case LearningDFA:
		metrics.EnergyPerUpdate = 30e-15 * float64(totalParams)
		metrics.UpdateLatencyUs = 10.0
		metrics.AccuracyVsBP = 92.0
		metrics.MemoryRequired = totalParams * 4 // Forward + feedback weights
		metrics.HardwareOverhead = 0.2

	default:
		metrics.EnergyPerUpdate = 100e-15 * float64(totalParams)
		metrics.UpdateLatencyUs = 50.0
		metrics.AccuracyVsBP = 90.0
		metrics.MemoryRequired = totalParams * 4
		metrics.HardwareOverhead = 0.25
	}

	return metrics
}

func sum(arr []int) int {
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
}

// String returns formatted metrics
func (m *OnChipLearningMetrics) String() string {
	return fmt.Sprintf(`On-Chip Learning Metrics (%s):
  Energy/Update:     %.2f pJ
  Latency:           %.1f µs
  Accuracy vs BP:    %.1f%%
  Memory Required:   %d KB
  Hardware Overhead: %.0f%%`,
		m.Algorithm,
		m.EnergyPerUpdate*1e12,
		m.UpdateLatencyUs,
		m.AccuracyVsBP,
		m.MemoryRequired/1024,
		m.HardwareOverhead*100)
}
