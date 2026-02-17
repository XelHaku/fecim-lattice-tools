# Memory Leak Fixes (2026-02-07)

> **Note:** This file was previously located at `docs/MEMORY_LEAK_FIXES.md`. It has moved to `docs/3-develop/memory-optimization.md`.

This document records memory leaks identified and fixed in the FeCIM Lattice Tools codebase.

## Summary

**Audit focus:** Long-running simulations and GUI updates  
**Primary issue:** `time.After()` timer leaks when goroutines are cancelled early

## Issues Fixed

### 1. `module3-mnist/pkg/gui/dualmode_demo.go` - Timer Leak in waitOrStop

**Problem:** The `waitOrStop()` function used `time.After(d)` which creates a timer that continues to run even after the channel read is abandoned. When the quick demo is stopped early, each remaining step's timer would leak until it naturally expired.

**Impact:** Memory proportional to (remaining steps × timer duration) leaked per early stop.

**Fix:** Changed to `time.NewTimer()` with deferred `Stop()`:
```go
// Before (leaks)
select {
case <-app.quickDemoStopChan:
    return true
case <-time.After(d):
    return false
}

// After (no leak)
timer := time.NewTimer(d)
defer timer.Stop()
select {
case <-app.quickDemoStopChan:
    return true
case <-timer.C:
    return false
}
```

### 2. `module3-mnist/pkg/gui/app.go` - Auto Demo Ticker Race Condition

**Problem:** The `stopAutoDemoLoop()` function had a problematic pattern:
1. Cancel context
2. Sleep 10ms (to "give goroutine time")
3. Stop ticker

This introduced a race condition and the sleep was a code smell.

**Fix:** Made the goroutine own its ticker cleanup via `defer`:
```go
func (ma *MNISTApp) autoDemoLoop(ctx context.Context) {
    ticker := ma.autoDemoTimer
    defer func() {
        ticker.Stop()
        ma.autoDemoTimer = nil
    }()
    // ... loop using ticker
}
```

This ensures proper cleanup regardless of how the goroutine exits.

### 3. `shared/widgets/demo_controller.go` - Multiple Timer Leaks

**Problem:** 
- `waitOrStop()` used `time.After()` (same issue as #1)
- Pause-checking loop used `time.After(100ms)` in a loop, creating a new timer each iteration

**Fix:**
- Changed `waitOrStop()` to use `time.NewTimer()` with deferred Stop()
- Created a single reusable timer for pause checking:
```go
pauseTimer := time.NewTimer(100 * time.Millisecond)
defer pauseTimer.Stop()

for d.IsPaused() {
    // Reset and reuse same timer
    if !pauseTimer.Stop() {
        select {
        case <-pauseTimer.C:
        default:
        }
    }
    pauseTimer.Reset(100 * time.Millisecond)
    // ...
}
```

### 4. `module4-circuits/pkg/gui/helpers.go` - Sleep Function Timer Leak

**Problem:** The `sleep()` helper used `time.After()`, leaking timers when animations were interrupted.

**Fix:** Changed to `time.NewTimer()` with deferred `Stop()`.

## Best Practices Going Forward

### DO:
```go
// Use time.NewTimer for interruptible waits
timer := time.NewTimer(duration)
defer timer.Stop()
select {
case <-ctx.Done():
    return
case <-timer.C:
    // handle timeout
}
```

### DON'T:
```go
// Avoid time.After in select with cancellation
select {
case <-ctx.Done():
    return  // Timer from time.After keeps running!
case <-time.After(duration):
    // ...
}
```

### For Reusable Timers in Loops:
```go
timer := time.NewTimer(interval)
defer timer.Stop()

for {
    if !timer.Stop() {
        select {
        case <-timer.C:
        default:
        }
    }
    timer.Reset(interval)
    
    select {
    case <-timer.C:
        // handle
    case <-ctx.Done():
        return
    }
}
```

## Testing

All demo controller tests pass after fixes:
```
=== RUN   TestDemoControllerStartStop
=== RUN   TestDemoControllerRunsAllSteps
=== RUN   TestDemoControllerLooping
=== RUN   TestDemoControllerCallbacks
=== RUN   TestDemoControllerWaitOrStop
=== RUN   TestTickerDemoController
=== RUN   TestDemoControllerPauseResume
PASS
```

## Notes on pprof

Go's `pprof` can be used to detect these leaks in long-running applications:

```go
import _ "net/http/pprof"

func init() {
    go http.ListenAndServe("localhost:6060", nil)
}
```

Then access:
- `http://localhost:6060/debug/pprof/heap` - Memory allocations
- `http://localhost:6060/debug/pprof/goroutine` - Goroutine count

Timer leaks would show up as growing heap allocations over time, especially after repeated start/stop cycles of demos.

## Files Modified

1. `module3-mnist/pkg/gui/dualmode_demo.go`
2. `module3-mnist/pkg/gui/app.go`
3. `shared/widgets/demo_controller.go`
4. `module4-circuits/pkg/gui/helpers.go`
