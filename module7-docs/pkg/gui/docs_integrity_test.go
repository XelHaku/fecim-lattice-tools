package gui

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

var (
	headingRE       = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+?)\s*$`)
	markdownLinkRE  = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	imagePrefixRE   = regexp.MustCompile(`^!\[[^\]]*\]\(`)
	htmlAnchorRE    = regexp.MustCompile(`(?i)<a\s+id=["']([^"']+)["']`)
)

func TestModule7DocsIntegrity(t *testing.T) {
	docsRoot := resolveDocsRoot(t)
	topicFiles := collectMarkdownFiles(t, docsRoot)

	if len(topicFiles) == 0 {
		t.Fatalf("no markdown topics found under %s", docsRoot)
	}

	t.Run("all doc topics load without error", func(t *testing.T) {
		for _, file := range topicFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed loading topic %s: %v", file, err)
			}
			if len(strings.TrimSpace(string(content))) == 0 {
				t.Fatalf("topic is empty: %s", file)
			}
		}
	})

	t.Run("search index covers all topics", func(t *testing.T) {
		index := NewSearchIndex(docsRoot)
		index.Build()

		index.mu.RLock()
		defer index.mu.RUnlock()

		if len(index.docs) == 0 {
			t.Fatal("search index contains no docs")
		}

		for _, file := range topicFiles {
			if _, ok := index.docs[file]; !ok {
				t.Fatalf("search index missing topic: %s", file)
			}
		}
	})

	t.Run("toc structure is consistent", func(t *testing.T) {
		for _, file := range topicFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			levels := extractHeadingLevels(string(content))
			if len(levels) == 0 {
				t.Fatalf("no headings found in %s", file)
			}
			if levels[0] != 1 {
				t.Fatalf("first heading should be H1 in %s, got H%d", file, levels[0])
			}

			seen := map[int]bool{1: true}
			for _, level := range levels[1:] {
				if level == 1 {
					seen[level] = true
					continue
				}
				hasParent := false
				for parent := level - 1; parent >= 1; parent-- {
					if seen[parent] {
						hasParent = true
						break
					}
				}
				if !hasParent {
					t.Fatalf("orphan section in %s: H%d has no parent heading", file, level)
				}
				seen[level] = true
			}
		}
	})

	t.Run("all internal cross-references resolve", func(t *testing.T) {
		anchorsByFile := make(map[string]map[string]struct{}, len(topicFiles))
		for _, file := range topicFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			anchorsByFile[file] = extractAnchors(string(content))
		}

		for _, file := range topicFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			links := extractMarkdownLinks(string(content))
			for _, raw := range links {
				if !isInternalDocLink(raw) {
					continue
				}

				targetFile, targetAnchor := resolveInternalLink(file, raw)
				if targetFile == "" {
					t.Fatalf("failed to resolve link %q in %s", raw, file)
				}
				rel, err := filepath.Rel(docsRoot, targetFile)
				if err != nil || strings.HasPrefix(rel, "..") {
					continue // only validate cross-references within the Module 7 docs corpus
				}
				if _, ok := anchorsByFile[targetFile]; !ok {
					t.Fatalf("link %q in %s points to missing doc: %s", raw, file, targetFile)
				}
				if targetAnchor != "" {
					anchors := anchorsByFile[targetFile]
					if _, ok := anchors[targetAnchor]; !ok {
						t.Fatalf("link %q in %s points to missing anchor %q in %s", raw, file, targetAnchor, targetFile)
					}
				}
			}
		}
	})
}

func resolveDocsRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Clean(filepath.Join("..", "..", "..", "docs", "documentation"))
	if !filepath.IsAbs(root) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("unable to resolve working directory: %v", err)
		}
		root = filepath.Join(cwd, root)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("docs root not found at %s: %v", root, err)
	}
	return root
}

func collectMarkdownFiles(t *testing.T, docsRoot string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(docsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(name), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs root: %v", err)
	}
	sort.Strings(files)
	return files
}

func extractHeadingLevels(content string) []int {
	matches := headingRE.FindAllStringSubmatch(content, -1)
	levels := make([]int, 0, len(matches))
	for _, m := range matches {
		levels = append(levels, len(m[1]))
	}
	return levels
}

func extractAnchors(content string) map[string]struct{} {
	anchors := make(map[string]struct{})
	for _, m := range headingRE.FindAllStringSubmatch(content, -1) {
		a := slugifyAnchor(m[2])
		if a != "" {
			anchors[a] = struct{}{}
		}
	}
	for _, m := range htmlAnchorRE.FindAllStringSubmatch(content, -1) {
		a := strings.TrimSpace(strings.ToLower(m[1]))
		if a != "" {
			anchors[a] = struct{}{}
		}
	}
	return anchors
}

func extractMarkdownLinks(content string) []string {
	matches := markdownLinkRE.FindAllStringSubmatch(content, -1)
	links := make([]string, 0, len(matches))
	for _, m := range matches {
		full := m[0]
		if imagePrefixRE.MatchString(full) {
			continue
		}
		links = append(links, strings.TrimSpace(m[1]))
	}
	return links
}

func isInternalDocLink(link string) bool {
	if link == "" {
		return false
	}
	lower := strings.ToLower(link)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "tel:") ||
		strings.HasPrefix(lower, "glossary://") {
		return false
	}
	if strings.HasPrefix(lower, "javascript:") {
		return false
	}
	return true
}

func resolveInternalLink(sourceFile, rawLink string) (string, string) {
	link := strings.TrimSpace(rawLink)
	parts := strings.SplitN(link, "#", 2)
	filePart := strings.TrimSpace(parts[0])
	anchorPart := ""
	if len(parts) == 2 {
		anchorPart = slugifyAnchor(parts[1])
	}

	if filePart == "" {
		return sourceFile, anchorPart
	}
	if q := strings.Index(filePart, "?"); q >= 0 {
		filePart = filePart[:q]
	}

	resolved := filePart
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(sourceFile), resolved)
	}
	resolved = filepath.Clean(resolved)
	if strings.EqualFold(filepath.Ext(resolved), ".md") {
		return resolved, anchorPart
	}

	if info, err := os.Stat(resolved); err == nil {
		if info.IsDir() {
			readme := filepath.Join(resolved, "README.md")
			if _, err := os.Stat(readme); err == nil {
				return readme, anchorPart
			}
		}
		return resolved, anchorPart
	}

	if filepath.Ext(resolved) == "" {
		candidate := resolved + ".md"
		if _, err := os.Stat(candidate); err == nil {
			return candidate, anchorPart
		}
	}
	return resolved, anchorPart
}

func slugifyAnchor(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	return out
}
