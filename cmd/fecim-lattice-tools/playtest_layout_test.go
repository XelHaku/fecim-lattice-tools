//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// ---------------------------------------------------------------------------
// Layout audit types and helpers
// ---------------------------------------------------------------------------

// boundEntry records a widget's absolute position and size.
type boundEntry struct {
	TypeName string
	Text     string
	X, Y     float32
	W, H     float32
}

type boundRect struct {
	X, Y float32
	W, H float32
}

// collectBounds walks the tree and records position/size of every visible widget.
func collectBounds(root fyne.CanvasObject) []boundEntry {
	seen := map[uintptr]bool{}
	var entries []boundEntry

	var walk func(o fyne.CanvasObject, parentPos fyne.Position, clip *boundRect)
	walk = func(o fyne.CanvasObject, parentPos fyne.Position, clip *boundRect) {
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

		if !o.Visible() {
			return
		}

		pos := o.Position()
		absPos := fyne.NewPos(parentPos.X+pos.X, parentPos.Y+pos.Y)
		sz := o.Size()
		if sz.Width > 0 && sz.Height > 0 {
			rect := boundRect{X: absPos.X, Y: absPos.Y, W: sz.Width, H: sz.Height}
			if clip != nil {
				var ok bool
				rect, ok = intersectRect(rect, *clip)
				if !ok {
					return
				}
			}
			typeName := fmt.Sprintf("%T", o)
			if idx := strings.LastIndex(typeName, "."); idx >= 0 {
				typeName = typeName[idx+1:]
			}
			entries = append(entries, boundEntry{
				TypeName: typeName,
				Text:     resolveText(o),
				X:        rect.X,
				Y:        rect.Y,
				W:        rect.W,
				H:        rect.H,
			})
		}

		// Recurse
		switch obj := o.(type) {
		case *container.AppTabs:
			for _, item := range obj.Items {
				walk(item.Content, absPos, clip)
			}
			return
		case *container.DocTabs:
			for _, item := range obj.Items {
				walk(item.Content, absPos, clip)
			}
			return
		case *container.Scroll:
			scrollClip := boundRect{X: absPos.X, Y: absPos.Y, W: sz.Width, H: sz.Height}
			if clip != nil {
				var ok bool
				scrollClip, ok = intersectRect(scrollClip, *clip)
				if !ok {
					return
				}
			}
			walk(obj.Content, absPos, &scrollClip)
			return
		case *container.Split:
			walk(obj.Leading, absPos, clip)
			walk(obj.Trailing, absPos, clip)
			return
		case *fyne.Container:
			for _, child := range obj.Objects {
				walk(child, absPos, clip)
			}
			return
		}

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
					walk(child, absPos, clip)
				} else if children, ok := f.Interface().([]fyne.CanvasObject); ok {
					for _, child := range children {
						walk(child, absPos, clip)
					}
				}
			}
		}
	}

	walk(root, fyne.NewPos(0, 0), nil)
	return entries
}

func TestCollectBounds_UsesAbsolutePositions(t *testing.T) {
	root := container.NewWithoutLayout()
	parent := container.NewWithoutLayout()
	parent.Move(fyne.NewPos(20, 30))
	parent.Resize(fyne.NewSize(200, 150))

	child := container.NewWithoutLayout()
	child.Move(fyne.NewPos(15, 25))
	child.Resize(fyne.NewSize(80, 60))

	parent.Add(child)
	root.Add(parent)

	entries := collectBounds(root)
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}

	found := false
	for _, entry := range entries {
		if entry.W == 80 && entry.H == 60 {
			found = true
			if entry.X != 35 || entry.Y != 55 {
				t.Fatalf("expected absolute child position 35x55, got %.0fx%.0f", entry.X, entry.Y)
			}
		}
	}
	if !found {
		t.Fatal("expected child entry not found")
	}
}

