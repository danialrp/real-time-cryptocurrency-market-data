# 📡 Real-Time Cryptocurrency Market Data Platform

> A high-throughput, event-driven microservices platform for real-time cryptocurrency market data ingestion, aggregation, and fan-out distribution — built in Go.

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat-square&logo=go&logoColor=white)
![NATS](https://img.shields.io/badge/NATS-JetStream-27AAE1?style=flat-square&logo=natsdotio&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7.x-DC382D?style=flat-square&logo=redis&logoColor=white)
![TimescaleDB](https://img.shields.io/badge/TimescaleDB-PostgreSQL-FDB515?style=flat-square&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

---

## Overview

This project is a production-grade, event-driven backend platform designed to ingest live cryptocurrency market streams, aggregate and persist time-series data, and fan out to WebSocket clients and REST consumers at scale.

Built as a learning ground and architectural reference for **Go microservices**, **NATS JetStream messaging**, and **real-time data pipelines** — every design decision prioritises throughput, resilience, and operational clarity.

---

## Architecture

The system is structured as a **4-service pipeline**:

```
Exchange WebSocket Feed
        │
        ▼
  ┌─────────────┐
  │   Producer  │  — Connects to exchange streams, publishes to NATS JetStream
  └──────┬──────┘
         │  NATS JetStream (8 subjects)
         ├──────────────────────────────────────┐
         ▼                                      ▼
  ┌─────────────┐                      ┌────────────────┐
  │  Aggregator │                      │  WebSocket Hub │  — Real-time fan-out to clients
  │  + Persister│                      └────────────────┘
  └──────┬──────┘
         │
         ▼
  ┌─────────────┐
  │  REST API   │  — Fiber v2, market detail + USDT/FIAT rate endpoints
  └─────────────┘
```

### Services

| Service | Responsibility |
|---|---|
| **Producer** | Connects to exchange WebSocket feeds; publishes raw market events to NATS JetStream |
| **Aggregator** | Consumes JetStream subjects; aggregates and persists time-series data to TimescaleDB |
| **WebSocket Hub** | Thread-safe pub/sub hub; broadcasts live market events to connected clients |
| **REST API** | Fiber v2 HTTP server; serves aggregated market data with USDT/FIAT rate conversion |

A **shared Go module** defines internal contracts (event types, interfaces) used across all services.

---

## Key Design Decisions

### Dual Connection Strategy

The producer implements a smart connection management layer:

- **Always-on** subscriptions for high-frequency streams (`ticker`, `detail`) — persistent connections maintained at all times
- **Lazy / demand-driven** connections for low-frequency streams (`kline`, `BBO`, `depth`, `trade`) — connections open on first subscriber and self-terminate when subscriber count drops to zero, minimising upstream socket pressure

### NATS JetStream Topology

8 dedicated subjects with per-subject retention policies:

- **TTL:** 10-minute message retention
- **Cap:** 50-message limit per subject
- Enables backpressure-safe fan-out to heterogeneous consumers without data loss

### Redis Layout (4 isolated databases)

| DB | Purpose |
|---|---|
| DB 0 | Response caching |
| DB 1 | Aggregation state |
| DB 2 | Rate limiting |
| DB 3 | Lazy connection tracking (ephemeral) |

Explicit DB isolation prevents cross-concern cache pollution and makes operational debugging straightforward.

### TimescaleDB via Ent ORM

Market OHLCV and tick data is persisted to **TimescaleDB** (PostgreSQL extension for time-series) using the **Ent ORM** with schema-driven migrations. Hypertable partitioning is applied automatically for query performance at scale.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.21 |
| Messaging | NATS JetStream |
| Cache / State | Redis 7.x (4 isolated DBs) |
| Persistence | TimescaleDB (PostgreSQL) |
| ORM | Ent ORM |
| HTTP Framework | Fiber v2 |
| WebSocket | Gorilla WebSocket |
| Containerisation | Docker Compose |

---

## Project Structure

```
.
├── producer/           # Exchange WebSocket feed consumer + NATS publisher
│   └── streams/        # Dual-mode connection handlers per stream type
├── aggregator/         # JetStream consumer + TimescaleDB writer
├── ws-hub/             # Thread-safe WebSocket pub/sub hub
├── api/                # Fiber v2 REST API server
├── shared/             # Internal contracts: event types, interfaces, config
├── docker-compose.yml  # Full stack: NATS, Redis, TimescaleDB, all services
└── README.md
```

---

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) & Docker Compose
- Go 1.21+ (for local development without Docker)

### Run with Docker Compose

```bash
# Clone the repository
git clone https://github.com/danialrp/real-time-cryptocurrency-market-data.git
cd real-time-cryptocurrency-market-data

# Copy environment config
cp .env.example .env

# Start the full stack
docker compose up -d
```

All services, NATS JetStream, Redis, and TimescaleDB will spin up together.

### Environment Variables

| Variable | Description |
|---|---|
| `EXCHANGE_WS_URL` | WebSocket endpoint of the exchange feed |
| `NATS_URL` | NATS server connection URL |
| `REDIS_ADDR` | Redis host and port |
| `DATABASE_URL` | TimescaleDB / PostgreSQL DSN |
| `API_PORT` | HTTP port for the REST API (default: `8080`) |

---

## REST API

Base URL: `http://localhost:8080`

| Endpoint | Method | Description |
|---|---|---|
| `/api/markets` | GET | List all available market pairs |
| `/api/markets/:symbol` | GET | Market detail for a symbol (with USDT/FIAT rate conversion) |
| `/api/markets/:symbol/klines` | GET | Historical OHLCV kline data |
| `/health` | GET | Service health check |

### WebSocket

Connect to `ws://localhost:8080/ws` to receive live market event streams.

Supported stream subscriptions: `ticker`, `detail`, `kline`, `BBO`, `depth`, `trade`

---

## Concepts Explored

This project was built to gain hands-on production experience with:

- **Go concurrency patterns** — goroutines, channels, sync primitives, thread-safe data structures
- **Event-driven microservices** — NATS JetStream subjects, consumer groups, retention policies
- **Time-series databases** — TimescaleDB hypertables, efficient OHLCV ingestion
- **Real-time WebSocket fan-out** — building a scalable pub/sub hub from scratch in Go
- **Operational Redis patterns** — DB isolation, TTL-based eviction, rate limiting
- **Ent ORM** — schema-first Go ORM with code generation and migration support

---

## License

MIT — see [LICENSE](LICENSE) for details.

---

> Built by [Danial Panah](https://danialrp.com) · [GitHub](https://github.com/danialrp) · [LinkedIn](https://linkedin.com/in/danialrp)
