// Package main provides a unified launcher for all FeCIM demos.
package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// FeCIM theme colors
var (
	colorBackground  = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary     = color.RGBA{0, 212, 255, 255} // Cyan
	colorCardBg      = color.RGBA{0, 40, 80, 255}   // Darker blue for cards
	colorDisabled    = color.RGBA{80, 80, 80, 255}  // Gray for coming soon
	colorReady       = color.RGBA{0, 200, 100, 255} // Green for ready
	colorComingSoon  = color.RGBA{150, 150, 150, 255}
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
	case theme.ColorNameDisabled:
		return colorDisabled
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Demo represents a demo application
type Demo struct {
	Number      int
	Title       string
	Description string
	Ready       bool
	BinaryPath  string // Relative path from project root
	BuildCmd    string // Build command if binary doesn't exist
}

var demos = []Demo{
	{
		Number:      1,
		Title:       "Memory Cell",
		Description: "P-E hysteresis curve\n30 discrete polarization levels",
		Ready:       true,
		BinaryPath:  "demo1-hysteresis/hysteresis",
		BuildCmd:    "demo1-hysteresis/cmd/hysteresis",
	},
	{
		Number:      2,
		Title:       "Crossbar MVM",
		Description: "Matrix-vector multiply\nIR drop & sneak paths",
		Ready:       true,
		BinaryPath:  "demo2-crossbar/crossbar-gui",
		BuildCmd:    "demo2-crossbar/cmd/crossbar-gui",
	},
	{
		Number:      3,
		Title:       "MNIST 87%",
		Description: "Handwritten digit recognition\nReal neural network",
		Ready:       true,
		BinaryPath:  "demo3-mnist/mnist-gui",
		BuildCmd:    "demo3-mnist/cmd/mnist-gui",
	},
	{
		Number:      4,
		Title:       "Circuits",
		Description: "DAC, ADC, TIA\nCharge pump peripherals",
		Ready:       true,
		BinaryPath:  "demo4-circuits/circuits-gui",
		BuildCmd:    "demo4-circuits/cmd/circuits-gui",
	},
	{
		Number:      5,
		Title:       "Thermal",
		Description: "Heat dissipation\nThermal management",
		Ready:       false,
		BinaryPath:  "demo5-thermal/thermal",
		BuildCmd:    "demo5-thermal/cmd/thermal",
	},
	{
		Number:      6,
		Title:       "3D Stack",
		Description: "Multi-layer architecture\n3D integration",
		Ready:       false,
		BinaryPath:  "demo6-multilayer/multilayer",
		BuildCmd:    "demo6-multilayer/cmd/multilayer",
	},
	{
		Number:      7,
		Title:       "Non-Idealities",
		Description: "Device variations\nEndurance & retention",
		Ready:       false,
		BinaryPath:  "demo7-nonidealities/nonidealities",
		BuildCmd:    "demo7-nonidealities/cmd/nonidealities",
	},
	{
		Number:      8,
		Title:       "Why FeCIM?",
		Description: "Architecture comparison\nFeCIM vs GPU/TPU/ASIC",
		Ready:       true,
		BinaryPath:  "demo8-comparison/comparison-gui",
		BuildCmd:    "demo8-comparison/cmd/comparison-gui",
	},
}

func main() {
	a := app.NewWithID("com.fecim.launcher")
	a.Settings().SetTheme(&feCIMTheme{})

	w := a.NewWindow("FeCIM Demo Suite")
	w.Resize(fyne.NewSize(900, 700))

	// Get project root directory
	projectRoot := getProjectRoot()

	// Header with title and quote
	header := createHeader()

	// Demo grid
	grid := createDemoGrid(projectRoot)

	// Footer with progress
	footer := createFooter()

	// Main layout
	content := container.NewBorder(
		header,  // top
		footer,  // bottom
		nil,     // left
		nil,     // right
		grid,    // center
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func getProjectRoot() string {
	// Try to find project root by looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}

func createHeader() fyne.CanvasObject {
	// Title
	title := canvas.NewText("Multilayer Ferroelectric CIM Visualizer", color.White)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Dr. Tour quote
	quote := widget.NewLabel("\"Compute in memory where the same device does memory and computation.\"")
	quote.Alignment = fyne.TextAlignCenter
	quote.TextStyle = fyne.TextStyle{Italic: true}

	attribution := widget.NewLabel("- Dr. external research group")
	attribution.Alignment = fyne.TextAlignCenter

	// Separator
	separator := canvas.NewRectangle(colorPrimary)
	separator.SetMinSize(fyne.NewSize(0, 2))

	return container.NewVBox(
		layout.NewSpacer(),
		title,
		quote,
		attribution,
		layout.NewSpacer(),
		separator,
		layout.NewSpacer(),
	)
}

func createDemoGrid(projectRoot string) fyne.CanvasObject {
	// Create cards for each demo
	var cards []fyne.CanvasObject
	for _, demo := range demos {
		cards = append(cards, createDemoCard(demo, projectRoot))
	}

	// 4 columns grid (2 rows of 4)
	grid := container.NewGridWithColumns(4, cards...)

	return container.NewPadded(grid)
}

func createDemoCard(demo Demo, projectRoot string) fyne.CanvasObject {
	// Demo number
	numText := canvas.NewText(fmt.Sprintf("DEMO %d", demo.Number), colorPrimary)
	numText.TextSize = 14
	numText.TextStyle = fyne.TextStyle{Bold: true}
	numText.Alignment = fyne.TextAlignCenter

	// Title
	titleText := canvas.NewText(demo.Title, color.White)
	titleText.TextSize = 16
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Alignment = fyne.TextAlignCenter

	// Description
	descLabel := widget.NewLabel(demo.Description)
	descLabel.Alignment = fyne.TextAlignCenter
	descLabel.Wrapping = fyne.TextWrapWord

	// Status indicator
	var statusText *canvas.Text
	if demo.Ready {
		statusText = canvas.NewText("READY", colorReady)
	} else {
		statusText = canvas.NewText("COMING SOON", colorComingSoon)
	}
	statusText.TextSize = 12
	statusText.TextStyle = fyne.TextStyle{Bold: true}
	statusText.Alignment = fyne.TextAlignCenter

	// Launch button
	var launchBtn *widget.Button
	if demo.Ready {
		launchBtn = widget.NewButton("Launch", func() {
			launchDemo(demo, projectRoot)
		})
	} else {
		launchBtn = widget.NewButton("Coming Soon", nil)
		launchBtn.Disable()
	}

	// Card content
	cardContent := container.NewVBox(
		layout.NewSpacer(),
		numText,
		titleText,
		descLabel,
		statusText,
		launchBtn,
		layout.NewSpacer(),
	)

	// Card background
	bg := canvas.NewRectangle(colorCardBg)
	bg.CornerRadius = 8

	// Combine background and content
	card := container.NewStack(bg, container.NewPadded(cardContent))
	card.Resize(fyne.NewSize(200, 180))

	return card
}

func launchDemo(demo Demo, projectRoot string) {
	binaryPath := filepath.Join(projectRoot, demo.BinaryPath)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to build it
		fmt.Printf("Building %s...\n", demo.Title)
		buildPath := filepath.Join(projectRoot, demo.BuildCmd)
		cmd := exec.Command("go", "build", "-o", binaryPath, buildPath)
		cmd.Dir = projectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to build demo %d: %v\n", demo.Number, err)
			return
		}
	}

	// Launch the demo
	fmt.Printf("Launching Demo %d: %s\n", demo.Number, demo.Title)
	cmd := exec.Command(binaryPath)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to launch demo %d: %v\n", demo.Number, err)
		return
	}
}

func createFooter() fyne.CanvasObject {
	// Count ready demos
	readyCount := 0
	for _, demo := range demos {
		if demo.Ready {
			readyCount++
		}
	}

	// Separator
	separator := canvas.NewRectangle(colorPrimary)
	separator.SetMinSize(fyne.NewSize(0, 2))

	// Progress text
	progressText := widget.NewLabel(fmt.Sprintf("%d/%d demos ready", readyCount, len(demos)))
	progressText.Alignment = fyne.TextAlignCenter

	// Progress bar
	progress := widget.NewProgressBar()
	progress.SetValue(float64(readyCount) / float64(len(demos)))
	progress.TextFormatter = func() string {
		return fmt.Sprintf("%.0f%% Complete", float64(readyCount)/float64(len(demos))*100)
	}

	// GitHub link
	githubLabel := widget.NewLabel("github.com/jtamez/multilayer-fecim-vis")
	githubLabel.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		separator,
		layout.NewSpacer(),
		progressText,
		progress,
		githubLabel,
		layout.NewSpacer(),
	)
}
