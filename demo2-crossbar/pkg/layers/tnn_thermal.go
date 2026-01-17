// Package layers provides neural network layer implementations for CIM simulation.
// This file implements Ternary Neural Networks (TNN) and thermal-aware scheduling.
//
// Ternary Neural Networks:
// - Weights constrained to {-1, 0, +1} (1.58 bits)
// - Multiplication-free inference via add/subtract/zero-skip
// - BitNet b1.58 compatibility (Microsoft, 2024)
// - FeFET TNN arrays: 1T, 2T-diff, 2T-PUPD configurations
//
// Thermal-Aware Scheduling:
// - Temperature modeling for crossbar arrays
// - Conductance drift compensation
// - Hot/cold bank classification (THOR)
// - DoRA-based calibration for drift recovery
//
// References:
// - BitNet b1.58 (arXiv 2402.17764, Microsoft 2024)
// - 1.58b FeFET TNN (ResearchGate 2024)
// - ReRAM Thermal Heating Challenges (IEEE OJCAS 2024)
// - DoRA Calibration for RRAM (arXiv 2504.03763)
// - Temperature-Resilient FeFET CIM (arXiv 2312.17442)

package layers

import (
	"fmt"
	"math"
	"sort"
)

// ============================================================================
// Ternary Neural Network (TNN) Types
// ============================================================================

// TernaryValue represents a ternary weight value.
type TernaryValue int8

const (
	TERNARY_NEG  TernaryValue = -1 // -1
	TERNARY_ZERO TernaryValue = 0  // 0
	TERNARY_POS  TernaryValue = 1  // +1
)

// TernaryWeight stores packed ternary weights using 2 bits per weight.
// Encoding: 00 = -1, 01 = 0, 10 = +1, 11 = reserved
type TernaryWeight struct {
	Data      []byte  // Packed ternary weights (4 weights per byte)
	NumWeights int    // Total number of weights
	Rows      int     // Number of output neurons
	Cols      int     // Number of input features
	Scale     float64 // Scaling factor (absmean)
}

// NewTernaryWeight creates a new packed ternary weight matrix.
func NewTernaryWeight(rows, cols int) *TernaryWeight {
	numWeights := rows * cols
	numBytes := (numWeights + 3) / 4 // 4 weights per byte
	return &TernaryWeight{
		Data:       make([]byte, numBytes),
		NumWeights: numWeights,
		Rows:       rows,
		Cols:       cols,
		Scale:      1.0,
	}
}

// SetWeight sets a ternary weight at position (row, col).
func (tw *TernaryWeight) SetWeight(row, col int, value TernaryValue) {
	idx := row*tw.Cols + col
	byteIdx := idx / 4
	bitOffset := uint((idx % 4) * 2)

	// Clear existing bits
	tw.Data[byteIdx] &^= (0x03 << bitOffset)

	// Encode ternary value
	var encoded byte
	switch value {
	case TERNARY_NEG:
		encoded = 0x00
	case TERNARY_ZERO:
		encoded = 0x01
	case TERNARY_POS:
		encoded = 0x02
	}

	tw.Data[byteIdx] |= (encoded << bitOffset)
}

// GetWeight retrieves the ternary weight at position (row, col).
func (tw *TernaryWeight) GetWeight(row, col int) TernaryValue {
	idx := row*tw.Cols + col
	byteIdx := idx / 4
	bitOffset := uint((idx % 4) * 2)

	encoded := (tw.Data[byteIdx] >> bitOffset) & 0x03

	switch encoded {
	case 0x00:
		return TERNARY_NEG
	case 0x01:
		return TERNARY_ZERO
	case 0x02:
		return TERNARY_POS
	default:
		return TERNARY_ZERO
	}
}

// GetFloat returns the floating-point value of the weight.
func (tw *TernaryWeight) GetFloat(row, col int) float64 {
	return float64(tw.GetWeight(row, col)) * tw.Scale
}

// Sparsity returns the fraction of zero weights.
func (tw *TernaryWeight) Sparsity() float64 {
	var zeros int
	for i := 0; i < tw.Rows; i++ {
		for j := 0; j < tw.Cols; j++ {
			if tw.GetWeight(i, j) == TERNARY_ZERO {
				zeros++
			}
		}
	}
	return float64(zeros) / float64(tw.NumWeights)
}

// ============================================================================
// Ternary Quantization (BitNet b1.58)
// ============================================================================

// TernaryQuantizer implements absmean quantization for ternary weights.
type TernaryQuantizer struct {
	// Configuration
	Threshold float64 // Threshold for zero (typically 0.5 * absmean)

	// Statistics
	OriginalScale float64
	Sparsity      float64
}

