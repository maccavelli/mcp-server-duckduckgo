#!/usr/bin/env bash
# Go pre-add checks: format, lint, vet, tests, vulnerabilities.
#
# The single implementation of the pre-add rule in AGENTS.md, called from two
# places so they cannot drift: `make pre-add-check`, and the machine-wide agent
# gate that runs before every agent `git commit`
# (~/.agents/hooks/lib/precommit-checks.sh), which prefers this script whenever
# the repository ships one and Go files are staged.
#
# Adapted from go-selfupdate-lib's scripts/go-precheck.sh under dotfiles
# docs/decisions/0028-MADR-retire-per-file-golint-from-the-fleet-pre-add-rule.md
# and this repository's
# docs/decisions/0010-MADR-adopt-the-fleet-pre-add-wrapper.md. That source
# script carries the fleet's recorded behavior: gofmt's own failure fails the
# check; a staged deletion still names its package; a docs-only change is a
# clean no-op; the lint step cannot drift from CI.
#
# Two differences from the source script, decided in
# docs/decisions/0010-MADR-adopt-the-fleet-pre-add-wrapper.md:
#
# - The lint step runs once for the host GOOS, the same command, config and
#   GOOS `make lint` runs, so the two cannot drift. The source script lints
#   once per build target with cgo off; this repository's `make lint` is
#   host-only, so a per-GOOS pre-add run would pass a stricter check than the
#   commit-time gate it mirrors. Making `make lint` per-GOOS is a separate
#   decision; until then the two gates change together.
# - govulncheck runs through the same pinned invocation as the `vuln` target
#   (`go run golang.org/x/vuln/cmd/govulncheck@v1.10.0`), so no second
#   govulncheck version is installed or PATH-probed.
#
# Usage:
#   scripts/go-precheck.sh [file.go ...]
#
# With no arguments it checks every tracked Go file; with arguments, only those
# (non-Go arguments are ignored, so callers can pass a whole changed-file list).
#
# Exit codes: 0 all clear · 1 a check failed · 2 a required tool is missing,
# or the vulnerability database cannot be reached.
#
# Env:
#   GOLANGCI_LINT=<path>      golangci-lint binary (default: $(go env GOPATH)/bin)
#   GO_PRECHECK_SKIP_VULN=1   skip govulncheck (offline work)
set -uo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT" || exit 1

# Collect the Go files to check: files that exist go to gofmt; every Go
# argument, deleted or not, names a package directory.
files=()
dirs=()
if [ "$#" -gt 0 ]; then
  for f in "$@"; do
    case "$f" in
    *.go)
      [ -f "$f" ] && files+=("$f")
      dirs+=("$(dirname "$f")")
      ;;
    esac
  done
else
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    # A tracked file deleted but not yet staged is not a file to format; its
    # package is still vetted and tested (the source script's 0010-MADR D3).
    [ -f "$f" ] && files+=("$f")
    dirs+=("$(dirname "$f")")
  done < <(git ls-files '*.go')
fi

# Nothing to check is not an error: a docs-only change. Return before any tool
# runs — gofmt with no file reads stdin.
if [ "${#dirs[@]}" -eq 0 ]; then
  echo "go-precheck: no Go files to check."
  exit 0
fi
ran=()

need() {
  command -v "$1" >/dev/null 2>&1 && return 0
  echo "go-precheck: $1 not found in PATH." >&2
  return 1
}

failed=0
fail() {
  # 2 (tool missing) outranks 1 (check failed).
  [ "$failed" -lt "$1" ] && failed="$1"
}

# 1. gofmt, over the files that exist. Its exit status counts as well as its
# list: a file it cannot read or parse makes it fail with nothing on its
# output.
if [ "${#files[@]}" -gt 0 ]; then
  need gofmt || exit 2
  gofmt_err="$(mktemp)"
  unformatted="$(gofmt -l "${files[@]}" 2>"$gofmt_err")"
  gofmt_rc=$?
  if [ "$gofmt_rc" -ne 0 ]; then
    echo "gofmt: failed (exit $gofmt_rc):" >&2
    sed 's/^/  /' "$gofmt_err" >&2
    fail 1
  fi
  rm -f "$gofmt_err"
  if [ -n "$unformatted" ]; then
    echo "gofmt: these files are not formatted (run 'gofmt -w <file>'):" >&2
    printf '%s\n' "$unformatted" | sed 's/^/  /' >&2
    fail 1
  fi
  ran+=(gofmt)
fi

