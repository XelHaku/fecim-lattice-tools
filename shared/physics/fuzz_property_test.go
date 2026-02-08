package physics

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"testing"
)

func isFinite(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}

func requireFinite(t *testing.T, name string, v float64) {
	t.Helper()
	if !isFinite(v) {
		t.Fatalf("%s is not finite: %v", name, v)
	}
}

func randRange(r *rand.Rand, lo, hi float64) float64 {
	return lo + (hi-lo)*r.Float64()
}

// TestFuzz_LKStep_NoNaNsAndBounds is a lightweight, bounded-runtime property test.
// It randomizes solver parameters and repeatedly calls Step() looking for:
//   - NaNs/Infs in state/time
//   - polarization escaping [-PMax, +PMax]
//
// It is deterministic by default (seeded), and can be expanded by setting:
//   FECIM_FUZZ_ITERS, FECIM_FUZZ_STEPS.
func TestFuzz_LKStep_NoNaNsAndBounds(t *testing.T) {
	iters := 50
	steps := 200
	if v := getenvInt("FECIM_FUZZ_ITERS", 0); v > 0 {
		iters = v
	}
	if v := getenvInt("FECIM_FUZZ_STEPS", 0); v > 0 {
		steps = v
	}

	baseSeed := int64(1337)
	for i := 0; i < iters; i++ {
		seed := baseSeed + int64(i)*1009
		r := rand.New(rand.NewSource(seed))

		s := NewLKSolver()
		// Keep ranges broad enough to shake out numerical issues but not so broad
		// that we only test obviously unphysical/overflow regimes.
		s.Beta = -randRange(r, 1e6, 5e9)     // negative
		s.Gamma = randRange(r, 1e8, 5e11)    // positive
		s.Rho = randRange(r, 1e-4, 1.0)      // >0
		s.Q12 = randRange(r, -0.05, 0.0)     // typical negative
		s.Stress = randRange(r, 0, 2.0e9)    // 0..2 GPa
		s.K_dep = randRange(r, 0, 5e8)       // allow 0
		s.SeriesResistance = randRange(r, 0, 200)
		s.Thickness = randRange(r, 5e-9, 30e-9)
		s.Area = randRange(r, 20e-9*20e-9, 80e-9*80e-9)
		s.CurieTemp = randRange(r, 500, 900)
		s.CurieConst = randRange(r, 5e4, 4e5)

		s.EnableNoise = r.Intn(10) == 0 // occasionally enable noise
		s.UseNLS = r.Intn(2) == 0
		// Keep NLS parameters in a stable numeric range.
		s.ActivationField = randRange(r, 1e8, 5e9)
		s.TauInf = randRange(r, 1e-15, 1e-10)
		// Avoid incubation getting stuck forever when enabled.
		s.IncubationEnd = randRange(r, 0, 5e-9)

		pMax := randRange(r, 0.05, 0.6)
		s.PMax = pMax
		// Start anywhere within bounds.
		s.P = randRange(r, -pMax, pMax)

		s.Temperature = randRange(r, 250, 450)
		// Mix dt and field variations.
		for step := 0; step < steps; step++ {
			E := randRange(r, -4e9, 4e9)
			dt := randRange(r, 1e-15, 2e-9)

			beforeP := s.P
			p := s.Step(E, dt)

			requireFinite(t, "P", p)
			requireFinite(t, "state.P", s.P)
			requireFinite(t, "Time", s.Time)

			if s.PMax > 0 {
				limit := 1.2*s.PMax + 1e-12 // must match clampP() guard-band
				if s.P < -limit || s.P > limit {
					t.Fatalf("seed=%d step=%d: P out of bounds: P=%g limit=%g PMax=%g (before=%g E=%g dt=%g)", seed, step, s.P, limit, s.PMax, beforeP, E, dt)
				}
			}
		}
	}
}

// simpleEverett implements a bounded, monotone Everett integral for testing.
// It produces E(sat, -sat) == 1, so polarization is typically in [-1, 1].
// This is not a physical distribution; it is a numerical invariant checker.
type simpleEverett struct{ sat float64 }

func (e simpleEverett) Calculate(alpha, beta float64) float64 {
	// In a Preisach plane, the meaningful region is alpha >= beta.
	// For robustness, tolerate any ordering.
	d := alpha - beta
	return d / (2 * e.sat)
}

func TestFuzz_PreisachUpdate_NoNaNsAndStackSanity(t *testing.T) {
	iters := 50
	steps := 300
	if v := getenvInt("FECIM_FUZZ_ITERS", 0); v > 0 {
		iters = v
	}
	if v := getenvInt("FECIM_FUZZ_STEPS", 0); v > 0 {
		steps = v
	}

	baseSeed := int64(20240207)
	for i := 0; i < iters; i++ {
		seed := baseSeed + int64(i)*9176
		r := rand.New(rand.NewSource(seed))

		sat := randRange(r, 1e8, 5e9)
		ps := NewPreisachStack(sat, simpleEverett{sat: sat})

		// Start from initialization; then drive with random walk + reversals.
		E := -sat
		for step := 0; step < steps; step++ {
			// Randomly do small moves or big excursions (including beyond saturation).
			jump := randRange(r, -0.4*sat, 0.4*sat)
			if r.Intn(20) == 0 {
				jump = randRange(r, -2.0*sat, 2.0*sat)
			}
			E += jump
			p := ps.Update(E)

			requireFinite(t, "P", p)
			requireFinite(t, "LastE", ps.LastE)

			// Stack size should not explode beyond O(steps).
			if len(ps.Stack) > steps+4 {
				t.Fatalf("seed=%d step=%d: stack too large: %d", seed, step, len(ps.Stack))
			}

			// Basic turning-point sanity: types should alternate and be +/-1.
			for j, tp := range ps.Stack {
				if tp.Type != -1 && tp.Type != 1 {
					t.Fatalf("seed=%d step=%d: invalid turning point type at %d: %+v", seed, step, j, tp)
				}
				if !isFinite(tp.E) {
					t.Fatalf("seed=%d step=%d: non-finite turning point E at %d: %+v", seed, step, j, tp)
				}
				if j > 0 && ps.Stack[j-1].Type == tp.Type {
					t.Fatalf("seed=%d step=%d: non-alternating turning point types at %d: prev=%+v cur=%+v", seed, step, j, ps.Stack[j-1], tp)
				}
			}
		}
	}
}

// getenvInt is a tiny helper to keep tests deterministic and avoid additional deps.
func getenvInt(key string, def int) int {
	v := def
	if s := os.Getenv(key); s != "" {
		if _, err := fmt.Sscanf(s, "%d", &v); err == nil {
			return v
		}
	}
	return def
}
