package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"microservice-template/config"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func newRateLimitHandler(t *testing.T, cfg *config.RateLimitConfig, exempt ...string) http.Handler {
	t.Helper()
	mw, err := RateLimit(cfg, exempt...)
	if err != nil {
		t.Fatalf("RateLimit: %v", err)
	}
	return mw(okHandler)
}

func doRequest(h http.Handler, path, remoteAddr, xff string) int {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimit_Disabled(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{Enabled: false})
	for i := 0; i < 10; i++ {
		if code := doRequest(h, "/test", "192.168.1.1:1234", ""); code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, code)
		}
	}
}

func TestRateLimit_Enabled(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{Enabled: true, RequestsPerSec: 2, Burst: 2})

	for i := 0; i < 2; i++ {
		if code := doRequest(h, "/test", "192.168.1.1:1234", ""); code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, code)
		}
	}
	if code := doRequest(h, "/test", "192.168.1.1:1234", ""); code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", code)
	}
}

func TestRateLimit_KeysOnIPNotPort(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{Enabled: true, RequestsPerSec: 1, Burst: 1})

	if code := doRequest(h, "/test", "192.168.1.1:1000", ""); code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", code)
	}
	// Same client, new source port (new TCP connection) must share the bucket.
	if code := doRequest(h, "/test", "192.168.1.1:2000", ""); code != http.StatusTooManyRequests {
		t.Fatalf("same IP, different port: got %d, want 429", code)
	}
	// IPv6 and IPv4-mapped forms are normalised.
	if code := doRequest(h, "/test", "[::ffff:192.168.1.1]:3000", ""); code != http.StatusTooManyRequests {
		t.Fatalf("IPv4-mapped IPv6 of same IP: got %d, want 429", code)
	}
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{Enabled: true, RequestsPerSec: 1, Burst: 1})
	for _, addr := range []string{"192.168.1.1:1234", "192.168.1.2:1234", "[2001:db8::1]:1234"} {
		if code := doRequest(h, "/test", addr, ""); code != http.StatusOK {
			t.Fatalf("request from %s: got %d, want 200", addr, code)
		}
	}
}

func TestRateLimit_IgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{
		Enabled: true, RequestsPerSec: 1, Burst: 1,
		TrustedProxies: []string{"10.0.0.0/8"},
	})

	if code := doRequest(h, "/test", "203.0.113.7:1000", "198.51.100.1"); code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", code)
	}
	// Spoofing a different X-Forwarded-For from an untrusted peer must not help.
	if code := doRequest(h, "/test", "203.0.113.7:1001", "198.51.100.2"); code != http.StatusTooManyRequests {
		t.Fatalf("spoofed XFF from untrusted peer: got %d, want 429", code)
	}
}

func TestRateLimit_UsesRightMostUntrustedHopFromTrustedProxy(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{
		Enabled: true, RequestsPerSec: 1, Burst: 1,
		TrustedProxies: []string{"10.0.0.0/8", "192.0.2.10"},
	})

	// Chain: client 198.51.100.1 -> 192.0.2.10 (trusted) -> 10.0.0.5 (trusted, peer).
	// The left-most entry is attacker-controlled and must be ignored.
	if code := doRequest(h, "/test", "10.0.0.5:1000", "1.1.1.1, 198.51.100.1, 192.0.2.10"); code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", code)
	}
	// Same real client, different spoofed left-most value: still limited.
	if code := doRequest(h, "/test", "10.0.0.6:1000", "2.2.2.2, 198.51.100.1, 192.0.2.10"); code != http.StatusTooManyRequests {
		t.Fatalf("same client behind proxies: got %d, want 429", code)
	}
	// A different real client behind the same proxies has its own bucket.
	if code := doRequest(h, "/test", "10.0.0.5:1000", "198.51.100.2, 192.0.2.10"); code != http.StatusOK {
		t.Fatalf("different client behind proxies: got %d, want 200", code)
	}
}

