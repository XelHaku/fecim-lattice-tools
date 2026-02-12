package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestSelectorCycleAndNavigation(t *testing.T) {
	selector := widget.NewSelect([]string{"View A", "View B", "View C"}, nil)
	selector.SetSelected("View A")

	nextView(selector)
	if selector.Selected != "View B" {
		t.Fatalf("nextView() = %q, want %q", selector.Selected, "View B")
	}

	prevView(selector)
	if selector.Selected != "View A" {
		t.Fatalf("prevView() = %q, want %q", selector.Selected, "View A")
	}

	prevView(selector)
	if selector.Selected != "View C" {
		t.Fatalf("prevView() wrap = %q, want %q", selector.Selected, "View C")
	}

	cycleViewSelector(selector)
	if selector.Selected != "View A" {
		t.Fatalf("cycleViewSelector() wrap = %q, want %q", selector.Selected, "View A")
	}
}

func TestSelectorNavigation_NilAndEmpty(t *testing.T) {
	cycleViewSelector(nil)
	nextView(nil)
	prevView(nil)

	empty := widget.NewSelect([]string{}, nil)
	cycleViewSelector(empty)
	nextView(empty)
	prevView(empty)

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

	selectWidget, stack := findSelectAndTwoViewStack(w.Content())
	if selectWidget == nil {
		t.Fatal("did not find view selector in window content")
	}
	if stack == nil || len(stack.Objects) != 2 {
		t.Fatal("did not find 2-view stack container in window content")
	}

	if !stack.Objects[0].Visible() || stack.Objects[1].Visible() {
		t.Fatalf("initial visibility unexpected: first=%v second=%v", stack.Objects[0].Visible(), stack.Objects[1].Visible())
	}

	// Same selection should be a no-op.
	selectWidget.SetSelected("1. Builder & Validation")
	if !stack.Objects[0].Visible() || stack.Objects[1].Visible() {
		t.Fatalf("same-selection should not toggle visibility: first=%v second=%v", stack.Objects[0].Visible(), stack.Objects[1].Visible())
	}

	selectWidget.SetSelected("2. Learn")
	if stack.Objects[0].Visible() || !stack.Objects[1].Visible() {
		t.Fatalf("after selecting Learn: first=%v second=%v", stack.Objects[0].Visible(), stack.Objects[1].Visible())
	}

	selectWidget.SetSelected("1. Builder & Validation")
	if !stack.Objects[0].Visible() || stack.Objects[1].Visible() {
		t.Fatalf("after selecting Builder: first=%v second=%v", stack.Objects[0].Visible(), stack.Objects[1].Visible())
	}
}

func findSelectAndTwoViewStack(root fyne.CanvasObject) (*widget.Select, *fyne.Container) {
	var foundSelect *widget.Select
	var foundStack *fyne.Container

	var walk func(obj fyne.CanvasObject)
	walk = func(obj fyne.CanvasObject) {
		if obj == nil {
			return
		}
		if sel, ok := obj.(*widget.Select); ok && foundSelect == nil {
			foundSelect = sel
		}
		if c, ok := obj.(*fyne.Container); ok {
			if len(c.Objects) == 2 && c.Objects[0].Visible() && !c.Objects[1].Visible() && foundStack == nil {
				foundStack = c
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
	return foundSelect, foundStack
}
