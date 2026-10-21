// Package order places orders. Each order is announced with an event.OrderPlaced message, which the
// service saves to the outbox in the transaction of the order, so the message is published if and only
// if the order is stored.
package order

import (
	"context"

	"github.com/efegreat/forgo/operation"
	"github.com/google/uuid"

	"example/internal/entity"
	"example/internal/errs"
	"example/internal/event"
	"example/internal/port"
)

// Service places and reads orders.
type Service struct {
	tx     operation.Transactor
	orders port.OrderRepository
	outbox port.Outbox
}

// NewService returns a Service that stores orders in orders and saves their messages to outbox, in
// transactions of tx.
func NewService(tx operation.Transactor, orders port.OrderRepository, outbox port.Outbox) *Service {
	return &Service{tx: tx, orders: orders, outbox: outbox}
}

// Place stores the order and saves its OrderPlaced message to the outbox, in one transaction.
func (s *Service) Place(ctx context.Context, req PlaceReq) (PlaceResp, error) {
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
