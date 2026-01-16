// Package layers provides edge AI deployment and optical-electronic hybrid
// compute-in-memory implementations for IronLattice ferroelectric CIM technology.
//
// This module implements:
// - TinyML deployment pipeline with model compression
// - FeFET-based edge accelerator simulation
// - Photonic tensor core for optical neural networks
// - Ferroelectric-silicon hybrid photonic memory
// - Phase-change material (PCM) optical weights
// - Reservoir computing on silicon photonics
//
// Based on research:
// - FeEdge: FeFET-based CIM for Edge AI (Fraunhofer IPMS)
// - MIT Photonic Processor (Nature 2024)
// - Ferroelectric Photonic Memory (Light: Sci & Apps 2024)
// - Silicon Photonic Reservoir Computing (Nature Communications)
// - Lightmatter Photonic AI Processor (2025)
package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// EDGE AI DEPLOYMENT FRAMEWORK
// =============================================================================

// EdgeDeviceConfig defines edge hardware constraints
type EdgeDeviceConfig struct {
	// Memory constraints
	FlashMemoryKB    int     // Flash memory for model storage (KB)
	RAMMemoryKB      int     // SRAM for inference (KB)
	WeightBits       int     // Supported weight precision

	// Compute constraints
	MACUnits         int     // Number of MAC units
	ClockFreqMHz     float64 // Operating frequency
	SupplyVoltage    float64 // Supply voltage (V)

	// FeFET-specific
	CrossbarSize     int     // FeFET crossbar array size
	ADCBits          int     // ADC resolution
	DACBits          int     // DAC resolution
	FeFETEndurance   float64 // Write cycles

	// Power budget
	PowerBudgetMW    float64 // Total power budget (mW)
	ActivePowerMW    float64 // Active inference power
	IdlePowerMW      float64 // Idle/sleep power
}

// TinyMLDevice returns configuration for typical TinyML device
func TinyMLDevice() *EdgeDeviceConfig {
	return &EdgeDeviceConfig{
		FlashMemoryKB:  256,
		RAMMemoryKB:    64,
		WeightBits:     8,
		MACUnits:       16,
		ClockFreqMHz:   80.0,
		SupplyVoltage:  1.8,
		CrossbarSize:   64,
		ADCBits:        6,
		DACBits:        8,
		FeFETEndurance: 1e10,
		PowerBudgetMW:  10.0,
		ActivePowerMW:  5.0,
		IdlePowerMW:    0.01,
	}
}

// EdgeTPUDevice returns configuration for Edge TPU-class device
func EdgeTPUDevice() *EdgeDeviceConfig {
	return &EdgeDeviceConfig{
		FlashMemoryKB:  8192,
		RAMMemoryKB:    2048,
		WeightBits:     8,
		MACUnits:       256,
		ClockFreqMHz:   500.0,
		SupplyVoltage:  0.9,
		CrossbarSize:   128,
		ADCBits:        8,
		DACBits:        8,
		FeFETEndurance: 1e12,
		PowerBudgetMW:  2000.0,
		ActivePowerMW:  500.0,
		IdlePowerMW:    5.0,
	}
}

// FeFETEdgeDevice returns configuration for FeFET CIM edge accelerator
func FeFETEdgeDevice() *EdgeDeviceConfig {
	return &EdgeDeviceConfig{
		FlashMemoryKB:  1024,
		RAMMemoryKB:    256,
		WeightBits:     6,
		MACUnits:       64, // Per crossbar
		ClockFreqMHz:   100.0,
		SupplyVoltage:  1.2,
		CrossbarSize:   64,
		ADCBits:        6,
		DACBits:        8,
		FeFETEndurance: 1e10,
		PowerBudgetMW:  50.0,
		ActivePowerMW:  10.0,
		IdlePowerMW:    0.1,
	}
}

// ModelProfile describes a neural network model for deployment
type ModelProfile struct {
	Name           string
	NumLayers      int
	TotalParams    int
	TotalMACs      int64
	InputShape     []int
	OutputShape    []int

	// Per-layer info
	LayerParams    []int
	LayerMACs      []int64
	LayerTypes     []string

	// Memory requirements
	WeightMemoryKB float64
	ActivationKB   float64
	PeakRAMKB      float64
}

// ModelCompressor implements model compression for edge deployment
type ModelCompressor struct {
	Device *EdgeDeviceConfig

	// Compression settings
	TargetWeightBits   int
	EnablePruning      bool
	PruningSparsity    float64
	EnableQuantization bool
	EnableDistillation bool

	// Statistics
	OriginalSizeKB     float64
	CompressedSizeKB   float64
	CompressionRatio   float64
	AccuracyDrop       float64
}

