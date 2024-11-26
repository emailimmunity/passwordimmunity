package router

import (
	"net/http"

	"github.com/emailimmunity/passwordimmunity/api/handlers"
	"github.com/emailimmunity/passwordimmunity/api/middleware"
	"github.com/emailimmunity/passwordimmunity/services/payment"
	"github.com/gorilla/mux"
)

func RegisterPaymentRoutes(r *mux.Router, paymentService payment.PaymentService, webhookSecret string) {
	handler := handlers.NewPaymentHandler(paymentService)
	paymentMiddleware := middleware.NewPaymentMiddleware(webhookSecret)

	// Public routes
	r.HandleFunc("/api/v1/payments", handler.CreatePayment).Methods("POST")
	r.HandleFunc("/api/v1/payments", handler.GetPayment).Methods("GET")
	r.HandleFunc("/api/v1/payments/cancel", handler.CancelPayment).Methods("POST")

	// Webhook route with signature validation
	r.HandleFunc("/api/v1/payments/webhook",
		paymentMiddleware.ValidateWebhook(handler.HandleWebhook),
	).Methods("POST")
}
