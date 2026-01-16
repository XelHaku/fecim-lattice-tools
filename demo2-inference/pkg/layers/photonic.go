// Package layers provides neural network layer implementations for CIM inference.
// photonic.go implements photonic CIM simulation and mixed-precision quantization.
//
// References:
// - MIT Photonic Processor (2024): 92% accuracy, <0.5ns latency
// - LightML Architecture (2025): 325 TOP/s at 3W, 17.1 TOP/s/W
// - Lightmatter Production Processor (2025): ResNet, BERT, RL support
// - MRR Crossbar Arrays: 4x4 prototype for MVM
// - In-memory photonic tensor cores: 880 TOPS/mm², 5.1 TOPS/W

package layers

import (
	"encoding/json"
	"fmt"
	"math"
)

// =============================================================================
// Mixed-Precision Quantization
// =============================================================================

// PrecisionLevel defines quantization bit-widths
type PrecisionLevel int

const (
	Precision2Bit  PrecisionLevel = 2
	Precision4Bit  PrecisionLevel = 4
	Precision6Bit  PrecisionLevel = 6
	Precision8Bit  PrecisionLevel = 8
	Precision16Bit PrecisionLevel = 16
	PrecisionFP32  PrecisionLevel = 32
)

// MixedPrecisionConfig configures per-layer precision
type MixedPrecisionConfig struct {
	// Default precision for layers not explicitly configured
	DefaultWeightBits     PrecisionLevel
	DefaultActivationBits PrecisionLevel

	// Per-layer overrides (layer name -> bits)
	WeightPrecision     map[string]PrecisionLevel
	ActivationPrecision map[string]PrecisionLevel

	// Automatic precision selection
	AutoSelect          bool
	SensitivityThreshold float64 // Accuracy drop threshold for auto-selection

	// Quantization method
	Symmetric           bool    // Symmetric vs asymmetric quantization
	PerChannel          bool    // Per-channel vs per-tensor quantization
	CalibrationSamples  int     // Number of samples for calibration
}

// DefaultMixedPrecisionConfig returns sensible defaults
func DefaultMixedPrecisionConfig() *MixedPrecisionConfig {
	return &MixedPrecisionConfig{
		DefaultWeightBits:     Precision6Bit,
		DefaultActivationBits: Precision8Bit,
		WeightPrecision:       make(map[string]PrecisionLevel),
		ActivationPrecision:   make(map[string]PrecisionLevel),
		AutoSelect:            false,
		SensitivityThreshold:  0.01, // 1% accuracy drop
		Symmetric:             true,
		PerChannel:            true,
		CalibrationSamples:    1000,
	}
}

// MixedPrecisionQuantizer handles mixed-precision quantization
type MixedPrecisionQuantizer struct {
	Config         *MixedPrecisionConfig
	LayerScales    map[string][]float64 // Per-layer quantization scales
	LayerZeroPoints map[string][]int    // Per-layer zero points (asymmetric)
	Sensitivity    map[string]float64   // Per-layer sensitivity scores
}

// NewMixedPrecisionQuantizer creates a new quantizer
func NewMixedPrecisionQuantizer(config *MixedPrecisionConfig) *MixedPrecisionQuantizer {
	if config == nil {
		config = DefaultMixedPrecisionConfig()
	}
	return &MixedPrecisionQuantizer{
		Config:          config,
		LayerScales:     make(map[string][]float64),
		LayerZeroPoints: make(map[string][]int),
		Sensitivity:     make(map[string]float64),
	}
}

