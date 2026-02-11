package validation

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateDEF_MissingFile verifies error handling for missing files
func TestValidateDEF_MissingFile(t *testing.T) {
	err := ValidateDEF("/nonexistent/path/file.def")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

// TestValidateDEF_MissingKeywords verifies keyword validation
func TestValidateDEF_MissingKeywords(t *testing.T) {
	// Create temp file with incomplete DEF
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "incomplete.def")

	content := `VERSION 5.8 ;
DESIGN test_design ;
`
	err := os.WriteFile(defPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = ValidateDEF(defPath)
	if err == nil {
		t.Error("Expected error for missing required keywords")
	}
}

// TestValidateDEF_ComponentCountMismatch verifies component count validation
func TestValidateDEF_ComponentCountMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "mismatch.def")

	// Declare 2 components but provide only 1
	content := `VERSION 5.8 ;
DESIGN test ;
UNITS DISTANCE MICRONS 1000 ;
DIEAREA ( 0 0 ) ( 1000 1000 ) ;

COMPONENTS 2 ;
    - cell_0 fecim_bitcell + FIXED ( 0 0 ) N ;
END COMPONENTS

END DESIGN
`
	err := os.WriteFile(defPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = ValidateDEF(defPath)
	if err == nil {
		t.Error("Expected error for component count mismatch")
	}
}

// TestValidateDEF_ValidFile verifies successful validation
func TestValidateDEF_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "valid.def")

	content := `VERSION 5.8 ;
DESIGN test ;
UNITS DISTANCE MICRONS 1000 ;
DIEAREA ( 0 0 ) ( 1000 1000 ) ;

COMPONENTS 2 ;
    - cell_0 fecim_bitcell + FIXED ( 0 0 ) N ;
    - cell_1 fecim_bitcell + FIXED ( 500 0 ) N ;
END COMPONENTS

END DESIGN
`
	err := os.WriteFile(defPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = ValidateDEF(defPath)
	if err != nil {
		t.Errorf("Expected valid DEF to pass, got error: %v", err)
	}
}

// TestGetDEFStats_ExtractsData verifies stats extraction
func TestGetDEFStats_ExtractsData(t *testing.T) {
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "stats.def")

	content := `VERSION 5.8 ;
DESIGN my_design ;
UNITS DISTANCE MICRONS 1000 ;
DIEAREA ( 0 0 ) ( 2000 3000 ) ;

COMPONENTS 3 ;
    - cell_0 fecim_bitcell + FIXED ( 0 0 ) N ;
    - cell_1 fecim_bitcell + FIXED ( 500 0 ) N ;
    - cell_2 fecim_bitcell + FIXED ( 1000 0 ) N ;
END COMPONENTS

END DESIGN
`
	err := os.WriteFile(defPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	stats, err := GetDEFStats(defPath)
	if err != nil {
		t.Fatalf("GetDEFStats failed: %v", err)
	}

	if stats["design_name"] != "my_design" {
		t.Errorf("design_name: got %v, want 'my_design'", stats["design_name"])
	}

	if stats["component_count"] != 3 {
		t.Errorf("component_count: got %v, want 3", stats["component_count"])
	}

	if stats["die_area"] == nil {
		t.Error("die_area should be extracted")
	}
}

// TestCrossCheckFiles_MissingFiles verifies error handling
func TestCrossCheckFiles_MissingFiles(t *testing.T) {
	err := CrossCheckFiles("/nonexistent.lef", "/nonexistent.lib", "/nonexistent.v")
	if err == nil {
		t.Error("Expected error for missing files")
	}
}

