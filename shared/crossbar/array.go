// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/mathutil"
	"fecim-lattice-tools/shared/physics"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
)

// Lazy-initialized logger to ensure it's created after EnableFileLogging() is called
var log *logging.Logger
var logOnce sync.Once

func getLog() *logging.Logger {
	logOnce.Do(func() {
		log = logging.NewLogger("crossbar")
	})
	return log
}

// DefaultQuantizationLevels is the standard number of discrete analog states.
// Alias to shared/physics for backward compatibility.
// Conference claim (COSM 2025), pending peer review: "It's got 30 discrete states."
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
	NoiseLevel float64 // Device-to-device variation sigma (Gaussian, normalized)
	ADCBits    int     // ADC resolution in bits
	DACBits    int     // DAC resolution in bits
	UseGPU     bool    `json:"use_gpu"` // Enable GPU acceleration for MVM operations

	// Conductance model configuration
	ConductanceModel ConductanceModel // Model type (linear, exponential, lookup)
	ConductanceTable []float64        // Calibration table for lookup model (length = 30)

	// Endurance tracking configuration
	Endurance *EnduranceConfig

	// Process variation configuration
	ProcessVariation *ProcessVariationConfig

	// Half-select disturb configuration
	HalfSelect *HalfSelectConfig

	// FeCAP (capacitive) mode configuration.
	// When CellType == CellTypeFeCAP, MVM uses charge-domain computation
	// (Q = C × V) instead of current-domain (I = G × V).
	CellType         CellType         // FeFET (default) or FeCAP
	CMin             float64          // Minimum cell capacitance at cNorm=0 (F)
	CMax             float64          // Maximum cell capacitance at cNorm=1 (F)
	PulseDuration    float64          // Word-line pulse duration for FeCAP reads (s)
	CapacitanceModel CapacitanceModel // How C scales with normalized state
}

// Cell represents a single ferroelectric memory cell.
type Cell struct {
	Conductance     float64 // Programmed conductance (normalized 0-1); FeFET mode
	Capacitance     float64 // Programmed capacitance (normalized 0-1); FeCAP mode
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
		Enabled:          false,             // Off by default for performance
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
//
// These defaults are heuristic placeholders for simulation sensitivity sweeps
// and should be calibrated for process-specific quantitative claims.
func DefaultProcessVariationConfig() *ProcessVariationConfig {
	return &ProcessVariationConfig{
		DeviceSigma: 0.02,  // 2% device-to-device variation (heuristic)
		GradientX:   0.001, // 0.1% per cell horizontal gradient (heuristic)
		GradientY:   0.001, // 0.1% per cell vertical gradient (heuristic)
		EdgeEffect:  0.05,  // 5% edge degradation (heuristic)
	}
}

// HalfSelectConfig configures half-select disturb modeling.
type HalfSelectConfig struct {
	Enabled          bool    // Enable half-select disturb tracking
	DisturbThreshold float64 // V/Vc threshold for disturb (typically 0.3)
	DisturbRate      float64 // Conductance shift per half-select pulse
}

// DefaultHalfSelectConfig returns default half-select settings.
// Disturb parameters are heuristic placeholders until calibrated to measured
// half-select disturb data.
func DefaultHalfSelectConfig() *HalfSelectConfig {
	return &HalfSelectConfig{
		Enabled:          false, // Off by default for performance
		DisturbThreshold: 0.3,   // 30% of Vc threshold (heuristic)
		DisturbRate:      0.001, // 0.1% conductance shift per pulse (heuristic)
	}
}

// Array represents a crossbar array of ferroelectric cells.
type Array struct {
	config *Config
	cells  [][]Cell

	// ADC/DAC quantization
	adcLevels int
	dacLevels int

	// GPU acceleration (nil if GPU not enabled/available)
	gpuAccelerator *GPUAccelerator
	gpuInitialized bool // true after first GPU init attempt

	// Statistics
	totalReads  int64
	totalWrites int64
}

// NewArray creates a new crossbar array.
func NewArray(cfg *Config) (*Array, error) {
	getLog().Input("NewArray", map[string]interface{}{
		"rows":       cfg.Rows,
		"cols":       cfg.Cols,
		"noiseLevel": cfg.NoiseLevel,
		"adcBits":    cfg.ADCBits,
		"dacBits":    cfg.DACBits,
		"useGPU":     cfg.UseGPU,
	})

	if cfg.Rows <= 0 || cfg.Cols <= 0 {
		err := fmt.Errorf("invalid array dimensions: %dx%d", cfg.Rows, cfg.Cols)
		getLog().Error(err, "NewArray validation failed")
		return nil, err
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
				NoiseFactor: initialNoiseFactor(cfg),
			}
		}
	}

	getLog().Output("NewArray", arr)
	return arr, nil
}

