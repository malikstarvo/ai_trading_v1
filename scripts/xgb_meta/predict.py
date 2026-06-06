import argparse
import os
import glob
import numpy as np
import pandas as pd
import psycopg2

from config import DB_CONFIG, FEATURE_COLUMNS, MODELS_DIR
from features import compute_all_features
from model import load_artifact


def get_latest_model_path(horizon=4):
    pattern = os.path.join(MODELS_DIR, f"xgb_*_{horizon}bar.pkl")
    files = glob.glob(pattern)
    if not files:
        raise FileNotFoundError(f"No model found for horizon={horizon} in {MODELS_DIR}")
    return max(files, key=os.path.getmtime)


def load_row_from_db(ts, symbol="BTCUSDT", timeframe="15m"):
    conn = psycopg2.connect(**DB_CONFIG)
    try:
        cur = conn.cursor()
        cur.execute("""
            SELECT
                fv.symbol, fv.timeframe, fv.ts,
                c.close, c.volume,
                fv.ema20, fv.ema50, fv.ema200, fv.rsi14, fv.atr14, fv.adx14,
                fv.volume_ema20,
                fv.oi_delta_1_pct, fv.oi_delta_4_pct, fv.oi_delta_12_pct,
                fv.funding_rate, fv.ls_ratio_raw,
                fv.liq_long_usd, fv.liq_short_usd,
                fv.volatility_14
            FROM feature_values fv
            JOIN candles c ON fv.symbol = c.symbol AND fv.timeframe = c.timeframe AND fv.ts = c.time
            WHERE fv.symbol = %s AND fv.timeframe = %s AND fv.ts = %s
        """, (symbol, timeframe, ts))
        row = cur.fetchone()
        if row is None:
            raise ValueError(f"No data found for ts={ts}")
        colnames = [desc[0] for desc in cur.description]
        return dict(zip(colnames, row))
    finally:
        conn.close()


def predict(ts, horizon=4, symbol="BTCUSDT", timeframe="15m", model_path=None):
    if model_path is None:
        model_path = get_latest_model_path(horizon)

    artifact = load_artifact(model_path)
    model = artifact["model"]
    model_features = artifact.get("features", FEATURE_COLUMNS)
    scaler = artifact.get("scaler")

    if isinstance(ts, str):
        ts = pd.Timestamp(ts).to_pydatetime()

    raw = load_row_from_db(ts, symbol, timeframe)
    features = compute_all_features(raw)
    X = np.array([[features[f] for f in model_features]], dtype=np.float64)

    if scaler is not None:
        X = scaler.transform(X)

    prob = float(model.predict_proba(X)[0, 1])
    return prob, features, artifact


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--ts", type=str, required=True)
    parser.add_argument("--horizon", type=int, default=4, choices=[4, 12, 24])
    parser.add_argument("--symbol", type=str, default="BTCUSDT")
    parser.add_argument("--timeframe", type=str, default="15m")
    parser.add_argument("--model", type=str, default=None)
    args = parser.parse_args()

    prob, features, artifact = predict(
        args.ts, args.horizon, args.symbol, args.timeframe, args.model
    )

    print(f"Model:       {artifact.get('version', '?')} (horizon={artifact.get('horizon', '?')})")
    print(f"Timestamp:   {args.ts}")
    print(f"Symbol:      {args.symbol} {args.timeframe}")
    print(f"Probability: {prob:.4f}")
    print(f"Decision:    {'TRADE' if prob >= 0.45 else 'NO TRADE'} (threshold: 0.45)")
    print(f"\nFeatures:")
    for k, v in features.items():
        print(f"  {k:20s} = {v:.4f}")


if __name__ == "__main__":
    main()
