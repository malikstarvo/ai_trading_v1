import asyncio
import json
import time
from fastapi import APIRouter, WebSocket, WebSocketDisconnect
from api.db import db

router = APIRouter()


class ConnectionManager:
    def __init__(self):
        self.clients: set[WebSocket] = set()
        self.last_broadcast: dict[str, float] = {}

    async def connect(self, ws: WebSocket):
        await ws.accept()
        self.clients.add(ws)

    def disconnect(self, ws: WebSocket):
        self.clients.discard(ws)

    async def broadcast(self, msg: dict):
        data = json.dumps(msg, default=str)
        stale = set()
        for ws in self.clients:
            try:
                await ws.send_text(data)
            except Exception:
                stale.add(ws)
        self.clients -= stale


manager = ConnectionManager()


async def collect_snapshot() -> dict:
    """Collect latest data from DB for broadcast."""
    result: dict = {"t": time.time()}

    try:
        row = await db.fetchrow(
            "SELECT time, open, high, low, close, volume "
            "FROM candles WHERE symbol = 'BTCUSDT' AND timeframe = '15m' "
            "ORDER BY time DESC LIMIT 1"
        )
        if row:
            result["candle"] = {
                "time": row["time"].isoformat(),
                "close": float(row["close"]),
                "volume": float(row["volume"]),
            }
    except Exception:
        pass

    try:
        row = await db.fetchrow(
            "SELECT * FROM feature_values "
            "WHERE symbol = 'BTCUSDT' AND timeframe = '15m' "
            "ORDER BY ts DESC LIMIT 1"
        )
        if row:
            out = {}
            for col in ("ts", "rsi14", "atr14", "adx14", "oi_delta_4_pct", "ls_ratio_raw"):
                v = row.get(col)
                if isinstance(v, float):
                    out[col] = round(v, 4)
                elif v is not None:
                    out[col] = str(v)
            result["feature"] = out
    except Exception:
        pass

    try:
        row = await db.fetchrow(
            "SELECT balance, equity, day_pnl, day_trades "
            "FROM paper_account_snapshots ORDER BY ts DESC LIMIT 1"
        )
        if row:
            result["account"] = {
                "balance": float(row["balance"]),
                "equity": float(row["equity"]),
                "day_pnl": float(row["day_pnl"]),
                "day_trades": int(row["day_trades"]),
            }
    except Exception:
        pass

    try:
        row = await db.fetchrow(
            "SELECT status FROM collector_health WHERE service_name = 'collector'"
        )
        if row:
            result["health"] = {"collector": row["status"]}
    except Exception:
        pass

    return result


async def broadcast_loop(interval: float = 10.0):
    """Periodically collect and broadcast data to all connected clients."""
    while True:
        try:
            snapshot = await collect_snapshot()
            if manager.clients:
                await manager.broadcast(snapshot)
        except Exception:
            pass
        await asyncio.sleep(interval)


@router.websocket("/ws")
async def websocket_endpoint(ws: WebSocket):
    await manager.connect(ws)
    try:
        # Send an initial snapshot immediately
        snapshot = await collect_snapshot()
        await ws.send_text(json.dumps(snapshot, default=str))

        # Keep connection alive — handle client messages
        while True:
            try:
                data = await asyncio.wait_for(ws.receive_text(), timeout=30)
                # Client can send "ping", respond "pong"
                if data == "ping":
                    await ws.send_text(json.dumps({"pong": True}))
            except asyncio.TimeoutError:
                # Send heartbeat to check connection
                try:
                    await ws.send_text(json.dumps({"heartbeat": True}))
                except Exception:
                    break
    except WebSocketDisconnect:
        pass
    except Exception:
        pass
    finally:
        manager.disconnect(ws)
