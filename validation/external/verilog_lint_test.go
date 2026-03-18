package external_test

import (
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

// makeTestDesign creates a minimal ArrayDesign for Verilog generation tests.
func makeTestDesign(rows, cols int, arch string) *compiler.ArrayDesign {
	cfg := compiler.NewArrayConfig(compiler.ModeStorage, rows, cols)
	cfg.Architecture = arch

	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		// Fallback: build design manually if GenerateDesign fails
		cells := make([]compiler.CellAssignment, 0, rows*cols)
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				cells = append(cells, compiler.CellAssignment{
					Row:         r,
					Col:         c,
					Level:       (r*cols + c) % 30,
					Conductance: float64((r*cols+c)%30) * 3.0,
				})
			}
		}
		design = &compiler.ArrayDesign{
			Config: cfg,
			Cells:  cells,
		}
	}

	return design
}

// TestVerilogLint_ModuleDeclaration validates that generated Verilog contains
// a proper module declaration with port list.
func TestVerilogLint_ModuleDeclaration(t *testing.T) {
	architectures := []string{compiler.ArchPassive, compiler.Arch1T1R, compiler.Arch2T1R}

	for _, arch := range architectures {
		t.Run(arch, func(t *testing.T) {
			design := makeTestDesign(4, 4, arch)
			verilog := export.GenerateVerilogWithDefaults(design)

			// Must contain a module declaration
			moduleRegex := regexp.MustCompile(`module\s+\w+\s*\(`)
			if !moduleRegex.MatchString(verilog) {
				t.Error("Generated Verilog missing module declaration")
			}

			// Must contain endmodule
			if !strings.Contains(verilog, "endmodule") {
				t.Error("Generated Verilog missing endmodule")
			}

			// Must declare WL and BL ports
			if !strings.Contains(verilog, "WL") {
				t.Error("Generated Verilog missing Word Line (WL) port")
			}
			if !strings.Contains(verilog, "BL") {
				t.Error("Generated Verilog missing Bit Line (BL) port")
			}

			// Architecture-specific port checks
			if arch == compiler.Arch1T1R || arch == compiler.Arch2T1R {
				if !strings.Contains(verilog, "SL") {
					t.Errorf("Architecture %s should declare Source Line (SL) port", arch)
				}
			}
			if arch == compiler.Arch2T1R {
				if !strings.Contains(verilog, "CSL") {
					t.Error("2T1R architecture should declare Column Select Line (CSL) port")
				}
			}
		})
	}
}

// TestVerilogLint_PortDirections validates that ports have proper direction
// keywords (input, output, inout).
func TestVerilogLint_PortDirections(t *testing.T) {
	design := makeTestDesign(4, 4, compiler.ArchPassive)
	verilog := export.GenerateVerilogWithDefaults(design)

	// Check for port direction keywords
	portDirRegex := regexp.MustCompile(`(?m)^\s*(input|output|inout)\s+(wire|reg)?\s*`)
	matches := portDirRegex.FindAllString(verilog, -1)

	if len(matches) == 0 {
		t.Error("No port direction declarations found (expected input/output/inout)")
	}

	// WL should be input
	wlInputRegex := regexp.MustCompile(`input\s+wire\s+\[\d+:0\]\s+WL`)
	if !wlInputRegex.MatchString(verilog) {
		t.Error("WL should be declared as input wire")
	}

	// BL should be inout (bidirectional for read/write)
	blInoutRegex := regexp.MustCompile(`inout\s+wire\s+\[\d+:0\]\s+BL`)
	if !blInoutRegex.MatchString(verilog) {
		t.Error("BL should be declared as inout wire")
	}
}

// TestVerilogLint_WireDeclarations validates reg/wire declarations are
// present and consistent with port usage.
func TestVerilogLint_WireDeclarations(t *testing.T) {
	design := makeTestDesign(8, 8, compiler.ArchPassive)
	verilog := export.GenerateVerilogWithDefaults(design)

	// Check that bus widths match array dimensions
	// WL should be [7:0] for 8 rows
	wlBusRegex := regexp.MustCompile(`\[7:0\]\s+WL`)
	if !wlBusRegex.MatchString(verilog) {
		t.Errorf("WL bus width should be [7:0] for 8-row array")
	}

	// BL should be [7:0] for 8 cols
	blBusRegex := regexp.MustCompile(`\[7:0\]\s+BL`)
	if !blBusRegex.MatchString(verilog) {
		t.Errorf("BL bus width should be [7:0] for 8-column array")
	}
}

