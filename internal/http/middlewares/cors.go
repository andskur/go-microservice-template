// Package middlewares contains HTTP middlewares.
package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"microservice-template/config"
)

// Cors middleware handles CORS (Cross-Origin Resource Sharing) headers.
func Cors(cfg *config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if cfg == nil || !cfg.Enabled {
			return next
		}

		allowedOrigins := buildOriginSet(cfg.AllowedOrigins)
		allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
		allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")
		maxAge := cfg.MaxAge

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && originAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if allowedOrigins["*"] {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			if allowedMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			}

			if allowedHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			}

			if maxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func buildOriginSet(origins []string) map[string]bool {
	set := make(map[string]bool, len(origins))
	for _, origin := range origins {
		set[origin] = true
	}
	return set
}

func originAllowed(origin string, allowed map[string]bool) bool {
	if allowed["*"] {
		return true
	}
	return allowed[origin]
}
