package httpserverclient

import (
	"bytes"
	"context"
	"fmt"
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

func (c *Client) UpdateStepExecution(ctx context.Context, in *api.UpdateStepExecutionRequest, opts ...grpc.CallOption) (*api.UpdateStepExecutionResponse, error) {
	u := c.u.JoinPath(fmt.Sprintf("/api/v1/jobs/%d/executions/%d/steps/%d", in.JobID, in.JobExecutionID,
		in.StepExecutionID))
	resp := &api.UpdateStepExecutionResponse{}
	err := c.doRequest(ctx, u.String(), http.MethodPost, in, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UploadLogLines(ctx context.Context, in *api.UpdateLogLinesRequest, opts ...grpc.CallOption) (*api.UpdateLogLinesResponse, error) {
	u := c.u.JoinPath(fmt.Sprintf("/api/v1/jobs/%d/executions/%d/logs", in.JobID, in.JobExecutionID))
	resp := &api.UpdateLogLinesResponse{}
	err := c.doRequest(ctx, u.String(), http.MethodPost, in, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetLogLines(ctx context.Context, in *api.GetLogLinesRequest, opts ...grpc.CallOption) (*api.GetLogLinesResponse, error) {
	url := fmt.Sprintf("/api/v1/jobs/%d/executions/%d/logs/%s", in.JobID, in.JobExecutionID, in.Name)

	u := c.u.JoinPath(url)
	query := u.Query()
	query.Add("offset", fmt.Sprintf("%d", in.Offset))
	if in.Limit != nil {
		query.Add("limit", fmt.Sprintf("%d", *in.Limit))
	}
	u.RawQuery = query.Encode()
	resp := &api.GetLogLinesResponse{}
	err := c.doRequest(ctx, u.String(), http.MethodGet, in, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateJobExecution(ctx context.Context, in *api.UpdateJobExecutionRequest,
	opts ...grpc.CallOption,
) (*api.UpdateJobExecutionResponse, error) {
	u := c.u.JoinPath(fmt.Sprintf("/api/v1/jobs/%d/executions/%d", in.JobID, in.JobExecutionID))
	resp := &api.UpdateJobExecutionResponse{}
	err := c.doRequest(ctx, u.String(), http.MethodPost, in, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) CreatePipeline(ctx context.Context, in *api.CreatePipelineRequest, opts ...grpc.CallOption) (*api.CreatePipelineResponse, error) {
	u := c.u.JoinPath("/api/v1/pipelines")
	resp := &api.CreatePipelineResponse{}
	err := c.doRequest(ctx, u.String(), http.MethodPost, in, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) RequestJob(ctx context.Context, in *api.RequestJobRequest, opts ...grpc.CallOption) (*api.RequestJobResponse, error) {
	u := c.u.JoinPath("/api/v1/jobs/request")
	resp := &api.RequestJobResponse{}
	err := c.doRequest(ctx, u.String(), http.MethodPost, in, resp)
	if err != nil {
		return nil, err
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

func (c *Client) doRequest(ctx context.Context, path string, method string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return errors.WithMessagef(err, "failed to marshal request body: %v", in)
	}
	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, body)
		_ = body.Close()
	}(response.Body)
	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	if response.StatusCode != http.StatusOK {
		return errors.Errorf("failed to request job, got status code: %d", response.StatusCode)
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.WithMessage(err, "failed to read response body")
	}
	err = json.Unmarshal(content, out)
	if err != nil {
		return errors.WithMessagef(err, "failed to unmarshal response body: %s", string(content))
	}
	return nil
}
