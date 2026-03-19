// Package gui provides Fyne-based GUI components for crossbar visualization.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/mathutil"
	gpurender "fecim-lattice-tools/shared/render"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// refreshMinInterval is the minimum time between heatmap refreshes (30 FPS max)
const refreshMinInterval = 33 * time.Millisecond

// gpuCellThreshold is the minimum number of cells (rows*cols) before the GPU
// renderer is engaged. Below this threshold the software path is fast enough.
const gpuCellThreshold = 4096 // 64x64

// Compile-time interface checks
var _ desktop.Hoverable = (*CrossbarHeatmap)(nil)

// CrossbarHeatmap is a custom widget for visualizing crossbar array states.
type CrossbarHeatmap struct {
	widget.BaseWidget

	data       [][]float64
	rows, cols int
	minVal     float64
	maxVal     float64
	colormap   string

	// Selection
	selectedRow   int
	selectedCol   int
	showSelection bool

	// Grid lines (subtle lines between cells for larger arrays)
	showGridLines bool

	// Animation state
	animPhase     int     // 0=none, 1=input, 2=compute, 3=output
	animProgress  float64 // 0-1 progress within phase
	highlightCols []int   // Columns to highlight (input)
	highlightRows []int   // Rows to highlight (output)

	// Callbacks
	OnCellTapped func(row, col int)
	OnCellHover  func(row, col int, value float64)

	// Internal
	raster   *canvas.Raster
	cellSize float32

	// Rate limiting for refresh (max 30 FPS)
	refreshMu      sync.Mutex
	lastRefresh    time.Time
	refreshPending bool

	// Data synchronization
	dataMu sync.RWMutex

	// Fixed scale (disables auto-scaling)
	useFixedScale bool
	fixedMinVal   float64
	fixedMaxVal   float64

	// GPU rendering (L09). When non-nil the base heatmap image is produced
	// by the GPU renderer; selection/animation overlays are still applied in
	// software. Created lazily on first call to EnableGPURendering().
	gpuRenderer *gpurender.GPUHeatmapRenderer
	useGPU      bool // true when GPU rendering is enabled AND available
}

// NewCrossbarHeatmap creates a new crossbar heatmap widget.
func NewCrossbarHeatmap(rows, cols int) *CrossbarHeatmap {
	h := &CrossbarHeatmap{
		rows:        rows,
		cols:        cols,
		minVal:      0,
		maxVal:      1,
		colormap:    "viridis",
		selectedRow: -1,
		selectedCol: -1,
		cellSize:    6, // Smaller cell size to fit better
	}

	// Initialize data
	h.data = make([][]float64, rows)
	for i := range h.data {
		h.data[i] = make([]float64, cols)
	}

	// Try to initialise GPU renderer. If Vulkan is available, GPU rendering
	// will be used automatically for arrays above gpuCellThreshold cells.
	h.gpuRenderer = gpurender.NewGPUHeatmapRenderer()
	h.useGPU = h.gpuRenderer.Available()

	h.ExtendBaseWidget(h)
	return h
}

// rateLimitedRefresh performs a refresh with rate limiting to prevent UI overload.
// Maximum refresh rate is 30 FPS (refreshMinInterval between calls).
// Also suppresses refreshes during startup stabilization period.
func (h *CrossbarHeatmap) rateLimitedRefresh() {
	// Skip refreshes during startup stabilization to prevent resize oscillation
	if sharedwidgets.IsStartupStabilizing() {
		return
	}

	h.refreshMu.Lock()

	// Check if we can refresh immediately
	now := time.Now()
	elapsed := now.Sub(h.lastRefresh)

	if elapsed >= refreshMinInterval {
		// Enough time has passed, refresh immediately
		h.lastRefresh = now
		h.refreshPending = false
		h.refreshMu.Unlock()
		// Use fyne.Do for thread safety in case called from background goroutine
		fyne.Do(func() {
			h.BaseWidget.Refresh()
		})
		return
	}

	// Too soon - schedule a delayed refresh if not already pending
	if h.refreshPending {
		h.refreshMu.Unlock()
		return
	}

	h.refreshPending = true
	delay := refreshMinInterval - elapsed
	h.refreshMu.Unlock()

	// Schedule delayed refresh
	go func() {
		time.Sleep(delay)
		h.refreshMu.Lock()
		h.refreshPending = false
		h.lastRefresh = time.Now()
		h.refreshMu.Unlock()

		fyne.Do(func() {
			h.BaseWidget.Refresh() // Call actual Fyne refresh
		})
	}()
}

