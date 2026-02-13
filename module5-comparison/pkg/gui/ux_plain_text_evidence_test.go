package gui

import (
	"strings"
	"testing"
)

func TestM5UX03_PlainTextEvidencePanelContent(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	txt := PlainTextEvidence("Energy chart evidence", run)
	if !strings.Contains(txt, "Energy chart evidence") || !strings.Contains(txt, "±") {
		t.Fatalf("plain-text evidence should include title and uncertainty: %s", txt)
	}
}
