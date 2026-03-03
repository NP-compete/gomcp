# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/gomcp ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S gomcp && \
    adduser -u 1001 -S gomcp -G gomcp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/gomcp /app/gomcp
COPY --from=builder /app/assets /app/assets

# Set ownership
RUN chown -R gomcp:gomcp /app

# Switch to non-root user
USER gomcp

# Expose port
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/app/gomcp"]

