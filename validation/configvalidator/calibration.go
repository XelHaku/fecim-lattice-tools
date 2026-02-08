package configvalidator

import (
	"fmt"
	"math"
)

// CalibrationConfig represents a ferroelectric calibration configuration.
type CalibrationConfig struct {
	Version      int                          `json:"version"`
	MaterialName string                       `json:"material_name"`
	NumLevels    int                          `json:"num_levels"`
	Calibrations map[string]CalibrationData   `json:"calibrations"`
	SavedAt      string                       `json:"saved_at,omitempty"`
}

// CalibrationData contains calibration arrays for a specific temperature.
type CalibrationData struct {
	TemperatureK    int       `json:"temperature_k"`
	CalibrationUp   []float64 `json:"calibration_up"`
	CalibrationDown []float64 `json:"calibration_down"`
	CalibUpLow      []float64 `json:"calib_up_low"`
	CalibUpHigh     []float64 `json:"calib_up_high"`
	CalibDownLow    []float64 `json:"calib_down_low"`
	CalibDownHigh   []float64 `json:"calib_down_high"`
	LastErrorUp     []float64 `json:"last_error_up"`
	LastErrorDown   []float64 `json:"last_error_down"`
	RelaxCompUp     []float64 `json:"relax_comp_up"`
	RelaxCompDown   []float64 `json:"relax_comp_down"`
}

// Calibration config constraints
const (
	MinCalibrationVersion = 1
	MaxCalibrationVersion = 4 // v4 adds extended calibration fields
	MinNumLevels          = 2
	MaxNumLevels          = 256
	MinTemperatureK       = 1
	MaxTemperatureK       = 1000
	MinRelaxComp          = 0.0
	MaxRelaxComp          = 1.0
)

// validateCalibrationConfig validates a calibration configuration.
func validateCalibrationConfig(data map[string]any, result *ValidationResult) {
	// Validate version
	version, ok := getInt(data, "version")
	if !ok {
		result.AddError("version", "required field missing or invalid type", data["version"])
	} else {
		if version < MinCalibrationVersion || version > MaxCalibrationVersion {
			result.AddError("version", fmt.Sprintf("must be between %d and %d", MinCalibrationVersion, MaxCalibrationVersion), version)
		}
	}
	
	// Validate material_name
	materialName, ok := getString(data, "material_name")
	if !ok {
		result.AddError("material_name", "required field missing or invalid type", data["material_name"])
	} else if materialName == "" {
		result.AddError("material_name", "must not be empty", materialName)
	}
	
	// Validate num_levels
	numLevels, ok := getInt(data, "num_levels")
	if !ok {
		result.AddError("num_levels", "required field missing or invalid type", data["num_levels"])
		numLevels = 0 // Set to 0 to skip array length validation
	} else {
		if numLevels < MinNumLevels {
			result.AddError("num_levels", fmt.Sprintf("must be at least %d", MinNumLevels), numLevels)
		}
		if numLevels > MaxNumLevels {
			result.AddError("num_levels", fmt.Sprintf("must be at most %d", MaxNumLevels), numLevels)
		}
	}
	
	// Validate calibrations
	calibrations, ok := getMap(data, "calibrations")
	if !ok {
		result.AddError("calibrations", "required field missing or invalid type", nil)
		return
	}
	
	if len(calibrations) == 0 {
		result.AddError("calibrations", "must contain at least one temperature calibration", nil)
		return
	}
	
	// Validate each temperature's calibration data
	for tempKey, tempData := range calibrations {
		fieldPrefix := fmt.Sprintf("calibrations.%s", tempKey)
		
		tempMap, ok := tempData.(map[string]any)
		if !ok {
			result.AddError(fieldPrefix, "must be an object", tempData)
			continue
		}
		
		validateCalibrationData(tempMap, fieldPrefix, numLevels, version, result)
	}
	
	// Validate saved_at if present
	if savedAt, ok := getString(data, "saved_at"); ok && savedAt != "" {
		if !validateTimestamp(savedAt) {
			result.AddWarning("saved_at", "invalid ISO 8601 timestamp format", savedAt)
		}
	}
}

