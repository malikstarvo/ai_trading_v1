import argparse
import json
import time
import os
from itertools import product

import numpy as np
from sklearn.metrics import roc_auc_score

from config import HORIZONS, MODELS_DIR, HYPERPARAM_GRID
from data import load_training_data
from model import train, compute_feature_importance


VARIANTS = {
    "A_score_only": [
        "technical_score",
        "regime_score",
        "confidence_score",
    ],
    "B_raw_only": [
        "atr14",
        "adx14",
        "funding_rate",
        "volume_delta",
    ],
    "C_v1_current": [
        "technical_score",
        "regime_score",
        "confidence_score",
        "atr14",
        "adx14",
        "funding_rate",
        "volume_delta",
    ],
    "D_with_orderflow": [
        "technical_score",
        "orderflow_score",
        "regime_score",
        "confidence_score",
        "atr14",
        "adx14",
        "funding_rate",
        "oi_delta_1_pct",
        "volume_delta",
    ],
}


def grid_search(X_train, y_train, X_val, y_val):
    keys = list(HYPERPARAM_GRID.keys())
    values = list(HYPERPARAM_GRID.values())
    best_auc = -1
    best_params = None
    best_model = None
    for combo in product(*values):
        params = dict(zip(keys, combo))
        model = train(X_train, y_train, X_val, y_val, params)
        val_pred = model.predict_proba(X_val)[:, 1]
        auc = roc_auc_score(y_val, val_pred)
        if auc > best_auc:
            best_auc = auc
            best_params = params
            best_model = model
    return best_model, best_params, best_auc


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--symbol", type=str, default="BTCUSDT")
    parser.add_argument("--timeframe", type=str, default="15m")
    parser.add_argument("--output", type=str,
                        default=os.path.join(MODELS_DIR, "ablation_report.json"))
    args = parser.parse_args()

    results = {"config": {"symbol": args.symbol, "timeframe": args.timeframe}}

    for variant_name, feature_set in VARIANTS.items():
        print(f"\n{'='*60}")
        print(f"  Variant: {variant_name} ({len(feature_set)} features)")
        print(f"{'='*60}")
        print(f"  Features: {feature_set}")

        results[variant_name] = {}

        for horizon in HORIZONS:
            print(f"\n  --- Horizon {horizon}-bar ---")
            try:
                data = load_training_data(horizon, args.symbol, args.timeframe,
                                          feature_columns=feature_set)
                X_train, X_val, X_test, y_train, y_val, y_test = data[:6]

                t0 = time.time()
                model, best_params, val_auc = grid_search(X_train, y_train, X_val, y_val)
                elapsed = time.time() - t0

                test_pred = model.predict_proba(X_test)[:, 1]
                test_auc = roc_auc_score(y_test, test_pred)

                importance = compute_feature_importance(model)

                print(f"    val AUC={val_auc:.4f}  test AUC={test_auc:.4f}  "
                      f"train={len(X_train)} val={len(X_val)} test={len(X_test)}  "
                      f"({elapsed:.1f}s)")

                results[variant_name][f"horizon_{horizon}"] = {
                    "val_auc": round(val_auc, 4),
                    "test_auc": round(test_auc, 4),
                    "best_params": best_params,
                    "feature_importance": importance,
                    "train_samples": len(X_train),
                    "val_samples": len(X_val),
                    "test_samples": len(X_test),
                }
            except Exception as e:
                print(f"    ERROR: {e}")
                results[variant_name][f"horizon_{horizon}"] = {"error": str(e)}

    # Summary table
    print(f"\n{'='*60}")
    print(f"  ABLATION SUMMARY — Test AUC")
    print(f"{'='*60}")
    print(f"  {'Variant':<25s}", end="")
    for h in HORIZONS:
        print(f"  {h:>8d}b", end="")
    hdr_sep = "-" * 25
    col_sep = "-" * 8
    print()
    print(f"  {hdr_sep:<25s}", end="")
    for _ in HORIZONS:
        print(f"  {col_sep:<8s}", end="")
    print()
    for variant_name in VARIANTS:
        print(f"  {variant_name:<25s}", end="")
        for h in HORIZONS:
            entry = results.get(variant_name, {}).get(f"horizon_{h}", {})
            auc = entry.get("test_auc", "ERR")
            flag = " *" if isinstance(auc, (int, float)) and auc >= 0.55 else ""
            print(f"  {auc:>8.4f}{flag}" if isinstance(auc, (int, float)) else f"  {auc:>8s}", end="")
        print()
    print(f"\n  * AUC >= 0.55 (potential signal)")

    with open(args.output, "w") as f:
        json.dump(results, f, indent=2, default=str)
    print(f"\n  Report saved: {args.output}")


if __name__ == "__main__":
    main()
