#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION:-v2.12.2}"
GOSEC_VERSION="${GOSEC_VERSION:-v2.22.3}"
GOLANGCI_LINT_ARGS=(--enable-only=govet,ineffassign --timeout=5m)
GOSEC_ARGS=(-exclude-generated -exclude-dir=.git ./...)

usage() {
  cat <<'USAGE'
Usage: scripts/quality.sh <command>

Commands:
  fmt         Format Go files in place with gofmt
  fmt-check   Fail if any Go files are not gofmt-formatted
  vet         Run go vet
  lint        Run golangci-lint with the repository's baseline linter set
  test        Run fast tests
  race-test   Run race-enabled tests with coverage
  sast        Run gosec
  build       Build CLI binaries
  pre-commit  Run checks suitable for local commits: fmt-check, vet, lint, test
  ci          Run CI checks: fmt-check, vet, lint, race-test, sast, build
  all         Alias for ci
USAGE
}

run_golangci_lint() {
  if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run "${GOLANGCI_LINT_ARGS[@]}"
  else
    go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}" run "${GOLANGCI_LINT_ARGS[@]}"
  fi
}

run_gosec() {
  if command -v gosec >/dev/null 2>&1; then
    gosec "${GOSEC_ARGS[@]}"
  else
    go run "github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}" "${GOSEC_ARGS[@]}"
  fi
}

fmt_go() {
  local files
  files="$(gofmt -l .)"
  if [[ -n "$files" ]]; then
    # shellcheck disable=SC2086 # gofmt -l emits newline-separated repository paths without spaces in this project.
    gofmt -w $files
  fi
}

fmt_check() {
  local files
  files="$(gofmt -l .)"
  if [[ -n "$files" ]]; then
    echo "The following files are not gofmt-formatted:" >&2
    echo "$files" >&2
    echo "Run: scripts/quality.sh fmt" >&2
    return 1
  fi
}

vet() {
  go vet ./...
}

lint() {
  run_golangci_lint
}

test_fast() {
  go test ./...
}

race_test() {
  go test -race -covermode=atomic -coverprofile=coverage.out ./...
}

sast() {
  run_gosec
}

build() {
  go build ./cmd/s2req ./cmd/s2req-schema
}

pre_commit() {
  fmt_check
  vet
  lint
  test_fast
}

ci() {
  fmt_check
  vet
  lint
  race_test
  sast
  build
}

cmd="${1:-}"
case "$cmd" in
  fmt) fmt_go ;;
  fmt-check) fmt_check ;;
  vet) vet ;;
  lint) lint ;;
  test) test_fast ;;
  race-test) race_test ;;
  sast) sast ;;
  build) build ;;
  pre-commit) pre_commit ;;
  ci|all) ci ;;
  -h|--help|help) usage ;;
  *)
    usage >&2
    exit 2
    ;;
esac
