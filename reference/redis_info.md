# Zen Bali Redis Information

## Current Status
⚠️ **Redis is RUNNING but NOT CURRENTLY USED in the application**

### Container Status
- **Container Name**: `zenbali-redis`
- **Image**: `redis:7-alpine`
- **Status**: Up and running ✅
- **Port**: `6379` (exposed to host)
- **Health**: Responding to PING ✅
- **Data Volume**: `app_zenbali_redis_data`

### Why Redis is Running but Not Used

The Docker Compose setup includes Redis as a service, but the application code does **not currently integrate with Redis**. This is evidenced by:

1. **No Redis client library** in `go.mod`
   - Missing: `github.com/redis/go-redis/v9` or similar

2. **No Redis configuration** in `.env`
   - No `REDIS_HOST`, `REDIS_PORT`, or `REDIS_PASSWORD` variables

3. **No Redis code** in the backend
   - No Redis client initialization
   - No caching logic
   - No session storage in Redis

4. **Current authentication uses PostgreSQL**
   - JWT tokens are generated in-memory
   - Session metadata is stored in PostgreSQL `sessions` table
   - No Redis caching layer

### Current Authentication Architecture

**How it works now:**
```
User Login
    ↓
Auth Service validates credentials (PostgreSQL)
    ↓
Generate JWT token (in-memory, signed with JWT_SECRET)
    ↓
Store session metadata in PostgreSQL sessions table
    ↓
Return JWT token to client
    ↓
Client sends JWT in Authorization header
    ↓
Server validates JWT signature (no database lookup needed)
```

**Sessions Table Structure:**
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_type VARCHAR(20) NOT NULL,  -- 'creator' or 'admin'
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Why Redis Was Included

Redis was likely included in the Docker Compose setup for **future expansion**. It's a common practice to include Redis early in the stack for future needs.

## Recommended Use Cases for Redis

If you want to utilize Redis in the future, here are recommended use cases for the Zen Bali platform:

### 1. Session Storage (Alternative to PostgreSQL)

**Current**: Sessions stored in PostgreSQL
**Future with Redis**: Store active sessions in Redis for faster lookups

**Benefits:**
- Faster session validation (in-memory vs disk I/O)
- Automatic expiration with TTL
- Reduced database load

**Example Implementation:**
```go
// Store session in Redis
SETEX session:{token_hash} 86400 "{user_id}:{user_type}"

// Validate session
GET session:{token_hash}

// Logout (revoke session)
DEL session:{token_hash}
```

### 2. Rate Limiting

**Use Case**: Prevent abuse of API endpoints (login attempts, event posting, etc.)

**Benefits:**
- Protect against brute force attacks
- Limit API calls per user/IP
- Prevent spam

**Example Implementation:**
```go
// Rate limit login attempts
INCR ratelimit:login:{ip_address}
EXPIRE ratelimit:login:{ip_address} 3600
// If count > 5, block for 1 hour
```

### 3. Caching Event Listings

**Use Case**: Cache frequently accessed event data

**Benefits:**
- Faster response times for public event listings
- Reduced database queries
- Better performance under high load

**Example Implementation:**
```go
// Cache public events for 5 minutes
SET events:public:page:1 "{json_data}" EX 300

// Cache event details for 10 minutes
SET event:{event_id} "{json_data}" EX 600

// Invalidate cache when event is updated
DEL event:{event_id}
DEL events:public:*
```

### 4. Real-time Statistics Caching

**Use Case**: Cache dashboard statistics (visitor counts, event counts, etc.)

**Benefits:**
- Instant dashboard load times
- Reduced load on PostgreSQL for analytics queries
- Can update cache periodically

**Example Implementation:**
```go
// Cache stats for 1 hour
SETEX stats:dashboard "{total_events}:{total_creators}:{today_visitors}" 3600

// Increment visitor count in real-time
INCR stats:visitors:today
```

### 5. Event Search Cache

**Use Case**: Cache search results for common queries

**Benefits:**
- Faster search response
- Better user experience
- Reduced database load for popular searches

**Example Implementation:**
```go
// Cache search results for location "ubud"
SET search:location:ubud "{json_results}" EX 300

// Cache search results for event type "yoga"
SET search:type:yoga "{json_results}" EX 300
```

