// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"fecim-lattice-tools/shared/physics"
	"fmt"
	"math"
	"math/rand"
)

// DefaultQuantizationLevels is the standard number of discrete analog states.
// Alias to shared/physics for backward compatibility.
const DefaultQuantizationLevels = physics.DefaultLevels

// Conductance range constants (physical units) - aliases to shared/physics
const (
	GMin = physics.GMin
	GMax = physics.GMax
)

// ConductanceModel is an alias to physics.ConductanceModel for backward compatibility.
type ConductanceModel = physics.ConductanceModel

// Conductance model constants - aliases to shared/physics
const (
	ConductanceLinear      = physics.ConductanceLinear
	ConductanceExponential = physics.ConductanceExponential
	ConductanceLookup      = physics.ConductanceLookup
)

// Config contains crossbar array configuration.
type Config struct {
	Rows       int     // Number of rows (word lines)
	Cols       int     // Number of columns (bit lines)
	NoiseLevel float64 // Device-to-device variation (0-1)
	ADCBits    int     // ADC resolution in bits
	DACBits    int     // DAC resolution in bits

	// Conductance model configuration
	ConductanceModel ConductanceModel // Model type (linear, exponential, lookup)
	ConductanceTable []float64        // Calibration table for lookup model (length = 30)

	// Endurance tracking configuration
	Endurance *EnduranceConfig

	// Process variation configuration
	ProcessVariation *ProcessVariationConfig

	// Half-select disturb configuration
	HalfSelect *HalfSelectConfig
}

// Cell represents a single ferroelectric memory cell.
type Cell struct {
	Conductance     float64 // Programmed conductance (normalized 0-1)
	NoiseFactor     float64 // Per-cell noise factor
	SwitchingCount  int64   // Number of write cycles
	HalfSelectCount int64   // Number of half-select (V/2) exposures
	DisturbShift    float64 // Accumulated drift from half-select disturb
}

// EnduranceConfig configures endurance/fatigue modeling.
type EnduranceConfig struct {
	Enabled          bool  // Enable endurance modeling
	FatigueThreshold int64 // Cycles before degradation starts (e.g., 10^8)
	FailureThreshold int64 // Cycles at 50% window degradation (e.g., 10^12)
}

// DefaultEnduranceConfig returns default endurance settings.
// Based on literature: FeFET 10^9-10^12 cycle endurance (IEEE IRPS 2022, Nano Letters 2024).
func DefaultEnduranceConfig() *EnduranceConfig {
	return &EnduranceConfig{
		Enabled:          false, // Off by default for performance
		FatigueThreshold: 100_000_000,       // 10^8 cycles
		FailureThreshold: 1_000_000_000_000, // 10^12 cycles
	}
}

// ProcessVariationConfig configures systematic process variation.
type ProcessVariationConfig struct {
	DeviceSigma float64 // Device-to-device variation (sigma, normalized)
	GradientX   float64 // Horizontal gradient (%/cell from center)
	GradientY   float64 // Vertical gradient (%/cell from center)
	EdgeEffect  float64 // Edge cell degradation factor (0-1)
}

// DefaultProcessVariationConfig returns default process variation settings.
func DefaultProcessVariationConfig() *ProcessVariationConfig {
	return &ProcessVariationConfig{
		DeviceSigma: 0.02,  // 2% device-to-device variation
		GradientX:   0.001, // 0.1% per cell horizontal gradient
		GradientY:   0.001, // 0.1% per cell vertical gradient
		EdgeEffect:  0.05,  // 5% edge degradation
	}
}

// HalfSelectConfig configures half-select disturb modeling.
type HalfSelectConfig struct {
	Enabled          bool    // Enable half-select disturb tracking
	DisturbThreshold float64 // V/Vc threshold for disturb (typically 0.3)
	DisturbRate      float64 // Conductance shift per half-select pulse
}

// DefaultHalfSelectConfig returns default half-select settings.
func DefaultHalfSelectConfig() *HalfSelectConfig {
	return &HalfSelectConfig{
		Enabled:          false, // Off by default for performance
		DisturbThreshold: 0.3,   // 30% of Vc threshold
		DisturbRate:      0.001, // 0.1% conductance shift per pulse
	}
}