// NewModelCompressor creates a new model compressor for target device
func NewModelCompressor(device *EdgeDeviceConfig) *ModelCompressor {
	return &ModelCompressor{
		Device:             device,
		TargetWeightBits:   device.WeightBits,
		EnablePruning:      true,
		PruningSparsity:    0.5,
		EnableQuantization: true,
		EnableDistillation: false,
	}
}

// CompressModel applies compression techniques to fit device constraints
func (mc *ModelCompressor) CompressModel(profile *ModelProfile) *CompressedModel {
	cm := &CompressedModel{
		Original:      profile,
		Device:        mc.Device,
		LayerBits:     make([]int, profile.NumLayers),
		LayerSparsity: make([]float64, profile.NumLayers),
	}

	mc.OriginalSizeKB = profile.WeightMemoryKB

	// Step 1: Quantization
	if mc.EnableQuantization {
		compressionFromQuant := 32.0 / float64(mc.TargetWeightBits)
		for i := 0; i < profile.NumLayers; i++ {
			cm.LayerBits[i] = mc.TargetWeightBits
		}
		cm.QuantizedSizeKB = profile.WeightMemoryKB / compressionFromQuant
	} else {
		cm.QuantizedSizeKB = profile.WeightMemoryKB
		for i := 0; i < profile.NumLayers; i++ {
			cm.LayerBits[i] = 32
		}
	}

	// Step 2: Pruning
	if mc.EnablePruning {
		for i := 0; i < profile.NumLayers; i++ {
			// Apply less pruning to first/last layers (more sensitive)
			if i == 0 || i == profile.NumLayers-1 {
				cm.LayerSparsity[i] = mc.PruningSparsity * 0.5
			} else {
				cm.LayerSparsity[i] = mc.PruningSparsity
			}
		}
		avgSparsity := 0.0
		for _, s := range cm.LayerSparsity {
			avgSparsity += s
		}
		avgSparsity /= float64(profile.NumLayers)
		cm.PrunedSizeKB = cm.QuantizedSizeKB * (1 - avgSparsity)
	} else {
		cm.PrunedSizeKB = cm.QuantizedSizeKB
	}

	cm.FinalSizeKB = cm.PrunedSizeKB
	mc.CompressedSizeKB = cm.FinalSizeKB
	mc.CompressionRatio = mc.OriginalSizeKB / mc.CompressedSizeKB

	// Estimate accuracy drop
	mc.AccuracyDrop = mc.estimateAccuracyDrop(cm)
	cm.EstimatedAccuracyDrop = mc.AccuracyDrop

	// Check if fits in device
	cm.FitsInFlash = cm.FinalSizeKB <= float64(mc.Device.FlashMemoryKB)
	cm.FitsInRAM = profile.PeakRAMKB <= float64(mc.Device.RAMMemoryKB)

	return cm
}

// estimateAccuracyDrop estimates accuracy loss from compression
func (mc *ModelCompressor) estimateAccuracyDrop(cm *CompressedModel) float64 {
	// Heuristic model based on literature
	quantDrop := 0.0
	if mc.EnableQuantization {
		// Lower bits = more drop
		if mc.TargetWeightBits <= 4 {
			quantDrop = 0.02 // ~2% for 4-bit
		} else if mc.TargetWeightBits <= 6 {
			quantDrop = 0.01 // ~1% for 6-bit
		} else {
			quantDrop = 0.005 // ~0.5% for 8-bit
		}
	}

	pruneDrop := 0.0
	if mc.EnablePruning {
		// Higher sparsity = more drop
		pruneDrop = mc.PruningSparsity * 0.02 // ~2% per 100% sparsity
	}

	return quantDrop + pruneDrop
}

// CompressedModel represents a compressed model ready for deployment
type CompressedModel struct {
	Original      *ModelProfile
	Device        *EdgeDeviceConfig

	// Per-layer compression
	LayerBits     []int
	LayerSparsity []float64

	// Size tracking
	QuantizedSizeKB float64
	PrunedSizeKB    float64
	FinalSizeKB     float64

	// Deployment status
	FitsInFlash           bool
	FitsInRAM             bool
	EstimatedAccuracyDrop float64
}

// EdgeDeployer handles deployment to edge devices
type EdgeDeployer struct {
	Device        *EdgeDeviceConfig
	Model         *CompressedModel

	// Deployment configuration
	UseFeFETCrossbar bool
	NumCrossbars     int
	TilingStrategy   string // "row", "column", "block"

	// Performance estimates
	InferenceLatencyMS float64
	InferencePowerMW   float64
	EnergyPerInfUJ     float64
	ThroughputIPS      float64 // Inferences per second
}

