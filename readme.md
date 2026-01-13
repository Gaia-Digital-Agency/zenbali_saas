# Zen Bali

**Created:** 2025-01-03
**Last Updated:** 2026-01-13
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
- Public event listing with search and filtering (location, date, event type, entrance fee, etc.).
- Creator registration, authentication, and event management portal.
- Enhanced event creation with:
  - Event time in 15-minute increments
  - Duration with days/hours/minutes breakdown
  - Entrance fee in Rupiah with detailed breakdown
  - Participant group type (Couples, Females Only, Males Only, Open)
  - Event leader information
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
├── start.sh                    # Local development startup script
├── stop.sh                     # Local development stop script
├── Dockerfile
├── docker-compose.yml
├── .env
├── .env.example
│
├── frontend/
│   └── public/
│       ├── index.html
│       ├── event.html
│       ├── admin/
│       │   ├── login.html
│       │   ├── dashboard.html
│       │   └── ...
│       ├── creator/
│       │   ├── login.html
│       │   ├── register.html
│       │   ├── dashboard.html
│       │   ├── new-event.html  # Enhanced with new fields
│       │   ├── edit-event.html
│       │   └── events.html
│       ├── css/
│       │   └── main.css
│       └── js/
│           ├── main.js
│           └── auth.js
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── database/
│   │   │   ├── database.go
│   │   │   └── migrations/
│   │   │       ├── 001_init.up.sql
│   │   │       ├── 002_seed_data.up.sql
│   │   │       └── 003_add_participant_group_and_lead_by.up.sql
│   │   ├── handlers/
│   │   │   ├── admin_handler.go
│   │   │   ├── auth_handler.go
│   │   │   ├── creator_handler.go
│   │   │   ├── public_handler.go
│   │   │   └── ...
│   │   ├── models/
│   │   │   ├── event.go        # Updated with new fields
│   │   │   ├── creator.go
│   │   │   └── ...
│   │   ├── repository/
│   │   │   ├── event_repo.go   # Updated with new fields
│   │   │   ├── creator_repo.go
│   │   │   └── ...
│   │   ├── services/
│   │   │   ├── event_service.go # Updated with new fields
│   │   │   ├── auth_service.go
│   │   │   └── ...
│   │   └── utils/
│   │       └── response.go
│   ├── go.mod
│   └── go.sum
│
├── reference/
│   ├── API.md
│   └── deployment.md
│
└── uploads/                    # Local file uploads directory
```

---

## Quick Start (Local Development)

### Prerequisites
- **Go 1.22+** - [Download](https://golang.org/dl/)
- **Docker Desktop** - [Download](https://www.docker.com/products/docker-desktop)
- **Git** - [Download](https://git-scm.com/downloads)

### Simple Setup (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/net1io/zenbali.git
   cd zenbali
   ```

2. **Start the application**
   ```bash
   ./start.sh
   ```

   This script will:
   - Start Docker containers (PostgreSQL & Redis)
   - Initialize the database with all migrations
   - Start the Go backend server
   - Display access URLs and credentials

3. **Access the application**
   - **Main Page:** http://localhost:8080
   - **API Health:** http://localhost:8080/api/health
   - **Creator Portal:** http://localhost:8080/creator/login.html
   - **Admin Panel:** http://localhost:8080/admin/login.html
     - Email: `admin@zenbali.org`
     - Password: `admin123`

4. **Stop the application**
   ```bash
   ./stop.sh
   ```

### Manual Setup

If you prefer to run services individually:

1. **Environment Variables**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

2. **Start Docker Services**
   ```bash
   docker-compose up -d
   ```

3. **Run Migrations**
   ```bash
   cd backend
   cat internal/database/migrations/001_init.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali
   cat internal/database/migrations/002_seed_data.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali
   cat internal/database/migrations/003_add_participant_group_and_lead_by.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali
   ```

4. **Start Backend Server**
   ```bash
   go run ./cmd/server
   ```

---

## Environment Configuration

### Development (.env)

```env
# Server Configuration
PORT=8080
ENV=development
BASE_URL=http://localhost:8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=zenbali
DB_PASSWORD=zenbali_dev_password
DB_NAME=zenbali
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5

# JWT Configuration
JWT_SECRET=zenbali-dev-secret-key-change-in-production-min-32-chars
JWT_EXPIRY_HOURS=24

# Stripe Configuration (use test keys)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_CENTS=1000

# Local Storage
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=5

# Admin Configuration
ADMIN_EMAIL=admin@zenbali.org
ADMIN_PASSWORD=admin123
```

