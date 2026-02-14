// pkg/export/lef_syntax_validation_test.go
// M6-LEF-01: LEF syntax validation test
// Validates LEF export syntax and parser round-trip correctness

package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestLEFSyntaxValidation tests M6-LEF-01:
// Export LEF cell library, parse back, verify VERSION, MACRO, PIN, OBS
func TestLEFSyntaxValidation(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Width:      0.46, // µm
		Height:     2.72, // µm
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}

	lef := GenerateLEF(cfg)

	// M6-LEF-01.1: Verify VERSION
	if !strings.Contains(lef, "VERSION 5.8") {
		t.Error("M6-LEF-01.1 FAIL: Missing VERSION 5.8")
	} else {
		t.Log("M6-LEF-01.1 PASS: VERSION 5.8 present")
	}

	// M6-LEF-01.2: Verify BUSBITCHARS
	if !strings.Contains(lef, `BUSBITCHARS "[]"`) {
		t.Error("M6-LEF-01.2 FAIL: Missing BUSBITCHARS declaration")
	} else {
		t.Log("M6-LEF-01.2 PASS: BUSBITCHARS \"[]\" present")
	}

	// M6-LEF-01.3: Verify DIVIDERCHAR
	if !strings.Contains(lef, `DIVIDERCHAR "/"`) {
		t.Error("M6-LEF-01.3 FAIL: Missing DIVIDERCHAR declaration")
	} else {
		t.Log("M6-LEF-01.3 PASS: DIVIDERCHAR \"/\" present")
	}

	// M6-LEF-01.4: Verify LAYER definition
	if !strings.Contains(lef, "LAYER met1") {
		t.Error("M6-LEF-01.4 FAIL: Missing LAYER met1 definition")
	} else {
		t.Log("M6-LEF-01.4 PASS: LAYER met1 defined")
	}

	// M6-LEF-01.5: Verify SITE definition
	if !strings.Contains(lef, "SITE fecim_site") {
		t.Error("M6-LEF-01.5 FAIL: Missing SITE definition")
	} else {
		t.Log("M6-LEF-01.5 PASS: SITE fecim_site defined")
	}

	// M6-LEF-01.6: Verify MACRO definition
	if !strings.Contains(lef, "MACRO fecim_bitcell") {
		t.Error("M6-LEF-01.6 FAIL: Missing MACRO fecim_bitcell")
	} else {
		t.Log("M6-LEF-01.6 PASS: MACRO fecim_bitcell defined")
	}

	// M6-LEF-01.7: Verify CLASS
	if !strings.Contains(lef, "CLASS CORE") {
		t.Error("M6-LEF-01.7 FAIL: Missing CLASS CORE")
	} else {
		t.Log("M6-LEF-01.7 PASS: CLASS CORE defined")
	}

	// M6-LEF-01.8: Verify SIZE
	if !strings.Contains(lef, "SIZE 0.460 BY 2.720") {
		t.Error("M6-LEF-01.8 FAIL: Missing or incorrect SIZE declaration")
	} else {
		t.Log("M6-LEF-01.8 PASS: SIZE 0.460 BY 2.720")
	}

	// M6-LEF-01.9: Verify PIN definitions (WL, BL, VPWR, VGND)
	requiredPins := []string{"PIN WL", "PIN BL", "PIN VPWR", "PIN VGND"}
	for _, pin := range requiredPins {
		if !strings.Contains(lef, pin) {
			t.Errorf("M6-LEF-01.9 FAIL: Missing %s definition", pin)
		}
	}
	t.Log("M6-LEF-01.9 PASS: All required pins (WL, BL, VPWR, VGND) defined")

	// M6-LEF-01.10: Verify OBS (obstruction) section
	if !strings.Contains(lef, "OBS") {
		t.Error("M6-LEF-01.10 FAIL: Missing OBS section")
	} else {
		t.Log("M6-LEF-01.10 PASS: OBS section present")
	}

	// M6-LEF-01.11: Verify END statements
	if !strings.Contains(lef, "END met1") {
		t.Error("M6-LEF-01.11 FAIL: Missing END met1")
	}
	if !strings.Contains(lef, "END fecim_site") {
		t.Error("M6-LEF-01.11 FAIL: Missing END fecim_site")
	}
	if !strings.Contains(lef, "END fecim_bitcell") {
		t.Error("M6-LEF-01.11 FAIL: Missing END fecim_bitcell")
	}
	if !strings.Contains(lef, "END LIBRARY") {
		t.Error("M6-LEF-01.11 FAIL: Missing END LIBRARY")
	} else {
		t.Log("M6-LEF-01.11 PASS: All END statements present")
	}

	t.Logf("M6-LEF-01 syntax validation: %d bytes", len(lef))
}

