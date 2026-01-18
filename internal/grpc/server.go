package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	"microservice-template/config"
	"microservice-template/pkg/logger"
)

// Server wraps gRPC server and listener.
type Server struct {
	server   *grpc.Server
	health   *health.Server
	listener net.Listener
	addr     string
	running  bool
}

// NewServer creates a new gRPC server instance with configured middleware.
func NewServer(cfg *config.GRPCConfig) (*Server, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("parse timeout: %w", err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     360 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}

	opts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.ConnectionTimeout(timeout),
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(),
			recoveryInterceptor(),
		),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
	}

	if cfg.NumStreamWorkers > 0 {
		opts = append(opts, grpc.NumStreamWorkers(cfg.NumStreamWorkers))
	}

	server := grpc.NewServer(opts...)

	return &Server{
		addr:     addr,
		listener: listener,
		server:   server,
		running:  false,
	}, nil
}

// RegisterHealthService registers the standard gRPC health check service.
func (s *Server) RegisterHealthService() error {
	s.health = health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.server, s.health)

	s.health.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	logger.Log().Info("grpc health check service registered")
	return nil
}

// Serve starts the gRPC server (blocking).
func (s *Server) Serve() error {
	defer func() { s.running = false }()

	logger.Log().Infof("grpc server listening on %s", s.addr)

	if err := s.server.Serve(s.listener); err != nil {
		return fmt.Errorf("grpc serve error: %w", err)
	}

	return nil
}

// GracefulStop gracefully stops the gRPC server.
func (s *Server) GracefulStop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Server returns the underlying gRPC server for handler registration.
func (s *Server) Server() *grpc.Server {
	return s.server
}

// IsRunning returns true if server is currently running.
func (s *Server) IsRunning() bool {
	return s.running
}

// MarkRunning marks the server as running (used when Serve starts in a goroutine).
func (s *Server) MarkRunning() {
	s.running = true
}
