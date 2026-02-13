package export

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReproducibilityPackInput defines required bundle content.
type ReproducibilityPackInput struct {
	ConfigYAML      []byte
	RandomSeeds     map[string]int64
	GitCommitHash   string
	TestResults     string
	GeneratedAssets []string
}

// CreateReproducibilityPack exports a reproducibility pack to a directory or zip.
func CreateReproducibilityPack(outputPath string, in ReproducibilityPackInput) (string, error) {
	if strings.HasSuffix(strings.ToLower(outputPath), ".zip") {
		return outputPath, createPackZip(outputPath, in)
	}
	return outputPath, createPackDir(outputPath, in)
}

func createPackDir(dir string, in ReproducibilityPackInput) error {
	if err := os.MkdirAll(filepath.Join(dir, "artifacts"), 0o755); err != nil {
		return err
	}
	if err := writePackFiles(dir, in); err != nil {
		return err
	}
	return copyArtifacts(filepath.Join(dir, "artifacts"), in.GeneratedAssets)
}

func createPackZip(zipPath string, in ReproducibilityPackInput) error {
	tmp, err := os.MkdirTemp("", "fecim-repro-pack-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := createPackDir(tmp, in); err != nil {
		return err
	}

	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()
	zw := zip.NewWriter(zf)
	defer zw.Close()

	return filepath.Walk(tmp, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(tmp, path)
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}

func writePackFiles(dir string, in ReproducibilityPackInput) error {
	if len(in.ConfigYAML) > 0 {
		if err := os.WriteFile(filepath.Join(dir, "config.yaml"), in.ConfigYAML, 0o644); err != nil {
			return err
		}
	}
	seeds, err := json.MarshalIndent(in.RandomSeeds, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "random_seeds.json"), seeds, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "git_commit.txt"), []byte(strings.TrimSpace(in.GitCommitHash)+"\n"), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "test_results.txt"), []byte(in.TestResults), 0o644)
}

func copyArtifacts(dest string, assets []string) error {
	for _, a := range assets {
		in, err := os.Open(a)
		if err != nil {
			return err
		}
		defer in.Close()
		outPath := filepath.Join(dest, filepath.Base(a))
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if _, err = io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		if err = out.Close(); err != nil {
			return err
		}
	}
	return nil
}

func ValidateReproducibilityPack(dir string) error {
	required := []string{"config.yaml", "random_seeds.json", "git_commit.txt", "test_results.txt", "artifacts"}
	for _, f := range required {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return fmt.Errorf("missing %s: %w", f, err)
		}
	}
	return nil
}
