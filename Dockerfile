# ===========================================
# Zen Bali Dockerfile
# Multi-stage build for minimal production image
# ===========================================

# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY backend/go.mod backend/go.sum ./backend/

# Download dependencies
WORKDIR /app/backend
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/zenbali \
    ./cmd/server

# Stage 2: Create minimal production image
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S zenbali && \
    adduser -u 1001 -S zenbali -G zenbali

# Copy binary from builder
COPY --from=builder /app/zenbali /app/zenbali

# Copy frontend files
COPY frontend/ /app/frontend/

# Copy migration files
COPY backend/internal/database/migrations/ /app/migrations/

# Create uploads directory
RUN mkdir -p /app/uploads && chown -R zenbali:zenbali /app

# Switch to non-root user
USER zenbali

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the binary
CMD ["/app/zenbali"]
