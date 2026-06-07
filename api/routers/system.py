from fastapi import APIRouter, Query
from api.db import db

router = APIRouter()


@router.get("/api/system")
async def get_system_info():
    # Collector health
    collector = await db.fetchrow(
        "SELECT status, last_success_at, last_error_at, updated_at "
        "FROM collector_health WHERE service_name = 'collector'"
    )

    # DB size
    db_size = await db.fetchrow(
        "SELECT ROUND(SUM(pg_total_relation_size(relid)) / 1048576.0, 1) AS size_mb "
        "FROM pg_stat_user_tables"
    )

    # Table counts
    tables = await db.fetch(
        "SELECT relname AS name, n_live_tup AS count "
        "FROM pg_stat_user_tables ORDER BY relname"
    )

    return {
        "collector": {
            "running": collector is not None,
            "ws_connected": collector["status"] == "ok" if collector else False,
            "uptime_hours": 0,
            "last_heartbeat": collector["updated_at"].isoformat() if collector and collector.get("updated_at") else None,
        },
        "db": {
            "size_mb": float(db_size["size_mb"]) if db_size and db_size["size_mb"] else 0,
            "tables": [{"name": r["name"], "count": r["count"]} for r in tables],
        },
        "crons": [
            {"schedule": "*/30 * * * *", "command": "scripts/run-feature-backfill.sh"},
            {"schedule": "0 2 * * 0", "command": "scripts/run-edge-study.sh"},
            {"schedule": "0 4 * * 0", "command": "scripts/run-xgb-retrain.sh"},
        ],
        "paper_trader": {
            "pid": None,
            "uptime_hours": 0,
        },
    }
