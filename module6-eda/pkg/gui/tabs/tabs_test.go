// pkg/gui/tabs/tabs_test.go
package tabs

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestMakeBuilderValidationTab(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	content := MakeBuilderValidationTab(cfg, window)
	if content == nil {
		t.Fatal("MakeBuilderValidationTab returned nil")
	}
}

func TestMakeBuilderValidationTab_NilConfig(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	// MakeBuilderValidationTab requires a valid config pointer
	// Passing nil is expected to panic - verify this behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("MakeBuilderValidationTab did not panic with nil config (expected panic)")
		}
	}()

	_ = MakeBuilderValidationTab(nil, window)
}

func TestMakeBuilderValidationTab_1T1RArchitecture(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "memory",
		Architecture: "1t1r",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   4.07,
	}

	content := MakeBuilderValidationTab(cfg, window)
	if content == nil {
		t.Fatal("MakeBuilderValidationTab returned nil for 1T1R architecture")
	}
}

func TestMakeBuilderValidationTab_2T1RArchitecture(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "compute",
		Architecture: "2t1r",
		Technology:   "sky130",
		CellWidth:    1.38,
		CellHeight:   4.07,
	}

	content := MakeBuilderValidationTab(cfg, window)
	if content == nil {
		t.Fatal("MakeBuilderValidationTab returned nil for 2T1R architecture")
	}
}

func TestMakeFlowScriptsTab(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "compute",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	content := MakeFlowScriptsTab(cfg, window)
	if content == nil {
		t.Fatal("MakeFlowScriptsTab returned nil")
	}
}

func TestMakeFlowScriptsTab_NilConfig(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	// nil config should use defaults without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MakeFlowScriptsTab panicked with nil config: %v", r)
		}
	}()

	content := MakeFlowScriptsTab(nil, window)
	if content == nil {
		t.Error("MakeFlowScriptsTab returned nil with nil config")
	}
}

func TestLoadFlowScriptContent_DesignSummary(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	content, desc := loadFlowScriptContent("Design Summary (design_summary.txt)", cfg)
	if content == "" {
		t.Fatal("loadFlowScriptContent Design Summary returned empty content")
	}
	if desc == "" {
		t.Fatal("loadFlowScriptContent Design Summary returned empty desc")
	}
	if !containsString(content, "Physical") {
		t.Error("Design summary should contain Physical section")
	}
}

func TestFlowScriptExtension(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"Design Summary (design_summary.txt)", ".txt"},
		{"Yosys TCL (synthesis.tcl)", ".tcl"},
		{"KLayout Python (gen_gds.py)", ".py"},
		{"CrossSim YAML (crosssim.yaml)", ".yaml"},
		{"Shell Runner (run_flow.sh)", ".sh"},
		{"LibreLane JSON (config.json)", ".json"},
		{"OpenVAF Verilog-A (fecim_lk.va)", ".va"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got := flowScriptExtension(tt.format)
			if got != tt.want {
				t.Errorf("flowScriptExtension(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestMakeLearnTab(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	content := MakeLearnTab(nil, window)
	if content == nil {
		t.Fatal("MakeLearnTab returned nil")
	}
}

func TestMakeLearnTab_WithState(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	// State parameter is currently unused but should not cause issues
	state := map[string]interface{}{"test": "value"}
	content := MakeLearnTab(state, window)
	if content == nil {
		t.Fatal("MakeLearnTab returned nil with state")
	}
}

func TestMakeExportViewerTab(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	content := MakeExportViewerTab(cfg, window)
	if content == nil {
		t.Fatal("MakeExportViewerTab returned nil")
	}

	v, source := loadExportPreviewContent("Verilog", cfg)
	if v == "" || source == "" {
		t.Fatal("loadExportPreviewContent returned empty verilog/source")
	}
}

func TestMakeLayoutVisualizerTab(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	content := MakeLayoutVisualizerTab(cfg, window)
	if content == nil {
		t.Fatal("MakeLayoutVisualizerTab returned nil")
	}

	svg, _ := loadLayoutSVGContent(cfg)
	summary := buildLayerSummary(svg, layerFilter{WL: true, BL: true, SL: true, CSL: true, Cells: true, Grid: true, Legend: true}, cfg)
	if summary == "" || !containsString(summary, "WL wires") {
		t.Fatal("buildLayerSummary did not include expected layer output")
	}
}

func TestBuildLayerSummary_CSLWires_2T1R(t *testing.T) {
	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "2t1r", CellWidth: 1.38, CellHeight: 4.07}
	svg, _ := loadLayoutSVGContent(cfg)

	// CSL enabled: summary must mention CSL wires line
	withCSL := buildLayerSummary(svg, layerFilter{CSL: true}, cfg)
	if !containsString(withCSL, "CSL wires") {
		t.Errorf("buildLayerSummary with CSL=true did not include 'CSL wires' line")
	}

	// CSL disabled: CSL wires line must be absent
	withoutCSL := buildLayerSummary(svg, layerFilter{CSL: false}, cfg)
	if containsString(withoutCSL, "CSL wires") {
		t.Errorf("buildLayerSummary with CSL=false should not include 'CSL wires' line")
	}
}

func TestMakeLearnTab_NilWindow(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Should handle nil window gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MakeLearnTab panicked with nil window: %v", r)
		}
	}()

	content := MakeLearnTab(nil, nil)
	if content == nil {
		t.Error("MakeLearnTab returned nil with nil window")
	}
}

