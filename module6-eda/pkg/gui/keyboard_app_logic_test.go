//go:build legacy_fyne

package gui

import (
	"fmt"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/keyboard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestSelectorCycleAndNavigation(t *testing.T) {
	selector := widget.NewSelect([]string{"View A", "View B", "View C"}, nil)
	selector.SetSelected("View A")

	keyboard.SelectNextOption(selector)
	if selector.Selected != "View B" {
		t.Fatalf("SelectNextOption() = %q, want %q", selector.Selected, "View B")
	}

	keyboard.SelectPrevOption(selector)
	if selector.Selected != "View A" {
		t.Fatalf("SelectPrevOption() = %q, want %q", selector.Selected, "View A")
	}

	keyboard.SelectPrevOption(selector)
	if selector.Selected != "View C" {
		t.Fatalf("SelectPrevOption() wrap = %q, want %q", selector.Selected, "View C")
	}

	keyboard.SelectNextOption(selector)
	if selector.Selected != "View A" {
		t.Fatalf("SelectNextOption() wrap = %q, want %q", selector.Selected, "View A")
	}
}

func TestSelectorNavigation_NilAndEmpty(t *testing.T) {
	keyboard.SelectNextOption(nil)
	keyboard.SelectNextOption(nil)
	keyboard.SelectPrevOption(nil)

	empty := widget.NewSelect([]string{}, nil)
	keyboard.SelectNextOption(empty)
	keyboard.SelectNextOption(empty)
	keyboard.SelectPrevOption(empty)

	if empty.Selected != "" {
		t.Fatalf("empty selector selected = %q, want empty", empty.Selected)
	}
}

func TestSetupKeyboard_OnTypedKeyNavigation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := app.NewWindow("keyboard")
	defer w.Close()

	selector := widget.NewSelect([]string{"One", "Two"}, nil)
	selector.SetSelected("One")
	SetupKeyboard(w, selector)

	onTypedKey := w.Canvas().OnTypedKey()
	if onTypedKey == nil {
		t.Fatal("expected typed-key handler to be registered")
	}

	onTypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	if selector.Selected != "Two" {
		t.Fatalf("KeyRight selected = %q, want %q", selector.Selected, "Two")
	}

	onTypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	if selector.Selected != "One" {
		t.Fatalf("KeyLeft selected = %q, want %q", selector.Selected, "One")
	}

	onTypedKey(&fyne.KeyEvent{Name: fyne.KeySpace})
	if selector.Selected != "Two" {
		t.Fatalf("KeySpace selected = %q, want %q", selector.Selected, "Two")
	}

	onTypedKey(&fyne.KeyEvent{Name: fyne.Key1})
	if selector.Selected != "One" {
		t.Fatalf("Key1 selected = %q, want %q", selector.Selected, "One")
	}

	onTypedKey(&fyne.KeyEvent{Name: fyne.Key2})
	if selector.Selected != "Two" {
		t.Fatalf("Key2 selected = %q, want %q", selector.Selected, "Two")
	}
}

func TestSetupKeyboard_NilSelectorAndHelpPath(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := app.NewWindow("keyboard")
	defer w.Close()

	SetupKeyboard(w, nil)
	onTypedKey := w.Canvas().OnTypedKey()
	if onTypedKey == nil {
		t.Fatal("expected typed-key handler to be registered")
	}

	// Exercise non-selector path and help dialog path.
	onTypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	onTypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	onTypedKey(&fyne.KeyEvent{Name: fyne.KeySlash})
}

func TestCreateMainWindow_ViewSelectionTogglesVisibleContent(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := CreateMainWindow(app)
	defer w.Close()

	selectWidget, stack := findSelectAndViewStack(w.Content())
	if selectWidget == nil {
		t.Fatal("did not find view selector in window content")
	}
	if stack == nil || len(stack.Objects) < 2 {
		t.Fatal("did not find multi-view stack container in window content")
	}

	if !stack.Objects[0].Visible() {
		t.Fatalf("initial visibility unexpected: first view should be visible")
	}
	for i := 1; i < len(stack.Objects); i++ {
		if stack.Objects[i].Visible() {
			t.Fatalf("initial visibility unexpected: view %d visible", i)
		}
	}

	// Same selection should be a no-op.
	selectWidget.SetSelected("1. Builder & Validation")
	if !stack.Objects[0].Visible() {
		t.Fatalf("same-selection should keep first view visible")
	}

	selectWidget.SetSelected("4. Learn")
	if stack.Objects[0].Visible() || !stack.Objects[3].Visible() {
		t.Fatalf("after selecting Learn: builder=%v learn=%v", stack.Objects[0].Visible(), stack.Objects[3].Visible())
	}

	selectWidget.SetSelected("2. Export Viewer")
	if !stack.Objects[1].Visible() {
		t.Fatalf("after selecting Export Viewer: second view not visible")
	}
}

func findSelectAndViewStack(root fyne.CanvasObject) (*widget.Select, *fyne.Container) {
	var foundSelect *widget.Select
	var candidates []*fyne.Container

	var walk func(obj fyne.CanvasObject)
	walk = func(obj fyne.CanvasObject) {
		if obj == nil {
			return
		}
		if sel, ok := obj.(*widget.Select); ok {
			if len(sel.Options) > 0 && strings.Contains(sel.Options[0], "Builder & Validation") {
				foundSelect = sel
			}
		}
		if c, ok := obj.(*fyne.Container); ok {
			if len(c.Objects) >= 2 {
				candidates = append(candidates, c)
			}
			for _, child := range c.Objects {
				walk(child)
			}
		}
		if tabs, ok := obj.(*container.AppTabs); ok {
			for _, item := range tabs.Items {
				walk(item.Content)
			}
		}
	}

	walk(root)

	if foundSelect == nil {
		return nil, nil
	}
	for _, c := range candidates {
		if len(c.Objects) != len(foundSelect.Options) {
			continue
		}
		if c.Layout == nil || !strings.Contains(fmt.Sprintf("%T", c.Layout), "stack") {
			continue
		}
		visibleCount := 0
		for _, o := range c.Objects {
			if o.Visible() {
				visibleCount++
			}
		}
		if visibleCount == 1 {
			return foundSelect, c
		}
	}

	return foundSelect, nil
}
