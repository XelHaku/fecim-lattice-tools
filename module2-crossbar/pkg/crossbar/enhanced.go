// Package crossbar provides enhanced crossbar array simulation with integrated non-idealities.
package crossbar

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// MVMOptions configures which non-idealities to include in MVM computation.
type MVMOptions struct {
	EnableIRDrop     bool
	EnableSneakPaths bool
	EnableVariation  bool
	EnableDrift      bool
	Temperature      float64 // Kelvin (default 300K = 27C)
	Architecture     string  // "1T1R" or "0T1R" - affects sneak path and IR drop calculations
}

// Is1T1R returns true if the architecture uses single transistor isolation (1T1R).
// 1T1R provides ~1000:1 sneak path isolation compared to passive 0T1R.
func (o *MVMOptions) Is1T1R() bool {
	if o == nil || o.Architecture == "" {
		return false // Default to 0T1R (passive crossbar)
	}
	// Check for 1T1R but NOT 2T1R
	if strings.Contains(o.Architecture, "2T1R") {
		return false
	}
	return o.Architecture == "1T1R" ||
		strings.Contains(o.Architecture, "1T1R") ||
		strings.Contains(o.Architecture, "Transistor")
}

// Is2T1R returns true if the architecture uses dual transistor isolation (2T1R).
// 2T1R provides individual cell addressing via WL+CSL AND-gate selection,
// offering even better isolation than 1T1R (~10000:1 vs ~1000:1).
func (o *MVMOptions) Is2T1R() bool {
	if o == nil || o.Architecture == "" {
		return false
	}
	return strings.Contains(o.Architecture, "2T1R") ||
		strings.Contains(o.Architecture, "Dual")
}

// HasTransistorIsolation returns true if the architecture has any transistor isolation (1T1R or 2T1R).
// Both provide significantly better isolation than passive 0T1R arrays.
func (o *MVMOptions) HasTransistorIsolation() bool {
	return o.Is1T1R() || o.Is2T1R()
}

// DefaultMVMOptions returns options with all non-idealities enabled.
func DefaultMVMOptions() *MVMOptions {
	return &MVMOptions{
		EnableIRDrop:     true,
		EnableSneakPaths: true,
		EnableVariation:  true,
		EnableDrift:      false, // Drift usually simulated separately over time
		Temperature:      300.0, // Room temperature
		Architecture:     "0T1R", // Default to passive crossbar (simpler, higher density)
	}
}

// MVMResult contains the results of an MVM operation with detailed analysis.
type MVMResult struct {
	// Outputs
	IdealOutput  []float64 // Output without non-idealities
	ActualOutput []float64 // Output with non-idealities

	// Error metrics
	RMSE          float64 // Root mean square error
	MaxError      float64 // Maximum absolute error
	MeanError     float64 // Mean absolute error
	AccuracyLoss  float64 // Estimated accuracy loss percentage

	// Energy metrics (estimated)
	ArrayEnergy float64 // Energy for MVM computation (pJ)
	ADCEnergy   float64 // Energy for ADC conversion (pJ)
	DACEnergy   float64 // Energy for DAC conversion (pJ)
	TotalEnergy float64 // Total energy (pJ)

	// Performance metrics
	MACOperations int     // Number of multiply-accumulate operations
	Latency       float64 // Estimated latency (ns)
	Throughput    float64 // Operations per second

	// Non-ideality analysis
	IRDropAnalysis    *IRDropAnalysis
	SneakPathAnalysis *SneakPathAnalysis

	// Comparison with GPU
	GPUEquivalentEnergy float64 // Estimated GPU energy for same operation (pJ)
	EnergyEfficiency    float64 // FeCIM energy / GPU energy ratio
}

