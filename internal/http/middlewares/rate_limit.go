package middlewares

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"microservice-template/config"
)

// RateLimit middleware implements token bucket rate limiting
func RateLimit(cfg *config.RateLimitConfig) func(http.Handler) http.Handler {
	// Skip rate limiting if disabled
	if cfg == nil || !cfg.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create a rate limiter per IP address
	type client struct {
		limiter  *rate.Limiter
		lastSeen int64
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Create limiter for requests per second with burst
	limiter := rate.Limit(cfg.RequestsPerSec)
	burst := cfg.Burst

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := r.RemoteAddr

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				// Create new limiter for this client
				c = &client{
					limiter: rate.NewLimiter(limiter, burst),
				}
				clients[ip] = c
			}
			mu.Unlock()

			// Check if request is allowed
			if !c.limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":429,"message":"rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
