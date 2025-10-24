// Package middleware provides HTTP middleware functions for the application.
package middleware

import (
	"net/http"

	"github.com/zjoart/countryxchange/pkg/logger"
)

// @Middleware		CorsMiddleware
// @Description	Handles Cross-Origin Resource Sharing (CORS) for HTTP requests
// @Usage			CorsMiddleware(allowedOrigins)
// @Checks			Validates origin against allowed origins, sets CORS headers, handles preflight requests
func CorsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqFields := logger.Fields{
				"path":   r.URL.Path,
				"method": r.Method,
			}

			origin := r.Header.Get("Origin")
			if origin != "" {
				reqFields["origin"] = origin
			}

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin || allowedOrigin == "*" {
					allowed = true
					break
				}
			}

			if !allowed && origin != "" {
				logger.Warn("blocked request from unauthorized origin", reqFields)
				http.Error(w, "Unauthorized origin", http.StatusForbidden)
				return
			}

			// Set CORS headers
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					reqFields["request_headers"] = reqHeaders
				}

				logger.Debug("handling CORS preflight request", reqFields)
				w.WriteHeader(http.StatusOK)
				return
			}

			logger.Debug("processing CORS request", reqFields)
			next.ServeHTTP(w, r)
		})
	}
}
