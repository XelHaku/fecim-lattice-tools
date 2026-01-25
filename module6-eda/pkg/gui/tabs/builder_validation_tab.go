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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/openlane"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/validation"
)

// MakeBuilderValidationTab creates a unified tab combining cell/array configuration,
// preview (Verilog/DEF/Layout), validation, and export functionality
func MakeBuilderValidationTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	// ========== CELL CONFIG SECTION ==========
	nameEntry := widget.NewEntry()
	nameEntry.SetText("fecim_bitcell")

	widthEntry := widget.NewEntry()
	widthEntry.SetText("0.460")

	heightEntry := widget.NewEntry()
	heightEntry.SetText("2.720")


	riseEntry := widget.NewEntry()
	riseEntry.SetText("0.1")

	fallEntry := widget.NewEntry()
	fallEntry.SetText("0.1")

	capEntry := widget.NewEntry()
	capEntry.SetText("0.002")

	leakageEntry := widget.NewEntry()
	leakageEntry.SetText("0.001")

	// Helper to parse cell config from inputs
	getCellConfig := func() config.CellConfig {
		width, _ := strconv.ParseFloat(widthEntry.Text, 64)
		height, _ := strconv.ParseFloat(heightEntry.Text, 64)
		rise, _ := strconv.ParseFloat(riseEntry.Text, 64)
		fall, _ := strconv.ParseFloat(fallEntry.Text, 64)
		cap, _ := strconv.ParseFloat(capEntry.Text, 64)
		leakage, _ := strconv.ParseFloat(leakageEntry.Text, 64)

		return config.CellConfig{
			Name:         nameEntry.Text,
			Width:        width,
			Height:       height,
			CellType:     cfg.Architecture, // Uses shared architecture selector
			Technology:   "sky130",
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

	archSelect := widget.NewSelect([]string{"passive", "1t1r"}, func(s string) {
		cfg.Architecture = s
	})
	archSelect.SetSelected(cfg.Architecture)

	// Statistics labels
	totalLabel := widget.NewLabel(fmt.Sprintf("Total Cells: %d", cfg.Rows*cfg.Cols))
	areaLabel := widget.NewLabel(fmt.Sprintf("Array Area: %.2f µm²", float64(cfg.Rows*cfg.Cols)*cfg.CellWidth*cfg.CellHeight))
	wlLengthLabel := widget.NewLabel(fmt.Sprintf("WL Length: %.2f µm", float64(cfg.Cols)*cfg.CellWidth))
	blLengthLabel := widget.NewLabel(fmt.Sprintf("BL Length: %.2f µm", float64(cfg.Rows)*cfg.CellHeight))

	// Update statistics function
	updateStats := func() {
		rows, _ := strconv.Atoi(rowsEntry.Text)
		cols, _ := strconv.Atoi(colsEntry.Text)
		cfg.Rows = rows
		cfg.Cols = cols

		// Update cell dimensions from entries
		cellW, _ := strconv.ParseFloat(widthEntry.Text, 64)
		cellH, _ := strconv.ParseFloat(heightEntry.Text, 64)
		cfg.CellWidth = cellW
		cfg.CellHeight = cellH

		total := rows * cols
		area := float64(total) * cfg.CellWidth * cfg.CellHeight
		wlLength := float64(cols) * cfg.CellWidth
		blLength := float64(rows) * cfg.CellHeight

		totalLabel.SetText(fmt.Sprintf("Total Cells: %d", total))
		areaLabel.SetText(fmt.Sprintf("Array Area: %.2f µm²", area))
		wlLengthLabel.SetText(fmt.Sprintf("WL Length: %.2f µm", wlLength))
		blLengthLabel.SetText(fmt.Sprintf("BL Length: %.2f µm", blLength))
		cellAreaLabel.SetText(fmt.Sprintf("Cell Area: %.4f µm²", cellW*cellH))
	}

	// Wire up change handlers
	rowsEntry.OnChanged = func(s string) { updateStats() }
	colsEntry.OnChanged = func(s string) { updateStats() }
	widthEntry.OnChanged = func(s string) { updateStats() }
	heightEntry.OnChanged = func(s string) { updateStats() }

	// ========== PREVIEW SECTION ==========
	verilogPreview := widget.NewMultiLineEntry()
	verilogPreview.Wrapping = fyne.TextWrapOff

	defPreview := widget.NewMultiLineEntry()
	defPreview.Wrapping = fyne.TextWrapOff

	layoutViz := widget.NewLabel("Generate to see layout")
	layoutViz.TextStyle.Monospace = true

	verilogStatsLabel := widget.NewLabel("Verilog: Pending")
	defStatsLabel := widget.NewLabel("DEF: Pending")

	// ========== VALIDATION SECTION ==========
	yosysResult := widget.NewLabel("Not validated")
	defResult := widget.NewLabel("Not validated")
	crossResult := widget.NewLabel("Not validated")
	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord

	addLog := func(msg string) {
		fyne.Do(func() {
			logOutput.SetText(logOutput.Text + msg + "\n")
		})
	}

	// ========== OPENLANE STATUS SECTION ==========
	dockerStatus := widget.NewLabel("Checking...")
	pdkStatus := widget.NewLabel("Checking...")
	placementResult := widget.NewLabel("Not validated")

	// Check OpenLane status on startup
	go func() {
		manager := openlane.NewManager()
		mode := manager.DetectMode()

		fyne.Do(func() {
			if mode == openlane.ModeDocker {
				if manager.IsDockerImagePulled() {
					dockerStatus.SetText("✓ Docker image ready")
				} else {
					dockerStatus.SetText("○ Docker image not pulled")
				}
			} else if mode == openlane.ModeNative {
				dockerStatus.SetText("✓ Native tools detected")
			} else {
				dockerStatus.SetText("✗ OpenLane not available")
			}

			if manager.IsPDKInstalled() {
				pdkStatus.SetText("✓ SKY130A PDK ready")
			} else {
				pdkStatus.SetText("○ PDK not installed (run: volare enable --pdk sky130 sky130A)")
			}
		})
	}()

	// Pull Docker Image button (only shown when needed)
	pullImageBtn := widget.NewButton("Pull OpenLane Image", func() {
		go func() {
			fyne.Do(func() {
				dockerStatus.SetText("Pulling image...")
			})
			addLog("=== Pulling OpenLane Docker Image ===")
			addLog("This may take several minutes...")

			manager := openlane.NewManager()
			err := manager.PullDockerImage(func(msg string) {
				addLog(msg)
			})
			if err != nil {
				fyne.Do(func() {
					dockerStatus.SetText("✗ Pull failed: " + err.Error())
				})
				addLog("ERROR: " + err.Error())
			} else {
				fyne.Do(func() {
					dockerStatus.SetText("✓ Docker image ready")
				})
				addLog("Docker image pulled successfully")
			}
		}()
	})

	// Placement validation checkbox
	enablePlacementCheck := widget.NewCheck("Enable OpenLane Placement Check", nil)

	// ========== STATUS ==========
	statusLabel := widget.NewLabel("Ready")

	// ========== GENERATE ALL BUTTON ==========
	var generateAllBtn *widget.Button
	var validateAllBtn *widget.Button
	var exportPackageBtn *widget.Button

	generateAllBtn = widget.NewButton("Generate All", func() {
		generateAllBtn.Disable()
		validateAllBtn.Disable()
		exportPackageBtn.Disable()
		generateAllBtn.SetText("Generating...")
		go func() {
			fyne.Do(func() {
				statusLabel.SetText("Generating...")
				logOutput.SetText("")
			})

			updateStats()
			cellCfg := getCellConfig()

			// Generate cell files
			addLog("Generating cell library...")
			lefContent := export.GenerateLEF(cellCfg)
			libContent := export.GenerateLiberty(cellCfg)
			cellVContent := export.GenerateCellVerilog(cellCfg)

			dir := "cells/fecim_bitcell"
			os.MkdirAll(dir, 0755)
			os.WriteFile(dir+"/fecim_bitcell.lef", []byte(lefContent), 0644)
			os.WriteFile(dir+"/fecim_bitcell.lib", []byte(libContent), 0644)
			os.WriteFile(dir+"/fecim_bitcell.v", []byte(cellVContent), 0644)
			addLog("  LEF/LIB/V written to " + dir)

			// Generate array Verilog
			addLog("Generating array Verilog...")
			vContent := export.GenerateArrayVerilog(*cfg)
			fyne.Do(func() {
				verilogPreview.SetText(vContent)
			})

			instances := cfg.Rows * cfg.Cols
			lines := strings.Count(vContent, "\n")
			size := float64(len(vContent)) / 1024
			fyne.Do(func() {
				verilogStatsLabel.SetText(fmt.Sprintf("Instances: %d | Lines: %d | Size: %.1fKB", instances, lines, size))
			})

			vFilename := fmt.Sprintf("output/exports/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
			os.MkdirAll("output/exports", 0755)
			os.WriteFile(vFilename, []byte(vContent), 0644)
			addLog("  Verilog: " + vFilename)

			// Generate DEF
			addLog("Generating DEF placement...")
			defContent := generateBuilderDEF(*cfg)
			fyne.Do(func() {
				defPreview.SetText(defContent)
			})

			defFilename := fmt.Sprintf("output/exports/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
			os.WriteFile(defFilename, []byte(defContent), 0644)
			fyne.Do(func() {
				defStatsLabel.SetText(fmt.Sprintf("Components: %d | File: %s", cfg.Rows*cfg.Cols, defFilename))
			})
			addLog("  DEF: " + defFilename)

			// Update layout visualization
			vizText := makeBuilderLayoutVisualization(cfg)
			fyne.Do(func() {
				layoutViz.SetText(vizText)
			})

			// Generate OpenLane config
			addLog("Generating OpenLane config...")
			configContent := export.GenerateOpenLaneConfig(*cfg)
			os.WriteFile("output/exports/config.json", []byte(configContent), 0644)
			addLog("  Config: output/exports/config.json")

			fyne.Do(func() {
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
		validateAllBtn.Disable()
		generateAllBtn.Disable()
		exportPackageBtn.Disable()
		validateAllBtn.SetText("Validating...")
		go func() {
			fyne.Do(func() {
				statusLabel.SetText("Validating...")
				logOutput.SetText("")
				yosysResult.SetText("...")
				defResult.SetText("...")
				crossResult.SetText("...")
			})

			allPassed := true

			// Yosys validation
			addLog("=== Yosys Verilog Validation ===")
			arrayPath := fmt.Sprintf("output/exports/fecim_crossbar_%dx%d.v", cfg.Rows, cfg.Cols)
			cellPath := "cells/fecim_bitcell/fecim_bitcell.v"
			addLog(fmt.Sprintf("Array: %s", arrayPath))
			addLog(fmt.Sprintf("Cell:  %s", cellPath))

			err1 := validation.ValidateVerilogWithCell(arrayPath, cellPath)
			if err1 != nil {
				fyne.Do(func() { yosysResult.SetText("FAIL") })
				addLog(fmt.Sprintf("ERROR: %v", err1))
				allPassed = false
			} else {
				fyne.Do(func() { yosysResult.SetText("PASS") })
				addLog("PASSED")
			}

			// DEF validation
			addLog("\n=== DEF Syntax Validation ===")
			defPath := fmt.Sprintf("output/exports/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)
			addLog(fmt.Sprintf("DEF: %s", defPath))

			err2 := validation.ValidateDEF(defPath)
			if err2 != nil {
				fyne.Do(func() { defResult.SetText("FAIL") })
				addLog(fmt.Sprintf("ERROR: %v", err2))
				allPassed = false
			} else {
				fyne.Do(func() { defResult.SetText("PASS") })
				stats, _ := validation.GetDEFStats(defPath)
				addLog(fmt.Sprintf("Design: %v, Components: %v", stats["design_name"], stats["component_count"]))
				addLog("PASSED")
			}

			// Cross-check validation
			addLog("\n=== LEF/LIB/V Cross-Check ===")
			lefPath := "cells/fecim_bitcell/fecim_bitcell.lef"
			libPath := "cells/fecim_bitcell/fecim_bitcell.lib"
			vPath := "cells/fecim_bitcell/fecim_bitcell.v"
			addLog(fmt.Sprintf("LEF: %s", lefPath))
			addLog(fmt.Sprintf("LIB: %s", libPath))
			addLog(fmt.Sprintf("V:   %s", vPath))

			err3 := validation.CrossCheckFiles(lefPath, libPath, vPath)
			if err3 != nil {
				fyne.Do(func() { crossResult.SetText("FAIL") })
				addLog(fmt.Sprintf("ERROR: %v", err3))
				allPassed = false
			} else {
				fyne.Do(func() { crossResult.SetText("PASS") })
				addLog("Pin names and cell names match")
				addLog("PASSED")
			}

			// OpenLane Placement validation (if enabled and available)
			if enablePlacementCheck.Checked {
				addLog("\n=== OpenLane Placement Validation ===")
				manager := openlane.NewManager()
				mode := manager.DetectMode()

				if mode == openlane.ModeNone {
					fyne.Do(func() { placementResult.SetText("SKIP - No OpenLane") })
					addLog("SKIPPED: OpenLane not available")
				} else {
					defPath := fmt.Sprintf("output/exports/fecim_crossbar_%dx%d.def", cfg.Rows, cfg.Cols)

					addLog(fmt.Sprintf("Mode: %s", mode))
					addLog(fmt.Sprintf("DEF: %s", defPath))

					config := openlane.DefaultConfig()
					result, err := validation.RunPlacementCheck(defPath, manager, config)
					if err != nil {
						fyne.Do(func() { placementResult.SetText("ERROR") })
						addLog(fmt.Sprintf("ERROR: %v", err))
						allPassed = false
					} else if result.Passed {
						fyne.Do(func() { placementResult.SetText("PASS") })
						addLog(result.RawOutput)
						addLog("PASSED")
					} else {
						fyne.Do(func() { placementResult.SetText("FAIL") })
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
				fyne.Do(func() {
					statusLabel.SetText("All validations passed")
					validateAllBtn.Enable()
					generateAllBtn.Enable()
					exportPackageBtn.Enable()
					validateAllBtn.SetText("Validate All")
				})
				addLog("\n=== ALL VALIDATIONS PASSED ===")
			} else {
				fyne.Do(func() {
					statusLabel.SetText("Some validations failed")
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
			fyne.Do(func() {
				statusLabel.SetText("Exporting package...")
				logOutput.SetText("")
			})

			designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
			outputDir := fmt.Sprintf("output/exports/%s", designName)
			os.MkdirAll(outputDir, 0755)
			os.MkdirAll(outputDir+"/cells", 0755)

			cellCfg := getCellConfig()

			// Step 1: Cell library
			addLog("[1/6] Generating cell library...")
			os.WriteFile(outputDir+"/cells/fecim_bitcell.lef", []byte(export.GenerateLEF(cellCfg)), 0644)
			os.WriteFile(outputDir+"/cells/fecim_bitcell.lib", []byte(export.GenerateLiberty(cellCfg)), 0644)
			os.WriteFile(outputDir+"/cells/fecim_bitcell.v", []byte(export.GenerateCellVerilog(cellCfg)), 0644)
			time.Sleep(100 * time.Millisecond)

			// Step 2: Array Verilog
			addLog("[2/6] Generating array Verilog...")
			os.WriteFile(outputDir+"/"+designName+".v", []byte(export.GenerateArrayVerilog(*cfg)), 0644)
			time.Sleep(100 * time.Millisecond)

			// Step 3: DEF
			addLog("[3/6] Generating DEF placement...")
			os.WriteFile(outputDir+"/"+designName+".def", []byte(generateBuilderDEF(*cfg)), 0644)
			time.Sleep(100 * time.Millisecond)

			// Step 4: Design JSON
			addLog("[4/6] Generating design data...")
			jsonContent := fmt.Sprintf(`{"design": "%s", "rows": %d, "cols": %d, "mode": "%s", "arch": "%s"}`,
				designName, cfg.Rows, cfg.Cols, cfg.Mode, cfg.Architecture)
			os.WriteFile(outputDir+"/"+designName+".json", []byte(jsonContent), 0644)
			time.Sleep(100 * time.Millisecond)

			// Step 5: OpenLane config
			addLog("[5/6] Generating OpenLane config...")
			os.WriteFile(outputDir+"/config.json", []byte(export.GenerateOpenLaneConfig(*cfg)), 0644)
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
			os.WriteFile(outputDir+"/README.md", []byte(readme), 0644)

			// Convert to absolute path for dialog
			absOutputDir, _ := filepath.Abs(outputDir)

			fyne.Do(func() {
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

	// Cell config form
	cellForm := widget.NewForm(
		widget.NewFormItem("Cell Name", nameEntry),
		widget.NewFormItem("Width (µm)", widthEntry),
		widget.NewFormItem("Height (µm)", heightEntry),
		widget.NewFormItem("Rise (ns)", riseEntry),
		widget.NewFormItem("Fall (ns)", fallEntry),
		widget.NewFormItem("Cap (pF)", capEntry),
		widget.NewFormItem("Leakage (nW)", leakageEntry),
	)

	cellPanel := container.NewVBox(
		widget.NewLabelWithStyle("Cell Config", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		cellForm,
		cellAreaLabel,
	)

	// Array config form
	arrayForm := widget.NewForm(
		widget.NewFormItem("Rows", rowsEntry),
		widget.NewFormItem("Columns", colsEntry),
		widget.NewFormItem("Mode", modeSelect),
		widget.NewFormItem("Architecture", archSelect),
	)

	statsBox := container.NewVBox(
		widget.NewLabelWithStyle("Statistics", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		totalLabel,
		areaLabel,
		wlLengthLabel,
		blLengthLabel,
	)

	arrayPanel := container.NewVBox(
		widget.NewLabelWithStyle("Array Config", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		arrayForm,
		modeHelpText,
		widget.NewSeparator(),
		statsBox,
	)

	// Config panels (left/right split)
	configSplit := container.NewHSplit(cellPanel, arrayPanel)
	configSplit.SetOffset(0.5)

	// Preview tabs
	verilogTab := container.NewBorder(
		verilogStatsLabel, nil, nil, nil,
		container.NewScroll(verilogPreview),
	)
	defTab := container.NewBorder(
		defStatsLabel, nil, nil, nil,
		container.NewScroll(defPreview),
	)
	layoutTab := container.NewScroll(layoutViz)

	previewTabs := container.NewAppTabs(
		container.NewTabItem("Verilog", verilogTab),
		container.NewTabItem("DEF", defTab),
		container.NewTabItem("Layout", layoutTab),
	)

	// Action buttons
	actionButtons := container.NewHBox(
		generateAllBtn,
		validateAllBtn,
		exportPackageBtn,
	)

	// Validation results (compact row)
	validationRow := container.NewHBox(
		widget.NewLabel("Yosys:"), yosysResult,
		widget.NewLabel(" | DEF:"), defResult,
		widget.NewLabel(" | Cross:"), crossResult,
		widget.NewLabel(" | Placement:"), placementResult,
	)

	// OpenLane status panel
	openLanePanel := container.NewVBox(
		widget.NewLabelWithStyle("OpenLane Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(widget.NewLabel("Docker:"), dockerStatus),
		container.NewHBox(widget.NewLabel("PDK:"), pdkStatus),
		pullImageBtn,
		enablePlacementCheck,
	)

	// Bottom section with validation
	validationSection := container.NewVBox(
		widget.NewSeparator(),
		container.NewHSplit(
			container.NewVBox(
				widget.NewLabelWithStyle("Validation Results", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				validationRow,
			),
			openLanePanel,
		),
		container.NewScroll(logOutput),
	)

	// Status bar
	statusBar := container.NewHBox(
		widget.NewLabel("Status:"),
		statusLabel,
	)

	// Top section: config + actions
	topSection := container.NewVBox(
		configSplit,
		widget.NewSeparator(),
		actionButtons,
		statusBar,
	)

	// Main layout: top (config/actions) | middle (preview) | bottom (validation)
	mainContent := container.NewBorder(
		topSection,
		validationSection,
		nil, nil,
		previewTabs,
	)

	return mainContent
}

// generateBuilderDEF generates DEF content for the unified builder tab
func generateBuilderDEF(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)

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

	// Components
	totalCells := cfg.Rows * cfg.Cols
	content.WriteString(fmt.Sprintf("COMPONENTS %d ;\n", totalCells))

	for row := 0; row < cfg.Rows; row++ {
		for col := 0; col < cfg.Cols; col++ {
			x := margin + col*cellWidthDBU
			y := margin + row*cellHeightDBU
			content.WriteString(fmt.Sprintf("    - cell_%d_%d fecim_bitcell + FIXED ( %d %d ) N ;\n", row, col, x, y))
		}
	}
	content.WriteString("END COMPONENTS\n\n")

	// Pins
	numPins := cfg.Rows + cfg.Cols + 2
	content.WriteString(fmt.Sprintf("PINS %d ;\n", numPins))
	content.WriteString("    - VPWR + NET VPWR + DIRECTION INOUT + USE POWER ;\n")
	content.WriteString("    - VGND + NET VGND + DIRECTION INOUT + USE GROUND ;\n")
	for i := 0; i < cfg.Rows; i++ {
		content.WriteString(fmt.Sprintf("    - WL[%d] + NET WL[%d] + DIRECTION INPUT + USE SIGNAL ;\n", i, i))
	}
	for i := 0; i < cfg.Cols; i++ {
		content.WriteString(fmt.Sprintf("    - BL[%d] + NET BL[%d] + DIRECTION OUTPUT + USE SIGNAL ;\n", i, i))
	}
	content.WriteString("END PINS\n\n")

	content.WriteString("END DESIGN\n")
	return content.String()
}

// makeBuilderLayoutVisualization creates a text-based visualization of the array layout
func makeBuilderLayoutVisualization(cfg *config.ArrayConfig) string {
	var viz strings.Builder

	rows := cfg.Rows
	cols := cfg.Cols
	if rows > 12 {
		rows = 12
	}
	if cols > 8 {
		cols = 8
	}

	viz.WriteString(fmt.Sprintf("FeCIM Crossbar %dx%d Layout\n\n", cfg.Rows, cfg.Cols))

	for r := 0; r < rows; r++ {
		viz.WriteString(fmt.Sprintf("WL[%d] ", r))
		for c := 0; c < cols; c++ {
			viz.WriteString("[=]")
		}
		viz.WriteString("\n      ")
		for c := 0; c < cols; c++ {
			viz.WriteString(" | ")
		}
		viz.WriteString("\n")
	}

	if cfg.Rows > 12 {
		viz.WriteString(fmt.Sprintf("      ... (%d more rows)\n", cfg.Rows-12))
	}

	viz.WriteString("      ")
	for c := 0; c < cols; c++ {
		viz.WriteString(fmt.Sprintf("BL%d ", c))
	}
	viz.WriteString("\n")

	return viz.String()
}
