package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Limiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	rate    int
	burst   int
	cleanup time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

func New(rate, burst int) *Limiter {
	l := &Limiter{
		clients: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		cleanup: 5 * time.Minute,
	}
	go l.cleanupLoop()
	return l
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, b := range l.clients {
			if now.Sub(b.lastReset) > l.cleanup {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *Limiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.clients[ip]
	if !ok {
		b = &bucket{tokens: l.burst, lastReset: time.Now()}
		l.clients[ip] = b
	}

	elapsed := time.Since(b.lastReset)
	refill := int(elapsed.Seconds()) * l.rate
	if refill > 0 {
		b.tokens += refill
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		b.lastReset = time.Now()
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