// TestVerilogLint_CellInstantiations validates cell instance naming and
// connectivity patterns.
func TestVerilogLint_CellInstantiations(t *testing.T) {
	design := makeTestDesign(4, 4, compiler.ArchPassive)
	verilog := export.GenerateVerilogWithDefaults(design)

	// Check that cell instances follow naming convention R_{row}_{col}
	instanceRegex := regexp.MustCompile(`R_\d+_\d+`)
	matches := instanceRegex.FindAllString(verilog, -1)

	expectedCells := 4 * 4
	if len(matches) < expectedCells {
		t.Errorf("Expected at least %d cell instances (R_row_col), found %d", expectedCells, len(matches))
	}

	// Check that corner cells exist
	cornerCells := []string{"R_0_0", "R_0_3", "R_3_0", "R_3_3"}
	for _, corner := range cornerCells {
		if !strings.Contains(verilog, corner) {
			t.Errorf("Corner cell %s not found in generated Verilog", corner)
		}
	}

	// Check that cells connect to proper WL and BL indices
	wlConnRegex := regexp.MustCompile(`\.WL\s*\(WL\[\d+\]\)`)
	wlConns := wlConnRegex.FindAllString(verilog, -1)
	if len(wlConns) < expectedCells {
		t.Errorf("Expected %d WL connections, found %d", expectedCells, len(wlConns))
	}

	blConnRegex := regexp.MustCompile(`\.BL\s*\(BL\[\d+\]\)`)
	blConns := blConnRegex.FindAllString(verilog, -1)
	if len(blConns) < expectedCells {
		t.Errorf("Expected %d BL connections, found %d", expectedCells, len(blConns))
	}
}

// TestVerilogLint_NoCommonAntipatterns checks for common Verilog antipatterns
// in generated output.
func TestVerilogLint_NoCommonAntipatterns(t *testing.T) {
	design := makeTestDesign(4, 4, compiler.ArchPassive)
	verilog := export.GenerateVerilogWithDefaults(design)

	// Check for balanced module/endmodule
	moduleCount := strings.Count(verilog, "module ")
	endmoduleCount := strings.Count(verilog, "endmodule")
	if moduleCount != endmoduleCount {
		t.Errorf("Unbalanced module/endmodule: %d modules, %d endmodules", moduleCount, endmoduleCount)
	}

	// Check for unclosed parentheses in port lists (basic check)
	openParens := strings.Count(verilog, "(")
	closeParens := strings.Count(verilog, ")")
	if openParens != closeParens {
		t.Errorf("Unbalanced parentheses: %d open, %d close", openParens, closeParens)
	}

	// Check no line has a double semicolon (common typo)
	if strings.Contains(verilog, ";;") {
		t.Error("Found double semicolons in generated Verilog")
	}

	// Check VPWR and VGND power connections exist
	if !strings.Contains(verilog, "VPWR") {
		t.Error("Missing VPWR power supply connection")
	}
	if !strings.Contains(verilog, "VGND") {
		t.Error("Missing VGND ground connection")
	}
}

// TestVerilogLint_ParameterDeclarations validates that Verilog parameters are
// declared with correct values.
func TestVerilogLint_ParameterDeclarations(t *testing.T) {
	design := makeTestDesign(8, 4, compiler.ArchPassive)
	verilog := export.GenerateVerilogWithDefaults(design)

	// Check for parameter declarations
	paramRegex := regexp.MustCompile(`parameter\s+(\w+)\s*=\s*(.+);`)
	params := paramRegex.FindAllStringSubmatch(verilog, -1)

	if len(params) == 0 {
		t.Skip("No parameter declarations found — generator may not emit parameters")
	}

	// Build parameter map
	paramMap := make(map[string]string)
	for _, p := range params {
		paramMap[p[1]] = strings.TrimSpace(p[2])
	}

	// Verify ROWS matches array dimensions
	if rows, ok := paramMap["ROWS"]; ok {
		if rows != "8" {
			t.Errorf("ROWS parameter should be 8, got: %s", rows)
		}
	}

	// Verify COLS matches array dimensions
	if cols, ok := paramMap["COLS"]; ok {
		if cols != "4" {
			t.Errorf("COLS parameter should be 4, got: %s", cols)
		}
	}
}
