package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
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

func main() {
	// Set up logrus
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	log.SetLevel(log.InfoLevel)

	// Set up HTTP server
	server := &http.Server{
		Addr: ":8080",
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
	order := Order{ID: time.Now().Nanosecond(), Amount: 99.99}
	orderChannel <- order

	log.WithFields(log.Fields{
		"orderID": order.ID,
		"amount":  order.Amount,
		"client":  r.RemoteAddr,
	}).Info("Received new order")

	fmt.Fprintf(w, "Order received: %d\n", order.ID)
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
	fmt.Fprintln(w, "OK")
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
