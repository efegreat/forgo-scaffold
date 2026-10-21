// Package shipping ships placed orders: it records a shipment for each event.OrderPlaced message, and a
// job dispatches the pending shipments.
package shipping

import (
	"context"

	"example/internal/entity"
	"example/internal/event"
	"example/internal/port"
)

// Service handles the messages of orders.
type Service struct {
	shipments port.ShipmentRepository
}

// NewService returns a Service that stores shipments in shipments.
func NewService(shipments port.ShipmentRepository) *Service {
	return &Service{shipments: shipments}
}

// OnOrderPlaced records the pending shipment of a placed order. The consumer runs it in a transaction
// together with the inbox, so a message delivered twice ships its order once.
func (s *Service) OnOrderPlaced(ctx context.Context, e event.OrderPlaced) error {
	return s.shipments.Insert(ctx, entity.Shipment{
		OrderID:  e.OrderID,
		SKU:      e.SKU,
		Quantity: e.Quantity,
		Status:   entity.ShipmentStatusPending,
	})
}
