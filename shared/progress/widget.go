// Package progress provides Fyne widgets for progress display.
package progress

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ProgressWidget is a Fyne widget that displays progress with ETA and cancellation.
type ProgressWidget struct {
	widget.BaseWidget

	progress    *Progress
	showCancel  bool
	showDetails bool
	compact     bool

	// UI elements
	titleLabel   *widget.Label
	phaseLabel   *widget.Label
	detailLabel  *widget.Label
	progressBar  *widget.ProgressBar
	etaLabel     *widget.Label
	rateLabel    *widget.Label
	cancelBtn    *widget.Button
	pauseBtn     *widget.Button
	statusIcon   *canvas.Text
	container    *fyne.Container

	// Animation for indeterminate progress
	indeterminateAnim *fyne.Animation
	indeterminatePos  float32
}

// ProgressWidgetOption configures a ProgressWidget
type ProgressWidgetOption func(*ProgressWidget)

// WithCancel enables the cancel button
func WithCancel(enabled bool) ProgressWidgetOption {
	return func(w *ProgressWidget) {
		w.showCancel = enabled
	}
}

// WithDetails enables detailed status messages
func WithDetails(enabled bool) ProgressWidgetOption {
	return func(w *ProgressWidget) {
		w.showDetails = enabled
	}
}

// WithCompact enables compact mode (smaller layout)
func WithCompact(compact bool) ProgressWidgetOption {
	return func(w *ProgressWidget) {
		w.compact = compact
	}
}

// NewProgressWidget creates a new progress widget
func NewProgressWidget(p *Progress, opts ...ProgressWidgetOption) *ProgressWidget {
	w := &ProgressWidget{
		progress:    p,
		showCancel:  true,
		showDetails: true,
		compact:     false,
	}

	for _, opt := range opts {
		opt(w)
	}

	w.ExtendBaseWidget(w)
	w.buildUI()

	// Register for progress updates
	p.OnProgress(func(p *Progress) {
		fyne.Do(func() {
			w.Refresh()
		})
	})

	return w
}

// buildUI constructs the widget UI
func (w *ProgressWidget) buildUI() {
	// Status icon
	w.statusIcon = canvas.NewText("⏳", theme.ForegroundColor())
	w.statusIcon.TextSize = 16
	w.statusIcon.TextStyle.Bold = true

	// Title label
	w.titleLabel = widget.NewLabelWithStyle(
		w.progress.Operation(),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	// Phase label
	w.phaseLabel = widget.NewLabel("")
	w.phaseLabel.TextStyle.Italic = true

	// Progress bar
	w.progressBar = widget.NewProgressBar()
	w.progressBar.Min = 0
	w.progressBar.Max = 1

	// Detail label
	w.detailLabel = widget.NewLabel("")
	w.detailLabel.Wrapping = fyne.TextTruncate

	// ETA label
	w.etaLabel = widget.NewLabel("")
	w.etaLabel.TextStyle = fyne.TextStyle{}

	// Rate label
	w.rateLabel = widget.NewLabel("")
	w.rateLabel.TextStyle = fyne.TextStyle{}

	// Cancel button
	w.cancelBtn = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		w.progress.Cancel()
	})

	// Pause button
	w.pauseBtn = widget.NewButtonWithIcon("", theme.MediaPauseIcon(), func() {
		if w.progress.State() == StatePaused {
			w.progress.Resume()
		} else {
			w.progress.Pause()
		}
	})

	// Build layout based on compact mode
	w.container = w.buildLayout()
}

// buildLayout constructs the appropriate layout
func (w *ProgressWidget) buildLayout() *fyne.Container {
	if w.compact {
		return w.buildCompactLayout()
	}
	return w.buildFullLayout()
}

// buildCompactLayout creates a minimal progress display
func (w *ProgressWidget) buildCompactLayout() *fyne.Container {
	topRow := container.NewHBox(
		w.statusIcon,
		w.titleLabel,
		w.phaseLabel,
	)

	if w.showCancel {
		topRow.Add(container.NewHBox(
			w.pauseBtn,
			w.cancelBtn,
		))
	}

	return container.NewVBox(
		topRow,
		w.progressBar,
	)
}

// buildFullLayout creates a detailed progress display
func (w *ProgressWidget) buildFullLayout() *fyne.Container {
	// Title row with status icon and operation name
	titleRow := container.NewHBox(
		w.statusIcon,
		w.titleLabel,
	)

	// Control buttons
	var controls *fyne.Container
	if w.showCancel {
		controls = container.NewHBox(
			w.pauseBtn,
			w.cancelBtn,
		)
	}

	// Progress info row (ETA and rate)
	infoRow := container.NewHBox(
		w.etaLabel,
		widget.NewSeparator(),
		w.rateLabel,
	)

	// Build main layout
	var content *fyne.Container
	if w.showDetails {
		content = container.NewVBox(
			container.NewBorder(nil, nil, titleRow, controls),
			w.phaseLabel,
			w.progressBar,
			infoRow,
			w.detailLabel,
		)
	} else {
		content = container.NewVBox(
			container.NewBorder(nil, nil, titleRow, controls),
			w.phaseLabel,
			w.progressBar,
			infoRow,
		)
	}

	return content
}

