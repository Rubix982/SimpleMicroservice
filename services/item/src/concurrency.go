package src

import (
	"go.opentelemetry.io/otel/trace"
	"sync"
)

var (
	tracer       trace.Tracer
	orderChannel = make(chan Item, 10) // Buffered channel for orders
	wg           *sync.WaitGroup       // WaitGroup to synchronize goroutines
	done         = make(chan struct{}) // Channel to signal workers to stop
)

func GetTracer() trace.Tracer {
	return tracer
}

func GetItemChannel() chan Item {
	if orderChannel == nil {
		orderChannel = make(chan Item, 10) // Buffered channel for orders
	}
	return orderChannel
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

func SetOrderChannel(oc chan Item) {
	orderChannel = oc
}

func SetWg(w *sync.WaitGroup) {
	wg = w
}
