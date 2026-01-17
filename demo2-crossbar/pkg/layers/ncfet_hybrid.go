// ncfet_hybrid.go - Negative Capacitance FETs and Hybrid Analog-Digital CIM
// for Ferroelectric Compute-in-Memory (CIM) Systems
//
// This module implements:
// 1. Negative Capacitance FET (NC-FET) device physics
// 2. NC-FET based SRAM CIM cells for energy efficiency
// 3. Hybrid analog-digital CIM architectures
// 4. Bit-serial and bit-parallel computing schemes
// 5. Mixed-precision CIM with configurable ADC
//
// Based on research:
// - NC-FET with HZO achieving SS < 10 mV/dec
// - HCiM: ADC-less hybrid analog-digital accelerator
// - NCFET 6T-SRAM CIM with 22.77× energy reduction
// - Mixed-precision CIM macros (1-16 bit configurable)

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// NEGATIVE CAPACITANCE FET CONFIGURATION
// =============================================================================

// NCFETConfig configures Negative Capacitance FET
type NCFETConfig struct {
	// Device parameters
	FerroelectricMaterial string  // "HZO", "ZrO2", "PZT"
	FEThicknessNm         float64 // Ferroelectric thickness (nm)
	DEThicknessNm         float64 // Dielectric thickness (nm)

	// Capacitance parameters
	FerroCapacitance      float64 // CFE (fF/μm²)
	MOSCapacitance        float64 // CMOS (fF/μm²)
	InterfaceCapacitance  float64 // Cint (fF/μm²)

	// NC effect parameters
	VoltageGain           float64 // Internal voltage amplification (>1)
	BodyFactor            float64 // m = 1/β (<1 for NC)
	SubthresholdSwing     float64 // SS (mV/dec)

	// Operating parameters
	SupplyVoltage         float64 // VDD (V)
	ThresholdVoltage      float64 // Vth (V)
	OnOffRatio            float64 // Ion/Ioff

	// Reliability
	EnduranceCycles       float64
	RetentionYears        float64
}

// DefaultNCFETConfig returns configuration for HZO-based NC-FET
func DefaultNCFETConfig() *NCFETConfig {
	return &NCFETConfig{
		FerroelectricMaterial: "HZO",
		FEThicknessNm:         12.0,
		DEThicknessNm:         2.0,
		FerroCapacitance:      15.0,  // 15 fF/μm²
		MOSCapacitance:        5.0,   // 5 fF/μm²
		InterfaceCapacitance:  10.0,  // 10 fF/μm²
		VoltageGain:           28.0,  // ~28× demonstrated
		BodyFactor:            0.5,   // < 1
		SubthresholdSwing:     8.0,   // 8 mV/dec (demonstrated)
		SupplyVoltage:         0.5,   // 0.5V low voltage
		ThresholdVoltage:      0.25,
		OnOffRatio:            1e8,
		EnduranceCycles:       1e10,
		RetentionYears:        10.0,
	}
}

// =============================================================================
// NC-FET DEVICE MODEL
// =============================================================================

// NCFETDevice models a Negative Capacitance FET
type NCFETDevice struct {
	Config *NCFETConfig

	// State
	GateVoltage       float64
	InternalVoltage   float64 // Vint (amplified)
	DrainCurrent      float64
	Polarization      float64

	// Computed parameters
	EffectiveCapacitance float64
	NegativeCapacitance  float64
}

// NewNCFETDevice creates an NC-FET device
func NewNCFETDevice(config *NCFETConfig) *NCFETDevice {
	device := &NCFETDevice{
		Config: config,
	}

	// Compute effective capacitance
	// In NC regime: Ceff = CFE * CMOS / (CFE + CMOS)
	// When CFE < 0: Ceff can be > CMOS (amplification)
	device.computeCapacitances()

	return device
}

// computeCapacitances calculates NC-related capacitances
func (dev *NCFETDevice) computeCapacitances() {
	// For stable NC: |CFE| > CMOS
	// Negative capacitance from Landau theory
	cfe := dev.Config.FerroCapacitance
	cmos := dev.Config.MOSCapacitance

	// In NC regime, effective CFE is negative
	dev.NegativeCapacitance = -cfe * 0.8 // ~80% of FE cap becomes negative

	// Effective gate capacitance with NC boost
	// Ceff = |CFE| * CMOS / (|CFE| - CMOS) when |CFE| > CMOS
	if math.Abs(dev.NegativeCapacitance) > cmos {
		dev.EffectiveCapacitance = math.Abs(dev.NegativeCapacitance) * cmos /
			(math.Abs(dev.NegativeCapacitance) - cmos)
	} else {
		dev.EffectiveCapacitance = cmos
	}
}

