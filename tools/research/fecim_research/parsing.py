from dataclasses import asdict, dataclass
import json
import os
from pathlib import Path
import shutil
import shlex
import subprocess


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
