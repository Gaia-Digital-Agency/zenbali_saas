package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/database"
	"github.com/net1io/zenbali/internal/handlers"
	"github.com/net1io/zenbali/internal/repository"
	"github.com/net1io/zenbali/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/stripe/stripe-go/v76"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	// if err := db.RunMigrations(); err != nil {
	// 	log.Fatalf("Failed to run migrations: %v", err)
	// }

	// Initialize Stripe
	stripe.Key = cfg.Stripe.SecretKey

	// Initialize repositories
	repos := &repository.Repositories{
		Creator:      repository.NewCreatorRepository(db.Pool),
		Event:        repository.NewEventRepository(db.Pool),
		Payment:      repository.NewPaymentRepository(db.Pool),
		Admin:        repository.NewAdminRepository(db.Pool),
		Location:     repository.NewLocationRepository(db.Pool),
		EventType:    repository.NewEventTypeRepository(db.Pool),
		EntranceType: repository.NewEntranceTypeRepository(db.Pool),
		Visitor:      repository.NewVisitorRepository(db.Pool),
	}

	// Initialize services
	svcs := &services.Services{
		Auth:    services.NewAuthService(repos, cfg.JWT),
		Event:   services.NewEventService(repos),
		Payment: services.NewPaymentService(repos, cfg.Stripe),
		Upload:  services.NewUploadService(cfg.Upload),
		Visitor: services.NewVisitorService(repos),
	}

	// Initialize handlers
	h := handlers.New(svcs, repos, cfg)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", h.HealthCheck)

		// Public routes
		r.Get("/events", h.Public.ListEvents)
		r.Get("/events/{id}", h.Public.GetEvent)
		r.Get("/locations", h.Public.ListLocations)
		r.Get("/event-types", h.Public.ListEventTypes)
		r.Get("/entrance-types", h.Public.ListEntranceTypes)

		// Visitor tracking
		r.Post("/visitors", h.Visitor.TrackVisitor)
		r.Get("/visitors/stats", h.Visitor.GetStats)

		// Creator authentication
		r.Post("/creator/register", h.Auth.CreatorRegister)
		r.Post("/creator/login", h.Auth.CreatorLogin)
		r.Post("/creator/logout", h.Auth.CreatorLogout)

		// Creator protected routes
		r.Group(func(r chi.Router) {
			r.Use(h.Auth.CreatorAuthMiddleware)

			r.Get("/creator/profile", h.Creator.GetProfile)
			r.Put("/creator/profile", h.Creator.UpdateProfile)
			r.Get("/creator/events", h.Creator.ListEvents)
			r.Post("/creator/events", h.Creator.CreateEvent)
			r.Get("/creator/events/{id}", h.Creator.GetEvent)
			r.Put("/creator/events/{id}", h.Creator.UpdateEvent)
			r.Delete("/creator/events/{id}", h.Creator.DeleteEvent)
			r.Post("/creator/events/{id}/upload-image", h.Creator.UploadEventImage)
			r.Post("/creator/events/{id}/pay", h.Creator.CreatePaymentSession)
			r.Get("/creator/payments", h.Creator.ListPayments)
		})

		// Admin authentication
		r.Post("/admin/login", h.Auth.AdminLogin)

		// Admin protected routes
		r.Group(func(r chi.Router) {
			r.Use(h.Auth.AdminAuthMiddleware)

			r.Get("/admin/dashboard", h.Admin.Dashboard)
			r.Get("/admin/events", h.Admin.ListEvents)
			r.Put("/admin/events/{id}", h.Admin.UpdateEvent)
			r.Delete("/admin/events/{id}", h.Admin.DeleteEvent)
			r.Get("/admin/creators", h.Admin.ListCreators)
			r.Put("/admin/creators/{id}", h.Admin.UpdateCreator)
			r.Get("/admin/payments", h.Admin.ListPayments)
			r.Get("/admin/payments/export", h.Admin.ExportPayments)
			r.Get("/admin/settings/locations", h.Admin.ListLocations)
			r.Post("/admin/settings/locations", h.Admin.CreateLocation)
			r.Put("/admin/settings/locations/{id}", h.Admin.UpdateLocation)
			r.Get("/admin/settings/event-types", h.Admin.ListEventTypes)
			r.Post("/admin/settings/event-types", h.Admin.CreateEventType)
			r.Put("/admin/settings/event-types/{id}", h.Admin.UpdateEventType)
		})

		// Stripe webhook
		r.Post("/webhooks/stripe", h.Webhook.HandleStripe)
	})

	// Serve uploaded files
	fileServer := http.FileServer(http.Dir(cfg.Upload.Dir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// Serve static frontend files
	r.Handle("/*", http.FileServer(http.Dir("../frontend/public")))

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌴 Zen Bali server starting on port %s", cfg.Port)
		log.Printf("📍 Environment: %s", cfg.Env)
		log.Printf("🌐 Base URL: %s", cfg.BaseURL)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
