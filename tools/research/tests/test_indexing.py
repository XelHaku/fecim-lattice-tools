import json
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from fecim_research.indexing import collect_chunk_files, run_index, write_index_manifest


class IndexingTest(unittest.TestCase):
    def test_collect_chunk_files_sorted(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunks = root / "research" / "chunks"
            chunks.mkdir(parents=True)
            (chunks / "b.jsonl").write_text("{}", encoding="utf-8")
            (chunks / "a.jsonl").write_text("{}", encoding="utf-8")
            got = collect_chunk_files(root)
            self.assertEqual([p.name for p in got], ["a.jsonl", "b.jsonl"])

    def test_write_index_manifest_records_backend_and_inputs(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "a.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text("{\"id\":\"a\",\"contents\":\"text\"}\n", encoding="utf-8")
            path = write_index_manifest(root, "pyserini", [chunk], semantic=False, embedding_model="")
            data = json.loads(path.read_text(encoding="utf-8"))
            self.assertEqual(data["backend"], "pyserini")
            self.assertFalse(data["semantic"])
            self.assertEqual(len(data["inputs"]), 1)
            backend_manifest = root / "research" / "manifests" / "index-pyserini.json"
            self.assertTrue(backend_manifest.exists())
            self.assertEqual(json.loads(backend_manifest.read_text(encoding="utf-8")), data)
            self.assertEqual(json.loads((root / "research" / "manifests" / "index-latest.json").read_text(encoding="utf-8")), data)

    def test_run_index_semantic_writes_rebuildable_local_vector_cache_without_lancedb(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps(
                    {
                        "id": "park::sec-01::chunk-001",
                        "paper_key": "park",
                        "section": "Results",
                        "contents": "HZO coercive field evidence.",
                        "source_path": "research/parsed/park/marker.md",
                    }
                )
                + "\n",
                encoding="utf-8",
            )

            with patch.dict(sys.modules, {"lancedb": None}):
                code = run_index(root, semantic=True, embedding_model="")

            self.assertEqual(code, 0)
            cache = root / "research" / "index" / "lancedb" / "chunks.jsonl"
            self.assertTrue(cache.exists())
            rows = [json.loads(line) for line in cache.read_text(encoding="utf-8").splitlines()]
            self.assertEqual(rows[0]["id"], "park::sec-01::chunk-001")
            self.assertEqual(len(rows[0]["vector"]), 64)
            self.assertEqual(rows[0]["embedding_model"], "fecim-hashing-bow-v1")
            manifest = json.loads((root / "research" / "manifests" / "index-latest.json").read_text(encoding="utf-8"))
            self.assertEqual(manifest["backend"], "local-vector-jsonl")
            self.assertTrue(manifest["semantic"])
            self.assertEqual(manifest["embedding_model"], "fecim-hashing-bow-v1")
            self.assertEqual(manifest["embedding_provider"], "local-hashing")
            self.assertFalse(manifest["external_ai"])
            self.assertEqual(manifest["vector_dimension"], 64)
            self.assertEqual(manifest["lancedb_index"], "research/index/lancedb")
            backend_manifest = json.loads((root / "research" / "manifests" / "index-lancedb.json").read_text(encoding="utf-8"))
            self.assertEqual(backend_manifest, manifest)


if __name__ == "__main__":
    unittest.main()
