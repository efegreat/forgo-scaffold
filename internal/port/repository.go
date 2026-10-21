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
	// Update stores the status of o, an order that Get returned.
	Update(ctx context.Context, o entity.Order) error
}

// ShipmentRepository stores shipments.
type ShipmentRepository interface {
	Insert(ctx context.Context, s entity.Shipment) error
	// DispatchPending dispatches up to limit pending shipments, the oldest first, and returns them. Call it in
	// a transaction: the shipments it returns are locked until it ends, and skipped by other calls.
	DispatchPending(ctx context.Context, limit int) ([]entity.Shipment, error)
}
