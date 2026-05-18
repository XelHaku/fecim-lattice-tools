from __future__ import annotations

import json
from pathlib import Path
from typing import Any


SOURCE_REPORT = "research/reports/search-latest.json"


def build_evidence_record(root: Path, claim_id: str) -> dict[str, object] | None:
    report = _read_search_report(root)
    if report is None:
        return None
    claim = report.get("claim")
    if not isinstance(claim, dict) or claim.get("id") != claim_id:
        return None
    if _uses_unreviewed_inbox_results(report):
        return None

    results = report.get("results", [])
    if not isinstance(results, list):
        results = []

    candidates = [_candidate(row) for row in results if isinstance(row, dict)]
    return {
        "claim": claim,
        "status": "candidate-evidence",
        "review": {
            "state": "needs-review",
            "notes": "Candidate evidence from retrieval; verify support before promotion.",
        },
        "source_report": SOURCE_REPORT,
        "query": str(report.get("query", "")),
        "backend": str(report.get("backend", "")),
        "candidate_count": len(candidates),
        "candidates": candidates,
    }


def run_evidence(root: Path, claim_id: str) -> int:
    record = build_evidence_record(root, claim_id)
    if record is None:
        print(
            f"no claim-linked search report found for {claim_id}; "
            f"run `fecim research search --claim {claim_id} --local` first"
        )
        return 1

    output = root / "research" / "evidence" / f"{claim_id}.json"
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(record, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    summary = {
        "ok": True,
        "claim_id": claim_id,
        "candidate_count": record["candidate_count"],
        "source_report": SOURCE_REPORT,
        "output": _repo_relative(root, output),
    }
    report = root / "research" / "reports" / "evidence-latest.json"
    report.parent.mkdir(parents=True, exist_ok=True)
    report.write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    print(f"research evidence complete: claim={claim_id} candidates={record['candidate_count']}")
    return 0


def _candidate(row: dict[str, Any]) -> dict[str, object]:
    keys = [
        "rank",
        "score",
        "docid",
        "paper_key",
        "section",
        "snippet",
        "chunk_file",
        "source_path",
        "source_parser",
        "page_start",
        "page_end",
        "char_start",
        "char_end",
        "sha256",
    ]
    return {key: row.get(key, "") for key in keys if key in row}


def _read_search_report(root: Path) -> dict[str, object] | None:
    path = root / SOURCE_REPORT
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return None
    return data if isinstance(data, dict) else None


def _uses_unreviewed_inbox_results(report: dict[str, object]) -> bool:
    backend = report.get("backend", "")
    if isinstance(backend, str) and backend.startswith("inbox"):
        return True
    if report.get("trust_state") == "unreviewed" or report.get("review_required") is True:
        return True

    results = report.get("results", [])
    if not isinstance(results, list):
        return False
    for row in results:
        if not isinstance(row, dict):
            continue
        if row.get("trust_state") == "unreviewed" or row.get("review_required") is True or row.get("inbox") is True:
            return True
    return False


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()
