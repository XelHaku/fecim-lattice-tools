// power_3d.go - Power/Energy Estimation and 3D Crossbar Architecture for CIM
//
// This module implements:
// - Comprehensive power breakdown (ADC, DAC, crossbar, peripherals)
// - Energy per MAC operation estimation
// - TOPS/W efficiency calculations
// - 3D FeNAND/crossbar array modeling
// - Vertical and horizontal NAND architectures
// - Layer-by-layer computation for 3D neural networks
//
// Based on research findings:
// - ADC/DAC account for ~10% of CIM power
// - Crossbar activation computation: ~83%
// - Data movement: ~7%
// - Target: <10 fJ per synaptic operation
// - State-of-art: 687.5 TOPS/W (SRAM CIM, 2024)
//
// References:
// - NeuroSim V1.5 (arXiv 2505.02314)
// - 3D Ferroelectric Memory Architectures (arXiv 2504.09713)
// - ISSCC 2024 CIM Macro Designs

package layers

import (
	"math"
)

// ================== Power Estimation ==================

// PowerConfig configures power estimation parameters
type PowerConfig struct {
	// Technology parameters
	TechNodeNm      float64 // Technology node (nm)
	VDD             float64 // Supply voltage (V)
	Temperature     float64 // Operating temperature (°C)

	// ADC parameters
	ADCBits         int     // ADC resolution
	ADCEnergyPerConv float64 // Energy per conversion (pJ)
	ADCSamplingRate float64 // Sampling rate (MS/s)
	ADCColumns      int     // Number of ADC columns

	// DAC parameters
	DACBits         int     // DAC resolution
	DACEnergyPerConv float64 // Energy per conversion (pJ)
	DACRows         int     // Number of DAC rows

	// Crossbar parameters
	CellReadEnergy  float64 // Energy per cell read (fJ)
	CellWriteEnergy float64 // Energy per cell write (pJ)
	LineResistance  float64 // Line resistance (Ohm)
	CellConductance float64 // Average cell conductance (S)

	// Peripheral parameters
	ShiftAddEnergy  float64 // Shift-add energy (fJ/bit)
	BufferEnergy    float64 // Buffer read/write (fJ/bit)
	ControlEnergy   float64 // Control logic (pJ/cycle)

	// Memory parameters
	LeakagePower    float64 // Static leakage (mW)
	RefreshEnergy   float64 // Refresh energy if needed (pJ)
}

// DefaultPowerConfig returns typical CIM power parameters
func DefaultPowerConfig() *PowerConfig {
	return &PowerConfig{
		TechNodeNm:       40,      // 40nm
		VDD:              1.0,     // 1.0V
		Temperature:      25,      // Room temp

		ADCBits:          8,
		ADCEnergyPerConv: 0.5,     // 0.5 pJ per conversion
		ADCSamplingRate:  100,     // 100 MS/s
		ADCColumns:       64,

		DACBits:          8,
		DACEnergyPerConv: 0.2,     // 0.2 pJ per conversion
		DACRows:          64,

		CellReadEnergy:   1.0,     // 1 fJ per cell read
		CellWriteEnergy:  100,     // 100 pJ per cell write
		LineResistance:   10,      // 10 Ohm
		CellConductance:  10e-6,   // 10 µS

		ShiftAddEnergy:   0.1,     // 0.1 fJ/bit
		BufferEnergy:     0.5,     // 0.5 fJ/bit
		ControlEnergy:    1.0,     // 1 pJ/cycle

		LeakagePower:     0.1,     // 0.1 mW
		RefreshEnergy:    0,       // No refresh for NVM
	}
}

// PowerBreakdown contains detailed power analysis
type PowerBreakdown struct {
	// Component energies (pJ per operation)
	ADCEnergy       float64
	DACEnergy       float64
	CrossbarEnergy  float64
	PeripheralEnergy float64
	DataMovement    float64
	LeakageEnergy   float64
	TotalEnergy     float64

	// Percentages
	ADCPercent      float64
	DACPercent      float64
	CrossbarPercent float64
	PeripheralPercent float64
	DataMovePercent float64

	// Efficiency metrics
	EnergyPerMAC    float64 // fJ/MAC
	TOPSW           float64 // TOPS/W
	OpsPerJoule     float64 // Operations per Joule
}

