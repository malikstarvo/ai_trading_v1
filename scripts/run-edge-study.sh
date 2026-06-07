#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p logs
TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
echo "[edge-study] $TIMESTAMP - Starting" >> logs/edge-study.log
./edge-study >> logs/edge-study.log 2>&1
echo "[edge-study] $TIMESTAMP - Done" >> logs/edge-study.log
