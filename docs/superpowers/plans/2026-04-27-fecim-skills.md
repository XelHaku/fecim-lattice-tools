# FeCIM Skills Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author 8 FeCIM-specific agent skills (`fecim-researcher`, `fecim-citation`, `fecim-builder`, `fecim-labtester`, `fecim-grill`, `fecim-fyne-thread-check`, `fecim-honesty-audit`, `fecim-gogpu-migrate`) distributed across Claude Code, Codex, and opencode from a single canonical SKILL.md per skill.

**Architecture:** Canonical SKILL.md files at `tools/fecim-skills/<name>/SKILL.md` (mattpocock-format frontmatter). A bash install script (`scripts/install-fecim-skills.sh`) generates per-harness adapters: symlinks for Claude Code (`.claude/skills/<name>/SKILL.md`), a managed block in `.codex/AGENTS.md` for Codex, and command files for opencode (`.opencode/command/<name>.md`). Generated adapters are committed. Install script has full TDD via `scripts/test-install-fecim-skills.sh`. CI runs `--check` to catch drift.

**Tech Stack:** bash 5+, GNU coreutils, GNU awk, `git`, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-04-27-fecim-skills-design.md`

**Phases:**
- **Phase 1 (PR1):** Pipeline + `fecim-builder` end-to-end (Tasks 1–24).
- **Phase 2 (PR2):** Remaining 7 skills (Tasks 25–32).

## File Structure

| Path | Responsibility | Phase |
|---|---|---|
| `tools/fecim-skills/README.md` | Index of skills + install instructions | 1 |
| `tools/fecim-skills/_shared/fecim-context.md` | FeCIM domain primer (citations, modules, honesty rules) | 1 |
| `tools/fecim-skills/_shared/tdd-evidence-template.md` | RED/GREEN/verification block referenced by skills | 1 |
| `tools/fecim-skills/fecim-builder/SKILL.md` | Build skill (legacy + next) | 1 |
| `tools/fecim-skills/fecim-{researcher,citation,labtester,grill,fyne-thread-check,honesty-audit,gogpu-migrate}/SKILL.md` | Remaining 7 skills | 2 |
| `scripts/install-fecim-skills.sh` | Generates per-harness adapters; supports `--check` | 1 |
| `scripts/test-install-fecim-skills.sh` | Test harness for install script | 1 |
| `.claude/skills/<name>/SKILL.md` | Generated Claude Code adapter (symlink or copy) | 1 (1 skill), 2 (rest) |
| `.codex/AGENTS.md` | Generated Codex managed block | 1 (1 skill), 2 (rest) |
| `.opencode/command/<name>.md` | Generated opencode command file | 1 (1 skill), 2 (rest) |
| `Makefile` | Add `install-skills`, `test-skills` targets | 1 |
| `.github/workflows/ci.yml` | Add sync-check step | 1 |
| `AGENTS.md` | One-line pointer to fecim-skills | 1 |
| `docs/3-develop/README.md` | One-line pointer to fecim-skills | 1 |

---

# Phase 1 — Pipeline + fecim-builder

## Task 1: Create directory skeleton and shared assets

**Files:**
- Create: `tools/fecim-skills/README.md`
- Create: `tools/fecim-skills/_shared/fecim-context.md`
- Create: `tools/fecim-skills/_shared/tdd-evidence-template.md`

- [ ] **Step 1: Create `tools/fecim-skills/README.md`**

```markdown
# FeCIM Skills

FeCIM-specific agent skills distributed across Claude Code, Codex, and opencode.

Each skill lives in `<name>/SKILL.md` with mattpocock-format YAML frontmatter (`name`, `description`). The install script `scripts/install-fecim-skills.sh` generates adapters for each harness; generated adapters are committed.

## Install / refresh

```bash
make install-skills        # regenerate all per-harness adapters
make test-skills           # run install-script tests
scripts/install-fecim-skills.sh --check    # CI sync check
```

## Skills

| Skill | Use when |
|---|---|
| `fecim-researcher` | Investigating a FeCIM physics topic |
| `fecim-citation` | Adding numeric/measurement claims to code or docs |
| `fecim-builder` | Building or debugging build failures |
| `fecim-labtester` | Running tests, regenerating physics goldens |
| `fecim-grill` | Starting a non-trivial physics/GUI change |
| `fecim-fyne-thread-check` | Reviewing PRs that add goroutines in `pkg/gui/` |
| `fecim-honesty-audit` | Committing docs/PRs with accuracy or efficiency numbers |
| `fecim-gogpu-migrate` | Porting a Fyne tab to the gogpu/ui shell |

## Editing

Edit only `tools/fecim-skills/<name>/SKILL.md`. Generated adapters carry a `do not edit` header — they're rewritten on next install. CI fails if generated files drift from canonical.
```

- [ ] **Step 2: Create `tools/fecim-skills/_shared/fecim-context.md`**

```markdown
# FeCIM Domain Context (Skill Reference)

Shared context referenced by FeCIM skills. Do not duplicate this content in individual SKILL.md files.

## Published-physics citation list (canonical short form)

| Short form | Full citation | Used for |
|---|---|---|
| Materlik 2015 | Materlik et al., J. Appl. Phys. 117, 134109 (2015) | HZO Landau parameters |
| Park 2015 | Park et al., Adv. Mater. 27, 1811 (2015) | HZO ferroelectricity |
| Alessandri 2018 | Alessandri et al., IEEE TED 65, 4503 (2018) | Polarization switching dynamics |
| Guo 2018 | Guo et al., Nano Lett. 18, 1727 (2018) | Crossbar device characterization |
| HZO FTJ 2025 | J. Alloys Compd. (2025) | 98.24% MNIST reservoir computing — NOT a FeCIM device claim |

## Honesty-audit policy (`docs/4-research/honesty-audit.md`)

**Verified claims (peer-reviewed):** 98.24% MNIST in HZO FTJ reservoir computing — must be attributed to HZO FTJ 2025 paper, NOT this simulator.

**Simulation defaults (educational, must be labeled as such):**
- 30 analog conductance levels (configurable simulation default)
- Material parameters (HZO, BTO, PZT) from published physics

**Removed/unverified — flag as policy violation:**
- "30 analog states" presented as device fact (conference-only baseline)
- "87% MNIST accuracy" attributed to FeCIM simulator (conference-only)
- Energy multipliers vs NAND or GPUs without published measurement evidence

## 5 known bug patterns (`MEMORY.md`)

