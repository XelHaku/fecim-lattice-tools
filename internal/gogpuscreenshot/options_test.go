package gogpuscreenshot

import (
	"path/filepath"
	"testing"
)

func TestParseOptionsSupportsLegacyScreenshotFlags(t *testing.T) {
	opts, err := ParseOptions([]string{
		"-out", "docs/assets",
		"-only", "hysteresis",
		"-tag", "readme",
		"-w", "1280",
		"-h", "820",
	})
	if err != nil {
		t.Fatalf("ParseOptions error: %v", err)
	}

	if opts.OutputDir != "docs/assets" {
		t.Fatalf("OutputDir = %q, want docs/assets", opts.OutputDir)
	}
	if opts.Only != "hysteresis" {
		t.Fatalf("Only = %q, want hysteresis", opts.Only)
	}
	if opts.Tag != "readme" {
		t.Fatalf("Tag = %q, want readme", opts.Tag)
	}
	if opts.Width != 1280 || opts.Height != 820 {
		t.Fatalf("size = %dx%d, want 1280x820", opts.Width, opts.Height)
	}

	got := opts.OutputPath("hysteresis-p-e-loop.png")
	want := filepath.Join("docs/assets", "hysteresis-p-e-loop_readme.png")
	if got != want {
		t.Fatalf("OutputPath = %q, want %q", got, want)
	}
}

func TestDefaultOptionsMatchGogpuScreenshotDefaults(t *testing.T) {
	opts := DefaultOptions()

	if opts.OutputDir != "screenshots" {
		t.Fatalf("OutputDir = %q, want screenshots", opts.OutputDir)
	}
	if opts.Width != 1400 || opts.Height != 900 {
		t.Fatalf("size = %dx%d, want 1400x900", opts.Width, opts.Height)
	}
	if opts.Only != "" || opts.Tag != "" {
		t.Fatalf("unexpected filters: only=%q tag=%q", opts.Only, opts.Tag)
	}
}
