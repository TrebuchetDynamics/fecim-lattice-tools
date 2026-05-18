from __future__ import annotations

from pathlib import Path
import hashlib
import json


def write_content_addressed_report(
    root: Path,
    latest_path: str,
    history_dir: str,
    payload: dict[str, object],
) -> dict[str, object]:
    canonical_payload = json.dumps(payload, sort_keys=True, separators=(",", ":")).encode("utf-8")
    run_id = hashlib.sha256(canonical_payload).hexdigest()[:16]
    history_path = f"{history_dir.rstrip('/')}/{run_id}.json"
    addressed_payload = {**payload, "run_id": run_id, "history_path": history_path}
    text = json.dumps(addressed_payload, indent=2, sort_keys=True) + "\n"

    latest = root / latest_path
    history = root / history_path
    latest.parent.mkdir(parents=True, exist_ok=True)
    history.parent.mkdir(parents=True, exist_ok=True)
    latest.write_text(text, encoding="utf-8")
    history.write_text(text, encoding="utf-8")
    return addressed_payload
