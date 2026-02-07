#!/usr/bin/env python3
r"""Compute basic streaming statistics for FeCIM LK hysteresis CSV logs.

Designed to work without pandas/numpy.

Examples:
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --columns e_field_v_m,polarization_c_m2
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --group-by waveform
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --group-by wrd_phase_name --where 'waveform=ISPP (Write/Read)'

Output:
  - Per-file summary
  - Optional per-group summaries
  - NaN/Inf counters
"""

from __future__ import annotations

import argparse
import csv
import glob
import math
import os
import sys
from dataclasses import dataclass, field
from typing import Dict, Iterable, List, Optional, Tuple


@dataclass
class RunningStats:
    n: int = 0
    mean: float = 0.0
    m2: float = 0.0
    min: float = float("inf")
    max: float = float("-inf")
    nan: int = 0
    inf: int = 0

    def push(self, x: float) -> None:
        if math.isnan(x):
            self.nan += 1
            return
        if math.isinf(x):
            self.inf += 1
            return

        self.n += 1
        if x < self.min:
            self.min = x
        if x > self.max:
            self.max = x

        # Welford
        delta = x - self.mean
        self.mean += delta / self.n
        delta2 = x - self.mean
        self.m2 += delta * delta2

    @property
    def var(self) -> float:
        if self.n < 2:
            return 0.0
        return self.m2 / (self.n - 1)

    @property
    def std(self) -> float:
        return math.sqrt(self.var)


def parse_where(where: Optional[str]) -> Optional[Tuple[str, str]]:
    if not where:
        return None
    if "=" not in where:
        raise ValueError("--where must look like col=value")
    k, v = where.split("=", 1)
    return k.strip(), v.strip()


def is_number(s: str) -> bool:
    if s is None:
        return False
    s = s.strip()
    if s == "":
        return False
    try:
        float(s)
        return True
    except Exception:
        return False


def fmt(x: float) -> str:
    if math.isnan(x):
        return "NaN"
    if math.isinf(x):
        return "Inf" if x > 0 else "-Inf"
    # keep readable for big/small
    ax = abs(x)
    if ax != 0 and (ax >= 1e6 or ax < 1e-3):
        return f"{x:.6e}"
    return f"{x:.6f}"


def summarize(stats_by_col: Dict[str, RunningStats], columns: List[str]) -> str:
    lines = []
    header = f"{'column':28} {'n':>9} {'min':>14} {'max':>14} {'mean':>14} {'std':>14} {'var':>14} {'nan':>6} {'inf':>6}"
    lines.append(header)
    lines.append("-" * len(header))
    for col in columns:
        st = stats_by_col.get(col)
        if not st:
            continue
        lines.append(
            f"{col:28} {st.n:9d} {fmt(st.min):>14} {fmt(st.max):>14} {fmt(st.mean):>14} {fmt(st.std):>14} {fmt(st.var):>14} {st.nan:6d} {st.inf:6d}"
        )
    return "\n".join(lines)


def iter_files(patterns: List[str]) -> List[str]:
    files: List[str] = []
    for p in patterns:
        expanded = glob.glob(p)
        if expanded:
            files.extend(expanded)
        else:
            files.append(p)
    # uniq + stable
    seen = set()
    out = []
    for f in files:
        if f not in seen:
            seen.add(f)
            out.append(f)
    return out


def main(argv: List[str]) -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("paths", nargs="+", help="CSV log paths (globs ok).")
    ap.add_argument(
        "--columns",
        help=(
            "Comma-separated numeric columns to summarize. Default: auto-detect a useful subset if present."
        ),
    )
    ap.add_argument(
        "--group-by",
        help="Optional column name to group summaries by (e.g., waveform, wrd_phase_name, controller_state).",
    )
    ap.add_argument(
        "--where",
        help="Optional filter like col=value (exact string match) applied before stats (e.g., waveform=LK_SWEEP).",
    )
    ap.add_argument(
        "--max-groups",
        type=int,
        default=25,
        help="Cap number of groups printed (default 25).",
    )

    args = ap.parse_args(argv)
    where = parse_where(args.where)

    files = iter_files(args.paths)
    if not files:
        print("No files matched.", file=sys.stderr)
        return 2

    for path in files:
        if not os.path.exists(path):
            print(f"Missing: {path}", file=sys.stderr)
            continue

        print(f"\n=== {path} ===")
        with open(path, "r", newline="") as f:
            reader = csv.DictReader(f)
            if not reader.fieldnames:
                print("(empty)")
                continue
            fieldnames = list(reader.fieldnames)

            # default columns: a curated set if present
            default_cols = [
                "sim_time_s",
                "dt_s",
                "e_field_v_m",
                "e_field_mv_cm",
                "polarization_c_m2",
                "polarization_uc_cm2",
                "normalized_p",
                "controller_current_field_v_m",
                "controller_current_field_mv_cm",
                "controller_phase_timer_s",
                "controller_pulse_count",
                "controller_retry_count",
                "controller_overshoot_count",
                "controller_overshoot_total",
                "wrd_cycle_energy_fj",
                "wrd_retry_count",
            ]
            if args.columns:
                columns = [c.strip() for c in args.columns.split(",") if c.strip()]
            else:
                columns = [c for c in default_cols if c in fieldnames]
                if not columns:
                    # fallback: take first 12 numeric-looking columns
                    columns = []
                    for c in fieldnames:
                        if c in ("timestamp", "waveform", "material", "wrd_phase_name", "controller_state"):
                            continue
                        columns.append(c)
                        if len(columns) >= 12:
                            break

            stats: Dict[str, RunningStats] = {c: RunningStats() for c in columns}
            grouped: Dict[str, Dict[str, RunningStats]] = {}
            group_counts: Dict[str, int] = {}

            n_rows = 0
            n_kept = 0
            for row in reader:
                n_rows += 1

                if where is not None:
                    wk, wv = where
                    if row.get(wk, "") != wv:
                        continue

                n_kept += 1
                gkey = None
                if args.group_by:
                    gkey = row.get(args.group_by, "")
                    if gkey not in grouped:
                        grouped[gkey] = {c: RunningStats() for c in columns}
                        group_counts[gkey] = 0
                    group_counts[gkey] += 1

                for c in columns:
                    raw = row.get(c, "")
                    if raw is None or raw == "":
                        continue
                    if not is_number(raw):
                        continue
                    x = float(raw)
                    stats[c].push(x)
                    if gkey is not None:
                        grouped[gkey][c].push(x)

            print(f"rows: {n_rows} (kept: {n_kept})")
            print(summarize(stats, columns))

            if args.group_by:
                keys = sorted(grouped.keys(), key=lambda k: group_counts.get(k, 0), reverse=True)
                if len(keys) > args.max_groups:
                    print(f"\n(grouped by {args.group_by}: showing top {args.max_groups} of {len(keys)} groups by row count)")
                    keys = keys[: args.max_groups]
                else:
                    print(f"\n(grouped by {args.group_by}: {len(keys)} groups)")

                for k in keys:
                    print(f"\n--- {args.group_by}={k!r} (rows={group_counts.get(k, 0)}) ---")
                    print(summarize(grouped[k], columns))

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
