package main

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func uiDo(fn func()) {
	fyne.DoAndWait(fn)
}

func saveGUISmokeScreenshot(t *testing.T, w fyne.Window, relPath string) {
	t.Helper()
	img := w.Canvas().Capture()
	if img == nil {
		t.Fatalf("captured screenshot is nil: %s", relPath)
	}
	abs := filepath.Join("testdata", "screenshots", "gui-smoke", relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir screenshot dir: %v", err)
	}
	f, err := os.Create(abs)
	if err != nil {
		t.Fatalf("create screenshot file: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode screenshot: %v", err)
	}
}

func walkCanvas(obj fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	if obj == nil {
		return
	}
	visit(obj)
	switch v := obj.(type) {
	case *fyne.Container:
		for _, child := range v.Objects {
			walkCanvas(child, visit)
		}
	case *container.AppTabs:
		for _, it := range v.Items {
			walkCanvas(it.Content, visit)
		}
	case *widget.Accordion:
		for _, it := range v.Items {
			walkCanvas(it.Detail, visit)
		}
	}
}

func findButton(root fyne.CanvasObject, text string) *widget.Button {
	var out *widget.Button
	walkCanvas(root, func(obj fyne.CanvasObject) {
		if out != nil {
			return
		}
		if b, ok := obj.(*widget.Button); ok && b.Text == text {
			out = b
		}
	})
	return out
}

func findButtonContains(root fyne.CanvasObject, needle string) *widget.Button {
	var out *widget.Button
	walkCanvas(root, func(obj fyne.CanvasObject) {
		if out != nil {
			return
		}
		if b, ok := obj.(*widget.Button); ok {
			if strings.Contains(b.Text, needle) {
				out = b
			}
		}
	})
	return out
}

func findSelectWithOptions(root fyne.CanvasObject, required ...string) *widget.Select {
	var out *widget.Select
	walkCanvas(root, func(obj fyne.CanvasObject) {
		if out != nil {
			return
		}
		s, ok := obj.(*widget.Select)
		if !ok {
			return
		}
		need := map[string]bool{}
		for _, r := range required {
			need[r] = true
		}
		for _, opt := range s.Options {
			delete(need, opt)
		}
		if len(need) == 0 {
			out = s
		}
	})
	return out
}

func findLabelByPrefix(root fyne.CanvasObject, prefix string) *widget.Label {
	var out *widget.Label
	walkCanvas(root, func(obj fyne.CanvasObject) {
		if out != nil {
			return
		}
		if l, ok := obj.(*widget.Label); ok {
			if len(l.Text) >= len(prefix) && l.Text[:len(prefix)] == prefix {
				out = l
			}
		}
	})
	return out
}
