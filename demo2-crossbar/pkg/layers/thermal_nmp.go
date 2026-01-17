// Package layers provides neural network layer implementations for CIM simulation.
// thermal_nmp.go implements thermal modeling for crossbar arrays and
// near-memory processing (NMP) / processing-in-memory (PIM) architectures.
//
// Research basis:
// - Thermal crosstalk: Joule heating spreads through shared electrodes
// - Temperature effects: RON/ROFF ratio degrades at elevated temperatures
// - Temperature-aware optimization: 58% accuracy improvement, 2.39× lifetime
// - Samsung Aquabolt-XL: First commercial HBM2-PIM (16-lane FP16 SIMD)
// - HBM-PIM bandwidth: 4.92 TB/s internal vs 1.23 TB/s external (4×)
// - Performance gains: 8.9× GEMV, 3.5× speech recognition, 60% energy reduction
//
// Key thermal parameters for HZO FeFET:
// - Thermal conductivity: ~1-2 W/(m·K) for HZO
// - Electrode conductivity: TiN ~20 W/(m·K), graphene 3000-5000 W/(m·K)
// - Temperature threshold: 330K for degradation onset
// - Retention: Stable to 10^6 s at room temperature
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// THERMAL MODELING FOR CROSSBAR ARRAYS
// =============================================================================

// ThermalConfig holds thermal simulation parameters
type ThermalConfig struct {
	// Physical properties
	AmbientTempK       float64 // Ambient temperature (default 300K)
	CellResistanceOhm  float64 // Typical cell resistance
	ReadCurrentA       float64 // Read current
	WriteCurrentA      float64 // Write/program current

	// Material thermal properties
	HZOThermalConductivity   float64 // W/(m·K) for HZO layer
	ElectrodeConductivity    float64 // W/(m·K) for TiN electrodes
	SubstrateConductivity    float64 // W/(m·K) for Si substrate

	// Geometry
	CellSizeNm    float64 // Cell dimension
	CellPitchNm   float64 // Cell-to-cell pitch
	HZOThicknessNm float64 // Ferroelectric layer thickness

	// Temperature thresholds
	DegradationThresholdK float64 // Temperature above which degradation starts
	ThermalRunawayK       float64 // Critical temperature for runaway

	// Timing
	HeatDissipationTimeNs float64 // Time constant for cooling
}

// DefaultThermalConfig returns typical HZO FeFET thermal parameters
func DefaultThermalConfig() *ThermalConfig {
	return &ThermalConfig{
		AmbientTempK:           300,     // Room temperature
		CellResistanceOhm:      10000,   // 10 kΩ typical
		ReadCurrentA:           1e-6,    // 1 μA read
		WriteCurrentA:          100e-6,  // 100 μA write

		HZOThermalConductivity:  1.5,    // W/(m·K)
		ElectrodeConductivity:   20,     // TiN
		SubstrateConductivity:   150,    // Silicon

		CellSizeNm:              50,     // 50 nm cell
		CellPitchNm:             100,    // 100 nm pitch
		HZOThicknessNm:          10,     // 10 nm HZO

		DegradationThresholdK:   330,    // 57°C
		ThermalRunawayK:         400,    // 127°C

		HeatDissipationTimeNs:   100,    // 100 ns cooling time
	}
}

// ThermalModel simulates thermal effects in crossbar arrays
type ThermalModel struct {
	config *ThermalConfig
	rng    *rand.Rand

	// Array dimensions
	rows int
	cols int

	// Temperature map for each cell (Kelvin)
	temperatureMap [][]float64

	// Heat accumulation per cell (Joules)
	heatMap [][]float64

	// Statistics
	maxTempReached    float64
	thermalRunawayCount int
	degradedCellCount   int
}

// NewThermalModel creates a thermal model for crossbar array
func NewThermalModel(rows, cols int, config *ThermalConfig) *ThermalModel {
	if config == nil {
		config = DefaultThermalConfig()
	}

	tm := &ThermalModel{
		config:         config,
		rng:            rand.New(rand.NewSource(42)),
		rows:           rows,
		cols:           cols,
		temperatureMap: make([][]float64, rows),
		heatMap:        make([][]float64, rows),
	}

	// Initialize temperature maps
	for i := 0; i < rows; i++ {
		tm.temperatureMap[i] = make([]float64, cols)
		tm.heatMap[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			tm.temperatureMap[i][j] = config.AmbientTempK
		}
	}

	return tm
}

