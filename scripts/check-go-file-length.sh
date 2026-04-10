#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LIMIT="${1:-200}"

violations=0

while IFS= read -r -d '' file; do
  lines=$(wc -l < "$file" | tr -d ' ')
  if (( lines > LIMIT )); then
    printf 'file exceeds %s lines: %s (%s)\n' "$LIMIT" "${file#$ROOT_DIR/}" "$lines"
    violations=1
  fi
done < <(
  find \
    "$ROOT_DIR/internal/application/product" \
    "$ROOT_DIR/internal/application/sale" \
    "$ROOT_DIR/internal/application/transfer" \
    "$ROOT_DIR/internal/adapters/repository/postgres/product" \
    "$ROOT_DIR/internal/adapters/repository/postgres/sale" \
    "$ROOT_DIR/internal/adapters/repository/postgres/transfer" \
    -type f -name '*.go' ! -name '*_test.go' -print0
)

if (( violations != 0 )); then
  exit 1
fi

printf 'all checked Go files are within %s lines\n' "$LIMIT"
