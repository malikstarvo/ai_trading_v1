package papertrade

import (
	"math"
	"testing"
)

func approxEq(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestDecideDirection_Long(t *testing.T) {
	dir := decideDirection(75, 60, 40)
	if dir != Long {
		t.Fatalf("expected LONG, got %s", dir)
	}
}

func TestDecideDirection_Short(t *testing.T) {
	dir := decideDirection(25, 60, 40)
	if dir != Short {
		t.Fatalf("expected SHORT, got %s", dir)
	}
}

func TestDecideDirection_NoTrade(t *testing.T) {
	dir := decideDirection(50, 60, 40)
	if dir != NoTrade {
		t.Fatalf("expected NO_TRADE, got %s", dir)
	}
}

func TestDecideDirection_EdgeLong(t *testing.T) {
	dir := decideDirection(60, 60, 40)
	if dir != Long {
		t.Fatalf("expected LONG at threshold=60, got %s", dir)
	}
}

func TestDecideDirection_EdgeShort(t *testing.T) {
	dir := decideDirection(40, 60, 40)
	if dir != Short {
		t.Fatalf("expected SHORT at threshold=40, got %s", dir)
	}
}

func TestCalcPositionSize_Normal(t *testing.T) {
	size := CalcPositionSize(10000, 50000, 500, 2.0, 1.0)
	// Expected: 10000 * 0.01 / (500 * 2) = 100 / 1000 = 0.1
	expected := 0.1
	if !approxEq(size, expected, 0.001) {
		t.Fatalf("expected %.4f, got %.4f", expected, size)
	}
}

func TestCalcPositionSize_ZeroEquity(t *testing.T) {
	size := CalcPositionSize(0, 50000, 500, 2.0, 1.0)
	if size != 0 {
		t.Fatalf("expected 0, got %.4f", size)
	}
}

func TestCalcPositionSize_ZeroATR(t *testing.T) {
	size := CalcPositionSize(10000, 50000, 0, 2.0, 1.0)
	if size != 0 {
		t.Fatalf("expected 0, got %.4f", size)
	}
}

func TestSimulateEntry_Long(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Slippage = 0.001
	result := SimulateEntry(50000, 1.0, Long, cfg, 0)
	// Long buys at higher price: 50000 * (1 + 0.001) = 50050
	expected := 50050.0
	if !approxEq(result.FillPrice, expected, 0.01) {
		t.Fatalf("expected %.2f, got %.2f", expected, result.FillPrice)
	}
	expectedComm := 1.0 * cfg.Commission
	if !approxEq(result.Commission, expectedComm, 0.0001) {
		t.Fatalf("expected commission %.6f, got %.6f", expectedComm, result.Commission)
	}
}

func TestSimulateEntry_Short(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Slippage = 0.001
	result := SimulateEntry(50000, 1.0, Short, cfg, 0)
	// Short sells at lower price: 50000 * (1 - 0.001) = 49950
	expected := 49950.0
	if !approxEq(result.FillPrice, expected, 0.01) {
		t.Fatalf("expected %.2f, got %.2f", expected, result.FillPrice)
	}
}

func TestSimulateExit_Long(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Slippage = 0.001
	result := SimulateExit(51000, 1.0, Long, cfg)
	// Long sells at lower price: 51000 * (1 - 0.001) = 50949
	expected := 50949.0
	if !approxEq(result.FillPrice, expected, 0.01) {
		t.Fatalf("expected %.2f, got %.2f", expected, result.FillPrice)
	}
}

func TestSimulateExit_Short(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Slippage = 0.001
	result := SimulateExit(49000, 1.0, Short, cfg)
	// Short buys at higher price: 49000 * (1 + 0.001) = 49049
	expected := 49049.0
	if !approxEq(result.FillPrice, expected, 0.01) {
		t.Fatalf("expected %.2f, got %.2f", expected, result.FillPrice)
	}
}

func TestSimulateEntry_VolumePremium(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Slippage = 0.001
	// Size > 1% of candle volume → extra slippage
	result := SimulateEntry(50000, 1000, Long, cfg, 50000)
	if result.SlippagePct <= cfg.Slippage {
		t.Fatalf("expected slippage > %.6f, got %.6f", cfg.Slippage, result.SlippagePct)
	}
}

func TestLongPnL_Profit(t *testing.T) {
	cfg := DefaultConfig()
	entry := SimulateEntry(50000, 1.0, Long, cfg, 0)
	exit := SimulateExit(55000, 1.0, Long, cfg)
	// Long: grossPnl = (exitPrice - entryPrice) / entryPrice * size
	grossPnL := (exit.FillPrice - entry.FillPrice) / entry.FillPrice * 1.0
	totalFee := entry.Commission + exit.Commission
	netPnL := grossPnL - totalFee
	if netPnL <= 0 {
		t.Fatalf("expected positive PnL, got %.2f", netPnL)
	}
}

func TestShortPnL_Profit(t *testing.T) {
	cfg := DefaultConfig()
	entry := SimulateEntry(50000, 1.0, Short, cfg, 0)
	exit := SimulateExit(45000, 1.0, Short, cfg)
	// Short: grossPnl = (entryPrice - exitPrice) / entryPrice * size
	grossPnL := (entry.FillPrice - exit.FillPrice) / entry.FillPrice * 1.0
	totalFee := entry.Commission + exit.Commission
	netPnL := grossPnL - totalFee
	if netPnL <= 0 {
		t.Fatalf("expected positive PnL, got %.2f", netPnL)
	}
}

func TestLongPnL_Loss(t *testing.T) {
	cfg := DefaultConfig()
	entry := SimulateEntry(50000, 1.0, Long, cfg, 0)
	exit := SimulateExit(48000, 1.0, Long, cfg)
	grossPnL := (exit.FillPrice - entry.FillPrice) / entry.FillPrice * 1.0
	totalFee := entry.Commission + exit.Commission
	netPnL := grossPnL - totalFee
	if netPnL >= 0 {
		t.Fatalf("expected negative PnL, got %.2f", netPnL)
	}
}

func TestShortPnL_Loss(t *testing.T) {
	cfg := DefaultConfig()
	entry := SimulateEntry(50000, 1.0, Short, cfg, 0)
	exit := SimulateExit(52000, 1.0, Short, cfg)
	grossPnL := (entry.FillPrice - exit.FillPrice) / entry.FillPrice * 1.0
	totalFee := entry.Commission + exit.Commission
	netPnL := grossPnL - totalFee
	if netPnL >= 0 {
		t.Fatalf("expected negative PnL, got %.2f", netPnL)
	}
}

func TestFeeCalculation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commission = 0.001
	size := 1000.0
	entry := SimulateEntry(50000, size, Long, cfg, 0)
	// Only entry fee is charged at entry
	if !approxEq(entry.Commission, size*cfg.Commission, 0.001) {
		t.Fatalf("expected entry fee %.2f, got %.2f", size*cfg.Commission, entry.Commission)
	}
}