// NewTernaryQuantizer creates a new ternary quantizer.
func NewTernaryQuantizer() *TernaryQuantizer {
	return &TernaryQuantizer{
		Threshold: 0.5,
	}
}

// Quantize converts float weights to ternary using absmean quantization.
// w_ternary = round_ternary(w / absmean(w))
func (tq *TernaryQuantizer) Quantize(floatWeights [][]float64) *TernaryWeight {
	rows := len(floatWeights)
	cols := 0
	if rows > 0 {
		cols = len(floatWeights[0])
	}

	tw := NewTernaryWeight(rows, cols)

	// Calculate absmean (γ)
	var sum float64
	count := 0
	for _, row := range floatWeights {
		for _, w := range row {
			sum += math.Abs(w)
			count++
		}
	}

	if count > 0 {
		tw.Scale = sum / float64(count)
		tq.OriginalScale = tw.Scale
	}

	// Quantize to ternary
	threshold := tq.Threshold * tw.Scale
	var zeros int

	for i, row := range floatWeights {
		for j, w := range row {
			if math.Abs(w) < threshold {
				tw.SetWeight(i, j, TERNARY_ZERO)
				zeros++
			} else if w > 0 {
				tw.SetWeight(i, j, TERNARY_POS)
			} else {
				tw.SetWeight(i, j, TERNARY_NEG)
			}
		}
	}

	tq.Sparsity = float64(zeros) / float64(count)
	return tw
}

// QuantizeActivations quantizes activations using absmax (8-bit per token).
func (tq *TernaryQuantizer) QuantizeActivations(activations []float64, bits int) ([]int8, float64) {
	// Find absmax
	var absmax float64
	for _, a := range activations {
		if abs := math.Abs(a); abs > absmax {
			absmax = abs
		}
	}

	// Quantize
	levels := float64(int(1) << (bits - 1)) // e.g., 128 for 8-bit
	scale := absmax / levels
	if scale == 0 {
		scale = 1.0
	}

	quantized := make([]int8, len(activations))
	for i, a := range activations {
		q := math.Round(a / scale)
		q = math.Max(-levels, math.Min(levels-1, q))
		quantized[i] = int8(q)
	}

	return quantized, scale
}

// ============================================================================
// BitLinear Layer (Multiplication-Free)
// ============================================================================

// BitLinearConfig configures a BitLinear layer.
type BitLinearConfig struct {
	InputSize    int
	OutputSize   int
	ActBits      int     // Activation bits (default 8)
	ZeroSkip     bool    // Skip zero weights (sparsity exploitation)
	UseRMSNorm   bool    // Apply RMSNorm before quantization
	Eps          float64 // RMSNorm epsilon
}

// DefaultBitLinearConfig returns default BitLinear configuration.
func DefaultBitLinearConfig(inputSize, outputSize int) *BitLinearConfig {
	return &BitLinearConfig{
		InputSize:  inputSize,
		OutputSize: outputSize,
		ActBits:    8,
		ZeroSkip:   true,
		UseRMSNorm: true,
		Eps:        1e-5,
	}
}

// BitLinear implements a multiplication-free ternary linear layer.
type BitLinear struct {
	config     *BitLinearConfig
	weights    *TernaryWeight
	quantizer  *TernaryQuantizer

	// Statistics
	TotalOps   int64   // Number of add/subtract operations
	SkippedOps int64   // Operations skipped due to zero weights
	Throughput float64 // GOPS
}

// NewBitLinear creates a new BitLinear layer.
func NewBitLinear(config *BitLinearConfig) *BitLinear {
	return &BitLinear{
		config:    config,
		quantizer: NewTernaryQuantizer(),
	}
}

// SetWeights initializes weights from float values.
func (bl *BitLinear) SetWeights(floatWeights [][]float64) {
	bl.weights = bl.quantizer.Quantize(floatWeights)
}

