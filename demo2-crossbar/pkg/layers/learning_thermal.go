// learning_thermal.go - On-Chip Learning and Thermal Modeling for CIM
//
// This module implements:
// - STDP (Spike-Timing-Dependent Plasticity) learning rules
// - Equilibrium propagation for energy-efficient training
// - Reward-modulated STDP (R-STDP) for reinforcement learning
// - Thermal modeling with self-heating and crosstalk
// - Temperature-aware optimization strategies
//
// Based on research findings:
// - STDP achieves 97% accuracy with 15% noise tolerance
// - Equilibrium propagation maintains 92% with 10% device defects
// - Thermal optimization reduces hotspot temperature by up to 4.04°C
//
// References:
// - "On-chip Learning with Ferroelectric Synapses" (Nature Electronics 2024)
// - "Equilibrium Propagation for Analog CIM" (IEEE JSSC 2024)
// - "Thermal Management in Memristive Crossbars" (IEEE TED 2024)

package layers

import (
	"math"
	"math/rand"
	"sort"
)

// ================== STDP Learning ==================

// STDPConfig configures spike-timing-dependent plasticity
type STDPConfig struct {
	TauPlus       float64 // Pre-before-post time constant (ms)
	TauMinus      float64 // Post-before-pre time constant (ms)
	APlus         float64 // LTP amplitude
	AMinus        float64 // LTD amplitude
	WMin          float64 // Minimum weight
	WMax          float64 // Maximum weight
	LearningRate  float64 // Global learning rate
	WeightDepend  bool    // Weight-dependent STDP
	Symmetric     bool    // Symmetric STDP window
	TripletSTDP   bool    // Triplet STDP rule
	TauTriplet    float64 // Triplet time constant
	NoiseLevel    float64 // Synaptic noise (0-1)
	UpdateMode    string  // "additive", "multiplicative", "soft_bounds"
}

// DefaultSTDPConfig returns standard STDP parameters
func DefaultSTDPConfig() *STDPConfig {
	return &STDPConfig{
		TauPlus:      20.0,  // 20 ms
		TauMinus:     20.0,  // 20 ms
		APlus:        0.01,  // LTP amplitude
		AMinus:       0.012, // LTD slightly stronger (heterosynaptic)
		WMin:         0.0,
		WMax:         1.0,
		LearningRate: 0.1,
		WeightDepend: true,
		Symmetric:    false,
		TripletSTDP:  false,
		TauTriplet:   100.0,
		NoiseLevel:   0.02,
		UpdateMode:   "soft_bounds",
	}
}

// STDPSynapse represents a synapse with STDP learning
type STDPSynapse struct {
	Config       *STDPConfig
	Weight       float64
	PreTrace     float64   // Pre-synaptic eligibility trace
	PostTrace    float64   // Post-synaptic eligibility trace
	TripletTrace float64   // For triplet STDP
	LastPreTime  float64   // Last pre-synaptic spike time
	LastPostTime float64   // Last post-synaptic spike time
	DeltaW       float64   // Accumulated weight change
	UpdateCount  int64     // Number of weight updates
	Conductance  float64   // Physical conductance (for CIM)
	GMin         float64   // Minimum conductance
	GMax         float64   // Maximum conductance
	NonLinearity float64   // Conductance nonlinearity factor
}

// NewSTDPSynapse creates a synapse with STDP
func NewSTDPSynapse(config *STDPConfig, initialWeight float64) *STDPSynapse {
	return &STDPSynapse{
		Config:       config,
		Weight:       initialWeight,
		PreTrace:     0,
		PostTrace:    0,
		TripletTrace: 0,
		LastPreTime:  -1000,
		LastPostTime: -1000,
		DeltaW:       0,
		UpdateCount:  0,
		Conductance:  initialWeight, // Normalized conductance
		GMin:         1e-9,          // 1 nS
		GMax:         100e-6,        // 100 µS
		NonLinearity: 0.5,
	}
}

// PreSpike processes a pre-synaptic spike
func (s *STDPSynapse) PreSpike(time float64) float64 {
	cfg := s.Config

	// Decay traces
	dt := time - s.LastPreTime
	if dt > 0 {
		s.PreTrace *= math.Exp(-dt / cfg.TauPlus)
		s.PostTrace *= math.Exp(-dt / cfg.TauMinus)
		if cfg.TripletSTDP {
			s.TripletTrace *= math.Exp(-dt / cfg.TauTriplet)
		}
	}

	// LTD: pre after post
	if s.PostTrace > 0 {
		dw := s.computeLTD()
		s.DeltaW += dw
	}

	// Update pre-trace
	s.PreTrace += 1.0
	s.LastPreTime = time

	// Return weighted output (for CIM: current proportional to conductance)
	return s.Weight * s.Conductance
}

// PostSpike processes a post-synaptic spike
func (s *STDPSynapse) PostSpike(time float64) {
	cfg := s.Config

	// Decay traces
	dt := time - s.LastPostTime
	if dt > 0 {
		s.PreTrace *= math.Exp(-dt / cfg.TauPlus)
		s.PostTrace *= math.Exp(-dt / cfg.TauMinus)
		if cfg.TripletSTDP {
			s.TripletTrace *= math.Exp(-dt / cfg.TauTriplet)
		}
	}

	// LTP: post after pre
	if s.PreTrace > 0 {
		dw := s.computeLTP()
		s.DeltaW += dw
	}

	// Update post-trace
	s.PostTrace += 1.0
	if cfg.TripletSTDP {
		s.TripletTrace += 1.0
	}
	s.LastPostTime = time
}

