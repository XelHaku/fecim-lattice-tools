from __future__ import annotations

from collections.abc import Callable
from dataclasses import asdict, dataclass, replace
from pathlib import Path
from urllib.parse import quote, urlencode
import hashlib
import json
import os
import re
import urllib.request

from .citations import CitationRecord, load_citation_records
from .discovery import discover_pdfs, match_pdf_to_record
from .yamlio import dumps_yaml


OPENALEX_WORKS_URL = "https://api.openalex.org/works/"
OPENALEX_SELECT = ",".join(
    [
        "id",
        "doi",
        "display_name",
        "open_access",
        "best_oa_location",
        "locations",
        "publication_year",
        "publication_date",
        "type",
        "is_retracted",
    ]
)


@dataclass(frozen=True)
class AcquisitionResult:
    paper_key: str
    status: str
    doi: str
    citation_path: str = ""
    openalex_id: str = ""
    pdf_url: str = ""
    landing_page_url: str = ""
    license: str = ""
    version: str = ""
    pdf_path: str = ""
    sha256: str = ""
    message: str = ""


def run_acquire(
    root: Path,
    keys: list[str],
    download: bool,
    opener: Callable[..., object] = urllib.request.urlopen,
    dois: list[str] | None = None,
) -> int:
    records = load_citation_records(root)
    selected = _select_records(root, records, keys)
    results: list[AcquisitionResult] = []

    for record in selected:
        result, work = _acquire_record(root, record, download, opener)
        results.append(result)
        if work:
            _write_openalex_record(root, record.key, work)
        _write_acquisition_record(root, result)
    for doi in sorted(dois or []):
        paper_key = provisional_key_for_doi(doi)
        result, work = _acquire_doi(root, paper_key, doi, download, opener)
        if work:
            _write_openalex_record(root, paper_key, work)
            citation_path = _write_provisional_citation_stub(root, paper_key, doi, work, result)
            result = replace(result, citation_path=citation_path)
        results.append(result)
        _write_acquisition_record(root, result)

    _write_report(root, results)
    _refresh_missing_report(root)
    planned = sum(1 for result in results if result.status in {"planned", "downloaded"})
    downloaded = sum(1 for result in results if result.status == "downloaded")
    print(f"acquire complete: planned={planned} downloaded={downloaded} checked={len(results)}")
    return 0


def _select_records(root: Path, records: dict[str, CitationRecord], keys: list[str]) -> list[CitationRecord]:
    existing = _existing_pdf_keys(root, records)
    if keys:
        selected = [records[key] for key in keys if key in records]
    else:
        selected = [record for key, record in sorted(records.items()) if key not in existing]
    return sorted(selected, key=lambda record: record.key)


def _existing_pdf_keys(root: Path, records: dict[str, CitationRecord]) -> set[str]:
    keys: set[str] = set()
    for pdf in discover_pdfs(root, extra_paths=[]):
        match = match_pdf_to_record(pdf, records)
        if match.paper_key is not None:
            keys.add(match.paper_key)
    return keys


def _acquire_record(
    root: Path,
    record: CitationRecord,
    download: bool,
    opener: Callable[..., object],
) -> tuple[AcquisitionResult, dict[str, object] | None]:
    if not record.doi:
        return (
            AcquisitionResult(
                record.key,
                "missing_doi",
                "",
                citation_path=_rel(root, record.path),
                message="citation record has no DOI",
            ),
            None,
        )

    result, work = _acquire_doi(root, record.key, record.doi, download, opener)
    return replace(result, citation_path=_rel(root, record.path)), work


