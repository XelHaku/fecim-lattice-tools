package widgets

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func resetEquationCachesForPerfTest() {
	equationSVGCacheMu.Lock()
	equationSVGCache = map[string]*equationSVGCacheEntry{}
	equationSVGCacheMu.Unlock()

	lkHotspotsOnce = sync.Once{}
	cachedLkSpots = nil
	cachedLkSize = fyne.Size{}
}

func measureEquationOpen(t *testing.T, budget time.Duration) time.Duration {
	t.Helper()
	start := time.Now()
	_ = NewPhysicsEquationsWidget(nil)
	dur := time.Since(start)
	if dur > budget {
		t.Fatalf("equation widget open exceeded budget: got=%s budget=%s", dur, budget)
	}
	return dur
}

func TestEquationWidgetPerf_OpenBudgets(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := test.NewWindow(widget.NewLabel("host"))
	defer w.Close()

	// Race mode adds ~2× overhead; use a generous budget to avoid false flakes.
	// Headless CI can be slower; keep a wider budget to reduce flakes.
	openBudget := 3000 * time.Millisecond
	if isRaceEnabled() {
		openBudget = 5000 * time.Millisecond
	}

	resetEquationCachesForPerfTest()
	cold := measureEquationOpen(t, openBudget)

	PrefetchEquationAssets()
	// Warm open should stay within the same budget even when cached path is used.
	warm := measureEquationOpen(t, openBudget)

	t.Logf("equation open timing: cold=%s warm=%s", cold, warm)
}

func BenchmarkEquationWidgetOpen_ColdWarm(b *testing.B) {
	app := test.NewApp()
	defer app.Quit()
	w := test.NewWindow(widget.NewLabel("host"))
	defer w.Close()

	b.Run("cold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resetEquationCachesForPerfTest()
			start := time.Now()
			_ = NewPhysicsEquationsWidget(nil)
			b.ReportMetric(float64(time.Since(start).Milliseconds()), "open_ms")
		}
	})

	b.Run("warm", func(b *testing.B) {
		PrefetchEquationAssets()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			_ = NewPhysicsEquationsWidget(nil)
			b.ReportMetric(float64(time.Since(start).Microseconds())/1000.0, "open_ms")
		}
	})
}
