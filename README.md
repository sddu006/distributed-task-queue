# Distributed Task Queue System

A production-inspired distributed task queue built in Go — similar to Celery or Redis Queue. Features a REST API, concurrent worker pool, AOF-based persistence, and a real-time web dashboard.

## Architecture
Client (HTTP)

│

▼

REST API (Gin)          ← Accepts jobs, enforces rate limiting

│

▼

Job Queue (Channel)     ← Thread-safe queue using Go channels

│

▼

Worker Pool             ← 5 concurrent goroutines processing jobs

│

▼

Job Store (In-Memory)   ← Mutex-protected map

│

▼

AOF Log (Disk)          ← Append-only file for persistence

## Features

- **REST API** — Submit and track jobs via HTTP endpoints
- **Concurrent Worker Pool** — 5 workers process jobs simultaneously using goroutines
- **AOF Persistence** — Append-only file logging inspired by Redis; jobs survive server restarts
- **Rate Limiting** — Token bucket algorithm limits clients to 5 requests per IP
- **Fault Tolerance** — Automatic job retry up to 3 times on failure
- **Graceful Shutdown** — In-progress jobs complete before server stops
- **Real-time Dashboard** — Live job monitoring in browser, auto-refreshes every 2 seconds

## Tech Stack

- **Language:** Go
- **Web Framework:** Gin
- **Concurrency:** Goroutines, Channels, Mutex
- **Persistence:** Custom AOF (Append Only File)
- **Dashboard:** HTML, CSS, JavaScript

## Project Structure
distributed-task-queue/

├── main.go              # Entry point, wires everything together

├── api/

│   ├── server.go        # REST API handlers

│   └── ratelimiter.go   # Token bucket rate limiting

├── queue/

│   └── queue.go         # Thread-safe job queue

├── worker/

│   └── worker.go        # Worker pool with retry logic

├── store/

│   ├── store.go         # In-memory job store

│   └── aof.go           # AOF persistence

└── dashboard/

└── index.html       # Real-time web dashboard

## Getting Started

### Prerequisites
- Go 1.21+

### Run locally

```bash
git clone https://github.com/YOUR_USERNAME/distributed-task-queue.git
cd distributed-task-queue
go mod download
go run main.go
```

Server starts on `http://localhost:8080`

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/jobs` | Submit a new job |
| GET | `/jobs` | List all jobs |
| GET | `/jobs/:id` | Get job by ID |
| GET | `/dashboard` | Web dashboard |

### Submit a job

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type": "send_email", "payload": "to: user@example.com"}'
```

### Response

```json
{
  "id": "job_1718123456789",
  "type": "send_email",
  "payload": "to: user@example.com",
  "status": "pending",
  "created_at": "2024-06-15T10:00:00Z",
  "updated_at": "2024-06-15T10:00:00Z",
  "retries": 0
}
```

### Supported Job Types

| Type | Simulated Duration |
|------|--------------------|
| `send_email` | 2 seconds |
| `resize_image` | 3 seconds |
| `send_notification` | 1 second |

## Key Concepts

**AOF Persistence** — Every write operation is appended to `aof.log`. On restart, the log is replayed line by line to restore exact state — the same mechanism Redis uses for durability.

**Token Bucket Rate Limiting** — Each IP gets a bucket of 5 tokens. Tokens refill at 1/second. Each request costs 1 token. Empty bucket = 429 Too Many Requests.

**Worker Pool** — 5 goroutines wait on a buffered channel. When a job arrives, one worker picks it up. Failed jobs are retried up to 3 times before being marked failed.

## License

MIT

## Author

**Siddardha**
B.Tech, Electronics and Communication Engineering
Indian Institute of Technology Kharagpur

- GitHub: [@sddu006](https://github.com/sddu006)