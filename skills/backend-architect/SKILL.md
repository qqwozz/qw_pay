# Backend Architect

**ID:** backend-architect
**Version:** 2.0
**Category:** Backend & Go Architecture
**Triggers:** Go, Gin, backend, services, repository pattern, dependency injection, error handling, concurrency

---

## Role

I am a senior backend architect specializing in Go. I design scalable, maintainable backend systems following clean architecture principles.

---

## Tech Stack

- **Go 1.23+** — Primary language
- **Gin** — HTTP framework
- **pgx v5** — PostgreSQL driver
- **Redis 7** — Caching & queues
- **shopspring/decimal** — Precise arithmetic

---

## Architecture: Handler → Service → Repository

### Handler Layer
```go
type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) CreateAccount(c *gin.Context) {
    var req CreateAccountRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, apperr.BadRequest(err.Error()))
        return
    }

    userID := c.GetString(contextkeys.UserIDKey)
    
    account, err := h.svc.CreateAccount(c.Request.Context(), userID, req.Currency)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Created(c, account)
}
```

### Service Layer
```go
type Service struct {
    repo   Repository
    logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
    return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateAccount(ctx context.Context, userID, currency string) (*model.Account, error) {
    // Business logic
    if !validCurrency(currency) {
        return nil, apperr.BadRequest("invalid currency")
    }

    account := &model.Account{
        ID:       uuid.New(),
        UserID:   userID,
        Currency: currency,
        Balance:  decimal.NewFromInt(100), // Welcome bonus
        Version:  1,
        Status:   "ACTIVE",
    }

    if err := s.repo.Create(ctx, account); err != nil {
        return nil, apperr.Wrap(err, "create account")
    }

    s.logger.Info("account created", "user_id", userID, "account_id", account.ID)
    return account, nil
}
```

### Repository Layer
```go
type Repository interface {
    Create(ctx context.Context, account *model.Account) error
    FindByID(ctx context.Context, id string) (*model.Account, error)
    UpdateBalance(ctx context.Context, id string, amount decimal.Decimal, version int) error
}

type PostgresRepository struct {
    db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
    return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, account *model.Account) error {
    _, err := r.db.Exec(ctx,
        `INSERT INTO accounts (id, user_id, currency, balance, version, status, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
        account.ID, account.UserID, account.Currency, account.Balance, account.Version, account.Status,
    )
    return err
}

func (r *PostgresRepository) UpdateBalance(ctx context.Context, id string, amount decimal.Decimal, version int) error {
    result, err := r.db.Exec(ctx,
        `UPDATE accounts 
         SET balance = balance + $1, version = version + 1, updated_at = NOW()
         WHERE id = $2 AND version = $3 AND status = 'ACTIVE'`,
        amount, id, version,
    )
    if err != nil {
        return err
    }
    if rows, _ := result.RowsAffected(); rows == 0 {
        return ErrOptimisticLock
    }
    return nil
}
```

---

## Error Handling

### Error Types
```go
// internal/errors/errors.go
var (
    ErrNotFound       = errors.New("resource not found")
    ErrOptimisticLock = errors.New("optimistic lock conflict")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrForbidden      = errors.New("forbidden")
)

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

func NotFound(msg string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: msg}
}

func BadRequest(msg string) *AppError {
    return &AppError{Code: "BAD_REQUEST", Message: msg}
}

func Wrap(err error, msg string) *AppError {
    return &AppError{
        Code:    "INTERNAL_ERROR",
        Message: fmt.Sprintf("%s: %v", msg, err),
        Err:     err,
    }
}
```

### Error Response
```go
// internal/response/response.go
func Error(c *gin.Context, err error) {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        status := httpStatusFromCode(appErr.Code)
        c.JSON(status, gin.H{
            "error": gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
            },
        })
        return
    }

    c.JSON(http.StatusInternalServerError, gin.H{
        "error": gin.H{
            "code":    "INTERNAL_ERROR",
            "message": "Internal server error",
        },
    })
}

