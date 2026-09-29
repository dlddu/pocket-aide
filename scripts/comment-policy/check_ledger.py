#!/usr/bin/env python3
import hashlib
import os
import re
import subprocess
import sys
from collections import defaultdict

ROOT = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip()
LEDGER = os.path.join(ROOT, "docs/comment-policy/ledger.md")
SCAN = os.path.join(ROOT, "scripts/comment-policy/scan.sh")
SURFACES = {"L": "## L ", "D": "## D ", "E": "## E "}
READ_ME = "## 읽는 법"
HEADER = ["판정일", "범위", "주석 줄", "지문", "판정", "결과"]
VERDICTS = {"완료", "—"}


def measure():
    out = subprocess.run(["bash", SCAN], cwd=ROOT, capture_output=True, text=True, check=True).stdout
    lines = [l for l in out.split("\n") if l]
    per = {s: defaultdict(list) for s in SURFACES}
    extra = []
    for l in lines:
        if l.startswith(("D:", "E:")):
            per[l[0]][l[2:].split(":", 1)[0]].append(l)
        elif l.startswith(("X:", "unclassified:")):
            extra.append(l)
        else:
            per["L"][l.split(":", 1)[0]].append(l)
    digest = hashlib.sha256(("\n".join(sorted(lines)) + "\n").encode()).hexdigest()
    return per, extra, digest


def fingerprint(entries):
    return hashlib.sha256(("\n".join(sorted(entries)) + "\n").encode()).hexdigest()[:12]


def cells(line):
    return [c.strip() for c in line.strip().strip("|").split(" | ")]


def parse(text, errors):
    tables = {s: [] for s in SURFACES}
    section = None
    for n, line in enumerate(text.split("\n"), 1):
        if line.startswith("## "):
            section = next((s for s, h in SURFACES.items() if line.startswith(h)), None)
            if line.strip() == READ_ME:
                section = "READ"
            elif section is None:
                errors.append(f"ledger.md:{n}: 허용되지 않은 절 {line.strip()!r}")
            continue
        if line.startswith("# ") or not line.strip() or section == "READ":
            continue
        if section in SURFACES and line.startswith("|"):
            row = cells(line)
            if row == HEADER or all(re.fullmatch(r":?-+:?", c) for c in row):
                continue
            if len(row) != len(HEADER):
                errors.append(f"ledger.md:{n}: 열이 {len(row)}개 (6개여야 한다)")
            else:
                tables[section].append((n, row))
            continue
        errors.append(f"ledger.md:{n}: 표와 「읽는 법」 밖의 산문")
    return tables


def check_table(surface, rows, measured, errors):
    seen = {}
    firsts = []
    done = pending = 0
    for n, (date, scope, count, fp, verdict, result) in rows:
        files = re.findall(r"`([^`]+)`", scope)
        if not files:
            errors.append(f"ledger.md:{n}: 범위 칸에 파일이 없다")
            continue
        firsts.append(files[0])
        entries = []
        for f in files:
            if f in seen:
                errors.append(f"ledger.md:{n}: {f} 가 {seen[f]}행에도 있다")
            seen[f] = n
            if f not in measured:
                errors.append(f"ledger.md:{n}: {f} 에 {surface} 주석이 없다(행에서 뺄 것)")
            entries += measured.get(f, [])
        if count != str(len(entries)):
            errors.append(f"ledger.md:{n}: 주석 줄 {count} ≠ 실측 {len(entries)}")
        if fp.strip("`") != fingerprint(entries):
            errors.append(f"ledger.md:{n}: 지문 {fp} ≠ 실측 {fingerprint(entries)}")
        if verdict not in VERDICTS:
            errors.append(f"ledger.md:{n}: 판정 칸 {verdict!r} — `완료`·`—` 만 유효")
        elif verdict == "완료":
            links = re.findall(r"\]\((passes/[^)]+)\)", result)
            if not links:
                errors.append(f"ledger.md:{n}: `완료` 행의 결과 칸에 passes/ 링크가 없다")
            for link in links:
                if not os.path.exists(os.path.join(ROOT, "docs/comment-policy", link)):
                    errors.append(f"ledger.md:{n}: 링크 {link} 가 없다")
            done += len(entries)
        else:
            pending += len(entries)
    if firsts != sorted(firsts):
        errors.append(f"{surface} 표: 행이 첫 파일 경로의 사전순이 아니다")
    for f in sorted(set(measured) - set(seen)):
        errors.append(f"{surface} 표: {f} ({len(measured[f])}줄) 가 어느 행에도 없다")
    total = sum(len(v) for v in measured.values())
    return f"{surface}: 행 {len(rows)} · 파일 {len(measured)} · 주석 {total}줄 · 완료 {done} · 미판정 {pending}"


def main():
    errors = []
    per, extra, digest = measure()
    errors += [f"실측: {l} — 미분류·파싱 실패는 원장 밖에서 먼저 해소할 것" for l in extra]
    tables = parse(open(LEDGER, encoding="utf-8").read(), errors)
    for surface, rows in tables.items():
        print(check_table(surface, rows, per[surface], errors))
    print(f"지문: {digest}")
    if "--rows" in sys.argv:
        for surface, rows in tables.items():
            for n, row in rows:
                entries = [e for f in re.findall(r"`([^`]+)`", row[1]) for e in per[surface].get(f, [])]
                print(f"{surface} {n} {len(entries)} {fingerprint(entries)}")
    if errors:
        print("\n".join(errors), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
