package eda

import "testing"

func TestDesignGenerationWorkflowBuildsDefaultEDAState(t *testing.T) {
	state := newDesignGenerationWorkflow(8, 8).buildState()

	if state.DesignName != "fecim_crossbar_8x8" {
		t.Fatalf("design name = %q", state.DesignName)
	}
	if state.ProcessNode != "sky130" {
		t.Fatalf("process node = %q", state.ProcessNode)
	}
	if state.ArrayRows != 8 || state.ArrayCols != 8 {
		t.Fatalf("array size = %dx%d, want 8x8", state.ArrayRows, state.ArrayCols)
	}
	if state.TotalCells <= 0 || state.AreaMM2 <= 0 || state.PowerMW <= 0 {
		t.Fatalf("missing design stats: cells=%d area=%.6f power=%.6f", state.TotalCells, state.AreaMM2, state.PowerMW)
	}
	if len(state.ExportFormats) != 5 {
		t.Fatalf("export format count = %d, want 5", len(state.ExportFormats))
	}
	for name, snippet := range map[string]string{
		"spice":   state.SPICESnippet,
		"verilog": state.VerilogSnippet,
		"def":     state.DEFSnippet,
		"lef":     state.LEFSnippet,
	} {
		if snippet == "" {
			t.Fatalf("%s snippet is empty", name)
		}
	}
}
