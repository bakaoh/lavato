package doll

import (
	"encoding/binary"
	"time"

	"github.com/bakaoh/lavato/pkg/db"
	bfn "github.com/bakaoh/lavato/plugins/bitfinex"
)

const (
	maxLength = 24 * 60 * 366
	beginTs   = int64(1546300800) // 2019-01-01 00:00:00 UTC
)

// Crawler ...
type Crawler struct {
	symbol   string
	ticker   *time.Ticker
	bitfinex *bfn.Client
	priceDb  *db.MMapDb
	volumeDb *db.MMapDb
	quit     chan struct{}
}

// NewCrawler ...
func NewCrawler(bitfinex *bfn.Client, symbol string) *Crawler {
	return &Crawler{
		symbol,
		time.NewTicker(1 * time.Minute),
		bitfinex,
		db.NewMMapDb(symbol + ".pdb"),
		db.NewMMapDb(symbol + ".vdb"),
		make(chan struct{}),
	}
}

func initDB(db *db.MMapDb) {
	err := db.Open()
	if err != nil {
		panic(err)
	}
	err = db.Resize(maxLength * 4)
	if err != nil {
		panic(err)
	}
	err = db.Mmap(maxLength * 4)
	if err != nil {
		panic(err)
	}
}

// Run ...
func (t *Crawler) Run() {
	initDB(t.priceDb)
	initDB(t.volumeDb)
	t.onTicker()
	for {
		select {
		case <-t.ticker.C:
			t.onTicker()
		case <-t.quit:
			t.ticker.Stop()
			return
		}
	}
}

// Stop ...
func (t *Crawler) Stop() {
	close(t.quit)
}

func (t *Crawler) onTicker() {
	tick, err := t.bitfinex.Ticker(t.symbol)
	if err == nil {
		i := minuteFromBegin(tick.Ts) * 4
		binary.LittleEndian.PutUint32(t.priceDb.Data[i:], uint32(tick.Last)*100)
		binary.LittleEndian.PutUint32(t.volumeDb.Data[i:], uint32(tick.Last)*10000)
	}
}

func minuteFromBegin(ts float64) int {
	return int((ts - 1546300800) / 60)
}
