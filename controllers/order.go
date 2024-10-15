package controllers

import (
	"context"
	"net/http"
	"time"

	"SimpleMicroserviceProject/constants"
	"SimpleMicroserviceProject/models"
	"SimpleMicroserviceProject/telemetry"

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	orderServiceTracer = otel.Tracer(constants.ServiceName)
	orderServiceLogger = otelslog.NewLogger(constants.ServiceName)
	orderServiceMeter  = otel.Meter(constants.ServiceName)
	orderCount         metric.Int64Counter
)

func init() {
	var err error
	orderCount, err = orderServiceMeter.Int64Counter("order.incr",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{incr}"))
	if err != nil {
		panic(err)
	}
}

// HandleOrder processes incoming order requests
func HandleOrder(w http.ResponseWriter, r *http.Request) {
	ctx, span := orderServiceTracer.Start(r.Context(), "HandleOrder")
	defer span.End()

	order := models.Order{ID: time.Now().Nanosecond(), Amount: 99.99}
	telemetry.GetOrderChannel() <- order

	orderServiceLogger.InfoContext(ctx, "Received new order", "result", log.Fields{
		"orderID": order.ID,
		"amount":  order.Amount,
		"client":  r.RemoteAddr,
		"method":  r.Method,
		"url":     r.URL.Path,
	})

	orderCountAttr := attribute.Int("order.count", 1)
	span.SetAttributes(orderCountAttr)
	orderCount.Add(ctx, 1, metric.WithAttributes(orderCountAttr))
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
