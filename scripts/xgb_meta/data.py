import psycopg2
import pandas as pd
import numpy as np

from config import DB_CONFIG, FEATURE_COLUMNS, HORIZON_COLUMNS, TRAIN_RATIO, VAL_RATIO
from features import compute_all_features


def get_conn():
    return psycopg2.connect(**DB_CONFIG)


def load_raw_data(conn, symbol="BTCUSDT", timeframe="15m", feature_set_id=1):
    query = """
    SELECT
        fv.symbol,
        fv.timeframe,
        fv.ts,
        c.close,
        c.volume,
        fv.ema20,
        fv.ema50,
        fv.ema200,
        fv.rsi14,
        fv.atr14,
        fv.adx14,
        fv.volume_ema20,
        fv.oi_delta_1_pct,
        fv.oi_delta_4_pct,
        fv.oi_delta_12_pct,
        fv.funding_rate,
        fv.ls_ratio_raw,
        fv.liq_long_usd,
        fv.liq_short_usd,
        fv.volatility_14,
        tl.success_4,
        tl.success_12,
        tl.success_24,
        tl.future_return_4,
        tl.future_return_12,
        tl.future_return_24
    FROM feature_values fv
    JOIN training_labels tl
        ON fv.symbol = tl.symbol
        AND fv.timeframe = tl.timeframe
        AND fv.ts = tl.ts
        AND fv.feature_set_id = tl.feature_set_id
    JOIN candles c
        ON fv.symbol = c.symbol
        AND fv.timeframe = c.timeframe
        AND fv.ts = c.time
    WHERE fv.symbol = %s
      AND fv.timeframe = %s
      AND fv.feature_set_id = %s
    ORDER BY fv.ts ASC
    """
    df = pd.read_sql(query, conn, params=(symbol, timeframe, feature_set_id))

    df.replace([np.inf, -np.inf], np.nan, inplace=True)
    return df


def _median_impute(df, cols):
    for c in cols:
        med = df[c].median()
        if pd.isna(med):
            med = 0.0
        df[c] = df[c].fillna(med)
    return df


def prepare_training_data(df, horizon, feature_columns=None):
    success_col, _ = HORIZON_COLUMNS[horizon]
    cols = feature_columns if feature_columns is not None else FEATURE_COLUMNS

    feature_rows = []
    labels = []
    timestamps = []

    for _, row in df.iterrows():
        ts = row.get("ts")
        label = row.get(success_col)

        if pd.isna(label):
            continue

        features = compute_all_features(row.to_dict())
        feature_rows.append([features[f] for f in cols])
        labels.append(int(label))
        timestamps.append(ts)

    X = np.array(feature_rows, dtype=np.float64)
    y = np.array(labels, dtype=np.int32)

    X = _median_impute(pd.DataFrame(X, columns=cols), cols)

    return X, y, np.array(timestamps)


def chronological_split(X, y, timestamps):
    n = len(X)
    train_end = int(n * TRAIN_RATIO)
    val_end = train_end + int(n * VAL_RATIO)

    indices = np.arange(n)
    train_idx = indices[:train_end]
    val_idx = indices[train_end:val_end]
    test_idx = indices[val_end:]

    return (
        X.iloc[train_idx].reset_index(drop=True),
        X.iloc[val_idx].reset_index(drop=True),
        X.iloc[test_idx].reset_index(drop=True),
        y[train_idx], y[val_idx], y[test_idx],
        timestamps[train_idx], timestamps[val_idx], timestamps[test_idx],
    )


def load_training_data(horizon, symbol="BTCUSDT", timeframe="15m", feature_columns=None):
    conn = get_conn()
    try:
        df = load_raw_data(conn, symbol, timeframe)
    finally:
        conn.close()

    print(f"Loaded {len(df)} raw rows from DB")

    # Label quality diagnostics
    _, ret_col = HORIZON_COLUMNS[horizon]
    rets = df[ret_col].dropna().values * 100  # convert to pct
    if len(rets) > 0:
        print(f"  {ret_col} percentiles: p25={np.percentile(rets, 25):.3f}% "
              f"p50={np.percentile(rets, 50):.3f}% p75={np.percentile(rets, 75):.3f}%")
        for thresh_pct in [0.25, 0.5, 1.0]:
            wr = (rets >= thresh_pct).mean() * 100
            print(f"  win_rate @ ≥{thresh_pct}%: {wr:.1f}%")

    X, y, timestamps = prepare_training_data(df, horizon, feature_columns)
    print(f"After prepare: {len(X)} samples, {y.sum()} positive ({y.mean()*100:.1f}%)")

    splits = chronological_split(X, y, timestamps)
    X_train, X_val, X_test, y_train, y_val, y_test, t_train, t_val, t_test = splits

    print(f"Train: {len(X_train)} | Val: {len(X_val)} | Test: {len(X_test)}")
    print(f"Train pos ratio: {y_train.mean()*100:.1f}%")
    print(f"Val pos ratio:   {y_val.mean()*100:.1f}%")
    print(f"Test pos ratio:  {y_test.mean()*100:.1f}%")
    print(f"Train date range: {t_train[0]} to {t_train[-1]}")
    print(f"Test date range:  {t_test[0]} to {t_test[-1]}")

    return (X_train, X_val, X_test, y_train, y_val, y_test,
            t_train, t_val, t_test)
