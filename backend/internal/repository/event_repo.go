package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/models"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (
			creator_id, title, event_date, event_time, location_id, event_type_id,
			duration, entrance_type_id, entrance_fee, contact_email, contact_mobile, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		event.CreatorID,
		event.Title,
		event.EventDate,
		event.EventTime,
		event.LocationID,
		event.EventTypeID,
		event.Duration,
		event.EntranceTypeID,
		event.EntranceFee,
		event.ContactEmail,
		event.ContactMobile,
		event.Notes,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	event := &models.Event{}
	query := `
		SELECT 
			e.id, e.creator_id, e.title, e.event_date, e.event_time, e.location_id,
			e.event_type_id, e.duration, e.entrance_type_id, e.entrance_fee,
			e.contact_email, e.contact_mobile, e.notes, e.image_url,
			e.is_paid, e.is_published, e.created_at, e.updated_at,
			c.name as creator_name, c.organization_name,
			l.name as location_name, et.name as event_type_name, ent.name as entrance_type_name
		FROM events e
		JOIN creators c ON e.creator_id = c.id
		JOIN locations l ON e.location_id = l.id
		JOIN event_types et ON e.event_type_id = et.id
		JOIN entrance_types ent ON e.entrance_type_id = ent.id
		WHERE e.id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.CreatorID, &event.Title, &event.EventDate, &event.EventTime,
		&event.LocationID, &event.EventTypeID, &event.Duration, &event.EntranceTypeID,
		&event.EntranceFee, &event.ContactEmail, &event.ContactMobile, &event.Notes,
		&event.ImageURL, &event.IsPaid, &event.IsPublished, &event.CreatedAt, &event.UpdatedAt,
		&event.CreatorName, &event.OrganizationName, &event.LocationName,
		&event.EventTypeName, &event.EntranceTypeName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (r *EventRepository) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events
		SET title = $1, event_date = $2, event_time = $3, location_id = $4,
		    event_type_id = $5, duration = $6, entrance_type_id = $7, entrance_fee = $8,
		    contact_email = $9, contact_mobile = $10, notes = $11, updated_at = NOW()
		WHERE id = $12
	`
	_, err := r.pool.Exec(ctx, query,
		event.Title, event.EventDate, event.EventTime, event.LocationID,
		event.EventTypeID, event.Duration, event.EntranceTypeID, event.EntranceFee,
		event.ContactEmail, event.ContactMobile, event.Notes, event.ID,
	)
	return err
}

