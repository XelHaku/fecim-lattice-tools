//go:build !ci
// +build !ci

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
)

// playtestGoldenUpdateEnv controls golden file updates.
const playtestGoldenUpdateEnv = "FECIM_UPDATE_PLAYTEST_MARKUP"

// playtestGoldenDir is relative to testdata/ (Fyne's convention).
const playtestGoldenDir = "markup-golden"

func playtestGoldenShouldUpdate() bool {
	v := strings.TrimSpace(os.Getenv(playtestGoldenUpdateEnv))
	return v != "" && v != "0" && !strings.EqualFold(v, "false")
}

// playtestAssertGolden renders an object to markup and compares against a golden file.
// When FECIM_UPDATE_PLAYTEST_MARKUP=1, the golden file is created/updated automatically.
func playtestAssertGolden(t *testing.T, masterFilename string, o fyne.CanvasObject) {
	t.Helper()

	if playtestGoldenShouldUpdate() {
		masterPath := filepath.Join("testdata", masterFilename)
		if err := os.MkdirAll(filepath.Dir(masterPath), 0755); err != nil {
			t.Logf("Warning: could not create golden dir: %v", err)
		}
		if err := os.WriteFile(masterPath, []byte(test.RenderObjectToMarkup(o)), 0644); err != nil {
			t.Fatalf("Could not update golden master %s: %v", masterPath, err)
		}
		t.Logf("Golden master updated: %s", masterPath)
		return
	}

	// Normal assertion mode — fail if markup doesn't match master.
	test.AssertObjectRendersToMarkup(t, masterFilename, o)
}

// ---------------------------------------------------------------------------
// Golden test helpers
// ---------------------------------------------------------------------------

type goldenModuleEntry struct {
	name   string
	create func() (moduleLifecycle, error)
	skip   string // non-empty = skip reason
}

func goldenAllModules() []goldenModuleEntry {
	mnistGated := ""
	if v := strings.TrimSpace(os.Getenv("FECIM_MNIST_PLAYTEST")); v == "" || v == "0" {
		mnistGated = "FECIM_MNIST_PLAYTEST not set"
	}

	return []goldenModuleEntry{
		{"hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil }, ""},
		{"crossbar", func() (moduleLifecycle, error) { return demo2gui.NewEmbeddedCrossbarApp() }, ""},
		{"mnist", func() (moduleLifecycle, error) { return demo3gui.NewEmbeddedDualModeApp(), nil }, mnistGated},
		{"circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }, ""},
		{"comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil }, ""},
		{"eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil }, ""},
		{"docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil }, ""},
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPlaytestGolden_Hysteresis(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	testPlaytestGoldenModule(t, "hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil })
}

func TestPlaytestGolden_Crossbar(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestGoldenModule(t, "crossbar", func() (moduleLifecycle, error) { return demo2gui.NewEmbeddedCrossbarApp() })
}

func TestPlaytestGolden_MNIST(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	if v := strings.TrimSpace(os.Getenv("FECIM_MNIST_PLAYTEST")); v == "" || v == "0" {
		t.Skip("MNIST golden requires FECIM_MNIST_PLAYTEST=1")
	}
	t.Setenv("FECIM_FYNE_TEST", "1")
	testPlaytestGoldenModule(t, "mnist", func() (moduleLifecycle, error) { return demo3gui.NewEmbeddedDualModeApp(), nil })
}

func TestPlaytestGolden_Circuits(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestGoldenModule(t, "circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil })
}

func TestPlaytestGolden_Comparison(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestGoldenModule(t, "comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil })
}

func TestPlaytestGolden_EDA(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestGoldenModule(t, "eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil })
}

func TestPlaytestGolden_Docs(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestGoldenModule(t, "docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil })
}

func testPlaytestGoldenModule(t *testing.T, name string, create func() (moduleLifecycle, error)) {
	t.Helper()

	sizes := []struct {
		label string
		size  fyne.Size
	}{
		{"desktop", fyne.NewSize(1200, 800)},
		{"mobile", fyne.NewSize(375, 812)},
	}

	for _, sz := range sizes {
		sz := sz
		t.Run(sz.label, func(t *testing.T) {
			fy := test.NewApp()
			t.Cleanup(func() { fy.Quit() })

			mod, err := create()
			if err != nil {
				t.Fatalf("Failed to create %s: %v", name, err)
			}

			var (
				w       fyne.Window
				content fyne.CanvasObject
			)
			fyne.DoAndWait(func() {
				w = fy.NewWindow("Golden - " + name + " - " + sz.label)
				content = mod.BuildContent(fy, w)
				w.SetContent(container.NewMax(content))
				w.Resize(sz.size)
				w.Show()
			})
			t.Cleanup(func() {
				fyne.DoAndWait(func() { w.Close() })
			})

			goldenFile := playtestGoldenDir + "/" + name + "_" + sz.label + ".xml"
			playtestAssertGolden(t, goldenFile, content)
		})
	}
}