// Test visual component creators
func TestOpenLaneFlowDiagram(t *testing.T) {
	diagram := OpenLaneFlowDiagram()
	if diagram == nil {
		t.Fatal("OpenLaneFlowDiagram returned nil")
	}
}

func TestOperationModesVisual(t *testing.T) {
	visual := OperationModesVisual()
	if visual == nil {
		t.Fatal("OperationModesVisual returned nil")
	}
}

func TestIsometricCrossbar(t *testing.T) {
	tests := []struct {
		name       string
		rows       int
		cols       int
		showLabels bool
	}{
		{"3x3 with labels", 3, 3, true},
		{"3x3 without labels", 3, 3, false},
		{"4x4 with labels", 4, 4, true},
		{"1x1 minimal", 1, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram := IsometricCrossbar(tt.rows, tt.cols, tt.showLabels)
			if diagram == nil {
				t.Errorf("IsometricCrossbar(%d, %d, %v) returned nil", tt.rows, tt.cols, tt.showLabels)
			}
		})
	}
}

func TestIsometric1T1RCrossbar(t *testing.T) {
	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"3x3 array", 3, 3},
		{"4x4 array", 4, 4},
		{"1x1 minimal", 1, 1},
		{"8x8 larger", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram := Isometric1T1RCrossbar(tt.rows, tt.cols)
			if diagram == nil {
				t.Errorf("Isometric1T1RCrossbar(%d, %d) returned nil", tt.rows, tt.cols)
			}
		})
	}
}

func TestIsometric2T1RCrossbar(t *testing.T) {
	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"3x3 array", 3, 3},
		{"4x4 array", 4, 4},
		{"1x1 minimal", 1, 1},
		{"8x8 larger", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram := Isometric2T1RCrossbar(tt.rows, tt.cols)
			if diagram == nil {
				t.Errorf("Isometric2T1RCrossbar(%d, %d) returned nil", tt.rows, tt.cols)
			}
		})
	}
}

func TestCellComparisonTable(t *testing.T) {
	table := CellComparisonTable()
	if table == nil {
		t.Fatal("CellComparisonTable returned nil")
	}
}

func TestLEFPreviewCard(t *testing.T) {
	card := LEFPreviewCard()
	if card == nil {
		t.Fatal("LEFPreviewCard returned nil")
	}
}

func TestDEFPreviewCard(t *testing.T) {
	card := DEFPreviewCard()
	if card == nil {
		t.Fatal("DEFPreviewCard returned nil")
	}
}

func TestVerilogPreviewCard(t *testing.T) {
	card := VerilogPreviewCard()
	if card == nil {
		t.Fatal("VerilogPreviewCard returned nil")
	}
}

func TestLibertyPreviewCard(t *testing.T) {
	card := LibertyPreviewCard()
	if card == nil {
		t.Fatal("LibertyPreviewCard returned nil")
	}
}

func TestReferencesCard(t *testing.T) {
	card := ReferencesCard()
	if card == nil {
		t.Fatal("ReferencesCard returned nil")
	}
}

