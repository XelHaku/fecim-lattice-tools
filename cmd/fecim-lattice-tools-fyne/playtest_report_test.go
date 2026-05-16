//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
	"fecim-lattice-tools/shared/peripherals"
)

const playtestReportPath = "testdata/playtest-report.txt"

// reportModuleEntry describes a module for the consolidated report.
type reportModuleEntry struct {
	name           string
	create         func() (moduleLifecycle, error)
	skipScreenshot bool
	skipTree       bool
	mnistGate      bool
}

func reportAllModules() []reportModuleEntry {
	return []reportModuleEntry{
		{"hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil }, false, false, false},
		{"crossbar", func() (moduleLifecycle, error) { return demo2gui.NewEmbeddedCrossbarApp() }, true, true, false},
		{"mnist", func() (moduleLifecycle, error) { return demo3gui.NewEmbeddedDualModeApp(), nil }, true, true, true},
		{"circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }, false, false, false},
		{"comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil }, false, false, false},
		{"eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil }, false, false, false},
		{"docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil }, false, false, false},
	}
}

// TestPlaytestReport_Full is the master orchestrator that produces a consolidated
// playtest report covering all 7 modules.
func TestPlaytestReport_Full(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	if err := os.MkdirAll(playtestScreenshotDir, 0755); err != nil {
		t.Logf("Warning: could not create screenshot dir: %v", err)
	}
	if err := os.MkdirAll(playtestTreeDumpDir, 0755); err != nil {
		t.Logf("Warning: could not create tree dump dir: %v", err)
	}

	var report strings.Builder
	report.WriteString("FeCIM Lattice Tools -- Playtest Report\n")
	report.WriteString("======================================\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	mnistEnabled := false
	if v := strings.TrimSpace(os.Getenv("FECIM_MNIST_PLAYTEST")); v != "" && v != "0" {
		mnistEnabled = true
	}

	totalCritical := 0
	totalWarning := 0
	totalInfo := 0

	modules := reportAllModules()
	for _, m := range modules {
		m := m
		report.WriteString(fmt.Sprintf("Module: %s\n", m.name))
		report.WriteString(strings.Repeat("-", 40) + "\n")

		// Gate checks
		if m.mnistGate && !mnistEnabled {
			report.WriteString("Status: SKIPPED (FECIM_MNIST_PLAYTEST not set)\n\n")
			t.Logf("Skipping %s: FECIM_MNIST_PLAYTEST not set", m.name)
			continue
		}
		if m.skipScreenshot && m.skipTree {
			report.WriteString("Status: SKIPPED (known test driver limitation)\n")
			if m.name == "crossbar" {
				report.WriteString("Reason: Bold+Monospace font crash in test.NewApp()\n")
			} else if m.name == "mnist" {
				report.WriteString("Reason: fyne.Do() deadlock in WeightComparisonWidget\n")
			}
			report.WriteString("\n")
			t.Logf("Skipping %s: known test driver limitation", m.name)
			continue
		}

		// Create module
		mod, err := m.create()
		if err != nil {
			report.WriteString(fmt.Sprintf("Status: ERROR creating module: %v\n\n", err))
			t.Logf("Error creating %s: %v", m.name, err)
			continue
		}

		var w fyne.Window
		var content fyne.CanvasObject
		fyne.DoAndWait(func() {
			w = fy.NewWindow("Report - " + m.name)
			content = mod.BuildContent(fy, w)
			w.SetContent(container.NewMax(content))
			w.Resize(fyne.NewSize(1200, 800))
			w.Show()
		})
		time.Sleep(200 * time.Millisecond)

		// Screenshot
		screenshotPath := ""
		if !m.skipScreenshot {
			var img image.Image
			fyne.DoAndWait(func() {
				img = captureWindow(w)
			})
			verifyImageNotEmpty(t, img, m.name)
			fname := filepath.Join(playtestScreenshotDir, "report_"+m.name+".png")
			if f, err := os.Create(fname); err == nil {
				if err := png.Encode(f, img); err == nil {
					screenshotPath = fname
				}
				f.Close()
			}
		}
		if screenshotPath != "" {
			report.WriteString(fmt.Sprintf("Screenshot: %s\n", screenshotPath))
		} else if m.skipScreenshot {
			report.WriteString("Screenshot: (skipped)\n")
		} else {
			report.WriteString("Screenshot: (save failed)\n")
		}

		// Widget tree
		if !m.skipTree {
			var nodes []widgetNode
			fyne.DoAndWait(func() {
				nodes = WalkWidgetTree(content)
			})

			totalCount := len(nodes)
			interactiveCount := 0
			for _, n := range nodes {
				if n.Interactive {
					interactiveCount++
				}
			}
			unlabeled := findUnlabeledTappables(nodes)

			report.WriteString(fmt.Sprintf("[INFO] Widget tree: total=%d interactive=%d unlabeled=%d\n",
				totalCount, interactiveCount, len(unlabeled)))
			totalInfo++

			if len(unlabeled) > 0 {
				report.WriteString("[CRITICAL] Unlabeled tappable widgets:\n")
				totalCritical += len(unlabeled)
				for _, n := range unlabeled {
					report.WriteString(fmt.Sprintf("  - %s at %.0f,%.0f (%.0fx%.0f)\n",
						n.TypeName, n.Position.X, n.Position.Y, n.Size.Width, n.Size.Height))
				}
			}

			// Save tree dump
			dump := DumpWidgetTree(nodes)
			dumpPath := filepath.Join(playtestTreeDumpDir, "tree-dump-"+m.name+".txt")
			if err := os.WriteFile(dumpPath, []byte(dump), 0644); err != nil {
				t.Logf("Warning: could not write tree dump for %s: %v", m.name, err)
			}

			// Accessibility checks (logged as warnings)
			smallTapTargets := 0
			for _, n := range nodes {
				if n.Interactive && (n.Size.Width < 44 || n.Size.Height < 44) {
					smallTapTargets++
				}
			}
			if smallTapTargets > 0 {
				report.WriteString(fmt.Sprintf("[WARNING] Tap targets < 44dp: %d (Fyne theme limitation)\n", smallTapTargets))
				totalWarning++
			}

			// Layout check at default size
			var entries []boundEntry
			fyne.DoAndWait(func() {
				entries = collectBounds(content)
			})
			overlaps := detectOverlaps(entries)
			if len(overlaps) > 0 {
				report.WriteString(fmt.Sprintf("[INFO] Layout overlaps at 1200x800: %d\n", len(overlaps)))
				totalInfo++
			} else {
				report.WriteString("[INFO] Layout overlaps at 1200x800: none\n")
				totalInfo++
			}
		} else {
			report.WriteString("Widget tree: (skipped)\n")
		}

		// Clean up window
		fyne.DoAndWait(func() {
			w.Close()
		})
		time.Sleep(100 * time.Millisecond)

		report.WriteString("\n")
	}

	report.WriteString(fmt.Sprintf("\nSummary: %d critical, %d warnings, %d info\n", totalCritical, totalWarning, totalInfo))

	// Signal Chain Fidelity section
	report.WriteString("\nSignal Chain Fidelity\n")
	report.WriteString(strings.Repeat("-", 40) + "\n")

	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0.5, Vmin: 0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: adc.Bits, Vmin: adc.VrefLow, Vmax: adc.VrefHigh},
	}
	imin, imax := sense.CurrentRange()
	lsb := sense.CurrentLSB()
	enob := adc.ENOB()
	effSNR := adc.EffectiveSNR()

	report.WriteString(fmt.Sprintf("DAC: %d-bit (%d levels), range %.2fV to %.2fV\n",
		dac.Bits, dac.Levels(), dac.VrefLow, dac.VrefHigh))
	report.WriteString(fmt.Sprintf("ADC ENOB: %.2f bits (nominal %d-bit, INL=%.1f, DNL=%.2f)\n",
		enob, adc.Bits, adc.INL, adc.DNL))
	report.WriteString(fmt.Sprintf("Effective SNR: %.1f dB\n", effSNR))
	report.WriteString(fmt.Sprintf("Sense chain current range: %.2f µA to %.2f µA\n",
		imin*1e6, imax*1e6))
	report.WriteString(fmt.Sprintf("Current LSB: %.2f nA\n", lsb*1e9))

	report.WriteString("======================================\n")
	report.WriteString("End of report\n")

	// Write report
	reportContent := report.String()
	if err := os.MkdirAll(filepath.Dir(playtestReportPath), 0755); err != nil {
		t.Logf("Warning: could not create report dir: %v", err)
	}
	if err := os.WriteFile(playtestReportPath, []byte(reportContent), 0644); err != nil {
		t.Fatalf("Could not write playtest report: %v", err)
	}
	t.Logf("Playtest report written to: %s", playtestReportPath)
	t.Logf("\n%s", reportContent)
}
