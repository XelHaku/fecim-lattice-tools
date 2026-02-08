package configvalidator

import (
	"fmt"
	"math"
)

// ArrayDesignConfig represents a crossbar array design configuration.
type ArrayDesignConfig struct {
	Config ArrayConfigParams `json:"config"`
	Cells  []CellConfig      `json:"cells"`
	Stats  ArrayStats        `json:"stats,omitempty"`
}

// ArrayConfigParams contains the main array configuration parameters.
type ArrayConfigParams struct {
	Name          string            `json:"name"`
	Mode          int               `json:"mode"`
	ArrayRows     int               `json:"array_rows"`
	ArrayCols     int               `json:"array_cols"`
	Technology    string            `json:"technology"`
	Architecture  string            `json:"architecture"`
	CellPitch     float64           `json:"cell_pitch"`
	RowHeight     float64           `json:"row_height"`
	Levels        int               `json:"levels"`
	GMin          float64           `json:"g_min"`
	GMax          float64           `json:"g_max"`
	VProgMin      float64           `json:"v_prog_min"`
	VProgMax      float64           `json:"v_prog_max"`
	TPulse        float64           `json:"t_pulse"`
	Peripherals   PeripheralConfig  `json:"peripherals"`
	ComputeConfig ComputeConfig     `json:"compute_config"`
}

// PeripheralConfig contains DAC/ADC configuration.
type PeripheralConfig struct {
	DACBits   int     `json:"dac_bits"`
	ADCBits   int     `json:"adc_bits"`
	TIAGain   float64 `json:"tia_gain"`
	VDD       float64 `json:"vdd"`
	ClockFreq float64 `json:"clock_freq"`
}

// ComputeConfig contains compute-specific parameters.
type ComputeConfig struct {
	InitialWeights  [][]float64 `json:"initial_weights"`
	QuantLevels     int         `json:"quant_levels"`
	AccumulatorBits int         `json:"accumulator_bits"`
	ActivationFunc  string      `json:"activation_func"`
}

// CellConfig represents a single memristor cell configuration.
type CellConfig struct {
	Row           int     `json:"row"`
	Col           int     `json:"col"`
	Level         int     `json:"level"`
	Conductance   float64 `json:"conductance"`
	Resistance    float64 `json:"resistance"`
	ProgramV      float64 `json:"program_v"`
	InitialWeight float64 `json:"initial_weight"`
}

// ArrayStats contains computed statistics for the array.
type ArrayStats struct {
	TotalCells    int     `json:"total_cells"`
	ActiveCells   int     `json:"active_cells"`
	AreaMM2       float64 `json:"area_mm2"`
	PowerMW       float64 `json:"power_mw"`
	ThroughputGOPS float64 `json:"throughput_gops"`
	QuantMSE      float64 `json:"quant_mse"`
	QuantPSNRdB   float64 `json:"quant_psnr_db"`
	WeightMin     float64 `json:"weight_min"`
	WeightMax     float64 `json:"weight_max"`
	UsedCells     int     `json:"used_cells"`
}

// Array design constraints
const (
	MinArrayDim     = 1
	MaxArrayDim     = 4096
	MinLevels       = 2
	MaxLevels       = 256
	MinConductance  = 0.0
	MaxConductance  = 1e6  // 1 MS
	MinVoltage      = 0.0
	MaxVoltage      = 10.0
	MinDACBits      = 1
	MaxDACBits      = 16
	MinTPulse       = 0.1   // ns
	MaxTPulse       = 1e9   // ns (1s)
)

// Valid technologies
var validTechnologies = map[string]bool{
	"SKY130":   true,
	"GF180":    true,
	"TSMC28":   true,
	"TSMC16":   true,
	"TSMC7":    true,
	"generic":  true,
}

// Valid architectures
var validArchitectures = map[string]bool{
	"passive": true,
	"1T1R":    true,
	"2T2R":    true,
	"1S1R":    true,
}

// Valid activation functions
var validActivationFuncs = map[string]bool{
	"none":    true,
	"relu":    true,
	"sigmoid": true,
	"tanh":    true,
	"softmax": true,
}

// validateArrayDesignConfig validates an array design configuration.
func validateArrayDesignConfig(data map[string]any, result *ValidationResult) {
	// Get the config object
	config, ok := getMap(data, "config")
	if !ok {
		result.AddError("config", "required field missing or invalid type", nil)
		return
	}
	
	// Validate config parameters
	arrayRows, arrayCols := validateArrayConfigParams(config, result)
	
	// Validate cells array
	validateCellsArray(data, arrayRows, arrayCols, config, result)
	
	// Validate stats if present
	if stats, ok := getMap(data, "stats"); ok {
		validateArrayStats(stats, arrayRows, arrayCols, result)
	}
}

