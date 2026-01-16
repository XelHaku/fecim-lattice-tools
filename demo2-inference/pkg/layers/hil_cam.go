// Package layers provides hardware-in-the-loop (HIL) testing frameworks and
// ferroelectric content-addressable memory (CAM) implementations for IronLattice CIM.
//
// Based on research findings:
// - NeuroSim: Validated CIM simulator framework (Frontiers AI 2021)
// - Hardware-aware training for CIM accelerators (Nature Communications 2023)
// - 28nm FeFET-based CAM for similarity search (IEEE JEDS 2025)
// - Combination-encoding CAM (CECAM) for high density (ACS AEM 2025)
// - TAP-CAM: Tunable approximate matching engine (2025)
// - High-temperature FeFET CAM (Adv. Intelligent Systems 2024)
//
// Reference: IronLattice Research Log Sections 280-281
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// HARDWARE-IN-THE-LOOP (HIL) TESTING FRAMEWORK
// ============================================================================

// HILConfig configures the hardware-in-the-loop testing framework
type HILConfig struct {
	// Simulation parameters
	SimulationMode      string  // "pure_sim", "chip_in_loop", "hybrid"
	RealTimeConstraint  float64 // Maximum latency for real-time (ms)
	SubMillisecondExec  bool    // Enable sub-millisecond execution

	// Device modeling
	DeviceTechnology    string  // "rram", "fefet", "sram", "pcm"
	ProcessNode         int     // Technology node (nm)
	TemperatureC        float64 // Operating temperature (°C)

	// Variation modeling
	DeviceVariation     float64 // Device-to-device variation (%)
	CycleVariation      float64 // Cycle-to-cycle variation (%)
	SpatialCorrelation  float64 // Spatial correlation coefficient

	// Non-idealities
	IRDropEnabled       bool    // Model IR drop
	ReadNoiseEnabled    bool    // Model read noise
	WriteNonlinear      bool    // Model nonlinear write
	RetentionLoss       bool    // Model retention degradation

	// Validation parameters
	ValidationSamples   int     // Number of validation samples
	AccuracyThreshold   float64 // Minimum acceptable accuracy
	EnergyBudget        float64 // Maximum energy budget (pJ/inference)
}

// DefaultHILConfig returns default HIL configuration
func DefaultHILConfig() *HILConfig {
	return &HILConfig{
		SimulationMode:     "pure_sim",
		RealTimeConstraint: 1.0,
		SubMillisecondExec: true,
		DeviceTechnology:   "fefet",
		ProcessNode:        28,
		TemperatureC:       25.0,
		DeviceVariation:    5.0,
		CycleVariation:     2.0,
		SpatialCorrelation: 0.3,
		IRDropEnabled:      true,
		ReadNoiseEnabled:   true,
		WriteNonlinear:     true,
		RetentionLoss:      true,
		ValidationSamples:  1000,
		AccuracyThreshold:  0.90,
		EnergyBudget:       1000.0,
	}
}

// DeviceModel represents a device-level model for CIM validation
type DeviceModel struct {
	Config *HILConfig

	// Device characteristics
	NominalConductance float64   // Nominal conductance (μS)
	ConductanceRange   [2]float64 // [Gmin, Gmax]
	NumLevels          int       // Number of programmable levels

	// Variation parameters
	DeviceSigma        float64   // Device variation sigma
	CycleSigma         float64   // Cycle variation sigma

	// Non-ideality models
	IRDropCoeff        float64   // IR drop coefficient
	ReadNoiseSigma     float64   // Read noise sigma
	WriteNonlinearity  float64   // Write nonlinearity exponent

	// State
	ProgrammedValues   [][]float64 // Current programmed states
}

// NewDeviceModel creates a device model based on technology
func NewDeviceModel(config *HILConfig, rows, cols int) *DeviceModel {
	dm := &DeviceModel{
		Config:           config,
		ProgrammedValues: make([][]float64, rows),
	}

	// Set technology-specific parameters
	switch config.DeviceTechnology {
	case "fefet":
		dm.NominalConductance = 50.0  // μS
		dm.ConductanceRange = [2]float64{1.0, 100.0}
		dm.NumLevels = 16             // 4-bit
		dm.DeviceSigma = 0.05
		dm.CycleSigma = 0.02
		dm.IRDropCoeff = 0.01
		dm.ReadNoiseSigma = 0.02
		dm.WriteNonlinearity = 0.8
	case "rram":
		dm.NominalConductance = 100.0
		dm.ConductanceRange = [2]float64{1.0, 200.0}
		dm.NumLevels = 8
		dm.DeviceSigma = 0.10
		dm.CycleSigma = 0.05
		dm.IRDropCoeff = 0.02
		dm.ReadNoiseSigma = 0.03
		dm.WriteNonlinearity = 0.6
	case "sram":
		dm.NominalConductance = 1.0   // Binary
		dm.ConductanceRange = [2]float64{0.0, 1.0}
		dm.NumLevels = 2
		dm.DeviceSigma = 0.01
		dm.CycleSigma = 0.01
		dm.IRDropCoeff = 0.005
		dm.ReadNoiseSigma = 0.01
		dm.WriteNonlinearity = 1.0
	case "pcm":
		dm.NominalConductance = 50.0
		dm.ConductanceRange = [2]float64{0.1, 100.0}
		dm.NumLevels = 16
		dm.DeviceSigma = 0.15
		dm.CycleSigma = 0.08
		dm.IRDropCoeff = 0.03
		dm.ReadNoiseSigma = 0.05
		dm.WriteNonlinearity = 0.5
	}

	// Initialize programmed values
	for i := 0; i < rows; i++ {
		dm.ProgrammedValues[i] = make([]float64, cols)
	}

	return dm
}

