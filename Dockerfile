# Build stage
FROM golang:1.20-slim AS builder

# Set up proper working directory
WORKDIR /build

# Install necessary build tools
RUN apk add --no-cache git make ca-certificates tzdata

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application (static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o aggregator ./cmd/aggregator

# Runtime stage
FROM slim:3.18

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /build/aggregator /app/aggregator

# Create directories for configs and output
RUN mkdir -p /app/configs /app/output

# Copy default configuration files
COPY --from=builder /build/configs/ /app/configs/

# Set environment variables
ENV CONFIG_PATH=/app/configs/config.yaml
ENV SOURCES_PATH=/app/configs/sources.yaml
ENV OUTPUT_DIR=/app/output

# Expose web & API ports
EXPOSE 8080

# Create a non-root user for security
RUN adduser -D -H -h /app appuser
RUN chown -R appuser:appuser /app
USER appuser

# Set entrypoint
ENTRYPOINT ["/app/aggregator"]

# Default command if none provided
CMD ["--config", "/app/configs/config.yaml", "--sources", "/app/configs/sources.yaml"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O- http://localhost:8080/health || exit 1

# Labels for metadata
LABEL maintainer="CyberwizD <https://github.com/CyberwizD>"
LABEL version="1.0.0"
LABEL description="Concurrent Web Content Aggregator"