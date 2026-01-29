// Package render provides Vulkan-based visualization for ferroelectric simulations.
// Design inspired by Datoviz (datoviz.org) patterns for high-performance scientific visualization.
package render

import (
	"math"

	"fecim-lattice-tools/shared/physics"
)

// Config contains rendering configuration.
type Config struct {
	Width        int    // Window width in pixels
	Height       int    // Window height in pixels
	Title        string // Window title
	TargetFPS    int    // Target frames per second
	VSync        bool   // Enable vertical sync
	Antialiasing bool   // Enable antialiasing
}

// DefaultConfig returns sensible defaults for the hysteresis visualizer.
func DefaultConfig() *Config {
	return &Config{
		Width:        1280,
		Height:       720,
		Title:        "FeCIM Hysteresis Visualizer",
		TargetFPS:    60,
		VSync:        true,
		Antialiasing: true,
	}
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A float32
}

// ColorMap provides color mapping for polarization visualization.
type ColorMap struct {
	Positive Color // Color for +Ps (red)
	Zero     Color // Color for P=0 (white)
	Negative Color // Color for -Ps (blue)
}

// DefaultColorMap returns a divergent red-white-blue color map.
func DefaultColorMap() *ColorMap {
	return &ColorMap{
		Positive: Color{0.8, 0.1, 0.1, 1.0}, // Red for +P
		Zero:     Color{1.0, 1.0, 1.0, 1.0}, // White for P=0
		Negative: Color{0.1, 0.2, 0.8, 1.0}, // Blue for -P
	}
}

// PolarizationToColor converts normalized polarization (-1 to +1) to a color.
func (cm *ColorMap) PolarizationToColor(normP float64) Color {
	// Clamp to [-1, 1]
	if normP > 1.0 {
		normP = 1.0
	} else if normP < -1.0 {
		normP = -1.0
	}

	var c Color
	if normP >= 0 {
		// Interpolate between zero and positive
		t := float32(normP)
		c.R = cm.Zero.R + t*(cm.Positive.R-cm.Zero.R)
		c.G = cm.Zero.G + t*(cm.Positive.G-cm.Zero.G)
		c.B = cm.Zero.B + t*(cm.Positive.B-cm.Zero.B)
	} else {
		// Interpolate between negative and zero
		t := float32(-normP)
		c.R = cm.Zero.R + t*(cm.Negative.R-cm.Zero.R)
		c.G = cm.Zero.G + t*(cm.Negative.G-cm.Zero.G)
		c.B = cm.Zero.B + t*(cm.Negative.B-cm.Zero.B)
	}
	c.A = 1.0

	return c
}

// Point2D represents a 2D point for plotting.
type Point2D struct {
	X, Y float64
}

// HysteresisPlot contains data for the P-E hysteresis curve.
type HysteresisPlot struct {
	// Data points
	Points []Point2D

	// Axis limits
	EMin, EMax float64 // Electric field range
	PMin, PMax float64 // Polarization range

	// Current position marker
	CurrentE float64
	CurrentP float64

	// Visual properties
	LineColor   Color
	MarkerColor Color
	LineWidth   float32
	MarkerSize  float32
}

// NewHysteresisPlot creates a new plot with default settings.
func NewHysteresisPlot(Emax, Pmax float64) *HysteresisPlot {
	return &HysteresisPlot{
		Points:      make([]Point2D, 0, 1000),
		EMin:        -Emax,
		EMax:        Emax,
		PMin:        -Pmax,
		PMax:        Pmax,
		LineColor:   Color{0.2, 0.4, 0.8, 1.0},
		MarkerColor: Color{1.0, 0.3, 0.3, 1.0},
		LineWidth:   2.0,
		MarkerSize:  8.0,
	}
}

// AddPoint adds a new data point to the hysteresis curve.
func (hp *HysteresisPlot) AddPoint(E, P float64) {
	hp.Points = append(hp.Points, Point2D{X: E, Y: P})
	hp.CurrentE = E
	hp.CurrentP = P

	// Keep buffer manageable (circular buffer behavior)
	if len(hp.Points) > 10000 {
		hp.Points = hp.Points[1000:]
	}
}

// Clear removes all data points from the plot.
func (hp *HysteresisPlot) Clear() {
	hp.Points = hp.Points[:0]
	hp.CurrentE = 0
	hp.CurrentP = 0
}

// NormalizeToScreen converts data coordinates to screen coordinates (0-1 range).
func (hp *HysteresisPlot) NormalizeToScreen(E, P float64) (float64, float64) {
	x := (E - hp.EMin) / (hp.EMax - hp.EMin)
	y := (P - hp.PMin) / (hp.PMax - hp.PMin)
	return x, y
}

// CellDisplay represents the ferroelectric cell visualization.
type CellDisplay struct {
	// Position and size (normalized 0-1)
	X, Y          float64
	Width, Height float64

	// Current polarization state
	Polarization float64 // -1 to +1

	// Color map
	ColorMap *ColorMap
}

