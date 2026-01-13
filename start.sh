#!/bin/bash

# ===========================================
# Zen Bali - Local Development Startup Script
# ===========================================

set -e  # Exit on any error

echo "🌴 Starting Zen Bali Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    exit 1
fi

# Stop any local PostgreSQL that might conflict with port 5432
echo "🔍 Checking for conflicting PostgreSQL instances..."
if lsof -i :5432 | grep -q postgres | grep -v docker; then
    echo "⚠️  Found local PostgreSQL on port 5432. Stopping it..."
    killall postgres 2>/dev/null || true
    sleep 2
fi

# Start Docker containers
echo "🐳 Starting Docker containers (PostgreSQL & Redis)..."
docker-compose up -d

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 5

MAX_RETRIES=30
RETRY_COUNT=0
until docker exec zenbali-postgres pg_isready -U zenbali -d zenbali > /dev/null 2>&1; do
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
TABLE_COUNT=$(docker exec -i zenbali-postgres psql -U zenbali -d zenbali -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null || echo "0")

if [ "$TABLE_COUNT" -lt 5 ]; then
    echo "📦 Initializing database with migrations..."

    cd backend

    # Run migrations
    cat internal/database/migrations/001_init.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali > /dev/null 2>&1
    echo "   ✓ Schema migration applied"

    cat internal/database/migrations/002_seed_data.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali > /dev/null 2>&1
    echo "   ✓ Seed data loaded"

    cat internal/database/migrations/003_add_participant_group_and_lead_by.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali > /dev/null 2>&1
    echo "   ✓ New fields migration applied"

    cd ..
else
    echo "✅ Database already initialized"
fi

# Start the backend server
echo "🚀 Starting backend server..."
cd backend

# Kill any existing Go server on port 8080
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

# Start server in background
nohup go run ./cmd/server > ../server.log 2>&1 &
SERVER_PID=$!

echo "⏳ Waiting for server to start..."
sleep 3

# Check if server is running
if ps -p $SERVER_PID > /dev/null; then
    echo "✅ Backend server started (PID: $SERVER_PID)"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🎉 Zen Bali is ready!"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📍 API:              http://localhost:8080"
    echo "📍 Frontend:         http://localhost:8080"
    echo "📍 Health Check:     http://localhost:8080/api/health"
    echo ""
    echo "🔐 Admin Login:      http://localhost:8080/admin/login.html"
    echo "   Email:            admin@zenbali.org"
    echo "   Password:         admin123"
    echo ""
    echo "👤 Creator Login:    http://localhost:8080/creator/login.html"
    echo ""
    echo "📋 Server logs:      tail -f server.log"
    echo "🛑 Stop server:      ./stop.sh"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
else
    echo "❌ Failed to start backend server. Check server.log for details."
    tail -20 ../server.log
    exit 1
fi

cd ..