// AddJouleHeat adds heat from a cell operation
func (tm *ThermalModel) AddJouleHeat(row, col int, current float64, durationNs float64) {
	// Joule heating: P = I²R, Q = P × t
	resistance := tm.config.CellResistanceOhm
	power := current * current * resistance
	heat := power * durationNs * 1e-9 // Convert to Joules

	tm.heatMap[row][col] += heat

	// Convert heat to temperature rise
	// ΔT = Q / (m × c) ≈ Q / thermal_mass
	// Simplified model: assume linear relationship
	thermalMass := tm.calculateThermalMass()
	tempRise := heat / thermalMass

	tm.temperatureMap[row][col] += tempRise

	// Track maximum temperature
	if tm.temperatureMap[row][col] > tm.maxTempReached {
		tm.maxTempReached = tm.temperatureMap[row][col]
	}

	// Check for thermal runaway
	if tm.temperatureMap[row][col] > tm.config.ThermalRunawayK {
		tm.thermalRunawayCount++
	}
}

// calculateThermalMass estimates thermal mass of a cell
func (tm *ThermalModel) calculateThermalMass() float64 {
	// Simplified: based on HZO volume and specific heat
	// HZO density: ~9500 kg/m³, specific heat: ~300 J/(kg·K)
	cellVolume := tm.config.CellSizeNm * tm.config.CellSizeNm * tm.config.HZOThicknessNm
	cellVolumeM3 := cellVolume * 1e-27 // nm³ to m³
	density := 9500.0                  // kg/m³
	specificHeat := 300.0              // J/(kg·K)

	return cellVolumeM3 * density * specificHeat
}

// SimulateThermalCrosstalk spreads heat to neighboring cells
func (tm *ThermalModel) SimulateThermalCrosstalk() {
	// Heat spreads through shared electrodes
	// Use diffusion model with thermal conductivity

	// Create temporary map for updated temperatures
	newTemp := make([][]float64, tm.rows)
	for i := range newTemp {
		newTemp[i] = make([]float64, tm.cols)
		copy(newTemp[i], tm.temperatureMap[i])
	}

	// Thermal diffusion coefficient
	// α = k / (ρ × c)
	diffusionCoeff := tm.config.ElectrodeConductivity / (9500 * 300)

	// Distance-based heat transfer
	pitch := tm.config.CellPitchNm * 1e-9 // Convert to meters

	for i := 0; i < tm.rows; i++ {
		for j := 0; j < tm.cols; j++ {
			// Heat flow to neighbors (4-connected)
			neighbors := []struct{ r, c int }{
				{i - 1, j}, {i + 1, j}, {i, j - 1}, {i, j + 1},
			}

			for _, n := range neighbors {
				if n.r >= 0 && n.r < tm.rows && n.c >= 0 && n.c < tm.cols {
					tempDiff := tm.temperatureMap[i][j] - tm.temperatureMap[n.r][n.c]
					heatFlow := diffusionCoeff * tempDiff / (pitch * pitch)
					transferFraction := 0.1 // Simplified coupling factor

					newTemp[i][j] -= heatFlow * transferFraction
					newTemp[n.r][n.c] += heatFlow * transferFraction
				}
			}
		}
	}

	// Update temperature map
	for i := range tm.temperatureMap {
		copy(tm.temperatureMap[i], newTemp[i])
	}
}

// DissipateHeat simulates cooling over time
func (tm *ThermalModel) DissipateHeat(timeNs float64) {
	// Exponential decay toward ambient temperature
	// τ = thermal time constant
	tau := tm.config.HeatDissipationTimeNs
	decayFactor := math.Exp(-timeNs / tau)

	for i := 0; i < tm.rows; i++ {
		for j := 0; j < tm.cols; j++ {
			tempAboveAmbient := tm.temperatureMap[i][j] - tm.config.AmbientTempK
			tm.temperatureMap[i][j] = tm.config.AmbientTempK + tempAboveAmbient*decayFactor
			tm.heatMap[i][j] *= decayFactor
		}
	}
}

// GetTemperature returns temperature at a cell
func (tm *ThermalModel) GetTemperature(row, col int) float64 {
	if row < 0 || row >= tm.rows || col < 0 || col >= tm.cols {
		return tm.config.AmbientTempK
	}
	return tm.temperatureMap[row][col]
}

