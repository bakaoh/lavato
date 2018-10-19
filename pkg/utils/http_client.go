package utils

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

const (
	maxIdleConnections = 100
	keepAliveTimeout   = 600 * time.Second
	requestTimeout     = 5 * time.Second
)

// CreateClient ...
func CreateClient() *http.Client {
	logged := &loggedRoundTripper{
		roundTripper: &http.Transport{
			Dial:                (&net.Dialer{KeepAlive: keepAliveTimeout}).Dial,
			MaxIdleConns:        maxIdleConnections,
			MaxIdleConnsPerHost: maxIdleConnections,
		},
	}
	return &http.Client{
		Transport: logged,
		Timeout:   requestTimeout,
	}
}

// Refs: https://github.com/ernesto-jimenez/httplogger
type loggedRoundTripper struct {
	roundTripper http.RoundTripper
}

func (c *loggedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	c.logRequest(request)
	startTime := time.Now()

	response, err := c.roundTripper.RoundTrip(request)

	duration := time.Since(startTime)
	c.logResponse(request, response, err, duration)
	return response, err
}

// logRequest doens't do anything since we'll be logging replies only
func (c *loggedRoundTripper) logRequest(req *http.Request) {
}

// logResponse logs path, host, status code and duration in milliseconds
func (c *loggedRoundTripper) logResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	if req.Method == "POST" || req.Method == "DELETE" {
		log.Printf(
			"status %d request %s response %s duration %d \n",
			res.StatusCode,
			dumpRequest(req),
			dumpResponse(res),
			duration/time.Millisecond,
		)
	}
}

// dump response for logging
func dumpResponse(resp *http.Response) string {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return "dump response failed: " + err.Error()
	}
	return string(dump)
}

// dump request for logging
func dumpRequest(req *http.Request) string {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return "dump request failed: " + err.Error()
	}
	return string(dump)
}