1. **Guard-band sign direction flip** — limit guard pulses to 2 max, clamp `calcLevel` to prevent overshoot.
2. **Bounds collapse `[VMin, VMax]`** — widen minimally using direction info, don't reset to full range.
3. **ACCEPT ±1 guard interaction** — skip ACCEPT ±1 when `guardActive=true`, raise overshoot threshold from 3 to 8.
4. **Zero-field bounds reset** — when `absField < 0.01*Ec`, reset to full `[0, MaxField]`.
5. **Preisach Everett zero-clamp** — use product-form (always non-negative) instead of factorized (goes negative).

## UI thread-safety rule (`docs/3-develop/gui/FYNE_NOTES.md`)

All UI updates from goroutines MUST use `fyne.Do(func() { ... })`. Direct widget mutation from a non-main goroutine causes hangs and freezes. Common patterns to audit:

```go
// WRONG
go func() {
    result := compute()
    label.SetText(result)   // RACE
}()

// RIGHT
go func() {
    result := compute()
    fyne.Do(func() { label.SetText(result) })
}()
```

## UI boundary rule (per `AGENTS.md`)

New UI-neutral, physics, simulation, validation, and export work must NOT add `fyne.io/...` or `github.com/gogpu/ui` imports. Use `shared/viewmodel/` as the UI-neutral bridge. Fyne and gogpu/ui imports belong only in shell/UI packages.

## Build target matrix

| Target | CGO | Entry | Use |
|---|---|---|---|
| Legacy | `CGO_ENABLED=1` (default) | `./cmd/fecim-lattice-tools` | Current Fyne GUI |
| Next | `CGO_ENABLED=0` | `./cmd/fecim-lattice-tools-next` | Future zero-CGO gogpu/ui shell |

## Test invocations

| Command | Use |
|---|---|
| `go test ./...` | Full suite |
| `go test -race ./...` | Race detection (mandatory when changing concurrency) |
| `make test-next-ui` | Future zero-CGO UI shell tests |
| `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...` | Regenerate physics regression goldens (only when divergence is intentional) |
```

- [ ] **Step 3: Create `tools/fecim-skills/_shared/tdd-evidence-template.md`**

```markdown
# TDD Evidence Block

Per CLAUDE.md TDD hard-rule, code-mutating skills must end their workflow output with this block:

```
RED: <command>
     <expected failure summary>

GREEN: <command>
       <expected pass summary>

VERIFY: <final command(s)>
```

For documentation-only, comments-only, formatting-only, generated files, or release metadata, mark `TDD: N/A` with a short reason.
```

- [ ] **Step 4: Commit**

```bash
git add tools/fecim-skills/README.md tools/fecim-skills/_shared/
git commit -m "feat(skills): scaffold tools/fecim-skills with shared context

TDD: N/A (documentation/configuration only, no behavior)"
```

---

## Task 2: Write canonical `fecim-builder/SKILL.md`

**Files:**
- Create: `tools/fecim-skills/fecim-builder/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
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

Build invocations are observation, not behavior change — `TDD: N/A`. Any code change discovered during triage triggers the project's TDD hard-rule. See `_shared/tdd-evidence-template.md`.
```

- [ ] **Step 2: Commit**

```bash
git add tools/fecim-skills/fecim-builder/SKILL.md
git commit -m "feat(skills): add fecim-builder canonical SKILL.md

TDD: N/A (documentation/configuration only, no behavior)"
```

---

## Task 3: Test harness scaffold + first failing test

**Files:**
- Create: `scripts/test-install-fecim-skills.sh`
- Create: `scripts/install-fecim-skills.sh` (empty stub)

- [ ] **Step 1: Create install script stub**

```bash
cat > scripts/install-fecim-skills.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "stub: not implemented" >&2
exit 1
EOF
chmod +x scripts/install-fecim-skills.sh
```

- [ ] **Step 2: Create test harness with the first scenario (fresh install)**

```bash
cat > scripts/test-install-fecim-skills.sh <<'BASH'
#!/usr/bin/env bash
# Tests for scripts/install-fecim-skills.sh.
# Each test sets up a fixture repo in $TMPDIR, runs the script, and asserts.
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
INSTALL_SCRIPT="$REPO_ROOT/scripts/install-fecim-skills.sh"

PASS=0
FAIL=0
FAIL_DETAILS=()

run_test() {
  local name="$1" body="$2"
  local fixture; fixture=$(mktemp -d -t fecim-skills-test-XXXXXX)
  trap "rm -rf '$fixture'" RETURN
  if (cd "$fixture" && eval "$body") >/dev/null 2>&1; then
    PASS=$((PASS+1))
    echo "PASS: $name"
  else
    FAIL=$((FAIL+1))
    FAIL_DETAILS+=("$name")
    echo "FAIL: $name"
  fi
}

# Helper: create a minimal fixture repo with one valid skill.
fixture_one_skill() {
  git init -q
  mkdir -p tools/fecim-skills/sample-skill
  cat > tools/fecim-skills/sample-skill/SKILL.md <<'SKILL'
---
name: sample-skill
description: A sample skill for testing.
---

# sample-skill
Body.
SKILL
  cp "$INSTALL_SCRIPT" install.sh
}

# ---------- Tests ----------

run_test "fresh install creates all three adapter targets" '
  fixture_one_skill
  ./install.sh
  test -e .claude/skills/sample-skill/SKILL.md
  test -f .codex/AGENTS.md && grep -q "fecim-skills:start" .codex/AGENTS.md
  test -f .opencode/command/sample-skill.md
'

# ---------- Summary ----------
echo
echo "Passed: $PASS  Failed: $FAIL"
if (( FAIL > 0 )); then
  for d in "${FAIL_DETAILS[@]}"; do echo "  - $d"; done
  exit 1
fi
BASH
chmod +x scripts/test-install-fecim-skills.sh
```

- [ ] **Step 3: Run the test to verify it fails (RED)**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `FAIL: fresh install creates all three adapter targets` (because install script is a stub).

- [ ] **Step 4: Commit RED state**

```bash
git add scripts/install-fecim-skills.sh scripts/test-install-fecim-skills.sh
git commit -m "test(skills): scaffold install-fecim-skills test harness (RED)

RED: scripts/test-install-fecim-skills.sh
     FAIL: fresh install creates all three adapter targets

TDD: behavioral test for install script."
```

---

## Task 4: Implement install script — minimal pass

**Files:**
- Modify: `scripts/install-fecim-skills.sh`

- [ ] **Step 1: Replace the stub with the full install logic**

```bash
cat > scripts/install-fecim-skills.sh <<'EOF'
#!/usr/bin/env bash
# Generate per-harness adapters from canonical tools/fecim-skills/<name>/SKILL.md.
# Source of truth: docs/superpowers/specs/2026-04-27-fecim-skills-design.md
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SKILLS_DIR="$REPO_ROOT/tools/fecim-skills"
CHECK_MODE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check) CHECK_MODE=1; shift ;;
    -h|--help) sed -n '2,4p' "$0"; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; exit 2 ;;
  esac
