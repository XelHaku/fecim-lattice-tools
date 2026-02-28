package render3d

import (
	"image"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Compile-time interface checks
var _ fyne.Widget = (*StackWidget)(nil)
var _ desktop.Hoverable = (*StackWidget)(nil)

// StackWidget is a Fyne widget that displays the 3D layer stack.
type StackWidget struct {
	widget.BaseWidget

	renderer *StackRenderer
	raster   *canvas.Raster

	mu sync.RWMutex // Protects renderer state

	// Interaction state
	dragging   bool
	lastMouseX float64
	lastMouseY float64

	// Interaction callbacks
	OnLayerSelected func(layer int)
	OnRotated       func(azimuth, elevation float64)
}

// NewStackWidget creates a new 3D stack visualization widget.
func NewStackWidget() *StackWidget {
	w := &StackWidget{
		renderer: NewStackRenderer(),
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetLayers updates the layer data to be rendered.
func (w *StackWidget) SetLayers(layers []LayerData) {
	w.mu.Lock()
	w.renderer.Layers = layers
	w.mu.Unlock()
	w.safeRefresh()
}

// SetCamera sets the camera parameters.
func (w *StackWidget) SetCamera(azimuth, elevation, zoom float64) {
	w.mu.Lock()
	w.renderer.Azimuth = azimuth
	w.renderer.Elevation = elevation
	w.renderer.Zoom = zoom
	w.mu.Unlock()
	w.safeRefresh()
}

// SetColormap changes the colormap used for rendering.
func (w *StackWidget) SetColormap(name string) {
	w.mu.Lock()
	w.renderer.Colormap = name
	w.mu.Unlock()
	w.safeRefresh()
}

// SetShowWires enables or disables inter-layer wire rendering.
func (w *StackWidget) SetShowWires(show bool) {
	w.mu.Lock()
	w.renderer.ShowWires = show
	w.mu.Unlock()
	w.safeRefresh()
}

// SetLayerGap sets the vertical gap between layers.
func (w *StackWidget) SetLayerGap(gap float64) {
	w.mu.Lock()
	w.renderer.LayerGap = gap
	w.mu.Unlock()
	w.safeRefresh()
}

// SetMaxVisibleLayers sets the maximum number of layers to render.
func (w *StackWidget) SetMaxVisibleLayers(n int) {
	w.mu.Lock()
	w.renderer.MaxVisibleLayers = n
	w.mu.Unlock()
	w.safeRefresh()
}

// SetSelectedLayer sets which layer is highlighted.
func (w *StackWidget) SetSelectedLayer(layer int) {
	w.mu.Lock()
	w.renderer.SelectedLayer = layer
	w.mu.Unlock()
	w.safeRefresh()
}

// AzimuthAngle returns the current azimuth angle in radians.
func (w *StackWidget) AzimuthAngle() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.renderer.Azimuth
}

// Elevation returns the current elevation angle in radians.
func (w *StackWidget) Elevation() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.renderer.Elevation
}

// ZoomLevel returns the current zoom level.
func (w *StackWidget) ZoomLevel() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.renderer.Zoom
}

// Rotate adjusts the camera by the given deltas.
func (w *StackWidget) Rotate(dAzimuth, dElevation float64) {
	w.mu.Lock()
	w.renderer.Azimuth += dAzimuth
	// Clamp elevation to avoid flipping
	w.renderer.Elevation += dElevation
	if w.renderer.Elevation < 0.05 {
		w.renderer.Elevation = 0.05
	}
	if w.renderer.Elevation > math.Pi/2-0.05 {
		w.renderer.Elevation = math.Pi/2 - 0.05
	}
	azimuth := w.renderer.Azimuth
	elevation := w.renderer.Elevation
	w.mu.Unlock()

	if w.OnRotated != nil {
		w.OnRotated(azimuth, elevation)
	}
	w.safeRefresh()
}

