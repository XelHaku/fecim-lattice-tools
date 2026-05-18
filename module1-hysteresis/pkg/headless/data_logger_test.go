package headless

import (
	"testing"
	"time"
)

func TestRecordBlocksForISPPTargetBoundarySnapshots(t *testing.T) {
	logger := &HysteresisDataLogger{rows: make(chan HysteresisSnapshot, 1)}
	logger.Record(HysteresisSnapshot{Waveform: "ISPP", WrdTargetLevel: 3, WrdPhaseTimer: 0.1})

	done := make(chan struct{})
	go func() {
		logger.Record(HysteresisSnapshot{Waveform: "ISPP", WrdTargetLevel: 15, WrdPhaseTimer: 0})
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("target-boundary snapshot returned while queue was full; want it to wait until recorded")
	case <-time.After(20 * time.Millisecond):
	}

	<-logger.rows

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("target-boundary snapshot stayed blocked after queue was drained")
	}

	got := <-logger.rows
	if got.WrdTargetLevel != 15 {
		t.Fatalf("recorded target = %d, want 15", got.WrdTargetLevel)
	}
}
