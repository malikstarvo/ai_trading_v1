#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p logs
TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
echo "[xgb-retrain] $TIMESTAMP - Starting" >> logs/xgb-retrain.log
bash scripts/xgb_meta/run_all.sh >> logs/xgb-retrain.log 2>&1
echo "[xgb-retrain] $TIMESTAMP - Done (exit=$?)" >> logs/xgb-retrain.log
