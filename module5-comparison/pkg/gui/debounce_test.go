//go:build legacy_fyne

package gui

import (
	"sync/atomic"
	"testing"
	"time"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

func TestModule5RecomputeDebounce_CoalescesBursts(t *testing.T) {
	bus := sharedwidgets.NewCoalesceBus(20 * time.Millisecond)
	defer bus.Close()

	var calls int32
	for i := 0; i < 20; i++ {
		bus.Submit("module5-calculation", func() {
			atomic.AddInt32(&calls, 1)
		})
	}
	time.Sleep(60 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 debounced call, got %d", got)
	}
}