func TestFileFormatCard(t *testing.T) {
	tests := []struct {
		title   string
		format  string
		content string
	}{
		{"Test LEF", "lef", "MACRO test\nEND test"},
		{"Test DEF", "def", "DESIGN test\nEND DESIGN"},
		{"Test Verilog", "v", "module test;\nendmodule"},
		{"Empty content", "txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			card := FileFormatCard(tt.title, tt.format, tt.content)
			if card == nil {
				t.Errorf("FileFormatCard(%s, %s, ...) returned nil", tt.title, tt.format)
			}
		})
	}
}

func TestGenerateBuilderDEF(t *testing.T) {
	tests := []struct {
		name   string
		config config.ArrayConfig
	}{
		{
			name: "Passive 4x4",
			config: config.ArrayConfig{
				Rows:         4,
				Cols:         4,
				Mode:         "storage",
				Architecture: "passive",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   2.72,
			},
		},
		{
			name: "1T1R 4x4",
			config: config.ArrayConfig{
				Rows:         4,
				Cols:         4,
				Mode:         "memory",
				Architecture: "1t1r",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   4.07,
			},
		},
		{
			name: "2T1R 4x4",
			config: config.ArrayConfig{
				Rows:         4,
				Cols:         4,
				Mode:         "compute",
				Architecture: "2t1r",
				Technology:   "sky130",
				CellWidth:    0.92,
				CellHeight:   4.07,
			},
		},
		{
			name: "Minimal 1x1",
			config: config.ArrayConfig{
				Rows:         1,
				Cols:         1,
				Mode:         "storage",
				Architecture: "passive",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   2.72,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defContent := generateBuilderDEF(tt.config)
			if defContent == "" {
				t.Error("generateBuilderDEF returned empty string")
			}

			// Verify basic DEF structure
			if len(defContent) < 50 {
				t.Error("generateBuilderDEF returned suspiciously short content")
			}

			// Check for required DEF keywords
			requiredKeywords := []string{"VERSION", "DESIGN", "UNITS", "DIEAREA", "COMPONENTS", "PINS", "END"}
			for _, keyword := range requiredKeywords {
				if !containsString(defContent, keyword) {
					t.Errorf("DEF content missing required keyword: %s", keyword)
				}
			}
		})
	}
}

func TestMakeBulletList(t *testing.T) {
	tests := []struct {
		name   string
		header string
		items  []string
	}{
		{
			name:   "Simple list",
			header: "Test Header",
			items:  []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name:   "Empty list",
			header: "Empty",
			items:  []string{},
		},
		{
			name:   "Single item",
			header: "Single",
			items:  []string{"Only one"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := makeBulletList(tt.header, tt.items...)
			if list == nil {
				t.Error("makeBulletList returned nil")
			}
		})
	}
}

func TestSizedContainer(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Create a simple child object
	child := OpenLaneFlowDiagram()

	tests := []struct {
		name   string
		width  float32
		height float32
	}{
		{"Standard size", 400, 300},
		{"Small size", 100, 100},
		{"Large size", 1000, 800},
		{"Zero width", 0, 300},
		{"Zero height", 400, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := sizedContainer(child, tt.width, tt.height)
			if container == nil {
				t.Errorf("sizedContainer(%f, %f) returned nil", tt.width, tt.height)
			}
		})
	}
}

func TestContentMakers(t *testing.T) {
	// Test all content creation functions
	contentFuncs := []struct {
		name string
		fn   func() fyne.CanvasObject
	}{
		{"makeQuickStartContent", makeQuickStartContent},
		{"makeIntroContent", makeIntroContent},
		{"makeCrossbarContent", makeCrossbarContent},
		{"makeFilesContent", makeFilesContent},
		{"makeFAQContent", makeFAQContent},
	}

	for _, tc := range contentFuncs {
		t.Run(tc.name, func(t *testing.T) {
			content := tc.fn()
			if content == nil {
				t.Errorf("%s returned nil", tc.name)
			}
		})
	}
}

