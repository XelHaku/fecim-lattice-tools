package circuits

import "testing"

func TestReferenceTimingWorkflowComputesActiveReadTiming(t *testing.T) {
	state := CircuitsState{TimingOperation: "READ"}

	updated := newReferenceTimingWorkflow(state).compute()

	if updated.TimingReadTotalNS != 76 || updated.TimingWriteTotalNS != 203 || updated.TimingComputeTotalNS != 76 {
		t.Fatalf("timing totals = write %d read %d compute %d", updated.TimingWriteTotalNS, updated.TimingReadTotalNS, updated.TimingComputeTotalNS)
	}
	if updated.TimingActiveOp != "READ" {
		t.Fatalf("active op = %q, want READ", updated.TimingActiveOp)
	}
	if updated.TimingActiveTotalNS != 76 {
		t.Fatalf("active total = %d, want 76", updated.TimingActiveTotalNS)
	}
	if updated.TimingActivePhases != "DAC 10 / Array 5 / TIA 11 / ADC 50 ns" {
		t.Fatalf("active phases = %q", updated.TimingActivePhases)
	}
	if len(updated.TimingWaveforms) != 3 {
		t.Fatalf("waveforms = %d, want 3", len(updated.TimingWaveforms))
	}
}

func TestReferenceTimingWorkflowComputesWriteAndComputeTiming(t *testing.T) {
	for _, tc := range []struct {
		operation string
		wantTotal int
		wantPhase string
	}{
		{operation: "WRITE", wantTotal: 203, wantPhase: "DAC 10 / Pump 88 / Pulse 100 / Array 5 ns"},
		{operation: "COMPUTE", wantTotal: 76, wantPhase: "DAC 10 / Array 5 / TIA+ADC 61 ns"},
	} {
		t.Run(tc.operation, func(t *testing.T) {
			updated := newReferenceTimingWorkflow(CircuitsState{TimingOperation: tc.operation}).compute()
			if updated.TimingActiveOp != tc.operation {
				t.Fatalf("active op = %q, want %q", updated.TimingActiveOp, tc.operation)
			}
			if updated.TimingActiveTotalNS != tc.wantTotal {
				t.Fatalf("active total = %d, want %d", updated.TimingActiveTotalNS, tc.wantTotal)
			}
			if updated.TimingActivePhases != tc.wantPhase {
				t.Fatalf("active phases = %q, want %q", updated.TimingActivePhases, tc.wantPhase)
			}
		})
	}
}
