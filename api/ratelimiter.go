package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type client struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	clients   map[string]*client
	rate      float64
	maxTokens float64
}

func NewRateLimiter(rate float64, maxTokens float64) *RateLimiter {
	rl := &RateLimiter{
		clients:   make(map[string]*client),
		rate:      rate,
		maxTokens: maxTokens,
	}
	go rl.cleanupOldClients()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		rl.clients[ip] = &client{
			tokens:   rl.maxTokens - 1,
			lastSeen: time.Now(),
		}
		return true
	}

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(c.lastSeen).Seconds()
	c.tokens += elapsed * rl.rate
	if c.tokens > rl.maxTokens {
		c.tokens = rl.maxTokens
	}
	c.lastSeen = now

	if c.tokens < 1 {
		return false
	}

	c.tokens--
	return true
}

func (rl *RateLimiter) cleanupOldClients() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > 5*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests, please wait before submitting more jobs",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
