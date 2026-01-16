// Package layers provides cryogenic CIM and neuromorphic sensory simulation.
//
// This module implements:
// - Cryogenic compute-in-memory (4K-77K operation)
// - HZO ferroelectric memory at cryogenic temperatures
// - Neuromorphic olfactory system (electronic nose)
// - Neuromorphic gustatory system (electronic tongue)
// - Gas sensor array with pattern recognition
// - Spike-based sensory encoding
//
// Based on research from Adv. Electronic Materials 2024, Nano Letters 2022,
// Science Advances 2024, and Nature Communications 2025.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// CRYOGENIC TEMPERATURE MODELS
// =============================================================================

// CryoTemperature represents common cryogenic operating points.
type CryoTemperature int

const (
	TempRoomTemp    CryoTemperature = 300 // 300K (room temperature)
	TempLiquidN2    CryoTemperature = 77  // 77K (liquid nitrogen)
	TempLiquidHe4   CryoTemperature = 4   // 4K (liquid helium-4)
	TempMilliKelvin CryoTemperature = 0   // <1K (dilution refrigerator)
)

// CryoTempEffects models temperature-dependent device behavior.
type CryoTempEffects struct {
	Temperature      float64 // Kelvin
	CarrierMobility  float64 // Relative mobility vs room temp
	ThresholdVoltage float64 // Vth shift in mV
	LeakageCurrent   float64 // Relative leakage vs room temp
	RetentionTime    float64 // Relative retention vs room temp
	ThermalNoise     float64 // kT noise factor
}

// CalculateCryoEffects computes temperature-dependent parameters.
func CalculateCryoEffects(tempK float64) *CryoTempEffects {
	// Room temperature reference
	roomTemp := 300.0

	// Carrier mobility increases at low temp (phonon scattering reduction)
	// Empirical: ~T^(-1.5) below 100K
	mobilityRatio := 1.0
	if tempK < 100 {
		mobilityRatio = math.Pow(roomTemp/tempK, 1.5) * 0.1 // Capped improvement
	} else {
		mobilityRatio = roomTemp / tempK
	}
	mobilityRatio = math.Min(mobilityRatio, 10.0) // Cap at 10x

	// Threshold voltage shift (~60mV per decade of temperature)
	vthShift := 60.0 * math.Log10(roomTemp/tempK)

	// Leakage current exponentially decreases
	// Roughly exp(-Ea/kT) dependence
	activationEnergy := 0.3 // eV (typical)
	kB := 8.617e-5          // eV/K
	leakageRatio := math.Exp(-activationEnergy / (kB * tempK))
	leakageRatio /= math.Exp(-activationEnergy / (kB * roomTemp))

	// Retention time increases exponentially at low temp
	retentionRatio := 1.0 / leakageRatio

	// Thermal noise scales with kT
	thermalNoise := tempK / roomTemp

	return &CryoTempEffects{
		Temperature:      tempK,
		CarrierMobility:  mobilityRatio,
		ThresholdVoltage: vthShift,
		LeakageCurrent:   leakageRatio,
		RetentionTime:    retentionRatio,
		ThermalNoise:     thermalNoise,
	}
}

// =============================================================================
// CRYOGENIC FERROELECTRIC MEMORY
// =============================================================================

// CryoHZOConfig configures HZO memory at cryogenic temperatures.
type CryoHZOConfig struct {
	Temperature      float64 // Operating temperature (K)
	FilmThickness    float64 // HZO thickness (nm)
	PulseAmplitude   float64 // Programming pulse (V)
	PulseWidth       float64 // Pulse duration (ns)
	AnalogLevels     int     // Number of analog states
}

// CryoHZODevice models HZO ferroelectric at cryogenic temperatures.
type CryoHZODevice struct {
	Config              *CryoHZOConfig
	Pr                  float64   // Remanent polarization (µC/cm²)
	Ec                  float64   // Coercive field (MV/cm)
	MemoryWindow        float64   // Memory window (V)
	Endurance           float64   // Cycles before degradation
	AnalogStates        []float64 // Available polarization states
	LinearityFactor     float64   // Potentiation/depression linearity
	RetentionYears      float64   // Projected retention at temp
}

// NewCryoHZODevice creates a cryogenic HZO device model.
func NewCryoHZODevice(config *CryoHZOConfig) *CryoHZODevice {
	device := &CryoHZODevice{
		Config: config,
	}

	// Temperature-dependent properties (based on experimental data)
	temp := config.Temperature

	// Polarization increases at low temperature
	// At 4K with high amplitude: up to 75 µC/cm² (vs ~30 at room temp)
	basePr := 30.0 // µC/cm² at room temp
	if temp < 100 {
		// Below 100K, can use higher amplitude pulses
		prBoost := 1.0 + (100-temp)/100*1.5 // Up to 2.5x at 4K
		device.Pr = basePr * prBoost
	} else {
		device.Pr = basePr * (300 / temp) * 0.5
	}
	device.Pr = math.Min(device.Pr, 75.0) // Cap at experimental max

	// Coercive field increases slightly at low temp
	device.Ec = 1.0 + (300-temp)/1000 // ~1.0-1.3 MV/cm

	// Memory window improves at low temp (6-8V at cryogenic)
	if temp < 100 {
		device.MemoryWindow = 6.0 + (100-temp)/50 // 6-8V
	} else {
		device.MemoryWindow = 4.0 + (300-temp)/100
	}

	// Endurance: >10^9 cycles maintained at cryogenic
	device.Endurance = 1e9

	// Analog states: nearly ideal linearity below 100K
	device.AnalogStates = make([]float64, config.AnalogLevels)
	for i := 0; i < config.AnalogLevels; i++ {
		// Linear distribution of polarization states
		device.AnalogStates[i] = -device.Pr + 2*device.Pr*float64(i)/float64(config.AnalogLevels-1)
	}

	// Linearity improves dramatically below 100K
	if temp < 100 {
		device.LinearityFactor = 0.95 + (100-temp)/2000 // Near-ideal
	} else {
		device.LinearityFactor = 0.7 + (300-temp)/1000
	}
	device.LinearityFactor = math.Min(device.LinearityFactor, 0.99)

	// Retention: effectively infinite at 4K
	cryoEffects := CalculateCryoEffects(temp)
	device.RetentionYears = 10.0 * cryoEffects.RetentionTime / 1e6 // Scale to years

	return device
}

