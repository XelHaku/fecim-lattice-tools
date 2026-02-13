package gui

import (
	"strings"
	"testing"
)

func TestM5UX02_EvidenceFirstIncludesAssumptionsOutputsCaveat(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	text := PlainTextEvidence("Evidence", run)
	if !strings.Contains(text, "Inputs:") || !strings.Contains(text, "Outputs:") {
		t.Fatalf("evidence text missing assumptions/output sections: %s", text)
	}
}
