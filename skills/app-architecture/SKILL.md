# Application Architecture

**ID:** app-architecture
**Version:** 2.0
**Category:** System Design
**Triggers:** architecture, patterns, design decisions, system design, scalability, microservices

---

## Role

I am a system architect. I design application architecture, define patterns, and ensure architectural consistency across the QW Pay platform.

---

## Architectural Patterns

### 1. Clean Architecture
```
┌─────────────────────────────────────────┐
│            Presentation Layer           │
│         (Handlers, Middleware)          │
├─────────────────────────────────────────┤
│            Business Logic Layer         │
│           (Services, Use Cases)         │
├─────────────────────────────────────────┤
│            Data Access Layer            │
│          (Repositories, Models)         │
├─────────────────────────────────────────┤
│            Infrastructure               │
│      (Database, Redis, External)        │
└─────────────────────────────────────────┘
```

**Dependency Rule:** Dependencies point inward. Inner layers never depend on outer layers.

### 2. Repository Pattern
```go
// Interface (domain level)
type AccountRepository interface {
    Create(ctx context.Context, account *Account) error
    FindByID(ctx context.Context, id string) (*Account, error)
    FindByUserID(ctx context.Context, userID string) ([]*Account, error)
    UpdateBalance(ctx context.Context, id string, amount decimal.Decimal, version int) error
}

// Implementation (infrastructure level)
type PostgresAccountRepository struct {
    db *pgxpool.Pool
}
```

### 3. CQRS (Command Query Responsibility Segregation)
```
Commands (Write):
  POST /accounts → CreateAccount
  POST /transactions → CreateTransfer

Queries (Read):
  GET /accounts → ListAccounts
  GET /transactions → ListTransactions
```

### 4. Event Sourcing for Anti-fraud
```
Transaction Created → Event → Redis Queue → Anti-fraud Check → Verdict Event
```

---

## System Architecture

### QW Pay Architecture
```
                    ┌──────────────────────┐
                    │      API Gateway      │
                    │   Gin + JWT + CORS    │
                    └──────────┬───────────┘
                               │
           ┌───────────────────┼───────────────────┐
           │                   │                   │
    ┌──────▼──────┐     ┌──────▼──────┐     ┌──────▼──────┐
    │  Auth       │     │  Account    │     │ Transaction │
    │  Service    │     │  Service    │     │  Service    │
    └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
           │                   │                   │
           └───────────────────┼───────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
  ┌──────▼──────┐      ┌──────▼──────┐      ┌───────▼──────┐
  │ PostgreSQL  │      │    Redis    │      │  Anti-Fraud  │
  │   (ACID)    │      │  (queues)   │      │  C++ + Python│
  └─────────────┘      └─────────────┘      └──────────────┘
```

### Data Flow: Transfer
```
1. Client → POST /api/v1/transactions
2. Handler validates JWT (middleware)
3. Handler validates request body
4. Handler checks account ownership
5. Service sends to Redis queue (anti-fraud)
6. C++ engine: velocity, limits, blacklists
7. Python engine: patterns, scoring
8. Verdict → Redis
9. Service receives verdict
10. If approved → ACID transaction:
    a. BEGIN
    b. UPDATE accounts (debit) WHERE version = X
    c. UPDATE accounts (credit) WHERE version = Y
    d. INSERT INTO transactions
    e. COMMIT
11. Client ← 200 OK / 403 Forbidden
```

---

## Design Patterns

### 1. Strategy Pattern (Anti-fraud)
```go
type FraudChecker interface {
    Check(ctx context.Context, tx *Transaction) (*Verdict, error)
}

type CPlusPlusChecker struct {
    client *RedisClient
}

type PythonChecker struct {
    client *RedisClient
}

func (c *CPlusPlusChecker) Check(ctx context.Context, tx *Transaction) (*Verdict, error) {
    // Fast checks: velocity, limits, blacklists
}

func (c *PythonChecker) Check(ctx context.Context, tx *Transaction) (*Verdict, error) {
    // Deep analysis: patterns, scoring
}
```

### 2. Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    failures    int
    threshold   int
    resetTimeout time.Duration
    state       State
    mu          sync.RWMutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.RLock()
    if cb.state == Open {
        cb.mu.RUnlock()
        return ErrCircuitOpen
    }
    cb.mu.RUnlock()

    err := fn()
    
    cb.mu.Lock()
    if err != nil {
        cb.failures++
        if cb.failures >= cb.threshold {
            cb.state = Open
            go cb.resetAfterTimeout()
        }
    } else {
        cb.failures = 0
        cb.state = Closed
    }
    cb.mu.Unlock()

    return err
}
```

### 3. Retry Pattern
```go
func Retry(attempts int, sleep time.Duration, fn func() error) error {
    if err := fn(); err != nil {
        if attempts--; attempts > 0 {
            time.Sleep(sleep)
            return Retry(attempts, sleep*2, fn) // Exponential backoff
        }
        return err
    }
    return nil
}

// Usage
err := Retry(3, 100*time.Millisecond, func() error {
    return service.Transfer(ctx, fromID, toID, amount)
})
```

---

## Scalability Considerations

### Horizontal Scaling
```
                    ┌─────────────┐
                    │   Load      │
                    │  Balancer   │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐      ┌────▼────┐      ┌────▼────┐
    │  App 1  │      │  App 2  │      │  App 3  │
    └────┬────┘      └────┬────┘      └────┬────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │
                    ┌──────▼──────┐
                    │  PostgreSQL │
                    │   (Primary) │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
         │ Replica │ │ Replica │ │ Replica │
         └─────────┘ └─────────┘ └─────────┘
```

### Database Sharding
```go
func getShard(userID string) string {
    hash := fnv.New32a()
    hash.Write([]byte(userID))
    shardID := hash.Sum32() % 4
    return fmt.Sprintf("shard_%d", shardID)
}
```

---

## Decision Log

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Performance, concurrency, simplicity |
| Database | PostgreSQL | ACID, reliability, JSON support |
| Cache | Redis | Speed, pub/sub, data structures |
| Anti-fraud | C++ + Python | Speed + ML capabilities |
| ORM | None (pgx) | Control, performance |
| API | REST | Simplicity, widely supported |
| Auth | JWT + Refresh | Stateless + revocable |

---

## Architectural Principles

1. **Single Responsibility** — Each service does one thing well
2. **Open/Closed** — Open for extension, closed for modification
3. **Dependency Inversion** — Depend on abstractions, not concretions
4. **Interface Segregation** — Small, focused interfaces
5. **Keep It Simple** — Prefer simple solutions over clever ones
