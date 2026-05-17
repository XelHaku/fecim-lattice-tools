package viewmodel

import (
	"errors"
	"testing"
)

func TestKnownDescriptorsCoverSevenModules(t *testing.T) {
	descriptors := KnownDescriptors()

	if len(descriptors) != 7 {
		t.Fatalf("descriptor count = %d, want 7", len(descriptors))
	}

	wantIDs := []ModuleID{
		ModuleHysteresis,
		ModuleCrossbar,
		ModuleMNIST,
		ModuleCircuits,
		ModuleComparison,
		ModuleEDA,
		ModuleDocs,
	}
	for i, want := range wantIDs {
		if descriptors[i].ID != want {
			t.Fatalf("descriptor[%d].ID = %q, want %q", i, descriptors[i].ID, want)
		}
		if descriptors[i].Title == "" {
			t.Fatalf("descriptor[%d] has empty title", i)
		}
		if descriptors[i].Description == "" {
			t.Fatalf("descriptor[%d] has empty description", i)
		}
	}
}

func TestKnownDescriptorsAreFunctionalForReleasedGogpuShell(t *testing.T) {
	for _, descriptor := range KnownDescriptors() {
		if descriptor.Status != StatusFunctional {
			t.Errorf("descriptor %s Status = %q, want %q", descriptor.ID, descriptor.Status, StatusFunctional)
		}
	}
}

func TestStaticModuleSnapshotIsDeterministic(t *testing.T) {
	descriptor := ModuleDescriptor{
		ID:          ModuleDocs,
		Title:       "Documentation",
		Description: "Documentation and validation references.",
		Status:      StatusPlaceholder,
	}
	port := NewStaticModule(descriptor, []Section{
		{
			ID:    "scope",
			Title: "Scope",
			Body:  "Placeholder card for future gogpu/ui module port.",
		},
	})

	snapshot := port.Snapshot()

	if snapshot.Descriptor != descriptor {
		t.Fatalf("snapshot descriptor = %#v, want %#v", snapshot.Descriptor, descriptor)
	}
	if len(snapshot.Sections) != 1 {
		t.Fatalf("section count = %d, want 1", len(snapshot.Sections))
	}
	if snapshot.Sections[0].ID != "scope" {
		t.Fatalf("section ID = %q, want scope", snapshot.Sections[0].ID)
	}
	if !snapshot.UpdatedAt.IsZero() {
		t.Fatalf("UpdatedAt = %v, want zero for deterministic placeholder", snapshot.UpdatedAt)
	}
}

func TestStaticModuleRejectsActions(t *testing.T) {
	port := NewStaticModule(ModuleDescriptor{
		ID:          ModuleComparison,
		Title:       "Comparison",
		Description: "Comparison placeholder.",
		Status:      StatusPlaceholder,
	}, nil)

	err := port.ApplyAction(Action{ID: "run"})
	if !errors.Is(err, ErrUnsupportedAction) {
		t.Fatalf("ApplyAction error = %v, want ErrUnsupportedAction", err)
	}
}
