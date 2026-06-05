package validate_test

import (
	"math"
	"testing"
	"time"

	"github.com/avav/ai_trading_v1/internal/collector/validate"
	"github.com/avav/ai_trading_v1/internal/model"
)

func newValidator() *validate.DataValidator {
	return validate.New(nil)
}

func TestValidateCandle_Valid(t *testing.T) {
	v := newValidator()
	c := &model.Candle{
		Time:   time.Now(),
		Symbol: "BTCUSDT",
		Open:   50000,
		High:   51000,
		Low:    49000,
		Close:  50500,
		Volume: 100,
	}
	err := v.ValidateCandle(c)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateCandle_HighLessThanLow(t *testing.T) {
	v := newValidator()
	c := &model.Candle{
		Time:   time.Now(),
		Open:   50000,
		High:   49000,
		Low:    51000,
		Close:  50500,
		Volume: 100,
	}
	err := v.ValidateCandle(c)
	if err == nil {
		t.Fatal("expected error for high < low")
	}
}

func TestValidateCandle_NegativeVolume(t *testing.T) {
	v := newValidator()
	c := &model.Candle{
		Time:   time.Now(),
		Open:   50000,
		High:   51000,
		Low:    49000,
		Close:  50500,
		Volume: -1,
	}
	err := v.ValidateCandle(c)
	if err == nil {
		t.Fatal("expected error for negative volume")
	}
}

func TestValidateCandle_ZeroTimestamp(t *testing.T) {
	v := newValidator()
	c := &model.Candle{
		Open:   50000,
		High:   51000,
		Low:    49000,
		Close:  50500,
		Volume: 100,
	}
	err := v.ValidateCandle(c)
	if err == nil {
		t.Fatal("expected error for zero timestamp")
	}
}

func TestValidateOI_Valid(t *testing.T) {
	v := newValidator()
	oi := &model.OIRecord{
		Time:       time.Now(),
		Symbol:     "BTCUSDT",
		OI:         100,
		OIValueUSD: 5000000,
	}
	err := v.ValidateOI(oi)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateOI_NotPositive(t *testing.T) {
	v := newValidator()
	oi := &model.OIRecord{
		Time:   time.Now(),
		Symbol: "BTCUSDT",
		OI:     0,
	}
	err := v.ValidateOI(oi)
	if err == nil {
		t.Fatal("expected error for zero OI")
	}
}

func TestValidateFunding_NaN(t *testing.T) {
	v := newValidator()
	f := &model.FundingRate{
		Time:   time.Now(),
		Symbol: "BTCUSDT",
		Rate:   math.NaN(),
	}
	err := v.ValidateFunding(f)
	if err == nil {
		t.Fatal("expected error for NaN rate")
	}
}

func TestValidateFunding_Inf(t *testing.T) {
	v := newValidator()
	f := &model.FundingRate{
		Time:   time.Now(),
		Symbol: "BTCUSDT",
		Rate:   math.Inf(1),
	}
	err := v.ValidateFunding(f)
	if err == nil {
		t.Fatal("expected error for Inf rate")
	}
}

func TestValidateLSRatio_Valid(t *testing.T) {
	v := newValidator()
	ls := &model.LSRatio{
		Time:      time.Now(),
		Symbol:    "BTCUSDT",
		Period:    "15m",
		BuyRatio:  0.5,
		SellRatio: 0.5,
	}
	err := v.ValidateLSRatio(ls)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateLiquidation_Valid(t *testing.T) {
	v := newValidator()
	liq := &model.Liquidation{
		Time:     time.Now(),
		Symbol:   "BTCUSDT",
		Side:     "Buy",
		Size:     10,
		Price:    50000,
		ValueUSD: 500000,
	}
	err := v.ValidateLiquidation(liq)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateLiquidation_InvalidSide(t *testing.T) {
	v := newValidator()
	liq := &model.Liquidation{
		Time:     time.Now(),
		Symbol:   "BTCUSDT",
		Side:     "Invalid",
		Size:     10,
		Price:    50000,
		ValueUSD: 500000,
	}
	err := v.ValidateLiquidation(liq)
	if err == nil {
		t.Fatal("expected error for invalid side")
	}
}
