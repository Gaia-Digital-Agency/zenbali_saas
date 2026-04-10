#!/bin/bash
# Rebuild and restart the zenbali server

set -e

GOTAR="/var/www/zenbali/go.tar.gz"
GODIR="/var/www/zenbali/go"
PROJECTDIR="/var/www/zenbali"

# Extract Go if not already done
if [ ! -f "$GODIR/bin/go" ]; then
    echo "Extracting Go..."
    tar xzf "$GOTAR" -C "$PROJECTDIR/"
    echo "Go extracted: $($GODIR/bin/go version)"
fi

# Build the server
echo "Building server..."
cd "$PROJECTDIR/backend"
"$GODIR/bin/go" build -o "$PROJECTDIR/bin/zenbali-server" ./cmd/server
echo "Build complete"

# Stop old server
echo "Stopping old server..."
pkill -f zenbali-server || true
sleep 2

# Start new server
echo "Starting server..."
cd "$PROJECTDIR"
source .env
export $(cat .env | grep -v '^#' | xargs)
nohup "$PROJECTDIR/bin/zenbali-server" > "$PROJECTDIR/server.log" 2>&1 &
echo "Server started with PID: $!"
echo "Done"
