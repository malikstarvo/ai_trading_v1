import subprocess
from fastapi import APIRouter

router = APIRouter()

@router.post("/api/actions/feature-backfill")
async def trigger_feature_backfill():
    try:
        subprocess.Popen(
            ["bash", "-c", "cd /home/ubuntu/ai_trading_v1 && go run ./cmd/feature-backfill/"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        return {"status": "started", "action": "feature-backfill"}
    except Exception as e:
        return {"status": "error", "message": str(e)}

@router.post("/api/actions/edge-study")
async def trigger_edge_study():
    try:
        subprocess.Popen(
            ["bash", "-c", "cd /home/ubuntu/ai_trading_v1 && go run ./cmd/edge-study/"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        return {"status": "started", "action": "edge-study"}
    except Exception as e:
        return {"status": "error", "message": str(e)}

@router.post("/api/actions/model-training")
async def trigger_model_training():
    try:
        subprocess.Popen(
            ["bash", "-c", "cd /home/ubuntu/ai_trading_v1/scripts/xgb_meta && bash run_all.sh"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        return {"status": "started", "action": "model-training"}
    except Exception as e:
        return {"status": "error", "message": str(e)}
