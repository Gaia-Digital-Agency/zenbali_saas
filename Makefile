# ===========================================
# Zen Bali Makefile
# ===========================================

.PHONY: help build run dev test clean docker-up docker-down migrate-up migrate-down seed

# Default target
help:
	@echo "Zen Bali - Available Commands:"
	@echo ""
	@echo "  make build        - Build the Go binary"
	@echo "  make run          - Run the compiled binary"
	@echo "  make dev          - Run with hot reload (requires air)"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Remove build artifacts"
	@echo ""
	@echo "  make docker-up    - Start PostgreSQL and Redis"
	@echo "  make docker-down  - Stop Docker services"
	@echo ""
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback last migration"
	@echo "  make seed         - Seed reference data"
	@echo ""
	@echo "  make deploy       - Deploy to GCP Cloud Run"

# Build the application
build:
	@echo "Building Zen Bali..."
	@cd backend && go build -o ../bin/zenbali ./cmd/server
	@echo "Build complete: bin/zenbali"

# Build for production (Linux)
build-prod:
	@echo "Building for production..."
	@cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/zenbali ./cmd/server
	@echo "Production build complete: bin/zenbali"

# Run the application
run: build
	@echo "Starting Zen Bali..."
	@./bin/zenbali

# Run with go run (development)
dev:
	@echo "Starting Zen Bali (dev mode)..."
	@cd backend && go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	@cd backend && go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@cd backend && go test -coverprofile=coverage.out ./...
	@cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: backend/coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf backend/coverage.out backend/coverage.html
	@echo "Clean complete"

# Start Docker services
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	@echo "Services started"

# Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "Services stopped"

# Run database migrations
migrate-up:
	@echo "Running migrations..."
	@cd backend && go run ./cmd/migrate up

# Rollback last migration
migrate-down:
	@echo "Rolling back migration..."
	@cd backend && go run ./cmd/migrate down

# Seed reference data
seed:
	@echo "Seeding reference data..."
	@PGPASSWORD=zenbali_dev_password psql -h localhost -U zenbali -d zenbali -f scripts/seed_locations.sql
	@PGPASSWORD=zenbali_dev_password psql -h localhost -U zenbali -d zenbali -f scripts/seed_event_types.sql
	@PGPASSWORD=zenbali_dev_password psql -h localhost -U zenbali -d zenbali -f scripts/seed_entrance_types.sql
	@PGPASSWORD=zenbali_dev_password psql -h localhost -U zenbali -d zenbali -f scripts/seed_admin.sql
	@echo "Seeding complete"

# Create uploads directory
setup-dirs:
	@mkdir -p uploads
	@mkdir -p bin
	@mkdir -p logs
	@touch uploads/.gitkeep
	@echo "Directories created"

# Full local setup
setup: docker-up setup-dirs
	@sleep 3
	@make migrate-up
	@make seed
	@echo "Setup complete! Run 'make dev' to start the server"

# Deploy to GCP
deploy:
	@echo "Building Docker image..."
	@docker build -t gcr.io/$(GCP_PROJECT)/zenbali:latest .
	@echo "Pushing to GCR..."
	@docker push gcr.io/$(GCP_PROJECT)/zenbali:latest
	@echo "Deploying to Cloud Run..."
	@gcloud run deploy zenbali \
		--image gcr.io/$(GCP_PROJECT)/zenbali:latest \
		--platform managed \
		--region asia-southeast1 \
		--allow-unauthenticated
	@echo "Deployment complete"
