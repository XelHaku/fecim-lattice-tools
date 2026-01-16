// Package layers provides neural network layer implementations for CIM deployment.
// This file implements neuromorphic olfaction and neural architecture search for CIM.
// Based on research: Nature Reviews EE (olfaction chips), CIMNAS (arXiv 2025),
// NeuroNAS (hardware-aware SNN NAS)
package layers

import (
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// NEUROMORPHIC OLFACTION SYSTEM
// Bio-inspired electronic nose with memristive synapses
// =============================================================================

// OlfactoryReceptorType defines types of olfactory receptors.
type OlfactoryReceptorType int

const (
	OlfactoryReceptorMOX    OlfactoryReceptorType = iota // Metal oxide (SnO2, ZnO)
	OlfactoryReceptorCP                                   // Conducting polymer
	OlfactoryReceptorCNT                                  // Carbon nanotube
	OlfactoryReceptorMXene                                // MXene-based
	OlfactoryReceptorFeFET                                // Ferroelectric FET
)

// GasSensorConfig configures a single gas sensor.
type GasSensorConfig struct {
	Type            OlfactoryReceptorType
	TargetGas       string   // Primary target gas
	Sensitivity     float64  // Sensitivity factor
	CrossSensitivity []float64 // Response to other gases

	// Temporal dynamics
	ResponseTimeMs  float64
	RecoveryTimeMs  float64

	// Operating conditions
	TemperatureC    float64
	HeaterPowerMW   float64

	// Drift parameters
	DriftRate       float64 // Baseline drift per hour
}

// DefaultGasSensorConfigs returns common sensor configurations.
func DefaultGasSensorConfigs() []*GasSensorConfig {
	return []*GasSensorConfig{
		{Type: OlfactoryReceptorMOX, TargetGas: "NH3", Sensitivity: 1.0, ResponseTimeMs: 100, RecoveryTimeMs: 500},
		{Type: OlfactoryReceptorMOX, TargetGas: "CO", Sensitivity: 0.8, ResponseTimeMs: 150, RecoveryTimeMs: 600},
		{Type: OlfactoryReceptorMOX, TargetGas: "NO2", Sensitivity: 1.2, ResponseTimeMs: 80, RecoveryTimeMs: 400},
		{Type: OlfactoryReceptorMOX, TargetGas: "H2S", Sensitivity: 1.5, ResponseTimeMs: 50, RecoveryTimeMs: 300},
		{Type: OlfactoryReceptorCP, TargetGas: "VOC", Sensitivity: 0.9, ResponseTimeMs: 200, RecoveryTimeMs: 800},
		{Type: OlfactoryReceptorCNT, TargetGas: "CO2", Sensitivity: 0.7, ResponseTimeMs: 120, RecoveryTimeMs: 550},
		{Type: OlfactoryReceptorMXene, TargetGas: "Ethanol", Sensitivity: 1.1, ResponseTimeMs: 90, RecoveryTimeMs: 450},
		{Type: OlfactoryReceptorFeFET, TargetGas: "Acetone", Sensitivity: 1.3, ResponseTimeMs: 60, RecoveryTimeMs: 350},
	}
}

// GasSensor implements a single gas sensing element.
type GasSensor struct {
	Config *GasSensorConfig

	// State
	Resistance     float64 // Current resistance
	BaseResistance float64 // Baseline resistance
	Response       float64 // Normalized response (0-1)

	// Temporal state
	LastUpdateTime float64
	IntegratedResp float64

	// Drift tracking
	DriftAccum     float64
	OperatingHours float64
}

// NewGasSensor creates a new gas sensor.
func NewGasSensor(config *GasSensorConfig) *GasSensor {
	if config == nil {
		config = DefaultGasSensorConfigs()[0]
	}
	return &GasSensor{
		Config:         config,
		BaseResistance: 10000.0, // 10 kOhm baseline
		Resistance:     10000.0,
	}
}

// Sense processes gas concentration input.
func (s *GasSensor) Sense(concentration float64, timeMs float64) float64 {
	dt := timeMs - s.LastUpdateTime
	if dt <= 0 {
		dt = 1
	}

	// Response dynamics (first-order)
	targetResponse := concentration * s.Config.Sensitivity
	if targetResponse > 1 {
		targetResponse = 1
	}

	// Rise or fall time constant
	var tau float64
	if targetResponse > s.Response {
		tau = s.Config.ResponseTimeMs
	} else {
		tau = s.Config.RecoveryTimeMs
	}

	alpha := 1.0 - math.Exp(-dt/tau)
	s.Response = s.Response + alpha*(targetResponse-s.Response)

	// Update resistance (assuming reducing gas)
	s.Resistance = s.BaseResistance / (1.0 + s.Response*10.0)

	// Apply drift
	s.OperatingHours += dt / 3600000.0
	s.DriftAccum += s.Config.DriftRate * dt / 3600000.0
	s.BaseResistance *= (1.0 + s.DriftAccum*0.01)

	s.LastUpdateTime = timeMs
	return s.Response
}

// =============================================================================
// NEUROMORPHIC OLFACTORY BULB
// =============================================================================

// OlfactoryBulbConfig configures the olfactory bulb model.
type OlfactoryBulbConfig struct {
	NumGlomeruli    int     // Number of glomerular units
	NumMitralCells  int     // Output neurons per glomerulus
	LateralInhibK   float64 // Lateral inhibition strength
	AdaptationTau   float64 // Adaptation time constant

	// Memristor synapse parameters
	UseMemristorSynapses bool
	SynapseOnOff         float64
	SynapseVariation     float64
}

// OlfactoryBulb implements the olfactory bulb processing.
type OlfactoryBulb struct {
	Config *OlfactoryBulbConfig

	// Glomerular layer (sensor input aggregation)
	GlomeruliActivation []float64

	// Mitral cell layer (output neurons)
	MitralActivation    []float64
	MitralThresholds    []float64

	// Lateral inhibition weights (granule cells)
	LateralWeights      [][]float64

	// Adaptation state
	AdaptationLevel     []float64
}

// NewOlfactoryBulb creates a new olfactory bulb.
func NewOlfactoryBulb(config *OlfactoryBulbConfig) *OlfactoryBulb {
	if config == nil {
		config = &OlfactoryBulbConfig{
			NumGlomeruli:         32,
			NumMitralCells:       4,
			LateralInhibK:        0.2,
			AdaptationTau:        1000.0,
			UseMemristorSynapses: true,
			SynapseOnOff:         1e4,
			SynapseVariation:     0.02,
		}
	}

	numMitral := config.NumGlomeruli * config.NumMitralCells

	// Initialize lateral inhibition weights
	lateralWeights := make([][]float64, numMitral)
	for i := range lateralWeights {
		lateralWeights[i] = make([]float64, numMitral)
		for j := range lateralWeights[i] {
			if i != j {
				// Distance-dependent inhibition
				dist := math.Abs(float64(i-j)) / float64(numMitral)
				lateralWeights[i][j] = config.LateralInhibK * math.Exp(-dist*5.0)
			}
		}
	}

	return &OlfactoryBulb{
		Config:              config,
		GlomeruliActivation: make([]float64, config.NumGlomeruli),
		MitralActivation:    make([]float64, numMitral),
		MitralThresholds:    make([]float64, numMitral),
		LateralWeights:      lateralWeights,
		AdaptationLevel:     make([]float64, numMitral),
	}
}

// Process processes sensor inputs through olfactory bulb.
func (ob *OlfactoryBulb) Process(sensorResponses []float64, dt float64) []float64 {
	// Map sensors to glomeruli
	numSensors := len(sensorResponses)
	sensorsPerGlom := numSensors / ob.Config.NumGlomeruli
	if sensorsPerGlom < 1 {
		sensorsPerGlom = 1
	}

	// Glomerular aggregation
	for g := 0; g < ob.Config.NumGlomeruli; g++ {
		sum := 0.0
		count := 0
		for s := g * sensorsPerGlom; s < (g+1)*sensorsPerGlom && s < numSensors; s++ {
			sum += sensorResponses[s]
			count++
		}
		if count > 0 {
			ob.GlomeruliActivation[g] = sum / float64(count)
		}
	}

	// Mitral cell activation with lateral inhibition
	numMitral := ob.Config.NumGlomeruli * ob.Config.NumMitralCells
	preInhibition := make([]float64, numMitral)

	for m := 0; m < numMitral; m++ {
		glom := m / ob.Config.NumMitralCells
		preInhibition[m] = ob.GlomeruliActivation[glom]

		// Add variation for mitral cell diversity
		preInhibition[m] *= (1.0 + 0.1*rand.NormFloat64())
	}

	// Apply lateral inhibition
	for m := 0; m < numMitral; m++ {
		inhibition := 0.0
		for n := 0; n < numMitral; n++ {
			inhibition += ob.LateralWeights[m][n] * preInhibition[n]
		}

		// Subtract inhibition
		ob.MitralActivation[m] = preInhibition[m] - inhibition

		// Apply adaptation
		adaptDt := dt / ob.Config.AdaptationTau
		ob.AdaptationLevel[m] = (1-adaptDt)*ob.AdaptationLevel[m] + adaptDt*ob.MitralActivation[m]
		ob.MitralActivation[m] -= ob.AdaptationLevel[m] * 0.5

		// Rectify
		if ob.MitralActivation[m] < 0 {
			ob.MitralActivation[m] = 0
		}
	}

	return ob.MitralActivation
}

// =============================================================================
// ELECTRONIC NOSE SYSTEM
// =============================================================================

// ElectronicNoseConfig configures the complete e-nose system.
type ElectronicNoseConfig struct {
	NumSensors      int
	SensorConfigs   []*GasSensorConfig
	BulbConfig      *OlfactoryBulbConfig

	// Classification
	NumOdorClasses  int
	UseSNN          bool
}

// ElectronicNose implements a complete neuromorphic e-nose.
type ElectronicNose struct {
	Config  *ElectronicNoseConfig
	Sensors []*GasSensor
	Bulb    *OlfactoryBulb

	// Classification layer
	ClassWeights    [][]float64
	ClassBiases     []float64

	// Training state
	TrainedOdors    []string
}

// NewElectronicNose creates a new electronic nose system.
func NewElectronicNose(config *ElectronicNoseConfig) *ElectronicNose {
	if config == nil {
		sensorConfigs := DefaultGasSensorConfigs()
		config = &ElectronicNoseConfig{
			NumSensors:     len(sensorConfigs),
			SensorConfigs:  sensorConfigs,
			NumOdorClasses: 10,
			UseSNN:         true,
		}
	}

	// Create sensors
	sensors := make([]*GasSensor, config.NumSensors)
	for i := 0; i < config.NumSensors; i++ {
		if i < len(config.SensorConfigs) {
			sensors[i] = NewGasSensor(config.SensorConfigs[i])
		} else {
			sensors[i] = NewGasSensor(nil)
		}
	}

	// Create bulb
	bulb := NewOlfactoryBulb(config.BulbConfig)

	// Initialize classification weights
	inputSize := bulb.Config.NumGlomeruli * bulb.Config.NumMitralCells
	classWeights := make([][]float64, inputSize)
	for i := range classWeights {
		classWeights[i] = make([]float64, config.NumOdorClasses)
		scale := math.Sqrt(2.0 / float64(inputSize))
		for j := range classWeights[i] {
			classWeights[i][j] = rand.NormFloat64() * scale
		}
	}

	return &ElectronicNose{
		Config:       config,
		Sensors:      sensors,
		Bulb:         bulb,
		ClassWeights: classWeights,
		ClassBiases:  make([]float64, config.NumOdorClasses),
		TrainedOdors: make([]string, 0),
	}
}

// Sense processes gas concentrations through the e-nose.
func (en *ElectronicNose) Sense(concentrations []float64, timeMs float64) *ENoseOutput {
	// Sensor layer
	sensorResponses := make([]float64, len(en.Sensors))
	for i, sensor := range en.Sensors {
		conc := 0.0
		if i < len(concentrations) {
			conc = concentrations[i]
		}
		sensorResponses[i] = sensor.Sense(conc, timeMs)
	}

	// Olfactory bulb processing
	dt := 10.0 // Assume 10ms timestep
	bulbOutput := en.Bulb.Process(sensorResponses, dt)

	// Classification
	classScores := make([]float64, en.Config.NumOdorClasses)
	for j := 0; j < en.Config.NumOdorClasses; j++ {
		sum := en.ClassBiases[j]
		for i := 0; i < len(bulbOutput) && i < len(en.ClassWeights); i++ {
			sum += bulbOutput[i] * en.ClassWeights[i][j]
		}
		classScores[j] = sum
	}

	// Softmax
	maxScore := classScores[0]
	for _, s := range classScores {
		if s > maxScore {
			maxScore = s
		}
	}
	sumExp := 0.0
	for i := range classScores {
		classScores[i] = math.Exp(classScores[i] - maxScore)
		sumExp += classScores[i]
	}
	for i := range classScores {
		classScores[i] /= sumExp
	}

	// Find predicted class
	predClass := 0
	predConf := classScores[0]
	for i, score := range classScores {
		if score > predConf {
			predConf = score
			predClass = i
		}
	}

	return &ENoseOutput{
		SensorResponses: sensorResponses,
		BulbActivation:  bulbOutput,
		ClassScores:     classScores,
		PredictedClass:  predClass,
		Confidence:      predConf,
	}
}

// ENoseOutput contains e-nose processing output.
type ENoseOutput struct {
	SensorResponses []float64
	BulbActivation  []float64
	ClassScores     []float64
	PredictedClass  int
	Confidence      float64
}

// =============================================================================
// NEURAL ARCHITECTURE SEARCH FOR CIM
// Hardware-aware NAS based on CIMNAS and NeuroNAS
// =============================================================================

// CIMNASConfig configures CIM-aware neural architecture search.
type CIMNASConfig struct {
	// Search space
	MinLayers       int
	MaxLayers       int
	LayerSizeOptions []int
	ActivationOptions []string

	// Hardware constraints
	MaxCrossbarSize int
	MaxTotalParams  int64
	TargetLatencyUs float64
	TargetEnergyPJ  float64
	TargetAreaMM2   float64

	// Quantization options
	WeightBitOptions []int
	ActBitOptions    []int

	// Search parameters
	PopulationSize  int
	NumGenerations  int
	MutationRate    float64
	CrossoverRate   float64
}

// DefaultCIMNASConfig returns default NAS configuration.
func DefaultCIMNASConfig() *CIMNASConfig {
	return &CIMNASConfig{
		MinLayers:        2,
		MaxLayers:        6,
		LayerSizeOptions: []int{32, 64, 128, 256, 512},
		ActivationOptions: []string{"relu", "sigmoid", "tanh"},
		MaxCrossbarSize:  256,
		MaxTotalParams:   500000,
		TargetLatencyUs:  100.0,
		TargetEnergyPJ:   1000.0,
		TargetAreaMM2:    1.0,
		WeightBitOptions: []int{2, 4, 6, 8},
		ActBitOptions:    []int{4, 6, 8},
		PopulationSize:   50,
		NumGenerations:   100,
		MutationRate:     0.1,
		CrossoverRate:    0.7,
	}
}

// NASArchitecture represents a neural architecture candidate.
type NASArchitecture struct {
	LayerSizes    []int
	Activations   []string
	WeightBits    []int
	ActBits       []int

	// Fitness metrics
	Accuracy      float64
	Latency       float64 // µs
	Energy        float64 // pJ
	Area          float64 // mm²
	ParamCount    int64

	// Combined fitness
	Fitness       float64
}

// CIMNASSearch implements hardware-aware NAS for CIM.
type CIMNASSearch struct {
	Config     *CIMNASConfig
	Population []*NASArchitecture
	BestArch   *NASArchitecture

	// Hardware model parameters
	MACLatencyNs   float64
	MACEnergyFJ    float64
	MemCellAreaUM2 float64

	// Search state
	Generation     int
	BestFitness    float64
	SearchHistory  []float64
}

// NewCIMNASSearch creates a new CIM-aware NAS.
func NewCIMNASSearch(config *CIMNASConfig) *CIMNASSearch {
	if config == nil {
		config = DefaultCIMNASConfig()
	}

	return &CIMNASSearch{
		Config:         config,
		Population:     make([]*NASArchitecture, config.PopulationSize),
		MACLatencyNs:   1.0,
		MACEnergyFJ:    10.0,
		MemCellAreaUM2: 0.01,
		SearchHistory:  make([]float64, 0),
	}
}

// InitializePopulation creates initial random architectures.
func (nas *CIMNASSearch) InitializePopulation(inputSize, outputSize int) {
	for i := 0; i < nas.Config.PopulationSize; i++ {
		arch := nas.RandomArchitecture(inputSize, outputSize)
		nas.Population[i] = arch
	}
}

// RandomArchitecture generates a random architecture.
func (nas *CIMNASSearch) RandomArchitecture(inputSize, outputSize int) *NASArchitecture {
	// Random number of layers
	numLayers := nas.Config.MinLayers + rand.Intn(nas.Config.MaxLayers-nas.Config.MinLayers+1)

	layerSizes := make([]int, numLayers+1)
	activations := make([]string, numLayers)
	weightBits := make([]int, numLayers)
	actBits := make([]int, numLayers)

	layerSizes[0] = inputSize
	layerSizes[numLayers] = outputSize

	for l := 1; l < numLayers; l++ {
		layerSizes[l] = nas.Config.LayerSizeOptions[rand.Intn(len(nas.Config.LayerSizeOptions))]
	}

	for l := 0; l < numLayers; l++ {
		activations[l] = nas.Config.ActivationOptions[rand.Intn(len(nas.Config.ActivationOptions))]
		weightBits[l] = nas.Config.WeightBitOptions[rand.Intn(len(nas.Config.WeightBitOptions))]
		actBits[l] = nas.Config.ActBitOptions[rand.Intn(len(nas.Config.ActBitOptions))]
	}

	return &NASArchitecture{
		LayerSizes:  layerSizes,
		Activations: activations,
		WeightBits:  weightBits,
		ActBits:     actBits,
	}
}

// EvaluateArchitecture computes hardware metrics for an architecture.
func (nas *CIMNASSearch) EvaluateArchitecture(arch *NASArchitecture) {
	// Count parameters
	var totalParams int64
	var totalMACs int64

	for l := 0; l < len(arch.LayerSizes)-1; l++ {
		params := int64(arch.LayerSizes[l]) * int64(arch.LayerSizes[l+1])
		totalParams += params
		totalMACs += params // One MAC per weight per input
	}
	arch.ParamCount = totalParams

	// Check if fits in crossbar constraints
	crossbarPenalty := 0.0
	for l := 0; l < len(arch.LayerSizes)-1; l++ {
		if arch.LayerSizes[l] > nas.Config.MaxCrossbarSize ||
			arch.LayerSizes[l+1] > nas.Config.MaxCrossbarSize {
			// Needs tiling
			tiles := math.Ceil(float64(arch.LayerSizes[l])/float64(nas.Config.MaxCrossbarSize)) *
				math.Ceil(float64(arch.LayerSizes[l+1])/float64(nas.Config.MaxCrossbarSize))
			crossbarPenalty += (tiles - 1) * 0.1
		}
	}

	// Latency estimation (ns -> µs)
	arch.Latency = float64(len(arch.LayerSizes)-1) * nas.MACLatencyNs / 1000.0
	// Add DAC/ADC latency per layer
	for l := 0; l < len(arch.LayerSizes)-1; l++ {
		dacLatency := float64(arch.ActBits[l]) * 0.5 // 0.5 ns per bit
		adcLatency := float64(arch.WeightBits[l]) * 1.0 // 1 ns per bit
		arch.Latency += (dacLatency + adcLatency) / 1000.0
	}

	// Energy estimation (fJ -> pJ)
	arch.Energy = float64(totalMACs) * nas.MACEnergyFJ / 1000.0
	// Add DAC/ADC energy
	for l := 0; l < len(arch.LayerSizes)-1; l++ {
		dacEnergy := float64(arch.ActBits[l]) * float64(arch.LayerSizes[l]) * 0.1 // fJ
		adcEnergy := float64(arch.WeightBits[l]) * float64(arch.LayerSizes[l+1]) * 0.5 // fJ
		arch.Energy += (dacEnergy + adcEnergy) / 1000.0
	}

	// Area estimation (µm² -> mm²)
	arch.Area = float64(totalParams) * nas.MemCellAreaUM2 / 1e6

	// Compute fitness (multi-objective)
	latencyScore := 1.0
	if arch.Latency > nas.Config.TargetLatencyUs {
		latencyScore = nas.Config.TargetLatencyUs / arch.Latency
	}

	energyScore := 1.0
	if arch.Energy > nas.Config.TargetEnergyPJ {
		energyScore = nas.Config.TargetEnergyPJ / arch.Energy
	}

	areaScore := 1.0
	if arch.Area > nas.Config.TargetAreaMM2 {
		areaScore = nas.Config.TargetAreaMM2 / arch.Area
	}

	paramScore := 1.0
	if totalParams > nas.Config.MaxTotalParams {
		paramScore = float64(nas.Config.MaxTotalParams) / float64(totalParams)
	}

	// Combined fitness (assuming accuracy is proxy-estimated)
	arch.Accuracy = nas.EstimateAccuracy(arch)
	arch.Fitness = arch.Accuracy * latencyScore * energyScore * areaScore * paramScore *
		(1.0 - crossbarPenalty)
}

// EstimateAccuracy provides a rough accuracy estimate based on architecture.
func (nas *CIMNASSearch) EstimateAccuracy(arch *NASArchitecture) float64 {
	// Heuristic: larger networks tend to have higher accuracy, but diminishing returns
	logParams := math.Log10(float64(arch.ParamCount) + 1)
	baseAccuracy := 0.5 + 0.1*logParams

	// Penalize very low precision
	avgBits := 0.0
	for _, b := range arch.WeightBits {
		avgBits += float64(b)
	}
	avgBits /= float64(len(arch.WeightBits))
	bitPenalty := 1.0 - math.Max(0, (6.0-avgBits)*0.05)

	// Depth bonus (up to a point)
	depthBonus := 1.0 + 0.02*float64(len(arch.LayerSizes)-2)
	if depthBonus > 1.1 {
		depthBonus = 1.1
	}

	accuracy := baseAccuracy * bitPenalty * depthBonus
	if accuracy > 0.99 {
		accuracy = 0.99
	}
	if accuracy < 0.1 {
		accuracy = 0.1
	}

	return accuracy
}

// Mutate applies random mutations to an architecture.
func (nas *CIMNASSearch) Mutate(arch *NASArchitecture) *NASArchitecture {
	newArch := &NASArchitecture{
		LayerSizes:  make([]int, len(arch.LayerSizes)),
		Activations: make([]string, len(arch.Activations)),
		WeightBits:  make([]int, len(arch.WeightBits)),
		ActBits:     make([]int, len(arch.ActBits)),
	}
	copy(newArch.LayerSizes, arch.LayerSizes)
	copy(newArch.Activations, arch.Activations)
	copy(newArch.WeightBits, arch.WeightBits)
	copy(newArch.ActBits, arch.ActBits)

	// Random mutation
	mutationType := rand.Intn(4)
	switch mutationType {
	case 0: // Mutate layer size
		if len(newArch.LayerSizes) > 2 {
			l := 1 + rand.Intn(len(newArch.LayerSizes)-2)
			newArch.LayerSizes[l] = nas.Config.LayerSizeOptions[rand.Intn(len(nas.Config.LayerSizeOptions))]
		}
	case 1: // Mutate weight bits
		if len(newArch.WeightBits) > 0 {
			l := rand.Intn(len(newArch.WeightBits))
			newArch.WeightBits[l] = nas.Config.WeightBitOptions[rand.Intn(len(nas.Config.WeightBitOptions))]
		}
	case 2: // Mutate activation
		if len(newArch.Activations) > 0 {
			l := rand.Intn(len(newArch.Activations))
			newArch.Activations[l] = nas.Config.ActivationOptions[rand.Intn(len(nas.Config.ActivationOptions))]
		}
	case 3: // Add or remove layer
		if rand.Float64() < 0.5 && len(newArch.LayerSizes) > nas.Config.MinLayers+1 {
			// Remove a hidden layer
			idx := 1 + rand.Intn(len(newArch.LayerSizes)-2)
			newArch.LayerSizes = append(newArch.LayerSizes[:idx], newArch.LayerSizes[idx+1:]...)
			newArch.Activations = append(newArch.Activations[:idx], newArch.Activations[idx+1:]...)
			newArch.WeightBits = append(newArch.WeightBits[:idx], newArch.WeightBits[idx+1:]...)
			newArch.ActBits = append(newArch.ActBits[:idx], newArch.ActBits[idx+1:]...)
		} else if len(newArch.LayerSizes) < nas.Config.MaxLayers+1 {
			// Add a hidden layer
			idx := 1 + rand.Intn(len(newArch.LayerSizes)-1)
			newSize := nas.Config.LayerSizeOptions[rand.Intn(len(nas.Config.LayerSizeOptions))]
			newArch.LayerSizes = append(newArch.LayerSizes[:idx], append([]int{newSize}, newArch.LayerSizes[idx:]...)...)
			newArch.Activations = append(newArch.Activations[:idx], append([]string{"relu"}, newArch.Activations[idx:]...)...)
			newArch.WeightBits = append(newArch.WeightBits[:idx], append([]int{6}, newArch.WeightBits[idx:]...)...)
			newArch.ActBits = append(newArch.ActBits[:idx], append([]int{6}, newArch.ActBits[idx:]...)...)
		}
	}

	return newArch
}

// Crossover combines two parent architectures.
func (nas *CIMNASSearch) Crossover(parent1, parent2 *NASArchitecture) *NASArchitecture {
	// Use shorter parent's length
	numLayers := len(parent1.LayerSizes)
	if len(parent2.LayerSizes) < numLayers {
		numLayers = len(parent2.LayerSizes)
	}

	newArch := &NASArchitecture{
		LayerSizes:  make([]int, numLayers),
		Activations: make([]string, numLayers-1),
		WeightBits:  make([]int, numLayers-1),
		ActBits:     make([]int, numLayers-1),
	}

	// Keep input/output sizes from parent1
	newArch.LayerSizes[0] = parent1.LayerSizes[0]
	newArch.LayerSizes[numLayers-1] = parent1.LayerSizes[len(parent1.LayerSizes)-1]

	// Randomly select from parents for hidden layers
	for l := 1; l < numLayers-1; l++ {
		if rand.Float64() < 0.5 {
			newArch.LayerSizes[l] = parent1.LayerSizes[l]
		} else {
			newArch.LayerSizes[l] = parent2.LayerSizes[l]
		}
	}

	for l := 0; l < numLayers-1; l++ {
		if rand.Float64() < 0.5 && l < len(parent1.Activations) {
			newArch.Activations[l] = parent1.Activations[l]
			newArch.WeightBits[l] = parent1.WeightBits[l]
			newArch.ActBits[l] = parent1.ActBits[l]
		} else if l < len(parent2.Activations) {
			newArch.Activations[l] = parent2.Activations[l]
			newArch.WeightBits[l] = parent2.WeightBits[l]
			newArch.ActBits[l] = parent2.ActBits[l]
		} else {
			newArch.Activations[l] = "relu"
			newArch.WeightBits[l] = 6
			newArch.ActBits[l] = 6
		}
	}

	return newArch
}

// RunSearch performs the NAS search.
func (nas *CIMNASSearch) RunSearch(inputSize, outputSize int) *NASArchitecture {
	// Initialize
	nas.InitializePopulation(inputSize, outputSize)

	// Evaluate initial population
	for _, arch := range nas.Population {
		nas.EvaluateArchitecture(arch)
	}

	// Evolution loop
	for gen := 0; gen < nas.Config.NumGenerations; gen++ {
		nas.Generation = gen

		// Sort by fitness
		sort.Slice(nas.Population, func(i, j int) bool {
			return nas.Population[i].Fitness > nas.Population[j].Fitness
		})

		// Track best
		if nas.Population[0].Fitness > nas.BestFitness {
			nas.BestFitness = nas.Population[0].Fitness
			nas.BestArch = nas.Population[0]
		}
		nas.SearchHistory = append(nas.SearchHistory, nas.BestFitness)

		// Selection (elitism + tournament)
		newPop := make([]*NASArchitecture, nas.Config.PopulationSize)
		eliteCount := nas.Config.PopulationSize / 10
		for i := 0; i < eliteCount; i++ {
			newPop[i] = nas.Population[i]
		}

		// Generate offspring
		for i := eliteCount; i < nas.Config.PopulationSize; i++ {
			// Tournament selection
			p1 := nas.TournamentSelect()
			p2 := nas.TournamentSelect()

			var offspring *NASArchitecture
			if rand.Float64() < nas.Config.CrossoverRate {
				offspring = nas.Crossover(p1, p2)
			} else {
				offspring = nas.Mutate(p1)
			}

			if rand.Float64() < nas.Config.MutationRate {
				offspring = nas.Mutate(offspring)
			}

			nas.EvaluateArchitecture(offspring)
			newPop[i] = offspring
		}

		nas.Population = newPop
	}

	return nas.BestArch
}

// TournamentSelect selects an architecture via tournament.
func (nas *CIMNASSearch) TournamentSelect() *NASArchitecture {
	tournamentSize := 3
	best := nas.Population[rand.Intn(len(nas.Population))]

	for i := 1; i < tournamentSize; i++ {
		candidate := nas.Population[rand.Intn(len(nas.Population))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}

	return best
}

// GetParetoFront returns Pareto-optimal architectures.
func (nas *CIMNASSearch) GetParetoFront() []*NASArchitecture {
	pareto := make([]*NASArchitecture, 0)

	for _, arch := range nas.Population {
		dominated := false
		for _, other := range nas.Population {
			if other == arch {
				continue
			}
			// Check if other dominates arch
			if other.Accuracy >= arch.Accuracy &&
				other.Latency <= arch.Latency &&
				other.Energy <= arch.Energy &&
				other.Area <= arch.Area &&
				(other.Accuracy > arch.Accuracy ||
					other.Latency < arch.Latency ||
					other.Energy < arch.Energy ||
					other.Area < arch.Area) {
				dominated = true
				break
			}
		}
		if !dominated {
			pareto = append(pareto, arch)
		}
	}

	return pareto
}

// =============================================================================
// NAS METRICS AND UTILITIES
// =============================================================================

// NASSearchResult contains NAS search results.
type NASSearchResult struct {
	BestArchitecture *NASArchitecture
	ParetoFront      []*NASArchitecture
	SearchHistory    []float64
	TotalGenerations int

	// Performance summary
	BestAccuracy     float64
	BestLatency      float64
	BestEnergy       float64
	BestArea         float64
}

// SummarizeSearch creates a summary of the NAS search.
func (nas *CIMNASSearch) SummarizeSearch() *NASSearchResult {
	pareto := nas.GetParetoFront()

	bestAcc := 0.0
	bestLat := math.MaxFloat64
	bestEnergy := math.MaxFloat64
	bestArea := math.MaxFloat64

	for _, arch := range pareto {
		if arch.Accuracy > bestAcc {
			bestAcc = arch.Accuracy
		}
		if arch.Latency < bestLat {
			bestLat = arch.Latency
		}
		if arch.Energy < bestEnergy {
			bestEnergy = arch.Energy
		}
		if arch.Area < bestArea {
			bestArea = arch.Area
		}
	}

	return &NASSearchResult{
		BestArchitecture: nas.BestArch,
		ParetoFront:      pareto,
		SearchHistory:    nas.SearchHistory,
		TotalGenerations: nas.Generation + 1,
		BestAccuracy:     bestAcc,
		BestLatency:      bestLat,
		BestEnergy:       bestEnergy,
		BestArea:         bestArea,
	}
}

// ExportArchitectureConfig exports architecture as configuration.
func ExportArchitectureConfig(arch *NASArchitecture) map[string]interface{} {
	return map[string]interface{}{
		"layer_sizes":  arch.LayerSizes,
		"activations":  arch.Activations,
		"weight_bits":  arch.WeightBits,
		"act_bits":     arch.ActBits,
		"param_count":  arch.ParamCount,
		"accuracy":     arch.Accuracy,
		"latency_us":   arch.Latency,
		"energy_pj":    arch.Energy,
		"area_mm2":     arch.Area,
	}
}
