#!/usr/bin/env python3
"""Record and docs-link tooling for the fixed docs tree.

Modes:
  --next         print the next record number: one more than the highest in
                 use anywhere in the repository, so a gap stays a gap
  --check        validate numbered records: filename shape, number pairing,
                 placement, and relative markdown links with their anchors
  --check-all    --check plus link checks for unnumbered docs
                 (docs/guides/**, docs/README.md, docs/architecture.md,
                 and the root README.md), and no unfilled template slot
                 ({{NAME}}) in those files or the root CLAUDE.md

This copy does not generate docs/README.md. The index is written by hand
so Status and extra sections survive.

Scope: only files git would track are scanned (tracked, or untracked and not
git-ignored), so a local working copy under .gitignore cannot collide with a
real number or move --next. A directory whose name starts with `_` is never
entered, tracked or not: `_name/` marks a tree that is deprecated or moved
elsewhere and kept as a fallback. Where this directory is not itself the top of
a git work tree, the filesystem is walked instead, and only the `_` rule and
SKIP_PARTS apply.

Errors print to stdout; exit status is 1 iff errors were found. Errors: a
number claimed by more than one MADR, a PLAN without a same-number MADR, a
GATES without a same-number PLAN, a record outside its kind directory, a
misnamed record (a markdown file in a docs/decisions or docs/reports
directory that is not NNNN-KIND-slug.md, or a file anywhere named like a
record in another case or width), a MADR or PLAN that does not open with
YAML frontmatter holding `status` and `date`, a MADR or PLAN status outside
its kind's vocabulary (or a `superseded by` naming no existing record of the
same kind), a relative link to a missing file, and a link whose #fragment
names no heading or HTML anchor of its markdown target. With --check-all, also
a {{NAME}} slot (uppercase name, outside fences and inline code) left
unfilled in a generated document (dotfiles 0007-MADR). Several PLANs and
REPORTs may share a number.

Status vocabulary (dotfiles 0005-MADR): a MADR uses MADR v4's list; a PLAN
uses MADR's words plus the Kubernetes KEP lifecycle. Either may be
`superseded by NNNN-<KIND>-<slug>.md`.

Anchors are computed as GitHub computes them from ATX headings (# ...):
inline markup removed, lowercased, characters other than letters, digits,
underscores, spaces and hyphens removed, spaces turned into hyphens, and a
repeated heading suffixed -1, -2, ... Headings in fenced code are ignored.
"""

from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
RECORD_RE = re.compile(r"^(\d{4})-(MADR|PLAN|REPORT|GATES)-(.+)\.md$")
LOOKALIKE_RE = re.compile(r"^\d+-(madr|plan|report|gates)-", re.IGNORECASE)
LINK_RE = re.compile(r"\]\(([^)]+)\)")
SCHEME_RE = re.compile(r"^[a-zA-Z][a-zA-Z0-9+.\-]*:")
FENCE_RE = re.compile(r"^\s*(```|~~~)")
PLACEHOLDER_RE = re.compile(r"\{\{[A-Z][A-Z0-9_]*\}\}")
HEADING_RE = re.compile(r"^ {0,3}#{1,6}(?:[ \t]+(.*?))?(?:[ \t]+#+)?[ \t]*$")
HTML_ANCHOR_RE = re.compile(r"<a\s+(?:[^>]*\s)?(?:id|name)\s*=\s*[\"']([^\"']+)[\"']", re.IGNORECASE)
FRONTMATTER_KEY_RE = re.compile(r"^([A-Za-z][\w-]*)\s*:")
FRONTMATTER_KINDS = ("MADR", "PLAN")
FRONTMATTER_KEYS = ("status", "date")
STATUSES = {
    "MADR": ("proposed", "accepted", "rejected", "deprecated"),
    "PLAN": ("proposed", "accepted", "in-progress", "deferred", "implemented", "rejected", "withdrawn"),
}
SUPERSEDED_RE = re.compile(r"^superseded by (\d{4}-(?:MADR|PLAN)-.+\.md)$")
KIND_SUFFIX = {"MADR": "docs/decisions", "PLAN": "docs/decisions",
               "REPORT": "docs/reports", "GATES": "docs/reports"}
