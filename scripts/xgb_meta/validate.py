import argparse
import json

import numpy as np
from sklearn.metrics import (
    roc_auc_score,
    precision_score,
    recall_score,
    f1_score,
    brier_score_loss,
)

from config import MODELS_DIR
from data import load_training_data
from model import load_artifact, compute_feature_importance


def top_k_precision(y_true, y_pred_prob, k=0.1):
    n_top = max(1, int(len(y_true) * k))
    top_idx = np.argsort(y_pred_prob)[-n_top:]
    return float(y_true[top_idx].mean())


def profit_factor_simulation(y_true, y_pred_prob, threshold=0.45):
    trades = y_pred_prob >= threshold
    if trades.sum() == 0:
        return 0.0
    winners = (y_true[trades] == 1).sum()
    losers = (y_true[trades] == 0).sum()
    if losers == 0:
        return float("inf")
    return winners / losers


def calibration_report(y_true, y_pred_prob, bins=10):
    indices = np.digitize(y_pred_prob, np.linspace(0, 1, bins + 1)) - 1
    report = []
    for b in range(bins):
        mask = indices == b
        if mask.sum() == 0:
            continue
        mean_pred = y_pred_prob[mask].mean()
        mean_obs = y_true[mask].mean()
        report.append({
            "bin": b,
            "n": int(mask.sum()),
            "mean_pred": round(float(mean_pred), 3),
            "mean_obs": round(float(mean_obs), 3),
        })
    return report


def evaluate_model(model, X, y, name="", threshold=0.45):
    pred_prob = model.predict_proba(X)[:, 1]
    pred_label = (pred_prob >= threshold).astype(int)

    metrics = {
        "set": name,
        "n": len(y),
        "pos_rate": float(y.mean()),
    }

    metrics["auc"] = round(float(roc_auc_score(y, pred_prob)), 4)
    metrics["precision"] = round(float(precision_score(y, pred_label, zero_division=0)), 4)
    metrics["recall"] = round(float(recall_score(y, pred_label, zero_division=0)), 4)
    metrics["f1"] = round(float(f1_score(y, pred_label, zero_division=0)), 4)
    metrics["brier"] = round(float(brier_score_loss(y, pred_prob)), 4)
    metrics["profit_factor"] = round(float(profit_factor_simulation(y, pred_prob, threshold)), 4)
    metrics["top10p_precision"] = round(float(top_k_precision(y, pred_prob, 0.10)), 4)

    pred_prob_precise = np.clip(pred_prob, 0.001, 0.999)
    log_loss = -(y * np.log(pred_prob_precise) + (1 - y) * np.log(1 - pred_prob_precise)).mean()
    metrics["log_loss"] = round(float(log_loss), 4)

    return metrics, pred_prob


def print_report(metrics, importance):
    print(f"\n{'='*60}")
    print(f"Validation Report: {metrics['set']}")
    print(f"{'='*60}")
    print(f"  Samples:          {metrics['n']}")
    print(f"  Positive rate:    {metrics['pos_rate']*100:.1f}%")
    print(f"  AUC:              {metrics['auc']:.4f}")
    print(f"  Precision (0.45): {metrics['precision']:.4f}")
    print(f"  Recall (0.45):    {metrics['recall']:.4f}")
    print(f"  F1 (0.45):        {metrics['f1']:.4f}")
    print(f"  Top 10% prec:     {metrics['top10p_precision']:.4f}")
    print(f"  Profit Factor:    {metrics['profit_factor']:.4f}")
    print(f"  Brier score:      {metrics['brier']:.4f}")
    print(f"  Log loss:         {metrics['log_loss']:.4f}")

    if importance:
        print(f"\n  Feature Importance (gain):")
        for feat, imp in importance.items():
            bar = "█" * int(imp * 50)
            print(f"    {feat:20s} {imp:.4f} {bar}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", type=str, required=True)
    parser.add_argument("--horizon", type=int, default=4, choices=[4, 12, 24])
    parser.add_argument("--symbol", type=str, default="BTCUSDT")
    parser.add_argument("--timeframe", type=str, default="15m")
    args = parser.parse_args()

    print(f"Loading model: {args.model}")
    artifact = load_artifact(args.model)
    model = artifact["model"]
    print(f"  Version: {artifact.get('version', '?')}")
    print(f"  Horizon: {artifact.get('horizon', '?')}")

    print("\nLoading test data...")
    data = load_training_data(args.horizon, args.symbol, args.timeframe)
    _, _, X_test, _, _, y_test, _, _, t_test = data

    print("\n=== Full Evaluation ===")
    metrics, pred_prob = evaluate_model(model, X_test, y_test, name="Test")

    importance = artifact.get("feature_importance") or compute_feature_importance(model)
    print_report(metrics, importance)

    cal = calibration_report(y_test, pred_prob)
    print(f"\n  Calibration (10 bins):")
    for c in cal:
        marker = "✓" if abs(c["mean_pred"] - c["mean_obs"]) < 0.1 else "⚠"
        print(f"    bin={c['bin']} n={c['n']:4d} pred={c['mean_pred']:.3f} obs={c['mean_obs']:.3f} {marker}")

    report_path = f"{MODELS_DIR}/validation_{artifact.get('version', 'model')}.json"
    with open(report_path, "w") as f:
        json.dump({"metrics": metrics, "calibration": cal, "feature_importance": importance}, f, indent=2)
    print(f"\nReport saved: {report_path}")


if __name__ == "__main__":
    main()
