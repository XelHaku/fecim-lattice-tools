package configvalidator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateCalibrationConfig tests calibration config validation.
func TestValidateCalibrationConfig(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectValid bool
		expectErrs  []string // Substrings to check for in errors
		expectWarns []string // Substrings to check for in warnings
	}{
		{
			name: "valid calibration v2",
			json: `{
				"version": 2,
				"material_name": "Test HZO",
				"num_levels": 4,
				"calibrations": {
					"300": {
						"temperature_k": 300,
						"calibration_up": [0, 1e7, 2e7, 3e7],
						"calibration_down": [-3e7, -2e7, -1e7, 0],
						"calib_up_low": [0, 0, 1e7, 2e7],
						"calib_up_high": [3e7, 1e7, 2e7, 3e7],
						"calib_down_low": [-3e7, -3e7, -2e7, -3e7],
						"calib_down_high": [-2e7, -1e7, 0, 0],
						"last_error_up": [0, 0, 0, 0],
						"last_error_down": [0, 0, 0, 0],
						"relax_comp_up": [0, 0.02, 0.04, 0],
						"relax_comp_down": [0, 0.02, 0.04, 0]
					}
				},
				"saved_at": "2026-01-29T10:00:00-06:00"
			}`,
			expectValid: true,
		},
		{
			name: "missing version",
			json: `{
				"material_name": "Test HZO",
				"num_levels": 4,
				"calibrations": {}
			}`,
			expectValid: false,
			expectErrs:  []string{"version", "required"},
		},
		{
			name: "invalid version",
			json: `{
				"version": 99,
				"material_name": "Test HZO",
				"num_levels": 4,
				"calibrations": {}
			}`,
			expectValid: false,
			expectErrs:  []string{"version", "between"},
		},
		{
			name: "empty material name",
			json: `{
				"version": 2,
				"material_name": "",
				"num_levels": 4,
				"calibrations": {"300": {"temperature_k": 300, "calibration_up": [0,1,2,3], "calibration_down": [-3,-2,-1,0], "calib_up_low": [0,0,0,0], "calib_up_high": [3,3,3,3], "calib_down_low": [-3,-3,-3,-3], "calib_down_high": [0,0,0,0], "last_error_up": [0,0,0,0], "last_error_down": [0,0,0,0], "relax_comp_up": [0,0,0,0], "relax_comp_down": [0,0,0,0]}}
			}`,
			expectValid: false,
			expectErrs:  []string{"material_name", "empty"},
		},
		{
			name: "num_levels too small",
			json: `{
				"version": 2,
				"material_name": "Test",
				"num_levels": 1,
				"calibrations": {}
			}`,
			expectValid: false,
			expectErrs:  []string{"num_levels", "at least"},
		},
		{
			name: "array length mismatch",
			json: `{
				"version": 2,
				"material_name": "Test",
				"num_levels": 4,
				"calibrations": {
					"300": {
						"temperature_k": 300,
						"calibration_up": [0, 1, 2],
						"calibration_down": [-3, -2, -1, 0],
						"calib_up_low": [0,0,0,0],
						"calib_up_high": [4,4,4,4],
						"calib_down_low": [-4,-4,-4,-4],
						"calib_down_high": [0,0,0,0],
						"last_error_up": [0,0,0,0],
						"last_error_down": [0,0,0,0],
						"relax_comp_up": [0,0,0,0],
						"relax_comp_down": [0,0,0,0]
					}
				}
			}`,
			expectValid: false,
			expectErrs:  []string{"calibration_up", "length"},
		},
		{
			name: "relax_comp out of range",
			json: `{
				"version": 2,
				"material_name": "Test",
				"num_levels": 2,
				"calibrations": {
					"300": {
						"temperature_k": 300,
						"calibration_up": [0, 1e7],
						"calibration_down": [-1e7, 0],
						"calib_up_low": [0,0],
						"calib_up_high": [2e7,2e7],
						"calib_down_low": [-2e7,-2e7],
						"calib_down_high": [0,0],
						"last_error_up": [0,0],
						"last_error_down": [0,0],
						"relax_comp_up": [0, 1.5],
						"relax_comp_down": [0, 0]
					}
				}
			}`,
			expectValid: false,
			expectErrs:  []string{"relax_comp_up", "between"},
		},
		{
			name: "temperature out of range",
			json: `{
				"version": 2,
				"material_name": "Test",
				"num_levels": 2,
				"calibrations": {
					"0": {
						"temperature_k": 0,
						"calibration_up": [0, 1e7],
						"calibration_down": [-1e7, 0],
						"calib_up_low": [0,0],
						"calib_up_high": [2e7,2e7],
						"calib_down_low": [-2e7,-2e7],
						"calib_down_high": [0,0],
						"last_error_up": [0,0],
						"last_error_down": [0,0],
						"relax_comp_up": [0,0],
						"relax_comp_down": [0,0]
					}
				}
			}`,
			expectValid: false,
			expectErrs:  []string{"temperature_k", "between"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON([]byte(tt.json))
			
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v\nErrors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			for _, errStr := range tt.expectErrs {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Error(), errStr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, but not found. Errors: %v", errStr, result.Errors)
				}
			}
			
			for _, warnStr := range tt.expectWarns {
				found := false
				for _, warn := range result.Warnings {
					if strings.Contains(warn.Error(), warnStr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning containing %q, but not found. Warnings: %v", warnStr, result.Warnings)
				}
			}
		})
	}
}

