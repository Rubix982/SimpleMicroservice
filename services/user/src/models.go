package src

import "OrderService/src"

type User struct {
	ID     int
	Name   string
	Email  string
	Orders []src.Order
}