// SetGateVoltage sets gate voltage and computes internal voltage
func (dev *NCFETDevice) SetGateVoltage(vg float64) {
	dev.GateVoltage = vg

	// Internal voltage amplification: Vint = β * Vg
	dev.InternalVoltage = dev.Config.VoltageGain * vg

	// Clamp to physical limits
	if dev.InternalVoltage > dev.Config.SupplyVoltage*2 {
		dev.InternalVoltage = dev.Config.SupplyVoltage * 2
	}

	// Compute drain current (subthreshold)
	dev.computeDrainCurrent()
}

// computeDrainCurrent calculates drain current using steep SS
func (dev *NCFETDevice) computeDrainCurrent() {
	vth := dev.Config.ThresholdVoltage
	vint := dev.InternalVoltage
	ss := dev.Config.SubthresholdSwing / 1000.0 // Convert to V/dec

	// Thermal voltage
	vt := 0.026 // 26 mV at room temperature

	// Subthreshold current with NC-enhanced SS
	// I = I0 * 10^((Vint - Vth) / SS)
	i0 := 1e-12 // Reference current (1 pA)

	if vint < vth {
		// Subthreshold region
		exponent := (vint - vth) / ss
		dev.DrainCurrent = i0 * math.Pow(10, exponent)
	} else {
		// Above threshold (simplified linear region)
		dev.DrainCurrent = i0 * dev.Config.OnOffRatio * (1 + (vint-vth)/vth)
	}

	// Add noise
	dev.DrainCurrent *= (1 + rand.NormFloat64()*0.01)
}

// GetSubthresholdSwing returns effective SS
func (dev *NCFETDevice) GetSubthresholdSwing() float64 {
	// SS = (kT/q) * ln(10) * m
	// where m = body factor = 1/β
	boltzmannSS := 60.0 // Boltzmann limit at room temperature (mV/dec)
	return boltzmannSS * dev.Config.BodyFactor
}

// GetEnergyPerSwitch returns switching energy
func (dev *NCFETDevice) GetEnergyPerSwitch() float64 {
	// E = 0.5 * Ceff * VDD²
	// NC reduces effective capacitance, lowering energy
	ceff := dev.EffectiveCapacitance * 1e-15 // Convert to F
	vdd := dev.Config.SupplyVoltage
	return 0.5 * ceff * vdd * vdd * 1e12 // pJ
}

// =============================================================================
// NC-FET SRAM CIM CELL
// =============================================================================

// NCFETSRAMCell represents a 6T SRAM cell with NC-FETs
type NCFETSRAMCell struct {
	Config *NCFETConfig

	// Cell transistors (6T: 2 access, 4 inverter)
	AccessTransistors [2]*NCFETDevice
	InverterNMOS      [2]*NCFETDevice
	InverterPMOS      [2]*NCFETDevice

	// Storage
	StoredValue int  // 0 or 1
	Weight      float64 // Analog weight (for MLC)

	// CIM support
	ComputeEnabled bool
	BitlineVoltage float64
	WordlineVoltage float64
}

// NewNCFETSRAMCell creates an NC-FET SRAM cell
func NewNCFETSRAMCell(config *NCFETConfig) *NCFETSRAMCell {
	cell := &NCFETSRAMCell{
		Config: config,
	}

	// Create access transistors
	for i := 0; i < 2; i++ {
		cell.AccessTransistors[i] = NewNCFETDevice(config)
	}

	// Create inverter transistors
	for i := 0; i < 2; i++ {
		cell.InverterNMOS[i] = NewNCFETDevice(config)
		cell.InverterPMOS[i] = NewNCFETDevice(config)
	}

	return cell
}

// Write stores a value in the cell
func (cell *NCFETSRAMCell) Write(value int) {
	cell.StoredValue = value & 1
	cell.Weight = float64(cell.StoredValue)

	// Set internal state
	if cell.StoredValue == 1 {
		cell.InverterNMOS[0].SetGateVoltage(cell.Config.SupplyVoltage)
		cell.InverterNMOS[1].SetGateVoltage(0)
	} else {
		cell.InverterNMOS[0].SetGateVoltage(0)
		cell.InverterNMOS[1].SetGateVoltage(cell.Config.SupplyVoltage)
	}
}

// Read returns the stored value
func (cell *NCFETSRAMCell) Read() int {
	return cell.StoredValue
}

