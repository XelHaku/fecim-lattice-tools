// emerging_neuromorphic_cim.go - Emerging Memory Technologies and Neuromorphic Integration for CIM
// Iteration 146: FTJ memristors, NDR neurons, and integrated neuromorphic systems
//
// Emerging Memory Technologies:
// - Ferroelectric Tunnel Junction (FTJ) memristors
// - HfO2/ZrO2-based FTJ with high TER ratio
// - Multi-level conductance programming
// - NDR (Negative Differential Resistance) memristors
// - Fitzhugh-Nagumo neuron circuits
//
// Neuromorphic Integration:
// - Leaky Integrate-and-Fire (LIF) neuron models
// - STDP (Spike-Timing Dependent Plasticity) learning
// - Memristive crossbar arrays for SNNs
// - On-chip learning systems
// - Event-driven computation
//
// References:
// - HZO-Based FTJ for Speech Recognition (ACS AMI 2025)
// - Ultra Robust NDR Memristor (Nature Communications 2025)
// - TEXEL Neuromorphic Processor (Nature Communications 2025)
// - Review of Memristors for IMC and SNNs (Adv. Intelligent Systems 2025)

package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// Ferroelectric Tunnel Junction (FTJ) Memristor
// ============================================================================

// FTJConfig holds FTJ device parameters
type FTJConfig struct {
	// Material parameters (HfO2/ZrO2 based)
	ThicknessNm       float64 // Ferroelectric layer thickness
	HfZrRatio         float64 // Hf:Zr ratio (0.5 for Hf0.5Zr0.5O2)
	InterlayerType    string  // "TiO2", "Al2O3", "none"
	InterlayerNm      float64 // Interlayer thickness

	// Electrical parameters
	TERRatio          float64 // Tunneling electroresistance ratio
	OnOffRatio        float64 // ON/OFF current ratio
	ConductanceLevels int     // Number of programmable states
	ReadVoltage       float64 // Read voltage (V)
	WriteVoltage      float64 // Write pulse voltage (V)
	WritePulseNs      float64 // Write pulse width (ns)

	// Reliability parameters
	Endurance         float64 // Number of cycles
	RetentionS        float64 // Retention time (seconds)
	CycleVariation    float64 // Cycle-to-cycle variation (%)
}

// FTJMemristor simulates a ferroelectric tunnel junction device
type FTJMemristor struct {
	Config            *FTJConfig
	Polarization      float64   // Current polarization state (-1 to +1)
	Conductance       float64   // Current conductance (S)
	ConductanceRange  []float64 // [Gmin, Gmax]
	ProgrammedStates  []float64 // Available conductance levels
	CurrentState      int       // Current state index
	CycleCount        int       // Total program/erase cycles
	Statistics        *FTJStats
}

// FTJStats tracks FTJ device statistics
type FTJStats struct {
	TotalReads        int
	TotalWrites       int
	AverageReadCurrent float64
	StateProgrammingError float64
	EnduranceRemaining float64
}

// NewFTJMemristor creates a new FTJ device
func NewFTJMemristor(config *FTJConfig) *FTJMemristor {
	ftj := &FTJMemristor{
		Config:           config,
		Polarization:     0,
		Statistics:       &FTJStats{},
	}

	// Calculate conductance range based on TER ratio
	gMin := 1e-9 // Minimum conductance (1 nS)
	gMax := gMin * config.TERRatio
	ftj.ConductanceRange = []float64{gMin, gMax}
	ftj.Conductance = gMin

	// Generate programmable conductance states
	ftj.ProgrammedStates = make([]float64, config.ConductanceLevels)
	for i := 0; i < config.ConductanceLevels; i++ {
		// Logarithmic spacing for better resolution at low conductance
		ratio := float64(i) / float64(config.ConductanceLevels-1)
		ftj.ProgrammedStates[i] = gMin * math.Pow(config.TERRatio, ratio)
	}

	ftj.Statistics.EnduranceRemaining = config.Endurance

	return ftj
}