# 2. golangci-lint, with this repository's configuration: the same command,
# config file and GOOS `make lint` runs, so a commit cannot pass a different
# check than the tree's own lint gate (see the header note on why this is one
# host-GOOS run and not the source script's per-target loop).
need go || exit 2
GOLANGCI="${GOLANGCI_LINT:-$(go env GOPATH)/bin/golangci-lint}"
if [ -x "$GOLANGCI" ]; then
  if ! lint_out="$("$GOLANGCI" run -c .golangci.yml ./... 2>&1)"; then
    echo "golangci-lint:" >&2
    printf '%s\n' "$lint_out" | tail -40 | sed 's/^/  /' >&2
    fail 1
  fi
  ran+=(golangci-lint)
else
  echo "go-precheck: golangci-lint not found at $GOLANGCI." >&2
  echo "  install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0" >&2
  fail 2
fi

# 3. go vet and go test, over the packages the files belong to (./... with no
# arguments), each in the module that owns it. A package directory with no Go
# file left, after a deletion, widens its module's run to ./... .

# module_of DIR: the directory, relative to the root, of the go.mod that owns
# DIR; "." for the main module.
module_of() {
  local d="$1"
  while [ "$d" != "." ] && [ ! -f "$d/go.mod" ]; do
    d="$(dirname "$d")"
  done
  printf '%s\n' "$d"
}

# entries: one "MODULE|PACKAGE" line per package to check, PACKAGE relative
# to MODULE.
entries=()
while IFS= read -r d; do
  [ -z "$d" ] && continue
  m="$(module_of "$d")"
  if [ "$#" -eq 0 ] || ! compgen -G "$d/*.go" >/dev/null; then
    entries+=("$m|./...")
  elif [ "$d" = "$m" ]; then
    entries+=("$m|.")
  elif [ "$m" = "." ]; then
    entries+=(".|./$d")
  else
    entries+=("$m|./${d#"$m"/}")
  fi
done < <(printf '%s\n' "${dirs[@]}" | sort -u)

vet_out=""
test_out=""
while IFS= read -r m; do
  pkgs=()
  while IFS= read -r p; do
    pkgs+=("$p")
  done < <(printf '%s\n' "${entries[@]}" | awk -F'|' -v m="$m" '$1 == m { print $2 }' | sort -u)
  gocmd=(go)
  if [ "$m" != "." ]; then
    gocmd=(env GOWORK=off go -C "$m")
  fi
  if ! out="$("${gocmd[@]}" vet "${pkgs[@]}" 2>&1)"; then
    vet_out+="$out"$'\n'
  fi
  if ! out="$("${gocmd[@]}" test "${pkgs[@]}" 2>&1)"; then
    test_out+="$out"$'\n'
  fi
done < <(printf '%s\n' "${entries[@]}" | cut -d'|' -f1 | sort -u)
if [ -n "$vet_out" ]; then
  echo "go vet:" >&2
  printf '%s' "$vet_out" | sed 's/^/  /' >&2
  fail 1
fi
if [ -n "$test_out" ]; then
  echo "go test:" >&2
  printf '%s' "$test_out" | tail -40 | sed 's/^/  /' >&2
  fail 1
fi
ran+=("go vet" "go test")

# 4. govulncheck, over the module, through the same pinned invocation as the
# `vuln` target. It reports *called* vulnerabilities, so it is a property of
# the whole build rather than of the edited files.
#
# Exit status 3 is govulncheck's "vulnerabilities found" and always fails. Only
# another non-zero status whose output looks like a network failure is treated
# as an unreachable database: matching those words on a status-3 run would let a
# finding whose trace names a `proxy` package or a `Timeout` function through.
if [ "${GO_PRECHECK_SKIP_VULN:-0}" = "1" ]; then
  echo "govulncheck: skipped (GO_PRECHECK_SKIP_VULN=1)" >&2
else
  vuln_out="$(go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./... 2>&1)"
  vuln_rc=$?
  if [ "$vuln_rc" -eq 3 ]; then
    echo "govulncheck: vulnerabilities found:" >&2
    printf '%s\n' "$vuln_out" | tail -30 | sed 's/^/  /' >&2
    fail 1
  elif [ "$vuln_rc" -ne 0 ]; then
    if printf '%s' "$vuln_out" | grep -qiE 'no such host|connection refused|timeout|dial tcp|proxy'; then
      echo "govulncheck: could not reach the vulnerability database;" \
        "set GO_PRECHECK_SKIP_VULN=1 to skip it offline." >&2
      fail 2
    else
      echo "govulncheck: failed (exit $vuln_rc):" >&2
      printf '%s\n' "$vuln_out" | tail -30 | sed 's/^/  /' >&2
      fail 1
    fi
  else
    ran+=(govulncheck)
  fi
fi

if [ "$failed" -eq 0 ]; then
  summary="$(printf '%s, ' "${ran[@]}")"
  echo "go-precheck: ${#dirs[@]} file(s) clean (${summary%, })."
fi
exit "$failed"
