package regus

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	bin "github.com/bakaoh/lavato/plugins/binance"
	"github.com/pkg/errors"
)

// Action ...
func (b *Barrack) Action(ctx context.Context, id, symbol, act string) error {
	paladin, err := b.storage.LoadPaladin(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "paladin %s not found", id)
	}

	switch act {
	case "cancel":
		if paladin.OutID > 0 {
			err := b.cancelOrder(ctx, paladin.OutID)
			if err != nil {
				return err
			}
			paladin.OutID = 0
		} else if paladin.InID > 0 {
			err := b.cancelOrder(ctx, paladin.InID)
			if err != nil {
				return err
			}
			paladin.InID = 0
		}
		b.storage.SavePaladin(context.Background(), paladin)
	case "atk":
		return b.buySymbol(ctx, paladin, symbol)
	case "quick":
	case "flash":
	case "target_5":
		return b.setTarget(ctx, paladin, 5)
	case "target_10":
		return b.setTarget(ctx, paladin, 10)
	case "target_20":
		return b.setTarget(ctx, paladin, 20)
	case "def":
		return b.setTarget(ctx, paladin, -100)
	}
	return nil
}

func (b *Barrack) cancelOrder(ctx context.Context, orderID int64) error {
	order, err := b.getUpToDateOrder(context.Background(), orderID)
	if err != nil {
		return err
	}
	if order.Status == "CANCELED" {
		return nil
	}
	if order.Status != "NEW" {
		return fmt.Errorf("can not cancel order %d with status %s", orderID, order.Status)
	}

	_, err = b.binance.DeleteOrder(order.Symbol, orderID)
	if err != nil {
		return errors.Wrapf(err, "cancel order %d err", orderID)
	}
	order.Status = "CANCELED"
	b.storage.SaveOrder(ctx, order)
	return nil
}

func (b *Barrack) getUpToDateOrder(ctx context.Context, orderID int64) (*bin.GetOrderResponse, error) {
	order, err := b.storage.LoadOrder(context.Background(), orderID)
	if err != nil {
		return nil, errors.Wrapf(err, "storage load order %d error", orderID)
	}
	order, err = b.binance.GetOrder(order.Symbol, orderID)
	if err != nil {
		return nil, errors.Wrapf(err, "binance get order %d error", orderID)
	}
	b.storage.SaveOrder(ctx, order)
	return order, nil
}

func (b *Barrack) buySymbol(ctx context.Context, paladin *Paladin, symbol string) error {
	// check status
	if paladin.InID > 0 {
		return errors.New("paladin is attacking")
	}

	symbol = strings.ToUpper(symbol) + "ETH"
	priceStr := b.provider.GetPrice(symbol)
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return errors.Wrapf(err, "can not parse price %s", priceStr)
	}
	qty := float64(paladin.RawHP+paladin.ModHP) / (1000 * price)
	stepSize := fmt.Sprintf("%d", b.provider.GetStepSize(symbol))
	buy, err := b.binance.BuyLimit(symbol, fmt.Sprintf("%."+stepSize+"f", qty), priceStr)
	if err != nil || buy == nil || buy.OrderID <= 0 {
		return errors.Wrapf(err, "buy order err")
	}

	// update storage
	paladin.InID = buy.OrderID
	b.storage.SavePaladin(context.Background(), paladin)

	inOrder, err := b.binance.GetOrder(symbol, paladin.InID)
	if err != nil {
		return errors.Wrapf(err, "binance get order %d error", paladin.InID)
	}
	b.storage.SaveOrder(context.Background(), inOrder)

	return nil
}

func (b *Barrack) setTarget(ctx context.Context, paladin *Paladin, percent int) error {
	// check status
	if paladin.InID == 0 {
		return errors.New("paladin is not attacking")
	}
	inOrder, err := b.getUpToDateOrder(context.Background(), paladin.InID)
	if err != nil {
		return err
	}
	if inOrder.Status != "FILLED" {
		return fmt.Errorf("order %d is not FILLED", inOrder.OrderID)
	}
	if paladin.OutID != 0 {
		return errors.New("out order is pending")
	}

	// post sell order
	var sell *bin.PostOrderResponse
	if percent == -100 {
		sell, err = b.binance.SellMarket(inOrder.Symbol, inOrder.ExecutedQty)
	} else {
		price, err := strconv.ParseFloat(inOrder.Price, 64)
		if err != nil {
			return errors.Wrapf(err, "can not parse price %s", inOrder.Price)
		}
		price = price * float64(100+percent) / 100
		tickSize := fmt.Sprintf("%d", b.provider.GetTickSize(inOrder.Symbol))
		sell, err = b.binance.SellLimit(inOrder.Symbol, inOrder.ExecutedQty, fmt.Sprintf("%."+tickSize+"f", price))
	}
	if err != nil || sell == nil || sell.OrderID <= 0 {
		return errors.Wrapf(err, "sell order err")
	}

	// update storage
	paladin.OutID = sell.OrderID
	b.storage.SavePaladin(context.Background(), paladin)

	outOrder, err := b.binance.GetOrder(inOrder.Symbol, paladin.OutID)
	if err != nil {
		return errors.Wrapf(err, "binance get order %d error", paladin.OutID)
	}
	b.storage.SaveOrder(context.Background(), outOrder)

	return nil
}
