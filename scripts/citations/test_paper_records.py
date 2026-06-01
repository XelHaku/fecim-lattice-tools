import os
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]


class CitationPaperRecordScriptsTest(unittest.TestCase):
    def run_script(self, script: str, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["bash", f"scripts/citations/{script}", *args],
            cwd=REPO_ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def test_compile_bib_discovers_records_in_subdirectories(self):
        nested_dir = REPO_ROOT / "citations" / "papers" / "_tmp_test_records"
        nested_file = nested_dir / "nested_compile_probe.md"
        refs = REPO_ROOT / "citations" / "refs.bib"
        original_refs = refs.read_text(encoding="utf-8") if refs.exists() else None
        nested_dir.mkdir(parents=True, exist_ok=True)
        nested_file.write_text(
            "# Nested Compile Probe\n\n"
            "```bibtex\n"
            "@article{nested_compile_probe, title={Nested Compile Probe}}\n"
            "```\n",
            encoding="utf-8",
        )
        try:
            result = self.run_script("compile_bib.sh")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("nested_compile_probe", refs.read_text(encoding="utf-8"))
        finally:
            nested_file.unlink(missing_ok=True)
            try:
                nested_dir.rmdir()
            except OSError:
                pass
            if original_refs is not None:
                refs.write_text(original_refs, encoding="utf-8")
            else:
                refs.unlink(missing_ok=True)

    def test_search_discovers_records_in_subdirectories(self):
        nested_dir = REPO_ROOT / "citations" / "papers" / "_tmp_test_records"
        nested_file = nested_dir / "nested_search_probe.md"
        nested_dir.mkdir(parents=True, exist_ok=True)
        nested_file.write_text("# Nested Search Probe\n\nunique-nested-search-token\n", encoding="utf-8")
        try:
            result = self.run_script("search.sh", "unique-nested-search-token")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn(str(nested_file.relative_to(REPO_ROOT)), result.stdout)
        finally:
            nested_file.unlink(missing_ok=True)
            try:
                nested_dir.rmdir()
            except OSError:
                pass


if __name__ == "__main__":
    unittest.main()
