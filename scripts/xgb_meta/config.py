import os

DB_CONFIG = {
    "host": os.environ.get("DB_HOST", "localhost"),
    "port": int(os.environ.get("DB_PORT", "5432")),
    "user": os.environ.get("DB_USER", "trader"),
    "password": os.environ.get("DB_PASSWORD", "trader_pass"),
    "dbname": os.environ.get("DB_NAME", "ai_trading"),
}

# V1.1: agent scores dropped — ablation showed raw features outperform.
# orderflow also dropped (OI/LS >80% NaN on server).
# TODO: re-enable after 3+ months of data collection.
FEATURE_COLUMNS = [
    "close",
    "atr14",
    "adx14",
    "funding_rate",
    "volume_delta",
    "volatility_14",
]

HORIZONS = [4, 12, 24]

HORIZON_COLUMNS = {
    4: ("success_4", "future_return_4"),
    12: ("success_12", "future_return_12"),
    24: ("success_24", "future_return_24"),
}

USE_SCALER = False

TRAIN_RATIO = 0.6
VAL_RATIO = 0.2
TEST_RATIO = 0.2

HYPERPARAM_GRID = {
    "max_depth": [3, 5],
    "learning_rate": [0.03, 0.05, 0.1],
    "n_estimators": [200, 500],
    "subsample": [0.8, 1.0],
    "colsample_bytree": [0.8, 1.0],
}

XGB_FIXED_PARAMS = {
    "objective": "binary:logistic",
    "eval_metric": "auc",
    "random_state": 42,
    "early_stopping_rounds": 50,
    "verbosity": 0,
}

VALIDATION_THRESHOLDS = {
    "test_auc_min": 0.60,
    "profit_factor_min": 1.1,
}

MODELS_DIR = os.path.join(os.path.dirname(__file__), "models")
os.makedirs(MODELS_DIR, exist_ok=True)