// Forward performs multiplication-free forward pass.
// For ternary weights: y = Σ(w_i * x_i) where w_i ∈ {-1, 0, +1}
// This becomes: y = Σ(x_i where w_i=+1) - Σ(x_i where w_i=-1)
func (bl *BitLinear) Forward(input []float64) []float64 {
	// Optional RMSNorm
	if bl.config.UseRMSNorm {
		input = bl.rmsNorm(input)
	}

	// Quantize activations
	quantizedAct, actScale := bl.quantizer.QuantizeActivations(input, bl.config.ActBits)

	// Multiplication-free matmul
	output := make([]float64, bl.config.OutputSize)

	for i := 0; i < bl.config.OutputSize; i++ {
		var sum int32
		for j := 0; j < bl.config.InputSize && j < len(quantizedAct); j++ {
			w := bl.weights.GetWeight(i, j)

			if bl.config.ZeroSkip && w == TERNARY_ZERO {
				bl.SkippedOps++
				continue
			}

			// Add or subtract based on ternary weight
			switch w {
			case TERNARY_POS:
				sum += int32(quantizedAct[j])
			case TERNARY_NEG:
				sum -= int32(quantizedAct[j])
			}
			bl.TotalOps++
		}

		// Rescale output
		output[i] = float64(sum) * actScale * bl.weights.Scale
	}

	return output
}

// rmsNorm applies RMSNorm to input.
func (bl *BitLinear) rmsNorm(x []float64) []float64 {
	var sumSq float64
	for _, v := range x {
		sumSq += v * v
	}

	rms := math.Sqrt(sumSq/float64(len(x)) + bl.config.Eps)

	normalized := make([]float64, len(x))
	for i, v := range x {
		normalized[i] = v / rms
	}

	return normalized
}

// GetEfficiency returns compute efficiency metrics.
func (bl *BitLinear) GetEfficiency() (float64, float64) {
	total := bl.TotalOps + bl.SkippedOps
	if total == 0 {
		return 0, 0
	}

	// Sparsity-based efficiency
	sparsityGain := float64(bl.SkippedOps) / float64(total) * 100.0

	// Energy efficiency estimate (add vs. multiply)
	// Ternary add/sub ~5x more efficient than FP multiply
	energyGain := 5.0 * (1.0 - float64(bl.SkippedOps)/float64(total))

	return sparsityGain, energyGain
}

// ============================================================================
// FeFET TNN Array Configurations
// ============================================================================

// FeFETTNNConfig enumerates FeFET TNN array designs.
type FeFETTNNDesign int

const (
	FEFET_1T_158B   FeFETTNNDesign = iota // 1T 1.58b FeFET
	FEFET_2T_DIFF                         // 2T differential
	FEFET_2T_PUPD                         // 2T pull-up/pull-down
)

// FeFETTNNConfig configures FeFET-based TNN arrays.
type FeFETTNNConfig struct {
	Design         FeFETTNNDesign
	ArrayRows      int
	ArrayCols      int
	TechNode       int     // Technology node in nm (e.g., 7)
	ReadLatency    float64 // ns
	WriteLatency   float64 // ns
	EnergyPerOp    float64 // fJ per ternary op
	UseWT          bool    // Weight transformation for robustness
	UseWIT         bool    // Weight-input transformation
}

// DefaultFeFETTNNConfig returns default FeFET TNN configuration.
func DefaultFeFETTNNConfig() *FeFETTNNConfig {
	return &FeFETTNNConfig{
		Design:       FEFET_2T_DIFF,
		ArrayRows:    256,
		ArrayCols:    256,
		TechNode:     7,
		ReadLatency:  5.0,
		WriteLatency: 50.0,
		EnergyPerOp:  3.0,
		UseWT:        true,
		UseWIT:       false,
	}
}

// FeFETTNNCell represents a FeFET ternary synapse cell.
type FeFETTNNCell struct {
	Weight      TernaryValue
	PosState    float64 // Positive FeFET conductance
	NegState    float64 // Negative FeFET conductance (2T designs)
	Variance    float64 // Device variation
	Temperature float64 // Current temperature (K)
}

// FeFETTNNArray simulates a FeFET-based TNN crossbar array.
type FeFETTNNArray struct {
	config *FeFETTNNConfig
	cells  [][]*FeFETTNNCell

	// Temperature state
	Temperature float64 // Average array temperature (K)
	Hotspots    [][]float64

	// Performance metrics
	TotalOps    int64
	TotalEnergy float64 // pJ
	Throughput  float64 // TOPS
	Efficiency  float64 // TOPS/W
}

// NewFeFETTNNArray creates a new FeFET TNN array.
func NewFeFETTNNArray(config *FeFETTNNConfig) *FeFETTNNArray {
	cells := make([][]*FeFETTNNCell, config.ArrayRows)
	hotspots := make([][]float64, config.ArrayRows)

	for i := range cells {
		cells[i] = make([]*FeFETTNNCell, config.ArrayCols)
		hotspots[i] = make([]float64, config.ArrayCols)
		for j := range cells[i] {
			cells[i][j] = &FeFETTNNCell{
				Weight:      TERNARY_ZERO,
				PosState:    0.0,
				NegState:    0.0,
				Variance:    0.02,
				Temperature: 300.0, // Room temperature
			}
			hotspots[i][j] = 300.0
		}
	}

	return &FeFETTNNArray{
		config:      config,
		cells:       cells,
		Temperature: 300.0,
		Hotspots:    hotspots,
	}
}

