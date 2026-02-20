// pkg/export/svg_test.go
// Tests for SVG layout visualization generator
package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// ============================================================================
// DefaultSVGConfig Tests
// ============================================================================

func TestDefaultSVGConfig(t *testing.T) {
	cfg := DefaultSVGConfig()

	// Check default values
	if cfg.CellWidth != 40 {
		t.Errorf("Default CellWidth should be 40, got %f", cfg.CellWidth)
	}
	if cfg.CellHeight != 60 {
		t.Errorf("Default CellHeight should be 60, got %f", cfg.CellHeight)
	}
	if cfg.Margin != 80 {
		t.Errorf("Default Margin should be 80, got %f", cfg.Margin)
	}
	if !cfg.ShowGrid {
		t.Error("ShowGrid should be true by default")
	}
	if !cfg.ShowLabels {
		t.Error("ShowLabels should be true by default")
	}
	if cfg.ShowCellIDs {
		t.Error("ShowCellIDs should be false by default")
	}
	if cfg.ColorScheme != "default" {
		t.Errorf("Default ColorScheme should be 'default', got %s", cfg.ColorScheme)
	}
}

// ============================================================================
// GenerateLayoutSVG Tests - Passive Architecture
// ============================================================================

func TestGenerateLayoutSVG_PassiveBasic(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check SVG header
	if !strings.Contains(svg, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>") {
		t.Error("Missing XML declaration")
	}
	if !strings.Contains(svg, "<svg xmlns=\"http://www.w3.org/2000/svg\"") {
		t.Error("Missing SVG namespace")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("Missing closing SVG tag")
	}
}

func TestGenerateLayoutSVG_PassiveTitle(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check title
	if !strings.Contains(svg, "FeCIM 4x4 Crossbar (Passive)") {
		t.Error("Missing or incorrect title")
	}
}

func TestGenerateLayoutSVG_PassiveWordLines(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check WL labels (4 rows)
	for i := 0; i < 4; i++ {
		if !strings.Contains(svg, "WL["+itoa(i)+"]") {
			t.Errorf("Missing WL[%d] label", i)
		}
	}

	// Check wire class
	if !strings.Contains(svg, "class=\"wire-wl\"") {
		t.Error("Missing WL wire class")
	}
}

func TestGenerateLayoutSVG_PassiveBitLines(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check BL labels (4 cols)
	for i := 0; i < 4; i++ {
		if !strings.Contains(svg, "BL["+itoa(i)+"]") {
			t.Errorf("Missing BL[%d] label", i)
		}
	}

	// Check wire class
	if !strings.Contains(svg, "class=\"wire-bl\"") {
		t.Error("Missing BL wire class")
	}
}

func TestGenerateLayoutSVG_PassiveNoSourceLines(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Passive architecture should NOT have SL
	if strings.Contains(svg, "SL[") {
		t.Error("Passive architecture should not have SL labels")
	}
	if strings.Contains(svg, "class=\"wire-sl\"") {
		t.Error("Passive architecture should not have SL wires")
	}
}

func TestGenerateLayoutSVG_PassiveCells(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 3
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check for passive cell class
	if !strings.Contains(svg, "cell-passive") {
		t.Error("Missing cell-passive class")
	}

	// Count cell rectangles (should be 2*3 = 6 for passive, single rect per cell)
	cellCount := strings.Count(svg, "<rect")
	// Passive cells have 1 rect each (6 cells) + 1 background rect
	// Grid lines use <line> not <rect>
	if cellCount < 7 {
		t.Errorf("Expected at least 7 rects (6 cells + 1 background), got %d", cellCount)
	}
}

// ============================================================================
// GenerateLayoutSVG Tests - 1T1R Architecture
// ============================================================================

func TestGenerateLayoutSVG_1T1RTitle(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "1t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check title shows 1T1R
	if !strings.Contains(svg, "FeCIM 4x4 Crossbar (1T1R)") {
		t.Error("Missing or incorrect 1T1R title")
	}
}

func TestGenerateLayoutSVG_1T1RSourceLines(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "1t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// 1T1R should have SL
	for i := 0; i < 4; i++ {
		if !strings.Contains(svg, "SL["+itoa(i)+"]") {
			t.Errorf("Missing SL[%d] label", i)
		}
	}

	// Check wire class
	if !strings.Contains(svg, "class=\"wire-sl\"") {
		t.Error("Missing SL wire class")
	}
}