// IsDegraded returns true if cell temperature exceeds degradation threshold
func (tm *ThermalModel) IsDegraded(row, col int) bool {
	return tm.GetTemperature(row, col) > tm.config.DegradationThresholdK
}

// GetResistanceRatioDegradation returns RON/ROFF ratio degradation factor
func (tm *ThermalModel) GetResistanceRatioDegradation(row, col int) float64 {
	temp := tm.GetTemperature(row, col)

	// Model: RON/ROFF ratio decreases linearly above threshold
	// At 300K: ratio = 1.0 (ideal)
	// At 330K: ratio = 0.8
	// At 400K: ratio = 0.3

	if temp <= tm.config.AmbientTempK {
		return 1.0
	}

	threshold := tm.config.DegradationThresholdK
	if temp <= threshold {
		return 1.0
	}

	// Linear degradation model
	runaway := tm.config.ThermalRunawayK
	degradation := (temp - threshold) / (runaway - threshold)
	ratio := 1.0 - 0.7*degradation // 70% max degradation

	if ratio < 0.3 {
		return 0.3 // Minimum ratio before failure
	}
	return ratio
}

// GetThermalStatistics returns thermal analysis statistics
func (tm *ThermalModel) GetThermalStatistics() ThermalStats {
	// Count degraded cells
	degraded := 0
	avgTemp := 0.0
	maxTemp := tm.config.AmbientTempK

	for i := 0; i < tm.rows; i++ {
		for j := 0; j < tm.cols; j++ {
			temp := tm.temperatureMap[i][j]
			avgTemp += temp
			if temp > maxTemp {
				maxTemp = temp
			}
			if temp > tm.config.DegradationThresholdK {
				degraded++
			}
		}
	}

	totalCells := float64(tm.rows * tm.cols)
	avgTemp /= totalCells

	return ThermalStats{
		AverageTemperatureK:  avgTemp,
		MaxTemperatureK:      maxTemp,
		DegradedCells:        degraded,
		DegradedCellPercent:  float64(degraded) / totalCells * 100,
		ThermalRunawayEvents: tm.thermalRunawayCount,
	}
}

// ThermalStats holds thermal analysis statistics
type ThermalStats struct {
	AverageTemperatureK  float64
	MaxTemperatureK      float64
	DegradedCells        int
	DegradedCellPercent  float64
	ThermalRunawayEvents int
}

// =============================================================================
// TEMPERATURE-AWARE WEIGHT ADJUSTMENT (TAWA)
// =============================================================================

// TAWAConfig holds temperature-aware weight adjustment parameters
type TAWAConfig struct {
	// Monitoring
	MonitoringIntervalNs float64
	TemperatureThresholdK float64

	// Adjustment strategies
	EnableWeightCompensation bool
	EnableRemapping          bool
	EnableCoolingPause       bool

	// Compensation parameters
	CompensationGain float64 // Gain adjustment per degree above threshold
}

// DefaultTAWAConfig returns typical TAWA configuration
func DefaultTAWAConfig() *TAWAConfig {
	return &TAWAConfig{
		MonitoringIntervalNs:     1000,  // 1 μs
		TemperatureThresholdK:    330,   // 57°C
		EnableWeightCompensation: true,
		EnableRemapping:          true,
		EnableCoolingPause:       false,
		CompensationGain:         0.01,  // 1% per degree
	}
}

// TAWAController implements temperature-aware weight adjustment
type TAWAController struct {
	config       *TAWAConfig
	thermalModel *ThermalModel

	// Weight compensation map
	compensationMap [][]float64

	// Remapping table (original -> cooler location)
	remapTable map[CellLocation]CellLocation

	// Statistics
	compensationEvents int
	remapEvents        int
}

// CellLocation represents a cell position
type CellLocation struct {
	Row int
	Col int
}

// NewTAWAController creates a new temperature-aware controller
func NewTAWAController(thermalModel *ThermalModel, config *TAWAConfig) *TAWAController {
	if config == nil {
		config = DefaultTAWAConfig()
	}

	rows := thermalModel.rows
	cols := thermalModel.cols

	tc := &TAWAController{
		config:          config,
		thermalModel:    thermalModel,
		compensationMap: make([][]float64, rows),
		remapTable:      make(map[CellLocation]CellLocation),
	}

	// Initialize compensation map to 1.0 (no compensation)
	for i := 0; i < rows; i++ {
		tc.compensationMap[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			tc.compensationMap[i][j] = 1.0
		}
	}

	return tc
}

