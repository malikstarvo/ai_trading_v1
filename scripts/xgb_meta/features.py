import math
import numpy as np

TECH_WEIGHT = 0.40
OF_WEIGHT = 0.40
REGIME_WEIGHT = 0.20


def _valid(v):
    if v is None:
        return False
    if isinstance(v, float) and (math.isnan(v) or math.isinf(v)):
        return False
    return True


# ─── Technical Score ────────────────────────────────────────────────────────────

def compute_technical_score(price, ema20, ema50, ema200, rsi14, atr14, adx14, volume, vol_ema20):
    if not _valid(price):
        return 0.0

    trend = _calc_trend(price, ema20, ema50, ema200)
    momentum = _calc_momentum(rsi14)
    volume_score = _calc_volume(volume, vol_ema20)
    volatility = _calc_volatility(atr14, price)
    adx_bonus = _calc_adx_bonus(adx14)

    total = trend + momentum + volume_score + volatility + adx_bonus
    return max(0, min(100, total))


def _calc_trend(price, ema20, ema50, ema200):
    if not all(_valid(v) for v in [ema20, ema50, ema200]):
        return 0.0

    if price > ema20 and ema20 > ema50 and ema50 > ema200:
        base = 30
    elif price > ema20 and ema20 > ema50:
        base = 25
    elif price > ema20:
        base = 20
    elif price < ema20 and ema20 > ema50 and ema50 > ema200:
        base = 20
    elif price > ema50:
        base = 15
    elif price > ema200:
        base = 10
    else:
        base = 0

    alignment = 5 if (ema20 > ema50 and ema50 > ema200) else 0
    result = base + alignment
    return min(35, result)


def _calc_momentum(rsi):
    if not _valid(rsi):
        return 0.0

    if 45 <= rsi <= 65:
        return 22 + ((rsi - 45) / 20) * 8
    elif 30 <= rsi < 45:
        return 15 + ((rsi - 30) / 15) * 7
    elif 65 < rsi <= 75:
        return 22 - ((rsi - 65) / 10) * 7
    elif rsi > 75:
        val = 15 - ((rsi - 75) / 25) * 15
        return max(0, val)
    elif 0 <= rsi < 30:
        return 5 + (rsi / 30) * 10
    return 0.0


def _calc_volume(volume, vol_ema20):
    if not _valid(volume) or not _valid(vol_ema20) or vol_ema20 <= 0:
        return 0.0
    ratio = volume / vol_ema20
    score = math.log1p(ratio) * 12
    return max(0, min(20, score))


def _calc_volatility(atr, price):
    if not _valid(atr) or not _valid(price) or price <= 0 or atr <= 0:
        return 0.0
    atr_pct = atr / price * 100
    if atr_pct < 1.0:
        return 3
    elif atr_pct < 2.0:
        return 5
    elif atr_pct < 3.5:
        return 10
    elif atr_pct < 5.0:
        return 7
    else:
        return 4


def _calc_adx_bonus(adx):
    if not _valid(adx) or adx <= 0:
        return 0.0
    if adx >= 35:
        return 10
    elif adx >= 25:
        return 7
    elif adx >= 20:
        return 3
    else:
        return 0


# ─── OrderFlow Score ────────────────────────────────────────────────────────────

def compute_orderflow_score(funding_rate, oi_delta_pct, ls_ratio, long_liq_usd, short_liq_usd):
    funding_score = _calc_funding_score(funding_rate)
    oi_score = _calc_oi_score(oi_delta_pct)
    ls_score = _calc_ls_score(ls_ratio)
    liq_score = _calc_liq_score(long_liq_usd, short_liq_usd)
    total = funding_score + oi_score + ls_score + liq_score
    return max(0, min(100, total))


def _calc_funding_score(funding_rate):
    if not _valid(funding_rate):
        return 10.0
    bps = funding_rate * 10000
    if bps <= 0:
        score = 10 + bps * 3
        return max(0, score)
    elif bps < 2:
        return 10 + bps * 5
    elif bps < 5:
        score = 20 - (bps - 2) * 3.33
        return max(10, score)
    else:
        score = 10 - (bps - 5) * 2
        return max(0, score)


def _calc_oi_score(oi_delta_pct):
    if not _valid(oi_delta_pct):
        return 12.5
    score = 15 + oi_delta_pct * 10
    return max(0, min(25, score))


def _calc_ls_score(ls_ratio):
    if not _valid(ls_ratio) or ls_ratio <= 0:
        return 10.0
    if 1.0 <= ls_ratio <= 1.3:
        return 12 + (ls_ratio - 1.0) / 0.3 * 6
    elif 1.3 < ls_ratio <= 2.0:
        score = 18 - (ls_ratio - 1.3) / 0.7 * 8
        return max(10, score)
    elif ls_ratio > 2.0:
        score = 10 - (ls_ratio - 2.0) * 8
        return max(2, score)
    else:
        score = 12 - (1.0 - ls_ratio) * 20
        return max(0, score)


