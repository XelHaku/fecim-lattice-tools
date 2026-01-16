// Package layers provides optical interconnect and CIM compiler simulation.
//
// Optical Interconnect Topics:
// - Intel OCI chiplet (4 Tbps, 64 PCIe 5.0 channels)
// - Silicon photonics neural network accelerators
// - MRR-based MAC operations (880 TOPS/mm², 5.1 TOPS/W)
// - Photonic in-memory computing with magneto-optic cells
// - MZI-based weight programming
// - Phase-change material (PCM) optical memory
//
// CIM Compiler Topics:
// - CIM-MLC multi-level compilation (3.2× speedup)
// - CIM-Explorer for BNN/TNN optimization
// - Weight mapping and tiling strategies
// - Dataflow optimization (weight-stationary, output-stationary)
// - NeuroSim-style performance estimation
// - Crossbar utilization optimization
//
// Key findings:
// - Intel OCI: 4 Tbps, 64 channels, 100m reach
// - Photonic CIM: 880 TOPS/mm², 5.1 TOPS/W (64×64 @ 25 GHz)
// - MRR compute density: 420 TOPS/mm² per channel
// - CIM-MLC: 3.2× speedup vs prior compilers
// - Magneto-optic: 10¹² endurance, 4-bit weights
// - Photonic fabric: 6.2 pJ/bit vs 62.5 pJ NVLink
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// OPTICAL INTERCONNECT FOR CIM CHIPLETS
// =============================================================================

// OpticalInterconnectConfig configures optical interconnect parameters.
type OpticalInterconnectConfig struct {
	// Protocol
	Protocol string // "OCI", "UCIe-Optical", "Photonic-Fabric"

	// Bandwidth
	TotalBandwidthTbps float64 // Tbps
	NumChannels        int
	DataRateGbps       float64 // Per channel

	// Distance and latency
	MaxReachMeters  float64
	LatencyNs       float64
	PropagationNsM  float64 // ns per meter

	// Energy
	EnergyPJPerBit float64

	// Wavelength Division Multiplexing
	EnableDWDM       bool
	NumWavelengths   int
	ChannelSpacingGHz float64
}

// DefaultOCIConfig returns Intel OCI chiplet parameters.
func DefaultOCIConfig() *OpticalInterconnectConfig {
	return &OpticalInterconnectConfig{
		Protocol:          "OCI",
		TotalBandwidthTbps: 4.0,
		NumChannels:       64,
		DataRateGbps:      32.0, // PCIe 5.0
		MaxReachMeters:    100,
		LatencyNs:         5,
		PropagationNsM:    5, // ~5 ns/m in fiber
		EnergyPJPerBit:    5,
		EnableDWDM:        true,
		NumWavelengths:    8,
		ChannelSpacingGHz: 100,
	}
}

// DefaultPhotonicFabricConfig returns Celestial AI-style parameters.
func DefaultPhotonicFabricConfig() *OpticalInterconnectConfig {
	return &OpticalInterconnectConfig{
		Protocol:          "Photonic-Fabric",
		TotalBandwidthTbps: 14.4,
		NumChannels:       128,
		DataRateGbps:      112.5,
		MaxReachMeters:    10, // Chip-to-chip
		LatencyNs:         2,
		PropagationNsM:    5,
		EnergyPJPerBit:    6.2, // vs 62.5 pJ for NVLink
		EnableDWDM:        true,
		NumWavelengths:    16,
		ChannelSpacingGHz: 50,
	}
}

// OpticalChannel represents a single optical channel.
type OpticalChannel struct {
	ID           int
	Wavelength   float64 // nm
	DataRateGbps float64
	Active       bool
	BER          float64 // Bit error rate
	PowerDBm     float64 // Optical power
}

// OpticalLink represents an optical interconnect link.
type OpticalLink struct {
	Config   *OpticalInterconnectConfig
	Channels []*OpticalChannel

	// State
	TotalBitsTransferred uint64
	TotalEnergy          float64
	Errors               uint64

	// Performance
	EffectiveBandwidth float64 // Actual achieved
	Utilization        float64
}

// NewOpticalLink creates an optical interconnect link.
func NewOpticalLink(config *OpticalInterconnectConfig) *OpticalLink {
	link := &OpticalLink{
		Config:   config,
		Channels: make([]*OpticalChannel, config.NumChannels),
	}

	// Initialize channels with DWDM wavelengths
	baseWavelength := 1550.0 // nm, C-band
	for i := 0; i < config.NumChannels; i++ {
		wavelength := baseWavelength
		if config.EnableDWDM {
			// DWDM channel spacing
			wavelengthIdx := i % config.NumWavelengths
			wavelength = baseWavelength + float64(wavelengthIdx)*0.8 // ~0.8nm per 100GHz
		}

		link.Channels[i] = &OpticalChannel{
			ID:           i,
			Wavelength:   wavelength,
			DataRateGbps: config.DataRateGbps,
			Active:       true,
			BER:          1e-15, // Typical optical BER
			PowerDBm:     0,     // Reference power
		}
	}

	return link
}

