FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and ca-certificates (required for private repos)
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# Create final lightweight image
FROM alpine:latest

WORKDIR /app

# Copy the binary
COPY --from=builder /app/bin/api /app/api

# Copy templates directory
COPY templates/ /app/templates/

# Expose the API port
EXPOSE 8000

# Run the application
CMD ["/app/api"] 