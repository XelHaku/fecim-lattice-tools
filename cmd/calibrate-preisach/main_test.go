package main

import (
	"os"
	"path/filepath"
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
