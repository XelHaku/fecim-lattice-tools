package widgets

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// LazyTabItem defers heavy content construction until the tab is first selected.
type LazyTabItem struct {
	Title   string
	Builder func() fyne.CanvasObject

	once    sync.Once
	content fyne.CanvasObject
}

func NewLazyTabItem(title string, builder func() fyne.CanvasObject) *LazyTabItem {
	return &LazyTabItem{Title: title, Builder: builder}
}

func (l *LazyTabItem) Content() fyne.CanvasObject {
	l.once.Do(func() {
		if l.Builder != nil {
			l.content = l.Builder()
		} else {
			l.content = container.NewWithoutLayout()
		}
	})
	return l.content
}

func (l *LazyTabItem) CurrentContent() fyne.CanvasObject {
	return l.content
}

// ApplyToTabs wires lazy initialization to tab selection changes.
func ApplyToTabs(tabs *container.AppTabs, lazyItems map[string]*LazyTabItem) {
	if tabs == nil {
		return
	}
	prev := tabs.OnSelected
	tabs.OnSelected = func(item *container.TabItem) {
		if item != nil {
			if l, ok := lazyItems[item.Text]; ok {
				item.Content = l.Content()
			}
		}
		if prev != nil {
			prev(item)
		}
	}
}
