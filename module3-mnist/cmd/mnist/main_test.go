package mnistcli

import (
	"encoding/json"
	"testing"
)

// TestParseLevelList covers the comma-separated level parser used by --core-levels
// and --export-levels flags.
func TestParseLevelList(t *testing.T) {
	cases := []struct {
		in      string
		want    []int
		wantErr bool
	}{
		{"8,16,24,31", []int{8, 16, 24, 31}, false},
		{"4", []int{4}, false},
		{"16,8", []int{8, 16}, false},   // must be sorted
		{"8,8,16", []int{8, 16}, false}, // deduplication
		{"", nil, false},
		{" ", nil, false},
		{"bad", nil, true},
		{"8,bad,16", nil, true},
	}
	for _, tc := range cases {
		got, err := parseLevelList(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseLevelList(%q): expected error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseLevelList(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if len(got) != len(tc.want) {
			t.Errorf("parseLevelList(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("parseLevelList(%q)[%d] = %d, want %d", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}

// TestParseDirList covers the comma-separated directory parser.
func TestParseDirList(t *testing.T) {
	cases := []struct {
		in      string
		wantLen int
		wantErr bool
	}{
		{"out1,out2", 2, false},
		{"single", 1, false},
		{"", 0, false},
		{" ", 0, false},
	}
	for _, tc := range cases {
		got, err := parseDirList(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseDirList(%q): expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseDirList(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if len(got) != tc.wantLen {
			t.Errorf("parseDirList(%q): got %d dirs, want %d", tc.in, len(got), tc.wantLen)
		}
	}
}

// TestMNISTConfigDefaults checks that a zero-value MNISTConfig marshals cleanly.
func TestMNISTConfigDefaults(t *testing.T) {
	cfg := MNISTConfig{HiddenSize: 128, NoiseLevel: 0.02, Epochs: 5, Levels: []int{8, 16, 30}}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(MNISTConfig): %v", err)
	}
	var rt MNISTConfig
	if err := json.Unmarshal(b, &rt); err != nil {
		t.Fatalf("json.Unmarshal(MNISTConfig): %v", err)
	}
	if rt.HiddenSize != cfg.HiddenSize {
		t.Errorf("HiddenSize round-trip: got %d, want %d", rt.HiddenSize, cfg.HiddenSize)
	}
	if rt.Epochs != cfg.Epochs {
		t.Errorf("Epochs round-trip: got %d, want %d", rt.Epochs, cfg.Epochs)
	}
}

// TestEvaluationResultJSONRoundtrip verifies EvaluationResult serialises and
// deserialises correctly (used for --json CLI output).
func TestEvaluationResultJSONRoundtrip(t *testing.T) {
	orig := EvaluationResult{
		Samples:     1000,
		Accuracy:    0.876,
		FPAccuracy:  0.912,
		CIMAccuracy: 0.876,
		AgreeRate:   0.941,
		AvgKL:       0.023,
		AvgEnergy:   1.234,
		Levels:      30,
	}
	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var rt EvaluationResult
	if err := json.Unmarshal(b, &rt); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if rt.Accuracy != orig.Accuracy {
		t.Errorf("Accuracy: got %f, want %f", rt.Accuracy, orig.Accuracy)
	}
	if rt.Samples != orig.Samples {
		t.Errorf("Samples: got %d, want %d", rt.Samples, orig.Samples)
	}
}

// TestResolveWeightsPathExplicit ensures an explicit path is returned unchanged.
func TestResolveWeightsPathExplicit(t *testing.T) {
	got, err := resolveWeightsPath("/tmp/weights.json")
	if err != nil {
		t.Fatalf("resolveWeightsPath(explicit): %v", err)
	}
	if got != "/tmp/weights.json" {
		t.Errorf("resolveWeightsPath: got %q, want %q", got, "/tmp/weights.json")
	}
}
