package bybit

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	baseURL            string
	httpClient         *http.Client
	rateLimiter        RateLimiter
	logger             *slog.Logger
}

func NewClient(baseURL string, rateLimiter RateLimiter, logger *slog.Logger, insecureSkipVerify bool) *Client {
	transport := &http.Transport{}
	if insecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second, Transport: transport},
		rateLimiter: rateLimiter,
		logger:      logger.With("module", "bybit_client"),
	}
}

func (c *Client) GetKlines(ctx context.Context, symbol, interval string, start, end int64, limit int) (*KlineResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	params := url.Values{}
	params.Set("category", "linear")
	params.Set("symbol", symbol)
	params.Set("interval", interval)
	params.Set("start", strconv.FormatInt(start, 10))
	params.Set("end", strconv.FormatInt(end, 10))
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doGET(ctx, "/v5/market/kline", params)
	if err != nil {
		return nil, err
	}

	var result KlineResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal kline: %w", err)
	}
	return &result, nil
}

func (c *Client) GetOpenInterest(ctx context.Context, symbol, interval string, limit int) (*OIResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	params := url.Values{}
	params.Set("category", "linear")
	params.Set("symbol", symbol)
	params.Set("intervalTime", interval)
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doGET(ctx, "/v5/market/open-interest", params)
	if err != nil {
		return nil, err
	}

	var result OIResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal open-interest: %w", err)
	}
	return &result, nil
}

func (c *Client) GetLongShortRatio(ctx context.Context, symbol, period string, limit int) (*LSRatioResponse, error) {
	if limit <= 0 {
		limit = 500
	}
	params := url.Values{}
	params.Set("category", "linear")
	params.Set("symbol", symbol)
	params.Set("period", period)
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doGET(ctx, "/v5/market/account-ratio", params)
	if err != nil {
		return nil, err
	}

	var result LSRatioResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal account-ratio: %w", err)
	}
	return &result, nil
}

func (c *Client) GetFundingRateHistory(ctx context.Context, symbol string, start, end int64, limit int) (*FundingRateResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	params := url.Values{}
	params.Set("category", "linear")
	params.Set("symbol", symbol)
	if start > 0 {
		params.Set("startTime", strconv.FormatInt(start, 10))
	}
	if end > 0 {
		params.Set("endTime", strconv.FormatInt(end, 10))
	}
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doGET(ctx, "/v5/market/funding/history", params)
	if err != nil {
		return nil, err
	}

	var result FundingRateResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal funding-history: %w", err)
	}
	return &result, nil
}

func (c *Client) LatestCandleTime(ctx context.Context, symbol, interval string) (time.Time, error) {
	resp, err := c.GetKlines(ctx, symbol, interval, 0, time.Now().UnixMilli(), 1)
	if err != nil {
		return time.Time{}, fmt.Errorf("latest candle time: %w", err)
	}
	if len(resp.List) == 0 {
		return time.Time{}, fmt.Errorf("no candles returned for %s %s", symbol, interval)
	}
	return time.UnixMilli(resp.List[0].Start), nil
}

func (c *Client) doGET(ctx context.Context, path string, params url.Values) (*APIResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.logger.Debug("rest request", "url", u)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if apiResp.RetCode != 0 {
		return nil, fmt.Errorf("bybit error %d: %s", apiResp.RetCode, apiResp.RetMsg)
	}

	return &apiResp, nil
}
