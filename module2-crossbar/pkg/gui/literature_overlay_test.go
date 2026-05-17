//go:build legacy_fyne

package gui

import (
	"math"
	"strings"
	"testing"
)

func TestReferenceBenchmarks(t *testing.T) {
	benchmarks := referenceBenchmarks()
	if len(benchmarks) != 2 {
		t.Fatalf("expected 2 benchmarks, got %d", len(benchmarks))
	}
	if benchmarks[0].Published != 97.0 {
		t.Fatalf("expected 97.0 for with-limiter, got %.1f", benchmarks[0].Published)
	}
	if benchmarks[1].Published != 9.8 {
		t.Fatalf("expected 9.8 for without-limiter, got %.1f", benchmarks[1].Published)
	}
	for i, b := range benchmarks {
		if b.Reference == "" || b.Condition == "" || b.Metric == "" {
			t.Fatalf("benchmark %d has empty field: %+v", i, b)
		}
	}
}

func TestComputeCrossbarFitMetrics_Empty(t *testing.T) {
	m := ComputeCrossbarFitMetrics(nil)
	if m.NSamples != 0 || m.RMSE != 0 {
		t.Fatalf("expected zero metrics for nil input, got %+v", m)
	}
}

func TestComputeCrossbarFitMetrics_Perfect(t *testing.T) {
	comparisons := []CrossbarLitComparison{
		{Benchmark: CrossbarBenchmark{Published: 97.0}, Simulated: 97.0},
		{Benchmark: CrossbarBenchmark{Published: 9.8}, Simulated: 9.8},
	}
	m := ComputeCrossbarFitMetrics(comparisons)
	if m.NSamples != 2 {
		t.Fatalf("expected 2 samples, got %d", m.NSamples)
	}
	if m.RMSE > 1e-10 || m.MAE > 1e-10 || m.MaxErr > 1e-10 {
		t.Fatalf("expected zero error for perfect match, got RMSE=%.6f MAE=%.6f MaxErr=%.6f",
			m.RMSE, m.MAE, m.MaxErr)
	}
}

func TestComputeCrossbarFitMetrics_KnownError(t *testing.T) {
	comparisons := []CrossbarLitComparison{
		{Benchmark: CrossbarBenchmark{Published: 97.0}, Simulated: 90.0},
	}
	m := ComputeCrossbarFitMetrics(comparisons)
	if m.NSamples != 1 {
		t.Fatalf("expected 1 sample, got %d", m.NSamples)
	}
	expectedErr := 7.0
	if math.Abs(m.RMSE-expectedErr) > 1e-10 {
		t.Fatalf("expected RMSE=%.1f, got %.6f", expectedErr, m.RMSE)
	}
	if math.Abs(m.MAE-expectedErr) > 1e-10 {
		t.Fatalf("expected MAE=%.1f, got %.6f", expectedErr, m.MAE)
	}
	if math.Abs(m.MaxErr-expectedErr) > 1e-10 {
		t.Fatalf("expected MaxErr=%.1f, got %.6f", expectedErr, m.MaxErr)
	}
}

func TestBuildComparisonTable(t *testing.T) {
	comparisons := []CrossbarLitComparison{
		{
			Benchmark: CrossbarBenchmark{
				Reference: "IEEE TED 2022",
				Condition: "With current limiter",
				Published: 97.0,
			},
			Simulated: 89.5,
		},
	}
	table := buildComparisonTable(comparisons)
	if !strings.Contains(table, "Reference") {
		t.Fatal("expected header with Reference column")
	}
	if !strings.Contains(table, "IEEE TED 2022") {
		t.Fatal("expected reference citation in table")
	}
	if !strings.Contains(table, "97.0%") {
		t.Fatal("expected published value in table")
	}
	if !strings.Contains(table, "89.5%") {
		t.Fatal("expected simulated value in table")
	}
	if !strings.Contains(table, "-7.5%") {
		t.Fatal("expected delta value in table")
	}
}

func TestConfidenceBadgeText(t *testing.T) {
	text := confidenceBadgeText()
	if !strings.Contains(text, "External Benchmark") {
		t.Fatalf("expected external benchmark disclaimer, got %q", text)
	}
	if !strings.Contains(text, "not this simulator") {
		t.Fatalf("expected simulator disclaimer, got %q", text)
	}
}
