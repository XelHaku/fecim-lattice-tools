package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

func TestFileGeneration8x8(t *testing.T) {
	tmpDir := t.TempDir()

	architectures := []struct {
		name     string
		arch     string
		cellName string
		hasSL    bool
		hasCSL   bool
	}{
		{"passive", "passive", "fecim_bit", false, false},
		{"1T1R", "1t1r", "fecim_1t1r", true, false},
		{"2T1R", "2t1r", "fecim_2t1r", true, true},
	}

	for _, tc := range architectures {
		t.Run(tc.name, func(t *testing.T) {
			archDir := filepath.Join(tmpDir, tc.name)
			os.MkdirAll(archDir, 0755)

			// Generate design
			cfg := compiler.NewComputeConfig(8, 8)
			cfg.Architecture = tc.arch
			design, err := compiler.GenerateDesign(cfg)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Generate and write DEF
			def := export.GenerateDEFWithDefaults(design)
			defPath := filepath.Join(archDir, "array.def")
			if err := os.WriteFile(defPath, []byte(def), 0644); err != nil {
				t.Fatalf("Failed to write DEF: %v", err)
			}

			// Generate and write Verilog
			// Convert compiler.ArrayConfig to config.ArrayConfig
			var mode string
			switch design.Config.Mode {
			case compiler.ModeStorage:
				mode = "storage"
			case compiler.ModeMemory:
				mode = "memory"
			case compiler.ModeCompute:
				mode = "compute"
			}
			arrayCfg := config.ArrayConfig{
				Rows:         design.Config.ArrayRows,
				Cols:         design.Config.ArrayCols,
				Mode:         mode,
				Architecture: design.Config.Architecture,
				Technology:   design.Config.Technology,
				CellWidth:    design.Config.CellPitch,
				CellHeight:   design.Config.RowHeight,
			}
			verilog := export.GenerateArrayVerilog(arrayCfg)
			verilogPath := filepath.Join(archDir, "array.v")
			if err := os.WriteFile(verilogPath, []byte(verilog), 0644); err != nil {
				t.Fatalf("Failed to write Verilog: %v", err)
			}

			// Generate and write LEF
			cellCfg := config.DefaultCellConfig()
			cellCfg.CellType = tc.arch
			lef := export.GenerateLEF(cellCfg)
			lefPath := filepath.Join(archDir, "cell.lef")
			if err := os.WriteFile(lefPath, []byte(lef), 0644); err != nil {
				t.Fatalf("Failed to write LEF: %v", err)
			}

			// Generate and write Liberty
			lib := export.GenerateLiberty(cellCfg)
			libPath := filepath.Join(archDir, "cell.lib")
			if err := os.WriteFile(libPath, []byte(lib), 0644); err != nil {
				t.Fatalf("Failed to write Liberty: %v", err)
			}

			// Validate DEF content
			defContent, _ := os.ReadFile(defPath)
			defStr := string(defContent)

			// Check architecture in DEF
			if !strings.Contains(defStr, "COMPONENTS 64") {
				t.Error("DEF should have 64 components for 8x8")
			}

			// Check cell naming in DEF
			if !strings.Contains(defStr, tc.cellName) {
				t.Errorf("DEF should use %s cell", tc.cellName)
			}

			// Check SL pins
			hasSL := strings.Contains(defStr, "- SL[")
			if hasSL != tc.hasSL {
				t.Errorf("SL pins: got %v, want %v", hasSL, tc.hasSL)
			}

			// Check CSL pins
			hasCSL := strings.Contains(defStr, "- CSL[")
			if hasCSL != tc.hasCSL {
				t.Errorf("CSL pins: got %v, want %v", hasCSL, tc.hasCSL)
			}

			// Validate Verilog
			verilogContent, _ := os.ReadFile(verilogPath)
			verilogStr := string(verilogContent)
			expectedModule := fmt.Sprintf("module fecim_crossbar_%dx%d", 8, 8)
			if !strings.Contains(verilogStr, expectedModule) {
				t.Errorf("Verilog should contain %s module", expectedModule)
			}

			// Validate LEF has correct pins
			lefContent, _ := os.ReadFile(lefPath)
			lefStr := string(lefContent)
			if !strings.Contains(lefStr, "MACRO") {
				t.Error("LEF should contain MACRO definition")
			}
			if !strings.Contains(lefStr, "PIN WL") {
				t.Error("LEF should have WL pin")
			}
			if !strings.Contains(lefStr, "PIN BL") {
				t.Error("LEF should have BL pin")
			}

			// Validate Liberty
			libContent, _ := os.ReadFile(libPath)
			libStr := string(libContent)
			if !strings.Contains(libStr, "library(") {
				t.Error("Liberty should have library definition")
			}
			if !strings.Contains(libStr, "pin(WL)") {
				t.Error("Liberty should have WL pin")
			}

			t.Logf("%s: Generated DEF (%d bytes), Verilog (%d bytes), LEF (%d bytes), Liberty (%d bytes)",
				tc.name, len(defContent), len(verilogContent), len(lefContent), len(libContent))
		})
	}
}

func TestDetectArchitectureFromDEF(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		content  string
		wantArch string
	}{
		{
			"passive",
			"VERSION 5.8 ;\nCOMPONENTS 4 ;\n- R_0_0 fecim_bit + FIXED ( 1000 1000 ) N ;\n",
			"passive",
		},
		{
			"1t1r",
			"VERSION 5.8 ;\nCOMPONENTS 4 ;\n- R_0_0 fecim_1t1r + FIXED ( 1000 1000 ) N ;\n",
			"1t1r",
		},
		{
			"2t1r",
			"VERSION 5.8 ;\nCOMPONENTS 4 ;\n- R_0_0 fecim_2t1r + FIXED ( 1000 1000 ) N ;\n",
			"2t1r",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defPath := filepath.Join(tmpDir, tc.name+".def")
			if err := os.WriteFile(defPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write test DEF: %v", err)
			}

			// Use the existing detectArchitectureFromDEF function from openlane.go
			gotArch := detectArchitectureFromDEF(defPath)
			if gotArch != tc.wantArch {
				t.Errorf("detectArchitectureFromDEF() = %q, want %q", gotArch, tc.wantArch)
			}
		})
	}
}
