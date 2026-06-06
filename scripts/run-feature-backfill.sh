#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
exec go run ./cmd/feature-backfill/ >> /tmp/feature-backfill.log 2>&1