// TestValidatePreisachConfig tests Preisach state config validation.
func TestValidatePreisachConfig(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectValid bool
		expectErrs  []string
	}{
		{
			name: "valid preisach config",
			json: `{
				"version": 1,
				"material": "HZO",
				"temperature_k": 300,
				"grid_size": 3,
				"distribution_type": "gaussian",
				"hysteron_states": [1, 1, 1, -1, 1, 1],
				"alpha_mean": 1.2e8,
				"alpha_sigma": 7.8e7,
				"beta_mean": -1.2e8,
				"beta_sigma": 7.8e7
			}`,
			expectValid: true,
		},
		{
			name: "invalid distribution type",
			json: `{
				"version": 1,
				"material": "HZO",
				"temperature_k": 300,
				"grid_size": 3,
				"distribution_type": "invalid_type",
				"hysteron_states": [1, 1, 1, -1, 1, 1],
				"alpha_mean": 1.2e8,
				"alpha_sigma": 7.8e7,
				"beta_mean": -1.2e8,
				"beta_sigma": 7.8e7
			}`,
			expectValid: false,
			expectErrs:  []string{"distribution_type", "one of"},
		},
		{
			name: "invalid hysteron state",
			json: `{
				"version": 1,
				"material": "HZO",
				"temperature_k": 300,
				"grid_size": 2,
				"distribution_type": "gaussian",
				"hysteron_states": [1, 0, 1],
				"alpha_mean": 1.2e8,
				"alpha_sigma": 7.8e7,
				"beta_mean": -1.2e8,
				"beta_sigma": 7.8e7
			}`,
			expectValid: false,
			expectErrs:  []string{"hysteron_states", "must be -1 or 1"},
		},
		{
			name: "negative sigma",
			json: `{
				"version": 1,
				"material": "HZO",
				"temperature_k": 300,
				"grid_size": 2,
				"distribution_type": "gaussian",
				"hysteron_states": [1, -1, 1],
				"alpha_mean": 1.2e8,
				"alpha_sigma": -1,
				"beta_mean": -1.2e8,
				"beta_sigma": 7.8e7
			}`,
			expectValid: false,
			expectErrs:  []string{"alpha_sigma", "positive"},
		},
		{
			name: "correlation out of range",
			json: `{
				"version": 1,
				"material": "HZO",
				"temperature_k": 300,
				"grid_size": 2,
				"distribution_type": "gaussian",
				"hysteron_states": [1, -1, 1],
				"alpha_mean": 1.2e8,
				"alpha_sigma": 7.8e7,
				"beta_mean": -1.2e8,
				"beta_sigma": 7.8e7,
				"correlation": 1.5
			}`,
			expectValid: false,
			expectErrs:  []string{"correlation", "between"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON([]byte(tt.json))
			
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v\nErrors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			for _, errStr := range tt.expectErrs {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Error(), errStr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, but not found. Errors: %v", errStr, result.Errors)
				}
			}
		})
	}
}

