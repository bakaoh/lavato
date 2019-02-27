package regus

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	bin "github.com/bakaoh/lavato/plugins/binance"
	"github.com/pkg/errors"
)

// ActionInfoResponse ...
type ActionInfoResponse struct {
	Msg string `json:"msg,omitempty"`
}

func (i *ActionInfoResponse) prepend(s string) *ActionInfoResponse {
	i.Msg = fmt.Sprintf("%s<br/>%s", s, i.Msg)
	return i
}

func (i *ActionInfoResponse) append(k, v string) *ActionInfoResponse {
	i.Msg += fmt.Sprintf("<b>%s:</b> %s <br/>", k, v)
	return i
}

// ActionInfo ...
func (b *Barrack) ActionInfo(ctx context.Context, id, symbol, act string) (*ActionInfoResponse, error) {
	symbol = strings.ToUpper(symbol) + "ETH"
	paladin, err := b.storage.LoadPaladin(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "paladin %s not found", id)
	}
	var inOrder, outOrder *bin.GetOrderResponse
	info := &ActionInfoResponse{}
	if paladin.InID > 0 {
		inOrder, err = b.storage.LoadOrder(context.Background(), paladin.InID)
		if err != nil {
			return nil, errors.Wrapf(err, "storage load order %d error", paladin.InID)
		}
		info.append("Symbol", inOrder.Symbol)
		info.append("Current price", b.provider.GetPrice(inOrder.Symbol))
		info.append("Buy price", inOrder.Price)
		info.append("Buy date", time.Unix(inOrder.Time/1000, 0).String())
	}
	if paladin.OutID > 0 {
		outOrder, err = b.storage.LoadOrder(context.Background(), paladin.OutID)
		if err != nil {
			return nil, errors.Wrapf(err, "storage load order %d error", paladin.OutID)
		}
		info.append("Target price", outOrder.Price)
		info.append("Sell date", time.Unix(outOrder.Time/1000, 0).String())
	}

	switch act {
	case "cancel":
		if paladin.OutID > 0 {
			info.prepend("Cancel selling")
		} else if paladin.InID > 0 {
			info.prepend("Cancel buying")
		}
		return info, nil
	case "attack":
		info.append("Symbol", symbol)
		info.append("Current price", b.provider.GetPrice(symbol))
		return info.prepend("<b>Attack</b> open position"), nil
	case "hit":
		target, err := b.increasePriceString(inOrder.Price, 5, inOrder.Symbol)
		info.append("Target price", target)
		return info.prepend("<b>Hit</b> set target at 5%"), err
	case "strike":
		target, err := b.increasePriceString(inOrder.Price, 15, inOrder.Symbol)
		info.append("Target price", target)
		return info.prepend("<b>Strike</b> set target at 15%"), err
	case "bash":
		target, err := b.increasePriceString(inOrder.Price, 25, inOrder.Symbol)
		info.append("Target price", target)
		return info.prepend("<b>Bash</b> set target at 25%"), err
	case "defend":
		return info.prepend("<b>Defend</b> close position"), nil
	}
	return nil, fmt.Errorf("invalid action %s", act)
}

// Action ...
func (b *Barrack) Action(ctx context.Context, id, symbol, act string) error {
	symbol = strings.ToUpper(symbol) + "ETH"
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
	case "attack":
		return b.buySymbol(ctx, paladin, symbol)
	case "hit":
		return b.setTarget(ctx, paladin, 5)
	case "strike":
		return b.setTarget(ctx, paladin, 15)
	case "bash":
		return b.setTarget(ctx, paladin, 25)
	case "defend":
		return b.setTarget(ctx, paladin, -100)
	}
	return fmt.Errorf("invalid action %s", act)
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
	if order.Status == "FILLED" {
		return order, nil
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
		price, err := b.increasePriceString(inOrder.Price, percent, inOrder.Symbol)
		if err != nil {
			return err
		}
		if b.lessthanPriceString(price, b.provider.GetPrice(inOrder.Symbol)) {
			// current price is greater than target price, use take profit
			sell, err = b.binance.SellTakeProfit(inOrder.Symbol, inOrder.ExecutedQty, price)
		} else {
			sell, err = b.binance.SellLimit(inOrder.Symbol, inOrder.ExecutedQty, price)
		}
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

func (b *Barrack) increasePriceString(priceStr string, percent int, symbol string) (string, error) {
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return "", errors.Wrapf(err, "can not parse price %s", priceStr)
	}
	price = price * float64(100+percent) / 100
	tickSize := fmt.Sprintf("%d", b.provider.GetTickSize(symbol))
	return fmt.Sprintf("%."+tickSize+"f", price), nil
}

func (b *Barrack) lessthanPriceString(priceStr1 string, priceStr2 string) bool {
	price1, err := strconv.ParseFloat(priceStr1, 64)
	if err != nil {
		return true
	}
	price2, err := strconv.ParseFloat(priceStr2, 64)
	if err != nil {
		return true
	}
	return price1 < price2
}
