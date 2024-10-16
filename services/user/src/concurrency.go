package src

import (
	"sync"

	"go.opentelemetry.io/otel/trace"
)

var (
	tracer      trace.Tracer
	userChannel = make(chan User, 10) // Buffered channel for users
	wg          *sync.WaitGroup       // WaitGroup to synchronize goroutines
	done        = make(chan struct{}) // Channel to signal workers to stop
)

func GetTracer() trace.Tracer {
	return tracer
}

func GetUserChannel() chan User {
	if userChannel == nil {
		userChannel = make(chan User, 10) // Buffered channel for users
	}
	return userChannel
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

func SetUserChannel(oc chan User) {
	userChannel = oc
}

func SetWg(w *sync.WaitGroup) {
	wg = w
}