// ProgramState programs FTJ to a specific state
func (ftj *FTJMemristor) ProgramState(targetState int) error {
	if targetState < 0 || targetState >= ftj.Config.ConductanceLevels {
		return fmt.Errorf("invalid state %d, must be 0-%d", targetState, ftj.Config.ConductanceLevels-1)
	}

	// Check endurance
	if ftj.Statistics.EnduranceRemaining <= 0 {
		return fmt.Errorf("device endurance exhausted")
	}

	// Apply programming with variation
	targetConductance := ftj.ProgrammedStates[targetState]
	variation := rand.NormFloat64() * ftj.Config.CycleVariation / 100
	ftj.Conductance = targetConductance * (1 + variation)
	ftj.CurrentState = targetState

	// Update polarization (linear mapping)
	ftj.Polarization = 2*float64(targetState)/float64(ftj.Config.ConductanceLevels-1) - 1

	// Update statistics
	ftj.CycleCount++
	ftj.Statistics.TotalWrites++
	ftj.Statistics.EnduranceRemaining--

	// Track programming error
	error := math.Abs(ftj.Conductance-targetConductance) / targetConductance
	ftj.Statistics.StateProgrammingError = (ftj.Statistics.StateProgrammingError*float64(ftj.Statistics.TotalWrites-1) +
		error) / float64(ftj.Statistics.TotalWrites)

	return nil
}

// Read reads the current conductance state
func (ftj *FTJMemristor) Read() float64 {
	ftj.Statistics.TotalReads++

	// Add read noise
	readNoise := rand.NormFloat64() * 0.01 * ftj.Conductance
	current := ftj.Conductance * ftj.Config.ReadVoltage

	ftj.Statistics.AverageReadCurrent = (ftj.Statistics.AverageReadCurrent*float64(ftj.Statistics.TotalReads-1) +
		current) / float64(ftj.Statistics.TotalReads)

	return ftj.Conductance + readNoise
}

// ============================================================================
// FTJ Crossbar Array for Neuromorphic Computing
// ============================================================================

// FTJCrossbarConfig holds crossbar array parameters
type FTJCrossbarConfig struct {
	Rows              int
	Cols              int
	FTJConfig         *FTJConfig
	LineResistance    float64 // Wire resistance (Ohms)
	ADCBits           int     // ADC resolution
	DACBits           int     // DAC resolution
}

// FTJCrossbar represents an FTJ-based crossbar array
type FTJCrossbar struct {
	Config      *FTJCrossbarConfig
	Devices     [][]*FTJMemristor
	Statistics  *FTJCrossbarStats
}

// FTJCrossbarStats tracks array statistics
type FTJCrossbarStats struct {
	TotalMVMs         int
	AverageEnergy     float64 // fJ per MAC
	ArrayYield        float64 // Percentage of working devices
	NonlinearityRatio float64
}

// NewFTJCrossbar creates a new FTJ crossbar array
func NewFTJCrossbar(config *FTJCrossbarConfig) *FTJCrossbar {
	cb := &FTJCrossbar{
		Config:     config,
		Devices:    make([][]*FTJMemristor, config.Rows),
		Statistics: &FTJCrossbarStats{},
	}

	// Initialize devices
	workingDevices := 0
	for i := 0; i < config.Rows; i++ {
		cb.Devices[i] = make([]*FTJMemristor, config.Cols)
		for j := 0; j < config.Cols; j++ {
			cb.Devices[i][j] = NewFTJMemristor(config.FTJConfig)
			if cb.Devices[i][j] != nil {
				workingDevices++
			}
		}
	}

	cb.Statistics.ArrayYield = float64(workingDevices) / float64(config.Rows*config.Cols) * 100

	return cb
}

// ProgramWeights programs weight matrix to crossbar
func (cb *FTJCrossbar) ProgramWeights(weights [][]float64) error {
	if len(weights) != cb.Config.Rows {
		return fmt.Errorf("weight rows %d doesn't match array rows %d", len(weights), cb.Config.Rows)
	}

	// Find weight range
	wMin, wMax := math.MaxFloat64, -math.MaxFloat64
	for _, row := range weights {
		for _, w := range row {
			wMin = math.Min(wMin, w)
			wMax = math.Max(wMax, w)
		}
	}

	if wMax <= wMin {
		wMax = wMin + 1
	}

	// Program each device
	levels := cb.Config.FTJConfig.ConductanceLevels
	for i := range weights {
		for j, w := range weights[i] {
			if j >= cb.Config.Cols {
				break
			}
			// Map weight to state
			normalized := (w - wMin) / (wMax - wMin)
			state := int(normalized * float64(levels-1))
			state = max(0, min(levels-1, state))

			cb.Devices[i][j].ProgramState(state)
		}
	}

	return nil
}

