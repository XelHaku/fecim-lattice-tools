package configvalidator

import (
	"fmt"
	"regexp"
	"strings"
)

// OpenLaneConfig represents an OpenLane ASIC flow configuration.
type OpenLaneConfig struct {
	DesignName         string  `json:"DESIGN_NAME"`
	VerilogFiles       string  `json:"VERILOG_FILES"`
	ClockPeriod        float64 `json:"CLOCK_PERIOD"`
	ClockPort          string  `json:"CLOCK_PORT"`
	ClockNet           string  `json:"CLOCK_NET"`
	PDK                string  `json:"PDK"`
	STDCellLibrary     string  `json:"STD_CELL_LIBRARY"`
	ExtraLEFs          string  `json:"EXTRA_LEFS,omitempty"`
	ExtraGDSFiles      string  `json:"EXTRA_GDS_FILES,omitempty"`
	ExtraLibs          string  `json:"EXTRA_LIBS,omitempty"`
	VerilogBlackbox    string  `json:"VERILOG_FILES_BLACKBOX,omitempty"`
	SynthElaborateOnly int     `json:"SYNTH_ELABORATE_ONLY,omitempty"`
	FPSizing           string  `json:"FP_SIZING,omitempty"`
	DIEArea            string  `json:"DIE_AREA,omitempty"`
	DesignIsCore       int     `json:"DESIGN_IS_CORE,omitempty"`
}

// OpenLane config constraints
const (
	MinClockPeriod = 0.1   // ns (10 GHz)
	MaxClockPeriod = 10000 // ns (100 kHz)
)

// Valid PDKs
var validPDKs = map[string]bool{
	"sky130A":   true,
	"sky130B":   true,
	"gf180mcuC": true,
	"gf180mcuD": true,
	"asap7":     true,
}

// Valid standard cell libraries per PDK
var validStdCellLibraries = map[string][]string{
	"sky130A":   {"sky130_fd_sc_hd", "sky130_fd_sc_hs", "sky130_fd_sc_ms", "sky130_fd_sc_ls", "sky130_fd_sc_hdll"},
	"sky130B":   {"sky130_fd_sc_hd", "sky130_fd_sc_hs", "sky130_fd_sc_ms", "sky130_fd_sc_ls", "sky130_fd_sc_hdll"},
	"gf180mcuC": {"gf180mcu_fd_sc_mcu7t5v0", "gf180mcu_fd_sc_mcu9t5v0"},
	"gf180mcuD": {"gf180mcu_fd_sc_mcu7t5v0", "gf180mcu_fd_sc_mcu9t5v0"},
}

// Valid FP_SIZING values
var validFPSizing = map[string]bool{
	"absolute": true,
	"relative": true,
}