// NewEdgeDeployer creates a deployer for target device
func NewEdgeDeployer(device *EdgeDeviceConfig, model *CompressedModel) *EdgeDeployer {
	return &EdgeDeployer{
		Device:           device,
		Model:            model,
		UseFeFETCrossbar: true,
		TilingStrategy:   "block",
	}
}

// PlanDeployment creates deployment plan with resource allocation
func (ed *EdgeDeployer) PlanDeployment() *DeploymentPlan {
	plan := &DeploymentPlan{
		Device:       ed.Device,
		Model:        ed.Model,
		LayerMapping: make([]*LayerMapping, ed.Model.Original.NumLayers),
	}

	// Calculate number of crossbars needed
	totalParams := ed.Model.Original.TotalParams
	paramsPerCrossbar := ed.Device.CrossbarSize * ed.Device.CrossbarSize
	ed.NumCrossbars = (totalParams + paramsPerCrossbar - 1) / paramsPerCrossbar

	// Map layers to crossbars
	crossbarIdx := 0
	for i := 0; i < ed.Model.Original.NumLayers; i++ {
		layerParams := ed.Model.Original.LayerParams[i]
		crossbarsNeeded := (layerParams + paramsPerCrossbar - 1) / paramsPerCrossbar

		plan.LayerMapping[i] = &LayerMapping{
			LayerIndex:      i,
			StartCrossbar:   crossbarIdx,
			NumCrossbars:    crossbarsNeeded,
			Bits:            ed.Model.LayerBits[i],
			Sparsity:        ed.Model.LayerSparsity[i],
		}

		crossbarIdx += crossbarsNeeded
	}

	plan.TotalCrossbars = crossbarIdx

	// Estimate performance
	ed.estimatePerformance(plan)
	plan.InferenceLatencyMS = ed.InferenceLatencyMS
	plan.InferencePowerMW = ed.InferencePowerMW
	plan.EnergyPerInfUJ = ed.EnergyPerInfUJ
	plan.ThroughputIPS = ed.ThroughputIPS

	return plan
}

// estimatePerformance calculates latency, power, throughput
func (ed *EdgeDeployer) estimatePerformance(plan *DeploymentPlan) {
	// FeFET crossbar MVM latency: ~10-100ns per operation
	mvmLatencyNS := 50.0 // 50ns per MVM

	// Total MVMs needed
	totalMVMs := 0
	for i := 0; i < ed.Model.Original.NumLayers; i++ {
		// Each layer needs MACs / crossbar_size MVMs
		layerMVMs := int(ed.Model.Original.LayerMACs[i]) / (ed.Device.CrossbarSize * ed.Device.CrossbarSize)
		if layerMVMs < 1 {
			layerMVMs = 1
		}
		totalMVMs += layerMVMs
	}

	// Account for ADC/DAC overhead
	adcLatencyNS := float64(ed.Device.ADCBits) * 5.0 // ~5ns per bit
	totalLatencyNS := float64(totalMVMs) * (mvmLatencyNS + adcLatencyNS)

	ed.InferenceLatencyMS = totalLatencyNS / 1e6

	// Power estimation
	// FeFET CIM: ~0.1-1 pJ per MAC
	energyPerMAC_pJ := 0.5
	totalEnergy_pJ := float64(ed.Model.Original.TotalMACs) * energyPerMAC_pJ

	// ADC power dominates
	adcPower_mW := float64(ed.Device.ADCBits) * 0.1 * float64(ed.NumCrossbars)
	ed.InferencePowerMW = adcPower_mW + ed.Device.IdlePowerMW

	ed.EnergyPerInfUJ = totalEnergy_pJ / 1e6
	ed.ThroughputIPS = 1000.0 / ed.InferenceLatencyMS
}

// DeploymentPlan contains the full deployment strategy
type DeploymentPlan struct {
	Device       *EdgeDeviceConfig
	Model        *CompressedModel
	LayerMapping []*LayerMapping

	TotalCrossbars     int
	InferenceLatencyMS float64
	InferencePowerMW   float64
	EnergyPerInfUJ     float64
	ThroughputIPS      float64
}

// LayerMapping describes how a layer maps to hardware
type LayerMapping struct {
	LayerIndex    int
	StartCrossbar int
	NumCrossbars  int
	Bits          int
	Sparsity      float64
}

// =============================================================================
// OPTICAL-ELECTRONIC HYBRID CIM
// =============================================================================

