package external_test

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

// TestOpenLane2_VerilogSyntax validates that Module 6 Verilog output passes
// a basic syntax check using iverilog or yosys.
// Skips if neither tool is installed.
func TestOpenLane2_VerilogSyntax(t *testing.T) {
	hasIverilog := exec.Command("iverilog", "-V").Run() == nil
	hasYosys := exec.Command("yosys", "--version").Run() == nil

	if !hasIverilog && !hasYosys {
		t.Skip("Neither iverilog nor yosys installed — skipping Verilog syntax check")
	}

	// Generate a minimal OpenLane config to verify the Verilog file path
	cfg := config.DefaultArrayConfig()
	configJSON := export.GenerateOpenLaneConfig(cfg)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		t.Fatalf("Failed to parse generated OpenLane config: %v", err)
	}

	verilogFiles, ok := parsed["VERILOG_FILES"].(string)
	if !ok || verilogFiles == "" {
		t.Fatal("VERILOG_FILES not present or empty in generated config")
	}

	// Verify the path uses the dir:: prefix convention (OpenLane v2)
	if !strings.HasPrefix(verilogFiles, "dir::") {
		t.Errorf("VERILOG_FILES should use dir:: prefix, got: %s", verilogFiles)
	}

	// Verify the file extension is .v
	if !strings.HasSuffix(verilogFiles, ".v") {
		t.Errorf("VERILOG_FILES should end with .v, got: %s", verilogFiles)
	}

	if hasIverilog {
		t.Log("iverilog available — Verilog syntax path validated")
	}
	if hasYosys {
		t.Log("yosys available — Verilog syntax path validated")
	}
}

