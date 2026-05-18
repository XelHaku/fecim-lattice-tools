import json
import hashlib
import tempfile
import unittest
from pathlib import Path

from fecim_research.claims import audit_claim_registry, load_claim_records, run_audit


class ClaimsTest(unittest.TestCase):
    def test_load_claim_records_reads_reviewed_yaml_files(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            claim_path = root / "citations" / "claims" / "hzo-remanent-polarization-range.yaml"
            claim_path.parent.mkdir(parents=True)
            claim_path.write_text(
                "id: hzo-remanent-polarization-range\n"
                "claim: HZO devices commonly report remanent polarization in a bounded literature range.\n"
                "status: literature-backed\n"
                "sources:\n"
                "  - park2015_advmat_hzo\n"
                "used_in:\n"
                "  - config/materials.yaml\n"
                "confidence: medium\n",
                encoding="utf-8",
            )

            records = load_claim_records(root)

            self.assertEqual(["hzo-remanent-polarization-range"], sorted(records))
            self.assertEqual(records["hzo-remanent-polarization-range"].sources, ["park2015_advmat_hzo"])
            self.assertEqual(records["hzo-remanent-polarization-range"].used_in, ["config/materials.yaml"])

    def test_audit_passes_for_claim_with_existing_source_and_used_path(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            (root / "config").mkdir()
            (root / "config" / "materials.yaml").write_text("# [claim: hzo-remanent-polarization-range]\n")
            (root / "citations" / "facts.md").write_text(
                "- HZO Pr range is cited. [claim: hzo-remanent-polarization-range]\n",
                encoding="utf-8",
            )

            report = audit_claim_registry(root)

            self.assertTrue(report.ok, report.errors)
            self.assertEqual(report.claims_checked, 1)

    def test_audit_fails_for_missing_source_or_used_path(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["missing_source"],
                used_in=["config/missing.yaml"],
            )

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn("missing source missing_source", "\n".join(report.errors))
            self.assertIn("missing used_in path config/missing.yaml", "\n".join(report.errors))

    def test_audit_fails_when_used_path_does_not_reference_claim_id(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            (root / "config").mkdir()
            (root / "config" / "materials.yaml").write_text("# missing claim marker\n")

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn("config/materials.yaml does not reference [claim: hzo-remanent-polarization-range]", "\n".join(report.errors))

    def test_audit_fails_when_facts_reference_unknown_claim(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            facts = root / "citations" / "facts.md"
            facts.parent.mkdir(parents=True)
            facts.write_text("- A cited-looking fact. [claim: missing-claim]\n", encoding="utf-8")

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn("unknown claim id missing-claim", "\n".join(report.errors))

    def test_disputed_claims_cannot_be_promoted_to_facts(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "unstable-hzo-range",
                status="disputed",
                sources=["park2015_advmat_hzo"],
                used_in=["citations/disputed.md"],
            )
            disputed = root / "citations" / "disputed.md"
            disputed.parent.mkdir(parents=True, exist_ok=True)
            disputed.write_text("- Disputed claim lives here. [claim: unstable-hzo-range]\n", encoding="utf-8")
            (root / "citations" / "facts.md").write_text(
                "- Disputed claim promoted. [claim: unstable-hzo-range]\n",
                encoding="utf-8",
            )

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn("disputed claim unstable-hzo-range is referenced from citations/facts.md", "\n".join(report.errors))

    def test_audit_fails_when_citation_record_pdf_path_is_missing(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo", pdf="docs/4-research/papers/missing.pdf")

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn(
                "citations/papers/park2015_advmat_hzo.md PDF path docs/4-research/papers/missing.pdf does not exist",
                "\n".join(report.errors),
            )

    def test_audit_fails_when_citation_pdf_points_at_ignored_research_inbox(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf_path = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf_path.parent.mkdir(parents=True)
            pdf_path.write_bytes(b"%PDF-1.7\n")
            self._write_paper(root, "park2015_advmat_hzo", pdf="research/papers/park2015_advmat_hzo.pdf")

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn(
                "citations/papers/park2015_advmat_hzo.md PDF path research/papers/park2015_advmat_hzo.pdf points at ignored local inbox; use not stored until promoted",
                "\n".join(report.errors),
            )

    def test_audit_fails_when_citation_record_pdf_path_is_not_repo_relative(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf_path = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf_path.parent.mkdir(parents=True)
            pdf_path.write_bytes(b"%PDF-1.7\n")
            self._write_paper(root, "park2015_advmat_hzo", pdf=str(pdf_path))

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn(
                "citations/papers/park2015_advmat_hzo.md PDF path must be repo-relative",
                "\n".join(report.errors),
            )

    def test_audit_passes_for_source_ledger_with_existing_pdf_digest(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf_bytes = b"%PDF-1.7\nsource ledger\n"
            digest = hashlib.sha256(pdf_bytes).hexdigest()
            pdf_path = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf_path.parent.mkdir(parents=True)
            pdf_path.write_bytes(pdf_bytes)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_source_ledger(
                root,
                "park2015_advmat_hzo",
                citation_path="citations/papers/park2015_advmat_hzo.md",
                pdf_path="research/papers/park2015_advmat_hzo.pdf",
                sha256=digest,
            )

            report = audit_claim_registry(root)

            self.assertTrue(report.ok, report.errors)

    def test_audit_fails_when_source_ledger_pdf_digest_does_not_match(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf_path = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf_path.parent.mkdir(parents=True)
            pdf_path.write_bytes(b"%PDF-1.7\nsource ledger\n")
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_source_ledger(
                root,
                "park2015_advmat_hzo",
                citation_path="citations/papers/park2015_advmat_hzo.md",
                pdf_path="research/papers/park2015_advmat_hzo.pdf",
                sha256="stale-digest",
            )

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn(
                "research/sources/park2015_advmat_hzo.yaml pdf sha256 stale-digest does not match actual",
                "\n".join(report.errors),
            )

    def test_audit_fails_when_source_ledger_citation_path_is_missing(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf_bytes = b"%PDF-1.7\nsource ledger\n"
            digest = hashlib.sha256(pdf_bytes).hexdigest()
            pdf_path = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf_path.parent.mkdir(parents=True)
            pdf_path.write_bytes(pdf_bytes)
            self._write_source_ledger(
                root,
                "park2015_advmat_hzo",
                citation_path="citations/papers/missing.md",
                pdf_path="research/papers/park2015_advmat_hzo.pdf",
                sha256=digest,
            )

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn(
                "research/sources/park2015_advmat_hzo.yaml citation_path citations/papers/missing.md does not exist",
                "\n".join(report.errors),
            )

    def test_audit_passes_for_candidate_evidence_with_existing_claim_and_chunk(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            (root / "config").mkdir()
            (root / "config" / "materials.yaml").write_text("# [claim: hzo-remanent-polarization-range]\n")
            self._write_chunk(
                root,
                "research/chunks/park2015_advmat_hzo.jsonl",
                {"id": "park2015_advmat_hzo::intro::chunk-001", "sha256": "chunk-digest"},
            )
            self._write_evidence(
                root,
                "hzo-remanent-polarization-range",
                candidate_count=1,
                candidates=[
                    {
                        "docid": "park2015_advmat_hzo::intro::chunk-001",
                        "chunk_file": "research/chunks/park2015_advmat_hzo.jsonl",
                        "sha256": "chunk-digest",
                    }
                ],
            )

            report = audit_claim_registry(root)

            self.assertTrue(report.ok, report.errors)

    def test_audit_fails_for_evidence_with_unknown_claim_and_missing_chunk(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_evidence(
                root,
                "missing-claim",
                candidate_count=1,
                candidates=[
                    {
                        "docid": "missing-claim::intro::chunk-001",
                        "chunk_file": "research/chunks/missing.jsonl",
                        "sha256": "chunk-digest",
                    }
                ],
            )

            report = audit_claim_registry(root)

            errors = "\n".join(report.errors)
            self.assertFalse(report.ok)
            self.assertIn("research/evidence/missing-claim.json references unknown claim missing-claim", errors)
            self.assertIn("candidate missing-claim::intro::chunk-001 missing chunk file research/chunks/missing.jsonl", errors)

    def test_audit_fails_for_evidence_candidate_count_or_digest_mismatch(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            (root / "config").mkdir()
            (root / "config" / "materials.yaml").write_text("# [claim: hzo-remanent-polarization-range]\n")
            self._write_chunk(
                root,
                "research/chunks/park2015_advmat_hzo.jsonl",
                {"id": "park2015_advmat_hzo::intro::chunk-001", "sha256": "actual-digest"},
            )
            self._write_evidence(
                root,
                "hzo-remanent-polarization-range",
                candidate_count=2,
                candidates=[
                    {
                        "docid": "park2015_advmat_hzo::intro::chunk-001",
                        "chunk_file": "research/chunks/park2015_advmat_hzo.jsonl",
                        "sha256": "stale-digest",
                    }
                ],
            )

            report = audit_claim_registry(root)

            errors = "\n".join(report.errors)
            self.assertFalse(report.ok)
            self.assertIn("candidate_count 2 does not match candidates length 1", errors)
            self.assertIn("candidate park2015_advmat_hzo::intro::chunk-001 sha256 stale-digest does not match chunk sha256 actual-digest", errors)

    def test_audit_fails_for_evidence_candidate_with_absolute_chunk_path(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            (root / "config").mkdir()
            (root / "config" / "materials.yaml").write_text("# [claim: hzo-remanent-polarization-range]\n")
            rel_chunk_path = "research/chunks/park2015_advmat_hzo.jsonl"
            self._write_chunk(
                root,
                rel_chunk_path,
                {"id": "park2015_advmat_hzo::intro::chunk-001", "sha256": "chunk-digest"},
            )
            self._write_evidence(
                root,
                "hzo-remanent-polarization-range",
                candidate_count=1,
                candidates=[
                    {
                        "docid": "park2015_advmat_hzo::intro::chunk-001",
                        "chunk_file": str(root / rel_chunk_path),
                        "sha256": "chunk-digest",
                    }
                ],
            )

            report = audit_claim_registry(root)

            self.assertFalse(report.ok)
            self.assertIn("candidate park2015_advmat_hzo::intro::chunk-001 chunk_file must be repo-relative", "\n".join(report.errors))

    def test_run_audit_writes_git_trackable_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            (root / "config").mkdir()
            (root / "config" / "materials.yaml").write_text("# [claim: hzo-remanent-polarization-range]\n")

            code = run_audit(root)

            self.assertEqual(code, 0)
            report_path = root / "research" / "reports" / "claim-audit-latest.json"
            self.assertTrue(report_path.exists())
            payload = json.loads(report_path.read_text())
            self.assertTrue(payload["ok"])
            self.assertEqual(payload["claims_checked"], 1)

    def _write_paper(self, root: Path, key: str, pdf: str = "not stored"):
        path = root / "citations" / "papers" / f"{key}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(f"**Key:** `{key}`\n**PDF:** `{pdf}`\n", encoding="utf-8")

    def _write_claim(
        self,
        root: Path,
        claim_id: str,
        sources: list[str],
        used_in: list[str],
        status: str = "literature-backed",
    ):
        path = root / "citations" / "claims" / f"{claim_id}.yaml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            f"id: {claim_id}\n"
            "claim: HZO devices commonly report remanent polarization in a bounded literature range.\n"
            f"status: {status}\n"
            "sources:\n"
            + "".join(f"  - {source}\n" for source in sources)
            + "used_in:\n"
            + "".join(f"  - {item}\n" for item in used_in)
            + "confidence: medium\n",
            encoding="utf-8",
        )

    def _write_chunk(self, root: Path, rel_path: str, record: dict[str, object]):
        path = root / rel_path
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(record, sort_keys=True) + "\n", encoding="utf-8")

    def _write_evidence(
        self,
        root: Path,
        claim_id: str,
        candidate_count: int,
        candidates: list[dict[str, object]],
    ):
        path = root / "research" / "evidence" / f"{claim_id}.json"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            json.dumps(
                {
                    "claim": {
                        "id": claim_id,
                        "claim": "HZO devices commonly report remanent polarization in a bounded literature range.",
                    },
                    "status": "candidate-evidence",
                    "review": {"state": "needs-review"},
                    "source_report": "research/reports/search-latest.json",
                    "query": "HZO remanent polarization range",
                    "backend": "local-jsonl",
                    "candidate_count": candidate_count,
                    "candidates": candidates,
                },
                indent=2,
                sort_keys=True,
            )
            + "\n",
            encoding="utf-8",
        )

    def _write_source_ledger(
        self,
        root: Path,
        paper_key: str,
        citation_path: str,
        pdf_path: str,
        sha256: str,
    ):
        path = root / "research" / "sources" / f"{paper_key}.yaml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            f"citation_path: {citation_path}\n"
            "doi: needs-review\n"
            "match:\n"
            "  confidence: 0.95\n"
            "  method: filename\n"
            "  status: matched\n"
            f"paper_key: {paper_key}\n"
            "pdf:\n"
            f"  path: {pdf_path}\n"
            f"  sha256: {sha256}\n"
            "  size: 24\n"
            "title: Park HZO\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    unittest.main()
