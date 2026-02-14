// pkg/export/def_placement_test.go
// M6-DEF-02: DEF placement validation test
// Verifies component placements don't overlap and cells have unique coordinates

package export

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// PlacementInfo represents a component's placement in the DEF
type PlacementInfo struct {
	Name   string
	X      int // in database units
	Y      int
	Width  int // in database units
	Height int
}

// TestDEFPlacementNoOverlap tests M6-DEF-02:
// Check component placements don't overlap, each cell has unique (x,y) or non-overlapping bounding boxes
func TestDEFPlacementNoOverlap(t *testing.T) {
	// Create 4×4 weight matrix
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	defConfig := DEFConfigFrom(design)
	def := GenerateDEF(design, defConfig)

	// Parse placements from DEF
	placements := parsePlacements(def, defConfig)

	if len(placements) != 16 {
		t.Errorf("M6-DEF-02 FAIL: Expected 16 placements, got %d", len(placements))
	}

	// M6-DEF-02.1: Check for overlapping placements
	overlaps := 0
	for i := 0; i < len(placements); i++ {
		for j := i + 1; j < len(placements); j++ {
			if placementsOverlap(placements[i], placements[j]) {
				t.Errorf("M6-DEF-02.1 FAIL: Overlap detected between %s and %s", placements[i].Name, placements[j].Name)
				overlaps++
			}
		}
	}
	if overlaps == 0 {
		t.Logf("M6-DEF-02.1 PASS: No overlapping placements (checked %d cell pairs)", len(placements)*(len(placements)-1)/2)
	}

	// M6-DEF-02.2: Verify unique coordinates
	coordMap := make(map[string]bool)
	duplicates := 0
	for _, p := range placements {
		coord := fmt.Sprintf("%d,%d", p.X, p.Y)
		if coordMap[coord] {
			t.Errorf("M6-DEF-02.2 FAIL: Duplicate coordinate (%d, %d) for %s", p.X, p.Y, p.Name)
			duplicates++
		}
		coordMap[coord] = true
	}
	if duplicates == 0 {
		t.Logf("M6-DEF-02.2 PASS: All 16 cells have unique coordinates")
	}

	// M6-DEF-02.3: Verify regular grid placement
	cellWidthDBU := defConfig.DatabaseUnit * int(defConfig.CellWidth*1000) / 1000
	cellHeightDBU := defConfig.DatabaseUnit * int(defConfig.CellHeight*1000) / 1000
	originXDBU := int(defConfig.OriginX * float64(defConfig.DatabaseUnit))
	originYDBU := int(defConfig.OriginY * float64(defConfig.DatabaseUnit))

	gridErrors := 0
	for _, p := range placements {
		// Extract row/col from instance name R_{row}_{col}
		parts := strings.Split(p.Name, "_")
		if len(parts) != 3 {
			continue
		}
		row, _ := strconv.Atoi(parts[1])
		col, _ := strconv.Atoi(parts[2])

		expectedX := originXDBU + col*cellWidthDBU
		expectedY := originYDBU + row*cellHeightDBU

		if p.X != expectedX || p.Y != expectedY {
			t.Errorf("M6-DEF-02.3 FAIL: %s at (%d,%d), expected (%d,%d)", p.Name, p.X, p.Y, expectedX, expectedY)
			gridErrors++
		}
	}
	if gridErrors == 0 {
		t.Logf("M6-DEF-02.3 PASS: All cells placed on regular grid (pitch=%.2f×%.2f µm)",
			defConfig.CellWidth, defConfig.CellHeight)
	}

	t.Logf("M6-DEF-02 placement validation: 16 cells, 0 overlaps, grid pitch=%.3f µm", defConfig.CellWidth)
}

// TestDEFPlacementBounds verifies all cells are within die area
func TestDEFPlacementBounds(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 4
	config.ArrayCols = 6
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	defConfig := DEFConfigFrom(design)
	def := GenerateDEF(design, defConfig)

	// Parse die area
	dieWidth, dieHeight := parseDieArea(def)
	if dieWidth == 0 || dieHeight == 0 {
		t.Fatalf("M6-DEF-02 FAIL: Could not parse die area")
	}

	// Parse placements
	placements := parsePlacements(def, defConfig)

	// Check all cells are within die bounds
	outOfBounds := 0
	for _, p := range placements {
		if p.X < 0 || p.Y < 0 || p.X+p.Width > dieWidth || p.Y+p.Height > dieHeight {
			t.Errorf("M6-DEF-02 FAIL: %s placement (%d,%d) size (%d×%d) exceeds die area (%d×%d)",
				p.Name, p.X, p.Y, p.Width, p.Height, dieWidth, dieHeight)
			outOfBounds++
		}
	}

	if outOfBounds == 0 {
		t.Logf("M6-DEF-02 PASS: All %d cells within die area (%d×%d DBU)",
			len(placements), dieWidth, dieHeight)
	}
}

