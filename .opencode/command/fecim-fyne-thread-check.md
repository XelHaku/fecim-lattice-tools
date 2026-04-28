---
description: Audits Go code for goroutine-to-widget access without fyne.Do(...) wrapping, the project's most common GUI freeze cause. Use when reviewing a PR that adds goroutines, async I/O, or simulation tickers in any pkg/gui/ or shell package.
agent: build
---

<!-- generated-from: tools/fecim-skills/fecim-fyne-thread-check/SKILL.md -->
<!-- do not edit; run scripts/install-fecim-skills.sh -->

Read and follow the workflow in `tools/fecim-skills/fecim-fyne-thread-check/SKILL.md`.

Use the user's request as $ARGUMENTS.
