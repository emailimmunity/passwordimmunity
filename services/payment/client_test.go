package payment

import (
    "bytes"
    "errors"
    "io"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestNewMollieClient(t *testing.T) {
    tests := []struct {
        name          string
        config        *MollieConfig
        expectedError string
    }{
        {
            name: "valid config",
            config: &MollieConfig{
                APIKey:     "live_test_key",
                TestAPIKey: "test_key",
                IsTestMode: false,
            },
        },
        {
            name:          "nil config",
            config:        nil,
            expectedError: "mollie configuration is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, err := NewMollieClient(tt.config)

            if tt.expectedError != "" {
                assert.EqualError(t, err, tt.expectedError)
                assert.Nil(t, client)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, client)
            }
        })
    }
}

func TestMollieClientCreatePayment(t *testing.T) {
    config := &MollieConfig{
        APIKey:     "test_key",
        IsTestMode: true,
    }

    client, err := NewMollieClient(config)
    assert.NoError(t, err)

    mockClient := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            // Verify request headers and body
            assert.Equal(t, "Bearer test_key", req.Header.Get("Authorization"))

            // Return mock response
            return &http.Response{
                StatusCode: http.StatusOK,
                Body: io.NopCloser(bytes.NewBufferString(`{
                    "id": "tr_test123",
                    "status": "open",
                    "amount": {"currency": "EUR", "value": "10.00"}
                }`)),
            }, nil
        },
    }

    mollieClientImpl := client.(*mollieClient)
    mollieClientImpl.WithHTTPClient(mockClient)

    payment, err := client.CreatePayment("10.00", "EUR", "Test payment")
    assert.NoError(t, err)
    assert.Equal(t, "tr_test123", payment.ID)
}

func TestMollieClientGetPayment(t *testing.T) {
    config := &MollieConfig{
        APIKey:     "test_key",
        IsTestMode: true,
    }

    client, err := NewMollieClient(config)
    assert.NoError(t, err)

    mockClient := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            assert.Equal(t, "Bearer test_key", req.Header.Get("Authorization"))
            assert.Contains(t, req.URL.Path, "tr_test123")

            return &http.Response{
                StatusCode: http.StatusOK,
                Body: io.NopCloser(bytes.NewBufferString(`{
                    "id": "tr_test123",
                    "status": "paid",
                    "amount": {"currency": "EUR", "value": "10.00"}
                }`)),
            }, nil
        },
    }

    mollieClientImpl := client.(*mollieClient)
    mollieClientImpl.WithHTTPClient(mockClient)

    payment, err := client.GetPayment("tr_test123")
    assert.NoError(t, err)
    assert.Equal(t, "tr_test123", payment.ID)
    assert.Equal(t, "paid", payment.Status)
}

func TestMollieClientErrors(t *testing.T) {
    config := &MollieConfig{
        APIKey:     "test_key",
        IsTestMode: true,
    }

    client, err := NewMollieClient(config)
    assert.NoError(t, err)

    t.Run("create payment error", func(t *testing.T) {
        mockClient := &mockHTTPClient{
            DoFunc: func(req *http.Request) (*http.Response, error) {
                return nil, errors.New("network error")
            },
        }

        mollieClientImpl := client.(*mollieClient)
        mollieClientImpl.WithHTTPClient(mockClient)

        _, err := client.CreatePayment("10.00", "EUR", "Test payment")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to create Mollie payment")
    })

    t.Run("get payment error", func(t *testing.T) {
        mockClient := &mockHTTPClient{
            DoFunc: func(req *http.Request) (*http.Response, error) {
                return nil, errors.New("network error")
            },
        }

        mollieClientImpl := client.(*mollieClient)
        mollieClientImpl.WithHTTPClient(mockClient)

        _, err := client.GetPayment("tr_test123")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to get Mollie payment")
    })
}

func TestMollieClientGetPayment(t *testing.T) {
    config := &MollieConfig{
        APIKey:     "test_key",
        IsTestMode: true,
    }

    client, err := NewMollieClient(config)
    assert.NoError(t, err)

    mockClient := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            assert.Equal(t, "Bearer test_key", req.Header.Get("Authorization"))
            assert.Contains(t, req.URL.Path, "tr_test123")

            return &http.Response{
                StatusCode: http.StatusOK,
                Body: io.NopCloser(bytes.NewBufferString(`{
                    "id": "tr_test123",
                    "status": "paid",
                    "amount": {"currency": "EUR", "value": "10.00"}
                }`)),
            }, nil
        },
    }

    mollieClientImpl := client.(*mollieClient)
    mollieClientImpl.WithHTTPClient(mockClient)

    payment, err := client.GetPayment("tr_test123")
    assert.NoError(t, err)
    assert.Equal(t, "tr_test123", payment.ID)
    assert.Equal(t, "paid", payment.Status)
}