// Array represents a crossbar array of ferroelectric cells.
type Array struct {
	config *Config
	cells  [][]Cell

	// ADC/DAC quantization
	adcLevels int
	dacLevels int

	// Statistics
	totalReads  int64
	totalWrites int64
}

// NewArray creates a new crossbar array.
func NewArray(cfg *Config) (*Array, error) {
	if cfg.Rows <= 0 || cfg.Cols <= 0 {
		return nil, fmt.Errorf("invalid array dimensions: %dx%d", cfg.Rows, cfg.Cols)
	}

	arr := &Array{
		config:    cfg,
		adcLevels: 1 << cfg.ADCBits,
		dacLevels: 1 << cfg.DACBits,
	}

	// Initialize cells
	arr.cells = make([][]Cell, cfg.Rows)
	for i := range arr.cells {
		arr.cells[i] = make([]Cell, cfg.Cols)
		for j := range arr.cells[i] {
			arr.cells[i][j] = Cell{
				Conductance: 0.0,
				NoiseFactor: 1.0 + cfg.NoiseLevel*(rand.Float64()*2-1),
			}
		}
	}

	return arr, nil
}

// ProgramWeight programs a weight value to a specific cell.
// Weights are automatically quantized to discrete levels.
func (a *Array) ProgramWeight(row, col int, weight float64) error {
	if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
		return fmt.Errorf("cell index out of range: (%d, %d)", row, col)
	}

	// Quantize to discrete levels
	quantized := QuantizeToLevels(weight)

	a.cells[row][col].Conductance = quantized
	a.cells[row][col].SwitchingCount++
	a.totalWrites++

	return nil
}

// QuantizeToLevels quantizes a value to exactly discrete levels (0-29).
// This matches the standard 30 discrete analog states.
func QuantizeToLevels(value float64) float64 {
	// Clamp to [0, 1]
	value = math.Max(0, math.Min(1, value))
	// Quantize to levels (0 to N-1)
	level := math.Round(value * float64(DefaultQuantizationLevels-1))
	return level / float64(DefaultQuantizationLevels-1)
}

// GetLevel returns the discrete level (0 to N-1) for a conductance value.
func GetLevel(conductance float64) int {
	return int(math.Round(conductance * float64(DefaultQuantizationLevels-1)))
}

// GetPhysicalConductance converts normalized conductance [0,1] to physical units (Siemens).
// The model type determines the interpolation:
//   - Linear: G = Gmin + gNorm*(Gmax-Gmin)  [simple, less accurate at extremes]
//   - Exponential: G = Gmin * exp(ln(Gmax/Gmin) * gNorm)  [realistic FeFET behavior]
//   - Lookup: G = ConductanceTable[level]  [calibration data]
//
// Level 0 → Gmin (10 µS), Level 29 → Gmax (100 µS)
// For exponential model, midpoint (level 15) = geometric mean ≈ 31.6 µS
func (a *Array) GetPhysicalConductance(gNorm float64) float64 {
	// Clamp to valid range
	if gNorm < 0 {
		gNorm = 0
	}
	if gNorm > 1 {
		gNorm = 1
	}

	switch a.config.ConductanceModel {
	case ConductanceExponential:
		// G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
		// This gives exponential scaling where:
		//   gNorm=0   → Gmin
		//   gNorm=0.5 → sqrt(Gmin*Gmax) = geometric mean
		//   gNorm=1   → Gmax
		ratio := GMax / GMin
		return GMin * math.Exp(math.Log(ratio)*gNorm)

	case ConductanceLookup:
		// Use calibration table if available
		if len(a.config.ConductanceTable) == DefaultQuantizationLevels {
			level := int(math.Round(gNorm * float64(DefaultQuantizationLevels-1)))
			if level < 0 {
				level = 0
			}
			if level >= len(a.config.ConductanceTable) {
				level = len(a.config.ConductanceTable) - 1
			}
			return a.config.ConductanceTable[level]
		}
		// Fall back to linear if no table
		fallthrough

	case ConductanceLinear:
		fallthrough
	default:
		// Linear interpolation (original model)
		return GMin + gNorm*(GMax-GMin)
	}
}