// validateArrayConfigParams validates the main config parameters and returns array dimensions.
func validateArrayConfigParams(config map[string]any, result *ValidationResult) (int, int) {
	// Validate name (optional for partial configs)
	name, ok := getString(config, "name")
	if ok && name == "" {
		result.AddWarning("config.name", "empty name provided", name)
	}
	
	// Validate array dimensions
	arrayRows, ok := getInt(config, "array_rows")
	if !ok {
		result.AddError("config.array_rows", "required field missing or invalid type", config["array_rows"])
		arrayRows = 0
	} else {
		if arrayRows < MinArrayDim || arrayRows > MaxArrayDim {
			result.AddError("config.array_rows", fmt.Sprintf("must be between %d and %d", MinArrayDim, MaxArrayDim), arrayRows)
		}
	}
	
	arrayCols, ok := getInt(config, "array_cols")
	if !ok {
		result.AddError("config.array_cols", "required field missing or invalid type", config["array_cols"])
		arrayCols = 0
	} else {
		if arrayCols < MinArrayDim || arrayCols > MaxArrayDim {
			result.AddError("config.array_cols", fmt.Sprintf("must be between %d and %d", MinArrayDim, MaxArrayDim), arrayCols)
		}
	}
	
	// Validate technology (optional - some configs don't specify)
	if tech, ok := getString(config, "technology"); ok {
		if !validTechnologies[tech] {
			result.AddWarning("config.technology", "unknown technology (expected: SKY130, GF180, TSMC28, TSMC16, TSMC7, generic)", tech)
		}
	}
	
	// Validate architecture if present
	if arch, ok := getString(config, "architecture"); ok {
		if !validArchitectures[arch] {
			result.AddWarning("config.architecture", "unknown architecture (expected: passive, 1T1R, 2T2R, 1S1R)", arch)
		}
	}
	
	// Validate levels
	levels, ok := getInt(config, "levels")
	if !ok {
		result.AddError("config.levels", "required field missing or invalid type", config["levels"])
	} else {
		if levels < MinLevels || levels > MaxLevels {
			result.AddError("config.levels", fmt.Sprintf("must be between %d and %d", MinLevels, MaxLevels), levels)
		}
	}
	
	// Validate conductance range
	gMin, hasGMin := getFloat(config, "g_min")
	gMax, hasGMax := getFloat(config, "g_max")
	if hasGMin && hasGMax {
		if gMin < MinConductance {
			result.AddError("config.g_min", "must be non-negative", gMin)
		}
		if gMax <= gMin {
			result.AddError("config.g_max", "must be greater than g_min", gMax)
		}
		if gMax > MaxConductance {
			result.AddWarning("config.g_max", "unusually high conductance value", gMax)
		}
	}
	
	// Validate voltage range
	vMin, hasVMin := getFloat(config, "v_prog_min")
	vMax, hasVMax := getFloat(config, "v_prog_max")
	if hasVMin && hasVMax {
		if vMin < MinVoltage {
			result.AddError("config.v_prog_min", "must be non-negative", vMin)
		}
		if vMax <= vMin {
			result.AddError("config.v_prog_max", "must be greater than v_prog_min", vMax)
		}
		if vMax > MaxVoltage {
			result.AddWarning("config.v_prog_max", "unusually high programming voltage", vMax)
		}
	}
	
	// Validate t_pulse if present
	if tPulse, ok := getFloat(config, "t_pulse"); ok {
		if tPulse < MinTPulse {
			result.AddError("config.t_pulse", fmt.Sprintf("must be at least %.1f ns", MinTPulse), tPulse)
		}
		if tPulse > MaxTPulse {
			result.AddWarning("config.t_pulse", "unusually long pulse duration", tPulse)
		}
	}
	
	// Validate peripherals if present
	if periph, ok := getMap(config, "peripherals"); ok {
		validatePeripherals(periph, result)
	}
	
	// Validate compute_config if present
	if computeCfg, ok := getMap(config, "compute_config"); ok {
		validateComputeConfig(computeCfg, arrayRows, arrayCols, result)
	}
	
	return arrayRows, arrayCols
}

