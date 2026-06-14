# Development Guide

## Requirements

- Go 1.22+
- PostgreSQL 16+ (or Docker)
- Redis 7+ (for anti-fraud)
- g++ and libhiredis-dev (for C++ engine)
- Python 3.12+ (for Python engine)

---

## Installation

### Ubuntu / Debian

```bash
# Go
sudo snap install go --classic

# PostgreSQL + Redis
sudo apt install postgresql redis-server

# C++ dependencies
sudo apt install g++ libhiredis-dev

# Python
sudo apt install python3 python3-pip
pip3 install redis
```

### macOS

```bash
brew install go postgresql redis python3 hiredis
```

### Arch Linux

```bash
sudo pacman -S go postgresql redis hiredis python python-pip
pip install redis
```

---

## Setup

```bash
# 1. Clone and configure
git clone <repo-url> && cd qw_pay
cp .env.example .env

# 2. Edit .env if needed
# DATABASE_URL, JWT_SECRET, PORT, REDIS_ADDR

# 3. Start PostgreSQL + Redis
make db

# 4. Run migrations (auto via Docker, manual otherwise)
psql $DATABASE_URL < migrations/001_init.sql
psql $DATABASE_URL < migrations/002_exchange_rates.sql
psql $DATABASE_URL < migrations/003_audit_fraud_tables.sql
```

---

## Running

### Server Only

```bash
make run
```

### Server + Anti-Fraud

```bash
# Terminal 1: Server
make run

# Terminal 2: Anti-fraud engines
make antifraud
```

### Full Docker Stack

```bash
cd deploy
docker-compose up -d
```

---

## Project Structure

```
internal/
├── auth/              # Authentication (register, OTP, JWT)
│   ├── handler.go     # HTTP handlers
│   ├── service.go     # Business logic
│   └── repository.go  # Database access
├── account/           # Accounts (CRUD, block, bonus)
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── transaction/       # Transfers (ACID, idempotent, retry)
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── currency/          # Exchange rates
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── audit/             # Audit log + admin API
│   ├── handler.go
│   ├── service.go
│   └── repository.go
├── antifraud/         # Anti-fraud Redis client
│   └── client.go
├── config/            # Environment config
├── database/          # pgxpool connection
├── middleware/        # JWT + AdminRequired
└── model/             # Data models + enums
```

---

## Adding a New Endpoint

### 1. Create Repository (if DB access needed)

```go
// internal/myfeature/repository.go
package myfeature

type Repository struct {
    db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
    return &Repository{db: db}
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*model.Thing, error) {
    // SQL query
}
```

### 2. Create Service

```go
// internal/myfeature/service.go
package myfeature

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) DoSomething(ctx context.Context, id uuid.UUID) error {
    // Business logic
}
```

### 3. Create Handler

```go
// internal/myfeature/handler.go
package myfeature

type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) HandleRequest(c *gin.Context) {
    // Validation, service call, response
}
```

### 4. Register in main.go

```go
myRepo := myfeature.NewRepository(database.Pool)
mySvc := myfeature.NewService(myRepo)
myH := myfeature.NewHandler(mySvc)

v1.POST("/my-endpoint", myH.HandleRequest)
```

---

## Testing

### Run Tests

```bash
make test
```

### Manual Testing with curl

```bash
# Register
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","phone":"+79001234567","password":"secret123"}'

# Get OTP from logs/server.log, then verify
curl -X POST http://localhost:8080/api/v1/verify \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","otp_code":"123456"}'

# Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"secret123"}'

# Create account
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"currency":"USD"}'

# Transfer
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id":"<from-uuid>",
    "to_account_id":"<to-uuid>",
    "amount":10,
    "idempotency_key":"test-key-1"
  }'

# View audit logs (requires ADMIN role)
curl http://localhost:8080/api/v1/admin/audit-logs \
  -H "Authorization: Bearer <admin-token>"
```

---

## Anti-Fraud

### C++ Engine

```bash
make antifraud-build
./antifraud/cpp/fraud_engine
```

### Python Engine

```bash
python3 antifraud/python/service.py
```

### Orchestrator (both engines)

```bash
make antifraud
```

---

## Troubleshooting

### "connection refused" to PostgreSQL

```bash
docker-compose ps
make stop && make db
```

### "connection refused" to Redis

```bash
redis-server --daemonize yes
# or
docker run -d -p 6379:6379 redis:7-alpine
```

### C++ build error

```bash
# macOS:
brew install hiredis

# Ubuntu:
sudo apt install g++ libhiredis-dev
```

### Migration already applied

The Docker setup auto-runs migrations via `/docker-entrypoint-initdb.d`. For manual setup, run each migration file once.

---

## Make Commands

```bash
make help              # Show all commands
make db                # Start PostgreSQL + Redis
make build             # Build Go server
make run               # Build and run server
make lint              # Run golangci-lint
make test              # Run tests
make antifraud-build   # Build C++ engine
make antifraud         # Start C++ + Python engines
make demo              # Open demo page
make logs              # Tail server logs
make clean             # Remove binaries
```