// computeLTP calculates long-term potentiation
func (s *STDPSynapse) computeLTP() float64 {
	cfg := s.Config
	dw := cfg.APlus * s.PreTrace * cfg.LearningRate

	// Triplet modulation
	if cfg.TripletSTDP {
		dw *= (1.0 + s.TripletTrace)
	}

	// Weight dependence
	if cfg.WeightDepend {
		switch cfg.UpdateMode {
		case "multiplicative":
			dw *= (cfg.WMax - s.Weight)
		case "soft_bounds":
			dw *= math.Pow(cfg.WMax-s.Weight, 0.5)
		}
	}

	// Add noise
	if cfg.NoiseLevel > 0 {
		dw *= (1.0 + cfg.NoiseLevel*(rand.Float64()*2-1))
	}

	return dw
}

// computeLTD calculates long-term depression
func (s *STDPSynapse) computeLTD() float64 {
	cfg := s.Config
	dw := -cfg.AMinus * s.PostTrace * cfg.LearningRate

	// Weight dependence
	if cfg.WeightDepend {
		switch cfg.UpdateMode {
		case "multiplicative":
			dw *= s.Weight
		case "soft_bounds":
			dw *= math.Pow(s.Weight-cfg.WMin, 0.5)
		}
	}

	// Add noise
	if cfg.NoiseLevel > 0 {
		dw *= (1.0 + cfg.NoiseLevel*(rand.Float64()*2-1))
	}

	return dw
}

// ApplyUpdate applies accumulated weight changes
func (s *STDPSynapse) ApplyUpdate() {
	cfg := s.Config

	s.Weight += s.DeltaW
	s.Weight = math.Max(cfg.WMin, math.Min(cfg.WMax, s.Weight))

	// Update physical conductance (nonlinear mapping)
	s.Conductance = s.weightToConductance(s.Weight)

	s.DeltaW = 0
	s.UpdateCount++
}

// weightToConductance maps logical weight to physical conductance
func (s *STDPSynapse) weightToConductance(w float64) float64 {
	// Nonlinear conductance response (typical for memristors)
	normalizedW := (w - s.Config.WMin) / (s.Config.WMax - s.Config.WMin)
	// Apply nonlinearity
	nonlinearW := math.Pow(normalizedW, s.NonLinearity)
	return s.GMin + nonlinearW*(s.GMax-s.GMin)
}

// ================== STDP Layer ==================

// STDPLayer implements a layer with STDP learning
type STDPLayer struct {
	Config    *STDPConfig
	Synapses  [][]*STDPSynapse // [post][pre]
	NumInputs int
	NumOutputs int
	CurrentTime float64
	InputSpikes  []float64 // Last spike times
	OutputSpikes []float64
	Threshold    float64 // Output neuron threshold
	Potential    []float64 // Membrane potentials
	TauMembrane  float64
}

// NewSTDPLayer creates an STDP learning layer
func NewSTDPLayer(config *STDPConfig, numInputs, numOutputs int) *STDPLayer {
	synapses := make([][]*STDPSynapse, numOutputs)
	for i := range synapses {
		synapses[i] = make([]*STDPSynapse, numInputs)
		for j := range synapses[i] {
			// Initialize with random weights
			w := config.WMin + rand.Float64()*(config.WMax-config.WMin)*0.5
			synapses[i][j] = NewSTDPSynapse(config, w)
		}
	}

	return &STDPLayer{
		Config:       config,
		Synapses:     synapses,
		NumInputs:    numInputs,
		NumOutputs:   numOutputs,
		CurrentTime:  0,
		InputSpikes:  make([]float64, numInputs),
		OutputSpikes: make([]float64, numOutputs),
		Threshold:    1.0,
		Potential:    make([]float64, numOutputs),
		TauMembrane:  20.0,
	}
}

// ProcessSpike processes input spikes and returns output spikes
func (l *STDPLayer) ProcessSpike(inputIdx int, time float64) []bool {
	l.CurrentTime = time
	l.InputSpikes[inputIdx] = time

	// Process pre-synaptic spike through all connected synapses
	outputSpikes := make([]bool, l.NumOutputs)
	for postIdx := 0; postIdx < l.NumOutputs; postIdx++ {
		synapse := l.Synapses[postIdx][inputIdx]

		// Get weighted contribution
		current := synapse.PreSpike(time)

		// Update membrane potential (LIF neuron)
		dt := time - l.OutputSpikes[postIdx]
		if dt > 0 {
			l.Potential[postIdx] *= math.Exp(-dt / l.TauMembrane)
		}
		l.Potential[postIdx] += current

		// Check for output spike
		if l.Potential[postIdx] >= l.Threshold {
			outputSpikes[postIdx] = true
			l.OutputSpikes[postIdx] = time
			l.Potential[postIdx] = 0 // Reset

			// Trigger post-synaptic spike in all incoming synapses
			for preIdx := 0; preIdx < l.NumInputs; preIdx++ {
				l.Synapses[postIdx][preIdx].PostSpike(time)
			}
		}
	}

	return outputSpikes
}

// ApplyUpdates applies all accumulated weight changes
func (l *STDPLayer) ApplyUpdates() {
	for postIdx := 0; postIdx < l.NumOutputs; postIdx++ {
		for preIdx := 0; preIdx < l.NumInputs; preIdx++ {
			l.Synapses[postIdx][preIdx].ApplyUpdate()
		}
	}
}