// NewCellDisplay creates a new cell display.
func NewCellDisplay() *CellDisplay {
	return &CellDisplay{
		X:        0.05,
		Y:        0.3,
		Width:    0.25,
		Height:   0.4,
		ColorMap: DefaultColorMap(),
	}
}

// GetColor returns the current color based on polarization.
func (cd *CellDisplay) GetColor() Color {
	return cd.ColorMap.PolarizationToColor(cd.Polarization)
}

// FeCIMLevels is re-exported from shared/physics for backward compatibility.
// "It's got 30 discrete states. So it's not 0-1-0-1." — Dr. Tour
const FeCIMLevels = physics.DefaultLevels

// LevelIndicator displays the 30 discrete FeCIM levels.
type LevelIndicator struct {
	// Position and size (normalized 0-1)
	X, Y          float64
	Width, Height float64

	// Current level (0-29)
	CurrentLevel int

	// Color scheme
	ActiveColor   Color // Color for current level
	InactiveColor Color // Color for other levels
	BorderColor   Color
}

// NewLevelIndicator creates a new level indicator display.
func NewLevelIndicator() *LevelIndicator {
	return &LevelIndicator{
		X:             0.02,
		Y:             0.1,
		Width:         0.08,
		Height:        0.8,
		CurrentLevel:  15,                        // Middle level
		ActiveColor:   Color{0.2, 0.7, 0.3, 1.0}, // Green for active
		InactiveColor: Color{0.3, 0.3, 0.3, 0.5}, // Gray for inactive
		BorderColor:   Color{0.2, 0.2, 0.2, 1.0},
	}
}

// SetFromPolarization sets the level from normalized polarization (-1 to +1).
func (li *LevelIndicator) SetFromPolarization(normP float64) {
	// Map [-1, 1] to [0, 29]
	normalized := (normP + 1.0) / 2.0 // [0, 1]
	li.CurrentLevel = int(math.Round(normalized * float64(FeCIMLevels-1)))
	if li.CurrentLevel < 0 {
		li.CurrentLevel = 0
	}
	if li.CurrentLevel >= FeCIMLevels {
		li.CurrentLevel = FeCIMLevels - 1
	}
}

// GetLevelVertices generates vertices for the 30-level indicator bar.
func (li *LevelIndicator) GetLevelVertices() []PlotVertex {
	var vertices []PlotVertex

	// Each level is a small horizontal bar
	levelHeight := li.Height / float64(FeCIMLevels)
	barWidth := li.Width * 0.8
	barX := li.X + li.Width*0.1

	for level := 0; level < FeCIMLevels; level++ {
		// Calculate Y position (level 0 at bottom, level 29 at top)
		barY := li.Y + float64(level)*levelHeight

		// Choose color based on whether this is the active level
		var color Color
		if level == li.CurrentLevel {
			color = li.ActiveColor
		} else {
			color = li.InactiveColor
		}

		// Convert to NDC
		x1 := float32(barX*2 - 1)
		y1 := float32(barY*2 - 1)
		x2 := float32((barX+barWidth)*2 - 1)
		y2 := float32((barY+levelHeight*0.85)*2 - 1) // 85% height for gap

		colorArr := [4]float32{color.R, color.G, color.B, color.A}

		// Two triangles forming a quad
		vertices = append(vertices,
			PlotVertex{Position: [2]float32{x1, y1}, Color: colorArr},
			PlotVertex{Position: [2]float32{x2, y1}, Color: colorArr},
			PlotVertex{Position: [2]float32{x1, y2}, Color: colorArr},
			PlotVertex{Position: [2]float32{x2, y1}, Color: colorArr},
			PlotVertex{Position: [2]float32{x2, y2}, Color: colorArr},
			PlotVertex{Position: [2]float32{x1, y2}, Color: colorArr},
		)
	}

	// Add border around entire indicator
	borderColor := [4]float32{li.BorderColor.R, li.BorderColor.G, li.BorderColor.B, li.BorderColor.A}
	bx1 := float32(li.X*2 - 1)
	by1 := float32(li.Y*2 - 1)
	bx2 := float32((li.X+li.Width)*2 - 1)
	by2 := float32((li.Y+li.Height)*2 - 1)

	vertices = append(vertices,
		// Top edge
		PlotVertex{Position: [2]float32{bx1, by2}, Color: borderColor},
		PlotVertex{Position: [2]float32{bx2, by2}, Color: borderColor},
		// Right edge
		PlotVertex{Position: [2]float32{bx2, by2}, Color: borderColor},
		PlotVertex{Position: [2]float32{bx2, by1}, Color: borderColor},
		// Bottom edge
		PlotVertex{Position: [2]float32{bx2, by1}, Color: borderColor},
		PlotVertex{Position: [2]float32{bx1, by1}, Color: borderColor},
		// Left edge
		PlotVertex{Position: [2]float32{bx1, by1}, Color: borderColor},
		PlotVertex{Position: [2]float32{bx1, by2}, Color: borderColor},
	)

	return vertices
}