// ProgramWeights programs ternary weights into the array.
func (arr *FeFETTNNArray) ProgramWeights(weights *TernaryWeight) error {
	if weights.Rows > arr.config.ArrayRows || weights.Cols > arr.config.ArrayCols {
		return fmt.Errorf("weight matrix (%d×%d) exceeds array size (%d×%d)",
			weights.Rows, weights.Cols, arr.config.ArrayRows, arr.config.ArrayCols)
	}

	for i := 0; i < weights.Rows; i++ {
		for j := 0; j < weights.Cols; j++ {
			w := weights.GetWeight(i, j)
			arr.cells[i][j].Weight = w

			// Set conductance states based on design
			switch arr.config.Design {
			case FEFET_1T_158B:
				// Single FeFET with 3 states
				switch w {
				case TERNARY_NEG:
					arr.cells[i][j].PosState = 0.1 // Low
				case TERNARY_ZERO:
					arr.cells[i][j].PosState = 0.5 // Mid
				case TERNARY_POS:
					arr.cells[i][j].PosState = 0.9 // High
				}

			case FEFET_2T_DIFF, FEFET_2T_PUPD:
				// Differential: pos and neg FeFETs
				switch w {
				case TERNARY_NEG:
					arr.cells[i][j].PosState = 0.1
					arr.cells[i][j].NegState = 0.9
				case TERNARY_ZERO:
					arr.cells[i][j].PosState = 0.5
					arr.cells[i][j].NegState = 0.5
				case TERNARY_POS:
					arr.cells[i][j].PosState = 0.9
					arr.cells[i][j].NegState = 0.1
				}
			}
		}
	}

	return nil
}

// TernaryCompute performs in-memory ternary computation.
func (arr *FeFETTNNArray) TernaryCompute(input []int8, outputRows int) []int32 {
	results := make([]int32, outputRows)

	for row := 0; row < outputRows; row++ {
		var sum int32
		for col := 0; col < len(input) && col < arr.config.ArrayCols; col++ {
			cell := arr.cells[row][col]

			// Apply weight transformation if enabled
			effectiveWeight := cell.Weight
			if arr.config.UseWT {
				effectiveWeight = arr.applyWeightTransform(cell, col)
			}

			// Ternary MAC (add/subtract/skip)
			switch effectiveWeight {
			case TERNARY_POS:
				sum += int32(input[col])
			case TERNARY_NEG:
				sum -= int32(input[col])
			}

			// Update thermal state
			arr.updateThermal(row, col)
		}

		results[row] = sum
		arr.TotalOps += int64(len(input))
	}

	// Update energy
	arr.TotalEnergy += float64(outputRows*len(input)) * arr.config.EnergyPerOp / 1000.0

	return results
}

// applyWeightTransform applies robustness transformation.
func (arr *FeFETTNNArray) applyWeightTransform(cell *FeFETTNNCell, col int) TernaryValue {
	// Weight transformation for robustness against variation
	// Maps weights to reduce impact of device non-idealities

	// Simple transformation: flip polarity for odd columns
	if arr.config.UseWT && col%2 == 1 {
		switch cell.Weight {
		case TERNARY_POS:
			return TERNARY_NEG
		case TERNARY_NEG:
			return TERNARY_POS
		}
	}

	return cell.Weight
}

// updateThermal updates thermal state after computation.
func (arr *FeFETTNNArray) updateThermal(row, col int) {
	// Simple thermal model: heat generation proportional to switching
	cell := arr.cells[row][col]
	if cell.Weight != TERNARY_ZERO {
		// Active cell generates heat
		heatGen := 0.1 // K per operation
		arr.Hotspots[row][col] += heatGen

		// Thermal dissipation
		dissipation := 0.05
		arr.Hotspots[row][col] = math.Max(300.0,
			arr.Hotspots[row][col]-dissipation)

		cell.Temperature = arr.Hotspots[row][col]
	}
}

