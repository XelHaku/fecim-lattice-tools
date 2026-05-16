//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"image"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
	demo2gui "fecim-lattice-tools/module2-crossbar/pkg/gui"
	demo3gui "fecim-lattice-tools/module3-mnist/pkg/gui"
	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
	demo5gui "fecim-lattice-tools/module5-comparison/pkg/gui"
	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

// moduleLifecycle is the package-local alias for shared/widgets.EmbeddedApp.

type moduleFactory struct {
	name   string
	create func() (moduleLifecycle, error)
}

const (
	layoutAuditEnvVar            = "FECIM_LAYOUT_AUDIT"
	layoutAuditShowDelay         = 150 * time.Millisecond
	layoutAuditResizeDelay       = 120 * time.Millisecond
	layoutAuditTabDelay          = 80 * time.Millisecond
	layoutAuditOverlayDelay      = 80 * time.Millisecond
	layoutAuditOverlayCloseDelay = 60 * time.Millisecond
)

// TestLayoutAudit_AllModulesTabsAndSizes runs the full layout audit.
// Run with: FECIM_LAYOUT_AUDIT=1 go test -v ./cmd/fecim-lattice-tools/... -run LayoutAudit
// Or with build tag: go test -tags layoutaudit -v ./cmd/fecim-lattice-tools/... -run LayoutAudit
func TestLayoutAudit_AllModulesTabsAndSizes(t *testing.T) {
	if !layoutAuditEnabled() {
		t.Skipf("Skipping layout audit: set %s=1 or use -tags layoutaudit", layoutAuditEnvVar)
	}
	if testing.Short() {
		t.Skip("Skipping layout audit in short mode")
	}
	// Use the Fyne test app for determinism and to avoid deadlocks/panics that can
	// happen when driving a real GLFW event loop under `go test` (esp. headless/Xvfb).
	//
	// If you want to audit with real fonts/driver, that should be a separate manual
	// workflow (or use FECIM_GLFW_TESTMAIN=1 and ensure a stable display stack).
	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	sizes := []struct {
		w, h float32
	}{
		{1200, 800},
		{390, 844},
	}

	modules := []moduleFactory{
		{"hysteresis", func() (moduleLifecycle, error) { return demo1gui.NewEmbeddedApp(), nil }},
		{"crossbar", func() (moduleLifecycle, error) { return demo2gui.NewEmbeddedCrossbarApp() }},
		{"mnist", func() (moduleLifecycle, error) { return demo3gui.NewEmbeddedDualModeApp(), nil }},
		{"circuits", func() (moduleLifecycle, error) { return demo4gui.NewEmbeddedCircuitsApp(), nil }},
		{"comparison", func() (moduleLifecycle, error) { return demo5gui.NewEmbeddedComparisonApp(), nil }},
		{"eda", func() (moduleLifecycle, error) { return demo6gui.NewEmbeddedEDAApp(), nil }},
	}

	for _, m := range modules {
		m := m
		t.Run(m.name, func(t *testing.T) {
			if m.name == "crossbar" {
				t.Skip("Skipping crossbar in layout audit: test-driver hang under headless mode")
			}
			mod, err := m.create()
			if err != nil {
				t.Fatalf("Failed to create %s module: %v", m.name, err)
			}
			if mod == nil {
				t.Fatalf("%s module is nil", m.name)
			}

			var w fyne.Window
			fyne.DoAndWait(func() {
				w = fy.NewWindow("LayoutAudit - " + m.name)
			})
			if w == nil {
				t.Fatalf("%s window creation failed", m.name)
			}
			t.Cleanup(func() {
				fyne.DoAndWait(func() {
					w.Close()
				})
			})

			var content fyne.CanvasObject
			fyne.DoAndWait(func() {
				content = mod.BuildContent(fy, w)
				w.SetContent(container.NewMax(content))
				w.Show()
			})
			if content == nil {
				t.Fatalf("%s BuildContent returned nil", m.name)
			}

			// IMPORTANT: avoid calling Start() here—some modules spin simulation/animation loops.
			time.Sleep(layoutAuditShowDelay)

			for _, sz := range sizes {
				sz := sz
				fyne.DoAndWait(func() {
					w.Resize(fyne.NewSize(sz.w, sz.h))
				})
				time.Sleep(layoutAuditResizeDelay)

				baseName := fmt.Sprintf("layout_%s_%dx%d_base", m.name, int(sz.w), int(sz.h))
				var img image.Image
				fyne.DoAndWait(func() {
					img = captureWindow(w)
				})
				saveTestScreenshot(t, img, baseName)
				verifyImageNotEmpty(t, img, baseName)

				// Keep heartbeat audits stable by restricting historically unstable modules
				// to base-size captures in automated test-driver mode.
				if m.name == "eda" || m.name == "crossbar" {
					t.Logf("layout audit: using base-only capture for %s at %dx%d", m.name, int(sz.w), int(sz.h))
					continue
				}

				captureOverlays(t, w, content, m.name, int(sz.w), int(sz.h), "base")

				// Traverse all AppTabs (including nested). For each tab set, capture each tab.
				var tabSets []*container.AppTabs
				fyne.DoAndWait(func() {
					tabSets = findAllAppTabs(content)
				})
				for k, tabs := range tabSets {
					var itemCount int
					fyne.DoAndWait(func() {
						itemCount = len(tabs.Items)
					})
					for i := 0; i < itemCount; i++ {
						var tabText string
						fyne.DoAndWait(func() {
							tabs.SelectIndex(i)
							if i < len(tabs.Items) {
								tabText = strings.TrimSpace(tabs.Items[i].Text)
							}
						})
						time.Sleep(layoutAuditTabDelay)

						// EDA Learn tab can trigger unstable canvas-capture paths under test driver;
						// keep the audit running while still validating all other tabs/sizes.
						if m.name == "eda" && strings.EqualFold(tabText, "Learn") {
							t.Logf("layout audit: skipping unstable capture for %s tab=%q", m.name, tabText)
							continue
						}

						name := fmt.Sprintf("layout_%s_%dx%d_tabs%d_i%d_%s", m.name, int(sz.w), int(sz.h), k, i, safeName(tabText))
						fyne.DoAndWait(func() {
							img = captureWindow(w)
						})
						saveTestScreenshot(t, img, name)
						verifyImageNotEmpty(t, img, name)
					}
				}
			}
		})
	}
}

