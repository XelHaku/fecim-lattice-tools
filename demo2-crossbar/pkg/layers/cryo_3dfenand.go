// Package layers provides cryogenic ferroelectric computing and 3D FeNAND
// array simulation for IronLattice CIM systems.
//
// Based on research findings:
// - Cryogenic HZO FeFET: 75 µC/cm² Pr at 4K, 6-8V memory window (Adv. Elec. Mat. 2024)
// - Enhanced linearity below 100K for in-memory computing (Frontiers Nano 2024)
// - VNW-FeFET: 25% memory window increase at 14K (IEEE 2024)
// - SK hynix 3D FeNAND: 256 levels/cell, 4000× density vs 2D (IEDM 2024)
// - Ultra-low power FeFET: 5-bit/cell, 96% power savings (Nature 2025)
// - Hybrid FeNAND paradigm for next-gen memory (arXiv Dec 2025)
//
// Reference: IronLattice Research Log Sections 282-283
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// CRYOGENIC FERROELECTRIC COMPUTING
// ============================================================================

// CryogenicConfig configures cryogenic ferroelectric operation
type CryogenicConfig struct {
	// Temperature settings
	TemperatureK       float64 // Operating temperature (Kelvin)
	AmbientThermalLoad float64 // Thermal load from environment (mW)

	// Material properties
	Material           string  // "hzo", "pzt", "bto"
	FilmThicknessNm    float64 // Ferroelectric film thickness (nm)

	// Cryogenic enhancements
	EnhancedLinearity  bool    // Enable enhanced linearity below 100K
	FrozenDefects      bool    // Model frozen charge defects

	// Quantum interface
	QuantumInterface   bool    // Enable quantum-classical interface
	QubitProximityMm   float64 // Distance to qubits (mm)
	MaxThermalBudgetUW float64 // Max thermal budget for quantum (µW)

	// Power constraints
	PowerBudgetMW      float64 // Total power budget (mW)
}

// DefaultCryogenicConfig returns default cryogenic configuration
func DefaultCryogenicConfig() *CryogenicConfig {
	return &CryogenicConfig{
		TemperatureK:       4.0,
		AmbientThermalLoad: 0.1,
		Material:           "hzo",
		FilmThicknessNm:    10.0,
		EnhancedLinearity:  true,
		FrozenDefects:      true,
		QuantumInterface:   true,
		QubitProximityMm:   10.0,
		MaxThermalBudgetUW: 100.0,
		PowerBudgetMW:      1.0,
	}
}

// TemperatureDependentFE models temperature-dependent ferroelectric properties
type TemperatureDependentFE struct {
	Config *CryogenicConfig

	// Material parameters (temperature-dependent)
	CoerciveFieldMVcm float64 // Coercive field (MV/cm)
	RemPolarization   float64 // Remanent polarization (µC/cm²)
	SatPolarization   float64 // Saturation polarization (µC/cm²)
	CurieTemperatureK float64 // Curie temperature (K)

	// Kinetic parameters
	ActivationEnergy  float64 // Activation energy for switching (eV)
	AttemptFrequency  float64 // Attempt frequency (Hz)

	// Current state
	Polarization      float64 // Current polarization state
	DomainFraction    float64 // Fraction of switched domains
}

// NewTemperatureDependentFE creates a temperature-dependent FE model
func NewTemperatureDependentFE(config *CryogenicConfig) *TemperatureDependentFE {
	tdfe := &TemperatureDependentFE{
		Config:           config,
		AttemptFrequency: 1e9, // 1 GHz typical
	}

	// Set material-specific parameters
	switch config.Material {
	case "hzo":
		tdfe.CurieTemperatureK = 600.0 // HZO Curie temp
		tdfe.ActivationEnergy = 0.5    // eV
		// Temperature-dependent Pr: increases at cryogenic
		tdfe.updateHZOParameters()
	case "pzt":
		tdfe.CurieTemperatureK = 650.0
		tdfe.ActivationEnergy = 0.3
		tdfe.updatePZTParameters()
	case "bto":
		tdfe.CurieTemperatureK = 393.0
		tdfe.ActivationEnergy = 0.4
		tdfe.updateBTOParameters()
	}

	return tdfe
}

