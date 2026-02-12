package comparison

import (
	"strings"
	"testing"
)

func TestAdditionalWorkloadAndRendererCoverage(t *testing.T) {
	r := NewRenderer()

	// Cover LLM workload constructor
	llm := LLMWorkload()
	if llm.Name != "LLM-70B" || llm.TotalOps <= 0 {
		t.Fatalf("unexpected LLM workload: %+v", llm)
	}

	// Build representative comparison data for rendering paths
	workload := ResNet50Workload()
	cmp := CompareArchitectures(workload, 4, 25000)
	adv := CalculateAdvantages(cmp)

	inferenceText := r.RenderInferenceComparison(cmp.Results, workload)
	if !strings.Contains(inferenceText, "Inference Comparison") || !strings.Contains(inferenceText, "Architecture") {
		t.Fatalf("unexpected RenderInferenceComparison output: %q", inferenceText)
	}

	dcText := r.RenderDataCenterComparison(cmp.DataCenter, 25000)
	if !strings.Contains(dcText, "Data Center Comparison") || !strings.Contains(dcText, "Target Throughput") {
		t.Fatalf("unexpected RenderDataCenterComparison output: %q", dcText)
	}

	advText := r.RenderAdvantages(adv)
	if !strings.Contains(advText, "FeCIM Advantages") || !strings.Contains(advText, "vs GPU Accelerator") {
		t.Fatalf("unexpected RenderAdvantages output: %q", advText)
	}
}

func TestAdditionalFormattingAndBranchCoverage(t *testing.T) {
	r := NewRenderer()

	// RenderBarChart branches: maxVal==0, lowerIsBetter=false (indexOfMax), value formatting tiers
	chart := r.RenderBarChart(
		"Branch Chart",
		[]string{"tiny", "small", "one", "kilo", "mega", "giga", "zero"},
		[]float64{0.0001, 0.01, 1, 1200, 2.5e6, 3.4e9, 0},
		"u",
		false,
	)
	if !strings.Contains(chart, "← Best") {
		t.Fatalf("expected best indicator in chart: %q", chart)
	}

	zeroChart := r.RenderBarChart("Zero", []string{"a"}, []float64{0}, "u", true)
	if !strings.Contains(zeroChart, "Zero:") {
		t.Fatalf("unexpected zero chart output: %q", zeroChart)
	}

	// formatThroughput coverage
	if got := formatThroughput(2.4e6); !strings.Contains(got, "M inf/s") {
		t.Fatalf("expected M throughput formatting, got %q", got)
	}
	if got := formatThroughput(2400); !strings.Contains(got, "K inf/s") {
		t.Fatalf("expected K throughput formatting, got %q", got)
	}
	if got := formatThroughput(24); !strings.Contains(got, "inf/s") {
		t.Fatalf("expected base throughput formatting, got %q", got)
	}

	// makeAdvantageBar low/high branches
	if bar := r.makeAdvantageBar(0.5); len(bar) == 0 {
		t.Fatal("expected non-empty bar for sub-1x advantage")
	}
	if bar := r.makeAdvantageBar(1e6); len(bar) == 0 {
		t.Fatal("expected non-empty bar for huge advantage")
	}

	// ScaleToDataCenter branch where chipsRequired would be <1 before clamp
	m := ScaleToDataCenter(FeCIMChip(), 0, MNISTWorkload())
	if m.ChipsRequired != 1 {
		t.Fatalf("expected chips clamp to 1, got %d", m.ChipsRequired)
	}
}
