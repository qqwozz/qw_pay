# Computer Science Fundamentals

**ID:** computer-science
**Version:** 2.0
**Category:** Computer Science & Algorithms
**Triggers:** algorithms, data structures, complexity, optimization, algorithms,Big O, performance, computer science

---

## Role

I am a computer science specialist. I apply fundamental CS principles to optimize the QW Pay platform, including algorithms, data structures, complexity analysis, and system design.

---

## 1. Algorithmic Complexity

### Big O Reference

| Complexity | Name | Example |
|------------|------|---------|
| O(1) | Constant | Hash table lookup |
| O(log n) | Logarithmic | Binary search |
| O(n) | Linear | Array scan |
| O(n log n) | Linearithmic | Merge sort |
| O(n²) | Quadratic | Nested loops |
| O(2ⁿ) | Exponential | Recursive Fibonacci |

### Applied to QW Pay

```go
// O(1) - Account lookup by ID (hash map)
account := accounts[accountID]

// O(log n) - Binary search for sorted transactions
func binarySearch(transactions []Transaction, target time.Time) int {
    low, high := 0, len(transactions)-1
    for low <= high {
        mid := (low + high) / 2
        if transactions[mid].CreatedAt.Equal(target) {
            return mid
        } else if transactions[mid].CreatedAt.Before(target) {
            low = mid + 1
        } else {
            high = mid - 1
        }
    }
    return -1
}

// O(n) - Find all user transactions
func findByUserID(transactions []Transaction, userID string) []Transaction {
    var result []Transaction
    for _, tx := range transactions {
        if tx.UserID == userID {
            result = append(result, tx)
        }
    }
    return result
}

// O(n log n) - Sort transactions by date
sort.Slice(transactions, func(i, j int) bool {
    return transactions[i].CreatedAt.Before(transactions[j].CreatedAt)
})
```

---

## 2. Data Structures

### Hash Table (Redis)
```go
// O(1) average case for lookups
// Used for: session storage, caching, rate limiting

// Redis SET command - O(1)
rdb.Set(ctx, "user:123:session", sessionData, 24*time.Hour)

// Redis GET command - O(1)
session, err := rdb.Get(ctx, "user:123:session").Result()
```

### B-Tree (PostgreSQL Index)
```
PostgreSQL uses B-tree indexes for:
- Primary keys (UUID)
- UNIQUE constraints (email, idempotency_key)
- Foreign keys (user_id, account_id)

Lookup: O(log n)
Range query: O(log n + k) where k is result size
```

### Queue (Redis Lists)
```go
// FIFO queue for anti-fraud processing
// LPUSH: O(1)
rdb.LPush(ctx, "anti-fraud:queue", transactionJSON)

// RPOP: O(1)
data, err := rdb.RPop(ctx, "anti-fraud:queue").Result()
```

### Sorted Set (Redis)
```go
// For leaderboards and ranking
// ZADD: O(log n)
rdb.ZAdd(ctx, "leaderboard", &redis.Z{
    Score:  float64(transactionCount),
    Member: userID,
})

// ZRANGE: O(log n + m) where m is result size
topUsers, err := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
```

---

## 3. Concurrency

### Mutex (Mutual Exclusion)
```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Get() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### WaitGroup
```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        processTransaction(id)
    }(i)
}

wg.Wait()
```

### Channels
```go
// Producer-consumer pattern
func producer(ch chan<- Transaction) {
    for _, tx := range transactions {
        ch <- tx
    }
    close(ch)
}

func consumer(ch <-chan Transaction, wg *sync.WaitGroup) {
    defer wg.Done()
    for tx := range ch {
        processTransaction(tx)
    }
}

ch := make(chan Transaction, 100)
go producer(ch)

var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)
    go consumer(ch, &wg)
}
wg.Wait()
```

### Context Cancellation
```go
func longRunningTask(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // Context cancelled or timeout
        default:
            // Do work
            if err := process(); err != nil {
                return err
            }
        }
    }
}

// Usage with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := longRunningTask(ctx)
```

---

## 4. Sorting Algorithms

### Quick Sort (for in-memory sorting)
```go
func quickSort(arr []int, low, high int) {
    if low < high {
        pivot := partition(arr, low, high)
        quickSort(arr, low, pivot-1)
        quickSort(arr, pivot+1, high)
    }
}

func partition(arr []int, low, high int) int {
    pivot := arr[high]
    i := low - 1
    for j := low; j < high; j++ {
        if arr[j] < pivot {
            i++
            arr[i], arr[j] = arr[j], arr[i]
        }
    }
    arr[i+1], arr[high] = arr[high], arr[i+1]
    return i + 1
}

// Average: O(n log n)
// Worst: O(n²)
// Space: O(log n)
```

### Merge Sort (for linked lists, stable sort)
```go
func mergeSort(arr []int) []int {
    if len(arr) <= 1 {
        return arr
    }
    mid := len(arr) / 2
    left := mergeSort(arr[:mid])
    right := mergeSort(arr[mid:])
    return merge(left, right)
}

func merge(left, right []int) []int {
    result := make([]int, 0, len(left)+len(right))
    i, j := 0, 0
    for i < len(left) && j < len(right) {
        if left[i] <= right[j] {
            result = append(result, left[i])
            i++
        } else {
            result = append(result, right[j])
            j++
        }
    }
    result = append(result, left[i:]...)
    result = append(result, right[j:]...)
    return result
}