func TestCollectBounds_ClipsScrollDescendantsToViewport(t *testing.T) {
	root := container.NewWithoutLayout()
	root.Resize(fyne.NewSize(240, 180))

	content := container.NewWithoutLayout()
	content.Resize(fyne.NewSize(300, 300))

	clipped := widget.NewLabel("clipped")
	clipped.Move(fyne.NewPos(0, 80))
	clipped.Resize(fyne.NewSize(160, 60))
	content.Add(clipped)

	scroll := container.NewScroll(content)
	scroll.Move(fyne.NewPos(10, 20))
	scroll.Resize(fyne.NewSize(100, 100))
	root.Add(scroll)

	entries := collectBounds(root)
	for _, entry := range entries {
		if entry.Text != "clipped" {
			continue
		}
		if entry.X != 10 || entry.Y != 100 || entry.W != 100 || entry.H != 20 {
			t.Fatalf("expected clipped bounds 10,100 100x20, got %.0f,%.0f %.0fx%.0f", entry.X, entry.Y, entry.W, entry.H)
		}
		return
	}
	t.Fatal("expected clipped label entry not found")
}

// detectOverlaps checks all pairs of bound entries for bounding-box intersection.
// Returns human-readable descriptions of overlapping pairs.
func detectOverlaps(entries []boundEntry) []string {
	var overlaps []string
	for i := 0; i < len(entries); i++ {
		a := entries[i]
		if a.TypeName == "Spacer" {
			continue
		}
		for j := i + 1; j < len(entries); j++ {
			b := entries[j]
			if b.TypeName == "Spacer" {
				continue
			}
			// Skip if one is fully inside the other (parent-child nesting is normal)
			if contains(a, b) || contains(b, a) {
				continue
			}
			if intersects(a, b) {
				overlaps = append(overlaps, fmt.Sprintf(
					"%s %q (%.0f,%.0f %.0fx%.0f) overlaps %s %q (%.0f,%.0f %.0fx%.0f)",
					a.TypeName, a.Text, a.X, a.Y, a.W, a.H,
					b.TypeName, b.Text, b.X, b.Y, b.W, b.H))
			}
		}
	}
	return overlaps
}

func intersects(a, b boundEntry) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X &&
		a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func contains(outer, inner boundEntry) bool {
	return inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.W <= outer.X+outer.W &&
		inner.Y+inner.H <= outer.Y+outer.H
}

