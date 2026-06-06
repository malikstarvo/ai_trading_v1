#!/usr/bin/env python3
"""
Parity Test: Python vs Go scoring implementations.

Verifies that the Python re-implementation of Technical, OrderFlow,
and Regime scorers produces identical outputs (±0.01 tolerance) as the
canonical Go implementations.

This is a BLOCKING GATE before training. Run:
    python parity_test.py

Expected: ALL PASS
"""

import math
import sys

from features import (
    compute_technical_score,
    compute_orderflow_score,
    compute_regime_score,
    compute_confidence_score,
)

TOLERANCE = 0.01


def check(name, got, expected, label=""):
    label_str = f" ({label})" if label else ""
    if abs(got - expected) <= TOLERANCE:
        print(f"  [PASS] {name}{label_str}: {got:.6f} ~ {expected}")
        return True
    else:
        print(f"  [FAIL] {name}{label_str}: {got:.6f} != {expected} (diff={abs(got-expected):.6f})")
        return False


def test_bullish():
    print("\n[Test: bullish]")
    price = 51000.0
    ok = True

    tech = compute_technical_score(
        price=price, ema20=price*0.99, ema50=price*0.97, ema200=price*0.95,
        rsi14=58, atr14=price*0.012, adx14=32,
        volume=1_100_000, vol_ema20=1_000_000,
    )
    ok &= check("tech", tech, 83.10324813675253)

    of = compute_orderflow_score(
        funding_rate=0.00002, oi_delta_pct=5.0, ls_ratio=1.2,
        long_liq_usd=5_000_000, short_liq_usd=1_000_000,
    )
    ok &= check("orderflow", of, 65.68)

    regime, label = compute_regime_score(
        adx14=32, atr14=price*0.012, price=price, volatility14=1.8,
    )
    ok &= check("regime", regime, 67.33333333333333)
    ok &= check("regime_label", 1.0 if label == "trending_high_vol" else 0.0, 1.0, label)

    conf = compute_confidence_score(tech, of, regime)
    ok &= check("confidence", conf, 72.97996592136768)

    return ok


def test_ranging():
    print("\n[Test: ranging]")
    price = 50000.0
    ok = True

    tech = compute_technical_score(
        price=price, ema20=price*0.998, ema50=price*0.997, ema200=price*0.996,
        rsi14=50, atr14=price*0.003, adx14=16,
        volume=1_000_000, vol_ema20=1_000_000,
    )
    ok &= check("tech", tech, 70.31776616671934)

    of = compute_orderflow_score(
        funding_rate=0.00001, oi_delta_pct=0.1, ls_ratio=1.0,
        long_liq_usd=100_000, short_liq_usd=100_000,
    )
    ok &= check("orderflow", of, 51.018)

    regime, label = compute_regime_score(
        adx14=16, atr14=price*0.003, price=price, volatility14=0.4,
    )
    ok &= check("regime", regime, 6.428571428571429)
    ok &= check("regime_label", 1.0 if label == "ranging_low_vol" else 0.0, 1.0, label)

    conf = compute_confidence_score(tech, of, regime)
    ok &= check("confidence", conf, 49.82002075240202)

    return ok


def test_chaotic():
    print("\n[Test: chaotic]")
    price = 50500.0
    ok = True

    tech = compute_technical_score(
        price=price, ema20=price*0.98, ema50=price*0.97, ema200=price*0.96,
        rsi14=45, atr14=price*0.04, adx14=18,
        volume=1_500_000, vol_ema20=1_000_000,
    )
    ok &= check("tech", tech, 74.99548878248986)

    of = compute_orderflow_score(
        funding_rate=-0.0002, oi_delta_pct=-0.5, ls_ratio=0.8,
        long_liq_usd=2_000_000, short_liq_usd=3_000_000,
    )
    ok &= check("orderflow", of, 34.79)

    regime, label = compute_regime_score(
        adx14=18, atr14=price*0.04, price=price, volatility14=4.0,
    )
    ok &= check("regime", regime, 27.5)
    ok &= check("regime_label", 1.0 if label == "ranging_high_vol" else 0.0, 1.0, label)

    conf = compute_confidence_score(tech, of, regime)
    ok &= check("confidence", conf, 49.414195512995946)

    return ok


def test_neutral():
    print("\n[Test: neutral]")
    price = 50000.0
    ok = True

    tech = compute_technical_score(
        price=price, ema20=price*0.999, ema50=price*0.998, ema200=price*0.997,
        rsi14=50, atr14=price*0.005, adx14=20,
        volume=500_000, vol_ema20=500_000,
    )
    ok &= check("tech", tech, 73.31776616671934)

    of = compute_orderflow_score(
        funding_rate=0.00001, oi_delta_pct=0.0, ls_ratio=1.0,
        long_liq_usd=100_000, short_liq_usd=100_000,
    )
    ok &= check("orderflow", of, 50.018)

    regime, label = compute_regime_score(
        adx14=20, atr14=price*0.005, price=price, volatility14=0.6,
    )
    ok &= check("regime", regime, 9.285714285714286)
    ok &= check("regime_label", 1.0 if label == "ranging_low_vol" else 0.0, 1.0, label)

    conf = compute_confidence_score(tech, of, regime)
    ok &= check("confidence", conf, 51.19144932383059)

    return ok


def test_nan():
    print("\n[Test: nan inputs]")
    ok = True

    tech = compute_technical_score(
        price=51000.0,
        ema20=math.nan, ema50=math.nan, ema200=math.nan,
        rsi14=math.nan, atr14=0, adx14=0,
        volume=0, vol_ema20=0,
    )
    ok &= check("tech", tech, 0.0)

    # Go: only FundingRate=NaN, others default to 0
    of = compute_orderflow_score(
        funding_rate=math.nan, oi_delta_pct=0.0,
        ls_ratio=0.0, long_liq_usd=0.0, short_liq_usd=0.0,
    )
    ok &= check("orderflow", of, 47.5)

    # Go: regime.Input{} (all zero)
    regime, label = compute_regime_score(
        adx14=0, atr14=0, price=0, volatility14=0,
    )
    ok &= check("regime", regime, 25.0)
    ok &= check("regime_label", 1.0 if label == "ranging_high_vol" else 0.0, 1.0, label)

    conf = compute_confidence_score(tech, of, regime)
    ok &= check("confidence", conf, 24.0)

    return ok


def main():
    print("=" * 60)
    print("Parity Test: Python vs Go Scoring")
    print("=" * 60)
    print(f"Tolerance: ±{TOLERANCE}")

    results = [
        ("bullish", test_bullish()),
        ("ranging", test_ranging()),
        ("chaotic", test_chaotic()),
        ("neutral", test_neutral()),
        ("nan", test_nan()),
    ]

    print("\n" + "=" * 60)
    passed = sum(1 for _, r in results if r)
    failed = sum(1 for _, r in results if not r)
    print(f"Passed: {passed}/{len(results)}  Failed: {failed}/{len(results)}")

    if failed > 0:
        print("\nFAILED - fix scoring implementation before training")
        sys.exit(1)
    else:
        print("\nALL PASSED - parity confirmed. Ready for training.")


if __name__ == "__main__":
    main()