// PhotonicConfig configures photonic computing parameters
type PhotonicConfig struct {
	// Optical parameters
	Wavelength       float64 // Operating wavelength (nm), typically 1550
	LaserPowerMW     float64 // Laser power (mW)
	ModulatorBits    int     // Modulator resolution
	DetectorBits     int     // Photodetector ADC bits

	// Array configuration
	ArraySize        int     // MZI array size
	NumWavelengths   int     // WDM channels
	ClockRateGHz     float64 // Operating frequency

	// Nonvolatile memory type
	MemoryType       string  // "PCM", "FeFET", "MRAM"
	WeightBits       int     // Weight precision
	WriteEnergyPJ    float64 // Energy per weight write

	// Noise and precision
	OpticalLossDB    float64 // Insertion loss
	CrosstalkDB      float64 // Inter-channel crosstalk
	DetectorNoise    float64 // Photodetector noise (relative)
}

// DefaultPhotonicConfig returns standard silicon photonic parameters
func DefaultPhotonicConfig() *PhotonicConfig {
	return &PhotonicConfig{
		Wavelength:     1550.0,
		LaserPowerMW:   10.0,
		ModulatorBits:  8,
		DetectorBits:   8,
		ArraySize:      64,
		NumWavelengths: 8,
		ClockRateGHz:   25.0,
		MemoryType:     "PCM",
		WeightBits:     4,
		WriteEnergyPJ:  1.7, // From literature: 1.7 nJ/dB for PCM
		OpticalLossDB:  3.0,
		CrosstalkDB:    -30.0,
		DetectorNoise:  0.01,
	}
}

// FeFETPhotonicConfig returns ferroelectric-photonic hybrid parameters
func FeFETPhotonicConfig() *PhotonicConfig {
	return &PhotonicConfig{
		Wavelength:     1550.0,
		LaserPowerMW:   5.0,
		ModulatorBits:  6,
		DetectorBits:   6,
		ArraySize:      64,
		NumWavelengths: 4,
		ClockRateGHz:   10.0,
		MemoryType:     "FeFET",
		WeightBits:     6,
		WriteEnergyPJ:  50.0,
		OpticalLossDB:  6.6, // From ferroelectric ring resonator
		CrosstalkDB:    -25.0,
		DetectorNoise:  0.02,
	}
}

// MachZehnderModulator simulates an MZI modulator
type MachZehnderModulator struct {
	Config       *PhotonicConfig
	PhaseShift   float64 // Current phase shift (radians)
	Transmission float64 // Current transmission (0-1)
	WeightValue  float64 // Programmed weight

	// Nonvolatile state
	MemoryState  float64 // PCM/FeFET state
	NumWrites    int
}

// NewMachZehnderModulator creates a new MZI modulator
func NewMachZehnderModulator(config *PhotonicConfig) *MachZehnderModulator {
	return &MachZehnderModulator{
		Config:       config,
		PhaseShift:   0,
		Transmission: 1.0,
		WeightValue:  0.5,
		MemoryState:  0.5,
	}
}

// SetWeight programs a weight value to the modulator
func (mzm *MachZehnderModulator) SetWeight(weight float64) {
	// Clamp weight to [0, 1]
	if weight < 0 {
		weight = 0
	} else if weight > 1 {
		weight = 1
	}

	// Map weight to phase shift (0 to π)
	mzm.PhaseShift = weight * math.Pi
	mzm.WeightValue = weight

	// Update memory state
	mzm.MemoryState = weight
	mzm.NumWrites++

	// Calculate transmission: T = cos²(Δφ/2)
	mzm.Transmission = math.Pow(math.Cos(mzm.PhaseShift/2), 2)
}

// Modulate applies the weight to input optical power
func (mzm *MachZehnderModulator) Modulate(inputPower float64) float64 {
	// Apply transmission and loss
	lossFactor := math.Pow(10, -mzm.Config.OpticalLossDB/10)
	return inputPower * mzm.Transmission * lossFactor
}

// PhotonicCrossbar implements optical matrix-vector multiplication
type PhotonicCrossbar struct {
	Config      *PhotonicConfig
	Rows        int
	Cols        int

	// MZI modulator array
	Modulators  [][]*MachZehnderModulator

	// Performance metrics
	ComputeDensityTOPS float64 // TOPS/mm²
	EnergyEffTOPSW     float64 // TOPS/W
	LatencyNS          float64 // Per MVM
}

// NewPhotonicCrossbar creates a new photonic crossbar array
func NewPhotonicCrossbar(config *PhotonicConfig, rows, cols int) *PhotonicCrossbar {
	pc := &PhotonicCrossbar{
		Config:     config,
		Rows:       rows,
		Cols:       cols,
		Modulators: make([][]*MachZehnderModulator, rows),
	}

	for i := 0; i < rows; i++ {
		pc.Modulators[i] = make([]*MachZehnderModulator, cols)
		for j := 0; j < cols; j++ {
			pc.Modulators[i][j] = NewMachZehnderModulator(config)
		}
	}

	// Calculate performance metrics
	// From literature: 880 TOPS/mm² at 25 GHz for 64x64 crossbar
	pc.ComputeDensityTOPS = 880.0 * (float64(rows*cols) / (64.0 * 64.0)) *
		(config.ClockRateGHz / 25.0)
	pc.EnergyEffTOPSW = 5.1 * (float64(rows*cols) / (64.0 * 64.0))
	pc.LatencyNS = 1000.0 / config.ClockRateGHz // One clock cycle

	return pc
}

