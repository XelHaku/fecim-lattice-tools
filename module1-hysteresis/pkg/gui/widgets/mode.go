package widgets

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ModeIndicator shows the current mode (WRITE/READ) with colored background
type ModeIndicator struct {
	widget.BaseWidget

	mu      sync.RWMutex
	isWrite bool
	minSize fyne.Size
}

// NewModeIndicator creates a new mode indicator
func NewModeIndicator() *ModeIndicator {
	m := &ModeIndicator{
		isWrite: false,
		minSize: fyne.NewSize(140, 50),
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *ModeIndicator) SetMinSize(size fyne.Size) {
	m.minSize = size
}

func (m *ModeIndicator) MinSize() fyne.Size {
	return m.minSize
}

func (m *ModeIndicator) SetWrite(isWrite bool) {
	m.mu.Lock()
	m.isWrite = isWrite
	m.mu.Unlock()
}

func (m *ModeIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &modeRenderer{indicator: m}
}

type modeRenderer struct {
	indicator *ModeIndicator
	objects   []fyne.CanvasObject
	cache     sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *modeRenderer) MinSize() fyne.Size {
	return r.indicator.minSize
}

func (r *modeRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("modeRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *modeRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("modeRenderer", r.indicator.Size())
	size := r.indicator.Size()
	// Always re-layout on Refresh for this dynamic widget (mode changes)
	r.layoutWithSize(size)
}

func (r *modeRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.indicator.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.indicator.mu.RLock()
	isWrite := r.indicator.isWrite
	r.indicator.mu.RUnlock()

	r.objects = r.objects[:0]

	// Constrain to minimum size to prevent growing
	minSize := r.indicator.minSize
	if size.Width > minSize.Width {
		size.Width = minSize.Width
	}
	if size.Height > minSize.Height {
		size.Height = minSize.Height
	}

	// Box colors
	var bgColor, borderColor color.RGBA
	var modeText, conditionText string

	if isWrite {
		bgColor = color.RGBA{180, 50, 50, 255} // Red background
		borderColor = color.RGBA{255, 100, 100, 255}
		modeText = "WRITE"
		conditionText = "|E| > Ec"
	} else {
		bgColor = color.RGBA{50, 150, 80, 255} // Green background
		borderColor = color.RGBA{100, 220, 130, 255}
		modeText = "READ"
		conditionText = "|E| < Ec"
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background with padding
	padding := float32(3)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Mode text (centered) - scale with size
	modeTextSize := size.Height * 0.4
	if modeTextSize < 14 {
		modeTextSize = 14
	}
	if modeTextSize > 24 {
		modeTextSize = 24
	}
	modeLabel := canvas.NewText(modeText, color.White)
	modeLabel.TextSize = modeTextSize
	modeLabel.TextStyle = fyne.TextStyle{Bold: true}
	modeTextW := float32(len(modeText)) * modeTextSize * 0.6
	modeLabel.Move(fyne.NewPos((size.Width-modeTextW)/2, size.Height*0.15))
	r.objects = append(r.objects, modeLabel)

	// Condition text (centered, bottom) - scale with size
	condTextSize := size.Height * 0.25
	if condTextSize < 14 {
		condTextSize = 14
	}
	if condTextSize > 14 {
		condTextSize = 14
	}
	condLabel := canvas.NewText(conditionText, color.RGBA{230, 230, 230, 255})
	condLabel.TextSize = condTextSize
	condLabelW := float32(len(conditionText)) * condTextSize * 0.6
	condLabel.Move(fyne.NewPos((size.Width-condLabelW)/2, size.Height*0.6))
	r.objects = append(r.objects, condLabel)

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *modeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *modeRenderer) Destroy() {}
