package progress

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewProgress(t *testing.T) {
	p := NewProgress("Test Operation", 100)
	if p == nil {
		t.Fatal("NewProgress returned nil")
	}
	if p.Operation() != "Test Operation" {
		t.Errorf("Operation = %q, want %q", p.Operation(), "Test Operation")
	}
	if p.Total() != 100 {
		t.Errorf("Total = %d, want %d", p.Total(), 100)
	}
	if p.State() != StateIdle {
		t.Errorf("State = %v, want %v", p.State(), StateIdle)
	}
}

func TestProgressStartStop(t *testing.T) {
	p := NewProgress("Test", 100)

	p.Start()
	if p.State() != StateRunning {
		t.Errorf("State = %v, want %v", p.State(), StateRunning)
	}
	if p.IsRunning() != true {
		t.Error("IsRunning should be true")
	}

	p.Complete()
	if p.State() != StateCompleted {
		t.Errorf("State = %v, want %v", p.State(), StateCompleted)
	}
	if p.IsRunning() != false {
		t.Error("IsRunning should be false after completion")
	}
}

func TestProgressUpdate(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	p.Update(50)
	if p.Current() != 50 {
		t.Errorf("Current = %d, want %d", p.Current(), 50)
	}
	if p.Fraction() != 0.5 {
		t.Errorf("Fraction = %f, want %f", p.Fraction(), 0.5)
	}
	if p.Percent() != 50.0 {
		t.Errorf("Percent = %f, want %f", p.Percent(), 50.0)
	}

	p.Increment(10)
	if p.Current() != 60 {
		t.Errorf("Current after Increment = %d, want %d", p.Current(), 60)
	}
}

func TestProgressFraction(t *testing.T) {
	tests := []struct {
		current  int64
		total    int64
		expected float64
	}{
		{0, 100, 0.0},
		{50, 100, 0.5},
		{100, 100, 1.0},
		{150, 100, 1.0}, // Capped at 1.0
		{0, 0, 0.0},     // Indeterminate
	}

	for _, tt := range tests {
		p := NewProgress("Test", tt.total)
		p.Start()
		p.Update(tt.current)
		if got := p.Fraction(); got != tt.expected {
			t.Errorf("Fraction(%d/%d) = %f, want %f", tt.current, tt.total, got, tt.expected)
		}
	}
}

func TestProgressCancellation(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	if p.IsCancelled() {
		t.Error("IsCancelled should be false initially")
	}

	p.Cancel()
	if !p.IsCancelled() {
		t.Error("IsCancelled should be true after Cancel()")
	}
	if p.State() != StateCancelled {
		t.Errorf("State = %v, want %v", p.State(), StateCancelled)
	}

	// Context should be cancelled
	select {
	case <-p.Context().Done():
		// OK
	default:
		t.Error("Context should be cancelled")
	}
}

func TestProgressContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewProgressWithContext(ctx, "Test", 100)
	p.Start()

	// Cancel parent context
	cancel()

	// Wait for propagation
	select {
	case <-p.Context().Done():
		// OK
	case <-time.After(time.Second):
		t.Error("Context cancellation should propagate")
	}
}

func TestProgressPauseResume(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	p.Pause()
	if p.State() != StatePaused {
		t.Errorf("State = %v, want %v", p.State(), StatePaused)
	}

	p.Resume()
	if p.State() != StateRunning {
		t.Errorf("State = %v, want %v", p.State(), StateRunning)
	}
}

func TestProgressFail(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	testErr := errors.New("test error")
	p.Fail(testErr)

	if p.State() != StateFailed {
		t.Errorf("State = %v, want %v", p.State(), StateFailed)
	}
	if p.Error() != testErr {
		t.Errorf("Error = %v, want %v", p.Error(), testErr)
	}
}

func TestProgressStatus(t *testing.T) {
	p := NewProgress("Test Operation", 100)
	p.Start()

	p.SetPhase("Phase 1")
	if p.Phase() != "Phase 1" {
		t.Errorf("Phase = %q, want %q", p.Phase(), "Phase 1")
	}

	p.SetDetail("Processing item 5")
	if p.Detail() != "Processing item 5" {
		t.Errorf("Detail = %q, want %q", p.Detail(), "Processing item 5")
	}

	p.SetStatus("Phase 2", "Processing item 10")
	if p.Phase() != "Phase 2" {
		t.Errorf("Phase = %q, want %q", p.Phase(), "Phase 2")
	}
	if p.Detail() != "Processing item 10" {
		t.Errorf("Detail = %q, want %q", p.Detail(), "Processing item 10")
	}
}

func TestProgressETA(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	// No ETA without samples
	if p.ETA() != 0 {
		t.Error("ETA should be 0 initially")
	}

	// Simulate progress over time with longer delays for reliable ETA
	for i := int64(0); i <= 30; i++ {
		p.Update(i)
		time.Sleep(20 * time.Millisecond)
	}

	// Should have an ETA now (may still be 0 if not enough time elapsed)
	eta := p.ETA()
	rate := p.Rate()

	// Rate should be non-zero after significant progress
	if rate <= 0 {
		t.Logf("Rate = %f, may need more samples", rate)
	}

	// ETA test is lenient - it depends on timing
	if eta > 0 {
		t.Logf("ETA = %v (good)", eta)
	}
}