// updateHZOParameters updates HZO parameters for current temperature
func (tdfe *TemperatureDependentFE) updateHZOParameters() {
	T := tdfe.Config.TemperatureK
	Tc := tdfe.CurieTemperatureK

	// Landau theory: Ps ∝ (Tc - T)^0.5 below Tc
	if T < Tc {
		// Room temperature values as baseline
		PsRT := 45.0  // µC/cm² at 300K
		PrRT := 30.0  // µC/cm² at 300K
		EcRT := 1.0   // MV/cm at 300K

		// Temperature scaling
		tempFactor := math.Sqrt((Tc - T) / (Tc - 300.0))

		// Below 100K: significant enhancement (from papers)
		if T < 100.0 {
			// Record Pr of 75 µC/cm² reported at cryogenic
			tdfe.RemPolarization = PrRT * tempFactor * 1.5 // Enhanced
			tdfe.SatPolarization = PsRT * tempFactor * 1.5
			// Memory window increases 25% at 14K
			tdfe.CoerciveFieldMVcm = EcRT * (1.0 + 0.25*(100.0-T)/100.0)
		} else {
			tdfe.RemPolarization = PrRT * tempFactor
			tdfe.SatPolarization = PsRT * tempFactor
			tdfe.CoerciveFieldMVcm = EcRT
		}
	}
}

// updatePZTParameters updates PZT parameters
func (tdfe *TemperatureDependentFE) updatePZTParameters() {
	T := tdfe.Config.TemperatureK
	Tc := tdfe.CurieTemperatureK

	if T < Tc {
		tempFactor := math.Sqrt((Tc - T) / (Tc - 300.0))
		tdfe.RemPolarization = 25.0 * tempFactor
		tdfe.SatPolarization = 35.0 * tempFactor
		tdfe.CoerciveFieldMVcm = 0.5
	}
}

// updateBTOParameters updates BTO parameters
func (tdfe *TemperatureDependentFE) updateBTOParameters() {
	T := tdfe.Config.TemperatureK
	Tc := tdfe.CurieTemperatureK

	if T < Tc {
		tempFactor := math.Sqrt((Tc - T) / (Tc - 300.0))
		tdfe.RemPolarization = 15.0 * tempFactor
		tdfe.SatPolarization = 20.0 * tempFactor
		tdfe.CoerciveFieldMVcm = 0.3
	}
}

// SwitchingProbability calculates switching probability at current temperature
func (tdfe *TemperatureDependentFE) SwitchingProbability(field float64, pulseTime float64) float64 {
	T := tdfe.Config.TemperatureK
	Ec := tdfe.CoerciveFieldMVcm
	Ea := tdfe.ActivationEnergy

	// Boltzmann constant
	kB := 8.617e-5 // eV/K

	// Below ~10K: quantum tunneling dominates over thermal activation
	if T < 10.0 {
		// Quantum tunneling rate (simplified)
		tunnelFactor := math.Exp(-Ea / (0.01 + T*kB)) // Avoid divide by zero
		fieldFactor := math.Abs(field) / Ec
		return math.Min(1.0, tunnelFactor*fieldFactor*pulseTime*tdfe.AttemptFrequency)
	}

	// Thermal activation (Arrhenius)
	thermalFactor := math.Exp(-Ea / (kB * T))

	// Field-assisted switching
	fieldFactor := 1.0
	if math.Abs(field) > Ec {
		fieldFactor = math.Pow(math.Abs(field)/Ec, 2.0)
	}

	return math.Min(1.0, thermalFactor*fieldFactor*pulseTime*tdfe.AttemptFrequency)
}

// CryogenicFeFET represents a FeFET operating at cryogenic temperatures
type CryogenicFeFET struct {
	Config    *CryogenicConfig
	FEModel   *TemperatureDependentFE

	// Device parameters
	ChannelLength     float64 // nm
	ChannelWidth      float64 // nm
	IsVerticalNanowire bool   // VNW structure

	// Electrical characteristics
	VthHigh           float64 // High Vth state (V)
	VthLow            float64 // Low Vth state (V)
	MemoryWindow      float64 // Vth difference (V)

	// Analog states
	CurrentVth        float64 // Current threshold voltage
	ConductanceStates int     // Number of analog levels
	CurrentLevel      int     // Current analog level

	// Reliability
	WriteCycles       int
	RetentionTime     float64 // seconds at current temp
}