// PowerEstimator estimates CIM power consumption
type PowerEstimator struct {
	Config    *PowerConfig
	ArrayRows int
	ArrayCols int
	NumArrays int
	Frequency float64 // Operating frequency (MHz)

	// Running statistics
	TotalOps      int64
	TotalEnergy   float64 // pJ
	PeakPower     float64 // mW
	AveragePower  float64 // mW
}

// NewPowerEstimator creates a power estimation module
func NewPowerEstimator(config *PowerConfig, rows, cols, numArrays int, freqMHz float64) *PowerEstimator {
	return &PowerEstimator{
		Config:    config,
		ArrayRows: rows,
		ArrayCols: cols,
		NumArrays: numArrays,
		Frequency: freqMHz,
		TotalOps:  0,
		TotalEnergy: 0,
		PeakPower:  0,
		AveragePower: 0,
	}
}

// EstimateMVMEnergy estimates energy for matrix-vector multiplication
func (pe *PowerEstimator) EstimateMVMEnergy(inputSparsity, weightSparsity float64) *PowerBreakdown {
	cfg := pe.Config
	rows := pe.ArrayRows
	cols := pe.ArrayCols

	// Effective operations considering sparsity
	effectiveOps := float64(rows * cols) * (1 - inputSparsity) * (1 - weightSparsity)

	// ADC energy: one conversion per column per input cycle
	numADCConversions := float64(cols)
	adcEnergy := numADCConversions * cfg.ADCEnergyPerConv

	// DAC energy: one conversion per row
	numDACConversions := float64(rows) * (1 - inputSparsity)
	dacEnergy := numDACConversions * cfg.DACEnergyPerConv

	// Crossbar energy: cell access for active inputs
	crossbarEnergy := effectiveOps * cfg.CellReadEnergy / 1000 // Convert fJ to pJ

	// IR drop penalty (increases with array size and current)
	irDropFactor := 1.0 + cfg.LineResistance*cfg.CellConductance*float64(rows)/10
	crossbarEnergy *= irDropFactor

	// Peripheral energy: shift-add for bit-serial, buffers
	bitsPerMAC := float64(cfg.ADCBits + cfg.DACBits)
	shiftAddEnergy := effectiveOps * bitsPerMAC * cfg.ShiftAddEnergy / 1000
	bufferEnergy := float64(rows+cols) * float64(cfg.DACBits) * cfg.BufferEnergy / 1000
	peripheralEnergy := shiftAddEnergy + bufferEnergy + cfg.ControlEnergy

	// Data movement (input loading, output transfer)
	inputBits := float64(rows) * float64(cfg.DACBits) * (1 - inputSparsity)
	outputBits := float64(cols) * float64(cfg.ADCBits)
	dataMovement := (inputBits + outputBits) * cfg.BufferEnergy / 1000

	// Leakage during operation
	cycleTime := 1.0 / pe.Frequency * 1000 // ns
	leakageEnergy := cfg.LeakagePower * cycleTime / 1e6 // pJ

	// Total energy
	totalEnergy := adcEnergy + dacEnergy + crossbarEnergy + peripheralEnergy + dataMovement + leakageEnergy

	// Calculate percentages
	breakdown := &PowerBreakdown{
		ADCEnergy:       adcEnergy,
		DACEnergy:       dacEnergy,
		CrossbarEnergy:  crossbarEnergy,
		PeripheralEnergy: peripheralEnergy,
		DataMovement:    dataMovement,
		LeakageEnergy:   leakageEnergy,
		TotalEnergy:     totalEnergy,

		ADCPercent:      adcEnergy / totalEnergy * 100,
		DACPercent:      dacEnergy / totalEnergy * 100,
		CrossbarPercent: crossbarEnergy / totalEnergy * 100,
		PeripheralPercent: peripheralEnergy / totalEnergy * 100,
		DataMovePercent: dataMovement / totalEnergy * 100,
	}

	// Efficiency metrics
	numMACs := float64(rows * cols)
	breakdown.EnergyPerMAC = totalEnergy * 1000 / numMACs // fJ/MAC
	breakdown.OpsPerJoule = numMACs / (totalEnergy * 1e-12)
	breakdown.TOPSW = breakdown.OpsPerJoule / 1e12

	// Update running statistics
	pe.TotalOps += int64(numMACs)
	pe.TotalEnergy += totalEnergy

	// Power calculation
	power := totalEnergy * pe.Frequency / 1e6 // mW
	if power > pe.PeakPower {
		pe.PeakPower = power
	}
	pe.AveragePower = pe.TotalEnergy * pe.Frequency / float64(pe.TotalOps) * float64(rows*cols) / 1e6

	return breakdown
}

