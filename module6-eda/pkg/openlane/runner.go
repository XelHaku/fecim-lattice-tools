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

// runDockerOpenROAD runs OpenROAD in Docker container with correct --entrypoint pattern
func (r *Runner) runDockerOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	pdkRoot := r.manager.GetPDKRoot()
	if pdkRoot == "" {
		return nil, fmt.Errorf("PDK_ROOT not set - run GetPDKSetupInstructions() for setup help")
	}

	// Build Docker command with --entrypoint openroad
	args := []string{
		"run", "--rm",
		"--entrypoint", "openroad",
		"-v", fmt.Sprintf("%s:/design", workDir),
		"-w", "/design",
		"-v", fmt.Sprintf("%s:/pdk:ro", pdkRoot),
	}

	// Add environment variables
	for k, v := range envVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add default PDK paths if not specified
	if _, ok := envVars["TECH_LEF"]; !ok {
		args = append(args, "-e", "TECH_LEF=/pdk/sky130A/libs.tech/openlane/sky130_fd_sc_hd/sky130_fd_sc_hd.tlef")
	}
	if _, ok := envVars["CELL_LEF"]; !ok {
		args = append(args, "-e", "CELL_LEF=/pdk/sky130A/libs.ref/sky130_fd_sc_hd/lef/sky130_fd_sc_hd.lef")
	}

	// Add image and OpenROAD flags
	args = append(args, r.manager.GetDockerImage())
	args = append(args, "-no_splash", "-exit", fmt.Sprintf("/design/%s", scriptName))

	return r.runWithTimeout("docker", args, workDir, r.config.TimeoutPlacement)
}

// runNativeOpenROAD runs OpenROAD directly
func (r *Runner) runNativeOpenROAD(scriptName string, workDir string, envVars map[string]string) (*Result, error) {
	pdkRoot := r.manager.GetPDKRoot()
	if pdkRoot == "" {
		return nil, fmt.Errorf("PDK_ROOT not set - run GetPDKSetupInstructions() for setup help")
	}

	scriptPath := filepath.Join(workDir, scriptName)
	args := []string{"-no_splash", "-exit", scriptPath}

	// Set up environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("PDK_ROOT=%s", pdkRoot))

	// Add default PDK paths if not specified
	techLEF := filepath.Join(pdkRoot, "sky130A/libs.tech/openlane/sky130_fd_sc_hd/sky130_fd_sc_hd.tlef")
	cellLEF := filepath.Join(pdkRoot, "sky130A/libs.ref/sky130_fd_sc_hd/lef/sky130_fd_sc_hd.lef")

	if v, ok := envVars["TECH_LEF"]; ok {
		techLEF = v
	}
	if v, ok := envVars["CELL_LEF"]; ok {
		cellLEF = v
	}

	env = append(env, fmt.Sprintf("TECH_LEF=%s", techLEF))
	env = append(env, fmt.Sprintf("CELL_LEF=%s", cellLEF))

	// Add user-provided env vars
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return r.runWithTimeoutEnv("openroad", args, workDir, r.config.TimeoutPlacement, env)
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
