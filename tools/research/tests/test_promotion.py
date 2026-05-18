import hashlib
import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.promotion import run_promote_pdf


class PromotionTest(unittest.TestCase):
    def test_promote_pdf_copies_reviewed_inbox_pdf_to_tracked_path_and_updates_citation(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            source.parent.mkdir(parents=True)
            pdf_bytes = b"%PDF-1.7\nreviewed fixture\n"
            source.write_bytes(pdf_bytes)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text(
                "# Park 2015\n\n"
                "**Key:** `park2015_advmat_hzo`\n"
                "**DOI:** `10.1002/adma.201404531`\n"
                "**Status:** `needs-review`\n"
                "**PDF:** `not stored`\n"
                "**Local PDF:** `research/papers/park2015_advmat_hzo.pdf`\n",
                encoding="utf-8",
            )
            destination = "docs/4-research/papers/by-topic/01-ferroelectric-materials/park2015_advmat_hzo.pdf"

            code = run_promote_pdf(
                root=root,
                key="park2015_advmat_hzo",
                destination=destination,
                source="",
                license_name="CC BY 4.0",
                license_url="https://example.org/license",
                review_note="Publisher page shows redistribution-compatible license.",
            )

            digest = hashlib.sha256(pdf_bytes).hexdigest()
            self.assertEqual(code, 0)
            self.assertTrue(source.exists())
            self.assertEqual((root / destination).read_bytes(), pdf_bytes)
            citation_text = citation.read_text(encoding="utf-8")
            self.assertIn(f"**PDF:** `{destination}`", citation_text)
            self.assertIn(f"**SHA256:** `{digest}`", citation_text)
            self.assertIn(f"**Size:** `{len(pdf_bytes)}`", citation_text)
            report = json.loads((root / "research" / "reports" / "pdf-promotion-latest.json").read_text())
            self.assertEqual(report["status"], "promoted")
            self.assertEqual(report["paper_key"], "park2015_advmat_hzo")
            self.assertEqual(report["source_path"], "research/papers/park2015_advmat_hzo.pdf")
            self.assertEqual(report["destination_path"], destination)
            self.assertEqual(report["license"], "CC BY 4.0")
            self.assertEqual(report["license_url"], "https://example.org/license")
            self.assertEqual(report["review_note"], "Publisher page shows redistribution-compatible license.")
            self.assertEqual(report["promotion_ledger_path"], "research/sources/park2015_advmat_hzo.promotion.yaml")
            ledger = root / "research" / "sources" / "park2015_advmat_hzo.promotion.yaml"
            self.assertTrue(ledger.exists())
            ledger_text = ledger.read_text(encoding="utf-8")
            self.assertIn("paper_key: park2015_advmat_hzo", ledger_text)
            self.assertIn("license: CC BY 4.0", ledger_text)
            self.assertIn('license_url: "https://example.org/license"', ledger_text)
            self.assertIn("review_note: Publisher page shows redistribution-compatible license.", ledger_text)
            self.assertIn(f"sha256: {digest}", ledger_text)
            self.assertIn(f"destination_path: {destination}", ledger_text)
            missing_report = json.loads((root / "research" / "reports" / "missing-papers-latest.json").read_text())
            self.assertEqual(missing_report["stored"], 1)
            self.assertEqual(missing_report["missing"], 0)

    def test_promote_pdf_requires_review_metadata_before_copying(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            source.parent.mkdir(parents=True)
            source.write_bytes(b"%PDF-1.7\nreviewed fixture\n")
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text(
                "**Key:** `park2015_advmat_hzo`\n"
                "**PDF:** `not stored`\n"
                "**Local PDF:** `research/papers/park2015_advmat_hzo.pdf`\n",
                encoding="utf-8",
            )
            destination = "docs/4-research/papers/by-topic/01-ferroelectric-materials/park2015_advmat_hzo.pdf"

            code = run_promote_pdf(
                root=root,
                key="park2015_advmat_hzo",
                destination=destination,
                source="",
                license_name="",
                license_url="https://example.org/license",
                review_note="Publisher page shows redistribution-compatible license.",
            )

            self.assertEqual(code, 1)
            self.assertFalse((root / destination).exists())
            self.assertFalse((root / "research" / "sources" / "park2015_advmat_hzo.promotion.yaml").exists())
            self.assertIn("**PDF:** `not stored`", citation.read_text(encoding="utf-8"))
            report = json.loads((root / "research" / "reports" / "pdf-promotion-latest.json").read_text())
            self.assertEqual(report["status"], "failed")
            self.assertIn("license, license_url, and review_note are required", report["message"])

    def test_promote_pdf_refuses_ignored_research_papers_destination(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            source.parent.mkdir(parents=True)
            source.write_bytes(b"%PDF-1.7\nreviewed fixture\n")
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text(
                "**Key:** `park2015_advmat_hzo`\n"
                "**PDF:** `not stored`\n"
                "**Local PDF:** `research/papers/park2015_advmat_hzo.pdf`\n",
                encoding="utf-8",
            )

            code = run_promote_pdf(
                root=root,
                key="park2015_advmat_hzo",
                destination="research/papers/promoted.pdf",
                source="",
                license_name="CC BY 4.0",
                license_url="https://example.org/license",
                review_note="Publisher page shows redistribution-compatible license.",
            )

            self.assertEqual(code, 1)
            self.assertFalse((root / "research" / "papers" / "promoted.pdf").exists())
            citation_text = citation.read_text(encoding="utf-8")
            self.assertIn("**PDF:** `not stored`", citation_text)
            report = json.loads((root / "research" / "reports" / "pdf-promotion-latest.json").read_text())
            self.assertEqual(report["status"], "failed")
            self.assertIn("tracked canonical path", report["message"])

    def test_promote_pdf_requires_existing_source_pdf(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text(
                "**Key:** `park2015_advmat_hzo`\n"
                "**PDF:** `not stored`\n"
                "**Local PDF:** `research/papers/missing.pdf`\n",
                encoding="utf-8",
            )

            code = run_promote_pdf(
                root=root,
                key="park2015_advmat_hzo",
                destination="docs/4-research/papers/by-topic/01-ferroelectric-materials/park2015_advmat_hzo.pdf",
                source="",
                license_name="CC BY 4.0",
                license_url="https://example.org/license",
                review_note="Publisher page shows redistribution-compatible license.",
            )

            self.assertEqual(code, 1)
            report = json.loads((root / "research" / "reports" / "pdf-promotion-latest.json").read_text())
            self.assertEqual(report["status"], "failed")
            self.assertIn("source PDF does not exist", report["message"])


if __name__ == "__main__":
    unittest.main()
