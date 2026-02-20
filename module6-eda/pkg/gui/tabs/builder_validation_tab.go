// pkg/gui/tabs/builder_validation_tab.go
// Unified Builder & Validation Tab - consolidates Cell Builder, Array Builder,
// Verilog Export, DEF Export, Validation, and Export All functionality
package tabs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/module6-eda/pkg/gui/widgets"
	"fecim-lattice-tools/module6-eda/pkg/openlane"
	"fecim-lattice-tools/module6-eda/pkg/validation"
	"fecim-lattice-tools/shared/logging"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MakeBuilderValidationTab creates a unified tab combining cell/array configuration,
// preview (Verilog/DEF/Layout), validation, and export functionality
func MakeBuilderValidationTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	logging.GlobalDebug("[EDA-Builder] Creating Builder & Validation tab")

	// ========== CELL CONFIG SECTION ==========
	nameEntry := widget.NewEntry()
	nameEntry.SetText("fecim_bitcell")

	widthEntry := widget.NewEntry()
	widthEntry.SetText("0.460")

	heightEntry := widget.NewEntry()
	heightEntry.SetText("2.720")

	riseEntry := widget.NewEntry()
	riseEntry.SetText("10.0")

	fallEntry := widget.NewEntry()
	fallEntry.SetText("10.0")

	capEntry := widget.NewEntry()
	capEntry.SetText("0.015")

	leakageEntry := widget.NewEntry()
	leakageEntry.SetText("0.0003")

	// Helper to parse cell config from inputs
	getCellConfig := func() config.CellConfig {
		width, err := strconv.ParseFloat(widthEntry.Text, 64)
		if err != nil {
			width = 0.460 // Default value
		}
		height, err := strconv.ParseFloat(heightEntry.Text, 64)
		if err != nil {
			height = 2.720 // Default value
		}
		rise, err := strconv.ParseFloat(riseEntry.Text, 64)
		if err != nil {
			rise = 10.0 // Default value
		}
		fall, err := strconv.ParseFloat(fallEntry.Text, 64)
		if err != nil {
			fall = 10.0 // Default value
		}
		cap, err := strconv.ParseFloat(capEntry.Text, 64)
		if err != nil {
			cap = 0.015 // Default value
		}
		leakage, err := strconv.ParseFloat(leakageEntry.Text, 64)
		if err != nil {
			leakage = 0.0003 // Default value
		}

		return config.CellConfig{
			Name:         nameEntry.Text,
			Width:        width,
			Height:       height,
			CellType:     cfg.Architecture, // Uses shared architecture selector
			Technology:   cfg.Technology,
			RiseTime:     rise,
			FallTime:     fall,
			InputCap:     cap,
			LeakagePower: leakage,
		}
	}

	cellAreaLabel := widget.NewLabel(fmt.Sprintf("Cell Area: %.4f µm²", 0.46*2.72))

	// ========== ARRAY CONFIG SECTION ==========
	rowsEntry := widget.NewEntry()
	rowsEntry.SetText(fmt.Sprintf("%d", cfg.Rows))

	colsEntry := widget.NewEntry()
	colsEntry.SetText(fmt.Sprintf("%d", cfg.Cols))

	// Mode selector with help text
	modeHelpText := widget.NewLabel("")
	modeHelpText.Wrapping = fyne.TextWrapWord

	updateModeHelp := func(mode string) {
		switch mode {
		case "storage":
			modeHelpText.SetText("Storage mode: Non-volatile data retention using FeCIM cells as memory elements")
		case "memory":
			modeHelpText.SetText("Memory mode: Fast read/write operations with row/column addressing for DRAM-like access")
		case "compute":
			modeHelpText.SetText("Compute mode: In-memory matrix-vector multiplication using analog conductance states")
		default:
			modeHelpText.SetText("")
		}
	}

	modeSelect := widget.NewSelect([]string{"storage", "memory", "compute"}, func(s string) {
		cfg.Mode = s
		updateModeHelp(s)
	})
	modeSelect.SetSelected(cfg.Mode)
	updateModeHelp(cfg.Mode) // Initialize help text

	// Helper to create zoomable image tab
	makeZoomableImageTab := func(img *canvas.Image, label *widget.Label, status *widget.Label) fyne.CanvasObject {
		zoomLevel := 1.0
		baseWidth := float32(600)
		baseHeight := float32(450)

		zoomLabel := widget.NewLabel("100%")

		updateZoom := func() {
			newW := baseWidth * float32(zoomLevel)
			newH := baseHeight * float32(zoomLevel)
			img.SetMinSize(fyne.NewSize(newW, newH))
			img.Refresh()
			zoomLabel.SetText(fmt.Sprintf("%.0f%%", zoomLevel*100))
		}

		zoomInBtn := widget.NewButton("+", func() {
			if zoomLevel < 3.0 {
				zoomLevel += 0.25
				updateZoom()
			}
		})
		zoomOutBtn := widget.NewButton("-", func() {
			if zoomLevel > 0.25 {
				zoomLevel -= 0.25
				updateZoom()
			}
		})
		fitBtn := widget.NewButton("Fit", func() {
			zoomLevel = 1.0
			updateZoom()
		})

		zoomControls := container.NewHBox(
			label,
			widget.NewLabel(" - "),
			status,
			layout.NewSpacer(),
			widget.NewLabel("Zoom:"),
			zoomOutBtn,
			zoomLabel,
			zoomInBtn,
			fitBtn,
		)

		return container.NewBorder(
			zoomControls,
			nil, nil, nil,
			container.NewScroll(img),
		)
	}

	// Layout image display - shows KLayout, OpenROAD, and Yosys generated images
	// === TABBED INTERFACE FOR BETTER VISIBILITY ===

	// 1. KLayout image (physical layout from DEF/LEF)
	klayoutImage := canvas.NewImageFromFile("")
	klayoutImage.FillMode = canvas.ImageFillContain
	klayoutImage.SetMinSize(fyne.NewSize(600, 450))
	klayoutLabel := widget.NewLabelWithStyle("KLayout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	klayoutStatus := widget.NewLabel("Not generated")

	// 2. OpenROAD image (placement visualization)
	openroadImage := canvas.NewImageFromFile("")
	openroadImage.FillMode = canvas.ImageFillContain
	openroadImage.SetMinSize(fyne.NewSize(600, 450))
	openroadLabel := widget.NewLabelWithStyle("OpenROAD", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	openroadStatus := widget.NewLabel("Not generated")

	// 3. Yosys schematic image (circuit diagram - PNG converted from DOT)
	yosysImage := canvas.NewImageFromFile("")
	yosysImage.FillMode = canvas.ImageFillContain
	yosysImage.SetMinSize(fyne.NewSize(600, 450))
	yosysLabel := widget.NewLabelWithStyle("Yosys Schematic", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	yosysStatus := widget.NewLabel("Not generated")

	// Create zoomable image tabs
	klayoutTab := makeZoomableImageTab(klayoutImage, klayoutLabel, klayoutStatus)
	openroadTab := makeZoomableImageTab(openroadImage, openroadLabel, openroadStatus)
	yosysTab := makeZoomableImageTab(yosysImage, yosysLabel, yosysStatus)
	imageTabs := container.NewAppTabs(
		container.NewTabItem("KLayout", klayoutTab),
		container.NewTabItem("OpenROAD", openroadTab),
		container.NewTabItem("Yosys Schematic", yosysTab),
	)
	imageTabs.SetTabLocation(container.TabLocationTop)

	// Replace layoutStack with imageTabs
	layoutStack := imageTabs

	// Helper to update KLayout image from file
	updateLayoutImage := func() {
		pngPath := fmt.Sprintf("data/fecim_crossbar_%dx%d.png", cfg.Rows, cfg.Cols)
		absPath, _ := filepath.Abs(pngPath)
		if fileExists(absPath) {
			sharedwidgets.SafeDo(func() {
				klayoutImage.File = absPath
				klayoutImage.Resource = nil
				klayoutStatus.SetText("Generated: " + filepath.Base(pngPath))
				klayoutImage.Refresh()
			})
		}
	}

	// Helper to update Yosys schematic image (PNG converted from DOT)
	updateYosysImage := func() {
		// Try PNG first (converted from DOT by graphviz)
		pngPath := fmt.Sprintf("data/fecim_crossbar_%dx%d_schematic.png", cfg.Rows, cfg.Cols)
		absPath, _ := filepath.Abs(pngPath)
		if fileExists(absPath) {
			sharedwidgets.SafeDo(func() {
				yosysImage.File = absPath
				yosysImage.Resource = nil
				yosysStatus.SetText("Generated: " + filepath.Base(pngPath))
				yosysImage.Refresh()
			})
			return
		}
		// Fallback: check if DOT file exists but PNG conversion failed
		dotPath := fmt.Sprintf("data/fecim_crossbar_%dx%d_schematic.dot", cfg.Rows, cfg.Cols)
		dotAbs, _ := filepath.Abs(dotPath)
		if fileExists(dotAbs) {
			sharedwidgets.SafeDo(func() {
				yosysStatus.SetText("DOT only (install graphviz)")
			})
		}
	}

	// Helper to update OpenROAD image
	updateOpenROADImage := func() {
		pngPath := fmt.Sprintf("data/fecim_crossbar_%dx%d_openroad.png", cfg.Rows, cfg.Cols)
		absPath, _ := filepath.Abs(pngPath)
		if fileExists(absPath) {
			sharedwidgets.SafeDo(func() {
				openroadImage.File = absPath
				openroadImage.Resource = nil
				openroadStatus.SetText("Generated: " + filepath.Base(pngPath))
				openroadImage.Refresh()
			})
		}
	}

	// Try to load existing images on startup
	updateLayoutImage()
	updateYosysImage()
	updateOpenROADImage()

	// Keep widgets import used
	_ = widgets.NewLayoutCanvas

	// Architecture toggle buttons (same style as crossbar module)
	archPassiveBtn := widget.NewButton("PASSIVE", nil)
	arch1T1RBtn := widget.NewButton("1T1R", nil)
	arch2T1RBtn := widget.NewButton("2T1R", nil)

	// Helper to update button styles based on selection
	updateArchButtons := func() {
		archPassiveBtn.Importance = widget.LowImportance
		arch1T1RBtn.Importance = widget.LowImportance
		arch2T1RBtn.Importance = widget.LowImportance
		switch cfg.Architecture {
		case "passive":
			archPassiveBtn.Importance = widget.HighImportance
		case "1t1r":
			arch1T1RBtn.Importance = widget.HighImportance
		case "2t1r":
			arch2T1RBtn.Importance = widget.HighImportance
		default:
			archPassiveBtn.Importance = widget.HighImportance
		}
		archPassiveBtn.Refresh()
		arch1T1RBtn.Refresh()
		arch2T1RBtn.Refresh()
	}

	// Set initial state
	updateArchButtons()

	// Wire up callbacks
	archPassiveBtn.OnTapped = func() {
		if cfg.Architecture == "passive" {
			return // Already selected
		}
		logging.GlobalDebug("[EDA-Builder] Architecture changed to: passive")
		cfg.Architecture = "passive"
		// Reset to passive cell dimensions (triggers updateStats via OnChanged)
		widthEntry.SetText("0.460")
		heightEntry.SetText("2.720")
		updateArchButtons()
		updateLayoutImage()
	}

	arch1T1RBtn.OnTapped = func() {
		if cfg.Architecture == "1t1r" {
			return // Already selected
		}
		logging.GlobalDebug("[EDA-Builder] Architecture changed to: 1t1r")
		cfg.Architecture = "1t1r"
		// Update to 1T1R cell dimensions - SKY130 hvl standard (triggers updateStats via OnChanged)
		widthEntry.SetText("0.460")
		heightEntry.SetText("4.070")
		updateArchButtons()
		updateLayoutImage()
	}

	arch2T1RBtn.OnTapped = func() {
		if cfg.Architecture == "2t1r" {
			return // Already selected
		}
		logging.GlobalDebug("[EDA-Builder] Architecture changed to: 2t1r")
		cfg.Architecture = "2t1r"
		// Update to 2T1R cell dimensions - two SKY130 sites wide, hvl height (triggers updateStats via OnChanged)
		widthEntry.SetText("0.920")
		heightEntry.SetText("4.070")
		updateArchButtons()
		updateLayoutImage()
	}

	// Architecture toggle container (3 buttons for passive, 1T1R, 2T1R)
	archToggle := container.NewGridWithColumns(3, archPassiveBtn, arch1T1RBtn, arch2T1RBtn)

	// Statistics labels - now with more metrics
	totalLabel := widget.NewLabel(fmt.Sprintf("Total Cells: %d", cfg.Rows*cfg.Cols))
	areaLabel := widget.NewLabel(fmt.Sprintf("Array Area: %.2f µm²", float64(cfg.Rows*cfg.Cols)*cfg.CellWidth*cfg.CellHeight))
	wlLengthLabel := widget.NewLabel(fmt.Sprintf("WL Length: %.2f µm", float64(cfg.Cols)*cfg.CellWidth))
	blLengthLabel := widget.NewLabel(fmt.Sprintf("BL Length: %.2f µm", float64(cfg.Rows)*cfg.CellHeight))
	densityLabel := widget.NewLabel(fmt.Sprintf("Density: %.2f cells/µm²", 0.0))
	utilizationLabel := widget.NewLabel(fmt.Sprintf("Utilization: %.1f%%", 0.0))

	// Update statistics function
	updateStats := func() {
		rows, err := strconv.Atoi(rowsEntry.Text)
		if err != nil || rows <= 0 {
			rows = cfg.Rows // Keep current value
		}
		if rows > 1024 {
			rows = 1024 // Prevent OOM from unreasonable array sizes
		}
		cols, err := strconv.Atoi(colsEntry.Text)
		if err != nil || cols <= 0 {
			cols = cfg.Cols // Keep current value
		}
		if cols > 1024 {
			cols = 1024 // Prevent OOM from unreasonable array sizes
		}
		cfg.Rows = rows
		cfg.Cols = cols

		// Update cell dimensions from entries
		cellW, err := strconv.ParseFloat(widthEntry.Text, 64)
		if err != nil || cellW <= 0 {
			cellW = cfg.CellWidth // Keep current value
		}
		cellH, err := strconv.ParseFloat(heightEntry.Text, 64)
		if err != nil || cellH <= 0 {
			cellH = cfg.CellHeight // Keep current value
		}
		cfg.CellWidth = cellW
		cfg.CellHeight = cellH

		total := rows * cols
		area := float64(total) * cfg.CellWidth * cfg.CellHeight
		wlLength := float64(cols) * cfg.CellWidth
		blLength := float64(rows) * cfg.CellHeight

		// Calculate density (cells per µm²)
		arrayWidth := wlLength
		arrayHeight := blLength
		arrayTotalArea := arrayWidth * arrayHeight
		density := 0.0
		if arrayTotalArea > 0 {
			density = float64(total) / arrayTotalArea
		}

		// Calculate utilization (cell area / total area)
		utilization := 0.0
		if arrayTotalArea > 0 {
			utilization = (area / arrayTotalArea) * 100.0
		}

		logging.GlobalDebug("[EDA-Builder] Stats updated: %dx%d array, area=%.2f µm²", rows, cols, area)

		sharedwidgets.SafeDo(func() {
			totalLabel.SetText(fmt.Sprintf("Total Cells: %d", total))
			areaLabel.SetText(fmt.Sprintf("Array Area: %.2f µm²", area))
			wlLengthLabel.SetText(fmt.Sprintf("WL Length: %.2f µm", wlLength))
			blLengthLabel.SetText(fmt.Sprintf("BL Length: %.2f µm", blLength))
			densityLabel.SetText(fmt.Sprintf("Density: %.4f cells/µm²", density))
			utilizationLabel.SetText(fmt.Sprintf("Utilization: %.1f%%", utilization))
			cellAreaLabel.SetText(fmt.Sprintf("Cell Area: %.4f µm²", cellW*cellH))
		})
	}

	// Wire up change handlers
	rowsEntry.OnChanged = func(s string) { updateStats() }
	colsEntry.OnChanged = func(s string) { updateStats() }
	widthEntry.OnChanged = func(s string) { updateStats() }
	heightEntry.OnChanged = func(s string) { updateStats() }

	// ========== PREVIEW SECTION ==========
	verilogPreview := widget.NewMultiLineEntry()
	verilogPreview.Wrapping = fyne.TextWrapOff
	verilogPreview.SetText("// Example Verilog Array Structure\n// After clicking 'Generate All', this will show:\n//   - Module definition with WL/BL ports\n//   - Cell instantiation grid\n//   - Power/ground connections\n\nmodule fecim_crossbar_NxM (\n    input [N-1:0] WL,\n    output [M-1:0] BL,\n    input VPWR, VGND\n);\n    // ... array instances ...\nendmodule")

	defPreview := widget.NewMultiLineEntry()
	defPreview.Wrapping = fyne.TextWrapOff
	defPreview.SetText("# Example DEF Placement Structure\n# After clicking 'Generate All', this will show:\n#   - Design header with units and die area\n#   - Component placement coordinates\n#   - Pin definitions\n\nVERSION 5.8 ;\nDESIGN fecim_crossbar_NxM ;\nUNITS DISTANCE MICRONS 1000 ;\nDIEAREA ( 0 0 ) ( ... ) ;\nCOMPONENTS ... ;\n  - cell_0_0 fecim_bitcell + FIXED ( ... ) N ;\nEND COMPONENTS")

	// Button to generate Yosys schematic SVG
	genSchematicBtn := widget.NewButton("Gen Schematic (Yosys)", func() {
		logging.GlobalInfo("[EDA-Builder] Yosys schematic generation started")
		sharedwidgets.SafeDo(func() {
			yosysStatus.SetText("Generating...")
		})
		go func() {
			verilogPath := fmt.Sprintf("data/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
			outputPrefix := fmt.Sprintf("data/fecim_crossbar_%dx%d_schematic", cfg.Rows, cfg.Cols)
			topModule := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)

			manager := openlane.NewManager()
			config := openlane.DefaultConfig()

			result, err := validation.GenerateYosysSchematic(verilogPath, outputPrefix, topModule, cfg.Architecture, manager, config)
			if err != nil {
				logging.GlobalError("[EDA-Builder] Yosys schematic generation failed: %v", err)
			}
			sharedwidgets.SafeDo(func() {
				if result != nil && result.Success {
					logging.GlobalInfo("[EDA-Builder] Yosys schematic generated successfully: %s", result.ImagePath)
					updateYosysImage()
					dialog.ShowInformation("Schematic Generated", "Yosys schematic saved to:\n"+result.ImagePath, window)
				} else {
					errMsg := "Unknown error"
					rawOutput := ""
					if result != nil {
						if result.Error != "" {
							errMsg = result.Error
						}
						rawOutput = result.RawOutput
					}
					logging.GlobalError("[EDA-Builder] Yosys schematic generation failed: %s", errMsg)
					yosysStatus.SetText("Failed: " + errMsg)
					// Show detailed error with raw output
					detailMsg := fmt.Sprintf("Schematic generation failed: %s", errMsg)
					if rawOutput != "" {
						detailMsg += "\n\nYosys output:\n" + rawOutput
					}
					dialog.ShowError(fmt.Errorf("%s", detailMsg), window)
				}
			})
		}()
	})

	// Button to generate OpenROAD layout PNG
	genOpenROADBtn := widget.NewButton("Gen Layout (OpenROAD)", func() {
		logging.GlobalInfo("[EDA-Builder] OpenROAD layout generation started")
		sharedwidgets.SafeDo(func() {
			openroadStatus.SetText("Generating...")
		})
		go func() {
			defPath := fmt.Sprintf("data/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
			var lefPath string
			switch cfg.Architecture {
			case "1t1r":
				lefPath = "cells/fecim_1t1r_bitcell/fecim_1t1r_bitcell.lef"
			case "2t1r":
				lefPath = "cells/fecim_2t1r_bitcell/fecim_2t1r_bitcell.lef"
			default:
				lefPath = "cells/fecim_bitcell/fecim_bitcell.lef"
			}
			outputPath := fmt.Sprintf("data/fecim_crossbar_%dx%d_openroad.png", cfg.Rows, cfg.Cols)

			manager := openlane.NewManager()
			config := openlane.DefaultConfig()

			result, err := validation.GenerateOpenROADImage(defPath, lefPath, outputPath, manager, config)
			if err != nil {
				logging.GlobalError("[EDA-Builder] OpenROAD layout generation failed: %v", err)
			}
			sharedwidgets.SafeDo(func() {
				if result != nil && result.Success {
					logging.GlobalInfo("[EDA-Builder] OpenROAD layout generated successfully: %s", result.ImagePath)
					updateOpenROADImage()
					dialog.ShowInformation("Layout Generated", "OpenROAD layout saved to:\n"+result.ImagePath, window)
				} else {
					errMsg := "Unknown error"
					rawOutput := ""
					if result != nil {
						if result.Error != "" {
							errMsg = result.Error
						}
						rawOutput = result.RawOutput
					}
					logging.GlobalError("[EDA-Builder] OpenROAD layout generation failed: %s", errMsg)
					openroadStatus.SetText("Failed: " + errMsg)
					// Show detailed error with raw output
					detailMsg := fmt.Sprintf("Layout generation failed: %s", errMsg)
					if rawOutput != "" {
						detailMsg += "\n\nOpenROAD output:\n" + rawOutput
					}
					dialog.ShowError(fmt.Errorf("%s", detailMsg), window)
				}
			})
		}()
	})

	verilogStatsLabel := widget.NewLabel("Verilog: Pending")
	defStatsLabel := widget.NewLabel("DEF: Pending")

	// ========== VALIDATION SECTION ==========
	yosysResult := widget.NewLabel("Not validated")
	defResult := widget.NewLabel("Not validated")
	crossResult := widget.NewLabel("Not validated")
	validationSummary := widget.NewLabel("")
	validationSummary.TextStyle.Bold = true
	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.TextStyle.Monospace = true

	addLog := func(msg string) {
		logging.GlobalInfo("%s", msg)
		sharedwidgets.SafeDo(func() {
			logOutput.SetText(logOutput.Text + msg + "\n")
		})
	}

	clearLogBtn := widget.NewButton("Clear Log", func() {
		logOutput.SetText("")
	})

	// ========== OPENLANE STATUS SECTION ==========
	dockerStatus := widget.NewLabel("Checking...")
	pdkStatus := widget.NewLabel("Checking...")
	placementResult := widget.NewLabel("Not validated")

	// Pull Docker Image button (only shown when needed)
	var pullImageBtn *widget.Button
	pullImageBtn = widget.NewButton("Pull OpenLane Image", func() {
		logging.GlobalInfo("[EDA-Builder] Docker image pull started")
		go func() {
			sharedwidgets.SafeDo(func() {
				dockerStatus.SetText("Pulling image...")
			})
			addLog("=== Pulling OpenLane Docker Image ===")
			addLog("This may take several minutes...")

			manager := openlane.NewManager()
			err := manager.PullDockerImage(func(msg string) {
				addLog(msg)
			})
			if err != nil {
				logging.GlobalError("[EDA-Builder] Docker image pull failed: %v", err)
				sharedwidgets.SafeDo(func() {
					dockerStatus.SetText("✗ Pull failed: " + err.Error())
				})
				addLog("ERROR: " + err.Error())
			} else {
				logging.GlobalInfo("[EDA-Builder] Docker image pull completed successfully")
				sharedwidgets.SafeDo(func() {
					dockerStatus.SetText("✓ Docker image ready")
					pullImageBtn.Hide()
				})
				addLog("Docker image pulled successfully")
			}
		}()
	})
	pullImageBtn.Hide() // Initially hidden until check completes

	// Check OpenLane status on startup
	go func() {
		manager := openlane.NewManager()
		mode := manager.DetectMode()

		sharedwidgets.SafeDo(func() {
			if mode == openlane.ModeDocker {
				if manager.IsDockerImagePulled() {
					dockerStatus.SetText("✓ Docker image ready")
					pullImageBtn.Hide()
				} else {
					// Should not happen with current DetectMode logic, but for safety
					dockerStatus.SetText("○ Docker image not pulled")
					pullImageBtn.Show()
				}
			} else if mode == openlane.ModeNative {
				dockerStatus.SetText("✓ Native tools detected")
				pullImageBtn.Hide()
			} else {
				// ModeNone
				if manager.IsDockerAvailable() {
					dockerStatus.SetText("○ Docker found, image missing")
					pullImageBtn.Show()
					pullImageBtn.Enable()
				} else {
					dockerStatus.SetText("✗ OpenLane/Docker not available")
					pullImageBtn.Disable()
				}
			}

			if manager.IsPDKInstalled() {
				pdkStatus.SetText("✓ SKY130A PDK available (optional)")
			} else {
				pdkStatus.SetText("○ Not needed - uses FeCIM cell library")
			}
		})
	}()

	// ========== STATUS ==========
	statusLabel := widget.NewLabel("Ready")

	// ========== GENERATE ALL BUTTON ==========
	var generateAllBtn *widget.Button
	var validateAllBtn *widget.Button
	var exportPackageBtn *widget.Button

	generateAllBtn = widget.NewButton("Generate All", func() {
		logging.GlobalInfo("[EDA-Builder] Generate All started")
		generateAllBtn.Disable()
		validateAllBtn.Disable()
		exportPackageBtn.Disable()
		generateAllBtn.SetText("Generating...")
		go func() {
			sharedwidgets.SafeDo(func() {
				statusLabel.SetText("Generating...")
				logOutput.SetText("")
			})

			updateStats()
			cellCfg := getCellConfig()

			// Generate cell files - use appropriate directory based on architecture
			addLog("Generating cell library...")
			lefContent := export.GenerateLEF(cellCfg)
			libContent := export.GenerateLiberty(cellCfg)
			cellVContent := export.GenerateCellVerilog(cellCfg)

			// Choose cell directory and name based on architecture
			var dir, cellFileName string
			switch cfg.Architecture {
			case "1t1r":
				dir = "cells/fecim_1t1r_bitcell"
				cellFileName = "fecim_1t1r_bitcell"
			case "2t1r":
				dir = "cells/fecim_2t1r_bitcell"
				cellFileName = "fecim_2t1r_bitcell"
			default:
				dir = "cells/fecim_bitcell"
				cellFileName = "fecim_bitcell"
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				addLog("ERROR: Failed to create directory " + dir + ": " + err.Error())
				sharedwidgets.SafeDo(func() {
					statusLabel.SetText("Generation failed")
					generateAllBtn.Enable()
					validateAllBtn.Enable()
					exportPackageBtn.Enable()
					generateAllBtn.SetText("Generate All")
				})
				return
			}
			if err := os.WriteFile(dir+"/"+cellFileName+".lef", []byte(lefContent), 0644); err != nil {
				addLog("ERROR: Failed to write LEF: " + err.Error())
			}
			if err := os.WriteFile(dir+"/"+cellFileName+".lib", []byte(libContent), 0644); err != nil {
				addLog("ERROR: Failed to write LIB: " + err.Error())
			}
			if err := os.WriteFile(dir+"/"+cellFileName+".v", []byte(cellVContent), 0644); err != nil {
				addLog("ERROR: Failed to write V: " + err.Error())
			}
			addLog("  LEF/LIB/V written to " + dir)

			// Generate array Verilog
			addLog("Generating array Verilog...")
			vContent := export.GenerateArrayVerilog(*cfg)
			sharedwidgets.SafeDo(func() {
				verilogPreview.SetText(vContent)
			})

			instances := cfg.Rows * cfg.Cols
			lines := strings.Count(vContent, "\n")
			size := float64(len(vContent)) / 1024
			sharedwidgets.SafeDo(func() {
				verilogStatsLabel.SetText(fmt.Sprintf("Instances: %d | Lines: %d | Size: %.1fKB", instances, lines, size))
			})

			vFilename := fmt.Sprintf("data/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
			if err := os.MkdirAll("data", 0755); err != nil {
				addLog("ERROR: Failed to create data directory: " + err.Error())
			}
			if err := os.WriteFile(vFilename, []byte(vContent), 0644); err != nil {
				addLog("ERROR: Failed to write Verilog: " + err.Error())
			}
			addLog("  Verilog: " + vFilename)

			// Generate DEF
			addLog("Generating DEF placement...")
			defContent := generateBuilderDEF(*cfg)
			sharedwidgets.SafeDo(func() {
				defPreview.SetText(defContent)
			})

			defFilename := fmt.Sprintf("data/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
			if err := os.WriteFile(defFilename, []byte(defContent), 0644); err != nil {
				addLog("ERROR: Failed to write DEF: " + err.Error())
			}
			sharedwidgets.SafeDo(func() {
				defStatsLabel.SetText(fmt.Sprintf("Components: %d | File: %s", cfg.Rows*cfg.Cols, defFilename))
			})
			addLog("  DEF: " + defFilename)

			// Generate layout image using KLayout (if available)
			addLog("Generating layout image...")
			pngFilename := fmt.Sprintf("data/fecim_crossbar_%dx%d.png", cfg.Rows, cfg.Cols)

			// In headless/unit-test mode, skip external tool probing to keep actions fast/deterministic.
			if window == nil {
				addLog("  Skipped KLayout generation (headless mode)")
				sharedwidgets.SafeDo(func() {
					klayoutStatus.SetText("Skipped (headless)")
				})
			} else {
				// Determine LEF path based on architecture
				var cellLEFPath string
				switch cfg.Architecture {
				case "1t1r":
					cellLEFPath = "cells/fecim_1t1r_bitcell/fecim_1t1r_bitcell.lef"
				case "2t1r":
					cellLEFPath = "cells/fecim_2t1r_bitcell/fecim_2t1r_bitcell.lef"
				default:
					cellLEFPath = "cells/fecim_bitcell/fecim_bitcell.lef"
				}

				imgManager := openlane.NewManager()
				imgConfig := openlane.DefaultConfig()
				sharedwidgets.SafeDo(func() {
					klayoutStatus.SetText("Generating...")
				})
				if validation.IsKLayoutAvailable(imgManager) {
					imgResult, err := validation.GenerateLayoutImage(
						defFilename,
						cellLEFPath,
						pngFilename,
						imgManager,
						imgConfig,
					)
					if err != nil {
						addLog("ERROR: " + err.Error())
					}
					if imgResult != nil && imgResult.Success {
						addLog("  PNG (KLayout): " + pngFilename)
						// Update the layout image display
						updateLayoutImage()
					} else {
						errMsg := "unknown error"
						if imgResult != nil && imgResult.Error != "" {
							errMsg = imgResult.Error
						}
						addLog("  KLayout failed: " + errMsg)
						// Show raw output for debugging
						if imgResult != nil && imgResult.RawOutput != "" {
							addLog("  KLayout output:")
							for _, line := range strings.Split(imgResult.RawOutput, "\n") {
								if line != "" {
									addLog("    " + line)
								}
							}
						}
						addLog("  Use 'Gen Layout (OpenROAD)' button for alternative")
						sharedwidgets.SafeDo(func() {
							klayoutStatus.SetText("Failed: " + errMsg)
						})
					}
				} else {
					addLog("  KLayout not available (install Docker with OpenLane image)")
					addLog("  Use 'Gen Layout (OpenROAD)' button for alternative")
					sharedwidgets.SafeDo(func() {
						klayoutStatus.SetText("Not available (need Docker)")
					})
				}
			}

			// Generate OpenLane config
			addLog("Generating OpenLane config...")
			configContent := export.GenerateOpenLaneConfig(*cfg)
			if err := os.WriteFile("data/config.json", []byte(configContent), 0644); err != nil {
				addLog("ERROR: Failed to write OpenLane config: " + err.Error())
			}
			addLog("  Config: data/config.json")

			logging.GlobalInfo("[EDA-Builder] Generate All completed successfully")
			sharedwidgets.SafeDo(func() {
				statusLabel.SetText("All files generated")
				generateAllBtn.Enable()
				validateAllBtn.Enable()
				exportPackageBtn.Enable()
				generateAllBtn.SetText("Generate All")
			})
			addLog("Generation complete!")
		}()
	})

	// ========== VALIDATE ALL BUTTON ==========
	validateAllBtn = widget.NewButton("Validate All", func() {
		logging.GlobalInfo("[EDA-Builder] Validate All started")
		validateAllBtn.Disable()
		generateAllBtn.Disable()
		exportPackageBtn.Disable()
		validateAllBtn.SetText("Validating...")
		go func() {
			sharedwidgets.SafeDo(func() {
				statusLabel.SetText("Validating...")
				logOutput.SetText("")
				yosysResult.SetText("...")
				defResult.SetText("...")
				crossResult.SetText("...")
				validationSummary.SetText("Running validations...")
			})

			allPassed := true

			// Determine cell paths based on architecture
			var cellDir, cellFileName string
			switch cfg.Architecture {
			case "1t1r":
				cellDir = "cells/fecim_1t1r_bitcell"
				cellFileName = "fecim_1t1r_bitcell"
			case "2t1r":
				cellDir = "cells/fecim_2t1r_bitcell"
				cellFileName = "fecim_2t1r_bitcell"
			default:
				cellDir = "cells/fecim_bitcell"
				cellFileName = "fecim_bitcell"
			}

			// Yosys validation
			addLog("=== Yosys Verilog Validation ===")
			arrayPath := fmt.Sprintf("data/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
			cellPath := cellDir + "/" + cellFileName + ".v"
			addLog(fmt.Sprintf("Array: %s", arrayPath))
			addLog(fmt.Sprintf("Cell:  %s", cellPath))

			err1 := validation.ValidateVerilogWithCell(arrayPath, cellPath)
			if err1 != nil {
				logging.GlobalDebug("[EDA-Builder] Yosys validation: FAIL")
				sharedwidgets.SafeDo(func() { yosysResult.SetText("✗ FAIL") })
				addLog(fmt.Sprintf("ERROR: %v", err1))
				allPassed = false
			} else {
				logging.GlobalDebug("[EDA-Builder] Yosys validation: PASS")
				sharedwidgets.SafeDo(func() { yosysResult.SetText("✓ PASS") })
				addLog("PASSED")
			}

			// DEF validation
			addLog("\n=== DEF Syntax Validation ===")
			defPath := fmt.Sprintf("data/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
			addLog(fmt.Sprintf("DEF: %s", defPath))

			err2 := validation.ValidateDEF(defPath)
			if err2 != nil {
				logging.GlobalDebug("[EDA-Builder] DEF validation: FAIL")
				sharedwidgets.SafeDo(func() { defResult.SetText("✗ FAIL") })
				addLog(fmt.Sprintf("ERROR: %v", err2))
				allPassed = false
			} else {
				logging.GlobalDebug("[EDA-Builder] DEF validation: PASS")
				sharedwidgets.SafeDo(func() { defResult.SetText("✓ PASS") })
				stats, _ := validation.GetDEFStats(defPath)
				addLog(fmt.Sprintf("Design: %v, Components: %v", stats["design_name"], stats["component_count"]))
				addLog("PASSED")
			}

			// Cross-check validation
			addLog("\n=== LEF/LIB/V Cross-Check ===")
			lefPath := cellDir + "/" + cellFileName + ".lef"
			libPath := cellDir + "/" + cellFileName + ".lib"
			vPath := cellDir + "/" + cellFileName + ".v"
			addLog(fmt.Sprintf("LEF: %s", lefPath))
			addLog(fmt.Sprintf("LIB: %s", libPath))
			addLog(fmt.Sprintf("V:   %s", vPath))

			err3 := validation.CrossCheckFiles(lefPath, libPath, vPath)
			if err3 != nil {
				logging.GlobalDebug("[EDA-Builder] Cross-check validation: FAIL")
				sharedwidgets.SafeDo(func() { crossResult.SetText("✗ FAIL") })
				addLog(fmt.Sprintf("ERROR: %v", err3))
				allPassed = false
			} else {
				logging.GlobalDebug("[EDA-Builder] Cross-check validation: PASS")
				sharedwidgets.SafeDo(func() { crossResult.SetText("✓ PASS") })
				addLog("Pin names and cell names match")
				addLog("PASSED")
			}

			// OpenLane Placement validation (runs when Docker/OpenROAD is available)
			// Uses our custom FeCIM cell LEF - no external PDK required
			addLog("\n=== OpenLane Placement Validation ===")
			if window == nil {
				sharedwidgets.SafeDo(func() { placementResult.SetText("⊝ SKIP") })
				addLog("SKIPPED: headless mode")
			} else {
				manager := openlane.NewManager()
				mode := manager.DetectMode()

				if mode == openlane.ModeNone {
					sharedwidgets.SafeDo(func() { placementResult.SetText("⊝ SKIP") })
					addLog("SKIPPED: OpenLane/Docker not available")
				} else {
					defPath := fmt.Sprintf("data/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)

					addLog(fmt.Sprintf("Mode: %s", mode))
					addLog(fmt.Sprintf("DEF: %s", defPath))
					addLog(fmt.Sprintf("Cell LEF: %s", lefPath))

					config := openlane.DefaultConfig()
					result, err := validation.RunPlacementCheckWithCell(defPath, lefPath, manager, config)
					if err != nil {
						logging.GlobalDebug("[EDA-Builder] Placement validation: ERROR")
						sharedwidgets.SafeDo(func() { placementResult.SetText("✗ ERROR") })
						addLog(fmt.Sprintf("ERROR: %v", err))
						allPassed = false
					} else if result.Passed {
						logging.GlobalDebug("[EDA-Builder] Placement validation: PASS")
						sharedwidgets.SafeDo(func() { placementResult.SetText("✓ PASS") })
						addLog(result.RawOutput)
						addLog("PASSED")
					} else {
						logging.GlobalDebug("[EDA-Builder] Placement validation: FAIL (%d violations)", result.ViolationCount)
						sharedwidgets.SafeDo(func() { placementResult.SetText("✗ FAIL") })
						addLog(result.RawOutput)
						addLog(fmt.Sprintf("FAILED: %d violations", result.ViolationCount))
						for _, v := range result.Violations {
							addLog(fmt.Sprintf("  - %s: %s", v.Issue, v.Message))
						}
						allPassed = false
					}
				}
			}

			// Summary
			if allPassed {
				logging.GlobalInfo("[EDA-Builder] Validate All completed: all checks passed")
				sharedwidgets.SafeDo(func() {
					statusLabel.SetText("All validations passed")
					validationSummary.SetText("✓ All checks passed")
					validateAllBtn.Enable()
					generateAllBtn.Enable()
					exportPackageBtn.Enable()
					validateAllBtn.SetText("Validate All")
				})
				addLog("\n=== ALL VALIDATIONS PASSED ===")
			} else {
				logging.GlobalInfo("[EDA-Builder] Validate All completed: some checks failed")
				sharedwidgets.SafeDo(func() {
					statusLabel.SetText("Some validations failed")
					validationSummary.SetText("✗ Some checks failed - see log for details")
					validateAllBtn.Enable()
					generateAllBtn.Enable()
					exportPackageBtn.Enable()
					validateAllBtn.SetText("Validate All")
				})
				addLog("\n=== SOME VALIDATIONS FAILED ===")
			}
		}()
	})

	// ========== EXPORT PACKAGE BUTTON ==========
	exportPackageBtn = widget.NewButton("Export Package", func() {
		exportPackageBtn.Disable()
		generateAllBtn.Disable()
		validateAllBtn.Disable()
		exportPackageBtn.SetText("Exporting...")
		go func() {
			sharedwidgets.SafeDo(func() {
				statusLabel.SetText("Exporting package...")
				logOutput.SetText("")
			})

			designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
			outputDir := fmt.Sprintf("data/%s", designName)
			logging.GlobalInfo("[EDA-Builder] Export Package started to: %s", outputDir)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				addLog("ERROR: Failed to create directory " + outputDir + ": " + err.Error())
				sharedwidgets.SafeDo(func() {
					statusLabel.SetText("Export failed")
					exportPackageBtn.Enable()
					generateAllBtn.Enable()
					validateAllBtn.Enable()
					exportPackageBtn.SetText("Export Package")
				})
				return
			}
			if err := os.MkdirAll(outputDir+"/cells", 0755); err != nil {
				addLog("ERROR: Failed to create cells directory: " + err.Error())
				sharedwidgets.SafeDo(func() {
					validateAllBtn.Enable()
					exportPackageBtn.SetText("Export Package")
				})
				return
			}

			cellCfg := getCellConfig()

			// Step 1: Cell library
			addLog("[1/6] Generating cell library...")
			if err := os.WriteFile(outputDir+"/cells/fecim_bitcell.lef", []byte(export.GenerateLEF(cellCfg)), 0644); err != nil {
				addLog("ERROR: Failed to write LEF: " + err.Error())
			}
			if err := os.WriteFile(outputDir+"/cells/fecim_bitcell.lib", []byte(export.GenerateLiberty(cellCfg)), 0644); err != nil {
				addLog("ERROR: Failed to write LIB: " + err.Error())
			}
			if err := os.WriteFile(outputDir+"/cells/fecim_bitcell.v", []byte(export.GenerateCellVerilog(cellCfg)), 0644); err != nil {
				addLog("ERROR: Failed to write cell Verilog: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)

			// Step 2: Array Verilog
			addLog("[2/6] Generating array Verilog...")
			if err := os.WriteFile(outputDir+"/"+designName+".v", []byte(export.GenerateArrayVerilog(*cfg)), 0644); err != nil {
				addLog("ERROR: Failed to write array Verilog: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)

			// Step 3: DEF
			addLog("[3/6] Generating DEF placement...")
			if err := os.WriteFile(outputDir+"/"+designName+".def", []byte(generateBuilderDEF(*cfg)), 0644); err != nil {
				addLog("ERROR: Failed to write DEF: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)

			// Step 4: Design JSON
			addLog("[4/6] Generating design data...")
			jsonContent := fmt.Sprintf(`{"design": "%s", "rows": %d, "cols": %d, "mode": "%s", "arch": "%s"}`,
				designName, cfg.Rows, cfg.Cols, cfg.Mode, cfg.Architecture)
			if err := os.WriteFile(outputDir+"/"+designName+".json", []byte(jsonContent), 0644); err != nil {
				addLog("ERROR: Failed to write design JSON: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)

			// Step 5: OpenLane config
			addLog("[5/6] Generating OpenLane config...")
			if err := os.WriteFile(outputDir+"/config.json", []byte(export.GenerateOpenLaneConfig(*cfg)), 0644); err != nil {
				addLog("ERROR: Failed to write OpenLane config: " + err.Error())
			}
			time.Sleep(100 * time.Millisecond)

			// Step 6: README
			addLog("[6/6] Generating README...")
			readme := fmt.Sprintf(`# %s

Generated by FeCIM Design Suite
Date: %s

## Files

- cells/ - FeCIM bitcell library (LEF/LIB/V)
- %s.v - Verilog netlist
- %s.def - Physical placement
- config.json - OpenLane configuration

## Usage

1. Copy this directory to your OpenLane designs/
2. Run: flow.tcl -design %s
`, designName, time.Now().Format("2006-01-02"), designName, designName, designName)
			if err := os.WriteFile(outputDir+"/README.md", []byte(readme), 0644); err != nil {
				addLog("ERROR: Failed to write README: " + err.Error())
			}

			// Convert to absolute path for dialog
			absOutputDir, _ := filepath.Abs(outputDir)

			logging.GlobalInfo("[EDA-Builder] Export Package completed: %s", absOutputDir)
			sharedwidgets.SafeDo(func() {
				statusLabel.SetText("Package exported to " + outputDir)
				exportPackageBtn.Enable()
				generateAllBtn.Enable()
				validateAllBtn.Enable()
				exportPackageBtn.SetText("Export Package")

				// Show success dialog with export directory path
				if window != nil {
					message := fmt.Sprintf("Package exported successfully to:\n\n%s\n\nContents:\n• Cell library (LEF/LIB/V)\n• Verilog netlist\n• DEF placement\n• OpenLane config\n• README", absOutputDir)
					dialog.ShowInformation("Export Complete", message, window)
				}
			})
			addLog("\nExport complete: " + outputDir)
		}()
	})

	// ========== BUILD LAYOUT ==========

	// Ultra-compact cell config - 8 columns, single row where possible
	// Set narrower placeholders for entries
	nameEntry.SetPlaceHolder("name")
	widthEntry.SetPlaceHolder("0.46")
	heightEntry.SetPlaceHolder("2.72")
	riseEntry.SetPlaceHolder("10")
	fallEntry.SetPlaceHolder("10")
	capEntry.SetPlaceHolder("0.015")
	leakageEntry.SetPlaceHolder("0.0003")

	// Cell config in two rows of 4 columns to fit within 1024px minimum width
	cellConfigRow1 := container.NewGridWithColumns(4,
		widget.NewLabel("Name"), nameEntry,
		widget.NewLabel("Width"), widthEntry,
	)
	cellConfigRow2 := container.NewGridWithColumns(4,
		widget.NewLabel("Height"), heightEntry,
		widget.NewLabel("Rise"), riseEntry,
	)
	cellConfigRow3 := container.NewGridWithColumns(4,
		widget.NewLabel("Fall"), fallEntry,
		widget.NewLabel("Capacitance"), capEntry,
	)
	cellConfigRow4 := container.NewGridWithColumns(4,
		widget.NewLabel("Leakage"), leakageEntry,
		cellAreaLabel, widget.NewLabel(""),
	)
	cellConfigGrid := container.NewVBox(cellConfigRow1, cellConfigRow2, cellConfigRow3, cellConfigRow4)

	cellPanel := container.NewVBox(
		cellConfigGrid,
	)

	// Ultra-compact array config - combine everything in single row
	arrayConfigRow := container.NewHBox(
		widget.NewLabel("Rows:"), rowsEntry,
		widget.NewLabel("Cols:"), colsEntry,
		widget.NewLabel("Mode:"), modeSelect,
		widget.NewSeparator(),
		widget.NewLabel("Arch:"), archToggle,
	)

	// Horizontal stats in two rows to fit within 1024px minimum width
	statsRow1 := container.NewHBox(
		totalLabel,
		widget.NewLabel("|"),
		areaLabel,
		widget.NewLabel("|"),
		wlLengthLabel,
	)
	statsRow2 := container.NewHBox(
		blLengthLabel,
		widget.NewLabel("|"),
		utilizationLabel,
		widget.NewLabel("|"),
		densityLabel,
	)
	statsRow := container.NewVBox(statsRow1, statsRow2)

	arrayPanel := container.NewVBox(
		arrayConfigRow,
		modeHelpText,
		statsRow,
	)

	// Collapsible config sections using Accordion
	configAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Cell Config", cellPanel),
		widget.NewAccordionItem("Array Config", arrayPanel),
	)
	// Open Array Config by default (more commonly used)
	configAccordion.Open(1)

	// Preview tabs - larger, scrollable containers
	verilogTab := container.NewBorder(
		verilogStatsLabel, nil, nil, nil,
		container.NewScroll(verilogPreview),
	)
	defTab := container.NewBorder(
		defStatsLabel, nil, nil, nil,
		container.NewScroll(defPreview),
	)

	// Layout tab with real EDA tool images only
	layoutHelp := widget.NewLabel("KLayout: auto on Generate All | OpenROAD/Yosys: buttons above")
	layoutScroll := container.NewScroll(layoutStack)
	layoutTab := container.NewBorder(
		container.NewHBox(genSchematicBtn, genOpenROADBtn, layoutHelp),
		nil, nil, nil,
		layoutScroll,
	)

	previewTabs := container.NewAppTabs(
		container.NewTabItem("Verilog", verilogTab),
		container.NewTabItem("DEF", defTab),
		container.NewTabItem("Layout", layoutTab),
	)
	previewTabs.SetTabLocation(container.TabLocationTop)

	// Action buttons - highlight primary actions
	generateAllBtn.Importance = widget.HighImportance
	validateAllBtn.Importance = widget.MediumImportance
	actionButtons := container.NewHBox(
		generateAllBtn,
		validateAllBtn,
		exportPackageBtn,
	)

	// Builder action shortcuts:
	// Cmd/Ctrl+Shift+G => Generate All
	// Cmd/Ctrl+Shift+V => Validate All
	// Cmd/Ctrl+Shift+E => Export Package
	triggerAction := func(btn *widget.Button) {
		if btn == nil || btn.Disabled() || btn.OnTapped == nil {
			return
		}
		btn.OnTapped()
	}
	if window != nil {
		window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyG, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}, func(fyne.Shortcut) {
			triggerAction(generateAllBtn)
		})
		window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}, func(fyne.Shortcut) {
			triggerAction(validateAllBtn)
		})
		window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}, func(fyne.Shortcut) {
			triggerAction(exportPackageBtn)
		})
	}

	// Validation results - compact 4-column single row
	validationResultsGrid := container.NewGridWithColumns(4,
		container.NewHBox(
			widget.NewLabelWithStyle("Yosys:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			yosysResult,
		),
		container.NewHBox(
			widget.NewLabelWithStyle("DEF:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			defResult,
		),
		container.NewHBox(
			widget.NewLabelWithStyle("Cross-Check:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			crossResult,
		),
		container.NewHBox(
			widget.NewLabelWithStyle("Placement:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			placementResult,
		),
	)
	validationHeader := container.NewHBox(
		widget.NewLabelWithStyle("Validation Results:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		validationSummary,
		layout.NewSpacer(),
		widget.NewLabelWithStyle("OpenLane:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dockerStatus,
		pdkStatus,
		pullImageBtn,
	)
	validationRow := container.NewVBox(
		validationHeader,
		container.NewHScroll(validationResultsGrid),
	)

	// Log section with improved visibility
	logOutput.SetMinRowsVisible(6)
	logHeader := container.NewHBox(
		widget.NewLabelWithStyle("Validation Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		clearLogBtn,
	)
	logScroll := container.NewScroll(logOutput)
	logScroll.SetMinSize(fyne.NewSize(0, 140))

	// Bottom validation section - more compact
	validationSection := container.NewVBox(
		widget.NewSeparator(),
		validationRow,
		logHeader,
		logScroll,
	)

	// Status bar - compact inline with actions
	exportConfidence := container.NewHBox(
		widget.NewLabel("Export confidence:"),
		sharedwidgets.NewConfidenceBadge(sharedwidgets.Estimated),
	)
	statusBar := container.NewHBox(
		widget.NewLabel("Status:"),
		statusLabel,
	)

	// Top section: config + actions (compact)
	actionsStatusRow := container.NewHBox(
		actionButtons,
		widget.NewSeparator(),
		statusBar,
		layout.NewSpacer(),
		exportConfidence,
	)
	topSection := container.NewVBox(
		configAccordion,
		widget.NewSeparator(),
		actionsStatusRow,
	)

	// Use VSplit for resizable preview/validation areas
	// Preview gets 65% of space, validation gets 35%
	mainSplit := container.NewVSplit(
		previewTabs,
		validationSection,
	)
	mainSplit.SetOffset(0.55)

	// Main layout: fixed top section, resizable middle/bottom
	mainContent := container.NewBorder(
		topSection,
		nil,
		nil, nil,
		mainSplit,
	)

	return mainContent
}

// generateBuilderDEF generates DEF content for the unified builder tab
// Supports passive, 1T1R, and 2T1R architectures:
//   - passive: WL[], BL[] pins
//   - 1t1r: WL[], BL[], SL[] pins (SL at bottom edge for transistor source)
//   - 2t1r: WL[], BL[], SL[], CSL[] pins (CSL at right edge for column transistor gates)
//
// Includes ROW definitions for OpenROAD placement validation
func generateBuilderDEF(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	is1T1R := cfg.Architecture == "1t1r"
	is2T1R := cfg.Architecture == "2t1r"

	// Determine cell name and site name based on architecture
	cellName := "fecim_bitcell"
	siteName := "fecim_site"
	if is1T1R {
		cellName = "fecim_1t1r_bitcell"
		siteName = "fecim_1t1r_site"
	}
	if is2T1R {
		cellName = "fecim_2t1r_bitcell"
		siteName = "fecim_2t1r_site"
	}

	dbu := 1000 // Database units per micron
	cellWidthDBU := int(cfg.CellWidth * float64(dbu))
	cellHeightDBU := int(cfg.CellHeight * float64(dbu))

	margin := 1000
	dieWidth := cfg.Cols*cellWidthDBU + 2*margin
	dieHeight := cfg.Rows*cellHeightDBU + 2*margin

	var content strings.Builder
	content.WriteString(`VERSION 5.8 ;
DIVIDERCHAR "/" ;
BUSBITCHARS "[]" ;
`)
	content.WriteString(fmt.Sprintf("DESIGN %s ;\n", designName))
	content.WriteString(fmt.Sprintf("UNITS DISTANCE MICRONS %d ;\n\n", dbu))
	content.WriteString(fmt.Sprintf("DIEAREA ( 0 0 ) ( %d %d ) ;\n\n", dieWidth, dieHeight))

	// ROW definitions - required for OpenROAD placement check
	content.WriteString(fmt.Sprintf("ROW ROW_0 %s %d %d N DO %d BY 1 STEP %d 0 ;\n\n",
		siteName, margin, margin, cfg.Cols, cellWidthDBU))
	for row := 1; row < cfg.Rows; row++ {
		y := margin + row*cellHeightDBU
		orient := "N"
		if row%2 == 1 {
			orient = "FS" // Flip alternate rows for proper power grid
		}
		content.WriteString(fmt.Sprintf("ROW ROW_%d %s %d %d %s DO %d BY 1 STEP %d 0 ;\n",
			row, siteName, margin, y, orient, cfg.Cols, cellWidthDBU))
	}
	content.WriteString("\n")

	// Components
	totalCells := cfg.Rows * cfg.Cols
	content.WriteString(fmt.Sprintf("COMPONENTS %d ;\n", totalCells))

	for row := 0; row < cfg.Rows; row++ {
		for col := 0; col < cfg.Cols; col++ {
			x := margin + col*cellWidthDBU
			y := margin + row*cellHeightDBU
			orient := "N"
			if row%2 == 1 {
				orient = "FS"
			}
			content.WriteString(fmt.Sprintf("    - cell_%d_%d %s + FIXED ( %d %d ) %s ;\n", row, col, cellName, x, y, orient))
		}
	}
	content.WriteString("END COMPONENTS\n\n")

	// Pins - add SL for 1T1R/2T1R, add CSL for 2T1R
	numPins := cfg.Rows + cfg.Cols + 2
	if is1T1R {
		numPins += cfg.Cols // Add SL pins (one per column)
	}
	if is2T1R {
		numPins += cfg.Cols // Add SL pins (one per column)
		numPins += cfg.Cols // Add CSL pins (one per column)
	}
	content.WriteString(fmt.Sprintf("PINS %d ;\n", numPins))
	content.WriteString("    - VPWR + NET VPWR + DIRECTION INOUT + USE POWER ;\n")
	content.WriteString("    - VGND + NET VGND + DIRECTION INOUT + USE GROUND ;\n")
	for i := 0; i < cfg.Rows; i++ {
		content.WriteString(fmt.Sprintf("    - WL[%d] + NET WL[%d] + DIRECTION INPUT + USE SIGNAL ;\n", i, i))
	}
	for i := 0; i < cfg.Cols; i++ {
		content.WriteString(fmt.Sprintf("    - BL[%d] + NET BL[%d] + DIRECTION OUTPUT + USE SIGNAL ;\n", i, i))
	}
	if is1T1R || is2T1R {
		// SL pins at bottom edge (opposite from BL) - one per column for transistor source
		for i := 0; i < cfg.Cols; i++ {
			content.WriteString(fmt.Sprintf("    - SL[%d] + NET SL[%d] + DIRECTION INPUT + USE SIGNAL ;\n", i, i))
		}
	}
	if is2T1R {
		// CSL pins at right edge - one per column for column transistor gates
		for i := 0; i < cfg.Cols; i++ {
			content.WriteString(fmt.Sprintf("    - CSL[%d] + NET CSL[%d] + DIRECTION INPUT + USE SIGNAL ;\n", i, i))
		}
	}
	content.WriteString("END PINS\n\n")

	content.WriteString("END DESIGN\n")
	return content.String()
}