// SetWeights programs weight matrix to the crossbar
func (pc *PhotonicCrossbar) SetWeights(weights [][]float64) {
	for i := 0; i < pc.Rows && i < len(weights); i++ {
		for j := 0; j < pc.Cols && j < len(weights[i]); j++ {
			// Normalize weight to [0, 1]
			w := (weights[i][j] + 1) / 2
			pc.Modulators[i][j].SetWeight(w)
		}
	}
}

// MatVecMul performs optical matrix-vector multiplication
func (pc *PhotonicCrossbar) MatVecMul(input []float64) []float64 {
	output := make([]float64, pc.Cols)

	// Optical power per input
	powerPerInput := pc.Config.LaserPowerMW / float64(pc.Rows)

	for j := 0; j < pc.Cols; j++ {
		sum := 0.0
		for i := 0; i < pc.Rows && i < len(input); i++ {
			// Input modulation
			inputPower := powerPerInput * input[i]

			// Weight modulation
			outputPower := pc.Modulators[i][j].Modulate(inputPower)

			// Photodetector summation
			sum += outputPower
		}

		// Add detector noise
		noise := rand.NormFloat64() * pc.Config.DetectorNoise * sum
		output[j] = sum + noise
	}

	return output
}

// FerroelectricRingResonator implements FeFET-silicon photonic memory
type FerroelectricRingResonator struct {
	Config *PhotonicConfig

	// Ring parameters
	Radius           float64 // Ring radius (µm)
	CouplingGap      float64 // Bus-ring gap (nm)
	FeFETLength      float64 // FeFET gate length (µm)

	// State
	Polarization     float64 // Ferroelectric polarization (-1 to 1)
	ResonanceShift   float64 // Resonance wavelength shift (nm)
	ExtinctionRatio  float64 // ON/OFF ratio (dB)

	// Memory characteristics
	Endurance        int     // Write cycles completed
	MaxEndurance     int     // Maximum write cycles
	RetentionYears   float64 // Retention time at 85°C
}

// NewFerroelectricRingResonator creates a new FeFET ring resonator
func NewFerroelectricRingResonator(config *PhotonicConfig) *FerroelectricRingResonator {
	return &FerroelectricRingResonator{
		Config:          config,
		Radius:          10.0,    // 10 µm radius
		CouplingGap:     200.0,   // 200 nm gap
		FeFETLength:     2.0,     // 2 µm FeFET
		Polarization:    0.0,
		ExtinctionRatio: 6.6,     // From literature
		MaxEndurance:    40000,   // 4×10⁴ cycles demonstrated
		RetentionYears:  10.0,
	}
}

// Program writes a weight value to the ring resonator
func (frr *FerroelectricRingResonator) Program(weight float64) {
	// Map weight to polarization
	frr.Polarization = 2*weight - 1 // [0,1] -> [-1,1]

	// Calculate resonance shift
	// ~0.5 nm shift per polarization state
	frr.ResonanceShift = frr.Polarization * 0.5

	frr.Endurance++
}

// GetTransmission calculates transmission at operating wavelength
func (frr *FerroelectricRingResonator) GetTransmission(wavelength float64) float64 {
	// Lorentzian lineshape around resonance
	resonance := frr.Config.Wavelength + frr.ResonanceShift
	detuning := wavelength - resonance
	linewidth := 0.1 // nm

	// On-resonance: low transmission, off-resonance: high transmission
	transmissionDB := -frr.ExtinctionRatio * math.Exp(-math.Pow(detuning/linewidth, 2))
	return math.Pow(10, transmissionDB/10)
}

// PCMPhotonicCell implements phase-change material optical weight cell
type PCMPhotonicCell struct {
	Config *PhotonicConfig

	// PCM state
	CrystallineFraction float64 // 0 (amorphous) to 1 (crystalline)
	RefractiveIndex     complex128
	AbsorptionCoeff     float64

	// Multi-level storage
	NumLevels        int
	CurrentLevel     int
	WriteEnergy      float64 // Total write energy (pJ)

	// Performance
	ContrastDB       float64 // Switching contrast
	WriteSpeedNS     float64 // Programming speed
}

