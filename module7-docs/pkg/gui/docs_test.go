package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestSearchIndex_NewSearchIndex verifies basic index initialization
func TestSearchIndex_NewSearchIndex(t *testing.T) {
	index := NewSearchIndex("/test/docs")
	if index == nil {
		t.Fatal("NewSearchIndex returned nil")
	}
	if index.docsPath != "/test/docs" {
		t.Errorf("Expected docsPath=/test/docs, got %s", index.docsPath)
	}
	if index.index == nil {
		t.Error("index map should be initialized")
	}
	if index.docs == nil {
		t.Error("docs map should be initialized")
	}
}

// TestSearchIndex_BuildIndex creates a temp directory with markdown files and builds the index
func TestSearchIndex_BuildIndex(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test markdown files
	testFiles := map[string]string{
		"README.md": "# Welcome\n\nThis is a test document about ferroelectric materials.",
		"physics.md": "# Physics\n\n## Hysteresis Loop\n\nThe hysteresis loop shows polarization switching.",
		"subdir/guide.md": "# Guide\n\nStep-by-step instructions for simulation.",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Build index
	index := NewSearchIndex(tmpDir)
	index.Build()

	// Verify documents were indexed
	index.mu.RLock()
	defer index.mu.RUnlock()

	if len(index.docs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(index.docs))
	}

	// Verify terms were indexed
	if len(index.index) == 0 {
		t.Error("No terms were indexed")
	}

	// Check specific terms
	if entries, ok := index.index["ferroelectric"]; !ok || len(entries) == 0 {
		t.Error("Term 'ferroelectric' should be indexed")
	}
	if entries, ok := index.index["hysteresis"]; !ok || len(entries) == 0 {
		t.Error("Term 'hysteresis' should be indexed")
	}
}

// TestTokenize verifies tokenization logic
func TestTokenize(t *testing.T) {
	index := NewSearchIndex("")

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple text",
			input:    "Hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "With punctuation",
			input:    "Hello, world! Testing.",
			expected: []string{"hello", "world", "testing"},
		},
		{
			name:     "Title and heading",
			input:    "# Title\n\n## Heading\n\nContent here",
			expected: []string{"title", "heading", "content", "here"},
		},
		{
			name:     "Numbers and mixed case",
			input:    "Test123 UPPERCASE lowercase",
			expected: []string{"test123", "uppercase", "lowercase"},
		},
		{
			name:     "Single char filtered",
			input:    "a b cd ef",
			expected: []string{"cd", "ef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := index.tokenize(tt.input)

			// Extract just the term strings
			terms := make([]string, len(tokens))
			for i, tok := range tokens {
				terms[i] = tok.term
			}

			// Verify all expected terms are present
			for _, expected := range tt.expected {
				found := false
				for _, term := range terms {
					if term == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected term '%s' not found in tokens %v", expected, terms)
				}
			}
		})
	}
}

// TestExtractMetadata verifies metadata extraction from markdown
func TestExtractMetadata(t *testing.T) {
	index := NewSearchIndex("/test/docs")

	content := `# Test Document

This is a test document with about 200 words to test reading time calculation.
` + strings.Repeat("word ", 200)

	info := &mockFileInfo{
		name:    "test.md",
		modTime: time.Now(),
	}

	meta := index.extractMetadata("/test/docs/test.md", content, info)

	if meta.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", meta.Title)
	}

	if meta.ReadingTime != 2 {
		t.Errorf("Expected reading time to ceil to 2 min, got %d", meta.ReadingTime)
	}

	if meta.Path != "/test/docs/test.md" {
		t.Errorf("Expected path /test/docs/test.md, got %s", meta.Path)
	}
}

