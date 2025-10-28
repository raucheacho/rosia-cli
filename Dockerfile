# Multi-stage Dockerfile for Rosia CLI
# This creates a minimal container image for running Rosia in containerized environments

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/raucheacho/rosia-cli/cmd.version=${VERSION:-dev} -X github.com/raucheacho/rosia-cli/cmd.commit=${COMMIT:-none} -X github.com/raucheacho/rosia-cli/cmd.date=${DATE:-unknown}" \
    -o rosia \
    ./main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 rosia && \
    adduser -D -u 1000 -G rosia rosia

# Set working directory
WORKDIR /home/rosia

# Copy binary from builder
COPY --from=builder /build/rosia /usr/local/bin/rosia

# Copy default profiles
COPY --chown=rosia:rosia profiles /home/rosia/.rosia/profiles

# Create necessary directories
RUN mkdir -p /home/rosia/.rosia/trash && \
    mkdir -p /home/rosia/.rosia/plugins && \
    chown -R rosia:rosia /home/rosia/.rosia

# Switch to non-root user
USER rosia

# Set default config
RUN echo '{"trash_retention_days":3,"profiles":["node","python","rust","flutter","go"],"ignore_paths":[],"plugins":[],"concurrency":0,"telemetry_enabled":false}' > /home/rosia/.rosiarc.json

# Default command
ENTRYPOINT ["/usr/local/bin/rosia"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="Rosia CLI"
LABEL org.opencontainers.image.description="Clean development dependencies and caches across multiple project types"
LABEL org.opencontainers.image.url="https://github.com/raucheacho/rosia-cli"
LABEL org.opencontainers.image.source="https://github.com/raucheacho/rosia-cli"
LABEL org.opencontainers.image.vendor="raucheacho"
LABEL org.opencontainers.image.licenses="MIT"
