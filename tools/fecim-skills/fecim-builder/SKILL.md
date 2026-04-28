---
name: fecim-builder
description: Runs build flows for both UI paths (legacy Fyne and zero-CGO gogpu/ui shell) on this Go 1.25 monorepo. Use when building, packaging, or debugging build failures in cmd/fecim-lattice-tools or cmd/fecim-lattice-tools-next.
---

# fecim-builder

Build the FeCIM Lattice Tools binary on either the legacy Fyne path or the zero-CGO gogpu/ui shell.

See `tools/fecim-skills/_shared/fecim-context.md` (Build target matrix) for the canonical CGO/entry-point mapping.

## Workflow

1. **Identify the target** — ask the user if unclear:
   - Legacy Fyne shell: `cmd/fecim-lattice-tools`
   - Next gogpu/ui shell: `cmd/fecim-lattice-tools-next`

2. **Set the build environment:**
   - Legacy: leave `CGO_ENABLED` at its default (`1`); ensure GLFW/X11 deps installed (`sudo apt-get install -y libgl1-mesa-dev xorg-dev` on Linux).
   - Next: `export CGO_ENABLED=0`.

3. **Run the build:**
   - Legacy single-binary: `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`
   - Legacy launch: `./launch.sh`
   - Next: `CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next`
   - Whole repo: `go build ./...`

4. **On failure, triage:**

   | Symptom | Cause | Fix |
   |---|---|---|
   | `fatal error: GL/gl.h: No such file` | Missing OpenGL headers | `sudo apt-get install -y libgl1-mesa-dev xorg-dev` |
   | `cannot find -lvulkan` | Vulkan loader missing | `sudo apt-get install -y libvulkan-dev` (optional dep, can be omitted) |
   | `gcc not found` | CGO toolchain missing | `sudo apt-get install -y gcc` |
   | `package github.com/gogpu/ui: cannot find module` | gogpu/ui import in non-shell pkg | UI-boundary violation; move logic to `shared/viewmodel/` per AGENTS.md |
   | `imports fyne.io/fyne/v2` from `viewmodel` | UI-boundary violation | Same — strip Fyne import, use viewmodel pure types |

5. **Verify:**
   - Binary exists and is executable.
   - For Next path, confirm `CGO_ENABLED=0` was respected (`go env CGO_ENABLED` should print `0` in the same shell).

## Verification

- Input: "Build the legacy GUI."
  Expected: runs `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`, reports success.
- Input: "Build the next shell." with missing libvulkan-dev.
  Expected: succeeds since libvulkan is optional, OR triages with the table above.

## TDD

Build invocations are observation, not behavior change — `TDD: N/A`. Any code change discovered during triage triggers the project's TDD hard-rule. See `tools/fecim-skills/_shared/tdd-evidence-template.md`.
