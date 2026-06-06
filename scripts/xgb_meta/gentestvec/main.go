package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/avav/ai_trading_v1/internal/agent/orderflow"
	"github.com/avav/ai_trading_v1/internal/agent/regime"
	"github.com/avav/ai_trading_v1/internal/agent/technical"
)

type TestCase struct {
	Name     string  `json:"name"`
	Tech     float64 `json:"tech"`
	OF       float64 `json:"of"`
	Regime   float64 `json:"regime"`
	RegimeLabel string `json:"regime_label"`
}

func main() {
	cases := []TestCase{}

	// Test 1: Strong bullish (index 200, trend=5)
	price := 51000.0
	techIn := technical.Input{
		Price:    price,
		EMA20:    price * 0.99,
		EMA50:    price * 0.97,
		EMA200:   price * 0.95,
		RSI14:    58,
		ATR14:    price * 0.012,
		Volume:   1_100_000,
		VolEMA20: 1_000_000,
		ADX14:    32,
	}
	techOut := technical.Calculate(techIn)

	ofIn := orderflow.Input{
		FundingRate:  0.00002,
		OIDeltaPct:   5.0,
		LSRatio:      1.2,
		LongLiqUSD:   5_000_000,
		ShortLiqUSD:  1_000_000,
	}
	ofOut := orderflow.Calculate(ofIn)

	regIn := regime.Input{
		ADX14:      32,
		ATR14:      price * 0.012,
		Price:      price,
		Volatility: 1.8,
	}
	regOut := regime.Calculate(regIn)

	cases = append(cases, TestCase{
		Name:     "bullish",
		Tech:     techOut.TechnicalScore,
		OF:       ofOut.OrderFlowScore,
		Regime:   regOut.RegimeScore,
		RegimeLabel: regOut.Regime,
	})

	// Test 2: Ranging
	price2 := 50000.0
	techIn2 := technical.Input{
		Price:    price2,
		EMA20:    price2 * 0.998,
		EMA50:    price2 * 0.997,
		EMA200:   price2 * 0.996,
		RSI14:    50,
		ATR14:    price2 * 0.003,
		Volume:   1_000_000,
		VolEMA20: 1_000_000,
		ADX14:    16,
	}
	techOut2 := technical.Calculate(techIn2)

	ofIn2 := orderflow.Input{
		FundingRate:  0.00001,
		OIDeltaPct:   0.1,
		LSRatio:      1.0,
		LongLiqUSD:   100_000,
		ShortLiqUSD:  100_000,
	}
	ofOut2 := orderflow.Calculate(ofIn2)

	regIn2 := regime.Input{
		ADX14:      16,
		ATR14:      price2 * 0.003,
		Price:      price2,
		Volatility: 0.4,
	}
	regOut2 := regime.Calculate(regIn2)

	cases = append(cases, TestCase{
		Name:     "ranging",
		Tech:     techOut2.TechnicalScore,
		OF:       ofOut2.OrderFlowScore,
		Regime:   regOut2.RegimeScore,
		RegimeLabel: regOut2.Regime,
	})

	// Test 3: Chaotic
	price3 := 50500.0
	techIn3 := technical.Input{
		Price:    price3,
		EMA20:    price3 * 0.98,
		EMA50:    price3 * 0.97,
		EMA200:   price3 * 0.96,
		RSI14:    45,
		ATR14:    price3 * 0.04,
		Volume:   1_500_000,
		VolEMA20: 1_000_000,
		ADX14:    18,
	}
	techOut3 := technical.Calculate(techIn3)

	ofIn3 := orderflow.Input{
		FundingRate:  -0.0002,
		OIDeltaPct:   -0.5,
		LSRatio:      0.8,
		LongLiqUSD:   2_000_000,
		ShortLiqUSD:  3_000_000,
	}
	ofOut3 := orderflow.Calculate(ofIn3)

	regIn3 := regime.Input{
		ADX14:      18,
		ATR14:      price3 * 0.04,
		Price:      price3,
		Volatility: 4.0,
	}
	regOut3 := regime.Calculate(regIn3)

	cases = append(cases, TestCase{
		Name:     "chaotic",
		Tech:     techOut3.TechnicalScore,
		OF:       ofOut3.OrderFlowScore,
		Regime:   regOut3.RegimeScore,
		RegimeLabel: regOut3.Regime,
	})

	// Test 4: Neutral
	price4 := 50000.0
	techIn4 := technical.Input{
		Price:    price4,
		EMA20:    price4 * 0.999,
		EMA50:    price4 * 0.998,
		EMA200:   price4 * 0.997,
		RSI14:    50,
		ATR14:    price4 * 0.005,
		Volume:   500_000,
		VolEMA20: 500_000,
		ADX14:    20,
	}
	techOut4 := technical.Calculate(techIn4)

	ofIn4 := orderflow.Input{
		FundingRate:  0.00001,
		OIDeltaPct:   0.0,
		LSRatio:      1.0,
		LongLiqUSD:   100_000,
		ShortLiqUSD:  100_000,
	}
	ofOut4 := orderflow.Calculate(ofIn4)

	regIn4 := regime.Input{
		ADX14:      20,
		ATR14:      price4 * 0.005,
		Price:      price4,
		Volatility: 0.6,
	}
	regOut4 := regime.Calculate(regIn4)

	cases = append(cases, TestCase{
		Name:     "neutral",
		Tech:     techOut4.TechnicalScore,
		OF:       ofOut4.OrderFlowScore,
		Regime:   regOut4.RegimeScore,
		RegimeLabel: regOut4.Regime,
	})

	// Test 5: NaN inputs (all invalid → should return 0 scores)
	techIn5 := technical.Input{
		Price:    price,
		EMA20:    math.NaN(),
		RSI14:    math.NaN(),
	}
	techOut5 := technical.Calculate(techIn5)

	ofIn5 := orderflow.Input{
		FundingRate: math.NaN(),
	}
	ofOut5 := orderflow.Calculate(ofIn5)

	regIn5 := regime.Input{}
	regOut5 := regime.Calculate(regIn5)

	cases = append(cases, TestCase{
		Name:     "nan",
		Tech:     techOut5.TechnicalScore,
		OF:       ofOut5.OrderFlowScore,
		Regime:   regOut5.RegimeScore,
		RegimeLabel: regOut5.Regime,
	})

	// Compute confidence scores
	type FullCase struct {
		Name       string  `json:"name"`
		Tech       float64 `json:"tech"`
		OF         float64 `json:"of"`
		Regime     float64 `json:"regime"`
		RegimeLabel string `json:"regime_label"`
		Confidence float64 `json:"confidence"`
	}

	var full []FullCase
	for _, c := range cases {
		conf := c.Tech*0.40 + c.OF*0.40 + c.Regime*0.20
		full = append(full, FullCase{
			Name:        c.Name,
			Tech:        c.Tech,
			OF:          c.OF,
			Regime:      c.Regime,
			RegimeLabel: c.RegimeLabel,
			Confidence:  conf,
		})
	}

	b, _ := json.MarshalIndent(full, "", "  ")
	fmt.Println(string(b))
}