// Update performs TAWA based on current thermal state
func (tc *TAWAController) Update() {
	for i := 0; i < tc.thermalModel.rows; i++ {
		for j := 0; j < tc.thermalModel.cols; j++ {
			temp := tc.thermalModel.GetTemperature(i, j)

			if temp > tc.config.TemperatureThresholdK {
				if tc.config.EnableWeightCompensation {
					tc.applyWeightCompensation(i, j, temp)
				}

				if tc.config.EnableRemapping {
					tc.considerRemapping(i, j)
				}
			}
		}
	}
}

// applyWeightCompensation adjusts weight based on temperature
func (tc *TAWAController) applyWeightCompensation(row, col int, temp float64) {
	// Compensate for RON/ROFF ratio degradation
	tempAboveThreshold := temp - tc.config.TemperatureThresholdK
	compensation := 1.0 + tc.config.CompensationGain*tempAboveThreshold

	tc.compensationMap[row][col] = compensation
	tc.compensationEvents++
}

// considerRemapping checks if cell should be remapped to cooler location
func (tc *TAWAController) considerRemapping(row, col int) {
	// Find coolest cell that can accept remapping
	hotLocation := CellLocation{Row: row, Col: col}

	// Already remapped?
	if _, exists := tc.remapTable[hotLocation]; exists {
		return
	}

	// Find coolest available cell
	coolest := tc.findCoolestCell()
	if coolest.Row == row && coolest.Col == col {
		return // Already at coolest location
	}

	coolestTemp := tc.thermalModel.GetTemperature(coolest.Row, coolest.Col)
	hotTemp := tc.thermalModel.GetTemperature(row, col)

	// Only remap if significant temperature difference
	if hotTemp-coolestTemp > 10 { // 10K difference threshold
		tc.remapTable[hotLocation] = coolest
		tc.remapEvents++
	}
}

// findCoolestCell finds the cell with lowest temperature
func (tc *TAWAController) findCoolestCell() CellLocation {
	minTemp := math.MaxFloat64
	coolest := CellLocation{Row: 0, Col: 0}

	for i := 0; i < tc.thermalModel.rows; i++ {
		for j := 0; j < tc.thermalModel.cols; j++ {
			temp := tc.thermalModel.GetTemperature(i, j)
			if temp < minTemp {
				minTemp = temp
				coolest = CellLocation{Row: i, Col: j}
			}
		}
	}

	return coolest
}

// GetCompensatedWeight returns weight with thermal compensation
func (tc *TAWAController) GetCompensatedWeight(row, col int, weight float64) float64 {
	// Apply remapping first
	location := CellLocation{Row: row, Col: col}
	if remapped, exists := tc.remapTable[location]; exists {
		row = remapped.Row
		col = remapped.Col
	}

	// Apply compensation
	compensation := tc.compensationMap[row][col]
	return weight * compensation
}

// GetTAWAStatistics returns TAWA operation statistics
func (tc *TAWAController) GetTAWAStatistics() TAWAStats {
	return TAWAStats{
		CompensationEvents: tc.compensationEvents,
		RemapEvents:        tc.remapEvents,
		RemappedCells:      len(tc.remapTable),
	}
}

// TAWAStats holds TAWA statistics
type TAWAStats struct {
	CompensationEvents int
	RemapEvents        int
	RemappedCells      int
}

// =============================================================================
// NEAR-MEMORY PROCESSING (NMP) / PROCESSING-IN-MEMORY (PIM)
// =============================================================================

// PIMArchitecture represents different PIM architectures
type PIMArchitecture int

const (
	PIMSamsungAquabolt PIMArchitecture = iota // HBM2-PIM with FP16 SIMD
	PIMUPMEMDRAMProc                          // DRAM with processing cores
	PIMAMDAxDIMM                              // Acceleration buffer DIMM
	PIMHBMPIMStandard                         // JEDEC HBM3-PIM standard
	PIMCIMAnalog                              // Analog compute-in-memory
)

