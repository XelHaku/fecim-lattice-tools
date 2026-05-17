//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestISPP_UsesSharedCalculatorWhenNil(t *testing.T) {
	ds := newTestDeviceState(4, 4)

	ds.mu.Lock()
	ds.isppCalc = nil
	ds.mu.Unlock()

	ds.StartISPP(0, 0, 20, 10)
	start := ds.GetISPPStatus().Voltage
	if start <= 0 {
		t.Fatalf("expected positive start voltage, got %.6f", start)
	}

	result := ds.ISPPIterate(11)
	if result != ISPPResultContinue {
		t.Fatalf("expected ISPPResultContinue, got %v", result)
	}

	status := ds.GetISPPStatus()
	if status.Voltage <= start {
		t.Fatalf("expected voltage to increase after continue step, start=%.6f now=%.6f", start, status.Voltage)
	}

	ds.mu.RLock()
	calc := ds.isppCalc
	ds.mu.RUnlock()
	if calc == nil {
		t.Fatal("expected shared ISPP calculator to be lazily initialized")
	}

	wantStep := ds.GetMaterial().CoerciveVoltage() * 0.01
	gotStep := status.Voltage - start
	if math.Abs(gotStep-wantStep) > 1e-6 {
		t.Fatalf("expected shared step %.6f, got %.6f", wantStep, gotStep)
	}
}
