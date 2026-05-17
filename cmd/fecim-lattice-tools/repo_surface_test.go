package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNonLegacyPackagesDoNotDependOnLegacyGraphics(t *testing.T) {
	root := repoRootForRepoSurface()
	for _, pkg := range listRepoPackages(t, root) {
		if isLegacyGraphicsPackage(pkg) {
			continue
		}
		for _, dep := range listRepoDeps(t, root, pkg) {
			if isLegacyGraphicsDependency(pkg, dep) {
				t.Errorf("non-legacy package %s must not depend on legacy graphics surface %s", pkg, dep)
			}
		}
	}
}

func TestDefaultRepoGraphDoesNotExposeLegacyFynePackages(t *testing.T) {
	root := repoRootForRepoSurface()
	for _, pkg := range listRepoPackages(t, root) {
		if isLegacyGraphicsPackage(pkg) {
			t.Fatalf("default repo graph must not expose legacy Fyne package %s", pkg)
		}
	}
}

func TestDefaultRepoGraphDoesNotExposeTransitionRedirectPackages(t *testing.T) {
	root := repoRootForRepoSurface()
	for _, pkg := range listRepoPackages(t, root) {
		if pkg == "fecim-lattice-tools/internal/gogpucommand" {
			t.Fatalf("default repo graph must not expose transition redirect package %s", pkg)
		}
	}
}

func TestFyneImportsAreLegacyTagged(t *testing.T) {
	root := repoRootForRepoSurface()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if isSkippedRepoSurfaceDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(data)
		importsFyne, err := fileImportsFyne(path, data)
		if err != nil {
			return err
		}
		if importsFyne && !hasLegacyFyneBuildTag(source) {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			t.Errorf("Go file importing Fyne must be tagged legacy_fyne: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo source files: %v", err)
	}
}

func TestLivingGuidanceUsesCanonicalGogpuSurface(t *testing.T) {
	root := repoRootForRepoSurface()
	files := []string{
		"CONTRIBUTING.md",
		"docs/README.md",
		"docs/1-getting-started/README.md",
		"tools/fecim-skills/_shared/fecim-context.md",
		"tools/fecim-skills/fecim-builder/SKILL.md",
		"tools/fecim-skills/fecim-gogpu-migrate/SKILL.md",
		"tools/fecim-skills/fecim-labtester/SKILL.md",
	}
	stalePhrases := []string{
		"current default desktop app remains the Fyne shell",
		"future zero-CGO",
		"cmd/fecim-lattice-tools-next",
		"make test-next-ui",
		"Next gogpu/ui shell",
		"Future shell",
		"Legacy Fyne shell: `cmd/fecim-lattice-tools`",
		"placeholder path until it reaches module parity",
		"go run ./cmd/demo-frames",
		"go run ./cmd/fecim-web",
		"go run ./cmd/write-proof",
	}
	for _, file := range files {
		body, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(body)
		for _, phrase := range stalePhrases {
			if strings.Contains(text, phrase) {
				t.Errorf("%s contains stale gogpu/Fyne guidance %q", file, phrase)
			}
		}
	}
}

func TestPublicGettingStartedDocsPresentGogpuAsDefault(t *testing.T) {
	root := repoRootForRepoSurface()
	staleFyneImportError := "**Error:** `cannot find package \"" + "fyne.io/" + "fyne/v2\"`"
	cases := map[string]struct {
		mustContain []string
		stale       []string
	}{
		"docs/README.md": {
			mustContain: []string{
				"gogpu/ui",
				"legacy Fyne",
			},
			stale: []string{
				"- **GUI:** Fyne 2.7.2",
			},
		},
		"docs/1-getting-started/README.md": {
			mustContain: []string{
				"No C compiler is required for the default gogpu/ui app.",
				"legacy_fyne",
			},
			stale: []string{
				"- **GCC:** C compiler (for Fyne GUI)",
				"sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev",
				"**Error:** `gcc: command not found`",
				staleFyneImportError,
				"**Error:** `undefined: GL_VERSION`",
				"FYNE_NO_GL=1 ./fecim-lattice-tools",
				"GDK_BACKEND=x11 ./fecim-lattice-tools",
			},
		},
	}
	for file, tc := range cases {
		body, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(body)
		for _, phrase := range tc.mustContain {
			if !strings.Contains(text, phrase) {
				t.Errorf("%s must present %q", file, phrase)
			}
		}
		for _, phrase := range tc.stale {
			if strings.Contains(text, phrase) {
				t.Errorf("%s presents stale default Fyne guidance %q", file, phrase)
			}
		}
	}
}

