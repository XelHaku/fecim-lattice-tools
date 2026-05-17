//go:build legacy_fyne

package widgets

import (
	"testing"

	"fecim-lattice-tools/shared/theme"
)

// TestEmbeddedAppBase_Lifecycle verifies Start/Stop state transitions
func TestEmbeddedAppBase_Lifecycle(t *testing.T) {
	base := &EmbeddedAppBase{}

	// Initial state
	if base.IsRunning() {
		t.Error("Should not be running initially")
	}

	// Start transition
	base.Start()
	if !base.IsRunning() {
		t.Error("Should be running after Start()")
	}

	// Stop transition
	base.Stop()
	if base.IsRunning() {
		t.Error("Should not be running after Stop()")
	}

	// Multiple Start/Stop cycles
	for i := 0; i < 5; i++ {
		base.Start()
		if !base.IsRunning() {
			t.Errorf("Cycle %d: Should be running after Start()", i)
		}

		base.Stop()
		if base.IsRunning() {
			t.Errorf("Cycle %d: Should not be running after Stop()", i)
		}
	}
}

// TestEmbeddedAppBase_LifecycleWithDemo verifies demo controller integration
func TestEmbeddedAppBase_LifecycleWithDemo(t *testing.T) {
	steps := []DemoStep{
		{Name: "Test", Duration: 1000000, Action: func() {}}, // Very long duration
	}
	demo := NewDemoController(steps)

	base := &EmbeddedAppBase{}
	base.SetDemoController(demo)

	// Demo should not be running initially
	if demo.IsRunning() {
		t.Error("Demo should not be running initially")
	}

	// Start should start the demo
	base.Start()
	if !demo.IsRunning() {
		t.Error("Demo should be running after Start()")
	}

	// Stop should stop the demo
	base.Stop()
	if demo.IsRunning() {
		t.Error("Demo should not be running after Stop()")
	}
}

// TestEmbeddedAppBase_InitAndGetters verifies initialization and getters
func TestEmbeddedAppBase_InitAndGetters(t *testing.T) {
	base := &EmbeddedAppBase{}

	// Before Init
	if base.GetFyneApp() != nil {
		t.Error("FyneApp should be nil before Init")
	}
	if base.GetWindow() != nil {
		t.Error("Window should be nil before Init")
	}
	if base.GetContent() != nil {
		t.Error("Content should be nil before Init")
	}

	// After Init with nil (simulating no-op)
	base.Init(nil, nil)

	// Should remain nil
	if base.GetFyneApp() != nil {
		t.Error("FyneApp should remain nil after Init(nil, nil)")
	}
	if base.GetWindow() != nil {
		t.Error("Window should remain nil after Init(nil, nil)")
	}
}

// TestEmbeddedAppBase_StatusBarIntegration verifies status bar updates
func TestEmbeddedAppBase_StatusBarIntegration(t *testing.T) {
	base := &EmbeddedAppBase{}
	status := NewStatusBar("Status: ")
	base.SetStatusBar(status)

	// Verify status bar is set
	if base.GetStatusBar() != status {
		t.Error("GetStatusBar should return the set status bar")
	}

	// Update status
	testMsg := "Running test"
	base.UpdateStatus(testMsg)

	if status.GetText() != testMsg {
		t.Errorf("Status text: got '%s', want '%s'", status.GetText(), testMsg)
	}

	// Multiple updates
	for i := 0; i < 10; i++ {
		msg := "Update"
		base.UpdateStatus(msg)
		if status.GetText() != msg {
			t.Errorf("Update %d: status text: got '%s', want '%s'", i, status.GetText(), msg)
		}
	}
}

// TestEmbeddedAppBase_SafeOperations verifies safe operations with nil components
func TestEmbeddedAppBase_SafeOperations(t *testing.T) {
	base := &EmbeddedAppBase{}

	// These should not panic
	base.UpdateStatus("test")                 // No status bar set
	base.RefreshContent()                     // No content set
	base.ShowNotification("title", "content") // No app set

	// Verify state is still valid
	if base.IsRunning() {
		t.Error("Should not be running after safe operations")
	}
}

// TestThemeConsistency verifies theme colors are valid
func TestThemeConsistency(t *testing.T) {
	// Verify all primary colors are non-nil
	colors := []struct {
		name  string
		color interface{}
	}{
		{"ColorPrimary", theme.ColorPrimary},
		{"ColorSecondary", theme.ColorSecondary},
		{"ColorAccent", theme.ColorAccent},
		{"ColorWarning", theme.ColorWarning},
		{"ColorSuccess", theme.ColorSuccess},
		{"ColorError", theme.ColorError},
		{"ColorBackground", theme.ColorBackground},
		{"ColorSurface", theme.ColorSurface},
		{"ColorText", theme.ColorText},
		{"ColorTextDim", theme.ColorTextDim},
	}

	for _, c := range colors {
		if c.color == nil {
			t.Errorf("%s is nil", c.name)
		}
	}
}