// ProgramWeight programs a weight value with variations
func (dm *DeviceModel) ProgramWeight(row, col int, target float64) float64 {
	// Clamp to valid range
	gMin, gMax := dm.ConductanceRange[0], dm.ConductanceRange[1]
	target = math.Max(gMin, math.Min(gMax, target))

	// Quantize to available levels
	levelSize := (gMax - gMin) / float64(dm.NumLevels-1)
	level := int(math.Round((target - gMin) / levelSize))
	quantized := gMin + float64(level)*levelSize

	// Add device variation
	deviceNoise := rand.NormFloat64() * dm.DeviceSigma * quantized

	// Add write nonlinearity
	if dm.Config.WriteNonlinear {
		// Asymmetric write (potentiation vs depression)
		if quantized > dm.ProgrammedValues[row][col] {
			// Potentiation - sublinear
			delta := quantized - dm.ProgrammedValues[row][col]
			delta *= math.Pow(delta/gMax, dm.WriteNonlinearity)
			quantized = dm.ProgrammedValues[row][col] + delta
		} else {
			// Depression - superlinear
			delta := dm.ProgrammedValues[row][col] - quantized
			delta *= math.Pow(delta/gMax, 2.0-dm.WriteNonlinearity)
			quantized = dm.ProgrammedValues[row][col] - delta
		}
	}

	programmed := quantized + deviceNoise
	programmed = math.Max(gMin, math.Min(gMax, programmed))
	dm.ProgrammedValues[row][col] = programmed

	return programmed
}

// ReadWeight reads a weight with noise and IR drop
func (dm *DeviceModel) ReadWeight(row, col int, rowActivity float64) float64 {
	value := dm.ProgrammedValues[row][col]

	// Add cycle-to-cycle variation
	cycleNoise := rand.NormFloat64() * dm.CycleSigma * value

	// Add read noise
	readNoise := 0.0
	if dm.Config.ReadNoiseEnabled {
		readNoise = rand.NormFloat64() * dm.ReadNoiseSigma * value
	}

	// Apply IR drop (depends on row activity)
	irDrop := 0.0
	if dm.Config.IRDropEnabled {
		irDrop = dm.IRDropCoeff * rowActivity * value
	}

	// Apply temperature effect
	tempFactor := 1.0 + 0.001*(dm.Config.TemperatureC-25.0)

	return (value + cycleNoise + readNoise - irDrop) * tempFactor
}

// NeuroSimValidator implements NeuroSim-style validation
type NeuroSimValidator struct {
	Config      *HILConfig
	DeviceModel *DeviceModel

	// Array configuration
	Rows, Cols  int

	// Validation results
	MeasuredAccuracy    float64
	MeasuredEnergy      float64
	MeasuredLatency     float64
	ValidationPassed    bool

	// Detailed metrics
	LayerAccuracies     []float64
	PerSampleErrors     []float64
	EnergyBreakdown     map[string]float64
}

// NewNeuroSimValidator creates a NeuroSim-style validator
func NewNeuroSimValidator(config *HILConfig, rows, cols int) *NeuroSimValidator {
	return &NeuroSimValidator{
		Config:          config,
		DeviceModel:     NewDeviceModel(config, rows, cols),
		Rows:            rows,
		Cols:            cols,
		EnergyBreakdown: make(map[string]float64),
	}
}

// ProgramArray programs weights with realistic variations
func (nsv *NeuroSimValidator) ProgramArray(idealWeights [][]float64) [][]float64 {
	programmed := make([][]float64, len(idealWeights))

	for i, row := range idealWeights {
		programmed[i] = make([]float64, len(row))
		for j, w := range row {
			// Scale to conductance range
			gMin, gMax := nsv.DeviceModel.ConductanceRange[0], nsv.DeviceModel.ConductanceRange[1]
			targetG := gMin + (w+1.0)/2.0*(gMax-gMin) // Assume w in [-1, 1]
			programmed[i][j] = nsv.DeviceModel.ProgramWeight(i, j, targetG)
		}
	}

	return programmed
}

