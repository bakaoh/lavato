package main

import (
	"encoding/binary"

	"github.com/bakaoh/lavato/pkg/db"
)

const (
	maxLength = 24 * 60 * 366
	beginTs   = int64(1546300800) // 2019-01-01 00:00:00 UTC
)

// Reader ...
type Reader struct {
	symbol   string
	priceDb  *db.MMapDb
	volumeDb *db.MMapDb
}

// NewReader ...
func NewReader(symbol string) *Reader {
	return &Reader{
		symbol,
		db.NewMMapDb("data/" + symbol + ".pdb"),
		db.NewMMapDb("data/" + symbol + ".vdb"),
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

// Init ...
func (t *Reader) Init() {
	initDB(t.priceDb)
	initDB(t.volumeDb)
}

// Price ...
func (t *Reader) Price(ts int64) float32 {
	i := minuteFromBegin(ts) * 4
	return float32(binary.LittleEndian.Uint32(t.priceDb.Data[i:])) / 100
}

// Volume ...
func (t *Reader) Volume(ts int64) float32 {
	i := minuteFromBegin(ts) * 4
	return float32(binary.LittleEndian.Uint32(t.volumeDb.Data[i:])) / 10000
}

func minuteFromBegin(ts int64) int {
	return int((ts - 1546300800) / 60)
}