// SetData updates the heatmap data.
func (h *CrossbarHeatmap) SetData(data [][]float64) {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	// Copy data
	for i := 0; i < h.rows && i < len(data); i++ {
		for j := 0; j < h.cols && j < len(data[i]); j++ {
			h.data[i][j] = data[i][j]
		}
	}

	// Only auto-scale if not using fixed scale
	if !h.useFixedScale {
		h.minVal = math.Inf(1)
		h.maxVal = math.Inf(-1)

		for i := 0; i < h.rows && i < len(data); i++ {
			for j := 0; j < h.cols && j < len(data[i]); j++ {
				if data[i][j] < h.minVal {
					h.minVal = data[i][j]
				}
				if data[i][j] > h.maxVal {
					h.maxVal = data[i][j]
				}
			}
		}

		if h.maxVal <= h.minVal {
			h.maxVal = h.minVal + 1
		}
	}

	h.rateLimitedRefresh()
}

// SetFixedScale sets a fixed min/max scale for the heatmap.
// When enabled, the heatmap will NOT auto-scale to the data range.
// Use this for consistent comparison between different data sets.
func (h *CrossbarHeatmap) SetFixedScale(min, max float64) {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	h.useFixedScale = true
	h.fixedMinVal = min
	h.fixedMaxVal = max
	h.minVal = min
	h.maxVal = max
	h.rateLimitedRefresh()
}

// ClearFixedScale disables fixed scale and reverts to auto-scaling.
func (h *CrossbarHeatmap) ClearFixedScale() {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	h.useFixedScale = false
	h.rateLimitedRefresh()
}

// SetGPUEnabled enables or disables GPU-accelerated heatmap rendering.
// When enabled and Vulkan is available, heatmaps with more than
// gpuCellThreshold cells are rendered on the GPU. The viridis colourmap is
// used for GPU rendering regardless of the selected software colourmap.
// Selection highlights, animation overlays, and grid lines are still
// composited in software on top of the GPU-rendered base image.
func (h *CrossbarHeatmap) SetGPUEnabled(enabled bool) {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	if enabled {
		if h.gpuRenderer == nil {
			h.gpuRenderer = gpurender.NewGPUHeatmapRenderer()
		}
		h.useGPU = h.gpuRenderer.Available()
	} else {
		h.useGPU = false
	}
}

// IsGPURenderingActive reports whether GPU rendering is currently in use.
func (h *CrossbarHeatmap) IsGPURenderingActive() bool {
	h.dataMu.RLock()
	defer h.dataMu.RUnlock()
	return h.useGPU
}

// DestroyGPURenderer releases GPU resources. Called when the widget is
// no longer needed or the application is shutting down.
func (h *CrossbarHeatmap) DestroyGPURenderer() {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()
	if h.gpuRenderer != nil {
		h.gpuRenderer.Destroy()
		h.gpuRenderer = nil
	}
	h.useGPU = false
}

// SetColormap changes the colormap.
// Available colormaps:
//   - viridis: Perceptually uniform, colorblind-safe (default for IR drop)
//   - plasma: Perceptually uniform, colorblind-safe (default for sneak paths)
//   - coolwarm: Diverging blue-white-red, generally colorblind-safe
//   - fecim: Custom FeCIM branding, not colorblind-safe
func (h *CrossbarHeatmap) SetColormap(name string) {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	h.colormap = name
	h.rateLimitedRefresh()
}