// ComputeMAC performs in-memory multiply-accumulate
func (cell *NCFETSRAMCell) ComputeMAC(input float64) float64 {
	// NC-FET advantage: lower energy at same precision
	// Weight × Input using current-domain computation

	// Set wordline (input activation)
	cell.WordlineVoltage = input * cell.Config.SupplyVoltage
	cell.AccessTransistors[0].SetGateVoltage(cell.WordlineVoltage)

	// Current through cell proportional to weight × input
	weight := cell.Weight
	inputScaled := input

	// Compute with NC-enhanced efficiency
	result := weight * inputScaled

	// Add noise (reduced due to NC voltage gain)
	noise := rand.NormFloat64() * 0.005 // Lower noise than conventional
	result += noise

	return result
}

// GetEnergy returns compute energy (reduced by NC effect)
func (cell *NCFETSRAMCell) GetEnergy() float64 {
	// NC-FET reduces energy by ~22.77× at 0.3V
	// and ~12.41× at 0.5V compared to baseline CMOS
	baseEnergy := 1.0 // Normalized baseline

	if cell.Config.SupplyVoltage <= 0.3 {
		return baseEnergy / 22.77
	} else if cell.Config.SupplyVoltage <= 0.5 {
		return baseEnergy / 12.41
	}
	return baseEnergy / 3.0 // Still 3× improvement at higher voltage
}

// =============================================================================
// NC-FET CIM ARRAY
// =============================================================================

// NCFETCIMArray represents an array of NC-FET SRAM CIM cells
type NCFETCIMArray struct {
	Config     *NCFETConfig
	Rows       int
	Cols       int
	Cells      [][]*NCFETSRAMCell

	// Statistics
	TotalMACs      int64
	TotalEnergy    float64
	ComputeCycles  int
}

// NewNCFETCIMArray creates an NC-FET CIM array
func NewNCFETCIMArray(config *NCFETConfig, rows, cols int) *NCFETCIMArray {
	array := &NCFETCIMArray{
		Config: config,
		Rows:   rows,
		Cols:   cols,
		Cells:  make([][]*NCFETSRAMCell, rows),
	}

	for r := 0; r < rows; r++ {
		array.Cells[r] = make([]*NCFETSRAMCell, cols)
		for c := 0; c < cols; c++ {
			array.Cells[r][c] = NewNCFETSRAMCell(config)
		}
	}

	return array
}

// LoadWeights loads weight matrix into the array
func (array *NCFETCIMArray) LoadWeights(weights [][]float64) {
	for r := 0; r < array.Rows && r < len(weights); r++ {
		for c := 0; c < array.Cols && c < len(weights[r]); c++ {
			// Quantize to binary for basic SRAM
			if weights[r][c] > 0.5 {
				array.Cells[r][c].Write(1)
			} else {
				array.Cells[r][c].Write(0)
			}
			array.Cells[r][c].Weight = weights[r][c]
		}
	}
}

// ComputeMVM performs matrix-vector multiplication
func (array *NCFETCIMArray) ComputeMVM(inputs []float64) []float64 {
	outputs := make([]float64, array.Cols)

	// Each column accumulates weighted inputs
	for c := 0; c < array.Cols; c++ {
		sum := 0.0
		for r := 0; r < array.Rows && r < len(inputs); r++ {
			sum += array.Cells[r][c].ComputeMAC(inputs[r])
			array.TotalMACs++
			array.TotalEnergy += array.Cells[r][c].GetEnergy()
		}
		outputs[c] = sum
	}

	array.ComputeCycles++
	return outputs
}

// GetEnergyEfficiency returns TOPS/W
func (array *NCFETCIMArray) GetEnergyEfficiency() float64 {
	if array.TotalEnergy == 0 {
		return 0
	}
	// TOPS/W = (MACs / Time) / Power
	// Simplified: (MACs / Energy) * scaling
	return float64(array.TotalMACs) / array.TotalEnergy * 1e3
}

// =============================================================================
// HYBRID ANALOG-DIGITAL CIM CONFIGURATION
// =============================================================================

