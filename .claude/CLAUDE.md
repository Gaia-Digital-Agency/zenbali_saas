# Zen Bali - Project Context for Claude

## Project Overview

Zen Bali is a SaaS events platform for Bali, Indonesia. Content creators post events (for $10 USD fee), visitors browse for free.

## Tech Stack

- **Backend:** Go 1.22+ with Chi router
- **Database:** PostgreSQL 16 (Docker)
- **Cache:** Redis (Docker)
- **Frontend:** Vanilla HTML/CSS/JS (no framework)
- **Payments:** Stripe
- **Auth:** JWT tokens

## Directory Structure

```
/backend           - Go backend (API server)
  /cmd/server      - Main entry point
  /internal        - Internal packages
    /handlers      - HTTP handlers
    /services      - Business logic
    /repository    - Database queries
    /models        - Data models
    /database      - DB connection & migrations
/frontend/public   - Static frontend files
  /admin           - Admin portal pages
  /creator         - Creator portal pages
  /js              - JavaScript files
  /css             - Stylesheets
```

## Key Files

- `start.sh` - Start development environment
- `stop.sh` - Stop development environment
- `docker-compose.yml` - PostgreSQL & Redis containers
- `.env` - Environment configuration
- `backend/cmd/server/main.go` - Server entry point & routes

## Development Commands

```bash
# Start app
./start.sh

# Stop app
./stop.sh

# View logs
tail -f server.log

# Database access
docker exec -it zenbali-postgres psql -U zenbali -d zenbali
```

## Test Accounts

**Admin:**
- URL: http://localhost:8080/admin/login.html
- Email: admin@zenbali.org
- Password: admin123

**Creator:**
- URL: http://localhost:8080/creator/login.html
- Email: creator@test.com
- Password: admin123

## API Base URL

- Development: http://localhost:8080/api

## Database

- Container: `zenbali-postgres`
- User: `zenbali`
- Database: `zenbali`
- Port: 5432

## Key Concepts

- **Events** must be `is_paid=true` AND `is_published=true` to appear on landing page
- Payment via Stripe sets both flags to true
- Landing page defaults to showing today's and tomorrow's events
- Images stored locally in `/uploads` directory

## Common Issues

1. **Port 5432 conflict:** Stop local PostgreSQL with `brew services stop postgresql@14`
2. **Login fails after restart:** Test accounts need to be recreated if Docker volumes are removed
3. **No events showing:** Events must have `is_paid=true` and `is_published=true`
