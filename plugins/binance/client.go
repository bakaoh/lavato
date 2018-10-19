package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/bakaoh/lavato/pkg/utils"
	"github.com/google/go-querystring/query"
)

// Client is Binance rest API client
type Client struct {
	baseURL     *url.URL
	httpClient  *http.Client
	apiKey      string
	secretKey   string
	rateLimited time.Time
}

// ErrorResponse ...
type ErrorResponse struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

// NewClient returns a new Binance API client.
func NewClient(apiKey, secretKey string) (*Client, error) {
	baseURL, err := url.Parse("https://api.binance.com")
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: utils.CreateClient(),
		apiKey:     apiKey,
		secretKey:  secretKey,
	}, nil
}

func sign(secret, data string) string {
	signature := hmac.New(sha256.New, []byte(secret))
	signature.Write([]byte(data))
	return hex.EncodeToString(signature.Sum(nil))
}

func (c *Client) newSignedRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.baseURL.ResolveReference(rel)

	data, err := query.Values(body)
	if err != nil {
		return nil, err
	}
	signature := sign(c.secretKey, data.Encode())
	u.RawQuery = data.Encode() + "&signature=" + signature

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-MBX-APIKEY", c.apiKey)
	return req, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	if time.Now().Sub(c.rateLimited) < 1*time.Hour {
		return nil, errors.New("rate limit is violated")
	}
	rel := &url.URL{Path: path}
	u := c.baseURL.ResolveReference(rel)

	data, err := query.Values(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusTooManyRequests {
			c.rateLimited = time.Now()
		}
		errRes := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errRes)
		err = fmt.Errorf("status %d, code %d, msg %s", resp.StatusCode, errRes.Code, errRes.Msg)
	} else {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	io.Copy(ioutil.Discard, resp.Body)
	return resp, err
}

func (c *Client) dump(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := httputil.DumpResponse(resp, true)
	fmt.Println(string(data))
	io.Copy(ioutil.Discard, resp.Body)
	return resp, err
}