// HybridCIMConfig configures hybrid analog-digital CIM
type HybridCIMConfig struct {
	// Precision settings
	InputPrecision   int  // Input bits (1-16)
	WeightPrecision  int  // Weight bits (1-8)
	OutputPrecision  int  // Output bits (8-32)
	AccumPrecision   int  // Accumulator bits

	// Computing mode
	ComputeMode      string // "Analog", "Digital", "Hybrid"
	BitSerialEnabled bool   // Bit-serial computation
	BitParallelCols  int    // Columns for bit-parallel

	// ADC configuration
	ADCEnabled       bool
	ADCBits          int     // ADC resolution
	ADCsPerCrossbar  int     // Number of ADCs
	ADCLessMode      bool    // HCiM-style ADC-less

	// Scale factor handling
	ScaleFactorBits  int     // Bits for scale factors
	DigitalScaling   bool    // Digital scale factor processing

	// Array configuration
	CrossbarRows     int
	CrossbarCols     int
	NumCrossbars     int
}

// DefaultHybridCIMConfig returns default hybrid configuration
func DefaultHybridCIMConfig() *HybridCIMConfig {
	return &HybridCIMConfig{
		InputPrecision:   8,
		WeightPrecision:  8,
		OutputPrecision:  32,
		AccumPrecision:   32,
		ComputeMode:      "Hybrid",
		BitSerialEnabled: true,
		BitParallelCols:  8,
		ADCEnabled:       true,
		ADCBits:          6,
		ADCsPerCrossbar:  8,
		ADCLessMode:      false,
		ScaleFactorBits:  16,
		DigitalScaling:   true,
		CrossbarRows:     256,
		CrossbarCols:     256,
		NumCrossbars:     16,
	}
}

// =============================================================================
// HYBRID CIM CROSSBAR
// =============================================================================

// HybridCIMCrossbar implements hybrid analog-digital computation
type HybridCIMCrossbar struct {
	Config *HybridCIMConfig

	// Weight storage
	Weights         [][]float64
	QuantizedWeights [][]int
	ScaleFactors    []float64 // Per-column scale factors

	// Bit slices for bit-serial
	WeightSlices    [][][]int // [slice][row][col]
	NumSlices       int

	// ADC model
	ADCQuantLevels  int
	ADCNoiseLevel   float64

	// Statistics
	AnalogMACs      int64
	DigitalMACs     int64
	ADCConversions  int64
	TotalEnergy     float64
}

// NewHybridCIMCrossbar creates a hybrid CIM crossbar
func NewHybridCIMCrossbar(config *HybridCIMConfig) *HybridCIMCrossbar {
	crossbar := &HybridCIMCrossbar{
		Config:         config,
		Weights:        make([][]float64, config.CrossbarRows),
		QuantizedWeights: make([][]int, config.CrossbarRows),
		ScaleFactors:   make([]float64, config.CrossbarCols),
		ADCQuantLevels: 1 << config.ADCBits,
		ADCNoiseLevel:  0.5 / float64(1<<config.ADCBits),
	}

	// Initialize weight storage
	for r := 0; r < config.CrossbarRows; r++ {
		crossbar.Weights[r] = make([]float64, config.CrossbarCols)
		crossbar.QuantizedWeights[r] = make([]int, config.CrossbarCols)
	}

	// Initialize scale factors
	for c := 0; c < config.CrossbarCols; c++ {
		crossbar.ScaleFactors[c] = 1.0
	}

	// Setup bit slices if bit-serial enabled
	if config.BitSerialEnabled {
		crossbar.NumSlices = (config.WeightPrecision + 3) / 4 // 4-bit slices
		crossbar.WeightSlices = make([][][]int, crossbar.NumSlices)
		for s := 0; s < crossbar.NumSlices; s++ {
			crossbar.WeightSlices[s] = make([][]int, config.CrossbarRows)
			for r := 0; r < config.CrossbarRows; r++ {
				crossbar.WeightSlices[s][r] = make([]int, config.CrossbarCols)
			}
		}
	}

	return crossbar
}

// LoadWeights loads and quantizes weights
func (cb *HybridCIMCrossbar) LoadWeights(weights [][]float64) {
	maxVal := 0.0
	for r := range weights {
		for c := range weights[r] {
			if math.Abs(weights[r][c]) > maxVal {
				maxVal = math.Abs(weights[r][c])
			}
		}
	}

	// Quantize weights
	quantLevels := 1 << cb.Config.WeightPrecision
	for r := 0; r < cb.Config.CrossbarRows && r < len(weights); r++ {
		for c := 0; c < cb.Config.CrossbarCols && c < len(weights[r]); c++ {
			cb.Weights[r][c] = weights[r][c]

			// Normalize and quantize
			normalized := weights[r][c] / maxVal
			cb.QuantizedWeights[r][c] = int((normalized + 1.0) / 2.0 * float64(quantLevels-1))
		}
	}

	// Compute per-column scale factors
	for c := 0; c < cb.Config.CrossbarCols; c++ {
		colMax := 0.0
		for r := 0; r < cb.Config.CrossbarRows && r < len(weights); r++ {
			if c < len(weights[r]) && math.Abs(weights[r][c]) > colMax {
				colMax = math.Abs(weights[r][c])
			}
		}
		cb.ScaleFactors[c] = colMax
	}

	// Create bit slices for bit-serial
	if cb.Config.BitSerialEnabled {
		cb.createBitSlices()
	}
}