// QuantizeWeights quantizes weights to specified precision
func (q *MixedPrecisionQuantizer) QuantizeWeights(layerName string, weights [][]float64) ([][]int, []float64) {
	precision := q.Config.DefaultWeightBits
	if p, ok := q.Config.WeightPrecision[layerName]; ok {
		precision = p
	}

	rows := len(weights)
	cols := len(weights[0])
	quantized := make([][]int, rows)
	for i := range quantized {
		quantized[i] = make([]int, cols)
	}

	var scales []float64

	if q.Config.PerChannel {
		// Per-channel quantization (per output channel)
		scales = make([]float64, rows)
		for i := 0; i < rows; i++ {
			// Find min/max for this channel
			minVal, maxVal := weights[i][0], weights[i][0]
			for j := 1; j < cols; j++ {
				if weights[i][j] < minVal {
					minVal = weights[i][j]
				}
				if weights[i][j] > maxVal {
					maxVal = weights[i][j]
				}
			}

			// Compute scale
			qmax := float64(int(1)<<(precision-1) - 1)
			if q.Config.Symmetric {
				absMax := math.Max(math.Abs(minVal), math.Abs(maxVal))
				scales[i] = absMax / qmax
				if scales[i] == 0 {
					scales[i] = 1e-10
				}
			} else {
				scales[i] = (maxVal - minVal) / (2 * qmax)
				if scales[i] == 0 {
					scales[i] = 1e-10
				}
			}

			// Quantize
			for j := 0; j < cols; j++ {
				q := int(math.Round(weights[i][j] / scales[i]))
				qmin := -int(1 << (precision - 1))
				if q < qmin {
					q = qmin
				}
				if q > int(qmax) {
					q = int(qmax)
				}
				quantized[i][j] = q
			}
		}
	} else {
		// Per-tensor quantization
		minVal, maxVal := weights[0][0], weights[0][0]
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				if weights[i][j] < minVal {
					minVal = weights[i][j]
				}
				if weights[i][j] > maxVal {
					maxVal = weights[i][j]
				}
			}
		}

		qmax := float64(int(1)<<(precision-1) - 1)
		scale := math.Max(math.Abs(minVal), math.Abs(maxVal)) / qmax
		if scale == 0 {
			scale = 1e-10
		}
		scales = []float64{scale}

		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				qv := int(math.Round(weights[i][j] / scale))
				qmin := -int(1 << (precision - 1))
				if qv < qmin {
					qv = qmin
				}
				if qv > int(qmax) {
					qv = int(qmax)
				}
				quantized[i][j] = qv
			}
		}
	}

	q.LayerScales[layerName] = scales
	return quantized, scales
}

// DequantizeWeights converts quantized weights back to float
func (q *MixedPrecisionQuantizer) DequantizeWeights(layerName string, quantized [][]int) [][]float64 {
	scales, ok := q.LayerScales[layerName]
	if !ok {
		return nil
	}

	rows := len(quantized)
	cols := len(quantized[0])
	weights := make([][]float64, rows)
	for i := range weights {
		weights[i] = make([]float64, cols)
	}

	if len(scales) == rows {
		// Per-channel
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				weights[i][j] = float64(quantized[i][j]) * scales[i]
			}
		}
	} else {
		// Per-tensor
		scale := scales[0]
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				weights[i][j] = float64(quantized[i][j]) * scale
			}
		}
	}

	return weights
}

// ComputeSensitivity estimates layer sensitivity to quantization
func (q *MixedPrecisionQuantizer) ComputeSensitivity(layerName string, weights [][]float64, sampleInputs [][]float64) float64 {
	// Reference output with full precision
	refOutputs := make([][]float64, len(sampleInputs))
	for i, input := range sampleInputs {
		refOutputs[i] = matVecMul(weights, input)
	}

	// Test different precisions and compute error
	precisions := []PrecisionLevel{Precision8Bit, Precision6Bit, Precision4Bit, Precision2Bit}
	var totalError float64

	for _, precision := range precisions {
		// Temporarily set precision
		q.Config.WeightPrecision[layerName] = precision
		quantized, _ := q.QuantizeWeights(layerName, weights)
		dequantized := q.DequantizeWeights(layerName, quantized)

		// Compute output error
		for i, input := range sampleInputs {
			quantOutput := matVecMul(dequantized, input)
			for j := range quantOutput {
				diff := quantOutput[j] - refOutputs[i][j]
				totalError += diff * diff
			}
		}
	}

	sensitivity := totalError / float64(len(precisions)*len(sampleInputs))
	q.Sensitivity[layerName] = sensitivity

	// Clean up temporary precision setting
	delete(q.Config.WeightPrecision, layerName)

	return sensitivity
}

// AutoSelectPrecision automatically selects precision based on sensitivity
func (q *MixedPrecisionQuantizer) AutoSelectPrecision(layers map[string][][]float64, sampleInputs [][]float64) {
	if !q.Config.AutoSelect {
		return
	}

	// Compute sensitivity for all layers
	for name, weights := range layers {
		q.ComputeSensitivity(name, weights, sampleInputs)
	}

	// Sort layers by sensitivity
	type layerSens struct {
		name string
		sens float64
	}
	sorted := make([]layerSens, 0, len(q.Sensitivity))
	for name, sens := range q.Sensitivity {
		sorted = append(sorted, layerSens{name, sens})
	}

	// Assign precision based on sensitivity
	// Higher sensitivity = higher precision
	for _, ls := range sorted {
		if ls.sens > q.Config.SensitivityThreshold*10 {
			q.Config.WeightPrecision[ls.name] = Precision8Bit
		} else if ls.sens > q.Config.SensitivityThreshold*5 {
			q.Config.WeightPrecision[ls.name] = Precision6Bit
		} else if ls.sens > q.Config.SensitivityThreshold {
			q.Config.WeightPrecision[ls.name] = Precision4Bit
		} else {
			q.Config.WeightPrecision[ls.name] = Precision2Bit
		}
	}
}

