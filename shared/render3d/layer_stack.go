package render3d

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"
)

// LayerData holds the conductance data for a single crossbar layer.
type LayerData struct {
	Values []float64 // Row-major, normalized [0,1]
	Rows   int
	Cols   int
	Label  string // e.g., "Layer 0", "Layer 127"
}

// StackRenderer renders a stack of crossbar layers in 3D using isometric projection.
type StackRenderer struct {
	Layers    []LayerData
	Width     int     // Output image width in pixels
	Height    int     // Output image height in pixels
	Azimuth   float64 // Camera rotation around Y (radians)
	Elevation float64 // Camera tilt from horizontal (radians)
	Zoom      float64 // Camera zoom level (1.0 = default)
	LayerGap  float64 // Vertical gap between layers (0-1 relative to layer size)
	ShowWires bool    // Draw interconnect wires between layers
	Colormap  string  // "viridis", "plasma", "coolwarm"

	// Interaction
	SelectedLayer int // -1 for none

	// Performance
	MaxVisibleLayers int // Limit rendered layers for performance (0=all)
}

// NewStackRenderer creates a StackRenderer with sensible defaults.
func NewStackRenderer() *StackRenderer {
	return &StackRenderer{
		Width:            800,
		Height:           600,
		Azimuth:          math.Pi / 6,
		Elevation:        math.Pi / 6,
		Zoom:             1.0,
		LayerGap:         0.15,
		Colormap:         "viridis",
		SelectedLayer:    -1,
		MaxVisibleLayers: 0,
	}
}

// layerQuad represents the projected 2D corners of a single layer for
// painter's-algorithm depth sorting and hit testing.
type layerQuad struct {
	index int     // Layer index in the Layers slice
	depth float64 // Average depth in camera space (for sorting)
	// 2D screen corners (top-left, top-right, bottom-right, bottom-left)
	corners [4][2]float64
}

