from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path
import hashlib
import json
import re

from .reporting import write_content_addressed_report


CLAIM_REF_RE = re.compile(r"\[claim:\s*([a-z0-9][a-z0-9-]*)\]")
PAPER_PDF_RE = re.compile(r"^\*\*PDF:\*\*\s*`([^`]+)`", re.MULTILINE)
CITATION_FIELD_RE = re.compile(r"^\*\*(?P<name>[^*]+):\*\*\s*`?(?P<value>[^`\n]+)`?", re.MULTILINE)
PDF_REVIEW_BACKLOG_PATH = Path("research/manifests/pdf-review-backlog.json")
PDF_REVIEW_BACKLOG_SCHEMA = "fecim.pdf-review-backlog.v1"
PDF_REVIEW_BACKLOG_STATUS = "legacy-needs-license-review"
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
    pdf_review_backlog = _load_pdf_review_backlog(root, errors)

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

    _audit_citation_pdf_paths(root, errors, pdf_review_backlog)
    _audit_citation_record_identity(root, errors)
    _audit_source_ledgers(root, errors)
    _audit_openalex_ledgers(root, errors)
    _audit_acquisition_ledgers(root, errors)
    _audit_promotion_ledgers(root, errors)
    _audit_pdf_review_backlog(root, pdf_review_backlog, errors)
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


def _audit_citation_pdf_paths(
    root: Path,
    errors: list[str],
    pdf_review_backlog: dict[str, dict[str, object]],
) -> None:
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
        pdf_file = root / pdf_path
        if not pdf_file.is_file():
            errors.append(f"{rel_path} PDF path {pdf_path} does not exist")
            continue
        fields = _parse_markdown_fields(path)
        expected_sha = fields.get("sha256", "").strip()
        if expected_sha:
            actual_sha = _sha256_file(pdf_file)
            if expected_sha != actual_sha:
                errors.append(
                    f"{rel_path} SHA256 {expected_sha} does not match actual {actual_sha}"
                )
        expected_size = fields.get("size", "").strip()
        if expected_size:
            try:
                size = int(expected_size)
            except ValueError:
                errors.append(f"{rel_path} Size must be an integer byte count")
            else:
                actual_size = pdf_file.stat().st_size
                if size != actual_size:
                    errors.append(f"{rel_path} Size {size} does not match actual {actual_size}")
        if not _has_promotion_ledger(root, path.stem):
            backlog_entry = pdf_review_backlog.get(path.stem)
            if not backlog_entry or str(backlog_entry.get("pdf_path", "")).strip() != pdf_path:
                errors.append(
                    f"{rel_path} PDF path {pdf_path} "
                    "has no promotion ledger or legacy review backlog entry"
                )


def _citation_pdf_path(path: Path) -> str:
    match = PAPER_PDF_RE.search(path.read_text(encoding="utf-8", errors="replace"))
    if match is None:
        return ""
    return match.group(1).strip()


def _audit_citation_record_identity(root: Path, errors: list[str]) -> None:
    papers_dir = root / "citations" / "papers"
    if not papers_dir.exists():
        return
    for path in sorted(papers_dir.glob("*.md")):
        if path.name == ".gitkeep":
            continue
        rel_path = _rel(root, path)
        fields = _parse_markdown_fields(path)
        key = str(fields.get("key", "")).strip()
        if key and key != path.stem:
            errors.append(f"{rel_path} key {key} must match filename {path.stem}")


def _parse_markdown_fields(path: Path) -> dict[str, str]:
    fields: dict[str, str] = {}
    for match in CITATION_FIELD_RE.finditer(path.read_text(encoding="utf-8", errors="replace")):
        fields[match.group("name").strip().lower()] = match.group("value").strip()
    return fields


def _is_ignored_pdf_inbox_path(path: str) -> bool:
    return path == "research/papers" or path.startswith("research/papers/")