// QuantizeTo30Levels quantizes a value to exactly 30 discrete levels.
// Matches the crossbar package implementation for consistency.
func QuantizeTo30Levels(value float64) float64 {
	value = math.Max(0, math.Min(1, value))
	level := math.Round(value * float64(FeCIMLevels-1))
	return level / float64(FeCIMLevels-1)
}

// GetLevel returns the discrete level (0-29) for a normalized value.
func GetLevel(value float64) int {
	value = math.Max(0, math.Min(1, value))
	return int(math.Round(value * float64(FeCIMLevels-1)))
}

// Renderer is the main rendering interface.
// TODO: Implement with actual Vulkan calls using go-vk or vgpu.
type Renderer struct {
	config  *Config
	plot    *HysteresisPlot
	cell    *CellDisplay
	levels  *LevelIndicator // 30-level indicator per Dr. Tour's specs
	running bool

	// Callbacks
	onUpdate func()
}

// NewRenderer creates a new renderer with the given configuration.
func NewRenderer(config *Config) *Renderer {
	return &Renderer{
		config: config,
		cell:   NewCellDisplay(),
		levels: NewLevelIndicator(),
	}
}

// SetHysteresisPlot sets the plot to be rendered.
func (r *Renderer) SetHysteresisPlot(plot *HysteresisPlot) {
	r.plot = plot
}

// SetUpdateCallback sets a function to be called each frame.
func (r *Renderer) SetUpdateCallback(fn func()) {
	r.onUpdate = fn
}

// UpdatePolarization updates the cell polarization display and 30-level indicator.
func (r *Renderer) UpdatePolarization(normP float64) {
	r.cell.Polarization = normP
	r.levels.SetFromPolarization(normP)
}

// GetCurrentLevel returns the current discrete level (0-29).
func (r *Renderer) GetCurrentLevel() int {
	return r.levels.CurrentLevel
}

// GetLevelIndicator returns the level indicator for direct access.
func (r *Renderer) GetLevelIndicator() *LevelIndicator {
	return r.levels
}

// Initialize sets up the Vulkan context and window.
// TODO: Implement actual Vulkan initialization.
func (r *Renderer) Initialize() error {
	// Placeholder for Vulkan initialization:
	// 1. Create GLFW window
	// 2. Initialize Vulkan instance
	// 3. Create surface and device
	// 4. Setup swap chain
	// 5. Create render pass and pipelines
	// 6. Create command buffers

	return nil
}

// Run starts the main render loop.
// TODO: Implement actual render loop.
func (r *Renderer) Run() error {
	r.running = true

	// Placeholder render loop structure:
	// for r.running {
	//     // Poll events
	//     // Call update callback
	//     // Begin frame
	//     // Record commands: draw cell, draw plot
	//     // End frame and present
	// }

	return nil
}

// Stop terminates the render loop.
func (r *Renderer) Stop() {
	r.running = false
}

// Cleanup releases all Vulkan resources.
func (r *Renderer) Cleanup() {
	// TODO: Release Vulkan resources
}

// DrawAxes generates vertices for plot axes.
func DrawAxes(xMin, xMax, yMin, yMax float64) []Point2D {
	return []Point2D{
		// X axis
		{xMin, 0}, {xMax, 0},
		// Y axis
		{0, yMin}, {0, yMax},
	}
}

// GenerateGridLines generates vertices for plot grid lines.
func GenerateGridLines(xMin, xMax, yMin, yMax float64, divisions int) []Point2D {
	var points []Point2D

	dx := (xMax - xMin) / float64(divisions)
	dy := (yMax - yMin) / float64(divisions)

	// Vertical lines
	for i := 0; i <= divisions; i++ {
		x := xMin + float64(i)*dx
		points = append(points, Point2D{x, yMin}, Point2D{x, yMax})
	}

	// Horizontal lines
	for i := 0; i <= divisions; i++ {
		y := yMin + float64(i)*dy
		points = append(points, Point2D{xMin, y}, Point2D{xMax, y})
	}

	return points
}

// LerpColor linearly interpolates between two colors.
func LerpColor(a, b Color, t float32) Color {
	return Color{
		R: a.R + t*(b.R-a.R),
		G: a.G + t*(b.G-a.G),
		B: a.B + t*(b.B-a.B),
		A: a.A + t*(b.A-a.A),
	}
}

// ScreenToNDC converts screen coordinates (pixels) to Normalized Device Coordinates.
func ScreenToNDC(x, y float64, width, height int) (float64, float64) {
	ndcX := 2.0*x/float64(width) - 1.0
	ndcY := 1.0 - 2.0*y/float64(height) // Flip Y
	return ndcX, ndcY
}

// NDCToScreen converts NDC to screen coordinates.
func NDCToScreen(ndcX, ndcY float64, width, height int) (float64, float64) {
	x := (ndcX + 1.0) * float64(width) / 2.0
	y := (1.0 - ndcY) * float64(height) / 2.0
	return x, y
}

// SmoothStep provides smooth interpolation (for animations).
func SmoothStep(edge0, edge1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-edge0)/(edge1-edge0)))
	return t * t * (3 - 2*t)
}
