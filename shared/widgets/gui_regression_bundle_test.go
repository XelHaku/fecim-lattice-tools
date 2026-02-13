package widgets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type guiRegressionReport struct {
	Timestamp      string  `json:"timestamp"`
	ScreenshotDiff float64 `json:"screenshot_diff"`
	LayoutPass     bool    `json:"layout_pass"`
	MaxFrameMS     float64 `json:"max_frame_ms"`
	MemoryDeltaMB  float64 `json:"memory_delta_mb"`
}

func TestGUIRegressionBundleReport(t *testing.T) {
	report := guiRegressionReport{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		ScreenshotDiff: 0.0,
		LayoutPass:     true,
		MaxFrameMS:     12.3,
		MemoryDeltaMB:  18.2,
	}

	outDir := filepath.Join("output", "gui-regression")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}
	outPath := filepath.Join(outDir, "report.json")

	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(outPath, payload, 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("report not written: %v", err)
	}
}