// PIMConfig holds PIM architecture configuration
type PIMConfig struct {
	Architecture PIMArchitecture

	// Processing units
	NumProcessors       int     // Number of PIM processors
	LanesPerProcessor   int     // SIMD lanes per processor
	DataWidthBits       int     // FP16, FP32, etc.
	ClockFrequencyMHz   float64 // Processing clock

	// Memory
	NumChannels         int     // Memory channels
	InternalBandwidthTBs float64 // Internal memory bandwidth
	ExternalBandwidthTBs float64 // External I/O bandwidth

	// Energy
	PIMEnergyPJPerOp    float64 // Energy per PIM operation
	DataMovementPJPerByte float64 // Energy per byte moved externally
}

// AquaboltConfig returns Samsung Aquabolt-XL HBM2-PIM configuration
func AquaboltConfig() *PIMConfig {
	return &PIMConfig{
		Architecture:          PIMSamsungAquabolt,
		NumProcessors:         128,   // 32 per die × 4 PIM dies
		LanesPerProcessor:     16,    // 16-lane FP16 SIMD
		DataWidthBits:         16,    // FP16
		ClockFrequencyMHz:     300,   // ~300 MHz PIM clock

		NumChannels:           16,    // 16 pseudo-channels
		InternalBandwidthTBs:  4.92,  // 4.92 TB/s internal
		ExternalBandwidthTBs:  1.23,  // 1.23 TB/s external

		PIMEnergyPJPerOp:      0.5,   // ~0.5 pJ per FP16 op
		DataMovementPJPerByte: 10,    // ~10 pJ per byte external
	}
}

// HBM3PIMConfig returns JEDEC HBM3-PIM standard configuration
func HBM3PIMConfig() *PIMConfig {
	return &PIMConfig{
		Architecture:          PIMHBMPIMStandard,
		NumProcessors:         256,   // More processors in HBM3
		LanesPerProcessor:     32,    // 32-lane SIMD (FP64 capable)
		DataWidthBits:         64,    // FP64 support
		ClockFrequencyMHz:     400,   // Higher clock

		NumChannels:           32,    // 32 channels
		InternalBandwidthTBs:  8.0,   // 8 TB/s internal
		ExternalBandwidthTBs:  2.0,   // 2 TB/s external

		PIMEnergyPJPerOp:      1.0,   // ~1 pJ per FP64 op
		DataMovementPJPerByte: 8,     // Improved efficiency
	}
}

// PIMProcessor simulates a PIM processing unit
type PIMProcessor struct {
	config *PIMConfig

	// Per-processor state
	registerFile []float64 // Local registers
	resultBuffer []float64 // Output buffer

	// Statistics
	opsExecuted   int64
	cyclesActive  int64
}

// NewPIMProcessor creates a new PIM processor
func NewPIMProcessor(config *PIMConfig) *PIMProcessor {
	return &PIMProcessor{
		config:       config,
		registerFile: make([]float64, 32), // 32 registers typical
		resultBuffer: make([]float64, config.LanesPerProcessor),
	}
}

// ExecuteSIMD executes SIMD operation across all lanes
func (p *PIMProcessor) ExecuteSIMD(op PIMOperation, a, b []float64) []float64 {
	lanes := p.config.LanesPerProcessor
	result := make([]float64, lanes)

	for i := 0; i < lanes; i++ {
		aVal := 0.0
		bVal := 0.0
		if i < len(a) {
			aVal = a[i]
		}
		if i < len(b) {
			bVal = b[i]
		}

		switch op {
		case PIMOpAdd:
			result[i] = aVal + bVal
		case PIMOpMul:
			result[i] = aVal * bVal
		case PIMOpFMA:
			// Fused multiply-add: a*b + result[i]
			result[i] = aVal*bVal + p.resultBuffer[i]
		case PIMOpMax:
			result[i] = math.Max(aVal, bVal)
		case PIMOpMin:
			result[i] = math.Min(aVal, bVal)
		}
	}

	p.opsExecuted += int64(lanes)
	p.cyclesActive++
	copy(p.resultBuffer, result)

	return result
}

// PIMOperation represents operations supported by PIM
type PIMOperation int

const (
	PIMOpAdd PIMOperation = iota
	PIMOpMul
	PIMOpFMA  // Fused multiply-add
	PIMOpMax
	PIMOpMin
)