// ComputeMVM performs matrix-vector multiply with non-idealities
func (nsv *NeuroSimValidator) ComputeMVM(input []float64) []float64 {
	if len(input) != nsv.Rows {
		return nil
	}

	output := make([]float64, nsv.Cols)

	// Calculate row activity for IR drop
	rowActivity := 0.0
	for _, v := range input {
		rowActivity += math.Abs(v)
	}
	rowActivity /= float64(len(input))

	// Perform MAC with non-idealities
	for j := 0; j < nsv.Cols; j++ {
		sum := 0.0
		for i := 0; i < nsv.Rows && i < len(input); i++ {
			weight := nsv.DeviceModel.ReadWeight(i, j, rowActivity)
			// Scale back to [-1, 1] range
			gMin, gMax := nsv.DeviceModel.ConductanceRange[0], nsv.DeviceModel.ConductanceRange[1]
			scaledWeight := 2.0*(weight-gMin)/(gMax-gMin) - 1.0
			sum += input[i] * scaledWeight
		}
		output[j] = sum
	}

	return output
}

// ValidateInference runs validation on test samples
func (nsv *NeuroSimValidator) ValidateInference(
	testInputs [][]float64,
	testLabels []int,
	idealWeights [][][]float64,
) error {
	if len(testInputs) == 0 {
		return fmt.Errorf("no test samples provided")
	}

	// Program all layers
	programmedWeights := make([][][]float64, len(idealWeights))
	for l, layerWeights := range idealWeights {
		programmedWeights[l] = nsv.ProgramArray(layerWeights)
	}

	// Run inference on test samples
	correct := 0
	nsv.PerSampleErrors = make([]float64, len(testInputs))

	startTime := time.Now()

	for i, input := range testInputs {
		activations := input

		// Forward pass through layers
		for l := range programmedWeights {
			// Simplified: just use first layer for demo
			if l == 0 {
				output := nsv.ComputeMVM(activations)
				// Apply activation
				for j := range output {
					output[j] = math.Tanh(output[j])
				}
				activations = output
			}
		}

		// Find predicted class
		maxIdx := 0
		maxVal := activations[0]
		for j, v := range activations {
			if v > maxVal {
				maxVal = v
				maxIdx = j
			}
		}

		if maxIdx == testLabels[i] {
			correct++
		}

		// Record error
		if testLabels[i] < len(activations) {
			nsv.PerSampleErrors[i] = 1.0 - (activations[testLabels[i]]+1.0)/2.0
		}
	}

	nsv.MeasuredLatency = float64(time.Since(startTime).Microseconds()) / float64(len(testInputs))
	nsv.MeasuredAccuracy = float64(correct) / float64(len(testInputs))

	// Estimate energy
	nsv.estimateEnergy(len(testInputs), idealWeights)

	// Check validation criteria
	nsv.ValidationPassed = nsv.MeasuredAccuracy >= nsv.Config.AccuracyThreshold &&
		nsv.MeasuredEnergy <= nsv.Config.EnergyBudget

	return nil
}

func (nsv *NeuroSimValidator) estimateEnergy(numSamples int, weights [][][]float64) {
	// Energy model based on NeuroSim methodology
	totalMACs := 0
	for _, layer := range weights {
		if len(layer) > 0 {
			totalMACs += len(layer) * len(layer[0])
		}
	}

	// Energy per MAC based on technology
	var energyPerMAC float64
	switch nsv.Config.DeviceTechnology {
	case "fefet":
		energyPerMAC = 0.5 // pJ
	case "rram":
		energyPerMAC = 1.0
	case "sram":
		energyPerMAC = 0.1
	case "pcm":
		energyPerMAC = 2.0
	}

	// Add ADC/DAC overhead
	adcEnergy := float64(nsv.Cols) * 0.5 // pJ per ADC conversion
	dacEnergy := float64(nsv.Rows) * 0.2 // pJ per DAC conversion

	nsv.EnergyBreakdown["compute"] = float64(totalMACs) * energyPerMAC
	nsv.EnergyBreakdown["adc"] = adcEnergy * float64(numSamples)
	nsv.EnergyBreakdown["dac"] = dacEnergy * float64(numSamples)
	nsv.EnergyBreakdown["peripheral"] = (adcEnergy + dacEnergy) * 0.2

	nsv.MeasuredEnergy = 0
	for _, e := range nsv.EnergyBreakdown {
		nsv.MeasuredEnergy += e
	}
	nsv.MeasuredEnergy /= float64(numSamples) // Per inference
}