// EstimatePerformance calculates performance metrics.
func (arr *FeFETTNNArray) EstimatePerformance(batchSize int) {
	opsPerInference := int64(arr.config.ArrayRows) * int64(arr.config.ArrayCols)
	totalOps := opsPerInference * int64(batchSize)
	latency := float64(batchSize) * arr.config.ReadLatency * 1e-9 // seconds

	arr.Throughput = float64(totalOps) / latency / 1e12 // TOPS

	// Energy efficiency
	totalEnergy := float64(totalOps) * arr.config.EnergyPerOp * 1e-15 // Joules
	power := totalEnergy / latency
	if power > 0 {
		arr.Efficiency = float64(totalOps) / 1e12 / power // TOPS/W
	}
}

// ============================================================================
// Thermal-Aware Scheduling
// ============================================================================

// ThermalConfig configures thermal management.
type ThermalConfig struct {
	AmbientTemp     float64 // Ambient temperature (K)
	MaxTemp         float64 // Maximum safe temperature (K)
	HotThreshold    float64 // Hot bank threshold (K)
	ColdThreshold   float64 // Cold bank threshold (K)
	CoolingRate     float64 // Cooling rate (K/ms)
	HeatingPerOp    float64 // Heat generated per operation (K)
	NumBanks        int     // Number of memory banks
	NumSensors      int     // Temperature sensors per bank
}

// DefaultThermalConfig returns default thermal configuration.
func DefaultThermalConfig() *ThermalConfig {
	return &ThermalConfig{
		AmbientTemp:   300.0,  // 27°C
		MaxTemp:       358.0,  // 85°C
		HotThreshold:  330.0,  // 57°C
		ColdThreshold: 310.0,  // 37°C
		CoolingRate:   0.5,    // K/ms
		HeatingPerOp:  0.001,  // K per operation
		NumBanks:      8,
		NumSensors:    4,
	}
}

// MemoryBank represents a CIM memory bank with thermal state.
type MemoryBank struct {
	ID          int
	Temperature float64
	IsHot       bool
	Utilization float64 // Current utilization
	PendingOps  int     // Pending operations
	LastAccess  int64   // Timestamp of last access
}

// ThermalScheduler implements thermal-aware task scheduling.
type ThermalScheduler struct {
	config       *ThermalConfig
	banks        []*MemoryBank
	currentTime  int64 // Simulated time in ms

	// Statistics
	DelayedOps   int64
	ThrottledOps int64
	TotalOps     int64
}

// NewThermalScheduler creates a new thermal scheduler.
func NewThermalScheduler(config *ThermalConfig) *ThermalScheduler {
	banks := make([]*MemoryBank, config.NumBanks)
	for i := range banks {
		banks[i] = &MemoryBank{
			ID:          i,
			Temperature: config.AmbientTemp,
			IsHot:       false,
		}
	}

	return &ThermalScheduler{
		config: config,
		banks:  banks,
	}
}

// ScheduleOperation schedules an operation considering thermal constraints.
func (ts *ThermalScheduler) ScheduleOperation(ops int, preferredBank int) (int, int64) {
	ts.TotalOps += int64(ops)

	// Update temperatures
	ts.updateTemperatures()

	// Classify banks
	ts.classifyBanks()

	// Find best bank
	targetBank := ts.selectBank(preferredBank)
	delay := int64(0)

	// Check if bank is too hot
	if ts.banks[targetBank].IsHot {
		// Calculate delay needed for cooling
		tempDiff := ts.banks[targetBank].Temperature - ts.config.ColdThreshold
		delay = int64(tempDiff / ts.config.CoolingRate)
		ts.DelayedOps += int64(ops)
	}

	// Execute operation
	ts.banks[targetBank].PendingOps += ops
	ts.banks[targetBank].LastAccess = ts.currentTime
	ts.banks[targetBank].Temperature += float64(ops) * ts.config.HeatingPerOp

	// Check throttling
	if ts.banks[targetBank].Temperature > ts.config.MaxTemp {
		ts.ThrottledOps += int64(ops)
	}

	return targetBank, delay
}

// updateTemperatures updates bank temperatures based on cooling.
func (ts *ThermalScheduler) updateTemperatures() {
	for _, bank := range ts.banks {
		// Natural cooling towards ambient
		if bank.Temperature > ts.config.AmbientTemp {
			cooling := ts.config.CoolingRate * 0.1 // Per update cycle
			bank.Temperature = math.Max(ts.config.AmbientTemp,
				bank.Temperature-cooling)
		}
	}
}

// classifyBanks classifies banks as hot or cold.
func (ts *ThermalScheduler) classifyBanks() {
	for _, bank := range ts.banks {
		bank.IsHot = bank.Temperature > ts.config.HotThreshold
	}
}

