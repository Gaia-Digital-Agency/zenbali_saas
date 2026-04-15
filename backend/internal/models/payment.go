package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID                    uuid.UUID `json:"id"`
	EventID               uuid.UUID `json:"event_id"`
	CreatorID             uuid.UUID `json:"creator_id"`
	StripeSessionID       string    `json:"stripe_session_id,omitempty"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id,omitempty"`
	AmountCents           int       `json:"amount_cents"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// Joined fields
	EventTitle  string `json:"event_title,omitempty"`
	CreatorName string `json:"creator_name,omitempty"`
}

const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

type PaymentResponse struct {
	ID          uuid.UUID `json:"id"`
	EventID     uuid.UUID `json:"event_id"`
	EventTitle  string    `json:"event_title"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaymentListResponse struct {
	Payments   []*Payment `json:"payments"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}

type CheckoutSessionResponse struct {
	SessionID  string `json:"session_id"`
	SessionURL string `json:"session_url"`
}

func (p *Payment) ToResponse() *PaymentResponse {
	return &PaymentResponse{
		ID:         p.ID,
		EventID:    p.EventID,
		EventTitle: p.EventTitle,
		Amount:     float64(p.AmountCents) / 100,
		Currency:   p.Currency,
		Status:     p.Status,
		CreatedAt:  p.CreatedAt,
	}
}
