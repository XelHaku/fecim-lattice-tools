package export

import "testing"

func TestValidateLatticeDimensions(t *testing.T) {
	if err := ValidateLatticeDimensions(64, 64); err != nil {
		t.Fatalf("expected valid dimensions, got %v", err)
	}
	if err := ValidateLatticeDimensions(0, 10); err == nil {
		t.Fatal("expected error for invalid dimensions")
	}
	if err := ValidateLatticeDimensions(maxLatticeDim+1, 1); err == nil {
		t.Fatal("expected error for oversize dimensions")
	}
}

func TestGenerateLatticeVerilogRejectsInvalidDimensions(t *testing.T) {
	out := GenerateLatticeVerilog(0, 4)
	if out == "" || out[:2] != "//" {
		t.Fatalf("expected error comment output for invalid dimensions, got %q", out)
	}
}
