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

var paymentInstrument = telemetry.GetNewInstrumentation(ServiceName)

// HandlePayment processes incoming payment requests
func HandlePayment(w http.ResponseWriter, r *http.Request) {
	ctx, span := paymentInstrument.Tracer.Start(r.Context(), "HandlePayment /payment")
	defer span.End()

	span.AddEvent("Processing payment", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
	payment := Payment{ID: time.Now().Nanosecond(), Amount: 99.99}
	GetPaymentChannel() <- payment

	paymentInstrument.Logger.InfoContext(ctx, "Received new payment", "result", log.Fields{
		"paymentID": payment.ID,
		"amount":    payment.Amount,
		"client":    r.RemoteAddr,
		"method":    r.Method,
		"url":       r.URL.Path,
	})

	paymentCountAttr := attribute.Int("payment.count", 1)
	span.SetAttributes(paymentCountAttr)
	paymentInstrument.Counter.Add(ctx, 1, metric.WithAttributes(paymentCountAttr))
	span.AddEvent("Completed payment", trace.WithAttributes(
		attribute.String("method", r.Method),
		attribute.String("url", r.URL.Path)),
	)
}

// ProcessPayments processes payments in the paymentChannel
func ProcessPayments(ctx context.Context, workerID int) {
	defer GetWg().Done()

	for {
		select {
		case payment, ok := <-GetPaymentChannel():
			if !ok {
				return // Channel closed
			}

			// Start a new span for processing payments
			var span trace.Span
			ctx, span = GetTracer().Start(ctx, "processPayments")
			defer span.End()

			logCtx := log.WithContext(ctx)
			logCtx.WithFields(log.Fields{
				"workerID":  workerID,
				"paymentID": payment.ID,
			}).Info("Processing payment")

			time.Sleep(2 * time.Second) // Simulate payment processing

			logCtx.WithFields(log.Fields{
				"workerID":  workerID,
				"paymentID": payment.ID,
			}).Info("Completed payment")

		case <-GetDone():
			log.WithFields(log.Fields{
				"workerID": workerID,
			}).Info("Shutting down worker")
			return
		}
	}
}