def _audit_source_ledgers(root: Path, errors: list[str]) -> None:
    sources_dir = root / "research" / "sources"
    if not sources_dir.exists():
        return
    for path in sorted(sources_dir.glob("*.yaml")):
        if path.name.endswith(".acquisition.yaml") or path.name.endswith(".promotion.yaml"):
            continue
        rel_path = _rel(root, path)
        data = _parse_mapping_yaml(path)
        paper_key = str(data.get("paper_key", "")).strip()
        if not paper_key:
            errors.append(f"{rel_path} missing paper_key")
        elif paper_key != path.stem:
            errors.append(f"{rel_path} paper_key {paper_key} must match filename {path.stem}")

        citation_file = _audit_source_file_reference(root, rel_path, "citation_path", data.get("citation_path"), errors)
        if citation_file is not None and citation_file.stem != path.stem:
            errors.append(f"{rel_path} citation_path must point at citations/papers/{path.stem}.md")

        pdf = data.get("pdf")
        if not isinstance(pdf, dict):
            errors.append(f"{rel_path} missing pdf metadata")
            continue
        source_pdf_path = str(pdf.get("path", "")).strip()
        if _is_ignored_pdf_inbox_path(source_pdf_path):
            errors.append(
                f"{rel_path} pdf path {source_pdf_path} points at ignored local inbox; "
                "promote it before writing source ledgers"
            )
            continue
        pdf_file = _audit_source_file_reference(root, rel_path, "pdf path", pdf.get("path"), errors)
        expected_sha = str(pdf.get("sha256", "")).strip()
        if not expected_sha:
            errors.append(f"{rel_path} missing pdf sha256")
        elif pdf_file is not None:
            actual_sha = _sha256_file(pdf_file)
            if expected_sha != actual_sha:
                errors.append(f"{rel_path} pdf sha256 {expected_sha} does not match actual {actual_sha}")


def _audit_openalex_ledgers(root: Path, errors: list[str]) -> None:
    sources_dir = root / "research" / "sources"
    if not sources_dir.exists():
        return
    for path in sorted(sources_dir.glob("*.openalex.json")):
        rel_path = _rel(root, path)
        expected_key = path.name.removesuffix(".openalex.json")
        try:
            data = json.loads(path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError) as exc:
            errors.append(f"{rel_path} invalid JSON: {exc}")
            continue
        if not isinstance(data, dict):
            errors.append(f"{rel_path} must contain a JSON object")
            continue

        openalex_id = str(data.get("id", "")).strip()
        openalex_doi = str(data.get("doi", "")).strip()
        if not openalex_id:
            errors.append(f"{rel_path} missing id")
        elif not openalex_id.startswith("https://openalex.org/"):
            errors.append(f"{rel_path} id must be an OpenAlex URL")
        if not str(data.get("display_name", "")).strip():
            errors.append(f"{rel_path} missing display_name")

        citation_path = root / "citations" / "papers" / f"{expected_key}.md"
        if citation_path.is_file():
            citation_fields = _parse_markdown_fields(citation_path)
            citation_openalex_id = citation_fields.get("openalex", "").strip()
            if citation_openalex_id and openalex_id and openalex_id != citation_openalex_id:
                errors.append(
                    f"{rel_path} id {openalex_id} "
                    f"does not match citation OpenAlex {citation_openalex_id}"
                )
            citation_doi = citation_fields.get("doi", "").strip()
            if citation_doi and openalex_doi and _normalize_doi(openalex_doi) != _normalize_doi(citation_doi):
                errors.append(f"{rel_path} doi {openalex_doi} does not match citation DOI {citation_doi}")

        acquisition_path = sources_dir / f"{expected_key}.acquisition.yaml"
        if acquisition_path.is_file():
            acquisition = _parse_mapping_yaml(acquisition_path)
            acquisition_openalex_id = str(acquisition.get("openalex_id", "")).strip()
            if acquisition_openalex_id and openalex_id and openalex_id != acquisition_openalex_id:
                errors.append(
                    f"{rel_path} id {openalex_id} "
                    f"does not match acquisition openalex_id {acquisition_openalex_id}"
                )
            acquisition_doi = str(acquisition.get("doi", "")).strip()
            if acquisition_doi and openalex_doi and _normalize_doi(openalex_doi) != _normalize_doi(acquisition_doi):
                errors.append(f"{rel_path} doi {openalex_doi} does not match acquisition DOI {acquisition_doi}")


