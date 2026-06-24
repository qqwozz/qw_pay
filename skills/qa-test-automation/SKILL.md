# QA & Test Automation

**ID:** qa-test-automation
**Version:** 2.0
**Category:** Quality Assurance
**Triggers:** test writing, test coverage, regression, edge cases, TDD, integration tests, unit tests

---

## Role

I am a senior QA engineer specializing in financial systems. I write comprehensive tests, identify edge cases, and ensure ACID compliance, idempotency, and race condition safety.

---

## Test Types

### Unit Tests
```go
func TestTransfer_Success(t *testing.T) {
    // Arrange
    fromAccount := &model.Account{
        ID:      uuid.New(),
        Balance: decimal.NewFromInt(1000),
        Version: 1,
        Status:  "ACTIVE",
    }
    toAccount := &model.Account{
        ID:      uuid.New(),
        Balance: decimal.NewFromInt(500),
        Version: 1,
        Status:  "ACTIVE",
    }
    amount := decimal.NewFromInt(100)

    // Act
    err := service.Transfer(ctx, fromAccount.ID, toAccount.ID, amount)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, decimal.NewFromInt(900), fromAccount.Balance)
    assert.Equal(t, decimal.NewFromInt(600), toAccount.Balance)
}
```

### Integration Tests
```go
func TestTransfer_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := transaction.NewRepository(db)
    svc := transaction.NewService(repo)

    // Create accounts
    from, _ := svc.CreateAccount(ctx, userID, "USD")
    to, _ := svc.CreateAccount(ctx, userID, "USD")

    // Transfer
    err := svc.Transfer(ctx, from.ID, to.ID, decimal.NewFromInt(50))

    // Verify in database
    var balance decimal.Decimal
    err = db.QueryRow(ctx, 
        "SELECT balance FROM accounts WHERE id = $1", from.ID,
    ).Scan(&balance)
    assert.Equal(t, decimal.NewFromInt(50), balance)
}
```

### Race Condition Tests
```go
func TestTransfer_ConcurrentAccess(t *testing.T) {
    account := createTestAccount(t, 1000)
    var wg sync.WaitGroup
    errors := make(chan error, 10)

    // Launch 10 concurrent transfers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := service.Transfer(ctx, account.ID, otherAccount.ID, decimal.NewFromInt(100))
            if err != nil {
                errors <- err
            }
        }()
    }

    wg.Wait()
    close(errors)

    // Verify final balance
    balance := getBalance(t, account.ID)
    // Should be 0 (1000 - 10*100 = 0)
    assert.Equal(t, decimal.NewFromInt(0), balance)
}
```

---

## Test Cases for QW Pay

### Authentication
| Test Case | Expected | Priority |
|-----------|----------|----------|
| Register new user | 201 Created | P0 |
| Register duplicate email | 400 Error | P0 |
| Verify correct OTP | 200 OK | P0 |
| Verify wrong OTP | 400 Error | P0 |
| Verify expired OTP | 400 Error | P1 |
| Login with correct password | 200 + tokens | P0 |
| Login with wrong password | 401 Error | P0 |
| Login unverified user | 403 Error | P0 |
| Refresh valid token | 200 + new tokens | P0 |
| Refresh revoked token | 401 Error | P0 |
| Logout | 200 OK | P1 |

### Accounts
| Test Case | Expected | Priority |
|-----------|----------|----------|
| Create account | 201 + 100 bonus | P0 |
| List user accounts | 200 + array | P0 |
| Block own account | 200 OK | P0 |
| Block others account | 403 Error | P0 |
| Block non-existent | 404 Error | P1 |

### Transactions
| Test Case | Expected | Priority |
|-----------|----------|----------|
| Successful transfer | 201 + EXECUTED | P0 |
| Insufficient funds | 400 Error | P0 |
| Transfer to self | 400 Error | P1 |
| Transfer blocked account | 403 Error | P0 |
| Idempotency (duplicate key) | 200 + same result | P0 |
| Cross-currency transfer | 201 + conversion | P1 |
| Exceed daily limit | 400 Error | P0 |
| Optimistic lock conflict | Retry succeeds | P0 |

### Anti-fraud
| Test Case | Expected | Priority |
|-----------|----------|----------|
| Velocity > 5/min | BLOCKED | P0 |
| Velocity > 20/hour | BLOCKED | P0 |
| Amount > 500K | BLOCKED | P0 |
| Blacklisted user | BLOCKED | P0 |
| Risk score >= 60 | BLOCKED | P0 |

---

## Test Commands

```bash
# Run all tests
make test

# Run with race detector
go test -race ./...

# Run specific package
go test -v ./internal/auth/...

# Run specific test
go test -v -run TestTransfer ./internal/transaction/

# Run without cache
go test -count=1 ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark tests
go test -bench=. ./internal/transaction/
```

---

## Mocking Pattern

```go
// Repository interface
type Repository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*Account, error)
    UpdateBalance(ctx context.Context, id string, amount decimal.Decimal, version int) error
}

// Mock repository
type MockRepository struct {
    accounts map[string]*Account
}

func (m *MockRepository) FindByID(ctx context.Context, id uuid.UUID) (*Account, error) {
    if acc, ok := m.accounts[id.String()]; ok {
        return acc, nil
    }
    return nil, ErrNotFound
}

func (m *MockRepository) UpdateBalance(ctx context.Context, id string, amount decimal.Decimal, version int) error {
    acc, ok := m.accounts[id]
    if !ok {
        return ErrNotFound
    }
    if acc.Version != version {
        return ErrOptimisticLock
    }
    acc.Balance = acc.Balance.Add(amount)
    acc.Version++
    return nil
}
```

---

## Test Report Format

```markdown
## Test Report — [Date]

### Summary
- Total: N
- Passed: N
- Failed: N
- Skipped: N
- Coverage: X%

### Failed Tests
- TestXxx: [error message]

### Performance
- Transfer latency p99: Xms
- RPS: X
```