// TestThemeColorValues verifies specific color values
func TestThemeColorValues(t *testing.T) {
	// ColorPrimary should be cyan #00D4FF
	if r, g, b, _ := theme.ColorPrimary.RGBA(); r != 0 || g>>8 != 212 || b>>8 != 255 {
		t.Errorf("ColorPrimary has unexpected value")
	}

	// ColorBackground should be dark blue #003264
	if r, g, b, _ := theme.ColorBackground.RGBA(); r != 0 || g>>8 != 50 || b>>8 != 100 {
		t.Errorf("ColorBackground has unexpected value")
	}

	// ColorText should be off-white #F0F4F8
	if r, g, b, _ := theme.ColorText.RGBA(); r>>8 != 240 || g>>8 != 244 || b>>8 != 248 {
		t.Errorf("ColorText has unexpected value")
	}
}

// TestThemeWithAlpha verifies alpha channel utility
func TestThemeWithAlpha(t *testing.T) {
	original := theme.ColorPrimary
	alpha := uint8(128)

	modified := theme.WithAlpha(original, alpha)

	r, g, b, a := modified.RGBA()
	if uint8(a>>8) != alpha {
		t.Errorf("Alpha: got %d, want %d", uint8(a>>8), alpha)
	}

	// RGB should be preserved
	or, og, ob, _ := original.RGBA()
	if r != or || g != og || b != ob {
		t.Error("RGB values should be preserved when modifying alpha")
	}
}

// TestThemeGetContrastColor verifies contrast color selection
func TestThemeGetContrastColor(t *testing.T) {
	// Dark background should return light text
	contrastDark := theme.GetContrastColor(theme.ColorBackground)
	if contrastDark != theme.ColorText {
		t.Error("Dark background should return ColorText for contrast")
	}

	// Light background should return dark text
	contrastLight := theme.GetContrastColor(theme.ColorText)
	if contrastLight != theme.ColorBackground {
		t.Error("Light background should return ColorBackground for contrast")
	}
}

// TestEmbeddedAppBase_ContentManagement verifies content get/set
func TestEmbeddedAppBase_ContentManagement(t *testing.T) {
	base := &EmbeddedAppBase{}

	// Initially nil
	if base.GetContent() != nil {
		t.Error("Content should be nil initially")
	}

	// Set and get (we can't create real Fyne objects in tests, but nil is valid)
	base.SetContent(nil)
	if base.GetContent() != nil {
		t.Error("Content should remain nil after SetContent(nil)")
	}

	// Multiple set operations
	for i := 0; i < 5; i++ {
		base.SetContent(nil)
		if base.GetContent() != nil {
			t.Errorf("Iteration %d: Content should be nil", i)
		}
	}
}

// TestEmbeddedAppBase_DemoControllerManagement verifies demo controller get/set
func TestEmbeddedAppBase_DemoControllerManagement(t *testing.T) {
	base := &EmbeddedAppBase{}

	// Initially nil
	if base.GetDemoController() != nil {
		t.Error("DemoController should be nil initially")
	}

	// Set and get
	steps := []DemoStep{{Name: "Test", Duration: 1000, Action: func() {}}}
	demo := NewDemoController(steps)
	base.SetDemoController(demo)

	if base.GetDemoController() != demo {
		t.Error("GetDemoController should return the set controller")
	}

	// Replace with new controller
	demo2 := NewLoopingDemoController(steps)
	base.SetDemoController(demo2)

	if base.GetDemoController() != demo2 {
		t.Error("GetDemoController should return the new controller")
	}
}

// TestEmbeddedAppBaseBuilder_Complete verifies builder pattern
func TestEmbeddedAppBaseBuilder_Complete(t *testing.T) {
	steps := []DemoStep{{Name: "Step", Duration: 100, Action: func() {}}}

	base := NewEmbeddedAppBaseBuilder().
		WithStatusBar("Test: ").
		WithDemoController(steps).
		Build()

	// Verify all components are set
	if base.GetStatusBar() == nil {
		t.Error("Builder should have created status bar")
	}
	if base.GetDemoController() == nil {
		t.Error("Builder should have created demo controller")
	}

	// Verify status bar prefix
	status := base.GetStatusBar()
	testMsg := "testing"
	base.UpdateStatus(testMsg)
	if status.GetText() != testMsg {
		t.Errorf("Status text: got '%s', want '%s'", status.GetText(), testMsg)
	}

	// Verify demo controller works
	demo := base.GetDemoController()
	if demo.loop {
		t.Error("Builder with WithDemoController should not create looping demo")
	}
}

// TestEmbeddedAppBaseBuilder_Looping verifies looping demo creation
func TestEmbeddedAppBaseBuilder_Looping(t *testing.T) {
	steps := []DemoStep{{Name: "Loop", Duration: 50, Action: func() {}}}

	base := NewEmbeddedAppBaseBuilder().
		WithLoopingDemo(steps).
		Build()

	demo := base.GetDemoController()
	if demo == nil {
		t.Fatal("Builder should have created demo controller")
	}
	if !demo.loop {
		t.Error("Builder with WithLoopingDemo should create looping demo")
	}
}