// MVMWithNonIdealities performs MVM with all specified non-idealities.
// This is the most accurate simulation mode.
func (a *Array) MVMWithNonIdealities(input []float64, opts *MVMOptions) (*MVMResult, error) {
	if len(input) > a.config.Cols {
		return nil, fmt.Errorf("input size (%d) exceeds array columns (%d)", len(input), a.config.Cols)
	}

	if opts == nil {
		opts = DefaultMVMOptions()
	}

	result := &MVMResult{
		IdealOutput:   make([]float64, a.config.Rows),
		ActualOutput:  make([]float64, a.config.Rows),
		MACOperations: a.config.Rows * len(input),
	}

	// Step 1: Compute ideal output (no non-idealities)
	for i := 0; i < a.config.Rows; i++ {
		var sum float64
		for j := 0; j < len(input); j++ {
			sum += a.cells[i][j].Conductance * input[j]
		}
		// Normalize to keep in reasonable range
		result.IdealOutput[i] = sum / float64(len(input))
	}

	// Step 2: Get effective voltages considering IR drop
	var effectiveVoltages [][]float64
	if opts.EnableIRDrop {
		params := DefaultWireParams()
		archName := "1T1R"

		// Architecture affects IR drop via sneak current contribution:
		// - 0T1R (passive): Highest IR drop due to full sneak currents
		// - 1T1R: Reduced IR drop due to transistor isolation (~1000:1)
		// - 2T1R: Lowest IR drop due to dual transistor AND-gate (~10000:1)
		if opts.Is2T1R() {
			// 2T1R has best isolation - lowest effective resistance
			params.RwordLine *= 0.85 // 15% lower due to minimal sneak currents
			params.RbitLine *= 0.85
			archName = "2T1R"
		} else if opts.Is1T1R() {
			// 1T1R has good isolation - no modifier (baseline)
			archName = "1T1R"
		} else {
			// 0T1R (passive) has highest effective resistance due to sneak currents
			params.RwordLine *= 1.5 // 50% higher effective resistance for 0T1R
			params.RbitLine *= 1.5
			archName = "0T1R"
		}

		// Apply temperature effect on wire resistance
		if opts.Temperature != 300.0 {
			tempFactor := 1.0 + 0.00393*(opts.Temperature-300.0) // Copper TCR
			params.RwordLine *= tempFactor
			params.RbitLine *= tempFactor
		}
		irAnalysis := a.AnalyzeIRDrop(input, params)
		// Debug: log the architecture and max IR drop
		fmt.Printf("[IR Drop] Architecture=%s, RwordLine=%.2f, MaxDrop=%.4f%%\n",
			archName, params.RwordLine, irAnalysis.MaxIRDrop*100)
		result.IRDropAnalysis = irAnalysis
		effectiveVoltages = irAnalysis.EffectiveVoltage
	} else {
		// No IR drop - use ideal voltages
		effectiveVoltages = make([][]float64, a.config.Rows)
		for i := range effectiveVoltages {
			effectiveVoltages[i] = make([]float64, a.config.Cols)
			for j := range effectiveVoltages[i] {
				effectiveVoltages[i][j] = 1.0 // Ideal voltage
			}
		}
	}

	// Step 3: Compute actual output with non-idealities
	for i := 0; i < a.config.Rows; i++ {
		var sum float64
		for j := 0; j < len(input); j++ {
			// Get conductance
			G := a.cells[i][j].Conductance

			// Apply device variation
			if opts.EnableVariation {
				G *= a.cells[i][j].NoiseFactor
			}

			// Get effective input voltage (DAC quantized, IR drop affected)
			vIn := a.quantizeDAC(input[j])
			if opts.EnableIRDrop && i < len(effectiveVoltages) && j < len(effectiveVoltages[i]) {
				vIn *= effectiveVoltages[i][j]
			}

			// Ohm's Law: I = G × V
			sum += G * vIn
		}

		// Add sneak path currents
		if opts.EnableSneakPaths {
			sneakCurrent := a.computeSneakCurrentForRow(i, input, opts)
			sum += sneakCurrent
		}

		// Normalize and quantize through ADC
		normalizedSum := sum / float64(len(input))
		result.ActualOutput[i] = a.quantizeADC(normalizedSum)
		a.totalReads++
	}

	// Compute sneak path analysis for center cell (architecture-aware)
	if opts.EnableSneakPaths {
		centerRow := a.config.Rows / 2
		centerCol := a.config.Cols / 2
		// Use architecture-specific isolation factor:
		// 0T1R: 1.0 (full sneak), 1T1R: 0.001 (1000x), 2T1R: 0.0001 (10000x)
		isolationFactor := 1.0 // Default: 0T1R passive
		if opts.Is2T1R() {
			isolationFactor = 0.0001 // 10000x isolation from dual transistors
		} else if opts.Is1T1R() {
			isolationFactor = 0.001 // 1000x isolation from single transistor
		}
		result.SneakPathAnalysis = a.AnalyzeSneakPathsWithIsolation(centerRow, centerCol, isolationFactor)
	}

	// Step 4: Compute error metrics
	result.computeErrorMetrics()

	// Step 5: Compute energy metrics
	result.computeEnergyMetrics(a.config.Rows, len(input), a.config.ADCBits)

	return result, nil
}

// SneakPathMode controls the sneak path calculation mode.
type SneakPathMode int

const (
	// SneakPathSimplified uses a fixed factor model (fast, less accurate)
	SneakPathSimplified SneakPathMode = iota
	// SneakPathFull computes actual three-cell series paths (slower, more accurate)
	SneakPathFull
)

// SneakPathThreshold is the array size above which simplified mode is used automatically.
// Full sneak calculation is O(n^4) so becomes expensive for large arrays.
const SneakPathThreshold = 32 // Use simplified mode for arrays larger than 32x32