// HardwareAwareTrainer implements hardware-aware training
type HardwareAwareTrainer struct {
	Config      *HILConfig
	Validator   *NeuroSimValidator

	// Training state
	Epoch           int
	LearningRate    float64
	NoiseInjection  bool
	QuantizationAware bool

	// Statistics
	TrainingAccuracy  []float64
	ValidationAccuracy []float64
}

// NewHardwareAwareTrainer creates a hardware-aware trainer
func NewHardwareAwareTrainer(config *HILConfig, rows, cols int) *HardwareAwareTrainer {
	return &HardwareAwareTrainer{
		Config:            config,
		Validator:         NewNeuroSimValidator(config, rows, cols),
		LearningRate:      0.01,
		NoiseInjection:    true,
		QuantizationAware: true,
	}
}

// TrainStep performs one training step with hardware awareness
func (hat *HardwareAwareTrainer) TrainStep(
	weights [][]float64,
	inputs [][]float64,
	labels []int,
) [][]float64 {
	// Inject noise during forward pass to simulate hardware
	if hat.NoiseInjection {
		for i := range weights {
			for j := range weights[i] {
				noise := rand.NormFloat64() * hat.Config.DeviceVariation / 100.0
				weights[i][j] += noise * weights[i][j]
			}
		}
	}

	// Quantization-aware forward pass
	if hat.QuantizationAware {
		levels := hat.Validator.DeviceModel.NumLevels
		for i := range weights {
			for j := range weights[i] {
				// Quantize to available levels
				quantized := math.Round(weights[i][j]*float64(levels)) / float64(levels)
				weights[i][j] = quantized
			}
		}
	}

	// Gradient computation would go here (simplified)
	// In practice, use straight-through estimator for quantization

	hat.Epoch++
	return weights
}

// ChipInLoopInterface provides interface to real hardware
type ChipInLoopInterface struct {
	Config       *HILConfig
	Connected    bool
	ChipID       string

	// Communication
	CommandQueue chan ChipCommand
	ResponseChan chan ChipResponse

	// State
	mutex        sync.Mutex
}

// ChipCommand represents a command to send to hardware
type ChipCommand struct {
	Type    string    // "program", "read", "compute"
	Address [2]int    // [row, col]
	Data    []float64
}

// ChipResponse represents a response from hardware
type ChipResponse struct {
	Success bool
	Data    []float64
	Latency float64 // μs
	Energy  float64 // pJ
}

// NewChipInLoopInterface creates a chip-in-loop interface
func NewChipInLoopInterface(config *HILConfig) *ChipInLoopInterface {
	return &ChipInLoopInterface{
		Config:       config,
		Connected:    false,
		CommandQueue: make(chan ChipCommand, 100),
		ResponseChan: make(chan ChipResponse, 100),
	}
}

// Connect simulates connection to hardware
func (cil *ChipInLoopInterface) Connect(chipID string) error {
	cil.mutex.Lock()
	defer cil.mutex.Unlock()

	// In real implementation, this would establish hardware connection
	cil.ChipID = chipID
	cil.Connected = true
	return nil
}

// ProgramChip sends programming commands to hardware
func (cil *ChipInLoopInterface) ProgramChip(weights [][]float64) error {
	if !cil.Connected {
		return fmt.Errorf("not connected to hardware")
	}

	for i, row := range weights {
		for j, w := range row {
			cmd := ChipCommand{
				Type:    "program",
				Address: [2]int{i, j},
				Data:    []float64{w},
			}
			cil.CommandQueue <- cmd
		}
	}

	return nil
}

// ============================================================================
// FERROELECTRIC CONTENT-ADDRESSABLE MEMORY (CAM)
// ============================================================================

// FeFETCAMConfig configures the ferroelectric CAM
type FeFETCAMConfig struct {
	// Array dimensions
	NumRows       int     // Number of CAM entries (words)
	NumCols       int     // Bits per entry (word width)

	// Cell type
	CellType      string  // "bcam", "tcam", "acam", "cecam"
	FeFETsPerCell int     // Number of FeFETs per CAM cell

	// Precision
	AnalogLevels  int     // For ACAM: number of analog levels
	ThresholdBits int     // For approximate matching

	// Operating conditions
	TemperatureC  float64 // Operating temperature
	VoltageV      float64 // Operating voltage

	// Endurance
	MaxEndurance  int     // Maximum write cycles
	RetentionSec  float64 // Retention time (seconds)
}

// DefaultFeFETCAMConfig returns default CAM configuration
func DefaultFeFETCAMConfig() *FeFETCAMConfig {
	return &FeFETCAMConfig{
		NumRows:       256,
		NumCols:       64,
		CellType:      "tcam",
		FeFETsPerCell: 2,
		AnalogLevels:  16,
		ThresholdBits: 4,
		TemperatureC:  25.0,
		VoltageV:      1.8,
		MaxEndurance:  1000000,
		RetentionSec:  100000.0, // >10⁵ seconds from paper
	}
}

