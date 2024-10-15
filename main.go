package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SimpleMicroserviceProject/constants"
	"SimpleMicroserviceProject/controllers"
	"SimpleMicroserviceProject/middleware"
	"SimpleMicroserviceProject/models"
	"SimpleMicroserviceProject/pkg/log"
	"SimpleMicroserviceProject/telemetry"
)

func main() {
	models.ConnectDatabase()
	logger := log.InitLogger()

	// Set up logrus
	ctx := context.Background()

	// Set up OpenTelemetry.
	openTelemetryShutdown, tp, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	telemetry.SetTracer(tp.Tracer(constants.ServiceName))
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, openTelemetryShutdown(context.Background()))
	}()

	// Set up HTTP server with timeouts
	server := &http.Server{
		Addr:              ":8080",
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 1 * time.Second,  // Timeout for reading request headers
		ReadTimeout:       10 * time.Second, // Timeout for reading the entire request
		WriteTimeout:      10 * time.Second, // Timeout for writing responses
		Handler:           middleware.NewHTTPHandler(),
	}

	// Set up signal handling for graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start worker goroutines to process orders
	for i := 1; i <= 3; i++ {
		telemetry.GetWg().Add(1)
		go controllers.ProcessOrders(ctx, i)
	}

	// Run the server in a goroutine
	go func() {
		logger.WithField("event", "startup").
			WithField("port", 8080).
			Info("Server is starting")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("Server failed")
		}
	}()

	// Wait for shutdown signal
	<-shutdownChan
	logger.Info("Shutdown signal received, stopping server...")

	// Gracefully shut down the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server Shutdown Failed")
	}

	// Signal workers to stop and wait for them to finish processing
	close(telemetry.GetDone())
	close(telemetry.GetOrderChannel())
	telemetry.GetWg().Wait()

	logger.Info("Server gracefully stopped.")
}
