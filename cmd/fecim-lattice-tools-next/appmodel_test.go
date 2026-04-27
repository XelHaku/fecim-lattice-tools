package main

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestDefaultAppSpecNamesFutureDefaultShell(t *testing.T) {
	spec := DefaultAppSpec()

	if spec.Title != "FeCIM Lattice Tools Next" {
		t.Fatalf("Title = %q, want FeCIM Lattice Tools Next", spec.Title)
	}
	if spec.Command != "fecim-lattice-tools-next" {
		t.Fatalf("Command = %q, want fecim-lattice-tools-next", spec.Command)
	}
	if spec.Width < 1200 || spec.Height < 800 {
		t.Fatalf("unexpected window size: %dx%d", spec.Width, spec.Height)
	}
}

func TestBuildPlaceholderPortsCoversAllKnownDescriptors(t *testing.T) {
	ports := BuildPlaceholderPorts()
	descriptors := viewmodel.KnownDescriptors()

	if len(ports) != len(descriptors) {
		t.Fatalf("port count = %d, want %d", len(ports), len(descriptors))
	}

	for i, port := range ports {
		got := port.Descriptor()
		want := descriptors[i]
		if got != want {
			t.Fatalf("port[%d] descriptor = %#v, want %#v", i, got, want)
		}
		snapshot := port.Snapshot()
		if len(snapshot.Sections) == 0 {
			t.Fatalf("port[%d] snapshot has no sections", i)
		}
	}
}
