from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path
import json
import re


CLAIM_REF_RE = re.compile(r"\[claim:\s*([a-z0-9][a-z0-9-]*)\]")
ALLOWED_STATUS = {
    "literature-backed",
    "validation-backed",
    "educational",
    "planned",
    "disputed",
    "not-validated",
}
ALLOWED_CONFIDENCE = {"low", "medium", "high"}


@dataclass(frozen=True)
class ClaimRecord:
    id: str
    path: Path
    claim: str
    status: str
    sources: list[str]
    used_in: list[str]
    confidence: str


@dataclass(frozen=True)
class ClaimAuditReport:
    ok: bool
    claims_checked: int
    errors: list[str]
    warnings: list[str]


def run_audit(root: Path) -> int:
    report = audit_claim_registry(root)
    _write_report(root, report)
    if report.ok:
        print(f"research audit complete: claims={report.claims_checked} errors=0")
        return 0
    print(f"research audit failed: claims={report.claims_checked} errors={len(report.errors)}")
    for error in report.errors:
        print(f"- {error}")
    return 1


def audit_claim_registry(root: Path) -> ClaimAuditReport:
    errors: list[str] = []
    warnings: list[str] = []
    claims = load_claim_records(root, errors)
    source_keys = _source_keys(root)

    for claim_id, record in sorted(claims.items()):
        if record.id != claim_id:
            errors.append(f"{_rel(root, record.path)} id {record.id} must match filename {claim_id}")
        if record.status not in ALLOWED_STATUS:
            errors.append(f"{_rel(root, record.path)} has invalid status {record.status}")
        if record.confidence not in ALLOWED_CONFIDENCE:
            errors.append(f"{_rel(root, record.path)} has invalid confidence {record.confidence}")
        if not record.claim:
            errors.append(f"{_rel(root, record.path)} missing claim text")
        if not record.sources:
            errors.append(f"{_rel(root, record.path)} must list at least one source")
        if not record.used_in:
            errors.append(f"{_rel(root, record.path)} must list at least one used_in path")
        for source in record.sources:
            if source not in source_keys:
                errors.append(f"{_rel(root, record.path)} missing source {source}")
        for used_path in record.used_in:
            full_path = root / used_path
            if not full_path.exists():
                errors.append(f"{_rel(root, record.path)} missing used_in path {used_path}")
            elif f"[claim: {claim_id}]" not in full_path.read_text(encoding="utf-8", errors="replace"):
                errors.append(f"{used_path} does not reference [claim: {claim_id}]")

    for rel_path in _claim_reference_files(root):
        path = root / rel_path
        if not path.exists():
            continue
        for claim_id in _claim_refs(path):
            record = claims.get(claim_id)
            if record is None:
                errors.append(f"{rel_path} references unknown claim id {claim_id}")
                continue
            if rel_path == "citations/facts.md" and record.status == "disputed":
                errors.append(f"disputed claim {claim_id} is referenced from citations/facts.md")

    return ClaimAuditReport(
        ok=not errors,
        claims_checked=len(claims),
        errors=errors,
        warnings=warnings,
    )


def load_claim_records(root: Path, errors: list[str] | None = None) -> dict[str, ClaimRecord]:
    claims_dir = root / "citations" / "claims"
    records: dict[str, ClaimRecord] = {}
    if not claims_dir.exists():
        return records
    for path in sorted(claims_dir.glob("*.yaml")):
        data = _parse_claim_yaml(path)
        claim_id = str(data.get("id", "")).strip()
        if not claim_id:
            claim_id = path.stem
            if errors is not None:
                errors.append(f"{_rel(root, path)} missing id")
        records[path.stem] = ClaimRecord(
            id=claim_id,
            path=path,
            claim=str(data.get("claim", "")).strip(),
            status=str(data.get("status", "")).strip(),
            sources=list(data.get("sources", [])),
            used_in=list(data.get("used_in", [])),
            confidence=str(data.get("confidence", "")).strip(),
        )
    return records


def _parse_claim_yaml(path: Path) -> dict[str, object]:
    data: dict[str, object] = {}
    current_list: str | None = None
    for raw_line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        line = raw_line.rstrip()
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if stripped.startswith("- "):
            if current_list is not None:
                casted = data.setdefault(current_list, [])
                if isinstance(casted, list):
                    casted.append(_unquote(stripped[2:].strip()))
            continue
        current_list = None
        if ":" not in stripped:
            continue
        key, value = stripped.split(":", 1)
        key = key.strip()
        value = value.strip()
        if value == "":
            data[key] = []
            current_list = key
        else:
            data[key] = _unquote(value)
    return data


def _unquote(value: str) -> str:
    if len(value) >= 2 and value[0] == value[-1] == '"':
        return value[1:-1].replace('\\"', '"')
    if len(value) >= 2 and value[0] == value[-1] == "'":
        return value[1:-1]
    return value


def _source_keys(root: Path) -> set[str]:
    keys: set[str] = set()
    for path in (root / "citations" / "papers").glob("*.md"):
        if path.name != ".gitkeep":
            keys.add(path.stem)
    for path in (root / "research" / "sources").glob("*.openalex.json"):
        keys.add(path.name.removesuffix(".openalex.json"))
    return keys


def _claim_reference_files(root: Path) -> list[str]:
    paths = ["citations/facts.md", "citations/disputed.md", "docs/TRUST.md"]
    paths.extend(str(path.relative_to(root)) for path in sorted((root / "config").glob("*.yaml")))
    return paths


def _claim_refs(path: Path) -> list[str]:
    return CLAIM_REF_RE.findall(path.read_text(encoding="utf-8", errors="replace"))


def _write_report(root: Path, report: ClaimAuditReport) -> None:
    path = root / "research" / "reports" / "claim-audit-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(asdict(report), indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