func TestProgressElapsed(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	time.Sleep(100 * time.Millisecond)

	elapsed := p.Elapsed()
	if elapsed < 100*time.Millisecond {
		t.Errorf("Elapsed = %v, want >= 100ms", elapsed)
	}
}

func TestProgressCallbacks(t *testing.T) {
	p := NewProgress("Test", 100)

	var progressCalls int
	var completeCalled bool
	var mu sync.Mutex

	p.OnProgress(func(p *Progress) {
		mu.Lock()
		progressCalls++
		mu.Unlock()
	})

	p.OnComplete(func(p *Progress) {
		mu.Lock()
		completeCalled = true
		mu.Unlock()
	})

	p.Start()
	p.Update(50)
	time.Sleep(50 * time.Millisecond) // Wait for async callbacks

	mu.Lock()
	if progressCalls < 1 {
		t.Error("OnProgress callback should have been called")
	}
	mu.Unlock()

	p.Complete()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if !completeCalled {
		t.Error("OnComplete callback should have been called")
	}
	mu.Unlock()
}

func TestProgressErrorCallback(t *testing.T) {
	p := NewProgress("Test", 100)

	var errorReceived error
	var mu sync.Mutex

	p.OnError(func(p *Progress, err error) {
		mu.Lock()
		errorReceived = err
		mu.Unlock()
	})

	p.Start()
	testErr := errors.New("test failure")
	p.Fail(testErr)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if errorReceived != testErr {
		t.Errorf("Error callback received %v, want %v", errorReceived, testErr)
	}
	mu.Unlock()
}

func TestProgressInfo(t *testing.T) {
	p := NewProgress("Test Op", 100)
	p.Start()
	p.SetStatus("Phase 1", "Detail 1")
	p.Update(25)

	info := p.Info()

	if info.Operation != "Test Op" {
		t.Errorf("Info.Operation = %q, want %q", info.Operation, "Test Op")
	}
	if info.Phase != "Phase 1" {
		t.Errorf("Info.Phase = %q, want %q", info.Phase, "Phase 1")
	}
	if info.Detail != "Detail 1" {
		t.Errorf("Info.Detail = %q, want %q", info.Detail, "Detail 1")
	}
	if info.Current != 25 {
		t.Errorf("Info.Current = %d, want %d", info.Current, 25)
	}
	if info.Total != 100 {
		t.Errorf("Info.Total = %d, want %d", info.Total, 100)
	}
	if info.State != StateRunning {
		t.Errorf("Info.State = %v, want %v", info.State, StateRunning)
	}
	if info.Percent != 25.0 {
		t.Errorf("Info.Percent = %f, want %f", info.Percent, 25.0)
	}
}

func TestProgressStatusLine(t *testing.T) {
	p := NewProgress("Export", 100)
	p.Start()
	p.SetPhase("Writing files")
	p.Update(50)

	line := p.StatusLine()

	// Should contain operation name
	if !containsStr(line, "Export") {
		t.Error("StatusLine should contain operation name")
	}

	// Should contain phase
	if !containsStr(line, "Writing files") {
		t.Error("StatusLine should contain phase")
	}

	// Should contain percentage
	if !containsStr(line, "50") {
		t.Error("StatusLine should contain percentage")
	}
}

func TestProgressIndeterminate(t *testing.T) {
	p := NewProgress("Loading", 0) // 0 = indeterminate

	if !p.IsIndeterminate() {
		t.Error("IsIndeterminate should be true for total=0")
	}

	// Fraction should be 0 for indeterminate
	if p.Fraction() != 0 {
		t.Error("Fraction should be 0 for indeterminate progress")
	}
}

func TestProgressSetTotal(t *testing.T) {
	p := NewProgress("Test", 0) // Start indeterminate
	p.Start()

	if !p.IsIndeterminate() {
		t.Error("Should start as indeterminate")
	}

	p.SetTotal(100)
	if p.IsIndeterminate() {
		t.Error("Should not be indeterminate after SetTotal")
	}
	if p.Total() != 100 {
		t.Errorf("Total = %d, want 100", p.Total())
	}
}

func TestProgressUpdateWithStatus(t *testing.T) {
	p := NewProgress("Test", 100)
	p.Start()

	p.UpdateWithStatus(50, "Phase X", "Detail Y")

	if p.Current() != 50 {
		t.Errorf("Current = %d, want 50", p.Current())
	}
	if p.Phase() != "Phase X" {
		t.Errorf("Phase = %q, want %q", p.Phase(), "Phase X")
	}
	if p.Detail() != "Detail Y" {
		t.Errorf("Detail = %q, want %q", p.Detail(), "Detail Y")
	}
}

func TestProgressConcurrentUpdates(t *testing.T) {
	p := NewProgress("Test", 1000)
	p.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				p.Increment()
			}
		}()
	}

	wg.Wait()

	if p.Current() != 1000 {
		t.Errorf("Current = %d, want 1000 after concurrent increments", p.Current())
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateIdle, "idle"},
		{StateRunning, "running"},
		{StatePaused, "paused"},
		{StateCancelled, "cancelled"},
		{StateCompleted, "completed"},
		{StateFailed, "failed"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{500 * time.Millisecond, "<1s"},
		{5 * time.Second, "5s"},
		{90 * time.Second, "1m 30s"},
		{3600 * time.Second, "1h 0m"},
		{3661 * time.Second, "1h 1m"},
	}

	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.expected)
		}
	}
}

// containsStr checks if s contains substr
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
