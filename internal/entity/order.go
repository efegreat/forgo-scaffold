package entity

import (
	"time"

	"example/internal/errs"
)

// OrderStatus is the status of an Order.
type OrderStatus string

const (
	// OrderStatusPlaced is the status of an order that was placed.
	OrderStatusPlaced OrderStatus = "placed"
	// OrderStatusShipped is the status of an order whose shipment was dispatched.
	OrderStatusShipped OrderStatus = "shipped"
)

// Order is an order of a quantity of one SKU.
type Order struct {
	ID        string
	SKU       string
	Quantity  int
	Status    OrderStatus
	CreatedAt time.Time
}

// Ship marks a placed order shipped. An order in any other status fails with errs.ErrOrderNotShippable.
func (o *Order) Ship() error {
	if o.Status != OrderStatusPlaced {
		return errs.ErrOrderNotShippable
	}
	o.Status = OrderStatusShipped
	return nil
}
