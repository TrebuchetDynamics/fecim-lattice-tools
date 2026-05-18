import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.searching import load_chunk_lookup, render_text_results


class SearchingTest(unittest.TestCase):
    def test_load_chunk_lookup_reads_jsonl_chunks(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps({"id": "park::sec-01::chunk-001", "paper_key": "park", "contents": "HZO coercive field evidence"}) + "\n",
                encoding="utf-8",
            )
            lookup = load_chunk_lookup(root)
            self.assertIn("park::sec-01::chunk-001", lookup)

    def test_render_text_results_includes_score_key_and_snippet(self):
        rows = [
            {
                "rank": 1,
                "score": 7.5,
                "docid": "park::sec-01::chunk-001",
                "paper_key": "park",
                "section": "Results",
                "snippet": "HZO coercive field evidence",
            }
        ]
        text = render_text_results(rows)
        self.assertIn("1. park", text)
        self.assertIn("score=7.5", text)
        self.assertIn("HZO coercive field evidence", text)


if __name__ == "__main__":
    unittest.main()
