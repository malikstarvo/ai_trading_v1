from fastapi import APIRouter, Query
from api.db import db

router = APIRouter()


@router.get("/api/features/nan")
async def get_feature_nan(
    symbol: str = Query("BTCUSDT"),
    tf: str = Query("15m"),
):
    row = await db.fetchrow(
        """
        SELECT COUNT(*) AS total_rows,
               COUNT(oi_delta_1_pct)  AS nn_oi1,
               COUNT(oi_delta_4_pct)  AS nn_oi4,
               COUNT(oi_delta_12_pct) AS nn_oi12,
               COUNT(ls_ratio_raw)    AS nn_ls,
               COUNT(funding_rate)    AS nn_fr,
               COUNT(liq_long_usd)    AS nn_liq_l,
               COUNT(liq_short_usd)   AS nn_liq_s,
               COUNT(return_1)        AS nn_r1,
               COUNT(return_4)        AS nn_r4,
               COUNT(return_12)       AS nn_r12,
               COUNT(volatility_14)   AS nn_v14,
               COUNT(volatility_50)   AS nn_v50,
               COUNT(atr14)           AS nn_atr,
               COUNT(adx14)           AS nn_adx,
               COUNT(rsi14)           AS nn_rsi,
               COUNT(ema20)           AS nn_ema20,
               COUNT(volume_ema20)    AS nn_vol
        FROM feature_values
        WHERE symbol = $1 AND timeframe = $2
        """,
        symbol.upper(),
        tf,
    )
    if not row:
        return []

    total = row["total_rows"]
    columns = [
        ("oi_delta_1_pct", "nn_oi1"),
        ("oi_delta_4_pct", "nn_oi4"),
        ("oi_delta_12_pct", "nn_oi12"),
        ("ls_ratio_raw", "nn_ls"),
        ("funding_rate", "nn_fr"),
        ("liq_long_usd", "nn_liq_l"),
        ("liq_short_usd", "nn_liq_s"),
        ("return_1", "nn_r1"),
        ("return_4", "nn_r4"),
        ("return_12", "nn_r12"),
        ("volatility_14", "nn_v14"),
        ("volatility_50", "nn_v50"),
        ("atr14", "nn_atr"),
        ("adx14", "nn_adx"),
        ("rsi14", "nn_rsi"),
        ("ema20", "nn_ema20"),
        ("volume_ema20", "nn_vol"),
    ]
    result = []
    for col, key in columns:
        nn = row[key] or 0
        nan_pct = round((1 - nn / total) * 100, 1) if total > 0 else 100.0
        result.append({
            "column": col,
            "non_null": nn,
            "total_rows": total,
            "nan_pct": nan_pct,
        })
    return result


@router.get("/api/features/latest")
async def get_latest_features(
    symbol: str = Query("BTCUSDT"),
    tf: str = Query("15m"),
):
    row = await db.fetchrow(
        "SELECT * FROM feature_values "
        "WHERE symbol = $1 AND timeframe = $2 "
        "ORDER BY ts DESC LIMIT 1",
        symbol.upper(),
        tf,
    )
    if not row:
        return {}
    result = {}
    for k, v in row.items():
        if isinstance(v, float):
            result[k] = round(v, 6)
        elif isinstance(v, int):
            result[k] = v
        elif isinstance(v, bool):
            result[k] = v
        elif v is None:
            result[k] = None
        else:
            result[k] = str(v)
    return result


@router.get("/api/features/ranking")
async def get_feature_ranking():
    rows = await db.fetch(
        "SELECT rank, feature, score FROM research_results "
        "WHERE run_id = (SELECT MAX(id) FROM research_runs) "
        "ORDER BY rank"
    )
    return [
        {"rank": r["rank"], "feature": r["feature"], "score": round(float(r["score"]), 3)}
        for r in rows
    ]


@router.get("/api/stats/overview")
async def get_overview():
    query = """
    SELECT 'candles_15m' AS tbl, COUNT(*) AS cnt,
           MIN(time)::TEXT AS first_ts, MAX(time)::TEXT AS last_ts,
           ROUND(EXTRACT(EPOCH FROM (MAX(time) - MIN(time))) / 86400, 1) AS span_days
    FROM candles WHERE timeframe = '15m'
    UNION ALL
    SELECT 'candles_1h', COUNT(*), MIN(time)::TEXT, MAX(time)::TEXT,
           ROUND(EXTRACT(EPOCH FROM (MAX(time) - MIN(time))) / 86400, 1)
    FROM candles WHERE timeframe = '1h'
    UNION ALL
    SELECT 'feature_values', COUNT(*), MIN(ts)::TEXT, MAX(ts)::TEXT,
           ROUND(EXTRACT(EPOCH FROM (MAX(ts) - MIN(ts))) / 86400, 1)
    FROM feature_values
    UNION ALL
    SELECT 'training_labels', COUNT(*), MIN(ts)::TEXT, MAX(ts)::TEXT,
           ROUND(EXTRACT(EPOCH FROM (MAX(ts) - MIN(ts))) / 86400, 1)
    FROM training_labels
    """
    rows = await db.fetch(query)
    return [
        {
            "table": r["tbl"],
            "count": r["cnt"],
            "first_ts": r["first_ts"],
            "last_ts": r["last_ts"],
            "span_days": r["span_days"],
        }
        for r in rows
    ]


@router.get("/api/labels/distribution")
async def get_label_distribution(
    symbol: str = Query("BTCUSDT"),
    tf: str = Query("15m"),
):
    rows = await db.fetch(
        "SELECT horizon, COUNT(*) AS total, "
        "ROUND(AVG(CASE WHEN success THEN 1 ELSE 0 END), 4) AS success_rate "
        "FROM training_labels "
        "WHERE symbol = $1 AND timeframe = $2 "
        "GROUP BY horizon ORDER BY horizon",
        symbol.upper(),
        tf,
    )
    return [
        {
            "horizon": r["horizon"],
            "total": r["total"],
            "success_rate": round(float(r["success_rate"]), 4),
        }
        for r in rows
    ]