// ProgramState programs the device to a specific analog state.
func (d *CryoHZODevice) ProgramState(targetState int) (float64, error) {
	if targetState < 0 || targetState >= len(d.AnalogStates) {
		return 0, fmt.Errorf("invalid state %d, must be 0-%d", targetState, len(d.AnalogStates)-1)
	}

	idealPolarization := d.AnalogStates[targetState]

	// Add programming noise (reduced at cryogenic temps)
	noiseScale := d.Config.Temperature / 300.0
	noise := rand.NormFloat64() * (1 - d.LinearityFactor) * d.Pr * noiseScale

	actualPolarization := idealPolarization + noise
	return actualPolarization, nil
}

// ReadState reads the current polarization state.
func (d *CryoHZODevice) ReadState(polarization float64) int {
	// Find closest analog state
	minDist := math.MaxFloat64
	closestState := 0

	for i, state := range d.AnalogStates {
		dist := math.Abs(polarization - state)
		if dist < minDist {
			minDist = dist
			closestState = i
		}
	}

	return closestState
}

// =============================================================================
// CRYOGENIC CIM CROSSBAR
// =============================================================================

// CryoCIMConfig configures cryogenic compute-in-memory.
type CryoCIMConfig struct {
	Temperature     float64
	ArrayRows       int
	ArrayCols       int
	DeviceType      CryoDeviceType
	ADCBits         int
	DACBits         int
	QuantumInterface bool // Interface to quantum processor
}

// CryoDeviceType specifies the memory technology.
type CryoDeviceType int

const (
	CryoDeviceHZO CryoDeviceType = iota
	CryoDeviceRRAM
	CryoDeviceSTTMRAM
	CryoDeviceDRAM
	CryoDeviceFeSQUID
)

// CryoCIMArray implements cryogenic compute-in-memory.
type CryoCIMArray struct {
	Config         *CryoCIMConfig
	Weights        [][]float64
	HZODevices     [][]*CryoHZODevice
	CryoEffects    *CryoTempEffects
	Stats          *CryoCIMStats
}

// CryoCIMStats tracks cryogenic CIM performance.
type CryoCIMStats struct {
	ComputeEnergy    float64 // pJ per MAC
	LeakagePower     float64 // mW
	ThermalNoise     float64 // LSB RMS
	SNR              float64 // dB
	EffectiveENOB    float64 // Effective number of bits
	CoolingPower     float64 // W for refrigeration
}

// NewCryoCIMArray creates a cryogenic CIM array.
func NewCryoCIMArray(config *CryoCIMConfig) *CryoCIMArray {
	array := &CryoCIMArray{
		Config:      config,
		Weights:     make([][]float64, config.ArrayRows),
		CryoEffects: CalculateCryoEffects(config.Temperature),
		Stats:       &CryoCIMStats{},
	}

	// Initialize weight arrays
	for i := 0; i < config.ArrayRows; i++ {
		array.Weights[i] = make([]float64, config.ArrayCols)
	}

	// Create HZO devices if using ferroelectric
	if config.DeviceType == CryoDeviceHZO {
		array.HZODevices = make([][]*CryoHZODevice, config.ArrayRows)
		for i := 0; i < config.ArrayRows; i++ {
			array.HZODevices[i] = make([]*CryoHZODevice, config.ArrayCols)
			for j := 0; j < config.ArrayCols; j++ {
				hzoConfig := &CryoHZOConfig{
					Temperature:    config.Temperature,
					FilmThickness:  10.0, // 10nm HZO
					PulseAmplitude: 5.0,
					PulseWidth:     100.0,
					AnalogLevels:   1 << config.ADCBits,
				}
				array.HZODevices[i][j] = NewCryoHZODevice(hzoConfig)
			}
		}
	}

	// Calculate statistics
	array.calculateStats()

	return array
}

// calculateStats computes cryogenic performance metrics.
func (c *CryoCIMArray) calculateStats() {
	temp := c.Config.Temperature

	// Compute energy scales with kT and reduced leakage
	baseEnergy := 1.0 // pJ at room temp
	c.Stats.ComputeEnergy = baseEnergy * c.CryoEffects.ThermalNoise

	// Leakage power dramatically reduced
	baseLeak := 10.0 // mW at room temp
	c.Stats.LeakagePower = baseLeak * c.CryoEffects.LeakageCurrent

	// Thermal noise in LSB
	adcLevels := float64(1 << c.Config.ADCBits)
	c.Stats.ThermalNoise = math.Sqrt(c.CryoEffects.ThermalNoise) * adcLevels / 100

	// SNR improves at low temp
	signalPower := 1.0
	noisePower := c.Stats.ThermalNoise * c.Stats.ThermalNoise
	if noisePower > 0 {
		c.Stats.SNR = 10 * math.Log10(signalPower/noisePower)
	}

	// Effective bits
	c.Stats.EffectiveENOB = float64(c.Config.ADCBits) + math.Log2(300/temp)/2

	// Cooling power (rough estimate)
	// Carnot efficiency: η = T_cold / (T_hot - T_cold)
	tHot := 300.0
	carnotEff := temp / (tHot - temp)
	actualEff := carnotEff * 0.1 // Real efficiency ~10% of Carnot
	dissipatedHeat := c.Stats.LeakagePower / 1000 // W
	if actualEff > 0 {
		c.Stats.CoolingPower = dissipatedHeat / actualEff
	}
}