def _calc_liq_score(long_liq, short_liq):
    if not _valid(long_liq) or not _valid(short_liq):
        return 12.5
    total_liq = long_liq + short_liq
    if total_liq <= 0:
        return 12.5
    imbalance = (short_liq - long_liq) / total_liq
    norm = min(1, max(0, total_liq / 50_000_000))
    direction = imbalance * norm * 8
    magnitude = norm * 4.5
    score = 12.5 - direction + magnitude
    return max(0, min(25, score))


# ─── Regime Score ───────────────────────────────────────────────────────────────

def compute_regime_score(adx14, atr14, price, volatility14):
    has_adx = _valid(adx14)
    has_atr = _valid(atr14)
    has_price = _valid(price)
    has_vol = _valid(volatility14)

    if not has_adx and not has_atr and not has_price and not has_vol:
        return 50.0, "unknown"

    trend_score = _calc_trend_score(adx14)
    atr_vol = _calc_vol_from_atr(atr14, price)
    feat_vol = _calc_vol_from_feature(volatility14)
    vol_score = (atr_vol + feat_vol) / 2
    total = trend_score + vol_score
    total = max(0, min(100, total))
    regime = _classify_regime(adx14, vol_score)
    return total, regime


def _calc_trend_score(adx):
    if not _valid(adx):
        return 25.0
    if adx < 20:
        return 0.0
    if adx >= 35:
        return 50.0
    return (adx - 20) / 15 * 50


def _calc_vol_from_atr(atr, price):
    if not _valid(atr) or not _valid(price) or price <= 0 or atr <= 0:
        return 25.0
    atr_pct = atr / price * 100
    if atr_pct < 0.5:
        return 5.0
    elif atr_pct < 1.5:
        return 5 + (atr_pct - 0.5) / 1.0 * 20
    elif atr_pct < 3.0:
        return 25 + (atr_pct - 1.5) / 1.5 * 20
    elif atr_pct < 5.0:
        return 45 - (atr_pct - 3.0) / 2.0 * 20
    else:
        return 15.0


def _calc_vol_from_feature(vol):
    if not _valid(vol) or vol <= 0:
        return 25.0
    if vol < 0.3:
        return 5.0
    elif vol < 1.0:
        return 5 + (vol - 0.3) / 0.7 * 20
    elif vol < 2.5:
        return 25 + (vol - 1.0) / 1.5 * 20
    elif vol < 4.0:
        return 45 - (vol - 2.5) / 1.5 * 15
    else:
        return 20.0


def _classify_regime(adx, vol_score):
    if not _valid(adx):
        return "unknown"
    is_trending = adx >= 25
    is_high_vol = vol_score >= 25
    if is_trending and is_high_vol:
        return "trending_high_vol"
    elif is_trending and not is_high_vol:
        return "trending_low_vol"
    elif not is_trending and is_high_vol:
        return "ranging_high_vol"
    else:
        return "ranging_low_vol"


# ─── Confidence Score ───────────────────────────────────────────────────────────

def compute_confidence_score(technical_score, orderflow_score, regime_score):
    return (
        technical_score * TECH_WEIGHT
        + orderflow_score * OF_WEIGHT
        + regime_score * REGIME_WEIGHT
    )


# ─── Volume Delta ───────────────────────────────────────────────────────────────

def compute_volume_delta(volume, volume_ema20):
    if not _valid(volume) or not _valid(volume_ema20) or volume <= 0 or volume_ema20 <= 0:
        return 0.0
    ratio = volume / volume_ema20
    return math.log(ratio)


# ─── Full Pipeline ────────────────────────────────────────────────────────────

def compute_all_features(row):
    tech = compute_technical_score(
        price=row.get("close"),
        ema20=row.get("ema20"),
        ema50=row.get("ema50"),
        ema200=row.get("ema200"),
        rsi14=row.get("rsi14"),
        atr14=row.get("atr14"),
        adx14=row.get("adx14"),
        volume=row.get("volume"),
        vol_ema20=row.get("volume_ema20"),
    )
    of = compute_orderflow_score(
        funding_rate=row.get("funding_rate"),
        oi_delta_pct=row.get("oi_delta_1_pct"),
        ls_ratio=row.get("ls_ratio_raw"),
        long_liq_usd=row.get("liq_long_usd"),
        short_liq_usd=row.get("liq_short_usd"),
    )
    regime_score, regime_label = compute_regime_score(
        adx14=row.get("adx14"),
        atr14=row.get("atr14"),
        price=row.get("close"),
        volatility14=row.get("volatility_14"),
    )
    conf = compute_confidence_score(tech, of, regime_score)
    vol_delta = compute_volume_delta(
        volume=row.get("volume"),
        volume_ema20=row.get("volume_ema20"),
    )
    return {
        "technical_score": tech,
        "orderflow_score": of,
        "regime_score": regime_score,
        "confidence_score": conf,
        "atr14": row.get("atr14") if _valid(row.get("atr14")) else 0.0,
        "adx14": row.get("adx14") if _valid(row.get("adx14")) else 0.0,
        "funding_rate": row.get("funding_rate") if _valid(row.get("funding_rate")) else 0.0,
        "oi_delta_1_pct": row.get("oi_delta_1_pct") if _valid(row.get("oi_delta_1_pct")) else 0.0,
        "volume_delta": vol_delta,
    }