func TestGenerateLayoutSVG_1T1RCells(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2
	arrayCfg.Architecture = "1t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check for 1T1R cell class
	if !strings.Contains(svg, "cell-1t1r") {
		t.Error("Missing cell-1t1r class")
	}

	// 1T1R cells have transistor + FeFET, so should have cell-transistor
	if !strings.Contains(svg, "cell-transistor") {
		t.Error("Missing cell-transistor class for 1T1R")
	}
}

func TestGenerateLayoutSVG_1T1RDimensions(t *testing.T) {
	passiveCfg := config.DefaultArrayConfig()
	passiveCfg.Rows = 4
	passiveCfg.Cols = 4
	passiveCfg.Architecture = "passive"

	cfg1T1R := config.DefaultArrayConfig()
	cfg1T1R.Rows = 4
	cfg1T1R.Cols = 4
	cfg1T1R.Architecture = "1t1r"

	svgPassive := GenerateLayoutSVG(passiveCfg, DefaultSVGConfig())
	svg1T1R := GenerateLayoutSVG(cfg1T1R, DefaultSVGConfig())

	// 1T1R should have larger height to accommodate SL labels
	// Extract viewBox height (rough check - 1T1R adds 30px for SL labels)
	if !strings.Contains(svg1T1R, "height=") {
		t.Error("Missing height attribute in 1T1R SVG")
	}

	// Both should have viewBox
	if !strings.Contains(svgPassive, "viewBox") || !strings.Contains(svg1T1R, "viewBox") {
		t.Error("Missing viewBox attribute")
	}
}

// ============================================================================
// SVG Configuration Options Tests
// ============================================================================

func TestGenerateLayoutSVG_ShowGrid(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4

	// With grid
	svgCfg := DefaultSVGConfig()
	svgCfg.ShowGrid = true
	svgWithGrid := GenerateLayoutSVG(arrayCfg, svgCfg)

	// Without grid
	svgCfg.ShowGrid = false
	svgNoGrid := GenerateLayoutSVG(arrayCfg, svgCfg)

	if !strings.Contains(svgWithGrid, "<!-- Grid -->") {
		t.Error("SVG with ShowGrid=true should have grid section")
	}
	if strings.Contains(svgNoGrid, "<!-- Grid -->") {
		t.Error("SVG with ShowGrid=false should not have grid section")
	}
}

func TestGenerateLayoutSVG_ShowLabels(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2

	// With labels
	svgCfg := DefaultSVGConfig()
	svgCfg.ShowLabels = true
	svgWithLabels := GenerateLayoutSVG(arrayCfg, svgCfg)

	// Without labels
	svgCfg.ShowLabels = false
	svgNoLabels := GenerateLayoutSVG(arrayCfg, svgCfg)

	// With labels should have WL/BL text
	if !strings.Contains(svgWithLabels, "WL[0]</text>") {
		t.Error("SVG with ShowLabels=true should have WL text labels")
	}

	// Without labels - WL/BL text should not appear in label format
	if strings.Contains(svgNoLabels, "WL[0]</text>") {
		t.Error("SVG with ShowLabels=false should not have WL text labels")
	}
}

func TestGenerateLayoutSVG_ShowCellIDs(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2

	// With cell IDs
	svgCfg := DefaultSVGConfig()
	svgCfg.ShowCellIDs = true
	svgWithIDs := GenerateLayoutSVG(arrayCfg, svgCfg)

	// Without cell IDs
	svgCfg.ShowCellIDs = false
	svgNoIDs := GenerateLayoutSVG(arrayCfg, svgCfg)

	// With IDs should have coordinate labels like "0,0", "0,1", etc.
	if !strings.Contains(svgWithIDs, "0,0</text>") {
		t.Error("SVG with ShowCellIDs=true should have cell ID labels")
	}
	if !strings.Contains(svgWithIDs, "class=\"label-small\"") {
		t.Error("Cell ID labels should use label-small class")
	}

	// Without IDs
	if strings.Contains(svgNoIDs, "0,0</text>") {
		t.Error("SVG with ShowCellIDs=false should not have cell ID labels")
	}
}

