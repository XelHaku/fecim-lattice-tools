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
   - Legacy: leave `CGO_ENABLED` at its default (`1`); if GLFW/X11 headers are missing, report the exact blocker. Do **not** install host packages yourself.
   - Next: `export CGO_ENABLED=0`.

3. **Run the build:**
   - Legacy single-binary: `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`
   - Legacy launch: `./launch.sh`
   - Next: `CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next`
   - Whole repo: `go build ./...`
   - Preflight first: verify repo path, `go`, `git`, and `rg`/fallback using `_shared/fecim-context.md`.

4. **On failure, triage:**

   | Symptom | Cause | Fix |
   |---|---|---|
   | `fatal error: GL/gl.h: No such file` | Missing OpenGL headers | Report blocker with command output; operator/admin installs packages. |
   | `cannot find -lvulkan` | Vulkan loader missing | Report blocker for legacy/Vulkan path; optional dep may be omitted when not required. |
   | `gcc not found` | CGO toolchain missing | Report blocker with `command -v gcc` evidence; operator/admin installs packages. |
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
