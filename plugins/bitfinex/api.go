package bitfinex

import (
	"context"
)

// TickerResponse return from Ticker request
type TickerResponse struct {
	Mid    float32 `json:"mid,string"`
	Bid    float32 `json:"bid,string"`
	Ask    float32 `json:"ask,string"`
	Last   float32 `json:"last_price,string"`
	Low    float32 `json:"low,string"`
	High   float32 `json:"high,string"`
	Volume float32 `json:"volume,string"`
	Ts     float64 `json:"timestamp,string"`
}

// Ticker ...
func (c *Client) Ticker(symbol string) (*TickerResponse, error) {
	req, err := c.newRequest("GET", "/v1/pubticker/"+symbol, nil)
	if err != nil {
		return nil, err
	}

	var rs TickerResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}