// All cases: O(n log n)
// Space: O(n)
```

---

## 5. Graph Algorithms

### BFS (Breadth-First Search)
```go
// For finding shortest path in transaction network
func bfs(graph map[string][]string, start string) map[string]int {
    distances := make(map[string]int)
    queue := []string{start}
    distances[start] = 0

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        for _, neighbor := range graph[current] {
            if _, visited := distances[neighbor]; !visited {
                distances[neighbor] = distances[current] + 1
                queue = append(queue, neighbor)
            }
        }
    }
    return distances
}
```

### Cycle Detection
```go
// Detect circular transfers
func hasCycle(graph map[string][]string) bool {
    visited := make(map[string]bool)
    recStack := make(map[string]bool)

    var dfs func(node string) bool
    dfs = func(node string) bool {
        visited[node] = true
        recStack[node] = true

        for _, neighbor := range graph[node] {
            if !visited[neighbor] {
                if dfs(neighbor) {
                    return true
                }
            } else if recStack[neighbor] {
                return true
            }
        }
        recStack[node] = false
        return false
    }

    for node := range graph {
        if !visited[node] {
            if dfs(node) {
                return true
            }
        }
    }
    return false
}
```

---

## 6. Dynamic Programming

### Memoization (Top-Down)
```go
// Cache exchange rate calculations
var rateCache = make(map[string]decimal.Decimal)

func getExchangeRate(from, to string) decimal.Decimal {
    key := from + ":" + to
    if rate, ok := rateCache[key]; ok {
        return rate
    }
    
    // Calculate rate
    rate := calculateRate(from, to)
    rateCache[key] = rate
    return rate
}
```

### Tabulation (Bottom-Up)
```go
// Calculate maximum transfer amount with limits
func maxTransfer(amounts []decimal.Decimal, limit decimal.Decimal) decimal.Decimal {
    n := len(amounts)
    dp := make([]decimal.Decimal, n+1)
    
    for i := 1; i <= n; i++ {
        dp[i] = dp[i-1].Add(amounts[i-1])
        if dp[i].GreaterThan(limit) {
            dp[i] = limit
        }
    }
    return dp[n]
}
```

---

## 7. Trees

### Binary Search Tree (for sorted data)
```go
type TreeNode struct {
    Value int
    Left  *TreeNode
    Right *TreeNode
}

func insert(root *TreeNode, value int) *TreeNode {
    if root == nil {
        return &TreeNode{Value: value}
    }
    if value < root.Value {
        root.Left = insert(root.Left, value)
    } else {
        root.Right = insert(root.Right, value)
    }
    return root
}

func search(root *TreeNode, value int) bool {
    if root == nil {
        return false
    }
    if value == root.Value {
        return true
    }
    if value < root.Value {
        return search(root.Left, value)
    }
    return search(root.Right, value)
}

// Average: O(log n)
// Worst: O(n) (unbalanced)
```

---

## 8. Hashing

### SHA-256 (for refresh tokens)
```go
import "crypto/sha256"

func hashToken(token string) string {
    h := sha256.New()
    h.Write([]byte(token))
    return hex.EncodeToString(h.Sum(nil))
}

// Collision resistance: 2^128 operations
// Pre-image resistance: 2^256 operations
```

### FNV-1a (for consistent hashing)
```go
import "hash/fnv"

func hashKey(key string) uint32 {
    h := fnv.New32a()
    h.Write([]byte(key))
    return h.Sum32()
}

// Used for: sharding, load balancing
```

---

## 9. Complexity Analysis for QW Pay

| Operation | Current | Optimal | Notes |
|-----------|---------|---------|-------|
| User lookup | O(1) | O(1) | Hash index on email |
| Account lookup | O(1) | O(1) | Primary key index |
| Transaction list | O(n) | O(log n + k) | Add cursor-based pagination |
| Transfer | O(1) | O(1) | Direct updates |
| Anti-fraud check | O(1) | O(1) | Redis in-memory |
| Daily limit check | O(n) | O(1) | Cache in Redis |

---

## 10. Memory Management

### Go Memory Model
```go
// Escape analysis
func example() *int {
    x := 42       // Escapes to heap (returned)
    return &x
}

// Stack allocation
func example2() int {
    x := 42       // Stays on stack (not escaped)
    return x
}
```

### Garbage Collection Tuning
```bash
# Set GOGC (default 100)
GOGC=200 ./server  # Less frequent GC, more memory

# Set memory limit
GOMEMLIMIT=1GiB ./server
```

---

## 11. Network Algorithms

### Consistent Hashing (for distributed systems)
```go
type ConsistentHash struct {
    ring       map[uint32]string
    sortedKeys []uint32
    replicas   int
}

func (ch *ConsistentHash) GetNode(key string) string {
    hash := fnv.New32a()
    hash.Write([]byte(key))
    h := hash.Sum32()
    
    // Binary search for nearest node
    idx := sort.Search(len(ch.sortedKeys), func(i int) bool {
        return ch.sortedKeys[i] >= h
    })
    
    if idx >= len(ch.sortedKeys) {
        idx = 0
    }
    return ch.ring[ch.sortedKeys[idx]]
}
```

---

## 12. Cryptographic Algorithms

### bcrypt (password hashing)
```go
import "golang.org/x/crypto/bcrypt"

// Hash password - O(2^cost)
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verify password - O(2^cost)
err := bcrypt.CompareHashAndPassword(hash, []byte(password))
```

### HMAC-SHA256 (JWT signing)
```go
import "crypto/hmac"
import "crypto/sha256"

func sign(data, secret []byte) []byte {
    mac := hmac.New(sha256.New, secret)
    mac.Write(data)
    return mac.Sum(nil)
}

func verify(data, signature, secret []byte) bool {
    expected := sign(data, secret)
    return hmac.Equal(signature, expected)
}
```
