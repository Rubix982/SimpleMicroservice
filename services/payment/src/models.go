package src

import "OrderService/src"

type Payment struct {
	ID               int
	Amount           float64
	OrderID          int
	Order            *src.Order
	Status           string
	PaymentGatewayID int
}
