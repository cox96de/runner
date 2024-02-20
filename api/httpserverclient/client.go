package httpserverclient

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/lib"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"

	"github.com/pkg/errors"
)

var (
	_    api.ServerClient = (*Client)(nil)
	json                  = jsoniter.ConfigCompatibleWithStandardLibrary
)

func init() {
	json.RegisterExtension(&lib.ProtobufTypeExtension{})
}

type Client struct {
	client *http.Client
	u      *url.URL
}

func (c *Client) CreatePipeline(ctx context.Context, in *api.CreatePipelineRequest, opts ...grpc.CallOption) (*api.CreatePipelineResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) RequestJob(ctx context.Context, in *api.RequestJobRequest, opts ...grpc.CallOption) (*api.RequestJobResponse, error) {
	u := c.u.JoinPath("/api/v1/jobs/request")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, body)
		_ = body.Close()
	}(response.Body)
	if response.StatusCode == http.StatusNoContent {
		return &api.RequestJobResponse{}, nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to request job, got status code: %d", response.StatusCode)
	}
	resp := &api.RequestJobResponse{}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read response body")
	}
	err = json.Unmarshal(content, resp)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to unmarshal response body: %s", string(content))
	}
	return resp, nil
}

func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{client: client, u: u}, nil
}
