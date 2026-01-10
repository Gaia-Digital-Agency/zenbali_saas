# Zen Bali Docker Configuration Guide

## Docker Status
✅ **Docker services are RUNNING**

### Active Containers
| Container Name | Image | Status | Ports | Purpose |
|---------------|-------|--------|-------|---------|
| `zenbali-postgres` | postgres:15-alpine | Up 21 mins (healthy) | 0.0.0.0:5432->5432 | PostgreSQL Database |
| `zenbali-redis` | redis:7-alpine | Up 21 mins | 0.0.0.0:6379->6379 | Redis Cache |

## Docker Architecture

### Services Overview

The Zen Bali project uses Docker Compose to manage two essential services:

1. **PostgreSQL Database** - Persistent data storage
2. **Redis Cache** - Session management and caching

The application itself (Go backend) can run:
- **Natively** on your Mac (recommended for development)
- **In Docker** for production deployment

## Configuration Files

### 1. docker-compose.yml
Location: [docker-compose.yml](docker-compose.yml)

This file defines the development environment with two services:

```yaml
services:
  postgres:
    - Image: postgres:15-alpine
    - Container: zenbali-postgres
    - Port: 5432 (exposed to host)
    - Health check enabled
    - Data persisted in volume

  redis:
    - Image: redis:7-alpine
    - Container: zenbali-redis
    - Port: 6379 (exposed to host)
    - Data persisted in volume
```

### 2. Dockerfile
Location: [Dockerfile](Dockerfile)

Multi-stage build for production deployment:

**Stage 1: Builder**
- Base: `golang:1.22-alpine`
- Compiles Go backend into optimized binary
- Static binary (CGO_ENABLED=0) for portability

**Stage 2: Runtime**
- Base: `alpine:3.19`
- Minimal image (~20MB + app)
- Non-root user (zenbali:1001)
- Includes frontend files and migrations
- Health check endpoint: `/api/health`

### 3. Database Initialization
Location: [scripts/init.sql](scripts/init.sql)

Runs automatically when PostgreSQL container first starts:
- Grants database privileges
- Schema creation is handled by Go migrations

## Docker Commands

### Using Makefile (Recommended)

The project includes a Makefile with convenient commands:

```bash
# Start Docker services
make docker-up

# Stop Docker services
make docker-down

# Complete local setup (Docker + migrations + seed data)
make setup
```

### Using Docker Compose Directly

```bash
# Start services in background
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f postgres
docker-compose logs -f redis

# Restart services
docker-compose restart

# Remove containers and volumes (CAUTION: deletes data)
docker-compose down -v
```

### Using Docker CLI

```bash
# List running containers
docker ps

# View container logs
docker logs zenbali-postgres
docker logs zenbali-redis

# Execute commands in containers
docker exec -it zenbali-postgres psql -U zenbali -d zenbali
docker exec -it zenbali-redis redis-cli

# Check container health
docker inspect zenbali-postgres --format='{{.State.Health.Status}}'

# Stop specific container
docker stop zenbali-postgres
docker stop zenbali-redis

# Start stopped container
docker start zenbali-postgres
docker start zenbali-redis

# Remove container (must be stopped first)
docker rm zenbali-postgres
```

## Data Persistence

### Docker Volumes

Data is persisted in Docker volumes:

```bash
# List volumes
docker volume ls | grep zenbali

# Current volumes:
# - app_zenbali_postgres_data (PostgreSQL data)
# - app_zenbali_redis_data (Redis data)

# Inspect volume
docker volume inspect app_zenbali_postgres_data

# Remove volume (CAUTION: deletes all data)
docker volume rm app_zenbali_postgres_data
```

### Volume Locations

- **PostgreSQL Data**: Stored in `app_zenbali_postgres_data` volume
  - Maps to: `/var/lib/postgresql/data` inside container
  - Persists: All database tables, indexes, and data

