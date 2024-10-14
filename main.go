package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Order represents a simple order request.
type Order struct {
	ID     int
	Amount float64
}

// Channel to handle orders.
var orderChannel = make(chan Order, 10)

// WaitGroup to wait for all goroutines to finish.
var wg sync.WaitGroup

func main() {
	// Set up the server
	server := &http.Server{
		Addr: ":8080",
	}

	// Set up graceful shutdown handling
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start worker goroutines to process orders.
	for i := 1; i <= 3; i++ { // 3 workers
		go processOrders(i)
	}

	// Start HTTP server
	http.HandleFunc("/order", handleOrder)

	// Run server in a separate goroutine
	go func() {
		fmt.Println("Server is starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	// Listen for shutdown signal
	<-shutdownChan
	fmt.Println("Shutdown signal received, stopping server...")

	// Create a context with a timeout to ensure graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	// Close the order channel and wait for all orders to be processed
	close(orderChannel)
	wg.Wait()

	fmt.Println("Server gracefully stopped.")
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	// Simulate creating an order
	order := Order{ID: time.Now().Nanosecond(), Amount: 99.99}

	// Send order to the order channel.
	orderChannel <- order

	fmt.Fprintf(w, "Order received: %d\n", order.ID)
}

func processOrders(workerID int) {
	for order := range orderChannel {
		wg.Add(1)
		go func(order Order) {
			defer wg.Done()
			fmt.Printf("Worker %d processing order ID: %d\n", workerID, order.ID)
			time.Sleep(2 * time.Second) // Simulate some processing time
			fmt.Printf("Worker %d completed order ID: %d\n", workerID, order.ID)
		}(order)
	}
}
