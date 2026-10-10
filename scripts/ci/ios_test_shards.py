#!/usr/bin/env python3
import argparse
import json
import os
import re
import statistics
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
UI_TARGET = "PocketAideTests"
UNIT_TARGET = "PocketAideUnitTests"
DURATIONS = Path(__file__).with_name("ios-test-durations.json")
CLASS_RE = re.compile(r"^(?:final )?class (\w+)\s*:\s*XCTestCase\b", re.M)


def ui_classes():
    names = set()
    for f in (ROOT / "ios" / UI_TARGET).glob("*.swift"):
        names |= set(CLASS_RE.findall(f.read_text(encoding="utf-8")))
    return sorted(names)


def plan(total):
    data = json.loads(DURATIONS.read_text(encoding="utf-8"))
    known = data[UI_TARGET]
    # A class missing from the durations file is a new one: weigh it like a
    # typical class so it still lands somewhere instead of being dropped.
    default = statistics.median(known.values())
    units = [(f"{UI_TARGET}/{c}", known.get(c, default)) for c in ui_classes()]
    units.append((UNIT_TARGET, data[UNIT_TARGET]))
    shards = [[] for _ in range(total)]
    loads = [0.0] * total
    for ident, weight in sorted(units, key=lambda u: (-u[1], u[0])):
        i = loads.index(min(loads))
        shards[i].append(ident)
        loads[i] += weight
    return shards, loads


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--index", type=int, required=True)
    p.add_argument("--total", type=int, required=True)
    a = p.parse_args()
    if not 0 <= a.index < a.total:
        sys.exit(f"index {a.index} out of range for total {a.total}")
    shards, loads = plan(a.total)
    for i, (s, load) in enumerate(zip(shards, loads)):
        print(f"shard {i}: ~{load / 60:.0f} min, {len(s)} units{'  <- this job' if i == a.index else ''}")
    mine = sorted(shards[a.index])
    for ident in mine:
        print(f"  {ident}")
    args = " ".join(f"-only-testing:{ident}" for ident in mine)
    pr_monitor = f"{UI_TARGET}/PRMonitorUITests" in mine
    out = os.environ.get("GITHUB_OUTPUT")
    if out:
        with open(out, "a", encoding="utf-8") as fh:
            fh.write(f"only_testing={args}\n")
            fh.write(f"pr_monitor={'true' if pr_monitor else 'false'}\n")


if __name__ == "__main__":
    main()
