package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// DemoInfo holds information about a demo
type DemoInfo struct {
	Number      int
	Title       string
	Subtitle    string
	Description string
	Icon        string // Unicode icon for the demo
	Ready       bool
	WIP         bool // Work in progress - show indicator but still accessible
}

// GetDemos returns all demo information (6 consolidated demos)
func GetDemos() []DemoInfo {
	return []DemoInfo{
		{
			Number:      1,
			Title:       "Hysteresis",
			Subtitle:    "P-E Curve Physics",
			Description: "Discover how ferroelectric materials remember: watch polarization flip between 30 discrete states to store multi-bit data in a single device",
			Icon:        "~",
			Ready:       true,
		},
		{
			Number:      2,
			Title:       "Crossbar+",
			Subtitle:    "Compute-in-Memory Array",
			Description: "Watch matrix multiplication happen (simulated physics): see how crossbar arrays compute using Ohm's law, and explore real-world challenges like IR drop and sneak currents",
			Icon:        "#",
			Ready:       true,
		},
		{
			Number:      3,
			Title:       "MNIST",
			Subtitle:    "Neural Network Demo",
			Description: "Draw your own digits and watch AI recognize them instantly: experience 87% accuracy on the classic handwriting benchmark, powered by analog computing",
			Icon:        "9",
			Ready:       true,
		},
		{
			Number:      4,
			Title:       "Circuits",
			Subtitle:    "Peripheral Electronics",
			Description: "Learn how analog meets digital: explore DACs that write voltage pulses, TIAs that measure tiny currents, and ADCs that read multi-level states",
			Icon:        "V",
			Ready:       true,
		},
		{
			Number:      5,
			Title:       "Comparison",
			Subtitle:    "Technology Benchmarks",
			Description: "Compare FeCIM to alternatives (with TRL caveats): energy, speed, and density vs NAND flash, DRAM, and emerging memory - all projections marked",
			Icon:        "$",
			Ready:       true,
		},
		{
			Number:      6,
			Title:       "EDA",
			Subtitle:    "Chip Layout Tools",
			Description: "Explore chip layout concepts: visualize placement, generate layouts from code, and learn industry-standard formats (educational tool)",
			Icon:        "L",
			Ready:       true,
			WIP:         true,
		},
	}
}

// DemoCard creates a card widget for a demo
type DemoCard struct {
	widget.BaseWidget
	info     DemoInfo
	onTapped func()
	minSize  fyne.Size
}

// NewDemoCard creates a new demo card
func NewDemoCard(info DemoInfo, onTapped func()) *DemoCard {
	card := &DemoCard{
		info:     info,
		onTapped: onTapped,
		minSize:  fyne.NewSize(300, 180), // Moderate size
	}
	card.ExtendBaseWidget(card)
	return card
}

func (c *DemoCard) MinSize() fyne.Size {
	return c.minSize
}

func (c *DemoCard) Tapped(*fyne.PointEvent) {
	if c.info.Ready && c.onTapped != nil {
		c.onTapped()
	}
}

func (c *DemoCard) TappedSecondary(*fyne.PointEvent) {}

func (c *DemoCard) CreateRenderer() fyne.WidgetRenderer {
	return &demoCardRenderer{card: c}
}

type demoCardRenderer struct {
	card    *DemoCard
	objects []fyne.CanvasObject
	cache   sharedwidgets.LayoutCache // Shared utility for safe layout
}

func (r *demoCardRenderer) MinSize() fyne.Size {
	return r.card.minSize
}

func (r *demoCardRenderer) Layout(size fyne.Size) {
	sharedwidgets.DebugLayoutCall("demoCardRenderer", size)
	if !r.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *demoCardRenderer) Refresh() {
	sharedwidgets.DebugRefreshCall("demoCardRenderer", r.card.Size())
	size := r.card.Size()
	// Always rebuild if objects are empty (first render) or size changed
	if len(r.objects) == 0 || r.cache.ShouldLayout(size) {
		r.layoutWithSize(size)
		if size.Width > 0 && size.Height > 0 {
			r.cache.MarkLayout(size)
		}
	}
}

