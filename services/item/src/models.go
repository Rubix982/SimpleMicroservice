package src

import "github.com/Rubix982/SimpleMicroserviceProject/services/order/src"

type Item struct {
	ID      int
	Name    string
	Price   float64
	Count   int
	OrderID int
	Order   *src.Order
}
