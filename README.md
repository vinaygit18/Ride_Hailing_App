# GoComet Ride-Hailing Application

https://github.com/user-attachments/assets/36d500af-68d1-4b2e-a6a0-81f5148146ab


A production-grade ride-hailing backend system built with Go, featuring real-time driver matching, dynamic pricing, and WebSocket-based live updates.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24+ with Gin framework |
| Database | PostgreSQL 15+ |
| Cache | Redis 7+ (geo-spatial indexing) |
| Real-time | WebSocket (Gorilla) |
| Monitoring | New Relic APM |
| Logging | Uber Zap |

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make

## Quick Start

```bash
# 1. Start infrastructure (PostgreSQL & Redis)
make docker-up

# 2. Run database migrations
make migrate-up

# 3. Seed sample data (50 drivers, 20 riders)
make seed

# 4. Run the application
make run
```

**Access Points:**

| Interface | URL |
|-----------|-----|
| Rider UI | http://localhost:8080/rider |
| Driver UI | http://localhost:8080/driver |
| Health Check | http://localhost:8080/health |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/rides` | Create ride request |
| GET | `/v1/rides/:id` | Get ride details |
| GET | `/v1/drivers/all` | List all drivers |
| GET | `/v1/drivers/random` | Get random driver |
| POST | `/v1/drivers/:id/location` | Update driver location |
| POST | `/v1/drivers/:id/accept` | Accept ride |
| POST | `/v1/trips/:id/end` | End trip & calculate fare |
| POST | `/v1/payments` | Process payment |
| GET | `/v1/riders/random` | Get random rider |
| GET | `/v1/ws` | WebSocket connection |

## Project Structure

```
├── cmd/api/            # Application entry point
├── internal/
│   ├── api/            # HTTP handlers, routes, DTOs
│   ├── config/         # Configuration management
│   ├── domain/         # Business entities (driver, rider, ride, trip, payment)
│   ├── repository/     # Data access layer interfaces
│   └── service/        # Business logic (matching, pricing)
├── pkg/                # Shared packages
│   ├── cache/          # Redis client
│   ├── database/       # PostgreSQL connection
│   ├── logger/         # Zap logging
│   ├── monitoring/     # New Relic APM
│   └── websocket/      # WebSocket hub
├── migrations/         # SQL migrations
├── scripts/            # Utility scripts
└── web/                # Frontend (HTML, JS, CSS)
```

## Configuration

Copy `.env.example` to `.env`:

```env
# Server
SERVER_PORT=8080
SERVER_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gocomet
DB_USER=postgres
DB_PASSWORD=postgres

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
```

See `.env.example` for all available options including pricing, matching, and rate limiting settings.

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make docker-up` | Start PostgreSQL & Redis |
| `make docker-down` | Stop containers |
| `make docker-clean` | Stop & remove volumes |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback migrations |
| `make seed` | Seed sample data |
| `make run` | Start application |
| `make build` | Build binary |
| `make test-unit` | Run unit tests |
| `make test-coverage` | Generate coverage report |
| `make setup` | Complete setup (docker + migrate + deps) |
| `make dev` | Run with hot reload |

## Documentation

| Document | Description |
|----------|-------------|
| [HLD.md](HLD.md) | System architecture and design decisions |
| [QUICK_TEST_GUIDE.md](QUICK_TEST_GUIDE.md) | Step-by-step testing instructions |
