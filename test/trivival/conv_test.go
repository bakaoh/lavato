package trivival

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseFloat(t *testing.T) {
	price := 0.012903
	strPrice := strconv.FormatFloat(price, 'f', 0, 64)
	fmt.Println(strPrice)
	fmt.Printf("%f", price)
}

func TestGetStepSize(t *testing.T) {
	stepSize := "0.00000100"
	size, err := strconv.ParseFloat(stepSize, 64)
	assert.Nil(t, err)
	fmt.Println(math.Log10(1 / size))
}

func TestInteval(t *testing.T) {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case a := <-ticker.C:
			fmt.Println(a)
		}
	}
}

func TestPrintDate(t *testing.T) {
	fmt.Printf("%s", time.Now().String()[:19])
}
