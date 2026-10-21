// Package order places orders, and ships them once their shipments are dispatched. Each order is announced
// with an event.OrderPlaced message, which the service saves to the outbox in the transaction of the
// order, so the message is published if and only if the order is stored.
package order

import (
	"context"
	"fmt"

	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo/operation"
	"github.com/efegreat/forgo/server/worker"
	"github.com/google/uuid"

	"example/internal/entity"
	"example/internal/errs"
	"example/internal/event"
	"example/internal/port"
)

// Service places, reads and ships orders.
type Service struct {
	tx      operation.Transactor
	orders  port.OrderRepository
	outbox  port.Outbox
	configs *forgo.LiveConfigs
}

// NewService returns a Service that stores orders in orders and saves their messages to outbox, in
// transactions of tx, within the Limits that configs holds.
func NewService(tx operation.Transactor, orders port.OrderRepository, outbox port.Outbox, configs *forgo.LiveConfigs) *Service {
	return &Service{tx: tx, orders: orders, outbox: outbox, configs: configs}
}

// Place stores the order and saves its OrderPlaced message to the outbox, in one transaction. An order of
// more than Limits.MaxQuantity fails with errs.ErrQuantityAboveLimit.
func (s *Service) Place(ctx context.Context, req PlaceReq) (PlaceResp, error) {
	limits, err := s.configs.Load[Limits](ctx)
	if err != nil {
		return PlaceResp{}, err // the store failed: never place an order within limits that may be stale
	}
	if req.Quantity > limits.MaxQuantity {
		return PlaceResp{}, errs.ErrQuantityAboveLimit
	}

	o := entity.Order{
		ID:       uuid.Must(uuid.NewV7()).String(),
		SKU:      req.SKU,
		Quantity: req.Quantity,
		Status:   entity.OrderStatusPlaced,
	}
	msg, err := event.OrderPlaced{OrderID: o.ID, SKU: o.SKU, Quantity: o.Quantity}.Message()
	if err != nil {
		return PlaceResp{}, err
	}

	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.orders.Insert(ctx, o); err != nil {
			return err
		}
		return s.outbox.Save(ctx, msg)
	})
	if err != nil {
		return PlaceResp{}, err
	}
	return PlaceResp{ID: o.ID}, nil
}

// Get returns the order of req.ID, or errs.ErrOrderNotFound.
func (s *Service) Get(ctx context.Context, req GetReq) (GetResp, error) {
	o, ok, err := s.orders.Get(ctx, req.ID)
	if err != nil {
		return GetResp{}, err
	}
	if !ok {
		return GetResp{}, errs.ErrOrderNotFound
	}
	return newGetResp(o), nil
}

// OnShipmentDispatched ships the order whose shipment was dispatched. The consumer runs it in a transaction
// together with the inbox. Retrying cannot help a message about an order that does not exist or cannot be
// shipped, so those fail at once, with worker.ErrNoRetry, to the dead-letter queue.
func (s *Service) OnShipmentDispatched(ctx context.Context, e event.ShipmentDispatched) error {
	o, ok, err := s.orders.Get(ctx, e.OrderID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("order %s: %w: %w", e.OrderID, errs.ErrOrderNotFound, worker.ErrNoRetry)
	}
	if err := o.Ship(); err != nil {
		return fmt.Errorf("order %s: %w: %w", e.OrderID, err, worker.ErrNoRetry)
	}
	return s.orders.Update(ctx, o)
}
