package arraysim

import (
	"strings"
	"testing"
)

func TestBuildDesignSpacePointsAndPareto(t *testing.T) {
	points := BuildDesignSpacePoints([]int{8, 16}, []int{4, 6}, []string{"FeFET", "RRAM"})
	if got, want := len(points), 8; got != want {
		t.Fatalf("len(points)=%d want %d", got, want)
	}
	front := ParetoFront(points)
	if len(front) == 0 {
		t.Fatal("pareto front empty")
	}
	for _, p := range front {
		if p.ArraySize <= 0 || p.ADCBits <= 0 || p.Device == "" {
			t.Fatalf("invalid pareto point: %+v", p)
		}
	}
}

func TestExportParetoCSV(t *testing.T) {
	points := ParetoFront(BuildDesignSpacePoints([]int{8}, []int{5}, []string{"FeFET"}))
	var b strings.Builder
	if err := ExportParetoCSV(points, &b); err != nil {
		t.Fatalf("export csv: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, "array_size,adc_bits,device,latency_ns,energy_pj,accuracy") {
		t.Fatalf("missing csv header: %q", out)
	}
	if !strings.Contains(out, "FeFET") {
		t.Fatalf("missing device in csv: %q", out)
	}
}