// EstimateWriteEnergy estimates energy for weight programming
func (pe *PowerEstimator) EstimateWriteEnergy(numCells int, iterationsPerCell float64) float64 {
	// Write energy per cell (typically much higher than read)
	totalEnergy := float64(numCells) * iterationsPerCell * pe.Config.CellWriteEnergy

	// Add verify reads (write-verify programming)
	verifyEnergy := float64(numCells) * iterationsPerCell * pe.Config.CellReadEnergy / 1000
	totalEnergy += verifyEnergy

	return totalEnergy // pJ
}

// EstimateInferenceEnergy estimates total inference energy for a network
func (pe *PowerEstimator) EstimateInferenceEnergy(layerSizes [][]int, batchSize int, sparsity float64) *InferenceEnergyReport {
	report := &InferenceEnergyReport{
		LayerEnergies: make([]float64, len(layerSizes)),
		LayerMACs:     make([]int64, len(layerSizes)),
	}

	totalEnergy := 0.0
	totalMACs := int64(0)

	for i, size := range layerSizes {
		rows, cols := size[0], size[1]

		// Number of array tiles needed
		tilesX := (rows + pe.ArrayRows - 1) / pe.ArrayRows
		tilesY := (cols + pe.ArrayCols - 1) / pe.ArrayCols
		numTiles := tilesX * tilesY

		// Energy per tile MVM
		breakdown := pe.EstimateMVMEnergy(sparsity, 0)
		layerEnergy := breakdown.TotalEnergy * float64(numTiles) * float64(batchSize)

		report.LayerEnergies[i] = layerEnergy
		report.LayerMACs[i] = int64(rows) * int64(cols) * int64(batchSize)

		totalEnergy += layerEnergy
		totalMACs += report.LayerMACs[i]
	}

	report.TotalEnergy = totalEnergy
	report.TotalMACs = totalMACs
	report.EnergyPerMAC = totalEnergy * 1000 / float64(totalMACs) // fJ/MAC
	report.TOPSW = float64(totalMACs) / (totalEnergy * 1e-12) / 1e12

	return report
}

// InferenceEnergyReport contains network-level energy analysis
type InferenceEnergyReport struct {
	LayerEnergies []float64
	LayerMACs     []int64
	TotalEnergy   float64 // pJ
	TotalMACs     int64
	EnergyPerMAC  float64 // fJ/MAC
	TOPSW         float64 // TOPS/W
}

// GetStatistics returns cumulative statistics
func (pe *PowerEstimator) GetStatistics() map[string]float64 {
	return map[string]float64{
		"TotalOps":      float64(pe.TotalOps),
		"TotalEnergyPJ": pe.TotalEnergy,
		"PeakPowerMW":   pe.PeakPower,
		"AvgPowerMW":    pe.AveragePower,
		"EfficiencyTOPSW": float64(pe.TotalOps) / (pe.TotalEnergy * 1e-12) / 1e12,
	}
}

// ================== 3D Crossbar Architecture ==================

// Array3DConfig configures 3D crossbar array
type Array3DConfig struct {
	// Physical dimensions
	NumLayers     int     // Number of vertical layers
	RowsPerLayer  int     // Rows per layer
	ColsPerLayer  int     // Columns per layer

	// Architecture type
	Architecture  string  // "vertical_nand", "horizontal_nand", "stacked_crossbar"

	// Cell parameters
	CellSize      float64 // Cell size (nm)
	LayerPitch    float64 // Vertical layer pitch (nm)
	MinFeature    float64 // Minimum feature size F (nm)

	// Electrical parameters
	InterLayerVia float64 // Via resistance (Ohm)
	LayerCapacitance float64 // Parasitic capacitance (fF)

	// Memory window
	MemoryWindow  float64 // Threshold voltage window (V)
	ProgramVoltage float64 // Program voltage (V)
	EraseVoltage  float64 // Erase voltage (V)

	// Multi-level cell
	BitsPerCell   int     // Bits per cell (1=SLC, 2=MLC, 3=TLC)
}

