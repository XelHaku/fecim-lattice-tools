package physics

import (
	"math/rand"
	"testing"
)

func TestC2CSigma_ScalesInversely(t *testing.T) {
	G_max := 1e-5 // 10 µS
	cfg := DefaultStateDepC2CConfig(G_max)

	sigmaHigh, err := cfg.C2CSigma(G_max)
	if err != nil {
		t.Fatal(err)
	}
	// Use G_max/2: natural sigma = 0.1*G_max, clamp threshold = 0.5*(G_max/2) = 0.25*G_max → no clamp
	sigmaLow, err := cfg.C2CSigma(G_max / 2)
	if err != nil {
		t.Fatal(err)
	}
	// At G_max/2 natural sigma (0.1*G_max) > at G_max (0.05*G_max), no clamping involved
	if sigmaLow <= sigmaHigh {
		t.Errorf("lower G should have higher absolute sigma: low=%g high=%g", sigmaLow, sigmaHigh)
	}
}

func TestC2CRelativeNoise_1OverG(t *testing.T) {
	G_max := 1e-5
	cfg := DefaultStateDepC2CConfig(G_max)

	rel1, _ := C2CRelativeNoise(G_max, cfg)
	rel2, _ := C2CRelativeNoise(G_max/2, cfg)
	// rel2 should be ~2x rel1 (σ/G ∝ 1/G)
	if rel2 < 1.5*rel1 {
		t.Errorf("relative noise should scale ~1/G: rel1=%g rel2=%g", rel1, rel2)
	}
}

func TestApplyStateDepC2C_NonNegative(t *testing.T) {
	G_max := 1e-5
	cfg := DefaultStateDepC2CConfig(G_max)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		noisy, err := ApplyStateDepC2C(G_max, cfg, rng)
		if err != nil {
			t.Fatal(err)
		}
		if noisy < 0 {
			t.Errorf("noisy conductance must be non-negative, got %g", noisy)
		}
	}
}

func TestSweepC2CRelativeNoise(t *testing.T) {
	G_max := 1e-5
	cfg := DefaultStateDepC2CConfig(G_max)
	// Values stay above the 0.5*G clamp threshold (G > sqrt(AbsoluteNoiseSigma*G_ref/0.5) ≈ 0.316*G_max).
	// G_max/5 and G_max/10 both saturate at rel=0.5 and would make the monotonicity check spuriously fail.
	Gs := []float64{G_max, G_max / 2, G_max / 2.5, G_max / 3}
	rels, err := SweepC2CRelativeNoise(Gs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != len(Gs) {
		t.Fatalf("expected %d results, got %d", len(Gs), len(rels))
	}
	// Relative noise should increase as G decreases (1/G relationship), all in non-clamped region.
	for i := 1; i < len(rels); i++ {
		if rels[i] <= rels[i-1] {
			t.Errorf("relative noise should increase as G decreases: rels[%d]=%g rels[%d]=%g", i, rels[i], i-1, rels[i-1])
		}
	}
}
