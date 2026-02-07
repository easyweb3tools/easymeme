#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v swag >/dev/null 2>&1; then
  echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest" >&2
  exit 1
fi

cd "$ROOT_DIR/server"
swag init -g cmd/server/main.go -o docs
