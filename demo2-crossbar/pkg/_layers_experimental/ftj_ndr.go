// ftj_ndr.go - Ferroelectric Tunnel Junction and Negative Differential Resistance Devices
// Implements FTJ memristors for synaptic devices and NDR-based neuromorphic circuits
// Based on Nature Communications 2025 (NDR memristor) and HZO FTJ research (2024-2025)

package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// FERROELECTRIC TUNNEL JUNCTION (FTJ) FUNDAMENTALS
// =============================================================================

// TunnelingMechanism defines the dominant transport mechanism
type TunnelingMechanism string

const (
	TunnelDirect        TunnelingMechanism = "direct_tunneling"
	TunnelFowlerNordheim TunnelingMechanism = "fowler_nordheim"
	TunnelThermionic    TunnelingMechanism = "thermionic_emission"
	TunnelMixed         TunnelingMechanism = "mixed"
)

// FTJMaterial defines ferroelectric tunnel junction materials
type FTJMaterial string

const (
	MaterialHZO  FTJMaterial = "Hf0.5Zr0.5O2"
	MaterialBTO  FTJMaterial = "BaTiO3"
	MaterialPTO  FTJMaterial = "PbTiO3"
	MaterialBFO  FTJMaterial = "BiFeO3"
	MaterialBSO  FTJMaterial = "Sm:BiO"
)

// FTJConfig configures ferroelectric tunnel junction device
type FTJConfig struct {
	// Material properties
	Material             FTJMaterial
	FerroelectricThickness float64 // nm (typically 2-10 nm)
	InterlayerThickness  float64 // nm (e.g., TiO2, Al2O3)
	InterlayerMaterial   string

	// Ferroelectric properties
	RemanentPolarization float64 // uC/cm^2
	CoerciveField        float64 // MV/cm
	CurieTemperature     float64 // K

	// Barrier properties
	BarrierHeightHigh    float64 // eV (polarization up)
	BarrierHeightLow     float64 // eV (polarization down)
	EffectiveMass        float64 // m*/m0

	// Device properties
	Area                 float64 // um^2
	TemperatureK         float64 // Operating temperature
}

// DefaultHZOFTJConfig returns optimized HZO FTJ configuration
func DefaultHZOFTJConfig() FTJConfig {
	return FTJConfig{
		Material:             MaterialHZO,
		FerroelectricThickness: 5.0, // 5nm optimal per research
		InterlayerThickness:  1.0,
		InterlayerMaterial:   "TiO2",
		RemanentPolarization: 23.7, // uC/cm^2 (2Pr ~ 47.33)
		CoerciveField:        1.0,  // MV/cm
		CurieTemperature:     723,  // K
		BarrierHeightHigh:    1.2,  // eV
		BarrierHeightLow:     0.4,  // eV
		EffectiveMass:        0.5,
		Area:                 0.01, // 0.01 um^2 (100nm x 100nm)
		TemperatureK:         300,
	}
}

// FTJState holds current FTJ device state
type FTJState struct {
	Polarization       float64 // -1 to +1 (normalized)
	PolarizationDomains []float64 // Domain distribution
	Resistance         float64 // Current resistance (Ohm)
	CurrentDensity     float64 // A/cm^2
	TERRatio           float64 // Tunneling electroresistance ratio
	ConductanceState   int     // Discrete state index
	NumStates          int     // Total conductance states
}

// FTJDevice models a ferroelectric tunnel junction
type FTJDevice struct {
	Config FTJConfig
	State  FTJState
}

// NewFTJDevice creates a new FTJ device
func NewFTJDevice(config FTJConfig) *FTJDevice {
	ftj := &FTJDevice{
		Config: config,
		State: FTJState{
			Polarization:     0,
			NumStates:        128, // 128 states demonstrated
			ConductanceState: 64,
		},
	}

	// Initialize domain distribution
	ftj.State.PolarizationDomains = make([]float64, 100)
	for i := range ftj.State.PolarizationDomains {
		ftj.State.PolarizationDomains[i] = rand.Float64()*2 - 1
	}

	ftj.updateResistance()
	return ftj
}