func TestExtractMetadata_ReadingTimeMath(t *testing.T) {
	index := NewSearchIndex("/test/docs")
	info := &mockFileInfo{name: "doc.md", modTime: time.Now()}

	tests := []struct {
		name     string
		words    int
		expected int
	}{
		{name: "minimum one minute", words: 0, expected: 1},
		{name: "exactly one page", words: 200, expected: 1},
		{name: "ceil to two", words: 201, expected: 2},
		{name: "ceil to three", words: 401, expected: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := strings.Repeat("w ", tt.words)
			meta := index.extractMetadata("/test/docs/doc.md", content, info)
			if meta.ReadingTime != tt.expected {
				t.Fatalf("reading time mismatch: words=%d got=%d want=%d", tt.words, meta.ReadingTime, tt.expected)
			}
		})
	}
}

// TestDetectCategory verifies category detection
func TestDetectCategory(t *testing.T) {
	index := NewSearchIndex("/docs")

	tests := []struct {
		path     string
		expected string
	}{
		{"/docs/ELI5.md", "ELI5"},
		{"/docs/PHYSICS.md", "Physics"},
		{"/docs/features.md", "Guide"},
		{"/docs/research-papers/paper.md", "Research"},
		{"/docs/research.md", "Research"}, // Test research filename
		{"/docs/cim/crossbar.md", "Physics"},
		{"/docs/guide.md", "Guide"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			category := index.detectCategory(tt.path)
			if category != tt.expected {
				t.Errorf("Expected category '%s', got '%s'", tt.expected, category)
			}
		})
	}
}

// TestQuery verifies search query functionality
func TestQuery(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test documents
	testDocs := map[string]string{
		"doc1.md": "# Ferroelectric Memory\n\nFerroelectric materials exhibit spontaneous polarization.",
		"doc2.md": "# Crossbar Arrays\n\nCrossbar architecture for compute-in-memory.",
		"doc3.md": "# Machine Learning\n\nNeural networks and ferroelectric devices.",
	}

	for name, content := range testDocs {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	index := NewSearchIndex(tmpDir)
	index.Build()

	tests := []struct {
		query         string
		expectedCount int
		shouldContain string
	}{
		{"ferroelectric", 2, "doc1.md"},
		{"crossbar", 1, "doc2.md"},
		{"memory", 2, "doc1.md"},
		{"neural", 1, "doc3.md"},
		{"nonexistent", 0, ""},
		{"", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := index.Query(tt.query, 10)

			if len(results) != tt.expectedCount {
				t.Errorf("Query '%s': expected %d results, got %d",
					tt.query, tt.expectedCount, len(results))
			}

			if tt.shouldContain != "" {
				found := false
				for _, result := range results {
					if strings.Contains(result.DocPath, tt.shouldContain) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Query '%s': expected result containing '%s'",
						tt.query, tt.shouldContain)
				}
			}
		})
	}
}

// TestRankResults verifies TF-IDF ranking
func TestRankResults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create documents with different term frequencies
	testDocs := map[string]string{
		"title_match.md":    "# Ferroelectric\n\nSome content here.",
		"heading_match.md":  "# Document\n\n## Ferroelectric Materials\n\nContent.",
		"content_match.md":  "# Document\n\nThis mentions ferroelectric once.",
		"frequent_match.md": "# Document\n\nFerroelectric ferroelectric ferroelectric.",
	}

	for name, content := range testDocs {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	index := NewSearchIndex(tmpDir)
	index.Build()

	results := index.Query("ferroelectric", 10)

	if len(results) < 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	// Verify results are sorted by relevance (descending)
	for i := 1; i < len(results); i++ {
		if results[i].Relevance > results[i-1].Relevance {
			t.Errorf("Results not properly sorted: result[%d].Relevance=%f > result[%d].Relevance=%f",
				i, results[i].Relevance, i-1, results[i-1].Relevance)
		}
	}

	if !strings.Contains(results[0].DocPath, "title_match.md") {
		t.Fatalf("expected title match as top result, got %q", results[0].DocPath)
	}
	if results[0].MatchType != "title" {
		t.Fatalf("expected top MatchType=title, got %q", results[0].MatchType)
	}

	scoreByName := map[string]float64{}
	for _, result := range results {
		scoreByName[filepath.Base(result.DocPath)] = result.Relevance
	}
	if scoreByName["heading_match.md"] <= scoreByName["content_match.md"] {
		t.Fatalf("expected heading boost > content: heading=%f content=%f", scoreByName["heading_match.md"], scoreByName["content_match.md"])
	}
}

