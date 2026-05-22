package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBaseMaterial(t *testing.T) {
	if baseMaterial("default_hzo") == nil {
		t.Fatal("default_hzo should resolve")
	}
	if baseMaterial("literature_superlattice") == nil {
		t.Fatal("literature_superlattice should resolve")
	}
	if baseMaterial("unknown") != nil {
		t.Fatal("unknown preset should return nil")
	}
}

func TestLoadCSVLoop(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "loop.csv")
	content := "E_MV_cm,P_uC_cm2\n-1.0,-10.0\n0.0,0.0\n1.0,10.0\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	E, P, err := loadCSVLoop(p)
	if err != nil {
		t.Fatalf("loadCSVLoop error: %v", err)
	}
	if len(E) != 3 || len(P) != 3 {
		t.Fatalf("unexpected lengths E=%d P=%d", len(E), len(P))
	}
	if E[0] != -1.0 || P[2] != 10.0 {
		t.Fatalf("unexpected parsed values E0=%v P2=%v", E[0], P[2])
	}
}

func TestLoadCSVLoopRejectsMalformedRows(t *testing.T) {
	t.Run("missing column", func(t *testing.T) {
		d := t.TempDir()
		p := filepath.Join(d, "loop.csv")
		content := "E_MV_cm,P_uC_cm2\n-1.0\n"
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		if _, _, err := loadCSVLoop(p); err == nil || !strings.Contains(err.Error(), "expected at least 2 columns") {
			t.Fatalf("expected missing-column error, got %v", err)
		}
	})

	t.Run("bad float", func(t *testing.T) {
		d := t.TempDir()
		p := filepath.Join(d, "loop.csv")
		content := "E_MV_cm,P_uC_cm2\nabc,1.0\n"
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		if _, _, err := loadCSVLoop(p); err == nil || !strings.Contains(err.Error(), "parse E_MV_cm") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestRunCalibratePreisachUnknownPresetWithoutExiting(t *testing.T) {
	d := t.TempDir()
	csvPath := filepath.Join(d, "loop.csv")
	content := "E_MV_cm,P_uC_cm2\n-1.0,-10.0\n0.0,0.0\n1.0,10.0\n"
	if err := os.WriteFile(csvPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCalibratePreisach([]string{"-csv", csvPath, "-preset", "unknown"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code=%d, want 1; stderr=%q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q, want empty output", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unknown preset \"unknown\"") {
		t.Fatalf("stderr=%q, want unknown preset context", stderr.String())
	}
}