func (r *demoCardRenderer) layoutWithSize(size fyne.Size) {
	// Use minSize if provided size is invalid (for initial render)
	if size.Width <= 0 || size.Height <= 0 {
		size = r.card.minSize
		if size.Width <= 0 || size.Height <= 0 {
			return
		}
	}

	r.objects = r.objects[:0]
	info := r.card.info

	// Colors
	cyanColor := color.RGBA{0, 212, 255, 255}
	var bgColor, borderColor, headerBgColor, textColor, subtitleColor, descColor, numberBgColor color.RGBA

	if info.Ready {
		borderColor = cyanColor
		bgColor = color.RGBA{0, 45, 90, 255}
		headerBgColor = color.RGBA{0, 55, 110, 255}
		textColor = color.RGBA{255, 255, 255, 255}
		subtitleColor = cyanColor
		descColor = color.RGBA{180, 200, 220, 255}
		numberBgColor = color.RGBA{0, 80, 160, 255}
	} else {
		borderColor = color.RGBA{80, 90, 100, 255}
		bgColor = color.RGBA{30, 40, 50, 200}
		headerBgColor = color.RGBA{35, 45, 55, 200}
		textColor = color.RGBA{120, 130, 140, 255}
		subtitleColor = color.RGBA{100, 110, 120, 255}
		descColor = color.RGBA{100, 110, 120, 255}
		numberBgColor = color.RGBA{50, 60, 70, 255}
	}

	// Height scaling for responsive layouts (base reference: 140px)
	heightScale := size.Height / 140.0
	if heightScale < 0.9 {
		heightScale = 0.9 // Don't shrink too much
	}
	if heightScale > 1.5 {
		heightScale = 1.5 // Moderate growth
	}

	borderWidth := float32(2)
	cornerRadius := float32(8)
	headerHeight := float32(50) * heightScale

	// Outer border with rounded corners
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	border.CornerRadius = cornerRadius
	r.objects = append(r.objects, border)

	// Main background
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-borderWidth*2, size.Height-borderWidth*2))
	bg.Move(fyne.NewPos(borderWidth, borderWidth))
	bg.CornerRadius = cornerRadius - 1
	r.objects = append(r.objects, bg)

	// Header background
	headerBg := canvas.NewRectangle(headerBgColor)
	headerBg.Resize(fyne.NewSize(size.Width-borderWidth*2, headerHeight))
	headerBg.Move(fyne.NewPos(borderWidth, borderWidth))
	r.objects = append(r.objects, headerBg)

	// Number badge - compact circle (scales with height)
	badgeSize := float32(36) * heightScale
	badgeX := float32(14)
	badgeY := float32(8)*heightScale + borderWidth

	badgeBg := canvas.NewCircle(numberBgColor)
	badgeBg.Resize(fyne.NewSize(badgeSize, badgeSize))
	badgeBg.Move(fyne.NewPos(badgeX, badgeY))
	r.objects = append(r.objects, badgeBg)

	// Number text (scales with badge)
	numTextSize := float32(18) * heightScale
	numText := canvas.NewText(string('0'+byte(info.Number)), textColor)
	numText.TextSize = numTextSize
	numText.TextStyle = fyne.TextStyle{Bold: true}
	numText.Move(fyne.NewPos(badgeX+badgeSize/2-numTextSize*0.3, badgeY+badgeSize/2-numTextSize*0.6))
	r.objects = append(r.objects, numText)

	// Title (scales with height)
	titleSize := float32(22) * heightScale
	title := canvas.NewText(info.Title, textColor)
	title.TextSize = titleSize
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(badgeX+badgeSize+12, 10*heightScale))
	r.objects = append(r.objects, title)

	// Subtitle
	subtitleSize := float32(14) * heightScale
	subtitle := canvas.NewText(info.Subtitle, subtitleColor)
	subtitle.TextSize = subtitleSize
	subtitle.Move(fyne.NewPos(badgeX+badgeSize+12, 10*heightScale+titleSize+2))
	r.objects = append(r.objects, subtitle)

	// Status indicator - WIP badge or green dot (scales with card)
	if info.Ready {
		if info.WIP {
			// Work In Progress badge
			wipWidth := float32(70) * heightScale
			wipHeight := float32(20) * heightScale
			wipBg := canvas.NewRectangle(color.RGBA{255, 165, 0, 255})
			wipBg.Resize(fyne.NewSize(wipWidth, wipHeight))
			wipBg.Move(fyne.NewPos(size.Width-wipWidth-10*heightScale, 10*heightScale))
			wipBg.CornerRadius = 4 * heightScale
			r.objects = append(r.objects, wipBg)

			wipTextSize := float32(12) * heightScale
			wipText := canvas.NewText("WIP", color.RGBA{0, 0, 0, 255})
			wipText.TextSize = wipTextSize
			wipText.TextStyle = fyne.TextStyle{Bold: true}
			wipText.Move(fyne.NewPos(size.Width-wipWidth-10*heightScale+wipWidth/2-wipTextSize, 10*heightScale+wipHeight/2-wipTextSize*0.6))
			r.objects = append(r.objects, wipText)
		} else {
			// Green dot for ready
			dotSize := float32(12) * heightScale
			statusDot := canvas.NewCircle(color.RGBA{100, 255, 150, 255})
			statusDot.Resize(fyne.NewSize(dotSize, dotSize))
			statusDot.Move(fyne.NewPos(size.Width-dotSize-12*heightScale, 12*heightScale))
			r.objects = append(r.objects, statusDot)
		}
	}

	// Calculate available content area
	contentHeight := size.Height - headerHeight - borderWidth*2
	contentWidth := size.Width - borderWidth*2

	// Preview thumbnail - scale with card size, positioned at bottom right
	// Use ~35% of width and ~55% of content height for preview
	previewWidth := contentWidth * 0.35
	if previewWidth < 80 {
		previewWidth = 80
	}
	if previewWidth > 250 {
		previewWidth = 250
	}
	previewHeight := contentHeight * 0.6
	if previewHeight < 50 {
		previewHeight = 50
	}
	if previewHeight > 180 {
		previewHeight = 180
	}
	previewX := size.Width - previewWidth - 14*heightScale
	previewY := size.Height - previewHeight - 14*heightScale
	previewObjects := drawPreviewThumbnail(info.Number, previewX, previewY, previewWidth, previewHeight, cyanColor)
	r.objects = append(r.objects, previewObjects...)

	// Description - wrapped text below header, left of thumbnail area
	desc := info.Description
	descSize := float32(14) * heightScale // Scale with card
	if descSize < 12 {
		descSize = 12
	}
	maxWidth := size.Width - previewWidth - 40*heightScale // Leave space for preview
	lineY := headerHeight + borderWidth + 14*heightScale
	lineHeight := descSize + 6*heightScale

	// Calculate max lines based on available height
	// Reserve space for: header (top) + description top padding (10) + hint text at bottom (20)
	availableDescHeight := size.Height - headerHeight - borderWidth*2 - 10 - 20
	maxLines := int(availableDescHeight / lineHeight)
	if maxLines < 2 {
		maxLines = 2
	}
	if maxLines > 8 {
		maxLines = 8
	}

	words := splitWords(desc)
	line := ""
	lineCount := 0
	for _, word := range words {
		if lineCount >= maxLines {
			break
		}
		testLine := line + word + " "
		testText := canvas.NewText(testLine, descColor)
		testText.TextSize = descSize
		if testText.MinSize().Width > maxWidth && line != "" {
			lineText := canvas.NewText(line, descColor)
			lineText.TextSize = descSize
			lineText.Move(fyne.NewPos(14, lineY))
			r.objects = append(r.objects, lineText)
			lineY += lineHeight
			lineCount++
			line = word + " "
		} else {
			line = testLine
		}
	}
	if line != "" && lineCount < maxLines {
		lineText := canvas.NewText(line, descColor)
		lineText.TextSize = descSize
		lineText.Move(fyne.NewPos(14, lineY))
		r.objects = append(r.objects, lineText)
	}

	// Click hint at bottom left (scales with card)
	hintSize := float32(12) * heightScale
	if hintSize < 10 {
		hintSize = 10
	}
	hintText := canvas.NewText("Click to explore →", color.RGBA{100, 130, 160, 200})
	hintText.TextSize = hintSize
	hintText.Move(fyne.NewPos(14*heightScale, size.Height-20*heightScale))
	r.objects = append(r.objects, hintText)

	r.cache.MarkLayout(size)
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// drawPreviewThumbnail draws a preview graphic for the given demo number
// All elements scale proportionally with width/height
func drawPreviewThumbnail(demoNum int, x, y, width, height float32, accentColor color.RGBA) []fyne.CanvasObject {
	var objects []fyne.CanvasObject

	// Scale factors based on reference size of 100x65
	scaleX := width / 100
	scaleY := height / 65

	// Background for preview area
	previewBg := canvas.NewRectangle(color.RGBA{0, 35, 70, 200})
	previewBg.Resize(fyne.NewSize(width, height))
	previewBg.Move(fyne.NewPos(x, y))
	previewBg.CornerRadius = 4 * scaleX
	objects = append(objects, previewBg)

	// Border
	previewBorder := canvas.NewRectangle(color.RGBA{0, 100, 150, 255})
	previewBorder.Resize(fyne.NewSize(width, height))
	previewBorder.Move(fyne.NewPos(x, y))
	previewBorder.CornerRadius = 4 * scaleX
	previewBorder.StrokeWidth = 1
	previewBorder.StrokeColor = color.RGBA{0, 150, 200, 150}
	previewBorder.FillColor = color.Transparent
	objects = append(objects, previewBorder)

	centerX := x + width/2
	centerY := y + height/2

	switch demoNum {
	case 1: // Hysteresis - P-E curve (S-shaped loop)
		// Axes first (behind curve)
		axisColor := color.RGBA{80, 110, 140, 150}
		xAxis := canvas.NewLine(axisColor)
		xAxis.Position1 = fyne.NewPos(x+8*scaleX, centerY)
		xAxis.Position2 = fyne.NewPos(x+width-8*scaleX, centerY)
		xAxis.StrokeWidth = 1
		objects = append(objects, xAxis)
		yAxis := canvas.NewLine(axisColor)
		yAxis.Position1 = fyne.NewPos(centerX, y+6*scaleY)
		yAxis.Position2 = fyne.NewPos(centerX, y+height-6*scaleY)
		yAxis.StrokeWidth = 1
		objects = append(objects, yAxis)

		// Draw hysteresis loop
		loopColor := accentColor
		points := []fyne.Position{
			{X: centerX - 40*scaleX, Y: centerY + 20*scaleY},
			{X: centerX - 25*scaleX, Y: centerY + 18*scaleY},
			{X: centerX - 10*scaleX, Y: centerY + 8*scaleY},
			{X: centerX - 3*scaleX, Y: centerY - 5*scaleY},
			{X: centerX + 8*scaleX, Y: centerY - 18*scaleY},
			{X: centerX + 22*scaleX, Y: centerY - 22*scaleY},
			{X: centerX + 40*scaleX, Y: centerY - 20*scaleY},
		}
		for i := 0; i < len(points)-1; i++ {
			line := canvas.NewLine(loopColor)
			line.Position1 = points[i]
			line.Position2 = points[i+1]
			line.StrokeWidth = 2 * scaleX
			objects = append(objects, line)
		}
		// Return path
		pointsReturn := []fyne.Position{
			{X: centerX + 40*scaleX, Y: centerY - 20*scaleY},
			{X: centerX + 25*scaleX, Y: centerY - 18*scaleY},
			{X: centerX + 10*scaleX, Y: centerY - 8*scaleY},
			{X: centerX + 3*scaleX, Y: centerY + 5*scaleY},
			{X: centerX - 8*scaleX, Y: centerY + 18*scaleY},
			{X: centerX - 22*scaleX, Y: centerY + 22*scaleY},
			{X: centerX - 40*scaleX, Y: centerY + 20*scaleY},
		}
		for i := 0; i < len(pointsReturn)-1; i++ {
			line := canvas.NewLine(loopColor)
			line.Position1 = pointsReturn[i]
			line.Position2 = pointsReturn[i+1]
			line.StrokeWidth = 2 * scaleX
			objects = append(objects, line)
		}

	case 2: // Crossbar+ - 4x4 grid showing MVM operation
		gridSize := 4
		cellSize := 12 * scaleX
		spacing := 3 * scaleX
		gridWidth := float32(gridSize)*cellSize + float32(gridSize-1)*spacing
		startX := centerX - gridWidth/2
		startY := centerY - gridWidth/2

		// Word lines (horizontal)
		for row := 0; row < gridSize; row++ {
			line := canvas.NewLine(color.RGBA{255, 180, 50, 180})
			yPos := startY + float32(row)*(cellSize+spacing) + cellSize/2
			line.Position1 = fyne.NewPos(startX-8*scaleX, yPos)
			line.Position2 = fyne.NewPos(startX+gridWidth+8*scaleX, yPos)
			line.StrokeWidth = 1.5 * scaleX
			objects = append(objects, line)
		}
		// Bit lines (vertical)
		for col := 0; col < gridSize; col++ {
			line := canvas.NewLine(color.RGBA{0, 180, 255, 180})
			xPos := startX + float32(col)*(cellSize+spacing) + cellSize/2
			line.Position1 = fyne.NewPos(xPos, startY-8*scaleY)
			line.Position2 = fyne.NewPos(xPos, startY+gridWidth+8*scaleY)
			line.StrokeWidth = 1.5 * scaleX
			objects = append(objects, line)
		}

		// Cells with varying conductance
		conductances := [][]uint8{
			{255, 180, 60, 200},
			{100, 255, 140, 80},
			{200, 90, 255, 160},
			{150, 220, 100, 255},
		}
		for row := 0; row < gridSize; row++ {
			for col := 0; col < gridSize; col++ {
				intensity := conductances[row][col]
				cellColor := color.RGBA{0, intensity, intensity/2 + 80, 255}
				cell := canvas.NewRectangle(cellColor)
				cell.Resize(fyne.NewSize(cellSize, cellSize))
				cell.Move(fyne.NewPos(startX+float32(col)*(cellSize+spacing), startY+float32(row)*(cellSize+spacing)))
				cell.CornerRadius = 2 * scaleX
				objects = append(objects, cell)
			}
		}

		// Input dots (left)
		for row := 0; row < gridSize; row++ {
			inputDot := canvas.NewCircle(color.RGBA{255, 100, 100, 255})
			yPos := startY + float32(row)*(cellSize+spacing) + cellSize/2
			dotSize := 4 * scaleX
			inputDot.Resize(fyne.NewSize(dotSize, dotSize))
			inputDot.Move(fyne.NewPos(startX-12*scaleX, yPos-dotSize/2))
			objects = append(objects, inputDot)
		}

	case 3: // MNIST - handwritten digit "7"
		digitColor := color.RGBA{255, 200, 100, 255}
		strokeWidth := 3 * scaleX

		// Top stroke
		topLine := canvas.NewLine(digitColor)
		topLine.Position1 = fyne.NewPos(centerX-15*scaleX, centerY-20*scaleY)
		topLine.Position2 = fyne.NewPos(centerX+15*scaleX, centerY-20*scaleY)
		topLine.StrokeWidth = strokeWidth
		objects = append(objects, topLine)

		// Diagonal stroke
		diagLine := canvas.NewLine(digitColor)
		diagLine.Position1 = fyne.NewPos(centerX+12*scaleX, centerY-20*scaleY)
		diagLine.Position2 = fyne.NewPos(centerX-6*scaleX, centerY+22*scaleY)
		diagLine.StrokeWidth = strokeWidth
		objects = append(objects, diagLine)

		// Serif
		serifLine := canvas.NewLine(digitColor)
		serifLine.Position1 = fyne.NewPos(centerX-15*scaleX, centerY-20*scaleY)
		serifLine.Position2 = fyne.NewPos(centerX-15*scaleX, centerY-14*scaleY)
		serifLine.StrokeWidth = strokeWidth
		objects = append(objects, serifLine)

	case 4: // Circuits - TIA
		circuitColor := accentColor
		wireColor := color.RGBA{150, 180, 200, 200}

		// Op-amp triangle
		triPoints := []fyne.Position{
			{X: centerX - 20*scaleX, Y: centerY - 18*scaleY},
			{X: centerX - 20*scaleX, Y: centerY + 18*scaleY},
			{X: centerX + 15*scaleX, Y: centerY},
			{X: centerX - 20*scaleX, Y: centerY - 18*scaleY},
		}
		for i := 0; i < len(triPoints)-1; i++ {
			line := canvas.NewLine(circuitColor)
			line.Position1 = triPoints[i]
			line.Position2 = triPoints[i+1]
			line.StrokeWidth = 1.5 * scaleX
			objects = append(objects, line)
		}

		// Input/output lines
		inLine1 := canvas.NewLine(wireColor)
		inLine1.Position1 = fyne.NewPos(centerX-35*scaleX, centerY-10*scaleY)
		inLine1.Position2 = fyne.NewPos(centerX-20*scaleX, centerY-10*scaleY)
		inLine1.StrokeWidth = 1.5 * scaleX
		objects = append(objects, inLine1)
		inLine2 := canvas.NewLine(wireColor)
		inLine2.Position1 = fyne.NewPos(centerX-35*scaleX, centerY+10*scaleY)
		inLine2.Position2 = fyne.NewPos(centerX-20*scaleX, centerY+10*scaleY)
		inLine2.StrokeWidth = 1.5 * scaleX
		objects = append(objects, inLine2)
		outLine := canvas.NewLine(wireColor)
		outLine.Position1 = fyne.NewPos(centerX+15*scaleX, centerY)
		outLine.Position2 = fyne.NewPos(centerX+35*scaleX, centerY)
		outLine.StrokeWidth = 1.5 * scaleX
		objects = append(objects, outLine)

		// +/- symbols
		plusText := canvas.NewText("+", circuitColor)
		plusText.TextSize = 10 * scaleX
		plusText.Move(fyne.NewPos(centerX-17*scaleX, centerY-14*scaleY))
		objects = append(objects, plusText)
		minusText := canvas.NewText("−", circuitColor)
		minusText.TextSize = 12 * scaleX
		minusText.Move(fyne.NewPos(centerX-17*scaleX, centerY+2*scaleY))
		objects = append(objects, minusText)

	case 5: // Comparison - Bar chart
		barColors := []color.RGBA{
			{0, 230, 180, 255},
			{180, 80, 80, 255},
			{140, 100, 60, 255},
			{100, 80, 120, 255},
		}
		// Scale bar heights with available space
		maxBarH := height * 0.7
		barHeights := []float32{maxBarH, maxBarH * 0.25, maxBarH * 0.17, maxBarH * 0.37}
		barWidth := 16 * scaleX
		spacing := 6 * scaleX
		totalWidth := 4*barWidth + 3*spacing
		startX := centerX - totalWidth/2
		baseY := y + height - 8*scaleY

		for i := 0; i < 4; i++ {
			bar := canvas.NewRectangle(barColors[i])
			bar.Resize(fyne.NewSize(barWidth, barHeights[i]))
			bar.Move(fyne.NewPos(startX+float32(i)*(barWidth+spacing), baseY-barHeights[i]))
			bar.CornerRadius = 2 * scaleX
			objects = append(objects, bar)
		}

		// Star on winner
		starText := canvas.NewText("★", color.RGBA{255, 220, 100, 255})
		starText.TextSize = 10 * scaleX
		starText.Move(fyne.NewPos(startX+barWidth/2-4*scaleX, baseY-barHeights[0]-10*scaleY))
		objects = append(objects, starText)

	case 6: // EDA - Chip with I/O pads
		// Chip die
		chipColor := color.RGBA{40, 60, 80, 255}
		chipInner := canvas.NewRectangle(chipColor)
		chipW := width - 20*scaleX
		chipH := height - 20*scaleY
		chipX := x + 10*scaleX
		chipY := y + 10*scaleY
		chipInner.Resize(fyne.NewSize(chipW, chipH))
		chipInner.Move(fyne.NewPos(chipX, chipY))
		chipInner.CornerRadius = 2 * scaleX
		objects = append(objects, chipInner)

		// I/O pads
		padColor := color.RGBA{255, 200, 80, 255}
		padSize := 6 * scaleX
		// Top/bottom pads
		padCount := int(chipW / (padSize * 2.5))
		if padCount < 3 {
			padCount = 3
		}
		padSpacing := chipW / float32(padCount+1)
		for i := 1; i <= padCount; i++ {
			// Top
			pad := canvas.NewRectangle(padColor)
			pad.Resize(fyne.NewSize(padSize, padSize))
			pad.Move(fyne.NewPos(chipX+float32(i)*padSpacing-padSize/2, chipY-padSize/2))
			objects = append(objects, pad)
			// Bottom
			pad2 := canvas.NewRectangle(padColor)
			pad2.Resize(fyne.NewSize(padSize, padSize))
			pad2.Move(fyne.NewPos(chipX+float32(i)*padSpacing-padSize/2, chipY+chipH-padSize/2))
			objects = append(objects, pad2)
		}
		// Left/right pads
		padCountV := int(chipH / (padSize * 3))
		if padCountV < 2 {
			padCountV = 2
		}
		padSpacingV := chipH / float32(padCountV+1)
		for i := 1; i <= padCountV; i++ {
			pad := canvas.NewRectangle(padColor)
			pad.Resize(fyne.NewSize(padSize, padSize))
			pad.Move(fyne.NewPos(chipX-padSize/2, chipY+float32(i)*padSpacingV-padSize/2))
			objects = append(objects, pad)
			pad2 := canvas.NewRectangle(padColor)
			pad2.Resize(fyne.NewSize(padSize, padSize))
			pad2.Move(fyne.NewPos(chipX+chipW-padSize/2, chipY+float32(i)*padSpacingV-padSize/2))
			objects = append(objects, pad2)
		}

		// Internal array
		arrayColor := color.RGBA{0, 180, 220, 200}
		arraySize := 20 * scaleX
		arrayCells := 3
		cellSz := arraySize / float32(arrayCells)
		arrayX := centerX - arraySize/2
		arrayY := centerY - arraySize/2
		for row := 0; row < arrayCells; row++ {
			for col := 0; col < arrayCells; col++ {
				cell := canvas.NewRectangle(arrayColor)
				cell.Resize(fyne.NewSize(cellSz-1, cellSz-1))
				cell.Move(fyne.NewPos(arrayX+float32(col)*cellSz, arrayY+float32(row)*cellSz))
				objects = append(objects, cell)
			}
		}
	}

	return objects
}

