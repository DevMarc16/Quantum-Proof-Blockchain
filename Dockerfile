# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o quantum-node ./cmd/quantum-node

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/quantum-node .

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/

# Create data directory
RUN mkdir -p /root/data

# Expose ports
EXPOSE 30303 8545 8546

# Create non-root user for security
RUN addgroup -g 1001 quantum && \
    adduser -D -s /bin/sh -u 1001 -G quantum quantum

# Change ownership
RUN chown -R quantum:quantum /root/

# Switch to non-root user
USER quantum

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8545/ || exit 1

# Set entrypoint
ENTRYPOINT ["./quantum-node"]

# Default command
CMD ["--config", "configs/default.json"]