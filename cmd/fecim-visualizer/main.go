// Command fecim-visualizer provides a unified GUI application with all FeCIM demos as tabs.
//
// This is the main entry point for the FeCIM Visualization Suite.
// It combines all 5 demos into a single application with tab navigation.
//
// The 5-Demo Story:
//   Demo 1: The Memory Cell (Hysteresis) - How the cell works
//   Demo 2: The Crossbar Computer (MVM + Non-Idealities) - How we compute
//   Demo 3: The AI Brain (MNIST) - What we can build
//   Demo 4: The Chip System (Circuits) - How it fits in a chip
//   Demo 5: Why FeCIM Wins (Comparison) - The business case
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
	demo5gui "multilayer-ferroelectric-cim-visualizer/demo8-comparison/pkg/gui"
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
	demo1  *demo1gui.EmbeddedApp                // Hysteresis
	demo2  *demo2gui.EmbeddedTabbedCrossbarApp  // Crossbar + Non-Idealities (merged)
	demo3  *demo3gui.EmbeddedMNISTApp           // MNIST (basic)
	demo3b *demo3gui.EmbeddedDualModeApp        // MNIST FP vs CIM
	demo4  *demo4gui.EmbeddedCircuitsApp        // Circuits
	demo5  *demo5gui.EmbeddedComparisonApp      // Comparison (technical briefing)
}

func main() {
	// Create Fyne app
	fyneApp := app.NewWithID("com.fecim.visualizer")
	fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Create main window
	window := fyneApp.NewWindow("FeCIM Visualization Suite - 5 World-Class Demos")
	window.Resize(fyne.NewSize(1400, 900))

	// Create demo instances
	demos := &DemoApp{
		demo1:  demo1gui.NewEmbeddedApp(),
		demo2:  demo2gui.NewEmbeddedTabbedCrossbarApp(), // New tabbed version with non-idealities
		demo3:  demo3gui.NewEmbeddedMNISTApp(),
		demo3b: demo3gui.NewEmbeddedDualModeApp(),
		demo4:  demo4gui.NewEmbeddedCircuitsApp(),
		demo5:  demo5gui.NewEmbeddedComparisonApp(),
	}

	// Create tabs container (will be populated below)
	var tabs *container.AppTabs

	// Create launcher content with callback to switch tabs
	launcherContent := CreateLauncherContent(func(demoNum int) {
		if tabs != nil {
			// Map demo number to tab index
			// Home=0, Demo1=1, Demo2=2, Demo3=3, Demo3b=4, Demo4=5, Demo5=6
			tabIndex := 0
			switch demoNum {
			case 1:
				tabIndex = 1
			case 2:
				tabIndex = 2
			case 3:
				tabIndex = 3
			case 4:
				tabIndex = 5
			case 5:
				tabIndex = 6
			}
			tabs.SelectIndex(tabIndex)
		}
	})

	// Build content for each demo
	demo1Content := demos.demo1.BuildContent(fyneApp, window)
	demo2Content := demos.demo2.BuildContent(fyneApp, window)
	demo3Content := demos.demo3.BuildContent(fyneApp, window)
	demo3bContent := demos.demo3b.BuildContent(fyneApp, window)
	demo4Content := demos.demo4.BuildContent(fyneApp, window)
	demo5Content := demos.demo5.BuildContent(fyneApp, window)

	// Create tabs - 5 demos total (plus home and MNIST variant)
	tabs = container.NewAppTabs(
		container.NewTabItem("Home", launcherContent),
		container.NewTabItem("1. Hysteresis", container.NewMax(demo1Content)),
		container.NewTabItem("2. Crossbar+", container.NewMax(demo2Content)),
		container.NewTabItem("3. MNIST", container.NewMax(demo3Content)),
		container.NewTabItem("3b. MNIST FP/CIM", container.NewMax(demo3bContent)),
		container.NewTabItem("4. Circuits", container.NewMax(demo4Content)),
		container.NewTabItem("5. Comparison", container.NewMax(demo5Content)),
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
		case 31: // 3b
			demos.demo3b.Stop()
		case 4:
			demos.demo4.Stop()
		case 5:
			demos.demo5.Stop()
		}

		// Start new demo
		switch tab.Text {
		case "1. Hysteresis":
			currentDemo = 1
			demos.demo1.Start()
		case "2. Crossbar+":
			currentDemo = 2
			demos.demo2.Start()
		case "3. MNIST":
			currentDemo = 3
			demos.demo3.Start()
		case "3b. MNIST FP/CIM":
			currentDemo = 31
			demos.demo3b.Start()
		case "4. Circuits":
			currentDemo = 4
			demos.demo4.Start()
		case "5. Comparison":
			currentDemo = 5
			demos.demo5.Start()
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
	demos.demo3b.Stop()
	demos.demo4.Stop()
	demos.demo5.Stop()
}
