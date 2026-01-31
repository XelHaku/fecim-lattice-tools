package openlane

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Result contains the output of a tool execution
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Runner executes OpenLane tools in Docker or native mode
type Runner struct {
	manager *Manager
	config  *Config
}

// NewRunner creates a new runner
func NewRunner(manager *Manager, config *Config) *Runner {
	return &Runner{
		manager: manager,
		config:  config,
	}
}

// RunOpenROAD executes an OpenROAD TCL script
// workDir should contain the script file and design files
// scriptName is the name of the TCL script file in workDir
func (r *Runner) RunOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	mode := r.manager.DetectMode()

	switch mode {
	case ModeDocker:
		return r.runDockerOpenROAD(scriptName, workDir, envVars)
	case ModeNative:
		return r.runNativeOpenROAD(scriptName, workDir, envVars)
	default:
		return nil, fmt.Errorf("no OpenROAD execution mode available (install Docker or OpenROAD)")
	}
}

// runDockerOpenROAD runs OpenROAD in Docker container with Xvfb for headless image export
// Uses Xvfb (virtual framebuffer) to enable save_image without X11 forwarding
func (r *Runner) runDockerOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	// Docker requires absolute paths for volume mounts
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Build Docker command with --entrypoint sh to run Xvfb wrapper
	args := []string{
		"run", "--rm",
		"--entrypoint", "sh",
		"-v", fmt.Sprintf("%s:/design", absWorkDir),
		"-w", "/design",
	}

	// Add environment variables (caller provides CELL_LEF, DEF_FILE, etc.)
	for k, v := range envVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add image
	args = append(args, r.manager.GetDockerImage())
	args = append(args, "-c")

	// Build OpenROAD command with Xvfb wrapper for headless operation
	xvfbCmd := fmt.Sprintf("Xvfb :99 -screen 0 1024x768x24 -nolisten tcp > /dev/null 2>&1 & sleep 1 && export DISPLAY=:99 && openroad -no_splash -exit /design/%s", scriptName)
	args = append(args, xvfbCmd)

	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutPlacement)
}

// runNativeOpenROAD runs OpenROAD directly
func (r *Runner) runNativeOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	scriptPath := filepath.Join(workDir, scriptName)
	args := []string{"-no_splash", "-exit", scriptPath}

	// Set up environment from caller-provided vars
	// For FeCIM validation, caller provides CELL_LEF path - no external PDK required
	env := os.Environ()

	// Add user-provided env vars (CELL_LEF, DEF_FILE, etc.)
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return r.runWithTimeoutEnv("openroad", args, workDir, r.config.TimeoutPlacement, env)
}

// RunYosys executes a Yosys command
// workDir should contain the Verilog files
// yosysCmd is the yosys command string (e.g., "read_verilog file.v; hierarchy -check")
func (r *Runner) RunYosys(yosysCmd string, workDir string) (*Result, error) {
	mode := r.manager.DetectMode()

	switch mode {
	case ModeDocker:
		return r.runDockerYosys(yosysCmd, workDir)
	case ModeNative:
		return r.runNativeYosys(yosysCmd, workDir)
	default:
		return nil, fmt.Errorf("no Yosys execution mode available (install Docker with OpenLane image or native yosys)")
	}
}

// runDockerYosys runs Yosys in Docker container
func (r *Runner) runDockerYosys(yosysCmd string, workDir string) (*Result, error) {
	// Docker requires absolute paths for volume mounts
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Build Docker command with --entrypoint yosys
	args := []string{
		"run", "--rm",
		"--entrypoint", "yosys",
		"-v", fmt.Sprintf("%s:/design", absWorkDir),
		"-w", "/design",
	}

	// Add image and Yosys command
	args = append(args, r.manager.GetDockerImage())
	args = append(args, "-p", yosysCmd)

	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutSynthesis)
}

// runNativeYosys runs Yosys directly
func (r *Runner) runNativeYosys(yosysCmd string, workDir string) (*Result, error) {
	args := []string{"-p", yosysCmd}
	return r.runWithTimeout("yosys", args, workDir, r.config.TimeoutSynthesis)
}

// RunKLayout executes a KLayout script for layout visualization
// workDir should contain the DEF and LEF files
// scriptPath is the path to the Python/Ruby script
func (r *Runner) RunKLayout(scriptPath string, workDir string, envVars map[string]string) (*Result, error) {
	mode := r.manager.DetectMode()

	switch mode {
	case ModeDocker:
		return r.runDockerKLayout(scriptPath, workDir, envVars)
	case ModeNative:
		return r.runNativeKLayout(scriptPath, workDir, envVars)
	default:
		return nil, fmt.Errorf("no KLayout execution mode available (install Docker with OpenLane image or native klayout)")
	}
}

// runDockerKLayout runs KLayout in Docker container with Xvfb for headless image export
// Uses -rd flags to pass variables to scripts (standard KLayout pattern)
// Uses Xvfb (virtual framebuffer) for headless GUI operations
func (r *Runner) runDockerKLayout(scriptPath string, workDir string, envVars map[string]string) (*Result, error) {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Build Docker command with --entrypoint sh to run Xvfb wrapper
	args := []string{
		"run", "--rm",
		"--entrypoint", "sh",
		"-v", fmt.Sprintf("%s:/design", absWorkDir),
		"-w", "/design",
		r.manager.GetDockerImage(),
		"-c",
	}

	// Build KLayout command with Xvfb wrapper for headless operation
	// Xvfb :99 creates virtual display, sleep ensures it's ready, then run klayout
	klayoutArgs := "-z"
	for k, v := range envVars {
		klayoutArgs += fmt.Sprintf(" -rd %s=%s", k, v)
	}
	klayoutArgs += fmt.Sprintf(" -r /design/%s", filepath.Base(scriptPath))

	xvfbCmd := fmt.Sprintf("Xvfb :99 -screen 0 1024x768x24 -nolisten tcp > /dev/null 2>&1 & sleep 1 && export DISPLAY=:99 && klayout %s", klayoutArgs)
	args = append(args, xvfbCmd)

	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutPlacement)
}

// runNativeKLayout runs KLayout directly
// Uses -rd flags to pass variables to scripts (standard KLayout pattern)
func (r *Runner) runNativeKLayout(scriptPath string, workDir string, envVars map[string]string) (*Result, error) {
	// KLayout flags: -z for batch mode with main window (required for image export)
	args := []string{"-z"}

	// Add variables using -rd (standard KLayout pattern per docs)
	for k, v := range envVars {
		args = append(args, "-rd", fmt.Sprintf("%s=%s", k, v))
	}

	// Add script
	args = append(args, "-r", scriptPath)

	return r.runWithTimeout("klayout", args, workDir, r.config.TimeoutPlacement)
}

// runWithTimeout executes a command with timeout
func (r *Runner) runWithTimeout(command string, args []string, workDir string, timeout time.Duration) (*Result, error) {
	return r.runWithTimeoutEnv(command, args, workDir, timeout, nil)
}

// runWithTimeoutEnv executes a command with timeout and custom environment
func (r *Runner) runWithTimeoutEnv(command string, args []string, workDir string, timeout time.Duration, env []string) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	if env != nil {
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %v", timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}