// initialNoiseFactor returns a per-cell device variation factor.
// Uses Gaussian noise to match PHYSICS.md: G = G_nominal * (1 + sigma * N(0,1)).
// If ProcessVariation is configured, its DeviceSigma overrides NoiseLevel.
func initialNoiseFactor(cfg *Config) float64 {
	sigma := cfg.NoiseLevel
	if cfg.ProcessVariation != nil {
		sigma = cfg.ProcessVariation.DeviceSigma
	}
	if sigma <= 0 {
		return 1.0
	}
	factor := 1.0 + rand.NormFloat64()*sigma
	if factor < 0 {
		factor = 0
	}
	return factor
}

// initGPU lazily initializes GPU accelerator on first use.
// Silently falls back to CPU if GPU is unavailable.
func (a *Array) initGPU() {
	if a.gpuInitialized || !a.config.UseGPU {
		return
	}
	a.gpuInitialized = true

	accel, err := NewGPUAccelerator(a.config.Rows, a.config.Cols)
	if err != nil {
		// Log warning but continue with CPU fallback
		// GPU unavailable is not a fatal error
		return
	}
	a.gpuAccelerator = accel

	// Upload device variation factors to GPU
	if a.gpuAccelerator != nil {
		a.uploadVariationToGPU()
	}
}

// uploadVariationToGPU uploads device variation factors to GPU.
func (a *Array) uploadVariationToGPU() {
	if a.gpuAccelerator == nil {
		return
	}

	// Build variation matrix from cells
	variation := make([]float32, a.config.Rows*a.config.Cols)
	for i := 0; i < a.config.Rows; i++ {
		for j := 0; j < a.config.Cols; j++ {
			variation[i*a.config.Cols+j] = float32(a.GetProcessVariationFactor(i, j))
		}
	}

	// Upload to GPU (ignore errors - will fall back to CPU if this fails)
	_ = a.gpuAccelerator.SetDeviceVariation(variation)
}