// GetWeights returns the weight matrix
func (l *STDPLayer) GetWeights() [][]float64 {
	weights := make([][]float64, l.NumOutputs)
	for i := range weights {
		weights[i] = make([]float64, l.NumInputs)
		for j := range weights[i] {
			weights[i][j] = l.Synapses[i][j].Weight
		}
	}
	return weights
}

// ================== Reward-Modulated STDP ==================

// RSTDPConfig configures reward-modulated STDP
type RSTDPConfig struct {
	*STDPConfig
	TauEligibility float64 // Eligibility trace time constant
	TauReward      float64 // Reward signal time constant
	BaselineReward float64 // Baseline reward for variance reduction
	RewardScale    float64 // Reward scaling factor
	Discount       float64 // Temporal discount factor
	ExplorationRate float64 // ε-greedy exploration
}

// DefaultRSTDPConfig returns standard R-STDP parameters
func DefaultRSTDPConfig() *RSTDPConfig {
	return &RSTDPConfig{
		STDPConfig:      DefaultSTDPConfig(),
		TauEligibility:  200.0, // 200 ms
		TauReward:       50.0,  // 50 ms
		BaselineReward:  0.5,
		RewardScale:     1.0,
		Discount:        0.99,
		ExplorationRate: 0.1,
	}
}

// RSTDPSynapse extends STDP with reward modulation
type RSTDPSynapse struct {
	*STDPSynapse
	RConfig        *RSTDPConfig
	EligibilityTrace float64
	AccumulatedReward float64
	RewardHistory    []float64
	BaselineEstimate float64
}

// NewRSTDPSynapse creates a reward-modulated STDP synapse
func NewRSTDPSynapse(config *RSTDPConfig, initialWeight float64) *RSTDPSynapse {
	return &RSTDPSynapse{
		STDPSynapse:      NewSTDPSynapse(config.STDPConfig, initialWeight),
		RConfig:          config,
		EligibilityTrace: 0,
		AccumulatedReward: 0,
		RewardHistory:    make([]float64, 0, 100),
		BaselineEstimate: config.BaselineReward,
	}
}

// UpdateEligibility updates the eligibility trace based on STDP
func (s *RSTDPSynapse) UpdateEligibility(time float64) {
	// Decay eligibility trace
	dt := 1.0 // Assuming 1 ms time step
	s.EligibilityTrace *= math.Exp(-dt / s.RConfig.TauEligibility)

	// Add STDP contribution to eligibility
	s.EligibilityTrace += s.DeltaW
	s.DeltaW = 0 // Reset STDP accumulator
}

// ApplyReward applies reward signal to modulate learning
func (s *RSTDPSynapse) ApplyReward(reward float64) {
	// Update baseline estimate (moving average)
	s.RewardHistory = append(s.RewardHistory, reward)
	if len(s.RewardHistory) > 100 {
		s.RewardHistory = s.RewardHistory[1:]
	}
	sum := 0.0
	for _, r := range s.RewardHistory {
		sum += r
	}
	s.BaselineEstimate = sum / float64(len(s.RewardHistory))

	// Compute reward prediction error
	rpe := (reward - s.BaselineEstimate) * s.RConfig.RewardScale

	// Modulate weight change by reward and eligibility
	dw := rpe * s.EligibilityTrace * s.RConfig.LearningRate
	s.Weight += dw
	s.Weight = math.Max(s.Config.WMin, math.Min(s.Config.WMax, s.Weight))

	// Update conductance
	s.Conductance = s.weightToConductance(s.Weight)

	// Decay eligibility after reward
	s.EligibilityTrace *= s.RConfig.Discount
}

// ================== Equilibrium Propagation ==================

// EPConfig configures equilibrium propagation
type EPConfig struct {
	Beta          float64 // Nudging strength
	Alpha         float64 // Learning rate
	NumIterations int     // Free phase iterations
	NumClampedIter int    // Clamped phase iterations
	Dt            float64 // Integration time step
	Tau           float64 // Energy time constant
	Epsilon       float64 // Convergence threshold
	Symmetric     bool    // Use symmetric weights
	Noise         float64 // Noise level for annealing
}

// DefaultEPConfig returns standard EP parameters
func DefaultEPConfig() *EPConfig {
	return &EPConfig{
		Beta:           0.5,
		Alpha:          0.1,
		NumIterations:  50,
		NumClampedIter: 10,
		Dt:             0.5,
		Tau:            10.0,
		Epsilon:        1e-4,
		Symmetric:      true,
		Noise:          0.01,
	}
}

// EPLayer implements equilibrium propagation
type EPLayer struct {
	Config  *EPConfig
	Weights [][]float64
	Biases  []float64
	State   []float64 // Neuron activations
	Nudge   []float64 // Nudging values (targets)
	Energy  float64
	NumInputs  int
	NumOutputs int
	GradW      [][]float64 // Weight gradients
	GradB      []float64   // Bias gradients
}

// NewEPLayer creates an equilibrium propagation layer
func NewEPLayer(config *EPConfig, numInputs, numOutputs int) *EPLayer {
	weights := make([][]float64, numOutputs)
	gradW := make([][]float64, numOutputs)
	for i := range weights {
		weights[i] = make([]float64, numInputs)
		gradW[i] = make([]float64, numInputs)
		for j := range weights[i] {
			// Xavier initialization
			weights[i][j] = rand.NormFloat64() * math.Sqrt(2.0/float64(numInputs+numOutputs))
		}
	}

	return &EPLayer{
		Config:     config,
		Weights:    weights,
		Biases:     make([]float64, numOutputs),
		State:      make([]float64, numOutputs),
		Nudge:      make([]float64, numOutputs),
		Energy:     0,
		NumInputs:  numInputs,
		NumOutputs: numOutputs,
		GradW:      gradW,
		GradB:      make([]float64, numOutputs),
	}
}