// matVecMul performs matrix-vector multiplication
func matVecMul(matrix [][]float64, vec []float64) []float64 {
	rows := len(matrix)
	result := make([]float64, rows)
	for i := 0; i < rows; i++ {
		sum := 0.0
		for j := 0; j < len(vec) && j < len(matrix[i]); j++ {
			sum += matrix[i][j] * vec[j]
		}
		result[i] = sum
	}
	return result
}

// EstimateModelSize estimates model size in bytes for given precision config
func (q *MixedPrecisionQuantizer) EstimateModelSize(layerParams map[string]int64) int64 {
	var totalBits int64
	for name, params := range layerParams {
		precision := q.Config.DefaultWeightBits
		if p, ok := q.Config.WeightPrecision[name]; ok {
			precision = p
		}
		totalBits += params * int64(precision)
	}
	return totalBits / 8
}

// =============================================================================
// Photonic CIM Simulation
// =============================================================================

// PhotonicTechnology defines the photonic platform
type PhotonicTechnology string

const (
	PhotonicMRR       PhotonicTechnology = "mrr"        // Microring resonator
	PhotonicMZI       PhotonicTechnology = "mzi"        // Mach-Zehnder interferometer
	PhotonicPCM       PhotonicTechnology = "pcm"        // Phase-change memory photonic
	PhotonicFerro     PhotonicTechnology = "ferro"      // Ferroelectric EO modulator
)

// PhotonicCrossbarConfig configures photonic crossbar simulation
type PhotonicCrossbarConfig struct {
	Technology        PhotonicTechnology
	ArraySize         int     // NxN crossbar size
	ClockSpeedGHz     float64 // Operating frequency
	WavelengthNm      float64 // Operating wavelength

	// MRR-specific parameters
	MRRQFactor        float64 // Quality factor
	MRRFSRNm          float64 // Free spectral range

	// Modulator parameters
	ModulationDepthDB float64 // Extinction ratio
	InsertionLossDB   float64 // Per-element loss

	// Noise and non-idealities
	ShotNoisePower    float64 // Shot noise power
	ThermalNoise      float64 // Thermal noise
	CrosstalkDB       float64 // Inter-channel crosstalk

	// Power consumption
	LaserPowerMW      float64
	ModulatorPowerMW  float64 // Per modulator
	DetectorPowerMW   float64 // Per photodetector
}

// DefaultPhotonicConfig returns typical photonic crossbar parameters
func DefaultPhotonicConfig() *PhotonicCrossbarConfig {
	return &PhotonicCrossbarConfig{
		Technology:        PhotonicMRR,
		ArraySize:         64,
		ClockSpeedGHz:     25.0,
		WavelengthNm:      1550.0,
		MRRQFactor:        10000,
		MRRFSRNm:          10.0,
		ModulationDepthDB: 20.0,
		InsertionLossDB:   0.5,
		ShotNoisePower:    1e-12,
		ThermalNoise:      1e-11,
		CrosstalkDB:       -30.0,
		LaserPowerMW:      100.0,
		ModulatorPowerMW:  0.5,
		DetectorPowerMW:   0.1,
	}
}

// PhotonicCrossbar simulates a photonic MVM accelerator
type PhotonicCrossbar struct {
	Config       *PhotonicCrossbarConfig
	Weights      [][]float64 // Weight matrix (encoded in photonic elements)
	WeightPhases [][]float64 // Phase encoding for weights
}

// NewPhotonicCrossbar creates a photonic crossbar simulator
func NewPhotonicCrossbar(config *PhotonicCrossbarConfig) *PhotonicCrossbar {
	if config == nil {
		config = DefaultPhotonicConfig()
	}
	return &PhotonicCrossbar{
		Config:       config,
		Weights:      make([][]float64, config.ArraySize),
		WeightPhases: make([][]float64, config.ArraySize),
	}
}