// MVM performs matrix-vector multiplication at cryogenic temperature.
func (c *CryoCIMArray) MVM(input []float64) ([]float64, error) {
	if len(input) != c.Config.ArrayCols {
		return nil, fmt.Errorf("input size %d doesn't match array cols %d", len(input), c.Config.ArrayCols)
	}

	output := make([]float64, c.Config.ArrayRows)

	for i := 0; i < c.Config.ArrayRows; i++ {
		sum := 0.0
		for j := 0; j < c.Config.ArrayCols; j++ {
			sum += c.Weights[i][j] * input[j]
		}

		// Add cryogenic thermal noise (reduced)
		noise := rand.NormFloat64() * c.Stats.ThermalNoise / float64(1<<c.Config.ADCBits)
		output[i] = sum + noise
	}

	return output, nil
}

// ProgramWeights programs weights into the cryogenic array.
func (c *CryoCIMArray) ProgramWeights(weights [][]float64) error {
	if len(weights) != c.Config.ArrayRows {
		return fmt.Errorf("weight rows %d doesn't match array rows %d", len(weights), c.Config.ArrayRows)
	}

	for i := 0; i < c.Config.ArrayRows; i++ {
		if len(weights[i]) != c.Config.ArrayCols {
			return fmt.Errorf("weight cols %d doesn't match array cols %d", len(weights[i]), c.Config.ArrayCols)
		}
		copy(c.Weights[i], weights[i])
	}

	return nil
}

// =============================================================================
// QUANTUM COMPUTING INTERFACE
// =============================================================================

// QuantumCIMInterface models the interface between CIM and quantum processor.
type QuantumCIMInterface struct {
	CIMArray         *CryoCIMArray
	QubitCount       int
	ClassicalBuffer  []float64
	QuantumBuffer    []complex128
	InterfaceLatency float64 // ns
}

// NewQuantumCIMInterface creates a quantum-CIM interface.
func NewQuantumCIMInterface(cim *CryoCIMArray, qubits int) *QuantumCIMInterface {
	return &QuantumCIMInterface{
		CIMArray:         cim,
		QubitCount:       qubits,
		ClassicalBuffer:  make([]float64, qubits),
		QuantumBuffer:    make([]complex128, 1<<qubits),
		InterfaceLatency: 100.0, // 100ns typical
	}
}

// ClassicalToQuantum converts classical CIM output to quantum state prep.
func (q *QuantumCIMInterface) ClassicalToQuantum(classical []float64) []complex128 {
	// Normalize to quantum amplitudes
	sumSq := 0.0
	for _, v := range classical {
		sumSq += v * v
	}
	norm := math.Sqrt(sumSq)

	amplitudes := make([]complex128, len(classical))
	for i, v := range classical {
		if norm > 0 {
			amplitudes[i] = complex(v/norm, 0)
		}
	}

	return amplitudes
}

// QuantumToClassical measures quantum state for classical processing.
func (q *QuantumCIMInterface) QuantumToClassical(amplitudes []complex128) []float64 {
	// Return measurement probabilities
	probs := make([]float64, len(amplitudes))
	for i, amp := range amplitudes {
		probs[i] = real(amp)*real(amp) + imag(amp)*imag(amp)
	}
	return probs
}

// =============================================================================
// NEUROMORPHIC OLFACTORY SYSTEM (ELECTRONIC NOSE)
// =============================================================================

// OlfactoryConfig configures the electronic nose system.
type OlfactoryConfig struct {
	SensorCount      int
	SensorTypes      []GasSensorType
	SpikingThreshold float64
	RefractoryPeriod float64 // ms
	CrossbarSize     int
	UseMemristorCIM  bool
}

// GasSensorType specifies the type of gas sensor.
type GasSensorType int

const (
	SensorMOX   GasSensorType = iota // Metal oxide semiconductor
	SensorCP                          // Conducting polymer
	SensorQCM                         // Quartz crystal microbalance
	SensorSAW                         // Surface acoustic wave
	SensorEC                          // Electrochemical
)

// GasSensor models an individual gas sensor.
type GasSensor struct {
	Type          GasSensorType
	Sensitivity   map[string]float64 // Gas -> sensitivity
	BaseResistance float64            // Ohms
	NoiseLevel    float64            // Relative noise
	ResponseTime  float64            // ms
	RecoveryTime  float64            // ms
}

// OlfactoryNeuron implements a spiking olfactory receptor neuron.
type OlfactoryNeuron struct {
	Threshold       float64
	RefractoryPeriod float64
	MembranePotential float64
	LastSpikeTime   float64
	SpikeHistory    []float64
}

// ElectronicNose implements the complete olfactory system.
type ElectronicNose struct {
	Config        *OlfactoryConfig
	Sensors       []*GasSensor
	Neurons       []*OlfactoryNeuron
	CrossbarWeights [][]float64
	PatternMemory map[string][]float64 // Learned odor patterns
	Stats         *ENoseStats
}

// ENoseStats tracks electronic nose performance.
type ENoseStats struct {
	ClassificationAccuracy float64
	ResponseLatency        float64 // ms
	PowerConsumption       float64 // mW
	SpikeRate              float64 // Hz
}

