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
