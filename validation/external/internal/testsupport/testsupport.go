package testsupport

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// HasCommand reports whether an external executable is available.
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RequireCommand skips the calling test when an external executable is absent.
func RequireCommand(t *testing.T, name, skipMessage string) {
	t.Helper()
	if !HasCommand(name) {
		t.Skip(skipMessage)
	}
}

// RequirePythonModule skips the calling test when python3 or the named module is absent.
func RequirePythonModule(t *testing.T, module, skipMessage string) {
	t.Helper()
	RequireCommand(t, "python3", "python3 not installed")
	if err := exec.Command("python3", "-c", "import "+module).Run(); err != nil {
		t.Skip(skipMessage)
	}
}

// ProjectRoot returns the repository root from validation/external/internal/testsupport.
func ProjectRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate testsupport source file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))
}

// ExternalArtifactDir returns and creates the shared external-validation artifact directory.
func ExternalArtifactDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(ProjectRoot(t), "output", "validation", "external")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create external artifact directory: %v", err)
	}
	return dir
}

// WriteExternalArtifact writes an indented JSON artifact into output/validation/external.
func WriteExternalArtifact(t *testing.T, name string, artifact map[string]interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s artifact: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(ExternalArtifactDir(t), name), b, 0o644); err != nil {
		t.Fatalf("write %s artifact: %v", name, err)
	}
}
