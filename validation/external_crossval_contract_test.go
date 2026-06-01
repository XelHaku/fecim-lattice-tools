package validation

import "testing"

func TestExternalMVMCrossvalArtifacts_Contract(t *testing.T) {
	paths := externalCrossvalArtifactPaths(4, 8, 16)
	seen := map[int]bool{}
	for _, p := range paths {
		rec := readExternalCrossvalArtifact(t, p)
		if rec.MaxVErr > 1e-12 {
			t.Fatalf("%s maxVErr_V=%g exceeds contract 1e-12", p, rec.MaxVErr)
		}
		seen[rec.N] = true
	}
	for _, n := range []int{4, 8, 16} {
		if !seen[n] {
			t.Fatalf("missing crossval artifact for n=%d", n)
		}
	}
}
