from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path
import hashlib
import json
import re


CLAIM_REF_RE = re.compile(r"\[claim:\s*([a-z0-9][a-z0-9-]*)\]")
PAPER_PDF_RE = re.compile(r"^\*\*PDF:\*\*\s*`([^`]+)`", re.MULTILINE)
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

    _audit_citation_pdf_paths(root, errors)
    _audit_source_ledgers(root, errors)
    _audit_evidence_ledgers(root, claims, errors)

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


def _audit_citation_pdf_paths(root: Path, errors: list[str]) -> None:
    papers_dir = root / "citations" / "papers"
    if not papers_dir.exists():
        return
    for path in sorted(papers_dir.glob("*.md")):
        rel_path = _rel(root, path)
        pdf_path = _citation_pdf_path(path)
        if not pdf_path or pdf_path.lower() == "not stored":
            continue
        if not _is_repo_relative_path(pdf_path):
            errors.append(f"{rel_path} PDF path must be repo-relative")
            continue
        if _is_ignored_pdf_inbox_path(pdf_path):
            errors.append(
                f"{rel_path} PDF path {pdf_path} points at ignored local inbox; "
                "use not stored until promoted"
            )
            continue
        if not (root / pdf_path).is_file():
            errors.append(f"{rel_path} PDF path {pdf_path} does not exist")


def _citation_pdf_path(path: Path) -> str:
    match = PAPER_PDF_RE.search(path.read_text(encoding="utf-8", errors="replace"))
    if match is None:
        return ""
    return match.group(1).strip()


def _is_ignored_pdf_inbox_path(path: str) -> bool:
    return path == "research/papers" or path.startswith("research/papers/")


def _audit_source_ledgers(root: Path, errors: list[str]) -> None:
    sources_dir = root / "research" / "sources"
    if not sources_dir.exists():
        return
    for path in sorted(sources_dir.glob("*.yaml")):
        if path.name.endswith(".acquisition.yaml"):
            continue
        rel_path = _rel(root, path)
        data = _parse_mapping_yaml(path)
        paper_key = str(data.get("paper_key", "")).strip()
        if not paper_key:
            errors.append(f"{rel_path} missing paper_key")
        elif paper_key != path.stem:
            errors.append(f"{rel_path} paper_key {paper_key} must match filename {path.stem}")

        _audit_source_file_reference(root, rel_path, "citation_path", data.get("citation_path"), errors)

        pdf = data.get("pdf")
        if not isinstance(pdf, dict):
            errors.append(f"{rel_path} missing pdf metadata")
            continue
        pdf_file = _audit_source_file_reference(root, rel_path, "pdf path", pdf.get("path"), errors)
        expected_sha = str(pdf.get("sha256", "")).strip()
        if not expected_sha:
            errors.append(f"{rel_path} missing pdf sha256")
        elif pdf_file is not None:
            actual_sha = _sha256_file(pdf_file)
            if expected_sha != actual_sha:
                errors.append(f"{rel_path} pdf sha256 {expected_sha} does not match actual {actual_sha}")


def _audit_source_file_reference(
    root: Path,
    owner: str,
    label: str,
    value: object,
    errors: list[str],
) -> Path | None:
    rel_path = str(value or "").strip()
    if not rel_path:
        errors.append(f"{owner} missing {label}")
        return None
    if not _is_repo_relative_path(rel_path):
        errors.append(f"{owner} {label} must be repo-relative")
        return None
    full_path = root / rel_path
    if not full_path.is_file():
        errors.append(f"{owner} {label} {rel_path} does not exist")
        return None
    return full_path


