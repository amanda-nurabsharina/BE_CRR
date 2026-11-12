# Multi-stage build for BE_CRR (Go API)
FROM golang:alpine AS builder

# Install build dependencies & git
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Enable automatic toolchain download if go.mod requires newer toolchain
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Invalidate cache on new commits
ARG GIT_SHA

# Copy application source
COPY . .

# Build statically compiled binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server cmd/api/main.go

# Minimal production image
FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata

# Set timezone to Asia/Jakarta (WIB)
ENV TZ=Asia/Jakarta
RUN cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && echo "Asia/Jakarta" > /etc/timezone

WORKDIR /app

# Create directory for audio & chunks storage
RUN mkdir -p /app/storage_data

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Expose Fiber API port
EXPOSE 8080

# Production default envs
ENV PORT=8080 \
    APP_ENV=production \
    STORAGE_DIR=/app/storage_data

# Persistent volume for recording files in Easypanel
VOLUME ["/app/storage_data"]

ENTRYPOINT ["/app/server"]

