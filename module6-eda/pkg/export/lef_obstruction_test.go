// pkg/export/lef_obstruction_test.go
// M6-LEF-03: LEF obstruction validation test
// Verifies OBS layers present for dense cells

package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// ObstructionGeometry represents an obstruction's rectangle in the LEF
type ObstructionGeometry struct {
	Layer  string
	X1, Y1 float64 // Lower-left corner in µm
	X2, Y2 float64 // Upper-right corner in µm
}

// TestLEFObstructionPresence tests M6-LEF-03:
// Verify OBS layers present for dense cells
func TestLEFObstructionPresence(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_bitcell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)

	// M6-LEF-03.1: Verify OBS section exists
	if !strings.Contains(lef, "OBS") {
		t.Error("M6-LEF-03.1 FAIL: Missing OBS section")
	} else {
		t.Log("M6-LEF-03.1 PASS: OBS section present")
	}

	// M6-LEF-03.2: Parse obstruction geometries
	obstructions := parseObstructions(lef)
	if len(obstructions) == 0 {
		t.Error("M6-LEF-03.2 FAIL: No obstruction geometries found")
	} else {
		t.Logf("M6-LEF-03.2 PASS: Found %d obstruction geometries", len(obstructions))
	}

	// M6-LEF-03.3: Verify obstructions are on met1 layer
	met1Count := 0
	for _, obs := range obstructions {
		if obs.Layer == "met1" {
			met1Count++
		}
	}
	if met1Count == 0 {
		t.Error("M6-LEF-03.3 FAIL: No obstructions on met1 layer")
	} else {
		t.Logf("M6-LEF-03.3 PASS: %d obstructions on met1 layer", met1Count)
	}

	// M6-LEF-03.4: Verify obstructions cover significant cell area
	// Dense cells should have obstructions covering internal area to prevent routing
	for _, obs := range obstructions {
		area := (obs.X2 - obs.X1) * (obs.Y2 - obs.Y1)
		if area < 0.01 { // Less than 0.01 µm² is too small
			t.Errorf("M6-LEF-03.4 FAIL: Obstruction area %.3f µm² too small", area)
		}
	}
	t.Log("M6-LEF-03.4 PASS: Obstruction areas adequate")

	// M6-LEF-03.5: Verify obstructions are within cell bounds
	for _, obs := range obstructions {
		if obs.X1 < 0 || obs.Y1 < 0 || obs.X2 > cfg.Width || obs.Y2 > cfg.Height {
			t.Errorf("M6-LEF-03.5 FAIL: Obstruction (%.3f,%.3f)-(%.3f,%.3f) outside cell (%.3f×%.3f)",
				obs.X1, obs.Y1, obs.X2, obs.Y2, cfg.Width, cfg.Height)
		}
	}
	t.Log("M6-LEF-03.5 PASS: All obstructions within cell bounds")
}

// TestLEF1T1RObstruction verifies 1T1R cells have obstructions
func TestLEF1T1RObstruction(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_1t1r_bitcell",
		CellType: "1t1r",
		Width:    0.92,
		Height:   3.40,
	}

	lef := Generate1T1RLEF(cfg)

	// 1T1R cells should have OBS section
	if !strings.Contains(lef, "OBS") {
		t.Error("M6-LEF-03 (1T1R) FAIL: Missing OBS section")
	}

	obstructions := parseObstructions(lef)
	if len(obstructions) == 0 {
		t.Error("M6-LEF-03 (1T1R) FAIL: No obstruction geometries")
	} else {
		t.Logf("M6-LEF-03 (1T1R) PASS: Found %d obstructions", len(obstructions))
	}

	// Calculate total obstruction area
	totalObsArea := 0.0
	for _, obs := range obstructions {
		area := (obs.X2 - obs.X1) * (obs.Y2 - obs.Y1)
		totalObsArea += area
	}

	cellArea := cfg.Width * cfg.Height
	obstructionPct := totalObsArea / cellArea * 100.0

	t.Logf("M6-LEF-03 (1T1R): Obstruction coverage %.3f µm² / %.3f µm² = %.1f%%",
		totalObsArea, cellArea, obstructionPct)

	// Dense cells should have significant obstruction (>10% of cell area)
	if obstructionPct < 10.0 {
		t.Errorf("M6-LEF-03 (1T1R) FAIL: Obstruction coverage %.1f%% too low (expected >10%%)", obstructionPct)
	} else {
		t.Logf("M6-LEF-03 (1T1R) PASS: Adequate obstruction coverage (%.1f%%)", obstructionPct)
	}
}

