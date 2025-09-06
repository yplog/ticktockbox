# Multi-stage build for TickTockBox (Go)

# 1) Builder image: compiles the server binary
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Allow multi-arch builds via buildx
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Enable module mode and cache deps early
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the rest of the source and build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w" -o /out/ticktockbox ./cmd/server


# 2) Runtime image: small, non-root, with writable /data for SQLite
FROM alpine:3.20 AS runtime

# Create non-root user and writable directories
RUN adduser -D -u 10001 appuser \
    && mkdir -p /app /data \
    && chown -R appuser:appuser /app /data

WORKDIR /app

# Copy binary from builder
COPY --from=builder /out/ticktockbox /app/ticktockbox

# Default environment (can be overridden at runtime)
ENV ADDR=:3000 \
    SQLITE_PATH=/data/app.db \
    RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/ \
    RABBITMQ_QUEUE=reminders.due

EXPOSE 3000
VOLUME ["/data"]
USER appuser

ENTRYPOINT ["/app/ticktockbox"]
