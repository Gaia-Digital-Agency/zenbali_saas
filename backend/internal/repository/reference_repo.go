package repository

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/models"
)

// LocationRepository handles location data operations
type LocationRepository struct {
	pool *pgxpool.Pool
}

func NewLocationRepository(pool *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{pool: pool}
}

func (r *LocationRepository) List(ctx context.Context, onlyActive bool) ([]*models.Location, error) {
	query := `SELECT id, name, slug, is_active, created_at, updated_at FROM locations`
	if onlyActive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*models.Location
	for rows.Next() {
		loc := &models.Location{}
		if err := rows.Scan(&loc.ID, &loc.Name, &loc.Slug, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

func (r *LocationRepository) Create(ctx context.Context, loc *models.Location) error {
	loc.Slug = generateSlug(loc.Name)
	query := `INSERT INTO locations (name, slug) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, loc.Name, loc.Slug).Scan(&loc.ID, &loc.CreatedAt, &loc.UpdatedAt)
}

func (r *LocationRepository) Update(ctx context.Context, id int, name string, isActive bool) error {
	slug := generateSlug(name)
	query := `UPDATE locations SET name = $1, slug = $2, is_active = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, name, slug, isActive, id)
	return err
}

// EventTypeRepository handles event type data operations
type EventTypeRepository struct {
	pool *pgxpool.Pool
}

func NewEventTypeRepository(pool *pgxpool.Pool) *EventTypeRepository {
	return &EventTypeRepository{pool: pool}
}

func (r *EventTypeRepository) List(ctx context.Context, onlyActive bool) ([]*models.EventType, error) {
	query := `SELECT id, name, slug, is_active, created_at, updated_at FROM event_types`
	if onlyActive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*models.EventType
	for rows.Next() {
		t := &models.EventType{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func (r *EventTypeRepository) Create(ctx context.Context, et *models.EventType) error {
	et.Slug = generateSlug(et.Name)
	query := `INSERT INTO event_types (name, slug) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, et.Name, et.Slug).Scan(&et.ID, &et.CreatedAt, &et.UpdatedAt)
}

func (r *EventTypeRepository) Update(ctx context.Context, id int, name string, isActive bool) error {
	slug := generateSlug(name)
	query := `UPDATE event_types SET name = $1, slug = $2, is_active = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, name, slug, isActive, id)
	return err
}

// EntranceTypeRepository handles entrance type data operations
type EntranceTypeRepository struct {
	pool *pgxpool.Pool
}

func NewEntranceTypeRepository(pool *pgxpool.Pool) *EntranceTypeRepository {
	return &EntranceTypeRepository{pool: pool}
}

func (r *EntranceTypeRepository) List(ctx context.Context, onlyActive bool) ([]*models.EntranceType, error) {
	query := `SELECT id, name, slug, is_active, created_at, updated_at FROM entrance_types`
	if onlyActive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*models.EntranceType
	for rows.Next() {
		t := &models.EntranceType{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

// Helper function to generate URL-friendly slugs
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}
