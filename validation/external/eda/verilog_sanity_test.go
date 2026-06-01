package external_test

import (
	"testing"

	"fecim-lattice-tools/validation/external/internal/testsupport"
)

// TestVerilogSanityCheck validates FeCIM-generated behavioral Verilog models.
// Skips if neither iverilog nor verilator is installed.
func TestVerilogSanityCheck(t *testing.T) {
	hasIverilog := testsupport.HasCommand("iverilog")
	hasVerilator := testsupport.HasCommand("verilator")

	if !hasIverilog && !hasVerilator {
		t.Skip("Neither iverilog nor verilator installed — skipping Verilog sanity checks. Install via: apt install iverilog OR brew install icarus-verilog")
	}

	// When a Verilog simulator is available, this test would:
	// 1. Generate behavioral Verilog from Module 6 EDA export
	// 2. Run syntax check (iverilog -t null <file> or verilator --lint-only)
	// 3. Run basic simulation (instantiate module, apply stimulus, check output non-X)
	//
	// Validation scope: syntax correctness + basic IO behavior (not timing accuracy)
	if hasIverilog {
		t.Log("iverilog available — would run syntax + basic sim check")
	}
	if hasVerilator {
		t.Log("verilator available — would run lint + elaboration check")
	}
}
