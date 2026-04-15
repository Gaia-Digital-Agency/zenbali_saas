package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/net1io/zenbali/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Database struct {
	Pool *pgxpool.Pool
}

func New(cfg config.DatabaseConfig) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")

	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
	log.Println("Database connection closed")
}

func (db *Database) RunMigrations() error {
	ctx := context.Background()

	// Create migrations tracking table
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	rows, err := db.Pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		applied[version] = true
	}

	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort .up.sql files
	var upMigrations []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	sort.Strings(upMigrations)

	// Apply pending migrations
	for _, filename := range upMigrations {
		version := strings.TrimSuffix(filename, ".up.sql")
		if applied[version] {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		log.Printf("Applying migration: %s", version)

		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("✅ Applied migration: %s", version)
	}

	log.Println("✅ All migrations applied")
	return nil
}