func intersectsRect(a, b boundRect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X &&
		a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func intersectRect(a, b boundRect) (boundRect, bool) {
	if !intersectsRect(a, b) {
		return boundRect{}, false
	}
	x1 := maxFloat32(a.X, b.X)
	y1 := maxFloat32(a.Y, b.Y)
	x2 := minFloat32(a.X+a.W, b.X+b.W)
	y2 := minFloat32(a.Y+a.H, b.Y+b.H)
	if x2 <= x1 || y2 <= y1 {
		return boundRect{}, false
	}
	return boundRect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}, true
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// checkMinSizes compares each widget's MinSize() against its actual Size().
func checkMinSizes(t *testing.T, root fyne.CanvasObject, module string, viewportSize fyne.Size) {
	t.Helper()
	seen := map[uintptr]bool{}
	violations := 0

	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
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

		if !o.Visible() {
			return
		}

		minSz := o.MinSize()
		sz := o.Size()
		// Only flag significant undersize (>2dp tolerance for rounding)
		if sz.Width > 0 && sz.Height > 0 {
			if sz.Width+2 < minSz.Width || sz.Height+2 < minSz.Height {
				typeName := fmt.Sprintf("%T", o)
				if idx := strings.LastIndex(typeName, "."); idx >= 0 {
					typeName = typeName[idx+1:]
				}
				t.Logf("[LAYOUT] %s@%.0fx%.0f: %s %q size %.0fx%.0f < minSize %.0fx%.0f",
					module, viewportSize.Width, viewportSize.Height,
					typeName, resolveText(o),
					sz.Width, sz.Height, minSz.Width, minSz.Height)
				violations++
			}
		}

		switch obj := o.(type) {
		case *container.AppTabs:
			for _, item := range obj.Items {
				walk(item.Content)
			}
			return
		case *container.DocTabs:
			for _, item := range obj.Items {
				walk(item.Content)
			}
			return
		case *container.Split:
			walk(obj.Leading)
			walk(obj.Trailing)
			return
		case *fyne.Container:
			for _, child := range obj.Objects {
				walk(child)
			}
			return
		}

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
					walk(child)
				} else if children, ok := f.Interface().([]fyne.CanvasObject); ok {
					for _, child := range children {
						walk(child)
					}
				}
			}
		}
	}

	walk(root)
	if violations > 0 {
		t.Logf("[LAYOUT] %s@%.0fx%.0f: %d widget(s) smaller than MinSize",
			module, viewportSize.Width, viewportSize.Height, violations)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

var playtestViewportSizes = []struct {
	name string
	w, h float32
}{
	{"320x480", 320, 480},
	{"375x812", 375, 812},
	{"768x1024", 768, 1024},
	{"1024x768", 1024, 768},
	{"1200x800", 1200, 800},
}

func TestPlaytestLayout_BoundsAtMultipleSizes(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	if err := os.MkdirAll(playtestScreenshotDir, 0755); err != nil {
		t.Logf("Warning: could not create screenshot dir: %v", err)
	}

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, content, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			for _, sz := range playtestViewportSizes {
				sz := sz
				t.Run(sz.name, func(t *testing.T) {
					fyne.DoAndWait(func() {
						w.Resize(fyne.NewSize(sz.w, sz.h))
					})
					time.Sleep(120 * time.Millisecond)

					var entries []boundEntry
					fyne.DoAndWait(func() {
						entries = collectBounds(content)
					})
					t.Logf("%s@%s: %d widgets with bounds", m.name, sz.name, len(entries))

					// Capture screenshot
					var img image.Image
					fyne.DoAndWait(func() {
						img = captureWindow(w)
					})
					fname := filepath.Join(playtestScreenshotDir, fmt.Sprintf("playtest_%s_%s.png", m.name, sz.name))
					if f, err := os.Create(fname); err == nil {
						_ = png.Encode(f, img)
						f.Close()
						t.Logf("Screenshot saved: %s", fname)
					}
				})
			}
		})
	}
}

func TestPlaytestLayout_OverlapDetection(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, content, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			for _, sz := range playtestViewportSizes {
				sz := sz
				t.Run(sz.name, func(t *testing.T) {
					fyne.DoAndWait(func() {
						w.Resize(fyne.NewSize(sz.w, sz.h))
					})
					time.Sleep(120 * time.Millisecond)

					var entries []boundEntry
					fyne.DoAndWait(func() {
						entries = collectBounds(content)
					})

					overlaps := detectOverlaps(entries)
					for _, msg := range overlaps {
						t.Logf("[LAYOUT] %s@%s: %s", m.name, sz.name, msg)
					}
					if len(overlaps) > 0 {
						t.Logf("[LAYOUT] %s@%s: %d overlap(s) detected", m.name, sz.name, len(overlaps))
					}
				})
			}
		})
	}
}

func TestPlaytestLayout_MinSizeValidation(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	fy := test.NewApp()
	t.Cleanup(func() { fy.Quit() })

	for _, m := range playtestSafeModules() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			w, content, _ := playtestSetupModule(t, fy, m)
			t.Cleanup(func() { fyne.DoAndWait(func() { w.Close() }) })

			for _, sz := range playtestViewportSizes {
				sz := sz
				t.Run(sz.name, func(t *testing.T) {
					viewportSize := fyne.NewSize(sz.w, sz.h)
					fyne.DoAndWait(func() {
						w.Resize(viewportSize)
					})
					time.Sleep(120 * time.Millisecond)

					fyne.DoAndWait(func() {
						checkMinSizes(t, content, m.name, viewportSize)
					})
				})
			}
		})
	}
}