// Render produces the 3D visualization as an image.
func (s *StackRenderer) Render() image.Image {
	w, h := s.Width, s.Height
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Dark background
	bg := color.RGBA{25, 25, 35, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	numLayers := len(s.Layers)
	if numLayers == 0 {
		return img
	}

	// Determine how many layers to render
	visibleCount := numLayers
	if s.MaxVisibleLayers > 0 && visibleCount > s.MaxVisibleLayers {
		visibleCount = s.MaxVisibleLayers
	}

	// Build layer list (subsample if needed)
	layerIndices := s.selectVisibleLayers(numLayers, visibleCount)

	// Compute scale: fit all layers within the viewport
	// Each layer occupies a 1x1 footprint in XZ, stacked along Y
	maxRows, maxCols := 0, 0
	for _, li := range layerIndices {
		ld := &s.Layers[li]
		if ld.Rows > maxRows {
			maxRows = ld.Rows
		}
		if ld.Cols > maxCols {
			maxCols = ld.Cols
		}
	}
	if maxRows == 0 || maxCols == 0 {
		return img
	}

	totalHeight := float64(len(layerIndices)-1)*(1.0+s.LayerGap) + 1.0
	footprint := math.Max(float64(maxCols), float64(maxRows))

	// Scale to fit in viewport with margin
	margin := 0.1
	scaleX := float64(w) * (1 - 2*margin) / (footprint * 1.5) // 1.5 to account for isometric spread
	scaleY := float64(h) * (1 - 2*margin) / (totalHeight + footprint)
	baseScale := math.Min(scaleX, scaleY) * s.Zoom

	proj := NewIsometricProjection(s.Azimuth, s.Elevation, baseScale)

	// Center offset: translate so the center of the stack maps to screen center
	centerWorld := Vec3{
		X: float64(maxCols) / 2,
		Y: totalHeight / 2,
		Z: float64(maxRows) / 2,
	}
	cx, cy := proj.Project(centerWorld)
	offsetX := float64(w)/2 - cx
	offsetY := float64(h)/2 - cy

	// Build layerQuads for depth sorting
	quads := make([]layerQuad, 0, len(layerIndices))
	for _, li := range layerIndices {
		ld := &s.Layers[li]
		// Layer Y position: distribute evenly
		// Use the position in the subsampled list for visual stacking
		stackPos := 0
		for si, idx := range layerIndices {
			if idx == li {
				stackPos = si
				break
			}
		}
		yPos := float64(stackPos) * (1.0 + s.LayerGap)

		// Four corners of the layer in world space
		c0 := Vec3{0, yPos, 0}                               // top-left
		c1 := Vec3{float64(ld.Cols), yPos, 0}                // top-right
		c2 := Vec3{float64(ld.Cols), yPos, float64(ld.Rows)} // bottom-right
		c3 := Vec3{0, yPos, float64(ld.Rows)}                // bottom-left

		// Project corners
		var corners [4][2]float64
		for ci, cv := range []Vec3{c0, c1, c2, c3} {
			sx, sy := proj.Project(cv)
			corners[ci] = [2]float64{sx + offsetX, sy + offsetY}
		}

		// Depth: transform center to camera space, use Z for sorting
		center := Vec3{float64(ld.Cols) / 2, yPos, float64(ld.Rows) / 2}
		transformed := proj.TransformVec3(center)

		quads = append(quads, layerQuad{
			index:   li,
			depth:   transformed.Z,
			corners: corners,
		})
	}

	// Sort back-to-front (painter's algorithm): larger depth = further away = draw first
	sort.Slice(quads, func(i, j int) bool {
		return quads[i].depth > quads[j].depth
	})

	cmap := GetColormap(s.Colormap)

	// Render each layer back-to-front
	for _, q := range quads {
		ld := &s.Layers[q.index]
		s.renderLayer(img, ld, q, proj, offsetX, offsetY, cmap)

		// Draw selected layer highlight
		if q.index == s.SelectedLayer {
			s.drawLayerHighlight(img, q.corners, color.RGBA{0, 255, 255, 200})
		}
	}

	// Draw inter-layer wires if enabled
	if s.ShowWires && len(quads) > 1 {
		s.drawWires(img, quads, proj, offsetX, offsetY)
	}

	// Draw layer labels along the left edge
	s.drawLabels(img, quads)

	// If layers were subsampled, draw an indicator
	if visibleCount < numLayers {
		s.drawEllipsis(img, numLayers, visibleCount)
	}

	return img
}

// selectVisibleLayers picks which layer indices to render when we have more
// layers than MaxVisibleLayers. Always includes first and last; evenly samples
// the rest.
func (s *StackRenderer) selectVisibleLayers(total, visible int) []int {
	if visible >= total {
		indices := make([]int, total)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	indices := make([]int, 0, visible)
	indices = append(indices, 0) // Always include first

	// Evenly sample middle layers
	step := float64(total-1) / float64(visible-1)
	for i := 1; i < visible-1; i++ {
		idx := int(math.Round(float64(i) * step))
		if idx > 0 && idx < total-1 {
			indices = append(indices, idx)
		}
	}

	indices = append(indices, total-1) // Always include last

	// Remove duplicates and sort
	seen := make(map[int]bool)
	unique := indices[:0]
	for _, idx := range indices {
		if !seen[idx] {
			seen[idx] = true
			unique = append(unique, idx)
		}
	}
	sort.Ints(unique)
	return unique
}

// renderLayer draws a single layer (all its cells) onto the image.
func (s *StackRenderer) renderLayer(img *image.RGBA, ld *LayerData, q layerQuad, proj Mat4, offsetX, offsetY float64, cmap ColormapFunc) {
	if ld.Rows == 0 || ld.Cols == 0 {
		return
	}

	// Compute the Y position for this layer from its quad center
	// We reconstruct it from the corners: average Y of all corners relates to yPos
	stackPos := 0
	indices := s.selectVisibleLayers(len(s.Layers), func() int {
		if s.MaxVisibleLayers > 0 && s.MaxVisibleLayers < len(s.Layers) {
			return s.MaxVisibleLayers
		}
		return len(s.Layers)
	}())
	for si, idx := range indices {
		if idx == q.index {
			stackPos = si
			break
		}
	}
	yPos := float64(stackPos) * (1.0 + s.LayerGap)

	// Draw each cell as a projected quadrilateral
	for row := 0; row < ld.Rows; row++ {
		for col := 0; col < ld.Cols; col++ {
			idx := row*ld.Cols + col
			val := 0.0
			if idx < len(ld.Values) {
				val = ld.Values[idx]
			}

			cellColor := cmap(val)

			// Cell corners in world space
			x0 := float64(col)
			x1 := float64(col + 1)
			z0 := float64(row)
			z1 := float64(row + 1)

			// Project 4 corners
			sx0, sy0 := proj.Project(Vec3{x0, yPos, z0})
			sx1, sy1 := proj.Project(Vec3{x1, yPos, z0})
			sx2, sy2 := proj.Project(Vec3{x1, yPos, z1})
			sx3, sy3 := proj.Project(Vec3{x0, yPos, z1})

			// Apply offset
			sx0 += offsetX
			sy0 += offsetY
			sx1 += offsetX
			sy1 += offsetY
			sx2 += offsetX
			sy2 += offsetY
			sx3 += offsetX
			sy3 += offsetY

			// Fill the quadrilateral using scanline
			fillQuad(img, [4][2]float64{
				{sx0, sy0}, {sx1, sy1}, {sx2, sy2}, {sx3, sy3},
			}, cellColor)
		}
	}

	// Draw layer edge lines (subtle border around the entire layer)
	edgeColor := color.RGBA{80, 100, 120, 180}
	drawLine(img, int(q.corners[0][0]), int(q.corners[0][1]), int(q.corners[1][0]), int(q.corners[1][1]), edgeColor)
	drawLine(img, int(q.corners[1][0]), int(q.corners[1][1]), int(q.corners[2][0]), int(q.corners[2][1]), edgeColor)
	drawLine(img, int(q.corners[2][0]), int(q.corners[2][1]), int(q.corners[3][0]), int(q.corners[3][1]), edgeColor)
	drawLine(img, int(q.corners[3][0]), int(q.corners[3][1]), int(q.corners[0][0]), int(q.corners[0][1]), edgeColor)
}

// drawLayerHighlight draws a bright border around a layer's projected quad.
func (s *StackRenderer) drawLayerHighlight(img *image.RGBA, corners [4][2]float64, c color.RGBA) {
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		x0, y0 := int(corners[i][0]), int(corners[i][1])
		x1, y1 := int(corners[j][0]), int(corners[j][1])
		drawLine(img, x0, y0, x1, y1, c)
		// Draw a second line offset by 1 pixel for thickness
		drawLine(img, x0+1, y0, x1+1, y1, c)
		drawLine(img, x0, y0+1, x1, y1+1, c)
	}
}

// drawWires draws vertical interconnect lines between corresponding cell centers
// on adjacent layers. Only draws a subset of wires to avoid clutter.
func (s *StackRenderer) drawWires(img *image.RGBA, quads []layerQuad, proj Mat4, offsetX, offsetY float64) {
	wireColor := color.RGBA{60, 80, 100, 100}

	// Draw wires only at corners and center of each layer pair
	for qi := 0; qi < len(quads)-1; qi++ {
		// Connect corners
		for ci := 0; ci < 4; ci++ {
			x0 := int(quads[qi].corners[ci][0])
			y0 := int(quads[qi].corners[ci][1])
			x1 := int(quads[qi+1].corners[ci][0])
			y1 := int(quads[qi+1].corners[ci][1])
			drawLine(img, x0, y0, x1, y1, wireColor)
		}
	}
}

// drawLabels renders layer labels next to each layer.
func (s *StackRenderer) drawLabels(img *image.RGBA, quads []layerQuad) {
	labelColor := color.RGBA{180, 190, 200, 220}

	for _, q := range quads {
		ld := &s.Layers[q.index]
		label := ld.Label
		if label == "" {
			continue
		}

		// Position label to the left of the layer's leftmost corner
		x := int(q.corners[0][0]) - 8
		y := int(q.corners[0][1])

		// Draw a simple label using pixel dots for each character
		// (full font rendering is out of scope; draw a small indicator dot)
		if x >= 0 && x < s.Width && y >= 0 && y < s.Height {
			// Draw a small marker dot
			for dx := -2; dx <= 2; dx++ {
				for dy := -2; dy <= 2; dy++ {
					px := x + dx
					py := y + dy
					if px >= 0 && px < s.Width && py >= 0 && py < s.Height {
						img.Set(px, py, labelColor)
					}
				}
			}
		}
	}
}

// drawEllipsis draws a "..." indicator when layers are subsampled.
func (s *StackRenderer) drawEllipsis(img *image.RGBA, total, visible int) {
	c := color.RGBA{150, 150, 160, 200}
	// Draw three dots near the bottom-right
	baseX := s.Width - 40
	baseY := s.Height - 15
	for i := 0; i < 3; i++ {
		x := baseX + i*8
		for dx := 0; dx < 3; dx++ {
			for dy := 0; dy < 3; dy++ {
				px := x + dx
				py := baseY + dy
				if px >= 0 && px < s.Width && py >= 0 && py < s.Height {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// HitTest returns which layer was clicked at the given screen coordinates.
// Returns -1 if no layer was hit. Tests front-to-back (reverse of draw order).
func (s *StackRenderer) HitTest(screenX, screenY float64) int {
	if len(s.Layers) == 0 {
		return -1
	}

	numLayers := len(s.Layers)
	visibleCount := numLayers
	if s.MaxVisibleLayers > 0 && visibleCount > s.MaxVisibleLayers {
		visibleCount = s.MaxVisibleLayers
	}
	layerIndices := s.selectVisibleLayers(numLayers, visibleCount)

	maxRows, maxCols := 0, 0
	for _, li := range layerIndices {
		ld := &s.Layers[li]
		if ld.Rows > maxRows {
			maxRows = ld.Rows
		}
		if ld.Cols > maxCols {
			maxCols = ld.Cols
		}
	}
	if maxRows == 0 || maxCols == 0 {
		return -1
	}

	totalHeight := float64(len(layerIndices)-1)*(1.0+s.LayerGap) + 1.0
	footprint := math.Max(float64(maxCols), float64(maxRows))
	margin := 0.1
	scaleX := float64(s.Width) * (1 - 2*margin) / (footprint * 1.5)
	scaleY := float64(s.Height) * (1 - 2*margin) / (totalHeight + footprint)
	baseScale := math.Min(scaleX, scaleY) * s.Zoom

	proj := NewIsometricProjection(s.Azimuth, s.Elevation, baseScale)

	centerWorld := Vec3{
		X: float64(maxCols) / 2,
		Y: totalHeight / 2,
		Z: float64(maxRows) / 2,
	}
	cx, cy := proj.Project(centerWorld)
	offsetX := float64(s.Width)/2 - cx
	offsetY := float64(s.Height)/2 - cy

	type depthLayer struct {
		index   int
		depth   float64
		corners [4][2]float64
	}

	dls := make([]depthLayer, 0, len(layerIndices))
	for si, li := range layerIndices {
		ld := &s.Layers[li]
		yPos := float64(si) * (1.0 + s.LayerGap)

		c0 := Vec3{0, yPos, 0}
		c1 := Vec3{float64(ld.Cols), yPos, 0}
		c2 := Vec3{float64(ld.Cols), yPos, float64(ld.Rows)}
		c3 := Vec3{0, yPos, float64(ld.Rows)}

		var corners [4][2]float64
		for ci, cv := range []Vec3{c0, c1, c2, c3} {
			sx, sy := proj.Project(cv)
			corners[ci] = [2]float64{sx + offsetX, sy + offsetY}
		}

		center := Vec3{float64(ld.Cols) / 2, yPos, float64(ld.Rows) / 2}
		transformed := proj.TransformVec3(center)

		dls = append(dls, depthLayer{
			index:   li,
			depth:   transformed.Z,
			corners: corners,
		})
	}

	// Sort front-to-back (smallest depth first) for hit testing
	sort.Slice(dls, func(i, j int) bool {
		return dls[i].depth < dls[j].depth
	})

	// Test point-in-quad for each layer (front to back)
	for _, dl := range dls {
		if pointInQuad(screenX, screenY, dl.corners) {
			return dl.index
		}
	}

	return -1
}

// fillQuad fills a convex quadrilateral using a scanline approach.
// corners are in screen coordinates (float64).
func fillQuad(img *image.RGBA, corners [4][2]float64, c color.RGBA) {
	bounds := img.Bounds()

	// Find bounding box
	minY := corners[0][1]
	maxY := corners[0][1]
	for i := 1; i < 4; i++ {
		if corners[i][1] < minY {
			minY = corners[i][1]
		}
		if corners[i][1] > maxY {
			maxY = corners[i][1]
		}
	}

	startY := int(math.Floor(minY))
	endY := int(math.Ceil(maxY))

	if startY < bounds.Min.Y {
		startY = bounds.Min.Y
	}
	if endY > bounds.Max.Y {
		endY = bounds.Max.Y
	}

	// For each scanline, find the X range where the quad intersects
	edges := [4][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}

	for y := startY; y < endY; y++ {
		fy := float64(y) + 0.5
		xMin := math.Inf(1)
		xMax := math.Inf(-1)

		for _, e := range edges {
			y0 := corners[e[0]][1]
			y1 := corners[e[1]][1]
			x0 := corners[e[0]][0]
			x1 := corners[e[1]][0]

			// Does this edge cross the scanline?
			if (y0 <= fy && y1 > fy) || (y1 <= fy && y0 > fy) {
				// Interpolate X at this Y
				t := (fy - y0) / (y1 - y0)
				x := x0 + t*(x1-x0)
				if x < xMin {
					xMin = x
				}
				if x > xMax {
					xMax = x
				}
			}
		}

		if xMin > xMax {
			continue
		}

		startX := int(math.Floor(xMin))
		endX := int(math.Ceil(xMax))
		if startX < bounds.Min.X {
			startX = bounds.Min.X
		}
		if endX > bounds.Max.X {
			endX = bounds.Max.X
		}

		// Fill scanline using direct pixel buffer access
		stride := img.Stride
		rowOffset := y * stride
		for x := startX; x < endX; x++ {
			off := rowOffset + x*4
			if off >= 0 && off+3 < len(img.Pix) {
				img.Pix[off] = c.R
				img.Pix[off+1] = c.G
				img.Pix[off+2] = c.B
				img.Pix[off+3] = c.A
			}
		}
	}
}

// drawLine draws a line using Bresenham's algorithm.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	bounds := img.Bounds()

	dx := x1 - x0
	dy := y1 - y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy

	for {
		if x0 >= bounds.Min.X && x0 < bounds.Max.X && y0 >= bounds.Min.Y && y0 < bounds.Max.Y {
			img.Set(x0, y0, c)
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// pointInQuad tests whether a point (px, py) is inside a convex quadrilateral.
// Uses the cross product sign test.
func pointInQuad(px, py float64, corners [4][2]float64) bool {
	positive := 0
	negative := 0

	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		// Edge vector
		ex := corners[j][0] - corners[i][0]
		ey := corners[j][1] - corners[i][1]
		// Point vector from corner[i] to point
		vx := px - corners[i][0]
		vy := py - corners[i][1]
		// Cross product
		cross := ex*vy - ey*vx
		if cross > 0 {
			positive++
		} else if cross < 0 {
			negative++
		}
	}

	// Point is inside if all cross products have the same sign
	return positive == 0 || negative == 0
}