// TestValidateWeightMatrixConfig tests weight matrix config validation.
func TestValidateWeightMatrixConfig(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectValid bool
		expectErrs  []string
	}{
		{
			name: "valid weight matrix",
			json: `{
				"name": "test_weights",
				"rows": 2,
				"cols": 3,
				"weights": [
					[0.1, -0.2, 0.3],
					[-0.1, 0.2, -0.3]
				]
			}`,
			expectValid: true,
		},
		{
			name: "row count mismatch",
			json: `{
				"name": "test_weights",
				"rows": 3,
				"cols": 2,
				"weights": [
					[0.1, -0.2],
					[-0.1, 0.2]
				]
			}`,
			expectValid: false,
			expectErrs:  []string{"weights", "row count"},
		},
		{
			name: "column count mismatch",
			json: `{
				"name": "test_weights",
				"rows": 2,
				"cols": 3,
				"weights": [
					[0.1, -0.2],
					[-0.1, 0.2, 0.3]
				]
			}`,
			expectValid: false,
			expectErrs:  []string{"weights[0]", "column count"},
		},
		{
			name: "empty name",
			json: `{
				"name": "",
				"rows": 2,
				"cols": 2,
				"weights": [[0.1, 0.2], [0.3, 0.4]]
			}`,
			expectValid: false,
			expectErrs:  []string{"name", "empty"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON([]byte(tt.json))
			
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v\nErrors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			for _, errStr := range tt.expectErrs {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Error(), errStr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, but not found. Errors: %v", errStr, result.Errors)
				}
			}
		})
	}
}

// TestValidateOpenLaneConfig tests OpenLane config validation.
func TestValidateOpenLaneConfig(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectValid bool
		expectErrs  []string
	}{
		{
			name: "valid openlane config",
			json: `{
				"DESIGN_NAME": "my_design",
				"VERILOG_FILES": "dir::src/design.v",
				"CLOCK_PERIOD": 10,
				"CLOCK_PORT": "clk",
				"PDK": "sky130A",
				"STD_CELL_LIBRARY": "sky130_fd_sc_hd"
			}`,
			expectValid: true,
		},
		{
			name: "invalid design name",
			json: `{
				"DESIGN_NAME": "123invalid",
				"VERILOG_FILES": "design.v",
				"CLOCK_PERIOD": 10,
				"PDK": "sky130A"
			}`,
			expectValid: false,
			expectErrs:  []string{"DESIGN_NAME", "Verilog identifier"},
		},
		{
			name: "clock period too small",
			json: `{
				"DESIGN_NAME": "design",
				"VERILOG_FILES": "design.v",
				"CLOCK_PERIOD": 0.01,
				"PDK": "sky130A"
			}`,
			expectValid: false,
			expectErrs:  []string{"CLOCK_PERIOD", "at least"},
		},
		{
			name: "invalid FP_SIZING",
			json: `{
				"DESIGN_NAME": "design",
				"VERILOG_FILES": "design.v",
				"CLOCK_PERIOD": 10,
				"PDK": "sky130A",
				"FP_SIZING": "invalid"
			}`,
			expectValid: false,
			expectErrs:  []string{"FP_SIZING", "absolute"},
		},
		{
			name: "invalid DIE_AREA",
			json: `{
				"DESIGN_NAME": "design",
				"VERILOG_FILES": "design.v",
				"CLOCK_PERIOD": 10,
				"PDK": "sky130A",
				"DIE_AREA": "0 0 0 100"
			}`,
			expectValid: false,
			expectErrs:  []string{"DIE_AREA", "x1 must be greater"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON([]byte(tt.json))
			
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v\nErrors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			for _, errStr := range tt.expectErrs {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Error(), errStr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, but not found. Errors: %v", errStr, result.Errors)
				}
			}
		})
	}
}

