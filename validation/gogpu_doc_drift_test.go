package validation

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type staleGogpuMigrationDocFinding struct {
	Path   string
	Line   int
	Phrase string
	Text   string
}

func (f staleGogpuMigrationDocFinding) String() string {
	return fmt.Sprintf("%s:%d contains %q: %s", f.Path, f.Line, f.Phrase, f.Text)
}

var staleGogpuMigrationDocPhrases = []string{
	"fecim-lattice-tools-next",
	"future default",
	"future-default",
	"future gogpu",
	"future `gogpu/ui`",
	"future zero",
	"next shell",
	"current Fyne",
	"stable Fyne",
	"Fyne remains",
	"reaches parity",
}

func TestStaleGogpuMigrationDocWordingScannerFlagsLiveDocsAndIgnoresArchivedSpecs(t *testing.T) {
	root := t.TempDir()
	writeMarkdown(t, root, "go.mod", "module example.com/docguard\n")
	writeMarkdown(t, root, "AGENTS.md", "Default UI work must not point at the future default gogpu shell.\n")
	writeMarkdown(t, root, filepath.Join("docs", "superpowers", "plans", "old-plan.md"), "The future default gogpu shell was still planned here.\n")
	writeMarkdown(t, root, filepath.Join("docs", "superpowers", "specs", "old-spec.md"), "The next shell was still planned here.\n")

	findings, err := staleGogpuMigrationDocWordingFindings(root)
	if err != nil {
		t.Fatalf("staleGogpuMigrationDocWordingFindings error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("finding count = %d, want 1: %+v", len(findings), findings)
	}
	if findings[0].Path != "AGENTS.md" {
		t.Fatalf("finding path = %q, want AGENTS.md", findings[0].Path)
	}
	if !strings.Contains(findings[0].Text, "future default gogpu") {
		t.Fatalf("finding text = %q, want stale gogpu wording", findings[0].Text)
	}
}

func TestLiveContributorDocsAvoidStaleGogpuMigrationWording(t *testing.T) {
	root := repoRoot(t)
	findings, err := staleGogpuMigrationDocWordingFindings(root)
	if err != nil {
		t.Fatalf("staleGogpuMigrationDocWordingFindings error: %v", err)
	}
	if len(findings) == 0 {
		return
	}

	var report strings.Builder
	for _, finding := range findings {
		report.WriteString("\n")
		report.WriteString(finding.String())
	}
	t.Fatalf("live contributor docs contain stale gogpu migration wording:%s", report.String())
}

func staleGogpuMigrationDocWordingFindings(root string) ([]staleGogpuMigrationDocFinding, error) {
	var findings []staleGogpuMigrationDocFinding
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if shouldSkipDocDriftDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		fileFindings, err := staleGogpuMigrationDocFindingsInFile(path, rel)
		if err != nil {
			return err
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	return findings, err
}

func staleGogpuMigrationDocFindingsInFile(path, rel string) ([]staleGogpuMigrationDocFinding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", rel, err)
	}
	defer file.Close()

	var findings []staleGogpuMigrationDocFinding
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		lowerLine := strings.ToLower(line)
		for _, phrase := range staleGogpuMigrationDocPhrases {
			if strings.Contains(lowerLine, strings.ToLower(phrase)) {
				findings = append(findings, staleGogpuMigrationDocFinding{
					Path:   rel,
					Line:   lineNumber,
					Phrase: phrase,
					Text:   strings.TrimSpace(line),
				})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", rel, err)
	}
	return findings, nil
}

func shouldSkipDocDriftDir(rel string) bool {
	switch rel {
	case ".git", ".cognee_system", ".worktrees", "artifacts", "logs", "output", "recordings", "screenshots":
		return true
	}
	return strings.HasPrefix(rel, "docs/superpowers/plans/") ||
		strings.HasPrefix(rel, "docs/superpowers/specs/") ||
		rel == "docs/superpowers/plans" ||
		rel == "docs/superpowers/specs"
}

func writeMarkdown(t *testing.T, root, rel, contents string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root containing go.mod")
		}
		dir = parent
	}
}
