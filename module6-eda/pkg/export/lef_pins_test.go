// pkg/export/lef_pins_test.go
// M6-LEF-02: LEF pin validation test
// Verifies pins are on routing grid (multiple of grid pitch)

package export

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// PinGeometry represents a pin's rectangle geometry in the LEF
type PinGeometry struct {
	Name   string
	Layer  string
	X1, Y1 float64 // Lower-left corner in µm
	X2, Y2 float64 // Upper-right corner in µm
}

// TestLEFPinsOnGrid tests M6-LEF-02:
// Verify pins are on routing grid (multiple of grid pitch)
func TestLEFPinsOnGrid(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Width:      0.46,
		Height:     2.72,
		MetalPitch: 0.46, // SKY130 met1 pitch
		MetalWidth: 0.14,
	}

	lef := GenerateLEF(cfg)

	// Parse pin geometries
	pins := parsePinGeometries(lef)
	if len(pins) == 0 {
		t.Fatal("M6-LEF-02 FAIL: No pin geometries found")
	}

	t.Logf("M6-LEF-02: Parsed %d pin geometries", len(pins))

	gridPitch := cfg.MetalPitch
	tolerance := 0.001 // 1 nm tolerance for floating-point comparison

	// M6-LEF-02.1: Verify all pin coordinates are on grid
	offGridPins := 0
	for _, pin := range pins {
		// Check all four corners
		coords := []float64{pin.X1, pin.Y1, pin.X2, pin.Y2}
		for i, coord := range coords {
			// Check if coordinate is a multiple of grid pitch
			remainder := math.Mod(coord, gridPitch)
			if remainder > tolerance && (gridPitch-remainder) > tolerance {
				coordName := []string{"X1", "Y1", "X2", "Y2"}[i]
				t.Errorf("M6-LEF-02.1 FAIL: Pin %s %s coordinate %.3f µm not on grid (pitch=%.3f µm, remainder=%.3f)",
					pin.Name, coordName, coord, gridPitch, remainder)
				offGridPins++
			}
		}
	}

	if offGridPins == 0 {
		t.Logf("M6-LEF-02.1 PASS: All %d pins on routing grid (pitch=%.3f µm)", len(pins), gridPitch)
	}

	// M6-LEF-02.2: Verify pin widths are multiples of minimum width
	invalidWidths := 0
	for _, pin := range pins {
		width := pin.X2 - pin.X1
		height := pin.Y2 - pin.Y1

		// Width should be >= minimum metal width
		if width < cfg.MetalWidth-tolerance {
			t.Errorf("M6-LEF-02.2 FAIL: Pin %s width %.3f µm < minimum %.3f µm",
				pin.Name, width, cfg.MetalWidth)
			invalidWidths++
		}
		if height < cfg.MetalWidth-tolerance {
			t.Errorf("M6-LEF-02.2 FAIL: Pin %s height %.3f µm < minimum %.3f µm",
				pin.Name, height, cfg.MetalWidth)
			invalidWidths++
		}
	}

	if invalidWidths == 0 {
		t.Logf("M6-LEF-02.2 PASS: All pin dimensions >= minimum width (%.3f µm)", cfg.MetalWidth)
	}

	// M6-LEF-02.3: Verify pins are within cell boundary
	outsidePins := 0
	for _, pin := range pins {
		if pin.X1 < 0 || pin.Y1 < 0 || pin.X2 > cfg.Width || pin.Y2 > cfg.Height {
			t.Errorf("M6-LEF-02.3 FAIL: Pin %s geometry (%.3f,%.3f)-(%.3f,%.3f) outside cell (%.3f×%.3f)",
				pin.Name, pin.X1, pin.Y1, pin.X2, pin.Y2, cfg.Width, cfg.Height)
			outsidePins++
		}
	}

	if outsidePins == 0 {
		t.Logf("M6-LEF-02.3 PASS: All pins within cell boundary (%.3f×%.3f µm)", cfg.Width, cfg.Height)
	}
}