// selectBank selects the best bank for operation.
func (ts *ThermalScheduler) selectBank(preferred int) int {
	// If preferred bank is cold, use it
	if !ts.banks[preferred].IsHot {
		return preferred
	}

	// Find coldest bank
	coldest := preferred
	minTemp := ts.banks[preferred].Temperature

	for i, bank := range ts.banks {
		if bank.Temperature < minTemp {
			minTemp = bank.Temperature
			coldest = i
		}
	}

	return coldest
}

// GetBankStatus returns status of all banks.
func (ts *ThermalScheduler) GetBankStatus() []map[string]interface{} {
	status := make([]map[string]interface{}, len(ts.banks))
	for i, bank := range ts.banks {
		status[i] = map[string]interface{}{
			"id":          bank.ID,
			"temperature": bank.Temperature,
			"is_hot":      bank.IsHot,
			"utilization": bank.Utilization,
		}
	}
	return status
}

// AdvanceTime advances the simulation time.
func (ts *ThermalScheduler) AdvanceTime(ms int64) {
	ts.currentTime += ms
	ts.updateTemperatures()
}

// ============================================================================
// Conductance Drift Model
// ============================================================================

// DriftModel models conductance drift in ReRAM/PCM devices.
type DriftModel struct {
	DriftCoefficient float64 // Drift coefficient (ν)
	InitialTime      float64 // Reference time (t0) in seconds
	Temperature      float64 // Temperature (K)
	ActivationEnergy float64 // Activation energy (eV)
}

// NewDriftModel creates a new drift model.
func NewDriftModel() *DriftModel {
	return &DriftModel{
		DriftCoefficient: 0.1,   // Typical for PCM
		InitialTime:      1.0,   // 1 second
		Temperature:      300.0, // Room temperature
		ActivationEnergy: 0.3,   // eV for ReRAM
	}
}

// CalculateDrift calculates conductance after drift.
// G(t) = G0 × (t/t0)^(-ν)
func (dm *DriftModel) CalculateDrift(initialConductance float64, timeSeconds float64) float64 {
	if timeSeconds <= dm.InitialTime {
		return initialConductance
	}

	// Power-law drift
	ratio := timeSeconds / dm.InitialTime
	driftFactor := math.Pow(ratio, -dm.DriftCoefficient)

	return initialConductance * driftFactor
}

// CalculateTempDependentDrift calculates temperature-dependent drift.
func (dm *DriftModel) CalculateTempDependentDrift(initialConductance, timeSeconds, temperature float64) float64 {
	// Arrhenius temperature dependence
	kB := 8.617e-5 // Boltzmann constant in eV/K

	// Effective drift coefficient increases with temperature
	tempFactor := math.Exp(-dm.ActivationEnergy / (kB * temperature))
	effectiveDrift := dm.DriftCoefficient * (1.0 + tempFactor)

	if timeSeconds <= dm.InitialTime {
		return initialConductance
	}

	ratio := timeSeconds / dm.InitialTime
	driftFactor := math.Pow(ratio, -effectiveDrift)

	return initialConductance * driftFactor
}

// ============================================================================
// DoRA-Based Calibration
// ============================================================================

// DoRAConfig configures DoRA (Low-Rank Adaptation) calibration.
type DoRAConfig struct {
	Rank           int     // Low-rank dimension
	Alpha          float64 // Scaling factor
	CalibSamples   int     // Number of calibration samples
	UpdateFraction float64 // Fraction of parameters to update
}

// DefaultDoRAConfig returns default DoRA configuration.
func DefaultDoRAConfig() *DoRAConfig {
	return &DoRAConfig{
		Rank:           8,
		Alpha:          16.0,
		CalibSamples:   10,
		UpdateFraction: 0.0234, // 2.34% of parameters
	}
}

// DoRACalibrator implements DoRA-based drift calibration.
type DoRACalibrator struct {
	config *DoRAConfig

	// Low-rank matrices (stored in SRAM)
	A [][]float64 // Down-projection (d × r)
	B [][]float64 // Up-projection (r × d)

	// Original weight statistics
	OriginalNorm float64

	// Calibration state
	IsCalibrated bool
	AccuracyGain float64
}

// NewDoRACalibrator creates a new DoRA calibrator.
func NewDoRACalibrator(config *DoRAConfig) *DoRACalibrator {
	return &DoRACalibrator{
		config: config,
	}
}