// NewCryogenicFeFET creates a cryogenic FeFET
func NewCryogenicFeFET(config *CryogenicConfig, isVNW bool) *CryogenicFeFET {
	cfet := &CryogenicFeFET{
		Config:            config,
		FEModel:           NewTemperatureDependentFE(config),
		IsVerticalNanowire: isVNW,
		ConductanceStates: 256, // From SK hynix IEDM 2024
	}

	if isVNW {
		cfet.ChannelLength = 50.0  // nm
		cfet.ChannelWidth = 20.0   // nm (diameter)
	} else {
		cfet.ChannelLength = 100.0
		cfet.ChannelWidth = 100.0
	}

	// Calculate memory window based on temperature
	cfet.updateMemoryWindow()

	return cfet
}

// updateMemoryWindow calculates temperature-dependent memory window
func (cfet *CryogenicFeFET) updateMemoryWindow() {
	T := cfet.Config.TemperatureK

	// Base memory window at 300K
	baseWindow := 2.3 // V (from papers)

	// Temperature enhancement
	if T < 100.0 {
		// 25% increase at 14K for VNW-FeFET
		if cfet.IsVerticalNanowire {
			enhancement := 1.0 + 0.25*(100.0-T)/100.0
			cfet.MemoryWindow = baseWindow * enhancement
		} else {
			// Standard enhancement
			cfet.MemoryWindow = baseWindow * (1.0 + 0.15*(100.0-T)/100.0)
		}

		// Below 100K: can use higher amplitude pulses (6-8V window)
		if T < 50.0 {
			cfet.MemoryWindow = math.Min(8.0, cfet.MemoryWindow*1.5)
		}
	} else {
		cfet.MemoryWindow = baseWindow
	}

	cfet.VthLow = -cfet.MemoryWindow / 2
	cfet.VthHigh = cfet.MemoryWindow / 2
	cfet.CurrentVth = 0.0
}

// ProgramAnalog programs to a specific analog level
func (cfet *CryogenicFeFET) ProgramAnalog(targetLevel int) error {
	if targetLevel < 0 || targetLevel >= cfet.ConductanceStates {
		return fmt.Errorf("invalid level: %d (max %d)", targetLevel, cfet.ConductanceStates-1)
	}

	// Calculate target Vth
	levelFraction := float64(targetLevel) / float64(cfet.ConductanceStates-1)
	targetVth := cfet.VthLow + levelFraction*cfet.MemoryWindow

	// Enhanced linearity at cryogenic temperatures
	if cfet.Config.EnhancedLinearity && cfet.Config.TemperatureK < 100.0 {
		// Linear update (improved at low temp)
		cfet.CurrentVth = targetVth
	} else {
		// Non-linear update at room temp
		delta := targetVth - cfet.CurrentVth
		nonlinearity := 0.8 // Asymmetric potentiation/depression
		if delta > 0 {
			cfet.CurrentVth += delta * nonlinearity
		} else {
			cfet.CurrentVth += delta * (2.0 - nonlinearity)
		}
	}

	cfet.CurrentLevel = targetLevel
	cfet.WriteCycles++

	return nil
}

// ReadConductance reads the current conductance state
func (cfet *CryogenicFeFET) ReadConductance() float64 {
	// Normalize Vth to conductance (0 to 1)
	normalized := (cfet.CurrentVth - cfet.VthLow) / cfet.MemoryWindow

	// Add read noise (reduced at cryogenic)
	noiseLevel := 0.01 // 1% at 300K
	if cfet.Config.TemperatureK < 100.0 {
		noiseLevel *= cfet.Config.TemperatureK / 100.0 // Scales with T
	}

	noise := rand.NormFloat64() * noiseLevel
	return math.Max(0, math.Min(1, normalized+noise))
}

