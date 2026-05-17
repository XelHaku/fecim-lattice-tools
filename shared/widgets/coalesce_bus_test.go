//go:build legacy_fyne

package widgets

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoalesceBusRapidFireExecutesOnlyLast(t *testing.T) {
	bus := NewCoalesceBus(35 * time.Millisecond)
	defer bus.Close()

	var last atomic.Int64
	var called atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)

	for i := int64(1); i <= 20; i++ {
		idx := i
		bus.Submit("status", func() {
			last.Store(idx)
			if called.Add(1) == 1 {
				wg.Done()
			}
		})
		time.Sleep(2 * time.Millisecond)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for coalesced update")
	}

	if got := called.Load(); got != 1 {
		t.Fatalf("callback invocation count mismatch: got %d want 1", got)
	}
	if got := last.Load(); got != 20 {
		t.Fatalf("last callback value mismatch: got %d want 20", got)
	}
}

func TestCoalesceBusSeparateKeysAreIndependent(t *testing.T) {
	bus := NewCoalesceBus(20 * time.Millisecond)
	defer bus.Close()

	var a, b atomic.Int64
	bus.Submit("a", func() { a.Add(1) })
	bus.Submit("b", func() { b.Add(1) })

	time.Sleep(80 * time.Millisecond)

	if a.Load() != 1 || b.Load() != 1 {
		t.Fatalf("expected both keys to execute once, got a=%d b=%d", a.Load(), b.Load())
	}
}