// TestCrossCheckFiles_CellNameMismatch verifies cell name validation
func TestCrossCheckFiles_CellNameMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create LEF with cell name "cell_a"
	lefPath := filepath.Join(tmpDir, "test.lef")
	lefContent := `MACRO cell_a
  PIN A
  END A
END cell_a
`
	err := os.WriteFile(lefPath, []byte(lefContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write LEF: %v", err)
	}

	// Create LIB with cell name "cell_b"
	libPath := filepath.Join(tmpDir, "test.lib")
	libContent := `cell(cell_b) {
  pin(A) {
  }
}
`
	err = os.WriteFile(libPath, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write LIB: %v", err)
	}

	// Create Verilog with module name "cell_a"
	vPath := filepath.Join(tmpDir, "test.v")
	vContent := `module cell_a (
  input wire A
);
endmodule
`
	err = os.WriteFile(vPath, []byte(vContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write Verilog: %v", err)
	}

	err = CrossCheckFiles(lefPath, libPath, vPath)
	if err == nil {
		t.Error("Expected error for cell name mismatch")
	}
}

// TestCrossCheckFiles_PinMismatch verifies pin validation
func TestCrossCheckFiles_PinMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create LEF with pins A, B
	lefPath := filepath.Join(tmpDir, "test.lef")
	lefContent := `MACRO mycell
  PIN A
  END A
  PIN B
  END B
END mycell
`
	err := os.WriteFile(lefPath, []byte(lefContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write LEF: %v", err)
	}

	// Create LIB with pins A, C (mismatch)
	libPath := filepath.Join(tmpDir, "test.lib")
	libContent := `cell(mycell) {
  pin(A) {
  }
  pin(C) {
  }
}
`
	err = os.WriteFile(libPath, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write LIB: %v", err)
	}

	// Create Verilog with pins A, B
	vPath := filepath.Join(tmpDir, "test.v")
	vContent := `module mycell (
  input wire A,
  input wire B
);
endmodule
`
	err = os.WriteFile(vPath, []byte(vContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write Verilog: %v", err)
	}

	err = CrossCheckFiles(lefPath, libPath, vPath)
	if err == nil {
		t.Error("Expected error for pin mismatch")
	}
}

// TestCrossCheckFiles_Valid verifies successful cross-check
func TestCrossCheckFiles_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create matching LEF
	lefPath := filepath.Join(tmpDir, "test.lef")
	lefContent := `MACRO mycell
  PIN WL
  END WL
  PIN BL
  END BL
  PIN VPWR
  END VPWR
  PIN VGND
  END VGND
END mycell
`
	err := os.WriteFile(lefPath, []byte(lefContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write LEF: %v", err)
	}

	// Create matching LIB
	libPath := filepath.Join(tmpDir, "test.lib")
	libContent := `cell(mycell) {
  pin(WL) {
  }
  pin(BL) {
  }
  pin(VPWR) {
  }
  pin(VGND) {
  }
}
`
	err = os.WriteFile(libPath, []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write LIB: %v", err)
	}

	// Create matching Verilog
	vPath := filepath.Join(tmpDir, "test.v")
	vContent := `module mycell (
  input wire WL,
  output wire BL,
  input wire VPWR,
  input wire VGND
);
endmodule
`
	err = os.WriteFile(vPath, []byte(vContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write Verilog: %v", err)
	}

	err = CrossCheckFiles(lefPath, libPath, vPath)
	if err != nil {
		t.Errorf("Expected valid cross-check to pass, got error: %v", err)
	}
}

// TestSlicesEqual verifies the helper function
func TestSlicesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{"empty slices", []string{}, []string{}, true},
		{"identical", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different order", []string{"a", "b"}, []string{"b", "a"}, true},
		{"different length", []string{"a"}, []string{"a", "b"}, false},
		{"different elements", []string{"a", "b"}, []string{"a", "c"}, false},
		{"nil vs empty", nil, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slicesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("slicesEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestExtractLEFData verifies LEF parsing
func TestExtractLEFData(t *testing.T) {
	tmpDir := t.TempDir()
	lefPath := filepath.Join(tmpDir, "test.lef")

	content := `VERSION 5.8 ;
MACRO test_cell
  CLASS CORE ;
  ORIGIN 0.0 0.0 ;
  SIZE 1.0 BY 2.0 ;
  PIN A
    DIRECTION INPUT ;
  END A
  PIN B
    DIRECTION OUTPUT ;
  END B
END test_cell
`
	err := os.WriteFile(lefPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write LEF: %v", err)
	}

	pins, cellName, err := extractLEFData(lefPath)
	if err != nil {
		t.Fatalf("extractLEFData failed: %v", err)
	}

	if cellName != "test_cell" {
		t.Errorf("cellName: got %s, want test_cell", cellName)
	}

	expectedPins := []string{"A", "B"}
	if !slicesEqual(pins, expectedPins) {
		t.Errorf("pins: got %v, want %v", pins, expectedPins)
	}
}

// TestExtractLibData verifies Liberty parsing
func TestExtractLibData(t *testing.T) {
	tmpDir := t.TempDir()
	libPath := filepath.Join(tmpDir, "test.lib")

	content := `library(test_lib) {
  cell(my_cell) {
    pin(IN1) {
      direction : input;
    }
    pin(OUT1) {
      direction : output;
    }
  }
}
`
	err := os.WriteFile(libPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write LIB: %v", err)
	}

	pins, cellName, err := extractLibData(libPath)
	if err != nil {
		t.Fatalf("extractLibData failed: %v", err)
	}

	if cellName != "my_cell" {
		t.Errorf("cellName: got %s, want my_cell", cellName)
	}

	expectedPins := []string{"IN1", "OUT1"}
	if !slicesEqual(pins, expectedPins) {
		t.Errorf("pins: got %v, want %v", pins, expectedPins)
	}
}

// TestExtractVerilogData verifies Verilog parsing
func TestExtractVerilogData(t *testing.T) {
	tmpDir := t.TempDir()
	vPath := filepath.Join(tmpDir, "test.v")

	content := `module top_module (
  input wire clk,
  input wire rst,
  output wire out
);
endmodule
`
	err := os.WriteFile(vPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write Verilog: %v", err)
	}

	pins, moduleName, err := extractVerilogData(vPath)
	if err != nil {
		t.Fatalf("extractVerilogData failed: %v", err)
	}

	if moduleName != "top_module" {
		t.Errorf("moduleName: got %s, want top_module", moduleName)
	}

	expectedPins := []string{"clk", "rst", "out"}
	if !slicesEqual(pins, expectedPins) {
		t.Errorf("pins: got %v, want %v", pins, expectedPins)
	}
}