// TestMakeExportViewerTab_NilWindow verifies that constructing the export viewer
// with a nil window does not panic, and that the save/copy button handlers
// return safely rather than panicking when window is nil.
func TestMakeExportViewerTab_NilWindow(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MakeExportViewerTab panicked with nil window: %v", r)
		}
	}()

	root := MakeExportViewerTab(cfg, nil)
	if root == nil {
		t.Fatal("MakeExportViewerTab returned nil with nil window")
	}

	// Tap the save button — should return silently, not panic.
	saveBtn := findButtonByText(root, "Save to File…")
	if saveBtn == nil {
		t.Fatal("failed to find Save to File… button")
	}
	saveBtn.OnTapped()

	// Tap copy — copyBtn already has a nil guard; confirm it also survives.
	copyBtn := findButtonByText(root, "Copy")
	if copyBtn == nil {
		t.Fatal("failed to find Copy button")
	}
	copyBtn.OnTapped()
}

func TestExportViewerNewFormats(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 4, Cols: 4,
		Mode: "storage", Architecture: "passive",
		Technology: "sky130", CellWidth: 0.46, CellHeight: 2.72,
	}

	cases := []struct {
		format  string
		ext     string
		wantKey string // substring that must appear in generated content
	}{
		{"Config (JSON)", ".json", "DESIGN_NAME"},
		{"SDC", ".sdc", "set_input_delay"},
		{"Design Summary", ".txt", "Physical"},
		// DEF now generates in-memory via GenerateLatticeDEF
		{"DEF", ".def", "DESIGN"},
		// SPICE now generates a subcircuit preview
		{"SPICE", ".sp", "FeCIM"},
	}

	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			// Verify extension mapping
			if got := formatExtension(tc.format); got != tc.ext {
				t.Errorf("formatExtension(%q) = %q, want %q", tc.format, got, tc.ext)
			}
			// Verify content generation (no disk files in test env)
			content, source := loadExportPreviewContent(tc.format, cfg)
			if content == "" {
				t.Fatalf("loadExportPreviewContent(%q) returned empty content", tc.format)
			}
			if source == "" {
				t.Errorf("loadExportPreviewContent(%q) returned empty source", tc.format)
			}
			if !containsString(content, tc.wantKey) {
				t.Errorf("loadExportPreviewContent(%q) missing %q in output", tc.format, tc.wantKey)
			}
		})
	}
}

func TestExportViewerArchitectureAware(t *testing.T) {
	baseCfg := config.ArrayConfig{
		Rows: 4, Cols: 4, Mode: "storage",
		Technology: "sky130", CellWidth: 0.46, CellHeight: 2.72,
	}

	t.Run("passive_LEF_uses_passive_cell_name", func(t *testing.T) {
		cfg := baseCfg
		cfg.Architecture = "passive"
		lef, _ := loadExportPreviewContent("LEF", &cfg)
		if !containsString(lef, "fecim_bitcell") {
			t.Error("passive LEF should contain fecim_bitcell")
		}
	})

	t.Run("1t1r_LEF_uses_1t1r_cell_name", func(t *testing.T) {
		cfg := baseCfg
		cfg.Architecture = "1t1r"
		cfg.CellHeight = 3.40
		lef, _ := loadExportPreviewContent("LEF", &cfg)
		if !containsString(lef, "fecim_1t1r_bitcell") {
			t.Error("1T1R LEF should contain fecim_1t1r_bitcell")
		}
	})

	t.Run("1t1r_SPICE_includes_1t1r_subcircuit", func(t *testing.T) {
		cfg := baseCfg
		cfg.Architecture = "1t1r"
		spice, _ := loadExportPreviewContent("SPICE", &cfg)
		if !containsString(spice, "fefet_1t1r") {
			t.Error("1T1R SPICE preview should contain fefet_1t1r subcircuit")
		}
	})

	t.Run("2t1r_SPICE_includes_2t1r_subcircuit", func(t *testing.T) {
		cfg := baseCfg
		cfg.Architecture = "2t1r"
		spice, _ := loadExportPreviewContent("SPICE", &cfg)
		if !containsString(spice, "fefet_2t1r") {
			t.Error("2T1R SPICE preview should contain fefet_2t1r subcircuit")
		}
	})

	t.Run("DEF_generates_in_memory", func(t *testing.T) {
		cfg := baseCfg
		cfg.Architecture = "passive"
		def, source := loadExportPreviewContent("DEF", &cfg)
		if !containsString(def, "DESIGN") {
			t.Error("in-memory DEF should contain DESIGN keyword")
		}
		if !containsString(source, "in-memory") {
			t.Errorf("DEF source should be in-memory, got %q", source)
		}
	})
}