// CryogenicCrossbar implements a crossbar array for cryogenic operation
type CryogenicCrossbar struct {
	Config    *CryogenicConfig
	Rows, Cols int

	// FeFET array
	Cells     [][]*CryogenicFeFET

	// Power tracking
	TotalPower    float64 // Current power consumption (mW)
	ThermalOutput float64 // Heat generated (mW)

	// Quantum interface
	QuantumReady  bool // Safe for quantum proximity
}

// NewCryogenicCrossbar creates a cryogenic crossbar
func NewCryogenicCrossbar(config *CryogenicConfig, rows, cols int) *CryogenicCrossbar {
	cells := make([][]*CryogenicFeFET, rows)
	for i := 0; i < rows; i++ {
		cells[i] = make([]*CryogenicFeFET, cols)
		for j := 0; j < cols; j++ {
			cells[i][j] = NewCryogenicFeFET(config, true) // VNW for density
		}
	}

	return &CryogenicCrossbar{
		Config:   config,
		Rows:     rows,
		Cols:     cols,
		Cells:    cells,
	}
}

// ComputeMVM performs matrix-vector multiplication
func (cc *CryogenicCrossbar) ComputeMVM(input []float64) []float64 {
	output := make([]float64, cc.Cols)

	// Track power
	cc.TotalPower = 0

	for j := 0; j < cc.Cols; j++ {
		sum := 0.0
		for i := 0; i < cc.Rows && i < len(input); i++ {
			weight := cc.Cells[i][j].ReadConductance()
			sum += input[i] * (2*weight - 1) // Scale to [-1, 1]

			// Power per MAC (reduced at cryogenic due to lower voltage)
			powerPerMAC := 0.1 // pJ at 4K (vs ~1 pJ at 300K)
			cc.TotalPower += powerPerMAC / 1000.0 // Convert to mW
		}
		output[j] = sum
	}

	// Check thermal budget for quantum
	cc.ThermalOutput = cc.TotalPower
	cc.QuantumReady = cc.ThermalOutput < cc.Config.MaxThermalBudgetUW/1000.0

	return output
}

// ============================================================================
// 3D FeNAND ARRAY ARCHITECTURE
// ============================================================================

// FeNANDConfig configures 3D FeNAND array
type FeNANDConfig struct {
	// Architecture
	Architecture      string  // "vc_fenand", "hc_fenand", "fe_and"
	NumLayers         int     // Number of vertical layers
	NumStrings        int     // Number of NAND strings
	CellsPerString    int     // Cells per vertical string

	// Cell parameters
	LevelsPerCell     int     // Multi-level capability (256 from SK hynix)
	BitsPerCell       int     // Derived from levels

	// Material
	FEMaterial        string  // "hzo", "sizo"
	FEThicknessNm     float64 // Ferroelectric layer thickness

	// Operating parameters
	PassVoltage       float64 // Pass voltage for unselected cells (V)
	ProgramVoltage    float64 // Programming voltage (V)
	ReadVoltage       float64 // Read voltage (V)

	// Power
	PowerSavingsRatio float64 // vs conventional NAND (0.96 from Nature 2025)
}

// DefaultFeNANDConfig returns default 3D FeNAND configuration
func DefaultFeNANDConfig() *FeNANDConfig {
	return &FeNANDConfig{
		Architecture:      "vc_fenand",
		NumLayers:         128,
		NumStrings:        1024,
		CellsPerString:    128,
		LevelsPerCell:     256, // From SK hynix IEDM 2024
		BitsPerCell:       8,   // log2(256)
		FEMaterial:        "hzo",
		FEThicknessNm:     10.0,
		PassVoltage:       0.0,  // Near-zero pass voltage from Nature 2025
		ProgramVoltage:    4.0,
		ReadVoltage:       1.0,
		PowerSavingsRatio: 0.96,
	}
}

// FeNANDCell represents a single 3D FeNAND cell
type FeNANDCell struct {
	// Position
	Layer         int
	StringIdx     int
	CellIdx       int

	// State
	PolarizationState float64 // -1 to 1
	AnalogLevel       int     // 0 to LevelsPerCell-1
	ThresholdVoltage  float64 // Vth

	// Reliability
	ProgramCycles     int
	DisturbCount      int
}

