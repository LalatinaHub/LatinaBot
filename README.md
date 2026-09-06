# LatinaBot

Official Telegram Bot service for the [LalatinaHub](https://github.com/LalatinaHub) VPN ecosystem, rewritten natively in **Go 1.24+**.

## Features

- **Telegram Bot Engine:** Built with `telebot.v3`, supporting custom inline keyboards, error reporting, rate limiting, and private chat authorization.
- **VPN Configuration Wizard:** Multi-step interactive flow for VMess, VLESS, and Trojan protocol setup, edge server selection with live latency/ping stats, and relay country pagination.
- **Donation Receipt OCR Verification:** Automatic verification of Trakteer & local receipts using Google Cloud Vision OCR (`DetectTexts`).
- **Payment Gateway Integration:** Dynamic QR code generation & verification via Xendit API.
- **Cloudflare Integration:** Automated Worker subdomains registration and DNS records management.
- **Ecosystem Integration:** Leverages [`github.com/LalatinaHub/common`](../common) for Turso LibSQL pooling, models (`User`, `Server`, `ProxyNode`, `KeyValue`), and proxy protocol serializations.
- **Background Cron Scheduler:** Automatic cleanup of expired accounts, quota enforcement, tenant load balancing, and scheduled free public proxy broadcasts.
- **Embedded HTTP Service:** Lightweight `/api/v1/proxy/check` endpoint with raw TLS socket validation.

## Prerequisites

- Go 1.24+
- Access to Turso LibSQL database (`TURSO_DATABASE_URL`, `TURSO_AUTH_TOKEN`)
- Telegram Bot Token (`BOT_TOKEN`)

## Quick Start

1. Copy `.env.example` to `.env` and fill in credentials:
   ```bash
   cp .env.example .env
   ```

2. Run the test suite:
   ```bash
   go test -v ./...
   ```

3. Build and run locally:
   ```bash
   go run ./cmd/bot
   ```

4. Build binary:
   ```bash
   go build -o latinabot ./cmd/bot
   ```

## Docker Build

```bash
docker build -t latinabot:latest -f Dockerfile ..
```
