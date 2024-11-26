package payment

import (
    "net/http"
)

// HTTPClient interface
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// mockHTTPClient implements HTTPClient for testing
type mockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    if m.DoFunc != nil {
        return m.DoFunc(req)
    }
    return nil, nil
}

// WithHTTPClient allows injection of custom HTTP client
func (m *mollieClient) WithHTTPClient(client HTTPClient) {
    if client != nil {
        m.client.WithHTTPClient(client)
    }
}