// FeFETCAMCell represents a single CAM cell
type FeFETCAMCell struct {
	// Cell state
	StoredValue   float64   // Stored value (0-1 for binary, analog for ACAM)
	Threshold     float64   // Match threshold (for TCAM/ACAM)
	DontCareMask  bool      // Don't care bit (for TCAM)

	// FeFET states
	FeFET1Vth     float64   // Threshold voltage of FeFET 1
	FeFET2Vth     float64   // Threshold voltage of FeFET 2

	// Statistics
	WriteCycles   int
	LastWriteTime float64
}

// NewFeFETCAMCell creates a new CAM cell
func NewFeFETCAMCell() *FeFETCAMCell {
	return &FeFETCAMCell{
		StoredValue:  0.0,
		Threshold:    0.5,
		DontCareMask: false,
		FeFET1Vth:    0.5,
		FeFET2Vth:    0.5,
	}
}

// Program programs the CAM cell with a value
func (cell *FeFETCAMCell) Program(value float64, dontCare bool) {
	cell.StoredValue = value
	cell.DontCareMask = dontCare

	// Program FeFET threshold voltages
	// 2FeFET cell: one stores high Vth for '1', one for '0'
	if value > 0.5 {
		cell.FeFET1Vth = 0.8 // High Vth (programmed)
		cell.FeFET2Vth = 0.2 // Low Vth
	} else {
		cell.FeFET1Vth = 0.2
		cell.FeFET2Vth = 0.8
	}

	cell.WriteCycles++
	cell.LastWriteTime = float64(time.Now().UnixNano()) / 1e9
}

// Match checks if search data matches stored data
func (cell *FeFETCAMCell) Match(searchValue float64, cellType string) bool {
	if cell.DontCareMask {
		return true // Don't care always matches
	}

	switch cellType {
	case "bcam":
		// Binary CAM: exact match
		storedBit := cell.StoredValue > 0.5
		searchBit := searchValue > 0.5
		return storedBit == searchBit

	case "tcam":
		// Ternary CAM: exact match (don't care handled above)
		storedBit := cell.StoredValue > 0.5
		searchBit := searchValue > 0.5
		return storedBit == searchBit

	case "acam":
		// Analog CAM: threshold-based match
		diff := math.Abs(cell.StoredValue - searchValue)
		return diff <= cell.Threshold

	default:
		return cell.StoredValue == searchValue
	}
}

// FeFETCAM implements a ferroelectric content-addressable memory
type FeFETCAM struct {
	Config *FeFETCAMConfig

	// CAM array
	Cells  [][]*FeFETCAMCell

	// Match lines
	MatchLines []bool

	// Statistics
	SearchCount     int
	TotalMatches    int
	AverageLatency  float64
	EnergyPerSearch float64
}

// NewFeFETCAM creates a new FeFET CAM
func NewFeFETCAM(config *FeFETCAMConfig) *FeFETCAM {
	cells := make([][]*FeFETCAMCell, config.NumRows)
	for i := 0; i < config.NumRows; i++ {
		cells[i] = make([]*FeFETCAMCell, config.NumCols)
		for j := 0; j < config.NumCols; j++ {
			cells[i][j] = NewFeFETCAMCell()
		}
	}

	return &FeFETCAM{
		Config:     config,
		Cells:      cells,
		MatchLines: make([]bool, config.NumRows),
	}
}

// ProgramEntry programs a CAM entry (row)
func (cam *FeFETCAM) ProgramEntry(row int, data []float64, mask []bool) error {
	if row >= cam.Config.NumRows {
		return fmt.Errorf("row %d out of range", row)
	}

	for j := 0; j < cam.Config.NumCols && j < len(data); j++ {
		dontCare := false
		if mask != nil && j < len(mask) {
			dontCare = mask[j]
		}
		cam.Cells[row][j].Program(data[j], dontCare)
	}

	return nil
}

// Search performs parallel associative search
func (cam *FeFETCAM) Search(query []float64) []int {
	matchingRows := make([]int, 0)

	// Parallel search across all rows
	for i := 0; i < cam.Config.NumRows; i++ {
		rowMatch := true

		// Check all columns in the row
		for j := 0; j < cam.Config.NumCols && j < len(query); j++ {
			if !cam.Cells[i][j].Match(query[j], cam.Config.CellType) {
				rowMatch = false
				break
			}
		}

		cam.MatchLines[i] = rowMatch
		if rowMatch {
			matchingRows = append(matchingRows, i)
		}
	}

	cam.SearchCount++
	cam.TotalMatches += len(matchingRows)

	// Estimate energy (from 28nm paper: ~fJ per search)
	cam.EnergyPerSearch = float64(cam.Config.NumRows*cam.Config.NumCols) * 0.1 // fJ

	return matchingRows
}