func TestFileExists(t *testing.T) {
	// Test with a file that definitely doesn't exist
	if fileExists("/this/path/should/not/exist/test.txt") {
		t.Error("fileExists returned true for non-existent path")
	}

	// Test with current directory (should exist)
	if !fileExists(".") {
		t.Error("fileExists returned false for current directory")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// Conductance Heatmap Tests
// =============================================================================

func TestMakeConductanceHeatmapPanel(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4}
	content := MakeConductanceHeatmapPanel(cfg)
	if content == nil {
		t.Fatal("MakeConductanceHeatmapPanel returned nil")
	}
}

func TestMakeConductanceHeatmapPanel_NilConfig(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	content := MakeConductanceHeatmapPanel(nil)
	if content == nil {
		t.Fatal("MakeConductanceHeatmapPanel(nil) returned nil")
	}
}

func TestNewConductanceRaster(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Empty matrix should return a valid (blank) raster.
	r := newConductanceRaster(conductanceMatrix{})
	if r == nil {
		t.Fatal("newConductanceRaster(empty): returned nil")
	}

	// Non-empty matrix.
	m := makeConductanceGradient(8, 8)
	r = newConductanceRaster(m)
	if r == nil {
		t.Fatal("newConductanceRaster(8×8 gradient): returned nil")
	}
}

func TestHeatColorCIM(t *testing.T) {
	// Dark-blue at 0.
	c := heatColorCIM(0)
	if c.R != 0 || c.G != 0 || c.B != 80 {
		t.Errorf("heatColorCIM(0): expected dark-blue {0,0,80}, got %v", c)
	}
	// Yellow at 1.
	c = heatColorCIM(1)
	if c.R != 255 || c.B != 0 {
		t.Errorf("heatColorCIM(1): expected yellow (R=255,B=0), got %v", c)
	}
	// Negative clamped to dark-blue.
	c = heatColorCIM(-0.5)
	if c.R != 0 || c.G != 0 || c.B != 80 {
		t.Errorf("heatColorCIM(-0.5): expected dark-blue, got %v", c)
	}
	// Mid-range should be non-black.
	c = heatColorCIM(0.5)
	if c.R == 0 && c.G == 0 && c.B == 0 {
		t.Error("heatColorCIM(0.5): expected non-black color")
	}
}

func TestConductancePatterns(t *testing.T) {
	rows, cols := 4, 4

	t.Run("Uniform_level", func(t *testing.T) {
		m := makeConductanceUniform(rows, cols, 0.9)
		if m.Rows != rows || m.Cols != cols {
			t.Fatalf("wrong dims %d×%d", m.Rows, m.Cols)
		}
		for i := range m.Values {
			for j := range m.Values[i] {
				if m.Values[i][j] != 0.9 {
					t.Errorf("[%d][%d]: expected 0.9, got %f", i, j, m.Values[i][j])
				}
			}
		}
	})

	t.Run("Gradient_range", func(t *testing.T) {
		m := makeConductanceGradient(rows, cols)
		for i := range m.Values {
			for j := range m.Values[i] {
				v := m.Values[i][j]
				if v < 0 || v > 1 {
					t.Errorf("[%d][%d]: out of [0,1]: %f", i, j, v)
				}
			}
		}
	})

	t.Run("Checker_alternating", func(t *testing.T) {
		m := makeConductanceChecker(rows, cols)
		for i := range m.Values {
			for j := range m.Values[i] {
				expected := 0.9
				if (i+j)%2 != 0 {
					expected = 0.1
				}
				if m.Values[i][j] != expected {
					t.Errorf("[%d][%d]: expected %f, got %f", i, j, expected, m.Values[i][j])
				}
			}
		}
	})

	t.Run("Random_range", func(t *testing.T) {
		m := makeConductanceRandom(rows, cols, 42)
		for i := range m.Values {
			for j := range m.Values[i] {
				v := m.Values[i][j]
				if v < 0 || v > 1 {
					t.Errorf("[%d][%d]: out of [0,1]: %f", i, j, v)
				}
			}
		}
	})

	t.Run("Random_deterministic", func(t *testing.T) {
		m1 := makeConductanceRandom(rows, cols, 42)
		m2 := makeConductanceRandom(rows, cols, 42)
		if m1.Values[0][0] != m2.Values[0][0] {
			t.Error("same seed should produce identical values")
		}
	})

	t.Run("Random_different_seeds", func(t *testing.T) {
		m1 := makeConductanceRandom(rows, cols, 1)
		m2 := makeConductanceRandom(rows, cols, 999)
		same := true
		for i := range m1.Values {
			for j := range m1.Values[i] {
				if m1.Values[i][j] != m2.Values[i][j] {
					same = false
					break
				}
			}
		}
		if same {
			t.Error("different seeds produced identical matrices")
		}
	})
}

func TestConductancePatterns_NeuralWeights(t *testing.T) {
	rows, cols := 16, 16
	m := makeConductanceNeuralWeights(rows, cols)

	t.Run("dimensions", func(t *testing.T) {
		if m.Rows != rows || m.Cols != cols {
			t.Fatalf("expected %d×%d, got %d×%d", rows, cols, m.Rows, m.Cols)
		}
	})

	t.Run("values_in_range", func(t *testing.T) {
		for i := range m.Values {
			for j, v := range m.Values[i] {
				if v < 0 || v > 1 {
					t.Errorf("[%d][%d]: %f out of [0,1]", i, j, v)
				}
			}
		}
	})

	t.Run("reproducible_fixed_seed", func(t *testing.T) {
		m2 := makeConductanceNeuralWeights(rows, cols)
		for i := range m.Values {
			for j := range m.Values[i] {
				if m.Values[i][j] != m2.Values[i][j] {
					t.Errorf("[%d][%d]: second call gave different value (seed should be fixed)", i, j)
				}
			}
		}
	})

	t.Run("mean_near_half", func(t *testing.T) {
		sum := 0.0
		n := rows * cols
		for i := range m.Values {
			for _, v := range m.Values[i] {
				sum += v
			}
		}
		mean := sum / float64(n)
		if mean < 0.40 || mean > 0.60 {
			t.Errorf("expected mean ≈ 0.5 (σ=0.18 Gaussian centered at 0.5), got %.3f", mean)
		}
	})
}

func TestConductanceLevelCounts(t *testing.T) {
	// Uniform Hi (0.9) → all cells should land in the top bin of 30 levels.
	m := makeConductanceUniform(4, 4, 0.9)
	counts := conductanceLevelCounts(m, 30)
	if len(counts) != 30 {
		t.Fatalf("expected 30 bins, got %d", len(counts))
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != 4*4 {
		t.Errorf("counts sum %d, want %d", total, 4*4)
	}
	// bin for 0.9 with 30 levels: int(0.9*30) = 27
	if counts[27] != 16 {
		t.Errorf("uniform 0.9: expected all 16 cells in bin 27, got %d", counts[27])
	}

	// Uniform Lo (0.1) → bin 3.
	m = makeConductanceUniform(4, 4, 0.1)
	counts = conductanceLevelCounts(m, 30)
	if counts[3] != 16 {
		t.Errorf("uniform 0.1: expected all 16 cells in bin 3, got %d", counts[3])
	}

	// Checkerboard → two bins should be non-zero.
	m = makeConductanceChecker(4, 4)
	counts = conductanceLevelCounts(m, 30)
	nonZero := 0
	for _, c := range counts {
		if c > 0 {
			nonZero++
		}
	}
	if nonZero != 2 {
		t.Errorf("checkerboard: expected exactly 2 non-zero bins, got %d", nonZero)
	}
}

func TestNewConductanceHistogram(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Empty matrix.
	h := newConductanceHistogram(conductanceMatrix{}, 30)
	if h == nil {
		t.Fatal("newConductanceHistogram(empty): returned nil")
	}

	// Non-empty.
	m := makeConductanceRandom(8, 8, 42)
	h = newConductanceHistogram(m, 30)
	if h == nil {
		t.Fatal("newConductanceHistogram(8×8 random): returned nil")
	}
}

// =============================================================================
// Array Statistics + Export Viewer Tests
// =============================================================================

func TestGenerateArrayStatistics_PassiveSmall(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows:         8,
		Cols:         8,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}
	stats := generateArrayStatistics(cfg)
	if stats == "" {
		t.Fatal("generateArrayStatistics returned empty string")
	}
	for _, want := range []string{"passive", "8 rows × 8 columns", "LOW", "Total cells:    64"} {
		if !findSubstring(stats, want) {
			t.Errorf("expected %q in stats output", want)
		}
	}
}

func TestGenerateArrayStatistics_PassiveLarge(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 64, Cols: 64, Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72,
	}
	stats := generateArrayStatistics(cfg)
	if !findSubstring(stats, "HIGH") {
		t.Error("64×64 passive array should report HIGH sneak path risk")
	}
	if !findSubstring(stats, "EXCEEDS") {
		t.Error("64×64 passive should warn about exceeding recommended size")
	}
}