// NewFeNANDCell creates a new FeNAND cell
func NewFeNANDCell(layer, stringIdx, cellIdx int) *FeNANDCell {
	return &FeNANDCell{
		Layer:             layer,
		StringIdx:         stringIdx,
		CellIdx:           cellIdx,
		PolarizationState: 0.0,
		AnalogLevel:       128, // Mid-level
		ThresholdVoltage:  0.0,
	}
}

// FeNANDString represents a vertical NAND string
type FeNANDString struct {
	Config    *FeNANDConfig
	StringIdx int

	// Cells in string (vertical stack)
	Cells     []*FeNANDCell

	// String-level state
	Selected  bool
	BitLine   float64 // Bit line voltage
	SourceLine float64 // Source line voltage
}

// NewFeNANDString creates a vertical NAND string
func NewFeNANDString(config *FeNANDConfig, stringIdx int) *FeNANDString {
	cells := make([]*FeNANDCell, config.CellsPerString)
	for i := 0; i < config.CellsPerString; i++ {
		layer := i % config.NumLayers
		cells[i] = NewFeNANDCell(layer, stringIdx, i)
	}

	return &FeNANDString{
		Config:    config,
		StringIdx: stringIdx,
		Cells:     cells,
	}
}

// ReadCell reads a specific cell in the string
func (fs *FeNANDString) ReadCell(cellIdx int) (int, error) {
	if cellIdx >= len(fs.Cells) {
		return 0, fmt.Errorf("cell index out of range")
	}

	cell := fs.Cells[cellIdx]

	// Apply pass voltage to unselected cells
	// With near-zero pass voltage (Nature 2025), minimal disturb
	for i, c := range fs.Cells {
		if i != cellIdx {
			// Check disturb from pass voltage
			if fs.Config.PassVoltage > 0.5 {
				c.DisturbCount++
			}
		}
	}

	return cell.AnalogLevel, nil
}

// ProgramCell programs a specific cell using ISPP
func (fs *FeNANDString) ProgramCell(cellIdx int, targetLevel int) error {
	if cellIdx >= len(fs.Cells) {
		return fmt.Errorf("cell index out of range")
	}

	cell := fs.Cells[cellIdx]

	// Incremental Step Pulse Programming (ISPP)
	// From the papers: achieves precise analog levels
	currentLevel := cell.AnalogLevel
	stepSize := 1

	for currentLevel != targetLevel {
		if currentLevel < targetLevel {
			currentLevel = min(currentLevel+stepSize, targetLevel)
		} else {
			currentLevel = max(currentLevel-stepSize, targetLevel)
		}
		cell.ProgramCycles++
	}

	cell.AnalogLevel = targetLevel
	cell.PolarizationState = 2.0*float64(targetLevel)/float64(fs.Config.LevelsPerCell-1) - 1.0
	cell.ThresholdVoltage = cell.PolarizationState * 2.0 // Scale to Vth range

	return nil
}

// FeNAND3DArray represents a complete 3D FeNAND array
type FeNAND3DArray struct {
	Config     *FeNANDConfig

	// Array of strings
	Strings    []*FeNANDString

	// Word lines (horizontal)
	WordLines  []float64

	// Statistics
	TotalCapacityBits int64
	DensityBitsMm2    float64

	// CIM capabilities
	CIMEnabled        bool
	MACOperations     int64
	ComputeEfficiency float64 // TOPS/mm²
}

// NewFeNAND3DArray creates a 3D FeNAND array
func NewFeNAND3DArray(config *FeNANDConfig) *FeNAND3DArray {
	strings := make([]*FeNANDString, config.NumStrings)
	for i := 0; i < config.NumStrings; i++ {
		strings[i] = NewFeNANDString(config, i)
	}

	array := &FeNAND3DArray{
		Config:    config,
		Strings:   strings,
		WordLines: make([]float64, config.CellsPerString),
		CIMEnabled: true,
	}

	// Calculate capacity
	cellsTotal := int64(config.NumStrings) * int64(config.CellsPerString)
	array.TotalCapacityBits = cellsTotal * int64(config.BitsPerCell)

	// Density: ~4000x improvement vs 2D (SK hynix)
	// Assuming 1mm² chip area for reference
	array.DensityBitsMm2 = float64(array.TotalCapacityBits) // bits/mm²

	// Compute efficiency: 1000x vs 2D arrays
	array.ComputeEfficiency = 1000.0 // TOPS/mm² (from SK hynix)

	return array
}

