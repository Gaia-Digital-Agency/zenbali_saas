package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrNotEventOwner = errors.New("not the owner of this event")
	ErrEventInPast   = errors.New("cannot modify past events")
	ErrInvalidDate   = errors.New("invalid date format")
)

type EventService struct {
	repos  *repository.Repositories
	upload *UploadService
}

func NewEventService(repos *repository.Repositories, upload *UploadService) *EventService {
	return &EventService{repos: repos, upload: upload}
}

func resolveEntranceFee(entranceFee float64, priceThousands *int) (float64, error) {
	if priceThousands == nil {
		return entranceFee, nil
	}
	if *priceThousands < 0 || *priceThousands > 100000 {
		return 0, fmt.Errorf("price_thousands must be between 0 and 100000")
	}
	return float64(*priceThousands * 1000), nil
}

func (s *EventService) Create(ctx context.Context, creatorID uuid.UUID, req *models.EventCreateRequest) (*models.Event, error) {
	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		return nil, ErrInvalidDate
	}

	entranceFee, err := resolveEntranceFee(req.EntranceFee, req.PriceThousands)
	if err != nil {
		return nil, err
	}

	var eventTime *string
	if req.EventTime != "" {
		eventTime = &req.EventTime
	}

	var duration *string
	if req.Duration != "" {
		duration = &req.Duration
	}

	var participantGroupType *string
	if req.ParticipantGroupType != "" {
		participantGroupType = &req.ParticipantGroupType
	}

	var leadBy *string
	if req.LeadBy != "" {
		leadBy = &req.LeadBy
	}

	var contactMobile *string
	if req.ContactMobile != "" {
		contactMobile = &req.ContactMobile
	}

	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	event := &models.Event{
		CreatorID:            creatorID,
		Title:                req.Title,
		EventDate:            eventDate,
		EventTime:            eventTime,
		LocationID:           req.LocationID,
		EventTypeID:          req.EventTypeID,
		Duration:             duration,
		EntranceTypeID:       req.EntranceTypeID,
		EntranceFee:          entranceFee,
		ParticipantGroupType: participantGroupType,
		LeadBy:               leadBy,
		ContactEmail:         req.ContactEmail,
		ContactMobile:        contactMobile,
		Notes:                notes,
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
		event.Duration = &req.Duration
	}
	if req.EntranceTypeID > 0 {
		event.EntranceTypeID = req.EntranceTypeID
	}
	if req.PriceThousands != nil {
		entranceFee, err := resolveEntranceFee(0, req.PriceThousands)
		if err != nil {
			return nil, err
		}
		event.EntranceFee = entranceFee
	} else if req.EntranceFee != nil {
		event.EntranceFee = *req.EntranceFee
	}
	if req.ContactEmail != "" {
		event.ContactEmail = req.ContactEmail
	}
	if req.ContactMobile != "" {
		event.ContactMobile = &req.ContactMobile
	}
	if req.Notes != "" {
		event.Notes = &req.Notes
	}
	if req.ParticipantGroupType != "" {
		event.ParticipantGroupType = &req.ParticipantGroupType
	}
	if req.LeadBy != "" {
		event.LeadBy = &req.LeadBy
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

	if err := s.repos.Event.Delete(ctx, id); err != nil {
		return err
	}

	if event.ImageURL != nil && s.upload != nil {
		if err := s.upload.DeleteFile(*event.ImageURL); err != nil {
			return err
		}
	}

	return nil
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

	if err := s.repos.Event.UpdateImageURL(ctx, id, imageURL); err != nil {
		if s.upload != nil {
			_ = s.upload.DeleteFile(imageURL)
		}
		return err
	}

	if event.ImageURL != nil && *event.ImageURL != "" && *event.ImageURL != imageURL && s.upload != nil {
		if err := s.upload.DeleteFile(*event.ImageURL); err != nil {
			return err
		}
	}

	return nil
}

func (s *EventService) ListPublic(ctx context.Context, filter models.EventListFilter) (*models.EventListResponse, error) {
	today := time.Now().Truncate(24 * time.Hour)
	filter.OnlyPublished = true
	filter.IncludePast = false
	filter.MinEventDate = today

	events, total, err := s.repos.Event.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		filter.IncludePast = true
		filter.ShowPastEvents = true
		filter.Limit = 100
		filter.MinEventDate = time.Time{} // a zero value for time.Time
		events, total, err = s.repos.Event.List(ctx, filter)
		if err != nil {
			return nil, err
		}
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
	today := time.Now().Truncate(24 * time.Hour)
	filter := models.EventListFilter{
		CreatorID:    creatorID,
		IncludePast:  includePast,
		MinEventDate: today.AddDate(0, 0, -90),
		Page:         page,
		Limit:        limit,
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
	filter.MinEventDate = time.Now().Truncate(24*time.Hour).AddDate(0, 0, -120)

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