// validatePeripherals validates peripheral configuration.
func validatePeripherals(periph map[string]any, result *ValidationResult) {
	// Validate DAC bits
	if dacBits, ok := getInt(periph, "dac_bits"); ok {
		if dacBits < MinDACBits || dacBits > MaxDACBits {
			result.AddError("config.peripherals.dac_bits", fmt.Sprintf("must be between %d and %d", MinDACBits, MaxDACBits), dacBits)
		}
	}
	
	// Validate ADC bits
	if adcBits, ok := getInt(periph, "adc_bits"); ok {
		if adcBits < MinDACBits || adcBits > MaxDACBits {
			result.AddError("config.peripherals.adc_bits", fmt.Sprintf("must be between %d and %d", MinDACBits, MaxDACBits), adcBits)
		}
	}
	
	// Validate TIA gain
	if tiaGain, ok := getFloat(periph, "tia_gain"); ok {
		if tiaGain <= 0 {
			result.AddError("config.peripherals.tia_gain", "must be positive", tiaGain)
		}
	}
	
	// Validate VDD
	if vdd, ok := getFloat(periph, "vdd"); ok {
		if vdd <= 0 || vdd > MaxVoltage {
			result.AddError("config.peripherals.vdd", fmt.Sprintf("must be between 0 and %.1f V", MaxVoltage), vdd)
		}
	}
	
	// Validate clock frequency (MHz)
	if clockFreq, ok := getFloat(periph, "clock_freq"); ok {
		if clockFreq <= 0 {
			result.AddError("config.peripherals.clock_freq", "must be positive", clockFreq)
		}
		if clockFreq > 10000 { // 10 GHz
			result.AddWarning("config.peripherals.clock_freq", "unusually high clock frequency (MHz)", clockFreq)
		}
	}
}

// validateComputeConfig validates compute configuration.
func validateComputeConfig(computeCfg map[string]any, arrayRows, arrayCols int, result *ValidationResult) {
	// Validate initial_weights dimensions
	if weights, ok := computeCfg["initial_weights"].([]any); ok {
		if arrayRows > 0 && len(weights) != arrayRows {
			result.AddError("config.compute_config.initial_weights", 
				fmt.Sprintf("row count (%d) must match array_rows (%d)", len(weights), arrayRows), len(weights))
		}
		
		for i, row := range weights {
			if rowArr, ok := row.([]any); ok {
				if arrayCols > 0 && len(rowArr) != arrayCols {
					result.AddError(fmt.Sprintf("config.compute_config.initial_weights[%d]", i),
						fmt.Sprintf("column count (%d) must match array_cols (%d)", len(rowArr), arrayCols), len(rowArr))
				}
				
				// Validate weight values
				for j, w := range rowArr {
					if wf, ok := w.(float64); ok {
						if wf < -1.0 || wf > 1.0 {
							result.AddWarning(fmt.Sprintf("config.compute_config.initial_weights[%d][%d]", i, j),
								"weight outside [-1, 1] range", wf)
						}
						if math.IsNaN(wf) || math.IsInf(wf, 0) {
							result.AddError(fmt.Sprintf("config.compute_config.initial_weights[%d][%d]", i, j),
								"must be a finite number", wf)
						}
					}
				}
			}
		}
	}
	
	// Validate quant_levels
	if quantLevels, ok := getInt(computeCfg, "quant_levels"); ok {
		if quantLevels < MinLevels {
			result.AddError("config.compute_config.quant_levels", fmt.Sprintf("must be at least %d", MinLevels), quantLevels)
		}
	}
	
	// Validate accumulator_bits
	if accBits, ok := getInt(computeCfg, "accumulator_bits"); ok {
		if accBits < 8 || accBits > 64 {
			result.AddError("config.compute_config.accumulator_bits", "must be between 8 and 64", accBits)
		}
	}
	
	// Validate activation_func
	if actFunc, ok := getString(computeCfg, "activation_func"); ok {
		if !validActivationFuncs[actFunc] {
			result.AddWarning("config.compute_config.activation_func", "unknown activation function", actFunc)
		}
	}
}

