import math
import os
import glob
import json
import numpy as np
import joblib
from fastapi import APIRouter, Query
from pydantic import BaseModel
from api.db import db

router = APIRouter()

MODELS_DIR = "scripts/xgb_meta/models"
_cached_models: dict[int, dict] = {}
_cached_models_loaded = False

FEATURE_COLUMNS = ["atr14", "adx14", "funding_rate", "volume_delta"]

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))


def _get_models_dir():
    return os.path.join(PROJECT_ROOT, MODELS_DIR)


def _load_models():
    global _cached_models, _cached_models_loaded
    if _cached_models_loaded:
        return
    models_dir = _get_models_dir()
    for horizon in [4, 12, 24]:
        pattern = os.path.join(models_dir, f"xgb_*_{horizon}bar.pkl")
        files = glob.glob(pattern)
        if not files:
            continue
        path = max(files, key=os.path.getmtime)
        try:
            artifact = joblib.load(path)
            _cached_models[horizon] = artifact
        except Exception:
            pass
    _cached_models_loaded = True


def _get_features_from_row(row: dict) -> dict:
    volume = row.get("volume") or 0
    volume_ema20 = row.get("volume_ema20") or 0
    vol_delta = 0.0
    if volume > 0 and volume_ema20 > 0:
        vol_delta = math.log(volume / volume_ema20)
    return {
        "atr14": row.get("atr14") if row.get("atr14") is not None and not (isinstance(row.get("atr14"), float) and (math.isnan(row["atr14"]) or math.isinf(row["atr14"]))) else 0.0,
        "adx14": row.get("adx14") if row.get("adx14") is not None and not (isinstance(row.get("adx14"), float) and (math.isnan(row["adx14"]) or math.isinf(row["adx14"]))) else 0.0,
        "funding_rate": row.get("funding_rate") if row.get("funding_rate") is not None and not (isinstance(row.get("funding_rate"), float) and (math.isnan(row["funding_rate"]) or math.isinf(row["funding_rate"]))) else 0.0,
        "volume_delta": vol_delta,
    }


@router.get("/api/model/status")
async def get_model_status():
    models_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(__file__))), MODELS_DIR) # noqa
    results = []
    for horizon in [4, 12, 24]:
        model_path = os.path.join(models_dir, f"xgb_v1.0_{horizon}bar.pkl")
        report_path = os.path.join(models_dir, f"report_v1.0_{horizon}bar.json")

        exists = os.path.isfile(model_path)
        auc = None
        trained_at = None
        if exists and os.path.isfile(report_path):
            try:
                with open(report_path) as f:
                    report = json.load(f)
                auc = report.get("test_auc")
                trained_at = report.get("trained_at")
            except Exception:
                pass

        results.append({
            "horizon": horizon,
            "exists": exists,
            "auc": round(auc, 4) if auc else None,
            "trained_at": trained_at,
        })

    return {
        "symbol": "BTCUSDT",
        "timeframe": "15m",
        "models": results,
    }


@router.get("/api/model/predict")
async def predict(
    ts: str = Query(..., description="Timestamp (ISO format)"),
    horizon: int = Query(4, description="Forecast horizon in bars (4, 12, 24)"),
    symbol: str = Query("BTCUSDT"),
    timeframe: str = Query("15m"),
):
    _load_models()

    # If no model for this horizon, return neutral prob (no gating)
    if horizon not in _cached_models:
        return {
            "prob": 1.0,
            "horizon": horizon,
            "model_exists": False,
            "features": {},
        }

    # Query feature_values + candles for the given timestamp
    row = await db.fetchrow("""
        SELECT
            fv.atr14, fv.adx14, fv.funding_rate,
            fv.volume_ema20,
            c.volume
        FROM feature_values fv
        JOIN candles c
            ON fv.symbol = c.symbol
            AND fv.timeframe = c.timeframe
            AND fv.ts = c.time
        WHERE fv.symbol = $1
          AND fv.timeframe = $2
          AND fv.ts = $3
        LIMIT 1
    """, symbol, timeframe, ts)

    if row is None:
        return {"prob": 1.0, "horizon": horizon, "model_exists": True, "features": {}, "error": "no_data"}

    features = _get_features_from_row(row)
    artifact = _cached_models[horizon]
    model = artifact["model"]
    model_features = artifact.get("features", FEATURE_COLUMNS)
    scaler = artifact.get("scaler")

    X = np.array([[features[f] for f in model_features]], dtype=np.float64)

    if scaler is not None:
        X = scaler.transform(X)

    prob = float(model.predict_proba(X)[0, 1])

    return {
        "prob": round(prob, 4),
        "horizon": horizon,
        "model_exists": True,
        "features": features,
    }


class PredictBatchItem(BaseModel):
    ts: str


class PredictBatchRequest(BaseModel):
    timestamps: list[PredictBatchItem]
    horizon: int = 4
    symbol: str = "BTCUSDT"
    timeframe: str = "15m"


class PredictBatchResponse(BaseModel):
    results: list[dict]


@router.post("/api/model/predict-batch")
async def predict_batch(body: PredictBatchRequest):
    _load_models()

    horizon = body.horizon
    if horizon not in _cached_models:
        return PredictBatchResponse(results=[
            {"ts": item.ts, "prob": 1.0, "model_exists": False}
            for item in body.timestamps
        ])

    artifact = _cached_models[horizon]
    model = artifact["model"]
    model_features = artifact.get("features", FEATURE_COLUMNS)
    scaler = artifact.get("scaler")

    results = []
    for item in body.timestamps:
        ts = item.ts
        row = await db.fetchrow("""
            SELECT
                fv.atr14, fv.adx14, fv.funding_rate,
                fv.volume_ema20,
                c.volume
            FROM feature_values fv
            JOIN candles c
                ON fv.symbol = c.symbol
                AND fv.timeframe = c.timeframe
                AND fv.ts = c.time
            WHERE fv.symbol = $1
              AND fv.timeframe = $2
              AND fv.ts = $3
            LIMIT 1
        """, body.symbol, body.timeframe, ts)

        if row is None:
            results.append({"ts": ts, "prob": 1.0, "error": "no_data"})
            continue

        features = _get_features_from_row(row)
        X = np.array([[features[f] for f in model_features]], dtype=np.float64)
        if scaler is not None:
            X = scaler.transform(X)
        prob = float(model.predict_proba(X)[0, 1])
        results.append({"ts": ts, "prob": round(prob, 4)})

    return PredictBatchResponse(results=results)