func TestRateLimit_ClientIP(t *testing.T) {
	rl, err := newRateLimiter(&config.RateLimitConfig{
		Enabled: true, RequestsPerSec: 1, Burst: 1, TrustedProxies: []string{"10.0.0.1", "10.1.0.0/16"},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, remote, xff, want string
	}{
		{"no proxy", "203.0.113.7:5555", "", "203.0.113.7"},
		{"no port", "203.0.113.7", "", "203.0.113.7"},
		{"untrusted peer with xff", "203.0.113.7:5555", "198.51.100.1", "203.0.113.7"},
		{"trusted peer no xff", "10.0.0.1:5555", "", "10.0.0.1"},
		{"trusted peer single hop", "10.0.0.1:5555", "198.51.100.1", "198.51.100.1"},
		{"all hops trusted", "10.0.0.1:5555", "10.1.2.3, 10.1.2.4", "10.1.2.3"},
		{"malformed hop", "10.0.0.1:5555", "198.51.100.1, garbage", "10.0.0.1"},
		{"ipv6 client", "10.0.0.1:5555", "2001:db8::1", "2001:db8::1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remote
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if got := rl.clientIP(req); got != tt.want {
				t.Fatalf("clientIP = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimit_ExemptPaths(t *testing.T) {
	h := newRateLimitHandler(t, &config.RateLimitConfig{Enabled: true, RequestsPerSec: 1, Burst: 1},
		"/health", "/api/v1/health")

	for i := 0; i < 5; i++ {
		for _, path := range []string{"/health", "/api/v1/health"} {
			if code := doRequest(h, path, "192.168.1.1:1234", ""); code != http.StatusOK {
				t.Fatalf("health request %d to %s: got %d, want 200", i, path, code)
			}
		}
	}
	// Health checks did not consume the client's tokens.
	if code := doRequest(h, "/test", "192.168.1.1:1234", ""); code != http.StatusOK {
		t.Fatalf("first non-exempt request: got %d, want 200", code)
	}
}

func TestRateLimit_EvictsIdleClients(t *testing.T) {
	rl, err := newRateLimiter(&config.RateLimitConfig{Enabled: true, RequestsPerSec: 1, Burst: 1, IdleTTL: "1m"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0)
	rl.now = func() time.Time { return now }

	rl.allow("198.51.100.1")
	rl.allow("198.51.100.2")
	if got := rl.size(); got != 2 {
		t.Fatalf("size = %d, want 2", got)
	}

	now = now.Add(45 * time.Second)
	rl.allow("198.51.100.2") // keeps .2 alive
	now = now.Add(30 * time.Second)
	rl.allow("198.51.100.3")

	// .1 idle 75s > 1m: evicted; .2 idle 30s and .3 new: kept.
	if got := rl.size(); got != 2 {
		t.Fatalf("size after sweep = %d, want 2", got)
	}
	rl.mu.Lock()
	_, stale := rl.clients["198.51.100.1"]
	rl.mu.Unlock()
	if stale {
		t.Fatal("idle client was not evicted")
	}
}

func TestRateLimit_InvalidConfig(t *testing.T) {
	bad := []*config.RateLimitConfig{
		{Enabled: true, RequestsPerSec: 0, Burst: 1},
		{Enabled: true, RequestsPerSec: 1, Burst: 0},
		{Enabled: true, RequestsPerSec: 1, Burst: 1, IdleTTL: "soon"},
		{Enabled: true, RequestsPerSec: 1, Burst: 1, TrustedProxies: []string{"not-an-ip"}},
		{Enabled: true, RequestsPerSec: 1, Burst: 1, TrustedProxies: []string{"10.0.0.0/99"}},
	}
	for i, cfg := range bad {
		if _, err := RateLimit(cfg); err == nil {
			t.Errorf("case %d: expected error for %+v", i, cfg)
		}
	}
}
