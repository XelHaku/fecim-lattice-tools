// Package widgets provides custom Fyne widgets for the hysteresis GUI.
package widgets

import (
	"fmt"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// CellVisualizer displays a single ferroelectric memory cell with its polarization state
type CellVisualizer struct {
	widget.BaseWidget

	mu        sync.RWMutex
	level     int
	numLevels int // Total number of levels (default 30)
	minSize   fyne.Size
}

// NewCellVisualizer creates a new cell visualizer
func NewCellVisualizer() *CellVisualizer {
	c := &CellVisualizer{
		level:     15,
		numLevels: 30,                     // Default to 30 levels
		minSize:   fyne.NewSize(180, 200), // Larger cell display (30% increase)
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *CellVisualizer) SetMinSize(size fyne.Size) {
	c.minSize = size
}

func (c *CellVisualizer) MinSize() fyne.Size {
	return c.minSize
}

func (c *CellVisualizer) SetLevel(level int) {
	c.mu.Lock()
	c.level = level
	c.mu.Unlock()
}

// SetNumLevels sets the total number of levels (for display purposes)
func (c *CellVisualizer) SetNumLevels(numLevels int) {
	c.mu.Lock()
	c.numLevels = numLevels
	c.mu.Unlock()
}

func (c *CellVisualizer) CreateRenderer() fyne.WidgetRenderer {
	return &cellRenderer{cell: c}
}

type cellRenderer struct {
	cell    *CellVisualizer
	objects []fyne.CanvasObject
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *cellRenderer) MinSize() fyne.Size {
	return r.cell.minSize
}

func (r *cellRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("cellRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *cellRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("cellRenderer", r.cell.Size())
	size := r.cell.Size()
	// Always re-layout on Refresh for this dynamic widget (level changes)
	r.layoutWithSize(size)
}

func (r *cellRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.cell.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.cell.mu.RLock()
	level := r.cell.level
	numLevels := r.cell.numLevels
	r.cell.mu.RUnlock()

	// Fallback to 30 if numLevels not set
	if numLevels < 2 {
		numLevels = 30
	}

	r.objects = r.objects[:0]

	// Constrain to minimum size to prevent growing
	minSize := r.cell.minSize
	if size.Width > minSize.Width {
		size.Width = minSize.Width
	}
	if size.Height > minSize.Height {
		size.Height = minSize.Height
	}

	// Background with subtle gradient effect
	bg := canvas.NewRectangle(color.RGBA{0, 35, 70, 255})
	bg.Resize(size)
	r.objects = append(r.objects, bg)

	// Calculate cell size and position - larger cell
	margin := float32(15)
	cellSize := size.Width - 2*margin
	if size.Height-70 < cellSize {
		cellSize = size.Height - 70
	}
	cellX := (size.Width - cellSize) / 2
	cellY := margin + 5

	// Outer glow (based on cell color)
	t := float64(level) / 29.0
	var glowColor color.RGBA
	if t < 0.5 {
		glowColor = color.RGBA{100, 150, 255, 60} // Blue glow
	} else {
		glowColor = color.RGBA{255, 150, 100, 60} // Red glow
	}
	outerGlow := canvas.NewRectangle(glowColor)
	outerGlow.Resize(fyne.NewSize(cellSize+20, cellSize+20))
	outerGlow.Move(fyne.NewPos(cellX-10, cellY-10))
	r.objects = append(r.objects, outerGlow)

	// Cell border (electrode representation) - cyan accent
	borderWidth := float32(4)
	border := canvas.NewRectangle(color.RGBA{0, 180, 220, 255})
	border.Resize(fyne.NewSize(cellSize+borderWidth*2, cellSize+borderWidth*2))
	border.Move(fyne.NewPos(cellX-borderWidth, cellY-borderWidth))
	r.objects = append(r.objects, border)

	// Cell color based on level - intuitive gradient:
	// Level 1 (negative P) = Blue, Level 15 = White/neutral, Level 30 (positive P) = Red
	// This matches physics: negative polarization -> positive polarization
	// t is already calculated above (0.0 to 1.0)
	var cellColor color.RGBA
	if t < 0.5 {
		// Blue to white transition (levels 1-15, negative to neutral)
		t2 := t * 2 // 0 to 1
		cellColor = color.RGBA{
			uint8(80 + t2*175),  // R: 80 -> 255
			uint8(120 + t2*135), // G: 120 -> 255
			255,                 // B stays high
			255,
		}
	} else {
		// White to red transition (levels 16-30, neutral to positive)
		t2 := (t - 0.5) * 2 // 0 to 1
		cellColor = color.RGBA{
			255,                 // R stays high
			uint8(255 - t2*175), // G: 255 -> 80
			uint8(255 - t2*175), // B: 255 -> 80
			255,
		}
	}

	// The memory cell square
	cell := canvas.NewRectangle(cellColor)
	cell.Resize(fyne.NewSize(cellSize, cellSize))
	cell.Move(fyne.NewPos(cellX, cellY))
	r.objects = append(r.objects, cell)

	// Inner glow effect
	glowSize := cellSize * 0.5
	glowX := cellX + (cellSize-glowSize)/2
	glowY := cellY + (cellSize-glowSize)/2
	glow := canvas.NewRectangle(color.RGBA{
		uint8(min(int(cellColor.R)+20, 255)),
		uint8(min(int(cellColor.G)+20, 255)),
		uint8(min(int(cellColor.B)+20, 255)),
		80,
	})
	glow.Resize(fyne.NewSize(glowSize, glowSize))
	glow.Move(fyne.NewPos(glowX, glowY))
	r.objects = append(r.objects, glow)

	// Level text inside cell - larger and centered with cyan accent
	levelStr := fmt.Sprintf("%d", level+1)
	// Use cyan color for prominence (#00D4FF)
	textColor := color.RGBA{0, 212, 255, 255}

	// Scale text size with cell size - increased baseline
	textSize := cellSize * 0.45 // Increased from 0.35
	if textSize < 36 {
		textSize = 36 // Increased from 24
	}
	if textSize > 60 {
		textSize = 60 // Increased from 48
	}
	levelText := canvas.NewText(levelStr, textColor)
	levelText.TextSize = textSize
	levelText.TextStyle = fyne.TextStyle{Bold: true}
	// Calculate width based on text size
	textW := float32(len(levelStr)) * textSize * 0.6
	levelText.Move(fyne.NewPos(cellX+(cellSize-textW)/2, cellY+cellSize/2-textSize/2))
	r.objects = append(r.objects, levelText)

	// Label below cell (centered) - larger text
	labelY := cellY + cellSize + 10
	levelLabelStr := fmt.Sprintf("Level %d/%d", level+1, numLevels)
	levelLabel := canvas.NewText(levelLabelStr, color.RGBA{0, 212, 255, 255}) // Cyan for consistency
	levelLabel.TextSize = 16                                                  // Increased from 14
	levelLabel.TextStyle = fyne.TextStyle{Bold: true}
	labelW := float32(len(levelLabelStr)) * 9.5 // Adjusted for larger text
	levelLabel.Move(fyne.NewPos(cellX+(cellSize-labelW)/2, labelY))
	r.objects = append(r.objects, levelLabel)

	// State description (centered) - larger
	var stateText string
	if level < 10 {
		stateText = "Negative P"
	} else if level > 19 {
		stateText = "Positive P"
	} else {
		stateText = "Intermediate"
	}
	stateLabel := canvas.NewText(stateText, color.RGBA{220, 220, 220, 255})
	stateLabel.TextSize = 14 // Increased from 12
	stateLabelW := float32(len(stateText)) * 7.5
	stateLabel.Move(fyne.NewPos(cellX+(cellSize-stateLabelW)/2, labelY+20))
	r.objects = append(r.objects, stateLabel)

	// Mark cache with the effective size used
	r.cache.MarkLayout(size)
}

func (r *cellRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *cellRenderer) Destroy() {}