// MVM performs matrix-vector multiplication
func (cb *FTJCrossbar) MVM(input []float64) []float64 {
	output := make([]float64, cb.Config.Cols)

	// Quantize input with DAC
	dacLevels := float64(int(1) << cb.Config.DACBits)
	quantizedInput := make([]float64, len(input))
	for i, v := range input {
		quantizedInput[i] = math.Round(v*dacLevels) / dacLevels
	}

	// Perform analog MVM
	for j := 0; j < cb.Config.Cols; j++ {
		sum := 0.0
		for i := 0; i < cb.Config.Rows && i < len(quantizedInput); i++ {
			conductance := cb.Devices[i][j].Read()
			sum += quantizedInput[i] * conductance
		}

		// Add IR drop effect
		irDrop := sum * cb.Config.LineResistance * float64(cb.Config.Rows) / 1e9
		sum -= irDrop

		// Quantize with ADC
		adcLevels := float64(int(1) << cb.Config.ADCBits)
		output[j] = math.Round(sum*adcLevels) / adcLevels
	}

	cb.Statistics.TotalMVMs++

	// Estimate energy (simplified model: ~1fJ per MAC for FTJ)
	macs := float64(cb.Config.Rows * cb.Config.Cols)
	cb.Statistics.AverageEnergy = 1.0 * macs // fJ

	return output
}

// ============================================================================
// NDR (Negative Differential Resistance) Memristor for Neurons
// ============================================================================

// NDRConfig holds NDR device parameters
type NDRConfig struct {
	// Device structure (AlAs/InGaAs/AlAs quantum well)
	WellMaterial      string  // "InGaAs", "GaAs"
	InGaRatio         float64 // In:Ga ratio (e.g., 0.8 for In0.8Ga0.2As)
	BarrierMaterial   string  // "AlAs"

	// NDR characteristics
	PeakVoltage       float64 // Peak voltage (V)
	ValleyVoltage     float64 // Valley voltage (V)
	PeakCurrent       float64 // Peak current (A)
	ValleyCurrent     float64 // Valley current (A)
	PVCR              float64 // Peak-to-valley current ratio

	// Reliability
	Endurance         float64 // Switching cycles
	TempResistanceC   float64 // Max operating temperature (C)
	Variation         float64 // Device variation (%)
}

// NDRMemristor simulates an NDR device
type NDRMemristor struct {
	Config       *NDRConfig
	Voltage      float64 // Current voltage
	Current      float64 // Current current
	State        string  // "peak", "valley", "transition"
	CycleCount   int
	Statistics   *NDRStats
}

// NDRStats tracks NDR statistics
type NDRStats struct {
	PeakTransitions   int
	ValleyTransitions int
	TotalCycles       int
}

// NewNDRMemristor creates a new NDR device
func NewNDRMemristor(config *NDRConfig) *NDRMemristor {
	return &NDRMemristor{
		Config:     config,
		State:      "valley",
		Statistics: &NDRStats{},
	}
}

// ApplyVoltage applies voltage and returns resulting current
func (ndr *NDRMemristor) ApplyVoltage(voltage float64) float64 {
	ndr.Voltage = voltage

	// NDR I-V characteristic model
	vp := ndr.Config.PeakVoltage
	vv := ndr.Config.ValleyVoltage
	ip := ndr.Config.PeakCurrent
	iv := ndr.Config.ValleyCurrent

	var current float64

	if voltage < vp {
		// Pre-peak region (positive slope)
		current = ip * voltage / vp
		ndr.State = "rising"
	} else if voltage < vv {
		// NDR region (negative slope)
		ratio := (voltage - vp) / (vv - vp)
		current = ip - (ip-iv)*ratio
		ndr.State = "ndr"
	} else {
		// Post-valley region (positive slope)
		current = iv + (voltage-vv)*iv/vv
		ndr.State = "post_valley"
	}

	// Add variation
	variation := rand.NormFloat64() * ndr.Config.Variation / 100
	current *= (1 + variation)

	ndr.Current = current
	return current
}

// ============================================================================
// Fitzhugh-Nagumo Neuron with NDR Memristor
// ============================================================================

// FHNNeuronConfig holds Fitzhugh-Nagumo neuron parameters
type FHNNeuronConfig struct {
	NDRDevice         *NDRConfig
	MembraneCapF      float64 // Membrane capacitance (F)
	RecoveryTau       float64 // Recovery time constant
	ExcitabilityA     float64 // Excitability parameter a
	ExcitabilityB     float64 // Excitability parameter b
	Threshold         float64 // Firing threshold
}

// FHNNeuron implements a Fitzhugh-Nagumo neuron with NDR memristor
type FHNNeuron struct {
	Config         *FHNNeuronConfig
	NDR            *NDRMemristor
	V              float64 // Membrane potential
	W              float64 // Recovery variable
	Spiking        bool
	SpikeHistory   []float64 // Spike times
	Statistics     *FHNStats
}

// FHNStats tracks FHN neuron statistics
type FHNStats struct {
	TotalSpikes    int
	AverageFiringRate float64
	LastSpikeTime  float64
}