// SetSelection highlights a specific cell.
func (h *CrossbarHeatmap) SetSelection(row, col int) {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	h.selectedRow = row
	h.selectedCol = col
	h.showSelection = row >= 0 && col >= 0
	h.rateLimitedRefresh()
}

// ClearSelection removes cell selection highlight.
func (h *CrossbarHeatmap) ClearSelection() {
	h.showSelection = false
	h.selectedRow = -1
	h.selectedCol = -1
	h.rateLimitedRefresh()
}

// SetShowGridLines enables or disables grid line rendering.
func (h *CrossbarHeatmap) SetShowGridLines(show bool) {
	h.showGridLines = show
	h.rateLimitedRefresh()
}

// SetDimensions changes the dimensions of the heatmap and reinitializes data.
func (h *CrossbarHeatmap) SetDimensions(rows, cols int) {
	h.dataMu.Lock()
	defer h.dataMu.Unlock()

	h.rows = rows
	h.cols = cols
	h.selectedRow = -1
	h.selectedCol = -1
	h.showSelection = false

	// Only reset min/max if not using fixed scale
	if !h.useFixedScale {
		h.minVal = 0
		h.maxVal = 1 // Reset to default range
	}

	// Auto-enable grid lines for larger arrays (>16x16) for better cell distinction
	h.showGridLines = rows > 16 || cols > 16

	// Reinitialize data
	h.data = make([][]float64, rows)
	for i := range h.data {
		h.data[i] = make([]float64, cols)
	}

	h.rateLimitedRefresh()
}

// CreateRenderer implements fyne.Widget.
func (h *CrossbarHeatmap) CreateRenderer() fyne.WidgetRenderer {
	h.raster = canvas.NewRaster(h.generateImage)
	altLabel := widget.NewLabel(h.TextAlternative())
	altLabel.Wrapping = fyne.TextWrapWord
	altLabel.TextStyle = fyne.TextStyle{Italic: true}
	return &heatmapRenderer{
		heatmap:  h,
		raster:   h.raster,
		altLabel: altLabel,
	}
}

// TextAlternative returns a screen-reader-friendly textual summary of the heatmap.
func (h *CrossbarHeatmap) TextAlternative() string {
	h.dataMu.RLock()
	defer h.dataMu.RUnlock()

	selectedText := "none"
	if h.showSelection && h.selectedRow >= 0 && h.selectedRow < h.rows && h.selectedCol >= 0 && h.selectedCol < h.cols {
		selectedText = fmt.Sprintf("row %d, col %d = %.4g", h.selectedRow, h.selectedCol, h.data[h.selectedRow][h.selectedCol])
	}

	return fmt.Sprintf("Crossbar heatmap %dx%d. Range %.4g to %.4g. Colormap %s. Selected cell: %s.",
		h.rows, h.cols, h.minVal, h.maxVal, h.colormap, selectedText)
}

// MinSize returns the minimum size of the widget.
// Uses a small fixed minimum - actual size adapts to container.
func (h *CrossbarHeatmap) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

// Tapped handles tap events on the heatmap.
func (h *CrossbarHeatmap) Tapped(e *fyne.PointEvent) {
	// Calculate cell size dynamically based on current widget size
	size := h.Size()
	cellW := float64(size.Width-40) / float64(h.cols)
	cellH := float64(size.Height-40) / float64(h.rows)
	cellSize := math.Min(cellW, cellH)

	col := int((float64(e.Position.X) - 20) / cellSize)
	row := int((float64(e.Position.Y) - 20) / cellSize)

	if row >= 0 && row < h.rows && col >= 0 && col < h.cols {
		h.SetSelection(row, col)
		if h.OnCellTapped != nil {
			h.OnCellTapped(row, col)
		}
	}
}

// TappedSecondary handles secondary tap (right-click).
func (h *CrossbarHeatmap) TappedSecondary(*fyne.PointEvent) {
	h.ClearSelection()
}