// mvmGPU performs MVM using GPU acceleration.
// The GPU shader implements VMM physics, so we transpose to get MVM:
// MVM(W, x) = VMM(x, W^T)
func (a *Array) mvmGPU(input []float64) ([]float64, error) {
	// For MVM: y = W*x where W is [Rows x Cols], x has Cols elements, y has Rows elements
	// GPU shader does VMM: I_j = Sum_i(V_i * G_ij), input size = shader rows, output size = shader cols
	//
	// To get MVM from VMM shader:
	// - Transpose W: shader sees W^T which is [Cols x Rows]
	// - Swap dimensions: shader rows = our cols, shader cols = our rows
	// - Input (our cols) goes to shader rows
	// - Output (shader cols) = our rows

	// Convert input to float32 (input has Cols elements)
	input32 := make([]float32, len(input))
	for i, v := range input {
		input32[i] = float32(a.quantizeDAC(v))
	}

	// Build TRANSPOSED conductance matrix: G^T[j][i] = G[i][j]
	// Layout: G^T stored as [Cols x Rows] row-major = G stored column-major
	conductancesT := make([]float32, a.config.Rows*a.config.Cols)
	for i := 0; i < a.config.Rows; i++ {
		for j := 0; j < a.config.Cols; j++ {
			// G^T[j][i] at index j*Rows + i
			conductancesT[j*a.config.Rows+i] = float32(a.cells[i][j].Conductance)
		}
	}

	// Set up GPU parameters with SWAPPED dimensions
	// Shader "rows" = our Cols (input size)
	// Shader "cols" = our Rows (output size)
	params := CrossbarParams{
		Rows:           int32(a.config.Cols), // SWAPPED: shader rows = our cols
		Cols:           int32(a.config.Rows), // SWAPPED: shader cols = our rows
		NoiseLevel:     float32(a.config.NoiseLevel),
		ADCBits:        int32(a.config.ADCBits),
		DACBits:        int32(a.config.DACBits),
		Time:           0.0,
		WireResistance: 0.0,
		DriftCoeff:     0.0,
	}

	// Execute GPU MVM (shader computes VMM on transposed matrix = MVM on original)
	outputs32, err := a.gpuAccelerator.MVM(conductancesT, input32, params)
	if err != nil {
		// Fall back to CPU on GPU error
		return a.mvmCPU(input)
	}

	// Find max possible current for normalization (same as CPU)
	maxCurrent := float64(len(input))

	// Convert back to float64 and apply normalization/quantization
	// Output vector now has Rows elements (one per output row)
	output := make([]float64, a.config.Rows)
	for i := 0; i < a.config.Rows; i++ {
		normalizedSum := float64(outputs32[i]) / maxCurrent
		output[i] = a.quantizeADC(normalizedSum)
		atomic.AddInt64(&a.totalReads, 1)
	}

	return output, nil
}

// Destroy releases GPU resources. Call when array is no longer needed.
func (a *Array) Destroy() {
	if a.gpuAccelerator != nil {
		a.gpuAccelerator.Destroy()
		a.gpuAccelerator = nil
	}
}

// ProgramWeight programs a weight value to a specific cell.
// Weights are automatically quantized to discrete levels.
func (a *Array) ProgramWeight(row, col int, weight float64) error {
	if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
		err := fmt.Errorf("cell index out of range: (%d, %d)", row, col)
		getLog().Error(err, "ProgramWeight validation failed")
		return err
	}

	// Quantize to discrete levels
	quantized := QuantizeToLevels(weight)

	getLog().Calculation("ProgramWeight", map[string]interface{}{
		"row":            row,
		"col":            col,
		"originalWeight": weight,
		"quantized":      quantized,
		"level":          GetLevel(quantized),
	}, nil)

	a.cells[row][col].Conductance = quantized
	a.cells[row][col].SwitchingCount++
	atomic.AddInt64(&a.totalWrites, 1)

	return nil
}

// QuantizeToLevels quantizes a value to exactly discrete levels (0-29).
// This matches the demo 30-level baseline (conference claim).
// Wrapper for shared/physics.QuantizeTo30Levels for backward compatibility.
func QuantizeToLevels(value float64) float64 {
	quantized := physics.QuantizeTo30Levels(value)
	getLog().Calculation("QuantizeToLevels", map[string]interface{}{
		"input": value,
	}, quantized)
	return quantized
}

// GetLevel returns the discrete level (0 to N-1) for a conductance value.
// Wrapper for shared/physics.GetLevelFor30 for backward compatibility.
func GetLevel(conductance float64) int {
	return physics.GetLevelFor30(conductance)
}

