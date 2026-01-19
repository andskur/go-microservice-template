package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"microservice-template/config"
)

// Cors middleware handles CORS (Cross-Origin Resource Sharing) headers
func Cors(cfg *config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CORS if disabled
			if cfg == nil || !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Set CORS headers
			if len(cfg.AllowedOrigins) > 0 {
				origin := r.Header.Get("Origin")
				if origin != "" {
					// Check if origin is allowed
					allowed := false
					for _, allowedOrigin := range cfg.AllowedOrigins {
						if allowedOrigin == "*" || allowedOrigin == origin {
							allowed = true
							break
						}
					}

					if allowed {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
				} else if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			}

			if len(cfg.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			}

			if len(cfg.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}

			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			// Handle preflight OPTIONS request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
