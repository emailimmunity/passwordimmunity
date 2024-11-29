package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	client    *http.Client
	baseURL   string
	apiKey    string
	maxRetries int
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s", e.StatusCode, e.Message)
}

func NewHTTPClient(apiKey string, baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		apiKey:     apiKey,
		maxRetries: 3,
	}
}

func (c *HTTPClient) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.client.Do(req)
		if err != nil {
			continue // Retry on network errors
		}

		// Don't retry on client errors (4xx)
		if resp.StatusCode < 500 {
			break
		}

		resp.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("all retries failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var errResp struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    "unknown error",
			}
		}
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Message,
		}
	}

	return resp, nil
}
