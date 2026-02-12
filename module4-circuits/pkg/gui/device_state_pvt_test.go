package gui

import (
	"testing"

	"fecim-lattice-tools/shared/peripherals"
)

func TestDeviceState_PeripheralPVT_AffectsDACWriteVoltage(t *testing.T) {
	ds := newTestDeviceState(4, 4)
	ds.SetDACNonlinearity(true)

	target := 0.6
	ds.SetPeripheralPVT(300, peripherals.CornerTypical)
	nominal, _ := ds.DACWriteVoltage(target)

	ds.SetPeripheralPVT(400, peripherals.CornerSlow)
	hotSlow, _ := ds.DACWriteVoltage(target)

	if hotSlow == nominal {
		t.Fatalf("expected PVT-conditioned DAC voltage to differ, both %.6f", nominal)
	}
}

func TestISPPState_HistoryTracksIterations(t *testing.T) {
	ds := newTestDeviceState(4, 4)
	ds.StartISPP(0, 0, 10, 0)

	_ = ds.ISPPIterate(1)
	_ = ds.ISPPIterate(2)

	status := ds.GetISPPStatus()
	if len(status.History) < 3 {
		t.Fatalf("expected at least 3 history points (start + 2 iters), got %d", len(status.History))
	}
	last := status.History[len(status.History)-1]
	if last.Level != 2 {
		t.Fatalf("expected final history level 2, got %d", last.Level)
	}
}