// Activation function (hardtanh for hardware friendliness)
func (l *EPLayer) activation(x float64) float64 {
	return math.Max(-1, math.Min(1, x))
}

// Activation derivative
func (l *EPLayer) activationDeriv(x float64) float64 {
	if x > -1 && x < 1 {
		return 1.0
	}
	return 0.0
}

// FreePhase runs the free phase (no target)
func (l *EPLayer) FreePhase(input []float64) []float64 {
	cfg := l.Config

	// Initialize state
	for i := range l.State {
		l.State[i] = 0
	}

	// Iterate to equilibrium
	for iter := 0; iter < cfg.NumIterations; iter++ {
		prevEnergy := l.Energy
		l.Energy = 0

		newState := make([]float64, l.NumOutputs)
		for i := 0; i < l.NumOutputs; i++ {
			// Compute input to neuron
			sum := l.Biases[i]
			for j := 0; j < l.NumInputs; j++ {
				sum += l.Weights[i][j] * input[j]
			}

			// Energy gradient
			dE := -sum + l.State[i]

			// Update with noise for annealing
			noise := 0.0
			if cfg.Noise > 0 {
				noise = cfg.Noise * rand.NormFloat64()
			}
			newState[i] = l.State[i] - cfg.Dt/cfg.Tau*(dE+noise)
			newState[i] = l.activation(newState[i])

			// Accumulate energy
			l.Energy += 0.5 * l.State[i] * l.State[i]
			for j := 0; j < l.NumInputs; j++ {
				l.Energy -= l.Weights[i][j] * l.State[i] * input[j]
			}
		}

		// Update state
		copy(l.State, newState)

		// Check convergence
		if math.Abs(l.Energy-prevEnergy) < cfg.Epsilon {
			break
		}
	}

	return l.State
}

// ClampedPhase runs the clamped phase (with target nudging)
func (l *EPLayer) ClampedPhase(input, target []float64) []float64 {
	cfg := l.Config

	// Store nudging values
	copy(l.Nudge, target)

	// Iterate with nudging
	for iter := 0; iter < cfg.NumClampedIter; iter++ {
		newState := make([]float64, l.NumOutputs)
		for i := 0; i < l.NumOutputs; i++ {
			// Compute input to neuron
			sum := l.Biases[i]
			for j := 0; j < l.NumInputs; j++ {
				sum += l.Weights[i][j] * input[j]
			}

			// Energy gradient with nudging
			nudge := cfg.Beta * (target[i] - l.State[i])
			dE := -sum + l.State[i] - nudge

			// Update
			newState[i] = l.State[i] - cfg.Dt/cfg.Tau*dE
			newState[i] = l.activation(newState[i])
		}

		copy(l.State, newState)
	}

	return l.State
}

// ComputeGradients computes contrastive gradients
func (l *EPLayer) ComputeGradients(input []float64, freeState, clampedState []float64) {
	cfg := l.Config

	// Weight gradients: Δw = (1/β) * (s_clamped * x - s_free * x)
	for i := 0; i < l.NumOutputs; i++ {
		for j := 0; j < l.NumInputs; j++ {
			l.GradW[i][j] = (clampedState[i]*input[j] - freeState[i]*input[j]) / cfg.Beta
		}
		// Bias gradient
		l.GradB[i] = (clampedState[i] - freeState[i]) / cfg.Beta
	}
}

// ApplyGradients applies gradients to weights
func (l *EPLayer) ApplyGradients() {
	cfg := l.Config

	for i := 0; i < l.NumOutputs; i++ {
		for j := 0; j < l.NumInputs; j++ {
			l.Weights[i][j] += cfg.Alpha * l.GradW[i][j]
		}
		l.Biases[i] += cfg.Alpha * l.GradB[i]
	}
}

// Train performs one training step
func (l *EPLayer) Train(input, target []float64) float64 {
	// Free phase
	freeState := l.FreePhase(input)
	freeStateCopy := make([]float64, len(freeState))
	copy(freeStateCopy, freeState)

	// Clamped phase
	clampedState := l.ClampedPhase(input, target)

	// Compute and apply gradients
	l.ComputeGradients(input, freeStateCopy, clampedState)
	l.ApplyGradients()

	// Return loss (MSE)
	loss := 0.0
	for i := range target {
		diff := target[i] - freeStateCopy[i]
		loss += diff * diff
	}
	return loss / float64(len(target))
}

// ================== Thermal Modeling ==================

// ThermalConfig configures thermal simulation
type ThermalConfig struct {
	AmbientTemp      float64 // Ambient temperature (°C)
	ThermalResistance float64 // Thermal resistance (°C/W)
	ThermalCapacity   float64 // Thermal capacity (J/°C)
	CrosstalkCoeff    float64 // Thermal crosstalk coefficient
	MaxTemp           float64 // Maximum allowed temperature
	CoolingCoeff      float64 // Cooling coefficient (natural convection)
	DutyCycle         float64 // Operation duty cycle
	PulseWidth        float64 // Programming pulse width (ns)
	CellSpacing       float64 // Cell spacing (nm)
}