// NewFHNNeuron creates a new FHN neuron
func NewFHNNeuron(config *FHNNeuronConfig) *FHNNeuron {
	return &FHNNeuron{
		Config:       config,
		NDR:          NewNDRMemristor(config.NDRDevice),
		V:            -0.7, // Resting potential
		W:            0,
		SpikeHistory: make([]float64, 0),
		Statistics:   &FHNStats{},
	}
}

// Step advances the neuron by dt
func (fhn *FHNNeuron) Step(input float64, dt float64, currentTime float64) bool {
	// FHN equations with NDR memristor:
	// dV/dt = V - V^3/3 - W + I_input + I_NDR
	// dW/dt = (V + a - b*W) / tau

	// Get NDR current contribution
	iNDR := fhn.NDR.ApplyVoltage(fhn.V)

	// Voltage dynamics
	dV := fhn.V - math.Pow(fhn.V, 3)/3 - fhn.W + input + iNDR*1e6 // Scale NDR current
	fhn.V += dV * dt

	// Recovery dynamics
	dW := (fhn.V + fhn.Config.ExcitabilityA - fhn.Config.ExcitabilityB*fhn.W) / fhn.Config.RecoveryTau
	fhn.W += dW * dt

	// Check for spike
	spiked := false
	if fhn.V > fhn.Config.Threshold && !fhn.Spiking {
		spiked = true
		fhn.Spiking = true
		fhn.SpikeHistory = append(fhn.SpikeHistory, currentTime)
		fhn.Statistics.TotalSpikes++
		fhn.Statistics.LastSpikeTime = currentTime
	} else if fhn.V < fhn.Config.Threshold {
		fhn.Spiking = false
	}

	return spiked
}

// ============================================================================
// Leaky Integrate-and-Fire (LIF) Neuron
// ============================================================================

// LIFNeuronConfig holds LIF neuron parameters
type LIFNeuronConfig struct {
	MembraneCapPf     float64 // Membrane capacitance (pF)
	LeakConductanceNs float64 // Leak conductance (nS)
	RestingPotentialMv float64 // Resting potential (mV)
	ThresholdMv       float64 // Firing threshold (mV)
	ResetPotentialMv  float64 // Reset potential after spike (mV)
	RefractoryPeriodMs float64 // Refractory period (ms)
}

// LIFNeuron implements a leaky integrate-and-fire neuron
type LIFNeuron struct {
	Config           *LIFNeuronConfig
	MembranePotential float64 // Current membrane potential (mV)
	LastSpikeTime    float64  // Time of last spike (ms)
	InRefractory     bool     // Currently in refractory period
	SpikeCount       int
	TimeConstantMs   float64  // Membrane time constant (ms)
}

// NewLIFNeuron creates a new LIF neuron
func NewLIFNeuron(config *LIFNeuronConfig) *LIFNeuron {
	tau := config.MembraneCapPf / config.LeakConductanceNs // ms
	return &LIFNeuron{
		Config:            config,
		MembranePotential: config.RestingPotentialMv,
		LastSpikeTime:     -1000,
		TimeConstantMs:    tau,
	}
}

// Step advances the LIF neuron by dt
func (lif *LIFNeuron) Step(inputCurrent float64, currentTimeMs float64, dtMs float64) bool {
	// Check refractory period
	if currentTimeMs-lif.LastSpikeTime < lif.Config.RefractoryPeriodMs {
		lif.InRefractory = true
		return false
	}
	lif.InRefractory = false

	// LIF dynamics: tau * dV/dt = -(V - V_rest) + R * I
	// Simplified: dV/dt = (V_rest - V + I/g_leak) / tau

	resistance := 1.0 / lif.Config.LeakConductanceNs * 1000 // MOhm
	dV := (lif.Config.RestingPotentialMv - lif.MembranePotential + inputCurrent*resistance) / lif.TimeConstantMs

	lif.MembranePotential += dV * dtMs

	// Check for spike
	if lif.MembranePotential >= lif.Config.ThresholdMv {
		lif.MembranePotential = lif.Config.ResetPotentialMv
		lif.LastSpikeTime = currentTimeMs
		lif.SpikeCount++
		return true
	}

	return false
}

// Reset resets the neuron to initial state
func (lif *LIFNeuron) Reset() {
	lif.MembranePotential = lif.Config.RestingPotentialMv
	lif.LastSpikeTime = -1000
	lif.InRefractory = false
}

// ============================================================================
// STDP (Spike-Timing Dependent Plasticity) Learning
// ============================================================================