// createBitSlices creates weight slices for bit-serial computation
func (cb *HybridCIMCrossbar) createBitSlices() {
	sliceMask := (1 << 4) - 1 // 4-bit slices

	for s := 0; s < cb.NumSlices; s++ {
		shift := s * 4
		for r := 0; r < cb.Config.CrossbarRows; r++ {
			for c := 0; c < cb.Config.CrossbarCols; c++ {
				cb.WeightSlices[s][r][c] = (cb.QuantizedWeights[r][c] >> shift) & sliceMask
			}
		}
	}
}

// ComputeAnalogMVM performs analog MVM
func (cb *HybridCIMCrossbar) ComputeAnalogMVM(inputs []float64) []float64 {
	outputs := make([]float64, cb.Config.CrossbarCols)

	for c := 0; c < cb.Config.CrossbarCols; c++ {
		sum := 0.0
		for r := 0; r < cb.Config.CrossbarRows && r < len(inputs); r++ {
			sum += cb.Weights[r][c] * inputs[r]
			cb.AnalogMACs++
		}

		// Add analog noise
		sum += rand.NormFloat64() * cb.ADCNoiseLevel * sum

		outputs[c] = sum
	}

	// ADC conversion
	if cb.Config.ADCEnabled && !cb.Config.ADCLessMode {
		outputs = cb.applyADC(outputs)
	}

	return outputs
}

// ComputeBitSerialMVM performs bit-serial MVM
func (cb *HybridCIMCrossbar) ComputeBitSerialMVM(inputs []float64) []float64 {
	outputs := make([]float64, cb.Config.CrossbarCols)

	// Quantize inputs
	inputBits := cb.Config.InputPrecision
	quantInputs := make([]int, len(inputs))
	for i, inp := range inputs {
		quantInputs[i] = int(inp * float64((1<<inputBits)-1))
	}

	// Process each bit position
	for inputBit := 0; inputBit < inputBits; inputBit++ {
		inputScale := float64(int(1) << inputBit)

		// Get input bit vector
		inputBitVector := make([]int, len(quantInputs))
		for i, qi := range quantInputs {
			inputBitVector[i] = (qi >> inputBit) & 1
		}

		// Process each weight slice
		for slice := 0; slice < cb.NumSlices; slice++ {
			weightScale := float64(int(1) << (slice * 4))

			// Compute partial products
			for c := 0; c < cb.Config.CrossbarCols; c++ {
				partialSum := 0
				for r := 0; r < cb.Config.CrossbarRows && r < len(inputBitVector); r++ {
					partialSum += inputBitVector[r] * cb.WeightSlices[slice][r][c]
					cb.DigitalMACs++
				}

				// Scale and accumulate
				outputs[c] += float64(partialSum) * inputScale * weightScale
			}
		}
	}

	// Apply scale factors
	for c := 0; c < cb.Config.CrossbarCols; c++ {
		outputs[c] *= cb.ScaleFactors[c] / float64((1<<cb.Config.InputPrecision)-1) /
			float64((1<<cb.Config.WeightPrecision)-1)
	}

	return outputs
}

// ComputeHybridMVM performs hybrid analog-digital MVM (HCiM style)
func (cb *HybridCIMCrossbar) ComputeHybridMVM(inputs []float64) []float64 {
	// HCiM: Binary/ternary partial sums from analog, digital scale factors

	outputs := make([]float64, cb.Config.CrossbarCols)

	// Step 1: Analog crossbar computes binary partial products
	binaryPartials := make([][]int, cb.Config.CrossbarCols)
	for c := 0; c < cb.Config.CrossbarCols; c++ {
		binaryPartials[c] = make([]int, cb.NumSlices)
	}

	for c := 0; c < cb.Config.CrossbarCols; c++ {
		for r := 0; r < cb.Config.CrossbarRows && r < len(inputs); r++ {
			// Binary weight × analog input → binary partial sum
			for slice := 0; slice < cb.NumSlices; slice++ {
				if cb.WeightSlices[slice][r][c] > 0 {
					// Threshold to binary
					if inputs[r] > 0.5 {
						binaryPartials[c][slice]++
					}
				}
			}
			cb.AnalogMACs++
		}
	}

	// Step 2: Digital CIM processes scale factors
	for c := 0; c < cb.Config.CrossbarCols; c++ {
		scaledSum := 0.0
		for slice := 0; slice < cb.NumSlices; slice++ {
			sliceScale := float64(int(1) << (slice * 4))
			scaledSum += float64(binaryPartials[c][slice]) * sliceScale
			cb.DigitalMACs++
		}
		outputs[c] = scaledSum * cb.ScaleFactors[c]
	}

	// HCiM achieves 28% energy reduction vs 7-bit ADC
	cb.TotalEnergy += float64(cb.Config.CrossbarRows*cb.Config.CrossbarCols) * 0.72 // 28% reduction

	return outputs
}