- **Redis Data**: Stored in `app_zenbali_redis_data` volume
  - Maps to: `/data` inside container
  - Persists: Cached data and sessions

- **Init Script**: `./scripts/init.sql` is mounted on first start

## Network Configuration

### Docker Network
- **Network Name**: `app_zenbali_default`
- **Type**: Bridge network
- **Purpose**: Allows containers to communicate

### Service Communication

**From Host Machine:**
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

**Between Containers:**
- PostgreSQL: `postgres:5432`
- Redis: `redis:6379`

**From Go Application (running on host):**
- PostgreSQL: `localhost:5432` (configured in `.env`)
- Redis: `localhost:6379`

## PostgreSQL Container Details

### Connection Information
```bash
# Connect from host
docker exec -it zenbali-postgres psql -U zenbali -d zenbali

# Or using psql directly
PGPASSWORD=zenbali_dev_password psql -h localhost -p 5432 -U zenbali -d zenbali
```

### Environment Variables
- `POSTGRES_USER`: zenbali
- `POSTGRES_PASSWORD`: zenbali_dev_password
- `POSTGRES_DB`: zenbali

### Health Check
The container runs health checks every 5 seconds:
```bash
pg_isready -U zenbali -d zenbali
```

Status: **healthy** ✅

### Useful Commands
```bash
# View database size
docker exec zenbali-postgres psql -U zenbali -d zenbali -c "\l+"

# Backup database
docker exec zenbali-postgres pg_dump -U zenbali zenbali > backup.sql

# Restore database
cat backup.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali

# View running queries
docker exec zenbali-postgres psql -U zenbali -d zenbali -c "SELECT * FROM pg_stat_activity;"
```

## Redis Container Details

### Connection Information
```bash
# Connect to Redis CLI
docker exec -it zenbali-redis redis-cli

# Test connection
docker exec zenbali-redis redis-cli ping
# Response: PONG ✅
```

### Useful Commands
```bash
# Inside redis-cli
INFO                    # Server information
DBSIZE                  # Number of keys
KEYS *                  # List all keys (use in dev only!)
GET key_name           # Get value of key
FLUSHALL               # Clear all data (CAUTION!)

# From host
docker exec zenbali-redis redis-cli INFO
docker exec zenbali-redis redis-cli DBSIZE
```

## Production Deployment

### Building Production Image

```bash
# Build the Docker image
docker build -t zenbali:latest .

# Or use the Makefile
make build-prod
```

### Image Features
- **Size**: Optimized minimal image (~50MB total)
- **Security**: Runs as non-root user (UID 1001)
- **Health Check**: Built-in health endpoint
- **Components Included**:
  - Go binary (static, optimized)
  - Frontend files
  - Database migrations
  - Uploads directory

### Google Cloud Run Deployment

The Makefile includes deployment to GCP Cloud Run:

```bash
# Set your GCP project
export GCP_PROJECT=your-project-id

# Deploy to Cloud Run
make deploy
```

This will:
1. Build Docker image
2. Push to Google Container Registry (GCR)
3. Deploy to Cloud Run in `asia-southeast1`
4. Configure as public (unauthenticated access)

### Environment Variables for Production

When deploying, set these environment variables:

```bash
# Database (use Cloud SQL or managed PostgreSQL)
DB_HOST=your-cloud-sql-host
DB_PORT=5432
DB_USER=zenbali
DB_PASSWORD=your-secure-password
DB_NAME=zenbali
DB_SSL_MODE=require

# Redis (use Cloud Memorystore or managed Redis)
REDIS_HOST=your-redis-host
REDIS_PORT=6379

# JWT (use strong secret)
JWT_SECRET=your-production-secret-min-64-chars

# Stripe (use production keys)
STRIPE_SECRET_KEY=sk_live_xxxxx
STRIPE_PUBLISHABLE_KEY=pk_live_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx

# Other
ENV=production
BASE_URL=https://your-domain.com
```

## Common Tasks

### 1. Start Fresh Development Environment