// TransferData simulates data transfer over optical link.
func (ol *OpticalLink) TransferData(sizeBytes uint64) (latencyNs float64, energyPJ float64) {
	bits := sizeBytes * 8

	// Calculate transfer time
	bandwidthBps := ol.Config.TotalBandwidthTbps * 1e12
	transferTimeNs := float64(bits) / bandwidthBps * 1e9

	// Total latency = propagation + transfer
	latencyNs = ol.Config.LatencyNs + transferTimeNs

	// Energy
	energyPJ = float64(bits) * ol.Config.EnergyPJPerBit

	// Update statistics
	ol.TotalBitsTransferred += bits
	ol.TotalEnergy += energyPJ

	// Simulate errors
	for _, ch := range ol.Channels {
		if ch.Active && rand.Float64() < ch.BER*float64(bits/uint64(ol.Config.NumChannels)) {
			ol.Errors++
		}
	}

	return latencyNs, energyPJ
}

// GetEfficiency returns energy efficiency in pJ/bit.
func (ol *OpticalLink) GetEfficiency() float64 {
	if ol.TotalBitsTransferred == 0 {
		return ol.Config.EnergyPJPerBit
	}
	return ol.TotalEnergy / float64(ol.TotalBitsTransferred)
}

// =============================================================================
// SILICON PHOTONICS NEURAL NETWORK ACCELERATOR
// =============================================================================

// MRRConfig configures microring resonator parameters.
type MRRConfig struct {
	// Geometry
	RadiusUm     float64 // Bend radius
	GapNm        float64 // Coupling gap
	WidthNm      float64 // Waveguide width

	// Optical properties
	FSRGHZ       float64 // Free spectral range
	QFactor      float64 // Quality factor
	ExtinctionDB float64 // On/off extinction ratio

	// Tuning
	TuningRangePm   float64 // Wavelength tuning range
	TuningPowerMW   float64 // Heater power for full tuning
	TuningSpeedKHz  float64 // Thermal tuning bandwidth
}

// DefaultMRRConfig returns typical silicon MRR parameters.
func DefaultMRRConfig() *MRRConfig {
	return &MRRConfig{
		RadiusUm:      5,
		GapNm:         200,
		WidthNm:       450,
		FSRGHZ:        2400, // ~20nm at 1550nm
		QFactor:       10000,
		ExtinctionDB:  20,
		TuningRangePm: 1000, // 1nm
		TuningPowerMW: 10,
		TuningSpeedKHz: 100,
	}
}

// PhotonicMACConfig configures photonic MAC unit.
type PhotonicMACConfig struct {
	// Architecture
	Architecture string // "MRR", "MZI", "Hybrid"
	ArraySize    int    // Weights per MAC unit

	// Performance
	ClockGHz       float64
	ComputeDensity float64 // TOPS/mm²
	EnergyEfficiency float64 // TOPS/W

	// Precision
	WeightBits int
	InputBits  int
	OutputBits int

	// Area
	CellSizeUm2 float64 // Per weight cell
}

// DefaultMRRMACConfig returns MRR-based MAC parameters.
func DefaultMRRMACConfig() *PhotonicMACConfig {
	return &PhotonicMACConfig{
		Architecture:     "MRR",
		ArraySize:        64,
		ClockGHz:         25,
		ComputeDensity:   420, // TOPS/mm² per channel
		EnergyEfficiency: 5.1, // TOPS/W
		WeightBits:       4,
		InputBits:        8,
		OutputBits:       8,
		CellSizeUm2:      900, // 30×30 μm²
	}
}

// PhotonicWeight represents an optical weight element.
type PhotonicWeight struct {
	Value       float64 // Weight value (normalized)
	Transmission float64 // Optical transmission (0-1)
	Phase       float64 // Phase shift (radians)
	Wavelength  float64 // Operating wavelength
}

// PhotonicMACUnit implements photonic MAC computation.
type PhotonicMACUnit struct {
	Config  *PhotonicMACConfig
	Weights []*PhotonicWeight

	// MRR state
	MRRConfigs []*MRRConfig
	Resonances []float64 // Current resonance wavelengths

	// Statistics
	TotalMACs    uint64
	TotalEnergy  float64
	Latency      float64
}

