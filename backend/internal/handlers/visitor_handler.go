package handlers

import (
	"net/http"
	"strings"

	"github.com/net1io/zenbali/internal/services"
	"github.com/net1io/zenbali/internal/utils"
)

type VisitorHandler struct {
	services *services.Services
}

func NewVisitorHandler(svcs *services.Services) *VisitorHandler {
	return &VisitorHandler{services: svcs}
}

func (h *VisitorHandler) TrackVisitor(w http.ResponseWriter, r *http.Request) {
	// Get client IP
	ip := getClientIP(r)
	userAgent := r.UserAgent()

	// Track the visitor
	if err := h.services.Visitor.TrackVisitor(r.Context(), ip, userAgent); err != nil {
		// Don't fail the request, just log
		utils.Success(w, map[string]string{"status": "tracked"})
		return
	}

	utils.Success(w, map[string]string{"status": "tracked"})
}

func (h *VisitorHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.services.Visitor.GetStats(r.Context())
	if err != nil {
		utils.InternalError(w, "Failed to fetch visitor stats")
		return
	}

	utils.Success(w, stats)
}

// getClientIP extracts the client's real IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}

	return ip
}
