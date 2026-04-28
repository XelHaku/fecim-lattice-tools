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