// NewPhotonicMACUnit creates a photonic MAC unit.
func NewPhotonicMACUnit(config *PhotonicMACConfig) *PhotonicMACUnit {
	pmu := &PhotonicMACUnit{
		Config:     config,
		Weights:    make([]*PhotonicWeight, config.ArraySize),
		MRRConfigs: make([]*MRRConfig, config.ArraySize),
		Resonances: make([]float64, config.ArraySize),
	}

	mrrConfig := DefaultMRRConfig()
	baseWavelength := 1550.0

	for i := 0; i < config.ArraySize; i++ {
		pmu.Weights[i] = &PhotonicWeight{
			Value:       0,
			Transmission: 1.0,
			Phase:       0,
			Wavelength:  baseWavelength + float64(i)*0.8,
		}
		pmu.MRRConfigs[i] = mrrConfig
		pmu.Resonances[i] = baseWavelength + float64(i)*0.8
	}

	return pmu
}

// ProgramWeights sets optical weight values.
func (pmu *PhotonicMACUnit) ProgramWeights(weights []float64) error {
	if len(weights) != pmu.Config.ArraySize {
		return fmt.Errorf("weight count mismatch")
	}

	for i, w := range weights {
		// Normalize weight to transmission (0-1)
		transmission := (w + 1) / 2 // Map [-1,1] to [0,1]
		if transmission < 0 {
			transmission = 0
		}
		if transmission > 1 {
			transmission = 1
		}

		pmu.Weights[i].Value = w
		pmu.Weights[i].Transmission = transmission

		// For MRR: adjust resonance to set transmission
		// Detuning from resonance controls transmission
		detuning := math.Acos(2*transmission-1) * 10 // Simplified model
		pmu.Resonances[i] = pmu.Weights[i].Wavelength + detuning*0.001
	}

	return nil
}

// ComputeMAC performs optical MAC operation.
func (pmu *PhotonicMACUnit) ComputeMAC(inputs []float64) (float64, error) {
	if len(inputs) != pmu.Config.ArraySize {
		return 0, fmt.Errorf("input count mismatch")
	}

	// Optical domain computation
	// Input modulation → Weight multiplication → Photodetector sum
	sum := 0.0
	for i := 0; i < pmu.Config.ArraySize; i++ {
		// Input modulates optical power
		inputPower := math.Abs(inputs[i])
		if inputs[i] < 0 {
			inputPower = -inputPower // Sign handling
		}

		// Weight via transmission
		weightedPower := inputPower * pmu.Weights[i].Transmission

		// For signed weights, use differential detection
		if pmu.Weights[i].Value < 0 {
			weightedPower = -weightedPower
		}

		sum += weightedPower
	}

	pmu.TotalMACs += uint64(pmu.Config.ArraySize)

	// Energy: modulation + detection
	energyPerMAC := 1.0 / pmu.Config.EnergyEfficiency / 1e12 // Convert TOPS/W to J/MAC
	pmu.TotalEnergy += energyPerMAC * float64(pmu.Config.ArraySize)

	return sum, nil
}

// =============================================================================
// MAGNETO-OPTIC MEMORY FOR PHOTONIC CIM
// =============================================================================

// MagnetoOpticConfig configures magneto-optic memory cell.
type MagnetoOpticConfig struct {
	// Material: Ce:YIG on silicon
	Material string

	// Performance
	Endurance      float64 // Cycles (10¹² demonstrated)
	SwitchingTimeNs float64
	SwitchingEnergyPJ float64

	// Multi-level
	NumLevels int // 4-bit = 16 levels demonstrated
	BitDepth  int

	// Optical
	ExtinctionDB    float64
	InsertionLossDB float64
}

// DefaultMagnetoOpticConfig returns Ce:YIG parameters.
func DefaultMagnetoOpticConfig() *MagnetoOpticConfig {
	return &MagnetoOpticConfig{
		Material:         "Ce:YIG",
		Endurance:        1e12,
		SwitchingTimeNs:  10,
		SwitchingEnergyPJ: 100,
		NumLevels:        16, // 4-bit
		BitDepth:         4,
		ExtinctionDB:     15,
		InsertionLossDB:  2,
	}
}

// MagnetoOpticCell represents a magneto-optic memory cell.
type MagnetoOpticCell struct {
	Config *MagnetoOpticConfig

	// State
	Level        int     // Current level (0 to NumLevels-1)
	Transmission float64 // Optical transmission

	// Statistics
	WriteCount uint64
}

// NewMagnetoOpticCell creates a magneto-optic cell.
func NewMagnetoOpticCell(config *MagnetoOpticConfig) *MagnetoOpticCell {
	return &MagnetoOpticCell{
		Config:       config,
		Level:        0,
		Transmission: 1.0,
	}
}

