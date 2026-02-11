#!/usr/bin/env python3
r"""Compute basic streaming statistics for FeCIM LK hysteresis CSV logs.

Designed to work without pandas/numpy.

This tool is intentionally dependency-light and works in a streaming fashion.
Optionally, it can also estimate a few percentiles (p1/p50/p99) using a
streaming quantile estimator (P^2 algorithm), and print LK/ISPP-specific
"alarms" for suspicious controller behavior.

Examples:
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --columns e_field_v_m,polarization_c_m2
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --group-by waveform
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --group-by wrd_phase_name --where 'waveform=ISPP (Write/Read)'

Percentiles:
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --percentiles
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --percentiles --columns wrd_read_level,wrd_target_level

Alarms:
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --alarms
  python3 scripts/lk_log_stats.py logs/hysteresis-*.csv --alarms --stuck-rows 30

Output:
  - Per-file summary
  - Optional per-group summaries
  - Optional percentile estimates
  - Optional alarm sections
  - NaN/Inf counters
"""

from __future__ import annotations

import argparse
import csv
import glob
import math
import os
import sys
from dataclasses import dataclass
from typing import Dict, Iterable, List, Optional, Tuple


# --- Quantiles (streaming, no numpy) ----------------------------------------


class P2Quantile:
    """P^2 streaming quantile estimator.

    References:
      Jain and Chlamtac, "The P^2 algorithm for dynamic calculation of
      quantiles and histograms without storing observations" (1985).

    This is an *approximate* quantile estimator that is well suited for logs
    with many rows.
    """

    def __init__(self, p: float):
        if not (0.0 < p < 1.0):
            raise ValueError("p must be in (0, 1)")
        self.p = p
        self._init: List[float] = []  # first 5 samples

        # marker positions (n), desired positions (np), and increments (dn)
        self.n: List[int] = []
        self.np: List[float] = []
        self.dn: List[float] = []

        # marker heights
        self.q: List[float] = []

    def push(self, x: float) -> None:
        if len(self.q) == 0:
            self._init.append(x)
            if len(self._init) == 5:
                self._init.sort()
                self.q = self._init[:]
                # Positions are 1-indexed in the paper
                self.n = [1, 2, 3, 4, 5]
                self.np = [1.0, 1.0 + 2 * self.p, 1.0 + 4 * self.p, 3.0 + 2 * self.p, 5.0]
                self.dn = [0.0, self.p / 2.0, self.p, (1.0 + self.p) / 2.0, 1.0]
            return

        # Find cell k
        k = 0
        if x < self.q[0]:
            self.q[0] = x
            k = 0
        elif x < self.q[1]:
            k = 0
        elif x < self.q[2]:
            k = 1
        elif x < self.q[3]:
            k = 2
        elif x <= self.q[4]:
            k = 3
        else:
            self.q[4] = x
            k = 3

        # Increment positions
        for i in range(5):
            if i > k:
                self.n[i] += 1
        for i in range(5):
            self.np[i] += self.dn[i]

        # Adjust heights of markers 2..4 (index 1..3)
        for i in (1, 2, 3):
            d = self.np[i] - self.n[i]
            if (d >= 1 and self.n[i + 1] - self.n[i] > 1) or (d <= -1 and self.n[i - 1] - self.n[i] < -1):
                di = 1 if d >= 0 else -1

                # Parabolic prediction
                qip1, qi, qim1 = self.q[i + 1], self.q[i], self.q[i - 1]
                nip1, ni, nim1 = self.n[i + 1], self.n[i], self.n[i - 1]

                num = di * (ni - nim1 + di) * (qip1 - qi) / (nip1 - ni) + di * (nip1 - ni - di) * (qi - qim1) / (ni - nim1)
                den = (nip1 - nim1)
                q_new = qi + num / den

                # If parabolic goes out of bounds, use linear
                if q_new <= qim1 or q_new >= qip1:
                    q_new = qi + di * (self.q[i + di] - qi) / (self.n[i + di] - ni)

                self.q[i] = q_new
                self.n[i] += di

    def value(self) -> float:
        if len(self.q) == 0:
            return float("nan")
        if len(self.q) < 5:
            # Not enough samples: exact from init buffer
            s = sorted(self._init)
            if not s:
                return float("nan")
            idx = int(round(self.p * (len(s) - 1)))
            return s[min(max(idx, 0), len(s) - 1)]
        return self.q[2]  # middle marker estimates the quantile


