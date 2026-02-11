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
		CellWidth:    0.92,
		CellHeight:   4.07,
	}

	content := MakeBuilderValidationTab(cfg, window)
	if content == nil {
		t.Fatal("MakeBuilderValidationTab returned nil for 2T1R architecture")
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
