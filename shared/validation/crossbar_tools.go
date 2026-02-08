// Package validation provides CLI detection and validation for external crossbar simulation tools.
// It supports CrossSim (Sandia National Labs) and BadCrossbar (UCL) for crossbar array validation.
package validation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ToolStatus represents the installation status of an external tool.
type ToolStatus int

const (
	// StatusUnknown means the tool status has not been checked.
	StatusUnknown ToolStatus = iota
	// StatusInstalled means the tool is installed and available.
	StatusInstalled
	// StatusNotInstalled means the tool is not installed.
	StatusNotInstalled
	// StatusError means there was an error checking the tool.
	StatusError
)

// String returns a human-readable status string.
func (s ToolStatus) String() string {
	switch s {
	case StatusInstalled:
		return "Installed"
	case StatusNotInstalled:
		return "Not Installed"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Symbol returns a status symbol for UI display.
func (s ToolStatus) Symbol() string {
	switch s {
	case StatusInstalled:
		return "✓"
	case StatusNotInstalled:
		return "✗"
	case StatusError:
		return "⚠"
	default:
		return "○"
	}
}

// ToolInfo contains information about an external tool.
type ToolInfo struct {
	Name        string
	Status      ToolStatus
	Version     string
	Error       string
	Description string
	InstallCmd  string
	DocURL      string
}

// CrossSimInfo returns information about CrossSim availability.
// CrossSim is a GPU-accelerated crossbar simulator from Sandia National Labs.
func CrossSimInfo() *ToolInfo {
	info := &ToolInfo{
		Name:        "CrossSim",
		Description: "Sandia National Labs GPU-accelerated crossbar simulator",
		InstallCmd:  "git clone https://github.com/sandialabs/cross-sim && pip install -e ./cross-sim",
		DocURL:      "https://github.com/sandialabs/cross-sim",
	}

	status, version, err := checkPythonModule("crosssim")
	info.Status = status
	info.Version = version
	if err != nil {
		info.Error = err.Error()
	}

	return info
}

// BadCrossbarInfo returns information about BadCrossbar availability.
// BadCrossbar is a Python tool for computing currents and voltages in passive crossbar arrays.
func BadCrossbarInfo() *ToolInfo {
	info := &ToolInfo{
		Name:        "BadCrossbar",
		Description: "UCL tool for passive crossbar current/voltage analysis",
		InstallCmd:  "pip install badcrossbar",
		DocURL:      "https://github.com/joksas/badcrossbar",
	}

	status, version, err := checkPythonModule("badcrossbar")
	info.Status = status
	info.Version = version
	if err != nil {
		info.Error = err.Error()
	}

	return info
}

// CheckAllTools returns the status of all supported crossbar simulation tools.
func CheckAllTools() []*ToolInfo {
	return []*ToolInfo{
		CrossSimInfo(),
		BadCrossbarInfo(),
	}
}

// IsCrossSimAvailable returns true if CrossSim is installed and available.
func IsCrossSimAvailable() bool {
	return CrossSimInfo().Status == StatusInstalled
}

// IsBadCrossbarAvailable returns true if BadCrossbar is installed and available.
func IsBadCrossbarAvailable() bool {
	return BadCrossbarInfo().Status == StatusInstalled
}

// checkPythonModule checks if a Python module is installed and returns its version.
func checkPythonModule(moduleName string) (ToolStatus, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First, check if Python is available
	pythonCmd := findPython()
	if pythonCmd == "" {
		return StatusError, "", fmt.Errorf("Python not found in PATH")
	}

	// Try to import the module and get its version
	script := fmt.Sprintf(`
import sys
try:
    import %s
    version = getattr(%s, '__version__', 'unknown')
    print(version)
except ImportError:
    sys.exit(1)
except Exception as e:
    print(str(e), file=sys.stderr)
    sys.exit(2)
`, moduleName, moduleName)

	cmd := exec.CommandContext(ctx, pythonCmd, "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return StatusError, "", fmt.Errorf("timeout checking module")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return StatusNotInstalled, "", nil
			}
			return StatusError, "", fmt.Errorf("import error: %s", stderr.String())
		}
		return StatusError, "", err
	}

	version := strings.TrimSpace(stdout.String())
	return StatusInstalled, version, nil
}

