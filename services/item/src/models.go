package src

import "OrderService/src"

type Item struct {
	ID      int
	Name    string
	Price   float64
	Count   int
	OrderID int
	Order   *src.Order
}
