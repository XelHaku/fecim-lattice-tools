package widgets

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

// Global UI lock used to serialize widget mutations during tests (and other
// headless environments) where fyne.Do does not provide a single-threaded UI
// runtime.
//
// This lock is re-entrant per goroutine to avoid deadlocks when safe helpers are
// layered (e.g. SafeRefresh inside SafeDo).
var (
	uiLockMu    sync.Mutex
	uiLockOwner atomic.Uint64
	uiLockDepth atomic.Int32
)

func lockUI() {
	gid := goroutineID()
	if uiLockOwner.Load() == gid {
		uiLockDepth.Add(1)
		return
	}

	uiLockMu.Lock()
	uiLockOwner.Store(gid)
	uiLockDepth.Store(1)
}

func unlockUI() {
	gid := goroutineID()
	owner := uiLockOwner.Load()
	if owner != gid {
		log.Println("widgets: unlockUI called from non-owner goroutine; ignoring unlock request")
		return
	}
	if uiLockDepth.Add(-1) == 0 {
		uiLockOwner.Store(0)
		uiLockMu.Unlock()
	}
}

// WithUILock runs fn while holding the global UI lock.
// This is primarily used by tests to serialize operations like window capture
// against background UI updates.
func WithUILock(fn func()) {
	lockUI()
	defer unlockUI()
	fn()
}

// goroutineID returns the current goroutine ID by parsing runtime.Stack output.
// This is a small internal helper to implement a re-entrant lock.
func goroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)

	// Stack header format: "goroutine 12345 ["
	const prefix = "goroutine "
	var id uint64
	for i := len(prefix); i < n; i++ {
		c := buf[i]
		if c < '0' || c > '9' {
			break
		}
		id = id*10 + uint64(c-'0')
	}
	return id
}
