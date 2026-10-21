package entity

import "time"

// OrderStatus is the status of an Order.
type OrderStatus string

// OrderStatusPlaced is the status of an order that was placed.
const OrderStatusPlaced OrderStatus = "placed"

// Order is an order of a quantity of one SKU.
type Order struct {
	ID        string
	SKU       string
	Quantity  int
	Status    OrderStatus
	CreatedAt time.Time
}
