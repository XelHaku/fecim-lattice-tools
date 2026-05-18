from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path
import json
import re


SCANNABLE_SUFFIXES = {".md", ".yaml", ".yml"}
DEFAULT_SCAN_PATHS = ["README.md", "docs/TRUST.md", "citations/facts.md", "config"]
CLAIM_REF_RE = re.compile(r"\[claim:\s*[a-z0-9][a-z0-9-]*\]")
DOI_RE = re.compile(r"\b10\.\d{4,9}/[-._;()/:A-Za-z0-9]+")
NUMERIC_UNIT_RE = re.compile(
    r"\b\d+(?:\.\d+)?(?:e[+-]?\d+|\^\d+)?\s*"
    r"(?:%|uC/cm2|µC/cm²|MV/cm|V/m|C/m²|C/m2|TOPS/W|GOPS/mm²|GOPS/mm2|cycles?|nm|ps|ns|K)\b",
    re.IGNORECASE,
)


@dataclass(frozen=True)
class ClaimScanFinding:
    path: str
    line: int
    reason: str
    text: str


@dataclass(frozen=True)
class ClaimScanReport:
    scanned_files: int
    findings_count: int
    findings: list[ClaimScanFinding]


def run_claim_scan(root: Path, paths: list[str], fail_on_findings: bool) -> int:
    report = scan_claims(root, paths)
    _write_report(root, report)
    print(f"claim scan complete: files={report.scanned_files} findings={report.findings_count}")
    return 1 if fail_on_findings and report.findings else 0


def scan_claims(root: Path, paths: list[str]) -> ClaimScanReport:
    files = _candidate_files(root, paths or DEFAULT_SCAN_PATHS)
    findings: list[ClaimScanFinding] = []
    for path in files:
        findings.extend(_scan_file(root, path))
    return ClaimScanReport(
        scanned_files=len(files),
        findings_count=len(findings),
        findings=findings,
    )


def _candidate_files(root: Path, paths: list[str]) -> list[Path]:
    seen: set[Path] = set()
    for item in paths:
        path = Path(item)
        if not path.is_absolute():
            path = root / path
        if path.is_file() and _is_scannable(path):
            seen.add(path.resolve())
        elif path.is_dir():
            for child in path.rglob("*"):
                if child.is_file() and _is_scannable(child):
                    seen.add(child.resolve())
    return sorted(seen, key=lambda path: _rel(root, path))


def _is_scannable(path: Path) -> bool:
    return path.suffix.lower() in SCANNABLE_SUFFIXES


def _scan_file(root: Path, path: Path) -> list[ClaimScanFinding]:
    findings: list[ClaimScanFinding] = []
    in_code_fence = False
    covered_indent: int | None = None
    for line_no, raw_line in enumerate(path.read_text(encoding="utf-8", errors="replace").splitlines(), start=1):
        stripped = raw_line.strip()
        if stripped.startswith("```") or stripped.startswith("~~~"):
            in_code_fence = not in_code_fence
            continue
        if in_code_fence or not stripped:
            continue
        indent = len(raw_line) - len(raw_line.lstrip())
        if covered_indent is not None and indent > covered_indent:
            continue
        covered_indent = None
        if "claim-scan: ignore" in stripped:
            continue
        if CLAIM_REF_RE.search(stripped):
            covered_indent = indent
            continue
        reason = _finding_reason(stripped)
        if reason:
            findings.append(
                ClaimScanFinding(
                    path=_rel(root, path),
                    line=line_no,
                    reason=reason,
                    text=stripped,
                )
            )
    return findings


def _finding_reason(line: str) -> str:
    normalized = line.replace("`", "").replace("*", "")
    if DOI_RE.search(normalized):
        return "doi"
    if NUMERIC_UNIT_RE.search(normalized):
        return "numeric-unit"
    return ""


def _write_report(root: Path, report: ClaimScanReport) -> None:
    path = root / "research" / "reports" / "claim-scan-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(asdict(report), indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
