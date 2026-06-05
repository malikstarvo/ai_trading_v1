package bybit

import "encoding/json"

type APIResponse struct {
	RetCode int             `json:"retCode"`
	RetMsg  string          `json:"retMsg"`
	Result  json.RawMessage `json:"result"`
	Time    int64           `json:"time"`
}

type KlineItem struct {
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Interval  string `json:"interval"`
	Open      string `json:"open"`
	Close     string `json:"close"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Volume    string `json:"volume"`
	Turnover  string `json:"turnover"`
	Confirm   bool   `json:"confirm"`
	Timestamp int64  `json:"timestamp"`
}

type KlineResponse struct {
	Category string      `json:"category"`
	Symbol   string      `json:"symbol"`
	List     []KlineItem `json:"list"`
}

type OIItem struct {
	OpenInterest string `json:"openInterest"`
	Timestamp    int64  `json:"timestamp"`
}

type OIResponse struct {
	Symbol   string   `json:"symbol"`
	Category string   `json:"category"`
	List     []OIItem `json:"list"`
}

type LSRatioItem struct {
	BuyRatio   string `json:"buyRatio"`
	SellRatio  string `json:"sellRatio"`
	Timestamp  int64  `json:"timestamp"`
}

type LSRatioResponse struct {
	Category string        `json:"category"`
	List     []LSRatioItem `json:"list"`
}

type FundingRateItem struct {
	Symbol     string `json:"symbol"`
	FundingRate string `json:"fundingRate"`
	FundingTime int64  `json:"fundingTime"`
}

type FundingRateResponse struct {
	Category string            `json:"category"`
	List     []FundingRateItem `json:"list"`
}