RECORD_DIRS = ("docs/decisions", "docs/reports")
SKIP_PARTS = {".git", "node_modules", "vendor", "dist", ".pub-cache",
              ".pub", ".symlinks", "build", ".dart_tool", ".gradle",
              ".idea", ".vscode"}


def errors(msg: str) -> None:
    print(msg)


def git_visible_markdown() -> set[Path] | None:
    """Markdown files git tracks or would let `git add` stage: tracked, plus untracked and not ignored.

    None when REPO is not itself the top of a git work tree, or git cannot answer, so a tarball or a
    directory inside some other repository is walked instead of trusting a verdict about a different tree."""
    try:
        top = subprocess.run(["git", "rev-parse", "--show-toplevel"], cwd=REPO,
                             capture_output=True, text=True, check=True).stdout.strip()
        if Path(top).resolve() != REPO:
            return None
        listing = subprocess.run(["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard",
                                  "--", "*.md"], cwd=REPO, capture_output=True, check=True).stdout
    except (OSError, subprocess.CalledProcessError):
        return None
    return {REPO / p.decode() for p in listing.split(b"\0") if p}


def markdown_files() -> list[Path]:
    visible = git_visible_markdown()
    candidates = REPO.rglob("*.md") if visible is None else visible
    out = []
    for p in sorted(candidates):
        dirs = p.relative_to(REPO).parts[:-1]
        if any(part in SKIP_PARTS or part.startswith("_") for part in dirs):
            continue
        if p.is_file():
            out.append(p)
    return out


class Record:
    def __init__(self, path: Path) -> None:
        m = RECORD_RE.match(path.name)
        if not m:
            raise ValueError(path.name)
        self.path = path
        self.rel = path.relative_to(REPO).as_posix()
        self.num, self.kind, self.slug = m.group(1), m.group(2), m.group(3)


def find_records() -> list[Record]:
    records = []
    for p in markdown_files():
        if RECORD_RE.match(p.name):
            records.append(Record(p))
    return records


def next_number() -> int:
    return max((int(r.num) for r in find_records()), default=0) + 1


def in_record_dir(path: Path) -> bool:
    parent = path.parent.relative_to(REPO).as_posix()
    return any(parent == d or parent.endswith("/" + d) for d in RECORD_DIRS)


def check_names() -> int:
    found = 0
    for p in markdown_files():
        if RECORD_RE.match(p.name):
            continue
        if in_record_dir(p) or LOOKALIKE_RE.match(p.name):
            errors(f"misnamed record: {p.relative_to(REPO).as_posix()} "
                   "(expected NNNN-MADR|PLAN|REPORT|GATES-slug.md)")
            found += 1
    return found


def check_structure(records: list[Record]) -> int:
    found = 0
    by_num: dict[str, list[Record]] = {}
    for r in records:
        by_num.setdefault(r.num, []).append(r)
    for num in sorted(by_num):
        items = by_num[num]
        madrs = [r for r in items if r.kind == "MADR"]
        plans = [r for r in items if r.kind == "PLAN"]
        gates = [r for r in items if r.kind == "GATES"]
        if len(madrs) > 1:
            errors(f"number {num} is claimed by {len(madrs)} MADRs: "
                   + ", ".join(r.rel for r in madrs)
                   + " (a number is never reused)")
            found += 1
        if plans and not madrs:
            errors(f"number {num} has a PLAN but no MADR: "
                   + ", ".join(r.rel for r in plans))
            found += 1
        if gates and not plans:
            errors(f"number {num} has a GATES but no PLAN: "
                   + ", ".join(r.rel for r in gates))
            found += 1
    for r in records:
        expected = KIND_SUFFIX[r.kind]
        parent = r.path.parent.relative_to(REPO).as_posix()
        if parent != expected and not parent.endswith("/" + expected):
            errors(f"record outside its directory: {r.rel} "
                   f"(expected a path ending in {expected}/)")
            found += 1
    return found