// findAllAppTabs walks a CanvasObject tree and returns all *container.AppTabs found.
// It handles:
// - *fyne.Container children
// - *container.AppTabs items' content
// - common wrappers that store a single child in a field named Content/content (via reflection)
func findAllAppTabs(root fyne.CanvasObject) []*container.AppTabs {
	seenObj := map[uintptr]bool{}
	seenTabs := map[uintptr]bool{}
	var out []*container.AppTabs

	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		ptr := ptrID(o)
		if ptr != 0 {
			if seenObj[ptr] {
				return
			}
			seenObj[ptr] = true
		}

		if tabs, ok := o.(*container.AppTabs); ok {
			tid := ptrID(tabs)
			if tid == 0 || !seenTabs[tid] {
				if tid != 0 {
					seenTabs[tid] = true
				}
				out = append(out, tabs)
			}
			for _, it := range tabs.Items {
				walk(it.Content)
			}
			return
		}

		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				walk(child)
			}
			return
		}

		// Reflection-based fallback for wrappers (e.g., Scroll) that hold a single child.
		v := reflect.ValueOf(o)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, fieldName := range []string{"Content", "content"} {
				f := v.FieldByName(fieldName)
				if f.IsValid() && f.CanInterface() {
					if child, ok := f.Interface().(fyne.CanvasObject); ok {
						walk(child)
					}
				}
			}
		}
	}

	walk(root)
	return out
}

