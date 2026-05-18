import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.citations import load_citation_records
from fecim_research.discovery import discover_pdfs, match_pdf_to_record
from fecim_research.registration import run_register_pdfs


class RegistrationTest(unittest.TestCase):
    def test_register_pdfs_reports_unmatched_without_writing_stubs_by_default(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_pdf(root, "research/papers/park2015_advmat_hzo.pdf", b"%PDF matched")
            self._write_pdf(root, "research/papers/new_hzo_device.pdf", b"%PDF new")
            self._write_pdf(root, "research/papers/new_hzo_device_copy.pdf", b"%PDF new")

            code = run_register_pdfs(root=root, extra_paths=[], write_stubs=False)

            self.assertEqual(code, 0)
            self.assertFalse((root / "citations" / "papers" / "new_hzo_device.md").exists())
            report = json.loads((root / "research" / "reports" / "pdf-registration-latest.json").read_text())
            self.assertEqual(report["discovered"], 3)
            self.assertEqual(report["matched"], 1)
            self.assertEqual(report["unmatched"], 1)
            self.assertEqual(report["duplicates"], 1)
            self.assertEqual(report["stubs_planned"], 1)
            self.assertEqual(report["stubs_written"], 0)
            statuses = {item["path"]: item["status"] for item in report["pdfs"]}
            self.assertEqual(statuses["research/papers/park2015_advmat_hzo.pdf"], "matched")
            self.assertEqual(statuses["research/papers/new_hzo_device.pdf"], "unmatched")
            self.assertEqual(statuses["research/papers/new_hzo_device_copy.pdf"], "duplicate")

    def test_register_pdfs_writes_review_stubs_for_unmatched_canonical_pdfs(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_pdf(root, "research/papers/new_hzo_device.pdf", b"%PDF new")

            code = run_register_pdfs(root=root, extra_paths=[], write_stubs=True)

            self.assertEqual(code, 0)
            stub = root / "citations" / "papers" / "new_hzo_device.md"
            self.assertTrue(stub.exists())
            text = stub.read_text(encoding="utf-8")
            self.assertIn("**Key:** `new_hzo_device`", text)
            self.assertIn("**Status:** `needs-review`", text)
            self.assertIn("**PDF:** `not stored`", text)
            self.assertIn("**Local PDF:** `research/papers/new_hzo_device.pdf`", text)
            self.assertIn("**SHA256:**", text)
            self.assertIn("Confirm bibliographic metadata", text)

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.status, "matched")
            self.assertEqual(match.paper_key, "new_hzo_device")

            report = json.loads((root / "research" / "reports" / "pdf-registration-latest.json").read_text())
            self.assertEqual(report["stubs_written"], 1)
            self.assertEqual(report["pdfs"][0]["stub_path"], "citations/papers/new_hzo_device.md")
            missing_report = json.loads((root / "research" / "reports" / "missing-papers-latest.json").read_text())
            self.assertEqual(missing_report["total_records"], 1)
            self.assertEqual(missing_report["stored"], 1)
            self.assertEqual(missing_report["missing"], 0)

    def test_register_pdfs_uses_hash_suffix_when_stub_key_would_collide(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_pdf(root, "incoming/a/colliding_name.pdf", b"%PDF first")
            self._write_pdf(root, "incoming/b/colliding_name.pdf", b"%PDF second")

            code = run_register_pdfs(root=root, extra_paths=[Path("incoming")], write_stubs=True)

            self.assertEqual(code, 0)
            stubs = sorted(path.name for path in (root / "citations" / "papers").glob("colliding_name*.md"))
            self.assertEqual(len(stubs), 2)
            self.assertEqual(stubs[0], "colliding_name.md")
            self.assertRegex(stubs[1], r"^colliding_name_[a-f0-9]{8}\.md$")

    def _write_paper(self, root: Path, key: str):
        path = root / "citations" / "papers" / f"{key}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(f"**Key:** `{key}`\n", encoding="utf-8")

    def _write_pdf(self, root: Path, rel_path: str, content: bytes):
        path = root / rel_path
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content)


if __name__ == "__main__":
    unittest.main()
