package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type externalCrossvalArtifact struct {
	N       int     `json:"n"`
	RWL     float64 `json:"RWL_ohm"`
	RBL     float64 `json:"RBL_ohm"`
	MaxIErr float64 `json:"maxIErr_A"`
	MaxVErr float64 `json:"maxVErr_V"`
	PassI   bool    `json:"pass_I"`
	PassV   bool    `json:"pass_V"`
}

func externalCrossvalArtifactPath(n int) string {
	return releaseArtifactPath(
		"validation",
		"output",
		"validation",
		"external",
		fmt.Sprintf("mvm_numpy_crossval_%dx%d.json", n, n),
	)
}

func externalCrossvalArtifactPaths(sizes ...int) []string {
	paths := make([]string, 0, len(sizes))
	for _, n := range sizes {
		paths = append(paths, externalCrossvalArtifactPath(n))
	}
	return paths
}

func readExternalCrossvalArtifact(t *testing.T, p string) externalCrossvalArtifact {
	t.Helper()
	b, err := os.ReadFile(filepath.Clean(p))
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec externalCrossvalArtifact
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