// PIMSystem simulates a complete PIM system
type PIMSystem struct {
	config     *PIMConfig
	processors []*PIMProcessor

	// Memory simulation
	memoryBanks [][]float64

	// Statistics
	totalOps        int64
	totalDataMoved  int64 // bytes
	totalEnergy     float64
}

// NewPIMSystem creates a new PIM system
func NewPIMSystem(config *PIMConfig) *PIMSystem {
	if config == nil {
		config = AquaboltConfig()
	}

	sys := &PIMSystem{
		config:      config,
		processors:  make([]*PIMProcessor, config.NumProcessors),
		memoryBanks: make([][]float64, config.NumChannels),
	}

	// Initialize processors
	for i := 0; i < config.NumProcessors; i++ {
		sys.processors[i] = NewPIMProcessor(config)
	}

	// Initialize memory banks (1MB per channel for simulation)
	for i := 0; i < config.NumChannels; i++ {
		sys.memoryBanks[i] = make([]float64, 1024*1024/8) // 1MB as float64
	}

	return sys
}

// ExecuteGEMV performs general matrix-vector multiply using PIM
func (sys *PIMSystem) ExecuteGEMV(matrix [][]float64, vector []float64) []float64 {
	if len(matrix) == 0 {
		return nil
	}

	rows := len(matrix)
	cols := len(matrix[0])
	result := make([]float64, rows)

	// Distribute work across processors
	rowsPerProcessor := (rows + sys.config.NumProcessors - 1) / sys.config.NumProcessors

	for procID := 0; procID < sys.config.NumProcessors; procID++ {
		startRow := procID * rowsPerProcessor
		endRow := startRow + rowsPerProcessor
		if endRow > rows {
			endRow = rows
		}
		if startRow >= rows {
			break
		}

		proc := sys.processors[procID]

		// Process assigned rows
		for row := startRow; row < endRow; row++ {
			// SIMD dot product
			sum := 0.0
			for col := 0; col < cols; col += proc.config.LanesPerProcessor {
				endCol := col + proc.config.LanesPerProcessor
				if endCol > cols {
					endCol = cols
				}

				// Load vector slice
				vecSlice := vector[col:endCol]

				// Load matrix row slice
				matSlice := matrix[row][col:endCol]

				// FMA operation
				products := proc.ExecuteSIMD(PIMOpMul, matSlice, vecSlice)
				for _, p := range products {
					sum += p
				}
			}
			result[row] = sum
		}
	}

	// Update statistics
	sys.totalOps += int64(rows * cols)
	sys.totalEnergy += float64(rows*cols) * sys.config.PIMEnergyPJPerOp

	return result
}

// ExecuteReduction performs parallel reduction operation
func (sys *PIMSystem) ExecuteReduction(data []float64, op PIMOperation) float64 {
	if len(data) == 0 {
		return 0
	}

	// Use tree reduction across processors
	current := make([]float64, len(data))
	copy(current, data)

	for len(current) > 1 {
		next := make([]float64, (len(current)+1)/2)

		for i := 0; i < len(current); i += 2 {
			procID := (i / 2) % sys.config.NumProcessors
			proc := sys.processors[procID]

			a := current[i]
			b := 0.0
			if i+1 < len(current) {
				b = current[i+1]
			}

			result := proc.ExecuteSIMD(op, []float64{a}, []float64{b})
			next[i/2] = result[0]
		}

		current = next
	}

	sys.totalOps += int64(len(data) - 1)
	return current[0]
}

// GetPerformanceMetrics returns PIM system performance
func (sys *PIMSystem) GetPerformanceMetrics() PIMMetrics {
	// Calculate throughput
	peakTOPS := float64(sys.config.NumProcessors) *
		float64(sys.config.LanesPerProcessor) *
		sys.config.ClockFrequencyMHz * 1e6 / 1e12

	// Calculate bandwidth advantage
	bandwidthRatio := sys.config.InternalBandwidthTBs / sys.config.ExternalBandwidthTBs

	// Energy efficiency
	energyPJPerOp := sys.config.PIMEnergyPJPerOp
	topsPerWatt := 1e12 / (energyPJPerOp * 1e-12) / 1e12 // TOPS/W

	return PIMMetrics{
		Architecture:       sys.config.Architecture,
		PeakTOPS:           peakTOPS,
		InternalBandwidth:  sys.config.InternalBandwidthTBs,
		ExternalBandwidth:  sys.config.ExternalBandwidthTBs,
		BandwidthRatio:     bandwidthRatio,
		EnergyPerOp:        energyPJPerOp,
		TOPSPerWatt:        topsPerWatt,
		TotalOpsExecuted:   sys.totalOps,
		TotalEnergyPJ:      sys.totalEnergy,
	}
}