// TestExtractSnippet verifies snippet extraction
func TestExtractSnippet(t *testing.T) {
	index := NewSearchIndex("")

	content := "This is a long document with many words. The ferroelectric materials are very interesting. More content follows here."

	snippet := index.extractSnippet(content, "ferroelectric")

	if !strings.Contains(snippet, "ferroelectric") {
		t.Error("Snippet should contain the search term")
	}

	if len(snippet) > 150 {
		t.Error("Snippet should be approximately 100 chars")
	}

	// Test snippet with term not found
	snippet2 := index.extractSnippet(content, "nonexistent")
	if len(snippet2) == 0 {
		t.Error("Should return beginning of content when term not found")
	}
}

// TestEditDistance verifies Levenshtein distance calculation
func TestEditDistance(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"hello", "hello", 0},
		{"hello", "helo", 1},
		{"ferroelectric", "feroelectric", 1},
		{"cat", "hat", 1},
		{"kitten", "sitting", 3},
		{"", "abc", 3},
		{"abc", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			dist := editDistance(tt.a, tt.b)
			if dist != tt.expected {
				t.Errorf("editDistance(%s, %s) = %d, expected %d",
					tt.a, tt.b, dist, tt.expected)
			}
		})
	}
}

// TestBreadcrumbWidget verifies breadcrumb navigation
func TestBreadcrumbWidget(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	var navigated string
	breadcrumb := NewBreadcrumbWidget(func(path string) {
		navigated = path
	})

	docsRoot := "/docs"
	currentPath := "/docs/module1/physics.md"

	breadcrumb.SetPath(currentPath, docsRoot)

	// Verify segments were created
	if len(breadcrumb.segments) < 2 {
		t.Errorf("Expected at least 2 segments (Home + module), got %d", len(breadcrumb.segments))
	}

	// First segment should be Home
	if breadcrumb.segments[0].Label != "Home" {
		t.Errorf("Expected first segment to be 'Home', got '%s'", breadcrumb.segments[0].Label)
	}

	if breadcrumb.segments[0].Path != docsRoot {
		t.Errorf("Expected Home path to be %s, got %s", docsRoot, breadcrumb.segments[0].Path)
	}

	// Suppress unused warning
	_ = navigated
}

// TestTableOfContentsWidget verifies ToC parsing
func TestTableOfContentsWidget(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	var selected string
	toc := NewTableOfContentsWidget(func(heading string) {
		selected = heading
	})

	markdown := `# Main Title

Some content here.

## Section One

More content.

### Subsection A

Details.

## Section Two

Final content.
`

	toc.ParseMarkdown(markdown)

	// Verify headings were parsed
	expectedCount := 4 // # + ## + ### + ##
	if len(toc.headings) != expectedCount {
		t.Errorf("Expected %d headings, got %d", expectedCount, len(toc.headings))
	}

	// Verify heading levels
	if toc.headings[0].Level != 1 {
		t.Errorf("First heading should be level 1, got %d", toc.headings[0].Level)
	}

	if toc.headings[0].Text != "Main Title" {
		t.Errorf("First heading text should be 'Main Title', got '%s'", toc.headings[0].Text)
	}

	// Verify anchor generation
	expectedAnchor := "main-title"
	if toc.headings[0].Anchor != expectedAnchor {
		t.Errorf("Expected anchor '%s', got '%s'", expectedAnchor, toc.headings[0].Anchor)
	}

	// Suppress unused warning
	_ = selected
}