func (r *demoCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *demoCardRenderer) Destroy() {}

// ResponsiveHeader is a custom widget that adapts text size based on width
type ResponsiveHeader struct {
	widget.BaseWidget
	titleText    *canvas.Text
	subtitleText *canvas.Text
	taglineText  *canvas.Text
	separator    *widget.Separator
	cache        sharedwidgets.LayoutCache
}

// NewResponsiveHeader creates a header that scales with screen width
func NewResponsiveHeader() *ResponsiveHeader {
	h := &ResponsiveHeader{
		titleText:    canvas.NewText("FeCIM Lattice Tools", color.RGBA{255, 255, 255, 255}),
		subtitleText: canvas.NewText("Ferroelectric Compute-in-Memory Educational Suite", color.RGBA{0, 212, 255, 255}),
		taglineText:  canvas.NewText("\"Compute in memory where the same device does the memory and the computation.\" — Dr. external research group", color.RGBA{180, 200, 220, 200}),
		separator:    widget.NewSeparator(),
	}
	h.titleText.TextStyle = fyne.TextStyle{Bold: true}
	h.taglineText.TextStyle = fyne.TextStyle{Italic: true}
	h.ExtendBaseWidget(h)
	return h
}

func (h *ResponsiveHeader) CreateRenderer() fyne.WidgetRenderer {
	return &responsiveHeaderRenderer{header: h}
}