// STDPConfig holds STDP learning parameters
type STDPConfig struct {
	APlusMax      float64 // Maximum potentiation
	AMinusMax     float64 // Maximum depression
	TauPlusMs     float64 // Potentiation time constant (ms)
	TauMinusMs    float64 // Depression time constant (ms)
	WeightMin     float64 // Minimum synaptic weight
	WeightMax     float64 // Maximum synaptic weight
	LearningRate  float64 // Learning rate multiplier
}

// STDPSynapse implements an STDP-enabled synapse
type STDPSynapse struct {
	Config           *STDPConfig
	Weight           float64
	PreTrace         float64 // Presynaptic trace
	PostTrace        float64 // Postsynaptic trace
	PreSpikeHistory  []float64
	PostSpikeHistory []float64
	Statistics       *STDPStats
}

// STDPStats tracks STDP synapse statistics
type STDPStats struct {
	TotalPotentiations int
	TotalDepressions   int
	WeightHistory      []float64
}

// NewSTDPSynapse creates a new STDP synapse
func NewSTDPSynapse(config *STDPConfig, initialWeight float64) *STDPSynapse {
	return &STDPSynapse{
		Config:           config,
		Weight:           initialWeight,
		PreSpikeHistory:  make([]float64, 0),
		PostSpikeHistory: make([]float64, 0),
		Statistics:       &STDPStats{WeightHistory: make([]float64, 0)},
	}
}

// PreSpike processes a presynaptic spike
func (syn *STDPSynapse) PreSpike(timeMs float64) {
	syn.PreSpikeHistory = append(syn.PreSpikeHistory, timeMs)
	syn.PreTrace = 1.0 // Reset trace to 1

	// Check for recent post-spike (LTD)
	for _, postTime := range syn.PostSpikeHistory {
		dt := timeMs - postTime
		if dt > 0 && dt < 5*syn.Config.TauMinusMs {
			// LTD: pre after post
			dW := -syn.Config.AMinusMax * math.Exp(-dt/syn.Config.TauMinusMs)
			syn.updateWeight(dW)
			syn.Statistics.TotalDepressions++
		}
	}
}

// PostSpike processes a postsynaptic spike
func (syn *STDPSynapse) PostSpike(timeMs float64) {
	syn.PostSpikeHistory = append(syn.PostSpikeHistory, timeMs)
	syn.PostTrace = 1.0 // Reset trace to 1

	// Check for recent pre-spike (LTP)
	for _, preTime := range syn.PreSpikeHistory {
		dt := timeMs - preTime
		if dt > 0 && dt < 5*syn.Config.TauPlusMs {
			// LTP: post after pre
			dW := syn.Config.APlusMax * math.Exp(-dt/syn.Config.TauPlusMs)
			syn.updateWeight(dW)
			syn.Statistics.TotalPotentiations++
		}
	}
}

// updateWeight updates synaptic weight with bounds
func (syn *STDPSynapse) updateWeight(dW float64) {
	syn.Weight += dW * syn.Config.LearningRate
	syn.Weight = math.Max(syn.Config.WeightMin, math.Min(syn.Config.WeightMax, syn.Weight))
	syn.Statistics.WeightHistory = append(syn.Statistics.WeightHistory, syn.Weight)
}

// DecayTraces decays pre/post traces
func (syn *STDPSynapse) DecayTraces(dtMs float64) {
	syn.PreTrace *= math.Exp(-dtMs / syn.Config.TauPlusMs)
	syn.PostTrace *= math.Exp(-dtMs / syn.Config.TauMinusMs)
}

// GetWeight returns current weight
func (syn *STDPSynapse) GetWeight() float64 {
	return syn.Weight
}

// ============================================================================
// Memristive SNN Crossbar Array
// ============================================================================

// SNNCrossbarConfig holds SNN crossbar parameters
type SNNCrossbarConfig struct {
	InputNeurons   int
	HiddenNeurons  int
	OutputNeurons  int
	LIFConfig      *LIFNeuronConfig
	STDPConfig     *STDPConfig
	FTJConfig      *FTJConfig // For memristive synapses
	EnableSTDP     bool
}

// SNNCrossbar implements a memristive SNN crossbar array
type SNNCrossbar struct {
	Config          *SNNCrossbarConfig
	InputLayer      []*LIFNeuron
	HiddenLayer     []*LIFNeuron
	OutputLayer     []*LIFNeuron
	IHSynapses      [][]*STDPSynapse // Input to Hidden
	HOSynapses      [][]*STDPSynapse // Hidden to Output
	FTJDevices      [][]*FTJMemristor // Memristive weight storage
	Statistics      *SNNCrossbarStats
}

// SNNCrossbarStats tracks SNN statistics
type SNNCrossbarStats struct {
	TotalSpikes         int
	InputSpikes         int
	HiddenSpikes        int
	OutputSpikes        int
	AverageLatencyMs    float64
	EnergyPerInference  float64 // pJ
}

