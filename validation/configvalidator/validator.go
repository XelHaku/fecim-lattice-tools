// Package configvalidator provides validation for FeCIM configuration files.
// It validates JSON configs for valid ranges, required fields, and type correctness.
package configvalidator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ValidationError represents a validation failure with context.
type ValidationError struct {
	Field   string // JSON path to the field (e.g., "calibrations.300.temperature_k")
	Message string // Human-readable error message
	Value   any    // The invalid value, if applicable
}

// Error implements the error interface for ValidationError.
func (e ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("%s: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult contains all validation errors found.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	FilePath string
	ConfigType string
}

// AddError adds an error to the result.
func (r *ValidationResult) AddError(field, message string, value any) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// AddWarning adds a warning (non-fatal) to the result.
func (r *ValidationResult) AddWarning(field, message string, value any) {
	r.Warnings = append(r.Warnings, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// String returns a human-readable summary.
func (r *ValidationResult) String() string {
	var sb strings.Builder
	if r.FilePath != "" {
		sb.WriteString(fmt.Sprintf("File: %s\n", r.FilePath))
	}
	if r.ConfigType != "" {
		sb.WriteString(fmt.Sprintf("Type: %s\n", r.ConfigType))
	}
	
	if r.Valid {
		sb.WriteString("Status: VALID\n")
	} else {
		sb.WriteString("Status: INVALID\n")
	}
	
	if len(r.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("Errors (%d):\n", len(r.Errors)))
		for _, err := range r.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
		}
	}
	
	if len(r.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("Warnings (%d):\n", len(r.Warnings)))
		for _, w := range r.Warnings {
			sb.WriteString(fmt.Sprintf("  - %s\n", w.Error()))
		}
	}
	
	return sb.String()
}

// ConfigType identifies the type of configuration.
type ConfigType string

const (
	ConfigTypeUnknown      ConfigType = "unknown"
	ConfigTypeCalibration  ConfigType = "calibration"
	ConfigTypePreisach     ConfigType = "preisach_state"
	ConfigTypeArrayDesign  ConfigType = "array_design"
	ConfigTypeWeightMatrix ConfigType = "weight_matrix"
	ConfigTypeOpenLane     ConfigType = "openlane"
)

// ValidateFile validates a JSON config file and returns detailed results.
func ValidateFile(path string) (*ValidationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	result := ValidateJSON(data)
	result.FilePath = path
	return result, nil
}

// ValidateJSON validates JSON config data and returns detailed results.
func ValidateJSON(data []byte) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// First, check if it's valid JSON (try object first, then array)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		// Check if it's a JSON array (which is valid JSON but not a config object)
		var arr []any
		if err2 := json.Unmarshal(data, &arr); err2 == nil {
			result.ConfigType = "json_array"
			result.AddWarning("", "JSON file contains an array, not a configuration object", nil)
			return result
		}
		result.AddError("", "invalid JSON syntax", err.Error())
		return result
	}
	
	// Detect config type and validate accordingly
	configType := detectConfigType(raw)
	result.ConfigType = string(configType)
	
	switch configType {
	case ConfigTypeCalibration:
		validateCalibrationConfig(raw, result)
	case ConfigTypePreisach:
		validatePreisachConfig(raw, result)
	case ConfigTypeArrayDesign:
		validateArrayDesignConfig(raw, result)
	case ConfigTypeWeightMatrix:
		validateWeightMatrixConfig(raw, result)
	case ConfigTypeOpenLane:
		validateOpenLaneConfig(raw, result)
	default:
		result.AddWarning("", "unknown config type, performing basic validation", nil)
		validateBasicJSON(raw, result)
	}
	
	return result
}

// ValidateDirectory validates all JSON files in a directory recursively.
func ValidateDirectory(dir string) ([]*ValidationResult, error) {
	var results []*ValidationResult
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip non-JSON files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}
		
		// Skip certain directories
		base := filepath.Base(filepath.Dir(path))
		if base == "node_modules" || base == ".git" || base == ".omc" {
			return nil
		}
		
		result, err := ValidateFile(path)
		if err != nil {
			result = &ValidationResult{
				Valid:    false,
				FilePath: path,
			}
			result.AddError("", "failed to validate", err.Error())
		}
		results = append(results, result)
		
		return nil
	})
	
	return results, err
}

// detectConfigType determines the config type from its structure.
func detectConfigType(data map[string]any) ConfigType {
	// Check for calibration config (has calibrations field with nested temperature data)
	// Detection is based on structure, not required fields - required fields are validated later
	if _, ok := data["calibrations"]; ok {
		// Has calibrations field - check for other calibration-specific fields
		if _, hasNumLevels := data["num_levels"]; hasNumLevels {
			return ConfigTypeCalibration
		}
		// Also detect if it has material_name (another strong indicator)
		if _, hasMaterial := data["material_name"]; hasMaterial {
			return ConfigTypeCalibration
		}
	}
	
	// Check for preisach state (has hysteron_states)
	if _, ok := data["hysteron_states"]; ok {
		return ConfigTypePreisach
	}
	
	// Check for array design (has config with array_rows/array_cols)
	if cfg, ok := data["config"].(map[string]any); ok {
		if _, hasRows := cfg["array_rows"]; hasRows {
			return ConfigTypeArrayDesign
		}
	}
	
	// Check for weight matrix (has weights array and rows/cols)
	if _, hasWeights := data["weights"]; hasWeights {
		if _, hasRows := data["rows"]; hasRows {
			return ConfigTypeWeightMatrix
		}
	}
	
	// Check for OpenLane config (has DESIGN_NAME and VERILOG_FILES)
	if _, ok := data["DESIGN_NAME"]; ok {
		if _, hasVerilog := data["VERILOG_FILES"]; hasVerilog {
			return ConfigTypeOpenLane
		}
	}
	
	return ConfigTypeUnknown
}

// Helper functions for type assertions with validation

func getInt(data map[string]any, key string) (int, bool) {
	v, ok := data[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}

func getFloat(data map[string]any, key string) (float64, bool) {
	v, ok := data[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func getString(data map[string]any, key string) (string, bool) {
	v, ok := data[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getFloatArray(data map[string]any, key string) ([]float64, bool) {
	v, ok := data[key]
	if !ok {
		return nil, false
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	result := make([]float64, len(arr))
	for i, item := range arr {
		switch n := item.(type) {
		case float64:
			result[i] = n
		case int:
			result[i] = float64(n)
		default:
			return nil, false
		}
	}
	return result, true
}

func getIntArray(data map[string]any, key string) ([]int, bool) {
	v, ok := data[key]
	if !ok {
		return nil, false
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	result := make([]int, len(arr))
	for i, item := range arr {
		switch n := item.(type) {
		case float64:
			result[i] = int(n)
		case int:
			result[i] = n
		default:
			return nil, false
		}
	}
	return result, true
}

func getMap(data map[string]any, key string) (map[string]any, bool) {
	v, ok := data[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

// validateTimestamp checks if a string is a valid ISO 8601 timestamp.
func validateTimestamp(s string) bool {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, format := range formats {
		if _, err := time.Parse(format, s); err == nil {
			return true
		}
	}
	return false
}

// validateBasicJSON performs basic validation on unknown config types.
func validateBasicJSON(data map[string]any, result *ValidationResult) {
	// Check for common issues
	if len(data) == 0 {
		result.AddWarning("", "empty config object", nil)
	}
}