@dataclass
class RunningStats:
    n: int = 0
    mean: float = 0.0
    m2: float = 0.0
    min: float = float("inf")
    max: float = float("-inf")
    nan: int = 0
    inf: int = 0

    # Optional percentiles
    q01: Optional[P2Quantile] = None
    q50: Optional[P2Quantile] = None
    q99: Optional[P2Quantile] = None

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

        if self.q01 is not None:
            self.q01.push(x)
        if self.q50 is not None:
            self.q50.push(x)
        if self.q99 is not None:
            self.q99.push(x)

    @property
    def var(self) -> float:
        if self.n < 2:
            return 0.0
        return self.m2 / (self.n - 1)

    @property
    def std(self) -> float:
        return math.sqrt(self.var)

    def p01(self) -> float:
        return self.q01.value() if self.q01 is not None else float("nan")

    def p50(self) -> float:
        return self.q50.value() if self.q50 is not None else float("nan")

    def p99(self) -> float:
        return self.q99.value() if self.q99 is not None else float("nan")


def make_running_stats(enable_percentiles: bool) -> RunningStats:
    if not enable_percentiles:
        return RunningStats()
    return RunningStats(q01=P2Quantile(0.01), q50=P2Quantile(0.50), q99=P2Quantile(0.99))


# --- Helpers ----------------------------------------------------------------


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
    ax = abs(x)
    if ax != 0 and (ax >= 1e6 or ax < 1e-3):
        return f"{x:.6e}"
    return f"{x:.6f}"


def summarize(stats_by_col: Dict[str, RunningStats], columns: List[str], *, show_percentiles: bool) -> str:
    lines = []
    if show_percentiles:
        header = (
            f"{'column':28} {'n':>9} {'min':>14} {'p01':>14} {'p50':>14} {'p99':>14} "
            f"{'max':>14} {'mean':>14} {'std':>14} {'var':>14} {'nan':>6} {'inf':>6}"
        )
    else:
        header = f"{'column':28} {'n':>9} {'min':>14} {'max':>14} {'mean':>14} {'std':>14} {'var':>14} {'nan':>6} {'inf':>6}"

    lines.append(header)
    lines.append("-" * len(header))
    for col in columns:
        st = stats_by_col.get(col)
        if not st:
            continue
        if show_percentiles:
            lines.append(
                f"{col:28} {st.n:9d} {fmt(st.min):>14} {fmt(st.p01()):>14} {fmt(st.p50()):>14} {fmt(st.p99()):>14} "
                f"{fmt(st.max):>14} {fmt(st.mean):>14} {fmt(st.std):>14} {fmt(st.var):>14} {st.nan:6d} {st.inf:6d}"
            )
        else:
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


def get_float(row: Dict[str, str], key: str) -> Optional[float]:
    raw = row.get(key, "")
    if raw is None or raw == "":
        return None
    if not is_number(raw):
        return None
    try:
        return float(raw)
    except Exception:
        return None


def get_str(row: Dict[str, str], key: str) -> str:
    v = row.get(key, "")
    return "" if v is None else str(v)


# --- Alarms (LK/ISPP-specific) ----------------------------------------------


