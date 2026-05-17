//go:build legacy_fyne

package tabs

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

func TestMakeLearnTab_ContentScrollUsesCompactMinSize(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	win := test.NewWindow(nil)
	defer win.Close()

	content := MakeLearnTab(nil, win)

	var maxScrollWidth float32
	walkLearnCanvas(content, func(obj fyne.CanvasObject) {
		s, ok := obj.(*container.Scroll)
		if !ok {
			return
		}
		if w := s.MinSize().Width; w > maxScrollWidth {
			maxScrollWidth = w
		}
	})

	if maxScrollWidth > 500 {
		t.Fatalf("learn tab scroll min width too large for narrow windows: got %.1fpx, want <= 500px", maxScrollWidth)
	}
}

func walkLearnCanvas(obj fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	if obj == nil {
		return
	}
	visit(obj)
	switch v := obj.(type) {
	case *fyne.Container:
		for _, child := range v.Objects {
			walkLearnCanvas(child, visit)
		}
	case *container.Scroll:
		walkLearnCanvas(v.Content, visit)
	}
}