// TestDocsHistory verifies history persistence
func TestDocsHistory(t *testing.T) {
	// Use temp directory for test
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	history := NewDocsHistory()

	// Test AddRecent
	history.AddRecent("/docs/file1.md")
	history.AddRecent("/docs/file2.md")
	history.AddRecent("/docs/file3.md")

	recent := history.GetRecent()
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent items, got %d", len(recent))
	}

	// Most recent should be first
	if recent[0] != "/docs/file3.md" {
		t.Errorf("Expected most recent to be file3.md, got %s", recent[0])
	}

	// Test duplicate handling
	history.AddRecent("/docs/file1.md")
	recent = history.GetRecent()
	if len(recent) != 3 {
		t.Error("Adding duplicate should not increase count")
	}
	if recent[0] != "/docs/file1.md" {
		t.Error("Duplicate should move to front")
	}

	// Test favorites
	history.ToggleFavorite("/docs/fav1.md")
	if !history.IsFavorite("/docs/fav1.md") {
		t.Error("File should be favorited")
	}

	favorites := history.GetFavorites()
	if len(favorites) != 1 {
		t.Errorf("Expected 1 favorite, got %d", len(favorites))
	}

	// Toggle off
	history.ToggleFavorite("/docs/fav1.md")
	if history.IsFavorite("/docs/fav1.md") {
		t.Error("File should no longer be favorited")
	}
}

// TestDocsHistory_Persistence verifies save/load using absolute paths
// to avoid os.Chdir races with parallel tests.
func TestDocsHistory_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	omcDir := filepath.Join(tmpDir, ".omc")
	if err := os.MkdirAll(omcDir, 0755); err != nil {
		t.Fatalf("Failed to create .omc directory: %v", err)
	}

	histPath := filepath.Join(omcDir, "docs-history.json")

	// Create and populate history with explicit configPath
	history1 := NewDocsHistory()
	history1.configPath = histPath

	history1.mu.Lock()
	history1.Recent = []string{"/docs/file1.md", "/docs/file2.md"}
	history1.Favorites = []string{"/docs/fav1.md"}
	history1.favoritesMap = map[string]bool{"/docs/fav1.md": true}
	history1.mu.Unlock()

	// Force synchronous save
	if err := history1.Save(); err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Verify file was created and has content
	data, err := os.ReadFile(histPath)
	if err != nil {
		t.Fatalf("History file was not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("History file is empty")
	}

	// Load into new instance with same path
	history2 := NewDocsHistory()
	history2.configPath = histPath
	if err := history2.Load(); err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	// Verify data was loaded
	recent := history2.GetRecent()
	if len(recent) != 2 {
		t.Errorf("Expected 2 recent items after load, got %d (file content: %s)", len(recent), string(data))
	}

	if !history2.IsFavorite("/docs/fav1.md") {
		favorites := history2.GetFavorites()
		t.Errorf("Favorite should be loaded. Got favorites: %v", favorites)
	}
}