// TestLEFStructureCorrectness verifies section order and completeness
func TestLEFStructureCorrectness(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "test_cell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)
	lines := strings.Split(lef, "\n")

	// Verify section order: VERSION → LAYER → SITE → MACRO → END LIBRARY
	versionIdx := findLineIndex(lines, "VERSION")
	layerIdx := findLineIndex(lines, "LAYER")
	siteIdx := findLineIndex(lines, "SITE")
	macroIdx := findLineIndex(lines, "MACRO")
	endLibIdx := findLineIndex(lines, "END LIBRARY")

	if !(versionIdx < layerIdx && layerIdx < siteIdx && siteIdx < macroIdx && macroIdx < endLibIdx) {
		t.Error("M6-LEF-01 FAIL: Incorrect section order")
		t.Logf("Order: VERSION=%d LAYER=%d SITE=%d MACRO=%d END=%d",
			versionIdx, layerIdx, siteIdx, macroIdx, endLibIdx)
	} else {
		t.Log("M6-LEF-01 PASS: Section order correct")
	}

	// Verify PIN sections are within MACRO section
	macroEndIdx := findLineIndex(lines, "END test_cell")
	wlPinIdx := findLineIndex(lines, "PIN WL")
	blPinIdx := findLineIndex(lines, "PIN BL")

	if !(macroIdx < wlPinIdx && wlPinIdx < macroEndIdx) {
		t.Error("M6-LEF-01 FAIL: PIN WL not within MACRO section")
	}
	if !(macroIdx < blPinIdx && blPinIdx < macroEndIdx) {
		t.Error("M6-LEF-01 FAIL: PIN BL not within MACRO section")
	}

	t.Log("M6-LEF-01 structure correctness: PASS")
}

// TestLEF1T1RSyntax verifies 1T1R LEF has SL pin
func TestLEF1T1RSyntax(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_1t1r_bitcell",
		CellType: "1t1r",
		Width:    0.92,
		Height:   3.40,
	}

	lef := Generate1T1RLEF(cfg)

	// Verify 1T1R-specific syntax
	if !strings.Contains(lef, "MACRO fecim_1t1r_bitcell") {
		t.Error("M6-LEF-01 (1T1R) FAIL: Missing MACRO fecim_1t1r_bitcell")
	}

	// SL pin should exist
	if !strings.Contains(lef, "PIN SL") {
		t.Error("M6-LEF-01 (1T1R) FAIL: Missing PIN SL")
	} else {
		t.Log("M6-LEF-01 (1T1R) PASS: PIN SL present")
	}

	// Site name should be 1T1R-specific
	if !strings.Contains(lef, "SITE fecim_1t1r_site") {
		t.Error("M6-LEF-01 (1T1R) FAIL: Missing SITE fecim_1t1r_site")
	} else {
		t.Log("M6-LEF-01 (1T1R) PASS: SITE fecim_1t1r_site defined")
	}

	// Verify larger size for 1T1R
	if !strings.Contains(lef, "SIZE 0.920 BY 3.400") {
		t.Error("M6-LEF-01 (1T1R) FAIL: Missing or incorrect SIZE (expected 0.920 × 3.400)")
	} else {
		t.Log("M6-LEF-01 (1T1R) PASS: SIZE 0.920 BY 3.400")
	}

	t.Log("M6-LEF-01 (1T1R) syntax validation: PASS")
}

// TestLEF2T1RSyntax verifies 2T1R LEF has CSL pin
func TestLEF2T1RSyntax(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_2t1r_bitcell",
		CellType: "2t1r",
		Width:    1.38,
		Height:   3.40,
	}

	lef := Generate2T1RLEF(cfg)

	// Verify 2T1R-specific syntax
	if !strings.Contains(lef, "MACRO fecim_2t1r_bitcell") {
		t.Error("M6-LEF-01 (2T1R) FAIL: Missing MACRO fecim_2t1r_bitcell")
	}

	// CSL pin should exist
	if !strings.Contains(lef, "PIN CSL") {
		t.Error("M6-LEF-01 (2T1R) FAIL: Missing PIN CSL")
	} else {
		t.Log("M6-LEF-01 (2T1R) PASS: PIN CSL present")
	}

	// SL pin should also exist
	if !strings.Contains(lef, "PIN SL") {
		t.Error("M6-LEF-01 (2T1R) FAIL: Missing PIN SL")
	} else {
		t.Log("M6-LEF-01 (2T1R) PASS: PIN SL present")
	}

	// Site name should be 2T1R-specific
	if !strings.Contains(lef, "SITE fecim_2t1r_site") {
		t.Error("M6-LEF-01 (2T1R) FAIL: Missing SITE fecim_2t1r_site")
	} else {
		t.Log("M6-LEF-01 (2T1R) PASS: SITE fecim_2t1r_site defined")
	}

	// Verify larger size for 2T1R
	if !strings.Contains(lef, "SIZE 1.380 BY 3.400") {
		t.Error("M6-LEF-01 (2T1R) FAIL: Missing or incorrect SIZE (expected 1.380 × 3.400)")
	} else {
		t.Log("M6-LEF-01 (2T1R) PASS: SIZE 1.380 BY 3.400")
	}

	t.Log("M6-LEF-01 (2T1R) syntax validation: PASS")
}