// validateCellsArray validates the cells array.
func validateCellsArray(data map[string]any, arrayRows, arrayCols int, config map[string]any, result *ValidationResult) {
	cells, ok := data["cells"].([]any)
	if !ok {
		result.AddError("cells", "required field missing or invalid type", nil)
		return
	}
	
	expectedCells := arrayRows * arrayCols
	if expectedCells > 0 && len(cells) != expectedCells {
		result.AddError("cells", fmt.Sprintf("length (%d) must match array_rows * array_cols (%d)", len(cells), expectedCells), len(cells))
	}
	
	// Get levels for validation
	levels, _ := getInt(config, "levels")
	gMin, hasGMin := getFloat(config, "g_min")
	gMax, hasGMax := getFloat(config, "g_max")
	vMin, hasVMin := getFloat(config, "v_prog_min")
	vMax, hasVMax := getFloat(config, "v_prog_max")
	
	seenPositions := make(map[string]bool)
	
	for i, cell := range cells {
		cellMap, ok := cell.(map[string]any)
		if !ok {
			result.AddError(fmt.Sprintf("cells[%d]", i), "must be an object", cell)
			continue
		}
		
		// Validate row/col
		row, hasRow := getInt(cellMap, "row")
		col, hasCol := getInt(cellMap, "col")
		
		if !hasRow {
			result.AddError(fmt.Sprintf("cells[%d].row", i), "required field missing", nil)
		} else if row < 0 || (arrayRows > 0 && row >= arrayRows) {
			result.AddError(fmt.Sprintf("cells[%d].row", i), fmt.Sprintf("must be between 0 and %d", arrayRows-1), row)
		}
		
		if !hasCol {
			result.AddError(fmt.Sprintf("cells[%d].col", i), "required field missing", nil)
		} else if col < 0 || (arrayCols > 0 && col >= arrayCols) {
			result.AddError(fmt.Sprintf("cells[%d].col", i), fmt.Sprintf("must be between 0 and %d", arrayCols-1), col)
		}
		
		// Check for duplicate positions
		if hasRow && hasCol {
			posKey := fmt.Sprintf("%d,%d", row, col)
			if seenPositions[posKey] {
				result.AddError(fmt.Sprintf("cells[%d]", i), fmt.Sprintf("duplicate cell at position (%d, %d)", row, col), nil)
			}
			seenPositions[posKey] = true
		}
		
		// Validate level
		if level, ok := getInt(cellMap, "level"); ok {
			if level < 0 || (levels > 0 && level >= levels) {
				result.AddError(fmt.Sprintf("cells[%d].level", i), fmt.Sprintf("must be between 0 and %d", levels-1), level)
			}
		}
		
		// Validate conductance
		if cond, ok := getFloat(cellMap, "conductance"); ok {
			if hasGMin && hasGMax {
				if cond < gMin || cond > gMax {
					result.AddWarning(fmt.Sprintf("cells[%d].conductance", i), fmt.Sprintf("outside configured range [%.2f, %.2f]", gMin, gMax), cond)
				}
			}
			if cond < 0 {
				result.AddError(fmt.Sprintf("cells[%d].conductance", i), "must be non-negative", cond)
			}
		}
		
		// Validate resistance
		if res, ok := getFloat(cellMap, "resistance"); ok {
			if res <= 0 {
				result.AddError(fmt.Sprintf("cells[%d].resistance", i), "must be positive", res)
			}
		}
		
		// Validate program_v
		if progV, ok := getFloat(cellMap, "program_v"); ok {
			if hasVMin && hasVMax {
				if progV < vMin || progV > vMax {
					result.AddWarning(fmt.Sprintf("cells[%d].program_v", i), fmt.Sprintf("outside configured range [%.2f, %.2f]", vMin, vMax), progV)
				}
			}
		}
		
		// Validate conductance/resistance consistency
		if cond, hasCond := getFloat(cellMap, "conductance"); hasCond {
			if res, hasRes := getFloat(cellMap, "resistance"); hasRes {
				if cond > 0 && res > 0 {
					expectedRes := 1.0 / (cond * 1e-6) // Assuming conductance is in µS
					relError := math.Abs(res-expectedRes) / expectedRes
					if relError > 0.01 { // 1% tolerance
						// Note: This is just a warning as units may vary
						result.AddWarning(fmt.Sprintf("cells[%d]", i), 
							"conductance and resistance may be inconsistent (check units)", nil)
					}
				}
			}
		}
	}
}

// validateArrayStats validates array statistics.
func validateArrayStats(stats map[string]any, arrayRows, arrayCols int, result *ValidationResult) {
	// Validate total_cells
	if totalCells, ok := getInt(stats, "total_cells"); ok {
		expected := arrayRows * arrayCols
		if expected > 0 && totalCells != expected {
			result.AddWarning("stats.total_cells", fmt.Sprintf("expected %d (rows * cols)", expected), totalCells)
		}
	}
	
	// Validate active_cells <= total_cells
	totalCells, _ := getInt(stats, "total_cells")
	if activeCells, ok := getInt(stats, "active_cells"); ok {
		if activeCells < 0 {
			result.AddError("stats.active_cells", "must be non-negative", activeCells)
		}
		if totalCells > 0 && activeCells > totalCells {
			result.AddError("stats.active_cells", "cannot exceed total_cells", activeCells)
		}
	}
	
	// Validate non-negative metrics
	nonNegativeFields := []string{"area_mm2", "power_mw", "throughput_gops", "quant_mse"}
	for _, field := range nonNegativeFields {
		if v, ok := getFloat(stats, field); ok && v < 0 {
			result.AddError("stats."+field, "must be non-negative", v)
		}
	}
	
	// Validate weight_min <= weight_max
	if wMin, hasMin := getFloat(stats, "weight_min"); hasMin {
		if wMax, hasMax := getFloat(stats, "weight_max"); hasMax {
			if wMin > wMax {
				result.AddError("stats.weight_min/weight_max", "weight_min must be <= weight_max", nil)
			}
		}
	}
}