// ComputeMAC performs analog MAC operation using FeNAND
func (fa *FeNAND3DArray) ComputeMAC(inputVector []float64, weightWordLine int) ([]float64, error) {
	if weightWordLine >= fa.Config.CellsPerString {
		return nil, fmt.Errorf("word line out of range")
	}

	output := make([]float64, len(fa.Strings))

	for i, str := range fa.Strings {
		if i >= len(inputVector) {
			break
		}

		// Read weight from specified word line
		level, _ := str.ReadCell(weightWordLine)

		// Normalize weight to [-1, 1]
		weight := 2.0*float64(level)/float64(fa.Config.LevelsPerCell-1) - 1.0

		// MAC operation
		output[i] = inputVector[i] * weight

		fa.MACOperations++
	}

	return output, nil
}

// ComputeLayerMVM performs full layer MVM using multiple word lines
func (fa *FeNAND3DArray) ComputeLayerMVM(input []float64, numOutputs int) []float64 {
	output := make([]float64, numOutputs)

	for outIdx := 0; outIdx < numOutputs && outIdx < fa.Config.CellsPerString; outIdx++ {
		sum := 0.0

		for inIdx, str := range fa.Strings {
			if inIdx >= len(input) {
				break
			}

			level, _ := str.ReadCell(outIdx)
			weight := 2.0*float64(level)/float64(fa.Config.LevelsPerCell-1) - 1.0
			sum += input[inIdx] * weight
		}

		output[outIdx] = sum
		fa.MACOperations += int64(len(input))
	}

	return output
}

// BayesianNNSupport implements Bayesian neural network on FeNAND
type BayesianNNSupport struct {
	Array     *FeNAND3DArray

	// Bayesian parameters
	WeightMeans    [][]float64 // Mean weights
	WeightVariances [][]float64 // Weight variances

	// Sampling
	NumSamples     int
	SampledWeights [][][]float64
}

// NewBayesianNNSupport creates Bayesian NN support
func NewBayesianNNSupport(array *FeNAND3DArray, rows, cols int) *BayesianNNSupport {
	means := make([][]float64, rows)
	variances := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		means[i] = make([]float64, cols)
		variances[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			means[i][j] = rand.NormFloat64() * 0.1
			variances[i][j] = 0.01 // Initial variance
		}
	}

	return &BayesianNNSupport{
		Array:           array,
		WeightMeans:     means,
		WeightVariances: variances,
		NumSamples:      10,
	}
}

// SampleWeights samples weights from posterior
func (bnn *BayesianNNSupport) SampleWeights() [][]float64 {
	rows := len(bnn.WeightMeans)
	cols := len(bnn.WeightMeans[0])

	sampled := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		sampled[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Sample from Gaussian posterior
			mean := bnn.WeightMeans[i][j]
			std := math.Sqrt(bnn.WeightVariances[i][j])
			sampled[i][j] = mean + rand.NormFloat64()*std
		}
	}

	return sampled
}

// ProgramSampledWeights programs sampled weights to array
func (bnn *BayesianNNSupport) ProgramSampledWeights(sampled [][]float64) error {
	levels := bnn.Array.Config.LevelsPerCell

	for i, row := range sampled {
		if i >= len(bnn.Array.Strings) {
			break
		}

		for j, w := range row {
			if j >= bnn.Array.Config.CellsPerString {
				break
			}

			// Convert weight to level
			normalized := (w + 1.0) / 2.0 // [-1,1] -> [0,1]
			level := int(normalized * float64(levels-1))
			level = max(0, min(levels-1, level))

			bnn.Array.Strings[i].ProgramCell(j, level)
		}
	}

	return nil
}

// ============================================================================
// INTEGRATED IRONLATTICE CRYOGENIC 3D SYSTEM
// ============================================================================

