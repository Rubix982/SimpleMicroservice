package src

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SimpleMicroserviceProject/pkg/db"
	"SimpleMicroserviceProject/pkg/log"
	"SimpleMicroserviceProject/pkg/middleware"
	"SimpleMicroserviceProject/pkg/telemetry"

	"github.com/sirupsen/logrus"
)

func main() {
	db.ConnectDatabase()
	logger := log.InitLogger()
	ctx := context.Background()

	if err := setupOpenTelemetry(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to setup OpenTelemetry")
		return
	}

	// Set up HTTP server with timeouts
	server := middleware.GetHttpServer(ctx, []middleware.RouteMeta{
		middleware.GetRouteMeta("/item", HandleItem, "Get random item"),
		middleware.GetRouteMeta("/health", HandleHealthCheck, "Health check"),
	})

	// Set up signal handling for graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start worker goroutines to process items
	for i := 1; i <= 3; i++ {
		GetWg().Add(1)
		go ProcessItems(ctx, i)
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

	<-shutdownChan // Wait for shutdown signal
	handleShutdown(logger, ctx, server)
}

func setupOpenTelemetry(ctx context.Context) error {
	// Set up OpenTelemetry.
	openTelemetryShutdown, tp, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return err
	}
	SetTracer(tp.Tracer(ServiceName))
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, openTelemetryShutdown(context.Background()))
	}()
	return nil
}

func handleShutdown(logger *logrus.Logger, ctx context.Context, server *http.Server) {
	logger.Info("Shutdown signal received, stopping server...")

	// Gracefully shut down the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server Shutdown Failed")
	}

	// Signal workers to stop and wait for them to finish processing
	close(GetDone())
	close(GetItemChannel())
	GetWg().Wait()

	logger.Info("Server gracefully stopped.")
}