@dataclass
class AlarmState:
    # Bounds
    bounds_rows: int = 0
    bounds_degenerate_rows: int = 0
    bounds_inverted_rows: int = 0
    vmin_ec_min: float = float("inf")
    vmin_ec_max: float = float("-inf")
    vmax_ec_min: float = float("inf")
    vmax_ec_max: float = float("-inf")

    # Overshoot/retry aggregation by (target_level, phase)
    overshoot_events: Dict[Tuple[str, str], int] = None  # type: ignore[assignment]
    overshoot_total: Dict[Tuple[str, str], float] = None  # type: ignore[assignment]
    retry_events: Dict[Tuple[str, str], int] = None  # type: ignore[assignment]
    retry_total: Dict[Tuple[str, str], float] = None  # type: ignore[assignment]

    # Stuck detection
    stuck_threshold_rows: int = 20
    stuck_max_reports: int = 10
    stuck_reports: List[str] = None  # type: ignore[assignment]

    # Current segment tracking
    _seg_key: Optional[Tuple[str, str]] = None
    _seg_start_step: Optional[int] = None
    _seg_last_error: Optional[float] = None
    _seg_nonimprove: int = 0
    _seg_last_read: Optional[float] = None
    _seg_last_target: Optional[float] = None

    def __post_init__(self) -> None:
        self.overshoot_events = {}
        self.overshoot_total = {}
        self.retry_events = {}
        self.retry_total = {}
        self.stuck_reports = []

    def _add_counter(self, d: Dict[Tuple[str, str], float], key: Tuple[str, str], v: float) -> None:
        d[key] = d.get(key, 0.0) + v

    def _add_int(self, d: Dict[Tuple[str, str], int], key: Tuple[str, str], v: int = 1) -> None:
        d[key] = d.get(key, 0) + v

    def push_row(self, row: Dict[str, str]) -> None:
        # Bounds degeneracy/inversion
        vmin_ec = get_float(row, "controller_vmin_ec")
        vmax_ec = get_float(row, "controller_vmax_ec")
        if vmin_ec is not None and vmax_ec is not None:
            self.bounds_rows += 1
            if vmin_ec < self.vmin_ec_min:
                self.vmin_ec_min = vmin_ec
            if vmin_ec > self.vmin_ec_max:
                self.vmin_ec_max = vmin_ec
            if vmax_ec < self.vmax_ec_min:
                self.vmax_ec_min = vmax_ec
            if vmax_ec > self.vmax_ec_max:
                self.vmax_ec_max = vmax_ec

            if vmin_ec == vmax_ec:
                self.bounds_degenerate_rows += 1
            if vmin_ec > vmax_ec:
                self.bounds_inverted_rows += 1

        # Overshoot/retry by target_level + phase
        target_level_s = get_str(row, "wrd_target_level")
        phase = get_str(row, "wrd_phase_name")
        key = (target_level_s, phase)

        # overshoot_count: event-ish, overshoot_total: magnitude-ish
        oc = get_float(row, "controller_overshoot_count")
        ot = get_float(row, "controller_overshoot_total")
        if oc is not None and oc > 0:
            self._add_int(self.overshoot_events, key, 1)
            self._add_counter(self.overshoot_total, key, oc)
        if ot is not None and ot > 0:
            # Track magnitude separately
            self._add_counter(self.overshoot_total, key, ot)

        # retries
        rc = get_float(row, "wrd_retry_count")
        crc = get_float(row, "controller_retry_count")
        # Some logs use controller_retry_count, some use wrd_retry_count; we aggregate both.
        rsum = 0.0
        if rc is not None and rc > 0:
            rsum += rc
        if crc is not None and crc > 0:
            rsum += crc
        if rsum > 0:
            self._add_int(self.retry_events, key, 1)
            self._add_counter(self.retry_total, key, rsum)

        # Stuck detection: track consecutive non-improving error towards target.
        # Reset segment when (target_level, phase) changes.
        # Use numeric target/read if available.
        try:
            step = int(float(get_str(row, "step") or "0"))
        except Exception:
            step = 0

        target = get_float(row, "wrd_target_level")
        read = get_float(row, "wrd_read_level")
        seg_key = (target_level_s, phase)

        if target is None or read is None:
            # Missing signal: break segment
            self._reset_segment()
            return

        err = abs(target - read)

        if self._seg_key != seg_key:
            self._reset_segment()
            self._seg_key = seg_key
            self._seg_start_step = step
            self._seg_last_error = err
            self._seg_last_read = read
            self._seg_last_target = target
            self._seg_nonimprove = 0
            return

        if self._seg_last_error is None:
            self._seg_last_error = err
            self._seg_last_read = read
            self._seg_last_target = target
            return

        # Non-improvement: error did not decrease.
        if err >= self._seg_last_error - 1e-12:
            self._seg_nonimprove += 1
        else:
            self._seg_nonimprove = 0

        self._seg_last_error = err
        self._seg_last_read = read
        self._seg_last_target = target

        if (
            self._seg_nonimprove >= self.stuck_threshold_rows
            and len(self.stuck_reports) < self.stuck_max_reports
        ):
            t_s = fmt(target)
            r_s = fmt(read)
            self.stuck_reports.append(
                f"stuck: target={t_s}, read={r_s}, err={fmt(err)}, rows_nonimprove={self._seg_nonimprove}, "
                f"phase={phase!r}, target_level={target_level_s!r}, step_start={self._seg_start_step}, step_now={step}"
            )
            # avoid spamming: require another full threshold before next report
            self._seg_nonimprove = 0

    def _reset_segment(self) -> None:
        self._seg_key = None
        self._seg_start_step = None
        self._seg_last_error = None
        self._seg_nonimprove = 0
        self._seg_last_read = None
        self._seg_last_target = None


