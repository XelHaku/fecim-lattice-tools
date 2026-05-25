//go:build legacy_fyne

package hysteresiscli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyFyneCommandHelpDeclaresDeprecation(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Run([]string{"--help"}); err != nil {
			t.Fatalf("Run(--help): %v", err)
		}
	})

	assertLegacyFyneDeprecationNotice(t, output)
	for _, stale := range []string{
		"recommended",
		"GPU accelerated",
	} {
		if strings.Contains(output, stale) {
			t.Fatalf("legacy Fyne help still markets deprecated UI with %q in output:\n%s", stale, output)
		}
	}
}

func TestLegacyFyneCommandListMaterialsDeclaresDeprecation(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Run([]string{"--list-materials"}); err != nil {
			t.Fatalf("Run(--list-materials): %v", err)
		}
	})

	assertLegacyFyneDeprecationNotice(t, output)
	if !strings.Contains(output, "Available materials") {
		t.Fatalf("list materials output missing material listing:\n%s", output)
	}
}

func TestLegacyFyneCommandDoesNotImportModuleGUI(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("main.go"))
	if err != nil {
		t.Fatalf("read legacy command source: %v", err)
	}
	if strings.Contains(string(body), "module1-hysteresis/pkg/gui") {
		t.Fatalf("legacy hysteresis-fyne command must be a deprecation shim, not import the Fyne GUI package")
	}
}

func TestLegacyFyneCommandDefaultInvocationFailsFastToGogpu(t *testing.T) {
	output := captureStdout(t, func() {
		err := Run(nil)
		if err == nil {
			t.Fatal("Run(nil) succeeded; default legacy Fyne invocation must fail fast to the gogpu/ui shell")
		}
		if !strings.Contains(err.Error(), "legacy Fyne GUI launch is fully deprecated") ||
			!strings.Contains(err.Error(), "CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module hysteresis") {
			t.Fatalf("default legacy error did not point to gogpu/ui migration path: %v", err)
		}
	})

	assertLegacyFyneDeprecationNotice(t, output)
	if strings.Contains(output, "Falling back") {
		t.Fatalf("legacy default invocation attempted graphical fallback instead of failing fast:\n%s", output)
	}
}

func assertLegacyFyneDeprecationNotice(t *testing.T, output string) {
	t.Helper()
	for _, want := range []string{
		"DEPRECATED",
		"legacy Fyne",
		"gogpu/ui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module hysteresis",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("legacy Fyne output missing %q in output:\n%s", want, output)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return buf.String()
}
