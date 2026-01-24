package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// DemoInfo holds information about a demo
type DemoInfo struct {
	Number      int
	Title       string
	Subtitle    string
	Description string
	Ready       bool
	WIP         bool // Work in progress - show indicator but still accessible
}

// GetDemos returns all demo information (6 consolidated demos)
func GetDemos() []DemoInfo {
	return []DemoInfo{
		{
			Number:      1,
			Title:       "Hysteresis",
			Subtitle:    "The Memory Cell",
			Description: "How the memory cell works: visualize ferroelectric polarization switching with 30 discrete states",
			Ready:       true,
		},
		{
			Number:      2,
			Title:       "Crossbar+",
			Subtitle:    "MVM + Non-Idealities",
			Description: "How we compute in memory: matrix-vector multiplication with IR drop, sneak paths, and drift analysis",
			Ready:       true,
		},
		{
			Number:      3,
			Title:       "MNIST",
			Subtitle:    "The AI Brain",
			Description: "What we can build: draw digits and watch the neural network classify them at 87% accuracy (FP vs CIM)",
			Ready:       true,
		},
		{
			Number:      4,
			Title:       "Circuits",
			Subtitle:    "The Chip System",
			Description: "How it fits in a real chip: DAC, ADC, TIA, and CMOS-compatible peripheral design",
			Ready:       true,
		},
		{
			Number:      5,
			Title:       "Comparison",
			Subtitle:    "Why FeCIM Wins",
			Description: "The business case: energy efficiency, competitive matrix, data center savings calculator",
			Ready:       true,
		},
		{
			Number:      6,
			Title:       "EDA",
			Subtitle:    "Design Suite",
			Description: "Bridge to open-source EDA: weight compiler, layout visualization, SPICE export for ngspice/KLayout",
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
		minSize:  fyne.NewSize(350, 220),
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
}

func (r *demoCardRenderer) MinSize() fyne.Size {
	return r.card.minSize
}

func (r *demoCardRenderer) Layout(size fyne.Size) {
	// Layout positions objects - use the passed size parameter
	sharedwidgets.DebugLayoutCall("demoCardRenderer", size)
	r.layoutWithSize(size)
}

func (r *demoCardRenderer) Refresh() {
	// Refresh uses current widget size
	sharedwidgets.DebugRefreshCall("demoCardRenderer", r.card.Size())
	r.layoutWithSize(r.card.Size())
}

func (r *demoCardRenderer) layoutWithSize(size fyne.Size) {
	// Skip layout with invalid sizes
	if size.Width <= 0 || size.Height <= 0 {
		return
	}

	r.objects = r.objects[:0]
	info := r.card.info

	// Constrain to minimum size to prevent growing
	minSize := r.card.minSize
	if size.Width > minSize.Width {
		size.Width = minSize.Width
	}
	if size.Height > minSize.Height {
		size.Height = minSize.Height
	}

	// Background color based on ready state
	var bgColor, borderColor, textColor, descColor color.RGBA
	if info.Ready {
		bgColor = color.RGBA{0, 60, 120, 255}       // Darker blue for ready
		borderColor = color.RGBA{0, 212, 255, 255}  // Cyan border
		textColor = color.RGBA{255, 255, 255, 255}  // White for titles
		descColor = color.RGBA{220, 235, 255, 255}  // Bright white-blue for descriptions
	} else {
		bgColor = color.RGBA{30, 40, 50, 200}
		borderColor = color.RGBA{80, 90, 100, 255}
		textColor = color.RGBA{120, 130, 140, 255}
		descColor = color.RGBA{100, 110, 120, 255}
	}

	// Border (thicker)
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background
	padding := float32(3)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Scale elements based on card size (base reference: 220px height)
	scale := size.Height / 220.0
	if scale < 1 {
		scale = 1
	}

	// Demo number badge - larger
	badgeSize := float32(48) * scale
	if badgeSize > 70 {
		badgeSize = 70
	}
	badgeX := float32(16)
	badgeY := float32(16)

	badgeBg := canvas.NewCircle(borderColor)
	badgeBg.Resize(fyne.NewSize(badgeSize, badgeSize))
	badgeBg.Move(fyne.NewPos(badgeX, badgeY))
	r.objects = append(r.objects, badgeBg)

	numTextSize := float32(26) * scale
	if numTextSize > 38 {
		numTextSize = 38
	}
	numText := canvas.NewText(string('0'+byte(info.Number)), bgColor)
	numText.TextSize = numTextSize
	numText.TextStyle = fyne.TextStyle{Bold: true}
	numText.Move(fyne.NewPos(badgeX+badgeSize/2-numTextSize/4, badgeY+badgeSize/2-numTextSize/2))
	r.objects = append(r.objects, numText)

	// Title - much larger
	titleSize := float32(24) * scale
	if titleSize > 36 {
		titleSize = 36
	}
	title := canvas.NewText(info.Title, textColor)
	title.TextSize = titleSize
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(badgeX+badgeSize+16, 18))
	r.objects = append(r.objects, title)

	// Subtitle - larger
	subtitleSize := float32(16) * scale
	if subtitleSize > 22 {
		subtitleSize = 22
	}
	subtitle := canvas.NewText(info.Subtitle, color.RGBA{0, 212, 255, 255}) // Cyan for subtitle
	subtitle.TextSize = subtitleSize
	subtitle.Move(fyne.NewPos(badgeX+badgeSize+16, 18+titleSize+6))
	r.objects = append(r.objects, subtitle)

	// Status indicator - WIP badge or green dot
	if info.Ready {
		if info.WIP {
			// Work In Progress badge - orange rounded rectangle with text
			badgeWidth := float32(110)
			badgeHeight := float32(20)
			wipBg := canvas.NewRectangle(color.RGBA{255, 165, 0, 255}) // Orange
			wipBg.Resize(fyne.NewSize(badgeWidth, badgeHeight))
			wipBg.Move(fyne.NewPos(size.Width-badgeWidth-10, 10))
			wipBg.CornerRadius = 4
			r.objects = append(r.objects, wipBg)

			wipText := canvas.NewText("Work In Progress", color.RGBA{0, 0, 0, 255})
			wipText.TextSize = 11
			wipText.TextStyle = fyne.TextStyle{Bold: true}
			wipText.Move(fyne.NewPos(size.Width-badgeWidth-10+6, 13))
			r.objects = append(r.objects, wipText)
		} else {
			// Green dot for ready
			dotSize := float32(12)
			statusDot := canvas.NewCircle(color.RGBA{100, 255, 150, 255})
			statusDot.Resize(fyne.NewSize(dotSize, dotSize))
			statusDot.Move(fyne.NewPos(size.Width-dotSize-12, 12))
			r.objects = append(r.objects, statusDot)
		}
	}

	// Description - larger and clearer
	desc := info.Description
	descSize := float32(15) * scale
	if descSize > 20 {
		descSize = 20
	}
	maxWidth := size.Width - 36
	lineY := badgeY + badgeSize + 20
	lineHeight := descSize + 6

	words := splitWords(desc)
	line := ""
	for _, word := range words {
		testLine := line + word + " "
		testText := canvas.NewText(testLine, descColor)
		testText.TextSize = descSize
		if testText.MinSize().Width > maxWidth && line != "" {
			lineText := canvas.NewText(line, descColor)
			lineText.TextSize = descSize
			lineText.Move(fyne.NewPos(18, lineY))
			r.objects = append(r.objects, lineText)
			lineY += lineHeight
			line = word + " "
		} else {
			line = testLine
		}
	}
	if line != "" {
		lineText := canvas.NewText(line, descColor)
		lineText.TextSize = descSize
		lineText.Move(fyne.NewPos(18, lineY))
		r.objects = append(r.objects, lineText)
	}

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

func (r *demoCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *demoCardRenderer) Destroy() {}

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

	// Use GridWrap layout for dynamic sizing - 3 columns, 2 rows
	// This will expand to fill available space
	grid := container.New(layout.NewGridLayoutWithRows(2),
		container.New(layout.NewGridLayoutWithColumns(3), cards[0], cards[1], cards[2]),
		container.New(layout.NewGridLayoutWithColumns(3), cards[3], cards[4], cards[5]),
	)

	// Compact footer with all info on one line
	metricsText := canvas.NewText("6 Demos: Physics ⇒ Compute ⇒ Application ⇒ System ⇒ Business ⇒ Design | 30 Levels (4.9 bits) | 87% MNIST | 10M× vs NAND | 1000× vs DRAM | TRL 4  —  Click any card to explore", color.RGBA{180, 200, 220, 255})
	metricsText.TextSize = 14
	metricsText.Alignment = fyne.TextAlignCenter

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(metricsText),
	)

	// Use border layout - grid expands to fill center (no header to save space)
	return container.NewBorder(
		nil,    // no header - saves vertical space
		footer, // footer with compact info
		nil, nil,
		container.NewPadded(grid),
	)
}