// SetLevel programs the cell to a specific level.
func (moc *MagnetoOpticCell) SetLevel(level int) {
	if level < 0 {
		level = 0
	}
	if level >= moc.Config.NumLevels {
		level = moc.Config.NumLevels - 1
	}

	moc.Level = level
	moc.WriteCount++

	// Calculate transmission based on level
	// Level 0 = max transmission, Level max = min transmission
	extinctionLinear := math.Pow(10, -moc.Config.ExtinctionDB/10)
	moc.Transmission = 1.0 - float64(level)/float64(moc.Config.NumLevels-1)*(1.0-extinctionLinear)
}

// GetWeight returns normalized weight value.
func (moc *MagnetoOpticCell) GetWeight() float64 {
	// Map level to weight [-1, 1]
	return 2.0*float64(moc.Level)/float64(moc.Config.NumLevels-1) - 1.0
}

// =============================================================================
// CIM COMPILER AND MAPPING OPTIMIZATION
// =============================================================================

// ComputingMode represents CIM computing granularity.
type ComputingMode int

const (
	ChipMode     ComputingMode = iota // Coarsest: entire chip
	CoreMode                          // Processing element level
	CrossbarMode                      // Crossbar array level
	WordlineMode                      // Finest: single wordline
)

// CIMCompilerConfig configures CIM compilation.
type CIMCompilerConfig struct {
	// Target hardware
	NumCores     int
	CrossbarsPerCore int
	CrossbarRows int
	CrossbarCols int

	// Computing mode
	Mode ComputingMode

	// Optimization targets
	OptimizeFor string // "latency", "energy", "utilization"

	// Dataflow
	Dataflow string // "weight-stationary", "output-stationary", "row-stationary"

	// Constraints
	MaxMemoryMB    float64
	MaxLatencyMs   float64
	MinUtilization float64
}

// DefaultCIMCompilerConfig returns typical CIM compiler settings.
func DefaultCIMCompilerConfig() *CIMCompilerConfig {
	return &CIMCompilerConfig{
		NumCores:         16,
		CrossbarsPerCore: 4,
		CrossbarRows:     128,
		CrossbarCols:     128,
		Mode:             CrossbarMode,
		OptimizeFor:      "latency",
		Dataflow:         "weight-stationary",
		MaxMemoryMB:      64,
		MaxLatencyMs:     100,
		MinUtilization:   0.7,
	}
}

// LayerMapping represents how a NN layer maps to CIM hardware.
type LayerMapping struct {
	LayerName    string
	LayerType    string // "conv2d", "fc", "matmul"

	// Input/output dimensions
	InputShape  []int
	OutputShape []int
	WeightShape []int

	// Tiling
	TileM int // Output tile size
	TileN int // Input tile size
	TileK int // Reduction dimension

	// Crossbar assignment
	CrossbarIDs    []int
	NumCrossbars   int
	CrossbarUtil   float64

	// Performance estimates
	Latency       float64 // ns
	Energy        float64 // pJ
	MACs          int64
}

// NetworkMapping represents full network mapping.
type NetworkMapping struct {
	Layers           []*LayerMapping
	TotalCrossbars   int
	TotalMACs        int64
	TotalLatency     float64
	TotalEnergy      float64
	AverageUtil      float64
	SpeedupVsBaseline float64
}

// CIMCompiler implements CIM compilation and mapping.
type CIMCompiler struct {
	Config *CIMCompilerConfig

	// Hardware model
	TotalCrossbars int
	CrossbarCells  int

	// Compiled mappings
	CurrentMapping *NetworkMapping
}

// NewCIMCompiler creates a CIM compiler instance.
func NewCIMCompiler(config *CIMCompilerConfig) *CIMCompiler {
	return &CIMCompiler{
		Config:         config,
		TotalCrossbars: config.NumCores * config.CrossbarsPerCore,
		CrossbarCells:  config.CrossbarRows * config.CrossbarCols,
	}
}

