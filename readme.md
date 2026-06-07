# AI Trading V1

Automated crypto trading system collecting perpetual futures data (candles, open interest, funding rates, LS ratios) with a real-time web dashboard.

**Goal:** Accumulate 60+ days of historical data → Train V2 ML model.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Collector** | Go (Bybit WebSocket + REST) |
| **Database** | TimescaleDB (PostgreSQL) |
| **Feature Engine** | Go (feature backfill) + Python (XGBoost) |
| **API** | Python FastAPI (asyncpg) |
| **Frontend** | React 19 + TypeScript + Tailwind CSS v4 |
| **UI Library** | shadcn/ui (Radix UI primitives) |
| **Charts** | Recharts (wrapped via shadcn ChartContainer) |
| **Real-time** | WebSocket push (FastAPI → React) |
| **Infra** | Docker Compose, Cloudflare Tunnel |

## Architecture

```
Bybit WS/REST
     │
     ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│  Collector   │───▶│  TimescaleDB │◀───│   FastAPI    │
│  (Go)        │    │  (Postgres)  │    │  (Python)   │
└─────────────┘    └──────────────┘    └──────┬──────┘
     │                                        │
     ▼                                        ▼
Feature Backfill ──▶ Edge Study ──▶ XGBoost   React SPA
(cron 30m)         (cron Sun 2AM)  (Sun 4AM)  (shadcn ui)
```

## Project Layout

```
├── api/                          # FastAPI backend
│   ├── main.py                   # App entry, lifespan, CORS
│   ├── db.py                     # asyncpg connection pool
│   └── routers/
│       ├── health.py             # GET /api/health
│       ├── market.py             # GET /api/candles, /api/orderflow, /api/liquidations
│       ├── features.py           # GET /api/features/*
│       ├── trading.py            # GET /api/paper/*
│       ├── model.py              # GET /api/model/status
│       ├── system.py             # GET /api/system
│       ├── data.py               # GET /api/data/overview (coverage + NaN stats)
│       └── ws.py                 # WebSocket /ws with 10s broadcast loop
├── cmd/
│   ├── trader/                   # Collector binary entry
│   ├── feature-backfill/         # Feature computation backfill
│   └── edge-study/               # Feature edge analysis
├── internal/
│   ├── collector/                # Bybit data collection (WS streams, REST polling)
│   ├── feature/                  # Feature computation engine
│   ├── agent/                    # Signal scorers (technical, orderflow, regime)
│   ├── backtest/                 # Backtesting engine
│   └── tradegate/                # Trade execution pipeline
├── scripts/
│   ├── run-feature-backfill.sh   # Cron: every 30 min
│   ├── run-edge-study.sh         # Cron: Sun 2AM UTC
│   ├── run-xgb-retrain.sh        # Cron: Sun 4AM UTC
│   ├── tunnel-url.sh             # Print current Cloudflare tunnel URL
│   └── xgb_meta/                 # Python XGBoost model
├── web/                          # React frontend
│   ├── src/
│   │   ├── components/ui/        # shadcn components (button, card, select, chart, etc.)
│   │   ├── components/layout/    # Sidebar
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx     # KPI cards, data progress bars, system health
│   │   │   ├── Market.tsx        # Close price chart, funding rate, OI
│   │   │   ├── Features.tsx      # NaN coverage, edge ranking, latest values
│   │   │   ├── Trading.tsx       # Positions, trades, equity curve
│   │   │   ├── Model.tsx         # AUC cards per horizon
│   │   │   └── System.tsx        # Service health, cron, DB tables
│   │   ├── hooks/useWebSocket.ts # WS hook with auto-reconnect
│   │   └── lib/api.ts            # Typed REST + WS client
│   ├── package.json
│   └── vite.config.ts
├── docker/
│   └── api.Dockerfile
├── docker-compose.yml            # postgres, trader (collector), api, prometheus, grafana
└── .env.example
```

## Services

| Service | Container | Port | Description |
|---------|-----------|------|-------------|
| **Postgres** | `postgres` | 5432 | TimescaleDB, hypertables for time-series |
| **Collector** | `trader` | — | Go binary: WS streams + REST polling |
| **API** | `api` | 8000 | FastAPI: REST + WebSocket |
| **Paper Trader** | `trader` (same) | — | Simulated trading inside collector |
| **Prometheus** | `prometheus` | 9090 | Metrics (optional) |
| **Grafana** | `grafana` | 3000 | Dashboards (optional) |
| **Cloudflare** | host systemd | — | `cloudflared` tunnel to localhost:8000 |

