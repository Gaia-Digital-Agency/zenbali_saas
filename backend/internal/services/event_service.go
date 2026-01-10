package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrNotEventOwner     = errors.New("not the owner of this event")
	ErrEventInPast       = errors.New("cannot modify past events")
	ErrInvalidDate       = errors.New("invalid date format")
)

type EventService struct {
	repos *repository.Repositories
}

func NewEventService(repos *repository.Repositories) *EventService {
	return &EventService{repos: repos}
}

func (s *EventService) Create(ctx context.Context, creatorID uuid.UUID, req *models.EventCreateRequest) (*models.Event, error) {
	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		return nil, ErrInvalidDate
	}

	var eventTime *string
	if req.EventTime != "" {
		eventTime = &req.EventTime
	}

	event := &models.Event{
		CreatorID:      creatorID,
		Title:          req.Title,
		EventDate:      eventDate,
		EventTime:      eventTime,
		LocationID:     req.LocationID,
		EventTypeID:    req.EventTypeID,
		Duration:       req.Duration,
		EntranceTypeID: req.EntranceTypeID,
		EntranceFee:    req.EntranceFee,
		ContactEmail:   req.ContactEmail,
		ContactMobile:  req.ContactMobile,
		Notes:          req.Notes,
	}

	if err := s.repos.Event.Create(ctx, event); err != nil {
		return nil, err
	}

	// Fetch with joined fields
	return s.repos.Event.GetByID(ctx, event.ID)
}

func (s *EventService) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	event, err := s.repos.Event.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}

func (s *EventService) Update(ctx context.Context, id, creatorID uuid.UUID, req *models.EventUpdateRequest, isAdmin bool) (*models.Event, error) {
	event, err := s.repos.Event.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}

	// Check ownership (unless admin)
	if !isAdmin && event.CreatorID != creatorID {
		return nil, ErrNotEventOwner
	}

	// Check if event is in past (unless admin)
	if !isAdmin && event.EventDate.Before(time.Now().Truncate(24*time.Hour)) {
		return nil, ErrEventInPast
	}

	// Update fields if provided
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.EventDate != "" {
		eventDate, err := time.Parse("2006-01-02", req.EventDate)
		if err != nil {
			return nil, ErrInvalidDate
		}
		event.EventDate = eventDate
	}
	if req.EventTime != "" {
		event.EventTime = &req.EventTime
	}
	if req.LocationID > 0 {
		event.LocationID = req.LocationID
	}
	if req.EventTypeID > 0 {
		event.EventTypeID = req.EventTypeID
	}
	if req.Duration != "" {
		event.Duration = req.Duration
	}
	if req.EntranceTypeID > 0 {
		event.EntranceTypeID = req.EntranceTypeID
	}
	if req.EntranceFee >= 0 {
		event.EntranceFee = req.EntranceFee
	}
	if req.ContactEmail != "" {
		event.ContactEmail = req.ContactEmail
	}
	if req.ContactMobile != "" {
		event.ContactMobile = req.ContactMobile
	}
	if req.Notes != "" {
		event.Notes = req.Notes
	}

	if err := s.repos.Event.Update(ctx, event); err != nil {
		return nil, err
	}

	return s.repos.Event.GetByID(ctx, id)
}

func (s *EventService) Delete(ctx context.Context, id, creatorID uuid.UUID, isAdmin bool) error {
	event, err := s.repos.Event.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}

	// Check ownership (unless admin)
	if !isAdmin && event.CreatorID != creatorID {
		return ErrNotEventOwner
	}

	return s.repos.Event.Delete(ctx, id)
}

func (s *EventService) UpdateImageURL(ctx context.Context, id, creatorID uuid.UUID, imageURL string) error {
	event, err := s.repos.Event.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}
	if event.CreatorID != creatorID {
		return ErrNotEventOwner
	}

	return s.repos.Event.UpdateImageURL(ctx, id, imageURL)
}

func (s *EventService) ListPublic(ctx context.Context, filter models.EventListFilter) (*models.EventListResponse, error) {
	filter.OnlyPublished = true
	filter.IncludePast = false

	events, total, err := s.repos.Event.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit

	return &models.EventListResponse{
		Events:     events,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *EventService) ListByCreator(ctx context.Context, creatorID uuid.UUID, page, limit int, includePast bool) (*models.EventListResponse, error) {
	filter := models.EventListFilter{
		CreatorID:   creatorID,
		IncludePast: includePast,
		Page:        page,
		Limit:       limit,
	}

	events, total, err := s.repos.Event.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}
	totalPages := (total + limit - 1) / limit

	return &models.EventListResponse{
		Events:     events,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *EventService) ListAll(ctx context.Context, filter models.EventListFilter) (*models.EventListResponse, error) {
	filter.IncludePast = true

	events, total, err := s.repos.Event.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit

	return &models.EventListResponse{
		Events:     events,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *EventService) PublishEvent(ctx context.Context, id uuid.UUID) error {
	return s.repos.Event.UpdatePaymentStatus(ctx, id, true, true)
}
