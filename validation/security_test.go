package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	hysteresisrender "fecim-lattice-tools/module1-hysteresis/pkg/render"
	"fecim-lattice-tools/shared/cli"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/presets"
	"fecim-lattice-tools/shared/recording"
)

func TestSecurity_FileOperationsRejectPathTraversal(t *testing.T) {
	base := t.TempDir()
	mgr := presets.NewManager(filepath.Join(base, "presets"))

	p := presets.NewPreset("unsafe", "path traversal attempt", presets.ModuleGlobal, presets.CategoryCustom, map[string]interface{}{"k": "v"})
	p.Metadata.ID = "../../outside"

	if err := mgr.Save(p); err == nil {
		t.Fatal("expected Save to reject unsafe preset ID")
	}

	outsidePath := filepath.Join(base, "outside.json")
	if _, err := os.Stat(outsidePath); err == nil {
		t.Fatalf("unexpected file created outside presets directory: %s", outsidePath)
	}
}

func TestSecurity_ConfigLoadingRejectsOversizeFiles(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "too-large.yaml")
	if err := os.WriteFile(p, []byte(strings.Repeat("a", 11*1024*1024)), 0o644); err != nil {
		t.Fatalf("write oversized config: %v", err)
	}

	loader := cli.NewConfigLoader(p)
	var cfg map[string]interface{}
	if err := loader.Load(&cfg); err == nil {
		t.Fatal("expected oversized config file to be rejected")
	}
}

func TestSecurity_UserInputsValidatedBeforeCalculations(t *testing.T) {
	badRenderConfig := &hysteresisrender.Config{Width: -1, Height: 720, TargetFPS: 60}
	if err := badRenderConfig.Validate(); err == nil {
		t.Fatal("expected renderer config validation to fail for invalid dimensions")
	}

	badRecordingSettings := recording.DefaultSettings()
	badRecordingSettings.FPS = 0
	if err := badRecordingSettings.Validate(); err == nil {
		t.Fatal("expected recording settings validation to fail for invalid FPS")
	}
}

func TestSecurity_LogsDoNotExposeSensitiveData(t *testing.T) {
	logging.ClearBuffer()
	logger := logging.NewLogger("security-test")
	defer logger.Close()

	oldVerbosity := logging.GetVerbosity()
	logging.SetVerbosity(logging.VerbosityDebug)
	defer logging.SetVerbosity(oldVerbosity)

	params := map[string]interface{}{
		"username": "alice",
		"password": "super-secret",
		"apiToken": "tok_123",
	}
	logger.Input("authenticate", params)

	entries := logging.ReadBuffer(1)
	if len(entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(entries))
	}
	msg := entries[0].Message
	if strings.Contains(msg, "super-secret") || strings.Contains(msg, "tok_123") {
		t.Fatalf("sensitive values leaked to log message: %s", msg)
	}
	if entries[0].Fields["password"] != "[REDACTED]" || entries[0].Fields["apiToken"] != "[REDACTED]" {
		t.Fatalf("expected redacted sensitive fields, got: %#v", entries[0].Fields)
	}
}

func TestSecurity_BufferOverflowProtectionsForArrayOperations(t *testing.T) {
	bp := recording.NewBufferPool(int(^uint(0)>>1), 2)
	if size := bp.Stats().BufferSize; size != 1 {
		t.Fatalf("expected overflow-safe fallback buffer size 1, got %d", size)
	}

	fb := recording.NewFrameBuffer(2, 2)
	before := append([]byte(nil), fb.Data()...)
	fb.SetPixel(10, 10, 255, 255, 255) // out of bounds must be ignored
	if got := fb.Data(); !equalBytes(before, got) {
		t.Fatal("out-of-bounds SetPixel modified framebuffer data")
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
