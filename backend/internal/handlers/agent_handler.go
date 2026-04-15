package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
	"github.com/net1io/zenbali/internal/services"
	"github.com/net1io/zenbali/internal/utils"
)

type AgentHandler struct {
	services *services.Services
	repos    *repository.Repositories
	config   *config.Config
}

type AgentEventCreateRequest struct {
	Title                string  `json:"title"`
	EventDate            string  `json:"event_date"`
	EventTime            string  `json:"event_time"`
	Location             string  `json:"location"`
	EventType            string  `json:"event_type"`
	DurationDays         int     `json:"duration_days"`
	DurationHours        int     `json:"duration_hours"`
	DurationMinutes      int     `json:"duration_minutes"`
	EntranceType         string  `json:"entrance_type"`
	ParticipantGroupType string  `json:"participant_group_type"`
	LeadBy               string  `json:"lead_by"`
	Venue                string  `json:"venue"`
	ContactEmail         string  `json:"contact_email"`
	ContactMobile        string  `json:"contact_mobile"`
	EventDescription     string  `json:"event_description"`
	ImageURL             string  `json:"image_url"`
	EntranceFee          float64 `json:"entrance_fee"`
	PriceThousands       *int    `json:"price_thousands,omitempty"`
}

func NewAgentHandler(svcs *services.Services, repos *repository.Repositories, cfg *config.Config) *AgentHandler {
	return &AgentHandler{
		services: svcs,
		repos:    repos,
		config:   cfg,
	}
}

func (h *AgentHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(h.config.Agent.CreatorEmail) == "" {
		utils.InternalError(w, "Agent creator email is not configured")
		return
	}

	var req AgentEventCreateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if err := validateAgentEventRequest(&req); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	creator, err := h.repos.Creator.GetByEmail(r.Context(), h.config.Agent.CreatorEmail)
	if err != nil {
		utils.InternalError(w, "Failed to load agent creator")
		return
	}
	if creator == nil || !creator.IsActive {
		utils.InternalError(w, "Configured agent creator is unavailable")
		return
	}

	locationID, err := resolveLocationID(r.Context(), h.repos, req.Location)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	eventTypeID, err := resolveEventTypeID(r.Context(), h.repos, req.EventType)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	entranceTypeID, err := resolveEntranceTypeID(r.Context(), h.repos, req.EntranceType)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	duration, err := formatDuration(req.DurationDays, req.DurationHours, req.DurationMinutes)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	eventTime, err := normalizeEventTime(req.EventTime)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	createReq := &models.EventCreateRequest{
		Title:                strings.TrimSpace(req.Title),
		EventDate:            strings.TrimSpace(req.EventDate),
		EventTime:            eventTime,
		LocationID:           locationID,
		EventTypeID:          eventTypeID,
		Duration:             duration,
		EntranceTypeID:       entranceTypeID,
		EntranceFee:          req.EntranceFee,
		PriceThousands:       req.PriceThousands,
		ParticipantGroupType: strings.TrimSpace(req.ParticipantGroupType),
		LeadBy:               strings.TrimSpace(req.LeadBy),
		Venue:                strings.TrimSpace(req.Venue),
		ContactEmail:         strings.TrimSpace(req.ContactEmail),
		ContactMobile:        strings.TrimSpace(req.ContactMobile),
		Notes:                strings.TrimSpace(req.EventDescription),
	}

	event, err := h.services.Event.Create(r.Context(), creator.ID, createReq)
	if err != nil {
		if err == services.ErrInvalidDate {
			utils.BadRequest(w, "Invalid date format. Use YYYY-MM-DD")
			return
		}
		utils.InternalError(w, "Failed to create event")
		return
	}

	imageURL := strings.TrimSpace(req.ImageURL)
	isPaid := true
	isPublished := true
	if err := h.repos.Event.UpdateAdminFields(r.Context(), event.ID, optionalString(imageURL), &isPaid, &isPublished); err != nil {
		utils.InternalError(w, "Failed to publish agent event")
		return
	}

	event, err = h.services.Event.GetByID(r.Context(), event.ID)
	if err != nil {
		utils.InternalError(w, "Failed to fetch created event")
		return
	}

	utils.Created(w, event.ToResponse())
}

func (h *AgentHandler) UploadEventImage(w http.ResponseWriter, r *http.Request) {
	maxSize := int64(h.config.Upload.MaxSizeMB * 1024 * 1024)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		utils.BadRequest(w, "File too large")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.BadRequest(w, "No image file provided")
		return
	}
	defer file.Close()

	imageURL, err := h.services.Upload.SaveEventImage(file, header)
	if err != nil {
		log.Printf("ERROR uploading agent event image: %v", err)
		switch err {
		case services.ErrFileTooLarge:
			utils.BadRequest(w, "File too large")
		case services.ErrInvalidFileType:
			utils.BadRequest(w, "Invalid file type. Allowed: jpg, jpeg, png, webp")
		default:
			utils.InternalError(w, "Failed to upload image")
		}
		return
	}

	utils.Created(w, map[string]string{
		"image_url": imageURL,
	})
}

