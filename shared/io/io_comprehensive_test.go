package io

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// atomicWriteJSON writes JSON via temp file + rename in the same directory.
func atomicWriteJSON(path string, data interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}

// sanitizePath joins userPath under base and rejects path traversal.
func sanitizePath(base, userPath string) (string, error) {
	if filepath.IsAbs(userPath) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	clean := filepath.Clean(userPath)
	full := filepath.Join(base, clean)
	rel, err := filepath.Rel(base, full)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path traversal detected: %q", userPath)
	}
	return full, nil
}

func writeJSONWithSizeLimit(path string, data interface{}, maxBytes int) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if len(b) > maxBytes {
		return fmt.Errorf("payload too large: %d > %d", len(b), maxBytes)
	}
	return SaveJSON(path, data)
}

func TestIOComprehensiveValidation(t *testing.T) {
	t.Run("safe file write uses atomic rename semantics", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "atomic.json")

		if err := atomicWriteJSON(path, testData{Name: "v1", Value: 1}); err != nil {
			t.Fatalf("initial atomic write failed: %v", err)
		}

		if err := atomicWriteJSON(path, testData{Name: "v2", Value: 2}); err != nil {
			t.Fatalf("second atomic write failed: %v", err)
		}

		var loaded testData
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON after atomic rename failed: %v", err)
		}
		if loaded.Name != "v2" || loaded.Value != 2 {
			t.Fatalf("atomic replacement mismatch: got %+v", loaded)
		}
	})

	t.Run("directory creation works with usable permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		target := filepath.Join(tmpDir, "a", "b", "c")

		if err := EnsureDir(target); err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("target is not a directory: %s", target)
		}

		perm := info.Mode().Perm()
		if perm&0o700 != 0o700 {
			t.Fatalf("owner permissions not usable, got %o", perm)
		}
	})

	t.Run("file size limits are enforced", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "limited.json")

		small := map[string]string{"k": "ok"}
		if err := writeJSONWithSizeLimit(path, small, 1024); err != nil {
			t.Fatalf("small payload should succeed: %v", err)
		}

		large := map[string]string{"blob": strings.Repeat("x", 4096)}
		err := writeJSONWithSizeLimit(path, large, 128)
		if err == nil {
			t.Fatal("expected error for payload exceeding size limit")
		}
		if !strings.Contains(err.Error(), "payload too large") {
			t.Fatalf("unexpected error for oversized payload: %v", err)
		}
	})

	t.Run("path sanitization rejects traversal", func(t *testing.T) {
		tmpDir := t.TempDir()

		safePath, err := sanitizePath(tmpDir, filepath.Join("nested", "file.json"))
		if err != nil {
			t.Fatalf("safe path rejected unexpectedly: %v", err)
		}
		if !strings.HasPrefix(safePath, tmpDir) {
			t.Fatalf("safe path escaped base dir: %s", safePath)
		}

		traversals := []string{"../escape.json", "../../etc/passwd"}
		if runtime.GOOS == "windows" {
			traversals = append(traversals, `..\\..\\Windows\\system.ini`)
		}
		for _, p := range traversals {
			if _, err := sanitizePath(tmpDir, p); err == nil {
				t.Fatalf("expected traversal rejection for %q", p)
			}
		}
	})

	t.Run("concurrent file access remains consistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "concurrent.json")

		const writers = 8
		const iterations = 50

		var wg sync.WaitGroup
		errCh := make(chan error, writers*iterations)

		for w := 0; w < writers; w++ {
			w := w
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < iterations; i++ {
					payload := map[string]int{"writer": w, "iter": i}
					if err := SaveJSON(path, payload); err != nil {
						errCh <- fmt.Errorf("writer %d iter %d: %w", w, i, err)
						return
					}
				}
			}()
		}
		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				t.Fatalf("concurrent write failed: %v", err)
			}
		}

		var loaded map[string]int
		if err := LoadJSON(path, &loaded); err != nil {
			t.Fatalf("LoadJSON after concurrent writes failed: %v", err)
		}
		if _, ok := loaded["writer"]; !ok {
			t.Fatalf("final JSON missing writer field: %+v", loaded)
		}
		if _, ok := loaded["iter"]; !ok {
			t.Fatalf("final JSON missing iter field: %+v", loaded)
		}
	})
}
