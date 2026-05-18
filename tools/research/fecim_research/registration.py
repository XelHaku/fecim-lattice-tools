from __future__ import annotations

from pathlib import Path
import json
import re

from .citations import CitationRecord, load_citation_records
from .discovery import DiscoveredPDF, discover_pdfs, match_pdf_to_record


def run_register_pdfs(root: Path, extra_paths: list[Path], write_stubs: bool) -> int:
    records = load_citation_records(root)
    pdfs = discover_pdfs(root, extra_paths)
    rows = _registration_rows(root, pdfs, records)
    stubs_written = 0
    used_keys = set(records)
    for row in rows:
        if row["status"] != "unmatched":
            continue
        key = str(row["paper_key"])
        if key in used_keys:
            key = f"{key}_{str(row['sha256'])[:8]}"
            row["paper_key"] = key
            row["stub_path"] = f"citations/papers/{key}.md"
        used_keys.add(key)
        if write_stubs:
            _write_stub(root, row)
            row["action"] = "stub_written"
            stubs_written += 1
        else:
            row["action"] = "stub_available"

    report = {
        "discovered": len(rows),
        "matched": sum(1 for row in rows if row["status"] == "matched"),
        "unmatched": sum(1 for row in rows if row["status"] == "unmatched"),
        "duplicates": sum(1 for row in rows if row["status"] == "duplicate"),
        "stubs_planned": sum(1 for row in rows if row["status"] == "unmatched"),
        "stubs_written": stubs_written,
        "write_stubs": write_stubs,
        "extra_paths": [_path_text(path) for path in extra_paths],
        "pdfs": rows,
    }
    report_path = root / "research" / "reports" / "pdf-registration-latest.json"
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    _refresh_missing_report(root)
    print(
        "pdf registration complete: "
        f"discovered={report['discovered']} unmatched={report['unmatched']} stubs_written={stubs_written}"
    )
    return 0


def _registration_rows(
    root: Path,
    pdfs: list[DiscoveredPDF],
    records: dict[str, CitationRecord],
) -> list[dict[str, object]]:
    rows: list[dict[str, object]] = []
    for pdf in pdfs:
        rel_path = _rel(root, pdf.path)
        if pdf.duplicate_of is not None:
            rows.append(
                {
                    "path": rel_path,
                    "sha256": pdf.sha256,
                    "size": pdf.size,
                    "status": "duplicate",
                    "duplicate_of": _rel(root, pdf.duplicate_of),
                    "paper_key": "",
                    "stub_path": "",
                    "action": "none",
                }
            )
            continue

        match = match_pdf_to_record(pdf, records)
        if match.paper_key is not None:
            record = records[match.paper_key]
            rows.append(
                {
                    "path": rel_path,
                    "sha256": pdf.sha256,
                    "size": pdf.size,
                    "status": "matched",
                    "paper_key": match.paper_key,
                    "citation_path": _rel(root, record.path),
                    "match_method": match.method,
                    "match_confidence": match.confidence,
                    "stub_path": "",
                    "action": "none",
                }
            )
            continue

        key = _key_from_pdf(pdf)
        rows.append(
            {
                "path": rel_path,
                "sha256": pdf.sha256,
                "size": pdf.size,
                "status": "unmatched",
                "paper_key": key,
                "stub_path": f"citations/papers/{key}.md",
                "action": "stub_available",
            }
        )
    return rows


def _write_stub(root: Path, row: dict[str, object]) -> None:
    key = str(row["paper_key"])
    path = root / str(row["stub_path"])
    path.parent.mkdir(parents=True, exist_ok=True)
    title = _title_from_key(key)
    pdf_path = str(row["path"])
    canonical_pdf = "not stored" if _is_ignored_pdf_inbox_path(pdf_path) else pdf_path
    text = (
        f"# {title}\n\n"
        f"**Key:** `{key}`\n"
        f"**Title:** `{title}`\n"
        "**DOI:** `needs-review`\n"
        "**Year:** `needs-review`\n"
        "**Venue:** `needs-review`\n"
        "**Authors:** `needs-review`\n"
        "**Tags:** `#needs-review`\n"
        "**Status:** `needs-review`\n"
        f"**PDF:** `{canonical_pdf}`\n"
        f"**Local PDF:** `{pdf_path}`\n"
        f"**SHA256:** `{row['sha256']}`\n"
        f"**Size:** `{row['size']}`\n"
        "\n---\n\n"
        "## Review TODO\n\n"
        "- [ ] Confirm bibliographic metadata.\n"
        "- [ ] Add DOI or arXiv ID when available.\n"
        "- [ ] Add source notes only after reading the paper.\n"
        "- [ ] Move evidence-bearing claims into `citations/claims/*.yaml`.\n"
    )
    path.write_text(text, encoding="utf-8")


def _key_from_pdf(pdf: DiscoveredPDF) -> str:
    key = re.sub(r"[^a-z0-9]+", "_", pdf.path.stem.lower()).strip("_")
    if not key:
        key = "paper"
    return key[:96]


def _title_from_key(key: str) -> str:
    return " ".join(part.upper() if part in {"ai", "hzo", "hfo2", "cim"} else part.capitalize() for part in key.split("_"))


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)


def _path_text(path: Path) -> str:
    return path.as_posix()


def _is_ignored_pdf_inbox_path(path: str) -> bool:
    return path == "research/papers" or path.startswith("research/papers/")


def _refresh_missing_report(root: Path) -> None:
    from .missing import run_missing

    run_missing(root=root)