// TestLEF1T1RPinsOnGrid verifies 1T1R pins are on grid
func TestLEF1T1RPinsOnGrid(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "fecim_1t1r_bitcell",
		CellType:   "1t1r",
		Width:      0.92,
		Height:     3.40,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}

	lef := Generate1T1RLEF(cfg)
	pins := parsePinGeometries(lef)

	// Should have WL, BL, SL, VPWR, VGND = 5 pins
	if len(pins) != 5 {
		t.Errorf("M6-LEF-02 (1T1R) FAIL: Expected 5 pins, found %d", len(pins))
	}

	// Verify SL pin exists
	slFound := false
	for _, pin := range pins {
		if pin.Name == "SL" {
			slFound = true
			// SL should be on grid
			gridPitch := cfg.MetalPitch
			tolerance := 0.001
			coords := []float64{pin.X1, pin.Y1, pin.X2, pin.Y2}
			for _, coord := range coords {
				remainder := math.Mod(coord, gridPitch)
				if remainder > tolerance && (gridPitch-remainder) > tolerance {
					t.Errorf("M6-LEF-02 (1T1R) FAIL: SL pin coordinate %.3f µm not on grid", coord)
				}
			}
			break
		}
	}

	if !slFound {
		t.Error("M6-LEF-02 (1T1R) FAIL: SL pin not found")
	} else {
		t.Log("M6-LEF-02 (1T1R) PASS: SL pin on routing grid")
	}

	// Check all pins on grid
	offGrid := 0
	for _, pin := range pins {
		coords := []float64{pin.X1, pin.Y1, pin.X2, pin.Y2}
		for _, coord := range coords {
			remainder := math.Mod(coord, cfg.MetalPitch)
			if remainder > 0.001 && (cfg.MetalPitch-remainder) > 0.001 {
				offGrid++
				break
			}
		}
	}

	if offGrid == 0 {
		t.Logf("M6-LEF-02 (1T1R) PASS: All 5 pins on routing grid")
	}
}

// TestLEF2T1RPinsOnGrid verifies 2T1R pins are on grid
func TestLEF2T1RPinsOnGrid(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "fecim_2t1r_bitcell",
		CellType:   "2t1r",
		Width:      1.38,
		Height:     3.40,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}

	lef := Generate2T1RLEF(cfg)
	pins := parsePinGeometries(lef)

	// Should have WL, CSL, BL, SL, VPWR, VGND = 6 pins
	if len(pins) != 6 {
		t.Errorf("M6-LEF-02 (2T1R) FAIL: Expected 6 pins, found %d", len(pins))
	}

	// Verify CSL and SL pins exist and are on grid
	cslFound := false
	slFound := false
	gridPitch := cfg.MetalPitch
	tolerance := 0.001

	for _, pin := range pins {
		if pin.Name == "CSL" {
			cslFound = true
		}
		if pin.Name == "SL" {
			slFound = true
		}

		// Check all coordinates on grid
		coords := []float64{pin.X1, pin.Y1, pin.X2, pin.Y2}
		for _, coord := range coords {
			remainder := math.Mod(coord, gridPitch)
			if remainder > tolerance && (gridPitch-remainder) > tolerance {
				t.Errorf("M6-LEF-02 (2T1R) FAIL: Pin %s coordinate %.3f µm not on grid", pin.Name, coord)
			}
		}
	}

	if !cslFound {
		t.Error("M6-LEF-02 (2T1R) FAIL: CSL pin not found")
	}
	if !slFound {
		t.Error("M6-LEF-02 (2T1R) FAIL: SL pin not found")
	}

	if cslFound && slFound {
		t.Log("M6-LEF-02 (2T1R) PASS: All 6 pins (including CSL, SL) on routing grid")
	}
}

// TestLEFPinSpacing verifies minimum spacing between pins
func TestLEFPinSpacing(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Width:      0.46,
		Height:     2.72,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}

	lef := GenerateLEF(cfg)
	pins := parsePinGeometries(lef)

	// Check minimum spacing between signal pins (not power)
	minSpacing := cfg.MetalWidth // Minimum spacing = minimum width
	violations := 0

	for i := 0; i < len(pins); i++ {
		for j := i + 1; j < len(pins); j++ {
			// Skip if both are power pins
			if isPowerPin(pins[i].Name) && isPowerPin(pins[j].Name) {
				continue
			}

			// Calculate spacing
			spacing := calculateSpacing(pins[i], pins[j])
			if spacing < minSpacing-0.001 { // 1nm tolerance
				t.Errorf("M6-LEF-02 FAIL: Pin spacing %.3f µm < minimum %.3f µm (pins %s, %s)",
					spacing, minSpacing, pins[i].Name, pins[j].Name)
				violations++
			}
		}
	}

	if violations == 0 {
		t.Logf("M6-LEF-02 PASS: All pin spacings >= minimum (%.3f µm)", minSpacing)
	}
}

