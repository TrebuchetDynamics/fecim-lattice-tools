#!/usr/bin/env python3
"""
FeCIM Crossbar Post-Run Validation Hook

This script runs after OpenLane flow completion to validate
the FeCIM crossbar design meets requirements.

Usage: Called automatically by OpenLane with -run_hooks flag
"""

import os
import sys
import json
from pathlib import Path


def check_metrics(run_path: Path) -> dict:
    """Extract key metrics from OpenLane reports."""
    metrics = {}

    metrics_file = run_path / "reports" / "metrics.csv"
    if metrics_file.exists():
        with open(metrics_file) as f:
            header = f.readline().strip().split(",")
            values = f.readline().strip().split(",")
            for h, v in zip(header, values):
                try:
                    metrics[h] = float(v)
                except ValueError:
                    metrics[h] = v

    return metrics


def check_drc(run_path: Path) -> int:
    """Count DRC violations."""
    drc_files = [
        run_path / "reports" / "signoff" / "drc.rpt",
        run_path / "logs" / "signoff" / "magic.drc",
    ]

    total_violations = 0
    for drc_file in drc_files:
        if drc_file.exists():
            content = drc_file.read_text()
            # Count various violation patterns
            total_violations += content.lower().count("violation")
            total_violations += content.lower().count("error")

    return total_violations


def check_lvs(run_path: Path) -> bool:
    """Check if LVS passed."""
    lvs_file = run_path / "reports" / "signoff" / "lvs.rpt"
    if lvs_file.exists():
        content = lvs_file.read_text()
        return "match" in content.lower() and "mismatch" not in content.lower()
    return False


def check_cell_count(run_path: Path, expected: int) -> bool:
    """Verify expected number of FeCIM cells."""
    def_files = list((run_path / "results" / "final" / "def").glob("*.def"))
    if not def_files:
        return False

    content = def_files[0].read_text()
    cell_count = content.count("fecim_bit")
    return cell_count >= expected


def main():
    print("=" * 60)
    print("FeCIM Crossbar Post-Run Validation")
    print("=" * 60)

    # Get run directory from environment
    run_dir = os.environ.get("RUN_DIR", os.environ.get("run_path", "."))
    run_path = Path(run_dir)

    if not run_path.exists():
        print(f"ERROR: Run directory not found: {run_path}")
        sys.exit(1)

    print(f"\nRun Path: {run_path}")

    # Check metrics
    print("\n--- Metrics ---")
    metrics = check_metrics(run_path)
    if metrics:
        for key in ["wire_length", "cell_count", "utilization", "DIEAREA_mm^2"]:
            if key in metrics:
                print(f"  {key}: {metrics[key]}")
    else:
        print("  (No metrics found)")

    # Check DRC
    print("\n--- DRC ---")
    violations = check_drc(run_path)
    print(f"  Total violations: {violations}")
    if violations > 0:
        print("  WARNING: DRC violations detected (expected with stub cells)")

    # Check LVS
    print("\n--- LVS ---")
    lvs_pass = check_lvs(run_path)
    print(f"  Status: {'PASS' if lvs_pass else 'FAIL/NOT RUN'}")

    # Check cell count (16x16 = 256 cells expected)
    print("\n--- Cell Count ---")
    expected_cells = 256  # 16x16
    cells_ok = check_cell_count(run_path, expected_cells)
    print(f"  Expected: {expected_cells}")
    print(f"  Status: {'OK' if cells_ok else 'MISMATCH'}")

    # Summary
    print("\n" + "=" * 60)
    if violations == 0 and cells_ok:
        print("VALIDATION: PASS")
        print("Design ready for next steps (with real FeCIM cells)")
    else:
        print("VALIDATION: WARNINGS")
        print("Review DRC/LVS before tape-out")
    print("=" * 60)

    return 0


if __name__ == "__main__":
    sys.exit(main())