// ComputeHammingDistance computes Hamming distance in-memory
func (cam *FeFETCAM) ComputeHammingDistance(query []float64) []int {
	distances := make([]int, cam.Config.NumRows)

	for i := 0; i < cam.Config.NumRows; i++ {
		dist := 0
		for j := 0; j < cam.Config.NumCols && j < len(query); j++ {
			if !cam.Cells[i][j].Match(query[j], "bcam") {
				dist++
			}
		}
		distances[i] = dist
	}

	return distances
}

// FindKNearest finds k nearest neighbors by Hamming distance
func (cam *FeFETCAM) FindKNearest(query []float64, k int) []int {
	distances := cam.ComputeHammingDistance(query)

	// Create index-distance pairs
	type idxDist struct {
		idx  int
		dist int
	}
	pairs := make([]idxDist, len(distances))
	for i, d := range distances {
		pairs[i] = idxDist{i, d}
	}

	// Sort by distance
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].dist < pairs[j].dist
	})

	// Return k nearest indices
	result := make([]int, 0, k)
	for i := 0; i < k && i < len(pairs); i++ {
		result = append(result, pairs[i].idx)
	}

	return result
}

// CombinationEncodingCAM implements CECAM for higher density
type CombinationEncodingCAM struct {
	Config *FeFETCAMConfig
	CAM    *FeFETCAM

	// Combination encoding parameters
	CombinationSize int     // Bits per combination
	NumCombinations int     // Total combinations

	// Encoding table
	EncodingTable   map[int][]int // Value -> bit positions
}

// NewCombinationEncodingCAM creates a CECAM
func NewCombinationEncodingCAM(config *FeFETCAMConfig, combSize int) *CombinationEncodingCAM {
	cecam := &CombinationEncodingCAM{
		Config:          config,
		CAM:             NewFeFETCAM(config),
		CombinationSize: combSize,
		EncodingTable:   make(map[int][]int),
	}

	// Build encoding table (combination encoding)
	// n choose k combinations for higher density
	cecam.buildEncodingTable()

	return cecam
}

func (cecam *CombinationEncodingCAM) buildEncodingTable() {
	// Generate combinations
	n := cecam.Config.NumCols
	k := cecam.CombinationSize

	// Calculate number of combinations: C(n,k)
	cecam.NumCombinations = binomial(n, k)

	// Generate all k-combinations of n bits
	combo := make([]int, k)
	for i := 0; i < k; i++ {
		combo[i] = i
	}

	idx := 0
	for {
		// Store current combination
		cecam.EncodingTable[idx] = make([]int, k)
		copy(cecam.EncodingTable[idx], combo)
		idx++

		// Generate next combination
		i := k - 1
		for i >= 0 && combo[i] == n-k+i {
			i--
		}
		if i < 0 {
			break
		}
		combo[i]++
		for j := i + 1; j < k; j++ {
			combo[j] = combo[j-1] + 1
		}
	}
}

func binomial(n, k int) int {
	if k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}
	result := 1
	for i := 0; i < k; i++ {
		result = result * (n - i) / (i + 1)
	}
	return result
}

// EncodeValue encodes a value using combination encoding
func (cecam *CombinationEncodingCAM) EncodeValue(value int) []float64 {
	encoded := make([]float64, cecam.Config.NumCols)

	if positions, exists := cecam.EncodingTable[value%cecam.NumCombinations]; exists {
		for _, pos := range positions {
			if pos < len(encoded) {
				encoded[pos] = 1.0
			}
		}
	}

	return encoded
}

// AnalogCAM implements analog content-addressable memory
type AnalogCAM struct {
	Config *FeFETCAMConfig
	CAM    *FeFETCAM

	// Analog-specific parameters
	ThresholdVoltages [][]float64 // Per-cell thresholds for analog matching
	Boundaries        [][]float64 // Decision boundaries (for random forest)
}

// NewAnalogCAM creates an analog CAM
func NewAnalogCAM(config *FeFETCAMConfig) *AnalogCAM {
	config.CellType = "acam"

	acam := &AnalogCAM{
		Config:            config,
		CAM:               NewFeFETCAM(config),
		ThresholdVoltages: make([][]float64, config.NumRows),
		Boundaries:        make([][]float64, config.NumRows),
	}

	for i := 0; i < config.NumRows; i++ {
		acam.ThresholdVoltages[i] = make([]float64, config.NumCols)
		acam.Boundaries[i] = make([]float64, config.NumCols)
	}

	return acam
}