// applyADC applies ADC quantization and noise
func (cb *HybridCIMCrossbar) applyADC(values []float64) []float64 {
	quantized := make([]float64, len(values))

	for i, v := range values {
		// Quantize
		levels := float64(cb.ADCQuantLevels)
		q := math.Round(v * levels) / levels

		// Add quantization noise
		q += rand.NormFloat64() * cb.ADCNoiseLevel

		quantized[i] = q
		cb.ADCConversions++
	}

	return quantized
}

// =============================================================================
// MIXED-PRECISION CIM MACRO
// =============================================================================

// MixedPrecisionCIMMacro implements configurable precision CIM
type MixedPrecisionCIMMacro struct {
	Config *HybridCIMConfig

	// Crossbars
	Crossbars []*HybridCIMCrossbar

	// Precision control
	CurrentInputPrecision  int
	CurrentWeightPrecision int
	CurrentOutputPrecision int

	// Accumulator
	Accumulator []float64
	AccumBits   int

	// Statistics
	TotalOps       int64
	EnergyConsumed float64
	Throughput     float64 // TOPS
}

// NewMixedPrecisionCIMMacro creates a mixed-precision CIM macro
func NewMixedPrecisionCIMMacro(config *HybridCIMConfig) *MixedPrecisionCIMMacro {
	macro := &MixedPrecisionCIMMacro{
		Config:                 config,
		Crossbars:              make([]*HybridCIMCrossbar, config.NumCrossbars),
		CurrentInputPrecision:  config.InputPrecision,
		CurrentWeightPrecision: config.WeightPrecision,
		CurrentOutputPrecision: config.OutputPrecision,
		Accumulator:            make([]float64, config.CrossbarCols),
		AccumBits:              config.AccumPrecision,
	}

	// Create crossbars
	for i := 0; i < config.NumCrossbars; i++ {
		macro.Crossbars[i] = NewHybridCIMCrossbar(config)
	}

	return macro
}

// SetPrecision configures runtime precision
func (macro *MixedPrecisionCIMMacro) SetPrecision(inputBits, weightBits, outputBits int) {
	// Validate ranges
	if inputBits < 1 || inputBits > 16 {
		inputBits = 8
	}
	if weightBits < 1 || weightBits > 8 {
		weightBits = 8
	}
	if outputBits < 8 || outputBits > 32 {
		outputBits = 32
	}

	macro.CurrentInputPrecision = inputBits
	macro.CurrentWeightPrecision = weightBits
	macro.CurrentOutputPrecision = outputBits

	// Update crossbar configurations
	for _, cb := range macro.Crossbars {
		cb.Config.InputPrecision = inputBits
		cb.Config.WeightPrecision = weightBits
		cb.Config.OutputPrecision = outputBits
	}
}

// LoadModel loads weights across crossbars
func (macro *MixedPrecisionCIMMacro) LoadModel(weights [][][]float64) {
	for i := 0; i < len(weights) && i < len(macro.Crossbars); i++ {
		macro.Crossbars[i].LoadWeights(weights[i])
	}
}

// Compute performs inference with configured precision
func (macro *MixedPrecisionCIMMacro) Compute(inputs []float64, mode string) []float64 {
	// Clear accumulator
	for i := range macro.Accumulator {
		macro.Accumulator[i] = 0
	}

	// Process through each crossbar
	for _, cb := range macro.Crossbars {
		var partial []float64

		switch mode {
		case "Analog":
			partial = cb.ComputeAnalogMVM(inputs)
		case "BitSerial":
			partial = cb.ComputeBitSerialMVM(inputs)
		case "Hybrid":
			partial = cb.ComputeHybridMVM(inputs)
		default:
			partial = cb.ComputeHybridMVM(inputs)
		}

		// Accumulate
		for i := range partial {
			if i < len(macro.Accumulator) {
				macro.Accumulator[i] += partial[i]
			}
		}

		macro.TotalOps += cb.AnalogMACs + cb.DigitalMACs
		macro.EnergyConsumed += cb.TotalEnergy
	}

	// Quantize output to specified precision
	outputs := make([]float64, len(macro.Accumulator))
	outputLevels := float64(1 << macro.CurrentOutputPrecision)
	for i, v := range macro.Accumulator {
		outputs[i] = math.Round(v*outputLevels) / outputLevels
	}

	return outputs
}

