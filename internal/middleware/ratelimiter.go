package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	limiter  *time.Ticker
	tokens   int
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	rate    int           // tokens allowed per interval
	burst   int           // max burst size
	ttl     time.Duration // cleanup time
}

func NewRateLimiter(rate int, burst int, ttl time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rate:    rate,
		burst:   burst,
		ttl:     ttl,
	}

	// Background cleanup routine to prevent memory leaks
	go func() {
		for {
			time.Sleep(ttl)
			rl.mu.Lock()
			for ip, client := range rl.clients {
				if time.Since(client.lastSeen) > rl.ttl {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // fallback if port is missing
		}

		rl.mu.Lock()
		c, exists := rl.clients[ip]
		if !exists {
			c = &client{
				tokens:   rl.burst,
				lastSeen: time.Now(),
			}
			rl.clients[ip] = c
		}
		c.lastSeen = time.Now()

		// Simple refill logic based on elapsed time (simplified for purity)
		// For production scaling, Redis is typically preferred over in-memory.
		if c.tokens < rl.burst {
			c.tokens++
		}

		if c.tokens <= 0 {
			rl.mu.Unlock()
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		c.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}