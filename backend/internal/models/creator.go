package models

import (
	"time"

	"github.com/google/uuid"
)

type Creator struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	OrganizationName string     `json:"organization_name"`
	Email            string     `json:"email"`
	Mobile           string     `json:"mobile"`
	PasswordHash     string     `json:"-"`
	IsVerified       bool       `json:"is_verified"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreatorRegisterRequest struct {
	Name             string `json:"name" validate:"required,min=2,max=255"`
	OrganizationName string `json:"organization_name" validate:"max=255"`
	Email            string `json:"email" validate:"required,email"`
	Mobile           string `json:"mobile" validate:"max=50"`
	Password         string `json:"password" validate:"required,min=8,max=100"`
}

type CreatorLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreatorUpdateRequest struct {
	Name             string `json:"name" validate:"min=2,max=255"`
	OrganizationName string `json:"organization_name" validate:"max=255"`
	Mobile           string `json:"mobile" validate:"max=50"`
}

type CreatorResponse struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	OrganizationName string    `json:"organization_name"`
	Email            string    `json:"email"`
	Mobile           string    `json:"mobile"`
	IsVerified       bool      `json:"is_verified"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

func (c *Creator) ToResponse() *CreatorResponse {
	return &CreatorResponse{
		ID:               c.ID,
		Name:             c.Name,
		OrganizationName: c.OrganizationName,
		Email:            c.Email,
		Mobile:           c.Mobile,
		IsVerified:       c.IsVerified,
		IsActive:         c.IsActive,
		CreatedAt:        c.CreatedAt,
	}
}