done

[[ -d "$SKILLS_DIR" ]] || { echo "ERROR: $SKILLS_DIR not found" >&2; exit 1; }
mapfile -t SKILLS < <(find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md | sort)
[[ ${#SKILLS[@]} -gt 0 ]] || { echo "ERROR: no SKILL.md files in $SKILLS_DIR" >&2; exit 1; }

drift=0
note_drift() { echo "DRIFT: $1" >&2; drift=1; }

skill_name() { basename "$(dirname "$1")"; }
skill_description() {
  awk '/^description:/{sub(/^description: */, ""); print; exit}' "$1"
}

# Validate every SKILL.md has description.
for s in "${SKILLS[@]}"; do
  if [[ -z "$(skill_description "$s")" ]]; then
    echo "ERROR: $s missing 'description:' frontmatter" >&2
    exit 1
  fi
done

# ----- Claude Code adapter -----
emit_claude_adapter() {
  local skill_path="$1" name; name=$(skill_name "$skill_path")
  local target="$REPO_ROOT/.claude/skills/$name/SKILL.md"
  local rel="../../../tools/fecim-skills/$name/SKILL.md"

  if (( CHECK_MODE )); then
    if [[ -L "$target" && "$(readlink "$target")" == "$rel" ]]; then
      return
    fi
    if [[ -f "$target" && ! -L "$target" ]]; then
      diff -q <(claude_copy_body "$skill_path") "$target" >/dev/null 2>&1 || note_drift "$target"
      return
    fi
    note_drift "$target"
    return
  fi

  mkdir -p "$(dirname "$target")"
  rm -f "$target"
  if ln -s "$rel" "$target" 2>/dev/null; then
    return
  fi
  claude_copy_body "$skill_path" > "$target"
}

claude_copy_body() {
  local skill_path="$1" name; name=$(skill_name "$skill_path")
  echo "<!-- generated-from: tools/fecim-skills/$name/SKILL.md -->"
  echo "<!-- do not edit; run scripts/install-fecim-skills.sh -->"
  cat "$skill_path"
}

# ----- opencode adapter -----
emit_opencode_adapter() {
  local skill_path="$1" name; name=$(skill_name "$skill_path")
  local desc; desc=$(skill_description "$skill_path")
  local target="$REPO_ROOT/.opencode/command/$name.md"
  local body; body=$(opencode_body "$name" "$desc")

  if (( CHECK_MODE )); then
    if [[ -f "$target" ]] && diff -q <(echo "$body") "$target" >/dev/null 2>&1; then
      return
    fi
    note_drift "$target"
    return
  fi

  mkdir -p "$(dirname "$target")"
  echo "$body" > "$target"
}

opencode_body() {
  local name="$1" desc="$2"
  printf -- '---\ndescription: %s\nagent: build\n---\n\n<!-- generated-from: tools/fecim-skills/%s/SKILL.md -->\n<!-- do not edit; run scripts/install-fecim-skills.sh -->\n\nRead and follow the workflow in `tools/fecim-skills/%s/SKILL.md`.\n\nUse the user'"'"'s request as $ARGUMENTS.\n' "$desc" "$name" "$name"
}

# ----- Codex managed block -----
codex_block() {
  echo "<!-- fecim-skills:start -->"
  echo "## FeCIM Skills"
  echo
  echo "When the user's request matches the trigger description below, follow the workflow in the linked file."
  echo
  for s in "${SKILLS[@]}"; do
    local n d
    n=$(skill_name "$s"); d=$(skill_description "$s")
    echo "- **$n** — $d → see \`tools/fecim-skills/$n/SKILL.md\`"
  done
  echo
  echo "For all skills above, the canonical body is the linked SKILL.md file. Read it before acting."
  echo "<!-- fecim-skills:end -->"
}

emit_codex() {
  local target="$REPO_ROOT/.codex/AGENTS.md"
  local block; block=$(codex_block)
  local merged

  if [[ -f "$target" ]]; then
    if grep -q "<!-- fecim-skills:start -->" "$target"; then
      merged=$(awk -v b="$block" '
        /<!-- fecim-skills:start -->/ {print b; in_block=1; next}
        /<!-- fecim-skills:end -->/   {in_block=0; next}
        !in_block {print}
      ' "$target")
    else
      merged=$(printf '%s\n\n%s\n' "$(cat "$target")" "$block")
    fi
  else
    merged="$block"
  fi

  if (( CHECK_MODE )); then
    if [[ -f "$target" ]] && diff -q <(echo "$merged") "$target" >/dev/null 2>&1; then
      return
    fi
    note_drift "$target"
    return
  fi

  mkdir -p "$(dirname "$target")"
  echo "$merged" > "$target"
}

# ----- Run -----
for s in "${SKILLS[@]}"; do
  emit_claude_adapter "$s"
  emit_opencode_adapter "$s"
done
emit_codex

if (( CHECK_MODE )); then
  exit "$drift"
fi
echo "Installed ${#SKILLS[@]} skills"
EOF
chmod +x scripts/install-fecim-skills.sh
```

- [ ] **Step 2: Run the test to verify it now passes (GREEN)**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `PASS: fresh install creates all three adapter targets`, `Passed: 1  Failed: 0`.

- [ ] **Step 3: Commit GREEN state**

```bash
git add scripts/install-fecim-skills.sh
git commit -m "feat(skills): implement install script (GREEN, 1 scenario)

RED: PASS: fresh install creates all three adapter targets (was FAIL)
GREEN: scripts/test-install-fecim-skills.sh passes 1/1
VERIFY: bash -n scripts/install-fecim-skills.sh"
```

---

## Task 5: Test — idempotency (RED → GREEN)

**Files:**
- Modify: `scripts/test-install-fecim-skills.sh`

- [ ] **Step 1: Add idempotency test**

Insert before the `# ---------- Summary ----------` line:

```bash
run_test "second install produces no diff against first install" '
  fixture_one_skill
  ./install.sh
  cp -r .claude .claude.bak
  cp -r .codex .codex.bak
  cp -r .opencode .opencode.bak
  ./install.sh
  diff -r .claude .claude.bak
  diff -r .codex .codex.bak
  diff -r .opencode .opencode.bak
'
```

- [ ] **Step 2: Run tests; expect PASS (idempotency was already designed in)**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `Passed: 2  Failed: 0`. If FAIL, the install script's output is non-deterministic — fix the source of nondeterminism (e.g., date stamps) before continuing.

- [ ] **Step 3: Commit**

```bash
git add scripts/test-install-fecim-skills.sh
git commit -m "test(skills): assert install is idempotent

GREEN: 2/2 tests pass"
```

---

## Task 6: Test — Codex preserves user content (RED → GREEN)

**Files:**
- Modify: `scripts/test-install-fecim-skills.sh`

- [ ] **Step 1: Add the test**

Insert before `# ---------- Summary ----------`:

```bash
run_test "codex install preserves user content outside markers" '
  fixture_one_skill
  mkdir -p .codex
  printf "Pre-existing line above\n\nAnother user line\n" > .codex/AGENTS.md
  ./install.sh
  grep -q "Pre-existing line above" .codex/AGENTS.md
  grep -q "Another user line" .codex/AGENTS.md
  grep -q "fecim-skills:start" .codex/AGENTS.md
'
```

- [ ] **Step 2: Run tests**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `Passed: 3  Failed: 0`. The append-when-no-marker branch in `emit_codex` already supports this; if FAIL, debug the awk merge.

- [ ] **Step 3: Commit**

```bash
git add scripts/test-install-fecim-skills.sh
git commit -m "test(skills): codex install preserves non-managed content

GREEN: 3/3 tests pass"
```

---

## Task 7: Test — `--check` detects drift (RED → GREEN)

**Files:**
- Modify: `scripts/test-install-fecim-skills.sh`

- [ ] **Step 1: Add the drift test**

Insert before `# ---------- Summary ----------`:

```bash
run_test "--check exits non-zero when generated adapter drifts" '
  fixture_one_skill
  ./install.sh
  echo "tampered" >> .opencode/command/sample-skill.md
  ! ./install.sh --check
'
```

- [ ] **Step 2: Run tests**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `Passed: 4  Failed: 0`. The `--check` mode in the script already sets `drift=1` and exits non-zero on diff; if FAIL, ensure `note_drift` is called from each emitter when content differs.

- [ ] **Step 3: Commit**

```bash
git add scripts/test-install-fecim-skills.sh
git commit -m "test(skills): --check detects adapter drift

GREEN: 4/4 tests pass"
```

---

## Task 8: Test — missing `description:` fails loudly (RED → GREEN)

**Files:**
- Modify: `scripts/test-install-fecim-skills.sh`

- [ ] **Step 1: Add the test**

Insert before `# ---------- Summary ----------`:

```bash
run_test "missing description frontmatter fails install" '
  git init -q
  mkdir -p tools/fecim-skills/broken
  cat > tools/fecim-skills/broken/SKILL.md <<SKILL
---
name: broken
---

# broken
SKILL
  cp "$INSTALL_SCRIPT" install.sh
  err=$(./install.sh 2>&1 || true)
  echo "$err" | grep -q "missing .description:. frontmatter"
  echo "$err" | grep -q "tools/fecim-skills/broken/SKILL.md"
'
```

- [ ] **Step 2: Run tests**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `Passed: 5  Failed: 0`. Validation already in install script; if FAIL, confirm `skill_description` returns empty for missing field and the validation loop runs first.

- [ ] **Step 3: Commit**

```bash
git add scripts/test-install-fecim-skills.sh
git commit -m "test(skills): install fails on missing description frontmatter

GREEN: 5/5 tests pass"
```

---

## Task 9: Test — symlink fallback path (RED → GREEN)

**Files:**
- Modify: `scripts/install-fecim-skills.sh`
- Modify: `scripts/test-install-fecim-skills.sh`

- [ ] **Step 1: Add the test**

Insert before `# ---------- Summary ----------`:

```bash
run_test "symlink failure falls back to copy with managed-block header" '
  fixture_one_skill
  FECIM_FORCE_NO_SYMLINK=1 ./install.sh
  test -f .claude/skills/sample-skill/SKILL.md
  test ! -L .claude/skills/sample-skill/SKILL.md
  head -1 .claude/skills/sample-skill/SKILL.md | grep -q "generated-from"
'
```

- [ ] **Step 2: Run test — expect FAIL (env var not yet honored)**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `FAIL: symlink failure falls back to copy with managed-block header`.

- [ ] **Step 3: Patch install script to honor `FECIM_FORCE_NO_SYMLINK`**

In `scripts/install-fecim-skills.sh`, replace the body of `emit_claude_adapter` block that does `if ln -s ...; then return; fi` with:

```bash
  if [[ "${FECIM_FORCE_NO_SYMLINK:-0}" != "1" ]] && ln -s "$rel" "$target" 2>/dev/null; then
    return
  fi
  claude_copy_body "$skill_path" > "$target"
```

(Find the existing `if ln -s "$rel" "$target" 2>/dev/null; then` line in `emit_claude_adapter` and update it to gate on the env var.)

- [ ] **Step 4: Run tests — expect GREEN**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `Passed: 6  Failed: 0`.

- [ ] **Step 5: Commit**

```bash
git add scripts/install-fecim-skills.sh scripts/test-install-fecim-skills.sh
git commit -m "feat(skills): symlink fallback to copy with header

RED: FAIL: symlink failure falls back to copy with managed-block header
GREEN: 6/6 tests pass
VERIFY: bash -n scripts/install-fecim-skills.sh"
```

---

## Task 10: Test — opencode and Claude bodies match expected shape (RED → GREEN)

**Files:**
- Modify: `scripts/test-install-fecim-skills.sh`

- [ ] **Step 1: Add structural assertions**

Insert before `# ---------- Summary ----------`:

```bash
run_test "opencode adapter has expected frontmatter and body" '
  fixture_one_skill
  ./install.sh
  head -1 .opencode/command/sample-skill.md | grep -q "^---$"
  grep -q "^description: A sample skill for testing.$" .opencode/command/sample-skill.md
  grep -q "^agent: build$" .opencode/command/sample-skill.md
  grep -q "tools/fecim-skills/sample-skill/SKILL.md" .opencode/command/sample-skill.md
'

run_test "claude adapter copy contains do-not-edit header" '
  fixture_one_skill
  FECIM_FORCE_NO_SYMLINK=1 ./install.sh
  grep -q "do not edit" .claude/skills/sample-skill/SKILL.md
  grep -q "name: sample-skill" .claude/skills/sample-skill/SKILL.md
'
```

- [ ] **Step 2: Run tests**

Run: `scripts/test-install-fecim-skills.sh`
Expected: `Passed: 8  Failed: 0`.

- [ ] **Step 3: Commit**

```bash
git add scripts/test-install-fecim-skills.sh
git commit -m "test(skills): assert opencode and claude adapter shapes

GREEN: 8/8 tests pass"
```

---

## Task 11: Wire up Makefile targets

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add `.PHONY` entries and targets**

Find the existing `.PHONY:` line near the top of `Makefile` and append `install-skills test-skills` to it. Then append at the end of the file:

```makefile
# Skills (FeCIM agent skills)
install-skills:
	scripts/install-fecim-skills.sh

test-skills:
	scripts/test-install-fecim-skills.sh
```

- [ ] **Step 2: Verify the targets work**

Run: `make test-skills`
Expected: `Passed: 8  Failed: 0`.

Run: `make install-skills`
Expected: `Installed 1 skills` (only `fecim-builder` exists in canonical so far).

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat(build): add install-skills and test-skills make targets

TDD: N/A (build glue)"
```

---

## Task 12: Run install once, commit generated adapters

**Files:**
- Create: `.claude/skills/fecim-builder/SKILL.md` (symlink)
- Create: `.codex/AGENTS.md`
- Create: `.opencode/command/fecim-builder.md`

- [ ] **Step 1: Run the installer**

Run: `make install-skills`
Expected output: `Installed 1 skills`, with three new files visible in `git status`.

- [ ] **Step 2: Verify shape**

Run:
```bash
git status --short .claude .codex .opencode
file .claude/skills/fecim-builder/SKILL.md
cat .opencode/command/fecim-builder.md
grep -A2 "fecim-skills:start" .codex/AGENTS.md
```

Expected:
- `.claude/skills/fecim-builder/SKILL.md`: symbolic link to `../../../tools/fecim-skills/fecim-builder/SKILL.md`
- `.opencode/command/fecim-builder.md`: has `agent: build` frontmatter
- `.codex/AGENTS.md`: contains the managed block with one bullet for `fecim-builder`

- [ ] **Step 3: Commit generated adapters**

```bash
git add .claude .codex .opencode
git commit -m "feat(skills): commit generated adapters for fecim-builder

TDD: N/A (generated artifacts)"
```

---

## Task 13: Add CI sync check

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Insert sync check after `Architecture check`**

Find the `- name: Architecture check` block in `.github/workflows/ci.yml`. Insert immediately after it:

```yaml
      - name: Verify fecim-skills are in sync
        run: scripts/install-fecim-skills.sh --check

      - name: Test fecim-skills install script
        run: scripts/test-install-fecim-skills.sh
```

- [ ] **Step 2: Lint the workflow file**

Run: `python3 -c 'import yaml; yaml.safe_load(open(".github/workflows/ci.yml"))'`
Expected: no output (YAML valid).

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci(skills): add fecim-skills sync check and test step

TDD: N/A (CI glue)"
```

---

## Task 14: Update `AGENTS.md` and `docs/3-develop/README.md` pointers

**Files:**
- Modify: `AGENTS.md`
- Modify: `docs/3-develop/README.md`

- [ ] **Step 1: Add row to the AGENTS.md "Quick Reference" table**

In `AGENTS.md`, locate the `### Quick Reference` table. Insert a row before the `If qmd emits CUDA build output...` line (i.e., as the last row of the table):

```markdown
| Use a FeCIM skill (researcher, builder, etc.) | `tools/fecim-skills/README.md` |
```

- [ ] **Step 2: Add a one-line bullet to `docs/3-develop/README.md` Documentation Index**

Open `docs/3-develop/README.md`, locate `## 📖 Documentation Index`. Add immediately under that heading (or at the end of its first list, whichever the file structure has — read first to confirm exact location):

```markdown
- **FeCIM Skills:** `tools/fecim-skills/README.md` — agent skills for Claude Code, Codex, opencode
```

- [ ] **Step 3: Commit**

```bash
git add AGENTS.md docs/3-develop/README.md
git commit -m "docs: point AGENTS.md and dev guide to fecim-skills

TDD: N/A (documentation only)"
```

---

## Task 15: Phase 1 verification

- [ ] **Step 1: Full local sanity**

Run:
```bash
make test-skills && \
make install-skills && \
scripts/install-fecim-skills.sh --check && \
git diff --exit-code
```
Expected: tests pass, install reports `Installed 1 skills`, `--check` exits 0, no uncommitted diff.

- [ ] **Step 2: Verify the skill is loadable**

Run: `cat .claude/skills/fecim-builder/SKILL.md | head -5`
Expected: starts with `---\nname: fecim-builder\n...` (whether via symlink or copy).

- [ ] **Step 3: Open PR1**

```bash
git push -u origin <branch>
gh pr create --title "feat(skills): fecim-skills pipeline + fecim-builder" --body "$(cat <<'BODY'
## Summary

- Add `tools/fecim-skills/` with `_shared/` context + `fecim-builder` SKILL.md
- Add `scripts/install-fecim-skills.sh` (TDD-built; 8 scenarios)
- Add `scripts/test-install-fecim-skills.sh` test harness
- Wire `make install-skills` / `make test-skills`
- CI sync check via `--check`
- Generated adapters for fecim-builder committed

## Test plan
- [ ] `make test-skills` passes
- [ ] `make install-skills` is idempotent
- [ ] `scripts/install-fecim-skills.sh --check` exits 0 on clean tree
- [ ] CI is green

🤖 Generated with [Claude Code](https://claude.com/claude-code)
BODY
)"
```

---

# Phase 2 — Remaining 7 skills

Each task in this phase has the same shape: write the canonical SKILL.md, run installer, run `--check`, commit canonical + generated adapters together.

## Task 16: `fecim-researcher`

**Files:**
- Create: `tools/fecim-skills/fecim-researcher/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-researcher
description: Surveys FeCIM domain knowledge by searching references/, citations/, docs/4-research/, and the local Cognee KG, then synthesizes a cited research note. Use when investigating a physics topic, evaluating a paper, or grounding a design decision in literature.
---

# fecim-researcher

Survey FeCIM literature and project knowledge to ground a design decision or answer a physics question. See `tools/fecim-skills/_shared/fecim-context.md` for the canonical citation list.

## Workflow

1. **Define the question.** Write it in one sentence. If it has multiple sub-questions, split them and run the workflow per sub-question.

2. **Search local sources** in this order:
   - `docs/4-research/` (audits, validation notes, error propagation)
   - `references/` (academic papers, simulation benchmarks)
   - `citations/` (project's citation registry, if present)
   - `experimental-data/` (HZO, HfO2, crossbar characterization)

   Use `rg` with focused patterns (e.g., `rg -i "preisach" docs/4-research/ references/`).

3. **If Cognee is configured locally** (`.env` has `LLM_API_KEY`, `.cognee_system/` exists), query the KG:
   ```bash
   python3 - <<'PY'
   import os, asyncio
   os.environ["COGNEE_SKIP_CONNECTION_TEST"] = "true"
   os.environ["ENABLE_BACKEND_ACCESS_CONTROL"] = "false"
   import cognee
   async def main():
       results = await cognee.search("YOUR QUERY HERE")
       print(results)
   asyncio.run(main())
   PY
   ```
   Otherwise skip silently.

4. **Cite findings** using the canonical short forms from `_shared/fecim-context.md`. Never invent a citation; if no source supports a claim, label it "no source found, requires verification".

5. **Output a structured note**:
   ```
   Question: <...>
   Sources: <list with short citation forms>
   Finding: <2-5 sentences>
   Gaps: <what is unsupported, what should be measured>
   Recommended next step: <action>
   ```

## Verification

- Input: "What HZO Landau parameters do we use?"
  Expected: searches `module1-hysteresis/pkg/ferroelectric/material.go` and `references/`, cites Materlik 2015, returns parameters with units.

## TDD

Research output is observation — `TDD: N/A`. Any code change suggested by the research triggers the project's TDD hard-rule.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-researcher/ .claude/skills/fecim-researcher/ .codex/AGENTS.md .opencode/command/fecim-researcher.md
git commit -m "feat(skills): add fecim-researcher

TDD: N/A (configuration only)"
```

---

## Task 17: `fecim-citation`

**Files:**
- Create: `tools/fecim-skills/fecim-citation/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-citation
description: Verifies and formats FeCIM physics/measurement claims against the project's published-source list and docs/4-research/honesty-audit.md. Use when adding a numeric claim, accuracy figure, or device-parameter assertion to code, docs, PR descriptions, or commit messages.
---

# fecim-citation

Validate any quantitative or device-physics claim before it lands in code, docs, or commit messages. See `tools/fecim-skills/_shared/fecim-context.md` (Honesty-audit policy) for the rule list.

## Workflow

1. **Parse the claim.** Extract: subject (e.g., HZO coercive field), value (e.g., 1.0 MV/cm), context (educational vs validated).

2. **Match to canonical sources** (`_shared/fecim-context.md` table). If the claim corresponds to a published source, format as:
   ```
   <claim text> (Materlik 2015)
   ```

3. **Check honesty-audit removed/unverified list:**
   - "30 analog states" presented as device fact → REPHRASE: "30 analog conductance levels (configurable simulation default)".
   - "87% MNIST accuracy" as a FeCIM device claim → REMOVE; if discussing reservoir computing instead, attribute to HZO FTJ 2025 paper at 98.24% and label "not a FeCIM device claim".
   - Energy multipliers vs NAND/GPUs without measurement evidence → REMOVE or replace with literature-backed comparison.

4. **For uncited claims**, suggest one of:
   - Add the citation if a published source exists.
   - Reframe as "simulation default" with the project's standard wording.
   - Block the claim if no source and no educational framing applies.

5. **Output:**
   ```
   Claim: <as written>
   Status: verified | educational-default | removed-unverified | needs-source
   Suggested wording: <if change needed>
   Citation: <short form, if applicable>
   ```

## Verification

- Input: "Add 'Our simulator achieves 87% MNIST accuracy' to README.md."
  Expected: status `removed-unverified`; suggested wording reframes as the simulation pipeline's accuracy with explicit educational label, OR points to HZO FTJ 2025's 98.24% with "not a FeCIM device claim" caveat.

## TDD

Citation review is observation — `TDD: N/A`. Code/doc edits triggered by review follow the project's TDD hard-rule per `_shared/tdd-evidence-template.md`.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-citation/ .claude/skills/fecim-citation/ .codex/AGENTS.md .opencode/command/fecim-citation.md
git commit -m "feat(skills): add fecim-citation

TDD: N/A (configuration only)"
```

---

## Task 18: `fecim-labtester`

**Files:**
- Create: `tools/fecim-skills/fecim-labtester/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-labtester
description: Runs the FeCIM test matrix (full, race, module-scoped, coverage, golden regen) and interprets physics regression failures using the 5 known bug patterns. Use when running tests, debugging test failures, or regenerating physics golden files.
---

# fecim-labtester

Pick the right test invocation, run it, and triage failures against the 5 known bug patterns. See `tools/fecim-skills/_shared/fecim-context.md` for the test matrix and bug list.

## Workflow

1. **Pick scope** by the change being verified:
   - Whole suite: `go test ./...`
   - Race detection: `go test -race ./...`
   - Module: `go test ./module1-hysteresis/...` (or `make test-hys`, `test-xbar`, `test-mnist`, `test-circuits`, `test-shared`)
   - Future shell: `make test-next-ui`
   - Coverage: `go test -cover ./...`

2. **For physics-regression failures**, classify against the 5 known patterns:
   1. Guard-band sign flip
   2. Bounds collapse `[VMin, VMax]`
   3. ACCEPT ±1 guard interaction
   4. Zero-field bounds reset
   5. Preisach Everett zero-clamp

3. **Golden regeneration** is allowed only when divergence is intentional and the user has confirmed:
   ```bash
   FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...
   ```
   Diff `validation/testdata/physics_regression/` before committing — unintended changes are bugs.

4. **Output the TDD evidence block** per `_shared/tdd-evidence-template.md`:
   ```
   RED:  go test ./module1-hysteresis/... -run TestX
         FAIL TestX (Preisach Everett zero-clamp pattern)
   GREEN: go test ./module1-hysteresis/... -run TestX
          ok
   VERIFY: go test ./... && go test -race ./...
   ```

## Verification

- Input: "TestPreisachEverett is failing in module1-hysteresis."
  Expected: maps to known bug pattern #5; suggests product-form Everett vs factorized; runs the targeted test, then full suite, then race.

## TDD

This skill is the TDD verifier itself — every change made under it must produce the RED/GREEN block above before commit.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-labtester/ .claude/skills/fecim-labtester/ .codex/AGENTS.md .opencode/command/fecim-labtester.md
git commit -m "feat(skills): add fecim-labtester

TDD: N/A (configuration only)"
```

---

## Task 19: `fecim-grill`

**Files:**
- Create: `tools/fecim-skills/fecim-grill/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-grill
description: Relentlessly interviews the user about a proposed FeCIM physics, simulation, or GUI change before any code is written — covering source citation, educational-vs-validated framing, TDD-RED test, thread-safety, and honesty-audit alignment. Use when starting a non-trivial change to physics, ISPP, crossbar, or GUI logic.
---

# fecim-grill

Domain-specific design grill. Differs from generic grill skills: every branch ties back to FeCIM's TDD hard-rule, honesty-audit, and Fyne thread-safety. See `tools/fecim-skills/_shared/fecim-context.md`.

## Workflow

Ask each question one at a time. Do not move on until the user answers. Mark each branch RESOLVED before exit.

1. **What changes, behaviorally?** One sentence. If the answer references a UI element, ask which goroutine/main thread.

2. **Source citation.** Is this grounded in:
   - A published source (Materlik 2015, Park 2015, Alessandri 2018, Guo 2018, HZO FTJ 2025)?
   - An educational simulation default (clearly labeled)?
   - Neither (BLOCK — must reframe before code).

3. **Educational vs validated framing.** If the change touches accuracy/efficiency numbers, run through the honesty-audit removed/unverified list (87% MNIST, 30-states-as-fact, NAND/GPU energy multipliers).

4. **TDD-RED test.** What is the focused failing test that proves the new behavior? Path + name. If the user can't name one, BLOCK until they can.

5. **Thread-safety (Fyne).** If the change runs on a goroutine and touches `*widget.*`, `*canvas.*`, `*container.*`, confirm `fyne.Do(func() { ... })` will wrap the mutation.

6. **UI boundary.** If the change touches `viewmodel/`, confirm zero `fyne.io/...` and `github.com/gogpu/ui` imports added.

7. **Output a one-paragraph design summary** the user signs off before any code is written:
   ```
   Behavior: ...
   Source: ...
   Framing: validated | educational
   RED test: <file:test>
   Thread-safety: covered | n/a
   UI-boundary: ok | n/a
   ```

## Verification

- Input: "Add ISPP guard-band relaxation that allows 4 guard pulses instead of 2."
  Expected: walks all 7 questions; surfaces bug pattern #1 (guard-band sign flip); requires RED test path; produces a sign-off summary.

## TDD

This skill exists to enforce TDD before code is written. It cannot violate TDD itself; output is observation — `TDD: N/A`.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-grill/ .claude/skills/fecim-grill/ .codex/AGENTS.md .opencode/command/fecim-grill.md
git commit -m "feat(skills): add fecim-grill

TDD: N/A (configuration only)"
```

---

## Task 20: `fecim-fyne-thread-check`

**Files:**
- Create: `tools/fecim-skills/fecim-fyne-thread-check/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-fyne-thread-check
description: Audits Go code for goroutine-to-widget access without fyne.Do(...) wrapping, the project's most common GUI freeze cause. Use when reviewing a PR that adds goroutines, async I/O, or simulation tickers in any pkg/gui/ or shell package.
---

# fecim-fyne-thread-check

Find places where a goroutine touches a Fyne widget without `fyne.Do(...)` wrapping. See `_shared/fecim-context.md` (UI thread-safety rule).

## Workflow

1. **Define audit scope.** Default: `module*/pkg/gui/`, `cmd/fecim-lattice-tools/`. Narrow to changed files for PR review:
   ```bash
   git diff --name-only main...HEAD | rg '\.go$' | rg 'pkg/gui|cmd/fecim'
   ```

2. **Find goroutine launches:**
   ```bash
   rg -nU 'go func\(' <scope>
   ```

3. **For each match, examine the body** for direct mutation of:
   - `*widget.*` (e.g., `Label.SetText`, `Button.SetText`, `Entry.SetText`, `ProgressBar.SetValue`)
   - `*canvas.*` (e.g., `canvas.Refresh`, `*canvas.Image.Image = ...`)
   - `*container.*` (`Add`, `Remove`, `Refresh`)
   - Direct field assignment to any `fyne.CanvasObject`

4. **Verify the call is wrapped:**
   - GOOD: `fyne.Do(func() { label.SetText("done") })`
   - GOOD: helper function whose body is itself wrapped
   - BAD: bare `label.SetText(...)` inside `go func()`

5. **Output a violation list:**
   ```
   <file>:<line>: <symbol>.<method>(...) inside goroutine — needs fyne.Do
     Suggested:
       fyne.Do(func() { <symbol>.<method>(...) })
   ```

6. **Cross-reference** `docs/3-develop/gui/FYNE_NOTES.md#threading-critical` for nuanced cases (animation tickers, blocking dialogs).

## Verification

- Input: PR adds `go func() { mylabel.SetText("done") }()` in `module1-hysteresis/pkg/gui/simulation.go`.
  Expected: skill flags `simulation.go:<line>: mylabel.SetText` and suggests the wrap.

## TDD

Audit is observation — `TDD: N/A`. Any code change made to fix a violation triggers the project's TDD hard-rule.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-fyne-thread-check/ .claude/skills/fecim-fyne-thread-check/ .codex/AGENTS.md .opencode/command/fecim-fyne-thread-check.md
git commit -m "feat(skills): add fecim-fyne-thread-check

TDD: N/A (configuration only)"
```

---

## Task 21: `fecim-honesty-audit`

**Files:**
- Create: `tools/fecim-skills/fecim-honesty-audit/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-honesty-audit
description: Enforces docs/4-research/honesty-audit.md policy by scanning PR diffs, READMEs, and presentation material for removed/unverified claims (87% MNIST, 30-states-as-fact, energy multipliers vs NAND/GPUs). Use before committing docs, PRs, or release notes that include accuracy or efficiency numbers.
---

# fecim-honesty-audit

Scan a diff, README, or PR description for the project's removed/unverified claim list before it lands. See `_shared/fecim-context.md` (Honesty-audit policy).

## Workflow

1. **Take input.** Either:
   - PR diff: `git diff main...HEAD -- '*.md' README.md`
   - Specific file(s).
   - Free-text the user pastes.

2. **Regex-scan for trigger phrases (case-insensitive):**
   - `\b87%?\s*MNIST\b`
   - `\b30\s*(analog\s+)?(states|levels)\b` (followed by `device|hardware|fact` → red flag; with `simulation default|configurable` → ok)
   - `vs\.?\s*(NAND|GPUs?|DRAM)` near energy/power numbers
   - `\bX×\s*(less|more)\s*(energy|power)\b`
   - Numeric percentages near `MNIST|CIFAR|accuracy|efficiency`

3. **Classify each hit** as one of:
   - `verified` — has a published citation in the project's list.
   - `educational-default` — properly labeled as a simulation default.
   - `removed-unverified` — matches the audit's removal list; must change.

4. **For `removed-unverified`, suggest the approved rephrasing:**
   - "30 analog states" → "30 analog conductance levels (configurable simulation default)"
   - "87% MNIST accuracy" → remove, OR attribute the 98.24% HZO FTJ 2025 figure with "not a FeCIM device claim".
   - Energy multipliers without source → remove or replace with literature-backed comparison.

5. **Output a structured report:**
   ```
   File:line: <claim text>
   Status: verified | educational-default | removed-unverified
   Suggested change: <wording or DELETE>
   ```

   Exit `0` if all clean, exit `1` (or print `CHANGES REQUESTED`) if any `removed-unverified`.

## Verification

- Input: a README diff with `our chip achieves 87% MNIST accuracy at 1000× lower energy than GPUs`.
  Expected: 2 hits flagged, both `removed-unverified`; suggests removal or reframing.

## TDD

Audit is observation — `TDD: N/A`. Wording changes are documentation-only and qualify under CLAUDE.md's `TDD: N/A` carve-out.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-honesty-audit/ .claude/skills/fecim-honesty-audit/ .codex/AGENTS.md .opencode/command/fecim-honesty-audit.md
git commit -m "feat(skills): add fecim-honesty-audit

TDD: N/A (configuration only)"
```

---

## Task 22: `fecim-gogpu-migrate`

**Files:**
- Create: `tools/fecim-skills/fecim-gogpu-migrate/SKILL.md`

- [ ] **Step 1: Write the canonical skill**

```markdown
---
name: fecim-gogpu-migrate
description: Migrates a Fyne tab/component to the gogpu/ui zero-CGO shell via the shared/viewmodel UI-neutral bridge. Use when porting a module from cmd/fecim-lattice-tools to cmd/fecim-lattice-tools-next, or when extracting UI-coupled logic into the viewmodel layer.
---

# fecim-gogpu-migrate

Port a Fyne tab/component to the future zero-CGO `gogpu/ui` shell. The viewmodel layer is the UI-neutral bridge — see `_shared/fecim-context.md` (UI boundary rule).

## Workflow

1. **Identify the Fyne-coupled file.** Read `module*/pkg/gui/<name>.go`. List which types touch `fyne.io/...` and which are pure data.

2. **Extract pure state and event interface into `shared/viewmodel/<name>/`:**
   - `state.go` — typed snapshot of what the UI displays.
   - `events.go` — events the UI sends back (button presses, value changes).
   - `<name>.go` — the viewmodel: pure-Go reducer over events, returns new state.

   No `fyne.io/...` or `github.com/gogpu/ui` imports allowed in `shared/viewmodel/`. Verify:
   ```bash
   grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/ && echo VIOLATION
   ```

3. **Write a `_test.go` for the viewmodel:**
   - Drive events, assert on state.
   - This is RED-first per CLAUDE.md TDD hard-rule.

4. **Reimplement the Fyne adapter** to render `state` and dispatch `events` to the viewmodel. Do not duplicate logic.

5. **Add (or stub) the gogpu/ui adapter** at `cmd/fecim-lattice-tools-next/...` rendering the same viewmodel.

6. **Verify both shells:**
   ```bash
   go test ./shared/viewmodel/... && go test ./module*/pkg/gui/... && make test-next-ui
   go build ./cmd/fecim-lattice-tools && CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
   ```

7. **Output the TDD evidence block** per `_shared/tdd-evidence-template.md`.

## Verification

- Input: "Migrate module1-hysteresis/pkg/gui/simulation.go to viewmodel."
  Expected: lists Fyne types in scope; proposes `shared/viewmodel/hysteresis_simulation/` layout; writes failing viewmodel test first.

## TDD

Full TDD applies. Behavior change ≠ moving code only — the viewmodel test must demonstrate the reducer logic before the adapter is touched.
```

- [ ] **Step 2: Install + check + commit**

```bash
make install-skills
scripts/install-fecim-skills.sh --check
git add tools/fecim-skills/fecim-gogpu-migrate/ .claude/skills/fecim-gogpu-migrate/ .codex/AGENTS.md .opencode/command/fecim-gogpu-migrate.md
git commit -m "feat(skills): add fecim-gogpu-migrate

TDD: N/A (configuration only)"
```

---

## Task 23: Phase 2 verification + PR

- [ ] **Step 1: Final sync check**

Run:
```bash
make test-skills && \
make install-skills && \
scripts/install-fecim-skills.sh --check && \
git diff --exit-code
```
Expected: tests `Passed: 8`, install reports `Installed 8 skills`, `--check` exits 0, no diff.

- [ ] **Step 2: Verify all 8 adapters exist on each harness**

Run:
```bash
ls .claude/skills/ | sort
ls .opencode/command/ | sort
grep -c "^- \*\*fecim-" .codex/AGENTS.md
```
Expected:
- `.claude/skills/`: 8 directories (`fecim-builder`, `fecim-citation`, `fecim-fyne-thread-check`, `fecim-gogpu-migrate`, `fecim-grill`, `fecim-honesty-audit`, `fecim-labtester`, `fecim-researcher`).
- `.opencode/command/`: 8 `.md` files.
- `.codex/AGENTS.md`: `8`.

- [ ] **Step 3: Open PR2**

```bash
git push -u origin <branch>
gh pr create --title "feat(skills): add remaining 7 fecim-skills" --body "$(cat <<'BODY'
## Summary

- Adds canonical SKILL.md for: fecim-researcher, fecim-citation, fecim-labtester, fecim-grill, fecim-fyne-thread-check, fecim-honesty-audit, fecim-gogpu-migrate
- Regenerates per-harness adapters via `make install-skills`

## Test plan
- [ ] `make test-skills` passes
- [ ] `scripts/install-fecim-skills.sh --check` exits 0
- [ ] CI is green
- [ ] `.claude/skills/` lists all 8 skills
- [ ] `.codex/AGENTS.md` lists all 8 in the managed block
- [ ] `.opencode/command/` lists all 8

🤖 Generated with [Claude Code](https://claude.com/claude-code)
BODY
)"
```

---

## Self-Review (Plan Author)

| Check | Result |
|---|---|
| **Spec coverage** — every section of the spec maps to a task | Catalog (T2, T16-T22), Directory layout (T1), Adapter formats (T4 install code), Install flow (T4, T11, T12), Failure modes (T5-T9), Testing (T3-T10), TDD compliance (each task), Rollout (T15 PR1, T23 PR2), Out of scope (not in plan, correct) — all covered. |
| **Placeholder scan** | No "TBD"/"TODO" remain; every code step shows the code; every test shows assertions. |
| **Type/name consistency** | Function names (`emit_claude_adapter`, `emit_opencode_adapter`, `emit_codex`, `claude_copy_body`, `opencode_body`, `skill_name`, `skill_description`, `note_drift`) used consistently across Tasks 4 and 9. Env var `FECIM_FORCE_NO_SYMLINK` used consistently in Tasks 9–10. |
| **Frontmatter `description:` fields** | Each Phase 2 SKILL.md `description:` matches the spec catalog text verbatim. |

No issues found that require restructure.
