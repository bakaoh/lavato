package main

import (
	"net/http"
	"time"

	chart "github.com/wcharczuk/go-chart"
)

func drawChart(res http.ResponseWriter, req *http.Request) {
	x := []time.Time{}
	y := []float64{}

	reader := NewReader("ethusd")
	reader.Init()
	now := time.Now().Add(-40 * time.Minute)
	lastPrice := float32(0)
	for i := 0; i < 24*60*5; i++ {
		now = now.Add(-1 * time.Minute)
		price := reader.Price(now.UnixNano() / int64(time.Second))
		if price == 0 {
			price = lastPrice
		} else {
			lastPrice = price
		}
		x = append([]time.Time{now}, x...)
		y = append([]float64{float64(price)}, y...)
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: x,
				YValues: y,
			},
		},
	}

	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func drawCustomChart(res http.ResponseWriter, req *http.Request) {
	/*
	   This is basically the other timeseries example, except we switch to hour intervals and specify a different formatter from default for the xaxis tick labels.
	*/
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: []time.Time{
					time.Now().Add(-10 * time.Hour),
					time.Now().Add(-9 * time.Hour),
					time.Now().Add(-8 * time.Hour),
					time.Now().Add(-7 * time.Hour),
					time.Now().Add(-6 * time.Hour),
					time.Now().Add(-5 * time.Hour),
					time.Now().Add(-4 * time.Hour),
					time.Now().Add(-3 * time.Hour),
					time.Now().Add(-2 * time.Hour),
					time.Now().Add(-1 * time.Hour),
					time.Now(),
				},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0},
			},
		},
	}

	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func main() {
	http.HandleFunc("/", drawChart)
	http.HandleFunc("/favicon.ico", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte{})
	})
	http.HandleFunc("/custom", drawCustomChart)
	http.ListenAndServe(":8080", nil)
}