### 6. Job Queue for Background Tasks

**Use Case**: Queue tasks like email notifications, image processing, etc.

**Benefits:**
- Asynchronous processing
- Better scalability
- Resilience to failures

**Example Implementation:**
```go
// Add job to queue
LPUSH queue:emails "{email_data}"

// Worker processes jobs
BRPOP queue:emails 0
```

### 7. Temporary Data Storage

**Use Case**: Store temporary data like password reset tokens, email verification codes

**Benefits:**
- Auto-expiration with TTL
- No database cleanup needed
- Fast access

**Example Implementation:**
```go
// Store password reset token for 1 hour
SETEX reset_token:{token} 3600 "{user_id}"

// Store email verification code for 15 minutes
SETEX verify_email:{code} 900 "{email}"
```

## How to Enable Redis in Your Application

If you decide to use Redis, follow these steps:

### 1. Add Redis Client Library

```bash
cd backend
go get github.com/redis/go-redis/v9
```

This will add to `go.mod`:
```go
require github.com/redis/go-redis/v9 v9.x.x
```

### 2. Add Redis Configuration to `.env`

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

### 3. Update Config Structure

Edit [backend/internal/config/config.go](backend/internal/config/config.go):

```go
type Config struct {
    Port     string
    Env      string
    BaseURL  string
    Database DatabaseConfig
    Redis    RedisConfig    // Add this
    JWT      JWTConfig
    Stripe   StripeConfig
    Upload   UploadConfig
    Admin    AdminConfig
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
    PoolSize int
}
```

### 4. Create Redis Client

Create `backend/internal/cache/redis.go`:

```go
package cache

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
    "github.com/net1io/zenbali/internal/config"
)

type RedisClient struct {
    client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
        PoolSize: cfg.PoolSize,
    })

    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis connection failed: %w", err)
    }

    return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
    return r.client.Close()
}

func (r *RedisClient) Client() *redis.Client {
    return r.client
}
```

### 5. Initialize in Main

Edit [backend/cmd/server/main.go](backend/cmd/server/main.go):

```go
// Initialize Redis
redisClient, err := cache.NewRedisClient(cfg.Redis)
if err != nil {
    log.Fatal("Failed to connect to Redis:", err)
}
defer redisClient.Close()

log.Println("✓ Redis connected")
```

## Accessing Redis

### Using redis-cli

```bash
# Connect to Redis container
docker exec -it zenbali-redis redis-cli

# Or from host (if redis-cli is installed)
redis-cli -h localhost -p 6379
```

### Common Redis Commands

```bash
# Test connection
PING
# Response: PONG

# Set a key
SET mykey "Hello Redis"

# Get a key
GET mykey

# Set with expiration (10 seconds)
SETEX mykey 10 "Hello"

# Check remaining TTL
TTL mykey

# List all keys (use cautiously in production!)
KEYS *

# Get number of keys
DBSIZE

# Delete a key
DEL mykey

# Flush all data (CAUTION!)
FLUSHALL

# Get Redis info
INFO
INFO memory
INFO stats

# Monitor all commands in real-time
MONITOR
```

### Using Redis from Go

```go
import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

ctx := context.Background()

// Set a value
err := redisClient.Set(ctx, "key", "value", 0).Err()

// Set with expiration
err := redisClient.Set(ctx, "key", "value", 5*time.Minute).Err()

// Get a value
val, err := redisClient.Get(ctx, "key").Result()

// Check if key exists
exists, err := redisClient.Exists(ctx, "key").Result()

// Delete a key
err := redisClient.Del(ctx, "key").Err()

// Increment a counter
count, err := redisClient.Incr(ctx, "counter").Result()

// Set expiration on existing key
err := redisClient.Expire(ctx, "key", 10*time.Minute).Err()

// Get TTL
ttl, err := redisClient.TTL(ctx, "key").Result()
```

## Current Docker Setup

### docker-compose.yml Configuration

```yaml
redis:
  image: redis:7-alpine
  container_name: zenbali-redis
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
```

### Starting/Stopping Redis