// DefaultThermalConfig returns standard thermal parameters
func DefaultThermalConfig() *ThermalConfig {
	return &ThermalConfig{
		AmbientTemp:       25.0,    // Room temperature
		ThermalResistance: 1000.0,  // 1000 °C/W for nanoscale device
		ThermalCapacity:   1e-12,   // pJ/°C
		CrosstalkCoeff:    0.15,    // 15% crosstalk to neighbors
		MaxTemp:           85.0,    // Industrial grade
		CoolingCoeff:      10.0,    // W/(m²·K)
		DutyCycle:         0.1,     // 10% duty cycle
		PulseWidth:        100.0,   // 100 ns pulse
		CellSpacing:       50.0,    // 50 nm spacing
	}
}

// ThermalCell represents thermal state of a crossbar cell
type ThermalCell struct {
	Config       *ThermalConfig
	Temperature  float64 // Current temperature (°C)
	PowerDissip  float64 // Power dissipation (W)
	Resistance   float64 // Cell resistance (Ω)
	Voltage      float64 // Applied voltage (V)
	Row          int
	Col          int
	SelfHeating  float64 // Self-heating contribution
	Crosstalk    float64 // Crosstalk from neighbors
	CoolingRate  float64 // Cooling rate
}

// NewThermalCell creates a thermal cell model
func NewThermalCell(config *ThermalConfig, row, col int, resistance float64) *ThermalCell {
	return &ThermalCell{
		Config:      config,
		Temperature: config.AmbientTemp,
		PowerDissip: 0,
		Resistance:  resistance,
		Voltage:     0,
		Row:         row,
		Col:         col,
		SelfHeating: 0,
		Crosstalk:   0,
		CoolingRate: 0,
	}
}

// UpdatePower updates power dissipation based on voltage
func (c *ThermalCell) UpdatePower(voltage float64) {
	c.Voltage = voltage
	// P = V²/R
	c.PowerDissip = (voltage * voltage) / c.Resistance
}

// ComputeSelfHeating calculates self-heating temperature rise
func (c *ThermalCell) ComputeSelfHeating(dt float64) float64 {
	cfg := c.Config

	// Temperature rise from power dissipation
	// ΔT = P * Rth * (1 - exp(-t/τth))
	// where τth = Rth * Cth
	tau := cfg.ThermalResistance * cfg.ThermalCapacity
	if tau > 0 {
		c.SelfHeating = c.PowerDissip * cfg.ThermalResistance * (1 - math.Exp(-dt/tau))
	}

	// Apply duty cycle
	c.SelfHeating *= cfg.DutyCycle

	return c.SelfHeating
}

// ThermalCrossbar models thermal effects in a crossbar array
type ThermalCrossbar struct {
	Config      *ThermalConfig
	Cells       [][]*ThermalCell
	Rows        int
	Cols        int
	HotspotRow  int
	HotspotCol  int
	MaxTemp     float64
	AvgTemp     float64
	ThermalMap  [][]float64 // Temperature distribution
}

// NewThermalCrossbar creates a thermal crossbar model
func NewThermalCrossbar(config *ThermalConfig, rows, cols int, resistances [][]float64) *ThermalCrossbar {
	cells := make([][]*ThermalCell, rows)
	thermalMap := make([][]float64, rows)
	for i := range cells {
		cells[i] = make([]*ThermalCell, cols)
		thermalMap[i] = make([]float64, cols)
		for j := range cells[i] {
			r := 10000.0 // Default 10 kΩ
			if resistances != nil && i < len(resistances) && j < len(resistances[i]) {
				r = resistances[i][j]
			}
			cells[i][j] = NewThermalCell(config, i, j, r)
			thermalMap[i][j] = config.AmbientTemp
		}
	}

	return &ThermalCrossbar{
		Config:     config,
		Cells:      cells,
		Rows:       rows,
		Cols:       cols,
		HotspotRow: -1,
		HotspotCol: -1,
		MaxTemp:    config.AmbientTemp,
		AvgTemp:    config.AmbientTemp,
		ThermalMap: thermalMap,
	}
}

// ApplyVoltages applies voltage pattern and updates temperatures
func (tc *ThermalCrossbar) ApplyVoltages(rowVoltages, colVoltages []float64, dt float64) {
	// Update power dissipation for each cell
	for i := 0; i < tc.Rows; i++ {
		for j := 0; j < tc.Cols; j++ {
			// Voltage across cell = Vrow - Vcol
			cellVoltage := rowVoltages[i] - colVoltages[j]
			tc.Cells[i][j].UpdatePower(cellVoltage)
		}
	}

	// Compute self-heating
	for i := 0; i < tc.Rows; i++ {
		for j := 0; j < tc.Cols; j++ {
			tc.Cells[i][j].ComputeSelfHeating(dt)
		}
	}

	// Compute thermal crosstalk
	tc.computeCrosstalk()

	// Update temperatures
	tc.updateTemperatures(dt)

	// Find hotspot
	tc.findHotspot()
}

// computeCrosstalk calculates thermal crosstalk between neighboring cells
func (tc *ThermalCrossbar) computeCrosstalk() {
	cfg := tc.Config

	for i := 0; i < tc.Rows; i++ {
		for j := 0; j < tc.Cols; j++ {
			crosstalk := 0.0
			count := 0

			// Sum contributions from 4-connected neighbors
			neighbors := []struct{ di, dj int }{
				{-1, 0}, {1, 0}, {0, -1}, {0, 1},
			}

			for _, n := range neighbors {
				ni, nj := i+n.di, j+n.dj
				if ni >= 0 && ni < tc.Rows && nj >= 0 && nj < tc.Cols {
					// Distance-weighted crosstalk
					neighborTemp := tc.Cells[ni][nj].Temperature - cfg.AmbientTemp
					crosstalk += cfg.CrosstalkCoeff * neighborTemp
					count++
				}
			}

			if count > 0 {
				tc.Cells[i][j].Crosstalk = crosstalk / float64(count)
			}
		}
	}
}

