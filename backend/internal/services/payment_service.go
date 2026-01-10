package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrAlreadyPaid     = errors.New("event already paid")
)

type PaymentService struct {
	repos  *repository.Repositories
	config config.StripeConfig
}

func NewPaymentService(repos *repository.Repositories, cfg config.StripeConfig) *PaymentService {
	return &PaymentService{
		repos:  repos,
		config: cfg,
	}
}

func (s *PaymentService) CreateCheckoutSession(ctx context.Context, event *models.Event, successURL, cancelURL string) (*models.CheckoutSessionResponse, error) {
	// Check if already paid
	if event.IsPaid {
		return nil, ErrAlreadyPaid
	}

	// Create Stripe checkout session
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Event Posting Fee - " + event.Title),
						Description: stripe.String("Zen Bali event posting fee"),
					},
					UnitAmount: stripe.Int64(s.config.PriceCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"event_id":   event.ID.String(),
			"creator_id": event.CreatorID.String(),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}

	// Create payment record
	payment := &models.Payment{
		EventID:         event.ID,
		CreatorID:       event.CreatorID,
		StripeSessionID: sess.ID,
		AmountCents:     int(s.config.PriceCents),
		Currency:        "USD",
		Status:          models.PaymentStatusPending,
	}

	if err := s.repos.Payment.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &models.CheckoutSessionResponse{
		SessionID:  sess.ID,
		SessionURL: sess.URL,
	}, nil
}

func (s *PaymentService) HandleSuccessfulPayment(ctx context.Context, sessionID string) error {
	// Get payment by session ID
	payment, err := s.repos.Payment.GetByStripeSessionID(ctx, sessionID)
	if err != nil {
		return err
	}
	if payment == nil {
		return ErrPaymentNotFound
	}

	// Get Stripe session to get payment intent ID
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return err
	}

	// Update payment status
	if err := s.repos.Payment.UpdateStatus(ctx, payment.ID, models.PaymentStatusCompleted, sess.PaymentIntent.ID); err != nil {
		return err
	}

	// Publish the event
	if err := s.repos.Event.UpdatePaymentStatus(ctx, payment.EventID, true, true); err != nil {
		return err
	}

	return nil
}

func (s *PaymentService) HandleFailedPayment(ctx context.Context, sessionID string) error {
	payment, err := s.repos.Payment.GetByStripeSessionID(ctx, sessionID)
	if err != nil {
		return err
	}
	if payment == nil {
		return ErrPaymentNotFound
	}

	return s.repos.Payment.UpdateStatus(ctx, payment.ID, models.PaymentStatusFailed, "")
}

func (s *PaymentService) ListByCreator(ctx context.Context, creatorID uuid.UUID, page, limit int) (*models.PaymentListResponse, error) {
	payments, total, err := s.repos.Payment.ListByCreator(ctx, creatorID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit

	return &models.PaymentListResponse{
		Payments:   payments,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *PaymentService) ListAll(ctx context.Context, page, limit int, status string) (*models.PaymentListResponse, error) {
	payments, total, err := s.repos.Payment.ListAll(ctx, page, limit, status)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit

	return &models.PaymentListResponse{
		Payments:   payments,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *PaymentService) GetWebhookSecret() string {
	return s.config.WebhookSecret
}
