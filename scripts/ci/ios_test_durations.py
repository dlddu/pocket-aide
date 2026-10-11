#!/usr/bin/env python3
import argparse
import json
import re
import statistics
import sys
from pathlib import Path

DURATIONS = Path(__file__).with_name("ios-test-durations.json")
UI_TARGET = "PocketAideTests"
UNIT_TARGET = "PocketAideUnitTests"
ANSI = re.compile(r"\x1b\[[0-9;]*m")
SUITE = re.compile(r"Test Suite '([^']+)' started at")
EXECUTED = re.compile(r"Executed \d+ tests?, with \d+ failures? .* in [\d.]+ \(([\d.]+)\) seconds")
DRIFT = 0.25


def extract(lines):
    measured = {}
    bundle = suite = None
    for raw in lines:
        line = ANSI.sub("", raw)
        started = SUITE.search(line)
        if started:
            name = started.group(1)
            if name.endswith(".xctest"):
                bundle = name.removesuffix(".xctest")
            elif name not in ("All tests", "Selected tests"):
                suite = name
            continue
        executed = EXECUTED.search(line)
        if not executed or suite is None:
            continue
        unit = UNIT_TARGET if bundle == UNIT_TARGET else suite
        measured[unit] = measured.get(unit, 0.0) + float(executed.group(1))
        suite = None
    return measured


def cmd_extract(a):
    with open(a.log, encoding="utf-8", errors="replace") as fh:
        measured = extract(fh)
    if not measured:
        sys.exit(f"no test suite timings found in {a.log}")
    total = sum(measured.values())
    Path(a.out).write_text(json.dumps({k: round(v) for k, v in sorted(measured.items())}, indent=2) + "\n", encoding="utf-8")
    print(f"{len(measured)} units, {total / 60:.1f} min measured")
    if a.planned:
        drift = (total - a.planned) / a.planned
        print(f"planned {a.planned / 60:.1f} min, drift {drift:+.0%}")
        if abs(drift) > DRIFT:
            print(
                f"::warning title=ios-test durations::This shard ran {total / 60:.1f} min against a planned "
                f"{a.planned / 60:.1f} min ({drift:+.0%}). Refresh {DURATIONS.name}: download the "
                f"ios-test-durations-* artifacts of a green run and pass them to `{Path(__file__).name} merge`."
            )


def cmd_merge(a):
    samples = {}
    for name in a.files:
        for unit, seconds in json.loads(Path(name).read_text(encoding="utf-8")).items():
            samples.setdefault(unit, []).append(seconds)
    if not samples:
        sys.exit("no measurements to merge")
    current = json.loads(DURATIONS.read_text(encoding="utf-8"))
    ui = dict(current[UI_TARGET])
    sys.dont_write_bytecode = True
    sys.path.insert(0, str(Path(__file__).parent))
    from ios_test_shards import ui_classes

    live = set(ui_classes())
    for unit, values in samples.items():
        value = max(1, round(statistics.median(values)))
        if unit == UNIT_TARGET:
            current[UNIT_TARGET] = value
        elif unit in live:
            ui[unit] = value
    for gone in sorted(set(ui) - live):
        print(f"dropped {gone}: no such UI test class")
        del ui[gone]
    for missing in sorted(live - set(ui)):
        print(f"::warning title=ios-test durations::{missing} has no measurement yet")
    current[UI_TARGET] = dict(sorted(ui.items()))
    DURATIONS.write_text(json.dumps(current, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(f"updated {DURATIONS} from {len(a.files)} files, {len(samples)} units")


def main():
    p = argparse.ArgumentParser(description="Measure iOS test class durations and refresh the shard planner's table.")
    sub = p.add_subparsers(dest="cmd", required=True)
    e = sub.add_parser("extract", help="Read an xcodebuild/xcbeautify log and write {class: seconds} for one shard.")
    e.add_argument("--log", required=True)
    e.add_argument("--out", required=True)
    e.add_argument("--planned", type=float, default=0, help="Seconds the shard planner expected for this shard.")
    e.set_defaults(func=cmd_extract)
    m = sub.add_parser("merge", help=f"Fold extract outputs (median per class) into {DURATIONS.name}.")
    m.add_argument("files", nargs="+")
    m.set_defaults(func=cmd_merge)
    a = p.parse_args()
    a.func(a)


if __name__ == "__main__":
    main()
