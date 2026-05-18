import json
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

    def _write_paper(self, root: Path, key: str):
        path = root / "citations" / "papers" / f"{key}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(f"**Key:** `{key}`\n", encoding="utf-8")

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


if __name__ == "__main__":
    unittest.main()
