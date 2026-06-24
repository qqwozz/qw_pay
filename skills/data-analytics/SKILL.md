# Data Analytics

**ID:** data-analytics
**Version:** 2.0
**Category:** Analytics & Reporting
**Triggers:** analytics, metrics, SQL queries, dashboards, reports, monitoring, business intelligence

---

## Role

I am a data analyst specializing in financial systems. I analyze transaction data, build reports, and identify patterns for the QW Pay platform.

---

## Business Metrics

### Key Performance Indicators

| Metric | Formula | Target |
|--------|---------|--------|
| Daily Transaction Volume | `SUM(amount) WHERE created_at > CURRENT_DATE` | - |
| Average Transaction Size | `AVG(amount) WHERE status = 'EXECUTED'` | - |
| Success Rate | `COUNT(EXECUTED) / COUNT(*) * 100` | > 95% |
| Anti-fraud Block Rate | `COUNT(BLOCKED) / COUNT(*) * 100` | < 5% |
| Registration Conversion | `COUNT(verified) / COUNT(registered) * 100` | > 80% |
| Active Users (DAU) | `COUNT(DISTINCT user_id) WHERE last_login > NOW() - INTERVAL '1 day'` | - |

### Financial Metrics

| Metric | Description |
|--------|-------------|
| Total Balance | `SUM(balance) GROUP BY currency` |
| Daily Inflow | `SUM(amount) WHERE to_account_id IN (user_accounts)` |
| Daily Outflow | `SUM(amount) WHERE from_account_id IN (user_accounts)` |
| Net Flow | `inflow - outflow` |

---

## SQL Queries

### Daily Transaction Report
```sql
SELECT 
    DATE(created_at) as day,
    COUNT(*) as transactions,
    SUM(amount) as volume,
    AVG(amount) as avg_amount,
    COUNT(CASE WHEN status = 'EXECUTED' THEN 1 END) as successful,
    COUNT(CASE WHEN status = 'REJECTED' THEN 1 END) as rejected,
    COUNT(CASE WHEN status = 'BLOCKED_BY_FRAUD' THEN 1 END) as blocked
FROM transactions
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY day DESC;
```

### Top Users by Volume
```sql
SELECT 
    u.id as user_id,
    u.email,
    COUNT(t.id) as transaction_count,
    SUM(t.amount) as total_volume,
    AVG(t.amount) as avg_amount,
    MAX(t.created_at) as last_transaction
FROM transactions t
JOIN accounts a ON t.from_account_id = a.id
JOIN users u ON a.user_id = u.id
WHERE t.status = 'EXECUTED'
    AND t.created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY u.id, u.email
ORDER BY total_volume DESC
LIMIT 20;
```

### Anti-fraud Statistics
```sql
SELECT 
    DATE(created_at) as day,
    status,
    COUNT(*) as count,
    SUM(amount) as total_amount
FROM transactions
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY DATE(created_at), status
ORDER BY day DESC, status;
```

### Velocity Check (Real-time)
```sql
-- Transactions per account in last minute
SELECT 
    from_account_id,
    COUNT(*) as tx_count,
    SUM(amount) as total_amount
FROM transactions
WHERE created_at > NOW() - INTERVAL '1 minute'
GROUP BY from_account_id
HAVING COUNT(*) > 5;
```

### Daily Limit Monitoring
```sql
-- Accounts approaching daily limit
SELECT 
    from_account_id,
    SUM(amount) as daily_total,
    50000000 - SUM(amount) as remaining
FROM transactions
WHERE created_at > CURRENT_DATE
    AND status = 'EXECUTED'
GROUP BY from_account_id
HAVING SUM(amount) > 40000000
ORDER BY daily_total DESC;
```

### Currency Distribution
```sql
SELECT 
    currency,
    COUNT(*) as account_count,
    SUM(balance) as total_balance,
    AVG(balance) as avg_balance
FROM accounts
WHERE status = 'ACTIVE'
GROUP BY currency
ORDER BY total_balance DESC;
```

### Hourly Transaction Pattern
```sql
SELECT 
    EXTRACT(HOUR FROM created_at) as hour,
    COUNT(*) as transactions,
    SUM(amount) as volume
FROM transactions
WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
    AND status = 'EXECUTED'
GROUP BY EXTRACT(HOUR FROM created_at)
ORDER BY hour;
```

---

## Anti-fraud Analytics

### Suspicious Patterns
```sql
-- Accounts with high velocity
SELECT 
    from_account_id,
    COUNT(*) as tx_count,
    COUNT(DISTINCT to_account_id) as unique_recipients
FROM transactions
WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY from_account_id
HAVING COUNT(*) > 20 OR COUNT(DISTINCT to_account_id) > 10;
```

### Circular Transfers
```sql
-- Detect circular transfer chains
WITH RECURSIVE transfer_chain AS (
    SELECT 
        from_account_id,
        to_account_id,
        1 as depth
    FROM transactions
    WHERE created_at > NOW() - INTERVAL '1 hour'
    
    UNION ALL
    
    SELECT 
        tc.from_account_id,
        t.to_account_id,
        tc.depth + 1
    FROM transfer_chain tc
    JOIN transactions t ON tc.to_account_id = t.from_account_id
    WHERE tc.depth < 5
)
SELECT from_account_id, to_account_id, depth
FROM transfer_chain
WHERE from_account_id = to_account_id
    AND depth > 1;
```

---

## Recommended Indexes

```sql
-- For time-based queries
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

-- For status filtering
CREATE INDEX idx_transactions_status ON transactions(status);

-- For account lookups
CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);

-- For idempotency checks
CREATE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);

-- For user lookups
CREATE INDEX idx_accounts_user_id ON accounts(user_id);

-- Composite index for daily reports
CREATE INDEX idx_transactions_daily ON transactions(created_at, status, amount);
```

---

## Report Templates

### Daily Summary
```markdown
## Daily Report — [Date]

### Transactions
- Total: N
- Volume: $X
- Successful: N (X%)
- Blocked: N (X%)

### Top Accounts
| Account | Volume | Transactions |
|---------|--------|--------------|
| ... | ... | ... |

### Anti-fraud
- Blocks: N
- Velocity triggers: N
- Blacklist hits: N
```

### Weekly Trend
```markdown
## Weekly Trend — [Week]

### Comparison to Previous Week
- Volume: +X%
- Transactions: +X%
- Block rate: +X%

### Anomalies
- [Description of unusual patterns]
```
