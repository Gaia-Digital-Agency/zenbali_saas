# Zen Bali Database Access Guide

## Database Status
✅ **Database is SET UP and RUNNING**

## Database Configuration

### Connection Details
- **Database Type**: PostgreSQL
- **Host**: localhost
- **Port**: 5432
- **Database Name**: zenbali
- **Username**: zenbali
- **Password**: zenbali_dev_password
- **SSL Mode**: disabled (development)

### Connection String
```
postgres://zenbali:zenbali_dev_password@localhost:5432/zenbali?sslmode=disable
```

## Accessing the Database

### Using psql (Command Line)

#### 1. Connect to Database
```bash
psql -h localhost -p 5432 -U zenbali -d zenbali
```

When prompted for password, enter: `zenbali_dev_password`

Or use this one-liner (password in command):
```bash
PGPASSWORD=zenbali_dev_password psql -h localhost -p 5432 -U zenbali -d zenbali
```

#### 2. Common psql Commands
Once connected:
```sql
\dt                    -- List all tables
\d table_name          -- Describe a table structure
\du                    -- List database users
\l                     -- List all databases
\q                     -- Quit psql
```

### Using GUI Tools

#### Option 1: pgAdmin
- **Download**: https://www.pgadmin.org/download/
- **Host**: localhost
- **Port**: 5432
- **Database**: zenbali
- **Username**: zenbali
- **Password**: zenbali_dev_password

#### Option 2: TablePlus
- **Download**: https://tableplus.com/
- **Connection Type**: PostgreSQL
- **Host**: localhost
- **Port**: 5432
- **Database**: zenbali
- **User**: zenbali
- **Password**: zenbali_dev_password

#### Option 3: DBeaver
- **Download**: https://dbeaver.io/download/
- Same connection details as above

## Database Schema

### Tables Overview

| Table Name | Purpose | Record Count |
|------------|---------|--------------|
| `locations` | Bali area locations (25 pre-seeded) | 25 |
| `event_types` | Event categories (25 pre-seeded) | 25 |
| `entrance_types` | Entry fee types (6 pre-seeded) | 6 |
| `creators` | Event organizers/creators | Variable |
| `admins` | Admin users | Variable |
| `events` | Event listings | Variable |
| `payments` | Payment records (Stripe) | Variable |
| `sessions` | User authentication sessions | Variable |
| `visitors` | Visitor tracking for stats | Variable |

### Core Tables Detail

#### 1. Events Table
Main table for event listings:
```sql
SELECT id, title, event_date, location_id, event_type_id, is_published
FROM events
LIMIT 10;
```

#### 2. Creators Table
Event organizers:
```sql
SELECT id, name, email, organization_name, is_verified, is_active
FROM creators;
```

#### 3. Reference Tables
Pre-seeded data:
```sql
-- View all locations
SELECT * FROM locations ORDER BY name;

-- View all event types
SELECT * FROM event_types ORDER BY name;

-- View all entrance types
SELECT * FROM entrance_types;
```

## Sample Queries

### View Events with Related Data
```sql
SELECT
    e.id,
    e.title,
    e.event_date,
    l.name as location,
    et.name as event_type,
    ent.name as entrance_type,
    c.name as creator_name,
    e.is_published
FROM events e
JOIN locations l ON e.location_id = l.id
JOIN event_types et ON e.event_type_id = et.id
JOIN entrance_types ent ON e.entrance_type_id = ent.id
JOIN creators c ON e.creator_id = c.id
ORDER BY e.event_date DESC;
```

### Check Event Statistics
```sql
SELECT
    COUNT(*) as total_events,
    COUNT(CASE WHEN is_published = true THEN 1 END) as published_events,
    COUNT(CASE WHEN is_paid = true THEN 1 END) as paid_events
FROM events;
```

### View Payments
```sql
SELECT
    p.id,
    e.title as event_title,
    c.name as creator_name,
    p.amount_cents / 100.0 as amount_dollars,
    p.status,
    p.created_at
FROM payments p
JOIN events e ON p.event_id = e.id
JOIN creators c ON p.creator_id = c.id
ORDER BY p.created_at DESC;
```

## Migration Files

Migrations are located in:
```
backend/internal/database/migrations/
```

Files:
- `001_init.up.sql` - Creates all tables and indexes
- `001_init.down.sql` - Rollback for initial schema
- `002_seed_data.up.sql` - Seeds reference data (locations, event types, entrance types)
- `002_seed_data.down.sql` - Rollback for seed data

## Database Configuration in Code

Configuration is loaded from `.env` file via [backend/internal/config/config.go:64-72](backend/internal/config/config.go#L64-L72)

The connection string (DSN) is built at [backend/internal/config/config.go:103-108](backend/internal/config/config.go#L103-L108)

## Environment Variables

All database settings are in `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=zenbali
DB_PASSWORD=zenbali_dev_password
DB_NAME=zenbali
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
```

## Common Tasks

### Create a Test Admin User
```sql
INSERT INTO admins (email, password_hash, name)
VALUES ('admin@zenbali.org', 'hashed_password_here', 'Admin User');
```

### Create a Test Creator
```sql
INSERT INTO creators (name, email, password_hash, is_verified)
VALUES ('Test Creator', 'creator@test.com', 'hashed_password_here', true);
```

### View Recent Visitors
```sql
SELECT COUNT(*) as visitor_count, DATE(visited_at) as visit_date
FROM visitors
GROUP BY DATE(visited_at)
ORDER BY visit_date DESC
LIMIT 30;
```

## Troubleshooting

### Cannot Connect
1. Ensure PostgreSQL is running:
   ```bash
   brew services list | grep postgresql
   # or
   ps aux | grep postgres
   ```

2. Check if database exists:
   ```bash
   psql -l | grep zenbali
   ```

3. Verify user exists:
   ```bash
   psql postgres -c "\du"
   ```

### Reset Database
To completely reset and rebuild:
```bash
# Drop database
dropdb -h localhost -U zenbali zenbali

# Recreate database
createdb -h localhost -U zenbali zenbali

# Run migrations
# (using your Go migration tool)
```

## Security Notes

⚠️ **WARNING**: The credentials in `.env` are for DEVELOPMENT ONLY.

- Never commit production credentials
- Change all passwords in production
- Enable SSL in production
- Use environment variables or secret management in production
- The current admin password (`admin123`) should be changed immediately

## Additional Resources

- PostgreSQL Documentation: https://www.postgresql.org/docs/
- Connection pooling is configured with max 25 connections
- Timezone support is enabled (all timestamps use `TIMESTAMP WITH TIME ZONE`)
- UUID support is enabled via `uuid-ossp` extension
