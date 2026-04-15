package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories holds all repository instances
type Repositories struct {
	Creator      *CreatorRepository
	Event        *EventRepository
	Payment      *PaymentRepository
	Admin        *AdminRepository
	Location     *LocationRepository
	EventType    *EventTypeRepository
	EntranceType *EntranceTypeRepository
	Visitor      *VisitorRepository
}

// BaseRepository provides common database functionality
type BaseRepository struct {
	pool *pgxpool.Pool
}