type responsiveHeaderRenderer struct {
	header  *ResponsiveHeader
	objects []fyne.CanvasObject
}

func (r *responsiveHeaderRenderer) MinSize() fyne.Size {
	return fyne.NewSize(280, 80)
}

func (r *responsiveHeaderRenderer) Layout(size fyne.Size) {
	if !r.header.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *responsiveHeaderRenderer) layoutWithSize(size fyne.Size) {
	h := r.header
	bp := sharedwidgets.GetBreakpoint(size.Width)

	// Scale text sizes based on breakpoint
	var titleSize, subtitleSize, taglineSize float32
	showTagline := true

	switch bp {
	case sharedwidgets.BreakpointSM:
		titleSize, subtitleSize, taglineSize = 22, 14, 11
		showTagline = false // Hide tagline on mobile
	case sharedwidgets.BreakpointMD:
		titleSize, subtitleSize, taglineSize = 24, 15, 12
	case sharedwidgets.BreakpointLG:
		titleSize, subtitleSize, taglineSize = 26, 15, 12
	default: // XL
		titleSize, subtitleSize, taglineSize = 28, 16, 13
	}

	h.titleText.TextSize = titleSize
	h.subtitleText.TextSize = subtitleSize
	h.taglineText.TextSize = taglineSize

	// Calculate positions (centered)
	padding := float32(8)
	y := padding

	// Title
	titleWidth := h.titleText.MinSize().Width
	h.titleText.Move(fyne.NewPos((size.Width-titleWidth)/2, y))
	y += titleSize + 4

	// Subtitle
	subtitleWidth := h.subtitleText.MinSize().Width
	h.subtitleText.Move(fyne.NewPos((size.Width-subtitleWidth)/2, y))
	y += subtitleSize + 4

	// Tagline (conditional)
	if showTagline {
		taglineWidth := h.taglineText.MinSize().Width
		// Truncate if too wide
		if taglineWidth > size.Width-20 {
			h.taglineText.Move(fyne.NewPos(10, y))
		} else {
			h.taglineText.Move(fyne.NewPos((size.Width-taglineWidth)/2, y))
		}
		y += taglineSize + 8
	}

	// Separator at bottom
	h.separator.Resize(fyne.NewSize(size.Width, 2))
	h.separator.Move(fyne.NewPos(0, size.Height-2))

	// Build objects list
	r.objects = []fyne.CanvasObject{h.titleText, h.subtitleText}
	if showTagline {
		r.objects = append(r.objects, h.taglineText)
	}
	r.objects = append(r.objects, h.separator)

	r.header.cache.MarkLayout(size)
}