// updateTemperatures updates all cell temperatures
func (tc *ThermalCrossbar) updateTemperatures(dt float64) {
	cfg := tc.Config

	sumTemp := 0.0
	for i := 0; i < tc.Rows; i++ {
		for j := 0; j < tc.Cols; j++ {
			cell := tc.Cells[i][j]

			// Compute cooling (Newton's law)
			tempDiff := cell.Temperature - cfg.AmbientTemp
			cooling := cfg.CoolingCoeff * tempDiff * dt

			// Update temperature
			newTemp := cell.Temperature + cell.SelfHeating + cell.Crosstalk - cooling

			// Clamp to physical limits
			newTemp = math.Max(cfg.AmbientTemp, math.Min(200, newTemp)) // Max 200°C

			cell.Temperature = newTemp
			tc.ThermalMap[i][j] = newTemp
			sumTemp += newTemp
		}
	}

	tc.AvgTemp = sumTemp / float64(tc.Rows*tc.Cols)
}

// findHotspot locates the hottest cell
func (tc *ThermalCrossbar) findHotspot() {
	tc.MaxTemp = tc.Config.AmbientTemp
	for i := 0; i < tc.Rows; i++ {
		for j := 0; j < tc.Cols; j++ {
			if tc.Cells[i][j].Temperature > tc.MaxTemp {
				tc.MaxTemp = tc.Cells[i][j].Temperature
				tc.HotspotRow = i
				tc.HotspotCol = j
			}
		}
	}
}

// GetThermalGradient returns temperature gradient at a cell
func (tc *ThermalCrossbar) GetThermalGradient(row, col int) (float64, float64) {
	dTdx, dTdy := 0.0, 0.0

	// X gradient (column direction)
	if col > 0 && col < tc.Cols-1 {
		dTdx = (tc.ThermalMap[row][col+1] - tc.ThermalMap[row][col-1]) / 2
	}

	// Y gradient (row direction)
	if row > 0 && row < tc.Rows-1 {
		dTdy = (tc.ThermalMap[row+1][col] - tc.ThermalMap[row-1][col]) / 2
	}

	return dTdx, dTdy
}

// ================== Thermal-Aware Optimization ==================

// ThermalOptimizer optimizes crossbar operations for thermal constraints
type ThermalOptimizer struct {
	Config         *ThermalConfig
	Crossbar       *ThermalCrossbar
	OperationQueue []*ThermalOperation
	CoolingPeriod  float64
	MaxBatchSize   int
	ThermalBudget  float64
}

// ThermalOperation represents a pending crossbar operation
type ThermalOperation struct {
	Type       string // "read", "write", "mvm"
	RowIdx     int
	ColIdx     int
	Voltage    float64
	Duration   float64
	Priority   int
	ThermalCost float64
}

// NewThermalOptimizer creates a thermal-aware optimizer
func NewThermalOptimizer(crossbar *ThermalCrossbar) *ThermalOptimizer {
	return &ThermalOptimizer{
		Config:         crossbar.Config,
		Crossbar:       crossbar,
		OperationQueue: make([]*ThermalOperation, 0),
		CoolingPeriod:  1e-6, // 1 µs cooling between operations
		MaxBatchSize:   16,
		ThermalBudget:  crossbar.Config.MaxTemp - crossbar.Config.AmbientTemp,
	}
}

// AddOperation adds an operation to the queue
func (to *ThermalOptimizer) AddOperation(op *ThermalOperation) {
	// Estimate thermal cost
	power := (op.Voltage * op.Voltage) / to.Crossbar.Cells[op.RowIdx][op.ColIdx].Resistance
	op.ThermalCost = power * op.Duration * to.Config.ThermalResistance

	to.OperationQueue = append(to.OperationQueue, op)
}

// OptimizeSchedule reorders operations to minimize peak temperature
func (to *ThermalOptimizer) OptimizeSchedule() []*ThermalOperation {
	if len(to.OperationQueue) == 0 {
		return nil
	}

	// Group operations by spatial locality
	groups := make(map[string][]*ThermalOperation)
	for _, op := range to.OperationQueue {
		// 4x4 spatial grouping
		groupKey := string(rune(op.RowIdx/4)) + string(rune(op.ColIdx/4))
		groups[groupKey] = append(groups[groupKey], op)
	}

	// Sort groups by total thermal cost (ascending)
	type groupInfo struct {
		key   string
		ops   []*ThermalOperation
		cost  float64
	}
	sortedGroups := make([]groupInfo, 0, len(groups))
	for key, ops := range groups {
		totalCost := 0.0
		for _, op := range ops {
			totalCost += op.ThermalCost
		}
		sortedGroups = append(sortedGroups, groupInfo{key, ops, totalCost})
	}
	sort.Slice(sortedGroups, func(i, j int) bool {
		return sortedGroups[i].cost < sortedGroups[j].cost
	})

	// Interleave operations from different spatial regions
	scheduled := make([]*ThermalOperation, 0, len(to.OperationQueue))
	for len(sortedGroups) > 0 {
		// Pick from alternating groups to spread heat
		for i := 0; i < len(sortedGroups); i++ {
			if len(sortedGroups[i].ops) > 0 {
				scheduled = append(scheduled, sortedGroups[i].ops[0])
				sortedGroups[i].ops = sortedGroups[i].ops[1:]
			}
		}

		// Remove empty groups
		newGroups := make([]groupInfo, 0)
		for _, g := range sortedGroups {
			if len(g.ops) > 0 {
				newGroups = append(newGroups, g)
			}
		}
		sortedGroups = newGroups
	}

	return scheduled
}

