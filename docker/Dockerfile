# Build stage
FROM golang:1.22-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# Final stage
FROM alpine:3.19

# Add non root user
RUN adduser -D -g '' appuser

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/api .

# Copy any templates or static files if needed
COPY --from=builder /app/templates ./templates

# Use non root user
USER appuser

# Expose the application port
EXPOSE 8000

# Set environment variables
ENV APP_API_HOST=:8000 \
  APP_API_READ_TIMEOUT=5s \
  APP_API_WRITE_TIMEOUT=10s \
  APP_API_IDLE_TIMEOUT=10s \
  APP_API_SHUTDOWN_TIMEOUT=30s

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8000/readiness || exit 1

# Run the application
CMD ["./api"] 