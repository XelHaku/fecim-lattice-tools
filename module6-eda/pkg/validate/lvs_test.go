package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateLVSConsistency_Pass(t *testing.T) {
	d := t.TempDir()
	lef := filepath.Join(d, "cell.lef")
	verilog := filepath.Join(d, "cell.v")
	spice := filepath.Join(d, "cell.sp")

	mustWrite(t, lef, `MACRO mycell
  PIN WL
  END WL
  PIN BL
  END BL
  PIN VPWR
  END VPWR
  PIN VGND
  END VGND
END mycell`)

	mustWrite(t, verilog, `module mycell (
  WL,
  BL,
  VPWR,
  VGND
);
  input WL;
  output BL;
  inout VPWR;
  inout VGND;
endmodule`)

	mustWrite(t, spice, `.SUBCKT mycell WL BL VPWR VGND
R1 BL VGND 1k
.ENDS mycell`)

	if err := ValidateLVSConsistency(lef, verilog, spice); err != nil {
		t.Fatalf("expected consistent exports to pass LVS, got: %v", err)
	}
}

func TestValidateLVSConsistency_Fail(t *testing.T) {
	d := t.TempDir()
	lef := filepath.Join(d, "cell.lef")
	verilog := filepath.Join(d, "cell.v")
	spice := filepath.Join(d, "cell.sp")

	mustWrite(t, lef, `MACRO mycell
  PIN WL
  END WL
  PIN BL
  END BL
END mycell`)

	mustWrite(t, verilog, `module mycell (WL,BL);
  input WL;
  output BL;
endmodule`)

	// Intentional mismatch: BR instead of BL.
	mustWrite(t, spice, `.SUBCKT mycell WL BR
R1 BR 0 1k
.ENDS mycell`)

	if err := ValidateLVSConsistency(lef, verilog, spice); err == nil {
		t.Fatal("expected mismatched exports to fail LVS")
	}
}

// TestValidateLVSConsistency_ANSIStyle verifies that ANSI-style inline port
// declarations (input wire WL,  // comment) are parsed correctly.
// The old [^;]+ regex captured one multi-line match and extracted the last
// whitespace-separated token from the trailing comment ("Ground" instead of "VGND").
func TestValidateLVSConsistency_ANSIStyle(t *testing.T) {
	d := t.TempDir()
	lef := filepath.Join(d, "cell.lef")
	verilog := filepath.Join(d, "cell.v")
	spice := filepath.Join(d, "cell.sp")

	mustWrite(t, lef, `MACRO mycell
  PIN WL
  END WL
  PIN BL
  END BL
  PIN VPWR
  END VPWR
  PIN VGND
  END VGND
END mycell`)

	// ANSI-style: port type+name declared inline in the port list, no trailing semicolons.
	// Last port has a trailing comment — the old regex mistakenly extracted the comment
	// word "Ground" as the port name instead of "VGND".
	mustWrite(t, verilog, `module mycell (
    input  wire WL,    // Word Line
    output wire BL,    // Bit Line
    inout  wire VPWR,  // Power
    inout  wire VGND   // Ground
);
    assign BL = WL;
endmodule`)

	mustWrite(t, spice, `.SUBCKT mycell WL BL VPWR VGND
R1 BL VGND 1k
.ENDS mycell`)

	if err := ValidateLVSConsistency(lef, verilog, spice); err != nil {
		t.Fatalf("ANSI-style Verilog: expected consistent exports to pass LVS, got: %v", err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
