package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/emailimmunity/passwordimmunity/services/payment"
    "github.com/emailimmunity/passwordimmunity/services/licensing"
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

type PaymentHandler struct {
    paymentService  payment.Service
    licenseService  licensing.Service
}

func NewPaymentHandler(ps payment.Service, ls licensing.Service) *PaymentHandler {
    return &PaymentHandler{
        paymentService:  ps,
        licenseService:  ls,
    }
}

type CreatePaymentRequest struct {
    LicenseType string `json:"license_type"`
    Period      string `json:"period"`
}

func (h *PaymentHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Post("/", h.CreatePayment)
    r.Post("/webhook", h.HandleWebhook)
    r.Get("/status/{id}", h.GetPaymentStatus)
    return r
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    var req CreatePaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    orgID := r.Context().Value("organization_id").(uuid.UUID)
    payment, err := h.paymentService.CreatePayment(r.Context(), orgID, req.LicenseType, req.Period)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }

    id := r.FormValue("id")
    status := r.FormValue("status")

    if err := h.paymentService.HandlePaymentWebhook(r.Context(), id, status); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        http.Error(w, "Invalid payment ID", http.StatusBadRequest)
        return
    }

    payment, err := h.paymentService.GetPayment(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if payment == nil {
        http.Error(w, "Payment not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(payment)
}
