//go:build !ci
// +build !ci

package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
)

// ---------------------------------------------------------------------------
// Module render benchmarks: BuildContent + Resize
// ---------------------------------------------------------------------------

func BenchmarkPlaytest_ModuleRender_Hysteresis(b *testing.B) {
	b.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	benchmarkModuleRender(b, "hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil })
}

func BenchmarkPlaytest_ModuleRender_Circuits(b *testing.B) {
	benchmarkModuleRender(b, "circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil })
}

func BenchmarkPlaytest_ModuleRender_Comparison(b *testing.B) {
	benchmarkModuleRender(b, "comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil })
}

func BenchmarkPlaytest_ModuleRender_EDA(b *testing.B) {
	benchmarkModuleRender(b, "eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil })
}

func BenchmarkPlaytest_ModuleRender_Docs(b *testing.B) {
	benchmarkModuleRender(b, "docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil })
}

func benchmarkModuleRender(b *testing.B, name string, create func() (moduleLifecycle, error)) {
	b.Helper()

	fy := test.NewApp()
	b.Cleanup(func() { fy.Quit() })

	var w fyne.Window
	fyne.DoAndWait(func() {
		w = fy.NewWindow("Bench - " + name)
		w.Show()
	})
	b.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod, err := create()
		if err != nil {
			b.Fatalf("create failed: %v", err)
		}

		fyne.DoAndWait(func() {
			content := mod.BuildContent(fy, w)
			w.SetContent(container.NewMax(content))
			w.Resize(fyne.NewSize(1200, 800))
		})
	}
}

// ---------------------------------------------------------------------------
// Module capture benchmarks: Canvas.Capture() time
// ---------------------------------------------------------------------------

func BenchmarkPlaytest_ModuleCapture_Hysteresis(b *testing.B) {
	b.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	benchmarkModuleCapture(b, "hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil })
}

func BenchmarkPlaytest_ModuleCapture_Circuits(b *testing.B) {
	benchmarkModuleCapture(b, "circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil })
}

func benchmarkModuleCapture(b *testing.B, name string, create func() (moduleLifecycle, error)) {
	b.Helper()

	fy := test.NewApp()
	b.Cleanup(func() { fy.Quit() })

	mod, err := create()
	if err != nil {
		b.Fatalf("create failed: %v", err)
	}

	var w fyne.Window
	fyne.DoAndWait(func() {
		w = fy.NewWindow("BenchCapture - " + name)
		content := mod.BuildContent(fy, w)
		w.SetContent(container.NewMax(content))
		w.Resize(fyne.NewSize(1200, 800))
		w.Show()
	})
	b.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fyne.DoAndWait(func() {
			_ = w.Canvas().Capture()
		})
	}
}

// ---------------------------------------------------------------------------
// Startup time benchmarks: build + start + capture pipeline
// ---------------------------------------------------------------------------

func BenchmarkPlaytest_StartupTime_Hysteresis(b *testing.B) {
	b.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	benchmarkStartupTime(b, "hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil })
}

func BenchmarkPlaytest_StartupTime_Circuits(b *testing.B) {
	benchmarkStartupTime(b, "circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil })
}

func BenchmarkPlaytest_StartupTime_EDA(b *testing.B) {
	benchmarkStartupTime(b, "eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil })
}

func BenchmarkPlaytest_StartupTime_Docs(b *testing.B) {
	benchmarkStartupTime(b, "docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil })
}

func benchmarkStartupTime(b *testing.B, name string, create func() (moduleLifecycle, error)) {
	b.Helper()

	fy := test.NewApp()
	b.Cleanup(func() { fy.Quit() })

	var w fyne.Window
	fyne.DoAndWait(func() {
		w = fy.NewWindow("BenchStartup - " + name)
		w.Show()
	})
	b.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod, err := create()
		if err != nil {
			b.Fatalf("create failed: %v", err)
		}

		fyne.DoAndWait(func() {
			content := mod.BuildContent(fy, w)
			w.SetContent(container.NewMax(content))
			w.Resize(fyne.NewSize(1200, 800))
			mod.Start()
			_ = w.Canvas().Capture()
			mod.Stop()
		})
	}
}