```bash
# Start Redis (via docker-compose)
docker-compose up -d redis

# Or start all services
make docker-up

# Stop Redis
docker-compose stop redis

# View Redis logs
docker logs zenbali-redis -f

# Restart Redis
docker-compose restart redis
```

## Monitoring Redis

### Check Memory Usage

```bash
docker exec zenbali-redis redis-cli INFO memory
```

### View Connected Clients

```bash
docker exec zenbali-redis redis-cli CLIENT LIST
```

### Monitor Commands in Real-time

```bash
docker exec -it zenbali-redis redis-cli MONITOR
```

### Check Performance Stats

```bash
docker exec zenbali-redis redis-cli INFO stats
```

## Data Persistence

### Current Setup

Redis data is persisted in Docker volume:
- **Volume Name**: `app_zenbali_redis_data`
- **Mount Point**: `/data` inside container
- **Persistence**: RDB snapshots (default Redis persistence)

### Backup Redis Data

```bash
# Create a snapshot
docker exec zenbali-redis redis-cli BGSAVE

# Export data
docker exec zenbali-redis redis-cli --rdb /data/dump.rdb

# Copy to host
docker cp zenbali-redis:/data/dump.rdb ./redis-backup.rdb
```

### Restore Redis Data

```bash
# Stop Redis
docker-compose stop redis

# Copy backup to volume
docker cp ./redis-backup.rdb zenbali-redis:/data/dump.rdb

# Start Redis
docker-compose start redis
```

## Production Considerations

### When to Use Redis in Production

Use Redis when you need:
- ✅ High-performance caching
- ✅ Session management at scale
- ✅ Rate limiting
- ✅ Real-time features
- ✅ Job queues

### Managed Redis Services

For production, consider managed Redis:

**Google Cloud (Recommended for this project):**
- **Cloud Memorystore for Redis**
- Fully managed
- High availability
- Automatic failover
- VPC integration
- Pricing: ~$50-200/month depending on size

**Other Options:**
- AWS ElastiCache
- Azure Cache for Redis
- Redis Enterprise Cloud
- Upstash (serverless)

### Security Best Practices

1. **Enable Authentication**
   ```bash
   # In production, always use password
   requirepass your-strong-password-here
   ```

2. **Use Private Networks**
   - Never expose Redis to public internet
   - Use VPC/private networking

3. **Disable Dangerous Commands**
   ```bash
   rename-command FLUSHDB ""
   rename-command FLUSHALL ""
   rename-command CONFIG ""
   ```

4. **Enable SSL/TLS**
   - Use encrypted connections in production

5. **Set Memory Limits**
   ```bash
   maxmemory 256mb
   maxmemory-policy allkeys-lru
   ```

## Summary & Recommendations

### Current State
- ✅ Redis container is running
- ❌ Not integrated into application code
- ❌ Not being utilized

### Recommendations

**Option 1: Remove Redis (If not needed)**
If you don't plan to use Redis soon, you can remove it to reduce resource usage:
```bash
# Stop and remove Redis
docker-compose stop redis
docker-compose rm redis
docker volume rm app_zenbali_redis_data

# Update docker-compose.yml to remove redis service
```

**Option 2: Implement Redis (Recommended)**
If you want to improve performance, implement Redis for:
1. **Start Simple**: Event listing cache (easiest, immediate benefit)
2. **Add Rate Limiting**: Protect your API
3. **Session Management**: Faster authentication
4. **Real-time Stats**: Dashboard performance

**Option 3: Keep for Future**
Keep Redis running as-is, ready for when you need it. The resource overhead is minimal (~10-20MB RAM).

### Next Steps

If you decide to implement Redis:
1. Add `github.com/redis/go-redis/v9` to dependencies
2. Create cache package with Redis client
3. Start with simple event listing cache
4. Gradually add more features
5. Monitor performance improvements

## Additional Resources

- **Redis Documentation**: https://redis.io/docs/
- **go-redis Documentation**: https://redis.uptrace.dev/
- **Redis Best Practices**: https://redis.io/docs/management/optimization/
- **Redis Persistence**: https://redis.io/docs/management/persistence/
- **Google Cloud Memorystore**: https://cloud.google.com/memorystore
