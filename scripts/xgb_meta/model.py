import joblib
import numpy as np
import xgboost as xgb
from sklearn.preprocessing import RobustScaler

from config import FEATURE_COLUMNS, XGB_FIXED_PARAMS, USE_SCALER


def train(X_train, y_train, X_val, y_val, params):
    model = xgb.XGBClassifier(**{**XGB_FIXED_PARAMS, **params})
    model.fit(
        X_train, y_train,
        eval_set=[(X_val, y_val)],
        verbose=False,
    )
    return model


def predict_proba(model, X, scaler=None):
    if scaler is not None:
        X = scaler.transform(X)
    return model.predict_proba(X)[:, 1]


def save_artifact(model, path, horizon, version, metrics, scaler=None, feature_importance=None):
    artifact = {
        "model": model,
        "features": FEATURE_COLUMNS,
        "horizon": horizon,
        "version": version,
        "metrics": metrics,
        "feature_importance": feature_importance,
        "scaler": scaler,
    }
    joblib.dump(artifact, path)


def load_artifact(path):
    return joblib.load(path)


def make_scaler(X_train):
    if not USE_SCALER:
        return None
    scaler = RobustScaler()
    scaler.fit(X_train)
    return scaler


def compute_feature_importance(model):
    if not hasattr(model, "get_booster"):
        return {}
    importance = model.get_booster().get_score(importance_type="gain")
    total = sum(importance.values())
    if total <= 0:
        return {}
    # Map f0→fN back to real feature names
    mapping = {}
    if hasattr(model, "feature_names_in_"):
        mapping = {f"f{i}": n for i, n in enumerate(model.feature_names_in_)}
    named = {}
    for feat, val in importance.items():
        name = mapping.get(feat, feat)
        named[name] = round(val / total, 4)
    return dict(sorted(named.items(), key=lambda x: x[1], reverse=True))
