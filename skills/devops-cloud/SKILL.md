# DevOps & Cloud

**ID:** devops-cloud
**Version:** 2.0
**Category:** Infrastructure & Deployment
**Triggers:** Docker, CI/CD, deployment, infrastructure, monitoring, logs, production, Kubernetes

---

## Role

I am a senior DevOps engineer. I manage infrastructure, CI/CD pipelines, containerization, and monitoring for the QW Pay platform.

---

## Infrastructure

### Docker Compose

```yaml
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: qw_pay
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/qw_pay?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      REDIS_ADDR: redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy

  antifraud-cpp:
    build:
      context: ./antifraud/cpp
      dockerfile: Dockerfile
    depends_on:
      - redis

  antifraud-python:
    build:
      context: ./antifraud
      dockerfile: Dockerfile.python
    depends_on:
      - redis

volumes:
  pgdata:
```

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
```

---

## CI/CD Pipeline

### GitHub Actions

```yaml
name: CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: qw_pay_test
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run tests
        run: go test -race -v ./...
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/qw_pay_test?sslmode=disable
          JWT_SECRET: test-secret-key
          REDIS_ADDR: localhost:6379

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        run: go build -o server ./cmd/server

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: server
          path: server

  release:
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      
      - name: Download artifact
        uses: actions/download-artifact@v4
        with:
          name: server

      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          tag_name: v1.0.${{ github.run_number }}
          files: server
```

---

## Make Commands

```makefile
.PHONY: db stop build run test lint antifraud clean help

# Infrastructure
db:
	docker-compose up -d

stop:
	docker-compose down

# Server
build:
	go build -o server ./cmd/server

run: build
	./server

test:
	go test -v ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run

# Anti-fraud
antifraud-build:
	cd antifraud/cpp && make

antifraud: antifraud-build
	python3 antifraud/orchestrator.py &

# Utilities
demo:
	open http://localhost:8080/demo

logs:
	tail -f logs/server.log

clean:
	rm -f server
	rm -f antifraud/cpp/fraud_engine

help:
	@echo "Available commands:"
	@echo "  make db              - Start PostgreSQL + Redis"
	@echo "  make stop            - Stop all containers"
	@echo "  make build           - Build server binary"
	@echo "  make run             - Build and run server"
	@echo "  make test            - Run tests"
	@echo "  make lint            - Run linter"
	@echo "  make antifraud       - Start anti-fraud engines"
	@echo "  make demo            - Open demo in browser"
	@echo "  make logs            - View logs"
	@echo "  make clean           - Remove binaries"
```

---

## Environment Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable` | PostgreSQL connection |
| `JWT_SECRET` | — | JWT signing secret (required) |
| `PORT` | `8080` | Server port |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis address |
| `MAX_TRANSFER_AMOUNT` | `10000000` | Max transfer limit |
| `DAILY_LIMIT` | `50000000` | Daily transfer limit |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:8080` | Allowed CORS origins |

---

## Monitoring

### Health Check Endpoint
```go
func healthHandler(c *gin.Context) {
    // Check PostgreSQL
    if err := db.Ping(c.Request.Context()); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "db": "disconnected"})
        return
    }

    // Check Redis
    if err := redis.Ping(c.Request.Context()).Err(); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "redis": "disconnected"})
        return
    }

    c.JSON(200, gin.H{"status": "healthy"})
}
```

### Metrics to Track
- Request latency (p50, p95, p99)
- Error rate (4xx, 5xx)
- Active connections
- Transaction throughput
- Anti-fraud verdicts

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `connection refused` to PostgreSQL | `docker-compose down && docker-compose up -d` |
| `connection refused` to Redis | `redis-server --daemonize yes` |
| C++ build error | `brew install hiredis` (macOS) or `apt install libhiredis-dev` (Linux) |
| Port 8080 in use | `lsof -i :8080` and kill process |