def frontmatter_problem(lines: list[str]) -> str | None:
    """What is wrong with a record's opening frontmatter, or None. Only the
    block between the first line's `---` and the next `---` counts, so a
    `status:` line in the body does not satisfy it."""
    if not lines or lines[0].rstrip() != "---":
        return "does not open with a '---' frontmatter line"
    end = next((i for i in range(1, len(lines)) if lines[i].rstrip() == "---"), None)
    if end is None:
        return "frontmatter is not closed by a '---' line"
    keys = {m.group(1) for line in lines[1:end] if (m := FRONTMATTER_KEY_RE.match(line))}
    missing = [k for k in FRONTMATTER_KEYS if k not in keys]
    return f"frontmatter lacks {', '.join(missing)}" if missing else None


def check_frontmatter(records: list[Record]) -> int:
    found = 0
    for r in records:
        if r.kind not in FRONTMATTER_KINDS:
            continue
        problem = frontmatter_problem(r.path.read_text(encoding="utf-8").splitlines())
        if problem:
            errors(f"frontmatter: {r.rel}: {problem}")
            found += 1
    return found


def frontmatter_status(lines: list[str]) -> str | None:
    """The frontmatter `status` value with quotes stripped, or None when the
    frontmatter is malformed (check_frontmatter reports that)."""
    if frontmatter_problem(lines):
        return None
    end = next(i for i in range(1, len(lines)) if lines[i].rstrip() == "---")
    for line in lines[1:end]:
        m = FRONTMATTER_KEY_RE.match(line)
        if m and m.group(1) == "status":
            return line.split(":", 1)[1].strip().strip("\"'")
    return None


def check_statuses(records: list[Record]) -> int:
    found = 0
    by_name = {r.path.name: r for r in records}
    for r in records:
        if r.kind not in STATUSES:
            continue
        value = frontmatter_status(r.path.read_text(encoding="utf-8").splitlines())
        if value is None or value in STATUSES[r.kind]:
            continue
        m = SUPERSEDED_RE.match(value)
        if m:
            target = by_name.get(m.group(1))
            if target is not None and target.kind == r.kind:
                continue
            errors(f"status: {r.rel}: '{value}' does not name an existing {r.kind}")
        else:
            allowed = " | ".join(STATUSES[r.kind])
            errors(f"status: {r.rel}: '{value}' is not a {r.kind} status "
                   f"({allowed} | superseded by NNNN-{r.kind}-<slug>.md)")
        found += 1
    return found


TICKS_RE = re.compile(r"`+")


def code_spans(line: str) -> list[tuple[int, int, int]]:
    """(start, end, width) of each inline code span. As in CommonMark, a span
    opens on a run of N backticks and closes on the next run of exactly N, so
    ``a `b` c`` is one span. A run with no closer is literal text."""
    runs = list(TICKS_RE.finditer(line))
    spans, i = [], 0
    while i < len(runs):
        width = len(runs[i].group())
        close = next((j for j in range(i + 1, len(runs))
                      if len(runs[j].group()) == width), None)
        if close is None:
            i += 1
            continue
        spans.append((runs[i].start(), runs[close].end(), width))
        i = close + 1
    return spans


def strip_code_spans(line: str) -> str:
    """Blank out `inline code` spans so example links inside them, like
    `[text](url)` or `SendRequest[T any](c *Connection, …)`, are not
    mistaken for live markdown links. Links after an unmatched backtick
    stay live."""
    out, pos = [], 0
    for start, end, _ in code_spans(line):
        out.append(line[pos:start])
        pos = end
    out.append(line[pos:])
    return "".join(out)


def strip_inline_markup(text: str) -> str:
    text = re.sub(r"<[^>]+>", "", text)
    text = text.replace("*", "")
    return re.sub(r"(?<!\w)_+|_+(?!\w)", "", text)


