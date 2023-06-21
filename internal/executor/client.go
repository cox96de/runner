package executor

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	pingEndpoint = "/ping"
)

type Client struct {
	c        *http.Client
	endpoint string
}

func NewClient(endpoint string) *Client {
	c := &Client{
		endpoint: endpoint,
		c: &http.Client{
			// Default timeout.
			Timeout: time.Second * 10,
		},
	}
	return c
}

// Ping checks the executor is bootstrapped and ready to serve.
func (c *Client) Ping(ctx context.Context) error {
	path, err := url.JoinPath(c.endpoint, pingEndpoint)
	if err != nil {
		return errors.WithStack(err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := c.c.Do(request)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		// Ignore too much payload.
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return errors.Errorf("invalid status code: %d with payload: %s", response.StatusCode, string(body))
	}
	return nil
}
