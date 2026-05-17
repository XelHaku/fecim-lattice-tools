//go:build legacy_fyne

package widgets

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDemoController(t *testing.T) {
	steps := []DemoStep{
		{Name: "Step 1", Duration: 10 * time.Millisecond, Action: func() {}},
	}
	dc := NewDemoController(steps)

	if dc == nil {
		t.Fatal("NewDemoController should not return nil")
	}
	if dc.loop {
		t.Error("Default demo should not loop")
	}
}

func TestNewLoopingDemoController(t *testing.T) {
	steps := []DemoStep{
		{Name: "Step 1", Duration: 10 * time.Millisecond, Action: func() {}},
	}
	dc := NewLoopingDemoController(steps)

	if dc == nil {
		t.Fatal("NewLoopingDemoController should not return nil")
	}
	if !dc.loop {
		t.Error("Looping demo should have loop=true")
	}
}

func TestDemoControllerStartStop(t *testing.T) {
	var executed int32

	steps := []DemoStep{
		{Name: "Step 1", Duration: 50 * time.Millisecond, Action: func() {
			atomic.AddInt32(&executed, 1)
		}},
	}
	dc := NewDemoController(steps)

	dc.Start()
	if !dc.IsRunning() {
		t.Error("Demo should be running after Start()")
	}

	// Wait for step to execute
	time.Sleep(20 * time.Millisecond)

	dc.Stop()
	if dc.IsRunning() {
		t.Error("Demo should not be running after Stop()")
	}

	if atomic.LoadInt32(&executed) == 0 {
		t.Error("Step action should have been executed")
	}
}

func TestDemoControllerRunsAllSteps(t *testing.T) {
	var step1, step2, step3 int32

	steps := []DemoStep{
		{Name: "Step 1", Duration: 10 * time.Millisecond, Action: func() {
			atomic.AddInt32(&step1, 1)
		}},
		{Name: "Step 2", Duration: 10 * time.Millisecond, Action: func() {
			atomic.AddInt32(&step2, 1)
		}},
		{Name: "Step 3", Duration: 10 * time.Millisecond, Action: func() {
			atomic.AddInt32(&step3, 1)
		}},
	}
	dc := NewDemoController(steps)

	dc.Start()

	// Wait for all steps to complete
	time.Sleep(50 * time.Millisecond)

	// Demo should auto-stop after completing (no loop)
	if dc.IsRunning() {
		t.Error("Non-looping demo should stop after all steps")
	}

	if atomic.LoadInt32(&step1) != 1 {
		t.Errorf("Step 1 should have run once, ran %d times", step1)
	}
	if atomic.LoadInt32(&step2) != 1 {
		t.Errorf("Step 2 should have run once, ran %d times", step2)
	}
	if atomic.LoadInt32(&step3) != 1 {
		t.Errorf("Step 3 should have run once, ran %d times", step3)
	}
}

func TestDemoControllerLooping(t *testing.T) {
	var count int32

	steps := []DemoStep{
		{Name: "Step", Duration: 10 * time.Millisecond, Action: func() {
			atomic.AddInt32(&count, 1)
		}},
	}
	dc := NewLoopingDemoController(steps)

	dc.Start()

	// Let it loop a few times
	time.Sleep(45 * time.Millisecond)

	dc.Stop()

	// Should have run multiple times
	runs := atomic.LoadInt32(&count)
	if runs < 2 {
		t.Errorf("Looping demo should run multiple times, ran %d times", runs)
	}
}

func TestDemoControllerCallbacks(t *testing.T) {
	var started, stopped int32
	var stepsCompleted int32

	steps := []DemoStep{
		{Name: "Step 1", Duration: 10 * time.Millisecond, Action: func() {}},
		{Name: "Step 2", Duration: 10 * time.Millisecond, Action: func() {}},
	}
	dc := NewDemoController(steps)

	dc.SetOnStart(func() {
		atomic.AddInt32(&started, 1)
	})
	dc.SetOnStop(func() {
		atomic.AddInt32(&stopped, 1)
	})
	dc.SetOnStepDone(func(idx int, step DemoStep) {
		atomic.AddInt32(&stepsCompleted, 1)
	})

	dc.Start()
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&started) != 1 {
		t.Error("OnStart should have been called")
	}
	if atomic.LoadInt32(&stopped) != 1 {
		t.Error("OnStop should have been called")
	}
	if atomic.LoadInt32(&stepsCompleted) != 2 {
		t.Errorf("OnStepDone should have been called twice, called %d times", stepsCompleted)
	}
}

