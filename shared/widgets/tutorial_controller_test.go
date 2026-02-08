// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"sync"
	"testing"
	"time"
)

func TestNewTutorialController(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Duration: 100 * time.Millisecond},
		{Title: "Step 2", Duration: 100 * time.Millisecond},
	}

	ctrl := NewTutorialController(steps)

	if ctrl == nil {
		t.Fatal("Expected non-nil controller")
	}

	if ctrl.TotalSteps() != 2 {
		t.Errorf("Expected 2 steps, got %d", ctrl.TotalSteps())
	}

	if ctrl.IsRunning() {
		t.Error("Controller should not be running initially")
	}
}

func TestTutorialControllerStartStop(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Duration: 50 * time.Millisecond},
		{Title: "Step 2", Duration: 50 * time.Millisecond},
	}

	ctrl := NewTutorialController(steps)

	ctrl.Start()
	time.Sleep(10 * time.Millisecond)

	if !ctrl.IsRunning() {
		t.Error("Controller should be running after Start()")
	}

	ctrl.Stop()
	time.Sleep(10 * time.Millisecond)

	if ctrl.IsRunning() {
		t.Error("Controller should not be running after Stop()")
	}
}

func TestTutorialControllerPauseResume(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Duration: 200 * time.Millisecond},
	}

	ctrl := NewTutorialController(steps)
	ctrl.Start()
	time.Sleep(10 * time.Millisecond)

	ctrl.Pause()
	if !ctrl.IsPaused() {
		t.Error("Controller should be paused after Pause()")
	}

	ctrl.Resume()
	if ctrl.IsPaused() {
		t.Error("Controller should not be paused after Resume()")
	}

	ctrl.Stop()
}

func TestTutorialControllerCallbacks(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Duration: 50 * time.Millisecond},
		{Title: "Step 2", Duration: 50 * time.Millisecond},
	}

	ctrl := NewTutorialController(steps)

	var mu sync.Mutex
	stepStarts := []int{}
	stepEnds := []int{}
	completed := false

	ctrl.SetOnStepStart(func(step int, total int, ts TutorialStep) {
		mu.Lock()
		stepStarts = append(stepStarts, step)
		mu.Unlock()
	})

	ctrl.SetOnStepEnd(func(step int, total int, ts TutorialStep) {
		mu.Lock()
		stepEnds = append(stepEnds, step)
		mu.Unlock()
	})

	ctrl.SetOnComplete(func(stats TutorialStats) {
		mu.Lock()
		completed = true
		mu.Unlock()
	})

	ctrl.Start()
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(stepStarts) != 2 {
		t.Errorf("Expected 2 step starts, got %d", len(stepStarts))
	}

	if len(stepEnds) != 2 {
		t.Errorf("Expected 2 step ends, got %d", len(stepEnds))
	}

	if !completed {
		t.Error("Expected tutorial to complete")
	}
}

func TestTutorialControllerProgress(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Duration: 50 * time.Millisecond},
		{Title: "Step 2", Duration: 50 * time.Millisecond},
		{Title: "Step 3", Duration: 50 * time.Millisecond},
		{Title: "Step 4", Duration: 50 * time.Millisecond},
	}

	ctrl := NewTutorialController(steps)

	if ctrl.GetProgress() != 0 {
		t.Error("Progress should be 0 before start")
	}

	ctrl.Start()
	time.Sleep(75 * time.Millisecond) // Should be past first step

	progress := ctrl.GetProgress()
	if progress < 0.2 || progress > 0.5 {
		t.Errorf("Expected progress around 0.25, got %f", progress)
	}

	ctrl.Stop()
}

func TestTutorialControllerLevelFilter(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Beginner", Duration: 50 * time.Millisecond, DifficultyLevel: LevelBeginner},
		{Title: "Advanced", Duration: 50 * time.Millisecond, DifficultyLevel: LevelAdvanced},
		{Title: "Beginner 2", Duration: 50 * time.Millisecond, DifficultyLevel: LevelBeginner},
	}

	ctrl := NewTutorialController(steps)
	ctrl.SetLevelFilter(LevelBeginner)

	stepsExecuted := 0
	ctrl.SetOnStepStart(func(step int, total int, ts TutorialStep) {
		stepsExecuted++
		if ts.DifficultyLevel > LevelBeginner {
			t.Errorf("Should not execute step with level %v", ts.DifficultyLevel)
		}
	})

	ctrl.Start()
	time.Sleep(200 * time.Millisecond)

	if stepsExecuted != 2 {
		t.Errorf("Expected 2 beginner steps to execute, got %d", stepsExecuted)
	}
}

