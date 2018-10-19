package regus

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bakaoh/lavato/plugins/binance"
	"github.com/stretchr/testify/assert"
)

var testStorage *Storage

func TestMain(m *testing.M) {

	db, _ := NewStorage("../../regus.db")
	testStorage = db
	defer testStorage.Close()

	os.Exit(m.Run())
}

func TestBasicSetGet(t *testing.T) {
	item := &binance.GetOrderResponse{
		Symbol:        "BNBETH",
		OrderID:       42067974,
		ClientOrderID: "5uL1vtAx1aHPiOETnz3S30",
		UpdateTime:    1539576878377,
		Type:          "MARKET",
		Side:          "SELL",
	}

	err := testStorage.SaveOrder(context.Background(), item)
	assert.Nil(t, err)

	get, err := testStorage.LoadOrder(context.Background(), item.OrderID)
	assert.Nil(t, err)
	assert.Equal(t, item, get)
}

func TestGetMap(t *testing.T) {
	orders, err := testStorage.GetMapOrders(context.Background())
	assert.Nil(t, err)

	for _, o := range orders {
		fmt.Println(o)
	}
}

func TestLoadPaladin(t *testing.T) {
	paladin, err := testStorage.LoadPaladin(context.Background(), "0256")
	assert.Nil(t, err)

	fmt.Println(paladin)
}
