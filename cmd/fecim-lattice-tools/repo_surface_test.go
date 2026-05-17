package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

func repoRootForRepoSurface() string {
	return filepath.Clean(filepath.Join("..", ".."))
}
