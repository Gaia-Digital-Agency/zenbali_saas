package handlers

import (
	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/repository"
	"github.com/net1io/zenbali/internal/services"
)

// Handlers holds all handler instances
type Handlers struct {
	Auth    *AuthHandler
	Public  *PublicHandler
	Creator *CreatorHandler
	Admin   *AdminHandler
	Webhook *WebhookHandler
	Visitor *VisitorHandler

	services *services.Services
	repos    *repository.Repositories
	config   *config.Config
}

// New creates a new Handlers instance
func New(svcs *services.Services, repos *repository.Repositories, cfg *config.Config) *Handlers {
	h := &Handlers{
		services: svcs,
		repos:    repos,
		config:   cfg,
	}

	h.Auth = NewAuthHandler(svcs, cfg)
	h.Public = NewPublicHandler(svcs, repos)
	h.Creator = NewCreatorHandler(svcs, repos, cfg)
	h.Admin = NewAdminHandler(svcs, repos)
	h.Webhook = NewWebhookHandler(svcs)
	h.Visitor = NewVisitorHandler(svcs)

	return h
}