// Default3DConfig returns typical 3D FeNAND parameters
func Default3DConfig() *Array3DConfig {
	return &Array3DConfig{
		NumLayers:       128,
		RowsPerLayer:    256,
		ColsPerLayer:    256,

		Architecture:    "vertical_nand",

		CellSize:        50,    // 50 nm
		LayerPitch:      40,    // 40 nm layer pitch
		MinFeature:      20,    // 20 nm F

		InterLayerVia:   100,   // 100 Ohm via
		LayerCapacitance: 1.0,  // 1 fF

		MemoryWindow:    3.5,   // 3.5V (HZO superlattice)
		ProgramVoltage:  7.0,   // ±7V P/E
		EraseVoltage:    -7.0,

		BitsPerCell:     3,     // TLC
	}
}

// Cell3D represents a 3D memory cell
type Cell3D struct {
	Layer       int
	Row         int
	Col         int
	Conductance float64
	Threshold   float64  // For FeFET
	State       int      // Multi-level state
	Disturbed   bool     // Disturb flag
}

// Array3D represents a 3D crossbar array
type Array3D struct {
	Config      *Array3DConfig
	Cells       [][][]*Cell3D // [layer][row][col]
	TotalCells  int64
	Density     float64 // Gb/mm²
	AreaMM2     float64

	// String (NAND) organization
	StringLength int   // Cells per string
	NumStrings   int   // Total strings

	// Statistics
	ReadOps     int64
	WriteOps    int64
	DisturbEvents int64
}

// NewArray3D creates a 3D crossbar array
func NewArray3D(config *Array3DConfig) *Array3D {
	cells := make([][][]*Cell3D, config.NumLayers)
	for l := 0; l < config.NumLayers; l++ {
		cells[l] = make([][]*Cell3D, config.RowsPerLayer)
		for r := 0; r < config.RowsPerLayer; r++ {
			cells[l][r] = make([]*Cell3D, config.ColsPerLayer)
			for c := 0; c < config.ColsPerLayer; c++ {
				cells[l][r][c] = &Cell3D{
					Layer:       l,
					Row:         r,
					Col:         c,
					Conductance: 1e-6, // 1 µS default
					Threshold:   0,
					State:       0,
					Disturbed:   false,
				}
			}
		}
	}

	totalCells := int64(config.NumLayers) * int64(config.RowsPerLayer) * int64(config.ColsPerLayer)

	// Calculate area (simplified)
	// Vertical NAND: footprint = rows × cols × F²
	footprintNM2 := float64(config.RowsPerLayer) * float64(config.ColsPerLayer) *
		config.MinFeature * config.MinFeature
	areaMM2 := footprintNM2 / 1e12

	// Density: total bits / area
	totalBits := totalCells * int64(config.BitsPerCell)
	densityGbMM2 := float64(totalBits) / 1e9 / areaMM2

	return &Array3D{
		Config:      config,
		Cells:       cells,
		TotalCells:  totalCells,
		Density:     densityGbMM2,
		AreaMM2:     areaMM2,
		StringLength: config.NumLayers,
		NumStrings:  config.RowsPerLayer * config.ColsPerLayer,
	}
}

// ReadCell reads a single cell
func (a *Array3D) ReadCell(layer, row, col int) (float64, error) {
	cell := a.Cells[layer][row][col]
	a.ReadOps++

	// Check for read disturb in adjacent cells
	a.checkReadDisturb(layer, row, col)

	return cell.Conductance, nil
}

// WriteCell writes a single cell
func (a *Array3D) WriteCell(layer, row, col int, state int) error {
	cell := a.Cells[layer][row][col]
	cfg := a.Config

	// Calculate target conductance for state
	numStates := 1 << cfg.BitsPerCell
	cell.State = state % numStates
	cell.Conductance = float64(cell.State) / float64(numStates-1) * 100e-6 // 0-100 µS

	// Calculate threshold voltage
	cell.Threshold = float64(cell.State) * cfg.MemoryWindow / float64(numStates-1)

	a.WriteOps++

	// Check for program disturb
	a.checkProgramDisturb(layer, row, col)

	return nil
}

