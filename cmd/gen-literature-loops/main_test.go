package main

import "testing"

func TestMaterialForPreset(t *testing.T) {
	mat, emax, err := materialForPreset("park")
	if err != nil || mat == nil || emax != 3.0 {
		t.Fatalf("park preset invalid: mat=%v emax=%v err=%v", mat, emax, err)
	}
	if _, _, err := materialForPreset("bad"); err == nil {
		t.Fatal("expected error for unknown preset")
	}
}

func TestBuildSweep(t *testing.T) {
	E := buildSweep(3.0)
	if len(E) != 61 {
		t.Fatalf("len(E)=%d, want 61", len(E))
	}
	if E[0] != -3.0 || E[30] != 3.0 || E[len(E)-1] != -3.0 {
		t.Fatalf("unexpected endpoints: first=%v mid=%v last=%v", E[0], E[30], E[len(E)-1])
	}
}