// MapLayer compiles a single layer to CIM hardware.
func (cc *CIMCompiler) MapLayer(name, layerType string, inputShape, weightShape, outputShape []int) *LayerMapping {
	mapping := &LayerMapping{
		LayerName:   name,
		LayerType:   layerType,
		InputShape:  inputShape,
		OutputShape: outputShape,
		WeightShape: weightShape,
	}

	// Calculate weight matrix dimensions
	var M, N, K int
	switch layerType {
	case "fc", "matmul":
		if len(weightShape) >= 2 {
			M = weightShape[0] // Output features
			K = weightShape[1] // Input features
			N = 1
			if len(inputShape) > 0 {
				N = inputShape[0] // Batch size
			}
		}
	case "conv2d":
		if len(weightShape) >= 4 {
			// [OutChannels, InChannels, H, W]
			outChannels := weightShape[0]
			inChannels := weightShape[1]
			kernelH := weightShape[2]
			kernelW := weightShape[3]

			// Im2col transformation
			M = outChannels
			K = inChannels * kernelH * kernelW
			if len(outputShape) >= 2 {
				N = outputShape[len(outputShape)-2] * outputShape[len(outputShape)-1]
			}
		}
	}

	// Calculate tiling based on crossbar size
	mapping.TileM = cc.Config.CrossbarCols
	mapping.TileK = cc.Config.CrossbarRows

	// Calculate number of crossbars needed
	crossbarsM := (M + mapping.TileM - 1) / mapping.TileM
	crossbarsK := (K + mapping.TileK - 1) / mapping.TileK
	mapping.NumCrossbars = crossbarsM * crossbarsK

	// Assign crossbar IDs
	mapping.CrossbarIDs = make([]int, mapping.NumCrossbars)
	for i := 0; i < mapping.NumCrossbars; i++ {
		mapping.CrossbarIDs[i] = i % cc.TotalCrossbars
	}

	// Calculate utilization
	usedCells := M * K
	totalCells := mapping.NumCrossbars * cc.CrossbarCells
	if totalCells > 0 {
		mapping.CrossbarUtil = float64(usedCells) / float64(totalCells)
	}

	// Estimate MACs
	mapping.MACs = int64(M) * int64(N) * int64(K)

	// Estimate latency (ns)
	// Assuming 1 cycle per MVM, 1 GHz clock
	mvmOps := (M + mapping.TileM - 1) / mapping.TileM *
		(K + mapping.TileK - 1) / mapping.TileK *
		N
	mapping.Latency = float64(mvmOps) * 1.0 // 1 ns per MVM at 1 GHz

	// Estimate energy (pJ)
	// Typical: 0.1-1 pJ per MAC in CIM
	mapping.Energy = float64(mapping.MACs) * 0.5

	return mapping
}

// MapNetwork compiles an entire network.
func (cc *CIMCompiler) MapNetwork(layers []struct {
	Name        string
	Type        string
	InputShape  []int
	WeightShape []int
	OutputShape []int
}) *NetworkMapping {
	nm := &NetworkMapping{
		Layers: make([]*LayerMapping, len(layers)),
	}

	totalUtil := 0.0
	for i, layer := range layers {
		mapping := cc.MapLayer(layer.Name, layer.Type,
			layer.InputShape, layer.WeightShape, layer.OutputShape)
		nm.Layers[i] = mapping

		nm.TotalCrossbars += mapping.NumCrossbars
		nm.TotalMACs += mapping.MACs
		nm.TotalLatency += mapping.Latency
		nm.TotalEnergy += mapping.Energy
		totalUtil += mapping.CrossbarUtil
	}

	if len(layers) > 0 {
		nm.AverageUtil = totalUtil / float64(len(layers))
	}

	// Speedup estimate (vs sequential baseline)
	// CIM-MLC achieves 3.2× on average
	nm.SpeedupVsBaseline = 3.2

	cc.CurrentMapping = nm
	return nm
}

// OptimizeTiling finds optimal tile sizes for a layer.
func (cc *CIMCompiler) OptimizeTiling(M, K, N int) (tileM, tileK int, utilization float64) {
	bestUtil := 0.0
	bestTileM := cc.Config.CrossbarCols
	bestTileK := cc.Config.CrossbarRows

	// Search tile sizes that fit crossbar
	for tm := 16; tm <= cc.Config.CrossbarCols; tm *= 2 {
		for tk := 16; tk <= cc.Config.CrossbarRows; tk *= 2 {
			// Calculate utilization with this tiling
			crossbarsM := (M + tm - 1) / tm
			crossbarsK := (K + tk - 1) / tk

			usedCells := M * K
			totalCells := crossbarsM * crossbarsK * tm * tk
			util := float64(usedCells) / float64(totalCells)

			// Check constraints
			neededCrossbars := crossbarsM * crossbarsK
			if neededCrossbars <= cc.TotalCrossbars && util > bestUtil {
				bestUtil = util
				bestTileM = tm
				bestTileK = tk
			}
		}
	}

	return bestTileM, bestTileK, bestUtil
}

// =============================================================================
// NEUROSIM-STYLE PERFORMANCE ESTIMATION
// =============================================================================

// NeuroSimConfig configures NeuroSim-style estimation.
type NeuroSimConfig struct {
	// Technology
	ProcessNm      int     // Technology node
	MemoryType     string  // "SRAM", "RRAM", "FeFET"

	// Array parameters
	ArrayRows      int
	ArrayCols      int
	CellResistance float64 // Ohms

	// ADC/DAC
	ADCBits        int
	ADCLatencyNs   float64
	ADCEnergyPJ    float64
	DACBits        int
	DACLatencyNs   float64
	DACEnergyPJ    float64

	// Non-idealities
	CellVariation  float64 // Standard deviation
	WireResistance float64 // Ohms per cell
	ReadNoise      float64 // Percentage
}

