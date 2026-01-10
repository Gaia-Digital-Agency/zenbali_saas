package handlers

import (
	"encoding/csv"
	"fmt"
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

type AdminHandler struct {
	services *services.Services
	repos    *repository.Repositories
}

func NewAdminHandler(svcs *services.Services, repos *repository.Repositories) *AdminHandler {
	return &AdminHandler{
		services: svcs,
		repos:    repos,
	}
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	totalEvents, publishedEvents, upcomingEvents, err := h.repos.Event.Count(ctx)
	if err != nil {
		utils.InternalError(w, "Failed to fetch event stats")
		return
	}

	totalCreators, activeCreators, err := h.repos.Creator.Count(ctx)
	if err != nil {
		utils.InternalError(w, "Failed to fetch creator stats")
		return
	}

	totalPayments, totalRevenue, err := h.repos.Payment.GetStats(ctx)
	if err != nil {
		utils.InternalError(w, "Failed to fetch payment stats")
		return
	}

	totalVisitors, _ := h.repos.Visitor.GetTotalCount(ctx)
	todayVisitors, _ := h.repos.Visitor.GetTodayCount(ctx)
	recentEvents, _ := h.repos.Event.GetRecent(ctx, 5)
	recentPayments, _ := h.repos.Payment.GetRecent(ctx, 5)

	if recentEvents == nil {
		recentEvents = []*models.Event{}
	}
	if recentPayments == nil {
		recentPayments = []*models.Payment{}
	}

	stats := models.DashboardStats{
		TotalEvents:     totalEvents,
		PublishedEvents: publishedEvents,
		UpcomingEvents:  upcomingEvents,
		TotalCreators:   totalCreators,
		ActiveCreators:  activeCreators,
		TotalPayments:   totalPayments,
		TotalRevenue:    totalRevenue,
		TotalVisitors:   totalVisitors,
		TodayVisitors:   todayVisitors,
		RecentEvents:    recentEvents,
		RecentPayments:  recentPayments,
	}

	utils.Success(w, stats)
}

func (h *AdminHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := models.EventListFilter{IncludePast: true}

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

	if search := query.Get("search"); search != "" {
		filter.Search = search
	}

	result, err := h.services.Event.ListAll(r.Context(), filter)
	if err != nil {
		utils.InternalError(w, "Failed to fetch events")
		return
	}

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

func (h *AdminHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid event ID")
		return
	}

	var req models.EventUpdateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	event, err := h.services.Event.Update(r.Context(), id, uuid.Nil, &req, true)
	if err != nil {
		if err == services.ErrEventNotFound {
			utils.NotFound(w, "Event not found")
			return
		}
		utils.InternalError(w, "Failed to update event")
		return
	}

	utils.Success(w, event.ToResponse())
}

func (h *AdminHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid event ID")
		return
	}

	if err := h.services.Event.Delete(r.Context(), id, uuid.Nil, true); err != nil {
		if err == services.ErrEventNotFound {
			utils.NotFound(w, "Event not found")
			return
		}
		utils.InternalError(w, "Failed to delete event")
		return
	}

	utils.Message(w, "Event deleted successfully")
}

func (h *AdminHandler) ListCreators(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, limit := 1, 20

	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	creators, total, err := h.repos.Creator.List(r.Context(), page, limit)
	if err != nil {
		utils.InternalError(w, "Failed to fetch creators")
		return
	}

	var responses []*models.CreatorResponse
	for _, c := range creators {
		responses = append(responses, c.ToResponse())
	}

	utils.Success(w, map[string]interface{}{
		"creators":    responses,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	})
}

func (h *AdminHandler) UpdateCreator(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid creator ID")
		return
	}

	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.IsActive != nil {
		if err := h.repos.Creator.UpdateStatus(r.Context(), id, *req.IsActive); err != nil {
			utils.InternalError(w, "Failed to update creator")
			return
		}
	}

	creator, err := h.repos.Creator.GetByID(r.Context(), id)
	if err != nil || creator == nil {
		utils.NotFound(w, "Creator not found")
		return
	}

	utils.Success(w, creator.ToResponse())
}

func (h *AdminHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, limit := 1, 20
	status := query.Get("status")

	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	result, err := h.services.Payment.ListAll(r.Context(), page, limit, status)
	if err != nil {
		utils.InternalError(w, "Failed to fetch payments")
		return
	}

	var payments []*models.PaymentResponse
	for _, p := range result.Payments {
		payments = append(payments, p.ToResponse())
	}

	utils.Success(w, map[string]interface{}{
		"payments":    payments,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

func (h *AdminHandler) ExportPayments(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Payment.ListAll(r.Context(), 1, 10000, "completed")
	if err != nil {
		utils.InternalError(w, "Failed to fetch payments")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=payments_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"ID", "Event", "Creator", "Amount", "Currency", "Status", "Date"})

	for _, p := range result.Payments {
		writer.Write([]string{
			p.ID.String(),
			p.EventTitle,
			p.CreatorName,
			fmt.Sprintf("%.2f", float64(p.AmountCents)/100),
			p.Currency,
			p.Status,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func (h *AdminHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := h.repos.Location.List(r.Context(), false)
	if err != nil {
		utils.InternalError(w, "Failed to fetch locations")
		return
	}
	utils.Success(w, locations)
}

func (h *AdminHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req models.LocationRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		utils.BadRequest(w, "Name is required")
		return
	}

	loc := &models.Location{Name: req.Name, IsActive: true}
	if err := h.repos.Location.Create(r.Context(), loc); err != nil {
		utils.InternalError(w, "Failed to create location")
		return
	}

	utils.Created(w, loc)
}

func (h *AdminHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid location ID")
		return
	}

	var req models.LocationRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := h.repos.Location.Update(r.Context(), id, req.Name, isActive); err != nil {
		utils.InternalError(w, "Failed to update location")
		return
	}

	utils.Message(w, "Location updated successfully")
}

func (h *AdminHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.repos.EventType.List(r.Context(), false)
	if err != nil {
		utils.InternalError(w, "Failed to fetch event types")
		return
	}
	utils.Success(w, types)
}

func (h *AdminHandler) CreateEventType(w http.ResponseWriter, r *http.Request) {
	var req models.EventTypeRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		utils.BadRequest(w, "Name is required")
		return
	}

	et := &models.EventType{Name: req.Name, IsActive: true}
	if err := h.repos.EventType.Create(r.Context(), et); err != nil {
		utils.InternalError(w, "Failed to create event type")
		return
	}

	utils.Created(w, et)
}

func (h *AdminHandler) UpdateEventType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid event type ID")
		return
	}

	var req models.EventTypeRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := h.repos.EventType.Update(r.Context(), id, req.Name, isActive); err != nil {
		utils.InternalError(w, "Failed to update event type")
		return
	}

	utils.Message(w, "Event type updated successfully")
}
