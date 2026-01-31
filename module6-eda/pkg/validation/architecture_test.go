package validation

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

func TestArchitecture8x8_Passive(t *testing.T) {
	// Create 8x8 passive array
	cfg := compiler.NewComputeConfig(8, 8)
	cfg.Architecture = "passive"

	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Generate and validate DEF
	def := export.GenerateDEFWithDefaults(design)
	if !strings.Contains(def, "Architecture: passive") {
		t.Error("Should indicate passive architecture")
	}
	if !strings.Contains(def, "COMPONENTS 64") {
		t.Error("Should have 64 cells for 8x8 array")
	}
	// Passive should NOT have SL pins
	if strings.Contains(def, "- SL[") {
		t.Error("Passive should not have SL pins")
	}
	// Should use VPWR/VGND
	if !strings.Contains(def, "- VPWR") {
		t.Error("Should use VPWR power pin")
	}
	if !strings.Contains(def, "- VGND") {
		t.Error("Should use VGND ground pin")
	}

	t.Log("8x8 Passive DEF validated successfully")
}

func TestArchitecture8x8_1T1R(t *testing.T) {
	// Create 8x8 1T1R array
	cfg := compiler.NewComputeConfig(8, 8)
	cfg.Architecture = "1t1r"

	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Generate and validate DEF
	def := export.GenerateDEFWithDefaults(design)
	if !strings.Contains(def, "Architecture: 1t1r") {
		t.Error("Should indicate 1t1r architecture")
	}
	if !strings.Contains(def, "COMPONENTS 64") {
		t.Error("Should have 64 cells for 8x8 array")
	}
	// 1T1R should have SL pins but NOT CSL
	if !strings.Contains(def, "- SL[") {
		t.Error("1T1R should have SL pins")
	}
	if strings.Contains(def, "- CSL[") {
		t.Error("1T1R should not have CSL pins")
	}
	// Should use fecim_1t1r cell
	if !strings.Contains(def, "fecim_1t1r") {
		t.Error("1T1R should use fecim_1t1r cell")
	}

	t.Log("8x8 1T1R DEF validated successfully")
}

func TestArchitecture8x8_2T1R(t *testing.T) {
	// Create 8x8 2T1R array
	cfg := compiler.NewComputeConfig(8, 8)
	cfg.Architecture = "2t1r"

	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Generate and validate DEF
	def := export.GenerateDEFWithDefaults(design)
	if !strings.Contains(def, "Architecture: 2t1r") {
		t.Error("Should indicate 2t1r architecture")
	}
	if !strings.Contains(def, "COMPONENTS 64") {
		t.Error("Should have 64 cells for 8x8 array")
	}
	// 2T1R should have BOTH SL and CSL pins
	if !strings.Contains(def, "- SL[") {
		t.Error("2T1R should have SL pins")
	}
	if !strings.Contains(def, "- CSL[") {
		t.Error("2T1R should have CSL pins")
	}
	// Should use fecim_2t1r cell
	if !strings.Contains(def, "fecim_2t1r") {
		t.Error("2T1R should use fecim_2t1r cell")
	}

	t.Log("8x8 2T1R DEF validated successfully")
}

func TestArchitectureDetection(t *testing.T) {
	testCases := []struct {
		name string
		arch string
	}{
		{"passive", "passive"},
		{"1t1r lowercase", "1t1r"},
		{"1T1R uppercase", "1T1R"}, // Should normalize
		{"2t1r lowercase", "2t1r"},
		{"2T1R uppercase", "2T1R"}, // Should normalize
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := compiler.NewComputeConfig(4, 4)
			cfg.Architecture = tc.arch

			design, err := compiler.GenerateDesign(cfg)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			def := export.GenerateDEFWithDefaults(design)

			// Should contain valid DEF header
			if !strings.Contains(def, "VERSION") {
				t.Error("DEF should contain VERSION header")
			}
			if !strings.Contains(def, "COMPONENTS") {
				t.Error("DEF should contain COMPONENTS section")
			}
		})
	}
}

func TestLibertyGeneration(t *testing.T) {
	testCases := []struct {
		cellType  string
		expectLib string
		expectCSL bool
		expectSL  bool
	}{
		{"passive", "fecim_cells", false, false},
		{"1t1r", "fecim_1t1r_cells", false, true},
		{"2t1r", "fecim_2t1r_cells", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.cellType, func(t *testing.T) {
			cfg := config.DefaultCellConfig()
			cfg.CellType = tc.cellType

			lib := export.GenerateLiberty(cfg)

			if !strings.Contains(lib, tc.expectLib) {
				t.Errorf("Expected library %q, got: %s...", tc.expectLib, lib[:100])
			}

			hasCSL := strings.Contains(lib, "pin(CSL)")
			if hasCSL != tc.expectCSL {
				t.Errorf("CSL pin: got %v, want %v", hasCSL, tc.expectCSL)
			}

			hasSL := strings.Contains(lib, "pin(SL)")
			if hasSL != tc.expectSL {
				t.Errorf("SL pin: got %v, want %v", hasSL, tc.expectSL)
			}
		})
	}
}