// validateCalibrationData validates a single temperature's calibration data.
func validateCalibrationData(data map[string]any, prefix string, numLevels, version int, result *ValidationResult) {
	// Validate temperature_k
	tempK, ok := getInt(data, "temperature_k")
	if !ok {
		result.AddError(prefix+".temperature_k", "required field missing or invalid type", data["temperature_k"])
	} else {
		if tempK < MinTemperatureK || tempK > MaxTemperatureK {
			result.AddError(prefix+".temperature_k", fmt.Sprintf("must be between %d and %d K", MinTemperatureK, MaxTemperatureK), tempK)
		}
	}
	
	// Required arrays
	requiredArrays := []string{"calibration_up", "calibration_down"}
	if version >= 2 {
		requiredArrays = append(requiredArrays,
			"calib_up_low", "calib_up_high",
			"calib_down_low", "calib_down_high",
			"last_error_up", "last_error_down",
			"relax_comp_up", "relax_comp_down",
		)
	}
	
	for _, arrayName := range requiredArrays {
		fieldPath := prefix + "." + arrayName
		arr, ok := getFloatArray(data, arrayName)
		if !ok {
			result.AddError(fieldPath, "required field missing or invalid type", data[arrayName])
			continue
		}
		
		// Validate array length matches num_levels
		if numLevels > 0 && len(arr) != numLevels {
			result.AddError(fieldPath, fmt.Sprintf("array length (%d) must match num_levels (%d)", len(arr), numLevels), len(arr))
		}
		
		// Validate array values
		validateCalibrationArray(arr, arrayName, fieldPath, result)
	}
	
	// Validate monotonicity of calibration arrays
	if numLevels > 0 {
		validateCalibrationMonotonicity(data, prefix, result)
	}
}

// validateCalibrationArray validates individual calibration array values.
func validateCalibrationArray(arr []float64, arrayName, fieldPath string, result *ValidationResult) {
	for i, v := range arr {
		// Check for NaN or Inf
		if math.IsNaN(v) {
			result.AddError(fmt.Sprintf("%s[%d]", fieldPath, i), "contains NaN", v)
		}
		if math.IsInf(v, 0) {
			result.AddError(fmt.Sprintf("%s[%d]", fieldPath, i), "contains Inf", v)
		}
		
		// Validate relax_comp arrays are in [0, 1] range
		if arrayName == "relax_comp_up" || arrayName == "relax_comp_down" {
			if v < MinRelaxComp || v > MaxRelaxComp {
				result.AddError(fmt.Sprintf("%s[%d]", fieldPath, i),
					fmt.Sprintf("must be between %.1f and %.1f", MinRelaxComp, MaxRelaxComp), v)
			}
		}
	}
}

// validateCalibrationMonotonicity checks that calibration arrays are properly ordered.
func validateCalibrationMonotonicity(data map[string]any, prefix string, result *ValidationResult) {
	// calibration_up should generally be non-decreasing
	if upArr, ok := getFloatArray(data, "calibration_up"); ok && len(upArr) > 1 {
		isMonotonic := true
		for i := 1; i < len(upArr); i++ {
			if upArr[i] < upArr[i-1] {
				isMonotonic = false
				break
			}
		}
		if !isMonotonic {
			result.AddWarning(prefix+".calibration_up", "array is not monotonically non-decreasing (may indicate calibration issue)", nil)
		}
	}
	
	// calibration_down should generally be non-increasing (all negative, becoming less negative)
	if downArr, ok := getFloatArray(data, "calibration_down"); ok && len(downArr) > 1 {
		isMonotonic := true
		for i := 1; i < len(downArr); i++ {
			if downArr[i] < downArr[i-1] {
				isMonotonic = false
				break
			}
		}
		if !isMonotonic {
			result.AddWarning(prefix+".calibration_down", "array is not monotonically non-decreasing (may indicate calibration issue)", nil)
		}
	}
	
	// Validate that up values are non-negative and down values are non-positive at boundaries
	if upArr, ok := getFloatArray(data, "calibration_up"); ok && len(upArr) > 0 {
		if upArr[0] < 0 {
			result.AddWarning(prefix+".calibration_up[0]", "first element is typically 0 or positive", upArr[0])
		}
		if upArr[len(upArr)-1] < 0 {
			result.AddWarning(prefix+".calibration_up", "last element is typically positive (max positive coercive field)", upArr[len(upArr)-1])
		}
	}
	
	if downArr, ok := getFloatArray(data, "calibration_down"); ok && len(downArr) > 0 {
		if downArr[0] > 0 {
			result.AddWarning(prefix+".calibration_down[0]", "first element is typically negative (max negative coercive field)", downArr[0])
		}
		if downArr[len(downArr)-1] > 0 {
			result.AddWarning(prefix+".calibration_down", "last element is typically 0 or negative", downArr[len(downArr)-1])
		}
	}
}
