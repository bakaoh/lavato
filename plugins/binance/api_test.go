package binance

import (
	"fmt"
	"testing"

	"github.com/bakaoh/lavato/private"
	"github.com/stretchr/testify/assert"
)

var client *Client

func init() {
	client, _ = NewClient(
		private.BinanceApiKey,
		private.BinanceSecretKey,
	)
}

func TestSign(t *testing.T) {
	secret := "NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j"
	data := "symbol=LTCBTC&side=BUY&type=LIMIT&timeInForce=GTC&quantity=1&price=0.1&recvWindow=5000&timestamp=1499827319559"
	signature := "c8db56825ae71d6d79447849e617115f4a920fa2acdcab2b053c4b2838bd6b71"
	assert.Equal(t, signature, sign(secret, data))
}

func TestTickerPrice(t *testing.T) {
	prices, err := client.TickerPrice()
	assert.Nil(t, err)

	fmt.Println(prices)
}

func TestExchangeInfo(t *testing.T) {
	infos, err := client.ExchangeInfo()
	assert.Nil(t, err)

	fmt.Println(infos)
}

func TestTime(t *testing.T) {
	serverTime, err := client.Time()
	assert.Nil(t, err)

	fmt.Println(serverTime.ServerTime)
}

func TestAllOrders(t *testing.T) {
	orders, err := client.AllOrders("IOTXETH")
	assert.Nil(t, err)

	fmt.Println(orders)
}

func TestNewOrder(t *testing.T) {
	order, err := client.GetOrder("BNBETH", 42067974)
	assert.Nil(t, err)

	fmt.Println(order)
}
