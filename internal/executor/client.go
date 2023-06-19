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

	internalmodel "github.com/cox96de/runner/internal/model"

	"github.com/pkg/errors"
)

const (
	pingEndpoint           = "/ping"
	startCommandEndpoint   = "/commands/%s"
	getCommandLogsEndpoint = "/commands/%s/logs"
)

const (
	// defaultAPITimeout is the default timeout for simple json RPC calls.
	defaultAPITimeout = time.Second * 10
	// defaultStreamTimeout is the default timeout for stream RPC calls.
	defaultStreamTimeout = time.Hour

	limitReaderSize = 1024
)

type Client struct {
	c        *http.Client
	endpoint string
}

func NewClient(endpoint string) *Client {
	c := &Client{
		endpoint: endpoint,
		// This http client don't assigned a timeout. Timeout is assigned in each request.
		// Different requests may have different timeout.
		c: &http.Client{},
	}
	return c
}

// Ping checks the executor is bootstrapped and ready to serve.
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()
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
// `id` is the unique id of the command, and it's unique in the executor.
// Use that id to get logs and status.
func (c *Client) StartCommand(ctx context.Context, id string, opt *StartCommandRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()
	request, err := c.newRPCRequest(ctx, http.MethodPost, fmt.Sprintf(startCommandEndpoint, url.PathEscape(id)), opt)
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
		body, _ := io.ReadAll(io.LimitReader(response.Body, limitReaderSize))
		return errors.Errorf("invalid status code: %d with payload: %s", response.StatusCode, string(body))
	}
	return nil
}

// GetCommandLogs gets command logs.
// It returns a reader that will be closed when the context is done or the logs are finished.
func (c *Client) GetCommandLogs(ctx context.Context, id string) io.ReadCloser {
	reader, writer := io.Pipe()
	// Combines multiple requests into one stream.
	// Omit io.EOF in response body.
	go func() {
		bs := 1024
		buf := make([]byte, bs)
		_ = buf
		for {
			select {
			case <-ctx.Done():
				_ = writer.CloseWithError(ctx.Err())
				return
			default:
			}
			// TODO: cancel called too many times, refactor it
			ctx, cancel := context.WithTimeout(ctx, defaultStreamTimeout)
			request, err := c.newRPCRequest(ctx, http.MethodGet, fmt.Sprintf(getCommandLogsEndpoint, url.PathEscape(id)),
				nil)
			if err != nil {
				_ = writer.CloseWithError(errors.WithStack(err))
				cancel()
				return
			}
			request.Header.Add("Content-Type", "application/octet-stream")
			response, err := c.c.Do(request)
			if err != nil {
				_ = writer.CloseWithError(errors.WithStack(err))
				cancel()
				return
			}
			// TODO: drain response body if not 200 ?
			if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusPartialContent {
				// Ignore too much payload.
				body, _ := io.ReadAll(io.LimitReader(response.Body, limitReaderSize))
				err := errors.Errorf("invalid status code: %d with payload: %s", response.StatusCode, string(body))
				_ = writer.CloseWithError(errors.WithStack(err))
				cancel()
				return
			}
			for {
				n, readErr := response.Body.Read(buf)
				if _, err = writer.Write(buf[:n]); err != nil {
					_ = writer.CloseWithError(errors.WithStack(err))
					cancel()
					return
				}
				if readErr != nil && readErr == io.EOF {
					break
				}
			}
			cancel()
			if response.StatusCode == http.StatusOK {
				_ = writer.Close()
				return
			}
		}
	}()
	return reader
}

func (c *Client) GetCommandStatus(ctx context.Context, id string) (*internalmodel.GetCommandStatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultAPITimeout)
	defer cancel()
	request, err := c.newRPCRequest(ctx, http.MethodGet, fmt.Sprintf(startCommandEndpoint, url.PathEscape(id)),
		nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := c.c.Do(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		// Ignore too much payload.
		body, _ := io.ReadAll(io.LimitReader(response.Body, limitReaderSize))
		return nil, errors.Errorf("invalid status code: %d with payload: %s", response.StatusCode, string(body))
	}
	resp := &internalmodel.GetCommandStatusResponse{}
	all, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = json.Unmarshal(all, resp)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
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
	return request, nil
}
