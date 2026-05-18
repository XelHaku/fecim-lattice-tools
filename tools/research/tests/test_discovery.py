import tempfile
import unittest
from pathlib import Path

from fecim_research.citations import load_citation_records
from fecim_research.discovery import discover_pdfs, match_pdf_to_record, sha256_file


class DiscoveryTest(unittest.TestCase):
    def test_discovers_pdf_roots_and_hashes_files(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF-1.4\nfixture\n")

            found = discover_pdfs(root, extra_paths=[])
            self.assertEqual(len(found), 1)
            self.assertEqual(found[0].path, pdf)
            self.assertEqual(found[0].sha256, sha256_file(pdf))

    def test_loads_existing_citation_keys_from_markdown(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text(
                "# Park 2015\n\n"
                "**Key:** `park2015_advmat_hzo`\n"
                "**DOI:** `10.1002/adma.201404531`\n"
                "**Title:** `Ferroelectric HZO`\n",
                encoding="utf-8",
            )
            records = load_citation_records(root)
            self.assertIn("park2015_advmat_hzo", records)
            self.assertEqual(records["park2015_advmat_hzo"].doi, "10.1002/adma.201404531")

    def test_uses_markdown_h1_as_title_fallback(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text(
                "# Ferroelectricity in Hafnium Oxide\n\n"
                "**Key:** `park2015_advmat_hzo`\n",
                encoding="utf-8",
            )

            records = load_citation_records(root)
            self.assertEqual(
                records["park2015_advmat_hzo"].title,
                "Ferroelectricity in Hafnium Oxide",
            )

    def test_discovers_uppercase_pdf_extensions(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.PDF"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            found = discover_pdfs(root, extra_paths=[])
            self.assertEqual([item.path for item in found], [pdf])

    def test_matches_pdf_filename_to_existing_citation_key(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text("**Key:** `park2015_advmat_hzo`\n", encoding="utf-8")
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.paper_key, "park2015_advmat_hzo")
            self.assertEqual(match.status, "matched")
            self.assertEqual(match.method, "filename")

    def test_filename_matching_prefers_exact_key_before_substring(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            papers = root / "citations" / "papers"
            papers.mkdir(parents=True)
            (papers / "advmat_hzo.md").write_text("**Key:** `advmat_hzo`\n", encoding="utf-8")
            (papers / "park2015_advmat_hzo.md").write_text(
                "**Key:** `park2015_advmat_hzo`\n",
                encoding="utf-8",
            )
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.paper_key, "park2015_advmat_hzo")

    def test_filename_matching_prefers_longer_substring_key(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            papers = root / "citations" / "papers"
            papers.mkdir(parents=True)
            (papers / "park2015.md").write_text("**Key:** `park2015`\n", encoding="utf-8")
            (papers / "park2015_advmat_hzo.md").write_text(
                "**Key:** `park2015_advmat_hzo`\n",
                encoding="utf-8",
            )
            pdf = root / "research" / "papers" / "scan_park2015_advmat_hzo_copy.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.paper_key, "park2015_advmat_hzo")

    def test_unmatched_pdf_is_quarantined(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "research" / "papers" / "unknown.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records={})
            self.assertEqual(match.status, "unmatched")
            self.assertIsNone(match.paper_key)


if __name__ == "__main__":
    unittest.main()
