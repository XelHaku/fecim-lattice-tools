package hysteresis

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestBuildFORCMetadataExportPayloadSummarizesState(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunFORC, Kind: viewmodel.ActionCommand, Payload: map[string]string{"reversals": "9"}}); err != nil {
		t.Fatalf("run FORC: %v", err)
	}

	payload := buildFORCMetadataExportPayload(m.state)

	if payload.Material != m.state.SelectedMaterial {
		t.Fatalf("material = %q, want %q", payload.Material, m.state.SelectedMaterial)
	}
	if payload.Curves != m.state.FORC.Curves || payload.DensityRows != m.state.FORC.DensityRows || payload.DensityCols != m.state.FORC.DensityCols {
		t.Fatalf("FORC dimensions = %+v, want state dimensions", payload)
	}
	if !strings.Contains(payload.Boundary, "Not measured") {
		t.Fatalf("boundary = %q, want simulation boundary", payload.Boundary)
	}
}
