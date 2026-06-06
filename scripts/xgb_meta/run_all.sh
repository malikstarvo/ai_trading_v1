#!/usr/bin/env bash
#
# run_all.sh — Full XGBoost Meta Model Pipeline
#
# Prerequisites:
#   - Database already running and populated (docker compose up -d)
#   - collector + feature-backfill already completed
#   - Go toolchain available (for parity test reference if needed)
#
# Usage:
#   bash scripts/xgb_meta/run_all.sh
#
# Exits on first failure. Run from repo root or scripts/xgb_meta/.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT_DIR="$REPO_ROOT/scripts/xgb_meta"
MODEL_DIR="$SCRIPT_DIR/models"

SYMBOL="BTCUSDT"
TIMEFRAME="15m"
THRESHOLD_AUC=0.60

echo "========================================"
echo " XGBoost Meta Model — Full Pipeline"
echo "========================================"
echo "Symbol:     $SYMBOL"
echo "Timeframe:  $TIMEFRAME"
echo "Min AUC:    $THRESHOLD_AUC"
echo "========================================"

# ---- Check prerequisites ----
echo ""
echo "[1/5] Checking prerequisites..."

command -v python3 >/dev/null 2>&1 || { echo "ERROR: python3 not found"; exit 1; }
command -v pg_isready >/dev/null 2>&1 || { echo "WARNING: pg_isready not found, skipping DB check"; }

if command -v pg_isready >/dev/null 2>&1; then
    if pg_isready -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-trader}" -d "${DB_NAME:-ai_trading}" >/dev/null 2>&1; then
        echo "  Database: OK"
    else
        echo "  WARNING: Database not reachable. Did you run 'docker compose up -d' ?"
        echo "  Continuing anyway (may fail later)..."
    fi
fi

cd "$SCRIPT_DIR"

# ---- Virtual environment ----
echo ""
echo "[2/5] Setting up virtual environment..."
if [ ! -d "venv" ]; then
    python3 -m venv venv
    echo "  Created venv/"
fi
venv/bin/pip install -q -r requirements.txt
echo "  Dependencies installed"

# ---- Parity test ----
echo ""
echo "[3/5] Running parity test (blocking gate)..."
venv/bin/python parity_test.py
echo "  Parity confirmed."

# ---- Train + Validate ----
ALL_PASSED=true
for H in 4 12 24; do
    echo ""
    echo "----------------------------------------"
    echo " Horizon: ${H}-bar"
    echo "----------------------------------------"

    echo "[4/$((3+H/10))] Training ${H}-bar model..."
    venv/bin/python train.py --horizon "$H" --symbol "$SYMBOL" --timeframe "$TIMEFRAME"

    MODEL_FILE="xgb_v1.0_${H}bar.pkl"
    MODEL_PATH="$MODEL_DIR/$MODEL_FILE"
    if [ ! -f "$MODEL_PATH" ]; then
        echo "  ERROR: Model not found at $MODEL_PATH"
        ALL_PASSED=false
        continue
    fi

    echo "[5/$((3+H/10))] Validating ${H}-bar model..."
    venv/bin/python validate.py --model "$MODEL_PATH" --horizon "$H" --symbol "$SYMBOL" --timeframe "$TIMEFRAME"

    # Check test AUC threshold
    REPORT_FILE="$MODEL_DIR/report_v1.0_${H}bar.json"
    if [ -f "$REPORT_FILE" ]; then
        AUC=$(python3 -c "import json; print(json.load(open('$REPORT_FILE'))['test_auc'])")
        if [ "$(echo "$AUC >= $THRESHOLD_AUC" | bc -l 2>/dev/null)" = "1" ] || python3 -c "exit(0 if $AUC >= $THRESHOLD_AUC else 1)"; then
            echo "  Test AUC $AUC >= $THRESHOLD_AUC — PASS"
        else
            echo "  WARNING: Test AUC $AUC < $THRESHOLD_AUC"
            ALL_PASSED=false
        fi
    fi
done

# ---- Summary ----
echo ""
echo "========================================"
if [ "$ALL_PASSED" = true ]; then
    echo " RESULT: ALL MODELS PASSED"
else
    echo " RESULT: Some models did not meet thresholds"
fi
echo "========================================"
echo ""
echo "Models saved in: $MODEL_DIR"
ls -la "$MODEL_DIR"/*.pkl 2>/dev/null || echo "(no models found)"
echo ""
echo "To run a prediction:"
echo "  venv/bin/python predict.py --ts \"2025-06-06T12:00Z\" --horizon 4"
