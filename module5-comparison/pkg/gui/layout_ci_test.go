//go:build legacy_fyne

package gui

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
)

func TestModule5LayoutCI_NoOverlapAtCommonResolutions(t *testing.T) {
	app := fyneTest.NewApp()
	defer app.Quit()

	ca := &ComparisonApp{currentWorkload: "GPT-2", currentInferences: 10000, uiMode: "Technical Review", scenarioProfile: ScenarioBaseline}
	w := app.NewWindow("m5-layout-ci")
	defer w.Close()

	root := ca.createMainLayout()
	w.SetContent(container.NewMax(root))
	w.Show()

	sizes := []fyne.Size{fyne.NewSize(1024, 768), fyne.NewSize(1200, 800), fyne.NewSize(1366, 768)}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", int(size.Width), int(size.Height)), func(t *testing.T) {
			fyne.DoAndWait(func() {
				w.Resize(size)
				w.Canvas().Refresh(root)
			})
			time.Sleep(50 * time.Millisecond)
			nodes := flattenNodes(root)
			for _, n := range nodes {
				obj := n.obj
				if obj == nil || !obj.Visible() {
					continue
				}
				if _, ok := obj.(fyne.Widget); ok {
					sz := obj.Size()
					if sz.Width <= 0 || sz.Height <= 0 {
						t.Fatalf("zero-size widget %T", obj)
					}
				}
				if n.inScroll {
					continue
				}
				p, s := obj.Position(), obj.Size()
				if p.X < 0 || p.Y < 0 || p.X+s.Width > size.Width+1 || p.Y+s.Height > size.Height+1 {
					t.Fatalf("out-of-bounds object %T pos=%v size=%v win=%v", obj, p, s, size)
				}
			}
		})
	}
}

type nodeWithScroll struct {
	obj      fyne.CanvasObject
	inScroll bool
}

func ptrID(o fyne.CanvasObject) uintptr {
	v := reflect.ValueOf(o)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		return 0
	}
	return v.Pointer()
}

func flattenNodes(root fyne.CanvasObject) []nodeWithScroll {
	seen := map[uintptr]bool{}
	out := make([]nodeWithScroll, 0, 128)
	var walk func(fyne.CanvasObject, bool)
	walk = func(o fyne.CanvasObject, inScroll bool) {
		if o == nil {
			return
		}
		if p := ptrID(o); p != 0 {
			if seen[p] {
				return
			}
			seen[p] = true
		}
		out = append(out, nodeWithScroll{obj: o, inScroll: inScroll})
		if _, ok := o.(*container.Scroll); ok {
			inScroll = true
		}
		if c, ok := o.(*fyne.Container); ok {
			for _, child := range c.Objects {
				walk(child, inScroll)
			}
			return
		}
		if tabs, ok := o.(*container.AppTabs); ok {
			for _, item := range tabs.Items {
				walk(item.Content, inScroll)
			}
			return
		}
		v := reflect.ValueOf(o)
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if v.IsValid() && v.Kind() == reflect.Struct {
			for _, fieldName := range []string{"Content", "content"} {
				f := v.FieldByName(fieldName)
				if f.IsValid() && f.CanInterface() {
					if child, ok := f.Interface().(fyne.CanvasObject); ok {
						walk(child, inScroll)
					}
				}
			}
		}
	}
	walk(root, false)
	return out
}
