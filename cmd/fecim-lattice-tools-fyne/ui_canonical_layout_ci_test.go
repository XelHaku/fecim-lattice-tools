package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

func TestCanonicalSizeLayout_NoOutOfBoundsOrZeroSize(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("canonical-layout-ci")
	defer w.Close()

	root := CreateLauncherContent(nil)
	w.SetContent(container.NewMax(root))
	w.Show()

	sizes := []fyne.Size{
		fyne.NewSize(1024, 768),
		fyne.NewSize(1200, 800),
		fyne.NewSize(1366, 768),
	}

	for _, size := range sizes {
		size := size
		t.Run(fmt.Sprintf("%dx%d", int(size.Width), int(size.Height)), func(t *testing.T) {
			fyne.DoAndWait(func() {
				w.Resize(size)
				w.Canvas().Refresh(root)
			})
			time.Sleep(50 * time.Millisecond)

			nodes := flattenCanvasNodes(root)
			for _, node := range nodes {
				obj := node.obj
				if obj == nil || !obj.Visible() {
					continue
				}

				if _, ok := obj.(fyne.Widget); ok {
					sz := obj.Size()
					if sz.Width <= 0 || sz.Height <= 0 {
						t.Fatalf("zero-size widget %T at pos=%v size=%v", obj, obj.Position(), sz)
					}
				}

				if node.inScroll {
					continue
				}
				pos := obj.Position()
				sz := obj.Size()
				if pos.X < 0 || pos.Y < 0 || pos.X+sz.Width > size.Width+1 || pos.Y+sz.Height > size.Height+1 {
					t.Fatalf("out-of-bounds object %T pos=%v size=%v window=%v", obj, pos, sz, size)
				}
			}
		})
	}
}

type canvasNode struct {
	obj      fyne.CanvasObject
	inScroll bool
}

func flattenCanvasNodes(root fyne.CanvasObject) []canvasNode {
	seen := map[uintptr]bool{}
	out := make([]canvasNode, 0, 128)

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
		out = append(out, canvasNode{obj: o, inScroll: inScroll})

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