// findPython locates Python 3 executable in PATH.
func findPython() string {
	// Try python3 first, then python
	for _, name := range []string{"python3", "python"} {
		path, err := exec.LookPath(name)
		if err == nil {
			// Verify it's Python 3
			cmd := exec.Command(path, "--version")
			out, err := cmd.Output()
			if err == nil && strings.Contains(string(out), "Python 3") {
				return path
			}
		}
	}
	return ""
}

// ValidateCrossSim runs a basic CrossSim validation test.
// It returns success status, output message, and any error.
func ValidateCrossSim() (bool, string, error) {
	if !IsCrossSimAvailable() {
		return false, "", fmt.Errorf("CrossSim not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Basic CrossSim validation: import core module and check GPU availability
	script := `
import crosssim
try:
    # Check if CrossSim core is functional
    from crosssim.core import NeuralCore
    print("CrossSim core imported successfully")
    
    # Check for GPU support (optional)
    try:
        import torch
        gpu_available = torch.cuda.is_available()
        print(f"GPU available: {gpu_available}")
    except ImportError:
        print("GPU check skipped (PyTorch not installed)")
    
    print("VALIDATION_PASSED")
except Exception as e:
    print(f"ERROR: {e}")
`

	pythonCmd := findPython()
	cmd := exec.CommandContext(ctx, pythonCmd, "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		return false, output + stderr.String(), err
	}

	passed := strings.Contains(output, "VALIDATION_PASSED")
	return passed, output, nil
}

// ValidateBadCrossbar runs a basic BadCrossbar validation test.
// It returns success status, output message, and any error.
func ValidateBadCrossbar() (bool, string, error) {
	if !IsBadCrossbarAvailable() {
		return false, "", fmt.Errorf("BadCrossbar not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Basic BadCrossbar validation: import and run a simple computation
	script := `
import badcrossbar
import numpy as np

try:
    # Create a simple 2x2 crossbar for validation
    resistances = np.array([[100, 200], [300, 400]])  # Ohms
    applied_voltages = np.array([[1.0], [0.5]])  # Volts
    
    # Compute currents (the core functionality)
    from badcrossbar import compute
    solution = compute.compute(applied_voltages, resistances)
    
    # Check that we got valid output
    if solution is not None and hasattr(solution, 'currents'):
        print(f"Computed currents shape: {solution.currents.word_line.shape}")
        print("VALIDATION_PASSED")
    else:
        print("ERROR: Invalid solution format")
except Exception as e:
    print(f"ERROR: {e}")
`

	pythonCmd := findPython()
	cmd := exec.CommandContext(ctx, pythonCmd, "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		return false, output + stderr.String(), err
	}

	passed := strings.Contains(output, "VALIDATION_PASSED")
	return passed, output, nil
}

// ValidationResult represents the result of a tool validation.
type ValidationResult struct {
	Tool    string
	Passed  bool
	Output  string
	Error   error
	Elapsed time.Duration
}

// ValidateAllTools runs validation tests on all installed tools.
// Returns results for each tool, skipping tools that are not installed.
func ValidateAllTools() []*ValidationResult {
	var results []*ValidationResult

	// CrossSim
	start := time.Now()
	passed, output, err := ValidateCrossSim()
	results = append(results, &ValidationResult{
		Tool:    "CrossSim",
		Passed:  passed,
		Output:  output,
		Error:   err,
		Elapsed: time.Since(start),
	})

	// BadCrossbar
	start = time.Now()
	passed, output, err = ValidateBadCrossbar()
	results = append(results, &ValidationResult{
		Tool:    "BadCrossbar",
		Passed:  passed,
		Output:  output,
		Error:   err,
		Elapsed: time.Since(start),
	})

	return results
}

// GetProjectRoot returns the project root directory.
//
// Prefer finding the repo root by walking up the filesystem from a few likely
// anchors. This needs to work in CI and in `go test`, where runtime.Caller
// paths may point at temporary build directories.
func GetProjectRoot() (string, error) {
	findFrom := func(start string) (string, bool) {
		dir := start
		for i := 0; i < 10; i++ {
			// Prefer CLAUDE.md (project convention), but fall back to go.mod/.git.
			if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
				return dir, true
			}
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, true
			}
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
				return dir, true
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return "", false
	}

	if _, filename, _, ok := runtime.Caller(0); ok {
		if root, ok := findFrom(filepath.Dir(filename)); ok {
			return root, nil
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if root, ok := findFrom(wd); ok {
			return root, nil
		}
	}

	if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
		if root, ok := findFrom(ws); ok {
			return root, nil
		}
	}

	return "", fmt.Errorf("project root not found")
}

// GetLocalClonePaths returns paths to local clones of CrossSim and BadCrossbar.
// Returns crosssimPath, badcrossbarPath, error.
func GetLocalClonePaths() (string, string, error) {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return "", "", err
	}

	crosssimPath := filepath.Join(projectRoot, "opensource", "crossbar", "cross-sim")
	badcrossbarPath := filepath.Join(projectRoot, "opensource", "crossbar", "badcrossbar")

	return crosssimPath, badcrossbarPath, nil
}

// HasLocalClone checks if a local clone exists at the given path.
func HasLocalClone(path string) bool {
	// Check for setup.py or pyproject.toml as indicators of a Python package
	if _, err := os.Stat(filepath.Join(path, "setup.py")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(path, "pyproject.toml")); err == nil {
		return true
	}
	return false
}

// InstallResult represents the result of an installation attempt.
type InstallResult struct {
	Tool    string
	Success bool
	Output  string
	Error   error
}

// InstallCrossSim installs CrossSim from the local clone using pip install -e.
func InstallCrossSim() *InstallResult {
	crosssimPath, _, err := GetLocalClonePaths()
	if err != nil {
		return &InstallResult{
			Tool:    "CrossSim",
			Success: false,
			Error:   fmt.Errorf("could not find project root: %w", err),
		}
	}

	if !HasLocalClone(crosssimPath) {
		return &InstallResult{
			Tool:    "CrossSim",
			Success: false,
			Error:   fmt.Errorf("local clone not found at %s", crosssimPath),
		}
	}

	return installPythonPackage("CrossSim", crosssimPath)
}

// InstallBadCrossbar installs BadCrossbar from the local clone using pip install -e.
func InstallBadCrossbar() *InstallResult {
	_, badcrossbarPath, err := GetLocalClonePaths()
	if err != nil {
		return &InstallResult{
			Tool:    "BadCrossbar",
			Success: false,
			Error:   fmt.Errorf("could not find project root: %w", err),
		}
	}

	if !HasLocalClone(badcrossbarPath) {
		return &InstallResult{
			Tool:    "BadCrossbar",
			Success: false,
			Error:   fmt.Errorf("local clone not found at %s", badcrossbarPath),
		}
	}

	return installPythonPackage("BadCrossbar", badcrossbarPath)
}

// installPythonPackage runs pip install -e on a package directory.
func installPythonPackage(name, path string) *InstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pythonCmd := findPython()
	if pythonCmd == "" {
		return &InstallResult{
			Tool:    name,
			Success: false,
			Error:   fmt.Errorf("Python not found in PATH"),
		}
	}

	// Use pip install -e for editable install
	cmd := exec.CommandContext(ctx, pythonCmd, "-m", "pip", "install", "-e", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()

	if err != nil {
		return &InstallResult{
			Tool:    name,
			Success: false,
			Output:  output,
			Error:   err,
		}
	}

	return &InstallResult{
		Tool:    name,
		Success: true,
		Output:  output,
	}
}

// InstallToolsIfNeeded checks if tools are installed and installs them from local clones if not.
// Returns results for each tool (install attempt or already-installed status).
func InstallToolsIfNeeded() []*InstallResult {
	var results []*InstallResult

	// Check and install CrossSim
	if !IsCrossSimAvailable() {
		crosssimPath, _, _ := GetLocalClonePaths()
		if HasLocalClone(crosssimPath) {
			results = append(results, InstallCrossSim())
		} else {
			results = append(results, &InstallResult{
				Tool:    "CrossSim",
				Success: false,
				Error:   fmt.Errorf("not installed and local clone not found at opensource/crossbar/cross-sim"),
			})
		}
	} else {
		results = append(results, &InstallResult{
			Tool:    "CrossSim",
			Success: true,
			Output:  "Already installed",
		})
	}

	// Check and install BadCrossbar
	if !IsBadCrossbarAvailable() {
		_, badcrossbarPath, _ := GetLocalClonePaths()
		if HasLocalClone(badcrossbarPath) {
			results = append(results, InstallBadCrossbar())
		} else {
			results = append(results, &InstallResult{
				Tool:    "BadCrossbar",
				Success: false,
				Error:   fmt.Errorf("not installed and local clone not found at opensource/crossbar/badcrossbar"),
			})
		}
	} else {
		results = append(results, &InstallResult{
			Tool:    "BadCrossbar",
			Success: true,
			Output:  "Already installed",
		})
	}

	return results
}
