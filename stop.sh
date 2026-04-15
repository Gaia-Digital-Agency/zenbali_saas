#!/bin/bash

# ===========================================
# Zen Bali - Local Development Stop Script
# ===========================================

echo "🛑 Stopping Zen Bali Development Environment..."

# Stop backend server
echo "   Stopping backend server..."
pkill -f "go run ./cmd/server" 2>/dev/null || true

# Stop Docker containers
echo "   Stopping Docker containers..."
docker compose down > /dev/null 2>&1

# Force remove containers by name to prevent conflicts, silencing errors if they don't exist
echo "   Ensuring conflicting containers are removed..."
docker rm -f zenbali-postgres zenbali-redis > /dev/null 2>&1 || true

echo ""
echo "✅ Zen Bali stopped successfully"
echo ""
echo "💡 To start again, run: ./start.sh"
