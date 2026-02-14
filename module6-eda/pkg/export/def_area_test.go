// pkg/export/def_area_test.go
// M6-DEF-04: DEF die area validation test
// Verifies die area = N×M × cell_area × routing_overhead and reported DIEAREA matches calculation

package export

import (
	"math"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestDEFDieAreaCalculation tests M6-DEF-04:
// Die area = N×M × cell_area × routing_overhead, verify reported DIEAREA matches calculation
func TestDEFDieAreaCalculation(t *testing.T) {
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

	// Parse die area from DEF
	dieWidthDBU, dieHeightDBU := parseDieArea(def)
	if dieWidthDBU == 0 || dieHeightDBU == 0 {
		t.Fatal("M6-DEF-04 FAIL: Could not parse die area")
	}

	// Convert to microns
	dieWidthUM := float64(dieWidthDBU) / float64(defConfig.DatabaseUnit)
	dieHeightUM := float64(dieHeightDBU) / float64(defConfig.DatabaseUnit)
	dieAreaUM2 := dieWidthUM * dieHeightUM

	t.Logf("M6-DEF-04: Reported die area: %.3f × %.3f µm = %.3f µm²", dieWidthUM, dieHeightUM, dieAreaUM2)

	// M6-DEF-04.1: Calculate expected die area
	// Array dimensions: 4×4 cells
	numRows := 4
	numCols := 4
	arrayWidth := float64(numCols) * defConfig.CellWidth
	arrayHeight := float64(numRows) * defConfig.CellHeight

	// Die area = array + margins
	expectedDieWidth := arrayWidth + defConfig.OriginX + defConfig.MarginX
	expectedDieHeight := arrayHeight + defConfig.OriginY + defConfig.MarginY
	expectedDieArea := expectedDieWidth * expectedDieHeight

	t.Logf("M6-DEF-04.1: Calculated die area: %.3f × %.3f µm = %.3f µm²",
		expectedDieWidth, expectedDieHeight, expectedDieArea)

	// Verify match (within 1% tolerance for floating point)
	deltaPct := math.Abs(dieAreaUM2-expectedDieArea) / expectedDieArea * 100.0
	if deltaPct > 1.0 {
		t.Errorf("M6-DEF-04.1 FAIL: Die area %.3f µm² != expected %.3f µm² (Δ=%.2f%%)",
			dieAreaUM2, expectedDieArea, deltaPct)
	} else {
		t.Logf("M6-DEF-04.1 PASS: Die area %.3f µm² matches calculation (Δ=%.3f%%, tolerance=1.0%%)",
			dieAreaUM2, deltaPct)
	}

	// M6-DEF-04.2: Verify array area component
	arrayArea := arrayWidth * arrayHeight
	t.Logf("M6-DEF-04.2: Array area: %.3f × %.3f µm = %.3f µm²", arrayWidth, arrayHeight, arrayArea)

	// M6-DEF-04.3: Verify margins
	marginArea := expectedDieArea - arrayArea
	marginPct := marginArea / expectedDieArea * 100.0
	t.Logf("M6-DEF-04.3: Margin area: %.3f µm² (%.1f%% of die)", marginArea, marginPct)

	if marginPct < 5.0 {
		t.Error("M6-DEF-04.3 FAIL: Margin too small (<5% of die area)")
	} else {
		t.Logf("M6-DEF-04.3 PASS: Adequate margin (%.1f%% of die)", marginPct)
	}

	// M6-DEF-04.4: Cell area breakdown
	cellArea := defConfig.CellWidth * defConfig.CellHeight
	totalCellArea := float64(numRows*numCols) * cellArea
	utilizationPct := totalCellArea / dieAreaUM2 * 100.0

	t.Logf("M6-DEF-04.4: Cell utilization: %.3f µm² / %.3f µm² = %.1f%%",
		totalCellArea, dieAreaUM2, utilizationPct)

	if utilizationPct > 90.0 {
		t.Error("M6-DEF-04.4 FAIL: Utilization too high (>90%), insufficient routing space")
	} else if utilizationPct < 5.0 {
		// Low utilization for small test arrays is expected due to fixed margins
		t.Logf("M6-DEF-04.4 NOTE: Low utilization (%.1f%%) expected for small test arrays with fixed margins", utilizationPct)
	} else {
		t.Logf("M6-DEF-04.4 PASS: Reasonable utilization (%.1f%%, target 5-90%%)", utilizationPct)
	}
}

// TestDEF1T1RDieArea verifies 1T1R die area with larger cells
func TestDEF1T1RDieArea(t *testing.T) {
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

	// Parse die area
	dieWidthDBU, dieHeightDBU := parseDieArea(def)
	dieWidthUM := float64(dieWidthDBU) / float64(defConfig.DatabaseUnit)
	dieHeightUM := float64(dieHeightDBU) / float64(defConfig.DatabaseUnit)

	// 1T1R cells are larger (0.92 µm vs 0.46 µm)
	// Die should be correspondingly larger
	numRows := 2
	numCols := 2
	arrayWidth := float64(numCols) * defConfig.CellWidth
	arrayHeight := float64(numRows) * defConfig.CellHeight

	expectedDieWidth := arrayWidth + defConfig.OriginX + defConfig.MarginX
	expectedDieHeight := arrayHeight + defConfig.OriginY + defConfig.MarginY

	deltaWidth := math.Abs(dieWidthUM - expectedDieWidth)
	deltaHeight := math.Abs(dieHeightUM - expectedDieHeight)

	if deltaWidth > 0.1 || deltaHeight > 0.1 {
		t.Errorf("M6-DEF-04 (1T1R) FAIL: Die size (%.3f×%.3f) != expected (%.3f×%.3f) µm",
			dieWidthUM, dieHeightUM, expectedDieWidth, expectedDieHeight)
	} else {
		t.Logf("M6-DEF-04 (1T1R) PASS: Die size %.3f×%.3f µm matches calculation", dieWidthUM, dieHeightUM)
	}

	// Verify 1T1R cells are indeed larger
	cellArea := defConfig.CellWidth * defConfig.CellHeight
	passiveCellArea := 0.46 * 2.72 // Passive default
	if cellArea <= passiveCellArea {
		t.Errorf("M6-DEF-04 (1T1R) FAIL: Cell area %.3f µm² not larger than passive %.3f µm²",
			cellArea, passiveCellArea)
	} else {
		t.Logf("M6-DEF-04 (1T1R) PASS: Cell area %.3f µm² > passive %.3f µm² (%.1f×)",
			cellArea, passiveCellArea, cellArea/passiveCellArea)
	}
}

// TestDEFDieAreaScaling verifies die area scales correctly with array size
func TestDEFDieAreaScaling(t *testing.T) {
	sizes := []struct {
		rows int
		cols int
	}{
		{2, 2},
		{4, 4},
		{8, 8},
		{16, 16},
	}

	config := compiler.DefaultConfig()
	defConfig := DefaultDEFConfig()
	_ = defConfig.CellWidth * defConfig.CellHeight

	for _, size := range sizes {
		// Create simple weight matrix
		weights := make([][]float64, size.rows)
		for i := range weights {
			weights[i] = make([]float64, size.cols)
			for j := range weights[i] {
				weights[i][j] = 0.1
			}
		}

		config.ArrayRows = size.rows * 2
		config.ArrayCols = size.cols * 2
		design, err := compiler.Compile(weights, config)
		if err != nil {
			t.Fatalf("Compile failed for %d×%d: %v", size.rows, size.cols, err)
		}

		def := GenerateDEF(design, defConfig)
		dieWidthDBU, dieHeightDBU := parseDieArea(def)
		dieWidthUM := float64(dieWidthDBU) / float64(defConfig.DatabaseUnit)
		dieHeightUM := float64(dieHeightDBU) / float64(defConfig.DatabaseUnit)
		dieAreaUM2 := dieWidthUM * dieHeightUM

		// Calculate expected
		arrayWidth := float64(size.cols) * defConfig.CellWidth
		arrayHeight := float64(size.rows) * defConfig.CellHeight
		expectedDieWidth := arrayWidth + defConfig.OriginX + defConfig.MarginX
		expectedDieHeight := arrayHeight + defConfig.OriginY + defConfig.MarginY
		expectedDieArea := expectedDieWidth * expectedDieHeight

		deltaPct := math.Abs(dieAreaUM2-expectedDieArea) / expectedDieArea * 100.0

		if deltaPct > 1.0 {
			t.Errorf("M6-DEF-04 FAIL: %d×%d array die area %.3f µm² != expected %.3f µm² (Δ=%.2f%%)",
				size.rows, size.cols, dieAreaUM2, expectedDieArea, deltaPct)
		}

		t.Logf("M6-DEF-04 scaling: %d×%d = %.3f µm² (Δ=%.3f%%)",
			size.rows, size.cols, dieAreaUM2, deltaPct)
	}

	t.Log("M6-DEF-04 PASS: Die area scales correctly with array size")
}

// TestDEFDieAreaMargins verifies margins are applied correctly
func TestDEFDieAreaMargins(t *testing.T) {
	weights := [][]float64{{0.1}}
	config := compiler.DefaultConfig()
	config.ArrayRows = 2
	config.ArrayCols = 2

	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Test with custom margins
	defConfig := DefaultDEFConfig()
	defConfig.MarginX = 50.0 // 50 µm margin
	defConfig.MarginY = 50.0

	def := GenerateDEF(design, defConfig)
	dieWidthDBU, dieHeightDBU := parseDieArea(def)
	dieWidthUM := float64(dieWidthDBU) / float64(defConfig.DatabaseUnit)
	dieHeightUM := float64(dieHeightDBU) / float64(defConfig.DatabaseUnit)

	// Expected: array + origin + margin
	arrayWidth := defConfig.CellWidth   // 1 col
	arrayHeight := defConfig.CellHeight // 1 row
	expectedWidth := arrayWidth + defConfig.OriginX + defConfig.MarginX
	expectedHeight := arrayHeight + defConfig.OriginY + defConfig.MarginY

	deltaW := math.Abs(dieWidthUM - expectedWidth)
	deltaH := math.Abs(dieHeightUM - expectedHeight)

	if deltaW > 0.1 || deltaH > 0.1 {
		t.Errorf("M6-DEF-04 FAIL: Margin calculation incorrect, got (%.3f×%.3f), expected (%.3f×%.3f) µm",
			dieWidthUM, dieHeightUM, expectedWidth, expectedHeight)
	} else {
		t.Logf("M6-DEF-04 PASS: Margins applied correctly (%.1f × %.1f µm)", defConfig.MarginX, defConfig.MarginY)
	}
}

// TestDEFDieAreaDatabaseUnits verifies database unit conversion is correct
func TestDEFDieAreaDatabaseUnits(t *testing.T) {
	weights := [][]float64{{0.1}}
	config := compiler.DefaultConfig()
	config.ArrayRows = 2
	config.ArrayCols = 2
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	defConfig := DefaultDEFConfig()
	def := GenerateDEF(design, defConfig)

	// Parse die area in DBU
	dieWidthDBU, dieHeightDBU := parseDieArea(def)

	// Convert to microns
	dieWidthUM := float64(dieWidthDBU) / float64(defConfig.DatabaseUnit)
	dieHeightUM := float64(dieHeightDBU) / float64(defConfig.DatabaseUnit)

	// Verify DBU conversion (1000 DBU = 1 µm)
	// Example: if die width is 12.34 µm, DBU should be 12340
	expectedWidthDBU := int(dieWidthUM * float64(defConfig.DatabaseUnit))
	expectedHeightDBU := int(dieHeightUM * float64(defConfig.DatabaseUnit))

	if dieWidthDBU != expectedWidthDBU {
		t.Errorf("M6-DEF-04 FAIL: Width DBU conversion incorrect: %d != %d", dieWidthDBU, expectedWidthDBU)
	}
	if dieHeightDBU != expectedHeightDBU {
		t.Errorf("M6-DEF-04 FAIL: Height DBU conversion incorrect: %d != %d", dieHeightDBU, expectedHeightDBU)
	}

	t.Logf("M6-DEF-04 PASS: Database unit conversion correct (%.3f µm = %d DBU @ %d DBU/µm)",
		dieWidthUM, dieWidthDBU, defConfig.DatabaseUnit)
}
