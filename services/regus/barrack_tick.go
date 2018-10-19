package regus

import (
	"context"
	"strconv"
)

// TickResponse ...
type TickResponse struct {
	FullUpdate []PaladinResponse `json:"full,omitempty"`
	Prices     map[string]string `json:"prices,omitempty"`
}

// OnTick ...
func (b *Barrack) OnTick() *TickResponse {
	prices := make(map[string]string)
	if b.cache != nil {
		for _, paladin := range b.cache {
			prices[paladin.ID] = b.provider.GetPrice(paladin.Symbol)
			b.checkOrder(context.Background(), &paladin, prices[paladin.ID])
		}
	}
	if b.fullUpdate {
		b.fullUpdate = false
		return &TickResponse{b.GetPaladins(context.Background()), prices}
	}
	return &TickResponse{nil, prices}
}

func (b *Barrack) checkOrder(ctx context.Context, paladin *PaladinResponse, priceStr string) {
	price, _ := strconv.ParseFloat(priceStr, 64)

	inPrice, _ := strconv.ParseFloat(paladin.InPrice, 64)
	if paladin.InStatus != "" && paladin.InStatus != "FILLED" && inPrice >= price {
		b.getAndSaveOrder(ctx, paladin.Symbol, paladin.OutID)
	}

	outPrice, _ := strconv.ParseFloat(paladin.OutPrice, 64)
	if paladin.OutStatus != "" && paladin.OutStatus != "FILLED" && outPrice <= price {
		b.getAndSaveOrder(ctx, paladin.Symbol, paladin.OutID)
	}
}

func (b *Barrack) getAndSaveOrder(ctx context.Context, symbol string, orderID int64) {
	order, err := b.binance.GetOrder(symbol, orderID)
	if err == nil {
		b.storage.SaveOrder(ctx, order)
		if order.Status == "FILLED" {
			b.fullUpdate = true
		}
	}
}

// ShouldFullUpdate ...
func (b *Barrack) ShouldFullUpdate() {
	b.fullUpdate = true
}