// ProgramBoundary programs an analog decision boundary
func (acam *AnalogCAM) ProgramBoundary(row, col int, boundary, threshold float64) {
	if row < acam.Config.NumRows && col < acam.Config.NumCols {
		acam.CAM.Cells[row][col].StoredValue = boundary
		acam.CAM.Cells[row][col].Threshold = threshold
		acam.ThresholdVoltages[row][col] = threshold
		acam.Boundaries[row][col] = boundary
	}
}

// SearchAnalog performs analog similarity search
func (acam *AnalogCAM) SearchAnalog(query []float64) ([]int, []float64) {
	matchingRows := make([]int, 0)
	similarities := make([]float64, acam.Config.NumRows)

	for i := 0; i < acam.Config.NumRows; i++ {
		// Compute similarity as inverse of distance
		totalDist := 0.0
		matches := 0

		for j := 0; j < acam.Config.NumCols && j < len(query); j++ {
			diff := math.Abs(acam.CAM.Cells[i][j].StoredValue - query[j])
			totalDist += diff

			if diff <= acam.CAM.Cells[i][j].Threshold {
				matches++
			}
		}

		// Row matches if all cells match
		if matches == acam.Config.NumCols || matches == len(query) {
			matchingRows = append(matchingRows, i)
		}

		// Similarity score (higher is better)
		similarities[i] = 1.0 / (1.0 + totalDist)
	}

	return matchingRows, similarities
}

// ============================================================================
// INTEGRATED IRONLATTICE HIL-CAM SYSTEM
// ============================================================================

// IronLatticeHILCAM integrates HIL testing with CAM functionality
type IronLatticeHILCAM struct {
	// Configuration
	HILConfig *HILConfig
	CAMConfig *FeFETCAMConfig

	// Components
	Validator   *NeuroSimValidator
	Trainer     *HardwareAwareTrainer
	CAM         *FeFETCAM
	ACAM        *AnalogCAM
	CECAM       *CombinationEncodingCAM

	// Statistics
	TotalValidations int
	TotalSearches    int
	ValidationHistory []ValidationResult
}

// ValidationResult stores validation run results
type ValidationResult struct {
	Timestamp      time.Time
	Accuracy       float64
	Energy         float64
	Latency        float64
	Passed         bool
	Configuration  string
}

// NewIronLatticeHILCAM creates the integrated system
func NewIronLatticeHILCAM(hilConfig *HILConfig, camConfig *FeFETCAMConfig) *IronLatticeHILCAM {
	return &IronLatticeHILCAM{
		HILConfig:         hilConfig,
		CAMConfig:         camConfig,
		Validator:         NewNeuroSimValidator(hilConfig, 64, 64),
		Trainer:           NewHardwareAwareTrainer(hilConfig, 64, 64),
		CAM:               NewFeFETCAM(camConfig),
		ACAM:              NewAnalogCAM(camConfig),
		CECAM:             NewCombinationEncodingCAM(camConfig, 3),
		ValidationHistory: make([]ValidationResult, 0),
	}
}

// RunValidation runs a complete validation cycle
func (ilhc *IronLatticeHILCAM) RunValidation(
	testInputs [][]float64,
	testLabels []int,
	weights [][][]float64,
) (*ValidationResult, error) {
	err := ilhc.Validator.ValidateInference(testInputs, testLabels, weights)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{
		Timestamp:     time.Now(),
		Accuracy:      ilhc.Validator.MeasuredAccuracy,
		Energy:        ilhc.Validator.MeasuredEnergy,
		Latency:       ilhc.Validator.MeasuredLatency,
		Passed:        ilhc.Validator.ValidationPassed,
		Configuration: fmt.Sprintf("%s_%dnm", ilhc.HILConfig.DeviceTechnology, ilhc.HILConfig.ProcessNode),
	}

	ilhc.ValidationHistory = append(ilhc.ValidationHistory, *result)
	ilhc.TotalValidations++

	return result, nil
}

// RunCAMSearch performs CAM-based associative search
func (ilhc *IronLatticeHILCAM) RunCAMSearch(query []float64, searchType string) ([]int, error) {
	ilhc.TotalSearches++

	switch searchType {
	case "exact":
		return ilhc.CAM.Search(query), nil
	case "hamming":
		return ilhc.CAM.FindKNearest(query, 5), nil
	case "analog":
		matches, _ := ilhc.ACAM.SearchAnalog(query)
		return matches, nil
	case "cecam":
		// Encode query using combination encoding
		encodedQuery := ilhc.CECAM.EncodeValue(int(query[0] * 100))
		return ilhc.CECAM.CAM.Search(encodedQuery), nil
	default:
		return ilhc.CAM.Search(query), nil
	}
}