def format_alarm_section(al: AlarmState) -> str:
    lines: List[str] = []

    # Bounds
    if al.bounds_rows > 0:
        lines.append("bounds:")
        lines.append(
            f"  controller_vmin_ec range: [{fmt(al.vmin_ec_min)}, {fmt(al.vmin_ec_max)}]"
        )
        lines.append(
            f"  controller_vmax_ec range: [{fmt(al.vmax_ec_min)}, {fmt(al.vmax_ec_max)}]"
        )
        if al.bounds_degenerate_rows > 0:
            lines.append(
                f"  ALARM: degenerate bounds (vmin==vmax): {al.bounds_degenerate_rows}/{al.bounds_rows} rows"
            )
        if al.bounds_inverted_rows > 0:
            lines.append(
                f"  ALARM: inverted bounds (vmin>vmax): {al.bounds_inverted_rows}/{al.bounds_rows} rows"
            )
    else:
        lines.append("bounds: (controller_vmin_ec/controller_vmax_ec not present)")

    # Overshoot/retries by key
    def topk(d: Dict[Tuple[str, str], float], k: int = 10) -> List[Tuple[Tuple[str, str], float]]:
        return sorted(d.items(), key=lambda kv: kv[1], reverse=True)[:k]

    if al.retry_events:
        lines.append("retries_by_target_and_phase (top):")
        for (tl, ph), tot in topk(al.retry_total, 10):
            ev = al.retry_events.get((tl, ph), 0)
            lines.append(f"  target={tl!r} phase={ph!r}: events={ev} total={fmt(tot)}")
    else:
        lines.append("retries_by_target_and_phase: (none)")

    if al.overshoot_events:
        lines.append("overshoot_by_target_and_phase (top):")
        for (tl, ph), tot in topk(al.overshoot_total, 10):
            ev = al.overshoot_events.get((tl, ph), 0)
            lines.append(f"  target={tl!r} phase={ph!r}: events={ev} total={fmt(tot)}")
    else:
        lines.append("overshoot_by_target_and_phase: (none)")

    if al.stuck_reports:
        lines.append(f"stuck_segments (showing up to {len(al.stuck_reports)}):")
        for r in al.stuck_reports:
            lines.append(f"  {r}")
    else:
        lines.append("stuck_segments: (none detected)")

    return "\n".join(lines)


# --- Main -------------------------------------------------------------------


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
    ap.add_argument(
        "--percentiles",
        action="store_true",
        help="Estimate and print p01/p50/p99 per numeric column (streaming approximation).",
    )
    ap.add_argument(
        "--alarms",
        action="store_true",
        help="Print LK/ISPP-specific alarm summaries (bounds degeneracy, retries/overshoot by phase, stuck segments).",
    )
    ap.add_argument(
        "--stuck-rows",
        type=int,
        default=20,
        help="Rows of consecutive non-improving wrd_read_level error (towards wrd_target_level) to flag a stuck segment (default 20).",
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
                "controller_vmin_ec",
                "controller_vmax_ec",
                "wrd_read_level",
                "wrd_target_level",
                "wrd_cycle_energy_fj",
                "wrd_retry_count",
            ]
            if args.columns:
                columns = [c.strip() for c in args.columns.split(",") if c.strip()]
            else:
                columns = [c for c in default_cols if c in fieldnames]
                if not columns:
                    # fallback: take first 12 non-obviously-string columns
                    columns = []
                    for c in fieldnames:
                        if c in (
                            "timestamp",
                            "waveform",
                            "material",
                            "wrd_phase_name",
                            "controller_state",
                        ):
                            continue
                        columns.append(c)
                        if len(columns) >= 12:
                            break

            stats: Dict[str, RunningStats] = {c: make_running_stats(args.percentiles) for c in columns}
            grouped: Dict[str, Dict[str, RunningStats]] = {}
            group_counts: Dict[str, int] = {}

            al = AlarmState(stuck_threshold_rows=args.stuck_rows) if args.alarms else None

            n_rows = 0
            n_kept = 0
            for row in reader:
                n_rows += 1

                if where is not None:
                    wk, wv = where
                    if row.get(wk, "") != wv:
                        continue

                n_kept += 1

                if al is not None:
                    al.push_row(row)

                gkey = None
                if args.group_by:
                    gkey = row.get(args.group_by, "")
                    if gkey not in grouped:
                        grouped[gkey] = {c: make_running_stats(args.percentiles) for c in columns}
                        group_counts[gkey] = 0
                    group_counts[gkey] += 1

                for c in columns:
                    x = get_float(row, c)
                    if x is None:
                        continue
                    stats[c].push(x)
                    if gkey is not None:
                        grouped[gkey][c].push(x)

            print(f"rows: {n_rows} (kept: {n_kept})")
            print(summarize(stats, columns, show_percentiles=args.percentiles))

            if al is not None:
                print("\n[alarms]")
                print(format_alarm_section(al))

            if args.group_by:
                keys = sorted(grouped.keys(), key=lambda k: group_counts.get(k, 0), reverse=True)
                if len(keys) > args.max_groups:
                    print(
                        f"\n(grouped by {args.group_by}: showing top {args.max_groups} of {len(keys)} groups by row count)"
                    )
                    keys = keys[: args.max_groups]
                else:
                    print(f"\n(grouped by {args.group_by}: {len(keys)} groups)")

                for k in keys:
                    print(f"\n--- {args.group_by}={k!r} (rows={group_counts.get(k, 0)}) ---")
                    print(summarize(grouped[k], columns, show_percentiles=args.percentiles))

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
