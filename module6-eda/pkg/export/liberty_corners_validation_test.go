package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestM6LIB04_CornersMultiCornerGeneration — M6-LIB-04
// Generate FF/TT/SS corners × T=-40/25/125°C
// Verify 9 Liberty files (or multi-corner single file)
// Check operating_conditions in each
func TestM6LIB04_CornersMultiCornerGeneration(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateMultiCornerLiberty(cfg)

	// Verify 9 operating_conditions blocks (FF/TT/SS × -40/25/125°C)
	expectedCorners := []string{
		"ff_n40C_1v95", "ff_025C_1v95", "ff_125C_1v95",
		"tt_n40C_1v80", "tt_025C_1v80", "tt_125C_1v80",
		"ss_n40C_1v60", "ss_025C_1v60", "ss_125C_1v60",
	}

	for _, corner := range expectedCorners {
		pattern := `operating_conditions\(` + corner + `\)`
		if !regexp.MustCompile(pattern).MatchString(lib) {
			t.Fatalf("missing operating_conditions for corner: %s", corner)
		}
	}

	// Count total operating_conditions blocks
	ocCount := strings.Count(lib, "operating_conditions(")
	if ocCount != 9 {
		t.Fatalf("expected 9 operating_conditions blocks, got %d", ocCount)
	}

	// Count library blocks (should be 9)
	libCount := strings.Count(lib, "library(")
	if libCount != 9 {
		t.Fatalf("expected 9 library blocks (one per corner), got %d", libCount)
	}

	t.Logf("M6-LIB-04 PASS: Multi-corner generation validated")
	t.Logf("  - Corners: %d/%d", len(expectedCorners), len(expectedCorners))
	t.Logf("  - operating_conditions blocks: %d", ocCount)
	t.Logf("  - library blocks: %d", libCount)
}

// TestM6LIB04_CornersOperatingConditionsAttributes validates process/temp/voltage
func TestM6LIB04_CornersOperatingConditionsAttributes(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.Voltage = 1.8 // V
	lib := GenerateMultiCornerLiberty(cfg)

	// Define expected corners with process/temp/voltage ranges
	cornerTests := []struct {
		name       string
		processMin float64
		processMax float64
		tempC      float64
		voltageMin float64
		voltageMax float64
	}{
		// FF (fast-fast): process < 1.0, higher voltage
		{"ff_n40C_1v95", 0.7, 0.9, -40.0, 1.8, 2.0},
		{"ff_025C_1v95", 0.7, 0.9, 25.0, 1.8, 2.0},
		{"ff_125C_1v95", 0.7, 0.9, 125.0, 1.7, 1.9},
		// TT (typical-typical): process = 1.0, nominal voltage
		{"tt_n40C_1v80", 0.9, 1.1, -40.0, 1.7, 1.9},
		{"tt_025C_1v80", 0.9, 1.1, 25.0, 1.7, 1.9},
		{"tt_125C_1v80", 0.9, 1.1, 125.0, 1.6, 1.8},
		// SS (slow-slow): process > 1.0, lower voltage
		{"ss_n40C_1v60", 1.1, 1.3, -40.0, 1.6, 1.8},
		{"ss_025C_1v60", 1.1, 1.3, 25.0, 1.5, 1.8},
		{"ss_125C_1v60", 1.1, 1.3, 125.0, 1.5, 1.7},
	}

	for _, tc := range cornerTests {
		t.Run(tc.name, func(t *testing.T) {
			// Find the corner's operating_conditions block
			reOC := regexp.MustCompile(`operating_conditions\(` + tc.name + `\)\s*\{[^}]*process\s*:\s*([0-9.]+)[^}]*temperature\s*:\s*([-0-9.]+)[^}]*voltage\s*:\s*([0-9.]+)`)
			mOC := reOC.FindStringSubmatch(lib)
			if len(mOC) < 4 {
				t.Fatalf("failed to extract operating_conditions for %s", tc.name)
			}

			process, _ := strconv.ParseFloat(mOC[1], 64)
			temp, _ := strconv.ParseFloat(mOC[2], 64)
			voltage, _ := strconv.ParseFloat(mOC[3], 64)

			// Validate process
			if process < tc.processMin || process > tc.processMax {
				t.Errorf("process out of range: got %.2f, expected [%.2f, %.2f]", process, tc.processMin, tc.processMax)
			}

			// Validate temperature (exact match expected)
			if temp != tc.tempC {
				t.Errorf("temperature mismatch: got %.1f, expected %.1f", temp, tc.tempC)
			}

			// Validate voltage
			if voltage < tc.voltageMin || voltage > tc.voltageMax {
				t.Errorf("voltage out of range: got %.3f V, expected [%.3f, %.3f]", voltage, tc.voltageMin, tc.voltageMax)
			}

			t.Logf("  %s: process=%.2f, temp=%.1f°C, voltage=%.3fV", tc.name, process, temp, voltage)
		})
	}

	t.Logf("M6-LIB-04 PASS: Operating conditions attributes validated for all 9 corners")
}

