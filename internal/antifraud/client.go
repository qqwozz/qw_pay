package antifraud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Verdict struct {
	ID        string `json:"id"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason"`
	RiskScore int    `json:"risk_score"`
	Engine    string `json:"engine"`
}

type TransferRequest struct {
	ID          string  `json:"id"`
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	UserID      string  `json:"user_id"`
}

type Client struct {
	rdb *redis.Client
}

func NewClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	return &Client{rdb: rdb}
}

func (c *Client) Check(ctx context.Context, fromAccount, toAccount string, amount float64, currency, userID string) (*Verdict, error) {
	req := TransferRequest{
		ID:          uuid.New().String(),
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
		Currency:    currency,
		UserID:      userID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	if err := c.rdb.LPush(ctx, "antifraud:queue", data).Err(); err != nil {
		return nil, fmt.Errorf("push to cpp queue: %w", err)
	}

	if err := c.rdb.LPush(ctx, "antifraud:queue:python", data).Err(); err != nil {
		log.Printf("[ANTIFRAUD] Failed to push to python queue: %v", err)
	}

	key := fmt.Sprintf("antifraud:verdict:%s", req.ID)
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			log.Printf("[ANTIFRAUD] Timeout waiting for verdict on %s, defaulting to APPROVED", req.ID)
			return &Verdict{
				ID:        req.ID,
				Approved:  true,
				Reason:    "Anti-fraud timeout — default approved",
				RiskScore: 0,
				Engine:    "timeout",
			}, nil
		case <-ticker.C:
			val, err := c.rdb.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			var verdict Verdict
			if err := json.Unmarshal([]byte(val), &verdict); err != nil {
				continue
			}
			log.Printf("[ANTIFRAUD] Verdict received: id=%s approved=%v risk=%d engine=%s reason=%s",
				verdict.ID, verdict.Approved, verdict.RiskScore, verdict.Engine, verdict.Reason)
			return &verdict, nil
		}
	}
}

func (c *Client) BlockUser(ctx context.Context, userID string) error {
	return c.rdb.SAdd(ctx, "antifraud:blocked_users", userID).Err()
}

func (c *Client) BlockAccount(ctx context.Context, accountID string) error {
	return c.rdb.SAdd(ctx, "antifraud:blocked_accounts", accountID).Err()
}

func (c *Client) RegisterAccountCreation(ctx context.Context, accountID string) error {
	return c.rdb.Set(ctx, "account:created:"+accountID, time.Now().Unix(), 24*time.Hour).Err()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
