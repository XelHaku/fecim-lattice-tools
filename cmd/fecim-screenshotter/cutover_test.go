package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultScreenshotterDoesNotDependOnFyne(t *testing.T) {
	needle := "fyne.io/" + "fyne"
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(body), needle) {
			t.Fatalf("default screenshotter must not import Fyne; found import in %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "list", "-deps", ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}
	for _, dep := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(dep, needle) {
			t.Fatalf("default screenshotter must not depend on Fyne; found transitive dependency %s", dep)
		}
	}
}
