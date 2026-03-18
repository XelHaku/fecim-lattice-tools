package export

import (
	"strings"
	"testing"
)

func TestDefaultSimulationMetadata_Fields(t *testing.T) {
	m := DefaultSimulationMetadata()

	if m.Tool == "" {
		t.Error("Tool field is empty")
	}
	if m.Version == "" {
		t.Error("Version field is empty")
	}
	if m.Disclaimer == "" {
		t.Error("Disclaimer field is empty")
	}
	if m.GeneratedAt == "" {
		t.Error("GeneratedAt field is empty")
	}
	if !strings.Contains(m.Disclaimer, "SIMULATION ONLY") {
		t.Errorf("Disclaimer should contain 'SIMULATION ONLY', got: %s", m.Disclaimer)
	}
	if !strings.Contains(m.Disclaimer, "honesty-audit.md") {
		t.Errorf("Disclaimer should reference honesty-audit.md, got: %s", m.Disclaimer)
	}
}

func TestDefaultSimulationMetadata_VersionFormat(t *testing.T) {
	m := DefaultSimulationMetadata()
	if !strings.Contains(m.Version, "education") {
		t.Errorf("Version should contain 'education' tag, got: %s", m.Version)
	}
}

func TestSPICEDisclaimer_ContainsWarning(t *testing.T) {
	d := SPICEDisclaimer()
	if !strings.Contains(d, "SIMULATION ONLY") {
		t.Errorf("SPICE disclaimer should contain 'SIMULATION ONLY', got:\n%s", d)
	}
	if !strings.Contains(d, "honesty-audit.md") {
		t.Errorf("SPICE disclaimer should reference honesty-audit.md, got:\n%s", d)
	}
}

func TestSPICEDisclaimer_UsesCommentSyntax(t *testing.T) {
	d := SPICEDisclaimer()
	lines := strings.Split(strings.TrimRight(d, "\n"), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "*") {
			t.Errorf("Line %d should start with '*' (SPICE comment), got: %q", i+1, line)
		}
	}
}