def _audit_evidence_ledgers(
    root: Path,
    claims: dict[str, ClaimRecord],
    errors: list[str],
) -> None:
    evidence_dir = root / "research" / "evidence"
    if not evidence_dir.exists():
        return
    for path in sorted(evidence_dir.glob("*.json")):
        rel_path = _rel(root, path)
        try:
            data = json.loads(path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError) as exc:
            errors.append(f"{rel_path} invalid JSON: {exc}")
            continue
        if not isinstance(data, dict):
            errors.append(f"{rel_path} must contain a JSON object")
            continue

        claim_id = _evidence_claim_id(data)
        if not claim_id:
            errors.append(f"{rel_path} missing claim id")
        else:
            if claim_id != path.stem:
                errors.append(f"{rel_path} claim id {claim_id} must match filename {path.stem}")
            if claim_id not in claims:
                errors.append(f"{rel_path} references unknown claim {claim_id}")

        if data.get("status") != "candidate-evidence":
            errors.append(f"{rel_path} status must be candidate-evidence")

        review = data.get("review")
        if not isinstance(review, dict) or review.get("state") != "needs-review":
            errors.append(f"{rel_path} review.state must be needs-review")

        candidates = data.get("candidates", [])
        if not isinstance(candidates, list):
            errors.append(f"{rel_path} candidates must be a list")
            continue

        candidate_count = data.get("candidate_count")
        if candidate_count != len(candidates):
            errors.append(
                f"{rel_path} candidate_count {candidate_count} "
                f"does not match candidates length {len(candidates)}"
            )

        for index, candidate in enumerate(candidates):
            if not isinstance(candidate, dict):
                errors.append(f"{rel_path} candidate {index} must be a JSON object")
                continue
            _audit_evidence_candidate(root, rel_path, candidate, errors)


def _evidence_claim_id(data: dict[str, object]) -> str:
    claim = data.get("claim")
    if not isinstance(claim, dict):
        return ""
    return str(claim.get("id", "")).strip()


def _audit_evidence_candidate(
    root: Path,
    evidence_rel_path: str,
    candidate: dict[str, object],
    errors: list[str],
) -> None:
    docid = str(candidate.get("docid", "")).strip()
    chunk_file = str(candidate.get("chunk_file", "")).strip()
    if not docid:
        errors.append(f"{evidence_rel_path} candidate missing docid")
        return
    if not chunk_file:
        errors.append(f"{evidence_rel_path} candidate {docid} missing chunk_file")
        return

    chunk_rel_path = Path(chunk_file)
    if chunk_rel_path.is_absolute() or ".." in chunk_rel_path.parts:
        errors.append(f"{evidence_rel_path} candidate {docid} chunk_file must be repo-relative")
        return

    chunk_path = root / chunk_rel_path
    if not chunk_path.exists():
        errors.append(f"{evidence_rel_path} candidate {docid} missing chunk file {chunk_file}")
        return

    record = _find_chunk_record(chunk_path, docid)
    if record is None:
        errors.append(f"{evidence_rel_path} candidate {docid} missing from {chunk_file}")
        return

    candidate_sha = str(candidate.get("sha256", "")).strip()
    chunk_sha = str(record.get("sha256", "")).strip()
    if candidate_sha and chunk_sha and candidate_sha != chunk_sha:
        errors.append(
            f"{evidence_rel_path} candidate {docid} sha256 {candidate_sha} "
            f"does not match chunk sha256 {chunk_sha}"
        )


def _find_chunk_record(path: Path, docid: str) -> dict[str, object] | None:
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError:
        return None
    for line in lines:
        if not line.strip():
            continue
        try:
            data = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(data, dict) and data.get("id") == docid:
            return data
    return None


def _parse_mapping_yaml(path: Path) -> dict[str, object]:
    root: dict[str, object] = {}
    stack: list[tuple[int, dict[str, object]]] = [(-1, root)]
    for raw_line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue
        if ":" not in raw_line:
            continue
        indent = len(raw_line) - len(raw_line.lstrip())
        stripped = raw_line.strip()
        key, value = stripped.split(":", 1)
        while stack and indent <= stack[-1][0]:
            stack.pop()
        parent = stack[-1][1] if stack else root
        value = value.strip()
        if value == "":
            child: dict[str, object] = {}
            parent[key.strip()] = child
            stack.append((indent, child))
        else:
            parent[key.strip()] = _unquote(value)
    return root


def _is_repo_relative_path(value: str) -> bool:
    path = Path(value)
    return not path.is_absolute() and ".." not in path.parts


def _sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def _write_report(root: Path, report: ClaimAuditReport) -> None:
    path = root / "research" / "reports" / "claim-audit-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(asdict(report), indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