func TestGenerateLayoutSVG_CustomDimensions(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4

	svgCfg := SVGConfig{
		CellWidth:   100, // Larger cells
		CellHeight:  150,
		Margin:      50,
		ShowGrid:    true,
		ShowLabels:  true,
		ColorScheme: "default",
	}

	svg := GenerateLayoutSVG(arrayCfg, svgCfg)

	// SVG should be generated without error
	if !strings.Contains(svg, "<svg") {
		t.Error("Failed to generate SVG with custom dimensions")
	}

	// Dimensions should affect viewBox
	if !strings.Contains(svg, "viewBox=") {
		t.Error("Missing viewBox")
	}
}

// ============================================================================
// GenerateLayoutSVGWithDefaults Tests
// ============================================================================

func TestGenerateLayoutSVGWithDefaults(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4

	svg := GenerateLayoutSVGWithDefaults(arrayCfg)

	// Should produce valid SVG
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Error("GenerateLayoutSVGWithDefaults should produce valid SVG")
	}

	// Should use defaults (show grid, show labels, no cell IDs)
	if !strings.Contains(svg, "<!-- Grid -->") {
		t.Error("Defaults should include grid")
	}
}

// ============================================================================
// SVG Structure Tests
// ============================================================================

func TestGenerateLayoutSVG_StyleDefinitions(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check style definitions exist
	if !strings.Contains(svg, "<style>") {
		t.Error("Missing style definitions")
	}

	// Check key CSS classes are defined
	requiredClasses := []string{
		".cell",
		".cell-passive",
		".cell-1t1r",
		".wire-wl",
		".wire-bl",
		".label",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(svg, class+" {") && !strings.Contains(svg, class+"{") {
			t.Errorf("Missing CSS class definition: %s", class)
		}
	}
}

func TestGenerateLayoutSVG_Legend(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "passive"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check legend exists
	if !strings.Contains(svg, "<!-- Legend -->") {
		t.Error("Missing legend section")
	}

	// Legend should have WL and BL labels
	if !strings.Contains(svg, "WL (Word Line)") {
		t.Error("Legend missing WL description")
	}
	if !strings.Contains(svg, "BL (Bit Line)") {
		t.Error("Legend missing BL description")
	}
}

func TestGenerateLayoutSVG_1T1RLegend(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "1t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// 1T1R legend should include SL
	if !strings.Contains(svg, "SL (Source Line)") {
		t.Error("1T1R legend missing SL description")
	}
}

