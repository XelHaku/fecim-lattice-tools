import tempfile
import unittest
from pathlib import Path

from fecim_research.citations import load_citation_records
from fecim_research.discovery import discover_pdfs, match_pdf_to_record, sha256_file
from fecim_research.yamlio import dumps_yaml


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

    def test_load_citation_records_uses_filename_as_repo_identity(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text(
                "# Park 2015\n\n"
                "**Key:** `wrong_external_key`\n"
                "**DOI:** `10.1002/adma.201404531`\n",
                encoding="utf-8",
            )

            records = load_citation_records(root)

            self.assertIn("park2015_advmat_hzo", records)
            self.assertNotIn("wrong_external_key", records)
            self.assertEqual(records["park2015_advmat_hzo"].key, "park2015_advmat_hzo")

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

    def test_discovers_relative_and_absolute_extra_path_directories(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            relative_pdf = root / "incoming" / "relative.pdf"
            absolute_pdf = root / "outside" / "absolute.PDF"
            relative_pdf.parent.mkdir(parents=True)
            absolute_pdf.parent.mkdir(parents=True)
            relative_pdf.write_bytes(b"%PDF relative")
            absolute_pdf.write_bytes(b"%PDF absolute")

            found = discover_pdfs(root, extra_paths=[Path("incoming"), absolute_pdf.parent])
            self.assertEqual([item.path for item in found], [relative_pdf, absolute_pdf])

    def test_marks_later_duplicate_pdf_hashes_with_canonical_path(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            first = root / "research" / "papers" / "a_first.pdf"
            second = root / "research" / "papers" / "b_second.pdf"
            first.parent.mkdir(parents=True)
            first.write_bytes(b"%PDF duplicate")
            second.write_bytes(b"%PDF duplicate")

            found = discover_pdfs(root, extra_paths=[])
            self.assertEqual([item.path for item in found], [first, second])
            self.assertIsNone(found[0].duplicate_of)
            self.assertEqual(found[1].duplicate_of, first)

    def test_dumps_yaml_sorts_mapping_keys(self):
        self.assertEqual(
            dumps_yaml({"status": "matched", "paper_key": "park2015", "confidence": 0.95}),
            "confidence: 0.95\npaper_key: park2015\nstatus: matched\n",
        )

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

    def test_matches_explicit_citation_pdf_path_before_filename_heuristics(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text(
                "**Key:** `park2015_advmat_hzo`\n"
                "**PDF:** `docs/4-research/papers/by-topic/01-ferroelectric-materials/Reviewed_Park_2015.pdf`\n",
                encoding="utf-8",
            )
            pdf = root / "docs" / "4-research" / "papers" / "by-topic" / "01-ferroelectric-materials"
            pdf.mkdir(parents=True)
            reviewed = pdf / "Reviewed_Park_2015.pdf"
            reviewed.write_bytes(b"%PDF fixture")

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.paper_key, "park2015_advmat_hzo")
            self.assertEqual(match.status, "matched")
            self.assertEqual(match.method, "citation_pdf")

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

    def test_filename_matching_normalizes_case_and_punctuation(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "ibm_aihwkit_arxiv_2307_09357.md"
            paper.parent.mkdir(parents=True)
            paper.write_text("**Key:** `ibm_aihwkit_arxiv_2307_09357`\n", encoding="utf-8")
            pdf = root / "research" / "papers" / "IBM_AIHWKit_arXiv_2307.09357.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.paper_key, "ibm_aihwkit_arxiv_2307_09357")
            self.assertEqual(match.status, "matched")

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