func (r *responsiveHeaderRenderer) Refresh() {
	r.layoutWithSize(r.header.Size())
}

func (r *responsiveHeaderRenderer) Objects() []fyne.CanvasObject {
	if len(r.objects) == 0 {
		// Initial objects before first layout
		return []fyne.CanvasObject{r.header.titleText, r.header.subtitleText, r.header.taglineText, r.header.separator}
	}
	return r.objects
}

func (r *responsiveHeaderRenderer) Destroy() {}

// ResponsiveFooter adapts footer content based on screen width
type ResponsiveFooter struct {
	widget.BaseWidget
	metricsText *canvas.Text
	journeyText *canvas.Text
	separator   *widget.Separator
	cache       sharedwidgets.LayoutCache
}

// NewResponsiveFooter creates a footer that simplifies on small screens
func NewResponsiveFooter() *ResponsiveFooter {
	f := &ResponsiveFooter{
		metricsText: canvas.NewText("30 Analog States  |  87% MNIST  |  100× Efficiency  |  10⁹ Cycles  |  TRL 4", color.RGBA{0, 212, 255, 230}),
		journeyText: canvas.NewText("1. Physics → 2. Compute → 3. Application → 4. System → 5. Business → 6. Design", color.RGBA{150, 170, 190, 200}),
		separator:   widget.NewSeparator(),
	}
	f.metricsText.Alignment = fyne.TextAlignCenter
	f.journeyText.Alignment = fyne.TextAlignCenter
	f.ExtendBaseWidget(f)
	return f
}

