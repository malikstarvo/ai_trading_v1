from fastapi import APIRouter, Query
from api.db import db

router = APIRouter()


@router.get("/api/paper/status")
async def get_paper_status():
    engine = await db.fetchrow("SELECT * FROM paper_trading_state ORDER BY updated_at DESC LIMIT 1")
    account = await db.fetchrow(
        "SELECT balance, equity, unrealized_pnl, day_pnl, day_trades "
        "FROM paper_account_snapshots ORDER BY ts DESC LIMIT 1"
    )

    if engine and engine.get("config"):
        config = engine["config"]
    else:
        config = {}

    return {
        "state": engine["state"] if engine else "stopped",
        "symbol": config.get("symbol", "BTCUSDT"),
        "timeframe": config.get("timeframe", "15m"),
        "initial_capital": float(config.get("initial_capital", 10000)),
        "balance": float(account["balance"]) if account else 10000,
        "equity": float(account["equity"]) if account else 10000,
        "day_pnl": float(account["day_pnl"]) if account else 0,
        "day_trades": int(account["day_trades"]) if account else 0,
        "total_pnl": float(account["balance"]) - float(config.get("initial_capital", 10000))
        if account
        else 0,
        "uptime_hours": 0,
    }


@router.get("/api/paper/positions")
async def get_paper_positions(status: str = Query("")):
    if status:
        rows = await db.fetch(
            "SELECT * FROM paper_positions WHERE status = $1 ORDER BY open_ts DESC",
            status,
        )
    else:
        rows = await db.fetch(
            "SELECT * FROM paper_positions ORDER BY open_ts DESC LIMIT 50"
        )
    return [
        {
            "id": r["id"],
            "symbol": r["symbol"],
            "timeframe": r["timeframe"],
            "direction": r["direction"],
            "entry_price": float(r["entry_price"]),
            "quantity": float(r["quantity"]),
            "stop_price": float(r["stop_price"]),
            "open_ts": r["open_ts"].isoformat(),
            "bars_held": r["bars_held"],
            "status": r["status"],
            "exit_price": float(r["exit_price"]) if r.get("exit_price") else None,
            "net_pnl": float(r["net_pnl"]) if r.get("net_pnl") else None,
            "return_pct": float(r["return_pct"]) if r.get("return_pct") else None,
            "exit_reason": r.get("exit_reason"),
        }
        for r in rows
    ]


@router.get("/api/paper/trades")
async def get_paper_trades(limit: int = Query(50)):
    rows = await db.fetch(
        "SELECT * FROM paper_trades ORDER BY exit_ts DESC NULLS LAST LIMIT $1",
        limit,
    )
    return [
        {
            "id": r["id"],
            "symbol": r["symbol"],
            "direction": r["direction"],
            "entry_ts": r["entry_ts"].isoformat(),
            "exit_ts": r["exit_ts"].isoformat() if r.get("exit_ts") else None,
            "entry_price": float(r["entry_price"]),
            "exit_price": float(r["exit_price"]),
            "net_pnl": float(r["net_pnl"]),
            "return_pct": float(r["return_pct"]),
            "holding_bars": r["holding_bars"],
            "exit_reason": r.get("exit_reason"),
        }
        for r in rows
    ]


@router.get("/api/paper/account")
async def get_paper_account(limit: int = Query(200)):
    rows = await db.fetch(
        "SELECT ts, balance, equity, unrealized_pnl, day_pnl, day_trades "
        "FROM paper_account_snapshots ORDER BY ts DESC LIMIT $1",
        limit,
    )
    rows.reverse()
    return [
        {
            "ts": r["ts"].isoformat(),
            "balance": float(r["balance"]),
            "equity": float(r["equity"]),
            "unrealized_pnl": float(r["unrealized_pnl"]),
            "day_pnl": float(r["day_pnl"]),
            "day_trades": r["day_trades"],
        }
        for r in rows
    ]