// ExecuteWithCooling executes operations with cooling periods
func (to *ThermalOptimizer) ExecuteWithCooling(operations []*ThermalOperation) *ThermalExecutionResult {
	result := &ThermalExecutionResult{
		Operations:    len(operations),
		TotalTime:    0,
		PeakTemp:     to.Config.AmbientTemp,
		AvgTemp:      to.Config.AmbientTemp,
		CoolingTime:  0,
		ThermalViolations: 0,
	}

	for _, op := range operations {
		// Check if cooling needed
		currentTemp := to.Crossbar.Cells[op.RowIdx][op.ColIdx].Temperature
		if currentTemp + op.ThermalCost > to.Config.MaxTemp {
			// Insert cooling period
			coolTime := to.estimateCoolingTime(currentTemp, to.Config.MaxTemp-op.ThermalCost)
			result.CoolingTime += coolTime
			result.TotalTime += coolTime

			// Apply cooling
			to.applyCooling(coolTime)
		}

		// Execute operation
		rowV := make([]float64, to.Crossbar.Rows)
		colV := make([]float64, to.Crossbar.Cols)
		rowV[op.RowIdx] = op.Voltage
		to.Crossbar.ApplyVoltages(rowV, colV, op.Duration)

		result.TotalTime += op.Duration

		// Track peak temperature
		if to.Crossbar.MaxTemp > result.PeakTemp {
			result.PeakTemp = to.Crossbar.MaxTemp
		}

		// Check for thermal violations
		if to.Crossbar.MaxTemp > to.Config.MaxTemp {
			result.ThermalViolations++
		}
	}

	result.AvgTemp = to.Crossbar.AvgTemp
	return result
}

// estimateCoolingTime estimates time needed to cool from current to target temp
func (to *ThermalOptimizer) estimateCoolingTime(currentTemp, targetTemp float64) float64 {
	if currentTemp <= targetTemp {
		return 0
	}

	// Newton's cooling: T(t) = Ta + (T0 - Ta) * exp(-t/τ)
	// Solve for t: t = -τ * ln((Tt - Ta) / (T0 - Ta))
	tau := to.Config.ThermalResistance * to.Config.ThermalCapacity
	Ta := to.Config.AmbientTemp
	ratio := (targetTemp - Ta) / (currentTemp - Ta)
	if ratio <= 0 {
		return 10 * tau // Max cooling time
	}
	return -tau * math.Log(ratio)
}

// applyCooling simulates cooling period
func (to *ThermalOptimizer) applyCooling(duration float64) {
	// Zero voltage during cooling
	rowV := make([]float64, to.Crossbar.Rows)
	colV := make([]float64, to.Crossbar.Cols)
	to.Crossbar.ApplyVoltages(rowV, colV, duration)
}

// ThermalExecutionResult contains execution statistics
type ThermalExecutionResult struct {
	Operations        int
	TotalTime         float64
	PeakTemp          float64
	AvgTemp           float64
	CoolingTime       float64
	ThermalViolations int
}

// ================== Integrated Learning-Thermal System ==================

// LearningThermalConfig combines learning and thermal parameters
type LearningThermalConfig struct {
	STDP    *STDPConfig
	RSTDP   *RSTDPConfig
	EP      *EPConfig
	Thermal *ThermalConfig
	ThermalAwareLearning bool    // Enable thermal-aware learning rate
	TempLRScale          float64 // Temperature-dependent LR scaling
	MaxTempForLearning   float64 // Disable learning above this temp
}

// DefaultLearningThermalConfig returns standard combined config
func DefaultLearningThermalConfig() *LearningThermalConfig {
	return &LearningThermalConfig{
		STDP:                 DefaultSTDPConfig(),
		RSTDP:                DefaultRSTDPConfig(),
		EP:                   DefaultEPConfig(),
		Thermal:              DefaultThermalConfig(),
		ThermalAwareLearning: true,
		TempLRScale:          0.02, // 2% LR reduction per °C above ambient
		MaxTempForLearning:   70.0, // Stop learning above 70°C
	}
}

// ThermalAwareSTDPLayer combines STDP learning with thermal constraints
type ThermalAwareSTDPLayer struct {
	STDPLayer  *STDPLayer
	Thermal    *ThermalCrossbar
	Config     *LearningThermalConfig
	Optimizer  *ThermalOptimizer
	Statistics *LearningThermalStats
}

// LearningThermalStats tracks combined learning/thermal statistics
type LearningThermalStats struct {
	TotalSpikes      int64
	WeightUpdates    int64
	ThermalPauses    int64
	AvgTemperature   float64
	MaxTemperature   float64
	EffectiveLR      float64
	LearningDisabled int64
}

// NewThermalAwareSTDPLayer creates a thermal-aware STDP layer
func NewThermalAwareSTDPLayer(config *LearningThermalConfig, numInputs, numOutputs int) *ThermalAwareSTDPLayer {
	// Create STDP layer
	stdpLayer := NewSTDPLayer(config.STDP, numInputs, numOutputs)

	// Extract resistances from STDP weights (inverse relationship)
	resistances := make([][]float64, numOutputs)
	for i := range resistances {
		resistances[i] = make([]float64, numInputs)
		for j := range resistances[i] {
			// R = R0 / G (conductance proportional to weight)
			g := stdpLayer.Synapses[i][j].Conductance
			if g > 0 {
				resistances[i][j] = 1.0 / g
			} else {
				resistances[i][j] = 1e9 // Very high resistance
			}
		}
	}

	// Create thermal crossbar
	thermal := NewThermalCrossbar(config.Thermal, numOutputs, numInputs, resistances)

	return &ThermalAwareSTDPLayer{
		STDPLayer:  stdpLayer,
		Thermal:    thermal,
		Config:     config,
		Optimizer:  NewThermalOptimizer(thermal),
		Statistics: &LearningThermalStats{},
	}
}