// updateResistance computes resistance based on polarization
func (ftj *FTJDevice) updateResistance() {
	// Barrier height modulation by polarization
	P := ftj.State.Polarization
	phiHigh := ftj.Config.BarrierHeightHigh
	phiLow := ftj.Config.BarrierHeightLow

	// Effective barrier height interpolated by polarization
	phi := (phiHigh+phiLow)/2 + (phiHigh-phiLow)/2*P

	// Direct tunneling current density (Simmons model approximation)
	d := ftj.Config.FerroelectricThickness * 1e-9 // m
	m := ftj.Config.EffectiveMass * 9.109e-31     // kg
	hbar := 1.054e-34                             // J·s
	q := 1.602e-19                                // C

	// Tunneling probability
	kappa := math.Sqrt(2 * m * phi * q) / hbar
	T := math.Exp(-2 * kappa * d)

	// Current density at low voltage
	V := 0.1 // Read voltage
	J := (q * q * V / (4 * math.Pi * math.Pi * hbar * d * d)) * T

	ftj.State.CurrentDensity = J

	// Resistance
	area := ftj.Config.Area * 1e-12 // m^2
	current := J * area * 1e4       // A (J is per cm^2)
	if current > 1e-15 {
		ftj.State.Resistance = V / current
	} else {
		ftj.State.Resistance = 1e15
	}

	// TER ratio
	// Calculate high and low resistance states
	kappaHigh := math.Sqrt(2 * m * phiHigh * q) / hbar
	kappaLow := math.Sqrt(2 * m * phiLow * q) / hbar
	THigh := math.Exp(-2 * kappaHigh * d)
	TLow := math.Exp(-2 * kappaLow * d)

	if THigh > 1e-30 {
		ftj.State.TERRatio = TLow / THigh
	} else {
		ftj.State.TERRatio = 1e6
	}
}

// ApplyVoltage applies programming voltage pulse
func (ftj *FTJDevice) ApplyVoltage(voltage float64, duration float64) {
	// Coercive voltage
	Vc := ftj.Config.CoerciveField * ftj.Config.FerroelectricThickness * 1e-3 // V

	if math.Abs(voltage) > Vc {
		// Polarization switching
		// Nucleation-limited switching dynamics
		switchingRate := math.Abs(voltage)/Vc - 1
		if switchingRate > 0 {
			targetP := math.Copysign(1.0, voltage)
			dP := switchingRate * (targetP - ftj.State.Polarization) * duration * 1e6
			ftj.State.Polarization += dP

			// Clamp polarization
			if ftj.State.Polarization > 1 {
				ftj.State.Polarization = 1
			} else if ftj.State.Polarization < -1 {
				ftj.State.Polarization = -1
			}
		}
	}

	// Update domain distribution
	for i := range ftj.State.PolarizationDomains {
		// Domain has individual coercive field with distribution
		localVc := Vc * (0.8 + 0.4*rand.Float64())
		if math.Abs(voltage) > localVc {
			targetP := math.Copysign(1.0, voltage)
			ftj.State.PolarizationDomains[i] += 0.1 * (targetP - ftj.State.PolarizationDomains[i])
		}
	}

	// Average domain polarization
	avgP := 0.0
	for _, p := range ftj.State.PolarizationDomains {
		avgP += p
	}
	avgP /= float64(len(ftj.State.PolarizationDomains))
	ftj.State.Polarization = avgP

	ftj.updateResistance()
}

// ProgramState programs to specific conductance state
func (ftj *FTJDevice) ProgramState(stateIndex int) float64 {
	if stateIndex < 0 {
		stateIndex = 0
	}
	if stateIndex >= ftj.State.NumStates {
		stateIndex = ftj.State.NumStates - 1
	}

	targetP := float64(stateIndex)/float64(ftj.State.NumStates-1)*2 - 1
	currentP := ftj.State.Polarization

	// Apply pulses to reach target state
	energy := 0.0
	maxPulses := 100

	for i := 0; i < maxPulses && math.Abs(targetP-ftj.State.Polarization) > 0.01; i++ {
		voltage := 0.0
		if targetP > ftj.State.Polarization {
			voltage = 2.0 // Positive pulse
		} else {
			voltage = -2.0 // Negative pulse
		}

		ftj.ApplyVoltage(voltage, 50e-9) // 50ns pulse

		// Energy consumed
		current := voltage / ftj.State.Resistance
		energy += math.Abs(voltage * current * 50e-9)
	}

	ftj.State.ConductanceState = stateIndex
	_ = currentP // Silence unused variable warning
	return energy
}

// GetConductance returns current conductance
func (ftj *FTJDevice) GetConductance() float64 {
	if ftj.State.Resistance > 0 {
		return 1.0 / ftj.State.Resistance
	}
	return 0
}