// checkReadDisturb checks for read disturb effects
func (a *Array3D) checkReadDisturb(layer, row, col int) {
	cfg := a.Config

	// In NAND, cells in the same string experience pass voltage
	// This can cause threshold shifts
	for l := 0; l < cfg.NumLayers; l++ {
		if l != layer {
			cell := a.Cells[l][row][col]
			// Small probability of disturb (simplified model)
			if !cell.Disturbed {
				// Accumulate small threshold shift
				cell.Threshold += 0.001 // 1mV per read of string
				if cell.Threshold > cfg.MemoryWindow*1.1 {
					cell.Disturbed = true
					a.DisturbEvents++
				}
			}
		}
	}
}

// checkProgramDisturb checks for program disturb effects
func (a *Array3D) checkProgramDisturb(layer, row, col int) {
	cfg := a.Config

	// Cells sharing WL or BL may be disturbed
	// Same WL (same layer, same row)
	for c := 0; c < cfg.ColsPerLayer; c++ {
		if c != col {
			cell := a.Cells[layer][row][c]
			// Program inhibit stress
			cell.Threshold += 0.01 // 10mV per neighbor program
		}
	}

	// Same BL (same row, same col, different layer)
	for l := 0; l < cfg.NumLayers; l++ {
		if l != layer {
			cell := a.Cells[l][row][col]
			// Pass disturb
			cell.Threshold += 0.005 // 5mV
		}
	}
}

// ReadString reads an entire NAND string (all layers at row,col)
func (a *Array3D) ReadString(row, col int) []float64 {
	cfg := a.Config
	values := make([]float64, cfg.NumLayers)

	for l := 0; l < cfg.NumLayers; l++ {
		values[l] = a.Cells[l][row][col].Conductance
	}

	a.ReadOps += int64(cfg.NumLayers)
	return values
}

// ================== 3D CIM Operations ==================

// CIM3DConfig configures 3D CIM operations
type CIM3DConfig struct {
	*Array3DConfig
	ComputeMode   string  // "layer_by_layer", "pipe_parallel", "full_3d"
	InputPrecision int    // Input precision (bits)
	WeightPrecision int   // Weight precision (bits)
	OutputPrecision int   // Output precision (bits)
	UseADCSharing bool    // Share ADCs across layers
}

// DefaultCIM3DConfig returns standard 3D CIM config
func DefaultCIM3DConfig() *CIM3DConfig {
	return &CIM3DConfig{
		Array3DConfig:   Default3DConfig(),
		ComputeMode:     "layer_by_layer",
		InputPrecision:  8,
		WeightPrecision: 4,
		OutputPrecision: 16,
		UseADCSharing:   true,
	}
}

// CIM3DAccelerator implements 3D CIM computation
type CIM3DAccelerator struct {
	Config      *CIM3DConfig
	Array       *Array3D
	Power       *PowerEstimator
	LayerMap    []int  // Which physical layers map to which network layers

	// Buffers
	InputBuffer  [][]float64
	OutputBuffer [][]float64

	// Statistics
	TotalMACs    int64
	TotalLatency float64 // ns
}

// NewCIM3DAccelerator creates a 3D CIM accelerator
func NewCIM3DAccelerator(config *CIM3DConfig) *CIM3DAccelerator {
	array := NewArray3D(config.Array3DConfig)

	powerConfig := DefaultPowerConfig()
	power := NewPowerEstimator(powerConfig, config.RowsPerLayer, config.ColsPerLayer, 1, 100)

	return &CIM3DAccelerator{
		Config:   config,
		Array:    array,
		Power:    power,
		LayerMap: make([]int, config.NumLayers),
	}
}

// LoadWeights loads neural network weights into 3D array
func (acc *CIM3DAccelerator) LoadWeights(weights [][][]float64) error {
	cfg := acc.Config

	for l := 0; l < len(weights) && l < cfg.NumLayers; l++ {
		for r := 0; r < len(weights[l]) && r < cfg.RowsPerLayer; r++ {
			for c := 0; c < len(weights[l][r]) && c < cfg.ColsPerLayer; c++ {
				// Quantize weight to state
				w := weights[l][r][c]
				numStates := 1 << cfg.WeightPrecision
				state := int((w + 1.0) / 2.0 * float64(numStates-1)) // Map [-1,1] to [0, numStates-1]
				acc.Array.WriteCell(l, r, c, state)
			}
		}
		acc.LayerMap[l] = l
	}

	return nil
}