// ProgramWeights programs weights into the photonic crossbar
func (p *PhotonicCrossbar) ProgramWeights(weights [][]float64) error {
	rows := len(weights)
	if rows > p.Config.ArraySize {
		return fmt.Errorf("weight matrix rows (%d) exceeds array size (%d)", rows, p.Config.ArraySize)
	}
	cols := 0
	if rows > 0 {
		cols = len(weights[0])
	}
	if cols > p.Config.ArraySize {
		return fmt.Errorf("weight matrix cols (%d) exceeds array size (%d)", cols, p.Config.ArraySize)
	}

	// Initialize full-size arrays
	p.Weights = make([][]float64, p.Config.ArraySize)
	p.WeightPhases = make([][]float64, p.Config.ArraySize)
	for i := 0; i < p.Config.ArraySize; i++ {
		p.Weights[i] = make([]float64, p.Config.ArraySize)
		p.WeightPhases[i] = make([]float64, p.Config.ArraySize)
	}

	// Copy weights and compute phase encoding
	for i := 0; i < rows; i++ {
		for j := 0; j < len(weights[i]); j++ {
			p.Weights[i][j] = weights[i][j]
			// Phase encoding: weight -> phase shift (0 to 2π)
			// Normalize weight to [-1, 1] range
			normalized := math.Max(-1, math.Min(1, weights[i][j]))
			// Map to phase: w=1 -> φ=0, w=-1 -> φ=π
			p.WeightPhases[i][j] = math.Acos(normalized)
		}
	}

	return nil
}

// Forward performs optical MVM with realistic noise model
func (p *PhotonicCrossbar) Forward(input []float64) []float64 {
	n := p.Config.ArraySize
	output := make([]float64, n)

	// Optical power at input (modulated by input values)
	inputPower := make([]float64, n)
	for i := 0; i < len(input) && i < n; i++ {
		// Input encoding: value -> optical power modulation
		inputPower[i] = p.Config.LaserPowerMW / float64(n) * math.Abs(input[i])
	}

	// Compute MVM through optical interference
	for i := 0; i < n; i++ {
		var sum float64
		for j := 0; j < n; j++ {
			// Optical field amplitude
			amplitude := math.Sqrt(inputPower[j]) * math.Cos(p.WeightPhases[i][j])

			// Apply insertion loss
			lossLinear := math.Pow(10, -p.Config.InsertionLossDB/10)
			amplitude *= math.Sqrt(lossLinear)

			// Accumulate (coherent addition)
			sum += amplitude * amplitude // Photodetector measures power
		}

		// Add shot noise
		shotNoise := gaussianNoise() * math.Sqrt(p.Config.ShotNoisePower)

		// Add thermal noise
		thermalNoise := gaussianNoise() * math.Sqrt(p.Config.ThermalNoise)

		// Add crosstalk from adjacent channels
		crosstalk := 0.0
		crosstalkLinear := math.Pow(10, p.Config.CrosstalkDB/10)
		if i > 0 {
			crosstalk += output[i-1] * crosstalkLinear
		}

		output[i] = sum + shotNoise + thermalNoise + crosstalk
	}

	return output
}

// EstimatePerformance calculates photonic accelerator metrics
func (p *PhotonicCrossbar) EstimatePerformance() *PhotonicPerformance {
	n := p.Config.ArraySize
	clockGHz := p.Config.ClockSpeedGHz

	// MACs per cycle = N² (full matrix-vector multiply)
	macsPerCycle := int64(n * n)

	// TOPS = MACs/cycle × clock × 2 (multiply-accumulate = 2 ops)
	tops := float64(macsPerCycle) * clockGHz * 2 / 1000.0

	// Power breakdown
	laserPower := p.Config.LaserPowerMW
	modulatorPower := p.Config.ModulatorPowerMW * float64(n) // Input modulators
	weightPower := p.Config.ModulatorPowerMW * float64(n*n)  // Weight elements
	detectorPower := p.Config.DetectorPowerMW * float64(n)   // Output detectors
	totalPowerMW := laserPower + modulatorPower + weightPower + detectorPower
	totalPowerW := totalPowerMW / 1000.0

	// Efficiency
	topsPerW := tops / totalPowerW

	// Latency (single MVM)
	latencyNs := 1.0 / clockGHz

	// Area estimate (assuming 100 µm² per MRR element)
	areaPerElementUM2 := 100.0
	totalAreaMM2 := float64(n*n) * areaPerElementUM2 / 1e6

	return &PhotonicPerformance{
		TOPS:              tops,
		PowerW:            totalPowerW,
		TOPSPerW:          topsPerW,
		LatencyNs:         latencyNs,
		AreaMM2:           totalAreaMM2,
		TOPSPerMM2:        tops / totalAreaMM2,
		EnergyPerMACfJ:    totalPowerW * 1e15 / (float64(macsPerCycle) * clockGHz * 1e9),
		LaserPowerMW:      laserPower,
		ModulatorPowerMW:  modulatorPower + weightPower,
		DetectorPowerMW:   detectorPower,
	}
}

// PhotonicPerformance holds performance metrics
type PhotonicPerformance struct {
	TOPS             float64
	PowerW           float64
	TOPSPerW         float64
	LatencyNs        float64
	AreaMM2          float64
	TOPSPerMM2       float64
	EnergyPerMACfJ   float64
	LaserPowerMW     float64
	ModulatorPowerMW float64
	DetectorPowerMW  float64
}

