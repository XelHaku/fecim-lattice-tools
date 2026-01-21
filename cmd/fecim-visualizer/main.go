// Command fecim-visualizer provides a unified GUI application with all FeCIM demos as tabs.
//
// This is the main entry point for the FeCIM Visualization Suite.
// It combines all individual demos into a single application with tab navigation.
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	demo1gui "multilayer-ferroelectric-cim-visualizer/demo1-hysteresis/pkg/gui"
	demo2gui "multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/gui"
	demo3gui "multilayer-ferroelectric-cim-visualizer/demo3-mnist/pkg/gui"
	demo4gui "multilayer-ferroelectric-cim-visualizer/demo4-circuits/pkg/gui"
	demo5gui "multilayer-ferroelectric-cim-visualizer/demo5-thermal/pkg/gui"
	demo6gui "multilayer-ferroelectric-cim-visualizer/demo6-multilayer/pkg/gui"
	demo7gui "multilayer-ferroelectric-cim-visualizer/demo7-nonidealities/pkg/gui"
	demo8gui "multilayer-ferroelectric-cim-visualizer/demo8-comparison/pkg/gui"
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
	demo4 *demo4gui.EmbeddedCircuitsApp
	demo5 *demo5gui.EmbeddedThermalApp
	demo6 *demo6gui.EmbeddedMultilayerApp
	demo7 *demo7gui.EmbeddedNonIdealitiesApp
	demo8 *demo8gui.EmbeddedComparisonApp
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
		demo4: demo4gui.NewEmbeddedCircuitsApp(),
		demo5: demo5gui.NewEmbeddedThermalApp(),
		demo6: demo6gui.NewEmbeddedMultilayerApp(),
		demo7: demo7gui.NewEmbeddedNonIdealitiesApp(),
		demo8: demo8gui.NewEmbeddedComparisonApp(),
	}

	// Create tabs container (will be populated below)
	var tabs *container.AppTabs

	// Create launcher content with callback to switch tabs
	launcherContent := CreateLauncherContent(func(demoNum int) {
		if tabs != nil && demoNum >= 1 && demoNum <= 8 {
			tabs.SelectIndex(demoNum) // Demo 1 is at index 1, etc.
		}
	})

	// Build content for each demo
	demo1Content := demos.demo1.BuildContent(fyneApp, window)
	demo2Content := demos.demo2.BuildContent(fyneApp, window)
	demo3Content := demos.demo3.BuildContent(fyneApp, window)
	demo4Content := demos.demo4.BuildContent(fyneApp, window)
	demo5Content := demos.demo5.BuildContent(fyneApp, window)
	demo6Content := demos.demo6.BuildContent(fyneApp, window)
	demo7Content := demos.demo7.BuildContent(fyneApp, window)
	demo8Content := demos.demo8.BuildContent(fyneApp, window)

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

	// Track current demo for start/stop
	currentDemo := 0

	// Handle tab changes - start/stop simulations as needed
	tabs.OnSelected = func(tab *container.TabItem) {
		// Stop previous demo
		switch currentDemo {
		case 1:
			demos.demo1.Stop()
		case 2:
			demos.demo2.Stop()
		case 3:
			demos.demo3.Stop()
		case 4:
			demos.demo4.Stop()
		case 5:
			demos.demo5.Stop()
		case 6:
			demos.demo6.Stop()
		case 7:
			demos.demo7.Stop()
		case 8:
			demos.demo8.Stop()
		}

		// Start new demo
		switch tab.Text {
		case "1. Hysteresis":
			currentDemo = 1
			demos.demo1.Start()
		case "2. Crossbar":
			currentDemo = 2
			demos.demo2.Start()
		case "3. MNIST":
			currentDemo = 3
			demos.demo3.Start()
		case "4. Circuits":
			currentDemo = 4
			demos.demo4.Start()
		case "5. Thermal":
			currentDemo = 5
			demos.demo5.Start()
		case "6. 3D Stack":
			currentDemo = 6
			demos.demo6.Start()
		case "7. Non-Idealities":
			currentDemo = 7
			demos.demo7.Start()
		case "8. Comparison":
			currentDemo = 8
			demos.demo8.Start()
		default:
			currentDemo = 0
		}
	}

	// Set window content
	window.SetContent(tabs)

	// Run the application
	window.ShowAndRun()

	// Cleanup all demos on exit
	demos.demo1.Stop()
	demos.demo2.Stop()
	demos.demo3.Stop()
	demos.demo4.Stop()
	demos.demo5.Stop()
	demos.demo6.Stop()
	demos.demo7.Stop()
	demos.demo8.Stop()
}
