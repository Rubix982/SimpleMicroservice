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

var userInstrument = telemetry.GetNewInstrumentation(ServiceName)

// HandleUser processes incoming user requests
func HandleUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := userInstrument.Tracer.Start(r.Context(), "HandleUser /user")
	defer span.End()

	span.AddEvent("Processing user", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
	user := User{ID: time.Now().Nanosecond(), Email: "abc@example.com"}
	GetUserChannel() <- user

	userInstrument.Logger.InfoContext(ctx, "Received new user", "result", log.Fields{
		"userID": user.ID,
		"email":  user.Email,
		"client": r.RemoteAddr,
		"method": r.Method,
		"url":    r.URL.Path,
	})

	userCountAttr := attribute.Int("user.count", 1)
	span.SetAttributes(userCountAttr)
	userInstrument.Counter.Add(ctx, 1, metric.WithAttributes(userCountAttr))
	span.AddEvent("Completed user", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
}

// ProcessUsers processes users in the userChannel
func ProcessUsers(ctx context.Context, workerID int) {
	defer GetWg().Done()

	for {
		select {
		case user, ok := <-GetUserChannel():
			if !ok {
				return // Channel closed
			}

			// Start a new span for processing users
			var span trace.Span
			ctx, span = GetTracer().Start(ctx, "processUsers")
			defer span.End()

			logCtx := log.WithContext(ctx)
			logCtx.WithFields(log.Fields{
				"workerID": workerID,
				"userID":   user.ID,
			}).Info("Processing user")

			time.Sleep(2 * time.Second) // Simulate user processing

			logCtx.WithFields(log.Fields{
				"workerID": workerID,
				"userID":   user.ID,
			}).Info("Completed user")

		case <-GetDone():
			log.WithFields(log.Fields{
				"workerID": workerID,
			}).Info("Shutting down worker")
			return
		}
	}
}