func (f *ResponsiveFooter) CreateRenderer() fyne.WidgetRenderer {
	return &responsiveFooterRenderer{footer: f}
}

type responsiveFooterRenderer struct {
	footer  *ResponsiveFooter
	objects []fyne.CanvasObject
}

func (r *responsiveFooterRenderer) MinSize() fyne.Size {
	return fyne.NewSize(280, 40)
}

func (r *responsiveFooterRenderer) Layout(size fyne.Size) {
	if !r.footer.cache.ShouldLayout(size) {
		return
	}
	r.layoutWithSize(size)
}

func (r *responsiveFooterRenderer) layoutWithSize(size fyne.Size) {
	f := r.footer
	bp := sharedwidgets.GetBreakpoint(size.Width)

	showJourney := bp != sharedwidgets.BreakpointSM // Hide journey on mobile

	// Adjust text sizes and content based on breakpoint
	var metricsSize, journeySize float32
	switch bp {
	case sharedwidgets.BreakpointSM:
		metricsSize = 11
		// Shorter metrics text for mobile
		f.metricsText.Text = "30 States | 87% MNIST | 100× Eff"
	case sharedwidgets.BreakpointMD:
		metricsSize, journeySize = 12, 11
		f.metricsText.Text = "30 States | 87% MNIST | 100× Efficiency | TRL 4"
	default:
		metricsSize, journeySize = 13, 12
		f.metricsText.Text = "30 Analog States  |  87% MNIST  |  100× Efficiency  |  10⁹ Cycles  |  TRL 4"
	}

	f.metricsText.TextSize = metricsSize
	f.journeyText.TextSize = journeySize

	// Separator at top
	f.separator.Resize(fyne.NewSize(size.Width, 2))
	f.separator.Move(fyne.NewPos(0, 0))

	// Position metrics text
	y := float32(8)
	metricsWidth := f.metricsText.MinSize().Width
	if metricsWidth > size.Width-20 {
		f.metricsText.Move(fyne.NewPos(10, y))
	} else {
		f.metricsText.Move(fyne.NewPos((size.Width-metricsWidth)/2, y))
	}
	y += metricsSize + 4

	// Journey text (conditional)
	r.objects = []fyne.CanvasObject{f.separator, f.metricsText}
	if showJourney {
		journeyWidth := f.journeyText.MinSize().Width
		if journeyWidth > size.Width-20 {
			f.journeyText.Move(fyne.NewPos(10, y))
		} else {
			f.journeyText.Move(fyne.NewPos((size.Width-journeyWidth)/2, y))
		}
		r.objects = append(r.objects, f.journeyText)
	}

	r.footer.cache.MarkLayout(size)
}

