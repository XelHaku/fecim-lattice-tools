package gui

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestImpactHelpers(t *testing.T) {
	if got := getImpactAssessment(0.001); !strings.Contains(got, "negligible") {
		t.Fatalf("expected negligible assessment, got %q", got)
	}
	if got := getImpactAssessment(0.02); !strings.Contains(got, "Moderate") {
		t.Fatalf("expected moderate assessment, got %q", got)
	}
	if got := getImpactAssessment(0.5); !strings.Contains(got, "Significant") {
		t.Fatalf("expected significant assessment, got %q", got)
	}

	if got := getImpactSummary(0.001); !strings.Contains(got, "Negligible") {
		t.Fatalf("expected negligible summary, got %q", got)
	}
	if got := getImpactSummary(0.02); !strings.Contains(got, "Moderate") {
		t.Fatalf("expected moderate summary, got %q", got)
	}
	if got := getImpactSummary(0.5); !strings.Contains(got, "High") {
		t.Fatalf("expected high summary, got %q", got)
	}
}

func TestTooltips_GracefulAndSeverityBranches(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	_ = arr.ProgramWeight(0, 0, 0.1)
	_ = arr.ProgramWeight(0, 1, 0.9)
	_ = arr.ProgramWeight(1, 0, 0.2)
	_ = arr.ProgramWeight(1, 1, 0.8)

	if got := IRDropTooltipWithArch(0, 0, nil, arr, ""); !strings.Contains(got, "Run MVM") {
		t.Fatalf("expected nil-analysis hint, got %q", got)
	}

	ir := &crossbar.IRDropAnalysis{
		EffectiveVoltage: [][]float64{{0.98, 0.93}, {0.89, 0.70}},
		WorstCaseCell:    [2]int{1, 1},
		MaxIRDrop:        0.30,
		AvgIRDrop:        0.10,
	}

	ok := IRDropTooltipWithArch(0, 0, ir, arr, "")
	moderate := IRDropTooltipWithArch(0, 1, ir, arr, "1T1R")
	high := IRDropTooltipWithArch(1, 1, ir, arr, "2T1R")

	if !strings.Contains(ok, "OK") || !strings.Contains(ok, "0T1R") {
		t.Fatalf("expected OK + default arch, got %q", ok)
	}
	if !strings.Contains(moderate, "Moderate") || !strings.Contains(moderate, "1T1R") {
		t.Fatalf("expected Moderate + custom arch, got %q", moderate)
	}
	if !strings.Contains(high, "High") {
		t.Fatalf("expected High severity, got %q", high)
	}
	if got := IRDropTooltipWithArch(3, 0, ir, arr, ""); got != "Cell out of range" {
		t.Fatalf("expected out-of-range guard, got %q", got)
	}

	if got := SneakPathTooltipWithArch(0, 0, nil, 0, 0, arr, ""); !strings.Contains(got, "Run MVM") {
		t.Fatalf("expected nil-analysis hint, got %q", got)
	}

	sneak := &crossbar.SneakPathAnalysis{
		SneakCurrents: [][]float64{{0.0001, 0.0003}, {0.002, 0.02}},
		TotalSignal:   0.001,
		MaxSneakRatio: 1.5,
	}

	tt := SneakPathTooltipWithArch(0, 0, sneak, 0, 0, arr, "")
	if !strings.Contains(tt, "TARGET") || !strings.Contains(tt, "High") {
		t.Fatalf("expected target/high branch, got %q", tt)
	}
	rowPath := SneakPathTooltipWithArch(0, 1, sneak, 0, 0, arr, "1T1R")
	if !strings.Contains(rowPath, "Row") || !strings.Contains(rowPath, "1T1R: ~1000×") {
		t.Fatalf("expected row + 1T1R note, got %q", rowPath)
	}
	diagCritical := SneakPathTooltipWithArch(1, 1, sneak, 0, 0, arr, "2T1R")
	if !strings.Contains(diagCritical, "Critical") || !strings.Contains(diagCritical, "actual") {
		t.Fatalf("expected critical + capped-ratio note, got %q", diagCritical)
	}
	if got := SneakPathTooltipWithArch(4, 0, sneak, 0, 0, arr, ""); got != "Cell out of range" {
		t.Fatalf("expected out-of-range guard, got %q", got)
	}
}

func TestTooltips_MVMAndComprehensive(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	ir := &crossbar.IRDropAnalysis{EffectiveVoltage: [][]float64{{0.95, 0.9}, {0.85, 0.8}}}
	sneak := &crossbar.SneakPathAnalysis{SneakCurrents: [][]float64{{0.01, 0.02}, {0.03, 0.04}}, TotalSignal: 0.1}
	mvm := &crossbar.MVMResult{
		IdealOutput:      []float64{1.0, 0.0},
		ActualOutput:     []float64{0.9, 0.0},
		RMSE:             0.1,
		AccuracyLoss:     2.0,
		TotalEnergy:      3.0,
		EnergyEfficiency: 25,
		MACOperations:    4,
		Latency:          5,
	}

	if got := MVMResultTooltip(0, mvm); !strings.Contains(got, "Error") {
		t.Fatalf("expected formatted mvm tooltip, got %q", got)
	}
	if got := MVMResultTooltip(1, mvm); !strings.Contains(got, "Error:  0.00%") {
		t.Fatalf("expected zero-ideal branch, got %q", got)
	}
	if got := MVMResultTooltip(99, mvm); got != "Row out of range" {
		t.Fatalf("expected out-of-range guard, got %q", got)
	}

	summary := ComprehensiveTooltip(0, 1, arr, ir, sneak, mvm)
	for _, want := range []string{"SUMMARY", "IR Drop", "Sneak", "Energy"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("expected summary to contain %q, got %q", want, summary)
		}
	}
}
