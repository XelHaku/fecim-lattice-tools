package main

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func BenchmarkUILatency_TabSwitchAndResize(b *testing.B) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("ui-latency-bench")
	defer w.Close()

	tabs := container.NewAppTabs(
		container.NewTabItem("Launcher", CreateLauncherContent(nil)),
		container.NewTabItem("Simple1", container.NewVBox(widget.NewLabel("A"), widget.NewButton("Run", nil))),
		container.NewTabItem("Simple2", container.NewVBox(widget.NewLabel("B"), widget.NewSlider(0, 1))),
	)
	fyne.DoAndWait(func() {
		w.SetContent(tabs)
		w.Resize(fyne.NewSize(1200, 800))
		w.Show()
	})

	b.ResetTimer()
	tabDurations := make([]time.Duration, 0, b.N)
	resizeDurations := make([]time.Duration, 0, b.N)

	sizes := []fyne.Size{fyne.NewSize(1024, 768), fyne.NewSize(1200, 800), fyne.NewSize(1366, 768)}
	for i := 0; i < b.N; i++ {
		startTab := time.Now()
		fyne.DoAndWait(func() {
			tabs.SelectIndex(i % len(tabs.Items))
		})
		tabDurations = append(tabDurations, time.Since(startTab))

		startResize := time.Now()
		sz := sizes[i%len(sizes)]
		fyne.DoAndWait(func() {
			w.Resize(sz)
		})
		resizeDurations = append(resizeDurations, time.Since(startResize))
	}

	reportLatencyBudget(b, "tab_switch", tabDurations)
	reportLatencyBudget(b, "resize", resizeDurations)
}

func reportLatencyBudget(b *testing.B, name string, samples []time.Duration) {
	if len(samples) == 0 {
		b.Fatalf("no latency samples for %s", name)
	}
	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	p50 := percentile(sorted, 50)
	p95 := percentile(sorted, 95)
	p99 := percentile(sorted, 99)

	b.Logf("%s latency p50=%s p95=%s p99=%s", name, p50, p95, p99)
	b.ReportMetric(float64(p50.Microseconds())/1000.0, fmt.Sprintf("%s_p50_ms", name))
	b.ReportMetric(float64(p95.Microseconds())/1000.0, fmt.Sprintf("%s_p95_ms", name))
	b.ReportMetric(float64(p99.Microseconds())/1000.0, fmt.Sprintf("%s_p99_ms", name))

	if p99 > 500*time.Millisecond {
		b.Fatalf("%s p99 latency regression: %s > 500ms", name, p99)
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 1 {
		return sorted[0]
	}
	idx := int(float64(len(sorted)-1) * (float64(p) / 100.0))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
