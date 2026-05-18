import hashlib
import json
from pathlib import Path
import re


HEADING_RE = re.compile(r"^(#{1,6})\s+(?P<title>.+)$", re.MULTILINE)


def _sha(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


def _sections(markdown: str) -> list[tuple[str, str]]:
    sections: list[tuple[str, str]] = []
    matches = list(HEADING_RE.finditer(markdown))
    if not matches:
        body = markdown.strip()
        return [("Body", body)] if body else []

    prefix = markdown[: matches[0].start()].strip()
    if prefix:
        sections.append(("Body", prefix))

    for index, match in enumerate(matches):
        start = match.end()
        end = matches[index + 1].start() if index + 1 < len(matches) else len(markdown)
        title = match.group("title").strip()
        body = markdown[start:end].strip()
        if body:
            sections.append((title, body))
    return sections


def chunk_markdown(paper_key: str, markdown: str, max_chars: int = 1800) -> list[dict[str, object]]:
    chunks: list[dict[str, object]] = []
    chunk_number = 1
    for section_number, (section, body) in enumerate(_sections(markdown), start=1):
        for contents in _chunk_body(body, max_chars):
            chunks.append(_record(paper_key, section, section_number, chunk_number, contents))
            chunk_number += 1
    return chunks


def _chunk_body(body: str, max_chars: int) -> list[str]:
    limit = max(1, max_chars)
    chunks: list[str] = []
    current = ""
    for paragraph in _paragraphs(body):
        if len(paragraph) > limit:
            if current:
                chunks.append(current)
                current = ""
            chunks.extend(_split_long_paragraph(paragraph, limit))
            continue

        candidate = paragraph if current == "" else current + "\n\n" + paragraph
        if len(candidate) <= limit:
            current = candidate
        else:
            if current:
                chunks.append(current)
            current = paragraph

    if current:
        chunks.append(current)
    return chunks


def _paragraphs(body: str) -> list[str]:
    return [part.strip() for part in re.split(r"\n\s*\n", body) if part.strip()]


def _split_long_paragraph(paragraph: str, max_chars: int) -> list[str]:
    pieces: list[str] = []
    current = ""
    for word in paragraph.split():
        candidate = word if current == "" else current + " " + word
        if len(candidate) <= max_chars:
            current = candidate
        else:
            if current:
                pieces.append(current)
            current = word
    if current:
        pieces.append(current)
    return pieces


def _record(
    paper_key: str,
    section: str,
    section_number: int,
    chunk_number: int,
    contents: str,
) -> dict[str, object]:
    return {
        "id": f"{paper_key}::sec-{section_number:02d}::chunk-{chunk_number:03d}",
        "paper_key": paper_key,
        "contents": contents,
        "section": section,
        "section_number": section_number,
        "chunk_number": chunk_number,
        "source_parser": "marker",
        "source_path": f"research/parsed/{paper_key}/marker.md",
        "page_start": None,
        "page_end": None,
        "char_start": None,
        "char_end": None,
        "sha256": _sha(contents),
    }


def write_chunks_jsonl(path: Path, chunks: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        for chunk in chunks:
            f.write(json.dumps(chunk, ensure_ascii=False, sort_keys=True) + "\n")