def _acquire_doi(
    root: Path,
    paper_key: str,
    doi: str,
    download: bool,
    opener: Callable[..., object],
) -> tuple[AcquisitionResult, dict[str, object] | None]:
    try:
        work = fetch_openalex_work(doi, opener=opener)
    except Exception as exc:
        return AcquisitionResult(paper_key, "metadata_failed", doi, message=str(exc)), None

    candidate = best_oa_pdf_candidate(work)
    if candidate is None:
        return (
            AcquisitionResult(
                paper_key=paper_key,
                status="no_oa_pdf",
                doi=doi,
                openalex_id=str(work.get("id", "")),
                message="OpenAlex did not report an open-access PDF URL",
            ),
            work,
        )

    result = AcquisitionResult(
        paper_key=paper_key,
        status="planned",
        doi=doi,
        openalex_id=str(work.get("id", "")),
        pdf_url=str(candidate.get("pdf_url", "")),
        landing_page_url=str(candidate.get("landing_page_url", "")),
        license=str(candidate.get("license") or ""),
        version=str(candidate.get("version") or ""),
        pdf_path=f"research/papers/{paper_key}.pdf",
        message="open-access PDF located via OpenAlex",
    )
    if not download:
        return result, work

    pdf_path = root / "research" / "papers" / f"{paper_key}.pdf"
    try:
        digest = download_pdf(result.pdf_url, pdf_path, opener=opener)
    except Exception as exc:
        return (
            AcquisitionResult(
                **{
                    **asdict(result),
                    "status": "download_failed",
                    "message": str(exc),
                }
            ),
            work,
        )
    return (
        AcquisitionResult(
            **{
                **asdict(result),
                "status": "downloaded",
                "sha256": digest,
                "message": "downloaded open-access PDF",
            }
        ),
        work,
    )


def fetch_openalex_work(doi: str, opener: Callable[..., object] = urllib.request.urlopen) -> dict[str, object]:
    request = urllib.request.Request(openalex_work_url(doi), headers={"Accept": "application/json"})
    with opener(request, timeout=30) as response:
        return json.loads(response.read().decode("utf-8", errors="replace"))


def openalex_work_url(doi: str) -> str:
    identifier = doi.strip()
    if identifier.startswith("https://doi.org/"):
        external_id = identifier
    elif identifier.startswith("doi:"):
        external_id = identifier
    else:
        external_id = "https://doi.org/" + identifier
    query: dict[str, str] = {"select": OPENALEX_SELECT}
    api_key = os.environ.get("FECIM_OPENALEX_API_KEY", "").strip()
    if api_key:
        query["api_key"] = api_key
    mailto = os.environ.get("FECIM_OPENALEX_MAILTO", "").strip()
    if mailto:
        query["mailto"] = mailto
    return OPENALEX_WORKS_URL + quote(external_id, safe=":/") + "?" + urlencode(query)


def provisional_key_for_doi(doi: str) -> str:
    identifier = doi.strip()
    if identifier.startswith("https://doi.org/"):
        identifier = identifier.removeprefix("https://doi.org/")
    elif identifier.startswith("doi:"):
        identifier = identifier.removeprefix("doi:")
    slug = re.sub(r"[^a-z0-9]+", "_", identifier.lower()).strip("_")
    if not slug:
        slug = "unknown"
    return "doi_" + slug[:96]


def best_oa_pdf_candidate(work: dict[str, object]) -> dict[str, object] | None:
    candidates: list[dict[str, object]] = []
    best = work.get("best_oa_location")
    if isinstance(best, dict):
        candidates.append(best)
    locations = work.get("locations")
    if isinstance(locations, list):
        candidates.extend(location for location in locations if isinstance(location, dict))
    for location in candidates:
        pdf_url = location.get("pdf_url")
        if location.get("is_oa") is True and isinstance(pdf_url, str) and _is_http_url(pdf_url):
            return location
    return None


def download_pdf(url: str, path: Path, opener: Callable[..., object] = urllib.request.urlopen) -> str:
    if not _is_http_url(url):
        raise ValueError(f"refusing non-http PDF URL: {url}")
    request = urllib.request.Request(url, headers={"Accept": "application/pdf"})
    with opener(request, timeout=60) as response:
        data = response.read()
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)
    return hashlib.sha256(data).hexdigest()


def _write_openalex_record(root: Path, paper_key: str, work: dict[str, object]) -> None:
    path = root / "research" / "sources" / f"{paper_key}.openalex.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(work, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _write_acquisition_record(root: Path, result: AcquisitionResult) -> None:
    path = root / "research" / "sources" / f"{result.paper_key}.acquisition.yaml"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(dumps_yaml(asdict(result)), encoding="utf-8")