func TestDemoControllerDoubleStart(t *testing.T) {
	var count int32

	steps := []DemoStep{
		{Name: "Step", Duration: 100 * time.Millisecond, Action: func() {
			atomic.AddInt32(&count, 1)
		}},
	}
	dc := NewDemoController(steps)

	dc.Start()
	dc.Start() // Should be ignored
	dc.Start() // Should be ignored

	time.Sleep(20 * time.Millisecond)
	dc.Stop()

	// Should have only started once
	if atomic.LoadInt32(&count) != 1 {
		t.Error("Double start should be ignored")
	}
}

func TestDemoControllerWaitOrStop(t *testing.T) {
	dc := NewDemoController(nil)

	// Not running - should return false
	if dc.WaitOrStop(10 * time.Millisecond) {
		t.Error("WaitOrStop should return false when not running")
	}

	// Create a long-running demo to test WaitOrStop
	steps := []DemoStep{
		{Name: "Long step", Duration: 500 * time.Millisecond, Action: func() {}},
	}
	dc2 := NewDemoController(steps)
	dc2.Start()

	// Wait should complete before the step duration
	start := time.Now()
	result := dc2.WaitOrStop(20 * time.Millisecond)
	elapsed := time.Since(start)

	dc2.Stop()

	if !result {
		t.Error("WaitOrStop should return true when wait completes")
	}
	if elapsed < 15*time.Millisecond {
		t.Errorf("Wait should take at least 15ms, took %v", elapsed)
	}
}

func TestTickerDemoController(t *testing.T) {
	var count int32

	tc := NewTickerDemoController(15*time.Millisecond, 0, func(step int) {
		atomic.AddInt32(&count, 1)
	})

	tc.Start()
	if !tc.IsRunning() {
		t.Error("Ticker should be running after Start()")
	}

	// Let it tick a few times
	time.Sleep(50 * time.Millisecond)

	tc.Stop()
	if tc.IsRunning() {
		t.Error("Ticker should not be running after Stop()")
	}

	runs := atomic.LoadInt32(&count)
	if runs < 2 {
		t.Errorf("Ticker should have run multiple times, ran %d times", runs)
	}
}

func TestTickerDemoControllerMaxSteps(t *testing.T) {
	var count int32

	tc := NewTickerDemoController(10*time.Millisecond, 3, func(step int) {
		atomic.AddInt32(&count, 1)
	})

	tc.Start()

	// Wait for max steps to be reached
	time.Sleep(60 * time.Millisecond)

	if tc.IsRunning() {
		t.Error("Ticker should auto-stop after max steps")
	}

	runs := atomic.LoadInt32(&count)
	if runs != 3 {
		t.Errorf("Ticker should have run exactly 3 times, ran %d times", runs)
	}
}

func TestTickerDemoControllerGetStep(t *testing.T) {
	tc := NewTickerDemoController(10*time.Millisecond, 0, func(step int) {})

	if tc.GetStep() != 0 {
		t.Error("Initial step should be 0")
	}

	tc.Start()
	deadline := time.Now().Add(250 * time.Millisecond)
	for tc.GetStep() < 1 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	tc.Stop()

	step := tc.GetStep()
	if step < 1 {
		t.Errorf("Step should have incremented, got %d", step)
	}
}

func TestDemoControllerPauseResume(t *testing.T) {
	dc := NewLoopingDemoController(nil)

	// Test pause/resume state transitions
	dc.Start()
	time.Sleep(10 * time.Millisecond)

	if dc.IsPaused() {
		t.Error("Demo should not be paused initially")
	}

	dc.Pause()
	if !dc.IsPaused() {
		t.Error("Demo should be paused after Pause()")
	}
	if !dc.IsRunning() {
		t.Error("Demo should still be running while paused")
	}

	dc.Resume()
	if dc.IsPaused() {
		t.Error("Demo should not be paused after Resume()")
	}

	dc.Stop()
	if dc.IsRunning() {
		t.Error("Demo should not be running after Stop()")
	}
	if dc.IsPaused() {
		t.Error("Demo should not be paused after Stop()")
	}
}