// =============================================================================
// FTJ CROSSBAR ARRAY
// =============================================================================

// FTJCrossbarConfig configures FTJ crossbar array
type FTJCrossbarConfig struct {
	Rows            int
	Cols            int
	DeviceConfig    FTJConfig
	LineResistance  float64 // Ohm per unit
	VariationPct    float64 // Device-to-device variation
	StuckOnRatio    float64
	StuckOffRatio   float64
}

// DefaultFTJCrossbarConfig returns typical crossbar configuration
func DefaultFTJCrossbarConfig(rows, cols int) FTJCrossbarConfig {
	return FTJCrossbarConfig{
		Rows:           rows,
		Cols:           cols,
		DeviceConfig:   DefaultHZOFTJConfig(),
		LineResistance: 1.0,
		VariationPct:   2.75, // 2.75% cycle-to-cycle variation
		StuckOnRatio:   0.001,
		StuckOffRatio:  0.001,
	}
}

// FTJCrossbar implements FTJ-based crossbar array
type FTJCrossbar struct {
	Config     FTJCrossbarConfig
	Devices    [][]*FTJDevice
	StuckOn    [][]bool
	StuckOff   [][]bool
	TotalEnergy float64
}

// NewFTJCrossbar creates FTJ crossbar array
func NewFTJCrossbar(config FTJCrossbarConfig) *FTJCrossbar {
	cb := &FTJCrossbar{
		Config:   config,
		Devices:  make([][]*FTJDevice, config.Rows),
		StuckOn:  make([][]bool, config.Rows),
		StuckOff: make([][]bool, config.Rows),
	}

	for i := range cb.Devices {
		cb.Devices[i] = make([]*FTJDevice, config.Cols)
		cb.StuckOn[i] = make([]bool, config.Cols)
		cb.StuckOff[i] = make([]bool, config.Cols)

		for j := range cb.Devices[i] {
			// Apply device-to-device variation
			devConfig := config.DeviceConfig
			variation := 1 + (rand.Float64()*2-1)*config.VariationPct/100
			devConfig.RemanentPolarization *= variation
			devConfig.BarrierHeightHigh *= variation

			cb.Devices[i][j] = NewFTJDevice(devConfig)

			// Apply stuck defects
			if rand.Float64() < config.StuckOnRatio {
				cb.StuckOn[i][j] = true
			}
			if rand.Float64() < config.StuckOffRatio {
				cb.StuckOff[i][j] = true
			}
		}
	}

	return cb
}

// ProgramWeights programs weight matrix into FTJ crossbar
func (cb *FTJCrossbar) ProgramWeights(weights [][]float64) float64 {
	energy := 0.0
	numStates := cb.Devices[0][0].State.NumStates

	for i := 0; i < cb.Config.Rows && i < len(weights); i++ {
		for j := 0; j < cb.Config.Cols && j < len(weights[i]); j++ {
			if cb.StuckOn[i][j] || cb.StuckOff[i][j] {
				continue // Skip stuck devices
			}

			// Normalize weight to state index
			normalized := (weights[i][j] + 1) / 2 // [-1, 1] -> [0, 1]
			stateIdx := int(normalized * float64(numStates-1))
			energy += cb.Devices[i][j].ProgramState(stateIdx)
		}
	}

	cb.TotalEnergy += energy
	return energy
}

// MVM performs matrix-vector multiplication
func (cb *FTJCrossbar) MVM(input []float64) []float64 {
	output := make([]float64, cb.Config.Rows)

	for i := 0; i < cb.Config.Rows; i++ {
		for j := 0; j < cb.Config.Cols && j < len(input); j++ {
			var conductance float64

			if cb.StuckOn[i][j] {
				conductance = 1e-3 // High conductance
			} else if cb.StuckOff[i][j] {
				conductance = 1e-9 // Low conductance
			} else {
				conductance = cb.Devices[i][j].GetConductance()
			}

			// Apply variation
			conductance *= 1 + (rand.Float64()*2-1)*cb.Config.VariationPct/100

			// Weight from conductance (normalized)
			maxG := 1.0 / 1e3  // Max conductance
			minG := 1.0 / 1e9  // Min conductance
			weight := (conductance - minG) / (maxG - minG) * 2 - 1

			output[i] += weight * input[j]
		}
	}

	return output
}

