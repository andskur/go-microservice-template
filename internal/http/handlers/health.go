package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"microservice-template/internal/http/models"
	"microservice-template/internal/http/server/operations/health"
	"microservice-template/pkg/logger"
	"microservice-template/pkg/version"
)

// healthCheckTimeout bounds how long the health endpoint waits for module checks.
const healthCheckTimeout = 5 * time.Second

// HealthChecker returns the health of every module, keyed by module name
// (nil = healthy). module.Manager.HealthCheckAll satisfies it.
type HealthChecker func(ctx context.Context) map[string]error

// NewHealth creates new Health handler. A nil checker reports healthy.
func NewHealth(check HealthChecker) *Health {
	return &Health{check: check}
}

// Health handler reports service health based on the registered modules.
type Health struct {
	check HealthChecker
}

// Handle Health endpoint.
func (h *Health) Handle(params health.GetHealthParams) middleware.Responder {
	if unhealthy := h.unhealthyModules(params.HTTPRequest); len(unhealthy) > 0 {
		return h.HandleUnhealthy("unhealthy modules: " + strings.Join(unhealthy, ", "))
	}

	status := "healthy"
	ver, _ := version.NewVersion()
	versionStr := ver.Service
	if ver.Tag != "" {
		versionStr += " " + ver.Tag
	}

	return health.NewGetHealthOK().WithPayload(&models.Health{
		Status:    &status,
		Version:   versionStr,
		Timestamp: strfmt.DateTime(time.Now()),
	})
}

// unhealthyModules runs the module checks and returns the failing module names.
// Error details are logged, not returned to the caller.
func (h *Health) unhealthyModules(r *http.Request) []string {
	if h.check == nil {
		return nil
	}

	ctx := context.Background()
	if r != nil {
		ctx = r.Context()
	}
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	var unhealthy []string
	for name, err := range h.check(ctx) {
		if err != nil {
			logger.Log().Warnf("health check failed for module %s: %v", name, err)
			unhealthy = append(unhealthy, name)
		}
	}
	sort.Strings(unhealthy)
	return unhealthy
}

// HandleUnhealthy returns unhealthy status (for testing or when checks fail).
func (h *Health) HandleUnhealthy(message string) middleware.Responder {
	code := int64(http.StatusServiceUnavailable)
	return health.NewGetHealthServiceUnavailable().
		WithPayload(&models.Error{
			Code:    &code,
			Message: &message,
		})
}