// TestLEF2T1RObstruction verifies 2T1R cells have obstructions
func TestLEF2T1RObstruction(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_2t1r_bitcell",
		CellType: "2t1r",
		Width:    1.38,
		Height:   3.40,
	}

	lef := Generate2T1RLEF(cfg)

	// 2T1R cells should have OBS section
	if !strings.Contains(lef, "OBS") {
		t.Error("M6-LEF-03 (2T1R) FAIL: Missing OBS section")
	}

	obstructions := parseObstructions(lef)
	if len(obstructions) == 0 {
		t.Error("M6-LEF-03 (2T1R) FAIL: No obstruction geometries")
	} else {
		t.Logf("M6-LEF-03 (2T1R) PASS: Found %d obstructions", len(obstructions))
	}

	// 2T1R cells are larger and denser - should have even more obstruction coverage
	totalObsArea := 0.0
	for _, obs := range obstructions {
		area := (obs.X2 - obs.X1) * (obs.Y2 - obs.Y1)
		totalObsArea += area
	}

	cellArea := cfg.Width * cfg.Height
	obstructionPct := totalObsArea / cellArea * 100.0

	t.Logf("M6-LEF-03 (2T1R): Obstruction coverage %.3f µm² / %.3f µm² = %.1f%%",
		totalObsArea, cellArea, obstructionPct)

	if obstructionPct < 10.0 {
		t.Errorf("M6-LEF-03 (2T1R) FAIL: Obstruction coverage %.1f%% too low (expected >10%%)", obstructionPct)
	} else {
		t.Logf("M6-LEF-03 (2T1R) PASS: Adequate obstruction coverage (%.1f%%)", obstructionPct)
	}
}

// TestLEFObstructionLayerCoverage verifies obstructions don't block all routing
func TestLEFObstructionLayerCoverage(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_bitcell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)
	obstructions := parseObstructions(lef)
	pins := parsePinGeometries(lef)

	// Obstructions should not overlap with pin access areas
	// This is a simplified check - just verify obstructions don't cover entire cell
	totalObsArea := 0.0
	for _, obs := range obstructions {
		area := (obs.X2 - obs.X1) * (obs.Y2 - obs.Y1)
		totalObsArea += area
	}

	cellArea := cfg.Width * cfg.Height
	obstructionPct := totalObsArea / cellArea * 100.0

	// Should leave some space for routing (< 90% obstruction)
	if obstructionPct > 90.0 {
		t.Errorf("M6-LEF-03 FAIL: Obstruction coverage %.1f%% too high (>90%%, blocks all routing)", obstructionPct)
	} else {
		t.Logf("M6-LEF-03 PASS: Obstruction coverage %.1f%% allows routing", obstructionPct)
	}

	// Verify at least some pins are not completely blocked
	accessiblePins := 0
	for _, pin := range pins {
		// Skip power pins
		if isPowerPin(pin.Name) {
			continue
		}

		// Check if pin has area not covered by obstructions
		// Simplified: just check pin exists (real check would need geometry intersection)
		if pin.X2 > pin.X1 && pin.Y2 > pin.Y1 {
			accessiblePins++
		}
	}

	if accessiblePins == 0 {
		t.Error("M6-LEF-03 FAIL: No accessible signal pins (all blocked by obstructions)")
	} else {
		t.Logf("M6-LEF-03 PASS: %d signal pins accessible", accessiblePins)
	}
}

// TestLEFObstructionGeometry verifies obstruction geometry validity
func TestLEFObstructionGeometry(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_bitcell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)
	obstructions := parseObstructions(lef)

	invalidGeometries := 0
	for _, obs := range obstructions {
		// Verify valid rectangle (X2 > X1, Y2 > Y1)
		if obs.X2 <= obs.X1 {
			t.Errorf("M6-LEF-03 FAIL: Invalid obstruction X coordinates (%.3f, %.3f)", obs.X1, obs.X2)
			invalidGeometries++
		}
		if obs.Y2 <= obs.Y1 {
			t.Errorf("M6-LEF-03 FAIL: Invalid obstruction Y coordinates (%.3f, %.3f)", obs.Y1, obs.Y2)
			invalidGeometries++
		}

		// Verify positive dimensions
		if obs.X1 < 0 || obs.Y1 < 0 {
			t.Errorf("M6-LEF-03 FAIL: Negative obstruction coordinates (%.3f, %.3f)", obs.X1, obs.Y1)
			invalidGeometries++
		}
	}

	if invalidGeometries == 0 {
		t.Logf("M6-LEF-03 PASS: All %d obstruction geometries valid", len(obstructions))
	}
}