// MouseMoved tracks mouse position for hover info.
func (h *CrossbarHeatmap) MouseMoved(e *desktop.MouseEvent) {
	h.dataMu.RLock()
	defer h.dataMu.RUnlock()

	size := h.Size()
	cellW := float64(size.Width-40) / float64(h.cols)
	cellH := float64(size.Height-40) / float64(h.rows)
	cellSize := math.Min(cellW, cellH)

	col := int((float64(e.Position.X) - 20) / cellSize)
	row := int((float64(e.Position.Y) - 20) / cellSize)

	if row >= 0 && row < h.rows && col >= 0 && col < h.cols {
		if h.OnCellHover != nil {
			h.OnCellHover(row, col, h.data[row][col])
		}
	}
}

// MouseIn is called when mouse enters the widget.
func (h *CrossbarHeatmap) MouseIn(*desktop.MouseEvent) {}

// MouseOut is called when mouse leaves the widget.
func (h *CrossbarHeatmap) MouseOut() {
	if h.OnCellHover != nil {
		h.OnCellHover(-1, -1, 0) // Signal mouse left
	}
}

// SetAnimPhase sets the current animation phase.
// Phase 0: No animation
// Phase 1: Input voltages being applied (highlight columns)
// Phase 2: Computing (wave animation through cells)
// Phase 3: Output currents (highlight rows)
func (h *CrossbarHeatmap) SetAnimPhase(phase int, progress float64) {
	h.animPhase = phase
	h.animProgress = progress
	h.rateLimitedRefresh()
}

// SetInputHighlight highlights specific columns (for input voltage visualization).
func (h *CrossbarHeatmap) SetInputHighlight(cols []int) {
	h.highlightCols = cols
	h.rateLimitedRefresh()
}

// SetOutputHighlight highlights specific rows (for output current visualization).
func (h *CrossbarHeatmap) SetOutputHighlight(rows []int) {
	h.highlightRows = rows
	h.rateLimitedRefresh()
}

// ClearAnimation clears all animation state.
func (h *CrossbarHeatmap) ClearAnimation() {
	h.animPhase = 0
	h.animProgress = 0
	h.highlightCols = nil
	h.highlightRows = nil
	h.rateLimitedRefresh()
}