// Initialize initializes DoRA matrices for a weight layer.
func (dc *DoRACalibrator) Initialize(inputDim, outputDim int) {
	rank := dc.config.Rank

	// Initialize A (down-projection)
	dc.A = make([][]float64, inputDim)
	for i := range dc.A {
		dc.A[i] = make([]float64, rank)
		// Kaiming initialization
		scale := math.Sqrt(2.0 / float64(inputDim))
		for j := range dc.A[i] {
			dc.A[i][j] = (float64(i*rank+j)/float64(inputDim*rank) - 0.5) * scale
		}
	}

	// Initialize B (up-projection) to zeros
	dc.B = make([][]float64, rank)
	for i := range dc.B {
		dc.B[i] = make([]float64, outputDim)
	}
}

// Calibrate performs calibration using sample data.
func (dc *DoRACalibrator) Calibrate(driftedWeights [][]float64, targetOutputs [][]float64, inputs [][]float64) {
	if len(inputs) == 0 || len(targetOutputs) == 0 {
		return
	}

	// Simple gradient-based calibration
	lr := 0.01

	for sample := 0; sample < dc.config.CalibSamples && sample < len(inputs); sample++ {
		input := inputs[sample]
		target := targetOutputs[sample]

		// Forward with drift compensation
		output := dc.ForwardWithCompensation(driftedWeights, input)

		// Compute error
		for i := range output {
			if i < len(target) {
				error := target[i] - output[i]

				// Update B (simplified gradient)
				for r := range dc.B {
					for j := range dc.B[r] {
						if j < len(input) {
							dc.B[r][j] += lr * error * dc.A[j][r]
						}
					}
				}
			}
		}
	}

	dc.IsCalibrated = true
}

// ForwardWithCompensation applies weight with DoRA compensation.
// W_effective = W_drifted + (alpha/rank) × B × A
func (dc *DoRACalibrator) ForwardWithCompensation(driftedWeights [][]float64, input []float64) []float64 {
	outputDim := len(driftedWeights)
	inputDim := len(input)

	output := make([]float64, outputDim)

	// Original drifted computation
	for i := 0; i < outputDim && i < len(driftedWeights); i++ {
		for j := 0; j < inputDim && j < len(driftedWeights[i]); j++ {
			output[i] += driftedWeights[i][j] * input[j]
		}
	}

	// Add DoRA compensation: (alpha/rank) × (B × A) × x
	if dc.IsCalibrated || len(dc.A) > 0 {
		scale := dc.config.Alpha / float64(dc.config.Rank)

		// Compute A × x
		Ax := make([]float64, dc.config.Rank)
		for r := 0; r < dc.config.Rank; r++ {
			for j := 0; j < inputDim && j < len(dc.A); j++ {
				if r < len(dc.A[j]) {
					Ax[r] += dc.A[j][r] * input[j]
				}
			}
		}

		// Compute B × (A × x) and add to output
		for i := 0; i < outputDim; i++ {
			for r := 0; r < dc.config.Rank && r < len(dc.B); r++ {
				if i < len(dc.B[r]) {
					output[i] += scale * dc.B[r][i] * Ax[r]
				}
			}
		}
	}

	return output
}

// GetOverhead returns memory overhead of DoRA calibration.
func (dc *DoRACalibrator) GetOverhead(totalParams int) float64 {
	doraParams := len(dc.A)*dc.config.Rank + dc.config.Rank*len(dc.B)
	return float64(doraParams) / float64(totalParams) * 100.0
}

// ============================================================================
// Temperature Compensation
// ============================================================================

// TempCompensator implements temperature compensation for CIM arrays.
type TempCompensator struct {
	// Calibration data at reference temperatures
	RefTemps    []float64            // Reference temperatures (K)
	ScaleFactors map[float64]float64 // Scale factors per temperature

	// Interpolation
	Interpolated bool
}

// NewTempCompensator creates a new temperature compensator.
func NewTempCompensator() *TempCompensator {
	return &TempCompensator{
		RefTemps:     []float64{300.0, 320.0, 340.0, 360.0}, // 27-87°C
		ScaleFactors: make(map[float64]float64),
	}
}

// Calibrate calibrates compensation at a reference temperature.
func (tc *TempCompensator) Calibrate(temperature float64, measuredScale float64) {
	tc.ScaleFactors[temperature] = measuredScale
}

// GetCompensation returns compensation factor for a temperature.
func (tc *TempCompensator) GetCompensation(temperature float64) float64 {
	// Check exact match
	if scale, ok := tc.ScaleFactors[temperature]; ok {
		return scale
	}

	// Linear interpolation between reference temperatures
	temps := make([]float64, 0, len(tc.ScaleFactors))
	for t := range tc.ScaleFactors {
		temps = append(temps, t)
	}
	sort.Float64s(temps)

	if len(temps) < 2 {
		return 1.0
	}

	// Find bracketing temperatures
	var lower, upper float64
	for i := 0; i < len(temps)-1; i++ {
		if temps[i] <= temperature && temps[i+1] >= temperature {
			lower = temps[i]
			upper = temps[i+1]
			break
		}
	}

	if lower == upper {
		return tc.ScaleFactors[lower]
	}

	// Linear interpolation
	t := (temperature - lower) / (upper - lower)
	return tc.ScaleFactors[lower]*(1-t) + tc.ScaleFactors[upper]*t
}