// Refresh updates the widget display
func (w *ProgressWidget) Refresh() {
	info := w.progress.Info()

	// Update status icon
	switch info.State {
	case StateRunning:
		w.statusIcon.Text = "⏳"
		w.statusIcon.Color = theme.PrimaryColor()
	case StatePaused:
		w.statusIcon.Text = "⏸️"
		w.statusIcon.Color = theme.WarningColor()
	case StateCancelled:
		w.statusIcon.Text = "❌"
		w.statusIcon.Color = theme.ErrorColor()
	case StateCompleted:
		w.statusIcon.Text = "✅"
		w.statusIcon.Color = successColor()
	case StateFailed:
		w.statusIcon.Text = "❗"
		w.statusIcon.Color = theme.ErrorColor()
	default:
		w.statusIcon.Text = "⏹️"
		w.statusIcon.Color = theme.ForegroundColor()
	}
	w.statusIcon.Refresh()

	// Update labels
	w.titleLabel.SetText(info.Operation)
	w.phaseLabel.SetText(info.Phase)

	if w.showDetails {
		w.detailLabel.SetText(info.Detail)
	}

	// Update progress bar
	if info.Total > 0 {
		w.progressBar.SetValue(float64(info.Current) / float64(info.Total))
	} else {
		// Indeterminate progress - animate
		w.startIndeterminateAnimation()
	}

	// Update ETA
	if info.State == StateRunning && info.ETA > 0 {
		w.etaLabel.SetText(fmt.Sprintf("ETA: %s", formatDuration(info.ETA)))
	} else if info.State == StateCompleted {
		w.etaLabel.SetText(fmt.Sprintf("Completed in %s", formatDuration(info.Elapsed)))
	} else if info.State == StateCancelled {
		w.etaLabel.SetText("Cancelled")
	} else if info.State == StateFailed {
		w.etaLabel.SetText("Failed")
	} else {
		w.etaLabel.SetText("")
	}

	// Update rate
	if info.State == StateRunning && info.Rate > 0 {
		w.rateLabel.SetText(formatRate(info.Rate, info.Total))
	} else {
		w.rateLabel.SetText("")
	}

	// Update pause button
	if info.State == StatePaused {
		w.pauseBtn.SetIcon(theme.MediaPlayIcon())
	} else {
		w.pauseBtn.SetIcon(theme.MediaPauseIcon())
	}

	// Enable/disable buttons based on state
	canCancel := info.State == StateRunning || info.State == StatePaused
	if canCancel {
		w.cancelBtn.Enable()
		w.pauseBtn.Enable()
	} else {
		w.cancelBtn.Disable()
		w.pauseBtn.Disable()
	}

	w.BaseWidget.Refresh()
}

// startIndeterminateAnimation starts the indeterminate progress animation
func (w *ProgressWidget) startIndeterminateAnimation() {
	if w.indeterminateAnim != nil {
		return // Already running
	}

	w.indeterminateAnim = canvas.NewPositionAnimation(
		fyne.NewPos(0, 0),
		fyne.NewPos(1, 0),
		time.Second*2,
		func(pos fyne.Position) {
			w.indeterminatePos = pos.X
			// The progress bar doesn't support native indeterminate,
			// so we'll pulse the value
			pulse := (pos.X * 0.3) + 0.1
			w.progressBar.SetValue(float64(pulse))
		},
	)
	w.indeterminateAnim.RepeatCount = fyne.AnimationRepeatForever
	w.indeterminateAnim.AutoReverse = true
	w.indeterminateAnim.Start()
}

// stopIndeterminateAnimation stops the animation
func (w *ProgressWidget) stopIndeterminateAnimation() {
	if w.indeterminateAnim != nil {
		w.indeterminateAnim.Stop()
		w.indeterminateAnim = nil
	}
}

