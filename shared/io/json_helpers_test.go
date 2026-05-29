package io

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testData represents a sample struct for JSON testing.
type testData struct {
	Name    string  `json:"name"`
	Value   int     `json:"value"`
	Enabled bool    `json:"enabled"`
	Score   float64 `json:"score"`
}

// nestedTestData represents nested struct for JSON testing.
type nestedTestData struct {
	ID       string            `json:"id"`
	Items    []testData        `json:"items"`
	Metadata map[string]string `json:"metadata"`
}

func TestSaveJSON(t *testing.T) {
	t.Run("saves simple struct", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.json")

		data := testData{Name: "test", Value: 42, Enabled: true, Score: 3.14}
		err := SaveJSON(path, data)
		if err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		// Verify file exists
		if !FileExists(path) {
			t.Fatal("File was not created")
		}

		// Verify content is pretty-printed
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		// Check for indentation (pretty print)
		if len(content) == 0 {
			t.Fatal("File is empty")
		}

		// Verify it's valid JSON and can be parsed back
		var loaded testData
		if err := json.Unmarshal(content, &loaded); err != nil {
			t.Fatalf("Written content is not valid JSON: %v", err)
		}

		if loaded.Name != data.Name || loaded.Value != data.Value {
			t.Errorf("Data mismatch: got %+v, want %+v", loaded, data)
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nested", "deep", "test.json")

		data := testData{Name: "nested"}
		err := SaveJSON(path, data)
		if err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		if !FileExists(path) {
			t.Fatal("File was not created in nested directory")
		}
	})

	t.Run("saves nested struct", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nested.json")

		data := nestedTestData{
			ID: "test-id",
			Items: []testData{
				{Name: "item1", Value: 1},
				{Name: "item2", Value: 2},
			},
			Metadata: map[string]string{"key": "value"},
		}

		err := SaveJSON(path, data)
		if err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded nestedTestData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded.ID != data.ID || len(loaded.Items) != len(data.Items) {
			t.Errorf("Nested data mismatch: got %+v, want %+v", loaded, data)
		}
	})

	t.Run("returns error for unmarshalable data", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.json")

		// Channels cannot be marshaled to JSON
		data := make(chan int)
		err := SaveJSON(path, data)
		if err == nil {
			t.Fatal("Expected error for unmarshalable data")
		}
	})
}

func TestLoadJSON(t *testing.T) {
	t.Run("loads valid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.json")

		original := testData{Name: "test", Value: 42, Enabled: true, Score: 3.14}
		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		var loaded testData
		err := LoadJSON(path, &loaded)
		if err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded.Name != original.Name {
			t.Errorf("Name mismatch: got %s, want %s", loaded.Name, original.Name)
		}
		if loaded.Value != original.Value {
			t.Errorf("Value mismatch: got %d, want %d", loaded.Value, original.Value)
		}
		if loaded.Enabled != original.Enabled {
			t.Errorf("Enabled mismatch: got %v, want %v", loaded.Enabled, original.Enabled)
		}
		if loaded.Score != original.Score {
			t.Errorf("Score mismatch: got %f, want %f", loaded.Score, original.Score)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		var data testData
		err := LoadJSON("/non/existent/path.json", &data)
		if err == nil {
			t.Fatal("Expected error for non-existent file")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "invalid.json")

		// Write invalid JSON
		if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		var data testData
		err := LoadJSON(path, &data)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
	})

	t.Run("loads into map", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "map.json")

		original := map[string]interface{}{
			"key1": "value1",
			"key2": float64(42),
		}
		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		var loaded map[string]interface{}
		err := LoadJSON(path, &loaded)
		if err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded["key1"] != original["key1"] {
			t.Errorf("key1 mismatch: got %v, want %v", loaded["key1"], original["key1"])
		}
	})
}

func TestSaveJSONCompact(t *testing.T) {
	t.Run("saves without formatting", func(t *testing.T) {
		tmpDir := t.TempDir()
		pathCompact := filepath.Join(tmpDir, "compact.json")
		pathPretty := filepath.Join(tmpDir, "pretty.json")

		data := testData{Name: "test", Value: 42, Enabled: true}

		if err := SaveJSONCompact(pathCompact, data); err != nil {
			t.Fatalf("SaveJSONCompact failed: %v", err)
		}
		if err := SaveJSON(pathPretty, data); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		compactContent, _ := os.ReadFile(pathCompact)
		prettyContent, _ := os.ReadFile(pathPretty)

		// Compact should be smaller (no whitespace/indentation)
		if len(compactContent) >= len(prettyContent) {
			t.Errorf("Compact (%d bytes) should be smaller than pretty (%d bytes)",
				len(compactContent), len(prettyContent))
		}

		// Both should parse to same data
		var loadedCompact, loadedPretty testData
		json.Unmarshal(compactContent, &loadedCompact)
		json.Unmarshal(prettyContent, &loadedPretty)

		if loadedCompact != loadedPretty {
			t.Errorf("Loaded data differs: compact=%+v, pretty=%+v", loadedCompact, loadedPretty)
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "a", "b", "c", "test.json")

		err := SaveJSONCompact(path, testData{Name: "test"})
		if err != nil {
			t.Fatalf("SaveJSONCompact failed: %v", err)
		}

		if !FileExists(path) {
			t.Fatal("File was not created")
		}
	})

	t.Run("returns error for unmarshalable data", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.json")

		data := make(chan int)
		err := SaveJSONCompact(path, data)
		if err == nil {
			t.Fatal("Expected error for unmarshalable data")
		}
	})
}

