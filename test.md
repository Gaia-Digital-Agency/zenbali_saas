# Zen Bali - Testing Instructions

## Prerequisites

- Docker Desktop installed and running
- Go 1.22+ installed
- Port 5432 available (no local PostgreSQL running)
- Port 8080 available

---

## Start the Application

```bash
./start.sh
```

This will:
1. Start PostgreSQL and Redis containers
2. Run database migrations
3. Start the Go backend server

**Expected output:**
```
🌴 Starting Zen Bali Development Environment...
🐳 Starting Docker containers (PostgreSQL & Redis)...
✅ Database is ready!
✅ Backend server started
🎉 Zen Bali is ready!
```

---

## Stop the Application

```bash
./stop.sh
```

This will:
1. Stop the backend server (port 8080)
2. Stop Docker containers
3. Stop any local PostgreSQL services on port 5432

---

## Test Accounts

### Admin Account

| Field | Value |
|-------|-------|
| URL | http://localhost:8080/admin/login.html |
| Email | `admin@zenbali.org` |
| Password | `admin123` |

**Admin can:**
- View dashboard statistics
- Manage all events
- Manage creators
- Add locations and event types

### Creator Account

| Field | Value |
|-------|-------|
| URL | http://localhost:8080/creator/login.html |
| Email | `creator@test.com` |
| Password | `admin123` |

**Creator can:**
- Create and manage events
- Upload event images
- Process payments (Stripe test mode)

---

## Application URLs

| Page | URL |
|------|-----|
| Landing Page | http://localhost:8080 |
| Health Check | http://localhost:8080/api/health |
| Creator Login | http://localhost:8080/creator/login.html |
| Creator Register | http://localhost:8080/creator/register.html |
| Creator Dashboard | http://localhost:8080/creator/dashboard.html |
| Admin Login | http://localhost:8080/admin/login.html |
| Admin Dashboard | http://localhost:8080/admin/dashboard.html |

---

## Sample Test Data

The database includes 2 sample paid and published events:

1. **Morning Yoga in Ubud** (Today)
   - Time: 06:30
   - Location: Ubud
   - Type: Yoga
   - Fee: IDR 150,000

2. **Sound Healing Meditation Workshop** (Tomorrow)
   - Time: 19:00
   - Location: Canggu
   - Type: Sound Healing
   - Fee: IDR 250,000

---

## Testing Workflows

### 1. View Public Events

1. Go to http://localhost:8080
2. Events from today and tomorrow should display
3. Use filters to search by location, type, date

### 2. Creator Registration

1. Go to http://localhost:8080/creator/register.html
2. Fill in registration form
3. Login with new credentials
4. Access creator dashboard

### 3. Create an Event

1. Login as creator
2. Click "Create Event"
3. Fill in event details
4. Save event (will be unpublished until paid)

### 4. Admin Management

1. Login as admin
2. View dashboard statistics
3. Browse all events and creators

---

## View Logs

```bash
# Server logs
tail -f server.log

# Docker container logs
docker logs zenbali-postgres
docker logs zenbali-redis
```

---

## Database Access

```bash
# Connect to PostgreSQL
docker exec -it zenbali-postgres psql -U zenbali -d zenbali

# Useful queries
SELECT * FROM events;
SELECT * FROM creators;
SELECT * FROM admins;
```

---

## Troubleshooting

### Can't login after restart

If Docker volumes were removed, test accounts need to be recreated:

```bash
# Generate password hash and insert admin
docker exec -i zenbali-postgres psql -U zenbali -d zenbali -c "INSERT INTO admins (email, password_hash, name) VALUES ('admin@zenbali.org', '\$2a\$10\$oWNHRvZQ9IKrsmElaYcYjumtJgrCJx87XkhbdzNxalNCR.4RpSaJ6', 'Admin') ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;"

# Insert creator
docker exec -i zenbali-postgres psql -U zenbali -d zenbali -c "INSERT INTO creators (name, organization_name, email, mobile, password_hash) VALUES ('Test Creator', 'Test Org', 'creator@test.com', '+628123456789', '\$2a\$10\$oWNHRvZQ9IKrsmElaYcYjumtJgrCJx87XkhbdzNxalNCR.4RpSaJ6') ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;"
```

### Port 5432 in use

```bash
# Stop local PostgreSQL
brew services stop postgresql@14
brew services stop postgresql@15
brew services stop postgresql@16
brew services stop postgresql

# Or kill process on port
lsof -ti:5432 | xargs kill -9
```

### Port 8080 in use

```bash
lsof -ti:8080 | xargs kill -9
```

### No events showing on landing page

Events must have:
- `is_paid = true`
- `is_published = true`
- `event_date` within filter range (default: today to tomorrow)

Check events in database:
```bash
docker exec -i zenbali-postgres psql -U zenbali -d zenbali -c "SELECT title, event_date, is_paid, is_published FROM events;"
```

---

## API Testing

### Health Check
```bash
curl http://localhost:8080/api/health
```

### List Events
```bash
curl "http://localhost:8080/api/events?date_from=2026-01-16&date_to=2026-01-17"
```

### Creator Login
```bash
curl -X POST http://localhost:8080/api/creator/login \
  -H "Content-Type: application/json" \
  -d '{"email":"creator@test.com","password":"admin123"}'
```

### Admin Login
```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@zenbali.org","password":"admin123"}'
```