// PIMMetrics holds PIM performance metrics
type PIMMetrics struct {
	Architecture       PIMArchitecture
	PeakTOPS           float64
	InternalBandwidth  float64 // TB/s
	ExternalBandwidth  float64 // TB/s
	BandwidthRatio     float64
	EnergyPerOp        float64 // pJ
	TOPSPerWatt        float64
	TotalOpsExecuted   int64
	TotalEnergyPJ      float64
}

// =============================================================================
// HYBRID CIM + PIM SYSTEM
// =============================================================================

// HybridCIMPIMConfig configures hybrid compute system
type HybridCIMPIMConfig struct {
	// CIM for analog MAC
	CIMArrayRows   int
	CIMArrayCols   int
	CIMWeightBits  int

	// PIM for digital operations
	PIMConfig *PIMConfig

	// Workload partitioning
	CIMRatio float64 // Fraction of MACs on CIM (0-1)
}

// DefaultHybridConfig returns balanced CIM+PIM configuration
func DefaultHybridConfig() *HybridCIMPIMConfig {
	return &HybridCIMPIMConfig{
		CIMArrayRows:  128,
		CIMArrayCols:  128,
		CIMWeightBits: 6,
		PIMConfig:     AquaboltConfig(),
		CIMRatio:      0.8, // 80% on CIM, 20% on PIM
	}
}

// HybridSystem combines CIM and PIM for optimal efficiency
type HybridSystem struct {
	config       *HybridCIMPIMConfig
	pimSystem    *PIMSystem
	thermalModel *ThermalModel

	// Statistics
	cimOps int64
	pimOps int64
}

// NewHybridSystem creates a hybrid CIM+PIM system
func NewHybridSystem(config *HybridCIMPIMConfig) *HybridSystem {
	if config == nil {
		config = DefaultHybridConfig()
	}

	return &HybridSystem{
		config:       config,
		pimSystem:    NewPIMSystem(config.PIMConfig),
		thermalModel: NewThermalModel(config.CIMArrayRows, config.CIMArrayCols, nil),
	}
}

// ExecuteLayer executes a neural network layer on hybrid system
func (hs *HybridSystem) ExecuteLayer(inputs []float64, weights [][]float64) []float64 {
	rows := len(weights)
	if rows == 0 {
		return nil
	}
	cols := len(weights[0])

	// Partition workload based on thermal state and config
	cimCols := int(float64(cols) * hs.config.CIMRatio)
	pimCols := cols - cimCols

	result := make([]float64, rows)

	// CIM portion (analog MAC)
	if cimCols > 0 {
		cimInputs := inputs[:cimCols]
		for row := 0; row < rows; row++ {
			cimWeights := weights[row][:cimCols]
			// Analog MAC simulation
			sum := 0.0
			for col, w := range cimWeights {
				// Add Joule heating
				hs.thermalModel.AddJouleHeat(row, col, 1e-6, 10)

				// Apply thermal degradation
				degradation := hs.thermalModel.GetResistanceRatioDegradation(row, col)
				sum += cimInputs[col] * w * degradation
			}
			result[row] += sum
			hs.cimOps += int64(cimCols)
		}

		// Thermal crosstalk
		hs.thermalModel.SimulateThermalCrosstalk()
		hs.thermalModel.DissipateHeat(100) // 100ns between operations
	}

	// PIM portion (digital SIMD)
	if pimCols > 0 {
		pimInputs := inputs[cimCols:]
		pimWeights := make([][]float64, rows)
		for row := range pimWeights {
			pimWeights[row] = weights[row][cimCols:]
		}

		pimResult := hs.pimSystem.ExecuteGEMV(pimWeights, pimInputs)
		for row := range result {
			result[row] += pimResult[row]
		}
		hs.pimOps += int64(rows * pimCols)
	}

	return result
}

