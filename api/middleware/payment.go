package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

type PaymentMiddleware struct {
	webhookSecret string
}

func NewPaymentMiddleware(webhookSecret string) *PaymentMiddleware {
	return &PaymentMiddleware{
		webhookSecret: webhookSecret,
	}
}

func (m *PaymentMiddleware) ValidateWebhook(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("Mollie-Signature")
		if signature == "" {
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}

		body := make([]byte, r.ContentLength)
		if _, err := r.Body.Read(body); err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		mac := hmac.New(sha256.New, []byte(m.webhookSecret))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
