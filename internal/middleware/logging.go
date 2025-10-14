package middleware

import (
	"net/http"
	"time"

	"github.com/NP-compete/gomcp/internal/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		log := logger.WithFields(map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      wrapped.statusCode,
			"duration":    duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
		})

		if wrapped.statusCode >= 500 {
			log.Error().Msg("HTTP request failed")
		} else if wrapped.statusCode >= 400 {
			log.Warn().Msg("HTTP request with client error")
		} else {
			log.Info().Msg("HTTP request completed")
		}
	})
}
