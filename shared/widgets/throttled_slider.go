//go:build legacy_fyne

package widgets

import (
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
)

// ThrottledSlider provides lightweight preview updates during drag and a full commit on release.
type ThrottledSlider struct {
	Slider          *widget.Slider
	previewInterval time.Duration
	onPreview       func(value float64)
	onCommit        func(value float64)

	mu          sync.Mutex
	lastPreview time.Time
	latestValue float64
}

// NewThrottledSlider creates a slider with throttled preview callback and full commit callback.
func NewThrottledSlider(min, max float64, previewInterval time.Duration, onPreview, onCommit func(value float64)) *ThrottledSlider {
	if previewInterval <= 0 {
		previewInterval = 40 * time.Millisecond
	}
	ts := &ThrottledSlider{
		Slider:          widget.NewSlider(min, max),
		previewInterval: previewInterval,
		onPreview:       onPreview,
		onCommit:        onCommit,
	}
	ts.Slider.OnChanged = ts.handleChanged
	ts.Slider.OnChangeEnded = ts.handleChangeEnded
	return ts
}

func (t *ThrottledSlider) handleChanged(value float64) {
	t.mu.Lock()
	now := time.Now()
	t.latestValue = value
	shouldPreview := now.Sub(t.lastPreview) >= t.previewInterval
	if shouldPreview {
		t.lastPreview = now
	}
	onPreview := t.onPreview
	t.mu.Unlock()

	if shouldPreview && onPreview != nil {
		onPreview(value)
	}
}

func (t *ThrottledSlider) handleChangeEnded(value float64) {
	t.mu.Lock()
	t.latestValue = value
	t.lastPreview = time.Time{}
	onCommit := t.onCommit
	t.mu.Unlock()

	if onCommit != nil {
		onCommit(value)
	}
}
