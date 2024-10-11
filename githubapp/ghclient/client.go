package ghclient

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v64/github"
)

type Client struct {
	*github.Client
	httpClient *http.Client
}

func NewClient(client *github.Client) *Client {
	return &Client{Client: client, httpClient: &http.Client{}}
}

// SetDefaultHTTPClient sets the http.Client to be used by the client.
// It usually to mock the http.Client for testing.
func (h *Client) SetDefaultHTTPClient(c *http.Client) {
	h.httpClient = c
}

func (h *Client) AppInstallClient(appInstall int64) (*github.Client, error) {
	token, _, err := h.Client.Apps.CreateInstallationToken(context.Background(), appInstall, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to create install token for '%d'", appInstall)
	}
	// TODO: cache the token.
	return github.NewClient(h.httpClient).WithAuthToken(token.GetToken()), nil
}