// String returns a formatted performance summary
func (p *PhotonicPerformance) String() string {
	return fmt.Sprintf(`Photonic Accelerator Performance:
  TOPS:           %.1f
  Power:          %.2f W
  Efficiency:     %.1f TOPS/W
  Latency:        %.2f ns
  Area:           %.4f mm²
  Area Efficiency: %.1f TOPS/mm²
  Energy/MAC:     %.1f fJ`,
		p.TOPS, p.PowerW, p.TOPSPerW, p.LatencyNs,
		p.AreaMM2, p.TOPSPerMM2, p.EnergyPerMACfJ)
}

// gaussianNoise returns a sample from standard normal distribution
func gaussianNoise() float64 {
	// Box-Muller transform (simplified)
	u1 := 0.5 // Would use rand.Float64() in real implementation
	u2 := 0.5
	return math.Sqrt(-2*math.Log(u1+0.001)) * math.Cos(2*math.Pi*u2)
}

// =============================================================================
// Antiferroelectric Device Simulation
// =============================================================================

// AFEDeviceType defines antiferroelectric device variants
type AFEDeviceType string

const (
	AFENeuron  AFEDeviceType = "neuron"  // LIF neuron
	AFESynapse AFEDeviceType = "synapse" // Synaptic device
	AFEHybrid  AFEDeviceType = "hybrid"  // CNN-SNN hybrid
)

// AFEDeviceConfig configures antiferroelectric device parameters
type AFEDeviceConfig struct {
	DeviceType        AFEDeviceType
	Material          string  // "HZO", "PZT", etc.
	ZrContent         float64 // Zr/(Hf+Zr) ratio (e.g., 0.8 for Hf0.2Zr0.8O2)
	ThicknessNm       float64
	AreaUM2           float64

	// Electrical parameters
	CoerciveFieldMV   float64 // MV/cm
	SaturationPolUC   float64 // µC/cm²
	LeakageCurrentA   float64

	// Neuron-specific (LIF behavior)
	ThresholdV        float64
	LeakTimeConstMs   float64 // Spontaneous depolarization time
	RefractoryPeriodUs float64

	// Synapse-specific
	NumStates         int
	PotentiationV     float64
	DepressionV       float64

	// Reliability
	EnduranceCycles   float64
	RetentionSeconds  float64
}

// DefaultAFENeuronConfig returns typical AFE neuron parameters
func DefaultAFENeuronConfig() *AFEDeviceConfig {
	return &AFEDeviceConfig{
		DeviceType:        AFENeuron,
		Material:          "HZO",
		ZrContent:         0.8, // Hf0.2Zr0.8O2 (antiferroelectric)
		ThicknessNm:       6.0,
		AreaUM2:           1.0,
		CoerciveFieldMV:   1.5,
		SaturationPolUC:   30.0,
		LeakageCurrentA:   1e-12,
		ThresholdV:        1.5,
		LeakTimeConstMs:   10.0,
		RefractoryPeriodUs: 100.0,
		EnduranceCycles:   1e12,
		RetentionSeconds:  1e4,
	}
}

// DefaultAFESynapseConfig returns typical AFE synapse parameters
func DefaultAFESynapseConfig() *AFEDeviceConfig {
	return &AFEDeviceConfig{
		DeviceType:        AFESynapse,
		Material:          "HZO",
		ZrContent:         0.5, // Hf0.5Zr0.5O2 (ferroelectric for synapse)
		ThicknessNm:       10.0,
		AreaUM2:           1.0,
		CoerciveFieldMV:   1.0,
		SaturationPolUC:   25.0,
		LeakageCurrentA:   1e-13,
		NumStates:         64,
		PotentiationV:     2.5,
		DepressionV:       -2.5,
		EnduranceCycles:   1e10,
		RetentionSeconds:  1e5,
	}
}

// AFEDevice simulates antiferroelectric neuromorphic devices
type AFEDevice struct {
	Config         *AFEDeviceConfig
	Polarization   float64 // Current polarization state
	MembranePot    float64 // For neurons: membrane potential
	SynapticWeight float64 // For synapses: weight value
	SpikeHistory   []float64
	LastSpikeTime  float64
	InRefractory   bool
}

// NewAFEDevice creates a new AFE device simulator
func NewAFEDevice(config *AFEDeviceConfig) *AFEDevice {
	return &AFEDevice{
		Config:       config,
		SpikeHistory: make([]float64, 0),
	}
}