// NewPCMPhotonicCell creates a new PCM optical weight cell
func NewPCMPhotonicCell(config *PhotonicConfig) *PCMPhotonicCell {
	return &PCMPhotonicCell{
		Config:              config,
		CrystallineFraction: 0.5,
		NumLevels:           16, // 4-bit from literature
		CurrentLevel:        8,
		ContrastDB:          158.5, // Record-high from literature
		WriteSpeedNS:        100.0,
	}
}

// SetLevel programs a specific weight level
func (pcm *PCMPhotonicCell) SetLevel(level int) {
	if level < 0 {
		level = 0
	} else if level >= pcm.NumLevels {
		level = pcm.NumLevels - 1
	}

	pcm.CurrentLevel = level
	pcm.CrystallineFraction = float64(level) / float64(pcm.NumLevels-1)

	// Energy per write: 1.7 nJ/dB from literature
	pcm.WriteEnergy += pcm.Config.WriteEnergyPJ
}

// GetTransmission returns optical transmission based on PCM state
func (pcm *PCMPhotonicCell) GetTransmission() float64 {
	// Amorphous = high transmission, Crystalline = low transmission
	transmissionDB := -pcm.ContrastDB * pcm.CrystallineFraction / float64(pcm.NumLevels)
	return math.Pow(10, transmissionDB/10)
}

// GetWeight returns normalized weight value
func (pcm *PCMPhotonicCell) GetWeight() float64 {
	return float64(pcm.CurrentLevel) / float64(pcm.NumLevels-1)
}

// PhotonicReservoir implements silicon photonic reservoir computing
type PhotonicReservoir struct {
	Config *PhotonicConfig

	// Reservoir parameters
	NumNodes        int
	ConnectivityMatrix [][]float64
	NodeStates      []float64
	InputWeights    [][]float64
	OutputWeights   [][]float64

	// Ring resonator nodes
	Resonators      []*FerroelectricRingResonator

	// Performance
	SpeedTOPS       float64 // From literature: 200 TOPS
	EnergyRatio     float64 // vs digital: 100× better
}

// NewPhotonicReservoir creates a new photonic reservoir computer
func NewPhotonicReservoir(config *PhotonicConfig, numNodes int) *PhotonicReservoir {
	pr := &PhotonicReservoir{
		Config:             config,
		NumNodes:           numNodes,
		NodeStates:         make([]float64, numNodes),
		Resonators:         make([]*FerroelectricRingResonator, numNodes),
		SpeedTOPS:          200.0, // From literature
		EnergyRatio:        100.0, // 2 orders of magnitude better
	}

	// Initialize resonator nodes
	for i := 0; i < numNodes; i++ {
		pr.Resonators[i] = NewFerroelectricRingResonator(config)
	}

	// Random connectivity matrix
	pr.ConnectivityMatrix = make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		pr.ConnectivityMatrix[i] = make([]float64, numNodes)
		for j := 0; j < numNodes; j++ {
			if rand.Float64() < 0.3 { // 30% connectivity
				pr.ConnectivityMatrix[i][j] = rand.NormFloat64() * 0.5
			}
		}
	}

	return pr
}

// SetInputWeights configures input-to-reservoir mapping
func (pr *PhotonicReservoir) SetInputWeights(weights [][]float64) {
	pr.InputWeights = weights
}

// SetOutputWeights configures reservoir-to-output mapping
func (pr *PhotonicReservoir) SetOutputWeights(weights [][]float64) {
	pr.OutputWeights = weights
}

// Process performs reservoir computation on input
func (pr *PhotonicReservoir) Process(input []float64) []float64 {
	// Input to reservoir
	if pr.InputWeights != nil {
		for i := 0; i < pr.NumNodes; i++ {
			sum := 0.0
			for j := 0; j < len(input) && j < len(pr.InputWeights[i]); j++ {
				sum += input[j] * pr.InputWeights[i][j]
			}
			pr.NodeStates[i] = sum
		}
	}

	// Reservoir dynamics (one step)
	newStates := make([]float64, pr.NumNodes)
	for i := 0; i < pr.NumNodes; i++ {
		sum := 0.0
		for j := 0; j < pr.NumNodes; j++ {
			sum += pr.NodeStates[j] * pr.ConnectivityMatrix[i][j]
		}

		// Nonlinear activation via ring resonator
		// Transmission depends on input power
		pr.Resonators[i].Program((sum + 1) / 2) // Normalize to [0,1]
		newStates[i] = pr.Resonators[i].GetTransmission(pr.Config.Wavelength)
	}
	pr.NodeStates = newStates

	// Reservoir to output
	if pr.OutputWeights == nil {
		return pr.NodeStates
	}

	output := make([]float64, len(pr.OutputWeights))
	for i := 0; i < len(pr.OutputWeights); i++ {
		sum := 0.0
		for j := 0; j < pr.NumNodes && j < len(pr.OutputWeights[i]); j++ {
			sum += pr.NodeStates[j] * pr.OutputWeights[i][j]
		}
		output[i] = sum
	}

	return output
}

