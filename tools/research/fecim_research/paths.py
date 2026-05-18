from pathlib import Path


def repo_root(start: Path | None = None) -> Path:
    current = (start or Path.cwd()).resolve()
    for candidate in [current, *current.parents]:
        if (candidate / "go.mod").exists() and (candidate / "citations").is_dir():
            return candidate
    raise RuntimeError("could not locate repository root")


def research_root(root: Path) -> Path:
    return root / "research"