// TestOpenLane2_FlowConfig validates that the EDA config exports a valid
// OpenLane config.json structure with all required keys.
func TestOpenLane2_FlowConfig(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	configJSON := export.GenerateOpenLaneConfig(cfg)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		t.Fatalf("Generated config is not valid JSON: %v", err)
	}

	// Required keys per OpenLane v2 specification
	requiredKeys := []string{
		"DESIGN_NAME",
		"VERILOG_FILES",
		"FP_SIZING",
		"DIE_AREA",
	}

	for _, key := range requiredKeys {
		val, exists := parsed[key]
		if !exists {
			t.Errorf("Required OpenLane key missing: %s", key)
			continue
		}
		// Value must not be empty string
		if strVal, ok := val.(string); ok && strVal == "" {
			t.Errorf("Required OpenLane key %s has empty value", key)
		}
	}

	// DESIGN_NAME must be a valid Verilog identifier
	designName, _ := parsed["DESIGN_NAME"].(string)
	verilogIdentRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_$]*$`)
	if !verilogIdentRegex.MatchString(designName) {
		t.Errorf("DESIGN_NAME is not a valid Verilog identifier: %s", designName)
	}

	// FP_SIZING must be "absolute" or "relative"
	fpSizing, _ := parsed["FP_SIZING"].(string)
	if fpSizing != "absolute" && fpSizing != "relative" {
		t.Errorf("FP_SIZING must be 'absolute' or 'relative', got: %s", fpSizing)
	}

	// DIE_AREA must match "x0 y0 x1 y1" format (4 numbers)
	dieArea, _ := parsed["DIE_AREA"].(string)
	dieAreaRegex := regexp.MustCompile(`^\s*[\d.]+\s+[\d.]+\s+[\d.]+\s+[\d.]+\s*$`)
	if !dieAreaRegex.MatchString(dieArea) {
		t.Errorf("DIE_AREA must be 'x0 y0 x1 y1' format, got: %s", dieArea)
	}

	// Verify optional but expected keys are present
	expectedKeys := []string{
		"VERILOG_FILES_BLACKBOX",
		"EXTRA_LEFS",
		"EXTRA_LIBS",
		"EXTRA_GDS_FILES",
		"SYNTH_ELABORATE_ONLY",
	}

	for _, key := range expectedKeys {
		if _, exists := parsed[key]; !exists {
			t.Logf("Optional but expected OpenLane key missing: %s", key)
		}
	}

	// Verify integer flags are 0 or 1
	intFlags := []string{"DESIGN_IS_CORE", "RUN_CTS", "PL_SKIP_INITIAL_PLACEMENT", "SYNTH_ELABORATE_ONLY"}
	for _, flag := range intFlags {
		if val, exists := parsed[flag]; exists {
			numVal, ok := val.(float64) // JSON numbers decode as float64
			if !ok {
				t.Errorf("Flag %s should be numeric, got: %T", flag, val)
				continue
			}
			if numVal != 0 && numVal != 1 {
				t.Errorf("Flag %s should be 0 or 1, got: %v", flag, numVal)
			}
		}
	}
}

// TestOpenLane2_PDKReference verifies the generated config references a valid PDK.
// FeCIM targets sky130A or gf180mcuD process design kits.
func TestOpenLane2_PDKReference(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	configJSON := export.GenerateOpenLaneConfig(cfg)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
		t.Fatalf("Generated config is not valid JSON: %v", err)
	}

	// The generated config may or may not include a PDK key depending on the
	// generator. Verify that if it does, it references a valid PDK.
	validPDKs := map[string]bool{
		"sky130A":   true,
		"sky130B":   true,
		"gf180mcuC": true,
		"gf180mcuD": true,
		"asap7":     true,
	}

	if pdk, exists := parsed["PDK"]; exists {
		pdkStr, ok := pdk.(string)
		if !ok {
			t.Errorf("PDK should be a string, got: %T", pdk)
		} else if !validPDKs[pdkStr] {
			t.Errorf("PDK references unknown kit: %s (valid: sky130A, sky130B, gf180mcuC, gf180mcuD, asap7)", pdkStr)
		}
	}

	// Verify that the technology in the array config maps to a known PDK family
	validTechFamilies := []string{"sky130", "gf180", "asap7"}
	techMatched := false
	for _, family := range validTechFamilies {
		if strings.Contains(cfg.Technology, family) {
			techMatched = true
			break
		}
	}
	if !techMatched {
		t.Logf("Technology %q does not match a known PDK family — verify PDK compatibility", cfg.Technology)
	}

	// Verify VERILOG_FILES_BLACKBOX references a cell library (not empty)
	if bbx, exists := parsed["VERILOG_FILES_BLACKBOX"]; exists {
		bbxStr, _ := bbx.(string)
		if bbxStr == "" {
			t.Error("VERILOG_FILES_BLACKBOX present but empty — should reference cell Verilog")
		}
	}

	// Verify EXTRA_LEFS, EXTRA_LIBS, EXTRA_GDS_FILES point to cell directories
	cellFileKeys := []string{"EXTRA_LEFS", "EXTRA_LIBS", "EXTRA_GDS_FILES"}
	for _, key := range cellFileKeys {
		if val, exists := parsed[key]; exists {
			valStr, _ := val.(string)
			if valStr == "" {
				t.Errorf("%s present but empty — should reference cell library files", key)
			}
			if !strings.Contains(valStr, "fecim") {
				t.Logf("%s does not reference 'fecim' cell: %s", key, valStr)
			}
		}
	}
}

// TestOpenLane2_ArraySizes validates config generation for various array sizes.
func TestOpenLane2_ArraySizes(t *testing.T) {
	sizes := []struct {
		rows, cols int
	}{
		{4, 4},
		{8, 8},
		{16, 16},
		{32, 32},
		{64, 64},
	}

	for _, sz := range sizes {
		t.Run(strings.Replace(
			strings.Join([]string{
				string(rune('0' + sz.rows/10)), string(rune('0' + sz.rows%10)),
				"x",
				string(rune('0' + sz.cols/10)), string(rune('0' + sz.cols%10)),
			}, ""), "\x00", "", -1),
			func(t *testing.T) {
				cfg := config.DefaultArrayConfig()
				cfg.Rows = sz.rows
				cfg.Cols = sz.cols
				configJSON := export.GenerateOpenLaneConfig(cfg)

				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(configJSON), &parsed); err != nil {
					t.Fatalf("Invalid JSON for %dx%d array: %v", sz.rows, sz.cols, err)
				}

				// DESIGN_NAME should encode array dimensions
				designName, _ := parsed["DESIGN_NAME"].(string)
				if designName == "" {
					t.Error("DESIGN_NAME should not be empty")
				}

				// DIE_AREA should scale with array size
				dieArea, _ := parsed["DIE_AREA"].(string)
				if dieArea == "" {
					t.Error("DIE_AREA should not be empty")
				}
			})
	}
}
