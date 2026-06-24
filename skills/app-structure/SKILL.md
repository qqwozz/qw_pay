# Application Structure

**ID:** app-structure
**Version:** 2.0
**Category:** Project Navigation
**Triggers:** file location, project structure, navigation, where is, directory, module

---

## Role

I help navigate the QW Pay codebase. I know exactly where every file and module is located.

---

## Complete Project Structure

```
qw_pay/
│
├── cmd/
│   └── server/
│       └── main.go                    # Entry point, DI, graceful shutdown
│
├── internal/
│   ├── account/                        # Account management
│   │   ├── handler.go                 # HTTP handlers
│   │   ├── service.go                 # Business logic
│   │   ├── repository.go              # Database queries
│   │   └── account_test.go            # Tests
│   │
│   ├── auth/                           # Authentication
│   │   ├── handler.go                 # HTTP handlers
│   │   ├── service.go                 # JWT, OTP, password hashing
│   │   ├── repository.go              # User persistence
│   │   └── auth_test.go               # Tests
│   │
│   ├── transaction/                    # Transfer logic
│   │   ├── handler.go                 # HTTP handlers
│   │   ├── service.go                 # ACID transfers, idempotency
│   │   ├── repository.go              # Transaction persistence
│   │   └── transaction_test.go        # Tests
│   │
│   ├── antifraud/                      # Anti-fraud client
│   │   └── client.go                  # Redis queue integration
│   │
│   ├── config/                         # Configuration
│   │   └── config.go                  # env parsing, validation
│   │
│   ├── contextkeys/                    # Context utilities
│   │   └── keys.go                    # Context key constants
│   │
│   ├── database/                       # Database connection
│   │   └── postgres.go                # pgxpool setup, health check
│   │
│   ├── errors/                         # Error types
│   │   └── errors.go                  # AppError, NotFound, BadRequest
│   │
│   ├── logger/                         # Logging
│   │   └── logger.go                  # slog setup
│   │
│   ├── middleware/                      # HTTP middleware
│   │   ├── auth.go                    # JWT verification
│   │   ├── cors.go                    # CORS headers
│   │   └── requestid.go               # Request ID injection
│   │
│   ├── model/                          # Data models
│   │   ├── user.go                    # User struct
│   │   ├── account.go                 # Account struct
│   │   ├── transaction.go             # Transaction struct
│   │   └── refresh_token.go           # RefreshToken struct
│   │
│   ├── ratelimit/                      # Rate limiting
│   │   └── ratelimit.go               # Token bucket implementation
│   │
│   └── response/                       # Response helpers
│       └── response.go                # OK, Created, Error, Paginated
│
├── antifraud/                          # Anti-fraud engines
│   ├── cpp/                            # C++ engine
│   │   ├── fraud_engine.cpp           # Main logic
│   │   ├── Makefile                   # Build configuration
│   │   └── Dockerfile                 # Container build
│   │
│   ├── orchestrator.py                 # Python orchestrator
│   └── Dockerfile.python              # Python container
│
├── migrations/                         # SQL migrations
│   ├── 001_init.sql                   # Users, accounts, transactions
│   └── 002_refresh_tokens.sql         # Refresh tokens table
│
├── web/                                # Frontend
│   └── index.html                     # Demo interface
│
├── docs/                               # Documentation
│   ├── API.md                         # API reference
│   ├── ARCHITECTURE.md                # System architecture
│   ├── DEVELOPMENT.md                 # Developer guide
│   └── TECHNICAL_SPECIFICATION.md     # Technical spec
│
├── skills/                             # AI skills
│   ├── security-engineer/
│   ├── api-platform/
│   ├── qa-test-automation/
│   ├── devops-cloud/
│   ├── data-analytics/
│   ├── frontend-architect/
│   ├── backend-architect/
│   ├── app-architecture/
│   ├── app-structure/
│   ├── computer-science/
│   └── ai-coordinator/
│
├── .github/
│   └── workflows/
│       └── ci.yml                     # GitHub Actions CI/CD
│
├── Makefile                           # Build commands
├── Dockerfile                         # App container
├── docker-compose.yml                 # Local infrastructure
├── go.mod                            # Go modules
├── go.sum                            # Dependency checksums
├── .golangci.yml                     # Linter config
├── .env.example                      # Environment template
├── .gitignore                        # Git ignore rules
├── AGENTS.md                         # AI instructions
├── README.md                         # Project readme
└── LICENSE                           # MIT License
```