// IronLatticeCryo3D integrates cryogenic computing with 3D FeNAND
type IronLatticeCryo3D struct {
	// Configuration
	CryoConfig   *CryogenicConfig
	FeNANDConfig *FeNANDConfig

	// Components
	CryoCrossbar *CryogenicCrossbar
	FeNANDArray  *FeNAND3DArray
	BayesianNN   *BayesianNNSupport

	// Quantum interface
	QuantumControlReady bool
	QuantumGates        []QuantumGateControl

	// Statistics
	TotalInferences int64
	TotalEnergy     float64 // pJ
	AverageAccuracy float64
}

// QuantumGateControl represents control interface for quantum gates
type QuantumGateControl struct {
	GateID        int
	ControlVoltage float64
	Duration      float64 // ns
	Fidelity      float64
}

// NewIronLatticeCryo3D creates the integrated system
func NewIronLatticeCryo3D(cryoConfig *CryogenicConfig, fenandConfig *FeNANDConfig) *IronLatticeCryo3D {
	system := &IronLatticeCryo3D{
		CryoConfig:   cryoConfig,
		FeNANDConfig: fenandConfig,
		CryoCrossbar: NewCryogenicCrossbar(cryoConfig, 64, 64),
		FeNANDArray:  NewFeNAND3DArray(fenandConfig),
		QuantumGates: make([]QuantumGateControl, 0),
	}

	// Initialize Bayesian NN support
	system.BayesianNN = NewBayesianNNSupport(system.FeNANDArray, 64, 64)

	// Check quantum readiness
	system.QuantumControlReady = cryoConfig.QuantumInterface &&
		cryoConfig.TemperatureK < 20.0

	return system
}

// RunCryogenicInference runs inference on cryogenic crossbar
func (ilc *IronLatticeCryo3D) RunCryogenicInference(input []float64) ([]float64, error) {
	// Run MVM on cryogenic crossbar
	output := ilc.CryoCrossbar.ComputeMVM(input)

	// Check thermal budget
	if !ilc.CryoCrossbar.QuantumReady {
		return output, fmt.Errorf("thermal budget exceeded for quantum proximity")
	}

	ilc.TotalInferences++

	// Estimate energy (much lower at cryogenic)
	energyPerMAC := 0.1 // pJ at 4K
	ilc.TotalEnergy += float64(len(input)*len(output)) * energyPerMAC

	return output, nil
}

// RunHighDensityInference runs inference on 3D FeNAND
func (ilc *IronLatticeCryo3D) RunHighDensityInference(input []float64, outputSize int) []float64 {
	output := ilc.FeNANDArray.ComputeLayerMVM(input, outputSize)

	ilc.TotalInferences++

	// Energy with 96% power savings
	baseEnergy := 1.0 // pJ per MAC baseline
	actualEnergy := baseEnergy * (1.0 - ilc.FeNANDConfig.PowerSavingsRatio)
	ilc.TotalEnergy += float64(len(input)*outputSize) * actualEnergy

	return output
}

// RunBayesianInference runs Bayesian inference with uncertainty
func (ilc *IronLatticeCryo3D) RunBayesianInference(input []float64, numSamples int) ([]float64, []float64) {
	outputSize := ilc.BayesianNN.Array.Config.CellsPerString
	allOutputs := make([][]float64, numSamples)

	// Run multiple forward passes with sampled weights
	for s := 0; s < numSamples; s++ {
		// Sample weights
		sampled := ilc.BayesianNN.SampleWeights()

		// Program to array
		ilc.BayesianNN.ProgramSampledWeights(sampled)

		// Run inference
		allOutputs[s] = ilc.RunHighDensityInference(input, outputSize)
	}

	// Compute mean and variance
	meanOutput := make([]float64, outputSize)
	varOutput := make([]float64, outputSize)

	for i := 0; i < outputSize; i++ {
		sum := 0.0
		for s := 0; s < numSamples; s++ {
			if i < len(allOutputs[s]) {
				sum += allOutputs[s][i]
			}
		}
		meanOutput[i] = sum / float64(numSamples)

		// Variance
		varSum := 0.0
		for s := 0; s < numSamples; s++ {
			if i < len(allOutputs[s]) {
				diff := allOutputs[s][i] - meanOutput[i]
				varSum += diff * diff
			}
		}
		varOutput[i] = varSum / float64(numSamples)
	}

	return meanOutput, varOutput
}

