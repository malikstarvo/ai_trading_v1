from fastapi import APIRouter
from api.db import db

router = APIRouter()

TARGET_DAYS = 60


@router.get("/api/data/overview")
async def data_overview():
    candles_rows = await db.fetch(
        "SELECT symbol, timeframe, count(*)::int as rows, min(time) as first_ts, max(time) as last_ts, "
        "round(extract(epoch from (max(time) - min(time)))/86400)::int as days_span "
        "FROM candles GROUP BY symbol, timeframe ORDER BY symbol, timeframe"
    )

    feature_nan = await db.fetch(
        "SELECT symbol, timeframe, feature_set_id, count(*)::int as total_rows, "
        "round(100.0 * sum(case when oi_delta_1_pct is null or oi_delta_1_pct <> oi_delta_1_pct then 1 else 0 end) / count(*), 1)::text as oi_nan_pct, "
        "round(100.0 * sum(case when ls_ratio_normalized is null or ls_ratio_normalized <> ls_ratio_normalized then 1 else 0 end) / count(*), 1)::text as ls_nan_pct "
        "FROM feature_values GROUP BY symbol, timeframe, feature_set_id ORDER BY symbol, timeframe, feature_set_id"
    )

    feature_rows = await db.fetch(
        "SELECT symbol, timeframe, count(*)::int as rows, min(ts) as first_ts, max(ts) as last_ts, "
        "round(extract(epoch from (max(ts) - min(ts)))/86400)::int as days_span "
        "FROM feature_values GROUP BY symbol, timeframe ORDER BY symbol, timeframe"
    )

    funding_row = await db.fetchrow(
        "SELECT count(*)::int as rows, min(time) as first_ts, max(time) as last_ts "
        "FROM funding_rates WHERE time > '2020-01-01'"
    )

    oi_row = await db.fetchrow(
        "SELECT count(*)::int as rows, min(time) as first_ts, max(time) as last_ts, "
        "round(extract(epoch from (max(time) - min(time)))/86400)::int as days_span "
        "FROM open_interest"
    )

    ls_row = await db.fetchrow(
        "SELECT count(*)::int as rows, min(time) as first_ts, max(time) as last_ts, "
        "round(extract(epoch from (max(time) - min(time)))/86400)::int as days_span "
        "FROM ls_ratios"
    )

    liq_row = await db.fetchrow(
        "SELECT count(*)::int as rows, min(time) as first_ts, max(time) as last_ts, "
        "round(extract(epoch from (max(time) - min(time)))/86400)::int as days_span "
        "FROM liquidations"
    )

    combined = {}
    for c in candles_rows:
        key = (c["symbol"], c["timeframe"])
        combined[key] = {
            "symbol": c["symbol"],
            "timeframe": c["timeframe"],
            "candle_rows": c["rows"],
            "candle_first_ts": str(c["first_ts"]) if c["first_ts"] else None,
            "candle_last_ts": str(c["last_ts"]) if c["last_ts"] else None,
            "candle_days_span": c["days_span"],
            "feature_rows": 0,
            "feature_first_ts": None,
            "feature_last_ts": None,
            "feature_days_span": 0,
            "oi_nan_pct": None,
            "ls_nan_pct": None,
        }

    for f in feature_rows:
        key = (f["symbol"], f["timeframe"])
        if key in combined:
            combined[key]["feature_rows"] = f["rows"]
            combined[key]["feature_first_ts"] = str(f["first_ts"]) if f["first_ts"] else None
            combined[key]["feature_last_ts"] = str(f["last_ts"]) if f["last_ts"] else None
            combined[key]["feature_days_span"] = f["days_span"]

    for n in feature_nan:
        key = (n["symbol"], n["timeframe"])
        if key in combined and n["feature_set_id"] == 1:
            combined[key]["oi_nan_pct"] = n["oi_nan_pct"]
            combined[key]["ls_nan_pct"] = n["ls_nan_pct"]

    for key, val in combined.items():
        val["candle_progress_pct"] = round(min(val["candle_days_span"] / TARGET_DAYS * 100, 100), 1)
        val["target_days"] = TARGET_DAYS
        remaining = TARGET_DAYS - val["candle_days_span"]
        val["remaining_days"] = max(remaining, 0)
        if val["candle_days_span"] > 0:
            rate = val["candle_days_span"] / val["candle_days_span"]
            val["est_completion_date"] = "—"
        else:
            val["est_completion_date"] = "—"

    return {
        "symbols": sorted(combined.values(), key=lambda x: (x["symbol"], x["timeframe"])),
        "funding_rates": {
            "rows": funding_row["rows"] if funding_row else 0,
            "first_ts": str(funding_row["first_ts"]) if funding_row and funding_row["first_ts"] else None,
            "last_ts": str(funding_row["last_ts"]) if funding_row and funding_row["last_ts"] else None,
        } if funding_row else None,
        "open_interest": {
            "rows": oi_row["rows"] if oi_row else 0,
            "first_ts": str(oi_row["first_ts"]) if oi_row and oi_row["first_ts"] else None,
            "last_ts": str(oi_row["last_ts"]) if oi_row and oi_row["last_ts"] else None,
            "days_span": oi_row["days_span"] if oi_row else 0,
        } if oi_row else None,
        "ls_ratios": {
            "rows": ls_row["rows"] if ls_row else 0,
            "first_ts": str(ls_row["first_ts"]) if ls_row and ls_row["first_ts"] else None,
            "last_ts": str(ls_row["last_ts"]) if ls_row and ls_row["last_ts"] else None,
            "days_span": ls_row["days_span"] if ls_row else 0,
        } if ls_row else None,
        "liquidations": {
            "rows": liq_row["rows"] if liq_row else 0,
            "first_ts": str(liq_row["first_ts"]) if liq_row and liq_row["first_ts"] else None,
            "last_ts": str(liq_row["last_ts"]) if liq_row and liq_row["last_ts"] else None,
            "days_span": liq_row["days_span"] if liq_row else 0,
        } if liq_row else None,
        "target_days": TARGET_DAYS,
    }