func (r *EventRepository) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL string) error {
	query := `UPDATE events SET image_url = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, imageURL, id)
	return err
}

func (r *EventRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, isPaid, isPublished bool) error {
	query := `UPDATE events SET is_paid = $1, is_published = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, isPaid, isPublished, id)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *EventRepository) List(ctx context.Context, filter models.EventListFilter) ([]*models.Event, int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	// Base query
	baseQuery := `
		FROM events e
		JOIN creators c ON e.creator_id = c.id
		JOIN locations l ON e.location_id = l.id
		JOIN event_types et ON e.event_type_id = et.id
		JOIN entrance_types ent ON e.entrance_type_id = ent.id
		WHERE 1=1
	`

	// Apply filters
	if filter.OnlyPublished {
		conditions = append(conditions, "e.is_published = true")
	}

	if !filter.IncludePast {
		conditions = append(conditions, fmt.Sprintf("e.event_date >= $%d", argNum))
		args = append(args, time.Now().Format("2006-01-02"))
		argNum++
	}

	if filter.LocationID > 0 {
		conditions = append(conditions, fmt.Sprintf("e.location_id = $%d", argNum))
		args = append(args, filter.LocationID)
		argNum++
	}

	if filter.EventTypeID > 0 {
		conditions = append(conditions, fmt.Sprintf("e.event_type_id = $%d", argNum))
		args = append(args, filter.EventTypeID)
		argNum++
	}

	if filter.EntranceTypeID > 0 {
		conditions = append(conditions, fmt.Sprintf("e.entrance_type_id = $%d", argNum))
		args = append(args, filter.EntranceTypeID)
		argNum++
	}

	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("e.event_date >= $%d", argNum))
		args = append(args, filter.DateFrom.Format("2006-01-02"))
		argNum++
	}

	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("e.event_date <= $%d", argNum))
		args = append(args, filter.DateTo.Format("2006-01-02"))
		argNum++
	}

	if filter.CreatorID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("e.creator_id = $%d", argNum))
		args = append(args, filter.CreatorID)
		argNum++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.title ILIKE $%d OR c.name ILIKE $%d OR c.organization_name ILIKE $%d OR e.notes ILIKE $%d)", argNum, argNum, argNum, argNum))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	// Build WHERE clause
	whereClause := baseQuery
	if len(conditions) > 0 {
		whereClause += " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Set defaults
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	// Data query
	selectQuery := `
		SELECT 
			e.id, e.creator_id, e.title, e.event_date, e.event_time, e.location_id,
			e.event_type_id, e.duration, e.entrance_type_id, e.entrance_fee,
			e.contact_email, e.contact_mobile, e.notes, e.image_url,
			e.is_paid, e.is_published, e.created_at, e.updated_at,
			c.name as creator_name, c.organization_name,
			l.name as location_name, et.name as event_type_name, ent.name as entrance_type_name
	` + whereClause + fmt.Sprintf(" ORDER BY e.event_date ASC, e.created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)

	args = append(args, filter.Limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		if err := rows.Scan(
			&event.ID, &event.CreatorID, &event.Title, &event.EventDate, &event.EventTime,
			&event.LocationID, &event.EventTypeID, &event.Duration, &event.EntranceTypeID,
			&event.EntranceFee, &event.ContactEmail, &event.ContactMobile, &event.Notes,
			&event.ImageURL, &event.IsPaid, &event.IsPublished, &event.CreatedAt, &event.UpdatedAt,
			&event.CreatorName, &event.OrganizationName, &event.LocationName,
			&event.EventTypeName, &event.EntranceTypeName,
		); err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	return events, total, nil
}

func (r *EventRepository) Count(ctx context.Context) (total, published, upcoming int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_published = true) as published,
			COUNT(*) FILTER (WHERE event_date >= CURRENT_DATE AND is_published = true) as upcoming
		FROM events
	`
	err = r.pool.QueryRow(ctx, query).Scan(&total, &published, &upcoming)
	return
}

func (r *EventRepository) GetRecent(ctx context.Context, limit int) ([]*models.Event, error) {
	query := `
		SELECT 
			e.id, e.creator_id, e.title, e.event_date, e.event_time, e.location_id,
			e.event_type_id, e.duration, e.entrance_type_id, e.entrance_fee,
			e.contact_email, e.contact_mobile, e.notes, e.image_url,
			e.is_paid, e.is_published, e.created_at, e.updated_at,
			c.name as creator_name, c.organization_name,
			l.name as location_name, et.name as event_type_name, ent.name as entrance_type_name
		FROM events e
		JOIN creators c ON e.creator_id = c.id
		JOIN locations l ON e.location_id = l.id
		JOIN event_types et ON e.event_type_id = et.id
		JOIN entrance_types ent ON e.entrance_type_id = ent.id
		ORDER BY e.created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		if err := rows.Scan(
			&event.ID, &event.CreatorID, &event.Title, &event.EventDate, &event.EventTime,
			&event.LocationID, &event.EventTypeID, &event.Duration, &event.EntranceTypeID,
			&event.EntranceFee, &event.ContactEmail, &event.ContactMobile, &event.Notes,
			&event.ImageURL, &event.IsPaid, &event.IsPublished, &event.CreatedAt, &event.UpdatedAt,
			&event.CreatorName, &event.OrganizationName, &event.LocationName,
			&event.EventTypeName, &event.EntranceTypeName,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
