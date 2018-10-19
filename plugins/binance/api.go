package binance

import (
	"context"
	"time"
)

// SymbolRequest is request with symbol field
type SymbolRequest struct {
	Symbol    string `url:"symbol,omitempty"`
	Timestamp int64  `url:"timestamp,omitempty"`
}

// TickerPriceResponse return from TickerPrice request
type TickerPriceResponse struct {
	Symbol string `json:"symbol,omitempty"`
	Price  string `json:"price,omitempty"`
}

// TickerPrice ...
func (c *Client) TickerPrice() ([]TickerPriceResponse, error) {
	req, err := c.newRequest("GET", "/api/v3/ticker/price", nil)
	if err != nil {
		return nil, err
	}

	var rs []TickerPriceResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// GetOrderRequest ...
type GetOrderRequest struct {
	Symbol            string `url:"symbol,omitempty"`
	OrderID           int64  `url:"orderId,omitempty"`
	OrigClientOrderID string `url:"origClientOrderId,omitempty"`
	RecvWindow        int64  `url:"recvWindow,omitempty"`
	Timestamp         int64  `url:"timestamp,omitempty"`
}

// GetOrderResponse return from get order request
type GetOrderResponse struct {
	Symbol        string `json:"symbol,omitempty"`
	OrderID       int64  `json:"orderId,omitempty" gorm:"primary_key"`
	ClientOrderID string `json:"clientOrderId,omitempty"`
	Price         string `json:"price,omitempty"`
	OrigQty       string `json:"origQty,omitempty"`
	ExecutedQty   string `json:"executedQty,omitempty"`
	CumQuoteQty   string `json:"cummulativeQuoteQty,omitempty"`
	Status        string `json:"status,omitempty"`
	TimeInForce   string `json:"timeInForce,omitempty"`
	Type          string `json:"type,omitempty"`
	Side          string `json:"side,omitempty"`
	StopPrice     string `json:"stopPrice,omitempty"`
	Time          int64  `json:"time,omitempty"`
	UpdateTime    int64  `json:"updateTime,omitempty"`
	IsWorking     bool   `json:"isWorking,omitempty"`
}

// AllOrders ...
func (c *Client) AllOrders(symbol string) ([]GetOrderResponse, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	req, err := c.newSignedRequest("GET", "/api/v3/allOrders", SymbolRequest{symbol, timestamp})
	if err != nil {
		return nil, err
	}

	var rs []GetOrderResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// GetOrder ...
func (c *Client) GetOrder(symbol string, orderID int64) (*GetOrderResponse, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	req, err := c.newSignedRequest("GET", "/api/v3/order", GetOrderRequest{symbol, orderID, "", 0, timestamp})
	if err != nil {
		return nil, err
	}

	var rs GetOrderResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

// PostOrderRequest ...
type PostOrderRequest struct {
	Symbol           string `url:"symbol,omitempty"`
	Side             string `url:"side,omitempty"`
	Type             string `url:"type,omitempty"`
	TimeInForce      string `url:"timeInForce,omitempty"`
	Quantity         string `url:"quantity,omitempty"`
	Price            string `url:"price,omitempty"`
	NewClientOrderID string `url:"newClientOrderId,omitempty"`
	StopPrice        string `url:"stopPrice,omitempty"`
	IcebergQty       string `url:"icebergQty,omitempty"`
	NewOrderRespType string `url:"newOrderRespType,omitempty"`
	RecvWindow       int64  `url:"recvWindow,omitempty"`
	Timestamp        int64  `url:"timestamp,omitempty"`
}

// PostOrderFill ...
type PostOrderFill struct {
	Price           string `json:"price,omitempty"`
	Qty             string `json:"qty,omitempty"`
	Commission      string `json:"commission,omitempty"`
	CommissionAsset string `json:"commissionAsset,omitempty"`
}

// PostOrderResponse is Response FULL
type PostOrderResponse struct {
	Symbol        string          `json:"symbol,omitempty"`
	OrderID       int64           `json:"orderId,omitempty"`
	ClientOrderID string          `json:"clientOrderId,omitempty"`
	TransactTime  int64           `json:"transactTime,omitempty"`
	Price         string          `json:"price,omitempty"`
	OrigQty       string          `json:"origQty,omitempty"`
	ExecutedQty   string          `json:"executedQty,omitempty"`
	CumQuoteQty   string          `json:"cummulativeQuoteQty,omitempty"`
	Status        string          `json:"status,omitempty"`
	TimeInForce   string          `json:"timeInForce,omitempty"`
	Type          string          `json:"type,omitempty"`
	Side          string          `json:"side,omitempty"`
	Fills         []PostOrderFill `json:"fills,omitempty"`
}

// PostOrder ...
func (c *Client) PostOrder(order PostOrderRequest) (*PostOrderResponse, error) {
	order.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	req, err := c.newSignedRequest("POST", "/api/v3/order", order)
	if err != nil {
		return nil, err
	}

	var rs PostOrderResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

// SellMarket ...
func (c *Client) SellMarket(symbol string, quantity string) (*PostOrderResponse, error) {
	return c.PostOrder(PostOrderRequest{
		Symbol:   symbol,
		Type:     "MARKET",
		Side:     "SELL",
		Quantity: quantity,
	})
}

// SellLimit ...
func (c *Client) SellLimit(symbol string, quantity string, price string) (*PostOrderResponse, error) {
	return c.PostOrder(PostOrderRequest{
		Symbol:      symbol,
		Type:        "LIMIT",
		Side:        "SELL",
		Quantity:    quantity,
		Price:       price,
		TimeInForce: "GTC",
	})
}

// BuyMarket ...
func (c *Client) BuyMarket(symbol string, quantity string) (*PostOrderResponse, error) {
	return c.PostOrder(PostOrderRequest{
		Symbol:   symbol,
		Type:     "MARKET",
		Side:     "BUY",
		Quantity: quantity,
	})
}

// BuyLimit ...
func (c *Client) BuyLimit(symbol string, quantity string, price string) (*PostOrderResponse, error) {
	return c.PostOrder(PostOrderRequest{
		Symbol:      symbol,
		Type:        "LIMIT",
		Side:        "BUY",
		Quantity:    quantity,
		Price:       price,
		TimeInForce: "GTC",
	})
}

// DeleteOrderRequest ...
type DeleteOrderRequest struct {
	Symbol            string `url:"symbol,omitempty"`
	OrderID           int64  `url:"orderId,omitempty"`
	OrigClientOrderID string `url:"origClientOrderId,omitempty"`
	NewClientOrderID  string `url:"newClientOrderId,omitempty"`
	RecvWindow        int64  `url:"recvWindow,omitempty"`
	Timestamp         int64  `url:"timestamp,omitempty"`
}

// DeleteOrderResponse ...
type DeleteOrderResponse struct {
	Symbol            string `json:"symbol,omitempty"`
	OrderID           int64  `json:"orderId,omitempty"`
	ClientOrderID     string `json:"clientOrderId,omitempty"`
	OrigClientOrderID string `json:"origClientOrderId,omitempty"`
}

// DeleteOrder ...
func (c *Client) DeleteOrder(symbol string, orderID int64) (*DeleteOrderResponse, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	req, err := c.newSignedRequest("DELETE", "/api/v3/order", DeleteOrderRequest{symbol, orderID, "", "", 0, timestamp})
	if err != nil {
		return nil, err
	}

	var rs DeleteOrderResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

// FilterInfoResponse ...
type FilterInfoResponse struct {
	FilterType string `json:"filterType,omitempty"`
	// PRICE_FILTER
	MinPrice string `json:"minPrice,omitempty"`
	MaxPrice string `json:"maxPrice,omitempty"`
	TickSize string `json:"tickSize,omitempty"`
	// LOT_SIZE
	MinQty   string `json:"minQty,omitempty"`
	MaxQty   string `json:"maxQty,omitempty"`
	StepSize string `json:"stepSize,omitempty"`
}

// SymbolInfoResponse ...
type SymbolInfoResponse struct {
	Symbol             string               `json:"symbol,omitempty"`
	Status             string               `json:"status,omitempty"`
	BaseAsset          string               `json:"baseAsset,omitempty"`
	BaseAssetPrecision int                  `json:"baseAssetPrecision,omitempty"`
	QuoteAsset         string               `json:"quoteAsset,omitempty"`
	QuotePrecision     int                  `json:"quotePrecision,omitempty"`
	Filters            []FilterInfoResponse `json:"filters,omitempty"`
}

// ExchangeInfoResponse ...
type ExchangeInfoResponse struct {
	Symbols []SymbolInfoResponse `json:"symbols,omitempty"`
}

// ExchangeInfo ...
func (c *Client) ExchangeInfo() (*ExchangeInfoResponse, error) {
	req, err := c.newRequest("GET", "/api/v1/exchangeInfo", nil)
	if err != nil {
		return nil, err
	}

	var rs ExchangeInfoResponse
	_, err = c.do(context.Background(), req, &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}
