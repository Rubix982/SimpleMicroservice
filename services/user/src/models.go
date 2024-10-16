package src

import (
	src2 "github.com/Rubix982/SimpleMicroserviceProject/services/order/src"
)

type User struct {
	ID     int
	Name   string
	Email  string
	Orders []src2.Order
}
