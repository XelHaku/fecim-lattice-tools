package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var researchRunner = runResearchTool

func runResearchSubcommand(args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return researchRunner(args)
}

func runResearchTool(args []string) error {
	python := os.Getenv("FECIM_RESEARCH_PYTHON")
	if python == "" {
		python = "python3"
	}
	root, err := researchRepoRoot(args)
	if err != nil {
		return err
	}
	script := filepath.Join(root, "tools", "research", "research_cli.py")
	cmdArgs := append([]string{script}, args...)
	cmd := exec.Command(python, cmdArgs...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("research tool: %w", err)
	}
	return nil
}

func researchRepoRoot(args []string) (string, error) {
	if root := repoRootFromResearchArgs(args); root != "" {
		return validateResearchRepoRoot(root)
	}
	if cwd, err := os.Getwd(); err == nil {
		if root, ok := findResearchRepoRoot(cwd); ok {
			return root, nil
		}
	}
	if exe, err := os.Executable(); err == nil {
		if root, ok := findResearchRepoRoot(filepath.Dir(exe)); ok {
			return root, nil
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		if root, ok := findResearchRepoRoot(filepath.Dir(file)); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("research tool: could not locate repository root containing tools/research/research_cli.py")
}

func repoRootFromResearchArgs(args []string) string {
	for i, arg := range args {
		if arg == "--repo-root" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--repo-root=") {
			return strings.TrimPrefix(arg, "--repo-root=")
		}
	}
	return ""
}

func validateResearchRepoRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("research tool: resolve repo root %q: %w", root, err)
	}
	abs = filepath.Clean(abs)
	if researchScriptExists(abs) {
		return abs, nil
	}
	return "", fmt.Errorf("research tool: could not find tools/research/research_cli.py under %s", abs)
}

func findResearchRepoRoot(start string) (string, bool) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	current = filepath.Clean(current)
	for {
		if researchScriptExists(current) {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func researchScriptExists(root string) bool {
	info, err := os.Stat(filepath.Join(root, "tools", "research", "research_cli.py"))
	return err == nil && !info.IsDir()
}