// NewSNNCrossbar creates a new SNN crossbar array
func NewSNNCrossbar(config *SNNCrossbarConfig) *SNNCrossbar {
	snn := &SNNCrossbar{
		Config:      config,
		InputLayer:  make([]*LIFNeuron, config.InputNeurons),
		HiddenLayer: make([]*LIFNeuron, config.HiddenNeurons),
		OutputLayer: make([]*LIFNeuron, config.OutputNeurons),
		IHSynapses:  make([][]*STDPSynapse, config.InputNeurons),
		HOSynapses:  make([][]*STDPSynapse, config.HiddenNeurons),
		Statistics:  &SNNCrossbarStats{},
	}

	// Initialize neurons
	for i := 0; i < config.InputNeurons; i++ {
		snn.InputLayer[i] = NewLIFNeuron(config.LIFConfig)
	}
	for i := 0; i < config.HiddenNeurons; i++ {
		snn.HiddenLayer[i] = NewLIFNeuron(config.LIFConfig)
	}
	for i := 0; i < config.OutputNeurons; i++ {
		snn.OutputLayer[i] = NewLIFNeuron(config.LIFConfig)
	}

	// Initialize synapses
	for i := 0; i < config.InputNeurons; i++ {
		snn.IHSynapses[i] = make([]*STDPSynapse, config.HiddenNeurons)
		for j := 0; j < config.HiddenNeurons; j++ {
			initialWeight := rand.Float64() * 0.5
			snn.IHSynapses[i][j] = NewSTDPSynapse(config.STDPConfig, initialWeight)
		}
	}
	for i := 0; i < config.HiddenNeurons; i++ {
		snn.HOSynapses[i] = make([]*STDPSynapse, config.OutputNeurons)
		for j := 0; j < config.OutputNeurons; j++ {
			initialWeight := rand.Float64() * 0.5
			snn.HOSynapses[i][j] = NewSTDPSynapse(config.STDPConfig, initialWeight)
		}
	}

	// Initialize FTJ devices if config provided
	if config.FTJConfig != nil {
		totalSynapses := config.InputNeurons*config.HiddenNeurons + config.HiddenNeurons*config.OutputNeurons
		rows := config.InputNeurons + config.HiddenNeurons
		cols := config.HiddenNeurons + config.OutputNeurons
		snn.FTJDevices = make([][]*FTJMemristor, rows)
		for i := 0; i < rows; i++ {
			snn.FTJDevices[i] = make([]*FTJMemristor, cols)
			for j := 0; j < cols; j++ {
				snn.FTJDevices[i][j] = NewFTJMemristor(config.FTJConfig)
			}
		}
		_ = totalSynapses // Used for statistics
	}

	return snn
}

// ForwardPass performs one timestep of SNN inference
func (snn *SNNCrossbar) ForwardPass(inputSpikes []bool, currentTimeMs float64, dtMs float64) []bool {
	// Process input spikes
	hiddenCurrents := make([]float64, snn.Config.HiddenNeurons)
	for i, spike := range inputSpikes {
		if spike {
			snn.Statistics.InputSpikes++
			// Propagate to hidden layer
			for j := 0; j < snn.Config.HiddenNeurons; j++ {
				weight := snn.IHSynapses[i][j].GetWeight()
				hiddenCurrents[j] += weight * 10 // Convert spike to current

				// STDP: pre-spike
				if snn.Config.EnableSTDP {
					snn.IHSynapses[i][j].PreSpike(currentTimeMs)
				}
			}
		}
	}

	// Update hidden layer
	hiddenSpikes := make([]bool, snn.Config.HiddenNeurons)
	for j := 0; j < snn.Config.HiddenNeurons; j++ {
		spiked := snn.HiddenLayer[j].Step(hiddenCurrents[j], currentTimeMs, dtMs)
		hiddenSpikes[j] = spiked
		if spiked {
			snn.Statistics.HiddenSpikes++
			// STDP: post-spike for IH synapses
			if snn.Config.EnableSTDP {
				for i := 0; i < snn.Config.InputNeurons; i++ {
					snn.IHSynapses[i][j].PostSpike(currentTimeMs)
				}
			}
		}
	}

	// Process hidden spikes to output
	outputCurrents := make([]float64, snn.Config.OutputNeurons)
	for i, spike := range hiddenSpikes {
		if spike {
			for j := 0; j < snn.Config.OutputNeurons; j++ {
				weight := snn.HOSynapses[i][j].GetWeight()
				outputCurrents[j] += weight * 10

				if snn.Config.EnableSTDP {
					snn.HOSynapses[i][j].PreSpike(currentTimeMs)
				}
			}
		}
	}

	// Update output layer
	outputSpikes := make([]bool, snn.Config.OutputNeurons)
	for j := 0; j < snn.Config.OutputNeurons; j++ {
		spiked := snn.OutputLayer[j].Step(outputCurrents[j], currentTimeMs, dtMs)
		outputSpikes[j] = spiked
		if spiked {
			snn.Statistics.OutputSpikes++
			if snn.Config.EnableSTDP {
				for i := 0; i < snn.Config.HiddenNeurons; i++ {
					snn.HOSynapses[i][j].PostSpike(currentTimeMs)
				}
			}
		}
	}

	// Decay STDP traces
	if snn.Config.EnableSTDP {
		for i := range snn.IHSynapses {
			for j := range snn.IHSynapses[i] {
				snn.IHSynapses[i][j].DecayTraces(dtMs)
			}
		}
		for i := range snn.HOSynapses {
			for j := range snn.HOSynapses[i] {
				snn.HOSynapses[i][j].DecayTraces(dtMs)
			}
		}
	}

	snn.Statistics.TotalSpikes = snn.Statistics.InputSpikes + snn.Statistics.HiddenSpikes + snn.Statistics.OutputSpikes

	return outputSpikes
}

