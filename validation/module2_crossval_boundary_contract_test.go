package validation

import "testing"

func TestModule2CrossvalBoundaryInvariant_NScaling(t *testing.T) {
	r4 := readExternalCrossvalArtifact(t, externalCrossvalArtifactPath(4))
	r8 := readExternalCrossvalArtifact(t, externalCrossvalArtifactPath(8))
	r16 := readExternalCrossvalArtifact(t, externalCrossvalArtifactPath(16))

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
