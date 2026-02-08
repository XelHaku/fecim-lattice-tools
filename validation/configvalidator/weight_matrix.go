package configvalidator

import (
	"fmt"
	"math"
)

// WeightMatrixConfig represents a neural network weight matrix configuration.
type WeightMatrixConfig struct {
	Name    string      `json:"name"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
	Weights [][]float64 `json:"weights"`
}

// Weight matrix constraints
const (
	MinWeightMatrixDim = 1
	MaxWeightMatrixDim = 65536
)

// validateWeightMatrixConfig validates a weight matrix configuration.
func validateWeightMatrixConfig(data map[string]any, result *ValidationResult) {
	// Validate name
	name, ok := getString(data, "name")
	if !ok {
		result.AddError("name", "required field missing or invalid type", data["name"])
	} else if name == "" {
		result.AddError("name", "must not be empty", name)
	}
	
	// Validate rows
	rows, ok := getInt(data, "rows")
	if !ok {
		result.AddError("rows", "required field missing or invalid type", data["rows"])
		rows = 0
	} else {
		if rows < MinWeightMatrixDim {
			result.AddError("rows", fmt.Sprintf("must be at least %d", MinWeightMatrixDim), rows)
		}
		if rows > MaxWeightMatrixDim {
			result.AddError("rows", fmt.Sprintf("must be at most %d", MaxWeightMatrixDim), rows)
		}
	}
	
	// Validate cols
	cols, ok := getInt(data, "cols")
	if !ok {
		result.AddError("cols", "required field missing or invalid type", data["cols"])
		cols = 0
	} else {
		if cols < MinWeightMatrixDim {
			result.AddError("cols", fmt.Sprintf("must be at least %d", MinWeightMatrixDim), cols)
		}
		if cols > MaxWeightMatrixDim {
			result.AddError("cols", fmt.Sprintf("must be at most %d", MaxWeightMatrixDim), cols)
		}
	}
	
	// Validate weights array
	weights, ok := data["weights"].([]any)
	if !ok {
		result.AddError("weights", "required field missing or invalid type", nil)
		return
	}
	
	// Validate row count matches
	if rows > 0 && len(weights) != rows {
		result.AddError("weights", fmt.Sprintf("row count (%d) must match rows (%d)", len(weights), rows), len(weights))
	}
	
	// Track weight statistics
	var weightMin, weightMax float64 = math.Inf(1), math.Inf(-1)
	var weightSum float64
	var weightCount int
	var hasNaN, hasInf bool
	
	for i, row := range weights {
		rowArr, ok := row.([]any)
		if !ok {
			result.AddError(fmt.Sprintf("weights[%d]", i), "must be an array", row)
			continue
		}
		
		// Validate column count matches
		if cols > 0 && len(rowArr) != cols {
			result.AddError(fmt.Sprintf("weights[%d]", i), fmt.Sprintf("column count (%d) must match cols (%d)", len(rowArr), cols), len(rowArr))
		}
		
		// Validate each weight value
		for j, w := range rowArr {
			var wf float64
			switch v := w.(type) {
			case float64:
				wf = v
			case int:
				wf = float64(v)
			default:
				result.AddError(fmt.Sprintf("weights[%d][%d]", i, j), "must be a number", w)
				continue
			}
			
			// Check for NaN/Inf
			if math.IsNaN(wf) {
				hasNaN = true
				result.AddError(fmt.Sprintf("weights[%d][%d]", i, j), "contains NaN", wf)
			} else if math.IsInf(wf, 0) {
				hasInf = true
				result.AddError(fmt.Sprintf("weights[%d][%d]", i, j), "contains Inf", wf)
			} else {
				// Track statistics
				if wf < weightMin {
					weightMin = wf
				}
				if wf > weightMax {
					weightMax = wf
				}
				weightSum += wf
				weightCount++
			}
		}
	}
	
	// Add warnings for unusual weight distributions
	if weightCount > 0 && !hasNaN && !hasInf {
		weightMean := weightSum / float64(weightCount)
		
		// Check if weights are outside typical normalized ranges
		if weightMax > 10 || weightMin < -10 {
			result.AddWarning("weights", fmt.Sprintf("weights outside typical range [-10, 10]: min=%.4f, max=%.4f", weightMin, weightMax), nil)
		}
		
		// Check for potentially uninitialized weights (all zeros)
		if weightMin == 0 && weightMax == 0 {
			result.AddWarning("weights", "all weights are zero (may be uninitialized)", nil)
		}
		
		// Check for highly imbalanced weights
		if math.Abs(weightMean) > 1 {
			result.AddWarning("weights", fmt.Sprintf("weight mean (%.4f) is unusually large, may indicate initialization issue", weightMean), nil)
		}
	}
}
