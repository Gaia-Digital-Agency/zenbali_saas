#!/bin/bash

# ===========================================
# Zen Bali - Local Development Stop Script
# ===========================================

echo "🛑 Stopping Zen Bali Development Environment..."

# Stop backend server
echo "   Stopping backend server..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
pkill -f "go run ./cmd/server" 2>/dev/null || true

# Stop Docker containers
echo "   Stopping Docker containers..."
docker-compose down

echo ""
echo "✅ Zen Bali stopped successfully"
echo ""
echo "💡 To start again, run: ./start.sh"