// Reset resets all neurons
func (snn *SNNCrossbar) Reset() {
	for _, n := range snn.InputLayer {
		n.Reset()
	}
	for _, n := range snn.HiddenLayer {
		n.Reset()
	}
	for _, n := range snn.OutputLayer {
		n.Reset()
	}
}

// ============================================================================
// Integrated Neuromorphic CIM System
// ============================================================================

// NeuromorphicCIMConfig holds integrated system config
type NeuromorphicCIMConfig struct {
	// Architecture
	UsesFTJ          bool
	UsesNDR          bool
	UsesSNN          bool

	// Scale
	ArraySize        int
	NumNeurons       int
	NumSynapses      int

	// Learning
	EnableOnChipLearning bool
	LearningRule     string // "stdp", "bcall", "supervised"

	// Energy
	TargetEnergyPj   float64 // Target energy per spike (pJ)
}

// NeuromorphicCIMSystem integrates all neuromorphic components
type NeuromorphicCIMSystem struct {
	Config          *NeuromorphicCIMConfig
	FTJCrossbar     *FTJCrossbar
	SNNCrossbar     *SNNCrossbar
	NDRNeurons      []*NDRMemristor
	FHNNeurons      []*FHNNeuron
	Statistics      *NeuromorphicStats
}

// NeuromorphicStats tracks system statistics
type NeuromorphicStats struct {
	TotalInferences     int
	TotalSpikes         int
	AverageLatencyUs    float64
	EnergyPerSpikePj    float64
	LearningUpdates     int
	Accuracy            float64
}