// TestDEF1T1RPlacement verifies 1T1R cells use correct larger pitch
func TestDEF1T1RPlacement(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.Config1T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	defConfig := DEFConfigFrom(design)
	def := GenerateDEF(design, defConfig)

	// 1T1R should use larger cell pitch (0.92 µm vs 0.46 µm)
	if defConfig.CellWidth < 0.9 {
		t.Errorf("M6-DEF-02 (1T1R) FAIL: Cell width %.3f µm too small for 1T1R (expected ≥0.92 µm)", defConfig.CellWidth)
	}

	// Parse placements and check spacing
	placements := parsePlacements(def, defConfig)

	// Check adjacent cells have correct spacing
	// Find R_0_0 and R_0_1
	var p00, p01 *PlacementInfo
	for i := range placements {
		if placements[i].Name == "R_0_0" {
			p00 = &placements[i]
		}
		if placements[i].Name == "R_0_1" {
			p01 = &placements[i]
		}
	}

	if p00 != nil && p01 != nil {
		spacing := p01.X - p00.X
		expectedSpacing := int(defConfig.CellWidth * float64(defConfig.DatabaseUnit))
		if spacing != expectedSpacing {
			t.Errorf("M6-DEF-02 (1T1R) FAIL: Horizontal spacing %d DBU, expected %d DBU",
				spacing, expectedSpacing)
		} else {
			t.Logf("M6-DEF-02 (1T1R) PASS: Correct spacing %d DBU (%.2f µm)", spacing, defConfig.CellWidth)
		}
	}

	// Verify no overlaps with larger cells
	overlaps := 0
	for i := 0; i < len(placements); i++ {
		for j := i + 1; j < len(placements); j++ {
			if placementsOverlap(placements[i], placements[j]) {
				overlaps++
			}
		}
	}
	if overlaps > 0 {
		t.Errorf("M6-DEF-02 (1T1R) FAIL: Found %d overlapping cells", overlaps)
	} else {
		t.Log("M6-DEF-02 (1T1R) PASS: No overlapping cells")
	}
}

// Helper: parse placements from DEF
func parsePlacements(def string, config DEFConfig) []PlacementInfo {
	var placements []PlacementInfo

	// Match: - R_0_0 fecim_bit + FIXED ( 10000 10000 ) N ;
	re := regexp.MustCompile(`-\s+(R_\d+_\d+)\s+\S+\s+\+\s+FIXED\s+\(\s+(\d+)\s+(\d+)\s+\)`)
	matches := re.FindAllStringSubmatch(def, -1)

	cellWidthDBU := int(config.CellWidth * float64(config.DatabaseUnit))
	cellHeightDBU := int(config.CellHeight * float64(config.DatabaseUnit))

	for _, match := range matches {
		if len(match) >= 4 {
			name := match[1]
			x, _ := strconv.Atoi(match[2])
			y, _ := strconv.Atoi(match[3])

			placements = append(placements, PlacementInfo{
				Name:   name,
				X:      x,
				Y:      y,
				Width:  cellWidthDBU,
				Height: cellHeightDBU,
			})
		}
	}

	return placements
}

// Helper: check if two placements overlap
func placementsOverlap(a, b PlacementInfo) bool {
	// Two rectangles overlap if they intersect in both X and Y
	xOverlap := a.X < b.X+b.Width && a.X+a.Width > b.X
	yOverlap := a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
	return xOverlap && yOverlap
}

// Helper: parse die area from DEF
func parseDieArea(def string) (width, height int) {
	// Match: DIEAREA ( 0 0 ) ( 12200 35440 ) ;
	re := regexp.MustCompile(`DIEAREA\s+\(\s+\d+\s+\d+\s+\)\s+\(\s+(\d+)\s+(\d+)\s+\)`)
	match := re.FindStringSubmatch(def)
	if len(match) >= 3 {
		width, _ = strconv.Atoi(match[1])
		height, _ = strconv.Atoi(match[2])
	}
	return
}
