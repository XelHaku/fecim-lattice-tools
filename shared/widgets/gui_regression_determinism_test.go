//go:build legacy_fyne

package widgets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGUIRegressionBundleReport_DeterministicTimestamp(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	t.Setenv("FECIM_ARTIFACT_TIMESTAMP_UTC", "")
	TestGUIRegressionBundleReport(t)

	p := filepath.Join("output", "gui-regression", "report.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var got guiRegressionReport
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if got.Timestamp != "1970-01-01T00:00:00Z" {
		t.Fatalf("expected deterministic timestamp, got %q", got.Timestamp)
	}
}
