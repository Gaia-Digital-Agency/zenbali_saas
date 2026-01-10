package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/models"
)

type VisitorRepository struct {
	pool *pgxpool.Pool
}

func NewVisitorRepository(pool *pgxpool.Pool) *VisitorRepository {
	return &VisitorRepository{pool: pool}
}

func (r *VisitorRepository) Create(ctx context.Context, visitor *models.Visitor) error {
	query := `
		INSERT INTO visitors (ip_address, user_agent, country, city)
		VALUES ($1, $2, $3, $4)
		RETURNING id, visited_at
	`
	return r.pool.QueryRow(ctx, query,
		visitor.IPAddress,
		visitor.UserAgent,
		visitor.Country,
		visitor.City,
	).Scan(&visitor.ID, &visitor.VisitedAt)
}

func (r *VisitorRepository) GetStats(ctx context.Context) (*models.VisitorStats, error) {
	stats := &models.VisitorStats{}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM visitors`
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&stats.TotalVisitors); err != nil {
		return nil, err
	}

	// Get last visitor info
	lastQuery := `
		SELECT visited_at, COALESCE(city, ''), COALESCE(country, '')
		FROM visitors
		ORDER BY visited_at DESC
		LIMIT 1
	`
	err := r.pool.QueryRow(ctx, lastQuery).Scan(
		&stats.LastVisitorDate,
		&stats.LastVisitorCity,
		&stats.LastVisitorCountry,
	)
	if err != nil {
		// No visitors yet, return default stats
		stats.LastVisitorDate = time.Now()
		return stats, nil
	}

	return stats, nil
}

func (r *VisitorRepository) GetTodayCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM visitors WHERE DATE(visited_at) = CURRENT_DATE`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *VisitorRepository) GetTotalCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM visitors`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}
