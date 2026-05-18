//go:build !cgo

package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
)

func TestCircuitsOverlayStateIncludesHalfSelectStress(t *testing.T) {
	vm := circuitsvm.New()
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetOperationMode,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"mode": circuitsvm.OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.halfSelectState != circuitsvm.HalfSelectStateColumnWriteActive {
		t.Fatalf("halfSelectState = %q, want %q", state.halfSelectState, circuitsvm.HalfSelectStateColumnWriteActive)
	}
	if state.halfSelectCells != 7 {
		t.Fatalf("halfSelectCells = %d, want 7", state.halfSelectCells)
	}
	if state.stressBudget != "400 pulses/level" {
		t.Fatalf("stressBudget = %q, want 400 pulses/level", state.stressBudget)
	}
}

func TestCircuitsOverlayStateIncludesPVTInvestigationSummaries(t *testing.T) {
	vm := circuitsvm.New()
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.pvtProcessYield != "100.0% (20/20)" {
		t.Fatalf("pvtProcessYield = %q, want 100.0%% (20/20)", state.pvtProcessYield)
	}
	if state.pvtTempSweep != "pass -40/25/85/125 C" {
		t.Fatalf("pvtTempSweep = %q, want temperature sweep summary", state.pvtTempSweep)
	}
	if state.pvtCornerENOB != "FF 4.51 / TT 4.42 / SS 4.30 bits" {
		t.Fatalf("pvtCornerENOB = %q, want FF/TT/SS summary", state.pvtCornerENOB)
	}
	if state.pvtNoiseCeiling != "13.61 bits at 16-bit ADC" {
		t.Fatalf("pvtNoiseCeiling = %q, want thermal-noise ceiling", state.pvtNoiseCeiling)
	}
}

func TestCircuitsOverlayStateIncludesReferenceSpecSummaries(t *testing.T) {
	vm := circuitsvm.New()
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.specPowerLatency != "5.8 mW / 76 ns" {
		t.Fatalf("specPowerLatency = %q, want reference power/latency", state.specPowerLatency)
	}
	if state.specThroughput != "0.84 GOPS / 145 GOPS/W" {
		t.Fatalf("specThroughput = %q, want reference throughput", state.specThroughput)
	}
	if state.specCompliance != "OK: DAC/ADC cover 30 levels" {
		t.Fatalf("specCompliance = %q, want compliance summary", state.specCompliance)
	}
}

func TestCircuitsOverlayStateIncludesReferenceTimingSummaries(t *testing.T) {
	vm := circuitsvm.New()
	state := circuitsOverlayStateFromSnapshot(vm.Snapshot())

	if state.timingActive != "READ 76 ns total" {
		t.Fatalf("timingActive = %q, want READ timing summary", state.timingActive)
	}
	if state.timingActivePhases != "DAC 10 / Array 5 / TIA 11 / ADC 50 ns" {
		t.Fatalf("timingActivePhases = %q, want read timing phases", state.timingActivePhases)
	}
}
