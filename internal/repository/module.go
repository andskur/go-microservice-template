package repository

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"

	"microservice-template/config"
	"microservice-template/pkg/logger"
)

// Module implements module.Module interface for repository layer.
// It manages database connection lifecycle and provides repository instances.
type Module struct {
	config *config.DatabaseConfig
	db     *pg.DB
	repo   IRepository
}

// NewModule creates a new repository module instance.
func NewModule(cfg *config.DatabaseConfig) *Module {
	return &Module{
		config: cfg,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "repository"
}

// Init initializes the repository module and establishes database connection.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module with driver: %s on %s:%d (database %s, sslmode %s)",
		m.Name(), m.config.Driver, m.config.Host, m.config.Port, m.config.Name, m.config.SSLMode)

	opts, err := pgOptions(m.config)
	if err != nil {
		return fmt.Errorf("database options: %w", err)
	}
	db := pg.Connect(opts)

	m.db = db
	m.repo = NewPostgresRepository(db)

	if err := m.HealthCheck(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Log().Infof("%s module initialized successfully", m.Name())
	return nil
}

// pgOptions converts the database config into go-pg connection options.
func pgOptions(cfg *config.DatabaseConfig) (*pg.Options, error) {
	tlsConfig, err := tlsConfigForSSLMode(cfg.SSLMode, cfg.Host)
	if err != nil {
		return nil, err
	}

	if cfg.ConnMaxLifetime < 0 {
		return nil, fmt.Errorf("conn_max_lifetime must be >= 0 seconds, got %d", cfg.ConnMaxLifetime)
	}

	return &pg.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:         cfg.User,
		Password:     cfg.Password,
		Database:     cfg.Name,
		TLSConfig:    tlsConfig,
		PoolSize:     cfg.MaxOpenConns,
		MinIdleConns: cfg.MaxIdleConns,
		MaxConnAge:   time.Duration(cfg.ConnMaxLifetime) * time.Second, // 0 = connections are not recycled
	}, nil
}

// tlsConfigForSSLMode maps libpq-style sslmode values to a TLS configuration.
// Supported: disable, require (encrypt, no certificate verification),
// verify-ca (verify the chain, not the host name) and verify-full.
func tlsConfigForSSLMode(mode, host string) (*tls.Config, error) {
	switch mode {
	case "", "disable":
		return nil, nil //nolint:nilnil // nil TLS config disables TLS in go-pg.
	case "require":
		return &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, //nolint:gosec // sslmode=require encrypts without verifying the server, as in libpq.
		}, nil
	case "verify-ca":
		return &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, //nolint:gosec // chain is verified in VerifyConnection; only the host name check is skipped.
			VerifyConnection:   verifyChainOnly,
		}, nil
	case "verify-full":
		return &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host}, nil
	default:
		return nil, fmt.Errorf("unsupported ssl_mode %q (use disable, require, verify-ca or verify-full)", mode)
	}
}

// verifyChainOnly verifies the server certificate chain against the system roots
// without checking the host name (libpq sslmode=verify-ca semantics).
func verifyChainOnly(cs tls.ConnectionState) error {
	if len(cs.PeerCertificates) == 0 {
		return errors.New("server presented no certificate")
	}
	intermediates := x509.NewCertPool()
	for _, cert := range cs.PeerCertificates[1:] {
		intermediates.AddCert(cert)
	}
	if _, err := cs.PeerCertificates[0].Verify(x509.VerifyOptions{Intermediates: intermediates}); err != nil {
		return fmt.Errorf("verify server certificate: %w", err)
	}
	return nil
}

// Start begins module operation (no-op for repository).
func (m *Module) Start(_ context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())
	return nil
}

// Stop gracefully shuts down the module and closes database connection.
func (m *Module) Stop(_ context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.db != nil {
		if err := m.db.Close(); err != nil {
			return fmt.Errorf("close database connection: %w", err)
		}
		logger.Log().Info("database connection closed")
	}

	return nil
}

// HealthCheck verifies database connectivity.
func (m *Module) HealthCheck(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if _, err := m.db.WithContext(ctx).Exec("SELECT 1"); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Repository returns the repository instance.
// This is used by other parts of the application (e.g., Service layer).
func (m *Module) Repository() IRepository {
	return m.repo
}