// generateImage creates the heatmap image with optimized rendering.
func (h *CrossbarHeatmap) generateImage(w, h_size int) image.Image {
	h.dataMu.RLock()
	defer h.dataMu.RUnlock()

	img := image.NewRGBA(image.Rect(0, 0, w, h_size))

	// Calculate cell size
	cellW := float64(w-40) / float64(h.cols)
	cellH := float64(h_size-40) / float64(h.rows)
	cellSize := math.Min(cellW, cellH)

	// Fill background using draw.Draw (batch operation)
	bgColor := color.RGBA{30, 30, 40, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Pre-calculate wave position for phase 2 animation
	wavePos := 0
	if h.animPhase == 2 {
		wavePos = int(h.animProgress * float64(h.rows))
	}

	// Build highlight column lookup map for O(1) access
	highlightColMap := make(map[int]bool)
	for _, col := range h.highlightCols {
		highlightColMap[col] = true
	}

	// Build highlight row lookup map for O(1) access
	highlightRowMap := make(map[int]bool)
	for _, row := range h.highlightRows {
		highlightRowMap[row] = true
	}

	// GPU path: when GPU is available and the array is large enough, render
	// the full-resolution heatmap via Vulkan offscreen rendering. The GPU uses
	// the viridis colourmap. Selection, animation and grid overlays are still
	// composited in software on top of the GPU image.
	var gpuBaseImg *image.RGBA
	cellCount := h.rows * h.cols
	if h.useGPU && h.gpuRenderer != nil && cellCount >= gpuCellThreshold {
		// Flatten and normalise data to [0,1] for GPU.
		valRange := h.maxVal - h.minVal
		if valRange <= 0 {
			valRange = 1
		}
		flatValues := make([]float64, cellCount)
		for i := 0; i < h.rows; i++ {
			for j := 0; j < h.cols; j++ {
				flatValues[i*h.cols+j] = (h.data[i][j] - h.minVal) / valRange
			}
		}
		if gpuResult := h.gpuRenderer.RenderHeatmap(flatValues, h.rows, h.cols, w, h_size); gpuResult != nil {
			if rgba, ok := gpuResult.(*image.RGBA); ok {
				gpuBaseImg = rgba
			}
		}
	}

	// If GPU produced the base image, use it as the background (replaces
	// cell-by-cell drawing below). Overlays are still applied in software.
	if gpuBaseImg != nil {
		copy(img.Pix, gpuBaseImg.Pix)
	}

	// Draw cells using batch draw operations.
	// When GPU rendered the base image, skip per-cell colour fill (already in img)
	// but still apply selection highlights and animation overlays.
	for i := 0; i < h.rows; i++ {
		for j := 0; j < h.cols; j++ {
			// Calculate cell position
			x0 := int(20 + float64(j)*cellSize)
			y0 := int(20 + float64(i)*cellSize)
			x1 := int(20 + float64(j+1)*cellSize - 1)
			y1 := int(20 + float64(i+1)*cellSize - 1)

			// Software path: fill cell colour when GPU did not produce the image.
			if gpuBaseImg == nil {
				normVal := (h.data[i][j] - h.minVal) / (h.maxVal - h.minVal)
				cellColor := h.valueToColor(normVal)
				cellRect := image.Rect(x0, y0, x1, y1)
				draw.Draw(img, cellRect, &image.Uniform{cellColor}, image.Point{}, draw.Src)
			}

			// Draw selection highlight using draw.Draw for borders
			if h.showSelection && i == h.selectedRow && j == h.selectedCol {
				highlightColor := h.selectionColor()
				highlightUniform := &image.Uniform{highlightColor}
				// Top border (2px)
				draw.Draw(img, image.Rect(x0, y0, x1, y0+2), highlightUniform, image.Point{}, draw.Src)
				// Bottom border (2px)
				draw.Draw(img, image.Rect(x0, y1-2, x1, y1), highlightUniform, image.Point{}, draw.Src)
				// Left border (2px)
				draw.Draw(img, image.Rect(x0, y0, x0+2, y1), highlightUniform, image.Point{}, draw.Src)
				// Right border (2px)
				draw.Draw(img, image.Rect(x1-2, y0, x1, y1), highlightUniform, image.Point{}, draw.Src)
			}

			// Animation overlays - apply blending only when needed
			if h.animPhase > 0 {
				var overlay color.RGBA
				shouldOverlay := false

				// Phase 1: Input - highlight active columns with cyan
				if h.animPhase == 1 && highlightColMap[j] {
					overlay = color.RGBA{0, 255, 255, 100}
					shouldOverlay = true
				}

				// Phase 2: Compute - wave animation
				if h.animPhase == 2 && i <= wavePos {
					overlay = color.RGBA{255, 200, 0, 80}
					shouldOverlay = true
				}

				// Phase 3: Output - highlight active rows with orange
				if h.animPhase == 3 && highlightRowMap[i] {
					overlay = color.RGBA{255, 150, 0, 100}
					shouldOverlay = true
				}

				// Apply overlay using direct buffer access (faster than Set)
				if shouldOverlay {
					alpha := float64(overlay.A) / 255.0
					invAlpha := 1.0 - alpha
					stride := img.Stride
					for y := y0; y < y1; y++ {
						rowOffset := y * stride
						for x := x0; x < x1; x++ {
							pixOffset := rowOffset + x*4
							img.Pix[pixOffset] = uint8(float64(img.Pix[pixOffset])*invAlpha + float64(overlay.R)*alpha)
							img.Pix[pixOffset+1] = uint8(float64(img.Pix[pixOffset+1])*invAlpha + float64(overlay.G)*alpha)
							img.Pix[pixOffset+2] = uint8(float64(img.Pix[pixOffset+2])*invAlpha + float64(overlay.B)*alpha)
							// Alpha stays at 255
						}
					}
				}
			}
		}
	}

	// Draw grid lines if enabled (after cells for visibility)
	if h.showGridLines {
		gridColor := color.RGBA{60, 80, 100, 100} // Subtle blue-gray
		// Draw horizontal lines
		for i := 0; i <= h.rows; i++ {
			y := int(20 + float64(i)*cellSize)
			if y >= 0 && y < h_size {
				for x := 20; x < int(20+float64(h.cols)*cellSize) && x < w; x++ {
					img.Set(x, y, gridColor)
				}
			}
		}
		// Draw vertical lines
		for j := 0; j <= h.cols; j++ {
			x := int(20 + float64(j)*cellSize)
			if x >= 0 && x < w {
				for y := 20; y < int(20+float64(h.rows)*cellSize) && y < h_size; y++ {
					img.Set(x, y, gridColor)
				}
			}
		}
	}

	// Draw axis labels (simplified)
	labelColor := color.RGBA{200, 200, 200, 255}
	// Draw corner markers
	for i := 0; i < 15; i++ {
		img.Set(10+i, 10, labelColor)
		img.Set(10, 10+i, labelColor)
	}

	return img
}

// valueToColor converts a normalized value to a color using the selected colormap.
func (h *CrossbarHeatmap) valueToColor(t float64) color.RGBA {
	// For diverging colormap, t is already in [-1, 1] range
	if h.colormap == "diverging" {
		return divergingColor(t)
	}

	// Standard colormaps expect t in [0, 1]
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	switch h.colormap {
	case "viridis":
		return viridisColor(t)
	case "plasma":
		return plasmaColor(t)
	case "coolwarm":
		return coolwarmColor(t)
	case "fecim":
		return fecimColor(t)
	default:
		return viridisColor(t)
	}
}

// Colormap implementations
func viridisColor(t float64) color.RGBA {
	// Viridis approximation
	r := 0.267 + t*(0.993*t-0.068)
	g := 0.005 + t*(0.991-0.149*t)
	b := 0.329 + t*(0.288-0.147*t)

	return color.RGBA{
		R: uint8(mathutil.Clamp(r, 0, 1) * 255),
		G: uint8(mathutil.Clamp(g, 0, 1) * 255),
		B: uint8(mathutil.Clamp(b, 0, 1) * 255),
		A: 255,
	}
}

func plasmaColor(t float64) color.RGBA {
	r := 0.05 + t*0.89
	g := 0.03 + t*0.95*t
	b := 0.53 - t*0.40

	return color.RGBA{
		R: uint8(mathutil.Clamp(r, 0, 1) * 255),
		G: uint8(mathutil.Clamp(g, 0, 1) * 255),
		B: uint8(mathutil.Clamp(b, 0, 1) * 255),
		A: 255,
	}
}

func coolwarmColor(t float64) color.RGBA {
	if t < 0.5 {
		s := t * 2
		return color.RGBA{
			R: uint8(s * 255),
			G: uint8(s * 255),
			B: 255,
			A: 255,
		}
	}
	s := (t - 0.5) * 2
	return color.RGBA{
		R: 255,
		G: uint8((1 - s) * 255),
		B: uint8((1 - s) * 255),
		A: 255,
	}
}

// FeCIM custom colormap: purple (low) -> blue -> cyan -> green -> yellow -> red (high)
func fecimColor(t float64) color.RGBA {
	// 30-level inspired colormap matching FeCIM's discrete states
	if t < 0.2 {
		s := t * 5
		return color.RGBA{
			R: uint8(60 + s*20),
			G: uint8(s * 100),
			B: uint8(120 + s*80),
			A: 255,
		}
	} else if t < 0.4 {
		s := (t - 0.2) * 5
		return color.RGBA{
			R: uint8(80 - s*50),
			G: uint8(100 + s*155),
			B: uint8(200 - s*50),
			A: 255,
		}
	} else if t < 0.6 {
		s := (t - 0.4) * 5
		return color.RGBA{
			R: uint8(30 + s*180),
			G: uint8(255),
			B: uint8(150 - s*100),
			A: 255,
		}
	} else if t < 0.8 {
		s := (t - 0.6) * 5
		return color.RGBA{
			R: uint8(210 + s*45),
			G: uint8(255 - s*100),
			B: uint8(50 - s*50),
			A: 255,
		}
	}
	s := (t - 0.8) * 5
	return color.RGBA{
		R: 255,
		G: uint8(155 - s*155),
		B: 0,
		A: 255,
	}
}


// selectionColor returns a contrasting selection border color based on the current colormap.
func (h *CrossbarHeatmap) selectionColor() color.RGBA {
	switch h.colormap {
	case "viridis":
		// Viridis is blue-green-yellow, use white for contrast
		return color.RGBA{255, 255, 255, 255}
	case "plasma":
		// Plasma is purple-orange-yellow, use cyan for contrast
		return color.RGBA{0, 255, 255, 255}
	case "coolwarm":
		// Coolwarm is blue-white-red, use bright green for contrast
		return color.RGBA{0, 255, 0, 255}
	case "fecim":
		// FeCIM has many colors, use white for best visibility
		return color.RGBA{255, 255, 255, 255}
	case "diverging":
		// Diverging is blue-white-red, use bright green
		return color.RGBA{0, 255, 0, 255}
	default:
		// Default to white (most visible on dark backgrounds)
		return color.RGBA{255, 255, 255, 255}
	}
}

// divergingColor returns a red-white-blue diverging colormap for difference visualization.
// t ranges from -1 to 1 where: negative = blue, zero = white, positive = red.
func divergingColor(t float64) color.RGBA {
	// Clamp t to [-1, 1]
	if t < -1 {
		t = -1
	} else if t > 1 {
		t = 1
	}

	if t < 0 {
		// Blue gradient (negative differences)
		s := -t // 0 to 1
		return color.RGBA{
			R: uint8(255 * (1 - s)),
			G: uint8(255 * (1 - s)),
			B: 255,
			A: 255,
		}
	}
	// Red gradient (positive differences)
	s := t // 0 to 1
	return color.RGBA{
		R: 255,
		G: uint8(255 * (1 - s)),
		B: uint8(255 * (1 - s)),
		A: 255,
	}
}

// heatmapRenderer implements fyne.WidgetRenderer.
type heatmapRenderer struct {
	heatmap  *CrossbarHeatmap
	raster   *canvas.Raster
	altLabel *widget.Label
	cache    sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *heatmapRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("heatmapRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	labelHeight := float32(36)
	if size.Height <= labelHeight {
		r.raster.Resize(size)
		r.altLabel.Hide()
		r.cache.MarkLayout(size)
		return
	}
	r.altLabel.Show()
	r.raster.Move(fyne.NewPos(0, 0))
	r.raster.Resize(fyne.NewSize(size.Width, size.Height-labelHeight))
	r.altLabel.Move(fyne.NewPos(0, size.Height-labelHeight))
	r.altLabel.Resize(fyne.NewSize(size.Width, labelHeight))
	r.cache.MarkLayout(size)
}

func (r *heatmapRenderer) MinSize() fyne.Size {
	return r.heatmap.MinSize()
}

func (r *heatmapRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("heatmapRenderer", r.heatmap.Size())
	// Only refresh if data has actually changed - the heatmap controls this via its rate limiter
	r.raster.Refresh()
	r.altLabel.SetText(r.heatmap.TextAlternative())
	r.altLabel.Refresh()
}

func (r *heatmapRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster, r.altLabel}
}

func (r *heatmapRenderer) Destroy() {}
