// Package progress provides progress tracking for long-running operations in FeCIM tools.
// It supports progress bars with ETA calculation, cancellation, and detailed status messages
// for simulations and exports.
package progress

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents the current state of an operation
type State int

const (
	StateIdle State = iota
	StateRunning
	StatePaused
	StateCancelled
	StateCompleted
	StateFailed
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateCancelled:
		return "cancelled"
	case StateCompleted:
		return "completed"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ProgressCallback is called whenever progress updates
type ProgressCallback func(p *Progress)

// Progress tracks the progress of a long-running operation with ETA calculation,
// cancellation support, and detailed status messages.
type Progress struct {
	mu sync.RWMutex

	// Core progress tracking
	current   int64 // Current progress value
	total     int64 // Total expected value (0 = indeterminate)
	state     State
	startTime time.Time
	endTime   time.Time

	// Status messages
	operation   string // e.g., "Simulation", "Export", "Training"
	phase       string // e.g., "Initializing", "Processing batch 5/10"
	detail      string // e.g., "Computing hysteresis for cell (3,4)"
	lastMessage string // Last status message for logging

	// ETA tracking
	samples     []etaSample // Ring buffer of recent samples for smoothed ETA
	sampleIdx   int
	sampleCount int

	// Cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Callbacks
	onProgress []ProgressCallback
	onComplete func(p *Progress)
	onError    func(p *Progress, err error)

	// Error tracking
	lastError error
}

// etaSample stores a progress sample for ETA calculation
type etaSample struct {
	progress int64
	time     time.Time
}

const maxSamples = 20 // Number of samples for smoothed ETA

// NewProgress creates a new progress tracker for an operation.
// If total is 0, the progress is indeterminate (spinner mode).
func NewProgress(operation string, total int64) *Progress {
	ctx, cancel := context.WithCancel(context.Background())
	return &Progress{
		operation: operation,
		total:     total,
		state:     StateIdle,
		ctx:       ctx,
		cancel:    cancel,
		samples:   make([]etaSample, maxSamples),
	}
}

// NewProgressWithContext creates a progress tracker using an existing context.
// This allows external cancellation (e.g., from a parent context).
func NewProgressWithContext(ctx context.Context, operation string, total int64) *Progress {
	childCtx, cancel := context.WithCancel(ctx)
	return &Progress{
		operation: operation,
		total:     total,
		state:     StateIdle,
		ctx:       childCtx,
		cancel:    cancel,
		samples:   make([]etaSample, maxSamples),
	}
}

// Start begins tracking progress
func (p *Progress) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = StateRunning
	p.startTime = time.Now()
	p.current = 0
	p.sampleIdx = 0
	p.sampleCount = 0
	p.addSample(0)
	p.notify()
}

// SetTotal updates the total expected value (useful for dynamic totals)
func (p *Progress) SetTotal(total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
	p.notify()
}

// SetPhase sets the current phase description
func (p *Progress) SetPhase(phase string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phase = phase
	p.notify()
}

// SetDetail sets the detailed status message
func (p *Progress) SetDetail(detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.detail = detail
	p.notify()
}

// SetStatus sets both phase and detail at once
func (p *Progress) SetStatus(phase, detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phase = phase
	p.detail = detail
	p.notify()
}

// Update sets the current progress value
func (p *Progress) Update(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StateRunning {
		return
	}

	p.current = current
	p.addSample(current)
	p.notify()
}

// Increment increases progress by delta (default 1)
func (p *Progress) Increment(delta ...int64) {
	d := int64(1)
	if len(delta) > 0 {
		d = delta[0]
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StateRunning {
		return
	}

	p.current += d
	p.addSample(p.current)
	p.notify()
}

// UpdateWithStatus updates progress and status in one call
func (p *Progress) UpdateWithStatus(current int64, phase, detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StateRunning {
		return
	}

	p.current = current
	p.phase = phase
	p.detail = detail
	p.addSample(current)
	p.notify()
}

// Pause pauses the operation
func (p *Progress) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StateRunning {
		p.state = StatePaused
		p.notify()
	}
}

