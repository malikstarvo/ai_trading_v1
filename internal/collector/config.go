package collector

import "time"

type Config struct {
	WSURL              string   `mapstructure:"ws_url"`
	BaseURL            string   `mapstructure:"base_url"`
	ProxyURL           string   `mapstructure:"proxy_url"`
	Symbols            []string `mapstructure:"symbols"`
	Testnet            bool     `mapstructure:"testnet"`
	InsecureSkipVerify bool     `mapstructure:"insecure_skip_verify"`

	BackfillDays      int   `mapstructure:"backfill_days"`
	BatchInsertSize   int   `mapstructure:"batch_insert_size"`

	Recovery struct {
		Enabled       bool          `mapstructure:"enabled"`
		CheckInterval time.Duration `mapstructure:"check_interval"`
		GapBars       int           `mapstructure:"gap_bars"`
	} `mapstructure:"recovery"`

	RateLimit struct {
		RequestsPerSecond float64 `mapstructure:"requests_per_second"`
		Burst             int     `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`
}

func DefaultConfig() Config {
	cfg := Config{
		WSURL:              "wss://stream.bybit.com/v5/public/linear",
		BaseURL:            "https://api.bybit.com",
		Symbols:            []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"},
		Testnet:            false,
		InsecureSkipVerify: false,
		BackfillDays:       30,
		BatchInsertSize:    50,
	}
	cfg.Recovery.Enabled = true
	cfg.Recovery.CheckInterval = 5 * time.Minute
	cfg.Recovery.GapBars = 2
	cfg.RateLimit.RequestsPerSecond = 15
	cfg.RateLimit.Burst = 8
	return cfg
}
