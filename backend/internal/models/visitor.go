package models

import (
	"time"

	"github.com/google/uuid"
)

type Visitor struct {
	ID        uuid.UUID `json:"id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
	VisitedAt time.Time `json:"visited_at"`
}

type VisitorStats struct {
	TotalVisitors    int       `json:"total_visitors"`
	LastVisitorDate  time.Time `json:"last_visitor_date"`
	LastVisitorCity  string    `json:"last_visitor_city"`
	LastVisitorCountry string  `json:"last_visitor_country"`
}

type TrackVisitorRequest struct {
	UserAgent string `json:"user_agent"`
}
