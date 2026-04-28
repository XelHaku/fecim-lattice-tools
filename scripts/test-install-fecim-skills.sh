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
