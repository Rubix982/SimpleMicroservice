package controllers

import (
	"SimpleMicroserviceProject/models"
	"SimpleMicroserviceProject/telemetry"
	"context"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
)

// HandleOrder processes incoming order requests
func HandleOrder(w http.ResponseWriter, r *http.Request) {
	// Start a new span for the order handling
	ctx, span := telemetry.GetTracer().Start(r.Context(), "HandleOrder")
	defer span.End()

	order := models.Order{ID: time.Now().Nanosecond(), Amount: 99.99}
	telemetry.GetOrderChannel() <- order

	logCtx := log.WithContext(ctx)
	logCtx.WithFields(log.Fields{
		"orderID": order.ID,
		"amount":  order.Amount,
		"client":  r.RemoteAddr,
	}).Info("Received new order")

	// Log the order received a message
	logCtx.WithContext(ctx).WithFields(log.Fields{
		"orderID": order.ID,
	}).Info("Order received")
}

// ProcessOrders processes orders in the orderChannel
func ProcessOrders(ctx context.Context, workerID int) {
	defer telemetry.GetWg().Done()

	for {
		select {
		case order, ok := <-telemetry.GetOrderChannel():
			if !ok {
				return // Channel closed
			}

			// Start a new span for processing orders
			var span trace.Span
			ctx, span = telemetry.GetTracer().Start(ctx, "processOrders")
			defer span.End()

			logCtx := log.WithContext(ctx)
			logCtx.WithFields(log.Fields{
				"workerID": workerID,
				"orderID":  order.ID,
			}).Info("Processing order")

			time.Sleep(2 * time.Second) // Simulate order processing

			logCtx.WithFields(log.Fields{
				"workerID": workerID,
				"orderID":  order.ID,
			}).Info("Completed order")

		case <-telemetry.GetDone():
			log.WithFields(log.Fields{
				"workerID": workerID,
			}).Info("Shutting down worker")
			return
		}
	}
}
