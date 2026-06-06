import argparse
import json
import os
import time
from itertools import product

import numpy as np
from sklearn.metrics import roc_auc_score

from config import HYPERPARAM_GRID, MODELS_DIR, VALIDATION_THRESHOLDS
from data import load_training_data
from model import train, save_artifact, make_scaler, compute_feature_importance


def grid_search(X_train, y_train, X_val, y_val):
    keys = list(HYPERPARAM_GRID.keys())
    values = list(HYPERPARAM_GRID.values())
    best_auc = -1
    best_params = None
    best_model = None
    total = 1
    for v in values:
        total *= len(v)
    print(f"Grid search: {total} combinations")

    for i, combo in enumerate(product(*values)):
        params = dict(zip(keys, combo))
        model = train(X_train, y_train, X_val, y_val, params)
        val_pred = model.predict_proba(X_val)[:, 1]
        auc = roc_auc_score(y_val, val_pred)
        if auc > best_auc:
            best_auc = auc
            best_params = params
            best_model = model
        if (i + 1) % 10 == 0 or i == total - 1:
            print(f"  [{i+1}/{total}] current best AUC={best_auc:.4f}")

    print(f"Best params: {best_params}  (val AUC={best_auc:.4f})")
    return best_model, best_params, best_auc


def compute_metrics(model, X, y):
    pred = model.predict_proba(X)[:, 1]
    auc = roc_auc_score(y, pred)
    return {"auc": round(float(auc), 4)}, pred


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--horizon", type=int, default=4, choices=[4, 12, 24])
    parser.add_argument("--symbol", type=str, default="BTCUSDT")
    parser.add_argument("--timeframe", type=str, default="15m")
    parser.add_argument("--version", type=str, default="v1.0")
    args = parser.parse_args()

    print(f"=== XGBoost Meta Model Training (horizon={args.horizon}) ===")

    print("\nLoading training data...")
    data = load_training_data(args.horizon, args.symbol, args.timeframe)
    X_train, X_val, X_test, y_train, y_val, y_test = data[:6]

    print("\nCreating scaler (if enabled)...")
    scaler = make_scaler(X_train)

    print("\nRunning grid search...")
    t0 = time.time()
    best_model, best_params, best_val_auc = grid_search(X_train, y_train, X_val, y_val)
    elapsed = time.time() - t0
    print(f"Grid search completed in {elapsed:.1f}s")

    print("\nEvaluating on test set...")
    test_metrics, _ = compute_metrics(best_model, X_test, y_test)
    print(f"Test AUC: {test_metrics['auc']:.4f}")

    all_metrics = {
        "val_auc": round(best_val_auc, 4),
        "test_auc": test_metrics["auc"],
        "best_params": best_params,
        "train_samples": len(X_train),
        "val_samples": len(X_val),
        "test_samples": len(X_test),
    }

    print("\nComputing feature importance...")
    importance = compute_feature_importance(best_model)
    print("Feature importance (gain):")
    for feat, imp in importance.items():
        print(f"  {feat}: {imp:.4f}")
    all_metrics["feature_importance"] = importance

    version_str = f"{args.version}_{args.horizon}bar"
    model_path = os.path.join(MODELS_DIR, f"xgb_{version_str}.pkl")
    save_artifact(best_model, model_path, args.horizon, version_str, all_metrics, scaler, importance)
    print(f"\nModel saved: {model_path}")

    test_auc = test_metrics["auc"]
    min_auc = VALIDATION_THRESHOLDS["test_auc_min"]
    if test_auc < min_auc:
        print(f"\n⚠ WARNING: Test AUC {test_auc:.4f} < minimum {min_auc}. Consider revising features.")
    else:
        print(f"\n✓ Test AUC {test_auc:.4f} meets threshold {min_auc}.")

    report_path = os.path.join(MODELS_DIR, f"report_{version_str}.json")
    with open(report_path, "w") as f:
        json.dump(all_metrics, f, indent=2, default=str)
    print(f"Report saved: {report_path}")


if __name__ == "__main__":
    main()
