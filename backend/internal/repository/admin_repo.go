package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/models"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) Create(ctx context.Context, admin *models.Admin) error {
	query := `
		INSERT INTO admins (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		admin.Email,
		admin.PasswordHash,
		admin.Name,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
}

func (r *AdminRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	admin := &models.Admin{}
	query := `
		SELECT id, email, password_hash, name, is_active, created_at, updated_at
		FROM admins
		WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Name,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (*models.Admin, error) {
	admin := &models.Admin{}
	query := `
		SELECT id, email, password_hash, name, is_active, created_at, updated_at
		FROM admins
		WHERE email = $1
	`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Name,
		&admin.IsActive,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (r *AdminRepository) EnsureDefaultAdmin(ctx context.Context, email, passwordHash string) error {
	// Check if admin exists
	existing, err := r.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // Admin already exists
	}

	// Create default admin
	admin := &models.Admin{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         "Admin",
	}
	return r.Create(ctx, admin)
}