// IntegrateSpike adds input spike to LIF neuron
func (d *AFEDevice) IntegrateSpike(inputCurrent float64, timestamp float64) bool {
	if d.Config.DeviceType != AFENeuron {
		return false
	}

	// Check refractory period
	if d.InRefractory {
		if timestamp-d.LastSpikeTime > d.Config.RefractoryPeriodUs/1000.0 {
			d.InRefractory = false
		} else {
			return false
		}
	}

	// Leaky integration
	dt := 0.1 // Time step in ms
	leakFactor := math.Exp(-dt / d.Config.LeakTimeConstMs)

	// AFE polarization accumulation models integration
	d.Polarization *= leakFactor
	d.Polarization += inputCurrent * 0.1 // Scale input

	d.MembranePot = d.Polarization

	// Check threshold
	if d.MembranePot > d.Config.ThresholdV {
		// Fire!
		d.SpikeHistory = append(d.SpikeHistory, timestamp)
		d.LastSpikeTime = timestamp
		d.InRefractory = true

		// AFE spontaneous depolarization resets the state
		d.Polarization = 0
		d.MembranePot = 0

		return true
	}

	return false
}

// UpdateSynapse updates synaptic weight with STDP-like rule
func (d *AFEDevice) UpdateSynapse(preSpikeTime, postSpikeTime float64) {
	if d.Config.DeviceType != AFESynapse {
		return
	}

	dt := postSpikeTime - preSpikeTime
	tauPlus := 20.0  // ms
	tauMinus := 20.0 // ms

	var deltaW float64
	if dt > 0 {
		// Potentiation (pre before post)
		deltaW = math.Exp(-dt / tauPlus)
	} else {
		// Depression (post before pre)
		deltaW = -math.Exp(dt / tauMinus)
	}

	// Update weight with saturation
	d.SynapticWeight += deltaW * 0.1
	if d.SynapticWeight < 0 {
		d.SynapticWeight = 0
	}
	if d.SynapticWeight > 1 {
		d.SynapticWeight = 1
	}

	// Map weight to polarization state
	d.Polarization = d.SynapticWeight * d.Config.SaturationPolUC
}

// GetConductance returns current synaptic conductance
func (d *AFEDevice) GetConductance() float64 {
	// Model conductance from polarization
	// G = G_min + (G_max - G_min) * P/Psat
	gMin := 1e-9  // 1 nS
	gMax := 1e-6  // 1 µS
	pNorm := d.Polarization / d.Config.SaturationPolUC
	return gMin + (gMax-gMin)*math.Abs(pNorm)
}

// EstimateEnergy estimates energy per spike or weight update
func (d *AFEDevice) EstimateEnergy() float64 {
	// E = C * V^2 where C ~ area * ε/thickness
	epsilon := 25.0 * 8.85e-12 // HZO relative permittivity
	area := d.Config.AreaUM2 * 1e-12
	thickness := d.Config.ThicknessNm * 1e-9
	capacitance := epsilon * area / thickness

	var voltage float64
	if d.Config.DeviceType == AFENeuron {
		voltage = d.Config.ThresholdV
	} else {
		voltage = d.Config.PotentiationV
	}

	return capacitance * voltage * voltage // Joules
}

// AFENetwork simulates a network of AFE neurons and synapses
type AFENetwork struct {
	Neurons   []*AFEDevice
	Synapses  [][]*AFEDevice // Synapse matrix [pre][post]
	NumInputs int
	NumHidden int
	NumOutput int
}

// NewAFENetwork creates a fully-connected AFE network
func NewAFENetwork(numInputs, numHidden, numOutput int) *AFENetwork {
	net := &AFENetwork{
		NumInputs: numInputs,
		NumHidden: numHidden,
		NumOutput: numOutput,
	}

	// Create hidden and output neurons
	totalNeurons := numHidden + numOutput
	net.Neurons = make([]*AFEDevice, totalNeurons)
	for i := 0; i < totalNeurons; i++ {
		net.Neurons[i] = NewAFEDevice(DefaultAFENeuronConfig())
	}

	// Create synapse matrix
	// Input -> Hidden synapses
	totalPre := numInputs + numHidden
	net.Synapses = make([][]*AFEDevice, totalPre)
	for i := 0; i < totalPre; i++ {
		net.Synapses[i] = make([]*AFEDevice, totalNeurons)
		for j := 0; j < totalNeurons; j++ {
			net.Synapses[i][j] = NewAFEDevice(DefaultAFESynapseConfig())
			// Initialize with random weight
			net.Synapses[i][j].SynapticWeight = 0.5
		}
	}

	return net
}