// GetTERStatistics returns TER distribution across array
func (cb *FTJCrossbar) GetTERStatistics() (mean, stddev, max float64) {
	var sum, sumSq float64
	var maxTER float64
	count := 0

	for i := range cb.Devices {
		for j := range cb.Devices[i] {
			ter := cb.Devices[i][j].State.TERRatio
			sum += ter
			sumSq += ter * ter
			if ter > maxTER {
				maxTER = ter
			}
			count++
		}
	}

	mean = sum / float64(count)
	variance := sumSq/float64(count) - mean*mean
	if variance > 0 {
		stddev = math.Sqrt(variance)
	}
	max = maxTER

	return mean, stddev, max
}

// =============================================================================
// NEGATIVE DIFFERENTIAL RESISTANCE (NDR) DEVICES
// =============================================================================

// NDRType defines the type of NDR device
type NDRType string

const (
	NDRQuantumWell   NDRType = "quantum_well_rtd"
	NDRTunnelDiode   NDRType = "tunnel_diode"
	NDRThyristor     NDRType = "thyristor"
	NDRNBO2          NDRType = "nbo2_mott"
	NDRFerroelectric NDRType = "ferroelectric"
)

// NDRConfig configures NDR device
type NDRConfig struct {
	Type            NDRType
	PeakVoltage     float64 // V (voltage at peak current)
	ValleyVoltage   float64 // V (voltage at valley current)
	PeakCurrent     float64 // A
	ValleyCurrent   float64 // A
	PeakValleyRatio float64 // I_peak / I_valley
	TemperatureK    float64
	Endurance       float64 // Cycles
}

// DefaultQuantumWellNDRConfig returns config based on Nature Communications 2025
func DefaultQuantumWellNDRConfig() NDRConfig {
	return NDRConfig{
		Type:            NDRQuantumWell,
		PeakVoltage:     0.3,
		ValleyVoltage:   0.6,
		PeakCurrent:     1e-3,
		ValleyCurrent:   1e-5,
		PeakValleyRatio: 100,
		TemperatureK:    300,
		Endurance:       1e11, // >10^11 cycles at room temp
	}
}

// NDRDevice models a negative differential resistance device
type NDRDevice struct {
	Config  NDRConfig
	Voltage float64
	Current float64
	State   string // "low", "ndr", "high"
}

// NewNDRDevice creates NDR device
func NewNDRDevice(config NDRConfig) *NDRDevice {
	return &NDRDevice{
		Config: config,
		State:  "low",
	}
}

// GetCurrent returns current for given voltage (N-shaped I-V)
func (ndr *NDRDevice) GetCurrent(voltage float64) float64 {
	ndr.Voltage = voltage
	Vp := ndr.Config.PeakVoltage
	Vv := ndr.Config.ValleyVoltage
	Ip := ndr.Config.PeakCurrent
	Iv := ndr.Config.ValleyCurrent

	if voltage < 0 {
		// Symmetric for negative voltages
		return -ndr.GetCurrent(-voltage)
	}

	if voltage < Vp {
		// Rising region (positive differential resistance)
		ndr.Current = Ip * voltage / Vp
		ndr.State = "low"
	} else if voltage < Vv {
		// NDR region (negative differential resistance)
		// Cubic interpolation for smooth NDR
		t := (voltage - Vp) / (Vv - Vp)
		ndr.Current = Ip - (Ip-Iv)*(3*t*t-2*t*t*t)
		ndr.State = "ndr"
	} else {
		// High voltage region (positive differential resistance again)
		ndr.Current = Iv + (voltage-Vv)*Iv/Vv
		ndr.State = "high"
	}

	return ndr.Current
}

// GetDifferentialResistance returns dV/dI at current operating point
func (ndr *NDRDevice) GetDifferentialResistance() float64 {
	// Numerical derivative
	dV := 0.001
	I1 := ndr.GetCurrent(ndr.Voltage)
	I2 := ndr.GetCurrent(ndr.Voltage + dV)
	dI := I2 - I1

	if math.Abs(dI) < 1e-15 {
		return 1e15
	}
	return dV / dI
}

// IsInNDRRegion returns true if operating in NDR region
func (ndr *NDRDevice) IsInNDRRegion() bool {
	return ndr.State == "ndr"
}

// =============================================================================
// FITZHUGH-NAGUMO NEURON WITH NDR
// =============================================================================

