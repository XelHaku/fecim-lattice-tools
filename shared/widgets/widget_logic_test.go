package widgets

import (
	"testing"
)

// ============================================================================
// MISSING FORMATTING FUNCTIONS (not in material_picker_test.go)
// ============================================================================

func TestFormatEnergy(t *testing.T) {
	tests := []struct {
		name     string
		ev       float64
		expected string
	}{
		{"Zero", 0, "0.00 eV"},
		{"Small", 0.1, "0.10 eV"},
		{"Typical activation", 0.65, "0.65 eV"},
		{"Large", 2.5, "2.50 eV"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEnergy(tt.ev)
			if result != tt.expected {
				t.Errorf("FormatEnergy(%v) = %q, want %q", tt.ev, result, tt.expected)
			}
		})
	}
}

func TestFormatConductanceRatio(t *testing.T) {
	tests := []struct {
		name     string
		ratio    float64
		expected string
	}{
		{"Zero", 0, "N/A"},
		{"Negative", -1, "N/A"},
		{"Small", 10, "10:1"},
		{"Hundreds", 500, "500:1"},
		{"Thousands", 5000, "5k:1"},
		{"Ten thousands", 50000, "5×10^4:1"},
		{"Hundred thousands", 200000, ">10^5:1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatConductanceRatio(tt.ratio)
			if result != tt.expected {
				t.Errorf("FormatConductanceRatio(%v) = %q, want %q", tt.ratio, result, tt.expected)
			}
		})
	}
}

func TestFormatVoltage(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		expected string
	}{
		{"Zero", 0, "0 mV"},
		{"Millivolt", 0.1, "100 mV"},
		{"Sub-volt", 0.5, "500 mV"},
		{"One volt", 1.0, "1.0 V"},
		{"Multiple volts", 3.3, "3.3 V"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatVoltage(tt.v)
			if result != tt.expected {
				t.Errorf("FormatVoltage(%v) = %q, want %q", tt.v, result, tt.expected)
			}
		})
	}
}

