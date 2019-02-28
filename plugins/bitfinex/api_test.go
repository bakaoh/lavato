package bitfinex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var client *Client

func init() {
	client, _ = NewClient(
		"",
		"",
	)
}

func TestTicker(t *testing.T) {
	ticker, err := client.Ticker("ethusd")
	assert.Nil(t, err)

	m := int((ticker.Ts - 1546300800) / 60)
	fmt.Printf("%d", m)
}