def _audit_acquisition_ledgers(root: Path, errors: list[str]) -> None:
    sources_dir = root / "research" / "sources"
    if not sources_dir.exists():
        return
    for path in sorted(sources_dir.glob("*.acquisition.yaml")):
        rel_path = _rel(root, path)
        data = _parse_mapping_yaml(path)
        expected_key = path.name.removesuffix(".acquisition.yaml")
        paper_key = str(data.get("paper_key", "")).strip()
        if not paper_key:
            errors.append(f"{rel_path} missing paper_key")
        elif paper_key != expected_key:
            errors.append(f"{rel_path} paper_key {paper_key} must match filename {expected_key}")

        citation_path = str(data.get("citation_path", "")).strip()
        if citation_path:
            citation_file = _audit_source_file_reference(root, rel_path, "citation_path", citation_path, errors)
            if citation_file is not None and citation_file.stem != expected_key:
                errors.append(f"{rel_path} citation_path must point at citations/papers/{expected_key}.md")

        status = str(data.get("status", "")).strip()
        pdf_path = str(data.get("pdf_path", "")).strip()
        pdf_file: Path | None = None
        if pdf_path:
            if not _is_repo_relative_path(pdf_path):
                errors.append(f"{rel_path} pdf_path must be repo-relative")
            elif not _is_ignored_pdf_inbox_path(pdf_path):
                errors.append(f"{rel_path} pdf_path {pdf_path} must point at ignored local inbox")
            elif status == "downloaded":
                pdf_file = root / pdf_path
                if not pdf_file.is_file():
                    errors.append(f"{rel_path} downloaded pdf_path {pdf_path} does not exist")
                    pdf_file = None
        elif status == "downloaded":
            errors.append(f"{rel_path} downloaded acquisition missing pdf_path")

        expected_sha = str(data.get("sha256", "")).strip()
        if status == "downloaded":
            if not expected_sha:
                errors.append(f"{rel_path} downloaded acquisition missing sha256")
            elif pdf_file is not None:
                actual_sha = _sha256_file(pdf_file)
                if expected_sha != actual_sha:
                    errors.append(
                        f"{rel_path} acquisition sha256 {expected_sha} does not match actual {actual_sha}"
                    )


def _audit_promotion_ledgers(root: Path, errors: list[str]) -> None:
    sources_dir = root / "research" / "sources"
    if not sources_dir.exists():
        return
    for path in sorted(sources_dir.glob("*.promotion.yaml")):
        rel_path = _rel(root, path)
        data = _parse_mapping_yaml(path)
        paper_key = str(data.get("paper_key", "")).strip()
        expected_key = path.name.removesuffix(".promotion.yaml")
        if not paper_key:
            errors.append(f"{rel_path} missing paper_key")
        elif paper_key != expected_key:
            errors.append(f"{rel_path} paper_key {paper_key} must match filename {expected_key}")

        if str(data.get("status", "")).strip() != "promoted":
            errors.append(f"{rel_path} status must be promoted")

        license_url = str(data.get("license_url", "")).strip()
        for field in ["license", "license_url", "review_note", "source_path"]:
            if not str(data.get(field, "")).strip():
                errors.append(f"{rel_path} missing {field}")
        if license_url and not license_url.startswith(("http://", "https://")):
            errors.append(f"{rel_path} license_url must be an http or https URL")

        citation_file = _audit_source_file_reference(root, rel_path, "citation_path", data.get("citation_path"), errors)
        destination_path = str(data.get("destination_path", "")).strip()
        if _is_ignored_pdf_inbox_path(destination_path):
            errors.append(f"{rel_path} destination_path {destination_path} points at ignored local inbox")
            destination_file = None
        else:
            destination_file = _audit_source_file_reference(
                root,
                rel_path,
                "destination_path",
                data.get("destination_path"),
                errors,
            )

        if citation_file is not None and destination_path:
            citation_pdf = _citation_pdf_path(citation_file)
            if citation_pdf != destination_path:
                errors.append(
                    f"{rel_path} citation PDF path {citation_pdf or 'not stored'} "
                    f"does not match destination_path {destination_path}"
                )

        source_path = str(data.get("source_path", "")).strip()
        if source_path and not _is_repo_relative_path(source_path):
            errors.append(f"{rel_path} source_path must be repo-relative")
        elif source_path and not source_path.lower().endswith(".pdf"):
            errors.append(f"{rel_path} source_path must end with .pdf")

        expected_sha = str(data.get("sha256", "")).strip()
        if not expected_sha:
            errors.append(f"{rel_path} missing sha256")
        elif destination_file is not None:
            actual_sha = _sha256_file(destination_file)
            if expected_sha != actual_sha:
                errors.append(f"{rel_path} promotion sha256 {expected_sha} does not match actual {actual_sha}")