```bash
# Stop and remove everything
docker-compose down -v

# Start fresh
make setup

# This will:
# - Start Docker containers
# - Create directories
# - Run migrations
# - Seed reference data
```

### 2. View Real-time Logs

```bash
# All services
docker-compose logs -f

# PostgreSQL only
docker-compose logs -f postgres

# Redis only
docker-compose logs -f redis
```

### 3. Access Database Shell

```bash
# PostgreSQL
docker exec -it zenbali-postgres psql -U zenbali -d zenbali

# Redis
docker exec -it zenbali-redis redis-cli
```

### 4. Check Resource Usage

```bash
# Container stats (CPU, memory, network)
docker stats zenbali-postgres zenbali-redis

# Disk usage
docker system df
```

### 5. Troubleshooting Connection Issues

```bash
# Check if containers are running
docker ps | grep zenbali

# Check container health
docker inspect zenbali-postgres --format='{{.State.Health.Status}}'

# Check logs for errors
docker logs zenbali-postgres --tail 50
docker logs zenbali-redis --tail 50

# Test PostgreSQL connection
docker exec zenbali-postgres pg_isready -U zenbali -d zenbali

# Test Redis connection
docker exec zenbali-redis redis-cli ping

# Check network connectivity
docker network inspect app_zenbali_default
```

## Development Workflow

### Recommended Setup

1. **Docker for Services** (PostgreSQL, Redis)
   ```bash
   make docker-up
   ```

2. **Native Go for Application** (faster, easier debugging)
   ```bash
   make dev
   ```

This approach gives you:
- Fast rebuild times (Go compiler runs natively)
- Easy debugging (use Delve or VS Code)
- Full access to services (PostgreSQL, Redis)
- Hot reload (if using Air)

### Alternative: Full Docker Development

If you prefer running everything in Docker:

1. Modify `docker-compose.yml` to add your app service
2. Use Docker Compose to run all services together
3. Mount your code as a volume for live reloading

## Cleanup

### Remove Containers Only
```bash
docker-compose down
```

### Remove Containers and Volumes (deletes data)
```bash
docker-compose down -v
```

### Remove Unused Images
```bash
docker image prune
```

### Complete Docker Cleanup
```bash
# Remove all stopped containers
docker container prune

# Remove all unused images
docker image prune -a

# Remove all unused volumes
docker volume prune

# Remove all unused networks
docker network prune

# Nuclear option: Remove everything
docker system prune -a --volumes
```

## Health Monitoring

### Container Health Status

```bash
# Check health
docker inspect zenbali-postgres --format='{{.State.Health.Status}}'
# Output: healthy

# View health check logs
docker inspect zenbali-postgres --format='{{json .State.Health}}' | jq
```

### Application Health Check

The Dockerfile includes a health check that pings `/api/health`:

```bash
# Test health endpoint (when app is running)
curl http://localhost:8080/api/health
```

## Security Considerations

### Development
- ⚠️ Containers expose ports to `0.0.0.0` (all interfaces)
- ⚠️ Passwords are in plaintext in `docker-compose.yml`
- ⚠️ SSL is disabled for PostgreSQL
- ✅ Acceptable for local development only

### Production
- ✅ Use secrets management (GCP Secret Manager, AWS Secrets, etc.)
- ✅ Enable SSL for PostgreSQL (`DB_SSL_MODE=require`)
- ✅ Use private networks (no public IPs)
- ✅ Run as non-root user (already configured in Dockerfile)
- ✅ Keep images updated regularly
- ✅ Use specific image tags (not `latest`)

## Additional Resources

- **Docker Documentation**: https://docs.docker.com/
- **Docker Compose Reference**: https://docs.docker.com/compose/
- **PostgreSQL Docker Hub**: https://hub.docker.com/_/postgres
- **Redis Docker Hub**: https://hub.docker.com/_/redis
- **Google Cloud Run**: https://cloud.google.com/run/docs