func (r *responsiveFooterRenderer) Refresh() {
	r.layoutWithSize(r.footer.Size())
}

func (r *responsiveFooterRenderer) Objects() []fyne.CanvasObject {
	if len(r.objects) == 0 {
		return []fyne.CanvasObject{r.footer.separator, r.footer.metricsText, r.footer.journeyText}
	}
	return r.objects
}

func (r *responsiveFooterRenderer) Destroy() {}

// CreateLauncherContent creates the launcher tab content
func CreateLauncherContent(onDemoSelected func(demoNum int)) fyne.CanvasObject {
	demos := GetDemos()

	// Create demo cards
	cards := make([]fyne.CanvasObject, len(demos))
	for i, demo := range demos {
		d := demo // Capture for closure
		cards[i] = NewDemoCard(d, func() {
			if onDemoSelected != nil {
				onDemoSelected(d.Number)
			}
		})
	}

	// Create responsive header that scales with screen width
	header := NewResponsiveHeader()

	// Responsive grid layout - adapts columns AND card height per breakpoint
	gridLayout := sharedwidgets.NewResponsiveGridLayout()
	grid := container.New(gridLayout, cards...)

	// Wrap grid in VScroll for small screens where content overflows
	scrollableGrid := container.NewVScroll(container.NewPadded(grid))

	// Create responsive footer that simplifies on small screens
	footer := NewResponsiveFooter()

	// Use border layout with header and footer
	return container.NewBorder(
		container.NewPadded(header),
		container.NewPadded(footer),
		nil, nil,
		scrollableGrid,
	)
}
