package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"microservice-template/config"
)

// defaultRateLimitIdleTTL is used when http.rate_limit.idle_ttl is not set.
const defaultRateLimitIdleTTL = 5 * time.Minute

// RateLimit returns a per-client token bucket rate limiting middleware.
//
// Clients are keyed by IP address (the port of RemoteAddr is ignored). The
// X-Forwarded-For header is only honored when the direct peer is in
// cfg.TrustedProxies; the client is then the right-most address in the header
// that is not itself a trusted proxy. Limiters idle for longer than cfg.IdleTTL
// are evicted. Requests to exemptPaths (e.g. health checks) are never limited.
func RateLimit(cfg *config.RateLimitConfig, exemptPaths ...string) (func(http.Handler) http.Handler, error) {
	if cfg == nil || !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }, nil
	}

	rl, err := newRateLimiter(cfg)
	if err != nil {
		return nil, err
	}

	exempt := make(map[string]struct{}, len(exemptPaths))
	for _, p := range exemptPaths {
		exempt[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := exempt[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			if !rl.allow(rl.clientIP(r)) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"code":429,"message":"rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	now       func() time.Time
	clients   map[string]*limiterEntry
	lastSweep time.Time
	trusted   []netip.Prefix
	limit     rate.Limit
	burst     int
	idleTTL   time.Duration
	mu        sync.Mutex
}

func newRateLimiter(cfg *config.RateLimitConfig) (*rateLimiter, error) {
	if cfg.RequestsPerSec <= 0 {
		return nil, fmt.Errorf("rate_limit.requests_per_sec must be > 0, got %v", cfg.RequestsPerSec)
	}
	if cfg.Burst <= 0 {
		return nil, fmt.Errorf("rate_limit.burst must be > 0, got %d", cfg.Burst)
	}

	idleTTL := defaultRateLimitIdleTTL
	if cfg.IdleTTL != "" {
		d, err := time.ParseDuration(cfg.IdleTTL)
		if err != nil || d <= 0 {
			return nil, fmt.Errorf("rate_limit.idle_ttl must be a positive duration, got %q", cfg.IdleTTL)
		}
		idleTTL = d
	}

	trusted, err := parseTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}

	return &rateLimiter{
		now:     time.Now,
		clients: make(map[string]*limiterEntry),
		trusted: trusted,
		limit:   rate.Limit(cfg.RequestsPerSec),
		burst:   cfg.Burst,
		idleTTL: idleTTL,
	}, nil
}

// parseTrustedProxies accepts IP addresses and CIDR prefixes.
func parseTrustedProxies(values []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if strings.Contains(v, "/") {
			p, err := netip.ParsePrefix(v)
			if err != nil {
				return nil, fmt.Errorf("rate_limit.trusted_proxies: invalid CIDR %q: %w", v, err)
			}
			prefixes = append(prefixes, p.Masked())
			continue
		}
		addr, err := netip.ParseAddr(v)
		if err != nil {
			return nil, fmt.Errorf("rate_limit.trusted_proxies: invalid IP %q: %w", v, err)
		}
		addr = addr.Unmap()
		prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
	}
	return prefixes, nil
}

func (rl *rateLimiter) isTrusted(addr netip.Addr) bool {
	addr = addr.Unmap()
	for _, p := range rl.trusted {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// clientIP returns the rate limiting key for the request.
func (rl *rateLimiter) clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr // no port (e.g. unix socket or tests)
	}

	peer, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}
	peer = peer.Unmap()

	if len(rl.trusted) == 0 || !rl.isTrusted(peer) {
		return peer.String()
	}

	// The peer is a trusted proxy: walk X-Forwarded-For from the right and
	// return the first hop that is not a trusted proxy.
	hops := forwardedFor(r.Header.Values("X-Forwarded-For"))
	client := peer
	for i := len(hops) - 1; i >= 0; i-- {
		addr, err := netip.ParseAddr(hops[i])
		if err != nil {
			// Malformed entry appended by a trusted proxy: stop here and use
			// the last address we could trust.
			break
		}
		addr = addr.Unmap()
		client = addr
		if !rl.isTrusted(addr) {
			break
		}
	}
	return client.String()
}

// forwardedFor flattens (possibly repeated) X-Forwarded-For headers.
func forwardedFor(values []string) []string {
	var hops []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if part = strings.TrimSpace(part); part != "" {
				hops = append(hops, part)
			}
		}
	}
	return hops
}

// allow reports whether a request from key may proceed.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.sweepLocked(now)

	e, ok := rl.clients[key]
	if !ok {
		e = &limiterEntry{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.clients[key] = e
	}
	e.lastSeen = now

	return e.limiter.AllowN(now, 1)
}

// sweepLocked evicts idle limiters at most once per idleTTL/2.
func (rl *rateLimiter) sweepLocked(now time.Time) {
	if now.Sub(rl.lastSweep) < rl.idleTTL/2 {
		return
	}
	rl.lastSweep = now
	for key, e := range rl.clients {
		if now.Sub(e.lastSeen) > rl.idleTTL {
			delete(rl.clients, key)
		}
	}
}

// size returns the number of tracked clients (for tests).
func (rl *rateLimiter) size() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.clients)
}