// FitzHughNagumoConfig configures FHN neuron
type FitzHughNagumoConfig struct {
	// Model parameters
	A     float64 // Recovery time constant
	B     float64 // Sensitivity of recovery
	Tau   float64 // Time scale ratio (tau_v / tau_w)
	Iext  float64 // External input current

	// NDR device parameters
	NDRConfig NDRConfig

	// Threshold parameters
	VThreshold float64 // Spike threshold
	VReset     float64 // Reset voltage
}

// DefaultFHNConfig returns optimized FHN configuration
func DefaultFHNConfig() FitzHughNagumoConfig {
	return FitzHughNagumoConfig{
		A:          0.7,
		B:          0.8,
		Tau:        12.5,
		Iext:       0.5,
		NDRConfig:  DefaultQuantumWellNDRConfig(),
		VThreshold: 1.0,
		VReset:     -1.0,
	}
}

// FitzHughNagumoNeuron implements FHN neuron model with NDR
type FitzHughNagumoNeuron struct {
	Config    FitzHughNagumoConfig
	V         float64   // Membrane potential (fast variable)
	W         float64   // Recovery variable (slow variable)
	NDR       *NDRDevice
	SpikeTime float64   // Time of last spike
	Spikes    []float64 // Spike times
	Time      float64
}

// NewFitzHughNagumoNeuron creates FHN neuron
func NewFitzHughNagumoNeuron(config FitzHughNagumoConfig) *FitzHughNagumoNeuron {
	return &FitzHughNagumoNeuron{
		Config: config,
		V:      -1.0,
		W:      -0.5,
		NDR:    NewNDRDevice(config.NDRConfig),
		Spikes: make([]float64, 0),
	}
}

// Step advances neuron by one timestep
func (fhn *FitzHughNagumoNeuron) Step(dt float64, input float64) bool {
	// FitzHugh-Nagumo equations:
	// dV/dt = V - V³/3 - W + I_ext + I_input
	// dW/dt = (V + A - B*W) / tau

	// NDR contribution to membrane dynamics
	ndrCurrent := fhn.NDR.GetCurrent(fhn.V)
	ndrContrib := ndrCurrent * 1e3 // Scale NDR current

	// Cubic nullcline for V (simplified FHN)
	Vcubic := fhn.V - math.Pow(fhn.V, 3)/3

	// Update equations
	dV := (Vcubic - fhn.W + fhn.Config.Iext + input + ndrContrib) * dt
	dW := ((fhn.V + fhn.Config.A - fhn.Config.B*fhn.W) / fhn.Config.Tau) * dt

	fhn.V += dV
	fhn.W += dW
	fhn.Time += dt

	// Spike detection
	spiked := false
	if fhn.V > fhn.Config.VThreshold && (fhn.Time-fhn.SpikeTime) > 0.01 {
		spiked = true
		fhn.Spikes = append(fhn.Spikes, fhn.Time)
		fhn.SpikeTime = fhn.Time
	}

	return spiked
}

// GetFiringRate returns firing rate in Hz
func (fhn *FitzHughNagumoNeuron) GetFiringRate() float64 {
	if len(fhn.Spikes) < 2 {
		return 0
	}
	duration := fhn.Spikes[len(fhn.Spikes)-1] - fhn.Spikes[0]
	if duration > 0 {
		return float64(len(fhn.Spikes)-1) / duration
	}
	return 0
}

// GetNeuronBehavior returns current behavior type
func (fhn *FitzHughNagumoNeuron) GetNeuronBehavior() string {
	// Determine behavior based on dynamics
	if len(fhn.Spikes) == 0 {
		return "quiescent"
	}

	rate := fhn.GetFiringRate()
	if rate < 1 {
		return "excitable"
	} else if rate < 10 {
		return "phasic_spiking"
	} else if rate < 50 {
		return "tonic_spiking"
	} else {
		return "bursting"
	}
}

// =============================================================================
// NDR-BASED OSCILLATOR NETWORK
// =============================================================================

// OscillatorNetworkConfig configures NDR oscillator network
type OscillatorNetworkConfig struct {
	NumOscillators   int
	CouplingStrength float64
	CouplingTopology string // "all_to_all", "ring", "random"
	NDRConfig        NDRConfig
}

// OscillatorNetwork implements coupled NDR oscillator network
type OscillatorNetwork struct {
	Config      OscillatorNetworkConfig
	Oscillators []*FitzHughNagumoNeuron
	Coupling    [][]float64 // Coupling matrix
	Time        float64
}