func TestFormatDimensionless(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		expected string
	}{
		{"Zero", 0, "0"},
		{"Integer", 5, "5"},
		{"Decimal", 3.14, "3.14"},
		{"Small decimal", 0.05, "0.05"},
		{"Large decimal", 123.456, "123.46"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDimensionless(tt.v)
			if result != tt.expected {
				t.Errorf("FormatDimensionless(%v) = %q, want %q", tt.v, result, tt.expected)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		expected string
	}{
		{"Zero", 0, "0.0%"},
		{"Half", 0.5, "50.0%"},
		{"Full", 1.0, "100.0%"},
		{"Decimal", 0.873, "87.3%"},
		{"Over 100", 1.5, "150.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercent(tt.v)
			if result != tt.expected {
				t.Errorf("FormatPercent(%v) = %q, want %q", tt.v, result, tt.expected)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxWidth int
		expected string
	}{
		{"Empty string", "", 10, ""},
		{"Short string", "Hello", 10, "Hello"},
		{"Single long word", "Supercalifragilistic", 10, "Supercalifragilistic"},
		{"Multiple words wrap", "Hello world from tests", 10, "Hello\nworld from\ntests"},
		{"Exact width", "Hello", 5, "Hello"},
		{"MaxWidth 0", "Hello", 0, "Hello"},
		{"Negative maxWidth", "Hello", -1, "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.s, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("WrapText(%q, %d) = %q, want %q", tt.s, tt.maxWidth, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// MISSING CATEGORY FUNCTIONS (not in material_picker_test.go)
// ============================================================================

func TestHasCategory(t *testing.T) {
	props := []FormattedProperty{
		{Name: "Pr", Category: CategoryCore},
		{Name: "d", Category: CategoryGeometry},
	}

	if !HasCategory(props, CategoryCore) {
		t.Error("HasCategory(Core) = false, want true")
	}

	if !HasCategory(props, CategoryGeometry) {
		t.Error("HasCategory(Geometry) = false, want true")
	}

	if HasCategory(props, CategoryLandau) {
		t.Error("HasCategory(Landau) = true, want false")
	}
}

// ============================================================================
// ARCHITECTURE SELECTOR (architecture_selector.go)
// ============================================================================

func TestArchitectureInfo(t *testing.T) {
	tests := []struct {
		name string
		arch string
	}{
		{"1T1R", Architecture1T1R},
		{"0T1R", Architecture0T1R},
		{"2T1R", Architecture2T1R},
		{"1T1R short", "1T1R"},
		{"Empty defaults to 1T1R", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, content := ArchitectureInfo(tt.arch)
			if title == "" {
				t.Error("ArchitectureInfo returned empty title")
			}
			if content == "" {
				t.Error("ArchitectureInfo returned empty content")
			}
		})
	}
}

func TestArchitectureSelectorConstants(t *testing.T) {
	// Verify constants are set correctly
	if Architecture1T1R != "1T1R (Transistor)" {
		t.Errorf("Architecture1T1R = %q, want %q", Architecture1T1R, "1T1R (Transistor)")
	}
	if Architecture0T1R != "0T1R (Passive)" {
		t.Errorf("Architecture0T1R = %q, want %q", Architecture0T1R, "0T1R (Passive)")
	}
	if Architecture2T1R != "2T1R (Dual Transistor)" {
		t.Errorf("Architecture2T1R = %q, want %q", Architecture2T1R, "2T1R (Dual Transistor)")
	}
}

// ============================================================================
// RESPONSIVE GRID LAYOUT (responsive_grid_layout.go) - Not in resize_detector_test.go
// ============================================================================

func TestResponsiveGridLayout_getMaxColumns(t *testing.T) {
	layout := NewResponsiveGridLayout()

	tests := []struct {
		name     string
		width    float32
		expected int
	}{
		{"Mobile", 500, 1},
		{"Tablet", 700, 1},
		{"Small laptop", 900, 1},
		{"Desktop", 1200, 2},
		{"Ultra-wide", 1800, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := layout.getMaxColumns(tt.width)
			if result != tt.expected {
				t.Errorf("getMaxColumns(%v) = %d, want %d", tt.width, result, tt.expected)
			}
		})
	}
}

func TestResponsiveGridLayout_getBreakpointHeight(t *testing.T) {
	layout := NewResponsiveGridLayout()

	tests := []struct {
		name     string
		width    float32
		expected float32
	}{
		{"SM", 500, layout.ItemHeightSM},
		{"MD", 700, layout.ItemHeightMD},
		{"LG", 900, layout.ItemHeightLG},
		{"XL", 1200, layout.ItemHeightXL},
		{"XXL", 1800, layout.ItemHeightXL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := layout.getBreakpointHeight(tt.width)
			if result != tt.expected {
				t.Errorf("getBreakpointHeight(%v) = %v, want %v", tt.width, result, tt.expected)
			}
		})
	}
}

func TestResponsiveGridLayout_calculateLayout(t *testing.T) {
	layout := NewResponsiveGridLayout()

	tests := []struct {
		name       string
		width      float32
		expectCols int
		maxWidth   float32 // item width should not exceed MaxItemWidth
	}{
		{"Mobile single column", 400, 1, 550},
		{"Desktop two columns", 1200, 2, 550},    // Capped at MaxItemWidth
		{"Ultra-wide three columns", 1800, 3, 550}, // Capped at MaxItemWidth
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, itemWidth, _, _ := layout.calculateLayout(tt.width)
			if cols != tt.expectCols {
				t.Errorf("calculateLayout(%v) cols = %d, want %d", tt.width, cols, tt.expectCols)
			}
			if itemWidth > tt.maxWidth {
				t.Errorf("calculateLayout(%v) itemWidth = %v, exceeds max %v", tt.width, itemWidth, tt.maxWidth)
			}
			if itemWidth <= 0 {
				t.Errorf("calculateLayout(%v) itemWidth = %v, should be > 0", tt.width, itemWidth)
			}
		})
	}
}

func TestResponsiveGridLayoutDefaults(t *testing.T) {
	layout := NewResponsiveGridLayout()

	// Check default values are set
	if layout.MinItemWidth <= 0 {
		t.Error("MinItemWidth should be > 0")
	}
	if layout.MaxItemWidth <= 0 {
		t.Error("MaxItemWidth should be > 0")
	}
	if layout.ItemHeightSM <= 0 {
		t.Error("ItemHeightSM should be > 0")
	}
	if layout.ItemHeightMD <= 0 {
		t.Error("ItemHeightMD should be > 0")
	}
	if layout.ItemHeightLG <= 0 {
		t.Error("ItemHeightLG should be > 0")
	}
	if layout.ItemHeightXL <= 0 {
		t.Error("ItemHeightXL should be > 0")
	}
}

// ============================================================================
// UI LOCK (ui_lock.go)
// ============================================================================

func TestUILockReentrant(t *testing.T) {
	// Test re-entrant lock from same goroutine
	lockUI()
	defer unlockUI()

	// Should not deadlock
	lockUI()
	defer unlockUI()

	// Verify we can execute logic while holding nested locks
	result := 42
	if result != 42 {
		t.Error("Failed to execute under nested locks")
	}
}

func TestWithUILock(t *testing.T) {
	executed := false
	WithUILock(func() {
		executed = true
	})

	if !executed {
		t.Error("WithUILock did not execute function")
	}
}

func TestGoroutineID(t *testing.T) {
	// Test that goroutineID returns a non-zero value
	id := goroutineID()
	if id == 0 {
		t.Error("goroutineID() returned 0, expected non-zero")
	}

	// Test that same goroutine has consistent ID
	id2 := goroutineID()
	if id != id2 {
		t.Errorf("goroutineID() inconsistent: first=%d, second=%d", id, id2)
	}
}

// ============================================================================
// SAFE_DO (safe_do.go)
// ============================================================================

func TestSafeDo(t *testing.T) {
	executed := false
	SafeDo(func() {
		executed = true
	})

	if !executed {
		t.Error("SafeDo did not execute function")
	}
}

func TestSafeDoNested(t *testing.T) {
	outer := false
	inner := false

	SafeDo(func() {
		outer = true
		SafeDo(func() {
			inner = true
		})
	})

	if !outer {
		t.Error("SafeDo outer function not executed")
	}
	if !inner {
		t.Error("SafeDo inner function not executed")
	}
}

// ============================================================================
// LAYOUT HELPERS (layout_helpers.go)
// ============================================================================

func TestCreateSectionDividerDefault(t *testing.T) {
	divider := CreateSectionDividerDefault()
	if divider == nil {
		t.Error("CreateSectionDividerDefault returned nil")
	}
}

// ============================================================================
// ADAPTIVE LAYOUT (adaptive_layout.go)
// ============================================================================

func TestGridWrapLayoutDefaults(t *testing.T) {
	layout := NewGridWrapLayout(300, 150, 10, 10)

	if layout.MinItemWidth != 300 {
		t.Errorf("MinItemWidth = %v, want 300", layout.MinItemWidth)
	}
	if layout.ItemHeight != 150 {
		t.Errorf("ItemHeight = %v, want 150", layout.ItemHeight)
	}
	if layout.RowSpacing != 10 {
		t.Errorf("RowSpacing = %v, want 10", layout.RowSpacing)
	}
	if layout.ColSpacing != 10 {
		t.Errorf("ColSpacing = %v, want 10", layout.ColSpacing)
	}
}

// ============================================================================
// STATUS BAR (status_helper.go)
// ============================================================================

func TestStatusBarGetText(t *testing.T) {
	sb := NewStatusBar("Status: ")
	sb.Update("Testing")

	text := sb.GetText()
	if text != "Testing" {
		t.Errorf("GetText() = %q, want %q", text, "Testing")
	}
}

func TestStatusBarGetTextEmpty(t *testing.T) {
	sb := NewStatusBar("")
	sb.Update("Ready")

	text := sb.GetText()
	if text != "Ready" {
		t.Errorf("GetText() = %q, want %q", text, "Ready")
	}
}

// ============================================================================
// CATEGORY ORDER (material_format.go)
// ============================================================================

func TestCategoryOrderDefined(t *testing.T) {
	if len(CategoryOrder) == 0 {
		t.Error("CategoryOrder should not be empty")
	}

	// Verify all expected categories are present
	expectedCategories := []string{
		CategoryCore,
		CategoryGeometry,
		CategoryLandau,
		CategoryAlpha,
		CategoryDepol,
		CategoryCircuit,
		CategoryNLS,
		CategoryConductance,
	}

	for _, expected := range expectedCategories {
		found := false
		for _, cat := range CategoryOrder {
			if cat == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CategoryOrder missing expected category: %s", expected)
		}
	}
}
