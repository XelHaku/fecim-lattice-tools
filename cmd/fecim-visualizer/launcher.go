package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// DemoInfo holds information about a demo
type DemoInfo struct {
	Number      int
	Title       string
	Subtitle    string
	Description string
	Ready       bool
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
		minSize:  fyne.NewSize(280, 180),
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
	r.Refresh()
}

func (r *demoCardRenderer) Refresh() {
	r.objects = r.objects[:0]
	size := r.card.Size()
	info := r.card.info

	// Background color based on ready state
	var bgColor, borderColor, textColor color.RGBA
	if info.Ready {
		bgColor = color.RGBA{0, 60, 120, 255}      // Darker blue for ready
		borderColor = color.RGBA{0, 212, 255, 255} // Cyan border
		textColor = color.RGBA{255, 255, 255, 255}
	} else {
		bgColor = color.RGBA{30, 40, 50, 200} // Gray for coming soon
		borderColor = color.RGBA{80, 90, 100, 255}
		textColor = color.RGBA{120, 130, 140, 255}
	}

	// Border
	border := canvas.NewRectangle(borderColor)
	border.Resize(size)
	r.objects = append(r.objects, border)

	// Background
	padding := float32(2)
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(size.Width-padding*2, size.Height-padding*2))
	bg.Move(fyne.NewPos(padding, padding))
	r.objects = append(r.objects, bg)

	// Demo number badge
	badgeSize := float32(36)
	badgeX := float32(12)
	badgeY := float32(12)

	badgeBg := canvas.NewCircle(borderColor)
	badgeBg.Resize(fyne.NewSize(badgeSize, badgeSize))
	badgeBg.Move(fyne.NewPos(badgeX, badgeY))
	r.objects = append(r.objects, badgeBg)

	numText := canvas.NewText(string('0'+byte(info.Number)), bgColor)
	numText.TextSize = 20
	numText.TextStyle = fyne.TextStyle{Bold: true}
	numText.Move(fyne.NewPos(badgeX+12, badgeY+6))
	r.objects = append(r.objects, numText)

	// Title
	title := canvas.NewText(info.Title, textColor)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(badgeX+badgeSize+12, 14))
	r.objects = append(r.objects, title)

	// Subtitle
	subtitle := canvas.NewText(info.Subtitle, color.RGBA{textColor.R, textColor.G, textColor.B, 180})
	subtitle.TextSize = 12
	subtitle.Move(fyne.NewPos(badgeX+badgeSize+12, 36))
	r.objects = append(r.objects, subtitle)

	// Status badge
	var statusText string
	var statusColor color.RGBA
	if info.Ready {
		statusText = "READY"
		statusColor = color.RGBA{0, 200, 100, 255}
	} else {
		statusText = "COMING SOON"
		statusColor = color.RGBA{150, 150, 150, 255}
	}
	status := canvas.NewText(statusText, statusColor)
	status.TextSize = 10
	status.TextStyle = fyne.TextStyle{Bold: true}
	status.Move(fyne.NewPos(size.Width-80, 16))
	r.objects = append(r.objects, status)

	// Description (wrapped manually)
	desc := info.Description
	descColor := color.RGBA{textColor.R, textColor.G, textColor.B, 200}

	// Simple word wrap - split into lines
	maxWidth := size.Width - 24
	lineY := float32(60)
	lineHeight := float32(16)

	words := splitWords(desc)
	line := ""
	for _, word := range words {
		testLine := line + word + " "
		testText := canvas.NewText(testLine, descColor)
		testText.TextSize = 12
		if testText.MinSize().Width > maxWidth && line != "" {
			// Write current line
			lineText := canvas.NewText(line, descColor)
			lineText.TextSize = 12
			lineText.Move(fyne.NewPos(12, lineY))
			r.objects = append(r.objects, lineText)
			lineY += lineHeight
			line = word + " "
		} else {
			line = testLine
		}
	}
	if line != "" {
		lineText := canvas.NewText(line, descColor)
		lineText.TextSize = 12
		lineText.Move(fyne.NewPos(12, lineY))
		r.objects = append(r.objects, lineText)
	}

	// Hover hint for ready cards
	if info.Ready {
		hint := canvas.NewText("Click to open", color.RGBA{0, 212, 255, 150})
		hint.TextSize = 10
		hint.Move(fyne.NewPos(size.Width-80, size.Height-20))
		r.objects = append(r.objects, hint)
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

	// Title
	titleLabel := widget.NewLabelWithStyle(
		"FeCIM Visualization Suite",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Subtitle with narrative
	subtitleLabel := widget.NewLabelWithStyle(
		"6 Demos: Physics → Compute → Application → System → Business → Design",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	// Progress indicator
	readyCount := 0
	for _, d := range demos {
		if d.Ready {
			readyCount++
		}
	}
	progressLabel := widget.NewLabelWithStyle(
		"6/6 Demos Ready",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	header := container.NewVBox(
		titleLabel,
		subtitleLabel,
		widget.NewSeparator(),
		progressLabel,
		widget.NewSeparator(),
	)

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

	// Grid layout: 3 columns for 6 demos (2 rows)
	// Row 1: Demo 1, 2, 3
	// Row 2: Demo 4, 5, 6
	row1 := container.New(layout.NewGridLayoutWithColumns(3),
		cards[0], cards[1], cards[2],
	)
	row2 := container.New(layout.NewGridLayoutWithColumns(3),
		cards[3], cards[4], cards[5],
	)

	grid := container.NewVBox(row1, row2)

	// Instructions at bottom
	instructionsLabel := widget.NewLabelWithStyle(
		"Click on any demo card to explore. All 6 demos are fully functional!",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Key metrics
	metricsLabel := widget.NewLabelWithStyle(
		"Key Specs: 30 Levels (4.9 bits) | 87% MNIST Accuracy | 10M× Energy vs NAND | 1000× vs DRAM | TRL 4",
		fyne.TextAlignCenter,
		fyne.TextStyle{Monospace: true},
	)

	// Narrative flow
	narrativeLabel := widget.NewLabelWithStyle(
		"The FeCIM Story: Cell → Crossbar → AI → Chip → Business → Design",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		instructionsLabel,
		metricsLabel,
		narrativeLabel,
	)

	// Center the grid in available space
	centeredGrid := container.NewCenter(grid)

	return container.NewBorder(
		header,
		footer,
		nil, nil,
		centeredGrid,
	)
}
