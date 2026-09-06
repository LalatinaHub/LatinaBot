# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata git

# Copy common library and bot source code
COPY common /common
COPY LatinaBot /src/LatinaBot

WORKDIR /src/LatinaBot

# Download dependencies and compile static binary
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/latinabot ./cmd/bot

# Runtime Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/latinabot /app/latinabot

# Create temporary directory for local order vouchers and QR codes
RUN mkdir -p /app/temp

EXPOSE 8080

ENTRYPOINT ["/app/latinabot"]