// GetPhysicalConductanceForCell returns the physical conductance for a specific cell,
// including effects from endurance fatigue and half-select disturb if enabled.
func (a *Array) GetPhysicalConductanceForCell(row, col int) float64 {
	if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
		return GMin
	}

	cell := &a.cells[row][col]
	gNorm := cell.Conductance

	// Apply half-select disturb shift
	if a.config.HalfSelect != nil && a.config.HalfSelect.Enabled {
		gNorm += cell.DisturbShift
		if gNorm > 1 {
			gNorm = 1
		}
		if gNorm < 0 {
			gNorm = 0
		}
	}

	// Get base physical conductance
	gPhys := a.GetPhysicalConductance(gNorm)

	// Apply endurance fatigue
	if a.config.Endurance != nil && a.config.Endurance.Enabled {
		gPhys = a.applyEnduranceFatigue(cell, gPhys)
	}

	return gPhys
}

// applyEnduranceFatigue applies cycle-dependent degradation to conductance.
// Degradation narrows the conductance window after fatigue threshold.
func (a *Array) applyEnduranceFatigue(cell *Cell, gPhys float64) float64 {
	endCfg := a.config.Endurance
	cycles := cell.SwitchingCount

	if cycles < endCfg.FatigueThreshold {
		return gPhys // No degradation yet
	}

	// Exponential degradation model
	// As cycles approach failure threshold, window narrows to 50%
	fatigueRatio := float64(cycles-endCfg.FatigueThreshold) /
		float64(endCfg.FailureThreshold-endCfg.FatigueThreshold)
	if fatigueRatio > 1 {
		fatigueRatio = 1
	}

	// Degradation factor: 1.0 → 0.5 (50% window narrowing at failure)
	degradation := 1.0 - 0.5*(1-math.Exp(-3*fatigueRatio))

	// Apply degradation: conductance moves toward midpoint
	gMid := (GMax + GMin) / 2
	return gMid + (gPhys-gMid)*degradation
}

// ProgramWeightMatrix programs an entire weight matrix to the array.
func (a *Array) ProgramWeightMatrix(weights [][]float64) error {
	if len(weights) > a.config.Rows {
		return fmt.Errorf("weight matrix rows (%d) exceed array rows (%d)", len(weights), a.config.Rows)
	}

	for i, row := range weights {
		if len(row) > a.config.Cols {
			return fmt.Errorf("weight matrix cols (%d) exceed array cols (%d)", len(row), a.config.Cols)
		}
		for j, w := range row {
			if err := a.ProgramWeight(i, j, w); err != nil {
				return err
			}
		}
	}

	return nil
}

// MVM performs matrix-vector multiplication: y = W * x
// Input x is applied to columns (bit lines), output y is read from rows (word lines).
// Physics: I_row = Σ(G_ij × V_j) - each cell contributes current via Ohm's law.
func (a *Array) MVM(input []float64) ([]float64, error) {
	if len(input) > a.config.Cols {
		return nil, fmt.Errorf("input size (%d) exceeds array columns (%d)", len(input), a.config.Cols)
	}

	output := make([]float64, a.config.Rows)

	// Find max possible current for normalization
	// This occurs when all weights = 1.0 and all inputs = 1.0
	maxCurrent := float64(len(input)) // Theoretical maximum

	for i := 0; i < a.config.Rows; i++ {
		var sum float64
		for j := 0; j < len(input); j++ {
			// Quantize input through DAC
			vIn := a.quantizeDAC(input[j])

			// Read conductance with device variation noise
			g := a.cells[i][j].Conductance * a.cells[i][j].NoiseFactor

			// Ohm's Law: I = G × V
			// Accumulate current (physical summation via Kirchhoff's current law)
			sum += g * vIn
		}

		// Normalize by max possible current to keep in [0,1] range
		// This allows stacking multiple MVMs in neural networks
		normalizedSum := sum / maxCurrent

		// Quantize output through ADC
		output[i] = a.quantizeADC(normalizedSum)
		a.totalReads++
	}

	return output, nil
}