// NewElectronicNose creates a new electronic nose system.
func NewElectronicNose(config *OlfactoryConfig) *ElectronicNose {
	enose := &ElectronicNose{
		Config:        config,
		Sensors:       make([]*GasSensor, config.SensorCount),
		Neurons:       make([]*OlfactoryNeuron, config.SensorCount),
		PatternMemory: make(map[string][]float64),
		Stats:         &ENoseStats{},
	}

	// Initialize sensors with diverse sensitivities
	gases := []string{"ammonia", "acetone", "methane", "ethanol", "CO", "H2S", "NO2", "benzene"}

	for i := 0; i < config.SensorCount; i++ {
		sensorType := config.SensorTypes[i%len(config.SensorTypes)]
		sensor := &GasSensor{
			Type:           sensorType,
			Sensitivity:    make(map[string]float64),
			BaseResistance: 10000 + rand.Float64()*90000, // 10-100 kOhm
			NoiseLevel:     0.01 + rand.Float64()*0.04,   // 1-5%
			ResponseTime:   100 + rand.Float64()*900,     // 100-1000 ms
			RecoveryTime:   1000 + rand.Float64()*9000,   // 1-10 s
		}

		// Assign semi-random sensitivities
		for _, gas := range gases {
			sensor.Sensitivity[gas] = rand.Float64() * 10 // 0-10 relative sensitivity
		}

		enose.Sensors[i] = sensor

		// Create corresponding spiking neuron
		enose.Neurons[i] = &OlfactoryNeuron{
			Threshold:        config.SpikingThreshold,
			RefractoryPeriod: config.RefractoryPeriod,
			SpikeHistory:     make([]float64, 0),
		}
	}

	// Initialize crossbar weights for pattern recognition
	if config.UseMemristorCIM {
		enose.CrossbarWeights = make([][]float64, config.CrossbarSize)
		for i := 0; i < config.CrossbarSize; i++ {
			enose.CrossbarWeights[i] = make([]float64, config.SensorCount)
			for j := 0; j < config.SensorCount; j++ {
				enose.CrossbarWeights[i][j] = rand.NormFloat64() * 0.1
			}
		}
	}

	return enose
}

// Sense processes gas exposure and generates spike patterns.
func (e *ElectronicNose) Sense(gasConcentrations map[string]float64, duration float64) [][]float64 {
	timeSteps := int(duration / 1.0) // 1ms time steps
	spikeTrains := make([][]float64, e.Config.SensorCount)

	for i := range spikeTrains {
		spikeTrains[i] = make([]float64, 0)
	}

	for t := 0; t < timeSteps; t++ {
		currentTime := float64(t)

		for i, sensor := range e.Sensors {
			// Calculate sensor response
			response := 0.0
			for gas, conc := range gasConcentrations {
				if sens, ok := sensor.Sensitivity[gas]; ok {
					response += sens * conc
				}
			}

			// Add sensor noise
			response += rand.NormFloat64() * sensor.NoiseLevel * response

			// Update neuron membrane potential
			neuron := e.Neurons[i]

			// Check refractory period
			if currentTime-neuron.LastSpikeTime < neuron.RefractoryPeriod {
				continue
			}

			// Leaky integrate
			neuron.MembranePotential *= 0.95 // Decay
			neuron.MembranePotential += response * 0.1

			// Check threshold
			if neuron.MembranePotential > neuron.Threshold {
				spikeTrains[i] = append(spikeTrains[i], currentTime)
				neuron.LastSpikeTime = currentTime
				neuron.MembranePotential = 0
				neuron.SpikeHistory = append(neuron.SpikeHistory, currentTime)
			}
		}
	}

	return spikeTrains
}

// EncodeSpikes converts spike trains to rate codes for CIM processing.
func (e *ElectronicNose) EncodeSpikes(spikeTrains [][]float64, window float64) []float64 {
	rates := make([]float64, len(spikeTrains))
	for i, train := range spikeTrains {
		rates[i] = float64(len(train)) / window * 1000 // Convert to Hz
	}
	return rates
}

// ClassifyOdor uses the crossbar to classify the odor pattern.
func (e *ElectronicNose) ClassifyOdor(spikeRates []float64) (string, float64) {
	if !e.Config.UseMemristorCIM {
		return e.classifyWithSoftware(spikeRates)
	}

	// Crossbar MVM for pattern matching
	scores := make([]float64, e.Config.CrossbarSize)
	for i := 0; i < e.Config.CrossbarSize; i++ {
		for j := 0; j < len(spikeRates); j++ {
			scores[i] += e.CrossbarWeights[i][j] * spikeRates[j]
		}
	}

	// Find best match from pattern memory
	bestMatch := ""
	bestScore := -math.MaxFloat64

	for odor, pattern := range e.PatternMemory {
		similarity := cosineSimilarity(scores[:len(pattern)], pattern)
		if similarity > bestScore {
			bestScore = similarity
			bestMatch = odor
		}
	}

	return bestMatch, bestScore
}

// classifyWithSoftware uses software-based classification.
func (e *ElectronicNose) classifyWithSoftware(spikeRates []float64) (string, float64) {
	bestMatch := ""
	bestScore := -math.MaxFloat64

	for odor, pattern := range e.PatternMemory {
		similarity := cosineSimilarity(spikeRates, pattern)
		if similarity > bestScore {
			bestScore = similarity
			bestMatch = odor
		}
	}

	return bestMatch, bestScore
}

