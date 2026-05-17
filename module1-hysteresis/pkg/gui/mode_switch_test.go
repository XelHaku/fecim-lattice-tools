//go:build legacy_fyne

package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/shared/logging"
)

func newModeSwitchTestApp(t *testing.T) *App {
	t.Helper()
	test.NewApp()
	if log == nil {
		log = logging.NewLogger("gui-test")
	}
	a := NewApp()
	a.createControlsPanel()
	return a
}

func seedHistory(a *App, n int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := 0; i < n; i++ {
		a.appendHistoryLocked(float64(i), float64(i)*0.1)
	}
}

func assertHistoryClearedAndTimeReset(t *testing.T, a *App) {
	t.Helper()
	a.mu.RLock()
	defer a.mu.RUnlock()
	if got := a.historyLengthLocked(); got != 0 {
		t.Fatalf("history should be cleared after waveform switch; got=%d", got)
	}
	if a.simTime != 0 {
		t.Fatalf("simTime should reset to 0 on waveform switch; got=%g", a.simTime)
	}
}

func TestWaveformSwitchClearsHistory_SineToTriangle(t *testing.T) {
	a := newModeSwitchTestApp(t)
	a.waveformSelect.SetSelected("Sine Wave")
	seedHistory(a, 100)

	a.mu.Lock()
	a.simTime = 42
	a.mu.Unlock()

	a.waveformSelect.SetSelected("Triangle Wave")
	assertHistoryClearedAndTimeReset(t, a)
}

func TestWaveformSwitchClearsHistory_ISPPToManual(t *testing.T) {
	a := newModeSwitchTestApp(t)
	a.waveformSelect.SetSelected("ISPP (Write/Read)")
	seedHistory(a, 100)

	a.mu.Lock()
	a.simTime = 17
	a.mu.Unlock()

	a.waveformSelect.SetSelected("Manual")
	assertHistoryClearedAndTimeReset(t, a)
}

func TestWaveformSwitchClearsHistory_ManualToSine(t *testing.T) {
	a := newModeSwitchTestApp(t)
	a.waveformSelect.SetSelected("Manual")
	seedHistory(a, 100)

	a.mu.Lock()
	a.simTime = 9
	a.mu.Unlock()

	a.waveformSelect.SetSelected("Sine Wave")
	assertHistoryClearedAndTimeReset(t, a)
}
