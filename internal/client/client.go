// Package client is a minimal Fireworks AI control-plane API client.
// API reference: https://docs.fireworks.ai/api-reference/introduction
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.fireworks.ai"

type Client struct {
	baseURL   string
	accountID string
	apiKey    string
	http      *http.Client
}

func New(baseURL, accountID, apiKey string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		accountID: accountID,
		apiKey:    apiKey,
		http:      &http.Client{Timeout: 60 * time.Second},
	}
}

// APIError is returned for any non-2xx response.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("fireworks API error (HTTP %d): %s", e.StatusCode, e.Body)
}

func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// ResourceID returns the final segment of a full resource name such as
// "accounts/my-account/deployments/abc123".
func ResourceID(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}
