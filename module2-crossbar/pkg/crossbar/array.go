// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"fmt"
	"math"
	"math/rand"
)

// DefaultQuantizationLevels is the standard number of discrete analog states.
// "It's got 30 discrete states. So it's not 0-1-0-1."
const DefaultQuantizationLevels = 30

// Conductance range constants (physical units)
const (
	GMin = 10e-6  // 10 µS minimum conductance (OFF state)
	GMax = 100e-6 // 100 µS maximum conductance (ON state)
)

// ConductanceModel specifies the G(V) relationship model.
type ConductanceModel int

const (
	// ConductanceLinear uses linear interpolation: G = Gmin + gNorm*(Gmax-Gmin)
	ConductanceLinear ConductanceModel = iota
	// ConductanceExponential uses exponential scaling: G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
	// This models realistic FeFET behavior where conductance varies exponentially with polarization
	ConductanceExponential
	// ConductanceLookup uses a pre-defined calibration table
	ConductanceLookup
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
