import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.graphing import build_provenance_graph, run_graph


class GraphingTest(unittest.TestCase):
    def test_build_provenance_graph_links_sources_claims_and_used_files(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["config/materials.yaml"],
            )
            config = root / "config" / "materials.yaml"
            config.parent.mkdir(parents=True)
            config.write_text("# [claim: hzo-remanent-polarization-range]\n", encoding="utf-8")

            graph = build_provenance_graph(root)

            node_ids = {node["id"] for node in graph["nodes"]}
            self.assertIn("paper:park2015_advmat_hzo", node_ids)
            self.assertIn("claim:hzo-remanent-polarization-range", node_ids)
            self.assertIn("file:config/materials.yaml", node_ids)
            edges = {(edge["from"], edge["to"], edge["relation"]) for edge in graph["edges"]}
            self.assertIn(
                ("paper:park2015_advmat_hzo", "claim:hzo-remanent-polarization-range", "supports"),
                edges,
            )
            self.assertIn(
                ("claim:hzo-remanent-polarization-range", "file:config/materials.yaml", "used_in"),
                edges,
            )
            self.assertIn(
                ("file:config/materials.yaml", "claim:hzo-remanent-polarization-range", "references"),
                edges,
            )

    def test_run_graph_writes_git_trackable_graph_and_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "park2015_advmat_hzo")
            self._write_claim(
                root,
                "hzo-remanent-polarization-range",
                sources=["park2015_advmat_hzo"],
                used_in=["docs/TRUST.md"],
            )
            docs = root / "docs"
            docs.mkdir()
            (docs / "TRUST.md").write_text("[claim: hzo-remanent-polarization-range]\n", encoding="utf-8")

            code = run_graph(root)

            self.assertEqual(code, 0)
            graph_path = root / "research" / "graphs" / "provenance-graph.json"
            report_path = root / "research" / "reports" / "graph-latest.json"
            self.assertTrue(graph_path.exists())
            self.assertTrue(report_path.exists())
            graph = json.loads(graph_path.read_text())
            report = json.loads(report_path.read_text())
            self.assertEqual(report["nodes"], len(graph["nodes"]))
            self.assertEqual(report["edges"], len(graph["edges"]))
            self.assertEqual(report["claims"], 1)
            self.assertEqual(report["sources"], 1)

    def _write_paper(self, root: Path, key: str):
        path = root / "citations" / "papers" / f"{key}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(f"**Key:** `{key}`\n", encoding="utf-8")

    def _write_claim(self, root: Path, claim_id: str, sources: list[str], used_in: list[str]):
        path = root / "citations" / "claims" / f"{claim_id}.yaml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            f"id: {claim_id}\n"
            "claim: HZO devices commonly report remanent polarization in a bounded literature range.\n"
            "status: literature-backed\n"
            "sources:\n"
            + "".join(f"  - {source}\n" for source in sources)
            + "used_in:\n"
            + "".join(f"  - {item}\n" for item in used_in)
            + "confidence: medium\n",
            encoding="utf-8",
        )


if __name__ == "__main__":
    unittest.main()