func httpStatusFromCode(code string) int {
    switch code {
    case "NOT_FOUND":
        return http.StatusNotFound
    case "BAD_REQUEST":
        return http.StatusBadRequest
    case "UNAUTHORIZED":
        return http.StatusUnauthorized
    case "FORBIDDEN":
        return http.StatusForbidden
    case "CONFLICT":
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}
```

---

## Dependency Injection

```go
// cmd/server/main.go
func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("failed to load config:", err)
    }

    // Database
    db, err := database.NewPool(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("failed to connect to database:", err)
    }
    defer db.Close()

    // Redis
    rdb := redis.NewClient(&redis.Options{
        Addr: cfg.RedisAddr,
    })

    // Logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Repositories
    userRepo := user.NewPostgresRepository(db)
    accountRepo := account.NewPostgresRepository(db)
    transactionRepo := transaction.NewPostgresRepository(db)

    // Services
    authService := auth.NewService(userRepo, cfg.JWTSecret, logger)
    accountService := account.NewService(accountRepo, logger)
    transactionService := transaction.NewService(transactionRepo, accountRepo, rdb, logger)

    // Handlers
    authHandler := auth.NewHandler(authService)
    accountHandler := account.NewHandler(accountService)
    transactionHandler := transaction.NewHandler(transactionService)

    // Router
    r := gin.Default()
    r.Use(middleware.CORS())
    r.Use(middleware.RequestID())

    v1 := r.Group("/api/v1")
    {
        // Public
        v1.POST("/register", authHandler.Register)
        v1.POST("/verify", authHandler.Verify)
        v1.POST("/login", authHandler.Login)
        v1.POST("/refresh", authHandler.Refresh)

        // Protected
        protected := v1.Group("")
        protected.Use(middleware.AuthRequired(cfg.JWTSecret))
        {
            protected.POST("/logout", authHandler.Logout)
            protected.POST("/accounts", accountHandler.Create)
            protected.GET("/accounts", accountHandler.List)
            protected.POST("/accounts/:id/block", accountHandler.Block)
            protected.POST("/transactions", transactionHandler.Create)
            protected.GET("/transactions", transactionHandler.List)
        }
    }

    r.Run(":" + cfg.Port)
}
```

---

## Concurrency Patterns

### Worker Pool
```go
type WorkerPool struct {
    jobs    chan Job
    results chan Result
    workers int
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        jobs:    make(chan Job, 100),
        results: make(chan Result, 100),
        workers: workers,
    }
}

func (wp *WorkerPool) Start(ctx context.Context) {
    var wg sync.WaitGroup
    for i := 0; i < wp.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range wp.jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    result := process(job)
                    wp.results <- result
                }
            }
        }()
    }
    go func() {
        wg.Wait()
        close(wp.results)
    }()
}
```

### Context Cancellation
```go
func (s *Service) Transfer(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    tx, err := s.db.Begin(ctx)
    if err != nil {
        return apperr.Wrap(err, "begin transaction")
    }
    defer tx.Rollback(ctx)

    // Check context before each operation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Perform operations...
    return tx.Commit(ctx)
}
```

---

## Logging

```go
// Structured logging with slog
slog.Info("transfer completed",
    "from", fromID,
    "to", toID,
    "amount", amount.String(),
    "duration", time.Since(start),
)

slog.Error("transfer failed",
    "from", fromID,
    "to", toID,
    "error", err,
)
```

---

## Testing

```go
// Unit test with mock
func TestCreateAccount(t *testing.T) {
    mockRepo := &MockRepository{}
    svc := NewService(mockRepo, slog.Default())

    account, err := svc.CreateAccount(context.Background(), "user-1", "USD")

    assert.NoError(t, err)
    assert.Equal(t, "USD", account.Currency)
    assert.Equal(t, decimal.NewFromInt(100), account.Balance)
}
```