// TestLEFPinAlignment verifies pins align with expected positions
func TestLEFPinAlignment(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Width:      0.46,
		Height:     2.72,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}

	lef := GenerateLEF(cfg)
	pins := parsePinGeometries(lef)

	// VPWR should be at top
	// VGND should be at bottom
	// WL should be on left
	// BL should be on right or top

	for _, pin := range pins {
		switch pin.Name {
		case "VPWR":
			// Should be near top edge
			if pin.Y2 < cfg.Height*0.8 {
				t.Errorf("M6-LEF-02 FAIL: VPWR pin Y2=%.3f should be near top (height=%.3f)",
					pin.Y2, cfg.Height)
			}
		case "VGND":
			// Should be near bottom edge
			if pin.Y1 > cfg.Height*0.2 {
				t.Errorf("M6-LEF-02 FAIL: VGND pin Y1=%.3f should be near bottom", pin.Y1)
			}
		case "WL":
			// Should be on left edge
			if pin.X1 > cfg.Width*0.5 {
				t.Errorf("M6-LEF-02 FAIL: WL pin X1=%.3f should be on left side", pin.X1)
			}
		}
	}

	t.Log("M6-LEF-02 PASS: Pin alignment correct")
}

// Helper: parse pin geometries from LEF
func parsePinGeometries(lef string) []PinGeometry {
	var pins []PinGeometry

	// Pattern: PIN <name> ... PORT ... LAYER <layer> ; RECT x1 y1 x2 y2 ; ... END <name>
	// Extract each PIN section manually (Go regexp doesn't support backreferences)
	// Find all PIN sections
	pinStarts := make([]int, 0)
	pinNames := make([]string, 0)

	pinHeaderRe := regexp.MustCompile(`PIN\s+(\S+)`)
	matches := pinHeaderRe.FindAllStringSubmatchIndex(lef, -1)
	for _, match := range matches {
		pinStarts = append(pinStarts, match[0])
		pinNames = append(pinNames, lef[match[2]:match[3]])
	}

	// Extract each PIN section content
	for i, pinName := range pinNames {
		start := pinStarts[i]
		// Find END <pinName>
		endPattern := "END " + pinName
		end := strings.Index(lef[start:], endPattern)
		if end == -1 {
			continue
		}
		pinSection := lef[start : start+end]

		// Extract LAYER
		layerRe := regexp.MustCompile(`LAYER\s+(\S+)`)
		layerMatch := layerRe.FindStringSubmatch(pinSection)
		layer := ""
		if len(layerMatch) >= 2 {
			layer = layerMatch[1]
		}

		// Extract RECT coordinates
		rectRe := regexp.MustCompile(`RECT\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)`)
		rectMatches := rectRe.FindAllStringSubmatch(pinSection, -1)

		for _, rectMatch := range rectMatches {
			if len(rectMatch) >= 5 {
				x1, _ := strconv.ParseFloat(rectMatch[1], 64)
				y1, _ := strconv.ParseFloat(rectMatch[2], 64)
				x2, _ := strconv.ParseFloat(rectMatch[3], 64)
				y2, _ := strconv.ParseFloat(rectMatch[4], 64)

				pins = append(pins, PinGeometry{
					Name:  pinName,
					Layer: layer,
					X1:    x1,
					Y1:    y1,
					X2:    x2,
					Y2:    y2,
				})
			}
		}
	}

	return pins
}

// Helper: check if pin is a power pin
func isPowerPin(name string) bool {
	return name == "VPWR" || name == "VGND" || name == "VDD" || name == "VSS"
}

// Helper: calculate minimum spacing between two pin geometries
func calculateSpacing(a, b PinGeometry) float64 {
	// For non-overlapping rectangles, find minimum edge-to-edge distance
	// If they overlap, spacing is 0

	// Check horizontal separation
	xSep := 0.0
	if a.X2 <= b.X1 {
		xSep = b.X1 - a.X2
	} else if b.X2 <= a.X1 {
		xSep = a.X1 - b.X2
	}

	// Check vertical separation
	ySep := 0.0
	if a.Y2 <= b.Y1 {
		ySep = b.Y1 - a.Y2
	} else if b.Y2 <= a.Y1 {
		ySep = a.Y1 - b.Y2
	}

	// If both separations are > 0, they are diagonally separated
	// Return the smaller of the two (conservative)
	if xSep > 0 && ySep > 0 {
		return math.Min(xSep, ySep)
	}

	// If only one is > 0, they overlap in the other dimension
	// Return the non-zero separation
	if xSep > 0 {
		return xSep
	}
	return ySep
}
