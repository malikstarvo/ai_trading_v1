package bybit

import "encoding/json"

type APIResponse struct {
	RetCode int             `json:"retCode"`
	RetMsg  string          `json:"retMsg"`
	Result  json.RawMessage `json:"result"`
	Time    int64           `json:"time"`
}

// KlineItem is a positional string array from the REST kline API.
// Index: 0=startTime, 1=openPrice, 2=highPrice, 3=lowPrice, 4=closePrice, 5=volume, 6=turnover
type KlineItem []string

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
	Timestamp  string `json:"timestamp"`
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

type InstrumentItem struct {
	Symbol string `json:"symbol"`
}

type InstrumentResponse struct {
	Category string           `json:"category"`
	List     []InstrumentItem `json:"list"`
}
