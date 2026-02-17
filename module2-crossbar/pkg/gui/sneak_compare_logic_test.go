package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func TestSneakCompareWidget_HelperLogic(t *testing.T) {
	w := NewSneakCompareWidget(2, 2)
	w.statsLabel = widget.NewLabel("")
	w.comparisonLabel = widget.NewLabel("")

	empty := w.analysisToHeatmapData(nil)
	if len(empty) != 2 || len(empty[0]) != 2 {
		t.Fatalf("expected empty heatmap shape 2x2, got %dx%d", len(empty), len(empty[0]))
	}

	analysis := &crossbar.SneakPathAnalysis{SneakCurrents: [][]float64{{0.1, 0.2}, {0.3, 0.4}}, TotalSneak: 0.4}
	if got := w.analysisToHeatmapData(analysis); got[1][1] != 0.4 {
		t.Fatalf("analysis mapping failed: %#v", got)
	}

	max := w.findMaxSneak([][]float64{{0, 2}}, [][]float64{{1}})
	if max != 2 {
		t.Fatalf("expected max=2, got %v", max)
	}
	if got := w.findMaxSneak([][]float64{{0}}, [][]float64{{0}}); got <= 0 {
		t.Fatalf("expected non-zero fallback max, got %v", got)
	}

	w.passiveAnalysis = &crossbar.SneakPathAnalysis{TotalSneak: 1e-3}
	w.activeAnalysis = &crossbar.SneakPathAnalysis{TotalSneak: 1e-6}
	w.updateStats()
	if !strings.Contains(w.comparisonLabel.Text, "virtually all") {
		t.Fatalf("expected >99%% insight, got %q", w.comparisonLabel.Text)
	}

	w.activeAnalysis = &crossbar.SneakPathAnalysis{TotalSneak: 5e-5}
	w.updateStats()
	if !strings.Contains(w.comparisonLabel.Text, "excellent") {
		t.Fatalf("expected >90%% insight, got %q", w.comparisonLabel.Text)
	}

	w.activeAnalysis = &crossbar.SneakPathAnalysis{TotalSneak: 4e-4}
	w.updateStats()
	if !strings.Contains(w.comparisonLabel.Text, "moderate") {
		t.Fatalf("expected >50%% insight, got %q", w.comparisonLabel.Text)
	}

	w.activeAnalysis = &crossbar.SneakPathAnalysis{TotalSneak: 9e-4}
	w.updateStats()
	if !strings.Contains(w.comparisonLabel.Text, "Compare") {
		t.Fatalf("expected low-reduction fallback, got %q", w.comparisonLabel.Text)
	}
}

func TestConductanceTooltipAndWrapperHelpers(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{Rows: 1, Cols: 1, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	text := ConductanceTooltip(0, 0, 0.5, arr)
	if !strings.Contains(text, "CONDUCTANCE") || !strings.Contains(text, "Bits/cell") {
		t.Fatalf("unexpected conductance tooltip: %q", text)
	}

	ir := &crossbar.IRDropAnalysis{EffectiveVoltage: [][]float64{{0.9}}, WorstCaseCell: [2]int{0, 0}}
	if !strings.Contains(IRDropTooltip(0, 0, ir, arr), "IR DROP") {
		t.Fatalf("wrapper IRDropTooltip should include IR DROP")
	}
	sn := &crossbar.SneakPathAnalysis{SneakCurrents: [][]float64{{0.01}}, TotalSignal: 1, MaxSneakRatio: 0.01}
	if !strings.Contains(SneakPathTooltip(0, 0, sn, 0, 0, arr), "SNEAK PATH") {
		t.Fatalf("wrapper SneakPathTooltip should include SNEAK PATH")
	}
}
