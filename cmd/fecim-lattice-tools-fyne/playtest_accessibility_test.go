//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// ---------------------------------------------------------------------------
// Accessibility audit helpers
// ---------------------------------------------------------------------------

// checkTapTargets warns on interactive widgets smaller than 44x44 dp.
func checkTapTargets(t *testing.T, nodes []widgetNode, module string) {
	t.Helper()
	const minDp float32 = 44
	for _, n := range nodes {
		if !n.Interactive {
			continue
		}
		if n.Size.Width < minDp || n.Size.Height < minDp {
			label := n.Text
			if n.A11yLabel != "" {
				label = n.A11yLabel
			}
			t.Logf("[A11Y] %s: small tap target %s %q (%.0fx%.0f < 44x44) at %.0f,%.0f",
				module, n.TypeName, label, n.Size.Width, n.Size.Height, n.Position.X, n.Position.Y)
		}
	}
}

// checkLabelCompleteness warns on unlabeled interactive widgets.
func checkLabelCompleteness(t *testing.T, nodes []widgetNode, module string) {
	t.Helper()
	unlabeled := findUnlabeledTappables(nodes)
	for _, n := range unlabeled {
		t.Logf("[A11Y] %s: unlabeled interactive %s at %.0f,%.0f (%.0fx%.0f)",
			module, n.TypeName, n.Position.X, n.Position.Y, n.Size.Width, n.Size.Height)
	}
	if len(unlabeled) > 0 {
		t.Logf("[A11Y] %s: %d unlabeled interactive widget(s) found", module, len(unlabeled))
	}
}

// checkFocusTraversal steps focus 50 times on the canvas and logs the traversal order.
func checkFocusTraversal(t *testing.T, c fyne.Canvas, module string) {
	t.Helper()
	seen := 0
	var order []string
	for i := 0; i < 50; i++ {
		test.FocusNext(c)
		focused := c.Focused()
		if focused == nil {
			continue
		}
		seen++
		desc := fmt.Sprintf("%T", focused)
		if idx := strings.LastIndex(desc, "."); idx >= 0 {
			desc = desc[idx+1:]
		}
		order = append(order, desc)
	}
	t.Logf("[A11Y] %s: focus traversal hit %d focusable(s) in 50 steps: %s",
		module, seen, strings.Join(order, " -> "))
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPlaytestA11y_TextLabelCompleteness(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, content, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			var nodes []widgetNode
			fyne.DoAndWait(func() {
				nodes = WalkWidgetTree(content)
			})
			checkLabelCompleteness(t, nodes, m.name)
		})
	}
}

func TestPlaytestA11y_TapTargetSize(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, content, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			var nodes []widgetNode
			fyne.DoAndWait(func() {
				nodes = WalkWidgetTree(content)
			})
			checkTapTargets(t, nodes, m.name)
		})
	}
}

func TestPlaytestA11y_FocusTraversal(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, _, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			time.Sleep(100 * time.Millisecond)
			fyne.DoAndWait(func() {
				checkFocusTraversal(t, w.Canvas(), m.name)
			})
		})
	}
}

func TestPlaytestA11y_WidgetTreeDump(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	if err := os.MkdirAll(playtestTreeDumpDir, 0755); err != nil {
		t.Fatalf("Could not create tree dump dir: %v", err)
	}

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, content, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			var nodes []widgetNode
			fyne.DoAndWait(func() {
				nodes = WalkWidgetTree(content)
			})
			dump := DumpWidgetTree(nodes)

			// Summary stats
			interactiveCount := 0
			for _, n := range nodes {
				if n.Interactive {
					interactiveCount++
				}
			}
			unlabeled := findUnlabeledTappables(nodes)
			header := fmt.Sprintf("Module: %s\nTotal widgets: %d\nInteractive: %d\nUnlabeled interactive: %d\n\n",
				m.name, len(nodes), interactiveCount, len(unlabeled))

			fullDump := header + dump
			path := filepath.Join(playtestTreeDumpDir, "tree-dump-"+m.name+".txt")
			if err := os.WriteFile(path, []byte(fullDump), 0644); err != nil {
				t.Logf("Warning: could not write tree dump: %v", err)
			} else {
				t.Logf("Tree dump saved: %s", path)
			}

			// Log for test output
			t.Logf("\n%s", fullDump)
		})
	}
}