func TestFileExists(t *testing.T) {
	t.Run("returns true for existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "exists.txt")

		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if !FileExists(path) {
			t.Error("FileExists should return true for existing file")
		}
	})

	t.Run("returns false for non-existent file", func(t *testing.T) {
		if FileExists("/non/existent/file.txt") {
			t.Error("FileExists should return false for non-existent file")
		}
	})

	t.Run("returns false for directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		if FileExists(tmpDir) {
			t.Error("FileExists should return false for directory")
		}
	})

	t.Run("returns false for empty path", func(t *testing.T) {
		if FileExists("") {
			t.Error("FileExists should return false for empty path")
		}
	})
}

func TestDirExists(t *testing.T) {
	t.Run("returns true for existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		if !DirExists(tmpDir) {
			t.Error("DirExists should return true for existing directory")
		}
	})

	t.Run("returns false for non-existent directory", func(t *testing.T) {
		if DirExists("/non/existent/directory") {
			t.Error("DirExists should return false for non-existent directory")
		}
	})

	t.Run("returns false for file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "file.txt")

		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		if DirExists(path) {
			t.Error("DirExists should return false for file")
		}
	})

	t.Run("returns false for empty path", func(t *testing.T) {
		if DirExists("") {
			t.Error("DirExists should return false for empty path")
		}
	})
}

func TestEnsureDir(t *testing.T) {
	t.Run("creates new directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "newdir")

		err := EnsureDir(path)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		if !DirExists(path) {
			t.Error("Directory was not created")
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "a", "b", "c", "d")

		err := EnsureDir(path)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		if !DirExists(path) {
			t.Error("Nested directories were not created")
		}
	})

	t.Run("succeeds for existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Should not error if directory already exists
		err := EnsureDir(tmpDir)
		if err != nil {
			t.Fatalf("EnsureDir failed for existing directory: %v", err)
		}
	})

	t.Run("preserves existing directory contents", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "existing.txt")

		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		err := EnsureDir(tmpDir)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		// File should still exist
		if !FileExists(filePath) {
			t.Error("EnsureDir should not delete existing files")
		}
	})
}

func TestRoundTrip(t *testing.T) {
	t.Run("complex data survives save/load cycle", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "roundtrip.json")

		original := nestedTestData{
			ID: "complex-test",
			Items: []testData{
				{Name: "first", Value: 1, Enabled: true, Score: 1.1},
				{Name: "second", Value: 2, Enabled: false, Score: 2.2},
				{Name: "third", Value: 3, Enabled: true, Score: 3.3},
			},
			Metadata: map[string]string{
				"version": "1.0",
				"author":  "test",
			},
		}

		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded nestedTestData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded.ID != original.ID {
			t.Errorf("ID mismatch: got %s, want %s", loaded.ID, original.ID)
		}
		if len(loaded.Items) != len(original.Items) {
			t.Errorf("Items length mismatch: got %d, want %d", len(loaded.Items), len(original.Items))
		}
		for i, item := range loaded.Items {
			if item != original.Items[i] {
				t.Errorf("Item %d mismatch: got %+v, want %+v", i, item, original.Items[i])
			}
		}
		if loaded.Metadata["version"] != original.Metadata["version"] {
			t.Errorf("Metadata mismatch")
		}
	})

	t.Run("compact also survives round trip", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "compact_roundtrip.json")

		original := testData{Name: "compact", Value: 999, Enabled: true, Score: 99.99}

		if err := SaveJSONCompact(path, original); err != nil {
			t.Fatalf("SaveJSONCompact failed: %v", err)
		}

		var loaded testData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded != original {
			t.Errorf("Data mismatch: got %+v, want %+v", loaded, original)
		}
	})
}

func TestBufferOrWriteJSONArtifactBuffersWhenPathMissing(t *testing.T) {
	artifact, err := BufferOrWriteJSONArtifact("", map[string]int{"cells": 64})
	if err != nil {
		t.Fatalf("BufferOrWriteJSONArtifact returned error: %v", err)
	}
	if artifact.StatusVerb != "buffered" || artifact.Path != "artifact buffer" || artifact.WroteFile {
		t.Fatalf("artifact = %+v, want buffered artifact", artifact)
	}
	if !strings.Contains(artifact.Content, "\"cells\": 64") || artifact.Bytes != len(artifact.Content) {
		t.Fatalf("artifact content/bytes = %d %q", artifact.Bytes, artifact.Content)
	}
}