// GetHybridMetrics returns combined system metrics
func (hs *HybridSystem) GetHybridMetrics() HybridMetrics {
	thermalStats := hs.thermalModel.GetThermalStatistics()
	pimMetrics := hs.pimSystem.GetPerformanceMetrics()

	return HybridMetrics{
		CIMOps:             hs.cimOps,
		PIMOps:             hs.pimOps,
		TotalOps:           hs.cimOps + hs.pimOps,
		CIMRatio:           float64(hs.cimOps) / float64(hs.cimOps+hs.pimOps+1),
		ThermalStats:       thermalStats,
		PIMMetrics:         pimMetrics,
	}
}

// HybridMetrics holds hybrid system performance metrics
type HybridMetrics struct {
	CIMOps       int64
	PIMOps       int64
	TotalOps     int64
	CIMRatio     float64
	ThermalStats ThermalStats
	PIMMetrics   PIMMetrics
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// FormatThermalReport generates thermal analysis report
func FormatThermalReport(tm *ThermalModel) string {
	stats := tm.GetThermalStatistics()

	report := "=== Crossbar Thermal Analysis Report ===\n\n"
	report += fmt.Sprintf("Array Size: %d × %d\n", tm.rows, tm.cols)
	report += fmt.Sprintf("Average Temperature: %.1f K (%.1f °C)\n",
		stats.AverageTemperatureK, stats.AverageTemperatureK-273.15)
	report += fmt.Sprintf("Maximum Temperature: %.1f K (%.1f °C)\n",
		stats.MaxTemperatureK, stats.MaxTemperatureK-273.15)
	report += fmt.Sprintf("Degraded Cells: %d (%.2f%%)\n",
		stats.DegradedCells, stats.DegradedCellPercent)
	report += fmt.Sprintf("Thermal Runaway Events: %d\n", stats.ThermalRunawayEvents)

	if stats.MaxTemperatureK > tm.config.DegradationThresholdK {
		report += "\n⚠️  WARNING: Temperature exceeds degradation threshold!\n"
	}

	return report
}

// FormatPIMReport generates PIM performance report
func FormatPIMReport(sys *PIMSystem) string {
	metrics := sys.GetPerformanceMetrics()

	archNames := map[PIMArchitecture]string{
		PIMSamsungAquabolt: "Samsung Aquabolt-XL HBM2-PIM",
		PIMHBMPIMStandard:  "JEDEC HBM3-PIM Standard",
		PIMUPMEMDRAMProc:   "UPMEM DRAM Processing",
		PIMAMDAxDIMM:       "AMD AxDIMM",
		PIMCIMAnalog:       "Analog CIM",
	}

	report := "=== PIM Performance Report ===\n\n"
	report += fmt.Sprintf("Architecture: %s\n", archNames[metrics.Architecture])
	report += fmt.Sprintf("Peak Throughput: %.2f TOPS\n", metrics.PeakTOPS)
	report += fmt.Sprintf("Internal Bandwidth: %.2f TB/s\n", metrics.InternalBandwidth)
	report += fmt.Sprintf("External Bandwidth: %.2f TB/s\n", metrics.ExternalBandwidth)
	report += fmt.Sprintf("Bandwidth Advantage: %.1f×\n", metrics.BandwidthRatio)
	report += fmt.Sprintf("Energy Efficiency: %.2f TOPS/W\n", metrics.TOPSPerWatt)
	report += fmt.Sprintf("Operations Executed: %d\n", metrics.TotalOpsExecuted)
	report += fmt.Sprintf("Total Energy: %.2f pJ\n", metrics.TotalEnergyPJ)

	return report
}

// CalculateThermalBudget estimates thermal budget for given workload
func CalculateThermalBudget(rows, cols int, opsPerSecond float64, config *ThermalConfig) float64 {
	if config == nil {
		config = DefaultThermalConfig()
	}

	// Heat generated per operation
	heatPerOp := config.ReadCurrentA * config.ReadCurrentA * config.CellResistanceOhm * 10e-9 // 10ns per op

	// Total heat generation rate (Watts)
	totalHeatRate := float64(rows*cols) * opsPerSecond * heatPerOp

	// Maximum temperature rise (simplified model)
	// ΔT = P × R_thermal
	thermalResistance := 100.0 // K/W typical for CIM array
	maxTempRise := totalHeatRate * thermalResistance

	return config.AmbientTempK + maxTempRise
}
