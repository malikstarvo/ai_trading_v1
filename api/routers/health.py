from fastapi import APIRouter
from api.db import db

router = APIRouter()


@router.get("/api/health")
async def health():
    db_ok = False
    try:
        row = await db.fetchrow("SELECT 1 AS ok")
        db_ok = row is not None
    except Exception:
        pass

    collector_ok = False
    try:
        row = await db.fetchrow(
            "SELECT status FROM collector_health WHERE service_name = 'collector'"
        )
        collector_ok = row is not None and row["status"] in ("ok", "healthy")
    except Exception:
        pass

    paper_status = "stopped"
    try:
        row = await db.fetchrow(
            "SELECT COUNT(*) AS cnt FROM paper_account_snapshots"
        )
        if row and row["cnt"] > 0:
            paper_status = "running"
    except Exception:
        pass

    uptime = 0
    try:
        row = await db.fetchrow(
            "SELECT EXTRACT(EPOCH FROM NOW() - MIN(updated_at)) / 3600 AS hours "
            "FROM collector_health"
        )
        if row:
            uptime = round(float(row["hours"] or 0), 1)
    except Exception:
        pass

    return {
        "db": "ok" if db_ok else "error",
        "collector": "ok" if collector_ok else "error",
        "paper_trader": paper_status,
        "uptime_hours": uptime,
    }