def heading_slug(text: str) -> str:
    """Emphasis and HTML are stripped outside code spans only: inside a span,
    `_bulk` keeps its underscore, as GitHub and GitLab slug it."""
    text = re.sub(r"!?\[([^\]]*)\]\([^)]*\)", r"\1", text)
    pieces, pos = [], 0
    for start, end, width in code_spans(text):
        pieces.append(strip_inline_markup(text[pos:start]))
        content = text[start + width:end - width]
        if len(content) > 1 and content.startswith(" ") and content.endswith(" "):
            content = content[1:-1]
        pieces.append(content)
        pos = end
    pieces.append(strip_inline_markup(text[pos:]))
    text = re.sub(r"[^\w\- ]", "", "".join(pieces).strip().lower())
    return text.replace(" ", "-")


_anchor_cache: dict[Path, set[str]] = {}


def anchors(path: Path) -> set[str]:
    if path in _anchor_cache:
        return _anchor_cache[path]
    out: set[str] = set()
    seen: dict[str, int] = {}
    in_fence = False
    for line in path.read_text(encoding="utf-8").splitlines():
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        out.update(HTML_ANCHOR_RE.findall(line))
        m = HEADING_RE.match(line)
        if m:
            slug = heading_slug(m.group(1) or "")
            n = seen.get(slug, 0)
            seen[slug] = n + 1
            out.add(slug if n == 0 else f"{slug}-{n}")
    _anchor_cache[path] = out
    return out


def check_links(path: Path) -> int:
    rel = path.relative_to(REPO).as_posix()
    found = 0
    in_fence = False
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError as exc:
        errors(f"unreadable file: {rel}: {exc}")
        return 1
    for i, line in enumerate(lines, 1):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        for m in LINK_RE.finditer(strip_code_spans(line)):
            target = m.group(1).strip()
            if not target or SCHEME_RE.match(target):
                continue
            if any(c.isspace() for c in target):
                continue
            file_part, _, fragment = target.partition("#")
            if file_part:
                resolved = Path(os.path.normpath(path.parent / file_part))
                if not resolved.exists():
                    errors(f"broken relative link: {rel}:{i}: {target}")
                    found += 1
                    continue
            else:
                resolved = path
            if fragment and resolved.suffix == ".md" and resolved.is_file():
                if fragment not in anchors(resolved):
                    errors(f"broken anchor: {rel}:{i}: {target}")
                    found += 1
    return found


def unnumbered_docs() -> list[Path]:
    out = []
    for p in markdown_files():
        rel = p.relative_to(REPO).as_posix()
        if rel == "README.md" or rel in ("docs/README.md", "docs/architecture.md"):
            out.append(p)
        elif rel.startswith("docs/guides/"):
            out.append(p)
    return out


def placeholder_docs() -> list[Path]:
    """The documents generated from template slots: the unnumbered docs plus the root CLAUDE.md."""
    out = unnumbered_docs()
    claude = REPO / "CLAUDE.md"
    if claude.is_file() and claude not in out:
        out.append(claude)
    return out


def check_placeholders(path: Path) -> int:
    rel = path.relative_to(REPO).as_posix()
    found = 0
    in_fence = False
    for i, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        for m in PLACEHOLDER_RE.finditer(strip_code_spans(line)):
            errors(f"unfilled placeholder: {rel}:{i}: {m.group(0)}")
            found += 1
    return found


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--next", action="store_true", help="print the next record number")
    ap.add_argument("--check", action="store_true", help="validate numbered records")
    ap.add_argument("--check-all", action="store_true", help="--check plus unnumbered docs")
    args = ap.parse_args()
    if not (args.next or args.check or args.check_all):
        ap.print_help()
        return 2
    records = find_records()
    if args.next:
        print(f"{next_number():04d}")
    if args.check or args.check_all:
        found = (check_names() + check_structure(records) + check_frontmatter(records)
                 + check_statuses(records))
        found += sum(check_links(r.path) for r in records)
        if args.check_all:
            for p in unnumbered_docs():
                found += check_links(p)
            for p in placeholder_docs():
                found += check_placeholders(p)
        if found:
            print(f"{found} error(s)")
            return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