// TestValidateArrayDesignConfig tests array design config validation.
func TestValidateArrayDesignConfig(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectValid bool
		expectErrs  []string
	}{
		{
			name: "valid array design",
			json: `{
				"config": {
					"name": "test_array",
					"array_rows": 2,
					"array_cols": 2,
					"technology": "SKY130",
					"levels": 8,
					"g_min": 10,
					"g_max": 100
				},
				"cells": [
					{"row": 0, "col": 0, "level": 0, "conductance": 10},
					{"row": 0, "col": 1, "level": 4, "conductance": 50},
					{"row": 1, "col": 0, "level": 7, "conductance": 100},
					{"row": 1, "col": 1, "level": 2, "conductance": 30}
				]
			}`,
			expectValid: true,
		},
		{
			name: "cells count mismatch",
			json: `{
				"config": {
					"name": "test_array",
					"array_rows": 2,
					"array_cols": 2,
					"technology": "SKY130",
					"levels": 8
				},
				"cells": [
					{"row": 0, "col": 0, "level": 0}
				]
			}`,
			expectValid: false,
			expectErrs:  []string{"cells", "length"},
		},
		{
			name: "duplicate cell position",
			json: `{
				"config": {
					"name": "test_array",
					"array_rows": 2,
					"array_cols": 2,
					"technology": "SKY130",
					"levels": 8
				},
				"cells": [
					{"row": 0, "col": 0, "level": 0},
					{"row": 0, "col": 0, "level": 1},
					{"row": 1, "col": 0, "level": 2},
					{"row": 1, "col": 1, "level": 3}
				]
			}`,
			expectValid: false,
			expectErrs:  []string{"duplicate cell"},
		},
		{
			name: "level out of range",
			json: `{
				"config": {
					"name": "test_array",
					"array_rows": 1,
					"array_cols": 1,
					"technology": "SKY130",
					"levels": 8
				},
				"cells": [
					{"row": 0, "col": 0, "level": 10}
				]
			}`,
			expectValid: false,
			expectErrs:  []string{"level", "between"},
		},
		{
			name: "g_max not greater than g_min",
			json: `{
				"config": {
					"name": "test_array",
					"array_rows": 1,
					"array_cols": 1,
					"technology": "SKY130",
					"levels": 8,
					"g_min": 100,
					"g_max": 50
				},
				"cells": [{"row": 0, "col": 0, "level": 0}]
			}`,
			expectValid: false,
			expectErrs:  []string{"g_max", "greater than g_min"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJSON([]byte(tt.json))
			
			if result.Valid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v\nErrors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			for _, errStr := range tt.expectErrs {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Error(), errStr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, but not found. Errors: %v", errStr, result.Errors)
				}
			}
		})
	}
}

// TestDetectConfigType tests automatic config type detection.
func TestDetectConfigType(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		expectType ConfigType
	}{
		{
			name:       "calibration config",
			json:       `{"version": 2, "num_levels": 30, "calibrations": {"300": {}}}`,
			expectType: ConfigTypeCalibration,
		},
		{
			name:       "preisach config",
			json:       `{"hysteron_states": [1, -1, 1]}`,
			expectType: ConfigTypePreisach,
		},
		{
			name:       "weight matrix",
			json:       `{"rows": 8, "cols": 8, "weights": []}`,
			expectType: ConfigTypeWeightMatrix,
		},
		{
			name:       "openlane config",
			json:       `{"DESIGN_NAME": "test", "VERILOG_FILES": "test.v"}`,
			expectType: ConfigTypeOpenLane,
		},
		{
			name:       "array design",
			json:       `{"config": {"array_rows": 8, "array_cols": 8}}`,
			expectType: ConfigTypeArrayDesign,
		},
		{
			name:       "unknown config",
			json:       `{"random": "data"}`,
			expectType: ConfigTypeUnknown,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			if err := json.Unmarshal([]byte(tt.json), &data); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			
			configType := detectConfigType(data)
			if configType != tt.expectType {
				t.Errorf("expected config type %s, got %s", tt.expectType, configType)
			}
		})
	}
}

// TestValidateInvalidJSON tests handling of malformed JSON.
func TestValidateInvalidJSON(t *testing.T) {
	// These are truly invalid JSON
	invalidJSONs := []string{
		`{invalid json}`,
		`{"unclosed": "string`,
		``,          // Empty
	}
	
	for _, js := range invalidJSONs {
		result := ValidateJSON([]byte(js))
		if result.Valid {
			t.Errorf("expected invalid JSON to fail validation: %s", js)
		}
	}
	
	// JSON array is valid JSON but not a config object - should be valid with warning
	arrayResult := ValidateJSON([]byte(`[1, 2, 3]`))
	if !arrayResult.Valid {
		t.Error("JSON array should be valid (but with warnings)")
	}
	if len(arrayResult.Warnings) == 0 {
		t.Error("JSON array should produce a warning")
	}
	if arrayResult.ConfigType != "json_array" {
		t.Errorf("expected config type 'json_array', got %s", arrayResult.ConfigType)
	}
}