// GetPhysicalConductance converts normalized conductance [0,1] to physical units (Siemens).
// The model type determines the interpolation:
//   - Linear: G = Gmin + gNorm*(Gmax-Gmin)  [simple, less accurate at extremes]
//   - Exponential: G = Gmin * exp(ln(Gmax/Gmin) * gNorm)  [realistic FeFET behavior]
//   - Lookup: G = ConductanceTable[level]  [calibration data]
//
// Level 0 → Gmin (10 µS), Level 29 → Gmax (100 µS)
// For exponential model, midpoint (level 15) = geometric mean ≈ 31.6 µS
//
// Uses shared/physics.NormalizedToPhysicalRange for the base calculation.
func (a *Array) GetPhysicalConductance(gNorm float64) float64 {
	// Use shared physics implementation with array's conductance table for lookup
	return physics.NormalizedToPhysicalRange(gNorm, a.config.ConductanceModel, GMin, GMax, a.config.ConductanceTable)
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
	// Apply process/device variation (random + gradients + edge)
	gPhys *= a.GetProcessVariationFactor(row, col)

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
	if len(weights) == 0 {
		return fmt.Errorf("weight matrix is empty or nil")
	}

	getLog().Input("ProgramWeightMatrix", map[string]interface{}{
		"matrixRows": len(weights),
		"matrixCols": len(weights[0]),
		"arrayRows":  a.config.Rows,
		"arrayCols":  a.config.Cols,
	})

	if len(weights) > a.config.Rows {
		err := fmt.Errorf("weight matrix rows (%d) exceed array rows (%d)", len(weights), a.config.Rows)
		getLog().Error(err, "ProgramWeightMatrix validation failed")
		return err
	}

	for i, row := range weights {
		if len(row) > a.config.Cols {
			err := fmt.Errorf("weight matrix cols (%d) exceed array cols (%d)", len(row), a.config.Cols)
			getLog().Error(err, "ProgramWeightMatrix validation failed")
			return err
		}
		for j, w := range row {
			if err := a.ProgramWeight(i, j, w); err != nil {
				return err
			}
		}
	}

	getLog().Output("ProgramWeightMatrix", "success")
	return nil
}

// MVM performs matrix-vector multiplication: y = W * x
// Input x is applied to columns (bit lines), output y is read from rows (word lines).
// Physics: I_row = Σ(G_ij × V_j) - each cell contributes current via Ohm's law.
func (a *Array) MVM(input []float64) ([]float64, error) {
	getLog().Input("MVM", map[string]interface{}{
		"inputLen": len(input),
		"rows":     a.config.Rows,
		"cols":     a.config.Cols,
	})

	if len(input) == 0 {
		return nil, fmt.Errorf("MVM input is empty")
	}
	if len(input) > a.config.Cols {
		err := fmt.Errorf("input size (%d) exceeds array columns (%d)", len(input), a.config.Cols)
		getLog().Error(err, "MVM validation failed")
		return nil, err
	}

	// Try GPU path first
	a.initGPU()
	if a.gpuAccelerator != nil && a.gpuAccelerator.IsAvailable() {
		output, err := a.mvmGPU(input)
		if err == nil {
			getLog().Calculation("MVM", map[string]interface{}{
				"mode": "GPU",
			}, output)
		}
		return output, err
	}

	// CPU fallback
	output, err := a.mvmCPU(input)
	if err == nil {
		getLog().Calculation("MVM", map[string]interface{}{
			"mode": "CPU",
		}, output)
	}
	return output, err
}

