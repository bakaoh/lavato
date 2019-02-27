package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/bakaoh/lavato/plugins/binance"
	"github.com/bakaoh/lavato/private"
	"github.com/bakaoh/lavato/services/regus"
	"github.com/stretchr/testify/assert"
)

var (
	testStorage *regus.Storage
	testClient  *binance.Client
)

var sheets = map[string]string{
	"XEMETH":   "0075",
	"IOTXETH":  "0256",
	"STRATETH": "0295",
	"MTLETH":   "0553",
	"PPTETH":   "0880",
	"GTOETH":   "0553",
	"ELFETH":   "0880",
	"NPXSETH":  "0553",
	"BCDETH":   "0256",
	"AIONETH":  "0256",
	"SKYETH":   "0256",
	"CLOAKETH": "0553",
}

var paladins = map[string]*regus.Paladin{
	"0075": &regus.Paladin{ID: "0075", Name: "PEREGRINE PALADIN / LARUT", RawHP: 1000},
	"0256": &regus.Paladin{ID: "0256", Name: "MAGE PALADIN / DISTRIER", RawHP: 1000},
	"0295": &regus.Paladin{ID: "0295", Name: "FLASH PALADIN / IBERT", RawHP: 1000},
	"0553": &regus.Paladin{ID: "0553", Name: "LADY PALADIN / MIRELIA", RawHP: 1000},
	"0880": &regus.Paladin{ID: "0880", Name: "PALADIN OF TRUTH / INZAGHI", RawHP: 1000},
}

func TestMain(m *testing.M) {

	testStorage, _ = regus.NewStorage("../../regus.db")
	testClient, _ = binance.NewClient(
		private.BinanceApiKey,
		private.BinanceSecretKey,
	)

	defer testStorage.Close()

	os.Exit(m.Run())
}

func TestFixBug(t *testing.T) {
	t.Skip()
	paladin, _ := testStorage.LoadPaladin(context.Background(), "0295")
	paladin.OutID = 0
	testStorage.SavePaladin(context.Background(), paladin)
}

func TestEvent_STRATEGIST(t *testing.T) {
	// eventPaladins := map[string]*regus.Paladin{
	// 	"0075": &regus.Paladin{ID: "0079", Name: "AZURE DRAGON - EAST / VORDORE", RawHP: 1000},
	// 	"0256": &regus.Paladin{ID: "0334", Name: "RED PHOENIX - SOUTH / SNAF", RawHP: 1000},
	// 	"0295": &regus.Paladin{ID: "0634", Name: "BLACK TORTOISE - NORTH / LADOL", RawHP: 1000},
	// 	"0553": &regus.Paladin{ID: "0881", Name: "WHITE TIGER - WEST / ROZARI", RawHP: 1000},
	// }
}

func TestEvent_ARCHMAGE(t *testing.T) {
	t.Skip()
	eventPaladins := map[string]*regus.Paladin{
		"0080": &regus.Paladin{ID: "0080", Name: "AQUA SORCERER / MYSTERE", RawHP: 1000, Type: "Event"},
		"0259": &regus.Paladin{ID: "0259", Name: "WIND SORCERESS / FEMIEL", RawHP: 1000, Type: "Event"},
		"0343": &regus.Paladin{ID: "0343", Name: "FIRE SORCERESS / ADDITION", RawHP: 1000, Type: "Event"},
		"0459": &regus.Paladin{ID: "0459", Name: "FOREST SORCERESS / ANTTILA", RawHP: 1000, Type: "Event"},
		"0556": &regus.Paladin{ID: "0556", Name: "FROST SORCERESS / RASAM", RawHP: 1000, Type: "Event"},
	}

	eventSymbol := map[string]string{
		"ADXETH":  "0080",
		"DENTETH": "0259",
		"ASTETH":  "0343",
		"ENGETH":  "0459",
		"BLZETH":  "0556",
	}
	for s, p := range eventSymbol {
		time.Sleep(1 * time.Second)
		orders, err := testClient.AllOrders(s)
		if err != nil {
			continue
		}
		paladin, ok := eventPaladins[p]
		if !ok {
			continue
		}
		for _, o := range orders {
			testStorage.SaveOrder(context.Background(), &o)
			if o.Status == "FILLED" && o.Side == "BUY" {
				paladin.InID = o.OrderID
			}
			if o.Status == "NEW" && o.Side == "SELL" {
				paladin.OutID = o.OrderID
			}
		}

		testStorage.SavePaladin(context.Background(), paladin)
	}
}

func TestStorePaladin(t *testing.T) {
	t.Skip()
	orders, err := testStorage.LoadAllOrders(context.Background())
	assert.Nil(t, err)

	ongoing := make(map[string]*regus.Battle)
	for _, o := range orders {
		if o.Status == "FILLED" && o.Side == "BUY" {
			battle := &regus.Battle{Symbol: o.Symbol, PaladinID: sheets[o.Symbol]}
			battle.InID = o.OrderID
			ongoing[o.Symbol] = battle
		}

		if o.Status == "FILLED" && o.Side == "SELL" {
			delete(ongoing, o.Symbol)
		}

		if o.Status == "NEW" && o.Side == "SELL" {
			battle := ongoing[o.Symbol]
			battle.OutID = o.OrderID
		}
	}

	for _, o := range ongoing {
		paladins[o.PaladinID].InID = o.InID
		paladins[o.PaladinID].OutID = o.OutID
		inOrder, _ := testStorage.LoadOrder(context.Background(), o.InID)
		hp, _ := strconv.ParseFloat(inOrder.CumQuoteQty, 64)
		paladins[o.PaladinID].ModHP = int(hp*1000) - paladins[o.PaladinID].RawHP + 1
		testStorage.SavePaladin(context.Background(), paladins[o.PaladinID])
	}
}

func TestStoreBattle(t *testing.T) {
	t.Skip()
	orders, err := testStorage.LoadAllOrders(context.Background())
	assert.Nil(t, err)

	ongoing := make(map[string]*regus.Battle)
	for _, o := range orders {
		if o.Status != "FILLED" {
			continue
		}

		battle := ongoing[o.Symbol]
		if battle == nil {
			battle = &regus.Battle{Symbol: o.Symbol, PaladinID: sheets[o.Symbol]}
		}

		if o.Side == "BUY" {
			battle.InID = o.OrderID
		} else if o.Side == "SELL" {
			battle.OutID = o.OrderID
		}
		ongoing[o.Symbol] = battle

		if battle.InID*battle.OutID > 0 {
			testStorage.SaveBattle(context.Background(), battle)
			delete(ongoing, o.Symbol)
		}
	}

	fmt.Println(ongoing)
}

func TestStoreOrder(t *testing.T) {
	t.Skip()
	for s := range sheets {
		time.Sleep(1 * time.Second)
		orders, err := testClient.AllOrders(s)
		if err != nil {
			continue
		}
		for _, o := range orders {
			testStorage.SaveOrder(context.Background(), &o)
		}
	}
}