// DefaultNeuroSimConfig returns NeuroSim default parameters.
func DefaultNeuroSimConfig() *NeuroSimConfig {
	return &NeuroSimConfig{
		ProcessNm:      28,
		MemoryType:     "RRAM",
		ArrayRows:      128,
		ArrayCols:      128,
		CellResistance: 100e3, // 100 kΩ
		ADCBits:        8,
		ADCLatencyNs:   100,
		ADCEnergyPJ:    500,
		DACBits:        4,
		DACLatencyNs:   10,
		DACEnergyPJ:    50,
		CellVariation:  0.05,
		WireResistance: 2.5,
		ReadNoise:      0.02,
	}
}

// NeuroSimEstimator provides performance estimation.
type NeuroSimEstimator struct {
	Config *NeuroSimConfig

	// Cached calculations
	ArrayArea      float64 // mm²
	ArrayLatency   float64 // ns per MVM
	ArrayEnergy    float64 // pJ per MVM
	ComputeDensity float64 // TOPS/mm²
	Efficiency     float64 // TOPS/W
}

// NewNeuroSimEstimator creates a NeuroSim estimator.
func NewNeuroSimEstimator(config *NeuroSimConfig) *NeuroSimEstimator {
	nse := &NeuroSimEstimator{Config: config}
	nse.calculateMetrics()
	return nse
}

// calculateMetrics computes performance metrics.
func (nse *NeuroSimEstimator) calculateMetrics() {
	cfg := nse.Config

	// Area estimation (simplified)
	// Cell area scales with process node
	cellAreaUm2 := float64(cfg.ProcessNm) * 0.5 // Simplified scaling
	nse.ArrayArea = float64(cfg.ArrayRows*cfg.ArrayCols) * cellAreaUm2 * 1e-6 // mm²

	// Latency estimation
	// MVM latency = DAC + compute + ADC
	computeLatency := 10.0 // ns for analog computation
	nse.ArrayLatency = cfg.DACLatencyNs + computeLatency + cfg.ADCLatencyNs

	// Energy estimation
	// Per MVM = DAC + array read + ADC
	arrayReadEnergy := float64(cfg.ArrayRows*cfg.ArrayCols) * 0.1 // pJ per cell
	nse.ArrayEnergy = cfg.DACEnergyPJ*float64(cfg.ArrayRows) +
		arrayReadEnergy +
		cfg.ADCEnergyPJ*float64(cfg.ArrayCols)

	// Compute density (TOPS/mm²)
	macs := cfg.ArrayRows * cfg.ArrayCols
	opsPerSecond := float64(macs) * 2 / (nse.ArrayLatency * 1e-9) // 2 ops per MAC
	nse.ComputeDensity = opsPerSecond / 1e12 / nse.ArrayArea

	// Energy efficiency (TOPS/W)
	powerW := nse.ArrayEnergy * 1e-12 / (nse.ArrayLatency * 1e-9)
	nse.Efficiency = opsPerSecond / 1e12 / powerW
}

// EstimateLayer estimates performance for a layer.
func (nse *NeuroSimEstimator) EstimateLayer(M, K, N int) map[string]float64 {
	cfg := nse.Config

	// Number of MVMs needed
	tilesM := (M + cfg.ArrayCols - 1) / cfg.ArrayCols
	tilesK := (K + cfg.ArrayRows - 1) / cfg.ArrayRows
	numMVMs := tilesM * tilesK * N

	// Total MACs
	macs := M * K * N

	// Latency (sequential MVMs)
	latencyNs := float64(numMVMs) * nse.ArrayLatency

	// Energy
	energyPJ := float64(numMVMs) * nse.ArrayEnergy

	// Area (number of arrays needed)
	numArrays := tilesM * tilesK
	areaMm2 := float64(numArrays) * nse.ArrayArea

	// Accuracy degradation from non-idealities
	variationLoss := cfg.CellVariation * 10 // Simplified model
	noiseLoss := cfg.ReadNoise * 5
	accuracyLoss := variationLoss + noiseLoss

	return map[string]float64{
		"latency_ns":     latencyNs,
		"energy_pJ":      energyPJ,
		"area_mm2":       areaMm2,
		"num_arrays":     float64(numArrays),
		"num_mvms":       float64(numMVMs),
		"macs":           float64(macs),
		"accuracy_loss":  accuracyLoss,
		"tops_mm2":       nse.ComputeDensity,
		"tops_w":         nse.Efficiency,
	}
}

