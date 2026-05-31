package toolchain_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDriftDetectorAcceptsPinFileOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell scripts are validated on Unix-like environments")
	}

	tmp := t.TempDir()
	pinFile := filepath.Join(tmp, "README.md")
	if err := os.WriteFile(pinFile, []byte(`| Tool | Version Pin | License | Install Method | FeCIM Use-Case |
|---|---:|---|---|---|
| ngspice | `+"`42`"+` | BSD-3-Clause | distro pkg | SPICE |
| Icarus Verilog (`+"`iverilog`"+`) | `+"`12.0`"+` | GPL | distro pkg | RTL |
| Verilator | `+"`5.028`"+` | LGPL | distro pkg | lint |
| Python scientific stack (`+"`numpy`"+`, `+"`scipy`"+`) | `+"`numpy==2.1.3, scipy==1.14.1`"+` | BSD | pip | scripts |
| Go toolchain | `+"`1.24.x`"+` | BSD | tarball | build |
`), 0o644); err != nil {
		t.Fatal(err)
	}

	bin := filepath.Join(tmp, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "ngspice"), "#!/usr/bin/env bash\necho 'ngspice-42'\n")
	writeExecutable(t, filepath.Join(bin, "iverilog"), "#!/usr/bin/env bash\necho 'Icarus Verilog version 12.0'\n")
	writeExecutable(t, filepath.Join(bin, "verilator"), "#!/usr/bin/env bash\necho 'Verilator 5.028'\n")
	writeExecutable(t, filepath.Join(bin, "go"), "#!/usr/bin/env bash\necho 'go version go1.24.7 linux/amd64'\n")
	writeExecutable(t, filepath.Join(bin, "python3"), "#!/usr/bin/env bash\necho 'Python 2.1.3'\n")

	cmd := exec.Command("bash", "../drift_detector.sh")
	cmd.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "PIN_FILE="+pinFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("drift_detector.sh failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "drift") || strings.Contains(string(out), "missing") {
		t.Fatalf("expected all tools to match pins, got:\n%s", out)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}
