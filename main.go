package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	trace2 "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

// Order represents a simple order request.
type Order struct {
	ID     int
	Amount float64
}

var (
	orderChannel = make(chan Order, 10) // Buffered channel for orders
	wg           sync.WaitGroup         // WaitGroup to synchronize goroutines
	done         = make(chan struct{})  // Channel to signal workers to stop
)

const serviceName = "order-service"

var tracer trace.Tracer

func main() {
	// Set up logrus
	ctx := context.Background()
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	log.SetLevel(log.InfoLevel)

	// Initialize OpenTelemetry
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create the exporter: %v", err)
	}

	tp := trace2.NewTracerProvider(
		trace2.WithBatcher(exporter),
		trace2.WithResource(resource.NewWithAttributes(
			resource.Default().String(),
			attribute.String("service.name", serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)

	// Set up HTTP server with timeouts
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 1 * time.Second,  // Timeout for reading request headers
		ReadTimeout:       10 * time.Second, // Timeout for reading the entire request
		WriteTimeout:      10 * time.Second, // Timeout for writing responses
	}

	// Set up signal handling for graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start worker goroutines to process orders
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go processOrders(i)
	}

	// Register HTTP handlers
	http.HandleFunc("/order", loggingMiddleware(HandleOrder))
	http.HandleFunc("/health", loggingMiddleware(HandleHealthCheck))

	// Wrap the server handler with OpenTelemetry
	server.Handler = otelhttp.NewHandler(http.DefaultServeMux, "HTTP Server")

	// Run the server in a goroutine
	go func() {
		log.WithFields(log.Fields{
			"event": "startup",
			"port":  8080,
		}).Info("Server is starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Server failed")
		}
	}()

	// Wait for shutdown signal
	<-shutdownChan
	log.Info("Shutdown signal received, stopping server...")

	// Gracefully shut down the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server Shutdown Failed")
	}

	// Signal workers to stop and wait for them to finish processing
	close(done)
	close(orderChannel)
	wg.Wait()

	log.Info("Server gracefully stopped.")
}

// HandleOrder processes incoming order requests
func HandleOrder(w http.ResponseWriter, r *http.Request) {
	// Start a new span for the order handling
	ctx, span := tracer.Start(r.Context(), "HandleOrder")
	defer span.End()

	order := Order{ID: time.Now().Nanosecond(), Amount: 99.99}
	orderChannel <- order

	log.WithFields(log.Fields{
		"orderID": order.ID,
		"amount":  order.Amount,
		"client":  r.RemoteAddr,
	}).Info("Received new order")

	// Log the order received message
	log.WithFields(log.Fields{
		"orderID": order.ID,
	}).Info("Order received")
}

// processOrders processes orders in the orderChannel
func processOrders(workerID int) {
	defer wg.Done()

	for {
		select {
		case order, ok := <-orderChannel:
			if !ok {
				return // Channel closed
			}

			// Start a new span for processing orders
			ctx, span := tracer.Start(context.Background(), "processOrders")
			defer span.End()

			log.WithFields(log.Fields{
				"workerID": workerID,
				"orderID":  order.ID,
			}).Info("Processing order")

			time.Sleep(2 * time.Second) // Simulate order processing

			log.WithFields(log.Fields{
				"workerID": workerID,
				"orderID":  order.ID,
			}).Info("Completed order")

		case <-done:
			log.WithFields(log.Fields{
				"workerID": workerID,
			}).Info("Shutting down worker")
			return
		}
	}
}

// HandleHealthCheck provides a health check endpoint
func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Info("Health check requested")
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