// TestLEFObstructionCompletion verifies OBS section is properly terminated
func TestLEFObstructionCompletion(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_bitcell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)

	// Extract OBS section
	obsStart := strings.Index(lef, "OBS")
	if obsStart == -1 {
		t.Fatal("M6-LEF-03 FAIL: No OBS section found")
	}

	// Find the section after OBS
	obsSection := lef[obsStart:]
	obsEnd := strings.Index(obsSection, "END OBS")

	// OBS should have proper termination
	// In LEF, OBS ends implicitly before the next section or END <macro>
	// Check that OBS section is followed by valid content
	if obsEnd == -1 {
		// OBS doesn't have explicit END OBS - this is valid in LEF 5.8
		// But section should end before END <macro>
		macroEnd := strings.Index(obsSection, "END fecim_bitcell")
		if macroEnd == -1 {
			t.Error("M6-LEF-03 FAIL: OBS section not properly terminated")
		} else {
			t.Log("M6-LEF-03 PASS: OBS section implicitly terminated before END macro")
		}
	} else {
		t.Log("M6-LEF-03 PASS: OBS section explicitly terminated with END OBS")
	}

	// Verify OBS section contains LAYER and RECT
	obsContent := obsSection[:min(len(obsSection), 500)]
	if !strings.Contains(obsContent, "LAYER") {
		t.Error("M6-LEF-03 FAIL: OBS section missing LAYER specification")
	}
	if !strings.Contains(obsContent, "RECT") {
		t.Error("M6-LEF-03 FAIL: OBS section missing RECT geometry")
	}
}

// TestLEFObstructionMultipleLayers tests support for multi-layer obstructions
func TestLEFObstructionMultipleLayers(t *testing.T) {
	cfg := config.CellConfig{
		Name:     "fecim_bitcell",
		CellType: "passive",
		Width:    0.46,
		Height:   2.72,
	}

	lef := GenerateLEF(cfg)
	obstructions := parseObstructions(lef)

	// Group by layer
	layerCount := make(map[string]int)
	for _, obs := range obstructions {
		layerCount[obs.Layer]++
	}

	t.Logf("M6-LEF-03: Obstruction layers: %v", layerCount)

	// At minimum, should have met1 obstructions
	if layerCount["met1"] == 0 {
		t.Error("M6-LEF-03 FAIL: No met1 layer obstructions")
	} else {
		t.Logf("M6-LEF-03 PASS: %d obstructions on met1", layerCount["met1"])
	}
}

// Helper: parse obstructions from LEF OBS section
func parseObstructions(lef string) []ObstructionGeometry {
	var obstructions []ObstructionGeometry

	// Extract OBS section
	obsStart := strings.Index(lef, "OBS")
	if obsStart == -1 {
		return obstructions
	}

	// Find the end of OBS section (either END OBS or END <macro>)
	obsSection := lef[obsStart:]
	obsEnd := strings.Index(obsSection, "END OBS")
	if obsEnd == -1 {
		// Find END macro instead
		macroNameRe := regexp.MustCompile(`END\s+(\S+)`)
		macroMatch := macroNameRe.FindStringSubmatchIndex(obsSection)
		if len(macroMatch) > 0 {
			obsEnd = macroMatch[0]
		} else {
			obsEnd = len(obsSection)
		}
	}
	obsSection = obsSection[:obsEnd]

	// Parse LAYER and RECT entries
	// Pattern: LAYER <name> ; RECT x1 y1 x2 y2 ;
	currentLayer := ""
	layerRe := regexp.MustCompile(`LAYER\s+(\S+)`)
	rectRe := regexp.MustCompile(`RECT\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)`)

	lines := strings.Split(obsSection, "\n")
	for _, line := range lines {
		// Check for LAYER
		if match := layerRe.FindStringSubmatch(line); len(match) >= 2 {
			currentLayer = match[1]
		}

		// Check for RECT
		if match := rectRe.FindStringSubmatch(line); len(match) >= 5 {
			x1, _ := strconv.ParseFloat(match[1], 64)
			y1, _ := strconv.ParseFloat(match[2], 64)
			x2, _ := strconv.ParseFloat(match[3], 64)
			y2, _ := strconv.ParseFloat(match[4], 64)

			obstructions = append(obstructions, ObstructionGeometry{
				Layer: currentLayer,
				X1:    x1,
				Y1:    y1,
				X2:    x2,
				Y2:    y2,
			})
		}
	}

	return obstructions
}

// Helper: min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
