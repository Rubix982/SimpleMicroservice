package src

import "github.com/Rubix982/SimpleMicroserviceProject/services/order/src"

type Payment struct {
	ID               int
	Amount           float64
	OrderID          int
	Order            *src.Order
	Status           string
	PaymentGatewayID int
}