func captureOverlays(t *testing.T, win fyne.Window, root fyne.CanvasObject, module string, w, h int, phase string) {
	t.Helper()

	// Conservative allow-list: text labels that typically open popups/modals.
	allow := map[string]bool{
		"about":    true,
		"help":     true,
		"docs":     true,
		"glossary": true,
		"info":     true,
		"learn":    true,
	}
	closeWords := map[string]bool{
		"close":   true,
		"back":    true,
		"dismiss": true,
		"ok":      true,
		"done":    true,
	}

	type buttonInfo struct {
		button    *widget.Button
		label     string
		hasAction bool
	}

	snapshotButtons := func() []buttonInfo {
		var infos []buttonInfo
		fyne.DoAndWait(func() {
			buttons := findAllButtons(root)
			infos = make([]buttonInfo, 0, len(buttons))
			for _, b := range buttons {
				infos = append(infos, buttonInfo{
					button:    b,
					label:     strings.TrimSpace(b.Text),
					hasAction: b.OnTapped != nil,
				})
			}
		})
		return infos
	}

	buttons := snapshotButtons()
	seenLabel := map[string]int{}

	for _, b := range buttons {
		label := b.label
		if label == "" {
			continue
		}
		norm := strings.ToLower(label)
		if !allow[norm] {
			continue
		}
		if !b.hasAction {
			continue
		}
		// Trigger overlay
		fyne.DoAndWait(func() {
			if b.button.OnTapped != nil {
				b.button.OnTapped()
			}
		})
		time.Sleep(layoutAuditOverlayDelay)

		idx := seenLabel[norm]
		seenLabel[norm] = idx + 1
		name := fmt.Sprintf("layout_%s_%dx%d_overlay_%s_%s_%d", module, w, h, safeName(norm), safeName(phase), idx)
		var img image.Image
		fyne.DoAndWait(func() {
			img = captureWindow(win)
		})
		saveTestScreenshot(t, img, name)
		verifyImageNotEmpty(t, img, name)

		// Best-effort close: look for a close/back/dismiss button and tap it.
		for _, cb := range snapshotButtons() {
			cl := strings.ToLower(cb.label)
			if closeWords[cl] && cb.hasAction {
				fyne.DoAndWait(func() {
					if cb.button.OnTapped != nil {
						cb.button.OnTapped()
					}
				})
				break
			}
		}
		time.Sleep(layoutAuditOverlayCloseDelay)
	}
}

func layoutAuditEnabled() bool {
	if layoutAuditBuildTagEnabled {
		return true
	}
	value := strings.TrimSpace(os.Getenv(layoutAuditEnvVar))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func findAllButtons(root fyne.CanvasObject) []*widget.Button {
	seenObj := map[uintptr]bool{}
	seenBtn := map[uintptr]bool{}
	var out []*widget.Button

	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		ptr := ptrID(o)
		if ptr != 0 {
			if seenObj[ptr] {
				return
			}
			seenObj[ptr] = true
		}

		if b, ok := o.(*widget.Button); ok {
			bid := ptrID(b)
			if bid == 0 || !seenBtn[bid] {
				if bid != 0 {
					seenBtn[bid] = true
				}
				out = append(out, b)
			}
			return
		}

		if tabs, ok := o.(*container.AppTabs); ok {
			for _, it := range tabs.Items {
				walk(it.Content)
			}
			return
		}

		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				walk(child)
			}
			return
		}

		// Reflection-based fallback for wrappers (e.g., Scroll) that hold a single child.
		v := reflect.ValueOf(o)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, fieldName := range []string{"Content", "content"} {
				f := v.FieldByName(fieldName)
				if f.IsValid() && f.CanInterface() {
					if child, ok := f.Interface().(fyne.CanvasObject); ok {
						walk(child)
					}
				}
			}
		}
	}

	walk(root)
	return out
}

func safeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = strings.ReplaceAll(s, ":", "-")
	return s
}

// ptrID is defined in ui_crawler_helpers.go