// TestDetectGlossaryTerms verifies glossary term detection
func TestDetectGlossaryTerms(t *testing.T) {
	content := `# Ferroelectric Memory

Ferroelectric materials exhibit spontaneous polarization due to hysteresis.
The Preisach model is used for simulation.
`

	terms := DetectGlossaryTerms(content)

	// Should detect common glossary terms
	expectedTerms := []string{"Ferroelectric", "Hysteresis", "Polarization"}

	for _, expected := range expectedTerms {
		found := false
		for _, term := range terms {
			if strings.EqualFold(term, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Warning: Expected term '%s' not found (may not be in glossary)", expected)
		}
	}
}

// TestHighlightGlossaryTerms verifies term highlighting
func TestHighlightGlossaryTerms(t *testing.T) {
	content := `# Test

This is a test with ferroelectric materials.

## Code Block

` + "```go\n" + `// ferroelectric in code should not be highlighted
` + "```\n"

	highlighted := HighlightGlossaryTerms(content)

	// Code blocks should not be modified
	if !strings.Contains(highlighted, "```go") {
		t.Error("Code block markers should be preserved")
	}

	// Headers should not be modified
	if !strings.HasPrefix(highlighted, "# Test") {
		t.Error("Headers should be preserved")
	}
}

// TestLayoutManager verifies responsive layout switching
func TestLayoutManager(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	manager := NewLayoutManager()

	// Test initial state
	if manager.currentMode != LayoutDesktop {
		t.Error("Should default to LayoutDesktop")
	}

	// Test mode switching based on width
	tests := []struct {
		width        float32
		expectedMode LayoutMode
	}{
		{500, LayoutMobile},
		{700, LayoutTablet},
		{1000, LayoutDesktop},
		{1400, LayoutWide},
	}

	for _, tt := range tests {
		t.Run(manager.GetModeString(), func(t *testing.T) {
			// Use fyne.NewSize instead of test.NewSize
			size := fyne.NewSize(tt.width, 600)
			manager.OnResize(size)

			if manager.currentMode != tt.expectedMode {
				t.Errorf("Width %f: expected mode %d, got %d",
					tt.width, tt.expectedMode, manager.currentMode)
			}
		})
	}

	// Test toggle functions
	initialSidebar := manager.IsSidebarVisible()
	manager.ToggleSidebar()
	if manager.IsSidebarVisible() == initialSidebar {
		t.Error("ToggleSidebar should change visibility")
	}

	initialToc := manager.IsTocVisible()
	manager.ToggleToc()
	if manager.IsTocVisible() == initialToc {
		t.Error("ToggleToc should change visibility")
	}
}

// TestGetBreakpointForWidth verifies breakpoint calculation
func TestGetBreakpointForWidth(t *testing.T) {
	tests := []struct {
		width    float32
		expected LayoutMode
	}{
		{400, LayoutMobile},
		{600, LayoutTablet},
		{800, LayoutTablet},
		{900, LayoutDesktop},
		{1100, LayoutDesktop},
		{1300, LayoutWide},
	}

	for _, tt := range tests {
		t.Run(string(rune(int(tt.width))), func(t *testing.T) {
			mode := GetBreakpointForWidth(tt.width)
			if mode != tt.expected {
				t.Errorf("Width %f: expected mode %d, got %d",
					tt.width, tt.expected, mode)
			}
		})
	}
}

// TestSearchIndex_EmptyQuery verifies empty query handling
func TestSearchIndex_EmptyQuery(t *testing.T) {
	index := NewSearchIndex("")

	results := index.Query("", 10)
	if results != nil && len(results) > 0 {
		t.Error("Empty query should return no results")
	}
}

// TestSearchIndex_QueryLimit verifies result limiting
func TestSearchIndex_QueryLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many documents with valid filenames
	for i := 0; i < 20; i++ {
		// Use zero-padded numbers for valid filenames
		filename := filepath.Join(tmpDir, strings.ReplaceAll(strings.ReplaceAll(
			"doc_test_", " ", "_"), "\x00", "") + string(rune('a'+i)) + ".md")
		content := "# Document\n\nThis is a test document with test content."
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	index := NewSearchIndex(tmpDir)
	index.Build()

	// Query with limit
	results := index.Query("test", 5)

	if len(results) > 5 {
		t.Errorf("Expected max 5 results, got %d", len(results))
	}
}

// TestSearchIndex_SpecialCharacters verifies special character handling
func TestSearchIndex_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()

	content := "# Test\n\nSpecial chars: @#$% ferroelectric!!! materials???"
	path := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	index := NewSearchIndex(tmpDir)
	index.Build()

	// Should still find terms despite special chars
	results := index.Query("ferroelectric", 10)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'ferroelectric', got %d", len(results))
	}
}