// Forward processes input spikes through the network
func (net *AFENetwork) Forward(inputSpikes []bool, timestamp float64) []bool {
	outputSpikes := make([]bool, net.NumOutput)

	// Process input -> hidden
	for j := 0; j < net.NumHidden; j++ {
		var totalCurrent float64
		for i := 0; i < net.NumInputs; i++ {
			if inputSpikes[i] {
				totalCurrent += net.Synapses[i][j].GetConductance() * 1e6
			}
		}
		net.Neurons[j].IntegrateSpike(totalCurrent, timestamp)
	}

	// Process hidden -> output
	for j := 0; j < net.NumOutput; j++ {
		var totalCurrent float64
		for i := 0; i < net.NumHidden; i++ {
			// Check if hidden neuron spiked recently
			if len(net.Neurons[i].SpikeHistory) > 0 {
				lastSpike := net.Neurons[i].SpikeHistory[len(net.Neurons[i].SpikeHistory)-1]
				if timestamp-lastSpike < 1.0 { // Within 1ms
					synIdx := net.NumInputs + i
					totalCurrent += net.Synapses[synIdx][net.NumHidden+j].GetConductance() * 1e6
				}
			}
		}
		outputSpikes[j] = net.Neurons[net.NumHidden+j].IntegrateSpike(totalCurrent, timestamp)
	}

	return outputSpikes
}

// EstimateNetworkPower estimates total network power consumption
func (net *AFENetwork) EstimateNetworkPower(spikeRateHz float64) float64 {
	// Energy per spike × spike rate × number of active elements
	energyPerSpike := net.Neurons[0].EstimateEnergy()
	numNeurons := len(net.Neurons)

	// Average spikes per neuron
	avgSpikesPerSec := spikeRateHz * 0.1 // Assume 10% firing rate

	totalPower := float64(numNeurons) * avgSpikesPerSec * energyPerSpike

	// Add synapse update power (assuming STDP)
	energyPerUpdate := net.Synapses[0][0].EstimateEnergy()
	numSynapses := 0
	for i := range net.Synapses {
		numSynapses += len(net.Synapses[i])
	}
	synapsePower := float64(numSynapses) * avgSpikesPerSec * 0.01 * energyPerUpdate

	return totalPower + synapsePower
}

// =============================================================================
// Ferroelectric Photonic Devices
// =============================================================================

// FerroPhotonicConfig for ferroelectric electro-optic modulators
type FerroPhotonicConfig struct {
	Material          string  // "LiNbO3", "BaTiO3", "PZT"
	EOCoefficientPmV  float64 // Electro-optic coefficient (pm/V)
	RefractiveIndex   float64
	WavelengthNm      float64
	ModulatorLengthMm float64
	VPiV              float64 // Half-wave voltage

	// Synaptic behavior
	NumPolarizationStates int
	SwitchingSpeedNs      float64
	RetentionHours        float64
}

// DefaultLiNbO3Config returns LiNbO3 modulator parameters
func DefaultLiNbO3Config() *FerroPhotonicConfig {
	return &FerroPhotonicConfig{
		Material:              "LiNbO3",
		EOCoefficientPmV:      31.0, // r33 coefficient
		RefractiveIndex:       2.21,
		WavelengthNm:          1550.0,
		ModulatorLengthMm:     10.0,
		VPiV:                  3.5,
		NumPolarizationStates: 128,
		SwitchingSpeedNs:      1.0,
		RetentionHours:        1000,
	}
}

// DefaultBaTiO3Config returns BaTiO3 modulator parameters
func DefaultBaTiO3Config() *FerroPhotonicConfig {
	return &FerroPhotonicConfig{
		Material:              "BaTiO3",
		EOCoefficientPmV:      1640.0, // Much higher than LiNbO3
		RefractiveIndex:       2.4,
		WavelengthNm:          1550.0,
		ModulatorLengthMm:     0.1, // Can be much shorter
		VPiV:                  0.5,
		NumPolarizationStates: 64,
		SwitchingSpeedNs:      10.0,
		RetentionHours:        100,
	}
}

// FerroPhotonicSynapse simulates a ferroelectric photonic synapse
type FerroPhotonicSynapse struct {
	Config           *FerroPhotonicConfig
	PolarizationState int     // Current polarization level
	Weight           float64 // Normalized weight [0, 1]
	PhaseShift       float64 // Current optical phase shift
}

// NewFerroPhotonicSynapse creates a new ferroelectric photonic synapse
func NewFerroPhotonicSynapse(config *FerroPhotonicConfig) *FerroPhotonicSynapse {
	if config == nil {
		config = DefaultLiNbO3Config()
	}
	return &FerroPhotonicSynapse{
		Config: config,
	}
}

