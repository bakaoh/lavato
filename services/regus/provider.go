package regus

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	bin "github.com/bakaoh/lavato/plugins/binance"
)

// Provider ...
type Provider struct {
	ticker         *time.Ticker
	binance        *bin.Client
	prices         map[string]string
	priceTickSizes map[string]int
	lotStepSizes   map[string]int
	quit           chan struct{}
}

// NewProvider ...
func NewProvider(binance *bin.Client) *Provider {
	return &Provider{
		time.NewTicker(2 * time.Second),
		binance,
		make(map[string]string),
		make(map[string]int),
		make(map[string]int),
		make(chan struct{}),
	}
}

// Run ...
func (t *Provider) Run() {
	t.getInfos()
	for {
		select {
		case <-t.ticker.C:
			t.getPrices()
		case <-t.quit:
			t.ticker.Stop()
			return
		}
	}
}

// GetPrice ...
func (t *Provider) GetPrice(symbol string) string {
	return t.prices[symbol]
}

// GetTickSize ...
func (t *Provider) GetTickSize(symbol string) int {
	return t.priceTickSizes[symbol]
}

// GetStepSize ...
func (t *Provider) GetStepSize(symbol string) int {
	return t.lotStepSizes[symbol]
}

// Stop ...
func (t *Provider) Stop() {
	close(t.quit)
}

func (t *Provider) getPrices() {
	prices, err := t.binance.TickerPrice()
	if err == nil {
		for _, price := range prices {
			t.check(t.prices[price.Symbol], price.Price, price.Symbol)
			t.prices[price.Symbol] = price.Price
		}
	}
}

func (t *Provider) check(prev, curr, symbol string) {
	if !strings.HasSuffix(symbol, "ETH") {
		return
	}

	prevPrice, _ := strconv.ParseFloat(prev, 64)
	currPrice, _ := strconv.ParseFloat(curr, 64)
	if prevPrice > 0 && (currPrice-prevPrice)/prevPrice > 0.01 {
		fmt.Printf("%s: %s %s => %s (%f) \n",
			time.Now().String()[:19],
			symbol,
			prev,
			curr,
			(currPrice-prevPrice)*100/prevPrice,
		)
	}
}

func (t *Provider) getInfos() {
	infos, err := t.binance.ExchangeInfo()
	if err != nil {
		log.Fatal("can not get exchange info", err)
	}
	for _, symbol := range infos.Symbols {
		for _, filter := range symbol.Filters {
			switch filter.FilterType {
			case "PRICE_FILTER":
				t.priceTickSizes[symbol.Symbol] = parseSize(filter.TickSize)
			case "LOT_SIZE":
				t.lotStepSizes[symbol.Symbol] = parseSize(filter.StepSize)
			}
		}
	}
}

func parseSize(size string) int {
	s, _ := strconv.ParseFloat(size, 64)
	return int(math.Log10(1 / s))
}
