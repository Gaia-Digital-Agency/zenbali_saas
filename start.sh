#!/bin/bash

# ===========================================
# Zen Bali - Local Development Startup Script
# ===========================================

set -e  # Exit on any error

if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

APP_PORT="${PORT:-8081}"
DB_PORT="${DB_PORT:-5433}"
DB_USER="${DB_USER:-zenbali}"
DB_NAME="${DB_NAME:-zenbali}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@zenbali.org}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Teameditor@123}"
CREATOR_EMAIL="${CREATOR_EMAIL:-creator@zenbali.org}"
CREATOR_PASSWORD="${CREATOR_PASSWORD:-admin123}"
LOG_DIR="${LOG_DIR:-logs}"
SERVER_LOG="${SERVER_LOG:-$LOG_DIR/server.log}"

echo "🌴 Starting Zen Bali Development Environment..."

mkdir -p "$LOG_DIR"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    exit 1
fi

# Start Docker containers
echo "🐳 Starting Docker containers (PostgreSQL & Redis)..."
docker compose up -d

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 5

MAX_RETRIES=30
RETRY_COUNT=0
until docker exec zenbali-postgres pg_isready -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "❌ Database failed to start after $MAX_RETRIES attempts"
        exit 1
    fi
    echo "   Still waiting... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 1
done

echo "✅ Database is ready!"

# Check if database is initialized
echo "🔍 Checking database initialization..."
TABLE_COUNT=$(docker exec -i zenbali-postgres psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null || echo "0")

if [ "$TABLE_COUNT" -lt 5 ]; then
    echo "📦 Initializing database with migrations..."

    cd backend

    # Run migrations
    cat internal/database/migrations/001_init.up.sql | docker exec -i zenbali-postgres psql -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1
    echo "   ✓ Schema migration applied"

    cat internal/database/migrations/002_seed_data.up.sql | docker exec -i zenbali-postgres psql -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1
    echo "   ✓ Seed data loaded"

    cat internal/database/migrations/003_add_participant_group_and_lead_by.up.sql | docker exec -i zenbali-postgres psql -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1
    echo "   ✓ New fields migration applied"

    cd ..
else
    echo "✅ Database already initialized"
fi

# Start the backend server
echo "🚀 Starting backend server..."
cd backend

if lsof -ti:"$APP_PORT" > /dev/null 2>&1; then
    echo "❌ Port $APP_PORT is already in use. Free that port or update PORT in .env."
    exit 1
fi

# Start server in background
nohup go run ./cmd/server > "../$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "⏳ Waiting for server to start..."
MAX_SERVER_RETRIES=20
SERVER_RETRY_COUNT=0
until curl -fsS "http://localhost:$APP_PORT/api/health" > /dev/null 2>&1; do
    SERVER_RETRY_COUNT=$((SERVER_RETRY_COUNT + 1))
    if ! ps -p $SERVER_PID > /dev/null 2>&1; then
        echo "❌ Backend server exited during startup. Check $SERVER_LOG for details."
        tail -20 "../$SERVER_LOG"
        exit 1
    fi
    if [ $SERVER_RETRY_COUNT -ge $MAX_SERVER_RETRIES ]; then
        echo "❌ Backend server did not become healthy after $MAX_SERVER_RETRIES checks."
        tail -20 "../$SERVER_LOG"
        exit 1
    fi
    sleep 1
done

# Check if server is running
if ps -p $SERVER_PID > /dev/null 2>&1; then
    echo "✅ Backend server started (PID: $SERVER_PID)"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🎉 Zen Bali is ready!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📍 API:              http://localhost:$APP_PORT"
    echo "📍 Frontend:         http://localhost:$APP_PORT"
    echo "📍 Health Check:     http://localhost:$APP_PORT/api/health"
    echo ""
    echo "🔐 Admin Login:      http://localhost:$APP_PORT/admin/login.html"
    echo "   Email:            $ADMIN_EMAIL"
    echo "   Password:         $ADMIN_PASSWORD"
    echo ""
    echo "👤 Creator Login:    http://localhost:$APP_PORT/creator/login.html"
    echo "   Email:            $CREATOR_EMAIL"
    echo "   Password:         $CREATOR_PASSWORD"
    echo ""
    echo "📋 Server logs:      tail -f $SERVER_LOG"
    echo "🛑 Stop server:      ./stop.sh"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
else
    echo "❌ Failed to start backend server. Check $SERVER_LOG for details."
    tail -20 "../$SERVER_LOG"
    exit 1
fi

cd ..
