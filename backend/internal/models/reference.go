package models

import "time"

// Location represents a Bali area/region
type Location struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventType represents a type of event
type EventType struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EntranceType represents how attendees can enter an event
type EntranceType struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LocationRequest for creating/updating locations
type LocationRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	IsActive *bool  `json:"is_active"`
}

// EventTypeRequest for creating/updating event types
type EventTypeRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	IsActive *bool  `json:"is_active"`
}

// EntranceTypeRequest for creating/updating entrance types
type EntranceTypeRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	IsActive *bool  `json:"is_active"`
}
