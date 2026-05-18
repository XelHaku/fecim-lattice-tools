from collections.abc import Mapping, Sequence


def dumps_yaml(value: object, indent: int = 0) -> str:
    lines: list[str] = []
    _emit(value, lines, indent)
    return "\n".join(lines) + "\n"


def _scalar(value: object) -> str:
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "true" if value else "false"
    text = str(value)
    if text == "" or text.lower() in {"null", "true", "false"} or any(c in text for c in ":#[]{}"):
        return '"' + text.replace('"', '\\"') + '"'
    return text


def _emit(value: object, lines: list[str], indent: int) -> None:
    pad = " " * indent
    if isinstance(value, Mapping):
        for key in sorted(value.keys()):
            item = value[key]
            if isinstance(item, (Mapping, list, tuple)):
                lines.append(f"{pad}{key}:")
                _emit(item, lines, indent + 2)
            else:
                lines.append(f"{pad}{key}: {_scalar(item)}")
        return
    if isinstance(value, Sequence) and not isinstance(value, (str, bytes, bytearray)):
        for item in value:
            if isinstance(item, Mapping):
                lines.append(f"{pad}-")
                _emit(item, lines, indent + 2)
            else:
                lines.append(f"{pad}- {_scalar(item)}")
        return
    lines.append(f"{pad}{_scalar(value)}")
