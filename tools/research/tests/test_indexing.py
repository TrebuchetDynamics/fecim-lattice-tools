import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.indexing import collect_chunk_files, write_index_manifest


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


if __name__ == "__main__":
    unittest.main()
