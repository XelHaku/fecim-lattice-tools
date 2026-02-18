package comparisoncli

import (
	"testing"

	"fecim-lattice-tools/module5-comparison/pkg/comparison"
)

func TestBuildComparisonResult_PopulatesArchitectures(t *testing.T) {
	w := comparison.MNISTWorkload()
	comp := comparison.CompareArchitectures(w, 1, 10000)
	adv := comparison.CalculateAdvantages(comp)

	res := buildComparisonResult(comp, adv, "mnist", 10000)
	if res.Workload != "mnist" {
		t.Fatalf("workload=%q want mnist", res.Workload)
	}
	if len(res.Architectures) == 0 {
		t.Fatal("expected architectures in JSON result")
	}
	for _, a := range res.Architectures {
		if a.Name == "" {
			t.Fatal("architecture name should not be empty")
		}
	}
}