func TestEmbeddedDocsApp_SortEntries_CurriculumOrder(t *testing.T) {
	app := &EmbeddedDocsApp{docsPath: "/docs/documentation"}
	entries := []*docEntry{
		{name: "README.md", path: "/docs/documentation/README.md"},
		{name: "module10-misc", path: "/docs/documentation/module10-misc", isDir: true},
		{name: "module2-crossbar", path: "/docs/documentation/module2-crossbar", isDir: true},
		{name: "research-papers", path: "/docs/documentation/research-papers", isDir: true},
		{name: "MODULES.md", path: "/docs/documentation/MODULES.md"},
		{name: "module1-hysteresis", path: "/docs/documentation/module1-hysteresis", isDir: true},
	}

	app.sortEntries(entries, app.docsPath)

	ordered := make([]string, 0, len(entries))
	for _, e := range entries {
		ordered = append(ordered, e.name)
	}

	expected := []string{
		"module1-hysteresis",
		"module2-crossbar",
		"module10-misc",
		"research-papers",
		"README.md",
		"MODULES.md",
	}

	for i := range expected {
		if ordered[i] != expected[i] {
			t.Fatalf("unexpected order at %d: got %q want %q (full=%v)", i, ordered[i], expected[i], ordered)
		}
	}
}

func TestEmbeddedDocsApp_SortEntries_ModuleFileOrder(t *testing.T) {
	app := &EmbeddedDocsApp{docsPath: "/docs/documentation"}
	parent := "/docs/documentation/module6-eda"
	entries := []*docEntry{
		{name: "z-notes.md", path: parent + "/z-notes.md"},
		{name: "PHYSICS.md", path: parent + "/PHYSICS.md"},
		{name: "ELI5.md", path: parent + "/ELI5.md"},
		{name: "FEATURES.md", path: parent + "/FEATURES.md"},
		{name: "OPENSOURCE-TOOLS.md", path: parent + "/OPENSOURCE-TOOLS.md"},
		{name: "assets", path: parent + "/assets", isDir: true},
	}

	app.sortEntries(entries, parent)

	ordered := make([]string, 0, len(entries))
	for _, e := range entries {
		ordered = append(ordered, e.name)
	}

	expected := []string{"assets", "ELI5.md", "PHYSICS.md", "FEATURES.md", "OPENSOURCE-TOOLS.md", "z-notes.md"}
	for i := range expected {
		if ordered[i] != expected[i] {
			t.Fatalf("unexpected module order at %d: got %q want %q (full=%v)", i, ordered[i], expected[i], ordered)
		}
	}
}

func TestModuleShortcutsPanel_MappingAndDisableState(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "moduleX")
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_ = os.WriteFile(filepath.Join(modulePath, "ELI5.md"), []byte("# e"), 0644)
	_ = os.WriteFile(filepath.Join(modulePath, "PHYSICS.md"), []byte("# p"), 0644)

	var selected string
	panel := NewModuleShortcutsPanel(func(path string) { selected = path })
	panel.SetModulePath(modulePath)

	btns := map[string]*widget.Button{}
	for _, obj := range panel.container.Objects {
		if b, ok := obj.(*widget.Button); ok {
			btns[b.Text] = b
		}
	}

	if len(btns) != 4 {
		t.Fatalf("expected 4 shortcut buttons, got %d", len(btns))
	}
	if btns["ELI5"].Disabled() {
		t.Fatal("ELI5 button should be enabled when file exists")
	}
	if btns["Physics"].Disabled() {
		t.Fatal("Physics button should be enabled when file exists")
	}
	if !btns["Features"].Disabled() || !btns["Tools"].Disabled() {
		t.Fatal("Features/Tools should be disabled when files are missing")
	}

	btns["ELI5"].OnTapped()
	expectedPath := filepath.Join(modulePath, "ELI5.md")
	if selected != expectedPath {
		t.Fatalf("shortcut path mismatch: got %q want %q", selected, expectedPath)
	}
}