// VMM performs vector-matrix multiplication: y = x * W
// Input x is applied to rows (word lines), output y is read from columns (bit lines).
func (a *Array) VMM(input []float64) ([]float64, error) {
	if len(input) > a.config.Rows {
		return nil, fmt.Errorf("input size (%d) exceeds array rows (%d)", len(input), a.config.Rows)
	}

	output := make([]float64, a.config.Cols)

	for j := 0; j < a.config.Cols; j++ {
		var sum float64
		for i := 0; i < len(input); i++ {
			// Quantize input through DAC
			quantizedInput := a.quantizeDAC(input[i])

			// Read conductance with noise
			g := a.cells[i][j].Conductance * a.cells[i][j].NoiseFactor

			// Accumulate current
			sum += g * quantizedInput
		}

		// Quantize output through ADC
		output[j] = a.quantizeADC(sum / float64(len(input)))
		a.totalReads++
	}

	return output, nil
}

// quantizeDAC applies DAC quantization to input voltage.
func (a *Array) quantizeDAC(value float64) float64 {
	// Clamp to [0, 1]
	value = math.Max(0, math.Min(1, value))
	// Quantize based on DAC bits
	levels := float64(a.dacLevels - 1)
	return math.Round(value*levels) / levels
}

// quantizeADC applies ADC quantization to output current.
func (a *Array) quantizeADC(value float64) float64 {
	// Clamp to [0, 1]
	value = math.Max(0, math.Min(1, value))
	// Quantize based on ADC bits
	levels := float64(a.adcLevels - 1)
	return math.Round(value*levels) / levels
}

// GetStats returns array statistics.
func (a *Array) GetStats() (reads, writes int64) {
	return a.totalReads, a.totalWrites
}

// GetConductanceMatrix returns the current conductance values as a matrix.
func (a *Array) GetConductanceMatrix() [][]float64 {
	matrix := make([][]float64, a.config.Rows)
	for i := range matrix {
		matrix[i] = make([]float64, a.config.Cols)
		for j := range matrix[i] {
			matrix[i][j] = a.cells[i][j].Conductance
		}
	}
	return matrix
}

// Rows returns the number of rows in the array.
func (a *Array) Rows() int {
	return a.config.Rows
}

// Cols returns the number of columns in the array.
func (a *Array) Cols() int {
	return a.config.Cols
}

// GetConfig returns a copy of the array configuration.
func (a *Array) GetConfig() Config {
	return *a.config
}

// SetConductanceModel changes the conductance model type.
func (a *Array) SetConductanceModel(model ConductanceModel) {
	a.config.ConductanceModel = model
}

// SetConductanceTable sets the lookup table for ConductanceLookup model.
// The table should have exactly 30 entries (one per level).
func (a *Array) SetConductanceTable(table []float64) error {
	if len(table) != DefaultQuantizationLevels {
		return fmt.Errorf("conductance table must have exactly %d entries, got %d",
			DefaultQuantizationLevels, len(table))
	}
	a.config.ConductanceTable = make([]float64, len(table))
	copy(a.config.ConductanceTable, table)
	return nil
}

// ProgramWeightWithDisturb programs a weight with half-select disturb tracking.
// This models the real behavior in passive (0T1R) crossbars where cells sharing
// the selected row or column experience a V/2 voltage stress.
// For 1T1R architectures, isPassive should be false (no disturb).
func (a *Array) ProgramWeightWithDisturb(row, col int, weight float64, isPassive bool) error {
	// First, program the target cell normally
	if err := a.ProgramWeight(row, col, weight); err != nil {
		return err
	}

	// No disturb for 1T1R architecture
	if !isPassive {
		return nil
	}

	// Check if half-select tracking is enabled
	hsConfig := a.config.HalfSelect
	if hsConfig == nil || !hsConfig.Enabled {
		return nil
	}

	// Half-select disturb on same row (different columns)
	// These cells see V/2 on their word line during the write pulse
	for j := 0; j < a.config.Cols; j++ {
		if j == col {
			continue
		}
		a.cells[row][j].HalfSelectCount++
		// Apply small disturb if V/2 exceeds disturb threshold
		// halfSelectRatio represents the voltage ratio V_half/V_coercive
		halfSelectRatio := 0.5 // V/2 scheme
		if halfSelectRatio > hsConfig.DisturbThreshold {
			// Disturb magnitude scales with how far above threshold
			disturb := hsConfig.DisturbRate * (halfSelectRatio - hsConfig.DisturbThreshold) / (1.0 - hsConfig.DisturbThreshold)
			a.cells[row][j].DisturbShift += disturb
		}
	}

	// Half-select disturb on same column (different rows)
	// These cells see V/2 on their bit line during the write pulse
	for i := 0; i < a.config.Rows; i++ {
		if i == row {
			continue
		}
		a.cells[i][col].HalfSelectCount++
		halfSelectRatio := 0.5 // V/2 scheme
		if halfSelectRatio > hsConfig.DisturbThreshold {
			disturb := hsConfig.DisturbRate * (halfSelectRatio - hsConfig.DisturbThreshold) / (1.0 - hsConfig.DisturbThreshold)
			a.cells[i][col].DisturbShift += disturb
		}
	}

	return nil
}

