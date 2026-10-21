package port

import (
	"context"

	"example/internal/entity"
)

// OrderRepository stores orders.
type OrderRepository interface {
	Insert(ctx context.Context, o entity.Order) error
	// Get returns the order of id, and false when there is none.
	Get(ctx context.Context, id string) (entity.Order, bool, error)
}

// ShipmentRepository stores shipments.
type ShipmentRepository interface {
	Insert(ctx context.Context, s entity.Shipment) error
	// DispatchPending dispatches every pending shipment, and returns how many it dispatched.
	DispatchPending(ctx context.Context) (int64, error)
}