// ComputeMVM3D performs 3D matrix-vector multiplication
func (acc *CIM3DAccelerator) ComputeMVM3D(input []float64, layerIdx int) []float64 {
	cfg := acc.Config
	layer := acc.LayerMap[layerIdx]

	output := make([]float64, cfg.ColsPerLayer)

	// Layer-by-layer MVM
	for c := 0; c < cfg.ColsPerLayer; c++ {
		sum := 0.0
		for r := 0; r < cfg.RowsPerLayer && r < len(input); r++ {
			g := acc.Array.Cells[layer][r][c].Conductance
			sum += input[r] * g
		}
		output[c] = sum
	}

	acc.TotalMACs += int64(cfg.RowsPerLayer) * int64(cfg.ColsPerLayer)
	acc.Power.EstimateMVMEnergy(0, 0)

	return output
}

// ComputePipelined performs pipelined 3D computation
func (acc *CIM3DAccelerator) ComputePipelined(inputs [][]float64) [][]float64 {
	cfg := acc.Config
	numInputs := len(inputs)

	outputs := make([][]float64, numInputs)
	for i := range outputs {
		outputs[i] = make([]float64, cfg.ColsPerLayer)
	}

	// Pipeline across layers
	// Each layer processes while next receives input
	for stage := 0; stage < numInputs+cfg.NumLayers-1; stage++ {
		// Process all active layers in parallel
		for l := 0; l < cfg.NumLayers; l++ {
			inputIdx := stage - l
			if inputIdx >= 0 && inputIdx < numInputs {
				// This layer processes this input
				if l == 0 {
					// First layer uses external input
					outputs[inputIdx] = acc.ComputeMVM3D(inputs[inputIdx], l)
				} else {
					// Subsequent layers use previous layer output
					// (simplified: use same input for demo)
					acc.ComputeMVM3D(inputs[inputIdx], l)
				}
			}
		}

		// Latency per stage
		acc.TotalLatency += 10.0 // 10 ns per stage (simplified)
	}

	return outputs
}

// GetEfficiency returns efficiency metrics
func (acc *CIM3DAccelerator) GetEfficiency() map[string]float64 {
	stats := acc.Power.GetStatistics()

	return map[string]float64{
		"TotalMACs":      float64(acc.TotalMACs),
		"TotalLatencyNs": acc.TotalLatency,
		"ThroughputGOPS": float64(acc.TotalMACs) / acc.TotalLatency,
		"TOPSW":          stats["EfficiencyTOPSW"],
		"DensityGbMM2":   acc.Array.Density,
		"DisturbEvents":  float64(acc.Array.DisturbEvents),
	}
}

// ================== Vertical NAND Specific ==================

// VerticalNANDConfig configures vertical channel FeNAND
type VerticalNANDConfig struct {
	*Array3DConfig
	ChannelDiameter  float64 // Channel hole diameter (nm)
	GateLength       float64 // Gate length (nm)
	ChargeTrapping   bool    // Hybrid with charge trap
	FerroelectricThk float64 // Ferroelectric thickness (nm)
}

// DefaultVerticalNANDConfig returns typical VC-FeNAND parameters
func DefaultVerticalNANDConfig() *VerticalNANDConfig {
	return &VerticalNANDConfig{
		Array3DConfig:    Default3DConfig(),
		ChannelDiameter:  80,   // 80 nm hole
		GateLength:       30,   // 30 nm gate
		ChargeTrapping:   true, // Hybrid FE+CT
		FerroelectricThk: 10,   // 10 nm HZO
	}
}

// VerticalNANDArray models vertical channel NAND
type VerticalNANDArray struct {
	Config        *VerticalNANDConfig
	Strings       [][]*NANDString // [row][col]
	NumStrings    int
	StringLength  int
	TotalCapacity int64 // bits
}

// NANDString represents a vertical NAND string
type NANDString struct {
	Row     int
	Col     int
	Cells   []*FeFETCell
	SSL     bool // String select transistor state
	GSL     bool // Ground select transistor state
}

// FeFETCell represents a ferroelectric FET cell
type FeFETCell struct {
	Layer         int
	Vth           float64 // Threshold voltage
	Polarization  float64 // Ferroelectric polarization
	ChargeTrapped float64 // Trapped charge (if hybrid)
	State         int     // Multi-level state
	Endurance     int64   // Remaining cycles
}

