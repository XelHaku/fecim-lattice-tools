package crossbar

import "testing"

func TestArray_RejectsInvalidDimensions(t *testing.T) {
	cases := []Config{
		{Rows: 0, Cols: 4, ADCBits: 8, DACBits: 8},
		{Rows: 4, Cols: 0, ADCBits: 8, DACBits: 8},
		{Rows: -1, Cols: 4, ADCBits: 8, DACBits: 8},
		{Rows: 4, Cols: -1, ADCBits: 8, DACBits: 8},
	}

	for _, cfg := range cases {
		cfg := cfg
		t.Run("dims", func(t *testing.T) {
			arr, err := NewArray(&cfg)
			if err == nil {
				t.Fatalf("NewArray(%+v) expected error, got array=%v", cfg, arr)
			}
		})
	}
}

func TestArray_OutOfBoundsAccessReturnsErrors(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	if err := arr.ProgramWeight(-1, 0, 0.5); err == nil {
		t.Fatal("ProgramWeight with negative row should error")
	}
	if err := arr.ProgramWeight(0, 2, 0.5); err == nil {
		t.Fatal("ProgramWeight with out-of-range col should error")
	}

	if _, err := arr.GetCellStats(3, 0); err == nil {
		t.Fatal("GetCellStats out of bounds should error")
	}
	if g := arr.GetPhysicalConductanceForCell(3, 0); g != GMin {
		t.Fatalf("GetPhysicalConductanceForCell out of bounds = %e, want GMin=%e", g, GMin)
	}
}
