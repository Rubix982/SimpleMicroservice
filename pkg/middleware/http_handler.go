package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type RouteMeta struct {
	Route       string
	Handler     http.HandlerFunc
	Description string
}

func GetRouteMeta(route string, handler http.HandlerFunc, description string) RouteMeta {
	return RouteMeta{
		Route:       route,
		Handler:     handler,
		Description: description,
	}
}

func GetHttpServer(ctx context.Context, routeMeta []RouteMeta) *http.Server {
	server := &http.Server{
		Addr:              ":8080",
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 1 * time.Second,  // Timeout for reading request headers
		ReadTimeout:       10 * time.Second, // Timeout for reading the entire request
		WriteTimeout:      10 * time.Second, // Timeout for writing responses
		Handler:           NewHTTPHandler(routeMeta),
	}
	return server
}

func NewHTTPHandler(routeMeta []RouteMeta) http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register HTTP handlers
	for _, route := range routeMeta {
		handleFunc(route.Route, loggingMiddleware(route.Handler))
	}

	// Add HTTP instrumentation for the whole server.
	return otelhttp.NewHandler(mux, "/")
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