// =============================================================================
// CIM-EXPLORER DESIGN SPACE EXPLORATION
// =============================================================================

// CIMExplorerConfig configures design space exploration.
type CIMExplorerConfig struct {
	// Search space
	ArraySizes     []int   // e.g., [64, 128, 256]
	ADCBits        []int   // e.g., [4, 6, 8]
	WeightBits     []int   // e.g., [1, 2, 4]
	MemoryTypes    []string // e.g., ["SRAM", "RRAM", "FeFET"]

	// Objectives
	MaxLatencyMs   float64
	MaxEnergyMJ    float64
	MinAccuracy    float64

	// Network
	NetworkMACs    int64
	NetworkLayers  int
}

// DesignPoint represents a single design configuration.
type DesignPoint struct {
	ArraySize   int
	ADCBits     int
	WeightBits  int
	MemoryType  string

	// Estimated metrics
	Latency     float64
	Energy      float64
	Area        float64
	Accuracy    float64
	Efficiency  float64

	// Pareto status
	IsParetoOptimal bool
}

// CIMExplorer performs design space exploration.
type CIMExplorer struct {
	Config *CIMExplorerConfig
	Points []*DesignPoint

	// Pareto frontier
	ParetoFrontier []*DesignPoint
}

// NewCIMExplorer creates a design space explorer.
func NewCIMExplorer(config *CIMExplorerConfig) *CIMExplorer {
	return &CIMExplorer{
		Config: config,
		Points: make([]*DesignPoint, 0),
	}
}

// Explore searches the design space.
func (ce *CIMExplorer) Explore() []*DesignPoint {
	ce.Points = make([]*DesignPoint, 0)

	for _, arraySize := range ce.Config.ArraySizes {
		for _, adcBits := range ce.Config.ADCBits {
			for _, weightBits := range ce.Config.WeightBits {
				for _, memType := range ce.Config.MemoryTypes {
					point := ce.evaluateDesign(arraySize, adcBits, weightBits, memType)
					ce.Points = append(ce.Points, point)
				}
			}
		}
	}

	// Find Pareto frontier
	ce.findParetoFrontier()

	return ce.ParetoFrontier
}

// evaluateDesign estimates metrics for a design point.
func (ce *CIMExplorer) evaluateDesign(arraySize, adcBits, weightBits int, memType string) *DesignPoint {
	dp := &DesignPoint{
		ArraySize:  arraySize,
		ADCBits:    adcBits,
		WeightBits: weightBits,
		MemoryType: memType,
	}

	// Create NeuroSim config for this design
	nsConfig := DefaultNeuroSimConfig()
	nsConfig.ArrayRows = arraySize
	nsConfig.ArrayCols = arraySize
	nsConfig.ADCBits = adcBits
	nsConfig.MemoryType = memType

	// Adjust parameters based on memory type
	switch memType {
	case "RRAM":
		nsConfig.CellVariation = 0.1
		nsConfig.CellResistance = 100e3
	case "FeFET":
		nsConfig.CellVariation = 0.05
		nsConfig.CellResistance = 1e6
	case "SRAM":
		nsConfig.CellVariation = 0.02
		nsConfig.CellResistance = 10e3
	}

	estimator := NewNeuroSimEstimator(nsConfig)

	// Estimate for typical layer
	avgLayerMACs := int(ce.Config.NetworkMACs / int64(ce.Config.NetworkLayers))
	M := int(math.Sqrt(float64(avgLayerMACs)))
	K := M
	N := 1

	metrics := estimator.EstimateLayer(M, K, N)

	dp.Latency = metrics["latency_ns"] * float64(ce.Config.NetworkLayers) / 1e6 // ms
	dp.Energy = metrics["energy_pJ"] * float64(ce.Config.NetworkLayers) / 1e9   // mJ
	dp.Area = metrics["area_mm2"] * float64(ce.Config.NetworkLayers)
	dp.Accuracy = 100 - metrics["accuracy_loss"]
	dp.Efficiency = metrics["tops_w"]

	return dp
}

// findParetoFrontier identifies Pareto-optimal designs.
func (ce *CIMExplorer) findParetoFrontier() {
	ce.ParetoFrontier = make([]*DesignPoint, 0)

	for _, p1 := range ce.Points {
		isDominated := false
		for _, p2 := range ce.Points {
			if p1 == p2 {
				continue
			}
			// Check if p2 dominates p1 (better in all objectives)
			if p2.Latency <= p1.Latency &&
				p2.Energy <= p1.Energy &&
				p2.Accuracy >= p1.Accuracy &&
				(p2.Latency < p1.Latency || p2.Energy < p1.Energy || p2.Accuracy > p1.Accuracy) {
				isDominated = true
				break
			}
		}
		if !isDominated {
			p1.IsParetoOptimal = true
			ce.ParetoFrontier = append(ce.ParetoFrontier, p1)
		}
	}

	// Sort by efficiency
	sort.Slice(ce.ParetoFrontier, func(i, j int) bool {
		return ce.ParetoFrontier[i].Efficiency > ce.ParetoFrontier[j].Efficiency
	})
}

