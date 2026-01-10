package models

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AdminResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *Admin) ToResponse() *AdminResponse {
	return &AdminResponse{
		ID:        a.ID,
		Email:     a.Email,
		Name:      a.Name,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
	}
}

type DashboardStats struct {
	TotalEvents      int     `json:"total_events"`
	PublishedEvents  int     `json:"published_events"`
	UpcomingEvents   int     `json:"upcoming_events"`
	TotalCreators    int     `json:"total_creators"`
	ActiveCreators   int     `json:"active_creators"`
	TotalPayments    int     `json:"total_payments"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalVisitors    int     `json:"total_visitors"`
	TodayVisitors    int     `json:"today_visitors"`
	RecentEvents     []*Event `json:"recent_events"`
	RecentPayments   []*Payment `json:"recent_payments"`
}
