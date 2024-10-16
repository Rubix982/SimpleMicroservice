package src

import (
	"sync"

	"go.opentelemetry.io/otel/trace"
)

var (
	tracer         trace.Tracer
	paymentChannel = make(chan Payment, 10) // Buffered channel for payments
	wg             *sync.WaitGroup          // WaitGroup to synchronize goroutines
	done           = make(chan struct{})    // Channel to signal workers to stop
)

func GetTracer() trace.Tracer {
	return tracer
}

func GetPaymentChannel() chan Payment {
	if paymentChannel == nil {
		paymentChannel = make(chan Payment, 10) // Buffered channel for payments
	}
	return paymentChannel
}

func GetWg() *sync.WaitGroup {
	if wg == nil {
		wg = &sync.WaitGroup{}
	}
	return wg
}

func GetDone() chan struct{} {
	if done == nil {
		done = make(chan struct{})
	}
	return done
}

func SetTracer(t trace.Tracer) {
	tracer = t
}

func SetPaymentChannel(oc chan Payment) {
	paymentChannel = oc
}

func SetWg(w *sync.WaitGroup) {
	wg = w
}
