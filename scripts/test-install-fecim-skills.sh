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

run_test "codex install preserves user content outside markers" '
  fixture_one_skill
  mkdir -p .codex
  printf "Pre-existing line above\n\nAnother user line\n" > .codex/AGENTS.md
  ./install.sh
  grep -q "Pre-existing line above" .codex/AGENTS.md
  grep -q "Another user line" .codex/AGENTS.md
  grep -q "fecim-skills:start" .codex/AGENTS.md
'

run_test "--check exits non-zero when generated adapter drifts" '
  fixture_one_skill
  ./install.sh
  echo "tampered" >> .opencode/command/sample-skill.md
  ! ./install.sh --check
'

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

run_test "symlink failure falls back to copy with managed-block header" '
  fixture_one_skill
  FECIM_FORCE_NO_SYMLINK=1 ./install.sh
  test -f .claude/skills/sample-skill/SKILL.md
  test ! -L .claude/skills/sample-skill/SKILL.md
  head -1 .claude/skills/sample-skill/SKILL.md | grep -q "generated-from"
'

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

# ---------- Summary ----------
echo
echo "Passed: $PASS  Failed: $FAIL"
if (( FAIL > 0 )); then
  for d in "${FAIL_DETAILS[@]}"; do echo "  - $d"; done
  exit 1
fi