// NewOscillatorNetwork creates coupled oscillator network
func NewOscillatorNetwork(config OscillatorNetworkConfig) *OscillatorNetwork {
	on := &OscillatorNetwork{
		Config:      config,
		Oscillators: make([]*FitzHughNagumoNeuron, config.NumOscillators),
		Coupling:    make([][]float64, config.NumOscillators),
	}

	fhnConfig := DefaultFHNConfig()
	fhnConfig.NDRConfig = config.NDRConfig

	for i := range on.Oscillators {
		on.Oscillators[i] = NewFitzHughNagumoNeuron(fhnConfig)
		// Random initial conditions
		on.Oscillators[i].V = rand.Float64()*2 - 1
		on.Oscillators[i].W = rand.Float64()*2 - 1
	}

	// Initialize coupling matrix
	for i := range on.Coupling {
		on.Coupling[i] = make([]float64, config.NumOscillators)
		for j := range on.Coupling[i] {
			if i != j {
				switch config.CouplingTopology {
				case "all_to_all":
					on.Coupling[i][j] = config.CouplingStrength
				case "ring":
					if j == (i+1)%config.NumOscillators || j == (i-1+config.NumOscillators)%config.NumOscillators {
						on.Coupling[i][j] = config.CouplingStrength
					}
				case "random":
					if rand.Float64() < 0.3 {
						on.Coupling[i][j] = config.CouplingStrength * rand.Float64()
					}
				}
			}
		}
	}

	return on
}

// Step advances all oscillators
func (on *OscillatorNetwork) Step(dt float64, inputs []float64) []bool {
	n := len(on.Oscillators)
	spikes := make([]bool, n)

	// Compute coupling inputs
	couplingInputs := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j {
				// Diffusive coupling based on voltage difference
				couplingInputs[i] += on.Coupling[i][j] * (on.Oscillators[j].V - on.Oscillators[i].V)
			}
		}
	}

	// Update each oscillator
	for i := 0; i < n; i++ {
		extInput := couplingInputs[i]
		if i < len(inputs) {
			extInput += inputs[i]
		}
		spikes[i] = on.Oscillators[i].Step(dt, extInput)
	}

	on.Time += dt
	return spikes
}

// ComputeSynchronization returns order parameter (0=desync, 1=sync)
func (on *OscillatorNetwork) ComputeSynchronization() float64 {
	// Kuramoto order parameter: R = |1/N * sum(e^(i*theta_j))|
	// Approximate phase from (V, W) coordinates

	var sumCos, sumSin float64
	n := float64(len(on.Oscillators))

	for _, osc := range on.Oscillators {
		// Phase approximation
		phase := math.Atan2(osc.W, osc.V)
		sumCos += math.Cos(phase)
		sumSin += math.Sin(phase)
	}

	R := math.Sqrt(sumCos*sumCos+sumSin*sumSin) / n
	return R
}

// =============================================================================
// STDP WITH FTJ SYNAPSES
// =============================================================================

// FTJSynapseConfig configures FTJ-based synapse
type FTJSynapseConfig struct {
	FTJConfig    FTJConfig
	STDPEnabled  bool
	LTPAmplitude float64 // Long-term potentiation amplitude
	LTDAmplitude float64 // Long-term depression amplitude
	TauPlus      float64 // LTP time constant (ms)
	TauMinus     float64 // LTD time constant (ms)
}

// FTJSynapse implements FTJ-based synapse with STDP
type FTJSynapse struct {
	Config       FTJSynapseConfig
	Device       *FTJDevice
	PreSpikeTime float64 // Last pre-synaptic spike
	PostSpikeTime float64 // Last post-synaptic spike
}

// NewFTJSynapse creates FTJ synapse
func NewFTJSynapse(config FTJSynapseConfig) *FTJSynapse {
	return &FTJSynapse{
		Config:        config,
		Device:        NewFTJDevice(config.FTJConfig),
		PreSpikeTime:  -1000,
		PostSpikeTime: -1000,
	}
}

// OnPreSpike handles pre-synaptic spike
func (syn *FTJSynapse) OnPreSpike(time float64) {
	syn.PreSpikeTime = time

	if syn.Config.STDPEnabled && syn.PostSpikeTime > 0 {
		// LTD: pre after post
		dt := time - syn.PostSpikeTime
		if dt > 0 && dt < 100 {
			// Depression
			deltaW := -syn.Config.LTDAmplitude * math.Exp(-dt/syn.Config.TauMinus)
			syn.updateWeight(deltaW)
		}
	}
}

