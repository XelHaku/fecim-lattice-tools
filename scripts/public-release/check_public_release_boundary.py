#!/usr/bin/env python3

from __future__ import annotations

import csv
import re
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
AUDIT_PATH = ROOT / "docs/public-release/THIRD_PARTY_PDF_AUDIT.csv"
DISALLOWED_PATHS = (
    "docs/archive/",
    "docs/superpowers/",
    "docs/4-research/internal-analysis/",
    "docs/4-research/transcripts/COSM_2025_AI_Hardware_Breakthrough/",
    "docs/4-research/transcripts/ironlattice-youtube-script.md",
    "docs/4-research/tour-group-ironlattice-research.md",
    "docs/4-research/superlattice-material-analysis.md",
    "docs/research-papers/external-research/",
    "docs/4-research/papers/external-research/",
    "docs/4-research/papers/DOWNLOAD_PLAN.md",
    "docs/4-research/opensource-tools/research_notes_final.md",
    "demo2-crossbar/pkg/_layers_experimental/kvcache_fjh.go",
    "module2-crossbar/pkg/_layers_experimental/kvcache_fjh.go",
    "validation/literature/_incoming/",
)
GENERATED_PATH_PREFIXES = (
    "artifacts/",
    "exports/",
    "output/",
    "recordings/",
    "screenshots/",
    "docs/1-getting-started/demo-videos/",
    "validation/output/",
    "validation/literature/output/",
    "shared/physics/output/",
    "module1-hysteresis/pkg/controller/output/",
    "module4-circuits/pkg/arraysim/output/",
    "module4-circuits/pkg/gui/output/",
)
GENERATED_EXACT_PATHS = (
    "crossbar",
    "hysteresis",
    "gen_golden_loops",
    "module1-hysteresis/hysteresis",
    "module1-hysteresis/test_output.txt",
    "module1-hysteresis/test_output_final.txt",
    "module2-crossbar/inference",
    "module3-mnist/mnist",
)
BAN_RE = re.compile(
    r"restricted|under nda|\bnda\b|internal repo\s*=|internal draft|"
    r"internal draft|research planning|research planning|public summary candidate|"
    r"james\s+tour|external-research|rice university|investor|technical briefing|technical briefing|"
    r"active benchmark domain|scenario modeling",
    re.IGNORECASE,
)
SCAN_ROOTS = (
    "CLAUDE.md",
    "README.md",
    "docs/2-learn",
    "docs/3-develop",
    "docs/4-research",
    "module5-comparison",
    "module6-eda",
)


def tracked_files() -> list[str]:
    result = subprocess.run(
        ["git", "ls-files"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(result.stderr.strip() or "git ls-files failed")
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]


def tracked_pdfs() -> list[str]:
    result = subprocess.run(
        ["git", "ls-files", "*.pdf"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(result.stderr.strip() or "git ls-files failed")
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]


def load_pdf_decisions() -> dict[str, str]:
    if not AUDIT_PATH.exists():
        return {}
    with AUDIT_PATH.open(newline="", encoding="utf-8") as fh:
        reader = csv.DictReader(fh)
        decisions = {}
        for row in reader:
            path = (row.get("path") or "").strip()
            if path:
                decisions[path] = (row.get("decision") or "").strip()
        return decisions


def is_generated_output(path: str) -> bool:
    if path in GENERATED_EXACT_PATHS:
        return True
    if path.startswith(GENERATED_PATH_PREFIXES):
        return True
    if path.startswith("logs/"):
        return True
    if "/logs/" in path:
        return True
    if path.endswith(".log"):
        return True
    if path.endswith(".mp4"):
        return True
    if re.match(r"\d{4}-\d{2}-\d{2}.*\.png$", path):
        return True
    return False


def main() -> int:
    failures: list[str] = []
    tracked = tracked_files()

    for path in tracked:
        if any(path.startswith(disallowed) for disallowed in DISALLOWED_PATHS):
            failures.append(f"Blocked tracked path: {path}")
        if is_generated_output(path):
            failures.append(f"Tracked generated output: {path}")

    pdf_decisions = load_pdf_decisions()
    for path in tracked_pdfs():
        decision = pdf_decisions.get(path)
        if decision is None:
            failures.append(f"Missing PDF audit row: {path}")
            continue
        if decision not in {"keep", "keep-with-conditions"}:
            failures.append(f"Blocked PDF decision: {path} -> {decision or '<missing>'}")

    scan = subprocess.run(
        ["rg", "-n", "-i", BAN_RE.pattern, *SCAN_ROOTS],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if scan.returncode == 0:
        failures.append("Blocked phrases found:\n" + scan.stdout.strip())
    elif scan.returncode != 1:
        failures.append(f"rg failed with exit code {scan.returncode}:\n{scan.stderr.strip()}")

    if failures:
        for failure in failures:
            print(failure)
        return 1

    print("Public release boundary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
