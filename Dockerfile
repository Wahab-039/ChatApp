# Stage 1: Build the Go application
FROM golang:1.26-alpine AS builder

# Install build dependencies (gcc, musl-dev for CGO)
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary, -ldflags="-s -w" strips debug info for smaller size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ChatApp .

# Stage 2: Create minimal runtime image
FROM alpine:3.20

# Install ca-certificates for HTTPS calls
RUN apk add --no-cache ca-certificates || true

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /build/ChatApp .

# Copy migrations directory (needed for goose)
COPY --from=builder /build/migrations ./migrations

# Copy entrypoint script
COPY deploy/api/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Health check - simple HTTP check (alpine has wget built-in in newer versions, fallback to nc)
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD nc -z localhost 8080 || exit 1

# Use entrypoint script to handle migrations
ENTRYPOINT ["./entrypoint.sh"]
