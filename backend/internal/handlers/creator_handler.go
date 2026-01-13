package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
	"github.com/net1io/zenbali/internal/services"
	"github.com/net1io/zenbali/internal/utils"
)

type CreatorHandler struct {
	services *services.Services
	repos    *repository.Repositories
	config   *config.Config
}

func NewCreatorHandler(svcs *services.Services, repos *repository.Repositories, cfg *config.Config) *CreatorHandler {
	return &CreatorHandler{
		services: svcs,
		repos:    repos,
		config:   cfg,
	}
}

func (h *CreatorHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}
	utils.Success(w, creator.ToResponse())
}

func (h *CreatorHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

	var req models.CreatorUpdateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Name != "" {
		creator.Name = req.Name
	}
	if req.OrganizationName != "" {
		creator.OrganizationName = req.OrganizationName
	}
	if req.Mobile != "" {
		creator.Mobile = req.Mobile
	}

	if err := h.repos.Creator.Update(r.Context(), creator); err != nil {
		utils.InternalError(w, "Failed to update profile")
		return
	}

	utils.Success(w, creator.ToResponse())
}

func (h *CreatorHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

	query := r.URL.Query()
	page := 1
	limit := 20
	includePast := query.Get("include_past") == "true"

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

	result, err := h.services.Event.ListByCreator(r.Context(), creator.ID, page, limit, includePast)
	if err != nil {
		log.Printf("ERROR listing events for creator %s: %v", creator.ID, err)
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

func (h *CreatorHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

	var req models.EventCreateRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if req.Title == "" {
		utils.BadRequest(w, "Title is required")
		return
	}
	if req.EventDate == "" {
		utils.BadRequest(w, "Event date is required")
		return
	}
	if req.EventTime == "" {
		utils.BadRequest(w, "Event time is required")
		return
	}
	if req.LocationID <= 0 {
		utils.BadRequest(w, "Location is required")
		return
	}
	if req.EventTypeID <= 0 {
		utils.BadRequest(w, "Event type is required")
		return
	}
	if req.Duration == "" {
		utils.BadRequest(w, "Duration is required")
		return
	}
	if req.EntranceTypeID <= 0 {
		utils.BadRequest(w, "Entrance type is required")
		return
	}
	if req.ParticipantGroupType == "" {
		utils.BadRequest(w, "Participant group type is required")
		return
	}
	if req.LeadBy == "" {
		utils.BadRequest(w, "Lead by is required")
		return
	}
	if req.ContactEmail == "" {
		utils.BadRequest(w, "Contact email is required")
		return
	}
	if req.ContactMobile == "" {
		utils.BadRequest(w, "Contact mobile is required")
		return
	}
	if req.Notes == "" {
		utils.BadRequest(w, "Event description is required")
		return
	}

	event, err := h.services.Event.Create(r.Context(), creator.ID, &req)
	if err != nil {
		if err == services.ErrInvalidDate {
			utils.BadRequest(w, "Invalid date format. Use YYYY-MM-DD")
			return
		}
		log.Printf("ERROR creating event for creator %s: %v", creator.ID, err)
		utils.InternalError(w, "Failed to create event")
		return
	}

	utils.Created(w, event.ToResponse())
}

func (h *CreatorHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

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

	// Check ownership
	if event.CreatorID != creator.ID {
		utils.Forbidden(w, "Not authorized to view this event")
		return
	}

	utils.Success(w, event.ToResponse())
}

func (h *CreatorHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

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

	event, err := h.services.Event.Update(r.Context(), id, creator.ID, &req, false)
	if err != nil {
		switch err {
		case services.ErrEventNotFound:
			utils.NotFound(w, "Event not found")
		case services.ErrNotEventOwner:
			utils.Forbidden(w, "Not authorized to update this event")
		case services.ErrEventInPast:
			utils.BadRequest(w, "Cannot modify past events")
		case services.ErrInvalidDate:
			utils.BadRequest(w, "Invalid date format")
		default:
			utils.InternalError(w, "Failed to update event")
		}
		return
	}

	utils.Success(w, event.ToResponse())
}

func (h *CreatorHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid event ID")
		return
	}

	if err := h.services.Event.Delete(r.Context(), id, creator.ID, false); err != nil {
		switch err {
		case services.ErrEventNotFound:
			utils.NotFound(w, "Event not found")
		case services.ErrNotEventOwner:
			utils.Forbidden(w, "Not authorized to delete this event")
		default:
			utils.InternalError(w, "Failed to delete event")
		}
		return
	}

	utils.Message(w, "Event deleted successfully")
}

func (h *CreatorHandler) UploadEventImage(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.BadRequest(w, "Invalid event ID")
		return
	}

	// Parse multipart form
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

	// Save file
	imageURL, err := h.services.Upload.SaveEventImage(file, header)
	if err != nil {
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

	// Update event with image URL
	if err := h.services.Event.UpdateImageURL(r.Context(), id, creator.ID, imageURL); err != nil {
		switch err {
		case services.ErrEventNotFound:
			utils.NotFound(w, "Event not found")
		case services.ErrNotEventOwner:
			utils.Forbidden(w, "Not authorized")
		default:
			utils.InternalError(w, "Failed to update event")
		}
		return
	}

	utils.Success(w, map[string]string{
		"image_url": imageURL,
	})
}

func (h *CreatorHandler) CreatePaymentSession(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

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

	if event.CreatorID != creator.ID {
		utils.Forbidden(w, "Not authorized")
		return
	}

	if event.IsPaid {
		utils.BadRequest(w, "Event is already paid")
		return
	}

	// Build success and cancel URLs
	baseURL := h.config.BaseURL
	successURL := baseURL + "/creator/payment-success.html?event_id=" + id.String()
	cancelURL := baseURL + "/creator/payment-cancel.html?event_id=" + id.String()

	session, err := h.services.Payment.CreateCheckoutSession(r.Context(), event, successURL, cancelURL)
	if err != nil {
		if err == services.ErrAlreadyPaid {
			utils.BadRequest(w, "Event is already paid")
			return
		}
		utils.InternalError(w, "Failed to create payment session")
		return
	}

	utils.Success(w, session)
}

func (h *CreatorHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	creator := GetCreatorFromContext(r.Context())
	if creator == nil {
		utils.Unauthorized(w, "")
		return
	}

	query := r.URL.Query()
	page := 1
	limit := 20

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

	result, err := h.services.Payment.ListByCreator(r.Context(), creator.ID, page, limit)
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