func TestTutorialControllerFastMode(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Required", Duration: 50 * time.Millisecond, SkippedWhenFast: false},
		{Title: "Optional", Duration: 50 * time.Millisecond, SkippedWhenFast: true},
		{Title: "Required 2", Duration: 50 * time.Millisecond, SkippedWhenFast: false},
	}

	ctrl := NewTutorialController(steps)
	ctrl.SetFastMode(true)

	stepsExecuted := 0
	ctrl.SetOnStepStart(func(step int, total int, ts TutorialStep) {
		stepsExecuted++
		if ts.SkippedWhenFast {
			t.Error("Should not execute skippable step in fast mode")
		}
	})

	ctrl.Start()
	time.Sleep(200 * time.Millisecond)

	if stepsExecuted != 2 {
		t.Errorf("Expected 2 required steps to execute, got %d", stepsExecuted)
	}
}

func TestTutorialControllerStats(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Duration: 30 * time.Millisecond},
		{Title: "Step 2", Duration: 30 * time.Millisecond},
	}

	ctrl := NewTutorialController(steps)
	ctrl.Start()
	time.Sleep(100 * time.Millisecond)

	stats := ctrl.GetStats()

	if stats.TotalSteps != 2 {
		t.Errorf("Expected total steps 2, got %d", stats.TotalSteps)
	}

	if stats.CompletedSteps != 2 {
		t.Errorf("Expected 2 completed steps, got %d", stats.CompletedSteps)
	}

	if stats.TotalTime < 60*time.Millisecond {
		t.Errorf("Expected total time >= 60ms, got %v", stats.TotalTime)
	}
}

func TestAnimationController(t *testing.T) {
	frames := []AnimationFrame{
		{Title: "Frame 1", Duration: 50 * time.Millisecond},
		{Title: "Frame 2", Duration: 50 * time.Millisecond},
	}

	ctrl := NewAnimationController(frames)

	if ctrl.IsRunning() {
		t.Error("Animation should not be running initially")
	}

	var mu sync.Mutex
	framesShown := 0

	ctrl.SetOnFrame(func(frame int, af AnimationFrame) {
		mu.Lock()
		framesShown++
		mu.Unlock()
	})

	ctrl.Start()
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	count := framesShown
	mu.Unlock()

	if count != 2 {
		t.Errorf("Expected 2 frames shown, got %d", count)
	}
}

func TestAnimationControllerLoop(t *testing.T) {
	frames := []AnimationFrame{
		{Title: "Frame 1", Duration: 30 * time.Millisecond},
		{Title: "Frame 2", Duration: 30 * time.Millisecond},
	}

	ctrl := NewAnimationController(frames)
	ctrl.SetLoop(true)

	var mu sync.Mutex
	framesShown := 0

	ctrl.SetOnFrame(func(frame int, af AnimationFrame) {
		mu.Lock()
		framesShown++
		mu.Unlock()
	})

	ctrl.Start()
	time.Sleep(200 * time.Millisecond)
	ctrl.Stop()

	mu.Lock()
	count := framesShown
	mu.Unlock()

	// With looping and 200ms runtime, should show more than 2 frames
	if count < 3 {
		t.Errorf("Expected at least 3 frames with looping, got %d", count)
	}
}

func TestTutorialLevel(t *testing.T) {
	tests := []struct {
		level    TutorialLevel
		expected string
	}{
		{LevelBeginner, "Beginner"},
		{LevelIntermediate, "Intermediate"},
		{LevelAdvanced, "Advanced"},
		{LevelExpert, "Expert"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("Level %d: expected %q, got %q", tt.level, tt.expected, result)
		}
	}
}

func TestHighlightOverlay(t *testing.T) {
	overlay := NewHighlightOverlay()

	// Add highlight
	overlay.AddHighlight("test1", HighlightConfig{
		Label: "Test Highlight",
	})

	// Should have one highlight
	overlay.mu.RLock()
	count := len(overlay.highlights)
	overlay.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 highlight, got %d", count)
	}

	// Remove highlight
	overlay.RemoveHighlight("test1")

	overlay.mu.RLock()
	count = len(overlay.highlights)
	overlay.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 highlights after removal, got %d", count)
	}

	// Add multiple and clear
	overlay.AddHighlight("test1", HighlightConfig{})
	overlay.AddHighlight("test2", HighlightConfig{})
	overlay.ClearHighlights()

	overlay.mu.RLock()
	count = len(overlay.highlights)
	overlay.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 highlights after clear, got %d", count)
	}
}