// Zoom adjusts the zoom level by the given factor.
func (w *StackWidget) Zoom(factor float64) {
	w.mu.Lock()
	w.renderer.Zoom *= factor
	if w.renderer.Zoom < 0.1 {
		w.renderer.Zoom = 0.1
	}
	if w.renderer.Zoom > 10.0 {
		w.renderer.Zoom = 10.0
	}
	w.mu.Unlock()
	w.safeRefresh()
}

// MinSize returns the minimum size for this widget.
func (w *StackWidget) MinSize() fyne.Size {
	return fyne.NewSize(200, 150)
}

// CreateRenderer creates the Fyne widget renderer.
func (w *StackWidget) CreateRenderer() fyne.WidgetRenderer {
	w.raster = canvas.NewRaster(w.generateImage)
	return &stackWidgetRenderer{
		widget: w,
		raster: w.raster,
	}
}

// generateImage is called by the canvas.Raster to produce the image.
func (w *StackWidget) generateImage(width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	w.mu.RLock()
	w.renderer.Width = width
	w.renderer.Height = height
	img := w.renderer.Render()
	w.mu.RUnlock()

	return img
}

// Tapped handles tap/click events for layer selection.
func (w *StackWidget) Tapped(e *fyne.PointEvent) {
	w.mu.RLock()
	w.renderer.Width = int(w.Size().Width)
	w.renderer.Height = int(w.Size().Height)
	layer := w.renderer.HitTest(float64(e.Position.X), float64(e.Position.Y))
	w.mu.RUnlock()

	if layer >= 0 {
		w.mu.Lock()
		w.renderer.SelectedLayer = layer
		w.mu.Unlock()
		w.safeRefresh()

		if w.OnLayerSelected != nil {
			w.OnLayerSelected(layer)
		}
	}
}

// TappedSecondary clears selection.
func (w *StackWidget) TappedSecondary(*fyne.PointEvent) {
	w.mu.Lock()
	w.renderer.SelectedLayer = -1
	w.mu.Unlock()
	w.safeRefresh()
}

// --- desktop.Hoverable interface for drag rotation ---

// MouseIn is called when the mouse enters the widget.
func (w *StackWidget) MouseIn(*desktop.MouseEvent) {}

// MouseOut is called when the mouse leaves the widget.
func (w *StackWidget) MouseOut() {
	w.dragging = false
}

// MouseMoved handles mouse drag for rotation.
func (w *StackWidget) MouseMoved(e *desktop.MouseEvent) {
	if w.dragging {
		dx := float64(e.Position.X) - w.lastMouseX
		dy := float64(e.Position.Y) - w.lastMouseY

		// Convert pixel delta to rotation (radians)
		const sensitivity = 0.005
		w.Rotate(dx*sensitivity, -dy*sensitivity)
	}
	w.lastMouseX = float64(e.Position.X)
	w.lastMouseY = float64(e.Position.Y)
}

// Dragged handles drag events for rotation.
func (w *StackWidget) Dragged(e *fyne.DragEvent) {
	dx := float64(e.Dragged.DX)
	dy := float64(e.Dragged.DY)

	const sensitivity = 0.005
	w.Rotate(dx*sensitivity, -dy*sensitivity)
}

// DragEnd is called when dragging stops.
func (w *StackWidget) DragEnd() {}

// Scrolled handles scroll events for zoom.
func (w *StackWidget) Scrolled(e *fyne.ScrollEvent) {
	if e.Scrolled.DY > 0 {
		w.Zoom(1.1)
	} else if e.Scrolled.DY < 0 {
		w.Zoom(0.9)
	}
}

// safeRefresh schedules a widget refresh on the main thread.
func (w *StackWidget) safeRefresh() {
	fyne.Do(func() {
		w.BaseWidget.Refresh()
	})
}

// stackWidgetRenderer implements fyne.WidgetRenderer for the StackWidget.
type stackWidgetRenderer struct {
	widget *StackWidget
	raster *canvas.Raster
}

func (r *stackWidgetRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *stackWidgetRenderer) MinSize() fyne.Size {
	return r.widget.MinSize()
}

func (r *stackWidgetRenderer) Refresh() {
	r.raster.Refresh()
}

func (r *stackWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *stackWidgetRenderer) Destroy() {}