// NewVerticalNANDArray creates a vertical NAND array
func NewVerticalNANDArray(config *VerticalNANDConfig) *VerticalNANDArray {
	strings := make([][]*NANDString, config.RowsPerLayer)
	for r := 0; r < config.RowsPerLayer; r++ {
		strings[r] = make([]*NANDString, config.ColsPerLayer)
		for c := 0; c < config.ColsPerLayer; c++ {
			cells := make([]*FeFETCell, config.NumLayers)
			for l := 0; l < config.NumLayers; l++ {
				cells[l] = &FeFETCell{
					Layer:         l,
					Vth:           0,
					Polarization:  0,
					ChargeTrapped: 0,
					State:         0,
					Endurance:     1e10, // 10^10 cycles for HZO
				}
			}
			strings[r][c] = &NANDString{
				Row:   r,
				Col:   c,
				Cells: cells,
				SSL:   false,
				GSL:   false,
			}
		}
	}

	totalCapacity := int64(config.NumLayers) * int64(config.RowsPerLayer) *
		int64(config.ColsPerLayer) * int64(config.BitsPerCell)

	return &VerticalNANDArray{
		Config:        config,
		Strings:       strings,
		NumStrings:    config.RowsPerLayer * config.ColsPerLayer,
		StringLength:  config.NumLayers,
		TotalCapacity: totalCapacity,
	}
}

// ProgramCell programs a cell in the vertical NAND
func (vna *VerticalNANDArray) ProgramCell(row, col, layer, state int, voltage float64) error {
	str := vna.Strings[row][col]
	cell := str.Cells[layer]

	// Enable string
	str.SSL = true
	str.GSL = true

	// Apply program voltage
	// Ferroelectric polarization switching
	cfg := vna.Config
	if math.Abs(voltage) > cfg.ProgramVoltage*0.5 {
		// Polarization switching
		if voltage > 0 {
			cell.Polarization = math.Tanh(voltage / cfg.ProgramVoltage)
		} else {
			cell.Polarization = -math.Tanh(-voltage / cfg.ProgramVoltage)
		}
	}

	// Hybrid: charge trapping contribution
	if cfg.ChargeTrapping {
		cell.ChargeTrapped += voltage * 0.1 // Simplified
	}

	// Update threshold voltage
	cell.Vth = cell.Polarization*cfg.MemoryWindow/2 + cell.ChargeTrapped*0.5
	cell.State = state
	cell.Endurance--

	// Disable string
	str.SSL = false
	str.GSL = false

	return nil
}

// ReadString reads all cells in a string for CIM
func (vna *VerticalNANDArray) ReadString(row, col int) []float64 {
	str := vna.Strings[row][col]
	cfg := vna.Config

	// Enable string
	str.SSL = true
	str.GSL = true

	currents := make([]float64, cfg.NumLayers)
	for l, cell := range str.Cells {
		// Current depends on Vth relative to read voltage
		readV := cfg.MemoryWindow / 2 // Mid-window read
		currents[l] = math.Max(0, readV-cell.Vth) * 1e-6 // Simplified I-V
	}

	str.SSL = false
	str.GSL = false

	return currents
}

// ComputeMAC performs MAC using the vertical NAND
func (vna *VerticalNANDArray) ComputeMAC(inputs []float64) []float64 {
	cfg := vna.Config
	outputs := make([]float64, cfg.ColsPerLayer)

	// Each column (string) computes dot product
	for c := 0; c < cfg.ColsPerLayer; c++ {
		for r := 0; r < cfg.RowsPerLayer && r < len(inputs); r++ {
			// Read string and weight by input
			currents := vna.ReadString(r, c)
			for l := 0; l < cfg.NumLayers; l++ {
				outputs[c] += inputs[r] * currents[l]
			}
		}
	}

	return outputs
}

// GetDensity returns the array density
func (vna *VerticalNANDArray) GetDensity() float64 {
	cfg := vna.Config

	// Effective cell size: 4F² per cell due to vertical stacking
	cellArea := 4 * cfg.MinFeature * cfg.MinFeature // nm²
	totalArea := cellArea * float64(vna.NumStrings) // nm² (footprint only)

	// Total bits
	totalBits := float64(vna.TotalCapacity)

	// Density in Gb/mm²
	return totalBits / 1e9 / (totalArea / 1e12)
}
