package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckArchitectureIgnoresLocalWorktreesGoMod(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.test/fecim\n\ngo 1.25\n")
	mustWrite(t, filepath.Join(root, ".worktrees", "feature", "go.mod"), "module example.test/worktree\n\ngo 1.25\n")
	mustWrite(t, filepath.Join(root, "scripts", "check-architecture.sh"), readFile(t, filepath.Join(repoRoot(t), "scripts", "check-architecture.sh")))
	runGit(t, root, "init")

	cmd := exec.Command("bash", "scripts/check-architecture.sh", "--fast")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("architecture check failed with local .worktrees go.mod:\n%s", out)
	}
	if strings.Contains(string(out), ".worktrees/feature/go.mod") {
		t.Fatalf("architecture check reported local worktree go.mod:\n%s", out)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}