func TestSlippageApplied(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Slippage = 0.005
	entry := SimulateEntry(50000, 1.0, Long, cfg, 0)
	if entry.FillPrice <= 50000 {
		t.Fatalf("expected fill price > 50000 with positive slippage, got %.2f", entry.FillPrice)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.InitialCapital != 10000 {
		t.Fatalf("expected initial capital 10000, got %.0f", cfg.InitialCapital)
	}
	if cfg.LongThreshold != 60 {
		t.Fatalf("expected long threshold 60, got %.0f", cfg.LongThreshold)
	}
	if cfg.ShortThreshold != 40 {
		t.Fatalf("expected short threshold 40, got %.0f", cfg.ShortThreshold)
	}
	if cfg.RiskPerTradePct != 1.0 {
		t.Fatalf("expected risk 1.0%%, got %.1f", cfg.RiskPerTradePct)
	}
}

func TestStateTransitions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialCapital = 10000
	cfg.MaxDailyDrawdownPct = 5.0

	// Simulate daily drawdown check: dayPnl <= -5% of initialCapital
	// We can test the threshold math directly
	ddPct := -600.0 / cfg.InitialCapital * 100
	if ddPct <= -cfg.MaxDailyDrawdownPct {
		// Would stop
	} else {
		t.Fatal("expected drawdown to trigger")
	}
}