func validateAgentEventRequest(req *AgentEventCreateRequest) error {
	switch {
	case strings.TrimSpace(req.Title) == "":
		return fmt.Errorf("title is required")
	case strings.TrimSpace(req.EventDate) == "":
		return fmt.Errorf("event_date is required")
	case strings.TrimSpace(req.EventTime) == "":
		return fmt.Errorf("event_time is required")
	case strings.TrimSpace(req.Location) == "":
		return fmt.Errorf("location is required")
	case strings.TrimSpace(req.EventType) == "":
		return fmt.Errorf("event_type is required")
	case strings.TrimSpace(req.EntranceType) == "":
		return fmt.Errorf("entrance_type is required")
	case strings.TrimSpace(req.ParticipantGroupType) == "":
		return fmt.Errorf("participant_group_type is required")
	case strings.TrimSpace(req.LeadBy) == "":
		return fmt.Errorf("lead_by is required")
	case strings.TrimSpace(req.ContactEmail) == "":
		return fmt.Errorf("contact_email is required")
	case strings.TrimSpace(req.ContactMobile) == "":
		return fmt.Errorf("contact_mobile is required")
	case strings.TrimSpace(req.EventDescription) == "":
		return fmt.Errorf("event_description is required")
	}

	if req.DurationDays < 0 || req.DurationHours < 0 || req.DurationMinutes < 0 {
		return fmt.Errorf("duration values must be zero or positive")
	}
	if req.PriceThousands != nil && (*req.PriceThousands < 0 || *req.PriceThousands > 100000) {
		return fmt.Errorf("price_thousands must be between 0 and 100000")
	}

	return nil
}

func resolveLocationID(ctx context.Context, repos *repository.Repositories, input string) (int, error) {
	return resolveReferenceID(ctx, input, repos.Location.List, func(loc *models.Location) (int, string, string) {
		return loc.ID, loc.Name, loc.Slug
	}, "location")
}

func resolveEventTypeID(ctx context.Context, repos *repository.Repositories, input string) (int, error) {
	return resolveReferenceID(ctx, input, repos.EventType.List, func(item *models.EventType) (int, string, string) {
		return item.ID, item.Name, item.Slug
	}, "event_type")
}

func resolveEntranceTypeID(ctx context.Context, repos *repository.Repositories, input string) (int, error) {
	return resolveReferenceID(ctx, input, repos.EntranceType.List, func(item *models.EntranceType) (int, string, string) {
		return item.ID, item.Name, item.Slug
	}, "entrance_type")
}

type listFn[T any] func(ctx context.Context, onlyActive bool) ([]*T, error)
type refPartsFn[T any] func(item *T) (int, string, string)

func resolveReferenceID[T any](ctx context.Context, input string, list listFn[T], parts refPartsFn[T], field string) (int, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return 0, fmt.Errorf("%s is required", field)
	}
	if id, err := strconv.Atoi(raw); err == nil && id > 0 {
		return id, nil
	}

	items, err := list(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("failed to load %s options", field)
	}

	normalized := normalizeLookup(raw)
	for _, item := range items {
		id, name, slug := parts(item)
		if normalized == normalizeLookup(name) || normalized == normalizeLookup(slug) {
			return id, nil
		}
	}

	return 0, fmt.Errorf("unknown %s: %s", field, raw)
}

func normalizeLookup(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatDuration(days, hours, minutes int) (string, error) {
	if minutes%15 != 0 {
		return "", fmt.Errorf("duration_minutes must be in 15-minute increments")
	}
	if days == 0 && hours == 0 && minutes == 0 {
		return "", fmt.Errorf("at least one duration value is required")
	}

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, pluralize(days, "day"))
	}
	if hours > 0 {
		parts = append(parts, pluralize(hours, "hour"))
	}
	if minutes > 0 {
		parts = append(parts, pluralize(minutes, "minute"))
	}

	return strings.Join(parts, " "), nil
}

func pluralize(value int, unit string) string {
	if value == 1 {
		return fmt.Sprintf("%d %s", value, unit)
	}
	return fmt.Sprintf("%d %ss", value, unit)
}

func normalizeEventTime(value string) (string, error) {
	raw := strings.TrimSpace(value)
	parts := strings.Split(raw, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("event_time must use HH:MM or HH:MM:SS format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return "", fmt.Errorf("event_time hour must be between 00 and 23")
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return "", fmt.Errorf("event_time minute must be between 00 and 59")
	}

	if minute%15 != 0 {
		return "", fmt.Errorf("event_time must be in 15-minute increments")
	}

	if len(parts) == 3 {
		second, err := strconv.Atoi(parts[2])
		if err != nil || second < 0 || second > 59 {
			return "", fmt.Errorf("event_time second must be between 00 and 59")
		}
		if second != 0 {
			return "", fmt.Errorf("event_time seconds must be 00")
		}
	}

	return fmt.Sprintf("%02d:%02d", hour, minute), nil
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
