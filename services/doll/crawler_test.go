package doll

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/bakaoh/lavato/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestBasicDb(t *testing.T) {
	tdb := db.NewMMapDb("/tmp/lavato-test.db")
	initDB(tdb)

	i := 1000 * 4
	binary.LittleEndian.PutUint32(tdb.Data[i:], 123)
	v := binary.LittleEndian.Uint32(tdb.Data[i:])
	assert.Equal(t, v, uint32(123))
}

func TestBeginTime(t *testing.T) {
	b := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	s := b.UnixNano() / int64(time.Second)
	assert.Equal(t, s, beginTs)
}
