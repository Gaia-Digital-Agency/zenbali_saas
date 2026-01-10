package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/models"
)

type CreatorRepository struct {
	pool *pgxpool.Pool
}

func NewCreatorRepository(pool *pgxpool.Pool) *CreatorRepository {
	return &CreatorRepository{pool: pool}
}

func (r *CreatorRepository) Create(ctx context.Context, creator *models.Creator) error {
	query := `
		INSERT INTO creators (name, organization_name, email, mobile, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		creator.Name,
		creator.OrganizationName,
		creator.Email,
		creator.Mobile,
		creator.PasswordHash,
	).Scan(&creator.ID, &creator.CreatedAt, &creator.UpdatedAt)
}

func (r *CreatorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Creator, error) {
	creator := &models.Creator{}
	query := `
		SELECT id, name, organization_name, email, mobile, password_hash, 
		       is_verified, is_active, created_at, updated_at
		FROM creators
		WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&creator.ID,
		&creator.Name,
		&creator.OrganizationName,
		&creator.Email,
		&creator.Mobile,
		&creator.PasswordHash,
		&creator.IsVerified,
		&creator.IsActive,
		&creator.CreatedAt,
		&creator.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return creator, nil
}

func (r *CreatorRepository) GetByEmail(ctx context.Context, email string) (*models.Creator, error) {
	creator := &models.Creator{}
	query := `
		SELECT id, name, organization_name, email, mobile, password_hash, 
		       is_verified, is_active, created_at, updated_at
		FROM creators
		WHERE email = $1
	`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&creator.ID,
		&creator.Name,
		&creator.OrganizationName,
		&creator.Email,
		&creator.Mobile,
		&creator.PasswordHash,
		&creator.IsVerified,
		&creator.IsActive,
		&creator.CreatedAt,
		&creator.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return creator, nil
}

func (r *CreatorRepository) Update(ctx context.Context, creator *models.Creator) error {
	query := `
		UPDATE creators
		SET name = $1, organization_name = $2, mobile = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.pool.Exec(ctx, query,
		creator.Name,
		creator.OrganizationName,
		creator.Mobile,
		creator.ID,
	)
	return err
}

func (r *CreatorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	query := `UPDATE creators SET is_active = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, isActive, id)
	return err
}

func (r *CreatorRepository) List(ctx context.Context, page, limit int) ([]*models.Creator, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM creators`
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get creators
	query := `
		SELECT id, name, organization_name, email, mobile, password_hash,
		       is_verified, is_active, created_at, updated_at
		FROM creators
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var creators []*models.Creator
	for rows.Next() {
		creator := &models.Creator{}
		if err := rows.Scan(
			&creator.ID,
			&creator.Name,
			&creator.OrganizationName,
			&creator.Email,
			&creator.Mobile,
			&creator.PasswordHash,
			&creator.IsVerified,
			&creator.IsActive,
			&creator.CreatedAt,
			&creator.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		creators = append(creators, creator)
	}

	return creators, total, nil
}

func (r *CreatorRepository) Count(ctx context.Context) (int, int, error) {
	var total, active int
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM creators
	`
	err := r.pool.QueryRow(ctx, query).Scan(&total, &active)
	return total, active, err
}