// mvmCPU performs MVM using CPU implementation (original algorithm).
func (a *Array) mvmCPU(input []float64) ([]float64, error) {
	cols := len(input)
	rows := a.config.Rows
	output := make([]float64, rows)

	// Pre-quantize all DAC inputs once (avoids repeated quantization inside
	// the inner loop which runs rows×cols times vs. cols times).
	dacIn := make([]float64, cols)
	for j := 0; j < cols; j++ {
		dacIn[j] = a.quantizeDAC(input[j])
	}

	// Find max possible current for normalization
	// This occurs when all weights = 1.0 and all inputs = 1.0
	maxCurrent := float64(cols) // Theoretical maximum

	for i := 0; i < rows; i++ {
		var sum float64
		for j := 0; j < cols; j++ {
			// Read conductance with device variation noise
			g := a.cells[i][j].Conductance * a.GetProcessVariationFactor(i, j)

			// Ohm's Law: I = G × V
			// Accumulate current (physical summation via Kirchhoff's current law)
			sum += g * dacIn[j]
		}

		// Normalize by max possible current to keep in [0,1] range
		// This allows stacking multiple MVMs in neural networks
		normalizedSum := sum / maxCurrent

		// Quantize output through ADC
		output[i] = a.quantizeADC(normalizedSum)
	}

	// Batch the read counter update — one atomic op per MVM call instead of
	// one per row, which reduces contention on the cache line for the counter.
	atomic.AddInt64(&a.totalReads, int64(rows))

	return output, nil
}

// VMM performs vector-matrix multiplication: y = x * W
// Input x is applied to rows (word lines), output y is read from columns (bit lines).
func (a *Array) VMM(input []float64) ([]float64, error) {
	getLog().Input("VMM", map[string]interface{}{
		"inputLen": len(input),
		"rows":     a.config.Rows,
		"cols":     a.config.Cols,
	})

	if len(input) > a.config.Rows {
		err := fmt.Errorf("input size (%d) exceeds array rows (%d)", len(input), a.config.Rows)
		getLog().Error(err, "VMM validation failed")
		return nil, err
	}

	rows := len(input)
	cols := a.config.Cols
	output := make([]float64, cols)

	// Pre-quantize all DAC inputs once (avoids repeated quantization inside
	// the inner loop which runs cols×rows times vs. rows times).
	dacIn := make([]float64, rows)
	for i := 0; i < rows; i++ {
		dacIn[i] = a.quantizeDAC(input[i])
	}

	for j := 0; j < cols; j++ {
		var sum float64
		for i := 0; i < rows; i++ {
			// Read conductance with noise
			g := a.cells[i][j].Conductance * a.GetProcessVariationFactor(i, j)

			// Accumulate current
			sum += g * dacIn[i]
		}

		// Quantize output through ADC
		output[j] = a.quantizeADC(sum / float64(rows))
	}

	// Batch the read counter update — one atomic op per VMM call instead of
	// one per column, reducing false-sharing contention on the counter.
	atomic.AddInt64(&a.totalReads, int64(cols))

	getLog().Calculation("VMM", map[string]interface{}{
		"outputLen": len(output),
	}, output)

	return output, nil
}

// quantizeDAC applies DAC quantization to input voltage.
func (a *Array) quantizeDAC(value float64) float64 {
	value = mathutil.Clamp01(value)
	levels := float64(a.dacLevels - 1)
	return math.Round(value*levels) / levels
}

// quantizeADC applies ADC quantization to output current.
func (a *Array) quantizeADC(value float64) float64 {
	value = mathutil.Clamp01(value)
	// Quantize based on ADC bits
	levels := float64(a.adcLevels - 1)
	return math.Round(value*levels) / levels
}

// GetStats returns array statistics.
func (a *Array) GetStats() (reads, writes int64) {
	return atomic.LoadInt64(&a.totalReads), atomic.LoadInt64(&a.totalWrites)
}

// GetConductanceMatrix returns the current conductance values as a matrix.
// This returns the programmed (ideal) conductance values without noise.
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

// GetEffectiveConductanceMatrix returns the actual effective conductance values.
// This includes per-cell noise factors that represent process variation and noise.
// Use this for comparing "actual" vs "ideal" conductances.
func (a *Array) GetEffectiveConductanceMatrix() [][]float64 {
	matrix := make([][]float64, a.config.Rows)
	for i := range matrix {
		matrix[i] = make([]float64, a.config.Cols)
		for j := range matrix[i] {
			// Effective conductance = base conductance × variation factor
			matrix[i][j] = a.cells[i][j].Conductance * a.GetProcessVariationFactor(i, j)
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