// TestValidationResultString tests the String() method.
func TestValidationResultString(t *testing.T) {
	result := &ValidationResult{
		Valid:      false,
		FilePath:   "/test/config.json",
		ConfigType: string(ConfigTypeCalibration),
	}
	result.AddError("version", "missing required field", nil)
	result.AddWarning("saved_at", "invalid timestamp", "bad-date")
	
	str := result.String()
	
	if !strings.Contains(str, "INVALID") {
		t.Error("expected 'INVALID' in result string")
	}
	if !strings.Contains(str, "version") {
		t.Error("expected error field name in result string")
	}
	if !strings.Contains(str, "saved_at") {
		t.Error("expected warning field name in result string")
	}
}

// TestValidateRealConfigs tests validation against real config files in the repo.
func TestValidateRealConfigs(t *testing.T) {
	// Test calibration configs
	calibrationDir := "../../data/calibrations"
	if files, err := os.ReadDir(calibrationDir); err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
				t.Run("calibration/"+f.Name(), func(t *testing.T) {
					path := filepath.Join(calibrationDir, f.Name())
					result, err := ValidateFile(path)
					if err != nil {
						t.Fatalf("failed to validate file: %v", err)
					}
					
					if result.ConfigType != string(ConfigTypeCalibration) {
						t.Errorf("expected calibration type, got %s", result.ConfigType)
					}
					
					// Log any errors/warnings for debugging
					if !result.Valid {
						t.Logf("Validation issues in %s:\n%s", f.Name(), result.String())
					}
				})
			}
		}
	}
	
	// Test preisach state configs
	preisachDir := "../../data/preisach_states"
	if files, err := os.ReadDir(preisachDir); err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
				t.Run("preisach/"+f.Name(), func(t *testing.T) {
					path := filepath.Join(preisachDir, f.Name())
					result, err := ValidateFile(path)
					if err != nil {
						t.Fatalf("failed to validate file: %v", err)
					}
					
					if result.ConfigType != string(ConfigTypePreisach) {
						t.Errorf("expected preisach type, got %s", result.ConfigType)
					}
					
					if !result.Valid {
						t.Logf("Validation issues in %s:\n%s", f.Name(), result.String())
					}
				})
			}
		}
	}
	
	// Test weight matrix configs
	weightsDir := "../../module6-eda/data"
	if files, err := os.ReadDir(weightsDir); err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") && strings.Contains(f.Name(), "weight") {
				t.Run("weights/"+f.Name(), func(t *testing.T) {
					path := filepath.Join(weightsDir, f.Name())
					result, err := ValidateFile(path)
					if err != nil {
						t.Fatalf("failed to validate file: %v", err)
					}
					
					if !result.Valid {
						t.Logf("Validation issues in %s:\n%s", f.Name(), result.String())
					}
				})
			}
		}
	}
}

// BenchmarkValidateCalibration benchmarks calibration validation.
func BenchmarkValidateCalibration(b *testing.B) {
	calibJSON := `{
		"version": 2,
		"material_name": "Benchmark HZO",
		"num_levels": 32,
		"calibrations": {
			"300": {
				"temperature_k": 300,
				"calibration_up": [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31],
				"calibration_down": [-31,-30,-29,-28,-27,-26,-25,-24,-23,-22,-21,-20,-19,-18,-17,-16,-15,-14,-13,-12,-11,-10,-9,-8,-7,-6,-5,-4,-3,-2,-1,0],
				"calib_up_low": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
				"calib_up_high": [31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31,31],
				"calib_down_low": [-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31,-31],
				"calib_down_high": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
				"last_error_up": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
				"last_error_down": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
				"relax_comp_up": [0,0.01,0.02,0.03,0.04,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.04,0.03,0.02,0.01,0.01,0.01,0.01,0],
				"relax_comp_down": [0,0.01,0.02,0.03,0.04,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.05,0.04,0.03,0.02,0.01,0.01,0.01,0.01,0]
			}
		}
	}`
	
	data := []byte(calibJSON)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ValidateJSON(data)
	}
}