// computeSneakCurrentForRow computes total sneak current affecting a row.
// The sneak factor varies based on architecture:
// - 2T1R: Dual transistor provides ~10000:1 isolation, virtually no sneak paths
// - 1T1R: Single transistor provides ~1000:1 isolation, minimal sneak paths
// - 0T1R: Passive crossbar has full sneak path impact
func (a *Array) computeSneakCurrentForRow(row int, input []float64, opts *MVMOptions) float64 {
	// For 2T1R, dual transistor AND-gate makes sneak paths virtually zero
	if opts != nil && opts.Is2T1R() {
		return a.computeSimplifiedSneakCurrent(row, input, 0.000001) // 10x better than 1T1R
	}

	// For 1T1R, transistor isolation makes sneak paths negligible
	if opts != nil && opts.Is1T1R() {
		return a.computeSimplifiedSneakCurrent(row, input, 0.00001)
	}

	// For small passive arrays, use full calculation for accuracy
	if a.config.Rows <= SneakPathThreshold && a.config.Cols <= SneakPathThreshold {
		return a.computeFullSneakCurrent(row, input)
	}

	// For large passive arrays, use simplified model for performance
	return a.computeSimplifiedSneakCurrent(row, input, 0.01)
}

// computeSimplifiedSneakCurrent uses a fixed factor model for sneak paths.
// This is the original simplified model - fast but less accurate.
func (a *Array) computeSimplifiedSneakCurrent(row int, input []float64, sneakFactor float64) float64 {
	var sneakCurrent float64

	for j := 0; j < len(input); j++ {
		// Two-cell approximation: current through row cells to other rows
		for i := 0; i < a.config.Rows; i++ {
			if i == row {
				continue
			}
			g1 := a.cells[row][j].Conductance
			g2 := a.cells[i][j].Conductance
			if g1 > 0.01 && g2 > 0.01 {
				sneakCurrent += input[j] * sneakFactor * (g1 * g2) / (g1 + g2)
			}
		}
	}

	return sneakCurrent
}

// computeFullSneakCurrent computes actual three-cell sneak paths through all rows.
// This provides literature-matched sneak path magnitudes (5-20% for passive arrays).
//
// Physical model: For a given target row reading, sneak paths form when:
// Input voltage on column 'col' → cell[srcRow][col] → cell[srcRow][j] → cell[targetRow][j]
// This creates a three-cell series path that adds parasitic current to the target row.
func (a *Array) computeFullSneakCurrent(targetRow int, input []float64) float64 {
	var totalSneak float64

	inputLen := len(input)
	if inputLen > a.config.Cols {
		inputLen = a.config.Cols
	}

	// For each input column
	for col := 0; col < inputLen; col++ {
		vIn := input[col]
		if vIn < 0.01 {
			continue // Skip negligible inputs
		}

		// Sum sneak paths through every other row
		for srcRow := 0; srcRow < a.config.Rows; srcRow++ {
			if srcRow == targetRow {
				continue // Not a sneak path
			}

			// Entry conductance: input column on source row
			g1 := a.cells[srcRow][col].Conductance
			if g1 < 0.01 {
				continue // Path blocked by low conductance
			}

			// Three-cell paths through every other column
			for j := 0; j < a.config.Cols; j++ {
				if j == col {
					continue // Direct path, not sneak
				}

				// Middle conductance: source row, different column
				g2 := a.cells[srcRow][j].Conductance
				if g2 < 0.01 {
					continue // Path blocked
				}

				// Exit conductance: target row, exit column
				g3 := a.cells[targetRow][j].Conductance
				if g3 < 0.01 {
					continue // Path blocked
				}

				// Three conductances in series: G_path = 1/(1/g1 + 1/g2 + 1/g3)
				gPath := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)

				// Current through this sneak path: I = V * G_path
				totalSneak += vIn * gPath
			}
		}
	}

	return totalSneak
}

// ComputeFullMVMSneak computes sneak currents for all rows in a full MVM operation.
// Returns an array of sneak current per row.
// This is the comprehensive sneak path analysis for passive arrays.
func (a *Array) ComputeFullMVMSneak(input []float64, opts *MVMOptions) []float64 {
	sneakPerRow := make([]float64, a.config.Rows)

	// 1T1R and 2T1R have negligible sneak paths due to transistor isolation
	if opts != nil && opts.HasTransistorIsolation() {
		return sneakPerRow // All zeros
	}

	for targetRow := 0; targetRow < a.config.Rows; targetRow++ {
		if a.config.Rows <= SneakPathThreshold && a.config.Cols <= SneakPathThreshold {
			sneakPerRow[targetRow] = a.computeFullSneakCurrent(targetRow, input)
		} else {
			sneakPerRow[targetRow] = a.computeSimplifiedSneakCurrent(targetRow, input, 0.01)
		}
	}

	return sneakPerRow
}

// GetSneakPathContribution returns detailed sneak path analysis for a single row.
// Useful for visualization and debugging.
type SneakPathContribution struct {
	SourceRow    int
	SourceCol    int
	ExitCol      int
	PathG        float64 // Effective path conductance
	PathCurrent  float64 // Current through this path
}

