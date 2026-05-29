package circuits

import "testing"

func TestReferenceSpecWorkflowComputesDefaultSpecSummary(t *testing.T) {
	updated := newReferenceSpecWorkflow(CircuitsState{
		Rows:          8,
		Cols:          8,
		QuantLevels:   DefaultQuantLevels,
		DACResolution: 5,
		ADCResolution: 5,
	}).compute()

	if updated.SpecCells != 64 {
		t.Fatalf("cells = %d, want 64", updated.SpecCells)
	}
	if updated.SpecDACCount != 8 || updated.SpecTIACount != 8 || updated.SpecADCCount != 8 {
		t.Fatalf("component counts = DAC %d TIA %d ADC %d", updated.SpecDACCount, updated.SpecTIACount, updated.SpecADCCount)
	}
	if updated.SpecDACCodes != 32 || updated.SpecADCCodes != 32 {
		t.Fatalf("converter codes = DAC %d ADC %d", updated.SpecDACCodes, updated.SpecADCCodes)
	}
	if !near(updated.SpecBitsPerCell, 4.9068905956, 1e-9) {
		t.Fatalf("bits/cell = %.10f, want log2(30)", updated.SpecBitsPerCell)
	}
	if !near(updated.SpecTotalPowerMW, 5.8, 1e-9) || !near(updated.SpecLatencyNS, 76, 1e-9) {
		t.Fatalf("power/latency = %.10f / %.10f", updated.SpecTotalPowerMW, updated.SpecLatencyNS)
	}
	if !near(updated.SpecThroughputGOPS, 0.8421052632, 1e-9) || !near(updated.SpecEfficiencyGOPSW, 145.1905626134, 1e-9) {
		t.Fatalf("throughput/efficiency = %.10f / %.10f", updated.SpecThroughputGOPS, updated.SpecEfficiencyGOPSW)
	}
	if updated.SpecCompliance != "OK: DAC/ADC cover 30 levels" {
		t.Fatalf("compliance = %q", updated.SpecCompliance)
	}
}

func TestReferenceSpecWorkflowComputesResizedLowDACCompliance(t *testing.T) {
	updated := newReferenceSpecWorkflow(CircuitsState{
		Rows:          32,
		Cols:          32,
		QuantLevels:   DefaultQuantLevels,
		DACResolution: 4,
		ADCResolution: 5,
	}).compute()

	if updated.SpecCells != 1024 {
		t.Fatalf("cells = %d, want 1024", updated.SpecCells)
	}
	if updated.SpecDACCount != 32 || updated.SpecTIACount != 32 || updated.SpecADCCount != 32 {
		t.Fatalf("component counts = DAC %d TIA %d ADC %d", updated.SpecDACCount, updated.SpecTIACount, updated.SpecADCCount)
	}
	if updated.SpecDACCodes != 16 || updated.SpecADCCodes != 32 {
		t.Fatalf("converter codes = DAC %d ADC %d", updated.SpecDACCodes, updated.SpecADCCodes)
	}
	if !near(updated.SpecTotalPowerMW, 21.4, 1e-9) {
		t.Fatalf("total power = %.10f, want 21.4", updated.SpecTotalPowerMW)
	}
	if !near(updated.SpecThroughputGOPS, 13.4736842105, 1e-9) || !near(updated.SpecEfficiencyGOPSW, 629.6114117063, 1e-9) {
		t.Fatalf("throughput/efficiency = %.10f / %.10f", updated.SpecThroughputGOPS, updated.SpecEfficiencyGOPSW)
	}
	if updated.SpecCompliance != "CHECK: DAC 16 codes < 30 levels" {
		t.Fatalf("compliance = %q", updated.SpecCompliance)
	}
}
