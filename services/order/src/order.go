package src

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	orderServiceTracer = otel.Tracer(ServiceName)
	orderServiceLogger = otelslog.NewLogger(ServiceName)
	orderServiceMeter  = otel.Meter(ServiceName)
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
	ctx, span := orderServiceTracer.Start(r.Context(), "HandleOrder /order")
	defer span.End()

	span.AddEvent("Processing order", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
	order := Order{ID: time.Now().Nanosecond(), Amount: 99.99}
	GetOrderChannel() <- order

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
	span.AddEvent("Completed order", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
}

// ProcessOrders processes orders in the orderChannel
func ProcessOrders(ctx context.Context, workerID int) {
	defer GetWg().Done()

	for {
		select {
		case order, ok := <-GetOrderChannel():
			if !ok {
				return // Channel closed
			}

			// Start a new span for processing orders
			var span trace.Span
			ctx, span = GetTracer().Start(ctx, "processOrders")
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

		case <-GetDone():
			log.WithFields(log.Fields{
				"workerID": workerID,
			}).Info("Shutting down worker")
			return
		}
	}
}
