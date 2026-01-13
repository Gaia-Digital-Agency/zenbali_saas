package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/services"
	"github.com/net1io/zenbali/internal/utils"
)

type contextKey string

const (
	ContextKeyCreator contextKey = "creator"
	ContextKeyAdmin   contextKey = "admin"
	ContextKeyUserID  contextKey = "user_id"
)

type AuthHandler struct {
	services *services.Services
	config   *config.Config
}

func NewAuthHandler(svcs *services.Services, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		services: svcs,
		config:   cfg,
	}
}

func (h *AuthHandler) CreatorRegister(w http.ResponseWriter, r *http.Request) {
	var req models.CreatorRegisterRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Basic validation
	if req.Name == "" || req.Email == "" || req.Password == "" {
		utils.BadRequest(w, "Name, email, and password are required")
		return
	}

	if len(req.Password) < 8 {
		utils.BadRequest(w, "Password must be at least 8 characters")
		return
	}

	creator, err := h.services.Auth.RegisterCreator(r.Context(), &req)
	if err != nil {
		if err == services.ErrEmailExists {
			utils.BadRequest(w, "Email already registered")
			return
		}
		utils.InternalError(w, "Failed to register")
		return
	}

	utils.Created(w, creator.ToResponse())
}

func (h *AuthHandler) CreatorLogin(w http.ResponseWriter, r *http.Request) {
	var req models.CreatorLoginRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.BadRequest(w, "Email and password are required")
		return
	}

	creator, token, err := h.services.Auth.LoginCreator(r.Context(), &req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			utils.Unauthorized(w, "Invalid email or password")
			return
		}
		if err == services.ErrAccountDisabled {
			utils.Forbidden(w, "Account is disabled")
			return
		}
		utils.InternalError(w, "Login failed")
		return
	}

	utils.Success(w, map[string]interface{}{
		"token":   token,
		"creator": creator.ToResponse(),
	})
}

func (h *AuthHandler) CreatorLogout(w http.ResponseWriter, r *http.Request) {
	// For JWT-based auth, logout is handled client-side by removing the token
	utils.Message(w, "Logged out successfully")
}

func (h *AuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req models.AdminLoginRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.BadRequest(w, "Email and password are required")
		return
	}

	admin, token, err := h.services.Auth.LoginAdmin(r.Context(), &req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			utils.Unauthorized(w, "Invalid email or password")
			return
		}
		if err == services.ErrAccountDisabled {
			utils.Forbidden(w, "Account is disabled")
			return
		}
		utils.InternalError(w, "Login failed")
		return
	}

	utils.Success(w, map[string]interface{}{
		"token": token,
		"admin": admin.ToResponse(),
	})
}

func (h *AuthHandler) CreatorAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			utils.Unauthorized(w, "Missing authorization token")
			return
		}

		claims, err := h.services.Auth.ValidateToken(token)
		if err != nil {
			utils.Unauthorized(w, "Invalid or expired token")
			return
		}

		if claims.UserType != "creator" {
			utils.Forbidden(w, "Access denied")
			return
		}

		creator, err := h.services.Auth.GetCreatorByID(r.Context(), claims.UserID)
		if err != nil || creator == nil {
			utils.Unauthorized(w, "Creator not found")
			return
		}

		if !creator.IsActive {
			utils.Forbidden(w, "Account is disabled")
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyCreator, creator)
		ctx = context.WithValue(ctx, ContextKeyUserID, creator.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			utils.Unauthorized(w, "Missing authorization token")
			return
		}

		claims, err := h.services.Auth.ValidateToken(token)
		if err != nil {
			utils.Unauthorized(w, "Invalid or expired token")
			return
		}

		if claims.UserType != "admin" {
			utils.Forbidden(w, "Admin access required")
			return
		}

		admin, err := h.services.Auth.GetAdminByID(r.Context(), claims.UserID)
		if err != nil || admin == nil {
			utils.Unauthorized(w, "Admin not found")
			return
		}

		if !admin.IsActive {
			utils.Forbidden(w, "Account is disabled")
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyAdmin, admin)
		ctx = context.WithValue(ctx, ContextKeyUserID, admin.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalCreatorAuthMiddleware attempts to authenticate a creator if a token is present,
// but does not reject the request if no token is provided
func (h *AuthHandler) OptionalCreatorAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		claims, err := h.services.Auth.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		if claims.UserType != "creator" {
			// Not a creator token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		creator, err := h.services.Auth.GetCreatorByID(r.Context(), claims.UserID)
		if err != nil || creator == nil || !creator.IsActive {
			// Creator not found or inactive, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Valid creator token, add to context
		ctx := context.WithValue(r.Context(), ContextKeyCreator, creator)
		ctx = context.WithValue(ctx, ContextKeyUserID, creator.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to get creator from context
func GetCreatorFromContext(ctx context.Context) *models.Creator {
	if creator, ok := ctx.Value(ContextKeyCreator).(*models.Creator); ok {
		return creator
	}
	return nil
}

// Helper to get admin from context
func GetAdminFromContext(ctx context.Context) *models.Admin {
	if admin, ok := ctx.Value(ContextKeyAdmin).(*models.Admin); ok {
		return admin
	}
	return nil
}

// Helper to get user ID from context
func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// Extract token from Authorization header
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}