def _write_provisional_citation_stub(
    root: Path,
    paper_key: str,
    doi: str,
    work: dict[str, object],
    result: AcquisitionResult,
) -> str:
    path = root / "citations" / "papers" / f"{paper_key}.md"
    rel_path = _rel(root, path)
    if path.exists():
        return rel_path

    path.parent.mkdir(parents=True, exist_ok=True)
    title = _work_title(work, paper_key)
    year = _code_value(work.get("publication_year"), "needs-review")
    venue = _code_value(_venue_from_work(work), "needs-review")
    openalex_id = _code_value(work.get("id"), "needs-review")
    local_pdf = result.pdf_path or "not downloaded"
    sha256 = result.sha256 or "not downloaded"
    oa_pdf_url = result.pdf_url or "not found"
    landing_page_url = result.landing_page_url or "not found"
    license_value = result.license or "needs-review"
    version = result.version or "needs-review"
    text = (
        f"# {title}\n\n"
        f"**Key:** `{paper_key}`\n"
        f"**Title:** `{_code_value(title, 'needs-review')}`\n"
        f"**DOI:** `{_code_value(doi, 'needs-review')}`\n"
        f"**Year:** `{year}`\n"
        f"**Venue:** `{venue}`\n"
        "**Authors:** `needs-review`\n"
        "**Tags:** `#needs-review`\n"
        "**Status:** `needs-review`\n"
        "**PDF:** `not stored`\n"
        f"**OpenAlex:** `{openalex_id}`\n"
        "\n---\n\n"
        "## Acquisition\n\n"
        f"- **Status:** `{result.status}`\n"
        f"- **Local PDF:** `{_code_value(local_pdf, 'not downloaded')}`\n"
        f"- **SHA256:** `{_code_value(sha256, 'not downloaded')}`\n"
        f"- **OA PDF URL:** `{_code_value(oa_pdf_url, 'not found')}`\n"
        f"- **Landing page:** `{_code_value(landing_page_url, 'not found')}`\n"
        f"- **License:** `{_code_value(license_value, 'needs-review')}`\n"
        f"- **Version:** `{_code_value(version, 'needs-review')}`\n"
        "\n## Review TODO\n\n"
        "- [ ] Confirm bibliographic metadata.\n"
        "- [ ] Add authors and venue details from the paper.\n"
        "- [ ] Move the PDF into a tracked paper collection only if licensing permits.\n"
        "- [ ] Add source notes only after reading the paper.\n"
        "- [ ] Move evidence-bearing claims into `citations/claims/*.yaml`.\n"
    )
    path.write_text(text, encoding="utf-8")
    return rel_path


def _write_report(root: Path, results: list[AcquisitionResult]) -> None:
    path = root / "research" / "reports" / "acquisition-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "checked": len(results),
        "planned": sum(1 for result in results if result.status in {"planned", "downloaded"}),
        "downloaded": sum(1 for result in results if result.status == "downloaded"),
        "results": [asdict(result) for result in sorted(results, key=lambda item: item.paper_key)],
    }
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def _refresh_missing_report(root: Path) -> None:
    from .missing import run_missing

    run_missing(root=root)


def _is_http_url(url: str) -> bool:
    return url.startswith("https://") or url.startswith("http://")


def _work_title(work: dict[str, object], fallback_key: str) -> str:
    title = work.get("display_name")
    if isinstance(title, str) and title.strip():
        return title.strip()
    return " ".join(
        part.upper() if part in {"ai", "hzo", "hfo2", "cim"} else part.capitalize()
        for part in fallback_key.split("_")
    )


def _venue_from_work(work: dict[str, object]) -> str:
    locations: list[object] = []
    best = work.get("best_oa_location")
    if best is not None:
        locations.append(best)
    raw_locations = work.get("locations")
    if isinstance(raw_locations, list):
        locations.extend(raw_locations)
    for location in locations:
        if not isinstance(location, dict):
            continue
        source = location.get("source")
        if not isinstance(source, dict):
            continue
        display_name = source.get("display_name")
        if isinstance(display_name, str) and display_name.strip():
            return display_name.strip()
    return ""


def _code_value(value: object, fallback: str) -> str:
    text = str(value or "").strip()
    if not text:
        text = fallback
    return text.replace("`", "'")


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
