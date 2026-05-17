//go:build legacy_fyne

// Package widgets provides shared widget utilities for Fyne GUI development.
// This file implements a reusable KeyStat widget for displaying
// prominent statistics in demo applications.
package widgets

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// KeyStat is a reusable widget for displaying a key statistic prominently.
// It shows a label and a large value, suitable for dashboards and demos.
type KeyStat struct {
	widget.BaseWidget

	mu          sync.RWMutex
	label       string
	value       string
	minSize     fyne.Size
	labelWidget *widget.Label
	valueWidget *widget.Label
}

// KeyStatConfig holds configuration for creating a KeyStat.
type KeyStatConfig struct {
	Label   string
	Value   string
	MinSize fyne.Size
}

// NewKeyStat creates a new key stat widget.
func NewKeyStat(config KeyStatConfig) *KeyStat {
	if config.MinSize.Width <= 0 {
		config.MinSize.Width = 100
	}
	if config.MinSize.Height <= 0 {
		config.MinSize.Height = 50
	}
	if config.Label == "" {
		config.Label = "Stat"
	}
	if config.Value == "" {
		config.Value = "-"
	}

	k := &KeyStat{
		label:   config.Label,
		value:   config.Value,
		minSize: config.MinSize,
	}
	k.ExtendBaseWidget(k)
	return k
}

// SetValue updates the displayed value.
func (k *KeyStat) SetValue(value string) {
	k.mu.Lock()
	k.value = value
	k.mu.Unlock()
	if k.valueWidget != nil {
		fyne.Do(func() {
			k.valueWidget.SetText(value)
		})
	}
}

// GetValue returns the current value.
func (k *KeyStat) GetValue() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.value
}

// SetLabel updates the label.
func (k *KeyStat) SetLabel(label string) {
	k.mu.Lock()
	k.label = label
	k.mu.Unlock()
	if k.labelWidget != nil {
		fyne.Do(func() {
			k.labelWidget.SetText(label)
		})
	}
}

// SetLabelAndValue updates both the label and value.
func (k *KeyStat) SetLabelAndValue(label, value string) {
	k.mu.Lock()
	k.label = label
	k.value = value
	k.mu.Unlock()
	fyne.Do(func() {
		if k.labelWidget != nil {
			k.labelWidget.SetText(label)
		}
		if k.valueWidget != nil {
			k.valueWidget.SetText(value)
		}
	})
}

// MinSize returns the minimum size for the widget.
func (k *KeyStat) MinSize() fyne.Size {
	return k.minSize
}

// CreateRenderer implements fyne.Widget.
func (k *KeyStat) CreateRenderer() fyne.WidgetRenderer {
	k.mu.RLock()
	label := k.label
	value := k.value
	k.mu.RUnlock()

	labelWidget := widget.NewLabel(label)
	labelWidget.Alignment = fyne.TextAlignCenter

	valueWidget := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	box := container.NewVBox(labelWidget, valueWidget)

	k.mu.Lock()
	k.labelWidget = labelWidget
	k.valueWidget = valueWidget
	k.mu.Unlock()

	return widget.NewSimpleRenderer(box)
}

// KeyStatGroup manages a group of related KeyStat widgets.
type KeyStatGroup struct {
	stats map[string]*KeyStat
	mu    sync.RWMutex
}

// NewKeyStatGroup creates a new group of key stats.
func NewKeyStatGroup() *KeyStatGroup {
	return &KeyStatGroup{
		stats: make(map[string]*KeyStat),
	}
}

// Add adds a new stat to the group.
func (g *KeyStatGroup) Add(id string, config KeyStatConfig) *KeyStat {
	stat := NewKeyStat(config)
	g.mu.Lock()
	g.stats[id] = stat
	g.mu.Unlock()
	return stat
}

// Get returns a stat by ID.
func (g *KeyStatGroup) Get(id string) *KeyStat {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.stats[id]
}

// SetValue updates a stat's value by ID.
func (g *KeyStatGroup) SetValue(id, value string) {
	g.mu.RLock()
	stat := g.stats[id]
	g.mu.RUnlock()
	if stat != nil {
		stat.SetValue(value)
	}
}

// All returns all stats in the group as a slice.
func (g *KeyStatGroup) All() []*KeyStat {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]*KeyStat, 0, len(g.stats))
	for _, stat := range g.stats {
		result = append(result, stat)
	}
	return result
}

// AsContainer returns all stats in a horizontal container.
func (g *KeyStatGroup) AsContainer() *fyne.Container {
	g.mu.RLock()
	defer g.mu.RUnlock()
	objects := make([]fyne.CanvasObject, 0, len(g.stats))
	for _, stat := range g.stats {
		objects = append(objects, stat)
	}
	return container.NewHBox(objects...)
}