// AddQuantumGateControl adds a quantum gate control interface
func (ilc *IronLatticeCryo3D) AddQuantumGateControl(gateID int, voltage float64) error {
	if !ilc.QuantumControlReady {
		return fmt.Errorf("system not ready for quantum control (temp: %.1fK)", ilc.CryoConfig.TemperatureK)
	}

	gate := QuantumGateControl{
		GateID:         gateID,
		ControlVoltage: voltage,
		Duration:       10.0, // ns typical
		Fidelity:       0.999,
	}

	ilc.QuantumGates = append(ilc.QuantumGates, gate)
	return nil
}

// GetStatistics returns system statistics
func (ilc *IronLatticeCryo3D) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["temperature_K"] = ilc.CryoConfig.TemperatureK
	stats["quantum_ready"] = ilc.QuantumControlReady
	stats["total_inferences"] = ilc.TotalInferences
	stats["total_energy_pJ"] = ilc.TotalEnergy
	stats["fenand_capacity_bits"] = ilc.FeNANDArray.TotalCapacityBits
	stats["fenand_levels_per_cell"] = ilc.FeNANDConfig.LevelsPerCell
	stats["fenand_compute_efficiency_TOPS_mm2"] = ilc.FeNANDArray.ComputeEfficiency
	stats["fenand_power_savings"] = ilc.FeNANDConfig.PowerSavingsRatio
	stats["cryo_memory_window_V"] = ilc.CryoCrossbar.Cells[0][0].MemoryWindow
	stats["num_quantum_gates"] = len(ilc.QuantumGates)

	return stats
}

// Preset configurations
func IronLatticeCryo3DPreset(scenario string) (*CryogenicConfig, *FeNANDConfig) {
	switch scenario {
	case "quantum_control":
		// For quantum computing control electronics
		return &CryogenicConfig{
				TemperatureK:       4.0,
				Material:           "hzo",
				EnhancedLinearity:  true,
				QuantumInterface:   true,
				QubitProximityMm:   5.0,
				MaxThermalBudgetUW: 50.0,
			},
			&FeNANDConfig{
				Architecture:      "vc_fenand",
				NumLayers:         64,
				NumStrings:        256,
				CellsPerString:    64,
				LevelsPerCell:     64,
				PowerSavingsRatio: 0.96,
			}

	case "space_computing":
		// For space applications (radiation-hard, wide temp range)
		return &CryogenicConfig{
				TemperatureK:       77.0, // Liquid nitrogen
				Material:           "hzo",
				EnhancedLinearity:  true,
				FrozenDefects:      false, // Higher temp
				QuantumInterface:   false,
			},
			&FeNANDConfig{
				Architecture:      "vc_fenand",
				NumLayers:         128,
				NumStrings:        512,
				CellsPerString:    128,
				LevelsPerCell:     16, // Conservative for reliability
				PowerSavingsRatio: 0.90,
			}

	case "high_density_ai":
		// Maximum density for AI inference
		return DefaultCryogenicConfig(),
			&FeNANDConfig{
				Architecture:      "vc_fenand",
				NumLayers:         256,
				NumStrings:        4096,
				CellsPerString:    256,
				LevelsPerCell:     256, // SK hynix level
				BitsPerCell:       8,
				PassVoltage:       0.0, // Near-zero
				PowerSavingsRatio: 0.96,
			}

	case "bayesian_nn":
		// For Bayesian neural networks
		return &CryogenicConfig{
				TemperatureK:      77.0,
				Material:          "hzo",
				EnhancedLinearity: true,
			},
			&FeNANDConfig{
				Architecture:      "vc_fenand",
				NumLayers:         128,
				NumStrings:        1024,
				CellsPerString:    128,
				LevelsPerCell:     256,
				PowerSavingsRatio: 0.96,
			}

	default:
		return DefaultCryogenicConfig(), DefaultFeNANDConfig()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