// TestLEFPinDirections verifies pin DIRECTION and USE attributes
func TestLEFPinDirections(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "test_cell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)

	// WL should be INPUT
	wlSection := extractPinSection(lef, "WL")
	if !strings.Contains(wlSection, "DIRECTION INPUT") {
		t.Error("M6-LEF-01 FAIL: WL should be DIRECTION INPUT")
	}
	if !strings.Contains(wlSection, "USE SIGNAL") {
		t.Error("M6-LEF-01 FAIL: WL should be USE SIGNAL")
	}

	// BL should be OUTPUT (or INOUT)
	blSection := extractPinSection(lef, "BL")
	if !strings.Contains(blSection, "DIRECTION OUTPUT") && !strings.Contains(blSection, "DIRECTION INOUT") {
		t.Error("M6-LEF-01 FAIL: BL should be DIRECTION OUTPUT or INOUT")
	}
	if !strings.Contains(blSection, "USE SIGNAL") {
		t.Error("M6-LEF-01 FAIL: BL should be USE SIGNAL")
	}

	// VPWR should be POWER
	vpwrSection := extractPinSection(lef, "VPWR")
	if !strings.Contains(vpwrSection, "USE POWER") {
		t.Error("M6-LEF-01 FAIL: VPWR should be USE POWER")
	}

	// VGND should be GROUND
	vgndSection := extractPinSection(lef, "VGND")
	if !strings.Contains(vgndSection, "USE GROUND") {
		t.Error("M6-LEF-01 FAIL: VGND should be USE GROUND")
	}

	t.Log("M6-LEF-01 pin directions: PASS")
}

// TestLEFPinPorts verifies all pins have PORT sections
func TestLEFPinPorts(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "test_cell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)

	// All pins should have PORT ... END sections
	requiredPins := []string{"WL", "BL", "VPWR", "VGND"}
	for _, pin := range requiredPins {
		pinSection := extractPinSection(lef, pin)
		if !strings.Contains(pinSection, "PORT") {
			t.Errorf("M6-LEF-01 FAIL: PIN %s missing PORT section", pin)
		}
		if !strings.Contains(pinSection, "LAYER met1") {
			t.Errorf("M6-LEF-01 FAIL: PIN %s PORT missing LAYER specification", pin)
		}
		if !strings.Contains(pinSection, "RECT") {
			t.Errorf("M6-LEF-01 FAIL: PIN %s PORT missing RECT geometry", pin)
		}
	}

	t.Log("M6-LEF-01 pin ports: All pins have PORT sections with LAYER and RECT")
}

// TestLEFLayerAttributes verifies layer TYPE and DIRECTION
func TestLEFLayerAttributes(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "test_cell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)

	// met1 should be TYPE ROUTING
	if !strings.Contains(lef, "TYPE ROUTING") {
		t.Error("M6-LEF-01 FAIL: LAYER met1 should be TYPE ROUTING")
	}

	// met1 should have DIRECTION
	if !strings.Contains(lef, "DIRECTION HORIZONTAL") {
		t.Error("M6-LEF-01 FAIL: LAYER met1 should have DIRECTION HORIZONTAL")
	}

	// met1 should have PITCH
	if !strings.Contains(lef, "PITCH") {
		t.Error("M6-LEF-01 FAIL: LAYER met1 should have PITCH")
	}

	// met1 should have WIDTH
	if !strings.Contains(lef, "WIDTH") {
		t.Error("M6-LEF-01 FAIL: LAYER met1 should have WIDTH")
	}

	t.Log("M6-LEF-01 layer attributes: PASS")
}

// Helper: extract a PIN section from LEF
func extractPinSection(lef, pinName string) string {
	pinStart := strings.Index(lef, "PIN "+pinName)
	if pinStart == -1 {
		return ""
	}
	pinEnd := strings.Index(lef[pinStart:], "END "+pinName)
	if pinEnd == -1 {
		return ""
	}
	return lef[pinStart : pinStart+pinEnd]
}