// NewNeuromorphicCIMSystem creates an integrated system
func NewNeuromorphicCIMSystem(config *NeuromorphicCIMConfig) *NeuromorphicCIMSystem {
	system := &NeuromorphicCIMSystem{
		Config:     config,
		Statistics: &NeuromorphicStats{},
	}

	// Initialize FTJ crossbar
	if config.UsesFTJ {
		ftjConfig := &FTJConfig{
			ThicknessNm:       4.5,
			HfZrRatio:         0.5,
			InterlayerType:    "TiO2",
			InterlayerNm:      2.0,
			TERRatio:          580,
			OnOffRatio:        580,
			ConductanceLevels: 128,
			ReadVoltage:       0.3,
			WriteVoltage:      2.5,
			WritePulseNs:      50,
			Endurance:         2e8,
			RetentionS:        1e5,
			CycleVariation:    2.75,
		}

		cbConfig := &FTJCrossbarConfig{
			Rows:           config.ArraySize,
			Cols:           config.ArraySize,
			FTJConfig:      ftjConfig,
			LineResistance: 10,
			ADCBits:        6,
			DACBits:        8,
		}
		system.FTJCrossbar = NewFTJCrossbar(cbConfig)
	}

	// Initialize SNN crossbar
	if config.UsesSNN {
		lifConfig := &LIFNeuronConfig{
			MembraneCapPf:      10,
			LeakConductanceNs:  1,
			RestingPotentialMv: -70,
			ThresholdMv:        -55,
			ResetPotentialMv:   -70,
			RefractoryPeriodMs: 2,
		}

		stdpConfig := &STDPConfig{
			APlusMax:     0.01,
			AMinusMax:    0.012,
			TauPlusMs:    20,
			TauMinusMs:   20,
			WeightMin:    0,
			WeightMax:    1,
			LearningRate: 1.0,
		}

		snnConfig := &SNNCrossbarConfig{
			InputNeurons:  784,
			HiddenNeurons: 800,
			OutputNeurons: 10,
			LIFConfig:     lifConfig,
			STDPConfig:    stdpConfig,
			EnableSTDP:    config.EnableOnChipLearning,
		}
		system.SNNCrossbar = NewSNNCrossbar(snnConfig)
	}

	// Initialize NDR neurons
	if config.UsesNDR {
		ndrConfig := &NDRConfig{
			WellMaterial:    "InGaAs",
			InGaRatio:       0.8,
			BarrierMaterial: "AlAs",
			PeakVoltage:     0.3,
			ValleyVoltage:   0.5,
			PeakCurrent:     1e-3,
			ValleyCurrent:   1e-4,
			PVCR:            10,
			Endurance:       1e11,
			TempResistanceC: 400,
			Variation:       0.264,
		}

		fhnConfig := &FHNNeuronConfig{
			NDRDevice:     ndrConfig,
			MembraneCapF:  1e-12,
			RecoveryTau:   12.5,
			ExcitabilityA: 0.7,
			ExcitabilityB: 0.8,
			Threshold:     0.5,
		}

		system.FHNNeurons = make([]*FHNNeuron, config.NumNeurons)
		for i := 0; i < config.NumNeurons; i++ {
			system.FHNNeurons[i] = NewFHNNeuron(fhnConfig)
		}

		system.NDRNeurons = make([]*NDRMemristor, config.NumNeurons)
		for i := 0; i < config.NumNeurons; i++ {
			system.NDRNeurons[i] = NewNDRMemristor(ndrConfig)
		}
	}

	return system
}

// RunInference runs inference on input data
func (sys *NeuromorphicCIMSystem) RunInference(input []float64, durationMs float64, dtMs float64) []float64 {
	sys.Statistics.TotalInferences++

	// Encode input as spikes (rate coding)
	inputSpikes := make([]bool, len(input))

	// Run SNN simulation
	output := make([]float64, sys.SNNCrossbar.Config.OutputNeurons)
	spikeCount := make([]int, sys.SNNCrossbar.Config.OutputNeurons)

	for t := 0.0; t < durationMs; t += dtMs {
		// Generate input spikes based on rate
		for i, rate := range input {
			if rand.Float64() < rate*dtMs/1000 {
				inputSpikes[i] = true
			} else {
				inputSpikes[i] = false
			}
		}

		// Forward pass
		outputSpikes := sys.SNNCrossbar.ForwardPass(inputSpikes, t, dtMs)

		// Count output spikes
		for i, spike := range outputSpikes {
			if spike {
				spikeCount[i]++
			}
		}
	}

	// Convert spike counts to output
	for i, count := range spikeCount {
		output[i] = float64(count) / (durationMs / dtMs)
	}

	sys.Statistics.TotalSpikes += sys.SNNCrossbar.Statistics.TotalSpikes

	// Estimate energy
	sys.Statistics.EnergyPerSpikePj = 0.35 // From literature

	return output
}

// GetSystemSummary returns system summary
func (sys *NeuromorphicCIMSystem) GetSystemSummary() string {
	summary := "Neuromorphic CIM System Summary:\n"
	summary += fmt.Sprintf("  Total inferences: %d\n", sys.Statistics.TotalInferences)
	summary += fmt.Sprintf("  Total spikes: %d\n", sys.Statistics.TotalSpikes)
	summary += fmt.Sprintf("  Energy per spike: %.2f pJ\n", sys.Statistics.EnergyPerSpikePj)

	if sys.FTJCrossbar != nil {
		summary += fmt.Sprintf("  FTJ array: %dx%d\n", sys.Config.ArraySize, sys.Config.ArraySize)
		summary += fmt.Sprintf("  FTJ yield: %.1f%%\n", sys.FTJCrossbar.Statistics.ArrayYield)
	}

	if sys.SNNCrossbar != nil {
		summary += fmt.Sprintf("  SNN neurons: %d input, %d hidden, %d output\n",
			sys.SNNCrossbar.Config.InputNeurons,
			sys.SNNCrossbar.Config.HiddenNeurons,
			sys.SNNCrossbar.Config.OutputNeurons)
		summary += fmt.Sprintf("  On-chip learning: %v\n", sys.Config.EnableOnChipLearning)
	}

	if sys.NDRNeurons != nil {
		summary += fmt.Sprintf("  NDR neurons: %d\n", len(sys.NDRNeurons))
	}

	return summary
}
