package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID             uuid.UUID `json:"id"`
	CreatorID      uuid.UUID `json:"creator_id"`
	Title          string    `json:"title"`
	EventDate      time.Time `json:"event_date"`
	EventTime      *string   `json:"event_time,omitempty"`
	LocationID     int       `json:"location_id"`
	EventTypeID    int       `json:"event_type_id"`
	Duration       string    `json:"duration"`
	EntranceTypeID int       `json:"entrance_type_id"`
	EntranceFee    float64   `json:"entrance_fee"`
	ContactEmail   string    `json:"contact_email"`
	ContactMobile  string    `json:"contact_mobile"`
	Notes          string    `json:"notes"`
	ImageURL       string    `json:"image_url"`
	IsPaid         bool      `json:"is_paid"`
	IsPublished    bool      `json:"is_published"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Joined fields
	CreatorName      string `json:"creator_name,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	LocationName     string `json:"location_name,omitempty"`
	EventTypeName    string `json:"event_type_name,omitempty"`
	EntranceTypeName string `json:"entrance_type_name,omitempty"`
}

type EventCreateRequest struct {
	Title          string  `json:"title" validate:"required,min=3,max=255"`
	EventDate      string  `json:"event_date" validate:"required"`
	EventTime      string  `json:"event_time"`
	LocationID     int     `json:"location_id" validate:"required,min=1"`
	EventTypeID    int     `json:"event_type_id" validate:"required,min=1"`
	Duration       string  `json:"duration" validate:"max=100"`
	EntranceTypeID int     `json:"entrance_type_id" validate:"required,min=1"`
	EntranceFee    float64 `json:"entrance_fee" validate:"min=0"`
	ContactEmail   string  `json:"contact_email" validate:"required,email"`
	ContactMobile  string  `json:"contact_mobile" validate:"max=50"`
	Notes          string  `json:"notes" validate:"max=2000"`
}

type EventUpdateRequest struct {
	Title          string  `json:"title" validate:"min=3,max=255"`
	EventDate      string  `json:"event_date"`
	EventTime      string  `json:"event_time"`
	LocationID     int     `json:"location_id" validate:"min=1"`
	EventTypeID    int     `json:"event_type_id" validate:"min=1"`
	Duration       string  `json:"duration" validate:"max=100"`
	EntranceTypeID int     `json:"entrance_type_id" validate:"min=1"`
	EntranceFee    float64 `json:"entrance_fee" validate:"min=0"`
	ContactEmail   string  `json:"contact_email" validate:"email"`
	ContactMobile  string  `json:"contact_mobile" validate:"max=50"`
	Notes          string  `json:"notes" validate:"max=2000"`
}

type EventListFilter struct {
	LocationID     int       `json:"location_id"`
	EventTypeID    int       `json:"event_type_id"`
	EntranceTypeID int       `json:"entrance_type_id"`
	DateFrom       time.Time `json:"date_from"`
	DateTo         time.Time `json:"date_to"`
	Search         string    `json:"search"`
	CreatorID      uuid.UUID `json:"creator_id"`
	IncludePast    bool      `json:"include_past"`
	OnlyPublished  bool      `json:"only_published"`
	Page           int       `json:"page"`
	Limit          int       `json:"limit"`
}

type EventListResponse struct {
	Events     []*Event `json:"events"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}

type EventResponse struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	EventDate        string    `json:"event_date"`
	EventTime        string    `json:"event_time,omitempty"`
	Location         string    `json:"location"`
	LocationID       int       `json:"location_id"`
	EventType        string    `json:"event_type"`
	EventTypeID      int       `json:"event_type_id"`
	Duration         string    `json:"duration"`
	EntranceType     string    `json:"entrance_type"`
	EntranceTypeID   int       `json:"entrance_type_id"`
	EntranceFee      float64   `json:"entrance_fee"`
	ContactEmail     string    `json:"contact_email"`
	ContactMobile    string    `json:"contact_mobile"`
	Notes            string    `json:"notes"`
	ImageURL         string    `json:"image_url"`
	Organizer        string    `json:"organizer"`
	OrganizationName string    `json:"organization_name"`
	IsPaid           bool      `json:"is_paid"`
	IsPublished      bool      `json:"is_published"`
	CreatedAt        time.Time `json:"created_at"`
}

func (e *Event) ToResponse() *EventResponse {
	eventTime := ""
	if e.EventTime != nil {
		eventTime = *e.EventTime
	}

	return &EventResponse{
		ID:               e.ID,
		Title:            e.Title,
		EventDate:        e.EventDate.Format("2006-01-02"),
		EventTime:        eventTime,
		Location:         e.LocationName,
		LocationID:       e.LocationID,
		EventType:        e.EventTypeName,
		EventTypeID:      e.EventTypeID,
		Duration:         e.Duration,
		EntranceType:     e.EntranceTypeName,
		EntranceTypeID:   e.EntranceTypeID,
		EntranceFee:      e.EntranceFee,
		ContactEmail:     e.ContactEmail,
		ContactMobile:    e.ContactMobile,
		Notes:            e.Notes,
		ImageURL:         e.ImageURL,
		Organizer:        e.CreatorName,
		OrganizationName: e.OrganizationName,
		IsPaid:           e.IsPaid,
		IsPublished:      e.IsPublished,
		CreatedAt:        e.CreatedAt,
	}
}