// CreateRenderer implements fyne.Widget
func (w *ProgressWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

// formatRate formats the processing rate for display
func formatRate(rate float64, total int64) string {
	if rate < 1 {
		return fmt.Sprintf("%.2f /sec", rate)
	}
	if rate < 1000 {
		return fmt.Sprintf("%.1f /sec", rate)
	}
	return fmt.Sprintf("%.1fk /sec", rate/1000)
}

// successColor returns a success color
func successColor() color.Color {
	return color.RGBA{R: 50, G: 200, B: 100, A: 255}
}

// ProgressDialog shows a progress dialog that blocks until complete or cancelled
type ProgressDialog struct {
	progress *Progress
	widget   *ProgressWidget
	dialog   *widget.PopUp
	window   fyne.Window
	onDone   func(cancelled bool)
}

// NewProgressDialog creates a modal progress dialog
func NewProgressDialog(window fyne.Window, p *Progress, opts ...ProgressWidgetOption) *ProgressDialog {
	pd := &ProgressDialog{
		progress: p,
		widget:   NewProgressWidget(p, opts...),
		window:   window,
	}

	// Create dialog content with padding
	content := container.NewPadded(
		container.NewVBox(
			pd.widget,
		),
	)

	pd.dialog = widget.NewModalPopUp(content, window.Canvas())

	// Register completion callback
	p.OnComplete(func(p *Progress) {
		fyne.Do(func() {
			pd.Hide()
			if pd.onDone != nil {
				pd.onDone(false)
			}
		})
	})

	p.OnError(func(p *Progress, err error) {
		fyne.Do(func() {
			pd.Hide()
			if pd.onDone != nil {
				pd.onDone(true)
			}
		})
	})

	return pd
}

// Show displays the progress dialog
func (pd *ProgressDialog) Show() {
	pd.dialog.Show()
	pd.dialog.Resize(fyne.NewSize(400, 150))
}

// Hide closes the progress dialog
func (pd *ProgressDialog) Hide() {
	pd.dialog.Hide()
}

// OnDone registers a callback for when the dialog closes
func (pd *ProgressDialog) OnDone(fn func(cancelled bool)) {
	pd.onDone = fn
}

// Progress returns the underlying progress tracker
func (pd *ProgressDialog) Progress() *Progress {
	return pd.progress
}

// ProgressListWidget displays multiple progress items
type ProgressListWidget struct {
	widget.BaseWidget

	items     []*Progress
	widgets   []*ProgressWidget
	container *fyne.Container
}

// NewProgressListWidget creates a widget that shows multiple progress items
func NewProgressListWidget() *ProgressListWidget {
	w := &ProgressListWidget{
		container: container.NewVBox(),
	}
	w.ExtendBaseWidget(w)
	return w
}

// AddProgress adds a progress item to the list
func (w *ProgressListWidget) AddProgress(p *Progress, opts ...ProgressWidgetOption) {
	pw := NewProgressWidget(p, opts...)
	w.items = append(w.items, p)
	w.widgets = append(w.widgets, pw)
	w.container.Add(pw)
	w.container.Add(widget.NewSeparator())
	w.Refresh()
}

// RemoveProgress removes a progress item from the list
func (w *ProgressListWidget) RemoveProgress(p *Progress) {
	for i, item := range w.items {
		if item == p {
			w.items = append(w.items[:i], w.items[i+1:]...)
			w.widgets = append(w.widgets[:i], w.widgets[i+1:]...)
			w.rebuildContainer()
			break
		}
	}
}

// Clear removes all progress items
func (w *ProgressListWidget) Clear() {
	w.items = nil
	w.widgets = nil
	w.container.RemoveAll()
	w.Refresh()
}

// rebuildContainer reconstructs the container
func (w *ProgressListWidget) rebuildContainer() {
	w.container.RemoveAll()
	for _, pw := range w.widgets {
		w.container.Add(pw)
		w.container.Add(widget.NewSeparator())
	}
	w.Refresh()
}

// CreateRenderer implements fyne.Widget
func (w *ProgressListWidget) CreateRenderer() fyne.WidgetRenderer {
	scroll := container.NewVScroll(w.container)
	return widget.NewSimpleRenderer(scroll)
}

// MiniBadge creates a small progress indicator for status bars
type MiniBadge struct {
	widget.BaseWidget

	progress  *Progress
	circle    *canvas.Circle
	label     *canvas.Text
	container *fyne.Container
}

// NewMiniBadge creates a compact progress badge
func NewMiniBadge(p *Progress) *MiniBadge {
	b := &MiniBadge{
		progress: p,
		circle:   canvas.NewCircle(theme.PrimaryColor()),
		label:    canvas.NewText("", theme.ForegroundColor()),
	}

	b.circle.Resize(fyne.NewSize(16, 16))
	b.label.TextSize = 10

	b.container = container.NewHBox(b.circle, b.label)
	b.ExtendBaseWidget(b)

	p.OnProgress(func(p *Progress) {
		fyne.Do(func() {
			b.Refresh()
		})
	})

	return b
}

// Refresh updates the badge
func (b *MiniBadge) Refresh() {
	info := b.progress.Info()

	// Update circle color based on state
	switch info.State {
	case StateRunning:
		b.circle.FillColor = theme.PrimaryColor()
	case StatePaused:
		b.circle.FillColor = theme.WarningColor()
	case StateCompleted:
		b.circle.FillColor = successColor()
	case StateFailed, StateCancelled:
		b.circle.FillColor = theme.ErrorColor()
	default:
		b.circle.FillColor = theme.DisabledColor()
	}

	// Update label
	if info.Total > 0 {
		b.label.Text = fmt.Sprintf("%.0f%%", info.Percent)
	} else {
		b.label.Text = info.Phase
	}

	b.circle.Refresh()
	b.label.Refresh()
	b.BaseWidget.Refresh()
}

// CreateRenderer implements fyne.Widget
func (b *MiniBadge) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.container)
}
