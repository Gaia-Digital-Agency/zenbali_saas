# Zen Bali

**Created:** 2025-01-03  
**Last Updated:** 2026-01-10
**GitHub Remote:** https://github.com/net1io/zenbali  
**Developed by:** net1io.com  
**Copyright (C) 2024-2026**

---

## Introduction

Zen Bali is a SaaS events platform for Bali, Indonesia, accessible at **zenbali.org**. The platform enables content creators and service providers to post events happening across Bali, while visitors can freely browse, search, and filter upcoming events.

**Business Model:**
- Visitors browse events for free.
- Content creators pay a flat fee of USD $10 per event posting.
- Events are published immediately upon successful payment.

**Core Features:**
- Public event listing with search and filtering (location, date, event type, etc.).
- Creator registration, authentication, and event management portal.
- Stripe payment integration for event posting fees.
- Admin panel for platform management.
- Visitor tracking and statistics.

---

## Architecture

The application is designed for cloud-native deployment, primarily on Google Cloud Platform.

```
┌─────────────────────────────────────────────────────────────────────┐
│                         ARCHITECTURE OVERVIEW                        │
└─────────────────────────────────────────────────────────────────────┘

                        ┌───────────────────┐
                        │    CLOUDFLARE     │
                        │  (DNS + CDN/SSL)  │
                        │   zenbali.org     │
                        └─────────┬─────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      GOOGLE CLOUD PLATFORM                          │
│                                                                     │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │                  Cloud Load Balancer (HTTPS)                 │  │
│   └───────────────────────────┬─────────────────────────────────┘  │
│                               │                                     │
│               ┌───────────────┴───────────────┐                    │
│               ▼                               ▼                     │
│   ┌───────────────────────┐     ┌───────────────────────────┐      │
│   │      Cloud Run        │     │   Cloud Storage (GCS)      │      │
│   │    (Go Backend)       │     │    (Event Images)          │      │
│   │                       │     │                            │      │
│   │  - REST API Server    │     │  - Public bucket           │      │
│   │  - Static Frontend    │     │  - CDN enabled             │      │
│   │  - Stripe Webhooks    │     │  - Max 5MB per image       │      │
│   └───────────┬───────────┘     └───────────────────────────┘      │
│               │                                                     │
│               ▼                                                     │
│   ┌───────────────────────┐     ┌───────────────────────────┐      │
│   │      Cloud SQL        │     │    Secret Manager          │      │
│   │   (PostgreSQL 15)     │     │                            │      │
│   │                       │     │  - DB credentials          │      │
│   │  - Private IP         │     │  - Stripe API keys         │      │
│   │  - Daily backups      │     │  - JWT secrets             │      │
│   │  - 7-day retention    │     │  - Admin credentials       │      │
│   └───────────────────────┘     └───────────────────────────┘      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

                        ┌───────────────────┐
                        │      STRIPE       │
                        │  (Payment Gateway)│
                        │  $10 per posting  │
                        └───────────────────┘
```

---

## Technologies & Frameworks

| Layer | Technology | Version/Library | Purpose |
|-------|------------|---------|---------|
| **Backend** | Go | 1.22+ | REST API Server |
| **Router** | Chi | v5 | HTTP Routing & Middleware |
| **Database** | PostgreSQL | 15 | Primary Data Store |
| **DB Driver** | pgx | v5 | PostgreSQL Driver |
| **Payments** | Stripe | stripe-go/v76 | Payment Processing |
| **Auth** | JWT | jwt/v5 | Token-based Authentication |
| **Frontend** | HTML5, CSS3, JS (ES6+) | - | UI & Client-side Logic |
| **Container** | Docker | - | Application Containerization |
| **Orchestration**| Docker Compose | - | Local Development |

---

## File Structure

```
zenbali/
│
├── readme.md
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── .env.example
│
├── frontend/
│   ├── public/
│   │   ├── index.html
│   │   ├── event.html
│   │   ├── admin/
│   │   └── creator/
│   ├── css/
│   ├── js/
│   └── assets/
│
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/ (migrations)
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── services/
│   │   └── utils/
│   ├── go.mod
│   └── go.sum
│
├── reference/
│   ├── API.md
│   └── DEPLOYMENT.md
│
├── scripts/
│   └── init.sql
│
└── uploads/
```

---

## Local Development Setup

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- `make` (optional)

### 1. Clone the Repository
```bash
git clone https://github.com/net1io/zenbali.git
cd zenbali
```

### 2. Environment Variables
Copy the example `.env` file and customize it.
```bash
cp .env.example .env
```
**Required `.env` variables:**
```
# Server Configuration
PORT=8080
ENV=development
BASE_URL=http://localhost:8080

# Database (for local Docker setup)
DB_HOST=localhost
DB_PORT=5432
DB_USER=zenbali
DB_PASSWORD=zenbali_dev_password
DB_NAME=zenbali
DB_SSL_MODE=disable

# JWT Authentication
JWT_SECRET=a_secure_secret_of_at_least_32_characters

# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Uploads
UPLOAD_DIR=./uploads
```

### 3. Start Services
This command starts the PostgreSQL and Redis containers.
```bash
docker-compose up -d
```

### 4. Run Database Migrations
```bash
make migrate-up
```
*This uses `golang-migrate`. Ensure it's installed (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`).*

### 5. Run the Backend Server
```bash
make run
```
The server will be running at `http://localhost:8080`.

### 6. Access the Application
- **Main Page:** `http://localhost:8080`
- **Creator Portal:** `http://localhost:8080/creator/login.html`
- **Admin Panel:** `http://localhost:8080/admin/login.html`

---

## API Endpoints

A summary of the main API endpoints. For full details, see `reference/API.md`.

### Public
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health Check |
| GET | `/api/events` | List all published events |
| GET | `/api/events/{id}` | Get a single event |
| GET | `/api/locations` | List all locations |
| GET | `/api/event-types` | List all event types |
| POST | `/api/visitors` | Track a new visitor |

### Creator (Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/creator/register` | Register a new creator |
| POST | `/api/creator/login` | Login for a creator |
| GET | `/api/creator/events` | List events for the creator |
| POST | `/api/creator/events` | Create a new event |
| POST | `/api/creator/events/{id}/pay`| Create a Stripe payment session |

### Admin (Admin Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/login` | Login for an admin |
| GET | `/api/admin/dashboard` | Get dashboard statistics |
| GET | `/api/admin/events` | List all events |
| GET | `/api/admin/creators` | List all creators |

---

## License

Proprietary. All rights reserved.