## Data Pipeline

```
Bybit WS (real-time)
  ├── Candles (15m, 1h) ──▶ candles hypertable
  ├── Open Interest ──────▶ open_interest table
  ├── Funding Rate ───────▶ funding_rates table
  ├── LS Ratio ───────────▶ ls_ratios table
  └── Liquidations ───────▶ liquidations table

Feature Backfill (cron 30m)
  └── Computes 20+ features ──▶ feature_values table
       (RSI14, ATR14, ADX14, OI delta, LS ratio, EMA, etc.)

Edge Study (cron Sun 2AM)
  └── Feature importance ranking ──▶ research_results table

XGBoost Retrain (cron Sun 4AM)
  ├── 4-bar model (~1h horizon)
  ├── 12-bar model (~3h horizon)
  └── 24-bar model (~6h horizon)
```

## Database

19 tables across public schema:

| Table | Purpose | Rows (approx) |
|-------|---------|---------------|
| `candles` | OHLCV per (symbol, timeframe) | 11,600+ |
| `feature_values` | Computed features | 10,200+ |
| `funding_rates` | Perpetual funding rates | 1,500,000+ |
| `open_interest` | Real-time OI snapshots | 28,000+ |
| `ls_ratios` | Long/short ratios | 2,100+ |
| `liquidations` | Liquidation events | 68 |
| `paper_account_snapshots` | Paper trader equity history | varies |
| `paper_positions` | Open positions | varies |
| `paper_trades` | Completed trades | 0 |
| `paper_orders` | Order history | 0 |
| `collector_health` | Collector heartbeat status | varies |
| `feature_sets` | Feature set definitions | 1 |
| `feature_jobs` | Backfill job tracking | varies |
| `training_labels` | Forward-return labels | varies |
| `research_results` | Edge study scores | varies |

## Data Coverage (as of Jun 7, 2026)

**Target: 60 days minimum**

| Symbol | TF | Candles | Days | Progress | OI NaN | LS NaN |
|--------|----|---------|------|----------|--------|--------|
| BTCUSDT | 15m | 3,096 | 32/60 | 53% | 24.2% | 0.0% |
| BTCUSDT | 1h | 773 | 32/60 | 53% | 92.7% | 0.0% |
| ETHUSDT | 15m | 3,096 | 32/60 | 53% | 24.1% | 0.0% |
| ETHUSDT | 1h | 773 | 32/60 | 53% | 92.7% | 0.0% |
| SOLUSDT | 15m | 3,096 | 32/60 | 53% | 28.7% | 0.0% |
| SOLUSDT | 1h | 773 | 32/60 | 53% | 96.6% | 0.0% |

**LS Ratio NaN: 0% across all symbols.** OI NaN improving daily as WS data accumulates.
Estimated 60-day completion: **July 5, 2026**.

## Web Interface (6 pages)

All components use **shadcn/ui** (Radix UI primitives). Charts use **ChartContainer** wrapping Recharts.

- **Dashboard** — KPI cards (candles, features, balance, uptime), Data progress bars per symbol/tf with 60-day target, System health, Recent trades
- **Market** — Close price line chart, Funding rate chart, Open Interest chart (all wrapped in ChartContainer)
- **Features** — Feature edge ranking bar chart, NaN coverage table with visual bars, Latest feature values
- **Trading** — Balance/PnL/Win rate cards, Equity curve area chart, PnL per trade bar chart, Positions table, Trades table
- **Model** — AUC cards per horizon (4/12/24 bar), AUC bar chart, Training pipeline info
- **System** — Collector/DB/Paper trader status cards, Cron schedule table, DB table row counts

### Real-time WebSocket

- Server broadcasts every 10s: candle close, features (RSI, ATR, ADX, OI delta, LS ratio), account (balance, equity, day PnL), health
- React hook `useWS(type)` with auto-reconnect exponential backoff (1-30s)
- Dashboard, Market, Trading pages merge live WS data into UI

## Quick Start

