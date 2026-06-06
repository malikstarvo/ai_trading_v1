# AI Trading V1 — Server Setup & Operations

## Prerequisites

- Go 1.26+
- Python 3.11+
- Docker & Docker Compose
- Git
- Bybit API access (testnet or mainnet)

## Quick Start

```bash
# 1. Clone & configure
git clone <repo> ~/ai_trading_v1
cd ~/ai_trading_v1
cp .env.example .env
# Edit .env: set DB_HOST=localhost, adjust symbols/timeframes

# 2. Start database + monitoring
docker compose up -d postgres
# Wait ~10s, then verify:
docker compose exec postgres pg_isready -U trader -d ai_trading

# 3. Start collector (candles, OI, funding, L/S, liquidations)
docker compose up -d trader
# Check logs: docker compose logs -f trader

# 4. Backfill features + training labels
go run ./cmd/feature-backfill/

# 5. Run edge study (optional, research)
go run ./cmd/edge-study/
```

## Project Layout

```
├── cmd/
│   ├── trader/            # Collector binary
│   ├── backtest/          # Backtester CLI
│   ├── feature-backfill/  # Feature + label backfiller
│   └── edge-study/        # Edge analysis
├── internal/
│   ├── agent/technical/   # Technical signal scorer
│   ├── agent/orderflow/   # Orderflow signal scorer
│   ├── agent/regime/      # Market regime classifier
│   ├── agent/tradegate/   # Trade gate pipeline
│   ├── backtest/          # Backtest engine
│   └── collector/         # Bybit data collection
├── scripts/xgb_meta/      # XGBoost meta model (Python)
├── docker-compose.yml
└── .env
```

## XGBoost Meta Model Training

The meta model combines Technical (40%) + OrderFlow (40%) + Regime (20%) agent scores into a single probability using XGBoost. Three horizon models are trained separately: 4-bar (~1h), 12-bar (~3h), 24-bar (~6h).

### Setup

```bash
cd scripts/xgb_meta

# Create virtual environment
python3 -m venv venv

# Install dependencies
venv/bin/pip install -r requirements.txt
```

### 1. Parity Test (Blocking Gate)

Python scorers must match Go reference values within ±0.01 before training is allowed.

```bash
venv/bin/python parity_test.py
```

Expected output:
```
Passed: 5/5  Failed: 0/5
ALL PASSED - parity confirmed. Ready for training.
```

### 2. Train Models

```bash
# Train all 3 horizon models
venv/bin/python train.py --horizon 4
venv/bin/python train.py --horizon 12
venv/bin/python train.py --horizon 24
```

Each run:
- Loads data from DB (`feature_values` + `training_labels` + `candles`)
- Computes 9 features on-the-fly via Python re-implementation of Go scorers
- Grid search 48 hyperparameter combinations (val AUC)
- Saves model to `models/xgb_v1.0_{horizon}bar.pkl`
- Saves report to `models/report_v1.0_{horizon}bar.json`

### 3. Validate

```bash
venv/bin/python validate.py --model models/xgb_v1.0_4bar.pkl
venv/bin/python validate.py --model models/xgb_v1.0_12bar.pkl
venv/bin/python validate.py --model models/xgb_v1.0_24bar.pkl
```

Key thresholds: Test AUC ≥ 0.60, Profit Factor ≥ 1.1.

### 4. Predict

Query a specific timestamp for live probability:

```bash
venv/bin/python predict.py --ts "2025-06-06T12:00Z" --horizon 4
```

Output:
```
Probability: 0.7234
Decision:    TRADE (threshold: 0.45)
```

## Feature Set

| # | Feature | Source |
|---|---------|--------|
| 1 | technical_score | Trend(35) + Momentum(25) + Volume(20) + Volatility(10) + ADX Bonus(10) |
| 2 | orderflow_score | Funding(20) + OI Delta(25) + LS Ratio(30) + Liquidation(25) |
| 3 | regime_score | ADX trend component + ATR/Volatility component |
| 4 | confidence_score | tech*0.4 + of*0.4 + regime*0.2 |
| 5 | atr14 | Average True Range (14-bar) |
| 6 | adx14 | ADX (14-bar) |
| 7 | funding_rate | Perpetual funding rate |
| 8 | oi_delta_1_pct | Open Interest 1-bar % change |
| 9 | volume_delta | log(volume / volume_ema20) |

## Automation

```bash
# Full ML pipeline (after DB is already running)
bash scripts/xgb_meta/run_all.sh
```

See `scripts/xgb_meta/run_all.sh` for details.


On the VPS, just run bash scripts/xgb_meta/run_all.sh