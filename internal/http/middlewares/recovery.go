package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"microservice-template/pkg/logger"
)

// Recovery middleware recovers from panics and returns a 500 Internal Server Error
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					logger.Log().Errorf("panic recovered: %v\nStack trace:\n%s", err, debug.Stack())

					// Return 500 error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(fmt.Sprintf(`{"code":500,"message":"internal server error"}`)))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