// Resume resumes a paused operation
func (p *Progress) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StatePaused {
		p.state = StateRunning
		p.notify()
	}
}

// Cancel cancels the operation
func (p *Progress) Cancel() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StateRunning || p.state == StatePaused {
		p.state = StateCancelled
		p.endTime = time.Now()
		p.cancel()
		p.notify()
	}
}

// Complete marks the operation as successfully completed
func (p *Progress) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StateRunning || p.state == StatePaused {
		p.state = StateCompleted
		p.endTime = time.Now()
		if p.total > 0 {
			p.current = p.total
		}
		p.notifyComplete()
		p.notify()
	}
}

// Fail marks the operation as failed with an error
func (p *Progress) Fail(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = StateFailed
	p.endTime = time.Now()
	p.lastError = err
	p.cancel()
	p.notifyError(err)
	p.notify()
}

// Context returns the cancellation context
func (p *Progress) Context() context.Context {
	return p.ctx
}

// IsCancelled returns true if the operation was cancelled
func (p *Progress) IsCancelled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state == StateCancelled
}

// IsRunning returns true if the operation is running
func (p *Progress) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state == StateRunning
}

// Current returns the current progress value
func (p *Progress) Current() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

// Total returns the total expected value
func (p *Progress) Total() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.total
}

// Fraction returns progress as a fraction (0.0-1.0)
// Returns 0 for indeterminate progress
func (p *Progress) Fraction() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.total <= 0 {
		return 0
	}
	f := float64(p.current) / float64(p.total)
	if f > 1 {
		f = 1
	}
	return f
}

// Percent returns progress as a percentage (0-100)
func (p *Progress) Percent() float64 {
	return p.Fraction() * 100
}

// IsIndeterminate returns true if total is unknown
func (p *Progress) IsIndeterminate() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.total <= 0
}

// State returns the current operation state
func (p *Progress) State() State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// Operation returns the operation name
func (p *Progress) Operation() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.operation
}

// Phase returns the current phase
func (p *Progress) Phase() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.phase
}

// Detail returns the detailed status
func (p *Progress) Detail() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.detail
}

// Error returns the last error (if any)
func (p *Progress) Error() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastError
}

// Elapsed returns the elapsed time since start
func (p *Progress) Elapsed() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.startTime.IsZero() {
		return 0
	}
	if !p.endTime.IsZero() {
		return p.endTime.Sub(p.startTime)
	}
	return time.Since(p.startTime)
}

// ETA returns the estimated time remaining
// Returns 0 if indeterminate or not enough samples
func (p *Progress) ETA() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.calculateETA()
}

// ETATime returns the estimated completion time
func (p *Progress) ETATime() time.Time {
	eta := p.ETA()
	if eta <= 0 {
		return time.Time{}
	}
	return time.Now().Add(eta)
}

// Rate returns the current processing rate (items per second)
func (p *Progress) Rate() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.sampleCount < 2 {
		return 0
	}

	// Use oldest and newest samples for rate calculation
	oldest := p.samples[(p.sampleIdx-p.sampleCount+1+maxSamples)%maxSamples]
	newest := p.samples[(p.sampleIdx+maxSamples)%maxSamples]

	duration := newest.time.Sub(oldest.time).Seconds()
	if duration <= 0 {
		return 0
	}

	return float64(newest.progress-oldest.progress) / duration
}

