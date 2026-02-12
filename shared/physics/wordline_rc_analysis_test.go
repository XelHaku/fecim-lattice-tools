package physics_test

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

type wlTechNode struct {
	name       string
	rScale     float64
	cGateFF    float64
	pulseWidth float64
}

type wlRow struct {
	TechNode      string
	N             int
	RSegOhm       float64
	CSegFF        float64
	TauNS         float64
	PulseWidthNS  float64
	Viable        bool
}

func computeWordlineRows() []wlRow {
	geom := arraysim.DefaultCellGeometry()

	// Rseg = rho * pitch / (w * t)
	rBase := geom.MetalResistivity * geom.PitchX / (geom.WireWidth * geom.WireThickness)

	// Cwire ~= 0.1 fF/um * pitch(um)
	pitchUM := geom.PitchX * 1e6
	cWireFF := 0.1 * pitchUM

	nodes := []wlTechNode{
		{name: "SKY130", rScale: 1.00, cGateFF: 3.00, pulseWidth: 10.0},
		{name: "65nm", rScale: 1.30, cGateFF: 1.80, pulseWidth: 10.0},
		{name: "28nm", rScale: 1.80, cGateFF: 1.00, pulseWidth: 10.0},
		{name: "14nm", rScale: 2.40, cGateFF: 0.60, pulseWidth: 10.0},
	}

	sizes := []int{8, 16, 32, 64, 128, 256}
	rows := make([]wlRow, 0, len(nodes)*len(sizes))

	for _, node := range nodes {
		rSeg := rBase * node.rScale
		cSegFF := cWireFF + node.cGateFF // 1T1R estimate: wire + transistor gate cap
		cSeg := cSegFF * 1e-15

		for _, n := range sizes {
			// Elmore delay: tau = 0.5 * (N*Rseg) * (N*Cseg)
			tauNS := 0.5 * (float64(n) * rSeg) * (float64(n) * cSeg) * 1e9
			rows = append(rows, wlRow{
				TechNode:     node.name,
				N:            n,
				RSegOhm:      rSeg,
				CSegFF:       cSegFF,
				TauNS:        tauNS,
				PulseWidthNS: node.pulseWidth,
				Viable:       tauNS <= node.pulseWidth,
			})
		}
	}

	return rows
}

func TestWordlineRCPracticalCeiling(t *testing.T) {
	rows := computeWordlineRows()
	if len(rows) == 0 {
		t.Fatal("expected non-empty WL RC rows")
	}

	// Validate geometry-derived baseline Rseg from module4 defaults.
	baseline := rows[0].RSegOhm
	if math.Abs(baseline-0.790625) > 1e-6 {
		t.Fatalf("unexpected SKY130 Rseg: got %.9f ohm, expected ~0.790625 ohm", baseline)
	}

	// Ensure tau increases monotonically with N for each node and remains <= 10 ns in tested range.
	lastTau := map[string]float64{}
	maxViableSky130 := 0

	for _, r := range rows {
		if prev, ok := lastTau[r.TechNode]; ok && r.TauNS <= prev {
			t.Fatalf("non-monotonic tau for %s: N=%d tau=%.6f ns <= prev %.6f ns", r.TechNode, r.N, r.TauNS, prev)
		}
		lastTau[r.TechNode] = r.TauNS

		if r.TechNode == "SKY130" && r.Viable {
			maxViableSky130 = r.N
		}
		if !r.Viable {
			t.Fatalf("expected tested N to be viable for %s, got N=%d tau=%.6f ns > %.2f ns", r.TechNode, r.N, r.TauNS, r.PulseWidthNS)
		}
	}

	if maxViableSky130 != 256 {
		t.Fatalf("expected max viable SKY130 size in tested set to be 256, got %d", maxViableSky130)
	}

	for _, r := range rows {
		t.Logf("%s N=%3d Rseg=%.6f ohm Cseg=%.3f fF tau=%.6f ns pulse=%.2f ns viable=%v",
			r.TechNode, r.N, r.RSegOhm, r.CSegFF, r.TauNS, r.PulseWidthNS, r.Viable)
	}
}

func TestWordlineRCZeroTOneRReference(t *testing.T) {
	geom := arraysim.DefaultCellGeometry()
	rSeg := geom.MetalResistivity * geom.PitchX / (geom.WireWidth * geom.WireThickness)
	cWireFF := 0.1 * (geom.PitchX * 1e6)
	cWire := cWireFF * 1e-15

	n := 256.0
	tauNS := 0.5 * (n * rSeg) * (n * cWire) * 1e9
	if tauNS >= 10.0 {
		t.Fatalf("0T1R wire-only tau unexpectedly exceeds pulse width: %.6f ns", tauNS)
	}
	if tauNS <= 0 {
		t.Fatalf("expected positive 0T1R tau, got %.6f ns", tauNS)
	}
}