// =============================================================================
// INTEGRATED EDGE-OPTICAL SYSTEM
// =============================================================================

// IronLatticeEdgeOpticalConfig configures integrated system
type IronLatticeEdgeOpticalConfig struct {
	EdgeConfig    *EdgeDeviceConfig
	PhotonicConfig *PhotonicConfig

	// Mode selection
	UseEdgeFeFET    bool
	UsePhotonic     bool
	UseHybrid       bool

	// Hybrid configuration
	PhotonicLayers  []int // Which layers use photonic compute
	EdgeLayers      []int // Which layers use FeFET CIM

	// Target metrics
	TargetLatencyMS float64
	TargetPowerMW   float64
	TargetAccuracy  float64
}

// DefaultIronLatticeEdgeOpticalConfig returns standard configuration
func DefaultIronLatticeEdgeOpticalConfig() *IronLatticeEdgeOpticalConfig {
	return &IronLatticeEdgeOpticalConfig{
		EdgeConfig:     FeFETEdgeDevice(),
		PhotonicConfig: FeFETPhotonicConfig(),
		UseEdgeFeFET:   true,
		UsePhotonic:    true,
		UseHybrid:      true,
		PhotonicLayers: []int{0, 1}, // First layers on photonic
		EdgeLayers:     []int{2, 3, 4}, // Later layers on FeFET
		TargetLatencyMS: 1.0,
		TargetPowerMW:   100.0,
		TargetAccuracy:  0.95,
	}
}

// IronLatticeEdgeOptical implements the integrated system
type IronLatticeEdgeOptical struct {
	Config *IronLatticeEdgeOpticalConfig

	// Components
	Compressor      *ModelCompressor
	Deployer        *EdgeDeployer
	PhotonicArray   *PhotonicCrossbar
	Reservoir       *PhotonicReservoir

	// Current model
	Model           *CompressedModel
	DeploymentPlan  *DeploymentPlan

	// Performance metrics
	TotalLatencyMS  float64
	TotalPowerMW    float64
	EnergyPerInfUJ  float64
	Accuracy        float64

	// Comparison metrics
	SpeedupVsGPU    float64
	EnergyVsGPU     float64
	SpeedupVsCPU    float64
	EnergyVsCPU     float64
}

// NewIronLatticeEdgeOptical creates a new integrated system
func NewIronLatticeEdgeOptical(config *IronLatticeEdgeOpticalConfig) *IronLatticeEdgeOptical {
	if config == nil {
		config = DefaultIronLatticeEdgeOpticalConfig()
	}

	system := &IronLatticeEdgeOptical{
		Config: config,
	}

	if config.UseEdgeFeFET {
		system.Compressor = NewModelCompressor(config.EdgeConfig)
	}

	if config.UsePhotonic {
		system.PhotonicArray = NewPhotonicCrossbar(
			config.PhotonicConfig,
			config.PhotonicConfig.ArraySize,
			config.PhotonicConfig.ArraySize,
		)
		system.Reservoir = NewPhotonicReservoir(config.PhotonicConfig, 64)
	}

	return system
}

// PrepareModel compresses and prepares model for deployment
func (sys *IronLatticeEdgeOptical) PrepareModel(profile *ModelProfile) {
	if sys.Compressor != nil {
		sys.Model = sys.Compressor.CompressModel(profile)
		sys.Deployer = NewEdgeDeployer(sys.Config.EdgeConfig, sys.Model)
		sys.DeploymentPlan = sys.Deployer.PlanDeployment()
	}
}

// RunInference performs inference using the hybrid system
func (sys *IronLatticeEdgeOptical) RunInference(input []float64) []float64 {
	var output []float64

	if sys.Config.UseHybrid {
		// First, photonic layers
		intermediate := input
		if sys.PhotonicArray != nil {
			intermediate = sys.PhotonicArray.MatVecMul(input)
		}

		// Then, reservoir processing (optional)
		if sys.Reservoir != nil {
			intermediate = sys.Reservoir.Process(intermediate)
		}

		output = intermediate
	} else if sys.Config.UsePhotonic && sys.PhotonicArray != nil {
		output = sys.PhotonicArray.MatVecMul(input)
	} else {
		// FeFET-only path (simulated)
		output = make([]float64, len(input))
		for i := range output {
			output[i] = input[i] * 0.9 // Placeholder
		}
	}

	return output
}