// StatusLine returns a formatted status line for display
func (p *Progress) StatusLine() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var sb string

	// State indicator
	switch p.state {
	case StateRunning:
		sb = "⏳ "
	case StatePaused:
		sb = "⏸️ "
	case StateCancelled:
		sb = "❌ "
	case StateCompleted:
		sb = "✅ "
	case StateFailed:
		sb = "❗ "
	default:
		sb = "⏹️ "
	}

	// Operation and phase
	sb += p.operation
	if p.phase != "" {
		sb += ": " + p.phase
	}

	// Progress percentage
	if p.total > 0 {
		pct := float64(p.current) / float64(p.total) * 100
		sb += fmt.Sprintf(" (%.1f%%)", pct)
	}

	// ETA
	if p.state == StateRunning && p.total > 0 {
		eta := p.calculateETA()
		if eta > 0 {
			sb += fmt.Sprintf(" - ETA: %s", formatDuration(eta))
		}
	}

	// Detail
	if p.detail != "" {
		sb += " - " + p.detail
	}

	return sb
}

// OnProgress registers a callback for progress updates
func (p *Progress) OnProgress(fn ProgressCallback) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onProgress = append(p.onProgress, fn)
}

// OnComplete registers a callback for completion
func (p *Progress) OnComplete(fn func(p *Progress)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onComplete = fn
}

// OnError registers a callback for errors
func (p *Progress) OnError(fn func(p *Progress, err error)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onError = fn
}

// addSample adds a progress sample for ETA calculation (must hold lock)
func (p *Progress) addSample(progress int64) {
	p.sampleIdx = (p.sampleIdx + 1) % maxSamples
	p.samples[p.sampleIdx] = etaSample{
		progress: progress,
		time:     time.Now(),
	}
	if p.sampleCount < maxSamples {
		p.sampleCount++
	}
}

// calculateETA calculates estimated time remaining (must hold lock)
func (p *Progress) calculateETA() time.Duration {
	if p.total <= 0 || p.current >= p.total || p.sampleCount < 2 {
		return 0
	}

	// Use linear regression on recent samples for smoothed ETA
	rate := p.calculateRate()
	if rate <= 0 {
		return 0
	}

	remaining := float64(p.total - p.current)
	return time.Duration(remaining/rate) * time.Second
}

// calculateRate calculates items per second from samples (must hold lock)
func (p *Progress) calculateRate() float64 {
	if p.sampleCount < 2 {
		return 0
	}

	// Use oldest and newest samples
	oldest := p.samples[(p.sampleIdx-p.sampleCount+1+maxSamples)%maxSamples]
	newest := p.samples[(p.sampleIdx+maxSamples)%maxSamples]

	duration := newest.time.Sub(oldest.time).Seconds()
	if duration <= 0 {
		return 0
	}

	return float64(newest.progress-oldest.progress) / duration
}

// notify calls all registered progress callbacks (must hold lock)
func (p *Progress) notify() {
	for _, fn := range p.onProgress {
		go fn(p)
	}
}

// notifyComplete calls the completion callback (must hold lock)
func (p *Progress) notifyComplete() {
	if p.onComplete != nil {
		go p.onComplete(p)
	}
}

// notifyError calls the error callback (must hold lock)
func (p *Progress) notifyError(err error) {
	if p.onError != nil {
		go p.onError(p, err)
	}
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

// ProgressInfo returns a snapshot of the current progress state
type ProgressInfo struct {
	Operation   string
	Phase       string
	Detail      string
	Current     int64
	Total       int64
	Percent     float64
	State       State
	Elapsed     time.Duration
	ETA         time.Duration
	Rate        float64
	Error       error
	StartTime   time.Time
	EndTime     time.Time
}

// Info returns a snapshot of the current progress state
func (p *Progress) Info() ProgressInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var pct float64
	if p.total > 0 {
		pct = float64(p.current) / float64(p.total) * 100
	}

	return ProgressInfo{
		Operation: p.operation,
		Phase:     p.phase,
		Detail:    p.detail,
		Current:   p.current,
		Total:     p.total,
		Percent:   pct,
		State:     p.state,
		Elapsed:   p.Elapsed(),
		ETA:       p.calculateETA(),
		Rate:      p.calculateRate(),
		Error:     p.lastError,
		StartTime: p.startTime,
		EndTime:   p.endTime,
	}
}