// AnalyzeSneakContributions returns the top sneak path contributors for a row.
func (a *Array) AnalyzeSneakContributions(targetRow int, input []float64, maxPaths int) []SneakPathContribution {
	var contributions []SneakPathContribution

	inputLen := len(input)
	if inputLen > a.config.Cols {
		inputLen = a.config.Cols
	}

	for col := 0; col < inputLen; col++ {
		vIn := input[col]
		if vIn < 0.01 {
			continue
		}

		for srcRow := 0; srcRow < a.config.Rows; srcRow++ {
			if srcRow == targetRow {
				continue
			}

			g1 := a.cells[srcRow][col].Conductance
			if g1 < 0.01 {
				continue
			}

			for j := 0; j < a.config.Cols; j++ {
				if j == col {
					continue
				}

				g2 := a.cells[srcRow][j].Conductance
				g3 := a.cells[targetRow][j].Conductance
				if g2 < 0.01 || g3 < 0.01 {
					continue
				}

				gPath := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
				pathCurrent := vIn * gPath

				contributions = append(contributions, SneakPathContribution{
					SourceRow:   srcRow,
					SourceCol:   col,
					ExitCol:     j,
					PathG:       gPath,
					PathCurrent: pathCurrent,
				})
			}
		}
	}

	// Sort by current magnitude (descending) and limit to maxPaths
	// Using simple selection sort since maxPaths is typically small
	for i := 0; i < len(contributions) && i < maxPaths; i++ {
		maxIdx := i
		for j := i + 1; j < len(contributions); j++ {
			if contributions[j].PathCurrent > contributions[maxIdx].PathCurrent {
				maxIdx = j
			}
		}
		contributions[i], contributions[maxIdx] = contributions[maxIdx], contributions[i]
	}

	if len(contributions) > maxPaths {
		contributions = contributions[:maxPaths]
	}

	return contributions
}

// computeErrorMetrics calculates error statistics between ideal and actual outputs.
func (r *MVMResult) computeErrorMetrics() {
	if len(r.IdealOutput) != len(r.ActualOutput) || len(r.IdealOutput) == 0 {
		return
	}

	var sumSqError, sumAbsError, maxError float64
	for i := range r.IdealOutput {
		diff := r.IdealOutput[i] - r.ActualOutput[i]
		absDiff := math.Abs(diff)

		sumSqError += diff * diff
		sumAbsError += absDiff
		if absDiff > maxError {
			maxError = absDiff
		}
	}

	n := float64(len(r.IdealOutput))
	r.RMSE = math.Sqrt(sumSqError / n)
	r.MeanError = sumAbsError / n
	r.MaxError = maxError

	// Estimate accuracy loss (empirical relationship)
	// Based on literature: ~1% accuracy loss per 3% RMSE for neural networks
	r.AccuracyLoss = r.RMSE * 100 / 3.0
}

// computeEnergyMetrics estimates energy consumption.
func (r *MVMResult) computeEnergyMetrics(rows, cols, adcBits int) {
	// Energy estimates based on literature (all in pJ)
	// FeFET read energy: ~0.01 fJ per cell
	// ADC energy scales as 2^bits

	cellReadEnergy := 0.01e-3 // 0.01 fJ = 0.00001 pJ per cell read
	r.ArrayEnergy = float64(rows*cols) * cellReadEnergy

	// ADC energy: ~0.5 pJ per conversion for 6-bit ADC
	adcEnergyBase := 0.5
	r.ADCEnergy = float64(rows) * adcEnergyBase * math.Pow(2, float64(adcBits-6))

	// DAC energy: ~0.1 pJ per conversion
	r.DACEnergy = float64(cols) * 0.1

	r.TotalEnergy = r.ArrayEnergy + r.ADCEnergy + r.DACEnergy

	// GPU comparison: ~10 pJ per MAC operation (including memory access)
	gpuEnergyPerMAC := 10.0
	r.GPUEquivalentEnergy = float64(r.MACOperations) * gpuEnergyPerMAC
	r.EnergyEfficiency = r.GPUEquivalentEnergy / r.TotalEnergy

	// Latency: ~10 ns for analog MVM
	r.Latency = 10.0

	// Throughput: MACs per second
	r.Throughput = float64(r.MACOperations) / (r.Latency * 1e-9)
}

// DifferentialArray implements signed weight support using two crossbar arrays.
// Positive weights go to G+, negative weights go to G-.
// Output = I+ - I-
type DifferentialArray struct {
	positive *Array
	negative *Array
	config   *Config
}

