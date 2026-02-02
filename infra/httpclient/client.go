package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

type Doer struct {
	client HTTPClient
}

func NewDoer(client HTTPClient) *Doer {
	return &Doer{client: client}
}

func (d *Doer) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return d.doAndRead(req)
}

func (d *Doer) Head(ctx context.Context, url string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	return d.doAndRead(req)
}

func (d *Doer) doAndRead(req *http.Request) (*Response, error) {
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

type DefaultHTTPClient struct {
	client *http.Client
}

type ClientOption func(*DefaultHTTPClient)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *DefaultHTTPClient) {
		c.client.Timeout = timeout
	}
}

func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *DefaultHTTPClient) {
		c.client.Transport = transport
	}
}

func NewDefaultHTTPClient(opts ...ClientOption) *DefaultHTTPClient {
	c := &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

var _ HTTPClient = (*DefaultHTTPClient)(nil)
