package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileWithLimitRejectsOversize(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "big.yaml")
	if err := os.WriteFile(p, []byte(strings.Repeat("a", 2048)), 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if _, err := readFileWithLimit(p, 1024); err == nil {
		t.Fatal("expected oversize file error")
	}
}