// CompensateWeights applies temperature compensation to weights.
func (tc *TempCompensator) CompensateWeights(weights [][]float64, temperature float64) [][]float64 {
	scale := tc.GetCompensation(temperature)

	compensated := make([][]float64, len(weights))
	for i := range weights {
		compensated[i] = make([]float64, len(weights[i]))
		for j := range weights[i] {
			compensated[i][j] = weights[i][j] * scale
		}
	}

	return compensated
}

// ============================================================================
// ReTern: Fault-Tolerant Ternary CIM
// ============================================================================

// ReTernConfig configures ReTern fault tolerance.
type ReTernConfig struct {
	FaultRate      float64 // Stuck-at fault rate
	RedundancyMode string  // "none", "spatial", "temporal"
	ErrorThreshold float64 // Maximum tolerable error
}

// DefaultReTernConfig returns default ReTern configuration.
func DefaultReTernConfig() *ReTernConfig {
	return &ReTernConfig{
		FaultRate:      0.01, // 1% fault rate
		RedundancyMode: "spatial",
		ErrorThreshold: 0.05,
	}
}

// ReTernLayer implements fault-tolerant ternary layer.
type ReTernLayer struct {
	config   *ReTernConfig
	weights  *TernaryWeight
	faultMap [][]bool // True = faulty cell

	// Overhead metrics
	EnergyOverhead  float64 // 2-2.2% additional
	LatencyOverhead float64 // 3.2-6.6% additional
	AreaOverhead    float64 // <1% additional
}

// NewReTernLayer creates a new ReTern layer.
func NewReTernLayer(config *ReTernConfig, rows, cols int) *ReTernLayer {
	faultMap := make([][]bool, rows)
	for i := range faultMap {
		faultMap[i] = make([]bool, cols)
		for j := range faultMap[i] {
			// Simulate random faults
			faultMap[i][j] = float64(i*cols+j)/float64(rows*cols) < config.FaultRate
		}
	}

	return &ReTernLayer{
		config:   config,
		faultMap: faultMap,
		weights:  NewTernaryWeight(rows, cols),
		// ReTern overhead from paper
		EnergyOverhead:  2.1,  // 2-2.2%
		LatencyOverhead: 4.9,  // 3.2-6.6%
		AreaOverhead:    0.8,  // <1%
	}
}

// ComputeWithFaultTolerance performs fault-tolerant ternary computation.
func (rl *ReTernLayer) ComputeWithFaultTolerance(input []int8) []int32 {
	rows := rl.weights.Rows
	cols := rl.weights.Cols
	results := make([]int32, rows)

	for i := 0; i < rows; i++ {
		var sum int32
		var faultyCount int

		for j := 0; j < cols && j < len(input); j++ {
			// Check for fault
			if rl.faultMap[i][j] {
				faultyCount++
				// Apply redundancy
				if rl.config.RedundancyMode == "spatial" {
					// Use neighboring cell value
					sum += rl.getRedundantValue(i, j, input[j])
				}
				continue
			}

			// Normal ternary operation
			w := rl.weights.GetWeight(i, j)
			switch w {
			case TERNARY_POS:
				sum += int32(input[j])
			case TERNARY_NEG:
				sum -= int32(input[j])
			}
		}

		results[i] = sum
	}

	return results
}

// getRedundantValue gets value from redundant cell.
func (rl *ReTernLayer) getRedundantValue(row, col int, input int8) int32 {
	// Try neighboring cell
	neighbors := []struct{ r, c int }{
		{row, col - 1},
		{row, col + 1},
		{row - 1, col},
		{row + 1, col},
	}

	for _, n := range neighbors {
		if n.r >= 0 && n.r < rl.weights.Rows &&
			n.c >= 0 && n.c < rl.weights.Cols &&
			!rl.faultMap[n.r][n.c] {

			w := rl.weights.GetWeight(n.r, n.c)
			switch w {
			case TERNARY_POS:
				return int32(input)
			case TERNARY_NEG:
				return -int32(input)
			}
		}
	}

	return 0 // No redundant value available
}
