package crossbar

import (
	"strings"
	"testing"
)

func TestGenerateMVMSneakTrace(t *testing.T) {
	arr, _ := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8})
	_ = arr.ProgramWeightMatrix([][]float64{{0.9, 0.8, 0.7, 0.6}, {0.5, 0.4, 0.3, 0.2}, {0.6, 0.7, 0.8, 0.9}, {0.3, 0.5, 0.7, 0.9}})
	trace := arr.GenerateMVMSneakTrace([]float64{1, 0.8, 0.6, 0.4}, &MVMOptions{Architecture: "0T1R"}, 2)
	if trace == nil || len(trace.Rows) != 4 || trace.TotalSneak <= 0 {
		t.Fatalf("invalid trace: %+v", trace)
	}
	if !strings.Contains(trace.FormatText(2, 1), "MVM Sneak Path Trace") {
		t.Fatal("missing trace header")
	}
}

func TestMVMWithNonIdealities_PopulatesSneakTrace(t *testing.T) {
	arr, _ := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8})
	_ = arr.ProgramWeightMatrix([][]float64{{0.9, 0.9, 0.9, 0.9}, {0.8, 0.8, 0.8, 0.8}, {0.7, 0.7, 0.7, 0.7}, {0.6, 0.6, 0.6, 0.6}})
	opts := DefaultMVMOptions()
	opts.EnableIRDrop = false
	opts.EnableVariation = false
	opts.EnableSneakPaths = true
	opts.Architecture = "0T1R"
	res, err := arr.MVMWithNonIdealities([]float64{1, 1, 1, 1}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if res.SneakTrace == nil {
		t.Fatal("expected SneakTrace")
	}
}
