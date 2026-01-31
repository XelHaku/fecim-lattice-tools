// pkg/validation/circuit_image.go
// Circuit visualization using Yosys show command and OpenROAD save_image
// Generates schematic SVG and layout PNG from EDA files

package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/openlane"
)

// Note: log variable is defined in layout_image.go via init()

// CircuitImageResult contains the result of circuit image generation
type CircuitImageResult struct {
	Success   bool
	ImagePath string
	RawOutput string
	Error     string
}

// GenerateYosysSchematic creates a circuit schematic SVG using Yosys show command
// Requires: Verilog file
// Output: SVG schematic diagram
func GenerateYosysSchematic(verilogPath string, outputPrefix string, topModule string, architecture string, manager *openlane.Manager, config *openlane.Config) (*CircuitImageResult, error) {
	result := &CircuitImageResult{
		Success:   false,
		ImagePath: outputPrefix + ".dot", // DOT format - graphviz text file
	}

	log.Info("=== Yosys Schematic Generation ===")
	log.Info("  Verilog: %s", verilogPath)
	log.Info("  Output prefix: %s", outputPrefix)
	log.Info("  Top module: %s", topModule)

	// Check if Verilog file exists
	if _, err := os.Stat(verilogPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("Verilog file not found: %s", verilogPath)
		log.Printf("Yosys: %s", result.Error)
		return result, nil
	}

	// Check mode
	mode := manager.DetectMode()
	log.Info("  Mode: %s", mode)
	if mode == openlane.ModeNone {
		result.Error = "Yosys not available (install Docker with OpenLane image or native Yosys)"
		log.Printf("Yosys: %s", result.Error)
		return result, nil
	}

	workDir := filepath.Dir(verilogPath)
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get absolute path: %v", err)
		return result, nil
	}

	// Yosys command to generate schematic
	// show -format svg -prefix <name> -viewer none
	verilogName := filepath.Base(verilogPath)
	outputName := filepath.Base(outputPrefix)

	// Determine cell directory and filename based on architecture
	cellDir := "fecim_bitcell"
	cellFileName := "fecim_bitcell.v"
	switch strings.ToLower(architecture) {
	case "1t1r":
		cellDir = "fecim_1t1r_bitcell"
		cellFileName = "fecim_1t1r_bitcell.v"
	case "2t1r":
		cellDir = "fecim_2t1r_bitcell"
		cellFileName = "fecim_2t1r_bitcell.v"
	}

	// Also need to read the cell Verilog file for hierarchy to work
	// Try to find the cell file in cells/ directory
	cellVerilog := ""
	cellPaths := []string{
		filepath.Join(workDir, "../cells", cellDir, cellFileName),
		filepath.Join(workDir, "../../cells", cellDir, cellFileName),
		filepath.Join("cells", cellDir, cellFileName),
	}
	for _, cp := range cellPaths {
		if absCP, err := filepath.Abs(cp); err == nil {
			if _, err := os.Stat(absCP); err == nil {
				// Copy cell verilog to work directory for Docker access
				cellData, _ := os.ReadFile(absCP)
				cellDst := filepath.Join(absWorkDir, cellFileName)
				os.WriteFile(cellDst, cellData, 0644)
				cellVerilog = cellFileName
				log.Info("  Cell Verilog found: %s", absCP)
				break
			}
		}
	}

	// Build Yosys command - read cell first, then array
	// Use 'dot' format which doesn't require graphviz binary, then we handle conversion
	// For formats other than dot/ps, Yosys requires selecting only one module
	var yosysCmd string
	if cellVerilog != "" {
		// Use -format dot to avoid needing graphviz, Yosys writes DOT file directly
		yosysCmd = fmt.Sprintf(
			"read_verilog %s; read_verilog %s; hierarchy -check -top %s; select %s; show -format dot -prefix %s -viewer none",
			cellVerilog, verilogName, topModule, topModule, outputName,
		)
	} else {
		// Fallback: try without cell (will show blackbox)
		log.Info("  Warning: Cell Verilog not found, schematic may show blackboxes")
		yosysCmd = fmt.Sprintf(
			"read_verilog %s; hierarchy -auto-top; show -format dot -prefix %s -viewer none",
			verilogName, outputName,
		)
	}

	log.Info("  Yosys command: %s", yosysCmd)
	log.Info("  Running Yosys...")
	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunYosys(yosysCmd, absWorkDir)

	if runResult != nil {
		result.RawOutput = runResult.Stdout + "\n" + runResult.Stderr
		// Log output for debugging
		if runResult.Stdout != "" {
			for _, line := range strings.Split(runResult.Stdout, "\n") {
				if line != "" {
					log.Info("  [Yosys stdout] %s", line)
				}
			}
		}
		if runResult.Stderr != "" {
			for _, line := range strings.Split(runResult.Stderr, "\n") {
				if line != "" {
					log.Info("  [Yosys stderr] %s", line)
				}
			}
		}
		log.Info("  [Yosys] Exit code: %d, Duration: %v", runResult.ExitCode, runResult.Duration)
	}

	if err != nil {
		result.Error = fmt.Sprintf("Yosys execution failed: %v", err)
		log.Printf("Yosys error: %v", err)
		log.Printf("Yosys raw output:\n%s", result.RawOutput)
		return result, nil
	}

	// Check if output file was created (Yosys adds .dot extension)
	dotOutput := filepath.Join(absWorkDir, outputName+".dot")
	if _, err := os.Stat(dotOutput); os.IsNotExist(err) {
		result.Error = "Yosys did not produce DOT schematic"
		log.Printf("Yosys: %s (expected: %s)", result.Error, dotOutput)
		return result, nil
	}

	log.Info("  DOT file generated: %s", dotOutput)

	// Convert DOT to PNG using local graphviz (if available)
	pngOutput := filepath.Join(absWorkDir, outputName+".png")
	log.Info("  Converting DOT to PNG...")

	dotCmd := exec.Command("dot", "-Tpng", dotOutput, "-o", pngOutput)
	dotOut, dotErr := dotCmd.CombinedOutput()
	if dotErr != nil {
		// Graphviz not installed or failed - return DOT file path with warning
		log.Printf("  Warning: graphviz conversion failed: %v", dotErr)
		log.Printf("  Install graphviz: sudo apt install graphviz")
		log.Printf("  DOT output: %s", string(dotOut))
		result.Success = true
		result.ImagePath = dotOutput // Return DOT file, UI will show message
		result.Error = "DOT file created but PNG conversion failed (install graphviz)"
		return result, nil
	}

	// PNG conversion successful
	result.Success = true
	result.ImagePath = pngOutput
	log.Info("  Yosys schematic PNG generated: %s", pngOutput)
	return result, nil
}