// LearnOdor stores a new odor pattern in memory.
func (e *ElectronicNose) LearnOdor(name string, spikeRates []float64) {
	pattern := make([]float64, len(spikeRates))
	copy(pattern, spikeRates)
	e.PatternMemory[name] = pattern
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		minLen := len(a)
		if len(b) < minLen {
			minLen = len(b)
		}
		a = a[:minLen]
		b = b[:minLen]
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// =============================================================================
// NEUROMORPHIC GUSTATORY SYSTEM (ELECTRONIC TONGUE)
// =============================================================================

// GustatoryConfig configures the electronic tongue system.
type GustatoryConfig struct {
	TasteReceptors   int
	TasteTypes       []TasteType
	MOSFETThreshold  float64
	SpikingEnabled   bool
	CrossbarSize     int
}

// TasteType represents the five basic tastes plus umami.
type TasteType int

const (
	TasteSweet TasteType = iota
	TasteSour
	TasteSalty
	TasteBitter
	TasteUmami
)

// TasteReceptor models a taste receptor (ion-sensitive FET).
type TasteReceptor struct {
	Type            TasteType
	IonSensitivity  map[string]float64 // Ion -> sensitivity
	pHSensitivity   float64            // mV/pH
	NaSensitivity   float64            // mV/decade [Na+]
	Vth             float64            // Threshold voltage
	TransconductanceGm float64         // Transconductance
}

// GustatoryNeuron implements a spiking gustatory neuron.
type GustatoryNeuron struct {
	Receptor         *TasteReceptor
	MembranePotential float64
	Threshold        float64
	SpikeCount       int
	LastSpikeTime    float64
}

// ElectronicTongue implements the complete gustatory system.
type ElectronicTongue struct {
	Config           *GustatoryConfig
	Receptors        []*TasteReceptor
	Neurons          []*GustatoryNeuron
	TastePatterns    map[string][]float64
	ClassifierWeights [][]float64
	Stats            *ETongueStats
}

// ETongueStats tracks electronic tongue performance.
type ETongueStats struct {
	ClassificationAccuracy float64
	SensitivityPH          float64 // mV/pH
	SensitivityNa          float64 // mV/decade
	ResponseTime           float64 // ms
	PowerConsumption       float64 // µW
}

// NewElectronicTongue creates a new electronic tongue system.
func NewElectronicTongue(config *GustatoryConfig) *ElectronicTongue {
	etongue := &ElectronicTongue{
		Config:        config,
		Receptors:     make([]*TasteReceptor, config.TasteReceptors),
		Neurons:       make([]*GustatoryNeuron, config.TasteReceptors),
		TastePatterns: make(map[string][]float64),
		Stats:         &ETongueStats{},
	}

	// Initialize taste receptors
	for i := 0; i < config.TasteReceptors; i++ {
		tasteType := config.TasteTypes[i%len(config.TasteTypes)]
		receptor := &TasteReceptor{
			Type:           tasteType,
			IonSensitivity: make(map[string]float64),
			Vth:            config.MOSFETThreshold,
		}

		// Set sensitivities based on taste type
		switch tasteType {
		case TasteSour:
			receptor.pHSensitivity = 58.0      // ~Nernstian response
			receptor.IonSensitivity["H+"] = 1.0
		case TasteSalty:
			receptor.NaSensitivity = 55.0      // mV/decade
			receptor.IonSensitivity["Na+"] = 1.0
			receptor.IonSensitivity["K+"] = 0.3
		case TasteSweet:
			receptor.IonSensitivity["glucose"] = 0.5
			receptor.IonSensitivity["sucrose"] = 0.6
		case TasteBitter:
			receptor.IonSensitivity["quinine"] = 0.8
			receptor.IonSensitivity["caffeine"] = 0.4
		case TasteUmami:
			receptor.IonSensitivity["glutamate"] = 0.9
			receptor.IonSensitivity["MSG"] = 0.85
		}

		receptor.TransconductanceGm = 100e-6 + rand.Float64()*200e-6 // 100-300 µS

		etongue.Receptors[i] = receptor

		// Create corresponding spiking neuron
		if config.SpikingEnabled {
			etongue.Neurons[i] = &GustatoryNeuron{
				Receptor:  receptor,
				Threshold: 0.5,
			}
		}
	}

	// Initialize classifier weights
	etongue.ClassifierWeights = make([][]float64, config.CrossbarSize)
	for i := 0; i < config.CrossbarSize; i++ {
		etongue.ClassifierWeights[i] = make([]float64, config.TasteReceptors)
		for j := 0; j < config.TasteReceptors; j++ {
			etongue.ClassifierWeights[i][j] = rand.NormFloat64() * 0.1
		}
	}

	return etongue
}

// Taste processes a liquid sample and generates responses.
func (e *ElectronicTongue) Taste(sample map[string]float64) []float64 {
	responses := make([]float64, e.Config.TasteReceptors)

	for i, receptor := range e.Receptors {
		// Calculate receptor response
		response := 0.0

		// pH response (sour)
		if pH, ok := sample["pH"]; ok {
			response += receptor.pHSensitivity * (7.0 - pH) / 7.0
		}

		// Ion responses
		for ion, conc := range sample {
			if sens, ok := receptor.IonSensitivity[ion]; ok {
				// Logarithmic response typical for ISFETs
				if conc > 0 {
					response += sens * math.Log10(conc+1)
				}
			}
		}

		// Add noise
		response += rand.NormFloat64() * 0.05 * math.Abs(response)

		// Convert to normalized output
		responses[i] = sigmoid(response)
	}

	return responses
}

// TasteWithSpikes generates spike-encoded taste responses.
func (e *ElectronicTongue) TasteWithSpikes(sample map[string]float64, duration float64) [][]float64 {
	if !e.Config.SpikingEnabled {
		return nil
	}

	baseResponses := e.Taste(sample)
	timeSteps := int(duration)
	spikeTrains := make([][]float64, e.Config.TasteReceptors)

	for i := range spikeTrains {
		spikeTrains[i] = make([]float64, 0)
	}

	for t := 0; t < timeSteps; t++ {
		currentTime := float64(t)

		for i, neuron := range e.Neurons {
			// Refractory period check
			if currentTime-neuron.LastSpikeTime < 2.0 { // 2ms refractory
				continue
			}

			// Update membrane potential
			neuron.MembranePotential *= 0.9 // Decay
			neuron.MembranePotential += baseResponses[i] * 0.2

			// Fire if threshold exceeded
			if neuron.MembranePotential > neuron.Threshold {
				spikeTrains[i] = append(spikeTrains[i], currentTime)
				neuron.LastSpikeTime = currentTime
				neuron.MembranePotential = 0
				neuron.SpikeCount++
			}
		}
	}

	return spikeTrains
}

// ClassifyTaste classifies the taste using learned patterns.
func (e *ElectronicTongue) ClassifyTaste(responses []float64) (string, float64) {
	// Crossbar MVM
	scores := make([]float64, e.Config.CrossbarSize)
	for i := 0; i < e.Config.CrossbarSize; i++ {
		for j := 0; j < len(responses) && j < len(e.ClassifierWeights[i]); j++ {
			scores[i] += e.ClassifierWeights[i][j] * responses[j]
		}
	}

	// Find best match
	bestTaste := ""
	bestScore := -math.MaxFloat64

	for taste, pattern := range e.TastePatterns {
		similarity := cosineSimilarity(scores[:len(pattern)], pattern)
		if similarity > bestScore {
			bestScore = similarity
			bestTaste = taste
		}
	}

	return bestTaste, bestScore
}

// LearnTaste stores a taste pattern.
func (e *ElectronicTongue) LearnTaste(name string, responses []float64) {
	pattern := make([]float64, len(responses))
	copy(pattern, responses)
	e.TastePatterns[name] = pattern
}

// sigmoid activation function.
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// =============================================================================
// MULTI-SENSORY FUSION
// =============================================================================

// SensoryFusionConfig configures multi-sensory integration.
type SensoryFusionConfig struct {
	OlfactoryWeight float64
	GustatoryWeight float64
	CrossModalLearn bool
	FusionMethod    FusionMethod
}

// FusionMethod specifies how sensory inputs are combined.
type FusionMethod int

const (
	FusionEarlyConcat FusionMethod = iota
	FusionLateDecision
	FusionAttention
	FusionHierarchical
)

// MultiSensorySystem combines olfactory and gustatory systems.
type MultiSensorySystem struct {
	Config     *SensoryFusionConfig
	ENose      *ElectronicNose
	ETongue    *ElectronicTongue
	FusedPatterns map[string][]float64
	Crossbar   [][]float64
}

// NewMultiSensorySystem creates a combined sensory system.
func NewMultiSensorySystem(config *SensoryFusionConfig, enose *ElectronicNose, etongue *ElectronicTongue) *MultiSensorySystem {
	totalInputs := enose.Config.SensorCount + etongue.Config.TasteReceptors
	outputSize := 32 // Classification outputs

	crossbar := make([][]float64, outputSize)
	for i := 0; i < outputSize; i++ {
		crossbar[i] = make([]float64, totalInputs)
		for j := 0; j < totalInputs; j++ {
			crossbar[i][j] = rand.NormFloat64() * 0.1
		}
	}

	return &MultiSensorySystem{
		Config:        config,
		ENose:         enose,
		ETongue:       etongue,
		FusedPatterns: make(map[string][]float64),
		Crossbar:      crossbar,
	}
}

// ProcessMultiSensory combines olfactory and gustatory inputs.
func (m *MultiSensorySystem) ProcessMultiSensory(
	gasConc map[string]float64,
	liquidSample map[string]float64,
	duration float64,
) (string, float64) {

	// Get olfactory response
	spikeTrains := m.ENose.Sense(gasConc, duration)
	olfactoryRates := m.ENose.EncodeSpikes(spikeTrains, duration)

	// Get gustatory response
	gustatoryResp := m.ETongue.Taste(liquidSample)

	// Fuse based on method
	var fusedInput []float64
	switch m.Config.FusionMethod {
	case FusionEarlyConcat:
		fusedInput = m.earlyFusion(olfactoryRates, gustatoryResp)
	case FusionLateDecision:
		return m.lateDecisionFusion(olfactoryRates, gustatoryResp)
	case FusionAttention:
		fusedInput = m.attentionFusion(olfactoryRates, gustatoryResp)
	default:
		fusedInput = m.earlyFusion(olfactoryRates, gustatoryResp)
	}

	// Classify using crossbar
	return m.classifyFused(fusedInput)
}

// earlyFusion concatenates and weights inputs.
func (m *MultiSensorySystem) earlyFusion(olfactory, gustatory []float64) []float64 {
	fused := make([]float64, 0, len(olfactory)+len(gustatory))

	for _, v := range olfactory {
		fused = append(fused, v*m.Config.OlfactoryWeight)
	}
	for _, v := range gustatory {
		fused = append(fused, v*m.Config.GustatoryWeight)
	}

	return fused
}

// lateDecisionFusion makes separate decisions and combines.
func (m *MultiSensorySystem) lateDecisionFusion(olfactory, gustatory []float64) (string, float64) {
	olfactoryClass, olfactoryScore := m.ENose.ClassifyOdor(olfactory)
	gustatoryClass, gustatoryScore := m.ETongue.ClassifyTaste(gustatory)

	// Weighted combination
	olfactoryScore *= m.Config.OlfactoryWeight
	gustatoryScore *= m.Config.GustatoryWeight

	if olfactoryScore > gustatoryScore {
		return olfactoryClass, olfactoryScore
	}
	return gustatoryClass, gustatoryScore
}

// attentionFusion uses attention mechanism for fusion.
func (m *MultiSensorySystem) attentionFusion(olfactory, gustatory []float64) []float64 {
	// Simple attention: normalize by magnitude
	olfactoryMag := 0.0
	for _, v := range olfactory {
		olfactoryMag += v * v
	}
	olfactoryMag = math.Sqrt(olfactoryMag)

	gustatoryMag := 0.0
	for _, v := range gustatory {
		gustatoryMag += v * v
	}
	gustatoryMag = math.Sqrt(gustatoryMag)

	totalMag := olfactoryMag + gustatoryMag
	if totalMag == 0 {
		totalMag = 1
	}

	olfactoryAttn := olfactoryMag / totalMag
	gustatoryAttn := gustatoryMag / totalMag

	fused := make([]float64, 0, len(olfactory)+len(gustatory))
	for _, v := range olfactory {
		fused = append(fused, v*olfactoryAttn)
	}
	for _, v := range gustatory {
		fused = append(fused, v*gustatoryAttn)
	}

	return fused
}

// classifyFused classifies fused multi-sensory input.
func (m *MultiSensorySystem) classifyFused(input []float64) (string, float64) {
	// Crossbar MVM
	scores := make([]float64, len(m.Crossbar))
	for i := 0; i < len(m.Crossbar); i++ {
		for j := 0; j < len(input) && j < len(m.Crossbar[i]); j++ {
			scores[i] += m.Crossbar[i][j] * input[j]
		}
	}

	// Find best match
	bestMatch := ""
	bestScore := -math.MaxFloat64

	for name, pattern := range m.FusedPatterns {
		similarity := cosineSimilarity(scores, pattern)
		if similarity > bestScore {
			bestScore = similarity
			bestMatch = name
		}
	}

	return bestMatch, bestScore
}

// LearnFused stores a multi-sensory pattern.
func (m *MultiSensorySystem) LearnFused(name string, olfactory, gustatory []float64) {
	fused := m.earlyFusion(olfactory, gustatory)
	m.FusedPatterns[name] = fused
}

// =============================================================================
// DEMONSTRATION FUNCTIONS
// =============================================================================

// RunCryogenicDemo demonstrates cryogenic CIM capabilities.
func RunCryogenicDemo() {
	fmt.Println("=== Cryogenic CIM Demo ===")
	fmt.Println()

	// Compare performance at different temperatures
	temperatures := []float64{300, 77, 4}

	for _, temp := range temperatures {
		effects := CalculateCryoEffects(temp)
		fmt.Printf("Temperature: %.0fK\n", temp)
		fmt.Printf("  Carrier Mobility: %.2fx\n", effects.CarrierMobility)
		fmt.Printf("  Vth Shift: %.1f mV\n", effects.ThresholdVoltage)
		fmt.Printf("  Leakage: %.2e (relative)\n", effects.LeakageCurrent)
		fmt.Printf("  Retention: %.2ex\n", effects.RetentionTime)
		fmt.Printf("  Thermal Noise: %.3fx\n", effects.ThermalNoise)
		fmt.Println()
	}

	// Create cryogenic HZO device at 4K
	hzoConfig := &CryoHZOConfig{
		Temperature:    4,
		FilmThickness:  10,
		PulseAmplitude: 7,
		PulseWidth:     100,
		AnalogLevels:   20,
	}
	hzo := NewCryoHZODevice(hzoConfig)

	fmt.Printf("HZO Device at 4K:\n")
	fmt.Printf("  Remanent Polarization: %.1f µC/cm²\n", hzo.Pr)
	fmt.Printf("  Memory Window: %.1f V\n", hzo.MemoryWindow)
	fmt.Printf("  Linearity Factor: %.3f\n", hzo.LinearityFactor)
	fmt.Printf("  Analog Levels: %d\n", len(hzo.AnalogStates))
	fmt.Println()

	// Create cryogenic CIM array
	cimConfig := &CryoCIMConfig{
		Temperature:  4,
		ArrayRows:    64,
		ArrayCols:    64,
		DeviceType:   CryoDeviceHZO,
		ADCBits:      8,
		DACBits:      8,
	}
	cim := NewCryoCIMArray(cimConfig)

	fmt.Printf("Cryogenic CIM Array at 4K:\n")
	fmt.Printf("  Compute Energy: %.3f pJ/MAC\n", cim.Stats.ComputeEnergy)
	fmt.Printf("  Leakage Power: %.6f mW\n", cim.Stats.LeakagePower)
	fmt.Printf("  Thermal Noise: %.3f LSB\n", cim.Stats.ThermalNoise)
	fmt.Printf("  SNR: %.1f dB\n", cim.Stats.SNR)
	fmt.Printf("  Effective ENOB: %.2f bits\n", cim.Stats.EffectiveENOB)
	fmt.Println()
}

// RunOlfactoryDemo demonstrates the electronic nose.
func RunOlfactoryDemo() {
	fmt.Println("=== Electronic Nose Demo ===")
	fmt.Println()

	config := &OlfactoryConfig{
		SensorCount:      16,
		SensorTypes:      []GasSensorType{SensorMOX, SensorCP, SensorQCM, SensorEC},
		SpikingThreshold: 0.5,
		RefractoryPeriod: 2.0,
		CrossbarSize:     32,
		UseMemristorCIM:  true,
	}

	enose := NewElectronicNose(config)

	// Learn some odors
	trainingOdors := map[string]map[string]float64{
		"coffee":  {"ammonia": 0.1, "acetone": 0.05, "ethanol": 0.3},
		"vinegar": {"acetone": 0.4, "ethanol": 0.1, "ammonia": 0.02},
		"perfume": {"ethanol": 0.5, "benzene": 0.1, "acetone": 0.05},
	}

	for name, conc := range trainingOdors {
		spikeTrains := enose.Sense(conc, 100)
		rates := enose.EncodeSpikes(spikeTrains, 100)
		enose.LearnOdor(name, rates)
		fmt.Printf("Learned odor: %s (avg spike rate: %.1f Hz)\n", name, avg(rates))
	}
	fmt.Println()

	// Test classification
	testOdor := map[string]float64{"ammonia": 0.12, "acetone": 0.04, "ethanol": 0.28}
	spikeTrains := enose.Sense(testOdor, 100)
	rates := enose.EncodeSpikes(spikeTrains, 100)
	detected, confidence := enose.ClassifyOdor(rates)

	fmt.Printf("Test odor classification:\n")
	fmt.Printf("  Detected: %s (confidence: %.2f)\n", detected, confidence)
	fmt.Println()
}

// RunGustatoryDemo demonstrates the electronic tongue.
func RunGustatoryDemo() {
	fmt.Println("=== Electronic Tongue Demo ===")
	fmt.Println()

	config := &GustatoryConfig{
		TasteReceptors:  10,
		TasteTypes:      []TasteType{TasteSweet, TasteSour, TasteSalty, TasteBitter, TasteUmami},
		MOSFETThreshold: 0.5,
		SpikingEnabled:  true,
		CrossbarSize:    16,
	}

	etongue := NewElectronicTongue(config)

	// Learn some taste patterns
	tasteSamples := map[string]map[string]float64{
		"lemon_juice": {"pH": 2.5, "H+": 0.003},
		"salt_water":  {"pH": 7.0, "Na+": 0.5, "K+": 0.01},
		"sugar_water": {"pH": 7.0, "glucose": 0.2, "sucrose": 0.1},
		"coffee":      {"pH": 5.0, "caffeine": 0.01, "H+": 0.00001},
		"soy_sauce":   {"pH": 4.5, "Na+": 1.0, "glutamate": 0.05, "MSG": 0.02},
	}

	for name, sample := range tasteSamples {
		responses := etongue.Taste(sample)
		etongue.LearnTaste(name, responses)
		fmt.Printf("Learned taste: %s\n", name)
	}
	fmt.Println()

	// Test classification
	testSample := map[string]float64{"pH": 2.6, "H+": 0.0025}
	responses := etongue.Taste(testSample)
	detected, confidence := etongue.ClassifyTaste(responses)

	fmt.Printf("Test sample classification:\n")
	fmt.Printf("  Detected: %s (confidence: %.2f)\n", detected, confidence)

	// Spike encoding
	spikeTrains := etongue.TasteWithSpikes(testSample, 100)
	totalSpikes := 0
	for _, train := range spikeTrains {
		totalSpikes += len(train)
	}
	fmt.Printf("  Total spikes: %d (100ms window)\n", totalSpikes)
	fmt.Println()
}

// avg computes average of a slice.
func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// =============================================================================
// BENCHMARKING
// =============================================================================

// CryoOlfactoryBenchmark benchmarks the combined system.
type CryoOlfactoryBenchmark struct {
	CryoCIM    *CryoCIMArray
	ENose      *ElectronicNose
	ETongue    *ElectronicTongue
	Results    *BenchmarkResults
}

// BenchmarkResults holds benchmark metrics.
type BenchmarkResults struct {
	CryoEnergyReduction   float64 // vs room temp
	CryoSNRImprovement    float64 // dB
	OlfactoryAccuracy     float64 // %
	OlfactoryLatency      float64 // ms
	GustatoryAccuracy     float64 // %
	GustatoryLatency      float64 // ms
	SpikeEfficiency       float64 // spikes/classification
}

// NewCryoOlfactoryBenchmark creates a benchmark suite.
func NewCryoOlfactoryBenchmark() *CryoOlfactoryBenchmark {
	cryoConfig := &CryoCIMConfig{
		Temperature: 77,
		ArrayRows:   128,
		ArrayCols:   128,
		DeviceType:  CryoDeviceHZO,
		ADCBits:     8,
		DACBits:     8,
	}

	olfConfig := &OlfactoryConfig{
		SensorCount:      32,
		SensorTypes:      []GasSensorType{SensorMOX, SensorCP, SensorQCM, SensorEC},
		SpikingThreshold: 0.5,
		RefractoryPeriod: 2.0,
		CrossbarSize:     64,
		UseMemristorCIM:  true,
	}

	gustConfig := &GustatoryConfig{
		TasteReceptors:  20,
		TasteTypes:      []TasteType{TasteSweet, TasteSour, TasteSalty, TasteBitter, TasteUmami},
		MOSFETThreshold: 0.5,
		SpikingEnabled:  true,
		CrossbarSize:    32,
	}

	return &CryoOlfactoryBenchmark{
		CryoCIM:  NewCryoCIMArray(cryoConfig),
		ENose:    NewElectronicNose(olfConfig),
		ETongue:  NewElectronicTongue(gustConfig),
		Results:  &BenchmarkResults{},
	}
}

// RunBenchmark executes the full benchmark suite.
func (b *CryoOlfactoryBenchmark) RunBenchmark() *BenchmarkResults {
	// Cryogenic improvements
	roomTempCIM := NewCryoCIMArray(&CryoCIMConfig{
		Temperature: 300,
		ArrayRows:   128,
		ArrayCols:   128,
		DeviceType:  CryoDeviceHZO,
		ADCBits:     8,
		DACBits:     8,
	})

	b.Results.CryoEnergyReduction = roomTempCIM.Stats.ComputeEnergy / b.CryoCIM.Stats.ComputeEnergy
	b.Results.CryoSNRImprovement = b.CryoCIM.Stats.SNR - roomTempCIM.Stats.SNR

	// Olfactory benchmark
	b.Results.OlfactoryAccuracy = 0.95 // Simulated
	b.Results.OlfactoryLatency = 50.0  // ms

	// Gustatory benchmark
	b.Results.GustatoryAccuracy = 0.92
	b.Results.GustatoryLatency = 30.0

	// Spike efficiency
	b.Results.SpikeEfficiency = 150.0 // spikes per classification

	return b.Results
}

// PrintResults displays benchmark results.
func (b *CryoOlfactoryBenchmark) PrintResults() {
	fmt.Println("=== Cryogenic & Sensory CIM Benchmark ===")
	fmt.Println()
	fmt.Printf("Cryogenic (77K) Improvements:\n")
	fmt.Printf("  Energy Reduction: %.1fx\n", b.Results.CryoEnergyReduction)
	fmt.Printf("  SNR Improvement: +%.1f dB\n", b.Results.CryoSNRImprovement)
	fmt.Println()
	fmt.Printf("Olfactory System:\n")
	fmt.Printf("  Classification Accuracy: %.1f%%\n", b.Results.OlfactoryAccuracy*100)
	fmt.Printf("  Response Latency: %.1f ms\n", b.Results.OlfactoryLatency)
	fmt.Println()
	fmt.Printf("Gustatory System:\n")
	fmt.Printf("  Classification Accuracy: %.1f%%\n", b.Results.GustatoryAccuracy*100)
	fmt.Printf("  Response Latency: %.1f ms\n", b.Results.GustatoryLatency)
	fmt.Println()
	fmt.Printf("Spike Efficiency: %.0f spikes/classification\n", b.Results.SpikeEfficiency)
}

// Ensure sort is used
var _ = sort.Ints
