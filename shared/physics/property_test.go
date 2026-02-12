package physics

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// materialEverett is a simple bounded Everett integral scaled to material Ps.
// It guarantees E(sat, -sat) == Ps, so Preisach polarization stays in [-Ps, +Ps].
type materialEverett struct {
	satE float64
	Ps   float64
}

func (e materialEverett) Calculate(alpha, beta float64) float64 {
	return e.Ps * (alpha - beta) / (2 * e.satE)
}

func randomValidMaterial(r *rand.Rand, idx int) *HZOMaterial {
	ps := randRange(r, 0.15, 0.60)         // C/m^2
	pr := randRange(r, 0.05*ps, 0.95*ps)   // enforce Pr <= Ps
	ec := randRange(r, 0.4e8, 2.5e8)       // V/m, strictly > 0
	thickness := randRange(r, 5e-9, 30e-9) // m
	area := randRange(r, 25e-18, 6400e-18) // m^2 (25nm^2..80nm^2)
	eps := randRange(r, 20, 45)

	return &HZOMaterial{
		Name:            fmt.Sprintf("rand-material-%02d", idx),
		Pr:              pr,
		Ps:              ps,
		Ec:              ec,
		Thickness:       thickness,
		Area:            area,
		Epsilon:         eps,
		EpsilonLF:       eps * 1.2,
		Tau:             randRange(r, 1e-10, 5e-9),
		Tau0:            randRange(r, 1e-14, 1e-11),
		Alpha:           randRange(r, 1.5, 3.0),
		K_dep:           randRange(r, 0.5e8, 4e8),
		CurieTemp:       randRange(r, 600, 900),
		CurieConst:      randRange(r, 0.8e5, 2.5e5),
		Gmin:            randRange(r, 0.5e-6, 5e-6),
		Gmax:            randRange(r, 20e-6, 250e-6),
		NumLevels:       30,
		TargetRangeFrac: 0.9,
	}
}

func indexAtE(Es []float64, target float64, tol float64) int {
	for i := range Es {
		if math.Abs(Es[i]-target) <= tol {
			return i
		}
	}
	return -1
}

func interpolateZeroCrossing(E1, P1, E2, P2 float64) (float64, bool) {
	if P1 == 0 {
		return E1, true
	}
	if P2 == 0 {
		return E2, true
	}
	if (P1 < 0 && P2 > 0) || (P1 > 0 && P2 < 0) {
		frac := -P1 / (P2 - P1)
		return E1 + frac*(E2-E1), true
	}
	return 0, false
}

func firstZeroCrossing(Es, Ps []float64) (float64, bool) {
	for i := 0; i+1 < len(Es); i++ {
		if ec, ok := interpolateZeroCrossing(Es[i], Ps[i], Es[i+1], Ps[i+1]); ok {
			return ec, true
		}
	}
	return 0, false
}