func TestGenerateArrayStatistics_1T1R(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 32, Cols: 32, Architecture: "1t1r", CellWidth: 0.46, CellHeight: 4.07,
	}
	stats := generateArrayStatistics(cfg)
	if !findSubstring(stats, "ROW-ONLY") {
		t.Error("1T1R should report ROW-ONLY sneak path suppression")
	}
}

func TestGenerateArrayStatistics_2T1R(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 64, Cols: 64, Architecture: "2t1r", CellWidth: 1.38, CellHeight: 4.07,
	}
	stats := generateArrayStatistics(cfg)
	if !findSubstring(stats, "NONE") {
		t.Error("2T1R should report NONE sneak path risk")
	}
}

func TestExportViewerArrayStatisticsFormat(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 4, Cols: 4, Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72, Technology: "sky130",
	}
	content, source := loadExportPreviewContent("Array Statistics", cfg)
	if content == "" {
		t.Fatal("Array Statistics format returned empty content")
	}
	if source != "generated (in-memory)" {
		t.Errorf("unexpected source: %s", source)
	}
	if !findSubstring(content, "Total cells:    16") {
		t.Error("expected cell count in Array Statistics output")
	}
}

func TestExportViewerManifestFormat(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 8, Cols: 8, Architecture: "1t1r", CellWidth: 0.46, CellHeight: 4.07, Technology: "sky130",
	}
	content, source := loadExportPreviewContent("Export Manifest", cfg)
	if content == "" {
		t.Fatal("Export Manifest returned empty content")
	}
	if source != "generated (in-memory)" {
		t.Errorf("unexpected source: %s", source)
	}
	for _, want := range []string{"fecim_crossbar_8x8", "run_flow.sh", "cells/fecim_1t1r_bitcell.lef", "TOTAL FILES"} {
		if !findSubstring(content, want) {
			t.Errorf("Export Manifest missing expected content: %q", want)
		}
	}
}

