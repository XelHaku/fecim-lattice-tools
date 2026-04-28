package comparison

import (
	"strings"
	"testing"

	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestBuildSnapshot_DescriptorIsFunctionalComparison(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	if snap.Descriptor.ID != viewmodel.ModuleComparison {
		t.Errorf("Descriptor.ID = %q, want %q", snap.Descriptor.ID, viewmodel.ModuleComparison)
	}
	if snap.Descriptor.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor.Status = %q, want %q", snap.Descriptor.Status, viewmodel.StatusFunctional)
	}
}

func TestBuildSnapshot_HasOneSectionPerArchitecture(t *testing.T) {
	archs := pkg.Architectures()
	snap := buildSnapshot(archs)
	if len(snap.Sections) != len(archs) {
		t.Fatalf("len(Sections) = %d, want %d", len(snap.Sections), len(archs))
	}
	for i, a := range archs {
		if snap.Sections[i].Title != a.Name {
			t.Errorf("Sections[%d].Title = %q, want %q", i, snap.Sections[i].Title, a.Name)
		}
		if snap.Sections[i].ID == "" {
			t.Errorf("Sections[%d].ID is empty", i)
		}
	}
}

func TestBuildSnapshot_SectionBodyIncludesPhysicalAndPerformanceFields(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	cpu := snap.Sections[0].Body
	for _, want := range []string{"Technology", "Process", "TDP", "TOPS"} {
		if !strings.Contains(cpu, want) {
			t.Errorf("CPU section body missing %q label\nbody: %s", want, cpu)
		}
	}
}

func TestBuildSnapshot_FeCIMSectionFlagsEstimatedValues(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	fecim := snap.Sections[2]
	if fecim.Title != "FeCIM CIM" {
		t.Fatalf("Sections[2].Title = %q, want FeCIM CIM", fecim.Title)
	}
	if !strings.Contains(strings.ToLower(fecim.Body), "estimated") {
		t.Errorf("FeCIM section body must flag IsEstimated=true (per honesty-audit policy)\nbody: %s", fecim.Body)
	}
}

func TestBuildSnapshot_HasArchitectureCountMetric(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	if len(snap.Metrics) == 0 {
		t.Fatal("snapshot has no metrics")
	}
	got := snap.Metrics[0]
	if got.ID != "count" {
		t.Errorf("Metrics[0].ID = %q, want count", got.ID)
	}
	if got.Value != "3" {
		t.Errorf("Metrics[0].Value = %q, want 3", got.Value)
	}
	if got.Confidence != "deterministic" {
		t.Errorf("Metrics[0].Confidence = %q, want deterministic", got.Confidence)
	}
}

func TestBuildSnapshot_DeterministicForSameInput(t *testing.T) {
	archs := pkg.Architectures()
	a := buildSnapshot(archs)
	b := buildSnapshot(archs)
	if !a.UpdatedAt.IsZero() || !b.UpdatedAt.IsZero() {
		t.Fatal("buildSnapshot must use zero time for deterministic tests")
	}
	if len(a.Sections) != len(b.Sections) {
		t.Fatalf("section counts differ across calls: %d vs %d", len(a.Sections), len(b.Sections))
	}
	for i := range a.Sections {
		if a.Sections[i] != b.Sections[i] {
			t.Errorf("Sections[%d] differs across calls\n  a: %+v\n  b: %+v", i, a.Sections[i], b.Sections[i])
		}
	}
}

func TestBuildSnapshot_EmptyInputProducesNoSections(t *testing.T) {
	snap := buildSnapshot(nil)
	if len(snap.Sections) != 0 {
		t.Errorf("nil input: len(Sections) = %d, want 0", len(snap.Sections))
	}
	if snap.Metrics[0].Value != "0" {
		t.Errorf("nil input: count metric = %q, want 0", snap.Metrics[0].Value)
	}
}