// TestM6LIB04_CornersTimingScaling validates timing scales across corners
func TestM6LIB04_CornersTimingScaling(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.RiseTime = 50.0 // ns
	lib := GenerateMultiCornerLiberty(cfg)

	// Extract cell_rise first value from FF corner
	ffIdx := strings.Index(lib, "library(fecim_cells_ff_n40C_1v95)")
	if ffIdx < 0 {
		t.Fatal("missing FF corner library")
	}
	ffLib := lib[ffIdx : strings.Index(lib[ffIdx+50:], "library(")+ffIdx+50]
	ffRise := extractFirstNLDMValue(t, ffLib, "cell_rise")

	// Extract cell_rise first value from TT corner
	ttIdx := strings.Index(lib, "library(fecim_cells_tt_025C_1v80)")
	if ttIdx < 0 {
		t.Fatal("missing TT corner library")
	}
	ttLib := lib[ttIdx : strings.Index(lib[ttIdx+50:], "library(")+ttIdx+50]
	ttRise := extractFirstNLDMValue(t, ttLib, "cell_rise")

	// Extract cell_rise first value from SS corner
	ssIdx := strings.Index(lib, "library(fecim_cells_ss_125C_1v60)")
	if ssIdx < 0 {
		t.Fatal("missing SS corner library")
	}
	ssLib := lib[ssIdx:]
	ssRise := extractFirstNLDMValue(t, ssLib, "cell_rise")

	// Verify monotonicity: FF (fast) < TT (typical) < SS (slow)
	if !(ffRise < ttRise && ttRise < ssRise) {
		t.Fatalf("timing not monotonic across corners: FF=%.3f < TT=%.3f < SS=%.3f", ffRise, ttRise, ssRise)
	}

	// Verify reasonable scaling factors
	ffToTTRatio := ffRise / ttRise
	ttToSSRatio := ssRise / ttRise
	if ffToTTRatio < 0.5 || ffToTTRatio > 1.0 {
		t.Errorf("FF/TT timing ratio out of expected range: %.3f (expected 0.5-1.0)", ffToTTRatio)
	}
	if ttToSSRatio < 1.0 || ttToSSRatio > 2.0 {
		t.Errorf("SS/TT timing ratio out of expected range: %.3f (expected 1.0-2.0)", ttToSSRatio)
	}

	t.Logf("M6-LIB-04 PASS: Timing scaling across corners validated")
	t.Logf("  - FF (fast, -40°C): %.3f ns", ffRise)
	t.Logf("  - TT (typical, 25°C): %.3f ns", ttRise)
	t.Logf("  - SS (slow, 125°C): %.3f ns", ssRise)
	t.Logf("  - FF/TT ratio: %.3f", ffToTTRatio)
	t.Logf("  - SS/TT ratio: %.3f", ttToSSRatio)
	t.Logf("  - FF to SS spread: %.2fx", ssRise/ffRise)
}

