//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
	demo7gui "fecim-lattice-tools/module7-docs/pkg/gui"
)

// ---------------------------------------------------------------------------
// Shared playtest infrastructure
// ---------------------------------------------------------------------------

const playtestEnvVar = "FECIM_PLAYTEST"

func playtestSkipUnlessEnabled(t *testing.T) {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(playtestEnvVar))
	if v == "" || v == "0" || strings.EqualFold(v, "false") {
		t.Skipf("Skipping playtest: set %s=1 to enable", playtestEnvVar)
	}
}

// playtestModuleEntry describes one module for playtest purposes.
type playtestModuleEntry struct {
	name           string
	create         func() (moduleLifecycle, error)
	skipScreenshot bool // M2: font crash in test driver
	skipStart      bool // avoid Start() for deterministic tree/golden tests
	mnistGate      bool // require FECIM_MNIST_PLAYTEST=1
}

// playtestSafeModules returns modules safe for tree/accessibility/layout tests.
// Excludes M2 (font crash) and M3 (fyne.Do deadlock).
func playtestSafeModules() []playtestModuleEntry {
	return []playtestModuleEntry{
		{"hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil }, false, false, false},
		{"circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }, false, false, false},
		{"comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil }, false, false, false},
		{"eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil }, false, false, false},
		{"docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil }, false, false, false},
	}
}

// playtestScreenshotDir is the output directory for playtest screenshots.
const playtestScreenshotDir = "testdata/playtest/screenshots"

// playtestTreeDumpDir is the output directory for tree dump text files.
const playtestTreeDumpDir = "testdata/playtest"

// playtestSetupModule builds a module and returns the window + content.
// Does NOT call module.Start() (caller decides).
func playtestSetupModule(t *testing.T, fy fyne.App, entry playtestModuleEntry) (fyne.Window, fyne.CanvasObject, moduleLifecycle) {
	t.Helper()

	if entry.name == "hysteresis" {
		t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	}

	mod, err := entry.create()
	if err != nil {
		t.Fatalf("Failed to create %s module: %v", entry.name, err)
	}

	var w fyne.Window
	var content fyne.CanvasObject
	fyne.DoAndWait(func() {
		w = fy.NewWindow("Playtest - " + entry.name)
		content = mod.BuildContent(fy, w)
		w.SetContent(container.NewMax(content))
		w.Resize(fyne.NewSize(1200, 800))
		w.Show()
	})
	time.Sleep(150 * time.Millisecond) // let layout settle

	return w, content, mod
}

// ---------------------------------------------------------------------------
// widgetNode and tree walker
// ---------------------------------------------------------------------------

// widgetNode represents a single widget discovered during tree traversal.
type widgetNode struct {
	Depth       int
	TypeName    string
	Text        string
	Position    fyne.Position
	Size        fyne.Size
	Interactive bool
	A11yLabel   string
	Object      fyne.CanvasObject
}

// resolveText extracts human-readable text from a canvas object.
func resolveText(obj fyne.CanvasObject) string {
	if obj == nil {
		return ""
	}
	switch o := obj.(type) {
	case *widget.Button:
		return o.Text
	case *widget.Label:
		return o.Text
	case *widget.Entry:
		if o.Text != "" {
			return o.Text
		}
		return o.PlaceHolder
	case *widget.Select:
		if o.Selected != "" {
			return o.Selected
		}
		return o.PlaceHolder
	case *widget.Check:
		return o.Text
	case *widget.RadioGroup:
		return o.Selected
	case *widget.Hyperlink:
		return o.Text
	case *widget.Slider:
		return fmt.Sprintf("%.4g", o.Value)
	case *canvas.Text:
		return o.Text
	case *widget.RichText:
		var parts []string
		for _, seg := range o.Segments {
			if ts, ok := seg.(*widget.TextSegment); ok {
				parts = append(parts, ts.Text)
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}

// isInteractive returns true if the object can be interacted with.
func isInteractive(obj fyne.CanvasObject) bool {
	if obj == nil {
		return false
	}
	if _, ok := obj.(fyne.Tappable); ok {
		return true
	}
	if _, ok := obj.(fyne.Focusable); ok {
		return true
	}
	switch obj.(type) {
	case *widget.Select, *widget.Slider:
		return true
	}
	return false
}

// WalkWidgetTree recursively walks a canvas object tree and returns all discovered nodes.
func WalkWidgetTree(root fyne.CanvasObject) []widgetNode {
	seen := map[uintptr]bool{}
	var nodes []widgetNode

	var walk func(o fyne.CanvasObject, depth int)
	walk = func(o fyne.CanvasObject, depth int) {
		if o == nil {
			return
		}

		ptr := ptrID(o)
		if ptr != 0 && seen[ptr] {
			return
		}
		if ptr != 0 {
			seen[ptr] = true
		}

		typeName := fmt.Sprintf("%T", o)
		if idx := strings.LastIndex(typeName, "."); idx >= 0 {
			typeName = typeName[idx+1:]
		}

		a11yLabel := ""
		if label, ok := sharedwidgets.GetAccessibleLabel(o); ok {
			a11yLabel = label
		}

		node := widgetNode{
			Depth:       depth,
			TypeName:    typeName,
			Text:        resolveText(o),
			Position:    o.Position(),
			Size:        o.Size(),
			Interactive: isInteractive(o),
			A11yLabel:   a11yLabel,
			Object:      o,
		}
		nodes = append(nodes, node)

		// Recurse into children
		switch obj := o.(type) {
		case *container.AppTabs:
			for _, item := range obj.Items {
				walk(item.Content, depth+1)
			}
			return
		case *container.DocTabs:
			for _, item := range obj.Items {
				walk(item.Content, depth+1)
			}
			return
		case *container.Split:
			walk(obj.Leading, depth+1)
			walk(obj.Trailing, depth+1)
			return
		case *fyne.Container:
			for _, child := range obj.Objects {
				walk(child, depth+1)
			}
			return
		}

		// Reflection fallback for wrappers (Scroll, Accordion, etc.)
		v := reflect.ValueOf(o)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, fieldName := range []string{"Content", "content", "Objects", "objects", "Leading", "Trailing"} {
				f := v.FieldByName(fieldName)
				if !f.IsValid() || !f.CanInterface() {
					continue
				}
				if child, ok := f.Interface().(fyne.CanvasObject); ok {
					walk(child, depth+1)
				} else if children, ok := f.Interface().([]fyne.CanvasObject); ok {
					for _, child := range children {
						walk(child, depth+1)
					}
				}
			}
		}
	}

	walk(root, 0)
	return nodes
}

// DumpWidgetTree produces a screen-reader-friendly plain text dump of the widget tree.
func DumpWidgetTree(nodes []widgetNode) string {
	var sb strings.Builder
	sb.WriteString("Widget Tree Dump\n")
	sb.WriteString("================\n")
	sb.WriteString("Note: canvas.Text inside widget renderers not visible to this walker.\n\n")

	for _, n := range nodes {
		indent := strings.Repeat("  ", n.Depth)
		prefix := ""
		if n.Interactive && n.A11yLabel == "" && n.Text == "" {
			prefix = "[UNLABELED] "
		}

		label := n.Text
		if n.A11yLabel != "" {
			label = n.A11yLabel
		}

		if label != "" {
			sb.WriteString(fmt.Sprintf("%s%s%s: %q  (%.0fx%.0f at %.0f,%.0f)\n",
				indent, prefix, n.TypeName, label,
				n.Size.Width, n.Size.Height, n.Position.X, n.Position.Y))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s%s  (%.0fx%.0f at %.0f,%.0f)\n",
				indent, prefix, n.TypeName,
				n.Size.Width, n.Size.Height, n.Position.X, n.Position.Y))
		}
	}
	return sb.String()
}

// findUnlabeledTappables returns interactive widgets that have no text and no a11y label.
func findUnlabeledTappables(nodes []widgetNode) []widgetNode {
	var result []widgetNode
	for _, n := range nodes {
		if n.Interactive && n.Text == "" && n.A11yLabel == "" {
			result = append(result, n)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPlaytestTreeWalker_Hysteresis(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	testPlaytestTreeWalkerModule(t, "hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil })
}

func TestPlaytestTreeWalker_Circuits(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestTreeWalkerModule(t, "circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil })
}

func TestPlaytestTreeWalker_Comparison(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestTreeWalkerModule(t, "comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil })
}

func TestPlaytestTreeWalker_EDA(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestTreeWalkerModule(t, "eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil })
}

func TestPlaytestTreeWalker_Docs(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	testPlaytestTreeWalkerModule(t, "docs", func() (moduleLifecycle, error) { return demo7gui.NewEmbeddedDocsApp(), nil })
}

func testPlaytestTreeWalkerModule(t *testing.T, name string, create func() (moduleLifecycle, error)) {
	t.Helper()

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	mod, err := create()
	if err != nil {
		t.Fatalf("Failed to create %s: %v", name, err)
	}

	var w fyne.Window
	var content fyne.CanvasObject
	fyne.DoAndWait(func() {
		w = fy.NewWindow("TreeWalker - " + name)
		content = mod.BuildContent(fy, w)
		w.SetContent(container.NewMax(content))
		w.Resize(fyne.NewSize(1200, 800))
		w.Show()
	})
	t.Cleanup(func() {
		fyne.DoAndWait(func() { w.Close() })
	})

	time.Sleep(200 * time.Millisecond)

	var nodes []widgetNode
	fyne.DoAndWait(func() {
		nodes = WalkWidgetTree(content)
	})

	dump := DumpWidgetTree(nodes)
	t.Logf("--- %s widget tree ---\n%s", name, dump)

	// Count stats
	totalCount := len(nodes)
	interactiveCount := 0
	for _, n := range nodes {
		if n.Interactive {
			interactiveCount++
		}
	}
	unlabeled := findUnlabeledTappables(nodes)
	t.Logf("%s: total=%d interactive=%d unlabeled=%d", name, totalCount, interactiveCount, len(unlabeled))

	if totalCount == 0 {
		t.Errorf("%s: widget tree is empty", name)
	}

	// Save dump to file
	if err := os.MkdirAll(playtestTreeDumpDir, 0755); err != nil {
		t.Logf("Warning: could not create tree dump dir: %v", err)
		return
	}
	dumpPath := filepath.Join(playtestTreeDumpDir, "tree-dump-"+name+".txt")
	if err := os.WriteFile(dumpPath, []byte(dump), 0644); err != nil {
		t.Logf("Warning: could not write tree dump: %v", err)
	} else {
		t.Logf("Tree dump saved: %s", dumpPath)
	}
}