```bash
# Prerequisites
# - Go 1.26+, Python 3.11+, Docker & Docker Compose
# - Bybit API keys (testnet or mainnet)

# 1. Clone & configure
git clone <repo> ~/ai_trading_v1
cd ~/ai_trading_v1
cp .env.example .env
# Edit .env: set BYBIT_API_KEY, BYBIT_API_SECRET, adjust symbols/timeframes

# 2. Start database + collector + API
docker compose up -d

# 3. Verify
curl http://localhost:8000/api/health
curl http://localhost:8000/

# 4. Backfill features
go run ./cmd/feature-backfill/

# 5. Run edge study (after enough data)
go run ./cmd/edge-study/

# 6. Train XGBoost models
cd scripts/xgb_meta
python3 -m venv venv
venv/bin/pip install -r requirements.txt
venv/bin/python train.py --horizon 4
venv/bin/python train.py --horizon 12
venv/bin/python train.py --horizon 24
```

## Deployment (VPS)

```bash
# Docker services
docker compose up -d

# Cloudflare Tunnel (public access, no port opening)
# Install: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install
cloudflared tunnel --url http://localhost:8000

# Or use systemd service for auto-start:
sudo systemctl enable cloudflared-tunnel
sudo systemctl start cloudflared-tunnel

# Get current URL
./scripts/tunnel-url.sh
```

### Cron Jobs (installed on VPS)

```
*/30 * * * * /home/ubuntu/ai_trading_v1/scripts/run-feature-backfill.sh
0 2 * * 0 /home/ubuntu/ai_trading_v1/scripts/run-edge-study.sh
0 4 * * 0 /home/ubuntu/ai_trading_v1/scripts/run-xgb-retrain.sh
```

### Updating Frontend

```bash
# Build locally
cd web && npm run build

# Copy to VPS
scp -r web/dist ubuntu@<vps>:/home/ubuntu/ai_trading_v1/web/
ssh ubuntu@<vps> "docker cp /home/ubuntu/ai_trading_v1/web/dist <api-container>:/app/web/"

# Or git push + pull on VPS
```

## Feature Set (V1)

| Feature | Source | Weight |
|---------|--------|--------|
| technical_score | Trend(35) + Momentum(25) + Volume(20) + Volatility(10) + ADX(10) | 40% |
| orderflow_score | Funding(20) + OI Delta(25) + LS Ratio(30) + Liquidation(25) | 40% |
| regime_score | ADX trend + ATR volatility | 20% |
| confidence_score | tech*0.4 + of*0.4 + regime*0.2 | — |
| rsi14, atr14, adx14 | Technical indicators | — |
| funding_rate | Perpetual funding | — |
| oi_delta_1/4/12_pct | OI % change | — |
| oi_zscore_30 | OI z-score (30 bars) | — |
| ls_ratio_raw, ls_ratio_normalized | Long/short balance | — |
| volume_delta | log(volume / volume_ema20) | — |

## Paper Trader

Runs inside the collector. Polls latest candle every 60s, scores all agents, and simulates trades without real capital.

```bash
# Status
screen -ls

# Logs
tail -f paper-trader.log

# Docker exec queries
docker exec ai_trading_v1-postgres-1 psql -U trader -d ai_trading \
  -c 'SELECT ts, balance, equity FROM paper_account_snapshots ORDER BY ts DESC LIMIT 5;'
```

## Database Queries

```bash
# Latest candle
docker exec ai_trading_v1-postgres-1 psql -U trader -d ai_trading \
  -c "SELECT time, close FROM candles WHERE symbol='BTCUSDT' AND timeframe='15m' ORDER BY time DESC LIMIT 1;"

# Feature lag
docker exec ai_trading_v1-postgres-1 psql -U trader -d ai_trading \
  -c "SELECT 'candle' as src, max(time) FROM candles WHERE symbol='BTCUSDT' AND timeframe='15m' UNION ALL SELECT 'feature', max(ts) FROM feature_values WHERE symbol='BTCUSDT' AND timeframe='15m';"

# Data coverage
docker exec ai_trading_v1-postgres-1 psql -U trader -d ai_trading \
  -c "SELECT symbol, timeframe, count(*), min(time)::text, max(time)::text, round(extract(epoch from (max(time)-min(time)))/86400) as days_span FROM candles GROUP BY symbol, timeframe ORDER BY symbol, timeframe;"
```
