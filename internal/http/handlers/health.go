package handlers

import (
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"microservice-template/internal/http/models"
	"microservice-template/internal/http/server/operations/health"
	"microservice-template/pkg/version"
)

// NewHealth creates new Health handler
func NewHealth() *Health {
	return &Health{}
}

// Health handler checks service health
type Health struct{}

// Handle Health endpoint
func (h *Health) Handle(params health.GetHealthParams) middleware.Responder {
	// TODO: Add actual health checks here
	// - Check database connectivity (if database module enabled)
	// - Check external dependencies
	// - Check resource availability (disk space, memory, etc.)

	status := "healthy"
	ver, _ := version.NewVersion()
	versionStr := ver.Service
	if ver.Tag != "" {
		versionStr += " " + ver.Tag
	}
	timestamp := strfmt.DateTime(time.Now())

	healthResponse := &models.Health{
		Status:    &status,
		Version:   versionStr,
		Timestamp: timestamp,
	}

	return health.NewGetHealthOK().WithPayload(healthResponse)
}

// HandleUnhealthy returns unhealthy status (for testing or when checks fail)
func (h *Health) HandleUnhealthy(message string) middleware.Responder {
	code := int64(http.StatusServiceUnavailable)
	return health.NewGetHealthServiceUnavailable().
		WithPayload(&models.Error{
			Code:    &code,
			Message: &message,
		})
}
