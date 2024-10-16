package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	metric2 "go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	trace2 "go.opentelemetry.io/otel/trace"
)

type Instrumentation struct {
	Logger  *slog.Logger
	Tracer  trace2.Tracer
	Meter   metric2.Meter
	Counter metric2.Int64Counter
}

func GetNewInstrumentation(serviceName string) *Instrumentation {
	instrument := &Instrumentation{
		Logger: otelslog.NewLogger(serviceName),
		Tracer: otel.Tracer(serviceName),
		Meter:  otel.Meter(serviceName),
	}

	var err error
	instrument.Counter, err = instrument.Meter.Int64Counter(fmt.Sprintf("%s.incr", serviceName),
		metric2.WithDescription(fmt.Sprintf("The number of %v added", serviceName)),
		metric2.WithUnit("{incr}"))
	if err != nil {
		panic(err)
	}

	return instrument
}

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, tp *trace.TracerProvider, err error) {
	var shutdownFunctions []func(context.Context) error

	// Shutdown calls cleanup functions registered via shutdownFunc.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var joinedErr error
		for _, fn := range shutdownFunctions {
			if err := fn(ctx); err != nil {
				joinedErr = errors.Join(joinedErr, err)
			}
		}
		shutdownFunctions = nil
		return joinedErr
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := NewTraceProvider(5 * time.Second)
	if err != nil {
		handleErr(err)
		return
	}
	tp = tracerProvider
	shutdownFunctions = append(shutdownFunctions, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFunctions = append(shutdownFunctions, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up log provider.
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFunctions = append(shutdownFunctions, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func NewTraceProvider(batchTimeout time.Duration) (*trace.TracerProvider, error) {
	// Create an OTLP exporter to send traces to the Jaeger backend via OTLP
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("jaeger-collector.default.svc.cluster.local:4318"), // Use the OTLP HTTP endpoint for Jaeger
		otlptracehttp.WithInsecure(), // Disable TLS for local testing
	)

	traceExporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %v", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(batchTimeout)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("simple-microservice-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