// Design name regex (valid Verilog identifier)
var verilogIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_$]*$`)

// DIE_AREA regex (format: "x0 y0 x1 y1")
var dieAreaRegex = regexp.MustCompile(`^\s*[\d.]+\s+[\d.]+\s+[\d.]+\s+[\d.]+\s*$`)

// validateOpenLaneConfig validates an OpenLane configuration.
func validateOpenLaneConfig(data map[string]any, result *ValidationResult) {
	// Validate DESIGN_NAME (required)
	designName, ok := getString(data, "DESIGN_NAME")
	if !ok {
		result.AddError("DESIGN_NAME", "required field missing or invalid type", data["DESIGN_NAME"])
	} else {
		if designName == "" {
			result.AddError("DESIGN_NAME", "must not be empty", designName)
		} else if !verilogIdentifierRegex.MatchString(designName) {
			result.AddError("DESIGN_NAME", "must be a valid Verilog identifier", designName)
		}
	}
	
	// Validate VERILOG_FILES (required)
	verilogFiles, ok := getString(data, "VERILOG_FILES")
	if !ok {
		result.AddError("VERILOG_FILES", "required field missing or invalid type", data["VERILOG_FILES"])
	} else if verilogFiles == "" {
		result.AddError("VERILOG_FILES", "must not be empty", verilogFiles)
	}
	
	// Validate CLOCK_PERIOD (required)
	clockPeriod, ok := getFloat(data, "CLOCK_PERIOD")
	if !ok {
		result.AddError("CLOCK_PERIOD", "required field missing or invalid type", data["CLOCK_PERIOD"])
	} else {
		if clockPeriod < MinClockPeriod {
			result.AddError("CLOCK_PERIOD", fmt.Sprintf("must be at least %.1f ns", MinClockPeriod), clockPeriod)
		}
		if clockPeriod > MaxClockPeriod {
			result.AddWarning("CLOCK_PERIOD", "unusually long clock period (slow design)", clockPeriod)
		}
	}
	
	// Validate CLOCK_PORT if present
	if clockPort, ok := getString(data, "CLOCK_PORT"); ok {
		if clockPort == "" {
			result.AddWarning("CLOCK_PORT", "empty clock port name", clockPort)
		}
	}
	
	// Validate PDK (required)
	pdk, ok := getString(data, "PDK")
	if !ok {
		result.AddError("PDK", "required field missing or invalid type", data["PDK"])
	} else {
		if !validPDKs[pdk] {
			result.AddWarning("PDK", "unknown PDK (expected: sky130A, sky130B, gf180mcuC, gf180mcuD, asap7)", pdk)
		}
	}
	
	// Validate STD_CELL_LIBRARY
	if stdCellLib, ok := getString(data, "STD_CELL_LIBRARY"); ok {
		if pdk != "" {
			validLibs := validStdCellLibraries[pdk]
			isValid := false
			for _, lib := range validLibs {
				if lib == stdCellLib {
					isValid = true
					break
				}
			}
			if !isValid && len(validLibs) > 0 {
				result.AddWarning("STD_CELL_LIBRARY", fmt.Sprintf("unexpected library for PDK %s (expected one of: %s)", pdk, strings.Join(validLibs, ", ")), stdCellLib)
			}
		}
	}
	
	// Validate FP_SIZING if present
	if fpSizing, ok := getString(data, "FP_SIZING"); ok {
		if !validFPSizing[fpSizing] {
			result.AddError("FP_SIZING", "must be 'absolute' or 'relative'", fpSizing)
		}
	}
	
	// Validate DIE_AREA if present
	if dieArea, ok := getString(data, "DIE_AREA"); ok {
		if !dieAreaRegex.MatchString(dieArea) {
			result.AddError("DIE_AREA", "must be in format 'x0 y0 x1 y1' (4 numbers)", dieArea)
		} else {
			// Parse and validate coordinates
			validateDIEArea(dieArea, result)
		}
	}
	
	// Validate boolean/integer flags
	validateOpenLaneFlags(data, result)
	
	// Validate file references
	validateOpenLaneFileRefs(data, result)
}

// validateDIEArea parses and validates DIE_AREA coordinates.
func validateDIEArea(dieArea string, result *ValidationResult) {
	var x0, y0, x1, y1 float64
	n, err := fmt.Sscanf(dieArea, "%f %f %f %f", &x0, &y0, &x1, &y1)
	if err != nil || n != 4 {
		return // Already caught by regex
	}
	
	// Validate coordinates make sense
	if x1 <= x0 {
		result.AddError("DIE_AREA", "x1 must be greater than x0", dieArea)
	}
	if y1 <= y0 {
		result.AddError("DIE_AREA", "y1 must be greater than y0", dieArea)
	}
	if x0 < 0 || y0 < 0 {
		result.AddError("DIE_AREA", "coordinates must be non-negative", dieArea)
	}
	
	// Warn about unusually small/large areas
	width := x1 - x0
	height := y1 - y0
	area := width * height
	
	if area < 1 { // Less than 1 µm²
		result.AddWarning("DIE_AREA", "unusually small die area", area)
	}
	if area > 1e8 { // Greater than 100 mm²
		result.AddWarning("DIE_AREA", "unusually large die area", area)
	}
}

// validateOpenLaneFlags validates integer flag fields.
func validateOpenLaneFlags(data map[string]any, result *ValidationResult) {
	// These are typically 0 or 1
	booleanFlags := []string{
		"SYNTH_ELABORATE_ONLY",
		"DESIGN_IS_CORE",
		"FP_PDN_ENABLE_RAILS",
		"RUN_CTS",
		"RUN_FILL_INSERTION",
		"QUIT_ON_MAGIC_DRC",
		"QUIT_ON_LVS_ERROR",
		"QUIT_ON_TR_DRC",
		"PL_SKIP_INITIAL_PLACEMENT",
	}
	
	for _, flag := range booleanFlags {
		if v, ok := getInt(data, flag); ok {
			if v != 0 && v != 1 {
				result.AddWarning(flag, "expected 0 or 1", v)
			}
		}
	}
}

// validateOpenLaneFileRefs validates file reference fields.
func validateOpenLaneFileRefs(data map[string]any, result *ValidationResult) {
	fileRefFields := []string{
		"EXTRA_LEFS",
		"EXTRA_GDS_FILES",
		"EXTRA_LIBS",
		"VERILOG_FILES_BLACKBOX",
		"PLACEMENT_CURRENT_DEF",
	}
	
	for _, field := range fileRefFields {
		if v, ok := getString(data, field); ok {
			// Check for common OpenLane file reference patterns
			if v != "" && !strings.HasPrefix(v, "dir::") && !strings.HasPrefix(v, "ref::") {
				// Could be an absolute or relative path - just check it's not empty
				if strings.ContainsAny(v, "<>|\"") {
					result.AddWarning(field, "contains invalid path characters", v)
				}
			}
		}
	}
}