func TestProperty_PreisachInvariants_RandomMaterials(t *testing.T) {
	r := rand.New(rand.NewSource(20260212))
	const numMaterials = 20
	const pointsPerBranch = 400

	for i := 0; i < numMaterials; i++ {
		mat := randomValidMaterial(r, i)
		t.Run(mat.Name, func(t *testing.T) {
			if mat.Pr > mat.Ps {
				t.Fatalf("invalid random material: Pr=%e > Ps=%e", mat.Pr, mat.Ps)
			}
			if mat.Ec <= 0 {
				t.Fatalf("invalid random material: Ec=%e <= 0", mat.Ec)
			}

			satE := 3.0 * mat.Ec
			ps := NewPreisachStack(satE, materialEverett{satE: satE, Ps: mat.Ps})

			startP := ps.ComputePolarization(-satE)

			upE := make([]float64, 0, pointsPerBranch+1)
			upP := make([]float64, 0, pointsPerBranch+1)
			for k := 0; k <= pointsPerBranch; k++ {
				E := -satE + 2*satE*float64(k)/float64(pointsPerBranch)
				P := ps.Update(E)
				upE = append(upE, E)
				upP = append(upP, P)
			}

			downE := make([]float64, 0, pointsPerBranch+1)
			downP := make([]float64, 0, pointsPerBranch+1)
			for k := 0; k <= pointsPerBranch; k++ {
				E := satE - 2*satE*float64(k)/float64(pointsPerBranch)
				P := ps.Update(E)
				downE = append(downE, E)
				downP = append(downP, P)
			}

			endP := downP[len(downP)-1]

			// (1) Closed loop after full cycle.
			closeTol := 1e-6*mat.Ps + 1e-12
			if math.Abs(endP-startP) > closeTol {
				t.Fatalf("loop not closed: Pstart=%e Pend=%e |delta|=%e tol=%e", startP, endP, math.Abs(endP-startP), closeTol)
			}

			// (2) Positive loop area (energy dissipation): |∮ P dE| > 0.
			loopE := append(append([]float64{}, upE...), downE[1:]...)
			loopP := append(append([]float64{}, upP...), downP[1:]...)
			signedArea := 0.0
			for j := 0; j+1 < len(loopE); j++ {
				dE := loopE[j+1] - loopE[j]
				signedArea += 0.5 * (loopP[j] + loopP[j+1]) * dE
			}
			area := math.Abs(signedArea)
			if area <= 1e-15 {
				t.Fatalf("loop area not positive: area=%e", area)
			}

			// (3) Pr <= Ps on both branches at E=0.
			zeroTol := satE / float64(pointsPerBranch)
			upZero := indexAtE(upE, 0, zeroTol)
			downZero := indexAtE(downE, 0, zeroTol)
			if upZero < 0 || downZero < 0 {
				t.Fatalf("missing E=0 sample: upZero=%d downZero=%d", upZero, downZero)
			}
			prUp := math.Abs(upP[upZero])
			prDown := math.Abs(downP[downZero])
			boundTol := 1e-6*mat.Ps + 1e-12
			if prUp > mat.Ps+boundTol || prDown > mat.Ps+boundTol {
				t.Fatalf("remnant exceeds saturation: |Pr_up|=%e |Pr_down|=%e Ps=%e", prUp, prDown, mat.Ps)
			}

			// (4) Ec > 0 for ferroelectric materials.
			// This is a material invariant (ferroelectric definition), so we assert
			// directly on randomized valid materials and also ensure loop branches
			// have resolvable zero crossings.
			if mat.Ec <= 0 {
				t.Fatalf("coercive field must be positive for ferroelectric material: Ec=%e", mat.Ec)
			}
			ecUp, okUp := firstZeroCrossing(upE, upP)
			ecDown, okDown := firstZeroCrossing(downE, downP)
			if !okUp || !okDown {
				t.Fatalf("could not find zero crossing on major loop branches: up=%v down=%v", okUp, okDown)
			}
			if math.IsNaN(ecUp) || math.IsNaN(ecDown) || math.IsInf(ecUp, 0) || math.IsInf(ecDown, 0) {
				t.Fatalf("invalid zero-crossing coercive estimates: Ec_up=%e Ec_down=%e", ecUp, ecDown)
			}

			// (5) Monotonic major branches:
			// ascending branch (negative->positive sweep): P must not decrease.
			for j := 1; j < len(upP); j++ {
				if upP[j] < upP[j-1]-1e-12 {
					t.Fatalf("ascending branch non-monotone at j=%d: Pprev=%e P=%e", j, upP[j-1], upP[j])
				}
			}
			// descending branch (positive->negative sweep): P must not increase.
			for j := 1; j < len(downP); j++ {
				if downP[j] > downP[j-1]+1e-12 {
					t.Fatalf("descending branch non-monotone at j=%d: Pprev=%e P=%e", j, downP[j-1], downP[j])
				}
			}
		})
	}
}
