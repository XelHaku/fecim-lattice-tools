package widgets

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestEmbeddedAppBaseInit(t *testing.T) {
	base := &EmbeddedAppBase{}

	// Before init, should be nil
	if base.GetFyneApp() != nil {
		t.Error("FyneApp should be nil before Init()")
	}
	if base.GetWindow() != nil {
		t.Error("Window should be nil before Init()")
	}

	// Init is typically called with real Fyne objects, but we test nil handling
	base.Init(nil, nil)

	if base.GetFyneApp() != nil {
		t.Error("FyneApp should remain nil if passed nil")
	}
}

func TestEmbeddedAppBaseContent(t *testing.T) {
	base := &EmbeddedAppBase{}

	if base.GetContent() != nil {
		t.Error("Content should be nil initially")
	}

	// SetContent with nil
	base.SetContent(nil)
	if base.GetContent() != nil {
		t.Error("Content should be nil after SetContent(nil)")
	}
}

func TestEmbeddedAppBaseStatusBar(t *testing.T) {
	base := &EmbeddedAppBase{}

	// UpdateStatus should not panic without status bar
	base.UpdateStatus("Test")

	// Set status bar
	status := NewStatusBar("Status: ")
	base.SetStatusBar(status)

	if base.GetStatusBar() != status {
		t.Error("GetStatusBar should return the set status bar")
	}

	// UpdateStatus should work now
	base.UpdateStatus("Running")
	if status.GetText() != "Running" {
		t.Errorf("Expected 'Running', got '%s'", status.GetText())
	}
}

func TestEmbeddedAppBaseStartStop(t *testing.T) {
	base := &EmbeddedAppBase{}

	if base.IsRunning() {
		t.Error("Should not be running initially")
	}

	base.Start()
	if !base.IsRunning() {
		t.Error("Should be running after Start()")
	}

	base.Stop()
	if base.IsRunning() {
		t.Error("Should not be running after Stop()")
	}
}

func TestEmbeddedAppBaseWithDemoController(t *testing.T) {
	var executed int32

	steps := []DemoStep{
		{Name: "Test", Duration: 10 * time.Millisecond, Action: func() {
			atomic.AddInt32(&executed, 1)
		}},
	}
	demo := NewDemoController(steps)

	base := &EmbeddedAppBase{}
	base.SetDemoController(demo)

	if base.GetDemoController() != demo {
		t.Error("GetDemoController should return the set controller")
	}

	// Start should start the demo
	base.Start()
	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt32(&executed) == 0 {
		t.Error("Demo step should have executed")
	}

	// Stop should stop the demo
	base.Stop()
	if demo.IsRunning() {
		t.Error("Demo should be stopped after Stop()")
	}
}

func TestEmbeddedAppBaseBuilder(t *testing.T) {
	steps := []DemoStep{
		{Name: "Step 1", Duration: 10 * time.Millisecond, Action: func() {}},
	}

	base := NewEmbeddedAppBaseBuilder().
		WithStatusBar("Status: ").
		WithDemoController(steps).
		Build()

	if base.GetStatusBar() == nil {
		t.Error("Builder should have created status bar")
	}
	if base.GetDemoController() == nil {
		t.Error("Builder should have created demo controller")
	}
}

func TestEmbeddedAppBaseBuilderLooping(t *testing.T) {
	steps := []DemoStep{
		{Name: "Step 1", Duration: 10 * time.Millisecond, Action: func() {}},
	}

	base := NewEmbeddedAppBaseBuilder().
		WithLoopingDemo(steps).
		Build()

	demo := base.GetDemoController()
	if demo == nil {
		t.Error("Builder should have created demo controller")
	}
	if !demo.loop {
		t.Error("Demo should be looping")
	}
}

func TestEmbeddedAppBaseRefreshContent(t *testing.T) {
	base := &EmbeddedAppBase{}

	// RefreshContent should not panic with nil content
	base.RefreshContent()
}

func TestEmbeddedAppBaseShowNotification(t *testing.T) {
	base := &EmbeddedAppBase{}

	// ShowNotification should not panic with nil app
	base.ShowNotification("Title", "Content")
}

func TestEmbeddedAppBaseConcurrency(t *testing.T) {
	base := &EmbeddedAppBase{}
	status := NewStatusBar("")
	base.SetStatusBar(status)

	// Concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 50; i++ {
			base.Start()
			base.UpdateStatus("Running")
			base.IsRunning()
			base.Stop()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			base.GetFyneApp()
			base.GetWindow()
			base.GetContent()
			base.GetStatusBar()
			base.IsRunning()
		}
		done <- true
	}()

	<-done
	<-done
}