// =============================================================================
// BENCHMARK UTILITIES
// =============================================================================

// PhotonicBenchmark benchmarks photonic accelerators.
type PhotonicBenchmark struct {
	Results []PhotonicBenchmarkResult
}

// PhotonicBenchmarkResult stores benchmark data.
type PhotonicBenchmarkResult struct {
	Architecture    string
	ArraySize       int
	ClockGHz        float64
	ComputeDensity  float64 // TOPS/mm²
	Efficiency      float64 // TOPS/W
	WeightBits      int
	LatencyNs       float64
}

// RunPhotonicBenchmark compares photonic architectures.
func RunPhotonicBenchmark() *PhotonicBenchmark {
	benchmark := &PhotonicBenchmark{}

	configs := []struct {
		name string
		cfg  *PhotonicMACConfig
	}{
		{"MRR-64", &PhotonicMACConfig{
			Architecture: "MRR", ArraySize: 64, ClockGHz: 25,
			ComputeDensity: 420, EnergyEfficiency: 5.1, WeightBits: 4,
		}},
		{"MRR-128", &PhotonicMACConfig{
			Architecture: "MRR", ArraySize: 128, ClockGHz: 25,
			ComputeDensity: 880, EnergyEfficiency: 5.1, WeightBits: 4,
		}},
		{"MZI-64", &PhotonicMACConfig{
			Architecture: "MZI", ArraySize: 64, ClockGHz: 10,
			ComputeDensity: 100, EnergyEfficiency: 10, WeightBits: 8,
		}},
	}

	for _, c := range configs {
		pmu := NewPhotonicMACUnit(c.cfg)

		// Run benchmark MACs
		for i := 0; i < 1000; i++ {
			inputs := make([]float64, c.cfg.ArraySize)
			for j := range inputs {
				inputs[j] = rand.Float64()*2 - 1
			}
			pmu.ComputeMAC(inputs)
		}

		benchmark.Results = append(benchmark.Results, PhotonicBenchmarkResult{
			Architecture:   c.name,
			ArraySize:      c.cfg.ArraySize,
			ClockGHz:       c.cfg.ClockGHz,
			ComputeDensity: c.cfg.ComputeDensity,
			Efficiency:     c.cfg.EnergyEfficiency,
			WeightBits:     c.cfg.WeightBits,
			LatencyNs:      1000 / c.cfg.ClockGHz,
		})
	}

	return benchmark
}

// PrintPhotonicBenchmark outputs photonic benchmark results.
func (b *PhotonicBenchmark) Print() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           PHOTONIC ACCELERATOR BENCHMARK RESULTS                      ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-12s │ %5s │ %8s │ %12s │ %8s │ %4s ║\n",
		"Arch", "Size", "GHz", "TOPS/mm²", "TOPS/W", "Bits")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")

	for _, r := range b.Results {
		fmt.Printf("║ %-12s │ %5d │ %8.1f │ %12.1f │ %8.1f │ %4d ║\n",
			r.Architecture, r.ArraySize, r.ClockGHz,
			r.ComputeDensity, r.Efficiency, r.WeightBits)
	}
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
}

// CompilerBenchmark benchmarks CIM compiler.
type CompilerBenchmark struct {
	NetworkName      string
	TotalLayers      int
	TotalMACs        int64
	TotalCrossbars   int
	CompileTimeMs    float64
	SpeedupVsManual  float64
	AverageUtil      float64
}

// RunCompilerBenchmark tests CIM compilation.
func RunCompilerBenchmark(networkName string, layers []struct {
	Name        string
	Type        string
	InputShape  []int
	WeightShape []int
	OutputShape []int
}) *CompilerBenchmark {
	compiler := NewCIMCompiler(DefaultCIMCompilerConfig())

	mapping := compiler.MapNetwork(layers)

	return &CompilerBenchmark{
		NetworkName:     networkName,
		TotalLayers:     len(layers),
		TotalMACs:       mapping.TotalMACs,
		TotalCrossbars:  mapping.TotalCrossbars,
		CompileTimeMs:   1.5, // Simulated
		SpeedupVsManual: mapping.SpeedupVsBaseline,
		AverageUtil:     mapping.AverageUtil,
	}
}