// ProcessSpikeWithThermal processes spike considering thermal effects
func (l *ThermalAwareSTDPLayer) ProcessSpikeWithThermal(inputIdx int, time float64) []bool {
	cfg := l.Config
	l.Statistics.TotalSpikes++

	// Check thermal state
	avgTemp := l.Thermal.AvgTemp
	l.Statistics.AvgTemperature = avgTemp
	if l.Thermal.MaxTemp > l.Statistics.MaxTemperature {
		l.Statistics.MaxTemperature = l.Thermal.MaxTemp
	}

	// Thermal-aware learning rate adjustment
	effectiveLR := cfg.STDP.LearningRate
	if cfg.ThermalAwareLearning {
		tempAboveAmbient := avgTemp - cfg.Thermal.AmbientTemp
		lrScale := 1.0 - cfg.TempLRScale*tempAboveAmbient
		effectiveLR *= math.Max(0.1, lrScale) // Minimum 10% of base LR
		l.Statistics.EffectiveLR = effectiveLR

		// Check if learning should be disabled
		if avgTemp > cfg.MaxTempForLearning {
			effectiveLR = 0
			l.Statistics.LearningDisabled++
		}
	}

	// Temporarily adjust learning rate
	originalLR := l.STDPLayer.Config.LearningRate
	l.STDPLayer.Config.LearningRate = effectiveLR

	// Process spike
	outputSpikes := l.STDPLayer.ProcessSpike(inputIdx, time)

	// Restore learning rate
	l.STDPLayer.Config.LearningRate = originalLR

	// Update thermal model with spike activity
	l.updateThermalFromSpike(inputIdx, outputSpikes)

	return outputSpikes
}

// updateThermalFromSpike updates thermal state based on spike activity
func (l *ThermalAwareSTDPLayer) updateThermalFromSpike(inputIdx int, outputSpikes []bool) {
	// Create voltage pattern from spike activity
	rowVoltages := make([]float64, l.Thermal.Rows)
	colVoltages := make([]float64, l.Thermal.Cols)

	// Input spike generates column voltage
	colVoltages[inputIdx] = 0.1 // 100 mV read voltage

	// Output spikes indicate row activity
	for i, spiked := range outputSpikes {
		if spiked {
			rowVoltages[i] = 0.5 // 500 mV for post-spike
		}
	}

	// Update thermal model (1 ms time step)
	l.Thermal.ApplyVoltages(rowVoltages, colVoltages, 1e-3)
}

// ApplyUpdatesWithThermal applies weight updates with thermal scheduling
func (l *ThermalAwareSTDPLayer) ApplyUpdatesWithThermal() *ThermalExecutionResult {
	// Collect weight updates as operations
	l.Optimizer.OperationQueue = make([]*ThermalOperation, 0)

	for i := 0; i < l.STDPLayer.NumOutputs; i++ {
		for j := 0; j < l.STDPLayer.NumInputs; j++ {
			synapse := l.STDPLayer.Synapses[i][j]
			if math.Abs(synapse.DeltaW) > 1e-6 {
				// Create write operation
				writeV := 1.0 // 1V write pulse
				if synapse.DeltaW < 0 {
					writeV = -1.0 // Negative pulse for depression
				}
				op := &ThermalOperation{
					Type:     "write",
					RowIdx:   i,
					ColIdx:   j,
					Voltage:  writeV,
					Duration: l.Config.Thermal.PulseWidth * 1e-9, // Convert ns to s
					Priority: int(math.Abs(synapse.DeltaW) * 1000),
				}
				l.Optimizer.AddOperation(op)
			}
		}
	}

	// Optimize schedule
	scheduled := l.Optimizer.OptimizeSchedule()

	// Execute with cooling
	result := l.Optimizer.ExecuteWithCooling(scheduled)

	// Apply weight updates
	l.STDPLayer.ApplyUpdates()
	l.Statistics.WeightUpdates += int64(result.Operations)

	// Update resistances from new weights
	l.syncResistancesFromWeights()

	return result
}

// syncResistancesFromWeights updates thermal model resistances from weights
func (l *ThermalAwareSTDPLayer) syncResistancesFromWeights() {
	for i := 0; i < l.STDPLayer.NumOutputs; i++ {
		for j := 0; j < l.STDPLayer.NumInputs; j++ {
			g := l.STDPLayer.Synapses[i][j].Conductance
			if g > 0 {
				l.Thermal.Cells[i][j].Resistance = 1.0 / g
			}
		}
	}
}

// GetThermalMap returns the current temperature distribution
func (l *ThermalAwareSTDPLayer) GetThermalMap() [][]float64 {
	return l.Thermal.ThermalMap
}

// GetStatistics returns combined learning/thermal statistics
func (l *ThermalAwareSTDPLayer) GetStatistics() *LearningThermalStats {
	return l.Statistics
}

// ResetStatistics resets all statistics
func (l *ThermalAwareSTDPLayer) ResetStatistics() {
	l.Statistics = &LearningThermalStats{}
}
