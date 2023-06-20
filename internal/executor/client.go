package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	pingEndpoint         = "/ping"
	startCommandEndpoint = "/commands/%s"
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
	request, err := c.newRPCRequest(ctx, http.MethodGet, pingEndpoint, nil)
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

type StartCommandRequest struct {
	Dir     string            `json:"dir"`
	Path    string            `json:"path"`
	Command []string          `json:"command"`
	Env     map[string]string `json:"env"`
}

// StartCommand starts a command.
func (c *Client) StartCommand(ctx context.Context, id string, opt *StartCommandRequest) error {
	request, err := c.newRPCRequest(ctx, http.MethodPost, fmt.Sprintf(startCommandEndpoint, url.PathEscape(id)), opt)
	if err != nil {
		return errors.WithStack(err)
	}
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

func (c *Client) newRPCRequest(ctx context.Context, method string, path string, body any) (*http.Request, error) {
	path, err := url.JoinPath(c.endpoint, path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		bodyReader = bytes.NewReader(payload)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	request, err := http.NewRequestWithContext(ctx, method, path, bodyReader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	request.Header.Add("Content-Type", "application/json")
	return request, nil
}
