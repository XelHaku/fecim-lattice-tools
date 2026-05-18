from dataclasses import dataclass
from pathlib import Path
import re


FIELD_RE = re.compile(r"^\*\*(?P<name>[^*]+):\*\*\s*`?(?P<value>[^`\n]+)`?", re.MULTILINE)
H1_RE = re.compile(r"^#\s+(?P<title>.+)$", re.MULTILINE)


@dataclass(frozen=True)
class CitationRecord:
    key: str
    path: Path
    title: str = ""
    doi: str = ""
    arxiv_id: str = ""


def _fields(text: str) -> dict[str, str]:
    out: dict[str, str] = {}
    for match in FIELD_RE.finditer(text):
        out[match.group("name").strip().lower()] = match.group("value").strip()
    return out


def _h1_title(text: str) -> str:
    match = H1_RE.search(text)
    if match is None:
        return ""
    return match.group("title").strip()


def load_citation_records(root: Path) -> dict[str, CitationRecord]:
    records: dict[str, CitationRecord] = {}
    papers_dir = root / "citations" / "papers"
    if not papers_dir.exists():
        return records
    for path in sorted(papers_dir.glob("*.md")):
        text = path.read_text(encoding="utf-8", errors="replace")
        fields = _fields(text)
        key = fields.get("key") or path.stem
        records[key] = CitationRecord(
            key=key,
            path=path,
            title=fields.get("title", "") or _h1_title(text),
            doi=fields.get("doi", ""),
            arxiv_id=fields.get("arxiv", ""),
        )
    return records