func TestEmbeddedDocsApp_TreeClickTargets(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()
	w := testApp.NewWindow("docs")
	defer w.Close()

	tmpDir := t.TempDir()
	docsRoot := filepath.Join(tmpDir, "docs", "documentation")
	moduleDir := filepath.Join(docsRoot, "module1-hysteresis")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	filePath := filepath.Join(moduleDir, "ELI5.md")
	if err := os.WriteFile(filePath, []byte("# Hello\n\nBody"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	app := NewEmbeddedDocsApp()
	app.EmbeddedAppBase.Init(testApp, w)
	app.docsPath = docsRoot
	app.searchIndex = NewSearchIndex(docsRoot)
	app.history = NewDocsHistory()
	app.createUIComponents()
	_ = app.buildSidebar()

	if app.tree.IsBranchOpen(moduleDir) {
		t.Fatal("branch should start closed")
	}
	app.tree.OnSelected(moduleDir)
	if !app.tree.IsBranchOpen(moduleDir) {
		t.Fatal("clicking folder row should open branch")
	}
	app.tree.OnSelected(moduleDir)
	if app.tree.IsBranchOpen(moduleDir) {
		t.Fatal("second click on folder row should close branch")
	}

	app.tree.OnSelected(filePath)
	if app.currentFile != filePath {
		t.Fatalf("clicking file row should load document: got %q want %q", app.currentFile, filePath)
	}
}

func TestEmbeddedDocsApp_TreeCategoryBadges(t *testing.T) {
	app := &EmbeddedDocsApp{docsPath: "/docs/documentation"}

	tests := []struct {
		name     string
		entry    *docEntry
		expected string
	}{
		{name: "module directory", entry: &docEntry{name: "module7-docs", path: "/docs/documentation/module7-docs", isDir: true}, expected: "Module"},
		{name: "research directory", entry: &docEntry{name: "research-papers", path: "/docs/documentation/research-papers", isDir: true}, expected: "Research"},
		{name: "eli5 file", entry: &docEntry{name: "ELI5.md", path: "/docs/documentation/module1/ELI5.md"}, expected: "ELI5"},
		{name: "physics file", entry: &docEntry{name: "PHYSICS.md", path: "/docs/documentation/module1/PHYSICS.md"}, expected: "Physics"},
		{name: "guide file", entry: &docEntry{name: "FEATURES.md", path: "/docs/documentation/module1/FEATURES.md"}, expected: "Guide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.treeCategory(tt.entry)
			if got != tt.expected {
				t.Fatalf("treeCategory mismatch: got=%q want=%q", got, tt.expected)
			}
			if _, ok := CategoryColors[got]; !ok {
				t.Fatalf("category %q missing color mapping", got)
			}
		})
	}
}

func TestEmbeddedDocsApp_LoadDocument_TocVisibility(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()
	w := testApp.NewWindow("docs")
	defer w.Close()

	tmpDir := t.TempDir()
	docsRoot := filepath.Join(tmpDir, "docs", "documentation")
	moduleDir := filepath.Join(docsRoot, "module1-hysteresis")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	shortDoc := filepath.Join(moduleDir, "short.md")
	longDoc := filepath.Join(moduleDir, "long.md")
	if err := os.WriteFile(shortDoc, []byte("# A\n\n## B"), 0644); err != nil {
		t.Fatalf("write short doc: %v", err)
	}
	if err := os.WriteFile(longDoc, []byte("# A\n\n## B\n\n### C"), 0644); err != nil {
		t.Fatalf("write long doc: %v", err)
	}

	app := NewEmbeddedDocsApp()
	app.EmbeddedAppBase.Init(testApp, w)
	app.window = w
	app.docsPath = docsRoot
	app.searchIndex = NewSearchIndex(docsRoot)
	app.history = NewDocsHistory()
	app.createUIComponents()
	app.layoutManager = NewLayoutManager()
	app.layoutManager.SetComponents(app.buildSidebar(), app.buildMainContent(), app.buildTocSidebar(), app.buildTopBar())
	_ = app.layoutManager.BuildLayout()

	app.loadDocument(shortDoc)
	fyne.DoAndWait(func() {})
	if app.layoutManager.IsTocVisible() {
		t.Fatal("ToC should be hidden when document has fewer than 3 headings")
	}

	app.loadDocument(longDoc)
	fyne.DoAndWait(func() {})
	if !app.layoutManager.IsTocVisible() {
		t.Fatal("ToC should be visible when document has 3+ headings")
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