// NewDifferentialArray creates a differential pair array for signed weights.
func NewDifferentialArray(cfg *Config) (*DifferentialArray, error) {
	pos, err := NewArray(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create positive array: %w", err)
	}

	neg, err := NewArray(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create negative array: %w", err)
	}

	return &DifferentialArray{
		positive: pos,
		negative: neg,
		config:   cfg,
	}, nil
}

// ProgramSignedWeight programs a signed weight [-1, 1] to the differential array.
func (d *DifferentialArray) ProgramSignedWeight(row, col int, weight float64) error {
	// Clamp to [-1, 1]
	weight = math.Max(-1, math.Min(1, weight))

	if weight >= 0 {
		// Positive weight: G+ = weight, G- = 0
		if err := d.positive.ProgramWeight(row, col, weight); err != nil {
			return err
		}
		return d.negative.ProgramWeight(row, col, 0)
	}

	// Negative weight: G+ = 0, G- = |weight|
	if err := d.positive.ProgramWeight(row, col, 0); err != nil {
		return err
	}
	return d.negative.ProgramWeight(row, col, -weight)
}

// ProgramSignedMatrix programs an entire signed weight matrix.
func (d *DifferentialArray) ProgramSignedMatrix(weights [][]float64) error {
	for i, row := range weights {
		for j, w := range row {
			if err := d.ProgramSignedWeight(i, j, w); err != nil {
				return err
			}
		}
	}
	return nil
}

// MVM performs signed MVM using differential readout.
func (d *DifferentialArray) MVM(input []float64) ([]float64, error) {
	posOut, err := d.positive.MVM(input)
	if err != nil {
		return nil, fmt.Errorf("positive array MVM failed: %w", err)
	}

	negOut, err := d.negative.MVM(input)
	if err != nil {
		return nil, fmt.Errorf("negative array MVM failed: %w", err)
	}

	output := make([]float64, len(posOut))
	for i := range output {
		// Differential readout: I+ - I-
		output[i] = posOut[i] - negOut[i]
	}

	return output, nil
}

// MVMWithNonIdealities performs signed MVM with non-idealities.
func (d *DifferentialArray) MVMWithNonIdealities(input []float64, opts *MVMOptions) (*MVMResult, error) {
	posResult, err := d.positive.MVMWithNonIdealities(input, opts)
	if err != nil {
		return nil, fmt.Errorf("positive array MVM failed: %w", err)
	}

	negResult, err := d.negative.MVMWithNonIdealities(input, opts)
	if err != nil {
		return nil, fmt.Errorf("negative array MVM failed: %w", err)
	}

	// Combine results
	result := &MVMResult{
		IdealOutput:   make([]float64, len(posResult.IdealOutput)),
		ActualOutput:  make([]float64, len(posResult.ActualOutput)),
		MACOperations: posResult.MACOperations * 2, // Two arrays
	}

	for i := range result.IdealOutput {
		result.IdealOutput[i] = posResult.IdealOutput[i] - negResult.IdealOutput[i]
		result.ActualOutput[i] = posResult.ActualOutput[i] - negResult.ActualOutput[i]
	}

	result.computeErrorMetrics()

	// Energy is doubled for two arrays
	result.ArrayEnergy = posResult.ArrayEnergy + negResult.ArrayEnergy
	result.ADCEnergy = posResult.ADCEnergy + negResult.ADCEnergy
	result.DACEnergy = posResult.DACEnergy // DAC shared
	result.TotalEnergy = result.ArrayEnergy + result.ADCEnergy + result.DACEnergy
	result.GPUEquivalentEnergy = posResult.GPUEquivalentEnergy * 2
	result.EnergyEfficiency = result.GPUEquivalentEnergy / result.TotalEnergy
	result.Latency = posResult.Latency // Same latency (parallel)
	result.Throughput = float64(result.MACOperations) / (result.Latency * 1e-9)

	return result, nil
}

// GetSignedWeight returns the effective signed weight at a cell.
func (d *DifferentialArray) GetSignedWeight(row, col int) float64 {
	posMatrix := d.positive.GetConductanceMatrix()
	negMatrix := d.negative.GetConductanceMatrix()
	return posMatrix[row][col] - negMatrix[row][col]
}

// GetSignedWeightMatrix returns the full signed weight matrix.
func (d *DifferentialArray) GetSignedWeightMatrix() [][]float64 {
	posMatrix := d.positive.GetConductanceMatrix()
	negMatrix := d.negative.GetConductanceMatrix()

	result := make([][]float64, len(posMatrix))
	for i := range posMatrix {
		result[i] = make([]float64, len(posMatrix[i]))
		for j := range posMatrix[i] {
			result[i][j] = posMatrix[i][j] - negMatrix[i][j]
		}
	}
	return result
}

// WriteStatistics configures write variation/statistics.
type WriteStatistics struct {
	Enabled  bool       // Enable statistical write variation
	VthSigma float64    // Threshold voltage sigma (normalized, 0-1)
	RNG      *rand.Rand // Random number generator (seeded for reproducibility)
}