---

## Database Schema

### Main Tables

**events** - Stores all event information
- Basic info: title, date, time, duration, location, type
- Financial: entrance_type_id, entrance_fee
- **New fields:**
  - `participant_group_type` - Couples, Females Only, Males Only, Open
  - `lead_by` - Event leader/instructor name
- Status: is_paid, is_published
- Media: image_url

**creators** - Event organizers
- name, organization_name, email, password_hash
- is_verified, is_active

**locations** - Bali areas (Ubud, Canggu, Seminyak, etc.)

**event_types** - Categories (Yoga, Meditation, Workshop, etc.)

**entrance_types** - Free, Paid, Donation, etc.

**payments** - Stripe payment records

**admins** - Platform administrators

**sessions** - JWT session tokens

**visitors** - Visitor tracking statistics

---

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/events` | List all published events (with filters) |
| GET | `/api/events/{id}` | Get single event details |
| GET | `/api/locations` | List all locations |
| GET | `/api/event-types` | List all event types |
| GET | `/api/entrance-types` | List entrance fee types |
| POST | `/api/visitors` | Track visitor (for stats) |
| GET | `/api/visitors/stats` | Get visitor statistics |

### Creator Endpoints (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/creator/register` | Register new creator account |
| POST | `/api/creator/login` | Login to creator account |
| GET | `/api/creator/events` | List creator's events |
| POST | `/api/creator/events` | Create new event |
| GET | `/api/creator/events/{id}` | Get event details |
| PUT | `/api/creator/events/{id}` | Update event |
| DELETE | `/api/creator/events/{id}` | Delete event |
| POST | `/api/creator/events/{id}/upload` | Upload event image |
| POST | `/api/creator/events/{id}/pay` | Create Stripe payment session |
| POST | `/api/stripe/webhook` | Handle Stripe webhooks |

### Admin Endpoints (Admin Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/login` | Admin login |
| GET | `/api/admin/dashboard` | Dashboard statistics |
| GET | `/api/admin/events` | List all events |
| GET | `/api/admin/creators` | List all creators |
| POST | `/api/admin/locations` | Add new location |
| POST | `/api/admin/event-types` | Add new event type |

---

## Event Creation Fields

### Required Fields
- **Title** - Event name (min 3, max 255 chars)
- **Event Date** - Date of the event
- **Location** - Select from Bali locations
- **Event Type** - Category (Yoga, Workshop, etc.)
- **Entrance Type** - Free, Paid, Donation, etc.
- **Contact Email** - For attendee inquiries

### Optional Fields
- **Event Time** - Time in 15-minute increments (00:00 - 23:45)
- **Duration** - Breakdown in Days/Hours/Minutes (15-min increments)
- **Entrance Fee** - Rupiah breakdown:
  - 10K-90K selector (10,000 - 90,000)
  - 1K-9K selector (1,000 - 9,000)
  - 100-900 selector
  - 10-90 selector
  - Verbatim field for Rp 100 million+
- **Participant Group Type** - Couples, Females Only, Males Only, Open
- **Lead By** - Event leader/instructor name (max 255 chars)
- **Contact Mobile** - Phone number
- **Event Description** - Detailed notes (max 2000 chars)

---

## Troubleshooting

### Port 5432 Already in Use

If you have a local PostgreSQL instance running:
```bash
# Stop local PostgreSQL
brew services stop postgresql@14
# OR
killall postgres

# Then restart the app
./start.sh
```

### Database Connection Failed

```bash
# Check if containers are running
docker ps

# Restart containers
docker-compose down
docker-compose up -d

# Check logs
docker logs zenbali-postgres
```

### Server Won't Start

```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill process using port 8080
lsof -ti:8080 | xargs kill -9

# Check server logs
tail -f server.log
```

---

## Development Workflow

1. **Create a new branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test locally**
   ```bash
   ./start.sh
   # Test your changes
   ./stop.sh
   ```

3. **Commit and push**
   ```bash
   git add .
   git commit -m "Description of changes"
   git push origin feature/your-feature-name
   ```

4. **Create Pull Request** on GitHub

---

## Deployment

See [reference/deployment.md](reference/deployment.md) for detailed deployment instructions to Google Cloud Platform.

---

## License

Proprietary. All rights reserved.

---

## Support

For issues or questions:
- **GitHub Issues:** https://github.com/net1io/zenbali/issues
- **Email:** support@net1io.com
- **Website:** https://net1io.com

---

**Developed with ❤️ by net1io.com**
