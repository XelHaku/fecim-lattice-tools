from dataclasses import asdict, dataclass
import json
import os
from pathlib import Path
import shutil
import shlex
import subprocess
import urllib.request


@dataclass(frozen=True)
class ParseResult:
    paper_key: str
    parser: str
    status: str
    output_path: str
    message: str


def write_parse_manifest(path: Path, results: list[ParseResult]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    data = {"results": [asdict(result) for result in results]}
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def run_marker_if_configured(pdf: Path, out_md: Path) -> ParseResult:
    paper_key = pdf.stem
    marker_cmd = os.environ.get("FECIM_MARKER_CMD", "").strip()
    if marker_cmd == "":
        return ParseResult(
            paper_key=paper_key,
            parser="marker",
            status="skipped",
            output_path=str(out_md),
            message="FECIM_MARKER_CMD is not set",
        )

    out_md.parent.mkdir(parents=True, exist_ok=True)
    command = shlex.split(marker_cmd) + [str(pdf), str(out_md)]
    completed = subprocess.run(command, capture_output=True, text=True, check=False)
    status = "ok" if completed.returncode == 0 else "failed"
    return ParseResult(
        paper_key=paper_key,
        parser="marker",
        status=status,
        output_path=str(out_md),
        message=completed.stderr.strip(),
    )


def copy_sidecar_markdown_if_present(pdf: Path, out_md: Path) -> ParseResult:
    paper_key = pdf.stem
    sidecar = pdf.with_suffix(".md")
    if not sidecar.exists():
        return ParseResult(
            paper_key=paper_key,
            parser="sidecar_markdown",
            status="skipped",
            output_path=str(out_md),
            message=f"{sidecar} does not exist",
        )

    out_md.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(sidecar, out_md)
    return ParseResult(
        paper_key=paper_key,
        parser="sidecar_markdown",
        status="ok",
        output_path=str(out_md),
        message=f"copied {sidecar}",
    )


def run_grobid_if_available(pdf: Path, out_tei: Path) -> ParseResult:
    raw_url = os.environ.get("FECIM_GROBID_URL", "").strip()
    if not raw_url:
        return ParseResult("", "grobid", "skipped", str(out_tei), "FECIM_GROBID_URL is not set")
    url = raw_url.rstrip("/")
    endpoint = f"{url}/api/processFulltextDocument"
    try:
        boundary = "----fecimresearchboundary"
        data = pdf.read_bytes()
        body = (
            f"--{boundary}\r\n"
            f'Content-Disposition: form-data; name="input"; filename="{pdf.name}"\r\n'
            "Content-Type: application/pdf\r\n\r\n"
        ).encode("utf-8") + data + f"\r\n--{boundary}--\r\n".encode("utf-8")
        request = urllib.request.Request(
            endpoint,
            data=body,
            headers={"Content-Type": f"multipart/form-data; boundary={boundary}"},
            method="POST",
        )
        with urllib.request.urlopen(request, timeout=20) as response:
            text = response.read().decode("utf-8", errors="replace")
    except Exception as exc:
        return ParseResult("", "grobid", "failed", str(out_tei), str(exc))
    out_tei.parent.mkdir(parents=True, exist_ok=True)
    out_tei.write_text(text, encoding="utf-8")
    return ParseResult("", "grobid", "ok", str(out_tei), "grobid completed")