// DefaultWriteStatistics returns default write statistics configuration.
func DefaultWriteStatistics() *WriteStatistics {
	return &WriteStatistics{
		Enabled:  false,
		VthSigma: 0.05, // 5% variation = ~1.5 level spread
		RNG:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ProgramWeightWithVariation programs a weight with statistical write variation.
// Returns the actual achieved level (may differ from target due to Vth variation).
func (a *Array) ProgramWeightWithVariation(row, col int, targetLevel int, stats *WriteStatistics) (int, error) {
	if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
		return 0, fmt.Errorf("cell index out of range: (%d, %d)", row, col)
	}

	if targetLevel < 0 || targetLevel >= DefaultQuantizationLevels {
		return 0, fmt.Errorf("target level %d out of range [0, %d)", targetLevel, DefaultQuantizationLevels)
	}

	// Without variation, just program the exact level
	if stats == nil || !stats.Enabled {
		weight := float64(targetLevel) / float64(DefaultQuantizationLevels-1)
		if err := a.ProgramWeight(row, col, weight); err != nil {
			return 0, err
		}
		return targetLevel, nil
	}

	// Add Gaussian noise to target level based on Vth variation
	// VthSigma of 0.05 corresponds to about 1.5 levels of variation (0.05 * 29 ≈ 1.5)
	rng := stats.RNG
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	noise := rng.NormFloat64() * stats.VthSigma * float64(DefaultQuantizationLevels-1)
	actualLevel := int(math.Round(float64(targetLevel) + noise))

	// Clamp to valid range
	if actualLevel < 0 {
		actualLevel = 0
	}
	if actualLevel >= DefaultQuantizationLevels {
		actualLevel = DefaultQuantizationLevels - 1
	}

	// Program the actual achieved level
	weight := float64(actualLevel) / float64(DefaultQuantizationLevels-1)
	a.cells[row][col].Conductance = QuantizeToLevels(weight)
	a.cells[row][col].SwitchingCount++
	a.totalWrites++

	return actualLevel, nil
}

// ProgramMatrixWithVariation programs an entire weight matrix with statistical variation.
// Returns a matrix of actual achieved levels.
func (a *Array) ProgramMatrixWithVariation(targetLevels [][]int, stats *WriteStatistics) ([][]int, error) {
	if len(targetLevels) > a.config.Rows {
		return nil, fmt.Errorf("matrix rows (%d) exceed array rows (%d)", len(targetLevels), a.config.Rows)
	}

	actualLevels := make([][]int, len(targetLevels))
	for i, row := range targetLevels {
		if len(row) > a.config.Cols {
			return nil, fmt.Errorf("matrix cols (%d) exceed array cols (%d)", len(row), a.config.Cols)
		}
		actualLevels[i] = make([]int, len(row))
		for j, targetLevel := range row {
			actual, err := a.ProgramWeightWithVariation(i, j, targetLevel, stats)
			if err != nil {
				return nil, err
			}
			actualLevels[i][j] = actual
		}
	}
	return actualLevels, nil
}

// WriteVariationAnalysis contains statistics about write variation.
type WriteVariationAnalysis struct {
	NumWrites        int       // Total number of write operations
	NumErrors        int       // Number of writes with level error
	ErrorRate        float64   // Percentage of writes with errors
	AvgLevelError    float64   // Average absolute level error
	MaxLevelError    int       // Maximum level error observed
	ErrorDistribution []int    // Histogram of error magnitudes [0, ±1, ±2, ±3, ...]
}

// AnalyzeWriteVariation tests write variation by programming random levels.
func (a *Array) AnalyzeWriteVariation(stats *WriteStatistics, numTests int) *WriteVariationAnalysis {
	result := &WriteVariationAnalysis{
		NumWrites:        numTests,
		ErrorDistribution: make([]int, 10), // Track errors up to ±9 levels
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalError := 0.0

	for i := 0; i < numTests; i++ {
		// Pick random cell and target level
		row := rng.Intn(a.config.Rows)
		col := rng.Intn(a.config.Cols)
		targetLevel := rng.Intn(DefaultQuantizationLevels)

		actualLevel, err := a.ProgramWeightWithVariation(row, col, targetLevel, stats)
		if err != nil {
			continue
		}

		levelError := actualLevel - targetLevel
		absError := levelError
		if absError < 0 {
			absError = -absError
		}

		if levelError != 0 {
			result.NumErrors++
		}
		totalError += float64(absError)

		if absError > result.MaxLevelError {
			result.MaxLevelError = absError
		}

		if absError < len(result.ErrorDistribution) {
			result.ErrorDistribution[absError]++
		}
	}

	result.ErrorRate = float64(result.NumErrors) / float64(numTests) * 100
	result.AvgLevelError = totalError / float64(numTests)

	return result
}

// WriteVerifyConfig configures write-verify programming.
type WriteVerifyConfig struct {
	MaxIterations int     // Maximum programming iterations (default: 10)
	Tolerance     float64 // Acceptable tolerance in levels (default: 0.5)
	PulseStep     float64 // Pulse amplitude step (default: 0.1)
}

// DefaultWriteVerifyConfig returns default write-verify configuration.
func DefaultWriteVerifyConfig() *WriteVerifyConfig {
	return &WriteVerifyConfig{
		MaxIterations: 10,
		Tolerance:     0.5,
		PulseStep:     0.1,
	}
}

// WriteVerifyResult contains the result of a write-verify operation.
type WriteVerifyResult struct {
	TargetLevel   int
	AchievedLevel int
	Iterations    int
	Converged     bool
	FinalError    float64
}

// ProgramWeightVerified programs a weight with write-verify loop.
func (a *Array) ProgramWeightVerified(row, col int, targetLevel int, cfg *WriteVerifyConfig) (*WriteVerifyResult, error) {
	if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
		return nil, fmt.Errorf("cell index out of range: (%d, %d)", row, col)
	}

	if cfg == nil {
		cfg = DefaultWriteVerifyConfig()
	}

	if targetLevel < 0 || targetLevel >= DefaultQuantizationLevels {
		return nil, fmt.Errorf("target level %d out of range [0, %d)", targetLevel, DefaultQuantizationLevels)
	}

	result := &WriteVerifyResult{
		TargetLevel: targetLevel,
	}

	targetConductance := float64(targetLevel) / float64(DefaultQuantizationLevels-1)
	tolerance := cfg.Tolerance / float64(DefaultQuantizationLevels-1)

	currentConductance := a.cells[row][col].Conductance

	for iter := 0; iter < cfg.MaxIterations; iter++ {
		result.Iterations = iter + 1

		// Simulate write with variation
		error := (rand.Float64()*2 - 1) * a.config.NoiseLevel * cfg.PulseStep
		newConductance := currentConductance + (targetConductance-currentConductance)*cfg.PulseStep + error

		// Clamp to valid range
		newConductance = math.Max(0, math.Min(1, newConductance))

		// Quantize to 30 levels
		newConductance = QuantizeToLevels(newConductance)
		a.cells[row][col].Conductance = newConductance
		a.cells[row][col].SwitchingCount++
		a.totalWrites++
		currentConductance = newConductance

		// Check convergence
		currentLevel := GetLevel(currentConductance)
		result.AchievedLevel = currentLevel
		result.FinalError = math.Abs(float64(currentLevel-targetLevel)) / float64(DefaultQuantizationLevels-1)

		if math.Abs(currentConductance-targetConductance) <= tolerance {
			result.Converged = true
			return result, nil
		}
	}

	return result, nil
}

// AnalysisReport contains a complete analysis of the array state.
type AnalysisReport struct {
	Timestamp   time.Time `json:"timestamp"`
	ArraySize   [2]int    `json:"array_size"`
	TotalMACs   int       `json:"total_macs"`
	Levels      int       `json:"levels"`
	NoiseLevel  float64   `json:"noise_level"`
	ADCBits     int       `json:"adc_bits"`

	// Accuracy metrics
	IdealAccuracy  float64 `json:"ideal_accuracy,omitempty"`
	ActualAccuracy float64 `json:"actual_accuracy,omitempty"`
	AccuracyLoss   float64 `json:"accuracy_loss,omitempty"`
	RMSE           float64 `json:"rmse,omitempty"`

	// IR Drop metrics
	MaxIRDrop float64 `json:"max_ir_drop,omitempty"`
	AvgIRDrop float64 `json:"avg_ir_drop,omitempty"`

	// Sneak Path metrics
	MaxSneakRatio float64 `json:"max_sneak_ratio,omitempty"`
	AvgSneakRatio float64 `json:"avg_sneak_ratio,omitempty"`

	// Energy metrics
	TotalEnergy         float64 `json:"total_energy_pj,omitempty"`
	GPUEquivalentEnergy float64 `json:"gpu_equivalent_energy_pj,omitempty"`
	EnergyEfficiency    float64 `json:"energy_efficiency,omitempty"`

	// Statistics
	TotalReads  int64 `json:"total_reads"`
	TotalWrites int64 `json:"total_writes"`
}

// GenerateReport creates a complete analysis report.
func (a *Array) GenerateReport(mvmResult *MVMResult) *AnalysisReport {
	report := &AnalysisReport{
		Timestamp:  time.Now(),
		ArraySize:  [2]int{a.config.Rows, a.config.Cols},
		TotalMACs:  a.config.Rows * a.config.Cols,
		Levels:     DefaultQuantizationLevels,
		NoiseLevel: a.config.NoiseLevel,
		ADCBits:    a.config.ADCBits,
	}

	reads, writes := a.GetStats()
	report.TotalReads = reads
	report.TotalWrites = writes

	if mvmResult != nil {
		report.RMSE = mvmResult.RMSE
		report.AccuracyLoss = mvmResult.AccuracyLoss
		report.TotalEnergy = mvmResult.TotalEnergy
		report.GPUEquivalentEnergy = mvmResult.GPUEquivalentEnergy
		report.EnergyEfficiency = mvmResult.EnergyEfficiency

		if mvmResult.IRDropAnalysis != nil {
			report.MaxIRDrop = mvmResult.IRDropAnalysis.MaxIRDrop
			report.AvgIRDrop = mvmResult.IRDropAnalysis.AvgIRDrop
		}

		if mvmResult.SneakPathAnalysis != nil {
			report.MaxSneakRatio = mvmResult.SneakPathAnalysis.MaxSneakRatio
			report.AvgSneakRatio = mvmResult.SneakPathAnalysis.AvgSneakRatio
		}
	}

	return report
}

// ExportWeightsCSV exports the conductance matrix to a CSV file.
func (a *Array) ExportWeightsCSV(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"row", "col", "level", "conductance", "conductance_uS"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data
	matrix := a.GetConductanceMatrix()
	for i := 0; i < a.config.Rows; i++ {
		for j := 0; j < a.config.Cols; j++ {
			G := matrix[i][j]
			level := GetLevel(G)
			// Assume conductance range 1-100 µS
			conductanceUS := G*99 + 1

			record := []string{
				strconv.Itoa(i),
				strconv.Itoa(j),
				strconv.Itoa(level),
				fmt.Sprintf("%.6f", G),
				fmt.Sprintf("%.2f", conductanceUS),
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
	}

	return nil
}

// ExportAnalysisJSON exports the analysis report to a JSON file.
func (a *Array) ExportAnalysisJSON(path string, mvmResult *MVMResult) error {
	report := a.GenerateReport(mvmResult)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// AccuracyDegradation tracks accuracy degradation from non-idealities.
type AccuracyDegradation struct {
	BaselineAccuracy float64            `json:"baseline_accuracy"`
	Degradations     []DegradationStep  `json:"degradations"`
	FinalAccuracy    float64            `json:"final_accuracy"`
}

// DegradationStep represents one source of accuracy loss.
type DegradationStep struct {
	Source      string  `json:"source"`
	AccuracyNow float64 `json:"accuracy_now"`
	Loss        float64 `json:"loss"`
}

// ComputeAccuracyDegradation computes stepwise accuracy loss.
func (a *Array) ComputeAccuracyDegradation(input []float64, baselineAccuracy float64) (*AccuracyDegradation, error) {
	result := &AccuracyDegradation{
		BaselineAccuracy: baselineAccuracy,
		Degradations:     []DegradationStep{},
	}

	currentAccuracy := baselineAccuracy

	// Step 1: ADC/DAC quantization
	opts := &MVMOptions{}
	mvmIdeal, _ := a.MVMWithNonIdealities(input, opts)
	quantLoss := mvmIdeal.RMSE * 100 / 3.0 // Empirical: 3% RMSE = 1% accuracy loss
	currentAccuracy -= quantLoss
	result.Degradations = append(result.Degradations, DegradationStep{
		Source:      "ADC/DAC Quantization",
		AccuracyNow: currentAccuracy,
		Loss:        quantLoss,
	})

	// Step 2: Add IR drop
	opts.EnableIRDrop = true
	mvmIR, _ := a.MVMWithNonIdealities(input, opts)
	irLoss := (mvmIR.RMSE - mvmIdeal.RMSE) * 100 / 3.0
	currentAccuracy -= irLoss
	result.Degradations = append(result.Degradations, DegradationStep{
		Source:      "IR Drop",
		AccuracyNow: currentAccuracy,
		Loss:        irLoss,
	})

	// Step 3: Add device variation
	opts.EnableVariation = true
	mvmVar, _ := a.MVMWithNonIdealities(input, opts)
	varLoss := (mvmVar.RMSE - mvmIR.RMSE) * 100 / 3.0
	currentAccuracy -= varLoss
	result.Degradations = append(result.Degradations, DegradationStep{
		Source:      "Device Variation",
		AccuracyNow: currentAccuracy,
		Loss:        varLoss,
	})

	// Step 4: Add sneak paths
	opts.EnableSneakPaths = true
	mvmSneak, _ := a.MVMWithNonIdealities(input, opts)
	sneakLoss := (mvmSneak.RMSE - mvmVar.RMSE) * 100 / 3.0
	currentAccuracy -= sneakLoss
	result.Degradations = append(result.Degradations, DegradationStep{
		Source:      "Sneak Paths",
		AccuracyNow: currentAccuracy,
		Loss:        sneakLoss,
	})

	result.FinalAccuracy = currentAccuracy

	return result, nil
}