// SetWeight programs the synaptic weight
func (s *FerroPhotonicSynapse) SetWeight(weight float64) {
	s.Weight = math.Max(0, math.Min(1, weight))

	// Map weight to polarization state
	s.PolarizationState = int(s.Weight * float64(s.Config.NumPolarizationStates-1))

	// Compute phase shift from polarization
	// ΔΦ = (π/Vπ) × V_bias × (P/Psat)
	pNorm := float64(s.PolarizationState) / float64(s.Config.NumPolarizationStates-1)
	s.PhaseShift = math.Pi * pNorm
}

// GetTransmission returns optical transmission based on phase
func (s *FerroPhotonicSynapse) GetTransmission() float64 {
	// Mach-Zehnder interferometer transmission
	// T = cos²(ΔΦ/2)
	return math.Pow(math.Cos(s.PhaseShift/2), 2)
}

// ProcessOpticalSignal applies synaptic weighting to optical signal
func (s *FerroPhotonicSynapse) ProcessOpticalSignal(inputPowerMW float64) float64 {
	transmission := s.GetTransmission()
	return inputPowerMW * transmission
}

// FerroPhotonicPerformance metrics for ferroelectric photonic devices
type FerroPhotonicPerformance struct {
	BandwidthGHz       float64
	InsertionLossDB    float64
	ExtinctionRatioDB  float64
	PowerConsumptionMW float64
	FootprintUM2       float64
	WeightPrecisionBits int
}

// EstimatePerformance calculates device metrics
func (s *FerroPhotonicSynapse) EstimatePerformance() *FerroPhotonicPerformance {
	// Bandwidth limited by RC time constant
	// τ = ε × L / (n² × r × V)
	bandwidth := 1.0 / (s.Config.SwitchingSpeedNs * 1e-9) / 1e9 // GHz

	// Insertion loss (typical for ferroelectric modulators)
	insertionLoss := 3.0 // dB for LiNbO3
	if s.Config.Material == "BaTiO3" {
		insertionLoss = 1.0 // Lower for integrated BTO
	}

	// Extinction ratio
	extinctionRatio := 20.0 // dB

	// Power (capacitive switching)
	// P = f × C × V²
	// Assume C ~ 1 pF
	power := bandwidth * 1e9 * 1e-12 * s.Config.VPiV * s.Config.VPiV * 1000 // mW

	// Footprint
	footprint := s.Config.ModulatorLengthMm * 1000 * 10 // µm² (10 µm width)

	return &FerroPhotonicPerformance{
		BandwidthGHz:        bandwidth,
		InsertionLossDB:     insertionLoss,
		ExtinctionRatioDB:   extinctionRatio,
		PowerConsumptionMW:  power,
		FootprintUM2:        footprint,
		WeightPrecisionBits: int(math.Log2(float64(s.Config.NumPolarizationStates))),
	}
}

// =============================================================================
// Utility Functions
// =============================================================================

// CompareAccelerators compares electronic vs photonic CIM
func CompareAccelerators(arraySize int) map[string]interface{} {
	// Electronic FeFET CIM
	electronicConfig := &HardwareConstraints{
		MaxArraySize: arraySize,
		ADCBits:      6,
		WeightBits:   6,
	}

	// Photonic CIM
	photonicConfig := DefaultPhotonicConfig()
	photonicConfig.ArraySize = arraySize
	photonicCrossbar := NewPhotonicCrossbar(photonicConfig)
	photonicPerf := photonicCrossbar.EstimatePerformance()

	// Estimate electronic performance (from estimation.go patterns)
	electronicTOPS := float64(arraySize*arraySize) * 1.0 * 2 / 1000.0 // 1 GHz clock
	electronicPowerW := 0.1 * float64(arraySize*arraySize) / 4096    // Scale with array
	electronicTOPSW := electronicTOPS / electronicPowerW

	comparison := map[string]interface{}{
		"array_size": arraySize,
		"electronic": map[string]float64{
			"tops":       electronicTOPS,
			"power_w":    electronicPowerW,
			"tops_per_w": electronicTOPSW,
		},
		"photonic": map[string]float64{
			"tops":       photonicPerf.TOPS,
			"power_w":    photonicPerf.PowerW,
			"tops_per_w": photonicPerf.TOPSPerW,
			"latency_ns": photonicPerf.LatencyNs,
		},
		"photonic_advantage": map[string]float64{
			"speed":      photonicPerf.TOPS / electronicTOPS,
			"efficiency": photonicPerf.TOPSPerW / electronicTOPSW,
		},
	}

	// Add hardware constraints info
	_ = electronicConfig

	return comparison
}

// ExportMixedPrecisionConfig exports configuration to JSON
func ExportMixedPrecisionConfig(config *MixedPrecisionConfig) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}
