# Stage 1: Build binary & agent artifacts
FROM golang:alpine AS builder

# Install build dependencies (certificates, git if needed)
RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app

# Download dependencies first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile statically linked Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/bin/api ./cmd/api/main.go

# Compile agent binaries and prepare /app/downloads
RUN mkdir -p /app/downloads && \
    AGENT_SRC="" && \
    if [ -d "/app/agent" ] && [ -f "/app/agent/cmd/main.go" ]; then \
        AGENT_SRC="/app/agent"; \
    elif [ -d "/app/../agent" ] && [ -f "/app/../agent/cmd/main.go" ]; then \
        AGENT_SRC="/app/../agent"; \
    fi && \
    if [ -n "$AGENT_SRC" ]; then \
        echo "Building agent binaries from $AGENT_SRC..." && \
        cd "$AGENT_SRC" && \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/downloads/encertia-agent-linux ./cmd && \
        CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o /app/downloads/encertia-agent-darwin ./cmd && \
        CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o /app/downloads/encertia-agent-windows.exe ./cmd && \
        if [ -f "$AGENT_SRC/certs/encertia-codesign.cer" ]; then \
            cp "$AGENT_SRC/certs/encertia-codesign.cer" /app/downloads/encertia-codesign.cer; \
        elif [ -f "$AGENT_SRC/certs/encertia-codesign.crt" ]; then \
            grep -v -- '-----' "$AGENT_SRC/certs/encertia-codesign.crt" | tr -d '\r\n' | base64 -d > /app/downloads/encertia-codesign.cer; \
        fi && \
        cd /app; \
    fi && \
    echo "Files in /app/downloads:" && \
    ls -la /app/downloads || true

# Stage 2: Minimal runtime container
FROM alpine:3.20 AS runner

# Install essential runtime utilities for healthchecks and SSL/TLS certificates
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -g 10001 -S appgroup \
    && adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/bin/api /usr/local/bin/api

# Copy downloads directory
COPY --from=builder --chown=appuser:appgroup /app/downloads /app/downloads

# Switch to unprivileged user
USER appuser:appgroup

# Environment defaults
ENV PORT=8080 \
    GIN_MODE=release \
    AGENT_DOWNLOADS_DIR=/app/downloads

# Expose HTTP port
EXPOSE 8080

# Health check using the /health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
