package src

import (
	"context"
	"net/http"
	"time"

	"SimpleMicroserviceProject/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	log "github.com/sirupsen/logrus"
)

var itemInstrument = telemetry.GetNewInstrumentation(ServiceName)

// HandleItem processes incoming item requests
func HandleItem(w http.ResponseWriter, r *http.Request) {
	ctx, span := itemInstrument.Tracer.Start(r.Context(), "HandleItem /item")
	defer span.End()

	span.AddEvent("Processing item", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
	item := Item{ID: time.Now().Nanosecond(), Price: 99.99}
	GetItemChannel() <- item

	itemInstrument.Logger.InfoContext(ctx, "Received new item", "result", log.Fields{
		"itemID": item.ID,
		"price":  item.Price,
		"client": r.RemoteAddr,
		"method": r.Method,
		"url":    r.URL.Path,
	})

	itemCountAttr := attribute.Int("item.count", 1)
	span.SetAttributes(itemCountAttr)
	itemInstrument.Counter.Add(ctx, 1, metric.WithAttributes(itemCountAttr))
	span.AddEvent("Completed item", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
}

// ProcessItems processes items in the itemChannel
func ProcessItems(ctx context.Context, workerID int) {
	defer GetWg().Done()

	for {
		select {
		case item, ok := <-GetItemChannel():
			if !ok {
				return // Channel closed
			}

			// Start a new span for processing items
			var span trace.Span
			ctx, span = GetTracer().Start(ctx, "processItems")
			defer span.End()

			logCtx := log.WithContext(ctx)
			logCtx.WithFields(log.Fields{
				"workerID": workerID,
				"itemID":   item.ID,
			}).Info("Processing item")

			time.Sleep(2 * time.Second) // Simulate item processing

			logCtx.WithFields(log.Fields{
				"workerID": workerID,
				"itemID":   item.ID,
			}).Info("Completed item")

		case <-GetDone():
			log.WithFields(log.Fields{
				"workerID": workerID,
			}).Info("Shutting down worker")
			return
		}
	}
}