// OnPostSpike handles post-synaptic spike
func (syn *FTJSynapse) OnPostSpike(time float64) {
	syn.PostSpikeTime = time

	if syn.Config.STDPEnabled && syn.PreSpikeTime > 0 {
		// LTP: post after pre
		dt := time - syn.PreSpikeTime
		if dt > 0 && dt < 100 {
			// Potentiation
			deltaW := syn.Config.LTPAmplitude * math.Exp(-dt/syn.Config.TauPlus)
			syn.updateWeight(deltaW)
		}
	}
}

// updateWeight applies weight change via FTJ programming
func (syn *FTJSynapse) updateWeight(delta float64) {
	// Convert weight change to polarization change
	currentP := syn.Device.State.Polarization
	newP := currentP + delta

	// Clamp
	if newP > 1 {
		newP = 1
	} else if newP < -1 {
		newP = -1
	}

	// Apply voltage to reach new polarization
	voltage := (newP - currentP) * 2.0 // Scaled voltage
	syn.Device.ApplyVoltage(voltage, 50e-9)
}

// GetWeight returns current synaptic weight
func (syn *FTJSynapse) GetWeight() float64 {
	// Weight proportional to conductance
	G := syn.Device.GetConductance()
	// Normalize to [-1, 1]
	maxG := 1.0 / 1e3
	minG := 1.0 / 1e9
	return (G-minG)/(maxG-minG)*2 - 1
}

// =============================================================================
// FECIM FTJ INTEGRATION
// =============================================================================

// FeCIMFTJConfig configures HZO-based FTJ for FeCIM
type FeCIMFTJConfig struct {
	// HZO superlattice properties
	HfZrRatio          float64 // Typically 0.5 for Hf0.5Zr0.5O2
	SuperlatticeRepeat int     // Number of HfO2/ZrO2 repeats
	LayerThickness     float64 // Individual layer thickness (nm)

	// FTJ specific
	TotalThickness     float64 // Total ferroelectric thickness
	InterlayerMaterial string  // TiO2, Al2O3, etc.
	InterlayerThickness float64

	// Target performance
	TargetTER          float64 // Target TER ratio
	TargetEndurance    float64 // Target endurance cycles
	TargetRetention    float64 // Target retention (seconds)
}

// DefaultFeCIMFTJConfig returns FeCIM-optimized FTJ config
func DefaultFeCIMFTJConfig() FeCIMFTJConfig {
	return FeCIMFTJConfig{
		HfZrRatio:          0.5,
		SuperlatticeRepeat: 4,
		LayerThickness:     1.25, // nm
		TotalThickness:     5.0,  // nm (4 × 1.25)
		InterlayerMaterial: "TiO2",
		InterlayerThickness: 1.0,
		TargetTER:          3000, // Based on 5nm HZO research
		TargetEndurance:    2e8,  // 2×10^8 cycles
		TargetRetention:    1e5,  // >10^5 seconds
	}
}

// FeCIMFTJArray implements FeCIM-optimized FTJ array
type FeCIMFTJArray struct {
	Config      FeCIMFTJConfig
	Crossbar    *FTJCrossbar
	SynapticOps int64
	TotalEnergy float64
}

// NewFeCIMFTJArray creates FeCIM FTJ array
func NewFeCIMFTJArray(rows, cols int, config FeCIMFTJConfig) *FeCIMFTJArray {
	// Create FTJ config based on FeCIM parameters
	ftjConfig := FTJConfig{
		Material:             MaterialHZO,
		FerroelectricThickness: config.TotalThickness,
		InterlayerThickness:  config.InterlayerThickness,
		InterlayerMaterial:   config.InterlayerMaterial,
		RemanentPolarization: 23.7 * (1 + 0.1*float64(config.SuperlatticeRepeat-4)), // Enhanced by superlattice
		CoerciveField:        0.85, // Optimized for superlattice
		CurieTemperature:     723,
		BarrierHeightHigh:    1.2,
		BarrierHeightLow:     0.4,
		EffectiveMass:        0.5,
		Area:                 0.0016, // 40nm × 40nm (4F² at 20nm node)
		TemperatureK:         300,
	}

	cbConfig := FTJCrossbarConfig{
		Rows:           rows,
		Cols:           cols,
		DeviceConfig:   ftjConfig,
		LineResistance: 1.0,
		VariationPct:   2.75,
		StuckOnRatio:   0.0001,
		StuckOffRatio:  0.0001,
	}

	return &FeCIMFTJArray{
		Config:   config,
		Crossbar: NewFTJCrossbar(cbConfig),
	}
}

