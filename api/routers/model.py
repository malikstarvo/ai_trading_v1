import os
import glob
import json
from fastapi import APIRouter

router = APIRouter()

MODELS_DIR = "scripts/xgb_meta/models"


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
