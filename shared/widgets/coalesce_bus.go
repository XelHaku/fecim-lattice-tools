//go:build legacy_fyne

package widgets

import (
	"sync"
	"time"
)

type coalesceEntry struct {
	timer      *time.Timer
	fn         func()
	generation uint64
}

// CoalesceBus debounces bursty update requests and only executes the last update per key.
type CoalesceBus struct {
	window  time.Duration
	mu      sync.Mutex
	pending map[string]*coalesceEntry
	closed  bool
}

// NewCoalesceBus creates a bus with the debounce window (recommended 30-50ms).
func NewCoalesceBus(window time.Duration) *CoalesceBus {
	if window <= 0 {
		window = 40 * time.Millisecond
	}
	return &CoalesceBus{window: window, pending: make(map[string]*coalesceEntry)}
}

// Submit debounces updates by key. Within the debounce window only the latest callback runs.
func (b *CoalesceBus) Submit(key string, fn func()) {
	if fn == nil {
		return
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	entry := b.pending[key]
	if entry == nil {
		entry = &coalesceEntry{}
		b.pending[key] = entry
	}
	entry.fn = fn
	entry.generation++
	generation := entry.generation
	if entry.timer != nil {
		entry.timer.Stop()
	}
	entry.timer = time.AfterFunc(b.window, func() {
		b.fire(key, entry, generation)
	})
	b.mu.Unlock()
}

func (b *CoalesceBus) fire(key string, entry *coalesceEntry, generation uint64) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}

	current := b.pending[key]
	if current != entry || current == nil || current.generation != generation {
		b.mu.Unlock()
		return
	}

	fn := current.fn
	delete(b.pending, key)
	b.mu.Unlock()

	if fn != nil {
		fn()
	}
}

// Close cancels pending work and rejects new submissions.
func (b *CoalesceBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	for _, entry := range b.pending {
		if entry != nil && entry.timer != nil {
			entry.timer.Stop()
		}
	}
	b.pending = map[string]*coalesceEntry{}
	b.closed = true
}
