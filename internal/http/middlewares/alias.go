package middlewares

import "net/http"

// PathAlias serves requests for alias as if they were made to target
// (e.g. an unversioned /health for /api/v1/health).
func PathAlias(alias, target string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == alias {
				r2 := r.Clone(r.Context())
				r2.URL.Path = target
				r2.URL.RawPath = ""
				next.ServeHTTP(w, r2)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