def _load_pdf_review_backlog(root: Path, errors: list[str]) -> dict[str, dict[str, object]]:
    path = root / PDF_REVIEW_BACKLOG_PATH
    if not path.exists():
        return {}
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        errors.append(f"{PDF_REVIEW_BACKLOG_PATH} invalid JSON: {exc}")
        return {}
    if not isinstance(data, dict):
        errors.append(f"{PDF_REVIEW_BACKLOG_PATH} must contain a JSON object")
        return {}
    if data.get("schema") != PDF_REVIEW_BACKLOG_SCHEMA:
        errors.append(f"{PDF_REVIEW_BACKLOG_PATH} schema must be {PDF_REVIEW_BACKLOG_SCHEMA}")
    entries = data.get("entries", [])
    if not isinstance(entries, list):
        errors.append(f"{PDF_REVIEW_BACKLOG_PATH} entries must be a list")
        return {}

    backlog: dict[str, dict[str, object]] = {}
    for index, entry in enumerate(entries):
        if not isinstance(entry, dict):
            errors.append(f"{PDF_REVIEW_BACKLOG_PATH} entry {index} must be a JSON object")
            continue
        paper_key = str(entry.get("paper_key", "")).strip()
        if not paper_key:
            errors.append(f"{PDF_REVIEW_BACKLOG_PATH} entry {index} missing paper_key")
            continue
        if paper_key in backlog:
            errors.append(f"{PDF_REVIEW_BACKLOG_PATH} duplicate entry {paper_key}")
            continue
        backlog[paper_key] = entry
    return backlog


def _audit_pdf_review_backlog(
    root: Path,
    pdf_review_backlog: dict[str, dict[str, object]],
    errors: list[str],
) -> None:
    for paper_key, entry in sorted(pdf_review_backlog.items()):
        owner = f"{PDF_REVIEW_BACKLOG_PATH} entry {paper_key}"
        if str(entry.get("status", "")).strip() != PDF_REVIEW_BACKLOG_STATUS:
            errors.append(f"{owner} status must be {PDF_REVIEW_BACKLOG_STATUS}")
        for field in ["citation_path", "pdf_path", "sha256", "note"]:
            if not str(entry.get(field, "")).strip():
                errors.append(f"{owner} missing {field}")

        citation_path = str(entry.get("citation_path", "")).strip()
        citation_file = _audit_source_file_reference(root, owner, "citation_path", citation_path, errors)
        if citation_file is not None and citation_file.stem != paper_key:
            errors.append(f"{owner} citation_path must point at citations/papers/{paper_key}.md")

        pdf_path = str(entry.get("pdf_path", "")).strip()
        if _is_ignored_pdf_inbox_path(pdf_path):
            errors.append(f"{owner} pdf_path {pdf_path} points at ignored local inbox")
            pdf_file = None
        else:
            pdf_file = _audit_source_file_reference(root, owner, "pdf_path", pdf_path, errors)

        if citation_file is not None and pdf_path:
            citation_pdf = _citation_pdf_path(citation_file)
            if citation_pdf != pdf_path:
                errors.append(
                    f"{owner} citation PDF path {citation_pdf or 'not stored'} "
                    f"does not match pdf_path {pdf_path}"
                )

        expected_sha = str(entry.get("sha256", "")).strip()
        if expected_sha and pdf_file is not None:
            actual_sha = _sha256_file(pdf_file)
            if expected_sha != actual_sha:
                errors.append(f"{owner} sha256 {expected_sha} does not match actual {actual_sha}")


def _has_promotion_ledger(root: Path, paper_key: str) -> bool:
    return (root / "research" / "sources" / f"{paper_key}.promotion.yaml").is_file()


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


def _normalize_doi(value: str) -> str:
    doi = value.strip().lower()
    if doi.startswith("https://doi.org/"):
        doi = doi.removeprefix("https://doi.org/")
    elif doi.startswith("http://doi.org/"):
        doi = doi.removeprefix("http://doi.org/")
    elif doi.startswith("doi:"):
        doi = doi.removeprefix("doi:")
    return doi


def _sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def _write_report(root: Path, report: ClaimAuditReport) -> None:
    write_content_addressed_report(
        root,
        "research/reports/claim-audit-latest.json",
        "research/reports/claim-audits",
        asdict(report),
    )


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
