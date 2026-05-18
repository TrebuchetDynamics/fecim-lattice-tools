import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.cache import build_cache_report, run_cache
from fecim_research.indexing import write_index_manifest


class CacheTest(unittest.TestCase):
    def test_build_cache_report_marks_missing_pyserini_cache_rebuildable(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_research_gitignore(root)

            report = build_cache_report(root)

            self.assertFalse(report["ok"])
            pyserini = self._cache(report, "pyserini")
            self.assertEqual(pyserini["path"], "research/index/pyserini")
            self.assertTrue(pyserini["rebuildable"])
            self.assertTrue(pyserini["ignored_by_policy"])
            self.assertFalse(pyserini["exists"])
            self.assertFalse(pyserini["manifest_exists"])
            self.assertEqual(pyserini["status"], "missing")
            self.assertEqual(pyserini["rebuild_command"], "fecim research index")

    def test_build_cache_report_marks_pyserini_cache_ready_when_manifest_inputs_match(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_research_gitignore(root)
            chunk = root / "research" / "chunks" / "paper.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text('{"id":"paper::chunk-001","contents":"text"}\n', encoding="utf-8")
            write_index_manifest(root, "pyserini", [chunk], semantic=False, embedding_model="")
            (root / "research" / "index" / "pyserini").mkdir(parents=True)

            report = build_cache_report(root)

            pyserini = self._cache(report, "pyserini")
            self.assertTrue(report["ok"])
            self.assertEqual(pyserini["status"], "ready")
            self.assertTrue(pyserini["exists"])
            self.assertTrue(pyserini["manifest_exists"])
            self.assertFalse(pyserini["stale"])
            self.assertEqual(pyserini["inputs"], 1)

    def test_build_cache_report_marks_pyserini_cache_stale_when_input_hash_changes(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_research_gitignore(root)
            chunk = root / "research" / "chunks" / "paper.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text('{"id":"paper::chunk-001","contents":"old"}\n', encoding="utf-8")
            write_index_manifest(root, "pyserini", [chunk], semantic=False, embedding_model="")
            chunk.write_text('{"id":"paper::chunk-001","contents":"new"}\n', encoding="utf-8")
            (root / "research" / "index" / "pyserini").mkdir(parents=True)

            report = build_cache_report(root)

            pyserini = self._cache(report, "pyserini")
            self.assertFalse(report["ok"])
            self.assertEqual(pyserini["status"], "stale")
            self.assertTrue(pyserini["stale"])
            self.assertEqual(pyserini["stale_inputs"], ["research/chunks/paper.jsonl"])

    def test_run_cache_writes_git_trackable_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_research_gitignore(root)

            code = run_cache(root)

            self.assertEqual(code, 1)
            report_path = root / "research" / "reports" / "cache-latest.json"
            self.assertTrue(report_path.exists())
            report = json.loads(report_path.read_text(encoding="utf-8"))
            self.assertIn("pyserini", [cache["name"] for cache in report["caches"]])

    def _cache(self, report: dict[str, object], name: str) -> dict[str, object]:
        for cache in report["caches"]:
            if cache["name"] == name:
                return cache
        raise AssertionError(f"cache {name} not found")

    def _write_research_gitignore(self, root: Path) -> None:
        path = root / "research" / ".gitignore"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            "/papers/**/*.pdf\n"
            "/index/pyserini/\n"
            "/index/lancedb/\n"
            "/index/models/\n"
            "/.cache/\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    unittest.main()