// GetProcessVariationFactor returns the systematic variation factor for a cell.
// This combines random device variation with spatial gradients and edge effects.
func (a *Array) GetProcessVariationFactor(row, col int) float64 {
	pvConfig := a.config.ProcessVariation
	if pvConfig == nil {
		// No process variation configured - return cell's random noise
		return a.cells[row][col].NoiseFactor
	}

	// Start with the random component (device-to-device variation)
	random := a.cells[row][col].NoiseFactor

	// Calculate systematic gradient from center
	centerRow := float64(a.config.Rows-1) / 2.0
	centerCol := float64(a.config.Cols-1) / 2.0

	// Horizontal gradient (varies with column distance from center)
	gradX := 1.0 + pvConfig.GradientX*(float64(col)-centerCol)
	// Vertical gradient (varies with row distance from center)
	gradY := 1.0 + pvConfig.GradientY*(float64(row)-centerRow)

	// Edge effect (boundary cells have degraded performance)
	edgeFactor := 1.0
	if pvConfig.EdgeEffect > 0 {
		isEdge := row == 0 || row == a.config.Rows-1 ||
			col == 0 || col == a.config.Cols-1
		if isEdge {
			edgeFactor = 1.0 - pvConfig.EdgeEffect
		}
	}

	return random * gradX * gradY * edgeFactor
}

// GetCellStats returns statistics for a specific cell.
type CellStats struct {
	Row             int
	Col             int
	Conductance     float64 // Normalized [0,1]
	Level           int     // Discrete level [0-29]
	PhysicalG       float64 // Physical conductance (S)
	NoiseFactor     float64
	SwitchingCount  int64
	HalfSelectCount int64
	DisturbShift    float64
	VariationFactor float64
}

// GetCellStats returns detailed statistics for a cell.
func (a *Array) GetCellStats(row, col int) (*CellStats, error) {
	if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
		return nil, fmt.Errorf("cell index out of range: (%d, %d)", row, col)
	}

	cell := &a.cells[row][col]
	return &CellStats{
		Row:             row,
		Col:             col,
		Conductance:     cell.Conductance,
		Level:           GetLevel(cell.Conductance),
		PhysicalG:       a.GetPhysicalConductanceForCell(row, col),
		NoiseFactor:     cell.NoiseFactor,
		SwitchingCount:  cell.SwitchingCount,
		HalfSelectCount: cell.HalfSelectCount,
		DisturbShift:    cell.DisturbShift,
		VariationFactor: a.GetProcessVariationFactor(row, col),
	}, nil
}

// ResetDisturbTracking clears all half-select disturb data.
func (a *Array) ResetDisturbTracking() {
	for i := 0; i < a.config.Rows; i++ {
		for j := 0; j < a.config.Cols; j++ {
			a.cells[i][j].HalfSelectCount = 0
			a.cells[i][j].DisturbShift = 0
		}
	}
}

// ResetCycleCounts clears all switching cycle counts.
func (a *Array) ResetCycleCounts() {
	for i := 0; i < a.config.Rows; i++ {
		for j := 0; j < a.config.Cols; j++ {
			a.cells[i][j].SwitchingCount = 0
		}
	}
}

// AgeCycles artificially ages all cells by adding cycles.
// Useful for demonstrating endurance effects.
func (a *Array) AgeCycles(cycles int64) {
	for i := 0; i < a.config.Rows; i++ {
		for j := 0; j < a.config.Cols; j++ {
			a.cells[i][j].SwitchingCount += cycles
		}
	}
}