// Forward performs synaptic operation
func (ila *FeCIMFTJArray) Forward(input []float64) []float64 {
	output := ila.Crossbar.MVM(input)
	ila.SynapticOps += int64(ila.Crossbar.Config.Rows * ila.Crossbar.Config.Cols)
	return output
}

// GetEnergyPerSpike returns energy per synaptic operation
func (ila *FeCIMFTJArray) GetEnergyPerSpike() float64 {
	// Based on research: 1.8 pJ per spike for 3D FTJ
	return 1.8 // pJ
}

// GetPerformanceMetrics returns array performance
func (ila *FeCIMFTJArray) GetPerformanceMetrics() map[string]float64 {
	terMean, terStd, terMax := ila.Crossbar.GetTERStatistics()

	return map[string]float64{
		"TER_mean":           terMean,
		"TER_stddev":         terStd,
		"TER_max":            terMax,
		"Target_TER":         ila.Config.TargetTER,
		"Endurance_cycles":   ila.Config.TargetEndurance,
		"Retention_seconds":  ila.Config.TargetRetention,
		"Energy_per_spike_pJ": ila.GetEnergyPerSpike(),
		"Cell_size_F2":       4, // 4F² cell size
		"Total_synaptic_ops": float64(ila.SynapticOps),
	}
}

// =============================================================================
// BENCHMARKS
// =============================================================================

// FTJBenchmarkResult holds FTJ benchmark results
type FTJBenchmarkResult struct {
	Material        FTJMaterial
	Thickness       float64
	TERRatio        float64
	Endurance       float64
	States          int
	Variation       float64
	EnergyPerState  float64
	CurrentDensity  float64
}

// RunFTJBenchmarks returns benchmark comparisons
func RunFTJBenchmarks() []FTJBenchmarkResult {
	return []FTJBenchmarkResult{
		{
			Material:       MaterialHZO,
			Thickness:      5.0,
			TERRatio:       2974.44,
			Endurance:      2e8,
			States:         128,
			Variation:      2.75,
			EnergyPerState: 1.8,
			CurrentDensity: 75, // A/cm² at 0.1V
		},
		{
			Material:       MaterialBTO,
			Thickness:      3.6,
			TERRatio:       500,
			Endurance:      1e6,
			States:         32,
			Variation:      5.0,
			EnergyPerState: 10,
			CurrentDensity: 10,
		},
		{
			Material:       MaterialBSO,
			Thickness:      1.0,
			TERRatio:       7e5,
			Endurance:      1e5,
			States:         2,
			Variation:      10,
			EnergyPerState: 100,
			CurrentDensity: 1000,
		},
		{
			Material:       MaterialPTO,
			Thickness:      3.6,
			TERRatio:       500,
			Endurance:      1e4,
			States:         16,
			Variation:      8,
			EnergyPerState: 50,
			CurrentDensity: 5,
		},
	}
}

// NDRBenchmarkResult holds NDR device benchmark results
type NDRBenchmarkResult struct {
	Type            NDRType
	PeakValleyRatio float64
	Endurance       float64
	MaxTemp         float64 // °C
	Variation       float64 // %
	NeuronFunctions int
}

// RunNDRBenchmarks returns NDR benchmark comparisons
func RunNDRBenchmarks() []NDRBenchmarkResult {
	return []NDRBenchmarkResult{
		{
			Type:            NDRQuantumWell,
			PeakValleyRatio: 100,
			Endurance:       1e11,
			MaxTemp:         400,
			Variation:       0.264,
			NeuronFunctions: 9,
		},
		{
			Type:            NDRTunnelDiode,
			PeakValleyRatio: 10,
			Endurance:       1e9,
			MaxTemp:         150,
			Variation:       5,
			NeuronFunctions: 3,
		},
		{
			Type:            NDRNBO2,
			PeakValleyRatio: 1000,
			Endurance:       1e6,
			MaxTemp:         85,
			Variation:       10,
			NeuronFunctions: 5,
		},
	}
}
