# syntax=docker/dockerfile:1

# Multi-stage production build for LatinaBot
# Stage 1: Build statically linked binary
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /build

# Install CA certificates and timezone database
RUN apk add --no-cache ca-certificates tzdata

# Leverage Docker cache for module dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy application source code
COPY . .

# Target OS and architecture provided automatically by Docker Buildx
ARG TARGETOS
ARG TARGETARCH

# Build statically linked binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /build/bin/latinabot \
    ./cmd/bot

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies (SSL certificates and timezone)
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/bin/latinabot /app/latinabot

# Ensure temp directory exists
RUN mkdir -p /app/temp

# Default container environment variables
ENV PORT=8080 \
    APP_ENV=production

EXPOSE 8080

ENTRYPOINT ["/app/latinabot"]