// CalculateMetrics computes overall system performance
func (sys *IronLatticeEdgeOptical) CalculateMetrics() {
	// Latency: combine photonic and FeFET
	photonicLatency := 0.0
	if sys.PhotonicArray != nil {
		photonicLatency = sys.PhotonicArray.LatencyNS / 1e6 // Convert to ms
	}

	fefetLatency := 0.0
	if sys.DeploymentPlan != nil {
		fefetLatency = sys.DeploymentPlan.InferenceLatencyMS
	}

	sys.TotalLatencyMS = photonicLatency + fefetLatency

	// Power: combine both
	photonicPower := 0.0
	if sys.PhotonicArray != nil {
		// Laser + modulators + detectors
		photonicPower = sys.Config.PhotonicConfig.LaserPowerMW * 2
	}

	fefetPower := 0.0
	if sys.DeploymentPlan != nil {
		fefetPower = sys.DeploymentPlan.InferencePowerMW
	}

	sys.TotalPowerMW = photonicPower + fefetPower

	// Energy per inference
	sys.EnergyPerInfUJ = sys.TotalPowerMW * sys.TotalLatencyMS

	// Comparison metrics
	// GPU baseline: ~10ms latency, ~100W power
	gpuLatency := 10.0  // ms
	gpuPower := 100000.0 // mW

	sys.SpeedupVsGPU = gpuLatency / sys.TotalLatencyMS
	sys.EnergyVsGPU = (gpuPower * gpuLatency) / (sys.TotalPowerMW * sys.TotalLatencyMS)

	// CPU baseline: ~100ms latency, ~15W power
	cpuLatency := 100.0 // ms
	cpuPower := 15000.0 // mW

	sys.SpeedupVsCPU = cpuLatency / sys.TotalLatencyMS
	sys.EnergyVsCPU = (cpuPower * cpuLatency) / (sys.TotalPowerMW * sys.TotalLatencyMS)
}

// GetStatistics returns system performance statistics
func (sys *IronLatticeEdgeOptical) GetStatistics() map[string]float64 {
	stats := make(map[string]float64)

	stats["total_latency_ms"] = sys.TotalLatencyMS
	stats["total_power_mw"] = sys.TotalPowerMW
	stats["energy_per_inf_uj"] = sys.EnergyPerInfUJ
	stats["speedup_vs_gpu"] = sys.SpeedupVsGPU
	stats["energy_vs_gpu"] = sys.EnergyVsGPU
	stats["speedup_vs_cpu"] = sys.SpeedupVsCPU
	stats["energy_vs_cpu"] = sys.EnergyVsCPU

	if sys.Compressor != nil {
		stats["compression_ratio"] = sys.Compressor.CompressionRatio
		stats["accuracy_drop"] = sys.Compressor.AccuracyDrop
	}

	if sys.PhotonicArray != nil {
		stats["photonic_tops_mm2"] = sys.PhotonicArray.ComputeDensityTOPS
		stats["photonic_tops_w"] = sys.PhotonicArray.EnergyEffTOPSW
	}

	if sys.DeploymentPlan != nil {
		stats["num_crossbars"] = float64(sys.DeploymentPlan.TotalCrossbars)
		stats["fefet_latency_ms"] = sys.DeploymentPlan.InferenceLatencyMS
	}

	return stats
}

// Benchmark runs performance comparison
func (sys *IronLatticeEdgeOptical) Benchmark() *EdgeOpticalBenchmark {
	sys.CalculateMetrics()

	return &EdgeOpticalBenchmark{
		SystemConfig:     "IronLattice Edge-Optical Hybrid",
		TotalLatencyMS:   sys.TotalLatencyMS,
		TotalPowerMW:     sys.TotalPowerMW,
		EnergyUJ:         sys.EnergyPerInfUJ,
		SpeedupVsGPU:     sys.SpeedupVsGPU,
		EnergyVsGPU:      sys.EnergyVsGPU,
		PhotonicTOPSMM2:  880.0,  // Literature value
		PhotonicTOPSW:    5.1,    // Literature value
		FeFETEnergyPJ:    0.5,    // Per MAC
		CompressionRatio: sys.Compressor.CompressionRatio,
	}
}

// EdgeOpticalBenchmark contains benchmark results
type EdgeOpticalBenchmark struct {
	SystemConfig     string
	TotalLatencyMS   float64
	TotalPowerMW     float64
	EnergyUJ         float64
	SpeedupVsGPU     float64
	EnergyVsGPU      float64
	PhotonicTOPSMM2  float64
	PhotonicTOPSW    float64
	FeFETEnergyPJ    float64
	CompressionRatio float64
}