---

## Quick Reference

### What am I looking for?

| Need | Location |
|------|----------|
| Entry point | `cmd/server/main.go` |
| HTTP routing | `cmd/server/main.go` (line ~100) |
| Auth logic | `internal/auth/service.go` |
| JWT generation | `internal/auth/service.go:GenerateToken()` |
| Password hashing | `internal/auth/service.go:HashPassword()` |
| OTP generation | `internal/auth/service.go:GenerateOTP()` |
| Account creation | `internal/account/service.go:CreateAccount()` |
| Balance update | `internal/account/repository.go:UpdateBalance()` |
| Transfer logic | `internal/transaction/service.go:Transfer()` |
| ACID transaction | `internal/transaction/service.go:Transfer()` (lines 50-100) |
| Optimistic locking | `internal/account/repository.go:UpdateBalance()` |
| Anti-fraud check | `internal/antifraud/client.go:CheckTransaction()` |
| Redis queue | `internal/antifraud/client.go` |
| C++ engine | `antifraud/cpp/fraud_engine.cpp` |
| Python engine | `antifraud/orchestrator.py` |
| SQL migrations | `migrations/` |
| Docker config | `docker-compose.yml` |
| CI/CD pipeline | `.github/workflows/ci.yml` |
| API documentation | `docs/API.md` |
| Frontend demo | `web/index.html` |

---

## Module Dependencies

```
cmd/server/main.go
    │
    ├── internal/config
    │   └── (no dependencies)
    │
    ├── internal/database
    │   └── internal/config
    │
    ├── internal/logger
    │   └── (stdlib only)
    │
    ├── internal/errors
    │   └── (stdlib only)
    │
    ├── internal/model
    │   └── shopspring/decimal
    │
    ├── internal/response
    │   └── internal/errors
    │
    ├── internal/middleware
    │   ├── internal/config
    │   └── internal/errors
    │
    ├── internal/auth
    │   ├── internal/model
    │   ├── internal/errors
    │   ├── internal/config
    │   └── internal/database
    │
    ├── internal/account
    │   ├── internal/model
    │   ├── internal/errors
    │   └── internal/database
    │
    └── internal/transaction
        ├── internal/model
        ├── internal/errors
        ├── internal/database
        └── internal/antifraud
```

---

## File Sizes (Approximate)

| File | Lines | Complexity |
|------|-------|------------|
| `cmd/server/main.go` | 150 | Medium |
| `internal/auth/service.go` | 200 | High |
| `internal/account/service.go` | 150 | Medium |
| `internal/transaction/service.go` | 250 | High |
| `internal/errors/errors.go` | 100 | Low |
| `internal/response/response.go` | 80 | Low |
| `internal/middleware/auth.go` | 60 | Low |

---

## Common Tasks

### Adding a new endpoint
1. Create `internal/myfeature/handler.go`
2. Create `internal/myfeature/service.go`
3. Create `internal/myfeature/repository.go` (if needed)
4. Register in `cmd/server/main.go`
5. Add tests in `internal/myfeature/*_test.go`
6. Document in `docs/API.md`

### Adding a new model
1. Add struct to `internal/model/`
2. Add repository methods
3. Add migration in `migrations/`
4. Update service layer

### Modifying anti-fraud
1. C++ changes: `antifraud/cpp/`
2. Python changes: `antifraud/orchestrator.py`
3. Client changes: `internal/antifraud/client.go`
