package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/models"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	query := `
		INSERT INTO payments (event_id, creator_id, stripe_session_id, amount_cents, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		payment.EventID,
		payment.CreatorID,
		payment.StripeSessionID,
		payment.AmountCents,
		payment.Currency,
		payment.Status,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	payment := &models.Payment{}
	query := `
		SELECT p.id, p.event_id, p.creator_id, p.stripe_session_id, p.stripe_payment_intent_id,
		       p.amount_cents, p.currency, p.status, p.created_at, p.updated_at,
		       e.title as event_title, c.name as creator_name
		FROM payments p
		JOIN events e ON p.event_id = e.id
		JOIN creators c ON p.creator_id = c.id
		WHERE p.id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&payment.ID, &payment.EventID, &payment.CreatorID, &payment.StripeSessionID,
		&payment.StripePaymentIntentID, &payment.AmountCents, &payment.Currency,
		&payment.Status, &payment.CreatedAt, &payment.UpdatedAt,
		&payment.EventTitle, &payment.CreatorName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *PaymentRepository) GetByStripeSessionID(ctx context.Context, sessionID string) (*models.Payment, error) {
	payment := &models.Payment{}
	query := `
		SELECT p.id, p.event_id, p.creator_id, p.stripe_session_id, p.stripe_payment_intent_id,
		       p.amount_cents, p.currency, p.status, p.created_at, p.updated_at
		FROM payments p
		WHERE p.stripe_session_id = $1
	`
	err := r.pool.QueryRow(ctx, query, sessionID).Scan(
		&payment.ID, &payment.EventID, &payment.CreatorID, &payment.StripeSessionID,
		&payment.StripePaymentIntentID, &payment.AmountCents, &payment.Currency,
		&payment.Status, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, paymentIntentID string) error {
	query := `
		UPDATE payments 
		SET status = $1, stripe_payment_intent_id = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.pool.Exec(ctx, query, status, paymentIntentID, id)
	return err
}

func (r *PaymentRepository) ListByCreator(ctx context.Context, creatorID uuid.UUID, page, limit int) ([]*models.Payment, int, error) {
	offset := (page - 1) * limit

	// Count
	var total int
	countQuery := `SELECT COUNT(*) FROM payments WHERE creator_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, creatorID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data
	query := `
		SELECT p.id, p.event_id, p.creator_id, p.stripe_session_id, p.stripe_payment_intent_id,
		       p.amount_cents, p.currency, p.status, p.created_at, p.updated_at,
		       e.title as event_title
		FROM payments p
		JOIN events e ON p.event_id = e.id
		WHERE p.creator_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, creatorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		payment := &models.Payment{}
		if err := rows.Scan(
			&payment.ID, &payment.EventID, &payment.CreatorID, &payment.StripeSessionID,
			&payment.StripePaymentIntentID, &payment.AmountCents, &payment.Currency,
			&payment.Status, &payment.CreatedAt, &payment.UpdatedAt, &payment.EventTitle,
		); err != nil {
			return nil, 0, err
		}
		payments = append(payments, payment)
	}
	return payments, total, nil
}

func (r *PaymentRepository) ListAll(ctx context.Context, page, limit int, status string) ([]*models.Payment, int, error) {
	offset := (page - 1) * limit
	var conditions []string
	var args []interface{}
	argNum := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("p.status = $%d", argNum))
		args = append(args, status)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM payments p " + whereClause
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data
	query := fmt.Sprintf(`
		SELECT p.id, p.event_id, p.creator_id, p.stripe_session_id, p.stripe_payment_intent_id,
		       p.amount_cents, p.currency, p.status, p.created_at, p.updated_at,
		       e.title as event_title, c.name as creator_name
		FROM payments p
		JOIN events e ON p.event_id = e.id
		JOIN creators c ON p.creator_id = c.id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		payment := &models.Payment{}
		if err := rows.Scan(
			&payment.ID, &payment.EventID, &payment.CreatorID, &payment.StripeSessionID,
			&payment.StripePaymentIntentID, &payment.AmountCents, &payment.Currency,
			&payment.Status, &payment.CreatedAt, &payment.UpdatedAt,
			&payment.EventTitle, &payment.CreatorName,
		); err != nil {
			return nil, 0, err
		}
		payments = append(payments, payment)
	}
	return payments, total, nil
}

func (r *PaymentRepository) GetStats(ctx context.Context) (int, float64, error) {
	var count int
	var totalCents int64
	query := `
		SELECT COUNT(*), COALESCE(SUM(amount_cents), 0)
		FROM payments
		WHERE status = 'completed'
	`
	if err := r.pool.QueryRow(ctx, query).Scan(&count, &totalCents); err != nil {
		return 0, 0, err
	}
	return count, float64(totalCents) / 100, nil
}

func (r *PaymentRepository) GetRecent(ctx context.Context, limit int) ([]*models.Payment, error) {
	query := `
		SELECT p.id, p.event_id, p.creator_id, p.stripe_session_id, p.stripe_payment_intent_id,
		       p.amount_cents, p.currency, p.status, p.created_at, p.updated_at,
		       e.title as event_title, c.name as creator_name
		FROM payments p
		JOIN events e ON p.event_id = e.id
		JOIN creators c ON p.creator_id = c.id
		ORDER BY p.created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		payment := &models.Payment{}
		if err := rows.Scan(
			&payment.ID, &payment.EventID, &payment.CreatorID, &payment.StripeSessionID,
			&payment.StripePaymentIntentID, &payment.AmountCents, &payment.Currency,
			&payment.Status, &payment.CreatedAt, &payment.UpdatedAt,
			&payment.EventTitle, &payment.CreatorName,
		); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}
