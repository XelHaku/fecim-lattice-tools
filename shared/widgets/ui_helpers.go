//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SafeUpdateLabel updates a label's text safely from any goroutine.
// Skips the update if label is nil.
func SafeUpdateLabel(label *widget.Label, text string) {
	if label == nil {
		return
	}
	safeUIUpdate(func() {
		label.SetText(text)
	})
}

// SafeUpdateProgress updates a progress bar's value safely from any goroutine.
// Skips the update if progress is nil.
func SafeUpdateProgress(progress *widget.ProgressBar, value float64) {
	if progress == nil {
		return
	}
	safeUIUpdate(func() {
		progress.SetValue(value)
	})
}

// SafeUpdateProgressInfinite sets a progress bar to infinite mode safely.
// Skips the update if progress is nil.
func SafeUpdateProgressInfinite(progress *widget.ProgressBarInfinite, start bool) {
	if progress == nil {
		return
	}
	safeUIUpdate(func() {
		if start {
			progress.Start()
		} else {
			progress.Stop()
		}
	})
}

// SafeRefresh refreshes a canvas object safely from any goroutine.
// Skips the refresh if obj is nil.
func SafeRefresh(obj fyne.CanvasObject) {
	if obj == nil {
		return
	}
	safeUIUpdate(func() {
		obj.Refresh()
	})
}

// SafeShow shows a canvas object safely from any goroutine.
// Skips if obj is nil.
func SafeShow(obj fyne.CanvasObject) {
	if obj == nil {
		return
	}
	safeUIUpdate(func() {
		obj.Show()
	})
}

// SafeHide hides a canvas object safely from any goroutine.
// Skips if obj is nil.
func SafeHide(obj fyne.CanvasObject) {
	if obj == nil {
		return
	}
	safeUIUpdate(func() {
		obj.Hide()
	})
}

// SafeShowHide shows or hides a canvas object safely from any goroutine.
// Skips if obj is nil.
func SafeShowHide(obj fyne.CanvasObject, show bool) {
	if obj == nil {
		return
	}
	safeUIUpdate(func() {
		if show {
			obj.Show()
		} else {
			obj.Hide()
		}
	})
}

// SafeEnable enables a disableable widget safely from any goroutine.
// Skips if w is nil.
func SafeEnable(w fyne.Disableable) {
	if w == nil {
		return
	}
	safeUIUpdate(func() {
		w.Enable()
	})
}

// SafeDisable disables a disableable widget safely from any goroutine.
// Skips if w is nil.
func SafeDisable(w fyne.Disableable) {
	if w == nil {
		return
	}
	safeUIUpdate(func() {
		w.Disable()
	})
}

// SafeEnableDisable enables or disables a widget safely from any goroutine.
// Skips if w is nil.
func SafeEnableDisable(w fyne.Disableable, enable bool) {
	if w == nil {
		return
	}
	safeUIUpdate(func() {
		if enable {
			w.Enable()
		} else {
			w.Disable()
		}
	})
}

// SafeSetEntry sets an entry's text safely from any goroutine.
// Skips if entry is nil.
func SafeSetEntry(entry *widget.Entry, text string) {
	if entry == nil {
		return
	}
	safeUIUpdate(func() {
		entry.SetText(text)
	})
}

// SafeSetCheck sets a check widget's checked state safely from any goroutine.
// Skips if check is nil.
func SafeSetCheck(check *widget.Check, checked bool) {
	if check == nil {
		return
	}
	safeUIUpdate(func() {
		check.SetChecked(checked)
	})
}

// SafeSetSlider sets a slider's value safely from any goroutine.
// Skips if slider is nil.
func SafeSetSlider(slider *widget.Slider, value float64) {
	if slider == nil {
		return
	}
	safeUIUpdate(func() {
		slider.SetValue(value)
	})
}

// SafeSetSelect sets a select widget's selected value safely from any goroutine.
// Skips if sel is nil.
func SafeSetSelect(sel *widget.Select, selected string) {
	if sel == nil {
		return
	}
	safeUIUpdate(func() {
		sel.SetSelected(selected)
	})
}

// SafeSetSelectIndex sets a select widget's selected index safely from any goroutine.
// Skips if sel is nil.
func SafeSetSelectIndex(sel *widget.Select, index int) {
	if sel == nil {
		return
	}
	safeUIUpdate(func() {
		sel.SetSelectedIndex(index)
	})
}
