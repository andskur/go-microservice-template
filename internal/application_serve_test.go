package internal

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"microservice-template/config"
	httpmod "microservice-template/internal/http"
)

// crashingModule simulates a server whose background goroutine dies after Start.
type crashingModule struct {
	failed chan error
}

func (m *crashingModule) Name() string                      { return "crasher" }
func (m *crashingModule) Init(context.Context) error        { return nil }
func (m *crashingModule) Stop(context.Context) error        { return nil }
func (m *crashingModule) HealthCheck(context.Context) error { return nil }
func (m *crashingModule) Failed() <-chan error              { return m.failed }
func (m *crashingModule) Start(context.Context) error {
	m.failed = make(chan error, 1)
	go func() {
		defer close(m.failed)
		m.failed <- errors.New("accept: too many open files")
	}()
	return nil
}

func TestApp_ServeReturnsErrorWhenModuleFails(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatal(err)
	}
	app.modules.Register(&crashingModule{})

	done := make(chan error, 1)
	go func() { done <- app.serve(context.Background()) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected serve to return the module failure")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after a module failure")
	}
}

func TestApp_HTTPHealthReflectsModules(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatal(err)
	}
	app.config.HTTP = &config.HTTPConfig{
		Enabled:   true,
		Host:      "127.0.0.1",
		Port:      0,
		Timeout:   "5s",
		MockAuth:  true,
		CORS:      &config.CORSConfig{Enabled: false},
		RateLimit: &config.RateLimitConfig{Enabled: false},
	}
	if err := app.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		<-done
		_ = app.Stop()
	})

	var httpModule *httpmod.Module
	deadline := time.Now().Add(5 * time.Second)
	for httpModule == nil || httpModule.Addr() == nil {
		if time.Now().After(deadline) {
			t.Fatal("http module did not start")
		}
		for _, m := range app.modules.List() {
			if hm, ok := m.(*httpmod.Module); ok {
				httpModule = hm
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}, Timeout: 5 * time.Second}
	status := func(path string) int {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+httpModule.Addr().String()+path, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		_ = resp.Body.Close()
		return resp.StatusCode
	}

	for _, p := range []string{"/health", "/api/v1/health"} {
		if code := status(p); code != http.StatusOK {
			t.Fatalf("GET %s: got %d, want 200", p, code)
		}
	}

	// A failing module makes the HTTP health endpoints report 503.
	app.modules.Register(&unhealthyModule{})
	for _, p := range []string{"/health", "/api/v1/health"} {
		if code := status(p); code != http.StatusServiceUnavailable {
			t.Fatalf("GET %s with unhealthy module: got %d, want 503", p, code)
		}
	}
}

type unhealthyModule struct{ crashingModule }

func (m *unhealthyModule) Name() string                      { return "db" }
func (m *unhealthyModule) HealthCheck(context.Context) error { return errors.New("connection refused") }
