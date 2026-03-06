package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type module4ParityContract struct {
	Version       string `json:"version"`
	Profile       string `json:"profile"`
	GeneratedUnix int64  `json:"generated_unix"`
	Records       []any  `json:"records"`
}

func readModule4ParityContract(t *testing.T, p string) module4ParityContract {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec module4ParityContract
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if rec.Version != "v1" {
		t.Fatalf("%s version=%q want v1", p, rec.Version)
	}
	if rec.Profile == "" {
		t.Fatalf("%s profile empty", p)
	}
	if rec.GeneratedUnix != 0 {
		t.Fatalf("%s generated_unix=%d want 0", p, rec.GeneratedUnix)
	}
	return rec
}

func TestModule4ParityBoundaryConsistency_RootVsModulePath(t *testing.T) {
	repoRoot := filepath.Clean("..")
	rootPath := filepath.Join(repoRoot, "output", "regression", "module4", "gui_vs_headless_parity.json")
	modulePath := filepath.Join(repoRoot, "module4-circuits", "pkg", "gui", "output", "regression", "module4", "gui_vs_headless_parity.json")

	rootRec := readModule4ParityContract(t, rootPath)
	moduleRec := readModule4ParityContract(t, modulePath)

	if rootRec.Profile != moduleRec.Profile {
		t.Fatalf("profile mismatch root=%q module=%q", rootRec.Profile, moduleRec.Profile)
	}
	if len(rootRec.Records) != len(moduleRec.Records) {
		t.Fatalf("records length mismatch root=%d module=%d", len(rootRec.Records), len(moduleRec.Records))
	}
	if !reflect.DeepEqual(rootRec.Records, moduleRec.Records) {
		t.Fatalf("records payload mismatch between root and module parity artifacts")
	}
}