// FewShotClassify performs few-shot learning classification using CAM
func (ilhc *IronLatticeHILCAM) FewShotClassify(
	supportSet [][]float64,
	supportLabels []int,
	querySet [][]float64,
) []int {
	// Program support set into CAM
	for i, sample := range supportSet {
		if i < ilhc.CAMConfig.NumRows {
			ilhc.CAM.ProgramEntry(i, sample, nil)
		}
	}

	// Classify queries using nearest neighbor
	predictions := make([]int, len(querySet))
	for i, query := range querySet {
		neighbors := ilhc.CAM.FindKNearest(query, 1)
		if len(neighbors) > 0 && neighbors[0] < len(supportLabels) {
			predictions[i] = supportLabels[neighbors[0]]
		}
	}

	return predictions
}

// GenomeReadMapping performs genome read mapping using CAM
func (ilhc *IronLatticeHILCAM) GenomeReadMapping(
	referenceGenome [][]float64,
	reads [][]float64,
	maxMismatches int,
) [][]int {
	// Program reference genome into CAM
	for i, ref := range referenceGenome {
		if i < ilhc.CAMConfig.NumRows {
			ilhc.CAM.ProgramEntry(i, ref, nil)
		}
	}

	// Map each read
	mappings := make([][]int, len(reads))
	for i, read := range reads {
		distances := ilhc.CAM.ComputeHammingDistance(read)

		// Find positions with acceptable mismatches
		mappings[i] = make([]int, 0)
		for j, dist := range distances {
			if dist <= maxMismatches {
				mappings[i] = append(mappings[i], j)
			}
		}
	}

	return mappings
}

// GetStatistics returns system statistics
func (ilhc *IronLatticeHILCAM) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_validations"] = ilhc.TotalValidations
	stats["total_searches"] = ilhc.TotalSearches
	stats["cam_rows"] = ilhc.CAMConfig.NumRows
	stats["cam_cols"] = ilhc.CAMConfig.NumCols
	stats["device_technology"] = ilhc.HILConfig.DeviceTechnology
	stats["process_node_nm"] = ilhc.HILConfig.ProcessNode

	if len(ilhc.ValidationHistory) > 0 {
		last := ilhc.ValidationHistory[len(ilhc.ValidationHistory)-1]
		stats["last_accuracy"] = last.Accuracy
		stats["last_energy_pj"] = last.Energy
		stats["last_passed"] = last.Passed
	}

	stats["cam_search_count"] = ilhc.CAM.SearchCount
	stats["cam_energy_per_search_fj"] = ilhc.CAM.EnergyPerSearch

	return stats
}

// Preset configurations
func IronLatticeHILCAMPreset(scenario string) (*HILConfig, *FeFETCAMConfig) {
	switch scenario {
	case "high_accuracy_validation":
		return &HILConfig{
				SimulationMode:     "pure_sim",
				DeviceTechnology:   "fefet",
				ProcessNode:        28,
				DeviceVariation:    3.0,
				IRDropEnabled:      true,
				ReadNoiseEnabled:   true,
				ValidationSamples:  10000,
				AccuracyThreshold:  0.95,
			},
			&FeFETCAMConfig{
				NumRows:       1024,
				NumCols:       256,
				CellType:      "tcam",
				FeFETsPerCell: 2,
				TemperatureC:  25.0,
			}

	case "chip_in_loop":
		return &HILConfig{
				SimulationMode:     "chip_in_loop",
				RealTimeConstraint: 0.5,
				SubMillisecondExec: true,
				DeviceTechnology:   "fefet",
				ProcessNode:        28,
				DeviceVariation:    5.0,
			},
			DefaultFeFETCAMConfig()

	case "genome_mapping":
		return DefaultHILConfig(),
			&FeFETCAMConfig{
				NumRows:       4096,
				NumCols:       128,
				CellType:      "tcam",
				FeFETsPerCell: 2,
				MaxEndurance:  1000000,
				RetentionSec:  100000.0,
			}

	case "few_shot_learning":
		return &HILConfig{
				SimulationMode:   "pure_sim",
				DeviceTechnology: "fefet",
				ProcessNode:      28,
				DeviceVariation:  5.0,
			},
			&FeFETCAMConfig{
				NumRows:       256,
				NumCols:       512,
				CellType:      "acam",
				AnalogLevels:  16,
				FeFETsPerCell: 2,
			}

	case "high_temperature":
		return &HILConfig{
				SimulationMode:   "pure_sim",
				DeviceTechnology: "fefet",
				ProcessNode:      28,
				TemperatureC:     120.0, // From paper: up to 120°C
				DeviceVariation:  8.0,   // Higher at temperature
			},
			&FeFETCAMConfig{
				NumRows:      256,
				NumCols:      64,
				CellType:     "bcam",
				TemperatureC: 120.0,
			}

	default:
		return DefaultHILConfig(), DefaultFeFETCAMConfig()
	}
}
