package validate

import (
	"fmt"
	"math"

	"github.com/avav/ai_trading_v1/internal/collector/metrics"
	"github.com/avav/ai_trading_v1/internal/model"
)

type Validator interface {
	ValidateCandle(c *model.Candle) error
	ValidateOI(o *model.OIRecord) error
	ValidateFunding(f *model.FundingRate) error
	ValidateLSRatio(l *model.LSRatio) error
	ValidateLiquidation(l *model.Liquidation) error
}

type DataValidator struct {
	metrics *metrics.CollectorMetrics
}

func New(m *metrics.CollectorMetrics) *DataValidator {
	return &DataValidator{metrics: m}
}

func (v *DataValidator) ValidateCandle(c *model.Candle) error {
	if c.Open <= 0 {
		return v.fail("candle", "open_not_positive")
	}
	if c.High <= 0 {
		return v.fail("candle", "high_not_positive")
	}
	if c.Low <= 0 {
		return v.fail("candle", "low_not_positive")
	}
	if c.Close <= 0 {
		return v.fail("candle", "close_not_positive")
	}
	if c.High < c.Low {
		return v.fail("candle", "high_lt_low")
	}
	if c.High < c.Open {
		return v.fail("candle", "high_lt_open")
	}
	if c.High < c.Close {
		return v.fail("candle", "high_lt_close")
	}
	if c.Low > c.Open {
		return v.fail("candle", "low_gt_open")
	}
	if c.Low > c.Close {
		return v.fail("candle", "low_gt_close")
	}
	if c.Volume < 0 {
		return v.fail("candle", "negative_volume")
	}
	if c.Time.IsZero() {
		return v.fail("candle", "zero_timestamp")
	}
	return nil
}

func (v *DataValidator) ValidateOI(o *model.OIRecord) error {
	if o.OI <= 0 {
		return v.fail("open_interest", "oi_not_positive")
	}
	if o.OIValueUSD <= 0 {
		return v.fail("open_interest", "value_not_positive")
	}
	if o.Time.IsZero() {
		return v.fail("open_interest", "zero_timestamp")
	}
	return nil
}

func (v *DataValidator) ValidateFunding(f *model.FundingRate) error {
	if math.IsNaN(f.Rate) || math.IsInf(f.Rate, 0) {
		return v.fail("funding_rate", "invalid_rate")
	}
	if f.Time.IsZero() {
		return v.fail("funding_rate", "zero_timestamp")
	}
	return nil
}

func (v *DataValidator) ValidateLSRatio(l *model.LSRatio) error {
	if l.BuyRatio < 0 || l.SellRatio < 0 {
		return v.fail("ls_ratio", "negative_ratio")
	}
	if l.Time.IsZero() {
		return v.fail("ls_ratio", "zero_timestamp")
	}
	return nil
}

func (v *DataValidator) ValidateLiquidation(l *model.Liquidation) error {
	if l.Size <= 0 {
		return v.fail("liquidation", "non_positive_size")
	}
	if l.Price <= 0 {
		return v.fail("liquidation", "non_positive_price")
	}
	if l.Side != "Buy" && l.Side != "Sell" {
		return v.fail("liquidation", "invalid_side")
	}
	if l.Time.IsZero() {
		return v.fail("liquidation", "zero_timestamp")
	}
	return nil
}

func (v *DataValidator) fail(table, reason string) error {
	if v.metrics != nil {
		v.metrics.ValidationFailed.WithLabelValues(table, reason).Inc()
	}
	return fmt.Errorf("%s: %s", table, reason)
}