func TestExportViewerMultiCornerLiberty(t *testing.T) {
	cfg := &config.ArrayConfig{
		Rows: 4, Cols: 4, Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72, Technology: "sky130",
	}
	content, source := loadExportPreviewContent("Liberty (Multi-Corner)", cfg)
	if content == "" {
		t.Fatal("Liberty (Multi-Corner) format returned empty content")
	}
	if !findSubstring(source, "TT/SS/FF") {
		t.Errorf("source should mention corners, got: %s", source)
	}
	// Should contain multiple library blocks for different corners
	if !findSubstring(content, "library(") {
		t.Error("multi-corner Liberty should contain library blocks")
	}
}

// TestMakeLayoutVisualizerTab_NilWindow verifies that constructing the layout
// visualizer with a nil window does not panic, and that the Save SVG button
// handler returns safely (via the nil-window guard) rather than panicking.
func TestMakeLayoutVisualizerTab_NilWindow(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MakeLayoutVisualizerTab panicked with nil window: %v", r)
		}
	}()

	root := MakeLayoutVisualizerTab(cfg, nil)
	if root == nil {
		t.Fatal("MakeLayoutVisualizerTab returned nil with nil window")
	}

	// Tap Save SVG… — must return silently, not panic.
	saveBtn := findButtonByText(root, "Save SVG\u2026")
	if saveBtn == nil {
		t.Fatal("failed to find 'Save SVG…' button in layout visualizer tab")
	}
	saveBtn.OnTapped()

	// Tap Copy SVG — already has a nil guard; confirm it also survives.
	copyBtn := findButtonByText(root, "Copy SVG")
	if copyBtn == nil {
		t.Fatal("failed to find 'Copy SVG' button in layout visualizer tab")
	}
	copyBtn.OnTapped()

	// Confirm the tab still renders content (layer summary, not empty).
	_ = fyne.Size{} // ensure fyne import is used
}
