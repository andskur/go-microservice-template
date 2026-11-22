package repository

import (
	"testing"
	"time"

	"microservice-template/config"
)

func TestPgOptions_AppliesSSLModeAndConnMaxLifetime(t *testing.T) {
	base := config.DatabaseConfig{Host: "db.internal", Port: 5432, ConnMaxLifetime: 300}

	tests := []struct {
		mode       string
		wantTLS    bool
		wantVerify bool // host name verified by crypto/tls
		wantErr    bool
	}{
		{mode: "disable"},
		{mode: ""},
		{mode: "require", wantTLS: true},
		{mode: "verify-ca", wantTLS: true},
		{mode: "verify-full", wantTLS: true, wantVerify: true},
		{mode: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			cfg := base
			cfg.SSLMode = tt.mode

			opts, err := pgOptions(&cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for ssl_mode %q", tt.mode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := opts.TLSConfig != nil; got != tt.wantTLS {
				t.Fatalf("TLS enabled = %v, want %v", got, tt.wantTLS)
			}
			if tt.wantVerify && (opts.TLSConfig.InsecureSkipVerify || opts.TLSConfig.ServerName != "db.internal") {
				t.Fatalf("verify-full must verify host name, got %+v", opts.TLSConfig)
			}
			if tt.mode == "verify-ca" && opts.TLSConfig.VerifyConnection == nil {
				t.Fatalf("verify-ca must verify the certificate chain")
			}
			if opts.MaxConnAge != 300*time.Second {
				t.Fatalf("MaxConnAge = %v, want 5m", opts.MaxConnAge)
			}
		})
	}
}

func TestPgOptions_RejectsNegativeLifetime(t *testing.T) {
	if _, err := pgOptions(&config.DatabaseConfig{ConnMaxLifetime: -1}); err == nil {
		t.Fatal("expected error for negative conn_max_lifetime")
	}
}
