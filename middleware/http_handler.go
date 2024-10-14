package middleware

import (
	"net/http"
	"time"

	"SimpleMicroserviceProject/controllers"

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register HTTP handlers
	handleFunc("/order", loggingMiddleware(controllers.HandleOrder))
	handleFunc("/health", loggingMiddleware(controllers.HandleHealthCheck))

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

// loggingMiddleware wraps handlers for request logging
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Info("Request started")

		next.ServeHTTP(w, r)

		log.WithFields(log.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		}).Info("Request completed")
	}
}
