//go:build legacy_fyne

package widgets

import (
	"testing"
	"time"
)

func TestFrameWatchdog_SimulatedSlowRender(t *testing.T) {
	watchdog := NewFrameWatchdog()
	base := time.Unix(100, 0)
	watchdog.now = func() time.Time { return base }

	watchdog.StartFrame()
	watchdog.now = func() time.Time { return base.Add(40 * time.Millisecond) }
	sample := watchdog.EndFrame()

	if sample.Severity != FrameCritical {
		t.Fatalf("expected critical severity for slow render, got %s", sample.Severity)
	}
}

func TestFrameWatchdog_Thresholds(t *testing.T) {
	if got := classifyFrame(10 * time.Millisecond); got != FrameOK {
		t.Fatalf("10ms should be ok, got %s", got)
	}
	if got := classifyFrame(20 * time.Millisecond); got != FrameWarn {
		t.Fatalf("20ms should be warn, got %s", got)
	}
	if got := classifyFrame(40 * time.Millisecond); got != FrameCritical {
		t.Fatalf("40ms should be critical, got %s", got)
	}
}
