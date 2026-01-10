package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
	"github.com/net1io/zenbali/internal/services"
	"github.com/net1io/zenbali/internal/utils"
)

type PublicHandler struct {
	services *services.Services
	repos    *repository.Repositories
}

func NewPublicHandler(svcs *services.Services, repos *repository.Repositories) *PublicHandler {
	return &PublicHandler{
		services: svcs,
		repos:    repos,
	}
}

func (h *PublicHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	filter := models.EventListFilter{
		OnlyPublished: true,
		IncludePast:   false,
	}

	// Parse query parameters
	query := r.URL.Query()

	if locationID := query.Get("location_id"); locationID != "" {
		if id, err := strconv.Atoi(locationID); err == nil {
			filter.LocationID = id
		}
	}

	if eventTypeID := query.Get("event_type_id"); eventTypeID != "" {
		if id, err := strconv.Atoi(eventTypeID); err == nil {
			filter.EventTypeID = id
		}
	}

	if entranceTypeID := query.Get("entrance_type_id"); entranceTypeID != "" {
		if id, err := strconv.Atoi(entranceTypeID); err == nil {
			filter.EntranceTypeID = id
		}
	}

	if dateFrom := query.Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = t
		}
	}

	if dateTo := query.Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = t
		}
	}

	if search := query.Get("search"); search != "" {
		filter.Search = search
	}

	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	result, err := h.services.Event.ListPublic(r.Context(), filter)
	if err != nil {
		utils.InternalError(w, "Failed to fetch events")
		return
	}

	// Convert to response format
	var events []*models.EventResponse
	for _, e := range result.Events {
		events = append(events, e.ToResponse())
	}

	utils.Success(w, map[string]interface{}{
		"events":      events,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

func (h *PublicHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid event ID")
		return
	}

	event, err := h.services.Event.GetByID(r.Context(), id)
	if err != nil {
		if err == services.ErrEventNotFound {
			utils.NotFound(w, "Event not found")
			return
		}
		utils.InternalError(w, "Failed to fetch event")
		return
	}

	// Only show published events to public
	if !event.IsPublished {
		utils.NotFound(w, "Event not found")
		return
	}

	utils.Success(w, event.ToResponse())
}

func (h *PublicHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := h.repos.Location.List(r.Context(), true)
	if err != nil {
		utils.InternalError(w, "Failed to fetch locations")
		return
	}
	utils.Success(w, locations)
}

func (h *PublicHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.repos.EventType.List(r.Context(), true)
	if err != nil {
		utils.InternalError(w, "Failed to fetch event types")
		return
	}
	utils.Success(w, types)
}

func (h *PublicHandler) ListEntranceTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.repos.EntranceType.List(r.Context(), true)
	if err != nil {
		utils.InternalError(w, "Failed to fetch entrance types")
		return
	}
	utils.Success(w, types)
}

// HealthCheck handler for the main handlers struct
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.Success(w, map[string]string{
		"status":  "healthy",
		"service": "zenbali",
	})
}