func TestGenerateLayoutSVG_Background(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Check background
	if !strings.Contains(svg, "<!-- Background -->") {
		t.Error("Missing background section")
	}
	if !strings.Contains(svg, "fill=\"#0a1520\"") {
		t.Error("Missing dark background color")
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestGenerateLayoutSVG_SingleCell(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 1
	arrayCfg.Cols = 1

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	if !strings.Contains(svg, "1x1 Crossbar") {
		t.Error("Missing 1x1 in title")
	}

	// Should have exactly 1 WL and 1 BL label
	wlCount := strings.Count(svg, "WL[0]")
	if wlCount < 1 {
		t.Error("Missing WL[0] for single cell")
	}
}

func TestGenerateLayoutSVG_LargeArray(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 32
	arrayCfg.Cols = 32

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Should generate without error
	if !strings.Contains(svg, "32x32 Crossbar") {
		t.Error("Missing 32x32 in title")
	}

	// Should have WL[31] and BL[31]
	if !strings.Contains(svg, "WL[31]") {
		t.Error("Missing WL[31] for large array")
	}
	if !strings.Contains(svg, "BL[31]") {
		t.Error("Missing BL[31] for large array")
	}
}

func TestGenerateLayoutSVG_RectangularArray(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 8
	arrayCfg.Cols = 4

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	if !strings.Contains(svg, "8x4 Crossbar") {
		t.Error("Missing 8x4 in title")
	}

	// Should have WL[7] but not WL[8]
	if !strings.Contains(svg, "WL[7]") {
		t.Error("Missing WL[7]")
	}
	if strings.Contains(svg, "WL[8]") {
		t.Error("Should not have WL[8] for 8-row array")
	}
}

// ============================================================================
// GenerateLayoutSVG Tests - 2T1R Architecture
// ============================================================================

func TestGenerateLayoutSVG_2T1RTitle(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "2t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	if !strings.Contains(svg, "FeCIM 4x4 Crossbar (2T1R)") {
		t.Error("Missing or incorrect 2T1R title")
	}
}

func TestGenerateLayoutSVG_2T1RCSLLines(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "2t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// 2T1R should have CSL wires
	if !strings.Contains(svg, "class=\"wire-csl\"") {
		t.Error("2T1R architecture should have CSL wire class")
	}
	// CSL labels
	for i := 0; i < 4; i++ {
		if !strings.Contains(svg, "CSL["+itoa(i)+"]") {
			t.Errorf("Missing CSL[%d] label", i)
		}
	}
}

func TestGenerateLayoutSVG_2T1RHasBothSLAndCSL(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "2t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// 2T1R has both SL (source line, per-column) and CSL (column select, per-column).
	// Both should be rendered in the SVG.
	if !strings.Contains(svg, ">SL[") {
		t.Error("2T1R architecture should have SL[] text labels (SL is a 2T1R port)")
	}
	if !strings.Contains(svg, "class=\"wire-sl\"") {
		t.Error("2T1R architecture should have wire-sl elements (SL is a 2T1R port)")
	}
	if !strings.Contains(svg, ">CSL[") {
		t.Error("2T1R architecture should have CSL[] text labels")
	}
	if !strings.Contains(svg, "class=\"wire-csl\"") {
		t.Error("2T1R architecture should have wire-csl elements")
	}
}

func TestGenerateLayoutSVG_2T1RCells(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2
	arrayCfg.Architecture = "2t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// 2T1R cells should have both transistor classes
	if !strings.Contains(svg, "cell-2t1r") {
		t.Error("Missing cell-2t1r class")
	}
	if !strings.Contains(svg, "cell-transistor") {
		t.Error("Missing cell-transistor (row transistor) in 2T1R cell")
	}
	if !strings.Contains(svg, "cell-transistor2") {
		t.Error("Missing cell-transistor2 (column transistor) in 2T1R cell")
	}
}

func TestGenerateLayoutSVG_2T1RLegend(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 4
	arrayCfg.Architecture = "2t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	if !strings.Contains(svg, "CSL (Col Select)") {
		t.Error("2T1R legend missing CSL description")
	}
	// 2T1R has both SL and CSL ports — SL should appear in legend too
	if !strings.Contains(svg, "SL (Source Line)") {
		t.Error("2T1R legend missing SL description (2T1R has SL per column)")
	}
}

func TestGenerateLayoutSVG_2T1RStyleCSS(t *testing.T) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 2
	arrayCfg.Cols = 2

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// CSS classes for 2T1R should be defined in all SVGs (in the style block)
	if !strings.Contains(svg, ".cell-2t1r") {
		t.Error("Missing .cell-2t1r CSS class definition")
	}
	if !strings.Contains(svg, ".wire-csl") {
		t.Error("Missing .wire-csl CSS class definition")
	}
}

func TestGenerateLayoutSVG_2T1RCSLPerColumn(t *testing.T) {
	// Non-square array: 4 rows × 3 cols.
	// CSL is per-column (Verilog: input wire [Cols-1:0] CSL), so there must be
	// exactly numCols=3 CSL labels, not numRows=4.
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 4
	arrayCfg.Cols = 3
	arrayCfg.Architecture = "2t1r"

	svg := GenerateLayoutSVG(arrayCfg, DefaultSVGConfig())

	// Should have CSL[0..2] (one per column)
	for i := 0; i < 3; i++ {
		if !strings.Contains(svg, "CSL["+itoa(i)+"]") {
			t.Errorf("Missing CSL[%d] for 4x3 2T1R array", i)
		}
	}
	// Must NOT have CSL[3] (that would mean wrong per-row indexing)
	if strings.Contains(svg, "CSL[3]") {
		t.Error("CSL[3] should not exist for 3-column 2T1R array (CSL is per-column)")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGenerateLayoutSVG_Small(b *testing.B) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 8
	arrayCfg.Cols = 8
	svgCfg := DefaultSVGConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateLayoutSVG(arrayCfg, svgCfg)
	}
}

func BenchmarkGenerateLayoutSVG_Large(b *testing.B) {
	arrayCfg := config.DefaultArrayConfig()
	arrayCfg.Rows = 64
	arrayCfg.Cols = 64
	svgCfg := DefaultSVGConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateLayoutSVG(arrayCfg, svgCfg)
	}
}