func TestInstallationGuideScopesCgoToLegacyFyne(t *testing.T) {
	root := repoRootForRepoSurface()
	file := "docs/1-getting-started/installation.md"
	body, err := os.ReadFile(filepath.Join(root, file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	text := string(body)
	defaultSection := strings.Split(text, "## Legacy Fyne parity only (`-tags legacy_fyne`)")[0]
	mustContain := []string{
		"The default gogpu/ui app requires Go and Git only; no CGO, C compiler, or OpenGL headers are required.",
		"Legacy Fyne parity only (`-tags legacy_fyne`)",
	}
	staleDefaultGuidance := []string{
		"- **C compiler** (gcc/clang) for CGO",
		"- **OpenGL libraries**",
		"sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev",
		"sudo dnf install -y gcc mesa-libGL-devel",
		"xcode-select --install  # Install command line tools",
		"Install MSYS2 (https://www.msys2.org/) or TDM-GCC",
	}
	for _, phrase := range mustContain {
		if !strings.Contains(text, phrase) {
			t.Errorf("%s must present %q", file, phrase)
		}
	}
	for _, phrase := range staleDefaultGuidance {
		if strings.Contains(defaultSection, phrase) {
			t.Errorf("%s presents CGO/OpenGL as default-app installation guidance %q", file, phrase)
		}
	}
}

func TestDeveloperDocsPresentGogpuArchitectureAsDefault(t *testing.T) {
	root := repoRootForRepoSurface()
	cases := map[string]struct {
		mustContain []string
		stale       []string
	}{
		"docs/3-develop/README.md": {
			mustContain: []string{
				"Default UI shell: `gogpu/ui`",
				"Legacy Fyne shell: `cmd/fecim-lattice-tools-fyne` with `-tags legacy_fyne`",
				"`shared/viewmodel`",
			},
			stale: []string{
				"sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev",
				"BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject",
				"All UI updates from goroutines must use `fyne.Do()`",
				"| `shared/widgets` | `fecim-lattice-tools/shared/widgets` | Fyne GUI components |",
				"**Fyne Version:** 2.7.2",
			},
		},
		"docs/3-develop/architecture/ARCHITECTURE.md": {
			mustContain: []string{
				"Default shell: `gogpu/ui`",
				"UI-neutral state: `shared/viewmodel`",
				"Legacy Fyne shell: `cmd/fecim-lattice-tools-fyne` with `-tags legacy_fyne`",
			},
			stale: []string{
				"│  Fyne App",
				"All modules use **Fyne**",
				"BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject",
				"Fyne provides both high-level widgets and low-level canvas",
			},
		},
	}
	for file, tc := range cases {
		body, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(body)
		for _, phrase := range tc.mustContain {
			if !strings.Contains(text, phrase) {
				t.Errorf("%s must present %q", file, phrase)
			}
		}
		for _, phrase := range tc.stale {
			if strings.Contains(text, phrase) {
				t.Errorf("%s presents stale Fyne-default architecture guidance %q", file, phrase)
			}
		}
	}
}

func listRepoPackages(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list ./... failed: %v\n%s", err, out)
	}
	return strings.Fields(string(out))
}

func listRepoDeps(t *testing.T, root string, pkg string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "-deps", pkg)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list -deps %s failed: %v\n%s", pkg, err, out)
	}
	return strings.Fields(string(out))
}

func isLegacyGraphicsDependency(pkg string, dep string) bool {
	disallowed := []string{
		"fyne.io/" + "fyne",
		"github.com/go-gl/glfw",
		"fecim-lattice-tools/shared/theme",
		"fecim-lattice-tools/shared/themes",
		"fecim-lattice-tools/shared/widgets",
	}
	for _, needle := range disallowed {
		if strings.HasPrefix(dep, needle) {
			return true
		}
	}
	if strings.HasPrefix(dep, "github.com/vulkan-go/vulkan") && !isAllowedVulkanComputePackage(pkg) {
		return true
	}
	return false
}

func isAllowedVulkanComputePackage(pkg string) bool {
	return pkg == "fecim-lattice-tools/shared/compute" ||
		pkg == "fecim-lattice-tools/module4-circuits/pkg/gpuperiph"
}

func isLegacyGraphicsPackage(pkg string) bool {
	if strings.Contains(pkg, "-fyne") {
		return true
	}
	legacyAreas := []string{
		"/pkg/gui",
		"/shared/theme",
		"/shared/themes",
		"/shared/widgets",
	}
	for _, area := range legacyAreas {
		if strings.Contains(pkg, area) {
			return true
		}
	}
	return false
}

func isSkippedRepoSurfaceDir(name string) bool {
	switch name {
	case ".git", ".worktrees", "artifacts", "tmp":
		return true
	default:
		return false
	}
}

func hasLegacyFyneBuildTag(source string) bool {
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			return false
		}
		if strings.HasPrefix(trimmed, "//go:build") {
			return strings.Contains(trimmed, "legacy_fyne")
		}
	}
	return false
}

func fileImportsFyne(path string, data []byte) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
	if err != nil {
		return false, err
	}
	for _, imported := range file.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			return false, err
		}
		if strings.HasPrefix(importPath, "fyne.io/"+"fyne/v2") {
			return true, nil
		}
	}
	return false, nil
}

func repoRootForRepoSurface() string {
	return filepath.Clean(filepath.Join("..", ".."))
}