// GetEfficiency returns energy efficiency metrics
func (macro *MixedPrecisionCIMMacro) GetEfficiency() map[string]float64 {
	totalAnalog := int64(0)
	totalDigital := int64(0)
	totalADC := int64(0)
	totalEnergy := 0.0

	for _, cb := range macro.Crossbars {
		totalAnalog += cb.AnalogMACs
		totalDigital += cb.DigitalMACs
		totalADC += cb.ADCConversions
		totalEnergy += cb.TotalEnergy
	}

	topsw := 0.0
	if totalEnergy > 0 {
		topsw = float64(totalAnalog+totalDigital) / totalEnergy * 1e3
	}

	return map[string]float64{
		"analog_macs":     float64(totalAnalog),
		"digital_macs":    float64(totalDigital),
		"adc_conversions": float64(totalADC),
		"total_energy":    totalEnergy,
		"tops_w":          topsw,
	}
}

// =============================================================================
// INTEGRATED NC-FET HYBRID CIM SYSTEM
// =============================================================================

// NCFETHybridCIMSystem combines NC-FET and hybrid CIM
type NCFETHybridCIMSystem struct {
	// NC-FET array for ultra-low power
	NCFETArray *NCFETCIMArray

	// Hybrid macro for precision
	HybridMacro *MixedPrecisionCIMMacro

	// Configuration
	UseNCFET     bool
	UseHybrid    bool
	ComputeMode  string

	// Performance targets
	TargetEfficiency float64 // TOPS/W
	TargetAccuracy   float64 // %
}

// NewNCFETHybridCIMSystem creates an integrated system
func NewNCFETHybridCIMSystem(rows, cols int) *NCFETHybridCIMSystem {
	ncfetConfig := DefaultNCFETConfig()
	hybridConfig := DefaultHybridCIMConfig()
	hybridConfig.CrossbarRows = rows
	hybridConfig.CrossbarCols = cols

	return &NCFETHybridCIMSystem{
		NCFETArray:       NewNCFETCIMArray(ncfetConfig, rows, cols),
		HybridMacro:      NewMixedPrecisionCIMMacro(hybridConfig),
		UseNCFET:         true,
		UseHybrid:        true,
		ComputeMode:      "Hybrid",
		TargetEfficiency: 100.0, // 100 TOPS/W target
		TargetAccuracy:   99.0,  // 99% target
	}
}

// LoadWeights loads weights into both arrays
func (sys *NCFETHybridCIMSystem) LoadWeights(weights [][]float64) {
	sys.NCFETArray.LoadWeights(weights)
	sys.HybridMacro.LoadModel([][][]float64{weights})
}

// Inference performs inference using configured mode
func (sys *NCFETHybridCIMSystem) Inference(inputs []float64) []float64 {
	if sys.UseNCFET && !sys.UseHybrid {
		// NC-FET only: ultra-low power
		return sys.NCFETArray.ComputeMVM(inputs)
	} else if sys.UseHybrid && !sys.UseNCFET {
		// Hybrid only: high precision
		return sys.HybridMacro.Compute(inputs, sys.ComputeMode)
	} else {
		// Combined: use NC-FET for low-precision, hybrid for accumulation
		// First pass with NC-FET
		ncfetResult := sys.NCFETArray.ComputeMVM(inputs)

		// Refine with hybrid if needed
		// Check if result needs higher precision
		maxVal := 0.0
		for _, v := range ncfetResult {
			if math.Abs(v) > maxVal {
				maxVal = math.Abs(v)
			}
		}

		// If values are in high dynamic range, use hybrid
		if maxVal > 0.9 || maxVal < 0.1 {
			return sys.HybridMacro.Compute(inputs, "Hybrid")
		}

		return ncfetResult
	}
}

