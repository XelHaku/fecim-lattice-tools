// Command fecim-visualizer provides a unified GUI application with all FeCIM demos as tabs.
//
// This is the main entry point for the FeCIM Visualization Suite.
// It combines all individual demos into a single application with tab navigation.
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	demo1gui "multilayer-ferroelectric-cim-visualizer/demo1-hysteresis/pkg/gui"
	demo2gui "multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/gui"
	demo3gui "multilayer-ferroelectric-cim-visualizer/demo3-mnist/pkg/gui"
)

// FeCIM theme colors
var (
	colorBackground = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
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

// DemoApp holds the demo instances
type DemoApp struct {
	demo1 *demo1gui.EmbeddedApp
	demo2 *demo2gui.EmbeddedCrossbarApp
	demo3 *demo3gui.EmbeddedMNISTApp
}

// createComingSoonTab creates a placeholder tab for demos not yet ready
func createComingSoonTab(demoNum int, title, description string) fyne.CanvasObject {
	// Background
	bg := canvas.NewRectangle(color.RGBA{0, 40, 80, 255})

	// Title
	titleLabel := widget.NewLabelWithStyle(
		title,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Coming soon message
	comingSoonLabel := widget.NewLabelWithStyle(
		"COMING SOON",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Description
	descLabel := widget.NewLabel(description)
	descLabel.Alignment = fyne.TextAlignCenter
	descLabel.Wrapping = fyne.TextWrapWord

	// Icon placeholder (large number)
	numLabel := canvas.NewText(string('0'+byte(demoNum)), color.RGBA{0, 212, 255, 100})
	numLabel.TextSize = 120
	numLabel.TextStyle = fyne.TextStyle{Bold: true}

	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(numLabel),
		titleLabel,
		comingSoonLabel,
		widget.NewSeparator(),
		descLabel,
		layout.NewSpacer(),
	)

	return container.NewStack(bg, container.NewCenter(content))
}

func main() {
	// Create Fyne app
	fyneApp := app.NewWithID("com.fecim.visualizer")
	fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Create main window
	window := fyneApp.NewWindow("FeCIM Visualization Suite")
	window.Resize(fyne.NewSize(1400, 900))

	// Create demo instances
	demos := &DemoApp{
		demo1: demo1gui.NewEmbeddedApp(),
		demo2: demo2gui.NewEmbeddedCrossbarApp(),
		demo3: demo3gui.NewEmbeddedMNISTApp(),
	}

	// Create tabs container (will be populated below)
	var tabs *container.AppTabs

	// Create launcher content with callback to switch tabs
	launcherContent := CreateLauncherContent(func(demoNum int) {
		if tabs != nil && demoNum >= 1 && demoNum <= 3 {
			tabs.SelectIndex(demoNum) // Demo 1 is at index 1, etc.
		}
	})

	// Build content for each ready demo
	demo1Content := demos.demo1.BuildContent(fyneApp, window)
	demo2Content := demos.demo2.BuildContent(fyneApp, window)
	demo3Content := demos.demo3.BuildContent(fyneApp, window)

	// Create coming soon placeholders for demos 4-8
	demo4Content := createComingSoonTab(4, "Demo 4: Peripheral Circuits",
		"DAC, ADC, and signal conditioning circuits for real chip integration.\n\nHow it fits in a real chip.")
	demo5Content := createComingSoonTab(5, "Demo 5: Thermal Analysis",
		"Compare thermal profiles vs NAND and DRAM.\n\n1000× cooler than the competition.")
	demo6Content := createComingSoonTab(6, "Demo 6: 3D Stack",
		"Multi-layer stacking for massive parallelism.\n\nScalable architecture for the future.")
	demo7Content := createComingSoonTab(7, "Demo 7: Non-Idealities",
		"IR drop, sneak paths, and drift analysis.\n\nReal-world challenges and solutions.")
	demo8Content := createComingSoonTab(8, "Demo 8: Technology Comparison",
		"Energy, speed, and cost comparison with verified sources.\n\nWhy FeCIM wins.")

	// Create tabs
	tabs = container.NewAppTabs(
		container.NewTabItem("Home", launcherContent),
		container.NewTabItem("1. Hysteresis", container.NewMax(demo1Content)),
		container.NewTabItem("2. Crossbar", container.NewMax(demo2Content)),
		container.NewTabItem("3. MNIST", container.NewMax(demo3Content)),
		container.NewTabItem("4. Circuits", container.NewMax(demo4Content)),
		container.NewTabItem("5. Thermal", container.NewMax(demo5Content)),
		container.NewTabItem("6. 3D Stack", container.NewMax(demo6Content)),
		container.NewTabItem("7. Non-Idealities", container.NewMax(demo7Content)),
		container.NewTabItem("8. Comparison", container.NewMax(demo8Content)),
	)

	// Handle tab changes - start/stop simulations as needed
	tabs.OnSelected = func(tab *container.TabItem) {
		// Start Demo 1 simulation only when its tab is selected
		switch tab.Text {
		case "1. Hysteresis":
			demos.demo1.Start()
		default:
			demos.demo1.Stop()
		}
	}

	// Set window content
	window.SetContent(tabs)

	// Run the application
	window.ShowAndRun()

	// Cleanup
	demos.demo1.Stop()
	demos.demo2.Stop()
	demos.demo3.Stop()
}