// openroadImageScript is the TCL script for OpenROAD image export
const openroadImageScript = `# save_layout_image.tcl - OpenROAD layout image export
# Environment: CELL_LEF, DEF_FILE, OUTPUT_PNG

puts "=== OpenROAD Layout Image Export ==="

# Read LEF and DEF
read_lef $env(CELL_LEF)
read_def $env(DEF_FILE)

puts "Design loaded, saving image..."

# Save layout image
save_image $env(OUTPUT_PNG)

puts "=== Image Export Complete ==="
exit
`

// GenerateOpenROADImage creates a layout PNG using OpenROAD save_image command
// Requires: LEF and DEF files
// Output: PNG layout image
func GenerateOpenROADImage(defPath string, lefPath string, outputPath string, manager *openlane.Manager, config *openlane.Config) (*CircuitImageResult, error) {
	result := &CircuitImageResult{
		Success:   false,
		ImagePath: outputPath,
	}

	log.Info("=== OpenROAD Image Generation ===")
	log.Info("  DEF: %s", defPath)
	log.Info("  LEF: %s", lefPath)
	log.Info("  Output: %s", outputPath)

	// Check if files exist
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("DEF file not found: %s", defPath)
		log.Printf("OpenROAD: %s", result.Error)
		return result, nil
	}
	if _, err := os.Stat(lefPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("LEF file not found: %s", lefPath)
		log.Printf("OpenROAD: %s", result.Error)
		return result, nil
	}

	// Check mode
	mode := manager.DetectMode()
	log.Info("  Mode: %s", mode)
	if mode == openlane.ModeNone {
		result.Error = "OpenROAD not available (install Docker with OpenLane image or native OpenROAD)"
		log.Printf("OpenROAD: %s", result.Error)
		return result, nil
	}

	workDir := filepath.Dir(defPath)
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get absolute path: %v", err)
		return result, nil
	}

	// Write TCL script
	scriptPath := filepath.Join(absWorkDir, "save_layout_image.tcl")
	if err := os.WriteFile(scriptPath, []byte(openroadImageScript), 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write TCL script: %v", err)
		return result, nil
	}
	defer os.Remove(scriptPath)

	// Copy LEF to work directory if needed
	lefName := filepath.Base(lefPath)
	lefDst := filepath.Join(absWorkDir, lefName)
	if lefPath != lefDst {
		if lefData, err := os.ReadFile(lefPath); err == nil {
			os.WriteFile(lefDst, lefData, 0644)
			defer os.Remove(lefDst)
		}
	}

	// Set up environment variables
	var envVars map[string]string
	outputName := filepath.Base(outputPath)
	if mode == openlane.ModeDocker {
		envVars = map[string]string{
			"DEF_FILE":   "/design/" + filepath.Base(defPath),
			"CELL_LEF":   "/design/" + lefName,
			"OUTPUT_PNG": "/design/" + outputName,
		}
	} else {
		envVars = map[string]string{
			"DEF_FILE":   filepath.Join(absWorkDir, filepath.Base(defPath)),
			"CELL_LEF":   lefDst,
			"OUTPUT_PNG": filepath.Join(absWorkDir, outputName),
		}
	}

	// Run OpenROAD
	log.Info("  Running OpenROAD...")
	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunOpenROAD("save_layout_image.tcl", absWorkDir, envVars)

	if runResult != nil {
		result.RawOutput = runResult.Stdout + "\n" + runResult.Stderr
		// Log output for debugging
		if runResult.Stdout != "" {
			for _, line := range strings.Split(runResult.Stdout, "\n") {
				if line != "" {
					log.Info("  [OpenROAD stdout] %s", line)
				}
			}
		}
		if runResult.Stderr != "" {
			for _, line := range strings.Split(runResult.Stderr, "\n") {
				if line != "" {
					log.Info("  [OpenROAD stderr] %s", line)
				}
			}
		}
		log.Info("  [OpenROAD] Exit code: %d, Duration: %v", runResult.ExitCode, runResult.Duration)
	}

	if err != nil {
		result.Error = fmt.Sprintf("OpenROAD execution failed: %v", err)
		log.Printf("OpenROAD error: %v", err)
		log.Printf("OpenROAD raw output:\n%s", result.RawOutput)
		return result, nil
	}

	// Check if output file was created
	expectedOutput := filepath.Join(absWorkDir, outputName)
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		result.Error = "OpenROAD did not produce layout image"
		log.Printf("OpenROAD: %s (expected: %s)", result.Error, expectedOutput)
		return result, nil
	}

	// Move to final output path if different
	if expectedOutput != outputPath {
		os.Rename(expectedOutput, outputPath)
	}

	result.Success = true
	result.ImagePath = outputPath
	log.Info("  OpenROAD image generated: %s", outputPath)
	return result, nil
}
