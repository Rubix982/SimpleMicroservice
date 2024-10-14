package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
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
	// Start worker goroutines to process orders.
	for i := 1; i <= 3; i++ { // 3 workers
		go processOrders(i)
	}

	// Start HTTP server
	http.HandleFunc("/order", handleOrder)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	// Close the channel after use.
	defer close(orderChannel)
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
