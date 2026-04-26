#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

bash scripts/citations/_run_agent.sh prompts/citations/04_auditor.md "Audit repository citation coverage and write citations/reports/citation_check.md."
