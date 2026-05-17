//go:build legacy_fyne

package tests

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	m3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
)

func TestM3_GUIIntegration_LoadsAndBindsControls(t *testing.T) {
	t.Setenv("FECIM_FYNE_TEST", "1")
	fy := test.NewApp()
	w := fy.NewWindow("m3-test")
	defer w.Close()

	app := m3gui.NewDualModeApp()
	content := app.BuildContent(fy, w)
	if content == nil {
		t.Fatal("BuildContent returned nil")
	}
	// NOTE: Do not call w.SetContent(content) here.
	// The full adaptive split layout has an expensive MinSize traversal in headless test windows
	// and can hit test timeouts. These integration tests focus on widget wiring/binding,
	// which does not require a rendered layout.

	h := app.TestHooks()
	if h.LevelsSelect == nil || h.NoiseSlider == nil {
		t.Fatalf("expected controls to be initialized: levelsSelect=%v noiseSlider=%v", h.LevelsSelect, h.NoiseSlider)
	}
	if h.StatusLabel == nil || h.FPPredLabel == nil || h.CIMPredLabel == nil || h.AgreementLabel == nil {
		t.Fatalf("expected result/status widgets to be initialized: status=%v fp=%v cim=%v agree=%v",
			h.StatusLabel, h.FPPredLabel, h.CIMPredLabel, h.AgreementLabel)
	}

	cfg := app.NetworkConfig()
	if cfg == nil {
		t.Fatal("NetworkConfig returned nil")
	}

	// --- Parameter binding: noise slider updates network config.
	origNoise := cfg.NoiseLevel
	newNoise := origNoise + 0.02
	if newNoise > 0.20 {
		newNoise = 0.10
	}
	done := make(chan struct{})
	go func() {
		h.NoiseSlider.SetValue(newNoise)
		if h.NoiseSlider.OnChanged != nil {
			h.NoiseSlider.OnChanged(newNoise)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout applying noise slider change")
	}
	if cfg.NoiseLevel != newNoise {
		t.Fatalf("noise binding failed: cfg.NoiseLevel=%v want=%v", cfg.NoiseLevel, newNoise)
	}

	// --- Parameter binding: levels select updates network config.
	// We pick a value that is typically present in shipped weights (8, 16, 24, 30...).
	// If not present, we skip rather than failing due to environment-specific weight files.
	wantLevels := 8
	found := false
	for _, opt := range h.LevelsSelect.Options {
		if opt == "8" {
			found = true
			break
		}
	}
	if !found {
		t.Skipf("levels option 8 not available (options=%v)", h.LevelsSelect.Options)
	}

	done2 := make(chan struct{})
	go func() {
		h.LevelsSelect.SetSelected("8")
		close(done2)
	}()
	select {
	case <-done2:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout applying levels selection")
	}
	if cfg.NumLevels != wantLevels {
		t.Fatalf("levels binding failed: cfg.NumLevels=%d want=%d", cfg.NumLevels, wantLevels)
	}
}

func TestM3_GUIIntegration_ResultDisplayUpdatesAfterInference(t *testing.T) {
	t.Setenv("FECIM_FYNE_TEST", "1")
	fy := test.NewApp()
	w := fy.NewWindow("m3-test")
	defer w.Close()

	app := m3gui.NewDualModeApp()
	content := app.BuildContent(fy, w)
	if content == nil {
		t.Fatal("BuildContent returned nil")
	}
	// Intentionally do not call w.SetContent(content); see note in first test.

	h := app.TestHooks()
	beforeFP := h.FPPredLabel.Text
	beforeCIM := h.CIMPredLabel.Text

	// Deterministic input: all zeros is a valid 28x28 image.
	pixels := make([]float64, 28*28)
	done := make(chan struct{})
	go func() {
		app.RunInferenceForTest(pixels)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout running inference")
	}

	// runInference uses fyne.Do(...) to update labels; give the UI queue a moment.
	for i := 0; i < 100; i++ {
		if h.FPPredLabel.Text != beforeFP || h.CIMPredLabel.Text != beforeCIM {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	// If still unchanged, fail with captured text for audit.
	t.Fatalf("expected prediction labels to change after inference; beforeFP=%q afterFP=%q beforeCIM=%q afterCIM=%q",
		beforeFP, h.FPPredLabel.Text, beforeCIM, h.CIMPredLabel.Text)
}