func TestBufferOrWriteJSONArtifactWritesValidatedPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "artifact.json")
	artifact, err := BufferOrWriteJSONArtifact(path, map[string]int{"cells": 64})
	if err != nil {
		t.Fatalf("BufferOrWriteJSONArtifact returned error: %v", err)
	}
	if artifact.StatusVerb != "wrote" || artifact.Path != filepath.Clean(path) || !artifact.WroteFile {
		t.Fatalf("artifact = %+v, want wrote clean path", artifact)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if string(raw) != artifact.Content {
		t.Fatalf("file content differs from artifact content")
	}
}

func TestBufferOrWriteJSONArtifactRejectsTraversalPath(t *testing.T) {
	if _, err := BufferOrWriteJSONArtifact("../escape.json", map[string]int{"cells": 64}); err == nil {
		t.Fatal("expected traversal path to fail")
	}
}

func TestBufferOrWriteTextArtifactBuffersWhenPathMissing(t *testing.T) {
	artifact, err := BufferOrWriteTextArtifact("", "hello")
	if err != nil {
		t.Fatalf("BufferOrWriteTextArtifact returned error: %v", err)
	}
	if artifact.StatusVerb != "buffered" || artifact.Path != "artifact buffer" || artifact.WroteFile {
		t.Fatalf("artifact = %+v, want buffered artifact", artifact)
	}
}

func TestBufferOrWriteTextArtifactWritesValidatedPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "artifact.txt")
	artifact, err := BufferOrWriteTextArtifact(path, "hello")
	if err != nil {
		t.Fatalf("BufferOrWriteTextArtifact returned error: %v", err)
	}
	if artifact.StatusVerb != "wrote" || artifact.Path != filepath.Clean(path) || !artifact.WroteFile {
		t.Fatalf("artifact = %+v, want wrote clean path", artifact)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if string(raw) != "hello" {
		t.Fatalf("artifact contents = %q", raw)
	}
}

func TestBufferOrWriteTextArtifactRejectsTraversalPath(t *testing.T) {
	if _, err := BufferOrWriteTextArtifact("../escape.txt", "hello"); err == nil {
		t.Fatal("expected traversal path to fail")
	}
}

func TestValidatePath(t *testing.T) {
	t.Run("rejects empty path", func(t *testing.T) {
		_, err := ValidatePath("")
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})

	t.Run("rejects whitespace-only path", func(t *testing.T) {
		_, err := ValidatePath("   ")
		if err == nil {
			t.Fatal("expected error for whitespace-only path")
		}
	})

	t.Run("rejects relative traversal", func(t *testing.T) {
		traversals := []string{
			"../escape.json",
			"../../etc/passwd",
			"../../../tmp/evil",
		}
		for _, p := range traversals {
			_, err := ValidatePath(p)
			if err == nil {
				t.Fatalf("expected traversal rejection for %q", p)
			}
		}
	})

	t.Run("accepts valid relative path", func(t *testing.T) {
		got, err := ValidatePath("data/output.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != filepath.Clean("data/output.json") {
			t.Fatalf("unexpected cleaned path: %s", got)
		}
	})

	t.Run("accepts absolute path", func(t *testing.T) {
		got, err := ValidatePath("/tmp/test.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/tmp/test.json" {
			t.Fatalf("unexpected cleaned path: %s", got)
		}
	})

	t.Run("accepts nested relative without traversal", func(t *testing.T) {
		got, err := ValidatePath("exports/comparison/data.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != filepath.Clean("exports/comparison/data.json") {
			t.Fatalf("unexpected cleaned path: %s", got)
		}
	})
}

func TestSaveJSON_PathValidation(t *testing.T) {
	t.Run("rejects empty path", func(t *testing.T) {
		err := SaveJSON("", map[string]string{"k": "v"})
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})
}

func TestLoadJSON_PathValidation(t *testing.T) {
	t.Run("rejects empty path", func(t *testing.T) {
		var data map[string]string
		err := LoadJSON("", &data)
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})
}

func TestSaveJSONCompact_PathValidation(t *testing.T) {
	t.Run("rejects empty path", func(t *testing.T) {
		err := SaveJSONCompact("", map[string]string{"k": "v"})
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("handles empty struct", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty.json")

		original := testData{}
		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded testData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded != original {
			t.Errorf("Empty struct mismatch: got %+v, want %+v", loaded, original)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty_slice.json")

		original := []testData{}
		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded []testData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if len(loaded) != 0 {
			t.Errorf("Expected empty slice, got %d elements", len(loaded))
		}
	})

	t.Run("handles nil pointer in target", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nil.json")

		if err := SaveJSON(path, testData{Name: "test"}); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded *testData
		// This should work - LoadJSON creates the object
		err := LoadJSON(path, &loaded)
		if err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}
	})

	t.Run("handles special characters in string", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "special.json")

		original := testData{Name: "test\nwith\ttabs\"and\"quotes"}
		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded testData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded.Name != original.Name {
			t.Errorf("Special chars mismatch: got %q, want %q", loaded.Name, original.Name)
		}
	})

	t.Run("handles unicode", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "unicode.json")

		original := testData{Name: "日本語テスト 🎉 émojis"}
		if err := SaveJSON(path, original); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		var loaded testData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON failed: %v", err)
		}

		if loaded.Name != original.Name {
			t.Errorf("Unicode mismatch: got %q, want %q", loaded.Name, original.Name)
		}
	})
}
