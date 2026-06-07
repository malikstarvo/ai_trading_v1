from fastapi import APIRouter, Query
from api.db import db

router = APIRouter()


@router.get("/api/candles")
async def get_candles(
    symbol: str = Query("BTCUSDT"),
    tf: str = Query("15m"),
    limit: int = Query(200),
):
    rows = await db.fetch(
        "SELECT time, open, high, low, close, volume "
        "FROM candles "
        "WHERE symbol = $1 AND timeframe = $2 "
        "ORDER BY time DESC LIMIT $3",
        symbol.upper(),
        tf,
        limit,
    )
    rows.reverse()
    return [
        {
            "time": r["time"].isoformat(),
            "open": float(r["open"]),
            "high": float(r["high"]),
            "low": float(r["low"]),
            "close": float(r["close"]),
            "volume": float(r["volume"]),
        }
        for r in rows
    ]


@router.get("/api/orderflow")
async def get_orderflow(
    symbol: str = Query("BTCUSDT"),
    limit: int = Query(200),
):
    oi = await db.fetch(
        "SELECT time, oi, oi_value_usd "
        "FROM open_interest "
        "WHERE symbol = $1 "
        "ORDER BY time DESC LIMIT $2",
        symbol.upper(),
        limit,
    )
    fr = await db.fetch(
        "SELECT time, rate AS funding_rate "
        "FROM funding_rates "
        "WHERE symbol = $1 AND time > '2020-01-01' "
        "ORDER BY time DESC LIMIT $2",
        symbol.upper(),
        limit,
    )
    ls = await db.fetch(
        "SELECT time, buy_ratio, sell_ratio "
        "FROM ls_ratios "
        "WHERE symbol = $1 AND period = '5m' "
        "ORDER BY time DESC LIMIT $2",
        symbol.upper(),
        limit,
    )

    merged: dict = {}
    for r in oi:
        t = r["time"].isoformat()
        merged.setdefault(t, {"time": t})["oi"] = float(r["oi"])
        merged[t]["oi_value_usd"] = float(r["oi_value_usd"])
    for r in fr:
        t = r["time"].isoformat()
        merged.setdefault(t, {"time": t})["funding_rate"] = float(r["funding_rate"])
    for r in ls:
        t = r["time"].isoformat()
        merged.setdefault(t, {"time": t})["buy_ratio"] = float(r["buy_ratio"])
        merged[t]["sell_ratio"] = float(r["sell_ratio"])

    result = sorted(merged.values(), key=lambda x: x["time"])
    return result[-limit:]


@router.get("/api/liquidations")
async def get_liquidations(
    symbol: str = Query("BTCUSDT"),
    limit: int = Query(100),
):
    rows = await db.fetch(
        "SELECT time, side, size, price, value_usd "
        "FROM liquidations "
        "WHERE symbol = $1 "
        "ORDER BY time DESC LIMIT $2",
        symbol.upper(),
        limit,
    )
    return [
        {
            "time": r["time"].isoformat(),
            "side": r["side"],
            "size": float(r["size"]),
            "price": float(r["price"]),
            "value_usd": float(r["value_usd"]),
        }
        for r in rows
    ]