// GetPerformanceReport returns system performance summary
func (sys *NCFETHybridCIMSystem) GetPerformanceReport() string {
	ncfetEff := sys.NCFETArray.GetEnergyEfficiency()
	hybridEff := sys.HybridMacro.GetEfficiency()

	return fmt.Sprintf(`
NC-FET Hybrid CIM System Performance
=====================================
NC-FET Array:
  Size: %d × %d
  Total MACs: %d
  Energy Efficiency: %.2f TOPS/W
  Subthreshold Swing: %.2f mV/dec

Hybrid CIM Macro:
  Crossbars: %d
  Precision: %d/%d/%d (I/W/O)
  Total Energy: %.2f
  Efficiency: %.2f TOPS/W

Configuration:
  Use NC-FET: %v
  Use Hybrid: %v
  Compute Mode: %s

NC-FET Advantages:
  - 22.77× energy reduction at 0.3V
  - 12.41× energy reduction at 0.5V
  - SS = %.1f mV/dec (below Boltzmann limit)

Hybrid CIM Advantages:
  - ADC-less mode: 28%% energy reduction
  - Configurable 1-16 bit precision
  - Digital scale factor processing
`,
		sys.NCFETArray.Rows, sys.NCFETArray.Cols,
		sys.NCFETArray.TotalMACs,
		ncfetEff,
		sys.NCFETArray.Config.SubthresholdSwing,
		len(sys.HybridMacro.Crossbars),
		sys.HybridMacro.CurrentInputPrecision,
		sys.HybridMacro.CurrentWeightPrecision,
		sys.HybridMacro.CurrentOutputPrecision,
		hybridEff["total_energy"],
		hybridEff["tops_w"],
		sys.UseNCFET,
		sys.UseHybrid,
		sys.ComputeMode,
		sys.NCFETArray.Config.SubthresholdSwing,
	)
}

// =============================================================================
// IRONLATTICE NC-FET HYBRID SYSTEM
// =============================================================================

// IronLatticeNCFETHybrid represents IronLattice-optimized system
type IronLatticeNCFETHybrid struct {
	// Core system
	System *NCFETHybridCIMSystem

	// HZO ferroelectric parameters
	HZOParameters struct {
		FEThickness     float64 // nm
		VoltageGain     float64 // Internal amplification
		SubthresholdSS  float64 // mV/dec
		BodyFactor      float64 // < 1 for NC
	}

	// Performance achievements
	EnergyReduction float64 // × vs baseline CMOS
	SpeedEnhancement float64 // × vs baseline
}

// NewIronLatticeNCFETHybrid creates an IronLattice NC-FET hybrid system
func NewIronLatticeNCFETHybrid(rows, cols int) *IronLatticeNCFETHybrid {
	system := &IronLatticeNCFETHybrid{
		System: NewNCFETHybridCIMSystem(rows, cols),
	}

	// Configure HZO parameters
	system.HZOParameters.FEThickness = 12.0    // 12 nm HZO
	system.HZOParameters.VoltageGain = 28.0    // 28× demonstrated
	system.HZOParameters.SubthresholdSS = 8.0  // 8 mV/dec
	system.HZOParameters.BodyFactor = 0.5      // Below 1

	// Performance based on research
	system.EnergyReduction = 22.77  // At 0.3V
	system.SpeedEnhancement = 18.0  // 18× speed improvement

	return system
}

// ExportJSON exports system configuration
func (ils *IronLatticeNCFETHybrid) ExportJSON() ([]byte, error) {
	export := map[string]interface{}{
		"hzo_parameters": map[string]float64{
			"fe_thickness_nm":     ils.HZOParameters.FEThickness,
			"voltage_gain":        ils.HZOParameters.VoltageGain,
			"subthreshold_ss_mv":  ils.HZOParameters.SubthresholdSS,
			"body_factor":         ils.HZOParameters.BodyFactor,
		},
		"performance": map[string]float64{
			"energy_reduction_x":   ils.EnergyReduction,
			"speed_enhancement_x":  ils.SpeedEnhancement,
		},
		"ncfet_array": map[string]interface{}{
			"rows":       ils.System.NCFETArray.Rows,
			"cols":       ils.System.NCFETArray.Cols,
			"total_macs": ils.System.NCFETArray.TotalMACs,
		},
		"hybrid_config": map[string]interface{}{
			"compute_mode":      ils.System.ComputeMode,
			"input_precision":   ils.System.HybridMacro.CurrentInputPrecision,
			"weight_precision":  ils.System.HybridMacro.CurrentWeightPrecision,
			"output_precision":  ils.System.HybridMacro.CurrentOutputPrecision,
		},
	}

	return json.MarshalIndent(export, "", "  ")
}
