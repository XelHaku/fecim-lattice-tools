package progress

import (
	"os/exec"
	"strings"
	"testing"
)

func TestDefaultProgressPackageDoesNotImportFyne(t *testing.T) {
	cmd := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list shared/progress imports failed: %v\n%s", err, out)
	}
	for _, imp := range strings.Fields(string(out)) {
		if strings.HasPrefix(imp, "fyne.io/fyne") {
			t.Fatalf("shared/progress default build must not import Fyne; found %s", imp)
		}
	}
}
