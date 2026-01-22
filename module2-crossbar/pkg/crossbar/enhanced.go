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
	"time"
)

// MVMOptions configures which non-idealities to include in MVM computation.
type MVMOptions struct {
	EnableIRDrop     bool
	EnableSneakPaths bool
	EnableVariation  bool
	EnableDrift      bool
	Temperature      float64 // Kelvin (default 300K = 27C)
}

// DefaultMVMOptions returns options with all non-idealities enabled.
func DefaultMVMOptions() *MVMOptions {
	return &MVMOptions{
		EnableIRDrop:     true,
		EnableSneakPaths: true,
		EnableVariation:  true,
		EnableDrift:      false, // Drift usually simulated separately over time
		Temperature:      300.0, // Room temperature
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
		// Apply temperature effect on wire resistance
		if opts.Temperature != 300.0 {
			tempFactor := 1.0 + 0.00393*(opts.Temperature-300.0) // Copper TCR
			params.RwordLine *= tempFactor
			params.RbitLine *= tempFactor
		}
		irAnalysis := a.AnalyzeIRDrop(input, params)
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
			sneakCurrent := a.computeSneakCurrentForRow(i, input)
			sum += sneakCurrent
		}

		// Normalize and quantize through ADC
		normalizedSum := sum / float64(len(input))
		result.ActualOutput[i] = a.quantizeADC(normalizedSum)
		a.totalReads++
	}

	// Compute sneak path analysis for center cell
	if opts.EnableSneakPaths {
		centerRow := a.config.Rows / 2
		centerCol := a.config.Cols / 2
		result.SneakPathAnalysis = a.AnalyzeSneakPaths(centerRow, centerCol)
	}

	// Step 4: Compute error metrics
	result.computeErrorMetrics()

	// Step 5: Compute energy metrics
	result.computeEnergyMetrics(a.config.Rows, len(input), a.config.ADCBits)

	return result, nil
}

// computeSneakCurrentForRow computes total sneak current affecting a row.
func (a *Array) computeSneakCurrentForRow(row int, input []float64) float64 {
	var sneakCurrent float64
	sneakFactor := 0.001 // 0.1% sneak coupling factor

	for j := 0; j < len(input); j++ {
		// Three-cell sneak paths from other rows
		for i := 0; i < a.config.Rows; i++ {
			if i == row {
				continue
			}
			// Simplified sneak model
			g1 := a.cells[row][j].Conductance
			g2 := a.cells[i][j].Conductance
			if g1 > 0.01 && g2 > 0.01 {
				sneakCurrent += input[j] * sneakFactor * (g1 * g2) / (g1 + g2)
			}
		}
	}

	return sneakCurrent
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

	if targetLevel < 0 || targetLevel >= FeCIMLevels {
		return nil, fmt.Errorf("target level %d out of range [0, %d)", targetLevel, FeCIMLevels)
	}

	result := &WriteVerifyResult{
		TargetLevel: targetLevel,
	}

	targetConductance := float64(targetLevel) / float64(FeCIMLevels-1)
	tolerance := cfg.Tolerance / float64(FeCIMLevels-1)

	currentConductance := a.cells[row][col].Conductance

	for iter := 0; iter < cfg.MaxIterations; iter++ {
		result.Iterations = iter + 1

		// Simulate write with variation
		error := (rand.Float64()*2 - 1) * a.config.NoiseLevel * cfg.PulseStep
		newConductance := currentConductance + (targetConductance-currentConductance)*cfg.PulseStep + error

		// Clamp to valid range
		newConductance = math.Max(0, math.Min(1, newConductance))

		// Quantize to 30 levels
		newConductance = QuantizeTo30Levels(newConductance)
		a.cells[row][col].Conductance = newConductance
		a.cells[row][col].SwitchingCount++
		a.totalWrites++
		currentConductance = newConductance

		// Check convergence
		currentLevel := GetLevel(currentConductance)
		result.AchievedLevel = currentLevel
		result.FinalError = math.Abs(float64(currentLevel-targetLevel)) / float64(FeCIMLevels-1)

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
		Levels:     FeCIMLevels,
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