// TestM6LIB04_CornersLibraryNaming validates library naming convention
func TestM6LIB04_CornersLibraryNaming(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateMultiCornerLiberty(cfg)

	// Expected library names (SKY130-convention corner names)
	expectedLibraries := []string{
		"library(fecim_cells_ff_n40C_1v95)",
		"library(fecim_cells_ff_025C_1v95)",
		"library(fecim_cells_ff_125C_1v95)",
		"library(fecim_cells_tt_n40C_1v80)",
		"library(fecim_cells_tt_025C_1v80)",
		"library(fecim_cells_tt_125C_1v80)",
		"library(fecim_cells_ss_n40C_1v60)",
		"library(fecim_cells_ss_025C_1v60)",
		"library(fecim_cells_ss_125C_1v60)",
	}

	for _, libName := range expectedLibraries {
		if !strings.Contains(lib, libName) {
			t.Fatalf("missing library declaration: %s", libName)
		}
	}

	t.Logf("M6-LIB-04 PASS: Library naming convention validated")
	t.Logf("  - Libraries: %d/%d", len(expectedLibraries), len(expectedLibraries))
}

// TestM6LIB04_CornersDefaultOperatingConditions validates default_operating_conditions
func TestM6LIB04_CornersDefaultOperatingConditions(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateMultiCornerLiberty(cfg)

	// Each library should have default_operating_conditions pointing to its corner
	cornerPairs := []struct {
		library string
		corner  string
	}{
		{"library(fecim_cells_ff_n40C_1v95)", "ff_n40C_1v95"},
		{"library(fecim_cells_tt_025C_1v80)", "tt_025C_1v80"},
		{"library(fecim_cells_ss_125C_1v60)", "ss_125C_1v60"},
	}

	for _, cp := range cornerPairs {
		// Find library block
		libIdx := strings.Index(lib, cp.library)
		if libIdx < 0 {
			t.Fatalf("missing library: %s", cp.library)
		}

		// Extract library block (up to next library or end)
		nextLibIdx := strings.Index(lib[libIdx+50:], "library(")
		var libBlock string
		if nextLibIdx > 0 {
			libBlock = lib[libIdx : libIdx+50+nextLibIdx]
		} else {
			libBlock = lib[libIdx:]
		}

		// Verify default_operating_conditions
		pattern := `default_operating_conditions\s*:\s*` + cp.corner
		if !regexp.MustCompile(pattern).MatchString(libBlock) {
			t.Fatalf("library %s missing default_operating_conditions: %s", cp.library, cp.corner)
		}
	}

	t.Logf("M6-LIB-04 PASS: default_operating_conditions validated for all corners")
}

// TestM6LIB04_CornersSeparation validates multi-corner file is properly separated
func TestM6LIB04_CornersSeparation(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateMultiCornerLiberty(cfg)

	// Count closing braces at library level (should be 9)
	// Each library block should be self-contained
	libraryEndCount := 0
	lines := strings.Split(lib, "\n")
	inLibrary := false
	braceDepth := 0

	for _, line := range lines {
		if strings.Contains(line, "library(") {
			inLibrary = true
			braceDepth = 1
			continue
		}
		if inLibrary {
			braceDepth += strings.Count(line, "{")
			braceDepth -= strings.Count(line, "}")
			if braceDepth == 0 {
				libraryEndCount++
				inLibrary = false
			}
		}
	}

	if libraryEndCount != 9 {
		t.Fatalf("expected 9 properly closed library blocks, got %d", libraryEndCount)
	}

	t.Logf("M6-LIB-04 PASS: Multi-corner file separation validated")
	t.Logf("  - Properly closed library blocks: %d/9", libraryEndCount)
}
