// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"context"
	"sync"
	"time"
)

// DemoStep represents one step in an automated demo.
type DemoStep struct {
	Name     string        // Step name for logging/display
	Duration time.Duration // How long to wait after this step
	Action   func()        // Action to execute (caller wraps with fyne.Do if needed)
}

// DemoController manages automated demo playback.
// Supports both sequential demos (run once) and looping demos (repeat continuously).
type DemoController struct {
	steps   []DemoStep
	loop    bool // If true, loop back to first step after last
	running bool
	paused  bool
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex

	// Callbacks
	onStart    func()
	onStop     func()
	onStepDone func(stepIndex int, step DemoStep)
}

// NewDemoController creates a new demo controller with the given steps.
// By default, the demo runs once (no looping).
func NewDemoController(steps []DemoStep) *DemoController {
	return &DemoController{
		steps: steps,
		loop:  false,
	}
}

// NewLoopingDemoController creates a demo controller that loops continuously.
func NewLoopingDemoController(steps []DemoStep) *DemoController {
	return &DemoController{
		steps: steps,
		loop:  true,
	}
}

// SetLoop enables or disables looping.
func (d *DemoController) SetLoop(loop bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.loop = loop
}

// SetOnStart sets a callback called when the demo starts.
func (d *DemoController) SetOnStart(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onStart = fn
}

// SetOnStop sets a callback called when the demo stops.
func (d *DemoController) SetOnStop(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onStop = fn
}

// SetOnStepDone sets a callback called after each step completes.
func (d *DemoController) SetOnStepDone(fn func(stepIndex int, step DemoStep)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onStepDone = fn
}

// Start begins the demo playback.
// If already running, this is a no-op.
func (d *DemoController) Start() {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return
	}
	d.running = true
	d.paused = false
	d.ctx, d.cancel = context.WithCancel(context.Background())

	onStart := d.onStart
	d.mu.Unlock()

	if onStart != nil {
		onStart()
	}

	go d.run()
}

// Stop halts the demo playback.
func (d *DemoController) Stop() {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}

	d.running = false
	d.paused = false
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}

	onStop := d.onStop
	d.mu.Unlock()

	if onStop != nil {
		onStop()
	}
}

// Pause pauses the demo (if running).
func (d *DemoController) Pause() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.running {
		d.paused = true
	}
}

// Resume resumes a paused demo.
func (d *DemoController) Resume() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.running {
		d.paused = false
	}
}

// IsRunning returns true if the demo is currently running.
func (d *DemoController) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

// IsPaused returns true if the demo is paused.
func (d *DemoController) IsPaused() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.paused
}

// run executes demo steps.
func (d *DemoController) run() {
	// Create a single pause-check timer to avoid memory leaks from repeated time.After calls
	pauseTimer := time.NewTimer(100 * time.Millisecond)
	defer pauseTimer.Stop()

	for {
		for i, step := range d.steps {
			// Check if stopped
			d.mu.Lock()
			if !d.running {
				d.mu.Unlock()
				return
			}
			ctx := d.ctx
			d.mu.Unlock()

			// Wait while paused - reuse timer to avoid memory leaks
			for d.IsPaused() {
				// Reset timer for next pause check
				if !pauseTimer.Stop() {
					select {
					case <-pauseTimer.C:
					default:
					}
				}
				pauseTimer.Reset(100 * time.Millisecond)

				select {
				case <-ctx.Done():
					return
				case <-pauseTimer.C:
					// Check pause state again
				}
			}

			// Check if stopped again after pause
			if !d.IsRunning() {
				return
			}

			// Execute step action
			if step.Action != nil {
				step.Action()
			}

			// Notify step done
			d.mu.Lock()
			onStepDone := d.onStepDone
			d.mu.Unlock()
			if onStepDone != nil {
				onStepDone(i, step)
			}

			// Wait for step duration or stop
			if !d.waitOrStop(ctx, step.Duration) {
				return
			}
		}

		// Check if we should loop
		d.mu.Lock()
		shouldLoop := d.loop && d.running
		d.mu.Unlock()

		if !shouldLoop {
			// Demo complete
			d.Stop()
			return
		}
	}
}

// waitOrStop waits for duration or returns false if stopped.
// Uses time.NewTimer instead of time.After to avoid memory leaks when stopped early.
func (d *DemoController) waitOrStop(ctx context.Context, duration time.Duration) bool {
	if duration <= 0 {
		return d.IsRunning()
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return d.IsRunning()
	}
}

// WaitOrStop is a public helper for custom demo logic.
// Returns true if the wait completed normally, false if stopped.
func (d *DemoController) WaitOrStop(duration time.Duration) bool {
	d.mu.Lock()
	ctx := d.ctx
	running := d.running
	d.mu.Unlock()

	if !running || ctx == nil {
		return false
	}

	return d.waitOrStop(ctx, duration)
}

// TickerDemoController provides a simpler ticker-based demo pattern.
// Useful for demos that execute the same action repeatedly at fixed intervals.
type TickerDemoController struct {
	interval time.Duration
	action   func(step int)
	step     int
	maxSteps int // 0 = infinite
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
}

// NewTickerDemoController creates a ticker-based demo that calls action every interval.
// Set maxSteps to 0 for infinite looping.
func NewTickerDemoController(interval time.Duration, maxSteps int, action func(step int)) *TickerDemoController {
	return &TickerDemoController{
		interval: interval,
		action:   action,
		maxSteps: maxSteps,
	}
}

// Start begins the ticker demo.
func (t *TickerDemoController) Start() {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.step = 0
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.mu.Unlock()

	go t.run()
}

// Stop halts the ticker demo.
func (t *TickerDemoController) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	t.running = false
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
}

// IsRunning returns true if the ticker demo is running.
func (t *TickerDemoController) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

// GetStep returns the current step number.
func (t *TickerDemoController) GetStep() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.step
}

// run executes the ticker loop.
func (t *TickerDemoController) run() {
	// Run first action immediately
	t.mu.Lock()
	action := t.action
	t.mu.Unlock()

	if action != nil {
		action(0)
	}

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		t.mu.Lock()
		ctx := t.ctx
		running := t.running
		t.mu.Unlock()

		if !running {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			if !t.running {
				t.mu.Unlock()
				return
			}
			t.step++
			step := t.step
			maxSteps := t.maxSteps
			action := t.action
			t.mu.Unlock()

			// Check max steps (0 = infinite)
			if maxSteps > 0 && step >= maxSteps {
				t.Stop()
				return
			}

			if action != nil {
				action(step)
			}
		}
	}
}
