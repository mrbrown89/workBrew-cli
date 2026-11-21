// Package workbrew is a small client for the Workbrew workspace API.
package workbrew

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiVersion     = "v0"
	defaultTimeout = 30 * time.Second
	// maxErrorExcerpt is how much of a non-2xx response body is included in the error.
	maxErrorExcerpt = 512
)

// Client talks to a single Workbrew workspace.
type Client struct {
	baseURL *url.URL
	token   string
	http    *http.Client
}

// New returns a client for the workspace API at baseURL, authenticating with token.
func New(baseURL, token string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace URL %q: %w", baseURL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid workspace URL %q: missing scheme or host", baseURL)
	}
	if token == "" {
		return nil, errors.New("API token is empty")
	}

	return &Client{
		baseURL: u,
		token:   token,
		http:    &http.Client{Timeout: defaultTimeout},
	}, nil
}

// getList fetches endpoint and decodes it as a JSON array of T.
// It never returns a nil slice on success.
func getList[T any](ctx context.Context, c *Client, endpoint string) ([]T, error) {
	var list []T
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []T{}
	}
	return list, nil
}

// Devices returns every device in the workspace.
func (c *Client) Devices(ctx context.Context) ([]Device, error) {
	return getList[Device](ctx, c, endpointDevices)
}

// Formulae returns every formula installed across the workspace.
func (c *Client) Formulae(ctx context.Context) ([]Formula, error) {
	return getList[Formula](ctx, c, endpointFormulae)
}

// Casks returns every cask installed across the workspace.
func (c *Client) Casks(ctx context.Context) ([]Cask, error) {
	return getList[Cask](ctx, c, endpointCasks)
}

// Vulnerabilities returns the workspace vulnerability report.
func (c *Client) Vulnerabilities(ctx context.Context) ([]VulnerabilityReport, error) {
	return getList[VulnerabilityReport](ctx, c, endpointVulnerabilities)
}

// Analytics returns brew command analytics for the workspace.
func (c *Client) Analytics(ctx context.Context) ([]AnalyticsItem, error) {
	return getList[AnalyticsItem](ctx, c, endpointAnalytics)
}

// Brewfiles returns the workspace Brewfiles.
func (c *Client) Brewfiles(ctx context.Context) ([]Brewfile, error) {
	return getList[Brewfile](ctx, c, endpointBrewfiles)
}

// BrewCommands returns the workspace brew commands.
func (c *Client) BrewCommands(ctx context.Context) ([]BrewCommand, error) {
	return getList[BrewCommand](ctx, c, endpointBrewCommands)
}

// CreateBrewCommand asks Workbrew to run `brew <arguments>` across the workspace's default scope.
func (c *Client) CreateBrewCommand(ctx context.Context, arguments string) (BrewCommandCreateResponse, error) {
	var result BrewCommandCreateResponse
	err := c.do(ctx, http.MethodPost, endpointBrewCommands, BrewCommandCreateRequest{Arguments: arguments}, &result)
	return result, err
}

// do sends a request to endpoint, JSON-encoding payload (if non-nil) as the body
// and decoding the response into target (if non-nil).
func (c *Client) do(ctx context.Context, method, endpoint string, payload, target any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL.JoinPath(endpoint).String(), body)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("X-Workbrew-API-Version", apiVersion)
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	// Nothing useful can be done if closing a response body fails.
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		excerpt, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorExcerpt))
		trimmed := strings.TrimSpace(string(excerpt))
		if trimmed == "" {
			return fmt.Errorf("workbrew API returned %s", response.Status)
		}
		return fmt.Errorf("workbrew API returned %s: %s", response.Status, trimmed)
	}

	if target == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding %s response: %w", endpoint, err)
	}
	return nil
}

// BaseURL returns the workspace API base URL the client was created with.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}
