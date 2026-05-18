from dataclasses import dataclass
from pathlib import Path
import hashlib

from .citations import CitationRecord


DEFAULT_PDF_GLOBS = (
    "docs/4-research/papers/**/*.pdf",
    "research/papers/**/*.pdf",
    "citations/pdfs/**/*.pdf",
)


@dataclass(frozen=True)
class DiscoveredPDF:
    path: Path
    sha256: str
    size: int


@dataclass(frozen=True)
class PDFMatch:
    status: str
    paper_key: str | None
    method: str
    confidence: float


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def discover_pdfs(root: Path, extra_paths: list[Path]) -> list[DiscoveredPDF]:
    paths: set[Path] = set()
    for pattern in DEFAULT_PDF_GLOBS:
        paths.update(root.glob(pattern))
        if pattern.endswith(".pdf"):
            paths.update(root.glob(pattern[:-4] + ".PDF"))
    for extra in extra_paths:
        base = extra if extra.is_absolute() else root / extra
        if base.is_file() and base.suffix.lower() == ".pdf":
            paths.add(base)
        elif base.is_dir():
            paths.update(path for path in base.rglob("*") if path.suffix.lower() == ".pdf")
    out: list[DiscoveredPDF] = []
    for path in sorted(paths):
        if not path.is_file():
            continue
        out.append(DiscoveredPDF(path=path, sha256=sha256_file(path), size=path.stat().st_size))
    return out


def match_pdf_to_record(pdf: DiscoveredPDF, records: dict[str, CitationRecord]) -> PDFMatch:
    stem = pdf.path.stem.lower()
    for key in sorted(records):
        if stem == key.lower():
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
    for key in sorted(records, key=lambda item: (-len(item), item)):
        if key.lower() in stem:
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
    return PDFMatch(status="unmatched", paper_key=None, method="none", confidence=0.0)
