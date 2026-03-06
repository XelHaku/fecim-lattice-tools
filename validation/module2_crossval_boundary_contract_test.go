package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type module2CrossvalRecord struct {
	N       int     `json:"n"`
	RWL     float64 `json:"RWL_ohm"`
	RBL     float64 `json:"RBL_ohm"`
	MaxIErr float64 `json:"maxIErr_A"`
	MaxVErr float64 `json:"maxVErr_V"`
	PassI   bool    `json:"pass_I"`
	PassV   bool    `json:"pass_V"`
}

func readModule2Crossval(t *testing.T, p string) module2CrossvalRecord {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec module2CrossvalRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if rec.N <= 0 || rec.RWL <= 0 || rec.RBL <= 0 {
		t.Fatalf("%s invalid n/resistances: n=%d rwl=%g rbl=%g", p, rec.N, rec.RWL, rec.RBL)
	}
	if !rec.PassI || !rec.PassV {
		t.Fatalf("%s pass flags false: pass_I=%v pass_V=%v", p, rec.PassI, rec.PassV)
	}
	if rec.MaxIErr < 0 || rec.MaxVErr < 0 {
		t.Fatalf("%s negative errors: maxIErr=%g maxVErr=%g", p, rec.MaxIErr, rec.MaxVErr)
	}
	return rec
}

func TestModule2CrossvalBoundaryInvariant_NScaling(t *testing.T) {
	repoRoot := filepath.Clean("..")
	base := filepath.Join(repoRoot, "validation", "output", "validation", "external")

	r4 := readModule2Crossval(t, filepath.Join(base, "mvm_numpy_crossval_4x4.json"))
	r8 := readModule2Crossval(t, filepath.Join(base, "mvm_numpy_crossval_8x8.json"))
	r16 := readModule2Crossval(t, filepath.Join(base, "mvm_numpy_crossval_16x16.json"))

	if r4.N != 4 || r8.N != 8 || r16.N != 16 {
		t.Fatalf("unexpected n sequence: got [%d,%d,%d]", r4.N, r8.N, r16.N)
	}
	if r4.RWL != r8.RWL || r8.RWL != r16.RWL || r4.RBL != r8.RBL || r8.RBL != r16.RBL {
		t.Fatalf("RWL/RBL mismatch across boundary artifacts")
	}
	// Error should not improve when scaling from 4->8->16 for this fixed solver setup;
	// enforce monotonic non-decreasing envelope to catch accidental cross-file drift.
	if !(r4.MaxVErr <= r8.MaxVErr && r8.MaxVErr <= r16.MaxVErr) {
		t.Fatalf("maxVErr non-monotonic across n: 4x4=%g 8x8=%g 16x16=%g", r4.MaxVErr, r8.MaxVErr, r16.MaxVErr)
	}
	if !(r4.MaxIErr <= r8.MaxIErr && r8.MaxIErr <= r16.MaxIErr) {
		t.Fatalf("maxIErr non-monotonic across n: 4x4=%g 8x8=%g 16x16=%g", r4.MaxIErr, r8.MaxIErr, r16.MaxIErr)
	}
}
