package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleEngineRegistryBuildsPortsForKnownDescriptors(t *testing.T) {
	entries := ModuleEngineRegistry()
	descriptors := viewmodel.KnownDescriptors()

	if len(entries) != len(descriptors) {
		t.Fatalf("registry entries = %d, want %d known descriptors", len(entries), len(descriptors))
	}
	for i, entry := range entries {
		want := descriptors[i]
		if entry.Descriptor.ID != want.ID {
			t.Fatalf("entry[%d] ID = %q, want %q", i, entry.Descriptor.ID, want.ID)
		}
		if entry.New == nil {
			t.Fatalf("entry[%d] factory is nil", i)
		}
		port := entry.New()
		if port == nil {
			t.Fatalf("entry[%d] factory returned nil", i)
		}
		if got := port.Snapshot().Descriptor.ID; got != want.ID {
			t.Fatalf("entry[%d] snapshot descriptor ID = %q, want %q", i, got, want.ID)
		}
	}
}
